package main

import (
	"net/http"

	"appengine"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Host string
	}{
		r.Host,
	}
	err := templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		serveError(appengine.NewContext(r), w, err)
	}
}
