package lexer

import (
	"testing"
)

// TestTripleQuoteStrings tests heredoc-style (triple-quoted) string literals.
func TestTripleQuoteStrings(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedLiteral string
	}{
		{
			name:            "basic heredoc with matching indent",
			input:           "'''\n   hello world\n   '''",
			expectedLiteral: "hello world",
		},
		{
			name:            "no indent",
			input:           "'''\nno indent\n'''",
			expectedLiteral: "no indent",
		},
		{
			name:            "deeper content keeps extra indent",
			input:           "'''\n      three indents\n   '''",
			expectedLiteral: "   three indents",
		},
		{
			name:            "empty lines preserved, tab indent",
			input:           "'''\n\n\t\tempty lines\n\n\t\t'''",
			expectedLiteral: "\nempty lines\n",
		},
		{
			name:            "quotes inside heredoc are verbatim",
			input:           "'''\n   l'appel du \"lyon\" ou ''du loup''\n   '''",
			expectedLiteral: "l'appel du \"lyon\" ou ''du loup''",
		},
		{
			name:            "multiple content lines",
			input:           "'''\n  line one\n  line two\n  '''",
			expectedLiteral: "line one\nline two",
		},
		{
			name:            "CRLF line endings",
			input:           "'''\r\n   hello\r\n   '''",
			expectedLiteral: "hello",
		},
		{
			name:            "triple double quotes",
			input:           "\"\"\"\n   hello\n   \"\"\"",
			expectedLiteral: "hello",
		},
		{
			name:            "empty heredoc",
			input:           "'''\n'''",
			expectedLiteral: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != STRING {
				t.Fatalf("tokentype wrong. expected=%q, got=%q", STRING, tok.Type)
			}
			if tok.Literal != tt.expectedLiteral {
				t.Fatalf("literal wrong. expected=%q, got=%q", tt.expectedLiteral, tok.Literal)
			}
			if errs := l.Errors(); len(errs) != 0 {
				t.Fatalf("unexpected lexer errors: %v", errs)
			}
		})
	}
}

// TestTripleQuoteNotHeredoc verifies that quote sequences that are not
// heredoc openers keep their regular string semantics.
func TestTripleQuoteNotHeredoc(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedLiteral string
	}{
		{
			name:            "four quotes is an escaped quote",
			input:           "''''",
			expectedLiteral: "'",
		},
		{
			name:            "six quotes is two escaped quotes",
			input:           "''''''",
			expectedLiteral: "''",
		},
		{
			name:            "empty string",
			input:           "''",
			expectedLiteral: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != STRING {
				t.Fatalf("tokentype wrong. expected=%q, got=%q", STRING, tok.Type)
			}
			if tok.Literal != tt.expectedLiteral {
				t.Fatalf("literal wrong. expected=%q, got=%q", tt.expectedLiteral, tok.Literal)
			}
		})
	}
}

// TestTripleQuoteErrors tests error cases for heredoc strings.
func TestTripleQuoteErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unterminated heredoc",
			input: "'''\n   hello",
		},
		{
			name:  "terminator preceded by non-whitespace",
			input: "'''\n   hello x'''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			l.NextToken()
			if errs := l.Errors(); len(errs) == 0 {
				t.Fatalf("expected lexer error, got none")
			}
		})
	}
}
