// Package ident provides utilities for case-insensitive identifier handling.
//
// DWScript is a case-insensitive language, meaning that identifiers like
// "MyVariable", "myvariable", and "MYVARIABLE" all refer to the same entity.
//
// This package centralizes the normalization and comparison logic to ensure
// consistency across the entire codebase and prevent accidental case-sensitive
// comparisons.
//
// Usage Guidelines:
//
//   - Use Normalize() when storing identifiers as map keys
//   - Use Equal() for one-off case-insensitive string comparisons
//   - Use Compare() when sorting identifiers
//   - Avoid direct strings.ToLower() or strings.EqualFold() - use these helpers instead
//
// Example:
//
//	// Storing in a map
//	variables := make(map[string]Value)
//	variables[ident.Normalize("MyVar")] = someValue
//
//	// Looking up
//	val, ok := variables[ident.Normalize("myvar")] // Found!
//
//	// Comparing
//	if ident.Equal(name1, name2) {
//	    // Names are equal (case-insensitive)
//	}
package ident

import (
	"strings"
)

// Normalize returns the canonical normalized form of an identifier.
// In DWScript, identifiers are case-insensitive, so normalization converts
// to lowercase for consistent comparison and storage.
//
// Use this function when:
//   - Creating map keys for identifier-based lookups
//   - Storing normalized identifiers for later comparison
//   - Implementing identifier-based registries or symbol tables
//
// The original case should be preserved separately for display purposes
// (error messages, code generation, etc.).
//
// Example:
//
//	normalized := ident.Normalize("MyVariable") // "myvariable"
//	store[normalized] = value
func Normalize(s string) string {
	return strings.ToLower(s)
}

// Equal performs a case-insensitive comparison between two strings.
// Returns true if the strings are equal when ignoring case.
//
// Use this function for:
//   - One-off identifier comparisons
//   - Checking if an identifier matches a known value
//   - Validating identifier equality in semantic analysis
//
// This is more efficient than normalizing both strings and comparing,
// as it avoids allocating new strings.
//
// Example:
//
//	if ident.Equal(funcName, "PrintLn") {
//	    // Handle PrintLn function
//	}
func Equal(a, b string) bool {
	return strings.EqualFold(a, b)
}

// Compare performs a case-insensitive lexicographic comparison of two strings.
// Returns:
//   - negative value if a < b
//   - zero if a == b
//   - positive value if a > b
//
// Use this function when:
//   - Sorting identifiers
//   - Implementing ordered collections of identifiers
//   - Creating deterministic output from identifier sets
//
// Example:
//
//	names := []string{"Charlie", "alice", "BOB"}
//	sort.Slice(names, func(i, j int) bool {
//	    return ident.Compare(names[i], names[j]) < 0
//	})
//	// Result: ["alice", "BOB", "Charlie"] (case-insensitive sort)
func Compare(a, b string) int {
	return strings.Compare(Normalize(a), Normalize(b))
}

// Contains checks if a slice of strings contains the given string (case-insensitive).
// Returns true if any element in the slice equals s when ignoring case.
//
// Example:
//
//	keywords := []string{"begin", "end", "if", "then"}
//	if ident.Contains(keywords, "BEGIN") {
//	    // Found!
//	}
func Contains(slice []string, s string) bool {
	for _, item := range slice {
		if Equal(item, s) {
			return true
		}
	}
	return false
}

// Index returns the index of the first occurrence of s in slice (case-insensitive).
// Returns -1 if s is not found in slice.
//
// Example:
//
//	keywords := []string{"begin", "end", "if"}
//	idx := ident.Index(keywords, "END") // Returns 1
func Index(slice []string, s string) int {
	for i, item := range slice {
		if Equal(item, s) {
			return i
		}
	}
	return -1
}

// IsKeyword checks if a string matches any of the provided keywords (case-insensitive).
// This is a convenience function for checking membership in a keyword set.
//
// Example:
//
//	if ident.IsKeyword(name, "begin", "end", "if", "then") {
//	    // name is a keyword
//	}
func IsKeyword(s string, keywords ...string) bool {
	return Contains(keywords, s)
}
