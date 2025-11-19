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
	builder := p.StartNode()
	stmt := &ast.RaiseStatement{
		BaseNode: ast.BaseNode{Token: p.cursor.Current()},
	}

	// Check if this is a bare raise (no expression)
	if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.EOF) {
		// Bare raise - re-raise current exception
		return builder.Finish(stmt).(*ast.RaiseStatement)
	}

	// Parse the exception expression
	p.nextToken()
	stmt.Exception = p.parseExpressionCursor(LOWEST)

	if stmt.Exception == nil {
		p.addError("expected exception expression after 'raise'", ErrInvalidExpression)
		return nil
	}

	// End position is after the exception expression
	return builder.FinishWithNode(stmt, stmt.Exception).(*ast.RaiseStatement)
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
	builder := p.StartNode()
	stmt := &ast.TryStatement{
		BaseNode: ast.BaseNode{Token: p.cursor.Current()},
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
	return builder.Finish(stmt).(*ast.TryStatement)
}

// parseBlockStatementForTry parses statements until 'except', 'finally', or 'end'
// PRE: curToken is first statement token
// POST: curToken is EXCEPT, FINALLY, or END
func (p *Parser) parseBlockStatementForTry() *ast.BlockStatement {
	block := &ast.BlockStatement{
		BaseNode: ast.BaseNode{Token: p.cursor.Current()},
	}
	block.Statements = []ast.Statement{}

	for !p.curTokenIs(lexer.EXCEPT) && !p.curTokenIs(lexer.FINALLY) &&
		!p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {

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
	builder := p.StartNode()
	exceptToken := p.cursor.Current() // Save 'except' token before moving past it
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
			BaseNode: ast.BaseNode{Token: p.cursor.Current()},
		}
		bareBlock.Statements = []ast.Statement{}

		for !p.curTokenIs(lexer.FINALLY) && !p.curTokenIs(lexer.END) &&
			!p.curTokenIs(lexer.ELSE) && !p.curTokenIs(lexer.EOF) {

			if p.curTokenIs(lexer.SEMICOLON) {
				p.nextToken()
				continue
			}

			stmt := p.parseStatementCursor()
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
			handlerBuilder := p.StartNode()
			bareHandler := &ast.ExceptionHandler{
				Token:         exceptToken, // Use the 'except' token for synthetic handler
				Variable:      nil,         // No exception variable
				ExceptionType: nil,         // Catches all exception types
				Statement:     bareBlock,
			}
			if bareBlock != nil {
				handlerBuilder.FinishWithNode(bareHandler, bareBlock)
			}
			clause.Handlers = append(clause.Handlers, bareHandler)
		}
	}

	// Check for optional else block
	if p.curTokenIs(lexer.ELSE) {
		p.nextToken()
		// Parse else block statements until finally or end
		elseBlock := &ast.BlockStatement{
			BaseNode: ast.BaseNode{Token: p.cursor.Current()},
		}
		elseBlock.Statements = []ast.Statement{}

		for !p.curTokenIs(lexer.FINALLY) && !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
			if p.curTokenIs(lexer.SEMICOLON) {
				p.nextToken()
				continue
			}

			stmt := p.parseStatementCursor()
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
		return builder.FinishWithNode(clause, clause.ElseBlock).(*ast.ExceptClause)
	} else if len(clause.Handlers) > 0 {
		lastHandler := clause.Handlers[len(clause.Handlers)-1]
		return builder.FinishWithNode(clause, lastHandler).(*ast.ExceptClause)
	} else {
		// No handlers or else block - use current position
		return builder.Finish(clause).(*ast.ExceptClause)
	}
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
	builder := p.StartNode()
	onToken := p.cursor.Current() // Save 'on' token before moving past it
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
				Token: p.cursor.Current(),
			},
		},
		Value: p.cursor.Current().Literal,
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
		Token: p.cursor.Current(),
		Name:  p.cursor.Current().Literal,
	}

	// Expect 'do' keyword
	if !p.expectPeek(lexer.DO) {
		return nil
	}

	// Parse handler statement
	p.nextToken()
	handler.Statement = p.parseStatementCursor()

	if handler.Statement == nil {
		p.addError("expected statement after 'do'", ErrInvalidSyntax)
		return nil
	}

	// After parsing the statement, advance to the next token
	// This handles both block statements (begin...end) and single statements
	p.nextToken()

	// Set EndPos to the end of the statement
	return builder.FinishWithNode(handler, handler.Statement).(*ast.ExceptionHandler)
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
		BaseNode: ast.BaseNode{Token: p.cursor.Current()},
	}

	p.nextToken() // move past 'finally'

	// Parse finally block statements until 'end'
	block := &ast.BlockStatement{
		BaseNode: ast.BaseNode{Token: p.cursor.Current()},
	}
	block.Statements = []ast.Statement{}

	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
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

	clause.Block = block

	return clause
}

// ============================================================================
// Task 2.2.14.7: Cursor-mode exception handling statement handlers
// ============================================================================

// parseRaiseStatementCursor parses a raise statement using cursor mode.
// Task 2.2.14.7: Raise statement migration
// Syntax:
//   - raise <expression>;  (raise new exception)
//   - raise;               (re-raise current exception, only valid in except block)
//
// PRE: cursor is on RAISE token
// POST: cursor is on last token of exception expression (or RAISE for bare raise)
func (p *Parser) parseRaiseStatementCursor() *ast.RaiseStatement {
	builder := p.StartNode()
	raiseToken := p.cursor.Current()
	stmt := &ast.RaiseStatement{
		BaseNode: ast.BaseNode{Token: raiseToken},
	}

	// Check if this is a bare raise (no expression)
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.SEMICOLON || nextToken.Type == lexer.EOF {
		// Bare raise - re-raise current exception
		return builder.FinishWithToken(stmt, raiseToken).(*ast.RaiseStatement)
	}

	// Parse the exception expression
	p.cursor = p.cursor.Advance() // move past 'raise'
	stmt.Exception = p.parseExpressionCursor(LOWEST)

	if stmt.Exception == nil {
		// Use structured error
		currentToken := p.cursor.Current()
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidExpression).
			WithMessage("expected exception expression after 'raise'").
			WithPosition(currentToken.Pos, currentToken.Length()).
			WithSuggestion("provide an exception object to raise").
			WithParsePhase("raise statement").
			Build()
		p.addStructuredError(err)
		return nil
	}

	// End position is after the exception expression
	return builder.FinishWithNode(stmt, stmt.Exception).(*ast.RaiseStatement)
}

// parseTryStatementCursor parses a try...except...finally...end statement using cursor mode.
// Task 2.2.14.7: Try statement migration
// Syntax:
//   - try <statements> except <handlers> end;
//   - try <statements> finally <statements> end;
//   - try <statements> except <handlers> finally <statements> end;
//
// PRE: cursor is on TRY token
// POST: cursor is on END token
func (p *Parser) parseTryStatementCursor() *ast.TryStatement {
	builder := p.StartNode()
	tryToken := p.cursor.Current()
	stmt := &ast.TryStatement{
		BaseNode: ast.BaseNode{Token: tryToken},
	}

	// Track block context for better error messages
	p.pushBlockContext("try", tryToken.Pos)
	defer p.popBlockContext()

	// Parse try block
	p.cursor = p.cursor.Advance() // move past 'try'
	stmt.TryBlock = p.parseBlockStatementForTryCursor()

	if stmt.TryBlock == nil {
		// Use structured error
		currentToken := p.cursor.Current()
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidSyntax).
			WithMessage("expected statements after 'try'").
			WithPosition(currentToken.Pos, currentToken.Length()).
			WithSuggestion("add statements to execute in the try block").
			WithParsePhase("try statement").
			Build()
		p.addStructuredError(err)
		return nil
	}

	// Check for except clause
	currentToken := p.cursor.Current()
	if currentToken.Type == lexer.EXCEPT {
		stmt.ExceptClause = p.parseExceptClauseCursor()
		if stmt.ExceptClause == nil {
			return nil
		}
	}

	// Check for finally clause
	currentToken = p.cursor.Current()
	if currentToken.Type == lexer.FINALLY {
		stmt.FinallyClause = p.parseFinallyClauseCursor()
		if stmt.FinallyClause == nil {
			return nil
		}
	}

	// Validate that at least one of except or finally is present
	if stmt.ExceptClause == nil && stmt.FinallyClause == nil {
		// Use structured error
		currentToken = p.cursor.Current()
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrUnexpectedToken).
			WithMessage("expected 'except' or 'finally' after 'try' block").
			WithPosition(currentToken.Pos, currentToken.Length()).
			WithExpectedString("'except' or 'finally'").
			WithActual(currentToken.Type, currentToken.Literal).
			WithSuggestion("add an 'except' clause to handle exceptions or a 'finally' clause for cleanup").
			WithParsePhase("try statement").
			Build()
		p.addStructuredError(err)
		return nil
	}

	// Expect 'end' keyword
	currentToken = p.cursor.Current()
	if currentToken.Type != lexer.END {
		// Use structured error
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingEnd).
			WithMessage("expected 'end' to close try statement").
			WithPosition(currentToken.Pos, currentToken.Length()).
			WithExpectedString("'end'").
			WithActual(currentToken.Type, currentToken.Literal).
			WithSuggestion("add 'end' to close the try statement").
			WithParsePhase("try statement").
			Build()
		p.addStructuredError(err)
		return nil
	}

	// End position is at the 'end' keyword
	return builder.FinishWithToken(stmt, currentToken).(*ast.TryStatement)
}

// parseBlockStatementForTryCursor parses statements until 'except', 'finally', or 'end' using cursor mode.
// Task 2.2.14.7: Helper for try block parsing
// PRE: cursor is on first statement token
// POST: cursor is on EXCEPT, FINALLY, or END
func (p *Parser) parseBlockStatementForTryCursor() *ast.BlockStatement {
	startToken := p.cursor.Current()
	block := &ast.BlockStatement{
		BaseNode: ast.BaseNode{Token: startToken},
	}
	block.Statements = []ast.Statement{}

	for {
		currentToken := p.cursor.Current()

		// Termination conditions
		if currentToken.Type == lexer.EXCEPT ||
			currentToken.Type == lexer.FINALLY ||
			currentToken.Type == lexer.END ||
			currentToken.Type == lexer.EOF {
			break
		}

		// Skip semicolons at statement level
		if currentToken.Type == lexer.SEMICOLON {
			p.cursor = p.cursor.Advance()
			continue
		}

		// Parse statement using cursor mode
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

	return block
}

// parseExceptClauseCursor parses an except clause using cursor mode.
// Task 2.2.14.7: Exception handler clause migration
// Syntax:
//   - except <handlers> end
//   - except <handlers> else <statements> end
//
// PRE: cursor is on EXCEPT token
// POST: cursor is on token before FINALLY or END
func (p *Parser) parseExceptClauseCursor() *ast.ExceptClause {
	builder := p.StartNode()
	exceptToken := p.cursor.Current() // Save 'except' token before moving past it
	clause := &ast.ExceptClause{
		Token: exceptToken, // 'except' keyword token
	}
	clause.Handlers = []*ast.ExceptionHandler{}

	p.cursor = p.cursor.Advance() // move past 'except'

	// Parse exception handlers or bare except
	for p.cursor.Current().Type == lexer.ON {
		handler := p.parseExceptionHandlerCursor()
		if handler == nil {
			return nil
		}
		clause.Handlers = append(clause.Handlers, handler)

		// Skip semicolons between handlers
		for p.cursor.Current().Type == lexer.SEMICOLON {
			p.cursor = p.cursor.Advance()
		}
	}

	// If no handlers, this is a bare except - parse statements until finally or end
	if len(clause.Handlers) == 0 {
		// Parse bare except statements into a block
		bareBlock := &ast.BlockStatement{
			BaseNode: ast.BaseNode{Token: p.cursor.Current()},
		}
		bareBlock.Statements = []ast.Statement{}

		for {
			currentToken := p.cursor.Current()

			// Termination conditions
			if currentToken.Type == lexer.FINALLY ||
				currentToken.Type == lexer.END ||
				currentToken.Type == lexer.ELSE ||
				currentToken.Type == lexer.EOF {
				break
			}

			// Skip semicolons
			if currentToken.Type == lexer.SEMICOLON {
				p.cursor = p.cursor.Advance()
				continue
			}

			// Parse statement
			stmt := p.parseStatementCursor()
			if stmt != nil {
				bareBlock.Statements = append(bareBlock.Statements, stmt)
			}
			p.cursor = p.cursor.Advance()

			// Skip semicolons after statement
			for p.cursor.Current().Type == lexer.SEMICOLON {
				p.cursor = p.cursor.Advance()
			}
		}

		// Create a synthetic handler for bare except (catches all)
		if len(bareBlock.Statements) > 0 {
			handlerBuilder := p.StartNode()
			bareHandler := &ast.ExceptionHandler{
				Token:         exceptToken, // Use the 'except' token for synthetic handler
				Variable:      nil,         // No exception variable
				ExceptionType: nil,         // Catches all exception types
				Statement:     bareBlock,
			}
			if bareBlock != nil {
				bareHandler = handlerBuilder.FinishWithNode(bareHandler, bareBlock).(*ast.ExceptionHandler)
			}
			clause.Handlers = append(clause.Handlers, bareHandler)
		}
	}

	// Check for optional else block
	if p.cursor.Current().Type == lexer.ELSE {
		p.cursor = p.cursor.Advance() // move past 'else'

		// Parse else block statements until finally or end
		elseBlock := &ast.BlockStatement{
			BaseNode: ast.BaseNode{Token: p.cursor.Current()},
		}
		elseBlock.Statements = []ast.Statement{}

		for {
			currentToken := p.cursor.Current()

			// Termination conditions
			if currentToken.Type == lexer.FINALLY ||
				currentToken.Type == lexer.END ||
				currentToken.Type == lexer.EOF {
				break
			}

			// Skip semicolons
			if currentToken.Type == lexer.SEMICOLON {
				p.cursor = p.cursor.Advance()
				continue
			}

			// Parse statement
			stmt := p.parseStatementCursor()
			if stmt != nil {
				elseBlock.Statements = append(elseBlock.Statements, stmt)
			}

			p.cursor = p.cursor.Advance()

			// Skip semicolons after statement
			for p.cursor.Current().Type == lexer.SEMICOLON {
				p.cursor = p.cursor.Advance()
			}
		}

		clause.ElseBlock = elseBlock
	}

	// Set EndPos based on what was parsed last
	if clause.ElseBlock != nil {
		return builder.FinishWithNode(clause, clause.ElseBlock).(*ast.ExceptClause)
	} else if len(clause.Handlers) > 0 {
		lastHandler := clause.Handlers[len(clause.Handlers)-1]
		return builder.FinishWithNode(clause, lastHandler).(*ast.ExceptClause)
	} else {
		// No handlers or else block - use current position
		return builder.FinishWithToken(clause, p.cursor.Current()).(*ast.ExceptClause)
	}
}

// parseExceptionHandlerCursor parses an exception handler using cursor mode.
// Task 2.2.14.7: Exception handler migration
// Syntax: on <variable>: <type> do <statement>
//
// PRE: cursor is on ON token
// POST: cursor is on last token of handler statement
func (p *Parser) parseExceptionHandlerCursor() *ast.ExceptionHandler {
	builder := p.StartNode()
	onToken := p.cursor.Current() // Save 'on' token before moving past it
	handler := &ast.ExceptionHandler{
		Token: onToken, // 'on' keyword token
	}

	// Expect 'on' keyword (already checked by caller)
	p.cursor = p.cursor.Advance() // move past 'on'

	// Parse variable name
	currentToken := p.cursor.Current()
	if currentToken.Type != lexer.IDENT {
		// Use structured error
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrExpectedIdent).
			WithMessage("expected identifier after 'on'").
			WithPosition(currentToken.Pos, currentToken.Length()).
			WithExpectedString("exception variable name").
			WithActual(currentToken.Type, currentToken.Literal).
			WithSuggestion("provide a variable name to hold the exception object").
			WithParsePhase("exception handler").
			Build()
		p.addStructuredError(err)
		return nil
	}

	handler.Variable = &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: currentToken,
			},
		},
		Value: currentToken.Literal,
	}

	// Expect ':' token
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.COLON {
		// Use structured error
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingColon).
			WithMessage("expected ':' after exception variable").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpectedString("':'").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("add ':' before the exception type").
			WithParsePhase("exception handler").
			Build()
		p.addStructuredError(err)
		return nil
	}
	p.cursor = p.cursor.Advance() // move to ':'

	// Parse exception type
	p.cursor = p.cursor.Advance() // move past ':'
	currentToken = p.cursor.Current()
	if currentToken.Type != lexer.IDENT {
		// Use structured error
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrExpectedType).
			WithMessage("expected exception type after ':'").
			WithPosition(currentToken.Pos, currentToken.Length()).
			WithExpectedString("exception type name").
			WithActual(currentToken.Type, currentToken.Literal).
			WithSuggestion("specify the exception type, like 'Exception' or 'EMyException'").
			WithParsePhase("exception handler").
			Build()
		p.addStructuredError(err)
		return nil
	}

	handler.ExceptionType = &ast.TypeAnnotation{
		Token: currentToken,
		Name:  currentToken.Literal,
	}

	// Expect 'do' keyword
	nextToken = p.cursor.Peek(1)
	if nextToken.Type != lexer.DO {
		// Use structured error
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingDo).
			WithMessage("expected 'do' after exception type").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpectedString("'do'").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("add 'do' before the handler statement").
			WithParsePhase("exception handler").
			Build()
		p.addStructuredError(err)
		return nil
	}
	p.cursor = p.cursor.Advance() // move to 'do'

	// Parse handler statement
	p.cursor = p.cursor.Advance() // move past 'do'
	handler.Statement = p.parseStatementCursor()

	if handler.Statement == nil {
		// Use structured error
		currentToken = p.cursor.Current()
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidSyntax).
			WithMessage("expected statement after 'do'").
			WithPosition(currentToken.Pos, currentToken.Length()).
			WithSuggestion("add a statement to handle the exception").
			WithParsePhase("exception handler").
			Build()
		p.addStructuredError(err)
		return nil
	}

	// After parsing the statement, advance to the next token
	// This handles both block statements (begin...end) and single statements
	p.cursor = p.cursor.Advance()

	// Set EndPos to the end of the statement
	return builder.FinishWithNode(handler, handler.Statement).(*ast.ExceptionHandler)
}

// parseFinallyClauseCursor parses a finally clause using cursor mode.
// Task 2.2.14.7: Finally clause migration
// Syntax: finally <statements> end
//
// PRE: cursor is on FINALLY token
// POST: cursor is on END token (before END of try)
func (p *Parser) parseFinallyClauseCursor() *ast.FinallyClause {
	finallyToken := p.cursor.Current()
	clause := &ast.FinallyClause{
		BaseNode: ast.BaseNode{Token: finallyToken},
	}

	p.cursor = p.cursor.Advance() // move past 'finally'

	// Parse finally block statements until 'end'
	block := &ast.BlockStatement{
		BaseNode: ast.BaseNode{Token: p.cursor.Current()},
	}
	block.Statements = []ast.Statement{}

	for {
		currentToken := p.cursor.Current()

		// Termination conditions
		if currentToken.Type == lexer.END || currentToken.Type == lexer.EOF {
			break
		}

		// Skip semicolons
		if currentToken.Type == lexer.SEMICOLON {
			p.cursor = p.cursor.Advance()
			continue
		}

		// Parse statement
		stmt := p.parseStatementCursor()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		p.cursor = p.cursor.Advance()

		// Skip any semicolons after the statement
		for p.cursor.Current().Type == lexer.SEMICOLON {
			p.cursor = p.cursor.Advance()
		}
	}

	clause.Block = block

	return clause
}
