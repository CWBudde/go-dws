// Package parser implements the DWScript parser.
//
// POSITION TRACKING PATTERN (Task 10.10):
//
// All parsing functions should set the EndPos field on AST nodes they create.
// The general pattern is:
//
//  1. For single-token nodes (literals, identifiers):
//     node.EndPos = p.endPosFromToken(p.cursor.Current())
//
// 2. For multi-token nodes:
//   - Set EndPos after all tokens are consumed
//   - Usually: node.EndPos = p.endPosFromToken(p.cursor.Current())
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
//	    stmt.EndPos = p.endPosFromToken(p.cursor.Current())
//	}
//
// Note: As of task 10.10 implementation, position tracking is partially complete.
// Many parsing functions still need EndPos population. Follow the pattern above.
//
// LOOKAHEAD PATTERN (Phase 2.1):
//
// The parser supports N-token lookahead via helper methods that wrap the lexer's Peek() capability.
// Use lookahead for disambiguation when grammar is ambiguous or context-dependent.
//
// Available lookahead methods:
//
//  1. peek(n int) lexer.Token
//     - Returns the token N positions after peekToken
//     - peek(0) = token after peekToken (2 tokens ahead of curToken)
//     - peek(1) = 2 tokens after peekToken (3 tokens ahead of curToken)
//     - Direct wrapper around p.l.Peek(n)
//
//  2. peekAhead(n int) lexer.Token
//     - Returns the token N positions ahead from curToken
//     - peekAhead(1) = peekToken
//     - peekAhead(2) = peek(0)
//     - More intuitive counting from curToken
//
// Common lookahead patterns:
//
//  1. Disambiguation functions (looksLike* pattern):
//     func (p *Parser) looksLikeVarDeclaration() bool {
//     if !p.peekTokenIs(lexer.IDENT) {
//     return false
//     }
//     tokenAfterIdent := p.peek(0)  // Look past the IDENT
//     return tokenAfterIdent.Type == lexer.COLON ||
//     tokenAfterIdent.Type == lexer.COMMA
//     }
//
//  2. Scanning for specific tokens:
//     peekIndex := 0
//     for {
//     tok := p.peek(peekIndex)
//     if tok.Type == lexer.COLON {
//     return true
//     }
//     if tok.Type == lexer.SEMICOLON || tok.Type == lexer.EOF {
//     return false
//     }
//     peekIndex++
//     }
//
// Best practices:
//   - Use lookahead sparingly - only when truly needed for disambiguation
//   - Prefer peek() for direct lookahead, peekAhead() when counting from curToken is clearer
//   - Always check for EOF when scanning ahead in loops
//   - Document WHY lookahead is needed (what ambiguity it resolves)
//   - Keep lookahead functions pure (no side effects, no token consumption)
//
// ERROR RECOVERY PATTERN (Phase 2.8):
//
// The parser implements panic-mode error recovery with synchronization tokens
// to enable multiple error reporting in a single pass and prevent infinite loops.
//
// Key components:
//
//  1. Block Context Tracking:
//     - Use pushBlockContext() when entering a block (begin, if, while, for, case, etc.)
//     - Use defer popBlockContext() to ensure cleanup even on error
//     - Provides context for better error messages
//
//  2. Error Reporting:
//     - addError(): Basic error with position tracking
//     - addErrorWithContext(): Include block context in error message
//     - Example: "expected 'end' (in begin block starting at line 10)"
//
//  3. Synchronization:
//     - synchronize(tokens): Advance to a safe synchronization point
//     - Stops at statement starters, block closers, or specified tokens
//     - Prevents parser from getting stuck in error loops
//
// Example usage:
//
//	func (p *Parser) parseWhileStatement() *ast.WhileStatement {
//	    // Track block context
//	    p.pushBlockContext("while", p.cursor.Current().Pos)
//	    defer p.popBlockContext()
//
//	    // Parse condition
//	    stmt.Condition = p.parseExpression(LOWEST)
//	    if stmt.Condition == nil {
//	        p.addErrorWithContext("expected condition", ErrInvalidExpression)
//	        p.synchronize([]lexer.TokenType{lexer.DO, lexer.END})
//	        return nil
//	    }
//
//	    // Try to recover from missing 'do'
//	    if !p.expectPeek(lexer.DO) {
//	        p.addErrorWithContext("expected 'do'", ErrMissingDo)
//	        p.synchronize([]lexer.TokenType{lexer.DO, lexer.END})
//	        if !p.curTokenIs(lexer.DO) {
//	            return nil  // Cannot recover
//	        }
//	        // Found DO, try to continue
//	    }
//	}
//
// Synchronization points (statementStarters, blockClosers):
//   - Statement starters: VAR, CONST, TYPE, IF, WHILE, FOR, REPEAT, CASE, BEGIN, etc.
//   - Block closers: END, UNTIL, ELSE, EXCEPT, FINALLY
//   - Always: EOF (prevents infinite loops)
//
// Best practices:
//   - Always use block context for block-level constructs
//   - Synchronize after errors to enable multiple error reporting
//   - Try to continue parsing when possible (don't give up at first error)
//   - Document which errors are recoverable vs. fatal
//
// STRUCTURED ERROR REPORTING (Phase 2.1.1):
//
// The parser supports both legacy string-based errors and modern structured errors.
// Structured errors provide richer context, suggestions, and better IDE/LSP integration.
//
// Key components:
//
//  1. StructuredParserError type (structured_error.go):
//     - Error kind categorization (syntax, missing, unexpected, invalid, ambiguous)
//     - Expected vs actual token tracking
//     - Automatic block context inclusion
//     - Helpful suggestions for fixing errors
//     - Related positions for multi-part errors
//     - Parse phase tracking
//
//  2. Error creation methods:
//     - NewStructuredError(kind): Creates builder with fluent API
//     - NewUnexpectedTokenError(): Helper for common "expected X, got Y" errors
//     - NewMissingTokenError(): Helper for missing required tokens
//     - NewInvalidExpressionError(): Helper for invalid expressions
//
//  3. Integration:
//     - addStructuredError(err): Adds structured error to parser (auto-injects block context)
//     - Backward compatible: converts to legacy ParserError automatically
//     - Existing tests continue to work without modification
//
// Common patterns:
//
//  1. Missing keyword:
//     if !p.expectPeek(lexer.THEN) {
//     err := NewStructuredError(ErrKindMissing).
//     WithCode(ErrMissingThen).
//     WithMessage("expected 'then' after if condition").
//     WithPosition(p.cursor.Peek(1).Pos, p.cursor.Peek(1).Length()).
//     WithExpected(lexer.THEN).
//     WithActual(p.cursor.Peek(1).Type, p.cursor.Peek(1).Literal).
//     WithSuggestion("add 'then' keyword after the condition").
//     WithNote("DWScript if statements require: if <condition> then <statement>").
//     Build()
//     p.addStructuredError(err)
//     return nil
//     }
//
//  2. Invalid expression:
//     if stmt.Condition == nil {
//     err := NewStructuredError(ErrKindInvalid).
//     WithCode(ErrInvalidExpression).
//     WithMessage("expected condition after 'if'").
//     WithPosition(p.cursor.Current().Pos, p.cursor.Current().Length()).
//     WithExpectedString("boolean expression").
//     WithSuggestion("add a condition like 'x > 0' or 'flag = true'").
//     WithParsePhase("if statement condition").
//     Build()
//     p.addStructuredError(err)
//     return nil
//     }
//
//  3. Missing closing delimiter (with related position):
//     if !p.expectPeek(lexer.RBRACK) {
//     err := NewStructuredError(ErrKindMissing).
//     WithCode(ErrMissingRBracket).
//     WithMessage("expected ']' to close array index").
//     WithPosition(p.cursor.Peek(1).Pos, p.cursor.Peek(1).Length()).
//     WithExpected(lexer.RBRACK).
//     WithActual(p.cursor.Peek(1).Type, p.cursor.Peek(1).Literal).
//     WithSuggestion("add ']' to close the array index").
//     WithRelatedPosition(lbrackToken.Pos, "opening '[' here").
//     WithParsePhase("array index expression").
//     Build()
//     p.addStructuredError(err)
//     return nil
//     }
//
// Migration strategy:
//   - New code should use structured errors for better diagnostics
//   - Legacy addError() and addErrorWithContext() still work
//   - Gradually migrate existing error sites to structured errors
//   - See parseIfStatement(), parseWhileStatement(), parseArrayType() for examples
//
// Best practices:
//   - Use appropriate error kind (ErrKindMissing, ErrKindUnexpected, ErrKindInvalid)
//   - Always provide expected/actual values when applicable
//   - Add helpful suggestions that guide users to fix the error
//   - Include related positions for paired delimiters (parentheses, brackets, etc.)
//   - Set parse phase for better context ("array type", "if statement", etc.)
//   - Block context is auto-injected by addStructuredError() - no need to add manually
//
// ERROR-CONTEXT INTEGRATION (Phase 2.1.3):
//
// The parser automatically integrates ParseContext with structured errors for rich error messages.
//
// Automatic context capture:
//   - addStructuredError() auto-injects current block context if not explicitly set
//   - Context includes block type (begin, if, while, etc.) and start position
//   - Errors automatically show: "error message (in while block starting at line 5)"
//
// Context management:
//   - ParseContext tracks block nesting via PushBlock/PopBlock
//   - Context snapshots are saved/restored during speculative parsing
//   - Context flags (parsingPostCondition, etc.) are synchronized
//
// Example of automatic context in errors:
//
//	begin
//	  x := 10;
//	  while y < 10    // Missing 'do'
//	    z := 5;
//	end;
//
//	Error: "expected 'do' after while condition (in while block starting at line 3)"
//
// Nested blocks:
//   - Errors capture the INNERMOST block context
//   - Each error gets its own snapshot of the current context
//   - Context properly tracks nesting depth and block types
//
// Testing:
//   - See error_context_integration_test.go for comprehensive tests
//   - Tests cover: automatic capture, nested blocks, state persistence, multiple errors
//
// Migration examples:
//   - Variable declarations: statements.go (7 error sites)
//   - Control flow: control_flow.go (parseIfStatement, parseWhileStatement)
//   - Type parsing: types.go (parseArrayType)
//   - Expression parsing: expressions.go (parseOldExpression)
//
// PRATT PARSING (Core Architecture):
//
// The parser uses a Pratt parser (top-down operator precedence) for expressions.
// This provides elegant handling of operator precedence and associativity.
//
// Key concepts:
//
//  1. Prefix Parse Functions:
//     - Called when a token appears at the START of an expression
//     - Examples: literals (42, "hello"), unary operators (-x, not x), grouping ((expr))
//     - Registered via registerPrefix(tokenType, parseFn)
//
//  2. Infix Parse Functions:
//     - Called when a token appears BETWEEN expressions
//     - Examples: binary operators (x + y), function calls (foo()), member access (obj.field)
//     - Registered via registerInfix(tokenType, parseFn)
//     - Receive left expression as parameter, parse right side
//
//  3. Precedence Levels:
//     - Integer constants from LOWEST to MEMBER
//     - Higher number = higher precedence
//     - Determines how tightly operators bind
//     - Example: PRODUCT (5) > SUM (4), so 3+5*2 parses as 3+(5*2)
//
// The parseExpression(precedence) method:
//  1. Look up prefix function for current token
//  2. Parse prefix to get left expression
//  3. While peek token has higher precedence:
//     a. Look up infix function
//     b. Advance to infix operator
//     c. Parse infix (passing left expression)
//     d. Update left expression with result
//  4. Return final expression
//
// See docs/parser-architecture.md for detailed explanation.
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
	errors               []*ParserError
	parsingPostCondition bool           // True when parsing postconditions (for 'old' keyword)
	blockStack           []BlockContext // Stack of nested block contexts for error messages

	// ctx is the new structured parsing context (Task 2.1.2)
	// Consolidates scattered context flags and block stack into a single type.
	ctx *ParseContext

	// The parser uses the immutable cursor for token navigation.
	cursor *TokenCursor
}

// ParserState represents a HEAVYWEIGHT snapshot of the parser's complete state.
// It can be saved and restored to enable speculative parsing and backtracking
// with full error and context preservation.
//
// LIGHTWEIGHT vs HEAVYWEIGHT BACKTRACKING:
//
// Use ParserState (heavyweight) via saveState()/restoreState() when you need to:
//   - Save/restore full parser state (errors, context, lexer state)
//   - Backtrack across multiple parsing functions
//   - Try multiple parsing strategies speculatively
//   - Preserve or discard error states
//   - More comprehensive but slower than cursor.Mark()
//
// Use cursor.Mark()/ResetTo() (lightweight) when you need to:
//   - Save/restore cursor position only
//   - Backtrack within a single parsing function
//   - Minimal overhead (just 1 integer)
//   - No error state changes during backtracking
//
// Performance: ParserState copies significant state (errors slice, block stack, etc).
// Use it only when you need full state preservation, prefer cursor.Mark() otherwise.
//
// Example - try multiple parsing strategies:
//
//	state := p.saveState()
//	if !tryFirstStrategy(p) {
//	    p.restoreState(state)  // Discard errors and backtrack
//	    trySecondStrategy(p)
//	}
//
// ParserState represents a snapshot of the parser's state for speculative parsing.
type ParserState struct {
	lexerState           lexer.LexerState
	errors               []*ParserError
	parsingPostCondition bool
	blockStack           []BlockContext
	ctx                  *ParseContext // New structured context (Task 2.1.2)
	cursor               *TokenCursor  // Cursor position
}

// New creates a new Parser instance.
//
// This is a convenience wrapper around the ParserBuilder for the common case
// of creating a parser with default settings.
//
// For more control over parser configuration, use the builder pattern:
//
//	parser := NewParserBuilder(lexer).
//	    WithStrictMode(true).
//	    Build()
func New(l *lexer.Lexer) *Parser {
	return NewParserBuilder(l).Build()
}

// NewCursorParser creates a new Parser instance.
// Maintained for backward compatibility.
//
// Usage:
//
//	p := parser.NewCursorParser(lexer)
//	program := p.ParseProgram()
//
// This is a convenience wrapper around the ParserBuilder.
func NewCursorParser(l *lexer.Lexer) *Parser {
	p := NewParserBuilder(l).Build()

	// These functions take tokens explicitly as parameters instead of accessing parser state.
	//
	// Note: For now, these are adapters that wrap existing cursor functions.
	// The existing functions still access p.cursor internally, but the adapter pattern
	// allows us to establish the infrastructure for pure functional parsing.
	// Later tasks will refactor the actual functions to take tokens as parameters.

	// Prefix functions - wrap existing implementations
	p.registerPrefix(lexer.IDENT, func(tok lexer.Token) ast.Expression {
		return p.parseIdentifier()
	})
	p.registerPrefix(lexer.INT, func(tok lexer.Token) ast.Expression {
		return p.parseIntegerLiteral()
	})
	p.registerPrefix(lexer.FLOAT, func(tok lexer.Token) ast.Expression {
		return p.parseFloatLiteral()
	})
	p.registerPrefix(lexer.STRING, func(tok lexer.Token) ast.Expression {
		return p.parseStringLiteral()
	})
	p.registerPrefix(lexer.TRUE, func(tok lexer.Token) ast.Expression {
		return p.parseBooleanLiteral()
	})
	p.registerPrefix(lexer.FALSE, func(tok lexer.Token) ast.Expression {
		return p.parseBooleanLiteral()
	})

	p.registerPrefix(lexer.NIL, func(tok lexer.Token) ast.Expression {
		return p.parseNilLiteral()
	})
	p.registerPrefix(lexer.NULL, func(tok lexer.Token) ast.Expression {
		return p.parseNullIdentifier()
	})
	p.registerPrefix(lexer.UNASSIGNED, func(tok lexer.Token) ast.Expression {
		return p.parseUnassignedIdentifier()
	})
	p.registerPrefix(lexer.CHAR, func(tok lexer.Token) ast.Expression {
		return p.parseCharLiteral()
	})
	p.registerPrefix(lexer.MINUS, func(tok lexer.Token) ast.Expression {
		return p.parsePrefixExpression()
	})
	p.registerPrefix(lexer.PLUS, func(tok lexer.Token) ast.Expression {
		return p.parsePrefixExpression()
	})
	p.registerPrefix(lexer.NOT, func(tok lexer.Token) ast.Expression {
		return p.parsePrefixExpression()
	})
	p.registerPrefix(lexer.LPAREN, func(tok lexer.Token) ast.Expression {
		return p.parseGroupedExpression()
	})
	p.registerPrefix(lexer.LBRACK, func(tok lexer.Token) ast.Expression {
		return p.parseArrayLiteral()
	})

	p.registerPrefix(lexer.SELF, func(tok lexer.Token) ast.Expression {
		return p.parseSelfExpression()
	})
	p.registerPrefix(lexer.INHERITED, func(tok lexer.Token) ast.Expression {
		return p.parseInheritedExpression()
	})
	p.registerPrefix(lexer.NEW, func(tok lexer.Token) ast.Expression {
		return p.parseNewExpression()
	})
	p.registerPrefix(lexer.IF, func(tok lexer.Token) ast.Expression {
		return p.parseIfExpression()
	})
	p.registerPrefix(lexer.LAMBDA, func(tok lexer.Token) ast.Expression {
		return p.parseLambdaExpression()
	})
	p.registerPrefix(lexer.AT, func(tok lexer.Token) ast.Expression {
		return p.parseAddressOfExpression()
	})
	p.registerPrefix(lexer.DEFAULT, func(tok lexer.Token) ast.Expression {
		return p.parseDefaultExpression()
	})
	p.registerPrefix(lexer.OLD, func(tok lexer.Token) ast.Expression {
		return p.parseOldExpression()
	})

	// Note: Only functions with true cursor implementations are registered above.
	// When parseExpression encounters a token type without a cursor prefix function,

	// Infix functions - wrap existing cursor implementations
	p.registerInfix(lexer.PLUS, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.MINUS, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.ASTERISK, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.SLASH, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.DIV, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.MOD, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.SHL, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.SHR, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.SAR, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.EQ, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.NOT_EQ, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.LESS, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.GREATER, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.LESS_EQ, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.GREATER_EQ, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.AND, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.OR, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.XOR, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.IN, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})
	p.registerInfix(lexer.QUESTION_QUESTION, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseInfixExpression(left)
	})

	p.registerInfix(lexer.LPAREN, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseCallExpression(left)
	})
	p.registerInfix(lexer.DOT, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseMemberAccess(left)
	})
	p.registerInfix(lexer.LBRACK, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseIndexExpression(left)
	})

	p.registerInfix(lexer.IS, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseIsExpression(left)
	})
	p.registerInfix(lexer.AS, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseAsExpression(left)
	})
	p.registerInfix(lexer.IMPLEMENTS, func(left ast.Expression, tok lexer.Token) ast.Expression {
		return p.parseImplementsExpression(left)
	})

	return p
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
	if p.cursor != nil {
		p.cursor = p.cursor.Advance()
	}
}

// curTokenIs checks if the current token is of the given type.
func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.cursor.Current().Type == t
}

// peekTokenIs checks if the peek token is of the given type.
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.cursor.Peek(1).Type == t
}

// peek provides N-token lookahead using the lexer's Peek() method.
// n=0 returns the token after peekToken, n=1 returns 2 tokens ahead, etc.
// This is a convenience wrapper around p.l.Peek(n) for cleaner syntax.
//
// Example usage:
//   - To look 1 token ahead of curToken: p.peekTokenIs(lexer.IDENT)
//   - To look 2 tokens ahead of curToken: p.peek(0).Type == lexer.COLON
//   - To look 3 tokens ahead of curToken: p.peek(1).Type == lexer.ASSIGN
func (p *Parser) peek(n int) lexer.Token {
	return p.l.Peek(n)
}

// peekAhead is an alternative helper that looks N tokens ahead from curToken.
// n=1 returns peekToken, n=2 returns the token after peekToken, etc.
// This provides a more intuitive interface where n represents "tokens ahead from curToken".
//
// Example usage:
//   - To look 1 token ahead: p.peekAhead(1) (same as cursor.Peek(1))
//   - To look 2 tokens ahead: p.peekAhead(2) (same as cursor.Peek(2))
//   - To look 3 tokens ahead: p.peekAhead(3) (same as cursor.Peek(3))
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
	p.peekError(t)
	return false
}

// isIdentifierToken checks if a token type can be used as an identifier.
// This includes IDENT and contextual keywords like STEP that are keywords in
// specific contexts (for loops) but can be used as variable names elsewhere.
// Also includes SELF which can be the target of member assignments (Self.field := value).
func (p *Parser) isIdentifierToken(t lexer.TokenType) bool {
	return t == lexer.IDENT || t == lexer.STEP || t == lexer.SELF
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

// registerInfix registers an infix parse function for a token type.

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// peekPrecedence returns the precedence of the peek token.
func (p *Parser) peekPrecedence() int {
	tokenType := p.cursor.Peek(1).Type

	if prec, ok := precedences[tokenType]; ok {
		return prec
	}
	return LOWEST
}

// curPrecedence returns the precedence of the current token.
func (p *Parser) curPrecedence() int {
	tokenType := p.cursor.Current().Type

	if prec, ok := precedences[tokenType]; ok {
		return prec
	}
	return LOWEST
}

// getPrecedence returns the precedence of a token type without relying on parser state.
// needs to work with tokens passed as parameters rather than parser fields.
//
// Unlike curPrecedence() and peekPrecedence() which access p.cursor.Current() and p.cursor.Peek(1),
// this function is stateless and can be used in pure functional parsing contexts.
//
// Returns LOWEST if the token type is not in the precedences map.
func getPrecedence(tokenType lexer.TokenType) int {
	if prec, ok := precedences[tokenType]; ok {
		return prec
	}
	return LOWEST
}

// saveState captures the current parser state for later restoration.
// This enables speculative parsing: try one approach, and if it fails,
// restore the state and try a different approach without leaving errors behind.
//
// Example usage:
//
//	state := p.saveState()
//	if result := p.tryParseAsType(); result != nil {
//	    return result
//	}
//	p.restoreState(state)  // Failed, backtrack
//	return p.parseAsExpression()
func (p *Parser) saveState() ParserState {
	// Make a copy of errors slice to avoid sharing the backing array
	errorsCopy := make([]*ParserError, len(p.errors))
	copy(errorsCopy, p.errors)

	// Make a deep copy of blockStack
	blockStackCopy := make([]BlockContext, len(p.blockStack))
	copy(blockStackCopy, p.blockStack)

	return ParserState{
		errors:               errorsCopy,
		lexerState:           p.l.SaveState(),
		parsingPostCondition: p.parsingPostCondition,
		blockStack:           blockStackCopy,
		ctx:                  p.ctx.Snapshot(), // Save context snapshot (Task 2.1.2)
		cursor:               p.cursor,
	}
}

// restoreState restores the parser to a previously saved state.
// This undoes all parser and lexer changes made since saveState() was called.
// It's used after speculative parsing fails to cleanly backtrack.
func (p *Parser) restoreState(state ParserState) {
	p.errors = state.errors
	p.parsingPostCondition = state.parsingPostCondition
	p.blockStack = state.blockStack
	p.l.RestoreState(state.lexerState)
	// Restore context (Task 2.1.2)
	// This also restores parsingPostCondition in the context
	p.ctx.Restore(state.ctx)
	p.cursor = state.cursor
}

// pushBlockContext pushes a new block context onto the stack.
// This is used to track nested blocks for better error messages.
// Adapter method: delegates to context and synchronizes old field for backward compatibility.
func (p *Parser) pushBlockContext(blockType string, startPos lexer.Position) {
	// Update new context (Task 2.1.2)
	p.ctx.PushBlock(blockType, startPos)

	// Synchronize old field for backward compatibility
	p.blockStack = append(p.blockStack, BlockContext{
		BlockType: blockType,
		StartPos:  startPos,
		StartLine: startPos.Line,
	})
}

// popBlockContext pops the most recent block context from the stack.
// Call this when exiting a block to maintain proper nesting.
// Adapter method: delegates to context and synchronizes old field for backward compatibility.
func (p *Parser) popBlockContext() {
	// Update new context (Task 2.1.2)
	p.ctx.PopBlock()

	// Synchronize old field for backward compatibility
	if len(p.blockStack) > 0 {
		p.blockStack = p.blockStack[:len(p.blockStack)-1]
	}
}

// currentBlockContext returns the current block context, if any.
// Returns nil if no block is currently being parsed.
// Delegates to ParseContext for current block information.
func (p *Parser) currentBlockContext() *BlockContext {
	// Delegate to new context (Task 2.1.2)
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

// synchronize performs panic-mode error recovery by advancing to a synchronization point.
// It skips tokens until it finds one that's likely to be a safe point to resume parsing.
//
// Parameters:
//   - syncTokens: specific tokens to synchronize on (in addition to statement starters/block closers)
//
// The synchronize method will stop at:
//  1. Any token in syncTokens
//  2. Statement starters (if/while/for/begin/etc.)
//  3. Block closers (end/until/else/etc.)
//  4. EOF (to prevent infinite loops)
//
// Example usage:
//
//	if !p.expectPeek(lexer.THEN) {
//	    p.addError("expected 'then' after if condition", ErrMissingThen)
//	    p.synchronize([]lexer.TokenType{lexer.THEN, lexer.ELSE, lexer.END})
//	    return nil
//	}

// synchronize performs panic-mode error recovery using cursor.
// It advances the cursor until it finds a synchronization point.
//
// Parameters:
//   - syncTokens: specific tokens to synchronize on (in addition to statement starters/block closers)
//
// Returns the new cursor position and whether a sync token was found.
func (p *Parser) synchronize(syncTokens []lexer.TokenType) bool {
	// Build a map of all synchronization tokens for fast lookup
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

// addErrorWithContext adds an error with additional context from the block stack.
// This provides better error messages by including information about which block
// the error occurred in.
//
// Example output: "expected 'end' to close 'begin' block starting at line 10"
func (p *Parser) addErrorWithContext(msg string, code string) {
	if ctx := p.currentBlockContext(); ctx != nil {
		msg = fmt.Sprintf("%s (in %s block starting at line %d)", msg, ctx.BlockType, ctx.StartLine)
	}
	p.addError(msg, code)
}

// endPosFromToken calculates the end position of a token.
// This is a helper function to populate EndPos fields in AST nodes.
func (p *Parser) endPosFromToken(tok lexer.Token) lexer.Position {
	pos := tok.Pos
	pos.Column += tok.Length()
	pos.Offset += tok.Length()
	return pos
}

// LIST PARSING HELPERS (Task 2.5)
//
// These helpers reduce code duplication by providing reusable patterns for
// parsing separated lists. Common use cases include:
//   - Comma-separated expression lists: (expr1, expr2, expr3)
//   - Semicolon-separated parameter groups: (x: Integer; y: String)
//   - Field lists, argument lists, etc.
//
// Design principles:
//   - Flexible separator support (single or multiple separator types)
//   - Optional trailing separator handling
//   - Proper error recovery
//   - Works with callbacks to handle different item types
//
// The helpers handle the common looping and separator logic, while callers
// provide item-specific parsing via callbacks.

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

// parseSeparatedList is a generic helper for parsing lists of items separated by delimiters.
//
// This function handles the common pattern of:
//  1. Check if list is empty (curToken is terminator)
//  2. Parse first item
//  3. While peekToken is a separator:
//     - Consume separator
//     - Check for trailing separator
//     - Parse next item
//  4. Expect terminator (if RequireTerminator is true)
//
// Parameters:
//   - opts: Configuration options (separators, terminator, etc.)
//   - parseItem: Callback to parse one item. Returns true if successful, false on error.
//     The callback should NOT consume trailing separators or terminators.
//
// Returns:
//   - itemCount: Number of items successfully parsed
//   - success: true if parsing completed successfully, false on error
//
// Token position on entry:
//   - curToken should be the first item OR the terminator (for empty lists)
//
// Token position on exit:
//   - If RequireTerminator is true: curToken is the terminator
//   - If RequireTerminator is false: curToken is the last item, peekToken is first non-separator
//
// Example usage (comma-separated expressions):
//
//	opts := ListParseOptions{
//	    Separators:             []lexer.TokenType{lexer.COMMA},
//	    Terminator:             lexer.RPAREN,
//	    AllowTrailingSeparator: true,
//	    AllowEmpty:             true,
//	    RequireTerminator:      true,
//	}
//	exprs := []ast.Expression{}
//	count, ok := p.parseSeparatedList(opts, func() bool {
//	    expr := p.parseExpression(LOWEST)
//	    if expr == nil {
//	        return false
//	    }
//	    exprs = append(exprs, expr)
//	    return true
//	})
//	if !ok {
//	    return nil
//	}
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

// parseSeparatedListBeforeStart is a variant of parseSeparatedList for when
// the current token is BEFORE the list (e.g., at the opening paren).
//
// This helper:
//  1. Checks if peekToken is terminator (empty list)
//  2. Advances to first item (nextToken)
//  3. Calls parseSeparatedList with remaining logic
//
// Token position on entry:
//   - curToken should be BEFORE the first item (e.g., at LPAREN)
//   - peekToken should be first item OR terminator
//
// Token position on exit:
//   - If RequireTerminator is true: curToken is the terminator
//   - If RequireTerminator is false: curToken is the last item
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

// isVarDeclBlock checks if a BlockStatement was created by parseVarDeclaration()
// to wrap multiple var declarations. These should be unwrapped to avoid extra scope nesting.
// We distinguish them from begin...end blocks by checking the token type:
// - parseVarDeclaration() uses VAR token
// - parseBlockStatement() uses BEGIN token
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

// parseFieldInitializer parses an optional field initializer (= Value or := Value).
// Returns the initialization expression if present, or nil if not.
// Should be called when curToken is the type token, and peekToken might be '=' or ':='.
// PRE: cursor is last token of type annotation
// POST: cursor is last token of initialization expression if present; otherwise unchanged
func (p *Parser) parseFieldInitializer(fieldNames []*ast.Identifier) ast.Expression {
	// Check for initialization (= Value or := Value)
	// DWScript uses '=' for field initializers: Field : String = 'hello';
	// Also support ':=' for compatibility
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
