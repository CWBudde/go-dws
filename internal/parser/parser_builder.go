package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ParserConfig holds configuration options for the parser.
// This separates parser configuration from parser state, making it easier
// to create parsers with different configurations.
// Task 2.7.9: UseCursor removed - parser is now cursor-only.
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
// Task 2.7.9: Parser is now cursor-only (no mode selection).
func DefaultConfig() ParserConfig {
	return ParserConfig{
		AllowReservedKeywordsAsIdentifiers: true,
		StrictMode:                         false,
		MaxRecursionDepth:                  1000,
	}
}

// ParserBuilder provides a fluent API for constructing Parser instances.
// It reduces code duplication and makes parser configuration more explicit.
// Task 2.7.9: Parser is now cursor-only.
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
// Task 2.7.9: Parser is now cursor-only.
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

	// Task 2.7.9: Always use cursor mode
	p.cursor = NewTokenCursor(b.lexer)
	p.prefixParseFnsCursor = make(map[lexer.TokenType]prefixParseFnCursor)
	p.infixParseFnsCursor = make(map[lexer.TokenType]infixParseFnCursor)

	// Register all parse functions
	b.registerParseFunctions(p)

	// Task 2.7.9: Register cursor-specific parse functions
	b.registerCursorParseFunctions(p)

	// Task 2.7.13.3: No token state initialization needed - cursor-only mode

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

// registerCursorParseFunctions registers cursor-specific parse functions.
// Task 2.7.9: Extracted from NewCursorParser() to ensure Build() registers cursor functions.
func (b *ParserBuilder) registerCursorParseFunctions(p *Parser) {
	// Register prefix parse functions for cursor mode
	p.registerPrefixCursor(lexer.IDENT, func(tok lexer.Token) ast.Expression {
		return p.parseIdentifierCursor()
	})
	p.registerPrefixCursor(lexer.INT, func(tok lexer.Token) ast.Expression {
		return p.parseIntegerLiteralCursor()
	})
	p.registerPrefixCursor(lexer.FLOAT, func(tok lexer.Token) ast.Expression {
		return p.parseFloatLiteralCursor()
	})
	p.registerPrefixCursor(lexer.STRING, func(tok lexer.Token) ast.Expression {
		return p.parseStringLiteralCursor()
	})
	p.registerPrefixCursor(lexer.TRUE, func(tok lexer.Token) ast.Expression {
		return p.parseBooleanLiteralCursor()
	})
	p.registerPrefixCursor(lexer.FALSE, func(tok lexer.Token) ast.Expression {
		return p.parseBooleanLiteralCursor()
	})
	p.registerPrefixCursor(lexer.NIL, func(tok lexer.Token) ast.Expression {
		return p.parseNilLiteralCursor()
	})
	p.registerPrefixCursor(lexer.NULL, func(tok lexer.Token) ast.Expression {
		return p.parseNullIdentifierCursor()
	})
	p.registerPrefixCursor(lexer.UNASSIGNED, func(tok lexer.Token) ast.Expression {
		return p.parseUnassignedIdentifierCursor()
	})
	p.registerPrefixCursor(lexer.CHAR, func(tok lexer.Token) ast.Expression {
		return p.parseCharLiteralCursor()
	})
	p.registerPrefixCursor(lexer.MINUS, func(tok lexer.Token) ast.Expression {
		return p.parsePrefixExpressionCursor()
	})
	p.registerPrefixCursor(lexer.PLUS, func(tok lexer.Token) ast.Expression {
		return p.parsePrefixExpressionCursor()
	})
	p.registerPrefixCursor(lexer.NOT, func(tok lexer.Token) ast.Expression {
		return p.parsePrefixExpressionCursor()
	})
	p.registerPrefixCursor(lexer.LPAREN, func(tok lexer.Token) ast.Expression {
		return p.parseGroupedExpressionCursor()
	})
	p.registerPrefixCursor(lexer.LBRACK, func(tok lexer.Token) ast.Expression {
		return p.parseArrayLiteralCursor()
	})
	p.registerPrefixCursor(lexer.SELF, func(tok lexer.Token) ast.Expression {
		return p.parseSelfExpressionCursor()
	})
	p.registerPrefixCursor(lexer.INHERITED, func(tok lexer.Token) ast.Expression {
		return p.parseInheritedExpressionCursor()
	})
	p.registerPrefixCursor(lexer.NEW, func(tok lexer.Token) ast.Expression {
		return p.parseNewExpressionCursor()
	})
	p.registerPrefixCursor(lexer.IF, func(tok lexer.Token) ast.Expression {
		return p.parseIfExpressionCursor()
	})
	p.registerPrefixCursor(lexer.LAMBDA, func(tok lexer.Token) ast.Expression {
		return p.parseLambdaExpressionCursor()
	})
	p.registerPrefixCursor(lexer.AT, func(tok lexer.Token) ast.Expression {
		return p.parseAddressOfExpressionCursor()
	})
	p.registerPrefixCursor(lexer.DEFAULT, func(tok lexer.Token) ast.Expression {
		return p.parseDefaultExpressionCursor()
	})
	p.registerPrefixCursor(lexer.OLD, func(tok lexer.Token) ast.Expression {
		return p.parseOldExpressionCursor()
	})

	// Register infix parse functions for cursor mode
	p.registerInfixCursor(lexer.PLUS, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.MINUS, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.ASTERISK, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.SLASH, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.DIV, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.MOD, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.SHL, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.SHR, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.SAR, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.EQ, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.NOT_EQ, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.LESS, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.GREATER, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.LESS_EQ, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.GREATER_EQ, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.AND, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.OR, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.XOR, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.IN, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.QUESTION_QUESTION, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.LPAREN, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseCallExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.DOT, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseMemberAccessCursor(left)
	})
	p.registerInfixCursor(lexer.LBRACK, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseIndexExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.IS, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseIsExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.AS, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseAsExpressionCursor(left)
	})
	p.registerInfixCursor(lexer.IMPLEMENTS, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseImplementsExpressionCursor(left)
	})
}
