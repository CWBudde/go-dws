package parser

import (
	"fmt"

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
func (p *Parser) parseSetDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) ast.Statement {
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

	// Inline anonymous enum: type TMy = set of (A, B, C);
	// Desugared into an implicit enum declaration plus the set declaration.
	if p.cursor.Peek(1).Type == lexer.LPAREN {
		enumName := &ast.Identifier{
			Value: "$" + nameIdent.Value + "$InlineEnum",
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: nameIdent.Token},
			},
		}
		enumDecl := p.parseEnumDeclaration(enumName, typeToken, false, false)
		if enumDecl == nil {
			return nil
		}
		setDecl.ElementType = &ast.TypeAnnotation{
			Token: enumName.Token,
			Name:  enumName.Value,
		}
		return &ast.BlockStatement{
			BaseNode:   ast.BaseNode{Token: typeToken},
			Statements: []ast.Statement{enumDecl, setDecl},
		}
	}

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

// parseInlineSetEnum parses the `(a, b)` of an inline `set of (a, b)` in a type
// position, queues the implicit enum declaration for parseStatement to hoist,
// and returns the type expression naming it.
//
// PRE: cursor is OF, and the next token is LPAREN
// POST: cursor is the enum's closing RPAREN
func (p *Parser) parseInlineSetEnum(setToken lexer.Token) ast.TypeExpression {
	enumName := &ast.Identifier{
		Value: fmt.Sprintf("$InlineEnum$%d$%d", setToken.Pos.Line, setToken.Pos.Column),
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: setToken},
		},
	}

	p.parsingInlineEnum = true
	enumDecl := p.parseEnumDeclaration(enumName, setToken, false, false)
	p.parsingInlineEnum = false
	if enumDecl == nil {
		return nil
	}
	p.pendingTypeDecls = append(p.pendingTypeDecls, enumDecl)

	return &ast.TypeAnnotation{
		Token: enumName.Token,
		Name:  enumName.Value,
	}
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

	var elementType ast.TypeExpression

	if cursor.Peek(1).Type == lexer.LPAREN {
		// Inline anonymous enum: `var s : set of (a, b)`. Unlike the named form
		// (`type TMy = set of (a, b)`, handled in parseSetDeclaration) there is
		// no type name to derive the implicit enum's name from, so mint one from
		// the 'set' token's position and queue the declaration for
		// parseStatement to hoist ahead of the statement being parsed.
		elementType = p.parseInlineSetEnum(setToken)
		if elementType == nil {
			return nil
		}
	} else {
		// Parse element type
		cursor = cursor.Advance() // move to element type
		p.cursor = cursor

		// Element type can be:
		// 1. Simple identifier: TEnum
		// 2. Subrange: 1..100 - might need special handling in future
		elementType = p.parseTypeExpression()
		if elementType == nil {
			p.addError("expected type expression after 'set of'", ErrExpectedType)
			return nil
		}
	}

	setTypeNode := &ast.SetTypeNode{
		Token:       setToken,
		ElementType: elementType,
	}

	result, _ := builder.FinishWithNode(setTypeNode, elementType).(*ast.SetTypeNode)

	return result
}
