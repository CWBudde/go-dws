package evaluator

import (
	"fmt"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestEvaluator_LexicalUnitRestoration(t *testing.T) {
	e, ctx := createTestEvaluator(), createTestContext()
	unitNode := &ast.Identifier{Value: "observe"}
	mainNode := &ast.Identifier{Value: "observe"}
	e.typeSystem.RegisterNodeUnit(unitNode, "UnitA")
	e.typeSystem.RegisterNodeUnit(mainNode, "")
	ctx.SetCurrentUnit("caller")
	observed := []string{}
	nested := false
	ctx.Env().Define("observe", runtime.NewReferenceValue("observe", func() (runtime.Value, error) {
		observed = append(observed, ctx.CurrentUnit())
		if !nested {
			nested = true
			value := e.Eval(mainNode, ctx)
			if !isError(value) {
				t.Error("expected inner error")
			}
			observed = append(observed, ctx.CurrentUnit())
		}
		return nil, fmt.Errorf("test failure")
	}, nil))
	if value := e.Eval(unitNode, ctx); !isError(value) {
		t.Fatal("expected outer error")
	}
	if fmt.Sprint(observed) != "[unita  unita]" {
		t.Fatalf("observed unit transitions: %#v", observed)
	}
	if ctx.CurrentUnit() != "caller" {
		t.Fatalf("caller unit was not restored: %q", ctx.CurrentUnit())
	}
}
