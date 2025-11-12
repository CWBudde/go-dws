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
	case lexer.FUNCTION, lexer.PROCEDURE, lexer.METHOD:
		return p.parseFunctionDeclaration()
	case lexer.OPERATOR:
		return p.parseOperatorDeclaration()
	case lexer.CLASS:
		if p.peekTokenIs(lexer.FUNCTION) || p.peekTokenIs(lexer.PROCEDURE) || p.peekTokenIs(lexer.METHOD) {
			p.nextToken() // move to function/procedure/method token
			fn := p.parseFunctionDeclaration()
			if fn != nil {
				fn.IsClassMethod = true
			}
			return fn
		}
		p.addError("expected 'function', 'procedure', or 'method' after 'class'", ErrUnexpectedToken)
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
		// Dispatch to class or interface parser
		// Both parsers will handle the full parsing starting from TYPE token
		return p.parseTypeDeclaration()
	case lexer.USES:
		// Parse uses clause at program level
		return p.parseUsesClause()
	default:
		// Check for assignment (simple or member assignment)
		if p.isIdentifierToken(p.curToken.Type) {
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
		p.addError("expected 'end' to close block", ErrMissingEnd)
		for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}
	}

	// Set end position to the END keyword
	block.EndPos = p.endPosFromToken(p.curToken)

	return block
}

// parseExpressionStatement parses an expression statement.
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	// Set end position based on expression
	if stmt.Expression != nil {
		stmt.EndPos = stmt.Expression.End()
	} else {
		stmt.EndPos = p.endPosFromToken(stmt.Token)
	}

	// Optional semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
		stmt.EndPos = p.endPosFromToken(p.curToken) // Update to include semicolon
	}

	return stmt
}

// parseVarDeclaration parses a variable declaration statement.
// Can be called in two contexts:
//  1. After 'var' keyword: var x: Integer;
//  2. In a var section without 'var': x: Integer; (curToken is already the IDENT)
func (p *Parser) parseVarDeclaration() ast.Statement {
	blockToken := p.curToken // Save the initial VAR token for the block
	statements := []ast.Statement{}

	// Parse first var declaration
	firstStmt := p.parseSingleVarDeclaration()
	if firstStmt == nil {
		return nil
	}
	statements = append(statements, firstStmt)

	// Continue parsing additional var declarations without the 'var' keyword
	// As long as the next line looks like a var declaration (not just any identifier)
	for p.looksLikeVarDeclaration() {
		p.nextToken() // move to identifier
		varStmt := p.parseSingleVarDeclaration()
		if varStmt == nil {
			break
		}
		statements = append(statements, varStmt)
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

// parseSingleVarDeclaration parses a single variable declaration.
// Assumes we're already positioned at the identifier (or just before it).
func (p *Parser) parseSingleVarDeclaration() *ast.VarDeclStatement {
	stmt := &ast.VarDeclStatement{}

	// Check if we're already at the identifier (var section continuation)
	// or if we need to advance to it (after 'var' keyword)
	if p.curTokenIs(lexer.VAR) {
		stmt.Token = p.curToken
		// After 'var' keyword, expect identifier next
		if !p.expectIdentifier() {
			return nil
		}
	} else if !p.isIdentifierToken(p.curToken.Type) {
		// Should already be at an identifier
		p.addError("expected identifier in var declaration", ErrExpectedIdent)
		return nil
	} else {
		stmt.Token = p.curToken
	}

	// Collect comma-separated identifiers
	// Parse pattern: IDENT (, IDENT)* : TYPE [:= VALUE]
	stmt.Names = []*ast.Identifier{}
	for {
		if !p.isIdentifierToken(p.curToken.Type) {
			p.addError("expected identifier in var declaration", ErrExpectedIdent)
			return nil
		}

		stmt.Names = append(stmt.Names, &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		})

		// Check if there are more names (comma-separated)
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // move to ','
			if !p.expectIdentifier() {
				return nil
			}
			continue
		}

		// No more names, break to parse type
		break
	}

	hasExplicitType := false
	if p.peekTokenIs(lexer.COLON) {
		hasExplicitType = true
		p.nextToken() // move to ':'

		// Parse type expression (can be simple type, function pointer, or array type)
		p.nextToken() // move to type expression
		typeExpr := p.parseTypeExpression()
		if typeExpr == nil {
			p.addError("expected type expression after ':' in var declaration", ErrExpectedType)
			return stmt
		}

		// For now, we need to convert TypeExpression to TypeAnnotation for VarDeclStatement.Type
		// TODO: Update VarDeclStatement struct to accept TypeExpression instead of TypeAnnotation
		switch te := typeExpr.(type) {
		case *ast.TypeAnnotation:
			stmt.Type = te
		case *ast.FunctionPointerTypeNode:
			// For function pointer types, we create a synthetic TypeAnnotation
			stmt.Type = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(), // Use the full function pointer signature as the type name
			}
		case *ast.ArrayTypeNode:
			// For array types, we create a synthetic TypeAnnotation
			// Check if Token is nil to prevent panics (defensive programming)
			if te == nil {
				p.addError("array type expression is nil in var declaration", ErrInvalidType)
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
				Name:       te.String(), // Use the full array type signature as the type name
				InlineType: te,          // Store the AST node for semantic analyzer to evaluate bounds
			}
		case *ast.SetTypeNode:
			// For set types, we create a synthetic TypeAnnotation
			stmt.Type = &ast.TypeAnnotation{
				Token:      te.Token,
				Name:       te.String(), // Use the full set type signature as the type name
				InlineType: te,          // Store the AST node for semantic analyzer
			}
		case *ast.ClassOfTypeNode:
			// For metaclass types, we create a synthetic TypeAnnotation
			stmt.Type = &ast.TypeAnnotation{
				Token:      te.Token,
				Name:       te.String(), // Use the full metaclass type signature as the type name
				InlineType: te,          // Store the AST node for semantic analyzer
			}
		default:
			p.addError("unsupported type expression in var declaration", ErrInvalidType)
			return stmt
		}
	}

	if hasExplicitType {
		if p.peekTokenIs(lexer.ASSIGN) || p.peekTokenIs(lexer.EQ) {
			if len(stmt.Names) > 1 {
				p.addError("cannot use initializer with multiple variable names", ErrInvalidSyntax)
				return stmt
			}

			p.nextToken() // move to assignment operator
			p.nextToken()
			stmt.Value = p.parseExpression(ASSIGN)
		}
	} else {
		if p.peekTokenIs(lexer.ASSIGN) || p.peekTokenIs(lexer.EQ) {
			if len(stmt.Names) > 1 {
				p.addError("cannot use initializer with multiple variable names", ErrInvalidSyntax)
				return stmt
			}

			p.nextToken() // move to assignment operator
			stmt.Inferred = true
			p.nextToken()
			stmt.Value = p.parseExpression(ASSIGN)
		} else if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.EXTERNAL) {
			p.addError("variable declaration requires a type or initializer", ErrInvalidSyntax)
		} else {
			p.addError("expected ':', ':=' or '=' in variable declaration", ErrMissingColon)
		}
	}

	// Check for 'external' keyword
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
		return nil
	}

	// End position is at the semicolon
	stmt.EndPos = p.endPosFromToken(p.curToken)

	return stmt
}

// isAssignmentOperator checks if the given token type is an assignment operator.
func isAssignmentOperator(t lexer.TokenType) bool {
	return t == lexer.ASSIGN ||
		t == lexer.PLUS_ASSIGN ||
		t == lexer.MINUS_ASSIGN ||
		t == lexer.TIMES_ASSIGN ||
		t == lexer.DIVIDE_ASSIGN
}

// parseAssignmentOrExpression determines if we have an assignment or expression statement.
// This handles both simple assignments (x := value), compound assignments (x += value),
// and member assignments (obj.field := value).
func (p *Parser) parseAssignmentOrExpression() ast.Statement {
	// Save starting position
	startToken := p.curToken

	// Parse the left side as an expression (could be identifier or member access)
	left := p.parseExpression(LOWEST)

	// Check if next token is assignment (simple or compound)
	if isAssignmentOperator(p.peekToken.Type) {
		p.nextToken() // move to assignment operator
		assignOp := p.curToken.Type

		// Determine what kind of assignment this is
		switch leftExpr := left.(type) {
		case *ast.Identifier:
			// Simple or compound assignment: x := value, x += value
			stmt := &ast.AssignmentStatement{
				Token:    p.curToken,
				Target:   leftExpr,
				Operator: assignOp,
			}
			p.nextToken()
			stmt.Value = p.parseExpression(ASSIGN)

			// Set end position based on value expression
			if stmt.Value != nil {
				stmt.EndPos = stmt.Value.End()
			} else {
				stmt.EndPos = p.endPosFromToken(p.curToken)
			}

			// Optional semicolon
			if p.peekTokenIs(lexer.SEMICOLON) {
				p.nextToken()
				stmt.EndPos = p.endPosFromToken(p.curToken) // Update to include semicolon
			}
			return stmt

		case *ast.MemberAccessExpression:
			// Member assignment: obj.field := value, obj.field += value
			stmt := &ast.AssignmentStatement{
				Token:    p.curToken,
				Target:   leftExpr,
				Operator: assignOp,
			}

			p.nextToken()
			stmt.Value = p.parseExpression(ASSIGN)

			// Set end position based on value expression
			if stmt.Value != nil {
				stmt.EndPos = stmt.Value.End()
			} else {
				stmt.EndPos = p.endPosFromToken(p.curToken)
			}

			// Optional semicolon
			if p.peekTokenIs(lexer.SEMICOLON) {
				p.nextToken()
				stmt.EndPos = p.endPosFromToken(p.curToken) // Update to include semicolon
			}
			return stmt

		case *ast.IndexExpression:
			// Array index assignment: arr[i] := value, arr[i] += value
			stmt := &ast.AssignmentStatement{
				Token:    p.curToken,
				Target:   leftExpr,
				Operator: assignOp,
			}

			p.nextToken()
			stmt.Value = p.parseExpression(ASSIGN)

			// Set end position based on value expression
			if stmt.Value != nil {
				stmt.EndPos = stmt.Value.End()
			} else {
				stmt.EndPos = p.endPosFromToken(p.curToken)
			}

			// Optional semicolon
			if p.peekTokenIs(lexer.SEMICOLON) {
				p.nextToken()
				stmt.EndPos = p.endPosFromToken(p.curToken) // Update to include semicolon
			}
			return stmt

		default:
			p.addError("invalid assignment target", ErrInvalidSyntax)
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

// looksLikeVarDeclaration performs lookahead to check if the next tokens form a var declaration.
// A var declaration pattern is: IDENT followed by either:
// - COLON (for typed declaration: x : Integer)
// - ASSIGN or EQ (for inferred type: x := 5)
// - COMMA (for multi-var declaration: x, y : Integer)
// This prevents mis-parsing function calls or other statements as var declarations.
func (p *Parser) looksLikeVarDeclaration() bool {
	if !p.peekTokenIs(lexer.IDENT) {
		return false
	}

	// We can't look at token after peek without advancing, but we know common patterns
	// For now, conservatively return false to avoid the regression
	// A proper implementation would require a 2-token lookahead buffer
	return false
}

// looksLikeConstDeclaration performs lookahead to check if the next tokens form a const declaration.
// A const declaration pattern is: IDENT followed by either:
// - COLON (for typed const: C : Integer = 5)
// - EQ or ASSIGN (for untyped const: C = 5)
// This prevents mis-parsing other statements as const declarations.
func (p *Parser) looksLikeConstDeclaration() bool {
	if !p.peekTokenIs(lexer.IDENT) {
		return false
	}

	// For now, conservatively return false to avoid the regression
	// A proper implementation would require a 2-token lookahead buffer
	return false
}
