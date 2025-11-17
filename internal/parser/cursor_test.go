package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Helper function to create a cursor from source code
func newCursorFromSource(source string) *TokenCursor {
	l := lexer.New(source)
	return NewTokenCursor(l)
}

// TestTokenCursor_NewAndCurrent tests cursor creation and Current() method
func TestTokenCursor_NewAndCurrent(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		wantType token.TokenType
		wantLit  string
	}{
		{
			name:     "simple integer",
			source:   "42",
			wantType: token.INT,
			wantLit:  "42",
		},
		{
			name:     "keyword var",
			source:   "var x: Integer;",
			wantType: token.VAR,
			wantLit:  "var",
		},
		{
			name:     "identifier",
			source:   "myVariable",
			wantType: token.IDENT,
			wantLit:  "myVariable",
		},
		{
			name:     "empty source",
			source:   "",
			wantType: token.EOF,
			wantLit:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := newCursorFromSource(tt.source)
			current := cursor.Current()

			if current.Type != tt.wantType {
				t.Errorf("Current().Type = %v, want %v", current.Type, tt.wantType)
			}
			if current.Literal != tt.wantLit {
				t.Errorf("Current().Literal = %q, want %q", current.Literal, tt.wantLit)
			}
		})
	}
}

// TestTokenCursor_Peek tests the Peek() method with various lookahead distances
func TestTokenCursor_Peek(t *testing.T) {
	source := "var x: Integer := 42;"
	cursor := newCursorFromSource(source)

	tests := []struct {
		name     string
		peekN    int
		wantType token.TokenType
		wantLit  string
	}{
		{
			name:     "peek 0 (current)",
			peekN:    0,
			wantType: token.VAR,
			wantLit:  "var",
		},
		{
			name:     "peek 1 (next)",
			peekN:    1,
			wantType: token.IDENT,
			wantLit:  "x",
		},
		{
			name:     "peek 2",
			peekN:    2,
			wantType: token.COLON,
			wantLit:  ":",
		},
		{
			name:     "peek 3",
			peekN:    3,
			wantType: token.IDENT,
			wantLit:  "Integer",
		},
		{
			name:     "peek 4",
			peekN:    4,
			wantType: token.ASSIGN,
			wantLit:  ":=",
		},
		{
			name:     "peek 5",
			peekN:    5,
			wantType: token.INT,
			wantLit:  "42",
		},
		{
			name:     "peek 6",
			peekN:    6,
			wantType: token.SEMICOLON,
			wantLit:  ";",
		},
		{
			name:     "peek 7 (EOF)",
			peekN:    7,
			wantType: token.EOF,
			wantLit:  "",
		},
		{
			name:     "peek beyond EOF",
			peekN:    100,
			wantType: token.EOF,
			wantLit:  "",
		},
		{
			name:     "peek negative (returns current)",
			peekN:    -1,
			wantType: token.VAR,
			wantLit:  "var",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			peeked := cursor.Peek(tt.peekN)

			if peeked.Type != tt.wantType {
				t.Errorf("Peek(%d).Type = %v, want %v", tt.peekN, peeked.Type, tt.wantType)
			}
			if peeked.Literal != tt.wantLit {
				t.Errorf("Peek(%d).Literal = %q, want %q", tt.peekN, peeked.Literal, tt.wantLit)
			}
		})
	}
}

// TestTokenCursor_Advance tests the Advance() method
func TestTokenCursor_Advance(t *testing.T) {
	source := "var x := 42;"
	cursor := newCursorFromSource(source)

	// Initial position: VAR
	if cursor.Current().Type != token.VAR {
		t.Fatalf("expected VAR, got %v", cursor.Current().Type)
	}

	// Advance to IDENT
	cursor = cursor.Advance()
	if cursor.Current().Type != token.IDENT {
		t.Errorf("after 1st advance: got %v, want IDENT", cursor.Current().Type)
	}

	// Advance to ASSIGN
	cursor = cursor.Advance()
	if cursor.Current().Type != token.ASSIGN {
		t.Errorf("after 2nd advance: got %v, want ASSIGN", cursor.Current().Type)
	}

	// Advance to INT
	cursor = cursor.Advance()
	if cursor.Current().Type != token.INT {
		t.Errorf("after 3rd advance: got %v, want INT", cursor.Current().Type)
	}

	// Advance to SEMICOLON
	cursor = cursor.Advance()
	if cursor.Current().Type != token.SEMICOLON {
		t.Errorf("after 4th advance: got %v, want SEMICOLON", cursor.Current().Type)
	}

	// Advance to EOF
	cursor = cursor.Advance()
	if cursor.Current().Type != token.EOF {
		t.Errorf("after 5th advance: got %v, want EOF", cursor.Current().Type)
	}

	// Advance past EOF should stay at EOF
	cursor = cursor.Advance()
	if cursor.Current().Type != token.EOF {
		t.Errorf("after advancing past EOF: got %v, want EOF", cursor.Current().Type)
	}
}

// TestTokenCursor_AdvanceN tests the AdvanceN() method
func TestTokenCursor_AdvanceN(t *testing.T) {
	source := "var x := 42;"
	cursor := newCursorFromSource(source)

	tests := []struct {
		name     string
		n        int
		wantType token.TokenType
	}{
		{
			name:     "advance 0 (no change)",
			n:        0,
			wantType: token.VAR,
		},
		{
			name:     "advance 2",
			n:        2,
			wantType: token.ASSIGN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to beginning for each test
			cursor := newCursorFromSource(source)
			cursor = cursor.AdvanceN(tt.n)

			if cursor.Current().Type != tt.wantType {
				t.Errorf("AdvanceN(%d).Type = %v, want %v", tt.n, cursor.Current().Type, tt.wantType)
			}
		})
	}

	// Test advancing past EOF
	cursor = newCursorFromSource(source)
	cursor = cursor.AdvanceN(100)
	if cursor.Current().Type != token.EOF {
		t.Errorf("AdvanceN(100) should stop at EOF, got %v", cursor.Current().Type)
	}
}

// TestTokenCursor_Immutability tests that cursor operations are immutable
func TestTokenCursor_Immutability(t *testing.T) {
	source := "var x := 42;"
	original := newCursorFromSource(source)

	if original.Current().Type != token.VAR {
		t.Fatalf("original cursor should be at VAR, got %v", original.Current().Type)
	}

	// Advance should return a new cursor
	advanced := original.Advance()

	// Original should be unchanged
	if original.Current().Type != token.VAR {
		t.Errorf("original cursor was modified, got %v", original.Current().Type)
	}

	// Advanced should be different
	if advanced.Current().Type != token.IDENT {
		t.Errorf("advanced cursor should be at IDENT, got %v", advanced.Current().Type)
	}

	// Test that peeking doesn't change cursor
	peeked := original.Peek(1)
	if peeked.Type != token.IDENT {
		t.Errorf("Peek(1) should return IDENT, got %v", peeked.Type)
	}
	if original.Current().Type != token.VAR {
		t.Errorf("Peek() modified cursor, got %v", original.Current().Type)
	}
}

// TestTokenCursor_Is tests the Is() method
func TestTokenCursor_Is(t *testing.T) {
	source := "var x: Integer;"
	cursor := newCursorFromSource(source)

	tests := []struct {
		name      string
		checkType token.TokenType
		want      bool
	}{
		{
			name:      "matches VAR",
			checkType: token.VAR,
			want:      true,
		},
		{
			name:      "doesn't match CONST",
			checkType: token.CONST,
			want:      false,
		},
		{
			name:      "doesn't match IDENT",
			checkType: token.IDENT,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cursor.Is(tt.checkType)
			if got != tt.want {
				t.Errorf("Is(%v) = %v, want %v", tt.checkType, got, tt.want)
			}
		})
	}
}

// TestTokenCursor_IsAny tests the IsAny() method
func TestTokenCursor_IsAny(t *testing.T) {
	source := "var x := 42;"
	cursor := newCursorFromSource(source)

	tests := []struct {
		name      string
		types     []token.TokenType
		wantMatch bool
		wantType  token.TokenType
	}{
		{
			name:      "matches first type",
			types:     []token.TokenType{token.VAR, token.CONST},
			wantMatch: true,
			wantType:  token.VAR,
		},
		{
			name:      "matches second type",
			types:     []token.TokenType{token.CONST, token.VAR},
			wantMatch: true,
			wantType:  token.VAR,
		},
		{
			name:      "no match",
			types:     []token.TokenType{token.CONST, token.TYPE},
			wantMatch: false,
			wantType:  token.ILLEGAL,
		},
		{
			name:      "empty list",
			types:     []token.TokenType{},
			wantMatch: false,
			wantType:  token.ILLEGAL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, gotType := cursor.IsAny(tt.types...)
			if gotMatch != tt.wantMatch {
				t.Errorf("IsAny() match = %v, want %v", gotMatch, tt.wantMatch)
			}
			if gotType != tt.wantType {
				t.Errorf("IsAny() type = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

// TestTokenCursor_PeekIs tests the PeekIs() method
func TestTokenCursor_PeekIs(t *testing.T) {
	source := "var x: Integer;"
	cursor := newCursorFromSource(source)

	tests := []struct {
		name      string
		peekN     int
		checkType token.TokenType
		want      bool
	}{
		{
			name:      "peek 0 is VAR",
			peekN:     0,
			checkType: token.VAR,
			want:      true,
		},
		{
			name:      "peek 1 is IDENT",
			peekN:     1,
			checkType: token.IDENT,
			want:      true,
		},
		{
			name:      "peek 2 is COLON",
			peekN:     2,
			checkType: token.COLON,
			want:      true,
		},
		{
			name:      "peek 1 is not VAR",
			peekN:     1,
			checkType: token.VAR,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cursor.PeekIs(tt.peekN, tt.checkType)
			if got != tt.want {
				t.Errorf("PeekIs(%d, %v) = %v, want %v", tt.peekN, tt.checkType, got, tt.want)
			}
		})
	}
}

// TestTokenCursor_PeekIsAny tests the PeekIsAny() method
func TestTokenCursor_PeekIsAny(t *testing.T) {
	source := "var x: Integer;"
	cursor := newCursorFromSource(source)

	tests := []struct {
		name      string
		peekN     int
		types     []token.TokenType
		wantMatch bool
		wantType  token.TokenType
	}{
		{
			name:      "peek 1 matches IDENT",
			peekN:     1,
			types:     []token.TokenType{token.IDENT, token.INT},
			wantMatch: true,
			wantType:  token.IDENT,
		},
		{
			name:      "peek 2 matches COLON",
			peekN:     2,
			types:     []token.TokenType{token.SEMICOLON, token.COLON},
			wantMatch: true,
			wantType:  token.COLON,
		},
		{
			name:      "peek 1 no match",
			peekN:     1,
			types:     []token.TokenType{token.INT, token.FLOAT},
			wantMatch: false,
			wantType:  token.ILLEGAL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, gotType := cursor.PeekIsAny(tt.peekN, tt.types...)
			if gotMatch != tt.wantMatch {
				t.Errorf("PeekIsAny(%d, ...) match = %v, want %v", tt.peekN, gotMatch, tt.wantMatch)
			}
			if gotType != tt.wantType {
				t.Errorf("PeekIsAny(%d, ...) type = %v, want %v", tt.peekN, gotType, tt.wantType)
			}
		})
	}
}

// TestTokenCursor_Skip tests the Skip() method
func TestTokenCursor_Skip(t *testing.T) {
	source := "var x := 42;"

	tests := []struct {
		name          string
		skipType      token.TokenType
		wantSkipped   bool
		wantCursorPos int // Expected index after skip (or no skip)
	}{
		{
			name:          "skip VAR (matches)",
			skipType:      token.VAR,
			wantSkipped:   true,
			wantCursorPos: 1, // Should advance to IDENT
		},
		{
			name:          "skip CONST (doesn't match)",
			skipType:      token.CONST,
			wantSkipped:   false,
			wantCursorPos: 0, // Should stay at VAR
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := newCursorFromSource(source)
			newCursor, skipped := cursor.Skip(tt.skipType)

			if skipped != tt.wantSkipped {
				t.Errorf("Skip(%v) skipped = %v, want %v", tt.skipType, skipped, tt.wantSkipped)
			}

			if newCursor.index != tt.wantCursorPos {
				t.Errorf("Skip(%v) cursor index = %v, want %v", tt.skipType, newCursor.index, tt.wantCursorPos)
			}
		})
	}
}

// TestTokenCursor_SkipAny tests the SkipAny() method
func TestTokenCursor_SkipAny(t *testing.T) {
	source := "var x := 42;"

	tests := []struct {
		name          string
		types         []token.TokenType
		wantSkipped   bool
		wantType      token.TokenType
		wantCursorPos int
	}{
		{
			name:          "skip VAR or CONST (VAR matches)",
			types:         []token.TokenType{token.VAR, token.CONST},
			wantSkipped:   true,
			wantType:      token.VAR,
			wantCursorPos: 1,
		},
		{
			name:          "skip CONST or VAR (VAR matches)",
			types:         []token.TokenType{token.CONST, token.VAR},
			wantSkipped:   true,
			wantType:      token.VAR,
			wantCursorPos: 1,
		},
		{
			name:          "skip CONST or TYPE (no match)",
			types:         []token.TokenType{token.CONST, token.TYPE},
			wantSkipped:   false,
			wantType:      token.ILLEGAL,
			wantCursorPos: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := newCursorFromSource(source)
			newCursor, skipped, matchedType := cursor.SkipAny(tt.types...)

			if skipped != tt.wantSkipped {
				t.Errorf("SkipAny() skipped = %v, want %v", skipped, tt.wantSkipped)
			}
			if matchedType != tt.wantType {
				t.Errorf("SkipAny() type = %v, want %v", matchedType, tt.wantType)
			}
			if newCursor.index != tt.wantCursorPos {
				t.Errorf("SkipAny() cursor index = %v, want %v", newCursor.index, tt.wantCursorPos)
			}
		})
	}
}

// TestTokenCursor_Expect tests the Expect() method
func TestTokenCursor_Expect(t *testing.T) {
	source := "var x := 42;"
	cursor := newCursorFromSource(source)

	// Expect VAR (should match)
	newCursor, ok := cursor.Expect(token.VAR)
	if !ok {
		t.Errorf("Expect(VAR) failed, should have matched")
	}
	if newCursor.Current().Type != token.IDENT {
		t.Errorf("After Expect(VAR), cursor should be at IDENT, got %v", newCursor.Current().Type)
	}

	// Original cursor should be unchanged
	if cursor.Current().Type != token.VAR {
		t.Errorf("Original cursor was modified")
	}

	// Expect CONST on original cursor (should not match)
	newCursor2, ok2 := cursor.Expect(token.CONST)
	if ok2 {
		t.Errorf("Expect(CONST) should have failed")
	}
	if newCursor2.Current().Type != token.VAR {
		t.Errorf("After failed Expect, cursor should be unchanged, got %v", newCursor2.Current().Type)
	}
}

// TestTokenCursor_ExpectAny tests the ExpectAny() method
func TestTokenCursor_ExpectAny(t *testing.T) {
	source := "var x := 42;"
	cursor := newCursorFromSource(source)

	// ExpectAny VAR or CONST (VAR should match)
	newCursor, ok, matchedType := cursor.ExpectAny(token.VAR, token.CONST)
	if !ok {
		t.Errorf("ExpectAny(VAR, CONST) failed, should have matched")
	}
	if matchedType != token.VAR {
		t.Errorf("ExpectAny matched type = %v, want VAR", matchedType)
	}
	if newCursor.Current().Type != token.IDENT {
		t.Errorf("After ExpectAny, cursor should be at IDENT, got %v", newCursor.Current().Type)
	}

	// ExpectAny CONST or TYPE (should not match)
	_, ok2, matchedType2 := cursor.ExpectAny(token.CONST, token.TYPE)
	if ok2 {
		t.Errorf("ExpectAny(CONST, TYPE) should have failed")
	}
	if matchedType2 != token.ILLEGAL {
		t.Errorf("ExpectAny failed match type = %v, want ILLEGAL", matchedType2)
	}
}

// TestTokenCursor_Mark_ResetTo tests backtracking with Mark/ResetTo
func TestTokenCursor_Mark_ResetTo(t *testing.T) {
	source := "var x := 42;"
	cursor := newCursorFromSource(source)

	// Save mark at VAR
	mark := cursor.Mark()
	if cursor.Current().Type != token.VAR {
		t.Fatalf("cursor should be at VAR")
	}

	// Advance to IDENT
	cursor = cursor.Advance()
	if cursor.Current().Type != token.IDENT {
		t.Fatalf("cursor should be at IDENT")
	}

	// Advance to ASSIGN
	cursor = cursor.Advance()
	if cursor.Current().Type != token.ASSIGN {
		t.Fatalf("cursor should be at ASSIGN")
	}

	// Reset to mark (back to VAR)
	cursor = cursor.ResetTo(mark)
	if cursor.Current().Type != token.VAR {
		t.Errorf("after ResetTo, cursor should be at VAR, got %v", cursor.Current().Type)
	}
}

// TestTokenCursor_Mark_Multiple tests multiple marks and resets
func TestTokenCursor_Mark_Multiple(t *testing.T) {
	source := "var x := 42;"
	cursor := newCursorFromSource(source)

	// Mark 1 at VAR
	mark1 := cursor.Mark()

	// Advance to IDENT
	cursor = cursor.Advance()
	mark2 := cursor.Mark()

	// Advance to ASSIGN
	cursor = cursor.Advance()
	mark3 := cursor.Mark()

	// Reset to mark2 (IDENT)
	cursor = cursor.ResetTo(mark2)
	if cursor.Current().Type != token.IDENT {
		t.Errorf("ResetTo(mark2) should be at IDENT, got %v", cursor.Current().Type)
	}

	// Reset to mark1 (VAR)
	cursor = cursor.ResetTo(mark1)
	if cursor.Current().Type != token.VAR {
		t.Errorf("ResetTo(mark1) should be at VAR, got %v", cursor.Current().Type)
	}

	// Reset to mark3 (ASSIGN)
	cursor = cursor.ResetTo(mark3)
	if cursor.Current().Type != token.ASSIGN {
		t.Errorf("ResetTo(mark3) should be at ASSIGN, got %v", cursor.Current().Type)
	}
}

// TestTokenCursor_Clone tests the Clone() method
func TestTokenCursor_Clone(t *testing.T) {
	source := "var x := 42;"
	original := newCursorFromSource(source)

	// Clone the cursor
	cloned := original.Clone()

	// Both should be at VAR
	if original.Current().Type != token.VAR {
		t.Errorf("original should be at VAR, got %v", original.Current().Type)
	}
	if cloned.Current().Type != token.VAR {
		t.Errorf("cloned should be at VAR, got %v", cloned.Current().Type)
	}

	// Advance the clone
	cloned = cloned.Advance()
	if cloned.Current().Type != token.IDENT {
		t.Errorf("cloned after advance should be at IDENT, got %v", cloned.Current().Type)
	}

	// Original should be unchanged
	if original.Current().Type != token.VAR {
		t.Errorf("original should still be at VAR after cloned advance, got %v", original.Current().Type)
	}
}

// TestTokenCursor_IsEOF tests the IsEOF() method
func TestTokenCursor_IsEOF(t *testing.T) {
	// Test with non-empty source
	cursor := newCursorFromSource("var x")
	if cursor.IsEOF() {
		t.Errorf("IsEOF() at start should be false")
	}

	// Advance to EOF
	cursor = cursor.Advance() // IDENT
	cursor = cursor.Advance() // EOF
	if !cursor.IsEOF() {
		t.Errorf("IsEOF() at EOF should be true")
	}

	// Test with empty source
	emptySource := newCursorFromSource("")
	if !emptySource.IsEOF() {
		t.Errorf("IsEOF() for empty source should be true")
	}
}

// TestTokenCursor_Position tests the Position() method
func TestTokenCursor_Position(t *testing.T) {
	source := "var x := 42;"
	cursor := newCursorFromSource(source)

	pos := cursor.Position()
	if pos.Line != 1 {
		t.Errorf("Position().Line = %d, want 1", pos.Line)
	}
	if pos.Column != 1 {
		t.Errorf("Position().Column = %d, want 1", pos.Column)
	}

	// Advance and check position
	cursor = cursor.Advance()
	pos = cursor.Position()
	if pos.Column != 5 { // "var " is 4 chars, next token starts at column 5
		t.Errorf("Position().Column after advance = %d, want 5", pos.Column)
	}
}

// TestTokenCursor_Length tests the Length() method
func TestTokenCursor_Length(t *testing.T) {
	source := "var := 42"
	cursor := newCursorFromSource(source)

	// VAR has length 3
	if cursor.Length() != 3 {
		t.Errorf("Length() for 'var' = %d, want 3", cursor.Length())
	}

	// Advance to ASSIGN (:=)
	cursor = cursor.Advance()
	if cursor.Length() != 2 {
		t.Errorf("Length() for ':=' = %d, want 2", cursor.Length())
	}

	// Advance to INT (42)
	cursor = cursor.Advance()
	if cursor.Length() != 2 {
		t.Errorf("Length() for '42' = %d, want 2", cursor.Length())
	}
}

// TestTokenCursor_ComplexNavigation tests a complex navigation scenario
func TestTokenCursor_ComplexNavigation(t *testing.T) {
	source := `
		var x: Integer := 42;
		var y: String := "hello";
	`
	cursor := newCursorFromSource(source)

	// Start at VAR
	if !cursor.Is(token.VAR) {
		t.Fatalf("should start at VAR, got %v", cursor.Current().Type)
	}

	// Save mark
	startMark := cursor.Mark()

	// Navigate to first assignment value (42)
	cursor = cursor.Advance() // IDENT (x)
	cursor = cursor.Advance() // COLON
	cursor = cursor.Advance() // IDENT (Integer)
	cursor = cursor.Advance() // ASSIGN
	cursor = cursor.Advance() // INT (42)

	if !cursor.Is(token.INT) {
		t.Errorf("should be at INT, got %v", cursor.Current().Type)
	}
	if cursor.Current().Literal != "42" {
		t.Errorf("INT literal should be '42', got %q", cursor.Current().Literal)
	}

	// Navigate to second var statement
	cursor = cursor.Advance() // SEMICOLON
	cursor = cursor.Advance() // VAR (second)

	if !cursor.Is(token.VAR) {
		t.Errorf("should be at second VAR, got %v", cursor.Current().Type)
	}

	// Peek ahead to check string value
	stringTok := cursor.Peek(5) // VAR -> y -> COLON -> String -> ASSIGN -> STRING
	if stringTok.Type != token.STRING {
		t.Errorf("Peek(5) should be STRING, got %v", stringTok.Type)
	}
	if stringTok.Literal != "hello" {
		t.Errorf("STRING literal should be 'hello', got %q", stringTok.Literal)
	}

	// Reset to start
	cursor = cursor.ResetTo(startMark)
	if !cursor.Is(token.VAR) {
		t.Errorf("after reset, should be at first VAR, got %v", cursor.Current().Type)
	}
}

// TestTokenCursor_BufferSharing tests that multiple cursors can share the token buffer
func TestTokenCursor_BufferSharing(t *testing.T) {
	source := "var x := 42;"
	cursor1 := newCursorFromSource(source)

	// Advance cursor1 and peek ahead to populate buffer
	cursor1.Peek(5)

	// Create another cursor from same position
	cursor2 := cursor1.Clone()

	// Both should see the same tokens
	if cursor1.Current().Type != cursor2.Current().Type {
		t.Errorf("cursors should see same current token")
	}

	// Advance cursor2
	cursor2 = cursor2.Advance()

	// cursor1 should be unchanged
	if cursor1.Current().Type != token.VAR {
		t.Errorf("cursor1 should still be at VAR")
	}

	// cursor2 should be advanced
	if cursor2.Current().Type != token.IDENT {
		t.Errorf("cursor2 should be at IDENT")
	}
}
