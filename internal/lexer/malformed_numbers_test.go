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
			literal   string
			tokenType TokenType
		}
		expectedErrors int // number of expected lexer errors
	}{
		{
			name:  "hex literal with no digits: $",
			input: "$",
			expectedTokens: []struct {
				literal   string
				tokenType TokenType
			}{
				{"$", DOLLAR},
				{"", EOF},
			},
			expectedErrors: 0, // Currently no error, but should there be?
		},
		{
			name:  "hex literal with invalid continuation: $FFG",
			input: "$FFG",
			expectedTokens: []struct {
				literal   string
				tokenType TokenType
			}{
				{"$FF", INT},
				{"G", IDENT},
				{"", EOF},
			},
			expectedErrors: 0, // This is actually valid - two tokens
		},
		{
			name:  "binary literal with no digits: %",
			input: "%",
			expectedTokens: []struct {
				literal   string
				tokenType TokenType
			}{
				{"%", PERCENT},
				{"", EOF},
			},
			expectedErrors: 0, // Treated as modulo operator
		},
		{
			name:  "binary literal with invalid digit: %2",
			input: "%2",
			expectedTokens: []struct {
				literal   string
				tokenType TokenType
			}{
				{"%", PERCENT},
				{"2", INT},
				{"", EOF},
			},
			expectedErrors: 0, // % is modulo, 2 is separate number
		},
		{
			name:  "binary literal with mixed digits: %012",
			input: "%012",
			expectedTokens: []struct {
				literal   string
				tokenType TokenType
			}{
				{"%01", INT}, // Binary part
				{"2", INT},   // Separate decimal number
				{"", EOF},
			},
			expectedErrors: 0, // Two separate tokens - this is valid behavior
		},
		{
			name:  "0x hex with no digits: 0x",
			input: "0x",
			expectedTokens: []struct {
				literal   string
				tokenType TokenType
			}{
				{"0x", INT}, // Empty hex literal - invalid!
				{"", EOF},
			},
			expectedErrors: 1, // Should error: hex literal requires at least one digit
		},
		{
			name:  "0x hex with invalid continuation: 0xFFG",
			input: "0xFFG",
			expectedTokens: []struct {
				literal   string
				tokenType TokenType
			}{
				{"0xFF", INT},
				{"G", IDENT},
				{"", EOF},
			},
			expectedErrors: 0, // Valid - two tokens
		},
		{
			name:  "$ followed by non-hex: $Z",
			input: "$Z",
			expectedTokens: []struct {
				literal   string
				tokenType TokenType
			}{
				{"$", DOLLAR},
				{"Z", IDENT},
				{"", EOF},
			},
			expectedErrors: 0, // $ is a valid token (address-of operator)
		},
		{
			name:  "multiple $ in expression: x $$ y",
			input: "x $$ y",
			expectedTokens: []struct {
				literal   string
				tokenType TokenType
			}{
				{"x", IDENT},
				{"$", DOLLAR},
				{"$", DOLLAR},
				{"y", IDENT},
				{"", EOF},
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
		literal   string
		tokenType TokenType
	}{
		{
			name:      "valid hex $FF",
			input:     "$FF",
			literal:   "$FF",
			tokenType: INT,
		},
		{
			name:      "valid hex 0xFF",
			input:     "0xFF",
			literal:   "0xFF",
			tokenType: INT,
		},
		{
			name:      "valid binary %1010",
			input:     "%1010",
			literal:   "%1010",
			tokenType: INT,
		},
		{
			name:      "valid decimal 42",
			input:     "42",
			literal:   "42",
			tokenType: INT,
		},
		{
			name:      "valid float 3.14",
			input:     "3.14",
			literal:   "3.14",
			tokenType: FLOAT,
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
