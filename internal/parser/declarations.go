package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseConstDeclaration parses one or more constant declarations in a const block.
// Syntax: const NAME = VALUE; or const NAME := VALUE; or const NAME: TYPE = VALUE;
// DWScript allows block syntax: const C1 = 1; C2 = 2; (one const keyword, multiple declarations)
// This function returns a BlockStatement containing all const declarations in the block.
// PRE: curToken is CONST
// POST: curToken is last token of value expression in last declaration
// Dispatcher: delegates to cursor or traditional mode
func (p *Parser) parseConstDeclaration() ast.Statement {
	return p.parseConstDeclarationCursor()
}

// parseSingleConstDeclaration parses a single constant declaration.
// Assumes we're already positioned at the identifier (or just before it).
// PRE: curToken is CONST or IDENT
// POST: curToken is last token of value expression
func (p *Parser) parseSingleConstDeclaration() *ast.ConstDecl {
	builder := p.StartNode()

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

	// Use OptionalTypeAnnotation combinator (Task 2.3.3)
	stmt.Type = p.OptionalTypeAnnotation()

	// Use Choice combinator for '=' or ':=' (Task 2.3.3)
	if !p.Choice(lexer.EQ, lexer.ASSIGN) {
		p.addError("expected '=' or ':=' after const name", ErrMissingAssign)
		return stmt
	}

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
		return builder.Finish(stmt).(*ast.ConstDecl)
	} else if stmt.Value != nil {
		// No semicolon - end position is after the value expression
		return builder.FinishWithNode(stmt, stmt.Value).(*ast.ConstDecl)
	} else {
		// Fallback - use current token
		return builder.Finish(stmt).(*ast.ConstDecl)
	}
}

// parseProgramDeclaration parses an optional program declaration at the start of a file.
// Syntax: program ProgramName;
// The program declaration is optional in DWScript and doesn't affect execution.
// It is parsed and then discarded (not added to the AST).
// PRE: curToken is PROGRAM
// POST: curToken is SEMICOLON
// Dispatcher: delegates to cursor or traditional mode
func (p *Parser) parseProgramDeclaration() {
	p.parseProgramDeclarationCursor()
}

// parseProgramDeclarationCursor parses program declaration using cursor mode.
// PRE: cursor is on PROGRAM token
// POST: cursor is on SEMICOLON token
func (p *Parser) parseProgramDeclarationCursor() {
	// We're on the PROGRAM token
	currentToken := p.cursor.Current()
	if currentToken.Type != lexer.PROGRAM {
		return
	}

	// Expect identifier (program name)
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.IDENT {
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrExpectedIdent).
			WithMessage("expected program name after 'program' keyword").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpectedString("program name").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("provide a program name after 'program'").
			WithParsePhase("program declaration").
			Build()
		p.addStructuredError(err)
		return
	}
	p.cursor = p.cursor.Advance() // move to program name

	// Note: We could store the program name if needed, but DWScript ignores it
	// programName := p.cursor.Current().Literal

	// Expect semicolon
	nextToken = p.cursor.Peek(1)
	if nextToken.Type != lexer.SEMICOLON {
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingSemicolon).
			WithMessage("expected ';' after program name").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpectedString("';'").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("add ';' after program name").
			WithParsePhase("program declaration").
			Build()
		p.addStructuredError(err)
		return
	}
	p.cursor = p.cursor.Advance() // move to semicolon

	// Successfully parsed program declaration
	// The program name is not stored in the AST as it doesn't affect execution
}

// ============================================================================
// Task 2.2.14.6: Cursor-mode constant declaration handlers
// ============================================================================

// parseConstDeclarationCursor parses one or more constant declarations in a const block using cursor mode.
// Task 2.2.14.6: Constant declaration migration
// Syntax: const NAME = VALUE; or const NAME := VALUE; or const NAME: TYPE = VALUE;
// DWScript allows block syntax: const C1 = 1; C2 = 2; (one const keyword, multiple declarations)
// This function returns a BlockStatement containing all const declarations in the block.
// PRE: cursor is on CONST token
// POST: cursor is on last token of last declaration
func (p *Parser) parseConstDeclarationCursor() ast.Statement {
	blockToken := p.cursor.Current() // Save the initial CONST token for the block
	statements := []ast.Statement{}

	// Parse first const declaration
	firstStmt := p.parseSingleConstDeclarationCursor()
	if firstStmt == nil {
		return nil
	}
	statements = append(statements, firstStmt)

	// Continue parsing additional const declarations without the 'const' keyword
	// As long as the next line looks like a const declaration (not just any identifier)
	for p.looksLikeConstDeclarationCursor(p.cursor) {
		p.cursor = p.cursor.Advance() // move to identifier
		constStmt := p.parseSingleConstDeclarationCursor()
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
		BaseNode:   ast.BaseNode{Token: blockToken},
		Statements: statements,
	}
}

// parseSingleConstDeclarationCursor parses a single constant declaration using cursor mode.
// Task 2.2.14.6: Constant declaration migration
// Assumes we're already positioned at the identifier (or just before it).
// PRE: cursor is on CONST or IDENT token
// POST: cursor is on last token of value expression or SEMICOLON
func (p *Parser) parseSingleConstDeclarationCursor() *ast.ConstDecl {
	builder := p.StartNode()

	// If we're at CONST token, advance to identifier
	currentToken := p.cursor.Current()
	if currentToken.Type == lexer.CONST {
		p.cursor = p.cursor.Advance() // move to identifier
		currentToken = p.cursor.Current()
	}

	// We should now be at the identifier
	if !p.isIdentifierToken(currentToken.Type) {
		// Use structured error
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrExpectedIdent).
			WithMessage("expected identifier in const declaration").
			WithPosition(currentToken.Pos, currentToken.Length()).
			WithExpectedString("constant name").
			WithActual(currentToken.Type, currentToken.Literal).
			WithSuggestion("provide a constant name after 'const'").
			WithParsePhase("constant declaration").
			Build()
		p.addStructuredError(err)
		return nil
	}

	stmt := &ast.ConstDecl{
		BaseNode: ast.BaseNode{
			Token: currentToken,
		},
	}
	stmt.Name = &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: currentToken,
			},
		},
		Value: currentToken.Literal,
	}

	// Check for optional type annotation (: Type)
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.COLON {
		p.cursor = p.cursor.Advance() // move to ':'

		// Parse type expression (can be simple type, function pointer, or array type)
		p.cursor = p.cursor.Advance() // move to type expression

		// Task 2.7.4: Use cursor mode directly
		typeExpr := p.parseTypeExpressionCursor()

		if typeExpr == nil {
			// Use structured error
			currentToken = p.cursor.Current()
			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrExpectedType).
				WithMessage("expected type expression after ':' in const declaration").
				WithPosition(currentToken.Pos, currentToken.Length()).
				WithExpectedString("type name").
				WithSuggestion("specify the constant type, like 'Integer' or 'String'").
				WithParsePhase("constant declaration type").
				Build()
			p.addStructuredError(err)
			return stmt
		}

		// Directly assign the type expression without creating synthetic wrappers
		stmt.Type = typeExpr
	}

	// Expect '=' or ':=' token
	nextToken = p.cursor.Peek(1)
	if nextToken.Type != lexer.EQ && nextToken.Type != lexer.ASSIGN {
		// Use structured error
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingAssign).
			WithMessage("expected '=' or ':=' after const name").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpectedString("'=' or ':='").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("add '=' or ':=' before the constant value").
			WithParsePhase("constant declaration").
			Build()
		p.addStructuredError(err)
		return stmt
	}
	p.cursor = p.cursor.Advance() // move to '=' or ':='

	// Parse value expression
	p.cursor = p.cursor.Advance() // move to value expression
	stmt.Value = p.parseExpressionCursor(ASSIGN)

	// Check for optional 'deprecated' keyword
	nextToken = p.cursor.Peek(1)
	if nextToken.Type == lexer.DEPRECATED {
		p.cursor = p.cursor.Advance() // move to 'deprecated'
		stmt.IsDeprecated = true

		// Check for optional deprecation message string
		nextToken = p.cursor.Peek(1)
		if nextToken.Type == lexer.STRING {
			p.cursor = p.cursor.Advance() // move to string
			stmt.DeprecatedMessage = p.cursor.Current().Literal
		}
	}

	// Expect semicolon
	nextToken = p.cursor.Peek(1)
	if nextToken.Type == lexer.SEMICOLON {
		p.cursor = p.cursor.Advance() // move to semicolon
		// End position is at the semicolon
		return builder.Finish(stmt).(*ast.ConstDecl)
	} else if stmt.Value != nil {
		// No semicolon - end position is after the value expression
		return builder.FinishWithNode(stmt, stmt.Value).(*ast.ConstDecl)
	} else {
		// Fallback - use current token
		return builder.Finish(stmt).(*ast.ConstDecl)
	}
}
