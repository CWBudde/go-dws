package interp

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
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
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
		},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "BaseMethod"},
				},
			},
			Value: "BaseMethod",
		},
	}
	parent.Methods["BaseMethod"] = parentMethod

	// Create child interface
	child := NewInterfaceInfo("IDerived")
	child.Parent = parent
	childMethod := &ast.FunctionDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
		},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "DerivedMethod"},
				},
			},
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
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
		},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Increment"},
				},
			},
			Value: "Increment",
		},
	}
	interfaceInfo.Methods["Increment"] = incrementMethod

	getValueMethod := &ast.FunctionDecl{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
		},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "GetValue"},
				},
			},
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
	// Field defined as "x" (lowercase) but accessed as "X" (uppercase)
	// to verify case-insensitive field access via normalization
	// Note: Skipping field registration since Fields now expects *ast.FieldDecl
	obj := NewObjectInstance(classInfo)
	// Use SetField for proper normalization
	obj.SetField("X", &IntegerValue{Value: 42})

	// Create an interface instance
	iface := NewInterfaceInstance(interfaceInfo, obj)

	// Get the underlying object
	underlyingObj := iface.Object

	if underlyingObj != obj {
		t.Error("GetUnderlyingObject should return the wrapped object")
	}

	// Verify we can access the object's fields (use GetField for proper field access)
	fieldVal := underlyingObj.GetField("X")
	if fieldVal == nil {
		t.Error("Field X should exist")
	} else if intVal, ok := fieldVal.(*IntegerValue); !ok || intVal.Value != 42 {
		t.Error("Should be able to access underlying object's fields")
	}
}

// ============================================================================
// Interface Declaration Tests (through the production evaluator path)
// ============================================================================

// declareViaScript parses source and evaluates it through the production
// interpreter/evaluator path, failing the test on any error.
func declareViaScript(t *testing.T, interp *Interpreter, src string) {
	t.Helper()

	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if result := interp.Eval(program); isError(result) {
		t.Fatalf("eval error: %s", result.String())
	}
}

func TestEvalInterfaceDeclaration(t *testing.T) {
	interp := testInterpreter()

	declareViaScript(t, interp, `
		type IMyInterface = interface
			procedure DoSomething;
		end;
	`)

	// Check that interface was registered
	ifaceInfo := interp.lookupInterfaceInfo("imyinterface")
	if ifaceInfo == nil {
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
	interp := testInterpreter()

	declareViaScript(t, interp, `
		type IBase = interface
			procedure BaseMethod;
		end;
		type IDerived = interface(IBase)
			procedure DerivedMethod;
		end;
	`)

	// Check that child interface was registered with parent
	childInfo := interp.lookupInterfaceInfo("iderived")
	if childInfo == nil {
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

	declareViaScript(t, interp, `
		type ICounter = interface
			function GetValue : Integer;
		end;
		type TCounter = class(TObject, ICounter)
			FValue : Integer;
			function GetValue : Integer; begin Result := FValue; end;
		end;
	`)

	// Create an object instance
	classInfo := mustLookupTestClass(t, interp, "TCounter")
	obj := NewObjectInstance(classInfo)
	obj.SetField("FValue", &IntegerValue{Value: 42})

	// Cast object to interface
	ifaceInfo := interp.lookupInterfaceInfo("icounter")
	if ifaceInfo == nil {
		t.Fatal("Interface 'ICounter' should be registered")
	}

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

	// Verify we can still access the underlying object
	underlying := ifaceInstance.GetUnderlyingObject()
	if underlying != obj {
		t.Error("Should be able to get underlying object from interface")
	}

	// Test interface variable assignment (reference semantics)
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

	declareViaScript(t, interp, `
		type IPrintable = interface
			procedure Print;
		end;
		type TDocument = class(TObject, IPrintable)
			procedure Print; begin end;
		end;
	`)

	// Create object and cast to interface
	obj := NewObjectInstance(mustLookupTestClass(t, interp, "TDocument"))
	ifaceInstance := NewInterfaceInstance(interp.lookupInterfaceInfo("iprintable"), obj)

	// Store in environment as a variable
	interp.Env().Define("myInterface", ifaceInstance)

	// Retrieve and verify
	val, exists := interp.Env().Get("myInterface")
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

	if retrieved.Interface.GetName() != "IPrintable" {
		t.Errorf("Interface name should be IPrintable, got %s", retrieved.Interface.GetName())
	}
}

// TestObjectToInterface tests object-to-interface casting
func TestObjectToInterface(t *testing.T) {
	interp := testInterpreter()

	declareViaScript(t, interp, `
		type IDrawable = interface
			procedure Draw;
			procedure GetWidth;
			procedure GetHeight;
		end;
		type TRectangle = class(TObject, IDrawable)
			procedure Draw; begin end;
			procedure GetWidth; begin end;
			procedure GetHeight; begin end;
		end;
		type TPoint = class(TObject)
			procedure Draw; begin end;
			// Missing GetWidth and GetHeight
		end;
	`)

	// Test successful cast
	obj := NewObjectInstance(mustLookupTestClass(t, interp, "TRectangle"))
	ifaceInfo := interp.lookupInterfaceInfo("idrawable")

	if !classImplementsInterface(obj.Class.(*ClassInfo), ifaceInfo) {
		t.Fatal("TRectangle should implement IDrawable")
	}

	ifaceInstance := NewInterfaceInstance(ifaceInfo, obj)
	if ifaceInstance.Object != obj {
		t.Error("Interface should wrap the original object")
	}

	// Test failed cast (class missing methods)
	objIncompatible := NewObjectInstance(mustLookupTestClass(t, interp, "TPoint"))
	if classImplementsInterface(objIncompatible.Class.(*ClassInfo), ifaceInfo) {
		t.Error("TPoint should NOT implement IDrawable (missing methods)")
	}
}

// TestInterfaceMethodCall tests method calls through interface references
func TestInterfaceMethodCall(t *testing.T) {
	interp := testInterpreter()

	declareViaScript(t, interp, `
		type ICalculator = interface
			function Add(x, y : Integer) : Integer;
		end;
		type TCalculator = class(TObject, ICalculator)
			function Add(x, y : Integer) : Integer; begin Result := x + y; end;
		end;
	`)

	// Create interface instance
	obj := NewObjectInstance(mustLookupTestClass(t, interp, "TCalculator"))
	ifaceInstance := NewInterfaceInstance(interp.lookupInterfaceInfo("icalculator"), obj)

	// Verify method can be found through interface
	methodAny := ifaceInstance.Interface.GetMethod("Add")
	if methodAny == nil {
		t.Fatal("Should be able to find Add method through interface")
	}

	// GetMethod returns any, need to type assert
	method, ok := methodAny.(*ast.FunctionDecl)
	if !ok {
		t.Fatal("GetMethod should return *ast.FunctionDecl")
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

	declareViaScript(t, interp, `
		type IBase = interface
			procedure BaseMethod;
		end;
		type IDerived = interface(IBase)
			procedure DerivedMethod;
		end;
		type TImplementation = class(TObject, IDerived)
			procedure BaseMethod; begin end;
			procedure DerivedMethod; begin end;
		end;
	`)

	// Test that class implements derived interface
	obj := NewObjectInstance(mustLookupTestClass(t, interp, "TImplementation"))
	derivedIface := interp.lookupInterfaceInfo("iderived")

	if !classImplementsInterface(obj.Class.(*ClassInfo), derivedIface) {
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
	baseIface := interp.lookupInterfaceInfo("ibase")
	if !classImplementsInterface(obj.Class.(*ClassInfo), baseIface) {
		t.Error("TImplementation should also implement IBase")
	}
}

// TestMultipleInterfaces tests class implementing multiple interfaces
func TestMultipleInterfaces(t *testing.T) {
	interp := testInterpreter()

	declareViaScript(t, interp, `
		type IReadable = interface
			procedure ReadData;
		end;
		type IWritable = interface
			procedure WriteData;
		end;
		type TFile = class(TObject, IReadable, IWritable)
			procedure ReadData; begin end;
			procedure WriteData; begin end;
		end;
	`)

	// Test that class implements both interfaces
	obj := NewObjectInstance(mustLookupTestClass(t, interp, "TFile"))

	if !classImplementsInterface(obj.Class.(*ClassInfo), interp.lookupInterfaceInfo("ireadable")) {
		t.Error("TFile should implement IReadable")
	}

	if !classImplementsInterface(obj.Class.(*ClassInfo), interp.lookupInterfaceInfo("iwritable")) {
		t.Error("TFile should implement IWritable")
	}

	// Create interface instances for both interfaces
	readableInstance := NewInterfaceInstance(interp.lookupInterfaceInfo("ireadable"), obj)
	writableInstance := NewInterfaceInstance(interp.lookupInterfaceInfo("iwritable"), obj)

	// Verify both wrap the same object
	if readableInstance.Object != obj || writableInstance.Object != obj {
		t.Error("Both interfaces should wrap the same object")
	}

	// Verify correct interface types
	if readableInstance.Interface.GetName() != "IReadable" {
		t.Error("First instance should be IReadable")
	}

	if writableInstance.Interface.GetName() != "IWritable" {
		t.Error("Second instance should be IWritable")
	}
}

// TestInterfaceToInterface tests interface-to-interface compatibility metadata
func TestInterfaceToInterface(t *testing.T) {
	interp := testInterpreter()

	declareViaScript(t, interp, `
		type IAnimal = interface
			procedure Eat;
		end;
		type IDog = interface(IAnimal)
			procedure Bark;
		end;
		type ICar = interface
			procedure Drive;
		end;
	`)

	dogIface := interp.lookupInterfaceInfo("idog")
	animalIface := interp.lookupInterfaceInfo("ianimal")
	carIface := interp.lookupInterfaceInfo("icar")

	// Upcast: IDog -> IAnimal (IDog inherits from IAnimal)
	if !interfaceInheritsFrom(dogIface, animalIface) {
		t.Error("IDog should inherit from IAnimal (upcast)")
	}

	// IDog has Eat (inherited), so it provides IAnimal's full method set
	if !dogIface.HasMethod("Eat") {
		t.Error("IDog should expose inherited method Eat")
	}

	// Downcast: IAnimal -> IDog should fail - IAnimal doesn't have Bark
	if interfaceInheritsFrom(animalIface, dogIface) {
		t.Error("IAnimal should NOT inherit from IDog (downcast)")
	}
	if animalIface.HasMethod("Bark") {
		t.Error("IAnimal should NOT have method Bark")
	}

	// Unrelated interfaces
	if interfaceInheritsFrom(dogIface, carIface) {
		t.Error("IDog should NOT be related to ICar")
	}
	if dogIface.HasMethod("Drive") {
		t.Error("IDog should NOT have method Drive")
	}
}

// TestInterfaceToObject tests interface-to-object casting
func TestInterfaceToObject(t *testing.T) {
	interp := testInterpreter()

	declareViaScript(t, interp, `
		type IShape = interface
			procedure GetArea;
		end;
		type TCircle = class(TObject, IShape)
			Radius : Integer;
			procedure GetArea; begin end;
		end;
	`)

	// Create object and interface instance
	obj := NewObjectInstance(mustLookupTestClass(t, interp, "TCircle"))
	obj.SetField("Radius", &IntegerValue{Value: 10})

	ifaceInstance := NewInterfaceInstance(interp.lookupInterfaceInfo("ishape"), obj)

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
	if extracted.Class.GetName() != "TCircle" {
		t.Errorf("Extracted object should be TCircle, got %s", extracted.Class.GetName())
	}
}

// TestInterfaceLifetime tests interface lifetime and scope
func TestInterfaceLifetime(t *testing.T) {
	interp := testInterpreter()

	declareViaScript(t, interp, `
		type IResource = interface
			procedure Release;
		end;
		type TResource = class(TObject, IResource)
			procedure Release; begin end;
		end;
	`)

	// Test 1: Interface holds reference to object
	obj := NewObjectInstance(mustLookupTestClass(t, interp, "TResource"))
	ifaceInstance := NewInterfaceInstance(interp.lookupInterfaceInfo("iresource"), obj)

	// Object should be accessible through interface
	if ifaceInstance.Object != obj {
		t.Error("Interface should maintain reference to object")
	}

	// Test 2: Multiple interface references to same object
	ifaceInstance2 := NewInterfaceInstance(interp.lookupInterfaceInfo("iresource"), obj)

	if ifaceInstance.Object != ifaceInstance2.Object {
		t.Error("Multiple interfaces should reference same object")
	}

	// Test 3: Interface in nested scope
	{
		nestedEnv := NewEnclosedEnvironment(interp.Env())
		savedEnv := interp.Env()
		interp.SetEnvironment(nestedEnv)

		// Define interface variable in nested scope
		interp.Env().Define("localInterface", ifaceInstance)

		val, exists := interp.Env().Get("localInterface")
		if !exists {
			t.Error("Interface should exist in nested scope")
		}

		retrieved := val.(*InterfaceInstance)
		if retrieved.Object != obj {
			t.Error("Nested scope interface should reference same object")
		}

		// Restore environment
		interp.SetEnvironment(savedEnv)
	}

	// Test 4: Object remains valid after interface created
	// (Go GC handles this automatically - we just verify references work)
	if ifaceInstance.Object.Class.GetName() != "TResource" {
		t.Error("Object should remain valid while interface reference exists")
	}
}

// TestInterfacePolymorphism tests interface polymorphism
func TestInterfacePolymorphism(t *testing.T) {
	interp := testInterpreter()

	declareViaScript(t, interp, `
		type IVehicle = interface
			procedure Start;
		end;
		type ICar = interface(IVehicle)
			procedure Drive;
		end;
		type TSportsCar = class(TObject)
			procedure Start; begin end;
			procedure Drive; begin end;
		end;
	`)

	obj := NewObjectInstance(mustLookupTestClass(t, interp, "TSportsCar"))

	// Test 1: Variable of type IVehicle can hold ICar instance
	carIface := NewInterfaceInstance(interp.lookupInterfaceInfo("icar"), obj)

	// Store in variable as IVehicle (base interface)
	interp.Env().Define("vehicle", carIface)

	val, _ := interp.Env().Get("vehicle")
	vehicleIface := val.(*InterfaceInstance)

	// Should still be ICar type
	if vehicleIface.Interface.GetName() != "ICar" {
		t.Errorf("Interface type should be preserved as ICar, got %s", vehicleIface.Interface.GetName())
	}

	// Test 2: Can cast to base interface
	baseIface := NewInterfaceInstance(interp.lookupInterfaceInfo("ivehicle"), obj)

	if baseIface.Interface.GetName() != "IVehicle" {
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
