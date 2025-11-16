package parser

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseOperatorDeclaration parses a standalone (global) operator declaration.
// Examples:
//
//	operator + (String, Integer) : String uses StrPlusInt;
//	operator implicit (Integer) : String uses IntToStr;
//	operator in (Integer, Float) : Boolean uses DigitInFloat;
// PRE: curToken is OPERATOR
// POST: curToken is SEMICOLON
func (p *Parser) parseOperatorDeclaration() *ast.OperatorDecl {
	decl := &ast.OperatorDecl{
		BaseNode:   ast.BaseNode{Token: p.curToken},
		Kind:       ast.OperatorKindGlobal,
		Visibility: ast.VisibilityPublic,
	}

	// Advance to the operator symbol/keyword (e.g., '+', 'in', 'implicit')
	p.nextToken()
	if !isOperatorSymbolToken(p.curToken.Type) {
		p.addError("expected operator symbol after 'operator'", ErrExpectedOperator)
		return nil
	}

	decl.OperatorToken = p.curToken
	decl.OperatorSymbol = normalizeOperatorSymbol(p.curToken)

	// Conversion operators use the IMPLICIT / EXPLICIT keywords
	if p.curToken.Type == lexer.IMPLICIT || p.curToken.Type == lexer.EXPLICIT {
		decl.Kind = ast.OperatorKindConversion
	}

	// Parse operand type list (enclosed in parentheses)
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}
	decl.OperandTypes = p.parseOperatorOperandTypes()
	decl.Arity = len(decl.OperandTypes)
	if decl.Arity == 0 {
		p.addError("operator declaration requires at least one operand type", ErrInvalidSyntax)
		return nil
	}

	// Optional return type
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // move to ':'
		if !p.expectPeek(lexer.IDENT) {
			p.addError("expected return type after ':' in operator declaration", ErrExpectedType)
			return nil
		}
		decl.ReturnType = &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}
	}

	// Expect 'uses' clause
	if !p.expectPeek(lexer.USES) {
		p.addError("expected 'uses' in operator declaration", ErrUnexpectedToken)
		return nil
	}
	if !p.expectPeek(lexer.IDENT) {
		p.addError("expected identifier after 'uses' in operator declaration", ErrExpectedIdent)
		return nil
	}

	decl.Binding = &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

	// Expect terminating semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Set EndPos to the position after the semicolon
	decl.EndPos = p.endPosFromToken(p.curToken)

	return decl
}

// parseClassOperatorDeclaration parses a class operator declared within a class body.
// Examples:
//
//	class operator += String uses AppendString;
//	class operator IN array of Integer uses ContainsArray;
// PRE: curToken is OPERATOR
// POST: curToken is SEMICOLON
func (p *Parser) parseClassOperatorDeclaration(classToken lexer.Token, visibility ast.Visibility) *ast.OperatorDecl {
	if !p.curTokenIs(lexer.OPERATOR) {
		p.addError("expected 'operator' after 'class'", ErrUnexpectedToken)
		return nil
	}

	decl := &ast.OperatorDecl{
		BaseNode:   ast.BaseNode{Token: classToken},
		Kind:       ast.OperatorKindClass,
		Visibility: visibility,
	}

	// Advance to operator symbol
	p.nextToken()
	if !isOperatorSymbolToken(p.curToken.Type) {
		p.addError("expected operator symbol after 'class operator'", ErrExpectedOperator)
		return nil
	}

	decl.OperatorToken = p.curToken
	decl.OperatorSymbol = normalizeOperatorSymbol(p.curToken)

	// Parse operand type(s)
	if p.peekTokenIs(lexer.LPAREN) {
		if !p.expectPeek(lexer.LPAREN) {
			return nil
		}
		decl.OperandTypes = p.parseOperatorOperandTypes()
		decl.Arity = len(decl.OperandTypes)
	} else {
		if p.peekTokenIs(lexer.USES) || p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.COLON) {
			p.addError("expected operand type in class operator declaration", ErrExpectedType)
			return nil
		}

		p.nextToken() // move to first operand token
		operand, ok := p.parseTypeExpressionUntil(func(tt lexer.TokenType) bool {
			return tt == lexer.USES || tt == lexer.COLON || tt == lexer.SEMICOLON
		})
		if !ok {
			return nil
		}

		decl.OperandTypes = []ast.TypeExpression{operand}
		decl.Arity = len(decl.OperandTypes)
	}
	if decl.Arity == 0 {
		p.addError("class operator declaration requires at least one operand type", ErrInvalidSyntax)
		return nil
	}

	// Optional return type
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // move to ':'
		p.nextToken() // move to first return type token
		returnType, ok := p.parseTypeExpressionUntil(func(tt lexer.TokenType) bool {
			return tt == lexer.USES || tt == lexer.SEMICOLON
		})
		if !ok {
			return nil
		}
		decl.ReturnType = returnType
	}

	// Expect 'uses' clause
	if !p.expectPeek(lexer.USES) {
		p.addError("expected 'uses' in class operator declaration", ErrUnexpectedToken)
		return nil
	}
	if !p.expectPeek(lexer.IDENT) {
		p.addError("expected identifier after 'uses' in class operator declaration", ErrExpectedIdent)
		return nil
	}

	decl.Binding = &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Set EndPos to the position after the semicolon
	decl.EndPos = p.endPosFromToken(p.curToken)

	return decl
}

// parseOperatorOperandTypes parses the operand type list inside parentheses.
// Example: (String, Integer)
// PRE: curToken is LPAREN
// POST: curToken is RPAREN
func (p *Parser) parseOperatorOperandTypes() []ast.TypeExpression {
	operandTypes := []ast.TypeExpression{}

	p.nextToken() // move past '(' to first operand or ')'

	for !p.curTokenIs(lexer.RPAREN) {

		startToken := p.curToken
		nameParts := []string{p.curToken.Literal}

		// Collect tokens that belong to this type until ',' or ')'
		for !p.peekTokenIs(lexer.COMMA) && !p.peekTokenIs(lexer.RPAREN) {
			p.nextToken()
			nameParts = append(nameParts, p.curToken.Literal)
		}

		if !p.curTokenIs(lexer.IDENT) {
			// Allow keywords like 'array' or 'set' in operator operand types.
			if !p.curToken.Type.IsKeyword() {
				p.addError("expected type identifier in operator operand list", ErrExpectedType)
				return operandTypes
			}
		}

		operandTypes = append(operandTypes, &ast.TypeAnnotation{
			Token: startToken,
			Name:  strings.Join(nameParts, " "),
		})

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // move to ','
			p.nextToken() // move past ',' to next type
			continue
		}

		if p.peekTokenIs(lexer.RPAREN) {
			p.nextToken() // move to ')'
			break
		}

		if p.peekTokenIs(lexer.EOF) {
			p.addError("unterminated operator operand list", ErrMissingRParen)
			return operandTypes
		}

		p.addError("expected ',' or ')' in operator operand list", ErrUnexpectedToken)
		return operandTypes
	}

	return operandTypes
}

// isOperatorSymbolToken returns true if the token type is valid after 'operator'.
func isOperatorSymbolToken(t lexer.TokenType) bool {
	if t.IsOperator() {
		return true
	}

	switch t {
	case lexer.IN, lexer.NOT, lexer.IMPLICIT, lexer.EXPLICIT:
		return true
	default:
		return false
	}
}

// normalizeOperatorSymbol returns a canonical string representation for the operator.
func normalizeOperatorSymbol(tok lexer.Token) string {
	switch tok.Type {
	case lexer.IN, lexer.NOT:
		return strings.ToLower(tok.Literal)
	default:
		return tok.Literal
	}
}

// parseTypeExpressionUntil parses a type expression until the stop condition is met.
// It assumes the current token is the first token of the type expression.
// PRE: curToken is IDENT or type keyword
// POST: curToken is last token before stop condition
func (p *Parser) parseTypeExpressionUntil(stopFn func(lexer.TokenType) bool) (*ast.TypeAnnotation, bool) {
	if p.curToken.Type != lexer.IDENT && !p.curToken.Type.IsKeyword() {
		p.addError("expected type identifier", ErrExpectedType)
		return nil, false
	}

	startToken := p.curToken
	parts := []string{p.curToken.Literal}

	for !stopFn(p.peekToken.Type) {
		p.nextToken()
		parts = append(parts, p.curToken.Literal)
	}

	return &ast.TypeAnnotation{
		Token: startToken,
		Name:  strings.Join(parts, " "),
	}, true
}
