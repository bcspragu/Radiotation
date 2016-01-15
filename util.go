package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"room"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func withLogin(handler func(c Context)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// If they have the cookie, load it
		var user *room.User

		if cookie, err := r.Cookie("login"); err == nil {
			value := struct {
				ID string
			}{}
			if err = s.Decode("login", cookie.Value, &value); err != nil {
				serveError(w, err)
				return
			}
			// If we're here, we've decoded it
			if u, ok := users[value.ID]; ok {
				user = u
			}
		}

		// No valid user found, create one
		if user == nil {
			// If two people get the same ID randomly, I'll play Powerball more often
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
				http.Redirect(w, r, "/", http.StatusFound)
			} else {
				log.Println(err)
			}

			// We've written the user, we can persist them now
			user = room.NewUser(id)
			users[id] = user
		}

		c := NewContext(w, r)
		c.User = user
		roomName := mux.Vars(r)["key"]
		c.Room = rooms[roomName]
		if c.User != nil && c.Room != nil {
			c.Room.AddUser(c.User)
			c.Queue = c.User.Queues[roomName]
		}
		handler(c)
	}
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
	if f, err := ioutil.ReadFile(name); err != nil {
		if dat := securecookie.GenerateRandomKey(32); dat != nil {
			if err := ioutil.WriteFile(name, dat, 0777); err == nil {
				return dat, nil
			}
			return nil, errors.New("Error writing file")
		}
		return nil, errors.New("Failed to generate key")
	} else {
		return f, nil
	}
}

func servePaths() error {
	for _, dir := range []string{"js", "img", "css", "bower_components"} {
		http.Handle("/"+dir+"/", http.StripPrefix("/"+dir+"/", http.FileServer(http.Dir(dir))))
	}

	// Note, this is sketchy, because it means for a path a/b/file, you can access it at:
	// /a/b/file,
	// /b/file
	// Which isn't ideal, but it stems from the fact that we're inconsistent with
	// our usage, ie we use app/component/blahblah but want view#/ instead of
	// app/view#/
	// I have no intentions of fixing it #NORAGRETS
	err := filepath.Walk("app", func(path string, info os.FileInfo, err error) error {
		// Only serve app/ and first level subdirectories at the root
		if info.IsDir() && strings.Count(path, "/") < 2 {
			dirs := strings.Split(path, "/")
			dir := dirs[len(dirs)-1]
			http.Handle("/"+dir+"/", http.StripPrefix("/"+dir+"/", http.FileServer(http.Dir(path))))
		}

		return nil
	})
	return err
}

func loadTemplates() (RadioTemplate, error) {
	template := template.New("templates").Delims("[[", "]]")

	err := filepath.Walk("app", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			if _, err := template.ParseFiles(path); err != nil {
				return err
			}
		}

		return nil
	})

	return RadioTemplate{template}, err
}

func genName(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
