package srv

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/NaySoftware/go-fcm"
	"github.com/bcspragu/Radiotation/db"
	"github.com/bcspragu/Radiotation/hub"
	"github.com/bcspragu/Radiotation/radio"
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
	sc         *securecookie.SecureCookie
	h          *hub.Hub
	mux        *mux.Router
	fcm        *fcm.FcmClient
	authClient *auth.Client
	cfg        *Config

	roomDB    db.RoomDB
	userDB    db.UserDB
	queueDB   db.QueueDB
	historyDB db.HistoryDB
}

type Config struct {
	ClientID   string
	SongServer radio.SongServer
	FCMKey     string
	AuthClient *auth.Client
}

// New returns an initialized server.
func New(sdb db.DB, cfg *Config) (*Srv, error) {
	sc, err := loadKeys()
	if err != nil {
		return nil, err
	}

	s := &Srv{
		sc:         sc,
		h:          hub.New(),
		fcm:        fcm.NewFcmClient(cfg.FCMKey),
		cfg:        cfg,
		authClient: cfg.AuthClient,
		roomDB:     sdb,
		userDB:     sdb,
		queueDB:    sdb,
		historyDB:  sdb,
	}

	s.mux = s.initMux()

	return s, nil
}

func (s *Srv) initMux() *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc("/api/user", s.serveUser).Methods("GET")
	m.HandleFunc("/api/search", s.serveRoomSearch).Methods("GET")
	// Verifying login and storing a cookie.
	m.HandleFunc("/api/verifyToken", s.serveVerifyToken).Methods("POST")
	// Load room information for a user.
	m.HandleFunc("/api/room/{id}", s.withRoomAndUser(s.serveRoom)).Methods("GET")
	// Search for a song.
	m.HandleFunc("/api/room/{id}/search", s.withRoomAndUser(s.serveSearch)).Methods("GET")

	// Get the next song. This should be a POST action, but its GET for
	// debugging.
	m.HandleFunc("/api/room/{id}/pop", s.serveSong).Methods("GET")
	m.HandleFunc("/api/room/{id}/veto", s.withRoomAndUser(s.serveVeto)).Methods("POST")

	// Create a room.
	m.HandleFunc("/api/room", s.serveCreateRoom).Methods("POST")
	// Add a song to a queue as the next song.
	m.HandleFunc("/api/room/{id}/addNext", s.withRoomAndUser(s.addToQueueNext)).Methods("POST")
	// Add a song to a queue at the end.
	m.HandleFunc("/api/room/{id}/addLast", s.withRoomAndUser(s.addToQueueLast)).Methods("POST")
	// Remove a song from a queue.
	m.HandleFunc("/api/room/{id}/remove", s.withRoomAndUser(s.removeFromQueue)).Methods("POST")

	// WebSocket handler for new songs.
	m.HandleFunc("/api/ws/room/{id}", s.serveData).Methods("GET")

	return m
}

func (s *Srv) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Srv) serveUser(w http.ResponseWriter, r *http.Request) {
	u, err := s.user(r)
	if err != nil {
		jsonErr(w, err)
		return
	}
	jsonResp(w, u)
}

func (s *Srv) addToQueueNext(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room) error {
	return s.addToQueue(w, r, u, rm, s.addNext)
}

func (s *Srv) addToQueueLast(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room) error {
	return s.addToQueue(w, r, u, rm, s.addLast)
}

func (s *Srv) addToQueue(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room, add func(db.QueueID, *radio.Track) error) error {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	track, err := s.track(req.ID)
	if err != nil {
		return err
	}

	if err := add(db.QueueID{RoomID: rm.ID, UserID: u.ID}, &track); err != nil {
		log.Println(err)
	}

	jsonResp(w, struct{ ID string }{req.ID})
	return nil
}

func (s *Srv) removeFromQueue(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room) error {
	var req struct {
		QueueTrackID string `json:"queueTrackID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	if err := s.queueDB.RemoveTrack(db.QueueID{RoomID: rm.ID, UserID: u.ID}, req.QueueTrackID); err != nil {
		log.Println(err)
	}

	jsonResp(w, struct{}{})
	return nil
}

func (s *Srv) queueAction(w http.ResponseWriter, r *http.Request, remove bool) {
}

func (s *Srv) serveSong(w http.ResponseWriter, r *http.Request) {
	contTkn := r.FormValue("continuationToken")
	if contTkn != "" {
		// TODO: Don't load a new song, get it from history.
	}

	rm, err := s.room(r)
	if err != nil {
		jsonErr(w, err)
		return
	}

	u, t, err := s.roomDB.NextTrack(rm.ID)
	if err == db.ErrNoTracksInQueue {
		jsonErr(w, errors.New("No tracks to choose from"))
		return
	} else if err != nil {
		jsonErr(w, err)
		return
	}

	idx, err := s.historyDB.AddToHistory(rm.ID, &db.TrackEntry{
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
		Error             bool
		Message           string
		Track             *radio.Track
		ContinuationToken string
	}

	ct, err := makeContinuationToken(&continuationToken{
		HistoryIndex: idx,
		RoomID:       rm.ID,
		UserID:       u.ID,
		TrackID:      t.ID,
	})
	if err != nil {
		log.Printf("Failed to generate continuation token: %v", err)
	}

	err = json.NewEncoder(w).Encode(trackResponse{
		Track:             t,
		ContinuationToken: ct,
	})
}

func (s *Srv) serveCreateRoom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DisplayName  string `json:"roomName"`
		ShuffleOrder string `json:"shuffleOrder"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, err)
		return
	}

	if req.DisplayName == "" {
		jsonErr(w, errors.New("No room name given"))
		return
	}

	room := &db.Room{
		DisplayName: req.DisplayName,
		RotatorType: rotatorTypeByName(req.ShuffleOrder),
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

type resultRoom struct {
	DisplayName string `json:"displayName"`
	RoomCode    string `json:"roomCode"`
	NumberUsers int    `json:"numberUsers"`
}

type roomInfo struct {
	Room  *db.Room         `json:"room"`
	Queue []*db.QueueTrack `json:"queue"`
	Track *radio.Track     `json:"track"`
}

type roomResp struct {
	// Whether this is a room, or search results. Will
	// either be 'room' or 'results'.
	Type string `json:"type"`

	// Only populated for 'results'.
	Results []resultRoom `json:"results"`

	// Only populated for 'room'.
	RoomInfo roomInfo `json:"roomInfo"`
}

func (s *Srv) serveRoomSearch(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("query")

	if q == "" {
		jsonErr(w, errors.New("No query given"))
		return
	}

	u, err := s.user(r)
	if err != nil {
		jsonErr(w, err)
		return
	}

	rm, err := s.roomDB.Room(db.RoomID(strings.ToUpper(q)))
	switch err {
	case nil:
		qts, err := s.queueDB.Tracks(db.QueueID{
			RoomID: rm.ID,
			UserID: u.ID,
		}, &db.QueueOptions{Type: db.AllTracks})

		if err == db.ErrQueueNotFound {
			if err := s.roomDB.AddUserToRoom(rm.ID, u.ID); err != nil {
				jsonErr(w, err)
				return
			}
		} else if err != nil {
			jsonErr(w, err)
			return
		}

		jsonResp(w, roomResp{
			Type: "room",
			RoomInfo: roomInfo{
				Room:  rm,
				Queue: qts,
				Track: s.nowPlaying(rm.ID),
			},
		})
		return
	case db.ErrRoomNotFound:
		// This is fine, just search for it.
	default:
		jsonErr(w, err)
		return
	}

	rooms, err := s.roomDB.SearchRooms(q)
	if err != nil {
		jsonErr(w, err)
		return
	}

	rms := make([]resultRoom, 0, len(rooms))
	for _, rm := range rooms {
		us, err := s.userDB.Users(rm.ID)
		if err != nil {
			log.Printf("Failed to get user list for room %q: %v", rm.ID, err)
		}
		rms = append(rms, resultRoom{
			DisplayName: rm.DisplayName,
			RoomCode:    string(rm.ID),
			NumberUsers: len(us),
		})
	}

	jsonResp(w, roomResp{
		Type:    "results",
		Results: rms,
	})
	return
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

	nu, t, err := s.roomDB.NextTrack(rm.ID)
	if err == db.ErrNoTracksInQueue {
		return errors.New("No tracks left in queue")
	} else if err != nil {
		return err
	}

	_, err = s.historyDB.AddToHistory(rm.ID, &db.TrackEntry{
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
	qts, err := s.queueDB.Tracks(db.QueueID{
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

	jsonResp(w, roomInfo{
		Room:  rm,
		Queue: qts,
		Track: s.nowPlaying(rm.ID),
	})
	return nil
}

func (s *Srv) addNext(qID db.QueueID, t *radio.Track) error {
	return s.addTrackAfter(qID, t, db.PlayedOnly)
}

func (s *Srv) addLast(qID db.QueueID, t *radio.Track) error {
	return s.addTrackAfter(qID, t, db.AllTracks)
}

func (s *Srv) addTrackAfter(qID db.QueueID, t *radio.Track, qot db.QueueType) error {
	qts, err := s.queueDB.Tracks(qID, &db.QueueOptions{Type: qot})
	if err != nil {
		return err
	}

	afterID := ""
	if len(qts) > 0 {
		afterID = qts[len(qts)-1].ID
	}

	return s.queueDB.AddTrack(qID, t, afterID)
}

func (s *Srv) serveSearch(w http.ResponseWriter, r *http.Request, u *db.User, rm *db.Room) error {
	q := r.FormValue("query")

	if q == "" {
		jsonResp(w, []interface{}{})
		return nil
	}

	qts, err := s.queueDB.Tracks(db.QueueID{
		RoomID: rm.ID,
		UserID: u.ID,
	}, &db.QueueOptions{Type: db.UnplayedOnly})
	if err != nil {
		return err
	}

	inQueue := make(map[string]bool)
	for _, qt := range qts {
		if !qt.Played {
			inQueue[qt.Track.ID] = true
		}
	}

	type trackInQueue struct {
		Track   radio.Track `json:"track"`
		InQueue bool        `json:"inQueue"`
	}

	tracks, err := s.search(q)
	if err != nil {
		return err
	}

	var tracksInQueue []*trackInQueue
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
		return ts[len(ts)-1].Track
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
	log.Printf("Creating user with ID %s", u.ID)
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

type continuationToken struct {
	HistoryIndex int
	RoomID       db.RoomID
	UserID       db.UserID
	TrackID      string
}

func parseContinuationToken(str string) (*continuationToken, error) {
	dat, err := base64.RawURLEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	var ct continuationToken
	if err := gob.NewDecoder(bytes.NewReader(dat)).Decode(&ct); err != nil {
		return nil, err
	}
	return &ct, nil
}

func makeContinuationToken(ct *continuationToken) (string, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(ct); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf.Bytes()), nil
}
