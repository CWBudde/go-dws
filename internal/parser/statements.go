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
		// Task 2.2.14.3: Use cursor mode for block statements
		return p.parseBlockStatementCursor()

	case lexer.VAR:
		// Task 2.2.14.6: Use cursor mode for var declarations
		return p.parseVarDeclarationCursor()

	case lexer.CONST:
		// Task 2.2.14.6: Use cursor mode for const declarations
		return p.parseConstDeclarationCursor()

	case lexer.IF:
		// Task 2.2.14.4: Use cursor mode for if statements
		return p.parseIfStatementCursor()

	case lexer.WHILE:
		// Task 2.2.14.4: Use cursor mode for while statements
		return p.parseWhileStatementCursor()

	case lexer.REPEAT:
		// Task 2.2.14.4: Use cursor mode for repeat statements
		return p.parseRepeatStatementCursor()

	case lexer.FOR:
		// Task 2.2.14.5: Use cursor mode for for statements
		return p.parseForStatementCursor()

	case lexer.CASE:
		// Task 2.2.14.5: Use cursor mode for case statements
		return p.parseCaseStatementCursor()

	case lexer.BREAK:
		// Task 2.2.14.8: Use cursor mode for break statements
		return p.parseBreakStatementCursor()

	case lexer.CONTINUE:
		// Task 2.2.14.8: Use cursor mode for continue statements
		return p.parseContinueStatementCursor()

	case lexer.EXIT:
		// Task 2.2.14.8: Use cursor mode for exit statements
		return p.parseExitStatementCursor()

	case lexer.TRY:
		// Task 2.2.14.7: Use cursor mode for try statements
		return p.parseTryStatementCursor()

	case lexer.RAISE:
		// Task 2.2.14.7: Use cursor mode for raise statements
		return p.parseRaiseStatementCursor()

	case lexer.FUNCTION, lexer.PROCEDURE, lexer.METHOD:
		return p.parseFunctionDeclarationCursor()

	case lexer.OPERATOR:
		return p.parseOperatorDeclarationCursor()

	case lexer.CLASS:
		nextToken := p.cursor.Peek(1)
		if nextToken.Type == lexer.FUNCTION || nextToken.Type == lexer.PROCEDURE || nextToken.Type == lexer.METHOD {
			p.cursor = p.cursor.Advance() // move to function/procedure/method token
			fn := p.parseFunctionDeclarationCursor()
			if fn != nil {
				fn.IsClassMethod = true
			}
			return fn
		}
		p.addError("expected 'function', 'procedure', or 'method' after 'class'", ErrUnexpectedToken)
		return nil

	case lexer.CONSTRUCTOR:
		method := p.parseFunctionDeclarationCursor()
		if method != nil {
			method.IsConstructor = true
		}
		return method

	case lexer.DESTRUCTOR:
		method := p.parseFunctionDeclarationCursor()
		if method != nil {
			method.IsDestructor = true
		}
		return method

	case lexer.TYPE:
		return p.parseTypeDeclarationCursor()

	case lexer.USES:
		return p.parseUsesClauseCursor()

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
				return p.parseVarDeclarationCursor()
			}

			// Otherwise, parse as assignment or expression using cursor mode
			return p.parseAssignmentOrExpressionCursor()
		}

		// Expression statement - use cursor mode (expressions are fully migrated)
		return p.parseExpressionStatementCursor()
	}
}

// parseBlockStatement parses a begin...end block.
// PRE: curToken is BEGIN
// POST: curToken is END
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	builder := p.StartNode()

	block := &ast.BlockStatement{
		BaseNode: ast.BaseNode{Token: p.cursor.Current()},
	}
	block.Statements = []ast.Statement{}

	// Track block context for better error messages
	p.pushBlockContext("begin", p.cursor.Current().Pos)
	defer p.popBlockContext()

	p.nextToken() // advance past 'begin'

	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) && !p.curTokenIs(lexer.ENSURE) {
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatementCursor()
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

	return builder.Finish(block).(*ast.BlockStatement)
}

// parseExpressionStatement parses an expression statement.
// PRE: curToken is first token of expression
// POST: curToken is SEMICOLON if present, otherwise last token of expression
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	builder := p.StartNode()

	stmt := &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{Token: p.cursor.Current()},
	}

	stmt.Expression = p.parseExpressionCursor(LOWEST)

	// Optional semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
		// End at semicolon
		return builder.Finish(stmt).(*ast.ExpressionStatement)
	}

	// End at expression
	return builder.FinishWithNode(stmt, stmt.Expression).(*ast.ExpressionStatement)
}

// parseVarDeclaration parses a variable declaration statement.
// Can be called in two contexts:
//  1. After 'var' keyword: var x: Integer;
//  2. In a var section without 'var': x: Integer; (curToken is already the IDENT)
//
// PRE: curToken is VAR or IDENT
// POST: curToken is SEMICOLON of last var declaration
// Dispatcher: delegates to cursor or traditional mode
func (p *Parser) parseVarDeclaration() ast.Statement {
	return p.parseVarDeclarationCursor()
}

// parseVarDeclarationTraditional parses var declarations using traditional mode.
func (p *Parser) parseSingleVarDeclaration() *ast.VarDeclStatement {
	builder := p.StartNode()
	stmt := &ast.VarDeclStatement{}

	// Check if we're already at the identifier (var section continuation)
	// or if we need to advance to it (after 'var' keyword)
	if p.curTokenIs(lexer.VAR) {
		stmt.Token = p.cursor.Current()
		// After 'var' keyword, expect identifier next
		if !p.expectIdentifier() {
			return nil
		}
	} else if !p.isIdentifierToken(p.cursor.Current().Type) {
		// Should already be at an identifier
		// Task 2.7.7: Dual-mode - get current token for error reporting
		var curTok lexer.Token
		if p.cursor != nil {
			curTok = p.cursor.Current()
		} else {
			curTok = p.cursor.Current()
		}

		// Use structured error (Task 2.1.3)
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrExpectedIdent).
			WithMessage("expected identifier in var declaration").
			WithPosition(curTok.Pos, curTok.Length()).
			WithExpectedString("variable name").
			WithActual(curTok.Type, curTok.Literal).
			WithSuggestion("provide a variable name after 'var'").
			WithParsePhase("variable declaration").
			Build()
		p.addStructuredError(err)
		return nil
	} else {
		stmt.Token = p.cursor.Current()
	}

	// Use IdentifierList combinator to collect comma-separated identifiers (Task 2.3.3)
	// Parse pattern: IDENT (, IDENT)* : TYPE [:= VALUE]
	stmt.Names = p.IdentifierList(IdentifierListConfig{
		ErrorContext:      "variable declaration",
		RequireAtLeastOne: true,
	})
	if stmt.Names == nil {
		return nil
	}

	// Use OptionalTypeAnnotation combinator (Task 2.3.3)
	stmt.Type = p.OptionalTypeAnnotation()

	if stmt.Type != nil {
		if p.peekTokenIs(lexer.ASSIGN) || p.peekTokenIs(lexer.EQ) {
			if len(stmt.Names) > 1 {
				// Task 2.7.7: Dual-mode - get peek token for error reporting
				var peekTok lexer.Token
				if p.cursor != nil {
					peekTok = p.cursor.Peek(1)
				} else {
					peekTok = p.cursor.Peek(1)
				}

				// Use structured error (Task 2.1.3)
				err := NewStructuredError(ErrKindInvalid).
					WithCode(ErrInvalidSyntax).
					WithMessage("cannot use initializer with multiple variable names").
					WithPosition(peekTok.Pos, peekTok.Length()).
					WithSuggestion("declare variables separately or use the same value for all").
					WithNote("DWScript requires: var x, y: Integer (no initializer) or var x: Integer := 10").
					WithParsePhase("variable declaration").
					Build()
				p.addStructuredError(err)
				return stmt
			}

			p.nextToken() // move to assignment operator
			p.nextToken()
			stmt.Value = p.parseExpressionCursor(ASSIGN)
		}
	} else {
		if p.peekTokenIs(lexer.ASSIGN) || p.peekTokenIs(lexer.EQ) {
			if len(stmt.Names) > 1 {
				// Task 2.7.7: Dual-mode - get peek token for error reporting
				var peekTok lexer.Token
				if p.cursor != nil {
					peekTok = p.cursor.Peek(1)
				} else {
					peekTok = p.cursor.Peek(1)
				}

				// Use structured error (Task 2.1.3)
				err := NewStructuredError(ErrKindInvalid).
					WithCode(ErrInvalidSyntax).
					WithMessage("cannot use initializer with multiple variable names").
					WithPosition(peekTok.Pos, peekTok.Length()).
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
			stmt.Value = p.parseExpressionCursor(ASSIGN)
		} else if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.EXTERNAL) {
			// Task 2.7.7: Dual-mode - get current token for error reporting
			var curTok lexer.Token
			if p.cursor != nil {
				curTok = p.cursor.Current()
			} else {
				curTok = p.cursor.Current()
			}

			// Use structured error (Task 2.1.3)
			err := NewStructuredError(ErrKindInvalid).
				WithCode(ErrInvalidSyntax).
				WithMessage("variable declaration requires a type or initializer").
				WithPosition(curTok.Pos, curTok.Length()).
				WithSuggestion("add ': TypeName' or ':= value' after the variable name").
				WithNote("Examples: 'var x: Integer' or 'var x := 10'").
				WithParsePhase("variable declaration").
				Build()
			p.addStructuredError(err)
		} else {
			// Task 2.7.7: Dual-mode - get peek token for error reporting
			var peekTok lexer.Token
			if p.cursor != nil {
				peekTok = p.cursor.Peek(1)
			} else {
				peekTok = p.cursor.Peek(1)
			}

			// Use structured error (Task 2.1.3)
			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrMissingColon).
				WithMessage("expected ':', ':=' or '=' in variable declaration").
				WithPosition(peekTok.Pos, peekTok.Length()).
				WithExpectedString("':' or ':='").
				WithActual(peekTok.Type, peekTok.Literal).
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
			stmt.ExternalName = p.cursor.Current().Literal
		}
	}

	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return builder.Finish(stmt).(*ast.VarDeclStatement)
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
	builder := p.StartNode()

	// Save starting position
	startToken := p.cursor.Current()

	// Parse the left side as an expression (could be identifier or member access)
	left := p.parseExpressionCursor(LOWEST)

	// Check if next token is assignment (simple or compound)
	if isAssignmentOperator(p.cursor.Peek(1).Type) {
		p.nextToken() // move to assignment operator
		assignOp := p.cursor.Current().Type

		// Determine what kind of assignment this is
		switch leftExpr := left.(type) {
		case *ast.Identifier:
			// Simple or compound assignment: x := value, x += value
			stmt := &ast.AssignmentStatement{
				BaseNode: ast.BaseNode{Token: p.cursor.Current()},
				Target:   leftExpr,
				Operator: assignOp,
			}
			p.nextToken()
			stmt.Value = p.parseExpressionCursor(ASSIGN)

			// Optional semicolon
			if p.peekTokenIs(lexer.SEMICOLON) {
				p.nextToken()
				return builder.Finish(stmt).(*ast.AssignmentStatement)
			}

			// End at value expression (FinishWithNode handles nil by falling back to current token)
			return builder.FinishWithNode(stmt, stmt.Value).(*ast.AssignmentStatement)

		case *ast.MemberAccessExpression:
			// Member assignment: obj.field := value, obj.field += value
			stmt := &ast.AssignmentStatement{
				BaseNode: ast.BaseNode{Token: p.cursor.Current()},
				Target:   leftExpr,
				Operator: assignOp,
			}

			p.nextToken()
			stmt.Value = p.parseExpressionCursor(ASSIGN)

			// Optional semicolon
			if p.peekTokenIs(lexer.SEMICOLON) {
				p.nextToken()
				return builder.Finish(stmt).(*ast.AssignmentStatement)
			}

			// End at value expression (FinishWithNode handles nil by falling back to current token)
			return builder.FinishWithNode(stmt, stmt.Value).(*ast.AssignmentStatement)

		case *ast.IndexExpression:
			// Array index assignment: arr[i] := value, arr[i] += value
			stmt := &ast.AssignmentStatement{
				BaseNode: ast.BaseNode{Token: p.cursor.Current()},
				Target:   leftExpr,
				Operator: assignOp,
			}

			p.nextToken()
			stmt.Value = p.parseExpressionCursor(ASSIGN)

			// Optional semicolon
			if p.peekTokenIs(lexer.SEMICOLON) {
				p.nextToken()
				return builder.Finish(stmt).(*ast.AssignmentStatement)
			}

			// End at value expression (FinishWithNode handles nil by falling back to current token)
			return builder.FinishWithNode(stmt, stmt.Value).(*ast.AssignmentStatement)

		default:
			// Task 2.7.7: Dual-mode - get current token for error reporting
			var curTok lexer.Token
			if p.cursor != nil {
				curTok = p.cursor.Current()
			} else {
				curTok = p.cursor.Current()
			}

			// Use structured error (Task 2.1.3)
			err := NewStructuredError(ErrKindInvalid).
				WithCode(ErrInvalidSyntax).
				WithMessage("invalid assignment target").
				WithPosition(curTok.Pos, curTok.Length()).
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
		return builder.Finish(stmt).(*ast.ExpressionStatement)
	}

	// End at expression
	return builder.FinishWithNode(stmt, left).(*ast.ExpressionStatement)
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
	builder := p.StartNode()

	startToken := p.cursor.Current()
	stmt := &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{Token: startToken},
	}

	// Parse expression using cursor mode (expressions are fully migrated)
	stmt.Expression = p.parseExpressionCursor(LOWEST)

	// Optional semicolon
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.SEMICOLON {
		p.cursor = p.cursor.Advance()
		return builder.FinishWithToken(stmt, p.cursor.Current()).(*ast.ExpressionStatement)
	}

	// End at expression or current token
	if stmt.Expression != nil {
		return builder.FinishWithNode(stmt, stmt.Expression).(*ast.ExpressionStatement)
	}
	return builder.FinishWithToken(stmt, startToken).(*ast.ExpressionStatement)
}

// parseAssignmentOrExpressionCursor determines if we have an assignment or expression statement.
// Task 2.2.14.2: Assignment statement migration
// This handles both simple assignments (x := value), compound assignments (x += value),
// and member assignments (obj.field := value, arr[i] := value).
// PRE: cursor is on first token (typically an identifier or expression start)
// POST: cursor is on last token of statement (possibly SEMICOLON)
func (p *Parser) parseAssignmentOrExpressionCursor() ast.Statement {
	builder := p.StartNode()

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

			// Optional semicolon
			nextToken := p.cursor.Peek(1)
			if nextToken.Type == lexer.SEMICOLON {
				p.cursor = p.cursor.Advance()
				return builder.FinishWithToken(stmt, p.cursor.Current()).(*ast.AssignmentStatement)
			}

			// End at value expression or current token
			if stmt.Value != nil {
				return builder.FinishWithNode(stmt, stmt.Value).(*ast.AssignmentStatement)
			}
			return builder.FinishWithToken(stmt, p.cursor.Current()).(*ast.AssignmentStatement)

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

	// Optional semicolon
	nextToken = p.cursor.Peek(1)
	if nextToken.Type == lexer.SEMICOLON {
		p.cursor = p.cursor.Advance()
		return builder.FinishWithToken(stmt, p.cursor.Current()).(*ast.ExpressionStatement)
	}

	// End at expression or start token
	if left != nil {
		return builder.FinishWithNode(stmt, left).(*ast.ExpressionStatement)
	}
	return builder.FinishWithToken(stmt, startToken).(*ast.ExpressionStatement)
}

// ============================================================================
// Task 2.2.14.3: Cursor-mode handler for block statements
// ============================================================================

// parseBlockStatementCursor parses a begin...end block in cursor mode.
// Task 2.2.14.3: Block statement migration
// PRE: cursor is on BEGIN token
// POST: cursor is on END token
func (p *Parser) parseBlockStatementCursor() *ast.BlockStatement {
	builder := p.StartNode()

	beginToken := p.cursor.Current()
	block := &ast.BlockStatement{
		BaseNode: ast.BaseNode{Token: beginToken},
	}
	block.Statements = []ast.Statement{}

	// Track block context for better error messages
	p.pushBlockContext("begin", beginToken.Pos)
	defer p.popBlockContext()

	// Advance past 'begin'
	p.cursor = p.cursor.Advance()

	// Parse statements until we hit END, EOF, or ENSURE
	for {
		currentToken := p.cursor.Current()

		// Termination conditions
		if currentToken.Type == lexer.END ||
			currentToken.Type == lexer.EOF ||
			currentToken.Type == lexer.ENSURE {
			break
		}

		// Skip semicolons at statement level
		if currentToken.Type == lexer.SEMICOLON {
			p.cursor = p.cursor.Advance()
			continue
		}

		// Parse statement using cursor mode
		// Since we're in cursor mode, parseStatementCursor() will be called
		stmt := p.parseStatementCursor()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		// Advance to next token
		p.cursor = p.cursor.Advance()

		// Skip any semicolons after the statement
		for p.cursor.Current().Type == lexer.SEMICOLON {
			p.cursor = p.cursor.Advance()
		}
	}

	// Check for proper block termination
	currentToken := p.cursor.Current()
	if currentToken.Type != lexer.END && currentToken.Type != lexer.ENSURE {
		p.addErrorWithContext("expected 'end' to close block", ErrMissingEnd)
		// Synchronize to recover
		p.synchronizeCursor([]lexer.TokenType{lexer.END, lexer.ENSURE})
	}

	return builder.FinishWithToken(block, p.cursor.Current()).(*ast.BlockStatement)
}

// ============================================================================
// Task 2.2.14.6: Cursor-mode variable and constant declaration handlers
// ============================================================================

// looksLikeVarDeclarationCursor performs lookahead using cursor to check if
// the next tokens form a var declaration.
// A var declaration pattern is: IDENT followed by either:
// - COLON (for typed declaration: x : Integer)
// - COMMA (for multi-var declaration: x, y : Integer)
// This prevents mis-parsing function calls or assignments as var declarations.
func (p *Parser) looksLikeVarDeclarationCursor(cursor *TokenCursor) bool {
	nextToken := cursor.Peek(1)
	if !p.isIdentifierToken(nextToken.Type) {
		return false
	}

	// Check the token after the identifier
	tokenAfterIdent := cursor.Peek(2)

	// Only accept unambiguous var declaration patterns:
	// - name : Type         (explicit type - always a declaration)
	// - name, name2 : Type  (multi-var declaration)
	return tokenAfterIdent.Type == lexer.COLON ||
		tokenAfterIdent.Type == lexer.COMMA
}

// looksLikeConstDeclarationCursor performs lookahead using cursor to check if
// the next tokens form a const declaration.
// A const declaration pattern is: IDENT followed by either:
// - COLON (for typed const: C : Integer = 5)
// - EQ (for untyped const: C = value)
// This prevents mis-parsing other statements as const declarations.
func (p *Parser) looksLikeConstDeclarationCursor(cursor *TokenCursor) bool {
	nextToken := cursor.Peek(1)
	if !p.isIdentifierToken(nextToken.Type) {
		return false
	}

	// Check the token after the identifier
	tokenAfterIdent := cursor.Peek(2)

	// Const declaration patterns:
	// - NAME : Type = value  (typed const)
	// - NAME = value         (untyped const)
	return tokenAfterIdent.Type == lexer.COLON ||
		tokenAfterIdent.Type == lexer.EQ
}

// parseVarDeclarationCursor parses one or more variable declarations in a var block.
// Task 2.2.14.6: Variable declaration migration
// Syntax: var NAME : TYPE; or var NAME := VALUE; or var NAME : TYPE := VALUE;
// DWScript allows block syntax: var V1 : Integer; V2 : String; (one var keyword, multiple declarations)
// This function returns a BlockStatement containing all var declarations in the block.
// PRE: cursor is on VAR token
// POST: cursor is on last token of last declaration
func (p *Parser) parseVarDeclarationCursor() ast.Statement {
	blockToken := p.cursor.Current() // Save the initial VAR token for the block
	statements := []ast.Statement{}

	// Parse first var declaration
	firstStmt := p.parseSingleVarDeclarationCursor()
	if firstStmt == nil {
		return nil
	}
	statements = append(statements, firstStmt)

	// Continue parsing additional var declarations without the 'var' keyword
	// As long as the next line looks like a var declaration (not just any identifier)
	for p.looksLikeVarDeclarationCursor(p.cursor) {
		p.cursor = p.cursor.Advance() // move to identifier
		varStmt := p.parseSingleVarDeclarationCursor()
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

// parseSingleVarDeclarationCursor parses a single variable declaration using cursor mode.
// Task 2.2.14.6: Variable declaration migration
// Assumes we're already positioned at the identifier (or just before it).
// PRE: cursor is on VAR or variable name IDENT
// POST: cursor is on SEMICOLON
func (p *Parser) parseSingleVarDeclarationCursor() *ast.VarDeclStatement {
	builder := p.StartNode()
	stmt := &ast.VarDeclStatement{}

	// Check if we're already at the identifier (var section continuation)
	// or if we need to advance to it (after 'var' keyword)
	currentToken := p.cursor.Current()
	if currentToken.Type == lexer.VAR {
		stmt.Token = currentToken
		// After 'var' keyword, expect identifier next
		p.cursor = p.cursor.Advance()
		currentToken = p.cursor.Current()
		if !p.isIdentifierToken(currentToken.Type) {
			// Use structured error (Task 2.1.3)
			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrExpectedIdent).
				WithMessage("expected identifier in var declaration").
				WithPosition(currentToken.Pos, currentToken.Length()).
				WithExpectedString("variable name").
				WithActual(currentToken.Type, currentToken.Literal).
				WithSuggestion("provide a variable name after 'var'").
				WithParsePhase("variable declaration").
				Build()
			p.addStructuredError(err)
			return nil
		}
	} else if !p.isIdentifierToken(currentToken.Type) {
		// Should already be at an identifier
		// Use structured error (Task 2.1.3)
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrExpectedIdent).
			WithMessage("expected identifier in var declaration").
			WithPosition(currentToken.Pos, currentToken.Length()).
			WithExpectedString("variable name").
			WithActual(currentToken.Type, currentToken.Literal).
			WithSuggestion("provide a variable name after 'var'").
			WithParsePhase("variable declaration").
			Build()
		p.addStructuredError(err)
		return nil
	} else {
		stmt.Token = currentToken
	}

	// Collect comma-separated identifiers
	// Parse pattern: IDENT (, IDENT)* : TYPE [:= VALUE]
	stmt.Names = []*ast.Identifier{}
	for {
		currentToken = p.cursor.Current()
		if !p.isIdentifierToken(currentToken.Type) {
			// Use structured error (Task 2.1.3)
			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrExpectedIdent).
				WithMessage("expected identifier in var declaration").
				WithPosition(currentToken.Pos, currentToken.Length()).
				WithExpectedString("variable name").
				WithActual(currentToken.Type, currentToken.Literal).
				WithSuggestion("provide a variable name").
				WithParsePhase("variable declaration").
				Build()
			p.addStructuredError(err)
			return nil
		}

		stmt.Names = append(stmt.Names, &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: currentToken,
				},
			},
			Value: currentToken.Literal,
		})

		// Check if there are more names (comma-separated)
		nextToken := p.cursor.Peek(1)
		if nextToken.Type == lexer.COMMA {
			p.cursor = p.cursor.Advance() // move to ','
			p.cursor = p.cursor.Advance() // move to next identifier
			currentToken = p.cursor.Current()
			if !p.isIdentifierToken(currentToken.Type) {
				// Use structured error
				err := NewStructuredError(ErrKindMissing).
					WithCode(ErrExpectedIdent).
					WithMessage("expected identifier after comma in var declaration").
					WithPosition(currentToken.Pos, currentToken.Length()).
					WithExpectedString("variable name").
					WithActual(currentToken.Type, currentToken.Literal).
					WithSuggestion("provide a variable name after ','").
					WithParsePhase("variable declaration").
					Build()
				p.addStructuredError(err)
				return nil
			}
			continue
		}

		// No more names, break to parse type
		break
	}

	// Parse optional type annotation
	// Note: Not using OptionalTypeAnnotation() combinator here because cursor mode
	// requires careful state management and the combinator is designed for traditional mode.
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.COLON {
		p.cursor = p.cursor.Advance() // move to ':'

		// Parse type expression (can be simple type, function pointer, or array type)
		p.cursor = p.cursor.Advance() // move to type expression

		typeExpr := p.parseTypeExpressionCursor()

		if typeExpr == nil {
			// Use structured error (Task 2.1.3)
			currentToken = p.cursor.Current()
			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrExpectedType).
				WithMessage("expected type expression after ':' in var declaration").
				WithPosition(currentToken.Pos, currentToken.Length()).
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

	nextToken = p.cursor.Peek(1)
	if stmt.Type != nil {
		if nextToken.Type == lexer.ASSIGN || nextToken.Type == lexer.EQ {
			if len(stmt.Names) > 1 {
				// Use structured error (Task 2.1.3)
				err := NewStructuredError(ErrKindInvalid).
					WithCode(ErrInvalidSyntax).
					WithMessage("cannot use initializer with multiple variable names").
					WithPosition(nextToken.Pos, nextToken.Length()).
					WithSuggestion("declare variables separately or use the same value for all").
					WithNote("DWScript requires: var x, y: Integer (no initializer) or var x: Integer := 10").
					WithParsePhase("variable declaration").
					Build()
				p.addStructuredError(err)
				return stmt
			}

			p.cursor = p.cursor.Advance() // move to assignment operator
			p.cursor = p.cursor.Advance() // move to value expression
			stmt.Value = p.parseExpressionCursor(ASSIGN)
		}
	} else {
		if nextToken.Type == lexer.ASSIGN || nextToken.Type == lexer.EQ {
			if len(stmt.Names) > 1 {
				// Use structured error (Task 2.1.3)
				err := NewStructuredError(ErrKindInvalid).
					WithCode(ErrInvalidSyntax).
					WithMessage("cannot use initializer with multiple variable names").
					WithPosition(nextToken.Pos, nextToken.Length()).
					WithSuggestion("declare variables separately when using initializers").
					WithNote("DWScript requires separate declarations with initializers").
					WithParsePhase("variable declaration").
					Build()
				p.addStructuredError(err)
				return stmt
			}

			p.cursor = p.cursor.Advance() // move to assignment operator
			stmt.Inferred = true
			p.cursor = p.cursor.Advance() // move to value expression
			stmt.Value = p.parseExpressionCursor(ASSIGN)
		} else {
			nextToken = p.cursor.Peek(1)
			if nextToken.Type == lexer.SEMICOLON || nextToken.Type == lexer.EXTERNAL {
				// Use structured error (Task 2.1.3)
				currentToken = p.cursor.Current()
				err := NewStructuredError(ErrKindInvalid).
					WithCode(ErrInvalidSyntax).
					WithMessage("variable declaration requires a type or initializer").
					WithPosition(currentToken.Pos, currentToken.Length()).
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
					WithPosition(nextToken.Pos, nextToken.Length()).
					WithExpectedString("':' or ':='").
					WithActual(nextToken.Type, nextToken.Literal).
					WithSuggestion("add ':' for type declaration or ':=' for type inference").
					WithParsePhase("variable declaration").
					Build()
				p.addStructuredError(err)
			}
		}
	}

	// Check for 'external' keyword
	nextToken = p.cursor.Peek(1)
	if nextToken.Type == lexer.EXTERNAL {
		p.cursor = p.cursor.Advance() // move to 'external'
		stmt.IsExternal = true

		// Check for optional external name: external 'customName'
		nextToken = p.cursor.Peek(1)
		if nextToken.Type == lexer.STRING {
			p.cursor = p.cursor.Advance() // move to string literal
			stmt.ExternalName = p.cursor.Current().Literal
		}
	}

	// Expect semicolon
	nextToken = p.cursor.Peek(1)
	if nextToken.Type != lexer.SEMICOLON {
		// Use structured error
		currentToken = p.cursor.Current()
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingSemicolon).
			WithMessage("expected ';' after variable declaration").
			WithPosition(currentToken.Pos, currentToken.Length()).
			WithExpectedString("';'").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("add ';' at the end of the declaration").
			WithParsePhase("variable declaration").
			Build()
		p.addStructuredError(err)
		return stmt
	}

	p.cursor = p.cursor.Advance() // move to semicolon

	return builder.FinishWithToken(stmt, p.cursor.Current()).(*ast.VarDeclStatement)
}
