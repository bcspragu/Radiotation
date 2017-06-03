package spotify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/bcspragu/Radiotation/music"
)

type spotifySongServer struct {
	apiEndpoint string
	tr          *tokenRefresher
}

type spotifyResponse struct {
	Tracks music.Tracks
}

type tokenRefresher struct {
	clientID string
	secret   string
	tkn      string
	exp      time.Time
}

func (tr *tokenRefresher) token() string {
	if !tr.exp.IsZero() && time.Now().Sub(tr.exp) > 0 {
		return tr.tkn
	}
	return tr.getToken()
}

func (tr *tokenRefresher) getToken() string {
	url := fmt.Sprintf("http://%s/v1/api/token", s.apiEndpoint)
	req := http.NewRequest(http.MethodPost, url, nil)
	req.SetBasicAuth(tr.clientID, tr.secret)
	http.DefaultClient.Do(req)
}

func NewSongServer(apiEndpoint, clientID, secret string) music.SongServer {
	s := &spotifySongServer{
		apiEndpoint: apiEndpoint,
		tr: &tokenRefresher{
			clientID: clientID,
			secret:   secret,
		},
	}
}

func (s *spotifySongServer) Track(id string) (music.Track, error) {
	url := fmt.Sprintf("http://%s/v1/tracks/%s", s.apiEndpoint, url.QueryEscape(id))
	resp, err := http.Get(url)
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
	url := fmt.Sprintf("http://%s/v1/search?q=%s&type=track", s.apiEndpoint, url.QueryEscape(query))
	resp, err := http.Get(url)
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
