package radio

type SongServer interface {
	Search(query string) ([]Track, error)
	Track(id string) (Track, error)
}

type Tracks struct {
	Items []Track `json:"items"`
}

type Track struct {
	Artists []Artist `json:"artists"`
	Name    string   `json:"name"`
	ID      string   `json:"id"`
	Album   Album    `json:"album"`
}

type Album struct {
	Name   string  `json:"name"`
	Images []Image `json:"images"`
}

type Artist struct {
	Name string `json:"name"`
}

type Image struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	URL    string `json:"url"`
}
