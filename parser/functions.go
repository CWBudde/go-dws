package parser

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
)

// parseFunctionDeclaration parses a function or procedure declaration.
// Syntax: function Name(params): Type; begin ... end;
//
//	procedure Name(params); begin ... end;
func (p *Parser) parseFunctionDeclaration() *ast.FunctionDecl {
	fn := &ast.FunctionDecl{Token: p.curToken}

	// Parse function name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	fn.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Parse parameter list (if present)
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // move to '('
		fn.Parameters = p.parseParameterList()
		if !p.curTokenIs(lexer.RPAREN) {
			p.addError("expected ')' after parameter list")
			return nil
		}
	}

	// Parse return type for functions (not procedures)
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // move to ':'
		if !p.expectPeek(lexer.IDENT) {
			p.addError("expected return type after ':'")
			return nil
		}
		fn.ReturnType = &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}
	}

	// Expect semicolon after signature
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Check for optional 'static' keyword (for class methods)
	if p.peekTokenIs(lexer.STATIC) {
		p.nextToken() // move to 'static'
		// Note: IsClassMethod flag should have been set by the caller (parseClassDeclaration)
		// The 'static' keyword is optional and doesn't change the semantics
		if !p.expectPeek(lexer.SEMICOLON) {
			return nil
		}
	}

	// TODO: Handle forward declarations if we see another semicolon
	// For now, expect begin...end body

	// Parse function body
	if !p.expectPeek(lexer.BEGIN) {
		p.addError("expected 'begin' to start function body")
		return nil
	}

	fn.Body = p.parseBlockStatement()

	// Expect semicolon after end
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return fn
}

// parseParameterList parses a function parameter list.
// Syntax: (param: Type; var param: Type; a, b, c: Type)
func (p *Parser) parseParameterList() []*ast.Parameter {
	params := []*ast.Parameter{}

	p.nextToken() // move past '('

	if p.curTokenIs(lexer.RPAREN) {
		return params
	}

	for {
		// Parse one parameter group (may have multiple names with same type)
		groupParams := p.parseParameterGroup()
		params = append(params, groupParams...)

		if !p.peekTokenIs(lexer.SEMICOLON) {
			break
		}
		p.nextToken() // move to ';'
		p.nextToken() // move past ';'
	}

	if !p.expectPeek(lexer.RPAREN) {
		return params
	}

	return params
}

// parseParameterGroup parses a group of parameters with the same type.
// Syntax: name: Type  or  name1, name2, name3: Type  or  var name: Type
func (p *Parser) parseParameterGroup() []*ast.Parameter {
	params := []*ast.Parameter{}

	// Check for 'var' keyword (pass by reference)
	byRef := false
	if p.curTokenIs(lexer.VAR) {
		byRef = true
		p.nextToken() // move past 'var'
	}

	// Collect parameter names separated by commas
	names := []*ast.Identifier{}

	for {
		// Parse parameter name
		if !p.curTokenIs(lexer.IDENT) {
			p.addError("expected parameter name")
			return nil
		}

		names = append(names, &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		})

		// Check if there are more names (comma-separated)
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // move to ','
			p.nextToken() // move past ','
			continue
		}

		// No more names, expect ':' and type
		break
	}

	// Expect ':' and type
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	if !p.expectPeek(lexer.IDENT) {
		p.addError("expected type name after ':'")
		return nil
	}

	typeAnnotation := &ast.TypeAnnotation{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	// Create a parameter for each name with the same type
	for _, name := range names {
		param := &ast.Parameter{
			Token: name.Token,
			Name:  name,
			Type:  typeAnnotation,
			ByRef: byRef,
		}
		params = append(params, param)
	}

	return params
}
