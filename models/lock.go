package models

import "time"

type Dependencyes struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Hash    string `json:"hash"`
}

type PackageLock struct {
	lastUpdate time.Time                 `json:"last_update"`
	installed  map[string]string         `json:"installed_modules"`
	tree       map[string][]Dependencyes `json:"tree"`
}
