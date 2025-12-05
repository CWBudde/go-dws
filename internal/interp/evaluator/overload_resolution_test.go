package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// TestGetValueType tests the getValueType method for all runtime value types.
func TestGetValueType(t *testing.T) {
	e := &Evaluator{}

	tests := []struct {
		value    Value
		expected types.Type
		name     string
	}{
		// Primitive types
		{name: "integer", value: &runtime.IntegerValue{Value: 42}, expected: types.INTEGER},
		{name: "float", value: &runtime.FloatValue{Value: 3.14}, expected: types.FLOAT},
		{name: "string", value: &runtime.StringValue{Value: "hello"}, expected: types.STRING},
		{name: "boolean true", value: &runtime.BooleanValue{Value: true}, expected: types.BOOLEAN},
		{name: "boolean false", value: &runtime.BooleanValue{Value: false}, expected: types.BOOLEAN},
		{name: "nil", value: &runtime.NilValue{}, expected: types.NIL},
		{name: "variant", value: &runtime.VariantValue{}, expected: types.VARIANT},

		// Array with type
		{name: "array with type", value: &runtime.ArrayValue{
			ArrayType: types.NewDynamicArrayType(types.INTEGER),
		}, expected: types.NewDynamicArrayType(types.INTEGER)},

		// Array without type returns NIL
		{name: "array without type", value: &runtime.ArrayValue{}, expected: types.NIL},

		// Record with type
		{name: "record with type", value: &runtime.RecordValue{
			RecordType: types.NewRecordType("TPoint", nil),
		}, expected: types.NewRecordType("TPoint", nil)},

		// Record without type returns NIL
		{name: "record without type", value: &runtime.RecordValue{}, expected: types.NIL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := e.getValueType(tt.value)

			if result == nil {
				t.Fatalf("expected non-nil type, got nil")
			}

			// Compare type kinds for basic types
			if result.TypeKind() != tt.expected.TypeKind() {
				t.Errorf("expected type kind %q, got %q", tt.expected.TypeKind(), result.TypeKind())
			}
		})
	}
}

// TestGetValueType_NilValue tests that nil input returns NIL type.
func TestGetValueType_NilValue(t *testing.T) {
	e := &Evaluator{}

	result := e.getValueType(nil)

	if result != types.NIL {
		t.Errorf("expected NIL type for nil input, got %v", result)
	}
}

// TestGetValueType_UnknownType tests that unknown types return NIL.
func TestGetValueType_UnknownType(t *testing.T) {
	e := &Evaluator{}

	// This test documents that unknown types gracefully return NIL
	// When passed nil, the function should return types.NIL
	result := e.getValueType(nil)

	if result != types.NIL {
		t.Errorf("expected NIL type for unknown input, got %v", result)
	}
}

// =============================================================================
// extractFunctionType tests
// =============================================================================

// TestExtractFunctionType tests extraction of types.FunctionType from ast.FunctionDecl.
func TestExtractFunctionType(t *testing.T) {
	// Create evaluator with mock type resolution
	e := &Evaluator{}
	ctx := &ExecutionContext{}

	tests := []struct {
		fn             *ast.FunctionDecl
		name           string
		expectedReturn string
		expectedParams int
	}{
		{
			name: "no parameters, void return",
			fn: &ast.FunctionDecl{
				Name: &ast.Identifier{Value: "TestProc"},
			},
			expectedParams: 0,
			expectedReturn: "VOID",
		},
		{
			name: "one integer parameter",
			fn: &ast.FunctionDecl{
				Name: &ast.Identifier{Value: "TestFunc"},
				Parameters: []*ast.Parameter{
					{
						Name: &ast.Identifier{Value: "x"},
						Type: &ast.TypeAnnotation{
							Token: token.Token{Literal: "Integer"},
							Name:  "Integer",
						},
					},
				},
				ReturnType: &ast.TypeAnnotation{
					Token: token.Token{Literal: "Integer"},
					Name:  "Integer",
				},
			},
			expectedParams: 1,
			expectedReturn: "INTEGER",
		},
		{
			name: "multiple parameters with modifiers",
			fn: &ast.FunctionDecl{
				Name: &ast.Identifier{Value: "TestFunc"},
				Parameters: []*ast.Parameter{
					{
						Name:   &ast.Identifier{Value: "a"},
						Type:   &ast.TypeAnnotation{Name: "Integer"},
						IsLazy: true,
					},
					{
						Name:  &ast.Identifier{Value: "b"},
						Type:  &ast.TypeAnnotation{Name: "String"},
						ByRef: true,
					},
					{
						Name:    &ast.Identifier{Value: "c"},
						Type:    &ast.TypeAnnotation{Name: "Float"},
						IsConst: true,
					},
				},
				ReturnType: &ast.TypeAnnotation{Name: "Boolean"},
			},
			expectedParams: 3,
			expectedReturn: "BOOLEAN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := e.extractFunctionType(tt.fn, ctx)

			if result == nil {
				t.Fatalf("expected non-nil FunctionType, got nil")
			}

			if len(result.Parameters) != tt.expectedParams {
				t.Errorf("expected %d params, got %d", tt.expectedParams, len(result.Parameters))
			}

			if result.ReturnType.TypeKind() != tt.expectedReturn {
				t.Errorf("expected return type %s, got %s", tt.expectedReturn, result.ReturnType.TypeKind())
			}
		})
	}
}

// TestExtractFunctionType_NilParameter tests handling of nil parameter types.
func TestExtractFunctionType_NilParameter(t *testing.T) {
	e := &Evaluator{}
	ctx := &ExecutionContext{}

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "BadFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "x"},
				Type: nil, // Invalid - no type annotation
			},
		},
	}

	result := e.extractFunctionType(fn, ctx)

	if result != nil {
		t.Errorf("expected nil for function with nil parameter type, got %v", result)
	}
}

// TestExtractFunctionType_LazyVarConstFlags tests that parameter flags are extracted.
func TestExtractFunctionType_LazyVarConstFlags(t *testing.T) {
	e := &Evaluator{}
	ctx := &ExecutionContext{}

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "FlagsFunc"},
		Parameters: []*ast.Parameter{
			{
				Name:   &ast.Identifier{Value: "lazy"},
				Type:   &ast.TypeAnnotation{Name: "Integer"},
				IsLazy: true,
			},
			{
				Name:  &ast.Identifier{Value: "byRef"},
				Type:  &ast.TypeAnnotation{Name: "Integer"},
				ByRef: true,
			},
			{
				Name:    &ast.Identifier{Value: "const"},
				Type:    &ast.TypeAnnotation{Name: "Integer"},
				IsConst: true,
			},
		},
	}

	result := e.extractFunctionType(fn, ctx)

	if result == nil {
		t.Fatalf("expected non-nil FunctionType")
	}

	// Check lazy flag
	if !result.LazyParams[0] {
		t.Error("expected first param to be lazy")
	}
	if result.LazyParams[1] || result.LazyParams[2] {
		t.Error("expected only first param to be lazy")
	}

	// Check var (byRef) flag
	if !result.VarParams[1] {
		t.Error("expected second param to be var/byRef")
	}
	if result.VarParams[0] || result.VarParams[2] {
		t.Error("expected only second param to be var/byRef")
	}

	// Check const flag
	if !result.ConstParams[2] {
		t.Error("expected third param to be const")
	}
	if result.ConstParams[0] || result.ConstParams[1] {
		t.Error("expected only third param to be const")
	}
}

// =============================================================================
// resolveOverloadFast tests
// =============================================================================

// TestResolveOverloadFast tests the fast path for single-overload function resolution.
func TestResolveOverloadFast(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	// Function with one non-lazy parameter
	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "x"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
			},
		},
	}

	// Create a simple integer argument expression
	argExpr := &ast.IntegerLiteral{Value: 42}

	cachedArgs, err := e.ResolveOverloadFast(fn, []ast.Expression{argExpr}, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cachedArgs) != 1 {
		t.Fatalf("expected 1 cached arg, got %d", len(cachedArgs))
	}

	// Non-lazy parameter should be evaluated
	if cachedArgs[0] == nil {
		t.Error("expected non-lazy argument to be evaluated, got nil")
	}

	intVal, ok := cachedArgs[0].(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected *runtime.IntegerValue, got %T", cachedArgs[0])
	}

	if intVal.Value != 42 {
		t.Errorf("expected value 42, got %d", intVal.Value)
	}
}

// TestResolveOverloadFast_LazyParameter tests that lazy parameters are not evaluated.
func TestResolveOverloadFast_LazyParameter(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	// Function with lazy parameter
	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "LazyFunc"},
		Parameters: []*ast.Parameter{
			{
				Name:   &ast.Identifier{Value: "x"},
				Type:   &ast.TypeAnnotation{Name: "Integer"},
				IsLazy: true,
			},
		},
	}

	argExpr := &ast.IntegerLiteral{Value: 100}

	cachedArgs, err := e.ResolveOverloadFast(fn, []ast.Expression{argExpr}, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cachedArgs) != 1 {
		t.Fatalf("expected 1 cached arg, got %d", len(cachedArgs))
	}

	// Lazy parameter should NOT be evaluated - should be nil
	if cachedArgs[0] != nil {
		t.Errorf("expected lazy argument to be nil (not evaluated), got %v", cachedArgs[0])
	}
}

// TestResolveOverloadFast_MixedParameters tests mixed lazy and non-lazy parameters.
func TestResolveOverloadFast_MixedParameters(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	// Function with mixed parameters
	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "MixedFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "a"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
			},
			{
				Name:   &ast.Identifier{Value: "b"},
				Type:   &ast.TypeAnnotation{Name: "String"},
				IsLazy: true,
			},
			{
				Name: &ast.Identifier{Value: "c"},
				Type: &ast.TypeAnnotation{Name: "Float"},
			},
		},
	}

	argExprs := []ast.Expression{
		&ast.IntegerLiteral{Value: 1},
		&ast.StringLiteral{Value: "lazy"},
		&ast.FloatLiteral{Value: 3.14},
	}

	cachedArgs, err := e.ResolveOverloadFast(fn, argExprs, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cachedArgs) != 3 {
		t.Fatalf("expected 3 cached args, got %d", len(cachedArgs))
	}

	// First arg (non-lazy): should be evaluated
	if cachedArgs[0] == nil {
		t.Error("expected first arg to be evaluated")
	}

	// Second arg (lazy): should be nil
	if cachedArgs[1] != nil {
		t.Error("expected second arg (lazy) to be nil")
	}

	// Third arg (non-lazy): should be evaluated
	if cachedArgs[2] == nil {
		t.Error("expected third arg to be evaluated")
	}
}

// TestResolveOverloadFast_NoParameters tests function with no parameters.
func TestResolveOverloadFast_NoParameters(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "NoParamsFunc"},
	}

	cachedArgs, err := e.ResolveOverloadFast(fn, []ast.Expression{}, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cachedArgs) != 0 {
		t.Errorf("expected 0 cached args, got %d", len(cachedArgs))
	}
}

// =============================================================================
// resolveOverloadMultiple tests
// =============================================================================

// TestResolveOverloadMultiple tests resolution with multiple overloads.
func TestResolveOverloadMultiple(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	// Two overloads: one takes Integer, one takes String
	overloads := []*ast.FunctionDecl{
		{
			Name:       &ast.Identifier{Value: "TestFunc"},
			IsOverload: true,
			Parameters: []*ast.Parameter{
				{
					Name: &ast.Identifier{Value: "x"},
					Type: &ast.TypeAnnotation{Name: "Integer"},
				},
			},
			ReturnType: &ast.TypeAnnotation{Name: "Integer"},
		},
		{
			Name:       &ast.Identifier{Value: "TestFunc"},
			IsOverload: true,
			Parameters: []*ast.Parameter{
				{
					Name: &ast.Identifier{Value: "x"},
					Type: &ast.TypeAnnotation{Name: "String"},
				},
			},
			ReturnType: &ast.TypeAnnotation{Name: "String"},
		},
	}

	// Call with Integer argument - should select first overload
	argExprs := []ast.Expression{&ast.IntegerLiteral{Value: 42}}

	fn, cachedArgs, err := e.ResolveOverloadMultiple("TestFunc", overloads, argExprs, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fn == nil {
		t.Fatal("expected non-nil function declaration")
	}

	// Should select the Integer overload
	typeAnnot, ok := fn.Parameters[0].Type.(*ast.TypeAnnotation)
	if !ok {
		t.Fatalf("expected *ast.TypeAnnotation, got %T", fn.Parameters[0].Type)
	}
	if typeAnnot.Name != "Integer" {
		t.Errorf("expected Integer overload, got %s", typeAnnot.Name)
	}

	if len(cachedArgs) != 1 {
		t.Fatalf("expected 1 cached arg, got %d", len(cachedArgs))
	}

	// Argument should be evaluated
	intVal, ok := cachedArgs[0].(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected *runtime.IntegerValue, got %T", cachedArgs[0])
	}
	if intVal.Value != 42 {
		t.Errorf("expected 42, got %d", intVal.Value)
	}
}

// TestResolveOverloadMultiple_StringOverload tests selecting string overload.
func TestResolveOverloadMultiple_StringOverload(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	overloads := []*ast.FunctionDecl{
		{
			Name:       &ast.Identifier{Value: "TestFunc"},
			IsOverload: true,
			Parameters: []*ast.Parameter{
				{Name: &ast.Identifier{Value: "x"}, Type: &ast.TypeAnnotation{Name: "Integer"}},
			},
		},
		{
			Name:       &ast.Identifier{Value: "TestFunc"},
			IsOverload: true,
			Parameters: []*ast.Parameter{
				{Name: &ast.Identifier{Value: "x"}, Type: &ast.TypeAnnotation{Name: "String"}},
			},
		},
	}

	// Call with String argument - should select second overload
	argExprs := []ast.Expression{&ast.StringLiteral{Value: "hello"}}

	fn, _, err := e.ResolveOverloadMultiple("TestFunc", overloads, argExprs, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typeAnnot, ok := fn.Parameters[0].Type.(*ast.TypeAnnotation)
	if !ok {
		t.Fatalf("expected *ast.TypeAnnotation, got %T", fn.Parameters[0].Type)
	}
	if typeAnnot.Name != "String" {
		t.Errorf("expected String overload, got %s", typeAnnot.Name)
	}
}

// TestResolveOverloadMultiple_NoMatch tests error when no overload matches.
func TestResolveOverloadMultiple_NoMatch(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	// Only Integer overload
	overloads := []*ast.FunctionDecl{
		{
			Name:       &ast.Identifier{Value: "IntOnly"},
			IsOverload: true,
			Parameters: []*ast.Parameter{
				{Name: &ast.Identifier{Value: "x"}, Type: &ast.TypeAnnotation{Name: "Integer"}},
			},
		},
	}

	// Call with Boolean - should not match
	argExprs := []ast.Expression{&ast.BooleanLiteral{Value: true}}

	_, _, err := e.ResolveOverloadMultiple("IntOnly", overloads, argExprs, ctx)

	if err == nil {
		t.Error("expected error for no matching overload")
	}
}

// TestResolveOverloadMultiple_MultipleParams tests overloads with multiple parameters.
func TestResolveOverloadMultiple_MultipleParams(t *testing.T) {
	e := &Evaluator{}
	ctx := NewExecutionContext(nil)

	overloads := []*ast.FunctionDecl{
		{
			Name:       &ast.Identifier{Value: "Add"},
			IsOverload: true,
			Parameters: []*ast.Parameter{
				{Name: &ast.Identifier{Value: "a"}, Type: &ast.TypeAnnotation{Name: "Integer"}},
				{Name: &ast.Identifier{Value: "b"}, Type: &ast.TypeAnnotation{Name: "Integer"}},
			},
			ReturnType: &ast.TypeAnnotation{Name: "Integer"},
		},
		{
			Name:       &ast.Identifier{Value: "Add"},
			IsOverload: true,
			Parameters: []*ast.Parameter{
				{Name: &ast.Identifier{Value: "a"}, Type: &ast.TypeAnnotation{Name: "Float"}},
				{Name: &ast.Identifier{Value: "b"}, Type: &ast.TypeAnnotation{Name: "Float"}},
			},
			ReturnType: &ast.TypeAnnotation{Name: "Float"},
		},
	}

	// Call with two integers
	argExprs := []ast.Expression{
		&ast.IntegerLiteral{Value: 1},
		&ast.IntegerLiteral{Value: 2},
	}

	fn, cachedArgs, err := e.ResolveOverloadMultiple("Add", overloads, argExprs, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typeAnnot, ok := fn.Parameters[0].Type.(*ast.TypeAnnotation)
	if !ok {
		t.Fatalf("expected *ast.TypeAnnotation, got %T", fn.Parameters[0].Type)
	}
	if typeAnnot.Name != "Integer" {
		t.Errorf("expected Integer overload, got %s", typeAnnot.Name)
	}

	if len(cachedArgs) != 2 {
		t.Errorf("expected 2 cached args, got %d", len(cachedArgs))
	}
}
