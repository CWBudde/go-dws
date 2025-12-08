package parser

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// PRE: cursor is first token of expression
// POST: cursor is last token of expression
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// 1. Lookup and call prefix function
	currentToken := p.cursor.Current()
	prefixFn, ok := p.prefixParseFns[currentToken.Type]
	if !ok {
		p.noPrefixParseFnError(currentToken.Type)
		return nil
	}
	leftExp := prefixFn(currentToken)

	// 2. Main precedence climbing loop
	for {
		nextToken := p.cursor.Peek(1)

		// Termination condition 1: semicolon
		if nextToken.Type == lexer.SEMICOLON {
			break
		}

		// Get next token's precedence
		nextPrec := getPrecedence(nextToken.Type)

		// Termination condition 2: precedence
		// Special case: allow NOT at EQUALS precedence for "not in/is/as"
		if precedence >= nextPrec && (nextToken.Type != lexer.NOT || precedence >= EQUALS) {
			break
		}

		// 3. Special case: "not in/is/as"
		if nextToken.Type == lexer.NOT && precedence < EQUALS {
			leftExp = p.parseNotInIsAs(leftExp)
			if leftExp == nil {
				// Not a "not in/is/as" pattern, return what we have
				break
			}
			continue
		}

		// 4. Normal infix handling
		infixFn, ok := p.infixParseFns[nextToken.Type]
		if !ok {
			// No infix handler for this token type, stop parsing
			break
		}

		// Advance to operator
		p.cursor = p.cursor.Advance()
		operatorToken := p.cursor.Current()

		// All registered infix cursor functions now use parseInfixExpression,
		// which is pure cursor and recursively calls parseExpression
		// Call infix function
		leftExp = infixFn(leftExp, operatorToken)
	}

	// Sync cursor position back to curToken/peekToken for backward compatibility
	// External code (like parseIfStatement) uses curToken/peekToken, not cursor

	return leftExp
}

// Returns the wrapped NOT expression if successful, or nil if this is not a "not in/is/as" pattern.
func (p *Parser) parseNotInIsAs(leftExp ast.Expression) ast.Expression {
	// Mark current position for potential backtracking
	mark := p.cursor.Mark()

	// Advance to NOT token
	p.cursor = p.cursor.Advance()
	notToken := p.cursor.Current()

	// Check if next token is IN, IS, or AS
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.IN && nextToken.Type != lexer.IS && nextToken.Type != lexer.AS {
		// Not a "not in/is/as" pattern, backtrack
		p.cursor = p.cursor.ResetTo(mark)
		return nil
	}

	// This is "not in", "not is", or "not as"
	// Advance to IN/IS/AS token
	p.cursor = p.cursor.Advance()
	operatorToken := p.cursor.Current()

	// Look up infix function for the operator
	infixFn, ok := p.infixParseFns[operatorToken.Type]
	if !ok {
		// No infix function, backtrack
		p.cursor = p.cursor.ResetTo(mark)
		return nil
	}

	// Now that parseInfixExpression is pure cursor, no sync needed
	// Parse the comparison expression
	comparisonExp := infixFn(leftExp, operatorToken)

	// Wrap in NOT expression
	notExp := &ast.UnaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  notToken,
				EndPos: comparisonExp.End(),
			},
		},
		Operator: notToken.Literal,
		Right:    comparisonExp,
	}

	return notExp
}

// parseIdentifier parses an identifier.
// POST: cursor is IDENT (unchanged)
func (p *Parser) parseIdentifier() ast.Expression {
	currentToken := p.cursor.Current()
	return &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  currentToken,
				EndPos: p.endPosFromToken(currentToken),
			},
		},
		Value: currentToken.Literal,
	}
}

// parsePrefixExpression parses a prefix (unary) expression: -x, +x, not x
// PRE: cursor is prefix operator (NOT, MINUS, PLUS, etc.)
// POST: cursor is last token of right operand
func (p *Parser) parsePrefixExpression() ast.Expression {
	builder := p.StartNode()
	operatorToken := p.cursor.Current()

	expression := &ast.UnaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: operatorToken,
			},
		},
		Operator: operatorToken.Literal,
	}

	// Advance to operand
	p.cursor = p.cursor.Advance()

	// Parse the operand expression
	expression.Right = p.parseExpression(PREFIX)

	// End at right expression (FinishWithNode handles nil by falling back to current token)
	return builder.FinishWithNode(expression, expression.Right).(ast.Expression)
}

// parseInfixExpression parses a binary infix expression (dispatcher).
// PRE: cursor is the operator token
// POST: cursor is last token of right expression
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	builder := p.StartNode()
	operatorToken := p.cursor.Current()

	expression := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: operatorToken,
			},
		},
		Operator: operatorToken.Literal,
		Left:     left,
	}

	// Get precedence based on operator token type
	precedence := LOWEST
	if prec, ok := precedences[operatorToken.Type]; ok {
		precedence = prec
	}

	// Advance cursor to next token (the start of right expression)
	p.cursor = p.cursor.Advance()

	// Now that parseExpression is implemented, we can call it directly
	// for pure cursor-to-cursor recursion without state synchronization
	expression.Right = p.parseExpression(precedence)

	// End at right expression (FinishWithNode handles nil by falling back to current token)
	return builder.FinishWithNode(expression, expression.Right).(ast.Expression)
}

// parseExpressionList parses a comma-separated list of expressions.
// PRE: cursor is before the list (at opening delimiter)
// POST: cursor is at terminator (closing delimiter)
func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	// Check for empty list
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == end {
		p.cursor = p.cursor.Advance() // consume terminator
		return list
	}

	// Advance to first item
	p.cursor = p.cursor.Advance()

	// Parse first expression
	expr := p.parseExpression(LOWEST)
	if expr != nil {
		list = append(list, expr)
	}

	// Parse remaining expressions (separated by commas)
	for {
		nextToken = p.cursor.Peek(1)

		// Check for terminator
		if nextToken.Type == end {
			p.cursor = p.cursor.Advance() // consume terminator
			break
		}

		// Check for comma separator
		if nextToken.Type == lexer.COMMA {
			p.cursor = p.cursor.Advance() // consume comma

			// Check for trailing comma before terminator
			nextToken = p.cursor.Peek(1)
			if nextToken.Type == end {
				p.cursor = p.cursor.Advance() // consume terminator
				break
			}

			// Advance to next expression
			p.cursor = p.cursor.Advance()

			// Parse next expression
			expr = p.parseExpression(LOWEST)
			if expr != nil {
				list = append(list, expr)
			}
		} else {
			// Unexpected token - no separator found
			// Add error and break
			p.addError(fmt.Sprintf("expected ',' or '%s', got %s", end, nextToken.Type), ErrUnexpectedToken)
			break
		}
	}

	return list
}

// parseGroupedExpression parses a grouped expression (parentheses).
// Also handles:
//   - Record literals: (X: 10, Y: 20)
//   - Array literals: (1, 2, 3)
//
// PRE: cursor is LPAREN
// POST: cursor is RPAREN
func (p *Parser) parseGroupedExpression() ast.Expression {
	lparenToken := p.cursor.Current()

	// Handle empty parentheses: () -> empty array literal
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.RPAREN {
		p.cursor = p.cursor.Advance() // move to RPAREN
		return &ast.ArrayLiteralExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token:  lparenToken,
					EndPos: p.cursor.Current().End(),
				},
			},
			Elements: []ast.Expression{},
		}
	}

	// Check if this is a record literal: (IDENT : ...)
	secondToken := p.cursor.Peek(2)
	if nextToken.Type == lexer.IDENT && secondToken.Type == lexer.COLON {
		// Parse record literal inline
		recordLit := &ast.RecordLiteralExpression{
			BaseNode: ast.BaseNode{Token: lparenToken},
			TypeName: nil, // Anonymous record
			Fields:   []*ast.FieldInitializer{},
		}

		// Move to first field name
		p.cursor = p.cursor.Advance()

		// Parse fields in a loop
		for p.cursor.Current().Type != lexer.RPAREN && p.cursor.Current().Type != lexer.EOF {
			field := p.parseNamedFieldInitializer()
			if field == nil {
				return nil
			}
			recordLit.Fields = append(recordLit.Fields, field)

			// Check if we should continue to next field
			shouldContinue, ok := p.advanceToNextItem(lexer.RPAREN)
			if !ok {
				return nil
			}
			if !shouldContinue {
				break
			}
		}

		recordLit.EndPos = p.cursor.Current().End()
		return recordLit
	}

	// Move to first expression
	p.cursor = p.cursor.Advance()

	// Parse first expression
	exp := p.parseExpression(LOWEST)
	if exp == nil {
		return nil
	}

	// Check if this is an array literal: (expr, expr, ...)
	nextToken = p.cursor.Peek(1)
	if nextToken.Type == lexer.COMMA {
		// Parse array literal inline
		elements := []ast.Expression{exp}

		// Parse remaining elements
		for p.cursor.Peek(1).Type == lexer.COMMA {
			p.cursor = p.cursor.Advance() // move to COMMA
			p.cursor = p.cursor.Advance() // move to next element or RPAREN

			// Allow trailing comma: (1, 2, )
			if p.cursor.Current().Type == lexer.RPAREN {
				return &ast.ArrayLiteralExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token:  lparenToken,
							EndPos: p.cursor.Current().End(),
						},
					},
					Elements: elements,
				}
			}

			elementExpr := p.parseExpression(LOWEST)
			if elementExpr == nil {
				return nil
			}
			elements = append(elements, elementExpr)
		}

		// Expect closing paren
		if p.cursor.Peek(1).Type != lexer.RPAREN {
			p.addError(fmt.Sprintf("expected ')', got %s", p.cursor.Peek(1).Type), ErrUnexpectedToken)
			return nil
		}

		p.cursor = p.cursor.Advance() // move to RPAREN

		return &ast.ArrayLiteralExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token:  lparenToken,
					EndPos: p.cursor.Current().End(),
				},
			},
			Elements: elements,
		}
	}

	// Expect closing paren
	if nextToken.Type != lexer.RPAREN {
		p.addError(fmt.Sprintf("expected ')', got %s", nextToken.Type), ErrUnexpectedToken)
		return nil
	}

	// Advance to RPAREN
	p.cursor = p.cursor.Advance()

	// Return the expression directly, not wrapped
	// This avoids double parentheses in the string representation
	return exp
}
