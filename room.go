package main

import (
	"net/http"

	"appengine"
	"appengine/datastore"

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

func (r *Room) PopTrack() Track {
	c := 0
	for c < len(r.Queues) {
		queue := r.Queues[(c+r.Offset)%len(r.Queues)]
		if queue.HasTracks() {
			track := queue.PopTrack()
			r.Offset = (c + r.Offset + 1) % len(r.Queues)
			return track
		}
		c++
	}
	return Track{}
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
			serveError(appengine.NewContext(req), w, err)
		}
	} else {
		data := struct {
			Room  *Room
			Queue *Queue
			Host  string
		}{
			r,
			r.Queue(userID(req)),
			req.Host,
		}
		err := templates.ExecuteTemplate(w, "room.html", data)
		if err != nil {
			serveError(appengine.NewContext(req), w, err)
		}
	}
}

func createRoom(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	vars := mux.Vars(r)
	roomName := vars["key"]
	room := Room{
		Name:   roomName,
		Queues: []*Queue{},
	}

	key := datastore.NewKey(c, "Room", roomName, 0, roomKey(c))
	var z *Room
	err := datastore.Get(c, key, z)

	// Only create if it doesn't exist
	if err == datastore.ErrNoSuchEntity {
		_, err := datastore.Put(c, key, &room)
		if err != nil {
			serveError(c, w, err)
			return
		}
	}

	http.Redirect(w, r, "/rooms/"+roomName, 302)
}
