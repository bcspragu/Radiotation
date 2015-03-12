package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Queue struct {
	Login  int
	Tracks []Track
}

var loginMap map[int]*Queue

func init() {
	loginMap = make(map[int]*Queue)
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

func (q *Queue) Enqueue(newTrack Track) error {
	for _, track := range q.Tracks {
		if track.ID == newTrack.ID {
			return errors.New("Track is already in your queue, relax")
		}
	}
	q.Tracks = append(q.Tracks, newTrack)
	return nil
}

func addToQueue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	queue := FindQueue(r)
	songID := r.FormValue("id")

	track := getTrack(songID)

	data := QueueResponse{Error: false}
	err := queue.Enqueue(track)

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
