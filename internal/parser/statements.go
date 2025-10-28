package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseStatement parses a single statement.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.BEGIN:
		return p.parseBlockStatement()
	case lexer.VAR:
		return p.parseVarDeclaration()
	case lexer.CONST:
		return p.parseConstDeclaration()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.WHILE:
		return p.parseWhileStatement()
	case lexer.REPEAT:
		return p.parseRepeatStatement()
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.CASE:
		return p.parseCaseStatement()
	case lexer.BREAK:
		return p.parseBreakStatement()
	case lexer.CONTINUE:
		return p.parseContinueStatement()
	case lexer.EXIT:
		return p.parseExitStatement()
	case lexer.TRY:
		return p.parseTryStatement()
	case lexer.RAISE:
		return p.parseRaiseStatement()
	case lexer.FUNCTION, lexer.PROCEDURE:
		return p.parseFunctionDeclaration()
	case lexer.OPERATOR:
		return p.parseOperatorDeclaration()
	case lexer.CLASS:
		if p.peekTokenIs(lexer.FUNCTION) || p.peekTokenIs(lexer.PROCEDURE) {
			p.nextToken() // move to function/procedure token
			fn := p.parseFunctionDeclaration()
			if fn != nil {
				fn.IsClassMethod = true
			}
			return fn
		}
		p.addError("expected 'function' or 'procedure' after 'class'")
		return nil
	case lexer.CONSTRUCTOR:
		// Parse constructor implementation outside class body
		method := p.parseFunctionDeclaration()
		if method != nil {
			method.IsConstructor = true
		}
		return method
	case lexer.DESTRUCTOR:
		// Parse destructor implementation outside class body
		method := p.parseFunctionDeclaration()
		if method != nil {
			method.IsDestructor = true
		}
		return method
	case lexer.TYPE:
		// Task 7.85: Dispatch to class or interface parser
		// Both parsers will handle the full parsing starting from TYPE token
		return p.parseTypeDeclaration()
	case lexer.USES:
		// Parse uses clause at program level
		return p.parseUsesClause()
	default:
		// Check for assignment (simple or member assignment)
		if p.curToken.Type == lexer.IDENT {
			// Could be:
			// 1. x := value (assignment)
			// 2. obj.field := value (member assignment)
			// 3. x: Type; (var declaration without 'var' keyword - part of var section)

			// Check if this is a var declaration (IDENT COLON pattern)
			if p.peekTokenIs(lexer.COLON) {
				// This is a var declaration in a var section
				// Treat it like "var x: Type;"
				return p.parseVarDeclaration()
			}

			// Otherwise, parse as assignment or expression
			return p.parseAssignmentOrExpression()
		}
		return p.parseExpressionStatement()
	}
}

// parseBlockStatement parses a begin...end block.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken() // advance past 'begin'

	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		p.nextToken()

		// Skip any semicolons after the statement
		for p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}
	}

	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close block")
		for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}
		return block
	}

	return block
}

// parseExpressionStatement parses an expression statement.
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	// Optional semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseVarDeclaration parses a variable declaration statement.
// Can be called in two contexts:
//  1. After 'var' keyword: var x: Integer;
//  2. In a var section without 'var': x: Integer; (curToken is already the IDENT)
//
// Task 7.143: Now supports external variables:
//
//	var x: Integer; external;
//	var y: String; external 'externalName';
func (p *Parser) parseVarDeclaration() ast.Statement {
	stmt := &ast.VarDeclStatement{Token: p.curToken}

	// Check if we're already at the identifier (var section continuation)
	// or if we need to advance to it (after 'var' keyword)
	if p.curTokenIs(lexer.VAR) {
		// After 'var' keyword, expect identifier next
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
	} else if !p.curTokenIs(lexer.IDENT) {
		// Should already be at an identifier
		p.addError("expected identifier in var declaration")
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // move to ':'
		if !p.expectPeek(lexer.IDENT) {
			p.addError("expected type identifier after ':' in var declaration")
			return stmt
		}
		stmt.Type = &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}
	}

	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // move to ':='
		p.nextToken()
		stmt.Value = p.parseExpression(ASSIGN)
	}

	// Task 7.143: Check for 'external' keyword
	if p.peekTokenIs(lexer.EXTERNAL) {
		p.nextToken() // move to 'external'
		stmt.IsExternal = true

		// Check for optional external name: external 'customName'
		if p.peekTokenIs(lexer.STRING) {
			p.nextToken() // move to string literal
			stmt.ExternalName = p.curToken.Literal
		}
	}

	if !p.expectPeek(lexer.SEMICOLON) {
		return stmt
	}

	return stmt
}

// parseAssignmentOrExpression determines if we have an assignment or expression statement.
// This handles both simple assignments (x := value) and member assignments (obj.field := value).
func (p *Parser) parseAssignmentOrExpression() ast.Statement {
	// Save starting position
	startToken := p.curToken

	// Parse the left side as an expression (could be identifier or member access)
	left := p.parseExpression(LOWEST)

	// Check if next token is assignment
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // move to :=

		// Determine what kind of assignment this is
		switch leftExpr := left.(type) {
		case *ast.Identifier:
			// Simple assignment: x := value
			stmt := &ast.AssignmentStatement{
				Token:  p.curToken,
				Target: leftExpr,
			}
			p.nextToken()
			stmt.Value = p.parseExpression(ASSIGN)

			// Optional semicolon
			if p.peekTokenIs(lexer.SEMICOLON) {
				p.nextToken()
			}
			return stmt

		case *ast.MemberAccessExpression:
			// Member assignment: obj.field := value
			stmt := &ast.AssignmentStatement{
				Token:  p.curToken,
				Target: leftExpr,
			}

			p.nextToken()
			stmt.Value = p.parseExpression(ASSIGN)

			// Optional semicolon
			if p.peekTokenIs(lexer.SEMICOLON) {
				p.nextToken()
			}
			return stmt

		case *ast.IndexExpression:
			// Array index assignment: arr[i] := value
			stmt := &ast.AssignmentStatement{
				Token:  p.curToken,
				Target: leftExpr,
			}

			p.nextToken()
			stmt.Value = p.parseExpression(ASSIGN)

			// Optional semicolon
			if p.peekTokenIs(lexer.SEMICOLON) {
				p.nextToken()
			}
			return stmt

		default:
			p.addError("invalid assignment target")
			return nil
		}
	}

	// Not an assignment, treat as expression statement
	stmt := &ast.ExpressionStatement{
		Token:      startToken,
		Expression: left,
	}

	// Optional semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
