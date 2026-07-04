// Package encoding provides BOM-aware text decoding shared by the CLI,
// the interpreter, and tooling (e.g. cmd/fixture-report).
//
// DWScript's original test suite contains files in UTF-8 (with and without
// BOM), UTF-16 LE/BE (with BOM), and legacy ANSI/Latin-1. DecodeBytes detects
// the encoding from the BOM and returns UTF-8 text, falling back to Latin-1
// for non-UTF-8 byte content so legacy fixtures still round-trip.
package encoding

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// DecodeFile reads a file and decodes it to UTF-8 based on its BOM.
// See DecodeBytes for the detection rules.
func DecodeFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return DecodeBytes(data)
}

// DecodeBytes decodes raw file content to UTF-8 text.
// It supports UTF-8 (BOM stripped), UTF-16 LE, and UTF-16 BE (BOM required).
// Content without a BOM is assumed to be UTF-8; invalid UTF-8 falls back to
// Latin-1 (each byte promoted to the corresponding rune).
func DecodeBytes(data []byte) (string, error) {
	// UTF-8 BOM: EF BB BF
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return string(data[3:]), nil
	}

	// UTF-16 LE BOM: FF FE
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return decodeUTF16(data, unicode.LittleEndian)
	}

	// UTF-16 BE BOM: FE FF
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		return decodeUTF16(data, unicode.BigEndian)
	}

	// No BOM detected, assume UTF-8
	if utf8.Valid(data) {
		return string(data), nil
	}

	// Fallback: treat as Latin-1/bytes and promote to runes
	runes := make([]rune, len(data))
	for i, b := range data {
		runes[i] = rune(b)
	}
	return string(runes), nil
}

// decodeUTF16 decodes UTF-16 encoded data to a UTF-8 string.
func decodeUTF16(data []byte, endianness unicode.Endianness) (string, error) {
	decoder := unicode.UTF16(endianness, unicode.UseBOM).NewDecoder()

	utf8Data, _, err := transform.Bytes(decoder, data)
	if err != nil {
		return "", fmt.Errorf("failed to decode UTF-16: %w", err)
	}

	// Strip any BOM the decoder may have preserved.
	result := string(utf8Data)
	result = strings.TrimPrefix(result, "\uFEFF")
	return result, nil
}
