package runtime

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
)

// AssociativeArrayValue is the runtime value for a DWScript associative array
// (`array [KeyType] of ElementType`): a sparse map keyed by an arbitrary type.
//
// Storage is a direct port of DWScript's `TScriptAssociativeArray`
// (`Source/dwsAssociativeArrays.pas`): an open-addressing hash table with
// linear probing over three parallel bucket slices. This is not merely an
// optimisation over the previous insertion-ordered linear scan — the bucket
// layout is *observable*, because `Keys` walks the bucket array in index order
// and fixtures pin the resulting order. See `associative_hash.go` for the
// ported hash function.
//
// Invariants mirrored from upstream:
//   - a hash code of 0 marks an empty bucket (the hash function never yields 0),
//   - capacity is a power of two, starting at 32 and doubling,
//   - the table grows once count reaches capacity*11/16,
//   - growth and deletion preserve the probe invariant, so lookups stay correct.
//
// Associative arrays are reference types: assignment shares the backing map
// (Copy returns the receiver), like dynamic arrays.
type AssociativeArrayValue struct {
	AssocType *types.AssociativeArrayType

	hashes []uint32 // 0 == empty bucket
	keys   []Value  // parallel to hashes; value-typed keys are snapshotted
	values []Value  // parallel to hashes

	count    int
	capacity int
	growth   int

	bindings int // live named bindings sharing this map (ARC, see RetainBinding)
}

// RetainBinding records that one more named binding (variable, parameter,
// field) now shares this map. Because associative arrays are reference types,
// the entries they own must outlive every binding, not just the one whose
// scope happens to end first.
func (a *AssociativeArrayValue) RetainBinding() {
	if a == nil {
		return
	}
	a.bindings++
}

// ReleaseBinding drops one named binding and reports whether that was the last
// one, i.e. whether the caller must now release the map's retained keys and
// values. A map that was never retained (bindings == 0) never reports true, so
// an unbalanced release cannot empty a map somebody still holds.
func (a *AssociativeArrayValue) ReleaseBinding() bool {
	if a == nil || a.bindings <= 0 {
		return false
	}
	a.bindings--
	return a.bindings == 0
}

// Compile-time interface satisfaction check.
var _ CopyableValue = (*AssociativeArrayValue)(nil)

// NewAssociativeArrayValue creates an empty associative array of the given type.
func NewAssociativeArrayValue(t *types.AssociativeArrayType) *AssociativeArrayValue {
	return &AssociativeArrayValue{AssocType: t}
}

// Type returns "ASSOCIATIVE_ARRAY".
func (a *AssociativeArrayValue) Type() string { return "ASSOCIATIVE_ARRAY" }

// String returns a debug representation "[k: v, ...]" in bucket order.
func (a *AssociativeArrayValue) String() string {
	parts := make([]string, 0, a.count)
	for i := 0; i < a.capacity; i++ {
		if a.hashes[i] == 0 {
			continue
		}
		key := "nil"
		if a.keys[i] != nil {
			key = a.keys[i].String()
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

// grow doubles the bucket array (or allocates the initial 32 buckets) and
// reinserts every live entry. Ported from TScriptAssociativeArray.Grow.
func (a *AssociativeArrayValue) grow() {
	if a.capacity == 0 {
		a.capacity = 32
	} else {
		a.capacity *= 2
	}
	a.growth = (a.capacity * 11) / 16

	oldHashes, oldKeys, oldValues := a.hashes, a.keys, a.values
	a.hashes = make([]uint32, a.capacity)
	a.keys = make([]Value, a.capacity)
	a.values = make([]Value, a.capacity)

	mask := a.capacity - 1
	for i, h := range oldHashes {
		if h == 0 {
			continue
		}
		j := int(h) & mask
		for a.hashes[j] != 0 {
			j = (j + 1) & mask
		}
		a.hashes[j] = h
		a.keys[j] = oldKeys[i]
		a.values[j] = oldValues[i]
	}
}

// linearFind probes from the bucket index derived from hash until it finds a
// matching key (true) or an empty bucket (false). Either way index holds the
// bucket to use: the match, or the free slot where the key would be inserted.
// Ported from TScriptAssociativeArray.LinearFind.
func (a *AssociativeArrayValue) linearFind(key Value, hash uint32) (index int, found bool) {
	mask := a.capacity - 1
	i := int(hash) & mask
	for {
		if a.hashes[i] == 0 {
			return i, false
		}
		if a.hashes[i] == hash && associativeKeyEqual(a.keys[i], key) {
			return i, true
		}
		i = (i + 1) & mask
	}
}

// coerceKey applies the implicit conversion that DWScript's compiler inserts on
// an associative-array key expression: `ReadSymbolArrayExpr` wraps a key whose
// type is not the declared key type with `WrapWithImplicitConversion`, so an
// Integer index into an `array [Float] of ...` is converted before it ever
// reaches the hash table.
//
// The conversion is load-bearing here because the hash is type-sensitive: a
// script integer is a `varInt64` and a float a `varDouble`, which hash
// differently even when numerically equal. Without this coercion `a[1] := x`
// followed by `a[1.0]` would probe a different bucket and miss.
func (a *AssociativeArrayValue) coerceKey(key Value) Value {
	keyType := a.KeyType()
	if keyType == nil || !types.GetUnderlyingType(keyType).Equals(types.FLOAT) {
		return key
	}
	switch k := key.(type) {
	case *IntegerValue:
		return NewFloat(float64(k.Value))
	case *VariantValue:
		if inner, ok := k.Value.(*IntegerValue); ok {
			return NewFloat(float64(inner.Value))
		}
	}
	return key
}

// Get returns the value stored at key and whether the key is present.
func (a *AssociativeArrayValue) Get(key Value) (Value, bool) {
	if a.count == 0 {
		return nil, false
	}
	key = a.coerceKey(key)
	if i, found := a.linearFind(key, associativeHashCode(key)); found {
		return a.values[i], true
	}
	return nil, false
}

// Set inserts or updates the value at key. Value-typed keys are snapshotted so
// later mutation of the caller's key variable does not alter the stored key.
//
// It returns the value previously stored at key and whether an existing slot
// was overwritten, so callers can apply ARC to the displaced value.
func (a *AssociativeArrayValue) Set(key, value Value) (prev Value, replaced bool) {
	if a.count >= a.growth {
		a.grow()
	}
	key = a.coerceKey(key)
	hash := associativeHashCode(key)
	i, found := a.linearFind(key, hash)
	if !found {
		a.hashes[i] = hash
		a.keys[i] = cloneKey(key)
		a.count++
	}
	prev, replaced = a.values[i], found
	a.values[i] = value
	return prev, replaced
}

// Delete removes the entry at key, returning whether it was present.
//
// Ported from TScriptAssociativeArray.Delete: open addressing cannot simply
// blank a bucket, because that would cut the probe chain of any later entry
// that collided with it. Instead the following cluster is scanned and each
// entry that may legally move back is shifted into the gap.
func (a *AssociativeArrayValue) Delete(key Value) bool {
	_, _, ok := a.DeleteEntry(key)
	return ok
}

// DeleteEntry removes the entry at key and returns the stored key and value
// alongside whether the key was present. Callers own the ARC release of both.
//
// The backward shift below overwrites the matched bucket, so the entry is read
// out before the cluster is repaired.
func (a *AssociativeArrayValue) DeleteEntry(key Value) (storedKey, storedValue Value, ok bool) {
	if a.count == 0 {
		return nil, nil, false
	}
	key = a.coerceKey(key)
	i, found := a.linearFind(key, associativeHashCode(key))
	if !found {
		return nil, nil, false
	}
	storedKey, storedValue = a.keys[i], a.values[i]

	mask := a.capacity - 1
	gap := i
	for {
		i = (i + 1) & mask
		if a.hashes[i] == 0 {
			break
		}
		k := int(a.hashes[i]) & mask
		if ((gap >= k) || (k > i)) && ((i >= gap) || (k <= gap)) && ((i >= gap) || (k > i)) {
			a.hashes[gap] = a.hashes[i]
			a.keys[gap] = a.keys[i]
			a.values[gap] = a.values[i]
			a.hashes[i] = 0
			gap = i
		}
	}

	// Clear the freed slot so the removed key/value become eligible for GC.
	a.hashes[gap] = 0
	a.keys[gap] = nil
	a.values[gap] = nil
	a.count--
	return storedKey, storedValue, true
}

// Contains reports whether key is present.
func (a *AssociativeArrayValue) Contains(key Value) bool {
	_, ok := a.Get(key)
	return ok
}

// Len returns the number of live entries.
func (a *AssociativeArrayValue) Len() int { return a.count }

// Clear removes all entries.
func (a *AssociativeArrayValue) Clear() {
	a.hashes = nil
	a.keys = nil
	a.values = nil
	a.count = 0
	a.capacity = 0
	a.growth = 0
}

// TakeEntries empties the array and returns the stored keys and their parallel
// values, compacted out of the bucket array in iteration order. Callers own the
// ARC release of everything returned; emptying first makes a second take a
// no-op, so the contents can never be released twice.
func (a *AssociativeArrayValue) TakeEntries() (keys, values []Value) {
	if a.count > 0 {
		keys = make([]Value, 0, a.count)
		values = make([]Value, 0, a.count)
		for i := 0; i < a.capacity; i++ {
			if a.hashes[i] != 0 {
				keys = append(keys, a.keys[i])
				values = append(values, a.values[i])
			}
		}
	}
	a.Clear()
	return keys, values
}

// Keys returns the keys in bucket order (a fresh slice), matching DWScript's
// TScriptAssociativeArray.CopyKeys. Value-typed keys (records, static arrays)
// are snapshotted so a caller mutating a returned key cannot corrupt the map's
// internal key set; object keys keep their identity.
func (a *AssociativeArrayValue) Keys() []Value {
	out := make([]Value, 0, a.count)
	for i := 0; i < a.capacity; i++ {
		if a.hashes[i] != 0 {
			out = append(out, cloneKey(a.keys[i]))
		}
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
