package runtime

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestClassCallable_ImplementationPreservesIdentity(t *testing.T) {
	class := NewClassInfo("TBase")
	decl := &ast.FunctionDecl{Name: &ast.Identifier{Value: "Run"}, ClassName: &ast.Identifier{Value: "TBase"}, IsVirtual: true}
	class.AddMethodDeclaration(decl, "TBase", NewMethodRegistry())
	callable := class.Metadata.Methods["run"]
	pointer := &FunctionPointerValue{Callable: callable}
	impl := &ast.FunctionDecl{Name: &ast.Identifier{Value: "Run"}, ClassName: &ast.Identifier{Value: "TBase"}, Body: &ast.BlockStatement{}}
	class.RegisterMethodImplementation(impl, map[string]IClassInfo{"tbase": class})
	if class.Metadata.Methods["run"] != callable {
		t.Fatal("implementation replaced callable identity")
	}
	if callable.Body != impl.Body {
		t.Fatal("implementation body not bound to canonical callable")
	}
	if pointer.IsNil() || pointer.GetFunctionDecl().(*ast.FunctionDecl).Body != impl.Body || callable.SourceDeclaration != impl {
		t.Fatal("captured method pointer retained obsolete implementation")
	}
}

func TestClassCallable_ImplementationKeepsDeclarationState(t *testing.T) {
	class := NewClassInfo("TExample")
	defaultValue := &ast.IntegerLiteral{Value: 7}
	declaration := &ast.FunctionDecl{Name: &ast.Identifier{Value: "Run"}, IsClassMethod: true, IsStatic: true, IsVirtual: true, Parameters: []*ast.Parameter{{Name: &ast.Identifier{Value: "count"}, Type: &ast.TypeAnnotation{Name: "Integer"}, DefaultValue: defaultValue, IsLazy: true, IsConst: true}}}
	declaration.PreConditions = &ast.PreConditions{}
	declaration.PostConditions = &ast.PostConditions{}
	class.AddMethodDeclaration(declaration, class.Name, NewMethodRegistry())
	callable := class.LookupClassMethod("run")
	originalID := callable.ID
	implementation := &ast.FunctionDecl{Name: &ast.Identifier{Value: "Run"}, IsClassMethod: true, Parameters: []*ast.Parameter{{Name: &ast.Identifier{Value: "count"}, Type: &ast.TypeAnnotation{Name: "Integer"}}}, Body: &ast.BlockStatement{}}
	class.RegisterMethodImplementation(implementation, map[string]IClassInfo{"texample": class})
	if callable.Owner != class || callable.ID != originalID || !callable.IsVirtual || !callable.IsStatic {
		t.Fatal("implementation discarded declaration ownership or flags")
	}
	if callable.PreConditions != declaration.PreConditions || callable.PostConditions != declaration.PostConditions || !callable.Declaration.Parameters[0].IsLazy || !callable.Declaration.Parameters[0].IsConst {
		t.Fatal("implementation discarded declared contracts or parameter modifiers")
	}
	if callable.Parameters[0].DefaultValue != defaultValue {
		t.Fatal("implementation discarded declared default")
	}
	if implementation.Parameters[0].DefaultValue != nil || implementation.IsVirtual || implementation.IsStatic {
		t.Fatal("binding modified compiled source declaration")
	}
}

func TestClassCallable_InheritedBindingsShareCallable(t *testing.T) {
	parent := NewClassInfo("TBase")
	declaration := &ast.FunctionDecl{Name: &ast.Identifier{Value: "Run"}, IsVirtual: true}
	parent.AddMethodDeclaration(declaration, parent.Name, NewMethodRegistry())
	parent.BuildVirtualMethodTableDirect()
	child := NewClassInfo("TChild")
	child.SetParentClass(parent)
	child.BuildVirtualMethodTableDirect()
	callable := parent.LookupMethod("run")
	if child.LookupMethod("RUN") != callable || len(child.Metadata.Methods) != 0 {
		t.Fatal("inherited lookup copied the declaring class method table")
	}
	implementation := &ast.FunctionDecl{Name: &ast.Identifier{Value: "Run"}, Body: &ast.BlockStatement{}}
	parent.RegisterMethodImplementation(implementation, map[string]IClassInfo{"tbase": parent, "tchild": child})
	if child.LookupMethod("run") != callable || child.VirtualMethods["run_0"].Method != callable {
		t.Fatal("inherited binding lost canonical callable identity")
	}
	if callable.Body != implementation.Body {
		t.Fatal("inherited callable retained old implementation")
	}
}
