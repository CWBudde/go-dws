package parser

import (
	"strconv"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseTypeExpression is a dispatcher that routes to the appropriate implementation
// based on the parser mode (traditional vs cursor).
//
// Task 2.7.2: This dispatcher enables dual-mode operation during migration.
// Eventually (Phase 2.7), only the cursor version will remain.
func (p *Parser) parseTypeExpression() ast.TypeExpression {
	return p.parseTypeExpressionCursor()
}

// parseTypeExpressionCursor parses a type expression using cursor mode.
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
// Task 2.7.2: Migrated to cursor mode
// PRE: cursor is on first token of type (IDENT, CONST, FUNCTION, PROCEDURE, ARRAY, SET, CLASS)
// POST: cursor is on last token of type expression
func (p *Parser) parseTypeExpressionCursor() ast.TypeExpression {
	cursor := p.cursor
	builder := p.StartNode()
	currentToken := cursor.Current()

	switch currentToken.Type {
	case lexer.IDENT:
		// Simple type identifier
		typeAnnotation := &ast.TypeAnnotation{
			Token: currentToken,
			Name:  currentToken.Literal,
		}
		// EndPos is after the type identifier token
		return builder.Finish(typeAnnotation).(*ast.TypeAnnotation)

	case lexer.CONST:
		// Special case: "const" can be used as a type in "array of const"
		typeAnnotation := &ast.TypeAnnotation{
			Token: currentToken,
			Name:  "const",
		}
		// EndPos is after the const token
		return builder.Finish(typeAnnotation).(*ast.TypeAnnotation)

	case lexer.FUNCTION, lexer.PROCEDURE:
		// Inline function or procedure pointer type
		return p.parseFunctionPointerType()

	case lexer.ARRAY:
		// Array type: array of ElementType
		return p.parseArrayType()

	case lexer.SET:
		// Set type: set of ElementType
		return p.parseSetType()

	case lexer.CLASS:
		// Metaclass type: class of ClassName
		return p.parseClassOfType()

	default:
		p.addError("expected type expression, got "+currentToken.Literal, ErrExpectedType)
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

// parseFunctionPointerType is a dispatcher that routes to the appropriate implementation
// based on the parser mode (traditional vs cursor).
//
// Task 2.7.2: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseFunctionPointerType() *ast.FunctionPointerTypeNode {
	return p.parseFunctionPointerTypeCursor()
}

// parseFunctionPointerTypeCursor parses an inline function or procedure pointer type using cursor mode.
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
// Task 2.7.2: Migrated to cursor mode
// PRE: cursor is on FUNCTION or PROCEDURE token
// POST: cursor is on last token of function pointer type (OBJECT, return type, or RPAREN)
func (p *Parser) parseFunctionPointerTypeCursor() *ast.FunctionPointerTypeNode {
	cursor := p.cursor
	builder := p.StartNode()

	// Current token is FUNCTION or PROCEDURE
	funcOrProcToken := cursor.Current()
	isFunction := funcOrProcToken.Type == lexer.FUNCTION

	// Create the function pointer type node
	funcPtrType := &ast.FunctionPointerTypeNode{
		Token:      funcOrProcToken,
		Parameters: []*ast.Parameter{},
		OfObject:   false,
	}

	// Check if parameter list is present
	nextToken := cursor.Peek(1)
	hasParentheses := nextToken.Type == lexer.LPAREN

	// Track the token for EndPos calculation
	var endToken lexer.Token

	if hasParentheses {
		// Parameter list present with parentheses
		cursor = cursor.Advance() // move to LPAREN
		p.cursor = cursor

		// Check if there are parameters (not just empty parens)
		if cursor.Peek(1).Type != lexer.RPAREN {
			// Detect syntax type using lookahead
			isFullSyntax := p.detectFunctionPointerFullSyntax()

			// Advance to first parameter/type token
			cursor = cursor.Advance()
			p.cursor = cursor

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

			// Update cursor after parameter parsing
			cursor = p.cursor
		} else {
			// Empty parameter list
			cursor = cursor.Advance() // move to RPAREN
			p.cursor = cursor
		}

		// Expect closing parenthesis
		if cursor.Current().Type != lexer.RPAREN {
			p.addError("expected ')' after parameter list in function pointer type", ErrMissingRParen)
			return nil
		}

		// Save RPAREN token for EndPos calculation
		endToken = cursor.Current()
	} else {
		// No parentheses - parameterless function/procedure pointer
		// Current token is still FUNCTION or PROCEDURE
		// Save the function/procedure token for EndPos
		endToken = funcOrProcToken
	}

	// Parse return type for functions (not procedures)
	if isFunction {
		// Expect colon and return type
		if cursor.Peek(1).Type != lexer.COLON {
			p.addError("expected ':' after ')' in function pointer type", ErrMissingColon)
			return nil
		}
		cursor = cursor.Advance() // move to COLON
		p.cursor = cursor

		// Parse return type (can be any type expression)
		cursor = cursor.Advance() // move to return type
		p.cursor = cursor

		returnTypeExpr := p.parseTypeExpression()
		if returnTypeExpr == nil {
			return nil
		}

		// Convert type expression to TypeAnnotation
		switch rt := returnTypeExpr.(type) {
		case *ast.TypeAnnotation:
			funcPtrType.ReturnType = rt
		default:
			p.addError("complex return types not yet supported in function pointers", ErrInvalidType)
			return nil
		}

		// Update cursor after type expression parsing
		cursor = p.cursor
	}

	// Check for "of object" clause (method pointers)
	if cursor.Peek(1).Type == lexer.OF {
		cursor = cursor.Advance() // move to OF
		if cursor.Peek(1).Type != lexer.OBJECT {
			p.addError("expected 'object' after 'of' in function pointer type", ErrUnexpectedToken)
			return nil
		}
		cursor = cursor.Advance() // move to OBJECT
		p.cursor = cursor
		funcPtrType.OfObject = true
		// EndPos is after "object" token
		return builder.Finish(funcPtrType).(*ast.FunctionPointerTypeNode)
	} else if funcPtrType.ReturnType != nil {
		// EndPos is after return type for functions
		return builder.FinishWithNode(funcPtrType, funcPtrType.ReturnType).(*ast.FunctionPointerTypeNode)
	} else {
		// EndPos is after closing paren (if present) or function/procedure keyword
		return builder.FinishWithToken(funcPtrType, endToken).(*ast.FunctionPointerTypeNode)
	}
}

// dimensionPair represents a single dimension of an array with low and high bounds.
type dimensionPair struct {
	low, high ast.Expression
}

// parseArrayType is a dispatcher that routes to the appropriate implementation
// based on the parser mode (traditional vs cursor).
//
// Task 2.7.2: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseArrayType() *ast.ArrayTypeNode {
	return p.parseArrayTypeCursor()
}

// parseArrayTypeCursor parses an array type expression using cursor mode.
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
// Task 2.7.2: Migrated to cursor mode
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
//
// PRE: cursor is on ARRAY token
// POST: cursor is on last token of element type
func (p *Parser) parseArrayTypeCursor() *ast.ArrayTypeNode {
	cursor := p.cursor
	builder := p.StartNode()

	// Current token is ARRAY
	arrayToken := cursor.Current()

	// Collect all dimensions (comma-separated)
	var dimensions []dimensionPair
	var indexType ast.TypeExpression

	if cursor.Peek(1).Type == lexer.LBRACK {
		cursor = cursor.Advance() // move to '['
		p.cursor = cursor

		// Check for enum-indexed array: array[TEnum] of Type
		if cursor.Peek(1).Type == lexer.IDENT {
			cursor = cursor.Advance() // move to identifier
			p.cursor = cursor

			// Check if next token is ']' (enum-indexed) or something else
			if cursor.Peek(1).Type == lexer.RBRACK {
				// This is an enum-indexed array: array[TEnum] of Type
				typeBuilder := p.StartNode()
				indexType = typeBuilder.Finish(&ast.TypeAnnotation{
					Token: cursor.Current(),
					Name:  cursor.Current().Literal,
				}).(*ast.TypeAnnotation)

				// Move to ']'
				cursor = cursor.Advance()
				p.cursor = cursor
			} else {
				// Not enum-indexed, parse as normal bounds
				// We've already moved to the identifier, so this will be the low bound
				dimensions = p.parseArrayBoundsFromCurrent()
				if dimensions == nil {
					return nil
				}

				// Update cursor after bounds parsing
				cursor = p.cursor

				// Now expect ']'
				if cursor.Peek(1).Type != lexer.RBRACK {
					p.addError("expected ']' after array bounds", ErrMissingRBracket)
					return nil
				}
				cursor = cursor.Advance() // move to ']'
				p.cursor = cursor
			}
		} else {
			// Not starting with identifier, parse normally
			cursor = cursor.Advance() // move to low bound
			p.cursor = cursor

			dimensions = p.parseArrayBoundsFromCurrent()
			if dimensions == nil {
				return nil
			}

			// Update cursor after bounds parsing
			cursor = p.cursor

			// Now expect ']'
			if cursor.Peek(1).Type != lexer.RBRACK {
				// Use structured error for missing closing bracket
				err := NewStructuredError(ErrKindMissing).
					WithCode(ErrMissingRBracket).
					WithMessage("expected ']' after array bounds").
					WithPosition(cursor.Peek(1).Pos, cursor.Peek(1).Length()).
					WithExpected(lexer.RBRACK).
					WithActual(cursor.Peek(1).Type, cursor.Peek(1).Literal).
					WithSuggestion("add ']' to close the array bounds").
					WithParsePhase("array type bounds").
					Build()
				p.addStructuredError(err)
				return nil
			}
			cursor = cursor.Advance() // move to ']'
			p.cursor = cursor
		}
	}

	// Expect 'of' keyword
	if cursor.Peek(1).Type != lexer.OF {
		// Use structured error for missing 'of'
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingOf).
			WithMessage("expected 'of' after array declaration").
			WithPosition(cursor.Peek(1).Pos, cursor.Peek(1).Length()).
			WithExpected(lexer.OF).
			WithActual(cursor.Peek(1).Type, cursor.Peek(1).Literal).
			WithSuggestion("add 'of' keyword after 'array' or 'array[bounds]'").
			WithNote("DWScript array types use syntax: array [bounds] of ElementType").
			WithParsePhase("array type").
			Build()
		p.addStructuredError(err)
		return nil
	}
	cursor = cursor.Advance() // move to OF
	p.cursor = cursor

	// Parse element type
	cursor = cursor.Advance() // move to element type
	p.cursor = cursor

	elementType := p.parseTypeExpression()
	if elementType == nil {
		// Use structured error for missing element type
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrExpectedType).
			WithMessage("expected type expression after 'array of'").
			WithPosition(cursor.Current().Pos, cursor.Current().Length()).
			WithExpectedString("type name").
			WithSuggestion("specify the element type, like 'Integer' or 'String'").
			WithParsePhase("array element type").
			Build()
		p.addStructuredError(err)
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
		return builder.FinishWithNode(arrayNode, elementType).(*ast.ArrayTypeNode)
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
		return builder.FinishWithNode(arrayNode, elementType).(*ast.ArrayTypeNode)
	}

	// Build nested array types from innermost to outermost
	// This desugars: array[0..1, 0..2] of Integer
	//           into: array[0..1] of (array[0..2] of Integer)
	result := elementType
	for i := len(dimensions) - 1; i >= 0; i-- {
		dimBuilder := p.StartNode()
		arrayNode := &ast.ArrayTypeNode{
			Token:       arrayToken,
			ElementType: result,
			LowBound:    dimensions[i].low,
			HighBound:   dimensions[i].high,
		}
		// EndPos is after the element type (which could be nested)
		result = dimBuilder.FinishWithNode(arrayNode, result).(*ast.ArrayTypeNode)
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

// parseClassOfType is a dispatcher that routes to the appropriate implementation
// based on the parser mode (traditional vs cursor).
//
// Task 2.7.2: This dispatcher enables dual-mode operation during migration.
func (p *Parser) parseClassOfType() *ast.ClassOfTypeNode {
	return p.parseClassOfTypeCursor()
}

// parseClassOfTypeCursor parses a metaclass type expression using cursor mode.
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
// Task 2.7.2: Migrated to cursor mode
// PRE: cursor is on CLASS token
// POST: cursor is on class type IDENT
func (p *Parser) parseClassOfTypeCursor() *ast.ClassOfTypeNode {
	cursor := p.cursor
	builder := p.StartNode()

	classToken := cursor.Current() // The 'class' token

	// Expect 'of' keyword
	if cursor.Peek(1).Type != lexer.OF {
		p.addError("expected 'of' after 'class' in metaclass type", ErrMissingOf)
		return nil
	}
	cursor = cursor.Advance() // move to OF
	p.cursor = cursor

	// Parse class type (typically a simple identifier like TMyClass)
	cursor = cursor.Advance() // move to class type
	p.cursor = cursor

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
	return builder.FinishWithNode(classOfNode, classType).(*ast.ClassOfTypeNode)
}
