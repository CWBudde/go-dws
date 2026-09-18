package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestUnitQualified_ParameterlessPointerContext(t *testing.T) {
	e, ctx := createTestEvaluator(), createTestContext()
	registry := units.NewUnitRegistry(nil)
	unit := units.NewUnit("UnitA", "unita.dws")
	registry.RegisterUnit(unit.Name, unit)
	e.SetUnitRegistry(registry)
	fn := &ast.FunctionDecl{Name: &ast.Identifier{Value: "Execute"}, Body: &ast.BlockStatement{}}
	e.typeSystem.RegisterFunctionWithUnit(unit.Name, fn.Name.Value, fn)
	node := &ast.MemberAccessExpression{Object: &ast.Identifier{Value: "UnitA"}, Member: &ast.Identifier{Value: "Execute"}}
	info := ast.NewSemanticInfo()
	info.SetType(node, &ast.TypeAnnotation{Name: "Procedure"})
	info.SetResolvedType(node, types.NewFunctionPointerType(nil, nil))
	e.SetSemanticInfo(info)
	result := e.Eval(node, ctx)
	pointer, ok := result.(*runtime.FunctionPointerValue)
	if !ok || pointer.Function != fn {
		t.Fatalf("expected pointer to qualified procedure, got %#v", result)
	}
}
