package interp

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// InterfaceInfo Tests
// ============================================================================

func TestInterfaceInfoCreation(t *testing.T) {
	// Create a simple interface: IMyInterface
	interfaceInfo := NewInterfaceInfo("IMyInterface")

	if interfaceInfo.Name != "IMyInterface" {
		t.Errorf("interfaceInfo.Name = %s, want IMyInterface", interfaceInfo.Name)
	}

	if interfaceInfo.Parent != nil {
		t.Errorf("interfaceInfo.Parent should be nil for root interface")
	}

	if interfaceInfo.Methods == nil {
		t.Error("interfaceInfo.Methods should be initialized")
	}
}

func TestInterfaceInfoWithInheritance(t *testing.T) {
	// Create parent interface
	parent := NewInterfaceInfo("IBase")
	parentMethod := &ast.FunctionDecl{
		Token: lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "BaseMethod"},
			Value: "BaseMethod",
		},
	}
	parent.Methods["BaseMethod"] = parentMethod

	// Create child interface
	child := NewInterfaceInfo("IDerived")
	child.Parent = parent
	childMethod := &ast.FunctionDecl{
		Token: lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "DerivedMethod"},
			Value: "DerivedMethod",
		},
	}
	child.Methods["DerivedMethod"] = childMethod

	if child.Parent == nil {
		t.Fatal("child.Parent should not be nil")
	}

	if child.Parent.Name != "IBase" {
		t.Errorf("child.Parent.Name = %s, want IBase", child.Parent.Name)
	}
}

func TestInterfaceInfoAddMethod(t *testing.T) {
	interfaceInfo := NewInterfaceInfo("ICounter")

	// Create method AST nodes
	incrementMethod := &ast.FunctionDecl{
		Token: lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Increment"},
			Value: "Increment",
		},
	}
	interfaceInfo.Methods["Increment"] = incrementMethod

	getValueMethod := &ast.FunctionDecl{
		Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "GetValue"},
			Value: "GetValue",
		},
		ReturnType: &ast.TypeAnnotation{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
			Name:  "Integer",
		},
	}
	interfaceInfo.Methods["GetValue"] = getValueMethod

	if len(interfaceInfo.Methods) != 2 {
		t.Errorf("len(interfaceInfo.Methods) = %d, want 2", len(interfaceInfo.Methods))
	}

	if interfaceInfo.Methods["Increment"] == nil {
		t.Error("Method 'Increment' should exist")
	}

	if interfaceInfo.Methods["GetValue"] == nil {
		t.Error("Method 'GetValue' should exist")
	}
}

// ============================================================================
// InterfaceInstance Tests
// ============================================================================

func TestInterfaceInstanceCreation(t *testing.T) {
	// Create an interface
	interfaceInfo := NewInterfaceInfo("IMyInterface")

	// Create a class that implements the interface
	classInfo := NewClassInfo("TMyClass")

	// Create an object instance
	obj := NewObjectInstance(classInfo)

	// Create an interface instance wrapping the object
	iface := NewInterfaceInstance(interfaceInfo, obj)

	if iface.Interface != interfaceInfo {
		t.Error("iface.Interface should point to interfaceInfo")
	}

	if iface.Object != obj {
		t.Error("iface.Object should point to obj")
	}
}

func TestInterfaceInstanceImplementsValue(t *testing.T) {
	// Create an interface
	interfaceInfo := NewInterfaceInfo("IMyInterface")

	// Create a class and object
	classInfo := NewClassInfo("TMyClass")
	obj := NewObjectInstance(classInfo)

	// Create an interface instance
	iface := NewInterfaceInstance(interfaceInfo, obj)

	// Test that it implements Value interface
	var _ Value = iface

	// Test Type() method
	if iface.Type() != "INTERFACE" {
		t.Errorf("iface.Type() = %s, want INTERFACE", iface.Type())
	}

	// Test String() method
	expected := "IMyInterface instance (wrapping TMyClass)"
	if iface.String() != expected {
		t.Errorf("iface.String() = %s, want %s", iface.String(), expected)
	}
}

func TestInterfaceInstanceGetUnderlyingObject(t *testing.T) {
	// Create an interface
	interfaceInfo := NewInterfaceInfo("IMyInterface")

	// Create a class and object
	classInfo := NewClassInfo("TMyClass")
	obj := NewObjectInstance(classInfo)
	obj.Fields["X"] = &IntegerValue{Value: 42}

	// Create an interface instance
	iface := NewInterfaceInstance(interfaceInfo, obj)

	// Get the underlying object
	underlyingObj := iface.Object

	if underlyingObj != obj {
		t.Error("GetUnderlyingObject should return the wrapped object")
	}

	// Verify we can access the object's fields
	if underlyingObj.Fields["X"].(*IntegerValue).Value != 42 {
		t.Error("Should be able to access underlying object's fields")
	}
}

// ============================================================================
// evalInterfaceDeclaration Tests
// ============================================================================

func TestEvalInterfaceDeclaration(t *testing.T) {
	// Create interpreter
	interp := testInterpreter()

	// Create a simple interface AST
	interfaceDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "IMyInterface"},
			Value: "IMyInterface",
		},
		Methods: []*ast.InterfaceMethodDecl{
			{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
				},
				Name: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "DoSomething"},
					Value: "DoSomething",
				},
			},
		},
	}

	// Evaluate the interface declaration
	result := interp.evalInterfaceDeclaration(interfaceDecl)

	// Should return NilValue (interface declarations don't produce values)
	if result == nil || result.Type() != "NIL" {
		if result == nil {
			t.Error("evalInterfaceDeclaration should return NilValue, got nil")
		} else {
			t.Errorf("evalInterfaceDeclaration should return NIL, got %s", result.Type())
		}
	}

	// Check that interface was registered
	ifaceInfo, exists := interp.interfaces["imyinterface"]
	if !exists {
		t.Fatal("Interface 'IMyInterface' should be registered")
	}

	if ifaceInfo.Name != "IMyInterface" {
		t.Errorf("Interface name = %s, want IMyInterface", ifaceInfo.Name)
	}

	if len(ifaceInfo.Methods) != 1 {
		t.Errorf("Interface should have 1 method, got %d", len(ifaceInfo.Methods))
	}

	if !ifaceInfo.HasMethod("DoSomething") {
		t.Error("Interface should have method 'DoSomething'")
	}
}

func TestEvalInterfaceDeclarationWithInheritance(t *testing.T) {
	// Create interpreter
	interp := testInterpreter()

	// Create parent interface
	parentDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "IBase"},
			Value: "IBase",
		},
		Methods: []*ast.InterfaceMethodDecl{
			{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
				},
				Name: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "BaseMethod"},
					Value: "BaseMethod",
				},
			},
		},
	}
	interp.evalInterfaceDeclaration(parentDecl)

	// Create child interface
	childDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "IDerived"},
			Value: "IDerived",
		},
		Parent: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "IBase"},
			Value: "IBase",
		},
		Methods: []*ast.InterfaceMethodDecl{
			{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
				},
				Name: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "DerivedMethod"},
					Value: "DerivedMethod",
				},
			},
		},
	}
	interp.evalInterfaceDeclaration(childDecl)

	// Check that child interface was registered with parent
	childInfo, exists := interp.interfaces["iderived"]
	if !exists {
		t.Fatal("Interface 'IDerived' should be registered")
	}

	if childInfo.Parent == nil {
		t.Fatal("Child interface should have parent")
	}

	if childInfo.Parent.Name != "IBase" {
		t.Errorf("Parent name = %s, want IBase", childInfo.Parent.Name)
	}

	// Check that child has its own method
	if !childInfo.HasMethod("DerivedMethod") {
		t.Error("Child should have method 'DerivedMethod'")
	}

	// Check that child has inherited method
	if !childInfo.HasMethod("BaseMethod") {
		t.Error("Child should inherit method 'BaseMethod'")
	}

	// Check that child has both methods via AllMethods
	allMethods := childInfo.AllMethods()
	if len(allMethods) != 2 {
		t.Errorf("Child should have 2 methods total, got %d", len(allMethods))
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

// TestCompleteInterfaceWorkflow tests the complete interface workflow:
// - Interface declaration
// - Class implementing the interface
// - Object-to-interface casting
// - Method calls through interface
// - Interface variable assignment
func TestCompleteInterfaceWorkflow(t *testing.T) {
	interp := testInterpreter()

	// Step 1: Declare an interface
	interfaceDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "ICounter"},
			Value: "ICounter",
		},
		Methods: []*ast.InterfaceMethodDecl{
			{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "GetValue"},
					Value: "GetValue",
				},
				ReturnType: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
			},
		},
	}
	interp.evalInterfaceDeclaration(interfaceDecl)

	// Step 2: Declare a class that implements the interface
	classDecl := &ast.ClassDecl{
		Token: lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "TCounter"},
			Value: "TCounter",
		},
		Fields: []*ast.FieldDecl{
			{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "FValue"},
				Name: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "FValue"},
					Value: "FValue",
				},
				Type: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
			},
		},
		Methods: []*ast.FunctionDecl{
			{
				Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
				Name: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "GetValue"},
					Value: "GetValue",
				},
				ReturnType: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
				Body: &ast.BlockStatement{}, // Simplified - would have actual implementation
			},
		},
	}
	interp.evalClassDeclaration(classDecl)

	// Step 3: Create an object instance
	classInfo := interp.classes["TCounter"]
	obj := NewObjectInstance(classInfo)
	obj.SetField("FValue", &IntegerValue{Value: 42})

	// Step 4: Cast object to interface
	ifaceInfo := interp.interfaces["icounter"]

	// Verify the class implements the interface
	if !classImplementsInterface(classInfo, ifaceInfo) {
		t.Fatal("TCounter should implement ICounter")
	}

	// Create interface instance wrapping the object
	ifaceInstance := NewInterfaceInstance(ifaceInfo, obj)

	// Verify the interface instance was created correctly
	if ifaceInstance.Interface != ifaceInfo {
		t.Error("Interface instance should reference ICounter")
	}

	if ifaceInstance.Object != obj {
		t.Error("Interface instance should wrap the object")
	}

	// Step 5: Verify we can still access the underlying object
	underlying := ifaceInstance.GetUnderlyingObject()
	if underlying != obj {
		t.Error("Should be able to get underlying object from interface")
	}

	// Step 6: Test interface variable assignment (reference semantics)
	var interfaceVar Value = ifaceInstance
	if interfaceVar.Type() != "INTERFACE" {
		t.Errorf("Interface variable should have type INTERFACE, got %s", interfaceVar.Type())
	}
}

// ============================================================================
// Runtime Testing
// ============================================================================

// TestInterfaceVariable tests interface variable creation and assignment
func TestInterfaceVariable(t *testing.T) {
	interp := testInterpreter()

	// Create interface
	interfaceDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{Value: "IPrintable"},
		Methods: []*ast.InterfaceMethodDecl{
			{
				Name: &ast.Identifier{Value: "Print"},
			},
		},
	}
	interp.evalInterfaceDeclaration(interfaceDecl)

	// Create class implementing interface
	classDecl := &ast.ClassDecl{
		Token: lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:  &ast.Identifier{Value: "TDocument"},
		Methods: []*ast.FunctionDecl{
			{
				Name: &ast.Identifier{Value: "Print"},
				Body: &ast.BlockStatement{},
			},
		},
	}
	interp.evalClassDeclaration(classDecl)

	// Create object and cast to interface
	obj := NewObjectInstance(interp.classes["TDocument"])
	ifaceInstance := NewInterfaceInstance(interp.interfaces["iprintable"], obj)

	// Store in environment as a variable
	interp.env.Define("myInterface", ifaceInstance)

	// Retrieve and verify
	val, exists := interp.env.Get("myInterface")
	if !exists {
		t.Fatal("Interface variable should exist in environment")
	}

	if val.Type() != "INTERFACE" {
		t.Errorf("Variable should have type INTERFACE, got %s", val.Type())
	}

	// Verify it's the same instance
	retrieved, ok := val.(*InterfaceInstance)
	if !ok {
		t.Fatal("Variable should be an InterfaceInstance")
	}

	if retrieved.Interface.Name != "IPrintable" {
		t.Errorf("Interface name should be IPrintable, got %s", retrieved.Interface.Name)
	}
}

// TestObjectToInterface tests object-to-interface casting
func TestObjectToInterface(t *testing.T) {
	interp := testInterpreter()

	// Create interface with multiple methods
	interfaceDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{Value: "IDrawable"},
		Methods: []*ast.InterfaceMethodDecl{
			{Name: &ast.Identifier{Value: "Draw"}},
			{Name: &ast.Identifier{Value: "GetWidth"}},
			{Name: &ast.Identifier{Value: "GetHeight"}},
		},
	}
	interp.evalInterfaceDeclaration(interfaceDecl)

	// Create class implementing all methods
	classDecl := &ast.ClassDecl{
		Token: lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:  &ast.Identifier{Value: "TRectangle"},
		Methods: []*ast.FunctionDecl{
			{Name: &ast.Identifier{Value: "Draw"}, Body: &ast.BlockStatement{}},
			{Name: &ast.Identifier{Value: "GetWidth"}, Body: &ast.BlockStatement{}},
			{Name: &ast.Identifier{Value: "GetHeight"}, Body: &ast.BlockStatement{}},
		},
	}
	interp.evalClassDeclaration(classDecl)

	// Test successful cast
	obj := NewObjectInstance(interp.classes["TRectangle"])
	ifaceInfo := interp.interfaces["idrawable"]

	if !classImplementsInterface(obj.Class, ifaceInfo) {
		t.Fatal("TRectangle should implement IDrawable")
	}

	ifaceInstance := NewInterfaceInstance(ifaceInfo, obj)
	if ifaceInstance.Object != obj {
		t.Error("Interface should wrap the original object")
	}

	// Test failed cast (class missing methods)
	incompatibleClass := &ast.ClassDecl{
		Token: lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:  &ast.Identifier{Value: "TPoint"},
		Methods: []*ast.FunctionDecl{
			{Name: &ast.Identifier{Value: "Draw"}, Body: &ast.BlockStatement{}},
			// Missing GetWidth and GetHeight
		},
	}
	interp.evalClassDeclaration(incompatibleClass)

	objIncompatible := NewObjectInstance(interp.classes["TPoint"])
	if classImplementsInterface(objIncompatible.Class, ifaceInfo) {
		t.Error("TPoint should NOT implement IDrawable (missing methods)")
	}
}

// TestInterfaceMethodCall tests method calls through interface references
func TestInterfaceMethodCall(t *testing.T) {
	interp := testInterpreter()

	// Create interface
	interfaceDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{Value: "ICalculator"},
		Methods: []*ast.InterfaceMethodDecl{
			{
				Name: &ast.Identifier{Value: "Add"},
				Parameters: []*ast.Parameter{
					{Name: &ast.Identifier{Value: "x"}},
					{Name: &ast.Identifier{Value: "y"}},
				},
				ReturnType: &ast.TypeAnnotation{Name: "Integer"},
			},
		},
	}
	interp.evalInterfaceDeclaration(interfaceDecl)

	// Create implementing class
	classDecl := &ast.ClassDecl{
		Token: lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:  &ast.Identifier{Value: "TCalculator"},
		Methods: []*ast.FunctionDecl{
			{
				Name: &ast.Identifier{Value: "Add"},
				Parameters: []*ast.Parameter{
					{Name: &ast.Identifier{Value: "x"}},
					{Name: &ast.Identifier{Value: "y"}},
				},
				ReturnType: &ast.TypeAnnotation{Name: "Integer"},
				Body:       &ast.BlockStatement{},
			},
		},
	}
	interp.evalClassDeclaration(classDecl)

	// Create interface instance
	obj := NewObjectInstance(interp.classes["TCalculator"])
	ifaceInstance := NewInterfaceInstance(interp.interfaces["icalculator"], obj)

	// Verify method can be found through interface
	method := ifaceInstance.Interface.GetMethod("Add")
	if method == nil {
		t.Fatal("Should be able to find Add method through interface")
	}

	if method.Name.Value != "Add" {
		t.Errorf("Method name should be Add, got %s", method.Name.Value)
	}

	// Verify underlying object has the method
	objMethod := obj.GetMethod("Add")
	if objMethod == nil {
		t.Fatal("Underlying object should have Add method")
	}
}

// TestInterfaceInheritance tests interface inheritance at runtime
func TestInterfaceInheritance(t *testing.T) {
	interp := testInterpreter()

	// Create base interface
	baseDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{Value: "IBase"},
		Methods: []*ast.InterfaceMethodDecl{
			{Name: &ast.Identifier{Value: "BaseMethod"}},
		},
	}
	interp.evalInterfaceDeclaration(baseDecl)

	// Create derived interface
	derivedDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name:   &ast.Identifier{Value: "IDerived"},
		Parent: &ast.Identifier{Value: "IBase"},
		Methods: []*ast.InterfaceMethodDecl{
			{Name: &ast.Identifier{Value: "DerivedMethod"}},
		},
	}
	interp.evalInterfaceDeclaration(derivedDecl)

	// Create class implementing derived interface (must implement both methods)
	classDecl := &ast.ClassDecl{
		Token: lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:  &ast.Identifier{Value: "TImplementation"},
		Methods: []*ast.FunctionDecl{
			{Name: &ast.Identifier{Value: "BaseMethod"}, Body: &ast.BlockStatement{}},
			{Name: &ast.Identifier{Value: "DerivedMethod"}, Body: &ast.BlockStatement{}},
		},
	}
	interp.evalClassDeclaration(classDecl)

	// Test that class implements derived interface
	obj := NewObjectInstance(interp.classes["TImplementation"])
	derivedIface := interp.interfaces["iderived"]

	if !classImplementsInterface(obj.Class, derivedIface) {
		t.Fatal("TImplementation should implement IDerived")
	}

	// Create interface instance
	ifaceInstance := NewInterfaceInstance(derivedIface, obj)

	// Verify both methods accessible through interface
	if !ifaceInstance.Interface.HasMethod("DerivedMethod") {
		t.Error("Should have DerivedMethod")
	}

	if !ifaceInstance.Interface.HasMethod("BaseMethod") {
		t.Error("Should inherit BaseMethod from parent interface")
	}

	// Test that class also implements base interface
	baseIface := interp.interfaces["ibase"]
	if !classImplementsInterface(obj.Class, baseIface) {
		t.Error("TImplementation should also implement IBase")
	}
}

// TestMultipleInterfaces tests class implementing multiple interfaces
func TestMultipleInterfaces(t *testing.T) {
	interp := testInterpreter()

	// Create first interface
	interface1 := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{Value: "IReadable"},
		Methods: []*ast.InterfaceMethodDecl{
			{Name: &ast.Identifier{Value: "Read"}},
		},
	}
	interp.evalInterfaceDeclaration(interface1)

	// Create second interface
	interface2 := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{Value: "IWritable"},
		Methods: []*ast.InterfaceMethodDecl{
			{Name: &ast.Identifier{Value: "Write"}},
		},
	}
	interp.evalInterfaceDeclaration(interface2)

	// Create class implementing both interfaces
	classDecl := &ast.ClassDecl{
		Token: lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:  &ast.Identifier{Value: "TFile"},
		Methods: []*ast.FunctionDecl{
			{Name: &ast.Identifier{Value: "Read"}, Body: &ast.BlockStatement{}},
			{Name: &ast.Identifier{Value: "Write"}, Body: &ast.BlockStatement{}},
		},
	}
	interp.evalClassDeclaration(classDecl)

	// Test that class implements both interfaces
	obj := NewObjectInstance(interp.classes["TFile"])

	if !classImplementsInterface(obj.Class, interp.interfaces["ireadable"]) {
		t.Error("TFile should implement IReadable")
	}

	if !classImplementsInterface(obj.Class, interp.interfaces["iwritable"]) {
		t.Error("TFile should implement IWritable")
	}

	// Create interface instances for both interfaces
	readableInstance := NewInterfaceInstance(interp.interfaces["ireadable"], obj)
	writableInstance := NewInterfaceInstance(interp.interfaces["iwritable"], obj)

	// Verify both wrap the same object
	if readableInstance.Object != obj || writableInstance.Object != obj {
		t.Error("Both interfaces should wrap the same object")
	}

	// Verify correct interface types
	if readableInstance.Interface.Name != "IReadable" {
		t.Error("First instance should be IReadable")
	}

	if writableInstance.Interface.Name != "IWritable" {
		t.Error("Second instance should be IWritable")
	}
}

// TestInterfaceToInterface tests interface-to-interface casting
func TestInterfaceToInterface(t *testing.T) {
	interp := testInterpreter()

	// Create base interface
	baseDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{Value: "IAnimal"},
		Methods: []*ast.InterfaceMethodDecl{
			{Name: &ast.Identifier{Value: "Eat"}},
		},
	}
	interp.evalInterfaceDeclaration(baseDecl)

	// Create derived interface
	derivedDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name:   &ast.Identifier{Value: "IDog"},
		Parent: &ast.Identifier{Value: "IAnimal"},
		Methods: []*ast.InterfaceMethodDecl{
			{Name: &ast.Identifier{Value: "Bark"}},
		},
	}
	interp.evalInterfaceDeclaration(derivedDecl)

	// Test upcast: IDog → IAnimal (should succeed)
	dogIface := interp.interfaces["idog"]
	animalIface := interp.interfaces["ianimal"]

	// IDog has Eat (inherited) and Bark, so it's compatible with IAnimal
	if !interfaceIsCompatible(dogIface, animalIface) {
		t.Error("IDog should be compatible with IAnimal (upcast)")
	}

	// Test downcast: IAnimal → IDog (should fail - IAnimal doesn't have Bark)
	if interfaceIsCompatible(animalIface, dogIface) {
		t.Error("IAnimal should NOT be compatible with IDog (downcast without Bark)")
	}

	// Test unrelated interfaces
	unrelatedDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{Value: "ICar"},
		Methods: []*ast.InterfaceMethodDecl{
			{Name: &ast.Identifier{Value: "Drive"}},
		},
	}
	interp.evalInterfaceDeclaration(unrelatedDecl)

	carIface := interp.interfaces["icar"]
	if interfaceIsCompatible(dogIface, carIface) {
		t.Error("IDog should NOT be compatible with ICar (unrelated)")
	}
}

// TestInterfaceToObject tests interface-to-object casting
func TestInterfaceToObject(t *testing.T) {
	interp := testInterpreter()

	// Create interface
	interfaceDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{Value: "IShape"},
		Methods: []*ast.InterfaceMethodDecl{
			{Name: &ast.Identifier{Value: "GetArea"}},
		},
	}
	interp.evalInterfaceDeclaration(interfaceDecl)

	// Create class
	classDecl := &ast.ClassDecl{
		Token: lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:  &ast.Identifier{Value: "TCircle"},
		Fields: []*ast.FieldDecl{
			{
				Name: &ast.Identifier{Value: "Radius"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
			},
		},
		Methods: []*ast.FunctionDecl{
			{Name: &ast.Identifier{Value: "GetArea"}, Body: &ast.BlockStatement{}},
		},
	}
	interp.evalClassDeclaration(classDecl)

	// Create object and interface instance
	obj := NewObjectInstance(interp.classes["TCircle"])
	obj.SetField("Radius", &IntegerValue{Value: 10})

	ifaceInstance := NewInterfaceInstance(interp.interfaces["ishape"], obj)

	// Test extracting underlying object
	extracted := ifaceInstance.GetUnderlyingObject()

	if extracted != obj {
		t.Error("Should extract the original object")
	}

	// Verify we can access object fields through extracted reference
	radiusVal := extracted.GetField("Radius")
	if radiusVal == nil {
		t.Fatal("Should be able to access field through extracted object")
	}

	radiusInt, ok := radiusVal.(*IntegerValue)
	if !ok || radiusInt.Value != 10 {
		t.Error("Field value should be preserved")
	}

	// Verify object type is correct
	if extracted.Class.Name != "TCircle" {
		t.Errorf("Extracted object should be TCircle, got %s", extracted.Class.Name)
	}
}

// TestInterfaceLifetime tests interface lifetime and scope
func TestInterfaceLifetime(t *testing.T) {
	interp := testInterpreter()

	// Create interface and class
	interfaceDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{Value: "IResource"},
		Methods: []*ast.InterfaceMethodDecl{
			{Name: &ast.Identifier{Value: "Release"}},
		},
	}
	interp.evalInterfaceDeclaration(interfaceDecl)

	classDecl := &ast.ClassDecl{
		Token: lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:  &ast.Identifier{Value: "TResource"},
		Methods: []*ast.FunctionDecl{
			{Name: &ast.Identifier{Value: "Release"}, Body: &ast.BlockStatement{}},
		},
	}
	interp.evalClassDeclaration(classDecl)

	// Test 1: Interface holds reference to object
	obj := NewObjectInstance(interp.classes["TResource"])
	ifaceInstance := NewInterfaceInstance(interp.interfaces["iresource"], obj)

	// Object should be accessible through interface
	if ifaceInstance.Object != obj {
		t.Error("Interface should maintain reference to object")
	}

	// Test 2: Multiple interface references to same object
	ifaceInstance2 := NewInterfaceInstance(interp.interfaces["iresource"], obj)

	if ifaceInstance.Object != ifaceInstance2.Object {
		t.Error("Multiple interfaces should reference same object")
	}

	// Test 3: Interface in nested scope
	{
		nestedEnv := NewEnclosedEnvironment(interp.env)
		savedEnv := interp.env
		interp.env = nestedEnv

		// Define interface variable in nested scope
		interp.env.Define("localInterface", ifaceInstance)

		val, exists := interp.env.Get("localInterface")
		if !exists {
			t.Error("Interface should exist in nested scope")
		}

		retrieved := val.(*InterfaceInstance)
		if retrieved.Object != obj {
			t.Error("Nested scope interface should reference same object")
		}

		// Restore environment
		interp.env = savedEnv
	}

	// Test 4: Object remains valid after interface created
	// (Go GC handles this automatically - we just verify references work)
	if ifaceInstance.Object.Class.Name != "TResource" {
		t.Error("Object should remain valid while interface reference exists")
	}
}

// TestInterfacePolymorphism tests interface polymorphism
func TestInterfacePolymorphism(t *testing.T) {
	interp := testInterpreter()

	// Create base interface
	baseDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name: &ast.Identifier{Value: "IVehicle"},
		Methods: []*ast.InterfaceMethodDecl{
			{Name: &ast.Identifier{Value: "Start"}},
		},
	}
	interp.evalInterfaceDeclaration(baseDecl)

	// Create derived interface
	derivedDecl := &ast.InterfaceDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		},
		Name:   &ast.Identifier{Value: "ICar"},
		Parent: &ast.Identifier{Value: "IVehicle"},
		Methods: []*ast.InterfaceMethodDecl{
			{Name: &ast.Identifier{Value: "Drive"}},
		},
	}
	interp.evalInterfaceDeclaration(derivedDecl)

	// Create class implementing derived interface
	classDecl := &ast.ClassDecl{
		Token: lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:  &ast.Identifier{Value: "TSportsCar"},
		Methods: []*ast.FunctionDecl{
			{Name: &ast.Identifier{Value: "Start"}, Body: &ast.BlockStatement{}},
			{Name: &ast.Identifier{Value: "Drive"}, Body: &ast.BlockStatement{}},
		},
	}
	interp.evalClassDeclaration(classDecl)

	obj := NewObjectInstance(interp.classes["TSportsCar"])

	// Test 1: Variable of type IVehicle can hold ICar instance
	carIface := NewInterfaceInstance(interp.interfaces["icar"], obj)

	// Store in variable as IVehicle (base interface)
	interp.env.Define("vehicle", carIface)

	val, _ := interp.env.Get("vehicle")
	vehicleIface := val.(*InterfaceInstance)

	// Should still be ICar type
	if vehicleIface.Interface.Name != "ICar" {
		t.Errorf("Interface type should be preserved as ICar, got %s", vehicleIface.Interface.Name)
	}

	// Test 2: Can cast to base interface
	baseIface := NewInterfaceInstance(interp.interfaces["ivehicle"], obj)

	if baseIface.Interface.Name != "IVehicle" {
		t.Error("Should be able to create IVehicle instance")
	}

	if baseIface.Object != obj {
		t.Error("Both interfaces should wrap same object")
	}

	// Test 3: Base interface has base methods
	if !baseIface.Interface.HasMethod("Start") {
		t.Error("IVehicle should have Start method")
	}

	// Test 4: Derived interface has all methods
	if !carIface.Interface.HasMethod("Start") {
		t.Error("ICar should inherit Start from IVehicle")
	}

	if !carIface.Interface.HasMethod("Drive") {
		t.Error("ICar should have Drive method")
	}
}

// ============================================================================
// Helper functions
// ============================================================================

func testInterpreter() *Interpreter {
	return New(nil)
}
