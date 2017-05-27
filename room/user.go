package room

import (
	"sync"

	"github.com/bcspragu/Radiotation/music"
)

type User struct {
	ID     string
	queues map[string]*Queue
	m      *sync.RWMutex
}

func (u *User) Queue(id string) *Queue {
	u.m.RLock()
	q := u.queues[id]
	if q != nil {
		u.m.RUnlock()
		return q
	}
	u.m.RUnlock()

	u.m.Lock()
	defer u.m.Unlock()
	u.queues[id] = &Queue{
		tracks:   []music.Track{},
		trackMap: make(map[string]music.Track),
	}
	return u.queues[id]
}

func NewUser(id string) *User {
	return &User{
		ID:     id,
		queues: make(map[string]*Queue),
		m:      &sync.RWMutex{},
	}
}
