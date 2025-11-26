package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// parseIntegerLiteral parses an integer literal.
// PRE: cursor is on INT
// POST: cursor is on INT (unchanged)
func (p *Parser) parseIntegerLiteral() ast.Expression {
	currentToken := p.cursor.Current()

	lit := &ast.IntegerLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  currentToken,
				EndPos: p.endPosFromToken(currentToken),
			},
		},
	}

	literal := currentToken.Literal

	var (
		value int64
		err   error
	)

	switch {
	case len(literal) > 0 && literal[0] == '$':
		// Hexadecimal with $ prefix (Pascal style)
		value, err = strconv.ParseInt(strings.ReplaceAll(literal[1:], "_", ""), 16, 64)
	case len(literal) > 1 && (literal[0:2] == "0x" || literal[0:2] == "0X"):
		// Hexadecimal with 0x/0X prefix
		value, err = strconv.ParseInt(strings.ReplaceAll(literal[2:], "_", ""), 16, 64)
	case len(literal) > 0 && literal[0] == '%':
		// Binary with % prefix
		value, err = strconv.ParseInt(strings.ReplaceAll(literal[1:], "_", ""), 2, 64)
	case len(literal) > 1 && (literal[0:2] == "0b" || literal[0:2] == "0B"):
		// Binary with 0b/0B prefix
		value, err = strconv.ParseInt(strings.ReplaceAll(literal[2:], "_", ""), 2, 64)
	default:
		value, err = strconv.ParseInt(strings.ReplaceAll(literal, "_", ""), 10, 64)
	}

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", literal)
		p.addError(msg, ErrInvalidExpression)
		return nil
	}

	lit.Value = value
	return lit
}

// parseFloatLiteral parses a floating-point literal.
// POST: cursor is FLOAT (unchanged)
func (p *Parser) parseFloatLiteral() ast.Expression {
	currentToken := p.cursor.Current()

	lit := &ast.FloatLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  currentToken,
				EndPos: p.endPosFromToken(currentToken),
			},
		},
	}

	value, err := strconv.ParseFloat(strings.ReplaceAll(currentToken.Literal, "_", ""), 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", currentToken.Literal)
		p.addError(msg, ErrInvalidExpression)
		return nil
	}

	lit.Value = value
	return lit
}

// parseStringLiteral parses a string literal.
// POST: cursor is STRING (unchanged)
func (p *Parser) parseStringLiteral() ast.Expression {
	currentToken := p.cursor.Current()

	// The lexer has already processed the string, so we just need to
	// extract the value without the quotes
	value := currentToken.Literal

	// Handle escaped quotes ('' -> ')
	value = unescapeString(value)

	return &ast.StringLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  currentToken,
				EndPos: p.endPosFromToken(currentToken),
			},
		},
		Value: value,
	}
}

// unescapeString handles DWScript string escape sequences.
func unescapeString(s string) string {
	// Use strings.Builder for efficient string concatenation
	var result strings.Builder
	result.Grow(len(s)) // Pre-allocate approximate size

	// Convert to runes to handle UTF-8 correctly
	runes := []rune(s)
	i := 0
	for i < len(runes) {
		// Check for escaped single quote ('')
		if i < len(runes)-1 && runes[i] == '\'' && runes[i+1] == '\'' {
			result.WriteRune('\'')
			i += 2
		} else {
			result.WriteRune(runes[i])
			i++
		}
	}
	return result.String()
}

// parseBooleanLiteral parses a boolean literal.
// POST: cursor is TRUE or FALSE (unchanged)
func (p *Parser) parseBooleanLiteral() ast.Expression {
	currentToken := p.cursor.Current()
	return &ast.BooleanLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  currentToken,
				EndPos: p.endPosFromToken(currentToken),
			},
		},
		Value: currentToken.Type == lexer.TRUE,
	}
}

// parseNilLiteral parses a nil literal.
// PRE: cursor is NIL
// POST: cursor is NIL (unchanged)
func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  p.cursor.Current(),
				EndPos: p.endPosFromToken(p.cursor.Current()),
			},
		},
	}
}

// parseNullIdentifier parses the Null keyword as an identifier.
// Task 9.4.1: Null is a built-in constant, so we parse it as an identifier.
// PRE: cursor is NULL
// POST: cursor is NULL (unchanged)
func (p *Parser) parseNullIdentifier() ast.Expression {
	tok := p.cursor.Current()
	return &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  tok,
				EndPos: p.endPosFromToken(tok),
			},
		},
		Value: tok.Literal, // "Null" (preserves original casing)
	}
}

// parseUnassignedIdentifier parses the Unassigned keyword as an identifier.
// Task 9.4.1: Unassigned is a built-in constant, so we parse it as an identifier.
// PRE: cursor is UNASSIGNED
// POST: cursor is UNASSIGNED (unchanged)
func (p *Parser) parseUnassignedIdentifier() ast.Expression {
	tok := p.cursor.Current()
	return &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  tok,
				EndPos: p.endPosFromToken(tok),
			},
		},
		Value: tok.Literal, // "Unassigned" (preserves original casing)
	}
}

// parseCharLiteral parses a character literal (#65, #$41).
// PRE: cursor is CHAR
// POST: cursor is CHAR (unchanged)
func (p *Parser) parseCharLiteral() ast.Expression {
	tok := p.cursor.Current()
	lit := &ast.CharLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  tok,
				EndPos: p.endPosFromToken(tok),
			},
		},
	}

	// Parse the character value from the token literal
	// Token literal can be: "#65" (decimal) or "#$41" (hex)
	literal := tok.Literal
	if len(literal) < 2 || literal[0] != '#' {
		msg := fmt.Sprintf("invalid character literal format: %q", literal)
		p.addError(msg, ErrInvalidExpression)
		return nil
	}

	var value int64
	var err error

	if len(literal) >= 3 && literal[1] == '$' {
		// Hex format: #$41
		value, err = strconv.ParseInt(literal[2:], 16, 32)
	} else {
		// Decimal format: #65
		value, err = strconv.ParseInt(literal[1:], 10, 32)
	}

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as character literal", literal)
		p.addError(msg, ErrInvalidExpression)
		return nil
	}

	lit.Value = rune(value)
	return lit
}
