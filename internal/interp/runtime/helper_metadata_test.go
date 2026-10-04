package runtime

import (
	"slices"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestMutableHelperInfo_InheritedOverloads(t *testing.T) {
	parent := NewMutableHelperInfo("Parent", types.INTEGER, false)
	first, second, own := &ast.FunctionDecl{}, &ast.FunctionDecl{}, &ast.FunctionDecl{}
	parent.MethodOverloads["Pick"] = []*ast.FunctionDecl{first, second}
	parent.Methods["Pick"] = second
	child := NewMutableHelperInfo("Child", types.INTEGER, false)
	child.ParentHelper = parent
	grandchild := NewMutableHelperInfo("Grandchild", types.INTEGER, false)
	grandchild.ParentHelper = child
	for _, tt := range []struct {
		helper *MutableHelperInfo
		name   string
	}{
		{parent, "parent"}, {child, "child"}, {grandchild, "grandchild"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertHelperOverloads(t, tt.helper, parent, []*ast.FunctionDecl{first, second})
		})
	}
	child.Methods["pick"] = own
	assertHelperOverloads(t, grandchild, child, []*ast.FunctionDecl{own})
	child.MethodOverloads["pick"] = []*ast.FunctionDecl{own, first}
	assertHelperOverloads(t, grandchild, child, []*ast.FunctionDecl{own, first})
	if methods, owner, found := grandchild.GetMethodOverloads("Missing"); found || owner != nil || methods != nil {
		t.Fatal("missing name unexpectedly resolved")
	}
}

func assertHelperOverloads(t *testing.T, helper, wantOwner *MutableHelperInfo, wantMethods []*ast.FunctionDecl) {
	t.Helper()
	methods, owner, found := helper.GetMethodOverloads("pIcK")
	if !found || owner != wantOwner || !slices.Equal(methods, wantMethods) {
		t.Fatalf("lookup = (%d declarations, %p, %v), want %d original declarations owned by %p", len(methods), owner, found, len(wantMethods), wantOwner)
	}
}
