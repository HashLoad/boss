package ports

import (
	"github.com/hashload/boss/internal/infra"
)

// FileSystem is an alias for infra.FileSystem.
// This exists for backward compatibility with code that imports ports.FileSystem.
// New code should use infra.FileSystem directly.
type FileSystem = infra.FileSystem
