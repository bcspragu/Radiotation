package srv

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/bcspragu/Radiotation/music"
)

func TestPopTrack_RoundRobinRotator(t *testing.T) {
	r := &Room{
		ID:          "room",
		DisplayName: "Room",
		Rotator:     &roundRobinRotator{},
	}
	n := 4

	// Add 4 users to the queue
	for i := 0; i < n; i++ {
		u := NewUser(strconv.Itoa(i))
		q := u.Queue("room")
		// Add i+1 songs to this user's queue for this room
		for j := 0; j < i+1; j++ {
			q.AddTrack(music.Track{ID: strconv.Itoa(j)})
		}

		r.AddUser(u)
	}

	// First rotation, exhaust User 0s only song
	if u, tr := r.PopTrack(); u.ID != "0" || tr.ID != "0" {
		t.Errorf("r.PopTrack = (%s, %s), want (%s, %s)", u.ID, tr.ID, "0", "0")
	}
	if u, tr := r.PopTrack(); u.ID != "1" || tr.ID != "0" {
		t.Errorf("r.PopTrack = (%s, %s), want (%s, %s)", u.ID, tr.ID, "1", "0")
	}
	if u, tr := r.PopTrack(); u.ID != "2" || tr.ID != "0" {
		t.Errorf("r.PopTrack = (%s, %s), want (%s, %s)", u.ID, tr.ID, "2", "0")
	}
	if u, tr := r.PopTrack(); u.ID != "3" || tr.ID != "0" {
		t.Errorf("r.PopTrack = (%s, %s), want (%s, %s)", u.ID, tr.ID, "3", "0")
	}

	// Second rotation, exhaust User 1s last song
	if u, tr := r.PopTrack(); u.ID != "1" || tr.ID != "1" {
		t.Errorf("r.PopTrack = (%s, %s), want (%s, %s)", u.ID, tr.ID, "1", "1")
	}
	if u, tr := r.PopTrack(); u.ID != "2" || tr.ID != "1" {
		t.Errorf("r.PopTrack = (%s, %s), want (%s, %s)", u.ID, tr.ID, "2", "1")
	}
	if u, tr := r.PopTrack(); u.ID != "3" || tr.ID != "1" {
		t.Errorf("r.PopTrack = (%s, %s), want (%s, %s)", u.ID, tr.ID, "3", "1")
	}

	// Third rotation, exhaust User 2s last song
	if u, tr := r.PopTrack(); u.ID != "2" || tr.ID != "2" {
		t.Errorf("r.PopTrack = (%s, %s), want (%s, %s)", u.ID, tr.ID, "2", "2")
	}
	if u, tr := r.PopTrack(); u.ID != "3" || tr.ID != "2" {
		t.Errorf("r.PopTrack = (%s, %s), want (%s, %s)", u.ID, tr.ID, "3", "2")
	}

	// Last rotation, exhaust User 3s last song
	if u, tr := r.PopTrack(); u.ID != "3" || tr.ID != "3" {
		t.Errorf("r.PopTrack = (%s, %s), want (%s, %s)", u.ID, tr.ID, "3", "3")
	}

	// Make sure it returns garbage when we run out
	if u, tr := r.PopTrack(); u != nil || tr.ID != "" {
		t.Errorf("r.PopTrack = (%v, %v), want (%s, %s)", u, tr, nil, "empty track")
	}
}

func TestPopTrack_ShuffleRotator(t *testing.T) {
	r := &Room{
		ID:          "room",
		DisplayName: "Room",
		Rotator:     &shuffleRotator{R: rand.New(rand.NewSource(0))},
		users:       []*User{},
		pending:     []*User{},
		m:           &sync.RWMutex{},
	}
	// Add 2 users to the queue, ID "1" and ID "2"
	for i := 1; i <= 2; i++ {
		u := NewUser(strconv.Itoa(i))
		q := u.Queue("room")
		// Add 1,000 songs to each user's queue for this room
		for j := 0; j < 1000; j++ {
			q.AddTrack(music.Track{ID: strconv.Itoa(j)})
		}

		r.AddUser(u)
	}

	var u1, u2 float64

	// Pop 1,000 tracks
	for i := 0; i < 1000; i++ {
		u, _ := r.PopTrack()
		switch u.ID {
		case "1":
			u1++
		case "2":
			u2++
		}
	}

	if diff := abs(u1-u2) / ((u1 + u2) / 2); diff >= 0.05 {
		t.Errorf("diff between %f and %f = %f, want less than %f", u1, u2, diff, 0.2)
	}
}
