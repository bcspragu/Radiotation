package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var loginMap map[int]*Queue
var queues []*Queue
var queueIndex = 0

type Queue struct {
	Login  int
	Tracks []Track
}

func init() {
	loginMap = make(map[int]*Queue)
	queues = make([]*Queue, 0)
}

func NewQueue(login int) *Queue {
	q := new(Queue)
	q.Login = login
	q.Tracks = make([]Track, 0)
	return q
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

func (q *Queue) Pop() Track {
	var track Track
	track, q.Tracks = q.Tracks[0], q.Tracks[1:]
	return track
}

func (q *Queue) Remove(delTrack Track) error {
	for i, track := range q.Tracks {
		if track.ID == delTrack.ID {
			q.Tracks = append(q.Tracks[:i], q.Tracks[i+1:]...)
			return nil
		}
	}
	return errors.New("Track isn't in your queue, relax")
}

func addToQueue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	queue := FindQueue(r)
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

func removeFromQueue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	queue := FindQueue(r)
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

func serveQueue(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Queue *Queue
	}{
		FindQueue(r),
	}
	err := templates.ExecuteTemplate(w, "queue.html", data)
	if err != nil {
		fmt.Println(err)
	}
}

func FindQueue(r *http.Request) *Queue {
	login := LoginID(r)
	return loginMap[login]
}

func serveSong(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := TrackResponse{}
	if HasTracks() {
		for {
			// If there are no tracks in this queue, check the next one
			if len(queues[queueIndex].Tracks) == 0 {
				queueIndex = (queueIndex + 1) % len(queues)
			} else {
				data.Track = queues[queueIndex].Pop()
				h.connections[queues[queueIndex].Login].send <- []byte{}
				queueIndex = (queueIndex + 1) % len(queues)
				break
			}
		}
		respString, _ := json.Marshal(data)
		fmt.Fprint(w, string(respString))
	} else {
		data.Error = true
		data.Message = "No tracks to choose from"
		respString, _ := json.Marshal(data)
		fmt.Fprint(w, string(respString))
	}
}

func serveQueues(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	indices := make([]int, len(queues))
	data := TrackListResponse{}
	data.Tracks = make([]Track, CountTracks(queues))
	i := 0
	q_i := 0
	for {
		// Stop looping when we've filled our array
		if i == len(data.Tracks) {
			break
		}
		// Search until we find a queue that we haven't used all of
		for queues[q_i].TrackCount() == indices[q_i] {
			q_i = (q_i + 1) % len(queues)
		}
		// Add the track from the earliest unadded position in this queue
		data.Tracks[i] = queues[q_i].Tracks[indices[q_i]]
		indices[q_i]++
		q_i = (q_i + 1) % len(queues)
		i++
	}

	respString, _ := json.Marshal(data)
	fmt.Fprint(w, string(respString))
}

func HasTracks() bool {
	for _, q := range queues {
		if len(q.Tracks) > 0 {
			return true
		}
	}
	return false
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
