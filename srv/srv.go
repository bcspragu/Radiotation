package srv

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bcspragu/Radiotation/db"
	"github.com/bcspragu/Radiotation/hub"
	"github.com/bcspragu/Radiotation/music"
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
	errNoTracks = errors.New("radiotation: no tracks in room")
)

type tmplData struct {
	ClientID string
	Host     string
	Room     *db.Room
	User     *db.User
	Rooms    struct {
		ID          string
		DisplayName string
	}
}

type Srv struct {
	sc *securecookie.SecureCookie
	h  *hub.Hub
	r  *mux.Router

	roomDB    db.RoomDB
	userDB    db.UserDB
	queueDB   db.QueueDB
	historyDB db.HistoryDB
}

func New(sdb db.DB, h *hub.Hub) (http.Handler, error) {
	sc, err := loadKeys()
	if err != nil {
		return nil, err
	}
	s := &srv{
		sc:        sc,
		h:         h,
		roomDB:    sdb.RoomDB,
		userDB:    sdb.UserDB,
		queueDB:   sdb.QueueDB,
		historyDB: sdb.HistoryDB,
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

	s.r.Handle("/dist/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/dist/"))))

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

	// TODO: Check for errors.
	if remove {
		s.queueDB.RemoveTrack(rm.ID, u.ID, track)
	} else {
		s.queueDB.AddTrack(db.QueueID{RoomID: rm.ID, UserID: u.ID})
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

	q, err := s.queueDB.Queue(db.QueueID{RoomID: rm.ID, UserID: u.ID})
	if err != nil {
		log.Printf("Couldn't load queue: %v", err)
		return
	}

	err = s.ExecuteTemplate(w, "queue.html", struct {
		*tmplData
		Queue *db.Queue
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

	err = s.historyDB.AddToHistory(rm.ID, &db.TrackEntry{
		Track:  t,
		UserID: u.ID,
	})
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
	id := db.Normalize(dispName)

	_, err := s.roomDB.Room(id)
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
	rm := db.NewRoom(name, db.Spotify)
	switch rotator {
	case "robin":
		rm.Rotator = db.NewRotator(db.RoundRobin)
	case "shuffle":
		rm.Rotator = db.NewRotator(db.Shuffle)
	case "random":
		rm.Rotator = db.NewRotator(db.Random)
	}

	if err := s.roomDB.AddRoom(rm); err != nil {
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

	q, err := s.queueDB.Queue(db.QueueID{RoomID: rm.ID, UserID: u.ID})
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
		Queue  *db.Queue
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

	q, err := s.queueDB.Queue(db.QueueID{RoomID: rm.ID, UserID: u.ID})
	if err != nil {
		serveError(w, err)
		return
	}

	err = s.ExecuteTemplate(w, "search.html", struct {
		Host   string
		Tracks []music.Track
		Queue  *db.Queue
		Room   *db.Room
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
	r, err := s.db.Room(rid)
	if err != nil {
		return nil, music.Track{}, err
	}

	users, err := s.db.Users(rid)
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

		q, err := s.db.Queue(rid, u.ID)
		if err != nil {
			log.Printf("error retreiving queue for user %s in room %s: %v", u.ID, rid, err)
			continue
		}

		if !q.HasTracks() {
			continue
		}

		t := q.NextTrack()
		if err := s.db.AddToHistory(rid, u.ID, t); err != nil {
			log.Printf("Failed to add track %v from user %s to history for room %s: %v", t, u.ID, rid, err)
		}

		return u, t, nil
	}
	return nil, music.Track{}, errNoTracks
}

func (s *Srv) AddUser(rid db.RoomID, id db.UserID) {
	r, err := s.db.Room(rid)
	if err != nil {
		log.Printf("Error loading room %s: %v", rid, err)
		return
	}

	users, err := s.db.Users(rid)
	if err != nil {
		log.Printf("Error loading users in room %s: %v", rid, err)
		return
	}

	for _, u := range users {
		if u.ID == id {
			log.Printf("User %s is already in room %s", id, rid)
			return
		}
	}

	// If this is the first user, start the rotation
	if len(users) == 0 {
		r.Rotator.Start(1)
	}

	err = s.db.AddUserToRoom(rid, id)
	if err != nil {
		log.Printf("Error adding user %s to room %s: %v", id, rid, err)
		return
	}
}

func (s *Srv) withLogin(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := s.user(r); err != nil {
			log.Printf("Unable to load user from request: %v --- Redirecting to login", err)
			http.Redirect(w, r, "/", 302)
			return
		}

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
	if err := s.db.AddUser(u); err != nil {
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
	rm, err := s.db.Room(id)
	if err != nil {
		return nil, fmt.Errorf("Error loading room with key %s: %v", id, err)
	}

	return rm, nil
}

func (s *Srv) user(r *http.Request) (*db.User, error) {
	cookie, err := r.Cookie("user")

	if err != nil {
		return nil, fmt.Errorf("Error loading cookie, or no cookie found: %v", err)
	}

	var u *db.User
	if err := s.sc.Decode("user", cookie.Value, &u); err != nil {
		return nil, fmt.Errorf("Error decoding cookie: %v", err)
	}

	u, err = s.db.User(u.ID)
	if err != nil {
		return nil, fmt.Errorf("User not found in system...probably: %v", err)
	}

	return u, nil
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
