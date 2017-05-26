package room

import (
	"errors"
	"math/rand"
	"regexp"
	"strings"

	"github.com/bcspragu/Radiotation/music"
	"github.com/bcspragu/Radiotation/spotify"
)

var (
	nameRE = regexp.MustCompile(`[^a-zA-Z0-9\-]+`)
)

type Rotator interface {
	NextUser(*Room) *User
	LastIndex(*Room) int
}

type shuffleRotator struct {
	shuffleList []int
	index       int
}

func (s *shuffleRotator) getOrder(r *Room) []int {
	if s.shuffleList == nil || s.index >= len(s.shuffleList)-1 {
		s.shuffleList = rand.Perm(len(r.Users))
		s.index = 0
	}
	return s.shuffleList
}

func (s *shuffleRotator) LastIndex(r *Room) int {
	return len(r.Users)
}

func (s *shuffleRotator) NextUser(r *Room) *User {
	ind := s.getOrder(r)
	for i := 0; i < len(r.Users); i++ {
		user := r.Users[ind[i]]
		queue := user.Queues[r.ID]
		if queue.HasTracks() {
			s.index++
			return user
		}
	}
	return nil
}

func RoundRobin() Rotator {
	return &constantRotator{}
}

func Shuffle() Rotator {
	return &shuffleRotator{}
}

type constantRotator struct {
	offset int
}

func (c *constantRotator) LastIndex(r *Room) int {
	return (c.offset + len(r.Users) - 1) % len(r.Users)
}

func (c *constantRotator) NextUser(r *Room) *User {
	for i := 0; i < len(r.Users); i++ {
		user := r.Users[(i+c.offset)%len(r.Users)]
		queue := user.Queues[r.ID]
		if queue.HasTracks() {
			c.offset = (i + c.offset + 1) % len(r.Users)
			return user
		}
	}
	return nil
}

type Room struct {
	ID          string
	DisplayName string
	Users       Users
	Rotator     Rotator
	SongServer  music.SongServer
}

func New(name string) *Room {
	return &Room{
		DisplayName: name,
		ID:          Normalize(name),
		SongServer:  spotify.NewSongServer("api.spotify.com"),
		Users:       Users{},
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

func (r *Room) AddUser(user *User) error {
	for _, u := range r.Users {
		if u.ID == user.ID {
			return errors.New("user is already in room")
		}
	}

	// Add a user at the end of the queue
	if len(r.Users) > 0 {
		i := (r.Rotator.LastIndex(r) + len(r.Users) - 1) % len(r.Users)
		r.Users = append(r.Users, nil)
		copy(r.Users[i+1:], r.Users[i:])
		r.Users[i] = user
	} else {
		r.Users = append(r.Users, user)
	}

	// Add a queue for this room
	user.AddQueue(r.ID)
	return nil
}

func (r *Room) HasTracks() bool {
	for _, user := range r.Users {
		if user.Queues[r.ID].HasTracks() {
			return true
		}
	}
	return false
}

func (r *Room) PopTrack() (*User, music.Track) {
	u := r.Rotator.NextUser(r)
	if u != nil {
		if q := u.Queues[r.ID]; q.HasTracks() {
			return u, q.NextTrack()
		}
	}

	return nil, music.Track{}
}
