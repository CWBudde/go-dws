package runtime

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
)

// AssociativeArrayValue is the runtime value for a DWScript associative array
// (`array [KeyType] of ElementType`): a sparse map keyed by an arbitrary type.
//
// Storage is a pair of parallel slices in insertion order with linear-scan
// lookup. Fixture maps are small (≤150 entries) so this is sufficient; a hash
// index can be added later without changing the surface.
//
// Associative arrays are reference types: assignment shares the backing map
// (Copy returns the receiver), like dynamic arrays.
type AssociativeArrayValue struct {
	AssocType *types.AssociativeArrayType
	keys      []Value // insertion order; value-typed keys are snapshotted
	values    []Value // parallel to keys
}

// Compile-time interface satisfaction check.
var _ CopyableValue = (*AssociativeArrayValue)(nil)

// NewAssociativeArrayValue creates an empty associative array of the given type.
func NewAssociativeArrayValue(t *types.AssociativeArrayType) *AssociativeArrayValue {
	return &AssociativeArrayValue{AssocType: t}
}

// Type returns "ASSOCIATIVE_ARRAY".
func (a *AssociativeArrayValue) Type() string { return "ASSOCIATIVE_ARRAY" }

// String returns a debug representation "[k: v, ...]".
func (a *AssociativeArrayValue) String() string {
	parts := make([]string, 0, len(a.keys))
	for i, k := range a.keys {
		key := "nil"
		if k != nil {
			key = k.String()
		}
		v := "nil"
		if a.values[i] != nil {
			v = a.values[i].String()
		}
		parts = append(parts, key+": "+v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// Copy returns the receiver: associative arrays have reference semantics.
func (a *AssociativeArrayValue) Copy() Value { return a }

// KeyType returns the declared key type (may be nil).
func (a *AssociativeArrayValue) KeyType() types.Type {
	if a.AssocType == nil {
		return nil
	}
	return a.AssocType.KeyType
}

// ElementType returns the declared element (value) type (may be nil).
func (a *AssociativeArrayValue) ElementType() types.Type {
	if a.AssocType == nil {
		return nil
	}
	return a.AssocType.ElementType
}

func (a *AssociativeArrayValue) indexOf(key Value) int {
	for i, k := range a.keys {
		if associativeKeyEqual(k, key) {
			return i
		}
	}
	return -1
}

// Get returns the value stored at key and whether the key is present.
func (a *AssociativeArrayValue) Get(key Value) (Value, bool) {
	if i := a.indexOf(key); i >= 0 {
		return a.values[i], true
	}
	return nil, false
}

// Set inserts or updates the value at key. Value-typed keys are snapshotted so
// later mutation of the caller's key variable does not alter the stored key.
func (a *AssociativeArrayValue) Set(key, value Value) {
	if i := a.indexOf(key); i >= 0 {
		a.values[i] = value
		return
	}
	a.keys = append(a.keys, cloneKey(key))
	a.values = append(a.values, value)
}

// Delete removes the entry at key, returning whether it was present.
func (a *AssociativeArrayValue) Delete(key Value) bool {
	i := a.indexOf(key)
	if i < 0 {
		return false
	}
	// Shift down, then clear the freed tail slots so the removed key/value (and
	// anything they reference) become eligible for GC.
	copy(a.keys[i:], a.keys[i+1:])
	copy(a.values[i:], a.values[i+1:])
	last := len(a.keys) - 1
	a.keys[last] = nil
	a.values[last] = nil
	a.keys = a.keys[:last]
	a.values = a.values[:last]
	return true
}

// Contains reports whether key is present.
func (a *AssociativeArrayValue) Contains(key Value) bool { return a.indexOf(key) >= 0 }

// Len returns the number of live entries.
func (a *AssociativeArrayValue) Len() int { return len(a.keys) }

// Clear removes all entries.
func (a *AssociativeArrayValue) Clear() {
	a.keys = nil
	a.values = nil
}

// Keys returns the keys in insertion order (a fresh slice). Value-typed keys
// (records, static arrays) are snapshotted so a caller mutating a returned key
// cannot corrupt the map's internal key set; object keys keep their identity.
func (a *AssociativeArrayValue) Keys() []Value {
	out := make([]Value, len(a.keys))
	for i, k := range a.keys {
		out[i] = cloneKey(k)
	}
	return out
}

// associativeKeyEqual compares two associative-array keys: objects by pointer
// identity (their String() collides across instances), nil keys as equal, and
// everything else (primitives, records, static arrays) structurally via Equal.
func associativeKeyEqual(a, b Value) bool {
	aNil, bNil := isNilKey(a), isNilKey(b)
	if aNil || bNil {
		return aNil && bNil
	}
	if ao, ok := a.(*ObjectInstance); ok {
		bo, ok := b.(*ObjectInstance)
		return ok && ao == bo
	}
	if _, ok := b.(*ObjectInstance); ok {
		return false
	}
	eq, err := Equal(a, b)
	return err == nil && eq
}

func isNilKey(v Value) bool {
	if v == nil {
		return true
	}
	_, ok := v.(*NilValue)
	return ok
}

// cloneKey snapshots value-typed keys (records, static arrays) so that mutating
// the original key variable does not change a stored key; objects are kept by
// reference (identity is the key).
func cloneKey(k Value) Value {
	if _, isObj := k.(*ObjectInstance); isObj {
		return k
	}
	return CopyValue(k)
}
