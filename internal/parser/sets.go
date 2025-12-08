package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// parseSetDeclaration parses a set type declaration.
// Called after 'type Name =' has already been parsed.
// Current token should be 'set'.
//
// Syntax:
//   - type TDays = set of TWeekday;
//
// PRE: cursor is SET
// POST: cursor is SEMICOLON

// PRE: cursor is SET
// POST: cursor is SEMICOLON
func (p *Parser) parseSetDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.SetDecl {
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
// PRE: cursor is SET
// POST: cursor is last token of element type

// PRE: cursor is SET
// POST: cursor is last token of element type
func (p *Parser) parseSetType() *ast.SetTypeNode {
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
