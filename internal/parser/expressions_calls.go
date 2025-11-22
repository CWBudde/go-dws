package parser

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// parseCallExpression parses a function call expression.
// Also handles typed record literals: TypeName(field: value)
// PRE: cursor is LPAREN
// POST: cursor is RPAREN
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	// Check if this might be a typed record literal
	// Pattern: Identifier(Identifier:Expression, ...)
	if ident, ok := function.(*ast.Identifier); ok {
		// Parse the arguments, but check if they're all colon-based field initializers
		return p.parseCallOrRecordLiteral(ident)
	}

	// Normal function call (non-identifier function)
	builder := p.StartNode()
	lparenToken := p.cursor.Current()

	exp := &ast.CallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: lparenToken},
		},
		Function: function,
	}

	exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	return builder.Finish(exp).(ast.Expression) // cursor is now at RPAREN
}

// parseCallOrRecordLiteral parses either a function call or a typed record literal.
// They have the same syntax initially: Identifier(...)
// The difference is whether the arguments are field initializers (name: value) or expressions.
// PRE: cursor is LPAREN
// POST: cursor is RPAREN
func (p *Parser) parseCallOrRecordLiteral(typeName *ast.Identifier) ast.Expression {
	// Empty parentheses -> function call
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.RPAREN {
		return p.parseEmptyCall(typeName)
	}

	// Non-identifier first element -> must be function call
	if nextToken.Type != lexer.IDENT {
		return p.parseCallWithExpressionList(typeName)
	}

	// We have: TypeName(IDENT ...
	// Parse arguments/fields and determine type based on whether ALL have colons
	items, allHaveColons := p.parseArgumentsOrFields(lexer.RPAREN)

	if allHaveColons {
		// All items were field initializers -> record literal
		return p.buildRecordLiteral(typeName, items)
	}

	// Some or no items had colons -> function call
	return p.buildCallExpressionFromFields(typeName, items)
}

// parseEmptyCall creates a call expression with no arguments.
// PRE: cursor.Peek(1) is RPAREN
// POST: cursor is at RPAREN
func (p *Parser) parseEmptyCall(typeName *ast.Identifier) *ast.CallExpression {
	builder := p.StartNode()
	// Advance to RPAREN
	p.cursor = p.cursor.Advance()
	rparenToken := p.cursor.Current()

	exp := &ast.CallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: rparenToken,
			},
		},
		Function:  typeName,
		Arguments: []ast.Expression{},
	}
	return builder.Finish(exp).(*ast.CallExpression)
}

// parseCallWithExpressionList parses a function call using the expression list parser.
// PRE: cursor is at LPAREN, cursor.Peek(1) is not RPAREN
// POST: cursor is at RPAREN
func (p *Parser) parseCallWithExpressionList(typeName *ast.Identifier) *ast.CallExpression {
	builder := p.StartNode()
	lparenToken := p.cursor.Current()

	exp := &ast.CallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lparenToken,
			},
		},
		Function: typeName,
	}

	// Parse argument list using cursor version
	exp.Arguments = p.parseExpressionList(lexer.RPAREN)

	// Set end position to RPAREN
	return builder.Finish(exp).(*ast.CallExpression)
}

// buildRecordLiteral creates a record literal expression from field initializers.
func (p *Parser) buildRecordLiteral(typeName *ast.Identifier, fields []*ast.FieldInitializer) *ast.RecordLiteralExpression {
	return &ast.RecordLiteralExpression{
		BaseNode: ast.BaseNode{Token: p.cursor.Current()},
		TypeName: typeName,
		Fields:   fields,
	}
}

// buildCallExpressionFromFields creates a call expression by extracting arguments from field initializers.
// Handles the case where some items might have names (which shouldn't happen, but we're defensive).
func (p *Parser) buildCallExpressionFromFields(typeName *ast.Identifier, items []*ast.FieldInitializer) *ast.CallExpression {
	args := make([]ast.Expression, len(items))
	for i, item := range items {
		if item.Name != nil {
			// Shouldn't happen if allHaveColons is false, but handle defensively
			args[i] = item.Name
		} else {
			args[i] = item.Value
		}
	}

	return &ast.CallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: p.cursor.Current()},
		},
		Function:  typeName,
		Arguments: args,
	}
}

// parseArgumentsOrFields parses a list that could be either function arguments or record fields.
// Returns the parsed items and whether ALL of them were colon-based fields.
// PRE: cursor is on LPAREN
// POST: cursor is on end token
func (p *Parser) parseArgumentsOrFields(end lexer.TokenType) ([]*ast.FieldInitializer, bool) {
	var items []*ast.FieldInitializer
	allHaveColons := true

	// Check for empty list
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == end {
		p.cursor = p.cursor.Advance() // consume end token
		return items, true            // empty list
	}

	// Move to first element
	p.cursor = p.cursor.Advance()

	for {
		// Parse either a field initializer (name: value) or plain expression
		item, hasColon := p.parseSingleArgumentOrField()
		if item == nil {
			return items, false
		}

		if !hasColon {
			allHaveColons = false
		}

		items = append(items, item)

		// Check if we should continue to next item
		shouldContinue, ok := p.advanceToNextItem(end)
		if !ok {
			return items, false
		}
		if !shouldContinue {
			break
		}
	}

	return items, allHaveColons
}

// parseNamedFieldInitializer parses a field initializer: name : value
// PRE: cursor is at IDENT, cursor.Peek(1) is COLON
// POST: cursor is at value expression
func (p *Parser) parseNamedFieldInitializer() *ast.FieldInitializer {
	identToken := p.cursor.Current()

	fieldName := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  identToken,
				EndPos: p.endPosFromToken(identToken),
			},
		},
		Value: identToken.Literal,
	}

	// Advance to COLON
	p.cursor = p.cursor.Advance()

	// Advance to value
	p.cursor = p.cursor.Advance()

	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}

	return &ast.FieldInitializer{
		BaseNode: ast.BaseNode{
			Token:  fieldName.Token,
			EndPos: value.End(),
		},
		Name:  fieldName,
		Value: value,
	}
}

// parseArgumentAsFieldInitializer parses a plain expression as a field initializer (without name).
// Used to represent function arguments in the same data structure as record fields.
// PRE: cursor is at start of expression
// POST: cursor is at end of expression
func (p *Parser) parseArgumentAsFieldInitializer() *ast.FieldInitializer {
	exprStart := p.cursor.Current()

	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}

	return &ast.FieldInitializer{
		BaseNode: ast.BaseNode{
			Token:  exprStart,
			EndPos: expr.End(),
		},
		Name:  nil, // no name means regular argument
		Value: expr,
	}
}

// parseSingleArgumentOrField parses either a field initializer (name: value) or plain expression.
// Returns the item and whether it had a colon (i.e., was a field initializer).
// PRE: cursor is at start of argument/field
// POST: cursor is at end of argument/field
func (p *Parser) parseSingleArgumentOrField() (*ast.FieldInitializer, bool) {
	currentToken := p.cursor.Current()
	nextToken := p.cursor.Peek(1)

	// Check for field initializer pattern: IDENT COLON
	if currentToken.Type == lexer.IDENT && nextToken.Type == lexer.COLON {
		return p.parseNamedFieldInitializer(), true
	}

	// Otherwise, parse as plain argument
	return p.parseArgumentAsFieldInitializer(), false
}

// advanceToNextItem handles separator logic and advances to next item if needed.
// Returns (shouldContinue, ok) where:
// - shouldContinue: true if there's another item to parse
// - ok: true if no error occurred
// PRE: cursor is at current item
// POST: cursor is at next item (if shouldContinue), or at terminator (if !shouldContinue)
func (p *Parser) advanceToNextItem(end lexer.TokenType) (bool, bool) {
	nextToken := p.cursor.Peek(1)

	// Check for separator (comma or semicolon)
	if nextToken.Type == lexer.COMMA || nextToken.Type == lexer.SEMICOLON {
		p.cursor = p.cursor.Advance() // consume separator

		// Check for trailing separator before terminator
		nextToken = p.cursor.Peek(1)
		if nextToken.Type == end {
			p.cursor = p.cursor.Advance() // consume terminator
			return false, true
		}

		// Advance to next item
		p.cursor = p.cursor.Advance()
		return true, true
	}

	// Check if we're at terminator
	if nextToken.Type == end {
		p.cursor = p.cursor.Advance() // consume terminator
		return false, true
	}

	// Unexpected token
	p.addError(fmt.Sprintf("expected ',' or '%s' in argument list, got %s", end, nextToken.Type), ErrUnexpectedToken)
	return false, false
}

// parseRecordLiteralInline parses an anonymous record literal: (name: value, ...)
// PRE: cursor is IDENT (first field name), peekToken is COLON
// POST: cursor is RPAREN
func (p *Parser) parseRecordLiteralInline() *ast.RecordLiteralExpression {
	currentTok := p.cursor.Current()

	recordLit := &ast.RecordLiteralExpression{
		BaseNode: ast.BaseNode{Token: currentTok}, // The first field name token
		TypeName: nil,                             // Anonymous record
		Fields:   []*ast.FieldInitializer{},
	}

	// Parse fields in a loop
	for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
		field := p.parseRecordField()
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

	return recordLit
}

// parseRecordField parses a single record field: name : value
// PRE: cursor is IDENT or other token
func (p *Parser) parseRecordField() *ast.FieldInitializer {
	if !p.curTokenIs(lexer.IDENT) || !p.peekTokenIs(lexer.COLON) {
		// Positional field - not yet supported
		p.addError("positional record field initialization not yet supported", ErrInvalidSyntax)
		return nil
	}

	return p.parseNamedFieldInitializer()
}
