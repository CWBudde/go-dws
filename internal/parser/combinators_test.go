package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Helper function to create a parser from source code
func newTestParser(input string) *Parser {
	l := lexer.New(input)
	return New(l)
}

// TestOptional tests the Optional combinator
func TestOptional(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		tokenType lexer.TokenType
		expected  bool
	}{
		{
			name:      "present semicolon",
			input:     "foo ;",
			tokenType: lexer.SEMICOLON,
			expected:  true,
		},
		{
			name:      "missing semicolon",
			input:     "foo bar",
			tokenType: lexer.SEMICOLON,
			expected:  false,
		},
		{
			name:      "present comma",
			input:     "foo ,",
			tokenType: lexer.COMMA,
			expected:  true,
		},
		{
			name:      "wrong token",
			input:     "foo ;",
			tokenType: lexer.COMMA,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.Optional(tt.tokenType)
			if result != tt.expected {
				t.Errorf("Optional() = %v, want %v", result, tt.expected)
			}
			// Verify token was consumed if match
			if result && tt.expected {
				if p.cursor.Current().Type == tt.tokenType {
					// Correctly consumed
				} else {
					t.Errorf("Token not consumed: curToken = %v", p.cursor.Current().Type)
				}
			}
		})
	}
}

// TestOptionalOneOf tests the OptionalOneOf combinator
func TestOptionalOneOf(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		tokenTypes []lexer.TokenType
		expected   lexer.TokenType
	}{
		{
			name:       "first option matches",
			input:      "foo +",
			tokenTypes: []lexer.TokenType{lexer.PLUS, lexer.MINUS},
			expected:   lexer.PLUS,
		},
		{
			name:       "second option matches",
			input:      "foo -",
			tokenTypes: []lexer.TokenType{lexer.PLUS, lexer.MINUS},
			expected:   lexer.MINUS,
		},
		{
			name:       "no match",
			input:      "foo *",
			tokenTypes: []lexer.TokenType{lexer.PLUS, lexer.MINUS},
			expected:   lexer.ILLEGAL,
		},
		{
			name:       "visibility specifier",
			input:      "var public",
			tokenTypes: []lexer.TokenType{lexer.PUBLIC, lexer.PRIVATE, lexer.PROTECTED},
			expected:   lexer.PUBLIC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.OptionalOneOf(tt.tokenTypes...)
			if result != tt.expected {
				t.Errorf("OptionalOneOf() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestMany tests the Many combinator
func TestMany(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		desc     string
	}{
		{
			name:     "zero matches",
			input:    "foo bar",
			expected: 0,
			desc:     "no integers",
		},
		{
			name:     "three matches",
			input:    "begin 1 2 3",
			expected: 3,
			desc:     "three integers",
		},
		{
			name:     "one match",
			input:    "begin 42",
			expected: 1,
			desc:     "one integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			count := p.Many(func() bool {
				if p.cursor.PeekIs(1, lexer.INT) {
					p.nextToken()
					return true
				}
				return false
			})
			if count != tt.expected {
				t.Errorf("Many() = %d, want %d (%s)", count, tt.expected, tt.desc)
			}
		})
	}
}

// TestMany1 tests the Many1 combinator
func TestMany1(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		desc     string
	}{
		{
			name:     "no match returns 0",
			input:    "foo bar",
			expected: 0,
			desc:     "no integers",
		},
		{
			name:     "one match",
			input:    "begin 42",
			expected: 1,
			desc:     "one integer",
		},
		{
			name:     "multiple matches",
			input:    "begin 1 2 3",
			expected: 3,
			desc:     "three integers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			count := p.Many1(func() bool {
				if p.cursor.PeekIs(1, lexer.INT) {
					p.nextToken()
					return true
				}
				return false
			})
			if count != tt.expected {
				t.Errorf("Many1() = %d, want %d (%s)", count, tt.expected, tt.desc)
			}
		})
	}
}

// TestManyUntil tests the ManyUntil combinator
func TestManyUntil(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		terminator lexer.TokenType
		expected   int
		desc       string
	}{
		{
			name:       "parse until semicolon",
			input:      "begin 1 2 3 ;",
			terminator: lexer.SEMICOLON,
			expected:   3,
			desc:       "three integers before semicolon",
		},
		{
			name:       "empty before terminator",
			input:      "begin ;",
			terminator: lexer.SEMICOLON,
			expected:   0,
			desc:       "immediate terminator",
		},
		{
			name:       "parse until EOF",
			input:      "begin 1 2 3",
			terminator: lexer.SEMICOLON,
			expected:   3,
			desc:       "reaches EOF",
		},
		{
			name:       "parser function returns false",
			input:      "begin foo 1 2 ;",
			terminator: lexer.SEMICOLON,
			expected:   0,
			desc:       "parser function fails immediately",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			count := p.ManyUntil(tt.terminator, func() bool {
				if p.cursor.PeekIs(1, lexer.INT) {
					p.nextToken()
					return true
				}
				return false
			})
			if count != tt.expected {
				t.Errorf("ManyUntil() = %d, want %d (%s)", count, tt.expected, tt.desc)
			}
		})
	}
}

// TestChoice tests the Choice combinator
func TestChoice(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		choices    []lexer.TokenType
		expected   bool
		matchedTok lexer.TokenType
	}{
		{
			name:       "first choice matches",
			input:      "begin +",
			choices:    []lexer.TokenType{lexer.PLUS, lexer.MINUS},
			expected:   true,
			matchedTok: lexer.PLUS,
		},
		{
			name:       "second choice matches",
			input:      "begin -",
			choices:    []lexer.TokenType{lexer.PLUS, lexer.MINUS},
			expected:   true,
			matchedTok: lexer.MINUS,
		},
		{
			name:       "no match",
			input:      "begin *",
			choices:    []lexer.TokenType{lexer.PLUS, lexer.MINUS},
			expected:   false,
			matchedTok: lexer.ILLEGAL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.Choice(tt.choices...)
			if result != tt.expected {
				t.Errorf("Choice() = %v, want %v", result, tt.expected)
			}
			if result && p.cursor.Current().Type != tt.matchedTok {
				t.Errorf("Expected token %v, got %v", tt.matchedTok, p.cursor.Current().Type)
			}
		})
	}
}

// TestSequence tests the Sequence combinator
func TestSequence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sequence []lexer.TokenType
		expected bool
	}{
		{
			name:     "single token match",
			input:    "x :=",
			sequence: []lexer.TokenType{lexer.ASSIGN},
			expected: true,
		},
		{
			name:     "no match",
			input:    "x +",
			sequence: []lexer.TokenType{lexer.ASSIGN},
			expected: false,
		},
		{
			name:     "multi-token sequence match",
			input:    "begin ( )",
			sequence: []lexer.TokenType{lexer.LPAREN, lexer.RPAREN},
			expected: true,
		},
		{
			name:     "partial sequence fails",
			input:    "begin ( foo",
			sequence: []lexer.TokenType{lexer.LPAREN, lexer.RPAREN},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.Sequence(tt.sequence...)
			if result != tt.expected {
				t.Errorf("Sequence() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestBetween tests the Between combinator
func TestBetween(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opening  lexer.TokenType
		closing  lexer.TokenType
		expected bool
	}{
		{
			name:     "valid parentheses",
			input:    "x ( 42 )",
			opening:  lexer.LPAREN,
			closing:  lexer.RPAREN,
			expected: true,
		},
		{
			name:     "missing opening",
			input:    "x 42 )",
			opening:  lexer.LPAREN,
			closing:  lexer.RPAREN,
			expected: false,
		},
		{
			name:     "missing closing",
			input:    "x ( 42",
			opening:  lexer.LPAREN,
			closing:  lexer.RPAREN,
			expected: false,
		},
		{
			name:     "valid brackets",
			input:    "x [ 42 ]",
			opening:  lexer.LBRACK,
			closing:  lexer.RBRACK,
			expected: true,
		},
		{
			name:     "nil result from parser",
			input:    "x ( foo )",
			opening:  lexer.LPAREN,
			closing:  lexer.RPAREN,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.Between(tt.opening, tt.closing, func() ast.Expression {
				// Parse a simple integer literal
				if !p.expectPeek(lexer.INT) {
					return nil
				}
				lit := &ast.IntegerLiteral{
					Value: 42,
				}
				lit.Token = p.cursor.Current()
				return lit
			})
			if (result != nil) != tt.expected {
				t.Errorf("Between() returned %v, want success=%v", result, tt.expected)
			}
		})
	}
}

// TestSeparatedList tests the SeparatedList combinator
func TestSeparatedList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		config   SeparatorConfig
		expected int
	}{
		{
			name:  "empty list allowed",
			input: "( )",
			config: SeparatorConfig{
				Sep:         lexer.COMMA,
				Term:        lexer.RPAREN,
				AllowEmpty:  true,
				RequireTerm: false,
			},
			expected: 0,
		},
		{
			name:  "single item",
			input: "( 42 )",
			config: SeparatorConfig{
				Sep:         lexer.COMMA,
				Term:        lexer.RPAREN,
				AllowEmpty:  true,
				RequireTerm: false,
			},
			expected: 1,
		},
		{
			name:  "multiple items",
			input: "( 1, 2, 3 )",
			config: SeparatorConfig{
				Sep:         lexer.COMMA,
				Term:        lexer.RPAREN,
				AllowEmpty:  true,
				RequireTerm: false,
			},
			expected: 3,
		},
		{
			name:  "trailing separator allowed",
			input: "( 1, 2, 3, )",
			config: SeparatorConfig{
				Sep:           lexer.COMMA,
				Term:          lexer.RPAREN,
				AllowEmpty:    true,
				AllowTrailing: true,
				RequireTerm:   false,
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			// Consume the opening parenthesis, curToken should now be on first item or terminator
			p.nextToken()

			var items []int
			tt.config.ParseItem = func() bool {
				if p.cursor.Is(lexer.INT) {
					items = append(items, 1)
					return true
				}
				return false
			}

			count := p.SeparatedList(tt.config)
			if count != tt.expected {
				t.Errorf("SeparatedList() = %d, want %d", count, tt.expected)
			}
			if len(items) != tt.expected {
				t.Errorf("Items parsed = %d, want %d", len(items), tt.expected)
			}
		})
	}
}

// TestGuard tests the Guard combinator
func TestGuard(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "guard passes",
			input:    "var x: Integer;",
			expected: true,
		},
		{
			name:     "guard fails",
			input:    "42;",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.Guard(
				func() bool { return p.cursor.Is(lexer.VAR) },
				func() bool {
					p.nextToken()
					return true
				},
			)
			if result != tt.expected {
				t.Errorf("Guard() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestPeekNIs tests the PeekNIs combinator
func TestPeekNIs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		n         int
		tokenType lexer.TokenType
		expected  bool
	}{
		{
			name:      "peek 1 (immediate next)",
			input:     "var x",
			n:         1,
			tokenType: lexer.IDENT,
			expected:  true,
		},
		{
			name:      "peek 2 (two ahead)",
			input:     "var x y",
			n:         2,
			tokenType: lexer.IDENT,
			expected:  true,
		},
		{
			name:      "peek wrong token",
			input:     "var x",
			n:         1,
			tokenType: lexer.CONST,
			expected:  false,
		},
		{
			name:      "peek 0 is invalid",
			input:     "var x",
			n:         0,
			tokenType: lexer.VAR,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.PeekNIs(tt.n, tt.tokenType)
			if result != tt.expected {
				t.Errorf("PeekNIs(%d, %v) = %v, want %v", tt.n, tt.tokenType, result, tt.expected)
			}
		})
	}
}

// TestSkipUntil tests the SkipUntil combinator
func TestSkipUntil(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		targets  []lexer.TokenType
		expected bool
		finalTok lexer.TokenType
	}{
		{
			name:     "find semicolon",
			input:    "var x: Integer; var y",
			targets:  []lexer.TokenType{lexer.SEMICOLON},
			expected: true,
			finalTok: lexer.SEMICOLON,
		},
		{
			name:     "reach EOF",
			input:    "var x: Integer",
			targets:  []lexer.TokenType{lexer.SEMICOLON},
			expected: false,
			finalTok: lexer.EOF,
		},
		{
			name:     "find first of multiple",
			input:    "var x: Integer; end",
			targets:  []lexer.TokenType{lexer.SEMICOLON, lexer.END},
			expected: true,
			finalTok: lexer.SEMICOLON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.SkipUntil(tt.targets...)
			if result != tt.expected {
				t.Errorf("SkipUntil() = %v, want %v", result, tt.expected)
			}
			if p.cursor.Current().Type != tt.finalTok {
				t.Errorf("Final token = %v, want %v", p.cursor.Current().Type, tt.finalTok)
			}
		})
	}
}

// TestSkipPast tests the SkipPast combinator
func TestSkipPast(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		targets  []lexer.TokenType
		expected bool
		finalTok lexer.TokenType
	}{
		{
			name:     "skip past semicolon",
			input:    "var x: Integer; var y",
			targets:  []lexer.TokenType{lexer.SEMICOLON},
			expected: true,
			finalTok: lexer.VAR,
		},
		{
			name:     "reach EOF",
			input:    "var x: Integer",
			targets:  []lexer.TokenType{lexer.SEMICOLON},
			expected: false,
			finalTok: lexer.EOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.SkipPast(tt.targets...)
			if result != tt.expected {
				t.Errorf("SkipPast() = %v, want %v", result, tt.expected)
			}
			if p.cursor.Current().Type != tt.finalTok {
				t.Errorf("Final token = %v, want %v", p.cursor.Current().Type, tt.finalTok)
			}
		})
	}
}

// TestSeparatedListMultiSep tests the SeparatedListMultiSep combinator
func TestSeparatedListMultiSep(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		separators []lexer.TokenType
		expected   int
	}{
		{
			name:       "comma separated",
			input:      "( 1, 2, 3 )",
			separators: []lexer.TokenType{lexer.COMMA, lexer.SEMICOLON},
			expected:   3,
		},
		{
			name:       "semicolon separated",
			input:      "( 1; 2; 3 )",
			separators: []lexer.TokenType{lexer.COMMA, lexer.SEMICOLON},
			expected:   3,
		},
		{
			name:       "mixed separators",
			input:      "( 1, 2; 3 )",
			separators: []lexer.TokenType{lexer.COMMA, lexer.SEMICOLON},
			expected:   3,
		},
		{
			name:       "empty list",
			input:      "( )",
			separators: []lexer.TokenType{lexer.COMMA},
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			// Consume the opening parenthesis
			p.nextToken()

			var items []int

			count := p.SeparatedListMultiSep(
				tt.separators,
				lexer.RPAREN,
				func() bool {
					if p.cursor.Is(lexer.INT) {
						items = append(items, 1)
						return true
					}
					return false
				},
				true,  // allow empty
				false, // no trailing
				false, // don't require term
			)

			if count != tt.expected {
				t.Errorf("SeparatedListMultiSep() = %d, want %d", count, tt.expected)
			}
		})
	}
}

// TestTryParse tests the TryParse combinator
func TestTryParse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldSucceed bool
	}{
		{
			name:          "successful parse",
			input:         ": Integer",
			shouldSucceed: true,
		},
		{
			name:          "failed parse",
			input:         "42",
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.TryParse(func() ast.Expression {
				if p.cursor.Is(lexer.COLON) {
					if p.expectPeek(lexer.IDENT) {
						ident := &ast.Identifier{
							Value: p.cursor.Current().Literal,
						}
						ident.Token = p.cursor.Current()
						return ident
					}
				}
				return nil
			})

			if (result != nil) != tt.shouldSucceed {
				t.Errorf("TryParse() success = %v, want %v", result != nil, tt.shouldSucceed)
			}

			// Verify errors are rolled back on failure
			if !tt.shouldSucceed && len(p.errors) > 0 {
				t.Errorf("Expected no errors after failed TryParse, got %d", len(p.errors))
			}
		})
	}
}

// TestBetweenStatement tests the BetweenStatement combinator
func TestBetweenStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldSucceed bool
	}{
		{
			name:          "successful between statement",
			input:         "x begin var x: Integer; end",  // Need something before BEGIN so it's in peek position
			shouldSucceed: true,
		},
		{
			name:          "missing opening delimiter",
			input:         "x var x: Integer; end",
			shouldSucceed: false,
		},
		{
			name:          "missing closing delimiter",
			input:         "x begin var x: Integer;",
			shouldSucceed: false,
		},
		{
			name:          "nil result from parse function",
			input:         "x begin end",
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.BetweenStatement(lexer.BEGIN, lexer.END, func() ast.Statement {
				// After BEGIN is consumed, we need to advance to the first token inside
				p.nextToken()
				// Simple parser that returns a statement if there's a VAR token
				if p.curTokenIs(lexer.VAR) {
					stmt := &ast.ExpressionStatement{
						BaseNode: ast.BaseNode{Token: p.cursor.Current()},
					}
					// Skip to END
					for !p.cursor.PeekIs(1, lexer.END) && !p.cursor.PeekIs(1, lexer.EOF) {
						p.nextToken()
					}
					return stmt
				}
				return nil
			})

			if (result != nil) != tt.shouldSucceed {
				t.Errorf("BetweenStatement() success = %v, want %v", result != nil, tt.shouldSucceed)
			}
		})
	}
}

// TestTryParseStatement tests the TryParseStatement combinator
func TestTryParseStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldSucceed bool
	}{
		{
			name:          "successful statement parse",
			input:         "var x: Integer;",
			shouldSucceed: true,
		},
		{
			name:          "failed statement parse",
			input:         "42",
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.TryParseStatement(func() ast.Statement {
				if p.cursor.Is(lexer.VAR) {
					stmt := &ast.ExpressionStatement{
						BaseNode: ast.BaseNode{Token: p.cursor.Current()},
					}
					return stmt
				}
				// Add an error to test error rollback
				p.addError("test error", "TEST")
				return nil
			})

			if (result != nil) != tt.shouldSucceed {
				t.Errorf("TryParseStatement() success = %v, want %v", result != nil, tt.shouldSucceed)
			}

			// Verify errors are rolled back on failure
			if !tt.shouldSucceed && len(p.errors) > 0 {
				t.Errorf("Expected no errors after failed TryParseStatement, got %d", len(p.errors))
			}
		})
	}
}

// TestPeek1Is tests the Peek1Is combinator
func TestPeek1Is(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		tokenType lexer.TokenType
		expected bool
	}{
		{
			name:      "match next token",
			input:     "foo bar",
			tokenType: lexer.IDENT,
			expected:  true,
		},
		{
			name:      "no match",
			input:     "foo ;",
			tokenType: lexer.IDENT,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.Peek1Is(tt.tokenType)
			if result != tt.expected {
				t.Errorf("Peek1Is() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestPeek2Is tests the Peek2Is combinator
func TestPeek2Is(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		tokenType lexer.TokenType
		expected bool
	}{
		{
			name:      "match token 2 ahead",
			input:     "foo bar baz",
			tokenType: lexer.IDENT,
			expected:  true,
		},
		{
			name:      "match IDENT at position 2",
			input:     "foo ; bar",
			tokenType: lexer.IDENT,  // peek(0) sees "bar" (IDENT)
			expected:  true,
		},
		{
			name:      "no match - expecting INT",
			input:     "foo ; bar",
			tokenType: lexer.INT,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.Peek2Is(tt.tokenType)
			if result != tt.expected {
				t.Errorf("Peek2Is() = %v, want %v (peek(0) = %s)", result, tt.expected, p.peek(0).Type)
			}
		})
	}
}

// TestPeek3Is tests the Peek3Is combinator
func TestPeek3Is(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		tokenType lexer.TokenType
		expected bool
	}{
		{
			name:      "match token 3 ahead",
			input:     "foo bar baz qux",
			tokenType: lexer.IDENT,
			expected:  true,
		},
		{
			name:      "match IDENT at position 3",
			input:     "foo bar ; qux",
			tokenType: lexer.IDENT,  // peek(1) sees "qux" (IDENT)
			expected:  true,
		},
		{
			name:      "no match - expecting INT",
			input:     "foo bar baz qux",
			tokenType: lexer.INT,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser(tt.input)
			result := p.Peek3Is(tt.tokenType)
			if result != tt.expected {
				t.Errorf("Peek3Is() = %v, want %v (peek(1) = %s)", result, tt.expected, p.peek(1).Type)
			}
		})
	}
}

// Benchmark tests to ensure combinators don't add overhead
func BenchmarkOptional(b *testing.B) {
	input := "; foo"
	for i := 0; i < b.N; i++ {
		p := newTestParser(input)
		p.Optional(lexer.SEMICOLON)
	}
}

func BenchmarkMany(b *testing.B) {
	input := "1 2 3 4 5 6 7 8 9 10"
	for i := 0; i < b.N; i++ {
		p := newTestParser(input)
		p.Many(func() bool {
			if p.cursor.PeekIs(1, lexer.INT) {
				p.nextToken()
				return true
			}
			return false
		})
	}
}

func BenchmarkChoice(b *testing.B) {
	input := "+ foo"
	for i := 0; i < b.N; i++ {
		p := newTestParser(input)
		p.Choice(lexer.PLUS, lexer.MINUS, lexer.ASTERISK, lexer.SLASH)
	}
}
