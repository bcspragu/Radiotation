package db

import (
	"errors"

	"github.com/bcspragu/Radiotation/music"
)

var (
	ErrOperationNotImplemented = errors.New("radiotation: operation not implemented")
	ErrUserNotFound            = errors.New("radiotation: user not found")
	ErrRoomNotFound            = errors.New("radiotation: room not found")
	ErrQueueNotFound           = errors.New("radiotation: queue not found")
)

type TrackEntry struct {
	UserID UserID
	Track  music.Track

	// VetoedBy is only set if Vetoed is true.
	Vetoed   bool
	VetoedBy UserID
}

type RoomDB interface {
	Room(RoomID) (*Room, error)
	Rooms() ([]*Room, error)

	AddRoom(*Room) error
	AddUserToRoom(RoomID, UserID) error
}

type UserDB interface {
	User(UserID) (*User, error)
	Users(RoomID) ([]*User, error)

	AddUser(user *User) error
}

type QueueDB interface {
	Queue(QueueID) (*Queue, error)

	AddTrack(QueueID, music.Track) error
	RemoveTrack(QueueID, int) error
}

type HistoryDB interface {
	History(RoomID) ([]music.Track, error)
	AddToHistory(RoomID, *TrackEntry) error
}

type DB interface {
	RoomDB
	UserDB
	QueueDB
	HistoryDB
}
