package lexer

import (
	"testing"
)

// TestRosettaStringLiterals tests the string literal examples from examples/rosetta/Literals_String.dws
// This covers:
// 1. Single-quoted strings with embedded double quotes
// 2. Double-quoted strings with escaped double quotes (doubling)
// 3. String concatenation with character literals (#13#10 for CR+LF)
func TestRosettaStringLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			tokenType TokenType
			literal   string
		}
	}{
		{
			name:  "single-quoted string with embedded double quotes",
			input: `'quoted "word" in string'`,
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{STRING, `quoted "word" in string`},
				{EOF, ""},
			},
		},
		{
			name:  "double-quoted string with escaped double quotes",
			input: `"quoted ""word"" in string"`,
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{STRING, `quoted "word" in string`},
				{EOF, ""},
			},
		},
		{
			name:  "string concatenation with char literals (CR+LF)",
			input: `'first line'#13#10'second line'`,
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{STRING, "first line\r\nsecond line"},
				{EOF, ""},
			},
		},
		{
			name:  "complete first example from rosetta",
			input: `const s1 := 'quoted "word" in string';`,
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{CONST, "const"},
				{IDENT, "s1"},
				{ASSIGN, ":="},
				{STRING, `quoted "word" in string`},
				{SEMICOLON, ";"},
				{EOF, ""},
			},
		},
		{
			name:  "complete second example from rosetta with comment",
			input: `const s2 := "quoted ""word"" in string"; // sames as s1, shows the doubling of the delimiter`,
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{CONST, "const"},
				{IDENT, "s2"},
				{ASSIGN, ":="},
				{STRING, `quoted "word" in string`},
				{SEMICOLON, ";"},
				{EOF, ""},
			},
		},
		{
			name:  "complete third example with string+char concatenation",
			input: `const s2 := 'first line'#13#10'second line';`,
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{CONST, "const"},
				{IDENT, "s2"},
				{ASSIGN, ":="},
				{STRING, "first line\r\nsecond line"},
				{SEMICOLON, ";"},
				{EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			for i, exp := range tt.expected {
				tok := l.NextToken()

				if tok.Type != exp.tokenType {
					t.Errorf("token[%d] - wrong type. expected=%q, got=%q (literal=%q)",
						i, exp.tokenType, tok.Type, tok.Literal)
				}

				if tok.Literal != exp.literal {
					t.Errorf("token[%d] - wrong literal. expected=%q, got=%q (type=%q)",
						i, exp.literal, tok.Literal, tok.Type)
				}
			}
		})
	}
}

// TestStringLiteralEscaping specifically tests the escape handling in string literals
func TestStringLiteralEscaping(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedLiteral string
	}{
		{
			name:            "doubled single quote in single-quoted string",
			input:           `'it''s'`,
			expectedLiteral: `it's`,
		},
		{
			name:            "doubled double quote in double-quoted string",
			input:           `"say ""hello"""`,
			expectedLiteral: `say "hello"`,
		},
		{
			name:            "embedded double quotes in single-quoted string (no escaping needed)",
			input:           `'she said "hi"'`,
			expectedLiteral: `she said "hi"`,
		},
		{
			name:            "embedded single quotes in double-quoted string (no escaping needed)",
			input:           `"it's fine"`,
			expectedLiteral: `it's fine`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != STRING {
				t.Fatalf("expected STRING token, got %q", tok.Type)
			}

			if tok.Literal != tt.expectedLiteral {
				t.Errorf("literal wrong. expected=%q, got=%q",
					tt.expectedLiteral, tok.Literal)
			}
		})
	}
}

// TestCharLiteralsCRLF tests character literals specifically for CR and LF
// Character literals are now automatically converted to their actual character values
// and returned as STRING tokens (with implicit concatenation support)
func TestCharLiteralsCRLF(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedLiteral string
	}{
		{
			name:            "carriage return #13",
			input:           "#13",
			expectedLiteral: "\r",
		},
		{
			name:            "line feed #10",
			input:           "#10",
			expectedLiteral: "\n",
		},
		{
			name:            "CR+LF sequence",
			input:           "#13#10",
			expectedLiteral: "\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != STRING {
				t.Fatalf("expected STRING token, got %q", tok.Type)
			}

			if tok.Literal != tt.expectedLiteral {
				t.Errorf("literal wrong. expected=%q, got=%q",
					tt.expectedLiteral, tok.Literal)
			}
		})
	}
}
