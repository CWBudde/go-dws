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
