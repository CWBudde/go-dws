// Package testutil provides shared test helper functions.
package testutil

import "unicode"

// CapitalizeFirst capitalizes the first letter of a string.
// Used for testing case-insensitive keyword handling.
func CapitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// AlternatingCase converts a string to alternating case (e.g., "hello" -> "hElLo").
// Used for testing case-insensitive keyword handling.
func AlternatingCase(s string) string {
	runes := []rune(s)
	for i := range runes {
		if i%2 == 0 {
			runes[i] = unicode.ToLower(runes[i])
		} else {
			runes[i] = unicode.ToUpper(runes[i])
		}
	}
	return string(runes)
}
