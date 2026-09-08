package interp

import (
	"fmt"
	"io"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestClassVariableCompoundAssignment_PreservesRHSContext(t *testing.T) {
	tests := []struct {
		name string
		rhs  string
		want string
	}{
		{name: "parameter", rhs: "delta", want: "13\n"},
		{name: "function_call", rhs: "Twice(delta)", want: "16\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := fmt.Sprintf(`
function Twice(value: Integer): Integer;
begin
  Result := value * 2;
end;

type TCounter = class
  class var Count: Integer;

  class procedure Add(delta: Integer); static;
  begin
    Count += %s;
  end;
end;

TCounter.Count := 10;
TCounter.Add(3);
PrintLn(TCounter.Count);
`, tt.rhs)
			// Bare Count resolves through the current class, while the RHS
			// must retain the method's parameter scope and execution context.
			if got := runHelperScript(t, source); got != tt.want {
				t.Fatalf("output: got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInterpreterEvaluation_RestoresCurrentNode(t *testing.T) {
	tests := []struct {
		name string
		eval func(*Interpreter, ast.Node) Value
	}{
		{name: "Eval", eval: func(i *Interpreter, node ast.Node) Value {
			return i.Eval(node)
		}},
		{name: "EvalWithExpectedType", eval: func(i *Interpreter, node ast.Node) Value {
			return i.EvalWithExpectedType(node, types.INTEGER)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := New(io.Discard)
			marker := &ast.IntegerLiteral{Value: 1}
			i.ctx.SetCurrentNode(marker)
			result := tt.eval(i, &ast.IntegerLiteral{Value: 42})
			if result == nil || result.String() != "42" {
				t.Fatalf("literal evaluation returned %v", result)
			}
			if got := i.ctx.CurrentNode(); got != marker {
				t.Fatalf("current node was not restored: got %v, want marker %v", got, marker)
			}
		})
	}
}
