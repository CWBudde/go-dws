// Package parser implements the DWScript parser using Pratt parsing.
//
// Key patterns:
//   - Position tracking: Set EndPos on all AST nodes after parsing
//   - Lookahead: peek(n) for N tokens after peekToken, peekAhead(n) from curToken
//   - Error recovery: pushBlockContext/popBlockContext + synchronize() for panic-mode
//   - Structured errors: NewStructuredError() with auto-injected block context
//
// See docs/parser-architecture.md for detailed explanation.
package parser

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
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

// prefixParseFn parses prefix expressions (literals, unary ops, grouping).
type prefixParseFn func(lexer.Token) ast.Expression

// infixParseFn parses infix expressions (binary ops, calls, member access).
type infixParseFn func(ast.Expression, lexer.Token) ast.Expression

// BlockContext represents the context of a block being parsed.
// Used for better error messages and error recovery.
type BlockContext struct {
	BlockType string         // "begin", "if", "while", "for", "case", "try", "class", "function", etc.
	StartPos  lexer.Position // Position where the block started
	StartLine int            // Line number where the block started
}

// Parser represents the DWScript parser.
type Parser struct {
	l                    *lexer.Lexer
	prefixParseFns       map[lexer.TokenType]prefixParseFn
	infixParseFns        map[lexer.TokenType]infixParseFn
	ctx                  *ParseContext
	cursor               *TokenCursor
	errors               []*ParserError
	blockStack           []BlockContext
	parsingPostCondition bool
}

// ParserState is a heavyweight snapshot for speculative parsing with full backtracking.
// Use saveState()/restoreState() when you need to preserve/discard errors.
// For lightweight position-only backtracking, use cursor.Mark()/ResetTo() instead.
type ParserState struct {
	ctx                  *ParseContext
	cursor               *TokenCursor
	errors               []*ParserError
	blockStack           []BlockContext
	lexerState           lexer.LexerState
	parsingPostCondition bool
}

// New creates a new Parser with default settings.
// For custom configuration, use NewParserBuilder(l).WithStrictMode(true).Build().
func New(l *lexer.Lexer) *Parser {
	return NewParserBuilder(l).Build()
}

// NewCursorParser creates a new Parser (backward compatibility alias for New).
func NewCursorParser(l *lexer.Lexer) *Parser {
	return NewParserBuilder(l).Build()
}

// Errors returns the list of parsing errors.
func (p *Parser) Errors() []*ParserError {
	return p.errors
}

// LexerErrors returns all lexer errors accumulated during tokenization.
// This should be checked in addition to parser errors for complete error reporting.
func (p *Parser) LexerErrors() []lexer.LexerError {
	return p.l.Errors()
}

// nextToken advances the cursor.
func (p *Parser) nextToken() {
	p.cursor = p.cursor.Advance()
}

// curTokenIs checks if the current token is of the given type.
func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.cursor.Current().Type == t
}

// peekTokenIs checks if the peek token is of the given type.
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.cursor.Peek(1).Type == t
}

// peek returns token N positions after peekToken. peek(0) = 2 ahead, peek(1) = 3 ahead.
func (p *Parser) peek(n int) lexer.Token {
	return p.cursor.Peek(n + 2)
}

// peekAhead returns token N positions ahead from curToken. peekAhead(1) = peekToken.
func (p *Parser) peekAhead(n int) lexer.Token {
	if n <= 0 {
		return p.cursor.Current()
	}
	return p.cursor.Peek(n)
}

// expectPeek advances if peek token matches, otherwise adds error and returns false.
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	// Allow contextual keywords as identifiers (e.g., HELPER as method name)
	if t == lexer.IDENT && p.peekTokenIs(lexer.HELPER) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// isIdentifierToken returns true for IDENT and contextual keywords (STEP, SELF, HELPER).
func (p *Parser) isIdentifierToken(t lexer.TokenType) bool {
	return t == lexer.IDENT || t == lexer.STEP || t == lexer.SELF || t == lexer.HELPER
}

// expectIdentifier advances if peek token can be used as an identifier.
func (p *Parser) expectIdentifier() bool {
	peekType := p.cursor.Peek(1).Type

	if p.isIdentifierToken(peekType) {
		p.nextToken()
		return true
	}
	p.peekError(lexer.IDENT)
	return false
}

// peekError adds an error about an unexpected peek token.
func (p *Parser) peekError(t lexer.TokenType) {
	peekTok := p.cursor.Peek(1)

	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, peekTok.Type)
	err := NewParserError(
		peekTok.Pos,
		peekTok.Length(),
		msg,
		ErrUnexpectedToken,
	)
	p.errors = append(p.errors, err)
}

// addError adds a generic error message with the specified error code.
func (p *Parser) addError(msg string, code string) {
	curTok := p.cursor.Current()

	err := NewParserError(
		curTok.Pos,
		curTok.Length(),
		msg,
		code,
	)
	p.errors = append(p.errors, err)
}

// addStructuredError adds a structured error with auto-injected block context.
func (p *Parser) addStructuredError(structErr *StructuredParserError) {
	if structErr.BlockContext == nil {
		structErr.BlockContext = p.currentBlockContext()
	}
	p.errors = append(p.errors, structErr.ToParserError())
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

// getPrecedence returns the precedence of a token type (LOWEST if not found).
func getPrecedence(tokenType lexer.TokenType) int {
	if prec, ok := precedences[tokenType]; ok {
		return prec
	}
	return LOWEST
}

// saveState captures full parser state for speculative parsing with backtracking.
func (p *Parser) saveState() ParserState {
	errorsCopy := make([]*ParserError, len(p.errors))
	copy(errorsCopy, p.errors)
	blockStackCopy := make([]BlockContext, len(p.blockStack))
	copy(blockStackCopy, p.blockStack)

	return ParserState{
		errors:               errorsCopy,
		lexerState:           p.l.SaveState(),
		parsingPostCondition: p.parsingPostCondition,
		blockStack:           blockStackCopy,
		ctx:                  p.ctx.Snapshot(),
		cursor:               p.cursor,
	}
}

// restoreState restores parser to a previously saved state (undoes all changes).
func (p *Parser) restoreState(state ParserState) {
	p.errors = state.errors
	p.parsingPostCondition = state.parsingPostCondition
	p.blockStack = state.blockStack
	p.l.RestoreState(state.lexerState)
	p.ctx.Restore(state.ctx)
	p.cursor = state.cursor
}

// pushBlockContext tracks block nesting for better error messages.
func (p *Parser) pushBlockContext(blockType string, startPos lexer.Position) {
	p.ctx.PushBlock(blockType, startPos)
	p.blockStack = append(p.blockStack, BlockContext{
		BlockType: blockType,
		StartPos:  startPos,
		StartLine: startPos.Line,
	})
}

// popBlockContext exits the current block context.
func (p *Parser) popBlockContext() {
	p.ctx.PopBlock()
	if len(p.blockStack) > 0 {
		p.blockStack = p.blockStack[:len(p.blockStack)-1]
	}
}

// currentBlockContext returns the innermost block context, or nil if none.
func (p *Parser) currentBlockContext() *BlockContext {
	return p.ctx.CurrentBlock()
}

// Synchronization token sets for error recovery (safe points to resume parsing).
var (
	statementStarters = []lexer.TokenType{
		lexer.VAR, lexer.CONST, lexer.TYPE,
		lexer.IF, lexer.WHILE, lexer.FOR, lexer.REPEAT, lexer.CASE,
		lexer.BEGIN, lexer.TRY, lexer.RAISE,
		lexer.BREAK, lexer.CONTINUE, lexer.EXIT,
		lexer.FUNCTION, lexer.PROCEDURE,
		lexer.CLASS, lexer.RECORD, lexer.INTERFACE,
		lexer.IDENT,
	}

	blockClosers = []lexer.TokenType{
		lexer.END, lexer.UNTIL, lexer.ELSE, lexer.EXCEPT, lexer.FINALLY,
	}
	declarationStarters = []lexer.TokenType{
		lexer.VAR, lexer.CONST, lexer.TYPE,
		lexer.FUNCTION, lexer.PROCEDURE,
		lexer.CLASS, lexer.RECORD, lexer.INTERFACE,
	}
)

// synchronize advances to a safe point after an error.
// Stops at syncTokens, statement starters, block closers, or EOF.
func (p *Parser) synchronize(syncTokens []lexer.TokenType) bool {
	syncMap := make(map[lexer.TokenType]bool)
	for _, t := range syncTokens {
		syncMap[t] = true
	}
	for _, t := range statementStarters {
		syncMap[t] = true
	}
	for _, t := range blockClosers {
		syncMap[t] = true
	}

	for p.cursor.Current().Type != lexer.EOF {
		if syncMap[p.cursor.Current().Type] {
			return true
		}
		p.cursor = p.cursor.Advance()
	}
	return false
}

// addErrorWithContext adds an error with block context info appended.
func (p *Parser) addErrorWithContext(msg string, code string) {
	if ctx := p.currentBlockContext(); ctx != nil {
		msg = fmt.Sprintf("%s (in %s block starting at line %d)", msg, ctx.BlockType, ctx.StartLine)
	}
	p.addError(msg, code)
}

// endPosFromToken calculates the end position of a token for AST EndPos fields.
func (p *Parser) endPosFromToken(tok lexer.Token) lexer.Position {
	pos := tok.Pos
	pos.Column += tok.Length()
	pos.Offset += tok.Length()
	return pos
}

// ListParseOptions configures parseSeparatedList behavior.
type ListParseOptions struct {
	Separators             []lexer.TokenType // Token types that separate items (e.g., COMMA)
	Terminator             lexer.TokenType   // Token that ends the list (e.g., RPAREN)
	AllowTrailingSeparator bool              // Allow (1, 2, 3,) with trailing separator
	AllowEmpty             bool              // Allow empty lists
	RequireTerminator      bool              // Automatically consume terminator at end
}

// parseSeparatedList parses a delimited list using the provided callback.
// Entry: curToken is first item or terminator. Exit: curToken is terminator (if required).
func (p *Parser) parseSeparatedList(opts ListParseOptions, parseItem func() bool) (itemCount int, success bool) {
	if p.curTokenIs(opts.Terminator) {
		if !opts.AllowEmpty {
			return 0, false
		}
		return 0, true
	}

	// Parse first item
	if !parseItem() {
		return 0, false
	}
	itemCount = 1

	// Parse remaining items
	for p.peekTokenIsSomeOf(opts.Separators...) {
		p.nextToken() // consume separator
		if opts.AllowTrailingSeparator && p.peekTokenIs(opts.Terminator) {
			if opts.RequireTerminator {
				p.nextToken()
			}
			return itemCount, true
		}
		p.nextToken() // move to next item

		// Parse the item
		if !parseItem() {
			return itemCount, false
		}
		itemCount++
	}

	// Expect terminator
	if opts.RequireTerminator {
		if !p.expectPeek(opts.Terminator) {
			return itemCount, false
		}
	}

	return itemCount, true
}

// peekTokenIsSomeOf checks if peekToken is one of the given types.
func (p *Parser) peekTokenIsSomeOf(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.peekTokenIs(t) {
			return true
		}
	}
	return false
}

// parseSeparatedListBeforeStart is like parseSeparatedList but curToken is before list (e.g., LPAREN).
func (p *Parser) parseSeparatedListBeforeStart(opts ListParseOptions, parseItem func() bool) (itemCount int, success bool) {
	if p.peekTokenIs(opts.Terminator) {
		if !opts.AllowEmpty {
			return 0, false
		}
		p.nextToken()
		return 0, true
	}

	// Advance to first item
	p.nextToken()
	return p.parseSeparatedList(opts, parseItem)
}

// ParseProgram parses the entire program and returns the AST root node.
func (p *Parser) ParseProgram() *ast.Program {
	builder := p.StartNode()
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
		p.nextToken()
	}

	// Otherwise, parse as a regular program
	for !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// A DOT after the main program block terminates parsing
		if p.curTokenIs(lexer.DOT) {
			p.nextToken()
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			// Unwrap var declaration blocks to avoid extra scope nesting
			if blockStmt, ok := stmt.(*ast.BlockStatement); ok && p.isVarDeclBlock(blockStmt) {
				program.Statements = append(program.Statements, blockStmt.Statements...)
			} else {
				program.Statements = append(program.Statements, stmt)
			}
		}
		p.nextToken()
	}

	return builder.Finish(program).(*ast.Program)
}

// isVarDeclBlock checks if a BlockStatement wraps var declarations (VAR token, not BEGIN).
func (p *Parser) isVarDeclBlock(block *ast.BlockStatement) bool {
	if block.Token.Type != lexer.VAR || len(block.Statements) == 0 {
		return false
	}
	for _, stmt := range block.Statements {
		if _, ok := stmt.(*ast.VarDeclStatement); !ok {
			return false
		}
	}
	return true
}

// parseFieldInitializer parses an optional field initializer (= Value or := Value).
// PRE: cursor is last token of type annotation. POST: cursor is last token of init expr if present.
func (p *Parser) parseFieldInitializer(fieldNames []*ast.Identifier) ast.Expression {
	if !p.peekTokenIs(lexer.EQ) && !p.peekTokenIs(lexer.ASSIGN) {
		return nil
	}
	if len(fieldNames) > 1 {
		p.addError("initialization not allowed for comma-separated field declarations", ErrInvalidExpression)
		return nil
	}

	p.nextToken() // move to '=' or ':='
	p.nextToken() // move to value expression

	// Parse initialization expression
	initValue := p.parseExpression(LOWEST)
	if initValue == nil {
		p.addError("expected initialization expression after = or :=", ErrInvalidExpression)
		return nil
	}
	return initValue
}
