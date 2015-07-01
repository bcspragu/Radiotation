package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type SpotifyResponse struct {
	Tracks Tracks
}

type Tracks struct {
	Items []Track
}

type Track struct {
	Artists []Artist
	Name    string
	ID      string
	Album   Album
}

type Album struct {
	Name   string
	Images []Image
}

type Artist struct {
	Name string
}

type Image struct {
	Width  int
	Height int
	URL    string
}

func (a *Album) Image() string {
	return a.Images[0].URL
}

func (t *Track) ArtistList() string {
	names := make([]string, len(t.Artists))
	for i, a := range t.Artists {
		names[i] = a.Name
	}
	return strings.Join(names, ", ")
}

func (t *Track) InQueue(q Queue) bool {
	for _, track := range q.Tracks[q.Offset:] {
		if track.ID == t.ID {
			return true
		}
	}
	return false
}

func getTrack(trackID string) Track {
	url := fmt.Sprintf("http://api.spotify.com/v1/tracks/%s", url.QueryEscape(trackID))
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("Error querying Spotify API:", err)
		return Track{}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading from Spotify API:", err)
		return Track{}
	}
	var track Track
	err = json.Unmarshal(body, &track)
	if err != nil {
		fmt.Println("Error loading data from Spotify API:", err)
		return Track{}
	}
	return track
}

func searchTrack(trackName string) []Track {
	url := fmt.Sprintf("http://api.spotify.com/v1/search?q=%s&type=track", url.QueryEscape(trackName))
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("Error querying Spotify API:", err)
		return []Track{}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading from Spotify API:", err)
		return []Track{}
	}
	var spotifyResp SpotifyResponse
	err = json.Unmarshal(body, &spotifyResp)
	if err != nil {
		fmt.Println("Error loading data from Spotify API:", err)
		return []Track{}
	}
	return spotifyResp.Tracks.Items
}
