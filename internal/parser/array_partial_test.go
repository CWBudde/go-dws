package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestArrayLiteral_StopPreservesIfCondition(t *testing.T) {
	p := New(lexer.New("if item in [first..Test(1);"))
	program := p.ParseProgram()
	if len(program.Statements) != 1 {
		t.Fatalf("got %d statements, want recovered if", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok || stmt.Condition == nil {
		t.Fatalf("expected recovered if condition, got %#v", program.Statements[0])
	}
	if stmt.Consequence != nil {
		t.Fatalf("compiler stop must not invent a then branch: %#v", stmt.Consequence)
	}
	if len(p.Errors()) != 1 {
		t.Fatalf("expected only missing bracket, got %v", p.Errors())
	}
	foundRange := false
	ast.Inspect(stmt.Condition, func(node ast.Node) bool {
		if _, ok := node.(*ast.RangeExpression); ok {
			foundRange = true
		}
		return true
	})
	if !foundRange {
		t.Fatal("range was discarded before semantic analysis")
	}
}
