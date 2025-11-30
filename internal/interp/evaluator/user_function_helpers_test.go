package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// =============================================================================
// Task 3.5.144b.1: EvaluateDefaultParameters tests
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
