package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestPeekHelper tests the peek() helper method for N-token lookahead.
func TestPeekHelper(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		n        int
		expected lexer.TokenType
	}{
		{
			name:     "peek(0) returns token after peekToken",
			input:    "var x : Integer",
			n:        0,
			expected: lexer.COLON,
		},
		{
			name:     "peek(1) returns 2 tokens ahead of peekToken",
			input:    "var x : Integer",
			n:        1,
			expected: lexer.IDENT, // "Integer"
		},
		{
			name:     "peek(2) returns 3 tokens ahead of peekToken",
			input:    "var x : Integer := 42",
			n:        2,
			expected: lexer.ASSIGN,
		},
		{
			name:     "peek(0) with const declaration",
			input:    "const C = 5",
			n:        0,
			expected: lexer.EQ,
		},
		{
			name:     "peek(0) with type declaration",
			input:    "type TFoo = Integer",
			n:        0,
			expected: lexer.EQ,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			// After New(), curToken = first token, peekToken = second token

			result := p.peek(tt.n)
			if result.Type != tt.expected {
				t.Errorf("peek(%d) = %s, want %s", tt.n, result.Type, tt.expected)
			}
		})
	}
}

// TestPeekAheadHelper tests the peekAhead() helper method.
func TestPeekAheadHelper(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		n        int
		expected lexer.TokenType
	}{
		{
			name:     "peekAhead(0) returns curToken",
			input:    "var x : Integer",
			n:        0,
			expected: lexer.VAR,
		},
		{
			name:     "peekAhead(1) returns peekToken",
			input:    "var x : Integer",
			n:        1,
			expected: lexer.IDENT, // "x"
		},
		{
			name:     "peekAhead(2) returns token after peekToken",
			input:    "var x : Integer",
			n:        2,
			expected: lexer.COLON,
		},
		{
			name:     "peekAhead(3) returns 2 tokens after peekToken",
			input:    "var x : Integer",
			n:        3,
			expected: lexer.IDENT, // "Integer"
		},
		{
			name:     "peekAhead(4) returns 3 tokens after peekToken",
			input:    "var x : Integer := 42",
			n:        4,
			expected: lexer.ASSIGN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			// After New(), curToken = first token, peekToken = second token

			result := p.peekAhead(tt.n)
			if result.Type != tt.expected {
				t.Errorf("peekAhead(%d) = %s, want %s", tt.n, result.Type, tt.expected)
			}
		})
	}
}

// TestLooksLikeVarDeclaration tests the var declaration disambiguation.
func TestLooksLikeVarDeclaration(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		allowInferred bool
		expected      bool
	}{
		{
			name:          "explicit type - should be var declaration",
			input:         "var x : Integer",
			allowInferred: false,
			expected:      true,
		},
		{
			name:          "multi-var declaration with comma",
			input:         "var x, y : Integer",
			allowInferred: false,
			expected:      true,
		},
		{
			name:          "var with assignment - ambiguous when inference disabled",
			input:         "var x := 5",
			allowInferred: false,
			expected:      false,
		},
		{
			name:          "var with assignment - allowed in block var section",
			input:         "var x := 5",
			allowInferred: true,
			expected:      true,
		},
		{
			name:          "var with equals - ambiguous when inference disabled",
			input:         "var x = 5",
			allowInferred: false,
			expected:      false,
		},
		{
			name:          "var with equals - allowed in block var section",
			input:         "var x = 5",
			allowInferred: true,
			expected:      true,
		},
		{
			name:          "not starting with ident - should be false",
			input:         "var begin end",
			allowInferred: true,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			// After New(), curToken = VAR, peekToken = IDENT or other

			result := p.looksLikeVarDeclaration(p.cursor, tt.allowInferred)
			if result != tt.expected {
				t.Errorf("looksLikeVarDeclaration() = %v, want %v for input: %s",
					result, tt.expected, tt.input)
			}
		})
	}
}

// TestLooksLikeConstDeclaration tests the const declaration disambiguation.
func TestLooksLikeConstDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "typed const with colon",
			input:    "const C : Integer = 5",
			expected: true,
		},
		{
			name:     "untyped const with equals",
			input:    "const C = 5",
			expected: true,
		},
		{
			name:     "const with assignment operator - ambiguous",
			input:    "const C := 5",
			expected: false,
		},
		{
			name:     "not starting with ident - should be false",
			input:    "const begin end",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			// After New(), curToken = CONST, peekToken = IDENT or other

			result := p.looksLikeConstDeclaration(p.cursor)
			if result != tt.expected {
				t.Errorf("looksLikeConstDeclaration() = %v, want %v for input: %s",
					result, tt.expected, tt.input)
			}
		})
	}
}

// TestLooksLikeTypeDeclaration tests the type declaration disambiguation.
func TestLooksLikeTypeDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "type alias with equals",
			input:    "type TFoo = Integer",
			expected: true,
		},
		{
			name:     "type declaration with class",
			input:    "type TFoo = class",
			expected: true,
		},
		{
			name:     "not starting with ident - should be false",
			input:    "type class end",
			expected: false,
		},
		{
			name:     "ident not followed by equals - should be false",
			input:    "type TFoo : Integer",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			// After New(), curToken = TYPE, peekToken = IDENT or other

			result := p.looksLikeTypeDeclaration()
			if result != tt.expected {
				t.Errorf("looksLikeTypeDeclaration() = %v, want %v for input: %s",
					result, tt.expected, tt.input)
			}
		})
	}
}

// TestLookaheadInComplexDeclarations tests lookahead in complex declaration scenarios.
func TestLookaheadInComplexDeclarations(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "var with explicit type",
			input:       "var x : Integer := 42;",
			expectError: false,
		},
		{
			name:        "const with type and value",
			input:       "const C : Integer = 100;",
			expectError: false,
		},
		{
			name:        "const with inferred type",
			input:       "const C = 100;",
			expectError: false,
		},
		{
			name:        "multi-var declaration",
			input:       "var x, y, z : Integer;",
			expectError: false,
		},
		{
			name:        "type alias",
			input:       "type TMyInt = Integer;",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			hasError := len(p.Errors()) > 0
			if hasError != tt.expectError {
				t.Errorf("Expected error: %v, got errors: %v", tt.expectError, p.Errors())
			}

			if !tt.expectError && len(program.Statements) == 0 {
				t.Errorf("Expected at least one statement, got none")
			}
		})
	}
}

// TestPeekPerformance ensures lookahead doesn't cause performance regression.
// This is a basic sanity check - actual benchmarks are in parser_bench_test.go.
func TestPeekPerformance(t *testing.T) {
	input := `
var x : Integer := 42;
const C = 100;
type TFoo = Integer;
var y, z : String;
const MAX : Integer = 1000;
type TBar = class
	Field : Integer;
end;
`

	// This should complete quickly (< 100ms) even with lookahead
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Errorf("Unexpected parsing errors: %v", p.Errors())
	}

	if len(program.Statements) < 5 {
		t.Errorf("Expected at least 5 statements, got %d", len(program.Statements))
	}
}

// TestPeekConsistency verifies that peek(n) and peekAhead(n) are consistent.
func TestPeekConsistency(t *testing.T) {
	input := "var x : Integer := 42"
	l := lexer.New(input)
	p := New(l)

	// peekAhead(1) should equal peekToken
	if p.peekAhead(1).Type != p.cursor.Peek(1).Type {
		t.Errorf("peekAhead(1) = %s, but peekToken = %s",
			p.peekAhead(1).Type, p.cursor.Peek(1).Type)
	}

	// peekAhead(2) should equal peek(0)
	if p.peekAhead(2).Type != p.peek(0).Type {
		t.Errorf("peekAhead(2) = %s, but peek(0) = %s",
			p.peekAhead(2).Type, p.peek(0).Type)
	}

	// peekAhead(3) should equal peek(1)
	if p.peekAhead(3).Type != p.peek(1).Type {
		t.Errorf("peekAhead(3) = %s, but peek(1) = %s",
			p.peekAhead(3).Type, p.peek(1).Type)
	}
}

// TestPeekBeyondEOF tests that peeking beyond EOF returns EOF tokens.
func TestPeekBeyondEOF(t *testing.T) {
	input := "var x"
	l := lexer.New(input)
	p := New(l)

	// Peek far beyond the input
	tok := p.peek(100)
	if tok.Type != lexer.EOF {
		t.Errorf("peek(100) beyond EOF should return EOF, got %s", tok.Type)
	}
}

// TestLookaheadWithComplexTypes tests lookahead with complex type expressions.
func TestLookaheadWithComplexTypes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "array type",
			input: "var arr : array of Integer;",
		},
		{
			name:  "array type with bounds",
			input: "var arr : array[0..9] of Integer;",
		},
		{
			name:  "function pointer type",
			input: "var fn : function(x: Integer): Boolean;",
		},
		{
			name:  "procedure pointer type",
			input: "var proc : procedure(s: String);",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Errorf("Unexpected parsing errors: %v", p.Errors())
			}

			if len(program.Statements) == 0 {
				t.Errorf("Expected at least one statement, got none")
			}
		})
	}
}
