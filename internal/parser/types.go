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
func (p *Parser) parseArrayType() *ast.ArrayTypeNode {
	// Current token is ARRAY
	arrayToken := p.curToken

	// Check for bounds: array[low..high]
	var lowBound *int
	var highBound *int

	if p.peekTokenIs(lexer.LBRACK) {
		p.nextToken() // move to '['

		// Parse low bound (may be negative)
		p.nextToken() // move to low bound
		low, err := p.parseArrayBound()
		if err != nil {
			p.addError("invalid array lower bound: " + err.Error())
			return nil
		}
		lowBound = &low

		// Expect '..'
		if !p.expectPeek(lexer.DOTDOT) {
			p.addError("expected '..' in array bounds")
			return nil
		}

		// Parse high bound (may be negative)
		p.nextToken() // move to high bound
		high, err := p.parseArrayBound()
		if err != nil {
			p.addError("invalid array upper bound: " + err.Error())
			return nil
		}
		highBound = &high

		// Validate bounds
		if *lowBound > *highBound {
			p.addError("array lower bound cannot be greater than upper bound")
			return nil
		}

		// Expect ']'
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

	return &ast.ArrayTypeNode{
		Token:       arrayToken,
		ElementType: elementType,
		LowBound:    lowBound,
		HighBound:   highBound,
	}
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

// parseArrayBound parses an array bound which may be negative.
// Current token should be at the bound value (INT or MINUS).
// Returns the integer value and any error.
func (p *Parser) parseArrayBound() (int, error) {
	// Check for negative number
	if p.curTokenIs(lexer.MINUS) {
		// Negative bound: -N
		if !p.expectPeek(lexer.INT) {
			return 0, strconv.ErrSyntax
		}
		val, err := parseInt(p.curToken.Literal)
		if err != nil {
			return 0, err
		}
		return -val, nil
	}

	// Positive bound
	if !p.curTokenIs(lexer.INT) {
		return 0, strconv.ErrSyntax
	}
	return parseInt(p.curToken.Literal)
}
