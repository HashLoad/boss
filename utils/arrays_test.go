package utils_test

import (
	"testing"

	"github.com/hashload/boss/utils"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		element  string
		expected bool
	}{
		{
			name:     "element exists in slice",
			slice:    []string{"apple", "banana", "cherry"},
			element:  "banana",
			expected: true,
		},
		{
			name:     "element does not exist in slice",
			slice:    []string{"apple", "banana", "cherry"},
			element:  "grape",
			expected: false,
		},
		{
			name:     "case insensitive match",
			slice:    []string{"Apple", "Banana", "Cherry"},
			element:  "banana",
			expected: true,
		},
		{
			name:     "case insensitive match uppercase search",
			slice:    []string{"apple", "banana", "cherry"},
			element:  "BANANA",
			expected: true,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			element:  "banana",
			expected: false,
		},
		{
			name:     "empty element",
			slice:    []string{"apple", "banana", "cherry"},
			element:  "",
			expected: false,
		},
		{
			name:     "empty element in slice with empty string",
			slice:    []string{"apple", "", "cherry"},
			element:  "",
			expected: true,
		},
		{
			name:     "single element slice - found",
			slice:    []string{"only"},
			element:  "only",
			expected: true,
		},
		{
			name:     "single element slice - not found",
			slice:    []string{"only"},
			element:  "other",
			expected: false,
		},
		{
			name:     "mixed case elements",
			slice:    []string{"GitHub.com", "gitlab.COM", "BitBucket.ORG"},
			element:  "GITHUB.COM",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.Contains(tt.slice, tt.element)
			if result != tt.expected {
				t.Errorf("Contains(%v, %q) = %v, want %v", tt.slice, tt.element, result, tt.expected)
			}
		})
	}
}
