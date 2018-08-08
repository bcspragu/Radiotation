package db_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/bcspragu/Radiotation/db"
	"github.com/bcspragu/Radiotation/memdb"
	"github.com/bcspragu/Radiotation/radio"
	"github.com/bcspragu/Radiotation/rng"
	"github.com/bcspragu/Radiotation/sqldb"
	"github.com/google/go-cmp/cmp"
	"github.com/pressly/goose"

	// Init DB drivers.
	_ "github.com/mattn/go-sqlite3"
)

func TestRoomDoesntExist(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) { testRoomDoesntExist(t, newSQLDB) })
	t.Run("MemDB", func(t *testing.T) { testRoomDoesntExist(t, newMemDB) })
}

func testRoomDoesntExist(t *testing.T, newDB func(*testing.T) (db.DB, closeFn)) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	r, err := sdb.Room(db.RoomID("NOTA"))
	if r != nil {
		t.Errorf("Room(\"NOTA\") = %v, %v", r, nil)
	}

	if err != db.ErrRoomNotFound {
		t.Errorf("Room(\"NOTA\"): %v, wanted %v", err, db.ErrRoomNotFound)
	}
}

func TestAddRoom(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) { testAddRoom(t, newSQLDB) })
	t.Run("MemDB", func(t *testing.T) { testAddRoom(t, newMemDB) })
}

func testAddRoom(t *testing.T, newDB func(*testing.T) (db.DB, closeFn)) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	rID, err := sdb.AddRoom(&db.Room{DisplayName: "Test Room", RotatorType: db.RoundRobin})
	if err != nil {
		t.Errorf("AddRoom(): %v", err)
	}

	r, err := sdb.Room(rID)
	if err != nil {
		t.Errorf("Room(): %v", err)
	}

	if r.DisplayName != "Test Room" {
		t.Errorf("DisplayName = %q, want \"Test Room\"", r.DisplayName)
	}

	if r.ID != rID {
		t.Errorf("ID = %q, want %q", r.ID, rID)
	}

	if r.RotatorType != db.RoundRobin {
		t.Errorf("RotatorType = %q, want \"Random\"", r.RotatorType)
	}
}

func TestSearchRooms(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) { testSearchRooms(t, newSQLDB) })
	t.Run("MemDB", func(t *testing.T) { testSearchRooms(t, newMemDB) })
}

func testSearchRooms(t *testing.T, newDB func(*testing.T) (db.DB, closeFn)) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	rooms := []string{"Room One", "Room Two", "Another One", "Some Guy's Room"}

	for _, name := range rooms {
		if _, err := sdb.AddRoom(&db.Room{DisplayName: name, RotatorType: db.RoundRobin}); err != nil {
			t.Errorf("AddRoom(): %v", err)
		}
	}

	tests := []struct {
		query string
		want  []string
	}{
		{"Room", []string{"Room One", "Room Two", "Some Guy's Room"}},
		{"room", []string{"Room One", "Room Two", "Some Guy's Room"}},
		{"oNe", []string{"Another One", "Room One"}},
		{"TWO", []string{"Room Two"}},
		{"guy", []string{"Some Guy's Room"}},
		{"", []string{}},
		{"llaswd", []string{}},
	}

	for _, tc := range tests {
		rooms, err := sdb.SearchRooms(tc.query)
		if err != nil {
			t.Errorf("SearchRooms: %v", err)
			continue
		}

		names := []string{}
		for _, r := range rooms {
			names = append(names, r.DisplayName)
		}
		sort.Strings(names)

		if diff := cmp.Diff(tc.want, names); diff != "" {
			t.Errorf("SearchRooms(%q) (-want +got): \n%s", tc.query, diff)
		}
	}
}

func TestAddUserToRoom(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) { testAddUserToRoom(t, newSQLDB) })
	t.Run("MemDB", func(t *testing.T) { testAddUserToRoom(t, newMemDB) })
}

func testAddUserToRoom(t *testing.T, newDB func(*testing.T) (db.DB, closeFn)) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	rID, err := sdb.AddRoom(&db.Room{DisplayName: "Test Room", RotatorType: db.RoundRobin})
	if err != nil {
		t.Fatalf("AddRoom(): %v", err)
	}

	users, err := sdb.Users(rID)
	if err != nil {
		t.Fatalf("Users(): %v", err)
	}

	userCount(t, users, 0)

	var (
		uID1  = db.UserID{AccountType: db.GoogleAccount, ID: "testid"}
		user1 = &db.User{ID: uID1, First: "Test", Last: "Name"}
	)
	if err := sdb.AddUser(user1); err != nil {
		t.Fatalf("AddUser(): %v", err)
	}

	if err := sdb.AddUserToRoom(rID, uID1); err != nil {
		t.Fatalf("AddUserToRoom(): %v", err)
	}

	users, err = sdb.Users(rID)
	if err != nil {
		t.Fatalf("Users(): %v", err)
	}

	userCount(t, users, 1)
	userEquals(t, users[0], user1)

	var (
		uID2  = db.UserID{AccountType: db.GoogleAccount, ID: "testid2"}
		user2 = &db.User{ID: uID2, First: "Another", Last: "Test"}
	)

	if err := sdb.AddUser(user2); err != nil {
		t.Fatalf("AddUser(): %v", err)
	}

	if err := sdb.AddUserToRoom(rID, uID2); err != nil {
		t.Fatalf("AddUserToRoom(): %v", err)
	}

	users, err = sdb.Users(rID)
	if err != nil {
		t.Fatalf("Users(): %v", err)
	}

	userCount(t, users, 2)
}

func TestUser(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) { testUser(t, newSQLDB) })
	t.Run("MemDB", func(t *testing.T) { testUser(t, newMemDB) })
}

func testUser(t *testing.T, newDB func(*testing.T) (db.DB, closeFn)) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	uID := db.UserID{AccountType: db.GoogleAccount, ID: "testid"}
	wantUser := &db.User{ID: uID, First: "Test", Last: "Name"}
	if err := sdb.AddUser(wantUser); err != nil {
		t.Fatalf("AddUser(): %v", err)
	}

	gotUser, err := sdb.User(uID)
	if err != nil {
		t.Fatalf("User(): %v", err)
	}

	userEquals(t, gotUser, wantUser)
}

func TestTracks(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) { testTracks(t, newSQLDB) })
	t.Run("MemDB", func(t *testing.T) { testTracks(t, newMemDB) })
}

func testTracks(t *testing.T, newDB func(*testing.T) (db.DB, closeFn)) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	rID, err := sdb.AddRoom(&db.Room{DisplayName: "Test Room", RotatorType: db.RoundRobin})
	if err != nil {
		t.Fatalf("AddRoom(): %v", err)
	}

	uID := db.UserID{AccountType: db.GoogleAccount, ID: "testid"}
	user := &db.User{ID: uID, First: "Test", Last: "Name"}

	if err := sdb.AddUser(user); err != nil {
		t.Fatalf("AddUser(): %v", err)
	}

	if err := sdb.AddUserToRoom(rID, uID); err != nil {
		t.Fatalf("AddUserToRoom(): %v", err)
	}

	qID := db.QueueID{RoomID: rID, UserID: uID}
	track1 := &radio.Track{
		ID:      "testID1",
		Name:    "Test Track1",
		Artists: []radio.Artist{radio.Artist{Name: "Test Artist1"}},
	}
	if err := sdb.AddTrack(qID, track1, ""); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	ts, err := sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	trackCount(t, ts, 1)
	trackEquals(t, ts[0].Track, track1)
	trackNotPlayed(t, ts[0])

	track2 := &radio.Track{
		ID:      "testID2",
		Name:    "Test Track2",
		Artists: []radio.Artist{radio.Artist{Name: "Test Artist2"}},
	}
	if err := sdb.AddTrack(qID, track2, ts[0].ID); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	trackCount(t, ts, 2)
	trackEquals(t, ts[0].Track, track1)
	trackNotPlayed(t, ts[0])
	trackEquals(t, ts[1].Track, track2)
	trackNotPlayed(t, ts[1])
}

func TestNextTrackOneUser(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) { testNextTrackOneUser(t, newSQLDB) })
	t.Run("MemDB", func(t *testing.T) { testNextTrackOneUser(t, newMemDB) })
}

func testNextTrackOneUser(t *testing.T, newDB func(*testing.T) (db.DB, closeFn)) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	rID, err := sdb.AddRoom(&db.Room{DisplayName: "Test Room", RotatorType: db.RoundRobin})
	if err != nil {
		t.Fatalf("AddRoom(): %v", err)
	}

	uID := db.UserID{AccountType: db.GoogleAccount, ID: "testid"}
	user := &db.User{ID: uID, First: "Test", Last: "Name"}

	if err := sdb.AddUser(user); err != nil {
		t.Fatalf("AddUser(): %v", err)
	}

	if err := sdb.AddUserToRoom(rID, uID); err != nil {
		t.Fatalf("AddUserToRoom(): %v", err)
	}

	qID := db.QueueID{RoomID: rID, UserID: uID}
	track1 := &radio.Track{
		ID:      "testID1",
		Name:    "Test Track1",
		Artists: []radio.Artist{radio.Artist{Name: "Test Artist1"}},
	}
	if err := sdb.AddTrack(qID, track1, ""); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	ts, err := sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	track2 := &radio.Track{
		ID:      "testID2",
		Name:    "Test Track2",
		Artists: []radio.Artist{radio.Artist{Name: "Test Artist2"}},
	}
	if err := sdb.AddTrack(qID, track2, ts[0].ID); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	// Pull the first track from the queue.
	gotUser, gotTrack, err := sdb.NextTrack(rID)
	if err != nil {
		t.Fatalf("NextTrack(): %v", err)
	}

	// Make sure we got the right/only user.
	userEquals(t, gotUser, user)
	// Make sure we got the first track.
	trackEquals(t, gotTrack, track1)

	// Load our track list, make sure both tracks are still there and one is
	// played.
	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	trackCount(t, ts, 2)
	trackPlayed(t, ts[0])
	trackNotPlayed(t, ts[1])

	// Pull the second track from the queue.
	gotUser, gotTrack, err = sdb.NextTrack(rID)
	if err != nil {
		t.Fatalf("NextTrack(): %v", err)
	}

	userEquals(t, gotUser, user)
	trackEquals(t, gotTrack, track2)

	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	trackCount(t, ts, 2)
	trackPlayed(t, ts[0])
	trackPlayed(t, ts[1])

	// Pull the next track from the queue, which shouldn't exist.
	gotUser, gotTrack, err = sdb.NextTrack(rID)
	if err != db.ErrNoTracksInQueue {
		t.Errorf("NextTrack() got %v, want %v", err, db.ErrNoTracksInQueue)
	}

	userEquals(t, gotUser, nil)
	trackEquals(t, gotTrack, nil)
}

func TestNextTrackMultipleUsers(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) { testNextTrackMultipleUsers(t, newSQLDB) })
	t.Run("MemDB", func(t *testing.T) { testNextTrackMultipleUsers(t, newMemDB) })
}

func testNextTrackMultipleUsers(t *testing.T, newDB func(*testing.T) (db.DB, closeFn)) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	rID, err := sdb.AddRoom(&db.Room{DisplayName: "Test Room", RotatorType: db.RoundRobin})
	if err != nil {
		t.Fatalf("AddRoom(): %v", err)
	}

	var users []*db.User
	for i := 0; i < 3; i++ {
		users = append(users, &db.User{
			ID:    db.UserID{AccountType: db.GoogleAccount, ID: fmt.Sprintf("testid%d", i)},
			First: fmt.Sprintf("Test %d", i),
			Last:  fmt.Sprintf("Name %d", i),
		})
	}

	for _, u := range users {
		if err := sdb.AddUser(u); err != nil {
			t.Fatalf("AddUser(): %v", err)
		}
	}

	for _, u := range users {
		if err := sdb.AddUserToRoom(rID, u.ID); err != nil {
			t.Fatalf("AddUserToRoom(): %v", err)
		}
	}

	var tracks []*radio.Track
	for i := 0; i < 9; i++ {
		tracks = append(tracks, &radio.Track{
			ID:      fmt.Sprintf("testID%d", i),
			Name:    fmt.Sprintf("Test Track %d", i),
			Artists: []radio.Artist{radio.Artist{Name: fmt.Sprintf("Test Artist %d", i)}},
		})
	}

	for i, u := range users {
		for j, track := range tracks {
			if j%len(users) == i {
				if err := sdb.AddTrack(db.QueueID{RoomID: rID, UserID: u.ID}, track, ""); err != nil {
					t.Fatalf("AddTrack(): %v", err)
				}
			}
		}
	}

	// Since the round robin rotator tries to insert later users in the middle
	// somewhere, our users end up ordered 0 2 1 instead of 0 1 2.
	orderedUsers := []*db.User{users[0], users[2], users[1]}
	// Since we insert tracks at the head of the list (for simplicity, because
	// adding at the end requires a QueueTrack.id), we effectively add the songs
	// in reverse order.
	orderedTracks := []*radio.Track{
		tracks[6],
		tracks[8],
		tracks[7],
		tracks[3],
		tracks[5],
		tracks[4],
		tracks[0],
		tracks[2],
		tracks[1],
	}

	for i := 0; i < 9; i++ {
		gotUser, gotTrack, err := sdb.NextTrack(rID)
		if err != nil {
			t.Fatalf("NextTrack(): %v", err)
		}

		// Make sure we got the right user.
		userEquals(t, gotUser, orderedUsers[i%len(orderedUsers)])
		// Make sure we got the right track.
		trackEquals(t, gotTrack, orderedTracks[i])
	}
}

func TestAddTrack(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) { testAddTrack(t, newSQLDB) })
	t.Run("MemDB", func(t *testing.T) { testAddTrack(t, newMemDB) })
}

func testAddTrack(t *testing.T, newDB func(*testing.T) (db.DB, closeFn)) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	rID, err := sdb.AddRoom(&db.Room{DisplayName: "Test Room", RotatorType: db.RoundRobin})
	if err != nil {
		t.Fatalf("AddRoom(): %v", err)
	}

	user := &db.User{
		ID:    db.UserID{AccountType: db.GoogleAccount, ID: "testid"},
		First: "Test",
		Last:  "Name",
	}
	if err := sdb.AddUser(user); err != nil {
		t.Fatalf("AddUser(): %v", err)
	}

	if err := sdb.AddUserToRoom(rID, user.ID); err != nil {
		t.Fatalf("AddUserToRoom(): %v", err)
	}

	var tracks []*radio.Track
	for i := 0; i < 4; i++ {
		tracks = append(tracks, &radio.Track{
			ID:      fmt.Sprintf("testID%d", i),
			Name:    fmt.Sprintf("Test Track %d", i),
			Artists: []radio.Artist{radio.Artist{Name: fmt.Sprintf("Test Artist %d", i)}},
		})
	}

	qID := db.QueueID{RoomID: rID, UserID: user.ID}

	// Add a track first.
	if err := sdb.AddTrack(qID, tracks[0], ""); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	ts, err := sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	trackCount(t, ts, 1)
	trackEquals(t, ts[0].Track, tracks[0])

	// Add a track after.
	if err := sdb.AddTrack(qID, tracks[1], ts[0].ID); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	trackCount(t, ts, 2)
	trackEquals(t, ts[0].Track, tracks[0])
	trackEquals(t, ts[1].Track, tracks[1])

	// Add a track in the middle.
	if err := sdb.AddTrack(qID, tracks[2], ts[0].ID); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	trackCount(t, ts, 3)
	trackEquals(t, ts[0].Track, tracks[0])
	trackEquals(t, ts[1].Track, tracks[2])
	trackEquals(t, ts[2].Track, tracks[1])

	// Add a track at the end.
	if err := sdb.AddTrack(qID, tracks[3], ts[2].ID); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	trackCount(t, ts, 4)
	trackEquals(t, ts[0].Track, tracks[0])
	trackEquals(t, ts[1].Track, tracks[2])
	trackEquals(t, ts[2].Track, tracks[1])
	trackEquals(t, ts[3].Track, tracks[3])
}

func TestRemoveTrack(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) { testRemoveTrack(t, newSQLDB) })
	t.Run("MemDB", func(t *testing.T) { testRemoveTrack(t, newMemDB) })
}

func testRemoveTrack(t *testing.T, newDB func(*testing.T) (db.DB, closeFn)) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	rID, err := sdb.AddRoom(&db.Room{DisplayName: "Test Room", RotatorType: db.RoundRobin})
	if err != nil {
		t.Fatalf("AddRoom(): %v", err)
	}

	user := &db.User{
		ID:    db.UserID{AccountType: db.GoogleAccount, ID: "testid"},
		First: "Test",
		Last:  "Name",
	}
	if err := sdb.AddUser(user); err != nil {
		t.Fatalf("AddUser(): %v", err)
	}

	if err := sdb.AddUserToRoom(rID, user.ID); err != nil {
		t.Fatalf("AddUserToRoom(): %v", err)
	}

	var tracks []*radio.Track
	for i := 0; i < 4; i++ {
		tracks = append(tracks, &radio.Track{
			ID:      fmt.Sprintf("testID%d", i),
			Name:    fmt.Sprintf("Test Track %d", i),
			Artists: []radio.Artist{radio.Artist{Name: fmt.Sprintf("Test Artist %d", i)}},
		})
	}

	qID := db.QueueID{RoomID: rID, UserID: user.ID}

	// Add a track first.
	if err := sdb.AddTrack(qID, tracks[0], ""); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	ts, err := sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	// Add a track after.
	if err := sdb.AddTrack(qID, tracks[1], ts[0].ID); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	// Add a track after that.
	if err := sdb.AddTrack(qID, tracks[2], ts[1].ID); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	// Add one last track.
	if err := sdb.AddTrack(qID, tracks[3], ts[2].ID); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	// Make sure the tracks are where we expect them.
	trackCount(t, ts, 4)
	trackEquals(t, ts[0].Track, tracks[0])
	trackEquals(t, ts[1].Track, tracks[1])
	trackEquals(t, ts[2].Track, tracks[2])
	trackEquals(t, ts[3].Track, tracks[3])

	// Remove the first track.
	if err := sdb.RemoveTrack(qID, ts[0].ID); err != nil {
		t.Fatalf("RemoveTrack(): %v", err)
	}

	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	trackCount(t, ts, 3)
	trackEquals(t, ts[0].Track, tracks[1])
	trackEquals(t, ts[1].Track, tracks[2])
	trackEquals(t, ts[2].Track, tracks[3])

	// Remove the middle track.
	if err := sdb.RemoveTrack(qID, ts[1].ID); err != nil {
		t.Fatalf("RemoveTrack(): %v", err)
	}

	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	trackCount(t, ts, 2)
	trackEquals(t, ts[0].Track, tracks[1])
	trackEquals(t, ts[1].Track, tracks[3])

	// Remove the last track in the queue.
	if err := sdb.RemoveTrack(qID, ts[1].ID); err != nil {
		t.Fatalf("RemoveTrack(): %v", err)
	}

	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	trackCount(t, ts, 1)
	trackEquals(t, ts[0].Track, tracks[1])

	// Remove the only track.
	if err := sdb.RemoveTrack(qID, ts[0].ID); err != nil {
		t.Fatalf("RemoveTrack(): %v", err)
	}

	ts, err = sdb.Tracks(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("Tracks(): %v", err)
	}

	trackCount(t, ts, 0)
}

func TestHistory(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) { testHistory(t, newSQLDB) })
	t.Run("MemDB", func(t *testing.T) { testHistory(t, newMemDB) })
}

func testHistory(t *testing.T, newDB func(*testing.T) (db.DB, closeFn)) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	rID, err := sdb.AddRoom(&db.Room{DisplayName: "Test Room", RotatorType: db.RoundRobin})
	if err != nil {
		t.Fatalf("AddRoom(): %v", err)
	}

	uID := db.UserID{AccountType: db.GoogleAccount, ID: "testid"}
	wantUser := &db.User{ID: uID, First: "Test", Last: "Name"}
	if err := sdb.AddUser(wantUser); err != nil {
		t.Fatalf("AddUser(): %v", err)
	}

	if err := sdb.AddUserToRoom(rID, uID); err != nil {
		t.Fatalf("AddUserToRoom(): %v", err)
	}

	var tracks []*db.TrackEntry
	for i := 0; i < 3; i++ {
		tracks = append(tracks, &db.TrackEntry{
			UserID: uID,
			Track: &radio.Track{
				ID:      fmt.Sprintf("testID%d", i),
				Name:    fmt.Sprintf("Test Track %d", i),
				Artists: []radio.Artist{radio.Artist{Name: fmt.Sprintf("Test Artist %d", i)}},
			},
		})
	}

	tes, err := sdb.History(rID)
	if err != nil {
		t.Fatalf("History(): %v: ", err)
	}

	trackEntryCount(t, tes, 0)

	if err := sdb.AddToHistory(rID, tracks[0]); err != nil {
		t.Fatalf("AddToHistory(): %v", err)
	}

	tes, err = sdb.History(rID)
	if err != nil {
		t.Fatalf("History(): %v: ", err)
	}

	trackEntryCount(t, tes, 1)
	trackEntryEquals(t, tes[0], tracks[0])

	if err := sdb.AddToHistory(rID, tracks[1]); err != nil {
		t.Fatalf("AddToHistory(): %v", err)
	}

	tes, err = sdb.History(rID)
	if err != nil {
		t.Fatalf("History(): %v: ", err)
	}

	trackEntryCount(t, tes, 2)
	trackEntryEquals(t, tes[0], tracks[0])
	trackEntryEquals(t, tes[1], tracks[1])

	if err := sdb.AddToHistory(rID, tracks[2]); err != nil {
		t.Fatalf("AddToHistory(): %v", err)
	}

	tes, err = sdb.History(rID)
	if err != nil {
		t.Fatalf("History(): %v: ", err)
	}

	trackEntryCount(t, tes, 3)
	trackEntryEquals(t, tes[0], tracks[0])
	trackEntryEquals(t, tes[1], tracks[1])
	trackEntryEquals(t, tes[2], tracks[2])
}

type closeFn func()

func trackEntryCount(t *testing.T, ts []*db.TrackEntry, want int) {
	t.Helper()
	if got := len(ts); got != want {
		t.Fatalf("Got %d tracks, want %d", got, want)
	}
}

func trackCount(t *testing.T, ts []*db.QueueTrack, want int) {
	t.Helper()
	if got := len(ts); got != want {
		t.Fatalf("Got %d tracks, want %d", got, want)
	}
}

func trackPlayed(t *testing.T, qt *db.QueueTrack) {
	t.Helper()
	if !qt.Played {
		t.Errorf("track %v not played", qt)
	}
}

func trackNotPlayed(t *testing.T, qt *db.QueueTrack) {
	t.Helper()
	if qt.Played {
		t.Errorf("track %v played", qt)
	}
}

func userCount(t *testing.T, us []*db.User, want int) {
	t.Helper()
	if got := len(us); got != want {
		t.Fatalf("Got %d users, want %d", got, want)
	}
}

func trackEquals(t *testing.T, gotTrack, wantTrack *radio.Track) {
	t.Helper()
	if diff := cmp.Diff(wantTrack, gotTrack); diff != "" {
		t.Errorf("Track (-want +got)\n%s", diff)
	}
}

func trackEntryEquals(t *testing.T, gotTrack, wantTrack *db.TrackEntry) {
	t.Helper()
	if diff := cmp.Diff(wantTrack, gotTrack); diff != "" {
		t.Errorf("Track (-want +got)\n%s", diff)
	}
}

func userEquals(t *testing.T, gotUser, wantUser *db.User) {
	t.Helper()
	if diff := cmp.Diff(wantUser, gotUser); diff != "" {
		t.Errorf("User (-want +got)\n%s", diff)
	}
}

func newMemDB(t *testing.T) (db.DB, closeFn) {
	db, err := memdb.New(rng.NewSource(0))
	if err != nil {
		t.Fatalf("failed to create memdb: %v", err)
	}

	return db, func() {}
}

func newSQLDB(t *testing.T) (db.DB, closeFn) {
	prefix := strings.Replace(t.Name(), "/", "", -1)
	name, err := ioutil.TempDir("", prefix)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	sdb, err := sqldb.New(filepath.Join(name, prefix+".db"), rng.NewSource(0))
	if err != nil {
		t.Fatalf("failed to create sqldb: %v", err)
	}

	goose.SetLogger(&testLogger{t: t, log: false})
	goose.SetDialect("sqlite3")
	if err := goose.Up(sdb.DB, "../sqldb/migrations"); err != nil {
		t.Fatalf("failed to apply migrations to db: %v", err)
	}

	return sdb, func() {
		if err := sdb.Close(); err != nil {
			t.Errorf("failed to close DB: %v", err)
		}

		if err := os.RemoveAll(name); err != nil {
			t.Errorf("failed to remove DB temp dir: %v", err)
		}
	}
}

type testLogger struct {
	t   *testing.T
	log bool
}

func (t *testLogger) Fatal(v ...interface{}) {
	if t.log {
		t.t.Fatal(v...)
	}
}

func (t *testLogger) Fatalf(format string, v ...interface{}) {
	if t.log {
		t.t.Fatalf(format, v...)
	}
}

func (t *testLogger) Print(v ...interface{}) {
	if t.log {
		t.t.Log(v...)
	}
}

func (t *testLogger) Println(v ...interface{}) {
	if t.log {
		t.t.Log(v, "\n")
	}
}

func (t *testLogger) Printf(format string, v ...interface{}) {
	if t.log {
		t.t.Logf(format, v...)
	}
}
