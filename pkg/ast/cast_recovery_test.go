package ast_test

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestAsExpression_RecoveryTraversalAndSpan(t *testing.T) {
	left := &ast.Identifier{BaseNode: ast.BaseNode{Token: token.Token{Pos: token.Position{Line: 1, Column: 1}}}, Value: "obj"}
	for _, right := range []ast.Expression{
		&ast.StringLiteral{BaseNode: ast.BaseNode{EndPos: token.Position{Line: 1, Column: 15}}, Value: "hello"},
		&ast.InvalidExpression{BaseNode: ast.BaseNode{EndPos: token.Position{Line: 1, Column: 9}}, Reason: "expression expected"},
	} {
		expr := &ast.AsExpression{Left: left, Right: right}
		seen := false
		ast.Inspect(expr, func(node ast.Node) bool {
			if node == right {
				seen = true
			}
			return true
		})
		if !seen {
			t.Fatalf("did not traverse %T", right)
		}
		if expr.Pos() != left.Pos() || expr.End() != right.End() {
			t.Fatalf("wrong span: %v .. %v", expr.Pos(), expr.End())
		}
		if expr.String() == "" {
			t.Fatal("empty rendering")
		}
	}
	expr := &ast.AsExpression{Left: left}
	if got := expr.String(); got != "(obj as )" {
		t.Fatalf("missing RHS rendering = %q", got)
	}
}
