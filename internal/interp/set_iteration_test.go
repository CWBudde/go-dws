package interp

import (
	"fmt"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================ //
// Set iteration tests (for-in and membership-driven enumeration)
// ============================================================================ //

func TestLargeSet_ForInIteration(t *testing.T) {
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

	if setType.StorageKind != types.SetStorageMap {
		t.Errorf("expected map storage for 70-element enum, got %s", setType.StorageKind)
	}

	set := NewSetValue(setType)
	expectedOrdinals := []int{5, 15, 25, 35, 45, 55, 65}
	for _, ord := range expectedOrdinals {
		set.AddElement(ord)
	}

	for _, ord := range expectedOrdinals {
		if !set.HasElement(ord) {
			t.Errorf("expected element %d to be in set", ord)
		}
	}

	collectedOrdinals := []int{}
	for _, name := range enumType.OrderedNames {
		ordinal := enumType.Values[name]
		if set.HasElement(ordinal) {
			collectedOrdinals = append(collectedOrdinals, ordinal)
		}
	}

	if len(collectedOrdinals) != len(expectedOrdinals) {
		t.Errorf("expected %d iterations, got %d", len(expectedOrdinals), len(collectedOrdinals))
	}
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

func TestForInSet_EmptySet(t *testing.T) {
	enumType := types.NewEnumType("TColor", map[string]int{
		"Red":   0,
		"Green": 1,
		"Blue":  2,
	}, []string{"Red", "Green", "Blue"})
	setType := types.NewSetType(enumType)
	emptySet := NewSetValue(setType)

	executionCount := 0
	for _, name := range enumType.OrderedNames {
		ordinal := enumType.Values[name]
		if emptySet.HasElement(ordinal) {
			executionCount++
		}
	}

	if executionCount != 0 {
		t.Errorf("expected loop body to never execute for empty set, executed %d times", executionCount)
	}
}

func TestForInSet_SingleElement(t *testing.T) {
	enumType := types.NewEnumType("TColor", map[string]int{
		"Red":   0,
		"Green": 1,
		"Blue":  2,
	}, []string{"Red", "Green", "Blue"})
	setType := types.NewSetType(enumType)
	singleSet := NewSetValue(setType)
	singleSet.AddElement(1) // Green

	executionCount := 0
	var lastElement string
	for _, name := range enumType.OrderedNames {
		ordinal := enumType.Values[name]
		if singleSet.HasElement(ordinal) {
			executionCount++
			lastElement = name
		}
	}

	if executionCount != 1 {
		t.Errorf("expected loop body to execute exactly once, executed %d times", executionCount)
	}
	if lastElement != "Green" {
		t.Errorf("expected to iterate over Green, got %s", lastElement)
	}
}

func TestLargeSet_ForInAllElements(t *testing.T) {
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
	fullSet := NewSetValue(setType)
	for i := 0; i < 100; i++ {
		fullSet.AddElement(i)
	}

	count := 0
	for _, name := range enumType.OrderedNames {
		ordinal := enumType.Values[name]
		if fullSet.HasElement(ordinal) {
			count++
		}
	}

	if count != 100 {
		t.Errorf("expected to iterate over all 100 elements, got %d", count)
	}
}

func TestLargeSet_ForInFirstAndLastOnly(t *testing.T) {
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
	boundarySet := NewSetValue(setType)
	boundarySet.AddElement(0)
	boundarySet.AddElement(99)

	var collected []string
	for _, name := range enumType.OrderedNames {
		ordinal := enumType.Values[name]
		if boundarySet.HasElement(ordinal) {
			collected = append(collected, name)
		}
	}

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

func TestLargeSet_NestedForInIteration(t *testing.T) {
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

	outerSet := NewSetValue(setType)
	for i := 0; i < 50; i += 2 {
		outerSet.AddElement(i)
	}

	innerSet := NewSetValue(setType)
	for i := 0; i < 50; i += 5 {
		innerSet.AddElement(i)
	}

	outerCount := 0
	totalInnerCount := 0

	for _, outerName := range enumType.OrderedNames {
		outerOrdinal := enumType.Values[outerName]
		if !outerSet.HasElement(outerOrdinal) {
			continue
		}
		outerCount++

		for _, innerName := range enumType.OrderedNames {
			innerOrdinal := enumType.Values[innerName]
			if innerSet.HasElement(innerOrdinal) {
				totalInnerCount++
			}
		}
	}

	if outerCount != 25 {
		t.Errorf("outer loop should execute 25 times, got %d", outerCount)
	}

	expectedTotal := 25 * 10
	if totalInnerCount != expectedTotal {
		t.Errorf("total inner iterations should be %d, got %d", expectedTotal, totalInnerCount)
	}
}
