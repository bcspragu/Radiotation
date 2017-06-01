package app

import (
	"sync"

	"github.com/bcspragu/Radiotation/music"
)

type AccountType int

func (a AccountType) String() string {
	switch a {
	case GoogleAccount:
		return "Google"
	case FacebookAccount:
		return "Facebook"
	}
	return "Unknown"
}

const (
	GoogleAccount AccountType = iota
	FacebookAccount
)

type ID struct {
	AccountType AccountType
	ID          string
}

func (id ID) String() string {
	return id.AccountType.String() + id.ID
}

type User struct {
	ID          ID
	First, Last string
	queues      map[string]*Queue
	m           *sync.RWMutex
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

func GoogleUser(id, first, last string) *User {
	return newUser(ID{AccountType: GoogleAccount, ID: id}, first, last)
}

func newUser(id ID, first, last string) *User {
	return &User{
		ID:     id,
		First:  first,
		Last:   last,
		queues: make(map[string]*Queue),
		m:      &sync.RWMutex{},
	}
}
