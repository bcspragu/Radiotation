package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bcspragu/Radiotation/app"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

func (s *srv) withLogin(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := s.user(r); err != nil {
			log.Printf("Unable to load user from request: %v --- Redirecting to login", err)
			http.Redirect(w, r, "/", 302)
			return
		}

		handler(w, r)
	}
}

func (s *srv) createUser(w http.ResponseWriter, u *app.User) {
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

func roomID(r *http.Request) string {
	return app.Normalize(mux.Vars(r)["id"])
}

func (s *srv) getRoom(r *http.Request) (*app.Room, error) {
	id := roomID(r)
	rm, err := s.db.Room(id)
	if err != nil {
		return nil, fmt.Errorf("Error loading room with key %s: %v", id, err)
	}

	return rm, nil
}

func (s *srv) user(r *http.Request) (*app.User, error) {
	cookie, err := r.Cookie("user")

	if err != nil {
		return nil, fmt.Errorf("Error loading cookie, or no cookie found: %v", err)
	}

	var u *app.User
	if err := s.sc.Decode("user", cookie.Value, &u); err != nil {
		return nil, fmt.Errorf("Error decoding cookie: %v", err)
	}

	u, err = s.db.User(u.ID)
	if err != nil {
		return nil, fmt.Errorf("User not found in system...probably: %v", err)
	}

	return u, nil
}
