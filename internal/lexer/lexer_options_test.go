package lexer

import (
	"testing"
)

// TestLexerOptions tests the options pattern for Lexer configuration.
func TestLexerOptions(t *testing.T) {
	input := `var x := 42; // comment
{ block comment }`

	t.Run("default configuration", func(t *testing.T) {
		l := New(input)
		if l.preserveComments {
			t.Error("preserveComments should be false by default")
		}
		if l.tracing {
			t.Error("tracing should be false by default")
		}
	})

	t.Run("WithPreserveComments(true)", func(t *testing.T) {
		l := New(input, WithPreserveComments(true))
		if !l.preserveComments {
			t.Error("preserveComments should be true")
		}

		// Verify comments are preserved
		tokens := []TokenType{}
		for {
			tok := l.NextToken()
			tokens = append(tokens, tok.Type)
			if tok.Type == EOF {
				break
			}
		}

		// Should have COMMENT tokens
		hasComment := false
		for _, tt := range tokens {
			if tt == COMMENT {
				hasComment = true
				break
			}
		}
		if !hasComment {
			t.Error("expected COMMENT token when preserveComments is true")
		}
	})

	t.Run("WithPreserveComments(false)", func(t *testing.T) {
		l := New(input, WithPreserveComments(false))
		if l.preserveComments {
			t.Error("preserveComments should be false")
		}

		// Verify comments are skipped
		tokens := []TokenType{}
		for {
			tok := l.NextToken()
			tokens = append(tokens, tok.Type)
			if tok.Type == EOF {
				break
			}
		}

		// Should NOT have COMMENT tokens
		for _, tt := range tokens {
			if tt == COMMENT {
				t.Error("unexpected COMMENT token when preserveComments is false")
			}
		}
	})

	t.Run("WithTracing(true)", func(t *testing.T) {
		l := New(input, WithTracing(true))
		if !l.tracing {
			t.Error("tracing should be true")
		}
	})

	t.Run("WithTracing(false)", func(t *testing.T) {
		l := New(input, WithTracing(false))
		if l.tracing {
			t.Error("tracing should be false")
		}
	})

	t.Run("multiple options", func(t *testing.T) {
		l := New(input, WithPreserveComments(true), WithTracing(true))
		if !l.preserveComments {
			t.Error("preserveComments should be true")
		}
		if !l.tracing {
			t.Error("tracing should be true")
		}
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
