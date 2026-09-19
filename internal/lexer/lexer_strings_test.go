package lexer

import (
	"testing"
)

func TestStringLiterals(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedLiteral string
		expectedType    TokenType
	}{
		{
			name:            "simple single quoted",
			input:           `'hello'`,
			expectedType:    STRING,
			expectedLiteral: "hello",
		},
		{
			name:            "simple double quoted",
			input:           `"world"`,
			expectedType:    STRING,
			expectedLiteral: "world",
		},
		{
			name:            "escaped single quote",
			input:           `'it''s'`,
			expectedType:    STRING,
			expectedLiteral: "it's",
		},
		{
			name:            "empty string",
			input:           `''`,
			expectedType:    STRING,
			expectedLiteral: "",
		},
		{
			name:            "string with spaces",
			input:           `'hello world'`,
			expectedType:    STRING,
			expectedLiteral: "hello world",
		},
		{
			// Only double-quoted strings may span lines.
			name:            "multiline string",
			input:           "\"hello\nworld\"",
			expectedType:    STRING,
			expectedLiteral: "hello\nworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.expectedType {
				t.Fatalf("tokentype wrong. expected=%q, got=%q",
					tt.expectedType, tok.Type)
			}

			if tok.Literal != tt.expectedLiteral {
				t.Fatalf("literal wrong. expected=%q, got=%q",
					tt.expectedLiteral, tok.Literal)
			}
		})
	}
}

func TestCharLiterals(t *testing.T) {
	input := `#65 #$41 #13 #10`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"#65", CHAR},  // #65 = ASCII 'A'
		{"#$41", CHAR}, // #$41 = ASCII 'A' (hex)
		{"#13", CHAR},  // #13 = CR
		{"#10", CHAR},  // #10 = LF
		{"", EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestCharLiteralStandaloneStillWorks tests that isCharLiteralStandalone works after refactoring
func TestCharLiteralStandaloneStillWorks(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		isStandalone bool
	}{
		{
			name:         "standalone character literal",
			input:        "#65",
			isStandalone: true,
		},
		{
			name:         "character literal followed by space and string",
			input:        "#65 'hello'",
			isStandalone: true, // space separates them
		},
		{
			name:         "character literal immediately followed by string",
			input:        "#65'hello'",
			isStandalone: false, // no space, part of concatenation
		},
		{
			name:         "character literal followed by another char literal",
			input:        "#65#66",
			isStandalone: false, // concatenation
		},
		{
			name:         "hex character literal standalone",
			input:        "#$41",
			isStandalone: true,
		},
		{
			name:         "hex character literal in concatenation",
			input:        "#$41#$42",
			isStandalone: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			result := l.isCharLiteralStandalone()
			if result != tt.isStandalone {
				t.Errorf("isCharLiteralStandalone() = %v, expected %v", result, tt.isStandalone)
			}

			// Verify state wasn't changed
			if l.position != 0 {
				t.Errorf("isCharLiteralStandalone() changed position to %d, expected 0", l.position)
			}
			if l.ch != '#' {
				t.Errorf("isCharLiteralStandalone() changed ch to %c, expected '#'", l.ch)
			}
		})
	}
}

// TestInvalidCharConstant covers DWScript's fatal "Invalid char constant" error for
// character literals beyond U+10FFFF, anchored just past the literal.
func TestInvalidCharConstant(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		message string
		column  int
	}{
		{name: "standalone hex", input: "#$200000;", message: `Invalid char constant "$200000"`, column: 9},
		{name: "standalone decimal", input: "#1114112;", message: `Invalid char constant "1114112"`, column: 9},
		{name: "overflowing hex", input: "#$FFFFFFFFFFFFFFFFFF", message: `Invalid char constant "$FFFFFFFFFFFFFFFFFF"`, column: 21},
		{name: "in a string sequence", input: "'a'#$110000'b'", message: `Invalid char constant "$110000"`, column: 12},
		{name: "largest code point is valid", input: "#$10FFFF;"},
		{name: "supplementary plane is valid", input: "#$10000;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			drainTokens(l)
			diags := l.DirectiveDiagnostics()
			if tt.message == "" {
				if len(diags) != 0 {
					t.Fatalf("unexpected diagnostics: %v", diags)
				}
				return
			}
			if len(diags) != 1 || diags[0].Message != tt.message || diags[0].Pos.Line != 1 || diags[0].Pos.Column != tt.column {
				t.Fatalf("got %v, want %q at 1:%d", diags, tt.message, tt.column)
			}
		})
	}
}
