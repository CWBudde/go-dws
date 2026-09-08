package evaluator

import (
	"sync"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestEvaluator_IndependentExecutionContexts(t *testing.T) {
	e := createTestEvaluator()
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func(value int64) {
			defer wg.Done()
			ctx := createTestContext()
			marker := &ast.IntegerLiteral{Value: -1}
			ctx.SetCurrentNode(marker)
			node := &ast.BinaryExpression{Operator: "+", Left: &ast.IntegerLiteral{Value: value}, Right: &ast.IntegerLiteral{Value: 1}}
			for iteration := 0; iteration < 100; iteration++ {
				result := e.Eval(node, ctx)
				integer, ok := result.(*runtime.IntegerValue)
				if !ok || integer.Value != value+1 {
					t.Errorf("unexpected result: %v", result)
					return
				}
				if ctx.CurrentNode() != marker {
					t.Error("evaluation did not restore its context's node")
					return
				}
			}
		}(int64(worker))
	}
	wg.Wait()
}

func TestBuiltinContext_KeepsExplicitExecutionContext(t *testing.T) {
	e := createTestEvaluator()
	outer, inner := createTestContext(), createTestContext()
	outerNode := &ast.IntegerLiteral{Value: 1}
	innerNode := &ast.IntegerLiteral{Value: 2}
	outer.SetCurrentNode(outerNode)
	inner.SetCurrentNode(innerNode)
	outerBuiltin := e.builtinContext(outer)
	innerBuiltin := e.builtinContext(inner)
	e.Eval(&ast.IntegerLiteral{Value: 3}, inner)
	if outerBuiltin.CurrentNode() != outerNode || innerBuiltin.CurrentNode() != innerNode {
		t.Fatal("builtin call lost its execution context after nested evaluation")
	}
	outerBuiltin.RaiseException("Exception", "outer failure", nil)
	if outer.Exception() == nil || inner.Exception() != nil {
		t.Fatal("builtin exception was raised in the wrong execution context")
	}
}
