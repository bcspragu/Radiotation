package sql

import (
	"database/sql"
	"errors"

	"github.com/bcspragu/chameleon/c7n"
)

const ()

type DB struct {
	dbChan   chan func(db *sql.DB)
	doneChan chan struct{}

	sql *sql.DB
}

func New() (*DB, error) {
	sdb, err := sql.Open("sqlite3", "chameleon.db")
	if err != nil {
		return nil, err
	}

	db := &DB{
		dbChan:   make(chan func(db *sql.DB)),
		doneChan: make(chan struct{}),
		sql:      sdb,
	}
	go db.run()

	return db, nil
}

func (db *DB) run() {
	for {
		select {
		case dbFn := <-db.dbChan:
			dbFn(db.sql)
		case <-db.doneChan:
			return
		}
	}
}

func (db *DB) Close() error {
	close(db.doneChan)
	db.sql.Close()
}

func (db *DB) NewGame(g *c7n.Game) (c7n.GameID, error) {
	if g.ID != "" {
		return c7n.GameID(""), errors.New("game ID should not be set when creating a game")
	}

}

func (db *DB) Game(c7n.GameID) (*c7n.Game, error) {

}

func (db *DB) AddPlayer(id c7n.GameID, name string) error {

}

func (db *DB) StartGame(c7n.GameID) {

}
