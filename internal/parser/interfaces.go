package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseTypeDeclaration determines whether this is a class, interface, or enum declaration
// and dispatches to the appropriate parser.
// Syntax: type Name = class... OR type Name = interface... OR type Name = (...)
//
// Task 7.85: Dispatcher for type declarations
// Task 8.38: Added enum support
func (p *Parser) parseTypeDeclaration() ast.Statement {
	// Current token is TYPE
	// Pattern: TYPE IDENT EQ (CLASS|INTERFACE|LPAREN|ENUM)

	typeToken := p.curToken // Save the TYPE token

	// Advance and peek to see what type declaration this is
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	nameIdent := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.EQ) {
		return nil
	}

	// Now peek to see what kind of type declaration this is
	// Task 9.96: Check for subrange or type alias
	// Subrange: type TDigit = 0..9;
	// Type alias: type TUserID = Integer;

	// Check if this could be a subrange (expression followed by ..)
	// Expressions can start with: INT, MINUS, LPAREN, IDENT, etc.
	if p.peekTokenIs(lexer.INT) || p.peekTokenIs(lexer.MINUS) || p.peekTokenIs(lexer.FLOAT) {
		// Might be subrange - parse first expression
		p.nextToken() // move to expression start
		lowBound := p.parseExpression(LOWEST)

		// Check if followed by DOTDOT
		if p.peekTokenIs(lexer.DOTDOT) {
			// It's a subrange!
			p.nextToken() // move to DOTDOT

			// Parse high bound
			p.nextToken() // move past DOTDOT to high bound expression
			highBound := p.parseExpression(LOWEST)

			// Expect semicolon
			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}

			return &ast.TypeDeclaration{
				Token:      typeToken,
				Name:       nameIdent,
				IsSubrange: true,
				LowBound:   lowBound,
				HighBound:  highBound,
			}
		}
		// Not a subrange, fall through to error
		p.addError("unexpected expression after '=' in type declaration (expected type name or subrange)")
		return nil
	} else if p.peekTokenIs(lexer.IDENT) {
		// Type alias: type TUserID = Integer;
		p.nextToken() // move to aliased type identifier
		aliasedType := &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}

		// Expect semicolon
		if !p.expectPeek(lexer.SEMICOLON) {
			return nil
		}

		return &ast.TypeDeclaration{
			Token:       typeToken,
			Name:        nameIdent,
			IsAlias:     true,
			AliasedType: aliasedType,
		}
	} else if p.peekTokenIs(lexer.INTERFACE) {
		p.nextToken() // move to INTERFACE
		return p.parseInterfaceDeclarationBody(nameIdent)
	} else if p.peekTokenIs(lexer.CLASS) {
		p.nextToken() // move to CLASS
		return p.parseClassDeclarationBody(nameIdent)
	} else if p.peekTokenIs(lexer.RECORD) {
		// Record declaration: type TPoint = record X, Y: Integer; end;
		// Task 8.62: Integrated record parsing
		return p.parseRecordDeclaration(nameIdent, typeToken)
	} else if p.peekTokenIs(lexer.SET) {
		// Set declaration: type TDays = set of TWeekday;
		// Task 8.91-8.92: Integrated set parsing
		p.nextToken() // move to SET
		return p.parseSetDeclaration(nameIdent, typeToken)
	} else if p.peekTokenIs(lexer.ARRAY) {
		// Array declaration: type TMyArray = array[1..10] of Integer;
		// Task 8.122: Integrated array parsing
		p.nextToken() // move to ARRAY
		return p.parseArrayDeclaration(nameIdent, typeToken)
	} else if p.peekTokenIs(lexer.LPAREN) {
		// Enum declaration: type TColor = (Red, Green, Blue);
		return p.parseEnumDeclaration(nameIdent, typeToken)
	} else if p.peekTokenIs(lexer.ENUM) {
		// Scoped enum: type TEnum = enum (One, Two);
		p.nextToken() // move to ENUM
		return p.parseEnumDeclaration(nameIdent, typeToken)
	} else if p.peekTokenIs(lexer.FUNCTION) || p.peekTokenIs(lexer.PROCEDURE) {
		// Function pointer: type TFunc = function(x: Integer): Boolean;
		// Procedure pointer: type TProc = procedure(msg: String);
		// Method pointer: type TEvent = procedure(Sender: TObject) of object;
		// Task 9.155: Function pointer type declaration support
		p.nextToken() // move to FUNCTION or PROCEDURE
		return p.parseFunctionPointerTypeDeclaration(nameIdent, typeToken)
	}

	// Unknown type declaration
	p.addError("expected 'class', 'interface', 'enum', 'record', 'set', 'array', 'function', 'procedure', or '(' after '=' in type declaration")
	return nil
}

// parseFunctionPointerTypeDeclaration parses a function or procedure pointer type declaration.
// Called after 'type Name = function' or 'type Name = procedure' has been parsed.
// Current token should be FUNCTION or PROCEDURE.
//
// Examples:
//   - type TFunc = function(x: Integer): Boolean;
//   - type TProc = procedure(msg: String);
//   - type TCallback = procedure;
//   - type TEvent = procedure(Sender: TObject) of object;
//
// Task 9.155: Function pointer type declaration support
func (p *Parser) parseFunctionPointerTypeDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) ast.Statement {
	// Current token is FUNCTION or PROCEDURE
	funcOrProcToken := p.curToken
	isFunction := funcOrProcToken.Type == lexer.FUNCTION

	// Create the function pointer type node
	funcPtrType := &ast.FunctionPointerTypeNode{
		Token:      funcOrProcToken,
		Parameters: []*ast.Parameter{},
		OfObject:   false,
	}

	// Expect opening parenthesis for parameter list
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	// Check if there are parameters (not just empty parens)
	if !p.peekTokenIs(lexer.RPAREN) {
		// Parse parameter list using existing function
		funcPtrType.Parameters = p.parseParameterList()
	} else {
		// Empty parameter list
		p.nextToken() // move to RPAREN
	}

	// Expect closing parenthesis
	if !p.curTokenIs(lexer.RPAREN) {
		p.addError("expected ')' after parameter list in function pointer type")
		return nil
	}

	// Parse return type for functions (not procedures)
	if isFunction {
		// Expect colon and return type
		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		// Parse return type
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		funcPtrType.ReturnType = &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}
	}

	// Check for "of object" clause (method pointers)
	if p.peekTokenIs(lexer.OF) {
		p.nextToken() // move to OF
		if !p.expectPeek(lexer.OBJECT) {
			return nil
		}
		funcPtrType.OfObject = true
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Return the complete type declaration with function pointer type
	return &ast.TypeDeclaration{
		Token:               typeToken,
		Name:                nameIdent,
		FunctionPointerType: funcPtrType,
		IsFunctionPointer:   true,
	}
}

// parseInterfaceDeclarationBody parses the body of an interface declaration.
// Called after 'type Name = interface' has already been parsed.
// Current token should be 'interface'.
//
// Task 7.81: Parse interface declarations with inheritance and external support
func (p *Parser) parseInterfaceDeclarationBody(nameIdent *ast.Identifier) *ast.InterfaceDecl {
	interfaceDecl := &ast.InterfaceDecl{
		Token: p.curToken, // 'interface' token
		Name:  nameIdent,
	}

	// Check for optional parent interface (IDerived = interface(IBase))
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // move to '('
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		interfaceDecl.Parent = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	}

	// Check for 'external' keyword
	if p.peekTokenIs(lexer.EXTERNAL) {
		p.nextToken() // move to 'external'
		interfaceDecl.IsExternal = true

		// Check for optional external name string
		if p.peekTokenIs(lexer.STRING) {
			p.nextToken() // move to string
			interfaceDecl.ExternalName = p.curToken.Literal
		}
	}

	// Check for forward declaration: type IForward = interface;
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken() // move to semicolon
		return interfaceDecl
	}

	// Parse interface body (method declarations) until 'end'
	p.nextToken() // move past 'interface' or ')' or external name

	interfaceDecl.Methods = []*ast.InterfaceMethodDecl{}

	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
		// Skip semicolons
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// Parse method declaration (procedure or function)
		if p.curTokenIs(lexer.PROCEDURE) || p.curTokenIs(lexer.FUNCTION) {
			method := p.parseInterfaceMethodDecl()
			if method != nil {
				interfaceDecl.Methods = append(interfaceDecl.Methods, method)
			}
		} else {
			// Unknown token in interface body, skip it
			p.nextToken()
			continue
		}

		p.nextToken()
	}

	// Expect 'end'
	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close interface declaration")
		return nil
	}

	// Expect terminating semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return interfaceDecl
}

// parseInterfaceMethodDecl parses a method declaration within an interface.
// Syntax: procedure MethodName(params);
//
//	function MethodName(params): ReturnType;
//
// Task 7.82: Parse interface method declarations (abstract methods with no body)
func (p *Parser) parseInterfaceMethodDecl() *ast.InterfaceMethodDecl {
	methodDecl := &ast.InterfaceMethodDecl{Token: p.curToken}

	// Determine if this is a procedure or function
	isProcedure := p.curTokenIs(lexer.PROCEDURE)

	// Expect method name identifier
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	methodDecl.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Parse parameter list if present
	methodDecl.Parameters = []*ast.Parameter{}
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // move to '('
		methodDecl.Parameters = p.parseParameterList()
	}

	// Parse return type for functions
	if !isProcedure {
		// Expect ':' for return type
		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		// Expect type identifier
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}

		methodDecl.ReturnType = &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}
	}

	// Expect semicolon (interface methods have no body)
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Task 7.82: Error if body is present (interfaces are abstract)
	// If we see 'begin' next, it's an error
	if p.peekTokenIs(lexer.BEGIN) {
		p.addError("interface methods cannot have a body")
		return nil
	}

	return methodDecl
}
