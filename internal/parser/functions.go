package parser

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// isCallingConvention checks if a string is a calling convention keyword.
// Calling conventions are contextual identifiers, not reserved keywords,
// so they're tokenized as IDENT by the lexer.
func isCallingConvention(literal string) bool {
	lower := strings.ToLower(literal)
	return lower == "register" || lower == "pascal" || lower == "cdecl" ||
		lower == "safecall" || lower == "stdcall" || lower == "fastcall" ||
		lower == "reference"
}

// parseFunctionDeclaration is a dispatcher that routes to the appropriate implementation
// based on the parser mode (traditional vs cursor).
//
// Task 2.7.3: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseFunctionDeclaration() *ast.FunctionDecl {
	if p.useCursor {
		return p.parseFunctionDeclarationCursor()
	}
	return p.parseFunctionDeclarationTraditional()
}

// parseFunctionDeclarationTraditional parses a function or procedure declaration.
// Syntax: function Name(params): Type; begin ... end;
//
//	procedure Name(params); begin ... end;
//
// PRE: curToken is FUNCTION or PROCEDURE
// POST: curToken is END or SEMICOLON (forward declaration) or last token of body
// Task 2.7.3: Renamed to enable dual-mode operation
func (p *Parser) parseFunctionDeclarationTraditional() *ast.FunctionDecl {
	builder := p.StartNode()

	fn := &ast.FunctionDecl{
		BaseNode: ast.BaseNode{
			Token: p.curToken,
		},
	}

	// Parse function name (may be qualified: ClassName.MethodName)
	// In DWScript/Object Pascal, keywords can be used as identifiers in certain contexts
	// like method names, so we accept any token as a name here
	p.nextToken()
	firstIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

	// Check for qualified name (ClassName.MethodName for method implementations)
	if p.peekTokenIs(lexer.DOT) {
		p.nextToken() // move to '.'
		p.nextToken() // move past '.'
		// This is a qualified name: TExample.MethodName
		// firstIdent is the class name, current token is the method name
		fn.ClassName = firstIdent
		fn.Name = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: p.curToken,
				},
			},
			Value: p.curToken.Literal,
		}
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
			p.addError("expected ')' after parameter list", ErrMissingRParen)
			return nil
		}
	}

	// Parse return type for functions (not procedures)
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // move to ':'
		p.nextToken() // move past ':' to type expression start token

		// Support inline array types in return types
		// Parse type expression (can be simple type, function pointer, or array type)
		typeExpr := p.parseTypeExpression()
		if typeExpr == nil {
			p.addError("expected return type after ':'", ErrExpectedType)
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
		case *ast.SetTypeNode:
			// For set types, we create a synthetic TypeAnnotation
			// Handle inline set type expressions in return types
			fn.ReturnType = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(), // Use the full set type signature as the type name
			}
		case *ast.ArrayTypeNode:
			// For array types, we create a synthetic TypeAnnotation
			// Check if Token is nil to prevent panics (defensive programming)
			if te == nil {
				p.addError("array type expression is nil in return type", ErrInvalidType)
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
			p.addError("unsupported type expression in return type", ErrInvalidType)
			return nil
		}
	}

	// Expect semicolon after signature
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Check for optional directives: static, virtual, override, abstract, external, overload, calling conventions
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
		} else if p.peekTokenIs(lexer.REINTRODUCE) {
			p.nextToken() // move to 'reintroduce'
			fn.IsReintroduce = true
			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}
		} else if p.peekTokenIs(lexer.DEFAULT) {
			p.nextToken() // move to 'default'
			fn.IsDefault = true
			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}
		} else if p.peekTokenIs(lexer.ABSTRACT) {
			// Abstract method: function GetArea(): Float; abstract;
			p.nextToken() // move to 'abstract'
			fn.IsAbstract = true
			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}
			// Continue parsing directives (e.g., overload)
			// Don't return early here
		} else if p.peekTokenIs(lexer.EXTERNAL) {
			// External method: procedure Hello; external 'world';
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
			// Continue parsing directives (e.g., overload)
			// Don't return early here
		} else if p.peekTokenIs(lexer.OVERLOAD) {
			// Overload directive: function Test(x: Integer): Float; overload;
			p.nextToken() // move to 'overload'
			fn.IsOverload = true
			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}
		} else if p.peekTokenIs(lexer.FORWARD) {
			// Forward directive: function Test(x: Integer): Float; forward;
			p.nextToken() // move to 'forward'
			fn.IsForward = true
			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}
			// Forward declarations have no body, so we can return early
			// But continue to allow combined directives like "overload; forward;"
		} else if p.peekToken.Type == lexer.IDENT && isCallingConvention(p.peekToken.Literal) {
			// Calling convention directives: register, pascal, cdecl, safecall, stdcall, fastcall, reference
			// Syntax: procedure Test; register;
			// Note: These are contextual identifiers, not reserved keywords
			p.nextToken() // move to calling convention
			fn.CallingConvention = strings.ToLower(p.curToken.Literal)
			fn.CallingConventionPos = p.curToken.Pos
			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}
		} else if p.peekTokenIs(lexer.DEPRECATED) {
			// Deprecated directive: function Test(): Integer; deprecated;
			// Syntax: procedure Test; deprecated 'message';
			p.nextToken() // move to 'deprecated'
			fn.IsDeprecated = true

			// Check for optional deprecation message string
			if p.peekTokenIs(lexer.STRING) {
				p.nextToken() // move to string
				fn.DeprecatedMessage = p.curToken.Literal
			}

			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}
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
	// Forward declarations either:
	//   1. Have explicit forward directive (fn.IsForward = true)
	//   2. Or implicitly end with semicolon (no begin/var/const/require)
	if fn.IsForward || (!p.peekTokenIs(lexer.BEGIN) && !p.peekTokenIs(lexer.VAR) && !p.peekTokenIs(lexer.CONST) && !p.peekTokenIs(lexer.REQUIRE)) {
		// This is a forward declaration (or method declaration in class body)
		// Body will be provided later in implementation
		// End position is at the last semicolon we consumed
		return builder.Finish(fn).(*ast.FunctionDecl)
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
					fn.Body = &ast.BlockStatement{
						BaseNode: ast.BaseNode{Token: p.curToken},
					}
				}
				// If parseVarDeclaration() wrapped multiple declarations in a BlockStatement,
				// unwrap it to avoid creating an extra nested scope in the semantic analyzer
				if blockStmt, ok := varDecl.(*ast.BlockStatement); ok && p.isVarDeclBlock(blockStmt) {
					// Add each var declaration individually
					fn.Body.Statements = append(fn.Body.Statements, blockStmt.Statements...)
				} else {
					fn.Body.Statements = append(fn.Body.Statements, varDecl)
				}

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
					fn.Body = &ast.BlockStatement{
						BaseNode: ast.BaseNode{Token: p.curToken},
					}
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

	// Check if we stopped at ENSURE inside the begin...end block
	if p.curTokenIs(lexer.ENSURE) {
		fn.PostConditions = p.parsePostConditions()
		// After parsing postconditions, skip any semicolons
		for p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}
		// Now check if we need to advance to END or if we're already there
		if !p.curTokenIs(lexer.END) {
			if !p.expectPeek(lexer.END) {
				return nil
			}
		}
	}

	// Expect semicolon after end
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Parse postconditions (ensure block) if present AFTER the end keyword
	if p.peekTokenIs(lexer.ENSURE) {
		// Check if postconditions were already defined inline
		if fn.PostConditions != nil {
			p.addError("postconditions already defined inline; cannot define them again after 'end'", ErrInvalidSyntax)
			return nil
		}
		p.nextToken() // move to ENSURE
		fn.PostConditions = p.parsePostConditions()
		// End position is after the postconditions
		if fn.PostConditions != nil {
			return builder.FinishWithNode(fn, fn.PostConditions).(*ast.FunctionDecl)
		} else {
			// Fallback if postconditions failed to parse
			return builder.Finish(fn).(*ast.FunctionDecl)
		}
	} else {
		// No postconditions - end position is at the semicolon after 'end'
		return builder.Finish(fn).(*ast.FunctionDecl)
	}
}

// parseFunctionDeclarationCursor parses a function or procedure declaration (cursor mode).
// PRE: cursor is at FUNCTION or PROCEDURE
// POST: cursor is at END or SEMICOLON or last token of postconditions
//
// Task 2.7.3: New cursor-based implementation for immutable parsing.
func (p *Parser) parseFunctionDeclarationCursor() *ast.FunctionDecl {
	cursor := p.cursor
	builder := p.StartNode()

	fn := &ast.FunctionDecl{
		BaseNode: ast.BaseNode{
			Token: cursor.Current(),
		},
	}

	// Parse function name (may be qualified: ClassName.MethodName)
	cursor = cursor.Advance() // move to name
	p.cursor = cursor

	firstIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: cursor.Current(),
			},
		},
		Value: cursor.Current().Literal,
	}

	// Check for qualified name (ClassName.MethodName for method implementations)
	if cursor.Peek(1).Type == lexer.DOT {
		cursor = cursor.Advance() // move to '.'
		cursor = cursor.Advance() // move past '.'
		p.cursor = cursor

		fn.ClassName = firstIdent
		fn.Name = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: cursor.Current(),
				},
			},
			Value: cursor.Current().Literal,
		}
	} else {
		fn.Name = firstIdent
		fn.ClassName = nil
	}

	// Parse parameter list (if present)
	if cursor.Peek(1).Type == lexer.LPAREN {
		cursor = cursor.Advance() // move to '('
		p.cursor = cursor

		fn.Parameters = p.parseParameterList()
		cursor = p.cursor

		if cursor.Current().Type != lexer.RPAREN {
			p.addError("expected ')' after parameter list", ErrMissingRParen)
			return nil
		}
	}

	// Parse return type for functions (not procedures)
	if cursor.Peek(1).Type == lexer.COLON {
		cursor = cursor.Advance() // move to ':'
		cursor = cursor.Advance() // move past ':' to type expression
		p.cursor = cursor

		typeExpr := p.parseTypeExpression()
		if typeExpr == nil {
			p.addError("expected return type after ':'", ErrExpectedType)
			return nil
		}

		cursor = p.cursor

		// Convert TypeExpression to TypeAnnotation
		switch te := typeExpr.(type) {
		case *ast.TypeAnnotation:
			fn.ReturnType = te
		case *ast.FunctionPointerTypeNode:
			fn.ReturnType = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(),
			}
		case *ast.SetTypeNode:
			fn.ReturnType = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(),
			}
		case *ast.ArrayTypeNode:
			if te == nil {
				p.addError("array type expression is nil in return type", ErrInvalidType)
				return nil
			}
			token := te.Token
			if token.Type == 0 || token.Literal == "" {
				token = lexer.Token{Type: lexer.ARRAY, Literal: "array", Pos: lexer.Position{}}
			}
			fn.ReturnType = &ast.TypeAnnotation{
				Token: token,
				Name:  te.String(),
			}
		default:
			p.addError("unsupported type expression in return type", ErrInvalidType)
			return nil
		}
	}

	// Expect semicolon after signature
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after function signature", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	// Parse directives loop
	for {
		nextTok := cursor.Peek(1)

		if nextTok.Type == lexer.STATIC {
			cursor = cursor.Advance() // move to 'static'
			p.cursor = cursor

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after static", ErrMissingSemicolon)
				return nil
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.VIRTUAL {
			cursor = cursor.Advance() // move to 'virtual'
			p.cursor = cursor
			fn.IsVirtual = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after virtual", ErrMissingSemicolon)
				return nil
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.OVERRIDE {
			cursor = cursor.Advance() // move to 'override'
			p.cursor = cursor
			fn.IsOverride = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after override", ErrMissingSemicolon)
				return nil
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.REINTRODUCE {
			cursor = cursor.Advance() // move to 'reintroduce'
			p.cursor = cursor
			fn.IsReintroduce = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after reintroduce", ErrMissingSemicolon)
				return nil
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.DEFAULT {
			cursor = cursor.Advance() // move to 'default'
			p.cursor = cursor
			fn.IsDefault = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after default", ErrMissingSemicolon)
				return nil
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.ABSTRACT {
			cursor = cursor.Advance() // move to 'abstract'
			p.cursor = cursor
			fn.IsAbstract = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after abstract", ErrMissingSemicolon)
				return nil
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.EXTERNAL {
			cursor = cursor.Advance() // move to 'external'
			p.cursor = cursor
			fn.IsExternal = true

			// Check for optional external name string
			if cursor.Peek(1).Type == lexer.STRING {
				cursor = cursor.Advance() // move to string
				p.cursor = cursor
				fn.ExternalName = cursor.Current().Literal
			}

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after external", ErrMissingSemicolon)
				return nil
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.OVERLOAD {
			cursor = cursor.Advance() // move to 'overload'
			p.cursor = cursor
			fn.IsOverload = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after overload", ErrMissingSemicolon)
				return nil
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.FORWARD {
			cursor = cursor.Advance() // move to 'forward'
			p.cursor = cursor
			fn.IsForward = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after forward", ErrMissingSemicolon)
				return nil
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.IDENT && isCallingConvention(nextTok.Literal) {
			cursor = cursor.Advance() // move to calling convention
			p.cursor = cursor
			fn.CallingConvention = strings.ToLower(cursor.Current().Literal)
			fn.CallingConventionPos = cursor.Current().Pos

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after calling convention", ErrMissingSemicolon)
				return nil
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.DEPRECATED {
			cursor = cursor.Advance() // move to 'deprecated'
			p.cursor = cursor
			fn.IsDeprecated = true

			// Check for optional deprecation message
			if cursor.Peek(1).Type == lexer.STRING {
				cursor = cursor.Advance() // move to string
				p.cursor = cursor
				fn.DeprecatedMessage = cursor.Current().Literal
			}

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after deprecated", ErrMissingSemicolon)
				return nil
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else {
			break // No more directives
		}
	}

	// Parse preconditions (require block) if present
	if cursor.Peek(1).Type == lexer.REQUIRE {
		cursor = cursor.Advance() // move to REQUIRE
		p.cursor = cursor
		fn.PreConditions = p.parsePreConditions()
		cursor = p.cursor
	}

	// Check if this is a forward declaration (no body)
	nextTok := cursor.Peek(1)
	if fn.IsForward || (nextTok.Type != lexer.BEGIN && nextTok.Type != lexer.VAR && nextTok.Type != lexer.CONST && nextTok.Type != lexer.REQUIRE) {
		return builder.Finish(fn).(*ast.FunctionDecl)
	}

	// Parse local variable/constant declarations
	for cursor.Peek(1).Type == lexer.VAR || cursor.Peek(1).Type == lexer.CONST {
		cursor = cursor.Advance() // move to 'var' or 'const'
		p.cursor = cursor

		if cursor.Current().Type == lexer.VAR {
			// Parse multiple variable declarations
			for {
				if cursor.Current().Type == lexer.VAR {
					if cursor.Peek(1).Type != lexer.IDENT {
						break
					}
				} else if cursor.Current().Type != lexer.IDENT {
					break
				}

				varDecl := p.parseVarDeclaration()
				if varDecl == nil {
					break
				}
				cursor = p.cursor

				if fn.Body == nil {
					fn.Body = &ast.BlockStatement{
						BaseNode: ast.BaseNode{Token: cursor.Current()},
					}
				}

				if blockStmt, ok := varDecl.(*ast.BlockStatement); ok && p.isVarDeclBlock(blockStmt) {
					fn.Body.Statements = append(fn.Body.Statements, blockStmt.Statements...)
				} else {
					fn.Body.Statements = append(fn.Body.Statements, varDecl)
				}

				if cursor.Peek(1).Type != lexer.IDENT {
					break
				}

				cursor = cursor.Advance() // move to next identifier
				p.cursor = cursor
			}
		} else if cursor.Current().Type == lexer.CONST {
			constDecl := p.parseConstDeclaration()
			if constDecl != nil {
				if fn.Body == nil {
					fn.Body = &ast.BlockStatement{
						BaseNode: ast.BaseNode{Token: cursor.Current()},
					}
				}
				fn.Body.Statements = append(fn.Body.Statements, constDecl)
			}
			cursor = p.cursor
		}
	}

	// Parse function body (begin...end block)
	if cursor.Peek(1).Type != lexer.BEGIN {
		p.addError("expected 'begin' for function body", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to BEGIN
	p.cursor = cursor

	bodyBlock := p.parseBlockStatement()
	cursor = p.cursor

	if bodyBlock != nil {
		if fn.Body == nil {
			fn.Body = bodyBlock
		} else {
			fn.Body.Statements = append(fn.Body.Statements, bodyBlock.Statements...)
		}
	}

	// Check if we stopped at ENSURE inside the begin...end block
	if cursor.Current().Type == lexer.ENSURE {
		fn.PostConditions = p.parsePostConditions()
		cursor = p.cursor

		// Skip semicolons
		for cursor.Current().Type == lexer.SEMICOLON {
			cursor = cursor.Advance()
			p.cursor = cursor
		}

		// Advance to END if not already there
		if cursor.Current().Type != lexer.END {
			if cursor.Peek(1).Type != lexer.END {
				p.addError("expected 'end' after postconditions", ErrUnexpectedToken)
				return nil
			}
			cursor = cursor.Advance() // move to END
			p.cursor = cursor
		}
	}

	// Expect semicolon after end
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after 'end'", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	// Parse postconditions after end (if present and not already parsed inline)
	if cursor.Peek(1).Type == lexer.ENSURE {
		if fn.PostConditions != nil {
			p.addError("postconditions already defined inline; cannot define them again after 'end'", ErrInvalidSyntax)
			return nil
		}
		cursor = cursor.Advance() // move to ENSURE
		p.cursor = cursor

		fn.PostConditions = p.parsePostConditions()
		if fn.PostConditions != nil {
			return builder.FinishWithNode(fn, fn.PostConditions).(*ast.FunctionDecl)
		} else {
			return builder.Finish(fn).(*ast.FunctionDecl)
		}
	} else {
		return builder.Finish(fn).(*ast.FunctionDecl)
	}
}

// parseParameterList parses a function parameter list.
// Syntax: (param: Type; var param: Type; a, b, c: Type)
// PRE: curToken is LPAREN
// POST: curToken is RPAREN
func (p *Parser) parseParameterList() []*ast.Parameter {
	params := []*ast.Parameter{}

	p.nextToken() // move past '('

	if p.curTokenIs(lexer.RPAREN) {
		return params
	}

	// Use SeparatedList combinator (Task 2.3.2)
	p.SeparatedList(SeparatorConfig{
		Sep:         lexer.SEMICOLON,
		Term:        lexer.RPAREN,
		AllowEmpty:  true,
		RequireTerm: true,
		ParseItem: func() bool {
			groupParams := p.parseParameterGroup()
			if groupParams != nil {
				params = append(params, groupParams...)
				return true
			}
			return false
		},
	})

	return params
}

// parseParameterGroup parses a group of parameters with the same type.
// Syntax: name: Type  or  name1, name2, name3: Type  or  var name: Type  or  lazy name: Type  or  const name: Type
// PRE: curToken is VAR, CONST, LAZY, or first parameter name IDENT
// POST: curToken is last token of type expression or default value
func (p *Parser) parseParameterGroup() []*ast.Parameter {
	// Use the ParameterGroup combinator (Task 2.3.3)
	return p.ParameterGroup(ParameterGroupConfig{
		AllowModifiers: true,
		AllowDefaults:  true,
		ErrorContext:   "function parameter",
	})
}

// parseParameterListAtToken is a dispatcher that routes to the appropriate implementation
// based on the parser mode (traditional vs cursor).
//
// Task 2.7.3: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseParameterListAtToken() []*ast.Parameter {
	if p.useCursor {
		return p.parseParameterListAtTokenCursor()
	}
	return p.parseParameterListAtTokenTraditional()
}

// parseParameterListAtTokenTraditional parses a full parameter list with names when already
// positioned at the first parameter token (not at LPAREN).
// This is a wrapper used by function pointer type parsing.
// Syntax: name: Type; name2: Type; ...
// PRE: curToken is first parameter token (VAR, CONST, LAZY, or IDENT)
// POST: curToken is RPAREN
// Task 2.7.3: Renamed to enable dual-mode operation
func (p *Parser) parseParameterListAtTokenTraditional() []*ast.Parameter {
	params := []*ast.Parameter{}

	// Parse first parameter group (we're already at first token)
	groupParams := p.parseParameterGroup()
	if groupParams == nil {
		return nil
	}
	params = append(params, groupParams...)

	// Parse remaining parameter groups separated by semicolons
	for p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken() // move to ';'
		p.nextToken() // move past ';'

		groupParams = p.parseParameterGroup()
		if groupParams == nil {
			return nil
		}
		params = append(params, groupParams...)
	}

	// Expect closing parenthesis
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return params
}

// parseParameterListAtTokenCursor parses a full parameter list with names (cursor mode).
// PRE: cursor is at first parameter token (VAR, CONST, LAZY, or IDENT)
// POST: cursor is at RPAREN
//
// Task 2.7.3: New cursor-based implementation for immutable parsing.
func (p *Parser) parseParameterListAtTokenCursor() []*ast.Parameter {
	params := []*ast.Parameter{}

	// Parse first parameter group (we're already at first token)
	groupParams := p.parseParameterGroup()
	if groupParams == nil {
		return nil
	}
	params = append(params, groupParams...)

	// Update cursor after parameter group parsing
	cursor := p.cursor

	// Parse remaining parameter groups separated by semicolons
	for cursor.Peek(1).Type == lexer.SEMICOLON {
		cursor = cursor.Advance() // move to ';'
		cursor = cursor.Advance() // move past ';'
		p.cursor = cursor

		groupParams = p.parseParameterGroup()
		if groupParams == nil {
			return nil
		}
		params = append(params, groupParams...)

		// Update cursor after parameter group parsing
		cursor = p.cursor
	}

	// Expect closing parenthesis
	if cursor.Peek(1).Type != lexer.RPAREN {
		p.addError("expected ')' after parameter list", ErrMissingRParen)
		return nil
	}
	cursor = cursor.Advance() // move to RPAREN
	p.cursor = cursor

	return params
}

// parseTypeOnlyParameterListAtToken is a dispatcher that routes to the appropriate implementation
// based on the parser mode (traditional vs cursor).
//
// Task 2.7.3: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseTypeOnlyParameterListAtToken() []*ast.Parameter {
	if p.useCursor {
		return p.parseTypeOnlyParameterListAtTokenCursor()
	}
	return p.parseTypeOnlyParameterListAtTokenTraditional()
}

// parseTypeOnlyParameterListAtTokenTraditional parses a parameter list with only types (no names).
// Used for shorthand function pointer syntax: function(Integer, String): Boolean
//
// Syntax:
//   - function(Integer): Boolean                  - single param
//   - function(Integer, String): Boolean          - comma-separated
//   - function(Integer; String; Boolean): Float   - semicolon-separated
//
// The parser is positioned at the first type token when this is called.
// Parameters will have nil Name fields.
//
// This format is used in type declarations but not in actual function definitions.
// Example: type TFunc = function(Integer, String): Boolean;
// PRE: curToken is first type token or modifier (CONST, VAR, LAZY)
// POST: curToken is RPAREN
// Task 2.7.3: Renamed to enable dual-mode operation
func (p *Parser) parseTypeOnlyParameterListAtTokenTraditional() []*ast.Parameter {
	params := []*ast.Parameter{}

	// Current token is first type
	for {
		// Check for modifiers (const, var, lazy)
		isConst := false
		isLazy := false
		byRef := false

		if p.curTokenIs(lexer.CONST) {
			isConst = true
			p.nextToken()
		}
		if p.curTokenIs(lexer.LAZY) {
			isLazy = true
			p.nextToken()
		}
		if p.curTokenIs(lexer.VAR) {
			byRef = true
			p.nextToken()
		}

		// Parse type expression (could be complex like "array of Integer" or "function(Integer): Integer")
		typeExpr := p.parseTypeExpression()
		if typeExpr == nil {
			p.addError("expected type in function pointer parameter list", ErrExpectedType)
			return nil
		}

		// Convert TypeExpression to TypeAnnotation
		var typeAnnotation *ast.TypeAnnotation
		switch te := typeExpr.(type) {
		case *ast.TypeAnnotation:
			typeAnnotation = te
		case *ast.FunctionPointerTypeNode:
			// For nested function pointers, use the string representation as type name
			typeAnnotation = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(),
			}
		case *ast.ArrayTypeNode:
			// For array types, use string representation
			token := te.Token
			if token.Type == 0 || token.Literal == "" {
				token = lexer.Token{Type: lexer.ARRAY, Literal: "array", Pos: lexer.Position{}}
			}
			typeAnnotation = &ast.TypeAnnotation{
				Token: token,
				Name:  te.String(),
			}
		case *ast.SetTypeNode:
			// For set types, use string representation
			typeAnnotation = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(),
			}
		default:
			p.addError("unsupported type expression in function pointer parameter", ErrInvalidType)
			return nil
		}

		// Create parameter with nil name (shorthand syntax)
		param := &ast.Parameter{
			Token:   typeAnnotation.Token,
			Name:    nil, // Shorthand syntax has no parameter names
			Type:    typeAnnotation,
			IsLazy:  isLazy,
			ByRef:   byRef,
			IsConst: isConst,
		}
		params = append(params, param)

		// Check what comes next
		if p.peekTokenIs(lexer.COMMA) {
			// More parameters in same group
			p.nextToken() // move to comma
			p.nextToken() // move past comma to next type
			continue
		} else if p.peekTokenIs(lexer.SEMICOLON) {
			// Next parameter group
			p.nextToken() // move to semicolon
			p.nextToken() // move past semicolon to next type
			continue
		} else if p.peekTokenIs(lexer.RPAREN) {
			// End of parameter list
			p.nextToken() // move to RPAREN
			break
		} else {
			p.addError("expected ',', ';', or ')' in function pointer parameter list", ErrUnexpectedToken)
			return nil
		}
	}

	return params
}

// parseTypeOnlyParameterListAtTokenCursor parses a parameter list with only types (cursor mode).
// PRE: cursor is at first type token or modifier (CONST, VAR, LAZY)
// POST: cursor is at RPAREN
//
// Task 2.7.3: New cursor-based implementation for immutable parsing.
func (p *Parser) parseTypeOnlyParameterListAtTokenCursor() []*ast.Parameter {
	params := []*ast.Parameter{}
	cursor := p.cursor

	// Current token is first type
	for {
		// Check for modifiers (const, var, lazy)
		isConst := false
		isLazy := false
		byRef := false

		if cursor.Current().Type == lexer.CONST {
			isConst = true
			cursor = cursor.Advance()
			p.cursor = cursor
		}
		if cursor.Current().Type == lexer.LAZY {
			isLazy = true
			cursor = cursor.Advance()
			p.cursor = cursor
		}
		if cursor.Current().Type == lexer.VAR {
			byRef = true
			cursor = cursor.Advance()
			p.cursor = cursor
		}

		// Parse type expression (could be complex like "array of Integer" or "function(Integer): Integer")
		typeExpr := p.parseTypeExpression()
		if typeExpr == nil {
			p.addError("expected type in function pointer parameter list", ErrExpectedType)
			return nil
		}

		// Update cursor after type expression parsing
		cursor = p.cursor

		// Convert TypeExpression to TypeAnnotation
		var typeAnnotation *ast.TypeAnnotation
		switch te := typeExpr.(type) {
		case *ast.TypeAnnotation:
			typeAnnotation = te
		case *ast.FunctionPointerTypeNode:
			// For nested function pointers, use the string representation as type name
			typeAnnotation = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(),
			}
		case *ast.ArrayTypeNode:
			// For array types, use string representation
			token := te.Token
			if token.Type == 0 || token.Literal == "" {
				token = lexer.Token{Type: lexer.ARRAY, Literal: "array", Pos: lexer.Position{}}
			}
			typeAnnotation = &ast.TypeAnnotation{
				Token: token,
				Name:  te.String(),
			}
		case *ast.SetTypeNode:
			// For set types, use string representation
			typeAnnotation = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(),
			}
		default:
			p.addError("unsupported type expression in function pointer parameter", ErrInvalidType)
			return nil
		}

		// Create parameter with nil name (shorthand syntax)
		param := &ast.Parameter{
			Token:   typeAnnotation.Token,
			Name:    nil, // Shorthand syntax has no parameter names
			Type:    typeAnnotation,
			IsLazy:  isLazy,
			ByRef:   byRef,
			IsConst: isConst,
		}
		params = append(params, param)

		// Check what comes next
		if cursor.Peek(1).Type == lexer.COMMA {
			// More parameters in same group
			cursor = cursor.Advance() // move to comma
			cursor = cursor.Advance() // move past comma to next type
			p.cursor = cursor
			continue
		} else if cursor.Peek(1).Type == lexer.SEMICOLON {
			// Next parameter group
			cursor = cursor.Advance() // move to semicolon
			cursor = cursor.Advance() // move past semicolon to next type
			p.cursor = cursor
			continue
		} else if cursor.Peek(1).Type == lexer.RPAREN {
			// End of parameter list
			cursor = cursor.Advance() // move to RPAREN
			p.cursor = cursor
			break
		} else {
			p.addError("expected ',', ';', or ')' in function pointer parameter list", ErrUnexpectedToken)
			return nil
		}
	}

	return params
}
