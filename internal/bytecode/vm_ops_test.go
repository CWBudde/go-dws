package bytecode

import (
	"testing"
)

// TestValuesEqualExtended tests the valuesEqual function with additional value types
func TestValuesEqualExtended(t *testing.T) {
	vm := NewVM()

	t.Run("nil values", func(t *testing.T) {
		if !vm.valuesEqual(NilValue(), NilValue()) {
			t.Error("nil should equal nil")
		}
	})

	t.Run("integers equal", func(t *testing.T) {
		if !vm.valuesEqual(IntValue(42), IntValue(42)) {
			t.Error("equal integers should be equal")
		}
	})

	t.Run("integers not equal", func(t *testing.T) {
		if vm.valuesEqual(IntValue(42), IntValue(43)) {
			t.Error("different integers should not be equal")
		}
	})

	t.Run("floats equal", func(t *testing.T) {
		if !vm.valuesEqual(FloatValue(3.14), FloatValue(3.14)) {
			t.Error("equal floats should be equal")
		}
	})

	t.Run("floats not equal", func(t *testing.T) {
		if vm.valuesEqual(FloatValue(3.14), FloatValue(3.15)) {
			t.Error("different floats should not be equal")
		}
	})

	t.Run("strings equal", func(t *testing.T) {
		if !vm.valuesEqual(StringValue("hello"), StringValue("hello")) {
			t.Error("equal strings should be equal")
		}
	})

	t.Run("strings not equal", func(t *testing.T) {
		if vm.valuesEqual(StringValue("hello"), StringValue("world")) {
			t.Error("different strings should not be equal")
		}
	})

	t.Run("booleans equal", func(t *testing.T) {
		if !vm.valuesEqual(BoolValue(true), BoolValue(true)) {
			t.Error("equal booleans should be equal")
		}
	})

	t.Run("booleans not equal", func(t *testing.T) {
		if vm.valuesEqual(BoolValue(true), BoolValue(false)) {
			t.Error("different booleans should not be equal")
		}
	})

	t.Run("numeric types equal", func(t *testing.T) {
		// Numbers are converted to float for comparison
		if !vm.valuesEqual(IntValue(42), FloatValue(42.0)) {
			t.Error("numerically equal values should be equal regardless of type")
		}
	})

	t.Run("arrays same reference", func(t *testing.T) {
		// Array equality is pointer comparison, not deep comparison
		arr := NewArrayInstance([]Value{IntValue(1), IntValue(2), IntValue(3)})
		if !vm.valuesEqual(ArrayValue(arr), ArrayValue(arr)) {
			t.Error("same array reference should be equal")
		}
	})

	t.Run("arrays different references", func(t *testing.T) {
		// Even with same contents, different array instances are not equal
		arr1 := NewArrayInstance([]Value{IntValue(1), IntValue(2), IntValue(3)})
		arr2 := NewArrayInstance([]Value{IntValue(1), IntValue(2), IntValue(3)})
		if vm.valuesEqual(ArrayValue(arr1), ArrayValue(arr2)) {
			t.Error("different array references should not be equal (pointer comparison)")
		}
	})

	t.Run("arrays not equal - different length", func(t *testing.T) {
		arr1 := NewArrayInstance([]Value{IntValue(1), IntValue(2)})
		arr2 := NewArrayInstance([]Value{IntValue(1), IntValue(2), IntValue(3)})
		if vm.valuesEqual(ArrayValue(arr1), ArrayValue(arr2)) {
			t.Error("arrays with different lengths should not be equal")
		}
	})

	t.Run("arrays not equal - different elements", func(t *testing.T) {
		arr1 := NewArrayInstance([]Value{IntValue(1), IntValue(2), IntValue(3)})
		arr2 := NewArrayInstance([]Value{IntValue(1), IntValue(2), IntValue(4)})
		if vm.valuesEqual(ArrayValue(arr1), ArrayValue(arr2)) {
			t.Error("arrays with different elements should not be equal")
		}
	})

	t.Run("functions - same reference", func(t *testing.T) {
		fn := &FunctionObject{Name: "test", Arity: 0}
		if !vm.valuesEqual(FunctionValue(fn), FunctionValue(fn)) {
			t.Error("same function reference should be equal")
		}
	})

	t.Run("functions - different references", func(t *testing.T) {
		fn1 := &FunctionObject{Name: "test1", Arity: 0}
		fn2 := &FunctionObject{Name: "test2", Arity: 0}
		if vm.valuesEqual(FunctionValue(fn1), FunctionValue(fn2)) {
			t.Error("different function references should not be equal")
		}
	})
}

// TestRequireArray tests array type checking
func TestRequireArrayExtended(t *testing.T) {
	vm := NewVM()

	t.Run("valid array", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1), IntValue(2)})
		result, err := vm.requireArray(ArrayValue(arr), "test")
		if err != nil {
			t.Fatalf("requireArray error: %v", err)
		}
		if result.Length() != 2 {
			t.Errorf("array length = %d, want 2", result.Length())
		}
	})

	t.Run("non-array value", func(t *testing.T) {
		_, err := vm.requireArray(IntValue(42), "test")
		if err == nil {
			t.Error("requireArray should return error for non-array")
		}
	})

	t.Run("nil value", func(t *testing.T) {
		_, err := vm.requireArray(NilValue(), "test")
		if err == nil {
			t.Error("requireArray should return error for nil")
		}
	})

	t.Run("string value", func(t *testing.T) {
		_, err := vm.requireArray(StringValue("not an array"), "test")
		if err == nil {
			t.Error("requireArray should return error for string")
		}
	})
}

// TestRequireInt tests integer type checking
func TestRequireIntExtended(t *testing.T) {
	vm := NewVM()

	t.Run("valid integer", func(t *testing.T) {
		result, err := vm.requireInt(IntValue(42), "test")
		if err != nil {
			t.Fatalf("requireInt error: %v", err)
		}
		if result != 42 {
			t.Errorf("requireInt = %d, want 42", result)
		}
	})

	t.Run("non-integer value", func(t *testing.T) {
		_, err := vm.requireInt(StringValue("hello"), "test")
		if err == nil {
			t.Error("requireInt should return error for non-integer")
		}
	})

	t.Run("float value", func(t *testing.T) {
		_, err := vm.requireInt(FloatValue(3.14), "test")
		if err == nil {
			t.Error("requireInt should return error for float")
		}
	})

	t.Run("negative integer", func(t *testing.T) {
		result, err := vm.requireInt(IntValue(-100), "test")
		if err != nil {
			t.Fatalf("requireInt error: %v", err)
		}
		if result != -100 {
			t.Errorf("requireInt = %d, want -100", result)
		}
	})

	t.Run("zero integer", func(t *testing.T) {
		result, err := vm.requireInt(IntValue(0), "test")
		if err != nil {
			t.Fatalf("requireInt error: %v", err)
		}
		if result != 0 {
			t.Errorf("requireInt = %d, want 0", result)
		}
	})

	t.Run("boolean value", func(t *testing.T) {
		_, err := vm.requireInt(BoolValue(true), "test")
		if err == nil {
			t.Error("requireInt should return error for boolean")
		}
	})
}

// TestCompare tests the compare function for various comparison operations
func TestCompare(t *testing.T) {
	vm := NewVM()

	t.Run("numeric greater than", func(t *testing.T) {
		vm.push(IntValue(5))
		vm.push(IntValue(10))
		result, err := vm.compare(OpGreater)
		if err != nil {
			t.Fatalf("compare error: %v", err)
		}
		if result {
			t.Error("10 > 5 should be false")
		}
	})

	t.Run("numeric greater than true", func(t *testing.T) {
		vm.push(IntValue(10))
		vm.push(IntValue(5))
		result, err := vm.compare(OpGreater)
		if err != nil {
			t.Fatalf("compare error: %v", err)
		}
		if !result {
			t.Error("5 > 10 should be false, got true")
		}
	})

	t.Run("numeric greater or equal", func(t *testing.T) {
		vm.push(IntValue(10))
		vm.push(IntValue(10))
		result, err := vm.compare(OpGreaterEqual)
		if err != nil {
			t.Fatalf("compare error: %v", err)
		}
		if !result {
			t.Error("10 >= 10 should be true")
		}
	})

	t.Run("numeric less than", func(t *testing.T) {
		vm.push(IntValue(5))
		vm.push(IntValue(10))
		result, err := vm.compare(OpLess)
		if err != nil {
			t.Fatalf("compare error: %v", err)
		}
		if !result {
			t.Error("5 < 10 should be true")
		}
	})

	t.Run("numeric less or equal", func(t *testing.T) {
		vm.push(IntValue(5))
		vm.push(IntValue(5))
		result, err := vm.compare(OpLessEqual)
		if err != nil {
			t.Fatalf("compare error: %v", err)
		}
		if !result {
			t.Error("5 <= 5 should be true")
		}
	})

	t.Run("float comparison", func(t *testing.T) {
		vm.push(FloatValue(3.14))  // left
		vm.push(FloatValue(2.71))  // right
		result, err := vm.compare(OpGreater)
		if err != nil {
			t.Fatalf("compare error: %v", err)
		}
		if !result {
			t.Error("3.14 > 2.71 should be true")
		}
	})

	t.Run("string greater than", func(t *testing.T) {
		vm.push(StringValue("banana"))  // left
		vm.push(StringValue("apple"))   // right
		result, err := vm.compare(OpGreater)
		if err != nil {
			t.Fatalf("compare error: %v", err)
		}
		if !result {
			t.Error("'banana' > 'apple' should be true")
		}
	})

	t.Run("string less than", func(t *testing.T) {
		vm.push(StringValue("apple"))
		vm.push(StringValue("banana"))
		result, err := vm.compare(OpLess)
		if err != nil {
			t.Fatalf("compare error: %v", err)
		}
		if !result {
			t.Error("'apple' < 'banana' should be true")
		}
	})

	t.Run("string greater or equal", func(t *testing.T) {
		vm.push(StringValue("apple"))
		vm.push(StringValue("apple"))
		result, err := vm.compare(OpGreaterEqual)
		if err != nil {
			t.Fatalf("compare error: %v", err)
		}
		if !result {
			t.Error("'apple' >= 'apple' should be true")
		}
	})

	t.Run("string less or equal", func(t *testing.T) {
		vm.push(StringValue("apple"))
		vm.push(StringValue("banana"))
		result, err := vm.compare(OpLessEqual)
		if err != nil {
			t.Fatalf("compare error: %v", err)
		}
		if !result {
			t.Error("'apple' <= 'banana' should be true")
		}
	})

	t.Run("incompatible types", func(t *testing.T) {
		vm.push(IntValue(5))
		vm.push(StringValue("hello"))
		_, err := vm.compare(OpGreater)
		if err == nil {
			t.Error("compare should return error for incompatible types")
		}
	})

	t.Run("unsupported opcode", func(t *testing.T) {
		vm.push(IntValue(5))
		vm.push(IntValue(10))
		_, err := vm.compare(OpAddInt) // OpAddInt is not a comparison opcode
		if err == nil {
			t.Error("compare should return error for unsupported opcode")
		}
	})
}

// TestConstantAsString tests the constantAsString helper
func TestConstantAsString(t *testing.T) {
	vm := NewVM()
	chunk := NewChunk("test")

	t.Run("valid string constant", func(t *testing.T) {
		idx := chunk.AddConstant(StringValue("hello"))
		result, err := vm.constantAsString(chunk, idx, "test")
		if err != nil {
			t.Fatalf("constantAsString error: %v", err)
		}
		if result != "hello" {
			t.Errorf("constantAsString = %q, want 'hello'", result)
		}
	})

	t.Run("out of range index", func(t *testing.T) {
		_, err := vm.constantAsString(chunk, 999, "test")
		if err == nil {
			t.Error("constantAsString should error on out of range index")
		}
	})

	t.Run("negative index", func(t *testing.T) {
		_, err := vm.constantAsString(chunk, -1, "test")
		if err == nil {
			t.Error("constantAsString should error on negative index")
		}
	})

	t.Run("non-string constant", func(t *testing.T) {
		idx := chunk.AddConstant(IntValue(42))
		_, err := vm.constantAsString(chunk, idx, "test")
		if err == nil {
			t.Error("constantAsString should error when constant is not a string")
		}
	})

	t.Run("nil chunk", func(t *testing.T) {
		_, err := vm.constantAsString(nil, 0, "test")
		if err == nil {
			t.Error("constantAsString should error with nil chunk")
		}
	})
}

// TestBinaryIntOp tests the binaryIntOp function
func TestBinaryIntOp(t *testing.T) {
	t.Run("add operation", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(5))
		vm.push(IntValue(3))
		err := vm.binaryIntOp(func(a, b int64) int64 { return a + b })
		if err != nil {
			t.Fatalf("binaryIntOp error: %v", err)
		}
		result, _ := vm.pop()
		if result.AsInt() != 8 {
			t.Errorf("5 + 3 = %v, want 8", result.AsInt())
		}
	})

	t.Run("subtract operation", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(10))
		vm.push(IntValue(3))
		err := vm.binaryIntOp(func(a, b int64) int64 { return a - b })
		if err != nil {
			t.Fatalf("binaryIntOp error: %v", err)
		}
		result, _ := vm.pop()
		if result.AsInt() != 7 {
			t.Errorf("10 - 3 = %v, want 7", result.AsInt())
		}
	})

	t.Run("multiply operation", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(5))
		vm.push(IntValue(3))
		err := vm.binaryIntOp(func(a, b int64) int64 { return a * b })
		if err != nil {
			t.Fatalf("binaryIntOp error: %v", err)
		}
		result, _ := vm.pop()
		if result.AsInt() != 15 {
			t.Errorf("5 * 3 = %v, want 15", result.AsInt())
		}
	})

	t.Run("non-integer left", func(t *testing.T) {
		vm := NewVM()
		vm.push(StringValue("hello"))
		vm.push(IntValue(3))
		err := vm.binaryIntOp(func(a, b int64) int64 { return a + b })
		if err == nil {
			t.Error("binaryIntOp should error on non-integer left operand")
		}
	})

	t.Run("non-integer right", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(5))
		vm.push(FloatValue(3.14))
		err := vm.binaryIntOp(func(a, b int64) int64 { return a + b })
		if err == nil {
			t.Error("binaryIntOp should error on non-integer right operand")
		}
	})
}

// TestBinaryFloatOp tests the binaryFloatOp function
func TestBinaryFloatOp(t *testing.T) {
	t.Run("add operation", func(t *testing.T) {
		vm := NewVM()
		vm.push(FloatValue(5.5))
		vm.push(FloatValue(3.2))
		err := vm.binaryFloatOp(func(a, b float64) float64 { return a + b })
		if err != nil {
			t.Fatalf("binaryFloatOp error: %v", err)
		}
		result, _ := vm.pop()
		if result.AsFloat() != 8.7 {
			t.Errorf("5.5 + 3.2 = %v, want 8.7", result.AsFloat())
		}
	})

	t.Run("int to float conversion", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(5))
		vm.push(FloatValue(3.5))
		err := vm.binaryFloatOp(func(a, b float64) float64 { return a + b })
		if err != nil {
			t.Fatalf("binaryFloatOp error: %v", err)
		}
		result, _ := vm.pop()
		if result.AsFloat() != 8.5 {
			t.Errorf("5 + 3.5 = %v, want 8.5", result.AsFloat())
		}
	})

	t.Run("non-numeric left", func(t *testing.T) {
		vm := NewVM()
		vm.push(StringValue("hello"))
		vm.push(FloatValue(3.5))
		err := vm.binaryFloatOp(func(a, b float64) float64 { return a + b })
		if err == nil {
			t.Error("binaryFloatOp should error on non-numeric left operand")
		}
	})

	t.Run("non-numeric right", func(t *testing.T) {
		vm := NewVM()
		vm.push(FloatValue(5.5))
		vm.push(StringValue("world"))
		err := vm.binaryFloatOp(func(a, b float64) float64 { return a + b })
		if err == nil {
			t.Error("binaryFloatOp should error on non-numeric right operand")
		}
	})
}
