package main

import (
	"fmt"
	"net/http"
)

func serveSearch(w http.ResponseWriter, r *http.Request, room *Room) {
	data := struct {
		Host   string
		Tracks []Track
		Queue  *Queue
	}{
		r.Host,
		searchTrack(r.FormValue("search")),
		room.Queue(userID(r)),
	}
	err := templates.ExecuteTemplate(w, "search.html", data)
	if err != nil {
		fmt.Println(err)
	}
}
