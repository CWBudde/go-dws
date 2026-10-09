package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestPropertyAccessorRecovery_GroupedNamedWriterSource(t *testing.T) {
	p := testParser("const K=2; type T = class property P: Integer write ((K)); end;")
	program := p.ParseProgram()
	checkParserErrors(t, p)
	prop := program.Statements[1].(*ast.ClassDecl).Properties[0]
	assignment, ok := prop.WriteStmt.(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("source writer = %T", prop.WriteStmt)
	}
	group, ok := prop.WriteSourceExpression.(*ast.GroupedExpression)
	if !ok || group.Expression != assignment.Target {
		t.Fatalf("successful inner source group/alias lost: source=%T target=%T", prop.WriteSourceExpression, assignment.Target)
	}
	count := 0
	ast.Inspect(prop, func(n ast.Node) bool {
		if n == assignment.Target {
			count++
		}
		return true
	})
	if count != 1 {
		t.Fatalf("provenance alias changed source traversal count: %d", count)
	}
}

func TestPropertyAccessorRecovery_RecordWriterSourceAlias(t *testing.T) {
	p := testParser("type R = record F: Integer; property P: Integer read F write ((F)); end;")
	program := p.ParseProgram()
	checkParserErrors(t, p)
	prop := program.Statements[0].(*ast.RecordDecl).Properties[0]
	assignment, ok := prop.WriteStmt.(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("writer = %T", prop.WriteStmt)
	}
	group, ok := prop.WriteSourceExpression.(*ast.GroupedExpression)
	if !ok || group.Expression != assignment.Target {
		t.Fatalf("record source alias lost: source=%T target=%T", prop.WriteSourceExpression, assignment.Target)
	}
	count := 0
	ast.Inspect(&prop, func(n ast.Node) bool {
		if n == assignment.Target {
			count++
		}
		return true
	})
	if count != 1 {
		t.Fatalf("record provenance alias traversal count=%d", count)
	}
}
