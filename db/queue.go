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

func HasTracks(q *Queue) bool {
	return len(q.Tracks) > q.Offset
}

func HasTrack(q *Queue, track music.Track) bool {
	for _, t := range q.Tracks {
		if t.ID == track.ID {
			return true
		}
	}
	return false
}

func CountTracks(queues []*Queue) int {
	c := 0
	for _, q := range queues {
		c += len(q.Tracks)
	}
	return c
}
