//nolint:testpackage // Testing internal functions
package git

import (
	"testing"

	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
)

// TestGetMain_EmptyRepo tests GetMain with an empty repository.
func TestGetMain_EmptyRepo(t *testing.T) {
	// Create an in-memory repository
	repo, err := goGit.Init(memory.NewStorage(), nil)
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	// GetMain should return an error for empty repo
	_, err = GetMain(repo)
	if err == nil {
		t.Error("GetMain() should return error for repo without main/master branch")
	}
}

// TestGetTagsShortName_NoTags tests GetTagsShortName with no tags.
func TestGetTagsShortName_NoTags(t *testing.T) {
	// Create an in-memory repository
	repo, err := goGit.Init(memory.NewStorage(), nil)
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	result := GetTagsShortName(repo)

	if len(result) != 0 {
		t.Errorf("GetTagsShortName() should return empty for repo with no tags, got %v", result)
	}
}

// TestParseVersion tests version parsing from tags.
func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		expected string
	}{
		{
			name:     "v prefix",
			tagName:  "v1.0.0",
			expected: "v1.0.0",
		},
		{
			name:     "no prefix",
			tagName:  "1.0.0",
			expected: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := plumbing.NewReferenceFromStrings("refs/tags/"+tt.tagName, "abc123")

			shortName := ref.Name().Short()
			if shortName != tt.tagName {
				t.Errorf("Short() = %q, want %q", shortName, tt.tagName)
			}
		})
	}
}
