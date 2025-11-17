package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseExpression is a dispatcher that routes to the appropriate implementation
// based on the parser mode (traditional vs cursor).
//
// Task 2.2.7: This dispatcher enables dual-mode operation during migration.
// Eventually (Phase 2.7), only the cursor version will remain.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	if p.useCursor {
		return p.parseExpressionCursor(precedence)
	}
	return p.parseExpressionTraditional(precedence)
}

// parseExpressionTraditional parses an expression with the given precedence (traditional mode).
// PRE: curToken is first token of expression
// POST: curToken is last token of expression
//
// This is the original implementation using mutable parser state (curToken/peekToken).
// Task 2.2.7: Renamed from parseExpression to enable dual-mode operation.
func (p *Parser) parseExpressionTraditional(precedence int) ast.Expression {
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

// parseExpressionCursor parses an expression with the given precedence (cursor mode).
// PRE: cursor is at first token of expression
// POST: cursor is at last token of expression
//
// This is the cursor-based implementation using immutable cursor navigation.
// It uses registered cursor prefix/infix functions from prefixParseFnsCursor and
// infixParseFnsCursor maps. When encountering a token type without a cursor
// implementation, it gracefully falls back to traditional mode for that expression
// subtree. This allows incremental migration - as more functions are migrated to
// cursor mode (in Tasks 2.2.10, 2.2.11), cursor coverage will naturally increase.
//
// Currently registered cursor functions:
// - Prefix: IDENT, INT, FLOAT, STRING, TRUE, FALSE
// - Infix: +, -, *, /, div, mod, shl, shr, sar, =, <>, <, >, <=, >=, and, or, xor, in, ??
//
// Task 2.2.7: New implementation for pure functional parsing.
func (p *Parser) parseExpressionCursor(precedence int) ast.Expression {
	// 1. Lookup and call prefix function
	currentToken := p.cursor.Current()
	prefixFn, ok := p.prefixParseFnsCursor[currentToken.Type]
	if !ok {
		// No cursor version - fall back to traditional mode
		p.syncCursorToTokens() // Sync cursor → tokens
		p.useCursor = false
		result := p.parseExpressionTraditional(precedence)
		p.useCursor = true
		// CRITICAL FIX (Task 2.2.7): Sync tokens → cursor after traditional mode mutates state
		// Traditional parseExpression advanced curToken/peekToken, but cursor is still at old position.
		// Update cursor to match the new curToken position to prevent infinite fallback loops.
		p.syncTokensToCursor()
		return result
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
			leftExp = p.parseNotInIsAsCursor(leftExp)
			if leftExp == nil {
				// Not a "not in/is/as" pattern, return what we have
				break
			}
			continue
		}

		// 4. Normal infix handling
		infixFn, ok := p.infixParseFnsCursor[nextToken.Type]
		if !ok {
			// No cursor version - fall back to traditional mode for rest
			p.syncCursorToTokens() // Sync cursor → tokens
			p.useCursor = false
			result := p.parseExpressionTraditional(precedence)
			p.useCursor = true
			// CRITICAL FIX (Task 2.2.7): Sync tokens → cursor after traditional mode mutates state
			p.syncTokensToCursor()
			return result
		}

		// Advance to operator
		p.cursor = p.cursor.Advance()
		operatorToken := p.cursor.Current()

		// Task 2.2.9: Call infix function directly without state sync
		// All registered infix cursor functions now use parseInfixExpressionCursor,
		// which is pure cursor and recursively calls parseExpressionCursor
		// Call infix function
		leftExp = infixFn(leftExp, operatorToken)
	}

	// Sync cursor position back to curToken/peekToken for backward compatibility
	// External code (like parseIfStatement) uses curToken/peekToken, not cursor
	p.syncCursorToTokens()

	return leftExp
}

// parseNotInIsAsCursor handles special "not in", "not is", "not as" operators in cursor mode.
// Returns the wrapped NOT expression if successful, or nil if this is not a "not in/is/as" pattern.
//
// Task 2.2.7: Cursor-based implementation using Mark/ResetTo for backtracking.
func (p *Parser) parseNotInIsAsCursor(leftExp ast.Expression) ast.Expression {
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
		p.syncCursorToTokens()
		return nil
	}

	// This is "not in", "not is", or "not as"
	// Advance to IN/IS/AS token
	p.cursor = p.cursor.Advance()
	operatorToken := p.cursor.Current()

	// Look up infix function for the operator
	infixFn, ok := p.infixParseFnsCursor[operatorToken.Type]
	if !ok {
		// No infix function, backtrack
		p.cursor = p.cursor.ResetTo(mark)
		p.syncCursorToTokens()
		return nil
	}

	// Task 2.2.9: Call infix function directly without state sync
	// Now that parseInfixExpressionCursor is pure cursor, no sync needed
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
// PRE: curToken is IDENT (traditional) or cursor.Current() is IDENT (cursor)
// POST: curToken is IDENT (unchanged)
func (p *Parser) parseIdentifier() ast.Expression {
	return p.parseIdentifierTraditional()
}

// parseIdentifierTraditional parses an identifier using traditional state.
func (p *Parser) parseIdentifierTraditional() ast.Expression {
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

// parseIdentifierCursor parses an identifier using cursor navigation.
func (p *Parser) parseIdentifierCursor() ast.Expression {
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
// PRE: curToken is FLOAT (traditional) or cursor.Current() is FLOAT (cursor)
// POST: curToken is FLOAT (unchanged)
func (p *Parser) parseFloatLiteral() ast.Expression {
	return p.parseFloatLiteralTraditional()
}

// parseFloatLiteralTraditional parses a float literal using traditional state.
func (p *Parser) parseFloatLiteralTraditional() ast.Expression {
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

// parseFloatLiteralCursor parses a float literal using cursor navigation.
func (p *Parser) parseFloatLiteralCursor() ast.Expression {
	currentToken := p.cursor.Current()

	lit := &ast.FloatLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  currentToken,
				EndPos: p.endPosFromToken(currentToken),
			},
		},
	}

	value, err := strconv.ParseFloat(currentToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", currentToken.Literal)
		p.addError(msg, ErrInvalidExpression)
		return nil
	}

	lit.Value = value
	return lit
}

// parseStringLiteral parses a string literal.
// PRE: curToken is STRING (traditional) or cursor.Current() is STRING (cursor)
// POST: curToken is STRING (unchanged)
func (p *Parser) parseStringLiteral() ast.Expression {
	return p.parseStringLiteralTraditional()
}

// parseStringLiteralTraditional parses a string literal using traditional state.
func (p *Parser) parseStringLiteralTraditional() ast.Expression {
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

// parseStringLiteralCursor parses a string literal using cursor navigation.
func (p *Parser) parseStringLiteralCursor() ast.Expression {
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
// PRE: curToken is TRUE or FALSE (traditional) or cursor.Current() is TRUE/FALSE (cursor)
// POST: curToken is TRUE or FALSE (unchanged)
func (p *Parser) parseBooleanLiteral() ast.Expression {
	return p.parseBooleanLiteralTraditional()
}

// parseBooleanLiteralTraditional parses a boolean literal using traditional state.
func (p *Parser) parseBooleanLiteralTraditional() ast.Expression {
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

// parseBooleanLiteralCursor parses a boolean literal using cursor navigation.
func (p *Parser) parseBooleanLiteralCursor() ast.Expression {
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

// parseNilLiteralCursor parses a nil literal in cursor mode.
// Task 2.2.12: Cursor mode version
func (p *Parser) parseNilLiteralCursor() ast.Expression {
	return &ast.NilLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  p.cursor.Current(),
				EndPos: p.endPosFromToken(p.cursor.Current()),
			},
		},
	}
}

// parseNullIdentifierCursor parses the Null keyword as an identifier in cursor mode.
// Task 2.2.12: Cursor mode version
func (p *Parser) parseNullIdentifierCursor() ast.Expression {
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

// parseUnassignedIdentifierCursor parses the Unassigned keyword as an identifier in cursor mode.
// Task 2.2.12: Cursor mode version
func (p *Parser) parseUnassignedIdentifierCursor() ast.Expression {
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

// parseCharLiteralCursor parses a character literal (#65, #$41) in cursor mode.
// Task 2.2.12: Cursor mode version
func (p *Parser) parseCharLiteralCursor() ast.Expression {
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

// Task 2.2.12: parsePrefixExpressionCursor - Cursor mode version of parsePrefixExpression
// Parses unary prefix operators: -x, +x, not x
// PRE: cursor is on prefix operator token (MINUS, PLUS, NOT)
// POST: cursor is at last token of right expression
func (p *Parser) parsePrefixExpressionCursor() ast.Expression {
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
	expression.Right = p.parseExpressionCursor(PREFIX)

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

// parseInfixExpression parses a binary infix expression (dispatcher).
// PRE: curToken is the operator token (traditional) or cursor at operator (cursor)
// POST: curToken is last token of right expression (traditional)
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	return p.parseInfixExpressionTraditional(left)
}

// parseInfixExpressionTraditional parses a binary infix expression using traditional state.
// PRE: curToken is the operator token
// POST: curToken is last token of right expression
func (p *Parser) parseInfixExpressionTraditional(left ast.Expression) ast.Expression {
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

// parseInfixExpressionCursor parses a binary infix expression using cursor navigation.
// PRE: cursor at operator token
// POST: cursor position advanced (state mutation needed for now until parseExpression is migrated)
//
// Note: This cursor version still calls the traditional parseExpression internally,
// because full cursor integration requires migrating parseExpression itself (future task).
// For now, we sync the cursor state with traditional state before/after the recursive call.
func (p *Parser) parseInfixExpressionCursor(left ast.Expression) ast.Expression {
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

	// Task 2.2.9: Parse right expression using pure cursor mode
	// Now that parseExpressionCursor is implemented, we can call it directly
	// for pure cursor-to-cursor recursion without state synchronization
	expression.Right = p.parseExpressionCursor(precedence)

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

// Task 2.2.11: parseCallExpressionCursor - Cursor mode version of parseCallExpression
// Parses function call expressions and typed record literals using cursor navigation.
// PRE: cursor is on LPAREN
// POST: cursor is on RPAREN
func (p *Parser) parseCallExpressionCursor(function ast.Expression) ast.Expression {
	// Check if this might be a typed record literal
	// Pattern: Identifier(Identifier:Expression, ...)
	if ident, ok := function.(*ast.Identifier); ok {
		// Parse the arguments, but check if they're all colon-based field initializers
		return p.parseCallOrRecordLiteralCursor(ident)
	}

	// Normal function call (non-identifier function)
	lparenToken := p.cursor.Current()

	exp := &ast.CallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: lparenToken},
		},
		Function: function,
	}

	exp.Arguments = p.parseExpressionListCursor(lexer.RPAREN)
	exp.EndPos = p.endPosFromToken(p.cursor.Current()) // cursor is now at RPAREN

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

// parseEmptyCallCursor creates a call expression with no arguments in cursor mode.
// PRE: cursor.Peek(1) is RPAREN
// POST: cursor is at RPAREN
//
// Task 2.2.10 Phase 2: Cursor version of parseEmptyCall.
func (p *Parser) parseEmptyCallCursor(typeName *ast.Identifier) *ast.CallExpression {
	// Advance to RPAREN
	p.cursor = p.cursor.Advance()
	rparenToken := p.cursor.Current()

	return &ast.CallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  rparenToken,
				EndPos: p.endPosFromToken(rparenToken),
			},
		},
		Function:  typeName,
		Arguments: []ast.Expression{},
	}
}

// parseCallWithExpressionListCursor parses a function call using the cursor expression list parser.
// PRE: cursor is at LPAREN, cursor.Peek(1) is not RPAREN
// POST: cursor is at RPAREN
//
// Task 2.2.10 Phase 2: Cursor version of parseCallWithExpressionList.
// Uses parseExpressionListCursor instead of parseExpressionList.
func (p *Parser) parseCallWithExpressionListCursor(typeName *ast.Identifier) *ast.CallExpression {
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
	exp.Arguments = p.parseExpressionListCursor(lexer.RPAREN)

	// Set end position to RPAREN
	exp.EndPos = p.cursor.Current().End()

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

// parseNamedFieldInitializerCursor parses a field initializer: name : value (cursor mode).
// PRE: cursor is at IDENT, cursor.Peek(1) is COLON
// POST: cursor is at value expression
//
// Task 2.2.10 Phase 3: Cursor version of parseNamedFieldInitializer.
func (p *Parser) parseNamedFieldInitializerCursor() *ast.FieldInitializer {
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

	// Parse value using cursor mode
	value := p.parseExpressionCursor(LOWEST)
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

// parseArgumentAsFieldInitializerCursor parses a plain expression as a field initializer (cursor mode).
// Used to represent function arguments in the same data structure as record fields.
// PRE: cursor is at start of expression
// POST: cursor is at end of expression
//
// Task 2.2.10 Phase 3: Cursor version of parseArgumentAsFieldInitializer.
func (p *Parser) parseArgumentAsFieldInitializerCursor() *ast.FieldInitializer {
	exprStart := p.cursor.Current()

	expr := p.parseExpressionCursor(LOWEST)
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

// parseSingleArgumentOrFieldCursor parses either a field initializer (name: value) or plain expression (cursor mode).
// Returns the item and whether it had a colon (i.e., was a field initializer).
// PRE: cursor is at start of argument/field
// POST: cursor is at end of argument/field
//
// Task 2.2.10 Phase 3: Cursor version of parseSingleArgumentOrField.
func (p *Parser) parseSingleArgumentOrFieldCursor() (*ast.FieldInitializer, bool) {
	currentToken := p.cursor.Current()
	nextToken := p.cursor.Peek(1)

	// Check for field initializer pattern: IDENT COLON
	if currentToken.Type == lexer.IDENT && nextToken.Type == lexer.COLON {
		return p.parseNamedFieldInitializerCursor(), true
	}

	// Otherwise, parse as plain argument
	return p.parseArgumentAsFieldInitializerCursor(), false
}

// advanceToNextItemCursor handles separator logic and advances to next item if needed (cursor mode).
// Returns (shouldContinue, ok) where:
// - shouldContinue: true if there's another item to parse
// - ok: true if no error occurred
// PRE: cursor is at current item
// POST: cursor is at next item (if shouldContinue), or at terminator (if !shouldContinue)
//
// Task 2.2.10 Phase 3: Cursor version of advanceToNextItem.
func (p *Parser) advanceToNextItemCursor(end lexer.TokenType) (bool, bool) {
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

// Task 2.2.10 Phase 4: parseArgumentsOrFieldsCursor
// parseArgumentsOrFieldsCursor parses a list that could be either function arguments or record fields.
// Returns the parsed items and whether ALL of them were colon-based fields.
// PRE: cursor is on LPAREN
// POST: cursor is on end token
func (p *Parser) parseArgumentsOrFieldsCursor(end lexer.TokenType) ([]*ast.FieldInitializer, bool) {
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
		item, hasColon := p.parseSingleArgumentOrFieldCursor()
		if item == nil {
			return items, false
		}

		if !hasColon {
			allHaveColons = false
		}

		items = append(items, item)

		// Check if we should continue to next item
		shouldContinue, ok := p.advanceToNextItemCursor(end)
		if !ok {
			return items, false
		}
		if !shouldContinue {
			break
		}
	}

	return items, allHaveColons
}

// Task 2.2.10 Phase 5: parseCallOrRecordLiteralCursor
// parseCallOrRecordLiteralCursor orchestrates the disambiguation between function calls
// and record literals using cursor mode.
// PRE: cursor is on LPAREN
// POST: cursor is on RPAREN
func (p *Parser) parseCallOrRecordLiteralCursor(typeName *ast.Identifier) ast.Expression {
	// Empty parentheses -> function call
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.RPAREN {
		return p.parseEmptyCallCursor(typeName)
	}

	// Non-identifier first element -> must be function call
	if nextToken.Type != lexer.IDENT {
		return p.parseCallWithExpressionListCursor(typeName)
	}

	// We have: TypeName(IDENT ...
	// Parse arguments/fields and determine type based on whether ALL have colons
	items, allHaveColons := p.parseArgumentsOrFieldsCursor(lexer.RPAREN)

	if allHaveColons {
		// All items were field initializers -> record literal
		return p.buildRecordLiteral(typeName, items)
	}

	// Some or no items had colons -> function call
	return p.buildCallExpressionFromFields(typeName, items)
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

// parseExpressionListCursor parses a comma-separated list of expressions in cursor mode.
// PRE: cursor is before the list (at opening delimiter)
// POST: cursor is at terminator (closing delimiter)
//
// Task 2.2.10 Phase 1: Cursor version of parseExpressionList.
// Uses parseExpressionCursor for each element instead of parseExpression.
func (p *Parser) parseExpressionListCursor(end lexer.TokenType) []ast.Expression {
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
	expr := p.parseExpressionCursor(LOWEST)
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
			expr = p.parseExpressionCursor(LOWEST)
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

// Task 2.2.12: parseGroupedExpressionCursor - Cursor mode version of parseGroupedExpression
// Parses grouped expressions in parentheses: (expr)
// Also handles empty parentheses, array literals, and record literals
// PRE: cursor is on LPAREN
// POST: cursor is on RPAREN
func (p *Parser) parseGroupedExpressionCursor() ast.Expression {
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
		// Fall back to traditional mode for record literals (complex parsing)
		p.syncCursorToTokens()
		p.useCursor = false
		result := p.parseGroupedExpression()
		p.useCursor = true
		p.syncTokensToCursor()
		return result
	}

	// Move to first expression
	p.cursor = p.cursor.Advance()

	// Parse first expression
	exp := p.parseExpressionCursor(LOWEST)
	if exp == nil {
		return nil
	}

	// Check if this is an array literal: (expr, expr, ...)
	nextToken = p.cursor.Peek(1)
	if nextToken.Type == lexer.COMMA {
		// Fall back to traditional mode for array literals (complex parsing)
		p.syncCursorToTokens()
		p.useCursor = false
		result := p.parseGroupedExpression()
		p.useCursor = true
		p.syncTokensToCursor()
		return result
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

// parseDefaultExpression parses a Default() call expression.
// DWScript syntax: Default(TypeName) - returns the default/zero value for the type
// PRE: curToken is DEFAULT
// POST: curToken is RPAREN
func (p *Parser) parseDefaultExpression() ast.Expression {
	defaultToken := p.curToken // Save the 'default' token position

	// Expect LPAREN
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	// Parse the type name argument
	p.nextToken() // Move to type name

	// The type name could be an identifier (Integer, String, etc.)
	typeName := p.parseExpression(LOWEST)
	if typeName == nil {
		return nil
	}

	// Expect RPAREN
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

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

// parseIsExpressionCursor parses the 'is' operator in cursor mode.
// Task 2.2.13: Cursor mode version
// Example: obj is TClass  -> Boolean
// PRE: cursor is on IS token
// POST: cursor is on last token of type/expression
func (p *Parser) parseIsExpressionCursor(left ast.Expression) ast.Expression {
	isToken := p.cursor.Current()
	expression := &ast.IsExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: isToken},
		},
		Left: left,
	}

	p.cursor = p.cursor.Advance()

	// For type parsing, fall back to traditional mode since parseTypeExpression
	// is complex and doesn't have a cursor version yet
	p.syncCursorToTokens()
	p.useCursor = false

	// Try to parse as type expression first (speculatively)
	state := p.saveState()
	expression.TargetType = p.parseTypeExpression()
	if expression.TargetType != nil {
		expression.EndPos = expression.TargetType.End()
		p.useCursor = true
		p.syncTokensToCursor()
		return expression
	}

	// If type parsing failed, restore state and try as boolean expression
	p.restoreState(state)

	// Back to cursor mode for expression parsing
	p.useCursor = true
	p.syncTokensToCursor()

	// Parse as value expression (boolean comparison)
	// Use EQUALS precedence to prevent consuming following logical operators
	expression.Right = p.parseExpressionCursor(EQUALS)
	if expression.Right == nil {
		p.addError("expected expression after 'is' operator", ErrInvalidExpression)
		return expression
	}
	expression.EndPos = expression.Right.End()

	return expression
}

// parseAsExpressionCursor parses the 'as' type casting operator in cursor mode.
// Task 2.2.13: Cursor mode version
// Example: obj as IMyInterface
// PRE: cursor is on AS token
// POST: cursor is on last token of target type
func (p *Parser) parseAsExpressionCursor(left ast.Expression) ast.Expression {
	asToken := p.cursor.Current()
	expression := &ast.AsExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: asToken},
		},
		Left: left,
	}

	p.cursor = p.cursor.Advance()

	// Fall back to traditional mode for type parsing
	p.syncCursorToTokens()
	p.useCursor = false

	// Parse the target type (should be an interface type)
	expression.TargetType = p.parseTypeExpression()
	if expression.TargetType == nil {
		p.addError("expected type after 'as' operator", ErrExpectedType)
		p.useCursor = true
		p.syncTokensToCursor()
		return expression
	}

	// Set end position based on the target type
	expression.EndPos = expression.TargetType.End()

	// Return to cursor mode
	p.useCursor = true
	p.syncTokensToCursor()

	return expression
}

// parseImplementsExpressionCursor parses the 'implements' operator in cursor mode.
// Task 2.2.13: Cursor mode version
// Example: obj implements IMyInterface  -> Boolean
// PRE: cursor is on IMPLEMENTS token
// POST: cursor is on last token of target type
func (p *Parser) parseImplementsExpressionCursor(left ast.Expression) ast.Expression {
	implementsToken := p.cursor.Current()
	expression := &ast.ImplementsExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: implementsToken},
		},
		Left: left,
	}

	p.cursor = p.cursor.Advance()

	// Fall back to traditional mode for type parsing
	p.syncCursorToTokens()
	p.useCursor = false

	// Parse the target type (should be an interface type)
	expression.TargetType = p.parseTypeExpression()
	if expression.TargetType == nil {
		p.addError("expected type after 'implements' operator", ErrExpectedType)
		p.useCursor = true
		p.syncTokensToCursor()
		return expression
	}

	// Set end position based on the target type
	expression.EndPos = expression.TargetType.End()

	// Return to cursor mode
	p.useCursor = true
	p.syncTokensToCursor()

	return expression
}
