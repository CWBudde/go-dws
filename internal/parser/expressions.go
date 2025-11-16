package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseExpression parses an expression with the given precedence.
// PRE: curToken is first token of expression
// POST: curToken is last token of expression
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.SEMICOLON) && (precedence < p.peekPrecedence() || (p.peekTokenIs(lexer.NOT) && precedence < EQUALS)) {
		// Special handling for "not in", "not is", "not as" operators
		// DWScript allows syntax like "x not in set" which means "not (x in set)"
		// We only handle this if our current precedence allows EQUALS-level operators
		if p.peekTokenIs(lexer.NOT) && precedence < EQUALS {
			// Check if this is "not in/is/as" by looking two tokens ahead
			savedCurToken := p.curToken
			savedPeekToken := p.peekToken

			p.nextToken() // move to NOT
			notToken := p.curToken

			// Check if the next token is IN, IS, or AS
			if p.peekTokenIs(lexer.IN) || p.peekTokenIs(lexer.IS) || p.peekTokenIs(lexer.AS) {
				// This is "not in", "not is", or "not as"
				// Parse the comparison expression first
				p.nextToken() // move to IN/IS/AS
				infix := p.infixParseFns[p.curToken.Type]
				if infix != nil {
					comparisonExp := infix(leftExp)

					// Wrap in NOT expression
					leftExp = &ast.UnaryExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token:  notToken,
								EndPos: comparisonExp.End(),
							},
						},
						Operator: notToken.Literal,
						Right:    comparisonExp,
					}
					continue
				}
			}

			// Not a "not in/is/as" pattern, restore state and exit
			p.curToken = savedCurToken
			p.peekToken = savedPeekToken
			return leftExp
		}

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
// PRE: curToken is IDENT
// POST: curToken is IDENT (unchanged)
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  p.curToken,
				EndPos: p.endPosFromToken(p.curToken),
			},
		},
		Value: p.curToken.Literal,
	}
}

// parseIntegerLiteral parses an integer literal.
// PRE: curToken is INT
// POST: curToken is INT (unchanged)
//
// Note: This currently uses the traditional implementation. The cursor-based version
// (parseIntegerLiteralCursor) exists and is tested, but full integration requires
// migrating parseExpression first (Task 2.2.4). See migration_integer_literal_test.go
// for validation that both implementations produce identical results.
func (p *Parser) parseIntegerLiteral() ast.Expression {
	return p.parseIntegerLiteralTraditional()
}

// parseIntegerLiteralTraditional parses an integer literal using traditional mutable state.
// PRE: curToken is INT
// POST: curToken is INT (unchanged)
func (p *Parser) parseIntegerLiteralTraditional() ast.Expression {
	lit := &ast.IntegerLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  p.curToken,
				EndPos: p.endPosFromToken(p.curToken),
			},
		},
	}

	literal := p.curToken.Literal

	var (
		value int64
		err   error
	)

	switch {
	case len(literal) > 0 && literal[0] == '$':
		// Hexadecimal with $ prefix (Pascal style)
		value, err = strconv.ParseInt(literal[1:], 16, 64)
	case len(literal) > 1 && (literal[0:2] == "0x" || literal[0:2] == "0X"):
		// Hexadecimal with 0x/0X prefix
		value, err = strconv.ParseInt(literal[2:], 16, 64)
	case len(literal) > 0 && literal[0] == '%':
		// Binary with % prefix
		value, err = strconv.ParseInt(literal[1:], 2, 64)
	default:
		value, err = strconv.ParseInt(literal, 10, 64)
	}

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", literal)
		p.addError(msg, ErrInvalidExpression)
		return nil
	}

	lit.Value = value
	return lit
}

// parseIntegerLiteralCursor parses an integer literal using cursor-based navigation.
// This is the cursor-based version of parseIntegerLiteral (Task 2.2.3).
//
// PRE: cursor.Current() is INT
// POST: cursor position unchanged (parsing functions don't advance cursor)
//
// Key differences from traditional version:
//   - Uses cursor.Current() instead of p.curToken
//   - No state mutation (immutable cursor)
//   - Clearer separation of token access from parsing logic
func (p *Parser) parseIntegerLiteralCursor() ast.Expression {
	currentToken := p.cursor.Current()

	lit := &ast.IntegerLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  currentToken,
				EndPos: p.endPosFromToken(currentToken),
			},
		},
	}

	literal := currentToken.Literal

	var (
		value int64
		err   error
	)

	switch {
	case len(literal) > 0 && literal[0] == '$':
		// Hexadecimal with $ prefix (Pascal style)
		value, err = strconv.ParseInt(literal[1:], 16, 64)
	case len(literal) > 1 && (literal[0:2] == "0x" || literal[0:2] == "0X"):
		// Hexadecimal with 0x/0X prefix
		value, err = strconv.ParseInt(literal[2:], 16, 64)
	case len(literal) > 0 && literal[0] == '%':
		// Binary with % prefix
		value, err = strconv.ParseInt(literal[1:], 2, 64)
	default:
		value, err = strconv.ParseInt(literal, 10, 64)
	}

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", literal)
		p.addError(msg, ErrInvalidExpression)
		return nil
	}

	lit.Value = value
	return lit
}

// parseFloatLiteral parses a floating-point literal.
// PRE: curToken is FLOAT
// POST: curToken is FLOAT (unchanged)
func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  p.curToken,
				EndPos: p.endPosFromToken(p.curToken),
			},
		},
	}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.addError(msg, ErrInvalidExpression)
		return nil
	}

	lit.Value = value
	return lit
}

// parseStringLiteral parses a string literal.
// PRE: curToken is STRING
// POST: curToken is STRING (unchanged)
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

	return &ast.StringLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  p.curToken,
				EndPos: p.endPosFromToken(p.curToken),
			},
		},
		Value: value,
	}
}

// unescapeString handles DWScript string escape sequences.
func unescapeString(s string) string {
	// Use strings.Builder for efficient string concatenation
	var result strings.Builder
	result.Grow(len(s)) // Pre-allocate approximate size

	// Convert to runes to handle UTF-8 correctly
	runes := []rune(s)
	i := 0
	for i < len(runes) {
		// Check for escaped single quote ('')
		if i < len(runes)-1 && runes[i] == '\'' && runes[i+1] == '\'' {
			result.WriteRune('\'')
			i += 2
		} else {
			result.WriteRune(runes[i])
			i++
		}
	}
	return result.String()
}

// parseBooleanLiteral parses a boolean literal.
// PRE: curToken is TRUE or FALSE
// POST: curToken is TRUE or FALSE (unchanged)
func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  p.curToken,
				EndPos: p.endPosFromToken(p.curToken),
			},
		},
		Value: p.curTokenIs(lexer.TRUE),
	}
}

// parseNilLiteral parses a nil literal.
// PRE: curToken is NIL
// POST: curToken is NIL (unchanged)
func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  p.curToken,
				EndPos: p.endPosFromToken(p.curToken),
			},
		},
	}
}

// parseNullIdentifier parses the Null keyword as an identifier.
// Task 9.4.1: Null is a built-in constant, so we parse it as an identifier.
// PRE: curToken is NULL
// POST: curToken is NULL (unchanged)
func (p *Parser) parseNullIdentifier() ast.Expression {
	return &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  p.curToken,
				EndPos: p.endPosFromToken(p.curToken),
			},
		},
		Value: p.curToken.Literal, // "Null" (preserves original casing)
	}
}

// parseUnassignedIdentifier parses the Unassigned keyword as an identifier.
// Task 9.4.1: Unassigned is a built-in constant, so we parse it as an identifier.
// PRE: curToken is UNASSIGNED
// POST: curToken is UNASSIGNED (unchanged)
func (p *Parser) parseUnassignedIdentifier() ast.Expression {
	return &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  p.curToken,
				EndPos: p.endPosFromToken(p.curToken),
			},
		},
		Value: p.curToken.Literal, // "Unassigned" (preserves original casing)
	}
}

// parseCharLiteral parses a character literal (#65, #$41).
// PRE: curToken is CHAR
// POST: curToken is CHAR (unchanged)
func (p *Parser) parseCharLiteral() ast.Expression {
	lit := &ast.CharLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  p.curToken,
				EndPos: p.endPosFromToken(p.curToken),
			},
		},
	}

	// Parse the character value from the token literal
	// Token literal can be: "#65" (decimal) or "#$41" (hex)
	literal := p.curToken.Literal
	if len(literal) < 2 || literal[0] != '#' {
		msg := fmt.Sprintf("invalid character literal format: %q", literal)
		p.addError(msg, ErrInvalidExpression)
		return nil
	}

	var value int64
	var err error

	if len(literal) >= 3 && literal[1] == '$' {
		// Hex format: #$41
		value, err = strconv.ParseInt(literal[2:], 16, 32)
	} else {
		// Decimal format: #65
		value, err = strconv.ParseInt(literal[1:], 10, 32)
	}

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as character literal", literal)
		p.addError(msg, ErrInvalidExpression)
		return nil
	}

	lit.Value = rune(value)
	return lit
}

// parsePrefixExpression parses a prefix (unary) expression.
// PRE: curToken is prefix operator (NOT, MINUS, PLUS, etc.)
// POST: curToken is last token of right operand
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.UnaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	// Set end position based on the right expression
	if expression.Right != nil {
		expression.EndPos = expression.Right.End()
	} else {
		expression.EndPos = p.endPosFromToken(expression.Token)
	}

	return expression
}

// parseAddressOfExpression parses the address-of operator (@) applied to a function or procedure.
// Examples: @MyFunction, @TMyClass.MyMethod
// PRE: curToken is AT
// POST: curToken is last token of target expression
func (p *Parser) parseAddressOfExpression() ast.Expression {
	expression := &ast.AddressOfExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: p.curToken}, // The @ token
		},
	}

	p.nextToken() // advance to the target

	// Parse the target expression (function/procedure name or member access)
	expression.Operator = p.parseExpression(PREFIX)

	// Set end position based on the target expression
	if expression.Operator != nil {
		expression.EndPos = expression.Operator.End()
	} else {
		expression.EndPos = p.endPosFromToken(expression.Token)
	}

	return expression
}

// parseInfixExpression parses an infix (binary) expression.
// PRE: curToken is the operator token
// POST: curToken is last token of right operand
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	// Set end position based on the right expression
	if expression.Right != nil {
		expression.EndPos = expression.Right.End()
	} else {
		expression.EndPos = p.endPosFromToken(expression.Token)
	}

	return expression
}

// parseCallExpression parses a function call expression.
// Also handles typed record literals: TypeName(field: value)
// PRE: curToken is LPAREN
// POST: curToken is RPAREN
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	// Check if this might be a typed record literal
	// Pattern: Identifier(Identifier:Expression, ...)
	if ident, ok := function.(*ast.Identifier); ok {
		// Parse the arguments, but check if they're all colon-based field initializers
		return p.parseCallOrRecordLiteral(ident)
	}

	// Normal function call (non-identifier function)
	exp := &ast.CallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: p.curToken},
		},
		Function: function,
	}

	exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	exp.EndPos = p.endPosFromToken(p.curToken) // p.curToken is now at RPAREN

	return exp
}

// parseCallOrRecordLiteral parses either a function call or a typed record literal.
// They have the same syntax initially: Identifier(...)
// The difference is whether the arguments are field initializers (name: value) or expressions.
// PRE: curToken is LPAREN
// POST: curToken is RPAREN
// parseCallOrRecordLiteral disambiguates between function calls and record literals.
// DWScript syntax allows both: TypeName(args) for calls and TypeName(field: value) for records.
// PRE: curToken is LPAREN
// POST: curToken is RPAREN
func (p *Parser) parseCallOrRecordLiteral(typeName *ast.Identifier) ast.Expression {
	// Empty parentheses -> function call
	if p.peekTokenIs(lexer.RPAREN) {
		return p.parseEmptyCall(typeName)
	}

	// Non-identifier first element -> must be function call
	if !p.peekTokenIs(lexer.IDENT) {
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
// PRE: peekToken is RPAREN
// POST: curToken is RPAREN
func (p *Parser) parseEmptyCall(typeName *ast.Identifier) *ast.CallExpression {
	p.nextToken() // consume ')'
	return &ast.CallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: p.curToken},
		},
		Function:  typeName,
		Arguments: []ast.Expression{},
	}
}

// parseCallWithExpressionList parses a function call using the standard expression list parser.
// PRE: curToken is LPAREN, peekToken is not RPAREN
// POST: curToken is RPAREN
func (p *Parser) parseCallWithExpressionList(typeName *ast.Identifier) *ast.CallExpression {
	exp := &ast.CallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: p.curToken},
		},
		Function: typeName,
	}
	exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	return exp
}

// buildRecordLiteral creates a record literal expression from field initializers.
func (p *Parser) buildRecordLiteral(typeName *ast.Identifier, fields []*ast.FieldInitializer) *ast.RecordLiteralExpression {
	return &ast.RecordLiteralExpression{
		BaseNode: ast.BaseNode{Token: p.curToken},
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
			BaseNode: ast.BaseNode{Token: p.curToken},
		},
		Function:  typeName,
		Arguments: args,
	}
}

// parseArgumentsOrFields parses a list that could be either function arguments or record fields.
// Returns the parsed items and whether ALL of them were colon-based fields.
// PRE: curToken is LPAREN
// POST: curToken is end token
func (p *Parser) parseArgumentsOrFields(end lexer.TokenType) ([]*ast.FieldInitializer, bool) {
	var items []*ast.FieldInitializer
	allHaveColons := true

	if p.peekTokenIs(end) {
		p.nextToken()
		return items, true // empty list
	}

	p.nextToken() // move to first element

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

// parseSingleArgumentOrField parses either a field initializer (name: value) or plain expression.
// Returns the item and whether it had a colon (i.e., was a field initializer).
func (p *Parser) parseSingleArgumentOrField() (*ast.FieldInitializer, bool) {
	if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.COLON) {
		return p.parseNamedFieldInitializer(), true
	}
	return p.parseArgumentAsFieldInitializer(), false
}

// parseNamedFieldInitializer parses a field initializer: name : value
// PRE: curToken is IDENT, peekToken is COLON
func (p *Parser) parseNamedFieldInitializer() *ast.FieldInitializer {
	fieldName := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

	p.nextToken() // move to ':'
	p.nextToken() // move to value

	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}

	return &ast.FieldInitializer{
		BaseNode: ast.BaseNode{Token: fieldName.Token},
		Name:     fieldName,
		Value:    value,
	}
}

// parseArgumentAsFieldInitializer parses a plain expression as a field initializer (without name).
// Used to represent function arguments in the same data structure as record fields.
func (p *Parser) parseArgumentAsFieldInitializer() *ast.FieldInitializer {
	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}

	return &ast.FieldInitializer{
		BaseNode: ast.BaseNode{Token: p.curToken},
		Name:     nil, // no name means regular argument
		Value:    expr,
	}
}

// advanceToNextItem handles separator logic and advances to next item if needed.
// Returns (shouldContinue, ok) where:
// - shouldContinue: true if there's another item to parse
// - ok: true if no error occurred
func (p *Parser) advanceToNextItem(end lexer.TokenType) (bool, bool) {
	if p.peekTokenIs(lexer.COMMA) || p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken() // consume separator
		if p.peekTokenIs(end) {
			// Trailing separator
			p.nextToken()
			return false, true
		}
		p.nextToken() // move to next item
		return true, true
	}

	if p.peekTokenIs(end) {
		p.nextToken()
		return false, true
	}

	p.addError("expected ',' or ')' in argument list", ErrUnexpectedToken)
	return false, false
}

// parseExpressionList parses a comma-separated list of expressions.
// PRE: curToken is LPAREN (or opening token)
// POST: curToken is end token (typically RPAREN)
func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	opts := ListParseOptions{
		Separators:             []lexer.TokenType{lexer.COMMA},
		Terminator:             end,
		AllowTrailingSeparator: true,
		AllowEmpty:             true,
		RequireTerminator:      true,
	}

	_, _ = p.parseSeparatedListBeforeStart(opts, func() bool {
		exp := p.parseExpression(LOWEST)
		if exp != nil {
			list = append(list, exp)
			return true
		}
		return false
	})

	return list
}

// parseGroupedExpression parses a grouped expression (parentheses).
// Also handles:
//   - Record literals: (X: 10, Y: 20)
//   - Array literals: (1, 2, 3)
//
// PRE: curToken is LPAREN
// POST: curToken is RPAREN
func (p *Parser) parseGroupedExpression() ast.Expression {
	lparenToken := p.curToken

	// Handle empty parentheses: () -> empty array literal
	if p.peekTokenIs(lexer.RPAREN) {
		return p.parseEmptyArrayLiteral(lparenToken)
	}

	// Check if this is a record literal: (IDENT : ...)
	if p.isRecordLiteralPattern() {
		p.nextToken() // advance to IDENT
		return p.parseRecordLiteralInline()
	}

	// Parse first expression, then determine if it's an array literal or grouped expression
	return p.parseExpressionOrArrayLiteral(lparenToken)
}

// parseEmptyArrayLiteral creates an empty array literal from empty parentheses.
// PRE: curToken is LPAREN, peekToken is RPAREN
// POST: curToken is RPAREN
func (p *Parser) parseEmptyArrayLiteral(lparenToken lexer.Token) *ast.ArrayLiteralExpression {
	p.nextToken() // consume ')'
	return &ast.ArrayLiteralExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  lparenToken,
				EndPos: p.curToken.End(),
			},
		},
		Elements: []ast.Expression{},
	}
}

// isRecordLiteralPattern checks if we're looking at a record literal pattern: (IDENT : ...)
// PRE: curToken is LPAREN
func (p *Parser) isRecordLiteralPattern() bool {
	return p.peekTokenIs(lexer.IDENT) && p.peekAhead(2).Type == lexer.COLON
}

// parseExpressionOrArrayLiteral parses either a grouped expression or array literal.
// Decides based on whether a comma follows the first expression.
// PRE: curToken is LPAREN
// POST: curToken is RPAREN
func (p *Parser) parseExpressionOrArrayLiteral(lparenToken lexer.Token) ast.Expression {
	p.nextToken() // move to first expression

	exp := p.parseExpression(LOWEST)
	if exp == nil {
		return nil
	}

	// Check if this is an array literal: (expr, expr, ...)
	if p.peekTokenIs(lexer.COMMA) {
		return p.parseParenthesizedArrayLiteral(lparenToken, exp)
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	// Return the expression directly, not wrapped in GroupedExpression
	// This avoids double parentheses in the string representation
	return exp
}

// parseParenthesizedArrayLiteral parses an array literal with parentheses: (expr1, expr2, ...)
// Called when we've already parsed the first element and detected a comma.
// PRE: curToken is last token of first element expression
// POST: curToken is RPAREN
func (p *Parser) parseParenthesizedArrayLiteral(lparenToken lexer.Token, firstElement ast.Expression) ast.Expression {
	elements := []ast.Expression{firstElement}

	// We're at the first expression, peek is COMMA
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // move to ','
		p.nextToken() // advance to next element or ')'

		// Allow trailing comma: (1, 2, )
		if p.curTokenIs(lexer.RPAREN) {
			// Already at the closing paren, just return
			return &ast.ArrayLiteralExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token:  lparenToken,
						EndPos: p.curToken.End(),
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

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return &ast.ArrayLiteralExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  lparenToken,
				EndPos: p.curToken.End(),
			},
		},
		Elements: elements,
	}
}

// parseRecordLiteralInline parses a record literal when we're already positioned
// at the first field name (after detecting the pattern "(IDENT:").
// PRE: curToken is first field name IDENT
// POST: curToken is RPAREN
// parseRecordLiteralInline parses an anonymous record literal: (name: value, ...)
// PRE: curToken is IDENT (first field name), peekToken is COLON
// POST: curToken is RPAREN
func (p *Parser) parseRecordLiteralInline() *ast.RecordLiteralExpression {
	recordLit := &ast.RecordLiteralExpression{
		BaseNode: ast.BaseNode{Token: p.curToken}, // The first field name token
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
// PRE: curToken is IDENT or other token
func (p *Parser) parseRecordField() *ast.FieldInitializer {
	if !p.curTokenIs(lexer.IDENT) || !p.peekTokenIs(lexer.COLON) {
		// Positional field - not yet supported
		p.addError("positional record field initialization not yet supported", ErrInvalidSyntax)
		return nil
	}

	return p.parseNamedFieldInitializer()
}

// parseNewExpression parses a new expression for both classes and arrays.
// DWScript syntax:
//   - new ClassName(args)     // Class instantiation
//   - new TypeName[size]      // Array instantiation (1D)
//   - new TypeName[s1, s2]    // Array instantiation (multi-dimensional)
//
// This function dispatches to the appropriate parser based on the token
// following the type name: '(' for classes, '[' for arrays.
// PRE: curToken is NEW
// POST: curToken is last token of new expression (RPAREN, RBRACK, or IDENT for zero-arg)
func (p *Parser) parseNewExpression() ast.Expression {
	newToken := p.curToken // Save the 'new' token position

	// Expect a type name (identifier)
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	typeName := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

	// Check what follows: '(' for class, '[' for array, or nothing for zero-arg constructor
	if p.peekTokenIs(lexer.LBRACK) {
		// Array instantiation: new TypeName[size, ...]
		return p.parseNewArrayExpression(newToken, typeName)
	} else if p.peekTokenIs(lexer.LPAREN) {
		// Class instantiation: new ClassName(args)
		return p.parseNewClassExpression(newToken, typeName)
	} else {
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
// This is the original parseNewExpression logic, now extracted as a helper.
// PRE: curToken is className IDENT
// POST: curToken is RPAREN
func (p *Parser) parseNewClassExpression(newToken lexer.Token, className *ast.Identifier) ast.Expression {
	// Create NewExpression
	newExpr := &ast.NewExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: newToken,
			},
		},
		ClassName: className,
		Arguments: []ast.Expression{},
	}

	// Expect opening parenthesis
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	// Parse arguments
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
// PRE: curToken is element type IDENT
// POST: curToken is RBRACK
func (p *Parser) parseNewArrayExpression(newToken lexer.Token, elementTypeName *ast.Identifier) ast.Expression {
	// Expect opening bracket
	if !p.expectPeek(lexer.LBRACK) {
		return nil
	}

	dimensions := []ast.Expression{}

	// Parse first dimension expression
	p.nextToken()
	firstDim := p.parseExpression(LOWEST)
	if firstDim == nil {
		p.addError(fmt.Sprintf("expected expression for array dimension at %d:%d",
			p.curToken.Pos.Line, p.curToken.Pos.Column), ErrInvalidExpression)
		return nil
	}
	dimensions = append(dimensions, firstDim)

	// Parse additional dimensions (comma-separated)
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to dimension expression

		dim := p.parseExpression(LOWEST)
		if dim == nil {
			p.addError(fmt.Sprintf("expected expression for array dimension at %d:%d",
				p.curToken.Pos.Line, p.curToken.Pos.Column), ErrInvalidExpression)
			return nil
		}
		dimensions = append(dimensions, dim)
	}

	// Expect closing bracket
	if !p.expectPeek(lexer.RBRACK) {
		return nil
	}

	return &ast.NewArrayExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: newToken},
		},
		ElementTypeName: elementTypeName,
		Dimensions:      dimensions,
	}
}

// parseInheritedExpression parses an inherited expression.
// Supports three forms:
//   - inherited                  // Bare inherited (calls same method in parent)
//   - inherited MethodName       // Call parent method (no args)
//   - inherited MethodName(args) // Call parent method with args
//
// PRE: curToken is INHERITED
// POST: curToken is INHERITED, method IDENT, or RPAREN (depends on form)
func (p *Parser) parseInheritedExpression() ast.Expression {
	inheritedExpr := &ast.InheritedExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken, // The 'inherited' keyword
			},
		},
	}

	// Check if there's a method name following
	if p.peekTokenIs(lexer.IDENT) {
		p.nextToken() // move to identifier
		inheritedExpr.Method = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: p.curToken,
				},
			},
			Value: p.curToken.Literal,
		}
		inheritedExpr.IsMember = true

		// Check if there's a call (parentheses)
		if p.peekTokenIs(lexer.LPAREN) {
			p.nextToken() // move to '('
			inheritedExpr.IsCall = true

			// Parse arguments
			inheritedExpr.Arguments = p.parseExpressionList(lexer.RPAREN)
			// Set end position after closing parenthesis (p.curToken is now at RPAREN)
			inheritedExpr.EndPos = p.endPosFromToken(p.curToken)
		} else {
			// No call, just method name - end at method identifier
			inheritedExpr.EndPos = inheritedExpr.Method.End()
		}
	} else {
		// Bare 'inherited' keyword - end at the keyword itself
		inheritedExpr.EndPos = p.endPosFromToken(inheritedExpr.Token)
	}

	return inheritedExpr
}

// parseSelfExpression parses a self expression.
// The Self keyword refers to the current instance (in instance methods) or
// the current class (in class methods).
// Usage: Self, Self.Field, Self.Method()
// PRE: curToken is SELF
// POST: curToken is SELF (unchanged)
func (p *Parser) parseSelfExpression() ast.Expression {
	selfExpr := &ast.SelfExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken, // The 'self' keyword
			},
		},
		Token: p.curToken,
	}

	// Set end position at the Self keyword itself
	selfExpr.EndPos = p.endPosFromToken(p.curToken)

	return selfExpr
}

// parseLambdaExpression parses a lambda/anonymous function expression.
// Supports both full and shorthand syntax:
//   - Full: lambda(x: Integer): Integer begin Result := x * 2; end
//   - Shorthand: lambda(x) => x * 2
//
// PRE: curToken is LAMBDA
// POST: curToken is last token of lambda body (END for full syntax, expression for shorthand)
func (p *Parser) parseLambdaExpression() ast.Expression {
	lambdaExpr := &ast.LambdaExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: p.curToken}, // The 'lambda' keyword
		},
	}

	// Expect opening parenthesis
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	// Parse parameter list (may be empty)
	lambdaExpr.Parameters = p.parseLambdaParameterList()

	// Check for return type annotation (optional)
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // move to ':'

		// Parse return type
		if !p.expectPeek(lexer.IDENT) {
			p.addError("expected return type after ':'", ErrExpectedType)
			return nil
		}

		lambdaExpr.ReturnType = &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}
	}

	// Check which syntax is being used: shorthand (=>) or full (begin/end)
	if p.peekTokenIs(lexer.FAT_ARROW) {
		// Shorthand syntax: lambda(x) => expression
		p.nextToken() // move to '=>'
		p.nextToken() // move past '=>' to expression

		// Parse the expression
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			p.addError("expected expression after '=>'", ErrInvalidExpression)
			return nil
		}

		// Desugar shorthand to full syntax: wrap expression in return statement
		lambdaExpr.Body = &ast.BlockStatement{
			BaseNode: ast.BaseNode{Token: p.curToken}, // Use current token for position tracking
			Statements: []ast.Statement{
				&ast.ReturnStatement{
					BaseNode: ast.BaseNode{
						Token: p.curToken,
					},
					ReturnValue: expr,
				},
			},
		}
		lambdaExpr.IsShorthand = true

		// Set end position based on expression
		if expr != nil {
			lambdaExpr.EndPos = expr.End()
		} else {
			lambdaExpr.EndPos = p.endPosFromToken(p.curToken)
		}

	} else if p.peekTokenIs(lexer.BEGIN) {
		// Full syntax: lambda(x: Integer) begin ... end
		p.nextToken() // move to 'begin'

		// Parse block statement
		lambdaExpr.Body = p.parseBlockStatement()
		lambdaExpr.IsShorthand = false

		// Set end position based on body block
		if lambdaExpr.Body != nil {
			lambdaExpr.EndPos = lambdaExpr.Body.End()
		} else {
			lambdaExpr.EndPos = p.endPosFromToken(p.curToken)
		}

	} else {
		p.addError("expected '=>' or 'begin' after lambda parameters", ErrUnexpectedToken)
		return nil
	}

	return lambdaExpr
}

// parseLambdaParameterList parses the parameter list for a lambda expression.
// Lambda parameters follow the same syntax as function parameters:
//   - Semicolon-separated groups: lambda(x: Integer; y: Integer)
//   - Comma-separated names with shared type: lambda(x, y: Integer)
//   - Mixed groups: lambda(x, y: Integer; z: String)
//   - Supports by-ref: lambda(var x: Integer; y: Integer)
//
// Note: Lambda parameters use semicolons between groups, matching DWScript function syntax.
// PRE: curToken is LAMBDA
// POST: curToken is RPAREN
func (p *Parser) parseLambdaParameterList() []*ast.Parameter {
	params := []*ast.Parameter{}

	p.nextToken() // move past '('

	// Empty parameter list
	if p.curTokenIs(lexer.RPAREN) {
		return params
	}

	// Parse parameter groups separated by semicolons
	for {
		// Parse a parameter group (one or more names with shared type)
		group := p.parseLambdaParameterGroup()
		if group == nil {
			return params
		}
		params = append(params, group...)

		// Check for more parameter groups
		if p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken() // move to ';'
			p.nextToken() // move past ';'
			continue
		}

		// No more groups, expect ')'
		if !p.expectPeek(lexer.RPAREN) {
			return params
		}
		break
	}

	return params
}

// parseLambdaParameterGroup parses a group of lambda parameters with the same type.
// Syntax: name: Type  or  name1, name2: Type  or  var name: Type  or  name (optional type)
// PRE: curToken is VAR or first parameter IDENT
// POST: curToken is type IDENT or last parameter name (if no type)
func (p *Parser) parseLambdaParameterGroup() []*ast.Parameter {
	params := []*ast.Parameter{}

	// Check for 'var' keyword (pass by reference)
	byRef := false
	if p.curTokenIs(lexer.VAR) {
		byRef = true
		p.nextToken() // move past 'var'
	}

	// Collect parameter names separated by commas
	names := []*ast.Identifier{}

	for {
		// Parse parameter name
		if !p.curTokenIs(lexer.IDENT) {
			p.addError("expected parameter name", ErrExpectedIdent)
			return nil
		}

		names = append(names, &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: p.curToken,
				},
			},
			Value: p.curToken.Literal,
		})

		// Check if there are more names (comma-separated)
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // move to ','
			p.nextToken() // move past ','
			continue
		}

		// No more names, check for type annotation
		break
	}

	// Check for optional type annotation
	var typeExpr ast.TypeExpression
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // move to ':'

		if !p.expectPeek(lexer.IDENT) {
			p.addError("expected type name after ':'", ErrExpectedType)
			return nil
		}

		typeExpr = &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}
	}

	// Create a parameter for each name with the same type (or nil if untyped)
	for _, name := range names {
		param := &ast.Parameter{
			Token: name.Token,
			Name:  name,
			Type:  typeExpr,
			ByRef: byRef,
		}
		params = append(params, param)
	}

	return params
}

// parseCondition parses a single contract condition.
// Syntax: boolean_expression [: "error message"]
// Returns a Condition node with the test expression and optional custom message.
// PRE: curToken is first token of condition expression
// POST: curToken is last token of condition (message STRING or test expression)
func (p *Parser) parseCondition() *ast.Condition {
	// Parse the test expression (should be boolean, but type checking is done in semantic phase)
	testExpr := p.parseExpression(LOWEST)
	if testExpr == nil {
		return nil
	}

	condition := &ast.Condition{
		BaseNode: ast.BaseNode{Token: p.curToken},
		Test:     testExpr,
	}

	// Check for optional custom message: : "message"
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // consume the colon

		// Expect a string literal for the error message
		if !p.expectPeek(lexer.STRING) {
			p.addError("expected string literal after ':' in contract condition", ErrUnexpectedToken)
			return nil
		}

		condition.Message = &ast.StringLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: p.curToken,
				},
			},
			Value: p.curToken.Literal,
		}
		// EndPos is the end of the message string literal
		condition.EndPos = p.endPosFromToken(p.curToken)
	} else {
		// EndPos is the end of the test expression
		condition.EndPos = testExpr.End()
	}

	return condition
}

// parseOldExpression parses an 'old' expression for contract postconditions.
// Syntax: old identifier
// The 'old' keyword can only be used in postconditions to reference pre-execution values.
// PRE: curToken is OLD
// POST: curToken is IDENT (identifier)
func (p *Parser) parseOldExpression() ast.Expression {
	token := p.curToken // the OLD token

	// Validate that we're in a postcondition context
	// Use new context API (Task 2.1.2) instead of direct field access
	if !p.ctx.ParsingPostCondition() {
		msg := fmt.Sprintf("'old' keyword can only be used in postconditions at line %d, column %d",
			token.Pos.Line, token.Pos.Column)
		p.addError(msg, ErrInvalidSyntax)
		return nil
	}

	// Expect an identifier after 'old'
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	identifier := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

	return &ast.OldExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  token,
				EndPos: identifier.End(),
			},
		},
		Identifier: identifier,
	}
}

// parsePreConditions parses function preconditions (require block).
// Syntax: require condition1; condition2; ...
// Returns a PreConditions node containing all parsed conditions.
// PRE: curToken is REQUIRE
// POST: curToken is last token of last condition
func (p *Parser) parsePreConditions() *ast.PreConditions {
	requireToken := p.curToken // the REQUIRE token

	// Advance to the first condition
	p.nextToken()

	var conditions []*ast.Condition

	// Parse first condition
	condition := p.parseCondition()
	if condition == nil {
		p.addError("expected at least one condition after 'require'", ErrInvalidExpression)
		return nil
	}
	conditions = append(conditions, condition)

	// Parse additional conditions separated by semicolons
	for p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken() // consume the semicolon

		// Check if we've reached the end of preconditions (peek at next token)
		// (beginning of var/const/begin or postconditions or EOF)
		if p.peekTokenIs(lexer.VAR) || p.peekTokenIs(lexer.CONST) ||
			p.peekTokenIs(lexer.BEGIN) || p.peekTokenIs(lexer.ENSURE) ||
			p.peekTokenIs(lexer.EOF) {
			break
		}

		p.nextToken() // move to the next condition

		condition := p.parseCondition()
		if condition == nil {
			break
		}
		conditions = append(conditions, condition)
	}

	preConditions := &ast.PreConditions{
		BaseNode:   ast.BaseNode{Token: requireToken},
		Conditions: conditions,
	}

	// EndPos is the end of the last condition
	if len(conditions) > 0 {
		preConditions.EndPos = conditions[len(conditions)-1].End()
	}

	return preConditions
}

// parsePostConditions parses function postconditions (ensure block).
// Syntax: ensure condition1; condition2; ...
// Returns a PostConditions node containing all parsed conditions.
// Sets parsingPostCondition flag to enable 'old' keyword parsing.
// PRE: curToken is ENSURE
// POST: curToken is last token of last condition
func (p *Parser) parsePostConditions() *ast.PostConditions {
	ensureToken := p.curToken // the ENSURE token

	// Enable 'old' keyword parsing
	// Synchronize both old field and new context (Task 2.1.2)
	p.parsingPostCondition = true
	p.ctx.SetParsingPostCondition(true)
	defer func() {
		p.parsingPostCondition = false
		p.ctx.SetParsingPostCondition(false)
	}()

	// Advance to the first condition
	p.nextToken()

	var conditions []*ast.Condition

	// Parse first condition
	condition := p.parseCondition()
	if condition == nil {
		p.addError("expected at least one condition after 'ensure'", ErrInvalidExpression)
		return nil
	}
	conditions = append(conditions, condition)

	// Parse additional conditions separated by semicolons
	for p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken() // consume the semicolon

		// Check if we've reached the end of postconditions (peek at next token)
		// (next function/procedure/type/begin/end/etc. or EOF)
		if p.peekTokenIs(lexer.FUNCTION) || p.peekTokenIs(lexer.PROCEDURE) ||
			p.peekTokenIs(lexer.TYPE) || p.peekTokenIs(lexer.VAR) ||
			p.peekTokenIs(lexer.CONST) || p.peekTokenIs(lexer.BEGIN) ||
			p.peekTokenIs(lexer.END) || p.peekTokenIs(lexer.IMPLEMENTATION) ||
			p.peekTokenIs(lexer.EOF) {
			break
		}

		p.nextToken() // move to the next condition

		condition := p.parseCondition()
		if condition == nil {
			break
		}
		conditions = append(conditions, condition)
	}

	postConditions := &ast.PostConditions{
		BaseNode:   ast.BaseNode{Token: ensureToken},
		Conditions: conditions,
	}

	// EndPos is the end of the last condition
	if len(conditions) > 0 {
		postConditions.EndPos = conditions[len(conditions)-1].End()
	}

	return postConditions
}

// parseIsExpression parses the 'is' operator which can be used for:
// 1. Type checking: obj is TMyClass
// 2. Boolean value comparison: boolExpr is True, boolExpr is False
// This creates an IsExpression AST node that will be evaluated at runtime.
// PRE: curToken is IS
// POST: curToken is last token of type or right expression
func (p *Parser) parseIsExpression(left ast.Expression) ast.Expression {
	expression := &ast.IsExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: p.curToken}, // The 'is' token
		},
		Left: left,
	}

	p.nextToken()

	// Try to parse as type expression first (speculatively)
	// Save state before attempting, so we can cleanly backtrack if it fails
	state := p.saveState()
	expression.TargetType = p.parseTypeExpression()
	if expression.TargetType != nil {
		expression.EndPos = expression.TargetType.End()
		return expression
	}

	// If type parsing failed, restore state and try as boolean expression
	// This removes any errors from the speculative parse
	p.restoreState(state)

	// Parse as value expression (boolean comparison)
	// Use EQUALS precedence to prevent consuming following logical operators
	expression.Right = p.parseExpression(EQUALS)
	if expression.Right == nil {
		p.addError("expected expression after 'is' operator", ErrInvalidExpression)
		return expression
	}
	expression.EndPos = expression.Right.End()

	return expression
}

// parseAsExpression parses the 'as' type casting operator.
// Example: obj as IMyInterface
// This creates an AsExpression AST node that will be evaluated at runtime
// to wrap an object instance in an InterfaceInstance.
// PRE: curToken is AS
// POST: curToken is last token of target type
func (p *Parser) parseAsExpression(left ast.Expression) ast.Expression {
	expression := &ast.AsExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: p.curToken}, // The 'as' token
		},
		Left: left,
	}

	p.nextToken()

	// Parse the target type (should be an interface type)
	expression.TargetType = p.parseTypeExpression()
	if expression.TargetType == nil {
		p.addError("expected type after 'as' operator", ErrExpectedType)
		return expression
	}

	// Set end position based on the target type
	expression.EndPos = expression.TargetType.End()

	return expression
}

// parseImplementsExpression parses the 'implements' operator.
// Example: obj implements IMyInterface  -> Boolean
// This creates an ImplementsExpression AST node that will be evaluated
// to check whether the object's class implements the interface.
// PRE: curToken is IMPLEMENTS
// POST: curToken is last token of target type
func (p *Parser) parseImplementsExpression(left ast.Expression) ast.Expression {
	expression := &ast.ImplementsExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: p.curToken}, // The 'implements' token
		},
		Left: left,
	}

	p.nextToken()

	// Parse the target type (should be an interface type)
	expression.TargetType = p.parseTypeExpression()
	if expression.TargetType == nil {
		p.addError("expected type after 'implements' operator", ErrExpectedType)
		return expression
	}

	// Set end position based on the target type
	expression.EndPos = expression.TargetType.End()

	return expression
}
