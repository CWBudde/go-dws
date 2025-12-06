package evaluator

import (
	"testing"

	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestResolveTypeName_Primitives tests resolveTypeName for primitive types.
func TestResolveTypeName_Primitives(t *testing.T) {
	tests := []struct {
		expectedType types.Type
		name         string
		typeName     string
		expectError  bool
	}{
		// Integer type
		{name: "Integer lowercase", typeName: "integer", expectedType: types.INTEGER, expectError: false},
		{name: "Integer uppercase", typeName: "INTEGER", expectedType: types.INTEGER, expectError: false},
		{name: "Integer mixed case", typeName: "Integer", expectedType: types.INTEGER, expectError: false},

		// Float type
		{name: "Float lowercase", typeName: "float", expectedType: types.FLOAT, expectError: false},
		{name: "Float uppercase", typeName: "FLOAT", expectedType: types.FLOAT, expectError: false},
		{name: "Float mixed case", typeName: "Float", expectedType: types.FLOAT, expectError: false},

		// String type
		{name: "String lowercase", typeName: "string", expectedType: types.STRING, expectError: false},
		{name: "String uppercase", typeName: "STRING", expectedType: types.STRING, expectError: false},
		{name: "String mixed case", typeName: "String", expectedType: types.STRING, expectError: false},

		// Boolean type
		{name: "Boolean lowercase", typeName: "boolean", expectedType: types.BOOLEAN, expectError: false},
		{name: "Boolean uppercase", typeName: "BOOLEAN", expectedType: types.BOOLEAN, expectError: false},
		{name: "Boolean mixed case", typeName: "Boolean", expectedType: types.BOOLEAN, expectError: false},

		// Variant type
		{name: "Variant lowercase", typeName: "variant", expectedType: types.VARIANT, expectError: false},
		{name: "Variant uppercase", typeName: "VARIANT", expectedType: types.VARIANT, expectError: false},
		{name: "Variant mixed case", typeName: "Variant", expectedType: types.VARIANT, expectError: false},

		// Const type (deprecated, mapped to Variant)
		{name: "Const lowercase", typeName: "const", expectedType: types.VARIANT, expectError: false},
		{name: "Const uppercase", typeName: "CONST", expectedType: types.VARIANT, expectError: false},
		{name: "Const mixed case", typeName: "Const", expectedType: types.VARIANT, expectError: false},

		// TDateTime type
		{name: "TDateTime lowercase", typeName: "tdatetime", expectedType: types.DATETIME, expectError: false},
		{name: "TDateTime uppercase", typeName: "TDATETIME", expectedType: types.DATETIME, expectError: false},
		{name: "TDateTime mixed case", typeName: "TDateTime", expectedType: types.DATETIME, expectError: false},

		// Nil type
		{name: "Nil lowercase", typeName: "nil", expectedType: types.NIL, expectError: false},
		{name: "Nil uppercase", typeName: "NIL", expectedType: types.NIL, expectError: false},
		{name: "Nil mixed case", typeName: "Nil", expectedType: types.NIL, expectError: false},

		// Void type
		{name: "Void lowercase", typeName: "void", expectedType: types.VOID, expectError: false},
		{name: "Void uppercase", typeName: "VOID", expectedType: types.VOID, expectError: false},
		{name: "Void mixed case", typeName: "Void", expectedType: types.VOID, expectError: false},

		// Unknown types (should error)
		{name: "Unknown type", typeName: "Unknown", expectedType: nil, expectError: true},
		{name: "Custom type", typeName: "TMyClass", expectedType: nil, expectError: true},
		{name: "Empty string", typeName: "", expectedType: nil, expectError: true},
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

// TestResolveTypeName_FunctionPointerTypes ensures registered function pointer
// types can be resolved through the TypeSystem.
func TestResolveTypeName_FunctionPointerTypes(t *testing.T) {
	typeSystem := interptypes.NewTypeSystem()
	e := &Evaluator{typeSystem: typeSystem}
	ctx := &ExecutionContext{}

	funcPtrType := types.NewProcedurePointerType(nil)
	typeSystem.RegisterFunctionPointerType("TProc", funcPtrType)

	resolved, err := e.resolveTypeName("TProc", ctx)
	if err != nil {
		t.Fatalf("expected no error resolving function pointer type, got %v", err)
	}
	if resolved == nil {
		t.Fatalf("expected non-nil type for function pointer")
	}
	if resolved.TypeKind() != "FUNCTION_POINTER" {
		t.Errorf("expected FUNCTION_POINTER type kind, got %s", resolved.TypeKind())
	}
}

func TestResolveArrayTypeNode_FunctionPointerElement(t *testing.T) {
	typeSystem := interptypes.NewTypeSystem()
	e := &Evaluator{typeSystem: typeSystem}
	ctx := &ExecutionContext{}

	funcPtrType := types.NewProcedurePointerType(nil)
	typeSystem.RegisterFunctionPointerType("TProc", funcPtrType)

	arrayNode := &ast.ArrayTypeNode{
		ElementType: &ast.TypeAnnotation{Name: "TProc"},
	}

	result := e.resolveArrayTypeNode(arrayNode, ctx)
	if result == nil {
		t.Fatal("expected non-nil array type for function pointer element")
	}
	if result.ElementType == nil || result.ElementType.TypeKind() != "FUNCTION_POINTER" {
		t.Fatalf("expected function pointer element type, got %v", result.ElementType)
	}
}

// TestResolveTypeName_ArrayTypes tests resolveTypeName for array types.
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
func TestParseInlineArrayType(t *testing.T) {
	tests := []struct {
		name         string
		signature    string
		expectedType string
		lowBound     int
		highBound    int
		expectNil    bool
		isDynamic    bool
	}{
		// Dynamic arrays
		{name: "Dynamic array of Integer", signature: "array of Integer", expectNil: false, expectedType: "array of Integer", isDynamic: true, lowBound: 0, highBound: 0},
		{name: "Dynamic array of String", signature: "array of String", expectNil: false, expectedType: "array of String", isDynamic: true, lowBound: 0, highBound: 0},
		{name: "Dynamic array of Float", signature: "array of Float", expectNil: false, expectedType: "array of Float", isDynamic: true, lowBound: 0, highBound: 0},
		{name: "Dynamic array of Boolean", signature: "array of Boolean", expectNil: false, expectedType: "array of Boolean", isDynamic: true, lowBound: 0, highBound: 0},

		// Case insensitivity
		{name: "Dynamic array uppercase", signature: "ARRAY OF INTEGER", expectNil: false, expectedType: "array of Integer", isDynamic: true, lowBound: 0, highBound: 0},
		{name: "Dynamic array mixed case", signature: "Array Of String", expectNil: false, expectedType: "array of String", isDynamic: true, lowBound: 0, highBound: 0},

		// Static arrays
		{name: "Static array 0..9", signature: "array[0..9] of Integer", expectNil: false, expectedType: "array[0..9] of Integer", isDynamic: false, lowBound: 0, highBound: 9},
		{name: "Static array 1..10", signature: "array[1..10] of String", expectNil: false, expectedType: "array[1..10] of String", isDynamic: false, lowBound: 1, highBound: 10},
		{name: "Static array -5..5", signature: "array[-5..5] of Float", expectNil: false, expectedType: "array[-5..5] of Float", isDynamic: false, lowBound: -5, highBound: 5},

		// Nested arrays
		{name: "2D dynamic array", signature: "array of array of Integer", expectNil: false, expectedType: "array of array of Integer", isDynamic: true, lowBound: 0, highBound: 0},
		{name: "2D static array", signature: "array[0..2] of array[0..3] of Integer", expectNil: false, expectedType: "array[0..2] of array[0..3] of Integer", isDynamic: false, lowBound: 0, highBound: 2},

		// Error cases
		{name: "Invalid syntax - no 'of'", signature: "array Integer", expectNil: true, expectedType: "", isDynamic: false, lowBound: 0, highBound: 0},
		{name: "Invalid syntax - missing element type", signature: "array of", expectNil: true, expectedType: "", isDynamic: false, lowBound: 0, highBound: 0},
		{name: "Invalid syntax - bad bounds", signature: "array[0-9] of Integer", expectNil: true, expectedType: "", isDynamic: false, lowBound: 0, highBound: 0},
		{name: "Invalid syntax - non-numeric bounds", signature: "array[a..b] of Integer", expectNil: true, expectedType: "", isDynamic: false, lowBound: 0, highBound: 0},
		{name: "Unknown element type", signature: "array of UnknownType", expectNil: true, expectedType: "", isDynamic: false, lowBound: 0, highBound: 0},
		{name: "Not an array", signature: "Integer", expectNil: true, expectedType: "", isDynamic: false, lowBound: 0, highBound: 0},
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
func TestResolveTypeName_InlineArrays(t *testing.T) {
	tests := []struct {
		name         string
		typeName     string
		expectedType string
		expectError  bool
	}{
		// Dynamic arrays via resolveTypeName
		{name: "Resolve dynamic array", typeName: "array of Integer", expectError: false, expectedType: "array of Integer"},
		{name: "Resolve dynamic string array", typeName: "array of String", expectError: false, expectedType: "array of String"},

		// Static arrays via resolveTypeName
		{name: "Resolve static array", typeName: "array[0..9] of Integer", expectError: false, expectedType: "array[0..9] of Integer"},
		{name: "Resolve static bounds array", typeName: "array[1..100] of String", expectError: false, expectedType: "array[1..100] of String"},

		// Nested arrays
		{name: "Resolve 2D array", typeName: "array of array of Integer", expectError: false, expectedType: "array of array of Integer"},

		// Error cases
		{name: "Invalid array syntax", typeName: "array Integer", expectError: true, expectedType: ""},
		{name: "Unknown element type", typeName: "array of CustomType", expectError: true, expectedType: ""},
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
func TestResolveArrayTypeNode(t *testing.T) {
	tests := []struct {
		setupNode    func() *ast.ArrayTypeNode
		name         string
		expectedType string
		expectNil    bool
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
