package main

import (
	"fmt"
	"net/http"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Host  string
		Queue *Queue
	}{
		r.Host,
		FindQueue(r),
	}
	err := templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		fmt.Println(err)
	}
}
