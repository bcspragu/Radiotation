package main

import (
	"math/rand"

	"github.com/gorilla/mux"

	"fmt"
	"io"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func serveError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "Internal Server Error")
	fmt.Printf("%v", err)
}

func withRoom(hand func(http.ResponseWriter, *http.Request, *Room)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		roomName := vars["key"]
		hand(w, r, rooms[roomName])
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
