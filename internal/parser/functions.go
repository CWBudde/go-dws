package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseFunctionDeclaration parses a function or procedure declaration.
// Syntax: function Name(params): Type; begin ... end;
//
//	procedure Name(params); begin ... end;
func (p *Parser) parseFunctionDeclaration() *ast.FunctionDecl {
	fn := &ast.FunctionDecl{Token: p.curToken}

	// Parse function name (may be qualified: ClassName.MethodName)
	// In DWScript/Object Pascal, keywords can be used as identifiers in certain contexts
	// like method names, so we accept any token as a name here
	p.nextToken()
	firstIdent := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check for qualified name (ClassName.MethodName for method implementations)
	if p.peekTokenIs(lexer.DOT) {
		p.nextToken() // move to '.'
		p.nextToken() // move past '.'
		// This is a qualified name: TExample.MethodName
		// firstIdent is the class name, current token is the method name
		fn.ClassName = firstIdent
		fn.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	} else {
		// Simple function name (not a method implementation)
		fn.Name = firstIdent
		fn.ClassName = nil
	}

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
		p.nextToken() // move past ':' to type expression start token

		// Task 9.59: Support inline array types in return types
		// Parse type expression (can be simple type, function pointer, or array type)
		typeExpr := p.parseTypeExpression()
		if typeExpr == nil {
			p.addError("expected return type after ':'")
			return nil
		}

		// Convert TypeExpression to TypeAnnotation for FunctionDecl.ReturnType
		// TODO: Update FunctionDecl struct to accept TypeExpression instead of TypeAnnotation
		switch te := typeExpr.(type) {
		case *ast.TypeAnnotation:
			fn.ReturnType = te
		case *ast.FunctionPointerTypeNode:
			// For function pointer types, we create a synthetic TypeAnnotation
			fn.ReturnType = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(), // Use the full function pointer signature as the type name
			}
		case *ast.ArrayTypeNode:
			// For array types, we create a synthetic TypeAnnotation
			// Check if Token is nil to prevent panics (defensive programming)
			if te == nil {
				p.addError("array type expression is nil in return type")
				return nil
			}
			// Use the array token or create a dummy token if nil
			token := te.Token
			if token.Type == 0 || token.Literal == "" {
				// Create a dummy token to prevent nil pointer issues
				token = lexer.Token{Type: lexer.ARRAY, Literal: "array", Pos: lexer.Position{}}
			}
			fn.ReturnType = &ast.TypeAnnotation{
				Token: token,
				Name:  te.String(), // Use the full array type signature as the type name
			}
		default:
			p.addError("unsupported type expression in return type")
			return nil
		}
	}

	// Expect semicolon after signature
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Check for optional directives: static, virtual, override, external
	for {
		if p.peekTokenIs(lexer.STATIC) {
			p.nextToken() // move to 'static'
			// Note: IsClassMethod flag should have been set by the caller (parseClassDeclaration)
			// The 'static' keyword is optional and doesn't change the semantics
			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}
		} else if p.peekTokenIs(lexer.VIRTUAL) {
			p.nextToken() // move to 'virtual'
			fn.IsVirtual = true
			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}
		} else if p.peekTokenIs(lexer.OVERRIDE) {
			p.nextToken() // move to 'override'
			fn.IsOverride = true
			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}
		} else if p.peekTokenIs(lexer.ABSTRACT) {
			// Abstract method: function GetArea(): Float; abstract;
			// Task 7.65d - Parse abstract method declarations (no body)
			p.nextToken() // move to 'abstract'
			fn.IsAbstract = true
			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}
			// Abstract methods have no body, return early
			return fn
		} else if p.peekTokenIs(lexer.EXTERNAL) {
			// External method: procedure Hello; external 'world';
			// Task 7.140 - Parse external method declarations (no body)
			p.nextToken() // move to 'external'
			fn.IsExternal = true

			// Check for optional external name string
			if p.peekTokenIs(lexer.STRING) {
				p.nextToken() // move to string
				fn.ExternalName = p.curToken.Literal
			}

			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}
			// External methods have no body, return early
			return fn
		} else {
			break // No more directives
		}
	}

	// Parse preconditions (require block) if present
	if p.peekTokenIs(lexer.REQUIRE) {
		p.nextToken() // move to REQUIRE
		fn.PreConditions = p.parsePreConditions()
	}

	// Check if this is a forward declaration (no body)
	// Forward declarations end with a semicolon instead of begin...end or local declarations
	if !p.peekTokenIs(lexer.BEGIN) && !p.peekTokenIs(lexer.VAR) && !p.peekTokenIs(lexer.CONST) && !p.peekTokenIs(lexer.REQUIRE) {
		// This is a forward declaration (method declaration in class body)
		// Body will be provided later in method implementation outside class
		return fn
	}

	// Parse local variable/constant declarations (optional)
	// Syntax: var x: Integer; y: String; ... or const X = 5; ...
	for p.peekTokenIs(lexer.VAR) || p.peekTokenIs(lexer.CONST) {
		p.nextToken() // move to 'var' or 'const'

		if p.curTokenIs(lexer.VAR) {
			// Parse multiple variable declarations until we hit something else
			// First iteration: cur=VAR, peek=first_identifier
			// Subsequent iterations: cur=identifier_N, peek=colon
			for {
				// Check if there's an identifier to parse (either in peek for first iteration,
				// or in cur for subsequent iterations after we've advanced)
				if p.curTokenIs(lexer.VAR) {
					// First iteration - identifier is in peek position
					if !p.peekTokenIs(lexer.IDENT) {
						break
					}
				} else if !p.curTokenIs(lexer.IDENT) {
					// Subsequent iterations - we should be at an identifier
					break
				}

				// Parse one variable declaration
				varDecl := p.parseVarDeclaration()
				if varDecl == nil {
					break
				}
				// Add to function body as a local declaration
				if fn.Body == nil {
					fn.Body = &ast.BlockStatement{Token: p.curToken}
				}
				fn.Body.Statements = append(fn.Body.Statements, varDecl)

				// parseVarDeclaration() leaves us at the semicolon (cur=`;`)
				// Check if there's another identifier after the semicolon before advancing
				if !p.peekTokenIs(lexer.IDENT) {
					// No more variable declarations, stop here with cur at semicolon
					break
				}

				// Advance past semicolon to the next identifier
				p.nextToken()

				// Now cur is the next identifier, loop back to parse it
			}
		} else if p.curTokenIs(lexer.CONST) {
			// Parse constant declaration
			constDecl := p.parseConstDeclaration()
			if constDecl != nil {
				if fn.Body == nil {
					fn.Body = &ast.BlockStatement{Token: p.curToken}
				}
				fn.Body.Statements = append(fn.Body.Statements, constDecl)
			}
		}
	}

	// Parse function body (begin...end block)
	if !p.expectPeek(lexer.BEGIN) {
		return nil
	}

	bodyBlock := p.parseBlockStatement()
	if bodyBlock != nil {
		if fn.Body == nil {
			// No local declarations, use the body block directly
			fn.Body = bodyBlock
		} else {
			// Append body statements to local declarations
			fn.Body.Statements = append(fn.Body.Statements, bodyBlock.Statements...)
		}
	}

	// Expect semicolon after end
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Parse postconditions (ensure block) if present
	if p.peekTokenIs(lexer.ENSURE) {
		p.nextToken() // move to ENSURE
		fn.PostConditions = p.parsePostConditions()
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
// Syntax: name: Type  or  name1, name2, name3: Type  or  var name: Type  or  lazy name: Type  or  const name: Type
func (p *Parser) parseParameterGroup() []*ast.Parameter {
	params := []*ast.Parameter{}

	// Check for 'const' keyword (pass by const-reference)
	isConst := false
	if p.curTokenIs(lexer.CONST) {
		isConst = true
		p.nextToken() // move past 'const'
	}

	// Check for 'lazy' keyword (expression capture)
	isLazy := false
	if p.curTokenIs(lexer.LAZY) {
		isLazy = true
		p.nextToken() // move past 'lazy'
	}

	// Check for 'var' keyword (pass by reference)
	byRef := false
	if p.curTokenIs(lexer.VAR) {
		byRef = true
		p.nextToken() // move past 'var'
	}

	// Check for mutually exclusive modifiers
	if (isLazy && byRef) || (isConst && byRef) || (isConst && isLazy) {
		p.addError("parameter modifiers are mutually exclusive")
		return nil
	}

	// Collect parameter names separated by commas
	names := []*ast.Identifier{}

	for {
		// Parse parameter name (can be IDENT or contextual keywords like STEP)
		if !p.isIdentifierToken(p.curToken.Type) {
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

	// Parse type expression (can be simple type, function pointer, or array type)
	// Task 9.44: Changed from simple IDENT to parseTypeExpression() to support inline types
	// expectPeek() has already advanced us to the COLON token
	// Now advance past it to get to the type expression
	p.nextToken() // move past COLON to type expression start token
	typeExpr := p.parseTypeExpression()
	if typeExpr == nil {
		p.addError("expected type expression after ':'")
		return nil
	}

	// For now, we need to convert TypeExpression to TypeAnnotation for Parameter.Type
	// TODO: Update Parameter struct to accept TypeExpression instead of TypeAnnotation
	var typeAnnotation *ast.TypeAnnotation
	switch te := typeExpr.(type) {
	case *ast.TypeAnnotation:
		typeAnnotation = te
	case *ast.FunctionPointerTypeNode:
		// For function pointer types, we create a synthetic TypeAnnotation
		// The semantic analyzer will recognize function pointer parameters by checking the type string
		typeAnnotation = &ast.TypeAnnotation{
			Token: te.Token,
			Name:  te.String(), // Use the full function pointer signature as the type name
		}
	case *ast.ArrayTypeNode:
		// For array types, we create a synthetic TypeAnnotation
		// Check if Token is nil to prevent panics (defensive programming)
		if te == nil {
			p.addError("array type expression is nil in parameter type")
			return nil
		}
		// Use the array token or create a dummy token if nil
		token := te.Token
		if token.Type == 0 || token.Literal == "" {
			// Create a dummy token to prevent nil pointer issues
			token = lexer.Token{Type: lexer.ARRAY, Literal: "array", Pos: lexer.Position{}}
		}
		typeAnnotation = &ast.TypeAnnotation{
			Token: token,
			Name:  te.String(), // Use the full array type signature as the type name
		}
	default:
		p.addError("unsupported type expression in parameter")
		return nil
	}

	// Create a parameter for each name with the same type
	for _, name := range names {
		param := &ast.Parameter{
			Token:   name.Token,
			Name:    name,
			Type:    typeAnnotation,
			IsLazy:  isLazy,
			ByRef:   byRef,
			IsConst: isConst,
		}
		params = append(params, param)
	}

	return params
}
