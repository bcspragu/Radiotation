package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bcspragu/Radiotation/radio"
)

type spotifySongServer struct {
	apiEndpoint string
	clientID    string
	secret      string
	tr          *tokenRefresher
}

type spotifyResponse struct {
	Tracks radio.Tracks
}

type tokenRefresher struct {
	tkn       string
	exp       time.Time
	threshold time.Duration
}

func (s *spotifySongServer) token() string {
	remain := s.tr.exp.Sub(time.Now())
	if !s.tr.exp.IsZero() && remain > s.tr.threshold {
		log.Printf("Loading cached token, expires in %s", remain.String())
		return s.tr.tkn
	}
	return s.getToken()
}

func (s *spotifySongServer) getToken() string {
	u := fmt.Sprintf("https://accounts.%s/api/token", s.apiEndpoint)

	form := url.Values{}
	form.Add("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		log.Println(err)
		return ""
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(s.clientID + ":" + s.secret))
	req.Header.Set("Authorization", "Basic "+encoded)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer resp.Body.Close()

	var tkn struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tkn)
	if err != nil {
		log.Println(err)
		return ""
	}
	s.tr.tkn = tkn.AccessToken
	s.tr.exp = time.Now().Add(time.Duration(tkn.ExpiresIn) * time.Second)
	return s.tr.tkn
}

func NewSongServer(apiEndpoint, clientID, secret string) radio.SongServer {
	s := &spotifySongServer{
		apiEndpoint: apiEndpoint,
		clientID:    clientID,
		secret:      secret,
		tr: &tokenRefresher{
			threshold: 5 * time.Second, // Expire the token 5 seconds before it actually expires
		},
	}
	s.token() // Preload our token
	return s
}

func (s *spotifySongServer) requestWithAuth(u string) *http.Request {
	r, _ := http.NewRequest(http.MethodGet, u, nil)
	r.Header.Set("Authorization", "Bearer "+s.token())
	return r
}

func (s *spotifySongServer) Track(id string) (radio.Track, error) {
	url := fmt.Sprintf("https://api.%s/v1/tracks/%s", s.apiEndpoint, url.QueryEscape(id))
	req := s.requestWithAuth(url)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return radio.Track{}, fmt.Errorf("error querying Spotify API: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return radio.Track{}, fmt.Errorf("error reading from Spotify API: %v", err)
	}
	var track radio.Track
	err = json.Unmarshal(body, &track)
	if err != nil {
		return radio.Track{}, fmt.Errorf("error loading data from Spotify API: %v", err)
	}
	return track, nil
}

func (s *spotifySongServer) Search(query string) ([]radio.Track, error) {
	url := fmt.Sprintf("https://api.%s/v1/search?q=%s&type=track", s.apiEndpoint, url.QueryEscape(query))
	req := s.requestWithAuth(url)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []radio.Track{}, fmt.Errorf("error querying Spotify API: %v", err)
	}
	defer resp.Body.Close()

	var spotifyResp spotifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&spotifyResp); err != nil {
		return []radio.Track{}, fmt.Errorf("error loading data from Spotify API: %v", err)
	}
	return spotifyResp.Tracks.Items, nil
}
