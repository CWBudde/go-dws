package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ParserConfig holds configuration options for the parser.
// This separates parser configuration from parser state, making it easier
// to create parsers with different configurations.
type ParserConfig struct {
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
//	    WithStrictMode(false).
//	    Build()
//
// For simple cases, use the convenience constructors:
//
//	parser := New(lexer)              // Default settings
//	parser := NewCursorParser(lexer)  // Alias for New()
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
	}

	p.cursor = NewTokenCursor(b.lexer)
	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)

	// Register all parse functions
	b.registerParseFunctions(p)

	return p
}

// registerParseFunctions registers all prefix and infix parse functions.
// This centralizes the registration logic that was previously duplicated
// between New() and NewCursorParser().
func (b *ParserBuilder) registerParseFunctions(p *Parser) {
	// Register prefix parse functions for literals
	// Note: These parse functions don't take token parameters, so we wrap them in adapters
	p.registerPrefix(lexer.IDENT, func(_ lexer.Token) ast.Expression { return p.parseIdentifier() })
	p.registerPrefix(lexer.INT, func(_ lexer.Token) ast.Expression { return p.parseIntegerLiteral() })
	p.registerPrefix(lexer.FLOAT, func(_ lexer.Token) ast.Expression { return p.parseFloatLiteral() })
	p.registerPrefix(lexer.STRING, func(_ lexer.Token) ast.Expression { return p.parseStringLiteral() })
	p.registerPrefix(lexer.TRUE, func(_ lexer.Token) ast.Expression { return p.parseBooleanLiteral() })
	p.registerPrefix(lexer.FALSE, func(_ lexer.Token) ast.Expression { return p.parseBooleanLiteral() })
	p.registerPrefix(lexer.NIL, func(_ lexer.Token) ast.Expression { return p.parseNilLiteral() })
	p.registerPrefix(lexer.NULL, func(_ lexer.Token) ast.Expression { return p.parseNullIdentifier() })
	p.registerPrefix(lexer.UNASSIGNED, func(_ lexer.Token) ast.Expression { return p.parseUnassignedIdentifier() })
	p.registerPrefix(lexer.CHAR, func(_ lexer.Token) ast.Expression { return p.parseCharLiteral() })

	// Register prefix parse functions for operators
	p.registerPrefix(lexer.MINUS, func(_ lexer.Token) ast.Expression { return p.parsePrefixExpression() })
	p.registerPrefix(lexer.PLUS, func(_ lexer.Token) ast.Expression { return p.parsePrefixExpression() })
	p.registerPrefix(lexer.NOT, func(_ lexer.Token) ast.Expression { return p.parsePrefixExpression() })

	// Register prefix parse functions for grouping and collections
	p.registerPrefix(lexer.LPAREN, func(_ lexer.Token) ast.Expression { return p.parseGroupedExpression() })
	p.registerPrefix(lexer.LBRACK, func(_ lexer.Token) ast.Expression { return p.parseArrayLiteral() })

	// Register prefix parse functions for keywords
	p.registerPrefix(lexer.NEW, func(_ lexer.Token) ast.Expression { return p.parseNewExpression() })
	p.registerPrefix(lexer.DEFAULT, func(_ lexer.Token) ast.Expression { return p.parseDefaultExpression() })
	p.registerPrefix(lexer.AT, func(_ lexer.Token) ast.Expression { return p.parseAddressOfExpression() })
	p.registerPrefix(lexer.LAMBDA, func(_ lexer.Token) ast.Expression { return p.parseLambdaExpression() })
	p.registerPrefix(lexer.OLD, func(_ lexer.Token) ast.Expression { return p.parseOldExpression() })
	p.registerPrefix(lexer.INHERITED, func(_ lexer.Token) ast.Expression { return p.parseInheritedExpression() })
	p.registerPrefix(lexer.SELF, func(_ lexer.Token) ast.Expression { return p.parseSelfExpression() })
	p.registerPrefix(lexer.IF, func(_ lexer.Token) ast.Expression { return p.parseIfExpression() })

	// Register contextual keywords that can be used as identifiers
	if b.config.AllowReservedKeywordsAsIdentifiers {
		p.registerPrefix(lexer.HELPER, func(_ lexer.Token) ast.Expression { return p.parseIdentifier() })
		p.registerPrefix(lexer.STEP, func(_ lexer.Token) ast.Expression { return p.parseIdentifier() })
	}

	// Register infix parse functions for operators
	p.registerInfix(lexer.QUESTION_QUESTION, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) }) // Coalesce: a ?? b
	p.registerInfix(lexer.PLUS, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.MINUS, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.ASTERISK, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.SLASH, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.DIV, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.MOD, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.SHL, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.SHR, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.SAR, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })

	// Register infix parse functions for comparison operators
	p.registerInfix(lexer.EQ, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.NOT_EQ, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.LESS, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.GREATER, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.LESS_EQ, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.GREATER_EQ, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })

	// Register infix parse functions for logical operators
	p.registerInfix(lexer.AND, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.OR, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.XOR, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })

	// Register infix parse functions for type operations
	p.registerInfix(lexer.IN, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseInfixExpression(left) })
	p.registerInfix(lexer.IS, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseIsExpression(left) })
	p.registerInfix(lexer.AS, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseAsExpression(left) })
	p.registerInfix(lexer.IMPLEMENTS, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseImplementsExpression(left) })

	// Register infix parse functions for member access and calls
	p.registerInfix(lexer.DOT, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseMemberAccess(left) })
	p.registerInfix(lexer.LPAREN, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseCallExpression(left) })
	p.registerInfix(lexer.LBRACK, func(left ast.Expression, _ lexer.Token) ast.Expression { return p.parseIndexExpression(left) })
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
