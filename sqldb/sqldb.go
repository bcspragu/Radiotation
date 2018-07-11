package sqldb

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bcspragu/Radiotation/db"
	"github.com/bcspragu/Radiotation/radio"
	// Import SQLite driver
	_ "github.com/mattn/go-sqlite3"

	cryptorand "crypto/rand"
	"math/rand"
)

var (
	roomExistsStmt  = `SELECT EXISTS(SELECT 1 FROM Rooms WHERE id = ?)`
	getRoomStmt     = `SELECT id, display_name, rotator_type FROM Rooms WHERE id = ?`
	searchRoomsStmt = `SELECT id, display_name, rotator_type FROM Rooms WHERE normalized_name LIKE '%' || ? || '%'`
	addRoomStmt     = `INSERT INTO Rooms (id, display_name, normalized_name, rotator, rotator_type) VALUES (?, ?, ?, ?, ?)`

	getRotatorStmt    = `SELECT rotator FROM Rooms WHERE id = ?`
	updateRotatorStmt = `UPDATE Rooms SET rotator = ? WHERE id = ?`

	getUserStmt        = `SELECT id, first_name, last_name FROM Users WHERE id = ?`
	getUsersStmt       = `SELECT id, first_name, last_name FROM Users WHERE id IN (%s)`
	getUsersInRoomStmt = `SELECT user_id FROM Queues WHERE room_id = ? ORDER BY joined_at`
	addUserStmt        = `INSERT INTO Users (id, first_name, last_name) VALUES (?, ?, ?)`

	addQueueStmt      = `INSERT INTO Queues (room_id, user_id) VALUES (?, ?)`
	addQueueTrackStmt = `INSERT INTO QueueTracks (id, previous_id, next_id, track_id, room_id, user_id, played)
		VALUES (?, ?, ?, ?, ?, ?, 0)`
	getQueueStmt      = `SELECT next_queue_track_id FROM Queues WHERE room_id = ? AND user_id = ?`
	getQueueTrackStmt = `SELECT previous_id, next_id, played FROM QueueTracks
		WHERE id = ?`
	getFirstQueueTrackStmt    = `SELECT id FROM QueueTracks WHERE previous_id IS NULL AND room_id = ? AND user_id = ?`
	setQueueTrackPreviousStmt = `UPDATE QueueTracks SET previous_id = ? WHERE id = ?`
	setQueueTrackNextStmt     = `UPDATE QueueTracks SET next_id = ? WHERE id = ?`
	setQueueTrackPlayedStmt   = `UPDATE QueueTracks SET played = 1 WHERE id = ?`
	removeQueueTrackStmt      = `DELETE FROM QueueTracks WHERE id = ?`
	getTracksStmt             = `SELECT QueueTracks.id, previous_id, next_id, played, track FROM QueueTracks
		JOIN Tracks
		ON QueueTracks.track_id = Tracks.id
		WHERE room_id = ? AND user_id = ?`
	nextTrackStmt = `SELECT QueueTracks.id, track, next_id FROM QueueTracks
	JOIN Tracks
	ON QueueTracks.track_id = Tracks.id
	WHERE QueueTracks.id = (SELECT next_queue_track_id FROM Queues WHERE room_id = ? AND user_id = ?)`

	updateNextTrackStmt = `UPDATE Queues SET next_queue_track_id = ? WHERE room_id = ? AND user_id = ?`
	addTrackStmt        = `INSERT OR IGNORE INTO Tracks (id, track) VALUES (?, ?)`

	createHistoryStmt = `INSERT INTO History (room_id, track_entries) VALUES (?, ?)`
	getHistoryStmt    = `SELECT track_entries FROM History WHERE room_id = ?`
	updateHistoryStmt = `UPDATE History SET track_entries = ? WHERE room_id = ?`
)

// DB implements the Radiotation database API, backed by a SQLite database.
// NOTE: Since the database doesn't support concurrent writers, we don't
// actually hold the *sql.DB in this struct, we force all callers to get a
// handle via channels.
type DB struct {
	dbChan   chan func(*sql.DB)
	doneChan chan struct{}
	closeFn  func() error
	src      rand.Source

	sdb *sql.DB
}

// InitSQLiteDB creates a new *DB that is stored on disk as
// 'radiotation-sql.db'.
func New(fn string, src rand.Source) (*DB, error) {
	sdb, err := sql.Open("sqlite3", fn)
	if err != nil {
		return nil, err
	}

	db := &DB{
		dbChan:   make(chan func(*sql.DB)),
		doneChan: make(chan struct{}),
		closeFn: func() error {
			return sdb.Close()
		},
		src: src,
		sdb: sdb,
	}
	go db.run(sdb)
	return db, nil
}

// run handles all database calls, and ensures that only one thing is happening
// against the database at a time.
func (s *DB) run(sdb *sql.DB) {
	for {
		select {
		case dbFn := <-s.dbChan:
			dbFn(sdb)
		case <-s.doneChan:
			sdb.Close()
			return
		}
	}
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func loadTrackEntries(s scanner) ([]*db.TrackEntry, error) {
	var tBytes []byte
	if err := s.Scan(&tBytes); err != nil {
		return nil, err
	}

	var tracks []*db.TrackEntry
	if err := gob.NewDecoder(bytes.NewReader(tBytes)).Decode(&tracks); err != nil {
		return nil, fmt.Errorf("failed to decode tracks: %v", err)
	}

	return tracks, nil
}

func loadTrackList(tx *sql.Tx, qID db.QueueID, qo *db.QueueOptions) ([]db.QueueTrack, error) {
	var nextTrackID sql.NullString
	if err := tx.QueryRow(getQueueStmt, string(qID.RoomID), qID.UserID.String()).Scan(&nextTrackID); err != nil {
		return nil, err
	}

	suffix := ""
	switch qo.Type {
	case db.PlayedOnly:
		suffix = " AND played = 1"
	case db.UnplayedOnly:
		suffix = " AND played = 0"
	}

	stmt := getTracksStmt + suffix

	rows, err := tx.Query(stmt, string(qID.RoomID), qID.UserID.String())
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	type trackEntry struct {
		qtID       string
		previousID sql.NullString
		nextID     sql.NullString
		track      radio.Track
		played     bool
	}

	var first string
	links := make(map[string]trackEntry)

	for rows.Next() {
		var (
			trackBytes []byte
			te         trackEntry
		)
		if err := rows.Scan(&te.qtID, &te.previousID, &te.nextID, &te.played, &trackBytes); err != nil {
			return nil, err
		}

		if err := gob.NewDecoder(bytes.NewReader(trackBytes)).Decode(&te.track); err != nil {
			return nil, err
		}
		links[te.qtID] = te
		// If the track has no previous, it's the first one.
		if !te.previousID.Valid {
			if first != "" {
				return nil, errors.New("multiple tracks with no previous QueueTrack ID")
			}
			first = te.qtID
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// If we have at least one song in the queue, we should have a pointer to it
	// from the queue.
	if first == "" {
		return nil, errors.New("no tracks denoted as first")
	}

	var (
		tracks []db.QueueTrack
	)

	current := first
	idx := 0
	i := 0
	for {
		if current == nextTrackID.String {
			idx = i
		}

		te, ok := links[current]
		if !ok {
			return nil, fmt.Errorf("missing link for ID %q", current)
		}

		tracks = append(tracks, db.QueueTrack{
			ID:     te.qtID,
			Played: te.played,
			Track:  te.track,
		})

		// We're at the end of the chain.
		if !te.nextID.Valid {
			break
		}

		current = te.nextID.String
		i++
	}

	if len(tracks) != len(links) {
		return nil, fmt.Errorf("%d tracks, %d links, should be the same", len(tracks), len(links))
	}

	// Keep around the index of the next song in case we need it in the future.
	// TODO: Remove this, once the API has stabilized.
	_ = idx

	return tracks, nil
}

func loadUsers(tx *sql.Tx, rid db.RoomID) ([]*db.User, error) {
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

	var users []*db.User
	for rows.Next() {
		u, err := loadUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func loadUser(s scanner) (*db.User, error) {
	var ur struct {
		id        string
		firstName string
		lastName  string
	}
	if err := s.Scan(&ur.id, &ur.firstName, &ur.lastName); err != nil {
		return nil, err
	}
	uid, err := db.UserIDFromString(ur.id)
	if err != nil {
		return nil, err
	}

	return &db.User{
		ID:    uid,
		First: ur.firstName,
		Last:  ur.lastName,
	}, nil
}

func loadRotator(tx *sql.Tx, rID db.RoomID) (db.Rotator, error) {
	var rBytes []byte
	if err := tx.QueryRow(getRotatorStmt, string(rID)).Scan(&rBytes); err != nil {
		return nil, err
	}

	var rot db.Rotator
	if err := gob.NewDecoder(bytes.NewReader(rBytes)).Decode(&rot); err != nil {
		return nil, fmt.Errorf("failed to decode rotator: %v", err)
	}
	return rot, nil
}

func loadRoom(s scanner) (*db.Room, error) {
	var rr struct {
		id          string
		displayName string
		rotatorType int
	}
	if err := s.Scan(&rr.id, &rr.displayName, &rr.rotatorType); err != nil {
		return nil, err
	}

	return &db.Room{
		ID:          db.RoomID(rr.id),
		DisplayName: rr.displayName,
		RotatorType: db.RotatorType(rr.rotatorType),
	}, nil
}

func (s *DB) Room(rid db.RoomID) (*db.Room, error) {
	type result struct {
		room *db.Room
		err  error
	}
	rmChan := make(chan *result)
	s.dbChan <- func(sdb *sql.DB) {
		r, err := loadRoom(sdb.QueryRow(getRoomStmt, string(rid)))
		rmChan <- &result{room: r, err: err}
	}
	res := <-rmChan
	if res.err == sql.ErrNoRows {
		return nil, db.ErrRoomNotFound
	}
	if res.err != nil {
		return nil, fmt.Errorf("failed to load room: %v", res.err)
	}
	return res.room, nil
}

func (s *DB) NextTrack(rID db.RoomID) (*db.User, radio.Track, error) {
	type result struct {
		user  *db.User
		track radio.Track
		err   error
	}
	tChan := make(chan *result)
	s.dbChan <- func(sdb *sql.DB) {
		tx, err := sdb.Begin()
		if err != nil {
			tChan <- &result{err: err}
			return
		}
		defer tx.Rollback()

		rot, err := loadRotator(tx, rID)
		if err != nil {
			tChan <- &result{err: err}
			return
		}

		users, err := loadUsers(tx, rID)
		if err != nil {
			tChan <- &result{err: err}
			return
		}

		for i := 0; i < len(users); i++ {
			idx := rot.NextIndex()

			if idx >= len(users) {
				tChan <- &result{err: fmt.Errorf("rotator is broken, returned index %d for list of %d users", idx, len(users))}
				return
			}

			u := users[idx]
			if u == nil {
				log.Printf("everything is broken, returned a nil user at index %d of %d", idx, len(users))
				continue
			}

			var (
				qtID       string
				trackBytes []byte
				nextID     sql.NullString
			)
			err := tx.QueryRow(nextTrackStmt, string(rID), u.ID.String()).Scan(&qtID, &trackBytes, &nextID)
			if err == sql.ErrNoRows {
				continue
			} else if err != nil {
				tChan <- &result{err: err}
				return
			}

			var track radio.Track
			if err := gob.NewDecoder(bytes.NewReader(trackBytes)).Decode(&track); err != nil {
				tChan <- &result{err: err}
				return
			}

			// Update our next track, because we're taking this one.
			// next_queue_track_id, room_id, user_id
			if _, err := tx.Exec(updateNextTrackStmt, nextID, string(rID), u.ID.String()); err != nil {
				tChan <- &result{err: err}
				return
			}

			if _, err := tx.Exec(setQueueTrackPlayedStmt, qtID); err != nil {
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

			if _, err := tx.Exec(updateRotatorStmt, rBytes, string(rID)); err != nil {
				tChan <- &result{err: err}
				return
			}

			if err := tx.Commit(); err != nil {
				tChan <- &result{err: err}
				return
			}
			tChan <- &result{user: u, track: track}
			return
		}
		tChan <- &result{err: db.ErrNoTracksInQueue}
	}
	res := <-tChan
	if res.err != nil {
		return nil, radio.Track{}, res.err
	}
	return res.user, res.track, nil
}

func (s *DB) SearchRooms(q string) ([]*db.Room, error) {
	type result struct {
		rooms []*db.Room
		err   error
	}
	if q == "" {
		return []*db.Room{}, nil
	}

	rmChan := make(chan *result)
	s.dbChan <- func(sdb *sql.DB) {
		rows, err := sdb.Query(searchRoomsStmt, normalize(q))
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
	if res.rooms == nil {
		return []*db.Room{}, nil
	}
	return res.rooms, nil
}

func (s *DB) AddRoom(rm *db.Room) (db.RoomID, error) {
	type result struct {
		id  db.RoomID
		err error
	}

	resChan := make(chan *result)
	s.dbChan <- func(sdb *sql.DB) {
		tx, err := sdb.Begin()
		if err != nil {
			resChan <- &result{err: err}
			return
		}
		defer tx.Rollback()

		id, err := s.uniqueID(tx)
		if err != nil {
			resChan <- &result{err: err}
			return
		}

		rBytes, err := rotatorBytes(db.NewRotator(rm.RotatorType))
		if err != nil {
			resChan <- &result{err: err}
			return
		}

		_, err = tx.Exec(addRoomStmt, string(id), rm.DisplayName, normalize(rm.DisplayName), rBytes, rm.RotatorType)
		if err != nil {
			resChan <- &result{err: err}
			return
		}

		// Create an empty history list.
		teBytes, err := trackEntryBytes([]*db.TrackEntry{})
		if err != nil {
			resChan <- &result{err: err}
			return
		}
		if _, err := tx.Exec(createHistoryStmt, string(id), teBytes); err != nil {
			resChan <- &result{err: err}
			return
		}
		if err := tx.Commit(); err != nil {
			resChan <- &result{err: err}
			return
		}
		resChan <- &result{id: id}
	}

	res := <-resChan
	if res.err != nil {
		return db.RoomID(""), res.err
	}
	return res.id, nil
}

func rotatorBytes(r db.Rotator) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&r)
	return buf.Bytes(), err
}

func (s *DB) AddUserToRoom(rid db.RoomID, uid db.UserID) error {
	errChan := make(chan error)
	s.dbChan <- func(sdb *sql.DB) {
		tx, err := sdb.Begin()
		if err != nil {
			errChan <- err
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec(addQueueStmt, string(rid), uid.String())
		if err != nil {
			errChan <- err
			return
		}

		if err := addToRotator(tx, rid); err != nil {
			errChan <- err
			return
		}
		errChan <- tx.Commit()
	}
	return <-errChan
}

func addToRotator(tx *sql.Tx, rID db.RoomID) error {
	rot, err := loadRotator(tx, rID)
	if err != nil {
		return err
	}
	rot.Add()
	rBytes, err := rotatorBytes(rot)
	if err != nil {
		return err
	}
	_, err = tx.Exec(updateRotatorStmt, rBytes, string(rID))
	return err
}

func trackEntryBytes(entries []*db.TrackEntry) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(entries)
	return buf.Bytes(), err
}

func (s *DB) User(id db.UserID) (*db.User, error) {
	type result struct {
		user *db.User
		err  error
	}
	uChan := make(chan *result)
	s.dbChan <- func(sdb *sql.DB) {
		u, err := loadUser(sdb.QueryRow(getUserStmt, id.String()))
		uChan <- &result{user: u, err: err}
	}
	res := <-uChan
	if res.err == sql.ErrNoRows {
		return nil, db.ErrUserNotFound
	}
	return res.user, nil
}

func (s *DB) Users(rid db.RoomID) ([]*db.User, error) {
	type result struct {
		users []*db.User
		err   error
	}
	uChan := make(chan *result)
	s.dbChan <- func(sdb *sql.DB) {
		tx, err := sdb.Begin()
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

func (s *DB) AddUser(user *db.User) error {
	errChan := make(chan error)
	s.dbChan <- func(sdb *sql.DB) {
		_, err := sdb.Exec(addUserStmt, user.ID.String(), user.First, user.Last)
		errChan <- err
	}
	return <-errChan
}

func (s *DB) Tracks(qID db.QueueID, qo *db.QueueOptions) ([]db.QueueTrack, error) {
	type result struct {
		qts []db.QueueTrack
		err error
	}

	qChan := make(chan *result)
	s.dbChan <- func(sdb *sql.DB) {
		tx, err := sdb.Begin()
		if err != nil {
			qChan <- &result{err: err}
			return
		}
		defer tx.Rollback()

		qts, err := loadTrackList(tx, qID, qo)
		if err != nil {
			qChan <- &result{err: err}
			return
		}

		if err := tx.Commit(); err != nil {
			qChan <- &result{err: err}
			return
		}
		qChan <- &result{qts: qts}
	}
	res := <-qChan
	if res.err == sql.ErrNoRows {
		return nil, db.ErrQueueNotFound
	}
	if res.err != nil {
		return nil, fmt.Errorf("failed to load queue: %v", res.err)
	}
	return res.qts, nil
}

// AddTrack adds a track after the given qtID. If a blank ID is given, the song
// is added first. The song can't be added before a song that's already played.
func (s *DB) AddTrack(qID db.QueueID, track radio.Track, afterQTID string) error {
	errChan := make(chan error)
	s.dbChan <- func(sdb *sql.DB) {
		tx, err := sdb.Begin()
		if err != nil {
			errChan <- err
			return
		}
		defer tx.Rollback()

		var (
			id     = s.randomTrackID()
			prevID sql.NullString
			nextID sql.NullString
		)

		if afterQTID == "" {
			// This means they want to insert the track first. First, we look for an existing first track.
			var firstTrackID string
			if err := tx.QueryRow(getFirstQueueTrackStmt, string(qID.RoomID), qID.UserID.String()).Scan(&firstTrackID); err == sql.ErrNoRows {
				// There are no tracks in the queue, we're the first. We can leave
				// prevID and nextID as null.
			} else if err != nil {
				errChan <- err
				return
			} else {
				// No error, we found our first track. Make it the second track now.
				nextID.Valid = true
				nextID.String = firstTrackID
			}
		} else {
			// Load the track we want to insert our new track after.
			prevTrack, err := loadQueueTrack(tx, afterQTID)
			if err != nil {
				errChan <- err
				return
			}
			prevID.Valid = true
			prevID.String = afterQTID
			nextID = prevTrack.nextID
		}

		var nextQueueTrackID sql.NullString
		if err := tx.QueryRow(getQueueStmt, string(qID.RoomID), qID.UserID.String()).Scan(&nextQueueTrackID); err == sql.ErrNoRows {
			errChan <- err
			return
		}

		if nextID.Valid {
			// Before we insert the track, look up the track after us and make sure it
			// hasn't played yet.
			nextTrack, err := loadQueueTrack(tx, nextID.String)
			if err != nil {
				errChan <- err
				return
			}

			if nextTrack.played {
				errChan <- errors.New("can't add song before one that's already played")
				return
			}
		}

		// Insert the track, and once that's successful, start updating the
		// surrounding tracks.
		if _, err := tx.Exec(addQueueTrackStmt, id, prevID, nextID, track.ID, string(qID.RoomID), qID.UserID.String()); err != nil {
			errChan <- err
			return
		}

		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(track); err != nil {
			errChan <- err
			return
		}

		if _, err := tx.Exec(addTrackStmt, track.ID, buf.Bytes()); err != nil {
			errChan <- err
			return
		}

		// If there's no next track to be played, we need to set ourselves as the
		// next track.
		if !nextQueueTrackID.Valid {
			if _, err := tx.Exec(updateNextTrackStmt, id, string(qID.RoomID), qID.UserID.String()); err != nil {
				errChan <- err
				return
			}
		}

		// If we have a track before this track, update it to point to this track.
		if prevID.Valid {
			if _, err := tx.Exec(setQueueTrackNextStmt, id, prevID.String); err != nil {
				errChan <- err
				return
			}
		}

		// If we have a track after this track, update it to point to this track.
		if nextID.Valid {
			if _, err := tx.Exec(setQueueTrackPreviousStmt, id, nextID.String); err != nil {
				errChan <- err
				return
			}
		}

		errChan <- tx.Commit()
	}
	return <-errChan
}

type queueTrack struct {
	id     string
	prevID sql.NullString
	nextID sql.NullString
	played bool
}

func loadQueueTrack(tx *sql.Tx, id string) (*queueTrack, error) {
	qt := queueTrack{id: id}
	if err := tx.QueryRow(getQueueTrackStmt, id).Scan(&qt.prevID, &qt.nextID, &qt.played); err != nil {
		return nil, err
	}

	return &qt, nil
}

// RemoveTrack remove a given track from a queue. To do it, we find the
// QueueTrack in question, get it's previous/next tracks, and update their
// pointers to each other.
func (s *DB) RemoveTrack(qid db.QueueID, qtID string) error {
	errChan := make(chan error)
	s.dbChan <- func(sdb *sql.DB) {
		tx, err := sdb.Begin()
		if err != nil {
			errChan <- err
			return
		}
		defer tx.Rollback()

		var (
			prevID sql.NullString
			nextID sql.NullString
			played bool
		)

		if err := tx.QueryRow(getQueueTrackStmt, qtID).Scan(&prevID, &nextID, &played); err != nil {
			errChan <- err
			return
		}

		if played {
			errChan <- errors.New("can't delete songs that have already played")
			return
		}

		// If there was a previous song, we need to update it to point to the next
		// song, or nothing.
		if prevID.Valid {
			// If the song we're removing was the last song.
			var arg interface{} = nil
			// If the song we're removing wasn't the last song
			if nextID.Valid {
				arg = nextID.String
			}
			if _, err := tx.Exec(setQueueTrackNextStmt, arg, prevID.String); err != nil {
				errChan <- err
				return
			}
		}

		// If there was a next song, we need to update it to point to the previous
		// song, or nothing.
		if nextID.Valid {
			// If the song we're removing was the first song.
			var arg interface{} = nil
			// If the song we're removing wasn't the last song
			if prevID.Valid {
				arg = prevID.String
			}
			if _, err := tx.Exec(setQueueTrackPreviousStmt, arg, nextID.String); err != nil {
				errChan <- err
				return
			}
		}

		if _, err := tx.Exec(removeQueueTrackStmt, qtID); err != nil {
			errChan <- err
			return
		}

		errChan <- tx.Commit()
	}
	return <-errChan
}

func (s *DB) History(rid db.RoomID) ([]*db.TrackEntry, error) {
	type result struct {
		tracks []*db.TrackEntry
		err    error
	}
	hChan := make(chan *result)
	s.dbChan <- func(sdb *sql.DB) {
		ts, err := loadTrackEntries(sdb.QueryRow(getHistoryStmt, string(rid)))
		hChan <- &result{tracks: ts, err: err}
	}
	res := <-hChan
	if res.err == sql.ErrNoRows {
		return nil, db.ErrRoomNotFound
	}
	if res.err != nil {
		return nil, fmt.Errorf("failed to load history: %v", res.err)
	}
	return res.tracks, nil
}

func (s *DB) AddToHistory(rid db.RoomID, te *db.TrackEntry) error {
	errChan := make(chan error)
	s.dbChan <- func(sdb *sql.DB) {
		tx, err := sdb.Begin()
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

func (s *DB) MarkVetoed(rid db.RoomID, uid db.UserID) error {
	errChan := make(chan error)
	s.dbChan <- func(sdb *sql.DB) {
		tx, err := sdb.Begin()
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

func (s *DB) uniqueID(tx *sql.Tx) (db.RoomID, error) {
	i := 0
	var id string
	for {
		id = s.randomID()
		var n int
		if err := tx.QueryRow(roomExistsStmt, id).Scan(&n); err != nil {
			return db.RoomID(""), err
		}
		if n == 0 {
			break
		}
		i++
		if i >= 100 {
			return db.RoomID(""), errors.New("tried 100 random IDs, all were taken, which seems fishy")
		}
	}
	return db.RoomID(id), nil
}

func (s *DB) Close() error {
	close(s.doneChan)
	return s.closeFn()
}

func sqlInput(n int) string {
	if n <= 0 {
		return ""
	}
	return "?" + strings.Repeat(", ?", n-1)
}

// For now, normalizing a room name is just lowercasing it, to make searching
// easier.
func normalize(s string) string {
	return strings.ToLower(s)
}

var trackLetters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func (s *DB) randomTrackID() string {
	b := make([]byte, 64)
	r := rand.New(s.src)
	for i := range b {
		b[i] = trackLetters[r.Intn(len(trackLetters))]
	}
	return string(b)
}

var letters = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (s *DB) randomID() string {
	b := make([]byte, 4)
	r := rand.New(s.src)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

type CryptoRandSource struct{}

func (CryptoRandSource) Int63() int64 {
	var buf [8]byte
	_, err := cryptorand.Read(buf[:])
	if err != nil {
		panic(err)
	}
	return int64(buf[0]) |
		int64(buf[1])<<8 |
		int64(buf[2])<<16 |
		int64(buf[3])<<24 |
		int64(buf[4])<<32 |
		int64(buf[5])<<40 |
		int64(buf[6])<<48 |
		int64(buf[7]&0x7f)<<56
}

func (CryptoRandSource) Seed(int64) {}
