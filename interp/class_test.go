package interp

import (
	"testing"

	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
	"github.com/cwbudde/go-dws/types"
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
		Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "GetValue"},
			Value: "GetValue",
		},
		Parameters: []*ast.Parameter{},
		ReturnType: &ast.TypeAnnotation{Name: "Integer"},
		Body: &ast.BlockStatement{
			Statements: []ast.Statement{},
		},
	}

	classInfo.Methods["GetValue"] = method

	if len(classInfo.Methods) != 1 {
		t.Errorf("len(classInfo.Methods) = %d, want 1", len(classInfo.Methods))
	}

	if classInfo.Methods["GetValue"] == nil {
		t.Error("Method 'GetValue' should be registered")
	}

	if classInfo.Methods["GetValue"].Name.Value != "GetValue" {
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
		Name: &ast.Identifier{Value: "GetValue"},
	}
	classInfo.Methods["GetValue"] = method

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
		Name: &ast.Identifier{Value: "ToString"},
	}
	parent.Methods["ToString"] = parentMethod

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
		Name: &ast.Identifier{Value: "ToString"},
		Body: &ast.BlockStatement{}, // Different body
	}
	parent.Methods["ToString"] = parentMethod

	// Create child class that overrides the method
	child := NewClassInfo("TPerson")
	child.Parent = parent

	childMethod := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "ToString"},
		Body: &ast.BlockStatement{
			Statements: []ast.Statement{
				// Different implementation
			},
		},
	}
	child.Methods["ToString"] = childMethod

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
// Task 7.141: External Class Runtime Tests
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
