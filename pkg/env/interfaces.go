package env

import (
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

// ConfigProvider defines the interface for configuration access
// This allows dependency injection and easier testing
type ConfigProvider interface {
	GetDelphiPath() string
	GetGitEmbedded() bool
	GetAuth(repo string) transport.AuthMethod
	GetPurgeTime() int
	GetInternalRefreshRate() int
	GetLastPurge() time.Time
	GetLastInternalUpdate() time.Time
	GetConfigVersion() int64
	SetLastPurge(t time.Time)
	SetLastInternalUpdate(t time.Time)
	SetConfigVersion(version int64)
	SaveConfiguration()
}

// Ensure Configuration implements ConfigProvider
var _ ConfigProvider = (*Configuration)(nil)

// GetDelphiPath returns the Delphi path
func (c *Configuration) GetDelphiPath() string {
	return c.DelphiPath
}

// GetGitEmbedded returns whether to use embedded git
func (c *Configuration) GetGitEmbedded() bool {
	return c.GitEmbedded
}

// GetPurgeTime returns the purge time in days
func (c *Configuration) GetPurgeTime() int {
	return c.PurgeTime
}

// GetInternalRefreshRate returns the internal refresh rate
func (c *Configuration) GetInternalRefreshRate() int {
	return c.InternalRefreshRate
}

// GetLastPurge returns the last purge time
func (c *Configuration) GetLastPurge() time.Time {
	return c.LastPurge
}

// GetLastInternalUpdate returns the last internal update time
func (c *Configuration) GetLastInternalUpdate() time.Time {
	return c.LastInternalUpdate
}

// GetConfigVersion returns the configuration version
func (c *Configuration) GetConfigVersion() int64 {
	return c.ConfigVersion
}

// SetLastPurge sets the last purge time
func (c *Configuration) SetLastPurge(t time.Time) {
	c.LastPurge = t
}

// SetLastInternalUpdate sets the last internal update time
func (c *Configuration) SetLastInternalUpdate(t time.Time) {
	c.LastInternalUpdate = t
}

// SetConfigVersion sets the configuration version
func (c *Configuration) SetConfigVersion(version int64) {
	c.ConfigVersion = version
}
