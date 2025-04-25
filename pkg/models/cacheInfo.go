package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

type RepoInfo struct {
	Key        string    `json:"key"`
	Name       string    `json:"name"`
	LastUpdate time.Time `json:"last_update"`
	Versions   []string  `json:"versions"`
	References []string  `json:"references"` // TODO: check it on garbage collector
}

func CacheRepositoryDetails(dep Dependency, versions []string) {
	location := env.GetCacheDir()
	data := &RepoInfo{
		Key:        dep.HashName(),
		Name:       dep.Name(),
		Versions:   versions,
		LastUpdate: time.Now(),
	}

	buff, err := json.Marshal(data)
	if err != nil {
		msg.Err(err.Error())
	}

	infoPath := filepath.Join(location, "info")
	err = os.MkdirAll(infoPath, 0750)
	if err != nil {
		msg.Err(err.Error())
	}

	jsonFilePath := filepath.Join(infoPath, data.Key+".json")
	jsonFile, err := os.Create(jsonFilePath)
	if err != nil {
		msg.Err(err.Error())
		return
	}
	defer jsonFile.Close()

	_, err = jsonFile.Write(buff)
	if err != nil {
		msg.Err(err.Error())
	}
}

func RepoData(key string) (*RepoInfo, error) {
	location := env.GetCacheDir()
	cacheRepository := &RepoInfo{}
	cacheInfoPath := filepath.Join(location, "info", key+".json")
	cacheInfoData, err := os.ReadFile(cacheInfoPath)
	if err != nil {
		return &RepoInfo{}, err
	}
	err = json.Unmarshal(cacheInfoData, cacheRepository)
	if err != nil {
		return &RepoInfo{}, err
	}
	return cacheRepository, nil
}
