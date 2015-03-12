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
		r.ParseForm()
		cookie, err := r.Cookie("data")
		if err == nil {
			var data map[string]string
			// They have a cookie and we can decode it
			err = s.Decode("data", cookie.Value, &data)

			loginID, _ := strconv.Atoi(data["login"])
			// If they have a cookie but we don't have that in our map, it means they
			// have an old cookie and we need to get them a new one
			if _, ok := loginMap[loginID]; !ok {
				AddUser(currentID)
				SetUserCookie(w, currentID)
				r.Form.Set("login", strconv.Itoa(currentID))
				currentID++
			} else {
				r.Form.Set("login", data["login"])
			}

		} else {
			// Assume they don't have a cookie, give them an ID
			AddUser(currentID)
			SetUserCookie(w, currentID)
			r.Form.Set("login", strconv.Itoa(currentID))
			currentID++

		}
		hand(w, r)
	}
}

func LoginID(r *http.Request) int {
	id, _ := strconv.Atoi(r.FormValue("login"))
	return id
}

func AddUser(loginID int) {
	q := NewQueue(loginID)
	loginMap[loginID] = q
}

func SetUserCookie(w http.ResponseWriter, loginID int) {
	// Our cookie is a map from "login" to a string of the user's loginID
	data := map[string]string{
		"login": strconv.Itoa(currentID),
	}

	// We encode the data
	encoded, _ := s.Encode("data", data)

	// We put the encoded data in a cookie
	cookie := &http.Cookie{
		Name:  "data",
		Value: encoded,
		Path:  "/",
	}

	// We send the cookie to the client
	http.SetCookie(w, cookie)
}
