package runtime

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestCallStack_UnitOwnership(t *testing.T) {
	stack := NewCallStack(2)
	if err := stack.PushWithUnit("UnitProc", "", nil, "FirstUnit"); err != nil {
		t.Fatal(err)
	}
	if got := stack.CurrentUnit(); got != "FirstUnit" {
		t.Fatalf("unit = %q", got)
	}
	if err := stack.PushWithUnit("MainCallback", "", nil, ""); err != nil {
		t.Fatal(err)
	}
	if got := stack.CurrentUnit(); got != "" {
		t.Fatalf("main callback inherited unit %q", got)
	}
	if err := stack.PushWithUnit("Overflow", "", nil, "OtherUnit"); err == nil {
		t.Fatal("expected overflow")
	}
	clone := stack.Clone()
	clone.Pop()
	if got := clone.CurrentUnit(); got != "FirstUnit" {
		t.Fatalf("clone unit = %q", got)
	}
	if got := stack.CurrentUnit(); got != "" {
		t.Fatalf("clone changed original unit to %q", got)
	}
	stack.Pop()
	if got := stack.CurrentUnit(); got != "FirstUnit" {
		t.Fatalf("restored unit = %q", got)
	}
	stack.Clear()
	if got := stack.CurrentUnit(); got != "" {
		t.Fatalf("cleared unit = %q", got)
	}
	stack.Pop()
}

func TestExecutionContext_UnitOwnership(t *testing.T) {
	ctx := NewExecutionContext(NewEnvironment())
	ctx.SetCurrentUnit("DeclaringUnit")
	clone := ctx.Clone()
	if got := clone.CurrentUnit(); got != "DeclaringUnit" {
		t.Fatalf("clone unit = %q", got)
	}
	clone.SetCurrentUnit("")
	if got := ctx.CurrentUnit(); got != "DeclaringUnit" {
		t.Fatalf("clone changed original unit to %q", got)
	}
	ctx.Reset()
	if got := ctx.CurrentUnit(); got != "" {
		t.Fatalf("reset unit = %q", got)
	}
}

func TestMethodMetadata_BindingPreservesUnit(t *testing.T) {
	declaration := &ast.FunctionDecl{Name: &ast.Identifier{Value: "Execute"}}
	callable := MethodMetadataFromAST(declaration)
	callable.UnitName = "declaringunit"
	implementation := &ast.FunctionDecl{Name: &ast.Identifier{Value: "Execute"}, Body: &ast.BlockStatement{}}
	callable.BindImplementation(implementation)
	if callable.UnitName != "declaringunit" {
		t.Fatalf("binding lost unit: %q", callable.UnitName)
	}
}
