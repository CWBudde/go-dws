package interp

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// ClassInfo Tests
// ============================================================================

func TestClassInfoCreation(t *testing.T) {
	// Create a simple class: TPoint with X, Y fields
	classInfo := NewClassInfo("TPoint")

	if classInfo.Name != "TPoint" {
		t.Errorf("classInfo.Name = %s, want TPoint", classInfo.Name)
	}

	if classInfo.Parent != nil {
		t.Errorf("classInfo.Parent should be nil for root class")
	}

	if classInfo.Fields == nil {
		t.Error("classInfo.Fields should be initialized")
	}

	if classInfo.Methods == nil {
		t.Error("classInfo.Methods should be initialized")
	}
}

func TestClassInfoWithInheritance(t *testing.T) {
	// Create parent class
	parent := NewClassInfo("TObject")
	parent.Fields["ID"] = types.INTEGER

	// Create child class
	child := NewClassInfo("TPerson")
	child.Parent = parent
	child.Fields["Name"] = types.STRING

	if child.Parent == nil {
		t.Fatal("child.Parent should not be nil")
	}

	if child.Parent.Name != "TObject" {
		t.Errorf("child.Parent.Name = %s, want TObject", child.Parent.Name)
	}
}

func TestClassInfoAddField(t *testing.T) {
	classInfo := NewClassInfo("TPerson")
	classInfo.Fields["Name"] = types.STRING
	classInfo.Fields["Age"] = types.INTEGER

	if len(classInfo.Fields) != 2 {
		t.Errorf("len(classInfo.Fields) = %d, want 2", len(classInfo.Fields))
	}

	if classInfo.Fields["Name"] != types.STRING {
		t.Error("Field 'Name' should be STRING type")
	}

	if classInfo.Fields["Age"] != types.INTEGER {
		t.Error("Field 'Age' should be INTEGER type")
	}
}

func TestClassInfoAddMethod(t *testing.T) {
	classInfo := NewClassInfo("TCounter")

	// Create a simple method AST node
	method := &ast.FunctionDecl{
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
		Parameters: []*ast.Parameter{},
		ReturnType: &ast.TypeAnnotation{Name: "Integer"},
		Body: &ast.BlockStatement{
			Statements: []ast.Statement{},
		},
	}

	// Methods are stored with lowercase keys for case-insensitive lookup
	classInfo.Methods["getvalue"] = method

	if len(classInfo.Methods) != 1 {
		t.Errorf("len(classInfo.Methods) = %d, want 1", len(classInfo.Methods))
	}

	if classInfo.Methods["getvalue"] == nil {
		t.Error("Method 'GetValue' should be registered")
	}

	if classInfo.Methods["getvalue"].Name.Value != "GetValue" {
		t.Error("Method name should be 'GetValue'")
	}
}

// ============================================================================
// ObjectInstance Tests
// ============================================================================

func TestObjectInstanceCreation(t *testing.T) {
	// Create class info
	classInfo := NewClassInfo("TPoint")
	classInfo.Fields["X"] = types.INTEGER
	classInfo.Fields["Y"] = types.INTEGER

	// Create object instance
	obj := NewObjectInstance(classInfo)

	if obj.Class == nil {
		t.Fatal("obj.Class should not be nil")
	}

	if obj.Class.Name != "TPoint" {
		t.Errorf("obj.Class.Name = %s, want TPoint", obj.Class.Name)
	}

	if obj.Fields == nil {
		t.Fatal("obj.Fields should be initialized")
	}
}

func TestObjectInstanceGetSetField(t *testing.T) {
	// Create class info
	classInfo := NewClassInfo("TPoint")
	classInfo.Fields["X"] = types.INTEGER
	classInfo.Fields["Y"] = types.INTEGER

	// Create object instance
	obj := NewObjectInstance(classInfo)

	// Set fields
	obj.SetField("X", NewIntegerValue(10))
	obj.SetField("Y", NewIntegerValue(20))

	// Get fields
	xVal := obj.GetField("X")
	yVal := obj.GetField("Y")

	if xVal == nil {
		t.Fatal("GetField('X') should not return nil")
	}

	if yVal == nil {
		t.Fatal("GetField('Y') should not return nil")
	}

	// Check values
	xInt, err := GoInt(xVal)
	if err != nil {
		t.Fatalf("GetField('X') should return integer: %v", err)
	}

	if xInt != 10 {
		t.Errorf("GetField('X') = %d, want 10", xInt)
	}

	yInt, err := GoInt(yVal)
	if err != nil {
		t.Fatalf("GetField('Y') should return integer: %v", err)
	}

	if yInt != 20 {
		t.Errorf("GetField('Y') = %d, want 20", yInt)
	}
}

func TestObjectInstanceGetUndefinedField(t *testing.T) {
	classInfo := NewClassInfo("TPoint")
	classInfo.Fields["X"] = types.INTEGER

	obj := NewObjectInstance(classInfo)

	// Try to get undefined field
	val := obj.GetField("NonExistent")

	if val != nil {
		t.Error("GetField for undefined field should return nil")
	}
}

func TestObjectInstanceInitializeFields(t *testing.T) {
	// Create class with default values
	classInfo := NewClassInfo("TPerson")
	classInfo.Fields["Name"] = types.STRING
	classInfo.Fields["Age"] = types.INTEGER
	classInfo.Fields["Active"] = types.BOOLEAN

	obj := NewObjectInstance(classInfo)

	// Fields should be nil until explicitly set
	name := obj.GetField("Name")
	age := obj.GetField("Age")
	active := obj.GetField("Active")

	if name != nil {
		t.Error("Uninitialized field 'Name' should be nil")
	}

	if age != nil {
		t.Error("Uninitialized field 'Age' should be nil")
	}

	if active != nil {
		t.Error("Uninitialized field 'Active' should be nil")
	}

	// Set and verify
	obj.SetField("Name", NewStringValue("Alice"))
	obj.SetField("Age", NewIntegerValue(30))
	obj.SetField("Active", NewBooleanValue(true))

	nameStr, _ := GoString(obj.GetField("Name"))
	if nameStr != "Alice" {
		t.Errorf("Name = %s, want Alice", nameStr)
	}

	ageInt, _ := GoInt(obj.GetField("Age"))
	if ageInt != 30 {
		t.Errorf("Age = %d, want 30", ageInt)
	}

	activeBool, _ := GoBool(obj.GetField("Active"))
	if !activeBool {
		t.Error("Active should be true")
	}
}

// ============================================================================
// Method Lookup Tests
// ============================================================================

func TestMethodLookupBasic(t *testing.T) {
	// Create class with method
	classInfo := NewClassInfo("TCounter")

	method := &ast.FunctionDecl{
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{},
			},
			Value: "GetValue",
		},
	}
	// Methods are stored with lowercase keys for case-insensitive lookup
	classInfo.Methods["getvalue"] = method

	// Create object
	obj := NewObjectInstance(classInfo)

	// Look up method
	foundMethod := obj.GetMethod("GetValue")

	if foundMethod == nil {
		t.Fatal("GetMethod('GetValue') should not return nil")
	}

	if foundMethod.Name.Value != "GetValue" {
		t.Error("Found method should have name 'GetValue'")
	}
}

func TestMethodLookupWithInheritance(t *testing.T) {
	// Create parent class with method
	parent := NewClassInfo("TObject")
	parentMethod := &ast.FunctionDecl{
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{},
			},
			Value: "ToString",
		},
	}
	// Methods are stored with lowercase keys for case-insensitive lookup
	parent.Methods["tostring"] = parentMethod

	// Create child class
	child := NewClassInfo("TPerson")
	child.Parent = parent

	// Create object of child class
	obj := NewObjectInstance(child)

	// Should find parent's method
	foundMethod := obj.GetMethod("ToString")

	if foundMethod == nil {
		t.Fatal("GetMethod should find parent's method")
	}

	if foundMethod.Name.Value != "ToString" {
		t.Error("Found method should be 'ToString' from parent")
	}
}

func TestMethodOverriding(t *testing.T) {
	// Create parent class with method
	parent := NewClassInfo("TObject")
	parentMethod := &ast.FunctionDecl{
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{},
			},
			Value: "ToString",
		},
		Body: &ast.BlockStatement{}, // Different body
	}
	// Methods are stored with lowercase keys for case-insensitive lookup
	parent.Methods["tostring"] = parentMethod

	// Create child class that overrides the method
	child := NewClassInfo("TPerson")
	child.Parent = parent

	childMethod := &ast.FunctionDecl{
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{},
			},
			Value: "ToString",
		},
		Body: &ast.BlockStatement{
			Statements: []ast.Statement{
				// Different implementation
			},
		},
	}
	// Methods are stored with lowercase keys for case-insensitive lookup
	child.Methods["tostring"] = childMethod

	// Create object of child class
	obj := NewObjectInstance(child)

	// Should find child's overridden method, not parent's
	foundMethod := obj.GetMethod("ToString")

	if foundMethod == nil {
		t.Fatal("GetMethod should find method")
	}

	// Verify it's the child's method (different body)
	if foundMethod != childMethod {
		t.Error("Should find child's overridden method, not parent's")
	}
}

func TestMethodLookupNotFound(t *testing.T) {
	classInfo := NewClassInfo("TPoint")
	obj := NewObjectInstance(classInfo)

	method := obj.GetMethod("NonExistent")

	if method != nil {
		t.Error("GetMethod for non-existent method should return nil")
	}
}

// ============================================================================
// ObjectValue Tests (ObjectInstance as a Value)
// ============================================================================

func TestObjectValue(t *testing.T) {
	classInfo := NewClassInfo("TPoint")
	obj := NewObjectInstance(classInfo)

	// ObjectInstance should implement Value interface
	var _ Value = obj

	if obj.Type() != "OBJECT" {
		t.Errorf("obj.Type() = %s, want OBJECT", obj.Type())
	}

	// String representation should show class name
	str := obj.String()
	if str == "" {
		t.Error("obj.String() should not be empty")
	}
}

// ============================================================================
// External Class Runtime Tests
// ============================================================================

func TestExternalClassRuntime(t *testing.T) {
	t.Run("external class instantiation returns error", func(t *testing.T) {
		input := `
type TExternal = class external
end;

var obj: TExternal;
obj := TExternal.Create();
`
		result := testEval(input)
		if result == nil {
			t.Error("Expected error value, got nil")
			return
		}

		// Check if result is an error
		if result.Type() != "ERROR" {
			t.Errorf("Expected ERROR type, got %s", result.Type())
			return
		}

		errMsg := result.String()
		if !containsInMiddle(errMsg, "external") && !containsInMiddle(errMsg, "not supported") {
			t.Errorf("Expected error about external class, got: %s", errMsg)
		}
	})

	t.Run("external method call returns error", func(t *testing.T) {
		input := `
type TExternal = class external
  procedure Hello; external 'world';
end;

var obj: TExternal;
obj := TExternal.Create();
obj.Hello();
`
		result := testEval(input)
		// Should fail at instantiation, not method call
		// But if it somehow gets to method call, should also fail
		if result == nil {
			t.Error("Expected error value, got nil")
			return
		}

		if result.Type() != "ERROR" {
			t.Errorf("Expected ERROR type, got %s", result.Type())
		}
	})
}

// Helper function for string contains check
func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestCaseInsensitiveTypeNames tests that type names are case-insensitive
// DWScript (like Pascal) is case-insensitive for all identifiers including type names
func TestCaseInsensitiveTypeNames(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{
			name: "lowercase integer in class field",
			input: `
				type TMyClass = class
					FValue: integer;
				end;
				var obj := TMyClass.Create();
				obj.FValue := 42;
			`,
			shouldPass: true,
		},
		{
			name: "uppercase INTEGER in class field",
			input: `
				type TMyClass = class
					FValue: INTEGER;
				end;
				var obj := TMyClass.Create();
				obj.FValue := 99;
			`,
			shouldPass: true,
		},
		{
			name: "mixed case InTeGeR in class field",
			input: `
				type TMyClass = class
					FValue: InTeGeR;
				end;
				var obj := TMyClass.Create();
				obj.FValue := 123;
			`,
			shouldPass: true,
		},
		{
			name: "lowercase string in class field",
			input: `
				type TMyClass = class
					FName: string;
				end;
				var obj := TMyClass.Create();
				obj.FName := 'test';
			`,
			shouldPass: true,
		},
		{
			name: "lowercase boolean in class field",
			input: `
				type TMyClass = class
					FFlag: boolean;
				end;
				var obj := TMyClass.Create();
				obj.FFlag := true;
			`,
			shouldPass: true,
		},
		{
			name: "lowercase float in class field",
			input: `
				type TMyClass = class
					FValue: float;
				end;
				var obj := TMyClass.Create();
				obj.FValue := 3.14;
			`,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			interp := New(nil)
			result := interp.Eval(program)

			if tt.shouldPass {
				if result != nil && result.Type() == "ERROR" {
					t.Errorf("Unexpected error: %s", result.String())
				}
			} else {
				if result == nil || result.Type() != "ERROR" {
					t.Errorf("Expected error but got none")
				}
			}
		})
	}
}

// TestClassMetadataPopulation tests that ClassMetadata is populated during class declaration.
func TestClassMetadataPopulation(t *testing.T) {
	input := `
		type
			TPoint = class
			private
				FX: Integer;
				FY: Integer;
			public
				constructor Create(X, Y: Integer);
				function GetX(): Integer;
				function GetY(): Integer;
				procedure SetX(Value: Integer);
				destructor Destroy;
			end;
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("Expected no errors, got %v", result)
	}

	// Lookup the class
	classInfo := interp.classes["tpoint"]
	if classInfo == nil {
		t.Fatal("Expected class TPoint to be registered")
	}

	// Check that Metadata is initialized
	if classInfo.Metadata == nil {
		t.Fatal("Expected Metadata to be initialized")
	}

	// Check metadata name
	if classInfo.Metadata.Name != "TPoint" {
		t.Errorf("Expected metadata name 'TPoint', got %q", classInfo.Metadata.Name)
	}

	// Check fields are in metadata
	if len(classInfo.Metadata.Fields) != 2 {
		t.Errorf("Expected 2 fields in metadata, got %d", len(classInfo.Metadata.Fields))
	}

	// Check specific fields
	if _, exists := classInfo.Metadata.Fields["fx"]; !exists {
		t.Error("Expected field 'FX' in metadata")
	}
	if _, exists := classInfo.Metadata.Fields["fy"]; !exists {
		t.Error("Expected field 'FY' in metadata")
	}

	// Check methods are in metadata (excluding constructor and destructor)
	// GetX, GetY, SetX should be in Methods
	if len(classInfo.Metadata.Methods) != 3 {
		t.Errorf("Expected 3 methods in metadata, got %d", len(classInfo.Metadata.Methods))
	}

	// Check constructor is in metadata
	if len(classInfo.Metadata.Constructors) != 1 {
		t.Errorf("Expected 1 constructor in metadata, got %d", len(classInfo.Metadata.Constructors))
	}

	// Check destructor is in metadata
	if classInfo.Metadata.Destructor == nil {
		t.Error("Expected destructor in metadata")
	}

	// Check that methods are registered in MethodRegistry and have IDs
	for methodName, method := range classInfo.Metadata.Methods {
		if method.ID == 0 {
			t.Errorf("Method %q has invalid ID (0)", methodName)
		}
		// Verify we can retrieve the method by ID
		retrieved := interp.methodRegistry.GetMethod(method.ID)
		if retrieved == nil {
			t.Errorf("Could not retrieve method %q by ID %d from registry", methodName, method.ID)
		}
		if retrieved.Name != method.Name {
			t.Errorf("Retrieved method name mismatch: expected %q, got %q", method.Name, retrieved.Name)
		}
	}
}

// TestClassMetadataInheritance tests that metadata is properly set for inheritance.
func TestClassMetadataInheritance(t *testing.T) {
	input := `
		type
			TBase = class
				FValue: Integer;
				procedure DoSomething;
			end;

			TDerived = class(TBase)
				FExtra: String;
				procedure DoExtra;
			end;
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("Expected no errors, got %v", result)
	}

	// Check TDerived metadata
	derived := interp.classes["tderived"]
	if derived == nil {
		t.Fatal("Expected class TDerived to be registered")
	}

	// Check parent metadata reference
	if derived.Metadata.Parent == nil {
		t.Fatal("Expected TDerived metadata to have Parent set")
	}

	if derived.Metadata.ParentName != "TBase" {
		t.Errorf("Expected parent name 'TBase', got %q", derived.Metadata.ParentName)
	}

	// The parent metadata should be the base class's metadata
	base := interp.classes["tbase"]
	if base == nil {
		t.Fatal("Expected class TBase to be registered")
	}

	if derived.Metadata.Parent != base.Metadata {
		t.Error("Expected TDerived metadata parent to reference TBase metadata")
	}
}

// TestClassMetadataFlags tests that class flags are set in metadata.
func TestClassMetadataFlags(t *testing.T) {
	input := `
		type
			TAbstract = class abstract
				procedure AbstractMethod; virtual; abstract;
			end;

			TPartial = class partial
				FField: Integer;
			end;
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("Expected no errors, got %v", result)
	}

	// Check abstract flag
	abstract := interp.classes["tabstract"]
	if abstract == nil {
		t.Fatal("Expected class TAbstract to be registered")
	}

	if !abstract.Metadata.IsAbstract {
		t.Error("Expected TAbstract metadata IsAbstract to be true")
	}

	// Check partial flag
	partial := interp.classes["tpartial"]
	if partial == nil {
		t.Fatal("Expected class TPartial to be registered")
	}

	if !partial.Metadata.IsPartial {
		t.Error("Expected TPartial metadata IsPartial to be true")
	}
}
