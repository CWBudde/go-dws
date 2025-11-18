package lexer

import (
	"testing"
)

func TestComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wants []TokenType
	}{
		{
			name: "line comment",
			input: `x // this is a comment
y`,
			wants: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name:  "block comment with braces",
			input: `x { comment } y`,
			wants: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name:  "block comment with parens",
			input: `x (* comment *) y`,
			wants: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name:  "C-style comment",
			input: `x /* comment */ y`,
			wants: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name: "multiline C-style comment",
			input: `x /* this is a
				multiline comment */ y`,
			wants: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name: "multiline block comment",
			input: `x {
				this is a
				multiline comment
			} y`,
			wants: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name: "multiple comments",
			input: `x // comment 1
y { comment 2 } z`,
			wants: []TokenType{IDENT, IDENT, IDENT, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			for i, expectedType := range tt.wants {
				tok := l.NextToken()
				if tok.Type != expectedType {
					t.Fatalf("token[%d] wrong. expected=%q, got=%q (literal=%q)",
						i, expectedType, tok.Type, tok.Literal)
				}
			}
		})
	}
}

func TestCompilerDirective(t *testing.T) {
	input := `{$DEFINE DEBUG}
	x := 5;`

	tests := []struct {
		expectedType TokenType
	}{
		{IDENT}, // x
		{ASSIGN},
		{INT},
		{SEMICOLON},
		{EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType TokenType
	}{
		{
			name:         "empty input",
			input:        "",
			expectedType: EOF,
		},
		{
			name:         "only whitespace",
			input:        "   \t\n  ",
			expectedType: EOF,
		},
		{
			name:         "illegal character",
			input:        "`",
			expectedType: ILLEGAL,
		},
		{
			name:         "unterminated string",
			input:        "'hello",
			expectedType: STRING, // Now returns STRING token with error accumulated
		},
		{
			name:         "unterminated block comment",
			input:        "{ comment",
			expectedType: EOF, // Comments skip to EOF, error is accumulated
		},
		{
			name:         "unterminated C-style comment",
			input:        "/* comment",
			expectedType: EOF, // Comments skip to EOF, error is accumulated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.expectedType {
				t.Fatalf("tokentype wrong. expected=%q, got=%q (literal=%q)",
					tt.expectedType, tok.Type, tok.Literal)
			}

			// For error cases, verify error was accumulated (Task 12.1)
			if tt.name == "unterminated string" || tt.name == "unterminated block comment" ||
				tt.name == "unterminated C-style comment" || tt.name == "illegal character" {
				errors := l.Errors()
				if len(errors) == 0 {
					t.Errorf("expected error to be accumulated, but got none")
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkLexer(b *testing.B) {
	input := `
	function Fibonacci(n: Integer): Integer;
	begin
		if n <= 1 then
			Result := n
		else
			Result := Fibonacci(n-1) + Fibonacci(n-2);
	end;

	var i: Integer;
	begin
		for i := 0 to 10 do
			PrintLn(Fibonacci(i));
	end.
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

func BenchmarkLexerKeywords(b *testing.B) {
	input := "begin end if then else while for do function procedure class var const"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

func BenchmarkLexerNumbers(b *testing.B) {
	input := "123 456 789 $FF $10 0xFF 123.45 1.5e10 %1010"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

func BenchmarkLexerStrings(b *testing.B) {
	input := `'hello' 'world' 'this is a test' 'it''s working'`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}
