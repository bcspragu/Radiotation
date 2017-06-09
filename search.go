package main

import (
	"net/http"

	"github.com/bcspragu/Radiotation/app"
	"github.com/bcspragu/Radiotation/music"
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

	tracks, err := songServer(rm).Search(r.FormValue("search"))
	if err != nil {
		serveError(w, err)
		return
	}

	q, err := s.db.Queue(rm.ID, u.ID)
	if err != nil {
		serveError(w, err)
		return
	}

	err = s.tmpls.ExecuteTemplate(w, "search.html", struct {
		Host   string
		Tracks []music.Track
		Queue  *app.Queue
		Room   *app.Room
	}{
		Host:   r.Host,
		Tracks: tracks,
		Queue:  q,
		Room:   rm,
	})
	if err != nil {
		serveError(w, err)
	}
}
