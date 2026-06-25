package installer_test

import (
	"testing"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/installer"
)

func TestParseDependency(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name adds hashload prefix",
			input:    "horse",
			expected: "github.com/hashload/horse",
		},
		{
			name:     "owner/repo adds github.com prefix",
			input:    "hashload/boss",
			expected: "github.com/hashload/boss",
		},
		{
			name:     "full path unchanged",
			input:    "github.com/hashload/horse",
			expected: "github.com/hashload/horse",
		},
		{
			name:     "gitlab path unchanged",
			input:    "gitlab.com/user/repo",
			expected: "gitlab.com/user/repo",
		},
		{
			name:     "with version suffix",
			input:    "github.com/hashload/horse@1.0.0",
			expected: "github.com/hashload/horse@1.0.0",
		},
		{
			name:     "bitbucket path unchanged",
			input:    "bitbucket.org/user/repo",
			expected: "bitbucket.org/user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := installer.ParseDependency(tt.input)
			if result != tt.expected {
				t.Errorf("ParseDependency(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEnsureDependency(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedDeps map[string]string
	}{
		{
			name: "simple dependency",
			args: []string{"horse"},
			expectedDeps: map[string]string{
				"github.com/hashload/horse": ">0.0.0",
			},
		},
		{
			name: "dependency with version",
			args: []string{"github.com/hashload/horse@2.0.0"},
			expectedDeps: map[string]string{
				"github.com/hashload/horse": "2.0.0",
			},
		},
		{
			name: "dependency with caret version",
			args: []string{"github.com/hashload/horse@^1.5.0"},
			expectedDeps: map[string]string{
				"github.com/hashload/horse": "^1.5.0",
			},
		},
		{
			name: "multiple dependencies",
			args: []string{"horse", "boss-ide"},
			expectedDeps: map[string]string{
				"github.com/hashload/horse":    ">0.0.0",
				"github.com/hashload/boss-ide": ">0.0.0",
			},
		},
		{
			name: "dependency with .git suffix",
			args: []string{"github.com/hashload/horse.git"},
			expectedDeps: map[string]string{
				"github.com/hashload/horse": ">0.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &domain.Package{
				Dependencies: make(map[string]string),
			}

			installer.EnsureDependency(pkg, tt.args)

			if len(pkg.Dependencies) != len(tt.expectedDeps) {
				t.Errorf("Dependencies count = %d, want %d", len(pkg.Dependencies), len(tt.expectedDeps))
			}

			for dep, ver := range tt.expectedDeps {
				if pkg.Dependencies[dep] != ver {
					t.Errorf("Dependencies[%q] = %q, want %q", dep, pkg.Dependencies[dep], ver)
				}
			}
		})
	}
}

func TestEnsureDependency_OwnerRepo(t *testing.T) {
	pkg := &domain.Package{
		Dependencies: make(map[string]string),
	}

	installer.EnsureDependency(pkg, []string{"hashload/boss"})

	expected := "github.com/hashload/boss"
	if _, ok := pkg.Dependencies[expected]; !ok {
		t.Errorf("Should add dependency for %q", expected)
	}
}

func TestEnsureDependency_TildeVersion(t *testing.T) {
	pkg := &domain.Package{
		Dependencies: make(map[string]string),
	}

	installer.EnsureDependency(pkg, []string{"github.com/hashload/horse@~1.0.0"})

	if ver := pkg.Dependencies["github.com/hashload/horse"]; ver != "~1.0.0" {
		t.Errorf("Version = %q, want ~1.0.0", ver)
	}
}

func TestEnsureDependency_HTTPSUrl(t *testing.T) {
	pkg := &domain.Package{
		Dependencies: make(map[string]string),
	}

	installer.EnsureDependency(pkg, []string{"https://github.com/hashload/horse"})

	// Should strip https:// and add to dependencies
	if len(pkg.Dependencies) == 0 {
		t.Error("Should add dependency for HTTPS URL")
	}
}

func TestEnsureDependency_SSHUrl(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedDep  string
		expectedVer  string
	}{
		{
			name:        "SSH URL with version tag",
			input:       "git@github.com:hashload/boss.git:v1.0.0",
			expectedDep: "git@github.com:hashload/boss",
			expectedVer: "v1.0.0",
		},
		{
			name:        "SSH URL without version tag",
			input:       "git@github.com:hashload/boss.git",
			expectedDep: "git@github.com:hashload/boss",
			expectedVer: ">0.0.0",
		},
		{
			name:        "Self-hosted gitlab SSH URL without version",
			input:       "git@mygitlab.domain.de:delphi/libraries/mylib.git",
			expectedDep: "git@mygitlab.domain.de:delphi/libraries/mylib",
			expectedVer: ">0.0.0",
		},
		{
			name:        "Self-hosted gitlab SSH URL with version",
			input:       "git@mygitlab.domain.de:delphi/libraries/mylib.git:1.2.3",
			expectedDep: "git@mygitlab.domain.de:delphi/libraries/mylib",
			expectedVer: "1.2.3",
		},
		{
			name:        "SSH URL with @ version tag",
			input:       "git@github.com:hashload/boss.git@v2.5.0",
			expectedDep: "git@github.com:hashload/boss",
			expectedVer: "v2.5.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &domain.Package{
				Dependencies: make(map[string]string),
			}

			installer.EnsureDependency(pkg, []string{tt.input})

			if len(pkg.Dependencies) != 1 {
				t.Fatalf("EnsureDependency() did not add exactly 1 dependency for %s, got %d", tt.input, len(pkg.Dependencies))
			}

			actualVer, exists := pkg.Dependencies[tt.expectedDep]
			if !exists {
				// Print keys for debugging
				var keys []string
				for k := range pkg.Dependencies {
					keys = append(keys, k)
				}
				t.Fatalf("Expected dependency %q not found, got keys: %v", tt.expectedDep, keys)
			}

			if actualVer != tt.expectedVer {
				t.Errorf("pkg.Dependencies[%q] = %q, want %q", tt.expectedDep, actualVer, tt.expectedVer)
			}
		})
	}
}

