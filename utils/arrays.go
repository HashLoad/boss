// Package utils provides general utility functions used throughout Boss.
// It includes array manipulation, string operations, and helper functions.
package utils

import "strings"

// Contains checks if a string slice contains a specific string (case-insensitive).
func Contains(a []string, x string) bool {
	for _, n := range a {
		if strings.EqualFold(x, n) {
			return true
		}
	}
	return false
}
