package room

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/bcspragu/Radiotation/music"
)

func TestNormalize(t *testing.T) {
	testcases := []struct {
		desc string
		in   string
		want string
	}{
		{"empty name should return blank", "", "blank"},
		{"long name should be truncated", "thisnameiswaytoolong", "thisnameiswayto"},
		{"capital letters should be lower-cased", "YeLLiNG", "yelling"},
		{"dashes and spaces should become hyphens", "what_s goin-on", "what-s-goin-on"},
		{"other non-alphanumerics should be removed", "!@#$te st%^&*", "te-st"},
		{"more non-alphanumerics should be removed", "(){}te st:\"<>?", "te-st"},
	}

	for _, tc := range testcases {
		got := Normalize(tc.in)
		if got != tc.want {
			t.Errorf("%s: Normalize(%s) = %s, want %s", tc.desc, tc.in, got, tc.want)
		}
	}
}

func TestConstantRotator_Empty(t *testing.T) {
	r := &constantRotator{}

	r.start(0)
	if idx, last := r.nextIndex(); idx != 0 || !last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, true)
	}

	r.start(0)
	if idx, last := r.nextIndex(); idx != 0 || !last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, true)
	}
}

func TestConstantRotator(t *testing.T) {
	r := &constantRotator{}

	r.start(1)
	if idx, last := r.nextIndex(); idx != 0 || !last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, true)
	}

	r.start(2)
	if idx, last := r.nextIndex(); idx != 0 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, false)
	}
	if idx, last := r.nextIndex(); idx != 1 || !last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 1, true)
	}

	r.start(3)
	if idx, last := r.nextIndex(); idx != 0 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, false)
	}
	if idx, last := r.nextIndex(); idx != 1 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 1, false)
	}
	if idx, last := r.nextIndex(); idx != 2 || !last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 2, true)
	}

	r.start(4)
	if idx, last := r.nextIndex(); idx != 0 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, false)
	}
	if idx, last := r.nextIndex(); idx != 1 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 1, false)
	}
	// Restart in middle
	r.start(4)
	if idx, last := r.nextIndex(); idx != 0 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, false)
	}
	if idx, last := r.nextIndex(); idx != 1 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 1, false)
	}
	if idx, last := r.nextIndex(); idx != 2 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 2, false)
	}
	if idx, last := r.nextIndex(); idx != 3 || !last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 3, true)
	}
}

func TestShuffleRotator_Empty(t *testing.T) {
	r := &shuffleRotator{r: rand.New(rand.NewSource(0))}

	r.start(0)
	if idx, last := r.nextIndex(); idx != 0 || !last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, true)
	}

	r.start(0)
	if idx, last := r.nextIndex(); idx != 0 || !last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, true)
	}
}

func TestShuffleRotator(t *testing.T) {
	r := &shuffleRotator{r: rand.New(rand.NewSource(0))}
	r.start(1)
	if idx, last := r.nextIndex(); idx != 0 || !last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, true)
	}

	r = &shuffleRotator{r: rand.New(rand.NewSource(0))}
	r.start(2)
	if idx, last := r.nextIndex(); idx != 1 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 1, false)
	}
	if idx, last := r.nextIndex(); idx != 0 || !last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, true)
	}

	r = &shuffleRotator{r: rand.New(rand.NewSource(0))}
	r.start(3)
	if idx, last := r.nextIndex(); idx != 1 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 1, false)
	}
	if idx, last := r.nextIndex(); idx != 2 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 2, false)
	}
	if idx, last := r.nextIndex(); idx != 0 || !last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, true)
	}

	r = &shuffleRotator{r: rand.New(rand.NewSource(3))}
	r.start(4)
	if idx, last := r.nextIndex(); idx != 2 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 2, false)
	}
	if idx, last := r.nextIndex(); idx != 1 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 1, false)
	}
	// Restart in middle
	r = &shuffleRotator{r: rand.New(rand.NewSource(3))}
	r.start(4)
	if idx, last := r.nextIndex(); idx != 2 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, false)
	}
	if idx, last := r.nextIndex(); idx != 1 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 1, false)
	}
	if idx, last := r.nextIndex(); idx != 3 || last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 3, false)
	}
	if idx, last := r.nextIndex(); idx != 0 || !last {
		t.Errorf("r.nextIndex() = (%d, %t), want (%d, %t)", idx, last, 0, true)
	}
}

func TestPopTrack_ConstantRotator(t *testing.T) {
	r := &Room{
		ID:          "room",
		DisplayName: "Room",
		Rotator:     &constantRotator{},
		users:       []*User{},
		pending:     []*User{},
		m:           &sync.RWMutex{},
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
		Rotator:     &shuffleRotator{r: rand.New(rand.NewSource(0))},
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
		t.Errorf("diff between %d and %d = %f, want less than %f", u1, u2, diff, 0.2)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
