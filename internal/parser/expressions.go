package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseExpression parses an expression with the given precedence.
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
						Token:    notToken,
						Operator: notToken.Literal,
						Right:    comparisonExp,
						EndPos:   comparisonExp.End(),
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
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Token:  p.curToken,
		Value:  p.curToken.Literal,
		EndPos: p.endPosFromToken(p.curToken),
	}
}

// parseIntegerLiteral parses an integer literal.
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{
		Token:  p.curToken,
		EndPos: p.endPosFromToken(p.curToken),
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

// parseFloatLiteral parses a floating-point literal.
func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{
		Token:  p.curToken,
		EndPos: p.endPosFromToken(p.curToken),
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
		Token:  p.curToken,
		Value:  value,
		EndPos: p.endPosFromToken(p.curToken),
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
func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{
		Token:  p.curToken,
		Value:  p.curTokenIs(lexer.TRUE),
		EndPos: p.endPosFromToken(p.curToken),
	}
}

// parseNilLiteral parses a nil literal.
func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{
		Token:  p.curToken,
		EndPos: p.endPosFromToken(p.curToken),
	}
}

// parseCharLiteral parses a character literal (#65, #$41).
func (p *Parser) parseCharLiteral() ast.Expression {
	lit := &ast.CharLiteral{
		Token:  p.curToken,
		EndPos: p.endPosFromToken(p.curToken),
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
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.UnaryExpression{
		Token:    p.curToken,
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
func (p *Parser) parseAddressOfExpression() ast.Expression {
	expression := &ast.AddressOfExpression{
		Token: p.curToken, // The @ token
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
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.BinaryExpression{
		Token:    p.curToken,
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
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	// Check if this might be a typed record literal
	// Pattern: Identifier(Identifier:Expression, ...)
	if ident, ok := function.(*ast.Identifier); ok {
		// Parse the arguments, but check if they're all colon-based field initializers
		return p.parseCallOrRecordLiteral(ident)
	}

	// Normal function call (non-identifier function)
	exp := &ast.CallExpression{
		Token:    p.curToken,
		Function: function,
	}

	exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	exp.EndPos = p.endPosFromToken(p.curToken) // p.curToken is now at RPAREN

	return exp
}

// parseCallOrRecordLiteral parses either a function call or a typed record literal.
// They have the same syntax initially: Identifier(...)
// The difference is whether the arguments are field initializers (name: value) or expressions.
func (p *Parser) parseCallOrRecordLiteral(typeName *ast.Identifier) ast.Expression {
	// We're at '(' token
	// Peek ahead to see what's inside

	// Empty parentheses -> function call
	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken() // consume ')'
		return &ast.CallExpression{
			Token:     p.curToken,
			Function:  typeName,
			Arguments: []ast.Expression{},
		}
	}

	// Not empty - check if first element is "IDENT COLON"
	if !p.peekTokenIs(lexer.IDENT) {
		// First element is not an identifier, must be function call
		exp := &ast.CallExpression{
			Token:    p.curToken,
			Function: typeName,
		}
		exp.Arguments = p.parseExpressionList(lexer.RPAREN)
		return exp
	}

	// We have: TypeName(IDENT ...
	// Need to check if next token after IDENT is COLON
	// We'll use a special parsing mode to handle this

	// Try to parse as record literal fields
	fields, isRecordLiteral := p.tryParseRecordFields()

	if isRecordLiteral {
		// Successfully parsed as record literal
		return &ast.RecordLiteralExpression{
			Token:    p.curToken,
			TypeName: typeName,
			Fields:   fields,
		}
	}

	// Not a record literal, parse as normal function call
	// The problem: we may have already consumed some tokens in tryParseRecordFields
	// Solution: tryParseRecordFields should not consume tokens on failure
	// OR: we implement this differently

	// Actually, let's use a simpler approach:
	// Parse ALL arguments as a special list that handles both cases
	items, allHaveColons := p.parseArgumentsOrFields(lexer.RPAREN)

	if allHaveColons {
		// All items were field initializers -> record literal
		return &ast.RecordLiteralExpression{
			Token:    p.curToken,
			TypeName: typeName,
			Fields:   items,
		}
	}

	// Some or no items had colons -> function call
	// Extract just the expressions from the field initializers
	args := make([]ast.Expression, len(items))
	for i, item := range items {
		if item.Name != nil {
			// This shouldn't happen if allHaveColons is false, but handle it
			// Just use the identifier as the argument
			args[i] = item.Name
		} else {
			args[i] = item.Value
		}
	}

	return &ast.CallExpression{
		Token:     p.curToken,
		Function:  typeName,
		Arguments: args,
	}
}

// parseArgumentsOrFields parses a list that could be either function arguments or record fields.
// Returns the parsed items and whether ALL of them were colon-based fields.
func (p *Parser) parseArgumentsOrFields(end lexer.TokenType) ([]*ast.FieldInitializer, bool) {
	var items []*ast.FieldInitializer
	allHaveColons := true

	if p.peekTokenIs(end) {
		p.nextToken()
		return items, true // empty list
	}

	p.nextToken() // move to first element

	for {
		// Try to parse as "name : value"
		var item *ast.FieldInitializer

		if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.COLON) {
			// This is a field initializer: name : value
			fieldName := &ast.Identifier{
				Token: p.curToken,
				Value: p.curToken.Literal,
			}

			p.nextToken() // move to ':'
			p.nextToken() // move to value

			value := p.parseExpression(LOWEST)
			if value == nil {
				return items, false
			}

			item = &ast.FieldInitializer{
				Token: fieldName.Token,
				Name:  fieldName,
				Value: value,
			}
		} else {
			// Not a field initializer, just a regular expression
			expr := p.parseExpression(LOWEST)
			if expr == nil {
				return items, false
			}

			item = &ast.FieldInitializer{
				Token: p.curToken,
				Name:  nil, // no name means regular argument
				Value: expr,
			}
			allHaveColons = false
		}

		items = append(items, item)

		// Check for comma/semicolon separator
		if p.peekTokenIs(lexer.COMMA) || p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken() // consume separator
			if p.peekTokenIs(end) {
				// Trailing separator
				p.nextToken()
				break
			}
			p.nextToken() // move to next item
		} else if p.peekTokenIs(end) {
			p.nextToken()
			break
		} else {
			p.addError("expected ',' or ')' in argument list", ErrUnexpectedToken)
			return items, false
		}
	}

	return items, allHaveColons
}

// Stub functions to avoid compilation errors
func (p *Parser) tryParseRecordFields() ([]*ast.FieldInitializer, bool) {
	// This is replaced by parseArgumentsOrFields above
	return nil, false
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
// Also handles:
//   - Record literals: (X: 10, Y: 20)
//   - Array literals: (1, 2, 3)
func (p *Parser) parseGroupedExpression() ast.Expression {
	lparenToken := p.curToken

	// Handle empty parentheses: ()
	// This creates an empty array literal
	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken() // consume ')'
		return &ast.ArrayLiteralExpression{
			Token:    lparenToken,
			Elements: []ast.Expression{},
			EndPos:   p.curToken.End(),
		}
	}

	// Check if this is a record literal
	// Pattern: (IDENT : ...) indicates a named record literal
	// We need to look ahead two tokens: if peek is IDENT, advance and check if next peek is COLON
	if p.peekTokenIs(lexer.IDENT) {
		// Advance once to the IDENT
		p.nextToken()
		// Now check if peek is COLON
		if p.peekTokenIs(lexer.COLON) {
			// This is a named record literal!
			// Delegate to parseRecordLiteral() helper
			// We need to back up one token to let parseRecordLiteral start fresh
			// But since we can't back up, we'll pass the current position
			// Actually, parseRecordLiteral expects to be at '(' so we need to handle this
			// For now, parse inline but should refactor later
			return p.parseRecordLiteralInline()
		}
		// Not a record literal (no colon after ident)
		// We've already advanced past '(', so we're at IDENT
		// Parse this IDENT as an expression and continue
		exp := p.parseExpression(LOWEST)

		// Check if this is an array literal: (expr, expr, ...)
		if p.peekTokenIs(lexer.COMMA) {
			return p.parseParenthesizedArrayLiteral(lparenToken, exp)
		}

		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}

		return exp
	}

	// Not starting with IDENT, parse as normal grouped expression or array literal
	p.nextToken() // skip '('

	exp := p.parseExpression(LOWEST)

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
				Token:    lparenToken,
				Elements: elements,
				EndPos:   p.curToken.End(),
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
		Token:    lparenToken,
		Elements: elements,
		EndPos:   p.curToken.End(),
	}
}

// parseRecordLiteralInline parses a record literal when we're already positioned
// at the first field name (after detecting the pattern "(IDENT:").
func (p *Parser) parseRecordLiteralInline() *ast.RecordLiteralExpression {
	// We're currently at the IDENT after '(', and peek is COLON
	recordLit := &ast.RecordLiteralExpression{
		Token:    p.curToken, // The first field name token
		TypeName: nil,        // Anonymous record
		Fields:   []*ast.FieldInitializer{},
	}

	// Parse fields
	for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
		// We're at an IDENT, check if followed by COLON
		if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.COLON) {
			// Named field initialization
			fieldNameToken := p.curToken
			fieldName := &ast.Identifier{Token: fieldNameToken, Value: fieldNameToken.Literal}

			p.nextToken() // move to ':'
			p.nextToken() // move to value

			// Parse value expression
			value := p.parseExpression(LOWEST)
			if value == nil {
				p.addError("expected expression after ':' in record literal field", ErrInvalidExpression)
				return nil
			}

			fieldInit := &ast.FieldInitializer{
				Token: fieldNameToken,
				Name:  fieldName,
				Value: value,
			}

			recordLit.Fields = append(recordLit.Fields, fieldInit)
		} else {
			// Positional field - not yet supported
			p.addError("positional record field initialization not yet supported", ErrInvalidSyntax)
			return nil
		}

		// Check for separator (comma or semicolon)
		if p.peekTokenIs(lexer.COMMA) || p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken() // move to separator
			// Allow optional trailing separator
			if p.peekTokenIs(lexer.RPAREN) {
				p.nextToken() // move to ')'
				break
			}
			p.nextToken() // move to next field
		} else if p.peekTokenIs(lexer.RPAREN) {
			p.nextToken() // move to ')'
			break
		} else {
			p.addError("expected ',' or ';' or ')' in record literal", ErrUnexpectedToken)
			return nil
		}
	}

	return recordLit
}

// parseNewExpression parses a new expression for both classes and arrays.
// DWScript syntax:
//   - new ClassName(args)     // Class instantiation
//   - new TypeName[size]      // Array instantiation (1D)
//   - new TypeName[s1, s2]    // Array instantiation (multi-dimensional)
//
// This function dispatches to the appropriate parser based on the token
// following the type name: '(' for classes, '[' for arrays.
func (p *Parser) parseNewExpression() ast.Expression {
	newToken := p.curToken // Save the 'new' token position

	// Expect a type name (identifier)
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	typeName := &ast.Identifier{
		Token: p.curToken,
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
			Token:     newToken,
			ClassName: typeName,
			Arguments: []ast.Expression{},
		}
	}
}

// parseNewClassExpression parses class instantiation: new ClassName(args)
// This is the original parseNewExpression logic, now extracted as a helper.
func (p *Parser) parseNewClassExpression(newToken lexer.Token, className *ast.Identifier) ast.Expression {
	// Create NewExpression
	newExpr := &ast.NewExpression{
		Token:     newToken,
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
		Token:           newToken,
		ElementTypeName: elementTypeName,
		Dimensions:      dimensions,
	}
}

// parseInheritedExpression parses an inherited expression.
// Supports three forms:
//   - inherited                  // Bare inherited (calls same method in parent)
//   - inherited MethodName       // Call parent method (no args)
//   - inherited MethodName(args) // Call parent method with args
func (p *Parser) parseInheritedExpression() ast.Expression {
	inheritedExpr := &ast.InheritedExpression{
		Token: p.curToken, // The 'inherited' keyword
	}

	// Check if there's a method name following
	if p.peekTokenIs(lexer.IDENT) {
		p.nextToken() // move to identifier
		inheritedExpr.Method = &ast.Identifier{
			Token: p.curToken,
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

// parseLambdaExpression parses a lambda/anonymous function expression.
// Supports both full and shorthand syntax:
//   - Full: lambda(x: Integer): Integer begin Result := x * 2; end
//   - Shorthand: lambda(x) => x * 2
//
// Tasks 9.212-9.215: Parser support for lambda expressions
func (p *Parser) parseLambdaExpression() ast.Expression {
	lambdaExpr := &ast.LambdaExpression{
		Token: p.curToken, // The 'lambda' keyword
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
			Token: p.curToken, // Use current token for position tracking
			Statements: []ast.Statement{
				&ast.ReturnStatement{
					Token:       p.curToken,
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
			Token: p.curToken,
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
	var typeAnnotation *ast.TypeAnnotation
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // move to ':'

		if !p.expectPeek(lexer.IDENT) {
			p.addError("expected type name after ':'", ErrExpectedType)
			return nil
		}

		typeAnnotation = &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}
	}

	// Create a parameter for each name with the same type (or nil if untyped)
	for _, name := range names {
		param := &ast.Parameter{
			Token: name.Token,
			Name:  name,
			Type:  typeAnnotation,
			ByRef: byRef,
		}
		params = append(params, param)
	}

	return params
}

// parseCondition parses a single contract condition.
// Syntax: boolean_expression [: "error message"]
// Returns a Condition node with the test expression and optional custom message.
func (p *Parser) parseCondition() *ast.Condition {
	// Parse the test expression (should be boolean, but type checking is done in semantic phase)
	testExpr := p.parseExpression(LOWEST)
	if testExpr == nil {
		return nil
	}

	condition := &ast.Condition{
		Token: p.curToken,
		Test:  testExpr,
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
			Token: p.curToken,
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
func (p *Parser) parseOldExpression() ast.Expression {
	token := p.curToken // the OLD token

	// Validate that we're in a postcondition context
	if !p.parsingPostCondition {
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
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	return &ast.OldExpression{
		Token:      token,
		Identifier: identifier,
		EndPos:     identifier.End(),
	}
}

// parsePreConditions parses function preconditions (require block).
// Syntax: require condition1; condition2; ...
// Returns a PreConditions node containing all parsed conditions.
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
		Token:      requireToken,
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
func (p *Parser) parsePostConditions() *ast.PostConditions {
	ensureToken := p.curToken // the ENSURE token

	// Enable 'old' keyword parsing
	p.parsingPostCondition = true
	defer func() {
		p.parsingPostCondition = false
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
		Token:      ensureToken,
		Conditions: conditions,
	}

	// EndPos is the end of the last condition
	if len(conditions) > 0 {
		postConditions.EndPos = conditions[len(conditions)-1].End()
	}

	return postConditions
}

// parseIsExpression parses the 'is' type checking operator.
// Example: obj is TMyClass
// This creates an IsExpression AST node that will be evaluated at runtime
// to check if an object is an instance of a specific type.
func (p *Parser) parseIsExpression(left ast.Expression) ast.Expression {
	expression := &ast.IsExpression{
		Token: p.curToken, // The 'is' token
		Left:  left,
	}

	p.nextToken()

	// Parse the target type
	expression.TargetType = p.parseTypeExpression()
	if expression.TargetType == nil {
		p.addError("expected type after 'is' operator", ErrExpectedType)
		return expression
	}

	// Set end position based on the target type
	expression.EndPos = expression.TargetType.End()

	return expression
}

// parseAsExpression parses the 'as' type casting operator.
// Example: obj as IMyInterface
// This creates an AsExpression AST node that will be evaluated at runtime
// to wrap an object instance in an InterfaceInstance.
func (p *Parser) parseAsExpression(left ast.Expression) ast.Expression {
	expression := &ast.AsExpression{
		Token: p.curToken, // The 'as' token
		Left:  left,
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
func (p *Parser) parseImplementsExpression(left ast.Expression) ast.Expression {
	expression := &ast.ImplementsExpression{
		Token: p.curToken, // The 'implements' token
		Left:  left,
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
