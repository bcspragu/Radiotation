package room

import (
	"errors"
	"math/rand"

	"github.com/bcspragu/Radiotation/music"
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
		queue := user.Queues[r.Name]
		if queue.HasTracks() {
			s.index++
			return user
		}
	}
	return nil
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
		queue := user.Queues[r.Name]
		if queue.HasTracks() {
			c.offset = (i + c.offset + 1) % len(r.Users)
			return user
		}
	}
	return nil
}

type Room struct {
	Name       string
	Users      Users
	Rotator    Rotator
	SongServer music.SongServer
}

func New(name string) *Room {
	return &Room{
		Name:  name,
		Users: Users{},
	}
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
	user.AddQueue(r.Name)
	return nil
}

func (r *Room) HasTracks() bool {
	for _, user := range r.Users {
		if user.Queues[r.Name].HasTracks() {
			return true
		}
	}
	return false
}

func (r *Room) PopTrack() (*User, music.Track) {
	u := r.Rotator.NextUser(r)
	if u != nil {
		if q := u.Queues[r.Name]; q.HasTracks() {
			return u, q.NextTrack()
		}
	}

	return nil, music.Track{}
}
