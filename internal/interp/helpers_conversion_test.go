package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestResolveTypeFromExpression_FunctionPointerNode(t *testing.T) {
	interp := New(&bytes.Buffer{})
	node := &ast.FunctionPointerTypeNode{
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "value"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
			},
		},
		ReturnType: &ast.TypeAnnotation{Name: "String"},
	}

	result := interp.resolveTypeFromExpression(node)
	funcPtr, ok := result.(*types.FunctionPointerType)
	if !ok {
		t.Fatalf("expected FunctionPointerType, got %T", result)
	}
	if len(funcPtr.Parameters) != 1 || !funcPtr.Parameters[0].Equals(types.INTEGER) {
		t.Fatalf("unexpected function pointer parameters: %#v", funcPtr.Parameters)
	}
	if funcPtr.ReturnType == nil || !funcPtr.ReturnType.Equals(types.STRING) {
		t.Fatalf("unexpected function pointer return type: %#v", funcPtr.ReturnType)
	}
}

func TestResolveTypeFromExpression_MethodPointerNode(t *testing.T) {
	interp := New(&bytes.Buffer{})
	node := &ast.FunctionPointerTypeNode{
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "value"},
				Type: &ast.TypeAnnotation{Name: "Boolean"},
			},
		},
		OfObject: true,
	}

	result := interp.resolveTypeFromExpression(node)
	methodPtr, ok := result.(*types.MethodPointerType)
	if !ok {
		t.Fatalf("expected MethodPointerType, got %T", result)
	}
	if len(methodPtr.Parameters) != 1 || !methodPtr.Parameters[0].Equals(types.BOOLEAN) {
		t.Fatalf("unexpected method pointer parameters: %#v", methodPtr.Parameters)
	}
	if methodPtr.ReturnType != nil {
		t.Fatalf("expected procedure method pointer return type nil, got %#v", methodPtr.ReturnType)
	}
}
