package memdb

import (
	"errors"
	"math/rand"
	"sort"
	"strings"
	"sync"

	"github.com/bcspragu/Radiotation/db"
	"github.com/bcspragu/Radiotation/radio"
)

func New(src rand.Source) (*DB, error) {
	return &DB{
		rooms:   make(map[db.RoomID]*room),
		users:   make(map[db.UserID]*db.User),
		queues:  make(map[db.RoomID][]*queue),
		history: make(map[db.RoomID][]*db.TrackEntry),
		src:     src,
	}, nil
}

type room struct {
	room    *db.Room
	rotator db.Rotator
}

type queue struct {
	ID     db.QueueID
	Offset int
	Tracks []*db.QueueTrack
}

func (q *queue) nextTrack() (*radio.Track, bool) {
	if q.Offset >= len(q.Tracks) {
		return nil, false
	}
	return q.Tracks[q.Offset].Track, true
}

type DB struct {
	sync.RWMutex
	// Map from roomID -> room
	rooms map[db.RoomID]*room
	// Map from uid -> user
	users map[db.UserID]*db.User
	// Map from roomID -> list of queues
	queues map[db.RoomID][]*queue
	// Map from roomID -> list of played track entries
	history map[db.RoomID][]*db.TrackEntry
	src     rand.Source
}

func (m *DB) Room(id db.RoomID) (*db.Room, error) {
	m.RLock()
	defer m.RUnlock()
	r, ok := m.rooms[id]
	if !ok {
		return nil, db.ErrRoomNotFound
	}

	return r.room, nil
}

func (m *DB) NextTrack(rID db.RoomID) (*db.User, *radio.Track, error) {
	r, ok := m.rooms[rID]
	if !ok {
		return nil, nil, db.ErrRoomNotFound
	}

	qs, ok := m.queues[rID]
	if !ok {
		return nil, nil, db.ErrQueueNotFound
	}

	for i := 0; i < len(qs); i++ {
		idx := r.rotator.NextIndex()

		if idx >= len(qs) {
			return nil, nil, errors.New("invalid index in rotation")
		}

		q := qs[idx]
		u, ok := m.users[q.ID.UserID]
		if !ok {
			return nil, nil, db.ErrUserNotFound
		}

		nt, ok := q.nextTrack()
		if !ok {
			// Continue onto the next queue in the rotation.
			continue
		}

		q.Tracks[q.Offset].Played = true
		q.Offset++
		return u, nt, nil
	}
	return nil, nil, db.ErrNoTracksInQueue
}

func (m *DB) SearchRooms(q string) ([]*db.Room, error) {
	m.RLock()
	defer m.RUnlock()

	if q == "" {
		return []*db.Room{}, nil
	}

	q = strings.ToLower(q)

	var rs []*db.Room
	for _, r := range m.rooms {
		if strings.Contains(strings.ToLower(r.room.DisplayName), q) {
			rs = append(rs, r.room)
		}
	}

	return rs, nil
}

func (m *DB) Rooms() ([]*db.Room, error) {
	m.RLock()
	defer m.RUnlock()
	var rooms []*db.Room
	for _, r := range m.rooms {
		rooms = append(rooms, r.room)
	}

	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i].DisplayName < rooms[j].DisplayName
	})

	return rooms, nil
}

func (m *DB) AddRoom(r *db.Room) (db.RoomID, error) {
	m.Lock()
	defer m.Unlock()
	r.ID = db.RoomID(db.RandomID(m.src))
	m.rooms[r.ID] = &room{
		room:    r,
		rotator: db.NewRotator(r.RotatorType),
	}
	m.queues[r.ID] = []*queue{}
	m.history[r.ID] = []*db.TrackEntry{}
	return r.ID, nil
}

func (m *DB) AddUserToRoom(rID db.RoomID, uID db.UserID) error {
	m.Lock()
	defer m.Unlock()

	r, ok := m.rooms[rID]
	if !ok {
		return db.ErrRoomNotFound
	}
	r.rotator.Add()

	qs, ok := m.queues[rID]
	if !ok {
		return db.ErrQueueNotFound
	}

	m.queues[rID] = append(qs, &queue{
		ID:     db.QueueID{UserID: uID, RoomID: rID},
		Tracks: []*db.QueueTrack{},
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

func (m *DB) Users(rID db.RoomID) ([]*db.User, error) {
	m.RLock()
	defer m.RUnlock()

	if _, ok := m.rooms[rID]; !ok {
		return nil, db.ErrRoomNotFound
	}

	qs, ok := m.queues[rID]
	if !ok {
		return nil, db.ErrRoomNotFound
	}

	var us []*db.User
	for _, q := range qs {
		u, ok := m.users[q.ID.UserID]
		if !ok {
			return nil, db.ErrUserNotFound
		}
		us = append(us, u)
	}

	return us, nil
}

func (m *DB) AddUser(user *db.User) error {
	m.Lock()
	defer m.Unlock()
	m.users[user.ID] = user
	return nil
}

func (m *DB) Tracks(id db.QueueID, qo *db.QueueOptions) ([]*db.QueueTrack, error) {
	m.RLock()
	defer m.RUnlock()

	q, ok := m.queueByID(id)
	if !ok {
		return nil, db.ErrQueueNotFound
	}

	var keep func(*db.QueueTrack) bool
	switch qo.Type {
	case db.PlayedOnly:
		keep = func(qt *db.QueueTrack) bool { return qt.Played }
	case db.UnplayedOnly:
		keep = func(qt *db.QueueTrack) bool { return !qt.Played }
	case db.AllTracks:
		keep = func(*db.QueueTrack) bool { return true }
	}

	var qts []*db.QueueTrack
	for _, t := range q.Tracks {
		if keep(t) {
			qts = append(qts, t)
		}
	}

	return qts, nil
}

func (m *DB) queueByID(id db.QueueID) (*queue, bool) {
	qs, ok := m.queues[id.RoomID]
	if !ok {
		return nil, false
	}

	for _, q := range qs {
		if q.ID == id {
			return q, true
		}
	}

	return nil, false
}

func (m *DB) AddTrack(id db.QueueID, track *radio.Track, afterQTID string) error {
	m.Lock()
	defer m.Unlock()

	q, ok := m.queueByID(id)
	if !ok {
		return db.ErrQueueNotFound
	}

	if afterQTID == "" {
		q.Tracks = append([]*db.QueueTrack{&db.QueueTrack{
			ID:     db.RandomTrackID(m.src),
			Played: false,
			Track:  track,
		}}, q.Tracks...)

		return nil
	}

	for i, qt := range q.Tracks {
		if qt.ID != afterQTID {
			continue
		}

		// If the next track has already played, no inserting tracks.
		if i < len(q.Tracks)-1 && q.Tracks[i+1].Played {
			return errors.New("can't add a track before one that's already played")
		}

		// If we're here, we should insert the track.
		q.Tracks = append(q.Tracks, nil)
		copy(q.Tracks[i+2:], q.Tracks[i+1:])
		q.Tracks[i+1] = &db.QueueTrack{
			ID:     db.RandomTrackID(m.src),
			Played: false,
			Track:  track,
		}
		return nil
	}

	// If we're here, we didn't find the track.
	return db.ErrQueueNotFound
}

func (m *DB) RemoveTrack(id db.QueueID, qtID string) error {
	m.Lock()
	defer m.Unlock()

	q, ok := m.queueByID(id)
	if !ok {
		return db.ErrQueueNotFound
	}

	for i, qt := range q.Tracks {
		if qt.ID != qtID {
			continue
		}

		// If the track has already played, no removing it.
		if qt.Played {
			return errors.New("can't remove a track that's already played")
		}

		// If we're here, we should remove the track.
		copy(q.Tracks[i:], q.Tracks[i+1:])
		q.Tracks[len(q.Tracks)-1] = nil
		q.Tracks = q.Tracks[:len(q.Tracks)-1]
		return nil
	}

	// If we're here, we didn't find the track.
	return db.ErrQueueNotFound
}

func (m *DB) History(rID db.RoomID) ([]*db.TrackEntry, error) {
	m.RLock()
	defer m.RUnlock()
	tes, ok := m.history[rID]
	if !ok {
		return nil, db.ErrRoomNotFound
	}
	return tes, nil
}

func (m *DB) AddToHistory(rID db.RoomID, trackEntry *db.TrackEntry) (int, error) {
	m.Lock()
	defer m.Unlock()
	m.history[rID] = append(m.history[rID], trackEntry)
	return len(m.history[rID]) - 1, nil
}

func (m *DB) MarkVetoed(db.RoomID, db.UserID) error {
	return db.ErrOperationNotImplemented
}
