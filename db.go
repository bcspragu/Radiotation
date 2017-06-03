package main

import (
	"errors"
	"sync"
	"time"

	"github.com/bcspragu/Radiotation/app"
	"github.com/boltdb/bolt"
)

var (
	RoomBucket = []byte("Room")
	UserBucket = []byte("User")

	errOperationNotImplemented = errors.New("radiotation: operation not implemented")
	errUserNotFound            = errors.New("radiotation: user not found")
	errRoomNotFound            = errors.New("radiotation: room not found")
)

type db interface {
	Room(id string) (*app.Room, error)
	AddRoom(room *app.Room) error

	User(id app.ID) (*app.User, error)
	AddUser(user *app.User) error
}

type memDBImpl struct {
	sync.RWMutex
	rooms map[string]*app.Room
	users map[app.ID]*app.User
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

type boltDBImpl struct {
	*bolt.DB
}

func initInMemDB() (db, error) {
	return &memDBImpl{
		rooms: make(map[string]*app.Room),
		users: make(map[app.ID]*app.User),
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
