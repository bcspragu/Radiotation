package main

import (
	"fmt"
	"log"

	"github.com/bcspragu/Radiotation/db"
	"github.com/bcspragu/Radiotation/music"
)

type userTrack struct {
	id    db.UserID
	track music.Track
}

func (s *srv) nowPlaying(roomID string) music.Track {
	ts, err := s.db.History(roomID)
	if err != nil {
		log.Printf("Couldn't load history of tracks for room %s: %v", roomID, err)
	}

	if len(ts) > 0 {
		return ts[len(ts)-1]
	}
	return music.Track{}
}

func (s *srv) popTrack(roomID string) (*db.User, music.Track, error) {
	r, err := s.db.Room(roomID)
	if err != nil {
		return nil, music.Track{}, err
	}

	users, err := s.db.Users(roomID)
	if err != nil {
		return nil, music.Track{}, err
	}

	// Go through the queues, at most once each
	for i := 0; i < len(users); i++ {
		idx, last := r.Rotator.NextIndex()
		if last {
			// Start a rotation with any new users
			r.Rotator.Start(len(users))
		}

		if idx >= len(users) {
			return nil, music.Track{}, fmt.Errorf("Rotator is broken, returned index %d for list of %d users", idx, len(users))
		}

		u := users[idx]
		if u == nil {
			log.Printf("everything is broken, returned a nil user at index %d of %d", idx, len(users))
			continue
		}

		q, err := s.db.Queue(roomID, u.ID)
		if err != nil {
			log.Printf("error retreiving queue for user %s in room %s: %v", u.ID, roomID, err)
			continue
		}

		if !q.HasTracks() {
			continue
		}

		t := q.NextTrack()
		if err := s.db.AddToHistory(roomID, u.ID, t); err != nil {
			log.Printf("Failed to add track %v from user %s to history for room %s: %v", t, u.ID, roomID, err)
		}

		return u, t, nil
	}
	return nil, music.Track{}, errNoTracks
}

func (s *srv) AddUser(roomID string, id db.UserID) {
	r, err := s.db.Room(roomID)
	if err != nil {
		log.Printf("Error loading room %s: %v", roomID, err)
		return
	}

	users, err := s.db.Users(roomID)
	if err != nil {
		log.Printf("Error loading users in room %s: %v", roomID, err)
		return
	}

	for _, u := range users {
		if u.ID == id {
			log.Printf("User %s is already in room %s", id, roomID)
			return
		}
	}

	// If this is the first user, start the rotation
	if len(users) == 0 {
		r.Rotator.Start(1)
	}

	err = s.db.AddUserToRoom(roomID, id)
	if err != nil {
		log.Printf("Error adding user %s to room %s: %v", id, roomID, err)
		return
	}
}
