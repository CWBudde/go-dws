package runtime

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

func newTestAssoc(key, elem types.Type) *AssociativeArrayValue {
	return NewAssociativeArrayValue(types.NewAssociativeArrayType(key, elem))
}

func TestAssociativeArray_SetGetLenDelete(t *testing.T) {
	a := newTestAssoc(types.INTEGER, types.STRING)

	if _, ok := a.Get(&IntegerValue{Value: 1}); ok {
		t.Fatal("empty map should not contain key 1")
	}
	if a.Len() != 0 {
		t.Fatalf("Len = %d, want 0", a.Len())
	}

	a.Set(&IntegerValue{Value: 1}, &StringValue{Value: "one"})
	a.Set(&IntegerValue{Value: 2}, &StringValue{Value: "two"})
	if a.Len() != 2 {
		t.Fatalf("Len = %d, want 2", a.Len())
	}

	// Updating an existing key must not grow the map.
	a.Set(&IntegerValue{Value: 1}, &StringValue{Value: "uno"})
	if a.Len() != 2 {
		t.Fatalf("Len after update = %d, want 2", a.Len())
	}
	if v, ok := a.Get(&IntegerValue{Value: 1}); !ok || v.String() != "uno" {
		t.Fatalf("Get(1) = %v, %v; want uno,true", v, ok)
	}

	if !a.Contains(&IntegerValue{Value: 2}) {
		t.Fatal("Contains(2) = false, want true")
	}

	if !a.Delete(&IntegerValue{Value: 1}) {
		t.Fatal("Delete(1) = false, want true (was present)")
	}
	if a.Delete(&IntegerValue{Value: 1}) {
		t.Fatal("Delete(1) again = true, want false (already removed)")
	}
	if a.Len() != 1 {
		t.Fatalf("Len after delete = %d, want 1", a.Len())
	}

	a.Clear()
	if a.Len() != 0 {
		t.Fatalf("Len after Clear = %d, want 0", a.Len())
	}
}

func TestAssociativeArray_KeysInsertionOrder(t *testing.T) {
	a := newTestAssoc(types.STRING, types.INTEGER)
	a.Set(&StringValue{Value: "b"}, &IntegerValue{Value: 2})
	a.Set(&StringValue{Value: "a"}, &IntegerValue{Value: 1})
	keys := a.Keys()
	if len(keys) != 2 || keys[0].String() != "b" || keys[1].String() != "a" {
		t.Fatalf("Keys = %v, want [b a] (insertion order)", keys)
	}
}

func TestAssociativeArray_ObjectKeysUseIdentity(t *testing.T) {
	a := newTestAssoc(types.STRING, types.INTEGER) // key type unused for the check
	// Distinct instances with a nil class share String() == "<nil> instance",
	// so a String()-based comparison would wrongly collapse them. Keys must use
	// pointer identity.
	o1 := NewObjectInstance(nil)
	o2 := NewObjectInstance(nil)
	if o1.String() != o2.String() {
		t.Fatal("precondition: nil-class instances should share String()")
	}

	a.Set(o1, &IntegerValue{Value: 1})
	a.Set(o2, &IntegerValue{Value: 2})
	if a.Len() != 2 {
		t.Fatalf("two distinct object keys collapsed to Len %d, want 2", a.Len())
	}
	if v, ok := a.Get(o1); !ok || v.String() != "1" {
		t.Fatalf("Get(o1) = %v,%v; want 1,true", v, ok)
	}
	if v, ok := a.Get(o2); !ok || v.String() != "2" {
		t.Fatalf("Get(o2) = %v,%v; want 2,true", v, ok)
	}
	if a.Contains(NewObjectInstance(nil)) {
		t.Fatal("a third distinct instance must not be present")
	}
}

func TestAssociativeArray_ReferenceCopy(t *testing.T) {
	a := newTestAssoc(types.INTEGER, types.STRING)
	a.Set(&IntegerValue{Value: 1}, &StringValue{Value: "one"})
	if c := a.Copy(); c != Value(a) {
		t.Fatal("Copy must return the receiver (reference semantics)")
	}
}

// TestAssociativeArray_SetReportsReplacedSlot covers the ARC hook on Set: an
// overwrite must hand the displaced value back so the caller can release it,
// while a fresh insert reports nothing to release.
func TestAssociativeArray_SetReportsReplacedSlot(t *testing.T) {
	tests := []struct {
		name         string
		wantPrev     string
		preset       bool
		wantReplaced bool
	}{
		{name: "fresh insert", preset: false, wantReplaced: false},
		{name: "overwrite", preset: true, wantReplaced: true, wantPrev: "old"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAssoc(types.STRING, types.STRING)
			if tt.preset {
				a.Set(&StringValue{Value: "k"}, &StringValue{Value: "old"})
			}

			prev, replaced := a.Set(&StringValue{Value: "k"}, &StringValue{Value: "new"})
			if replaced != tt.wantReplaced {
				t.Fatalf("replaced = %v, want %v", replaced, tt.wantReplaced)
			}
			if !tt.wantReplaced {
				if prev != nil {
					t.Fatalf("prev = %v, want nil", prev)
				}
				return
			}
			str, ok := prev.(*StringValue)
			if !ok || str.Value != tt.wantPrev {
				t.Fatalf("prev = %v, want %q", prev, tt.wantPrev)
			}
			if a.Len() != 1 {
				t.Fatalf("Len = %d, want 1", a.Len())
			}
		})
	}
}

// TestAssociativeArray_DeleteEntry checks that DeleteEntry hands back both the
// stored key and the stored value, which the evaluator needs in order to
// release the map's references to them.
func TestAssociativeArray_DeleteEntry(t *testing.T) {
	a := newTestAssoc(types.STRING, types.STRING)
	a.Set(&StringValue{Value: "k"}, &StringValue{Value: "v"})

	key, value, ok := a.DeleteEntry(&StringValue{Value: "k"})
	if !ok {
		t.Fatal("DeleteEntry reported a missing key")
	}
	if k, isStr := key.(*StringValue); !isStr || k.Value != "k" {
		t.Fatalf("key = %v, want \"k\"", key)
	}
	if v, isStr := value.(*StringValue); !isStr || v.Value != "v" {
		t.Fatalf("value = %v, want \"v\"", value)
	}
	if a.Len() != 0 {
		t.Fatalf("Len = %d, want 0", a.Len())
	}

	if _, _, ok := a.DeleteEntry(&StringValue{Value: "k"}); ok {
		t.Fatal("second DeleteEntry should report a missing key")
	}
}

// TestAssociativeArray_TakeEntries checks that TakeEntries empties the map and
// that a second take yields nothing, so contents can never be released twice.
func TestAssociativeArray_TakeEntries(t *testing.T) {
	a := newTestAssoc(types.STRING, types.STRING)
	a.Set(&StringValue{Value: "k1"}, &StringValue{Value: "v1"})
	a.Set(&StringValue{Value: "k2"}, &StringValue{Value: "v2"})

	keys, values := a.TakeEntries()
	if len(keys) != 2 || len(values) != 2 {
		t.Fatalf("took %d keys / %d values, want 2 / 2", len(keys), len(values))
	}
	if a.Len() != 0 {
		t.Fatalf("Len = %d after take, want 0", a.Len())
	}

	keys, values = a.TakeEntries()
	if len(keys) != 0 || len(values) != 0 {
		t.Fatalf("second take yielded %d keys / %d values, want 0 / 0", len(keys), len(values))
	}
}
