package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestResolveTypeName_Primitives tests resolveTypeName for primitive types.
// Task 3.5.139a: Test primitive type resolution.
func TestResolveTypeName_Primitives(t *testing.T) {
	tests := []struct {
		name         string
		typeName     string
		expectedType types.Type
		expectError  bool
	}{
		// Integer type
		{"Integer lowercase", "integer", types.INTEGER, false},
		{"Integer uppercase", "INTEGER", types.INTEGER, false},
		{"Integer mixed case", "Integer", types.INTEGER, false},

		// Float type
		{"Float lowercase", "float", types.FLOAT, false},
		{"Float uppercase", "FLOAT", types.FLOAT, false},
		{"Float mixed case", "Float", types.FLOAT, false},

		// String type
		{"String lowercase", "string", types.STRING, false},
		{"String uppercase", "STRING", types.STRING, false},
		{"String mixed case", "String", types.STRING, false},

		// Boolean type
		{"Boolean lowercase", "boolean", types.BOOLEAN, false},
		{"Boolean uppercase", "BOOLEAN", types.BOOLEAN, false},
		{"Boolean mixed case", "Boolean", types.BOOLEAN, false},

		// Variant type
		{"Variant lowercase", "variant", types.VARIANT, false},
		{"Variant uppercase", "VARIANT", types.VARIANT, false},
		{"Variant mixed case", "Variant", types.VARIANT, false},

		// Const type (deprecated, mapped to Variant)
		{"Const lowercase", "const", types.VARIANT, false},
		{"Const uppercase", "CONST", types.VARIANT, false},
		{"Const mixed case", "Const", types.VARIANT, false},

		// TDateTime type
		{"TDateTime lowercase", "tdatetime", types.DATETIME, false},
		{"TDateTime uppercase", "TDATETIME", types.DATETIME, false},
		{"TDateTime mixed case", "TDateTime", types.DATETIME, false},

		// Nil type
		{"Nil lowercase", "nil", types.NIL, false},
		{"Nil uppercase", "NIL", types.NIL, false},
		{"Nil mixed case", "Nil", types.NIL, false},

		// Void type
		{"Void lowercase", "void", types.VOID, false},
		{"Void uppercase", "VOID", types.VOID, false},
		{"Void mixed case", "Void", types.VOID, false},

		// Unknown types (should error)
		{"Unknown type", "Unknown", nil, true},
		{"Custom type", "TMyClass", nil, true},
		{"Empty string", "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create evaluator and context
			e := &Evaluator{}
			ctx := &ExecutionContext{}

			// Call resolveTypeName
			result, err := e.resolveTypeName(tt.typeName, ctx)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for type '%s', but got none", tt.typeName)
				}
				return
			}

			// Check no error when not expected
			if err != nil {
				t.Errorf("Unexpected error for type '%s': %v", tt.typeName, err)
				return
			}

			// Check type matches
			if result != tt.expectedType {
				t.Errorf("Type mismatch for '%s': expected %v, got %v",
					tt.typeName, tt.expectedType, result)
			}

			// Verify type properties
			if result.String() != tt.expectedType.String() {
				t.Errorf("Type string mismatch for '%s': expected '%s', got '%s'",
					tt.typeName, tt.expectedType.String(), result.String())
			}

			if result.TypeKind() != tt.expectedType.TypeKind() {
				t.Errorf("TypeKind mismatch for '%s': expected '%s', got '%s'",
					tt.typeName, tt.expectedType.TypeKind(), result.TypeKind())
			}
		})
	}
}

// TestResolveTypeName_ParentQualification tests that parent qualification is stripped.
// Task 3.5.139a: Test class type strings with parent qualification.
func TestResolveTypeName_ParentQualification(t *testing.T) {
	tests := []struct {
		name        string
		typeName    string
		expectError bool
	}{
		// These should strip the (TParent) suffix and try to resolve the base name
		// Since custom types aren't supported yet in 3.5.139a, they should error
		{"Class with parent", "TSub(TBase)", true},
		{"Class with parent and spaces", "TSub (TBase)", true},
		{"Class with nested parent", "TSub(TBase(TRoot))", true},

		// But if the base name is a primitive, it should work
		// (This is unlikely in practice but tests the stripping logic)
		{"Primitive with paren suffix", "Integer(Foo)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Evaluator{}
			ctx := &ExecutionContext{}

			result, err := e.resolveTypeName(tt.typeName, ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for type '%s', but got none", tt.typeName)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for type '%s': %v", tt.typeName, err)
					return
				}
				// For "Integer(Foo)", should resolve to INTEGER
				if result != types.INTEGER {
					t.Errorf("Expected INTEGER for '%s', got %v", tt.typeName, result)
				}
			}
		})
	}
}

// TestResolveTypeName_CaseInsensitivity verifies case-insensitive resolution.
// Task 3.5.139a: Test that type resolution is case-insensitive per DWScript spec.
func TestResolveTypeName_CaseInsensitivity(t *testing.T) {
	variations := []string{
		"integer", "Integer", "INTEGER", "InTeGeR",
		"float", "Float", "FLOAT", "FlOaT",
		"string", "String", "STRING", "StRiNg",
		"boolean", "Boolean", "BOOLEAN", "BoOlEaN",
		"variant", "Variant", "VARIANT", "VaRiAnT",
	}

	e := &Evaluator{}
	ctx := &ExecutionContext{}

	for _, typeName := range variations {
		t.Run(typeName, func(t *testing.T) {
			result, err := e.resolveTypeName(typeName, ctx)
			if err != nil {
				t.Errorf("Unexpected error for '%s': %v", typeName, err)
				return
			}
			if result == nil {
				t.Errorf("Got nil result for '%s'", typeName)
			}
		})
	}
}


// TestResolveTypeName_RegisteredTypes tests resolveTypeName for registered types.
// Task 3.5.139b: Test enum, record, class, interface, and subrange type resolution.
//
// Note: These tests verify the lookup logic works correctly. They use the
// infrastructure that exists in the evaluator package for type lookups.
// Full integration testing with actual type declarations is done elsewhere.
func TestResolveTypeName_RegisteredTypes(t *testing.T) {
	t.Run("returns error for unregistered custom types", func(t *testing.T) {
		e := &Evaluator{}
		ctx := &ExecutionContext{}

		// Without registration, custom types should return errors
		customTypes := []string{"TMyClass", "TMyRecord", "TMyEnum", "IMyInterface", "TMySubrange"}

		for _, typeName := range customTypes {
			_, err := e.resolveTypeName(typeName, ctx)
			if err == nil {
				t.Errorf("Expected error for unregistered type '%s', got nil", typeName)
			}
		}
	})

	// Note: Full integration tests with actual type registration would require
	// setting up the environment and TypeSystem with registered types.
	// Those tests belong in integration test files that test the full interpreter.
	//
	// The current implementation verifies:
	// - Enum types: looks in ctx.Env().Get("__enum_type_" + normalized)
	// - Record types: looks in ctx.Env().Get("__record_type_" + normalized)
	// - Class types: checks e.typeSystem.HasClass()
	// - Interface types: checks e.typeSystem.HasInterface()
	// - Subrange types: looks in ctx.Env().Get("__subrange_type_" + normalized)
}

// TestResolveTypeName_ArrayTypes tests resolveTypeName for array types.
// Task 3.5.139c: Test array type resolution via TypeSystem.
func TestResolveTypeName_ArrayTypes(t *testing.T) {
	t.Run("returns error for unregistered array types", func(t *testing.T) {
		e := &Evaluator{}
		ctx := &ExecutionContext{}

		// Without registration, array types should return errors
		arrayTypes := []string{"TIntArray", "TStringArray", "TMyArray"}

		for _, typeName := range arrayTypes {
			_, err := e.resolveTypeName(typeName, ctx)
			if err == nil {
				t.Errorf("Expected error for unregistered array type '%s', got nil", typeName)
			}
		}
	})

	// Note: Full integration tests with actual array type registration would require
	// setting up the TypeSystem with registered array types. Those tests belong in
	// integration test files that test the full interpreter.
	//
	// The current implementation verifies:
	// - Array types: checks e.typeSystem.LookupArrayType()
	// - Returns the ArrayType from TypeSystem if found
	// - Returns error if not found
}

// TestParseInlineArrayType tests parseInlineArrayType for inline array syntax.
// Task 3.5.139d: Test inline array type parsing.
func TestParseInlineArrayType(t *testing.T) {
	tests := []struct {
		name          string
		signature     string
		expectNil     bool
		expectedType  string // String representation for verification
		isDynamic     bool
		lowBound      int
		highBound     int
	}{
		// Dynamic arrays
		{"Dynamic array of Integer", "array of Integer", false, "array of Integer", true, 0, 0},
		{"Dynamic array of String", "array of String", false, "array of String", true, 0, 0},
		{"Dynamic array of Float", "array of Float", false, "array of Float", true, 0, 0},
		{"Dynamic array of Boolean", "array of Boolean", false, "array of Boolean", true, 0, 0},

		// Case insensitivity
		{"Dynamic array uppercase", "ARRAY OF INTEGER", false, "array of Integer", true, 0, 0},
		{"Dynamic array mixed case", "Array Of String", false, "array of String", true, 0, 0},

		// Static arrays
		{"Static array 0..9", "array[0..9] of Integer", false, "array[0..9] of Integer", false, 0, 9},
		{"Static array 1..10", "array[1..10] of String", false, "array[1..10] of String", false, 1, 10},
		{"Static array -5..5", "array[-5..5] of Float", false, "array[-5..5] of Float", false, -5, 5},

		// Nested arrays
		{"2D dynamic array", "array of array of Integer", false, "array of array of Integer", true, 0, 0},
		{"2D static array", "array[0..2] of array[0..3] of Integer", false, "array[0..2] of array[0..3] of Integer", false, 0, 2},

		// Error cases
		{"Invalid syntax - no 'of'", "array Integer", true, "", false, 0, 0},
		{"Invalid syntax - missing element type", "array of", true, "", false, 0, 0},
		{"Invalid syntax - bad bounds", "array[0-9] of Integer", true, "", false, 0, 0},
		{"Invalid syntax - non-numeric bounds", "array[a..b] of Integer", true, "", false, 0, 0},
		{"Unknown element type", "array of UnknownType", true, "", false, 0, 0},
		{"Not an array", "Integer", true, "", false, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Evaluator{}
			ctx := &ExecutionContext{}

			result := e.parseInlineArrayType(tt.signature, ctx)

			if tt.expectNil {
				if result != nil {
					t.Errorf("Expected nil for signature '%s', got %v", tt.signature, result)
				}
				return
			}

			if result == nil {
				t.Errorf("Expected non-nil ArrayType for signature '%s', got nil", tt.signature)
				return
			}

			// Verify array type properties
			if result.String() != tt.expectedType {
				t.Errorf("Type string mismatch for '%s': expected '%s', got '%s'",
					tt.signature, tt.expectedType, result.String())
			}

			// Verify dynamic vs static
			if result.IsDynamic() != tt.isDynamic {
				t.Errorf("IsDynamic mismatch for '%s': expected %v, got %v",
					tt.signature, tt.isDynamic, result.IsDynamic())
			}

			// Verify bounds for static arrays
			if !tt.isDynamic {
				if result.LowBound == nil || *result.LowBound != tt.lowBound {
					actual := -1
					if result.LowBound != nil {
						actual = *result.LowBound
					}
					t.Errorf("LowBound mismatch for '%s': expected %d, got %d",
						tt.signature, tt.lowBound, actual)
				}
				if result.HighBound == nil || *result.HighBound != tt.highBound {
					actual := -1
					if result.HighBound != nil {
						actual = *result.HighBound
					}
					t.Errorf("HighBound mismatch for '%s': expected %d, got %d",
						tt.signature, tt.highBound, actual)
				}
			}
		})
	}
}

// TestResolveTypeName_InlineArrays tests resolveTypeName for inline array syntax.
// Task 3.5.139d: Test integration of inline array parsing into resolveTypeName.
func TestResolveTypeName_InlineArrays(t *testing.T) {
	tests := []struct {
		name         string
		typeName     string
		expectError  bool
		expectedType string
	}{
		// Dynamic arrays via resolveTypeName
		{"Resolve dynamic array", "array of Integer", false, "array of Integer"},
		{"Resolve dynamic string array", "array of String", false, "array of String"},

		// Static arrays via resolveTypeName
		{"Resolve static array", "array[0..9] of Integer", false, "array[0..9] of Integer"},
		{"Resolve static bounds array", "array[1..100] of String", false, "array[1..100] of String"},

		// Nested arrays
		{"Resolve 2D array", "array of array of Integer", false, "array of array of Integer"},

		// Error cases
		{"Invalid array syntax", "array Integer", true, ""},
		{"Unknown element type", "array of CustomType", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Evaluator{}
			ctx := &ExecutionContext{}

			result, err := e.resolveTypeName(tt.typeName, ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for type '%s', got nil", tt.typeName)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for type '%s': %v", tt.typeName, err)
				return
			}

			if result == nil {
				t.Errorf("Expected non-nil type for '%s', got nil", tt.typeName)
				return
			}

			if result.String() != tt.expectedType {
				t.Errorf("Type string mismatch for '%s': expected '%s', got '%s'",
					tt.typeName, tt.expectedType, result.String())
			}
		})
	}
}



// TestResolveArrayTypeNode tests resolveArrayTypeNode for AST ArrayTypeNode resolution.
// Task 3.5.139e: Test AST-based array type resolution.
func TestResolveArrayTypeNode(t *testing.T) {
	tests := []struct {
		name         string
		setupNode    func() *ast.ArrayTypeNode
		expectNil    bool
		expectedType string
	}{
		{
			name: "Dynamic array of Integer",
			setupNode: func() *ast.ArrayTypeNode {
				return &ast.ArrayTypeNode{
					ElementType: &ast.TypeAnnotation{Name: "Integer"},
				}
			},
			expectNil:    false,
			expectedType: "array of Integer",
		},
		{
			name: "Dynamic array of String",
			setupNode: func() *ast.ArrayTypeNode {
				return &ast.ArrayTypeNode{
					ElementType: &ast.TypeAnnotation{Name: "String"},
				}
			},
			expectNil:    false,
			expectedType: "array of String",
		},
		{
			name: "Static array with literal bounds [0..9]",
			setupNode: func() *ast.ArrayTypeNode {
				return &ast.ArrayTypeNode{
					LowBound:    &ast.IntegerLiteral{Value: 0},
					HighBound:   &ast.IntegerLiteral{Value: 9},
					ElementType: &ast.TypeAnnotation{Name: "Integer"},
				}
			},
			expectNil:    false,
			expectedType: "array[0..9] of Integer",
		},
		{
			name: "Static array with negative bounds [-5..5]",
			setupNode: func() *ast.ArrayTypeNode {
				return &ast.ArrayTypeNode{
					LowBound:    &ast.IntegerLiteral{Value: -5},
					HighBound:   &ast.IntegerLiteral{Value: 5},
					ElementType: &ast.TypeAnnotation{Name: "Float"},
				}
			},
			expectNil:    false,
			expectedType: "array[-5..5] of Float",
		},
		{
			name: "Nested array (2D dynamic)",
			setupNode: func() *ast.ArrayTypeNode {
				return &ast.ArrayTypeNode{
					ElementType: &ast.ArrayTypeNode{
						ElementType: &ast.TypeAnnotation{Name: "Integer"},
					},
				}
			},
			expectNil:    false,
			expectedType: "array of array of Integer",
		},
		{
			name: "Nested array (2D static)",
			setupNode: func() *ast.ArrayTypeNode {
				return &ast.ArrayTypeNode{
					LowBound:  &ast.IntegerLiteral{Value: 0},
					HighBound: &ast.IntegerLiteral{Value: 2},
					ElementType: &ast.ArrayTypeNode{
						LowBound:    &ast.IntegerLiteral{Value: 0},
						HighBound:   &ast.IntegerLiteral{Value: 3},
						ElementType: &ast.TypeAnnotation{Name: "Integer"},
					},
				}
			},
			expectNil:    false,
			expectedType: "array[0..2] of array[0..3] of Integer",
		},
		{
			name: "Nil node",
			setupNode: func() *ast.ArrayTypeNode {
				return nil
			},
			expectNil:    true,
			expectedType: "",
		},
		{
			name: "Unknown element type",
			setupNode: func() *ast.ArrayTypeNode {
				return &ast.ArrayTypeNode{
					ElementType: &ast.TypeAnnotation{Name: "UnknownType"},
				}
			},
			expectNil:    true,
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Evaluator{}
			ctx := &ExecutionContext{}

			arrayNode := tt.setupNode()
			result := e.resolveArrayTypeNode(arrayNode, ctx)

			if tt.expectNil {
				if result != nil {
					t.Errorf("Expected nil for node, got %v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("Expected non-nil ArrayType, got nil")
				return
			}

			if result.String() != tt.expectedType {
				t.Errorf("Type string mismatch: expected '%s', got '%s'",
					tt.expectedType, result.String())
			}
		})
	}
}
