// Package cache provides caching functionality for repository information.
// It stores and retrieves repository metadata to avoid repeated network requests.
package cache

import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/infra"
	"github.com/hashload/boss/pkg/env"
)

// CacheService provides cache management operations.
//
//nolint:revive // cache.CacheService is intentional for clarity
type CacheService struct {
	fs infra.FileSystem
}

// NewCacheService creates a new cache service.
func NewCacheService(fs infra.FileSystem) *CacheService {
	return &CacheService{fs: fs}
}

// SaveRepositoryDetails saves repository details to cache.
func (s *CacheService) SaveRepositoryDetails(dep domain.Dependency, versions []string) error {
	location := env.GetCacheDir()
	data := &domain.RepoInfo{
		Key:        dep.HashName(),
		Name:       dep.Name(),
		Versions:   versions,
		LastUpdate: time.Now(),
	}

	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}

	infoPath := filepath.Join(location, "info")
	if err := s.fs.MkdirAll(infoPath, 0755); err != nil {
		return err
	}

	jsonFilePath := filepath.Join(infoPath, data.Key+".json")
	return s.fs.WriteFile(jsonFilePath, buff, 0644)
}

// LoadRepositoryData loads repository data from cache.
func (s *CacheService) LoadRepositoryData(key string) (*domain.RepoInfo, error) {
	location := env.GetCacheDir()
	cacheInfoPath := filepath.Join(location, "info", key+".json")

	data, err := s.fs.ReadFile(cacheInfoPath)
	if err != nil {
		return nil, err
	}

	var repoInfo domain.RepoInfo
	if err := json.Unmarshal(data, &repoInfo); err != nil {
		return nil, err
	}

	return &repoInfo, nil
}
