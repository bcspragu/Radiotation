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

func withLogin(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := user(r); err != nil {
			log.Printf("Unable to load user from request: %v", err)
			createUser(w)
		}

		handler(w, r)
	}
}

func createUser(w http.ResponseWriter) {
	// If two people get the same ID randomly, I'll play Powerball more often

	// Future you checking in, was just wondering if we check for collisions.
	// Turns out we don't, but the above answer is very solid.
	id := genName(64)
	val := struct {
		ID string
	}{
		ID: id,
	}

	if encoded, err := s.Encode("login", val); err == nil {
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
	users[id] = room.NewUser(id)
}

func serveError(w http.ResponseWriter, err error) {
	w.Write([]byte("Internal Server Error"))
	log.Printf("Error: %v\n", err)
}

func loadKeys() error {
	var hashKey []byte
	var blockKey []byte

	if dat, err := loadOrGenKey("hashKey"); err != nil {
		return err
	} else {
		hashKey = dat
	}

	if dat, err := loadOrGenKey("blockKey"); err != nil {
		return err
	} else {
		blockKey = dat
	}

	s = securecookie.New(hashKey, blockKey)

	return nil
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

func getRoom(r *http.Request) (*room.Room, error) {
	id := roomID(r)
	rm, ok := rooms[id]
	if !ok {
		return nil, fmt.Errorf("No room found for key %s", id)
	}

	return rm, nil
}

func queue(r *http.Request) (*room.Queue, error) {
	id := roomID(r)
	user, err := user(r)
	if err != nil {
		return nil, fmt.Errorf("Error loading user: %v", err)
	}

	if _, ok := user.Queues[id]; !ok {
		user.AddQueue(id)
	}
	return user.Queues[id], nil
}

func user(r *http.Request) (*room.User, error) {
	cookie, err := r.Cookie("login")

	if err != nil {
		return nil, fmt.Errorf("Error loading cookie, or no cookie found: %v", err)
	}

	value := struct{ ID string }{}
	if err := s.Decode("login", cookie.Value, &value); err != nil {
		return nil, fmt.Errorf("Error decoding cookie: %v", err)
	}

	u, ok := users[value.ID]
	if !ok {
		return nil, fmt.Errorf("User not found in system")
	}

	return u, nil
}
