package parser

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// parseIsExpression parses the 'is' operator which can be used for:
// 1. Type checking: obj is TMyClass
// 2. Boolean value comparison: boolExpr is True, boolExpr is False
// This creates an IsExpression AST node that will be evaluated at runtime.
// Example: obj is TClass  -> Boolean
// PRE: cursor is on IS token
// POST: cursor is on last token of type/expression
func (p *Parser) parseIsExpression(left ast.Expression) ast.Expression {
	builder := p.StartNode()
	isToken := p.cursor.Current()
	expression := &ast.IsExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: isToken},
		},
		Left: left,
	}

	p.cursor = p.cursor.Advance()

	// Try to parse as type expression first (speculatively)
	// Save full parser state including errors for clean backtracking
	state := p.saveState()
	expression.TargetType = p.parseTypeExpression()
	if expression.TargetType != nil {
		return builder.FinishWithNode(expression, expression.TargetType).(ast.Expression)
	}

	// If type parsing failed, restore full state (errors + cursor) and try as boolean expression
	// Note: cursor is already positioned at the token after IS from the saved state
	p.restoreState(state)

	// Parse as value expression (boolean comparison)
	// Use EQUALS precedence to prevent consuming following logical operators
	expression.Right = p.parseExpression(EQUALS)
	if expression.Right == nil {
		p.addError("expected expression after 'is' operator", ErrInvalidExpression)
		return expression
	}
	return builder.FinishWithNode(expression, expression.Right).(ast.Expression)
}

// parseAsExpression parses the 'as' type casting operator.
// Example: obj as IMyInterface
// This creates an AsExpression AST node that will be evaluated at runtime
// to wrap an object instance in an InterfaceInstance.
// PRE: cursor is on AS token
// POST: cursor is on last token of target type
func (p *Parser) parseAsExpression(left ast.Expression) ast.Expression {
	builder := p.StartNode()
	asToken := p.cursor.Current()
	expression := &ast.AsExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: asToken},
		},
		Left: left,
	}

	p.cursor = p.cursor.Advance()

	// Parse the target type (should be an interface type)
	expression.TargetType = p.parseTypeExpression()
	if expression.TargetType == nil {
		p.addError("expected type after 'as' operator", ErrExpectedType)
		return expression
	}

	return builder.FinishWithNode(expression, expression.TargetType).(ast.Expression)
}

// parseImplementsExpression parses the 'implements' operator.
// Example: obj implements IMyInterface  -> Boolean
// This creates an ImplementsExpression AST node that will be evaluated
// to check whether the object's class implements the interface.
// PRE: cursor is on IMPLEMENTS token
// POST: cursor is on last token of target type
func (p *Parser) parseImplementsExpression(left ast.Expression) ast.Expression {
	builder := p.StartNode()
	implementsToken := p.cursor.Current()
	expression := &ast.ImplementsExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: implementsToken},
		},
		Left: left,
	}

	p.cursor = p.cursor.Advance()

	// Parse the target type (should be an interface type)
	expression.TargetType = p.parseTypeExpression()
	if expression.TargetType == nil {
		p.addError("expected type after 'implements' operator", ErrExpectedType)
		return expression
	}

	return builder.FinishWithNode(expression, expression.TargetType).(ast.Expression)
}
