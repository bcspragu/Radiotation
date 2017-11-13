package db

import (
	"time"

	"github.com/bcspragu/Radiotation/music"
	"github.com/boltdb/bolt"
)

var (
	RoomBucket = []byte("Room")
	UserBucket = []byte("User")
)

type boltDBImpl struct {
	*bolt.DB
}

func InitBoltDB() (DB, error) {
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

func (b *boltDBImpl) Room(RoomID) (*Room, error) {
	return nil, ErrOperationNotImplemented
}

func (b *boltDBImpl) Rooms() ([]*Room, error) {
	return nil, ErrOperationNotImplemented
}

func (b *boltDBImpl) AddRoom(*Room) error {
	return ErrOperationNotImplemented
}

func (b *boltDBImpl) AddUserToRoom(RoomID, UserID) error {
	return ErrOperationNotImplemented
}

func (b *boltDBImpl) User(id UserID) (*User, error) {
	return nil, ErrOperationNotImplemented
}

func (b *boltDBImpl) Users(rid RoomID) ([]*User, error) {
	return nil, ErrOperationNotImplemented
}

func (b *boltDBImpl) AddUser(user *User) error {
	return ErrOperationNotImplemented
}

func (b *boltDBImpl) Queue(QueueID) (*Queue, error) {
	return nil, ErrOperationNotImplemented
}

func (b *boltDBImpl) AddTrack(QueueID, music.Track) error {
	return ErrOperationNotImplemented
}

func (b *boltDBImpl) RemoveTrack(QueueID, int) error {
	return ErrOperationNotImplemented
}

func (b *boltDBImpl) History(RoomID) ([]music.Track, error) {
	return nil, ErrOperationNotImplemented
}

func (b *boltDBImpl) AddToHistory(RoomID, *TrackEntry) error {
	return ErrOperationNotImplemented
}
