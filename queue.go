package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Queue struct {
	ID     string
	Offset int
	Tracks []Track
}

type QueueResponse struct {
	Error   bool
	Message string
}

type TrackListResponse struct {
	Error   bool
	Message string
	Tracks  []Track
}

type TrackResponse struct {
	Error   bool
	Message string
	Track   Track
}

func (q *Queue) Enqueue(newTrack Track) error {
	for _, track := range q.Tracks {
		if track.ID == newTrack.ID {
			return errors.New("Track is already in your queue, relax")
		}
	}
	q.Tracks = append(q.Tracks, newTrack)
	return nil
}

func (q *Queue) PopTrack() Track {
	if q.Offset < len(q.Tracks) {
		track := q.Tracks[q.Offset]
		q.Offset++
		return track
	}
	// TODO(bsprague): Probably add errors back
	return Track{}
}

func (q *Queue) Remove(delTrack Track) error {
	for i, track := range q.Tracks {
		if track.ID == delTrack.ID && i >= q.Offset {
			q.Tracks = append(q.Tracks[:i], q.Tracks[i+1:]...)
			return nil
		}
	}
	return errors.New("Track isn't in your queue, relax")
}

func addToQueue(w http.ResponseWriter, r *http.Request, room *Room) {
	w.Header().Set("Content-Type", "application/json")

	queue := room.Queue(userID(r))
	songID := r.FormValue("id")

	track := getTrack(songID)

	data := QueueResponse{}
	err := queue.Enqueue(track)

	if err != nil {
		data.Error = true
		data.Message = err.Error()
	}
	respString, _ := json.Marshal(data)
	fmt.Fprint(w, string(respString))
}

func removeFromQueue(w http.ResponseWriter, r *http.Request, room *Room) {
	w.Header().Set("Content-Type", "application/json")

	queue := room.Queue(userID(r))
	songID := r.FormValue("id")

	track := getTrack(songID)

	data := QueueResponse{}
	err := queue.Remove(track)

	if err != nil {
		data.Error = true
		data.Message = err.Error()
	}
	respString, _ := json.Marshal(data)
	fmt.Fprint(w, string(respString))
}

func serveQueue(w http.ResponseWriter, r *http.Request, room *Room) {
	data := struct {
		Queue *Queue
	}{
		room.Queue(userID(r)),
	}
	err := templates.ExecuteTemplate(w, "queue.html", data)
	if err != nil {
		fmt.Println(err)
	}
}

func serveSong(w http.ResponseWriter, r *http.Request, room *Room) {
	w.Header().Set("Content-Type", "application/json")

	data := TrackResponse{}
	if room.HasTracks() {
		data.Track = room.PopTrack()
		// TODO(bsprague): Use channels to alert people the song is changing
		respString, _ := json.Marshal(data)
		fmt.Fprint(w, string(respString))
	} else {
		data.Error = true
		data.Message = "No tracks to choose from"
		respString, _ := json.Marshal(data)
		fmt.Fprint(w, string(respString))
	}
}

func (q *Queue) HasTracks() bool {
	return len(q.Tracks) >= q.Offset
}

func (q *Queue) TrackCount() int {
	return len(q.Tracks)
}

func CountTracks(queues []*Queue) int {
	trackCount := 0
	for _, q := range queues {
		trackCount += q.TrackCount()
	}
	return trackCount
}
