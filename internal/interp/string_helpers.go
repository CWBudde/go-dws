package interp

import (
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
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

// runeReplace replaces the rune at the given 1-based index in the string.
// Returns the updated string and true if the index was valid, or the original string and false otherwise.
func runeReplace(s string, index int, replacement rune) (string, bool) {
	if index < 1 {
		return s, false
	}

	runes := []rune(s)
	if index > len(runes) {
		return s, false
	}

	runes[index-1] = replacement
	return string(runes), true
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

// normalizeUnicode normalizes a string to the specified Unicode normalization form.
// Supported forms: NFC, NFD, NFKC, NFKD
func normalizeUnicode(s string, form string) string {
	switch form {
	case "NFC":
		return norm.NFC.String(s)
	case "NFD":
		return norm.NFD.String(s)
	case "NFKC":
		return norm.NFKC.String(s)
	case "NFKD":
		return norm.NFKD.String(s)
	default:
		// Default to NFC if form is unknown
		return norm.NFC.String(s)
	}
}

// stripAccents removes diacritical marks from a string.
// It works by normalizing the string to NFD (decomposed form) and then
// removing all combining marks (which include accents).
func stripAccents(s string) string {
	// Normalize to NFD (decomposed form)
	normalized := norm.NFD.String(s)

	// Filter out combining marks
	var result []rune
	for _, r := range normalized {
		if !unicode.Is(unicode.Mn, r) {
			result = append(result, r)
		}
	}

	return string(result)
}

// runeSetLength resizes a string to the specified character length.
// If the new length is shorter, the string is truncated.
// If the new length is longer, the string is padded with spaces.
// This matches DWScript's SetLength behavior for strings.
func runeSetLength(s string, newLength int) string {
	if newLength < 0 {
		newLength = 0
	}

	runes := []rune(s)
	currentLength := len(runes)

	if newLength == currentLength {
		return s
	}

	if newLength < currentLength {
		// Truncate
		return string(runes[:newLength])
	}

	// Extend with spaces to match DWScript semantics
	padding := newLength - currentLength
	spaces := make([]rune, padding)
	for i := range spaces {
		spaces[i] = ' '
	}
	return s + string(spaces)
}
