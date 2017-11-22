package db

import (
	"github.com/bcspragu/Radiotation/music"
)

type QueueID struct {
	RoomID RoomID
	UserID UserID
}

func (id QueueID) String() string {
	return string(id.RoomID) + id.UserID.String()
}

type Queue struct {
	ID     QueueID
	Offset int
	Tracks []music.Track
}

func nextTrack(q *Queue) (music.Track, error) {
	if q.Offset < len(q.Tracks) {
		return q.Tracks[q.Offset], nil
	}
	return music.Track{}, ErrNoTracksInQueue
}
