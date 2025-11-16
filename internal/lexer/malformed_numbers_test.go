package lexer

import (
	"testing"
)

// TestMalformedNumberLiterals tests edge cases and potentially malformed number literals.
// This test documents current behavior and will be updated as we add validation.
func TestMalformedNumberLiterals(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedTokens []struct {
			tokenType TokenType
			literal   string
		}
		expectedErrors int // number of expected lexer errors
	}{
		{
			name:  "hex literal with no digits: $",
			input: "$",
			expectedTokens: []struct {
				tokenType TokenType
				literal   string
			}{
				{DOLLAR, "$"},
				{EOF, ""},
			},
			expectedErrors: 0, // Currently no error, but should there be?
		},
		{
			name:  "hex literal with invalid continuation: $FFG",
			input: "$FFG",
			expectedTokens: []struct {
				tokenType TokenType
				literal   string
			}{
				{INT, "$FF"},
				{IDENT, "G"},
				{EOF, ""},
			},
			expectedErrors: 0, // This is actually valid - two tokens
		},
		{
			name:  "binary literal with no digits: %",
			input: "%",
			expectedTokens: []struct {
				tokenType TokenType
				literal   string
			}{
				{PERCENT, "%"},
				{EOF, ""},
			},
			expectedErrors: 0, // Treated as modulo operator
		},
		{
			name:  "binary literal with invalid digit: %2",
			input: "%2",
			expectedTokens: []struct {
				tokenType TokenType
				literal   string
			}{
				{PERCENT, "%"},
				{INT, "2"},
				{EOF, ""},
			},
			expectedErrors: 0, // % is modulo, 2 is separate number
		},
		{
			name:  "binary literal with mixed digits: %012",
			input: "%012",
			expectedTokens: []struct {
				tokenType TokenType
				literal   string
			}{
				{INT, "%01"},   // Binary part
				{INT, "2"},     // Separate decimal number
				{EOF, ""},
			},
			expectedErrors: 0, // Two separate tokens - this is valid behavior
		},
		{
			name:  "0x hex with no digits: 0x",
			input: "0x",
			expectedTokens: []struct {
				tokenType TokenType
				literal   string
			}{
				{INT, "0x"}, // Empty hex literal - invalid!
				{EOF, ""},
			},
			expectedErrors: 1, // Should error: hex literal requires at least one digit
		},
		{
			name:  "0x hex with invalid continuation: 0xFFG",
			input: "0xFFG",
			expectedTokens: []struct {
				tokenType TokenType
				literal   string
			}{
				{INT, "0xFF"},
				{IDENT, "G"},
				{EOF, ""},
			},
			expectedErrors: 0, // Valid - two tokens
		},
		{
			name:  "$ followed by non-hex: $Z",
			input: "$Z",
			expectedTokens: []struct {
				tokenType TokenType
				literal   string
			}{
				{DOLLAR, "$"},
				{IDENT, "Z"},
				{EOF, ""},
			},
			expectedErrors: 0, // $ is a valid token (address-of operator)
		},
		{
			name:  "multiple $ in expression: x $$ y",
			input: "x $$ y",
			expectedTokens: []struct {
				tokenType TokenType
				literal   string
			}{
				{IDENT, "x"},
				{DOLLAR, "$"},
				{DOLLAR, "$"},
				{IDENT, "y"},
				{EOF, ""},
			},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			// Collect all tokens
			for i, expected := range tt.expectedTokens {
				tok := l.NextToken()
				if tok.Type != expected.tokenType {
					t.Errorf("token[%d]: type = %v, want %v", i, tok.Type, expected.tokenType)
				}
				if tok.Literal != expected.literal {
					t.Errorf("token[%d]: literal = %q, want %q", i, tok.Literal, expected.literal)
				}
			}

			// Check error count
			errors := l.Errors()
			if len(errors) != tt.expectedErrors {
				t.Errorf("error count = %d, want %d", len(errors), tt.expectedErrors)
				for i, err := range errors {
					t.Logf("  error[%d]: %s at %s", i, err.Message, err.Pos)
				}
			}
		})
	}
}

// TestValidNumberLiterals ensures valid number literals are not flagged as errors
func TestValidNumberLiterals(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		tokenType TokenType
		literal   string
	}{
		{
			name:      "valid hex $FF",
			input:     "$FF",
			tokenType: INT,
			literal:   "$FF",
		},
		{
			name:      "valid hex 0xFF",
			input:     "0xFF",
			tokenType: INT,
			literal:   "0xFF",
		},
		{
			name:      "valid binary %1010",
			input:     "%1010",
			tokenType: INT,
			literal:   "%1010",
		},
		{
			name:      "valid decimal 42",
			input:     "42",
			tokenType: INT,
			literal:   "42",
		},
		{
			name:      "valid float 3.14",
			input:     "3.14",
			tokenType: FLOAT,
			literal:   "3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.tokenType {
				t.Errorf("type = %v, want %v", tok.Type, tt.tokenType)
			}
			if tok.Literal != tt.literal {
				t.Errorf("literal = %q, want %q", tok.Literal, tt.literal)
			}

			// Should have no errors
			if len(l.Errors()) > 0 {
				t.Errorf("unexpected errors: %v", l.Errors())
			}
		})
	}
}
