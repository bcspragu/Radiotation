package main

import (
	"errors"
	"sync"
	"time"

	"github.com/bcspragu/Radiotation/app"
	"github.com/bcspragu/Radiotation/music"
	"github.com/boltdb/bolt"
)

var (
	RoomBucket = []byte("Room")
	UserBucket = []byte("User")

	errOperationNotImplemented = errors.New("radiotation: operation not implemented")
	errUserNotFound            = errors.New("radiotation: user not found")
	errRoomNotFound            = errors.New("radiotation: room not found")
	errQueueNotFound           = errors.New("radiotation: queue not found")
)

type uqueue struct {
	id app.ID
	q  *app.Queue
}

type utrack struct {
	id app.ID
	t  music.Track
}

type db interface {
	Room(id string) (*app.Room, error)
	AddRoom(room *app.Room) error
	Users(roomID string) ([]*app.User, error)

	User(id app.ID) (*app.User, error)
	AddUser(user *app.User) error

	Queue(roomID string, userID app.ID) (*app.Queue, error)
	AddTrackToQueue(roomID string, userID app.ID, track music.Track) error
	RemoveTrackFromQueue(roomID string, userID app.ID, track music.Track) error

	AddUserToRoom(roomID string, userID app.ID) error

	AddToHistory(roomID string, userID app.ID, track music.Track) error
	History(roomID string) ([]music.Track, error)
}

type memDBImpl struct {
	sync.RWMutex
	// Map from roomID -> room
	rooms map[string]*app.Room
	// Map from uid -> user
	users map[app.ID]*app.User
	// Map from roomID -> list of (uid, queue) pairs
	queues map[string][]*uqueue
	// Map from roomID -> list of (uid, track) pairs
	history map[string][]*utrack
}

func (m *memDBImpl) Room(id string) (*app.Room, error) {
	m.RLock()
	defer m.RUnlock()
	r, ok := m.rooms[id]
	if !ok {
		return nil, errRoomNotFound
	}

	return r, nil
}

func (m *memDBImpl) AddRoom(room *app.Room) error {
	m.Lock()
	defer m.Unlock()
	m.rooms[room.ID] = room
	return nil
}

func (m *memDBImpl) Users(roomID string) ([]*app.User, error) {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.rooms[roomID]
	if !ok {
		return nil, errRoomNotFound
	}

	qs, ok := m.queues[roomID]
	if !ok {
		return []*app.User{}, nil
	}

	us := []*app.User{}
	for _, uq := range qs {
		us = append(us, m.users[uq.id])
	}
	return us, nil
}

func (m *memDBImpl) User(id app.ID) (*app.User, error) {
	m.RLock()
	defer m.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return nil, errUserNotFound
	}

	return u, nil
}

func (m *memDBImpl) AddUser(user *app.User) error {
	m.Lock()
	defer m.Unlock()
	m.users[user.ID] = user
	return nil
}

func (m *memDBImpl) Queue(roomID string, userID app.ID) (*app.Queue, error) {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.rooms[roomID]
	if !ok {
		return nil, errRoomNotFound
	}

	qs, ok := m.queues[roomID]
	if !ok {
		return nil, errQueueNotFound
	}

	for _, uq := range qs {
		if uq.id == userID {
			return uq.q, nil
		}
	}

	return nil, errQueueNotFound
}

func (m *memDBImpl) AddTrackToQueue(roomID string, userID app.ID, track music.Track) error {
	m.Lock()
	defer m.Unlock()
	qs, ok := m.queues[roomID]
	if !ok {
		return errQueueNotFound
	}

	for _, uq := range qs {
		if uq.id == userID {
			uq.q.Tracks = append(uq.q.Tracks, track)
			return nil
		}
	}

	return errQueueNotFound
}

func (m *memDBImpl) RemoveTrackFromQueue(roomID string, userID app.ID, track music.Track) error {
	m.Lock()
	defer m.Unlock()
	qs, ok := m.queues[roomID]
	if !ok {
		return errQueueNotFound
	}

	for _, uq := range qs {
		if uq.id == userID {
			for i, t := range uq.q.Tracks {
				if t.ID == track.ID && i >= uq.q.Offset {
					uq.q.Tracks = append(uq.q.Tracks[:i], uq.q.Tracks[i+1:]...)
					return nil
				}
			}
		}
	}

	return errQueueNotFound
}

func (m *memDBImpl) AddUserToRoom(roomID string, userID app.ID) error {
	m.Lock()
	defer m.Unlock()
	qs := m.queues[roomID]
	m.queues[roomID] = append(qs, &uqueue{id: userID, q: &app.Queue{}})
	return nil
}

func (m *memDBImpl) UserInRoom(roomID string, userID app.ID) (bool, error) {
	m.RLock()
	defer m.RUnlock()
	qs, ok := m.queues[roomID]
	if !ok {
		return false, errRoomNotFound
	}

	for _, uq := range qs {
		if uq.id == userID {
			return true, nil
		}
	}
	return false, nil
}

func (m *memDBImpl) AddToHistory(roomID string, userID app.ID, track music.Track) error {
	m.Lock()
	defer m.Unlock()
	m.history[roomID] = append(m.history[roomID], &utrack{id: userID, t: track})
	return nil
}

func (m *memDBImpl) History(roomID string) ([]music.Track, error) {
	m.RLock()
	defer m.RUnlock()
	ts := make([]music.Track, len(m.history[roomID]))
	for i, ut := range m.history[roomID] {
		ts[i] = ut.t
	}
	return ts, nil
}

type boltDBImpl struct {
	*bolt.DB
}

func initInMemDB() (db, error) {
	return &memDBImpl{
		rooms:   make(map[string]*app.Room),
		users:   make(map[app.ID]*app.User),
		queues:  make(map[string][]*uqueue),
		history: make(map[string][]*utrack),
	}, nil
}

func initBoltDB() (db, error) {
	bdb, err := bolt.Open("radiotation.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = bdb.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{RoomBucket, UserBucket} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}

		return nil
	})

	return &boltDBImpl{bdb}, err
}

func (b *boltDBImpl) Room(id string) (*app.Room, error) {
	return nil, errOperationNotImplemented
}

func (b *boltDBImpl) AddRoom(rm *app.Room) error {
	return errOperationNotImplemented
}

func (b *boltDBImpl) User(id app.ID) (*app.User, error) {
	return nil, errOperationNotImplemented
}

func (b *boltDBImpl) AddUser(user *app.User) error {
	return errOperationNotImplemented
}

func (b *boltDBImpl) Queue(roomID string, userID app.ID) (*app.Queue, error) {
	return nil, errOperationNotImplemented
}

func (b *boltDBImpl) AddTrackToQueue(roomID string, userID app.ID, track music.Track) error {
	return errOperationNotImplemented
}

func (b *boltDBImpl) RemoveTrackFromQueue(roomID string, userID app.ID, track music.Track) error {
	return errOperationNotImplemented
}

func (b *boltDBImpl) Users(roomID string) ([]*app.User, error) {
	return nil, errOperationNotImplemented
}

func (b *boltDBImpl) AddUserToRoom(roomID string, userID app.ID) error {
	return errOperationNotImplemented
}

func (b *boltDBImpl) AddToHistory(roomID string, userID app.ID, track music.Track) error {
	return errOperationNotImplemented
}

func (b *boltDBImpl) History(roomID string) ([]music.Track, error) {
	return []music.Track{}, errOperationNotImplemented
}
