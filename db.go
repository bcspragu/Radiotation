package main

import (
	"time"

	"github.com/bcspragu/Radiotation/app"
	"github.com/boltdb/bolt"
)

var (
	RoomBucket = []byte("Room")
	UserBucket = []byte("User")
)

type db interface {
	Room(id string) (*app.Room, error)
	AddRoom(room *app.Room) error

	User(id app.ID) (*app.User, error)
	AddUser(user *app.User) error
}

type boltDBImpl struct {
	*bolt.DB
}

func initBoltDB() (db, error) {
	bdb, err := bolt.Open("radiotation.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
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
	return nil, nil
}

func (b *boltDBImpl) AddRoom(rm *app.Room) error {
	return nil
}

func (b *boltDBImpl) User(id string) (*app.User, error) {
	return nil, nil
}

func (b *boltDBImpl) AddUser(rm *app.User) error {
	return nil
}
