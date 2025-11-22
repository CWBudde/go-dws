package parser

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// parseLambdaExpression parses a lambda/anonymous function expression.
// Supports both full and shorthand syntax:
//   - Full: lambda(x: Integer): Integer begin Result := x * 2; end
//   - Shorthand: lambda(x) => x * 2
//
// PRE: cursor is on LAMBDA token
// POST: cursor is on last token of lambda body (END for full syntax, expression for shorthand)
func (p *Parser) parseLambdaExpression() ast.Expression {
	builder := p.StartNode()
	currentToken := p.cursor.Current()
	lambdaExpr := &ast.LambdaExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: currentToken}, // The 'lambda' keyword
		},
	}

	// Check for opening parenthesis (optional for zero-parameter lambdas)
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.LPAREN {
		p.cursor = p.cursor.Advance() // move to '('

		// Parse parameter list (may be empty)
		lambdaExpr.Parameters = p.parseLambdaParameterList()

		// Check for return type annotation (optional)
		nextToken = p.cursor.Peek(1)
		if nextToken.Type == lexer.COLON {
			p.cursor = p.cursor.Advance() // move to ':'

			// Parse return type
			nextToken = p.cursor.Peek(1)
			if nextToken.Type != lexer.IDENT {
				p.addError("expected return type after ':'", ErrExpectedType)
				return nil
			}
			p.cursor = p.cursor.Advance() // move to type
			typeToken := p.cursor.Current()

			lambdaExpr.ReturnType = &ast.TypeAnnotation{
				Token: typeToken,
				Name:  typeToken.Literal,
			}
		}
	} else {
		// No parentheses - zero-parameter lambda
		// Parameters remain empty (nil)
		lambdaExpr.Parameters = []*ast.Parameter{}
	}

	// Check which syntax is being used: shorthand (=>) or full (begin/end)
	nextToken = p.cursor.Peek(1)
	switch nextToken.Type {
	case lexer.FAT_ARROW:
		// Shorthand syntax: lambda(x) => expression
		p.cursor = p.cursor.Advance() // move to '=>'
		p.cursor = p.cursor.Advance() // move past '=>' to expression

		// Parse the expression
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			p.addError("expected expression after '=>'", ErrInvalidExpression)
			return nil
		}

		// Desugar shorthand to full syntax: wrap expression in return statement
		lambdaExpr.Body = &ast.BlockStatement{
			BaseNode: ast.BaseNode{Token: p.cursor.Current()}, // Use current token for position tracking
			Statements: []ast.Statement{
				&ast.ReturnStatement{
					BaseNode: ast.BaseNode{
						Token: p.cursor.Current(),
					},
					ReturnValue: expr,
				},
			},
		}
		lambdaExpr.IsShorthand = true

		// Set end position based on expression
		if expr != nil {
			return builder.FinishWithNode(lambdaExpr, expr).(ast.Expression)
		} else {
			return builder.Finish(lambdaExpr).(ast.Expression)
		}

	case lexer.BEGIN:
		// Full syntax: lambda(x: Integer) begin ... end
		p.cursor = p.cursor.Advance() // move to 'begin'

		// Parse block statement
		lambdaExpr.Body = p.parseBlockStatement()
		lambdaExpr.IsShorthand = false

		// Set end position based on body block
		if lambdaExpr.Body != nil {
			return builder.FinishWithNode(lambdaExpr, lambdaExpr.Body).(ast.Expression)
		} else {
			return builder.Finish(lambdaExpr).(ast.Expression)
		}

	default:
		// Full syntax without 'begin': lambda result := value end
		// Parse statements until we hit 'end'
		p.cursor = p.cursor.Advance() // move to first token of body

		statements := []ast.Statement{}
		for p.cursor.Current().Type != lexer.END && p.cursor.Current().Type != lexer.EOF {
			stmt := p.parseStatement()
			if stmt != nil {
				statements = append(statements, stmt)
			}

			// Check if we're at the end
			if p.cursor.Current().Type == lexer.END || p.cursor.Current().Type == lexer.EOF {
				break
			}

			// If not at end, skip optional semicolon
			if p.cursor.Current().Type == lexer.SEMICOLON {
				p.cursor = p.cursor.Advance()
			}
		}

		if p.cursor.Current().Type != lexer.END {
			p.addError("expected 'end' to close lambda body", ErrUnexpectedToken)
			return nil
		}

		// Create block statement from parsed statements
		lambdaExpr.Body = &ast.BlockStatement{
			BaseNode:   ast.BaseNode{Token: currentToken},
			Statements: statements,
		}
		lambdaExpr.IsShorthand = false

		// Set end position to the 'end' keyword
		return builder.Finish(lambdaExpr).(ast.Expression)
	}
}

// parseLambdaParameterList parses the parameter list for a lambda expression.
// Lambda parameters follow the same syntax as function parameters:
//   - Semicolon-separated groups: lambda(x: Integer; y: Integer)
//   - Comma-separated names with shared type: lambda(x, y: Integer)
//   - Mixed groups: lambda(x, y: Integer; z: String)
//   - Supports by-ref: lambda(var x: Integer; y: Integer)
//
// Note: Lambda parameters use semicolons between groups, matching DWScript function syntax.
// PRE: cursor is on LPAREN
// POST: cursor is on RPAREN
func (p *Parser) parseLambdaParameterList() []*ast.Parameter {
	params := []*ast.Parameter{}

	// Check if empty parameter list
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.RPAREN {
		p.cursor = p.cursor.Advance() // move to ')'
		return params
	}

	// Parse first parameter group
	p.cursor = p.cursor.Advance() // move past '(' to first parameter
	groupParams := p.parseParameterGroup()
	params = append(params, groupParams...)

	// Parse additional parameter groups separated by semicolons
	for {
		nextToken = p.cursor.Peek(1)
		if nextToken.Type == lexer.EOF {
			break
		}
		if nextToken.Type == lexer.SEMICOLON {
			p.cursor = p.cursor.Advance() // move to ';'
			p.cursor = p.cursor.Advance() // move past ';' to next parameter
			groupParams = p.parseParameterGroup()
			params = append(params, groupParams...)
		} else if nextToken.Type == lexer.RPAREN {
			p.cursor = p.cursor.Advance() // move to ')'
			break
		} else {
			p.addError(fmt.Sprintf("expected ';' or ')' in parameter list, got %s", nextToken.Literal), ErrUnexpectedToken)
			break
		}
	}

	return params
}
