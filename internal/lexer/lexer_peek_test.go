package lexer

import (
	"fmt"
	"testing"
)

// TestPeekToken tests basic token peeking functionality (Task 12.3.2)
func TestPeekToken(t *testing.T) {
	input := "var x := 5;"
	l := New(input)

	// Peek at first token without consuming
	tok := l.Peek(0)
	if tok.Type != VAR || tok.Literal != "var" {
		t.Fatalf("Peek(0) expected VAR(var), got %s(%s)", tok.Type, tok.Literal)
	}

	// Peek again - should get same token
	tok2 := l.Peek(0)
	if tok2.Type != VAR || tok2.Literal != "var" {
		t.Fatalf("Peek(0) second call expected VAR(var), got %s(%s)", tok2.Type, tok2.Literal)
	}

	// Peek ahead 1 token
	tok3 := l.Peek(1)
	if tok3.Type != IDENT || tok3.Literal != "x" {
		t.Fatalf("Peek(1) expected IDENT(x), got %s(%s)", tok3.Type, tok3.Literal)
	}

	// First peek should still return var
	tok4 := l.Peek(0)
	if tok4.Type != VAR {
		t.Fatalf("Peek(0) after Peek(1) expected VAR, got %s", tok4.Type)
	}

	// Now consume the first token
	consumed := l.NextToken()
	if consumed.Type != VAR || consumed.Literal != "var" {
		t.Fatalf("NextToken() expected VAR(var), got %s(%s)", consumed.Type, consumed.Literal)
	}

	// Peek should now return the next token
	tok5 := l.Peek(0)
	if tok5.Type != IDENT || tok5.Literal != "x" {
		t.Fatalf("Peek(0) after NextToken() expected IDENT(x), got %s(%s)", tok5.Type, tok5.Literal)
	}
}

// TestPeekMultipleTokens tests peeking multiple tokens ahead (Task 12.3.2)
func TestPeekMultipleTokens(t *testing.T) {
	input := "var x := 5;"
	l := New(input)

	tests := []struct {
		expectedLit  string
		peekN        int
		expectedType TokenType
	}{
		{expectedLit: "var", peekN: 0, expectedType: VAR},
		{expectedLit: "x", peekN: 1, expectedType: IDENT},
		{expectedLit: ":=", peekN: 2, expectedType: ASSIGN},
		{expectedLit: "5", peekN: 3, expectedType: INT},
		{expectedLit: ";", peekN: 4, expectedType: SEMICOLON},
		{expectedLit: "", peekN: 5, expectedType: EOF},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Peek(%d)", tt.peekN), func(t *testing.T) {
			tok := l.Peek(tt.peekN)
			if tok.Type != tt.expectedType {
				t.Errorf("Peek(%d) type: expected %s, got %s", tt.peekN, tt.expectedType, tok.Type)
			}
			if tok.Literal != tt.expectedLit {
				t.Errorf("Peek(%d) literal: expected %q, got %q", tt.peekN, tt.expectedLit, tok.Literal)
			}
		})
	}

	// Peek should not consume - first NextToken should still return VAR
	tok := l.NextToken()
	if tok.Type != VAR {
		t.Errorf("After all Peeks, NextToken() expected VAR, got %s", tok.Type)
	}
}

// TestPeekAndConsumeInterleaved tests interleaving Peek and NextToken (Task 12.3.2)
func TestPeekAndConsumeInterleaved(t *testing.T) {
	input := "a b c d"
	l := New(input)

	// Peek ahead to 'b'
	tok := l.Peek(1)
	if tok.Literal != "b" {
		t.Fatalf("Peek(1) expected 'b', got %q", tok.Literal)
	}

	// Consume 'a'
	tok = l.NextToken()
	if tok.Literal != "a" {
		t.Fatalf("NextToken() expected 'a', got %q", tok.Literal)
	}

	// Peek ahead to 'c'
	tok = l.Peek(1)
	if tok.Literal != "c" {
		t.Fatalf("Peek(1) expected 'c', got %q", tok.Literal)
	}

	// Peek 0 should be 'b'
	tok = l.Peek(0)
	if tok.Literal != "b" {
		t.Fatalf("Peek(0) expected 'b', got %q", tok.Literal)
	}

	// Consume 'b'
	tok = l.NextToken()
	if tok.Literal != "b" {
		t.Fatalf("NextToken() expected 'b', got %q", tok.Literal)
	}

	// Peek 0 should now be 'c'
	tok = l.Peek(0)
	if tok.Literal != "c" {
		t.Fatalf("Peek(0) expected 'c', got %q", tok.Literal)
	}

	// Peek 1 should now be 'd'
	tok = l.Peek(1)
	if tok.Literal != "d" {
		t.Fatalf("Peek(1) expected 'd', got %q", tok.Literal)
	}
}

// TestPeekWithComments tests that Peek handles comments correctly (Task 12.3.2)
func TestPeekWithComments(t *testing.T) {
	input := "var // comment\nx := 5"
	l := New(input)

	// Peek should skip the comment
	tok := l.Peek(0)
	if tok.Type != VAR || tok.Literal != "var" {
		t.Fatalf("Peek(0) expected VAR(var), got %s(%s)", tok.Type, tok.Literal)
	}

	// Peek 1 should be 'x', skipping the comment
	tok = l.Peek(1)
	if tok.Type != IDENT || tok.Literal != "x" {
		t.Fatalf("Peek(1) expected IDENT(x), got %s(%s)", tok.Type, tok.Literal)
	}

	// Consume tokens should also skip comments
	tok = l.NextToken()
	if tok.Type != VAR {
		t.Fatalf("NextToken() expected VAR, got %s", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type != IDENT {
		t.Fatalf("NextToken() expected IDENT, got %s", tok.Type)
	}
}

// TestNextTokenAfterPeekConsumesBuffer tests that NextToken consumes buffered tokens (Task 12.3.3)
func TestNextTokenAfterPeekConsumesBuffer(t *testing.T) {
	input := "a b c"
	l := New(input)

	// Peek ahead multiple times - this buffers tokens
	_ = l.Peek(0) // buffers 'a'
	_ = l.Peek(1) // buffers 'b'
	_ = l.Peek(2) // buffers 'c'

	// NextToken should consume from buffer
	tok1 := l.NextToken()
	if tok1.Literal != "a" {
		t.Fatalf("NextToken() expected 'a', got %q", tok1.Literal)
	}

	tok2 := l.NextToken()
	if tok2.Literal != "b" {
		t.Fatalf("NextToken() expected 'b', got %q", tok2.Literal)
	}

	tok3 := l.NextToken()
	if tok3.Literal != "c" {
		t.Fatalf("NextToken() expected 'c', got %q", tok3.Literal)
	}

	// Should be at EOF now
	tok4 := l.NextToken()
	if tok4.Type != EOF {
		t.Fatalf("NextToken() expected EOF, got %s", tok4.Type)
	}
}

// TestPeekPreservesExistingBehavior tests that existing code still works (Task 12.3.3)
func TestPeekPreservesExistingBehavior(t *testing.T) {
	input := "var x := 5;"
	l := New(input)

	// Consume tokens as usual without using Peek
	tokens := []struct {
		lit string
		typ TokenType
	}{
		{lit: "var", typ: VAR},
		{lit: "x", typ: IDENT},
		{lit: ":=", typ: ASSIGN},
		{lit: "5", typ: INT},
		{lit: ";", typ: SEMICOLON},
		{lit: "", typ: EOF},
	}

	for i, expected := range tokens {
		tok := l.NextToken()
		if tok.Type != expected.typ {
			t.Errorf("token[%d] type: expected %s, got %s", i, expected.typ, tok.Type)
		}
		if tok.Literal != expected.lit {
			t.Errorf("token[%d] literal: expected %q, got %q", i, expected.lit, tok.Literal)
		}
	}
}
