package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
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
// PRE: cursor is TYPE
// POST: cursor is SEMICOLON of last type declaration

// PRE: cursor is TYPE
// POST: cursor is SEMICOLON of last type declaration
func (p *Parser) parseTypeDeclaration() ast.Statement {
	typeToken := p.cursor.Current() // Save the TYPE token
	statements := []ast.Statement{}

	// Parse first type declaration
	firstStmt := p.parseSingleTypeDeclaration(typeToken)
	if firstStmt == nil {
		return nil
	}
	statements = append(statements, firstStmt)

	// Continue parsing additional type declarations without the 'type' keyword
	// As long as the next line looks like a type declaration
	for p.looksLikeTypeDeclaration() {
		p.cursor = p.cursor.Advance() // move to identifier
		typeStmt := p.parseSingleTypeDeclaration(typeToken)
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
// This method uses lexer.Peek() to look ahead without modifying parser state.
// PRE: cursor is SEMICOLON (after previous type decl)
// POST: cursor is SEMICOLON (after previous type decl)

// looksLikeTypeDeclaration checks if the current position looks like the start of
// Pattern: IDENT EQ (CLASS|INTERFACE|LPAREN|RECORD|SET|ARRAY|ENUM|FUNCTION|PROCEDURE|HELPER|...)
// PRE: cursor is on SEMICOLON (after previous type decl)
// POST: cursor is on SEMICOLON (after previous type decl)
func (p *Parser) looksLikeTypeDeclaration() bool {
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
// PRE: cursor is TYPE or type name IDENT
// POST: cursor is SEMICOLON

// PRE: cursor is at TYPE or type name IDENT
// POST: cursor is at SEMICOLON
func (p *Parser) parseSingleTypeDeclaration(typeToken lexer.Token) ast.Statement {
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
	p.cursor = cursor

	// Check what kind of type declaration this is by peeking at the next token
	nextToken := cursor.Peek(1)

	// Subrange type: type TDigit = 0..9; or type TTemperature = -40..50;
	// Check if it starts with a number or minus (for negative ranges)
	if nextToken.Type == lexer.INT || nextToken.Type == lexer.MINUS {
		cursor = cursor.Advance() // move past '='
		p.cursor = cursor

		// Parse the low bound expression
		lowBound := p.parseExpression(LOWEST)
		if lowBound == nil {
			p.addError("expected expression for subrange low bound", ErrUnexpectedToken)
			return nil
		}

		// Check for '..' operator
		if p.cursor.Peek(1).Type != lexer.DOTDOT {
			p.addError("expected '..' in subrange type", ErrUnexpectedToken)
			return nil
		}
		p.cursor = p.cursor.Advance() // move to DOTDOT
		p.cursor = p.cursor.Advance() // move past DOTDOT

		// Parse the high bound expression
		highBound := p.parseExpression(LOWEST)
		if highBound == nil {
			p.addError("expected expression for subrange high bound", ErrUnexpectedToken)
			return nil
		}

		// Expect semicolon
		if p.cursor.Peek(1).Type != lexer.SEMICOLON {
			p.addError("expected ';' after subrange type declaration", ErrMissingSemicolon)
			return nil
		}
		p.cursor = p.cursor.Advance() // move to SEMICOLON

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

	// Type alias: type TUserID = Integer;
	if nextToken.Type == lexer.IDENT {
		cursor = cursor.Advance() // move past '=' to type name
		p.cursor = cursor
		currentToken := cursor.Current()
		aliasedType := &ast.TypeAnnotation{
			Token: currentToken,
			Name:  currentToken.Literal,
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

	// For other type declarations (class, interface, enum, etc.), handle each type
	switch nextToken.Type {
	case lexer.INTERFACE:
		// Interface declaration: type IMyInterface = interface ... end;
		cursor = cursor.Advance() // move to INTERFACE
		p.cursor = cursor
		return p.parseInterfaceDeclarationBody(nameIdent)
	case lexer.PARTIAL:
		// Partial class declaration: type TMyClass = partial class ... end;
		cursor = cursor.Advance() // move to PARTIAL
		p.cursor = cursor

		// Expect CLASS after PARTIAL
		if cursor.Peek(1).Type != lexer.CLASS {
			p.addError("expected 'class' after 'partial'", ErrUnexpectedToken)
			return nil
		}
		cursor = cursor.Advance() // move to CLASS
		p.cursor = cursor

		classDecl := p.parseClassDeclarationBody(nameIdent)
		if classDecl != nil {
			classDecl.IsPartial = true
		}
		return classDecl
	case lexer.CLASS:
		cursor = cursor.Advance() // move to CLASS
		p.cursor = cursor

		// Check if this is a metaclass type alias: type TBaseClass = class of TBase;
		// or a class declaration: type TMyClass = class ... end;
		if cursor.Peek(1).Type == lexer.OF {
			// Metaclass type alias: type TBaseClass = class of TBase;
			// Parse as type expression (class of ...)
			classOfType := p.parseClassOfType()
			if classOfType == nil {
				return nil
			}

			// Expect semicolon
			if p.cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after class of type", ErrMissingSemicolon)
				return nil
			}
			p.cursor = p.cursor.Advance() // move to SEMICOLON

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
		if cursor.Peek(1).Type == lexer.PARTIAL {
			p.cursor = p.cursor.Advance() // move to PARTIAL
			classDecl := p.parseClassDeclarationBody(nameIdent)
			if classDecl != nil {
				classDecl.IsPartial = true
			}
			return classDecl
		}
		// Regular class declaration: type TMyClass = class ... end;
		return p.parseClassDeclarationBody(nameIdent)
	case lexer.RECORD:
		// Could be either:
		//   - Record declaration: type TPoint = record X, Y: Integer; end;
		//   - Record helper: type THelper = record helper for TypeName ... end;
		cursor = cursor.Advance() // move to RECORD
		p.cursor = cursor
		return p.parseRecordOrHelperDeclaration(nameIdent, typeToken)
	case lexer.SET:
		// Set declaration: type TDays = set of TWeekday;
		cursor = cursor.Advance() // move to SET
		p.cursor = cursor
		return p.parseSetDeclaration(nameIdent, typeToken)
	case lexer.ARRAY:
		// Array declaration: type TMyArray = array[1..10] of Integer;
		cursor = cursor.Advance() // move to ARRAY
		p.cursor = cursor
		return p.parseArrayDeclaration(nameIdent, typeToken)
	case lexer.LPAREN:
		// Enum declaration: type TColor = (Red, Green, Blue);
		// Do NOT advance past '=' - parseEnumDeclaration expects cursor on '=', will peek ahead for LPAREN
		return p.parseEnumDeclaration(nameIdent, typeToken, false, false)
	case lexer.ENUM:
		// Scoped enum: type TEnum = enum (One, Two);
		cursor = cursor.Advance() // move to ENUM
		p.cursor = cursor
		return p.parseEnumDeclaration(nameIdent, typeToken, true, false)
	case lexer.FLAGS:
		// Flags enum: type TFlags = flags (a, b, c);
		cursor = cursor.Advance() // move to FLAGS
		p.cursor = cursor
		return p.parseEnumDeclaration(nameIdent, typeToken, true, true)
	case lexer.FUNCTION, lexer.PROCEDURE:
		// Function pointer: type TFunc = function(x: Integer): Boolean;
		// Procedure pointer: type TProc = procedure(msg: String);
		// Method pointer: type TEvent = procedure(Sender: TObject) of object;
		cursor = cursor.Advance() // move to FUNCTION or PROCEDURE
		p.cursor = cursor
		return p.parseFunctionPointerTypeDeclaration(nameIdent, typeToken)
	case lexer.HELPER:
		// Helper declaration (without "record" keyword):
		// type THelper = helper for TypeName ... end;
		cursor = cursor.Advance() // move to HELPER
		p.cursor = cursor
		return p.parseHelperDeclaration(nameIdent, typeToken, false)
	}

	// Unknown type declaration
	p.addError("expected 'class', 'partial', 'interface', 'enum', 'record', 'set', 'array', 'function', 'procedure', 'helper', or '(' after '=' in type declaration", ErrUnexpectedToken)
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
// PRE: cursor is FUNCTION or PROCEDURE
// POST: cursor is SEMICOLON
func (p *Parser) parseFunctionPointerTypeDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) ast.Statement {
	builder := p.StartNode()

	funcOrProcToken := p.cursor.Current()

	// Current token is FUNCTION or PROCEDURE
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

		endToken = p.cursor.Current()
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

		retTypeTok := p.cursor.Current()

		returnType := &ast.TypeAnnotation{
			Token: retTypeTok,
			Name:  retTypeTok.Literal,
		}
		// EndPos is after the type identifier token
		returnType.EndPos = p.endPosFromToken(retTypeTok)
		funcPtrType.ReturnType = returnType
		endToken = retTypeTok
	}

	// Check for "of object" clause (method pointers)
	if p.peekTokenIs(lexer.OF) {
		p.nextToken() // move to OF
		if !p.expectPeek(lexer.OBJECT) {
			return nil
		}
		funcPtrType.OfObject = true

		endToken = p.cursor.Current()

		// EndPos is after "object" token
		funcPtrType.EndPos = p.endPosFromToken(endToken)
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

// Called after 'type Name = interface' has already been parsed.
//

// Called after 'type Name = interface' has already been parsed.
// Current token should be 'interface'.
// PRE: cursor is INTERFACE
// POST: cursor is SEMICOLON
func (p *Parser) parseInterfaceDeclarationBody(nameIdent *ast.Identifier) *ast.InterfaceDecl {
	builder := p.StartNode()
	cursor := p.cursor

	interfaceDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: cursor.Current(), // 'interface' token
		},
		Name: nameIdent,
		// Initialize slices to avoid nil checks downstream
		Properties: []*ast.PropertyDecl{},
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

		switch cursor.Current().Type {
		case lexer.PROCEDURE, lexer.FUNCTION:
			// Parse method declaration (procedure or function)
			method := p.parseInterfaceMethodDecl()
			if method != nil {
				interfaceDecl.Methods = append(interfaceDecl.Methods, method)
			}
			cursor = p.cursor
		case lexer.PROPERTY:
			// Parse property declaration
			property := p.parsePropertyDeclaration()
			if property != nil {
				interfaceDecl.Properties = append(interfaceDecl.Properties, property)
			}
			cursor = p.cursor
		default:
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

// Syntax: procedure MethodName(params); or function MethodName(params): ReturnType;
//

// Syntax: procedure MethodName(params);
//
//	function MethodName(params): ReturnType;
//
// PRE: cursor is PROCEDURE or FUNCTION
// POST: cursor is SEMICOLON
func (p *Parser) parseInterfaceMethodDecl() *ast.InterfaceMethodDecl {
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
