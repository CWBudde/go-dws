package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ParserConfig holds configuration options for the parser.
// This separates parser configuration from parser state, making it easier
// to create parsers with different configurations.
type ParserConfig struct {
	// UseCursor enables cursor-based parsing mode (vs. traditional mode)
	UseCursor bool

	// AllowReservedKeywordsAsIdentifiers allows using reserved keywords as identifiers
	// in contexts where they're unambiguous (e.g., 'step' as a variable name)
	AllowReservedKeywordsAsIdentifiers bool

	// StrictMode enables stricter parsing rules (future use)
	StrictMode bool

	// MaxRecursionDepth sets the maximum recursion depth for expression parsing
	// to prevent stack overflow on deeply nested expressions (future use)
	MaxRecursionDepth int
}

// DefaultConfig returns a ParserConfig with default settings.
func DefaultConfig() ParserConfig {
	return ParserConfig{
		UseCursor:                          true, // Cursor mode is now the standard (Task 2.7.4.2)
		AllowReservedKeywordsAsIdentifiers: true,
		StrictMode:                         false,
		MaxRecursionDepth:                  1000,
	}
}

// ParserBuilder provides a fluent API for constructing Parser instances.
// It reduces code duplication and makes parser configuration more explicit.
//
// Example usage:
//
//	parser := NewParserBuilder(lexer).
//	    WithCursorMode(true).
//	    WithStrictMode(false).
//	    Build()
//
// For simple cases, use the convenience constructors:
//
//	parser := New(lexer)          // Traditional mode with defaults
//	parser := NewCursorParser(lexer)  // Cursor mode with defaults
type ParserBuilder struct {
	lexer  *lexer.Lexer
	config ParserConfig
}

// NewParserBuilder creates a new ParserBuilder with default configuration.
func NewParserBuilder(l *lexer.Lexer) *ParserBuilder {
	return &ParserBuilder{
		lexer:  l,
		config: DefaultConfig(),
	}
}

// WithConfig sets the entire configuration at once.
func (b *ParserBuilder) WithConfig(config ParserConfig) *ParserBuilder {
	b.config = config
	return b
}

// WithCursorMode enables or disables cursor-based parsing.
func (b *ParserBuilder) WithCursorMode(enabled bool) *ParserBuilder {
	b.config.UseCursor = enabled
	return b
}

// WithStrictMode enables or disables strict parsing mode.
func (b *ParserBuilder) WithStrictMode(enabled bool) *ParserBuilder {
	b.config.StrictMode = enabled
	return b
}

// WithReservedKeywordsAsIdentifiers allows or disallows reserved keywords as identifiers.
func (b *ParserBuilder) WithReservedKeywordsAsIdentifiers(enabled bool) *ParserBuilder {
	b.config.AllowReservedKeywordsAsIdentifiers = enabled
	return b
}

// WithMaxRecursionDepth sets the maximum recursion depth for expression parsing.
func (b *ParserBuilder) WithMaxRecursionDepth(depth int) *ParserBuilder {
	b.config.MaxRecursionDepth = depth
	return b
}

// Build constructs and returns a configured Parser instance.
// This is the main entry point for creating parsers via the builder pattern.
func (b *ParserBuilder) Build() *Parser {
	// Create parser with basic configuration
	p := &Parser{
		l:              b.lexer,
		errors:         []*ParserError{},
		prefixParseFns: make(map[lexer.TokenType]prefixParseFn),
		infixParseFns:  make(map[lexer.TokenType]infixParseFn),
		blockStack:     []BlockContext{},
		ctx:            NewParseContext(),
		useCursor:      b.config.UseCursor,
	}

	// Always initialize both Traditional and Cursor maps during transition
	p.prefixParseFnsCursor = make(map[lexer.TokenType]prefixParseFnCursor)
	p.infixParseFnsCursor = make(map[lexer.TokenType]infixParseFnCursor)

	// Configure cursor mode
	if b.config.UseCursor {
		p.cursor = NewTokenCursor(b.lexer)
	} else {
		p.cursor = nil
	}

	// Register all parse functions
	b.registerParseFunctions(p)

	// Initialize token state
	// In cursor mode, NewTokenCursor already positioned the cursor at the first token,
	// so we just sync the traditional token pointers to match.
	// In traditional mode, we need to read the first two tokens from the lexer.
	if b.config.UseCursor {
		// Cursor is already at the first token - just sync to traditional pointers
		p.curToken = p.cursor.Current()
		p.peekToken = p.cursor.Peek(1)
	} else {
		// Traditional mode: read first two tokens
		p.nextToken()
		p.nextToken()
	}

	return p
}

// registerParseFunctions registers all prefix and infix parse functions.
// This centralizes the registration logic that was previously duplicated
// between New() and NewCursorParser().
func (b *ParserBuilder) registerParseFunctions(p *Parser) {
	// Register prefix parse functions for literals (both traditional and cursor maps)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.NIL, p.parseNilLiteral)
	p.registerPrefix(lexer.NULL, p.parseNullIdentifier)
	p.registerPrefix(lexer.UNASSIGNED, p.parseUnassignedIdentifier)
	p.registerPrefix(lexer.CHAR, p.parseCharLiteral)

	// Also register in cursor map for pure cursor mode operation
	// TODO(Task 2.7.5): These wrapper functions ignore the tok parameter and call traditional
	// implementations. They should be replaced with pure cursor-mode implementations that
	// accept and use the token parameter directly, eliminating the need for closures.
	p.registerPrefixCursor(lexer.IDENT, func(tok lexer.Token) ast.Expression { return p.parseIdentifier() })
	p.registerPrefixCursor(lexer.INT, func(tok lexer.Token) ast.Expression { return p.parseIntegerLiteral() })
	p.registerPrefixCursor(lexer.FLOAT, func(tok lexer.Token) ast.Expression { return p.parseFloatLiteral() })
	p.registerPrefixCursor(lexer.STRING, func(tok lexer.Token) ast.Expression { return p.parseStringLiteral() })
	p.registerPrefixCursor(lexer.TRUE, func(tok lexer.Token) ast.Expression { return p.parseBooleanLiteral() })
	p.registerPrefixCursor(lexer.FALSE, func(tok lexer.Token) ast.Expression { return p.parseBooleanLiteral() })
	p.registerPrefixCursor(lexer.NIL, func(tok lexer.Token) ast.Expression { return p.parseNilLiteral() })
	p.registerPrefixCursor(lexer.NULL, func(tok lexer.Token) ast.Expression { return p.parseNullIdentifier() })
	p.registerPrefixCursor(lexer.UNASSIGNED, func(tok lexer.Token) ast.Expression { return p.parseUnassignedIdentifier() })
	p.registerPrefixCursor(lexer.CHAR, func(tok lexer.Token) ast.Expression { return p.parseCharLiteral() })

	// Register prefix parse functions for operators
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.PLUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.NOT, p.parsePrefixExpression)

	// TODO(Task 2.7.5): Same issue - wrapper ignores tok parameter
	p.registerPrefixCursor(lexer.MINUS, func(tok lexer.Token) ast.Expression { return p.parsePrefixExpression() })
	p.registerPrefixCursor(lexer.PLUS, func(tok lexer.Token) ast.Expression { return p.parsePrefixExpression() })
	p.registerPrefixCursor(lexer.NOT, func(tok lexer.Token) ast.Expression { return p.parsePrefixExpression() })

	// Register prefix parse functions for grouping and collections
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.LBRACK, p.parseArrayLiteral)

	// TODO(Task 2.7.5): Same issue - wrapper ignores tok parameter
	p.registerPrefixCursor(lexer.LPAREN, func(tok lexer.Token) ast.Expression { return p.parseGroupedExpression() })
	p.registerPrefixCursor(lexer.LBRACK, func(tok lexer.Token) ast.Expression { return p.parseArrayLiteral() })

	// Register prefix parse functions for keywords
	p.registerPrefix(lexer.NEW, p.parseNewExpression)
	p.registerPrefix(lexer.DEFAULT, p.parseDefaultExpression)
	p.registerPrefix(lexer.AT, p.parseAddressOfExpression)
	p.registerPrefix(lexer.LAMBDA, p.parseLambdaExpression)
	p.registerPrefix(lexer.OLD, p.parseOldExpression)
	p.registerPrefix(lexer.INHERITED, p.parseInheritedExpression)
	p.registerPrefix(lexer.SELF, p.parseSelfExpression)
	p.registerPrefix(lexer.IF, p.parseIfExpression)

	// TODO(Task 2.7.5): Same issue - wrapper ignores tok parameter
	p.registerPrefixCursor(lexer.NEW, func(tok lexer.Token) ast.Expression { return p.parseNewExpression() })
	p.registerPrefixCursor(lexer.DEFAULT, func(tok lexer.Token) ast.Expression { return p.parseDefaultExpression() })
	p.registerPrefixCursor(lexer.AT, func(tok lexer.Token) ast.Expression { return p.parseAddressOfExpression() })
	p.registerPrefixCursor(lexer.LAMBDA, func(tok lexer.Token) ast.Expression { return p.parseLambdaExpression() })
	p.registerPrefixCursor(lexer.OLD, func(tok lexer.Token) ast.Expression { return p.parseOldExpression() })
	p.registerPrefixCursor(lexer.INHERITED, func(tok lexer.Token) ast.Expression { return p.parseInheritedExpression() })
	p.registerPrefixCursor(lexer.SELF, func(tok lexer.Token) ast.Expression { return p.parseSelfExpression() })
	p.registerPrefixCursor(lexer.IF, func(tok lexer.Token) ast.Expression { return p.parseIfExpression() })

	// Register contextual keywords that can be used as identifiers
	if b.config.AllowReservedKeywordsAsIdentifiers {
		p.registerPrefix(lexer.HELPER, p.parseIdentifier)
		p.registerPrefix(lexer.STEP, p.parseIdentifier)
		p.registerPrefixCursor(lexer.HELPER, func(tok lexer.Token) ast.Expression { return p.parseIdentifier() })
		p.registerPrefixCursor(lexer.STEP, func(tok lexer.Token) ast.Expression { return p.parseIdentifier() })
	}

	// Register infix parse functions for operators (both traditional and cursor maps)
	p.registerInfix(lexer.QUESTION_QUESTION, p.parseInfixExpression) // Coalesce: a ?? b
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.ASTERISK, p.parseInfixExpression)
	p.registerInfix(lexer.SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.DIV, p.parseInfixExpression)
	p.registerInfix(lexer.MOD, p.parseInfixExpression)
	p.registerInfix(lexer.SHL, p.parseInfixExpression)
	p.registerInfix(lexer.SHR, p.parseInfixExpression)
	p.registerInfix(lexer.SAR, p.parseInfixExpression)

	// TODO(Task 2.7.5): All infix cursor registrations below also ignore the tok parameter.
	// They should be refactored to use pure cursor-mode implementations.
	p.registerInfixCursor(lexer.QUESTION_QUESTION, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.PLUS, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.MINUS, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.ASTERISK, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.SLASH, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.DIV, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.MOD, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.SHL, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.SHR, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.SAR, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })

	// Register infix parse functions for comparison operators
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LESS, p.parseInfixExpression)
	p.registerInfix(lexer.GREATER, p.parseInfixExpression)
	p.registerInfix(lexer.LESS_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.GREATER_EQ, p.parseInfixExpression)

	p.registerInfixCursor(lexer.EQ, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.NOT_EQ, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.LESS, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.GREATER, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.LESS_EQ, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.GREATER_EQ, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })

	// Register infix parse functions for logical operators
	p.registerInfix(lexer.AND, p.parseInfixExpression)
	p.registerInfix(lexer.OR, p.parseInfixExpression)
	p.registerInfix(lexer.XOR, p.parseInfixExpression)

	p.registerInfixCursor(lexer.AND, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.OR, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.XOR, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })

	// Register infix parse functions for type operations
	p.registerInfix(lexer.IN, p.parseInfixExpression)
	p.registerInfix(lexer.IS, p.parseIsExpression)
	p.registerInfix(lexer.AS, p.parseAsExpression)
	p.registerInfix(lexer.IMPLEMENTS, p.parseImplementsExpression)

	p.registerInfixCursor(lexer.IN, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfixCursor(lexer.IS, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseIsExpression(left) })
	p.registerInfixCursor(lexer.AS, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseAsExpression(left) })
	p.registerInfixCursor(lexer.IMPLEMENTS, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseImplementsExpression(left) })

	// Register infix parse functions for member access and calls
	p.registerInfix(lexer.DOT, p.parseMemberAccess)
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.LBRACK, p.parseIndexExpression)

	p.registerInfixCursor(lexer.DOT, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseMemberAccess(left) })
	p.registerInfixCursor(lexer.LPAREN, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseCallExpression(left) })
	p.registerInfixCursor(lexer.LBRACK, func(left ast.Expression, tok lexer.Token) ast.Expression { return p.parseIndexExpression(left) })
}

// MustBuild builds the parser and panics if there's an error.
// This is useful in test code where parser construction should never fail.
func (b *ParserBuilder) MustBuild() *Parser {
	p := b.Build()
	if p == nil {
		panic("failed to build parser")
	}
	return p
}
