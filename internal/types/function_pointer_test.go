package types

import (
	"strings"
	"testing"
)

// TestNewFunctionPointerType tests the creation of function pointer types.
func TestNewFunctionPointerType(t *testing.T) {
	t.Run("function with parameters and return type", func(t *testing.T) {
		params := []Type{&IntegerType{}, &StringType{}}
		returnType := Type(&BooleanType{})
		funcPtr := NewFunctionPointerType(params, returnType)

		if funcPtr == nil {
			t.Fatal("expected non-nil function pointer type")
		}
		if len(funcPtr.Parameters) != 2 {
			t.Errorf("expected 2 parameters, got %d", len(funcPtr.Parameters))
		}
		if funcPtr.ReturnType == nil {
			t.Error("expected non-nil return type")
		}
	})

	t.Run("function with no parameters", func(t *testing.T) {
		params := []Type{}
		returnType := Type(&IntegerType{})
		funcPtr := NewFunctionPointerType(params, returnType)

		if funcPtr == nil {
			t.Fatal("expected non-nil function pointer type")
		}
		if len(funcPtr.Parameters) != 0 {
			t.Errorf("expected 0 parameters, got %d", len(funcPtr.Parameters))
		}
		if funcPtr.ReturnType == nil {
			t.Error("expected non-nil return type")
		}
	})

	t.Run("function with many parameters", func(t *testing.T) {
		params := []Type{
			&IntegerType{},
			&FloatType{},
			&StringType{},
			&BooleanType{},
		}
		returnType := Type(&StringType{})
		funcPtr := NewFunctionPointerType(params, returnType)

		if funcPtr == nil {
			t.Fatal("expected non-nil function pointer type")
		}
		if len(funcPtr.Parameters) != 4 {
			t.Errorf("expected 4 parameters, got %d", len(funcPtr.Parameters))
		}
	})
}

// TestNewProcedurePointerType tests the creation of procedure pointer types.
func TestNewProcedurePointerType(t *testing.T) {
	t.Run("procedure with parameters", func(t *testing.T) {
		params := []Type{&StringType{}, &IntegerType{}}
		procPtr := NewProcedurePointerType(params)

		if procPtr == nil {
			t.Fatal("expected non-nil procedure pointer type")
		}
		if len(procPtr.Parameters) != 2 {
			t.Errorf("expected 2 parameters, got %d", len(procPtr.Parameters))
		}
		if procPtr.ReturnType != nil {
			t.Error("expected nil return type for procedure")
		}
	})

	t.Run("procedure with no parameters", func(t *testing.T) {
		params := []Type{}
		procPtr := NewProcedurePointerType(params)

		if procPtr == nil {
			t.Fatal("expected non-nil procedure pointer type")
		}
		if len(procPtr.Parameters) != 0 {
			t.Errorf("expected 0 parameters, got %d", len(procPtr.Parameters))
		}
		if procPtr.ReturnType != nil {
			t.Error("expected nil return type for procedure")
		}
	})
}

// TestNewMethodPointerType tests the creation of method pointer types.
func TestNewMethodPointerType(t *testing.T) {
	t.Run("method with parameters and return type", func(t *testing.T) {
		params := []Type{&IntegerType{}, &StringType{}}
		returnType := Type(&BooleanType{})
		methodPtr := NewMethodPointerType(params, returnType)

		if methodPtr == nil {
			t.Fatal("expected non-nil method pointer type")
		}
		if len(methodPtr.Parameters) != 2 {
			t.Errorf("expected 2 parameters, got %d", len(methodPtr.Parameters))
		}
		if methodPtr.ReturnType == nil {
			t.Error("expected non-nil return type")
		}
		if !methodPtr.OfObject {
			t.Error("expected OfObject to be true")
		}
	})

	t.Run("method procedure with no return type", func(t *testing.T) {
		params := []Type{&StringType{}}
		methodPtr := NewMethodPointerType(params, nil)

		if methodPtr == nil {
			t.Fatal("expected non-nil method pointer type")
		}
		if len(methodPtr.Parameters) != 1 {
			t.Errorf("expected 1 parameter, got %d", len(methodPtr.Parameters))
		}
		if methodPtr.ReturnType != nil {
			t.Error("expected nil return type for method procedure")
		}
		if !methodPtr.OfObject {
			t.Error("expected OfObject to be true")
		}
	})
}

// TestFunctionPointerTypeKind tests the TypeKind method for function pointers.
func TestFunctionPointerTypeKind(t *testing.T) {
	t.Run("function pointer type kind", func(t *testing.T) {
		funcPtr := NewFunctionPointerType(
			[]Type{&IntegerType{}},
			&BooleanType{},
		)
		if funcPtr.TypeKind() != "FUNCTION_POINTER" {
			t.Errorf("expected FUNCTION_POINTER, got %s", funcPtr.TypeKind())
		}
	})

	t.Run("method pointer type kind", func(t *testing.T) {
		methodPtr := NewMethodPointerType(
			[]Type{&IntegerType{}},
			&BooleanType{},
		)
		if methodPtr.TypeKind() != "METHOD_POINTER" {
			t.Errorf("expected METHOD_POINTER, got %s", methodPtr.TypeKind())
		}
	})
}

// TestFunctionPointerString tests the String method for function pointers.
func TestFunctionPointerString(t *testing.T) {
	tests := []struct {
		name     string
		funcPtr  *FunctionPointerType
		expected string
	}{
		{
			name: "function with two parameters",
			funcPtr: NewFunctionPointerType(
				[]Type{&IntegerType{}, &StringType{}},
				&BooleanType{},
			),
			expected: "function(Integer, String): Boolean",
		},
		{
			name: "function with no parameters",
			funcPtr: NewFunctionPointerType(
				[]Type{},
				&IntegerType{},
			),
			expected: "function(): Integer",
		},
		{
			name: "procedure with one parameter",
			funcPtr: NewProcedurePointerType(
				[]Type{&StringType{}},
			),
			expected: "procedure(String)",
		},
		{
			name: "procedure with no parameters",
			funcPtr: NewProcedurePointerType(
				[]Type{},
			),
			expected: "procedure()",
		},
		{
			name: "function with many parameters",
			funcPtr: NewFunctionPointerType(
				[]Type{&IntegerType{}, &FloatType{}, &StringType{}, &BooleanType{}},
				&StringType{},
			),
			expected: "function(Integer, Float, String, Boolean): String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.funcPtr.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestMethodPointerString tests the String method for method pointers.
func TestMethodPointerString(t *testing.T) {
	tests := []struct {
		name      string
		methodPtr *MethodPointerType
		expected  string
	}{
		{
			name: "method with parameters and return type",
			methodPtr: NewMethodPointerType(
				[]Type{&IntegerType{}, &StringType{}},
				&BooleanType{},
			),
			expected: "function(Integer, String): Boolean of object",
		},
		{
			name: "method procedure with one parameter",
			methodPtr: NewMethodPointerType(
				[]Type{&StringType{}},
				nil,
			),
			expected: "procedure(String) of object",
		},
		{
			name: "method with no parameters",
			methodPtr: NewMethodPointerType(
				[]Type{},
				&IntegerType{},
			),
			expected: "function(): Integer of object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.methodPtr.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFunctionPointerIsProcedure tests the IsProcedure method.
func TestFunctionPointerIsProcedure(t *testing.T) {
	tests := []struct {
		funcPtr  *FunctionPointerType
		name     string
		expected bool
	}{
		{
			name: "function with return type",
			funcPtr: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			expected: false,
		},
		{
			name: "procedure with no return type",
			funcPtr: NewProcedurePointerType(
				[]Type{&StringType{}},
			),
			expected: true,
		},
		{
			name: "function with nil explicitly",
			funcPtr: &FunctionPointerType{
				Parameters: []Type{&IntegerType{}},
				ReturnType: nil,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.funcPtr.IsProcedure()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestFunctionPointerIsFunction tests the IsFunction method.
func TestFunctionPointerIsFunction(t *testing.T) {
	tests := []struct {
		funcPtr  *FunctionPointerType
		name     string
		expected bool
	}{
		{
			name: "function with return type",
			funcPtr: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			expected: true,
		},
		{
			name: "procedure with no return type",
			funcPtr: NewProcedurePointerType(
				[]Type{&StringType{}},
			),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.funcPtr.IsFunction()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestFunctionPointerEquals tests the Equals method for function pointers.
func TestFunctionPointerEquals(t *testing.T) {
	tests := []struct {
		a        Type
		b        Type
		name     string
		expected bool
	}{
		{
			name: "identical function signatures",
			a: NewFunctionPointerType(
				[]Type{&IntegerType{}, &StringType{}},
				&BooleanType{},
			),
			b: NewFunctionPointerType(
				[]Type{&IntegerType{}, &StringType{}},
				&BooleanType{},
			),
			expected: true,
		},
		{
			name: "identical procedure signatures",
			a: NewProcedurePointerType(
				[]Type{&IntegerType{}, &StringType{}},
			),
			b: NewProcedurePointerType(
				[]Type{&IntegerType{}, &StringType{}},
			),
			expected: true,
		},
		{
			name: "different parameter count",
			a: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			b: NewFunctionPointerType(
				[]Type{&IntegerType{}, &StringType{}},
				&BooleanType{},
			),
			expected: false,
		},
		{
			name: "different parameter types",
			a: NewFunctionPointerType(
				[]Type{&IntegerType{}, &StringType{}},
				&BooleanType{},
			),
			b: NewFunctionPointerType(
				[]Type{&IntegerType{}, &FloatType{}},
				&BooleanType{},
			),
			expected: false,
		},
		{
			name: "different return types",
			a: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			b: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&StringType{},
			),
			expected: false,
		},
		{
			name: "function vs procedure",
			a: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			b: NewProcedurePointerType(
				[]Type{&IntegerType{}},
			),
			expected: false,
		},
		{
			name: "function pointer vs method pointer with same signature",
			a: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			b: NewMethodPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			expected: false,
		},
		{
			name: "function pointer vs different type",
			a: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			b:        &IntegerType{},
			expected: false,
		},
		{
			name: "empty parameter list",
			a: NewFunctionPointerType(
				[]Type{},
				&IntegerType{},
			),
			b: NewFunctionPointerType(
				[]Type{},
				&IntegerType{},
			),
			expected: true,
		},
		{
			name: "nil vs empty slice parameters (equivalent)",
			a: NewFunctionPointerType(
				nil,
				&IntegerType{},
			),
			b: NewFunctionPointerType(
				[]Type{},
				&IntegerType{},
			),
			expected: true, // nil and empty slice both mean "no parameters"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Equals(tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
			// Test symmetry
			result2 := tt.b.Equals(tt.a)
			if result2 != tt.expected {
				t.Errorf("symmetry failed: expected %v, got %v", tt.expected, result2)
			}
		})
	}
}

// TestMethodPointerEquals tests the Equals method for method pointers.
func TestMethodPointerEquals(t *testing.T) {
	tests := []struct {
		a        Type
		b        Type
		name     string
		expected bool
	}{
		{
			name: "identical method signatures",
			a: NewMethodPointerType(
				[]Type{&IntegerType{}, &StringType{}},
				&BooleanType{},
			),
			b: NewMethodPointerType(
				[]Type{&IntegerType{}, &StringType{}},
				&BooleanType{},
			),
			expected: true,
		},
		{
			name: "identical method procedure signatures",
			a: NewMethodPointerType(
				[]Type{&IntegerType{}},
				nil,
			),
			b: NewMethodPointerType(
				[]Type{&IntegerType{}},
				nil,
			),
			expected: true,
		},
		{
			name: "different parameter types",
			a: NewMethodPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			b: NewMethodPointerType(
				[]Type{&StringType{}},
				&BooleanType{},
			),
			expected: false,
		},
		{
			name: "method pointer vs function pointer with same signature",
			a: NewMethodPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			b: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			expected: false,
		},
		{
			name: "method pointer vs different type",
			a: NewMethodPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			b:        &StringType{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Equals(tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestFunctionPointerCompatibility tests the IsCompatibleWith method.
func TestFunctionPointerCompatibility(t *testing.T) {
	tests := []struct {
		source   Type
		target   Type
		name     string
		expected bool
	}{
		{
			name: "compatible function pointers",
			source: NewFunctionPointerType(
				[]Type{&IntegerType{}, &StringType{}},
				&BooleanType{},
			),
			target: NewFunctionPointerType(
				[]Type{&IntegerType{}, &StringType{}},
				&BooleanType{},
			),
			expected: true,
		},
		{
			name: "incompatible parameter types",
			source: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			target: NewFunctionPointerType(
				[]Type{&StringType{}},
				&BooleanType{},
			),
			expected: false,
		},
		{
			name: "incompatible return types",
			source: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			target: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&StringType{},
			),
			expected: false,
		},
		{
			name: "method pointer to function pointer (compatible)",
			source: NewMethodPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			target: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			expected: true,
		},
		{
			name: "function pointer to method pointer (incompatible)",
			source: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			target: NewMethodPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			expected: false,
		},
		{
			name: "function pointer to unrelated type",
			source: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			target:   &StringType{},
			expected: false,
		},
		{
			name: "compatible procedures",
			source: NewProcedurePointerType(
				[]Type{&IntegerType{}},
			),
			target: NewProcedurePointerType(
				[]Type{&IntegerType{}},
			),
			expected: true,
		},
		{
			name: "procedure vs function (incompatible)",
			source: NewProcedurePointerType(
				[]Type{&IntegerType{}},
			),
			target: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test function pointer compatibility
			if funcPtr, ok := tt.source.(*FunctionPointerType); ok {
				result := funcPtr.IsCompatibleWith(tt.target)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
			// Test method pointer compatibility
			if methodPtr, ok := tt.source.(*MethodPointerType); ok {
				result := methodPtr.IsCompatibleWith(tt.target)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// TestMethodPointerCompatibility tests method pointer specific compatibility rules.
func TestMethodPointerCompatibility(t *testing.T) {
	tests := []struct {
		target   Type
		source   *MethodPointerType
		name     string
		expected bool
	}{
		{
			name: "method to method (compatible)",
			source: NewMethodPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			target: NewMethodPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			expected: true,
		},
		{
			name: "method to function (compatible - can call method as function)",
			source: NewMethodPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			target: NewFunctionPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			expected: true,
		},
		{
			name: "method to incompatible function",
			source: NewMethodPointerType(
				[]Type{&IntegerType{}},
				&BooleanType{},
			),
			target: NewFunctionPointerType(
				[]Type{&StringType{}},
				&BooleanType{},
			),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.source.IsCompatibleWith(tt.target)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestFunctionPointerWithComplexTypes tests function pointers with complex parameter types.
func TestFunctionPointerWithComplexTypes(t *testing.T) {
	t.Run("function with array parameter", func(t *testing.T) {
		arrayType := NewDynamicArrayType(&IntegerType{})
		funcPtr := NewFunctionPointerType(
			[]Type{arrayType},
			&BooleanType{},
		)

		expected := "function(array of Integer): Boolean"
		if funcPtr.String() != expected {
			t.Errorf("expected %q, got %q", expected, funcPtr.String())
		}
	})

	t.Run("function with record parameter", func(t *testing.T) {
		recordType := NewRecordType("TPoint", map[string]Type{
			"X": &IntegerType{},
			"Y": &IntegerType{},
		})
		funcPtr := NewFunctionPointerType(
			[]Type{recordType},
			&BooleanType{},
		)

		result := funcPtr.String()
		// strings.Contains is available in the standard library
		if !strings.Contains(result, "TPoint") {
			t.Errorf("expected string to contain 'TPoint', got %q", result)
		}
	})

	t.Run("nested function pointer", func(t *testing.T) {
		innerFunc := NewFunctionPointerType(
			[]Type{&IntegerType{}},
			&BooleanType{},
		)
		outerFunc := NewFunctionPointerType(
			[]Type{innerFunc},
			&StringType{},
		)

		result := outerFunc.String()
		if !strings.Contains(result, "function") {
			t.Errorf("expected nested function in string, got %q", result)
		}
	})
}
