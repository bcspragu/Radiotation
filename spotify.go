package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func (a *Album) Image() string {
	return a.Images[0].URL
}

type Artist struct {
	Name string
}

type Image struct {
	Width  int
	Height int
	URL    string
}

func searchTrack(trackName string) []Track {
	url := fmt.Sprintf("http://api.spotify.com/v1/search?q=%s&type=track", trackName)
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
