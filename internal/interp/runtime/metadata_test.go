package runtime

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestMethodVisibilityString tests the String() method of MethodVisibility.
func TestMethodVisibilityString(t *testing.T) {
	tests := []struct {
		expected   string
		visibility MethodVisibility
	}{
		{visibility: VisibilityPublic, expected: "public"},
		{visibility: VisibilityPrivate, expected: "private"},
		{visibility: VisibilityProtected, expected: "protected"},
		{visibility: VisibilityPublished, expected: "published"},
		{visibility: MethodVisibility(999), expected: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.visibility.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFieldVisibilityString tests the String() method of FieldVisibility.
func TestFieldVisibilityString(t *testing.T) {
	tests := []struct {
		expected   string
		visibility FieldVisibility
	}{
		{visibility: FieldVisibilityPublic, expected: "public"},
		{visibility: FieldVisibilityPrivate, expected: "private"},
		{visibility: FieldVisibilityProtected, expected: "protected"},
		{visibility: FieldVisibilityPublished, expected: "published"},
		{visibility: FieldVisibility(999), expected: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.visibility.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestMethodMetadata_IsFunction tests the IsFunction method.
func TestMethodMetadata_IsFunction(t *testing.T) {
	tests := []struct {
		name           string
		returnTypeName string
		expected       bool
	}{
		{"function with return type", "Integer", true},
		{"procedure without return type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MethodMetadata{
				ReturnTypeName: tt.returnTypeName,
			}
			result := m.IsFunction()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestMethodMetadata_IsProcedure tests the IsProcedure method.
func TestMethodMetadata_IsProcedure(t *testing.T) {
	tests := []struct {
		name           string
		returnTypeName string
		expected       bool
	}{
		{"procedure without return type", "", true},
		{"function with return type", "Integer", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MethodMetadata{
				ReturnTypeName: tt.returnTypeName,
			}
			result := m.IsProcedure()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestMethodMetadata_RequiredParamCount tests the RequiredParamCount method.
func TestMethodMetadata_RequiredParamCount(t *testing.T) {
	tests := []struct {
		name       string
		parameters []ParameterMetadata
		expected   int
	}{
		{
			name:       "no parameters",
			parameters: []ParameterMetadata{},
			expected:   0,
		},
		{
			name: "all required parameters",
			parameters: []ParameterMetadata{
				{Name: "a", DefaultValue: nil},
				{Name: "b", DefaultValue: nil},
			},
			expected: 2,
		},
		{
			name: "mixed required and optional",
			parameters: []ParameterMetadata{
				{Name: "a", DefaultValue: nil},
				{Name: "b", DefaultValue: &ast.IntegerLiteral{Value: 42}},
				{Name: "c", DefaultValue: nil},
			},
			expected: 2,
		},
		{
			name: "all optional parameters",
			parameters: []ParameterMetadata{
				{Name: "a", DefaultValue: &ast.IntegerLiteral{Value: 1}},
				{Name: "b", DefaultValue: &ast.IntegerLiteral{Value: 2}},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MethodMetadata{
				Parameters: tt.parameters,
			}
			result := m.RequiredParamCount()
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestMethodMetadata_ParamCount tests the ParamCount method.
func TestMethodMetadata_ParamCount(t *testing.T) {
	tests := []struct {
		name       string
		parameters []ParameterMetadata
		expected   int
	}{
		{
			name:       "no parameters",
			parameters: []ParameterMetadata{},
			expected:   0,
		},
		{
			name: "three parameters",
			parameters: []ParameterMetadata{
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MethodMetadata{
				Parameters: tt.parameters,
			}
			result := m.ParamCount()
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestNewClassMetadata tests ClassMetadata creation and initialization.
func TestNewClassMetadata(t *testing.T) {
	metadata := NewClassMetadata("TMyClass")

	if metadata.Name != "TMyClass" {
		t.Errorf("Expected name 'TMyClass', got %q", metadata.Name)
	}

	// Check all maps are initialized
	if metadata.Fields == nil {
		t.Error("Fields map not initialized")
	}
	if metadata.Methods == nil {
		t.Error("Methods map not initialized")
	}
	if metadata.MethodOverloads == nil {
		t.Error("MethodOverloads map not initialized")
	}
	if metadata.ClassMethods == nil {
		t.Error("ClassMethods map not initialized")
	}
	if metadata.ClassMethodOverloads == nil {
		t.Error("ClassMethodOverloads map not initialized")
	}
	if metadata.Constructors == nil {
		t.Error("Constructors map not initialized")
	}
	if metadata.ConstructorOverloads == nil {
		t.Error("ConstructorOverloads map not initialized")
	}
	if metadata.VirtualMethods == nil {
		t.Error("VirtualMethods map not initialized")
	}
	if metadata.Constants == nil {
		t.Error("Constants map not initialized")
	}
	if metadata.ClassVars == nil {
		t.Error("ClassVars map not initialized")
	}
	if metadata.Properties == nil {
		t.Error("Properties map not initialized")
	}
}

// TestNewRecordMetadata tests RecordMetadata creation and initialization.
func TestNewRecordMetadata(t *testing.T) {
	recordType := &types.RecordType{Name: "TPoint"}
	metadata := NewRecordMetadata("TPoint", recordType)

	if metadata.Name != "TPoint" {
		t.Errorf("Expected name 'TPoint', got %q", metadata.Name)
	}

	if metadata.RecordType != recordType {
		t.Error("RecordType not set correctly")
	}

	// Check all maps are initialized
	if metadata.Fields == nil {
		t.Error("Fields map not initialized")
	}
	if metadata.Methods == nil {
		t.Error("Methods map not initialized")
	}
	if metadata.MethodOverloads == nil {
		t.Error("MethodOverloads map not initialized")
	}
	if metadata.StaticMethods == nil {
		t.Error("StaticMethods map not initialized")
	}
	if metadata.StaticMethodOverloads == nil {
		t.Error("StaticMethodOverloads map not initialized")
	}
	if metadata.Constants == nil {
		t.Error("Constants map not initialized")
	}
	if metadata.ClassVars == nil {
		t.Error("ClassVars map not initialized")
	}
}

// TestNewHelperMetadata tests HelperMetadata creation and initialization.
func TestNewHelperMetadata(t *testing.T) {
	targetType := &types.IntegerType{}
	metadata := NewHelperMetadata("TIntegerHelper", targetType, false)

	if metadata.Name != "TIntegerHelper" {
		t.Errorf("Expected name 'TIntegerHelper', got %q", metadata.Name)
	}

	if metadata.TargetType != targetType {
		t.Error("TargetType not set correctly")
	}

	if metadata.IsRecordHelper {
		t.Error("Expected IsRecordHelper to be false")
	}

	// Check all maps are initialized
	if metadata.Methods == nil {
		t.Error("Methods map not initialized")
	}
	if metadata.Properties == nil {
		t.Error("Properties map not initialized")
	}
	if metadata.ClassVars == nil {
		t.Error("ClassVars map not initialized")
	}
	if metadata.ClassConsts == nil {
		t.Error("ClassConsts map not initialized")
	}
	if metadata.BuiltinMethods == nil {
		t.Error("BuiltinMethods map not initialized")
	}
}

// TestMethodMetadataFromAST tests conversion from AST FunctionDecl to MethodMetadata.
func TestMethodMetadataFromAST(t *testing.T) {
	// Create a sample AST function declaration
	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "MyFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "x"},
				Type: &ast.TypeAnnotation{
					Name: "Integer",
				},
				ByRef: false,
			},
			{
				Name: &ast.Identifier{Value: "y"},
				Type: &ast.TypeAnnotation{
					Name: "String",
				},
				ByRef:        false,
				DefaultValue: &ast.StringLiteral{Value: "default"},
			},
		},
		ReturnType: &ast.TypeAnnotation{
			Name: "Boolean",
		},
		Body:       &ast.BlockStatement{Statements: []ast.Statement{}},
		IsVirtual:  true,
		IsOverride: false,
	}

	metadata := MethodMetadataFromAST(fn)

	if metadata == nil {
		t.Fatal("Expected non-nil metadata")
	}

	if metadata.Name != "MyFunc" {
		t.Errorf("Expected name 'MyFunc', got %q", metadata.Name)
	}

	if len(metadata.Parameters) != 2 {
		t.Fatalf("Expected 2 parameters, got %d", len(metadata.Parameters))
	}

	// Check first parameter
	if metadata.Parameters[0].Name != "x" {
		t.Errorf("Expected parameter name 'x', got %q", metadata.Parameters[0].Name)
	}
	if metadata.Parameters[0].TypeName != "Integer" {
		t.Errorf("Expected parameter type 'Integer', got %q", metadata.Parameters[0].TypeName)
	}
	if metadata.Parameters[0].ByRef {
		t.Error("Expected ByRef to be false")
	}
	if metadata.Parameters[0].DefaultValue != nil {
		t.Error("Expected no default value for first parameter")
	}

	// Check second parameter
	if metadata.Parameters[1].Name != "y" {
		t.Errorf("Expected parameter name 'y', got %q", metadata.Parameters[1].Name)
	}
	if metadata.Parameters[1].DefaultValue == nil {
		t.Error("Expected default value for second parameter")
	}

	// Check return type
	if metadata.ReturnTypeName != "Boolean" {
		t.Errorf("Expected return type 'Boolean', got %q", metadata.ReturnTypeName)
	}

	if metadata.Body == nil {
		t.Error("Expected non-nil body")
	}

	// Check flags
	if !metadata.IsVirtual {
		t.Error("Expected IsVirtual to be true")
	}
	if metadata.IsOverride {
		t.Error("Expected IsOverride to be false")
	}
}

// TestMethodMetadataFromAST_Nil tests handling of nil input.
func TestMethodMetadataFromAST_Nil(t *testing.T) {
	metadata := MethodMetadataFromAST(nil)
	if metadata != nil {
		t.Error("Expected nil metadata for nil input")
	}
}

// TestFieldMetadataFromAST tests conversion from AST FieldDecl to FieldMetadata.
func TestFieldMetadataFromAST(t *testing.T) {
	field := &ast.FieldDecl{
		Name: &ast.Identifier{Value: "FMyField"},
		Type: &ast.TypeAnnotation{
			Name: "Integer",
		},
		InitValue: &ast.IntegerLiteral{Value: 42},
	}

	metadata := FieldMetadataFromAST(field)

	if metadata == nil {
		t.Fatal("Expected non-nil metadata")
	}

	if metadata.Name != "FMyField" {
		t.Errorf("Expected name 'FMyField', got %q", metadata.Name)
	}

	if metadata.TypeName != "Integer" {
		t.Errorf("Expected type 'Integer', got %q", metadata.TypeName)
	}

	if metadata.InitValue == nil {
		t.Error("Expected non-nil init value")
	}

	if metadata.Visibility != FieldVisibilityPublic {
		t.Errorf("Expected public visibility, got %v", metadata.Visibility)
	}
}

// TestFieldMetadataFromAST_Nil tests handling of nil input.
func TestFieldMetadataFromAST_Nil(t *testing.T) {
	metadata := FieldMetadataFromAST(nil)
	if metadata != nil {
		t.Error("Expected nil metadata for nil input")
	}
}

// TestAddMethodToClass tests adding methods to ClassMetadata.
func TestAddMethodToClass(t *testing.T) {
	class := NewClassMetadata("TMyClass")

	// Add first method
	method1 := &MethodMetadata{Name: "DoSomething", Parameters: []ParameterMetadata{}}
	AddMethodToClass(class, method1, false)

	if class.Methods["dosomething"] != method1 {
		t.Error("Method not added correctly")
	}

	if len(class.MethodOverloads["dosomething"]) != 0 {
		t.Error("No overloads should exist yet")
	}

	// Add overload
	method2 := &MethodMetadata{
		Name:       "DoSomething",
		Parameters: []ParameterMetadata{{Name: "x"}},
	}
	AddMethodToClass(class, method2, false)

	if len(class.MethodOverloads["dosomething"]) != 2 {
		t.Errorf("Expected 2 overloads, got %d", len(class.MethodOverloads["dosomething"]))
	}

	if class.MethodOverloads["dosomething"][0] != method1 {
		t.Error("First overload should be method1")
	}

	if class.MethodOverloads["dosomething"][1] != method2 {
		t.Error("Second overload should be method2")
	}
}

// TestAddConstructorToClass tests adding constructors to ClassMetadata.
func TestAddConstructorToClass(t *testing.T) {
	class := NewClassMetadata("TMyClass")

	// Add first constructor
	ctor1 := &MethodMetadata{Name: "Create", Parameters: []ParameterMetadata{}}
	AddConstructorToClass(class, ctor1)

	if !ctor1.IsConstructor {
		t.Error("Constructor flag not set")
	}

	if class.Constructors["create"] != ctor1 {
		t.Error("Constructor not added correctly")
	}

	if class.DefaultConstructor != "Create" {
		t.Errorf("Expected default constructor 'Create', got %q", class.DefaultConstructor)
	}

	// Add overload
	ctor2 := &MethodMetadata{
		Name:       "Create",
		Parameters: []ParameterMetadata{{Name: "x"}},
	}
	AddConstructorToClass(class, ctor2)

	if len(class.ConstructorOverloads["create"]) != 2 {
		t.Errorf("Expected 2 constructor overloads, got %d", len(class.ConstructorOverloads["create"]))
	}
}

// TestAddFieldToClass tests adding fields to ClassMetadata.
func TestAddFieldToClass(t *testing.T) {
	class := NewClassMetadata("TMyClass")

	field := &FieldMetadata{Name: "FMyField", TypeName: "Integer"}
	AddFieldToClass(class, field)

	if class.Fields["fmyfield"] != field {
		t.Error("Field not added correctly")
	}
}

// Note: normalizeIdentifier test removed - function replaced by ident.Normalize
// which has its own tests in pkg/ident/
