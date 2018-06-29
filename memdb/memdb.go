package memdb

import (
	"fmt"
	"sort"
	"sync"

	"github.com/bcspragu/Radiotation/db"
	"github.com/bcspragu/Radiotation/radio"
)

type Queue struct {
	ID     db.QueueID
	Offset int
	Tracks []radio.Track
}

func New() (*DB, error) {
	return &DB{
		rooms:   make(map[db.RoomID]*db.Room),
		users:   make(map[db.UserID]*db.User),
		queues:  make(map[db.RoomID][]*Queue),
		history: make(map[db.RoomID][]*db.TrackEntry),
	}, nil
}

type DB struct {
	sync.RWMutex
	// Map from roomID -> room
	rooms map[db.RoomID]*db.Room
	// Map from uid -> user
	users map[db.UserID]*db.User
	// Map from roomID -> list of queues
	queues map[db.RoomID][]*Queue
	// Map from roomID -> list of played track entries
	history map[db.RoomID][]*db.TrackEntry
}

func (m *DB) Room(id db.RoomID) (*db.Room, error) {
	m.RLock()
	defer m.RUnlock()
	r, ok := m.rooms[id]
	if !ok {
		return nil, db.ErrRoomNotFound
	}

	return r, nil
}

func (m *DB) NextUser(db.RoomID) (*db.User, radio.Track, error) {
	return nil, radio.Track{}, db.ErrOperationNotImplemented
}

func (m *DB) MarkVetoed(db.RoomID, db.UserID) error {
	return db.ErrOperationNotImplemented
}

func (m *DB) Rooms() ([]*db.Room, error) {
	m.RLock()
	defer m.RUnlock()
	var rooms []*db.Room
	for _, r := range m.rooms {
		rooms = append(rooms, r)
	}

	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i].DisplayName < rooms[j].DisplayName
	})

	return rooms, nil
}

func (m *DB) AddRoom(room *db.Room) error {
	m.Lock()
	defer m.Unlock()
	m.rooms[room.ID] = room
	return nil
}

func (m *DB) AddUserToRoom(rid db.RoomID, uid db.UserID) error {
	m.Lock()
	defer m.Unlock()
	qs := m.queues[rid]
	m.queues[rid] = append(qs, &Queue{
		ID:     db.QueueID{UserID: uid, RoomID: rid},
		Tracks: []radio.Track{},
	})
	return nil
}

func (m *DB) User(id db.UserID) (*db.User, error) {
	m.RLock()
	defer m.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return nil, db.ErrUserNotFound
	}

	return u, nil
}

func (m *DB) Users(rid db.RoomID) ([]*db.User, error) {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.rooms[rid]
	if !ok {
		return nil, db.ErrRoomNotFound
	}

	qs, ok := m.queues[rid]
	if !ok {
		return []*db.User{}, nil
	}

	var us []*db.User
	for _, q := range qs {
		if u, ok := m.users[q.ID.UserID]; ok {
			us = append(us, u)
		}
	}
	return us, nil
}

func (m *DB) AddUser(user *db.User) error {
	m.Lock()
	defer m.Unlock()
	m.users[user.ID] = user
	return nil
}

func (m *DB) Queue(id db.QueueID) (*Queue, error) {
	m.RLock()
	defer m.RUnlock()
	qs, ok := m.queues[id.RoomID]
	if !ok {
		return nil, db.ErrQueueNotFound
	}

	for _, q := range qs {
		if q.ID == id {
			return q, nil
		}
	}

	return nil, db.ErrQueueNotFound
}

func (m *DB) AddTrack(id db.QueueID, track radio.Track) error {
	m.Lock()
	defer m.Unlock()
	qs, ok := m.queues[id.RoomID]
	if !ok {
		return db.ErrQueueNotFound
	}

	for _, q := range qs {
		if q.ID == id {
			q.Tracks = append(q.Tracks, track)
			return nil
		}
	}

	return db.ErrQueueNotFound
}

func (m *DB) RemoveTrack(id db.QueueID, idx int) error {
	m.Lock()
	defer m.Unlock()
	qs, ok := m.queues[id.RoomID]
	if !ok {
		return db.ErrQueueNotFound
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

	return db.ErrQueueNotFound
}

func (m *DB) History(rid db.RoomID) ([]*db.TrackEntry, error) {
	m.RLock()
	defer m.RUnlock()
	tes, ok := m.history[rid]
	if !ok {
		return nil, db.ErrRoomNotFound
	}
	return tes, nil
}

func (m *DB) AddToHistory(rid db.RoomID, trackEntry *db.TrackEntry) error {
	m.Lock()
	defer m.Unlock()
	m.history[rid] = append(m.history[rid], trackEntry)
	return nil
}

func (m *DB) Close() error {
	return nil
}
