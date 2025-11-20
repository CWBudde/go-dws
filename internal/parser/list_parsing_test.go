package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestParseSeparatedListBeforeStart_EmptyList tests parsing empty lists.
func TestParseSeparatedListBeforeStart_EmptyList(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectCount int
		allowEmpty  bool
		expectOk    bool
	}{
		{
			name:        "Empty list allowed",
			input:       "()",
			allowEmpty:  true,
			expectCount: 0,
			expectOk:    true,
		},
		{
			name:        "Empty list not allowed",
			input:       "()",
			allowEmpty:  false,
			expectCount: 0,
			expectOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l) // After New(): cur='(', peek=')'

			opts := ListParseOptions{
				Separators:             []lexer.TokenType{lexer.COMMA},
				Terminator:             lexer.RPAREN,
				AllowTrailingSeparator: true,
				AllowEmpty:             tt.allowEmpty,
				RequireTerminator:      true,
			}

			items := []string{}
			count, ok := p.parseSeparatedListBeforeStart(opts, func() bool {
				if !p.cursor.Is(lexer.IDENT) {
					return false
				}
				items = append(items, p.cursor.Current().Literal)
				return true
			})

			if count != tt.expectCount {
				t.Errorf("count = %d, want %d", count, tt.expectCount)
			}
			if ok != tt.expectOk {
				t.Errorf("ok = %v, want %v", ok, tt.expectOk)
			}
			if len(items) != tt.expectCount {
				t.Errorf("len(items) = %d, want %d", len(items), tt.expectCount)
			}
		})
	}
}

// TestParseSeparatedListBeforeStart_SingleItem tests parsing lists with one item.
func TestParseSeparatedListBeforeStart_SingleItem(t *testing.T) {
	input := "(x)"
	l := lexer.New(input)
	p := New(l) // After New(): cur='(', peek='x'

	opts := ListParseOptions{
		Separators:             []lexer.TokenType{lexer.COMMA},
		Terminator:             lexer.RPAREN,
		AllowTrailingSeparator: true,
		AllowEmpty:             true,
		RequireTerminator:      true,
	}

	items := []string{}
	count, ok := p.parseSeparatedListBeforeStart(opts, func() bool {
		if !p.cursor.Is(lexer.IDENT) {
			return false
		}
		items = append(items, p.cursor.Current().Literal)
		return true
	})

	if !ok {
		t.Fatalf("parseSeparatedListBeforeStart failed")
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if len(items) != 1 || items[0] != "x" {
		t.Errorf("items = %v, want [x]", items)
	}
	if !p.cursor.Is(lexer.RPAREN) {
		t.Errorf("curToken = %s, want RPAREN", p.cursor.Current().Type)
	}
}

// TestParseSeparatedListBeforeStart_MultipleItems tests parsing lists with multiple items.
func TestParseSeparatedListBeforeStart_MultipleItems(t *testing.T) {
	input := "(a, b, c)"
	l := lexer.New(input)
	p := New(l) // After New(): cur='(', peek='a'

	opts := ListParseOptions{
		Separators:             []lexer.TokenType{lexer.COMMA},
		Terminator:             lexer.RPAREN,
		AllowTrailingSeparator: false,
		AllowEmpty:             true,
		RequireTerminator:      true,
	}

	items := []string{}
	count, ok := p.parseSeparatedListBeforeStart(opts, func() bool {
		if !p.cursor.Is(lexer.IDENT) {
			return false
		}
		items = append(items, p.cursor.Current().Literal)
		return true
	})

	if !ok {
		t.Fatalf("parseSeparatedListBeforeStart failed")
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
	expected := []string{"a", "b", "c"}
	if len(items) != len(expected) {
		t.Fatalf("len(items) = %d, want %d", len(items), len(expected))
	}
	for i, want := range expected {
		if items[i] != want {
			t.Errorf("items[%d] = %s, want %s", i, items[i], want)
		}
	}
	if !p.cursor.Is(lexer.RPAREN) {
		t.Errorf("curToken = %s, want RPAREN", p.cursor.Current().Type)
	}
}

// TestParseSeparatedListBeforeStart_TrailingSeparator tests trailing separator handling.
func TestParseSeparatedListBeforeStart_TrailingSeparator(t *testing.T) {
	tests := []struct {
		name                   string
		input                  string
		expectCount            int
		allowTrailingSeparator bool
		expectOk               bool
	}{
		{
			name:                   "Trailing comma allowed",
			input:                  "(a, b, c,)",
			allowTrailingSeparator: true,
			expectCount:            3,
			expectOk:               true,
		},
		{
			name:                   "No trailing comma",
			input:                  "(a, b, c)",
			allowTrailingSeparator: false,
			expectCount:            3,
			expectOk:               true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l) // After New(): cur='('

			opts := ListParseOptions{
				Separators:             []lexer.TokenType{lexer.COMMA},
				Terminator:             lexer.RPAREN,
				AllowTrailingSeparator: tt.allowTrailingSeparator,
				AllowEmpty:             true,
				RequireTerminator:      true,
			}

			items := []string{}
			count, ok := p.parseSeparatedListBeforeStart(opts, func() bool {
				if !p.cursor.Is(lexer.IDENT) {
					return false
				}
				items = append(items, p.cursor.Current().Literal)
				return true
			})

			if ok != tt.expectOk {
				t.Errorf("ok = %v, want %v", ok, tt.expectOk)
			}
			if count != tt.expectCount {
				t.Errorf("count = %d, want %d", count, tt.expectCount)
			}
		})
	}
}

// TestParseSeparatedListBeforeStart_MultipleSeparators tests lists with multiple separator types.
func TestParseSeparatedListBeforeStart_MultipleSeparators(t *testing.T) {
	input := "(a; b, c; d)"
	l := lexer.New(input)
	p := New(l) // After New(): cur='('

	opts := ListParseOptions{
		Separators:             []lexer.TokenType{lexer.COMMA, lexer.SEMICOLON},
		Terminator:             lexer.RPAREN,
		AllowTrailingSeparator: false,
		AllowEmpty:             true,
		RequireTerminator:      true,
	}

	items := []string{}
	count, ok := p.parseSeparatedListBeforeStart(opts, func() bool {
		if !p.cursor.Is(lexer.IDENT) {
			return false
		}
		items = append(items, p.cursor.Current().Literal)
		return true
	})

	if !ok {
		t.Fatalf("parseSeparatedListBeforeStart failed")
	}
	if count != 4 {
		t.Errorf("count = %d, want 4", count)
	}
	expected := []string{"a", "b", "c", "d"}
	if len(items) != len(expected) {
		t.Fatalf("len(items) = %d, want %d", len(items), len(expected))
	}
	for i, want := range expected {
		if items[i] != want {
			t.Errorf("items[%d] = %s, want %s", i, items[i], want)
		}
	}
}

// TestParseSeparatedListBeforeStart_WithExpressions tests parsing actual expressions.
func TestParseSeparatedListBeforeStart_WithExpressions(t *testing.T) {
	input := "(1 + 2, x * 3, foo())"
	l := lexer.New(input)
	p := New(l) // After New(): cur='('

	opts := ListParseOptions{
		Separators:             []lexer.TokenType{lexer.COMMA},
		Terminator:             lexer.RPAREN,
		AllowTrailingSeparator: true,
		AllowEmpty:             true,
		RequireTerminator:      true,
	}

	exprs := []ast.Expression{}
	count, ok := p.parseSeparatedListBeforeStart(opts, func() bool {
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			return false
		}
		exprs = append(exprs, expr)
		return true
	})

	if !ok {
		t.Fatalf("parseSeparatedListBeforeStart failed: %v", p.Errors())
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
	if len(exprs) != 3 {
		t.Fatalf("len(exprs) = %d, want 3", len(exprs))
	}

	// Verify types
	if _, ok := exprs[0].(*ast.BinaryExpression); !ok {
		t.Errorf("exprs[0] is %T, want *ast.BinaryExpression", exprs[0])
	}
	if _, ok := exprs[1].(*ast.BinaryExpression); !ok {
		t.Errorf("exprs[1] is %T, want *ast.BinaryExpression", exprs[1])
	}
	if _, ok := exprs[2].(*ast.CallExpression); !ok {
		t.Errorf("exprs[2] is %T, want *ast.CallExpression", exprs[2])
	}
}

// TestParseSeparatedList_Direct tests the direct variant (curToken already at first item).
func TestParseSeparatedList_Direct(t *testing.T) {
	input := "a, b, c)"
	l := lexer.New(input)
	p := New(l) // After New(): cur='a', peek=','

	opts := ListParseOptions{
		Separators:             []lexer.TokenType{lexer.COMMA},
		Terminator:             lexer.RPAREN,
		AllowTrailingSeparator: false,
		AllowEmpty:             true,
		RequireTerminator:      true,
	}

	items := []string{}
	count, ok := p.parseSeparatedList(opts, func() bool {
		if !p.cursor.Is(lexer.IDENT) {
			return false
		}
		items = append(items, p.cursor.Current().Literal)
		return true
	})

	if !ok {
		t.Fatalf("parseSeparatedList failed")
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
	expected := []string{"a", "b", "c"}
	if len(items) != len(expected) {
		t.Fatalf("len(items) = %d, want %d", len(items), len(expected))
	}
	for i, want := range expected {
		if items[i] != want {
			t.Errorf("items[%d] = %s, want %s", i, items[i], want)
		}
	}
	if !p.cursor.Is(lexer.RPAREN) {
		t.Errorf("curToken = %s, want RPAREN", p.cursor.Current().Type)
	}
}

// TestParseSeparatedList_NoTerminator tests lists without requiring terminator.
func TestParseSeparatedList_NoTerminator(t *testing.T) {
	input := "a, b, c;"
	l := lexer.New(input)
	p := New(l) // After New(): cur='a'

	opts := ListParseOptions{
		Separators:             []lexer.TokenType{lexer.COMMA},
		Terminator:             lexer.SEMICOLON,
		AllowTrailingSeparator: false,
		AllowEmpty:             true,
		RequireTerminator:      false, // Don't require terminator
	}

	items := []string{}
	count, ok := p.parseSeparatedList(opts, func() bool {
		if !p.cursor.Is(lexer.IDENT) {
			return false
		}
		items = append(items, p.cursor.Current().Literal)
		return true
	})

	if !ok {
		t.Fatalf("parseSeparatedList failed")
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
	expected := []string{"a", "b", "c"}
	if len(items) != len(expected) {
		t.Fatalf("len(items) = %d, want %d", len(items), len(expected))
	}
	for i, want := range expected {
		if items[i] != want {
			t.Errorf("items[%d] = %s, want %s", i, items[i], want)
		}
	}
	// Should be positioned on last item when RequireTerminator is false
	if !p.cursor.Is(lexer.IDENT) || p.cursor.Current().Literal != "c" {
		t.Errorf("curToken = %v, want c", p.cursor.Current())
	}
	if !p.cursor.PeekIs(1, lexer.SEMICOLON) {
		t.Errorf("peekToken = %s, want SEMICOLON", p.cursor.Peek(1).Type)
	}
}

// TestParseSeparatedList_TrailingSeparatorNoTerminator tests the edge case where
// AllowTrailingSeparator=true AND RequireTerminator=false with a trailing separator.
// This is a regression test for the token position bug where curToken was left pointing
// to the separator instead of the last item.
func TestParseSeparatedList_TrailingSeparatorNoTerminator(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectCurToken  string
		expectCount     int
		expectPeekToken lexer.TokenType
		expectOk        bool
	}{
		{
			name:            "Trailing separator with optional terminator",
			input:           "a, b, c,)",
			expectCount:     3,
			expectOk:        true,
			expectCurToken:  "c", // Should be on last item, NOT on the comma
			expectPeekToken: lexer.RPAREN,
		},
		{
			name:            "No trailing separator with optional terminator",
			input:           "a, b, c)",
			expectCount:     3,
			expectOk:        true,
			expectCurToken:  "c",
			expectPeekToken: lexer.RPAREN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l) // After New(): cur='a'

			opts := ListParseOptions{
				Separators:             []lexer.TokenType{lexer.COMMA},
				Terminator:             lexer.RPAREN,
				AllowTrailingSeparator: true, // Allow trailing separator
				AllowEmpty:             true,
				RequireTerminator:      false, // Don't require terminator (caller handles it)
			}

			items := []string{}
			count, ok := p.parseSeparatedList(opts, func() bool {
				if !p.cursor.Is(lexer.IDENT) {
					return false
				}
				items = append(items, p.cursor.Current().Literal)
				return true
			})

			if ok != tt.expectOk {
				t.Errorf("ok = %v, want %v", ok, tt.expectOk)
			}
			if count != tt.expectCount {
				t.Errorf("count = %d, want %d", count, tt.expectCount)
			}

			// Critical assertion: curToken should be on the last item
			if !p.cursor.Is(lexer.IDENT) || p.cursor.Current().Literal != tt.expectCurToken {
				t.Errorf("curToken = %v (type=%s), want %s (type=IDENT)",
					p.cursor.Current().Literal, p.cursor.Current().Type, tt.expectCurToken)
			}

			// Verify peekToken is the terminator
			if !p.cursor.PeekIs(1, tt.expectPeekToken) {
				t.Errorf("peekToken = %s, want %s", p.cursor.Peek(1).Type, tt.expectPeekToken)
			}
		})
	}
}

// TestParseSeparatedList_ParseError tests that parse errors are handled correctly.
func TestParseSeparatedList_ParseError(t *testing.T) {
	input := "a, +, c)"
	l := lexer.New(input)
	p := New(l) // After New(): cur='a'

	opts := ListParseOptions{
		Separators:             []lexer.TokenType{lexer.COMMA},
		Terminator:             lexer.RPAREN,
		AllowTrailingSeparator: false,
		AllowEmpty:             true,
		RequireTerminator:      true,
	}

	items := []string{}
	count, ok := p.parseSeparatedList(opts, func() bool {
		// Only accept identifiers
		if !p.cursor.Is(lexer.IDENT) {
			return false
		}
		items = append(items, p.cursor.Current().Literal)
		return true
	})

	if ok {
		t.Errorf("parseSeparatedList should have failed but succeeded")
	}
	// Should have parsed 'a' before failing on '+'
	if count != 1 {
		t.Errorf("count = %d, want 1 (parsed before error)", count)
	}
	if len(items) != 1 || items[0] != "a" {
		t.Errorf("items = %v, want [a]", items)
	}
}

// TestPeekTokenIsSomeOf tests the helper function.
func TestPeekTokenIsSomeOf(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		types  []lexer.TokenType
		expect bool
	}{
		{
			name:   "Single match",
			input:  "x;",
			types:  []lexer.TokenType{lexer.SEMICOLON},
			expect: true,
		},
		{
			name:   "Multiple types, first matches",
			input:  "x,",
			types:  []lexer.TokenType{lexer.COMMA, lexer.SEMICOLON},
			expect: true,
		},
		{
			name:   "Multiple types, second matches",
			input:  "x;",
			types:  []lexer.TokenType{lexer.COMMA, lexer.SEMICOLON},
			expect: true,
		},
		{
			name:   "No match",
			input:  "x)",
			types:  []lexer.TokenType{lexer.COMMA, lexer.SEMICOLON},
			expect: false,
		},
		{
			name:   "Empty types list",
			input:  "x,",
			types:  []lexer.TokenType{},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l) // After New(): cur='x'

			result, _ := p.cursor.PeekIsAny(1, tt.types...)
			if result != tt.expect {
				t.Errorf("peekTokenIsSomeOf() = %v, want %v (peekToken = %s)",
					result, tt.expect, p.cursor.Peek(1).Type)
			}
		})
	}
}
