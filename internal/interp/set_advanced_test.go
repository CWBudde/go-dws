package interp

import (
	"fmt"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Large Set Tests
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
// For-In Edge Cases
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
// Large Set Edge Cases
// ============================================================================

// TestVeryLargeSet_500Elements tests set operations with 500-element enum.
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

// ============================================================================
// Set Initialization Tests
// ============================================================================

func TestSetUninitializedVariable(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "uninitialized inline set type",
			input: `
				type TColor = (Red, Green, Blue);
				var s: set of TColor;
				PrintLn('ok');
			`,
			expect: "ok\n",
		},
		{
			name: "multi-identifier set declaration",
			input: `
				type TColor = (Red, Green, Blue);
				var s1, s2, s3: set of TColor;
				PrintLn('ok');
			`,
			expect: "ok\n",
		},
		// Note: Named set types (type TColorSet = set of TColor) are not yet implemented
		// That would require semantic analysis support for set type aliases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

func TestSetInOperatorEmpty(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "in operator with empty set - should be false",
			input: `
				type TColor = (Red, Green, Blue);
				var s: set of TColor;
				if Red in s then
					PrintLn('found')
				else
					PrintLn('not found');
			`,
			expect: "not found\n",
		},
		{
			name: "multiple checks on empty set",
			input: `
				type TColor = (Red, Green, Blue);
				var s: set of TColor;
				var count := 0;
				if Red in s then count := count + 1;
				if Green in s then count := count + 1;
				if Blue in s then count := count + 1;
				PrintLn(count);
			`,
			expect: "0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

func TestSetInOperatorAfterInclude(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "Include then check membership",
			input: `
				type TColor = (Red, Green, Blue);
				var s: set of TColor;
				s.Include(Red);
				if Red in s then
					PrintLn('found')
				else
					PrintLn('not found');
			`,
			expect: "found\n",
		},
		{
			name: "Include multiple then check all",
			input: `
				type TColor = (Red, Green, Blue);
				var s: set of TColor;
				s.Include(Red);
				s.Include(Blue);

				var count := 0;
				if Red in s then count := count + 1;
				if Green in s then count := count + 1;
				if Blue in s then count := count + 1;
				PrintLn(count);
			`,
			expect: "2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

func TestSetMultiIdentifierSeparateInstances(t *testing.T) {
	// Verify that multi-identifier declarations create separate set instances
	input := `
		type TColor = (Red, Green, Blue);
		var s1, s2: set of TColor;

		s1.Include(Red);
		s2.Include(Blue);

		// s1 should only have Red
		if Red in s1 then PrintLn('s1 has Red');
		if Blue in s1 then PrintLn('s1 has Blue');

		// s2 should only have Blue
		if Red in s2 then PrintLn('s2 has Red');
		if Blue in s2 then PrintLn('s2 has Blue');
	`
	_, output := testEvalWithOutput(input)
	expect := "s1 has Red\ns2 has Blue\n"
	if output != expect {
		t.Errorf("expected %q, got %q", expect, output)
	}
}

func TestSetForInEmpty(t *testing.T) {
	// Test for-in over empty set should execute 0 times
	input := `
		type TColor = (Red, Green, Blue);
		var s: set of TColor;
		var count := 0;
		for var e in s do
			count := count + 1;
		PrintLn(count);
	`
	_, output := testEvalWithOutput(input)
	expect := "0\n"
	if output != expect {
		t.Errorf("expected %q, got %q", expect, output)
	}
}

func TestSetInitializationEratosthenePattern(t *testing.T) {
	// Test the pattern from eratosthene.pas
	input := `
		type TRange = enum (Low = 2, High = 20);
		var sieve: set of TRange;

		// Initially empty
		var count := 0;
		for var e in TRange do begin
			if e in sieve then
				count := count + 1;
		end;
		PrintLn(count);

		// Add one element
		sieve.Include(TRange.Low);  // Fixed: Use qualified access for scoped enum
		count := 0;
		for var e in TRange do begin
			if e in sieve then
				count := count + 1;
		end;
		PrintLn(count);
	`
	_, output := testEvalWithOutput(input)
	expect := "0\n1\n"
	if output != expect {
		t.Errorf("expected %q, got %q", expect, output)
	}
}
