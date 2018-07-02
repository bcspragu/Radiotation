package sqldb

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bcspragu/Radiotation/db"
	"github.com/bcspragu/Radiotation/rng"
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
		t.Errorf("Room(\"NOTA\") returned err %v, wanted %v", err, db.ErrRoomNotFound)
	}
}

func TestAddRoom(t *testing.T) {
	sdb, closeFn := newDB(t)
	defer closeFn()

	rID, err := sdb.AddRoom(&db.Room{DisplayName: "Test Room", RotatorType: db.Random})
	if err != nil {
		t.Errorf("AddRoom() returned err %v", err)
	}

	r, err := sdb.Room(rID)
	if err != nil {
		t.Errorf("Room() returned err %v", err)
	}
	t.Log(3)

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
