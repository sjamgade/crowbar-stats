package config

import (
	"encoding/json"
	"io/ioutil"
)

// Config contains the configuration of the url shortener.
type Config struct {
	Server struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"server"`
	Redis struct {
		Host     string `json:"host"`
		Password string `json:"password"`
		DB       string `json:"db"`
	} `json:"redis"`
	Sqlite struct {
		Dbpath string `json:"dbpath"`
	} `json:"sqlite"`
	Postgres struct {
		Host     string `json:"host"`
		User     string `json:"user"`
		Password string `json:"password"`
		DB       string `json:"db"`
	} `json:"postgres"`
	Logfile string `json:"logfile"`
}

// FromFile returns a configuration parsed from the given file.
func FromFile(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
