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

		// Parse type expression (can be simple type, function pointer, or array type)
		p.nextToken() // move to type expression
		typeExpr := p.parseTypeExpression()
		if typeExpr == nil {
			p.addError("expected type expression after ':' in const declaration")
			return stmt
		}

		// Convert TypeExpression to TypeAnnotation
		// TODO: Update ConstDecl struct to accept TypeExpression instead of TypeAnnotation
		switch te := typeExpr.(type) {
		case *ast.TypeAnnotation:
			stmt.Type = te
		case *ast.FunctionPointerTypeNode:
			stmt.Type = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(),
			}
		case *ast.ArrayTypeNode:
			stmt.Type = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(),
			}
		default:
			p.addError("unsupported type expression in const declaration")
			return stmt
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

// parseProgramDeclaration parses an optional program declaration at the start of a file.
// Syntax: program ProgramName;
// The program declaration is optional in DWScript and doesn't affect execution.
// It is parsed and then discarded (not added to the AST).
func (p *Parser) parseProgramDeclaration() {
	// We're on the PROGRAM token
	if !p.curTokenIs(lexer.PROGRAM) {
		return
	}

	// Expect identifier (program name)
	if !p.expectPeek(lexer.IDENT) {
		p.addError("expected program name after 'program' keyword")
		return
	}

	// Note: We could store the program name if needed, but DWScript ignores it
	// programName := p.curToken.Literal

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		p.addError("expected ';' after program name")
		return
	}

	// Successfully parsed program declaration
	// The program name is not stored in the AST as it doesn't affect execution
}
