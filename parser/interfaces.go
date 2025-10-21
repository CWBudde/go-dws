package parser

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
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
	if p.peekTokenIs(lexer.INTERFACE) {
		p.nextToken() // move to INTERFACE
		return p.parseInterfaceDeclarationBody(nameIdent)
	} else if p.peekTokenIs(lexer.CLASS) {
		p.nextToken() // move to CLASS
		return p.parseClassDeclarationBody(nameIdent)
	} else if p.peekTokenIs(lexer.LPAREN) {
		// Enum declaration: type TColor = (Red, Green, Blue);
		return p.parseEnumDeclaration(nameIdent, typeToken)
	} else if p.peekTokenIs(lexer.ENUM) {
		// Scoped enum: type TEnum = enum (One, Two);
		p.nextToken() // move to ENUM
		return p.parseEnumDeclaration(nameIdent, typeToken)
	}

	// Unknown type declaration
	p.addError("expected 'class', 'interface', 'enum', or '(' after '=' in type declaration")
	return nil
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

	// Check for 'external' keyword (Task 7.81)
	if p.peekTokenIs(lexer.EXTERNAL) {
		p.nextToken() // move to 'external'
		interfaceDecl.IsExternal = true

		// Check for optional external name string
		if p.peekTokenIs(lexer.STRING) {
			p.nextToken() // move to string
			interfaceDecl.ExternalName = p.curToken.Literal
		}
	}

	// Check for forward declaration: type IForward = interface; (Task 7.84)
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
