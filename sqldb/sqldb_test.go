package sqldb

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/bcspragu/Radiotation/db"
	"github.com/bcspragu/Radiotation/radio"
	"github.com/bcspragu/Radiotation/rng"
	"github.com/google/go-cmp/cmp"
	"github.com/pressly/goose"

	// Init DB drivers.
	_ "github.com/mattn/go-sqlite3"
)

func TestRoomDoesntExist(t *testing.T) {
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
	sdb, closeFn := newDB(t)
	defer closeFn()

	rID, err := sdb.AddRoom(&db.Room{DisplayName: "Test Room", RotatorType: db.Random})
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

	if r.RotatorType != db.Random {
		t.Errorf("RotatorType = %q, want \"Random\"", r.RotatorType)
	}
}

func TestSearchRooms(t *testing.T) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	rooms := []string{"Room One", "Room Two", "Another One", "Some Guy's Room"}

	for _, name := range rooms {
		if _, err := sdb.AddRoom(&db.Room{DisplayName: name, RotatorType: db.Random}); err != nil {
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
	sdb, closeFn := newDB(t)
	defer closeFn()

	rID, err := sdb.AddRoom(&db.Room{DisplayName: "Test Room", RotatorType: db.Random})
	if err != nil {
		t.Fatalf("AddRoom(): %v", err)
	}

	users, err := sdb.Users(rID)
	if err != nil {
		t.Fatalf("Users(): %v", err)
	}

	if len(users) != 0 {
		t.Errorf("Users() = %q, wanted none", users)
	}

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

	if len(users) != 1 {
		t.Fatalf("Users() = %q, wanted one user", users)
	}

	if diff := cmp.Diff(users[0], user1); diff != "" {
		t.Errorf("User (-want +got)\n%s", diff)
	}

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

	if len(users) != 2 {
		t.Fatalf("Users() = %q, wanted two users", users)
	}
}

func TestUser(t *testing.T) {
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

	if diff := cmp.Diff(wantUser, gotUser); diff != "" {
		t.Errorf("User (-want +got)\n%s", diff)
	}
}

func TestTrackList(t *testing.T) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	rID, err := sdb.AddRoom(&db.Room{DisplayName: "Test Room", RotatorType: db.Random})
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
	track1 := radio.Track{
		ID:      "testID1",
		Name:    "Test Track1",
		Artists: []radio.Artist{radio.Artist{Name: "Test Artist1"}},
	}
	track2 := radio.Track{
		ID:      "testID2",
		Name:    "Test Track2",
		Artists: []radio.Artist{radio.Artist{Name: "Test Artist2"}},
	}
	if err := sdb.AddTrack(qID, track1, ""); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	tl, err := sdb.TrackList(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("TrackList(): %v", err)
	}

	if c := len(tl.Tracks); c != 1 {
		t.Fatalf("Got %d tracks in track list, want %d", c, 1)
	}

	if diff := cmp.Diff(track1, tl.Tracks[0]); diff != "" {
		t.Errorf("Track (-want +got)\n%s", diff)
	}

	if err := sdb.AddTrack(qID, track2, tl.QueueTrackIDs[0]); err != nil {
		t.Fatalf("AddTrack(): %v", err)
	}

	tl, err = sdb.TrackList(qID, &db.QueueOptions{Type: db.AllTracks})
	if err != nil {
		t.Fatalf("TrackList(): %v", err)
	}

	if c := len(tl.Tracks); c != 2 {
		t.Fatalf("Got %d tracks in track list, want %d", c, 2)
	}

	if diff := cmp.Diff(track1, tl.Tracks[0]); diff != "" {
		t.Errorf("Track #1 (-want +got)\n%s", diff)
	}

	if diff := cmp.Diff(track2, tl.Tracks[1]); diff != "" {
		t.Errorf("Track #2 (-want +got)\n%s", diff)
	}
}

type closeFn func()

func newDB(t *testing.T) (db.DB, closeFn) {
	name, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	sdb, err := New(filepath.Join(name, t.Name()+".db"), rng.NewSource(0))
	if err != nil {
		t.Fatalf("failed to create sqldb: %v", err)
	}

	goose.SetLogger(&testLogger{t: t, log: false})
	goose.SetDialect("sqlite3")
	if err := goose.Up(sdb.sdb, "migrations"); err != nil {
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
