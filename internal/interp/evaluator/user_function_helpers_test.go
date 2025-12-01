package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// =============================================================================
// EvaluateDefaultParameters tests
// =============================================================================

// TestEvaluateDefaultParameters tests filling in missing optional arguments with defaults.
func TestEvaluateDefaultParameters(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	// Function with 3 parameters: x (required), y (default=10), z (default=20)
	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "x"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
				// No default - required
			},
			{
				Name:         &ast.Identifier{Value: "y"},
				Type:         &ast.TypeAnnotation{Name: "Integer"},
				DefaultValue: &ast.IntegerLiteral{Value: 10},
			},
			{
				Name:         &ast.Identifier{Value: "z"},
				Type:         &ast.TypeAnnotation{Name: "Integer"},
				DefaultValue: &ast.IntegerLiteral{Value: 20},
			},
		},
	}

	// Only provide first argument - should fill in y=10, z=20
	providedArgs := []Value{&runtime.IntegerValue{Value: 1}}

	result, err := e.EvaluateDefaultParameters(fn, providedArgs, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 args after defaults, got %d", len(result))
	}

	// First arg should be unchanged
	intVal, ok := result[0].(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected *runtime.IntegerValue for arg 0, got %T", result[0])
	}
	if intVal.Value != 1 {
		t.Errorf("expected arg 0 = 1, got %d", intVal.Value)
	}

	// Second arg should be default value 10
	intVal, ok = result[1].(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected *runtime.IntegerValue for arg 1, got %T", result[1])
	}
	if intVal.Value != 10 {
		t.Errorf("expected arg 1 = 10, got %d", intVal.Value)
	}

	// Third arg should be default value 20
	intVal, ok = result[2].(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected *runtime.IntegerValue for arg 2, got %T", result[2])
	}
	if intVal.Value != 20 {
		t.Errorf("expected arg 2 = 20, got %d", intVal.Value)
	}
}

// TestEvaluateDefaultParameters_AllProvided tests when all args are provided.
func TestEvaluateDefaultParameters_AllProvided(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
		Parameters: []*ast.Parameter{
			{
				Name:         &ast.Identifier{Value: "x"},
				Type:         &ast.TypeAnnotation{Name: "Integer"},
				DefaultValue: &ast.IntegerLiteral{Value: 100},
			},
		},
	}

	// Provide the argument - should not use default
	providedArgs := []Value{&runtime.IntegerValue{Value: 42}}

	result, err := e.EvaluateDefaultParameters(fn, providedArgs, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(result))
	}

	intVal, ok := result[0].(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected *runtime.IntegerValue, got %T", result[0])
	}
	if intVal.Value != 42 {
		t.Errorf("expected 42 (provided), got %d", intVal.Value)
	}
}

// TestEvaluateDefaultParameters_NoDefaults tests function with no default params.
func TestEvaluateDefaultParameters_NoDefaults(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "x"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
			},
			{
				Name: &ast.Identifier{Value: "y"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
			},
		},
	}

	providedArgs := []Value{
		&runtime.IntegerValue{Value: 1},
		&runtime.IntegerValue{Value: 2},
	}

	result, err := e.EvaluateDefaultParameters(fn, providedArgs, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 args, got %d", len(result))
	}
}

// TestEvaluateDefaultParameters_TooFewArgs tests error when too few args provided.
func TestEvaluateDefaultParameters_TooFewArgs(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "x"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
				// No default - required
			},
			{
				Name: &ast.Identifier{Value: "y"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
				// No default - required
			},
		},
	}

	// Provide only 1 arg but 2 are required
	providedArgs := []Value{&runtime.IntegerValue{Value: 1}}

	_, err := e.EvaluateDefaultParameters(fn, providedArgs, ctx)

	if err == nil {
		t.Error("expected error for too few arguments")
	}
}

// TestEvaluateDefaultParameters_TooManyArgs tests error when too many args provided.
func TestEvaluateDefaultParameters_TooManyArgs(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "x"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
			},
		},
	}

	// Provide 2 args but only 1 parameter
	providedArgs := []Value{
		&runtime.IntegerValue{Value: 1},
		&runtime.IntegerValue{Value: 2},
	}

	_, err := e.EvaluateDefaultParameters(fn, providedArgs, ctx)

	if err == nil {
		t.Error("expected error for too many arguments")
	}
}

// TestEvaluateDefaultParameters_ExpressionDefaults tests default values that are expressions.
func TestEvaluateDefaultParameters_ExpressionDefaults(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	// Default value is an expression: 5 + 5
	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "x"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
				DefaultValue: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 5},
					Operator: "+",
					Right:    &ast.IntegerLiteral{Value: 5},
				},
			},
		},
	}

	// Provide no args - should evaluate expression to 10
	providedArgs := []Value{}

	result, err := e.EvaluateDefaultParameters(fn, providedArgs, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(result))
	}

	intVal, ok := result[0].(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected *runtime.IntegerValue, got %T", result[0])
	}
	if intVal.Value != 10 {
		t.Errorf("expected 10 (5+5), got %d", intVal.Value)
	}
}

// TestEvaluateDefaultParameters_NoParameters tests function with no parameters.
func TestEvaluateDefaultParameters_NoParameters(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "NoParams"},
	}

	providedArgs := []Value{}

	result, err := e.EvaluateDefaultParameters(fn, providedArgs, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 args, got %d", len(result))
	}
}

// =============================================================================
// BindFunctionParameters tests
// =============================================================================

// testEnv is a simple mock environment for testing parameter binding.
type testEnv struct {
	bindings map[string]interface{}
}

func newTestEnv() *testEnv {
	return &testEnv{bindings: make(map[string]interface{})}
}

func (e *testEnv) Define(name string, value interface{}) {
	e.bindings[name] = value
}

func (e *testEnv) Get(name string) (interface{}, bool) {
	val, ok := e.bindings[name]
	return val, ok
}

func (e *testEnv) Set(name string, value interface{}) bool {
	if _, ok := e.bindings[name]; ok {
		e.bindings[name] = value
		return true
	}
	return false
}

func (e *testEnv) NewEnclosedEnvironment() Environment {
	child := newTestEnv()
	return child
}

// mockRefValue is a simple mock for reference values in tests.
// It's used to simulate var parameter references without needing runtime.ReferenceValue.
type mockRefValue struct {
	name string
}

func (m *mockRefValue) Type() string   { return "REFERENCE" }
func (m *mockRefValue) String() string { return "&" + m.name }

// TestBindFunctionParameters tests basic parameter binding to environment.
func TestBindFunctionParameters(t *testing.T) {
	e := &Evaluator{}
	// Create context with a fresh environment
	env := newTestEnv()
	ctx := NewExecutionContext(env)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "x"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
			},
			{
				Name: &ast.Identifier{Value: "y"},
				Type: &ast.TypeAnnotation{Name: "String"},
			},
		},
	}

	args := []Value{
		&runtime.IntegerValue{Value: 42},
		&runtime.StringValue{Value: "hello"},
	}

	// No conversion needed - pass nil converter
	err := e.BindFunctionParameters(fn, args, ctx, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check x was bound
	xVal, ok := ctx.Env().Get("x")
	if !ok {
		t.Fatal("expected 'x' to be defined in environment")
	}
	intVal, ok := xVal.(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected *runtime.IntegerValue for x, got %T", xVal)
	}
	if intVal.Value != 42 {
		t.Errorf("expected x = 42, got %d", intVal.Value)
	}

	// Check y was bound
	yVal, ok := ctx.Env().Get("y")
	if !ok {
		t.Fatal("expected 'y' to be defined in environment")
	}
	strVal, ok := yVal.(*runtime.StringValue)
	if !ok {
		t.Fatalf("expected *runtime.StringValue for y, got %T", yVal)
	}
	if strVal.Value != "hello" {
		t.Errorf("expected y = 'hello', got '%s'", strVal.Value)
	}
}

// TestBindFunctionParameters_WithConversion tests that implicit conversion is applied.
func TestBindFunctionParameters_WithConversion(t *testing.T) {
	e := &Evaluator{}
	env := newTestEnv()
	ctx := NewExecutionContext(env)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "x"},
				Type: &ast.TypeAnnotation{Name: "Float"},
			},
		},
	}

	// Pass an integer that should be converted to float
	args := []Value{&runtime.IntegerValue{Value: 42}}

	// Conversion callback that converts Integer to Float
	converter := func(value Value, targetTypeName string) (Value, bool) {
		if targetTypeName == "Float" {
			if intVal, ok := value.(*runtime.IntegerValue); ok {
				return &runtime.FloatValue{Value: float64(intVal.Value)}, true
			}
		}
		return value, false
	}

	err := e.BindFunctionParameters(fn, args, ctx, converter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check x was converted to float
	xVal, ok := ctx.Env().Get("x")
	if !ok {
		t.Fatal("expected 'x' to be defined in environment")
	}
	floatVal, ok := xVal.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("expected *runtime.FloatValue for x after conversion, got %T", xVal)
	}
	if floatVal.Value != 42.0 {
		t.Errorf("expected x = 42.0, got %f", floatVal.Value)
	}
}

// TestBindFunctionParameters_VarParameter tests that var parameters skip conversion.
func TestBindFunctionParameters_VarParameter(t *testing.T) {
	e := &Evaluator{}
	env := newTestEnv()
	ctx := NewExecutionContext(env)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
		Parameters: []*ast.Parameter{
			{
				Name:  &ast.Identifier{Value: "x"},
				Type:  &ast.TypeAnnotation{Name: "Integer"},
				ByRef: true, // var parameter
			},
		},
	}

	// Simulate a ReferenceValue that would be passed for var params
	refVal := &mockRefValue{name: "original"}
	args := []Value{refVal}

	// This converter should NOT be called for var parameters
	converterCalled := false
	converter := func(value Value, targetTypeName string) (Value, bool) {
		converterCalled = true
		return value, false
	}

	err := e.BindFunctionParameters(fn, args, ctx, converter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if converterCalled {
		t.Error("converter should not be called for var parameters")
	}

	// Check x was bound as reference (not converted)
	xVal, ok := ctx.Env().Get("x")
	if !ok {
		t.Fatal("expected 'x' to be defined in environment")
	}
	if xVal != refVal {
		t.Error("var parameter should keep its ReferenceValue unchanged")
	}
}

// TestBindFunctionParameters_NoParameters tests function with no parameters.
func TestBindFunctionParameters_NoParameters(t *testing.T) {
	e := &Evaluator{}
	env := newTestEnv()
	ctx := NewExecutionContext(env)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "NoParams"},
	}

	err := e.BindFunctionParameters(fn, []Value{}, ctx, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestBindFunctionParameters_MixedVarAndRegular tests mixed var and regular parameters.
func TestBindFunctionParameters_MixedVarAndRegular(t *testing.T) {
	e := &Evaluator{}
	env := newTestEnv()
	ctx := NewExecutionContext(env)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "MixedFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "a"},
				Type: &ast.TypeAnnotation{Name: "Float"},
				// Regular parameter - should convert
			},
			{
				Name:  &ast.Identifier{Value: "b"},
				Type:  &ast.TypeAnnotation{Name: "Integer"},
				ByRef: true, // var parameter - should not convert
			},
			{
				Name: &ast.Identifier{Value: "c"},
				Type: &ast.TypeAnnotation{Name: "Float"},
				// Regular parameter - should convert
			},
		},
	}

	refVal := &mockRefValue{name: "originalB"}
	args := []Value{
		&runtime.IntegerValue{Value: 1}, // Should be converted to Float
		refVal,                          // Should stay as reference
		&runtime.IntegerValue{Value: 3}, // Should be converted to Float
	}

	conversionCount := 0
	converter := func(value Value, targetTypeName string) (Value, bool) {
		if targetTypeName == "Float" {
			if intVal, ok := value.(*runtime.IntegerValue); ok {
				conversionCount++
				return &runtime.FloatValue{Value: float64(intVal.Value)}, true
			}
		}
		return value, false
	}

	err := e.BindFunctionParameters(fn, args, ctx, converter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have converted 2 values (a and c, but not b)
	if conversionCount != 2 {
		t.Errorf("expected 2 conversions, got %d", conversionCount)
	}

	// Check a was converted
	aVal, _ := ctx.Env().Get("a")
	if _, ok := aVal.(*runtime.FloatValue); !ok {
		t.Errorf("expected 'a' to be FloatValue, got %T", aVal)
	}

	// Check b is still reference
	bVal, _ := ctx.Env().Get("b")
	if bVal != refVal {
		t.Error("expected 'b' to stay as ReferenceValue")
	}

	// Check c was converted
	cVal, _ := ctx.Env().Get("c")
	if _, ok := cVal.(*runtime.FloatValue); !ok {
		t.Errorf("expected 'c' to be FloatValue, got %T", cVal)
	}
}

// TestBindFunctionParameters_NilType tests parameter without type annotation.
func TestBindFunctionParameters_NilType(t *testing.T) {
	e := &Evaluator{}
	env := newTestEnv()
	ctx := NewExecutionContext(env)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "x"},
				Type: nil, // No type annotation
			},
		},
	}

	args := []Value{&runtime.IntegerValue{Value: 42}}

	// Converter should not be called when there's no type
	converterCalled := false
	converter := func(value Value, targetTypeName string) (Value, bool) {
		converterCalled = true
		return value, false
	}

	err := e.BindFunctionParameters(fn, args, ctx, converter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if converterCalled {
		t.Error("converter should not be called when parameter has no type")
	}

	// Value should still be bound
	xVal, ok := ctx.Env().Get("x")
	if !ok {
		t.Fatal("expected 'x' to be defined")
	}
	if _, ok := xVal.(*runtime.IntegerValue); !ok {
		t.Errorf("expected IntegerValue, got %T", xVal)
	}
}

// =============================================================================
// InitializeResultVariable tests
// =============================================================================

// TestInitializeResultVariable_Procedure tests that procedures don't initialize Result.
func TestInitializeResultVariable_Procedure(t *testing.T) {
	e := &Evaluator{}
	env := newTestEnv()
	ctx := NewExecutionContext(env)

	// Procedure - no return type
	fn := &ast.FunctionDecl{
		Name:       &ast.Identifier{Value: "TestProc"},
		ReturnType: nil,
	}

	err := e.InitializeResultVariable(fn, ctx, nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Result should NOT be defined for procedures
	_, ok := ctx.Env().Get("Result")
	if ok {
		t.Error("Result should not be defined for procedures")
	}
}

// TestInitializeResultVariable_Function tests that functions initialize Result.
func TestInitializeResultVariable_Function(t *testing.T) {
	e := &Evaluator{}
	env := newTestEnv()
	ctx := NewExecutionContext(env)

	fn := &ast.FunctionDecl{
		Name:       &ast.Identifier{Value: "TestFunc"},
		ReturnType: &ast.TypeAnnotation{Name: "Integer"},
	}

	// Callback that returns default integer value
	defaultValueGetter := func(returnTypeName string) Value {
		if returnTypeName == "Integer" {
			return &runtime.IntegerValue{Value: 0}
		}
		return &runtime.NilValue{}
	}

	err := e.InitializeResultVariable(fn, ctx, defaultValueGetter, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Result should be defined
	resultVal, ok := ctx.Env().Get("Result")
	if !ok {
		t.Fatal("Result should be defined for functions")
	}

	intVal, ok := resultVal.(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected *runtime.IntegerValue for Result, got %T", resultVal)
	}
	if intVal.Value != 0 {
		t.Errorf("expected Result = 0, got %d", intVal.Value)
	}
}

// TestInitializeResultVariable_FunctionNameAlias tests that function name is aliased to Result.
func TestInitializeResultVariable_FunctionNameAlias(t *testing.T) {
	e := &Evaluator{}
	env := newTestEnv()
	ctx := NewExecutionContext(env)

	fn := &ast.FunctionDecl{
		Name:       &ast.Identifier{Value: "GetValue"},
		ReturnType: &ast.TypeAnnotation{Name: "String"},
	}

	defaultValueGetter := func(returnTypeName string) Value {
		return &runtime.StringValue{Value: ""}
	}

	// Track if alias creator was called with correct params
	aliasCreatorCalled := false
	aliasCreator := func(funcName string, funcEnv Environment) Value {
		aliasCreatorCalled = true
		if funcName != "GetValue" {
			t.Errorf("expected funcName 'GetValue', got '%s'", funcName)
		}
		// Return a mock reference value
		return &mockRefValue{name: "Result"}
	}

	err := e.InitializeResultVariable(fn, ctx, defaultValueGetter, aliasCreator)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !aliasCreatorCalled {
		t.Error("alias creator should be called for functions")
	}

	// Function name should be defined
	aliasVal, ok := ctx.Env().Get("GetValue")
	if !ok {
		t.Fatal("function name should be defined as alias")
	}

	// Should be the mock reference value
	if _, ok := aliasVal.(*mockRefValue); !ok {
		t.Errorf("expected mockRefValue, got %T", aliasVal)
	}
}

// TestInitializeResultVariable_NilCallbacks tests that nil callbacks are handled.
func TestInitializeResultVariable_NilCallbacks(t *testing.T) {
	e := &Evaluator{}
	env := newTestEnv()
	ctx := NewExecutionContext(env)

	fn := &ast.FunctionDecl{
		Name:       &ast.Identifier{Value: "TestFunc"},
		ReturnType: &ast.TypeAnnotation{Name: "Integer"},
	}

	// With nil defaultValueGetter, Result should be NilValue
	err := e.InitializeResultVariable(fn, ctx, nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultVal, ok := ctx.Env().Get("Result")
	if !ok {
		t.Fatal("Result should be defined")
	}

	// With nil callback, should default to NilValue
	if _, ok := resultVal.(*runtime.NilValue); !ok {
		t.Errorf("expected NilValue with nil callback, got %T", resultVal)
	}

	// Function name should NOT be defined with nil alias creator
	_, ok = ctx.Env().Get("TestFunc")
	if ok {
		t.Error("function name should not be defined with nil alias creator")
	}
}

// TestInitializeResultVariable_FloatDefault tests float return type initialization.
func TestInitializeResultVariable_FloatDefault(t *testing.T) {
	e := &Evaluator{}
	env := newTestEnv()
	ctx := NewExecutionContext(env)

	fn := &ast.FunctionDecl{
		Name:       &ast.Identifier{Value: "GetFloat"},
		ReturnType: &ast.TypeAnnotation{Name: "Float"},
	}

	defaultValueGetter := func(returnTypeName string) Value {
		if returnTypeName == "Float" {
			return &runtime.FloatValue{Value: 0.0}
		}
		return &runtime.NilValue{}
	}

	err := e.InitializeResultVariable(fn, ctx, defaultValueGetter, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultVal, ok := ctx.Env().Get("Result")
	if !ok {
		t.Fatal("Result should be defined")
	}

	floatVal, ok := resultVal.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("expected *runtime.FloatValue, got %T", resultVal)
	}
	if floatVal.Value != 0.0 {
		t.Errorf("expected 0.0, got %f", floatVal.Value)
	}
}
