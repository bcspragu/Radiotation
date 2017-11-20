package srv

import (
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
	s.r.HandleFunc("/verifyToken", s.serveVerifyToken)
	s.r.HandleFunc("/room", s.withLogin(s.serveCreateRoom)).Methods("POST")
	s.r.HandleFunc("/room/{id}", s.withLogin(s.serveRoom)).Methods("GET")
	s.r.HandleFunc("/room/{id}/search", s.withLogin(s.serveSearch)).Methods("GET")
	s.r.HandleFunc("/room/{id}/queue", s.withLogin(s.serveQueue)).Methods("GET")
	s.r.HandleFunc("/room/{id}/now", s.withLogin(s.serveNowPlaying)).Methods("GET")
	s.r.HandleFunc("/room/{id}/add", s.withLogin(s.addToQueue)).Methods("POST")
	s.r.HandleFunc("/room/{id}/remove", s.withLogin(s.removeFromQueue)).Methods("POST")
	s.r.HandleFunc("/room/{id}/pop", s.serveSong).Methods("GET")
	s.r.HandleFunc("/ws", s.withLogin(s.serveData))
	s.r.PathPrefix("/assets/").
		Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("frontend/static/"))))
}

func (s *Srv) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.r.ServeHTTP(w, r)
}

func (s *Srv) serveHome(w http.ResponseWriter, r *http.Request) {
	js := template.HTML("/assets/app.js")
	if s.cfg.Dev {
		js = template.HTML("//localhost:8081/app.js")
	}
	if err := s.tmpl.ExecuteTemplate(w, "index.html", struct {
		ClientID string
		JS       template.HTML
	}{s.cfg.ClientID, (js)}); err != nil {
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

func (s *Srv) addToQueue(w http.ResponseWriter, r *http.Request) {
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

	trackID := r.FormValue("id")
	track, err := s.track(rm, trackID)
	if err != nil {
		jsonErr(w, err)
		return
	}

	if err := s.queueDB.AddTrack(db.QueueID{RoomID: rm.ID, UserID: u.ID}, track); err != nil {
		log.Println(err)
	}

	jsonResp(w, struct{ ID string }{trackID})
}

func (s *Srv) removeFromQueue(w http.ResponseWriter, r *http.Request) {
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

	idx, err := strconv.Atoi(r.FormValue("index"))
	if err != nil {
		jsonErr(w, err)
		return
	}

	queue, err := s.queueDB.Queue(db.QueueID{RoomID: rm.ID, UserID: u.ID})
	if err != nil {
		jsonErr(w, err)
		return
	}

	// If there are less tracks than the index, it's invalid.
	if len(queue.Tracks) <= idx {
		jsonErr(w, fmt.Errorf("asked to remove track index %d, only have %d tracks", idx, len(queue.Tracks)))
		return
	}

	// If we're already passed the index, it's invalid.
	if idx < queue.Offset {
		jsonErr(w, fmt.Errorf("asked to remove track index %d, we're passed that on index %d", idx, queue.Offset))
		return
	}

	if err := s.queueDB.RemoveTrack(db.QueueID{RoomID: rm.ID, UserID: u.ID}, idx); err != nil {
		log.Println(err)
	}

	jsonResp(w, struct{}{})
}

func (s *Srv) queueAction(w http.ResponseWriter, r *http.Request, remove bool) {
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

	q, err := s.queueDB.Queue(db.QueueID{RoomID: rm.ID, UserID: u.ID})
	if err != nil {
		log.Printf("Couldn't load queue: %v", err)
		return
	}

	if err := json.NewEncoder(w).Encode(q); err != nil {
		serveError(w, err)
	}
}

func (s *Srv) serveNowPlaying(w http.ResponseWriter, r *http.Request) {
	rm, err := s.getRoom(r)
	if err != nil {
		log.Printf("Couldn't load room: %v", err)
		return
	}

	t := s.nowPlaying(rm.ID)

	if err := json.NewEncoder(w).Encode(t); err != nil {
		serveError(w, err)
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

	err = s.historyDB.AddToHistory(rm.ID, &db.TrackEntry{
		Track:  t,
		UserID: u.ID,
	})
	if err != nil {
		log.Printf("Failed to add track to history for room %s, moving on: %v", rm.ID, err)
	}

	// Let the user know we're playing their track
	//if c, ok := s.h.userconns[u]; ok {
	// TODO: Send more info, so the user can render stuff appropriately.
	s.h.Broadcast([]byte("pop"))
	//}
	s.h.Broadcast([]byte("playing"))

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

func (s *Srv) serveRoom(w http.ResponseWriter, r *http.Request) {
	rm, err := s.getRoom(r)
	if err != nil {
		jsonErr(w, errors.New("room not found"))
		return
	}

	u, err := s.user(r)
	if err != nil {
		jsonErr(w, err)
		return
	}

	q, err := s.queueDB.Queue(db.QueueID{RoomID: rm.ID, UserID: u.ID})
	if err != nil && err != db.ErrQueueNotFound {
		jsonErr(w, err)
		return
	}

	if err == db.ErrQueueNotFound {
		log.Printf("Adding user %s to room %s", u.ID, rm.ID)
		if q, err = s.AddUser(rm.ID, u.ID); err != nil {
			jsonErr(w, err)
			return
		}
	}

	t := s.nowPlaying(rm.ID)

	if err := json.NewEncoder(w).Encode(struct {
		Room  *db.Room
		Queue []music.Track
		Track music.Track
	}{rm, q.Tracks, t}); err != nil {
		serveError(w, err)
	}
}

func (s *Srv) serveSearch(w http.ResponseWriter, r *http.Request) {
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

	tracks, err := s.search(rm, r.FormValue("query"))
	if err != nil {
		jsonErr(w, err)
		return
	}

	queue, err := s.queueDB.Queue(db.QueueID{RoomID: rm.ID, UserID: u.ID})
	if err != nil {
		jsonErr(w, err)
		return
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

	var tracksInQueue []*trackInQueue
	for _, t := range tracks {
		idx, iq := inQueue[t.ID]
		tracksInQueue = append(tracksInQueue, &trackInQueue{
			Track:   t,
			InQueue: iq,
			Index:   idx,
		})
	}

	jsonResp(w, tracksInQueue)
}

// serveData handles websocket requests from the peer trying to connect.
func (s *Srv) serveData(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		serveError(w, err)
		return
	}

	s.h.Register(ws)
}

func (s *Srv) nowPlaying(rid db.RoomID) music.Track {
	ts, err := s.historyDB.History(rid)
	if err != nil {
		log.Printf("Couldn't load history of tracks for room %s: %v", rid, err)
	}

	if len(ts) > 0 {
		return ts[len(ts)-1]
	}
	return music.Track{}
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

		t := q.NextTrack()
		if err := s.historyDB.AddToHistory(rid, &db.TrackEntry{
			UserID: u.ID,
			Track:  t,
		}); err != nil {
			log.Printf("Failed to add track %v from user %s to history for room %s: %v", t, u.ID, rid, err)
		}

		return u, t, nil
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

func (s *Srv) withLogin(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Reintroduce login check
		handler(w, r)
	}
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

func roomID(r *http.Request) db.RoomID {
	return db.Normalize(mux.Vars(r)["id"])
}

func (s *Srv) getRoom(r *http.Request) (*db.Room, error) {
	id := roomID(r)
	rm, err := s.roomDB.Room(id)
	if err != nil {
		return nil, fmt.Errorf("Error loading room with key %s: %v", id, err)
	}

	return rm, nil
}

func (s *Srv) user(r *http.Request) (*db.User, error) {
	cookie, err := r.Cookie("user")

	if err != nil {
		return nil, errNotLoggedIn
	}

	var u *db.User
	if err := s.sc.Decode("user", cookie.Value, &u); err != nil {
		return nil, errNotLoggedIn
	}

	u, err = s.userDB.User(u.ID)
	if err == db.ErrUserNotFound {
		return nil, errNotLoggedIn
	}

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user: %v", err)
	}

	return u, nil
}

func jsonErr(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(struct {
		Error       bool
		Message     string
		NotLoggedIn bool
	}{
		Error:       true,
		Message:     err.Error(),
		NotLoggedIn: err == errNotLoggedIn,
	})
}

func jsonResp(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		serveError(w, err)
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
