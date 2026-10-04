package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// The intrinsic must retain its own span and following-token anchor rather
// than looking like a callable identifier or losing grouping in value contexts.
func TestParseDebugBreak_IntrinsicAndPositions(t *testing.T) {
	tests := []struct {
		name, source              string
		startCol, endLine, endCol int
		nextLine, nextCol         int
		grouped, incomplete       bool
	}{
		{"bare", "DebugBreak;", 1, 1, 11, 1, 11, false, false},
		{"mixed case", "dEbUgBrEaK;", 1, 1, 11, 1, 11, false, false},
		{"empty parentheses through comment", "DebugBreak {note}\n();", 1, 2, 3, 2, 3, false, false},
		{"grouped value", "(DebugBreak);", 2, 1, 12, 1, 12, true, false},
		{"interrupted EOF", "DebugBreak(", 1, 1, 12, 1, 11, false, true},
		{"argument is not parsed", "DebugBreak((1));", 1, 1, 12, 1, 12, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.source)
			program := p.ParseProgram()
			if tt.incomplete {
				if len(p.Errors()) != 1 {
					t.Fatalf("expected one punctuation stop, got %v", p.Errors())
				}
			} else {
				checkParserErrors(t, p)
			}
			if len(program.Statements) != 1 {
				t.Fatalf("expected one statement, got %d", len(program.Statements))
			}
			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("expected expression statement, got %T", program.Statements[0])
			}
			expr := stmt.Expression
			if tt.grouped {
				group, ok := expr.(*ast.GroupedExpression)
				if !ok {
					t.Fatalf("grouping must survive, got %T", expr)
				}
				expr = group.Expression
			}
			node, ok := expr.(*ast.DebugBreakExpression)
			if !ok {
				t.Fatalf("expected reserved intrinsic, got %T", expr)
			}
			span := [4]int{node.Pos().Line, node.Pos().Column, node.End().Line, node.End().Column}
			if span != [4]int{1, tt.startCol, tt.endLine, tt.endCol} {
				t.Fatalf("span = %v..%v, want 1:%d..%d:%d", node.Pos(), node.End(), tt.startCol, tt.endLine, tt.endCol)
			}
			if node.NextTokenPos.Line != tt.nextLine || node.NextTokenPos.Column != tt.nextCol || node.Incomplete != tt.incomplete {
				t.Fatalf("next token = %v, incomplete = %t; want %d:%d, %t", node.NextTokenPos, node.Incomplete, tt.nextLine, tt.nextCol, tt.incomplete)
			}
		})
	}
}

func TestParseDebugBreak_QualifiedMember(t *testing.T) {
	p := testParser("Obj.DebugBreak(7);")
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	call, ok := stmt.Expression.(*ast.MethodCallExpression)
	if !ok {
		t.Fatalf("qualified method must remain a call, got %T", stmt.Expression)
	}
	if call.Object.String() != "Obj" || call.Method.Value != "DebugBreak" || len(call.Arguments) != 1 {
		t.Fatalf("qualified member call lost its callee or argument: %s", call)
	}
}
