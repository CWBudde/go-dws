package parser

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// Precedence levels for operators (lowest to highest).
const (
	_ int = iota
	LOWEST
	ASSIGN      // :=
	OR          // or
	AND         // and
	EQUALS      // = <>
	LESSGREATER // < > <= >=
	SUM         // + -
	PRODUCT     // * / div mod
	PREFIX      // -x, not x, +x
	CALL        // function(args)
	INDEX       // array[index]
	MEMBER      // obj.field
)

// precedences maps token types to their precedence levels.
var precedences = map[lexer.TokenType]int{
	lexer.ASSIGN:       ASSIGN,
	lexer.OR:           OR,
	lexer.XOR:          OR,
	lexer.AND:          AND,
	lexer.EQ:           EQUALS,
	lexer.NOT_EQ:       EQUALS,
	lexer.IN:           EQUALS, // Set membership test
	lexer.LESS:         LESSGREATER,
	lexer.GREATER:      LESSGREATER,
	lexer.LESS_EQ:      LESSGREATER,
	lexer.GREATER_EQ:   LESSGREATER,
	lexer.PLUS:         SUM,
	lexer.MINUS:        SUM,
	lexer.ASTERISK:     PRODUCT,
	lexer.SLASH:        PRODUCT,
	lexer.DIV:          PRODUCT,
	lexer.MOD:          PRODUCT,
	lexer.LPAREN:       CALL,
	lexer.LBRACK:       INDEX,
	lexer.DOT:          MEMBER,
	lexer.PLUS_ASSIGN:  ASSIGN,
	lexer.MINUS_ASSIGN: ASSIGN,
}

// prefixParseFn is a function type for parsing prefix expressions.
type prefixParseFn func() ast.Expression

// infixParseFn is a function type for parsing infix expressions.
type infixParseFn func(ast.Expression) ast.Expression

// Parser represents the DWScript parser.
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn

	// Semantic analysis
	enableSemanticAnalysis bool
	semanticErrors         []string
}

// New creates a new Parser instance.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:                      l,
		errors:                 []string{},
		prefixParseFns:         make(map[lexer.TokenType]prefixParseFn),
		infixParseFns:          make(map[lexer.TokenType]infixParseFn),
		enableSemanticAnalysis: false,
		semanticErrors:         []string{},
	}

	// Register prefix parse functions
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.NIL, p.parseNilLiteral)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.PLUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.LBRACK, p.parseSetLiteral) // Set literals: [one, two]

	// Register keywords that can be used as identifiers in expression context
	// In DWScript/Object Pascal, some keywords can be used as identifiers
	p.registerPrefix(lexer.HELPER, p.parseIdentifier)

	// Register infix parse functions
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.LBRACK, p.parseIndexExpression) // Array/string indexing: arr[i]
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.ASTERISK, p.parseInfixExpression)
	p.registerInfix(lexer.SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.DIV, p.parseInfixExpression)
	p.registerInfix(lexer.MOD, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LESS, p.parseInfixExpression)
	p.registerInfix(lexer.GREATER, p.parseInfixExpression)
	p.registerInfix(lexer.LESS_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.GREATER_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.AND, p.parseInfixExpression)
	p.registerInfix(lexer.OR, p.parseInfixExpression)
	p.registerInfix(lexer.XOR, p.parseInfixExpression)
	p.registerInfix(lexer.IN, p.parseInfixExpression) // Set membership test
	p.registerInfix(lexer.DOT, p.parseMemberAccess)

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns the list of parsing errors.
func (p *Parser) Errors() []string {
	return p.errors
}

// EnableSemanticAnalysis enables or disables semantic analysis during parsing.
func (p *Parser) EnableSemanticAnalysis(enable bool) {
	p.enableSemanticAnalysis = enable
}

// SemanticErrors returns the list of semantic errors (if semantic analysis was enabled).
func (p *Parser) SemanticErrors() []string {
	return p.semanticErrors
}

// SetSemanticErrors sets the semantic errors (called by external semantic analyzer).
func (p *Parser) SetSemanticErrors(errors []string) {
	p.semanticErrors = errors
}

// nextToken advances both curToken and peekToken.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs checks if the current token is of the given type.
func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs checks if the peek token is of the given type.
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks if the peek token is of the given type and advances if so.
// Returns true if the token matches, false otherwise (and adds an error).
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// peekError adds an error about an unexpected peek token.
func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead at %d:%d",
		t, p.peekToken.Type, p.peekToken.Pos.Line, p.peekToken.Pos.Column)
	p.errors = append(p.errors, msg)
}

// addError adds a generic error message.
func (p *Parser) addError(msg string) {
	fullMsg := fmt.Sprintf("%s at %d:%d", msg, p.curToken.Pos.Line, p.curToken.Pos.Column)
	p.errors = append(p.errors, fullMsg)
}

// noPrefixParseFnError adds an error for missing prefix parse function.
func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.addError(msg)
}

// registerPrefix registers a prefix parse function for a token type.
func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix registers an infix parse function for a token type.
func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// peekPrecedence returns the precedence of the peek token.
func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}

// curPrecedence returns the precedence of the current token.
func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return LOWEST
}

// ParseProgram parses the entire program and returns the AST root node.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(lexer.EOF) {
		// Skip semicolons at statement level
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}
