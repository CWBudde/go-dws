package bytecode

import (
	"math"
	"testing"
)

// TestBuiltinMathFunctions tests math built-in functions
func TestBuiltinMathFunctions(t *testing.T) {
	vm := NewVM()

	t.Run("Pi", func(t *testing.T) {
		result, err := builtinPi(vm, []Value{})
		if err != nil {
			t.Fatalf("builtinPi() error = %v", err)
		}
		if !result.IsFloat() {
			t.Fatalf("builtinPi() should return float, got %v", result.Type)
		}
		if result.AsFloat() != math.Pi {
			t.Errorf("builtinPi() = %v, want %v", result.AsFloat(), math.Pi)
		}
	})

	t.Run("Pi with args", func(t *testing.T) {
		_, err := builtinPi(vm, []Value{IntValue(1)})
		if err == nil {
			t.Error("builtinPi() with args should return error")
		}
	})

	t.Run("Sign positive", func(t *testing.T) {
		result, err := builtinSign(vm, []Value{IntValue(42)})
		if err != nil {
			t.Fatalf("builtinSign() error = %v", err)
		}
		if result.AsInt() != 1 {
			t.Errorf("builtinSign(42) = %v, want 1", result.AsInt())
		}
	})

	t.Run("Sign negative", func(t *testing.T) {
		result, err := builtinSign(vm, []Value{FloatValue(-3.14)})
		if err != nil {
			t.Fatalf("builtinSign() error = %v", err)
		}
		if result.AsInt() != -1 {
			t.Errorf("builtinSign(-3.14) = %v, want -1", result.AsInt())
		}
	})

	t.Run("Sign zero", func(t *testing.T) {
		result, err := builtinSign(vm, []Value{IntValue(0)})
		if err != nil {
			t.Fatalf("builtinSign() error = %v", err)
		}
		if result.AsInt() != 0 {
			t.Errorf("builtinSign(0) = %v, want 0", result.AsInt())
		}
	})

	t.Run("Odd true", func(t *testing.T) {
		result, err := builtinOdd(vm, []Value{IntValue(3)})
		if err != nil {
			t.Fatalf("builtinOdd() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinOdd(3) = false, want true")
		}
	})

	t.Run("Odd false", func(t *testing.T) {
		result, err := builtinOdd(vm, []Value{IntValue(4)})
		if err != nil {
			t.Fatalf("builtinOdd() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinOdd(4) = true, want false")
		}
	})

	t.Run("Frac", func(t *testing.T) {
		result, err := builtinFrac(vm, []Value{FloatValue(3.14)})
		if err != nil {
			t.Fatalf("builtinFrac() error = %v", err)
		}
		expected := 0.14000000000000012 // floating point precision
		if math.Abs(result.AsFloat()-expected) > 0.0001 {
			t.Errorf("builtinFrac(3.14) = %v, want ~%v", result.AsFloat(), expected)
		}
	})

	t.Run("Int", func(t *testing.T) {
		result, err := builtinInt(vm, []Value{FloatValue(3.99)})
		if err != nil {
			t.Fatalf("builtinInt() error = %v", err)
		}
		if result.AsFloat() != 3.0 {
			t.Errorf("builtinInt(3.99) = %v, want 3.0", result.AsFloat())
		}
	})

	t.Run("Log10", func(t *testing.T) {
		result, err := builtinLog10(vm, []Value{FloatValue(100.0)})
		if err != nil {
			t.Fatalf("builtinLog10() error = %v", err)
		}
		if result.AsFloat() != 2.0 {
			t.Errorf("builtinLog10(100) = %v, want 2.0", result.AsFloat())
		}
	})

	t.Run("LogN", func(t *testing.T) {
		result, err := builtinLogN(vm, []Value{FloatValue(8.0), FloatValue(2.0)})
		if err != nil {
			t.Fatalf("builtinLogN() error = %v", err)
		}
		if result.AsFloat() != 3.0 {
			t.Errorf("builtinLogN(8, 2) = %v, want 3.0", result.AsFloat())
		}
	})

	t.Run("Infinity", func(t *testing.T) {
		result, err := builtinInfinity(vm, []Value{})
		if err != nil {
			t.Fatalf("builtinInfinity() error = %v", err)
		}
		if !math.IsInf(result.AsFloat(), 1) {
			t.Errorf("builtinInfinity() = %v, want +Inf", result.AsFloat())
		}
	})

	t.Run("NaN", func(t *testing.T) {
		result, err := builtinNaN(vm, []Value{})
		if err != nil {
			t.Fatalf("builtinNaN() error = %v", err)
		}
		if !math.IsNaN(result.AsFloat()) {
			t.Errorf("builtinNaN() = %v, want NaN", result.AsFloat())
		}
	})

	t.Run("IsFinite true", func(t *testing.T) {
		result, err := builtinIsFinite(vm, []Value{FloatValue(42.0)})
		if err != nil {
			t.Fatalf("builtinIsFinite() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinIsFinite(42) = false, want true")
		}
	})

	t.Run("IsFinite false", func(t *testing.T) {
		result, err := builtinIsFinite(vm, []Value{FloatValue(math.Inf(1))})
		if err != nil {
			t.Fatalf("builtinIsFinite() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinIsFinite(Inf) = true, want false")
		}
	})

	t.Run("IsInfinite true", func(t *testing.T) {
		result, err := builtinIsInfinite(vm, []Value{FloatValue(math.Inf(1))})
		if err != nil {
			t.Fatalf("builtinIsInfinite() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinIsInfinite(Inf) = false, want true")
		}
	})

	t.Run("IsInfinite false", func(t *testing.T) {
		result, err := builtinIsInfinite(vm, []Value{FloatValue(42.0)})
		if err != nil {
			t.Fatalf("builtinIsInfinite() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinIsInfinite(42) = true, want false")
		}
	})

	t.Run("IntPower", func(t *testing.T) {
		result, err := builtinIntPower(vm, []Value{IntValue(2), IntValue(8)})
		if err != nil {
			t.Fatalf("builtinIntPower() error = %v", err)
		}
		if result.AsFloat() != 256.0 {
			t.Errorf("builtinIntPower(2, 8) = %v, want 256.0", result.AsFloat())
		}
	})

	t.Run("Factorial", func(t *testing.T) {
		result, err := builtinFactorial(vm, []Value{IntValue(5)})
		if err != nil {
			t.Fatalf("builtinFactorial() error = %v", err)
		}
		if result.AsInt() != 120 {
			t.Errorf("builtinFactorial(5) = %v, want 120", result.AsInt())
		}
	})

	t.Run("Gcd", func(t *testing.T) {
		result, err := builtinGcd(vm, []Value{IntValue(48), IntValue(18)})
		if err != nil {
			t.Fatalf("builtinGcd() error = %v", err)
		}
		if result.AsInt() != 6 {
			t.Errorf("builtinGcd(48, 18) = %v, want 6", result.AsInt())
		}
	})

	t.Run("Lcm", func(t *testing.T) {
		result, err := builtinLcm(vm, []Value{IntValue(4), IntValue(6)})
		if err != nil {
			t.Fatalf("builtinLcm() error = %v", err)
		}
		if result.AsInt() != 12 {
			t.Errorf("builtinLcm(4, 6) = %v, want 12", result.AsInt())
		}
	})

	t.Run("IsPrime true", func(t *testing.T) {
		result, err := builtinIsPrime(vm, []Value{IntValue(17)})
		if err != nil {
			t.Fatalf("builtinIsPrime() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinIsPrime(17) = false, want true")
		}
	})

	t.Run("IsPrime false", func(t *testing.T) {
		result, err := builtinIsPrime(vm, []Value{IntValue(15)})
		if err != nil {
			t.Fatalf("builtinIsPrime() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinIsPrime(15) = true, want false")
		}
	})

	t.Run("LeastFactor", func(t *testing.T) {
		result, err := builtinLeastFactor(vm, []Value{IntValue(15)})
		if err != nil {
			t.Fatalf("builtinLeastFactor() error = %v", err)
		}
		if result.AsInt() != 3 {
			t.Errorf("builtinLeastFactor(15) = %v, want 3", result.AsInt())
		}
	})

	t.Run("PopCount", func(t *testing.T) {
		result, err := builtinPopCount(vm, []Value{IntValue(7)}) // 0b111
		if err != nil {
			t.Fatalf("builtinPopCount() error = %v", err)
		}
		if result.AsInt() != 3 {
			t.Errorf("builtinPopCount(7) = %v, want 3", result.AsInt())
		}
	})

	t.Run("TestBit true", func(t *testing.T) {
		result, err := builtinTestBit(vm, []Value{IntValue(5), IntValue(0)}) // 0b101, bit 0
		if err != nil {
			t.Fatalf("builtinTestBit() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinTestBit(5, 0) = false, want true")
		}
	})

	t.Run("TestBit false", func(t *testing.T) {
		result, err := builtinTestBit(vm, []Value{IntValue(5), IntValue(1)}) // 0b101, bit 1
		if err != nil {
			t.Fatalf("builtinTestBit() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinTestBit(5, 1) = true, want false")
		}
	})

	t.Run("Haversine", func(t *testing.T) {
		// Test with known coordinates (simple case: same point)
		result, err := builtinHaversine(vm, []Value{
			FloatValue(0.0), FloatValue(0.0), // lat1, lon1
			FloatValue(0.0), FloatValue(0.0), // lat2, lon2
		})
		if err != nil {
			t.Fatalf("builtinHaversine() error = %v", err)
		}
		if result.AsFloat() != 0.0 {
			t.Errorf("builtinHaversine(same point) = %v, want 0.0", result.AsFloat())
		}
	})

	t.Run("CompareNum", func(t *testing.T) {
		result, err := builtinCompareNum(vm, []Value{FloatValue(3.0), FloatValue(5.0)})
		if err != nil {
			t.Fatalf("builtinCompareNum() error = %v", err)
		}
		if result.AsInt() != -1 {
			t.Errorf("builtinCompareNum(3, 5) = %v, want -1", result.AsInt())
		}
	})

	t.Run("RandSeed", func(t *testing.T) {
		result, err := builtinRandSeed(vm, []Value{})
		if err != nil {
			t.Fatalf("builtinRandSeed() error = %v", err)
		}
		if !result.IsInt() {
			t.Errorf("builtinRandSeed() should return int")
		}
	})

	t.Run("SetRandSeed", func(t *testing.T) {
		_, err := builtinSetRandSeed(vm, []Value{IntValue(12345)})
		if err != nil {
			t.Fatalf("builtinSetRandSeed() error = %v", err)
		}
	})

	t.Run("Randomize", func(t *testing.T) {
		_, err := builtinRandomize(vm, []Value{})
		if err != nil {
			t.Fatalf("builtinRandomize() error = %v", err)
		}
	})

	t.Run("RandG", func(t *testing.T) {
		result, err := builtinRandG(vm, []Value{})
		if err != nil {
			t.Fatalf("builtinRandG() error = %v", err)
		}
		if !result.IsFloat() {
			t.Errorf("builtinRandG() should return float")
		}
	})
}

// TestBuiltinConversionFunctions tests conversion built-in functions
func TestBuiltinConversionFunctions(t *testing.T) {
	vm := NewVM()

	t.Run("StrToIntDef valid", func(t *testing.T) {
		result, err := builtinStrToIntDef(vm, []Value{StringValue("42"), IntValue(0)})
		if err != nil {
			t.Fatalf("builtinStrToIntDef() error = %v", err)
		}
		if result.AsInt() != 42 {
			t.Errorf("builtinStrToIntDef('42', 0) = %v, want 42", result.AsInt())
		}
	})

	t.Run("StrToIntDef invalid", func(t *testing.T) {
		result, err := builtinStrToIntDef(vm, []Value{StringValue("abc"), IntValue(99)})
		if err != nil {
			t.Fatalf("builtinStrToIntDef() error = %v", err)
		}
		if result.AsInt() != 99 {
			t.Errorf("builtinStrToIntDef('abc', 99) = %v, want 99", result.AsInt())
		}
	})

	t.Run("StrToFloatDef valid", func(t *testing.T) {
		result, err := builtinStrToFloatDef(vm, []Value{StringValue("3.14"), FloatValue(0.0)})
		if err != nil {
			t.Fatalf("builtinStrToFloatDef() error = %v", err)
		}
		if result.AsFloat() != 3.14 {
			t.Errorf("builtinStrToFloatDef('3.14', 0) = %v, want 3.14", result.AsFloat())
		}
	})

	t.Run("StrToFloatDef invalid", func(t *testing.T) {
		result, err := builtinStrToFloatDef(vm, []Value{StringValue("abc"), FloatValue(9.9)})
		if err != nil {
			t.Fatalf("builtinStrToFloatDef() error = %v", err)
		}
		if result.AsFloat() != 9.9 {
			t.Errorf("builtinStrToFloatDef('abc', 9.9) = %v, want 9.9", result.AsFloat())
		}
	})

	t.Run("Float", func(t *testing.T) {
		result, err := builtinFloat(vm, []Value{IntValue(42)})
		if err != nil {
			t.Fatalf("builtinFloat() error = %v", err)
		}
		if result.AsFloat() != 42.0 {
			t.Errorf("builtinFloat(42) = %v, want 42.0", result.AsFloat())
		}
	})

	t.Run("String", func(t *testing.T) {
		result, err := builtinString(vm, []Value{IntValue(42)})
		if err != nil {
			t.Fatalf("builtinString() error = %v", err)
		}
		if result.AsString() != "42" {
			t.Errorf("builtinString(42) = %v, want '42'", result.AsString())
		}
	})

	t.Run("Boolean from int", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{IntValue(1)})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinBoolean(1) = false, want true")
		}
	})
}

// TestBuiltinStringFunctions tests string built-in functions
func TestBuiltinStringFunctions(t *testing.T) {
	vm := NewVM()

	t.Run("SubStr", func(t *testing.T) {
		result, err := builtinSubStr(vm, []Value{StringValue("hello"), IntValue(2), IntValue(3)})
		if err != nil {
			t.Fatalf("builtinSubStr() error = %v", err)
		}
		if result.AsString() != "ell" {
			t.Errorf("builtinSubStr('hello', 2, 3) = %v, want 'ell'", result.AsString())
		}
	})

	t.Run("DupeString", func(t *testing.T) {
		result, err := builtinDupeString(vm, []Value{StringValue("ab"), IntValue(3)})
		if err != nil {
			t.Fatalf("builtinDupeString() error = %v", err)
		}
		if result.AsString() != "ababab" {
			t.Errorf("builtinDupeString('ab', 3) = %v, want 'ababab'", result.AsString())
		}
	})

	t.Run("NormalizeString", func(t *testing.T) {
		result, err := builtinNormalizeString(vm, []Value{StringValue("hello"), StringValue("nfc")})
		if err != nil {
			t.Fatalf("builtinNormalizeString() error = %v", err)
		}
		if !result.IsString() {
			t.Errorf("builtinNormalizeString() should return string")
		}
	})

	t.Run("StripAccents", func(t *testing.T) {
		result, err := builtinStripAccents(vm, []Value{StringValue("café")})
		if err != nil {
			t.Fatalf("builtinStripAccents() error = %v", err)
		}
		if result.AsString() != "cafe" {
			t.Errorf("builtinStripAccents('café') = %v, want 'cafe'", result.AsString())
		}
	})

	t.Run("SameText true", func(t *testing.T) {
		result, err := builtinSameText(vm, []Value{StringValue("Hello"), StringValue("hello")})
		if err != nil {
			t.Fatalf("builtinSameText() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinSameText('Hello', 'hello') = false, want true")
		}
	})

	t.Run("SameText false", func(t *testing.T) {
		result, err := builtinSameText(vm, []Value{StringValue("Hello"), StringValue("World")})
		if err != nil {
			t.Fatalf("builtinSameText() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinSameText('Hello', 'World') = true, want false")
		}
	})

	t.Run("CompareText", func(t *testing.T) {
		result, err := builtinCompareText(vm, []Value{StringValue("abc"), StringValue("ABC")})
		if err != nil {
			t.Fatalf("builtinCompareText() error = %v", err)
		}
		if result.AsInt() != 0 {
			t.Errorf("builtinCompareText('abc', 'ABC') = %v, want 0", result.AsInt())
		}
	})

	t.Run("CompareStr", func(t *testing.T) {
		result, err := builtinCompareStr(vm, []Value{StringValue("abc"), StringValue("ABC")})
		if err != nil {
			t.Fatalf("builtinCompareStr() error = %v", err)
		}
		// Case-sensitive comparison, 'abc' > 'ABC'
		if result.AsInt() <= 0 {
			t.Errorf("builtinCompareStr('abc', 'ABC') = %v, want > 0", result.AsInt())
		}
	})

	t.Run("AnsiCompareText", func(t *testing.T) {
		result, err := builtinAnsiCompareText(vm, []Value{StringValue("abc"), StringValue("def")})
		if err != nil {
			t.Fatalf("builtinAnsiCompareText() error = %v", err)
		}
		if result.AsInt() >= 0 {
			t.Errorf("builtinAnsiCompareText('abc', 'def') = %v, want < 0", result.AsInt())
		}
	})

	t.Run("AnsiCompareStr", func(t *testing.T) {
		result, err := builtinAnsiCompareStr(vm, []Value{StringValue("abc"), StringValue("def")})
		if err != nil {
			t.Fatalf("builtinAnsiCompareStr() error = %v", err)
		}
		if result.AsInt() >= 0 {
			t.Errorf("builtinAnsiCompareStr('abc', 'def') = %v, want < 0", result.AsInt())
		}
	})
}

// TestBuiltinMiscFunctions tests miscellaneous built-in functions
func TestBuiltinMiscFunctions(t *testing.T) {
	vm := NewVM()

	t.Run("Print", func(t *testing.T) {
		// builtinPrint outputs to vm.output, just verify it doesn't error
		_, err := builtinPrint(vm, []Value{StringValue("test")})
		if err != nil {
			t.Fatalf("builtinPrint() error = %v", err)
		}
	})
}

// TestBuiltinErrorCases tests error handling in built-in functions
func TestBuiltinErrorCases(t *testing.T) {
	vm := NewVM()

	t.Run("Sign wrong arg count", func(t *testing.T) {
		_, err := builtinSign(vm, []Value{})
		if err == nil {
			t.Error("builtinSign() with no args should return error")
		}
	})

	t.Run("Sign wrong type", func(t *testing.T) {
		_, err := builtinSign(vm, []Value{StringValue("hello")})
		if err == nil {
			t.Error("builtinSign() with string arg should return error")
		}
	})

	t.Run("Odd wrong type", func(t *testing.T) {
		_, err := builtinOdd(vm, []Value{FloatValue(3.14)})
		if err == nil {
			t.Error("builtinOdd() with float arg should return error")
		}
	})

	t.Run("IntPower wrong arg count", func(t *testing.T) {
		_, err := builtinIntPower(vm, []Value{IntValue(2)})
		if err == nil {
			t.Error("builtinIntPower() with one arg should return error")
		}
	})

	t.Run("Factorial negative", func(t *testing.T) {
		_, err := builtinFactorial(vm, []Value{IntValue(-1)})
		if err == nil {
			t.Error("builtinFactorial() with negative arg should return error")
		}
	})

	t.Run("TestBit negative bit", func(t *testing.T) {
		_, err := builtinTestBit(vm, []Value{IntValue(5), IntValue(-1)})
		if err == nil {
			t.Error("builtinTestBit() with negative bit should return error")
		}
	})

	t.Run("Haversine wrong arg count", func(t *testing.T) {
		_, err := builtinHaversine(vm, []Value{FloatValue(0.0)})
		if err == nil {
			t.Error("builtinHaversine() with wrong arg count should return error")
		}
	})
}
