package token

import (
	"testing"
	"unicode/utf8"
)

func TestTokenEnd(t *testing.T) {
	tests := []struct {
		name           string
		token          Token
		expectedLine   int
		expectedColumn int
		expectedOffset int
	}{
		{
			name: "ASCII token",
			token: Token{
				Type:    IDENT,
				Literal: "var",
				Pos:     Position{Line: 1, Column: 1, Offset: 0},
			},
			expectedLine:   1,
			expectedColumn: 4, // 1 + 3 runes
			expectedOffset: 3, // 0 + 3 bytes
		},
		{
			name: "multi-byte UTF-8 token",
			token: Token{
				Type:    IDENT,
				Literal: "Δ", // Greek delta: 2 bytes, 1 rune
				Pos:     Position{Line: 1, Column: 5, Offset: 4},
			},
			expectedLine:   1,
			expectedColumn: 6, // 5 + 1 rune (not 5 + 2 bytes!)
			expectedOffset: 6, // 4 + 2 bytes
		},
		{
			name: "emoji token",
			token: Token{
				Type:    STRING,
				Literal: "🚀", // Rocket emoji: 4 bytes, 1 rune
				Pos:     Position{Line: 2, Column: 10, Offset: 100},
			},
			expectedLine:   2,
			expectedColumn: 11,  // 10 + 1 rune (not 10 + 4 bytes!)
			expectedOffset: 104, // 100 + 4 bytes
		},
		{
			name: "mixed ASCII and multi-byte",
			token: Token{
				Type:    STRING,
				Literal: "hello世界", // 5 ASCII + 2 Chinese chars = 7 runes, 11 bytes
				Pos:     Position{Line: 3, Column: 1, Offset: 0},
			},
			expectedLine:   3,
			expectedColumn: 8,  // 1 + 7 runes
			expectedOffset: 11, // 0 + 11 bytes
		},
		{
			name: "empty token",
			token: Token{
				Type:    EOF,
				Literal: "",
				Pos:     Position{Line: 5, Column: 20, Offset: 200},
			},
			expectedLine:   5,
			expectedColumn: 20,  // 20 + 0 runes
			expectedOffset: 200, // 200 + 0 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			end := tt.token.End()

			if end.Line != tt.expectedLine {
				t.Errorf("End().Line = %d, want %d", end.Line, tt.expectedLine)
			}
			if end.Column != tt.expectedColumn {
				t.Errorf("End().Column = %d, want %d (literal: %q, %d bytes, %d runes)",
					end.Column, tt.expectedColumn, tt.token.Literal,
					len(tt.token.Literal), utf8.RuneCountInString(tt.token.Literal))
			}
			if end.Offset != tt.expectedOffset {
				t.Errorf("End().Offset = %d, want %d", end.Offset, tt.expectedOffset)
			}
		})
	}
}

func TestTokenLength(t *testing.T) {
	tests := []struct {
		name     string
		token    Token
		expected int // expected rune count
	}{
		{
			name:     "ASCII keyword",
			token:    Token{Type: BEGIN, Literal: "begin"},
			expected: 5,
		},
		{
			name:     "ASCII identifier",
			token:    Token{Type: IDENT, Literal: "myVariable"},
			expected: 10,
		},
		{
			name:     "single multi-byte character",
			token:    Token{Type: IDENT, Literal: "Δ"}, // 2 bytes, 1 rune
			expected: 1,
		},
		{
			name:     "emoji",
			token:    Token{Type: STRING, Literal: "🚀"}, // 4 bytes, 1 rune
			expected: 1,
		},
		{
			name:     "mixed ASCII and multi-byte",
			token:    Token{Type: STRING, Literal: "hello世界"}, // 11 bytes, 7 runes
			expected: 7,
		},
		{
			name:     "multiple emoji",
			token:    Token{Type: STRING, Literal: "🚀🌟💻"}, // 12 bytes, 3 runes
			expected: 3,
		},
		{
			name:     "empty literal",
			token:    Token{Type: EOF, Literal: ""},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.Length()
			if got != tt.expected {
				t.Errorf("Token.Length() = %d, want %d (literal: %q, bytes: %d)",
					got, tt.expected, tt.token.Literal, len(tt.token.Literal))
			}
		})
	}
}
