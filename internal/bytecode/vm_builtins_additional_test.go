package bytecode

import (
	"bytes"
	"math"
	"testing"
)

// TestBuiltinMiscFunctionsAdditional tests miscellaneous built-in functions
func TestBuiltinMiscFunctionsAdditional(t *testing.T) {
	vm := NewVM()

	t.Run("Print single string", func(t *testing.T) {
		var buf bytes.Buffer
		vm.output = &buf
		_, err := builtinPrint(vm, []Value{StringValue("Hello")})
		if err != nil {
			t.Fatalf("builtinPrint() error = %v", err)
		}
		if buf.String() != "Hello" {
			t.Errorf("builtinPrint() = %q, want %q", buf.String(), "Hello")
		}
	})

	t.Run("Print multiple values", func(t *testing.T) {
		var buf bytes.Buffer
		vm.output = &buf
		_, err := builtinPrint(vm, []Value{
			StringValue("Hello"),
			IntValue(42),
			FloatValue(3.14),
		})
		if err != nil {
			t.Fatalf("builtinPrint() error = %v", err)
		}
		if buf.String() != "Hello 42 3.14" {
			t.Errorf("builtinPrint() = %q, want %q", buf.String(), "Hello 42 3.14")
		}
	})

	t.Run("Print no args", func(t *testing.T) {
		var buf bytes.Buffer
		vm.output = &buf
		_, err := builtinPrint(vm, []Value{})
		if err != nil {
			t.Fatalf("builtinPrint() error = %v", err)
		}
		if buf.String() != "" {
			t.Errorf("builtinPrint() = %q, want empty", buf.String())
		}
	})

	t.Run("PrintLn single value", func(t *testing.T) {
		var buf bytes.Buffer
		vm.output = &buf
		_, err := builtinPrintLn(vm, []Value{StringValue("Test")})
		if err != nil {
			t.Fatalf("builtinPrintLn() error = %v", err)
		}
		if buf.String() != "Test\n" {
			t.Errorf("builtinPrintLn() = %q, want %q", buf.String(), "Test\n")
		}
	})

	t.Run("PrintLn multiple values", func(t *testing.T) {
		var buf bytes.Buffer
		vm.output = &buf
		_, err := builtinPrintLn(vm, []Value{
			IntValue(1),
			IntValue(2),
			IntValue(3),
		})
		if err != nil {
			t.Fatalf("builtinPrintLn() error = %v", err)
		}
		if buf.String() != "1 2 3\n" {
			t.Errorf("builtinPrintLn() = %q, want %q", buf.String(), "1 2 3\n")
		}
	})

	t.Run("Length of string", func(t *testing.T) {
		result, err := builtinLength(vm, []Value{StringValue("hello")})
		if err != nil {
			t.Fatalf("builtinLength() error = %v", err)
		}
		if result.AsInt() != 5 {
			t.Errorf("builtinLength('hello') = %v, want 5", result.AsInt())
		}
	})

	t.Run("Length of empty string", func(t *testing.T) {
		result, err := builtinLength(vm, []Value{StringValue("")})
		if err != nil {
			t.Fatalf("builtinLength() error = %v", err)
		}
		if result.AsInt() != 0 {
			t.Errorf("builtinLength('') = %v, want 0", result.AsInt())
		}
	})

	t.Run("Length of array", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1), IntValue(2), IntValue(3)})
		result, err := builtinLength(vm, []Value{ArrayValue(arr)})
		if err != nil {
			t.Fatalf("builtinLength() error = %v", err)
		}
		if result.AsInt() != 3 {
			t.Errorf("builtinLength(array) = %v, want 3", result.AsInt())
		}
	})

	t.Run("Length wrong type", func(t *testing.T) {
		_, err := builtinLength(vm, []Value{IntValue(42)})
		if err == nil {
			t.Error("builtinLength() with int should return error")
		}
	})

	t.Run("Length wrong arg count", func(t *testing.T) {
		_, err := builtinLength(vm, []Value{})
		if err == nil {
			t.Error("builtinLength() with no args should return error")
		}
	})
}

// TestBuiltinMathFunctionsAdditional tests additional math built-in functions
func TestBuiltinMathFunctionsAdditional(t *testing.T) {
	vm := NewVM()

	t.Run("Log10 of 100", func(t *testing.T) {
		result, err := builtinLog10(vm, []Value{FloatValue(100.0)})
		if err != nil {
			t.Fatalf("builtinLog10() error = %v", err)
		}
		if result.AsFloat() != 2.0 {
			t.Errorf("builtinLog10(100) = %v, want 2.0", result.AsFloat())
		}
	})

	t.Run("Log10 of 1000", func(t *testing.T) {
		result, err := builtinLog10(vm, []Value{FloatValue(1000.0)})
		if err != nil {
			t.Fatalf("builtinLog10() error = %v", err)
		}
		if result.AsFloat() != 3.0 {
			t.Errorf("builtinLog10(1000) = %v, want 3.0", result.AsFloat())
		}
	})

	t.Run("Log10 of 1", func(t *testing.T) {
		result, err := builtinLog10(vm, []Value{FloatValue(1.0)})
		if err != nil {
			t.Fatalf("builtinLog10() error = %v", err)
		}
		if result.AsFloat() != 0.0 {
			t.Errorf("builtinLog10(1) = %v, want 0.0", result.AsFloat())
		}
	})

	t.Run("Log10 with int", func(t *testing.T) {
		result, err := builtinLog10(vm, []Value{IntValue(100)})
		if err != nil {
			t.Fatalf("builtinLog10() error = %v", err)
		}
		if result.AsFloat() != 2.0 {
			t.Errorf("builtinLog10(100) = %v, want 2.0", result.AsFloat())
		}
	})

	t.Run("Log10 wrong type", func(t *testing.T) {
		_, err := builtinLog10(vm, []Value{StringValue("test")})
		if err == nil {
			t.Error("builtinLog10() with string should return error")
		}
	})

	t.Run("LogN base 2 of 8", func(t *testing.T) {
		result, err := builtinLogN(vm, []Value{FloatValue(8.0), FloatValue(2.0)})
		if err != nil {
			t.Fatalf("builtinLogN() error = %v", err)
		}
		if result.AsFloat() != 3.0 {
			t.Errorf("builtinLogN(8, 2) = %v, want 3.0", result.AsFloat())
		}
	})

	t.Run("LogN base 10 of 100", func(t *testing.T) {
		result, err := builtinLogN(vm, []Value{FloatValue(100.0), FloatValue(10.0)})
		if err != nil {
			t.Fatalf("builtinLogN() error = %v", err)
		}
		if result.AsFloat() != 2.0 {
			t.Errorf("builtinLogN(100, 10) = %v, want 2.0", result.AsFloat())
		}
	})

	t.Run("LogN with ints", func(t *testing.T) {
		result, err := builtinLogN(vm, []Value{IntValue(8), IntValue(2)})
		if err != nil {
			t.Fatalf("builtinLogN() error = %v", err)
		}
		if result.AsFloat() != 3.0 {
			t.Errorf("builtinLogN(8, 2) = %v, want 3.0", result.AsFloat())
		}
	})

	t.Run("LogN wrong arg count", func(t *testing.T) {
		_, err := builtinLogN(vm, []Value{FloatValue(8.0)})
		if err == nil {
			t.Error("builtinLogN() with 1 arg should return error")
		}
	})

	t.Run("IsFinite with normal number", func(t *testing.T) {
		result, err := builtinIsFinite(vm, []Value{FloatValue(42.0)})
		if err != nil {
			t.Fatalf("builtinIsFinite() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinIsFinite(42.0) = false, want true")
		}
	})

	t.Run("IsFinite with infinity", func(t *testing.T) {
		result, err := builtinIsFinite(vm, []Value{FloatValue(math.Inf(1))})
		if err != nil {
			t.Fatalf("builtinIsFinite() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinIsFinite(Inf) = true, want false")
		}
	})

	t.Run("IsFinite with NaN", func(t *testing.T) {
		result, err := builtinIsFinite(vm, []Value{FloatValue(math.NaN())})
		if err != nil {
			t.Fatalf("builtinIsFinite() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinIsFinite(NaN) = true, want false")
		}
	})

	t.Run("IsFinite with int", func(t *testing.T) {
		result, err := builtinIsFinite(vm, []Value{IntValue(42)})
		if err != nil {
			t.Fatalf("builtinIsFinite() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinIsFinite(42) = false, want true")
		}
	})

	t.Run("IsInfinite with infinity", func(t *testing.T) {
		result, err := builtinIsInfinite(vm, []Value{FloatValue(math.Inf(1))})
		if err != nil {
			t.Fatalf("builtinIsInfinite() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinIsInfinite(Inf) = false, want true")
		}
	})

	t.Run("IsInfinite with negative infinity", func(t *testing.T) {
		result, err := builtinIsInfinite(vm, []Value{FloatValue(math.Inf(-1))})
		if err != nil {
			t.Fatalf("builtinIsInfinite() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinIsInfinite(-Inf) = false, want true")
		}
	})

	t.Run("IsInfinite with normal number", func(t *testing.T) {
		result, err := builtinIsInfinite(vm, []Value{FloatValue(42.0)})
		if err != nil {
			t.Fatalf("builtinIsInfinite() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinIsInfinite(42.0) = true, want false")
		}
	})

	t.Run("IsInfinite with NaN", func(t *testing.T) {
		result, err := builtinIsInfinite(vm, []Value{FloatValue(math.NaN())})
		if err != nil {
			t.Fatalf("builtinIsInfinite() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinIsInfinite(NaN) = true, want false")
		}
	})

	t.Run("LeastFactor of 12", func(t *testing.T) {
		result, err := builtinLeastFactor(vm, []Value{IntValue(12)})
		if err != nil {
			t.Fatalf("builtinLeastFactor() error = %v", err)
		}
		if result.AsInt() != 2 {
			t.Errorf("builtinLeastFactor(12) = %v, want 2", result.AsInt())
		}
	})

	t.Run("LeastFactor of 15", func(t *testing.T) {
		result, err := builtinLeastFactor(vm, []Value{IntValue(15)})
		if err != nil {
			t.Fatalf("builtinLeastFactor() error = %v", err)
		}
		if result.AsInt() != 3 {
			t.Errorf("builtinLeastFactor(15) = %v, want 3", result.AsInt())
		}
	})

	t.Run("LeastFactor of prime 17", func(t *testing.T) {
		result, err := builtinLeastFactor(vm, []Value{IntValue(17)})
		if err != nil {
			t.Fatalf("builtinLeastFactor() error = %v", err)
		}
		if result.AsInt() != 17 {
			t.Errorf("builtinLeastFactor(17) = %v, want 17 (prime)", result.AsInt())
		}
	})

	t.Run("LeastFactor of 1", func(t *testing.T) {
		result, err := builtinLeastFactor(vm, []Value{IntValue(1)})
		if err != nil {
			t.Fatalf("builtinLeastFactor() error = %v", err)
		}
		if result.AsInt() != 1 {
			t.Errorf("builtinLeastFactor(1) = %v, want 1", result.AsInt())
		}
	})

	t.Run("LeastFactor of 35 (5*7)", func(t *testing.T) {
		result, err := builtinLeastFactor(vm, []Value{IntValue(35)})
		if err != nil {
			t.Fatalf("builtinLeastFactor() error = %v", err)
		}
		if result.AsInt() != 5 {
			t.Errorf("builtinLeastFactor(35) = %v, want 5", result.AsInt())
		}
	})

	t.Run("LeastFactor of 49 (7*7)", func(t *testing.T) {
		result, err := builtinLeastFactor(vm, []Value{IntValue(49)})
		if err != nil {
			t.Fatalf("builtinLeastFactor() error = %v", err)
		}
		if result.AsInt() != 7 {
			t.Errorf("builtinLeastFactor(49) = %v, want 7", result.AsInt())
		}
	})

	t.Run("LeastFactor wrong type", func(t *testing.T) {
		_, err := builtinLeastFactor(vm, []Value{FloatValue(12.5)})
		if err == nil {
			t.Error("builtinLeastFactor() with float should return error")
		}
	})

	t.Run("CompareNum equal", func(t *testing.T) {
		result, err := builtinCompareNum(vm, []Value{FloatValue(3.14), FloatValue(3.14)})
		if err != nil {
			t.Fatalf("builtinCompareNum() error = %v", err)
		}
		if result.AsInt() != 0 {
			t.Errorf("builtinCompareNum(3.14, 3.14) = %v, want 0", result.AsInt())
		}
	})

	t.Run("CompareNum less than", func(t *testing.T) {
		result, err := builtinCompareNum(vm, []Value{FloatValue(1.0), FloatValue(2.0)})
		if err != nil {
			t.Fatalf("builtinCompareNum() error = %v", err)
		}
		if result.AsInt() != -1 {
			t.Errorf("builtinCompareNum(1.0, 2.0) = %v, want -1", result.AsInt())
		}
	})

	t.Run("CompareNum greater than", func(t *testing.T) {
		result, err := builtinCompareNum(vm, []Value{FloatValue(5.0), FloatValue(2.0)})
		if err != nil {
			t.Fatalf("builtinCompareNum() error = %v", err)
		}
		if result.AsInt() != 1 {
			t.Errorf("builtinCompareNum(5.0, 2.0) = %v, want 1", result.AsInt())
		}
	})

	t.Run("CompareNum with ints", func(t *testing.T) {
		result, err := builtinCompareNum(vm, []Value{IntValue(10), IntValue(10)})
		if err != nil {
			t.Fatalf("builtinCompareNum() error = %v", err)
		}
		if result.AsInt() != 0 {
			t.Errorf("builtinCompareNum(10, 10) = %v, want 0", result.AsInt())
		}
	})

	t.Run("CompareNum mixed int and float", func(t *testing.T) {
		result, err := builtinCompareNum(vm, []Value{IntValue(3), FloatValue(3.5)})
		if err != nil {
			t.Fatalf("builtinCompareNum() error = %v", err)
		}
		if result.AsInt() != -1 {
			t.Errorf("builtinCompareNum(3, 3.5) = %v, want -1", result.AsInt())
		}
	})

	t.Run("CompareNum wrong type", func(t *testing.T) {
		_, err := builtinCompareNum(vm, []Value{StringValue("abc"), FloatValue(1.0)})
		if err == nil {
			t.Error("builtinCompareNum() with string should return error")
		}
	})

	t.Run("CompareNum wrong arg count", func(t *testing.T) {
		_, err := builtinCompareNum(vm, []Value{FloatValue(1.0)})
		if err == nil {
			t.Error("builtinCompareNum() with 1 arg should return error")
		}
	})
}
