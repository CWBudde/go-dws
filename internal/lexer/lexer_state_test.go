package lexer

import (
	"fmt"
	"testing"
)

// TestPeekCharDoesNotModifyState tests that peekChar() doesn't modify lexer state
func TestPeekCharDoesNotModifyState(t *testing.T) {
	input := "abc"
	l := New(input)

	// Current character should be 'a'
	if l.ch != 'a' {
		t.Fatalf("expected current char 'a', got %c", l.ch)
	}

	// Peek should return 'b'
	peeked := l.peekChar()
	if peeked != 'b' {
		t.Fatalf("peekChar() expected 'b', got %c", peeked)
	}

	// Current character should still be 'a'
	if l.ch != 'a' {
		t.Fatalf("after peekChar(), expected current char 'a', got %c", l.ch)
	}

	// Position should not have changed
	if l.position != 0 {
		t.Fatalf("after peekChar(), expected position 0, got %d", l.position)
	}
}

// TestPeekCharN tests peekCharN() for various N values
func TestPeekCharN(t *testing.T) {
	input := "abcdef"
	l := New(input)

	tests := []struct {
		n        int
		expected rune
	}{
		{1, 'b'}, // peek 1 ahead
		{2, 'c'}, // peek 2 ahead
		{3, 'd'}, // peek 3 ahead
		{4, 'e'}, // peek 4 ahead
		{5, 'f'}, // peek 5 ahead
		{6, 0},   // peek beyond EOF
		{10, 0},  // peek way beyond EOF
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("peek_%d", tt.n), func(t *testing.T) {
			peeked := l.peekCharN(tt.n)
			if peeked != tt.expected {
				t.Errorf("peekCharN(%d) expected %q, got %q", tt.n, tt.expected, peeked)
			}

			// Current character should still be 'a'
			if l.ch != 'a' {
				t.Errorf("after peekCharN(%d), expected current char 'a', got %c", tt.n, l.ch)
			}

			// Position should not have changed
			if l.position != 0 {
				t.Errorf("after peekCharN(%d), expected position 0, got %d", tt.n, l.position)
			}
		})
	}
}

// TestPeekCharNWithUTF8 tests peekCharN() with multi-byte UTF-8 characters
func TestPeekCharNWithUTF8(t *testing.T) {
	input := "aβγδ" // 'a' (1 byte), 'β' (2 bytes), 'γ' (2 bytes), 'δ' (2 bytes)
	l := New(input)

	tests := []struct {
		n        int
		expected rune
	}{
		{1, 'β'},
		{2, 'γ'},
		{3, 'δ'},
		{4, 0}, // beyond EOF
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("peek_%d", tt.n), func(t *testing.T) {
			peeked := l.peekCharN(tt.n)
			if peeked != tt.expected {
				t.Errorf("peekCharN(%d) expected %q, got %q", tt.n, tt.expected, peeked)
			}

			// Current character should still be 'a'
			if l.ch != 'a' {
				t.Errorf("after peekCharN(%d), expected current char 'a', got %c", tt.n, l.ch)
			}
		})
	}
}

// TestSaveRestoreStateSymmetry tests that saveState/restoreState is symmetric
func TestSaveRestoreStateSymmetry(t *testing.T) {
	input := "var x := 5;\ny := 10;"
	l := New(input)

	// Save initial state
	state1 := l.SaveState()

	// Advance lexer by reading some tokens
	l.NextToken() // var
	l.NextToken() // x
	l.NextToken() // :=

	// Save state after tokens
	state2 := l.SaveState()

	// Advance more
	l.NextToken() // 5
	l.NextToken() // ;

	// Check we're at a different position
	if l.position == state2.position {
		t.Fatal("lexer should have advanced")
	}

	// Restore to state2
	l.RestoreState(state2)

	// Verify state matches state2
	if l.position != state2.position {
		t.Errorf("position mismatch: expected %d, got %d", state2.position, l.position)
	}
	if l.readPosition != state2.readPosition {
		t.Errorf("readPosition mismatch: expected %d, got %d", state2.readPosition, l.readPosition)
	}
	if l.ch != state2.ch {
		t.Errorf("ch mismatch: expected %c, got %c", state2.ch, l.ch)
	}
	if l.line != state2.line {
		t.Errorf("line mismatch: expected %d, got %d", state2.line, l.line)
	}
	if l.column != state2.column {
		t.Errorf("column mismatch: expected %d, got %d", state2.column, l.column)
	}

	// Next token should be '5' again
	tok := l.NextToken()
	if tok.Type != INT || tok.Literal != "5" {
		t.Errorf("after restore, expected INT(5), got %s(%s)", tok.Type, tok.Literal)
	}

	// Restore to initial state
	l.RestoreState(state1)

	// Next token should be 'var' again
	tok = l.NextToken()
	if tok.Type != VAR || tok.Literal != "var" {
		t.Errorf("after restore to initial, expected VAR(var), got %s(%s)", tok.Type, tok.Literal)
	}
}

// TestSaveRestoreStatePreservesLineColumn tests that line/column are correctly saved/restored
func TestSaveRestoreStatePreservesLineColumn(t *testing.T) {
	input := "var x := 5;\ny := 10;\nz := 15;"
	l := New(input)

	// Read some tokens to advance through the input
	l.NextToken() // var
	l.NextToken() // x
	l.NextToken() // :=
	l.NextToken() // 5
	l.NextToken() // ;

	// Should be on line 2 now (after the newline)
	line1 := l.line
	col1 := l.column

	// Save state
	state := l.SaveState()

	// Read more tokens
	l.NextToken() // y
	l.NextToken() // :=
	l.NextToken() // 10
	l.NextToken() // ;

	// Should have advanced
	if l.line == line1 && l.column == col1 {
		t.Fatal("lexer should have advanced")
	}

	line2 := l.line
	col2 := l.column

	// Restore to saved state
	l.RestoreState(state)

	// Should be back at the saved position
	if l.line != state.line {
		t.Errorf("after restore, line mismatch: expected %d, got %d", state.line, l.line)
	}
	if l.column != state.column {
		t.Errorf("after restore, column mismatch: expected %d, got %d", state.column, l.column)
	}
	if l.line == line2 || l.column == col2 {
		t.Error("after restore, should not be at the advanced position")
	}
}

// TestMatchAndConsume tests the matchAndConsume helper method
func TestMatchAndConsume(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    rune
		shouldMatch bool
		expectedPos int  // position after matchAndConsume
		expectedCh  rune // current char after matchAndConsume
	}{
		{
			name:        "match succeeds",
			input:       "++",
			expected:    '+',
			shouldMatch: true,
			expectedPos: 1,
			expectedCh:  '+',
		},
		{
			name:        "match fails",
			input:       "+=",
			expected:    '+',
			shouldMatch: false,
			expectedPos: 0,
			expectedCh:  '+',
		},
		{
			name:        "match at end of input",
			input:       "+",
			expected:    0, // EOF
			shouldMatch: true,
			expectedPos: 1,
			expectedCh:  0,
		},
		{
			name:        "no match at end of input",
			input:       "+",
			expected:    'x',
			shouldMatch: false,
			expectedPos: 0,
			expectedCh:  '+',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			result := l.matchAndConsume(tt.expected)

			if result != tt.shouldMatch {
				t.Errorf("matchAndConsume(%q) = %v, expected %v", tt.expected, result, tt.shouldMatch)
			}

			if l.position != tt.expectedPos {
				t.Errorf("after matchAndConsume, position = %d, expected %d", l.position, tt.expectedPos)
			}

			if l.ch != tt.expectedCh {
				t.Errorf("after matchAndConsume, ch = %q, expected %q", l.ch, tt.expectedCh)
			}
		})
	}
}

// TestSaveRestoreStateWithPeekPreservesTokenBuffer tests that tokenBuffer is correctly saved/restored
// This is critical for parser backtracking when Peek() is used during speculative parsing.
// Bug: Before the fix, tokenBuffer was not included in LexerState, causing token duplication/skipping.
func TestSaveRestoreStateWithPeekPreservesTokenBuffer(t *testing.T) {
	input := "var x := 5 + 10;"
	l := New(input)

	// Save initial state (tokenBuffer is empty)
	state := l.SaveState()

	// Use Peek to fill the tokenBuffer
	tok0 := l.Peek(0) // var
	tok1 := l.Peek(1) // x
	tok2 := l.Peek(2) // :=

	// Verify peeked tokens
	if tok0.Type != VAR || tok0.Literal != "var" {
		t.Errorf("Peek(0) expected VAR(var), got %s(%s)", tok0.Type, tok0.Literal)
	}
	if tok1.Type != IDENT || tok1.Literal != "x" {
		t.Errorf("Peek(1) expected IDENT(x), got %s(%s)", tok1.Type, tok1.Literal)
	}
	if tok2.Type != ASSIGN || tok2.Literal != ":=" {
		t.Errorf("Peek(2) expected ASSIGN(:=), got %s(%s)", tok2.Type, tok2.Literal)
	}

	// At this point, tokenBuffer has 3 tokens: [var, x, :=]
	// Internal lexer position has advanced past these tokens

	// Restore to initial state
	// BUG: If tokenBuffer is not restored, it still contains [var, x, :=]
	// but position is back at start, causing tokens to be consumed from stale buffer
	l.RestoreState(state)

	// Now consume tokens via NextToken()
	// With the fix: tokenBuffer is empty, so tokens are generated fresh from position 0
	// Without the fix: tokenBuffer still has [var, x, :=], consuming them first

	tok := l.NextToken()
	if tok.Type != VAR || tok.Literal != "var" {
		t.Errorf("NextToken() after restore expected VAR(var), got %s(%s)", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != IDENT || tok.Literal != "x" {
		t.Errorf("NextToken() after restore expected IDENT(x), got %s(%s)", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != ASSIGN || tok.Literal != ":=" {
		t.Errorf("NextToken() after restore expected ASSIGN(:=), got %s(%s)", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != INT || tok.Literal != "5" {
		t.Errorf("NextToken() after restore expected INT(5), got %s(%s)", tok.Type, tok.Literal)
	}
}

// TestSaveRestoreStateWithPeekInMiddle tests save/restore with Peek() at a non-initial position
// This simulates the actual parser backtracking scenario in parseIsExpression.
func TestSaveRestoreStateWithPeekInMiddle(t *testing.T) {
	input := "if obj is function(x: Integer): String then end"
	l := New(input)

	// Advance to 'is' keyword
	l.NextToken() // if
	l.NextToken() // obj
	l.NextToken() // is

	// Save state at 'is' position
	state := l.SaveState()

	// Simulate parser's speculative parsing using Peek
	// This is what detectFunctionPointerFullSyntax does
	peeked := []Token{
		l.Peek(0), // function
		l.Peek(1), // (
		l.Peek(2), // x
		l.Peek(3), // :
		l.Peek(4), // Integer
	}

	// Verify peeked tokens
	expectedTypes := []TokenType{FUNCTION, LPAREN, IDENT, COLON, IDENT}
	for i, expected := range expectedTypes {
		if peeked[i].Type != expected {
			t.Errorf("Peek(%d) expected %s, got %s", i, expected, peeked[i].Type)
		}
	}

	// Now restore to the saved state (back to 'is' position)
	l.RestoreState(state)

	// Consume tokens - should start from 'function' again
	tok := l.NextToken()
	if tok.Type != FUNCTION || tok.Literal != "function" {
		t.Errorf("After restore, expected FUNCTION(function), got %s(%s)", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != LPAREN || tok.Literal != "(" {
		t.Errorf("After restore, expected LPAREN((), got %s(%s)", tok.Type, tok.Literal)
	}

	// Continue verifying the sequence is correct
	tok = l.NextToken()
	if tok.Type != IDENT || tok.Literal != "x" {
		t.Errorf("After restore, expected IDENT(x), got %s(%s)", tok.Type, tok.Literal)
	}
}
