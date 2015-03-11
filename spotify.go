package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Tracks []Track

type Track struct {
	Artists    []Artist
	Name       string
	ID         string
	IsPlayable bool
}

type Artist struct {
	Name   string
	Images []Image
}

type Image struct {
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
	var tracks Tracks
	err = json.Unmarshal(body, &tracks)
	if err != nil {
		fmt.Println("Error loading data from Spotify API:", err)
		return []Track{}
	}
	return tracks
}
