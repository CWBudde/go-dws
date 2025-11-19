package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseTypeDeclaration parses one or more type declarations in a type section.
// In DWScript, a single 'type' keyword can introduce multiple type declarations:
//
//	type
//	  TFirst = class ... end;
//	  TSecond = class ... end;
//	  TThird = Integer;
//
// This function handles multiple declarations and returns either a single statement
// or a BlockStatement containing multiple type declarations.
// PRE: curToken is TYPE
// POST: curToken is SEMICOLON of last type declaration
// Dispatcher: delegates to cursor or traditional mode
func (p *Parser) parseTypeDeclaration() ast.Statement {
	return p.parseTypeDeclarationCursor()
}

// parseTypeDeclarationTraditional parses type declarations using traditional mode.
// PRE: curToken is TYPE
// POST: curToken is SEMICOLON of last type declaration
func (p *Parser) parseTypeDeclarationCursor() ast.Statement {
	typeToken := p.cursor.Current() // Save the TYPE token
	statements := []ast.Statement{}

	// Parse first type declaration
	firstStmt := p.parseSingleTypeDeclarationCursor(typeToken)
	if firstStmt == nil {
		return nil
	}
	statements = append(statements, firstStmt)

	// Continue parsing additional type declarations without the 'type' keyword
	// As long as the next line looks like a type declaration
	for p.looksLikeTypeDeclarationCursor() {
		p.cursor = p.cursor.Advance() // move to identifier
		typeStmt := p.parseSingleTypeDeclarationCursor(typeToken)
		if typeStmt == nil {
			break
		}
		statements = append(statements, typeStmt)
	}

	// If only one declaration, return it directly
	if len(statements) == 1 {
		return statements[0]
	}

	// Multiple declarations: wrap in a BlockStatement
	return &ast.BlockStatement{
		BaseNode:   ast.BaseNode{Token: typeToken},
		Statements: statements,
	}
}

// looksLikeTypeDeclaration checks if the current position looks like the start of
// a type declaration (without the 'type' keyword).
// Pattern: IDENT EQ (CLASS|INTERFACE|LPAREN|RECORD|SET|ARRAY|ENUM|FUNCTION|PROCEDURE|HELPER|...)
//
// This method uses lexer.Peek() to look ahead without modifying parser state (Task 12.3.4).
// PRE: curToken is SEMICOLON (after previous type decl)
// POST: curToken is SEMICOLON (after previous type decl)
func (p *Parser) looksLikeTypeDeclaration() bool {
	// After a type declaration, we're typically at a semicolon
	// The next token should be an identifier (type name)
	if !p.peekTokenIs(lexer.IDENT) {
		return false
	}

	// Look ahead past peekToken to see if the next token is '='
	// Since parser is already 1 token ahead (peekToken = IDENT), peek(0) gives us
	// the token after peekToken, which should be '='
	tok := p.peek(0)
	return tok.Type == lexer.EQ
}

// looksLikeTypeDeclarationCursor checks if the current position looks like the start of
// a type declaration using cursor mode.
// Pattern: IDENT EQ (CLASS|INTERFACE|LPAREN|RECORD|SET|ARRAY|ENUM|FUNCTION|PROCEDURE|HELPER|...)
// PRE: cursor is on SEMICOLON (after previous type decl)
// POST: cursor is on SEMICOLON (after previous type decl)
func (p *Parser) looksLikeTypeDeclarationCursor() bool {
	// After a type declaration, we're typically at a semicolon
	// The next token should be an identifier (type name)
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.IDENT {
		return false
	}

	// Look ahead to see if the token after the identifier is '='
	tokenAfterIdent := p.cursor.Peek(2)
	return tokenAfterIdent.Type == lexer.EQ
}

// parseSingleTypeDeclaration parses a single type declaration.
// This is the core logic extracted from the original parseTypeDeclaration.
// Assumes we're already positioned at the identifier (or TYPE token).
// PRE: curToken is TYPE or type name IDENT
// POST: curToken is SEMICOLON
func (p *Parser) parseSingleTypeDeclaration(typeToken lexer.Token) ast.Statement {
	builder := p.StartNode()

	// Check if we're already at the identifier (type section continuation)
	// or if we need to advance to it (after 'type' keyword)
	var nameIdent *ast.Identifier

	if p.curTokenIs(lexer.TYPE) {
		// After 'type' keyword, expect identifier next
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		nameIdent = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: p.curToken,
				},
			},
			Value: p.curToken.Literal,
		}
	} else if !p.isIdentifierToken(p.curToken.Type) {
		// Should already be at an identifier
		p.addError("expected identifier in type declaration", ErrExpectedIdent)
		return nil
	} else {
		nameIdent = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: p.curToken,
				},
			},
			Value: p.curToken.Literal,
		}
	}

	// Expect '=' after type name
	if !p.expectPeek(lexer.EQ) {
		return nil
	}

	// Now peek to see what kind of type declaration this is
	//
	// Check for subrange or type alias
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

			typeDecl := &ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: typeToken,
				},
				Name:       nameIdent,
				IsSubrange: true,
				LowBound:   lowBound,
				HighBound:  highBound,
			}
			return builder.Finish(typeDecl).(*ast.TypeDeclaration)
		}
		// Not a subrange, fall through to error
		p.addError("unexpected expression after '=' in type declaration (expected type name or subrange)", ErrUnexpectedToken)
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

		typeDecl := &ast.TypeDeclaration{
			BaseNode: ast.BaseNode{
				Token: typeToken,
			},
			Name:        nameIdent,
			IsAlias:     true,
			AliasedType: aliasedType,
		}
		return builder.Finish(typeDecl).(*ast.TypeDeclaration)
	} else if p.peekTokenIs(lexer.INTERFACE) {
		p.nextToken() // move to INTERFACE
		return p.parseInterfaceDeclarationBody(nameIdent)
	} else if p.peekTokenIs(lexer.PARTIAL) {
		// Partial class: type TMyClass = partial class ... end;
		p.nextToken() // move to PARTIAL
		if !p.expectPeek(lexer.CLASS) {
			p.addError("expected 'class' after 'partial' keyword", ErrUnexpectedToken)
			return nil
		}
		// Parse class body and mark as partial
		classDecl := p.parseClassDeclarationBody(nameIdent)
		if classDecl != nil {
			classDecl.IsPartial = true
		}
		return classDecl
	} else if p.peekTokenIs(lexer.CLASS) {
		p.nextToken() // move to CLASS
		// Check if this is a metaclass type alias: type TBaseClass = class of TBase;
		// or a class declaration: type TMyClass = class ... end;
		if p.peekTokenIs(lexer.OF) {
			// Metaclass type alias: type TBaseClass = class of TBase;
			// Parse as type expression (class of ...)
			classOfType := p.parseClassOfType()
			if classOfType == nil {
				return nil
			}

			// Expect semicolon
			if !p.expectPeek(lexer.SEMICOLON) {
				return nil
			}

			// Create a type declaration with the metaclass type as an inline type
			typeDecl := &ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: typeToken,
				},
				Name:        nameIdent,
				IsAlias:     true,
				AliasedType: classOfType,
			}
			return builder.Finish(typeDecl).(*ast.TypeDeclaration)
		}
		// Check if followed by 'partial': type TMyClass = class partial ... end;
		if p.peekTokenIs(lexer.PARTIAL) {
			p.nextToken() // move to PARTIAL
			classDecl := p.parseClassDeclarationBody(nameIdent)
			if classDecl != nil {
				classDecl.IsPartial = true
			}
			return classDecl
		}
		// Regular class declaration: type TMyClass = class ... end;
		return p.parseClassDeclarationBody(nameIdent)
	} else if p.peekTokenIs(lexer.RECORD) {
		// Could be either:
		//   - Record declaration: type TPoint = record X, Y: Integer; end;
		//   - Record helper: type THelper = record helper for TypeName ... end;
		// Delegate to parseRecordDeclaration which will check for helper
		return p.parseRecordOrHelperDeclaration(nameIdent, typeToken)
	} else if p.peekTokenIs(lexer.SET) {
		// Set declaration: type TDays = set of TWeekday;
		p.nextToken() // move to SET
		return p.parseSetDeclaration(nameIdent, typeToken)
	} else if p.peekTokenIs(lexer.ARRAY) {
		// Array declaration: type TMyArray = array[1..10] of Integer;
		p.nextToken() // move to ARRAY
		return p.parseArrayDeclaration(nameIdent, typeToken)
	} else if p.peekTokenIs(lexer.LPAREN) {
		// Enum declaration: type TColor = (Red, Green, Blue);
		return p.parseEnumDeclaration(nameIdent, typeToken, false, false)
	} else if p.peekTokenIs(lexer.ENUM) {
		// Scoped enum: type TEnum = enum (One, Two);
		p.nextToken() // move to ENUM
		return p.parseEnumDeclaration(nameIdent, typeToken, true, false)
	} else if p.peekTokenIs(lexer.FLAGS) {
		// Flags enum: type TFlags = flags (a, b, c);
		p.nextToken() // move to FLAGS
		return p.parseEnumDeclaration(nameIdent, typeToken, true, true)
	} else if p.peekTokenIs(lexer.FUNCTION) || p.peekTokenIs(lexer.PROCEDURE) {
		// Function pointer: type TFunc = function(x: Integer): Boolean;
		// Procedure pointer: type TProc = procedure(msg: String);
		// Method pointer: type TEvent = procedure(Sender: TObject) of object;
		p.nextToken() // move to FUNCTION or PROCEDURE
		return p.parseFunctionPointerTypeDeclaration(nameIdent, typeToken)
	} else if p.peekTokenIs(lexer.HELPER) {
		// Helper declaration (without "record" keyword):
		// type THelper = helper for TypeName ... end;
		p.nextToken() // move to HELPER
		return p.parseHelperDeclaration(nameIdent, typeToken, false)
	}

	// Unknown type declaration
	p.addError("expected 'class', 'interface', 'enum', 'record', 'set', 'array', 'function', 'procedure', 'helper', or '(' after '=' in type declaration", ErrUnexpectedToken)
	return nil
}

// parseSingleTypeDeclarationCursor parses a single type declaration (cursor mode).
// Task 2.7.4: Cursor-based version for clean migration
// PRE: cursor is at TYPE or type name IDENT
// POST: cursor is at SEMICOLON
func (p *Parser) parseSingleTypeDeclarationCursor(typeToken lexer.Token) ast.Statement {
	builder := p.StartNode()
	cursor := p.cursor

	// Check if we're already at the identifier (type section continuation)
	// or if we need to advance to it (after 'type' keyword)
	var nameIdent *ast.Identifier

	if cursor.Current().Type == lexer.TYPE {
		// After 'type' keyword, expect identifier next
		if !p.isIdentifierToken(cursor.Peek(1).Type) {
			p.addError("expected identifier after 'type'", ErrExpectedIdent)
			return nil
		}
		cursor = cursor.Advance() // move to identifier
		p.cursor = cursor
		nameIdent = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: cursor.Current(),
				},
			},
			Value: cursor.Current().Literal,
		}
	} else if !p.isIdentifierToken(cursor.Current().Type) {
		// Should already be at an identifier
		p.addError("expected identifier in type declaration", ErrExpectedIdent)
		return nil
	} else {
		nameIdent = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: cursor.Current(),
				},
			},
			Value: cursor.Current().Literal,
		}
	}

	// Expect '=' after type name
	if cursor.Peek(1).Type != lexer.EQ {
		p.addError("expected '=' after type name", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to '='
	cursor = cursor.Advance() // move past '=' to type
	p.cursor = cursor

	// Check what kind of type declaration this is
	nextToken := cursor.Current()

	// Type alias: type TUserID = Integer;
	if nextToken.Type == lexer.IDENT {
		aliasedType := &ast.TypeAnnotation{
			Token: nextToken,
			Name:  nextToken.Literal,
		}

		// Expect semicolon
		if cursor.Peek(1).Type != lexer.SEMICOLON {
			p.addError("expected ';' after type declaration", ErrMissingSemicolon)
			return nil
		}
		cursor = cursor.Advance() // move to SEMICOLON
		p.cursor = cursor

		typeDecl := &ast.TypeDeclaration{
			BaseNode: ast.BaseNode{
				Token: typeToken,
			},
			Name:        nameIdent,
			IsAlias:     true,
			AliasedType: aliasedType,
		}
		return builder.Finish(typeDecl).(*ast.TypeDeclaration)
	}

	// For other type declarations (class, interface, enum, etc.), delegate to traditional mode
	// These will be migrated to cursor mode in future tasks
	// TODO: Complete cursor mode migration and remove this fallback

	// IMPORTANT: Restore parser state to the identifier position before delegating.
	// The traditional parser expects curToken to be on the identifier, not the type keyword.
	// We saved the identifier in nameIdent.Token, so restore to that position.
	p.curToken = nameIdent.Token
	// The traditional parser will call nextToken() to get to '=', so set peekToken accordingly
	// We need to manually find the '=' token - it should be 1 position after the identifier
	// Since cursor has already moved past '=', we can search the cursor's token buffer
	for i, tok := range p.cursor.tokens {
		if tok.Type == nameIdent.Token.Type &&
			tok.Pos.Offset == nameIdent.Token.Pos.Offset &&
			i+1 < len(p.cursor.tokens) {
			// Found the identifier, set peekToken to next token (should be '=')
			p.peekToken = p.cursor.tokens[i+1]
			break
		}
	}

	wasUsingCursor := p.useCursor
	p.useCursor = false

	result := p.parseSingleTypeDeclaration(typeToken)

	p.useCursor = wasUsingCursor
	// After traditional parsing, sync the cursor to match the new curToken position
	p.syncTokensToCursor()

	return result
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
// PRE: curToken is FUNCTION or PROCEDURE
// POST: curToken is SEMICOLON
func (p *Parser) parseFunctionPointerTypeDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) ast.Statement {
	builder := p.StartNode()

	// Current token is FUNCTION or PROCEDURE
	funcOrProcToken := p.curToken
	isFunction := funcOrProcToken.Type == lexer.FUNCTION

	// Create the function pointer type node
	funcPtrType := &ast.FunctionPointerTypeNode{
		Token:      funcOrProcToken,
		Parameters: []*ast.Parameter{},
		OfObject:   false,
	}

	// Check if parameter list is present (optional in DWScript)
	// Function pointer types can be:
	//   procedure - no parameters, no parentheses
	//   procedure() - no parameters, with parentheses
	//   procedure(x: Integer) - with parameters
	//   function : Integer - no parameters, no parentheses
	//   function() : Integer - no parameters, with parentheses
	//   function(x: Integer) : Integer - with parameters
	hasParentheses := p.peekTokenIs(lexer.LPAREN)

	// Track the last token for EndPos calculation
	var endToken = funcOrProcToken

	if hasParentheses {
		// Advance to opening parenthesis
		p.nextToken() // move to LPAREN

		// Check if there are parameters (not just empty parens)
		if !p.peekTokenIs(lexer.RPAREN) {
			// Detect syntax type: full (with names) vs shorthand (types only)
			// We need to determine if we have:
			//   Full syntax: "name: Type" or "name1, name2: Type"
			//   Shorthand: "Type" or "Type1, Type2"
			//
			// Strategy: Use simple lookahead WITHOUT advancing parser state.
			// After we detect, advance once and parse accordingly.

			isFullSyntax := p.detectFunctionPointerFullSyntax()

			// Now advance to first parameter/type token
			p.nextToken()

			if isFullSyntax {
				// Full syntax with parameter names
				funcPtrType.Parameters = p.parseParameterListAtToken()
			} else {
				// Shorthand syntax with only types
				funcPtrType.Parameters = p.parseTypeOnlyParameterListAtToken()
			}

			if funcPtrType.Parameters == nil {
				return nil
			}
		} else {
			// Empty parameter list
			p.nextToken() // move to RPAREN
		}

		// Expect closing parenthesis
		if !p.curTokenIs(lexer.RPAREN) {
			p.addError("expected ')' after parameter list in function pointer type", ErrMissingRParen)
			return nil
		}

		// Save RPAREN token for EndPos calculation
		endToken = p.curToken
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
		returnType := &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}
		// EndPos is after the type identifier token
		returnType.EndPos = p.endPosFromToken(p.curToken)
		funcPtrType.ReturnType = returnType
		endToken = p.curToken
	}

	// Check for "of object" clause (method pointers)
	if p.peekTokenIs(lexer.OF) {
		p.nextToken() // move to OF
		if !p.expectPeek(lexer.OBJECT) {
			return nil
		}
		funcPtrType.OfObject = true
		endToken = p.curToken
		// EndPos is after "object" token
		funcPtrType.EndPos = p.endPosFromToken(p.curToken)
	} else {
		// EndPos is after the last significant token (return type, RPAREN, or function/procedure keyword)
		funcPtrType.EndPos = p.endPosFromToken(endToken)
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Return the complete type declaration with function pointer type
	typeDecl := &ast.TypeDeclaration{
		BaseNode: ast.BaseNode{
			Token: typeToken,
		},
		Name:                nameIdent,
		FunctionPointerType: funcPtrType,
		IsFunctionPointer:   true,
	}
	return builder.Finish(typeDecl).(*ast.TypeDeclaration)
}

// parseInterfaceDeclarationBody parses the body of an interface declaration (dual-mode dispatcher).
// Called after 'type Name = interface' has already been parsed.
//
// Task 2.7.2: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseInterfaceDeclarationBody(nameIdent *ast.Identifier) *ast.InterfaceDecl {
	return p.parseInterfaceDeclarationBodyCursor(nameIdent)
}

// parseInterfaceDeclarationBodyTraditional parses the body of an interface declaration (traditional mode).
// Called after 'type Name = interface' has already been parsed.
// Current token should be 'interface'.
// PRE: curToken is INTERFACE
// POST: curToken is SEMICOLON
func (p *Parser) parseInterfaceDeclarationBodyCursor(nameIdent *ast.Identifier) *ast.InterfaceDecl {
	builder := p.StartNode()
	cursor := p.cursor

	interfaceDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: cursor.Current(), // 'interface' token
		},
		Name: nameIdent,
	}

	// Check for optional parent interface (IDerived = interface(IBase))
	if cursor.Peek(1).Type == lexer.LPAREN {
		cursor = cursor.Advance() // move to '('
		p.cursor = cursor

		if cursor.Peek(1).Type != lexer.IDENT {
			p.addError("expected identifier for parent interface", ErrExpectedIdent)
			return nil
		}
		cursor = cursor.Advance() // move to IDENT
		p.cursor = cursor

		interfaceDecl.Parent = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: cursor.Current(),
				},
			},
			Value: cursor.Current().Literal,
		}

		if cursor.Peek(1).Type != lexer.RPAREN {
			p.addError("expected ')' after parent interface", ErrMissingRParen)
			return nil
		}
		cursor = cursor.Advance() // move to ')'
		p.cursor = cursor
	}

	// Check for 'external' keyword
	if cursor.Peek(1).Type == lexer.EXTERNAL {
		cursor = cursor.Advance() // move to 'external'
		p.cursor = cursor
		interfaceDecl.IsExternal = true

		// Check for optional external name string
		if cursor.Peek(1).Type == lexer.STRING {
			cursor = cursor.Advance() // move to string
			p.cursor = cursor
			interfaceDecl.ExternalName = cursor.Current().Literal
		}
	}

	// Check for forward declaration: type IForward = interface;
	if cursor.Peek(1).Type == lexer.SEMICOLON {
		cursor = cursor.Advance() // move to semicolon
		p.cursor = cursor
		return builder.Finish(interfaceDecl).(*ast.InterfaceDecl)
	}

	// Parse interface body (method declarations) until 'end'
	cursor = cursor.Advance() // move past 'interface' or ')' or external name
	p.cursor = cursor

	interfaceDecl.Methods = []*ast.InterfaceMethodDecl{}

	for cursor.Current().Type != lexer.END && cursor.Current().Type != lexer.EOF {
		// Skip semicolons
		if cursor.Current().Type == lexer.SEMICOLON {
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Parse method declaration (procedure or function)
		if cursor.Current().Type == lexer.PROCEDURE || cursor.Current().Type == lexer.FUNCTION {
			method := p.parseInterfaceMethodDecl()
			if method != nil {
				interfaceDecl.Methods = append(interfaceDecl.Methods, method)
			}
			cursor = p.cursor
		} else {
			// Unknown token in interface body, skip it
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		}

		cursor = cursor.Advance()
		p.cursor = cursor
	}

	// Expect 'end'
	if cursor.Current().Type != lexer.END {
		p.addError("expected 'end' to close interface declaration", ErrMissingEnd)
		return nil
	}

	// Expect terminating semicolon
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after 'end'", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	return builder.Finish(interfaceDecl).(*ast.InterfaceDecl)
}

// parseInterfaceMethodDecl parses a method declaration within an interface (dual-mode dispatcher).
// Syntax: procedure MethodName(params); or function MethodName(params): ReturnType;
//
// Task 2.7.2: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseInterfaceMethodDecl() *ast.InterfaceMethodDecl {
	return p.parseInterfaceMethodDeclCursor()
}

// parseInterfaceMethodDeclTraditional parses a method declaration within an interface (traditional mode).
// Syntax: procedure MethodName(params);
//
//	function MethodName(params): ReturnType;
//
// PRE: curToken is PROCEDURE or FUNCTION
// POST: curToken is SEMICOLON
func (p *Parser) parseInterfaceMethodDeclCursor() *ast.InterfaceMethodDecl {
	cursor := p.cursor

	methodDecl := &ast.InterfaceMethodDecl{
		BaseNode: ast.BaseNode{
			Token: cursor.Current(),
		},
	}

	// Determine if this is a procedure or function
	isProcedure := cursor.Current().Type == lexer.PROCEDURE

	// Expect method name identifier
	if cursor.Peek(1).Type != lexer.IDENT {
		p.addError("expected identifier for method name", ErrExpectedIdent)
		return nil
	}
	cursor = cursor.Advance() // move to IDENT
	p.cursor = cursor

	methodDecl.Name = &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: cursor.Current(),
			},
		},
		Value: cursor.Current().Literal,
	}

	// Parse parameter list if present
	methodDecl.Parameters = []*ast.Parameter{}
	if cursor.Peek(1).Type == lexer.LPAREN {
		cursor = cursor.Advance() // move to '('
		p.cursor = cursor
		methodDecl.Parameters = p.parseParameterList()
		cursor = p.cursor
	}

	// Parse return type for functions
	if !isProcedure {
		// Expect ':' for return type
		if cursor.Peek(1).Type != lexer.COLON {
			p.addError("expected ':' for function return type", ErrMissingColon)
			return nil
		}
		cursor = cursor.Advance() // move to ':'
		p.cursor = cursor

		// Expect type identifier
		if cursor.Peek(1).Type != lexer.IDENT {
			p.addError("expected type identifier for return type", ErrExpectedIdent)
			return nil
		}
		cursor = cursor.Advance() // move to type IDENT
		p.cursor = cursor

		methodDecl.ReturnType = &ast.TypeAnnotation{
			Token: cursor.Current(),
			Name:  cursor.Current().Literal,
		}
	}

	// Expect semicolon (interface methods have no body)
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after interface method declaration", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	// Error if body is present (interfaces are abstract)
	// If we see 'begin' next, it's an error
	if cursor.Peek(1).Type == lexer.BEGIN {
		p.addError("interface methods cannot have a body", ErrInvalidSyntax)
		return nil
	}

	return methodDecl
}
