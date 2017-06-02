package main

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bcspragu/Radiotation/app"
	"github.com/bcspragu/Radiotation/music"
	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/namsral/flag"
)

type tmplData struct {
	Host string
	Room *app.Room
	User *app.User
}

type srv struct {
	db db

	tmpls *template.Template
	sc    *securecookie.SecureCookie
	h     hub
}

var (
	_        = flag.String(flag.DefaultConfigFlagname, "config", "path to config file")
	addr     = flag.String("addr", ":8000", "http service address")
	clientID = flag.String("client_id", "", "The Google ClientID to use")

	googleVerifier *oidc.IDTokenVerifier
)

func main() {
	rand.Seed(time.Now().Unix())
	flag.Parse()

	if *clientID == "" {
		log.Fatalf("Missing required --client_id")
	}

	tmpls, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Can't load templates, dying: %v", err)
	}

	sc, err := loadKeys()
	if err != nil {
		log.Fatalf("Can't load or generate keys, dying: %v", err)
	}

	googleProvider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		log.Fatalf("Failed to get provider for Google: %v", err)
	}
	googleVerifier = googleProvider.Verifier(&oidc.Config{
		ClientID: *clientID,
	})

	db, err := initBoltDB()
	if err != nil {
		log.Fatalf("Failed to initialize datastore: %v", err)
	}

	s := &srv{
		db:    db,
		tmpls: tmpls,
		sc:    sc,
		h: hub{
			broadcast:   make(chan []byte),
			register:    make(chan *connection),
			unregister:  make(chan *connection),
			connections: make(map[*connection]bool),
			userconns:   make(map[*app.User]*connection),
		},
	}
	go s.h.run()

	r := mux.NewRouter()
	r.HandleFunc("/", s.serveHome).Methods("GET")
	r.HandleFunc("/verifyToken", s.serveVerifyToken)
	r.HandleFunc("/rooms", s.withLogin(s.createRoom)).Methods("POST")
	r.HandleFunc("/rooms/{id}", s.withLogin(s.serveRoom)).Methods("GET")
	r.HandleFunc("/rooms/{id}/search", s.withLogin(s.serveSearch)).Methods("GET")
	r.HandleFunc("/rooms/{id}/queue", s.withLogin(s.serveQueue)).Methods("GET")
	r.HandleFunc("/rooms/{id}/now", s.withLogin(s.nowPlaying)).Methods("GET")
	r.HandleFunc("/rooms/{id}/add", s.withLogin(s.addToQueue)).Methods("POST")
	r.HandleFunc("/rooms/{id}/remove", s.withLogin(s.removeFromQueue)).Methods("POST")
	r.HandleFunc("/rooms/{id}/pop", s.withLogin(s.serveSong)).Methods("GET")
	r.HandleFunc("/rooms/{id}/ws", s.withLogin(s.serveData)).Methods("GET")

	http.Handle("/", r)

	if err := servePaths(); err != nil {
		log.Fatalf("Can't serve static assets: %v", err)
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// Got the signal to die, save some stuff first
		if err := s.saveState("server-state"); err != nil {
			log.Printf("Failed to save server state: %v", err)
		}
	}()

	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (s *srv) serveHome(w http.ResponseWriter, r *http.Request) {
	if err := s.tmpls.ExecuteTemplate(w, "index.html", s.data(r)); err != nil {
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

func (s *srv) addToQueue(w http.ResponseWriter, r *http.Request) {
	s.queueAction(w, r, false /* remove */)
}

func (s *srv) removeFromQueue(w http.ResponseWriter, r *http.Request) {
	s.queueAction(w, r, true /* remove */)
}

func (s *srv) queueAction(w http.ResponseWriter, r *http.Request, remove bool) {
	w.Header().Set("Content-Type", "application/json")
	rm, err := s.getRoom(r)
	if err != nil {
		jsonErr(w, err)
		return
	}

	u, err := s.user(r)
	if err != nil {
		jsonErr(w, err)
		return
	}

	track, err := rm.SongServer.Track(r.FormValue("id"))
	if err != nil {
		jsonErr(w, err)
		return
	}

	q := u.Queue(rm.ID)
	if remove {
		q.RemoveTrack(track)
	} else {
		q.AddTrack(track)
	}

	json.NewEncoder(w).Encode(QueueResponse{})
}

func jsonErr(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(QueueResponse{
		Error:   true,
		Message: err.Error(),
	})
}

func (s *srv) serveQueue(w http.ResponseWriter, r *http.Request) {
	rm, err := s.getRoom(r)
	if err != nil {
		log.Printf("Couldn't load room: %v", err)
		return
	}

	u, err := s.user(r)
	if err != nil {
		log.Printf("Couldn't load user: %v", err)
		return
	}

	err = s.tmpls.ExecuteTemplate(w, "queue.html", struct {
		Queue *app.Queue
		Room  *app.Room
	}{u.Queue(rm.ID), rm})
	if err != nil {
		log.Printf("Failed to execute queue template: %v", err)
	}
}

func (s *srv) nowPlaying(w http.ResponseWriter, r *http.Request) {
	rm, err := s.getRoom(r)
	if err != nil {
		log.Printf("Couldn't load room: %v", err)
		return
	}

	_, t := rm.NowPlaying()
	err = s.tmpls.ExecuteTemplate(w, "playing.html", []music.Track{t})

	if err != nil {
		log.Printf("Failed to execute queue template: %v", err)
	}
}

func (s *srv) serveSong(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rm, err := s.getRoom(r)
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
	if c, ok := s.h.userconns[u]; ok {
		c.send <- []byte("pop")
	}
	s.h.broadcast <- []byte("playing")

	err = json.NewEncoder(w).Encode(TrackResponse{
		Track: t,
	})
}

func (s *srv) createRoom(w http.ResponseWriter, r *http.Request) {
	dispName := r.PostFormValue("room")
	id := app.Normalize(dispName)
	exists, err := s.db.HasRoom(id)
	if err != nil {
		log.Printf("Failed to check for room: %v", err)
		return
	}
	// If the room exists, take them to it
	if exists {
		http.Redirect(w, r, "/rooms/"+id, 302)
		return
	}

	// Add the new, non-existent room
	rm := app.New(dispName)
	switch r.PostFormValue("shuffle_order") {
	case "robin":
		rm.Rotator = app.RoundRobin()
	case "random":
		rm.Rotator = app.Shuffle()
	}

	if err := s.db.AddRoom(rm); err != nil {
		log.Printf("Failed to add room %+v: %v", rm, err)
		return
	}

	http.Redirect(w, r, "/rooms/"+id, 302)

}

func (s *srv) serveRoom(w http.ResponseWriter, r *http.Request) {
	rm, err := s.getRoom(r)
	id := roomID(r)
	if err != nil {
		log.Printf("No room found with ID %s", id)

		err := s.tmpls.ExecuteTemplate(w, "new_room.html", struct {
			DisplayName string
			ID          string
			Host        string
			Room        *app.Room
		}{mux.Vars(r)["id"], id, r.Host, nil})
		if err != nil {
			serveError(w, err)
		}
		return
	}

	u, err := s.user(r)
	if err != nil {
		serveError(w, err)
		return
	}

	q := u.Queue(id)
	if !rm.HasUser(u) {
		log.Printf("Adding user %s to room %s", u.ID, rm.ID)
		rm.AddUser(u)
	}

	_, t := rm.NowPlaying()

	err = s.tmpls.ExecuteTemplate(w, "room.html", struct {
		Room   *app.Room
		Queue  *app.Queue
		Tracks []music.Track
		Host   string
	}{rm, q, []music.Track{t}, r.Host})
	if err != nil {
		serveError(w, err)
	}
}

func (s *srv) data(r *http.Request) *tmplData {
	rm, err := s.getRoom(r)
	if err != nil {
		log.Printf("Failed to load room: %v", err)
	}

	user, err := s.user(r)
	if err != nil {
		log.Printf("Failed to load user: %v", err)
	}
	return &tmplData{
		Host: r.Host,
		Room: rm,
		User: user,
	}
}

func (s *srv) saveState(filename string) error {
	// TODO: Implement this. You'll probably have to recursively implement some
	// serialization for the nested/unexported fields
	return nil
}
