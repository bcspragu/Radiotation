package radio

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

type TrackList struct {
	Tracks        []Track
	QueueTrackIDs []string
	NextIndex     int
}
