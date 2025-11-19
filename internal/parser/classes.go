package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseClassDeclaration parses a class declaration (dual-mode dispatcher).
// Syntax: type ClassName = class(Parent) ... end;
//
// Task 2.7.2: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseClassDeclaration() *ast.ClassDecl {
	return p.parseClassDeclarationCursor()
}

// parseClassDeclarationTraditional parses a class declaration with visibility sections (traditional mode).
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
// PRE: curToken is CLASS or TYPE; peekToken is class name IDENT
// POST: curToken is END
func (p *Parser) parseClassDeclarationCursor() *ast.ClassDecl {
	cursor := p.cursor

	// Expect class name identifier
	if cursor.Peek(1).Type != lexer.IDENT {
		p.addError("expected identifier for class name", ErrExpectedIdent)
		return nil
	}
	cursor = cursor.Advance() // move to IDENT
	p.cursor = cursor

	nameIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: cursor.Current(),
			},
		},
		Value: cursor.Current().Literal,
	}

	// Expect '='
	if cursor.Peek(1).Type != lexer.EQ {
		p.addError("expected '=' after class name", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to EQ
	p.cursor = cursor

	// Expect 'class' keyword
	if cursor.Peek(1).Type != lexer.CLASS {
		p.addError("expected 'class' keyword", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to CLASS
	p.cursor = cursor

	return p.parseClassDeclarationBodyCursor(nameIdent)
}

// parseClassParentAndInterfaces parses optional parent class and interfaces (dual-mode dispatcher).
// Can be called multiple times for syntax like: class abstract(TParent)
// Only updates classDecl if not already set to avoid overwriting previous parse
//
// Task 2.7.9: Cursor mode is now the only mode - dispatcher removed.
func (p *Parser) parseClassParentAndInterfaces(classDecl *ast.ClassDecl) {
	p.parseClassParentAndInterfacesCursor(classDecl)
}

// parseClassParentAndInterfacesTraditional parses optional parent class and interfaces from (...) (traditional mode).
// Can be called multiple times for syntax like: class abstract(TParent)
// Only updates classDecl if not already set to avoid overwriting previous parse
// PRE: curToken is token before parent list (CLASS, ABSTRACT, or EXTERNAL)
// POST: curToken is RPAREN if parentheses present; otherwise unchanged
func (p *Parser) parseClassParentAndInterfacesCursor(classDecl *ast.ClassDecl) {
	cursor := p.cursor

	if cursor.Peek(1).Type != lexer.LPAREN {
		return
	}

	cursor = cursor.Advance() // move to '('
	p.cursor = cursor

	// Parse comma-separated list of parent/interfaces
	identifiers := []*ast.Identifier{}

	for {
		if cursor.Peek(1).Type != lexer.IDENT {
			p.addError("expected identifier in class inheritance list", ErrExpectedIdent)
			return
		}
		cursor = cursor.Advance() // move to IDENT
		p.cursor = cursor

		identifiers = append(identifiers, &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: cursor.Current(),
				},
			},
			Value: cursor.Current().Literal,
		})

		// Check for comma (more items) or closing paren
		nextTok := cursor.Peek(1)
		if nextTok.Type == lexer.COMMA {
			cursor = cursor.Advance() // move to comma
			p.cursor = cursor
			continue
		} else if nextTok.Type == lexer.RPAREN {
			cursor = cursor.Advance() // move to ')'
			p.cursor = cursor
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

// parseClassDeclarationBody parses the body of a class declaration (dual-mode dispatcher).
// Called after 'type Name = class' has already been parsed.
//
// Task 2.7.2: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseClassDeclarationBody(nameIdent *ast.Identifier) *ast.ClassDecl {
	return p.parseClassDeclarationBodyCursor(nameIdent)
}

// parseClassDeclarationBodyTraditional parses the body of a class declaration (traditional mode).
// Called after 'type Name = class' has already been parsed.
// Current token should be 'class'.
// PRE: curToken is CLASS
// POST: curToken is END
func (p *Parser) parseClassDeclarationBodyCursor(nameIdent *ast.Identifier) *ast.ClassDecl {
	builder := p.StartNode()
	cursor := p.cursor

	classDecl := &ast.ClassDecl{
		BaseNode: ast.BaseNode{Token: cursor.Current()}, // 'class' token
		Name:     nameIdent,
	}

	// Check for optional parent class and/or interfaces
	classDecl.Interfaces = []*ast.Identifier{}

	p.parseClassParentAndInterfacesCursor(classDecl)
	cursor = p.cursor

	// Check for 'abstract' keyword
	if cursor.Peek(1).Type == lexer.ABSTRACT {
		cursor = cursor.Advance() // move to 'abstract'
		p.cursor = cursor
		classDecl.IsAbstract = true
		// Check again for parent/interfaces after abstract
		p.parseClassParentAndInterfacesCursor(classDecl)
		cursor = p.cursor
	}

	// Check for 'external' keyword
	if cursor.Peek(1).Type == lexer.EXTERNAL {
		cursor = cursor.Advance() // move to 'external'
		p.cursor = cursor
		classDecl.IsExternal = true

		// Check for optional external name string
		if cursor.Peek(1).Type == lexer.STRING {
			cursor = cursor.Advance() // move to string
			p.cursor = cursor
			classDecl.ExternalName = cursor.Current().Literal
		}
		// Check again for parent/interfaces after external
		p.parseClassParentAndInterfacesCursor(classDecl)
		cursor = p.cursor
	}

	// Check for forward declaration: type TForward = class;
	if cursor.Peek(1).Type == lexer.SEMICOLON {
		cursor = cursor.Advance() // move to semicolon
		p.cursor = cursor
		// Do NOT initialize the slices - leave them as nil so semantic analyzer
		// can detect this as a forward declaration
		return builder.Finish(classDecl).(*ast.ClassDecl)
	}

	// Parse class body (fields and methods) until 'end'
	cursor = cursor.Advance() // move past 'class' or ')' or 'abstract' or 'external' or external name
	p.cursor = cursor

	classDecl.Fields = []*ast.FieldDecl{}
	classDecl.Methods = []*ast.FunctionDecl{}
	classDecl.Operators = []*ast.OperatorDecl{}
	classDecl.Properties = []*ast.PropertyDecl{}
	classDecl.Constants = []*ast.ConstDecl{}

	// Default visibility is public
	currentVisibility := ast.VisibilityPublic

	for cursor.Current().Type != lexer.END && cursor.Current().Type != lexer.EOF {
		// Skip semicolons
		if cursor.Current().Type == lexer.SEMICOLON {
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Check for visibility section keywords
		if cursor.Current().Type == lexer.PRIVATE {
			currentVisibility = ast.VisibilityPrivate
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		} else if cursor.Current().Type == lexer.PROTECTED {
			currentVisibility = ast.VisibilityProtected
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		} else if cursor.Current().Type == lexer.PUBLIC {
			currentVisibility = ast.VisibilityPublic
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Check for 'class var', 'class const', 'class property', or 'class function' / 'class procedure'
		if cursor.Current().Type == lexer.CLASS {
			classToken := cursor.Current()
			cursor = cursor.Advance() // move past 'class'
			p.cursor = cursor

			if cursor.Current().Type == lexer.VAR {
				// Class variable: class var FieldName: Type;
				cursor = cursor.Advance() // move past 'var'
				p.cursor = cursor
				fields := p.parseFieldDeclarationsCursor(currentVisibility)
				for _, field := range fields {
					if field != nil {
						field.IsClassVar = true // Mark as class variable
						classDecl.Fields = append(classDecl.Fields, field)
					}
				}
				cursor = p.cursor
			} else if cursor.Current().Type == lexer.CONST {
				// Class constant: class const Name = Value;
				cursor = cursor.Advance() // move past 'const'
				p.cursor = cursor
				constant := p.parseClassConstantDeclarationCursor(currentVisibility, true)
				if constant != nil {
					classDecl.Constants = append(classDecl.Constants, constant)
				}
				cursor = p.cursor
			} else if cursor.Current().Type == lexer.PROPERTY {
				// Class property: class property Name: Type read GetName write SetName;
				property := p.parsePropertyDeclaration()
				if property != nil {
					property.IsClassProperty = true // Mark as class property
					classDecl.Properties = append(classDecl.Properties, property)
				}
				cursor = p.cursor
			} else if cursor.Current().Type == lexer.OPERATOR {
				operator := p.parseClassOperatorDeclarationCursor(classToken, currentVisibility)
				if operator != nil {
					classDecl.Operators = append(classDecl.Operators, operator)
				}
				cursor = p.cursor
			} else if cursor.Current().Type == lexer.FUNCTION || cursor.Current().Type == lexer.PROCEDURE || cursor.Current().Type == lexer.METHOD {
				// Class method: class function/procedure/method ...
				method := p.parseFunctionDeclarationCursor()
				if method != nil {
					method.IsClassMethod = true // Mark as class method
					method.Visibility = currentVisibility
					classDecl.Methods = append(classDecl.Methods, method)
				}
				cursor = p.cursor
			} else {
				p.addError("expected 'var', 'const', 'property', 'function', 'procedure', or 'method' after 'class' keyword", ErrUnexpectedToken)
				cursor = cursor.Advance()
				p.cursor = cursor
				continue
			}
		} else if cursor.Current().Type == lexer.CONST {
			// Regular class constant: const Name = Value;
			cursor = cursor.Advance() // move past 'const'
			p.cursor = cursor
			constant := p.parseClassConstantDeclarationCursor(currentVisibility, false)
			if constant != nil {
				classDecl.Constants = append(classDecl.Constants, constant)
			}
			cursor = p.cursor
		} else if cursor.Current().Type == lexer.IDENT && (cursor.Peek(1).Type == lexer.COLON || cursor.Peek(1).Type == lexer.COMMA || cursor.Peek(1).Type == lexer.ASSIGN || cursor.Peek(1).Type == lexer.EQ) {
			// This is a regular instance field declaration (may be comma-separated)
			// Supports: FieldName: Type; or FieldName := Value; or FieldName = Value; or FieldName: Type := Value;
			fields := p.parseFieldDeclarationsCursor(currentVisibility)
			for _, field := range fields {
				if field != nil {
					classDecl.Fields = append(classDecl.Fields, field)
				}
			}
			cursor = p.cursor
		} else if cursor.Current().Type == lexer.FUNCTION || cursor.Current().Type == lexer.PROCEDURE || cursor.Current().Type == lexer.METHOD {
			// This is a regular instance method declaration
			method := p.parseFunctionDeclarationCursor()
			if method != nil {
				method.Visibility = currentVisibility
				classDecl.Methods = append(classDecl.Methods, method)
			}
			cursor = p.cursor
		} else if cursor.Current().Type == lexer.CONSTRUCTOR {
			// This is a constructor declaration
			method := p.parseFunctionDeclarationCursor()
			if method != nil {
				method.IsConstructor = true
				method.Visibility = currentVisibility
				classDecl.Methods = append(classDecl.Methods, method)
			}
			cursor = p.cursor
		} else if cursor.Current().Type == lexer.DESTRUCTOR {
			// This is a destructor declaration
			method := p.parseFunctionDeclarationCursor()
			if method != nil {
				method.IsDestructor = true
				method.Visibility = currentVisibility
				classDecl.Methods = append(classDecl.Methods, method)
			}
			cursor = p.cursor
		} else if cursor.Current().Type == lexer.PROPERTY {
			// This is a property declaration
			property := p.parsePropertyDeclaration()
			if property != nil {
				// Note: We could track visibility here if needed
				// For now, properties are parsed without explicit visibility tracking
				classDecl.Properties = append(classDecl.Properties, property)
			}
			cursor = p.cursor
		} else if cursor.Current().Type == lexer.IDENT {
			// Unexpected identifier in class body - likely a field missing its type declaration
			p.addError("expected ':' after field name or method/property declaration keyword", ErrMissingColon)
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		} else {
			// Unknown token in class body, skip it
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		}

		cursor = cursor.Advance()
		p.cursor = cursor
	}

	// Expect 'end'
	if cursor.Current().Type != lexer.END {
		p.addError("expected 'end' to close class declaration", ErrMissingEnd)
		return nil
	}

	// Expect terminating semicolon
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after 'end'", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	return builder.Finish(classDecl).(*ast.ClassDecl)
}

// parseFieldDeclaration parses a field declaration within a class (dual-mode dispatcher).
// Syntax: FieldName: Type; or FieldName1, FieldName2, FieldName3: Type;
// Returns a slice of FieldDecl nodes (one per field name) since DWScript supports
// comma-separated field names with a single type annotation.
//
// Task 2.7.2: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseFieldDeclarations(visibility ast.Visibility) []*ast.FieldDecl {
	return p.parseFieldDeclarationsCursor(visibility)
}

// parseFieldDeclarationsTraditional parses a field declaration within a class (traditional mode).
// Syntax: FieldName: Type; or FieldName1, FieldName2, FieldName3: Type;
// The visibility parameter specifies the access level for this field.
// Returns a slice of FieldDecl nodes (one per field name) since DWScript supports
// comma-separated field names with a single type annotation.
// PRE: curToken is first field name IDENT
// POST: curToken is SEMICOLON or last token of initialization value
func (p *Parser) parseFieldDeclarationsCursor(visibility ast.Visibility) []*ast.FieldDecl {
	// Parse comma-separated field names using combinator
	// Note: IdentifierList uses parser state, so it should work with synced cursor
	fieldNames := p.IdentifierList(IdentifierListConfig{
		ErrorContext:      "field declaration",
		RequireAtLeastOne: true,
	})
	if fieldNames == nil {
		return nil
	}
	cursor := p.cursor

	// Parse optional type and/or initialization
	var fieldType ast.TypeExpression
	var initValue ast.Expression

	// Check for type annotation (: Type) or direct initialization (:= Value)
	if cursor.Peek(1).Type == lexer.COLON {
		cursor = cursor.Advance() // move to ':'
		cursor = cursor.Advance() // move to type
		p.cursor = cursor

		// Parse type expression (supports simple types, array types, function pointer types)
		fieldType = p.parseTypeExpressionCursor()
		if fieldType == nil {
			return nil
		}
		cursor = p.cursor

		// Parse optional field initializer after type
		initValue = p.parseFieldInitializer(fieldNames)
		cursor = p.cursor
	} else if cursor.Peek(1).Type == lexer.ASSIGN || cursor.Peek(1).Type == lexer.EQ {
		// Direct field initialization without type annotation: FField := Value; or FField = Value;
		// This is a shorthand syntax for field initialization from constants
		if len(fieldNames) > 1 {
			p.addError("initialization without type annotation not allowed for comma-separated field declarations", ErrInvalidExpression)
			return nil
		}

		cursor = cursor.Advance() // move to ':=' or '='
		cursor = cursor.Advance() // move to value expression
		p.cursor = cursor

		// Parse initialization expression
		initValue = p.parseExpressionCursor(LOWEST)
		if initValue == nil {
			p.addError("expected initialization expression after := or =", ErrInvalidExpression)
			return nil
		}
		cursor = p.cursor
	}

	// Expect semicolon
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after field declaration", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

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

// parseMemberAccess parses member access and method call expressions (dual-mode dispatcher).
// Handles obj.field, obj.method(), and TClass.Create() syntax.
// This is registered as an infix operator for the DOT token.
//
// Task 2.7.2: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseMemberAccess(left ast.Expression) ast.Expression {
	return p.parseMemberAccessCursor(left)
}

// parseMemberAccessTraditional parses member access and method call expressions (traditional mode).
// Handles obj.field, obj.method(), and TClass.Create() syntax.
// This is registered as an infix operator for the DOT token.
// PRE: curToken is DOT
// POST: curToken is member name IDENT, RPAREN (for method calls), or last token of right operand
func (p *Parser) parseMemberAccessCursor(left ast.Expression) ast.Expression {
	builder := p.StartNode()

	dotToken := p.cursor.Current() // Save the '.' token

	// Advance to the member name
	p.cursor = p.cursor.Advance()
	memberToken := p.cursor.Current()

	// The member name can be an identifier or a keyword (DWScript allows keywords as member names)
	// But it cannot be operators, numbers, or other invalid tokens
	if memberToken.Type == lexer.SEMICOLON || memberToken.Type == lexer.INT ||
		memberToken.Type == lexer.FLOAT || memberToken.Type == lexer.STRING ||
		memberToken.Type == lexer.LPAREN || memberToken.Type == lexer.RPAREN ||
		memberToken.Type == lexer.LBRACK || memberToken.Type == lexer.RBRACK ||
		memberToken.Type == lexer.COMMA || memberToken.Type == lexer.EOF {
		p.addError("expected identifier after '.'", ErrExpectedIdent)
		return nil
	}

	memberName := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: memberToken,
			},
		},
		Value: memberToken.Literal,
	}

	// Check if this is a method call (followed by '(')
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.LPAREN {
		// Check if this is object creation: TClass.Create()
		if ident, ok := left.(*ast.Identifier); ok && memberName.Value == "Create" {
			// This is a NewExpression
			p.cursor = p.cursor.Advance() // move to '('

			newExpr := &ast.NewExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: ident.Token,
					},
				},
				ClassName: ident,
				Arguments: []ast.Expression{},
			}

			// Parse arguments - cursor will be at RPAREN after parseExpressionListCursor
			newExpr.Arguments = p.parseExpressionListCursor(lexer.RPAREN)

			return builder.Finish(newExpr).(*ast.NewExpression)
		}

		// Regular method call: obj.Method()
		p.cursor = p.cursor.Advance() // move to '('

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

		// Parse arguments - cursor will be at RPAREN after parseExpressionListCursor
		methodCall.Arguments = p.parseExpressionListCursor(lexer.RPAREN)

		return builder.Finish(methodCall).(*ast.MethodCallExpression)
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

	return builder.FinishWithNode(memberAccess, memberName).(*ast.MemberAccessExpression)
}

// parseClassConstantDeclaration parses a constant declaration within a class (dual-mode dispatcher).
// Syntax: const Name = Value; or const Name: Type = Value;
// Also: class const Name = Value;
//
// Task 2.7.2: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseClassConstantDeclaration(visibility ast.Visibility, isClassConst bool) *ast.ConstDecl {
	return p.parseClassConstantDeclarationCursor(visibility, isClassConst)
}

// parseClassConstantDeclarationTraditional parses a constant declaration within a class (traditional mode).
// Syntax: const Name = Value; or const Name: Type = Value;
// Also: class const Name = Value;
// The visibility parameter specifies the access level for this constant.
// The isClassConst parameter indicates if it was declared with 'class const'.
// PRE: curToken is constant name IDENT
// POST: curToken is last token of value expression
func (p *Parser) parseClassConstantDeclarationCursor(visibility ast.Visibility, isClassConst bool) *ast.ConstDecl {
	builder := p.StartNode()
	cursor := p.cursor

	// Current token should be the constant name identifier
	if cursor.Current().Type != lexer.IDENT {
		p.addError("expected identifier for constant name", ErrExpectedIdent)
		return nil
	}

	constToken := cursor.Current()
	nameIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: cursor.Current(),
			},
		},
		Value: cursor.Current().Literal,
	}

	// Check for optional type annotation: const Name: Type = Value;
	var typeAnnotation *ast.TypeAnnotation
	if cursor.Peek(1).Type == lexer.COLON {
		cursor = cursor.Advance() // move to ':'
		cursor = cursor.Advance() // move to type
		p.cursor = cursor

		typeExpr := p.parseTypeExpressionCursor()
		if typeExpr != nil {
			cursor = p.cursor
			typeAnnotation = &ast.TypeAnnotation{
				Token:      cursor.Current(),
				InlineType: typeExpr,
			}
		}
	}

	// Expect '=' for the constant value
	if cursor.Peek(1).Type != lexer.EQ {
		p.addError("expected '=' after constant name", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to '='
	p.cursor = cursor

	// Parse the constant value expression
	cursor = cursor.Advance() // move to value expression
	p.cursor = cursor

	value := p.parseExpressionCursor(LOWEST)
	if value == nil {
		p.addError("expected constant value expression", ErrInvalidExpression)
		return nil
	}
	cursor = p.cursor

	// Expect semicolon
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after constant value", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	constant := &ast.ConstDecl{
		BaseNode: ast.BaseNode{
			Token: constToken,
		},
		Name:         nameIdent,
		Type:         typeAnnotation,
		Value:        value,
		Visibility:   visibility,
		IsClassConst: isClassConst,
	}

	return builder.Finish(constant).(*ast.ConstDecl)
}
