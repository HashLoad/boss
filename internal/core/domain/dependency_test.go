package domain_test

import (
	"testing"

	"github.com/hashload/boss/internal/core/domain"
)

func TestDependency_Name(t *testing.T) {
	tests := []struct {
		name       string
		repository string
		expected   string
	}{
		{
			name:       "github repository",
			repository: "github.com/hashload/boss",
			expected:   "boss",
		},
		{
			name:       "gitlab repository",
			repository: "gitlab.com/user/project",
			expected:   "project",
		},
		{
			name:       "bitbucket repository",
			repository: "bitbucket.org/team/repo",
			expected:   "repo",
		},
		{
			name:       "nested path repository",
			repository: "github.com/org/group/subgroup/repo",
			expected:   "repo",
		},
		{
			name:       "repository with trailing slash",
			repository: "github.com/hashload/boss/",
			expected:   "boss/",
		},
		{
			name:       "simple name",
			repository: "simple-repo",
			expected:   "simple-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := domain.Dependency{Repository: tt.repository}
			result := dep.Name()
			if result != tt.expected {
				t.Errorf("Name() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDependency_HashName(t *testing.T) {
	tests := []struct {
		name       string
		repository string
	}{
		{
			name:       "github repository",
			repository: "github.com/hashload/boss",
		},
		{
			name:       "empty repository",
			repository: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := domain.Dependency{Repository: tt.repository}
			hash := dep.HashName()

			// MD5 hash should be 32 hex characters
			if len(hash) != 32 {
				t.Errorf("HashName() length = %d, want 32", len(hash))
			}

			// Same repository should produce same hash
			dep2 := domain.Dependency{Repository: tt.repository}
			hash2 := dep2.HashName()
			if hash != hash2 {
				t.Errorf("Same repository should produce same hash: got %s and %s", hash, hash2)
			}
		})
	}

	t.Run("different repositories produce different hashes", func(t *testing.T) {
		dep1 := domain.Dependency{Repository: "github.com/user/repo1"}
		dep2 := domain.Dependency{Repository: "github.com/user/repo2"}

		hash1 := dep1.HashName()
		hash2 := dep2.HashName()

		if hash1 == hash2 {
			t.Error("Different repositories should produce different hashes")
		}
	})
}

func TestDependency_GetVersion(t *testing.T) {
	tests := []struct {
		name     string
		info     string
		expected string
	}{
		{
			name:     "semantic version",
			info:     "1.0.0",
			expected: "1.0.0",
		},
		{
			name:     "caret version",
			info:     "^1.0.0",
			expected: "^1.0.0",
		},
		{
			name:     "tilde version",
			info:     "~1.0.0",
			expected: "~1.0.0",
		},
		{
			name:     "two part version gets .0 appended",
			info:     "1.0",
			expected: "1.0.0",
		},
		{
			name:     "single part version gets .0.0 appended",
			info:     "1",
			expected: "1.0.0",
		},
		{
			name:     "caret two part version",
			info:     "^1.0",
			expected: "^1.0.0",
		},
		{
			name:     "tilde single part version",
			info:     "~1",
			expected: "~1.0.0",
		},
		{
			name:     "branch name",
			info:     "main",
			expected: "main",
		},
		{
			name:     "version with ssh suffix",
			info:     "1.0.0:ssh",
			expected: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := domain.ParseDependency("github.com/test/repo", tt.info)
			result := dep.GetVersion()
			if result != tt.expected {
				t.Errorf("GetVersion() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseDependency(t *testing.T) {
	tests := []struct {
		name         string
		repo         string
		info         string
		expectedRepo string
		expectedSSH  bool
	}{
		{
			name:         "simple version",
			repo:         "github.com/hashload/boss",
			info:         "1.0.0",
			expectedRepo: "github.com/hashload/boss",
			expectedSSH:  false,
		},
		{
			name:         "version with ssh",
			repo:         "github.com/hashload/boss",
			info:         "1.0.0:ssh",
			expectedRepo: "github.com/hashload/boss",
			expectedSSH:  true,
		},
		{
			name:         "version without ssh explicit",
			repo:         "github.com/hashload/boss",
			info:         "1.0.0:https",
			expectedRepo: "github.com/hashload/boss",
			expectedSSH:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := domain.ParseDependency(tt.repo, tt.info)

			if dep.Repository != tt.expectedRepo {
				t.Errorf("Repository = %q, want %q", dep.Repository, tt.expectedRepo)
			}
			if dep.UseSSH != tt.expectedSSH {
				t.Errorf("UseSSH = %v, want %v", dep.UseSSH, tt.expectedSSH)
			}
		})
	}
}

func TestGetDependencies(t *testing.T) {
	tests := []struct {
		name     string
		deps     map[string]string
		expected int
	}{
		{
			name:     "empty map",
			deps:     map[string]string{},
			expected: 0,
		},
		{
			name: "single dependency",
			deps: map[string]string{
				"github.com/hashload/boss": "1.0.0",
			},
			expected: 1,
		},
		{
			name: "multiple dependencies",
			deps: map[string]string{
				"github.com/hashload/boss":  "1.0.0",
				"github.com/hashload/horse": "^2.0.0",
				"github.com/user/repo":      "~1.5.0",
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.GetDependencies(tt.deps)
			if len(result) != tt.expected {
				t.Errorf("GetDependencies() returned %d dependencies, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestGetDependenciesNames(t *testing.T) {
	deps := []domain.Dependency{
		{Repository: "github.com/hashload/boss"},
		{Repository: "github.com/hashload/horse"},
		{Repository: "github.com/user/repo"},
	}

	names := domain.GetDependenciesNames(deps)

	if len(names) != 3 {
		t.Errorf("GetDependenciesNames() returned %d names, want 3", len(names))
	}

	expectedNames := []string{"boss", "horse", "repo"}
	for i, expected := range expectedNames {
		if names[i] != expected {
			t.Errorf("GetDependenciesNames()[%d] = %q, want %q", i, names[i], expected)
		}
	}
}

func TestDependency_GetURLPrefix(t *testing.T) {
	tests := []struct {
		name       string
		repository string
		expected   string
	}{
		{
			name:       "github.com",
			repository: "github.com/hashload/boss",
			expected:   "github.com",
		},
		{
			name:       "gitlab.com",
			repository: "gitlab.com/user/repo",
			expected:   "gitlab.com",
		},
		{
			name:       "bitbucket.org",
			repository: "bitbucket.org/team/project",
			expected:   "bitbucket.org",
		},
		{
			name:       "custom domain",
			repository: "git.mycompany.com/team/repo",
			expected:   "git.mycompany.com",
		},
		{
			name:       "https url",
			repository: "https://github.com/user/repo",
			expected:   "https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := domain.Dependency{Repository: tt.repository}
			result := dep.GetURLPrefix()
			if result != tt.expected {
				t.Errorf("GetURLPrefix() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDependency_GetURL(t *testing.T) {
	tests := []struct {
		name       string
		repository string
		wantPrefix string
	}{
		{
			name:       "adds https to plain repository",
			repository: "github.com/hashload/boss",
			wantPrefix: "https://github.com/hashload/boss",
		},
		{
			name:       "keeps https url as is",
			repository: "https://github.com/user/repo",
			wantPrefix: "https://github.com/user/repo",
		},
		{
			name:       "keeps http url as is",
			repository: "http://git.local/repo",
			wantPrefix: "http://git.local/repo",
		},
		{
			name:       "gitlab repository",
			repository: "gitlab.com/user/project",
			wantPrefix: "https://gitlab.com/user/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := domain.Dependency{Repository: tt.repository}
			result := dep.GetURL()
			if result != tt.wantPrefix {
				t.Errorf("GetURL() = %q, want %q", result, tt.wantPrefix)
			}
		})
	}
}

func TestDependency_SSHUrl(t *testing.T) {
	tests := []struct {
		name       string
		repository string
		expected   string
	}{
		{
			name:       "github repository converts to ssh format",
			repository: "github.com/hashload/boss",
			expected:   "git@github.com:hashload/boss",
		},
		{
			name:       "gitlab repository converts to ssh format",
			repository: "gitlab.com/user/project",
			expected:   "git@gitlab.com:user/project",
		},
		{
			name:       "already ssh format is returned as-is",
			repository: "git@github.com:hashload/boss",
			expected:   "git@github.com:hashload/boss",
		},
		{
			name:       "custom domain converts to ssh format",
			repository: "git.company.com/team/repo",
			expected:   "git@git.company.com:team/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := domain.Dependency{Repository: tt.repository}
			result := dep.SSHUrl()
			if result != tt.expected {
				t.Errorf("SSHUrl() = %q, want %q", result, tt.expected)
			}
		})
	}
}
