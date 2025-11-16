package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
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
//
// PRE: curToken is CLASS or TYPE (when called from parseTypeDeclaration)
// POST: curToken is END
func (p *Parser) parseClassDeclaration() *ast.ClassDecl {
	// This is the old entry point, still used by old code
	// Expect class name identifier
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	nameIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

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

// parseClassParentAndInterfaces parses optional parent class and interfaces from (...)
// Can be called multiple times for syntax like: class abstract(TParent)
// Only updates classDecl if not already set to avoid overwriting previous parse
// PRE: curToken is before LPAREN (peekToken is LPAREN)
// POST: curToken is RPAREN or unchanged if no parentheses
func (p *Parser) parseClassParentAndInterfaces(classDecl *ast.ClassDecl) {
	if !p.peekTokenIs(lexer.LPAREN) {
		return
	}

	p.nextToken() // move to '('

	// Parse comma-separated list of parent/interfaces
	identifiers := []*ast.Identifier{}

	for {
		if !p.expectPeek(lexer.IDENT) {
			return
		}
		identifiers = append(identifiers, &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: p.curToken,
				},
			},
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
			p.addError("expected ',' or ')' in class inheritance list", ErrUnexpectedToken)
			return
		}
	}

	// Distinguish parent class from interfaces
	//
	// Convention: First identifier is parent class if:
	//   1. It's a built-in class (Exception, EConvertError, etc.), OR
	//   2. It starts with 'T' (TObject, TMyClass, etc.)
	// Otherwise, all identifiers are treated as interfaces
	if len(identifiers) > 0 {
		firstIdent := identifiers[0]
		// Check if first identifier is a built-in class or starts with 'T' (case-insensitive)
		// Task 9.14.2: Make 'T' check case-insensitive for parent class detection
		firstChar := firstIdent.Value[0]
		if isBuiltinClass(firstIdent.Value) ||
			(len(firstIdent.Value) > 0 && (firstChar == 'T' || firstChar == 't')) {
			// First identifier is the parent class
			if classDecl.Parent == nil {
				classDecl.Parent = firstIdent
			}
			classDecl.Interfaces = append(classDecl.Interfaces, identifiers[1:]...)
		} else {
			// No parent class, all are interfaces
			classDecl.Interfaces = append(classDecl.Interfaces, identifiers...)
		}
	}
}

// isBuiltinClass checks if a class name is a built-in class that doesn't follow
// the 'T' prefix convention. These classes need special handling in parent/interface
// disambiguation since they don't start with 'T' but are parent classes, not interfaces.
func isBuiltinClass(name string) bool {
	builtinClasses := []string{
		"Exception",     // Base exception class
		"EConvertError", // Standard exception types
		"ERangeError",
		"EDivByZero",
		"EAssertionFailed",
		"EInvalidOp",
	}

	for _, builtin := range builtinClasses {
		if name == builtin {
			return true
		}
	}
	return false
}

// parseClassDeclarationBody parses the body of a class declaration.
// Called after 'type Name = class' has already been parsed.
// Current token should be 'class'.
// PRE: curToken is CLASS
// POST: curToken is END
func (p *Parser) parseClassDeclarationBody(nameIdent *ast.Identifier) *ast.ClassDecl {
	classDecl := &ast.ClassDecl{
		BaseNode: ast.BaseNode{Token: p.curToken}, // 'class' token
		Name:     nameIdent,
	}

	// Check for optional parent class and/or interfaces
	//
	// Syntax: class(TParent, IInterface1, IInterface2)
	// Syntax: class abstract(TParent) - parent after abstract
	// First identifier is parent class (if it starts with T)
	// Rest are interfaces (if they start with I)
	// OR: class(IInterface1, IInterface2) - no parent, just interfaces
	classDecl.Interfaces = []*ast.Identifier{}

	p.parseClassParentAndInterfaces(classDecl)

	// Check for 'abstract' keyword
	// Syntax: type TShape = class abstract
	// Syntax: type TShape = class abstract(TParent)
	if p.peekTokenIs(lexer.ABSTRACT) {
		p.nextToken() // move to 'abstract'
		classDecl.IsAbstract = true
		// Check again for parent/interfaces after abstract
		p.parseClassParentAndInterfaces(classDecl)
	}

	// Check for 'external' keyword
	// Syntax: type TExternal = class external
	// Syntax: type TExternal = class external 'ExternalName'
	// Syntax: type TExternal = class external(TParent)
	if p.peekTokenIs(lexer.EXTERNAL) {
		p.nextToken() // move to 'external'
		classDecl.IsExternal = true

		// Check for optional external name string
		if p.peekTokenIs(lexer.STRING) {
			p.nextToken() // move to string
			classDecl.ExternalName = p.curToken.Literal
		}
		// Check again for parent/interfaces after external
		p.parseClassParentAndInterfaces(classDecl)
	}

	// Check for forward declaration: type TForward = class;
	// Syntax: type TChild = class;
	// Syntax: type TChild = class(TParent);
	// This can also represent a short-form class declaration inheriting from parent
	// without adding any new members. The semantic analyzer distinguishes between
	// forward declarations (slices are nil) and empty classes (slices are empty but initialized).
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken() // move to semicolon
		classDecl.EndPos = p.endPosFromToken(p.curToken)
		// Do NOT initialize the slices - leave them as nil so semantic analyzer
		// can detect this as a forward declaration
		return classDecl
	}

	// Parse class body (fields and methods) until 'end'
	p.nextToken() // move past 'class' or ')' or 'abstract' or 'external' or external name

	classDecl.Fields = []*ast.FieldDecl{}
	classDecl.Methods = []*ast.FunctionDecl{}
	classDecl.Operators = []*ast.OperatorDecl{}
	classDecl.Properties = []*ast.PropertyDecl{}
	classDecl.Constants = []*ast.ConstDecl{}

	// Default visibility is public
	currentVisibility := ast.VisibilityPublic

	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
		// Skip semicolons
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// Check for visibility section keywords
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

		// Check for 'class var', 'class const', 'class property', or 'class function' / 'class procedure'
		if p.curTokenIs(lexer.CLASS) {
			classToken := p.curToken
			p.nextToken() // move past 'class'

			if p.curTokenIs(lexer.VAR) {
				// Class variable: class var FieldName: Type;
				p.nextToken() // move past 'var'
				fields := p.parseFieldDeclarations(currentVisibility)
				for _, field := range fields {
					if field != nil {
						field.IsClassVar = true // Mark as class variable
						classDecl.Fields = append(classDecl.Fields, field)
					}
				}
			} else if p.curTokenIs(lexer.CONST) {
				// Class constant: class const Name = Value;
				p.nextToken() // move past 'const'
				constant := p.parseClassConstantDeclaration(currentVisibility, true)
				if constant != nil {
					classDecl.Constants = append(classDecl.Constants, constant)
				}
			} else if p.curTokenIs(lexer.PROPERTY) {
				// Class property: class property Name: Type read GetName write SetName;
				property := p.parsePropertyDeclaration()
				if property != nil {
					property.IsClassProperty = true // Mark as class property
					classDecl.Properties = append(classDecl.Properties, property)
				}
			} else if p.curTokenIs(lexer.OPERATOR) {
				operator := p.parseClassOperatorDeclaration(classToken, currentVisibility)
				if operator != nil {
					classDecl.Operators = append(classDecl.Operators, operator)
				}
			} else if p.curTokenIs(lexer.FUNCTION) || p.curTokenIs(lexer.PROCEDURE) || p.curTokenIs(lexer.METHOD) {
				// Class method: class function/procedure/method ...
				method := p.parseFunctionDeclaration()
				if method != nil {
					method.IsClassMethod = true // Mark as class method
					method.Visibility = currentVisibility
					classDecl.Methods = append(classDecl.Methods, method)
				}
			} else {
				p.addError("expected 'var', 'const', 'property', 'function', 'procedure', or 'method' after 'class' keyword", ErrUnexpectedToken)
				p.nextToken()
				continue
			}
		} else if p.curTokenIs(lexer.CONST) {
			// Regular class constant: const Name = Value;
			p.nextToken() // move past 'const'
			constant := p.parseClassConstantDeclaration(currentVisibility, false)
			if constant != nil {
				classDecl.Constants = append(classDecl.Constants, constant)
			}
		} else if p.curToken.Type == lexer.IDENT && (p.peekTokenIs(lexer.COLON) || p.peekTokenIs(lexer.COMMA) || p.peekTokenIs(lexer.ASSIGN) || p.peekTokenIs(lexer.EQ)) {
			// This is a regular instance field declaration (may be comma-separated)
			// Supports: FieldName: Type; or FieldName := Value; or FieldName = Value; or FieldName: Type := Value;
			fields := p.parseFieldDeclarations(currentVisibility)
			for _, field := range fields {
				if field != nil {
					classDecl.Fields = append(classDecl.Fields, field)
				}
			}
		} else if p.curToken.Type == lexer.FUNCTION || p.curToken.Type == lexer.PROCEDURE || p.curToken.Type == lexer.METHOD {
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
		} else if p.curToken.Type == lexer.PROPERTY {
			// This is a property declaration
			property := p.parsePropertyDeclaration()
			if property != nil {
				// Note: We could track visibility here if needed
				// For now, properties are parsed without explicit visibility tracking
				classDecl.Properties = append(classDecl.Properties, property)
			}
		} else if p.curToken.Type == lexer.IDENT {
			// Unexpected identifier in class body - likely a field missing its type declaration
			p.addError("expected ':' after field name or method/property declaration keyword", ErrMissingColon)
			p.nextToken()
			continue
		} else {
			// Unknown token in class body, skip it
			p.nextToken()
			continue
		}

		p.nextToken()
	}

	// Expect 'end'
	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close class declaration", ErrMissingEnd)
		return nil
	}

	// Expect terminating semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	classDecl.EndPos = p.endPosFromToken(p.curToken)

	return classDecl
}

// parseFieldDeclaration parses a field declaration within a class.
// Syntax: FieldName: Type; or FieldName1, FieldName2, FieldName3: Type;
// The visibility parameter specifies the access level for this field.
// Returns a slice of FieldDecl nodes (one per field name) since DWScript supports
// comma-separated field names with a single type annotation.
// PRE: curToken is first field name IDENT
// POST: curToken is SEMICOLON or last token of initialization value
func (p *Parser) parseFieldDeclarations(visibility ast.Visibility) []*ast.FieldDecl {
	// Collect all field names (comma-separated)
	var fieldNames []*ast.Identifier

	// Current token should be the first field name identifier
	fieldNames = append(fieldNames, &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	})

	// Check for comma-separated field names
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next field name

		if p.curToken.Type != lexer.IDENT {
			p.addError("expected identifier after comma in field declaration", ErrExpectedIdent)
			return nil
		}

		fieldNames = append(fieldNames, &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: p.curToken,
				},
			},
			Value: p.curToken.Literal,
		})
	}

	// Parse optional type and/or initialization
	var fieldType ast.TypeExpression
	var initValue ast.Expression

	// Check for type annotation (: Type) or direct initialization (:= Value)
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // move to ':'
		p.nextToken() // move to type

		// Parse type expression (supports simple types, array types, function pointer types)
		fieldType = p.parseTypeExpression()
		if fieldType == nil {
			return nil
		}

		// Parse optional field initializer after type
		initValue = p.parseFieldInitializer(fieldNames)
	} else if p.peekTokenIs(lexer.ASSIGN) || p.peekTokenIs(lexer.EQ) {
		// Direct field initialization without type annotation: FField := Value; or FField = Value;
		// This is a shorthand syntax for field initialization from constants
		if len(fieldNames) > 1 {
			p.addError("initialization without type annotation not allowed for comma-separated field declarations", ErrInvalidExpression)
			return nil
		}

		p.nextToken() // move to ':=' or '='
		p.nextToken() // move to value expression

		// Parse initialization expression
		initValue = p.parseExpression(LOWEST)
		if initValue == nil {
			p.addError("expected initialization expression after := or =", ErrInvalidExpression)
			return nil
		}
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Create a FieldDecl for each field name with the same type
	fields := make([]*ast.FieldDecl, 0, len(fieldNames))
	for _, name := range fieldNames {
		field := &ast.FieldDecl{
			BaseNode: ast.BaseNode{
				Token: name.Token,
			},
			Name:       name,
			Type:       fieldType, // May be nil for type inference
			Visibility: visibility,
			InitValue:  initValue, // May be nil if no initialization
		}
		fields = append(fields, field)
	}

	return fields
}

// parseMemberAccess parses member access and method call expressions.
// Handles obj.field, obj.method(), and TClass.Create() syntax.
// This is registered as an infix operator for the DOT token.
// PRE: curToken is DOT
// POST: curToken is member name IDENT, RPAREN (for method calls), or last token of right operand
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
		p.addError("expected identifier after '.'", ErrExpectedIdent)
		return nil
	}

	memberName := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

	// Check if this is a method call (followed by '(')
	if p.peekTokenIs(lexer.LPAREN) {
		// Check if this is object creation: TClass.Create()
		if ident, ok := left.(*ast.Identifier); ok && memberName.Value == "Create" {
			// This is a NewExpression
			p.nextToken() // move to '('

			newExpr := &ast.NewExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: ident.Token,
					},
				},
				ClassName: ident,
				Arguments: []ast.Expression{},
			}

			// Parse arguments (parseExpressionList handles the advancement)
			newExpr.Arguments = p.parseExpressionList(lexer.RPAREN)
			newExpr.EndPos = p.endPosFromToken(p.curToken) // p.curToken is now at RPAREN

			return newExpr
		}

		// Regular method call: obj.Method()
		p.nextToken() // move to '('

		methodCall := &ast.MethodCallExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: dotToken,
				},
			},
			Object:    left,
			Method:    memberName,
			Arguments: []ast.Expression{},
		}

		// Parse arguments (parseExpressionList handles the advancement)
		methodCall.Arguments = p.parseExpressionList(lexer.RPAREN)
		methodCall.EndPos = p.endPosFromToken(p.curToken) // p.curToken is now at RPAREN

		return methodCall
	}

	// Otherwise, this is simple member access: obj.field
	memberAccess := &ast.MemberAccessExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: dotToken,
			},
		},
		Object: left,
		Member: memberName,
	}
	memberAccess.EndPos = memberName.End() // End position is after the member name

	return memberAccess
}

// parseClassConstantDeclaration parses a constant declaration within a class.
// Syntax: const Name = Value; or const Name: Type = Value;
// Also: class const Name = Value;
// The visibility parameter specifies the access level for this constant.
// The isClassConst parameter indicates if it was declared with 'class const'.
// PRE: curToken is constant name IDENT
// POST: curToken is last token of value expression
func (p *Parser) parseClassConstantDeclaration(visibility ast.Visibility, isClassConst bool) *ast.ConstDecl {
	// Current token should be the constant name identifier
	if !p.curTokenIs(lexer.IDENT) {
		p.addError("expected identifier for constant name", ErrExpectedIdent)
		return nil
	}

	constToken := p.curToken
	nameIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

	// Check for optional type annotation: const Name: Type = Value;
	var typeAnnotation *ast.TypeAnnotation
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // move to ':'
		p.nextToken() // move to type

		typeExpr := p.parseTypeExpression()
		if typeExpr != nil {
			typeAnnotation = &ast.TypeAnnotation{
				Token:      p.curToken,
				InlineType: typeExpr,
			}
		}
	}

	// Expect '=' for the constant value
	if !p.expectPeek(lexer.EQ) {
		return nil
	}

	// Parse the constant value expression
	p.nextToken()
	value := p.parseExpression(LOWEST)
	if value == nil {
		p.addError("expected constant value expression", ErrInvalidExpression)
		return nil
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	constant := &ast.ConstDecl{
		BaseNode: ast.BaseNode{
			Token:  constToken,
			EndPos: p.endPosFromToken(p.curToken),
		},
		Name:         nameIdent,
		Type:         typeAnnotation,
		Value:        value,
		Visibility:   visibility,
		IsClassConst: isClassConst,
	}

	return constant
}
