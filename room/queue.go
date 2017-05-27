package room

import "github.com/bcspragu/Radiotation/music"

type Queue struct {
	offset   int
	tracks   []music.Track
	trackMap map[string]music.Track
}

func (q *Queue) Offset() int {
	return q.offset
}

func (q *Queue) Tracks() []music.Track {
	return q.tracks
}

func (q *Queue) AddTrack(t music.Track) {
	q.tracks = append(q.tracks, t)
	q.trackMap[t.ID] = t
}

func (q *Queue) NextTrack() music.Track {
	if q.offset < len(q.tracks) {
		track := q.tracks[q.offset]
		delete(q.trackMap, track.ID)
		q.offset++
		return track
	}
	// TODO(bsprague): Probably add errors back
	return music.Track{}
}

func (q *Queue) RemoveTrack(delTrack music.Track) {
	for i, track := range q.tracks {
		if track.ID == delTrack.ID && i >= q.offset {
			q.tracks = append(q.tracks[:i], q.tracks[i+1:]...)
			delete(q.trackMap, track.ID)
		}
	}
}

func (q *Queue) HasTracks() bool {
	return len(q.tracks) > q.offset
}

func (q *Queue) NumTracks() int {
	return len(q.tracks)
}

func (q *Queue) HasTrack(track music.Track) bool {
	_, ok := q.trackMap[track.ID]
	return ok
}

func CountTracks(queues []*Queue) int {
	c := 0
	for _, q := range queues {
		c += q.NumTracks()
	}
	return c
}
