package parser

import (
	"strconv"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseTypeExpression parses a type expression.
// Type expressions can be:
//   - Simple types: Integer, String, TMyType
//   - Function pointer types: function(x: Integer): String
//   - Procedure pointer types: procedure(msg: String)
//   - Array types: array of Integer
//   - Nested arrays: array of array of String
//   - Complex combinations: array of function(x: Integer): Boolean
//
// This unified parser enables inline type syntax in parameters and variables
// without requiring type aliases.
//
// Task 9.49: Created to support inline type expressions
func (p *Parser) parseTypeExpression() ast.TypeExpression {
	switch p.curToken.Type {
	case lexer.IDENT:
		// Simple type identifier
		return &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}

	case lexer.FUNCTION, lexer.PROCEDURE:
		// Inline function or procedure pointer type
		return p.parseFunctionPointerType()

	case lexer.ARRAY:
		// Array type: array of ElementType
		return p.parseArrayType()

	case lexer.SET:
		// Set type: set of ElementType
		// Task 9.213: Parse inline set type expressions
		return p.parseSetType()

	default:
		p.addError("expected type expression, got " + p.curToken.Literal)
		return nil
	}
}

// parseFunctionPointerType parses an inline function or procedure pointer type.
// This is the reusable version extracted from parseFunctionPointerTypeDeclaration.
//
// Syntax:
//
//	function(param1: Type1, param2: Type2, ...): ReturnType
//	procedure(param1: Type1, param2: Type2, ...)
//	function(...): ReturnType of object
//	procedure(...) of object
//
// Task 9.50: Refactored from parseFunctionPointerTypeDeclaration
func (p *Parser) parseFunctionPointerType() *ast.FunctionPointerTypeNode {
	// Current token is FUNCTION or PROCEDURE
	funcOrProcToken := p.curToken
	isFunction := funcOrProcToken.Type == lexer.FUNCTION

	// Create the function pointer type node
	funcPtrType := &ast.FunctionPointerTypeNode{
		Token:      funcOrProcToken,
		Parameters: []*ast.Parameter{},
		OfObject:   false,
	}

	// Expect opening parenthesis for parameter list
	if !p.expectPeek(lexer.LPAREN) {
		p.addError("expected '(' after " + funcOrProcToken.Literal)
		return nil
	}

	// Check if there are parameters (not just empty parens)
	if !p.peekTokenIs(lexer.RPAREN) {
		// Parse parameter list using existing function
		funcPtrType.Parameters = p.parseParameterList()
		if funcPtrType.Parameters == nil {
			return nil
		}
	} else {
		// Empty parameter list
		p.nextToken() // move to RPAREN
	}

	// Expect closing parenthesis
	if !p.curTokenIs(lexer.RPAREN) {
		p.addError("expected ')' after parameter list in function pointer type")
		return nil
	}

	// Parse return type for functions (not procedures)
	if isFunction {
		// Expect colon and return type
		if !p.expectPeek(lexer.COLON) {
			p.addError("expected ':' after ')' in function pointer type")
			return nil
		}

		// Parse return type (can be any type expression)
		p.nextToken() // move to return type
		returnTypeExpr := p.parseTypeExpression()
		if returnTypeExpr == nil {
			return nil
		}

		// Convert type expression to TypeAnnotation
		// For now, we only support simple types as return types
		// TODO: Support complex return types (arrays, function pointers)
		switch rt := returnTypeExpr.(type) {
		case *ast.TypeAnnotation:
			funcPtrType.ReturnType = rt
		default:
			p.addError("complex return types not yet supported in function pointers")
			return nil
		}
	}

	// Check for "of object" clause (method pointers)
	if p.peekTokenIs(lexer.OF) {
		p.nextToken() // move to OF
		if !p.expectPeek(lexer.OBJECT) {
			p.addError("expected 'object' after 'of' in function pointer type")
			return nil
		}
		funcPtrType.OfObject = true
	}

	return funcPtrType
}

// parseArrayType parses an array type expression.
// Supports both dynamic and static arrays:
//   - Dynamic: array of ElementType
//   - Static: array[low..high] of ElementType
//
// ElementType can be any type expression:
//   - Simple type: array of Integer
//   - Array type: array of array of String (nested)
//   - Function pointer: array of function(x: Integer): Boolean
//   - Static nested: array[1..5] of array[1..10] of Integer
//
// Task 9.51: Created to support array of Type syntax
// Task 9.54: Extended to support static array bounds
// Task 9.212: Extended to support comma-separated multidimensional arrays
//
// Supports both single and multi-dimensional syntax:
//   - array[0..10] of Integer         (single dimension)
//   - array[0..1, 0..2] of Integer    (2D, comma-separated)
//   - array[1..3, 1..4, 1..5] of Float (3D, comma-separated)
//
// Multi-dimensional arrays are desugared into nested array types:
//
//	array[0..1, 0..2] of Integer
//	→ array[0..1] of array[0..2] of Integer
func (p *Parser) parseArrayType() *ast.ArrayTypeNode {
	// Current token is ARRAY
	arrayToken := p.curToken

	// Collect all dimensions (comma-separated)
	type dimensionPair struct {
		low, high ast.Expression
	}
	var dimensions []dimensionPair

	if p.peekTokenIs(lexer.LBRACK) {
		p.nextToken() // move to '['

		// Parse first dimension
		p.nextToken() // move to low bound
		lowBound := p.parseArrayBound()
		if lowBound == nil {
			p.addError("invalid array lower bound expression")
			return nil
		}

		// Expect '..'
		if !p.expectPeek(lexer.DOTDOT) {
			p.addError("expected '..' in array bounds")
			return nil
		}

		// Parse high bound expression
		p.nextToken() // move to high bound
		highBound := p.parseArrayBound()
		if highBound == nil {
			p.addError("invalid array upper bound expression")
			return nil
		}

		dimensions = append(dimensions, dimensionPair{lowBound, highBound})

		// Parse additional dimensions (comma-separated)
		// Task 9.212: Support multidimensional arrays like array[0..1, 0..2]
		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume comma
			p.nextToken() // move to next low bound
			lowBound := p.parseArrayBound()
			if lowBound == nil {
				p.addError("invalid array lower bound expression in multi-dimensional array")
				return nil
			}

			if !p.expectPeek(lexer.DOTDOT) {
				p.addError("expected '..' in array bounds")
				return nil
			}

			p.nextToken() // move to high bound
			highBound := p.parseArrayBound()
			if highBound == nil {
				p.addError("invalid array upper bound expression in multi-dimensional array")
				return nil
			}

			dimensions = append(dimensions, dimensionPair{lowBound, highBound})
		}

		// Note: Bounds validation is now deferred to semantic analysis phase
		// since bounds may be constant expressions that need evaluation

		// Now expect ']'
		if !p.expectPeek(lexer.RBRACK) {
			p.addError("expected ']' after array bounds")
			return nil
		}
	}

	// Expect 'of' keyword
	if !p.expectPeek(lexer.OF) {
		p.addError("expected 'of' after 'array' or 'array[bounds]'")
		return nil
	}

	// Parse element type
	p.nextToken() // move to element type
	elementType := p.parseTypeExpression()
	if elementType == nil {
		p.addError("expected type expression after 'array of'")
		return nil
	}

	// If no dimensions, return simple dynamic array
	if len(dimensions) == 0 {
		return &ast.ArrayTypeNode{
			Token:       arrayToken,
			ElementType: elementType,
			LowBound:    nil,
			HighBound:   nil,
		}
	}

	// Build nested array types from innermost to outermost
	// This desugars: array[0..1, 0..2] of Integer
	//           into: array[0..1] of (array[0..2] of Integer)
	result := elementType
	for i := len(dimensions) - 1; i >= 0; i-- {
		result = &ast.ArrayTypeNode{
			Token:       arrayToken,
			ElementType: result,
			LowBound:    dimensions[i].low,
			HighBound:   dimensions[i].high,
		}
	}

	return result.(*ast.ArrayTypeNode)
}

// parseInt parses a string as an integer.
// Helper function for parsing array bounds.
func parseInt(s string) (int, error) {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

// parseArrayBound parses an array bound expression.
// Array bounds can be:
// - Integer literals: 10, -5
// - Constant identifiers: size, maxIndex
// - Constant expressions: size - 1, maxIndex + 10
//
// Current token should be at the start of the bound expression.
// Returns the parsed expression or nil on error.
//
// Task 9.205: Changed to return ast.Expression instead of int to support const expressions
func (p *Parser) parseArrayBound() ast.Expression {
	// Parse as a general expression
	// This handles:
	// - Integer literals: 10
	// - Unary expressions: -5
	// - Identifiers: size
	// - Binary expressions: size - 1
	return p.parseExpression(LOWEST)
}
