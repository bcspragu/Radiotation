package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"room"
	"spotify"
	"time"

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
	r.HandleFunc("/rooms/{key}/remove", withLogin(removeFromQueue)).Methods("POST")
	r.HandleFunc("/rooms/{key}/pop", withLogin(serveSong)).Methods("GET")
	r.HandleFunc("/rooms/{key}/ws", withLogin(serveData)).Methods("GET")

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
	Tracks  []spotify.Track
}

type TrackResponse struct {
	Error   bool
	Message string
	Track   spotify.Track
}

func addToQueue(c Context) {
	c.w.Header().Set("Content-Type", "application/json")

	songID := c.r.FormValue("id")

	track := spotify.GetTrack(songID)

	data := QueueResponse{}
	err := c.Queue.Enqueue(track)
	if err != nil {
		data.Error = true
		data.Message = err.Error()
	}

	respString, _ := json.Marshal(data)
	fmt.Fprint(c.w, string(respString))
}

func removeFromQueue(c Context) {
	c.w.Header().Set("Content-Type", "application/json")

	songID := c.r.FormValue("id")

	track := spotify.GetTrack(songID)

	data := QueueResponse{}
	err := c.Queue.RemoveTrack(track)

	if err != nil {
		data.Error = true
		data.Message = err.Error()
	}
	respString, _ := json.Marshal(data)
	fmt.Fprint(c.w, string(respString))
}

func serveQueue(c Context) {
	data := allData{
		"Queue": c.Queue,
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
		_, t := c.Room.PopTrack()
		data.Track = t
		// TODO FIx This
		//if c, ok := h.connections[c.User.ID]; ok {
		//c.send <- []byte{}
		//}
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
		vars := mux.Vars(c.r)
		data := allData{
			"Room": vars["key"],
			"Host": c.r.Host,
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
