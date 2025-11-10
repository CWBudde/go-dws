// Package parser implements the DWScript parser.
//
// POSITION TRACKING PATTERN (Task 10.10):
//
// All parsing functions should set the EndPos field on AST nodes they create.
// The general pattern is:
//
//  1. For single-token nodes (literals, identifiers):
//     node.EndPos = p.endPosFromToken(p.curToken)
//
// 2. For multi-token nodes:
//   - Set EndPos after all tokens are consumed
//   - Usually: node.EndPos = p.endPosFromToken(p.curToken)
//   - Or delegate to child expression: node.EndPos = childExpr.End()
//
// 3. For nodes with optional semicolons:
//   - Set EndPos first based on main content
//   - Update EndPos if semicolon is consumed
//
// Example:
//
//	stmt.Expression = p.parseExpression(LOWEST)
//	stmt.EndPos = stmt.Expression.End()
//	if p.peekTokenIs(lexer.SEMICOLON) {
//	    p.nextToken()
//	    stmt.EndPos = p.endPosFromToken(p.curToken)
//	}
//
// Note: As of task 10.10 implementation, position tracking is partially complete.
// Many parsing functions still need EndPos population. Follow the pattern above.
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
	COALESCE    // ?? (higher than ASSIGN so it works in assignment RHS)
	OR          // or
	AND         // and
	EQUALS      // = <>
	LESSGREATER // < > <= >=
	SUM         // + -
	SHIFT       // shl shr
	PRODUCT     // * / div mod
	PREFIX      // -x, not x, +x
	CALL        // function(args)
	INDEX       // array[index]
	MEMBER      // obj.field
)

// precedences maps token types to their precedence levels.
var precedences = map[lexer.TokenType]int{
	lexer.QUESTION_QUESTION: COALESCE,
	lexer.ASSIGN:            ASSIGN,
	lexer.OR:                OR,
	lexer.XOR:               OR,
	lexer.AND:               AND,
	lexer.EQ:                EQUALS,
	lexer.NOT_EQ:            EQUALS,
	lexer.IN:                EQUALS, // Set membership test
	lexer.IS:                EQUALS, // Type checking: obj is TClass
	lexer.AS:                EQUALS, // Type casting: obj as IInterface
	lexer.IMPLEMENTS:        EQUALS, // Interface check: obj implements IInterface
	lexer.LESS:              LESSGREATER,
	lexer.GREATER:           LESSGREATER,
	lexer.LESS_EQ:           LESSGREATER,
	lexer.GREATER_EQ:        LESSGREATER,
	lexer.PLUS:              SUM,
	lexer.MINUS:             SUM,
	lexer.SHL:               SHIFT,
	lexer.SHR:               SHIFT,
	lexer.SAR:               SHIFT,
	lexer.ASTERISK:          PRODUCT,
	lexer.SLASH:             PRODUCT,
	lexer.DIV:               PRODUCT,
	lexer.MOD:               PRODUCT,
	lexer.LPAREN:            CALL,
	lexer.LBRACK:            INDEX,
	lexer.DOT:               MEMBER,
	// Note: Compound assignment operators (+=, -=, *=, /=) are NOT in this table
	// because they are statement-level operators, not expression operators.
	// They are handled in parseAssignmentOrExpression() in statements.go
}

// prefixParseFn is a function type for parsing prefix expressions.
type prefixParseFn func() ast.Expression

// infixParseFn is a function type for parsing infix expressions.
type infixParseFn func(ast.Expression) ast.Expression

// Parser represents the DWScript parser.
type Parser struct {
	l                      *lexer.Lexer
	prefixParseFns         map[lexer.TokenType]prefixParseFn
	infixParseFns          map[lexer.TokenType]infixParseFn
	errors                 []*ParserError
	semanticErrors         []string
	curToken               lexer.Token
	peekToken              lexer.Token
	enableSemanticAnalysis bool
	parsingPostCondition   bool // True when parsing postconditions (for 'old' keyword)
}

// New creates a new Parser instance.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:                      l,
		errors:                 []*ParserError{},
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
	p.registerPrefix(lexer.CHAR, p.parseCharLiteral)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.PLUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.LBRACK, p.parseArrayLiteral)           // Array/Set literals: [a, b]
	p.registerPrefix(lexer.NEW, p.parseNewExpression)             // new keyword: new Exception('msg')
	p.registerPrefix(lexer.AT, p.parseAddressOfExpression)        // Address-of operator: @FunctionName
	p.registerPrefix(lexer.LAMBDA, p.parseLambdaExpression)       // Lambda expressions: lambda(x) => x * 2
	p.registerPrefix(lexer.OLD, p.parseOldExpression)             // old keyword: old identifier (postconditions only)
	p.registerPrefix(lexer.INHERITED, p.parseInheritedExpression) // inherited keyword: inherited MethodName(args)

	// Register keywords that can be used as identifiers in expression context
	// In DWScript/Object Pascal, some keywords can be used as identifiers
	p.registerPrefix(lexer.HELPER, p.parseIdentifier)
	p.registerPrefix(lexer.STEP, p.parseIdentifier) // 'step' is contextual - keyword in for loops, identifier elsewhere

	// Register infix parse functions
	p.registerInfix(lexer.QUESTION_QUESTION, p.parseInfixExpression) // Coalesce: a ?? b
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.LBRACK, p.parseIndexExpression) // Array/string indexing: arr[i]
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.ASTERISK, p.parseInfixExpression)
	p.registerInfix(lexer.SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.DIV, p.parseInfixExpression)
	p.registerInfix(lexer.MOD, p.parseInfixExpression)
	p.registerInfix(lexer.SHL, p.parseInfixExpression)
	p.registerInfix(lexer.SHR, p.parseInfixExpression)
	p.registerInfix(lexer.SAR, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LESS, p.parseInfixExpression)
	p.registerInfix(lexer.GREATER, p.parseInfixExpression)
	p.registerInfix(lexer.LESS_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.GREATER_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.AND, p.parseInfixExpression)
	p.registerInfix(lexer.OR, p.parseInfixExpression)
	p.registerInfix(lexer.XOR, p.parseInfixExpression)
	p.registerInfix(lexer.IN, p.parseInfixExpression)              // Set membership test
	p.registerInfix(lexer.IS, p.parseIsExpression)                 // Type checking: obj is TClass
	p.registerInfix(lexer.AS, p.parseAsExpression)                 // Type casting: obj as IInterface
	p.registerInfix(lexer.IMPLEMENTS, p.parseImplementsExpression) // Interface check: obj implements IInterface
	p.registerInfix(lexer.DOT, p.parseMemberAccess)

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns the list of parsing errors.
func (p *Parser) Errors() []*ParserError {
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

// isIdentifierToken checks if a token type can be used as an identifier.
// This includes IDENT and contextual keywords like STEP that are keywords in
// specific contexts (for loops) but can be used as variable names elsewhere.
func (p *Parser) isIdentifierToken(t lexer.TokenType) bool {
	return t == lexer.IDENT || t == lexer.STEP
}

// expectIdentifier checks if the peek token can be used as an identifier and advances if so.
// Returns true if the token is valid as an identifier, false otherwise (and adds an error).
// This allows contextual keywords like 'step' to be used as variable names.
func (p *Parser) expectIdentifier() bool {
	if p.isIdentifierToken(p.peekToken.Type) {
		p.nextToken()
		return true
	}
	p.peekError(lexer.IDENT)
	return false
}

// peekError adds an error about an unexpected peek token.
func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	err := NewParserError(
		p.peekToken.Pos,
		p.peekToken.Length(),
		msg,
		ErrUnexpectedToken,
	)
	p.errors = append(p.errors, err)
}

// addError adds a generic error message with the specified error code.
func (p *Parser) addError(msg string, code string) {
	err := NewParserError(
		p.curToken.Pos,
		p.curToken.Length(),
		msg,
		code,
	)
	p.errors = append(p.errors, err)
}

// addGenericError adds a generic error message with a default error code.
func (p *Parser) addGenericError(msg string) {
	p.addError(msg, ErrInvalidExpression)
}

// noPrefixParseFnError adds an error for missing prefix parse function.
func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.addError(msg, ErrNoPrefixParse)
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

// endPosFromToken calculates the end position of a token.
// This is a helper function to populate EndPos fields in AST nodes.
func (p *Parser) endPosFromToken(tok lexer.Token) lexer.Position {
	pos := tok.Pos
	pos.Column += tok.Length()
	pos.Offset += tok.Length()
	return pos
}

// ParseProgram parses the entire program and returns the AST root node.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// If the file starts with 'unit', parse it as a unit
	if p.curTokenIs(lexer.UNIT) {
		unit := p.parseUnit()
		if unit != nil {
			program.Statements = append(program.Statements, unit)
		}
		return program
	}

	// If the file starts with 'program', parse and skip it
	if p.curTokenIs(lexer.PROGRAM) {
		p.parseProgramDeclaration()
		p.nextToken() // Move past the semicolon to the next token
	}

	// Otherwise, parse as a regular program
	for !p.curTokenIs(lexer.EOF) {
		// Skip semicolons at statement level
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// Handle program terminator: end.
		// A DOT after the main program block terminates parsing
		if p.curTokenIs(lexer.DOT) {
			p.nextToken() // consume the dot
			break         // end of program
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	// Set end position to the last token processed
	program.EndPos = p.endPosFromToken(p.curToken)

	return program
}
