package room

import (
	"log"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bcspragu/Radiotation/music"
	"github.com/bcspragu/Radiotation/spotify"
)

var (
	nameRE = regexp.MustCompile(`[^a-zA-Z0-9\-]+`)
)

// A Rotator says what the next index should be from a list of size n. They can
// assume that n will never decrease, but it can increase between any two
// invocations.
type Rotator interface {
	// nextIndex returns the next index in the current rotation, and if this is
	// the last entry in the current rotation
	nextIndex() (int, bool)
	// start takes the size of the current rotation and should be called before
	// we do our first rotation, and after each subsequent rotation is over
	start(n int)
}

type shuffleRotator struct {
	shuffleList []int
	index       int
	r           *rand.Rand
}

func (s *shuffleRotator) nextIndex() (int, bool) {
	if len(s.shuffleList) == 0 {
		return 0, true
	}

	if s.index >= len(s.shuffleList) {
		// Keep returning the last element if we've gone too far
		return s.shuffleList[len(s.shuffleList)-1], true
	}

	i := s.shuffleList[s.index]
	s.index++
	return i, s.index == len(s.shuffleList)
}

func (s *shuffleRotator) start(n int) {
	s.shuffleList = s.r.Perm(n)
	s.index = 0
}

type constantRotator struct {
	offset int
	n      int
}

func (c *constantRotator) nextIndex() (int, bool) {
	if c.n == 0 {
		return 0, true
	}

	i := c.offset
	c.offset = (c.offset + 1) % c.n
	return i, c.offset == 0
}

func (c *constantRotator) start(n int) {
	c.offset = 0
	c.n = n
}

func RoundRobin() Rotator {
	return &constantRotator{}
}

func Shuffle() Rotator {
	return &shuffleRotator{r: rand.New(rand.NewSource(time.Now().Unix()))}
}

type UserTrack struct {
	user  *User
	track music.Track
}

type Room struct {
	ID          string
	DisplayName string
	Rotator     Rotator
	SongServer  music.SongServer

	users   []*User
	pending []*User
	history []UserTrack
	m       *sync.RWMutex
}

func New(name string) *Room {
	return &Room{
		DisplayName: name,
		ID:          Normalize(name),
		SongServer:  spotify.NewSongServer("api.spotify.com"),
		users:       []*User{},
		pending:     []*User{},
		history:     []UserTrack{},
		m:           &sync.RWMutex{},
	}
}

func Normalize(name string) string {
	if len(name) == 0 {
		name = "blank"
	}

	if len(name) > 15 {
		name = name[:15]
	}
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)
	name = strings.Replace(name, " ", "-", -1)
	name = strings.Replace(name, "_", "-", -1)
	return nameRE.ReplaceAllString(name, "")
}

func (r *Room) HasUser(user *User) bool {
	r.m.Lock()
	defer r.m.Unlock()
	for _, u := range r.users {
		if u.ID == user.ID {
			return true
		}
	}
	return false
}

func (r *Room) AddUser(user *User) {
	r.m.Lock()
	defer r.m.Unlock()
	for _, u := range r.users {
		if u.ID == user.ID {
			log.Printf("User %s is already in room %s", user.ID, r.ID)
			return
		}
	}

	// Add a user at the end of the queue.
	if len(r.users) == 0 {
		r.users = append(r.users, user)
		r.Rotator.start(len(r.users))
		return
	}

	// Add the user to the end of our pending queue, we'll add them in once we
	// finish our current rotation.
	r.pending = append(r.pending, user)
}

func (r *Room) HasTracks() bool {
	for _, u := range r.users {
		if u.Queue(r.ID).HasTracks() {
			return true
		}
	}
	return false
}

func (r *Room) NowPlaying() (*User, music.Track) {
	r.m.RLock()
	defer r.m.RUnlock()
	if len(r.history) == 0 {
		return nil, music.Track{}
	}
	ut := r.history[len(r.history)-1]
	return ut.user, ut.track
}

func (r *Room) PopTrack() (*User, music.Track) {
	r.m.Lock()
	defer r.m.Unlock()
	for i := 0; i < len(r.users); i++ {
		idx, last := r.Rotator.nextIndex()
		if last {
			r.addPending()
			r.Rotator.start(len(r.users))
		}

		if idx >= len(r.users) {
			log.Printf("Rotator is broken, returned index %d for list of %d users", idx, len(r.users))
			return nil, music.Track{}
		}

		u := r.users[idx]
		if u == nil {
			continue
		}

		q := u.Queue(r.ID)
		if !q.HasTracks() {
			continue
		}

		t := q.NextTrack()
		r.history = append(r.history, UserTrack{
			user:  u,
			track: t,
		})

		return u, t
	}
	return nil, music.Track{}
}

func (r *Room) addPending() {
	r.users = append(r.users, r.pending...)
	r.pending = []*User{}
}
