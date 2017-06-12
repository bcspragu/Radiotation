package main

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/bcspragu/Radiotation/app"
	"github.com/bcspragu/Radiotation/music"
	"github.com/bcspragu/Radiotation/spotify"
	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/namsral/flag"
)

type tmplData struct {
	ClientID string
	Host     string
	Room     *app.Room
	User     *app.User
}

type srv struct {
	sync.RWMutex
	pendingUsers map[string][]app.ID

	db db

	tmpls *template.Template
	sc    *securecookie.SecureCookie
	h     hub
}

var (
	_             = flag.String(flag.DefaultConfigFlagname, "config", "path to config file")
	addr          = flag.String("addr", ":8000", "http service address")
	clientID      = flag.String("client_id", "", "The Google ClientID to use")
	spotifyClient = flag.String("spotify_client_id", "", "The client ID of the Spotify application")
	spotifySecret = flag.String("spotify_secret", "", "The secret of the Spotify application")

	spotifyServer  music.SongServer
	googleVerifier *oidc.IDTokenVerifier

	errNoTracks = errors.New("radiotation: no tracks in room")
)

func main() {
	rand.Seed(time.Now().Unix())
	flag.Parse()

	if *clientID == "" || *spotifyClient == "" || *spotifySecret == "" {
		log.Fatalf("Missing a required flag, all of  --client_id, --spotify_client_id, and --spotify_secret are required.")
	}

	spotifyServer = spotify.NewSongServer("spotify.com", *spotifyClient, *spotifySecret)

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

	//db, err := initBoltDB()
	db, err := initInMemDB()
	if err != nil {
		log.Fatalf("Failed to initialize datastore: %v", err)
	}

	s := &srv{
		pendingUsers: make(map[string][]app.ID),
		db:           db,
		tmpls:        tmpls,
		sc:           sc,
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
	r.HandleFunc("/rooms", s.withLogin(s.serveCreateRoom)).Methods("POST")
	r.HandleFunc("/rooms/{id}", s.withLogin(s.serveRoom)).Methods("GET")
	r.HandleFunc("/rooms/{id}/search", s.withLogin(s.serveSearch)).Methods("GET")
	r.HandleFunc("/rooms/{id}/queue", s.withLogin(s.serveQueue)).Methods("GET")
	r.HandleFunc("/rooms/{id}/now", s.withLogin(s.serveNowPlaying)).Methods("GET")
	r.HandleFunc("/rooms/{id}/add", s.withLogin(s.addToQueue)).Methods("POST")
	r.HandleFunc("/rooms/{id}/remove", s.withLogin(s.removeFromQueue)).Methods("POST")
	r.HandleFunc("/rooms/{id}/pop", s.serveSong).Methods("GET")
	r.HandleFunc("/ws", s.withLogin(s.serveData))

	http.Handle("/", r)

	if err := servePaths(); err != nil {
		log.Fatalf("Can't serve static assets: %v", err)
	}

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

	track, err := songServer(rm).Track(r.FormValue("id"))
	if err != nil {
		jsonErr(w, err)
		return
	}

	if remove {
		s.db.RemoveTrackFromQueue(rm.ID, u.ID, track)
	} else {
		s.db.AddTrackToQueue(rm.ID, u.ID, track)
	}

	json.NewEncoder(w).Encode(QueueResponse{})
}

func songServer(rm *app.Room) music.SongServer {
	switch rm.MusicService {
	case app.Spotify:
		return spotifyServer
	default:
		return nil
	}
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

	q, err := s.db.Queue(rm.ID, u.ID)
	if err != nil {
		log.Printf("Couldn't load queue: %v", err)
		return
	}

	err = s.tmpls.ExecuteTemplate(w, "queue.html", struct {
		*tmplData
		Queue *app.Queue
	}{s.data(r), q})
	if err != nil {
		log.Printf("Failed to execute queue template: %v", err)
	}
}

func (s *srv) serveNowPlaying(w http.ResponseWriter, r *http.Request) {
	rm, err := s.getRoom(r)
	if err != nil {
		log.Printf("Couldn't load room: %v", err)
		return
	}

	t := s.nowPlaying(rm.ID)

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

	u, t, err := s.PopTrack(rm.ID)
	if err == errNoTracks {
		jsonErr(w, errors.New("No tracks to choose from"))
		return
	} else if err != nil {
		log.Printf("Couldn't pop track: %v", err)
		return
	}

	err = s.db.AddToHistory(rm.ID, u.ID, t)
	if err != nil {
		log.Printf("Failed to add track to history for room %s, moving on: %v", rm.ID, err)
	}

	// Let the user know we're playing their track
	if c, ok := s.h.userconns[u]; ok {
		c.send <- []byte("pop")
	}
	s.h.broadcast <- []byte("playing")

	err = json.NewEncoder(w).Encode(TrackResponse{
		Track: t,
	})
}

func (s *srv) serveCreateRoom(w http.ResponseWriter, r *http.Request) {
	dispName := r.PostFormValue("room")
	id := app.Normalize(dispName)

	_, err := s.db.Room(id)
	if err == errRoomNotFound {
		s.createRoom(dispName, r.PostFormValue("shuffle_order"))
	} else if err != nil {
		log.Printf("Failed to check for room: %v", err)
		return
	}

	// If we're here, the room exists now
	http.Redirect(w, r, "/rooms/"+id, 302)
}

func (s *srv) createRoom(name, rotator string) {
	log.Printf("Creating room %s, with shuffle order %s", name, rotator)

	// Add the new, non-existent room
	rm := app.New(name, app.Spotify)
	switch rotator {
	case "robin":
		rm.Rotator = app.NewRotator(app.RoundRobin)
	case "shuffle":
		rm.Rotator = app.NewRotator(app.Shuffle)
	case "random":
		rm.Rotator = app.NewRotator(app.Random)
	}

	if err := s.db.AddRoom(rm); err != nil {
		log.Printf("Failed to add room %+v: %v", rm, err)
		return
	}
}

func (s *srv) serveNewRoom(w http.ResponseWriter, r *http.Request) {
	id := roomID(r)
	log.Printf("No room found with ID %s", id)

	err := s.tmpls.ExecuteTemplate(w, "new_room.html", struct {
		*tmplData
		DisplayName string
		ID          string
	}{s.data(r), mux.Vars(r)["id"], id})
	if err != nil {
		serveError(w, err)
	}
	return
}

func (s *srv) serveRoom(w http.ResponseWriter, r *http.Request) {
	rm, err := s.getRoom(r)
	if err != nil {
		s.serveNewRoom(w, r)
		return
	}

	u, err := s.user(r)
	if err != nil {
		serveError(w, err)
		return
	}

	q, err := s.db.Queue(rm.ID, u.ID)
	if err == errQueueNotFound {
		log.Printf("Adding user %s to room %s", u.ID, rm.ID)
		s.AddUser(rm.ID, u.ID)
	} else if err != nil {
		serveError(w, err)
		return
	}

	t := s.nowPlaying(rm.ID)

	err = s.tmpls.ExecuteTemplate(w, "room.html", struct {
		*tmplData
		Queue  *app.Queue
		Tracks []music.Track
	}{s.data(r), q, []music.Track{t}})
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
		ClientID: *clientID,
		Host:     r.Host,
		Room:     rm,
		User:     user,
	}
}
