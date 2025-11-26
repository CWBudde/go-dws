package interp

import (
	"bytes"
	"fmt"
	"os"
	"unicode/utf8"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// detectAndDecodeFile reads a file and detects its encoding based on BOM (Byte Order Mark).
// It supports UTF-8, UTF-16 LE, and UTF-16 BE. Files without BOM are assumed to be UTF-8.
// The function returns the file content as a UTF-8 string.
func detectAndDecodeFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Check for BOM and decode accordingly
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		// UTF-8 BOM: EF BB BF
		// Strip BOM and return as string
		return string(data[3:]), nil
	}

	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		// UTF-16 LE BOM: FF FE
		return decodeUTF16(data, unicode.LittleEndian)
	}

	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		// UTF-16 BE BOM: FE FF
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

// decodeUTF16 decodes UTF-16 encoded data to UTF-8 string.
func decodeUTF16(data []byte, endianness unicode.Endianness) (string, error) {
	// Create UTF-16 decoder
	decoder := unicode.UTF16(endianness, unicode.UseBOM).NewDecoder()

	// Transform UTF-16 bytes to UTF-8
	utf8Data, _, err := transform.Bytes(decoder, data)
	if err != nil {
		return "", fmt.Errorf("failed to decode UTF-16: %w", err)
	}

	// Remove BOM from the result if present (the decoder might include it)
	// UTF-8 BOM in the decoded string
	if len(utf8Data) >= 3 && utf8Data[0] == 0xEF && utf8Data[1] == 0xBB && utf8Data[2] == 0xBF {
		utf8Data = utf8Data[3:]
	}

	// Also check for UTF-16 BOM characters that might have been decoded
	// (U+FEFF ZERO WIDTH NO-BREAK SPACE)
	result := string(utf8Data)
	result = string(bytes.TrimPrefix([]byte(result), []byte("\uFEFF")))

	return result, nil
}
