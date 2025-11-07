package parser

import (
	"fmt"
	"strconv"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseEnumDeclaration parses an enum type declaration.
// Called after 'type Name =' has already been parsed.
// Current token should be '(' or 'enum'.
//
// Syntax:
//   - type TColor = (Red, Green, Blue);
//   - type TEnum = (One = 1, Two = 5);
//   - type TEnum = enum (One, Two);
func (p *Parser) parseEnumDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.EnumDecl {
	enumDecl := &ast.EnumDecl{
		Token:  typeToken, // The 'type' token
		Name:   nameIdent,
		Values: []ast.EnumValue{},
	}

	// Expect '('
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	// Parse enum values
	p.nextToken() // move to first value

	// Check for empty enum
	if p.curTokenIs(lexer.RPAREN) {
		p.addError("enum declaration cannot be empty", ErrInvalidSyntax)
		return nil
	}

	for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
		// Parse enum value name
		// Allow identifiers and some keywords (like True, False) as enum value names
		valueName := ""
		if p.curTokenIs(lexer.IDENT) {
			valueName = p.curToken.Literal
		} else if p.curTokenIs(lexer.TRUE) || p.curTokenIs(lexer.FALSE) {
			// Allow boolean keywords as enum value names
			valueName = p.curToken.Literal
		} else {
			p.addError("expected enum value name, got "+p.curToken.Type.String(), ErrExpectedIdent)
			return nil
		}

		enumValue := ast.EnumValue{
			Name:  valueName,
			Value: nil, // Default to implicit value
		}

		// Check for explicit value: Name = Value
		if p.peekTokenIs(lexer.EQ) {
			p.nextToken() // move to '='
			p.nextToken() // move to value

			// Parse the value (could be negative)
			value, err := p.parseEnumValue()
			if err != nil {
				p.addError("invalid enum value: "+err.Error(), ErrInvalidExpression)
				return nil
			}
			enumValue.Value = &value
		}

		enumDecl.Values = append(enumDecl.Values, enumValue)

		// Move to next token (comma or closing paren)
		p.nextToken()

		// If we hit a comma, skip it and continue
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken() // move past comma to next value name
		} else if !p.curTokenIs(lexer.RPAREN) {
			p.addError("expected ',' or ')' in enum declaration, got "+p.curToken.Type.String(), ErrUnexpectedToken)
			return nil
		}
	}

	// Expect semicolon after closing paren
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// End position is at the semicolon
	enumDecl.EndPos = p.endPosFromToken(p.curToken)

	return enumDecl
}

// parseEnumValue parses an enum value (integer, possibly negative)
func (p *Parser) parseEnumValue() (int, error) {
	// Handle negative values
	isNegative := false
	if p.curTokenIs(lexer.MINUS) {
		isNegative = true
		p.nextToken() // move past minus
	}

	// Parse integer value
	if !p.curTokenIs(lexer.INT) {
		return 0, fmt.Errorf("expected integer value")
	}

	value, err := strconv.Atoi(p.curToken.Literal)
	if err != nil {
		return 0, err
	}

	if isNegative {
		value = -value
	}

	return value, nil
}
