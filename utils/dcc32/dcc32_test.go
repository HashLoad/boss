//nolint:testpackage // Testing internal functions
package dcc32

import (
	"strings"
	"testing"
)

// TestGetDcc32DirByCmd tests the dcc32 directory detection.
func TestGetDcc32DirByCmd(_ *testing.T) {
	// This function calls system command "where dcc32"
	// On non-Windows or without Delphi, it will return empty
	// Just ensure it doesn't panic
	result := GetDcc32DirByCmd()

	// Result depends on system - just verify it's a slice
	_ = result
}

// TestGetDcc32DirByCmd_ProcessOutput tests output processing logic.
func TestGetDcc32DirByCmd_ProcessOutput(t *testing.T) {
	// Test the string processing logic used in GetDcc32DirByCmd
	testCases := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty output",
			input:    "",
			expected: 0,
		},
		{
			name:     "single path",
			input:    "C:\\Program Files\\Embarcadero\\Studio\\22.0\\bin\\dcc32.exe\n",
			expected: 1,
		},
		{
			name:     "multiple paths",
			input:    "C:\\path1\\dcc32.exe\nC:\\path2\\dcc32.exe\n",
			expected: 2,
		},
		{
			name:     "with tabs and carriage returns",
			input:    "C:\\path1\\dcc32.exe\r\n\tC:\\path2\\dcc32.exe\r\n",
			expected: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the processing in GetDcc32DirByCmd
			outputStr := strings.ReplaceAll(tc.input, "\t", "")
			outputStr = strings.ReplaceAll(outputStr, "\r", "")

			if len(strings.ReplaceAll(outputStr, "\n", "")) == 0 {
				if tc.expected != 0 {
					t.Errorf("Expected %d results, got 0", tc.expected)
				}
				return
			}

			count := 0
			for _, value := range strings.Split(outputStr, "\n") {
				if len(strings.TrimSpace(value)) > 0 {
					count++
				}
			}

			if count != tc.expected {
				t.Errorf("Expected %d results, got %d", tc.expected, count)
			}
		})
	}
}
