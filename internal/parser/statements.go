package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseStatementCursor parses a single statement in cursor mode.
// Task 2.2.14.1: Statement parsing infrastructure
// PRE: cursor is on first token of statement
// POST: cursor is on last token of statement
func (p *Parser) parseStatementCursor() ast.Statement {
	// For now, this is a dispatcher that will call cursor versions when available
	// As we implement each statement cursor handler, they'll be added here

	currentToken := p.cursor.Current()

	switch currentToken.Type {
	case lexer.BEGIN:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseBlockStatement()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.VAR:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseVarDeclaration()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.CONST:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseConstDeclaration()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.IF:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseIfStatement()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.WHILE:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseWhileStatement()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.REPEAT:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseRepeatStatement()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.FOR:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseForStatement()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.CASE:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseCaseStatement()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.BREAK:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseBreakStatement()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.CONTINUE:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseContinueStatement()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.EXIT:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseExitStatement()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.TRY:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseTryStatement()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.RAISE:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseRaiseStatement()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.FUNCTION, lexer.PROCEDURE, lexer.METHOD:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseFunctionDeclaration()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.OPERATOR:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseOperatorDeclaration()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.CLASS:
		nextToken := p.cursor.Peek(1)
		if nextToken.Type == lexer.FUNCTION || nextToken.Type == lexer.PROCEDURE || nextToken.Type == lexer.METHOD {
			// Fall back to traditional mode for now
			p.syncCursorToTokens()
			p.useCursor = false
			p.nextToken() // move to function/procedure/method token
			fn := p.parseFunctionDeclaration()
			if fn != nil {
				fn.IsClassMethod = true
			}
			p.useCursor = true
			p.syncTokensToCursor()
			return fn
		}
		p.addError("expected 'function', 'procedure', or 'method' after 'class'", ErrUnexpectedToken)
		return nil

	case lexer.CONSTRUCTOR:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		method := p.parseFunctionDeclaration()
		if method != nil {
			method.IsConstructor = true
		}
		p.useCursor = true
		p.syncTokensToCursor()
		return method

	case lexer.DESTRUCTOR:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		method := p.parseFunctionDeclaration()
		if method != nil {
			method.IsDestructor = true
		}
		p.useCursor = true
		p.syncTokensToCursor()
		return method

	case lexer.TYPE:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseTypeDeclaration()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	case lexer.USES:
		// Fall back to traditional mode for now
		p.syncCursorToTokens()
		p.useCursor = false
		stmt := p.parseUsesClause()
		p.useCursor = true
		p.syncTokensToCursor()
		return stmt

	default:
		// Check for assignment (simple or member assignment)
		// Handle SELF and INHERITED which can be assignment targets (e.g., Self.X := value)
		// Task 2.2.14.2: Use cursor mode for expressions and assignments
		if currentToken.Type == lexer.SELF || currentToken.Type == lexer.INHERITED {
			// Parse as assignment or expression using cursor mode
			return p.parseAssignmentOrExpressionCursor()
		}

		if p.isIdentifierToken(currentToken.Type) {
			// Could be:
			// 1. x := value (assignment)
			// 2. obj.field := value (member assignment)
			// 3. x: Type; (var declaration without 'var' keyword - part of var section)

			// Check if this is a var declaration (IDENT COLON pattern)
			nextToken := p.cursor.Peek(1)
			if nextToken.Type == lexer.COLON {
				// This is a var declaration in a var section
				// Treat it like "var x: Type;"
				// Fall back to traditional mode for var declarations (Task 2.2.14.6)
				p.syncCursorToTokens()
				p.useCursor = false
				stmt := p.parseVarDeclaration()
				p.useCursor = true
				p.syncTokensToCursor()
				return stmt
			}

			// Otherwise, parse as assignment or expression using cursor mode
			return p.parseAssignmentOrExpressionCursor()
		}

		// Expression statement - use cursor mode (expressions are fully migrated)
		return p.parseExpressionStatementCursor()
	}
}

// parseStatement parses a single statement.
// PRE: curToken is first token of statement
// POST: curToken is last token of statement
func (p *Parser) parseStatement() ast.Statement {
	// Task 2.2.14.2: Dispatch to cursor mode if enabled
	if p.useCursor {
		stmt := p.parseStatementCursor()
		// Sync traditional pointers from cursor after statement parsing
		p.syncCursorToTokens()
		return stmt
	}

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
		// Handle SELF and INHERITED which can be assignment targets (e.g., Self.X := value)
		if p.curToken.Type == lexer.SELF || p.curToken.Type == lexer.INHERITED {
			return p.parseAssignmentOrExpression()
		}

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
// PRE: curToken is BEGIN
// POST: curToken is END
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}
	block.Statements = []ast.Statement{}

	// Track block context for better error messages
	p.pushBlockContext("begin", p.curToken.Pos)
	defer p.popBlockContext()

	p.nextToken() // advance past 'begin'

	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) && !p.curTokenIs(lexer.ENSURE) {
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

	if !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.ENSURE) {
		p.addErrorWithContext("expected 'end' to close block", ErrMissingEnd)
		// Use synchronize for better error recovery
		p.synchronize([]lexer.TokenType{lexer.END, lexer.ENSURE})
	}

	// Set end position to the END keyword
	block.EndPos = p.endPosFromToken(p.curToken)

	return block
}

// parseExpressionStatement parses an expression statement.
// PRE: curToken is first token of expression
// POST: curToken is SEMICOLON if present, otherwise last token of expression
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}

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
//
// PRE: curToken is VAR or IDENT
// POST: curToken is SEMICOLON of last var declaration
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
		BaseNode:   ast.BaseNode{Token: blockToken},
		Statements: statements,
	}
}

// parseSingleVarDeclaration parses a single variable declaration.
// Assumes we're already positioned at the identifier (or just before it).
// PRE: curToken is VAR or variable name IDENT
// POST: curToken is SEMICOLON
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
		// Use structured error (Task 2.1.3)
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrExpectedIdent).
			WithMessage("expected identifier in var declaration").
			WithPosition(p.curToken.Pos, p.curToken.Length()).
			WithExpectedString("variable name").
			WithActual(p.curToken.Type, p.curToken.Literal).
			WithSuggestion("provide a variable name after 'var'").
			WithParsePhase("variable declaration").
			Build()
		p.addStructuredError(err)
		return nil
	} else {
		stmt.Token = p.curToken
	}

	// Collect comma-separated identifiers
	// Parse pattern: IDENT (, IDENT)* : TYPE [:= VALUE]
	stmt.Names = []*ast.Identifier{}
	for {
		if !p.isIdentifierToken(p.curToken.Type) {
			// Use structured error (Task 2.1.3)
			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrExpectedIdent).
				WithMessage("expected identifier in var declaration").
				WithPosition(p.curToken.Pos, p.curToken.Length()).
				WithExpectedString("variable name").
				WithActual(p.curToken.Type, p.curToken.Literal).
				WithSuggestion("provide a variable name").
				WithParsePhase("variable declaration").
				Build()
			p.addStructuredError(err)
			return nil
		}

		stmt.Names = append(stmt.Names, &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: p.curToken,
				},
			},
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
			// Use structured error (Task 2.1.3)
			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrExpectedType).
				WithMessage("expected type expression after ':' in var declaration").
				WithPosition(p.curToken.Pos, p.curToken.Length()).
				WithExpectedString("type name").
				WithSuggestion("specify the variable type, like 'Integer' or 'String'").
				WithParsePhase("variable declaration type").
				Build()
			p.addStructuredError(err)
			return stmt
		}

		// Directly assign the type expression without creating synthetic wrappers
		stmt.Type = typeExpr
	}

	if hasExplicitType {
		if p.peekTokenIs(lexer.ASSIGN) || p.peekTokenIs(lexer.EQ) {
			if len(stmt.Names) > 1 {
				// Use structured error (Task 2.1.3)
				err := NewStructuredError(ErrKindInvalid).
					WithCode(ErrInvalidSyntax).
					WithMessage("cannot use initializer with multiple variable names").
					WithPosition(p.peekToken.Pos, p.peekToken.Length()).
					WithSuggestion("declare variables separately or use the same value for all").
					WithNote("DWScript requires: var x, y: Integer (no initializer) or var x: Integer := 10").
					WithParsePhase("variable declaration").
					Build()
				p.addStructuredError(err)
				return stmt
			}

			p.nextToken() // move to assignment operator
			p.nextToken()
			stmt.Value = p.parseExpression(ASSIGN)
		}
	} else {
		if p.peekTokenIs(lexer.ASSIGN) || p.peekTokenIs(lexer.EQ) {
			if len(stmt.Names) > 1 {
				// Use structured error (Task 2.1.3)
				err := NewStructuredError(ErrKindInvalid).
					WithCode(ErrInvalidSyntax).
					WithMessage("cannot use initializer with multiple variable names").
					WithPosition(p.peekToken.Pos, p.peekToken.Length()).
					WithSuggestion("declare variables separately when using initializers").
					WithNote("DWScript requires separate declarations with initializers").
					WithParsePhase("variable declaration").
					Build()
				p.addStructuredError(err)
				return stmt
			}

			p.nextToken() // move to assignment operator
			stmt.Inferred = true
			p.nextToken()
			stmt.Value = p.parseExpression(ASSIGN)
		} else if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.EXTERNAL) {
			// Use structured error (Task 2.1.3)
			err := NewStructuredError(ErrKindInvalid).
				WithCode(ErrInvalidSyntax).
				WithMessage("variable declaration requires a type or initializer").
				WithPosition(p.curToken.Pos, p.curToken.Length()).
				WithSuggestion("add ': TypeName' or ':= value' after the variable name").
				WithNote("Examples: 'var x: Integer' or 'var x := 10'").
				WithParsePhase("variable declaration").
				Build()
			p.addStructuredError(err)
		} else {
			// Use structured error (Task 2.1.3)
			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrMissingColon).
				WithMessage("expected ':', ':=' or '=' in variable declaration").
				WithPosition(p.peekToken.Pos, p.peekToken.Length()).
				WithExpectedString("':' or ':='").
				WithActual(p.peekToken.Type, p.peekToken.Literal).
				WithSuggestion("add ':' for type declaration or ':=' for type inference").
				WithParsePhase("variable declaration").
				Build()
			p.addStructuredError(err)
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
				BaseNode: ast.BaseNode{Token: p.curToken},
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
				BaseNode: ast.BaseNode{Token: p.curToken},
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
				BaseNode: ast.BaseNode{Token: p.curToken},
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
			// Use structured error (Task 2.1.3)
			err := NewStructuredError(ErrKindInvalid).
				WithCode(ErrInvalidSyntax).
				WithMessage("invalid assignment target").
				WithPosition(p.curToken.Pos, p.curToken.Length()).
				WithSuggestion("assignment target must be a variable, field access, or array element").
				WithNote("Valid: x := 10, obj.field := 20, arr[i] := 30").
				WithParsePhase("assignment statement").
				Build()
			p.addStructuredError(err)
			return nil
		}
	}

	// Not an assignment, treat as expression statement
	stmt := &ast.ExpressionStatement{
		BaseNode:   ast.BaseNode{Token: startToken},
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
// - COMMA (for multi-var declaration: x, y : Integer)
// This prevents mis-parsing function calls or assignments as var declarations.
// NOTE: We do NOT accept ASSIGN/EQ patterns here because they're ambiguous:
//
//	"x := 5" could be a new var declaration OR an assignment to existing var.
//	To use type inference in var blocks, repeat the 'var' keyword for each declaration.
func (p *Parser) looksLikeVarDeclaration() bool {
	if !p.peekTokenIs(lexer.IDENT) {
		return false
	}

	// Use peek(0) for 2-token lookahead: check the token after peekToken
	// peekToken is the IDENT, peek(0) returns what comes after it
	tokenAfterIdent := p.peek(0)

	// Only accept unambiguous var declaration patterns:
	// - name : Type         (explicit type - always a declaration)
	// - name, name2 : Type  (multi-var declaration)
	return tokenAfterIdent.Type == lexer.COLON ||
		tokenAfterIdent.Type == lexer.COMMA
}

// looksLikeConstDeclaration performs lookahead to check if the next tokens form a const declaration.
// A const declaration pattern is: IDENT followed by either:
// - COLON (for typed const: C : Integer = 5)
// - EQ (for untyped const: C = value)
// This prevents mis-parsing other statements as const declarations.
// NOTE: We accept EQ but not ASSIGN here. While both can be used for consts,
//
//	ASSIGN is more commonly used for var declarations and assignments.
func (p *Parser) looksLikeConstDeclaration() bool {
	if !p.peekTokenIs(lexer.IDENT) {
		return false
	}

	// Use peek(0) for 2-token lookahead: check the token after peekToken
	// peekToken is the IDENT, peek(0) returns what comes after it
	tokenAfterIdent := p.peek(0)

	// Const declaration patterns:
	// - NAME : Type = value  (typed const)
	// - NAME = value         (untyped const)
	return tokenAfterIdent.Type == lexer.COLON ||
		tokenAfterIdent.Type == lexer.EQ
}

// ============================================================================
// Task 2.2.14.2: Cursor-mode statement handlers for expressions and assignments
// ============================================================================

// parseExpressionStatementCursor parses an expression statement in cursor mode.
// Task 2.2.14.2: Expression statement migration
// PRE: cursor is on first token of expression
// POST: cursor is on last token of statement (possibly SEMICOLON)
func (p *Parser) parseExpressionStatementCursor() *ast.ExpressionStatement {
	startToken := p.cursor.Current()
	stmt := &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{Token: startToken},
	}

	// Parse expression using cursor mode (expressions are fully migrated)
	stmt.Expression = p.parseExpressionCursor(LOWEST)

	// Set end position based on expression
	if stmt.Expression != nil {
		stmt.EndPos = stmt.Expression.End()
	} else {
		stmt.EndPos = p.endPosFromToken(stmt.Token)
	}

	// Optional semicolon
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.SEMICOLON {
		p.cursor = p.cursor.Advance()
		stmt.EndPos = p.endPosFromToken(p.cursor.Current()) // Update to include semicolon
	}

	return stmt
}

// parseAssignmentOrExpressionCursor determines if we have an assignment or expression statement.
// Task 2.2.14.2: Assignment statement migration
// This handles both simple assignments (x := value), compound assignments (x += value),
// and member assignments (obj.field := value, arr[i] := value).
// PRE: cursor is on first token (typically an identifier or expression start)
// POST: cursor is on last token of statement (possibly SEMICOLON)
func (p *Parser) parseAssignmentOrExpressionCursor() ast.Statement {
	startToken := p.cursor.Current()

	// Parse the left side as an expression (could be identifier, member access, or index)
	left := p.parseExpressionCursor(LOWEST)

	// Check if next token is assignment (simple or compound)
	nextToken := p.cursor.Peek(1)
	if isAssignmentOperator(nextToken.Type) {
		p.cursor = p.cursor.Advance() // Move to assignment operator
		assignOp := p.cursor.Current().Type

		// Determine what kind of assignment this is
		switch leftExpr := left.(type) {
		case *ast.Identifier, *ast.MemberAccessExpression, *ast.IndexExpression:
			// Valid assignment targets
			stmt := &ast.AssignmentStatement{
				BaseNode: ast.BaseNode{Token: p.cursor.Current()},
				Target:   leftExpr,
				Operator: assignOp,
			}

			// Move to value expression
			p.cursor = p.cursor.Advance()

			// Parse value expression
			stmt.Value = p.parseExpressionCursor(ASSIGN)

			// Set end position based on value expression
			if stmt.Value != nil {
				stmt.EndPos = stmt.Value.End()
			} else {
				stmt.EndPos = p.endPosFromToken(p.cursor.Current())
			}

			// Optional semicolon
			nextToken := p.cursor.Peek(1)
			if nextToken.Type == lexer.SEMICOLON {
				p.cursor = p.cursor.Advance()
				stmt.EndPos = p.endPosFromToken(p.cursor.Current()) // Update to include semicolon
			}

			return stmt

		default:
			// Invalid assignment target - use structured error
			err := NewStructuredError(ErrKindInvalid).
				WithCode(ErrInvalidSyntax).
				WithMessage("invalid assignment target").
				WithPosition(p.cursor.Current().Pos, p.cursor.Current().Length()).
				WithSuggestion("assignment target must be a variable, field access, or array element").
				WithNote("Valid: x := 10, obj.field := 20, arr[i] := 30").
				WithParsePhase("assignment statement").
				Build()
			p.addStructuredError(err)
			return nil
		}
	}

	// Not an assignment, treat as expression statement
	stmt := &ast.ExpressionStatement{
		BaseNode:   ast.BaseNode{Token: startToken},
		Expression: left,
	}

	// Set end position
	if left != nil {
		stmt.EndPos = left.End()
	} else {
		stmt.EndPos = p.endPosFromToken(startToken)
	}

	// Optional semicolon
	nextToken = p.cursor.Peek(1)
	if nextToken.Type == lexer.SEMICOLON {
		p.cursor = p.cursor.Advance()
		stmt.EndPos = p.endPosFromToken(p.cursor.Current()) // Update to include semicolon
	}

	return stmt
}
