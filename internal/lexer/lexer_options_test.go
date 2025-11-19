package lexer

import (
	"testing"
)

// Helper functions for TestLexerOptions

func checkBoolField(t *testing.T, actual bool, expected bool, fieldName string) {
	t.Helper()
	if actual != expected {
		t.Errorf("%s should be %v", fieldName, expected)
	}
}

func collectTokens(l *Lexer) []TokenType {
	tokens := []TokenType{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok.Type)
		if tok.Type == EOF {
			break
		}
	}
	return tokens
}

func hasCommentToken(tokens []TokenType) bool {
	for _, tt := range tokens {
		if tt == COMMENT {
			return true
		}
	}
	return false
}

// TestLexerOptions tests the options pattern for Lexer configuration.
func TestLexerOptions(t *testing.T) {
	input := `var x := 42; // comment
{ block comment }`

	t.Run("default configuration", func(t *testing.T) {
		l := New(input)
		checkBoolField(t, l.preserveComments, false, "preserveComments")
		checkBoolField(t, l.tracing, false, "tracing")
	})

	t.Run("WithPreserveComments(true)", func(t *testing.T) {
		l := New(input, WithPreserveComments(true))
		checkBoolField(t, l.preserveComments, true, "preserveComments")

		// Verify comments are preserved
		tokens := collectTokens(l)
		if !hasCommentToken(tokens) {
			t.Error("expected COMMENT token when preserveComments is true")
		}
	})

	t.Run("WithPreserveComments(false)", func(t *testing.T) {
		l := New(input, WithPreserveComments(false))
		checkBoolField(t, l.preserveComments, false, "preserveComments")

		// Verify comments are skipped
		tokens := collectTokens(l)
		if hasCommentToken(tokens) {
			t.Error("unexpected COMMENT token when preserveComments is false")
		}
	})

	t.Run("WithTracing(true)", func(t *testing.T) {
		l := New(input, WithTracing(true))
		checkBoolField(t, l.tracing, true, "tracing")
	})

	t.Run("WithTracing(false)", func(t *testing.T) {
		l := New(input, WithTracing(false))
		checkBoolField(t, l.tracing, false, "tracing")
	})

	t.Run("multiple options", func(t *testing.T) {
		l := New(input, WithPreserveComments(true), WithTracing(true))
		checkBoolField(t, l.preserveComments, true, "preserveComments")
		checkBoolField(t, l.tracing, true, "tracing")
	})

	t.Run("backwards compatibility - no options", func(t *testing.T) {
		// Should work without any options (backwards compatible)
		l := New(input)
		tok := l.NextToken()
		if tok.Type != VAR {
			t.Errorf("expected VAR token, got %s", tok.Type)
		}
	})
}

// TestOptionsBackwardsCompatibility ensures that existing code without options still works.
func TestOptionsBackwardsCompatibility(t *testing.T) {
	input := `var x: Integer := 42;`

	// Old usage pattern (no options)
	l := New(input)

	tokens := []struct {
		lit string
		typ TokenType
	}{
		{lit: "var", typ: VAR},
		{lit: "x", typ: IDENT},
		{lit: ":", typ: COLON},
		{lit: "Integer", typ: IDENT},
		{lit: ":=", typ: ASSIGN},
		{lit: "42", typ: INT},
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
