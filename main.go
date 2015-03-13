// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8080", "http service address")
var env = flag.String("env", "production", "the environment to run in")
var dev bool
var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	flag.Parse()
	dev = *env == "development"
	go h.run()

	r := mux.NewRouter()

	r.HandleFunc("/", withLogin(serveHome)).Methods("GET")
	r.HandleFunc("/search", serveSearch).Methods("GET")
	r.HandleFunc("/queue", serveQueue).Methods("GET")
	r.HandleFunc("/add", withLogin(addToQueue)).Methods("POST")
	r.HandleFunc("/remove", withLogin(removeFromQueue)).Methods("POST")
	r.HandleFunc("/next", serveSong).Methods("GET")
	r.HandleFunc("/ws", withLogin(serveWs)).Methods("GET")

	// In production, static assets are served by nginx
	if dev {
		http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))
		http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
		http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))
	}

	http.Handle("/", r)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
