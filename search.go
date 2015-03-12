package main

import (
	"fmt"
	"net/http"
)

func serveSearch(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Host   string
		Tracks []Track
		Queue  *Queue
	}{
		r.Host,
		searchTrack(r.FormValue("search")),
		FindQueue(r),
	}
	err := templates.ExecuteTemplate(w, "search.html", data)
	if err != nil {
		fmt.Println(err)
	}
}
