package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bcspragu/Radiotation/music"
)

type spotifySongServer struct {
	apiEndpoint string
	clientID    string
	secret      string
	tr          *tokenRefresher
}

type spotifyResponse struct {
	Tracks music.Tracks
}

type tokenRefresher struct {
	tkn       string
	exp       time.Time
	threshold time.Duration
}

func (s *spotifySongServer) token() string {
	remain := time.Now().Sub(s.tr.exp)
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

	r := io.TeeReader(resp.Body, os.Stdout)

	var tkn struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	err = json.NewDecoder(r).Decode(&tkn)
	if err != nil {
		log.Println(err)
		return ""
	}
	s.tr.tkn = tkn.AccessToken
	s.tr.exp = time.Now().Add(time.Duration(tkn.ExpiresIn) * time.Second)
	return s.tr.tkn
}

func NewSongServer(apiEndpoint, clientID, secret string) music.SongServer {
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
	r, _ := http.NewRequest(http.MethodPost, u, nil)
	r.Header.Set("Authorization", "Bearer "+s.token())
	return r
}

func (s *spotifySongServer) Track(id string) (music.Track, error) {
	url := fmt.Sprintf("http://api.%s/v1/tracks/%s", s.apiEndpoint, url.QueryEscape(id))
	req := s.requestWithAuth(url)
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return music.Track{}, fmt.Errorf("error querying Spotify API: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return music.Track{}, fmt.Errorf("error reading from Spotify API: %v", err)
	}
	var track music.Track
	err = json.Unmarshal(body, &track)
	if err != nil {
		return music.Track{}, fmt.Errorf("error loading data from Spotify API: %v", err)
	}
	return track, nil
}

func (s *spotifySongServer) Search(query string) ([]music.Track, error) {
	url := fmt.Sprintf("http://api.%s/v1/search?q=%s&type=track", s.apiEndpoint, url.QueryEscape(query))
	req := s.requestWithAuth(url)
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return []music.Track{}, fmt.Errorf("error querying Spotify API: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []music.Track{}, fmt.Errorf("error reading from Spotify API: %v", err)
	}
	var spotifyResp spotifyResponse
	err = json.Unmarshal(body, &spotifyResp)
	if err != nil {
		return []music.Track{}, fmt.Errorf("error loading data from Spotify API: %v", err)
	}
	return spotifyResp.Tracks.Items, nil
}
