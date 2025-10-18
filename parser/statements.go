package parser

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
)

// parseStatement parses a single statement.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.BEGIN:
		return p.parseBlockStatement()
	case lexer.VAR:
		return p.parseVarDeclaration()
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
	case lexer.FUNCTION, lexer.PROCEDURE:
		return p.parseFunctionDeclaration()
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
		return p.parseClassDeclaration()
	default:
		// Check for assignment (simple or member assignment)
		if p.curToken.Type == lexer.IDENT {
			// Could be: x := value OR obj.field := value
			// We need to parse the left side first to determine which it is
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
func (p *Parser) parseVarDeclaration() ast.Statement {
	stmt := &ast.VarDeclStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
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
				Token: p.curToken,
				Name:  leftExpr,
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
			// For now, we'll store the member access in Name field
			// The interpreter will need to handle this specially
			stmt := &ast.AssignmentStatement{
				Token: p.curToken,
				Name:  nil, // We'll use a special marker
			}

			// Store the object and member for the interpreter
			// We need to extract the object identifier and member name
			objIdent, ok := leftExpr.Object.(*ast.Identifier)
			if !ok {
				p.addError("member assignment requires identifier.member pattern")
				return nil
			}

			// Create a synthetic identifier that the interpreter can recognize
			// Format: "object.member" - interpreter will parse this
			stmt.Name = &ast.Identifier{
				Token: objIdent.Token,
				Value: objIdent.Value + "." + leftExpr.Member.Value,
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
