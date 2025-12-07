package parser

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// parseSelfExpression parses a self expression.
// The Self keyword refers to the current instance (in instance methods) or
// the current class (in class methods).
// Usage: Self, Self.Field, Self.Method()
// PRE: cursor is on SELF token
// POST: cursor unchanged (SELF token only)
func (p *Parser) parseSelfExpression() ast.Expression {
	builder := p.StartNode()
	currentToken := p.cursor.Current()
	selfExpr := &ast.SelfExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: currentToken, // The 'self' keyword
			},
		},
		Token: currentToken,
	}

	// Set end position at the Self keyword itself
	return builder.Finish(selfExpr).(ast.Expression)
}

// parseInheritedExpression parses an inherited expression.
// Supports three forms:
//   - inherited                  // Bare inherited (calls same method in parent)
//   - inherited MethodName       // Call parent method (no args)
//   - inherited MethodName(args) // Call parent method with args
//
// DWScript syntax: inherited [Method[(args)]]
// PRE: cursor is on INHERITED token
// POST: cursor is on last token of expression
func (p *Parser) parseInheritedExpression() ast.Expression {
	builder := p.StartNode()
	currentToken := p.cursor.Current()
	inheritedExpr := &ast.InheritedExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: currentToken, // The 'inherited' keyword
			},
		},
	}

	// Check if there's a method name following
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.IDENT {
		p.cursor = p.cursor.Advance() // move to identifier
		methodToken := p.cursor.Current()
		inheritedExpr.Method = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: methodToken,
				},
			},
			Value: methodToken.Literal,
		}
		inheritedExpr.IsMember = true

		// Check if there's a call (parentheses)
		nextToken = p.cursor.Peek(1)
		if nextToken.Type == lexer.LPAREN {
			p.cursor = p.cursor.Advance() // move to '('
			inheritedExpr.IsCall = true

			// Parse arguments
			inheritedExpr.Arguments = p.parseExpressionList(lexer.RPAREN)
			// Set end position after closing parenthesis (cursor is now at RPAREN)
			return builder.Finish(inheritedExpr).(ast.Expression)
		} else {
			// No call, just method name - end at method identifier
			return builder.FinishWithNode(inheritedExpr, inheritedExpr.Method).(ast.Expression)
		}
	} else {
		// Bare 'inherited' keyword - end at the keyword itself
		return builder.Finish(inheritedExpr).(ast.Expression)
	}
}

// parseNewExpression parses a new expression for both classes and arrays.
// DWScript syntax:
//   - new ClassName(args)     // Class instantiation
//   - new TypeName[size]      // Array instantiation (1D)
//   - new TypeName[s1, s2]    // Array instantiation (multi-dimensional)
//
// This function dispatches to the appropriate parser based on the token
// following the type name: '(' for classes, '[' for arrays.
// PRE: cursor is on NEW token
// POST: cursor is on last token of expression
func (p *Parser) parseNewExpression() ast.Expression {
	newToken := p.cursor.Current() // Save the 'new' token position

	// Expect a type name (identifier)
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.IDENT {
		p.addError("expected type name after 'new'", ErrExpectedIdent)
		return nil
	}

	p.cursor = p.cursor.Advance() // move to identifier
	typeToken := p.cursor.Current()
	parts := []string{typeToken.Literal}
	// Support qualified class names for nested classes: new TOuter.TInner(...)
	// But stop if the dot would actually start a member access on the constructed instance
	// (e.g., "new TObj.FField" should be parsed as (new TObj).FField).
	for {
		if p.cursor.Peek(1).Type != lexer.DOT || p.cursor.Peek(2).Type != lexer.IDENT {
			break
		}
		// Look ahead to see what follows the potential qualified identifier.
		// If it's another dot (chained nested type) or an argument/array list,
		// treat the dot as part of the type name. Otherwise, leave it for
		// regular member access parsing.
		afterIdent := p.cursor.Peek(3)
		if afterIdent.Type != lexer.DOT && afterIdent.Type != lexer.LPAREN && afterIdent.Type != lexer.LBRACK {
			break
		}

		p.cursor = p.cursor.Advance() // move to '.'
		p.cursor = p.cursor.Advance() // move to next ident
		typeToken = p.cursor.Current()
		parts = append(parts, typeToken.Literal)
	}
	typeName := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: typeToken,
			},
		},
		Value: strings.Join(parts, "."),
	}

	// Check what follows: '(' for class, '[' for array, or nothing for zero-arg constructor
	nextToken = p.cursor.Peek(1)
	switch nextToken.Type {
	case lexer.LBRACK:
		// Array instantiation: new TypeName[size, ...]
		return p.parseNewArrayExpression(newToken, typeName)
	case lexer.LPAREN:
		// Class instantiation: new ClassName(args)
		return p.parseNewClassExpression(newToken, typeName)
	default:
		// No parentheses - treat as zero-argument constructor
		// DWScript allows: new TTest (equivalent to new TTest())
		return &ast.NewExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: newToken,
				},
			},
			ClassName: typeName,
			Arguments: []ast.Expression{},
		}
	}
}

// parseNewClassExpression parses class instantiation: new ClassName(args)
// PRE: cursor is on className IDENT
// POST: cursor is on RPAREN
func (p *Parser) parseNewClassExpression(newToken lexer.Token, className *ast.Identifier) ast.Expression {
	// Create NewExpression
	newExpr := &ast.NewExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: newToken,
			},
		},
		ClassName: className,
	}

	// Move to LPAREN
	p.cursor = p.cursor.Advance()

	// Parse constructor arguments
	newExpr.Arguments = p.parseExpressionList(lexer.RPAREN)

	return newExpr
}

// parseNewArrayExpression parses array instantiation: new TypeName[size1, size2, ...]
// Supports both single-dimensional and multi-dimensional arrays.
// Examples:
//   - new Integer[16]
//   - new String[10, 20]
//   - new Float[Length(arr)+1]
//
// PRE: cursor is on type name identifier
// POST: cursor is on RBRACK
func (p *Parser) parseNewArrayExpression(newToken lexer.Token, elementTypeName *ast.Identifier) ast.Expression {
	// Move to '['
	p.cursor = p.cursor.Advance()

	// Parse dimension sizes (comma-separated)
	dimensions, ok := p.parseArrayDimensions(lexer.RBRACK)
	if !ok {
		return nil
	}

	return &ast.NewArrayExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: newToken,
			},
		},
		ElementTypeName: elementTypeName,
		Dimensions:      dimensions,
	}
}

// parseArrayDimensions parses the dimension list for a 'new' array expression.
// It disallows empty brackets and trailing commas to ensure each dimension
// has a corresponding expression.
// PRE: cursor is on '['
// POST: cursor is on ']'
func (p *Parser) parseArrayDimensions(end lexer.TokenType) ([]ast.Expression, bool) {
	dimensions := []ast.Expression{}

	// Empty brackets are not allowed
	if p.cursor.Peek(1).Type == end {
		p.addError("expected expression for array dimension", ErrInvalidExpression)
		p.cursor = p.cursor.Advance() // consume ']'
		return dimensions, false
	}

	// Move to first dimension
	p.cursor = p.cursor.Advance()

	for {
		// Parse dimension expression
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			return dimensions, false
		}
		dimensions = append(dimensions, expr)

		nextToken := p.cursor.Peek(1)
		switch nextToken.Type {
		case lexer.COMMA:
			p.cursor = p.cursor.Advance() // move to ','

			// Trailing comma before closing bracket is invalid
			if p.cursor.Peek(1).Type == end {
				p.addError("expected expression for array dimension", ErrInvalidExpression)
				p.cursor = p.cursor.Advance() // consume ']'
				return dimensions, false
			}

			p.cursor = p.cursor.Advance() // move to next dimension

		case end:
			p.cursor = p.cursor.Advance() // consume ']'
			return dimensions, true

		default:
			p.addError(fmt.Sprintf("expected ',' or '%s', got %s", end, nextToken.Type), ErrUnexpectedToken)
			return dimensions, false
		}
	}
}

// parseDefaultExpression parses a Default() call expression.
// DWScript syntax: Default(TypeName) - returns the default/zero value for the type
// PRE: cursor is on DEFAULT token
// POST: cursor is on RPAREN
func (p *Parser) parseDefaultExpression() ast.Expression {
	defaultToken := p.cursor.Current() // Save the 'default' token position

	// Expect LPAREN
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.LPAREN {
		p.addError("expected '(' after 'default'", ErrUnexpectedToken)
		return nil
	}
	p.cursor = p.cursor.Advance() // move to '('

	// Parse the type name argument
	p.cursor = p.cursor.Advance() // Move to type name

	// The type name could be an identifier (Integer, String, etc.)
	typeName := p.parseExpression(LOWEST)
	if typeName == nil {
		return nil
	}

	// Expect RPAREN
	nextToken = p.cursor.Peek(1)
	if nextToken.Type != lexer.RPAREN {
		p.addError("expected ')' after type name", ErrUnexpectedToken)
		return nil
	}
	p.cursor = p.cursor.Advance() // move to ')'

	// Return as a CallExpression with function name "Default"
	return &ast.CallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: defaultToken,
			},
		},
		Function: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: defaultToken,
				},
			},
			Value: "Default",
		},
		Arguments: []ast.Expression{typeName},
	}
}

// parseAddressOfExpression parses the address-of operator (@) applied to a function or procedure.
// Examples: @MyFunction, @TMyClass.MyMethod
// DWScript syntax: @variable or @function
// PRE: cursor is on AT token
// POST: cursor is on last token of target expression
func (p *Parser) parseAddressOfExpression() ast.Expression {
	builder := p.StartNode()
	currentToken := p.cursor.Current()
	expression := &ast.AddressOfExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: currentToken}, // The @ token
		},
	}

	p.cursor = p.cursor.Advance() // advance to the target

	// Parse the target expression (function/procedure name or member access)
	expression.Operator = p.parseExpression(PREFIX)

	// End at operator expression (FinishWithNode handles nil by falling back to current token)
	return builder.FinishWithNode(expression, expression.Operator).(ast.Expression)
}
