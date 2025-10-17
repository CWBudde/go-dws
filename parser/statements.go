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
	default:
		if p.curToken.Type == lexer.IDENT && p.peekTokenIs(lexer.ASSIGN) {
			return p.parseAssignmentStatement()
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

// parseAssignmentStatement parses an assignment statement.
func (p *Parser) parseAssignmentStatement() ast.Statement {
	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	stmt := &ast.AssignmentStatement{
		Token: p.curToken,
		Name:  name,
	}

	p.nextToken()
	stmt.Value = p.parseExpression(ASSIGN)

	// Optional semicolon (some contexts like repeat-until don't require it)
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
