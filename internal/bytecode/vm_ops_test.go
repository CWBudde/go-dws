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
