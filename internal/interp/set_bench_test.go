// Package interp_test contains benchmarks for set operations.
package interp

import (
	"fmt"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Small Set Benchmarks (Bitmask Storage, ≤64 elements)
// ============================================================================

// BenchmarkSmallSetUnion benchmarks union operation on 64-element sets (bitmask).
func BenchmarkSmallSetUnion(b *testing.B) {
	// Create enum with 64 elements
	enumType := createBenchEnumType("TSmall", 64)
	setType := types.NewSetType(enumType)

	// Create two sets with different elements
	set1 := NewSetValue(setType)
	set2 := NewSetValue(setType)
	for i := 0; i < 64; i += 2 {
		set1.AddElement(i)
	}
	for i := 1; i < 64; i += 2 {
		set2.AddElement(i)
	}

	interp := New(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interp.evalBinarySetOperation(set1, set2, "+")
	}
}

// BenchmarkSmallSetIntersection benchmarks intersection on 64-element sets (bitmask).
func BenchmarkSmallSetIntersection(b *testing.B) {
	enumType := createBenchEnumType("TSmall", 64)
	setType := types.NewSetType(enumType)

	set1 := NewSetValue(setType)
	set2 := NewSetValue(setType)
	for i := 0; i < 64; i += 2 {
		set1.AddElement(i)
	}
	for i := 0; i < 64; i += 3 {
		set2.AddElement(i)
	}

	interp := New(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interp.evalBinarySetOperation(set1, set2, "*")
	}
}

// BenchmarkSmallSetDifference benchmarks difference on 64-element sets (bitmask).
func BenchmarkSmallSetDifference(b *testing.B) {
	enumType := createBenchEnumType("TSmall", 64)
	setType := types.NewSetType(enumType)

	set1 := NewSetValue(setType)
	set2 := NewSetValue(setType)
	for i := 0; i < 64; i += 2 {
		set1.AddElement(i)
	}
	for i := 0; i < 64; i += 4 {
		set2.AddElement(i)
	}

	interp := New(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interp.evalBinarySetOperation(set1, set2, "-")
	}
}

// BenchmarkSmallSetMembership benchmarks membership testing on 64-element sets.
func BenchmarkSmallSetMembership(b *testing.B) {
	enumType := createBenchEnumType("TSmall", 64)
	setType := types.NewSetType(enumType)

	set := NewSetValue(setType)
	for i := 0; i < 64; i += 2 {
		set.AddElement(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test both present and absent elements
		_ = set.HasElement(10)
		_ = set.HasElement(11)
		_ = set.HasElement(50)
		_ = set.HasElement(51)
	}
}

// BenchmarkSmallSetForIn benchmarks for-in iteration over 64-element set.
func BenchmarkSmallSetForIn(b *testing.B) {
	enumType := createBenchEnumType("TSmall", 64)
	setType := types.NewSetType(enumType)

	set := NewSetValue(setType)
	for i := 0; i < 64; i += 2 {
		set.AddElement(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate for-in iteration
		count := 0
		for _, name := range enumType.OrderedNames {
			ordinal := enumType.Values[name]
			if set.HasElement(ordinal) {
				count++
			}
		}
		if count == 0 {
			b.Fatal("iteration found no elements")
		}
	}
}

// ============================================================================
// Large Set Benchmarks (Map Storage, >64 elements)
// ============================================================================

// BenchmarkLargeSetUnion100 benchmarks union on 100-element sets (map storage).
func BenchmarkLargeSetUnion100(b *testing.B) {
	benchmarkLargeSetUnion(b, 100)
}

// BenchmarkLargeSetUnion200 benchmarks union on 200-element sets (map storage).
func BenchmarkLargeSetUnion200(b *testing.B) {
	benchmarkLargeSetUnion(b, 200)
}

// BenchmarkLargeSetUnion500 benchmarks union on 500-element sets (map storage).
func BenchmarkLargeSetUnion500(b *testing.B) {
	benchmarkLargeSetUnion(b, 500)
}

func benchmarkLargeSetUnion(b *testing.B, size int) {
	enumType := createBenchEnumType("TLarge", size)
	setType := types.NewSetType(enumType)

	set1 := NewSetValue(setType)
	set2 := NewSetValue(setType)
	for i := 0; i < size; i += 2 {
		set1.AddElement(i)
	}
	for i := 1; i < size; i += 2 {
		set2.AddElement(i)
	}

	interp := New(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interp.evalBinarySetOperation(set1, set2, "+")
	}
}

// BenchmarkLargeSetIntersection100 benchmarks intersection on 100-element sets.
func BenchmarkLargeSetIntersection100(b *testing.B) {
	benchmarkLargeSetIntersection(b, 100)
}

// BenchmarkLargeSetIntersection200 benchmarks intersection on 200-element sets.
func BenchmarkLargeSetIntersection200(b *testing.B) {
	benchmarkLargeSetIntersection(b, 200)
}

// BenchmarkLargeSetIntersection500 benchmarks intersection on 500-element sets.
func BenchmarkLargeSetIntersection500(b *testing.B) {
	benchmarkLargeSetIntersection(b, 500)
}

func benchmarkLargeSetIntersection(b *testing.B, size int) {
	enumType := createBenchEnumType("TLarge", size)
	setType := types.NewSetType(enumType)

	set1 := NewSetValue(setType)
	set2 := NewSetValue(setType)
	for i := 0; i < size; i += 2 {
		set1.AddElement(i)
	}
	for i := 0; i < size; i += 3 {
		set2.AddElement(i)
	}

	interp := New(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interp.evalBinarySetOperation(set1, set2, "*")
	}
}

// BenchmarkLargeSetDifference100 benchmarks difference on 100-element sets.
func BenchmarkLargeSetDifference100(b *testing.B) {
	benchmarkLargeSetDifference(b, 100)
}

// BenchmarkLargeSetDifference200 benchmarks difference on 200-element sets.
func BenchmarkLargeSetDifference200(b *testing.B) {
	benchmarkLargeSetDifference(b, 200)
}

// BenchmarkLargeSetDifference500 benchmarks difference on 500-element sets.
func BenchmarkLargeSetDifference500(b *testing.B) {
	benchmarkLargeSetDifference(b, 500)
}

func benchmarkLargeSetDifference(b *testing.B, size int) {
	enumType := createBenchEnumType("TLarge", size)
	setType := types.NewSetType(enumType)

	set1 := NewSetValue(setType)
	set2 := NewSetValue(setType)
	for i := 0; i < size; i += 2 {
		set1.AddElement(i)
	}
	for i := 0; i < size; i += 4 {
		set2.AddElement(i)
	}

	interp := New(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interp.evalBinarySetOperation(set1, set2, "-")
	}
}

// BenchmarkLargeSetMembership100 benchmarks membership on 100-element sets.
func BenchmarkLargeSetMembership100(b *testing.B) {
	benchmarkLargeSetMembership(b, 100)
}

// BenchmarkLargeSetMembership200 benchmarks membership on 200-element sets.
func BenchmarkLargeSetMembership200(b *testing.B) {
	benchmarkLargeSetMembership(b, 200)
}

// BenchmarkLargeSetMembership500 benchmarks membership on 500-element sets.
func BenchmarkLargeSetMembership500(b *testing.B) {
	benchmarkLargeSetMembership(b, 500)
}

func benchmarkLargeSetMembership(b *testing.B, size int) {
	enumType := createBenchEnumType("TLarge", size)
	setType := types.NewSetType(enumType)

	set := NewSetValue(setType)
	for i := 0; i < size; i += 2 {
		set.AddElement(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test various positions
		_ = set.HasElement(10)
		_ = set.HasElement(11)
		_ = set.HasElement(size / 2)
		_ = set.HasElement(size/2 + 1)
		_ = set.HasElement(size - 2)
		_ = set.HasElement(size - 1)
	}
}

// BenchmarkLargeSetForIn100 benchmarks for-in iteration over 100-element set.
func BenchmarkLargeSetForIn100(b *testing.B) {
	benchmarkLargeSetForIn(b, 100)
}

// BenchmarkLargeSetForIn200 benchmarks for-in iteration over 200-element set.
func BenchmarkLargeSetForIn200(b *testing.B) {
	benchmarkLargeSetForIn(b, 200)
}

// BenchmarkLargeSetForIn500 benchmarks for-in iteration over 500-element set.
func BenchmarkLargeSetForIn500(b *testing.B) {
	benchmarkLargeSetForIn(b, 500)
}

func benchmarkLargeSetForIn(b *testing.B, size int) {
	enumType := createBenchEnumType("TLarge", size)
	setType := types.NewSetType(enumType)

	set := NewSetValue(setType)
	for i := 0; i < size; i += 2 {
		set.AddElement(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate for-in iteration
		count := 0
		for _, name := range enumType.OrderedNames {
			ordinal := enumType.Values[name]
			if set.HasElement(ordinal) {
				count++
			}
		}
		if count == 0 {
			b.Fatal("iteration found no elements")
		}
	}
}

// ============================================================================
// Set Literal Creation Benchmarks
// ============================================================================

// BenchmarkSetLiteralCreation_Small benchmarks creating small set literals.
func BenchmarkSetLiteralCreation_Small(b *testing.B) {
	enumType := createBenchEnumType("TSmall", 64)
	setType := types.NewSetType(enumType)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set := NewSetValue(setType)
		for j := 0; j < 10; j++ {
			set.AddElement(j * 2)
		}
	}
}

// BenchmarkSetLiteralCreation_Large100 benchmarks creating 100-element set literals.
func BenchmarkSetLiteralCreation_Large100(b *testing.B) {
	benchmarkSetLiteralCreation(b, 100, 20)
}

// BenchmarkSetLiteralCreation_Large200 benchmarks creating 200-element set literals.
func BenchmarkSetLiteralCreation_Large200(b *testing.B) {
	benchmarkSetLiteralCreation(b, 200, 20)
}

// BenchmarkSetLiteralCreation_Large500 benchmarks creating 500-element set literals.
func BenchmarkSetLiteralCreation_Large500(b *testing.B) {
	benchmarkSetLiteralCreation(b, 500, 20)
}

func benchmarkSetLiteralCreation(b *testing.B, enumSize, literalSize int) {
	enumType := createBenchEnumType("TLarge", enumSize)
	setType := types.NewSetType(enumType)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set := NewSetValue(setType)
		for j := 0; j < literalSize; j++ {
			set.AddElement(j * (enumSize / literalSize))
		}
	}
}

// ============================================================================
// Memory Allocation Benchmarks
// ============================================================================

// BenchmarkSetAllocation_Small benchmarks memory allocation for small sets.
func BenchmarkSetAllocation_Small(b *testing.B) {
	enumType := createBenchEnumType("TSmall", 64)
	setType := types.NewSetType(enumType)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewSetValue(setType)
	}
}

// BenchmarkSetAllocation_Large100 benchmarks memory allocation for 100-element sets.
func BenchmarkSetAllocation_Large100(b *testing.B) {
	enumType := createBenchEnumType("TLarge", 100)
	setType := types.NewSetType(enumType)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewSetValue(setType)
	}
}

// BenchmarkSetAllocation_Large500 benchmarks memory allocation for 500-element sets.
func BenchmarkSetAllocation_Large500(b *testing.B) {
	enumType := createBenchEnumType("TLarge", 500)
	setType := types.NewSetType(enumType)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewSetValue(setType)
	}
}

// ============================================================================
// Benchmark Helper Functions
// ============================================================================

// createBenchEnumType creates an EnumType for benchmarking with specified size.
func createBenchEnumType(name string, size int) *types.EnumType {
	values := make(map[string]int, size)
	orderedNames := make([]string, size)

	for i := 0; i < size; i++ {
		elementName := fmt.Sprintf("E%04d", i)
		values[elementName] = i
		orderedNames[i] = elementName
	}

	return &types.EnumType{
		Name:         name,
		Values:       values,
		OrderedNames: orderedNames,
	}
}
