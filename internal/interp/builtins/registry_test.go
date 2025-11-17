package builtins

import (
	"math/rand"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// mockErrorValue is a simple error value for testing
type mockErrorValue struct {
	Message string
}

func (e *mockErrorValue) Type() string   { return "ERROR" }
func (e *mockErrorValue) String() string { return "ERROR: " + e.Message }

// mockContext implements the Context interface for testing
type mockContext struct {
	randSeed int64
	rng      *rand.Rand
}

func newMockContext() *mockContext {
	return &mockContext{
		randSeed: 0,
		rng:      rand.New(rand.NewSource(0)),
	}
}

func (m *mockContext) NewError(format string, args ...interface{}) Value {
	return &mockErrorValue{Message: "mock error"}
}

func (m *mockContext) CurrentNode() ast.Node {
	return nil
}

func (m *mockContext) RandSource() *rand.Rand {
	return m.rng
}

func (m *mockContext) GetRandSeed() int64 {
	return m.randSeed
}

func (m *mockContext) SetRandSeed(seed int64) {
	m.randSeed = seed
	m.rng = rand.New(rand.NewSource(seed))
}

func (m *mockContext) UnwrapVariant(value Value) Value {
	return value
}

func (m *mockContext) ToInt64(value Value) (int64, bool) {
	// Simple mock implementation for testing
	if iv, ok := value.(*runtime.IntegerValue); ok {
		return iv.Value, true
	}
	return 0, false
}

func (m *mockContext) ToBool(value Value) (bool, bool) {
	// Simple mock implementation for testing
	if bv, ok := value.(*runtime.BooleanValue); ok {
		return bv.Value, true
	}
	return false, false
}

func (m *mockContext) ToFloat64(value Value) (float64, bool) {
	// Simple mock implementation for testing
	if fv, ok := value.(*runtime.FloatValue); ok {
		return fv.Value, true
	}
	if iv, ok := value.(*runtime.IntegerValue); ok {
		return float64(iv.Value), true
	}
	return 0.0, false
}

// Task 3.7.6 Context methods
func (m *mockContext) ParseJSONString(jsonStr string) (Value, error) {
	// Simple mock - not needed for most tests
	return nil, nil
}

func (m *mockContext) ValueToJSON(value Value, formatted bool) (string, error) {
	// Simple mock - not needed for most tests
	return "{}", nil
}

func (m *mockContext) ValueToJSONWithIndent(value Value, formatted bool, indent int) (string, error) {
	// Simple mock - not needed for most tests
	return "{}", nil
}

func (m *mockContext) GetTypeOf(value Value) string {
	if value == nil {
		return "NULL"
	}
	return value.Type()
}

func (m *mockContext) GetClassOf(value Value) string {
	return ""
}

func (m *mockContext) JSONHasField(value Value, fieldName string) bool {
	return false
}

func (m *mockContext) JSONGetKeys(value Value) []string {
	return []string{}
}

func (m *mockContext) JSONGetValues(value Value) []Value {
	return []Value{}
}

func (m *mockContext) JSONGetLength(value Value) int {
	return 0
}

func (m *mockContext) CreateStringArray(values []string) Value {
	// Simple mock - return a dummy value for testing
	return &runtime.StringValue{Value: "mock array"}
}

func (m *mockContext) CreateVariantArray(values []Value) Value {
	// Simple mock - return a dummy value for testing
	return &runtime.StringValue{Value: "mock array"}
}

func (m *mockContext) Write(s string) {
	// Simple mock - no-op for testing
}

func (m *mockContext) WriteLine(s string) {
	// Simple mock - no-op for testing
}

func (m *mockContext) GetEnumOrdinal(value Value) (int64, bool) {
	// Simple mock - check type string since we can't import EnumValue
	if value.Type() == "ENUM" {
		// For testing, return a dummy ordinal
		return 0, true
	}
	return 0, false
}

func (m *mockContext) GetJSONVarType(value Value) (int64, bool) {
	// Simple mock - check type string since we can't import JSONValue
	if value.Type() == "JSON" {
		// For testing, return a dummy varType (varJSON = 0x1000)
		return 0x1000, true
	}
	return 0, false
}

// Task 3.7.7 Context methods for array and collection functions
func (m *mockContext) GetArrayLength(value Value) (int64, bool) {
	// Simple mock - check type string since we can't import ArrayValue
	if value.Type() == "ARRAY" {
		// For testing, return a dummy length
		return 0, true
	}
	return 0, false
}

func (m *mockContext) SetArrayLength(array Value, newLength int) error {
	// Simple mock - no-op for testing
	return nil
}

func (m *mockContext) ArrayCopy(array Value) Value {
	// Simple mock - return the same value for testing
	return array
}

func (m *mockContext) ArrayReverse(array Value) Value {
	// Simple mock - return nil for testing
	return &runtime.NilValue{}
}

func (m *mockContext) ArraySort(array Value) Value {
	// Simple mock - return nil for testing
	return &runtime.NilValue{}
}

func (m *mockContext) EvalFunctionPointer(funcPtr Value, args []Value) Value {
	// Simple mock - return nil for testing
	return &runtime.NilValue{}
}

// Task 3.7.8 Context methods for system functions
func (m *mockContext) GetCallStackString() string {
	// Simple mock - return empty string for testing
	return ""
}

func (m *mockContext) GetCallStackArray() Value {
	// Simple mock - return empty array for testing
	return &runtime.StringValue{Value: "[]"}
}

func (m *mockContext) IsAssigned(value Value) bool {
	// Simple mock - return true for non-nil values
	return value != nil
}

func (m *mockContext) RaiseAssertionFailed(customMessage string) {
	// Simple mock - no-op for testing
}

func (m *mockContext) GetEnumSuccessor(enumVal Value) (Value, error) {
	// Simple mock - return nil for testing
	return &runtime.NilValue{}, nil
}

func (m *mockContext) GetEnumPredecessor(enumVal Value) (Value, error) {
	// Simple mock - return nil for testing
	return &runtime.NilValue{}, nil
}

func (m *mockContext) ParseInt(s string, base int) (int64, bool) {
	// Simple mock - return 0 for testing
	return 0, false
}

func (m *mockContext) ParseFloat(s string) (float64, bool) {
	// Simple mock - return 0.0 for testing
	return 0.0, false
}

func (m *mockContext) FormatString(format string, args []Value) (string, error) {
	// Simple mock - return empty string for testing
	return "", nil
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}

	if r.Count() != 0 {
		t.Errorf("NewRegistry should create empty registry, got %d functions", r.Count())
	}
}

func TestRegister(t *testing.T) {
	r := NewRegistry()

	// Mock function
	mockFunc := func(ctx Context, args []Value) Value {
		return &runtime.IntegerValue{Value: 42}
	}

	r.Register("TestFunc", mockFunc, CategoryMath, "Test function")

	if !r.Has("TestFunc") {
		t.Error("Register did not add function to registry")
	}

	if r.Count() != 1 {
		t.Errorf("Expected 1 function, got %d", r.Count())
	}
}

func TestLookupCaseInsensitive(t *testing.T) {
	r := NewRegistry()

	mockFunc := func(ctx Context, args []Value) Value {
		return &runtime.IntegerValue{Value: 42}
	}

	r.Register("TestFunc", mockFunc, CategoryMath, "Test function")

	tests := []string{
		"TestFunc",
		"testfunc",
		"TESTFUNC",
		"TeStFuNc",
	}

	for _, name := range tests {
		fn, ok := r.Lookup(name)
		if !ok {
			t.Errorf("Lookup(%q) failed, expected to find function", name)
			continue
		}

		// Call the function to verify it works
		result := fn(newMockContext(), []Value{})
		if intVal, ok := result.(*runtime.IntegerValue); !ok || intVal.Value != 42 {
			t.Errorf("Lookup(%q) returned wrong function", name)
		}
	}
}

func TestGetByCategory(t *testing.T) {
	r := NewRegistry()

	mathFunc := func(ctx Context, args []Value) Value {
		return &runtime.IntegerValue{Value: 1}
	}

	stringFunc := func(ctx Context, args []Value) Value {
		return &runtime.StringValue{Value: "test"}
	}

	r.Register("Abs", mathFunc, CategoryMath, "Absolute value")
	r.Register("Sin", mathFunc, CategoryMath, "Sine function")
	r.Register("UpperCase", stringFunc, CategoryString, "Uppercase function")

	mathFuncs := r.GetByCategory(CategoryMath)
	if len(mathFuncs) != 2 {
		t.Errorf("Expected 2 math functions, got %d", len(mathFuncs))
	}

	stringFuncs := r.GetByCategory(CategoryString)
	if len(stringFuncs) != 1 {
		t.Errorf("Expected 1 string function, got %d", len(stringFuncs))
	}

	// Verify sorting
	if len(mathFuncs) == 2 {
		if mathFuncs[0].Name != "Abs" {
			t.Errorf("Expected first math function to be 'Abs', got %q", mathFuncs[0].Name)
		}
		if mathFuncs[1].Name != "Sin" {
			t.Errorf("Expected second math function to be 'Sin', got %q", mathFuncs[1].Name)
		}
	}
}

func TestAllCategories(t *testing.T) {
	r := NewRegistry()

	mockFunc := func(ctx Context, args []Value) Value {
		return &runtime.IntegerValue{Value: 0}
	}

	r.Register("Abs", mockFunc, CategoryMath, "Test")
	r.Register("UpperCase", mockFunc, CategoryString, "Test")
	r.Register("Now", mockFunc, CategoryDateTime, "Test")

	categories := r.AllCategories()
	if len(categories) != 3 {
		t.Errorf("Expected 3 categories, got %d", len(categories))
	}

	// Should be sorted
	expectedOrder := []Category{CategoryDateTime, CategoryMath, CategoryString}
	for i, expected := range expectedOrder {
		if i < len(categories) && categories[i] != expected {
			t.Errorf("Expected category %d to be %q, got %q", i, expected, categories[i])
		}
	}
}

func TestAllFunctions(t *testing.T) {
	r := NewRegistry()

	mockFunc := func(ctx Context, args []Value) Value {
		return &runtime.IntegerValue{Value: 0}
	}

	r.Register("Zebra", mockFunc, CategoryMath, "Test")
	r.Register("Alpha", mockFunc, CategoryString, "Test")
	r.Register("Beta", mockFunc, CategoryDateTime, "Test")

	funcs := r.AllFunctions()
	if len(funcs) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(funcs))
	}

	// Should be sorted alphabetically
	expectedOrder := []string{"Alpha", "Beta", "Zebra"}
	for i, expected := range expectedOrder {
		if i < len(funcs) && funcs[i].Name != expected {
			t.Errorf("Expected function %d to be %q, got %q", i, expected, funcs[i].Name)
		}
	}
}

func TestCategoryCount(t *testing.T) {
	r := NewRegistry()

	mockFunc := func(ctx Context, args []Value) Value {
		return &runtime.IntegerValue{Value: 0}
	}

	r.Register("Abs", mockFunc, CategoryMath, "Test")
	r.Register("Sin", mockFunc, CategoryMath, "Test")
	r.Register("Cos", mockFunc, CategoryMath, "Test")
	r.Register("UpperCase", mockFunc, CategoryString, "Test")

	if count := r.CategoryCount(CategoryMath); count != 3 {
		t.Errorf("Expected 3 math functions, got %d", count)
	}

	if count := r.CategoryCount(CategoryString); count != 1 {
		t.Errorf("Expected 1 string function, got %d", count)
	}

	if count := r.CategoryCount(CategoryDateTime); count != 0 {
		t.Errorf("Expected 0 datetime functions, got %d", count)
	}
}

func TestClear(t *testing.T) {
	r := NewRegistry()

	mockFunc := func(ctx Context, args []Value) Value {
		return &runtime.IntegerValue{Value: 0}
	}

	r.Register("Test1", mockFunc, CategoryMath, "Test")
	r.Register("Test2", mockFunc, CategoryString, "Test")

	if r.Count() != 2 {
		t.Errorf("Expected 2 functions before clear, got %d", r.Count())
	}

	r.Clear()

	if r.Count() != 0 {
		t.Errorf("Expected 0 functions after clear, got %d", r.Count())
	}

	if r.Has("Test1") {
		t.Error("Function should not exist after clear")
	}
}

func TestDefaultRegistry(t *testing.T) {
	if DefaultRegistry == nil {
		t.Fatal("DefaultRegistry is nil")
	}

	// Should have many functions registered
	if DefaultRegistry.Count() == 0 {
		t.Error("DefaultRegistry should have functions registered")
	}

	// Test some known functions
	knownFunctions := []string{
		"Abs", "Sin", "Cos", "Sqrt",
		"UpperCase", "LowerCase", "Trim",
		"Now", "FormatDateTime", "EncodeDate",
	}

	for _, name := range knownFunctions {
		if !DefaultRegistry.Has(name) {
			t.Errorf("DefaultRegistry should have %q function", name)
		}

		// Also test case-insensitive lookup
		if _, ok := DefaultRegistry.Lookup(name); !ok {
			t.Errorf("DefaultRegistry.Lookup(%q) failed", name)
		}
	}

	// Test category counts
	mathCount := DefaultRegistry.CategoryCount(CategoryMath)
	if mathCount == 0 {
		t.Error("DefaultRegistry should have math functions")
	}

	stringCount := DefaultRegistry.CategoryCount(CategoryString)
	if stringCount == 0 {
		t.Error("DefaultRegistry should have string functions")
	}

	datetimeCount := DefaultRegistry.CategoryCount(CategoryDateTime)
	if datetimeCount == 0 {
		t.Error("DefaultRegistry should have datetime functions")
	}
}

func TestGet(t *testing.T) {
	r := NewRegistry()

	mockFunc := func(ctx Context, args []Value) Value {
		return &runtime.IntegerValue{Value: 42}
	}

	r.Register("TestFunc", mockFunc, CategoryMath, "Test function description")

	info, ok := r.Get("TestFunc")
	if !ok {
		t.Fatal("Get failed to find function")
	}

	if info.Name != "TestFunc" {
		t.Errorf("Expected name 'TestFunc', got %q", info.Name)
	}

	if info.Category != CategoryMath {
		t.Errorf("Expected category %q, got %q", CategoryMath, info.Category)
	}

	if info.Description != "Test function description" {
		t.Errorf("Expected description 'Test function description', got %q", info.Description)
	}

	// Test case-insensitive
	info2, ok := r.Get("testfunc")
	if !ok {
		t.Error("Get should be case-insensitive")
	}

	if info2.Name != "TestFunc" {
		t.Error("Get should return original case name")
	}
}

func TestRegisterBatch(t *testing.T) {
	r := NewRegistry()

	mockFunc := func(ctx Context, args []Value) Value {
		return &runtime.IntegerValue{Value: 0}
	}

	entries := []struct {
		Name        string
		Function    BuiltinFunc
		Category    Category
		Description string
	}{
		{"Func1", mockFunc, CategoryMath, "Function 1"},
		{"Func2", mockFunc, CategoryString, "Function 2"},
		{"Func3", mockFunc, CategoryDateTime, "Function 3"},
	}

	r.RegisterBatch(entries)

	if r.Count() != 3 {
		t.Errorf("Expected 3 functions after batch register, got %d", r.Count())
	}

	for _, entry := range entries {
		if !r.Has(entry.Name) {
			t.Errorf("Batch register did not add %q", entry.Name)
		}
	}
}

func TestDuplicateRegistration(t *testing.T) {
	r := NewRegistry()

	mockFunc1 := func(ctx Context, args []Value) Value {
		return &runtime.IntegerValue{Value: 42}
	}

	mockFunc2 := func(ctx Context, args []Value) Value {
		return &runtime.IntegerValue{Value: 99}
	}

	// Register a function
	r.Register("TestFunc", mockFunc1, CategoryMath, "First registration")

	// Verify initial state
	if r.Count() != 1 {
		t.Errorf("Expected 1 function after first registration, got %d", r.Count())
	}

	if r.CategoryCount(CategoryMath) != 1 {
		t.Errorf("Expected 1 math function after first registration, got %d", r.CategoryCount(CategoryMath))
	}

	// Register the same function name again (should replace, not duplicate)
	r.Register("TestFunc", mockFunc2, CategoryMath, "Second registration")

	// Verify no duplicates were created
	if r.Count() != 1 {
		t.Errorf("Expected 1 function after duplicate registration, got %d", r.Count())
	}

	if r.CategoryCount(CategoryMath) != 1 {
		t.Errorf("Expected 1 math function after duplicate registration, got %d (duplicate category entry created)", r.CategoryCount(CategoryMath))
	}

	// Verify the function was replaced (should return 99, not 42)
	fn, ok := r.Lookup("TestFunc")
	if !ok {
		t.Fatal("Function not found after duplicate registration")
	}

	result := fn(newMockContext(), []Value{})
	if intVal, ok := result.(*runtime.IntegerValue); !ok || intVal.Value != 99 {
		t.Errorf("Expected function to be replaced with new implementation returning 99, got %v", result)
	}

	// Verify case-insensitive duplicate registration
	r.Register("testfunc", mockFunc1, CategoryMath, "Third registration (different case)")

	if r.Count() != 1 {
		t.Errorf("Expected 1 function after case-insensitive duplicate, got %d", r.Count())
	}

	if r.CategoryCount(CategoryMath) != 1 {
		t.Errorf("Expected 1 math function after case-insensitive duplicate, got %d", r.CategoryCount(CategoryMath))
	}

	// Verify GetByCategory doesn't return duplicates
	mathFuncs := r.GetByCategory(CategoryMath)
	if len(mathFuncs) != 1 {
		t.Errorf("GetByCategory should return 1 function, got %d (duplicates in category list)", len(mathFuncs))
	}
}

func TestConcurrency(t *testing.T) {
	r := NewRegistry()

	mockFunc := func(ctx Context, args []Value) Value {
		return &runtime.IntegerValue{Value: 0}
	}

	// Test concurrent registration and lookup
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			r.Register("ConcurrentFunc", mockFunc, CategoryMath, "Test")
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			r.Lookup("ConcurrentFunc")
			r.Has("ConcurrentFunc")
			r.Count()
		}
		done <- true
	}()

	<-done
	<-done

	if !r.Has("ConcurrentFunc") {
		t.Error("Concurrent operations failed")
	}
}
