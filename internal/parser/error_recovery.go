package parser

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ErrorRecovery provides centralized error recovery functionality for the parser.
// It encapsulates synchronization logic, error reporting patterns, and recovery strategies.
//
// Design principles:
//   - Panic-mode error recovery: skip tokens until reaching a "safe" synchronization point
//   - Multiple error reporting: continue parsing after errors to report multiple issues
//   - Context-aware error messages: include block context information
//   - Suggestion-based recovery: provide actionable hints to the user
//
// Usage:
//
//	recovery := NewErrorRecovery(parser)
//	if !parser.expectPeek(lexer.THEN) {
//	    recovery.AddExpectError(lexer.THEN, "after if condition")
//	    recovery.SynchronizeOn(lexer.THEN, lexer.ELSE, lexer.END)
//	    return nil
//	}
type ErrorRecovery struct {
	parser *Parser // Reference to the parser for error reporting and token access
}

// NewErrorRecovery creates a new ErrorRecovery instance.
func NewErrorRecovery(p *Parser) *ErrorRecovery {
	return &ErrorRecovery{
		parser: p,
	}
}

// SynchronizationSet defines a set of tokens for error recovery.
type SynchronizationSet int

const (
	// SyncStatementStarters synchronizes on tokens that can start a statement
	SyncStatementStarters SynchronizationSet = iota

	// SyncBlockClosers synchronizes on tokens that close blocks
	SyncBlockClosers

	// SyncDeclarationStarters synchronizes on tokens that can start a declaration
	SyncDeclarationStarters

	// SyncAll synchronizes on all common recovery points
	SyncAll
)

// GetSyncTokens returns the synchronization tokens for a given set.
func (s SynchronizationSet) GetSyncTokens() []lexer.TokenType {
	switch s {
	case SyncStatementStarters:
		return statementStarters
	case SyncBlockClosers:
		return blockClosers
	case SyncDeclarationStarters:
		return declarationStarters
	case SyncAll:
		// Combine all sets
		all := make([]lexer.TokenType, 0)
		all = append(all, statementStarters...)
		all = append(all, blockClosers...)
		all = append(all, declarationStarters...)
		return all
	default:
		return nil
	}
}

// SynchronizeOn performs panic-mode error recovery by advancing to a synchronization point.
// It skips tokens until it finds one of the specified tokens, a statement starter,
// or a block closer.
//
// Parameters:
//   - tokens: specific tokens to synchronize on
//
// Returns:
//   - true if synchronization succeeded (found a sync token)
//   - false if reached EOF without finding a sync token
func (er *ErrorRecovery) SynchronizeOn(tokens ...lexer.TokenType) bool {
	return er.parser.synchronize(tokens)
}

// SynchronizeOnSet synchronizes using a predefined synchronization set.
func (er *ErrorRecovery) SynchronizeOnSet(set SynchronizationSet, additionalTokens ...lexer.TokenType) bool {
	syncTokens := set.GetSyncTokens()
	allTokens := make([]lexer.TokenType, 0, len(syncTokens)+len(additionalTokens))
	allTokens = append(allTokens, syncTokens...)
	allTokens = append(allTokens, additionalTokens...)
	return er.parser.synchronize(allTokens)
}

// AddExpectError adds an error when an expected token is missing.
// This is a high-level wrapper around the common "expected X, got Y" error pattern.
//
// Example:
//
//	if !p.expectPeek(lexer.THEN) {
//	    recovery.AddExpectError(lexer.THEN, "after if condition")
//	    return nil
//	}
func (er *ErrorRecovery) AddExpectError(expected lexer.TokenType, context string) {
	var msg string
	if context != "" {
		msg = fmt.Sprintf("expected %s %s, got %s instead",
			expected, context, er.parser.cursor.Peek(1).Type)
	} else {
		msg = fmt.Sprintf("expected %s, got %s instead",
			expected, er.parser.cursor.Peek(1).Type)
	}

	code := getErrorCodeForMissingToken(expected)
	err := NewParserError(
		er.parser.cursor.Peek(1).Pos,
		er.parser.cursor.Peek(1).Length(),
		msg,
		code,
	)
	er.parser.errors = append(er.parser.errors, err)
}

// AddExpectErrorWithSuggestion adds an expect error with a recovery suggestion.
func (er *ErrorRecovery) AddExpectErrorWithSuggestion(expected lexer.TokenType, context string, suggestion string) {
	structErr := NewStructuredError(ErrKindMissing).
		WithCode(getErrorCodeForMissingToken(expected)).
		WithPosition(er.parser.cursor.Peek(1).Pos, er.parser.cursor.Peek(1).Length()).
		WithExpected(expected).
		WithActual(er.parser.cursor.Peek(1).Type, er.parser.cursor.Peek(1).Literal).
		WithSuggestion(suggestion)

	if context != "" {
		structErr = structErr.WithParsePhase(context)
	}

	er.parser.addStructuredError(structErr.Build())
}

// AddContextError adds an error with block context information.
// This provides better error messages by showing which block the error occurred in.
//
// Example output: "expected 'end' (in begin block starting at line 10)"
func (er *ErrorRecovery) AddContextError(msg string, code string) {
	er.parser.addErrorWithContext(msg, code)
}

// AddError adds a basic error at the current token position.
func (er *ErrorRecovery) AddError(msg string, code string) {
	er.parser.addError(msg, code)
}

// AddStructuredError adds a rich structured error with full context.
func (er *ErrorRecovery) AddStructuredError(err *StructuredParserError) {
	er.parser.addStructuredError(err)
}

// TryRecover attempts to recover from an error by trying multiple recovery strategies.
// It returns true if recovery succeeded (found a valid synchronization point).
//
// Recovery strategies (tried in order):
//  1. Look for specific recovery tokens
//  2. Look for statement starters
//  3. Look for block closers
//  4. Give up at EOF
func (er *ErrorRecovery) TryRecover(specificTokens ...lexer.TokenType) bool {
	// First, try to find specific tokens
	if er.SynchronizeOn(specificTokens...) {
		return !er.parser.cursor.Is(lexer.EOF)
	}
	return false
}

// ExpectWithRecovery combines expectPeek with automatic error reporting and recovery.
// This is a convenience method that reduces boilerplate in parsing functions.
//
// Returns:
//   - true if the expected token was found
//   - false if recovery was needed
//
// Example:
//
//	if !recovery.ExpectWithRecovery(lexer.THEN, "after if condition", lexer.ELSE, lexer.END) {
//	    return nil  // Error already reported and synchronized
//	}
func (er *ErrorRecovery) ExpectWithRecovery(expected lexer.TokenType, context string, syncTokens ...lexer.TokenType) bool {
	if er.parser.cursor.PeekIs(1, expected) {
		er.parser.nextToken()
		return true
	}

	// Report error
	er.AddExpectError(expected, context)

	// Try to recover
	er.SynchronizeOn(syncTokens...)

	return false
}

// ExpectOneOf checks if the peek token is one of the expected types.
// If not, it reports an error and synchronizes.
//
// Returns:
//   - The matched token type (if found)
//   - lexer.ILLEGAL if none matched
func (er *ErrorRecovery) ExpectOneOf(expected []lexer.TokenType, context string, syncTokens ...lexer.TokenType) lexer.TokenType {
	for _, t := range expected {
		if er.parser.cursor.PeekIs(1, t) {
			er.parser.nextToken()
			return t
		}
	}

	// Build error message
	msg := "expected one of: "
	for i, t := range expected {
		if i > 0 {
			msg += ", "
		}
		msg += t.String()
	}
	if context != "" {
		msg += " " + context
	}
	msg += fmt.Sprintf(", got %s instead", er.parser.cursor.Peek(1).Type)

	er.parser.addError(msg, ErrUnexpectedToken)
	er.SynchronizeOn(syncTokens...)

	return lexer.ILLEGAL
}

// SkipUntil skips tokens until one of the specified tokens is found.
// Unlike SynchronizeOn, this doesn't include default sync points (statement starters, block closers).
// Use this for very specific recovery scenarios.
func (er *ErrorRecovery) SkipUntil(tokens ...lexer.TokenType) bool {
	tokenMap := make(map[lexer.TokenType]bool)
	for _, t := range tokens {
		tokenMap[t] = true
	}

	for !er.parser.cursor.Is(lexer.EOF) {
		if tokenMap[er.parser.cursor.Current().Type] {
			return true
		}
		er.parser.nextToken()
	}

	return false
}

// IsAtSyncPoint checks if the current token is a synchronization point.
// This is useful for deciding whether to continue parsing or give up.
func (er *ErrorRecovery) IsAtSyncPoint() bool {
	curType := er.parser.cursor.Current().Type

	// Check if current token is a statement starter
	for _, t := range statementStarters {
		if curType == t {
			return true
		}
	}

	// Check if current token is a block closer
	for _, t := range blockClosers {
		if curType == t {
			return true
		}
	}

	return false
}

// SuggestMissingDelimiter creates a suggestion for a missing delimiter.
func (er *ErrorRecovery) SuggestMissingDelimiter(delimiter lexer.TokenType, blockType string) string {
	return fmt.Sprintf("add '%s' to close %s", delimiter, blockType)
}

// SuggestMissingSeparator creates a suggestion for a missing separator.
func (er *ErrorRecovery) SuggestMissingSeparator(separator lexer.TokenType, context string) string {
	return fmt.Sprintf("add '%s' to separate %s", separator, context)
}

// getErrorCodeForMissingToken returns an appropriate error code for a missing token.
// This maps common tokens to their specific error codes.
func getErrorCodeForMissingToken(t lexer.TokenType) string {
	switch t {
	case lexer.THEN:
		return ErrMissingThen
	case lexer.DO:
		return ErrMissingDo
	case lexer.END:
		return ErrMissingEnd
	case lexer.SEMICOLON:
		return ErrMissingSemicolon
	case lexer.COLON:
		return ErrMissingColon
	case lexer.RPAREN:
		return ErrMissingRParen
	case lexer.RBRACE:
		return ErrMissingRBrace
	case lexer.RBRACK:
		return ErrMissingRBracket
	case lexer.ASSIGN:
		return ErrMissingAssign
	case lexer.OF:
		return ErrMissingOf
	case lexer.TO, lexer.DOWNTO:
		return ErrMissingTo
	default:
		return ErrUnexpectedToken
	}
}
