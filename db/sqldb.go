package db

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/bcspragu/Radiotation/music"
	// Import SQLite driver
	_ "github.com/mattn/go-sqlite3"
)

var (
	createUsersTableStmt = `CREATE TABLE IF NOT EXISTS Users (
  id TEXT PRIMARY KEY,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL
)`

	createRoomsTableStmt = `CREATE TABLE IF NOT EXISTS Rooms (
  id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  rotator BLOB NOT NULL,
	music_service INTEGER NOT NULL
)`

	createQueuesTableStmt = `CREATE TABLE IF NOT EXISTS Queues (
	room_id TEXT NOT NULL,
	user_id TEXT NOT NULL,
	offset INTEGER DEFAULT 0,
	tracks BLOB NOT NULL,
	FOREIGN KEY (room_id) REFERENCES Rooms(id),
	FOREIGN KEY (user_id) REFERENCES Users(id),
	PRIMARY KEY (room_id, user_id)
)`

	createHistoryTableStmt = `CREATE TABLE IF NOT EXISTS History (
	room_id TEXT PRIMARY KEY,
	tracks BLOB NOT NULL,
	FOREIGN KEY (room_id) REFERENCES Rooms(id)
)`

	getRoomStmt  = `SELECT id, display_name, rotator, music_service FROM Rooms WHERE id = ?`
	getRoomsStmt = `SELECT id, display_name, rotator, music_service FROM Rooms`
	addRoomStmt  = `INSERT INTO Rooms (id, display_name, rotator, music_service) VALUES (?, ?, ?, ?)`

	getUserStmt        = `SELECT id, first_name, last_name FROM Users WHERE id = ?`
	getUsersInRoomStmt = `SELECT user_id FROM Queues WHERE room_id = ?`
	addUserStmt        = `INSERT INTO Users (id, first_name, last_name) VALUES (?, ?, ?)`

	addQueueStmt = `INSERT INTO Queues (room_id, user_id, tracks) VALUES (?, ?, ?)`
	getQueueStmt = `SELECT room_id, user_id, offset, tracks FROM Queues WHERE room_id = ? AND user_id = ?`

	updateTracksStmt = `UPDATE Queues SET tracks = ? WHERE room_id = ? AND user_id = ?`
)

// sqlDB implements the Radiotation database API, backed by a SQLite database.
// NOTE: Since the database doesn't support concurrent writers, we don't
// actually hold the *sql.DB in this struct, we force all callers to get a
// handle via channels.
type sqlDB struct {
	dbChan   chan func(db *sql.DB)
	doneChan chan struct{}
	closeFn  func() error
}

func InitSQLiteDB() (DB, error) {
	db, err := sql.Open("sqlite3", "radiotation-sql.db")
	if err != nil {
		return nil, err
	}

	for _, stmt := range []string{createUsersTableStmt, createRoomsTableStmt, createQueuesTableStmt, createHistoryTableStmt} {
		if _, err := db.Exec(stmt); err != nil {
			return nil, err
		}
	}
	sdb := &sqlDB{
		dbChan:   make(chan func(*sql.DB)),
		doneChan: make(chan struct{}),
		closeFn: func() error {
			return db.Close()
		},
	}
	go sdb.run(db)
	return sdb, nil
}

func (s *sqlDB) run(db *sql.DB) {
	for {
		select {
		case dbFn := <-s.dbChan:
			dbFn(db)
		case <-s.doneChan:
			db.Close()
			return
		}
	}
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func loadQueue(s scanner) (*Queue, error) {
	var qr struct {
		roomID string
		userID string
		offset int
		tBytes []byte
	}
	if err := s.Scan(&qr.roomID, &qr.userID, &qr.offset, &qr.tBytes); err != nil {
		return nil, err
	}

	uid, err := userIDFromString(qr.userID)
	if err != nil {
		return nil, err
	}

	var tracks []music.Track
	if err := gob.NewDecoder(bytes.NewReader(qr.tBytes)).Decode(&tracks); err != nil {
		return nil, fmt.Errorf("failed to decode tracks: %v", err)
	}

	return &Queue{
		ID: QueueID{
			RoomID: RoomID(qr.roomID),
			UserID: uid,
		},
		Offset: qr.offset,
		Tracks: tracks,
	}, nil
}

func loadUser(s scanner) (*User, error) {
	var ur struct {
		id        string
		firstName string
		lastName  string
	}
	if err := s.Scan(&ur.id, &ur.firstName, &ur.lastName); err != nil {
		return nil, err
	}
	uid, err := userIDFromString(ur.id)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:    uid,
		First: ur.firstName,
		Last:  ur.lastName,
	}, nil
}

func loadRoom(s scanner) (*Room, error) {
	var rr struct {
		id           string
		displayName  string
		rBytes       []byte
		musicService int
	}
	if err := s.Scan(&rr.id, &rr.displayName, &rr.rBytes, &rr.musicService); err != nil {
		return nil, err
	}

	var rot Rotator
	if err := gob.NewDecoder(bytes.NewReader(rr.rBytes)).Decode(&rot); err != nil {
		return nil, fmt.Errorf("failed to decode rotator: %v", err)
	}

	return &Room{
		ID:           RoomID(rr.id),
		DisplayName:  rr.displayName,
		MusicService: MusicService(rr.musicService),
		Rotator:      rot,
	}, nil
}

func (s *sqlDB) Room(rid RoomID) (*Room, error) {
	type result struct {
		room *Room
		err  error
	}
	rmChan := make(chan *result)
	s.dbChan <- func(db *sql.DB) {
		r, err := loadRoom(db.QueryRow(getRoomStmt, string(rid)))
		rmChan <- &result{room: r, err: err}
	}
	res := <-rmChan
	if res.err == sql.ErrNoRows {
		return nil, ErrRoomNotFound
	}
	if res.err != nil {
		return nil, fmt.Errorf("failed to load room: %v", res.err)
	}
	return res.room, nil
}

func (s *sqlDB) NextTrack(RoomID) (music.Track, error) {
	return music.Track{}, ErrOperationNotImplemented
}

func (s *sqlDB) Rooms() ([]*Room, error) {
	type result struct {
		rooms []*Room
		err   error
	}
	rmChan := make(chan *result)
	s.dbChan <- func(db *sql.DB) {
		rows, err := db.Query(getRoomsStmt)
		if err != nil {
			rmChan <- &result{err: err}
			return
		}

		var res result
		for rows.Next() {
			r, err := loadRoom(rows)
			if err != nil {
				rmChan <- &result{err: err}
				return
			}
			res.rooms = append(res.rooms, r)
		}
		rmChan <- &res
	}
	res := <-rmChan
	if res.err != nil {
		return nil, fmt.Errorf("failed to load room: %v", res.err)
	}
	return res.rooms, nil
}

func (s *sqlDB) AddRoom(rm *Room) error {
	errChan := make(chan error)
	s.dbChan <- func(db *sql.DB) {
		rBytes, err := rotatorBytes(rm.Rotator)
		if err != nil {
			errChan <- err
			return
		}
		_, err = db.Exec(addRoomStmt, string(rm.ID), rm.DisplayName, rBytes, int(rm.MusicService))
		errChan <- err
	}
	return <-errChan
}

func rotatorBytes(r Rotator) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&r)
	return buf.Bytes(), err
}

func (s *sqlDB) AddUserToRoom(rid RoomID, uid UserID) error {
	errChan := make(chan error)
	s.dbChan <- func(db *sql.DB) {
		tBytes, err := tracksBytes([]music.Track{})
		_, err = db.Exec(addQueueStmt, string(rid), uid.String(), tBytes)
		errChan <- err
	}
	return <-errChan
}

func tracksBytes(tracks []music.Track) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(tracks)
	return buf.Bytes(), err
}

func (s *sqlDB) User(id UserID) (*User, error) {
	type result struct {
		user *User
		err  error
	}
	uChan := make(chan *result)
	s.dbChan <- func(db *sql.DB) {
		u, err := loadUser(db.QueryRow(getUserStmt, id.String()))
		uChan <- &result{user: u, err: err}
	}
	res := <-uChan
	if res.err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return res.user, nil
}

func (s *sqlDB) Users(rid RoomID) ([]*User, error) {
	type result struct {
		users []*User
		err   error
	}
	uChan := make(chan *result)
	s.dbChan <- func(db *sql.DB) {
		tx, err := db.Begin()
		defer tx.Rollback()
		if err != nil {
			uChan <- &result{err: err}
			return
		}
		rows, err := tx.Query(getUsersInRoomStmt, string(rid))
		if err != nil {
			uChan <- &result{err: err}
			return
		}
		var res result
		for rows.Next() {
			u, err := loadUser(rows)
			if err != nil {
				uChan <- &result{err: err}
				return
			}
			res.users = append(res.users, u)
		}
		if err := tx.Commit(); err != nil {
			uChan <- &result{err: err}
			return
		}
		uChan <- &res
	}
	res := <-uChan
	if res.err != nil {
		return nil, fmt.Errorf("failed to load users: %v", res.err)
	}
	return res.users, nil
}

func (s *sqlDB) AddUser(user *User) error {
	errChan := make(chan error)
	s.dbChan <- func(db *sql.DB) {
		_, err := db.Exec(addUserStmt, user.ID.String(), user.First, user.Last)
		errChan <- err
	}
	return <-errChan
}

func (s *sqlDB) Queue(qid QueueID) (*Queue, error) {
	type result struct {
		queue *Queue
		err   error
	}
	qChan := make(chan *result)
	s.dbChan <- func(db *sql.DB) {
		q, err := loadQueue(db.QueryRow(getQueueStmt, string(qid.RoomID), qid.UserID.String()))
		qChan <- &result{queue: q, err: err}
	}
	res := <-qChan
	if res.err == sql.ErrNoRows {
		return nil, ErrQueueNotFound
	}
	if res.err != nil {
		return nil, fmt.Errorf("failed to load queue: %v", res.err)
	}
	return res.queue, nil
}

func queueIDFromString(qid string) (QueueID, error) {
	idp := strings.SplitN(qid, ":", 2)
	if len(idp) != 2 {
		return QueueID{}, fmt.Errorf("malformed qid %q", qid)
	}
	uid, err := userIDFromString(idp[1])
	if err != nil {
		return QueueID{}, err
	}
	return QueueID{
		RoomID: RoomID(idp[0]),
		UserID: uid,
	}, nil
}

func (s *sqlDB) AddTrack(qid QueueID, track music.Track) error {
	errChan := make(chan error)
	s.dbChan <- func(db *sql.DB) {
		tx, err := db.Begin()
		defer tx.Rollback()
		if err != nil {
			errChan <- err
			return
		}
		q, err := loadQueue(tx.QueryRow(getQueueStmt, string(qid.RoomID), qid.UserID.String()))
		ts := append(q.Tracks, track)
		tBytes, err := tracksBytes(ts)
		if err != nil {
			errChan <- err
			return
		}
		if _, err := tx.Exec(updateTracksStmt, tBytes, string(qid.RoomID), qid.UserID.String()); err != nil {
			errChan <- err
			return
		}
		errChan <- tx.Commit()
	}
	return <-errChan
}

func (s *sqlDB) RemoveTrack(qid QueueID, idx int) error {
	errChan := make(chan error)
	s.dbChan <- func(db *sql.DB) {
		tx, err := db.Begin()
		defer tx.Rollback()
		if err != nil {
			errChan <- err
			return
		}
		q, err := loadQueue(tx.QueryRow(getQueueStmt, string(qid.RoomID), qid.UserID.String()))
		if err != nil {
			errChan <- err
			return
		}
		if idx >= len(q.Tracks) {
			errChan <- fmt.Errorf("asked to remove track index %d, only have %d tracks", idx, len(q.Tracks))
			return
		}
		if idx < q.Offset {
			errChan <- fmt.Errorf("asked to remove track index %d, we're passed that on index %d", idx, q.Offset)
			return
		}

		// Remove the track from the queue.
		ts := q.Tracks
		copy(ts[idx:], ts[idx+1:])
		ts[len(ts)-1] = music.Track{}
		ts = ts[:len(ts)-1]

		tBytes, err := tracksBytes(ts)
		if err != nil {
			errChan <- err
			return
		}
		if _, err := tx.Exec(updateTracksStmt, tBytes, string(qid.RoomID), qid.UserID.String()); err != nil {
			errChan <- err
			return
		}
		errChan <- tx.Commit()
	}
	return <-errChan
}

func (s *sqlDB) History(RoomID) ([]*TrackEntry, error) {
	return []*TrackEntry{}, ErrOperationNotImplemented
}

func (s *sqlDB) AddToHistory(RoomID, *TrackEntry) error {
	return ErrOperationNotImplemented
}

func (s *sqlDB) Close() error {
	return s.closeFn()
}
