package utils

import "strings"

func Contains(a []string, x string) bool {
	for _, n := range a {
		if strings.ToUpper(x) == strings.ToUpper(n) {
			return true
		}
	}
	return false
}
