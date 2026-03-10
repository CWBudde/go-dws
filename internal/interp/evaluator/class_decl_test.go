package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)


type testClassDeclarationInfo struct {
	parent     any
	properties map[string]*types.PropertyInfo
	classVars  map[string]Value
	constants  map[string]Value
	fields     map[string]types.Type
	name       string
	isPartial  bool
}

func newTestClassDeclarationInfo(name string) *testClassDeclarationInfo {
	return &testClassDeclarationInfo{
		name:       name,
		properties: make(map[string]*types.PropertyInfo),
		classVars:  make(map[string]Value),
		constants:  make(map[string]Value),
		fields:     make(map[string]types.Type),
	}
}

func (c *testClassDeclarationInfo) IsPartialClass() bool { return c.isPartial }
func (c *testClassDeclarationInfo) SetPartialClass(isPartial bool) {
	c.isPartial = isPartial
}
func (c *testClassDeclarationInfo) SetAbstractClass(isAbstract bool)                      {}
func (c *testClassDeclarationInfo) SetExternalClass(isExternal bool, externalName string) {}
func (c *testClassDeclarationInfo) DefineCurrentClassMarker(env *runtime.Environment) {
	env.Define("__CurrentClass__", nil)
}
func (c *testClassDeclarationInfo) HasNoParentClass() bool { return c.parent == nil }
func (c *testClassDeclarationInfo) SetParentClass(parent any) {
	c.parent = parent
}
func (c *testClassDeclarationInfo) AddImplementedInterface(iface any, ifaceName string) {}
func (c *testClassDeclarationInfo) AddConstantValue(constDecl *ast.ConstDecl, value Value) {
	c.constants[constDecl.Name.Value] = value
}
func (c *testClassDeclarationInfo) ConstantValuesCopy() map[string]Value {
	out := make(map[string]Value, len(c.constants))
	for k, v := range c.constants {
		out[k] = v
	}
	return out
}
func (c *testClassDeclarationInfo) InheritConstantValuesFrom(parent any)                 {}
func (c *testClassDeclarationInfo) AddNestedClassRef(nestedName string, nestedClass any) {}
func (c *testClassDeclarationInfo) AddFieldDeclaration(fieldDecl *ast.FieldDecl, fieldType types.Type) {
	c.fields[fieldDecl.Name.Value] = fieldType
}
func (c *testClassDeclarationInfo) AddClassVarValue(name string, value Value) {
	c.classVars[name] = value
}
func (c *testClassDeclarationInfo) AddMethodDeclaration(method *ast.FunctionDecl, className string, registry *runtime.MethodRegistry) bool {
	return true
}
func (c *testClassDeclarationInfo) LookupDeclaredMethod(methodName string, isClassMethod bool) (*ast.FunctionDecl, bool) {
	return nil, false
}
func (c *testClassDeclarationInfo) SetConstructorDecl(constructor *ast.FunctionDecl) {}
func (c *testClassDeclarationInfo) SetDestructorDecl(destructor *ast.FunctionDecl)   {}
func (c *testClassDeclarationInfo) InheritDestructorMetadataIfMissing()              {}
func (c *testClassDeclarationInfo) SynthesizeImplicitDefaultConstructor()            {}
func (c *testClassDeclarationInfo) SetPropertyInfo(name string, propInfo *types.PropertyInfo) {
	c.properties[name] = propInfo
}
func (c *testClassDeclarationInfo) DeterminePropertyAccessKind(specName string) types.PropAccessKind {
	return types.PropAccessMethod
}
func (c *testClassDeclarationInfo) InheritParentPropertyInfos() {}
func (c *testClassDeclarationInfo) RegisterOperatorBinding(operatorSymbol, bindingName string, operandTypes []string) error {
	return nil
}
func (c *testClassDeclarationInfo) BuildVirtualMethodTableDirect() {}
func (c *testClassDeclarationInfo) RegisterInTypeSystem(ts any, parentName string) {
	type registrar interface {
		RegisterClassWithParent(name string, classInfo interptypes.ClassInfo, parentName string)
		RegisterClass(name string, classInfo interptypes.ClassInfo)
	}
	r := ts.(registrar)
	if parentName != "" {
		r.RegisterClassWithParent(c.name, c, parentName)
		return
	}
	r.RegisterClass(c.name, c)
}
func (c *testClassDeclarationInfo) DefineInEnv(env *runtime.Environment) {
	env.Define(c.name, nil)
}
func (c *testClassDeclarationInfo) RegisterMethodImplementation(fn *ast.FunctionDecl, allClasses map[string]interptypes.ClassInfo) {
}

func TestVisitClassDecl_UsesTypeSystemFactoryInsteadOfDeclHandler(t *testing.T) {
	typeSystem := interptypes.NewTypeSystem()
	typeSystem.ClassInfoFactory = func(className string) interptypes.ClassInfo {
		return newTestClassDeclarationInfo(className)
	}

	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)
	e.SetRuntimeBridge(nil)

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
