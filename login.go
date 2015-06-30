package main

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

func init() {
	store = sessions.NewCookieStore([]byte("Trust me, this is secure"))
}

func withLogin(hand func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "login")
		if session.IsNew {
			// Create a new session
			session.Values["login"] = genName(20)
		}
		session.Save(r, w)
		hand(w, r)
	}
}

func userID(r *http.Request) string {
	session, _ := store.Get(r, "login")
	s, _ := session.Values["login"].(string)
	return s
}
