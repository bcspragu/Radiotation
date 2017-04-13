package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/bcspragu/Radiotation/music"
	"github.com/bcspragu/Radiotation/room"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

var s *securecookie.SecureCookie

var addr = flag.String("addr", ":8000", "http service address")
var env = flag.String("env", "development", "the environment to run in")
var dev bool

var templates RadioTemplate

var rooms = make(map[string]*room.Room)
var users = make(map[string]*room.User)

func main() {
	rand.Seed(time.Now().Unix())
	go h.run()

	flag.Parse()
	dev = *env == "development"

	r := mux.NewRouter()

	r.HandleFunc("/", withLogin(serveHome)).Methods("GET")
	r.HandleFunc("/rooms/{key}", withLogin(serveRoom)).Methods("GET")
	r.HandleFunc("/rooms/{key}/create", withLogin(createRoom)).Methods("POST")
	r.HandleFunc("/rooms/{key}/search", withLogin(serveSearch)).Methods("GET")
	r.HandleFunc("/rooms/{key}/queue", withLogin(serveQueue)).Methods("GET")
	r.HandleFunc("/rooms/{key}/add", withLogin(addToQueue)).Methods("POST")
	r.HandleFunc("/rooms/{key}/pop", withLogin(serveSong)).Methods("GET")
	r.HandleFunc("/rooms/{key}/ws", withLogin(serveData)).Methods("GET")

	http.Handle("/", r)

	var err error

	if err = servePaths(); err != nil {
		log.Fatal("Can't serve static assets", err)
	}

	if templates, err = loadTemplates(); err != nil {
		log.Fatal("Can't load templates, dying: ", err)
	}

	if err = loadKeys(); err != nil {
		log.Fatal("Can't load or generate keys, dying: ", err)
	}

	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serveHome(c Context) {
	dat := map[string]interface{}{}

	if err := templates.ExecuteTemplate(c, "index.html", dat); err != nil {
		serveError(c.w, err)
	}
}

type QueueResponse struct {
	Error   bool
	Message string
}

type TrackListResponse struct {
	Error   bool
	Message string
	Tracks  []music.Track
}

type TrackResponse struct {
	Error   bool
	Message string
	Track   music.Track
}

func addToQueue(c Context) {
	c.w.Header().Set("Content-Type", "application/json")

	data := QueueResponse{}

	track, err := c.Room.SongServer.Track(c.r.FormValue("id"))
	if err != nil {
		data.Error = true
		data.Message = err.Error()
		json.NewEncoder(c.w).Encode(data)
		return
	}

	if c.Queue.HasTrack(track) {
		err = c.Queue.RemoveTrack(track)
	} else {
		err = c.Queue.AddTrack(track)
	}

	if err != nil {
		data.Error = true
		data.Message = err.Error()
		json.NewEncoder(c.w).Encode(data)
		return
	}

}

func serveQueue(c Context) {
	data := allData{
		"Queue": c.Queue,
		"Raw":   true,
	}
	err := templates.ExecuteTemplate(c, "queue.html", data)
	if err != nil {
		fmt.Println(err)
	}
}

func serveSong(c Context) {
	c.w.Header().Set("Content-Type", "application/json")

	data := TrackResponse{}
	if c.Room.HasTracks() {
		u, t := c.Room.PopTrack()
		data.Track = t
		if c, ok := h.userconns[u]; ok {
			c.send <- []byte{}
		}
		respString, _ := json.Marshal(data)
		fmt.Fprint(c.w, string(respString))
	} else {
		data.Error = true
		data.Message = "No tracks to choose from"
		respString, _ := json.Marshal(data)
		fmt.Fprint(c.w, string(respString))
	}
}

func createRoom(c Context) {
	vars := mux.Vars(c.r)
	roomName := vars["key"]

	if _, ok := rooms[roomName]; !ok {
		// Add the new room
		rooms[roomName] = room.New(roomName)
	}

	http.Redirect(c.w, c.r, "/rooms/"+roomName, 302)
}

func serveRoom(c Context) {
	if c.Room == nil {
		// Make the user create it first
		data := allData{
			"Room": &room.Room{Name: mux.Vars(c.r)["key"]},
		}

		err := templates.ExecuteTemplate(c, "new_room.html", data)
		if err != nil {
			serveError(c.w, err)
			return
		}
	} else {

		if c.Queue == nil {
			c.Room.AddUser(c.User)
		}

		data := allData{
			"Room":  c.Room,
			"Queue": c.Queue,
			"Host":  c.r.Host,
		}

		err := templates.ExecuteTemplate(c, "room.html", data)
		if err != nil {
			serveError(c.w, err)
		}
	}
}
