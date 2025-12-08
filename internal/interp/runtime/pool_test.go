package runtime

import (
	"testing"
)

func TestIntegerPool(t *testing.T) {
	ResetPoolStats()

	// Create and release values
	v1 := NewInteger(42)
	if v1.Value != 42 {
		t.Errorf("Expected value 42, got %d", v1.Value)
	}
	ReleaseInteger(v1)

	v2 := NewInteger(100)
	if v2.Value != 100 {
		t.Errorf("Expected value 100, got %d", v2.Value)
	}

	// v2 should be the same instance as v1 (reused from pool)
	if v2 != v1 {
		t.Log("Note: Pool may not reuse on first get (this is OK)")
	}

	stats := GetPoolStats()
	if stats.IntegerGets != 2 {
		t.Errorf("Expected 2 gets, got %d", stats.IntegerGets)
	}
	if stats.IntegerPuts != 1 {
		t.Errorf("Expected 1 put, got %d", stats.IntegerPuts)
	}
}

func TestFloatPool(t *testing.T) {
	ResetPoolStats()

	v1 := NewFloat(3.14)
	if v1.Value != 3.14 {
		t.Errorf("Expected value 3.14, got %f", v1.Value)
	}
	ReleaseFloat(v1)

	v2 := NewFloat(2.71)
	if v2.Value != 2.71 {
		t.Errorf("Expected value 2.71, got %f", v2.Value)
	}

	stats := GetPoolStats()
	if stats.FloatGets != 2 {
		t.Errorf("Expected 2 gets, got %d", stats.FloatGets)
	}
}

func TestBooleanSingletons(t *testing.T) {
	// Booleans should always return the same instances
	t1 := NewBoolean(true)
	t2 := NewBoolean(true)
	if t1 != t2 {
		t.Error("Expected true booleans to be the same instance")
	}

	f1 := NewBoolean(false)
	f2 := NewBoolean(false)
	if f1 != f2 {
		t.Error("Expected false booleans to be the same instance")
	}

	if t1 == f1 {
		t.Error("Expected true and false to be different instances")
	}
}

func TestStringCreation(t *testing.T) {
	s1 := NewString("hello")
	s2 := NewString("hello")

	// Strings are not pooled, so these should be different instances
	if s1 == s2 {
		t.Error("Expected different string instances")
	}

	if s1.Value != "hello" || s2.Value != "hello" {
		t.Error("String values should be 'hello'")
	}
}

func TestPoolStats(t *testing.T) {
	ResetPoolStats()

	// Create some values
	for i := 0; i < 10; i++ {
		v := NewInteger(int64(i))
		if i%2 == 0 {
			ReleaseInteger(v) // Release half of them
		}
	}

	stats := GetPoolStats()
	if stats.IntegerGets != 10 {
		t.Errorf("Expected 10 gets, got %d", stats.IntegerGets)
	}
	if stats.IntegerPuts != 5 {
		t.Errorf("Expected 5 puts, got %d", stats.IntegerPuts)
	}

	// Check efficiency calculation
	intEff, _ := stats.PoolEfficiency()
	if intEff < 0 || intEff > 100 {
		t.Errorf("Pool efficiency should be between 0-100, got %f", intEff)
	}
}

func TestPoolNilSafety(t *testing.T) {
	// Releasing nil should not panic
	ReleaseInteger(nil)
	ReleaseFloat(nil)
	ReleaseBoolean(nil)
}

func TestValueInterfaces(t *testing.T) {
	// Test that pooled values implement the correct interfaces
	var _ Value = NewInteger(42)
	var _ NumericValue = NewInteger(42)
	var _ ComparableValue = NewInteger(42)
	var _ OrderableValue = NewInteger(42)
	var _ CopyableValue = NewInteger(42)

	var _ Value = NewFloat(3.14)
	var _ NumericValue = NewFloat(3.14)
	var _ ComparableValue = NewFloat(3.14)
	var _ OrderableValue = NewFloat(3.14)

	var _ Value = NewBoolean(true)
	var _ ComparableValue = NewBoolean(true)

	var _ Value = NewString("test")
	var _ ComparableValue = NewString("test")
	var _ OrderableValue = NewString("test")
	var _ IndexableValue = NewString("test")
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkIntegerPooled(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := NewInteger(42)
		ReleaseInteger(v)
	}
}

func BenchmarkIntegerDirect(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := &IntegerValue{Value: 42}
		_ = v
	}
}

func BenchmarkFloatPooled(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := NewFloat(3.14)
		ReleaseFloat(v)
	}
}

func BenchmarkFloatDirect(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := &FloatValue{Value: 3.14}
		_ = v
	}
}

func BenchmarkBooleanPooled(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := NewBoolean(true)
		_ = v
	}
}

func BenchmarkBooleanDirect(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := &BooleanValue{Value: true}
		_ = v
	}
}

func BenchmarkNumericInterfaces(b *testing.B) {
	v := NewInteger(42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if val, ok := v.AsInteger(); !ok || val != 42 {
			b.Fatal("AsInteger failed")
		}
		if val, ok := v.AsFloat(); !ok || val != 42.0 {
			b.Fatal("AsFloat failed")
		}
	}
}

func BenchmarkComparableInterface(b *testing.B) {
	v1 := NewInteger(42)
	v2 := NewInteger(42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if eq, err := v1.Equals(v2); err != nil || !eq {
			b.Fatal("Equals failed")
		}
	}
}

func BenchmarkOrderableInterface(b *testing.B) {
	v1 := NewInteger(42)
	v2 := NewInteger(100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if cmp, err := v1.CompareTo(v2); err != nil || cmp != -1 {
			b.Fatal("CompareTo failed")
		}
	}
}
