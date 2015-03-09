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

func loginID(w http.ResponseWriter, r *http.Request) int {
	var data map[string]string

	cookie, err := r.Cookie("data")
	if err == nil {
		// They have a cookie and we can decode it
		err = s.Decode("data", cookie.Value, &data)
		id, _ := strconv.Atoi(data["id"])
		return id
	} else {
		// Assume they don't have a cookie, give them an ID
		data = map[string]string{
			"id": strconv.Itoa(currentID),
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
		id, _ := strconv.Atoi(data["id"])
		return id
	}
}
