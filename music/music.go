package music

import "strings"

type SongServer interface {
	Search(query string) ([]Track, error)
	Track(id string) (Track, error)
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
