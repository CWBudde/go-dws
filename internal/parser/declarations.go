package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseConstDeclaration parses one or more constant declarations in a const block.
// Syntax: const NAME = VALUE; or const NAME := VALUE; or const NAME: TYPE = VALUE;
// DWScript allows block syntax: const C1 = 1; C2 = 2; (one const keyword, multiple declarations)
// This function returns a BlockStatement containing all const declarations in the block.
func (p *Parser) parseConstDeclaration() ast.Statement {
	blockToken := p.curToken // Save the initial CONST token for the block
	statements := []ast.Statement{}

	// Parse first const declaration
	firstStmt := p.parseSingleConstDeclaration()
	if firstStmt == nil {
		return nil
	}
	statements = append(statements, firstStmt)

	// Continue parsing additional const declarations without the 'const' keyword
	// As long as the next line looks like a const declaration (not just any identifier)
	for p.looksLikeConstDeclaration() {
		p.nextToken() // move to identifier
		constStmt := p.parseSingleConstDeclaration()
		if constStmt == nil {
			break
		}
		statements = append(statements, constStmt)
	}

	// If only one declaration, return it directly
	if len(statements) == 1 {
		return statements[0]
	}

	// Multiple declarations: wrap in a BlockStatement
	return &ast.BlockStatement{
		Token:      blockToken,
		Statements: statements,
	}
}

// parseSingleConstDeclaration parses a single constant declaration.
// Assumes we're already positioned at the identifier (or just before it).
func (p *Parser) parseSingleConstDeclaration() *ast.ConstDecl {
	// If we're at CONST token, advance to identifier
	if p.curTokenIs(lexer.CONST) {
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
	}

	// We should now be at the identifier
	if !p.isIdentifierToken(p.curToken.Type) {
		p.addError("expected identifier in const declaration", ErrExpectedIdent)
		return nil
	}

	stmt := &ast.ConstDecl{
		BaseNode: ast.BaseNode{
			Token: p.curToken,
		},
	}
	stmt.Name = &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

	// Check for optional type annotation (: Type)
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // move to ':'

		// Parse type expression (can be simple type, function pointer, or array type)
		p.nextToken() // move to type expression
		typeExpr := p.parseTypeExpression()
		if typeExpr == nil {
			p.addError("expected type expression after ':' in const declaration", ErrExpectedType)
			return stmt
		}

		// Convert TypeExpression to TypeAnnotation
		// TODO: Update ConstDecl struct to accept TypeExpression instead of TypeAnnotation
		switch te := typeExpr.(type) {
		case *ast.TypeAnnotation:
			stmt.Type = te
		case *ast.FunctionPointerTypeNode:
			stmt.Type = &ast.TypeAnnotation{
				Token:      te.Token,
				Name:       te.String(),
				InlineType: te, // Store the AST node for semantic analysis
			}
		case *ast.ArrayTypeNode:
			// Check if Token is nil to prevent panics (defensive programming)
			if te == nil {
				p.addError("array type expression is nil in const declaration", ErrInvalidType)
				return stmt
			}
			// Use the array token or create a dummy token if nil
			token := te.Token
			if token.Type == 0 || token.Literal == "" {
				// Create a dummy token to prevent nil pointer issues
				token = lexer.Token{Type: lexer.ARRAY, Literal: "array", Pos: lexer.Position{}}
			}
			stmt.Type = &ast.TypeAnnotation{
				Token:      token,
				Name:       te.String(),
				InlineType: te, // Store the AST node for semantic analysis
			}
		default:
			p.addError("unsupported type expression in const declaration", ErrInvalidType)
			return stmt
		}
	}

	// Expect '=' or ':=' token
	if !p.peekTokenIs(lexer.EQ) && !p.peekTokenIs(lexer.ASSIGN) {
		p.addError("expected '=' or ':=' after const name", ErrMissingAssign)
		return stmt
	}
	p.nextToken() // move to '=' or ':='

	// Parse value expression
	p.nextToken()
	stmt.Value = p.parseExpression(ASSIGN)

	// Check for optional 'deprecated' keyword
	if p.peekTokenIs(lexer.DEPRECATED) {
		p.nextToken() // move to 'deprecated'
		stmt.IsDeprecated = true

		// Check for optional deprecation message string
		if p.peekTokenIs(lexer.STRING) {
			p.nextToken() // move to string
			stmt.DeprecatedMessage = p.curToken.Literal
		}
	}

	// Expect semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
		// End position is at the semicolon
		stmt.EndPos = p.endPosFromToken(p.curToken)
	} else if stmt.Value != nil {
		// No semicolon - end position is after the value expression
		stmt.EndPos = stmt.Value.End()
	} else {
		// Fallback - use current token
		stmt.EndPos = p.endPosFromToken(p.curToken)
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
		p.addError("expected program name after 'program' keyword", ErrExpectedIdent)
		return
	}

	// Note: We could store the program name if needed, but DWScript ignores it
	// programName := p.curToken.Literal

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		p.addError("expected ';' after program name", ErrMissingSemicolon)
		return
	}

	// Successfully parsed program declaration
	// The program name is not stored in the AST as it doesn't affect execution
}
