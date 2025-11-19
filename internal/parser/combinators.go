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
//	// Match member access (dot followed by identifier)
//	if p.Sequence(lexer.DOT, lexer.IDENT) {
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

// ========================================================================
// High-Level DWScript-Specific Combinators
// ========================================================================
//
// These combinators encapsulate common DWScript language patterns and provide
// higher-level abstractions than the basic combinators above. They combine
// multiple basic combinators to handle domain-specific parsing tasks.

// OptionalTypeAnnotation parses an optional type annotation (: Type).
// Returns the type expression if present, nil otherwise.
// Does not consume tokens if no type annotation is present.
//
// Syntax: [: TypeExpression]
//
// PRE: curToken is the token before potential colon
// POST: curToken is last token of type expression if present; otherwise unchanged
//
// Example:
//
//	// Parse optional type annotation in variable declaration
//	typeExpr := p.OptionalTypeAnnotation()
//	if typeExpr != nil {
//	    // Type annotation was present
//	}
func (p *Parser) OptionalTypeAnnotation() ast.TypeExpression {
	if !p.peekTokenIs(lexer.COLON) {
		return nil
	}

	p.nextToken() // move to ':'
	p.nextToken() // move to type expression

	typeExpr := p.parseTypeExpression()
	if typeExpr == nil {
		// Error already reported by parseTypeExpression
		return nil
	}

	return typeExpr
}

// IdentifierListConfig configures the IdentifierList combinator.
type IdentifierListConfig struct {
	// ErrorContext provides context for error messages (e.g., "parameter declaration", "var declaration")
	ErrorContext string

	// AllowEmpty permits empty lists (returns empty slice instead of nil)
	AllowEmpty bool

	// RequireAtLeastOne ensures at least one identifier is parsed
	RequireAtLeastOne bool
}

// IdentifierList parses a comma-separated list of identifiers.
// Returns a slice of Identifier nodes, or nil if parsing fails.
//
// Syntax: IDENT (, IDENT)*
//
// PRE: curToken is first identifier (or error token if invalid)
// POST: curToken is last identifier
//
// Example:
//
//	// Parse identifier list: a, b, c
//	ids := p.IdentifierList(IdentifierListConfig{
//	    ErrorContext: "variable declaration",
//	    RequireAtLeastOne: true,
//	})
func (p *Parser) IdentifierList(config IdentifierListConfig) []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// Check first identifier
	if !p.isIdentifierToken(p.cursor.Current().Type) {
		if config.RequireAtLeastOne {
			context := config.ErrorContext
			if context == "" {
				context = "identifier list"
			}

			// Task 2.7.7: Dual-mode - get current token for error reporting
			var curTok lexer.Token
			if p.cursor != nil {
				curTok = p.cursor.Current()
			} else {
				curTok = p.curToken
			}

			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrExpectedIdent).
				WithMessage("expected identifier in "+context).
				WithPosition(curTok.Pos, curTok.Length()).
				WithExpectedString("identifier").
				WithActual(curTok.Type, curTok.Literal).
				WithParsePhase(context).
				Build()
			p.addStructuredError(err)
			return nil
		}
		if config.AllowEmpty {
			return identifiers
		}
		return nil
	}

	// Parse identifiers separated by commas
	for {
		// Task 2.7.8: Dual-mode - get current token for AST node creation
		var curTok lexer.Token
		if p.cursor != nil {
			curTok = p.cursor.Current()
		} else {
			curTok = p.curToken
		}

		identifiers = append(identifiers, &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: curTok,
				},
			},
			Value: curTok.Literal,
		})

		// Check for comma separator
		if !p.peekTokenIs(lexer.COMMA) {
			break
		}

		p.nextToken() // move to ','

		// Expect identifier after comma
		if !p.expectIdentifier() {
			// Error already reported by expectIdentifier
			return identifiers // Return what we have so far
		}
	}

	return identifiers
}

// StatementBlockConfig configures the StatementBlock combinator.
type StatementBlockConfig struct {
	// OpenToken is the token that starts the block (e.g., BEGIN, TRY)
	OpenToken lexer.TokenType

	// CloseToken is the token that ends the block (e.g., END)
	CloseToken lexer.TokenType

	// AdditionalTerminators are additional tokens that can end the block
	// (e.g., EXCEPT and FINALLY for TRY blocks)
	AdditionalTerminators []lexer.TokenType

	// SkipSemicolons controls whether to automatically skip semicolons
	SkipSemicolons bool

	// ContextName provides context for error messages (e.g., "try block", "function body")
	ContextName string

	// RequireClose controls whether the closing token is required
	RequireClose bool
}

// StatementBlock parses a block of statements with configurable delimiters.
// Returns a BlockStatement node or nil if parsing fails.
//
// PRE: curToken is the opening token (BEGIN, TRY, etc.)
// POST: curToken is the closing token (END, etc.) if RequireClose is true
//
// Example:
//
//	// Parse begin...end block
//	block := p.StatementBlock(StatementBlockConfig{
//	    OpenToken: lexer.BEGIN,
//	    CloseToken: lexer.END,
//	    SkipSemicolons: true,
//	    ContextName: "begin block",
//	    RequireClose: true,
//	})
func (p *Parser) StatementBlock(config StatementBlockConfig) *ast.BlockStatement {
	builder := p.StartNode()

	// Task 2.7.8: Dual-mode - get current token for AST node creation and context tracking
	var curTok lexer.Token
	if p.cursor != nil {
		curTok = p.cursor.Current()
	} else {
		curTok = p.curToken
	}

	block := &ast.BlockStatement{
		BaseNode:   ast.BaseNode{Token: curTok},
		Statements: []ast.Statement{},
	}

	// Track block context for better error messages
	contextName := config.ContextName
	if contextName == "" {
		contextName = config.OpenToken.String()
	}
	p.pushBlockContext(contextName, curTok.Pos)
	defer p.popBlockContext()

	p.nextToken() // advance past opening token

	// Build list of terminator tokens
	terminators := []lexer.TokenType{config.CloseToken, lexer.EOF}
	if len(config.AdditionalTerminators) > 0 {
		terminators = append([]lexer.TokenType{config.CloseToken, lexer.EOF}, config.AdditionalTerminators...)
	}

	// Parse statements until we hit a terminator
	for {
		// Check for terminators
		isTerminator := false
		for _, term := range terminators {
			if p.curTokenIs(term) {
				isTerminator = true
				break
			}
		}
		if isTerminator {
			break
		}

		// Skip semicolons if configured
		if config.SkipSemicolons && p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// Parse statement
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		p.nextToken()
	}

	// Verify we ended at the expected close token if required
	if config.RequireClose && !p.curTokenIs(config.CloseToken) {
		// Check if we hit an additional terminator instead
		hitAdditionalTerm := false
		for _, term := range config.AdditionalTerminators {
			if p.curTokenIs(term) {
				hitAdditionalTerm = true
				break
			}
		}

		if !hitAdditionalTerm {
			// Task 2.7.7: Dual-mode - get current token for error reporting
			var curTok lexer.Token
			if p.cursor != nil {
				curTok = p.cursor.Current()
			} else {
				curTok = p.curToken
			}

			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrUnexpectedToken).
				WithMessage("expected '"+config.CloseToken.String()+"' to close "+contextName).
				WithPosition(curTok.Pos, curTok.Length()).
				WithExpected(config.CloseToken).
				WithActual(curTok.Type, curTok.Literal).
				WithParsePhase(contextName).
				Build()
			p.addStructuredError(err)
		}
	}

	return builder.Finish(block).(*ast.BlockStatement)
}

// ParameterGroupConfig configures the ParameterGroup combinator.
type ParameterGroupConfig struct {
	// AllowModifiers controls whether modifiers (var, const, lazy) are allowed
	AllowModifiers bool

	// AllowDefaults controls whether default values are allowed
	AllowDefaults bool

	// ErrorContext provides context for error messages
	ErrorContext string
}

// ParameterGroup parses a parameter group with shared type.
// Returns a slice of Parameter nodes or nil if parsing fails.
//
// Syntax: [modifier] name1, name2: Type [= default]
// Where modifier is one of: var, const, lazy
//
// PRE: curToken is first token (modifier keyword or identifier)
// POST: curToken is last token of type/default expression
//
// Example:
//
//	// Parse parameter group: var a, b: Integer
//	params := p.ParameterGroup(ParameterGroupConfig{
//	    AllowModifiers: true,
//	    AllowDefaults: false,
//	    ErrorContext: "function parameter",
//	})
func (p *Parser) ParameterGroup(config ParameterGroupConfig) []*ast.Parameter {
	params := []*ast.Parameter{}

	// Parse optional modifiers
	isConst := false
	isLazy := false
	byRef := false

	if config.AllowModifiers {
		// Check for 'const' keyword (pass by const-reference)
		if p.curTokenIs(lexer.CONST) {
			isConst = true
			p.nextToken() // move past 'const'
		}

		// Check for 'lazy' keyword (expression capture)
		if p.curTokenIs(lexer.LAZY) {
			isLazy = true
			p.nextToken() // move past 'lazy'
		}

		// Check for 'var' keyword (pass by reference)
		if p.curTokenIs(lexer.VAR) {
			byRef = true
			p.nextToken() // move past 'var'
		}

		// Check for mutually exclusive modifiers
		if (isLazy && byRef) || (isConst && byRef) || (isConst && isLazy) {
			// Task 2.7.7: Dual-mode - get current token for error reporting
			var curTok lexer.Token
			if p.cursor != nil {
				curTok = p.cursor.Current()
			} else {
				curTok = p.curToken
			}

			err := NewStructuredError(ErrKindInvalid).
				WithCode(ErrInvalidSyntax).
				WithMessage("parameter modifiers are mutually exclusive").
				WithPosition(curTok.Pos, curTok.Length()).
				WithSuggestion("use only one of: var, const, or lazy").
				WithParsePhase(config.ErrorContext).
				Build()
			p.addStructuredError(err)
			return nil
		}
	}

	// Parse identifier list
	names := p.IdentifierList(IdentifierListConfig{
		ErrorContext:      config.ErrorContext,
		RequireAtLeastOne: true,
	})
	if names == nil {
		return nil
	}

	// Expect ':' and type
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken() // move past COLON to type expression start token

	typeExpr := p.parseTypeExpression()
	if typeExpr == nil {
		// Error already reported by parseTypeExpression
		return nil
	}

	// Parse optional default value
	var defaultValue ast.Expression
	if config.AllowDefaults && p.peekTokenIs(lexer.EQ) {
		// Validate that optional parameters don't have modifiers (lazy, var, const)
		if isLazy || byRef || isConst {
			// Task 2.7.7: Dual-mode - get current token for error reporting
			var curTok lexer.Token
			if p.cursor != nil {
				curTok = p.cursor.Current()
			} else {
				curTok = p.curToken
			}

			err := NewStructuredError(ErrKindInvalid).
				WithCode(ErrInvalidSyntax).
				WithMessage("optional parameters cannot have lazy, var, or const modifiers").
				WithPosition(curTok.Pos, curTok.Length()).
				WithSuggestion("remove the modifier or remove the default value").
				WithParsePhase(config.ErrorContext).
				Build()
			p.addStructuredError(err)
			return nil
		}

		p.nextToken() // move to '='
		p.nextToken() // move past '='
		defaultValue = p.parseExpression(LOWEST)
		if defaultValue == nil {
			// Task 2.7.7: Dual-mode - get current token for error reporting
			var curTok lexer.Token
			if p.cursor != nil {
				curTok = p.cursor.Current()
			} else {
				curTok = p.curToken
			}

			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrInvalidExpression).
				WithMessage("expected default value expression after '='").
				WithPosition(curTok.Pos, curTok.Length()).
				WithExpectedString("expression").
				WithParsePhase(config.ErrorContext).
				Build()
			p.addStructuredError(err)
			return nil
		}
	}

	// Create parameter nodes for each name
	for _, name := range names {
		param := &ast.Parameter{
			Token:        name.Token,
			Name:         name,
			Type:         typeExpr,
			ByRef:        byRef,
			IsConst:      isConst,
			IsLazy:       isLazy,
			DefaultValue: defaultValue,
		}
		params = append(params, param)
	}

	return params
}
