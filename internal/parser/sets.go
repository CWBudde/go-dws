package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseSetDeclaration parses a set type declaration.
// Called after 'type Name =' has already been parsed.
// Current token should be 'set'.
//
// Syntax:
//   - type TDays = set of TWeekday;
func (p *Parser) parseSetDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.SetDecl {
	setDecl := &ast.SetDecl{
		BaseNode: ast.BaseNode{Token: typeToken}, // The 'type' token
		Name:     nameIdent,
	}

	// Current token is 'set', advance to 'of'
	if !p.expectPeek(lexer.OF) {
		return nil
	}

	// Advance to type identifier
	if !p.expectPeek(lexer.IDENT) {
		p.addError("expected type identifier after 'of' in set declaration", ErrExpectedType)
		return nil
	}

	// Parse the element type
	setDecl.ElementType = &ast.TypeAnnotation{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return setDecl
}

// parseSetType parses an inline set type expression.
// Called when we encounter 'set' in a type context.
// Current token should be 'set'.
//
// Syntax:
//   - set of TypeName
//   - set of (A, B, C)  // inline anonymous enum (if supported)
func (p *Parser) parseSetType() *ast.SetTypeNode {
	setToken := p.curToken // The 'set' token

	// Expect 'of' keyword
	if !p.expectPeek(lexer.OF) {
		p.addError("expected 'of' after 'set' in set type", ErrMissingOf)
		return nil
	}

	// Parse element type
	p.nextToken() // move to element type

	// Element type can be:
	// 1. Simple identifier: TEnum
	// 2. Inline anonymous enum: (A, B, C) - would be handled by parseTypeExpression
	// 3. Subrange: 1..100 - might need special handling in future
	elementType := p.parseTypeExpression()
	if elementType == nil {
		p.addError("expected type expression after 'set of'", ErrExpectedType)
		return nil
	}

	setTypeNode := &ast.SetTypeNode{
		Token:       setToken,
		ElementType: elementType,
	}
	// EndPos is after element type
	setTypeNode.EndPos = elementType.End()

	return setTypeNode
}

// parseSetLiteral parses a set literal expression.
// Syntax:
//   - [one, two, three]       // elements
//   - [A..C]                  // range
//   - [one, three..five]      // mixed
//   - []                      // empty set
func (p *Parser) parseSetLiteral() ast.Expression {
	setLit := &ast.SetLiteral{
		Token:    p.curToken, // The '[' token
		Elements: []ast.Expression{},
	}

	// Check for empty set: []
	if p.peekTokenIs(lexer.RBRACK) {
		p.nextToken() // move to ']'
		// Set EndPos to after the ']'
		setLit.EndPos = p.endPosFromToken(p.curToken)
		return setLit
	}

	// Parse elements/ranges
	p.nextToken() // move to first element

	for !p.curTokenIs(lexer.RBRACK) && !p.curTokenIs(lexer.EOF) {
		// Parse an element (could be a simple identifier or a range)
		start := p.parseExpression(LOWEST)

		// Check if this is a range (element..element)
		if p.peekTokenIs(lexer.DOTDOT) {
			p.nextToken() // move to '..'
			rangeToken := p.curToken

			p.nextToken() // move to end expression
			end := p.parseExpression(LOWEST)

			// Create a RangeExpression
			rangeExpr := &ast.RangeExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token:  rangeToken,
						EndPos: end.End(),
					},
				},
				Start:    start,
				RangeEnd: end,
			}
			setLit.Elements = append(setLit.Elements, rangeExpr)
		} else {
			// Simple element (not a range)
			setLit.Elements = append(setLit.Elements, start)
		}

		// Check for comma (more elements) or closing bracket
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // move to comma
			p.nextToken() // move to next element
		} else if p.peekTokenIs(lexer.RBRACK) {
			p.nextToken() // move to ']'
			break
		} else {
			// Unexpected token
			p.addError("expected ',' or ']' in set literal", ErrUnexpectedToken)
			return nil
		}
	}

	// Set EndPos to after the ']'
	setLit.EndPos = p.endPosFromToken(p.curToken)
	return setLit
}
