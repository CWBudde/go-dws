package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Test Helpers
// ============================================================================

// makePosition creates a test position
func makePosition(line, col int) token.Position {
	return token.Position{Line: line, Column: col, Offset: 0}
}

// ============================================================================
// Basic Registration and Lookup Tests
// ============================================================================

func TestTypeRegistry_BasicRegistration(t *testing.T) {
	registry := NewTypeRegistry()

	// Register a simple type
	intType := &types.IntegerType{}
	err := registry.Register("MyInt", intType, makePosition(1, 1), 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify we can resolve it
	resolved, ok := registry.Resolve("MyInt")
	if !ok {
		t.Fatal("Expected to find registered type")
	}
	if !resolved.Equals(intType) {
		t.Errorf("Expected %v, got %v", intType, resolved)
	}
}

func TestTypeRegistry_CaseInsensitiveLookup(t *testing.T) {
	registry := NewTypeRegistry()

	intType := &types.IntegerType{}
	err := registry.Register("MyType", intType, makePosition(1, 1), 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test various case combinations
	testCases := []string{"MyType", "mytype", "MYTYPE", "mYtYpE"}
	for _, name := range testCases {
		resolved, ok := registry.Resolve(name)
		if !ok {
			t.Errorf("Expected to find type with name '%s'", name)
		}
		if !resolved.Equals(intType) {
			t.Errorf("Expected %v, got %v for name '%s'", intType, resolved, name)
		}
	}
}

func TestTypeRegistry_DuplicateDetection(t *testing.T) {
	registry := NewTypeRegistry()

	// Register first type
	intType := &types.IntegerType{}
	err := registry.Register("MyType", intType, makePosition(1, 1), 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Try to register duplicate (case-insensitive)
	strType := &types.StringType{}
	err = registry.Register("mytype", strType, makePosition(2, 1), 0)
	if err == nil {
		t.Fatal("Expected error for duplicate type, got nil")
	}
}

func TestTypeRegistry_EmptyName(t *testing.T) {
	registry := NewTypeRegistry()

	err := registry.Register("", &types.IntegerType{}, makePosition(1, 1), 0)
	if err == nil {
		t.Fatal("Expected error for empty name, got nil")
	}
}

func TestTypeRegistry_NilType(t *testing.T) {
	registry := NewTypeRegistry()

	err := registry.Register("MyType", nil, makePosition(1, 1), 0)
	if err == nil {
		t.Fatal("Expected error for nil type, got nil")
	}
}

func TestTypeRegistry_ResolveNonExistent(t *testing.T) {
	registry := NewTypeRegistry()

	resolved, ok := registry.Resolve("NonExistent")
	if ok {
		t.Fatal("Expected not to find non-existent type")
	}
	if resolved != nil {
		t.Errorf("Expected nil type, got %v", resolved)
	}
}

// ============================================================================
// Descriptor Tests
// ============================================================================

func TestTypeRegistry_ResolveDescriptor(t *testing.T) {
	registry := NewTypeRegistry()

	intType := &types.IntegerType{}
	pos := makePosition(10, 5)
	err := registry.Register("MyInt", intType, pos, 2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	descriptor, ok := registry.ResolveDescriptor("MyInt")
	if !ok {
		t.Fatal("Expected to find descriptor")
	}

	if descriptor.Name != "MyInt" {
		t.Errorf("Expected name 'MyInt', got '%s'", descriptor.Name)
	}
	if !descriptor.Type.Equals(intType) {
		t.Errorf("Expected type %v, got %v", intType, descriptor.Type)
	}
	if descriptor.Position.Line != 10 || descriptor.Position.Column != 5 {
		t.Errorf("Expected position 10:5, got %d:%d", descriptor.Position.Line, descriptor.Position.Column)
	}
	if descriptor.Visibility != 2 {
		t.Errorf("Expected visibility 2, got %d", descriptor.Visibility)
	}
}

func TestTypeRegistry_MustResolve(t *testing.T) {
	registry := NewTypeRegistry()

	intType := &types.IntegerType{}
	err := registry.Register("MyInt", intType, makePosition(1, 1), 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should not panic
	resolved := registry.MustResolve("MyInt")
	if !resolved.Equals(intType) {
		t.Errorf("Expected %v, got %v", intType, resolved)
	}
}

func TestTypeRegistry_MustResolvePanic(t *testing.T) {
	registry := NewTypeRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for non-existent type")
		}
	}()

	registry.MustResolve("NonExistent")
}

// ============================================================================
// Query and Iteration Tests
// ============================================================================

func TestTypeRegistry_AllTypes(t *testing.T) {
	registry := NewTypeRegistry()

	// Register multiple types
	registry.Register("Int1", &types.IntegerType{}, makePosition(1, 1), 0)
	registry.Register("Str1", &types.StringType{}, makePosition(2, 1), 0)
	registry.Register("Bool1", &types.BooleanType{}, makePosition(3, 1), 0)

	allTypes := registry.AllTypes()
	if len(allTypes) != 3 {
		t.Errorf("Expected 3 types, got %d", len(allTypes))
	}

	// Check that all types are present (case-preserved names)
	if _, ok := allTypes["Int1"]; !ok {
		t.Error("Expected to find 'Int1'")
	}
	if _, ok := allTypes["Str1"]; !ok {
		t.Error("Expected to find 'Str1'")
	}
	if _, ok := allTypes["Bool1"]; !ok {
		t.Error("Expected to find 'Bool1'")
	}
}

func TestTypeRegistry_AllDescriptors(t *testing.T) {
	registry := NewTypeRegistry()

	registry.Register("Int1", &types.IntegerType{}, makePosition(1, 1), 0)
	registry.Register("Str1", &types.StringType{}, makePosition(2, 1), 1)

	allDescriptors := registry.AllDescriptors()
	if len(allDescriptors) != 2 {
		t.Errorf("Expected 2 descriptors, got %d", len(allDescriptors))
	}

	// Check descriptor details
	if desc, ok := allDescriptors["Int1"]; ok {
		if desc.Visibility != 0 {
			t.Errorf("Expected visibility 0 for Int1, got %d", desc.Visibility)
		}
	} else {
		t.Error("Expected to find 'Int1' descriptor")
	}
}

func TestTypeRegistry_TypesByKind(t *testing.T) {
	registry := NewTypeRegistry()

	// Register types of different kinds
	registry.Register("Int1", &types.IntegerType{}, makePosition(1, 1), 0)
	registry.Register("Int2", &types.IntegerType{}, makePosition(2, 1), 0)
	registry.Register("Str1", &types.StringType{}, makePosition(3, 1), 0)

	class := &types.ClassType{Name: "MyClass"}
	registry.Register("MyClass", class, makePosition(4, 1), 0)

	enum := &types.EnumType{Name: "MyEnum"}
	registry.Register("MyEnum", enum, makePosition(5, 1), 0)

	// Query by kind
	integers := registry.TypesByKind("INTEGER")
	if len(integers) != 2 {
		t.Errorf("Expected 2 INTEGER types, got %d", len(integers))
	}

	strings := registry.TypesByKind("STRING")
	if len(strings) != 1 {
		t.Errorf("Expected 1 STRING type, got %d", len(strings))
	}

	classes := registry.TypesByKind("CLASS")
	if len(classes) != 1 {
		t.Errorf("Expected 1 CLASS type, got %d", len(classes))
	}

	enums := registry.TypesByKind("ENUM")
	if len(enums) != 1 {
		t.Errorf("Expected 1 ENUM type, got %d", len(enums))
	}

	// Query non-existent kind
	nonExistent := registry.TypesByKind("NONEXISTENT")
	if len(nonExistent) != 0 {
		t.Errorf("Expected 0 NONEXISTENT types, got %d", len(nonExistent))
	}
}

func TestTypeRegistry_Count(t *testing.T) {
	registry := NewTypeRegistry()

	if registry.Count() != 0 {
		t.Errorf("Expected 0 types in new registry, got %d", registry.Count())
	}

	registry.Register("Type1", &types.IntegerType{}, makePosition(1, 1), 0)
	if registry.Count() != 1 {
		t.Errorf("Expected 1 type, got %d", registry.Count())
	}

	registry.Register("Type2", &types.StringType{}, makePosition(2, 1), 0)
	if registry.Count() != 2 {
		t.Errorf("Expected 2 types, got %d", registry.Count())
	}
}

// ============================================================================
// Position-based Query Tests (LSP Support)
// ============================================================================

func TestTypeRegistry_FindTypeByPosition(t *testing.T) {
	registry := NewTypeRegistry()

	registry.Register("Type1", &types.IntegerType{}, makePosition(10, 5), 0)
	registry.Register("Type2", &types.StringType{}, makePosition(20, 10), 0)
	registry.Register("Type3", &types.BooleanType{}, makePosition(30, 15), 0)

	// Find exact position
	descriptor, ok := registry.FindTypeByPosition(makePosition(20, 10))
	if !ok {
		t.Fatal("Expected to find type at position 20:10")
	}
	if descriptor.Name != "Type2" {
		t.Errorf("Expected Type2, got %s", descriptor.Name)
	}

	// Non-existent position
	descriptor, ok = registry.FindTypeByPosition(makePosition(99, 99))
	if ok {
		t.Fatal("Expected not to find type at position 99:99")
	}
	if descriptor != nil {
		t.Errorf("Expected nil descriptor, got %v", descriptor)
	}
}

func TestTypeRegistry_TypesInRange(t *testing.T) {
	registry := NewTypeRegistry()

	registry.Register("Type1", &types.IntegerType{}, makePosition(5, 1), 0)
	registry.Register("Type2", &types.StringType{}, makePosition(10, 1), 0)
	registry.Register("Type3", &types.BooleanType{}, makePosition(15, 1), 0)
	registry.Register("Type4", &types.FloatType{}, makePosition(20, 1), 0)

	// Query range 8-17 (should get Type2 and Type3)
	inRange := registry.TypesInRange(8, 17)
	if len(inRange) != 2 {
		t.Errorf("Expected 2 types in range, got %d", len(inRange))
	}

	// Verify we got the right types
	names := make(map[string]bool)
	for _, desc := range inRange {
		names[desc.Name] = true
	}
	if !names["Type2"] || !names["Type3"] {
		t.Error("Expected Type2 and Type3 in range")
	}

	// Query range with no types
	emptyRange := registry.TypesInRange(100, 200)
	if len(emptyRange) != 0 {
		t.Errorf("Expected 0 types in range, got %d", len(emptyRange))
	}
}

// ============================================================================
// Dependency Analysis Tests
// ============================================================================

func TestTypeRegistry_GetTypeDependencies_Record(t *testing.T) {
	registry := NewTypeRegistry()

	// Create a record with fields (Fields is map[string]Type)
	record := &types.RecordType{
		Name: "MyRecord",
		Fields: map[string]types.Type{
			"X": &types.IntegerType{},
			"Y": &types.StringType{},
		},
	}
	registry.Register("MyRecord", record, makePosition(1, 1), 0)

	deps := registry.GetTypeDependencies("MyRecord")
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(deps))
	}

	// Check that we have Integer and String
	depsMap := make(map[string]bool)
	for _, dep := range deps {
		depsMap[dep] = true
	}
	if !depsMap["Integer"] || !depsMap["String"] {
		t.Error("Expected Integer and String dependencies")
	}
}

func TestTypeRegistry_GetTypeDependencies_Class(t *testing.T) {
	registry := NewTypeRegistry()

	// Create a class with parent and fields (Fields is map[string]Type)
	parent := &types.ClassType{Name: "TObject"}
	registry.Register("TObject", parent, makePosition(1, 1), 0)

	child := &types.ClassType{
		Name:   "TMyClass",
		Parent: parent,
		Fields: map[string]types.Type{
			"Value": &types.IntegerType{},
		},
	}
	registry.Register("TMyClass", child, makePosition(2, 1), 0)

	deps := registry.GetTypeDependencies("TMyClass")
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(deps))
	}

	// Check that we have TObject and Integer
	depsMap := make(map[string]bool)
	for _, dep := range deps {
		depsMap[dep] = true
	}
	if !depsMap["TObject"] || !depsMap["Integer"] {
		t.Errorf("Expected TObject and Integer dependencies, got %v", deps)
	}
}

func TestTypeRegistry_GetTypeDependencies_Array(t *testing.T) {
	registry := NewTypeRegistry()

	array := &types.ArrayType{
		ElementType: &types.StringType{},
	}
	registry.Register("MyArray", array, makePosition(1, 1), 0)

	deps := registry.GetTypeDependencies("MyArray")
	if len(deps) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(deps))
	}
	if deps[0] != "String" {
		t.Errorf("Expected String dependency, got %s", deps[0])
	}
}

func TestTypeRegistry_GetTypeDependencies_TypeAlias(t *testing.T) {
	registry := NewTypeRegistry()

	alias := &types.TypeAlias{
		Name:        "MyInt",
		AliasedType: &types.IntegerType{},
	}
	registry.Register("MyInt", alias, makePosition(1, 1), 0)

	deps := registry.GetTypeDependencies("MyInt")
	if len(deps) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(deps))
	}
	if deps[0] != "Integer" {
		t.Errorf("Expected Integer dependency, got %s", deps[0])
	}
}

func TestTypeRegistry_GetTypeDependencies_NonExistent(t *testing.T) {
	registry := NewTypeRegistry()

	deps := registry.GetTypeDependencies("NonExistent")
	if deps != nil {
		t.Errorf("Expected nil for non-existent type, got %v", deps)
	}
}

// ============================================================================
// Alias Resolution Tests
// ============================================================================

func TestTypeRegistry_ResolveUnderlying_SimpleAlias(t *testing.T) {
	registry := NewTypeRegistry()

	// Register base type and alias
	registry.Register("Integer", &types.IntegerType{}, makePosition(1, 1), 2)
	alias := &types.TypeAlias{
		Name:        "MyInt",
		AliasedType: &types.IntegerType{},
	}
	registry.Register("MyInt", alias, makePosition(2, 1), 0)

	// Resolve the alias to underlying type
	underlying, ok := registry.ResolveUnderlying("MyInt")
	if !ok {
		t.Fatal("Expected to resolve MyInt")
	}

	// Should get IntegerType, not TypeAlias
	if _, isAlias := underlying.(*types.TypeAlias); isAlias {
		t.Error("Expected underlying type to not be an alias")
	}
	if _, isInt := underlying.(*types.IntegerType); !isInt {
		t.Errorf("Expected IntegerType, got %T", underlying)
	}
}

func TestTypeRegistry_ResolveUnderlying_ChainedAliases(t *testing.T) {
	registry := NewTypeRegistry()

	// Create a chain of aliases: C -> B -> A -> Integer
	intType := &types.IntegerType{}
	aliasA := &types.TypeAlias{Name: "A", AliasedType: intType}
	aliasB := &types.TypeAlias{Name: "B", AliasedType: aliasA}
	aliasC := &types.TypeAlias{Name: "C", AliasedType: aliasB}

	registry.Register("Integer", intType, makePosition(1, 1), 2)
	registry.Register("A", aliasA, makePosition(2, 1), 0)
	registry.Register("B", aliasB, makePosition(3, 1), 0)
	registry.Register("C", aliasC, makePosition(4, 1), 0)

	// Resolve through the chain
	underlying, ok := registry.ResolveUnderlying("C")
	if !ok {
		t.Fatal("Expected to resolve C")
	}

	// Should get IntegerType at the end of the chain
	if _, isInt := underlying.(*types.IntegerType); !isInt {
		t.Errorf("Expected IntegerType, got %T", underlying)
	}
}

func TestTypeRegistry_ResolveUnderlying_NonAlias(t *testing.T) {
	registry := NewTypeRegistry()

	// Register a non-alias type
	intType := &types.IntegerType{}
	registry.Register("Integer", intType, makePosition(1, 1), 2)

	// Resolve should just return the type itself
	underlying, ok := registry.ResolveUnderlying("Integer")
	if !ok {
		t.Fatal("Expected to resolve Integer")
	}

	if !underlying.Equals(intType) {
		t.Errorf("Expected same type, got %v", underlying)
	}
}

func TestTypeRegistry_ResolveUnderlying_NonExistent(t *testing.T) {
	registry := NewTypeRegistry()

	underlying, ok := registry.ResolveUnderlying("NonExistent")
	if ok {
		t.Fatal("Expected not to resolve non-existent type")
	}
	if underlying != nil {
		t.Errorf("Expected nil type, got %v", underlying)
	}
}

// ============================================================================
// Helper Method Tests
// ============================================================================

func TestTypeRegistry_Has(t *testing.T) {
	registry := NewTypeRegistry()

	registry.Register("MyType", &types.IntegerType{}, makePosition(1, 1), 0)

	// Check existing type
	if !registry.Has("MyType") {
		t.Error("Expected Has to return true for MyType")
	}

	// Check case-insensitive
	if !registry.Has("mytype") {
		t.Error("Expected Has to be case-insensitive")
	}

	// Check non-existent type
	if registry.Has("NonExistent") {
		t.Error("Expected Has to return false for non-existent type")
	}
}

func TestTypeRegistry_RegisterBuiltIn(t *testing.T) {
	registry := NewTypeRegistry()

	// Register a built-in type
	err := registry.RegisterBuiltIn("Integer", &types.IntegerType{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify it was registered
	typ, ok := registry.Resolve("Integer")
	if !ok {
		t.Fatal("Expected to find Integer")
	}
	if _, isInt := typ.(*types.IntegerType); !isInt {
		t.Errorf("Expected IntegerType, got %T", typ)
	}

	// Verify descriptor has correct properties
	desc, ok := registry.ResolveDescriptor("Integer")
	if !ok {
		t.Fatal("Expected to find Integer descriptor")
	}
	if desc.Position.Line != 0 || desc.Position.Column != 0 {
		t.Errorf("Expected position 0:0 for built-in, got %d:%d", desc.Position.Line, desc.Position.Column)
	}
	if desc.Visibility != 2 {
		t.Errorf("Expected visibility 2 (public) for built-in, got %d", desc.Visibility)
	}
}

func TestTypeRegistry_MustRegisterBuiltIn(t *testing.T) {
	registry := NewTypeRegistry()

	// Should not panic for valid registration
	registry.MustRegisterBuiltIn("Integer", &types.IntegerType{})

	// Verify it was registered
	if !registry.Has("Integer") {
		t.Error("Expected Integer to be registered")
	}
}

func TestTypeRegistry_MustRegisterBuiltIn_Panic(t *testing.T) {
	registry := NewTypeRegistry()

	// Register first type
	registry.MustRegisterBuiltIn("Integer", &types.IntegerType{})

	// Try to register duplicate - should panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for duplicate built-in type")
		}
	}()

	registry.MustRegisterBuiltIn("Integer", &types.StringType{})
}

// ============================================================================
// Utility Method Tests
// ============================================================================

func TestTypeRegistry_Clear(t *testing.T) {
	registry := NewTypeRegistry()

	// Register some types
	registry.Register("Type1", &types.IntegerType{}, makePosition(1, 1), 0)
	registry.Register("Type2", &types.StringType{}, makePosition(2, 1), 0)

	if registry.Count() != 2 {
		t.Errorf("Expected 2 types before clear, got %d", registry.Count())
	}

	// Clear the registry
	registry.Clear()

	if registry.Count() != 0 {
		t.Errorf("Expected 0 types after clear, got %d", registry.Count())
	}

	// Verify we can't resolve cleared types
	_, ok := registry.Resolve("Type1")
	if ok {
		t.Error("Expected not to find Type1 after clear")
	}
}

func TestTypeRegistry_Unregister(t *testing.T) {
	registry := NewTypeRegistry()

	registry.Register("Type1", &types.IntegerType{}, makePosition(1, 1), 0)
	registry.Register("Type2", &types.StringType{}, makePosition(2, 1), 0)

	// Unregister Type1
	removed := registry.Unregister("Type1")
	if !removed {
		t.Error("Expected to remove Type1")
	}

	if registry.Count() != 1 {
		t.Errorf("Expected 1 type after unregister, got %d", registry.Count())
	}

	// Verify Type1 is gone
	_, ok := registry.Resolve("Type1")
	if ok {
		t.Error("Expected not to find Type1 after unregister")
	}

	// Verify Type2 is still there
	_, ok = registry.Resolve("Type2")
	if !ok {
		t.Error("Expected to find Type2 after unregister")
	}

	// Try to unregister non-existent type
	removed = registry.Unregister("NonExistent")
	if removed {
		t.Error("Expected not to remove non-existent type")
	}
}

func TestTypeRegistry_UnregisterCaseInsensitive(t *testing.T) {
	registry := NewTypeRegistry()

	registry.Register("MyType", &types.IntegerType{}, makePosition(1, 1), 0)

	// Unregister with different case
	removed := registry.Unregister("mytype")
	if !removed {
		t.Error("Expected to remove type with case-insensitive name")
	}

	if registry.Count() != 0 {
		t.Errorf("Expected 0 types after unregister, got %d", registry.Count())
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestTypeRegistry_ComplexScenario(t *testing.T) {
	registry := NewTypeRegistry()

	// Register various types
	registry.Register("Integer", &types.IntegerType{}, makePosition(1, 1), 2)
	registry.Register("String", &types.StringType{}, makePosition(2, 1), 2)

	enum := &types.EnumType{
		Name: "TColor",
		Values: map[string]int{
			"Red":   0,
			"Green": 1,
			"Blue":  2,
		},
		OrderedNames: []string{"Red", "Green", "Blue"},
	}
	registry.Register("TColor", enum, makePosition(10, 1), 1)

	record := &types.RecordType{
		Name: "TPoint",
		Fields: map[string]types.Type{
			"X": &types.IntegerType{},
			"Y": &types.IntegerType{},
		},
	}
	registry.Register("TPoint", record, makePosition(20, 1), 1)

	class := &types.ClassType{Name: "TMyClass"}
	registry.Register("TMyClass", class, makePosition(30, 1), 0)

	// Test various queries
	if registry.Count() != 5 {
		t.Errorf("Expected 5 types, got %d", registry.Count())
	}

	// Test TypesByKind
	enums := registry.TypesByKind("ENUM")
	if len(enums) != 1 {
		t.Errorf("Expected 1 enum, got %d", len(enums))
	}

	// Test TypesInRange
	inRange := registry.TypesInRange(5, 25)
	if len(inRange) != 2 {
		t.Errorf("Expected 2 types in range 5-25, got %d", len(inRange))
	}

	// Test dependencies
	deps := registry.GetTypeDependencies("TPoint")
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies for TPoint, got %d", len(deps))
	}

	// Test case-insensitive lookup
	_, ok := registry.Resolve("tpoint")
	if !ok {
		t.Error("Expected to find TPoint with lowercase name")
	}

	// Test descriptor retrieval
	desc, ok := registry.ResolveDescriptor("TColor")
	if !ok {
		t.Fatal("Expected to find TColor descriptor")
	}
	if desc.Visibility != 1 {
		t.Errorf("Expected visibility 1 for TColor, got %d", desc.Visibility)
	}
}

// ============================================================================
// Kind Index Rebuild Tests
// ============================================================================

func TestTypeRegistry_KindIndexRebuild(t *testing.T) {
	registry := NewTypeRegistry()

	// Register some types and query by kind (builds index)
	registry.Register("Int1", &types.IntegerType{}, makePosition(1, 1), 0)
	integers := registry.TypesByKind("INTEGER")
	if len(integers) != 1 {
		t.Errorf("Expected 1 integer, got %d", len(integers))
	}

	// Register more types (should invalidate index)
	registry.Register("Int2", &types.IntegerType{}, makePosition(2, 1), 0)
	registry.Register("Int3", &types.IntegerType{}, makePosition(3, 1), 0)

	// Query again (should rebuild index)
	integers = registry.TypesByKind("INTEGER")
	if len(integers) != 3 {
		t.Errorf("Expected 3 integers after rebuild, got %d", len(integers))
	}

	// Unregister a type (should invalidate index)
	registry.Unregister("Int2")

	// Query again (should rebuild index)
	integers = registry.TypesByKind("INTEGER")
	if len(integers) != 2 {
		t.Errorf("Expected 2 integers after unregister, got %d", len(integers))
	}
}
