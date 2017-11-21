package srv

import (
	"fmt"
	"net/http"

	"github.com/bcspragu/Radiotation/db"
	"github.com/gorilla/mux"
)

type roomHandler func(http.ResponseWriter, *http.Request, *db.User, *db.Room) error

func (s *Srv) withRoomAndUser(rh roomHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := s.user(r)
		if err != nil {
			jsonErr(w, err)
			return
		}

		rm, err := s.room(r)
		if err != nil {
			jsonErr(w, err)
			return
		}

		if err := rh(w, r, u, rm); err != nil {
			jsonErr(w, err)
			return
		}
	}
}

func (s *Srv) user(r *http.Request) (*db.User, error) {
	cookie, err := r.Cookie("user")

	if err != nil {
		return nil, errNotLoggedIn
	}

	var u *db.User
	// If we can't decode their cookie, it means we've probably updated our hash
	// and block keys. Treat them as not logged in, the Google Sign-In will auto
	// log them in and handle the refresh.
	if err := s.sc.Decode("user", cookie.Value, &u); err != nil {
		return nil, errNotLoggedIn
	}

	u, err = s.userDB.User(u.ID)
	if err == db.ErrUserNotFound {
		return nil, errNotLoggedIn
	}

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user: %v", err)
	}

	return u, nil
}

func roomID(r *http.Request) db.RoomID {
	return db.Normalize(mux.Vars(r)["id"])
}

func (s *Srv) room(r *http.Request) (*db.Room, error) {
	id := roomID(r)
	rm, err := s.roomDB.Room(id)
	if err != nil {
		return nil, err
	}

	return rm, nil
}
