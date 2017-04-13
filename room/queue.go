package room

import (
	"errors"

	"github.com/bcspragu/Radiotation/music"
)

type Queue struct {
	Offset   int
	Tracks   []music.Track
	TrackMap map[string]music.Track
}

func (q *Queue) AddTrack(newTrack music.Track) error {
	if q.HasTrack(newTrack) {
		return errors.New("Track is already in your queue, relax")
	}
	q.Tracks = append(q.Tracks, newTrack)
	q.TrackMap[newTrack.ID] = newTrack
	return nil
}

func (q *Queue) NextTrack() music.Track {
	if q.Offset < len(q.Tracks) {
		track := q.Tracks[q.Offset]
		delete(q.TrackMap, track.ID)
		q.Offset++
		return track
	}
	// TODO(bsprague): Probably add errors back
	return music.Track{}
}

func (q *Queue) RemoveTrack(delTrack music.Track) error {
	for i, track := range q.Tracks {
		if track.ID == delTrack.ID && i >= q.Offset {
			q.Tracks = append(q.Tracks[:i], q.Tracks[i+1:]...)
			delete(q.TrackMap, track.ID)
			return nil
		}
	}
	return errors.New("Track isn't in your queue, relax")
}

func (q *Queue) HasTracks() bool {
	return len(q.Tracks) > q.Offset
}

func (q *Queue) TrackCount() int {
	return len(q.Tracks)
}

func (q *Queue) HasTrack(track music.Track) bool {
	_, ok := q.TrackMap[track.ID]
	return ok
}

func CountTracks(queues []*Queue) int {
	trackCount := 0
	for _, q := range queues {
		trackCount += q.TrackCount()
	}
	return trackCount
}
