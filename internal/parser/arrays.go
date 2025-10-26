package parser

import (
	"strconv"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseArrayDeclaration parses an array type declaration.
// Called after 'type Name =' has already been parsed.
// Current token should be 'array'.
//
// Syntax:
//   - type TMyArray = array[1..10] of Integer;  (static array with bounds)
//   - type TDynamic = array of String;          (dynamic array without bounds)
//
// Task 8.122: Parse array type declarations
func (p *Parser) parseArrayDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.ArrayDecl {
	arrayDecl := &ast.ArrayDecl{
		Token: typeToken, // The 'type' token
		Name:  nameIdent,
	}

	arrayToken := p.curToken // Save 'array' token

	// Check for bounds: array[low..high]
	var lowBound *int
	var highBound *int

	if p.peekTokenIs(lexer.LBRACK) {
		p.nextToken() // move to '['

		// Parse low bound
		if !p.expectPeek(lexer.INT) {
			p.addError("expected integer for array lower bound")
			return nil
		}

		low, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
		if err != nil {
			p.addError("invalid integer for array lower bound")
			return nil
		}
		lowInt := int(low)
		lowBound = &lowInt

		// Expect '..'
		if !p.expectPeek(lexer.DOTDOT) {
			p.addError("expected '..' in array bounds")
			return nil
		}

		// Parse high bound
		if !p.expectPeek(lexer.INT) {
			p.addError("expected integer for array upper bound")
			return nil
		}

		high, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
		if err != nil {
			p.addError("invalid integer for array upper bound")
			return nil
		}
		highInt := int(high)
		highBound = &highInt

		// Expect ']'
		if !p.expectPeek(lexer.RBRACK) {
			return nil
		}
	}

	// Expect 'of'
	if !p.expectPeek(lexer.OF) {
		return nil
	}

	// Parse element type
	if !p.expectPeek(lexer.IDENT) {
		p.addError("expected type identifier after 'of' in array declaration")
		return nil
	}

	elementType := &ast.TypeAnnotation{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	// Create ArrayTypeAnnotation and assign to ArrayDecl
	arrayDecl.ArrayType = &ast.ArrayTypeAnnotation{
		Token:       arrayToken,
		ElementType: elementType,
		LowBound:    lowBound,
		HighBound:   highBound,
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return arrayDecl
}

// parseIndexExpression parses an array/string indexing operation.
// Called when we encounter '[' as an infix operator.
// Left side is the array/string being indexed.
//
// Syntax:
//   - arr[i]      (variable index)
//   - arr[0]      (literal index)
//   - arr[i + 1]  (expression index)
//   - arr[i][j]   (nested indexing)
//
// Task 8.124: Parse array indexing expressions
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	indexExpr := &ast.IndexExpression{
		Token: p.curToken, // The '[' token
		Left:  left,
	}

	// Move to index expression
	p.nextToken()

	// Parse the index expression
	indexExpr.Index = p.parseExpression(LOWEST)

	// Expect ']'
	if !p.expectPeek(lexer.RBRACK) {
		return nil
	}

	return indexExpr
}
