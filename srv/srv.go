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

	"github.com/NaySoftware/go-fcm"
	"github.com/bcspragu/Radiotation/db"
	"github.com/bcspragu/Radiotation/hub"
	"github.com/bcspragu/Radiotation/radio"
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
	errNotLoggedIn = errors.New("radiotation: user not found")
)

type Srv struct {
	sc   *securecookie.SecureCookie
	h    *hub.Hub
	r    *mux.Router
	tmpl *template.Template
	fcm  *fcm.FcmClient
	cfg  *Config

	googleVerifier *oidc.IDTokenVerifier

	roomDB    db.RoomDB
	userDB    db.UserDB
	queueDB   db.QueueDB
	historyDB db.HistoryDB
}

type Config struct {
	ClientID     string
	SongServer   radio.SongServer
	Dev          bool
	FCMKey       string
	FrontendGlob string
	StaticDir    string
}

// New returns an initialized server
func New(sdb db.DB, cfg *Config) (*Srv, error) {
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
		tmpl: template.Must(template.ParseGlob(cfg.FrontendGlob)),
		fcm:  fcm.NewFcmClient(cfg.FCMKey),
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
	s.r.HandleFunc("/room/{id}/veto", s.withRoomAndUser(s.serveVeto)).Methods("POST")

	// Create a room.
	s.r.HandleFunc("/room", s.serveCreateRoom).Methods("POST")
	// Add a song to a queue.
	s.r.HandleFunc("/room/{id}/add", s.withRoomAndUser(s.addToQueue)).Methods("POST")
	// Remove a song from a queue.
	s.r.HandleFunc("/room/{id}/remove", s.withRoomAndUser(s.removeFromQueue)).Methods("POST")

	// WebSocket handler for new songs.
	s.r.HandleFunc("/ws/room/{id}", s.serveData)

	// Static asset serving
	s.r.PathPrefix("/static/").
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(s.cfg.StaticDir))))
}

func (s *Srv) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.r.ServeHTTP(w, r)
}

func (s *Srv) serveHome(w http.ResponseWriter, r *http.Request) {
	ws := template.JSStr(fmt.Sprintf("wss://%s", r.Host))
	var js template.HTML
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
	afterQTID := r.FormValue("afterQTID")

	track, err := s.track(trackID)
	if err != nil {
		return err
	}

	if err := s.queueDB.AddTrack(db.QueueID{RoomID: rm.ID, UserID: u.ID}, track, afterQTID); err != nil {
		log.Println(err)
	}

	jsonResp(w, struct{ ID string }{trackID})
	return nil
}

func (s *Srv) removeFromQueue(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room) error {
	qtID := r.FormValue("queueTrackID")

	if err := s.queueDB.RemoveTrack(db.QueueID{RoomID: rm.ID, UserID: u.ID}, qtID); err != nil {
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

	u, tID, err := s.roomDB.NextTrack(rm.ID)
	if err == db.ErrNoTracksInQueue {
		jsonErr(w, errors.New("No tracks to choose from"))
		return
	} else if err != nil {
		jsonErr(w, err)
		return
	}

	t, err := s.track(tID)
	if err != nil {
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
		Track   radio.Track
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

	room := &db.Room{
		DisplayName: dispName,
		RotatorType: rotatorTypeByName(r.FormValue("shuffleOrder")),
	}

	rID, err := s.roomDB.AddRoom(room)
	if err != nil {
		jsonErr(w, err)
		return
	}

	jsonResp(w, struct{ ID string }{string(rID)})
}

func rotatorTypeByName(name string) db.RotatorType {
	typ := db.RoundRobin
	switch name {
	case "robin":
		typ = db.RoundRobin
	case "shuffle":
		typ = db.Shuffle
	case "random":
		typ = db.Random
	}
	return typ
}

func (s *Srv) serveRoomSearch(w http.ResponseWriter, r *http.Request) error {
	q := r.FormValue("query")
	if q == "" {
		return errors.New("No query given")
	}

	rooms, err := s.roomDB.SearchRooms(q)
	if err != nil {
		jsonErr(w, err)
		return nil
	}

	jsonResp(w, rooms)
	return nil
}

func (s *Srv) serveVeto(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room) error {
	return errors.New("sorry, vetoing not implemented yet")

	users, err := s.userDB.Users(rm.ID)
	if err != nil {
		return err
	}

	hist, err := s.historyDB.History(rm.ID)
	if err != nil {
		return err
	}
	if len(hist) == 0 {
		return errors.New("no tracks in history")
	}

	songsSince, vetoed := lastVeto(hist, u.ID)
	if vetoed && songsSince < 2*len(users) {
		return fmt.Errorf("Can only veto once every %d songs, you vetoed %d songs ago", 2*len(users), songsSince)
	}

	if err := s.historyDB.MarkVetoed(rm.ID, u.ID); err != nil {
		return err
	}

	nu, tID, err := s.roomDB.NextTrack(rm.ID)
	if err == db.ErrNoTracksInQueue {
		return errors.New("No tracks left in queue")
	} else if err != nil {
		return err
	}

	t, err := s.track(tID)
	if err != nil {
		return err
	}

	err = s.historyDB.AddToHistory(rm.ID, &db.TrackEntry{
		Track:  t,
		UserID: nu.ID,
	})
	if err != nil {
		log.Printf("failed to add track %v from user %s to history for room %s: %v", t, u.ID, rm.ID, err)
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(t); err != nil {
		return err
	}
	s.h.BroadcastRoom(buf.Bytes(), rm)

	vetoee, err := s.userDB.User(hist[len(hist)-1].UserID)
	if err != nil {
		return err
	}

	if err := s.pushVeto(rm, u, vetoee); err != nil {
		return err
	}

	jsonResp(w, struct{}{})
	return nil
}

func (s *Srv) pushVeto(rm *db.Room, vetoer, vetoee *db.User) error {
	s.fcm.NewFcmMsgTo(string(rm.ID), struct {
		Vetoer *db.User
		Vetoee *db.User
	}{vetoer, vetoee})

	status, err := s.fcm.Send()
	if err != nil {
		return fmt.Errorf("error sending FCM: %v", err)
	}

	log.Printf("Veto status: %+v", status)
	return nil
}

func lastVeto(history []*db.TrackEntry, uid db.UserID) (songsSince int, veto bool) {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Vetoed && history[i].VetoedBy == uid {
			veto = true
			return
		}
		songsSince++
	}
	return
}

func (s *Srv) serveRoom(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room) error {
	tl, err := s.queueDB.TrackList(db.QueueID{
		RoomID: rm.ID,
		UserID: u.ID,
	}, &db.QueueOptions{Type: db.AllTracks})

	if err == db.ErrQueueNotFound {
		if err := s.roomDB.AddUserToRoom(rm.ID, u.ID); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	type trackWithPlayed struct {
		radio.Track
		Played bool
	}

	tracksWithPlayed := []*trackWithPlayed{}
	for i, t := range tl.Tracks {
		tracksWithPlayed = append(tracksWithPlayed, &trackWithPlayed{
			Track:  t,
			Played: i < tl.NextIndex,
		})
	}

	jsonResp(w, struct {
		Room  *db.Room
		Queue []*trackWithPlayed
		Track *radio.Track
	}{rm, tracksWithPlayed, s.nowPlaying(rm.ID)})
	return nil
}

func (s *Srv) serveSearch(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room) error {
	tracks, err := s.search(r.FormValue("query"))
	if err != nil {
		return err
	}

	tl, err := s.queueDB.TrackList(db.QueueID{
		RoomID: rm.ID,
		UserID: u.ID,
	}, &db.QueueOptions{Type: db.PlayedOnly})
	if err != nil {
		return err
	}

	inQueue := make(map[string]bool)
	for _, t := range tl.Tracks {
		inQueue[t.ID] = true
	}

	type trackInQueue struct {
		radio.Track
		InQueue bool
	}

	tracksInQueue := []*trackInQueue{}
	for _, t := range tracks {
		tracksInQueue = append(tracksInQueue, &trackInQueue{
			Track:   t,
			InQueue: inQueue[t.ID],
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

func (s *Srv) nowPlaying(rid db.RoomID) *radio.Track {
	ts, err := s.historyDB.History(rid)
	if err != nil {
		log.Printf("Couldn't load history of tracks for room %s: %v", rid, err)
	}

	if len(ts) > 0 {
		return &ts[len(ts)-1].Track
	}
	return nil
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

func (s *Srv) search(query string) ([]radio.Track, error) {
	return s.cfg.SongServer.Search(query)
}

func (s *Srv) track(id string) (radio.Track, error) {
	return s.cfg.SongServer.Track(id)
}
