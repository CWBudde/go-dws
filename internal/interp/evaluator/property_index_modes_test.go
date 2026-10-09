package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestPropertyIndexModes_RuntimeMetadata(t *testing.T) {
	e := NewEvaluator(interptypes.NewTypeSystem(), nil, nil, nil, nil, runtime.NewRefCountManager())
	ctx := runtime.NewExecutionContext(runtime.NewEnvironment())
	prop := &ast.PropertyDecl{Name: &ast.Identifier{Value: "P"}, Type: &ast.TypeAnnotation{Name: "Integer"}, IndexParams: []*ast.Parameter{
		{Name: &ast.Identifier{Value: "A"}, ByRef: true},
		{Name: &ast.Identifier{Value: "B"}, ByRef: true},
		{Name: &ast.Identifier{Value: "C"}, IsConst: true},
		{Name: &ast.Identifier{Value: "D"}},
	}}
	// The same constructor registers class and interface properties, including unchecked runs.
	info := e.convertPropertyDecl(nil, prop, ctx)
	want := []types.PropertyIndexMode{types.PropertyIndexVar, types.PropertyIndexVar, types.PropertyIndexConst, types.PropertyIndexValue}
	if len(info.IndexParamModes) != len(want) {
		t.Fatalf("mode count = %d, want %d", len(info.IndexParamModes), len(want))
	}
	for i, mode := range want {
		if info.IndexMode(i) != mode {
			t.Errorf("mode %d = %d, want %d", i, info.IndexMode(i), mode)
		}
	}
}

func TestPropertyIndexModes_LegacyPreparation(t *testing.T) {
	e := NewEvaluator(interptypes.NewTypeSystem(), nil, nil, nil, nil, runtime.NewRefCountManager())
	ctx := runtime.NewExecutionContext(runtime.NewEnvironment())
	ctx.Env().Define("X", &runtime.IntegerValue{Value: 5})
	prop := &types.PropertyInfo{IndexParamNames: []string{"I"}}
	args, err := e.preparePropertyIndices(prop, []ast.Expression{&ast.Identifier{Value: "X"}}, nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 {
		t.Fatalf("args = %v", args)
	}
	if _, isRef := args[0].(ReferenceAccessor); isRef {
		t.Fatal("legacy metadata must pass a value")
	}
	if got := args[0].(*runtime.IntegerValue).Value; got != 5 {
		t.Fatalf("value = %d, want 5", got)
	}
}

func TestPropertyIndexModes_OriginalArgumentError(t *testing.T) {
	e := NewEvaluator(interptypes.NewTypeSystem(), nil, nil, nil, nil, runtime.NewRefCountManager())
	ctx := runtime.NewExecutionContext(runtime.NewEnvironment())
	original := &runtime.ErrorValue{Message: "original container error [line: 7, column: 3]"}
	ctx.Env().Define("Broken", original)
	prop := &types.PropertyInfo{IndexParamNames: []string{"I"}, IndexParamModes: []types.PropertyIndexMode{types.PropertyIndexVar}}
	index := &ast.IndexExpression{Left: &ast.Identifier{Value: "Broken"}, Index: &ast.IntegerLiteral{Value: 0}}
	_, err := e.preparePropertyIndices(prop, []ast.Expression{index}, index, ctx)
	if text := argumentEvaluationError(original, ctx).Error(); text != original.String() {
		t.Fatalf("existing argument error text changed: %q", text)
	}
	if err != original {
		t.Fatalf("argument error identity lost: got %v, want original %v", err, original)
	}
}

func TestPropertyIndexModes_RejectPartialSignature(t *testing.T) {
	e := NewEvaluator(interptypes.NewTypeSystem(), nil, nil, nil, nil, runtime.NewRefCountManager())
	ctx := runtime.NewExecutionContext(runtime.NewEnvironment())
	ctx.Env().Define("X", &runtime.IntegerValue{Value: 5})
	prop := &types.PropertyInfo{IndexParamNames: []string{"I", "J"}, IndexParamModes: []types.PropertyIndexMode{types.PropertyIndexVar}}
	args := []ast.Expression{&ast.Identifier{Value: "X"}, &ast.Identifier{Value: "X"}}
	if _, err := e.preparePropertyIndices(prop, args, nil, ctx); !isError(err) {
		t.Fatal("partial explicit mode signature must not silently default to value")
	}
}
