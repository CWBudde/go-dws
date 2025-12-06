// Package parser implements the DWScript parser.
//
// POSITION TRACKING:
// All parsing functions should set the EndPos field on AST nodes:
//   - Single-token nodes: node.EndPos = p.endPosFromToken(p.cursor.Current())
//   - Multi-token nodes: Set EndPos after all tokens consumed
//   - Optional semicolons: Update EndPos if semicolon is consumed
//
// LOOKAHEAD:
// N-token lookahead for disambiguation:
//   - peek(n): Returns token N positions after peekToken
//   - peekAhead(n): Returns token N positions ahead from curToken
//
// Use sparingly for grammar disambiguation, always check for EOF in loops.
//
// ERROR RECOVERY:
// Panic-mode error recovery with synchronization tokens:
//   - pushBlockContext/popBlockContext: Track nested blocks for error context
//   - addError/addErrorWithContext: Report errors with position and context
//   - synchronize(tokens): Advance to safe point (statement starters, block closers, EOF)
//
// STRUCTURED ERRORS:
// Rich error reporting with suggestions and context:
//   - NewStructuredError(kind): Creates builder with fluent API
//   - addStructuredError(err): Auto-injects block context
//   - Error kinds: Missing, Unexpected, Invalid, Ambiguous
//   - Include expected/actual tokens, suggestions, related positions
//
// PRATT PARSING:
// Top-down operator precedence for expressions:
//   - Prefix functions: Parse tokens at START of expression (literals, unary ops, grouping)
//   - Infix functions: Parse tokens BETWEEN expressions (binary ops, calls, member access)
//   - Precedence levels: LOWEST to MEMBER (higher number = tighter binding)
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

// prefixParseFn is a function type for parsing prefix expressions.
// Takes the token explicitly as a parameter instead of relying on shared mutable state.
// This enables pure functional parsing with immutable cursor navigation.
type prefixParseFn func(lexer.Token) ast.Expression

// infixParseFn is a function type for parsing infix expressions.
// Takes both the left expression and the operator token explicitly.
// This enables pure functional parsing with immutable cursor navigation.
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

// ParserState represents a snapshot of the parser's complete state for speculative parsing.
// Saves errors, context, block stack, and lexer state. Use saveState()/restoreState() for
// heavyweight backtracking across multiple parsing functions. For lightweight cursor-only
// backtracking within a single function, prefer cursor.Mark()/ResetTo() instead.
type ParserState struct {
	ctx                  *ParseContext
	cursor               *TokenCursor
	errors               []*ParserError
	blockStack           []BlockContext
	lexerState           lexer.LexerState
	parsingPostCondition bool
}

// New creates a new Parser instance with default settings.
// For custom configuration, use NewParserBuilder(lexer).WithStrictMode(true).Build().
func New(l *lexer.Lexer) *Parser {
	return NewParserBuilder(l).Build()
}

// NewCursorParser creates a new Parser instance. Alias for New().
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

// peek returns the token N positions after peekToken (n=0 is 2 ahead of curToken).
func (p *Parser) peek(n int) lexer.Token {
	return p.cursor.Peek(n + 2)
}

// peekAhead returns the token N positions ahead from curToken (n=1 is peekToken).
func (p *Parser) peekAhead(n int) lexer.Token {
	if n <= 0 {
		return p.cursor.Current()
	}
	return p.cursor.Peek(n)
}

// expectPeek checks if the peek token is of the given type and advances if so.
// Returns true if the token matches, false otherwise (and adds an error).
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	// Allow contextual keywords that can act as identifiers (e.g., HELPER used as method name)
	if t == lexer.IDENT && p.peekTokenIs(lexer.HELPER) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// isIdentifierToken checks if a token type can be used as an identifier.
// This includes IDENT and contextual keywords like STEP that are keywords in
// specific contexts (for loops) but can be used as variable names elsewhere.
// Also includes SELF which can be the target of member assignments (Self.field := value).
func (p *Parser) isIdentifierToken(t lexer.TokenType) bool {
	return t == lexer.IDENT || t == lexer.STEP || t == lexer.SELF || t == lexer.HELPER
}

// expectIdentifier checks if the peek token can be used as an identifier and advances if so.
// Returns true if the token is valid as an identifier, false otherwise (and adds an error).
// This allows contextual keywords like 'step' to be used as variable names.
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

// addStructuredError adds a structured error to the parser's error list.
// This method provides richer error reporting with context, suggestions, and better formatting.
// The structured error is automatically enhanced with block context if available.
//
// Example usage:
//
//	err := NewStructuredError(ErrKindMissing).
//	    WithCode(ErrMissingRParen).
//	    WithPosition(p.cursor.Current().Pos, p.cursor.Current().Length()).
//	    WithExpected(lexer.RPAREN).
//	    WithSuggestion("add ')' to close the expression").
//	    Build()
//	p.addStructuredError(err)
func (p *Parser) addStructuredError(structErr *StructuredParserError) {
	// Auto-inject block context if not already set
	if structErr.BlockContext == nil {
		structErr.BlockContext = p.currentBlockContext()
	}

	// Convert to legacy ParserError for backward compatibility
	// This ensures existing error handling code continues to work
	legacyErr := structErr.ToParserError()
	p.errors = append(p.errors, legacyErr)
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.addError(msg, ErrNoPrefixParse)
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// getPrecedence returns the precedence of a token type. Returns LOWEST if not found.
func getPrecedence(tokenType lexer.TokenType) int {
	if prec, ok := precedences[tokenType]; ok {
		return prec
	}
	return LOWEST
}

// saveState captures the current parser state for speculative parsing.
// Call restoreState() to backtrack after failed speculative parse.
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

// restoreState restores the parser to a previously saved state.
func (p *Parser) restoreState(state ParserState) {
	p.errors = state.errors
	p.parsingPostCondition = state.parsingPostCondition
	p.blockStack = state.blockStack
	p.l.RestoreState(state.lexerState)
	p.ctx.Restore(state.ctx)
	p.cursor = state.cursor
}

// pushBlockContext tracks nested blocks for better error messages.
func (p *Parser) pushBlockContext(blockType string, startPos lexer.Position) {
	p.ctx.PushBlock(blockType, startPos)
	p.blockStack = append(p.blockStack, BlockContext{
		BlockType: blockType,
		StartPos:  startPos,
		StartLine: startPos.Line,
	})
}

// popBlockContext pops the most recent block context from the stack.
func (p *Parser) popBlockContext() {
	p.ctx.PopBlock()
	if len(p.blockStack) > 0 {
		p.blockStack = p.blockStack[:len(p.blockStack)-1]
	}
}

// currentBlockContext returns the current block context, or nil if none.
func (p *Parser) currentBlockContext() *BlockContext {
	return p.ctx.CurrentBlock()
}

// Synchronization token sets for error recovery.
// These define "safe" points where parsing can resume after an error.
var (
	// statementStarters are tokens that can start a new statement
	statementStarters = []lexer.TokenType{
		lexer.VAR, lexer.CONST, lexer.TYPE,
		lexer.IF, lexer.WHILE, lexer.FOR, lexer.REPEAT, lexer.CASE,
		lexer.BEGIN, lexer.TRY, lexer.RAISE,
		lexer.BREAK, lexer.CONTINUE, lexer.EXIT,
		lexer.FUNCTION, lexer.PROCEDURE,
		lexer.CLASS, lexer.RECORD, lexer.INTERFACE,
		lexer.IDENT,
	}

	// blockClosers are tokens that close blocks
	blockClosers = []lexer.TokenType{
		lexer.END, lexer.UNTIL, lexer.ELSE, lexer.EXCEPT, lexer.FINALLY,
	}

	// declarationStarters are tokens that can start a declaration
	declarationStarters = []lexer.TokenType{
		lexer.VAR, lexer.CONST, lexer.TYPE,
		lexer.FUNCTION, lexer.PROCEDURE,
		lexer.CLASS, lexer.RECORD, lexer.INTERFACE,
	}
)

// synchronize performs panic-mode error recovery by advancing to a safe synchronization point.
// Stops at syncTokens, statement starters, block closers, or EOF. Returns true if sync token found.
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

	// Advance until we find a synchronization token or EOF
	for p.cursor.Current().Type != lexer.EOF {
		if syncMap[p.cursor.Current().Type] {
			return true // Found a sync token
		}
		p.cursor = p.cursor.Advance()
	}

	return false // Reached EOF without finding a sync token
}

// addErrorWithContext adds an error with block context information.
func (p *Parser) addErrorWithContext(msg string, code string) {
	if ctx := p.currentBlockContext(); ctx != nil {
		msg = fmt.Sprintf("%s (in %s block starting at line %d)", msg, ctx.BlockType, ctx.StartLine)
	}
	p.addError(msg, code)
}

// endPosFromToken calculates the end position of a token for AST node EndPos fields.
func (p *Parser) endPosFromToken(tok lexer.Token) lexer.Position {
	pos := tok.Pos
	pos.Column += tok.Length()
	pos.Offset += tok.Length()
	return pos
}

// LIST PARSING HELPERS
//
// Generic helpers for parsing separated lists (comma-separated expressions,
// semicolon-separated parameters, etc.) with flexible separator/terminator handling.

// ListParseOptions configures how parseSeparatedList behaves.
type ListParseOptions struct {
	// Separators are the token types that separate list items (e.g., COMMA, SEMICOLON).
	// At least one separator must match for the list to continue.
	Separators []lexer.TokenType

	// Terminator is the token type that ends the list (e.g., RPAREN, RBRACE).
	Terminator lexer.TokenType

	// AllowTrailingSeparator permits lists like (1, 2, 3,) with separator before terminator.
	AllowTrailingSeparator bool

	// AllowEmpty permits empty lists (when current token is already the terminator).
	// If false and list is empty, returns false to indicate failure.
	AllowEmpty bool

	// RequireTerminator controls whether expectPeek(Terminator) is called at the end.
	// If true, the list must end with the terminator and an error is added if missing.
	// If false, the list ends when no separator is found (caller handles terminator).
	RequireTerminator bool
}

// parseSeparatedList parses lists of items separated by delimiters.
// parseItem callback should parse one item and return success status.
// Returns (itemCount, success). On entry, curToken should be first item or terminator.
func (p *Parser) parseSeparatedList(opts ListParseOptions, parseItem func() bool) (itemCount int, success bool) {
	// Handle empty list
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

		// Check for trailing separator
		if opts.AllowTrailingSeparator && p.peekTokenIs(opts.Terminator) {
			if opts.RequireTerminator {
				p.nextToken() // consume terminator
			}
			// The contract "curToken is the last item" no longer applies.
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

// parseSeparatedListBeforeStart is a variant for when curToken is BEFORE the list (e.g., LPAREN).
// Checks for empty list, advances to first item, then calls parseSeparatedList.
func (p *Parser) parseSeparatedListBeforeStart(opts ListParseOptions, parseItem func() bool) (itemCount int, success bool) {
	// Check for empty list (peek is terminator)
	if p.peekTokenIs(opts.Terminator) {
		if !opts.AllowEmpty {
			return 0, false
		}
		p.nextToken() // consume terminator
		return 0, true
	}

	// Advance to first item
	p.nextToken()

	// Use main helper for the rest
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
			// If parseVarDeclaration() wrapped multiple declarations in a BlockStatement,
			// unwrap it to avoid creating an extra nested scope in the semantic analyzer
			if blockStmt, ok := stmt.(*ast.BlockStatement); ok && p.isVarDeclBlock(blockStmt) {
				// Add each var declaration individually
				program.Statements = append(program.Statements, blockStmt.Statements...)
			} else {
				program.Statements = append(program.Statements, stmt)
			}
		}
		p.nextToken()
	}

	return builder.Finish(program).(*ast.Program)
}

// isVarDeclBlock checks if a BlockStatement wraps multiple var declarations.
// These should be unwrapped to avoid extra scope nesting. Identified by VAR token (not BEGIN).
func (p *Parser) isVarDeclBlock(block *ast.BlockStatement) bool {
	// Must have VAR token (not BEGIN)
	if block.Token.Type != lexer.VAR {
		return false
	}
	// Must contain at least one statement
	if len(block.Statements) == 0 {
		return false
	}
	// All statements must be VarDeclStatement
	for _, stmt := range block.Statements {
		if _, ok := stmt.(*ast.VarDeclStatement); !ok {
			return false
		}
	}
	return true
}

// parseFieldInitializer parses optional field initializer (= Value or := Value).
// Returns initialization expression if present, nil otherwise.
// PRE: cursor at type annotation. POST: cursor at init expression or unchanged.
func (p *Parser) parseFieldInitializer(fieldNames []*ast.Identifier) ast.Expression {
	if p.peekTokenIs(lexer.EQ) || p.peekTokenIs(lexer.ASSIGN) {
		// Initialization is only allowed for single field declarations
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

	return nil
}
