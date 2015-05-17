package db

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
)

var configStr string
var d *sql.DB

var ErrRowCount = errors.New("Wrong number of rows updated")

func init() {
	config, err := NewConfig()
	if err != nil {
		fmt.Println("Failed to load db config:", err)
		panic(err)
	}

	configStr = fmt.Sprintf("user=%s dbname=%s", config.User, config.DBName)
	if config.Password != "" {
		configStr += " password=" + config.Password
	}
	if config.SSLMode != "" {
		configStr += " sslmode=" + config.SSLMode
	}
	database, err := sql.Open("postgres", configStr)
	if err != nil {
		fmt.Println("Failed to open SQL connection:", err)
		panic(err)
	}
	err = database.Ping()
	if err != nil {
		fmt.Println("Failed to open SQL connection:", err)
		panic(err)
	}

	d = database
}
