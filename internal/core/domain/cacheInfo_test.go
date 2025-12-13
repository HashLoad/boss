package domain_test

import (
	"testing"

	"github.com/hashload/boss/internal/core/domain"
)

func TestNewRepoInfo(t *testing.T) {
	dep := domain.ParseDependency("github.com/hashload/horse", "^1.0.0")
	versions := []string{"1.0.0", "1.1.0", "1.2.0"}

	info := domain.NewRepoInfo(dep, versions)

	if info.Key != dep.HashName() {
		t.Errorf("NewRepoInfo().Key = %q, want %q", info.Key, dep.HashName())
	}

	if info.Name != "horse" {
		t.Errorf("NewRepoInfo().Name = %q, want %q", info.Name, "horse")
	}

	if len(info.Versions) != 3 {
		t.Errorf("NewRepoInfo().Versions count = %d, want 3", len(info.Versions))
	}

	if info.LastUpdate.IsZero() {
		t.Error("NewRepoInfo().LastUpdate should not be zero")
	}
}

func TestRepoInfo_Struct(t *testing.T) {
	info := domain.RepoInfo{
		Key:      "abc123",
		Name:     "test-repo",
		Versions: []string{"1.0.0", "2.0.0"},
	}

	if info.Key != "abc123" {
		t.Errorf("Key = %q, want %q", info.Key, "abc123")
	}
	if info.Name != "test-repo" {
		t.Errorf("Name = %q, want %q", info.Name, "test-repo")
	}
	if len(info.Versions) != 2 {
		t.Errorf("Versions count = %d, want 2", len(info.Versions))
	}
}
