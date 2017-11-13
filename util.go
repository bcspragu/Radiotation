package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bcspragu/Radiotation/db"
	"github.com/gorilla/mux"
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

func (s *srv) createUser(w http.ResponseWriter, u *db.User) {
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

func servePaths() error {
	for _, dir := range []string{"js", "img", "css"} {
		http.Handle("/"+dir+"/", http.StripPrefix("/"+dir+"/", http.FileServer(http.Dir(dir))))
	}
	return nil
}

func roomID(r *http.Request) string {
	return db.Normalize(mux.Vars(r)["id"])
}

func (s *srv) getRoom(r *http.Request) (*db.Room, error) {
	id := roomID(r)
	rm, err := s.db.Room(id)
	if err != nil {
		return nil, fmt.Errorf("Error loading room with key %s: %v", id, err)
	}

	return rm, nil
}

func (s *srv) user(r *http.Request) (*db.User, error) {
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
