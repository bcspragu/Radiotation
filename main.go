package main

import (
	"encoding/json"
	"errors"
	"flag"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/bcspragu/Radiotation/music"
	"github.com/bcspragu/Radiotation/room"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

type tmplData struct {
	Host string
	Room *room.Room
}

var (
	rooms = make(map[string]*room.Room)
	users = make(map[string]*room.User)
	addr  = flag.String("addr", ":8000", "http service address")
	env   = flag.String("env", "development", "the environment to run in")

	dev   bool
	s     *securecookie.SecureCookie
	tmpls *template.Template
)

func main() {
	rand.Seed(time.Now().Unix())
	go h.run()

	flag.Parse()
	dev = *env == "development"

	r := mux.NewRouter()

	r.HandleFunc("/", withLogin(serveHome)).Methods("GET")
	r.HandleFunc("/rooms", withLogin(createRoom)).Methods("POST")
	r.HandleFunc("/rooms/{id}", withLogin(serveRoom)).Methods("GET")
	r.HandleFunc("/rooms/{id}/search", withLogin(serveSearch)).Methods("GET")
	r.HandleFunc("/rooms/{id}/queue", withLogin(serveQueue)).Methods("GET")
	r.HandleFunc("/rooms/{id}/add", withLogin(addToQueue)).Methods("POST")
	r.HandleFunc("/rooms/{id}/pop", withLogin(serveSong)).Methods("GET")
	r.HandleFunc("/rooms/{id}/ws", withLogin(serveData)).Methods("GET")

	http.Handle("/", r)

	var err error

	if err = servePaths(); err != nil {
		log.Fatalf("Can't serve static assets: %v", err)
	}

	if tmpls, err = template.ParseGlob("templates/*.html"); err != nil {
		log.Fatalf("Can't load templates, dying: %v", err)
	}

	if err = loadKeys(); err != nil {
		log.Fatalf("Can't load or generate keys, dying: %v", err)
	}

	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if err := tmpls.ExecuteTemplate(w, "index.html", data(r)); err != nil {
		serveError(w, err)
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

func addToQueue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rm, err := getRoom(r)
	if err != nil {
		jsonErr(w, err)
		return
	}

	queue, err := queue(r)
	if err != nil {
		jsonErr(w, err)
		return
	}

	track, err := rm.SongServer.Track(r.FormValue("id"))
	if err != nil {
		jsonErr(w, err)
		return
	}

	if queue.HasTrack(track) {
		err = queue.RemoveTrack(track)
	} else {
		err = queue.AddTrack(track)
	}

	if err != nil {
		jsonErr(w, err)
		return
	}
}

func jsonErr(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(QueueResponse{
		Error:   true,
		Message: err.Error(),
	})
}

func serveQueue(w http.ResponseWriter, r *http.Request) {
	rm, err := getRoom(r)
	if err != nil {
		log.Printf("Couldn't load room: %v", err)
		return
	}

	queue, err := queue(r)
	if err != nil {
		log.Printf("Couldn't load queue: %v", err)
		return
	}

	err = tmpls.ExecuteTemplate(w, "queue.html", struct {
		Queue *room.Queue
		Room  *room.Room
	}{queue, rm})
	if err != nil {
		log.Printf("Failed to execute queue template: %v", err)
	}
}

func serveSong(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rm, err := getRoom(r)
	if err != nil {
		log.Printf("Couldn't load room: %v", err)
		return
	}

	if !rm.HasTracks() {
		jsonErr(w, errors.New("No tracks to choose from"))
		return
	}

	u, t := rm.PopTrack()
	// Let the user know we're playing their track
	if c, ok := h.userconns[u]; ok {
		c.send <- []byte{}
	}

	err = json.NewEncoder(w).Encode(TrackResponse{
		Track: t,
	})
}

func createRoom(w http.ResponseWriter, r *http.Request) {
	dispName := r.PostFormValue("room")
	id := room.Normalize(dispName)
	if _, ok := rooms[id]; ok {
		http.Redirect(w, r, "/rooms/"+id, 302)
		return
	}

	// Add the new, non-existent room
	rm := room.New(dispName)
	switch r.PostFormValue("shuffle_order") {
	case "robin":
		rm.Rotator = room.RoundRobin()
	case "random":
		rm.Rotator = room.Shuffle()
	}
	rooms[id] = rm
	http.Redirect(w, r, "/rooms/"+id, 302)

}

func serveRoom(w http.ResponseWriter, r *http.Request) {
	rm, err := getRoom(r)
	if err != nil {
		id := roomID(r)
		log.Printf("No room found with ID %s", id)

		err := tmpls.ExecuteTemplate(w, "new_room.html", struct {
			DisplayName string
			ID          string
			Host        string
			Room        *room.Room
		}{mux.Vars(r)["id"], id, r.Host, nil})
		if err != nil {
			serveError(w, err)
		}
		return
	}

	u, err := user(r)
	if err != nil {
		serveError(w, err)
		return
	}

	queue, err := queue(r)
	if err != nil {
		log.Printf("Couldn't load queue for user in room, creating: %v", err)
		rm.AddUser(u)
	}

	err = tmpls.ExecuteTemplate(w, "room.html", struct {
		Room  *room.Room
		Queue *room.Queue
		Host  string
	}{rm, queue, r.Host})
	if err != nil {
		serveError(w, err)
	}
}

func data(r *http.Request) *tmplData {
	rm, _ := getRoom(r)
	return &tmplData{
		Host: r.Host,
		Room: rm,
	}
}
