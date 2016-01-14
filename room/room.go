package room

import (
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
	// Add a queue for this room
	user.AddQueue(r.Name)
	// TODO: Maybe don't append, insert somewhere, furthest?
	r.Users = append(r.Users, user)
}

func (r *Room) HasTracks() bool {
	for _, user := range r.Users {
		if user.Queues[r.Name].HasTracks() {
			return true
		}
	}
	return false
}

func (r *Room) PopTrack() (*Queue, spotify.Track) {
	c := 0
	for c < len(r.Users) {
		user := r.Users[(c+r.Offset)%len(r.Users)]
		queue := user.Queues[r.Name]
		if queue.HasTracks() {
			track := queue.NextTrack()
			r.Offset = (c + r.Offset + 1) % len(r.Users)
			return queue, track
		}
		c++
	}
	return nil, spotify.Track{}
}
