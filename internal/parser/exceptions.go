package parser

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
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
func (p *Parser) parseRaiseStatement() *ast.RaiseStatement {
	stmt := &ast.RaiseStatement{Token: p.curToken}

	// Check if this is a bare raise (no expression)
	if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.EOF) {
		// Bare raise - re-raise current exception
		return stmt
	}

	// Parse the exception expression
	p.nextToken()
	stmt.Exception = p.parseExpression(LOWEST)

	if stmt.Exception == nil {
		p.addError("expected exception expression after 'raise'")
		return nil
	}

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
func (p *Parser) parseTryStatement() *ast.TryStatement {
	stmt := &ast.TryStatement{Token: p.curToken}

	// Parse try block
	p.nextToken()
	stmt.TryBlock = p.parseBlockStatementForTry()

	if stmt.TryBlock == nil {
		p.addError("expected statements after 'try'")
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
		p.addError("expected 'except' or 'finally' after 'try' block")
		return nil
	}

	// Expect 'end' keyword
	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close try statement")
		return nil
	}

	return stmt
}

// parseBlockStatementForTry parses statements until 'except', 'finally', or 'end'
func (p *Parser) parseBlockStatementForTry() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
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
func (p *Parser) parseExceptClause() *ast.ExceptClause {
	clause := &ast.ExceptClause{Token: p.curToken}
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
		bareBlock := &ast.BlockStatement{Token: p.curToken}
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
				Token:         clause.Token,
				Variable:      nil, // No exception variable
				ExceptionType: nil, // Catches all exception types
				Statement:     bareBlock,
			}
			clause.Handlers = append(clause.Handlers, bareHandler)
		}
	}

	// Check for optional else block
	if p.curTokenIs(lexer.ELSE) {
		p.nextToken()
		// Parse else block statements until finally or end
		elseBlock := &ast.BlockStatement{Token: p.curToken}
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
func (p *Parser) parseExceptionHandler() *ast.ExceptionHandler {
	handler := &ast.ExceptionHandler{Token: p.curToken}

	// Expect 'on' keyword (already checked by caller)
	p.nextToken() // move past 'on'

	// Parse variable name
	if !p.curTokenIs(lexer.IDENT) {
		p.addError("expected identifier after 'on'")
		return nil
	}

	handler.Variable = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	// Expect ':' token
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	// Parse exception type
	p.nextToken()
	if !p.curTokenIs(lexer.IDENT) {
		p.addError("expected exception type after ':'")
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
		p.addError("expected statement after 'do'")
		return nil
	}

	// If the statement was a block (begin...end), we're positioned on END
	// Need to advance past it to see what comes next (semicolon, finally, etc.)
	if p.curTokenIs(lexer.END) {
		p.nextToken()
	}

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
func (p *Parser) parseFinallyClause() *ast.FinallyClause {
	clause := &ast.FinallyClause{Token: p.curToken}

	p.nextToken() // move past 'finally'

	// Parse finally block statements until 'end'
	block := &ast.BlockStatement{Token: p.curToken}
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
