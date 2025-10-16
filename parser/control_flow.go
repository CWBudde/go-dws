package parser

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
)

// parseIfStatement parses an if-then-else statement.
// Syntax: if <condition> then <statement> [else <statement>]
func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	// Move past 'if' and parse the condition
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if stmt.Condition == nil {
		p.addError("expected condition after 'if'")
		return nil
	}

	// Expect 'then' keyword
	if !p.expectPeek(lexer.THEN) {
		return nil
	}

	// Parse the consequence (then branch)
	p.nextToken()
	stmt.Consequence = p.parseStatement()

	if stmt.Consequence == nil {
		p.addError("expected statement after 'then'")
		return nil
	}

	// Check for optional 'else' branch
	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken() // move to 'else'
		p.nextToken() // move to statement after 'else'
		stmt.Alternative = p.parseStatement()

		if stmt.Alternative == nil {
			p.addError("expected statement after 'else'")
			return nil
		}
	}

	return stmt
}

// parseWhileStatement parses a while loop statement.
// Syntax: while <condition> do <statement>
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}

	// Move past 'while' and parse the condition
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if stmt.Condition == nil {
		p.addError("expected condition after 'while'")
		return nil
	}

	// Expect 'do' keyword
	if !p.expectPeek(lexer.DO) {
		return nil
	}

	// Parse the body statement
	p.nextToken()
	stmt.Body = p.parseStatement()

	if stmt.Body == nil {
		p.addError("expected statement after 'do'")
		return nil
	}

	return stmt
}

// parseRepeatStatement parses a repeat-until loop statement.
// Syntax: repeat <statement> until <condition>
func (p *Parser) parseRepeatStatement() *ast.RepeatStatement {
	stmt := &ast.RepeatStatement{Token: p.curToken}

	// Move past 'repeat' and parse the body statement
	p.nextToken()
	stmt.Body = p.parseStatement()

	if stmt.Body == nil {
		p.addError("expected statement after 'repeat'")
		return nil
	}

	// Expect 'until' keyword
	if !p.expectPeek(lexer.UNTIL) {
		return nil
	}

	// Parse the condition
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if stmt.Condition == nil {
		p.addError("expected condition after 'until'")
		return nil
	}

	return stmt
}

// parseForStatement parses a for loop statement.
// Syntax: for <variable> := <start> to|downto <end> do <statement>
func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	// Move past 'for' and parse the loop variable identifier
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Variable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Expect ':=' assignment operator
	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	// Parse the start expression
	p.nextToken()
	stmt.Start = p.parseExpression(LOWEST)

	if stmt.Start == nil {
		p.addError("expected start expression in for loop")
		return nil
	}

	// Parse direction keyword ('to' or 'downto')
	// We need to check the peek token and advance if it's either TO or DOWNTO
	if !p.peekTokenIs(lexer.TO) && !p.peekTokenIs(lexer.DOWNTO) {
		p.addError("expected 'to' or 'downto' in for loop")
		return nil
	}
	p.nextToken() // Move to TO or DOWNTO

	// Set direction based on token
	if p.curTokenIs(lexer.TO) {
		stmt.Direction = ast.ForTo
	} else if p.curTokenIs(lexer.DOWNTO) {
		stmt.Direction = ast.ForDownto
	} else {
		p.addError("expected 'to' or 'downto' in for loop")
		return nil
	}

	// Parse the end expression
	p.nextToken()
	stmt.End = p.parseExpression(LOWEST)

	if stmt.End == nil {
		p.addError("expected end expression in for loop")
		return nil
	}

	// Expect 'do' keyword
	if !p.expectPeek(lexer.DO) {
		return nil
	}

	// Parse the body statement
	p.nextToken()
	stmt.Body = p.parseStatement()

	if stmt.Body == nil {
		p.addError("expected statement after 'do'")
		return nil
	}

	return stmt
}

// parseCaseStatement parses a case statement.
// Syntax: case <expression> of <value>: <statement>; ... [else <statement>;] end;
func (p *Parser) parseCaseStatement() *ast.CaseStatement {
	stmt := &ast.CaseStatement{Token: p.curToken}

	// Move past 'case' and parse the case expression
	p.nextToken()
	stmt.Expression = p.parseExpression(LOWEST)

	if stmt.Expression == nil {
		p.addError("expected expression after 'case'")
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

		branch := &ast.CaseBranch{Token: p.curToken}

		// Parse comma-separated value list
		branch.Values = []ast.Expression{}

		// Parse first value
		value := p.parseExpression(LOWEST)
		if value == nil {
			p.addError("expected value in case branch")
			return nil
		}
		branch.Values = append(branch.Values, value)

		// Parse additional comma-separated values
		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // move to comma
			p.nextToken() // move to next value
			value := p.parseExpression(LOWEST)
			if value == nil {
				p.addError("expected value after comma in case branch")
				return nil
			}
			branch.Values = append(branch.Values, value)
		}

		// Expect ':' after value(s)
		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		// Parse the statement for this branch
		p.nextToken()
		branch.Statement = p.parseStatement()

		if branch.Statement == nil {
			p.addError("expected statement after ':' in case branch")
			return nil
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
		stmt.Else = p.parseStatement()

		if stmt.Else == nil {
			p.addError("expected statement after 'else' in case statement")
			return nil
		}

		// Move to 'end' or semicolon
		p.nextToken()

		// Skip any trailing semicolons before 'end'
		for p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}
	}

	// Expect 'end' keyword
	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close case statement")
		return nil
	}

	return stmt
}
