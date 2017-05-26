package main

import (
	"net/http"

	"github.com/bcspragu/Radiotation/music"
	"github.com/bcspragu/Radiotation/room"
)

func serveSearch(w http.ResponseWriter, r *http.Request) {
	rm, err := getRoom(r)
	if err != nil {
		serveError(w, err)
		return
	}

	queue, err := queue(r)
	if err != nil {
		serveError(w, err)
		return
	}

	tracks, err := rm.SongServer.Search(r.FormValue("search"))
	if err != nil {
		serveError(w, err)
		return
	}

	err = tmpls.ExecuteTemplate(w, "search.html", struct {
		Host   string
		Tracks []music.Track
		Queue  *room.Queue
		Room   *room.Room
	}{
		Host:   r.Host,
		Tracks: tracks,
		Queue:  queue,
		Room:   rm,
	})
	if err != nil {
		serveError(w, err)
	}
}
