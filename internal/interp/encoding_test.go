package interp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectAndDecodeFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "encoding_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "UTF-8 without BOM",
			data:     []byte("Hello, World!"),
			expected: "Hello, World!",
		},
		{
			name:     "UTF-8 with BOM",
			data:     []byte{0xEF, 0xBB, 0xBF, 'H', 'e', 'l', 'l', 'o'},
			expected: "Hello",
		},
		{
			name: "UTF-16 LE with BOM - simple ASCII",
			data: []byte{
				0xFF, 0xFE, // BOM
				'H', 0x00, 'i', 0x00, // "Hi" in UTF-16 LE
			},
			expected: "Hi",
		},
		{
			name: "UTF-16 LE with BOM - DWScript code",
			data: []byte{
				0xFF, 0xFE, // BOM
				'{', 0x00, '$', 0x00, 'i', 0x00, 'f', 0x00, 'd', 0x00, 'e', 0x00, 'f', 0x00, '}', 0x00,
			},
			expected: "{$ifdef}",
		},
		{
			name: "UTF-16 BE with BOM - simple ASCII",
			data: []byte{
				0xFE, 0xFF, // BOM
				0x00, 'H', 0x00, 'i', // "Hi" in UTF-16 BE
			},
			expected: "Hi",
		},
		{
			name: "UTF-16 BE with BOM - DWScript code",
			data: []byte{
				0xFE, 0xFF, // BOM
				0x00, '{', 0x00, '$', 0x00, 'i', 0x00, 'f', 0x00, '}',
			},
			expected: "{$if}",
		},
		{
			name:     "Empty file",
			data:     []byte{},
			expected: "",
		},
		{
			name:     "UTF-8 with special characters",
			data:     []byte("Ĥéļļö, Wőřłđ! 你好世界"),
			expected: "Ĥéļļö, Wőřłđ! 你好世界",
		},
		{
			name: "UTF-16 LE with Unicode characters",
			data: []byte{
				0xFF, 0xFE, // BOM
				0x24, 0x01, 0xE9, 0x00, // Ĥé in UTF-16 LE (U+0124 U+00E9)
			},
			expected: "Ĥé",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tmpDir, "test_"+tt.name+".txt")
			if err := os.WriteFile(testFile, tt.data, 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Test the function
			result, err := detectAndDecodeFile(testFile)
			if err != nil {
				t.Fatalf("detectAndDecodeFile failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDetectAndDecodeFile_NonExistentFile(t *testing.T) {
	_, err := detectAndDecodeFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestDetectAndDecodeFile_RealFixture(t *testing.T) {
	// Test with an actual UTF-16 LE file from the fixtures
	fixtureFile := "../../testdata/fixtures/FunctionsString/bytesizetostring.pas"

	// Check if the file exists
	if _, err := os.Stat(fixtureFile); os.IsNotExist(err) {
		t.Skip("Fixture file not found, skipping test")
	}

	result, err := detectAndDecodeFile(fixtureFile)
	if err != nil {
		t.Fatalf("Failed to decode fixture file: %v", err)
	}

	// Check that the result is not empty and starts with expected content
	if len(result) == 0 {
		t.Error("Decoded content is empty")
	}

	// The file should start with {$ifdef JS_CODEGEN}
	if len(result) < 20 {
		t.Errorf("Decoded content is too short: %d bytes", len(result))
	}

	// Check for expected content (should contain DWScript keywords)
	expectedStrings := []string{"{$ifdef", "for", "PrintLn"}
	for _, expected := range expectedStrings {
		if !contains(result, expected) {
			t.Errorf("Expected decoded content to contain %q, but it doesn't. Got: %q", expected, result[:min(100, len(result))])
		}
	}
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
