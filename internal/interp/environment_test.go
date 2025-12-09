package interp

import (
	"testing"
)

// TestNewEnvironment verifies that a new environment is created correctly.
func TestNewEnvironment(t *testing.T) {
	env := NewEnvironment()

	if env == nil {
		t.Fatal("NewEnvironment() returned nil")
	}

	// Verify the environment is properly initialized (store not nil)
	// We can't access store directly anymore, but Size() will work if store exists
	_ = env.Size()

	if env.Outer() != nil {
		t.Error("Root environment should have no outer environment")
	}

	if env.Size() != 0 {
		t.Errorf("New environment should be empty, got size %d", env.Size())
	}
}

// TestDefineAndGet verifies that we can define variables and retrieve them.
func TestDefineAndGet(t *testing.T) {
	env := NewEnvironment()

	// Define a variable
	env.Define("x", NewIntegerValue(42))

	// Retrieve it
	val, ok := env.Get("x")
	if !ok {
		t.Fatal("Variable 'x' not found after definition")
	}

	intVal, err := GoInt(val)
	if err != nil {
		t.Fatalf("Expected integer value, got: %v", err)
	}

	if intVal != 42 {
		t.Errorf("Expected value 42, got %d", intVal)
	}
}

// TestGetUndefined verifies that getting an undefined variable returns false.
func TestGetUndefined(t *testing.T) {
	env := NewEnvironment()

	val, ok := env.Get("undefined")
	if ok {
		t.Errorf("Expected undefined variable to return false, got value: %v", val)
	}

	if val != nil {
		t.Errorf("Expected nil value for undefined variable, got: %v", val)
	}
}

// TestDefineMultipleVariables verifies we can define multiple variables.
func TestDefineMultipleVariables(t *testing.T) {
	env := NewEnvironment()

	env.Define("a", NewIntegerValue(1))
	env.Define("b", NewStringValue("hello"))
	env.Define("c", NewBooleanValue(true))

	if env.Size() != 3 {
		t.Errorf("Expected 3 variables, got %d", env.Size())
	}

	// Verify each variable
	if val, ok := env.Get("a"); !ok {
		t.Error("Variable 'a' not found")
	} else if intVal, _ := GoInt(val); intVal != 1 {
		t.Errorf("Expected 'a' = 1, got %d", intVal)
	}

	if val, ok := env.Get("b"); !ok {
		t.Error("Variable 'b' not found")
	} else if strVal, _ := GoString(val); strVal != "hello" {
		t.Errorf("Expected 'b' = 'hello', got %s", strVal)
	}

	if val, ok := env.Get("c"); !ok {
		t.Error("Variable 'c' not found")
	} else if boolVal, _ := GoBool(val); !boolVal {
		t.Error("Expected 'c' = true, got false")
	}
}

// TestDefineOverwrite verifies that defining a variable twice overwrites it.
func TestDefineOverwrite(t *testing.T) {
	env := NewEnvironment()

	env.Define("x", NewIntegerValue(10))
	env.Define("x", NewIntegerValue(20))

	val, ok := env.Get("x")
	if !ok {
		t.Fatal("Variable 'x' not found")
	}

	intVal, _ := GoInt(val)
	if intVal != 20 {
		t.Errorf("Expected overwritten value 20, got %d", intVal)
	}

	if env.Size() != 1 {
		t.Errorf("Expected size 1 after overwrite, got %d", env.Size())
	}
}

// TestSet verifies that we can update existing variables with Set.
func TestSet(t *testing.T) {
	env := NewEnvironment()

	// Define a variable
	env.Define("x", NewIntegerValue(10))

	// Update it
	err := env.Set("x", NewIntegerValue(20))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify the update
	val, _ := env.Get("x")
	intVal, _ := GoInt(val)
	if intVal != 20 {
		t.Errorf("Expected updated value 20, got %d", intVal)
	}
}

// TestSetUndefined verifies that Set returns an error for undefined variables.
func TestSetUndefined(t *testing.T) {
	env := NewEnvironment()

	err := env.Set("undefined", NewIntegerValue(42))
	if err == nil {
		t.Error("Expected error when setting undefined variable")
	}
}

// TestHas verifies the Has method.
func TestHas(t *testing.T) {
	env := NewEnvironment()

	if env.Has("x") {
		t.Error("Expected Has('x') to return false for undefined variable")
	}

	env.Define("x", NewIntegerValue(42))

	if !env.Has("x") {
		t.Error("Expected Has('x') to return true after definition")
	}
}

// TestNewEnclosedEnvironment verifies that enclosed environments are created correctly.
func TestNewEnclosedEnvironment(t *testing.T) {
	outer := NewEnvironment()
	inner := NewEnclosedEnvironment(outer)

	if inner.Outer() != outer {
		t.Error("Enclosed environment does not reference correct outer environment")
	}

	if inner.Size() != 0 {
		t.Errorf("New enclosed environment should be empty, got size %d", inner.Size())
	}
}

// TestNestedScope verifies that inner scopes can access outer scope variables.
func TestNestedScope(t *testing.T) {
	outer := NewEnvironment()
	outer.Define("global", NewStringValue("I'm global"))

	inner := NewEnclosedEnvironment(outer)
	inner.Define("local", NewStringValue("I'm local"))

	// Inner scope should see both variables
	if val, ok := inner.Get("global"); !ok {
		t.Error("Inner scope cannot access outer variable 'global'")
	} else if str, _ := GoString(val); str != "I'm global" {
		t.Errorf("Wrong value for 'global': %s", str)
	}

	if val, ok := inner.Get("local"); !ok {
		t.Error("Inner scope cannot access local variable 'local'")
	} else if str, _ := GoString(val); str != "I'm local" {
		t.Errorf("Wrong value for 'local': %s", str)
	}

	// Outer scope should NOT see inner variables
	if _, ok := outer.Get("local"); ok {
		t.Error("Outer scope should not access inner variable 'local'")
	}

	// But outer should still see its own variables
	if _, ok := outer.Get("global"); !ok {
		t.Error("Outer scope lost its own variable 'global'")
	}
}

// TestNestedScopeSet verifies that Set updates variables in the correct scope.
func TestNestedScopeSet(t *testing.T) {
	outer := NewEnvironment()
	outer.Define("x", NewIntegerValue(10))

	inner := NewEnclosedEnvironment(outer)

	// Update outer variable from inner scope
	err := inner.Set("x", NewIntegerValue(20))
	if err != nil {
		t.Fatalf("Failed to set outer variable from inner scope: %v", err)
	}

	// Verify the update is visible from both scopes
	val, _ := outer.Get("x")
	if intVal, _ := GoInt(val); intVal != 20 {
		t.Errorf("Outer scope: expected x=20, got %d", intVal)
	}

	val, _ = inner.Get("x")
	if intVal, _ := GoInt(val); intVal != 20 {
		t.Errorf("Inner scope: expected x=20, got %d", intVal)
	}
}

// TestShadowing verifies that inner scopes can shadow outer variables.
func TestShadowing(t *testing.T) {
	outer := NewEnvironment()
	outer.Define("x", NewIntegerValue(10))

	inner := NewEnclosedEnvironment(outer)
	inner.Define("x", NewIntegerValue(20)) // Shadow outer 'x'

	// Inner scope sees its own 'x'
	val, _ := inner.Get("x")
	if intVal, _ := GoInt(val); intVal != 20 {
		t.Errorf("Inner scope: expected x=20, got %d", intVal)
	}

	// Outer scope still sees its own 'x'
	val, _ = outer.Get("x")
	if intVal, _ := GoInt(val); intVal != 10 {
		t.Errorf("Outer scope: expected x=10, got %d", intVal)
	}
}

// TestGetLocal verifies the GetLocal method only searches current scope.
func TestGetLocal(t *testing.T) {
	outer := NewEnvironment()
	outer.Define("outerVar", NewIntegerValue(10))

	inner := NewEnclosedEnvironment(outer)
	inner.Define("innerVar", NewIntegerValue(20))

	// GetLocal should find local variable
	if val, ok := inner.GetLocal("innerVar"); !ok {
		t.Error("GetLocal failed to find local variable")
	} else if intVal, _ := GoInt(val); intVal != 20 {
		t.Errorf("GetLocal returned wrong value: %d", intVal)
	}

	// GetLocal should NOT find outer variable
	if _, ok := inner.GetLocal("outerVar"); ok {
		t.Error("GetLocal should not find outer variable")
	}

	// But Get should find it
	if _, ok := inner.Get("outerVar"); !ok {
		t.Error("Get failed to find outer variable")
	}
}

// TestDeeplyScopedNest verifies multiple levels of nesting work correctly.
func TestDeeplyNestedScopes(t *testing.T) {
	// Create a chain: global -> func -> block
	global := NewEnvironment()
	global.Define("level", NewStringValue("global"))

	funcScope := NewEnclosedEnvironment(global)
	funcScope.Define("funcVar", NewIntegerValue(1))

	blockScope := NewEnclosedEnvironment(funcScope)
	blockScope.Define("blockVar", NewIntegerValue(2))

	// Block scope should see all three levels
	if _, ok := blockScope.Get("level"); !ok {
		t.Error("Block scope cannot see global variable")
	}

	if _, ok := blockScope.Get("funcVar"); !ok {
		t.Error("Block scope cannot see function scope variable")
	}

	if _, ok := blockScope.Get("blockVar"); !ok {
		t.Error("Block scope cannot see its own variable")
	}

	// Function scope should not see block variables
	if _, ok := funcScope.Get("blockVar"); ok {
		t.Error("Function scope should not see block variable")
	}

	// Global should not see function or block variables
	if _, ok := global.Get("funcVar"); ok {
		t.Error("Global scope should not see function variable")
	}

	if _, ok := global.Get("blockVar"); ok {
		t.Error("Global scope should not see block variable")
	}
}

// TestSetInNestedScope verifies Set works through multiple nesting levels.
func TestSetInNestedScope(t *testing.T) {
	global := NewEnvironment()
	global.Define("x", NewIntegerValue(1))

	level1 := NewEnclosedEnvironment(global)
	level2 := NewEnclosedEnvironment(level1)
	level3 := NewEnclosedEnvironment(level2)

	// Update from deeply nested scope
	err := level3.Set("x", NewIntegerValue(100))
	if err != nil {
		t.Fatalf("Failed to set variable from nested scope: %v", err)
	}

	// Verify update is visible from all levels
	val, _ := global.Get("x")
	if intVal, _ := GoInt(val); intVal != 100 {
		t.Errorf("Global: expected x=100, got %d", intVal)
	}

	val, _ = level3.Get("x")
	if intVal, _ := GoInt(val); intVal != 100 {
		t.Errorf("Level3: expected x=100, got %d", intVal)
	}
}

// TestEnvironmentIsolation verifies that sibling scopes are isolated.
func TestEnvironmentIsolation(t *testing.T) {
	parent := NewEnvironment()
	parent.Define("shared", NewIntegerValue(0))

	child1 := NewEnclosedEnvironment(parent)
	child1.Define("child1Var", NewStringValue("only in child1"))

	child2 := NewEnclosedEnvironment(parent)
	child2.Define("child2Var", NewStringValue("only in child2"))

	// Child1 should not see child2's variables
	if _, ok := child1.Get("child2Var"); ok {
		t.Error("Child1 should not see child2's variable")
	}

	// Child2 should not see child1's variables
	if _, ok := child2.Get("child1Var"); ok {
		t.Error("Child2 should not see child1's variable")
	}

	// Both should see parent's variables
	if _, ok := child1.Get("shared"); !ok {
		t.Error("Child1 cannot see parent's variable")
	}

	if _, ok := child2.Get("shared"); !ok {
		t.Error("Child2 cannot see parent's variable")
	}
}

// TestDifferentValueTypes verifies environment works with all value types.
func TestDifferentValueTypes(t *testing.T) {
	env := NewEnvironment()

	env.Define("int", NewIntegerValue(42))
	env.Define("float", NewFloatValue(3.14))
	env.Define("str", NewStringValue("hello"))
	env.Define("bool", NewBooleanValue(true))
	env.Define("nil", NewNilValue())

	// Verify each type
	if val, ok := env.Get("int"); !ok || val.Type() != "INTEGER" {
		t.Error("Integer value not stored correctly")
	}

	if val, ok := env.Get("float"); !ok || val.Type() != "FLOAT" {
		t.Error("Float value not stored correctly")
	}

	if val, ok := env.Get("str"); !ok || val.Type() != "STRING" {
		t.Error("String value not stored correctly")
	}

	if val, ok := env.Get("bool"); !ok || val.Type() != "BOOLEAN" {
		t.Error("Boolean value not stored correctly")
	}

	if val, ok := env.Get("nil"); !ok || val.Type() != "NIL" {
		t.Error("Nil value not stored correctly")
	}
}
