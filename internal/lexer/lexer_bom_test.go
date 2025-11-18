package lexer

import (
	"testing"
)

func TestBOMHandling(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectLit   string
		expectFirst TokenType
	}{
		{
			name:        "UTF-8 BOM followed by var",
			input:       "\xEF\xBB\xBFvar x := 5;",
			expectFirst: VAR,
			expectLit:   "var",
		},
		{
			name:        "UTF-8 BOM followed by comment then var",
			input:       "\xEF\xBB\xBF// comment\nvar x := 5;",
			expectFirst: VAR,
			expectLit:   "var",
		},
		{
			name:        "UTF-8 BOM followed by function",
			input:       "\xEF\xBB\xBFfunction Test(): Integer;\nbegin\nend;",
			expectFirst: FUNCTION,
			expectLit:   "function",
		},
		{
			name:        "No BOM - ensure no regression",
			input:       "var x := 5;",
			expectFirst: VAR,
			expectLit:   "var",
		},
		{
			name:        "Partial BOM (2 bytes) should treat as ILLEGAL",
			input:       "\xEF\xBBvar x := 5;",
			expectFirst: ILLEGAL,
			expectLit:   "\uFFFD", // Unicode replacement character for invalid UTF-8
		},
		{
			name:        "Empty file with just BOM",
			input:       "\xEF\xBB\xBF",
			expectFirst: EOF,
			expectLit:   "",
		},
		{
			name:        "UTF-8 BOM followed by integer",
			input:       "\xEF\xBB\xBF42",
			expectFirst: INT,
			expectLit:   "42",
		},
		{
			name:        "UTF-8 BOM followed by string",
			input:       "\xEF\xBB\xBF\"hello\"",
			expectFirst: STRING,
			expectLit:   "hello",
		},
		{
			name:        "UTF-8 BOM with whitespace",
			input:       "\xEF\xBB\xBF   var x := 5;",
			expectFirst: VAR,
			expectLit:   "var",
		},
		{
			name:        "UTF-8 BOM with block comment",
			input:       "\xEF\xBB\xBF{ comment }\nvar x;",
			expectFirst: VAR,
			expectLit:   "var",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			// Skip whitespace and comments to get to first real token
			for tok.Type == EOF {
				break
			}

			if tok.Type != tt.expectFirst {
				t.Errorf("expected first token type %v, got %v", tt.expectFirst, tok.Type)
			}

			if tok.Literal != tt.expectLit {
				t.Errorf("expected first token literal %q, got %q", tt.expectLit, tok.Literal)
			}

			// For ILLEGAL tokens with partial BOM, verify position is correct
			if tt.expectFirst == ILLEGAL && tok.Pos.Line != 1 {
				t.Errorf("expected ILLEGAL token at line 1, got line %d", tok.Pos.Line)
			}
		})
	}
}

func TestBOMWithRealFixtureContent(t *testing.T) {
	// Test with actual content similar to the failing fixture files
	tests := []struct {
		name      string
		input     string
		numTokens int // Approximate number of tokens to verify parsing continues
	}{
		{
			name: "BOM with comment and var declaration",
			input: "\xEF\xBB\xBF// rc algos\n" +
				"var i: Integer;\n" +
				"begin\n" +
				"  i := 1;\n" +
				"end;",
			numTokens: 10, // var, i, :, Integer, ;, begin, i, :=, 1, ;, end, ;
		},
		{
			name: "BOM with type declaration",
			input: "\xEF\xBB\xBFtype\n" +
				"  TMyType = Integer;\n" +
				"var x: TMyType;",
			numTokens: 8, // type, TMyType, =, Integer, ;, var, x, :, TMyType, ;
		},
		{
			name: "BOM with procedure",
			input: "\xEF\xBB\xBFprocedure Test;\n" +
				"begin\n" +
				"end;",
			numTokens: 4, // procedure, Test, ;, begin, end, ;
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			tokenCount := 0
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
				tokenCount++

				// Verify no ILLEGAL tokens are produced
				if tok.Type == ILLEGAL {
					t.Errorf("unexpected ILLEGAL token at position %d: %q", tok.Pos.Offset, tok.Literal)
				}
			}

			if tokenCount < tt.numTokens {
				t.Errorf("expected at least %d tokens, got %d", tt.numTokens, tokenCount)
			}
		})
	}
}
