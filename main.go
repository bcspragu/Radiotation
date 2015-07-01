// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var addr = flag.String("addr", ":8080", "http service address")
var env = flag.String("env", "production", "the environment to run in")
var dev bool
var rooms = make(map[string]*Room)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	flag.Parse()
	dev = *env == "development"
	go h.run()

	r := mux.NewRouter()

	r.HandleFunc("/", withLogin(serveHome)).Methods("GET")
	r.HandleFunc("/rooms/{key}", withLogin(withRoom(serveRoom))).Methods("GET")
	r.HandleFunc("/rooms/{key}/create", withLogin(createRoom)).Methods("POST")
	r.HandleFunc("/rooms/{key}/search", withLogin(withRoom(serveSearch))).Methods("GET")
	r.HandleFunc("/rooms/{key}/queue", withLogin(withRoom(serveQueue))).Methods("GET")
	r.HandleFunc("/rooms/{key}/add", withLogin(withRoom(addToQueue))).Methods("POST")
	r.HandleFunc("/rooms/{key}/remove", withLogin(withRoom(removeFromQueue))).Methods("POST")
	r.HandleFunc("/rooms/{key}/pop", withRoom(serveSong)).Methods("GET")
	r.HandleFunc("/rooms/{key}/ws", withLogin(withRoom(serveWs))).Methods("GET")

	// In production, static assets are served by nginx
	if dev {
		http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))
		http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
	}

	http.Handle("/", r)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
