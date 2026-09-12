package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// parseOperatorDeclaration parses a standalone (global) operator declaration.
// Examples:
//
//	operator + (String, Integer) : String uses StrPlusInt;
//	operator implicit (Integer) : String uses IntToStr;
//	operator in (Integer, Float) : Boolean uses DigitInFloat;
//
// PRE: cursor is OPERATOR
// POST: cursor is SEMICOLON

// PRE: cursor is OPERATOR
// POST: cursor is SEMICOLON
func (p *Parser) parseOperatorDeclaration() *ast.OperatorDecl {
	builder := p.StartNode()
	cursor := p.cursor

	decl := &ast.OperatorDecl{
		BaseNode:   ast.BaseNode{Token: cursor.Current()},
		Kind:       ast.OperatorKindGlobal,
		Visibility: ast.VisibilityPublic,
	}

	// Advance to the operator symbol/keyword (e.g., '+', 'in', 'implicit')
	cursor = cursor.Advance()
	if !isOperatorSymbolToken(cursor.Current().Type) {
		p.addError("expected operator symbol after 'operator'", ErrExpectedOperator)
		return nil
	}

	decl.OperatorToken = cursor.Current()
	decl.OperatorSymbol = normalizeOperatorSymbol(cursor.Current())

	// Conversion operators use the IMPLICIT / EXPLICIT keywords
	if cursor.Current().Type == lexer.IMPLICIT || cursor.Current().Type == lexer.EXPLICIT {
		decl.Kind = ast.OperatorKindConversion
	}

	// Parse operand type list (enclosed in parentheses)
	if cursor.Peek(1).Type != lexer.LPAREN {
		p.addError("expected '(' after operator symbol", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to '('
	p.cursor = cursor
	decl.OperandTypes = p.parseOperatorOperandTypes()
	cursor = p.cursor // Update cursor after helper
	decl.Arity = len(decl.OperandTypes)
	if decl.Arity == 0 {
		p.addError("operator declaration requires at least one operand type", ErrInvalidSyntax)
		return nil
	}

	// Optional return type
	if cursor.Peek(1).Type == lexer.COLON {
		cursor = cursor.Advance()   // move to ':'
		p.cursor = cursor.Advance() // move to the return type
		decl.ReturnType = p.parseTypeExpression()
		if isInvalidTypeExpression(decl.ReturnType) {
			return nil
		}
		cursor = p.cursor
	}

	// Expect 'uses' clause
	if cursor.Peek(1).Type != lexer.USES {
		p.addError("expected 'uses' in operator declaration", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to 'uses'
	if cursor.Peek(1).Type != lexer.IDENT {
		p.addError("expected identifier after 'uses' in operator declaration", ErrExpectedIdent)
		return nil
	}
	cursor = cursor.Advance() // move to identifier

	decl.Binding = &ast.Identifier{
		BaseNode: ast.BaseNode{
			Token: cursor.Current(),
		},
		Value: cursor.Current().Literal,
	}

	// Expect terminating semicolon
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' at end of operator declaration", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to ';'

	p.cursor = cursor
	decl, _ = builder.Finish(decl).(*ast.OperatorDecl)
	return decl
}

// parseClassOperatorDeclaration parses a class operator declared within a class body.
// Examples:
//
//	class operator += String uses AppendString;
//	class operator IN array of Integer uses ContainsArray;
//
// PRE: cursor is OPERATOR
// POST: cursor is SEMICOLON

// PRE: cursor is OPERATOR
// POST: cursor is SEMICOLON
//
//nolint:gocyclo // Operator parser handling multiple operator types
func (p *Parser) parseClassOperatorDeclaration(classToken lexer.Token, visibility ast.Visibility) *ast.OperatorDecl {
	builder := p.StartNode()
	cursor := p.cursor

	if cursor.Current().Type != lexer.OPERATOR {
		p.addError("expected 'operator' after 'class'", ErrUnexpectedToken)
		return nil
	}

	decl := &ast.OperatorDecl{
		BaseNode:   ast.BaseNode{Token: classToken},
		Kind:       ast.OperatorKindClass,
		Visibility: visibility,
	}

	// Advance to operator symbol
	cursor = cursor.Advance()
	if !isOperatorSymbolToken(cursor.Current().Type) {
		p.addError("expected operator symbol after 'class operator'", ErrExpectedOperator)
		return nil
	}

	decl.OperatorToken = cursor.Current()
	decl.OperatorSymbol = normalizeOperatorSymbol(cursor.Current())

	// Parse operand type(s)
	if cursor.Peek(1).Type == lexer.LPAREN {
		cursor = cursor.Advance() // move to '('
		p.cursor = cursor
		decl.OperandTypes = p.parseOperatorOperandTypes()
		cursor = p.cursor // Update cursor after helper
		decl.Arity = len(decl.OperandTypes)
	} else {
		if cursor.Peek(1).Type == lexer.USES || cursor.Peek(1).Type == lexer.SEMICOLON || cursor.Peek(1).Type == lexer.COLON {
			p.addError("expected operand type in class operator declaration", ErrExpectedType)
			return nil
		}

		cursor = cursor.Advance() // move to first operand token
		p.cursor = cursor
		operand, ok := p.parseTypeExpressionUntil(func(tt lexer.TokenType) bool {
			return tt == lexer.USES || tt == lexer.COLON || tt == lexer.SEMICOLON
		})
		if !ok {
			return nil
		}
		cursor = p.cursor // Update cursor after helper

		decl.OperandTypes = []ast.TypeExpression{operand}
		decl.Arity = len(decl.OperandTypes)
	}
	if decl.Arity == 0 {
		p.addError("class operator declaration requires at least one operand type", ErrInvalidSyntax)
		return nil
	}

	// Optional return type
	if cursor.Peek(1).Type == lexer.COLON {
		cursor = cursor.Advance() // move to ':'
		cursor = cursor.Advance() // move to first return type token
		p.cursor = cursor
		returnType, ok := p.parseTypeExpressionUntil(func(tt lexer.TokenType) bool {
			return tt == lexer.USES || tt == lexer.SEMICOLON
		})
		if !ok {
			return nil
		}
		cursor = p.cursor // Update cursor after helper
		decl.ReturnType = returnType
	}

	// Expect 'uses' clause
	if cursor.Peek(1).Type != lexer.USES {
		p.addError("expected 'uses' in class operator declaration", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to 'uses'
	if cursor.Peek(1).Type != lexer.IDENT {
		p.addError("expected identifier after 'uses' in class operator declaration", ErrExpectedIdent)
		return nil
	}
	cursor = cursor.Advance() // move to identifier

	decl.Binding = &ast.Identifier{
		BaseNode: ast.BaseNode{
			Token: cursor.Current(),
		},
		Value: cursor.Current().Literal,
	}

	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' at end of class operator declaration", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to ';'

	p.cursor = cursor
	decl, _ = builder.Finish(decl).(*ast.OperatorDecl)
	return decl
}

// parseOperatorOperandTypes parses the operand type list inside parentheses.
// Example: (String, Integer)
// PRE: cursor is LPAREN
// POST: cursor is RPAREN

// PRE: cursor is on LPAREN token
// POST: cursor is on RPAREN token
func (p *Parser) parseOperatorOperandTypes() []ast.TypeExpression {
	var operandTypes []ast.TypeExpression
	p.cursor = p.cursor.Advance()
	for p.cursor.Current().Type != lexer.RPAREN && p.cursor.Current().Type != lexer.EOF {
		operand := p.parseOperatorOperandType()
		if isInvalidTypeExpression(operand) {
			return operandTypes
		}
		operandTypes = append(operandTypes, operand)
		switch p.cursor.Peek(1).Type {
		case lexer.COMMA:
			p.cursor = p.cursor.Advance().Advance()
		case lexer.RPAREN:
			p.cursor = p.cursor.Advance()
			return operandTypes
		case lexer.EOF:
			p.addError("unterminated operator operand list", ErrMissingRParen)
			return operandTypes
		default:
			p.addError("expected ',' or ')' in operator operand list", ErrUnexpectedToken)
			return operandTypes
		}
	}
	return operandTypes
}

// parseOperatorOperandType accepts a type or a parameter-style operand such as
// "const items: array of const"; parameter modifiers do not change its type.
func (p *Parser) parseOperatorOperandType() ast.TypeExpression {
	if p.cursor.Current().Type == lexer.CONST || p.cursor.Current().Type == lexer.VAR {
		p.cursor = p.cursor.Advance()
	}
	if p.cursor.Current().Type == lexer.IDENT && p.cursor.Peek(1).Type == lexer.COLON {
		p.cursor = p.cursor.Advance().Advance()
	}
	return p.parseTypeExpression()
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
		return ident.Normalize(tok.Literal)
	default:
		return tok.Literal
	}
}

// parseTypeExpressionUntil parses a structured operand or return type and checks
// that the next token is a valid delimiter for the enclosing operator declaration.
// PRE: cursor is on IDENT or type keyword
// POST: cursor is on the last token of the type expression
func (p *Parser) parseTypeExpressionUntil(stopFn func(lexer.TokenType) bool) (ast.TypeExpression, bool) {
	typeExpr := p.parseTypeExpression()
	if isInvalidTypeExpression(typeExpr) {
		return nil, false
	}
	if !stopFn(p.cursor.Peek(1).Type) {
		p.addError("unexpected token after operator type", ErrUnexpectedToken)
		return nil, false
	}
	return typeExpr, true
}
