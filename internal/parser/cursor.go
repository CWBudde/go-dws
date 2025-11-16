package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/token"
)

// TokenCursor provides an immutable cursor abstraction over a stream of tokens.
// It replaces the mutable parser state (curToken/peekToken) with an immutable
// navigation interface that supports backtracking and lookahead.
//
// Key features:
//   - Immutable: All operations return new cursor instances
//   - Backtracking: Save/restore cursor position via Mark/ResetTo
//   - Lookahead: Peek arbitrary distances ahead
//   - Convenience: Is/IsAny/Expect methods for common patterns
//
// Design philosophy:
//   - Zero manual nextToken() calls in parsing code
//   - Explicit cursor state rather than hidden mutable state
//   - Composable navigation operations
//   - Type-safe token matching
//
// Usage example:
//
//	cursor := NewTokenCursor(lexer)
//	if cursor.Is(token.IF) {
//	    cursor = cursor.Advance()  // Move to next token
//	    condition := parseExpression(cursor)
//	    cursor = cursor.Expect(token.THEN)  // Advance if THEN, error otherwise
//	}
type TokenCursor struct {
	lexer   *lexer.Lexer
	current token.Token
	tokens  []token.Token // Buffered tokens for backtracking
	index   int           // Current position in buffered tokens
}

// NewTokenCursor creates a new TokenCursor from a lexer.
// The cursor starts at the first token in the stream.
func NewTokenCursor(l *lexer.Lexer) *TokenCursor {
	// Read the first token
	firstToken := l.NextToken()
	return &TokenCursor{
		lexer:   l,
		current: firstToken,
		tokens:  []token.Token{firstToken},
		index:   0,
	}
}

// Current returns the token at the current cursor position.
// This replaces the old `p.curToken` field.
func (c *TokenCursor) Current() token.Token {
	return c.current
}

// Peek returns the token N positions ahead of the current position.
// Peek(0) returns the current token (same as Current()).
// Peek(1) returns the next token (replaces old `p.peekToken`).
// Peek(2) returns the token after next, etc.
//
// This method buffers tokens as needed to support arbitrary lookahead.
func (c *TokenCursor) Peek(n int) token.Token {
	if n < 0 {
		// Invalid lookahead - return current
		return c.current
	}

	targetIndex := c.index + n

	// Ensure we have enough tokens buffered
	for len(c.tokens) <= targetIndex {
		nextTok := c.lexer.NextToken()
		c.tokens = append(c.tokens, nextTok)
		if nextTok.Type == token.EOF {
			// Don't fetch beyond EOF
			break
		}
	}

	// Return the token at the target index, or EOF if beyond buffer
	if targetIndex < len(c.tokens) {
		return c.tokens[targetIndex]
	}
	// Return last token (should be EOF)
	return c.tokens[len(c.tokens)-1]
}

// Advance returns a new cursor positioned at the next token.
// The original cursor is unchanged (immutable operation).
// This replaces the old `p.nextToken()` method.
//
// Example:
//
//	cursor = cursor.Advance()  // Move to next token
func (c *TokenCursor) Advance() *TokenCursor {
	return c.AdvanceN(1)
}

// AdvanceN returns a new cursor positioned N tokens ahead.
// The original cursor is unchanged (immutable operation).
// If N <= 0, returns the same cursor.
//
// Example:
//
//	cursor = cursor.AdvanceN(3)  // Skip 3 tokens
func (c *TokenCursor) AdvanceN(n int) *TokenCursor {
	if n <= 0 {
		return c
	}

	// Ensure we have tokens buffered
	c.Peek(n)

	newIndex := c.index + n
	if newIndex >= len(c.tokens) {
		// Can't advance past EOF - stop at last token
		newIndex = len(c.tokens) - 1
	}

	return &TokenCursor{
		lexer:   c.lexer,
		current: c.tokens[newIndex],
		tokens:  c.tokens,
		index:   newIndex,
	}
}

// Skip advances the cursor if the current token matches the given type.
// Returns (new cursor, true) if matched and advanced.
// Returns (original cursor, false) if not matched.
//
// This is useful for optional tokens:
//
//	if newCursor, ok := cursor.Skip(token.SEMICOLON); ok {
//	    cursor = newCursor
//	}
func (c *TokenCursor) Skip(t token.TokenType) (*TokenCursor, bool) {
	if c.current.Type == t {
		return c.Advance(), true
	}
	return c, false
}

// SkipAny advances the cursor if the current token matches any of the given types.
// Returns (new cursor, true, matched type) if matched and advanced.
// Returns (original cursor, false, ILLEGAL) if not matched.
//
// Example:
//
//	if newCursor, ok, _ := cursor.SkipAny(token.SEMICOLON, token.COMMA); ok {
//	    cursor = newCursor
//	}
func (c *TokenCursor) SkipAny(types ...token.TokenType) (*TokenCursor, bool, token.TokenType) {
	for _, t := range types {
		if c.current.Type == t {
			return c.Advance(), true, t
		}
	}
	return c, false, token.ILLEGAL
}

// Is checks if the current token matches the given type.
// This replaces the old `p.curTokenIs(t)` method.
func (c *TokenCursor) Is(t token.TokenType) bool {
	return c.current.Type == t
}

// IsAny checks if the current token matches any of the given types.
// Returns (true, matched type) if matched, (false, ILLEGAL) otherwise.
func (c *TokenCursor) IsAny(types ...token.TokenType) (bool, token.TokenType) {
	for _, t := range types {
		if c.current.Type == t {
			return true, t
		}
	}
	return false, token.ILLEGAL
}

// PeekIs checks if the token N positions ahead matches the given type.
// PeekIs(1, token.THEN) replaces the old `p.peekTokenIs(token.THEN)`.
func (c *TokenCursor) PeekIs(n int, t token.TokenType) bool {
	return c.Peek(n).Type == t
}

// PeekIsAny checks if the token N positions ahead matches any of the given types.
func (c *TokenCursor) PeekIsAny(n int, types ...token.TokenType) (bool, token.TokenType) {
	peekType := c.Peek(n).Type
	for _, t := range types {
		if peekType == t {
			return true, t
		}
	}
	return false, token.ILLEGAL
}

// Expect advances the cursor if the current token matches the given type.
// Returns (new cursor, true) if matched and advanced.
// Returns (original cursor, false) if not matched.
//
// This replaces the old `p.expectPeek(t)` pattern, but without adding errors.
// The caller should handle error reporting as needed.
//
// Example:
//
//	if newCursor, ok := cursor.Expect(token.THEN); ok {
//	    cursor = newCursor
//	} else {
//	    // Report error
//	}
func (c *TokenCursor) Expect(t token.TokenType) (*TokenCursor, bool) {
	return c.Skip(t)
}

// ExpectAny advances the cursor if the current token matches any of the given types.
// Returns (new cursor, true, matched type) if matched and advanced.
// Returns (original cursor, false, ILLEGAL) if not matched.
func (c *TokenCursor) ExpectAny(types ...token.TokenType) (*TokenCursor, bool, token.TokenType) {
	return c.SkipAny(types...)
}

// Mark represents a saved cursor position that can be restored later.
// This enables backtracking and speculative parsing.
type Mark struct {
	index int
}

// Mark saves the current cursor position for later restoration.
// Use this for speculative parsing or backtracking.
//
// Example:
//
//	mark := cursor.Mark()
//	if !tryParsePattern(cursor) {
//	    cursor = cursor.ResetTo(mark)  // Backtrack
//	}
func (c *TokenCursor) Mark() Mark {
	return Mark{index: c.index}
}

// ResetTo returns a new cursor positioned at the given mark.
// This enables backtracking to a previously saved position.
// The original cursor is unchanged (immutable operation).
//
// Example:
//
//	mark := cursor.Mark()
//	cursor = cursor.Advance()
//	cursor = cursor.ResetTo(mark)  // Back to original position
func (c *TokenCursor) ResetTo(mark Mark) *TokenCursor {
	if mark.index < 0 || mark.index >= len(c.tokens) {
		// Invalid mark - return current cursor
		return c
	}

	return &TokenCursor{
		lexer:   c.lexer,
		current: c.tokens[mark.index],
		tokens:  c.tokens,
		index:   mark.index,
	}
}

// Clone returns a deep copy of the cursor.
// This is useful when you need to pass a cursor to a function
// that might modify it, but you want to preserve the original.
//
// Note: Since cursors are immutable, Clone() is rarely needed.
// Most operations return new cursors automatically.
func (c *TokenCursor) Clone() *TokenCursor {
	// Since we share the token buffer, we only need a shallow copy
	return &TokenCursor{
		lexer:   c.lexer,
		current: c.current,
		tokens:  c.tokens, // Shared buffer is fine (tokens are immutable)
		index:   c.index,
	}
}

// IsEOF checks if the current token is EOF.
// This is a common check, so we provide a convenience method.
func (c *TokenCursor) IsEOF() bool {
	return c.current.Type == token.EOF
}

// Position returns the position of the current token.
// This is useful for error reporting.
func (c *TokenCursor) Position() token.Position {
	return c.current.Pos
}

// Length returns the length of the current token.
// This is useful for error reporting.
func (c *TokenCursor) Length() int {
	return c.current.Length()
}
