package lexer

import (
	"strings"
	"testing"
)

// TestInvalidUTF8Detection tests that the lexer properly detects and reports invalid UTF-8 sequences
func TestInvalidUTF8Detection(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
		errorContains  string
	}{
		{
			name:           "invalid UTF-8 byte sequence",
			input:          "var x\xff:= 5", // \xff is invalid UTF-8
			expectedErrors: 1,
			errorContains:  "invalid UTF-8",
		},
		{
			name:           "invalid UTF-8 in identifier",
			input:          "myVar\xfe\xffTest", // \xfe\xff is invalid
			expectedErrors: 2,                   // Two invalid bytes
			errorContains:  "invalid UTF-8",
		},
		{
			name:           "multiple invalid UTF-8 sequences",
			input:          "\xff\xfe\xfd", // Three invalid bytes
			expectedErrors: 3,
			errorContains:  "invalid UTF-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			// Consume all tokens
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
			}

			errors := l.Errors()
			if len(errors) != tt.expectedErrors {
				t.Errorf("error count = %d, want %d", len(errors), tt.expectedErrors)
				for i, err := range errors {
					t.Logf("  error[%d]: %s at %s", i, err.Message, err.Pos)
				}
			}

			// Check that errors contain expected text
			if tt.expectedErrors > 0 && len(errors) > 0 {
				found := false
				for _, err := range errors {
					if strings.Contains(err.Message, tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("no error contains %q, errors: %v", tt.errorContains, errors)
				}
			}
		})
	}
}

// TestValidUTF8 ensures valid UTF-8 doesn't produce errors
func TestValidUTF8(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "ASCII only",
			input: "var x := 42;",
		},
		{
			name:  "Greek letters",
			input: "var Δ := 3.14;",
		},
		{
			name:  "emoji in comment",
			input: "// 🚀 rocket",
		},
		{
			name:  "emoji in string",
			input: "'🚀🌟💻'",
		},
		{
			name:  "Chinese identifier",
			input: "var 变量 := 100;",
		},
		{
			name:  "mixed multi-byte identifiers",
			input: "函数 Δ test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			// Consume all tokens
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
			}

			// Should have no errors
			errors := l.Errors()
			if len(errors) > 0 {
				t.Errorf("unexpected errors: %v", errors)
			}
		})
	}
}
