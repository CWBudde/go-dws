package parser

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ParserError represents a structured parsing error with position information.
type ParserError struct {
	Message string
	Code    string
	Pos     lexer.Position
	Length  int
}

// Error implements the error interface.
func (e *ParserError) Error() string {
	return fmt.Sprintf("%s at %d:%d", e.Message, e.Pos.Line, e.Pos.Column)
}

// NewParserError creates a new ParserError with the given parameters.
func NewParserError(pos lexer.Position, length int, message, code string) *ParserError {
	return &ParserError{
		Message: message,
		Pos:     pos,
		Length:  length,
		Code:    code,
	}
}

// Error code constants for programmatic error handling
const (
	// ErrUnexpectedToken indicates an unexpected token was encountered
	ErrUnexpectedToken = "E_UNEXPECTED_TOKEN"

	// ErrMissingSemicolon indicates a missing semicolon
	ErrMissingSemicolon = "E_MISSING_SEMICOLON"

	// ErrMissingEnd indicates a missing 'end' keyword
	ErrMissingEnd = "E_MISSING_END"

	// ErrMissingLParen indicates a missing opening parenthesis
	ErrMissingLParen = "E_MISSING_LPAREN"

	// ErrMissingRParen indicates a missing closing parenthesis
	ErrMissingRParen = "E_MISSING_RPAREN"

	// ErrMissingRBracket indicates a missing closing bracket
	ErrMissingRBracket = "E_MISSING_RBRACKET"

	// ErrMissingRBrace indicates a missing closing brace
	ErrMissingRBrace = "E_MISSING_RBRACE"

	// ErrInvalidExpression indicates an invalid expression
	ErrInvalidExpression = "E_INVALID_EXPRESSION"

	// ErrNoPrefixParse indicates no prefix parse function found for a token
	ErrNoPrefixParse = "E_NO_PREFIX_PARSE"

	// ErrExpectedIdent indicates an identifier was expected
	ErrExpectedIdent = "E_EXPECTED_IDENT"

	// ErrExpectedType indicates a type name was expected
	ErrExpectedType = "E_EXPECTED_TYPE"

	// ErrExpectedOperator indicates an operator was expected
	ErrExpectedOperator = "E_EXPECTED_OPERATOR"

	// ErrInvalidSyntax indicates invalid syntax
	ErrInvalidSyntax = "E_INVALID_SYNTAX"

	// ErrMissingThen indicates a missing 'then' keyword
	ErrMissingThen = "E_MISSING_THEN"

	// ErrMissingDo indicates a missing 'do' keyword
	ErrMissingDo = "E_MISSING_DO"

	// ErrMissingOf indicates a missing 'of' keyword
	ErrMissingOf = "E_MISSING_OF"

	// ErrMissingTo indicates a missing 'to' or 'downto' keyword
	ErrMissingTo = "E_MISSING_TO"

	// ErrMissingIn indicates a missing 'in' keyword
	ErrMissingIn = "E_MISSING_IN"

	// ErrMissingColon indicates a missing colon
	ErrMissingColon = "E_MISSING_COLON"

	// ErrMissingAssign indicates a missing assignment operator
	ErrMissingAssign = "E_MISSING_ASSIGN"

	// ErrInvalidType indicates an invalid type
	ErrInvalidType = "E_INVALID_TYPE"
)
