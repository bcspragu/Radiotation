package main

import (
	"net/http"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Host string
	}{
		r.Host,
	}
	templates.ExecuteTemplate(w, "home.html", data)
}
