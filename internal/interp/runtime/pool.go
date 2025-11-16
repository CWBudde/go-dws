package runtime

import (
	"sync"
	"sync/atomic"
)

// ============================================================================
// Value Object Pooling
// ============================================================================
//
// Phase 3.2.3: Implement object pooling for frequently allocated value types
// to reduce garbage collection pressure and improve performance.
//
// Pooled types: IntegerValue, FloatValue, BooleanValue (most commonly allocated)
// Not pooled: StringValue (variable size), complex types (less frequent)
//
// Usage:
//   val := NewInteger(42)  // Gets from pool if available
//   ... use val ...
//   ReleaseInteger(val)    // Returns to pool for reuse
//
// The Release functions are optional - values will be garbage collected normally
// if not explicitly released. Pools are primarily beneficial in tight loops.
// ============================================================================

var (
	// Object pools for primitive value types
	integerPool = sync.Pool{
		New: func() interface{} {
			poolStats.integerAllocs.Add(1)
			return &IntegerValue{}
		},
	}

	floatPool = sync.Pool{
		New: func() interface{} {
			poolStats.floatAllocs.Add(1)
			return &FloatValue{}
		},
	}

	booleanPool = sync.Pool{
		New: func() interface{} {
			poolStats.booleanAllocs.Add(1)
			return &BooleanValue{}
		},
	}

	// Pool statistics for monitoring
	poolStats = struct {
		integerAllocs atomic.Uint64
		integerGets   atomic.Uint64
		integerPuts   atomic.Uint64

		floatAllocs atomic.Uint64
		floatGets   atomic.Uint64
		floatPuts   atomic.Uint64

		booleanAllocs atomic.Uint64
		booleanGets   atomic.Uint64
		booleanPuts   atomic.Uint64
	}{}
)

// ============================================================================
// Integer Value Pooling
// ============================================================================

// NewInteger creates a new IntegerValue, potentially reusing a pooled instance.
// This is more efficient than &IntegerValue{Value: v} for frequently allocated values.
func NewInteger(value int64) *IntegerValue {
	poolStats.integerGets.Add(1)
	v := integerPool.Get().(*IntegerValue)
	v.Value = value
	return v
}

// ReleaseInteger returns an IntegerValue to the pool for reuse.
// This is optional - if not called, the value will be garbage collected normally.
// Only call this when you're certain the value is no longer needed.
func ReleaseInteger(v *IntegerValue) {
	if v != nil {
		v.Value = 0 // Clear for safety
		poolStats.integerPuts.Add(1)
		integerPool.Put(v)
	}
}

// ============================================================================
// Float Value Pooling
// ============================================================================

// NewFloat creates a new FloatValue, potentially reusing a pooled instance.
func NewFloat(value float64) *FloatValue {
	poolStats.floatGets.Add(1)
	v := floatPool.Get().(*FloatValue)
	v.Value = value
	return v
}

// ReleaseFloat returns a FloatValue to the pool for reuse.
func ReleaseFloat(v *FloatValue) {
	if v != nil {
		v.Value = 0.0 // Clear for safety
		poolStats.floatPuts.Add(1)
		floatPool.Put(v)
	}
}

// ============================================================================
// Boolean Value Pooling
// ============================================================================

var (
	// Pre-allocated singleton boolean values for common cases.
	// Most boolean values are True or False, so we can reuse these.
	trueValue  = &BooleanValue{Value: true}
	falseValue = &BooleanValue{Value: false}
)

// NewBoolean creates a new BooleanValue.
// For true/false, returns singleton instances.
// This avoids allocations for the most common cases.
func NewBoolean(value bool) *BooleanValue {
	if value {
		return trueValue
	}
	return falseValue
}

// ReleaseBoolean is a no-op for booleans since we use singletons.
// Provided for API consistency.
func ReleaseBoolean(v *BooleanValue) {
	// No-op: booleans use singletons
}

// ============================================================================
// String Value Creation (no pooling - variable size)
// ============================================================================

// NewString creates a new StringValue.
// Strings are not pooled due to variable size.
func NewString(value string) *StringValue {
	return &StringValue{Value: value}
}

// ============================================================================
// Pool Statistics
// ============================================================================

// PoolStats holds statistics about value pool usage.
type PoolStats struct {
	IntegerAllocs uint64 // Total allocations (pool misses)
	IntegerGets   uint64 // Total gets from pool
	IntegerPuts   uint64 // Total returns to pool

	FloatAllocs uint64
	FloatGets   uint64
	FloatPuts   uint64

	BooleanAllocs uint64
	BooleanGets   uint64
	BooleanPuts   uint64
}

// GetPoolStats returns current pool statistics.
// Useful for monitoring and debugging pool effectiveness.
func GetPoolStats() PoolStats {
	return PoolStats{
		IntegerAllocs: poolStats.integerAllocs.Load(),
		IntegerGets:   poolStats.integerGets.Load(),
		IntegerPuts:   poolStats.integerPuts.Load(),

		FloatAllocs: poolStats.floatAllocs.Load(),
		FloatGets:   poolStats.floatGets.Load(),
		FloatPuts:   poolStats.floatPuts.Load(),

		BooleanAllocs: poolStats.booleanAllocs.Load(),
		BooleanGets:   poolStats.booleanGets.Load(),
		BooleanPuts:   poolStats.booleanPuts.Load(),
	}
}

// ResetPoolStats resets pool statistics to zero.
// Useful for benchmarking and testing.
func ResetPoolStats() {
	poolStats.integerAllocs.Store(0)
	poolStats.integerGets.Store(0)
	poolStats.integerPuts.Store(0)

	poolStats.floatAllocs.Store(0)
	poolStats.floatGets.Store(0)
	poolStats.floatPuts.Store(0)

	poolStats.booleanAllocs.Store(0)
	poolStats.booleanGets.Store(0)
	poolStats.booleanPuts.Store(0)
}

// PoolEfficiency returns the pool hit rate as a percentage (0-100).
// A higher percentage means the pool is more effective at reusing values.
// Formula: (Gets - Allocs) / Gets * 100
func (s PoolStats) PoolEfficiency() (integer, float, boolean float64) {
	intEff := 0.0
	if s.IntegerGets > 0 {
		intEff = float64(s.IntegerGets-s.IntegerAllocs) / float64(s.IntegerGets) * 100
	}

	floatEff := 0.0
	if s.FloatGets > 0 {
		floatEff = float64(s.FloatGets-s.FloatAllocs) / float64(s.FloatGets) * 100
	}

	boolEff := 0.0
	if s.BooleanGets > 0 {
		boolEff = float64(s.BooleanGets-s.BooleanAllocs) / float64(s.BooleanGets) * 100
	}

	return intEff, floatEff, boolEff
}
