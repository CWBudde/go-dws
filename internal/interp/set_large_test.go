package interp

import (
	"fmt"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================ //
// Large set storage and operations
// ============================================================================ //

// TestSetValue_BoundaryCase64Elements tests set with exactly 64 elements (boundary case).
// This should use bitmask storage (threshold is ≤64).
func TestSetValue_BoundaryCase64Elements(t *testing.T) {
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

	if setType.StorageKind != types.SetStorageBitmask {
		t.Errorf("expected bitmask storage for 64-element enum, got %s", setType.StorageKind)
	}

	set := NewSetValue(setType)
	set.AddElement(0)
	set.AddElement(31)
	set.AddElement(63)

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

	if str := set.String(); str == "[]" {
		t.Error("expected non-empty string representation")
	}
}

// TestSetValue_LargeEnum65Elements tests set with 65 elements (triggers map storage).
func TestSetValue_LargeEnum65Elements(t *testing.T) {
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

	if setType.StorageKind != types.SetStorageMap {
		t.Errorf("expected map storage for 65-element enum, got %s", setType.StorageKind)
	}

	set := NewSetValue(setType)
	if set.MapStore == nil {
		t.Fatal("expected MapStore to be initialized for large enum")
	}

	set.AddElement(0)
	set.AddElement(32)
	set.AddElement(64)

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
	values := make(map[string]int)
	orderedNames := make([]string, 100)
	for i := 0; i < 100; i++ {
		name := "E" + string(rune('0'+i/10)) + string(rune('0'+i%10))
		values[name] = i
		orderedNames[i] = name
	}

	enumType := types.NewEnumType("TLarge100", values, orderedNames)
	setType := types.NewSetType(enumType)

	if setType.StorageKind != types.SetStorageMap {
		t.Errorf("expected map storage for 100-element enum, got %s", setType.StorageKind)
	}

	set := NewSetValue(setType)
	for i := 0; i < 100; i += 5 {
		set.AddElement(i)
	}

	count := 0
	for i := 0; i < 100; i++ {
		if set.HasElement(i) {
			count++
		}
	}
	if count != 20 {
		t.Errorf("expected 20 elements, got %d", count)
	}

	if str := set.String(); str == "[]" {
		t.Error("expected non-empty string representation")
	}
}

// TestLargeSet_BinaryOperations tests set operations on large sets.
func TestLargeSet_BinaryOperations(t *testing.T) {
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

	set1 := NewSetValue(setType)
	set2 := NewSetValue(setType)

	for i := 0; i < 80; i += 10 {
		set1.AddElement(i)
	}
	for i := 0; i < 80; i += 20 {
		set2.AddElement(i)
	}

	interp := New(nil)

	unionResult := interp.evalBinarySetOperation(set1, set2, "+")
	if isError(unionResult) {
		t.Fatalf("union operation failed: %s", unionResult)
	}
	unionSet := unionResult.(*SetValue)
	expectedUnion := []int{0, 10, 20, 30, 40, 50, 60, 70}
	for _, ord := range expectedUnion {
		if !unionSet.HasElement(ord) {
			t.Errorf("union: expected element %d to be present", ord)
		}
	}

	intersectResult := interp.evalBinarySetOperation(set1, set2, "*")
	if isError(intersectResult) {
		t.Fatalf("intersection operation failed: %s", intersectResult)
	}
	intersectSet := intersectResult.(*SetValue)
	expectedIntersect := []int{0, 20, 40, 60}
	for _, ord := range expectedIntersect {
		if !intersectSet.HasElement(ord) {
			t.Errorf("intersection: expected element %d to be present", ord)
		}
	}
	for _, ord := range []int{10, 30, 50, 70} {
		if intersectSet.HasElement(ord) {
			t.Errorf("intersection: expected element %d to be absent", ord)
		}
	}

	diffResult := interp.evalBinarySetOperation(set1, set2, "-")
	if isError(diffResult) {
		t.Fatalf("difference operation failed: %s", diffResult)
	}
	diffSet := diffResult.(*SetValue)
	expectedDiff := []int{10, 30, 50, 70}
	for _, ord := range expectedDiff {
		if !diffSet.HasElement(ord) {
			t.Errorf("difference: expected element %d to be present", ord)
		}
	}
	for _, ord := range []int{0, 20, 40, 60} {
		if diffSet.HasElement(ord) {
			t.Errorf("difference: expected element %d to be absent", ord)
		}
	}
}

func TestLargeSet_Performance(t *testing.T) {
	values := make(map[string]int)
	orderedNames := make([]string, 200)
	for i := 0; i < 200; i++ {
		name := "E" + string(rune('0'+i/100)) + string(rune('0'+(i/10)%10)) + string(rune('0'+i%10))
		values[name] = i
		orderedNames[i] = name
	}

	enumType := types.NewEnumType("TLarge200", values, orderedNames)
	setType := types.NewSetType(enumType)

	set1 := NewSetValue(setType)
	set2 := NewSetValue(setType)
	for i := 0; i < 200; i += 2 {
		set1.AddElement(i)
	}
	for i := 1; i < 200; i += 2 {
		set2.AddElement(i)
	}

	interp := New(nil)
	if isError(interp.evalBinarySetOperation(set1, set2, "+")) {
		t.Fatalf("union failed")
	}
	if isError(interp.evalBinarySetOperation(set1, set2, "*")) {
		t.Fatalf("intersection failed")
	}
	if isError(interp.evalBinarySetOperation(set1, set2, "-")) {
		t.Fatalf("difference failed")
	}
}

func TestVeryLargeSet_500Elements(t *testing.T) {
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

	set1 := NewSetValue(setType)
	for i := 0; i < 500; i += 10 {
		set1.AddElement(i)
	}

	set2 := NewSetValue(setType)
	for i := 0; i < 500; i += 5 {
		set2.AddElement(i)
	}

	interp := New(nil)

	unionResult := interp.evalBinarySetOperation(set1, set2, "+")
	if isError(unionResult) {
		t.Fatalf("union operation failed: %s", unionResult)
	}
	union := unionResult.(*SetValue)
	if !union.HasElement(0) || !union.HasElement(5) || !union.HasElement(10) {
		t.Error("Union missing expected elements")
	}

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
	if intersection.HasElement(5) {
		t.Error("Intersection should not contain element 5")
	}

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

func TestVeryLargeSet_1000Elements(t *testing.T) {
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

	set1 := NewSetValue(setType)
	for i := 0; i < 100; i++ {
		set1.AddElement(i)
	}

	set2 := NewSetValue(setType)
	for i := 900; i < 1000; i++ {
		set2.AddElement(i)
	}

	interp := New(nil)

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

func TestLargeSet_MixedOperations(t *testing.T) {
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

	setA := NewSetValue(setType)
	for i := 0; i < 200; i += 3 {
		setA.AddElement(i)
	}

	setB := NewSetValue(setType)
	for i := 0; i < 200; i += 5 {
		setB.AddElement(i)
	}

	setC := NewSetValue(setType)
	for i := 0; i < 200; i += 7 {
		setC.AddElement(i)
	}

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

	if !result.HasElement(0) {
		t.Error("0 should be in result (multiple of 3, 5, and 7)")
	}
	if !result.HasElement(21) {
		t.Error("21 should be in result (multiple of 3 and 7)")
	}
	if !result.HasElement(35) {
		t.Error("35 should be in result (multiple of 5 and 7)")
	}
	if result.HasElement(15) {
		t.Error("15 should not be in result (not multiple of 7)")
	}

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

	if !diffResult.HasElement(3) {
		t.Error("3 should be in difference result")
	}
	if diffResult.HasElement(15) {
		t.Error("15 should not be in difference result (multiple of 5)")
	}
	if diffResult.HasElement(21) {
		t.Error("21 should not be in difference result (multiple of 7)")
	}
}
