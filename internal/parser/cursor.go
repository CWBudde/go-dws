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
	// Pre-allocate tokens slice with capacity to reduce reallocations
	// Most expressions have 20-50 tokens, so 32 is a good starting point
	tokens := make([]token.Token, 1, 32)
	tokens[0] = firstToken
	return &TokenCursor{
		lexer:   l,
		current: firstToken,
		tokens:  tokens,
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
	// Grow slice capacity aggressively to reduce reallocations
	if targetIndex >= len(c.tokens) {
		// Calculate how many tokens we need to fetch
		tokensNeeded := targetIndex - len(c.tokens) + 1

		// If we need to grow significantly, pre-allocate extra capacity
		if targetIndex >= cap(c.tokens) {
			// Grow by 1.5x or to target+16, whichever is larger
			newCap := max(targetIndex+16, cap(c.tokens)*3/2)
			newTokens := make([]token.Token, len(c.tokens), newCap)
			copy(newTokens, c.tokens)
			c.tokens = newTokens
		}

		// Fetch the needed tokens
		for i := 0; i < tokensNeeded; i++ {
			nextTok := c.lexer.NextToken()
			c.tokens = append(c.tokens, nextTok)
			if nextTok.Type == token.EOF {
				// Don't fetch beyond EOF
				break
			}
		}
	}

	// Return the token at the target index, or EOF if beyond buffer
	if targetIndex < len(c.tokens) {
		return c.tokens[targetIndex]
	}
	// Return last token (should be EOF)
	return c.tokens[len(c.tokens)-1]
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

// Mark represents a lightweight saved cursor position that can be restored later.
// This enables backtracking and speculative parsing with minimal overhead.
//
// LIGHTWEIGHT vs HEAVYWEIGHT MARKS:
//
// Use Mark (lightweight) when you need to:
//   - Save/restore cursor position only
//   - Backtrack within a single parsing function
//   - Minimal state (just 1 integer - very fast)
//   - No error state or parser context needed
//
// Use Parser.saveState()/restoreState() (heavyweight) when you need to:
//   - Save/restore full parser state (errors, context, lexer state)
//   - Backtrack across multiple parsing functions
//   - Speculative parsing that might add errors
//   - More state saved (slower, but comprehensive)
//
// Performance: Mark is extremely lightweight (just copies one int).
// Use it liberally for cursor-level backtracking.
type Mark struct {
	index int
}

// Mark saves the current cursor position for later restoration.
// This is a LIGHTWEIGHT operation that only saves the cursor position (1 integer).
// For full parser state backtracking, use Parser.saveState() instead.
//
// Use this for speculative parsing or backtracking at the cursor level.
//
// Example - try parsing a pattern, backtrack if it fails:
//
//	mark := cursor.Mark()
//	if !tryParsePattern(cursor) {
//	    cursor = cursor.ResetTo(mark)  // Backtrack to saved position
//	}
//
// Example - lookahead to check pattern without consuming:
//
//	mark := cursor.Mark()
//	pattern := scanPattern(cursor)
//	cursor = cursor.ResetTo(mark)  // Restore position after lookahead
//	if pattern.matches {
//	    // Now parse for real
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

// LookAhead scans forward from the current position to find a token matching
// the given predicate. Returns the matching token, its distance from current
// position, and true if found. Returns (zero token, 0, false) if not found
// before EOF.
//
// This makes lookahead declarative instead of imperative.
//
// Example - find next closing paren:
//
//	tok, distance, found := cursor.LookAhead(func(t token.Token) bool {
//	    return t.Type == token.RPAREN
//	})
//
// Example - find next statement terminator:
//
//	tok, distance, found := cursor.LookAhead(func(t token.Token) bool {
//	    return t.Type == token.SEMICOLON || t.Type == token.END
//	})
func (c *TokenCursor) LookAhead(predicate func(token.Token) bool) (token.Token, int, bool) {
	// Scan forward up to a reasonable limit (100 tokens)
	// This prevents infinite loops on malformed input
	const maxLookahead = 100

	for distance := 0; distance < maxLookahead; distance++ {
		tok := c.Peek(distance)

		// Stop at EOF
		if tok.Type == token.EOF {
			return token.Token{}, 0, false
		}

		// Check if predicate matches
		if predicate(tok) {
			return tok, distance, true
		}
	}

	return token.Token{}, 0, false
}

// ScanUntil collects tokens from the current position until the stop predicate
// returns true. Returns the collected tokens (not including the stop token).
// If EOF is reached before the stop condition, returns all tokens up to EOF.
//
// This is useful for gathering tokens within a delimited region.
//
// Example - collect tokens until semicolon:
//
//	tokens := cursor.ScanUntil(func(t token.Token) bool {
//	    return t.Type == token.SEMICOLON
//	})
//
// Example - collect tokens until end of parameter list:
//
//	tokens := cursor.ScanUntil(func(t token.Token) bool {
//	    return t.Type == token.RPAREN || t.Type == token.EOF
//	})
func (c *TokenCursor) ScanUntil(stop func(token.Token) bool) []token.Token {
	const maxScan = 100
	collected := make([]token.Token, 0, 16)

	for i := 0; i < maxScan; i++ {
		tok := c.Peek(i)

		// Stop at EOF
		if tok.Type == token.EOF {
			break
		}

		// Check stop condition
		if stop(tok) {
			break
		}

		collected = append(collected, tok)
	}

	return collected
}

// FindNext searches forward for the next occurrence of the given token type.
// Returns the distance to the token and true if found, or (0, false) if not
// found before EOF.
//
// This is a convenience wrapper around LookAhead for the common case of
// searching for a specific token type.
//
// Example - find distance to next comma:
//
//	distance, found := cursor.FindNext(token.COMMA)
//	if found {
//	    cursor = cursor.AdvanceN(distance)
//	}
//
// Example - check if there's a semicolon in the next 5 tokens:
//
//	if distance, found := cursor.FindNext(token.SEMICOLON); found && distance < 5 {
//	    // semicolon is nearby
//	}
func (c *TokenCursor) FindNext(t token.TokenType) (int, bool) {
	_, distance, found := c.LookAhead(func(tok token.Token) bool {
		return tok.Type == t
	})
	return distance, found
}
