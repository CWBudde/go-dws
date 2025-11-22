package runtime

import (
	"sync"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestNewMethodRegistry tests registry creation.
func TestNewMethodRegistry(t *testing.T) {
	registry := NewMethodRegistry()

	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}

	if registry.nextID != 1 {
		t.Errorf("Expected nextID to be 1, got %d", registry.nextID)
	}

	if registry.methods == nil {
		t.Error("Methods map not initialized")
	}

	if registry.nameIndex == nil {
		t.Error("Name index not initialized")
	}

	if registry.Count() != 0 {
		t.Errorf("Expected empty registry, got %d methods", registry.Count())
	}
}

// TestRegisterMethod tests method registration.
func TestRegisterMethod(t *testing.T) {
	registry := NewMethodRegistry()

	method := &MethodMetadata{
		Name:           "DoSomething",
		ReturnTypeName: "Integer",
		Parameters:     []ParameterMetadata{},
	}

	id := registry.RegisterMethod(method)

	if id == InvalidMethodID {
		t.Error("Expected valid method ID")
	}

	if id != 1 {
		t.Errorf("Expected first ID to be 1, got %d", id)
	}

	if method.ID != id {
		t.Errorf("Expected method ID to be set to %d, got %d", id, method.ID)
	}

	if registry.Count() != 1 {
		t.Errorf("Expected 1 method in registry, got %d", registry.Count())
	}
}

// TestRegisterMethod_Nil tests that registering nil returns InvalidMethodID.
func TestRegisterMethod_Nil(t *testing.T) {
	registry := NewMethodRegistry()

	id := registry.RegisterMethod(nil)

	if id != InvalidMethodID {
		t.Errorf("Expected InvalidMethodID for nil, got %d", id)
	}

	if registry.Count() != 0 {
		t.Errorf("Expected 0 methods in registry, got %d", registry.Count())
	}
}

// TestRegisterMethod_Multiple tests registering multiple methods.
func TestRegisterMethod_Multiple(t *testing.T) {
	registry := NewMethodRegistry()

	method1 := &MethodMetadata{Name: "Method1"}
	method2 := &MethodMetadata{Name: "Method2"}
	method3 := &MethodMetadata{Name: "Method3"}

	id1 := registry.RegisterMethod(method1)
	id2 := registry.RegisterMethod(method2)
	id3 := registry.RegisterMethod(method3)

	if id1 == InvalidMethodID || id2 == InvalidMethodID || id3 == InvalidMethodID {
		t.Error("Expected all IDs to be valid")
	}

	if id1 == id2 || id2 == id3 || id1 == id3 {
		t.Error("Expected unique IDs for each method")
	}

	if registry.Count() != 3 {
		t.Errorf("Expected 3 methods in registry, got %d", registry.Count())
	}

	// IDs should be sequential
	if id1 != 1 || id2 != 2 || id3 != 3 {
		t.Errorf("Expected sequential IDs 1,2,3, got %d,%d,%d", id1, id2, id3)
	}
}

// TestGetMethod tests method retrieval by ID.
func TestGetMethod(t *testing.T) {
	registry := NewMethodRegistry()

	original := &MethodMetadata{
		Name:           "MyMethod",
		ReturnTypeName: "String",
	}

	id := registry.RegisterMethod(original)

	retrieved := registry.GetMethod(id)

	if retrieved == nil {
		t.Fatal("Expected non-nil method")
	}

	if retrieved != original {
		t.Error("Expected same method instance")
	}

	if retrieved.Name != "MyMethod" {
		t.Errorf("Expected name 'MyMethod', got %q", retrieved.Name)
	}
}

// TestGetMethod_Invalid tests retrieving with invalid ID.
func TestGetMethod_Invalid(t *testing.T) {
	registry := NewMethodRegistry()

	// Try invalid IDs
	if method := registry.GetMethod(InvalidMethodID); method != nil {
		t.Error("Expected nil for InvalidMethodID")
	}

	if method := registry.GetMethod(999); method != nil {
		t.Error("Expected nil for non-existent ID")
	}
}

// TestGetMethodsByName tests name-based method lookup.
func TestGetMethodsByName(t *testing.T) {
	registry := NewMethodRegistry()

	method1 := &MethodMetadata{
		Name:       "Calculate",
		Parameters: []ParameterMetadata{{Name: "x"}},
	}
	method2 := &MethodMetadata{
		Name:       "Calculate",
		Parameters: []ParameterMetadata{{Name: "x"}, {Name: "y"}},
	}
	method3 := &MethodMetadata{
		Name: "Print",
	}

	registry.RegisterMethod(method1)
	registry.RegisterMethod(method2)
	registry.RegisterMethod(method3)

	// Find overloads of Calculate
	calcMethods := registry.GetMethodsByName("Calculate")
	if len(calcMethods) != 2 {
		t.Errorf("Expected 2 Calculate methods, got %d", len(calcMethods))
	}

	// Find Print
	printMethods := registry.GetMethodsByName("Print")
	if len(printMethods) != 1 {
		t.Errorf("Expected 1 Print method, got %d", len(printMethods))
	}

	// Non-existent method
	nonExistent := registry.GetMethodsByName("NonExistent")
	if len(nonExistent) != 0 {
		t.Errorf("Expected 0 methods for non-existent name, got %d", len(nonExistent))
	}
}

// TestGetMethodsByName_CaseInsensitive tests case-insensitive name lookup.
func TestGetMethodsByName_CaseInsensitive(t *testing.T) {
	registry := NewMethodRegistry()

	method := &MethodMetadata{Name: "MyMethod"}
	registry.RegisterMethod(method)

	tests := []string{
		"MyMethod",
		"mymethod",
		"MYMETHOD",
		"MYmEtHoD",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			methods := registry.GetMethodsByName(name)
			if len(methods) != 1 {
				t.Errorf("Expected 1 method for %q, got %d", name, len(methods))
			}
			if len(methods) > 0 && methods[0] != method {
				t.Error("Expected to find the same method")
			}
		})
	}
}

// TestHasMethod tests method existence checking.
func TestHasMethod(t *testing.T) {
	registry := NewMethodRegistry()

	method := &MethodMetadata{Name: "TestMethod"}
	id := registry.RegisterMethod(method)

	if !registry.HasMethod(id) {
		t.Errorf("Expected HasMethod to return true for valid ID %d", id)
	}

	if registry.HasMethod(InvalidMethodID) {
		t.Error("Expected HasMethod to return false for InvalidMethodID")
	}

	if registry.HasMethod(999) {
		t.Error("Expected HasMethod to return false for non-existent ID")
	}
}

// TestCount tests method counting.
func TestCount(t *testing.T) {
	registry := NewMethodRegistry()

	if registry.Count() != 0 {
		t.Errorf("Expected 0 methods initially, got %d", registry.Count())
	}

	registry.RegisterMethod(&MethodMetadata{Name: "M1"})
	if registry.Count() != 1 {
		t.Errorf("Expected 1 method, got %d", registry.Count())
	}

	registry.RegisterMethod(&MethodMetadata{Name: "M2"})
	if registry.Count() != 2 {
		t.Errorf("Expected 2 methods, got %d", registry.Count())
	}

	registry.RegisterMethod(&MethodMetadata{Name: "M3"})
	if registry.Count() != 3 {
		t.Errorf("Expected 3 methods, got %d", registry.Count())
	}
}

// TestClear tests clearing the registry.
func TestClear(t *testing.T) {
	registry := NewMethodRegistry()

	// Add some methods
	registry.RegisterMethod(&MethodMetadata{Name: "M1"})
	registry.RegisterMethod(&MethodMetadata{Name: "M2"})
	registry.RegisterMethod(&MethodMetadata{Name: "M3"})

	if registry.Count() != 3 {
		t.Errorf("Expected 3 methods before clear, got %d", registry.Count())
	}

	// Clear
	registry.Clear()

	if registry.Count() != 0 {
		t.Errorf("Expected 0 methods after clear, got %d", registry.Count())
	}

	// Next ID should reset to 1
	newMethod := &MethodMetadata{Name: "New"}
	newID := registry.RegisterMethod(newMethod)
	if newID != 1 {
		t.Errorf("Expected ID to reset to 1 after clear, got %d", newID)
	}
}

// TestStats tests registry statistics.
func TestStats(t *testing.T) {
	registry := NewMethodRegistry()

	// Initial stats
	stats := registry.Stats()
	if stats.MethodCount != 0 {
		t.Errorf("Expected 0 methods, got %d", stats.MethodCount)
	}
	if stats.NextID != 1 {
		t.Errorf("Expected NextID 1, got %d", stats.NextID)
	}
	if stats.UniqueNameCount != 0 {
		t.Errorf("Expected 0 unique names, got %d", stats.UniqueNameCount)
	}

	// Add methods with overloads
	registry.RegisterMethod(&MethodMetadata{
		Name:       "Calculate",
		Parameters: []ParameterMetadata{{Name: "x"}},
	})
	registry.RegisterMethod(&MethodMetadata{
		Name:       "Calculate",
		Parameters: []ParameterMetadata{{Name: "x"}, {Name: "y"}},
	})
	registry.RegisterMethod(&MethodMetadata{
		Name:       "Calculate",
		Parameters: []ParameterMetadata{{Name: "x"}, {Name: "y"}, {Name: "z"}},
	})
	registry.RegisterMethod(&MethodMetadata{Name: "Print"})

	stats = registry.Stats()
	if stats.MethodCount != 4 {
		t.Errorf("Expected 4 methods, got %d", stats.MethodCount)
	}
	if stats.UniqueNameCount != 2 {
		t.Errorf("Expected 2 unique names, got %d", stats.UniqueNameCount)
	}
	if stats.MaxOverloadCount != 3 {
		t.Errorf("Expected max 3 overloads, got %d", stats.MaxOverloadCount)
	}
	if stats.NextID != 5 {
		t.Errorf("Expected NextID 5, got %d", stats.NextID)
	}
}

// TestStats_String tests the string representation of stats.
func TestStats_String(t *testing.T) {
	registry := NewMethodRegistry()

	registry.RegisterMethod(&MethodMetadata{Name: "M1"})
	registry.RegisterMethod(&MethodMetadata{Name: "M2"})

	stats := registry.Stats()
	str := stats.String()

	if str == "" {
		t.Error("Expected non-empty string")
	}

	// String should contain key info
	if !contains(str, "2 methods") {
		t.Errorf("Expected string to contain method count, got: %s", str)
	}
}

// TestMethodRegistry_Concurrent tests concurrent access to the registry.
func TestMethodRegistry_Concurrent(t *testing.T) {
	registry := NewMethodRegistry()

	const numGoroutines = 10
	const methodsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(n int) {
			defer wg.Done()
			for j := 0; j < methodsPerGoroutine; j++ {
				method := &MethodMetadata{
					Name: "Method",
				}
				id := registry.RegisterMethod(method)
				if id == InvalidMethodID {
					t.Errorf("Got InvalidMethodID during concurrent registration")
				}
			}
		}(i)
	}

	wg.Wait()

	expectedCount := numGoroutines * methodsPerGoroutine
	if registry.Count() != expectedCount {
		t.Errorf("Expected %d methods after concurrent writes, got %d",
			expectedCount, registry.Count())
	}

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 1; j <= expectedCount; j++ {
				method := registry.GetMethod(MethodID(j))
				if method == nil {
					t.Errorf("Failed to get method %d during concurrent reads", j)
				}
			}
		}()
	}

	wg.Wait()
}

// TestMethodRegistry_WithRealMetadata tests the registry with realistic method metadata.
func TestMethodRegistry_WithRealMetadata(t *testing.T) {
	registry := NewMethodRegistry()

	// Function with return type
	function := &MethodMetadata{
		Name: "Add",
		Parameters: []ParameterMetadata{
			{Name: "a", TypeName: "Integer", ByRef: false},
			{Name: "b", TypeName: "Integer", ByRef: false},
		},
		ReturnTypeName: "Integer",
		Body:           &ast.BlockStatement{},
	}

	// Procedure without return type
	procedure := &MethodMetadata{
		Name: "PrintMessage",
		Parameters: []ParameterMetadata{
			{Name: "msg", TypeName: "String", ByRef: false},
		},
		ReturnTypeName: "",
		Body:           &ast.BlockStatement{},
	}

	// Virtual method
	virtualMethod := &MethodMetadata{
		Name:           "DoSomething",
		ReturnTypeName: "",
		IsVirtual:      true,
		Visibility:     VisibilityPublic,
	}

	// Register methods
	id1 := registry.RegisterMethod(function)
	id2 := registry.RegisterMethod(procedure)
	id3 := registry.RegisterMethod(virtualMethod)

	// Verify registration
	if id1 == InvalidMethodID || id2 == InvalidMethodID || id3 == InvalidMethodID {
		t.Fatal("Failed to register methods")
	}

	// Verify retrieval
	retrieved1 := registry.GetMethod(id1)
	if !retrieved1.IsFunction() {
		t.Error("Expected Add to be a function")
	}
	if retrieved1.ParamCount() != 2 {
		t.Errorf("Expected Add to have 2 parameters, got %d", retrieved1.ParamCount())
	}

	retrieved2 := registry.GetMethod(id2)
	if !retrieved2.IsProcedure() {
		t.Error("Expected PrintMessage to be a procedure")
	}

	retrieved3 := registry.GetMethod(id3)
	if !retrieved3.IsVirtual {
		t.Error("Expected DoSomething to be virtual")
	}
}

// TestInvalidMethodID tests the invalid method ID constant.
func TestInvalidMethodID(t *testing.T) {
	if InvalidMethodID != 0 {
		t.Errorf("Expected InvalidMethodID to be 0, got %d", InvalidMethodID)
	}

	registry := NewMethodRegistry()

	// Invalid ID should not be returned by registration
	method := &MethodMetadata{Name: "Test"}
	id := registry.RegisterMethod(method)
	if id == InvalidMethodID {
		t.Error("RegisterMethod should not return InvalidMethodID for valid metadata")
	}
}

// Helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexString(s, substr) >= 0)
}

// Helper function to find substring index.
func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
