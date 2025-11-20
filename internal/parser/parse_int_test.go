package parser

import (
	"testing"
)

// TestParseInt tests parseInt with various inputs
func TestParseInt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		shouldError bool
	}{
		{
			name:     "positive integer",
			input:    "42",
			expected: 42,
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "negative integer",
			input:    "-123",
			expected: -123,
		},
		{
			name:     "large number",
			input:    "999999",
			expected: 999999,
		},
		{
			name:        "invalid - non-numeric",
			input:       "abc",
			shouldError: true,
		},
		{
			name:        "invalid - float",
			input:       "42.5",
			shouldError: true,
		},
		{
			name:        "invalid - empty",
			input:       "",
			shouldError: true,
		},
		{
			name:        "invalid - overflow",
			input:       "99999999999999999999999999999",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseInt(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for input %q, got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("parseInt(%q) = %d, want %d", tt.input, result, tt.expected)
				}
			}
		})
	}
}
