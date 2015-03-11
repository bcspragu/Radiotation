package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Queue struct {
	ID     int
	Tracks []Track
}

var idToList map[int]*Queue

func init() {
	idToList = make(map[int]*Queue)
}

func NewQueue(id int) *Queue {
	q := new(Queue)
	q.ID = id
	q.Tracks = make([]Track, 0)
	return q
}

type QueueResponse struct {
	Error   bool
	Message string
}

func addToQueue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(r.Form)
	data := QueueResponse{Error: false}
	respString, _ := json.Marshal(data)
	fmt.Fprint(w, string(respString))
}
