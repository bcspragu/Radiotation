// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func init() {
	r := mux.NewRouter()

	r.HandleFunc("/", withLogin(serveHome)).Methods("GET")
	r.HandleFunc("/rooms/{key}", withLogin(withRoom(serveRoom))).Methods("GET")
	r.HandleFunc("/rooms/{key}/create", withLogin(createRoom)).Methods("POST")
	r.HandleFunc("/search", withLogin(withRoom(serveSearch))).Methods("GET")
	r.HandleFunc("/queue", withLogin(withRoom(serveQueue))).Methods("GET")
	r.HandleFunc("/add", withLogin(withRoom(addToQueue))).Methods("POST")
	r.HandleFunc("/remove", withLogin(withRoom(removeFromQueue))).Methods("POST")
	r.HandleFunc("/pop", withRoom(serveSong)).Methods("GET")

	http.Handle("/", r)
}
