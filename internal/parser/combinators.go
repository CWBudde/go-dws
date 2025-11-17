// Package parser - combinator library for reusable parsing patterns.
//
// This file implements parser combinators, which are higher-order functions
// that encapsulate common parsing patterns and can be composed together.
// Combinators make the parser code more declarative, reusable, and easier to test.
//
// Design Philosophy:
//   - Type-safe: Use Go's type system to catch errors at compile time
//   - Composable: Combinators can be nested and combined
//   - Non-invasive: Works with existing parser methods without modification
//   - Zero overhead: Direct function calls with no reflection or runtime penalties
//
// Common Usage Patterns:
//
//  1. Optional constructs:
//     if p.Optional(lexer.SEMICOLON) {
//         // semicolon was present and consumed
//     }
//
//  2. Repeated items:
//     count := p.Many(func() bool {
//         return p.parseStatement() != nil
//     })
//
//  3. Separated lists:
//     items := p.SeparatedList(SeparatorConfig{
//         Sep: lexer.COMMA,
//         Term: lexer.RPAREN,
//         ParseItem: func() bool { return p.parseExpression(LOWEST) != nil },
//     })
//
//  4. Choice between alternatives:
//     if p.Choice(lexer.PLUS, lexer.MINUS) {
//         // matched either + or -
//     }
//
//  5. Bracketed expressions:
//     expr := p.Between(lexer.LPAREN, lexer.RPAREN, func() ast.Expression {
//         return p.parseExpression(LOWEST)
//     })

package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ParserFunc is a generic parser function that returns success/failure.
// Used for combinators that don't need to return specific values.
type ParserFunc func() bool

// ExpressionParserFunc is a parser function that returns an expression node.
type ExpressionParserFunc func() ast.Expression

// StatementParserFunc is a parser function that returns a statement node.
type StatementParserFunc func() ast.Statement

// Optional attempts to consume a token of the given type.
// If the peek token matches, it consumes it and returns true.
// Otherwise, it leaves the parser state unchanged and returns false.
//
// Example:
//
//	// Parse optional semicolon
//	p.Optional(lexer.SEMICOLON)
//
//	// Parse optional "forward" keyword in class declaration
//	isForward := p.Optional(lexer.FORWARD)
func (p *Parser) Optional(tokenType lexer.TokenType) bool {
	if p.peekTokenIs(tokenType) {
		p.nextToken()
		return true
	}
	return false
}

// OptionalOneOf attempts to consume one of the given token types.
// If the peek token matches any of them, it consumes it and returns the matched type.
// Otherwise, it returns lexer.ILLEGAL.
//
// Example:
//
//	// Parse optional visibility specifier
//	visibility := p.OptionalOneOf(lexer.PUBLIC, lexer.PRIVATE, lexer.PROTECTED)
//	if visibility != lexer.ILLEGAL {
//	    // handle visibility
//	}
func (p *Parser) OptionalOneOf(tokenTypes ...lexer.TokenType) lexer.TokenType {
	for _, tt := range tokenTypes {
		if p.peekTokenIs(tt) {
			p.nextToken()
			return tt
		}
	}
	return lexer.ILLEGAL
}

// Many repeatedly applies the parse function until it fails or returns false.
// Returns the count of successful applications (which may be 0).
// Does not report errors if the function fails on the first attempt.
//
// Example:
//
//	// Parse zero or more statements
//	count := p.Many(func() bool {
//	    stmt := p.parseStatement()
//	    if stmt != nil {
//	        statements = append(statements, stmt)
//	        return true
//	    }
//	    return false
//	})
func (p *Parser) Many(parseFn ParserFunc) int {
	count := 0
	for parseFn() {
		count++
	}
	return count
}

// Many1 applies the parse function one or more times.
// Returns the count of successful applications, or 0 if it fails on first attempt.
// This is useful when at least one item is required.
//
// Example:
//
//	// Parse one or more digits (at least one required)
//	count := p.Many1(func() bool {
//	    if p.peekTokenIs(lexer.INT) {
//	        p.nextToken()
//	        return true
//	    }
//	    return false
//	})
//	if count == 0 {
//	    p.addError("expected at least one digit")
//	    return nil
//	}
func (p *Parser) Many1(parseFn ParserFunc) int {
	if !parseFn() {
		return 0
	}
	count := 1
	for parseFn() {
		count++
	}
	return count
}

// ManyUntil repeatedly applies the parse function until a terminator token is found.
// Returns the count of successful applications.
// Does not consume the terminator token.
//
// Example:
//
//	// Parse statements until 'end' keyword
//	count := p.ManyUntil(lexer.END, func() bool {
//	    stmt := p.parseStatement()
//	    if stmt != nil {
//	        statements = append(statements, stmt)
//	        return true
//	    }
//	    return false
//	})
func (p *Parser) ManyUntil(terminator lexer.TokenType, parseFn ParserFunc) int {
	count := 0
	for !p.peekTokenIs(terminator) && !p.peekTokenIs(lexer.EOF) {
		if !parseFn() {
			break
		}
		count++
	}
	return count
}

// Choice attempts to consume one of the given token types.
// Returns true if any token matches and advances the parser.
// Returns false if none match, leaving parser state unchanged.
//
// Example:
//
//	// Match either '+' or '-'
//	if p.Choice(lexer.PLUS, lexer.MINUS) {
//	    operator := p.curToken.Literal
//	    // ... parse unary expression
//	}
func (p *Parser) Choice(tokenTypes ...lexer.TokenType) bool {
	for _, tt := range tokenTypes {
		if p.peekTokenIs(tt) {
			p.nextToken()
			return true
		}
	}
	return false
}

// Sequence attempts to match a sequence of token types in order.
// All tokens must match for success. If any token doesn't match,
// returns false and leaves the parser in its original state.
//
// Example:
//
//	// Match ":=" sequence
//	if p.Sequence(lexer.COLON, lexer.ASSIGN) {
//	    // both tokens matched and consumed
//	}
//
// Note: This is mainly useful for lookahead checks. For actual parsing,
// prefer using expectPeek() which provides better error messages.
func (p *Parser) Sequence(tokenTypes ...lexer.TokenType) bool {
	// Check if all tokens match without consuming
	for i, tt := range tokenTypes {
		var checkToken lexer.Token
		if i == 0 {
			checkToken = p.peekToken
		} else {
			checkToken = p.peek(i - 1)
		}
		if checkToken.Type != tt {
			return false
		}
	}

	// All match - consume them
	for range tokenTypes {
		p.nextToken()
	}
	return true
}

// Between parses a construct that is surrounded by opening and closing delimiters.
// It expects to find the opening delimiter as the peek token, then parses the content,
// then expects the closing delimiter.
//
// Returns the result of the content parser, or nil if delimiters don't match.
//
// Example:
//
//	// Parse parenthesized expression: (expr)
//	expr := p.Between(lexer.LPAREN, lexer.RPAREN, func() ast.Expression {
//	    return p.parseExpression(LOWEST)
//	})
func (p *Parser) Between(opening, closing lexer.TokenType, parseFn ExpressionParserFunc) ast.Expression {
	if !p.expectPeek(opening) {
		return nil
	}

	result := parseFn()
	if result == nil {
		return nil
	}

	if !p.expectPeek(closing) {
		return nil
	}

	return result
}

// BetweenStatement is like Between but for statements.
//
// Example:
//
//	// Parse begin...end block
//	stmt := p.BetweenStatement(lexer.BEGIN, lexer.END, func() ast.Statement {
//	    return p.parseBlockStatement()
//	})
func (p *Parser) BetweenStatement(opening, closing lexer.TokenType, parseFn StatementParserFunc) ast.Statement {
	if !p.expectPeek(opening) {
		return nil
	}

	result := parseFn()
	if result == nil {
		return nil
	}

	if !p.expectPeek(closing) {
		return nil
	}

	return result
}

// SeparatorConfig configures the SeparatedList combinator.
type SeparatorConfig struct {
	// Sep is the separator token (e.g., COMMA)
	Sep lexer.TokenType

	// Term is the terminator token that ends the list (e.g., RPAREN)
	Term lexer.TokenType

	// ParseItem is called for each list item. Should return true if successful.
	ParseItem ParserFunc

	// AllowEmpty permits empty lists (when current token is the terminator)
	AllowEmpty bool

	// AllowTrailing permits a trailing separator before the terminator: (1, 2, 3,)
	AllowTrailing bool

	// RequireTerm controls whether the terminator is required at the end
	RequireTerm bool
}

// SeparatedList parses a list of items separated by a delimiter.
// This is a higher-level combinator that wraps the parser's parseSeparatedList method.
//
// Returns the count of items parsed, or -1 if parsing failed.
//
// Example:
//
//	// Parse parameter list: (a, b, c)
//	var params []*ast.Parameter
//	count := p.SeparatedList(SeparatorConfig{
//	    Sep: lexer.COMMA,
//	    Term: lexer.RPAREN,
//	    ParseItem: func() bool {
//	        param := p.parseParameter()
//	        if param != nil {
//	            params = append(params, param)
//	            return true
//	        }
//	        return false
//	    },
//	    AllowEmpty: true,
//	    RequireTerm: true,
//	})
func (p *Parser) SeparatedList(config SeparatorConfig) int {
	opts := ListParseOptions{
		Separators:             []lexer.TokenType{config.Sep},
		Terminator:             config.Term,
		AllowTrailingSeparator: config.AllowTrailing,
		AllowEmpty:             config.AllowEmpty,
		RequireTerminator:      config.RequireTerm,
	}

	count, success := p.parseSeparatedList(opts, config.ParseItem)
	if !success {
		return -1
	}
	return count
}

// SeparatedListMultiSep is like SeparatedList but allows multiple separator types.
// Useful for parsing lists that can be separated by either comma or semicolon.
//
// Example:
//
//	// Parse record fields: either comma or semicolon separated
//	count := p.SeparatedListMultiSep(
//	    []lexer.TokenType{lexer.COMMA, lexer.SEMICOLON},
//	    lexer.END,
//	    func() bool { return p.parseFieldDecl() != nil },
//	    true,  // allow empty
//	    false, // no trailing separator
//	    true,  // require terminator
//	)
func (p *Parser) SeparatedListMultiSep(
	separators []lexer.TokenType,
	terminator lexer.TokenType,
	parseItem ParserFunc,
	allowEmpty bool,
	allowTrailing bool,
	requireTerm bool,
) int {
	opts := ListParseOptions{
		Separators:             separators,
		Terminator:             terminator,
		AllowTrailingSeparator: allowTrailing,
		AllowEmpty:             allowEmpty,
		RequireTerminator:      requireTerm,
	}

	count, success := p.parseSeparatedList(opts, parseItem)
	if !success {
		return -1
	}
	return count
}

// Guard applies a lookahead check before attempting to parse.
// If the guard condition fails, returns false without consuming tokens.
// If the guard succeeds, applies the parse function.
//
// Example:
//
//	// Only parse if next token is an identifier
//	success := p.Guard(
//	    func() bool { return p.peekTokenIs(lexer.IDENT) },
//	    func() bool { return p.parseVarDecl() != nil },
//	)
func (p *Parser) Guard(guardFn func() bool, parseFn ParserFunc) bool {
	if !guardFn() {
		return false
	}
	return parseFn()
}

// TryParse attempts to parse using the given function.
// If parsing fails, it saves the error count and returns nil.
// This is useful for optional constructs where failure is acceptable.
//
// Note: This does NOT rollback token position on failure.
// Use only when the parse function doesn't consume tokens on failure.
//
// Example:
//
//	// Try to parse a type annotation (optional in some contexts)
//	typeAnnotation := p.TryParse(func() ast.Expression {
//	    if p.peekTokenIs(lexer.COLON) {
//	        p.nextToken()
//	        return p.parseTypeExpression()
//	    }
//	    return nil
//	})
func (p *Parser) TryParse(parseFn ExpressionParserFunc) ast.Expression {
	errorCount := len(p.errors)
	result := parseFn()
	if result == nil {
		// Restore error count (ignore errors from failed attempt)
		p.errors = p.errors[:errorCount]
		return nil
	}
	return result
}

// TryParseStatement is like TryParse but for statements.
func (p *Parser) TryParseStatement(parseFn StatementParserFunc) ast.Statement {
	errorCount := len(p.errors)
	result := parseFn()
	if result == nil {
		// Restore error count (ignore errors from failed attempt)
		p.errors = p.errors[:errorCount]
		return nil
	}
	return result
}

// Peek1Is checks if the next token (1 position ahead) matches the given type.
// This is a convenience wrapper around peekTokenIs.
func (p *Parser) Peek1Is(tokenType lexer.TokenType) bool {
	return p.peekTokenIs(tokenType)
}

// Peek2Is checks if the token 2 positions ahead matches the given type.
func (p *Parser) Peek2Is(tokenType lexer.TokenType) bool {
	return p.peek(0).Type == tokenType
}

// Peek3Is checks if the token 3 positions ahead matches the given type.
func (p *Parser) Peek3Is(tokenType lexer.TokenType) bool {
	return p.peek(1).Type == tokenType
}

// PeekNIs checks if the token N positions ahead matches the given type.
// N=1 checks peekToken, N=2 checks peek(0), etc.
func (p *Parser) PeekNIs(n int, tokenType lexer.TokenType) bool {
	if n <= 0 {
		return false
	}
	if n == 1 {
		return p.peekTokenIs(tokenType)
	}
	return p.peek(n-2).Type == tokenType
}

// SkipUntil advances the parser until it finds one of the given token types.
// Does not consume the found token.
// Returns true if a token was found, false if EOF was reached.
//
// This is useful for error recovery.
//
// Example:
//
//	// Skip to next semicolon or end
//	if !p.SkipUntil(lexer.SEMICOLON, lexer.END, lexer.EOF) {
//	    // reached EOF without finding target
//	}
func (p *Parser) SkipUntil(tokenTypes ...lexer.TokenType) bool {
	for !p.curTokenIs(lexer.EOF) {
		for _, tt := range tokenTypes {
			if p.curTokenIs(tt) {
				return true
			}
		}
		p.nextToken()
	}
	return false
}

// SkipPast advances the parser until it finds and consumes one of the given token types.
// Returns true if a token was found and consumed, false if EOF was reached.
//
// Example:
//
//	// Skip past the next semicolon
//	if p.SkipPast(lexer.SEMICOLON) {
//	    // semicolon was found and consumed
//	}
func (p *Parser) SkipPast(tokenTypes ...lexer.TokenType) bool {
	if p.SkipUntil(tokenTypes...) {
		p.nextToken() // consume the found token
		return true
	}
	return false
}
