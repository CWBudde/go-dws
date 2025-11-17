package parser

import (
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
		UseCursor:                          false,
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

	// Configure cursor mode
	if b.config.UseCursor {
		p.cursor = NewTokenCursor(b.lexer)
		p.prefixParseFnsCursor = make(map[lexer.TokenType]prefixParseFnCursor)
		p.infixParseFnsCursor = make(map[lexer.TokenType]infixParseFnCursor)
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
	// Register prefix parse functions for literals
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

	// Register prefix parse functions for operators
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.PLUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.NOT, p.parsePrefixExpression)

	// Register prefix parse functions for grouping and collections
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.LBRACK, p.parseArrayLiteral)

	// Register prefix parse functions for keywords
	p.registerPrefix(lexer.NEW, p.parseNewExpression)
	p.registerPrefix(lexer.DEFAULT, p.parseDefaultExpression)
	p.registerPrefix(lexer.AT, p.parseAddressOfExpression)
	p.registerPrefix(lexer.LAMBDA, p.parseLambdaExpression)
	p.registerPrefix(lexer.OLD, p.parseOldExpression)
	p.registerPrefix(lexer.INHERITED, p.parseInheritedExpression)
	p.registerPrefix(lexer.SELF, p.parseSelfExpression)
	p.registerPrefix(lexer.IF, p.parseIfExpression)

	// Register contextual keywords that can be used as identifiers
	if b.config.AllowReservedKeywordsAsIdentifiers {
		p.registerPrefix(lexer.HELPER, p.parseIdentifier)
		p.registerPrefix(lexer.STEP, p.parseIdentifier)
	}

	// Register infix parse functions for operators
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

	// Register infix parse functions for comparison operators
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LESS, p.parseInfixExpression)
	p.registerInfix(lexer.GREATER, p.parseInfixExpression)
	p.registerInfix(lexer.LESS_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.GREATER_EQ, p.parseInfixExpression)

	// Register infix parse functions for logical operators
	p.registerInfix(lexer.AND, p.parseInfixExpression)
	p.registerInfix(lexer.OR, p.parseInfixExpression)
	p.registerInfix(lexer.XOR, p.parseInfixExpression)

	// Register infix parse functions for type operations
	p.registerInfix(lexer.IN, p.parseInfixExpression)
	p.registerInfix(lexer.IS, p.parseIsExpression)
	p.registerInfix(lexer.AS, p.parseAsExpression)
	p.registerInfix(lexer.IMPLEMENTS, p.parseImplementsExpression)

	// Register infix parse functions for member access and calls
	p.registerInfix(lexer.DOT, p.parseMemberAccess)
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.LBRACK, p.parseIndexExpression)
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
