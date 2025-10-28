package interp

import (
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
// Set Literal Evaluation Tests (Tasks 8.106-8.107)
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
// Set Binary Operations Tests (Tasks 8.110-8.112)
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
// Include/Exclude Methods (Tasks 8.108-8.109)
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
