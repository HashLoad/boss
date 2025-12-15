package version_test

import (
	"testing"

	"github.com/hashload/boss/internal/version"
)

func TestGetVersion(t *testing.T) {
	v := version.GetVersion()

	if v == "" {
		t.Error("GetVersion() should not return empty string")
	}

	// Default version should start with 'v'
	if v[0] != 'v' {
		t.Errorf("GetVersion() = %q, should start with 'v'", v)
	}
}

func TestGet(t *testing.T) {
	info := version.Get()

	if info.Version == "" {
		t.Error("BuildInfo.Version should not be empty")
	}

	// Version should start with 'v'
	if info.Version[0] != 'v' {
		t.Errorf("BuildInfo.Version = %q, should start with 'v'", info.Version)
	}
}

func TestBuildInfo_Structure(t *testing.T) {
	info := version.Get()

	// Verify the struct has the expected fields
	_ = info.Version
	_ = info.GitCommit
	_ = info.GoVersion

	// In test mode, GoVersion is cleared
	if info.GoVersion != "" {
		t.Logf("GoVersion = %q (expected empty in test mode)", info.GoVersion)
	}
}

func TestGetVersion_Format(t *testing.T) {
	v := version.GetVersion()

	// Should match semver format (at minimum v0.0.1)
	if len(v) < 6 { // "v0.0.1" is 6 characters
		t.Errorf("GetVersion() = %q, too short for semver", v)
	}
}
