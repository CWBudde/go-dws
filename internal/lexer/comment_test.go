package lexer

import (
	"testing"
)

// TestCommentPreservation tests that comments are preserved when enabled
func TestCommentPreservation(t *testing.T) {
	input := `// Line comment
var x := 42; { Block comment }
(* Paren comment *)
/* C-style comment */`

	tests := []struct {
		name            string
		preserveComments bool
		wantCommentCount int
	}{
		{
			name:            "skip comments (default)",
			preserveComments: false,
			wantCommentCount: 0,
		},
		{
			name:            "preserve comments",
			preserveComments: true,
			wantCommentCount: 4, // 4 comments in the input
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(input)
			l.SetPreserveComments(tt.preserveComments)

			commentCount := 0
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
				if tok.Type == COMMENT {
					commentCount++
				}
			}

			if commentCount != tt.wantCommentCount {
				t.Errorf("got %d comments, want %d", commentCount, tt.wantCommentCount)
			}
		})
	}
}

// TestLineComment tests line comment preservation
func TestLineComment(t *testing.T) {
	input := `// This is a line comment
var x := 42;`

	l := New(input)
	l.SetPreserveComments(true)

	tok := l.NextToken()
	if tok.Type != COMMENT {
		t.Fatalf("expected COMMENT token, got %s", tok.Type)
	}

	if tok.Literal != "// This is a line comment" {
		t.Errorf("got %q, want %q", tok.Literal, "// This is a line comment")
	}

	// Next token should be VAR
	tok = l.NextToken()
	if tok.Type != VAR {
		t.Errorf("expected VAR token after comment, got %s", tok.Type)
	}
}

// TestBlockCommentCurly tests curly brace block comment preservation
func TestBlockCommentCurly(t *testing.T) {
	input := `{ This is a block comment }
var x := 42;`

	l := New(input)
	l.SetPreserveComments(true)

	tok := l.NextToken()
	if tok.Type != COMMENT {
		t.Fatalf("expected COMMENT token, got %s", tok.Type)
	}

	if tok.Literal != "{ This is a block comment }" {
		t.Errorf("got %q, want %q", tok.Literal, "{ This is a block comment }")
	}

	// Next token should be VAR
	tok = l.NextToken()
	if tok.Type != VAR {
		t.Errorf("expected VAR token after comment, got %s", tok.Type)
	}
}

// TestBlockCommentParen tests parenthesis block comment preservation
func TestBlockCommentParen(t *testing.T) {
	input := `(* This is a paren comment *)
var x := 42;`

	l := New(input)
	l.SetPreserveComments(true)

	tok := l.NextToken()
	if tok.Type != COMMENT {
		t.Fatalf("expected COMMENT token, got %s", tok.Type)
	}

	if tok.Literal != "(* This is a paren comment *)" {
		t.Errorf("got %q, want %q", tok.Literal, "(* This is a paren comment *)")
	}

	// Next token should be VAR
	tok = l.NextToken()
	if tok.Type != VAR {
		t.Errorf("expected VAR token after comment, got %s", tok.Type)
	}
}

// TestCStyleComment tests C-style comment preservation
func TestCStyleComment(t *testing.T) {
	input := `/* This is a C-style comment */
var x := 42;`

	l := New(input)
	l.SetPreserveComments(true)

	tok := l.NextToken()
	if tok.Type != COMMENT {
		t.Fatalf("expected COMMENT token, got %s", tok.Type)
	}

	if tok.Literal != "/* This is a C-style comment */" {
		t.Errorf("got %q, want %q", tok.Literal, "/* This is a C-style comment */")
	}

	// Next token should be VAR
	tok = l.NextToken()
	if tok.Type != VAR {
		t.Errorf("expected VAR token after comment, got %s", tok.Type)
	}
}

// TestMultilineComment tests multi-line comment preservation
func TestMultilineComment(t *testing.T) {
	input := `{ This is a
multi-line
block comment }
var x := 42;`

	l := New(input)
	l.SetPreserveComments(true)

	tok := l.NextToken()
	if tok.Type != COMMENT {
		t.Fatalf("expected COMMENT token, got %s", tok.Type)
	}

	expected := `{ This is a
multi-line
block comment }`
	if tok.Literal != expected {
		t.Errorf("got %q, want %q", tok.Literal, expected)
	}

	// Verify position tracking
	if tok.Pos.Line != 1 {
		t.Errorf("expected comment to start at line 1, got %d", tok.Pos.Line)
	}

	// Next token should be VAR on line 4
	tok = l.NextToken()
	if tok.Type != VAR {
		t.Errorf("expected VAR token after comment, got %s", tok.Type)
	}
	if tok.Pos.Line != 4 {
		t.Errorf("expected VAR at line 4, got line %d", tok.Pos.Line)
	}
}

// TestCommentPosition tests that comment positions are tracked correctly
func TestCommentPosition(t *testing.T) {
	input := `// Line 1
var x := 42; // Line 2
{ Line 3 }`

	l := New(input)
	l.SetPreserveComments(true)

	// First comment on line 1
	tok := l.NextToken()
	if tok.Type != COMMENT || tok.Pos.Line != 1 {
		t.Errorf("expected COMMENT at line 1, got %s at line %d", tok.Type, tok.Pos.Line)
	}

	// VAR on line 2
	tok = l.NextToken()
	if tok.Type != VAR || tok.Pos.Line != 2 {
		t.Errorf("expected VAR at line 2, got %s at line %d", tok.Type, tok.Pos.Line)
	}

	// Skip IDENT, ASSIGN, INT, SEMICOLON
	l.NextToken() // IDENT
	l.NextToken() // ASSIGN
	l.NextToken() // INT
	l.NextToken() // SEMICOLON

	// Second comment on line 2
	tok = l.NextToken()
	if tok.Type != COMMENT || tok.Pos.Line != 2 {
		t.Errorf("expected COMMENT at line 2, got %s at line %d", tok.Type, tok.Pos.Line)
	}

	// Third comment on line 3
	tok = l.NextToken()
	if tok.Type != COMMENT || tok.Pos.Line != 3 {
		t.Errorf("expected COMMENT at line 3, got %s at line %d", tok.Type, tok.Pos.Line)
	}
}

// TestUnterminatedComment tests unterminated comment handling
func TestUnterminatedComment(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unterminated curly brace comment",
			input: `{ This comment never ends`,
		},
		{
			name:  "unterminated paren comment",
			input: `(* This comment never ends`,
		},
		{
			name:  "unterminated C-style comment",
			input: `/* This comment never ends`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			l.SetPreserveComments(true)

			tok := l.NextToken()
			if tok.Type != ILLEGAL {
				t.Errorf("expected ILLEGAL token for unterminated comment, got %s", tok.Type)
			}
		})
	}
}

// TestPreserveCommentsSetting tests the preserve comments getter
func TestPreserveCommentsSetting(t *testing.T) {
	l := New("var x := 42;")

	// Default should be false
	if l.PreserveComments() {
		t.Error("expected PreserveComments to be false by default")
	}

	// Set to true
	l.SetPreserveComments(true)
	if !l.PreserveComments() {
		t.Error("expected PreserveComments to be true after setting")
	}

	// Set to false
	l.SetPreserveComments(false)
	if l.PreserveComments() {
		t.Error("expected PreserveComments to be false after unsetting")
	}
}
