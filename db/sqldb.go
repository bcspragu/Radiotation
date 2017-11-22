package db

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
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
	rotator_type INTEGER NOT NULL,
	music_service INTEGER NOT NULL
)`

	createQueuesTableStmt = `CREATE TABLE IF NOT EXISTS Queues (
	room_id TEXT NOT NULL,
	user_id TEXT NOT NULL,
	offset INTEGER DEFAULT 0,
	tracks BLOB NOT NULL,
	joined_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
	FOREIGN KEY (room_id) REFERENCES Rooms(id),
	FOREIGN KEY (user_id) REFERENCES Users(id),
	PRIMARY KEY (room_id, user_id)
)`

	createHistoryTableStmt = `CREATE TABLE IF NOT EXISTS History (
	room_id TEXT PRIMARY KEY,
	track_entries BLOB NOT NULL,
	FOREIGN KEY (room_id) REFERENCES Rooms(id)
)`

	getRoomStmt  = `SELECT id, display_name, rotator_type, music_service FROM Rooms WHERE id = ?`
	getRoomsStmt = `SELECT id, display_name, rotator_type, music_service FROM Rooms`
	addRoomStmt  = `INSERT INTO Rooms (id, display_name, rotator, rotator_type, music_service) VALUES (?, ?, ?, ?, ?)`

	getRotatorStmt = `SELECT rotator FROM Rooms WHERE id = ?`

	updateRotatorStmt = `UPDATE Rooms SET rotator = ? WHERE id = ?`

	getUserStmt        = `SELECT id, first_name, last_name FROM Users WHERE id = ?`
	getUsersStmt       = `SELECT id, first_name, last_name FROM Users WHERE id IN (%s)`
	getUsersInRoomStmt = `SELECT user_id FROM Queues WHERE room_id = ? ORDER BY joined_at`
	addUserStmt        = `INSERT INTO Users (id, first_name, last_name) VALUES (?, ?, ?)`

	addQueueStmt = `INSERT INTO Queues (room_id, user_id, tracks) VALUES (?, ?, ?)`
	getQueueStmt = `SELECT room_id, user_id, offset, tracks FROM Queues WHERE room_id = ? AND user_id = ?`

	updateOffsetStmt = `UPDATE Queues SET offset = ? WHERE room_id = ? AND user_id = ?`
	updateTracksStmt = `UPDATE Queues SET tracks = ? WHERE room_id = ? AND user_id = ?`

	createHistoryStmt = `INSERT INTO History (room_id, track_entries) VALUES (?, ?)`
	getHistoryStmt    = `SELECT track_entries FROM History WHERE room_id = ?`
	updateHistoryStmt = `UPDATE History SET track_entries = ? WHERE room_id = ?`
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

// InitSQLiteDB creates a new *sqlDB that is stored on disk as
// 'radiotation-sql.db'.
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

// run handles all database calls, and ensures that only one thing is happening
// against the database at a time.
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

func loadTrackEntries(s scanner) ([]*TrackEntry, error) {
	var tBytes []byte
	if err := s.Scan(&tBytes); err != nil {
		return nil, err
	}

	var tracks []*TrackEntry
	if err := gob.NewDecoder(bytes.NewReader(tBytes)).Decode(&tracks); err != nil {
		return nil, fmt.Errorf("failed to decode tracks: %v", err)
	}

	return tracks, nil
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

func loadUsers(tx *sql.Tx, rid RoomID) ([]*User, error) {
	rows, err := tx.Query(getUsersInRoomStmt, string(rid))
	if err != nil {
		return nil, err
	}
	var uids []interface{}
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		uids = append(uids, uid)
	}

	rows, err = tx.Query(fmt.Sprintf(getUsersStmt, sqlInput(len(uids))), uids...)
	if err != nil {
		return nil, err
	}

	var users []*User
	for rows.Next() {
		u, err := loadUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
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

func loadRotator(s scanner) (Rotator, error) {
	var rBytes []byte
	if err := s.Scan(&rBytes); err != nil {
		return nil, err
	}

	var rot Rotator
	if err := gob.NewDecoder(bytes.NewReader(rBytes)).Decode(&rot); err != nil {
		return nil, fmt.Errorf("failed to decode rotator: %v", err)
	}
	return rot, nil
}

func loadRoom(s scanner) (*Room, error) {
	var rr struct {
		id           string
		displayName  string
		rotatorType  int
		musicService int
	}
	if err := s.Scan(&rr.id, &rr.displayName, &rr.rotatorType, &rr.musicService); err != nil {
		return nil, err
	}

	return &Room{
		ID:           RoomID(rr.id),
		DisplayName:  rr.displayName,
		RotatorType:  RotatorType(rr.rotatorType),
		MusicService: MusicService(rr.musicService),
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

func (s *sqlDB) NextTrack(rid RoomID) (*User, music.Track, error) {
	type result struct {
		user  *User
		track music.Track
		err   error
	}
	tChan := make(chan *result)
	s.dbChan <- func(db *sql.DB) {
		tx, err := db.Begin()
		if err != nil {
			tChan <- &result{err: err}
			return
		}
		defer tx.Rollback()

		rot, err := loadRotator(tx.QueryRow(getRotatorStmt, string(rid)))
		if err != nil {
			tChan <- &result{err: err}
			return
		}

		users, err := loadUsers(tx, rid)
		if err != nil {
			tChan <- &result{err: err}
			return
		}

		for i := 0; i < len(users); i++ {
			idx, last := rot.NextIndex()
			if last {
				// Start a rotation with any new users
				rot.Start(len(users))
			}

			if idx >= len(users) {
				tChan <- &result{err: fmt.Errorf("rotator is broken, returned index %d for list of %d users", idx, len(users))}
				return
			}

			u := users[idx]
			if u == nil {
				log.Printf("everything is broken, returned a nil user at index %d of %d", idx, len(users))
				continue
			}

			q, err := loadQueue(tx.QueryRow(getQueueStmt, string(rid), u.ID.String()))
			if err != nil {
				tChan <- &result{err: err}
				return
			}

			t, err := nextTrack(q)
			if err == ErrNoTracksInQueue {
				continue
			}
			if err != nil {
				tChan <- &result{err: err}
				return
			}
			if _, err := tx.Exec(updateOffsetStmt, q.Offset+1, string(rid), u.ID.String()); err != nil {
				tChan <- &result{err: err}
				return
			}

			// If we're here, we've got a track, we just need to serialize/save the
			// rotator, commit the transaction, and go about our business.
			rBytes, err := rotatorBytes(rot)
			if err != nil {
				tChan <- &result{err: err}
				return
			}

			if _, err := tx.Exec(updateRotatorStmt, rBytes, string(rid)); err != nil {
				tChan <- &result{err: err}
				return
			}

			if err := tx.Commit(); err != nil {
				tChan <- &result{err: err}
				return
			}
			tChan <- &result{user: u, track: t}
			return
		}
		tChan <- &result{err: ErrNoTracksInQueue}
	}
	res := <-tChan
	if res.err != nil {
		return nil, music.Track{}, res.err
	}
	return res.user, res.track, nil
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
		tx, err := db.Begin()
		if err != nil {
			errChan <- err
			return
		}
		defer tx.Rollback()

		rBytes, err := rotatorBytes(newRotator(rm.RotatorType))
		if err != nil {
			errChan <- err
			return
		}
		_, err = tx.Exec(addRoomStmt, string(rm.ID), rm.DisplayName, rBytes, rm.RotatorType, int(rm.MusicService))
		if err != nil {
			errChan <- err
			return
		}

		// Create an empty history list.
		teBytes, err := trackEntryBytes([]*TrackEntry{})
		if err != nil {
			errChan <- err
			return
		}
		if _, err := tx.Exec(createHistoryStmt, string(rm.ID), teBytes); err != nil {
			errChan <- err
			return
		}
		errChan <- tx.Commit()
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
		tx, err := db.Begin()
		if err != nil {
			errChan <- err
			return
		}
		defer tx.Rollback()

		// Add the queue.
		tBytes, err := tracksBytes([]music.Track{})
		_, err = tx.Exec(addQueueStmt, string(rid), uid.String(), tBytes)
		if err != nil {
			errChan <- err
			return
		}

		users, err := loadUsers(tx, rid)
		if err != nil {
			errChan <- err
			return
		}

		if len(users) == 1 {
			if err := startRotator(tx, rid); err != nil {
				errChan <- err
				return
			}
		}
		errChan <- tx.Commit()
	}
	return <-errChan
}

func startRotator(tx *sql.Tx, rid RoomID) error {
	rot, err := loadRotator(tx.QueryRow(getRotatorStmt, string(rid)))
	if err != nil {
		return err
	}
	rot.Start(1)
	rBytes, err := rotatorBytes(rot)
	if err != nil {
		return err
	}
	_, err = tx.Exec(updateRotatorStmt, rBytes, string(rid))
	return err
}

func tracksBytes(tracks []music.Track) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(tracks)
	return buf.Bytes(), err
}

func trackEntryBytes(entries []*TrackEntry) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(entries)
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
		if err != nil {
			uChan <- &result{err: err}
			return
		}
		defer tx.Rollback()

		users, err := loadUsers(tx, rid)
		if err != nil {
			uChan <- &result{err: err}
			return
		}

		if err := tx.Commit(); err != nil {
			uChan <- &result{err: err}
			return
		}
		uChan <- &result{users: users}
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
		if err != nil {
			errChan <- err
			return
		}
		defer tx.Rollback()
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
		if err != nil {
			errChan <- err
			return
		}
		defer tx.Rollback()
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

func (s *sqlDB) History(rid RoomID) ([]*TrackEntry, error) {
	type result struct {
		tracks []*TrackEntry
		err    error
	}
	hChan := make(chan *result)
	s.dbChan <- func(db *sql.DB) {
		ts, err := loadTrackEntries(db.QueryRow(getHistoryStmt, string(rid)))
		hChan <- &result{tracks: ts, err: err}
	}
	res := <-hChan
	if res.err == sql.ErrNoRows {
		return nil, ErrRoomNotFound
	}
	if res.err != nil {
		return nil, fmt.Errorf("failed to load history: %v", res.err)
	}
	return res.tracks, nil
}

func (s *sqlDB) AddToHistory(rid RoomID, te *TrackEntry) error {
	errChan := make(chan error)
	s.dbChan <- func(db *sql.DB) {
		tx, err := db.Begin()
		if err != nil {
			errChan <- err
			return
		}
		defer tx.Rollback()
		ts, err := loadTrackEntries(tx.QueryRow(getHistoryStmt, string(rid)))
		ts = append(ts, te)
		teBytes, err := trackEntryBytes(ts)
		if err != nil {
			errChan <- err
			return
		}
		if _, err := tx.Exec(updateHistoryStmt, teBytes, string(rid)); err != nil {
			errChan <- err
			return
		}
		errChan <- tx.Commit()
	}
	return <-errChan
}

func (s *sqlDB) MarkVetoed(rid RoomID, uid UserID) error {
	errChan := make(chan error)
	s.dbChan <- func(db *sql.DB) {
		tx, err := db.Begin()
		if err != nil {
			errChan <- err
			return
		}
		defer tx.Rollback()
		ts, err := loadTrackEntries(tx.QueryRow(getHistoryStmt, string(rid)))
		if err != nil {
			errChan <- err
			return
		}

		if len(ts) == 0 {
			errChan <- errors.New("no tracks in history")
			return
		}

		te := ts[len(ts)-1]
		te.Vetoed = true
		te.VetoedBy = uid
		ts[len(ts)-1] = te

		teBytes, err := trackEntryBytes(ts)
		if err != nil {
			errChan <- err
			return
		}
		if _, err := tx.Exec(updateHistoryStmt, teBytes, string(rid)); err != nil {
			errChan <- err
			return
		}
		errChan <- tx.Commit()
	}
	return <-errChan
}

func (s *sqlDB) Close() error {
	return s.closeFn()
}

func sqlInput(n int) string {
	if n <= 0 {
		return ""
	}
	return "?" + strings.Repeat(", ?", n-1)
}
