package parser

import (
	"fmt"
	"reflect"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// isNilStatement checks if a statement is nil, including typed nils.
// In Go, an interface can contain a nil pointer but not be nil itself,
// which causes issues when calling methods on the interface.
func isNilStatement(stmt ast.Statement) bool {
	if stmt == nil {
		return true
	}
	// Use reflection to check if the underlying value is nil
	v := reflect.ValueOf(stmt)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

// parseBreakStatement parses a break statement.
// Syntax: break;
// PRE: curToken is BREAK
// POST: curToken is SEMICOLON
func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}

	// Expect semicolon after break
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	stmt.EndPos = p.endPosFromToken(p.curToken) // p.curToken is now at SEMICOLON
	return stmt
}

// parseContinueStatement parses a continue statement.
// Syntax: continue;
// PRE: curToken is CONTINUE
// POST: curToken is SEMICOLON
func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	stmt := &ast.ContinueStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}

	// Expect semicolon after continue
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	stmt.EndPos = p.endPosFromToken(p.curToken) // p.curToken is now at SEMICOLON
	return stmt
}

// parseExitStatement parses an exit statement.
// Syntax: exit; exit value; or exit(value);
// PRE: curToken is EXIT
// POST: curToken is SEMICOLON
func (p *Parser) parseExitStatement() *ast.ExitStatement {
	stmt := &ast.ExitStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}

	// Check if there's a parenthesized return value: exit(value)
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // move to '('
		p.nextToken() // move to expression

		stmt.ReturnValue = p.parseExpression(LOWEST)

		if stmt.ReturnValue == nil {
			p.addError("expected expression after 'exit('", ErrInvalidExpression)
			return nil
		}

		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	} else if _, ok := p.prefixParseFns[p.peekToken.Type]; ok && !p.peekTokenIs(lexer.SEMICOLON) {
		// Support exit with inline expression: exit value;
		p.nextToken()
		stmt.ReturnValue = p.parseExpression(LOWEST)

		if stmt.ReturnValue == nil {
			p.addError("expected expression after 'exit'", ErrInvalidExpression)
			return nil
		}
	}

	// Expect semicolon after exit or exit(value)
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	stmt.EndPos = p.endPosFromToken(p.curToken) // p.curToken is now at SEMICOLON
	return stmt
}

// parseIfStatement parses an if-then-else statement.
// Syntax: if <condition> then <statement> [else <statement>]
// PRE: curToken is IF
// POST: curToken is last token of consequence or alternative statement
func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}

	// Track block context for better error messages
	p.pushBlockContext("if", p.curToken.Pos)
	defer p.popBlockContext()

	// Move past 'if' and parse the condition
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if stmt.Condition == nil {
		// Use structured error for better diagnostics
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidExpression).
			WithMessage("expected condition after 'if'").
			WithPosition(p.curToken.Pos, p.curToken.Length()).
			WithExpectedString("boolean expression").
			WithSuggestion("add a condition like 'x > 0' or 'flag = true'").
			WithParsePhase("if statement condition").
			Build()
		p.addStructuredError(err)
		p.synchronize([]lexer.TokenType{lexer.THEN, lexer.ELSE, lexer.END})
		return nil
	}

	// Expect 'then' keyword
	if !p.expectPeek(lexer.THEN) {
		// Use structured error for missing 'then'
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingThen).
			WithMessage("expected 'then' after if condition").
			WithPosition(p.peekToken.Pos, p.peekToken.Length()).
			WithExpected(lexer.THEN).
			WithActual(p.peekToken.Type, p.peekToken.Literal).
			WithSuggestion("add 'then' keyword after the condition").
			WithNote("DWScript if statements require: if <condition> then <statement>").
			WithParsePhase("if statement").
			Build()
		p.addStructuredError(err)
		p.synchronize([]lexer.TokenType{lexer.THEN, lexer.ELSE, lexer.END})
		if !p.curTokenIs(lexer.THEN) {
			return nil
		}
	}

	// Parse the consequence (then branch)
	p.nextToken()
	stmt.Consequence = p.parseStatement()

	if stmt.Consequence == nil {
		// Use structured error for missing statement
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidSyntax).
			WithMessage("expected statement after 'then'").
			WithPosition(p.curToken.Pos, p.curToken.Length()).
			WithExpectedString("statement").
			WithSuggestion("add a statement like a variable assignment or function call").
			WithParsePhase("if statement consequence").
			Build()
		p.addStructuredError(err)
		p.synchronize([]lexer.TokenType{lexer.ELSE, lexer.END})
		return nil
	}

	// Check for optional 'else' branch
	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken() // move to 'else'
		p.nextToken() // move to statement after 'else'
		stmt.Alternative = p.parseStatement()

		if stmt.Alternative == nil {
			// Use structured error for missing else statement
			err := NewStructuredError(ErrKindInvalid).
				WithCode(ErrInvalidSyntax).
				WithMessage("expected statement after 'else'").
				WithPosition(p.curToken.Pos, p.curToken.Length()).
				WithExpectedString("statement").
				WithSuggestion("add a statement for the else branch").
				WithParsePhase("if statement alternative").
				Build()
			p.addStructuredError(err)
			p.synchronize([]lexer.TokenType{lexer.END})
			return nil
		}
		// End position is after the alternative statement
		stmt.EndPos = stmt.Alternative.End()
	} else {
		// No else branch - end position is after the consequence
		stmt.EndPos = stmt.Consequence.End()
	}

	return stmt
}

// parseWhileStatement parses a while loop statement.
// Syntax: while <condition> do <statement>
// PRE: curToken is WHILE
// POST: curToken is last token of body statement
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}

	// Track block context for better error messages
	p.pushBlockContext("while", p.curToken.Pos)
	defer p.popBlockContext()

	// Move past 'while' and parse the condition
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if stmt.Condition == nil {
		// Use structured error for better diagnostics
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidExpression).
			WithMessage("expected condition after 'while'").
			WithPosition(p.curToken.Pos, p.curToken.Length()).
			WithExpectedString("boolean expression").
			WithSuggestion("add a loop condition like 'count < 10'").
			WithParsePhase("while loop condition").
			Build()
		p.addStructuredError(err)
		p.synchronize([]lexer.TokenType{lexer.DO, lexer.END})
		return nil
	}

	// Expect 'do' keyword
	if !p.expectPeek(lexer.DO) {
		// Use structured error for missing 'do'
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingDo).
			WithMessage("expected 'do' after while condition").
			WithPosition(p.peekToken.Pos, p.peekToken.Length()).
			WithExpected(lexer.DO).
			WithActual(p.peekToken.Type, p.peekToken.Literal).
			WithSuggestion("add 'do' keyword after the condition").
			WithNote("DWScript while loops require: while <condition> do <statement>").
			WithParsePhase("while loop").
			Build()
		p.addStructuredError(err)
		p.synchronize([]lexer.TokenType{lexer.DO, lexer.END})
		if !p.curTokenIs(lexer.DO) {
			return nil
		}
	}

	// Parse the body statement
	p.nextToken()
	stmt.Body = p.parseStatement()

	if isNilStatement(stmt.Body) {
		// Use structured error for missing loop body
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidSyntax).
			WithMessage("expected statement after 'do'").
			WithPosition(p.curToken.Pos, p.curToken.Length()).
			WithExpectedString("statement").
			WithSuggestion("add a statement for the loop body").
			WithParsePhase("while loop body").
			Build()
		p.addStructuredError(err)
		p.synchronize([]lexer.TokenType{lexer.END})
		return nil
	}

	// End position is after the body statement
	stmt.EndPos = stmt.Body.End()

	return stmt
}

// parseRepeatStatement parses a repeat-until loop statement.
// Syntax: repeat <statements> until <condition>
// Note: The body can contain multiple statements
// PRE: curToken is REPEAT
// POST: curToken is last token of condition expression
func (p *Parser) parseRepeatStatement() *ast.RepeatStatement {
	stmt := &ast.RepeatStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}

	// Track block context for better error messages
	p.pushBlockContext("repeat", p.curToken.Pos)
	defer p.popBlockContext()

	// Move past 'repeat'
	p.nextToken()

	// Parse multiple statements until 'until' is encountered
	block := &ast.BlockStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}
	block.Statements = []ast.Statement{}

	for !p.curTokenIs(lexer.UNTIL) && !p.curTokenIs(lexer.EOF) {
		// Skip semicolons
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		bodyStmt := p.parseStatement()
		if bodyStmt != nil {
			block.Statements = append(block.Statements, bodyStmt)
		}

		p.nextToken()

		// Skip any semicolons after the statement
		for p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}
	}

	// If only one statement, use it directly; otherwise use the block
	if len(block.Statements) == 1 {
		stmt.Body = block.Statements[0]
	} else if len(block.Statements) > 1 {
		stmt.Body = block
	} else {
		p.addErrorWithContext("expected at least one statement in repeat body", ErrInvalidSyntax)
		p.synchronize([]lexer.TokenType{lexer.UNTIL, lexer.END})
		return nil
	}

	// Expect 'until' keyword
	if !p.curTokenIs(lexer.UNTIL) {
		p.addErrorWithContext(fmt.Sprintf("expected 'until' after repeat body, got %s instead", p.curToken.Type), ErrUnexpectedToken)
		p.synchronize([]lexer.TokenType{lexer.UNTIL, lexer.END})
		if !p.curTokenIs(lexer.UNTIL) {
			return nil
		}
	}

	// Parse the condition
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if stmt.Condition == nil {
		p.addError("expected condition after 'until'", ErrInvalidExpression)
		return nil
	}

	// End position is after the condition expression
	stmt.EndPos = stmt.Condition.End()

	return stmt
}

// parseForStatement parses a for loop statement.
// Syntax:
//
//	for <variable> := <start> to|downto <end> [step <step>] do <statement>
//	for [var] <variable> in <expression> do <statement>
//
// PRE: curToken is FOR
// POST: curToken is last token of body statement
func (p *Parser) parseForStatement() ast.Statement {
	forToken := p.curToken

	// Move past 'for' and parse optional inline var declaration
	inlineVar := false
	if p.peekTokenIs(lexer.VAR) {
		p.nextToken() // move to 'var'
		inlineVar = true
	}

	// Expect loop variable identifier
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	variable := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

	// Check if this is a for-in loop (IN) or for-to/downto loop (:=)
	if p.peekTokenIs(lexer.IN) {
		// Parse for-in loop: for [var] x in collection do statement
		return p.parseForInLoop(forToken, variable, inlineVar)
	}

	// Parse traditional for-to/downto loop
	stmt := &ast.ForStatement{
		BaseNode:  ast.BaseNode{Token: forToken},
		Variable:  variable,
		InlineVar: inlineVar,
	}

	// Expect ':=' assignment operator
	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	// Parse the start expression
	p.nextToken()
	stmt.Start = p.parseExpression(LOWEST)

	if stmt.Start == nil {
		p.addError("expected start expression in for loop", ErrInvalidExpression)
		return nil
	}

	// Parse direction keyword ('to' or 'downto')
	// We need to check the peek token and advance if it's either TO or DOWNTO
	if !p.peekTokenIs(lexer.TO) && !p.peekTokenIs(lexer.DOWNTO) {
		p.addError("expected 'to' or 'downto' in for loop", ErrMissingTo)
		return nil
	}
	p.nextToken() // Move to TO or DOWNTO

	// Set direction based on token
	if p.curTokenIs(lexer.TO) {
		stmt.Direction = ast.ForTo
	} else if p.curTokenIs(lexer.DOWNTO) {
		stmt.Direction = ast.ForDownto
	} else {
		p.addError("expected 'to' or 'downto' in for loop", ErrMissingTo)
		return nil
	}

	// Parse the end expression
	p.nextToken()
	stmt.EndValue = p.parseExpression(LOWEST)

	if stmt.EndValue == nil {
		p.addError("expected end expression in for loop", ErrInvalidExpression)
		return nil
	}

	// Check for optional 'step' keyword
	if p.peekTokenIs(lexer.STEP) {
		p.nextToken() // move to 'step'
		p.nextToken() // move to step expression
		stmt.Step = p.parseExpression(LOWEST)

		if stmt.Step == nil {
			p.addError("expected expression after 'step'", ErrInvalidExpression)
			return nil
		}
	}

	// Expect 'do' keyword
	if !p.expectPeek(lexer.DO) {
		return nil
	}

	// Parse the body statement
	p.nextToken()
	stmt.Body = p.parseStatement()

	if isNilStatement(stmt.Body) {
		p.addError("expected statement after 'do'", ErrInvalidSyntax)
		return nil
	}

	// End position is after the body statement
	stmt.EndPos = stmt.Body.End()

	return stmt
}

// parseForInLoop parses a for-in loop statement.
// Syntax: for [var] <variable> in <expression> do <statement>
// PRE: curToken is variable IDENT
// POST: curToken is last token of body statement
func (p *Parser) parseForInLoop(forToken lexer.Token, variable *ast.Identifier, inlineVar bool) *ast.ForInStatement {
	stmt := &ast.ForInStatement{
		BaseNode:  ast.BaseNode{Token: forToken},
		Variable:  variable,
		InlineVar: inlineVar,
	}

	// Move past variable to 'in' keyword
	if !p.expectPeek(lexer.IN) {
		return nil
	}

	// Parse the collection expression
	p.nextToken()
	stmt.Collection = p.parseExpression(LOWEST)

	if stmt.Collection == nil {
		p.addError("expected expression after 'in'", ErrInvalidExpression)
		return nil
	}

	// Expect 'do' keyword
	if !p.expectPeek(lexer.DO) {
		return nil
	}

	// Parse the body statement
	p.nextToken()
	stmt.Body = p.parseStatement()

	if isNilStatement(stmt.Body) {
		p.addError("expected statement after 'do'", ErrInvalidSyntax)
		return nil
	}

	// End position is after the body statement
	stmt.EndPos = stmt.Body.End()

	return stmt
}

// parseCaseStatement parses a case statement.
// Syntax: case <expression> of <value>: <statement>; ... [else <statement>;] end;
// PRE: curToken is CASE
// POST: curToken is END
func (p *Parser) parseCaseStatement() *ast.CaseStatement {
	stmt := &ast.CaseStatement{
		BaseNode: ast.BaseNode{Token: p.curToken},
	}

	// Track block context for better error messages
	p.pushBlockContext("case", p.curToken.Pos)
	defer p.popBlockContext()

	// Move past 'case' and parse the case expression
	p.nextToken()
	stmt.Expression = p.parseExpression(LOWEST)

	if stmt.Expression == nil {
		p.addErrorWithContext("expected expression after 'case'", ErrInvalidExpression)
		return nil
	}

	// Expect 'of' keyword
	if !p.expectPeek(lexer.OF) {
		return nil
	}

	// Parse case branches
	stmt.Cases = []*ast.CaseBranch{}

	// Move past 'of'
	p.nextToken()

	// Parse case branches until we hit 'else' or 'end'
	for !p.curTokenIs(lexer.ELSE) && !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
		// Skip any leading semicolons
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// Save the token of the first value for position tracking
		firstValueToken := p.curToken
		branch := &ast.CaseBranch{
			Token: firstValueToken, // First value token for position tracking
		}

		// Parse comma-separated value list (with range support)
		branch.Values = []ast.Expression{}

		// Parse first value or range
		value := p.parseExpression(LOWEST)
		if value == nil {
			p.addError("expected value in case branch", ErrInvalidExpression)
			return nil
		}

		// Check for range operator (..)
		if p.peekTokenIs(lexer.DOTDOT) {
			p.nextToken() // move to '..'
			rangeToken := p.curToken

			p.nextToken() // move to end expression
			endValue := p.parseExpression(LOWEST)
			if endValue == nil {
				p.addError("expected expression after '..' in case range", ErrInvalidExpression)
				return nil
			}

			// Create RangeExpression
			rangeExpr := &ast.RangeExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: rangeToken,
					},
				},
				Start:    value,
				RangeEnd: endValue,
			}
			branch.Values = append(branch.Values, rangeExpr)
		} else {
			// Simple value (not a range)
			branch.Values = append(branch.Values, value)
		}

		// Parse additional comma-separated values/ranges
		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // move to comma
			p.nextToken() // move to next value

			value := p.parseExpression(LOWEST)
			if value == nil {
				p.addError("expected value after comma in case branch", ErrInvalidExpression)
				return nil
			}

			// Check for range
			if p.peekTokenIs(lexer.DOTDOT) {
				p.nextToken() // move to '..'
				rangeToken := p.curToken

				p.nextToken() // move to end expression
				endValue := p.parseExpression(LOWEST)
				if endValue == nil {
					p.addError("expected expression after '..' in case range", ErrInvalidExpression)
					return nil
				}

				rangeExpr := &ast.RangeExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: rangeToken,
						},
					},
					Start:    value,
					RangeEnd: endValue,
				}
				branch.Values = append(branch.Values, rangeExpr)
			} else {
				branch.Values = append(branch.Values, value)
			}
		}

		// Expect ':' after value(s)
		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		// Parse the statement for this branch
		p.nextToken()
		branch.Statement = p.parseStatement()

		if branch.Statement == nil {
			p.addError("expected statement after ':' in case branch", ErrInvalidSyntax)
			return nil
		}

		// Set EndPos to the end of the statement
		if branch.Statement != nil {
			branch.EndPos = branch.Statement.End()
		}

		stmt.Cases = append(stmt.Cases, branch)

		// Move to next token (could be semicolon, else, or end)
		p.nextToken()

		// Skip any trailing semicolons
		for p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}
	}

	// Check for optional 'else' branch
	if p.curTokenIs(lexer.ELSE) {
		p.nextToken() // move past 'else'

		// Parse multiple statements until 'end' is encountered (like repeat-until)
		// DWScript allows multiple statements in else clause without begin-end
		block := &ast.BlockStatement{
			BaseNode: ast.BaseNode{Token: p.curToken},
		}
		block.Statements = []ast.Statement{}

		for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
			// Skip semicolons
			if p.curTokenIs(lexer.SEMICOLON) {
				p.nextToken()
				continue
			}

			elseStmt := p.parseStatement()
			if elseStmt != nil {
				block.Statements = append(block.Statements, elseStmt)
			}

			p.nextToken()

			// Skip any semicolons after the statement
			for p.curTokenIs(lexer.SEMICOLON) {
				p.nextToken()
			}
		}

		// If only one statement, use it directly; otherwise use the block
		if len(block.Statements) == 1 {
			stmt.Else = block.Statements[0]
		} else if len(block.Statements) > 1 {
			stmt.Else = block
		} else {
			p.addError("expected statement after 'else' in case statement", ErrInvalidSyntax)
			return nil
		}
	}

	// Expect 'end' keyword
	if !p.curTokenIs(lexer.END) {
		p.addErrorWithContext("expected 'end' to close case statement", ErrMissingEnd)
		p.synchronize([]lexer.TokenType{lexer.END})
		return nil
	}

	// End position is at the 'end' keyword
	stmt.EndPos = p.endPosFromToken(p.curToken)

	return stmt
}

// parseIfExpression parses an inline if-then-else conditional expression.
// Syntax: if <condition> then <expression> [else <expression>]
// This is similar to a ternary operator: condition ? value1 : value2
// PRE: curToken is IF
// POST: curToken is last token of consequence or alternative expression
func (p *Parser) parseIfExpression() ast.Expression {
	expr := &ast.IfExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: p.curToken},
		},
	}

	// Move past 'if' and parse the condition
	p.nextToken()
	expr.Condition = p.parseExpression(LOWEST)

	if expr.Condition == nil {
		p.addError("expected condition after 'if'", ErrInvalidExpression)
		return nil
	}

	// Expect 'then' keyword
	if !p.expectPeek(lexer.THEN) {
		return nil
	}

	// Parse the consequence (then branch) as an expression
	p.nextToken()
	expr.Consequence = p.parseExpression(LOWEST)

	if expr.Consequence == nil {
		p.addError("expected expression after 'then'", ErrInvalidSyntax)
		return nil
	}

	// Check for optional 'else' branch
	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken() // move to 'else'
		p.nextToken() // move to expression after 'else'
		expr.Alternative = p.parseExpression(LOWEST)

		if expr.Alternative == nil {
			p.addError("expected expression after 'else'", ErrInvalidSyntax)
			return nil
		}
		// End position is after the alternative expression
		expr.EndPos = expr.Alternative.End()
	} else {
		// No else branch - end position is after the consequence
		// The else clause is optional; if omitted, default value is returned
		expr.EndPos = expr.Consequence.End()
	}

	return expr
}

// ============================================================================
// Task 2.2.14.4: Cursor-mode handlers for control flow statements
// ============================================================================

// parseIfStatementCursor parses an if statement in cursor mode.
// Task 2.2.14.4: If statement migration
// Syntax: if <condition> then <statement> [else <statement>]
// PRE: cursor is on IF token
// POST: cursor is on last token of statement
func (p *Parser) parseIfStatementCursor() *ast.IfStatement {
	ifToken := p.cursor.Current()
	stmt := &ast.IfStatement{
		BaseNode: ast.BaseNode{Token: ifToken},
	}

	// Track block context for better error messages
	p.pushBlockContext("if", ifToken.Pos)
	defer p.popBlockContext()

	// Move past 'if' and parse the condition
	p.cursor = p.cursor.Advance()
	stmt.Condition = p.parseExpressionCursor(LOWEST)

	if stmt.Condition == nil {
		// Use structured error for better diagnostics
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidExpression).
			WithMessage("expected condition after 'if'").
			WithPosition(p.cursor.Current().Pos, p.cursor.Current().Length()).
			WithExpectedString("boolean expression").
			WithSuggestion("add a condition like 'x > 0' or 'flag = true'").
			WithParsePhase("if statement condition").
			Build()
		p.addStructuredError(err)
		// Synchronize using traditional mode
		p.syncCursorToTokens()
		p.useCursor = false
		p.synchronize([]lexer.TokenType{lexer.THEN, lexer.ELSE, lexer.END})
		p.useCursor = true
		p.syncTokensToCursor()
		return nil
	}

	// Expect 'then' keyword
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.THEN {
		// Use structured error for missing 'then'
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingThen).
			WithMessage("expected 'then' after if condition").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpected(lexer.THEN).
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("add 'then' keyword after the condition").
			WithNote("DWScript if statements require: if <condition> then <statement>").
			WithParsePhase("if statement").
			Build()
		p.addStructuredError(err)
		// Synchronize using traditional mode
		p.syncCursorToTokens()
		p.useCursor = false
		p.synchronize([]lexer.TokenType{lexer.THEN, lexer.ELSE, lexer.END})
		p.useCursor = true
		p.syncTokensToCursor()
		if p.cursor.Current().Type != lexer.THEN {
			return nil
		}
	}

	// Advance past 'then'
	p.cursor = p.cursor.Advance()

	// Parse the consequence (then branch)
	p.cursor = p.cursor.Advance()
	stmt.Consequence = p.parseStatementCursor()

	if stmt.Consequence == nil {
		// Use structured error for missing statement
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidSyntax).
			WithMessage("expected statement after 'then'").
			WithPosition(p.cursor.Current().Pos, p.cursor.Current().Length()).
			WithExpectedString("statement").
			WithSuggestion("add a statement like a variable assignment or function call").
			WithParsePhase("if statement consequence").
			Build()
		p.addStructuredError(err)
		// Synchronize using traditional mode
		p.syncCursorToTokens()
		p.useCursor = false
		p.synchronize([]lexer.TokenType{lexer.ELSE, lexer.END})
		p.useCursor = true
		p.syncTokensToCursor()
		return nil
	}

	// Check for optional 'else' branch
	nextToken = p.cursor.Peek(1)
	if nextToken.Type == lexer.ELSE {
		p.cursor = p.cursor.Advance() // move to 'else'
		p.cursor = p.cursor.Advance() // move to statement after 'else'
		stmt.Alternative = p.parseStatementCursor()

		if stmt.Alternative == nil {
			// Use structured error for missing else statement
			err := NewStructuredError(ErrKindInvalid).
				WithCode(ErrInvalidSyntax).
				WithMessage("expected statement after 'else'").
				WithPosition(p.cursor.Current().Pos, p.cursor.Current().Length()).
				WithExpectedString("statement").
				WithSuggestion("add a statement for the else branch").
				WithParsePhase("if statement alternative").
				Build()
			p.addStructuredError(err)
			// Synchronize using traditional mode
			p.syncCursorToTokens()
			p.useCursor = false
			p.synchronize([]lexer.TokenType{lexer.END})
			p.useCursor = true
			p.syncTokensToCursor()
			return nil
		}
		// End position is after the alternative statement
		stmt.EndPos = stmt.Alternative.End()
	} else {
		// No else branch - end position is after the consequence
		stmt.EndPos = stmt.Consequence.End()
	}

	return stmt
}

// parseWhileStatementCursor parses a while loop statement in cursor mode.
// Task 2.2.14.4: While statement migration
// Syntax: while <condition> do <statement>
// PRE: cursor is on WHILE token
// POST: cursor is on last token of body statement
func (p *Parser) parseWhileStatementCursor() *ast.WhileStatement {
	whileToken := p.cursor.Current()
	stmt := &ast.WhileStatement{
		BaseNode: ast.BaseNode{Token: whileToken},
	}

	// Track block context for better error messages
	p.pushBlockContext("while", whileToken.Pos)
	defer p.popBlockContext()

	// Move past 'while' and parse the condition
	p.cursor = p.cursor.Advance()
	stmt.Condition = p.parseExpressionCursor(LOWEST)

	if stmt.Condition == nil {
		// Use structured error for better diagnostics
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidExpression).
			WithMessage("expected condition after 'while'").
			WithPosition(p.cursor.Current().Pos, p.cursor.Current().Length()).
			WithExpectedString("boolean expression").
			WithSuggestion("add a loop condition like 'count < 10'").
			WithParsePhase("while loop condition").
			Build()
		p.addStructuredError(err)
		// Synchronize using traditional mode
		p.syncCursorToTokens()
		p.useCursor = false
		p.synchronize([]lexer.TokenType{lexer.DO, lexer.END})
		p.useCursor = true
		p.syncTokensToCursor()
		return nil
	}

	// Expect 'do' keyword
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.DO {
		// Use structured error for missing 'do'
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingDo).
			WithMessage("expected 'do' after while condition").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpected(lexer.DO).
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("add 'do' keyword after the condition").
			WithNote("DWScript while loops require: while <condition> do <statement>").
			WithParsePhase("while loop").
			Build()
		p.addStructuredError(err)
		// Synchronize using traditional mode
		p.syncCursorToTokens()
		p.useCursor = false
		p.synchronize([]lexer.TokenType{lexer.DO, lexer.END})
		p.useCursor = true
		p.syncTokensToCursor()
		if p.cursor.Current().Type != lexer.DO {
			return nil
		}
	}

	// Advance past 'do'
	p.cursor = p.cursor.Advance()

	// Parse the body statement
	p.cursor = p.cursor.Advance()
	stmt.Body = p.parseStatementCursor()

	if isNilStatement(stmt.Body) {
		// Use structured error for missing loop body
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidSyntax).
			WithMessage("expected statement after 'do'").
			WithPosition(p.cursor.Current().Pos, p.cursor.Current().Length()).
			WithExpectedString("statement").
			WithSuggestion("add a statement for the loop body").
			WithParsePhase("while loop body").
			Build()
		p.addStructuredError(err)
		// Synchronize using traditional mode
		p.syncCursorToTokens()
		p.useCursor = false
		p.synchronize([]lexer.TokenType{lexer.END})
		p.useCursor = true
		p.syncTokensToCursor()
		return nil
	}

	// End position is after the body statement
	stmt.EndPos = stmt.Body.End()

	return stmt
}

// parseRepeatStatementCursor parses a repeat-until loop statement in cursor mode.
// Task 2.2.14.4: Repeat statement migration
// Syntax: repeat <statements> until <condition>
// Note: The body can contain multiple statements
// PRE: cursor is on REPEAT token
// POST: cursor is on last token of condition expression
func (p *Parser) parseRepeatStatementCursor() *ast.RepeatStatement {
	repeatToken := p.cursor.Current()
	stmt := &ast.RepeatStatement{
		BaseNode: ast.BaseNode{Token: repeatToken},
	}

	// Track block context for better error messages
	p.pushBlockContext("repeat", repeatToken.Pos)
	defer p.popBlockContext()

	// Move past 'repeat'
	p.cursor = p.cursor.Advance()

	// Parse multiple statements until 'until' is encountered
	block := &ast.BlockStatement{
		BaseNode: ast.BaseNode{Token: p.cursor.Current()},
	}
	block.Statements = []ast.Statement{}

	for p.cursor.Current().Type != lexer.UNTIL && p.cursor.Current().Type != lexer.EOF {
		// Skip semicolons
		if p.cursor.Current().Type == lexer.SEMICOLON {
			p.cursor = p.cursor.Advance()
			continue
		}

		bodyStmt := p.parseStatementCursor()
		if bodyStmt != nil {
			block.Statements = append(block.Statements, bodyStmt)
		}

		p.cursor = p.cursor.Advance()

		// Skip any semicolons after the statement
		for p.cursor.Current().Type == lexer.SEMICOLON {
			p.cursor = p.cursor.Advance()
		}
	}

	// If only one statement, use it directly; otherwise use the block
	if len(block.Statements) == 1 {
		stmt.Body = block.Statements[0]
	} else if len(block.Statements) > 1 {
		stmt.Body = block
	} else {
		p.addErrorWithContext("expected at least one statement in repeat body", ErrInvalidSyntax)
		// Synchronize using traditional mode
		p.syncCursorToTokens()
		p.useCursor = false
		p.synchronize([]lexer.TokenType{lexer.UNTIL, lexer.END})
		p.useCursor = true
		p.syncTokensToCursor()
		return nil
	}

	// Expect 'until' keyword
	if p.cursor.Current().Type != lexer.UNTIL {
		p.addErrorWithContext(fmt.Sprintf("expected 'until' after repeat body, got %s instead", p.cursor.Current().Type), ErrUnexpectedToken)
		// Synchronize using traditional mode
		p.syncCursorToTokens()
		p.useCursor = false
		p.synchronize([]lexer.TokenType{lexer.UNTIL, lexer.END})
		p.useCursor = true
		p.syncTokensToCursor()
		if p.cursor.Current().Type != lexer.UNTIL {
			return nil
		}
	}

	// Parse the condition
	p.cursor = p.cursor.Advance()
	stmt.Condition = p.parseExpressionCursor(LOWEST)

	if stmt.Condition == nil {
		p.addError("expected condition after 'until'", ErrInvalidExpression)
		return nil
	}

	// End position is after the condition expression
	stmt.EndPos = stmt.Condition.End()

	return stmt
}
