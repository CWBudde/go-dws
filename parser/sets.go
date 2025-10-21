package parser

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
)

// parseSetDeclaration parses a set type declaration.
// Called after 'type Name =' has already been parsed.
// Current token should be 'set'.
//
// Syntax:
//   - type TDays = set of TWeekday;
//
// Task 8.91: Parse set type declarations
func (p *Parser) parseSetDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.SetDecl {
	setDecl := &ast.SetDecl{
		Token: typeToken, // The 'type' token
		Name:  nameIdent,
	}

	// Current token is 'set', advance to 'of'
	if !p.expectPeek(lexer.OF) {
		return nil
	}

	// Advance to type identifier
	if !p.expectPeek(lexer.IDENT) {
		p.addError("expected type identifier after 'of' in set declaration")
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

// parseSetLiteral parses a set literal expression.
// Syntax:
//   - [one, two, three]       // elements
//   - [A..C]                  // range
//   - [one, three..five]      // mixed
//   - []                      // empty set
//
// Task 8.93-8.95: Parse set literals
func (p *Parser) parseSetLiteral() ast.Expression {
	setLit := &ast.SetLiteral{
		Token:    p.curToken, // The '[' token
		Elements: []ast.Expression{},
	}

	// Check for empty set: []
	if p.peekTokenIs(lexer.RBRACK) {
		p.nextToken() // move to ']'
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
				Token: rangeToken,
				Start: start,
				End:   end,
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
			p.addError("expected ',' or ']' in set literal")
			return nil
		}
	}

	return setLit
}
