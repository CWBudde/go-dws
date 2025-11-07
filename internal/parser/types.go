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
		typeAnnotation := &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}
		// EndPos is after the type identifier token
		typeAnnotation.EndPos = p.endPosFromToken(p.curToken)
		return typeAnnotation

	case lexer.CONST:
		// Special case: "const" can be used as a type in "array of const"
		// This represents a variant/heterogeneous array type
		// Task 9.21.4: Support variadic parameters with array of const
		typeAnnotation := &ast.TypeAnnotation{
			Token: p.curToken,
			Name:  "const", // This will be interpreted as Variant type by semantic analyzer
		}
		// EndPos is after the const token
		typeAnnotation.EndPos = p.endPosFromToken(p.curToken)
		return typeAnnotation

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
		p.addError("expected type expression, got "+p.curToken.Literal, ErrExpectedType)
		return nil
	}
}

// detectFunctionPointerFullSyntax determines if we have full syntax (with parameter names)
// or shorthand syntax (types only) WITHOUT advancing the parser state.
//
// This method is called when curToken is LPAREN and peekToken is the first token inside.
//
// Returns:
//   - true for full syntax: "x: Type" or "x, y: Type"
//   - false for shorthand: "Type" or "Type1, Type2"
//
// Strategy: Create a temporary lexer from the current position and peek ahead.
// Task 9.301: Simplified detection without state save/restore
func (p *Parser) detectFunctionPointerFullSyntax() bool {
	// We need to look at peekToken and tokens after it WITHOUT calling nextToken().
	// Since we can't advance the parser, we'll create a temporary lexer starting
	// from the current parser position and manually scan tokens.

	// Get the lexer's current input starting from peekToken's position
	// peekToken.Pos.Offset gives us where peekToken starts in the input
	input := p.l.Input()
	if p.peekToken.Pos.Offset < 0 || p.peekToken.Pos.Offset >= len(input) {
		// Edge case: can't determine, assume shorthand
		return false
	}

	// Create a temporary lexer starting from peekToken's position
	tempLexer := lexer.New(input[p.peekToken.Pos.Offset:])

	// Scan through tokens looking for COLON, SEMICOLON, or RPAREN
	// If we find COLON before SEMICOLON/RPAREN, it's full syntax
	for {
		tok := tempLexer.NextToken()

		switch tok.Type {
		case lexer.COLON:
			// Found colon - this is full syntax
			return true
		case lexer.SEMICOLON, lexer.RPAREN, lexer.EOF:
			// Reached end without finding colon - this is shorthand
			return false
		case lexer.IDENT, lexer.COMMA:
			// Keep scanning
			continue
		default:
			// If we hit a type keyword or other token, likely shorthand
			// But keep scanning to be safe
			continue
		}
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
		p.addError("expected '(' after "+funcOrProcToken.Literal, ErrMissingLParen)
		return nil
	}

	// Check if there are parameters (not just empty parens)
	if !p.peekTokenIs(lexer.RPAREN) {
		// Detect syntax type: full (with names) vs shorthand (types only)
		// We need to determine if we have:
		//   Full syntax: "name: Type" or "name1, name2: Type"
		//   Shorthand: "Type" or "Type1, Type2"
		//
		// Strategy: Use simple lookahead WITHOUT advancing parser state.
		// After we detect, advance once and parse accordingly.

		isFullSyntax := p.detectFunctionPointerFullSyntax()

		// Now advance to first parameter/type token
		p.nextToken()

		if isFullSyntax {
			// Full syntax with parameter names
			funcPtrType.Parameters = p.parseParameterListAtToken()
		} else {
			// Shorthand syntax with only types
			funcPtrType.Parameters = p.parseTypeOnlyParameterListAtToken()
		}

		if funcPtrType.Parameters == nil {
			return nil
		}
	} else {
		// Empty parameter list
		p.nextToken() // move to RPAREN
	}

	// Expect closing parenthesis
	if !p.curTokenIs(lexer.RPAREN) {
		p.addError("expected ')' after parameter list in function pointer type", ErrMissingRParen)
		return nil
	}

	// Save RPAREN token for EndPos calculation (for procedures without return type)
	rparenToken := p.curToken

	// Parse return type for functions (not procedures)
	if isFunction {
		// Expect colon and return type
		if !p.expectPeek(lexer.COLON) {
			p.addError("expected ':' after ')' in function pointer type", ErrMissingColon)
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
			p.addError("complex return types not yet supported in function pointers", ErrInvalidType)
			return nil
		}
	}

	// Check for "of object" clause (method pointers)
	if p.peekTokenIs(lexer.OF) {
		p.nextToken() // move to OF
		if !p.expectPeek(lexer.OBJECT) {
			p.addError("expected .object. after .of. in function pointer type", ErrUnexpectedToken)
			return nil
		}
		funcPtrType.OfObject = true
		// EndPos is after "object" token
		funcPtrType.EndPos = p.endPosFromToken(p.curToken)
	} else if funcPtrType.ReturnType != nil {
		// EndPos is after return type for functions
		funcPtrType.EndPos = funcPtrType.ReturnType.End()
	} else {
		// EndPos is after closing paren for procedures without "of object"
		funcPtrType.EndPos = p.endPosFromToken(rparenToken)
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
			p.addError("invalid array lower bound expression", ErrInvalidExpression)
			return nil
		}

		// Expect '..'
		if !p.expectPeek(lexer.DOTDOT) {
			p.addError("expected .... in array bounds", ErrUnexpectedToken)
			return nil
		}

		// Parse high bound expression
		p.nextToken() // move to high bound
		highBound := p.parseArrayBound()
		if highBound == nil {
			p.addError("invalid array upper bound expression", ErrInvalidExpression)
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
				p.addError("invalid array lower bound expression in multi-dimensional array", ErrInvalidExpression)
				return nil
			}

			if !p.expectPeek(lexer.DOTDOT) {
				p.addError("expected .... in array bounds", ErrUnexpectedToken)
				return nil
			}

			p.nextToken() // move to high bound
			highBound := p.parseArrayBound()
			if highBound == nil {
				p.addError("invalid array upper bound expression in multi-dimensional array", ErrInvalidExpression)
				return nil
			}

			dimensions = append(dimensions, dimensionPair{lowBound, highBound})
		}

		// Note: Bounds validation is now deferred to semantic analysis phase
		// since bounds may be constant expressions that need evaluation

		// Now expect ']'
		if !p.expectPeek(lexer.RBRACK) {
			p.addError("expected ']' after array bounds", ErrMissingRBracket)
			return nil
		}
	}

	// Expect 'of' keyword
	if !p.expectPeek(lexer.OF) {
		p.addError("expected .of. after .array. or .array[bounds].", ErrMissingOf)
		return nil
	}

	// Parse element type
	p.nextToken() // move to element type
	elementType := p.parseTypeExpression()
	if elementType == nil {
		p.addError("expected type expression after .array of.", ErrExpectedType)
		return nil
	}

	// If no dimensions, return simple dynamic array
	if len(dimensions) == 0 {
		arrayNode := &ast.ArrayTypeNode{
			Token:       arrayToken,
			ElementType: elementType,
			LowBound:    nil,
			HighBound:   nil,
		}
		// EndPos is after element type
		arrayNode.EndPos = elementType.End()
		return arrayNode
	}

	// Build nested array types from innermost to outermost
	// This desugars: array[0..1, 0..2] of Integer
	//           into: array[0..1] of (array[0..2] of Integer)
	result := elementType
	for i := len(dimensions) - 1; i >= 0; i-- {
		arrayNode := &ast.ArrayTypeNode{
			Token:       arrayToken,
			ElementType: result,
			LowBound:    dimensions[i].low,
			HighBound:   dimensions[i].high,
		}
		// EndPos is after the element type (which could be nested)
		arrayNode.EndPos = result.End()
		result = arrayNode
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
