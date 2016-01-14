package room

import (
	"errors"
	"spotify"
)

type Queue struct {
	Offset   int
	Tracks   []spotify.Track
	TrackMap map[string]spotify.Track
}

func (q *Queue) Enqueue(newTrack spotify.Track) error {
	if q.HasTrack(newTrack) {
		return errors.New("Track is already in your queue, relax")
	}
	q.Tracks = append(q.Tracks, newTrack)
	return nil
}

func (q *Queue) NextTrack() spotify.Track {
	if q.Offset < len(q.Tracks) {
		track := q.Tracks[q.Offset]
		delete(q.TrackMap, track.ID)
		q.Offset++
		return track
	}
	// TODO(bsprague): Probably add errors back
	return spotify.Track{}
}

func (q *Queue) RemoveTrack(delTrack spotify.Track) error {
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

func (q *Queue) HasTrack(track spotify.Track) bool {
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
