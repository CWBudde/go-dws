package bytecode

import (
	"testing"
)

// TestBuiltinStringFunctionsAdditional tests uncovered string built-in functions
func TestBuiltinStringFunctionsAdditional(t *testing.T) {
	vm := NewVM()

	t.Run("SubString basic", func(t *testing.T) {
		result, err := builtinSubString(vm, []Value{
			StringValue("Hello, World!"),
			IntValue(1),
			IntValue(5),
		})
		if err != nil {
			t.Fatalf("builtinSubString() error = %v", err)
		}
		if result.AsString() != "Hello" {
			t.Errorf("builtinSubString('Hello, World!', 1, 5) = %v, want 'Hello'", result.AsString())
		}
	})

	t.Run("SubString middle", func(t *testing.T) {
		result, err := builtinSubString(vm, []Value{
			StringValue("Hello, World!"),
			IntValue(8),
			IntValue(12),
		})
		if err != nil {
			t.Fatalf("builtinSubString() error = %v", err)
		}
		if result.AsString() != "World" {
			t.Errorf("builtinSubString('Hello, World!', 8, 12) = %v, want 'World'", result.AsString())
		}
	})

	t.Run("SubString empty", func(t *testing.T) {
		result, err := builtinSubString(vm, []Value{
			StringValue("Hello"),
			IntValue(5),
			IntValue(3),
		})
		if err != nil {
			t.Fatalf("builtinSubString() error = %v", err)
		}
		if result.AsString() != "" {
			t.Errorf("builtinSubString('Hello', 5, 3) = %v, want ''", result.AsString())
		}
	})

	t.Run("LeftStr basic", func(t *testing.T) {
		result, err := builtinLeftStr(vm, []Value{
			StringValue("Hello, World!"),
			IntValue(5),
		})
		if err != nil {
			t.Fatalf("builtinLeftStr() error = %v", err)
		}
		if result.AsString() != "Hello" {
			t.Errorf("builtinLeftStr('Hello, World!', 5) = %v, want 'Hello'", result.AsString())
		}
	})

	t.Run("LeftStr zero", func(t *testing.T) {
		result, err := builtinLeftStr(vm, []Value{
			StringValue("Hello"),
			IntValue(0),
		})
		if err != nil {
			t.Fatalf("builtinLeftStr() error = %v", err)
		}
		if result.AsString() != "" {
			t.Errorf("builtinLeftStr('Hello', 0) = %v, want ''", result.AsString())
		}
	})

	t.Run("RightStr basic", func(t *testing.T) {
		result, err := builtinRightStr(vm, []Value{
			StringValue("Hello, World!"),
			IntValue(6),
		})
		if err != nil {
			t.Fatalf("builtinRightStr() error = %v", err)
		}
		if result.AsString() != "World!" {
			t.Errorf("builtinRightStr('Hello, World!', 6) = %v, want 'World!'", result.AsString())
		}
	})

	t.Run("RightStr zero", func(t *testing.T) {
		result, err := builtinRightStr(vm, []Value{
			StringValue("Hello"),
			IntValue(0),
		})
		if err != nil {
			t.Fatalf("builtinRightStr() error = %v", err)
		}
		if result.AsString() != "" {
			t.Errorf("builtinRightStr('Hello', 0) = %v, want ''", result.AsString())
		}
	})

	t.Run("MidStr basic", func(t *testing.T) {
		result, err := builtinMidStr(vm, []Value{
			StringValue("Hello, World!"),
			IntValue(8),
			IntValue(5),
		})
		if err != nil {
			t.Fatalf("builtinMidStr() error = %v", err)
		}
		if result.AsString() != "World" {
			t.Errorf("builtinMidStr('Hello, World!', 8, 5) = %v, want 'World'", result.AsString())
		}
	})

	t.Run("RevPos found", func(t *testing.T) {
		result, err := builtinRevPos(vm, []Value{
			StringValue("o"),
			StringValue("Hello, World!"),
		})
		if err != nil {
			t.Fatalf("builtinRevPos() error = %v", err)
		}
		if result.AsInt() != 9 {
			t.Errorf("builtinRevPos('o', 'Hello, World!') = %v, want 9", result.AsInt())
		}
	})

	t.Run("RevPos not found", func(t *testing.T) {
		result, err := builtinRevPos(vm, []Value{
			StringValue("x"),
			StringValue("Hello, World!"),
		})
		if err != nil {
			t.Fatalf("builtinRevPos() error = %v", err)
		}
		if result.AsInt() != 0 {
			t.Errorf("builtinRevPos('x', 'Hello, World!') = %v, want 0", result.AsInt())
		}
	})

	t.Run("StrFind found", func(t *testing.T) {
		result, err := builtinStrFind(vm, []Value{
			StringValue("Hello, World!"),
			StringValue("World"),
			IntValue(1),
		})
		if err != nil {
			t.Fatalf("builtinStrFind() error = %v", err)
		}
		if result.AsInt() != 8 {
			t.Errorf("builtinStrFind('Hello, World!', 'World', 1) = %v, want 8", result.AsInt())
		}
	})

	t.Run("StrJoin basic", func(t *testing.T) {
		arr := NewArrayInstance([]Value{
			StringValue("Hello"),
			StringValue("World"),
			StringValue("Test"),
		})
		result, err := builtinStrJoin(vm, []Value{
			ArrayValue(arr),
			StringValue(", "),
		})
		if err != nil {
			t.Fatalf("builtinStrJoin() error = %v", err)
		}
		if result.AsString() != "Hello, World, Test" {
			t.Errorf("builtinStrJoin() = %v, want 'Hello, World, Test'", result.AsString())
		}
	})

	t.Run("StrJoin empty array", func(t *testing.T) {
		arr := NewArrayInstance([]Value{})
		result, err := builtinStrJoin(vm, []Value{
			ArrayValue(arr),
			StringValue(", "),
		})
		if err != nil {
			t.Fatalf("builtinStrJoin() error = %v", err)
		}
		if result.AsString() != "" {
			t.Errorf("builtinStrJoin(empty) = %v, want ''", result.AsString())
		}
	})

	t.Run("StrArrayPack basic", func(t *testing.T) {
		// StrArrayPack takes an array and removes empty strings
		arr := NewArrayInstance([]Value{
			StringValue("abc"),
			StringValue(""),
			StringValue("def"),
			StringValue(""),
			StringValue("ghi"),
		})
		result, err := builtinStrArrayPack(vm, []Value{
			ArrayValue(arr),
		})
		if err != nil {
			t.Fatalf("builtinStrArrayPack() error = %v", err)
		}
		if !result.IsArray() {
			t.Fatalf("builtinStrArrayPack() should return array")
		}
		packed := result.AsArray()
		if packed.Length() != 3 {
			t.Errorf("builtinStrArrayPack() returned array with %d elements, want 3", packed.Length())
		}
	})

	t.Run("StrBeforeLast found", func(t *testing.T) {
		result, err := builtinStrBeforeLast(vm, []Value{
			StringValue("Hello, World!"),
			StringValue("o"),
		})
		if err != nil {
			t.Fatalf("builtinStrBeforeLast() error = %v", err)
		}
		if result.AsString() != "Hello, W" {
			t.Errorf("builtinStrBeforeLast('Hello, World!', 'o') = %v, want 'Hello, W'", result.AsString())
		}
	})

	t.Run("StrBeforeLast not found", func(t *testing.T) {
		result, err := builtinStrBeforeLast(vm, []Value{
			StringValue("Hello, World!"),
			StringValue("x"),
		})
		if err != nil {
			t.Fatalf("builtinStrBeforeLast() error = %v", err)
		}
		if result.AsString() != "Hello, World!" {
			t.Errorf("builtinStrBeforeLast('Hello, World!', 'x') = %v, want 'Hello, World!'", result.AsString())
		}
	})

	t.Run("StrBefore not found returns original", func(t *testing.T) {
		result, err := builtinStrBefore(vm, []Value{
			StringValue("Hello, World!"),
			StringValue("x"),
		})
		if err != nil {
			t.Fatalf("builtinStrBefore() error = %v", err)
		}
		if result.AsString() != "Hello, World!" {
			t.Errorf("builtinStrBefore('Hello, World!', 'x') = %v, want 'Hello, World!'", result.AsString())
		}
	})

	t.Run("StrBefore empty delimiter returns original", func(t *testing.T) {
		result, err := builtinStrBefore(vm, []Value{
			StringValue("Hello, World!"),
			StringValue(""),
		})
		if err != nil {
			t.Fatalf("builtinStrBefore() error = %v", err)
		}
		if result.AsString() != "Hello, World!" {
			t.Errorf("builtinStrBefore('Hello, World!', '') = %v, want 'Hello, World!'", result.AsString())
		}
	})

	t.Run("StrAfterLast found", func(t *testing.T) {
		result, err := builtinStrAfterLast(vm, []Value{
			StringValue("Hello, World!"),
			StringValue("o"),
		})
		if err != nil {
			t.Fatalf("builtinStrAfterLast() error = %v", err)
		}
		if result.AsString() != "rld!" {
			t.Errorf("builtinStrAfterLast('Hello, World!', 'o') = %v, want 'rld!'", result.AsString())
		}
	})

	t.Run("StrAfterLast not found", func(t *testing.T) {
		result, err := builtinStrAfterLast(vm, []Value{
			StringValue("Hello, World!"),
			StringValue("x"),
		})
		if err != nil {
			t.Fatalf("builtinStrAfterLast() error = %v", err)
		}
		if result.AsString() != "" {
			t.Errorf("builtinStrAfterLast('Hello, World!', 'x') = %v, want ''", result.AsString())
		}
	})

	t.Run("StrBetween found", func(t *testing.T) {
		result, err := builtinStrBetween(vm, []Value{
			StringValue("This is <example> text"),
			StringValue("<"),
			StringValue(">"),
		})
		if err != nil {
			t.Fatalf("builtinStrBetween() error = %v", err)
		}
		if result.AsString() != "example" {
			t.Errorf("builtinStrBetween('This is <example> text', '<', '>') = %v, want 'example'", result.AsString())
		}
	})

	t.Run("IsDelimiter true", func(t *testing.T) {
		result, err := builtinIsDelimiter(vm, []Value{
			StringValue(" ,;"),
			StringValue(" "),
			IntValue(1),
		})
		if err != nil {
			t.Fatalf("builtinIsDelimiter() error = %v", err)
		}
		if !result.AsBool() {
			t.Errorf("builtinIsDelimiter(' ,;', ' ', 1) = false, want true")
		}
	})

	t.Run("IsDelimiter false", func(t *testing.T) {
		result, err := builtinIsDelimiter(vm, []Value{
			StringValue(" ,;"),
			StringValue("a"),
			IntValue(1),
		})
		if err != nil {
			t.Fatalf("builtinIsDelimiter() error = %v", err)
		}
		if result.AsBool() {
			t.Errorf("builtinIsDelimiter(' ,;', 'a', 1) = true, want false")
		}
	})

	t.Run("LastDelimiter found", func(t *testing.T) {
		result, err := builtinLastDelimiter(vm, []Value{
			StringValue(" ,;"),
			StringValue("Hello, World!"),
		})
		if err != nil {
			t.Fatalf("builtinLastDelimiter() error = %v", err)
		}
		if result.AsInt() != 7 {
			t.Errorf("builtinLastDelimiter(' ,;', 'Hello, World!') = %v, want 7", result.AsInt())
		}
	})

	t.Run("FindDelimiter found", func(t *testing.T) {
		result, err := builtinFindDelimiter(vm, []Value{
			StringValue(" ,;"),
			StringValue("Hello, World!"),
			IntValue(1),
		})
		if err != nil {
			t.Fatalf("builtinFindDelimiter() error = %v", err)
		}
		if result.AsInt() != 6 {
			t.Errorf("builtinFindDelimiter(' ,;', 'Hello, World!', 1) = %v, want 6", result.AsInt())
		}
	})

	t.Run("PadLeft basic", func(t *testing.T) {
		result, err := builtinPadLeft(vm, []Value{
			StringValue("Hi"),
			IntValue(5),
			StringValue(" "),
		})
		if err != nil {
			t.Fatalf("builtinPadLeft() error = %v", err)
		}
		if result.AsString() != "   Hi" {
			t.Errorf("builtinPadLeft('Hi', 5, ' ') = %v, want '   Hi'", result.AsString())
		}
	})

	t.Run("PadLeft no padding needed", func(t *testing.T) {
		result, err := builtinPadLeft(vm, []Value{
			StringValue("Hello"),
			IntValue(3),
			StringValue(" "),
		})
		if err != nil {
			t.Fatalf("builtinPadLeft() error = %v", err)
		}
		if result.AsString() != "Hello" {
			t.Errorf("builtinPadLeft('Hello', 3, ' ') = %v, want 'Hello'", result.AsString())
		}
	})

	t.Run("PadRight basic", func(t *testing.T) {
		result, err := builtinPadRight(vm, []Value{
			StringValue("Hi"),
			IntValue(5),
			StringValue(" "),
		})
		if err != nil {
			t.Fatalf("builtinPadRight() error = %v", err)
		}
		if result.AsString() != "Hi   " {
			t.Errorf("builtinPadRight('Hi', 5, ' ') = %v, want 'Hi   '", result.AsString())
		}
	})

	t.Run("PadRight no padding needed", func(t *testing.T) {
		result, err := builtinPadRight(vm, []Value{
			StringValue("Hello"),
			IntValue(3),
			StringValue(" "),
		})
		if err != nil {
			t.Fatalf("builtinPadRight() error = %v", err)
		}
		if result.AsString() != "Hello" {
			t.Errorf("builtinPadRight('Hello', 3, ' ') = %v, want 'Hello'", result.AsString())
		}
	})
}
