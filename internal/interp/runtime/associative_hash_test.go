package runtime

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

// TestXXHash32_ReferenceVectors pins our xxHash32 port against the published
// xxHash32 reference vectors (seed 0). DWScript hashes string keys with this
// exact function, so any drift here silently reorders every associative array.
//
// The inputs are chosen to cover all three code paths: the short tail-only
// path (< 16 bytes), the boundary, and the four-accumulator path for longer
// inputs, which no fixture exercises.
func TestXXHash32_ReferenceVectors(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want uint32
	}{
		{"empty", "", 0x02cc5d05},
		{"1 byte", "a", 0x550d7456},
		{"3 bytes", "abc", 0x32d153ff},
		{"4 bytes", "abcd", 0xa3643705},
		{"8 bytes", "abcdefgh", 0x0bb3c6bb},
		{"10 bytes", "abcdefghij", 0x8b988cfe},
		{"16 bytes", "abcdefghijklmnop", 0x9d2d8b62},
		{"36 bytes", "abcdefghijklmnopqrstuvwxyz0123456789", 0x42ae804d},
		{"445 bytes", lipsum, 0x62b4ed00},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := xxHash32([]byte(tc.in), 0); got != tc.want {
				t.Errorf("xxHash32(%d bytes) = %#08x, want %#08x", len(tc.in), got, tc.want)
			}
		})
	}
}

// lipsum is the 445-byte reference-vector input.
const lipsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

// TestAssociativeHashCode_Buckets pins the bucket index of the keys that appear
// in order-sensitive fixtures, computed with the ported DWScript hash.
func TestAssociativeHashCode_Buckets(t *testing.T) {
	tests := []struct {
		key        Value
		name       string
		wantHash   uint32
		wantBucket int // hash mod the initial capacity of 32
	}{
		{&StringValue{Value: "a"}, "string a", 3664126901, 21},
		{&StringValue{Value: "b"}, "string b", 3229501451, 11},
		{&StringValue{Value: "toto"}, "string toto", 3440707429, 5},
		{&IntegerValue{Value: 10}, "integer 10", 2295946489, 25},
		{&IntegerValue{Value: 11}, "integer 11", 2111288658, 18},
		{&IntegerValue{Value: 20}, "integer 20", 2679957210, 26},
		{&IntegerValue{Value: 21}, "integer 21", 2433484647, 7},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := associativeHashCode(tc.key)
			if got != tc.wantHash {
				t.Errorf("associativeHashCode(%s) = %d, want %d", tc.key.String(), got, tc.wantHash)
			}
			if bucket := int(got) & 31; bucket != tc.wantBucket {
				t.Errorf("bucket(%s) = %d, want %d", tc.key.String(), bucket, tc.wantBucket)
			}
		})
	}
}

// TestAssociativeHashCode_NeverZero checks the invariant the bucket array
// relies on: 0 marks an empty bucket, so no live key may hash to 0.
func TestAssociativeHashCode_NeverZero(t *testing.T) {
	keys := []Value{
		&StringValue{Value: ""},
		&StringValue{Value: "a"},
		&IntegerValue{Value: 0},
		&IntegerValue{Value: -1},
		&FloatValue{Value: 0},
		&BooleanValue{Value: false},
		&BooleanValue{Value: true},
		&NilValue{},
	}
	for i := int64(-500); i <= 500; i++ {
		keys = append(keys, &IntegerValue{Value: i})
	}
	for _, k := range keys {
		if h := associativeHashCode(k); h == 0 {
			t.Errorf("associativeHashCode(%s) = 0, which would read as an empty bucket", k.String())
		}
	}
}

// TestAssociativeArray_FixtureKeyOrder reproduces the key orders that the
// DWScript fixture corpus pins. These are the observable consequence of the
// ported hash and bucket layout, and are the reason this port exists.
func TestAssociativeArray_FixtureKeyOrder(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		keyType types.Type
		insert  []Value
		want    []string
	}{
		{
			name:    "string keys a then b",
			fixture: "AssociativePass/records.pas",
			keyType: types.STRING,
			insert:  []Value{&StringValue{Value: "a"}, &StringValue{Value: "b"}},
			want:    []string{"b", "a"},
		},
		{
			name:    "single string key",
			fixture: "AssociativePass/records.pas",
			keyType: types.STRING,
			insert:  []Value{&StringValue{Value: "toto"}},
			want:    []string{"toto"},
		},
		{
			name:    "integer keys 10 11 20 21",
			fixture: "JSONConnectorPass/associative_array.pas",
			keyType: types.INTEGER,
			insert: []Value{
				&IntegerValue{Value: 10}, &IntegerValue{Value: 11},
				&IntegerValue{Value: 20}, &IntegerValue{Value: 21},
			},
			want: []string{"21", "11", "10", "20"},
		},
		{
			name:    "string keys 1.2 then 3.4",
			fixture: "JSONConnectorPass/associative_array.pas",
			keyType: types.STRING,
			insert:  []Value{&StringValue{Value: "1.2"}, &StringValue{Value: "3.4"}},
			want:    []string{"1.2", "3.4"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestAssoc(tc.keyType, types.INTEGER)
			for i, k := range tc.insert {
				a.Set(k, &IntegerValue{Value: int64(i)})
			}
			keys := a.Keys()
			got := make([]string, len(keys))
			for i, k := range keys {
				got[i] = k.String()
			}
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("Keys() = %v, want %v (order pinned by %s)", got, tc.want, tc.fixture)
			}
		})
	}
}

// TestAssociativeArray_GrowthPreservesLookups exercises the rehash path: the
// table starts at capacity 32 and doubles once count reaches capacity*11/16.
func TestAssociativeArray_GrowthPreservesLookups(t *testing.T) {
	const n = 200
	a := newTestAssoc(types.INTEGER, types.INTEGER)
	for i := 0; i < n; i++ {
		a.Set(&IntegerValue{Value: int64(i)}, &IntegerValue{Value: int64(i * 3)})
	}
	if a.Len() != n {
		t.Fatalf("Len = %d, want %d", a.Len(), n)
	}
	if a.capacity <= 32 {
		t.Fatalf("capacity = %d, expected the table to have grown past 32", a.capacity)
	}
	if a.count > a.growth {
		t.Errorf("count %d exceeds growth threshold %d", a.count, a.growth)
	}
	for i := 0; i < n; i++ {
		v, ok := a.Get(&IntegerValue{Value: int64(i)})
		if !ok {
			t.Fatalf("key %d lost across growth", i)
		}
		if iv, isInt := v.(*IntegerValue); !isInt || iv.Value != int64(i*3) {
			t.Fatalf("Get(%d) = %v, want %d", i, v, i*3)
		}
	}
	if len(a.Keys()) != n {
		t.Errorf("Keys() length = %d, want %d", len(a.Keys()), n)
	}
}

// TestAssociativeArray_DeleteKeepsProbeChains is the regression guard for the
// backward-shift deletion: blanking a bucket outright would cut the probe chain
// of any entry that had collided with it, silently losing keys.
func TestAssociativeArray_DeleteKeepsProbeChains(t *testing.T) {
	const n = 150
	a := newTestAssoc(types.INTEGER, types.INTEGER)
	for i := 0; i < n; i++ {
		a.Set(&IntegerValue{Value: int64(i)}, &IntegerValue{Value: int64(i)})
	}

	// Delete every third key, then assert both halves of the partition.
	for i := 0; i < n; i += 3 {
		if !a.Delete(&IntegerValue{Value: int64(i)}) {
			t.Fatalf("Delete(%d) reported absent", i)
		}
	}
	for i := 0; i < n; i++ {
		got := a.Contains(&IntegerValue{Value: int64(i)})
		want := i%3 != 0
		if got != want {
			t.Errorf("Contains(%d) = %v, want %v", i, got, want)
		}
	}
	if wantLen := n - (n+2)/3; a.Len() != wantLen {
		t.Errorf("Len = %d, want %d", a.Len(), wantLen)
	}

	// Reinserting a deleted key must find a slot and be reachable again.
	a.Set(&IntegerValue{Value: 0}, &IntegerValue{Value: 99})
	if v, ok := a.Get(&IntegerValue{Value: 0}); !ok || v.String() != "99" {
		t.Errorf("Get(0) after reinsert = %v,%v; want 99,true", v, ok)
	}
}

// TestAssociativeArray_KeyKindsRoundTrip checks that each key kind we hash is
// also matched correctly, so hashing and equality agree.
func TestAssociativeArray_KeyKindsRoundTrip(t *testing.T) {
	tests := []struct {
		key  Value
		miss Value
		name string
	}{
		{&IntegerValue{Value: 42}, &IntegerValue{Value: 43}, "integer"},
		{&IntegerValue{Value: -42}, &IntegerValue{Value: 42}, "negative integer"},
		{&FloatValue{Value: 1.5}, &FloatValue{Value: 2.5}, "float"},
		{&StringValue{Value: "hello"}, &StringValue{Value: "world"}, "string"},
		{&StringValue{Value: ""}, &StringValue{Value: " "}, "empty string"},
		{&StringValue{Value: "grüße"}, &StringValue{Value: "grusse"}, "unicode string"},
		{&BooleanValue{Value: true}, &BooleanValue{Value: false}, "boolean true"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestAssoc(types.VARIANT, types.INTEGER)
			a.Set(tc.key, &IntegerValue{Value: 1})
			if _, ok := a.Get(tc.key); !ok {
				t.Errorf("Get(%s) missed its own key", tc.key.String())
			}
			if a.Contains(tc.miss) {
				t.Errorf("Contains(%s) = true, want false", tc.miss.String())
			}
			if !a.Delete(tc.key) {
				t.Errorf("Delete(%s) = false, want true", tc.key.String())
			}
			if a.Len() != 0 {
				t.Errorf("Len after delete = %d, want 0", a.Len())
			}
		})
	}
}

// TestAssociativeArray_EqualKeysHashEqually is the invariant that makes the
// hash-first fast path in linearFind safe: structurally equal keys must produce
// the same hash code, or a lookup would probe past its own entry.
func TestAssociativeArray_EqualKeysHashEqually(t *testing.T) {
	pairs := [][2]Value{
		{&IntegerValue{Value: 7}, &IntegerValue{Value: 7}},
		{&StringValue{Value: "abc"}, &StringValue{Value: "abc"}},
		{&FloatValue{Value: 2.25}, &FloatValue{Value: 2.25}},
		{&BooleanValue{Value: true}, &BooleanValue{Value: true}},
	}
	for _, p := range pairs {
		if !associativeKeyEqual(p[0], p[1]) {
			t.Fatalf("precondition: %s and %s should compare equal", p[0].String(), p[1].String())
		}
		if h0, h1 := associativeHashCode(p[0]), associativeHashCode(p[1]); h0 != h1 {
			t.Errorf("equal keys %s hash differently: %d vs %d", p[0].String(), h0, h1)
		}
	}
}

// TestAssociativeArray_FloatKeyTypeCoercesIntegerKeys pins the implicit
// conversion DWScript's compiler performs on the key expression: for an
// `array [Float] of ...` an Integer index must be converted to Float before
// hashing, otherwise it hashes as varInt64 and misses the varDouble bucket.
func TestAssociativeArray_FloatKeyTypeCoercesIntegerKeys(t *testing.T) {
	a := newTestAssoc(types.FLOAT, types.INTEGER)
	a.Set(&IntegerValue{Value: 1}, &IntegerValue{Value: 7})

	if got, ok := a.Get(&FloatValue{Value: 1.0}); !ok {
		t.Error("Get(1.0) missed the key stored as integer 1")
	} else if iv, isInt := got.(*IntegerValue); !isInt || iv.Value != 7 {
		t.Errorf("Get(1.0) = %v, want 7", got)
	}
	if _, ok := a.Get(&IntegerValue{Value: 1}); !ok {
		t.Error("Get(1) missed its own key")
	}

	a.Set(&FloatValue{Value: 1.0}, &IntegerValue{Value: 8})
	if a.Len() != 1 {
		t.Errorf("Len = %d, want 1: integer and float key must share a bucket", a.Len())
	}

	if !a.Delete(&IntegerValue{Value: 1}) {
		t.Error("Delete(1) = false, want true")
	}
	if a.Len() != 0 {
		t.Errorf("Len after delete = %d, want 0", a.Len())
	}
}

// TestAssociativeArray_VariantKeyTypeKeepsNumericRepresentation documents the
// deliberate counterpart: with a Variant key type there is no declared type to
// convert to, so an integer and a float key stay distinct, exactly as upstream
// hashes varInt64 and varDouble differently.
func TestAssociativeArray_VariantKeyTypeKeepsNumericRepresentation(t *testing.T) {
	a := newTestAssoc(types.VARIANT, types.INTEGER)
	a.Set(&IntegerValue{Value: 1}, &IntegerValue{Value: 1})
	a.Set(&FloatValue{Value: 1.0}, &IntegerValue{Value: 2})
	if a.Len() != 2 {
		t.Errorf("Len = %d, want 2", a.Len())
	}
}
