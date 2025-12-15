package domain

import (
	"time"
)

// RepoInfo contains cached repository information.
type RepoInfo struct {
	Key        string    `json:"key"`
	Name       string    `json:"name"`
	LastUpdate time.Time `json:"last_update"`
	Versions   []string  `json:"versions"`
}

// NewRepoInfo creates a new RepoInfo for a dependency.
func NewRepoInfo(dep Dependency, versions []string) *RepoInfo {
	return &RepoInfo{
		Key:        dep.HashName(),
		Name:       dep.Name(),
		Versions:   versions,
		LastUpdate: time.Now(),
	}
}
