package srv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/bcspragu/Radiotation/db"
	"github.com/bcspragu/Radiotation/hub"
	"github.com/bcspragu/Radiotation/music"
	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	errNoTracks    = errors.New("radiotation: no tracks in room")
	errNotLoggedIn = errors.New("radiotation: user not found")
)

type Srv struct {
	sc   *securecookie.SecureCookie
	h    *hub.Hub
	r    *mux.Router
	tmpl *template.Template
	cfg  *Config

	googleVerifier *oidc.IDTokenVerifier

	roomDB    db.RoomDB
	userDB    db.UserDB
	queueDB   db.QueueDB
	historyDB db.HistoryDB
}

type Config struct {
	ClientID    string
	SongServers map[db.MusicService]music.SongServer
	Dev         bool
}

// New returns an initialized server
func New(sdb db.DB, cfg *Config) (http.Handler, error) {
	sc, err := loadKeys()
	if err != nil {
		return nil, err
	}

	googleProvider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		log.Fatalf("Failed to get provider for Google: %v", err)
	}

	s := &Srv{
		sc:   sc,
		h:    hub.New(),
		tmpl: template.Must(template.ParseGlob("frontend/*.html")),
		cfg:  cfg,
		googleVerifier: googleProvider.Verifier(&oidc.Config{
			ClientID: cfg.ClientID,
		}),
		roomDB:    sdb,
		userDB:    sdb,
		queueDB:   sdb,
		historyDB: sdb,
	}

	s.initHandlers()

	return s, nil
}

func (s *Srv) initHandlers() {
	s.r = mux.NewRouter()
	s.r.HandleFunc("/", s.serveHome).Methods("GET")
	s.r.HandleFunc("/user", s.serveUser).Methods("GET")
	// Verifying login and storing a cookie.
	s.r.HandleFunc("/verifyToken", s.serveVerifyToken).Methods("POST")
	// Load room information for a user.
	s.r.HandleFunc("/room/{id}", s.withRoomAndUser(s.serveRoom)).Methods("GET")
	// Search for a song.
	s.r.HandleFunc("/room/{id}/search", s.withRoomAndUser(s.serveSearch)).Methods("GET")

	// Get the next song. This should be a POST action, but its GET for
	// debugging.
	s.r.HandleFunc("/room/{id}/pop", s.serveSong).Methods("GET")

	// Create a room.
	s.r.HandleFunc("/room", s.serveCreateRoom).Methods("POST")
	// Add a song to a queue.
	s.r.HandleFunc("/room/{id}/add", s.withRoomAndUser(s.addToQueue)).Methods("POST")
	// Remove a song from a queue.
	s.r.HandleFunc("/room/{id}/remove", s.withRoomAndUser(s.removeFromQueue)).Methods("POST")

	// WebSocket handler for new songs.
	s.r.HandleFunc("/room/{id}/ws", s.serveData)

	// Static asset serving
	s.r.PathPrefix("/assets/").
		Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("frontend/static/"))))
}

func (s *Srv) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.r.ServeHTTP(w, r)
}

func (s *Srv) serveHome(w http.ResponseWriter, r *http.Request) {
	js := template.HTML("/assets/app.js")
	ws := template.JSStr(fmt.Sprintf("wss://%s", r.Host))
	if s.cfg.Dev {
		js = template.HTML("//localhost:8081/app.js")
		ws = template.JSStr(fmt.Sprintf("ws://%s", r.Host))
	}
	if err := s.tmpl.ExecuteTemplate(w, "index.html", struct {
		ClientID      string
		JS            template.HTML
		WebSocketAddr template.JSStr
	}{s.cfg.ClientID, js, ws}); err != nil {
		serveError(w, err)
	}
}

func (s *Srv) serveUser(w http.ResponseWriter, r *http.Request) {
	u, err := s.user(r)
	if err != nil {
		jsonErr(w, err)
		return
	}
	jsonResp(w, u)
}

func (s *Srv) addToQueue(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room) error {
	trackID := r.FormValue("id")
	track, err := s.track(rm, trackID)
	if err != nil {
		return err
	}

	if err := s.queueDB.AddTrack(db.QueueID{RoomID: rm.ID, UserID: u.ID}, track); err != nil {
		log.Println(err)
	}

	jsonResp(w, struct{ ID string }{trackID})
	return nil
}

func (s *Srv) removeFromQueue(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room) error {
	idx, err := strconv.Atoi(r.FormValue("index"))
	if err != nil {
		return err
	}

	queue, err := s.queueDB.Queue(db.QueueID{RoomID: rm.ID, UserID: u.ID})
	if err != nil {
		return err
	}

	// If there are less tracks than the index, it's invalid.
	if len(queue.Tracks) <= idx {
		return fmt.Errorf("asked to remove track index %d, only have %d tracks", idx, len(queue.Tracks))
	}

	// If we're already passed the index, it's invalid.
	if idx < queue.Offset {
		return fmt.Errorf("asked to remove track index %d, we're passed that on index %d", idx, queue.Offset)
	}

	if err := s.queueDB.RemoveTrack(db.QueueID{RoomID: rm.ID, UserID: u.ID}, idx); err != nil {
		log.Println(err)
	}

	jsonResp(w, struct{}{})
	return nil
}

func (s *Srv) queueAction(w http.ResponseWriter, r *http.Request, remove bool) {
}

func (s *Srv) serveSong(w http.ResponseWriter, r *http.Request) {
	rm, err := s.room(r)
	if err != nil {
		jsonErr(w, err)
		return
	}

	u, t, err := s.popTrack(rm.ID)
	if err == errNoTracks {
		jsonErr(w, errors.New("No tracks to choose from"))
		return
	} else if err != nil {
		jsonErr(w, err)
		return
	}

	err = s.historyDB.AddToHistory(rm.ID, &db.TrackEntry{
		Track:  t,
		UserID: u.ID,
	})
	if err != nil {
		jsonErr(w, fmt.Errorf("failed to add track %v from user %s to history for room %s: %v", t, u.ID, rm.ID, err))
		return
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(t); err != nil {
		jsonErr(w, err)
		return
	}
	s.h.BroadcastRoom(buf.Bytes(), rm)

	type trackResponse struct {
		Error   bool
		Message string
		Track   music.Track
	}

	err = json.NewEncoder(w).Encode(trackResponse{
		Track: t,
	})
}

func (s *Srv) serveCreateRoom(w http.ResponseWriter, r *http.Request) {
	dispName := r.FormValue("roomName")
	if dispName == "" {
		jsonErr(w, errors.New("No room name given"))
		return
	}
	id := db.Normalize(dispName)

	_, err := s.roomDB.Room(id)
	if err != nil && err != db.ErrRoomNotFound {
		jsonErr(w, err)
		return
	}

	if err == db.ErrRoomNotFound {
		room := &db.Room{
			ID:           id,
			DisplayName:  dispName,
			Rotator:      rotatorByName(r.FormValue("shuffleOrder")),
			MusicService: musicServiceByName(r.FormValue("musicSource")),
		}

		if err := s.roomDB.AddRoom(room); err != nil {
			jsonErr(w, err)
			return
		}
	}

	jsonResp(w, struct{ ID string }{string(id)})
}

func rotatorByName(name string) db.Rotator {
	typ := db.RoundRobin
	switch name {
	case "robin":
		typ = db.RoundRobin
	case "shuffle":
		typ = db.Shuffle
	case "random":
		typ = db.Random
	}
	return db.NewRotator(typ)
}

func musicServiceByName(name string) db.MusicService {
	typ := db.Spotify
	switch name {
	case "spotify":
		typ = db.Spotify
	case "playmusic":
		typ = db.PlayMusic
	}
	return typ
}

func (s *Srv) serveRoom(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room) error {
	q, err := s.queueDB.Queue(db.QueueID{RoomID: rm.ID, UserID: u.ID})
	if err != nil && err != db.ErrQueueNotFound {
		return err
	}

	if err == db.ErrQueueNotFound {
		log.Printf("Adding user %s to room %s", u.ID, rm.ID)
		if q, err = s.AddUser(rm.ID, u.ID); err != nil {
			return err
		}
	}

	t := s.nowPlaying(rm.ID)

	jsonResp(w, struct {
		Room  *db.Room
		Queue []music.Track
		Track *music.Track
	}{rm, q.Tracks, t})
	return nil
}

func (s *Srv) serveSearch(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room) error {
	tracks, err := s.search(rm, r.FormValue("query"))
	if err != nil {
		return err
	}

	queue, err := s.queueDB.Queue(db.QueueID{RoomID: rm.ID, UserID: u.ID})
	if err != nil {
		return err
	}

	inQueue := make(map[string]int)
	for i, t := range queue.Tracks[queue.Offset:] {
		inQueue[t.ID] = i + queue.Offset
	}

	type trackInQueue struct {
		music.Track
		InQueue bool
		Index   int
	}

	tracksInQueue := []*trackInQueue{}
	for _, t := range tracks {
		idx, iq := inQueue[t.ID]
		tracksInQueue = append(tracksInQueue, &trackInQueue{
			Track:   t,
			InQueue: iq,
			Index:   idx,
		})
	}

	jsonResp(w, tracksInQueue)
	return nil
}

// serveData handles websocket requests from the peer trying to connect.
func (s *Srv) serveData(w http.ResponseWriter, r *http.Request) {
	rm, err := s.room(r)
	if err != nil {
		jsonErr(w, err)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		jsonErr(w, err)
		return
	}

	// Register this connection with a room, and start reading from it.
	s.h.Register(ws, rm)
}

func (s *Srv) nowPlaying(rid db.RoomID) *music.Track {
	ts, err := s.historyDB.History(rid)
	if err != nil {
		log.Printf("Couldn't load history of tracks for room %s: %v", rid, err)
	}

	if len(ts) > 0 {
		return &ts[len(ts)-1]
	}
	return nil
}

func (s *Srv) popTrack(rid db.RoomID) (*db.User, music.Track, error) {
	r, err := s.roomDB.Room(rid)
	if err != nil {
		return nil, music.Track{}, err
	}

	users, err := s.userDB.Users(rid)
	if err != nil {
		return nil, music.Track{}, err
	}

	// Go through the queues, at most once each
	for i := 0; i < len(users); i++ {
		idx, last := r.Rotator.NextIndex()
		if last {
			// Start a rotation with any new users
			r.Rotator.Start(len(users))
		}

		if idx >= len(users) {
			return nil, music.Track{}, fmt.Errorf("Rotator is broken, returned index %d for list of %d users", idx, len(users))
		}

		u := users[idx]
		if u == nil {
			log.Printf("everything is broken, returned a nil user at index %d of %d", idx, len(users))
			continue
		}

		q, err := s.queueDB.Queue(db.QueueID{RoomID: rid, UserID: u.ID})
		if err != nil {
			log.Printf("error retreiving queue for user %s in room %s: %v", u.ID, rid, err)
			continue
		}

		if !q.HasTracks() {
			continue
		}

		return u, q.NextTrack(), nil
	}
	return nil, music.Track{}, errNoTracks
}

func (s *Srv) AddUser(rid db.RoomID, id db.UserID) (*db.Queue, error) {
	r, err := s.roomDB.Room(rid)
	if err != nil {
		return nil, fmt.Errorf("error loading room %s: %v", rid, err)
	}

	users, err := s.userDB.Users(rid)
	if err != nil {
		return nil, fmt.Errorf("error loading users in room %s: %v", rid, err)
	}

	for _, u := range users {
		if u.ID == id {
			return nil, fmt.Errorf("user %s is already in room %s", id, rid)
		}
	}

	// If this is the first user, start the rotation
	if len(users) == 0 {
		r.Rotator.Start(1)
	}

	err = s.roomDB.AddUserToRoom(rid, id)
	if err != nil {
		return nil, fmt.Errorf("error adding user %s to room %s: %v", id, rid, err)
	}
	return s.queueDB.Queue(db.QueueID{RoomID: rid, UserID: id})
}

func (s *Srv) createUser(w http.ResponseWriter, u *db.User) {
	if encoded, err := s.sc.Encode("user", u); err == nil {
		cookie := &http.Cookie{
			Name:  "user",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	} else {
		log.Printf("Error encoding cookie: %v", err)
	}

	// We've written the user, we can persist them now
	log.Printf("Creating user with ID %s", u.ID.String())
	if err := s.userDB.AddUser(u); err != nil {
		log.Printf("Failed to add user %+v: %v", u, err)
	}
}

func serveError(w http.ResponseWriter, err error) {
	w.Write([]byte("Internal Server Error"))
	log.Printf("Error: %v\n", err)
}

func jsonErr(w http.ResponseWriter, err error) {
	log.Printf("Returning error to client: %v", err)
	json.NewEncoder(w).Encode(struct {
		Error        bool
		Message      string
		NotLoggedIn  bool
		RoomNotFound bool
	}{
		Error:        true,
		Message:      err.Error(),
		NotLoggedIn:  err == errNotLoggedIn,
		RoomNotFound: err == db.ErrRoomNotFound,
	})
}

func jsonResp(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("jsonResp: %v", err)
	}
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

func (s *Srv) search(rm *db.Room, query string) ([]music.Track, error) {
	ss := s.songServer(rm)
	return ss.Search(query)
}

func (s *Srv) track(rm *db.Room, id string) (music.Track, error) {
	ss := s.songServer(rm)
	return ss.Track(id)
}

func (s *Srv) songServer(rm *db.Room) music.SongServer {
	ss, ok := s.cfg.SongServers[rm.MusicService]
	if !ok {
		log.Printf("Couldn't find song server for room %+v", rm)
		return s.cfg.SongServers[db.Spotify]
	}
	return ss
}
