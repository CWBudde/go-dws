package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// =============================================================================
// Array Functions Tests
// =============================================================================

func TestArrayLength(t *testing.T) {
	ctx := newMockContext()

	// Create test array
	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 3},
		},
	}

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "array with 3 elements",
			args:     []Value{testArray},
			expected: 3,
		},
		{
			name:     "string length",
			args:     []Value{&runtime.StringValue{Value: "Hello"}},
			expected: 5,
		},
		{
			name:     "empty string",
			args:     []Value{&runtime.StringValue{Value: ""}},
			expected: 0,
		},
		{
			name:     "Unicode string",
			args:     []Value{&runtime.StringValue{Value: "こんにちは"}}, // 5 Japanese characters
			expected: 5,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Length(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("Length() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestArrayCopy(t *testing.T) {
	ctx := newMockContext()

	// Test array copy (1 argument)
	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 3},
		},
	}

	result := Copy(ctx, []Value{testArray})
	if result.Type() == "ERROR" {
		t.Errorf("Copy() with array returned error: %v", result)
	}

	// Test string copy (3 arguments)
	tests := []struct {
		name     string
		str      string
		index    int64
		count    int64
		expected string
	}{
		{
			name:     "first 3 characters",
			str:      "Hello",
			index:    1,
			count:    3,
			expected: "Hel",
		},
		{
			name:     "middle substring",
			str:      "Hello World",
			index:    7,
			count:    5,
			expected: "World",
		},
		{
			name:     "count exceeds string length",
			str:      "Hello",
			index:    1,
			count:    100,
			expected: "Hello",
		},
		{
			name:     "index out of bounds",
			str:      "Hello",
			index:    10,
			count:    5,
			expected: "",
		},
		{
			name:     "negative count",
			str:      "Hello",
			index:    1,
			count:    -1,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Copy(ctx, []Value{
				&runtime.StringValue{Value: tt.str},
				&runtime.IntegerValue{Value: tt.index},
				&runtime.IntegerValue{Value: tt.count},
			})

			if result.Type() == "ERROR" {
				t.Errorf("Copy() returned error: %v", result)
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("Copy() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}

	// Test error cases
	result = Copy(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("Copy() with 0 arguments should error")
	}

	result = Copy(ctx, []Value{&runtime.IntegerValue{Value: 1}, &runtime.IntegerValue{Value: 2}})
	if result.Type() != "ERROR" {
		t.Errorf("Copy() with 2 arguments should error")
	}
}

func TestArrayLow(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 10},
			&runtime.IntegerValue{Value: 20},
		},
	}

	result := Low(ctx, []Value{testArray})
	if result.Type() == "ERROR" {
		t.Errorf("Low() returned error: %v", result)
		return
	}

	// Just verify it returns a value (actual behavior depends on Context mock)
	if result == nil {
		t.Errorf("Low() returned nil")
	}

	// Test error: wrong argument count
	result = Low(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("Low() with 0 arguments should error")
	}
}

func TestArrayHigh(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 10},
			&runtime.IntegerValue{Value: 20},
		},
	}

	result := High(ctx, []Value{testArray})
	if result.Type() == "ERROR" {
		t.Errorf("High() returned error: %v", result)
		return
	}

	// Just verify it returns a value (actual behavior depends on Context mock)
	if result == nil {
		t.Errorf("High() returned nil")
	}

	// Test error: wrong argument count
	result = High(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("High() with 0 arguments should error")
	}
}

func TestIndexOf(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 10},
			&runtime.IntegerValue{Value: 20},
			&runtime.IntegerValue{Value: 30},
			&runtime.IntegerValue{Value: 20}, // duplicate
		},
	}

	tests := []struct {
		name       string
		array      *runtime.ArrayValue
		value      Value
		startIndex *int64
		expected   int64
		isError    bool
	}{
		{
			name:     "find first occurrence",
			array:    testArray,
			value:    &runtime.IntegerValue{Value: 20},
			expected: 1,
		},
		{
			name:       "find from index 2",
			array:      testArray,
			value:      &runtime.IntegerValue{Value: 20},
			startIndex: ptr(int64(2)),
			expected:   3,
		},
		{
			name:     "value not found",
			array:    testArray,
			value:    &runtime.IntegerValue{Value: 99},
			expected: -1,
		},
		{
			name:       "start index out of bounds",
			array:      testArray,
			value:      &runtime.IntegerValue{Value: 20},
			startIndex: ptr(int64(10)),
			expected:   -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args []Value
			if tt.startIndex == nil {
				args = []Value{tt.array, tt.value}
			} else {
				args = []Value{tt.array, tt.value, &runtime.IntegerValue{Value: *tt.startIndex}}
			}

			result := IndexOf(ctx, args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("IndexOf() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}

	// Test error cases
	result := IndexOf(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("IndexOf() with 0 arguments should error")
	}

	result = IndexOf(ctx, []Value{&runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("IndexOf() with wrong type should error")
	}
}

func TestContains(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 10},
			&runtime.IntegerValue{Value: 20},
			&runtime.IntegerValue{Value: 30},
		},
	}

	tests := []struct {
		name     string
		array    *runtime.ArrayValue
		value    Value
		expected bool
	}{
		{
			name:     "value exists",
			array:    testArray,
			value:    &runtime.IntegerValue{Value: 20},
			expected: true,
		},
		{
			name:     "value does not exist",
			array:    testArray,
			value:    &runtime.IntegerValue{Value: 99},
			expected: false,
		},
		{
			name:     "check for string in integer array",
			array:    testArray,
			value:    &runtime.StringValue{Value: "20"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(ctx, []Value{tt.array, tt.value})

			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("Contains() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}

	// Test error cases
	result := Contains(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("Contains() with 0 arguments should error")
	}

	result = Contains(ctx, []Value{&runtime.IntegerValue{Value: 1}, &runtime.IntegerValue{Value: 2}})
	if result.Type() != "ERROR" {
		t.Errorf("Contains() with wrong type should error")
	}
}

func TestReverse(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 3},
		},
	}

	result := Reverse(ctx, []Value{testArray})
	if result.Type() == "ERROR" {
		t.Errorf("Reverse() returned error: %v", result)
	}

	// Test error: wrong argument count
	result = Reverse(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("Reverse() with 0 arguments should error")
	}
}

func TestSort(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 3},
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
		},
	}

	result := Sort(ctx, []Value{testArray})
	if result.Type() == "ERROR" {
		t.Errorf("Sort() returned error: %v", result)
	}

	// Test error: wrong argument count
	result = Sort(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("Sort() with 0 arguments should error")
	}

	// Test error: wrong type
	result = Sort(ctx, []Value{&runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("Sort() with wrong type should error")
	}
}

func TestAdd(t *testing.T) {
	ctx := newMockContext()

	// Create dynamic array
	testArray := &runtime.ArrayValue{
		Elements:  []Value{&runtime.IntegerValue{Value: 1}},
		ArrayType: &types.ArrayType{}, // dynamic array
	}

	result := Add(ctx, []Value{testArray, &runtime.IntegerValue{Value: 2}})
	if result.Type() == "ERROR" {
		t.Errorf("Add() returned error: %v", result)
		return
	}

	// Verify element was added
	if len(testArray.Elements) != 2 {
		t.Errorf("Add() did not append element, length = %d, want 2", len(testArray.Elements))
	}

	// Test error: wrong argument count
	result = Add(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("Add() with 0 arguments should error")
	}

	// Test error: wrong type
	result = Add(ctx, []Value{&runtime.IntegerValue{Value: 1}, &runtime.IntegerValue{Value: 2}})
	if result.Type() != "ERROR" {
		t.Errorf("Add() with wrong type should error")
	}
}

func TestDelete(t *testing.T) {
	ctx := newMockContext()

	// Create dynamic array
	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 10},
			&runtime.IntegerValue{Value: 20},
			&runtime.IntegerValue{Value: 30},
		},
		ArrayType: &types.ArrayType{}, // dynamic array
	}

	result := Delete(ctx, []Value{testArray, &runtime.IntegerValue{Value: 1}})
	if result.Type() == "ERROR" {
		t.Errorf("Delete() returned error: %v", result)
		return
	}

	// Verify element was removed
	if len(testArray.Elements) != 2 {
		t.Errorf("Delete() did not remove element, length = %d, want 2", len(testArray.Elements))
	}

	// Test error: index out of bounds
	result = Delete(ctx, []Value{testArray, &runtime.IntegerValue{Value: 10}})
	if result.Type() != "ERROR" {
		t.Errorf("Delete() with out of bounds index should error")
	}

	// Test error: wrong argument count
	result = Delete(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("Delete() with 0 arguments should error")
	}
}

func TestSetLength(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
		},
	}

	result := SetLength(ctx, []Value{testArray, &runtime.IntegerValue{Value: 5}})
	if result.Type() == "ERROR" {
		t.Errorf("SetLength() returned error: %v", result)
	}

	// Test error: negative length
	result = SetLength(ctx, []Value{testArray, &runtime.IntegerValue{Value: -1}})
	if result.Type() != "ERROR" {
		t.Errorf("SetLength() with negative length should error")
	}

	// Test error: wrong argument count
	result = SetLength(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("SetLength() with 0 arguments should error")
	}
}

func TestConcatArrays(t *testing.T) {
	ctx := newMockContext()

	array1 := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
		},
	}

	array2 := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 3},
			&runtime.IntegerValue{Value: 4},
		},
	}

	result := ConcatArrays(ctx, []Value{array1, array2})
	if result.Type() == "ERROR" {
		t.Errorf("ConcatArrays() returned error: %v", result)
		return
	}

	resultArray, ok := result.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %T", result)
	}

	if len(resultArray.Elements) != 4 {
		t.Errorf("ConcatArrays() length = %d, want 4", len(resultArray.Elements))
	}

	// Test error: no arguments
	result = ConcatArrays(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("ConcatArrays() with 0 arguments should error")
	}

	// Test error: wrong type
	result = ConcatArrays(ctx, []Value{array1, &runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("ConcatArrays() with wrong type should error")
	}
}

func TestSlice(t *testing.T) {
	ctx := newMockContext()

	testArray := &runtime.ArrayValue{
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 3},
			&runtime.IntegerValue{Value: 4},
			&runtime.IntegerValue{Value: 5},
		},
	}

	tests := []struct {
		name          string
		start         int64
		end           int64
		expectedCount int
	}{
		{
			name:          "slice middle",
			start:         1,
			end:           4,
			expectedCount: 3,
		},
		{
			name:          "slice from beginning",
			start:         0,
			end:           2,
			expectedCount: 2,
		},
		{
			name:          "slice to end",
			start:         3,
			end:           5,
			expectedCount: 2,
		},
		{
			name:          "empty slice",
			start:         2,
			end:           2,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slice(ctx, []Value{
				testArray,
				&runtime.IntegerValue{Value: tt.start},
				&runtime.IntegerValue{Value: tt.end},
			})

			if result.Type() == "ERROR" {
				t.Errorf("Slice() returned error: %v", result)
				return
			}

			resultArray, ok := result.(*runtime.ArrayValue)
			if !ok {
				t.Fatalf("expected ArrayValue, got %T", result)
			}

			if len(resultArray.Elements) != tt.expectedCount {
				t.Errorf("Slice() length = %d, want %d", len(resultArray.Elements), tt.expectedCount)
			}
		})
	}

	// Test error: wrong argument count
	result := Slice(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("Slice() with 0 arguments should error")
	}

	// Test error: wrong type
	result = Slice(ctx, []Value{&runtime.IntegerValue{Value: 1}, &runtime.IntegerValue{Value: 0}, &runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("Slice() with wrong type should error")
	}
}

// Helper function to create pointer to int64
func ptr(i int64) *int64 {
	return &i
}
