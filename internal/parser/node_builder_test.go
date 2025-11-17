package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestNodeBuilderFinishStatement tests that Finish() correctly sets EndPos
// on statement nodes based on the current token.
func TestNodeBuilderFinishStatement(t *testing.T) {
	input := "x := 42;"
	p := New(lexer.New(input))

	// Start parsing - curToken is 'x'
	builder := p.StartNode()

	// Advance through the tokens
	p.nextToken() // now at :=
	p.nextToken() // now at 42
	p.nextToken() // now at ;

	// Create a simple statement
	stmt := &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{
			Token: builder.StartToken(),
		},
	}

	// Finish should set EndPos to semicolon's end position
	builder.Finish(stmt)

	// Verify EndPos was set correctly
	// Semicolon is at position (1, 8), and has length 1, so EndPos should be (1, 9)
	expectedEndPos := lexer.Position{Line: 1, Column: 9, Offset: 8}
	if stmt.EndPos != expectedEndPos {
		t.Errorf("EndPos = %+v, want %+v", stmt.EndPos, expectedEndPos)
	}
}

// TestNodeBuilderFinishExpression tests that Finish() correctly sets EndPos
// on expression nodes (which embed TypedExpressionBase).
func TestNodeBuilderFinishExpression(t *testing.T) {
	input := "3 + 5"
	p := New(lexer.New(input))

	// Start parsing - curToken is '3'
	builder := p.StartNode()

	// Advance through the tokens
	p.nextToken() // now at +
	p.nextToken() // now at 5

	// Create a binary expression
	expr := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: builder.StartToken(),
			},
		},
		Operator: "+",
	}

	// Finish should set EndPos to '5's end position
	builder.Finish(expr)

	// Verify EndPos was set correctly
	// '5' is at position (1, 5), has length 1, so EndPos should be (1, 6)
	expectedEndPos := lexer.Position{Line: 1, Column: 6, Offset: 5}
	if expr.EndPos != expectedEndPos {
		t.Errorf("EndPos = %+v, want %+v", expr.EndPos, expectedEndPos)
	}
}

// TestNodeBuilderFinishWithNode tests that FinishWithNode() sets EndPos
// based on a child node's end position.
func TestNodeBuilderFinishWithNode(t *testing.T) {
	input := "x + y"
	p := New(lexer.New(input))

	builder := p.StartNode()

	// Create a child expression that ends at 'y'
	p.nextToken() // now at +
	p.nextToken() // now at y

	childExpr := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: "y",
	}
	// Set child's EndPos manually for this test
	childExpr.EndPos = lexer.Position{Line: 1, Column: 6, Offset: 5}

	// Create parent statement
	stmt := &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{
			Token: builder.StartToken(),
		},
		Expression: childExpr,
	}

	// FinishWithNode should use child's EndPos
	builder.FinishWithNode(stmt, childExpr)

	// Verify EndPos matches child's EndPos
	if stmt.EndPos != childExpr.EndPos {
		t.Errorf("EndPos = %+v, want %+v (child's EndPos)", stmt.EndPos, childExpr.EndPos)
	}
}

// TestNodeBuilderFinishWithNodeNil tests that FinishWithNode() falls back
// to current token when child is nil.
func TestNodeBuilderFinishWithNodeNil(t *testing.T) {
	input := "return;"
	p := New(lexer.New(input))

	builder := p.StartNode()
	p.nextToken() // move to semicolon

	stmt := &ast.ReturnStatement{
		BaseNode: ast.BaseNode{
			Token: builder.StartToken(),
		},
		ReturnValue: nil, // No return value
	}

	// FinishWithNode with nil child should use current token
	builder.FinishWithNode(stmt, nil)

	// Should have EndPos set based on current token (semicolon)
	expectedEndPos := lexer.Position{Line: 1, Column: 8, Offset: 7}
	if stmt.EndPos != expectedEndPos {
		t.Errorf("EndPos = %+v, want %+v", stmt.EndPos, expectedEndPos)
	}
}

// TestNodeBuilderFinishWithToken tests that FinishWithToken() sets EndPos
// based on a specific token's position.
func TestNodeBuilderFinishWithToken(t *testing.T) {
	input := "begin end"
	p := New(lexer.New(input))

	builder := p.StartNode()
	p.nextToken() // move to 'end'
	endToken := p.curToken

	block := &ast.BlockStatement{
		BaseNode: ast.BaseNode{
			Token: builder.StartToken(),
		},
	}

	// FinishWithToken should use the endToken's position
	builder.FinishWithToken(block, endToken)

	// 'end' starts at column 7, has length 3, so EndPos should be (1, 10)
	expectedEndPos := lexer.Position{Line: 1, Column: 10, Offset: 9}
	if block.EndPos != expectedEndPos {
		t.Errorf("EndPos = %+v, want %+v", block.EndPos, expectedEndPos)
	}
}

// TestNodeBuilderStartToken tests that StartToken() returns the correct token.
func TestNodeBuilderStartToken(t *testing.T) {
	input := "x := 42"
	p := New(lexer.New(input))

	startToken := p.curToken
	builder := p.StartNode()

	// Advance parser
	p.nextToken()
	p.nextToken()

	// StartToken() should return the original token
	if builder.StartToken() != startToken {
		t.Errorf("StartToken() = %+v, want %+v", builder.StartToken(), startToken)
	}
}

// TestNodeBuilderMultipleNodes tests using multiple builders for nested nodes.
func TestNodeBuilderMultipleNodes(t *testing.T) {
	input := "if x then y;"
	p := New(lexer.New(input))

	// Outer builder for if statement
	outerBuilder := p.StartNode()
	p.nextToken() // move past 'if' to 'x'

	// Inner builder for condition
	innerBuilder := p.StartNode()
	condToken := p.curToken

	condition := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: condToken,
			},
		},
		Value: "x",
	}
	// Finish while still on 'x' - this sets EndPos correctly
	innerBuilder.Finish(condition)

	p.nextToken() // move past 'x' to 'then'

	// Continue parsing
	p.nextToken() // move past 'then'
	p.nextToken() // move to 'y'
	p.nextToken() // move to ';'

	stmt := &ast.IfStatement{
		BaseNode: ast.BaseNode{
			Token: outerBuilder.StartToken(),
		},
		Condition: condition,
	}

	// Finish outer node
	outerBuilder.Finish(stmt)

	// Verify both nodes have correct EndPos
	// Condition 'x' should end at column 5
	expectedCondEnd := lexer.Position{Line: 1, Column: 5, Offset: 4}
	if condition.EndPos != expectedCondEnd {
		t.Errorf("condition.EndPos = %+v, want %+v", condition.EndPos, expectedCondEnd)
	}

	// Statement should end at semicolon (column 13)
	expectedStmtEnd := lexer.Position{Line: 1, Column: 13, Offset: 12}
	if stmt.EndPos != expectedStmtEnd {
		t.Errorf("stmt.EndPos = %+v, want %+v", stmt.EndPos, expectedStmtEnd)
	}
}

// TestNodeBuilderNilNode tests that setEndPos handles nil nodes gracefully.
func TestNodeBuilderNilNode(t *testing.T) {
	input := "x"
	p := New(lexer.New(input))

	builder := p.StartNode()

	// Should not panic with nil node
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Finish(nil) panicked: %v", r)
		}
	}()

	builder.Finish(nil)
}

// TestNodeBuilderRealWorldIfStatement tests NodeBuilder with a real if statement parsing.
func TestNodeBuilderRealWorldIfStatement(t *testing.T) {
	input := "if x > 5 then WriteLn('big');"
	p := New(lexer.New(input))

	builder := p.StartNode()

	// Parse the if keyword
	if !p.curTokenIs(lexer.IF) {
		t.Fatal("expected IF token")
	}
	p.nextToken()

	// Parse condition (simplified - just consume tokens until 'then')
	condStart := p.curToken
	for p.curToken.Type != lexer.THEN {
		p.nextToken()
	}
	condition := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: condStart,
			},
		},
		Value: "x > 5",
	}

	// Parse 'then'
	p.nextToken() // move past 'then'

	// Parse consequence (simplified - consume until semicolon)
	for p.curToken.Type != lexer.SEMICOLON {
		p.nextToken()
	}

	stmt := &ast.IfStatement{
		BaseNode: ast.BaseNode{
			Token: builder.StartToken(),
		},
		Condition: condition,
	}

	// Finish at semicolon
	builder.Finish(stmt)

	// Verify positions
	if stmt.Pos().Line != 1 || stmt.Pos().Column != 1 {
		t.Errorf("stmt.Pos() = %+v, want Line:1 Column:1", stmt.Pos())
	}

	// Semicolon is at the end, should be around column 31
	if stmt.End().Line != 1 || stmt.End().Column <= 1 {
		t.Errorf("stmt.End() = %+v, expected Line:1 Column>1", stmt.End())
	}
}

// TestNodeBuilderChaining tests that builder methods can be chained.
func TestNodeBuilderChaining(t *testing.T) {
	input := "x := 42;"
	p := New(lexer.New(input))

	builder := p.StartNode()
	p.nextToken() // :=
	p.nextToken() // 42
	p.nextToken() // ;

	// Test chaining with type assertion
	stmt := builder.Finish(&ast.ExpressionStatement{
		BaseNode: ast.BaseNode{
			Token: builder.StartToken(),
		},
	}).(*ast.ExpressionStatement)

	// Verify it worked
	expectedEndPos := lexer.Position{Line: 1, Column: 9, Offset: 8}
	if stmt.EndPos != expectedEndPos {
		t.Errorf("EndPos = %+v, want %+v", stmt.EndPos, expectedEndPos)
	}
}

// TestNodeBuilderWithComplexExpression tests NodeBuilder with nested binary expressions.
func TestNodeBuilderWithComplexExpression(t *testing.T) {
	input := "x + y * z"
	p := New(lexer.New(input))

	outerBuilder := p.StartNode()

	// Consume all tokens
	p.nextToken() // +
	p.nextToken() // y
	p.nextToken() // *
	p.nextToken() // z

	expr := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: outerBuilder.StartToken(),
			},
		},
		Operator: "+",
	}

	outerBuilder.Finish(expr)

	// Expression should span from 'x' (column 1) to 'z' (column 10)
	if expr.Pos().Column != 1 {
		t.Errorf("expr.Pos().Column = %d, want 1", expr.Pos().Column)
	}
	expectedEndCol := 10
	if expr.End().Column != expectedEndCol {
		t.Errorf("expr.End().Column = %d, want %d", expr.End().Column, expectedEndCol)
	}
}
