package models

import (
	"encoding/json"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/msg"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type RepoInfo struct {
	Key        string    `json:"key"`
	LastUpdate time.Time `json:"last_update"`
	Versions   []string  `json:"versions"`
}

func SaveRepoData(key string, versions []string) {
	location := env.GetCacheDir()
	data := &RepoInfo{}
	data.Key = key
	data.LastUpdate = time.Now()
	data.Versions = versions
	d, err := json.Marshal(data)
	if err != nil {
		msg.Err(err.Error())
	}

	pp := filepath.Join(location, "info")
	err = os.MkdirAll(pp, 0755)
	if err != nil {
		msg.Err(err.Error())
	}

	p := filepath.Join(pp, key+".json")
	f, err := os.Create(p)
	if err != nil {
		msg.Err(err.Error())
		return
	}
	defer f.Close()

	_, err = f.Write(d)
	if err != nil {
		msg.Err(err.Error())
	}
}

// RepoData retrieves cached information about a repo.
func RepoData(key string) (*RepoInfo, error) {
	location := env.GetCacheDir()
	c := &RepoInfo{}
	p := filepath.Join(location, "info", key+".json")
	f, err := ioutil.ReadFile(p)
	if err != nil {
		return &RepoInfo{}, err
	}
	err = json.Unmarshal(f, c)
	if err != nil {
		return &RepoInfo{}, err
	}
	return c, nil
}
