package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestDefaultConfig verifies that DefaultConfig returns expected default values.
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"UseCursor", config.UseCursor, false},
		{"AllowReservedKeywordsAsIdentifiers", config.AllowReservedKeywordsAsIdentifiers, true},
		{"StrictMode", config.StrictMode, false},
		{"MaxRecursionDepth", config.MaxRecursionDepth, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("DefaultConfig().%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestParserBuilderWithCursorMode tests the WithCursorMode builder method.
func TestParserBuilderWithCursorMode(t *testing.T) {
	input := "var x: Integer := 42;"
	l := lexer.New(input)

	// Test enabling cursor mode
	p := NewParserBuilder(l).WithCursorMode(true).Build()

	if !p.useCursor {
		t.Errorf("WithCursorMode(true) did not enable cursor mode")
	}
	if p.cursor == nil {
		t.Errorf("WithCursorMode(true) did not initialize cursor")
	}
	if p.prefixParseFnsCursor == nil {
		t.Errorf("WithCursorMode(true) did not initialize cursor prefix parse functions")
	}
	if p.infixParseFnsCursor == nil {
		t.Errorf("WithCursorMode(true) did not initialize cursor infix parse functions")
	}

	// Test disabling cursor mode
	l2 := lexer.New(input)
	p2 := NewParserBuilder(l2).WithCursorMode(false).Build()

	if p2.useCursor {
		t.Errorf("WithCursorMode(false) did not disable cursor mode")
	}
	if p2.cursor != nil {
		t.Errorf("WithCursorMode(false) should not initialize cursor")
	}
}

// TestParserBuilderWithStrictMode tests the WithStrictMode builder method.
func TestParserBuilderWithStrictMode(t *testing.T) {
	input := "var x: Integer := 42;"
	l := lexer.New(input)

	builder := NewParserBuilder(l).WithStrictMode(true)

	if !builder.config.StrictMode {
		t.Errorf("WithStrictMode(true) did not enable strict mode")
	}

	l2 := lexer.New(input)
	builder2 := NewParserBuilder(l2).WithStrictMode(false)

	if builder2.config.StrictMode {
		t.Errorf("WithStrictMode(false) did not disable strict mode")
	}
}

// TestParserBuilderWithReservedKeywordsAsIdentifiers tests the WithReservedKeywordsAsIdentifiers builder method.
func TestParserBuilderWithReservedKeywordsAsIdentifiers(t *testing.T) {
	input := "var step: Integer;"
	l := lexer.New(input)

	builder := NewParserBuilder(l).WithReservedKeywordsAsIdentifiers(true)

	if !builder.config.AllowReservedKeywordsAsIdentifiers {
		t.Errorf("WithReservedKeywordsAsIdentifiers(true) did not enable reserved keywords as identifiers")
	}

	l2 := lexer.New(input)
	builder2 := NewParserBuilder(l2).WithReservedKeywordsAsIdentifiers(false)

	if builder2.config.AllowReservedKeywordsAsIdentifiers {
		t.Errorf("WithReservedKeywordsAsIdentifiers(false) did not disable reserved keywords as identifiers")
	}
}

// TestParserBuilderWithMaxRecursionDepth tests the WithMaxRecursionDepth builder method.
func TestParserBuilderWithMaxRecursionDepth(t *testing.T) {
	input := "var x: Integer := 42;"
	l := lexer.New(input)

	builder := NewParserBuilder(l).WithMaxRecursionDepth(2048)

	if builder.config.MaxRecursionDepth != 2048 {
		t.Errorf("WithMaxRecursionDepth(2048) = %d, want 2048", builder.config.MaxRecursionDepth)
	}
}

// TestParserBuilderWithConfig tests the WithConfig builder method.
func TestParserBuilderWithConfig(t *testing.T) {
	input := "var x: Integer := 42;"
	l := lexer.New(input)

	customConfig := ParserConfig{
		UseCursor:                          true,
		AllowReservedKeywordsAsIdentifiers: false,
		StrictMode:                         true,
		MaxRecursionDepth:                  512,
	}

	builder := NewParserBuilder(l).WithConfig(customConfig)

	if builder.config.UseCursor != customConfig.UseCursor {
		t.Errorf("WithConfig().UseCursor = %v, want %v", builder.config.UseCursor, customConfig.UseCursor)
	}
	if builder.config.AllowReservedKeywordsAsIdentifiers != customConfig.AllowReservedKeywordsAsIdentifiers {
		t.Errorf("WithConfig().AllowReservedKeywordsAsIdentifiers = %v, want %v",
			builder.config.AllowReservedKeywordsAsIdentifiers, customConfig.AllowReservedKeywordsAsIdentifiers)
	}
	if builder.config.StrictMode != customConfig.StrictMode {
		t.Errorf("WithConfig().StrictMode = %v, want %v", builder.config.StrictMode, customConfig.StrictMode)
	}
	if builder.config.MaxRecursionDepth != customConfig.MaxRecursionDepth {
		t.Errorf("WithConfig().MaxRecursionDepth = %v, want %v", builder.config.MaxRecursionDepth, customConfig.MaxRecursionDepth)
	}
}

// TestParserBuilderChaining tests that builder methods can be chained.
func TestParserBuilderChaining(t *testing.T) {
	input := "var x: Integer := 42;"
	l := lexer.New(input)

	// Chain multiple configuration methods
	p := NewParserBuilder(l).
		WithCursorMode(true).
		WithStrictMode(true).
		WithReservedKeywordsAsIdentifiers(false).
		WithMaxRecursionDepth(2048).
		Build()

	if !p.useCursor {
		t.Errorf("Chained WithCursorMode(true) did not enable cursor mode")
	}
	// Note: StrictMode, AllowReservedKeywordsAsIdentifiers, and MaxRecursionDepth
	// are not directly exposed on Parser struct (they're build-time config),
	// so we can't verify them here. They would be tested via their behavioral effects.
}

// TestParserBuilderMustBuild tests the MustBuild method.
func TestParserBuilderMustBuild(t *testing.T) {
	input := "var x: Integer := 42;"
	l := lexer.New(input)

	// MustBuild should succeed with valid input
	p := NewParserBuilder(l).MustBuild()

	if p == nil {
		t.Errorf("MustBuild() returned nil for valid input")
	}

	// Note: MustBuild only panics if Build() returns nil, which shouldn't happen
	// with valid lexer input. We can't easily test the panic case without mocking.
}

// TestParserBuilderTraditionalModeTokenInitialization verifies that traditional mode
// correctly initializes curToken and peekToken by reading two tokens.
func TestParserBuilderTraditionalModeTokenInitialization(t *testing.T) {
	input := "x := 1;"
	l := lexer.New(input)

	p := NewParserBuilder(l).WithCursorMode(false).Build()

	// In traditional mode, parser should be positioned at first token
	if p.curToken.Type != lexer.IDENT {
		t.Errorf("Traditional mode: curToken.Type = %v, want IDENT", p.curToken.Type)
	}
	if p.curToken.Literal != "x" {
		t.Errorf("Traditional mode: curToken.Literal = %v, want 'x'", p.curToken.Literal)
	}

	// peekToken should be the second token
	if p.peekToken.Type != lexer.ASSIGN {
		t.Errorf("Traditional mode: peekToken.Type = %v, want ASSIGN", p.peekToken.Type)
	}
}

// TestParserBuilderCursorModeTokenInitialization verifies that cursor mode
// correctly initializes curToken and peekToken without advancing past the first token.
// This test specifically addresses the bug reported in PR #202.
func TestParserBuilderCursorModeTokenInitialization(t *testing.T) {
	input := "x := 1;"
	l := lexer.New(input)

	p := NewParserBuilder(l).WithCursorMode(true).Build()

	// In cursor mode, parser should be positioned at first token (not third!)
	if p.curToken.Type != lexer.IDENT {
		t.Errorf("Cursor mode: curToken.Type = %v, want IDENT", p.curToken.Type)
	}
	if p.curToken.Literal != "x" {
		t.Errorf("Cursor mode: curToken.Literal = %v, want 'x'", p.curToken.Literal)
	}

	// peekToken should be the second token
	if p.peekToken.Type != lexer.ASSIGN {
		t.Errorf("Cursor mode: peekToken.Type = %v, want ASSIGN", p.peekToken.Type)
	}

	// Verify cursor is also at the first token
	if p.cursor.Current().Type != lexer.IDENT {
		t.Errorf("Cursor mode: cursor.Current().Type = %v, want IDENT", p.cursor.Current().Type)
	}
	if p.cursor.Current().Literal != "x" {
		t.Errorf("Cursor mode: cursor.Current().Literal = %v, want 'x'", p.cursor.Current().Literal)
	}
}

// TestParserBuilderRegistersParseFunctions verifies that the builder
// registers all expected parse functions.
func TestParserBuilderRegistersParseFunctions(t *testing.T) {
	input := "var x: Integer := 42;"
	l := lexer.New(input)

	p := NewParserBuilder(l).Build()

	// Test a sample of prefix parse functions
	prefixTokens := []lexer.TokenType{
		lexer.IDENT,
		lexer.INT,
		lexer.FLOAT,
		lexer.STRING,
		lexer.TRUE,
		lexer.FALSE,
		lexer.MINUS,
		lexer.LPAREN,
		lexer.LBRACK,
	}

	for _, tt := range prefixTokens {
		if p.prefixParseFns[tt] == nil {
			t.Errorf("Prefix parse function not registered for token type %v", tt)
		}
	}

	// Test a sample of infix parse functions
	infixTokens := []lexer.TokenType{
		lexer.PLUS,
		lexer.MINUS,
		lexer.ASTERISK,
		lexer.SLASH,
		lexer.EQ,
		lexer.NOT_EQ,
		lexer.LPAREN,
		lexer.DOT,
		lexer.LBRACK,
	}

	for _, tt := range infixTokens {
		if p.infixParseFns[tt] == nil {
			t.Errorf("Infix parse function not registered for token type %v", tt)
		}
	}
}

// TestParserBuilderCursorModeDoesNotSkipTokens is a regression test for the
// P1 bug reported in PR #202: cursor mode was skipping the first two tokens
// because Build() unconditionally called nextToken() twice even though
// NewTokenCursor already positioned the cursor at the first token.
func TestParserBuilderCursorModeDoesNotSkipTokens(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedFirst lexer.TokenType
		expectedLit   string
	}{
		{
			name:          "simple assignment",
			input:         "x := 1;",
			expectedFirst: lexer.IDENT,
			expectedLit:   "x",
		},
		{
			name:          "var declaration",
			input:         "var y: Integer;",
			expectedFirst: lexer.VAR,
			expectedLit:   "var",
		},
		{
			name:          "expression",
			input:         "3 + 5 * 2",
			expectedFirst: lexer.INT,
			expectedLit:   "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := NewParserBuilder(l).WithCursorMode(true).Build()

			if p.curToken.Type != tt.expectedFirst {
				t.Errorf("Parser started at wrong token: got %v (%s), want %v (%s)",
					p.curToken.Type, p.curToken.Literal, tt.expectedFirst, tt.expectedLit)
			}

			if p.curToken.Literal != tt.expectedLit {
				t.Errorf("Parser started at wrong token literal: got %q, want %q",
					p.curToken.Literal, tt.expectedLit)
			}
		})
	}
}

// TestNewCursorParserUsesBuilder verifies that NewCursorParser properly uses
// the builder and doesn't bypass its initialization logic.
func TestNewCursorParserUsesBuilder(t *testing.T) {
	input := "x := 42;"
	l := lexer.New(input)

	p := NewCursorParser(l)

	// Verify cursor mode is enabled
	if !p.useCursor {
		t.Errorf("NewCursorParser did not enable cursor mode")
	}

	// Verify cursor is initialized and positioned at first token
	if p.cursor == nil {
		t.Errorf("NewCursorParser did not initialize cursor")
	}

	if p.curToken.Type != lexer.IDENT {
		t.Errorf("NewCursorParser: curToken.Type = %v, want IDENT", p.curToken.Type)
	}

	if p.curToken.Literal != "x" {
		t.Errorf("NewCursorParser: curToken.Literal = %v, want 'x'", p.curToken.Literal)
	}

	// Verify parse functions are registered
	if len(p.prefixParseFns) == 0 {
		t.Errorf("NewCursorParser did not register prefix parse functions")
	}

	if len(p.infixParseFns) == 0 {
		t.Errorf("NewCursorParser did not register infix parse functions")
	}

	// Verify cursor-specific parse functions are registered
	if len(p.prefixParseFnsCursor) == 0 {
		t.Errorf("NewCursorParser did not register cursor prefix parse functions")
	}

	if len(p.infixParseFnsCursor) == 0 {
		t.Errorf("NewCursorParser did not register cursor infix parse functions")
	}
}
