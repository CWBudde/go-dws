package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestIntegration_InterfaceDeclarationAndUsage tests complete interface workflow
func TestIntegration_InterfaceDeclarationAndUsage(t *testing.T) {
	source := `
		type
			IPrintable = interface
				procedure Print;
				function GetName: String;
			end;

		type
			TDocument = class(TObject, IPrintable)
				FName: String;

				procedure Print; begin PrintLn('Printing document'); end;
				function GetName: String; begin Result := FName; end;
			end;

		var doc: TDocument;
		doc := TDocument.Create();
		doc.FName := 'MyDoc';
		PrintLn(doc.GetName());
		doc.Print();
	`

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("Runtime error: %v", result.String())
	}

	// Verify interface was registered
	if _, exists := interp.interfaces["iprintable"]; !exists {
		t.Error("IPrintable interface should be registered")
	}

	// Verify class was registered
	class, exists := interp.classes["tdocument"]
	if !exists {
		t.Fatal("TDocument class should be registered")
	}

	// Verify class implements interface
	iface := interp.interfaces["iprintable"]
	if !classImplementsInterface(class, iface) {
		t.Error("TDocument should implement IPrintable")
	}

	// Verify output
	expectedOutput := "MyDoc\nPrinting document\n"
	if buf.String() != expectedOutput {
		t.Errorf("Output mismatch:\nExpected: %q\nActual: %q", expectedOutput, buf.String())
	}
}

// TestIntegration_InterfaceInheritanceHierarchy tests 3-level deep interface inheritance
func TestIntegration_InterfaceInheritanceHierarchy(t *testing.T) {
	source := `
		type
			IBase = interface
				procedure BaseMethod;
			end;

		type
			IMiddle = interface(IBase)
				procedure MiddleMethod;
			end;

		type
			IDerived = interface(IMiddle)
				procedure DerivedMethod;
			end;

		type
			TImplementation = class(TObject, IDerived)
				procedure BaseMethod; begin PrintLn('Base'); end;
				procedure MiddleMethod; begin PrintLn('Middle'); end;
				procedure DerivedMethod; begin PrintLn('Derived'); end;
			end;

		var instance: TImplementation;
		instance := TImplementation.Create();
		instance.BaseMethod();
		instance.MiddleMethod();
		instance.DerivedMethod();
	`

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("Runtime error: %v", result.String())
	}

	// Verify 3-level interface hierarchy
	base, existsBase := interp.interfaces["ibase"]
	if !existsBase {
		t.Fatal("IBase should be registered")
	}

	middle, existsMiddle := interp.interfaces["imiddle"]
	if !existsMiddle {
		t.Fatal("IMiddle should be registered")
	}

	derived, existsDerived := interp.interfaces["iderived"]
	if !existsDerived {
		t.Fatal("IDerived should be registered")
	}

	// Verify parent relationships
	if middle.Parent != base {
		t.Error("IMiddle's parent should be IBase")
	}

	if derived.Parent != middle {
		t.Error("IDerived's parent should be IMiddle")
	}

	// Verify method inheritance through all levels
	if !derived.HasMethod("BaseMethod") {
		t.Error("IDerived should inherit BaseMethod")
	}

	if !derived.HasMethod("MiddleMethod") {
		t.Error("IDerived should inherit MiddleMethod")
	}

	if !derived.HasMethod("DerivedMethod") {
		t.Error("IDerived should have DerivedMethod")
	}

	// Verify all methods count (should be 3)
	allMethods := derived.AllMethods()
	if len(allMethods) != 3 {
		t.Errorf("IDerived should have 3 methods total, got %d", len(allMethods))
	}

	// Verify class implements all levels
	class := interp.classes["timplementation"]
	if !classImplementsInterface(class, base) {
		t.Error("TImplementation should implement IBase")
	}

	if !classImplementsInterface(class, middle) {
		t.Error("TImplementation should implement IMiddle")
	}

	if !classImplementsInterface(class, derived) {
		t.Error("TImplementation should implement IDerived")
	}

	// Verify output
	expectedOutput := "Base\nMiddle\nDerived\n"
	if buf.String() != expectedOutput {
		t.Errorf("Output mismatch:\nExpected: %q\nActual: %q", expectedOutput, buf.String())
	}
}

// TestIntegration_ClassImplementingMultipleInterfaces tests single class with 3+ interfaces
func TestIntegration_ClassImplementingMultipleInterfaces(t *testing.T) {
	source := `
		type
			IReadable = interface
				function GetData: String;
			end;

		type
			IWritable = interface
				procedure SetData(s: String);
			end;

		type
			ICloseable = interface
				procedure Close;
			end;

		type
			TFile = class(TObject, IReadable, IWritable, ICloseable)
				FContent: String;
				FOpen: Boolean;

				function GetData: String;
				begin
					Result := FContent;
				end;

				procedure SetData(s: String);
				begin
					FContent := s;
				end;

				procedure Close;
				begin
					FOpen := False;
					PrintLn('File closed');
				end;
			end;

		var f: TFile;
		f := TFile.Create();
		f.FOpen := True;
		f.SetData('Hello, World');
		PrintLn(f.GetData());
		f.Close();
	`

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("Runtime error: %v", result.String())
	}

	// Verify all three interfaces registered
	readable, existsR := interp.interfaces["ireadable"]
	if !existsR {
		t.Fatal("IReadable should be registered")
	}

	writable, existsW := interp.interfaces["iwritable"]
	if !existsW {
		t.Fatal("IWritable should be registered")
	}

	closeable, existsC := interp.interfaces["icloseable"]
	if !existsC {
		t.Fatal("ICloseable should be registered")
	}

	// Verify class implements all three
	class := interp.classes["tfile"]
	if !classImplementsInterface(class, readable) {
		t.Error("TFile should implement IReadable")
	}

	if !classImplementsInterface(class, writable) {
		t.Error("TFile should implement IWritable")
	}

	if !classImplementsInterface(class, closeable) {
		t.Error("TFile should implement ICloseable")
	}

	// Verify output
	expectedOutput := "Hello, World\nFile closed\n"
	if buf.String() != expectedOutput {
		t.Errorf("Output mismatch:\nExpected: %q\nActual: %q", expectedOutput, buf.String())
	}
}

// TestIntegration_InterfaceCastingAllCombinations tests all casting combinations
func TestIntegration_InterfaceCastingAllCombinations(t *testing.T) {
	t.Run("ObjectToInterface", func(t *testing.T) {
		// Create interface
		iface := NewInterfaceInfo("ITest")
		iface.Methods[strings.ToLower("DoIt")] = &ast.FunctionDecl{
			Name: &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{},
				},
				Value: "DoIt",
			},
		}

		// Create class that implements interface
		class := NewClassInfo("TTest")
		class.Methods["doit"] = &ast.FunctionDecl{
			Name: &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{},
				},
				Value: "DoIt",
			},
		}
		class.Interfaces = append(class.Interfaces, iface)

		// Create object instance
		obj := NewObjectInstance(class)

		// Test: Object → Interface cast
		if !classImplementsInterface(class, iface) {
			t.Fatal("Class should implement interface")
		}

		ifaceInstance := NewInterfaceInstance(iface, obj)
		if ifaceInstance.Object != obj {
			t.Error("Interface should wrap the object")
		}

		if ifaceInstance.Interface != iface {
			t.Error("Interface instance should reference correct interface")
		}
	})

	t.Run("InterfaceToObject", func(t *testing.T) {
		// Create interface and class
		iface := NewInterfaceInfo("ITest")
		class := NewClassInfo("TTest")
		class.Methods["doit"] = &ast.FunctionDecl{Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "DoIt"}}
		// Register field so SetField uses normalized metadata/legacy paths
		class.Fields["testfield"] = types.INTEGER
		runtime.AddFieldToClass(class.Metadata, &runtime.FieldMetadata{
			Name:       "TestField",
			Type:       types.INTEGER,
			TypeName:   "Integer",
			Visibility: runtime.FieldVisibilityPublic,
		})

		// Create object and interface instance
		obj := NewObjectInstance(class)
		// Task 3.5.40: Use SetField for proper normalization
		obj.SetField("TestField", &IntegerValue{Value: 42})
		ifaceInstance := NewInterfaceInstance(iface, obj)

		// Test: Interface → Object cast
		extracted := ifaceInstance.GetUnderlyingObject()
		if extracted != obj {
			t.Error("Should extract original object")
		}

		// Verify object class is preserved
		if extracted.Class.Name != "TTest" {
			t.Errorf("Extracted object should be TTest, got %s", extracted.Class.Name)
		}

		// Verify object fields preserved (case-insensitive lookup)
		fieldVal := extracted.GetField("TestField")
		if fieldVal == nil {
			t.Fatal("TestField should exist in object")
		}

		intVal, ok := fieldVal.(*IntegerValue)
		if !ok || intVal.Value != 42 {
			t.Error("Field value should be preserved")
		}
	})

	t.Run("InterfaceToInterface_Upcast", func(t *testing.T) {
		// Create base interface
		base := NewInterfaceInfo("IBase")
		base.Methods["basemethod"] = &ast.FunctionDecl{Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "BaseMethod"}}

		// Create derived interface
		derived := NewInterfaceInfo("IDerived")
		derived.Parent = base
		derived.Methods["derivedmethod"] = &ast.FunctionDecl{Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "DerivedMethod"}}

		// Test: IDerived → IBase (upcast)
		if !interfaceIsCompatible(derived, base) {
			t.Error("Derived interface should be compatible with base (upcast)")
		}
	})

	t.Run("InterfaceToInterface_Downcast", func(t *testing.T) {
		// Create base interface
		base := NewInterfaceInfo("IBase")
		base.Methods["basemethod"] = &ast.FunctionDecl{Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "BaseMethod"}}

		// Create derived interface
		derived := NewInterfaceInfo("IDerived")
		derived.Parent = base
		derived.Methods["derivedmethod"] = &ast.FunctionDecl{Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "DerivedMethod"}}

		// Test: IBase → IDerived (downcast - should fail)
		if interfaceIsCompatible(base, derived) {
			t.Error("Base interface should NOT be compatible with derived (downcast)")
		}
	})

	t.Run("NilInterfaceCasting", func(t *testing.T) {
		// Create interface
		iface := NewInterfaceInfo("ITest")

		// Create interface instance with nil object
		var nilObj *ObjectInstance = nil
		ifaceInstance := NewInterfaceInstance(iface, nilObj)

		// Verify nil is preserved
		if ifaceInstance.Object != nil {
			t.Error("Interface with nil object should preserve nil")
		}

		// Extracting should return nil
		extracted := ifaceInstance.GetUnderlyingObject()
		if extracted != nil {
			t.Error("Extracting from nil interface should return nil")
		}
	})
}

// TestIntegration_InterfaceLifetimeManagement tests interface lifetime and scope
func TestIntegration_InterfaceLifetimeManagement(t *testing.T) {
	t.Run("VariableLifetime", func(t *testing.T) {
		interp := New(nil)

		// Create interface and class
		iface := NewInterfaceInfo("IResource")
		iface.Methods[strings.ToLower("Release")] = &ast.FunctionDecl{Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "Release"}}

		class := NewClassInfo("TResource")
		class.Methods["release"] = &ast.FunctionDecl{Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "Release"}}

		interp.interfaces["iresource"] = iface
		interp.classes["TResource"] = class

		// Create object and interface instance
		obj := NewObjectInstance(class)
		ifaceInstance := NewInterfaceInstance(iface, obj)

		// Store in environment
		interp.env.Define("resource", ifaceInstance)

		// Retrieve and verify
		retrieved, exists := interp.env.Get("resource")
		if !exists {
			t.Fatal("Interface variable should exist")
		}

		retrievedIface := retrieved.(*InterfaceInstance)
		if retrievedIface.Object != obj {
			t.Error("Retrieved interface should reference same object")
		}
	})

	t.Run("ScopeIsolation", func(t *testing.T) {
		interp := New(nil)

		// Create interface and object
		iface := NewInterfaceInfo("ITest")
		class := NewClassInfo("TTest")
		obj := NewObjectInstance(class)
		ifaceInstance := NewInterfaceInstance(iface, obj)

		// Define in outer scope
		interp.env.Define("outer", ifaceInstance)

		// Create nested scope
		nestedEnv := NewEnclosedEnvironment(interp.env)
		savedEnv := interp.env
		interp.env = nestedEnv

		// Define different instance in nested scope
		obj2 := NewObjectInstance(class)
		ifaceInstance2 := NewInterfaceInstance(iface, obj2)
		interp.env.Define("nested", ifaceInstance2)

		// Verify nested scope has both
		_, exists := interp.env.Get("nested")
		if !exists {
			t.Error("Nested variable should exist in nested scope")
		}

		_, exists2 := interp.env.Get("outer")
		if !exists2 {
			t.Error("Outer variable should be accessible from nested scope")
		}

		// Restore environment
		interp.env = savedEnv

		// Verify outer scope doesn't have nested variable
		_, exists3 := interp.env.Get("nested")
		if exists3 {
			t.Error("Nested variable should not exist in outer scope")
		}
	})

	t.Run("MultipleReferences", func(t *testing.T) {
		// Create interface and object
		iface := NewInterfaceInfo("IShared")
		class := NewClassInfo("TShared")
		obj := NewObjectInstance(class)
		// Task 3.5.40: Use SetField for proper normalization
		obj.SetField("Counter", &IntegerValue{Value: 0})

		// Create multiple interface instances referencing same object
		iface1 := NewInterfaceInstance(iface, obj)
		iface2 := NewInterfaceInstance(iface, obj)
		iface3 := NewInterfaceInstance(iface, obj)

		// All should reference same object
		if iface1.Object != obj || iface2.Object != obj || iface3.Object != obj {
			t.Error("All interface instances should reference same object")
		}

		// Modify through one interface (direct access to Fields map)
		iface1.Object.Fields["Counter"] = &IntegerValue{Value: 42}

		// Verify visible through all interfaces
		field2, exists2 := iface2.Object.Fields["Counter"]
		if !exists2 {
			t.Fatal("Counter field should exist")
		}
		val2 := field2.(*IntegerValue).Value

		field3, exists3 := iface3.Object.Fields["Counter"]
		if !exists3 {
			t.Fatal("Counter field should exist")
		}
		val3 := field3.(*IntegerValue).Value

		if val2 != 42 || val3 != 42 {
			t.Error("Changes through one interface should be visible through all")
		}
	})
}
