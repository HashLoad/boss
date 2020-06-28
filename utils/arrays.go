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

func Filter(a []string, condition func(string) bool) []string {
	var result []string
	for _, n := range a {
		if condition(n) {
			result = append(result, n)
		}
	}
	return result
}
