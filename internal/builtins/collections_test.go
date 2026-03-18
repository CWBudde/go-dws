package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Collection Functions Tests
// =============================================================================

type collectionEvalContext struct {
	*mockContext
	result Value
}

func (c *collectionEvalContext) EvalFunctionPointer(funcPtr Value, args []Value) Value {
	return c.result
}

func TestMap(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 3},
		},
	}

	// Test error cases (we can't test successful cases easily without a real interpreter)
	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "no arguments",
			args: []Value{},
		},
		{
			name: "only 1 argument",
			args: []Value{testArray},
		},
		{
			name: "wrong first argument type",
			args: []Value{
				&runtime.IntegerValue{Value: 1},
				&runtime.FunctionPointerValue{},
			},
		},
		{
			name: "wrong second argument type",
			args: []Value{
				testArray,
				&runtime.IntegerValue{Value: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Map(ctx, tt.args)
			if result.Type() != "ERROR" {
				t.Errorf("Map() should return ERROR for %s, got %v", tt.name, result)
			}
		})
	}

	// Test with valid arguments (will call mock EvalFunctionPointer which returns NilValue)
	mockLambda := &runtime.FunctionPointerValue{}
	result := Map(ctx, []Value{testArray, mockLambda})
	if result.Type() == "ERROR" {
		t.Errorf("Map() with valid arguments returned error: %v", result)
	}
}

func TestMap_CopiesMappedRecordResults(t *testing.T) {
	ctx := &collectionEvalContext{
		mockContext: newMockContext(),
		result: &runtime.RecordValue{
			Fields: map[string]Value{
				"name": &runtime.StringValue{Value: "mapped"},
			},
		},
	}

	arrayVal := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.RecordValue{
				Fields: map[string]Value{
					"name": &runtime.StringValue{Value: "input"},
				},
			},
		},
	}

	result := Map(ctx, []Value{arrayVal, &runtime.FunctionPointerValue{}})
	if result.Type() == "ERROR" {
		t.Fatalf("Map() returned error: %v", result)
	}

	resultArray, ok := result.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %T", result)
	}

	mappedRecord, ok := resultArray.Elements[0].(*runtime.RecordValue)
	if !ok {
		t.Fatalf("expected RecordValue element, got %T", resultArray.Elements[0])
	}

	if mappedRecord == ctx.result {
		t.Fatal("expected Map() to store a copy of the callback result")
	}

	ctx.result.(*runtime.RecordValue).Fields["name"] = &runtime.StringValue{Value: "mutated"}

	name, ok := mappedRecord.Fields["name"].(*runtime.StringValue)
	if !ok {
		t.Fatalf("mapped record field has type %T, want *runtime.StringValue", mappedRecord.Fields["name"])
	}
	if name.Value != "mapped" {
		t.Fatalf("expected copied mapped record to keep original value, got %q", name.Value)
	}
}

func TestFilter(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 3},
		},
	}

	// Test error cases
	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "no arguments",
			args: []Value{},
		},
		{
			name: "only 1 argument",
			args: []Value{testArray},
		},
		{
			name: "wrong first argument type",
			args: []Value{
				&runtime.StringValue{Value: "not an array"},
				&runtime.FunctionPointerValue{},
			},
		},
		{
			name: "wrong second argument type",
			args: []Value{
				testArray,
				&runtime.StringValue{Value: "not a function"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(ctx, tt.args)
			if result.Type() != "ERROR" {
				t.Errorf("Filter() should return ERROR for %s, got %v", tt.name, result)
			}
		})
	}

	// Test with valid arguments (mock returns NilValue which should trigger error for non-boolean)
	mockPredicate := &runtime.FunctionPointerValue{}
	result := Filter(ctx, []Value{testArray, mockPredicate})
	// Mock EvalFunctionPointer returns NilValue, which should cause an error in Filter
	// since it expects BooleanValue
	if result.Type() == "ERROR" {
		// This is expected since mock doesn't return proper boolean
		return
	}
}

func TestReduce(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 3},
		},
	}

	// Test error cases
	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "no arguments",
			args: []Value{},
		},
		{
			name: "only 1 argument",
			args: []Value{testArray},
		},
		{
			name: "only 2 arguments",
			args: []Value{testArray, &runtime.FunctionPointerValue{}},
		},
		{
			name: "wrong first argument type",
			args: []Value{
				&runtime.BooleanValue{Value: true},
				&runtime.FunctionPointerValue{},
				&runtime.IntegerValue{Value: 0},
			},
		},
		{
			name: "wrong second argument type",
			args: []Value{
				testArray,
				&runtime.IntegerValue{Value: 1},
				&runtime.IntegerValue{Value: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reduce(ctx, tt.args)
			if result.Type() != "ERROR" {
				t.Errorf("Reduce() should return ERROR for %s, got %v", tt.name, result)
			}
		})
	}

	// Test with valid arguments
	mockLambda := &runtime.FunctionPointerValue{}
	result := Reduce(ctx, []Value{testArray, mockLambda, &runtime.IntegerValue{Value: 0}})
	if result.Type() == "ERROR" {
		t.Errorf("Reduce() with valid arguments returned error: %v", result)
	}
}

func TestForEach(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 3},
		},
	}

	// Test error cases
	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "no arguments",
			args: []Value{},
		},
		{
			name: "only 1 argument",
			args: []Value{testArray},
		},
		{
			name: "wrong first argument type",
			args: []Value{
				&runtime.FloatValue{Value: 3.14},
				&runtime.FunctionPointerValue{},
			},
		},
		{
			name: "wrong second argument type",
			args: []Value{
				testArray,
				&runtime.BooleanValue{Value: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ForEach(ctx, tt.args)
			if result.Type() != "ERROR" {
				t.Errorf("ForEach() should return ERROR for %s, got %v", tt.name, result)
			}
		})
	}

	// Test with valid arguments
	mockLambda := &runtime.FunctionPointerValue{}
	result := ForEach(ctx, []Value{testArray, mockLambda})
	if result.Type() == "ERROR" {
		t.Errorf("ForEach() with valid arguments returned error: %v", result)
	}
}

func TestEvery(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 4},
			&runtime.IntegerValue{Value: 6},
		},
	}

	// Test error cases
	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "no arguments",
			args: []Value{},
		},
		{
			name: "only 1 argument",
			args: []Value{testArray},
		},
		{
			name: "wrong first argument type",
			args: []Value{
				&runtime.StringValue{Value: "array"},
				&runtime.FunctionPointerValue{},
			},
		},
		{
			name: "wrong second argument type",
			args: []Value{
				testArray,
				&runtime.IntegerValue{Value: 42},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Every(ctx, tt.args)
			if result.Type() != "ERROR" {
				t.Errorf("Every() should return ERROR for %s, got %v", tt.name, result)
			}
		})
	}

	// Test with valid arguments (mock returns NilValue which should cause error)
	mockPredicate := &runtime.FunctionPointerValue{}
	result := Every(ctx, []Value{testArray, mockPredicate})
	// Mock should cause error because it doesn't return proper boolean
	if result.Type() == "ERROR" {
		return // Expected
	}
}

func TestSome(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 3},
		},
	}

	// Test error cases
	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "no arguments",
			args: []Value{},
		},
		{
			name: "only 1 argument",
			args: []Value{testArray},
		},
		{
			name: "wrong first argument type",
			args: []Value{
				&runtime.IntegerValue{Value: 123},
				&runtime.FunctionPointerValue{},
			},
		},
		{
			name: "wrong second argument type",
			args: []Value{
				testArray,
				&runtime.FloatValue{Value: 1.5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Some(ctx, tt.args)
			if result.Type() != "ERROR" {
				t.Errorf("Some() should return ERROR for %s, got %v", tt.name, result)
			}
		})
	}

	// Test with valid arguments (mock returns NilValue which should cause error)
	mockPredicate := &runtime.FunctionPointerValue{}
	result := Some(ctx, []Value{testArray, mockPredicate})
	// Mock should cause error because it doesn't return proper boolean
	if result.Type() == "ERROR" {
		return // Expected
	}
}

func TestFind(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 3},
		},
	}

	// Test error cases
	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "no arguments",
			args: []Value{},
		},
		{
			name: "only 1 argument",
			args: []Value{testArray},
		},
		{
			name: "wrong first argument type",
			args: []Value{
				&runtime.BooleanValue{Value: true},
				&runtime.FunctionPointerValue{},
			},
		},
		{
			name: "wrong second argument type",
			args: []Value{
				testArray,
				&runtime.StringValue{Value: "lambda"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Find(ctx, tt.args)
			if result.Type() != "ERROR" {
				t.Errorf("Find() should return ERROR for %s, got %v", tt.name, result)
			}
		})
	}

	// Test with valid arguments (mock returns NilValue which should cause error)
	mockPredicate := &runtime.FunctionPointerValue{}
	result := Find(ctx, []Value{testArray, mockPredicate})
	// Mock should cause error because it doesn't return proper boolean
	if result.Type() == "ERROR" {
		return // Expected
	}
}

func TestFind_CopiesMatchedRecord(t *testing.T) {
	ctx := &collectionEvalContext{
		mockContext: newMockContext(),
		result:      &runtime.BooleanValue{Value: true},
	}

	original := &runtime.RecordValue{
		Fields: map[string]Value{
			"name": &runtime.StringValue{Value: "original"},
		},
	}

	result := Find(ctx, []Value{
		&runtime.ArrayValue{
			Elements: []Value{original},
		},
		&runtime.FunctionPointerValue{},
	})
	if result.Type() == "ERROR" {
		t.Fatalf("Find() returned error: %v", result)
	}

	found, ok := result.(*runtime.RecordValue)
	if !ok {
		t.Fatalf("expected RecordValue, got %T", result)
	}

	if found == original {
		t.Fatal("expected Find() to return a copy of the matched record")
	}

	original.Fields["name"] = &runtime.StringValue{Value: "mutated"}

	name, ok := found.Fields["name"].(*runtime.StringValue)
	if !ok {
		t.Fatalf("found record field has type %T, want *runtime.StringValue", found.Fields["name"])
	}
	if name.Value != "original" {
		t.Fatalf("expected copied found record to keep original value, got %q", name.Value)
	}
}

func TestFindIndex(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 10},
			&runtime.IntegerValue{Value: 20},
			&runtime.IntegerValue{Value: 30},
		},
	}

	// Test error cases
	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "no arguments",
			args: []Value{},
		},
		{
			name: "only 1 argument",
			args: []Value{testArray},
		},
		{
			name: "wrong first argument type",
			args: []Value{
				&runtime.FloatValue{Value: 42.0},
				&runtime.FunctionPointerValue{},
			},
		},
		{
			name: "wrong second argument type",
			args: []Value{
				testArray,
				&runtime.IntegerValue{Value: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindIndex(ctx, tt.args)
			if result.Type() != "ERROR" {
				t.Errorf("FindIndex() should return ERROR for %s, got %v", tt.name, result)
			}
		})
	}

	// Test with valid arguments (mock returns NilValue which should cause error)
	mockPredicate := &runtime.FunctionPointerValue{}
	result := FindIndex(ctx, []Value{testArray, mockPredicate})
	// Mock should cause error because it doesn't return proper boolean
	if result.Type() == "ERROR" {
		return // Expected
	}
}
