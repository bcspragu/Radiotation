package main

import (
	"fmt"
	"github.com/gorilla/securecookie"
	"io/ioutil"
	"net/http"
	"strconv"
)

var hashKey []byte
var blockKey []byte
var s *securecookie.SecureCookie
var currentID = 0

func init() {
	hashKey, err := ioutil.ReadFile("blockKey")
	if err != nil {
		fmt.Println("Error reading cookie code:", err)
		panic(err)
	}
	blockKey, err := ioutil.ReadFile("hashKey")
	if err != nil {
		fmt.Println("Error reading cookie code:", err)
		panic(err)
	}
	s = securecookie.New(hashKey, blockKey)
}

func withLogin(hand func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data map[string]string

		r.ParseForm()
		cookie, err := r.Cookie("data")
		if err == nil {
			// They have a cookie and we can decode it
			err = s.Decode("data", cookie.Value, &data)
			r.Form.Set("login", data["login"])
			fmt.Println(data)
			hand(w, r)
		} else {
			// Assume they don't have a cookie, give them an ID
			data = map[string]string{
				"login": strconv.Itoa(currentID),
			}

			q := NewQueue(currentID)
			idToList[currentID] = q
			currentID++

			encoded, _ := s.Encode("data", data)
			cookie := &http.Cookie{
				Name:  "data",
				Value: encoded,
				Path:  "/",
			}
			http.SetCookie(w, cookie)
			r.Form.Set("login", data["login"])
			fmt.Println(data)
			hand(w, r)
		}
	}
}
