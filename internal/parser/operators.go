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
//
// PRE: curToken is OPERATOR
// POST: curToken is SEMICOLON
// Dispatcher: delegates to cursor or traditional mode
func (p *Parser) parseOperatorDeclaration() *ast.OperatorDecl {
	return p.parseOperatorDeclarationCursor()
}

// parseOperatorDeclarationCursor parses operator declaration using cursor mode.
// Task 2.7.1.3: Operator declaration migration
// PRE: cursor is on OPERATOR token
// POST: cursor is on SEMICOLON token
func (p *Parser) parseOperatorDeclarationCursor() *ast.OperatorDecl {
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
	decl.OperandTypes = p.parseOperatorOperandTypesCursor()
	cursor = p.cursor // Update cursor after helper
	decl.Arity = len(decl.OperandTypes)
	if decl.Arity == 0 {
		p.addError("operator declaration requires at least one operand type", ErrInvalidSyntax)
		return nil
	}

	// Optional return type
	if cursor.Peek(1).Type == lexer.COLON {
		cursor = cursor.Advance() // move to ':'
		if cursor.Peek(1).Type != lexer.IDENT {
			p.addError("expected return type after ':' in operator declaration", ErrExpectedType)
			return nil
		}
		cursor = cursor.Advance() // move to type identifier
		decl.ReturnType = &ast.TypeAnnotation{
			Token: cursor.Current(),
			Name:  cursor.Current().Literal,
		}
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
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: cursor.Current(),
			},
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
	return builder.Finish(decl).(*ast.OperatorDecl)
}

// parseClassOperatorDeclaration parses a class operator declared within a class body.
// Examples:
//
//	class operator += String uses AppendString;
//	class operator IN array of Integer uses ContainsArray;
//
// PRE: curToken is OPERATOR
// POST: curToken is SEMICOLON
// Dispatcher: delegates to cursor or traditional mode
func (p *Parser) parseClassOperatorDeclaration(classToken lexer.Token, visibility ast.Visibility) *ast.OperatorDecl {
	return p.parseClassOperatorDeclarationCursor(classToken, visibility)
}

// parseClassOperatorDeclarationCursor parses class operator using cursor mode.
// Task 2.7.1.3: Class operator declaration migration
// PRE: cursor is on OPERATOR token
// POST: cursor is on SEMICOLON token
func (p *Parser) parseClassOperatorDeclarationCursor(classToken lexer.Token, visibility ast.Visibility) *ast.OperatorDecl {
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
		decl.OperandTypes = p.parseOperatorOperandTypesCursor()
		cursor = p.cursor // Update cursor after helper
		decl.Arity = len(decl.OperandTypes)
	} else {
		if cursor.Peek(1).Type == lexer.USES || cursor.Peek(1).Type == lexer.SEMICOLON || cursor.Peek(1).Type == lexer.COLON {
			p.addError("expected operand type in class operator declaration", ErrExpectedType)
			return nil
		}

		cursor = cursor.Advance() // move to first operand token
		p.cursor = cursor
		operand, ok := p.parseTypeExpressionUntilCursor(func(tt lexer.TokenType) bool {
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
		returnType, ok := p.parseTypeExpressionUntilCursor(func(tt lexer.TokenType) bool {
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
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: cursor.Current(),
			},
		},
		Value: cursor.Current().Literal,
	}

	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' at end of class operator declaration", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to ';'

	p.cursor = cursor
	return builder.Finish(decl).(*ast.OperatorDecl)
}

// parseOperatorOperandTypes parses the operand type list inside parentheses.
// Example: (String, Integer)
// PRE: curToken is LPAREN
// POST: curToken is RPAREN
// Note: Helper function - uses traditional mode internally
// parseOperatorOperandTypesCursor parses operator operand types using cursor mode.
// PRE: cursor is on LPAREN token
// POST: cursor is on RPAREN token
func (p *Parser) parseOperatorOperandTypesCursor() []ast.TypeExpression {
	operandTypes := []ast.TypeExpression{}
	cursor := p.cursor

	cursor = cursor.Advance() // move past '(' to first operand or ')'

	for cursor.Current().Type != lexer.RPAREN {
		startToken := cursor.Current()
		nameParts := []string{cursor.Current().Literal}

		// Collect tokens that belong to this type until ',' or ')'
		for cursor.Peek(1).Type != lexer.COMMA && cursor.Peek(1).Type != lexer.RPAREN {
			cursor = cursor.Advance()
			nameParts = append(nameParts, cursor.Current().Literal)
		}

		if cursor.Current().Type != lexer.IDENT {
			// Allow keywords like 'array' or 'set' in operator operand types.
			if !cursor.Current().Type.IsKeyword() {
				p.addError("expected type identifier in operator operand list", ErrExpectedType)
				p.cursor = cursor
				return operandTypes
			}
		}

		operandTypes = append(operandTypes, &ast.TypeAnnotation{
			Token: startToken,
			Name:  strings.Join(nameParts, " "),
		})

		if cursor.Peek(1).Type == lexer.COMMA {
			cursor = cursor.Advance() // move to ','
			cursor = cursor.Advance() // move past ',' to next type
			continue
		}

		if cursor.Peek(1).Type == lexer.RPAREN {
			cursor = cursor.Advance() // move to ')'
			break
		}

		if cursor.Peek(1).Type == lexer.EOF {
			p.addError("unterminated operator operand list", ErrMissingRParen)
			p.cursor = cursor
			return operandTypes
		}

		p.addError("expected ',' or ')' in operator operand list", ErrUnexpectedToken)
		p.cursor = cursor
		return operandTypes
	}

	p.cursor = cursor

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
// Note: Helper function - uses traditional mode internally
func (p *Parser) parseTypeExpressionUntil(stopFn func(lexer.TokenType) bool) (*ast.TypeAnnotation, bool) {
	// This helper is always called from within traditional or synced mode,
	// so no dispatcher needed
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

// parseTypeExpressionUntilCursor parses a type expression until the stop condition is met (cursor mode).
// PRE: cursor is on IDENT or type keyword
// POST: cursor is on last token before stop condition
func (p *Parser) parseTypeExpressionUntilCursor(stopFn func(lexer.TokenType) bool) (*ast.TypeAnnotation, bool) {
	cursor := p.cursor

	if cursor.Current().Type != lexer.IDENT && !cursor.Current().Type.IsKeyword() {
		p.addError("expected type identifier", ErrExpectedType)
		return nil, false
	}

	startToken := cursor.Current()
	parts := []string{cursor.Current().Literal}

	for !stopFn(cursor.Peek(1).Type) {
		cursor = cursor.Advance()
		parts = append(parts, cursor.Current().Literal)
	}

	p.cursor = cursor
	return &ast.TypeAnnotation{
		Token: startToken,
		Name:  strings.Join(parts, " "),
	}, true
}
