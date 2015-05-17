package db

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	User     string
	DBName   string
	Password string
	SSLMode  string
}

func NewConfig() (*Config, error) {
	j, err := ioutil.ReadFile("db/config.json")
	if err != nil {
		return nil, err
	}

	config := new(Config)
	err = json.Unmarshal(j, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
