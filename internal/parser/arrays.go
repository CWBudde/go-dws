package parser

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// Array Indexing Syntax in DWScript
//
// DWScript supports two equivalent syntaxes for multi-dimensional array indexing:
//
// 1. Nested Bracket Syntax: arr[i][j][k]
//    - Each dimension uses separate brackets
//    - Parsed as nested IndexExpression nodes
//    - Traditional Pascal/Delphi style
//
// 2. Comma Syntax: arr[i, j, k]
//    - All indices within a single pair of brackets, separated by commas
//    - Desugared at parse time to nested IndexExpression nodes
//    - More concise, common in mathematical notation
//
// Both syntaxes produce identical AST structures and runtime behavior:
//   arr[i, j]   → desugared to → ((arr[i])[j])
//   arr[i, j, k] → desugared to → (((arr[i])[j])[k])
//
// This desugaring approach means no changes are needed in:
//   - AST node structures
//   - Semantic analyzer
//   - Interpreter
//   - Type checker
//
// The parser simply transforms comma syntax into nested bracket syntax during parsing.

// parseArrayDeclaration parses an array type declaration.
// Called after 'type Name =' has already been parsed.
// Current token should be 'array'.
//
// Syntax:
//   - type TMyArray = array[1..10] of Integer;  (static array with bounds)
//   - type TDynamic = array of String;          (dynamic array without bounds)
//
// Supports both single and multi-dimensional syntax:
//   - type TMatrix = array[0..1, 0..2] of Integer;    (2D)
//   - type TCube = array[1..3, 1..4, 1..5] of Float;  (3D)
//
// Multi-dimensional arrays are desugared into nested array types.
// POST: cursor is on SEMICOLON

// PRE: cursor is ARRAY
// POST: cursor is SEMICOLON

// Parses array/string indexing expressions using cursor navigation.
// Handles multi-dimensional indexing: arr[i, j, k] → nested IndexExpression
// PRE: cursor is on LBRACK
// POST: cursor is on RBRACK
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	builder := p.StartNode()
	lbrackToken := p.cursor.Current() // Save the '[' token for error reporting

	indexExpr := &ast.IndexExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: lbrackToken},
		},
		Left: left,
	}

	// Move to index expression
	p.cursor = p.cursor.Advance()

	// Parse the first index expression
	indexExpr.Index = p.parseExpression(LOWEST)

	// Handle comma-separated indices: arr[i, j, k]
	// Desugar to nested IndexExpression nodes: ((arr[i])[j])[k]
	result := indexExpr
	for {
		nextToken := p.cursor.Peek(1)
		if nextToken.Type != lexer.COMMA {
			break
		}

		p.cursor = p.cursor.Advance() // consume the comma
		p.cursor = p.cursor.Advance() // move to next index expression

		// Create a new IndexExpression with the previous result as the Left
		nextIndex := &ast.IndexExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: lbrackToken},
			},
			Left:  result,
			Index: p.parseExpression(LOWEST),
		}
		result = nextIndex
	}

	// Expect ']'
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.RBRACK {
		// Use structured error for missing closing bracket
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingRBracket).
			WithMessage("expected ']' to close array index").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpected(lexer.RBRACK).
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("add ']' to close the array index").
			WithRelatedPosition(lbrackToken.Pos, "opening '[' here").
			WithParsePhase("array index expression").
			Build()
		p.addStructuredError(err)
		return nil
	}

	// Advance to RBRACK
	p.cursor = p.cursor.Advance()

	return builder.FinishWithToken(result, p.cursor.Current()).(*ast.IndexExpression)
}

// Parses array literal expressions: [expr1, expr2, expr3]
// PRE: cursor is on LBRACK
// POST: cursor is on RBRACK
func (p *Parser) parseArrayLiteral() ast.Expression {
	lbrackToken := p.cursor.Current()

	// Handle empty literal: []
	nextToken := p.cursor.Peek(1)
	if nextToken.Type == lexer.RBRACK {
		p.cursor = p.cursor.Advance() // move to RBRACK
		return &ast.ArrayLiteralExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token:  lbrackToken,
					EndPos: p.cursor.Current().End(),
				},
			},
			Elements: []ast.Expression{},
		}
	}

	elements := []ast.Expression{}

	// Move to first element
	p.cursor = p.cursor.Advance()

	for {
		currentToken := p.cursor.Current()
		if currentToken.Type == lexer.RBRACK || currentToken.Type == lexer.EOF {
			break
		}

		// Parse element expression
		elementExpr := p.parseExpression(LOWEST)
		if elementExpr == nil {
			return nil
		}

		elem := elementExpr

		// Handle range syntax for set literals: [one..five]
		nextToken = p.cursor.Peek(1)
		if nextToken.Type == lexer.DOTDOT {
			p.cursor = p.cursor.Advance() // move to '..'
			rangeToken := p.cursor.Current()

			p.cursor = p.cursor.Advance() // move to end expression
			endExpr := p.parseExpression(LOWEST)
			if endExpr == nil {
				return nil
			}

			rangeExpr := &ast.RangeExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token:  rangeToken,
						EndPos: endExpr.End(),
					},
				},
				Start:    elementExpr,
				RangeEnd: endExpr,
			}
			elem = rangeExpr
		}

		elements = append(elements, elem)

		// Check for comma or closing bracket
		nextToken = p.cursor.Peek(1)
		if nextToken.Type == lexer.COMMA {
			p.cursor = p.cursor.Advance() // move to ','
			p.cursor = p.cursor.Advance() // advance to next element or ']'

			// Allow trailing comma: [1, 2, ]
			if p.cursor.Current().Type == lexer.RBRACK {
				break
			}
			continue
		}

		if nextToken.Type == lexer.RBRACK {
			p.cursor = p.cursor.Advance() // consume ']'
			break
		}

		// Unexpected token between elements
		p.addError(fmt.Sprintf("expected ',' or ']', got %s", nextToken.Type), ErrUnexpectedToken)
		return nil
	}

	// Determine if this should be treated as a set literal (all elements are identifiers or ranges)
	if shouldParseAsSetLiteral(elements) {
		setLit := &ast.SetLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token:  lbrackToken,
					EndPos: p.cursor.Current().End(),
				},
			},
			Elements: elements,
		}
		return setLit
	}

	return &ast.ArrayLiteralExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  lbrackToken,
				EndPos: p.cursor.Current().End(),
			},
		},
		Elements: elements,
	}
}

// parseArrayLiteral parses an array literal expression.
// Syntax:
//   - [expr1, expr2, expr3]
//   - []
//   - [[1, 2], [3, 4]]
//   - [x + 1, y * 2, z - 3]
//   - Supports optional trailing comma and range syntax for set literals ([one..five])
//
// PRE: cursor is LBRACK
// POST: cursor is RBRACK

// shouldParseAsSetLiteral determines if the parsed elements represent a set literal.
// We conservatively treat literals that contain only identifiers and range expressions as sets
// to preserve compatibility with existing set literal syntax until semantic context is available.
func shouldParseAsSetLiteral(elements []ast.Expression) bool {
	if len(elements) == 0 {
		return false
	}

	// Sets are primarily for enum values. The key heuristic:
	// - If ANY element is a RangeExpression (e.g., Red..Blue), treat as set
	// - If ALL elements are plain Identifiers (potential enum values), treat as set
	// - If elements contain literals (integers, strings, booleans) WITHOUT ranges, treat as array
	//
	// This matches DWScript behavior:
	//   [1, 2, 3]        -> array literal
	//   [Red, Blue]      -> set literal (enum identifiers)
	//   [1..10]          -> set literal (range)
	//   [Red..Blue, Green] -> set literal (mix of range and identifier)

	hasRange := false
	hasLiteral := false
	allIdentifiers := true

	for _, elem := range elements {
		switch elem.(type) {
		case *ast.RangeExpression:
			hasRange = true
			allIdentifiers = false
		case *ast.Identifier:
			// Keep allIdentifiers true
		case *ast.IntegerLiteral, *ast.CharLiteral, *ast.StringLiteral, *ast.BooleanLiteral:
			hasLiteral = true
			allIdentifiers = false
		default:
			// Complex expressions (binary ops, calls, etc.) indicate array literal
			return false
		}
	}

	// If there's a range expression, it's definitely a set
	if hasRange {
		return true
	}

	// If all elements are identifiers (no literals), treat as set (enum values)
	if allIdentifiers {
		return true
	}

	// If there are literals (and no ranges), treat as array
	if hasLiteral {
		return false
	}

	// Default to array for safety
	return false
}

// PRE: cursor is on ARRAY token
// POST: cursor is on SEMICOLON token
func (p *Parser) parseArrayDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.ArrayDecl {
	cursor := p.cursor

	arrayDecl := &ast.ArrayDecl{
		BaseNode: ast.BaseNode{Token: typeToken}, // The 'type' token
		Name:     nameIdent,
	}

	arrayToken := cursor.Current() // Save 'array' token

	// Collect all dimensions (comma-separated)
	type dimensionPair struct {
		low, high ast.Expression
	}
	var dimensions []dimensionPair

	if cursor.Peek(1).Type == lexer.LBRACK {
		cursor = cursor.Advance() // move to '['

		// Parse first dimension
		cursor = cursor.Advance() // move to start of expression
		p.cursor = cursor
		lowBound := p.parseExpression(LOWEST)
		cursor = p.cursor // Update cursor after parseExpression
		if lowBound == nil {
			p.addError("invalid array lower bound expression", ErrInvalidExpression)
			return nil
		}

		// Expect '..'
		if cursor.Peek(1).Type != lexer.DOTDOT {
			p.addError("expected '..' in array bounds", ErrUnexpectedToken)
			return nil
		}
		cursor = cursor.Advance() // move to '..'

		// Parse high bound expression
		cursor = cursor.Advance() // move to start of expression
		p.cursor = cursor
		highBound := p.parseExpression(LOWEST)
		cursor = p.cursor // Update cursor after parseExpression
		if highBound == nil {
			p.addError("invalid array upper bound expression", ErrInvalidExpression)
			return nil
		}

		dimensions = append(dimensions, dimensionPair{lowBound, highBound})

		// Parse additional dimensions (comma-separated)
		for cursor.Peek(1).Type == lexer.COMMA {
			cursor = cursor.Advance() // consume comma
			cursor = cursor.Advance() // move to next low bound
			p.cursor = cursor
			lowBound := p.parseExpression(LOWEST)
			cursor = p.cursor // Update cursor after parseExpression
			if lowBound == nil {
				p.addError("invalid array lower bound expression in multi-dimensional array", ErrInvalidExpression)
				return nil
			}

			if cursor.Peek(1).Type != lexer.DOTDOT {
				p.addError("expected '..' in array bounds", ErrUnexpectedToken)
				return nil
			}
			cursor = cursor.Advance() // move to '..'

			cursor = cursor.Advance() // move to high bound
			p.cursor = cursor
			highBound := p.parseExpression(LOWEST)
			cursor = p.cursor // Update cursor after parseExpression
			if highBound == nil {
				p.addError("invalid array upper bound expression in multi-dimensional array", ErrInvalidExpression)
				return nil
			}

			dimensions = append(dimensions, dimensionPair{lowBound, highBound})
		}

		// Expect ']'
		if cursor.Peek(1).Type != lexer.RBRACK {
			p.addError("expected ']' after array bounds", ErrUnexpectedToken)
			return nil
		}
		cursor = cursor.Advance() // move to ']'
	}

	// Expect 'of'
	if cursor.Peek(1).Type != lexer.OF {
		p.addError("expected 'of' after 'array'", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to 'of'

	// Parse element type (can be any type expression, including nested arrays)
	cursor = cursor.Advance() // move to element type
	p.cursor = cursor
	elementTypeExpr := p.parseTypeExpression()
	cursor = p.cursor // Update cursor after parseTypeExpression
	if elementTypeExpr == nil {
		p.addError("expected type expression after 'array of'", ErrExpectedType)
		return nil
	}

	// Convert TypeExpression to string representation for TypeAnnotation
	// This allows the semantic analyzer to resolve it via resolveInlineArrayType
	elementType := &ast.TypeAnnotation{
		Token: cursor.Current(),
		Name:  elementTypeExpr.String(),
	}

	// Build nested array type annotations if we have dimensions
	// This desugars: array[0..1, 0..2] of Integer
	//           into: array[0..1] of (array[0..2] of Integer)
	var arrayType *ast.ArrayTypeAnnotation
	if len(dimensions) == 0 {
		// Dynamic array without bounds
		arrayType = &ast.ArrayTypeAnnotation{
			Token:       arrayToken,
			ElementType: elementType,
			LowBound:    nil,
			HighBound:   nil,
		}
	} else {
		// Build from innermost to outermost
		// Start with the element type
		currentElementType := elementType

		// For each dimension (starting from the last), create an array type annotation
		for i := len(dimensions) - 1; i >= 0; i-- {
			// Create a new array type annotation with the current element type
			newArrayType := &ast.ArrayTypeAnnotation{
				Token:       arrayToken,
				ElementType: currentElementType,
				LowBound:    dimensions[i].low,
				HighBound:   dimensions[i].high,
			}

			// For the next iteration, wrap this array type as a TypeAnnotation
			if i > 0 {
				// Create a wrapper TypeAnnotation pointing to this array type
				currentElementType = &ast.TypeAnnotation{
					Token: arrayToken,
					Name:  newArrayType.String(),
				}
			} else {
				// This is the outermost dimension, use it directly
				arrayType = newArrayType
			}
		}
	}

	arrayDecl.ArrayType = arrayType

	// Expect semicolon
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after array declaration", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to ';'

	p.cursor = cursor
	return arrayDecl
}
