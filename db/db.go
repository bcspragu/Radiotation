package db

import (
	"errors"

	"github.com/bcspragu/Radiotation/radio"
)

var (
	ErrOperationNotImplemented = errors.New("radiotation: operation not implemented")
	ErrUserNotFound            = errors.New("radiotation: user not found")
	ErrRoomNotFound            = errors.New("radiotation: room not found")
	ErrQueueNotFound           = errors.New("radiotation: queue not found")
	ErrNoTracksInQueue         = errors.New("radiotation: no tracks in queue")
)

type QueueID struct {
	RoomID RoomID
	UserID UserID
}

type TrackEntry struct {
	UserID UserID
	Track  radio.Track

	// VetoedBy is only set if Vetoed is true.
	Vetoed   bool
	VetoedBy UserID
}

type RoomDB interface {
	Room(RoomID) (*Room, error)
	NextUser(RoomID) (*User, error)

	Rooms() ([]*Room, error)

	AddRoom(*Room) error
	AddUserToRoom(RoomID, UserID) error
}

type UserDB interface {
	User(UserID) (*User, error)
	Users(RoomID) ([]*User, error)

	AddUser(user *User) error
}

type TrackDB interface {
	Tracks(QueueID) (radio.TrackList, error)
	NextTrack(QueueID) (radio.Track, error)
	AddTrack(QueueID, radio.Track, int) error
	RemoveTrack(QueueID, int) error
}

type HistoryDB interface {
	History(RoomID) ([]*TrackEntry, error)
	AddToHistory(RoomID, *TrackEntry) error
	MarkVetoed(RoomID, UserID) error
}

type DB interface {
	RoomDB
	UserDB
	TrackDB
	HistoryDB
	Close() error
}
