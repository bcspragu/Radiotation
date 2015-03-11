package main

import (
	"net/http"
)

func serveSearch(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Host   string
		Tracks []Track
	}{
		r.Host,
		searchTrack(r.FormValue("search")),
	}
	templates.ExecuteTemplate(w, "search.html", data)
}
