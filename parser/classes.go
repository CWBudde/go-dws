package parser

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
)

// parseClassDeclaration parses a class declaration with visibility sections.
// Syntax: type ClassName = class(Parent)
//
//	  private
//	    field1: Type;
//	  protected
//	    field2: Type;
//	  public
//	    field3: Type;
//	end;
func (p *Parser) parseClassDeclaration() *ast.ClassDecl {
	// This is the old entry point, still used by old code
	// Expect class name identifier
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	nameIdent := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Expect '='
	if !p.expectPeek(lexer.EQ) {
		return nil
	}

	// Expect 'class' keyword
	if !p.expectPeek(lexer.CLASS) {
		return nil
	}

	return p.parseClassDeclarationBody(nameIdent)
}

// parseClassDeclarationBody parses the body of a class declaration.
// Called after 'type Name = class' has already been parsed.
// Current token should be 'class'.
func (p *Parser) parseClassDeclarationBody(nameIdent *ast.Identifier) *ast.ClassDecl {
	classDecl := &ast.ClassDecl{
		Token: p.curToken, // 'class' token
		Name:  nameIdent,
	}

	// Task 7.83: Check for optional parent class and/or interfaces
	// Syntax: class(TParent, IInterface1, IInterface2)
	// First identifier is parent class (if it starts with T)
	// Rest are interfaces (if they start with I)
	// OR: class(IInterface1, IInterface2) - no parent, just interfaces
	classDecl.Interfaces = []*ast.Identifier{}

	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // move to '('

		// Parse comma-separated list of parent/interfaces
		identifiers := []*ast.Identifier{}

		for {
			if !p.expectPeek(lexer.IDENT) {
				return nil
			}
			identifiers = append(identifiers, &ast.Identifier{
				Token: p.curToken,
				Value: p.curToken.Literal,
			})

			// Check for comma (more items) or closing paren
			if p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // move to comma
				continue
			} else if p.peekTokenIs(lexer.RPAREN) {
				p.nextToken() // move to ')'
				break
			} else {
				p.addError("expected ',' or ')' in class inheritance list")
				return nil
			}
		}

		// Task 7.83: Distinguish parent class from interfaces
		// Convention: First identifier starting with 'T' is parent class
		// All others (or all if first doesn't start with 'T') are interfaces
		if len(identifiers) > 0 {
			firstIdent := identifiers[0]
			// In DWScript/Delphi, classes typically start with 'T', interfaces with 'I'
			// If first starts with 'T', it's the parent class
			// Otherwise, all are interfaces
			if len(firstIdent.Value) > 0 && firstIdent.Value[0] == 'T' {
				classDecl.Parent = firstIdent
				classDecl.Interfaces = identifiers[1:]
			} else {
				// No parent class, all are interfaces
				classDecl.Parent = nil
				classDecl.Interfaces = identifiers
			}
		}
	}

	// Check for 'abstract' keyword (Task 7.65b)
	// Syntax: type TShape = class abstract
	if p.peekTokenIs(lexer.ABSTRACT) {
		p.nextToken() // move to 'abstract'
		classDecl.IsAbstract = true
	}

	// Check for 'external' keyword (Task 7.138)
	// Syntax: type TExternal = class external
	// Syntax: type TExternal = class external 'ExternalName'
	if p.peekTokenIs(lexer.EXTERNAL) {
		p.nextToken() // move to 'external'
		classDecl.IsExternal = true

		// Check for optional external name string
		if p.peekTokenIs(lexer.STRING) {
			p.nextToken() // move to string
			classDecl.ExternalName = p.curToken.Literal
		}
	}

	// Parse class body (fields and methods) until 'end'
	p.nextToken() // move past 'class' or ')' or 'abstract' or 'external' or external name

	classDecl.Fields = []*ast.FieldDecl{}
	classDecl.Methods = []*ast.FunctionDecl{}
	classDecl.Operators = []*ast.OperatorDecl{}

	// Default visibility is public (Task 7.63e)
	currentVisibility := ast.VisibilityPublic

	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
		// Skip semicolons
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// Check for visibility section keywords (Task 7.63b-d)
		if p.curTokenIs(lexer.PRIVATE) {
			currentVisibility = ast.VisibilityPrivate
			p.nextToken()
			continue
		} else if p.curTokenIs(lexer.PROTECTED) {
			currentVisibility = ast.VisibilityProtected
			p.nextToken()
			continue
		} else if p.curTokenIs(lexer.PUBLIC) {
			currentVisibility = ast.VisibilityPublic
			p.nextToken()
			continue
		}

		// Check for 'class var' or 'class function' / 'class procedure'
		if p.curTokenIs(lexer.CLASS) {
			classToken := p.curToken
			p.nextToken() // move past 'class'

			if p.curTokenIs(lexer.VAR) {
				// Class variable: class var FieldName: Type;
				p.nextToken() // move past 'var'
				field := p.parseFieldDeclaration(currentVisibility)
				if field != nil {
					field.IsClassVar = true // Mark as class variable
					classDecl.Fields = append(classDecl.Fields, field)
				}
			} else if p.curTokenIs(lexer.OPERATOR) {
				operator := p.parseClassOperatorDeclaration(classToken, currentVisibility)
				if operator != nil {
					classDecl.Operators = append(classDecl.Operators, operator)
				}
			} else if p.curTokenIs(lexer.FUNCTION) || p.curTokenIs(lexer.PROCEDURE) {
				// Class method: class function/procedure ...
				method := p.parseFunctionDeclaration()
				if method != nil {
					method.IsClassMethod = true // Mark as class method
					method.Visibility = currentVisibility
					classDecl.Methods = append(classDecl.Methods, method)
				}
			} else {
				p.addError("expected 'var', 'function', or 'procedure' after 'class' keyword")
				p.nextToken()
				continue
			}
		} else if p.curToken.Type == lexer.IDENT && p.peekTokenIs(lexer.COLON) {
			// This is a regular instance field declaration
			field := p.parseFieldDeclaration(currentVisibility)
			if field != nil {
				classDecl.Fields = append(classDecl.Fields, field)
			}
		} else if p.curToken.Type == lexer.FUNCTION || p.curToken.Type == lexer.PROCEDURE {
			// This is a regular instance method declaration
			method := p.parseFunctionDeclaration()
			if method != nil {
				method.Visibility = currentVisibility
				classDecl.Methods = append(classDecl.Methods, method)
			}
		} else if p.curToken.Type == lexer.CONSTRUCTOR {
			// This is a constructor declaration
			method := p.parseFunctionDeclaration()
			if method != nil {
				method.IsConstructor = true
				method.Visibility = currentVisibility
				classDecl.Methods = append(classDecl.Methods, method)
			}
		} else if p.curToken.Type == lexer.DESTRUCTOR {
			// This is a destructor declaration
			method := p.parseFunctionDeclaration()
			if method != nil {
				method.IsDestructor = true
				method.Visibility = currentVisibility
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
// The visibility parameter specifies the access level for this field (Task 7.63f).
func (p *Parser) parseFieldDeclaration(visibility ast.Visibility) *ast.FieldDecl {
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

	// Set visibility from parameter (Task 7.63f)
	field.Visibility = visibility

	return field
}

// parseMemberAccess parses member access and method call expressions.
// Handles obj.field, obj.method(), and TClass.Create() syntax.
// This is registered as an infix operator for the DOT token.
func (p *Parser) parseMemberAccess(left ast.Expression) ast.Expression {
	dotToken := p.curToken // Save the '.' token

	// Advance to the member name
	p.nextToken()

	// The member name can be an identifier or a keyword (DWScript allows keywords as member names)
	// But it cannot be operators, numbers, or other invalid tokens
	if p.curToken.Type == lexer.SEMICOLON || p.curToken.Type == lexer.INT ||
		p.curToken.Type == lexer.FLOAT || p.curToken.Type == lexer.STRING ||
		p.curToken.Type == lexer.LPAREN || p.curToken.Type == lexer.RPAREN ||
		p.curToken.Type == lexer.LBRACK || p.curToken.Type == lexer.RBRACK ||
		p.curToken.Type == lexer.COMMA || p.curToken.Type == lexer.EOF {
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
