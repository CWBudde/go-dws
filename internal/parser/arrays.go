package parser

import (
	"strconv"

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
//    - Implemented in Task 9.172
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
//   - arr[i, j]   (multi-index comma syntax, desugared to arr[i][j])
//   - arr[i, j, k] (3D comma syntax, desugared to arr[i][j][k])
//
// Task 8.124: Parse array indexing expressions
// Task 9.172: Multi-index array syntax (arr[i, j])
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	lbrackToken := p.curToken // Save the '[' token for error reporting

	indexExpr := &ast.IndexExpression{
		Token: lbrackToken,
		Left:  left,
	}

	// Move to index expression
	p.nextToken()

	// Parse the first index expression
	indexExpr.Index = p.parseExpression(LOWEST)

	// Handle comma-separated indices: arr[i, j, k]
	// Desugar to nested IndexExpression nodes: ((arr[i])[j])[k]
	result := indexExpr
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume the comma
		p.nextToken() // move to next index expression

		// Create a new IndexExpression with the previous result as the Left
		nextIndex := &ast.IndexExpression{
			Token: lbrackToken, // Use the original '[' token
			Left:  result,
			Index: p.parseExpression(LOWEST),
		}
		result = nextIndex
	}

	// Expect ']'
	if !p.expectPeek(lexer.RBRACK) {
		return nil
	}

	return result
}

// parseArrayLiteral parses an array literal expression.
// Syntax:
//   - [expr1, expr2, expr3]
//   - []
//   - [[1, 2], [3, 4]]
//   - [x + 1, y * 2, z - 3]
//   - Supports optional trailing comma and range syntax for set literals ([one..five])
func (p *Parser) parseArrayLiteral() ast.Expression {
	lbrackToken := p.curToken

	// Handle empty literal: []
	if p.peekTokenIs(lexer.RBRACK) {
		p.nextToken() // consume ']'
		return &ast.ArrayLiteralExpression{
			Token:    lbrackToken,
			Elements: []ast.Expression{},
		}
	}

	elements := []ast.Expression{}

	// Move to first element
	p.nextToken()

	for !p.curTokenIs(lexer.RBRACK) && !p.curTokenIs(lexer.EOF) {
		elementExpr := p.parseExpression(LOWEST)
		if elementExpr == nil {
			return nil
		}

		elem := elementExpr

		// Handle range syntax for set literals: [one..five]
		if p.peekTokenIs(lexer.DOTDOT) {
			p.nextToken() // move to '..'
			rangeToken := p.curToken

			p.nextToken() // move to end expression
			endExpr := p.parseExpression(LOWEST)
			if endExpr == nil {
				return nil
			}

			elem = &ast.RangeExpression{
				Token: rangeToken,
				Start: elementExpr,
				End:   endExpr,
			}
		}

		elements = append(elements, elem)

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // move to ','
			p.nextToken() // advance to next element or ']'

			// Allow trailing comma: [1, 2, ]
			if p.curTokenIs(lexer.RBRACK) {
				break
			}
			continue
		}

		if p.peekTokenIs(lexer.RBRACK) {
			p.nextToken() // consume ']'
			break
		}

		// Unexpected token between elements
		p.addError("expected ',' or ']' in array literal")

		// Advance to the unexpected token to avoid infinite loops
		p.nextToken()

		// Attempt to recover by skipping tokens until we reach a closing bracket,
		// statement terminator, or EOF. This helps downstream parsing continue.
		for !p.curTokenIs(lexer.RBRACK) &&
			!p.curTokenIs(lexer.SEMICOLON) &&
			!p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}

		if !p.curTokenIs(lexer.RBRACK) {
			p.addError("expected closing ']' for array literal")
		}

		return nil
	}

	if !p.curTokenIs(lexer.RBRACK) {
		// Missing closing bracket
		p.addError("expected closing ']' for array literal")
		return nil
	}

	// Determine if this should be treated as a set literal (all elements are identifiers or ranges)
	if shouldParseAsSetLiteral(elements) {
		return &ast.SetLiteral{
			Token:    lbrackToken,
			Elements: elements,
		}
	}

	return &ast.ArrayLiteralExpression{
		Token:    lbrackToken,
		Elements: elements,
	}
}

// shouldParseAsSetLiteral determines if the parsed elements represent a set literal.
// We conservatively treat literals that contain only identifiers and range expressions as sets
// to preserve compatibility with existing set literal syntax until semantic context is available.
func shouldParseAsSetLiteral(elements []ast.Expression) bool {
	if len(elements) == 0 {
		return false
	}

	for _, elem := range elements {
		switch elem.(type) {
		case *ast.Identifier:
			// Identifiers remain valid set elements
		case *ast.RangeExpression:
			// Ranges are valid in set literals
		default:
			return false
		}
	}

	return true
}
