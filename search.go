package main

import (
	"net/http"
)

func serveSearch(w http.ResponseWriter, r *http.Request, room *Room) {
	data := struct {
		Host   string
		Tracks []Track
		Queue  *Queue
		Room   *Room
	}{
		r.Host,
		searchTrack(r.FormValue("search")),
		room.Queue(userID(r)),
		room,
	}
	err := templates.ExecuteTemplate(w, "search.html", data)
	if err != nil {
		serveError(w, err)
	}
}
