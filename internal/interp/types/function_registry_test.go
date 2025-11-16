package types

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
)

// Helper to create a simple function declaration for testing
func makeFunctionDecl(name string, paramCount int) *ast.FunctionDecl {
	params := make([]*ast.Parameter, paramCount)
	for i := 0; i < paramCount; i++ {
		params[i] = &ast.Parameter{
			Name: &ast.Identifier{Value: "param"},
		}
	}

	return &ast.FunctionDecl{
		Name:       &ast.Identifier{Value: name},
		Parameters: params,
	}
}

func TestFunctionRegistry_RegisterAndLookup(t *testing.T) {
	registry := NewFunctionRegistry()

	// Create test functions
	fn1 := makeFunctionDecl("TestFunc", 0)
	fn2 := makeFunctionDecl("TestFunc", 1) // Overload with 1 param

	// Register functions
	registry.Register("TestFunc", fn1)
	registry.Register("TestFunc", fn2)

	// Lookup should return both overloads
	overloads := registry.Lookup("TestFunc")
	if len(overloads) != 2 {
		t.Errorf("Expected 2 overloads, got %d", len(overloads))
	}

	// Lookup is case-insensitive
	overloads = registry.Lookup("testfunc")
	if len(overloads) != 2 {
		t.Errorf("Case-insensitive lookup failed: expected 2 overloads, got %d", len(overloads))
	}

	overloads = registry.Lookup("TESTFUNC")
	if len(overloads) != 2 {
		t.Errorf("Uppercase lookup failed: expected 2 overloads, got %d", len(overloads))
	}
}

func TestFunctionRegistry_Exists(t *testing.T) {
	registry := NewFunctionRegistry()
	fn := makeFunctionDecl("MyFunc", 0)

	registry.Register("MyFunc", fn)

	tests := []struct {
		name     string
		lookup   string
		expected bool
	}{
		{"exact case", "MyFunc", true},
		{"lowercase", "myfunc", true},
		{"uppercase", "MYFUNC", true},
		{"not found", "OtherFunc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := registry.Exists(tt.lookup)
			if exists != tt.expected {
				t.Errorf("Exists(%q) = %v, expected %v", tt.lookup, exists, tt.expected)
			}
		})
	}
}

func TestFunctionRegistry_RegisterWithUnit(t *testing.T) {
	registry := NewFunctionRegistry()

	fn1 := makeFunctionDecl("Abs", 1)
	fn2 := makeFunctionDecl("Abs", 1) // Different implementation

	// Register Math.Abs
	registry.RegisterWithUnit("Math", "Abs", fn1)

	// Should be available via global lookup
	if !registry.Exists("Abs") {
		t.Error("Function should exist in global namespace")
	}

	// Should be available via qualified lookup
	if !registry.ExistsQualified("Math", "Abs") {
		t.Error("Function should exist in qualified namespace")
	}

	// Lookup globally
	overloads := registry.Lookup("Abs")
	if len(overloads) != 1 {
		t.Errorf("Global lookup: expected 1 overload, got %d", len(overloads))
	}

	// Lookup qualified
	overloads = registry.LookupQualified("Math", "Abs")
	if len(overloads) != 1 {
		t.Errorf("Qualified lookup: expected 1 overload, got %d", len(overloads))
	}

	// Case-insensitive qualified lookup
	overloads = registry.LookupQualified("MATH", "abs")
	if len(overloads) != 1 {
		t.Errorf("Case-insensitive qualified lookup failed")
	}

	// Register another unit's Abs (different unit)
	registry.RegisterWithUnit("MyMath", "Abs", fn2)

	// Global lookup should now have 2
	overloads = registry.Lookup("Abs")
	if len(overloads) != 2 {
		t.Errorf("After second registration: expected 2 overloads, got %d", len(overloads))
	}

	// Each unit should still have 1
	overloads = registry.LookupQualified("Math", "Abs")
	if len(overloads) != 1 {
		t.Errorf("Math.Abs should have 1 overload, got %d", len(overloads))
	}

	overloads = registry.LookupQualified("MyMath", "Abs")
	if len(overloads) != 1 {
		t.Errorf("MyMath.Abs should have 1 overload, got %d", len(overloads))
	}
}

func TestFunctionRegistry_GetOverloadCount(t *testing.T) {
	registry := NewFunctionRegistry()

	fn1 := makeFunctionDecl("Func", 0)
	fn2 := makeFunctionDecl("Func", 1)
	fn3 := makeFunctionDecl("Func", 2)

	registry.Register("Func", fn1)
	registry.Register("Func", fn2)
	registry.Register("Func", fn3)

	count := registry.GetOverloadCount("Func")
	if count != 3 {
		t.Errorf("Expected 3 overloads, got %d", count)
	}

	count = registry.GetOverloadCount("NonExistent")
	if count != 0 {
		t.Errorf("Expected 0 for non-existent function, got %d", count)
	}
}

func TestFunctionRegistry_GetOverloadCountQualified(t *testing.T) {
	registry := NewFunctionRegistry()

	fn1 := makeFunctionDecl("Test", 0)
	fn2 := makeFunctionDecl("Test", 1)

	registry.RegisterWithUnit("Unit1", "Test", fn1)
	registry.RegisterWithUnit("Unit1", "Test", fn2)

	count := registry.GetOverloadCountQualified("Unit1", "Test")
	if count != 2 {
		t.Errorf("Expected 2 overloads, got %d", count)
	}

	count = registry.GetOverloadCountQualified("Unit2", "Test")
	if count != 0 {
		t.Errorf("Expected 0 for non-existent qualified function, got %d", count)
	}
}

func TestFunctionRegistry_GetAllFunctions(t *testing.T) {
	registry := NewFunctionRegistry()

	fn1 := makeFunctionDecl("Func1", 0)
	fn2 := makeFunctionDecl("Func2", 1)
	fn3 := makeFunctionDecl("Func2", 2) // Overload

	registry.Register("Func1", fn1)
	registry.Register("Func2", fn2)
	registry.Register("Func2", fn3)

	allFuncs := registry.GetAllFunctions()

	if len(allFuncs) != 2 {
		t.Errorf("Expected 2 unique functions, got %d", len(allFuncs))
	}

	if len(allFuncs["func1"]) != 1 {
		t.Errorf("Func1 should have 1 overload, got %d", len(allFuncs["func1"]))
	}

	if len(allFuncs["func2"]) != 2 {
		t.Errorf("Func2 should have 2 overloads, got %d", len(allFuncs["func2"]))
	}
}

func TestFunctionRegistry_GetFunctionNames(t *testing.T) {
	registry := NewFunctionRegistry()

	fn1 := makeFunctionDecl("Alpha", 0)
	fn2 := makeFunctionDecl("Beta", 0)
	fn3 := makeFunctionDecl("Beta", 1) // Overload

	registry.Register("Alpha", fn1)
	registry.Register("Beta", fn2)
	registry.Register("Beta", fn3)

	names := registry.GetFunctionNames()

	if len(names) != 2 {
		t.Errorf("Expected 2 unique names, got %d", len(names))
	}

	// Check that both names are present (order doesn't matter)
	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}

	if !nameSet["Alpha"] || !nameSet["Beta"] {
		t.Error("Expected names Alpha and Beta")
	}
}

func TestFunctionRegistry_Count(t *testing.T) {
	registry := NewFunctionRegistry()

	if count := registry.Count(); count != 0 {
		t.Errorf("Empty registry count = %d, expected 0", count)
	}

	fn1 := makeFunctionDecl("Func1", 0)
	registry.Register("Func1", fn1)

	if count := registry.Count(); count != 1 {
		t.Errorf("After 1 registration count = %d, expected 1", count)
	}

	fn2 := makeFunctionDecl("Func1", 1) // Overload
	registry.Register("Func1", fn2)

	// Count should still be 1 (same function name)
	if count := registry.Count(); count != 1 {
		t.Errorf("After overload count = %d, expected 1", count)
	}

	fn3 := makeFunctionDecl("Func2", 0)
	registry.Register("Func2", fn3)

	if count := registry.Count(); count != 2 {
		t.Errorf("After 2nd function count = %d, expected 2", count)
	}
}

func TestFunctionRegistry_TotalOverloads(t *testing.T) {
	registry := NewFunctionRegistry()

	if total := registry.TotalOverloads(); total != 0 {
		t.Errorf("Empty registry TotalOverloads = %d, expected 0", total)
	}

	fn1 := makeFunctionDecl("Func1", 0)
	fn2 := makeFunctionDecl("Func1", 1)
	fn3 := makeFunctionDecl("Func2", 0)

	registry.Register("Func1", fn1)
	registry.Register("Func1", fn2)
	registry.Register("Func2", fn3)

	if total := registry.TotalOverloads(); total != 3 {
		t.Errorf("TotalOverloads = %d, expected 3", total)
	}
}

func TestFunctionRegistry_Clear(t *testing.T) {
	registry := NewFunctionRegistry()

	fn := makeFunctionDecl("Test", 0)
	registry.Register("Test", fn)
	registry.RegisterWithUnit("Unit1", "Test2", fn)

	if count := registry.Count(); count == 0 {
		t.Error("Registry should not be empty before clear")
	}

	registry.Clear()

	if count := registry.Count(); count != 0 {
		t.Errorf("After clear count = %d, expected 0", count)
	}

	if exists := registry.Exists("Test"); exists {
		t.Error("Function should not exist after clear")
	}

	if exists := registry.ExistsQualified("Unit1", "Test2"); exists {
		t.Error("Qualified function should not exist after clear")
	}
}

func TestFunctionRegistry_RemoveFunction(t *testing.T) {
	registry := NewFunctionRegistry()

	fn1 := makeFunctionDecl("Test", 0)
	fn2 := makeFunctionDecl("Test", 1)

	registry.Register("Test", fn1)
	registry.Register("Test", fn2)

	// Remove should return true
	removed := registry.RemoveFunction("Test")
	if !removed {
		t.Error("RemoveFunction should return true when function exists")
	}

	// Function should no longer exist
	if exists := registry.Exists("Test"); exists {
		t.Error("Function should not exist after removal")
	}

	// Remove non-existent should return false
	removed = registry.RemoveFunction("NonExistent")
	if removed {
		t.Error("RemoveFunction should return false for non-existent function")
	}
}

func TestFunctionRegistry_FindFunctionsByParameterCount(t *testing.T) {
	registry := NewFunctionRegistry()

	fn1 := makeFunctionDecl("Func1", 0)
	fn2 := makeFunctionDecl("Func2", 1)
	fn3 := makeFunctionDecl("Func3", 1)
	fn4 := makeFunctionDecl("Func3", 2) // Overload with different param count

	registry.Register("Func1", fn1)
	registry.Register("Func2", fn2)
	registry.Register("Func3", fn3)
	registry.Register("Func3", fn4)

	// Find functions with 1 parameter
	funcs := registry.FindFunctionsByParameterCount(1)
	if len(funcs) != 2 {
		t.Errorf("Expected 2 functions with 1 parameter, got %d", len(funcs))
	}

	// Should include Func2 and Func3
	if _, ok := funcs["func2"]; !ok {
		t.Error("Should include Func2")
	}
	if _, ok := funcs["func3"]; !ok {
		t.Error("Should include Func3")
	}

	// Find functions with 0 parameters
	funcs = registry.FindFunctionsByParameterCount(0)
	if len(funcs) != 1 {
		t.Errorf("Expected 1 function with 0 parameters, got %d", len(funcs))
	}

	// Find functions with 3 parameters (none)
	funcs = registry.FindFunctionsByParameterCount(3)
	if len(funcs) != 0 {
		t.Errorf("Expected 0 functions with 3 parameters, got %d", len(funcs))
	}
}

func TestFunctionRegistry_GetFunctionMetadata(t *testing.T) {
	registry := NewFunctionRegistry()

	// Function with body
	fn1 := makeFunctionDecl("Test", 2)
	fn1.Body = &ast.BlockStatement{}

	// Forward declaration (no body)
	fn2 := makeFunctionDecl("Test", 3)
	fn2.Body = nil

	registry.RegisterWithUnit("MyUnit", "Test", fn1)
	registry.RegisterWithUnit("MyUnit", "Test", fn2)

	metadata := registry.GetFunctionMetadata("Test")
	if len(metadata) != 2 {
		t.Fatalf("Expected 2 metadata entries, got %d", len(metadata))
	}

	// Check first entry
	if metadata[0].Name != "Test" {
		t.Errorf("Expected name 'Test', got %q", metadata[0].Name)
	}
	if metadata[0].UnitName != "MyUnit" {
		t.Errorf("Expected unit 'MyUnit', got %q", metadata[0].UnitName)
	}
	if metadata[0].ParameterCount != 2 {
		t.Errorf("Expected 2 parameters, got %d", metadata[0].ParameterCount)
	}
	if metadata[0].IsForward {
		t.Error("First function should not be forward declaration")
	}

	// Check second entry
	if !metadata[1].IsForward {
		t.Error("Second function should be forward declaration")
	}
}

func TestFunctionRegistry_ValidateNoConflicts(t *testing.T) {
	registry := NewFunctionRegistry()

	// First function - no conflict
	err := registry.ValidateNoConflicts("Test", 1, false)
	if err != nil {
		t.Errorf("First function should have no conflict: %v", err)
	}

	// Register function without overload directive
	fn1 := makeFunctionDecl("Test", 1)
	fn1.IsOverload = false
	registry.Register("Test", fn1)

	// Try to register another with same param count, no overload directive
	err = registry.ValidateNoConflicts("Test", 1, false)
	if err == nil {
		t.Error("Expected conflict error for same param count without overload")
	}

	// With overload directive should be OK
	err = registry.ValidateNoConflicts("Test", 1, true)
	if err != nil {
		t.Errorf("With overload directive should be OK: %v", err)
	}

	// Different param count should be OK
	err = registry.ValidateNoConflicts("Test", 2, false)
	if err != nil {
		t.Errorf("Different param count should be OK: %v", err)
	}
}

func TestFunctionRegistry_GetFunctionsInUnit(t *testing.T) {
	registry := NewFunctionRegistry()

	fn1 := makeFunctionDecl("Func1", 0)
	fn2 := makeFunctionDecl("Func2", 1)
	fn3 := makeFunctionDecl("Func3", 0)

	registry.RegisterWithUnit("Math", "Func1", fn1)
	registry.RegisterWithUnit("Math", "Func2", fn2)
	registry.RegisterWithUnit("String", "Func3", fn3)

	// Get functions in Math unit
	mathFuncs := registry.GetFunctionsInUnit("Math")
	if len(mathFuncs) != 2 {
		t.Errorf("Expected 2 functions in Math unit, got %d", len(mathFuncs))
	}

	// Get functions in String unit
	stringFuncs := registry.GetFunctionsInUnit("String")
	if len(stringFuncs) != 1 {
		t.Errorf("Expected 1 function in String unit, got %d", len(stringFuncs))
	}

	// Get functions in non-existent unit
	otherFuncs := registry.GetFunctionsInUnit("Other")
	if len(otherFuncs) != 0 {
		t.Errorf("Expected 0 functions in Other unit, got %d", len(otherFuncs))
	}
}
