package main

import (
	"fmt"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"net/http"
)

var store *sessions.CookieStore

var hashKey []byte
var currentID = 0

func init() {
	hashKey, err := ioutil.ReadFile("hashKey")
	if err != nil {
		fmt.Println("Error reading cookie code:", err)
		panic(err)
	}
	store = sessions.NewCookieStore(hashKey)
}

func withLogin(hand func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "login")
		if session.IsNew {
			session.Values["login"] = currentID
			AddUser(currentID)
			currentID++
		} else {
			loginID := session.Values["login"].(int)
			// Their session is outdated, time to update it
			if _, ok := loginMap[loginID]; !ok {
				session.Values["login"] = currentID
				AddUser(currentID)
				currentID++
			}
		}
		session.Save(r, w)
		hand(w, r)
	}
}

func LoginID(r *http.Request) int {
	session, _ := store.Get(r, "login")
	return session.Values["login"].(int)
}

func AddUser(loginID int) {
	q := NewQueue(loginID)
	loginMap[loginID] = q
	queues = append(queues, q)
}
