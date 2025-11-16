package parser

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ErrorKind categorizes parser errors for better error handling and tooling integration.
type ErrorKind string

const (
	// ErrKindSyntax indicates a syntax error (malformed input)
	ErrKindSyntax ErrorKind = "syntax"

	// ErrKindUnexpected indicates an unexpected token was encountered
	ErrKindUnexpected ErrorKind = "unexpected"

	// ErrKindMissing indicates a required element is missing
	ErrKindMissing ErrorKind = "missing"

	// ErrKindInvalid indicates an invalid construct
	ErrKindInvalid ErrorKind = "invalid"

	// ErrKindAmbiguous indicates ambiguous syntax that needs clarification
	ErrKindAmbiguous ErrorKind = "ambiguous"
)

// StructuredParserError represents a rich, structured parsing error with contextual information.
// It provides better error messages and enables IDE/LSP integration by including:
//   - Error categorization (kind)
//   - Expected vs actual values
//   - Contextual information (block context, parsing phase)
//   - Helpful suggestions for fixing the error
//   - Multiple related positions for compound errors
//
// Example usage:
//
//	err := NewStructuredError(ErrKindMissing).
//	    WithMessage("missing closing parenthesis").
//	    WithCode(ErrMissingRParen).
//	    WithPosition(p.curToken.Pos, p.curToken.Length()).
//	    WithExpected(lexer.RPAREN).
//	    WithActual(p.curToken.Type, p.curToken.Literal).
//	    WithSuggestion("add ')' to close the expression").
//	    Build()
type StructuredParserError struct {
	// Core error information
	Kind    ErrorKind       // Error category (syntax, unexpected, missing, etc.)
	Message string          // Human-readable error message
	Code    string          // Machine-readable error code (e.g., ErrMissingRParen)
	Pos     lexer.Position  // Primary error position
	Length  int             // Length of the problematic token/region

	// Contextual information
	Expected     []string        // Expected tokens/constructs (e.g., [")", "end"])
	Actual       string          // Actual token/construct found
	ActualType   lexer.TokenType // Actual token type
	BlockContext *BlockContext   // Current block context (if any)
	ParsePhase   string          // Current parsing phase (e.g., "expression", "statement")

	// Additional information
	Suggestions    []string         // Helpful suggestions for fixing the error
	RelatedPos     []lexer.Position // Related positions (for multi-part errors)
	RelatedMessages []string        // Messages for related positions
	Notes          []string         // Additional notes or context
}

// Error implements the error interface.
// It formats the error with all available context for maximum clarity.
func (e *StructuredParserError) Error() string {
	var b strings.Builder

	// Start with the main message
	if e.Message != "" {
		b.WriteString(e.Message)
	} else {
		// Auto-generate message based on kind if not provided
		b.WriteString(e.autoGenerateMessage())
	}

	// Add position
	b.WriteString(fmt.Sprintf(" at %d:%d", e.Pos.Line, e.Pos.Column))

	// Add block context if available
	if e.BlockContext != nil {
		b.WriteString(fmt.Sprintf(" (in %s block starting at line %d)",
			e.BlockContext.BlockType, e.BlockContext.StartLine))
	}

	// Add parsing phase if available
	if e.ParsePhase != "" {
		b.WriteString(fmt.Sprintf(" [while parsing %s]", e.ParsePhase))
	}

	return b.String()
}

// autoGenerateMessage generates a reasonable error message based on the error kind and fields.
func (e *StructuredParserError) autoGenerateMessage() string {
	switch e.Kind {
	case ErrKindMissing:
		if len(e.Expected) == 1 {
			return fmt.Sprintf("missing %s", e.Expected[0])
		} else if len(e.Expected) > 1 {
			return fmt.Sprintf("missing one of: %s", strings.Join(e.Expected, ", "))
		}
		return "missing required element"

	case ErrKindUnexpected:
		if len(e.Expected) > 0 && e.Actual != "" {
			if len(e.Expected) == 1 {
				return fmt.Sprintf("expected %s, got %s", e.Expected[0], e.Actual)
			}
			return fmt.Sprintf("expected one of [%s], got %s",
				strings.Join(e.Expected, ", "), e.Actual)
		} else if e.Actual != "" {
			return fmt.Sprintf("unexpected %s", e.Actual)
		}
		return "unexpected token"

	case ErrKindInvalid:
		if e.Actual != "" {
			return fmt.Sprintf("invalid %s", e.Actual)
		}
		return "invalid syntax"

	case ErrKindAmbiguous:
		return "ambiguous syntax"

	case ErrKindSyntax:
		return "syntax error"

	default:
		return "parse error"
	}
}

// DetailedError returns a detailed, multi-line error message with all context.
// This is useful for CLI tools and detailed error reports.
func (e *StructuredParserError) DetailedError() string {
	var b strings.Builder

	// Header: error message and position
	b.WriteString(e.Error())
	b.WriteString("\n")

	// Expected vs actual (if available)
	if len(e.Expected) > 0 || e.Actual != "" {
		b.WriteString("  Details:\n")
		if len(e.Expected) > 0 {
			b.WriteString(fmt.Sprintf("    Expected: %s\n", strings.Join(e.Expected, " or ")))
		}
		if e.Actual != "" {
			b.WriteString(fmt.Sprintf("    Found:    %s\n", e.Actual))
		}
	}

	// Suggestions (if available)
	if len(e.Suggestions) > 0 {
		b.WriteString("  Suggestions:\n")
		for _, suggestion := range e.Suggestions {
			b.WriteString(fmt.Sprintf("    - %s\n", suggestion))
		}
	}

	// Related positions (if available)
	if len(e.RelatedPos) > 0 {
		b.WriteString("  Related:\n")
		for i, pos := range e.RelatedPos {
			msg := ""
			if i < len(e.RelatedMessages) {
				msg = e.RelatedMessages[i]
			}
			b.WriteString(fmt.Sprintf("    %d:%d: %s\n", pos.Line, pos.Column, msg))
		}
	}

	// Notes (if available)
	if len(e.Notes) > 0 {
		b.WriteString("  Notes:\n")
		for _, note := range e.Notes {
			b.WriteString(fmt.Sprintf("    - %s\n", note))
		}
	}

	return b.String()
}

// ToParserError converts a StructuredParserError to a legacy ParserError.
// This enables backward compatibility with existing error handling code.
// The block context is included in the message for backward compatibility.
func (e *StructuredParserError) ToParserError() *ParserError {
	msg := e.Message

	// Add block context to message for backward compatibility
	if e.BlockContext != nil {
		msg = fmt.Sprintf("%s (in %s block starting at line %d)",
			msg, e.BlockContext.BlockType, e.BlockContext.StartLine)
	}

	return &ParserError{
		Message: msg,
		Code:    e.Code,
		Pos:     e.Pos,
		Length:  e.Length,
	}
}

// StructuredErrorBuilder provides a fluent API for building StructuredParserError instances.
// This builder pattern makes it easy to construct rich errors with optional fields.
//
// Example:
//
//	err := NewStructuredError(ErrKindMissing).
//	    WithMessage("missing closing parenthesis").
//	    WithCode(ErrMissingRParen).
//	    WithPosition(pos, length).
//	    WithSuggestion("add ')' to close the expression").
//	    Build()
type StructuredErrorBuilder struct {
	err *StructuredParserError
}

// NewStructuredError creates a new StructuredErrorBuilder with the given error kind.
func NewStructuredError(kind ErrorKind) *StructuredErrorBuilder {
	return &StructuredErrorBuilder{
		err: &StructuredParserError{
			Kind:        kind,
			Expected:    []string{},
			Suggestions: []string{},
			RelatedPos:  []lexer.Position{},
			RelatedMessages: []string{},
			Notes:       []string{},
		},
	}
}

// WithMessage sets the primary error message.
func (b *StructuredErrorBuilder) WithMessage(msg string) *StructuredErrorBuilder {
	b.err.Message = msg
	return b
}

// WithCode sets the machine-readable error code.
func (b *StructuredErrorBuilder) WithCode(code string) *StructuredErrorBuilder {
	b.err.Code = code
	return b
}

// WithPosition sets the primary error position and length.
func (b *StructuredErrorBuilder) WithPosition(pos lexer.Position, length int) *StructuredErrorBuilder {
	b.err.Pos = pos
	b.err.Length = length
	return b
}

// WithExpected adds an expected token/construct.
// Can be called multiple times to add multiple expected values.
func (b *StructuredErrorBuilder) WithExpected(token lexer.TokenType) *StructuredErrorBuilder {
	b.err.Expected = append(b.err.Expected, token.String())
	return b
}

// WithExpectedString adds an expected construct as a string.
// Use this for non-token expectations like "expression" or "identifier".
func (b *StructuredErrorBuilder) WithExpectedString(expected string) *StructuredErrorBuilder {
	b.err.Expected = append(b.err.Expected, expected)
	return b
}

// WithActual sets the actual token/construct found.
func (b *StructuredErrorBuilder) WithActual(tokenType lexer.TokenType, literal string) *StructuredErrorBuilder {
	b.err.ActualType = tokenType
	if literal != "" {
		b.err.Actual = fmt.Sprintf("%s (%s)", tokenType, literal)
	} else {
		b.err.Actual = tokenType.String()
	}
	return b
}

// WithActualString sets the actual construct as a string.
func (b *StructuredErrorBuilder) WithActualString(actual string) *StructuredErrorBuilder {
	b.err.Actual = actual
	return b
}

// WithBlockContext sets the current block context.
func (b *StructuredErrorBuilder) WithBlockContext(ctx *BlockContext) *StructuredErrorBuilder {
	if ctx != nil {
		b.err.BlockContext = ctx
	}
	return b
}

// WithParsePhase sets the current parsing phase.
func (b *StructuredErrorBuilder) WithParsePhase(phase string) *StructuredErrorBuilder {
	b.err.ParsePhase = phase
	return b
}

// WithSuggestion adds a helpful suggestion for fixing the error.
// Can be called multiple times to add multiple suggestions.
func (b *StructuredErrorBuilder) WithSuggestion(suggestion string) *StructuredErrorBuilder {
	b.err.Suggestions = append(b.err.Suggestions, suggestion)
	return b
}

// WithRelatedPosition adds a related position with an optional message.
// Useful for multi-part errors (e.g., "opening brace at 10:5, but missing closing brace").
func (b *StructuredErrorBuilder) WithRelatedPosition(pos lexer.Position, msg string) *StructuredErrorBuilder {
	b.err.RelatedPos = append(b.err.RelatedPos, pos)
	b.err.RelatedMessages = append(b.err.RelatedMessages, msg)
	return b
}

// WithNote adds an additional note or context.
// Can be called multiple times to add multiple notes.
func (b *StructuredErrorBuilder) WithNote(note string) *StructuredErrorBuilder {
	b.err.Notes = append(b.err.Notes, note)
	return b
}

// Build returns the constructed StructuredParserError.
func (b *StructuredErrorBuilder) Build() *StructuredParserError {
	// Auto-generate message if not provided
	if b.err.Message == "" {
		b.err.Message = b.err.autoGenerateMessage()
	}
	return b.err
}

// Helper functions for common error patterns

// NewUnexpectedTokenError creates a structured error for an unexpected token.
// This is the most common error type in parsing.
func NewUnexpectedTokenError(pos lexer.Position, length int, expected lexer.TokenType, actual lexer.TokenType, actualLiteral string) *StructuredParserError {
	return NewStructuredError(ErrKindUnexpected).
		WithCode(ErrUnexpectedToken).
		WithPosition(pos, length).
		WithExpected(expected).
		WithActual(actual, actualLiteral).
		Build()
}

// NewMissingTokenError creates a structured error for a missing token.
func NewMissingTokenError(pos lexer.Position, length int, missing lexer.TokenType, code string) *StructuredParserError {
	return NewStructuredError(ErrKindMissing).
		WithCode(code).
		WithPosition(pos, length).
		WithExpected(missing).
		Build()
}

// NewInvalidExpressionError creates a structured error for an invalid expression.
func NewInvalidExpressionError(pos lexer.Position, length int, reason string) *StructuredParserError {
	builder := NewStructuredError(ErrKindInvalid).
		WithCode(ErrInvalidExpression).
		WithPosition(pos, length)

	if reason != "" {
		builder.WithMessage(fmt.Sprintf("invalid expression: %s", reason))
	}

	return builder.Build()
}
