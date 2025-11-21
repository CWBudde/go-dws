package bytecode

import (
	"fmt"
	"testing"
)

// TestEvaluateBinaryComparisonFloats tests comparison folding for float values
func TestEvaluateBinaryComparisonFloats(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		left     Value
		right    Value
		wantVal  Value
		wantOk   bool
	}{
		{
			name:    "float less than true",
			op:      "<",
			left:    FloatValue(3.5),
			right:   FloatValue(5.5),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "float less than false",
			op:      "<",
			left:    FloatValue(5.5),
			right:   FloatValue(3.5),
			wantVal: BoolValue(false),
			wantOk:  true,
		},
		{
			name:    "float less or equal true (less)",
			op:      "<=",
			left:    FloatValue(3.5),
			right:   FloatValue(5.5),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "float less or equal true (equal)",
			op:      "<=",
			left:    FloatValue(5.5),
			right:   FloatValue(5.5),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "float less or equal false",
			op:      "<=",
			left:    FloatValue(5.5),
			right:   FloatValue(3.5),
			wantVal: BoolValue(false),
			wantOk:  true,
		},
		{
			name:    "float greater than true",
			op:      ">",
			left:    FloatValue(5.5),
			right:   FloatValue(3.5),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "float greater than false",
			op:      ">",
			left:    FloatValue(3.5),
			right:   FloatValue(5.5),
			wantVal: BoolValue(false),
			wantOk:  true,
		},
		{
			name:    "float greater or equal true (greater)",
			op:      ">=",
			left:    FloatValue(5.5),
			right:   FloatValue(3.5),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "float greater or equal true (equal)",
			op:      ">=",
			left:    FloatValue(5.5),
			right:   FloatValue(5.5),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "float greater or equal false",
			op:      ">=",
			left:    FloatValue(3.5),
			right:   FloatValue(5.5),
			wantVal: BoolValue(false),
			wantOk:  true,
		},
		{
			name:    "int less than float (mixed)",
			op:      "<",
			left:    IntValue(3),
			right:   FloatValue(5.5),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "float greater than int (mixed)",
			op:      ">",
			left:    FloatValue(5.5),
			right:   IntValue(3),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := evaluateBinaryComparison(tt.op, tt.left, tt.right)
			if gotOk != tt.wantOk {
				t.Errorf("evaluateBinaryComparison() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk && gotVal != tt.wantVal {
				t.Errorf("evaluateBinaryComparison() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

// TestEvaluateBinaryComparisonStrings tests comparison folding for string values
func TestEvaluateBinaryComparisonStrings(t *testing.T) {
	tests := []struct {
		name    string
		op      string
		left    Value
		right   Value
		wantVal Value
		wantOk  bool
	}{
		{
			name:    "string less than true",
			op:      "<",
			left:    StringValue("abc"),
			right:   StringValue("def"),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "string less than false",
			op:      "<",
			left:    StringValue("def"),
			right:   StringValue("abc"),
			wantVal: BoolValue(false),
			wantOk:  true,
		},
		{
			name:    "string less or equal true (less)",
			op:      "<=",
			left:    StringValue("abc"),
			right:   StringValue("def"),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "string less or equal true (equal)",
			op:      "<=",
			left:    StringValue("abc"),
			right:   StringValue("abc"),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "string less or equal false",
			op:      "<=",
			left:    StringValue("def"),
			right:   StringValue("abc"),
			wantVal: BoolValue(false),
			wantOk:  true,
		},
		{
			name:    "string greater than true",
			op:      ">",
			left:    StringValue("def"),
			right:   StringValue("abc"),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "string greater than false",
			op:      ">",
			left:    StringValue("abc"),
			right:   StringValue("def"),
			wantVal: BoolValue(false),
			wantOk:  true,
		},
		{
			name:    "string greater or equal true (greater)",
			op:      ">=",
			left:    StringValue("def"),
			right:   StringValue("abc"),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "string greater or equal true (equal)",
			op:      ">=",
			left:    StringValue("abc"),
			right:   StringValue("abc"),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "string greater or equal false",
			op:      ">=",
			left:    StringValue("abc"),
			right:   StringValue("def"),
			wantVal: BoolValue(false),
			wantOk:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := evaluateBinaryComparison(tt.op, tt.left, tt.right)
			if gotOk != tt.wantOk {
				t.Errorf("evaluateBinaryComparison() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk && gotVal != tt.wantVal {
				t.Errorf("evaluateBinaryComparison() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

// TestEvaluateBinaryComparisonEqualityOperators tests = and <> operators
func TestEvaluateBinaryComparisonEqualityOperators(t *testing.T) {
	tests := []struct {
		name    string
		op      string
		left    Value
		right   Value
		wantVal Value
		wantOk  bool
	}{
		{
			name:    "equals operator true",
			op:      "=",
			left:    IntValue(5),
			right:   IntValue(5),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "equals operator false",
			op:      "=",
			left:    IntValue(5),
			right:   IntValue(10),
			wantVal: BoolValue(false),
			wantOk:  true,
		},
		{
			name:    "not equals operator true",
			op:      "<>",
			left:    IntValue(5),
			right:   IntValue(10),
			wantVal: BoolValue(true),
			wantOk:  true,
		},
		{
			name:    "not equals operator false",
			op:      "<>",
			left:    IntValue(5),
			right:   IntValue(5),
			wantVal: BoolValue(false),
			wantOk:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := evaluateBinaryComparison(tt.op, tt.left, tt.right)
			if gotOk != tt.wantOk {
				t.Errorf("evaluateBinaryComparison() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk && gotVal != tt.wantVal {
				t.Errorf("evaluateBinaryComparison() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

// TestInvokeMethodArrayAddDirect tests Array.Add method directly
func TestInvokeMethodArrayAddDirect(t *testing.T) {
	vm := &VM{}
	vm.stack = make([]Value, 0, 256)

	arr := NewArrayInstance(nil)
	arr.Resize(3)
	arr.Set(0, IntValue(1))
	arr.Set(1, IntValue(2))
	arr.Set(2, IntValue(3))

	err := vm.invokeMethod(ArrayValue(arr), "Add", []Value{IntValue(4)})
	if err != nil {
		t.Errorf("invokeMethod(Add) error = %v", err)
	}

	if arr.Length() != 4 {
		t.Errorf("Array length after Add = %d, want 4", arr.Length())
	}

	val, _ := arr.Get(3)
	if val.AsInt() != 4 {
		t.Errorf("Array[3] = %d, want 4", val.AsInt())
	}
}

// TestInvokeMethodArrayDeleteDirect tests Array.Delete method directly
func TestInvokeMethodArrayDeleteDirect(t *testing.T) {
	vm := &VM{}
	vm.stack = make([]Value, 0, 256)

	arr := NewArrayInstance(nil)
	arr.Resize(5)
	for i := 0; i < 5; i++ {
		arr.Set(i, IntValue(int64((i+1)*10)))
	}

	err := vm.invokeMethod(ArrayValue(arr), "Delete", []Value{IntValue(2)})
	if err != nil {
		t.Errorf("invokeMethod(Delete) error = %v", err)
	}

	if arr.Length() != 4 {
		t.Errorf("Array length after Delete = %d, want 4", arr.Length())
	}
}

// TestInvokeMethodArrayDeleteWithCountDirect tests Array.Delete with count
func TestInvokeMethodArrayDeleteWithCountDirect(t *testing.T) {
	vm := &VM{}
	vm.stack = make([]Value, 0, 256)

	arr := NewArrayInstance(nil)
	arr.Resize(5)
	for i := 0; i < 5; i++ {
		arr.Set(i, IntValue(int64((i+1)*10)))
	}

	err := vm.invokeMethod(ArrayValue(arr), "Delete", []Value{IntValue(1), IntValue(2)})
	if err != nil {
		t.Errorf("invokeMethod(Delete with count) error = %v", err)
	}

	if arr.Length() != 3 {
		t.Errorf("Array length after Delete = %d, want 3", arr.Length())
	}
}

// TestInvokeMethodArrayIndexOfDirect tests Array.IndexOf method directly
func TestInvokeMethodArrayIndexOfDirect(t *testing.T) {
	vm := &VM{}
	vm.stack = make([]Value, 0, 256)

	arr := NewArrayInstance(nil)
	arr.Resize(5)
	for i := 0; i < 5; i++ {
		arr.Set(i, IntValue(int64((i+1)*10)))
	}
	arr.Set(3, IntValue(20)) // Duplicate value at index 3

	err := vm.invokeMethod(ArrayValue(arr), "IndexOf", []Value{IntValue(20)})
	if err != nil {
		t.Errorf("invokeMethod(IndexOf) error = %v", err)
	}

	result, _ := vm.pop()
	if result.AsInt() != 1 {
		t.Errorf("IndexOf(20) = %d, want 1", result.AsInt())
	}
}

// TestInvokeMethodArrayIndexOfWithStartDirect tests Array.IndexOf with start index
func TestInvokeMethodArrayIndexOfWithStartDirect(t *testing.T) {
	vm := &VM{}
	vm.stack = make([]Value, 0, 256)

	arr := NewArrayInstance(nil)
	arr.Resize(5)
	for i := 0; i < 5; i++ {
		arr.Set(i, IntValue(int64((i+1)*10)))
	}
	arr.Set(3, IntValue(20)) // Duplicate value at index 3

	err := vm.invokeMethod(ArrayValue(arr), "IndexOf", []Value{IntValue(20), IntValue(2)})
	if err != nil {
		t.Errorf("invokeMethod(IndexOf with start) error = %v", err)
	}

	result, _ := vm.pop()
	if result.AsInt() != 3 {
		t.Errorf("IndexOf(20, 2) = %d, want 3", result.AsInt())
	}
}

// TestInvokeMethodArrayDeleteEdgeCases tests edge cases for Array.Delete
func TestInvokeMethodArrayDeleteEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		initial   []Value
		index     int64
		count     int64
		wantLen   int
		wantFirst int64 // expected first element after delete
	}{
		{
			name:      "delete first element",
			initial:   []Value{IntValue(1), IntValue(2), IntValue(3), IntValue(4)},
			index:     0,
			count:     1,
			wantLen:   3,
			wantFirst: 2,
		},
		{
			name:      "delete last element",
			initial:   []Value{IntValue(1), IntValue(2), IntValue(3), IntValue(4)},
			index:     3,
			count:     1,
			wantLen:   3,
			wantFirst: 1,
		},
		{
			name:      "delete multiple elements",
			initial:   []Value{IntValue(1), IntValue(2), IntValue(3), IntValue(4), IntValue(5)},
			index:     1,
			count:     3,
			wantLen:   2,
			wantFirst: 1,
		},
		{
			name:      "delete all elements",
			initial:   []Value{IntValue(1), IntValue(2), IntValue(3)},
			index:     0,
			count:     3,
			wantLen:   0,
			wantFirst: 0, // doesn't matter, array is empty
		},
		{
			name:      "delete with count overflow",
			initial:   []Value{IntValue(1), IntValue(2), IntValue(3), IntValue(4)},
			index:     2,
			count:     10, // more than remaining
			wantLen:   2,
			wantFirst: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VM{}
			vm.stack = make([]Value, 0, 256)

			arr := NewArrayInstance(nil)
			arr.Resize(len(tt.initial))
			for i, v := range tt.initial {
				arr.Set(i, v)
			}

			err := vm.invokeMethod(ArrayValue(arr), "Delete", []Value{IntValue(tt.index), IntValue(tt.count)})
			if err != nil {
				t.Errorf("invokeMethod(Delete) error = %v", err)
			}

			if arr.Length() != tt.wantLen {
				t.Errorf("Array length after Delete = %d, want %d", arr.Length(), tt.wantLen)
			}

			if tt.wantLen > 0 {
				val0, _ := arr.Get(0)
				if val0.AsInt() != tt.wantFirst {
					t.Errorf("First element after Delete = %d, want %d", val0.AsInt(), tt.wantFirst)
				}
			}
		})
	}
}

// TestInvokeMethodArrayIndexOfEdgeCases tests edge cases for Array.IndexOf
func TestInvokeMethodArrayIndexOfEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		initial    []Value
		searchVal  Value
		startIndex int64
		useStart   bool
		wantIndex  int64
	}{
		{
			name:       "find at start",
			initial:    []Value{IntValue(10), IntValue(20), IntValue(30)},
			searchVal:  IntValue(10),
			startIndex: 0,
			useStart:   false,
			wantIndex:  0,
		},
		{
			name:       "find at end",
			initial:    []Value{IntValue(10), IntValue(20), IntValue(30)},
			searchVal:  IntValue(30),
			startIndex: 0,
			useStart:   false,
			wantIndex:  2,
		},
		{
			name:       "not found",
			initial:    []Value{IntValue(10), IntValue(20), IntValue(30)},
			searchVal:  IntValue(99),
			startIndex: 0,
			useStart:   false,
			wantIndex:  -1,
		},
		{
			name:       "start from middle",
			initial:    []Value{IntValue(10), IntValue(20), IntValue(10)},
			searchVal:  IntValue(10),
			startIndex: 1,
			useStart:   true,
			wantIndex:  2, // should skip first occurrence
		},
		{
			name:       "empty array",
			initial:    []Value{},
			searchVal:  IntValue(10),
			startIndex: 0,
			useStart:   false,
			wantIndex:  -1,
		},
		{
			name:       "string search",
			initial:    []Value{StringValue("hello"), StringValue("world"), StringValue("test")},
			searchVal:  StringValue("world"),
			startIndex: 0,
			useStart:   false,
			wantIndex:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VM{}
			vm.stack = make([]Value, 0, 256)

			arr := NewArrayInstance(nil)
			arr.Resize(len(tt.initial))
			for i, v := range tt.initial {
				arr.Set(i, v)
			}

			var err error
			if tt.useStart {
				err = vm.invokeMethod(ArrayValue(arr), "IndexOf", []Value{tt.searchVal, IntValue(tt.startIndex)})
			} else {
				err = vm.invokeMethod(ArrayValue(arr), "IndexOf", []Value{tt.searchVal})
			}

			if err != nil {
				t.Errorf("invokeMethod(IndexOf) error = %v", err)
			}

			result, _ := vm.pop()
			if result.AsInt() != tt.wantIndex {
				t.Errorf("IndexOf result = %d, want %d", result.AsInt(), tt.wantIndex)
			}
		})
	}
}

// TestInvokeMethodArraySetLengthEdgeCases tests edge cases for Array.SetLength
func TestInvokeMethodArraySetLengthEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		initial   []Value
		newLength int64
		wantLen   int
	}{
		{
			name:      "grow array",
			initial:   []Value{IntValue(1), IntValue(2)},
			newLength: 5,
			wantLen:   5,
		},
		{
			name:      "shrink array",
			initial:   []Value{IntValue(1), IntValue(2), IntValue(3), IntValue(4)},
			newLength: 2,
			wantLen:   2,
		},
		{
			name:      "set to zero",
			initial:   []Value{IntValue(1), IntValue(2), IntValue(3)},
			newLength: 0,
			wantLen:   0,
		},
		{
			name:      "same length",
			initial:   []Value{IntValue(1), IntValue(2), IntValue(3)},
			newLength: 3,
			wantLen:   3,
		},
		{
			name:      "from empty to non-empty",
			initial:   []Value{},
			newLength: 3,
			wantLen:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VM{}
			vm.stack = make([]Value, 0, 256)

			arr := NewArrayInstance(nil)
			arr.Resize(len(tt.initial))
			for i, v := range tt.initial {
				arr.Set(i, v)
			}

			err := vm.invokeMethod(ArrayValue(arr), "SetLength", []Value{IntValue(tt.newLength)})
			if err != nil {
				t.Errorf("invokeMethod(SetLength) error = %v", err)
			}

			if arr.Length() != tt.wantLen {
				t.Errorf("Array length after SetLength = %d, want %d", arr.Length(), tt.wantLen)
			}

			// Verify that shrinking preserves existing elements
			if tt.wantLen > 0 && len(tt.initial) > 0 && tt.wantLen <= len(tt.initial) {
				val0, _ := arr.Get(0)
				if val0.AsInt() != tt.initial[0].AsInt() {
					t.Errorf("First element after SetLength = %d, want %d", val0.AsInt(), tt.initial[0].AsInt())
				}
			}
		})
	}
}

// TestInvokeMethodArrayAddMultiple tests adding multiple elements
func TestInvokeMethodArrayAddMultiple(t *testing.T) {
	vm := &VM{}
	vm.stack = make([]Value, 0, 256)

	arr := NewArrayInstance(nil)
	arr.Resize(0)

	// Add 10 elements
	for i := 0; i < 10; i++ {
		err := vm.invokeMethod(ArrayValue(arr), "Add", []Value{IntValue(int64(i * 10))})
		if err != nil {
			t.Errorf("invokeMethod(Add) iteration %d error = %v", i, err)
		}
	}

	if arr.Length() != 10 {
		t.Errorf("Array length after 10 Adds = %d, want 10", arr.Length())
	}

	// Verify all elements
	for i := 0; i < 10; i++ {
		val, _ := arr.Get(i)
		expected := int64(i * 10)
		if val.AsInt() != expected {
			t.Errorf("arr[%d] = %d, want %d", i, val.AsInt(), expected)
		}
	}
}

// TestInvokeMethodArrayDeleteSingleArg tests Delete with single argument (count=1)
func TestInvokeMethodArrayDeleteSingleArg(t *testing.T) {
	vm := &VM{}
	vm.stack = make([]Value, 0, 256)

	arr := NewArrayInstance(nil)
	arr.Resize(5)
	for i := 0; i < 5; i++ {
		arr.Set(i, IntValue(int64(i+1)))
	}

	// Delete with single argument should delete 1 element
	err := vm.invokeMethod(ArrayValue(arr), "Delete", []Value{IntValue(2)})
	if err != nil {
		t.Errorf("invokeMethod(Delete, single arg) error = %v", err)
	}

	if arr.Length() != 4 {
		t.Errorf("Array length after Delete(2) = %d, want 4", arr.Length())
	}

	// Element at index 2 should now be 4 (was 3, which was deleted)
	val, _ := arr.Get(2)
	if val.AsInt() != 4 {
		t.Errorf("arr[2] after Delete = %d, want 4", val.AsInt())
	}
}

// TestInvokeMethodNilArrayError tests error handling for nil array
func TestInvokeMethodNilArrayError(t *testing.T) {
	vm := &VM{}
	vm.stack = make([]Value, 0, 256)

	err := vm.invokeMethod(NilValue(), "Add", []Value{IntValue(1)})
	if err == nil {
		t.Error("Expected error for nil array, got nil")
	}
}

// TestInvokeMethodArraySetLengthDirect tests Array.SetLength method directly
func TestInvokeMethodArraySetLengthDirect(t *testing.T) {
	vm := &VM{}
	vm.stack = make([]Value, 0, 256)

	arr := NewArrayInstance(nil)
	arr.Resize(3)

	err := vm.invokeMethod(ArrayValue(arr), "SetLength", []Value{IntValue(5)})
	if err != nil {
		t.Errorf("invokeMethod(SetLength) error = %v", err)
	}

	if arr.Length() != 5 {
		t.Errorf("Array length after SetLength = %d, want 5", arr.Length())
	}
}

// TestBinaryIntOpChecked tests binaryIntOpChecked with various cases
func TestBinaryIntOpChecked(t *testing.T) {
	tests := []struct {
		name      string
		left      Value
		right     Value
		fn        func(a, b int64) (int64, error)
		wantVal   int64
		wantError bool
	}{
		{
			name:  "successful division",
			left:  IntValue(10),
			right: IntValue(2),
			fn: func(a, b int64) (int64, error) {
				if b == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				return a / b, nil
			},
			wantVal:   5,
			wantError: false,
		},
		{
			name:  "division by zero",
			left:  IntValue(10),
			right: IntValue(0),
			fn: func(a, b int64) (int64, error) {
				if b == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				return a / b, nil
			},
			wantError: true,
		},
		{
			name:  "non-integer left operand",
			left:  FloatValue(10.5),
			right: IntValue(2),
			fn: func(a, b int64) (int64, error) {
				return a / b, nil
			},
			wantError: true,
		},
		{
			name:  "non-integer right operand",
			left:  IntValue(10),
			right: StringValue("abc"),
			fn: func(a, b int64) (int64, error) {
				return a / b, nil
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VM{}
			vm.stack = make([]Value, 0, 256)

			vm.push(tt.left)
			vm.push(tt.right)

			err := vm.binaryIntOpChecked(tt.fn)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				result, _ := vm.pop()
				if result.AsInt() != tt.wantVal {
					t.Errorf("Result = %d, want %d", result.AsInt(), tt.wantVal)
				}
			}
		})
	}
}

// TestBinaryFloatOpChecked tests binaryFloatOpChecked with various cases
func TestBinaryFloatOpChecked(t *testing.T) {
	tests := []struct {
		name      string
		left      Value
		right     Value
		fn        func(a, b float64) (float64, error)
		wantVal   float64
		wantError bool
	}{
		{
			name:  "successful division",
			left:  FloatValue(10.0),
			right: FloatValue(2.0),
			fn: func(a, b float64) (float64, error) {
				if b == 0.0 {
					return 0.0, fmt.Errorf("division by zero")
				}
				return a / b, nil
			},
			wantVal:   5.0,
			wantError: false,
		},
		{
			name:  "division by zero",
			left:  FloatValue(10.0),
			right: FloatValue(0.0),
			fn: func(a, b float64) (float64, error) {
				if b == 0.0 {
					return 0.0, fmt.Errorf("division by zero")
				}
				return a / b, nil
			},
			wantError: true,
		},
		{
			name:  "integer to float conversion",
			left:  IntValue(10),
			right: FloatValue(2.5),
			fn: func(a, b float64) (float64, error) {
				return a / b, nil
			},
			wantVal:   4.0,
			wantError: false,
		},
		{
			name:  "non-numeric left operand",
			left:  StringValue("abc"),
			right: FloatValue(2.0),
			fn: func(a, b float64) (float64, error) {
				return a / b, nil
			},
			wantError: true,
		},
		{
			name:  "non-numeric right operand",
			left:  FloatValue(10.0),
			right: StringValue("abc"),
			fn: func(a, b float64) (float64, error) {
				return a / b, nil
			},
			wantError: true,
		},
		{
			name:  "custom error from function",
			left:  FloatValue(10.0),
			right: FloatValue(-1.0),
			fn: func(a, b float64) (float64, error) {
				if b < 0 {
					return 0.0, fmt.Errorf("negative divisor not allowed")
				}
				return a / b, nil
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VM{}
			vm.stack = make([]Value, 0, 256)

			vm.push(tt.left)
			vm.push(tt.right)

			err := vm.binaryFloatOpChecked(tt.fn)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				result, _ := vm.pop()
				if result.AsFloat() != tt.wantVal {
					t.Errorf("Result = %f, want %f", result.AsFloat(), tt.wantVal)
				}
			}
		})
	}
}

// TestBinaryIntOpTypeErrors tests type errors in binaryIntOp
func TestBinaryIntOpTypeErrors(t *testing.T) {
	tests := []struct {
		name  string
		left  Value
		right Value
	}{
		{"float left", FloatValue(1.5), IntValue(2)},
		{"float right", IntValue(1), FloatValue(2.5)},
		{"string left", StringValue("abc"), IntValue(2)},
		{"string right", IntValue(1), StringValue("abc")},
		{"both non-int", FloatValue(1.5), StringValue("abc")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VM{}
			vm.stack = make([]Value, 0, 256)

			vm.push(tt.left)
			vm.push(tt.right)

			err := vm.binaryIntOp(func(a, b int64) int64 { return a + b })
			if err == nil {
				t.Error("Expected type error, got nil")
			}
		})
	}
}

// TestBinaryFloatOpTypeErrors tests type errors in binaryFloatOp
func TestBinaryFloatOpTypeErrors(t *testing.T) {
	tests := []struct {
		name  string
		left  Value
		right Value
	}{
		{"string left", StringValue("abc"), FloatValue(2.0)},
		{"string right", FloatValue(1.0), StringValue("abc")},
		{"bool left", BoolValue(true), FloatValue(2.0)},
		{"bool right", FloatValue(1.0), BoolValue(false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VM{}
			vm.stack = make([]Value, 0, 256)

			vm.push(tt.left)
			vm.push(tt.right)

			err := vm.binaryFloatOp(func(a, b float64) float64 { return a + b })
			if err == nil {
				t.Error("Expected type error, got nil")
			}
		})
	}
}

// TestInvokeMethodStringHelpers tests string helper methods
func TestInvokeMethodStringHelpers(t *testing.T) {
	tests := []struct {
		name       string
		receiver   Value
		method     string
		args       []Value
		wantVal    string
		wantInt    int64
		wantBool   bool
		checkInt   bool
		checkBool  bool
		wantError  bool
	}{
		{
			name:     "ToUpper",
			receiver: StringValue("hello"),
			method:   "ToUpper",
			args:     []Value{},
			wantVal:  "HELLO",
		},
		{
			name:     "ToLower",
			receiver: StringValue("WORLD"),
			method:   "ToLower",
			args:     []Value{},
			wantVal:  "world",
		},
		{
			name:     "ToString identity",
			receiver: StringValue("test"),
			method:   "ToString",
			args:     []Value{},
			wantVal:  "test",
		},
		{
			name:      "StartsWith true",
			receiver:  StringValue("hello world"),
			method:    "StartsWith",
			args:      []Value{StringValue("hello")},
			checkBool: true,
			wantBool:  true,
		},
		{
			name:      "EndsWith true",
			receiver:  StringValue("hello world"),
			method:    "EndsWith",
			args:      []Value{StringValue("world")},
			checkBool: true,
			wantBool:  true,
		},
		{
			name:      "Contains true",
			receiver:  StringValue("hello world"),
			method:    "Contains",
			args:      []Value{StringValue("lo wo")},
			checkBool: true,
			wantBool:  true,
		},
		{
			name:     "IndexOf found",
			receiver: StringValue("hello world"),
			method:   "IndexOf",
			args:     []Value{StringValue("world")},
			checkInt: true,
			wantInt:  7, // Position in DWScript (1-based)
		},
		{
			name:     "Copy with start only",
			receiver: StringValue("hello world"),
			method:   "Copy",
			args:     []Value{IntValue(7)},
			wantVal:  "world",
		},
		{
			name:     "Copy with start and length",
			receiver: StringValue("hello world"),
			method:   "Copy",
			args:     []Value{IntValue(1), IntValue(5)},
			wantVal:  "hello",
		},
		{
			name:     "Before",
			receiver: StringValue("hello-world"),
			method:   "Before",
			args:     []Value{StringValue("-")},
			wantVal:  "hello",
		},
		{
			name:     "After",
			receiver: StringValue("hello-world"),
			method:   "After",
			args:     []Value{StringValue("-")},
			wantVal:  "world",
		},
		{
			name:     "Length",
			receiver: StringValue("hello"),
			method:   "Length",
			args:     []Value{},
			checkInt: true,
			wantInt:  5,
		},
		{
			name:      "ToUpper with args error",
			receiver:  StringValue("hello"),
			method:    "ToUpper",
			args:      []Value{IntValue(1)},
			wantError: true,
		},
		{
			name:      "StartsWith no args error",
			receiver:  StringValue("hello"),
			method:    "StartsWith",
			args:      []Value{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VM{}
			vm.stack = make([]Value, 0, 256)

			err := vm.invokeMethod(tt.receiver, tt.method, tt.args)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				result, _ := vm.pop()
				if tt.checkInt {
					if result.AsInt() != tt.wantInt {
						t.Errorf("Result = %d, want %d", result.AsInt(), tt.wantInt)
					}
				} else if tt.checkBool {
					if result.AsBool() != tt.wantBool {
						t.Errorf("Result = %v, want %v", result.AsBool(), tt.wantBool)
					}
				} else {
					if result.AsString() != tt.wantVal {
						t.Errorf("Result = %q, want %q", result.AsString(), tt.wantVal)
					}
				}
			}
		})
	}
}

// TestInvokeMethodStringToConversions tests string-to-number conversions
func TestInvokeMethodStringToConversions(t *testing.T) {
	tests := []struct {
		name      string
		receiver  Value
		method    string
		wantInt   int64
		wantFloat float64
		checkInt  bool
		wantError bool
	}{
		{
			name:     "ToInteger valid",
			receiver: StringValue("123"),
			method:   "ToInteger",
			wantInt:  123,
			checkInt: true,
		},
		{
			name:      "ToFloat valid",
			receiver:  StringValue("3.14"),
			method:    "ToFloat",
			wantFloat: 3.14,
			checkInt:  false,
		},
		{
			name:      "ToInteger invalid",
			receiver:  StringValue("abc"),
			method:    "ToInteger",
			wantError: true,
		},
		{
			name:      "ToFloat invalid",
			receiver:  StringValue("xyz"),
			method:    "ToFloat",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VM{}
			vm.stack = make([]Value, 0, 256)

			err := vm.invokeMethod(tt.receiver, tt.method, []Value{})

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				result, _ := vm.pop()
				if tt.checkInt {
					if result.AsInt() != tt.wantInt {
						t.Errorf("Result = %d, want %d", result.AsInt(), tt.wantInt)
					}
				} else {
					if result.AsFloat() != tt.wantFloat {
						t.Errorf("Result = %f, want %f", result.AsFloat(), tt.wantFloat)
					}
				}
			}
		})
	}
}
