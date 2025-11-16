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
// PRE: curToken is first token of type (IDENT, CONST, FUNCTION, PROCEDURE, ARRAY, SET, CLASS)
// POST: curToken is last token of type expression
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

	case lexer.CLASS:
		// Metaclass type: class of ClassName
		// Task 9.70: Parse metaclass type syntax
		return p.parseClassOfType()

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
// Strategy: Use lexer.Peek() to look ahead without modifying state.
// Task 12.3.4: Refactored to use Peek() instead of creating temporary lexer
// PRE: curToken is LPAREN
// POST: curToken is LPAREN (unchanged)
func (p *Parser) detectFunctionPointerFullSyntax() bool {
	// Use Peek() to look ahead through tokens
	// Peek(0) gives the token after peekToken
	// Scan through tokens looking for COLON, SEMICOLON, or RPAREN
	// If we find COLON before SEMICOLON/RPAREN, it's full syntax

	peekIndex := 0
	for {
		tok := p.peek(peekIndex)

		switch tok.Type {
		case lexer.COLON:
			// Found colon - this is full syntax
			return true
		case lexer.SEMICOLON, lexer.RPAREN, lexer.EOF:
			// Reached end without finding colon - this is shorthand
			return false
		case lexer.IDENT, lexer.COMMA:
			// Keep scanning
			peekIndex++
			continue
		default:
			// If we hit a type keyword or other token, likely shorthand
			// But keep scanning to be safe
			peekIndex++
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
// PRE: curToken is FUNCTION or PROCEDURE
// POST: curToken is last token of function pointer type (OBJECT, return type, or RPAREN)
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

	// Check if parameter list is present (optional in DWScript)
	// Function pointer types can be:
	//   procedure - no parameters, no parentheses
	//   procedure() - no parameters, with parentheses
	//   procedure(x: Integer) - with parameters
	//   function : Integer - no parameters, no parentheses
	//   function() : Integer - no parameters, with parentheses
	//   function(x: Integer) : Integer - with parameters
	hasParentheses := p.peekTokenIs(lexer.LPAREN)

	// Track the token for EndPos calculation
	var endToken lexer.Token

	if hasParentheses {
		// Parameter list present with parentheses
		p.nextToken() // move to LPAREN

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

		// Save RPAREN token for EndPos calculation
		endToken = p.curToken
	} else {
		// No parentheses - parameterless function/procedure pointer
		// Current token is still FUNCTION or PROCEDURE
		// Save the function/procedure token for EndPos
		endToken = funcOrProcToken
	}

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
			p.addError("expected 'object' after 'of' in function pointer type", ErrUnexpectedToken)
			return nil
		}
		funcPtrType.OfObject = true
		// EndPos is after "object" token
		funcPtrType.EndPos = p.endPosFromToken(p.curToken)
	} else if funcPtrType.ReturnType != nil {
		// EndPos is after return type for functions
		funcPtrType.EndPos = funcPtrType.ReturnType.End()
	} else {
		// EndPos is after closing paren (if present) or function/procedure keyword
		funcPtrType.EndPos = p.endPosFromToken(endToken)
	}

	return funcPtrType
}

// dimensionPair represents a single dimension of an array with low and high bounds.
type dimensionPair struct {
	low, high ast.Expression
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
// PRE: curToken is ARRAY
// POST: curToken is last token of element type
func (p *Parser) parseArrayType() *ast.ArrayTypeNode {
	// Current token is ARRAY
	arrayToken := p.curToken

	// Collect all dimensions (comma-separated)
	var dimensions []dimensionPair

	var indexType ast.TypeExpression

	if p.peekTokenIs(lexer.LBRACK) {
		p.nextToken() // move to '['

		// Check for enum-indexed array: array[TEnum] of Type
		// This is when we have an identifier followed directly by ']' (no '..')
		// Task 9.21.1: Support enum-indexed arrays
		if p.peekTokenIs(lexer.IDENT) {
			p.nextToken() // move to identifier

			// Check if next token is ']' (enum-indexed) or something else
			if p.peekTokenIs(lexer.RBRACK) {
				// This is an enum-indexed array: array[TEnum] of Type
				indexType = &ast.TypeAnnotation{
					Token:  p.curToken,
					Name:   p.curToken.Literal,
					EndPos: p.endPosFromToken(p.curToken),
				}

				// Move to ']'
				p.nextToken()
			} else {
				// Not enum-indexed, restore and parse as normal bounds
				// We've already moved to the identifier, so this will be the low bound
				dimensions = p.parseArrayBoundsFromCurrent()
				if dimensions == nil {
					return nil
				}

				// Now expect ']'
				if !p.expectPeek(lexer.RBRACK) {
					p.addError("expected ']' after array bounds", ErrMissingRBracket)
					return nil
				}
			}
		} else {
			// Not starting with identifier, parse normally
			p.nextToken() // move to low bound
			dimensions = p.parseArrayBoundsFromCurrent()
			if dimensions == nil {
				return nil
			}

			// Note: Bounds validation is now deferred to semantic analysis phase
			// since bounds may be constant expressions that need evaluation

			// Now expect ']'
			if !p.expectPeek(lexer.RBRACK) {
				p.addError("expected ']' after array bounds", ErrMissingRBracket)
				return nil
			}
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

	// If enum-indexed array, return with IndexType
	if indexType != nil {
		arrayNode := &ast.ArrayTypeNode{
			Token:       arrayToken,
			ElementType: elementType,
			IndexType:   indexType,
			LowBound:    nil,
			HighBound:   nil,
		}
		// EndPos is after element type
		arrayNode.EndPos = elementType.End()
		return arrayNode
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
// PRE: curToken is first token of bound expression
// POST: curToken is last token of bound expression
func (p *Parser) parseArrayBound() ast.Expression {
	// Parse as a general expression
	// This handles:
	// - Integer literals: 10
	// - Unary expressions: -5
	// - Identifiers: size
	// - Binary expressions: size - 1
	return p.parseExpression(LOWEST)
}

// parseArrayBoundsFromCurrent parses array dimensions starting from the current token.
// It expects to be positioned at the low bound of the first dimension.
// Returns a slice of dimension pairs (low, high) or nil on error.
//
// This helper function extracts the common array bound parsing logic to avoid duplication.
// PRE: curToken is first token of low bound expression
// POST: curToken is last token of last dimension's high bound
func (p *Parser) parseArrayBoundsFromCurrent() []dimensionPair {
	var dimensions []dimensionPair

	// Parse first dimension
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

	return dimensions
}

// parseClassOfType parses a metaclass type expression.
//
// Syntax:
//
//	class of ClassName
//
// Examples:
//
//	class of TMyClass
//	class of TObject
//
// Current token should be CLASS.
//
// Task 9.70: Parse metaclass type syntax
// PRE: curToken is CLASS
// POST: curToken is class type IDENT
func (p *Parser) parseClassOfType() *ast.ClassOfTypeNode {
	classToken := p.curToken // The 'class' token

	// Expect 'of' keyword
	if !p.expectPeek(lexer.OF) {
		p.addError("expected 'of' after 'class' in metaclass type", ErrMissingOf)
		return nil
	}

	// Parse class type (typically a simple identifier like TMyClass)
	p.nextToken() // move to class type

	classType := p.parseTypeExpression()
	if classType == nil {
		p.addError("expected class type after 'class of'", ErrExpectedType)
		return nil
	}

	classOfNode := &ast.ClassOfTypeNode{
		Token:     classToken,
		ClassType: classType,
	}

	// EndPos is after the class type
	classOfNode.EndPos = classType.End()

	return classOfNode
}
