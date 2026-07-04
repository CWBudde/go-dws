package interp

import (
	"github.com/cwbudde/go-dws/internal/encoding"
)

// detectAndDecodeFile reads a file and detects its encoding based on BOM (Byte Order Mark).
// It supports UTF-8, UTF-16 LE, and UTF-16 BE. Files without BOM are assumed to be UTF-8.
// The function returns the file content as a UTF-8 string.
func detectAndDecodeFile(path string) (string, error) {
	return encoding.DecodeFile(path)
}
