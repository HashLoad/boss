//nolint:testpackage // Testing internal implementation details
package installer

import (
	"testing"

	"github.com/hashload/boss/internal/core/domain"
)

func TestCollectAllDependencies(t *testing.T) {
	tests := []struct {
		name     string
		pkg      *domain.Package
		expected int
	}{
		{
			name: "empty dependencies",
			pkg: &domain.Package{
				Dependencies: nil,
			},
			expected: 0,
		},
		{
			name: "single dependency",
			pkg: &domain.Package{
				Dependencies: map[string]string{
					"dep1": "github.com/example/dep1",
				},
			},
			expected: 1,
		},
		{
			name: "multiple dependencies",
			pkg: &domain.Package{
				Dependencies: map[string]string{
					"dep1": "github.com/example/dep1",
					"dep2": "github.com/example/dep2",
					"dep3": "github.com/example/dep3",
				},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectAllDependencies(tt.pkg)
			if len(result) != tt.expected {
				t.Errorf("Expected %d dependencies, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestAddWarning(t *testing.T) {
	ctx := &installContext{
		warnings: make([]string, 0),
	}

	initialLen := len(ctx.warnings)
	ctx.addWarning("Test warning")

	if len(ctx.warnings) != initialLen+1 {
		t.Errorf("Expected %d warnings, got %d", initialLen+1, len(ctx.warnings))
	}

	if ctx.warnings[0] != "Test warning" {
		t.Errorf("Expected warning 'Test warning', got %q", ctx.warnings[0])
	}
}
