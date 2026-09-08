package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestVisitClassDecl_UsesRuntimeClassMetadata(t *testing.T) {
	typeSystem := interptypes.NewTypeSystem()

	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)

	ctx := runtime.NewExecutionContext(runtime.NewEnvironment())
	node := &ast.ClassDecl{
		Name: &ast.Identifier{Value: "TObject"},
	}

	result := e.VisitClassDecl(node, ctx)
	if isError(result) {
		t.Fatalf("VisitClassDecl returned error: %v", result)
	}

	if !typeSystem.HasClass("TObject") {
		t.Fatal("VisitClassDecl did not register class in TypeSystem")
	}
}
