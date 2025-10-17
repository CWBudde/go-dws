package parser

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
)

// parseClassDeclaration parses a class declaration.
// Syntax: type ClassName = class(Parent) ... end;
func (p *Parser) parseClassDeclaration() *ast.ClassDecl {
	classDecl := &ast.ClassDecl{Token: p.curToken}

	// Expect class name identifier
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	classDecl.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Expect '='
	if !p.expectPeek(lexer.EQ) {
		return nil
	}

	// Expect 'class' keyword
	if !p.expectPeek(lexer.CLASS) {
		return nil
	}

	// Check for optional parent class (TChild = class(TParent))
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // move to '('
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		classDecl.Parent = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	}

	// Parse class body (fields and methods) until 'end'
	p.nextToken() // move past 'class' or ')'

	classDecl.Fields = []*ast.FieldDecl{}
	classDecl.Methods = []*ast.FunctionDecl{}

	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
		// Skip semicolons
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// Check for field or method
		if p.curToken.Type == lexer.IDENT && p.peekTokenIs(lexer.COLON) {
			// This is a field declaration
			field := p.parseFieldDeclaration()
			if field != nil {
				classDecl.Fields = append(classDecl.Fields, field)
			}
		} else if p.curToken.Type == lexer.FUNCTION || p.curToken.Type == lexer.PROCEDURE {
			// This is a method declaration
			method := p.parseFunctionDeclaration()
			if method != nil {
				classDecl.Methods = append(classDecl.Methods, method)
			}
		} else {
			// Unknown token in class body, skip it
			p.nextToken()
			continue
		}

		p.nextToken()
	}

	// Expect 'end'
	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close class declaration")
		return nil
	}

	// Expect terminating semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return classDecl
}

// parseFieldDeclaration parses a field declaration within a class.
// Syntax: FieldName: Type;
func (p *Parser) parseFieldDeclaration() *ast.FieldDecl {
	field := &ast.FieldDecl{}

	// Current token should be the field name identifier
	field.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Expect ':'
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	// Expect type identifier
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	field.Type = &ast.TypeAnnotation{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Set default visibility (could be enhanced later with private/protected keywords)
	field.Visibility = "public"

	return field
}

// parseMemberAccess parses member access and method call expressions.
// Handles obj.field, obj.method(), and TClass.Create() syntax.
// This is registered as an infix operator for the DOT token.
func (p *Parser) parseMemberAccess(left ast.Expression) ast.Expression {
	dotToken := p.curToken // Save the '.' token

	// Advance to the member name
	p.nextToken()

	// The member name should be an identifier
	if p.curToken.Type != lexer.IDENT {
		p.addError("expected identifier after '.'")
		return nil
	}

	memberName := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check if this is a method call (followed by '(')
	if p.peekTokenIs(lexer.LPAREN) {
		// Check if this is object creation: TClass.Create()
		if ident, ok := left.(*ast.Identifier); ok && memberName.Value == "Create" {
			// This is a NewExpression
			p.nextToken() // move to '('

			newExpr := &ast.NewExpression{
				Token:     ident.Token,
				ClassName: ident,
				Arguments: []ast.Expression{},
			}

			// Parse arguments (parseExpressionList handles the advancement)
			newExpr.Arguments = p.parseExpressionList(lexer.RPAREN)

			return newExpr
		}

		// Regular method call: obj.Method()
		p.nextToken() // move to '('

		methodCall := &ast.MethodCallExpression{
			Token:     dotToken,
			Object:    left,
			Method:    memberName,
			Arguments: []ast.Expression{},
		}

		// Parse arguments (parseExpressionList handles the advancement)
		methodCall.Arguments = p.parseExpressionList(lexer.RPAREN)

		return methodCall
	}

	// Otherwise, this is simple member access: obj.field
	memberAccess := &ast.MemberAccessExpression{
		Token:  dotToken,
		Object: left,
		Member: memberName,
	}

	return memberAccess
}
