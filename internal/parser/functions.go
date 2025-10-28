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

	// Check if this is a forward declaration (no body)
	// Forward declarations end with a semicolon instead of begin...end or local declarations
	if !p.peekTokenIs(lexer.BEGIN) && !p.peekTokenIs(lexer.VAR) && !p.peekTokenIs(lexer.CONST) {
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
			for p.peekTokenIs(lexer.IDENT) {
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

				// Continue if next token is an identifier (another var declaration)
				if !p.peekTokenIs(lexer.IDENT) {
					break
				}
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
