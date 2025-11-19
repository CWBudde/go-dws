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
//
// PRE: curToken is SET
// POST: curToken is SEMICOLON
// Dispatcher: delegates to cursor or traditional mode
func (p *Parser) parseSetDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.SetDecl {
	return p.parseSetDeclarationCursor(nameIdent, typeToken)
}

// parseSetDeclarationCursor parses set declaration using cursor mode.
// Task 2.7.1.4: Set declaration migration
// PRE: cursor is on SET token
// POST: cursor is on SEMICOLON token
func (p *Parser) parseSetDeclarationCursor(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.SetDecl {
	setDecl := &ast.SetDecl{
		BaseNode: ast.BaseNode{Token: typeToken}, // The 'type' token
		Name:     nameIdent,
	}

	// Current token is 'set', expect 'of'
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.OF {
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingOf).
			WithMessage("expected 'of' after 'set' in set declaration").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpectedString("'of'").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("add 'of' after 'set'").
			WithParsePhase("set declaration").
			Build()
		p.addStructuredError(err)
		return nil
	}
	p.cursor = p.cursor.Advance() // move to 'of'

	// Expect type identifier
	nextToken = p.cursor.Peek(1)
	if nextToken.Type != lexer.IDENT {
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrExpectedType).
			WithMessage("expected type identifier after 'of' in set declaration").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpectedString("type name").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("provide a type name after 'of'").
			WithParsePhase("set declaration").
			Build()
		p.addStructuredError(err)
		return nil
	}
	p.cursor = p.cursor.Advance() // move to type identifier

	// Parse the element type
	currentToken := p.cursor.Current()
	setDecl.ElementType = &ast.TypeAnnotation{
		Token: currentToken,
		Name:  currentToken.Literal,
	}

	// Expect semicolon
	nextToken = p.cursor.Peek(1)
	if nextToken.Type != lexer.SEMICOLON {
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingSemicolon).
			WithMessage("expected ';' after set declaration").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpectedString("';'").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("add ';' after set declaration").
			WithParsePhase("set declaration").
			Build()
		p.addStructuredError(err)
		return nil
	}
	p.cursor = p.cursor.Advance() // move to semicolon

	return setDecl
}

// parseSetType parses an inline set type expression.
// Called when we encounter 'set' in a type context.
// Current token should be 'set'.
//
// Syntax:
//   - set of TypeName
//   - set of (A, B, C)  // inline anonymous enum (if supported)
//
// PRE: curToken is SET
// POST: curToken is last token of element type
// Dispatcher: delegates to cursor or traditional mode
func (p *Parser) parseSetType() *ast.SetTypeNode {
	return p.parseSetTypeCursor()
}

// parseSetTypeCursor parses set type using cursor mode.
// Task 2.7.1.4: Set type migration
// Task 2.7.2: Completed full cursor implementation
// PRE: cursor is on SET token
// POST: cursor is on last token of element type
func (p *Parser) parseSetTypeCursor() *ast.SetTypeNode {
	cursor := p.cursor
	builder := p.StartNode()

	setToken := cursor.Current() // The 'set' token

	// Expect 'of' keyword
	if cursor.Peek(1).Type != lexer.OF {
		p.addError("expected 'of' after 'set' in set type", ErrMissingOf)
		return nil
	}
	cursor = cursor.Advance() // move to OF
	p.cursor = cursor

	// Parse element type
	cursor = cursor.Advance() // move to element type
	p.cursor = cursor

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

	return builder.FinishWithNode(setTypeNode, elementType).(*ast.SetTypeNode)
}

// parseSetLiteral parses a set literal expression.
// Syntax:
//   - [one, two, three]       // elements
//   - [A..C]                  // range
//   - [one, three..five]      // mixed
//   - []                      // empty set
//
// PRE: curToken is LBRACK
// POST: curToken is RBRACK
// Dispatcher: delegates to cursor or traditional mode
func (p *Parser) parseSetLiteral() ast.Expression {
	return p.parseSetLiteralCursor()
}

// parseSetLiteralCursor parses set literal using cursor mode.
// Task 2.7.1.4: Set literal migration
// PRE: cursor is on LBRACK token
// POST: cursor is on RBRACK token
func (p *Parser) parseSetLiteralCursor() ast.Expression {
	builder := p.StartNode()
	cursor := p.cursor

	setLit := &ast.SetLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: cursor.Current()}, // The '[' token
		},
		Elements: []ast.Expression{},
	}

	// Check for empty set: []
	if cursor.Peek(1).Type == lexer.RBRACK {
		cursor = cursor.Advance() // move to ']'
		p.cursor = cursor
		return builder.Finish(setLit).(*ast.SetLiteral)
	}

	// Parse elements/ranges
	cursor = cursor.Advance() // move to first element
	p.cursor = cursor

	for cursor.Current().Type != lexer.RBRACK && cursor.Current().Type != lexer.EOF {
		// Parse an element (could be a simple identifier or a range)
		start := p.parseExpression(LOWEST)
		cursor = p.cursor // Update cursor after parseExpression

		// Check if this is a range (element..element)
		if cursor.Peek(1).Type == lexer.DOTDOT {
			cursor = cursor.Advance() // move to '..'
			rangeToken := cursor.Current()

			cursor = cursor.Advance() // move to end expression
			p.cursor = cursor
			end := p.parseExpression(LOWEST)
			cursor = p.cursor // Update cursor after parseExpression

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
		if cursor.Peek(1).Type == lexer.COMMA {
			cursor = cursor.Advance() // move to comma
			cursor = cursor.Advance() // move to next element
			p.cursor = cursor
		} else if cursor.Peek(1).Type == lexer.RBRACK {
			cursor = cursor.Advance() // move to ']'
			p.cursor = cursor
			break
		} else {
			// Unexpected token
			p.addError("expected ',' or ']' in set literal", ErrUnexpectedToken)
			return nil
		}
	}

	p.cursor = cursor
	return builder.Finish(setLit).(*ast.SetLiteral)
}
