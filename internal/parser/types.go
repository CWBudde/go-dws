package parser

import (
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

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
func (p *Parser) parseTypeExpression() ast.TypeExpression {
	cursor := p.cursor
	builder := p.StartNode()
	currentToken := cursor.Current()

	switch currentToken.Type {
	case lexer.IDENT:
		// Simple or qualified type identifier (supports nested types like TOuter.TInner)
		parts := []string{currentToken.Literal}
		endToken := currentToken
		// Consume dotted identifiers to build a qualified type name
		for cursor.Peek(1).Type == lexer.DOT && cursor.Peek(2).Type == lexer.IDENT {
			cursor = cursor.Advance() // move to '.'
			cursor = cursor.Advance() // move to next ident
			endToken = cursor.Current()
			parts = append(parts, endToken.Literal)
		}
		p.cursor = cursor

		qualifiedName := strings.Join(parts, ".")
		typeAnnotation := &ast.TypeAnnotation{
			Token:  currentToken,
			Name:   qualifiedName,
			EndPos: p.endPosFromToken(endToken),
		}
		result, _ := builder.Finish(typeAnnotation).(*ast.TypeAnnotation)
		return result

	case lexer.CONST:
		// Special case: "const" can be used as a type in "array of const"
		typeAnnotation := &ast.TypeAnnotation{
			Token: currentToken,
			Name:  "const",
		}
		// EndPos is after the const token
		result, _ := builder.Finish(typeAnnotation).(*ast.TypeAnnotation)
		return result

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
		return &ast.InvalidTypeExpression{
			BaseNode: ast.BaseNode{
				Token: currentToken,
			},
			Reason: "type expected",
		}
	}
}

func isInvalidTypeExpression(typeExpr ast.TypeExpression) bool {
	if typeExpr == nil {
		return true
	}
	_, ok := typeExpr.(*ast.InvalidTypeExpression)
	return ok
}

func invalidTypeExpression(tok lexer.Token, reason string) *ast.InvalidTypeExpression {
	return &ast.InvalidTypeExpression{
		BaseNode: ast.BaseNode{
			Token: tok,
		},
		Reason: reason,
	}
}

func isTypeExpressionStartToken(t lexer.TokenType) bool {
	switch t {
	case lexer.IDENT, lexer.CONST, lexer.FUNCTION, lexer.PROCEDURE, lexer.ARRAY, lexer.SET, lexer.CLASS:
		return true
	default:
		return false
	}
}

func isOpenArrayConstType(typeExpr ast.TypeExpression) bool {
	typeAnnotation, ok := typeExpr.(*ast.TypeAnnotation)
	return ok && ident.Equal(typeAnnotation.Name, "const")
}

func (p *Parser) recoverArrayType() {
	cursor := p.cursor

	for {
		nextToken := cursor.Peek(1)

		switch nextToken.Type {
		case lexer.OF:
			cursor = cursor.Advance()
			p.cursor = cursor

			if !isTypeExpressionStartToken(cursor.Peek(1).Type) {
				return
			}

			cursor = cursor.Advance()
			p.cursor = cursor
			_ = p.parseTypeExpression()
			return

		case lexer.SEMICOLON, lexer.EQ, lexer.ASSIGN, lexer.COMMA, lexer.RPAREN, lexer.EOF:
			return

		default:
			cursor = cursor.Advance()
			p.cursor = cursor
		}
	}
}

func (p *Parser) addPeekTokenError(message string, code string) {
	peekTok := p.cursor.Peek(1)
	err := NewParserError(
		peekTok.Pos,
		peekTok.Length(),
		message,
		code,
	)
	p.errors = append(p.errors, err)
}

// detectFunctionPointerFullSyntax determines if we have full syntax (with parameter names)
// or shorthand syntax (types only) WITHOUT advancing the parser state.
//
// Returns:
//   - true for full syntax: "x: Type" or "x, y: Type"
//   - false for shorthand: "Type" or "Type1, Type2"
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

// Syntax:
//
//	function(param1: Type1, param2: Type2, ...): ReturnType
//	procedure(param1: Type1, param2: Type2, ...)
//	function(...): ReturnType of object
//	procedure(...) of object
func (p *Parser) parseFunctionPointerType() *ast.FunctionPointerTypeNode {
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
		result, _ := builder.Finish(funcPtrType).(*ast.FunctionPointerTypeNode)
		return result
	} else if funcPtrType.ReturnType != nil {
		// EndPos is after return type for functions
		result, _ := builder.FinishWithNode(funcPtrType, funcPtrType.ReturnType).(*ast.FunctionPointerTypeNode)
		return result
	} else {
		// EndPos is after closing paren (if present) or function/procedure keyword
		result, _ := builder.FinishWithToken(funcPtrType, endToken).(*ast.FunctionPointerTypeNode)
		return result
	}
}

// dimensionPair represents a single dimension of an array with low and high bounds.
type dimensionPair struct {
	low, high ast.Expression
}

// Supports both dynamic and static arrays:
//   - Dynamic: array of ElementType
//   - Static: array[low..high] of ElementType
//   - Multidimensional: array[0..1, 0..2] of Integer
//
// Multi-dimensional arrays are desugared into nested array types:
//
//	array[0..1, 0..2] of Integer → array[0..1] of array[0..2] of Integer
func (p *Parser) parseArrayType() ast.TypeExpression {
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
					p.addPeekTokenError("\"]\" expected", ErrMissingRBracket)
					p.recoverArrayType()
					return invalidTypeExpression(arrayToken, "invalid array type")
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
					WithMessage("\"]\" expected").
					WithPosition(cursor.Peek(1).Pos, cursor.Peek(1).Length()).
					WithExpected(lexer.RBRACK).
					WithActual(cursor.Peek(1).Type, cursor.Peek(1).Literal).
					WithSuggestion("add ']' to close the array bounds").
					WithParsePhase("array type bounds").
					Build()
				p.addStructuredError(err)
				p.recoverArrayType()
				return invalidTypeExpression(arrayToken, "invalid array type")
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
			WithMessage("OF expected").
			WithPosition(cursor.Peek(1).Pos, cursor.Peek(1).Length()).
			WithExpected(lexer.OF).
			WithActual(cursor.Peek(1).Type, cursor.Peek(1).Literal).
			WithSuggestion("add 'of' keyword after 'array' or 'array[bounds]'").
			WithNote("DWScript array types use syntax: array [bounds] of ElementType").
			WithParsePhase("array type").
			Build()
		p.addStructuredError(err)
		p.recoverArrayType()
		return invalidTypeExpression(arrayToken, "invalid array type")
	}
	cursor = cursor.Advance() // move to OF
	p.cursor = cursor

	// Parse element type
	cursor = cursor.Advance() // move to element type
	p.cursor = cursor

	errorCount := len(p.errors)
	elementType := p.parseTypeExpression()
	if isInvalidTypeExpression(elementType) {
		if len(p.errors) == errorCount {
			err := NewStructuredError(ErrKindInvalid).
				WithCode(ErrExpectedType).
				WithMessage("expected type expression after 'array of'").
				WithPosition(cursor.Current().Pos, cursor.Current().Length()).
				WithExpectedString("type name").
				WithSuggestion("specify the element type, like 'Integer' or 'String'").
				WithParsePhase("array element type").
				Build()
			p.addStructuredError(err)
		}
		return invalidTypeExpression(arrayToken, "invalid array type")
	}

	if len(dimensions) > 0 && isOpenArrayConstType(elementType) {
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidType).
			WithMessage("No indices expected for open array").
			WithPosition(elementType.Pos(), elementType.End().Column-elementType.Pos().Column).
			WithParsePhase("array type").
			Build()
		p.addStructuredError(err)
		return invalidTypeExpression(arrayToken, "invalid open array type")
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
		result, _ := builder.FinishWithNode(arrayNode, elementType).(*ast.ArrayTypeNode)
		return result
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
		result, _ := builder.FinishWithNode(arrayNode, elementType).(*ast.ArrayTypeNode)
		return result
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

// parseArrayBound parses an array bound expression (integer literals,
// identifiers, or constant expressions like size - 1).
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
// Returns a slice of dimension pairs (low, high) or nil on error.
func (p *Parser) parseArrayBoundsFromCurrent() []dimensionPair {
	var dimensions []dimensionPair

	// Parse first dimension
	lowBound := p.parseArrayBound()
	if isInvalidExpression(lowBound) {
		p.addError("invalid array lower bound expression", ErrInvalidExpression)
		return nil
	}

	// Expect '..'
	if !p.peekTokenIs(lexer.DOTDOT) {
		p.addPeekTokenError("\"..\" expected", ErrUnexpectedToken)
		return []dimensionPair{{
			low: lowBound,
			high: &ast.InvalidExpression{
				Reason: "missing upper array bound",
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{Token: p.cursor.Peek(1)},
				},
			},
		}}
	}
	p.nextToken()

	// Parse high bound expression
	p.nextToken() // move to high bound
	highBound := p.parseArrayBound()
	if isInvalidExpression(highBound) {
		p.addError("invalid array upper bound expression", ErrInvalidExpression)
		return nil
	}

	dimensions = append(dimensions, dimensionPair{lowBound, highBound})

	// Parse additional dimensions (comma-separated)
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next low bound
		lowBound := p.parseArrayBound()
		if isInvalidExpression(lowBound) {
			p.addError("invalid array lower bound expression in multi-dimensional array", ErrInvalidExpression)
			return nil
		}

		if !p.peekTokenIs(lexer.DOTDOT) {
			p.addPeekTokenError("\"..\" expected", ErrUnexpectedToken)
			dimensions = append(dimensions, dimensionPair{
				low: lowBound,
				high: &ast.InvalidExpression{
					Reason: "missing upper array bound",
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{Token: p.cursor.Peek(1)},
					},
				},
			})
			return dimensions
		}
		p.nextToken()

		p.nextToken() // move to high bound
		highBound := p.parseArrayBound()
		if isInvalidExpression(highBound) {
			p.addError("invalid array upper bound expression in multi-dimensional array", ErrInvalidExpression)
			return nil
		}

		dimensions = append(dimensions, dimensionPair{lowBound, highBound})
	}

	return dimensions
}

// Syntax: class of ClassName
// Examples: class of TMyClass, class of TObject
func (p *Parser) parseClassOfType() *ast.ClassOfTypeNode {
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
	if isInvalidTypeExpression(classType) {
		p.addError("expected class type after 'class of'", ErrExpectedType)
		return nil
	}

	classOfNode := &ast.ClassOfTypeNode{
		Token:     classToken,
		ClassType: classType,
	}

	// EndPos is after the class type
	result, _ := builder.FinishWithNode(classOfNode, classType).(*ast.ClassOfTypeNode)
	return result
}
