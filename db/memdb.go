package db

import (
	"encoding/gob"
	"fmt"
	"io"
	"sort"
	"sync"

	"github.com/bcspragu/Radiotation/music"
)

func InitInMemDB() (DB, error) {
	return &memDBImpl{
		rooms:   make(map[RoomID]*Room),
		users:   make(map[UserID]*User),
		queues:  make(map[RoomID][]*Queue),
		history: make(map[RoomID][]*TrackEntry),
	}, nil
}

type memDBImpl struct {
	sync.RWMutex
	// Map from roomID -> room
	rooms map[RoomID]*Room
	// Map from uid -> user
	users map[UserID]*User
	// Map from roomID -> list of queues
	queues map[RoomID][]*Queue
	// Map from roomID -> list of played track entries
	history map[RoomID][]*TrackEntry
}

type memData struct {
	Rooms   map[RoomID]*Room
	Users   map[UserID]*User
	Queues  map[RoomID][]*Queue
	History map[RoomID][]*TrackEntry
}

func Load(r io.Reader, idb DB) error {
	m, ok := idb.(*memDBImpl)
	if !ok {
		return fmt.Errorf("Cannot load into %T", idb)
	}

	md := &memData{}
	if err := gob.NewDecoder(r).Decode(md); err != nil {
		return err
	}

	m.rooms = md.Rooms
	m.users = md.Users
	m.queues = md.Queues
	m.history = md.History
	return nil
}

func Save(w io.Writer, idb DB) error {
	m, ok := idb.(*memDBImpl)
	if !ok {
		return fmt.Errorf("Cannot load into %T", idb)
	}

	m.Lock()
	defer m.Unlock()
	md := &memData{
		Rooms:   m.rooms,
		Users:   m.users,
		Queues:  m.queues,
		History: m.history,
	}
	if err := gob.NewEncoder(w).Encode(md); err != nil {
		return err
	}

	return nil
}

func (m *memDBImpl) Room(id RoomID) (*Room, error) {
	m.RLock()
	defer m.RUnlock()
	r, ok := m.rooms[id]
	if !ok {
		return nil, ErrRoomNotFound
	}

	return r, nil
}

func (m *memDBImpl) NextTrack(RoomID) (*User, music.Track, error) {
	return nil, music.Track{}, ErrOperationNotImplemented
}

func (m *memDBImpl) Rooms() ([]*Room, error) {
	m.RLock()
	defer m.RUnlock()
	var rooms []*Room
	for _, r := range m.rooms {
		rooms = append(rooms, r)
	}

	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i].DisplayName < rooms[j].DisplayName
	})

	return rooms, nil
}

func (m *memDBImpl) AddRoom(room *Room) error {
	m.Lock()
	defer m.Unlock()
	m.rooms[room.ID] = room
	return nil
}

func (m *memDBImpl) AddUserToRoom(rid RoomID, uid UserID) error {
	m.Lock()
	defer m.Unlock()
	qs := m.queues[rid]
	m.queues[rid] = append(qs, &Queue{
		ID:     QueueID{UserID: uid, RoomID: rid},
		Tracks: []music.Track{},
	})
	return nil
}

func (m *memDBImpl) User(id UserID) (*User, error) {
	m.RLock()
	defer m.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}

	return u, nil
}

func (m *memDBImpl) Users(rid RoomID) ([]*User, error) {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.rooms[rid]
	if !ok {
		return nil, ErrRoomNotFound
	}

	qs, ok := m.queues[rid]
	if !ok {
		return []*User{}, nil
	}

	var us []*User
	for _, q := range qs {
		if u, ok := m.users[q.ID.UserID]; ok {
			us = append(us, u)
		}
	}
	return us, nil
}

func (m *memDBImpl) AddUser(user *User) error {
	m.Lock()
	defer m.Unlock()
	m.users[user.ID] = user
	return nil
}

func (m *memDBImpl) Queue(id QueueID) (*Queue, error) {
	m.RLock()
	defer m.RUnlock()
	qs, ok := m.queues[id.RoomID]
	if !ok {
		return nil, ErrQueueNotFound
	}

	for _, q := range qs {
		if q.ID == id {
			return q, nil
		}
	}

	return nil, ErrQueueNotFound
}

func (m *memDBImpl) AddTrack(id QueueID, track music.Track) error {
	m.Lock()
	defer m.Unlock()
	qs, ok := m.queues[id.RoomID]
	if !ok {
		return ErrQueueNotFound
	}

	for _, q := range qs {
		if q.ID == id {
			q.Tracks = append(q.Tracks, track)
			return nil
		}
	}

	return ErrQueueNotFound
}

func (m *memDBImpl) RemoveTrack(id QueueID, idx int) error {
	m.Lock()
	defer m.Unlock()
	qs, ok := m.queues[id.RoomID]
	if !ok {
		return ErrQueueNotFound
	}

	for _, q := range qs {
		if q.ID == id {
			if idx >= len(q.Tracks) {
				return fmt.Errorf("asked to remove track index %d, only have %d tracks", idx, len(q.Tracks))
			}
			if idx < q.Offset {
				return fmt.Errorf("asked to remove track index %d, we're passed that on index %d", idx, q.Offset)
			}
			q.Tracks = append(q.Tracks[:idx], q.Tracks[idx+1:]...)
			return nil
		}
	}

	return ErrQueueNotFound
}

func (m *memDBImpl) History(rid RoomID) ([]*TrackEntry, error) {
	m.RLock()
	defer m.RUnlock()
	tes, ok := m.history[rid]
	if !ok {
		return nil, ErrRoomNotFound
	}
	return tes, nil
}

func (m *memDBImpl) AddToHistory(rid RoomID, trackEntry *TrackEntry) error {
	m.Lock()
	defer m.Unlock()
	m.history[rid] = append(m.history[rid], trackEntry)
	return nil
}

func (m *memDBImpl) Close() error {
	return nil
}
