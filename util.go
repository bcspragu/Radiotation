package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"

	"github.com/bcspragu/Radiotation/room"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func (s *srv) withLogin(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := s.user(r); err != nil {
			log.Printf("Unable to load user from request: %v", err)
			s.createUser(w)
		}

		handler(w, r)
	}
}

func (s *srv) createUser(w http.ResponseWriter) {
	// If two people get the same ID randomly, I'll play Powerball more often

	// Future you checking in, was just wondering if we check for collisions.
	// Turns out we don't, but the above answer is very solid.
	id := genName(64)
	val := struct {
		ID string
	}{
		ID: id,
	}

	if encoded, err := s.sc.Encode("login", val); err == nil {
		cookie := &http.Cookie{
			Name:  "login",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	} else {
		log.Printf("Error encoding cookie: %v", err)
	}

	// We've written the user, we can persist them now
	log.Printf("Creating user with ID %s", id)
	s.um.Lock()
	s.users[id] = room.NewUser(id)
	s.um.Unlock()
}

func serveError(w http.ResponseWriter, err error) {
	w.Write([]byte("Internal Server Error"))
	log.Printf("Error: %v\n", err)
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

func servePaths() error {
	for _, dir := range []string{"js", "img", "css"} {
		http.Handle("/"+dir+"/", http.StripPrefix("/"+dir+"/", http.FileServer(http.Dir(dir))))
	}
	return nil
}

func genName(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func roomID(r *http.Request) string {
	return room.Normalize(mux.Vars(r)["id"])
}

func (s *srv) getRoom(r *http.Request) (*room.Room, error) {
	id := roomID(r)
	s.rm.RLock()
	rm, ok := s.rooms[id]
	s.rm.RUnlock()
	if !ok {
		return nil, fmt.Errorf("No room found for key %s", id)
	}

	return rm, nil
}

func (s *srv) user(r *http.Request) (*room.User, error) {
	cookie, err := r.Cookie("login")

	if err != nil {
		return nil, fmt.Errorf("Error loading cookie, or no cookie found: %v", err)
	}

	value := struct{ ID string }{}
	if err := s.sc.Decode("login", cookie.Value, &value); err != nil {
		return nil, fmt.Errorf("Error decoding cookie: %v", err)
	}

	s.um.RLock()
	u, ok := s.users[value.ID]
	s.um.RUnlock()
	if !ok {
		return nil, fmt.Errorf("User not found in system")
	}

	return u, nil
}
