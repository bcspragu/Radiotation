package main

import (
	"net/http"

	"github.com/bcspragu/Radiotation/music"
	"github.com/bcspragu/Radiotation/room"
)

func (s *srv) serveSearch(w http.ResponseWriter, r *http.Request) {
	rm, err := s.getRoom(r)
	if err != nil {
		serveError(w, err)
		return
	}

	u, err := s.user(r)
	if err != nil {
		serveError(w, err)
		return
	}

	tracks, err := rm.SongServer.Search(r.FormValue("search"))
	if err != nil {
		serveError(w, err)
		return
	}

	err = s.tmpls.ExecuteTemplate(w, "search.html", struct {
		Host   string
		Tracks []music.Track
		Queue  *room.Queue
		Room   *room.Room
	}{
		Host:   r.Host,
		Tracks: tracks,
		Queue:  u.Queue(rm.ID),
		Room:   rm,
	})
	if err != nil {
		serveError(w, err)
	}
}
