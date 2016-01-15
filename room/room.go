package room

import (
	"math/rand"
	"spotify"
)

type Room struct {
	Name   string
	Offset int
	Users  Users
}

func New(name string) *Room {
	return &Room{
		Name:  name,
		Users: Users{},
	}
}

func (r *Room) AddUser(user *User) {

	for _, u := range r.Users {
		if u.ID == user.ID {
			return
		}
	}

	// Add a user at the end of the queue
	if len(r.Users) > 0 {
		i := (r.Offset + len(r.Users) - 1) % len(r.Users)
		r.Users = append(r.Users, nil)
		copy(r.Users[i+1:], r.Users[i:])
		r.Users[i] = user
	} else {
		r.Users = append(r.Users, user)
	}

	// Add a queue for this room
	user.AddQueue(r.Name)
}

func (r *Room) HasTracks() bool {
	for _, user := range r.Users {
		if user.Queues[r.Name].HasTracks() {
			return true
		}
	}
	return false
}

func (r *Room) PopTrack() (*User, spotify.Track) {
	c := 0
	ind := rand.Perm(len(r.Users))
	for c < len(r.Users) {
		user := r.Users[ind[c]]
		//user := r.Users[(c+r.Offset)%len(r.Users)]
		queue := user.Queues[r.Name]
		if queue.HasTracks() {
			track := queue.NextTrack()
			//r.Offset = (c + r.Offset + 1) % len(r.Users)
			return user, track
		}
		c++
	}
	return nil, spotify.Track{}
}
