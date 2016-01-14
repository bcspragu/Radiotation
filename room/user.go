package room

import (
	"spotify"
)

type User struct {
	ID     string
	Queues map[string]*Queue
}

type Users []*User

func (u *User) AddQueue(name string) {
	u.Queues[name] = &Queue{
		Tracks:   []spotify.Track{},
		TrackMap: make(map[string]spotify.Track),
	}
}

func NewUser(id string) *User {
	return &User{
		ID:     id,
		Queues: make(map[string]*Queue),
	}
}
