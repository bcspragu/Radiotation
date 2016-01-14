package main

import (
	"spotify"
)

func serveSearch(c Context) {
	data := allData{
		"Host":   c.r.Host,
		"Tracks": spotify.SearchTrack(c.r.FormValue("search")),
		"Queue":  c.Queue,
		"Room":   c.Room,
	}

	err := templates.ExecuteTemplate(c, "search.html", data)
	if err != nil {
		serveError(c.w, err)
	}
}
