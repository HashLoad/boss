package installer

import (
	"sync"
)

// DependencyCache tracks which dependencies have been updated in current session.
// Thread-safe implementation to replace global variable.
type DependencyCache struct {
	updated map[string]bool
	mu      sync.RWMutex
}

// NewDependencyCache creates a new DependencyCache instance.
func NewDependencyCache() *DependencyCache {
	return &DependencyCache{
		updated: make(map[string]bool),
	}
}

// IsUpdated checks if a dependency has been updated in current session.
func (c *DependencyCache) IsUpdated(hashName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.updated[hashName]
}

// MarkUpdated marks a dependency as updated in current session.
func (c *DependencyCache) MarkUpdated(hashName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.updated[hashName] = true
}
