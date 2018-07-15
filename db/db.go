package db

import (
	"errors"
	"math/rand"

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
	NextTrack(RoomID) (*User, *radio.Track, error)

	SearchRooms(string) ([]*Room, error)

	AddRoom(*Room) (RoomID, error)
	AddUserToRoom(RoomID, UserID) error
}

type UserDB interface {
	User(UserID) (*User, error)
	Users(RoomID) ([]*User, error)

	AddUser(user *User) error
}

type QueueType int

const (
	AllTracks QueueType = iota
	PlayedOnly
	UnplayedOnly
)

type QueueOptions struct {
	Type QueueType
}

type QueueTrack struct {
	ID     string
	Played bool

	Track *radio.Track
}

type QueueDB interface {
	Tracks(QueueID, *QueueOptions) ([]*QueueTrack, error)
	AddTrack(QueueID, *radio.Track, string) error
	RemoveTrack(QueueID, string) error
}

type HistoryDB interface {
	History(RoomID) ([]*TrackEntry, error)
	AddToHistory(RoomID, *TrackEntry) error
	MarkVetoed(RoomID, UserID) error
}

type DB interface {
	RoomDB
	UserDB
	QueueDB
	HistoryDB
}

var trackLetters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandomTrackID(src rand.Source) string {
	b := make([]byte, 64)
	r := rand.New(src)
	for i := range b {
		b[i] = trackLetters[r.Intn(len(trackLetters))]
	}
	return string(b)
}

var letters = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomID(src rand.Source) string {
	b := make([]byte, 4)
	r := rand.New(src)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
