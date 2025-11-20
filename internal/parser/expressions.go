package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
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
		if precedence >= nextPrec && !(nextToken.Type == lexer.NOT && precedence < EQUALS) {
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

// parseIntegerLiteral parses an integer literal.
// PRE: cursor is on INT
// POST: cursor is on INT (unchanged)
func (p *Parser) parseIntegerLiteral() ast.Expression {
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
		value, err = strconv.ParseInt(strings.ReplaceAll(literal[1:], "_", ""), 16, 64)
	case len(literal) > 1 && (literal[0:2] == "0x" || literal[0:2] == "0X"):
		// Hexadecimal with 0x/0X prefix
		value, err = strconv.ParseInt(strings.ReplaceAll(literal[2:], "_", ""), 16, 64)
	case len(literal) > 0 && literal[0] == '%':
		// Binary with % prefix
		value, err = strconv.ParseInt(strings.ReplaceAll(literal[1:], "_", ""), 2, 64)
	case len(literal) > 1 && (literal[0:2] == "0b" || literal[0:2] == "0B"):
		// Binary with 0b/0B prefix
		value, err = strconv.ParseInt(strings.ReplaceAll(literal[2:], "_", ""), 2, 64)
	default:
		value, err = strconv.ParseInt(strings.ReplaceAll(literal, "_", ""), 10, 64)
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
// POST: cursor is FLOAT (unchanged)

func (p *Parser) parseFloatLiteral() ast.Expression {
	currentToken := p.cursor.Current()

	lit := &ast.FloatLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  currentToken,
				EndPos: p.endPosFromToken(currentToken),
			},
		},
	}

	value, err := strconv.ParseFloat(strings.ReplaceAll(currentToken.Literal, "_", ""), 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", currentToken.Literal)
		p.addError(msg, ErrInvalidExpression)
		return nil
	}

	lit.Value = value
	return lit
}

// parseStringLiteral parses a string literal.
// POST: cursor is STRING (unchanged)

func (p *Parser) parseStringLiteral() ast.Expression {
	currentToken := p.cursor.Current()

	// The lexer has already processed the string, so we just need to
	// extract the value without the quotes
	value := currentToken.Literal

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
				Token:  currentToken,
				EndPos: p.endPosFromToken(currentToken),
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
// POST: cursor is TRUE or FALSE (unchanged)

func (p *Parser) parseBooleanLiteral() ast.Expression {
	currentToken := p.cursor.Current()
	return &ast.BooleanLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  currentToken,
				EndPos: p.endPosFromToken(currentToken),
			},
		},
		Value: currentToken.Type == lexer.TRUE,
	}
}

// parseNilLiteral parses a nil literal.
// PRE: cursor is NIL
// POST: cursor is NIL (unchanged)

// parseNullIdentifier parses the Null keyword as an identifier.
// Task 9.4.1: Null is a built-in constant, so we parse it as an identifier.
// PRE: cursor is NULL
// POST: cursor is NULL (unchanged)

// parseUnassignedIdentifier parses the Unassigned keyword as an identifier.
// Task 9.4.1: Unassigned is a built-in constant, so we parse it as an identifier.
// PRE: cursor is UNASSIGNED
// POST: cursor is UNASSIGNED (unchanged)

func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  p.cursor.Current(),
				EndPos: p.endPosFromToken(p.cursor.Current()),
			},
		},
	}
}

func (p *Parser) parseNullIdentifier() ast.Expression {
	tok := p.cursor.Current()
	return &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  tok,
				EndPos: p.endPosFromToken(tok),
			},
		},
		Value: tok.Literal, // "Null" (preserves original casing)
	}
}

func (p *Parser) parseUnassignedIdentifier() ast.Expression {
	tok := p.cursor.Current()
	return &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  tok,
				EndPos: p.endPosFromToken(tok),
			},
		},
		Value: tok.Literal, // "Unassigned" (preserves original casing)
	}
}

// parseCharLiteral parses a character literal (#65, #$41).
// PRE: cursor is CHAR
// POST: cursor is CHAR (unchanged)

func (p *Parser) parseCharLiteral() ast.Expression {
	tok := p.cursor.Current()
	lit := &ast.CharLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  tok,
				EndPos: p.endPosFromToken(tok),
			},
		},
	}

	// Parse the character value from the token literal
	// Token literal can be: "#65" (decimal) or "#$41" (hex)
	literal := tok.Literal
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
// PRE: cursor is prefix operator (NOT, MINUS, PLUS, etc.)
// POST: cursor is last token of right operand

// Parses unary prefix operators: -x, +x, not x
// PRE: cursor is on prefix operator token (MINUS, PLUS, NOT)
// POST: cursor is at last token of right expression
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

// parseAddressOfExpression parses the address-of operator (@) applied to a function or procedure.
// Examples: @MyFunction, @TMyClass.MyMethod
// PRE: cursor is AT
// POST: cursor is last token of target expression

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

// parseCallExpression parses a function call expression.
// Also handles typed record literals: TypeName(field: value)
// PRE: cursor is LPAREN
// POST: cursor is RPAREN

// Parses function call expressions and typed record literals using cursor navigation.
// PRE: cursor is on LPAREN
// POST: cursor is on RPAREN
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
// parseCallOrRecordLiteral disambiguates between function calls and record literals.
// DWScript syntax allows both: TypeName(args) for calls and TypeName(field: value) for records.
// PRE: cursor is LPAREN
// POST: cursor is RPAREN

// parseEmptyCall creates a call expression with no arguments.
// PRE: cursor.Peek(1) is RPAREN
// POST: cursor is RPAREN

// parseCallWithExpressionList parses a function call using the standard expression list parser.
// PRE: cursor is LPAREN, peekToken is not RPAREN
// POST: cursor is RPAREN

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

// parseCallWithExpressionList parses a function call using the cursor expression list parser.
// PRE: cursor is at LPAREN, cursor.Peek(1) is not RPAREN
// POST: cursor is at RPAREN
//
// Uses parseExpressionList instead of parseExpressionList.
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
// PRE: cursor is LPAREN
// POST: cursor is end token

// parseSingleArgumentOrField parses either a field initializer (name: value) or plain expression.
// Returns the item and whether it had a colon (i.e., was a field initializer).

// parseNamedFieldInitializer parses a field initializer: name : value
// PRE: cursor is IDENT, peekToken is COLON

// parseArgumentAsFieldInitializer parses a plain expression as a field initializer (without name).
// Used to represent function arguments in the same data structure as record fields.

// advanceToNextItem handles separator logic and advances to next item if needed.
// Returns (shouldContinue, ok) where:
// - shouldContinue: true if there's another item to parse
// - ok: true if no error occurred

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

// parseCallOrRecordLiteral orchestrates the disambiguation between function calls
// PRE: cursor is on LPAREN
// POST: cursor is on RPAREN
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

// parseExpressionList parses a comma-separated list of expressions.
// PRE: cursor is LPAREN (or opening token)
// POST: cursor is end token (typically RPAREN)

// PRE: cursor is before the list (at opening delimiter)
// POST: cursor is at terminator (closing delimiter)
//
// Uses parseExpression for each element instead of parseExpression.
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

// parseEmptyArrayLiteral creates an empty array literal from empty parentheses.
// PRE: cursor is LPAREN, peekToken is RPAREN
// POST: cursor is RPAREN
func (p *Parser) parseEmptyArrayLiteral(lparenToken lexer.Token) *ast.ArrayLiteralExpression {
	p.nextToken() // consume ')'

	var currentTok lexer.Token
	if p.cursor != nil {
		currentTok = p.cursor.Current()
	} else {
		currentTok = p.cursor.Current()
	}

	return &ast.ArrayLiteralExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  lparenToken,
				EndPos: currentTok.End(),
			},
		},
		Elements: []ast.Expression{},
	}
}

// isRecordLiteralPattern checks if we're looking at a record literal pattern: (IDENT : ...)
// PRE: cursor is LPAREN
func (p *Parser) isRecordLiteralPattern() bool {
	return p.peekTokenIs(lexer.IDENT) && p.peekAhead(2).Type == lexer.COLON
}

// parseExpressionOrArrayLiteral parses either a grouped expression or array literal.
// Decides based on whether a comma follows the first expression.
// PRE: cursor is LPAREN
// POST: cursor is RPAREN
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

// Parses grouped expressions in parentheses: (expr)
// Also handles empty parentheses, array literals, and record literals
// PRE: cursor is on LPAREN
// POST: cursor is on RPAREN
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

// parseParenthesizedArrayLiteral parses an array literal with parentheses: (expr1, expr2, ...)
// Called when we've already parsed the first element and detected a comma.
// PRE: cursor is last token of first element expression
// POST: cursor is RPAREN
func (p *Parser) parseParenthesizedArrayLiteral(lparenToken lexer.Token, firstElement ast.Expression) ast.Expression {
	elements := []ast.Expression{firstElement}

	// We're at the first expression, peek is COMMA
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // move to ','
		p.nextToken() // advance to next element or ')'

		// Allow trailing comma: (1, 2, )
		if p.curTokenIs(lexer.RPAREN) {
			var currentTok lexer.Token
			if p.cursor != nil {
				currentTok = p.cursor.Current()
			} else {
				currentTok = p.cursor.Current()
			}

			// Already at the closing paren, just return
			return &ast.ArrayLiteralExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token:  lparenToken,
						EndPos: currentTok.End(),
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

	var currentTok lexer.Token
	if p.cursor != nil {
		currentTok = p.cursor.Current()
	} else {
		currentTok = p.cursor.Current()
	}

	return &ast.ArrayLiteralExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  lparenToken,
				EndPos: currentTok.End(),
			},
		},
		Elements: elements,
	}
}

// parseRecordLiteralInline parses a record literal when we're already positioned
// at the first field name (after detecting the pattern "(IDENT:").
// PRE: cursor is first field name IDENT
// POST: cursor is RPAREN
// parseRecordLiteralInline parses an anonymous record literal: (name: value, ...)
// PRE: cursor is IDENT (first field name), peekToken is COLON
// POST: cursor is RPAREN
func (p *Parser) parseRecordLiteralInline() *ast.RecordLiteralExpression {
	var currentTok lexer.Token
	if p.cursor != nil {
		currentTok = p.cursor.Current()
	} else {
		currentTok = p.cursor.Current()
	}

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

// parseNewExpression parses a new expression for both classes and arrays.
// DWScript syntax:
//   - new ClassName(args)     // Class instantiation
//   - new TypeName[size]      // Array instantiation (1D)
//   - new TypeName[s1, s2]    // Array instantiation (multi-dimensional)
//
// This function dispatches to the appropriate parser based on the token
// following the type name: '(' for classes, '[' for arrays.
// PRE: cursor is NEW
// POST: cursor is last token of new expression (RPAREN, RBRACK, or IDENT for zero-arg)

// parseDefaultExpression parses a Default() call expression.
// DWScript syntax: Default(TypeName) - returns the default/zero value for the type
// PRE: cursor is DEFAULT
// POST: cursor is RPAREN

// parseNewClassExpression parses class instantiation: new ClassName(args)
// This is the original parseNewExpression logic, now extracted as a helper.
// PRE: cursor is className IDENT
// POST: cursor is RPAREN

// parseNewArrayExpression parses array instantiation: new TypeName[size1, size2, ...]
// Supports both single-dimensional and multi-dimensional arrays.
// Examples:
//   - new Integer[16]
//   - new String[10, 20]
//   - new Float[Length(arr)+1]
//
// PRE: cursor is element type IDENT
// POST: cursor is RBRACK

// parseInheritedExpression parses an inherited expression.
// Supports three forms:
//   - inherited                  // Bare inherited (calls same method in parent)
//   - inherited MethodName       // Call parent method (no args)
//   - inherited MethodName(args) // Call parent method with args
//
// PRE: cursor is INHERITED
// POST: cursor is INHERITED, method IDENT, or RPAREN (depends on form)

// parseSelfExpression parses a self expression.
// The Self keyword refers to the current instance (in instance methods) or
// the current class (in class methods).
// Usage: Self, Self.Field, Self.Method()
// PRE: cursor is SELF
// POST: cursor is SELF (unchanged)

// parseLambdaExpression parses a lambda/anonymous function expression.
// Supports both full and shorthand syntax:
//   - Full: lambda(x: Integer): Integer begin Result := x * 2; end
//   - Shorthand: lambda(x) => x * 2
//
// PRE: cursor is LAMBDA
// POST: cursor is last token of lambda body (END for full syntax, expression for shorthand)

// parseLambdaParameterList parses the parameter list for a lambda expression.
// Lambda parameters follow the same syntax as function parameters:
//   - Semicolon-separated groups: lambda(x: Integer; y: Integer)
//   - Comma-separated names with shared type: lambda(x, y: Integer)
//   - Mixed groups: lambda(x, y: Integer; z: String)
//   - Supports by-ref: lambda(var x: Integer; y: Integer)
//
// Note: Lambda parameters use semicolons between groups, matching DWScript function syntax.
// PRE: cursor is LAMBDA
// POST: cursor is RPAREN

// parseLambdaParameterGroup parses a group of lambda parameters with the same type.
// Syntax: name: Type  or  name1, name2: Type  or  var name: Type  or  name (optional type)
// PRE: cursor is VAR or first parameter IDENT
// POST: cursor is type IDENT or last parameter name (if no type)
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

		var nameTok lexer.Token
		if p.cursor != nil {
			nameTok = p.cursor.Current()
		} else {
			nameTok = p.cursor.Current()
		}

		names = append(names, &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: nameTok,
				},
			},
			Value: nameTok.Literal,
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

		var typeTok lexer.Token
		if p.cursor != nil {
			typeTok = p.cursor.Current()
		} else {
			typeTok = p.cursor.Current()
		}

		typeExpr = &ast.TypeAnnotation{
			Token: typeTok,
			Name:  typeTok.Literal,
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
// PRE: cursor is first token of condition expression
// POST: cursor is last token of condition (message STRING or test expression)
func (p *Parser) parseCondition() *ast.Condition {
	builder := p.StartNode()

	var startToken lexer.Token
	if p.cursor != nil {
		startToken = p.cursor.Current()
	} else {
		startToken = p.cursor.Current()
	}

	// Parse the test expression (should be boolean, but type checking is done in semantic phase)
	testExpr := p.parseExpression(LOWEST)
	if testExpr == nil {
		return nil
	}

	condition := &ast.Condition{
		BaseNode: ast.BaseNode{Token: startToken},
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

		var msgToken lexer.Token
		if p.cursor != nil {
			msgToken = p.cursor.Current()
		} else {
			msgToken = p.cursor.Current()
		}

		condition.Message = &ast.StringLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: msgToken,
				},
			},
			Value: msgToken.Literal,
		}
		// EndPos is the end of the message string literal
		return builder.Finish(condition).(*ast.Condition)
	} else {
		// EndPos is the end of the test expression
		return builder.FinishWithNode(condition, testExpr).(*ast.Condition)
	}
}

// parseOldExpression parses an 'old' expression for contract postconditions.
// Syntax: old identifier
// The 'old' keyword can only be used in postconditions to reference pre-execution values.
// PRE: cursor is OLD
// POST: cursor is IDENT (identifier)

// parsePreConditions parses function preconditions (require block).
// Syntax: require condition1; condition2; ...
// Returns a PreConditions node containing all parsed conditions.
// PRE: cursor is REQUIRE
// POST: cursor is last token of last condition
func (p *Parser) parsePreConditions() *ast.PreConditions {
	builder := p.StartNode()

	var requireToken lexer.Token
	if p.cursor != nil {
		requireToken = p.cursor.Current()
	} else {
		requireToken = p.cursor.Current()
	}

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
		return builder.FinishWithNode(preConditions, conditions[len(conditions)-1]).(*ast.PreConditions)
	}

	return preConditions
}

// parsePostConditions parses function postconditions (ensure block).
// Syntax: ensure condition1; condition2; ...
// Returns a PostConditions node containing all parsed conditions.
// Sets parsingPostCondition flag to enable 'old' keyword parsing.
// PRE: cursor is ENSURE
// POST: cursor is last token of last condition
func (p *Parser) parsePostConditions() *ast.PostConditions {
	builder := p.StartNode()

	var ensureToken lexer.Token
	if p.cursor != nil {
		ensureToken = p.cursor.Current()
	} else {
		ensureToken = p.cursor.Current()
	}

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
		return builder.FinishWithNode(postConditions, conditions[len(conditions)-1]).(*ast.PostConditions)
	}

	return postConditions
}

// parseIsExpression parses the 'is' operator which can be used for:
// 1. Type checking: obj is TMyClass
// 2. Boolean value comparison: boolExpr is True, boolExpr is False
// This creates an IsExpression AST node that will be evaluated at runtime.
// PRE: cursor is IS
// POST: cursor is last token of type or right expression

// parseAsExpression parses the 'as' type casting operator.
// Example: obj as IMyInterface
// This creates an AsExpression AST node that will be evaluated at runtime
// to wrap an object instance in an InterfaceInstance.
// PRE: cursor is AS
// POST: cursor is last token of target type

// parseImplementsExpression parses the 'implements' operator.
// Example: obj implements IMyInterface  -> Boolean
// This creates an ImplementsExpression AST node that will be evaluated
// to check whether the object's class implements the interface.
// PRE: cursor is IMPLEMENTS
// POST: cursor is last token of target type

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

// Example: obj as IMyInterface
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

// Example: obj implements IMyInterface  -> Boolean
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

// ============================================================================
// ============================================================================

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

// DWScript syntax: inherited [Method[(args)]]
// Examples:
//   - inherited          (call parent constructor/destructor)
//   - inherited.Method   (access parent method)
//   - inherited Method(args) (call parent method with args)
//
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

// DWScript syntax: new ClassName[(args)] or new array [size] of Type
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
	typeName := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: typeToken,
			},
		},
		Value: typeToken.Literal,
	}

	// Check what follows: '(' for class, '[' for array, or nothing for zero-arg constructor
	nextToken = p.cursor.Peek(1)
	if nextToken.Type == lexer.LBRACK {
		// Array instantiation: new TypeName[size, ...]
		return p.parseNewArrayExpression(newToken, typeName)
	} else if nextToken.Type == lexer.LPAREN {
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

// Supports both single-dimensional and multi-dimensional arrays.
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

	// Expect opening parenthesis
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.LPAREN {
		p.addError("expected '(' after 'lambda'", ErrUnexpectedToken)
		return nil
	}
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

	// Check which syntax is being used: shorthand (=>) or full (begin/end)
	nextToken = p.cursor.Peek(1)
	if nextToken.Type == lexer.FAT_ARROW {
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

	} else if nextToken.Type == lexer.BEGIN {
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

	} else {
		p.addError("expected '=>' or 'begin' after lambda parameters", ErrUnexpectedToken)
		return nil
	}
}

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

// The 'old' keyword refers to the value of a variable before function execution (in postconditions).
// DWScript syntax: old(identifier)
// PRE: cursor is on OLD token
// POST: cursor is on identifier
func (p *Parser) parseOldExpression() ast.Expression {
	currentToken := p.cursor.Current() // the OLD token

	// Validate that we're in a postcondition context
	// Use new context API (Task 2.1.2) instead of direct field access
	if !p.ctx.ParsingPostCondition() {
		msg := fmt.Sprintf("'old' keyword can only be used in postconditions at line %d, column %d",
			currentToken.Pos.Line, currentToken.Pos.Column)
		p.addError(msg, ErrInvalidSyntax)
		return nil
	}

	// Expect an identifier after 'old'
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.IDENT {
		p.addError("expected identifier after 'old'", ErrExpectedIdent)
		return nil
	}
	p.cursor = p.cursor.Advance() // move to identifier
	identToken := p.cursor.Current()

	identifier := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: identToken,
			},
		},
		Value: identToken.Literal,
	}

	return &ast.OldExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  currentToken,
				EndPos: identifier.End(),
			},
		},
		Identifier: identifier,
	}
}
