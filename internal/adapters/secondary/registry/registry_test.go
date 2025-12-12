package registryadapter_test

import (
	"testing"

	registry "github.com/hashload/boss/internal/adapters/secondary/registry"
)

// TestGetDelphiPaths tests retrieval of Delphi paths.
func TestGetDelphiPaths(_ *testing.T) {
	// This function relies on system registry, so we just ensure it doesn't panic
	paths := registry.GetDelphiPaths()

	// Result can be nil on non-Windows or without Delphi installed
	// Just verify it doesn't panic - paths can be nil on Linux
	_ = paths
}

// TestGetCurrentDelphiVersion tests retrieval of current Delphi version.
func TestGetCurrentDelphiVersion(_ *testing.T) {
	// This function relies on system registry and configuration
	// Just ensure it doesn't panic
	version := registry.GetCurrentDelphiVersion()

	// Result can be empty on non-Windows or without Delphi installed
	// Version is a string, could be empty
	_ = version
}
