package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type Location struct {
	Lat  float64
	Long float64
}

type User struct {
	ID      int
	Login   string
	Lat     sql.NullFloat64
	Long    sql.NullFloat64
	Online  bool
	Blocked bool
}

var ErrUserExists = errors.New("User exists")

func UserExists(name string) bool {
	var login string
	err := d.QueryRow("SELECT login FROM users WHERE login=$1", name).Scan(&login)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		fmt.Println("Error querying user existence:", err)
	default:
		return true
	}
	return false
}

func CreateUser(name string) error {
	if UserExists(name) {
		return ErrUserExists
	}
	res, err := d.Exec("INSERT INTO users (login) VALUES ($1)", name)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return ErrRowCount
	}

	return err
}

func GetUser(name string) (User, error) {
	u := User{}
	err := d.QueryRow("SELECT id, login, lat, long, blocked FROM users WHERE login=$1", name).Scan(&u.ID, &u.Login, &u.Lat, &u.Long, &u.Blocked)
	return u, err
}

func (u *User) Message(msg string) error {
	res, err := d.Exec("INSERT INTO messages (user_id, body) VALUES ($1, $2)", u.ID, msg)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return ErrRowCount
	}

	return err
}

func (u *User) Post(msg string) error {
	res, err := d.Exec("INSERT INTO posts (user_id, body) VALUES ($1, $2)", u.ID, msg)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return ErrRowCount
	}

	return err
}

func (u *User) Comment(msg string, postID int) error {
	res, err := d.Exec("INSERT INTO comments (user_id, post_id, body) VALUES ($1, $2, $3)", u.ID, postID, msg)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return ErrRowCount
	}

	return err
}

func (u *User) SetLocation(lat, long float64) error {
	res, err := d.Exec("UPDATE users SET (lat, long) = ($1, $2) WHERE id = $3", lat, long, u.ID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return ErrRowCount
	}

	return err
}

func Locations() []Location {
	locs := []Location{}
	rows, err := d.Query("SELECT lat, long FROM users WHERE lat IS NOT NULL AND long IS NOT NULL")
	if err != nil {
		fmt.Println(err)
		return locs
	}

	defer rows.Close()
	for rows.Next() {
		var loc Location
		if err := rows.Scan(&loc.Lat, &loc.Long); err != nil {
			fmt.Println(err)
			return locs
		}
		locs = append(locs, loc)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return locs
	}
	return locs
}
