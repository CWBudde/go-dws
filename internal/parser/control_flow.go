package parser

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseBreakStatement parses a break statement.
// Syntax: break;
func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{Token: p.curToken}

	// Expect semicolon after break
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return stmt
}

// parseContinueStatement parses a continue statement.
// Syntax: continue;
func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	stmt := &ast.ContinueStatement{Token: p.curToken}

	// Expect semicolon after continue
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return stmt
}

// parseExitStatement parses an exit statement.
// Syntax: exit; or exit(value);
func (p *Parser) parseExitStatement() *ast.ExitStatement {
	stmt := &ast.ExitStatement{Token: p.curToken}

	// Check if there's a return value: exit(value)
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // move to '('
		p.nextToken() // move to expression

		stmt.Value = p.parseExpression(LOWEST)

		if stmt.Value == nil {
			p.addError("expected expression after 'exit('")
			return nil
		}

		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	}

	// Expect semicolon after exit or exit(value)
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return stmt
}

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

	// Advance past the statement
	p.nextToken()

	// Skip any optional semicolons before 'until'
	for p.curTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	// Expect 'until' keyword
	if !p.curTokenIs(lexer.UNTIL) {
		p.addError(fmt.Sprintf("expected 'until' after repeat body, got %s instead", p.curToken.Type))
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
// Syntax:
//
//	for <variable> := <start> to|downto <end> do <statement>
//	for [var] <variable> in <expression> do <statement>
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

	variable := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check if this is a for-in loop (IN) or for-to/downto loop (:=)
	if p.peekTokenIs(lexer.IN) {
		// Parse for-in loop: for [var] x in collection do statement
		return p.parseForInLoop(forToken, variable, inlineVar)
	}

	// Parse traditional for-to/downto loop
	stmt := &ast.ForStatement{Token: forToken, Variable: variable, InlineVar: inlineVar}

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

// parseForInLoop parses a for-in loop statement.
// Syntax: for [var] <variable> in <expression> do <statement>
func (p *Parser) parseForInLoop(forToken lexer.Token, variable *ast.Identifier, inlineVar bool) *ast.ForInStatement {
	stmt := &ast.ForInStatement{
		Token:     forToken,
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
		p.addError("expected expression after 'in'")
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
