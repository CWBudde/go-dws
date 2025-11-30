package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// PRE: cursor is on first token of statement
// POST: cursor is on last token of statement
// parseDefaultStatementCase handles the complex default case logic for statement parsing.
func (p *Parser) parseDefaultStatementCase(currentToken lexer.Token) ast.Statement {
	// Check for assignment (simple or member assignment)
	// Handle SELF and INHERITED which can be assignment targets (e.g., Self.X := value)
	if currentToken.Type == lexer.SELF || currentToken.Type == lexer.INHERITED {
		return p.parseAssignmentOrExpression()
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
			return p.parseVarDeclaration()
		}

		return p.parseAssignmentOrExpression()
	}

	return p.parseExpressionStatement()
}

func (p *Parser) parseStatement() ast.Statement {
	// As we implement each statement cursor handler, they'll be added here

	currentToken := p.cursor.Current()

	switch currentToken.Type {
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
		nextToken := p.cursor.Peek(1)
		if nextToken.Type == lexer.FUNCTION || nextToken.Type == lexer.PROCEDURE || nextToken.Type == lexer.METHOD {
			p.cursor = p.cursor.Advance() // move to function/procedure/method token
			fn := p.parseFunctionDeclaration()
			if fn != nil {
				fn.IsClassMethod = true
			}
			return fn
		}
		p.addError("expected 'function', 'procedure', or 'method' after 'class'", ErrUnexpectedToken)
		return nil

	case lexer.CONSTRUCTOR:
		method := p.parseFunctionDeclaration()
		if method != nil {
			method.IsConstructor = true
		}
		return method

	case lexer.DESTRUCTOR:
		method := p.parseFunctionDeclaration()
		if method != nil {
			method.IsDestructor = true
		}
		return method

	case lexer.TYPE:
		return p.parseTypeDeclaration()

	case lexer.USES:
		return p.parseUsesClause()

	default:
		return p.parseDefaultStatementCase(currentToken)
	}
}

// parseBlockStatement parses a begin...end block.
// PRE: cursor is BEGIN
// POST: cursor is END

// parseExpressionStatement parses an expression statement.
// PRE: cursor is first token of expression
// POST: cursor is SEMICOLON if present, otherwise last token of expression

// parseVarDeclaration parses a variable declaration statement.
// Can be called in two contexts:
//  1. After 'var' keyword: var x: Integer;
//  2. In a var section without 'var': x: Integer; (curToken is already the IDENT)
//
// PRE: cursor is VAR or IDENT
// POST: cursor is SEMICOLON of last var declaration

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

// looksLikeVarDeclaration performs lookahead to check if the next tokens form a var declaration.
// A var declaration pattern is: IDENT followed by either:
// - COLON (for typed declaration: x : Integer)
// - COMMA (for multi-var declaration: x, y : Integer)
// - ASSIGN/EQ when inference is explicitly allowed (block-style var sections)
// This prevents mis-parsing function calls or assignments as var declarations.

// looksLikeConstDeclaration performs lookahead to check if the next tokens form a const declaration.
// A const declaration pattern is: IDENT followed by either:
// - COLON (for typed const: C : Integer = 5)
// - EQ (for untyped const: C = value)
// This prevents mis-parsing other statements as const declarations.
// NOTE: We accept EQ but not ASSIGN here. While both can be used for consts,
//
//	ASSIGN is more commonly used for var declarations and assignments.

// ============================================================================
// ============================================================================

// PRE: cursor is on first token of expression
// POST: cursor is on last token of statement (possibly SEMICOLON)
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	builder := p.StartNode()

	startToken := p.cursor.Current()
	stmt := &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{Token: startToken},
	}

	stmt.Expression = p.parseExpression(LOWEST)

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

// parseAssignmentOrExpression determines if we have an assignment or expression statement.
// This handles both simple assignments (x := value), compound assignments (x += value),
// and member assignments (obj.field := value, arr[i] := value).
// PRE: cursor is on first token (typically an identifier or expression start)
// POST: cursor is on last token of statement (possibly SEMICOLON)
func (p *Parser) parseAssignmentOrExpression() ast.Statement {
	builder := p.StartNode()

	startToken := p.cursor.Current()

	// Parse the left side as an expression (could be identifier, member access, or index)
	left := p.parseExpression(LOWEST)

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
			stmt.Value = p.parseExpression(ASSIGN)

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
// ============================================================================

// PRE: cursor is on BEGIN token
// POST: cursor is on END token
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
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

		stmt := p.parseStatement()
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
		p.synchronize([]lexer.TokenType{lexer.END, lexer.ENSURE})
	}

	return builder.FinishWithToken(block, p.cursor.Current()).(*ast.BlockStatement)
}

// ============================================================================
// ============================================================================

// looksLikeVarDeclaration performs lookahead using cursor to check if
// the next tokens form a var declaration.
// A var declaration pattern is: IDENT followed by either:
// - COLON (for typed declaration: x : Integer)
// - COMMA (for multi-var declaration: x, y : Integer)
// This prevents mis-parsing function calls or assignments as var declarations.
func (p *Parser) looksLikeVarDeclaration(cursor *TokenCursor, allowInferred bool) bool {
	nextToken := cursor.Peek(1)
	if !p.isIdentifierToken(nextToken.Type) {
		return false
	}

	// Check the token after the identifier
	tokenAfterIdent := cursor.Peek(2)

	// Only accept unambiguous var declaration patterns:
	// - name : Type         (explicit type - always a declaration)
	// - name, name2 : Type  (multi-var declaration)
	if tokenAfterIdent.Type == lexer.COLON || tokenAfterIdent.Type == lexer.COMMA {
		return true
	}

	// Allow inferred declarations when we're in a block-style var section
	if allowInferred && (tokenAfterIdent.Type == lexer.ASSIGN || tokenAfterIdent.Type == lexer.EQ) {
		return true
	}

	return false
}

// looksLikeConstDeclaration performs lookahead using cursor to check if
// the next tokens form a const declaration.
// A const declaration pattern is: IDENT followed by either:
// - COLON (for typed const: C : Integer = 5)
// - EQ (for untyped const: C = value)
// This prevents mis-parsing other statements as const declarations.
func (p *Parser) looksLikeConstDeclaration(cursor *TokenCursor) bool {
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

// parseVarDeclaration parses one or more variable declarations in a var block.
// Syntax: var NAME : TYPE; or var NAME := VALUE; or var NAME : TYPE := VALUE;
// DWScript allows block syntax: var V1 : Integer; V2 : String; (one var keyword, multiple declarations)
// This function returns a BlockStatement containing all var declarations in the block.
// PRE: cursor is on VAR token
// POST: cursor is on last token of last declaration
func (p *Parser) parseVarDeclaration() ast.Statement {
	blockToken := p.cursor.Current() // Save the initial VAR token for the block
	statements := []ast.Statement{}

	// Detect block-style var sections where the 'var' keyword is on its own line.
	// In that case, allow inferred declarations (:=) in continuation lines.
	allowInferredContinuation := false
	if blockToken.Type == lexer.VAR {
		nextToken := p.cursor.Peek(1)
		allowInferredContinuation = nextToken.Pos.Line > blockToken.Pos.Line
	}

	// Parse first var declaration
	firstStmt := p.parseSingleVarDeclaration()
	if firstStmt == nil {
		return nil
	}
	statements = append(statements, firstStmt)

	// Continue parsing additional var declarations without the 'var' keyword
	// As long as the next line looks like a var declaration (not just any identifier)
	for p.looksLikeVarDeclaration(p.cursor, allowInferredContinuation) {
		p.cursor = p.cursor.Advance() // move to identifier
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

// Assumes we're already positioned at the identifier (or just before it).
// PRE: cursor is on VAR or variable name IDENT
// POST: cursor is on SEMICOLON
func (p *Parser) parseSingleVarDeclaration() *ast.VarDeclStatement {
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
			// Use structured error
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
		// Use structured error
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
			// Use structured error
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
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.COLON {
		p.cursor = p.cursor.Advance() // move to ':'

		// Parse type expression (can be simple type, function pointer, or array type)
		p.cursor = p.cursor.Advance() // move to type expression

		typeExpr := p.parseTypeExpression()

		if typeExpr == nil {
			// Use structured error
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
				// Use structured error
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
			stmt.Value = p.parseExpression(ASSIGN)
		}
	} else {
		if nextToken.Type == lexer.ASSIGN || nextToken.Type == lexer.EQ {
			if len(stmt.Names) > 1 {
				// Use structured error
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
			stmt.Value = p.parseExpression(ASSIGN)
		} else {
			nextToken = p.cursor.Peek(1)
			if nextToken.Type == lexer.SEMICOLON || nextToken.Type == lexer.EXTERNAL {
				// Use structured error
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
				// Use structured error
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
