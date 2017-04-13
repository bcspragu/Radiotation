package spotify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/bcspragu/Radiotation/music"
)

type spotifySongServer struct {
	apiEndpoint string
}

type spotifyResponse struct {
	Tracks music.Tracks
}

func NewSongServer(apiEndpoint string) music.SongServer {
	return &spotifySongServer{
		apiEndpoint: apiEndpoint,
	}
}

func (s *spotifySongServer) Track(id string) (music.Track, error) {
	url := fmt.Sprintf("http://api.spotify.com/v1/tracks/%s", url.QueryEscape(id))
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
	url := fmt.Sprintf("http://api.spotify.com/v1/search?q=%s&type=track", url.QueryEscape(query))
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
