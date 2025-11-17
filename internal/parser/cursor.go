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

// ============================================================================
// Declarative Lookahead Utilities (Task 2.6.1)
// ============================================================================

// LookAhead checks if the tokens starting from the current position match
// the given pattern of token types. This provides a declarative way to check
// for token patterns without manually peeking at each position.
//
// Returns true if all tokens in the pattern match, false otherwise.
//
// Example:
//
//	// Check if current token is IDENT followed by COLON
//	if cursor.LookAhead(token.IDENT, token.COLON) {
//	    // Pattern matched
//	}
//
//	// Check for const declaration pattern: IDENT followed by COLON or EQ
//	if cursor.LookAhead(token.IDENT, token.COLON) ||
//	   cursor.LookAhead(token.IDENT, token.EQ) {
//	    // Matches typed or untyped const
//	}
func (c *TokenCursor) LookAhead(pattern ...token.TokenType) bool {
	if len(pattern) == 0 {
		return true
	}

	// Check each position in the pattern
	for i, expectedType := range pattern {
		tok := c.Peek(i)
		if tok.Type != expectedType {
			return false
		}
	}

	return true
}

// LookAheadFunc checks if the tokens starting from the current position
// satisfy the given predicate function. This is a more flexible version of
// LookAhead that allows custom matching logic.
//
// The predicate receives the token at each position (0, 1, 2, ...) and
// should return true to continue checking or false to fail the match.
// If the predicate returns true for all positions up to maxDistance,
// the function returns true.
//
// Example:
//
//	// Check if next 3 tokens are all identifiers
//	cursor.LookAheadFunc(3, func(tok token.Token) bool {
//	    return tok.Type == token.IDENT
//	})
func (c *TokenCursor) LookAheadFunc(maxDistance int, predicate func(token.Token) bool) bool {
	for i := 0; i < maxDistance; i++ {
		tok := c.Peek(i)
		if !predicate(tok) {
			return false
		}
	}
	return true
}

// ScanUntil scans forward from the current position until the predicate
// returns true or EOF is reached. Returns the distance to the matching token
// and whether it was found.
//
// This is useful for finding matching delimiters or scanning to a specific
// context without manually looping.
//
// The predicate receives each token starting from Peek(0) and should return
// true when the desired token is found.
//
// Returns:
//   - distance: Number of positions to advance to reach the matching token
//   - found: true if the predicate matched, false if EOF was reached
//
// Example:
//
//	// Find the next semicolon
//	if distance, found := cursor.ScanUntil(func(t token.Token) bool {
//	    return t.Type == token.SEMICOLON
//	}); found {
//	    // Semicolon found at position 'distance'
//	}
//
//	// Find matching END for BEGIN (simplified - real code needs nesting count)
//	if distance, found := cursor.ScanUntil(func(t token.Token) bool {
//	    return t.Type == token.END
//	}); found {
//	    cursor = cursor.AdvanceN(distance)
//	}
func (c *TokenCursor) ScanUntil(predicate func(token.Token) bool) (distance int, found bool) {
	for i := 0; ; i++ {
		tok := c.Peek(i)

		// Check if we've reached EOF
		if tok.Type == token.EOF {
			return i, false
		}

		// Check if predicate matches
		if predicate(tok) {
			return i, true
		}
	}
}

// ScanUntilAny scans forward until any of the given token types is found.
// This is a convenience wrapper around ScanUntil for common token matching.
//
// Returns:
//   - distance: Number of positions to advance to reach the matching token
//   - found: true if any of the token types matched, false if EOF was reached
//   - matchedType: The token type that matched (or ILLEGAL if not found)
//
// Example:
//
//	// Find next semicolon or end keyword
//	if distance, found, tokType := cursor.ScanUntilAny(token.SEMICOLON, token.END); found {
//	    // Found token of type 'tokType' at distance 'distance'
//	}
func (c *TokenCursor) ScanUntilAny(types ...token.TokenType) (distance int, found bool, matchedType token.TokenType) {
	dist, ok := c.ScanUntil(func(t token.Token) bool {
		for _, tokType := range types {
			if t.Type == tokType {
				matchedType = tokType
				return true
			}
		}
		return false
	})
	return dist, ok, matchedType
}

// FindNext searches for the next occurrence of the given token type within
// maxDistance positions. Returns the distance to the token and whether it
// was found.
//
// This is more efficient than ScanUntil when you know the maximum distance
// to search, as it stops early.
//
// Returns:
//   - distance: Number of positions to advance to reach the token (0 = current position)
//   - found: true if the token was found within maxDistance, false otherwise
//
// Example:
//
//	// Find SEMICOLON within next 10 tokens
//	if distance, found := cursor.FindNext(token.SEMICOLON, 10); found {
//	    // Semicolon found at position 'distance'
//	}
//
//	// Check if current token is the target (distance 0)
//	if distance, found := cursor.FindNext(token.THEN, 5); found && distance == 0 {
//	    // Current token is THEN
//	}
func (c *TokenCursor) FindNext(tokenType token.TokenType, maxDistance int) (distance int, found bool) {
	for i := 0; i < maxDistance; i++ {
		tok := c.Peek(i)

		// Stop at EOF
		if tok.Type == token.EOF {
			return i, false
		}

		// Check if we found the target
		if tok.Type == tokenType {
			return i, true
		}
	}

	return maxDistance, false
}

// FindNextAny searches for the next occurrence of any of the given token types
// within maxDistance positions.
//
// Returns:
//   - distance: Number of positions to advance to reach the token
//   - found: true if any token was found within maxDistance, false otherwise
//   - matchedType: The token type that matched (or ILLEGAL if not found)
//
// Example:
//
//	// Find next comma or semicolon within 10 tokens
//	if distance, found, tokType := cursor.FindNextAny(10, token.COMMA, token.SEMICOLON); found {
//	    // Found 'tokType' at position 'distance'
//	}
func (c *TokenCursor) FindNextAny(maxDistance int, types ...token.TokenType) (distance int, found bool, matchedType token.TokenType) {
	for i := 0; i < maxDistance; i++ {
		tok := c.Peek(i)

		// Stop at EOF
		if tok.Type == token.EOF {
			return i, false, token.ILLEGAL
		}

		// Check if we found any of the targets
		for _, tokType := range types {
			if tok.Type == tokType {
				return i, true, tokType
			}
		}
	}

	return maxDistance, false, token.ILLEGAL
}
