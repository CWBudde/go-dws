package interp

import (
	"unicode/utf8"
)

// runeLength returns the number of Unicode characters (runes) in a string,
// not the byte length. This is important for UTF-8 strings where characters
// can be multiple bytes.
func runeLength(s string) int {
	return utf8.RuneCountInString(s)
}

// runeAt returns the rune at the given 1-based index in the string.
// Returns the rune and true if the index is valid, or 0 and false otherwise.
// DWScript uses 1-based indexing, so index 1 returns the first character.
func runeAt(s string, index int) (rune, bool) {
	if index < 1 {
		return 0, false
	}

	runes := []rune(s)
	if index > len(runes) {
		return 0, false
	}

	return runes[index-1], true
}

// runeSlice returns a substring based on 1-based character positions (not byte positions).
// start is inclusive, end is exclusive (like Go's slice notation).
// If start < 1, it's treated as 1. If end > length, it's treated as length.
func runeSlice(s string, start, end int) string {
	runes := []rune(s)
	length := len(runes)

	// Adjust start
	if start < 1 {
		start = 1
	}
	startIdx := start - 1 // Convert to 0-based

	// Adjust end
	if end > length {
		end = length
	}
	endIdx := end // Already 0-based for exclusive end

	// Bounds check
	if startIdx >= length || startIdx >= endIdx {
		return ""
	}

	return string(runes[startIdx:endIdx])
}

// runeSliceFrom returns a substring starting from a 1-based position and taking count characters.
// This is commonly used in DWScript's Copy function: Copy(str, start, count)
func runeSliceFrom(s string, start, count int) string {
	if start < 1 || count <= 0 {
		return ""
	}

	runes := []rune(s)
	length := len(runes)

	startIdx := start - 1 // Convert to 0-based
	if startIdx >= length {
		return ""
	}

	endIdx := startIdx + count
	if endIdx > length {
		endIdx = length
	}

	return string(runes[startIdx:endIdx])
}

// runeDelete removes count characters starting from a 1-based position.
// This is used for DWScript's Delete procedure.
func runeDelete(s string, pos, count int) string {
	if pos < 1 || count <= 0 {
		return s
	}

	runes := []rune(s)
	length := len(runes)

	startPos := pos - 1 // Convert to 0-based
	if startPos >= length {
		return s
	}

	endPos := startPos + count
	if endPos > length {
		endPos = length
	}

	// Concatenate the part before deletion and the part after
	return string(runes[:startPos]) + string(runes[endPos:])
}

// runeInsert inserts a substring at a 1-based position.
// This is used for DWScript's Insert procedure.
func runeInsert(source, target string, pos int) string {
	if pos < 1 {
		pos = 1
	}

	runes := []rune(target)
	length := len(runes)

	insertPos := pos - 1 // Convert to 0-based
	if insertPos > length {
		insertPos = length
	}

	return string(runes[:insertPos]) + source + string(runes[insertPos:])
}
