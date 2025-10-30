package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseConstDeclaration parses a constant declaration.
// Syntax: const NAME = VALUE; or const NAME := VALUE; or const NAME: TYPE = VALUE;
func (p *Parser) parseConstDeclaration() ast.Statement {
	stmt := &ast.ConstDecl{Token: p.curToken}

	// Expect identifier (const name)
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check for optional type annotation (: Type)
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // move to ':'
		if !p.expectPeek(lexer.IDENT) {
			p.addError("expected type identifier after ':' in const declaration")
			return stmt
		}
		stmt.Type = &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}
	}

	// Expect '=' or ':=' token
	if !p.peekTokenIs(lexer.EQ) && !p.peekTokenIs(lexer.ASSIGN) {
		p.addError("expected '=' or ':=' after const name")
		return stmt
	}
	p.nextToken() // move to '=' or ':='

	// Parse value expression
	p.nextToken()
	stmt.Value = p.parseExpression(ASSIGN)

	// Expect semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
