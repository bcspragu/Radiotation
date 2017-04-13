package room

import "github.com/bcspragu/Radiotation/music"

type User struct {
	ID     string
	Queues map[string]*Queue
}

type Users []*User

func (u *User) AddQueue(name string) {
	if u.Queues[name] == nil {
		u.Queues[name] = &Queue{
			Tracks:   []music.Track{},
			TrackMap: make(map[string]music.Track),
		}
	}
}

func NewUser(id string) *User {
	return &User{
		ID:     id,
		Queues: make(map[string]*Queue),
	}
}
