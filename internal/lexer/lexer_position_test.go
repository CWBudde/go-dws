package lexer

import (
	"testing"
)

func TestPositionTracking(t *testing.T) {
	input := `var x
y`

	tests := []struct {
		expectedType TokenType
		expectedLine int
		expectedCol  int
	}{
		{VAR, 1, 1},
		{IDENT, 1, 5},
		{IDENT, 2, 1},
		{EOF, 2, 2},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Pos.Line != tt.expectedLine {
			t.Fatalf("tests[%d] - line wrong. expected=%d, got=%d",
				i, tt.expectedLine, tok.Pos.Line)
		}

		if tok.Pos.Column != tt.expectedCol {
			t.Fatalf("tests[%d] - column wrong. expected=%d, got=%d",
				i, tt.expectedCol, tok.Pos.Column)
		}
	}
}

func TestDebugSHR(t *testing.T) {
	input := "shl shr"
	l := New(input)

	tok1 := l.NextToken()
	t.Logf("Token 1: Type=%s, Literal=%q", tok1.Type, tok1.Literal)

	tok2 := l.NextToken()
	t.Logf("Token 2: Type=%s, Literal=%q", tok2.Type, tok2.Literal)

	if tok1.Type != SHL {
		t.Errorf("Expected SHL, got %s", tok1.Type)
	}
	if tok2.Type != SHR {
		t.Errorf("Expected SHR, got %s", tok2.Type)
	}
}

func TestDebugPositions(t *testing.T) {
	l := New("shr")

	t.Logf("Initial: pos=%d, readPos=%d, ch=%q", l.position, l.readPosition, l.ch)

	// Manually test readIdentifier
	startPos := l.position
	l.readChar()
	t.Logf("After 1st readChar: pos=%d, readPos=%d, ch=%q", l.position, l.readPosition, l.ch)
	l.readChar()
	t.Logf("After 2nd readChar: pos=%d, readPos=%d, ch=%q", l.position, l.readPosition, l.ch)
	l.readChar()
	t.Logf("After 3rd readChar: pos=%d, readPos=%d, ch=%q", l.position, l.readPosition, l.ch)

	result := l.input[startPos:l.position]
	t.Logf("Identifier slice [%d:%d] = %q", startPos, l.position, result)
}
