package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestExecuteConversionFunction_Validation tests parameter validation.
func TestExecuteConversionFunction_Validation(t *testing.T) {
	// Create evaluator with minimal setup
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)

	// Create execution context
	ctx := NewExecutionContext(nil)

	// Test with function that has no parameters (should fail)
	t.Run("no parameters", func(t *testing.T) {
		fn := &ast.FunctionDecl{
			Name:       &ast.Identifier{Value: "InvalidConversion"},
			Parameters: []*ast.Parameter{}, // Empty
			ReturnType: &ast.TypeAnnotation{Name: "String"},
		}

		_, err := e.ExecuteConversionFunction(fn, &runtime.IntegerValue{Value: 42}, ctx, nil)
		if err == nil {
			t.Error("expected error for conversion function with no parameters")
		}
		if err.Error() != "conversion function 'InvalidConversion' must have exactly 1 parameter, got 0" {
			t.Errorf("unexpected error message: %s", err.Error())
		}
	})

	// Test with function that has multiple parameters (should fail)
	t.Run("multiple parameters", func(t *testing.T) {
		fn := &ast.FunctionDecl{
			Name: &ast.Identifier{Value: "InvalidConversion"},
			Parameters: []*ast.Parameter{
				{Name: &ast.Identifier{Value: "a"}},
				{Name: &ast.Identifier{Value: "b"}},
			},
			ReturnType: &ast.TypeAnnotation{Name: "String"},
		}

		_, err := e.ExecuteConversionFunction(fn, &runtime.IntegerValue{Value: 42}, ctx, nil)
		if err == nil {
			t.Error("expected error for conversion function with multiple parameters")
		}
		if err.Error() != "conversion function 'InvalidConversion' must have exactly 1 parameter, got 2" {
			t.Errorf("unexpected error message: %s", err.Error())
		}
	})

	// Test with procedure (no return type) - should fail
	t.Run("no return type", func(t *testing.T) {
		fn := &ast.FunctionDecl{
			Name: &ast.Identifier{Value: "InvalidConversion"},
			Parameters: []*ast.Parameter{
				{Name: &ast.Identifier{Value: "value"}},
			},
			ReturnType: nil, // No return type (procedure)
		}

		_, err := e.ExecuteConversionFunction(fn, &runtime.IntegerValue{Value: 42}, ctx, nil)
		if err == nil {
			t.Error("expected error for conversion function without return type")
		}
		if err.Error() != "conversion function 'InvalidConversion' must have a return type" {
			t.Errorf("unexpected error message: %s", err.Error())
		}
	})
}

// TestExecuteConversionFunction_DirectExecution verifies conversion functions run
// through the evaluator boundary without adapter scaffolding.
func TestExecuteConversionFunction_DirectExecution(t *testing.T) {
	// Create evaluator with minimal setup
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)

	// Create execution context with a proper environment
	env := runtime.NewEnvironment()
	ctx := NewExecutionContext(env)

	t.Run("returns assigned Result", func(t *testing.T) {
		fn := &ast.FunctionDecl{
			Name: &ast.Identifier{Value: "IntToStr"},
			Parameters: []*ast.Parameter{
				{
					Name: &ast.Identifier{Value: "value"},
					Type: &ast.TypeAnnotation{Name: "Integer"},
				},
			},
			ReturnType: &ast.TypeAnnotation{Name: "String"},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.AssignmentStatement{
					Target: &ast.Identifier{Value: "Result"},
					Value:  &ast.StringLiteral{Value: "42"},
				},
			}},
		}

		result, err := e.ExecuteConversionFunction(
			fn,
			&runtime.IntegerValue{Value: 42},
			ctx,
			nil, // No callbacks needed for this test
		)

		if err != nil {
			t.Fatalf("ExecuteConversionFunction failed: %v", err)
		}
		strVal, ok := result.(*runtime.StringValue)
		if !ok {
			t.Fatalf("expected StringValue, got %T", result)
		}
		if strVal.Value != "42" {
			t.Fatalf("expected Result to be %q, got %q", "42", strVal.Value)
		}
	})
}

// TestExecuteConversionFunctionSimple tests the simplified version.
func TestExecuteConversionFunctionSimple_Validation(t *testing.T) {
	// Create evaluator with minimal setup
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)

	// Create execution context
	ctx := NewExecutionContext(nil)

	// Test validation still works without adapter
	t.Run("validation without adapter", func(t *testing.T) {
		fn := &ast.FunctionDecl{
			Name:       &ast.Identifier{Value: "InvalidConversion"},
			Parameters: []*ast.Parameter{}, // Invalid: no parameters
			ReturnType: &ast.TypeAnnotation{Name: "String"},
		}

		_, err := e.ExecuteConversionFunctionSimple(fn, &runtime.IntegerValue{Value: 42}, ctx)
		if err == nil {
			t.Error("expected error for conversion function with no parameters")
		}
	})
}

// TestConversionCallbacks_NilHandling tests that nil callbacks are handled gracefully.
func TestConversionCallbacks_NilHandling(t *testing.T) {
	// Create evaluator with minimal setup
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)

	env := runtime.NewEnvironment()
	ctx := NewExecutionContext(env)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestConversion"},
		Parameters: []*ast.Parameter{
			{Name: &ast.Identifier{Value: "value"}},
		},
		ReturnType: &ast.TypeAnnotation{Name: "String"},
		Body: &ast.BlockStatement{Statements: []ast.Statement{
			&ast.AssignmentStatement{
				Target: &ast.Identifier{Value: "Result"},
				Value:  &ast.StringLiteral{Value: "ok"},
			},
		}},
	}

	t.Run("nil ConversionCallbacks", func(t *testing.T) {
		result, err := e.ExecuteConversionFunction(fn, &runtime.IntegerValue{Value: 1}, ctx, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strVal, ok := result.(*runtime.StringValue); !ok || strVal.Value != "ok" {
			t.Fatalf("expected conversion result %q, got %#v", "ok", result)
		}
	})

	t.Run("empty ConversionCallbacks", func(t *testing.T) {
		callbacks := &ConversionCallbacks{}
		result, err := e.ExecuteConversionFunction(fn, &runtime.IntegerValue{Value: 1}, ctx, callbacks)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strVal, ok := result.(*runtime.StringValue); !ok || strVal.Value != "ok" {
			t.Fatalf("expected conversion result %q, got %#v", "ok", result)
		}
	})

	t.Run("partial ConversionCallbacks", func(t *testing.T) {
		callbacks := &ConversionCallbacks{
			ImplicitConversion: func(value Value, targetTypeName string) (Value, bool) {
				return value, false
			},
		}
		result, err := e.ExecuteConversionFunction(fn, &runtime.IntegerValue{Value: 1}, ctx, callbacks)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strVal, ok := result.(*runtime.StringValue); !ok || strVal.Value != "ok" {
			t.Fatalf("expected conversion result %q, got %#v", "ok", result)
		}
	})
}

// TestTryImplicitConversion_NilValue tests nil value handling.
func TestTryImplicitConversion_NilValue(t *testing.T) {
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)
	ctx := NewExecutionContext(nil)

	result, ok := e.TryImplicitConversion(nil, "Integer", ctx)
	if ok {
		t.Error("expected no conversion for nil value")
	}
	if result != nil {
		t.Error("expected nil result for nil value")
	}
}

// TestTryImplicitConversion_SameType tests that same types return false (no conversion needed).
func TestTryImplicitConversion_SameType(t *testing.T) {
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)
	ctx := NewExecutionContext(nil)

	value := &runtime.IntegerValue{Value: 42}

	// INTEGER == INTEGER (exact match)
	result, ok := e.TryImplicitConversion(value, "INTEGER", ctx)
	if ok {
		t.Error("expected no conversion for same type")
	}
	if result != value {
		t.Error("expected original value returned for same type")
	}
}

// TestTryImplicitConversion_IntegerToFloat tests the built-in Integer→Float widening conversion.
func TestTryImplicitConversion_IntegerToFloat(t *testing.T) {
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)
	ctx := NewExecutionContext(nil)

	testCases := []struct {
		name     string
		input    int64
		expected float64
	}{
		{"positive", 42, 42.0},
		{"negative", -100, -100.0},
		{"zero", 0, 0.0},
		{"large", 9223372036854775807, 9223372036854775807.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := &runtime.IntegerValue{Value: tc.input}
			result, ok := e.TryImplicitConversion(value, "Float", ctx)

			if !ok {
				t.Error("expected conversion to succeed for Integer→Float")
			}

			floatResult, isFloat := result.(*runtime.FloatValue)
			if !isFloat {
				t.Errorf("expected FloatValue, got %T", result)
				return
			}

			if floatResult.Value != tc.expected {
				t.Errorf("expected %f, got %f", tc.expected, floatResult.Value)
			}
		})
	}
}

// TestTryImplicitConversion_EnumToInteger tests the built-in Enum→Integer conversion.
func TestTryImplicitConversion_EnumToInteger(t *testing.T) {
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)
	ctx := NewExecutionContext(nil)

	testCases := []struct {
		name            string
		enumType        string
		enumValue       string
		ordinal         int
		expectedInteger int64
	}{
		{"first_value", "TColor", "Red", 0, 0},
		{"second_value", "TColor", "Green", 1, 1},
		{"custom_ordinal", "TDay", "Wednesday", 3, 3},
		{"large_ordinal", "TMonth", "December", 12, 12},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := &runtime.EnumValue{
				TypeName:     tc.enumType,
				ValueName:    tc.enumValue,
				OrdinalValue: tc.ordinal,
			}

			result, ok := e.TryImplicitConversion(value, "Integer", ctx)

			if !ok {
				t.Error("expected conversion to succeed for Enum→Integer")
			}

			intResult, isInt := result.(*runtime.IntegerValue)
			if !isInt {
				t.Errorf("expected IntegerValue, got %T", result)
				return
			}

			if intResult.Value != tc.expectedInteger {
				t.Errorf("expected %d, got %d", tc.expectedInteger, intResult.Value)
			}
		})
	}
}

// TestTryImplicitConversion_NoConversionAvailable tests behavior when no conversion exists.
func TestTryImplicitConversion_NoConversionAvailable(t *testing.T) {
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)
	ctx := NewExecutionContext(nil)

	// String → Integer is not a built-in conversion
	value := &runtime.StringValue{Value: "42"}
	result, ok := e.TryImplicitConversion(value, "Integer", ctx)

	if ok {
		t.Error("expected no conversion for String→Integer without registered conversion")
	}

	if result != value {
		t.Error("expected original value returned when no conversion available")
	}
}

// TestTryImplicitConversion_FloatToInteger tests that Float→Integer is not allowed implicitly.
func TestTryImplicitConversion_FloatToInteger(t *testing.T) {
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)
	ctx := NewExecutionContext(nil)

	// Float → Integer is NOT an implicit conversion (would lose precision)
	value := &runtime.FloatValue{Value: 42.5}
	result, ok := e.TryImplicitConversion(value, "Integer", ctx)

	if ok {
		t.Error("expected no implicit conversion for Float→Integer (would lose precision)")
	}

	if result != value {
		t.Error("expected original value returned when no conversion available")
	}
}

// TestTryImplicitConversion_TypeNormalization tests that type names are properly normalized.
func TestTryImplicitConversion_TypeNormalization(t *testing.T) {
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)
	ctx := NewExecutionContext(nil)

	value := &runtime.IntegerValue{Value: 42}

	// Test various capitalizations of Float
	testCases := []string{"Float", "FLOAT", "float", "FloAt"}

	for _, targetType := range testCases {
		t.Run(targetType, func(t *testing.T) {
			result, ok := e.TryImplicitConversion(value, targetType, ctx)

			if !ok {
				t.Errorf("expected conversion to succeed for Integer→%s", targetType)
			}

			if _, isFloat := result.(*runtime.FloatValue); !isFloat {
				t.Errorf("expected FloatValue for Integer→%s, got %T", targetType, result)
			}
		})
	}
}

// TestIsErrorValue tests the isErrorValue helper function.
func TestIsErrorValue(t *testing.T) {
	testCases := []struct {
		value    Value
		name     string
		expected bool
	}{
		{name: "nil value", value: nil, expected: false},
		{name: "integer value", value: &runtime.IntegerValue{Value: 42}, expected: false},
		{name: "string value", value: &runtime.StringValue{Value: "test"}, expected: false},
		{name: "nil value struct", value: &runtime.NilValue{}, expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isErrorValue(tc.value)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
