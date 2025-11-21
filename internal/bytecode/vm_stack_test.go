package bytecode

import (
	"testing"
)

// TestGetGlobal tests the getGlobal function
func TestGetGlobal(t *testing.T) {
	vm := NewVM()

	t.Run("get from empty globals", func(t *testing.T) {
		result := vm.getGlobal(0)
		if !result.IsNil() {
			t.Errorf("getGlobal(0) from empty globals = %v, want nil", result)
		}
	})

	t.Run("get negative index", func(t *testing.T) {
		result := vm.getGlobal(-1)
		if !result.IsNil() {
			t.Errorf("getGlobal(-1) = %v, want nil", result)
		}
	})

	t.Run("get out of range index", func(t *testing.T) {
		vm.setGlobal(0, IntValue(42))
		result := vm.getGlobal(10)
		if !result.IsNil() {
			t.Errorf("getGlobal(10) with 1 global = %v, want nil", result)
		}
	})

	t.Run("get valid global", func(t *testing.T) {
		vm.setGlobal(0, IntValue(42))
		result := vm.getGlobal(0)
		if !result.IsInt() || result.AsInt() != 42 {
			t.Errorf("getGlobal(0) = %v, want 42", result)
		}
	})

	t.Run("get zero-value global", func(t *testing.T) {
		vm.globals = make([]Value, 1)
		// Default zero value should be treated as nil
		result := vm.getGlobal(0)
		if !result.IsNil() {
			t.Errorf("getGlobal(0) with zero value = %v, want nil", result)
		}
	})
}

// TestSetGlobal tests the setGlobal function
func TestSetGlobal(t *testing.T) {
	t.Run("set at index 0", func(t *testing.T) {
		vm := NewVM()
		vm.setGlobal(0, IntValue(42))
		if len(vm.globals) != 1 {
			t.Errorf("globals length = %d, want 1", len(vm.globals))
		}
		if vm.globals[0].AsInt() != 42 {
			t.Errorf("global[0] = %v, want 42", vm.globals[0])
		}
	})

	t.Run("set negative index", func(t *testing.T) {
		vm := NewVM()
		vm.setGlobal(-1, IntValue(42))
		if len(vm.globals) != 0 {
			t.Errorf("globals length = %d, want 0 (negative index should be ignored)", len(vm.globals))
		}
	})

	t.Run("set expands globals array", func(t *testing.T) {
		vm := NewVM()
		vm.setGlobal(5, IntValue(99))
		if len(vm.globals) != 6 {
			t.Errorf("globals length = %d, want 6", len(vm.globals))
		}
		if vm.globals[5].AsInt() != 99 {
			t.Errorf("global[5] = %v, want 99", vm.globals[5])
		}
	})

	t.Run("set preserves existing values", func(t *testing.T) {
		vm := NewVM()
		vm.setGlobal(0, IntValue(10))
		vm.setGlobal(1, IntValue(20))
		vm.setGlobal(5, IntValue(50))

		if vm.globals[0].AsInt() != 10 {
			t.Errorf("global[0] = %v, want 10", vm.globals[0])
		}
		if vm.globals[1].AsInt() != 20 {
			t.Errorf("global[1] = %v, want 20", vm.globals[1])
		}
		if vm.globals[5].AsInt() != 50 {
			t.Errorf("global[5] = %v, want 50", vm.globals[5])
		}
	})

	t.Run("set overwrites existing value", func(t *testing.T) {
		vm := NewVM()
		vm.setGlobal(0, IntValue(42))
		vm.setGlobal(0, IntValue(99))
		if vm.globals[0].AsInt() != 99 {
			t.Errorf("global[0] = %v, want 99", vm.globals[0])
		}
	})
}

// TestPublicGlobalAccessors tests the public SetGlobal and GetGlobal methods
func TestPublicGlobalAccessors(t *testing.T) {
	vm := NewVM()

	t.Run("SetGlobal and GetGlobal", func(t *testing.T) {
		vm.SetGlobal(0, IntValue(42))
		result := vm.GetGlobal(0)
		if result.AsInt() != 42 {
			t.Errorf("GetGlobal(0) = %v, want 42", result)
		}
	})

	t.Run("SetGlobal multiple values", func(t *testing.T) {
		vm.SetGlobal(0, StringValue("hello"))
		vm.SetGlobal(1, FloatValue(3.14))
		vm.SetGlobal(2, BoolValue(true))

		if vm.GetGlobal(0).AsString() != "hello" {
			t.Errorf("GetGlobal(0) = %v, want 'hello'", vm.GetGlobal(0))
		}
		if vm.GetGlobal(1).AsFloat() != 3.14 {
			t.Errorf("GetGlobal(1) = %v, want 3.14", vm.GetGlobal(1))
		}
		if !vm.GetGlobal(2).AsBool() {
			t.Errorf("GetGlobal(2) = %v, want true", vm.GetGlobal(2))
		}
	})
}

// TestPop tests the pop function
func TestPop(t *testing.T) {
	t.Run("pop from empty stack", func(t *testing.T) {
		vm := NewVM()
		_, err := vm.pop()
		if err == nil {
			t.Error("pop() from empty stack should return error")
		}
	})

	t.Run("pop single value", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(42))
		result, err := vm.pop()
		if err != nil {
			t.Fatalf("pop() error: %v", err)
		}
		if result.AsInt() != 42 {
			t.Errorf("pop() = %v, want 42", result)
		}
		if len(vm.stack) != 0 {
			t.Errorf("stack length = %d, want 0", len(vm.stack))
		}
	})

	t.Run("pop multiple values LIFO", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(1))
		vm.push(IntValue(2))
		vm.push(IntValue(3))

		v3, _ := vm.pop()
		v2, _ := vm.pop()
		v1, _ := vm.pop()

		if v3.AsInt() != 3 || v2.AsInt() != 2 || v1.AsInt() != 1 {
			t.Errorf("pop order = %v, %v, %v, want 3, 2, 1", v3, v2, v1)
		}
	})

	t.Run("pop until empty then error", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(42))
		_, err := vm.pop()
		if err != nil {
			t.Fatalf("first pop() error: %v", err)
		}
		_, err = vm.pop()
		if err == nil {
			t.Error("pop() from empty stack should return error")
		}
	})
}

// TestPeek tests the peek function
func TestPeek(t *testing.T) {
	t.Run("peek from empty stack", func(t *testing.T) {
		vm := NewVM()
		_, err := vm.peek()
		if err == nil {
			t.Error("peek() from empty stack should return error")
		}
	})

	t.Run("peek does not remove value", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(42))

		result1, err := vm.peek()
		if err != nil {
			t.Fatalf("peek() error: %v", err)
		}
		if result1.AsInt() != 42 {
			t.Errorf("peek() = %v, want 42", result1)
		}

		result2, err := vm.peek()
		if err != nil {
			t.Fatalf("second peek() error: %v", err)
		}
		if result2.AsInt() != 42 {
			t.Errorf("second peek() = %v, want 42", result2)
		}

		if len(vm.stack) != 1 {
			t.Errorf("stack length = %d, want 1 (peek should not remove)", len(vm.stack))
		}
	})

	t.Run("peek returns top value", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(1))
		vm.push(IntValue(2))
		vm.push(IntValue(3))

		result, err := vm.peek()
		if err != nil {
			t.Fatalf("peek() error: %v", err)
		}
		if result.AsInt() != 3 {
			t.Errorf("peek() = %v, want 3 (top of stack)", result)
		}
		if len(vm.stack) != 3 {
			t.Errorf("stack length = %d, want 3", len(vm.stack))
		}
	})
}

// TestTrimStack tests the trimStack function
func TestTrimStack(t *testing.T) {
	t.Run("trim to zero", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(1))
		vm.push(IntValue(2))
		vm.push(IntValue(3))
		vm.trimStack(0)
		if len(vm.stack) != 0 {
			t.Errorf("stack length = %d, want 0", len(vm.stack))
		}
	})

	t.Run("trim negative depth becomes zero", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(1))
		vm.push(IntValue(2))
		vm.trimStack(-5)
		if len(vm.stack) != 0 {
			t.Errorf("stack length = %d, want 0 (negative depth should become 0)", len(vm.stack))
		}
	})

	t.Run("trim to specific depth", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(1))
		vm.push(IntValue(2))
		vm.push(IntValue(3))
		vm.push(IntValue(4))
		vm.push(IntValue(5))

		vm.trimStack(2)
		if len(vm.stack) != 2 {
			t.Errorf("stack length = %d, want 2", len(vm.stack))
		}
		if vm.stack[0].AsInt() != 1 || vm.stack[1].AsInt() != 2 {
			t.Errorf("remaining stack = %v, want [1, 2]", vm.stack)
		}
	})

	t.Run("trim depth greater than stack length", func(t *testing.T) {
		vm := NewVM()
		vm.push(IntValue(1))
		vm.push(IntValue(2))

		vm.trimStack(10)
		if len(vm.stack) != 2 {
			t.Errorf("stack length = %d, want 2 (trim should not add)", len(vm.stack))
		}
	})

	t.Run("trim empty stack", func(t *testing.T) {
		vm := NewVM()
		vm.trimStack(0)
		if len(vm.stack) != 0 {
			t.Errorf("stack length = %d, want 0", len(vm.stack))
		}
	})

	t.Run("trim preserves lower stack values", func(t *testing.T) {
		vm := NewVM()
		for i := 1; i <= 10; i++ {
			vm.push(IntValue(int64(i)))
		}

		vm.trimStack(5)
		if len(vm.stack) != 5 {
			t.Errorf("stack length = %d, want 5", len(vm.stack))
		}
		for i := 0; i < 5; i++ {
			if vm.stack[i].AsInt() != int64(i+1) {
				t.Errorf("stack[%d] = %v, want %d", i, vm.stack[i], i+1)
			}
		}
	})
}

// TestStackPushPop tests push and pop together
func TestStackPushPop(t *testing.T) {
	vm := NewVM()

	t.Run("push and pop sequence", func(t *testing.T) {
		values := []Value{
			IntValue(1),
			FloatValue(2.5),
			StringValue("hello"),
			BoolValue(true),
			NilValue(),
		}

		for _, v := range values {
			vm.push(v)
		}

		// Pop in reverse order
		for i := len(values) - 1; i >= 0; i-- {
			result, err := vm.pop()
			if err != nil {
				t.Fatalf("pop() error: %v", err)
			}
			if result.Type != values[i].Type {
				t.Errorf("pop() type = %v, want %v", result.Type, values[i].Type)
			}
		}

		if len(vm.stack) != 0 {
			t.Errorf("stack length = %d, want 0 after all pops", len(vm.stack))
		}
	})
}
