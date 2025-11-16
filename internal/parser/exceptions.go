package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseRaiseStatement parses a raise statement.
// Syntax:
//   - raise <expression>;  (raise new exception)
//   - raise;               (re-raise current exception, only valid in except block)
//
// Examples:
//
//	raise Exception.Create('error');
//	raise new EMyException('custom error');
//	raise;  // re-raise
//
// PRE: curToken is RAISE
// POST: curToken is last token of exception expression (or RAISE for bare raise)
func (p *Parser) parseRaiseStatement() *ast.RaiseStatement {
	stmt := &ast.RaiseStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}

	// Check if this is a bare raise (no expression)
	if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.EOF) {
		// Bare raise - re-raise current exception
		stmt.EndPos = p.endPosFromToken(p.curToken)
		return stmt
	}

	// Parse the exception expression
	p.nextToken()
	stmt.Exception = p.parseExpression(LOWEST)

	if stmt.Exception == nil {
		p.addError("expected exception expression after 'raise'", ErrInvalidExpression)
		return nil
	}

	// End position is after the exception expression
	stmt.EndPos = stmt.Exception.End()

	return stmt
}

// parseTryStatement parses a try...except...finally...end statement.
// Syntax:
//   - try <statements> except <handlers> end;
//   - try <statements> finally <statements> end;
//   - try <statements> except <handlers> finally <statements> end;
//
// Examples:
//
//	try
//	  DoSomething();
//	except
//	  on E: Exception do
//	    PrintLn(E.Message);
//	end;
//
//	try
//	  DoSomething();
//	finally
//	  Cleanup();
//	end;
//
// PRE: curToken is TRY
// POST: curToken is END
func (p *Parser) parseTryStatement() *ast.TryStatement {
	stmt := &ast.TryStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}

	// Parse try block
	p.nextToken()
	stmt.TryBlock = p.parseBlockStatementForTry()

	if stmt.TryBlock == nil {
		p.addError("expected statements after 'try'", ErrInvalidSyntax)
		return nil
	}

	// Check for except clause
	if p.curTokenIs(lexer.EXCEPT) {
		stmt.ExceptClause = p.parseExceptClause()
		if stmt.ExceptClause == nil {
			return nil
		}
	}

	// Check for finally clause
	if p.curTokenIs(lexer.FINALLY) {
		stmt.FinallyClause = p.parseFinallyClause()
		if stmt.FinallyClause == nil {
			return nil
		}
	}

	// Validate that at least one of except or finally is present
	if stmt.ExceptClause == nil && stmt.FinallyClause == nil {
		p.addError("expected 'except' or 'finally' after 'try' block", ErrUnexpectedToken)
		return nil
	}

	// Expect 'end' keyword
	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close try statement", ErrMissingEnd)
		return nil
	}

	// End position is at the 'end' keyword
	stmt.EndPos = p.endPosFromToken(p.curToken)

	return stmt
}

// parseBlockStatementForTry parses statements until 'except', 'finally', or 'end'
// PRE: curToken is first statement token
// POST: curToken is EXCEPT, FINALLY, or END
func (p *Parser) parseBlockStatementForTry() *ast.BlockStatement {
	block := &ast.BlockStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}
	block.Statements = []ast.Statement{}

	for !p.curTokenIs(lexer.EXCEPT) && !p.curTokenIs(lexer.FINALLY) &&
		!p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {

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

	return block
}

// parseExceptClause parses an except clause.
// Syntax:
//   - except <handlers> end
//   - except <handlers> else <statements> end
//
// Examples:
//
//	except
//	  on E: Exception do
//	    PrintLn(E.Message);
//	end
//
//	except
//	  PrintLn('error');  // bare except
//	end
//
// PRE: curToken is EXCEPT
// POST: curToken is last token before FINALLY or END
func (p *Parser) parseExceptClause() *ast.ExceptClause {
	exceptToken := p.curToken // Save 'except' token before moving past it
	clause := &ast.ExceptClause{
		Token: exceptToken, // 'except' keyword token
	}
	clause.Handlers = []*ast.ExceptionHandler{}

	p.nextToken() // move past 'except'

	// Parse exception handlers or bare except
	for p.curTokenIs(lexer.ON) {
		handler := p.parseExceptionHandler()
		if handler == nil {
			return nil
		}
		clause.Handlers = append(clause.Handlers, handler)

		// Skip semicolons between handlers
		for p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}
	}

	// If no handlers, this is a bare except - parse statements until finally or end
	if len(clause.Handlers) == 0 {
		// Parse bare except statements into a block
		bareBlock := &ast.BlockStatement{
			BaseNode: ast.BaseNode{Token: p.curToken},
		}
		bareBlock.Statements = []ast.Statement{}

		for !p.curTokenIs(lexer.FINALLY) && !p.curTokenIs(lexer.END) &&
			!p.curTokenIs(lexer.ELSE) && !p.curTokenIs(lexer.EOF) {

			if p.curTokenIs(lexer.SEMICOLON) {
				p.nextToken()
				continue
			}

			stmt := p.parseStatement()
			if stmt != nil {
				bareBlock.Statements = append(bareBlock.Statements, stmt)
			}
			p.nextToken()

			for p.curTokenIs(lexer.SEMICOLON) {
				p.nextToken()
			}
		}

		// Create a synthetic handler for bare except (catches all)
		// Handler with nil Variable and nil ExceptionType catches everything
		if len(bareBlock.Statements) > 0 {
			bareHandler := &ast.ExceptionHandler{
				Token:         exceptToken, // Use the 'except' token for synthetic handler
				Variable:      nil,         // No exception variable
				ExceptionType: nil,         // Catches all exception types
				Statement:     bareBlock,
			}
			if bareBlock != nil {
				bareHandler.EndPos = bareBlock.End()
			}
			clause.Handlers = append(clause.Handlers, bareHandler)
		}
	}

	// Check for optional else block
	if p.curTokenIs(lexer.ELSE) {
		p.nextToken()
		// Parse else block statements until finally or end
		elseBlock := &ast.BlockStatement{
			BaseNode: ast.BaseNode{Token: p.curToken},
		}
		elseBlock.Statements = []ast.Statement{}

		for !p.curTokenIs(lexer.FINALLY) && !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
			if p.curTokenIs(lexer.SEMICOLON) {
				p.nextToken()
				continue
			}

			stmt := p.parseStatement()
			if stmt != nil {
				elseBlock.Statements = append(elseBlock.Statements, stmt)
			}

			p.nextToken()

			for p.curTokenIs(lexer.SEMICOLON) {
				p.nextToken()
			}
		}

		clause.ElseBlock = elseBlock
	}

	// Set EndPos based on what was parsed last
	if clause.ElseBlock != nil {
		clause.EndPos = clause.ElseBlock.End()
	} else if len(clause.Handlers) > 0 {
		lastHandler := clause.Handlers[len(clause.Handlers)-1]
		clause.EndPos = lastHandler.End()
	} else {
		// No handlers or else block - use current position
		clause.EndPos = p.endPosFromToken(p.curToken)
	}

	return clause
}

// parseExceptionHandler parses an exception handler.
// Syntax: on <variable>: <type> do <statement>
//
// Examples:
//
//	on E: Exception do
//	  PrintLn(E.Message);
//
//	on E: EMyException do begin
//	  HandleMyException(E);
//	end;
//
// PRE: curToken is ON
// POST: curToken is last token after statement
func (p *Parser) parseExceptionHandler() *ast.ExceptionHandler {
	onToken := p.curToken // Save 'on' token before moving past it
	handler := &ast.ExceptionHandler{
		Token: onToken, // 'on' keyword token
	}

	// Expect 'on' keyword (already checked by caller)
	p.nextToken() // move past 'on'

	// Parse variable name
	if !p.curTokenIs(lexer.IDENT) {
		p.addError("expected identifier after 'on'", ErrExpectedIdent)
		return nil
	}

	handler.Variable = &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

	// Expect ':' token
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	// Parse exception type
	p.nextToken()
	if !p.curTokenIs(lexer.IDENT) {
		p.addError("expected exception type after ':", ErrExpectedType)
		return nil
	}

	handler.ExceptionType = &ast.TypeAnnotation{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	// Expect 'do' keyword
	if !p.expectPeek(lexer.DO) {
		return nil
	}

	// Parse handler statement
	p.nextToken()
	handler.Statement = p.parseStatement()

	if handler.Statement == nil {
		p.addError("expected statement after 'do'", ErrInvalidSyntax)
		return nil
	}

	// Set EndPos to the end of the statement
	if handler.Statement != nil {
		handler.EndPos = handler.Statement.End()
	}

	// After parsing the statement, advance to the next token
	// This handles both block statements (begin...end) and single statements
	p.nextToken()

	return handler
}

// parseFinallyClause parses a finally clause.
// Syntax: finally <statements> end
//
// Example:
//
//	finally
//	  Cleanup();
//	end
//
// PRE: curToken is FINALLY
// POST: curToken is END (before END of try)
func (p *Parser) parseFinallyClause() *ast.FinallyClause {
	clause := &ast.FinallyClause{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}

	p.nextToken() // move past 'finally'

	// Parse finally block statements until 'end'
	block := &ast.BlockStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}
	block.Statements = []ast.Statement{}

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

	clause.Block = block

	return clause
}
