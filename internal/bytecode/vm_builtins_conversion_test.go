package bytecode

import (
	"testing"
)

// TestBuiltinConversionFunctionsComprehensive tests type conversion built-in functions comprehensively
func TestBuiltinConversionFunctionsComprehensive(t *testing.T) {
	vm := NewVM()

	// IntToStr tests
	t.Run("IntToStr positive", func(t *testing.T) {
		result, err := builtinIntToStr(vm, []Value{IntValue(42)})
		if err != nil {
			t.Fatalf("builtinIntToStr() error = %v", err)
		}
		if result.AsString() != "42" {
			t.Errorf("builtinIntToStr(42) = %v, want '42'", result.AsString())
		}
	})

	t.Run("IntToStr negative", func(t *testing.T) {
		result, err := builtinIntToStr(vm, []Value{IntValue(-123)})
		if err != nil {
			t.Fatalf("builtinIntToStr() error = %v", err)
		}
		if result.AsString() != "-123" {
			t.Errorf("builtinIntToStr(-123) = %v, want '-123'", result.AsString())
		}
	})

	t.Run("IntToStr zero", func(t *testing.T) {
		result, err := builtinIntToStr(vm, []Value{IntValue(0)})
		if err != nil {
			t.Fatalf("builtinIntToStr() error = %v", err)
		}
		if result.AsString() != "0" {
			t.Errorf("builtinIntToStr(0) = %v, want '0'", result.AsString())
		}
	})

	t.Run("IntToStr wrong type", func(t *testing.T) {
		_, err := builtinIntToStr(vm, []Value{FloatValue(3.14)})
		if err == nil {
			t.Error("builtinIntToStr() with float should return error")
		}
	})

	// FloatToStr tests
	t.Run("FloatToStr positive", func(t *testing.T) {
		result, err := builtinFloatToStr(vm, []Value{FloatValue(3.14)})
		if err != nil {
			t.Fatalf("builtinFloatToStr() error = %v", err)
		}
		if result.AsString() != "3.14" {
			t.Errorf("builtinFloatToStr(3.14) = %v, want '3.14'", result.AsString())
		}
	})

	t.Run("FloatToStr negative", func(t *testing.T) {
		result, err := builtinFloatToStr(vm, []Value{FloatValue(-2.5)})
		if err != nil {
			t.Fatalf("builtinFloatToStr() error = %v", err)
		}
		if result.AsString() != "-2.5" {
			t.Errorf("builtinFloatToStr(-2.5) = %v, want '-2.5'", result.AsString())
		}
	})

	t.Run("FloatToStr integer", func(t *testing.T) {
		result, err := builtinFloatToStr(vm, []Value{FloatValue(42.0)})
		if err != nil {
			t.Fatalf("builtinFloatToStr() error = %v", err)
		}
		if result.AsString() != "42" {
			t.Errorf("builtinFloatToStr(42.0) = %v, want '42'", result.AsString())
		}
	})

	t.Run("FloatToStr wrong type", func(t *testing.T) {
		_, err := builtinFloatToStr(vm, []Value{IntValue(42)})
		if err == nil {
			t.Error("builtinFloatToStr() with int should return error")
		}
	})

	// StrToInt tests
	t.Run("StrToInt positive", func(t *testing.T) {
		result, err := builtinStrToInt(vm, []Value{StringValue("42")})
		if err != nil {
			t.Fatalf("builtinStrToInt() error = %v", err)
		}
		if result.AsInt() != 42 {
			t.Errorf("builtinStrToInt('42') = %v, want 42", result.AsInt())
		}
	})

	t.Run("StrToInt negative", func(t *testing.T) {
		result, err := builtinStrToInt(vm, []Value{StringValue("-123")})
		if err != nil {
			t.Fatalf("builtinStrToInt() error = %v", err)
		}
		if result.AsInt() != -123 {
			t.Errorf("builtinStrToInt('-123') = %v, want -123", result.AsInt())
		}
	})

	t.Run("StrToInt invalid", func(t *testing.T) {
		_, err := builtinStrToInt(vm, []Value{StringValue("abc")})
		if err == nil {
			t.Error("builtinStrToInt('abc') should return error")
		}
	})

	t.Run("StrToInt wrong type", func(t *testing.T) {
		_, err := builtinStrToInt(vm, []Value{IntValue(42)})
		if err == nil {
			t.Error("builtinStrToInt() with int should return error")
		}
	})

	// StrToFloat tests
	t.Run("StrToFloat positive", func(t *testing.T) {
		result, err := builtinStrToFloat(vm, []Value{StringValue("3.14")})
		if err != nil {
			t.Fatalf("builtinStrToFloat() error = %v", err)
		}
		if result.AsFloat() != 3.14 {
			t.Errorf("builtinStrToFloat('3.14') = %v, want 3.14", result.AsFloat())
		}
	})

	t.Run("StrToFloat negative", func(t *testing.T) {
		result, err := builtinStrToFloat(vm, []Value{StringValue("-2.5")})
		if err != nil {
			t.Fatalf("builtinStrToFloat() error = %v", err)
		}
		if result.AsFloat() != -2.5 {
			t.Errorf("builtinStrToFloat('-2.5') = %v, want -2.5", result.AsFloat())
		}
	})

	t.Run("StrToFloat integer string", func(t *testing.T) {
		result, err := builtinStrToFloat(vm, []Value{StringValue("42")})
		if err != nil {
			t.Fatalf("builtinStrToFloat() error = %v", err)
		}
		if result.AsFloat() != 42.0 {
			t.Errorf("builtinStrToFloat('42') = %v, want 42.0", result.AsFloat())
		}
	})

	t.Run("StrToFloat invalid", func(t *testing.T) {
		_, err := builtinStrToFloat(vm, []Value{StringValue("abc")})
		if err == nil {
			t.Error("builtinStrToFloat('abc') should return error")
		}
	})

	// StrToIntDef tests
	t.Run("StrToIntDef valid", func(t *testing.T) {
		result, err := builtinStrToIntDef(vm, []Value{StringValue("42"), IntValue(999)})
		if err != nil {
			t.Fatalf("builtinStrToIntDef() error = %v", err)
		}
		if result.AsInt() != 42 {
			t.Errorf("builtinStrToIntDef('42', 999) = %v, want 42", result.AsInt())
		}
	})

	t.Run("StrToIntDef invalid uses default", func(t *testing.T) {
		result, err := builtinStrToIntDef(vm, []Value{StringValue("abc"), IntValue(999)})
		if err != nil {
			t.Fatalf("builtinStrToIntDef() error = %v", err)
		}
		if result.AsInt() != 999 {
			t.Errorf("builtinStrToIntDef('abc', 999) = %v, want 999", result.AsInt())
		}
	})

	t.Run("StrToIntDef with whitespace", func(t *testing.T) {
		result, err := builtinStrToIntDef(vm, []Value{StringValue("  123  "), IntValue(0)})
		if err != nil {
			t.Fatalf("builtinStrToIntDef() error = %v", err)
		}
		if result.AsInt() != 123 {
			t.Errorf("builtinStrToIntDef('  123  ', 0) = %v, want 123", result.AsInt())
		}
	})

	t.Run("StrToIntDef wrong arg count", func(t *testing.T) {
		_, err := builtinStrToIntDef(vm, []Value{StringValue("42")})
		if err == nil {
			t.Error("builtinStrToIntDef() with 1 arg should return error")
		}
	})

	// StrToFloatDef tests
	t.Run("StrToFloatDef valid", func(t *testing.T) {
		result, err := builtinStrToFloatDef(vm, []Value{StringValue("3.14"), FloatValue(0.0)})
		if err != nil {
			t.Fatalf("builtinStrToFloatDef() error = %v", err)
		}
		if result.AsFloat() != 3.14 {
			t.Errorf("builtinStrToFloatDef('3.14', 0.0) = %v, want 3.14", result.AsFloat())
		}
	})

	t.Run("StrToFloatDef invalid uses default", func(t *testing.T) {
		result, err := builtinStrToFloatDef(vm, []Value{StringValue("abc"), FloatValue(99.9)})
		if err != nil {
			t.Fatalf("builtinStrToFloatDef() error = %v", err)
		}
		if result.AsFloat() != 99.9 {
			t.Errorf("builtinStrToFloatDef('abc', 99.9) = %v, want 99.9", result.AsFloat())
		}
	})

	t.Run("StrToFloatDef with int default", func(t *testing.T) {
		result, err := builtinStrToFloatDef(vm, []Value{StringValue("abc"), IntValue(99)})
		if err != nil {
			t.Fatalf("builtinStrToFloatDef() error = %v", err)
		}
		if result.AsFloat() != 99.0 {
			t.Errorf("builtinStrToFloatDef('abc', 99) = %v, want 99.0", result.AsFloat())
		}
	})

	t.Run("StrToFloatDef with whitespace", func(t *testing.T) {
		result, err := builtinStrToFloatDef(vm, []Value{StringValue("  2.5  "), FloatValue(0.0)})
		if err != nil {
			t.Fatalf("builtinStrToFloatDef() error = %v", err)
		}
		if result.AsFloat() != 2.5 {
			t.Errorf("builtinStrToFloatDef('  2.5  ', 0.0) = %v, want 2.5", result.AsFloat())
		}
	})

	// Integer cast tests
	t.Run("Integer from int", func(t *testing.T) {
		result, err := builtinInteger(vm, []Value{IntValue(42)})
		if err != nil {
			t.Fatalf("builtinInteger() error = %v", err)
		}
		if result.AsInt() != 42 {
			t.Errorf("builtinInteger(42) = %v, want 42", result.AsInt())
		}
	})

	t.Run("Integer from float", func(t *testing.T) {
		result, err := builtinInteger(vm, []Value{FloatValue(3.7)})
		if err != nil {
			t.Fatalf("builtinInteger() error = %v", err)
		}
		if result.AsInt() != 4 {
			t.Errorf("builtinInteger(3.7) = %v, want 4 (rounded)", result.AsInt())
		}
	})

	t.Run("Integer from bool true", func(t *testing.T) {
		result, err := builtinInteger(vm, []Value{BoolValue(true)})
		if err != nil {
			t.Fatalf("builtinInteger() error = %v", err)
		}
		if result.AsInt() != 1 {
			t.Errorf("builtinInteger(true) = %v, want 1", result.AsInt())
		}
	})

	t.Run("Integer from bool false", func(t *testing.T) {
		result, err := builtinInteger(vm, []Value{BoolValue(false)})
		if err != nil {
			t.Fatalf("builtinInteger() error = %v", err)
		}
		if result.AsInt() != 0 {
			t.Errorf("builtinInteger(false) = %v, want 0", result.AsInt())
		}
	})

	t.Run("Integer from string", func(t *testing.T) {
		result, err := builtinInteger(vm, []Value{StringValue("123")})
		if err != nil {
			t.Fatalf("builtinInteger() error = %v", err)
		}
		if result.AsInt() != 123 {
			t.Errorf("builtinInteger('123') = %v, want 123", result.AsInt())
		}
	})

	t.Run("Integer from invalid string", func(t *testing.T) {
		_, err := builtinInteger(vm, []Value{StringValue("abc")})
		if err == nil {
			t.Error("builtinInteger('abc') should return error")
		}
	})

	// Float cast tests
	t.Run("Float from float", func(t *testing.T) {
		result, err := builtinFloat(vm, []Value{FloatValue(3.14)})
		if err != nil {
			t.Fatalf("builtinFloat() error = %v", err)
		}
		if result.AsFloat() != 3.14 {
			t.Errorf("builtinFloat(3.14) = %v, want 3.14", result.AsFloat())
		}
	})

	t.Run("Float from int", func(t *testing.T) {
		result, err := builtinFloat(vm, []Value{IntValue(42)})
		if err != nil {
			t.Fatalf("builtinFloat() error = %v", err)
		}
		if result.AsFloat() != 42.0 {
			t.Errorf("builtinFloat(42) = %v, want 42.0", result.AsFloat())
		}
	})

	t.Run("Float from bool true", func(t *testing.T) {
		result, err := builtinFloat(vm, []Value{BoolValue(true)})
		if err != nil {
			t.Fatalf("builtinFloat() error = %v", err)
		}
		if result.AsFloat() != 1.0 {
			t.Errorf("builtinFloat(true) = %v, want 1.0", result.AsFloat())
		}
	})

	t.Run("Float from bool false", func(t *testing.T) {
		result, err := builtinFloat(vm, []Value{BoolValue(false)})
		if err != nil {
			t.Fatalf("builtinFloat() error = %v", err)
		}
		if result.AsFloat() != 0.0 {
			t.Errorf("builtinFloat(false) = %v, want 0.0", result.AsFloat())
		}
	})

	t.Run("Float from string", func(t *testing.T) {
		result, err := builtinFloat(vm, []Value{StringValue("2.5")})
		if err != nil {
			t.Fatalf("builtinFloat() error = %v", err)
		}
		if result.AsFloat() != 2.5 {
			t.Errorf("builtinFloat('2.5') = %v, want 2.5", result.AsFloat())
		}
	})

	t.Run("Float from invalid string", func(t *testing.T) {
		_, err := builtinFloat(vm, []Value{StringValue("abc")})
		if err == nil {
			t.Error("builtinFloat('abc') should return error")
		}
	})

	// String cast tests
	t.Run("String from string", func(t *testing.T) {
		result, err := builtinString(vm, []Value{StringValue("hello")})
		if err != nil {
			t.Fatalf("builtinString() error = %v", err)
		}
		if result.AsString() != "hello" {
			t.Errorf("builtinString('hello') = %v, want 'hello'", result.AsString())
		}
	})

	t.Run("String from int", func(t *testing.T) {
		result, err := builtinString(vm, []Value{IntValue(42)})
		if err != nil {
			t.Fatalf("builtinString() error = %v", err)
		}
		if result.AsString() != "42" {
			t.Errorf("builtinString(42) = %v, want '42'", result.AsString())
		}
	})

	t.Run("String from float", func(t *testing.T) {
		result, err := builtinString(vm, []Value{FloatValue(3.14)})
		if err != nil {
			t.Fatalf("builtinString() error = %v", err)
		}
		if result.AsString() != "3.14" {
			t.Errorf("builtinString(3.14) = %v, want '3.14'", result.AsString())
		}
	})

	t.Run("String from bool true", func(t *testing.T) {
		result, err := builtinString(vm, []Value{BoolValue(true)})
		if err != nil {
			t.Fatalf("builtinString() error = %v", err)
		}
		if result.AsString() != "True" {
			t.Errorf("builtinString(true) = %v, want 'True'", result.AsString())
		}
	})

	t.Run("String from bool false", func(t *testing.T) {
		result, err := builtinString(vm, []Value{BoolValue(false)})
		if err != nil {
			t.Fatalf("builtinString() error = %v", err)
		}
		if result.AsString() != "False" {
			t.Errorf("builtinString(false) = %v, want 'False'", result.AsString())
		}
	})

	// Boolean cast tests
	t.Run("Boolean from bool", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{BoolValue(true)})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinBoolean(true) = false, want true")
		}
	})

	t.Run("Boolean from int non-zero", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{IntValue(42)})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinBoolean(42) = false, want true")
		}
	})

	t.Run("Boolean from int zero", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{IntValue(0)})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinBoolean(0) = true, want false")
		}
	})

	t.Run("Boolean from float non-zero", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{FloatValue(3.14)})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinBoolean(3.14) = false, want true")
		}
	})

	t.Run("Boolean from float zero", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{FloatValue(0.0)})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinBoolean(0.0) = true, want false")
		}
	})

	t.Run("Boolean from string 'true'", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{StringValue("true")})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinBoolean('true') = false, want true")
		}
	})

	t.Run("Boolean from string 'TRUE'", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{StringValue("TRUE")})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinBoolean('TRUE') = false, want true")
		}
	})

	t.Run("Boolean from string '1'", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{StringValue("1")})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinBoolean('1') = false, want true")
		}
	})

	t.Run("Boolean from string 'false'", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{StringValue("false")})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinBoolean('false') = true, want false")
		}
	})

	t.Run("Boolean from string 'FALSE'", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{StringValue("FALSE")})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinBoolean('FALSE') = true, want false")
		}
	})

	t.Run("Boolean from string '0'", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{StringValue("0")})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinBoolean('0') = true, want false")
		}
	})

	t.Run("Boolean from empty string", func(t *testing.T) {
		result, err := builtinBoolean(vm, []Value{StringValue("")})
		if err != nil {
			t.Fatalf("builtinBoolean() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinBoolean('') = true, want false")
		}
	})

	t.Run("Boolean from invalid string", func(t *testing.T) {
		_, err := builtinBoolean(vm, []Value{StringValue("abc")})
		if err == nil {
			t.Error("builtinBoolean('abc') should return error")
		}
	})
}
