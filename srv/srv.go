package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bcspragu/Radiotation/app"
	"github.com/bcspragu/Radiotation/music"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

type Srv struct {
	sc *securecookie.SecureCookie
	h  hub
	r  *mux.Router

	roomDB db.RoomDB
	userDB db.UserDB
}

func New(rdb db.RoomDB, udb db.UserDB) (http.Handler, error) {
	sc, err := loadKeys()
	if err != nil {
		return nil, err
	}
	s := &srv{
		sc:     sc,
		roomDB: rdb,
		userDB: udb,
	}
	s.r = mux.NewRouter()
	s.r.HandleFunc("/", s.serveHome).Methods("GET")
	s.r.HandleFunc("/", s.serveNewRoomHome).Methods("POST")
	s.r.HandleFunc("/verifyToken", s.serveVerifyToken)
	s.r.HandleFunc("/rooms", s.withLogin(s.serveCreateRoom)).Methods("POST")
	s.r.HandleFunc("/rooms/{id}", s.withLogin(s.serveRoom)).Methods("GET")
	s.r.HandleFunc("/rooms/{id}/search", s.withLogin(s.serveSearch)).Methods("GET")
	s.r.HandleFunc("/rooms/{id}/queue", s.withLogin(s.serveQueue)).Methods("GET")
	s.r.HandleFunc("/rooms/{id}/now", s.withLogin(s.serveNowPlaying)).Methods("GET")
	s.r.HandleFunc("/rooms/{id}/add", s.withLogin(s.addToQueue)).Methods("POST")
	s.r.HandleFunc("/rooms/{id}/remove", s.withLogin(s.removeFromQueue)).Methods("POST")
	s.r.HandleFunc("/rooms/{id}/pop", s.serveSong).Methods("GET")
	s.r.HandleFunc("/ws", s.withLogin(s.serveData))
	return s, nil
}

func (s *Srv) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.r.ServeHTTP(w, r)
}

func (s *Srv) serveHome(w http.ResponseWriter, r *http.Request) {
	if err := s.ExecuteTemplate(w, "index.html", s.data(r)); err != nil {
		serveError(w, err)
	}
}

func (s *Srv) addToQueue(w http.ResponseWriter, r *http.Request) {
	s.queueAction(w, r, false /* remove */)
}

func (s *Srv) removeFromQueue(w http.ResponseWriter, r *http.Request) {
	s.queueAction(w, r, true /* remove */)
}

func (s *Srv) queueAction(w http.ResponseWriter, r *http.Request, remove bool) {
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

func (s *Srv) serveQueue(w http.ResponseWriter, r *http.Request) {
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

	err = s.ExecuteTemplate(w, "queue.html", struct {
		*tmplData
		Queue *app.Queue
	}{s.data(r), q})
	if err != nil {
		log.Printf("Failed to execute queue template: %v", err)
	}
}

func (s *Srv) serveNowPlaying(w http.ResponseWriter, r *http.Request) {
	rm, err := s.getRoom(r)
	if err != nil {
		log.Printf("Couldn't load room: %v", err)
		return
	}

	t := s.nowPlaying(rm.ID)

	err = s.ExecuteTemplate(w, "playing.html", []music.Track{t})
	if err != nil {
		log.Printf("Failed to execute queue template: %v", err)
	}
}

func (s *Srv) serveSong(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rm, err := s.getRoom(r)
	if err != nil {
		log.Printf("Couldn't load room: %v", err)
		return
	}

	u, t, err := s.popTrack(rm.ID)
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

func (s *Srv) serveCreateRoom(w http.ResponseWriter, r *http.Request) {
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

func (s *Srv) createRoom(name, rotator string) {
	log.Printf("Creating room %s, with shuffle order %s", name, rotator)

	// Add the new, non-existent room
	rm := app.NewRoom(name, app.Spotify)
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

func (s *Srv) serveNewRoom(w http.ResponseWriter, r *http.Request) {
	id := roomID(r)
	log.Printf("No room found with ID %s", id)

	err := s.ExecuteTemplate(w, "new_room.html", struct {
		*tmplData
		DisplayName string
		ID          string
	}{s.data(r), mux.Vars(r)["id"], id})
	if err != nil {
		serveError(w, err)
	}
	return
}

func (s *Srv) serveNewRoomHome(w http.ResponseWriter, r *http.Request) {
	if err := s.ExecuteTemplate(w, "room_form.html", nil); err != nil {
		serveError(w, err)
	}
}

func (s *Srv) serveRoom(w http.ResponseWriter, r *http.Request) {
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

	err = s.ExecuteTemplate(w, "room.html", struct {
		*tmplData
		Queue  *app.Queue
		Tracks []music.Track
	}{s.data(r), q, []music.Track{t}})
	if err != nil {
		serveError(w, err)
	}
}

func (s *Srv) serveSearch(w http.ResponseWriter, r *http.Request) {
	rm, err := s.getRoom(r)
	if err != nil {
		serveError(w, err)
		return
	}

	u, err := s.user(r)
	if err != nil {
		serveError(w, err)
		return
	}

	tracks, err := songServer(rm).Search(r.FormValue("search"))
	if err != nil {
		serveError(w, err)
		return
	}

	q, err := s.db.Queue(rm.ID, u.ID)
	if err != nil {
		serveError(w, err)
		return
	}

	err = s.ExecuteTemplate(w, "search.html", struct {
		Host   string
		Tracks []music.Track
		Queue  *app.Queue
		Room   *app.Room
	}{
		Host:   r.Host,
		Tracks: tracks,
		Queue:  q,
		Room:   rm,
	})
	if err != nil {
		serveError(w, err)
	}
}

func (s *Srv) data(r *http.Request) *tmplData {
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
		Rooms:    nil,
	}
}

func jsonErr(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(QueueResponse{
		Error:   true,
		Message: err.Error(),
	})
}

func loadKeys() (*securecookie.SecureCookie, error) {
	hashKey, err := loadOrGenKey("hashKey")
	if err != nil {
		return nil, err
	}

	blockKey, err := loadOrGenKey("blockKey")
	if err != nil {
		return nil, err
	}

	return securecookie.New(hashKey, blockKey), nil
}

func loadOrGenKey(name string) ([]byte, error) {
	f, err := ioutil.ReadFile(name)
	if err == nil {
		return f, nil
	}

	dat := securecookie.GenerateRandomKey(32)
	if dat == nil {
		return nil, errors.New("Failed to generate key")
	}

	err = ioutil.WriteFile(name, dat, 0777)
	if err != nil {
		return nil, errors.New("Error writing file")
	}
	return dat, nil
}
