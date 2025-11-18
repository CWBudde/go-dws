package parser

import (
	"fmt"
	"strconv"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseEnumDeclaration parses an enum type declaration (dual-mode dispatcher).
// Called after 'type Name =' has already been parsed.
//
// Syntax:
//   - type TColor = (Red, Green, Blue);          // unscoped enum
//   - type TEnum = (One = 1, Two = 5);           // unscoped enum with values
//   - type TEnum = enum (One, Two);              // scoped enum
//   - type TFlags = flags (a, b, c);             // flags enum (scoped, power-of-2 values)
func (p *Parser) parseEnumDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token, scoped bool, flags bool) *ast.EnumDecl {
	if p.useCursor {
		return p.parseEnumDeclarationCursor(nameIdent, typeToken, scoped, flags)
	}
	return p.parseEnumDeclarationTraditional(nameIdent, typeToken, scoped, flags)
}

// parseEnumDeclarationTraditional parses an enum type declaration (traditional mode).
// Called after 'type Name =' has already been parsed.
// Current token should be '(' or 'enum' or 'flags'.
//
// Syntax:
//   - type TColor = (Red, Green, Blue);          // unscoped enum
//   - type TEnum = (One = 1, Two = 5);           // unscoped enum with values
//   - type TEnum = enum (One, Two);              // scoped enum
//   - type TFlags = flags (a, b, c);             // flags enum (scoped, power-of-2 values)
//
// PRE: curToken is LPAREN (or after ENUM/FLAGS, will advance to LPAREN)
// POST: curToken is SEMICOLON
func (p *Parser) parseEnumDeclarationTraditional(nameIdent *ast.Identifier, typeToken lexer.Token, scoped bool, flags bool) *ast.EnumDecl {
	builder := p.StartNode()
	enumDecl := &ast.EnumDecl{
		BaseNode: ast.BaseNode{Token: typeToken}, // The 'type' token
		Name:     nameIdent,
		Values:   []ast.EnumValue{},
		Scoped:   scoped,
		Flags:    flags,
	}

	// Expect '('
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	// Parse enum values using SeparatedList combinator (Task 2.3.2)
	p.nextToken() // move to first value

	// Check for empty enum
	if p.curTokenIs(lexer.RPAREN) {
		p.addError("enum declaration cannot be empty", ErrInvalidSyntax)
		return nil
	}

	count := p.SeparatedList(SeparatorConfig{
		Sep:           lexer.COMMA,
		Term:          lexer.RPAREN,
		AllowEmpty:    false,
		AllowTrailing: true,
		RequireTerm:   true,
		ParseItem: func() bool {
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
				return false
			}

			enumValue := ast.EnumValue{
				Name:  valueName,
				Value: nil, // Default to implicit value
			}

			// Check for optional 'deprecated' keyword (must come before the value assignment, if any)
			if p.peekTokenIs(lexer.DEPRECATED) {
				p.nextToken() // move to 'deprecated'
				enumValue.IsDeprecated = true

				// Check for optional deprecation message string
				if p.peekTokenIs(lexer.STRING) {
					p.nextToken() // move to string
					enumValue.DeprecatedMessage = p.curToken.Literal
				}
			}

			// Check for explicit value: Name [deprecated] = Value
			if p.peekTokenIs(lexer.EQ) {
				p.nextToken() // move to '='
				p.nextToken() // move to value

				// Parse the value (could be negative)
				value, err := p.parseEnumValue()
				if err != nil {
					p.addError("invalid enum value: "+err.Error(), ErrInvalidExpression)
					return false
				}
				enumValue.Value = &value
			}

			enumDecl.Values = append(enumDecl.Values, enumValue)
			return true
		},
	})

	// Check if parsing succeeded
	if count == -1 {
		return nil
	}

	// Expect semicolon after closing paren
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// End position is at the semicolon
	return builder.Finish(enumDecl).(*ast.EnumDecl)
}

// parseEnumDeclarationCursor parses an enum type declaration (cursor mode).
// Called after 'type Name =' has already been parsed.
//
// Syntax:
//   - type TColor = (Red, Green, Blue);          // unscoped enum
//   - type TEnum = (One = 1, Two = 5);           // unscoped enum with values
//   - type TEnum = enum (One, Two);              // scoped enum
//   - type TFlags = flags (a, b, c);             // flags enum (scoped, power-of-2 values)
//
// PRE: cursor is at LPAREN (or after ENUM/FLAGS, will advance to LPAREN)
// POST: cursor is at SEMICOLON
func (p *Parser) parseEnumDeclarationCursor(nameIdent *ast.Identifier, typeToken lexer.Token, scoped bool, flags bool) *ast.EnumDecl {
	builder := p.StartNode()
	cursor := p.cursor

	enumDecl := &ast.EnumDecl{
		BaseNode: ast.BaseNode{Token: typeToken}, // The 'type' token
		Name:     nameIdent,
		Values:   []ast.EnumValue{},
		Scoped:   scoped,
		Flags:    flags,
	}

	// Expect '('
	if cursor.Peek(1).Type != lexer.LPAREN {
		p.addError("expected '(' to start enum declaration", ErrMissingLParen)
		return nil
	}
	cursor = cursor.Advance() // move to '('
	p.cursor = cursor

	// Parse enum values using SeparatedList combinator (Task 2.3.2)
	cursor = cursor.Advance() // move to first value
	p.cursor = cursor

	// Check for empty enum
	if cursor.Current().Type == lexer.RPAREN {
		p.addError("enum declaration cannot be empty", ErrInvalidSyntax)
		return nil
	}

	count := p.SeparatedList(SeparatorConfig{
		Sep:           lexer.COMMA,
		Term:          lexer.RPAREN,
		AllowEmpty:    false,
		AllowTrailing: true,
		RequireTerm:   true,
		ParseItem: func() bool {
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
				return false
			}

			enumValue := ast.EnumValue{
				Name:  valueName,
				Value: nil, // Default to implicit value
			}

			// Check for optional 'deprecated' keyword (must come before the value assignment, if any)
			if p.peekTokenIs(lexer.DEPRECATED) {
				p.nextToken() // move to 'deprecated'
				enumValue.IsDeprecated = true

				// Check for optional deprecation message string
				if p.peekTokenIs(lexer.STRING) {
					p.nextToken() // move to string
					enumValue.DeprecatedMessage = p.curToken.Literal
				}
			}

			// Check for explicit value: Name [deprecated] = Value
			if p.peekTokenIs(lexer.EQ) {
				p.nextToken() // move to '='
				p.nextToken() // move to value

				// Parse the value (could be negative)
				value, err := p.parseEnumValue()
				if err != nil {
					p.addError("invalid enum value: "+err.Error(), ErrInvalidExpression)
					return false
				}
				enumValue.Value = &value
			}

			enumDecl.Values = append(enumDecl.Values, enumValue)
			return true
		},
	})
	cursor = p.cursor

	// Check if parsing succeeded
	if count == -1 {
		return nil
	}

	// Expect semicolon after closing paren
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after enum declaration", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	// End position is at the semicolon
	return builder.Finish(enumDecl).(*ast.EnumDecl)
}

// parseEnumValue parses an enum value (dual-mode dispatcher).
// Integer, possibly negative.
func (p *Parser) parseEnumValue() (int, error) {
	if p.useCursor {
		return p.parseEnumValueCursor()
	}
	return p.parseEnumValueTraditional()
}

// parseEnumValueTraditional parses an enum value (integer, possibly negative) (traditional mode).
// PRE: curToken is first token of value (INT or MINUS)
// POST: curToken is INT
func (p *Parser) parseEnumValueTraditional() (int, error) {
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

// parseEnumValueCursor parses an enum value (integer, possibly negative) (cursor mode).
// PRE: cursor is at first token of value (INT or MINUS)
// POST: cursor is at INT
func (p *Parser) parseEnumValueCursor() (int, error) {
	cursor := p.cursor

	// Handle negative values
	isNegative := false
	if cursor.Current().Type == lexer.MINUS {
		isNegative = true
		cursor = cursor.Advance() // move past minus
		p.cursor = cursor
	}

	// Parse integer value
	if cursor.Current().Type != lexer.INT {
		return 0, fmt.Errorf("expected integer value")
	}

	value, err := strconv.Atoi(cursor.Current().Literal)
	if err != nil {
		return 0, err
	}

	if isNegative {
		value = -value
	}

	return value, nil
}
