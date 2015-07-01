package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Room struct {
	Name   string
	Offset int
	Queues []*Queue
}

func (r *Room) NewQueue(name string) *Queue {
	q := &Queue{
		ID:     name,
		Tracks: []Track{},
	}
	r.Queues = append(r.Queues, q)
	return q
}

func (r *Room) Queue(id string) *Queue {
	for _, queue := range r.Queues {
		if queue.ID == id {
			return queue
		}
	}
	return nil
}

func (r *Room) HasTracks() bool {
	for _, queue := range r.Queues {
		if queue.HasTracks() {
			return true
		}
	}
	return false
}

func (r *Room) PopTrack() (*Queue, Track) {
	c := 0
	for c < len(r.Queues) {
		queue := r.Queues[(c+r.Offset)%len(r.Queues)]
		if queue.HasTracks() {
			track := queue.PopTrack()
			r.Offset = (c + r.Offset + 1) % len(r.Queues)
			return queue, track
		}
		c++
	}
	return nil, Track{}
}

func createRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomName := vars["key"]
	room := Room{
		Name:   roomName,
		Queues: []*Queue{},
	}

	if _, ok := rooms[roomName]; !ok {
		// Add the new room
		rooms[roomName] = &room
	}

	http.Redirect(w, r, "/rooms/"+roomName, 302)
}

func serveRoom(w http.ResponseWriter, req *http.Request, r *Room) {
	if r == nil {
		// Make the user create it first
		vars := mux.Vars(req)
		data := struct {
			Room string
			Host string
		}{
			vars["key"],
			req.Host,
		}
		err := templates.ExecuteTemplate(w, "new_room.html", data)
		if err != nil {
			serveError(w, err)
			return
		}
	} else {

		uID := userID(req)
		queue := r.Queue(uID)

		if queue == nil {
			queue = r.NewQueue(uID)
		}

		data := struct {
			Room  *Room
			Queue *Queue
			Host  string
		}{
			r,
			queue,
			req.Host,
		}
		err := templates.ExecuteTemplate(w, "room.html", data)
		if err != nil {
			serveError(w, err)
		}
	}
}
