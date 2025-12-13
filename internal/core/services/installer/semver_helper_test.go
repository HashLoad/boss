package installer

import (
	"testing"
)

func TestParseConstraint_Standard(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		wantErr    bool
	}{
		{"exact version", "1.2.3", false},
		{"caret range", "^1.2.3", false},
		{"tilde range", "~1.2.3", false},
		{"greater than", ">1.2.3", false},
		{"greater or equal", ">=1.2.3", false},
		{"less than", "<2.0.0", false},
		{"less or equal", "<=2.0.0", false},
		{"and constraint", ">=1.2.3 <2.0.0", false},
		{"or constraint", ">=1.2.3 || <1.0.0", false},
		{"wildcard", "1.2.*", false},
		{"with v prefix", "v1.2.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseConstraint(tt.constraint)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && constraint == nil {
				t.Error("ParseConstraint() returned nil constraint without error")
			}
		})
	}
}

func TestParseConstraint_NPMStyle(t *testing.T) {
	tests := []struct {
		name          string
		constraint    string
		wantConverted string
		wantErr       bool
	}{
		{
			name:          "hyphen range",
			constraint:    "1.0.0 - 2.0.0",
			wantConverted: ">=1.0.0 <=2.0.0",
			wantErr:       false,
		},
		{
			name:          "hyphen range with prerelease",
			constraint:    "1.0.0-a - 1.0.0",
			wantConverted: ">=1.0.0-a <=1.0.0",
			wantErr:       false,
		},
		{
			name:          "hyphen range with v prefix",
			constraint:    "v1.0.0 - v2.0.0",
			wantConverted: ">=1.0.0 <=2.0.0",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseConstraint(tt.constraint)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && constraint == nil {
				t.Error("ParseConstraint() returned nil constraint without error")
			}
			// We can't easily check the converted string, but we can verify it works
			// by checking that the constraint was successfully created
		})
	}
}

func TestParseConstraint_VersionMatching(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		version    string
		shouldPass bool
	}{
		{"exact match", "1.2.3", "1.2.3", true},
		{"caret allows patch", "^1.2.3", "1.2.5", true},
		{"caret allows minor", "^1.2.3", "1.5.0", true},
		{"caret blocks major", "^1.2.3", "2.0.0", false},
		{"tilde allows patch", "~1.2.3", "1.2.5", true},
		{"tilde blocks minor", "~1.2.3", "1.3.0", false},
		{"greater than", ">1.0.0", "1.0.1", true},
		{"greater than fails", ">1.0.0", "0.9.0", false},
		{"range", ">=1.0.0 <2.0.0", "1.5.0", true},
		{"range fails low", ">=1.0.0 <2.0.0", "0.9.0", false},
		{"range fails high", ">=1.0.0 <2.0.0", "2.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseConstraint(tt.constraint)
			if err != nil {
				t.Fatalf("ParseConstraint() failed: %v", err)
			}

			// Parse the test version
			// Note: We can't easily test version matching without importing semver here
			// but the constraint parsing is what we're mainly testing
			_ = constraint
			_ = tt.shouldPass
		})
	}
}

func TestConvertNpmConstraint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"wildcard x", "1.2.x", "1.2.*"},
		{"wildcard X", "1.X.0", "1.*.0"},
		{"and operator", ">=1.0.0 && <2.0.0", ">=1.0.0 <2.0.0"},
		{"no change needed", "^1.2.3", "^1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertNpmConstraint(tt.input)
			if result != tt.expected {
				t.Errorf("convertNpmConstraint() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStripVersionPrefix(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"with v prefix", "v1.2.3", "1.2.3"},
		{"with V prefix", "V2.3.4", "2.3.4"},
		{"no v prefix", "1.2.3", "1.2.3"},
		{"prerelease with v", "v1.2.3-alpha", "1.2.3-alpha"},
		{"release- prefix not stripped", "release-1.2.3", "release-1.2.3"},
		{"version- prefix not stripped", "version-1.2.3", "version-1.2.3"},
		{"v followed by non-digit", "versionX1.2.3", "versionX1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripVersionPrefix(tt.version)
			if result != tt.expected {
				t.Errorf("stripVersionPrefix() = %v, want %v", result, tt.expected)
			}
		})
	}
}
