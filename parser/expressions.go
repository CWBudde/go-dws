package parser

import (
	"fmt"
	"strconv"

	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
)

// parseExpression parses an expression with the given precedence.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

// parseIdentifier parses an identifier.
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral parses an integer literal.
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.addError(msg)
		return nil
	}

	lit.Value = value
	return lit
}

// parseFloatLiteral parses a floating-point literal.
func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.addError(msg)
		return nil
	}

	lit.Value = value
	return lit
}

// parseStringLiteral parses a string literal.
func (p *Parser) parseStringLiteral() ast.Expression {
	// The lexer has already processed the string, so we just need to
	// extract the value without the quotes
	value := p.curToken.Literal

	// Remove surrounding quotes
	if len(value) >= 2 {
		if (value[0] == '\'' && value[len(value)-1] == '\'') ||
			(value[0] == '"' && value[len(value)-1] == '"') {
			value = value[1 : len(value)-1]
		}
	}

	// Handle escaped quotes ('' -> ')
	value = unescapeString(value)

	return &ast.StringLiteral{Token: p.curToken, Value: value}
}

// unescapeString handles DWScript string escape sequences.
func unescapeString(s string) string {
	result := ""
	i := 0
	for i < len(s) {
		if i < len(s)-1 && s[i] == '\'' && s[i+1] == '\'' {
			result += "'"
			i += 2
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}

// parseBooleanLiteral parses a boolean literal.
func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(lexer.TRUE)}
}

// parseNilLiteral parses a nil literal.
func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{Token: p.curToken}
}

// parsePrefixExpression parses a prefix (unary) expression.
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.UnaryExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseInfixExpression parses an infix (binary) expression.
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.BinaryExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseCallExpression parses a function call expression.
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{
		Token:    p.curToken,
		Function: function,
	}

	exp.Arguments = p.parseExpressionList(lexer.RPAREN)

	return exp
}

// parseExpressionList parses a comma-separated list of expressions.
func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if exp != nil {
		list = append(list, exp)
	}

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // move to comma
		if p.peekTokenIs(end) {
			p.nextToken()
			return list
		}
		p.nextToken() // move to next expression
		exp := p.parseExpression(LOWEST)
		if exp != nil {
			list = append(list, exp)
		}
	}

	if !p.expectPeek(end) {
		return list
	}

	return list
}

// parseGroupedExpression parses a grouped expression (parentheses).
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken() // skip '('

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	// Return the expression directly, not wrapped in GroupedExpression
	// This avoids double parentheses in the string representation
	return exp
}
