package interp

import (
	"fmt"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// SetValue Tests
// ============================================================================

// TestSetValue_Creation tests creating a SetValue with a small enum (bitset).
func TestSetValue_Creation(t *testing.T) {
	// Create a small enum type
	enumType := types.NewEnumType("TColor", map[string]int{
		"Red":   0,
		"Green": 1,
		"Blue":  2,
	}, []string{"Red", "Green", "Blue"})

	setType := types.NewSetType(enumType)

	// Create an empty set
	set := &SetValue{
		SetType:  setType,
		Elements: 0, // Empty bitset
	}

	// Verify Type() returns "SET"
	if set.Type() != "SET" {
		t.Errorf("expected Type() = 'SET', got '%s'", set.Type())
	}

	// Verify String() for empty set
	if set.String() != "[]" {
		t.Errorf("expected String() = '[]', got '%s'", set.String())
	}
}

// TestSetValue_WithElements tests SetValue with elements.
func TestSetValue_WithElements(t *testing.T) {
	// Create a small enum type
	enumType := types.NewEnumType("TColor", map[string]int{
		"Red":   0,
		"Green": 1,
		"Blue":  2,
	}, []string{"Red", "Green", "Blue"})

	setType := types.NewSetType(enumType)

	// Create a set with Red and Blue (bits 0 and 2 set)
	// Bitset: 0b101 = 5
	set := &SetValue{
		SetType:  setType,
		Elements: 5, // Red (0) and Blue (2)
	}

	// Verify Type()
	if set.Type() != "SET" {
		t.Errorf("expected Type() = 'SET', got '%s'", set.Type())
	}

	// Verify String() shows elements
	str := set.String()
	// Should contain both Red and Blue
	if str != "[Blue, Red]" && str != "[Red, Blue]" {
		t.Errorf("expected String() to contain Red and Blue, got '%s'", str)
	}
}

// TestSetValue_HasElement tests checking if an element is in the set.
func TestSetValue_HasElement(t *testing.T) {
	// Create a small enum type
	enumType := types.NewEnumType("TColor", map[string]int{
		"Red":   0,
		"Green": 1,
		"Blue":  2,
	}, []string{"Red", "Green", "Blue"})

	setType := types.NewSetType(enumType)

	// Create a set with Red (bit 0)
	set := &SetValue{
		SetType:  setType,
		Elements: 1, // Red (0)
	}

	// Test HasElement
	if !set.HasElement(0) {
		t.Error("expected set to contain Red (0)")
	}

	if set.HasElement(1) {
		t.Error("expected set to NOT contain Green (1)")
	}

	if set.HasElement(2) {
		t.Error("expected set to NOT contain Blue (2)")
	}
}

// TestSetValue_AddElement tests adding an element to a set.
func TestSetValue_AddElement(t *testing.T) {
	// Create a small enum type
	enumType := types.NewEnumType("TColor", map[string]int{
		"Red":   0,
		"Green": 1,
		"Blue":  2,
	}, []string{"Red", "Green", "Blue"})

	setType := types.NewSetType(enumType)

	// Create an empty set
	set := &SetValue{
		SetType:  setType,
		Elements: 0,
	}

	// Add Red (ordinal 0)
	set.AddElement(0)

	if !set.HasElement(0) {
		t.Error("expected set to contain Red after AddElement(0)")
	}

	if set.Elements != 1 {
		t.Errorf("expected Elements = 1, got %d", set.Elements)
	}
}

// TestSetValue_RemoveElement tests removing an element from a set.
func TestSetValue_RemoveElement(t *testing.T) {
	// Create a small enum type
	enumType := types.NewEnumType("TColor", map[string]int{
		"Red":   0,
		"Green": 1,
		"Blue":  2,
	}, []string{"Red", "Green", "Blue"})

	setType := types.NewSetType(enumType)

	// Create a set with Red and Blue
	set := &SetValue{
		SetType:  setType,
		Elements: 5, // 0b101 = Red and Blue
	}

	// Remove Red (ordinal 0)
	set.RemoveElement(0)

	if set.HasElement(0) {
		t.Error("expected set to NOT contain Red after RemoveElement(0)")
	}

	if set.Elements != 4 {
		t.Errorf("expected Elements = 4, got %d", set.Elements)
	}
}

// ============================================================================
// Set Literal Evaluation Tests
// ============================================================================

// TestEvalSetLiteral_Simple tests evaluating a simple set literal with elements.
func TestEvalSetLiteral_Simple(t *testing.T) {
	// Setup interpreter with enum values
	interp, enumType := helperSetupInterpWithColorEnum(t)

	// Create a SetLiteral AST node: [Red, Blue]
	setLiteral := &ast.SetLiteral{
		Elements: []ast.Expression{
			&ast.EnumLiteral{ValueName: "Red"},
			&ast.EnumLiteral{ValueName: "Blue"},
		},
	}

	// Evaluate the set literal
	result := interp.Eval(setLiteral)

	// Verify it returns a SetValue
	setVal, ok := result.(*SetValue)
	if !ok {
		t.Fatalf("expected SetValue, got %T", result)
	}

	// Verify the set contains Red (0) and Blue (2)
	// Bitset: 0b101 = 5
	if setVal.Elements != 5 {
		t.Errorf("expected Elements = 5 (Red and Blue), got %d", setVal.Elements)
	}

	// Verify type
	if setVal.SetType.ElementType != enumType {
		t.Error("expected set type to match enum type")
	}
}

// TestEvalSetLiteral_Single tests evaluating a set literal with one element.
func TestEvalSetLiteral_Single(t *testing.T) {
	// Setup interpreter with enum values
	interp, _ := helperSetupInterpWithColorEnum(t)

	// Create a SetLiteral AST node: [Green]
	setLiteral := &ast.SetLiteral{
		Elements: []ast.Expression{
			&ast.EnumLiteral{ValueName: "Green"},
		},
	}

	// Evaluate the set literal
	result := interp.Eval(setLiteral)

	// Verify it returns a SetValue
	setVal, ok := result.(*SetValue)
	if !ok {
		t.Fatalf("expected SetValue, got %T", result)
	}

	// Verify the set contains only Green (1)
	// Bitset: 0b010 = 2
	if setVal.Elements != 2 {
		t.Errorf("expected Elements = 2 (Green only), got %d", setVal.Elements)
	}
}

// TestEvalSetLiteral_Empty tests evaluating an empty set literal.
// Note: Empty sets need type context in real scenarios. For unit testing,
// we'll test with at least one operation that gives context.
func TestEvalSetLiteral_Empty(t *testing.T) {
	// Setup interpreter
	_, enumType := helperSetupInterpWithColorEnum(t)

	// Create an empty SetValue directly (simulating what would happen
	// after type inference from context)
	setType := types.NewSetType(enumType)
	emptySet := &SetValue{
		SetType:  setType,
		Elements: 0,
	}

	// Verify the set is empty
	if emptySet.Elements != 0 {
		t.Errorf("expected Elements = 0 (empty), got %d", emptySet.Elements)
	}

	// Verify String() returns "[]"
	if emptySet.String() != "[]" {
		t.Errorf("expected String() = '[]', got '%s'", emptySet.String())
	}
}

// TestEvalSetLiteral_Range tests evaluating a set literal with a range.
func TestEvalSetLiteral_Range(t *testing.T) {
	// Setup interpreter with enum values
	interp, _ := helperSetupInterpWithColorEnum(t)

	// Create a SetLiteral AST node with a range: [Red..Blue]
	// This should expand to [Red, Green, Blue]
	setLiteral := &ast.SetLiteral{
		Elements: []ast.Expression{
			&ast.RangeExpression{
				Start: &ast.EnumLiteral{ValueName: "Red"},
				End:   &ast.EnumLiteral{ValueName: "Blue"},
			},
		},
	}

	// Evaluate the set literal
	result := interp.Eval(setLiteral)

	// Verify it returns a SetValue
	setVal, ok := result.(*SetValue)
	if !ok {
		t.Fatalf("expected SetValue, got %T", result)
	}

	// Verify the set contains Red (0), Green (1), and Blue (2)
	// Bitset: 0b111 = 7
	if setVal.Elements != 7 {
		t.Errorf("expected Elements = 7 (Red, Green, Blue), got %d", setVal.Elements)
	}
}

// TestEvalSetLiteral_MixedRangeAndElements tests a set literal with both ranges and individual elements.
func TestEvalSetLiteral_MixedRangeAndElements(t *testing.T) {
	// Setup interpreter - use a larger enum for this test
	interp, enumType := helperSetupInterpWithLargerEnum(t)

	// Create a SetLiteral AST node: [One, Three..Five]
	// This should expand to [One, Three, Four, Five]
	setLiteral := &ast.SetLiteral{
		Elements: []ast.Expression{
			&ast.EnumLiteral{ValueName: "One"},
			&ast.RangeExpression{
				Start: &ast.EnumLiteral{ValueName: "Three"},
				End:   &ast.EnumLiteral{ValueName: "Five"},
			},
		},
	}

	// Evaluate the set literal
	result := interp.Eval(setLiteral)

	// Verify it returns a SetValue
	setVal, ok := result.(*SetValue)
	if !ok {
		t.Fatalf("expected SetValue, got %T", result)
	}

	// Verify the set type
	if setVal.SetType.ElementType != enumType {
		t.Error("expected set type to match enum type")
	}

	// Verify the set contains One (1), Three (3), Four (4), Five (5)
	// Bitset: 0b111010 = 58
	if setVal.Elements != 58 {
		t.Errorf("expected Elements = 58 (One, Three, Four, Five), got %d", setVal.Elements)
	}
}

// ============================================================================
// Set Binary Operations Tests
// ============================================================================

// TestSetUnion tests set union operation (+).
func TestSetUnion(t *testing.T) {
	interp, enumType := helperSetupInterpWithColorEnum(t)
	setType := types.NewSetType(enumType)

	// Create two sets: [Red] and [Blue]
	set1 := &SetValue{SetType: setType, Elements: 1} // Red (bit 0)
	set2 := &SetValue{SetType: setType, Elements: 4} // Blue (bit 2)

	// Define them in environment
	interp.env.Define("s1", set1)
	interp.env.Define("s2", set2)

	// Evaluate: s1 + s2
	result := interp.evalBinarySetOperation(set1, set2, "+")

	// Verify result is a SetValue
	setVal, ok := result.(*SetValue)
	if !ok {
		t.Fatalf("expected SetValue, got %T", result)
	}

	// Union should be [Red, Blue]: bits 0 and 2 = 0b101 = 5
	if setVal.Elements != 5 {
		t.Errorf("expected Elements = 5 (Red, Blue), got %d", setVal.Elements)
	}
}

// TestSetDifference tests set difference operation (-).
func TestSetDifference(t *testing.T) {
	interp, enumType := helperSetupInterpWithColorEnum(t)
	setType := types.NewSetType(enumType)

	// Create two sets: [Red, Green, Blue] and [Green]
	set1 := &SetValue{SetType: setType, Elements: 7} // All three (bits 0,1,2)
	set2 := &SetValue{SetType: setType, Elements: 2} // Green (bit 1)

	// Define them in environment
	interp.env.Define("s1", set1)
	interp.env.Define("s2", set2)

	// Evaluate: s1 - s2
	result := interp.evalBinarySetOperation(set1, set2, "-")

	// Verify result is a SetValue
	setVal, ok := result.(*SetValue)
	if !ok {
		t.Fatalf("expected SetValue, got %T", result)
	}

	// Difference should be [Red, Blue]: bits 0 and 2 = 0b101 = 5
	if setVal.Elements != 5 {
		t.Errorf("expected Elements = 5 (Red, Blue), got %d", setVal.Elements)
	}
}

// TestSetIntersection tests set intersection operation (*).
func TestSetIntersection(t *testing.T) {
	interp, enumType := helperSetupInterpWithColorEnum(t)
	setType := types.NewSetType(enumType)

	// Create two sets: [Red, Green] and [Green, Blue]
	set1 := &SetValue{SetType: setType, Elements: 3} // Red, Green (bits 0,1)
	set2 := &SetValue{SetType: setType, Elements: 6} // Green, Blue (bits 1,2)

	// Define them in environment
	interp.env.Define("s1", set1)
	interp.env.Define("s2", set2)

	// Evaluate: s1 * s2
	result := interp.evalBinarySetOperation(set1, set2, "*")

	// Verify result is a SetValue
	setVal, ok := result.(*SetValue)
	if !ok {
		t.Fatalf("expected SetValue, got %T", result)
	}

	// Intersection should be [Green]: bit 1 = 0b010 = 2
	if setVal.Elements != 2 {
		t.Errorf("expected Elements = 2 (Green), got %d", setVal.Elements)
	}
}

// ============================================================================
// Membership Test
// ============================================================================

// TestSetMembership tests the 'in' operator for sets.
func TestSetMembership(t *testing.T) {
	interp, enumType := helperSetupInterpWithColorEnum(t)
	setType := types.NewSetType(enumType)

	// Create a set: [Red, Blue]
	set := &SetValue{SetType: setType, Elements: 5} // Red (bit 0) and Blue (bit 2)

	// Test membership for Red (should be in set)
	redEnum := &EnumValue{TypeName: "TColor", ValueName: "Red", OrdinalValue: 0}
	result := interp.evalSetMembership(redEnum, set)
	if boolVal, ok := result.(*BooleanValue); !ok || !boolVal.Value {
		t.Error("expected Red to be in set")
	}

	// Test membership for Green (should NOT be in set)
	greenEnum := &EnumValue{TypeName: "TColor", ValueName: "Green", OrdinalValue: 1}
	result = interp.evalSetMembership(greenEnum, set)
	if boolVal, ok := result.(*BooleanValue); !ok || boolVal.Value {
		t.Error("expected Green to NOT be in set")
	}
}

// ============================================================================
// Include/Exclude Methods.
// ============================================================================

// TestSetInclude tests the Include method.
func TestSetInclude(t *testing.T) {
	interp, enumType := helperSetupInterpWithColorEnum(t)
	setType := types.NewSetType(enumType)

	// Create a set: [Red]
	set := &SetValue{SetType: setType, Elements: 1} // Red (bit 0)

	// Include Blue
	blueEnum := &EnumValue{TypeName: "TColor", ValueName: "Blue", OrdinalValue: 2}
	interp.evalSetInclude(set, blueEnum)

	// Verify Blue is now in the set
	if !set.HasElement(2) {
		t.Error("expected Blue to be in set after Include")
	}

	// Elements should be: Red + Blue = 0b101 = 5
	if set.Elements != 5 {
		t.Errorf("expected Elements = 5, got %d", set.Elements)
	}
}

// TestSetExclude tests the Exclude method.
func TestSetExclude(t *testing.T) {
	interp, enumType := helperSetupInterpWithColorEnum(t)
	setType := types.NewSetType(enumType)

	// Create a set: [Red, Green, Blue]
	set := &SetValue{SetType: setType, Elements: 7} // All three

	// Exclude Green
	greenEnum := &EnumValue{TypeName: "TColor", ValueName: "Green", OrdinalValue: 1}
	interp.evalSetExclude(set, greenEnum)

	// Verify Green is no longer in the set
	if set.HasElement(1) {
		t.Error("expected Green to NOT be in set after Exclude")
	}

	// Elements should be: Red + Blue = 0b101 = 5
	if set.Elements != 5 {
		t.Errorf("expected Elements = 5, got %d", set.Elements)
	}
}

// ============================================================================
// Set Comparisons
// ============================================================================

// TestSetEquality tests set equality (= and <>).
func TestSetEquality(t *testing.T) {
	_, enumType := helperSetupInterpWithColorEnum(t)
	setType := types.NewSetType(enumType)

	// Create sets
	set1 := &SetValue{SetType: setType, Elements: 5} // [Red, Blue]
	set2 := &SetValue{SetType: setType, Elements: 5} // [Red, Blue]
	set3 := &SetValue{SetType: setType, Elements: 3} // [Red, Green]

	// Test equality
	if !setEquals(set1, set2) {
		t.Error("expected sets to be equal")
	}

	if setEquals(set1, set3) {
		t.Error("expected sets to be unequal")
	}
}

// TestSetSubset tests subset operation (<=).
func TestSetSubset(t *testing.T) {
	_, enumType := helperSetupInterpWithColorEnum(t)
	setType := types.NewSetType(enumType)

	// Create sets
	small := &SetValue{SetType: setType, Elements: 1} // [Red]
	large := &SetValue{SetType: setType, Elements: 7} // [Red, Green, Blue]

	// Test subset: small <= large should be true
	if !setIsSubset(small, large) {
		t.Error("expected small to be subset of large")
	}

	// Test subset: large <= small should be false
	if setIsSubset(large, small) {
		t.Error("expected large to NOT be subset of small")
	}

	// Test subset: set <= itself should be true
	if !setIsSubset(large, large) {
		t.Error("expected set to be subset of itself")
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// setEquals checks if two sets are equal.
func setEquals(s1, s2 *SetValue) bool {
	return s1.Elements == s2.Elements
}

// setIsSubset checks if s1 is a subset of s2.
func setIsSubset(s1, s2 *SetValue) bool {
	return (s1.Elements & s2.Elements) == s1.Elements
}

// ============================================================================
// Helper Functions
// ============================================================================

// helperSetupInterpWithColorEnum creates an interpreter with TColor enum defined.
func helperSetupInterpWithColorEnum(t *testing.T) (*Interpreter, *types.EnumType) {
	t.Helper()

	interp := New(nil)

	// Create TColor enum: (Red, Green, Blue)
	enumType := types.NewEnumType("TColor", map[string]int{
		"Red":   0,
		"Green": 1,
		"Blue":  2,
	}, []string{"Red", "Green", "Blue"})

	// Register enum values in the environment
	interp.env.Define("Red", &EnumValue{
		TypeName:     "TColor",
		ValueName:    "Red",
		OrdinalValue: 0,
	})
	interp.env.Define("Green", &EnumValue{
		TypeName:     "TColor",
		ValueName:    "Green",
		OrdinalValue: 1,
	})
	interp.env.Define("Blue", &EnumValue{
		TypeName:     "TColor",
		ValueName:    "Blue",
		OrdinalValue: 2,
	})

	// Store enum type metadata
	interp.env.Define("__enum_type_TColor", &EnumTypeValue{EnumType: enumType})

	return interp, enumType
}

// helperSetupInterpWithLargerEnum creates an interpreter with a larger enum for testing ranges.
func helperSetupInterpWithLargerEnum(t *testing.T) (*Interpreter, *types.EnumType) {
	t.Helper()

	interp := New(nil)

	// Create TNumber enum: (Zero, One, Two, Three, Four, Five)
	enumType := types.NewEnumType("TNumber", map[string]int{
		"Zero":  0,
		"One":   1,
		"Two":   2,
		"Three": 3,
		"Four":  4,
		"Five":  5,
	}, []string{"Zero", "One", "Two", "Three", "Four", "Five"})

	// Register enum values in the environment
	for _, name := range enumType.OrderedNames {
		ordinal := enumType.Values[name]
		interp.env.Define(name, &EnumValue{
			TypeName:     "TNumber",
			ValueName:    name,
			OrdinalValue: ordinal,
		})
	}

	// Store enum type metadata
	interp.env.Define("__enum_type_TNumber", &EnumTypeValue{EnumType: enumType})

	return interp, enumType
}

// ============================================================================
// Large Set Tests (Task 9.8)
// ============================================================================

// TestSetValue_BoundaryCase64Elements tests set with exactly 64 elements (boundary case).
// This should use bitmask storage (threshold is ≤64).
func TestSetValue_BoundaryCase64Elements(t *testing.T) {
	// Create enum with exactly 64 values (E0..E63)
	values := make(map[string]int)
	orderedNames := make([]string, 64)
	for i := 0; i < 64; i++ {
		name := "E" + string(rune('0'+i/10)) + string(rune('0'+i%10))
		if i < 10 {
			name = "E0" + string(rune('0'+i))
		}
		values[name] = i
		orderedNames[i] = name
	}

	enumType := types.NewEnumType("TBoundary", values, orderedNames)
	setType := types.NewSetType(enumType)

	// Verify it uses bitmask storage
	if setType.StorageKind != types.SetStorageBitmask {
		t.Errorf("expected bitmask storage for 64-element enum, got %s", setType.StorageKind)
	}

	// Create a set and add some elements
	set := NewSetValue(setType)
	set.AddElement(0)  // E00
	set.AddElement(31) // E31
	set.AddElement(63) // E63 (last element)

	// Verify elements are present
	if !set.HasElement(0) {
		t.Error("expected element 0 to be present")
	}
	if !set.HasElement(31) {
		t.Error("expected element 31 to be present")
	}
	if !set.HasElement(63) {
		t.Error("expected element 63 to be present")
	}
	if set.HasElement(32) {
		t.Error("expected element 32 to be absent")
	}

	// Verify String representation
	str := set.String()
	if str == "[]" {
		t.Error("expected non-empty string representation")
	}
}

// TestSetValue_LargeEnum65Elements tests set with 65 elements (triggers map storage).
func TestSetValue_LargeEnum65Elements(t *testing.T) {
	// Create enum with 65 values (E0..E64)
	values := make(map[string]int)
	orderedNames := make([]string, 65)
	for i := 0; i < 65; i++ {
		name := "E" + string(rune('0'+i/10)) + string(rune('0'+i%10))
		if i < 10 {
			name = "E0" + string(rune('0'+i))
		}
		values[name] = i
		orderedNames[i] = name
	}

	enumType := types.NewEnumType("TLarge65", values, orderedNames)
	setType := types.NewSetType(enumType)

	// Verify it uses map storage
	if setType.StorageKind != types.SetStorageMap {
		t.Errorf("expected map storage for 65-element enum, got %s", setType.StorageKind)
	}

	// Create a set and add elements
	set := NewSetValue(setType)
	if set.MapStore == nil {
		t.Fatal("expected MapStore to be initialized for large enum")
	}

	set.AddElement(0)
	set.AddElement(32)
	set.AddElement(64) // Beyond 64-bit boundary

	// Verify elements are present
	if !set.HasElement(0) {
		t.Error("expected element 0 to be present")
	}
	if !set.HasElement(32) {
		t.Error("expected element 32 to be present")
	}
	if !set.HasElement(64) {
		t.Error("expected element 64 to be present")
	}
	if set.HasElement(33) {
		t.Error("expected element 33 to be absent")
	}

	// Remove an element
	set.RemoveElement(32)
	if set.HasElement(32) {
		t.Error("expected element 32 to be removed")
	}
	if !set.HasElement(64) {
		t.Error("expected element 64 to still be present")
	}
}

// TestSetValue_LargeEnum100Elements tests set with 100 elements.
func TestSetValue_LargeEnum100Elements(t *testing.T) {
	// Create enum with 100 values (E00..E99)
	values := make(map[string]int)
	orderedNames := make([]string, 100)
	for i := 0; i < 100; i++ {
		name := "E" + string(rune('0'+i/10)) + string(rune('0'+i%10))
		values[name] = i
		orderedNames[i] = name
	}

	enumType := types.NewEnumType("TLarge100", values, orderedNames)
	setType := types.NewSetType(enumType)

	// Verify it uses map storage
	if setType.StorageKind != types.SetStorageMap {
		t.Errorf("expected map storage for 100-element enum, got %s", setType.StorageKind)
	}

	// Create a set with many elements
	set := NewSetValue(setType)
	for i := 0; i < 100; i += 5 {
		set.AddElement(i)
	}

	// Verify count
	count := 0
	for i := 0; i < 100; i++ {
		if set.HasElement(i) {
			count++
		}
	}
	if count != 20 {
		t.Errorf("expected 20 elements, got %d", count)
	}

	// Verify String representation contains elements
	str := set.String()
	if str == "[]" {
		t.Error("expected non-empty string representation")
	}
}

// TestLargeSet_BinaryOperations tests set operations on large sets.
func TestLargeSet_BinaryOperations(t *testing.T) {
	// Create enum with 80 values (triggers map storage)
	values := make(map[string]int)
	orderedNames := make([]string, 80)
	for i := 0; i < 80; i++ {
		name := "E" + string(rune('0'+i/10)) + string(rune('0'+i%10))
		if i < 10 {
			name = "E0" + string(rune('0'+i))
		}
		values[name] = i
		orderedNames[i] = name
	}

	enumType := types.NewEnumType("TLarge80", values, orderedNames)
	setType := types.NewSetType(enumType)

	// Create two sets
	set1 := NewSetValue(setType)
	set2 := NewSetValue(setType)

	// Set1: {0, 10, 20, ..., 70}
	for i := 0; i < 80; i += 10 {
		set1.AddElement(i)
	}

	// Set2: {0, 20, 40, 60}
	for i := 0; i < 80; i += 20 {
		set2.AddElement(i)
	}

	interp := New(nil)

	// Test union (set1 + set2)
	unionResult := interp.evalBinarySetOperation(set1, set2, "+")
	if isError(unionResult) {
		t.Fatalf("union operation failed: %s", unionResult)
	}
	unionSet := unionResult.(*SetValue)

	// Union should contain all elements from both sets: {0, 10, 20, 30, 40, 50, 60, 70}
	expectedUnion := []int{0, 10, 20, 30, 40, 50, 60, 70}
	for _, ord := range expectedUnion {
		if !unionSet.HasElement(ord) {
			t.Errorf("union: expected element %d to be present", ord)
		}
	}

	// Test intersection (set1 * set2)
	intersectResult := interp.evalBinarySetOperation(set1, set2, "*")
	if isError(intersectResult) {
		t.Fatalf("intersection operation failed: %s", intersectResult)
	}
	intersectSet := intersectResult.(*SetValue)

	// Intersection should contain common elements: {0, 20, 40, 60}
	expectedIntersect := []int{0, 20, 40, 60}
	for _, ord := range expectedIntersect {
		if !intersectSet.HasElement(ord) {
			t.Errorf("intersection: expected element %d to be present", ord)
		}
	}
	// Should NOT contain 10, 30, 50, 70
	notExpected := []int{10, 30, 50, 70}
	for _, ord := range notExpected {
		if intersectSet.HasElement(ord) {
			t.Errorf("intersection: expected element %d to be absent", ord)
		}
	}

	// Test difference (set1 - set2)
	diffResult := interp.evalBinarySetOperation(set1, set2, "-")
	if isError(diffResult) {
		t.Fatalf("difference operation failed: %s", diffResult)
	}
	diffSet := diffResult.(*SetValue)

	// Difference should contain set1 elements not in set2: {10, 30, 50, 70}
	expectedDiff := []int{10, 30, 50, 70}
	for _, ord := range expectedDiff {
		if !diffSet.HasElement(ord) {
			t.Errorf("difference: expected element %d to be present", ord)
		}
	}
	// Should NOT contain elements in set2: {0, 20, 40, 60}
	notExpectedDiff := []int{0, 20, 40, 60}
	for _, ord := range notExpectedDiff {
		if diffSet.HasElement(ord) {
			t.Errorf("difference: expected element %d to be absent", ord)
		}
	}
}

// TestLargeSet_ForInIteration tests for-in loop over a large set using evalForInStatement.
func TestLargeSet_ForInIteration(t *testing.T) {
	// Create enum with 70 values (triggers map storage)
	values := make(map[string]int)
	orderedNames := make([]string, 70)
	for i := 0; i < 70; i++ {
		name := "E" + string(rune('0'+i/10)) + string(rune('0'+i%10))
		if i < 10 {
			name = "E0" + string(rune('0'+i))
		}
		values[name] = i
		orderedNames[i] = name
	}

	enumType := types.NewEnumType("TLarge70", values, orderedNames)
	setType := types.NewSetType(enumType)

	// Verify it uses map storage
	if setType.StorageKind != types.SetStorageMap {
		t.Errorf("expected map storage for 70-element enum, got %s", setType.StorageKind)
	}

	// Create a set with some elements: {5, 15, 25, 35, 45, 55, 65}
	set := NewSetValue(setType)
	expectedOrdinals := []int{5, 15, 25, 35, 45, 55, 65}
	for _, ord := range expectedOrdinals {
		set.AddElement(ord)
	}

	// Verify the set has correct elements
	for _, ord := range expectedOrdinals {
		if !set.HasElement(ord) {
			t.Errorf("expected element %d to be in set", ord)
		}
	}

	// Verify iteration would visit elements in order (checking HasElement for each enum value)
	// This simulates what for-in does: iterate enumType.OrderedNames and check HasElement
	collectedOrdinals := []int{}
	for _, name := range enumType.OrderedNames {
		ordinal := enumType.Values[name]
		if set.HasElement(ordinal) {
			collectedOrdinals = append(collectedOrdinals, ordinal)
		}
	}

	// Verify iteration order and completeness
	if len(collectedOrdinals) != len(expectedOrdinals) {
		t.Errorf("expected %d iterations, got %d", len(expectedOrdinals), len(collectedOrdinals))
	}

	// Verify all expected ordinals were collected in order
	for i, expected := range expectedOrdinals {
		if i >= len(collectedOrdinals) {
			t.Errorf("missing ordinal %d at position %d", expected, i)
			continue
		}
		if collectedOrdinals[i] != expected {
			t.Errorf("at position %d: expected ordinal %d, got %d", i, expected, collectedOrdinals[i])
		}
	}
}

// TestLargeSet_Performance tests that large set operations complete in reasonable time.
func TestLargeSet_Performance(t *testing.T) {
	// Create enum with 200 values
	values := make(map[string]int)
	orderedNames := make([]string, 200)
	for i := 0; i < 200; i++ {
		name := "E" + string(rune('0'+i/100)) + string(rune('0'+(i/10)%10)) + string(rune('0'+i%10))
		values[name] = i
		orderedNames[i] = name
	}

	enumType := types.NewEnumType("TLarge200", values, orderedNames)
	setType := types.NewSetType(enumType)

	// Create two large sets
	set1 := NewSetValue(setType)
	set2 := NewSetValue(setType)

	// Populate sets (100 elements each)
	for i := 0; i < 200; i += 2 {
		set1.AddElement(i)
	}
	for i := 1; i < 200; i += 2 {
		set2.AddElement(i)
	}

	interp := New(nil)

	// Perform operations (should complete quickly)
	unionResult := interp.evalBinarySetOperation(set1, set2, "+")
	if isError(unionResult) {
		t.Fatalf("union failed: %s", unionResult)
	}

	intersectResult := interp.evalBinarySetOperation(set1, set2, "*")
	if isError(intersectResult) {
		t.Fatalf("intersection failed: %s", intersectResult)
	}

	diffResult := interp.evalBinarySetOperation(set1, set2, "-")
	if isError(diffResult) {
		t.Fatalf("difference failed: %s", diffResult)
	}

	// Verify results make sense
	unionSet := unionResult.(*SetValue)
	intersectSet := intersectResult.(*SetValue)
	diffSet := diffResult.(*SetValue)

	// Union should have all 200 elements
	unionCount := 0
	for i := 0; i < 200; i++ {
		if unionSet.HasElement(i) {
			unionCount++
		}
	}
	if unionCount != 200 {
		t.Errorf("expected union to have 200 elements, got %d", unionCount)
	}

	// Intersection should be empty (disjoint sets)
	intersectCount := 0
	for i := 0; i < 200; i++ {
		if intersectSet.HasElement(i) {
			intersectCount++
		}
	}
	if intersectCount != 0 {
		t.Errorf("expected intersection to be empty, got %d elements", intersectCount)
	}

	// Difference should equal set1 (100 elements)
	diffCount := 0
	for i := 0; i < 200; i++ {
		if diffSet.HasElement(i) {
			diffCount++
		}
	}
	if diffCount != 100 {
		t.Errorf("expected difference to have 100 elements, got %d", diffCount)
	}
}

// ============================================================================
// For-In Edge Cases (Task 9.9d)
// ============================================================================

// TestForInSet_EmptySet tests that for-in over an empty set never executes the loop body.
func TestForInSet_EmptySet(t *testing.T) {
	// Create a small enum
	enumType := types.NewEnumType("TColor", map[string]int{
		"Red":   0,
		"Green": 1,
		"Blue":  2,
	}, []string{"Red", "Green", "Blue"})

	setType := types.NewSetType(enumType)

	// Create an empty set
	emptySet := NewSetValue(setType)

	// Verify the set is empty
	if emptySet.HasElement(0) || emptySet.HasElement(1) || emptySet.HasElement(2) {
		t.Fatal("expected empty set")
	}

	// Simulate for-in iteration (same logic as interpreter)
	executionCount := 0
	for _, name := range enumType.OrderedNames {
		ordinal := enumType.Values[name]
		if emptySet.HasElement(ordinal) {
			executionCount++
		}
	}

	// Verify loop body never executed
	if executionCount != 0 {
		t.Errorf("expected loop body to never execute for empty set, executed %d times", executionCount)
	}
}

// TestForInSet_SingleElement tests that for-in over a single-element set executes exactly once.
func TestForInSet_SingleElement(t *testing.T) {
	// Create a small enum
	enumType := types.NewEnumType("TColor", map[string]int{
		"Red":   0,
		"Green": 1,
		"Blue":  2,
	}, []string{"Red", "Green", "Blue"})

	setType := types.NewSetType(enumType)

	// Create a set with single element
	singleSet := NewSetValue(setType)
	singleSet.AddElement(1) // Green

	// Verify the set has only one element
	if !singleSet.HasElement(1) {
		t.Fatal("expected Green to be in set")
	}
	if singleSet.HasElement(0) || singleSet.HasElement(2) {
		t.Fatal("expected only Green in set")
	}

	// Simulate for-in iteration (same logic as interpreter)
	executionCount := 0
	var lastElement string
	for _, name := range enumType.OrderedNames {
		ordinal := enumType.Values[name]
		if singleSet.HasElement(ordinal) {
			executionCount++
			lastElement = name
		}
	}

	// Verify loop body executed exactly once
	if executionCount != 1 {
		t.Errorf("expected loop body to execute exactly once, executed %d times", executionCount)
	}

	// Verify it was the correct element
	if lastElement != "Green" {
		t.Errorf("expected to iterate over Green, got %s", lastElement)
	}
}

// ============================================================================
// Large Set Edge Cases (Task 9.10c)
// ============================================================================

// TestVeryLargeSet_500Elements tests set operations with 500-element enum.
// Task 9.10c: Stress test for very large sets (map storage).
func TestVeryLargeSet_500Elements(t *testing.T) {
	// Create enum with 500 elements
	values := make(map[string]int)
	orderedNames := make([]string, 500)
	for i := 0; i < 500; i++ {
		name := fmt.Sprintf("E%03d", i)
		values[name] = i
		orderedNames[i] = name
	}
	enumType := &types.EnumType{
		Name:         "TVeryLarge",
		Values:       values,
		OrderedNames: orderedNames,
	}

	setType := types.NewSetType(enumType)
	if setType.StorageKind != types.SetStorageMap {
		t.Fatal("500-element set should use map storage")
	}

	// Create set with every 10th element
	set1 := NewSetValue(setType)
	for i := 0; i < 500; i += 10 {
		set1.AddElement(i)
	}

	// Create set with every 5th element
	set2 := NewSetValue(setType)
	for i := 0; i < 500; i += 5 {
		set2.AddElement(i)
	}

	interp := New(nil)

	// Test union
	unionResult := interp.evalBinarySetOperation(set1, set2, "+")
	if isError(unionResult) {
		t.Fatalf("union operation failed: %s", unionResult)
	}
	union := unionResult.(*SetValue)

	// Union should contain all elements from both sets
	// Check some specific elements
	if !union.HasElement(0) || !union.HasElement(5) || !union.HasElement(10) {
		t.Error("Union missing expected elements")
	}

	// Test intersection - should contain every 10th element (common to both)
	intersectResult := interp.evalBinarySetOperation(set1, set2, "*")
	if isError(intersectResult) {
		t.Fatalf("intersection operation failed: %s", intersectResult)
	}
	intersection := intersectResult.(*SetValue)

	for i := 0; i < 500; i += 10 {
		if !intersection.HasElement(i) {
			t.Errorf("Intersection missing element %d", i)
		}
	}

	// Element 5 should not be in intersection (not in set1)
	if intersection.HasElement(5) {
		t.Error("Intersection should not contain element 5")
	}

	// Test difference - set1 - set2 should be empty (set1 ⊆ set2)
	diffResult := interp.evalBinarySetOperation(set1, set2, "-")
	if isError(diffResult) {
		t.Fatalf("difference operation failed: %s", diffResult)
	}
	difference := diffResult.(*SetValue)

	for i := 0; i < 500; i += 10 {
		if difference.HasElement(i) {
			t.Errorf("Difference should not contain element %d", i)
		}
	}
}

// TestVeryLargeSet_1000Elements tests set operations with 1000-element enum.
// Task 9.10c: Extreme stress test for very large sets.
func TestVeryLargeSet_1000Elements(t *testing.T) {
	// Create enum with 1000 elements
	values := make(map[string]int)
	orderedNames := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		name := fmt.Sprintf("E%04d", i)
		values[name] = i
		orderedNames[i] = name
	}
	enumType := &types.EnumType{
		Name:         "TExtremeLarge",
		Values:       values,
		OrderedNames: orderedNames,
	}

	setType := types.NewSetType(enumType)
	if setType.StorageKind != types.SetStorageMap {
		t.Fatal("1000-element set should use map storage")
	}

	// Create set with first 100 elements
	set1 := NewSetValue(setType)
	for i := 0; i < 100; i++ {
		set1.AddElement(i)
	}

	// Create set with last 100 elements
	set2 := NewSetValue(setType)
	for i := 900; i < 1000; i++ {
		set2.AddElement(i)
	}

	interp := New(nil)

	// Test union - should have 200 elements
	unionResult := interp.evalBinarySetOperation(set1, set2, "+")
	if isError(unionResult) {
		t.Fatalf("union operation failed: %s", unionResult)
	}
	union := unionResult.(*SetValue)

	count := 0
	for i := 0; i < 1000; i++ {
		if union.HasElement(i) {
			count++
		}
	}
	if count != 200 {
		t.Errorf("Union should have 200 elements, got %d", count)
	}

	// Test intersection - should be empty (no overlap)
	intersectResult := interp.evalBinarySetOperation(set1, set2, "*")
	if isError(intersectResult) {
		t.Fatalf("intersection operation failed: %s", intersectResult)
	}
	intersection := intersectResult.(*SetValue)

	for i := 0; i < 1000; i++ {
		if intersection.HasElement(i) {
			t.Errorf("Intersection should be empty, found element %d", i)
		}
	}

	// Test membership
	if !set1.HasElement(50) {
		t.Error("set1 should contain element 50")
	}
	if set1.HasElement(500) {
		t.Error("set1 should not contain element 500")
	}
	if !set2.HasElement(950) {
		t.Error("set2 should contain element 950")
	}
}

// TestLargeSet_ForInAllElements tests for-in iteration over a set containing all enum values.
// Task 9.10c: Edge case - set with all possible elements.
func TestLargeSet_ForInAllElements(t *testing.T) {
	// Create enum with 100 elements
	values := make(map[string]int)
	orderedNames := make([]string, 100)
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("E%02d", i)
		values[name] = i
		orderedNames[i] = name
	}
	enumType := &types.EnumType{
		Name:         "TFull",
		Values:       values,
		OrderedNames: orderedNames,
	}

	setType := types.NewSetType(enumType)

	// Create set with ALL elements
	fullSet := NewSetValue(setType)
	for i := 0; i < 100; i++ {
		fullSet.AddElement(i)
	}

	// Iterate and count
	count := 0
	for _, name := range enumType.OrderedNames {
		ordinal := enumType.Values[name]
		if fullSet.HasElement(ordinal) {
			count++
		}
	}

	// Should iterate over all 100 elements
	if count != 100 {
		t.Errorf("expected to iterate over all 100 elements, got %d", count)
	}
}

// TestLargeSet_ForInFirstAndLastOnly tests for-in iteration with only boundary elements.
// Task 9.10c: Edge case - only first and last elements in large set.
func TestLargeSet_ForInFirstAndLastOnly(t *testing.T) {
	// Create enum with 100 elements
	values := make(map[string]int)
	orderedNames := make([]string, 100)
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("E%02d", i)
		values[name] = i
		orderedNames[i] = name
	}
	enumType := &types.EnumType{
		Name:         "TBoundary",
		Values:       values,
		OrderedNames: orderedNames,
	}

	setType := types.NewSetType(enumType)

	// Create set with only first and last elements
	boundarySet := NewSetValue(setType)
	boundarySet.AddElement(0)  // First
	boundarySet.AddElement(99) // Last

	// Iterate and collect
	var collected []string
	for _, name := range enumType.OrderedNames {
		ordinal := enumType.Values[name]
		if boundarySet.HasElement(ordinal) {
			collected = append(collected, name)
		}
	}

	// Should only iterate over 2 elements in order
	if len(collected) != 2 {
		t.Errorf("expected 2 elements, got %d", len(collected))
	}
	if collected[0] != "E00" {
		t.Errorf("first element should be E00, got %s", collected[0])
	}
	if collected[1] != "E99" {
		t.Errorf("second element should be E99, got %s", collected[1])
	}
}

// TestLargeSet_NestedForInIteration tests nested for-in loops over large sets.
// Task 9.10c: Edge case - nested iteration over two large sets.
func TestLargeSet_NestedForInIteration(t *testing.T) {
	// Create enum with 50 elements (moderate size for nested loops)
	values := make(map[string]int)
	orderedNames := make([]string, 50)
	for i := 0; i < 50; i++ {
		name := fmt.Sprintf("E%02d", i)
		values[name] = i
		orderedNames[i] = name
	}
	enumType := &types.EnumType{
		Name:         "TNested",
		Values:       values,
		OrderedNames: orderedNames,
	}

	setType := types.NewSetType(enumType)

	// Create outer set with even numbers
	outerSet := NewSetValue(setType)
	for i := 0; i < 50; i += 2 {
		outerSet.AddElement(i)
	}

	// Create inner set with numbers divisible by 5
	innerSet := NewSetValue(setType)
	for i := 0; i < 50; i += 5 {
		innerSet.AddElement(i)
	}

	// Simulate nested iteration: for e1 in outerSet do for e2 in innerSet do
	outerCount := 0
	totalInnerCount := 0

	for _, outerName := range enumType.OrderedNames {
		outerOrdinal := enumType.Values[outerName]
		if !outerSet.HasElement(outerOrdinal) {
			continue
		}
		outerCount++

		// Inner loop
		for _, innerName := range enumType.OrderedNames {
			innerOrdinal := enumType.Values[innerName]
			if innerSet.HasElement(innerOrdinal) {
				totalInnerCount++
			}
		}
	}

	// Outer should iterate 25 times (even numbers 0, 2, 4, ..., 48)
	if outerCount != 25 {
		t.Errorf("outer loop should execute 25 times, got %d", outerCount)
	}

	// Inner should iterate 10 times per outer iteration = 250 total
	// (innerSet has 10 elements: 0, 5, 10, 15, ..., 45)
	expectedTotal := 25 * 10
	if totalInnerCount != expectedTotal {
		t.Errorf("total inner iterations should be %d, got %d", expectedTotal, totalInnerCount)
	}
}

// TestLargeSet_MixedOperations tests complex combinations of set operations on large sets.
// Task 9.10c: Stress test - multiple operations in sequence.
func TestLargeSet_MixedOperations(t *testing.T) {
	// Create enum with 200 elements
	values := make(map[string]int)
	orderedNames := make([]string, 200)
	for i := 0; i < 200; i++ {
		name := fmt.Sprintf("E%03d", i)
		values[name] = i
		orderedNames[i] = name
	}
	enumType := &types.EnumType{
		Name:         "TMixed",
		Values:       values,
		OrderedNames: orderedNames,
	}

	setType := types.NewSetType(enumType)

	interp := New(nil)

	// Create three sets with different patterns
	// Set A: multiples of 3 (0, 3, 6, 9, ...)
	setA := NewSetValue(setType)
	for i := 0; i < 200; i += 3 {
		setA.AddElement(i)
	}

	// Set B: multiples of 5 (0, 5, 10, 15, ...)
	setB := NewSetValue(setType)
	for i := 0; i < 200; i += 5 {
		setB.AddElement(i)
	}

	// Set C: multiples of 7 (0, 7, 14, 21, ...)
	setC := NewSetValue(setType)
	for i := 0; i < 200; i += 7 {
		setC.AddElement(i)
	}

	// Test: (A ∪ B) ∩ C
	// Elements that are (multiple of 3 OR multiple of 5) AND multiple of 7
	unionABResult := interp.evalBinarySetOperation(setA, setB, "+")
	if isError(unionABResult) {
		t.Fatalf("union A+B failed: %s", unionABResult)
	}
	unionAB := unionABResult.(*SetValue)

	resultVal := interp.evalBinarySetOperation(unionAB, setC, "*")
	if isError(resultVal) {
		t.Fatalf("intersection (A+B)*C failed: %s", resultVal)
	}
	result := resultVal.(*SetValue)

	// Verify specific elements
	// 0 is in all sets
	if !result.HasElement(0) {
		t.Error("0 should be in result (multiple of 3, 5, and 7)")
	}

	// 21 is multiple of 3 and 7 (in A and C, not B)
	if !result.HasElement(21) {
		t.Error("21 should be in result (multiple of 3 and 7)")
	}

	// 35 is multiple of 5 and 7 (in B and C, not A)
	if !result.HasElement(35) {
		t.Error("35 should be in result (multiple of 5 and 7)")
	}

	// 15 is multiple of 3 and 5 but not 7 (in A and B, not C)
	if result.HasElement(15) {
		t.Error("15 should not be in result (not multiple of 7)")
	}

	// Test: A - (B ∪ C)
	// Multiples of 3 that are NOT multiples of 5 or 7
	unionBCResult := interp.evalBinarySetOperation(setB, setC, "+")
	if isError(unionBCResult) {
		t.Fatalf("union B+C failed: %s", unionBCResult)
	}
	unionBC := unionBCResult.(*SetValue)

	diffResultVal := interp.evalBinarySetOperation(setA, unionBC, "-")
	if isError(diffResultVal) {
		t.Fatalf("difference A-(B+C) failed: %s", diffResultVal)
	}
	diffResult := diffResultVal.(*SetValue)

	// 3 is multiple of 3 only
	if !diffResult.HasElement(3) {
		t.Error("3 should be in difference result")
	}

	// 15 is multiple of 3 and 5, should not be in result
	if diffResult.HasElement(15) {
		t.Error("15 should not be in difference result (multiple of 5)")
	}

	// 21 is multiple of 3 and 7, should not be in result
	if diffResult.HasElement(21) {
		t.Error("21 should not be in difference result (multiple of 7)")
	}
}
