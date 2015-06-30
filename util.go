package main

import (
	"math/rand"

	"github.com/gorilla/mux"

	"appengine"
	"appengine/datastore"

	"io"
	"net/http"
)

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "Internal Server Error")
	c.Errorf("%v", err)
}

func withRoom(hand func(http.ResponseWriter, *http.Request, *Room)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		vars := mux.Vars(r)
		roomName := vars["key"]
		var room Room
		err := datastore.Get(c, datastore.NewKey(c, "Room", roomName, 0, roomKey(c)), &room)
		switch err {
		case datastore.ErrNoSuchEntity:
			hand(w, r, nil)
		case nil:
			hand(w, r, &room)
		default:
			serveError(c, w, err)
		}
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func genName(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func roomKey(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "Room", "root", 0, nil)
}
