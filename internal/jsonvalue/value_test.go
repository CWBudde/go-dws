package jsonvalue

import "testing"

func TestKindString(t *testing.T) {
	tests := []struct {
		kind     Kind
		expected string
	}{
		{KindUndefined, "Undefined"},
		{KindNull, "Null"},
		{KindObject, "Object"},
		{KindArray, "Array"},
		{KindString, "String"},
		{KindNumber, "Number"},
		{KindInt64, "Int64"},
		{KindBoolean, "Boolean"},
		{Kind(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.expected {
			t.Errorf("Kind(%d).String() = %q, want %q", tt.kind, got, tt.expected)
		}
	}
}

func TestValueConstructors(t *testing.T) {
	if kind := NewNull().Kind(); kind != KindNull {
		t.Fatalf("NewNull kind = %v, want %v", kind, KindNull)
	}
	if kind := NewBoolean(true).Kind(); kind != KindBoolean {
		t.Fatalf("NewBoolean kind = %v, want %v", kind, KindBoolean)
	}
	if kind := NewNumber(1.23).Kind(); kind != KindNumber {
		t.Fatalf("NewNumber kind = %v, want %v", kind, KindNumber)
	}
	if kind := NewInt64(42).Kind(); kind != KindInt64 {
		t.Fatalf("NewInt64 kind = %v, want %v", kind, KindInt64)
	}
	if kind := NewString("foo").Kind(); kind != KindString {
		t.Fatalf("NewString kind = %v, want %v", kind, KindString)
	}
	if kind := NewArray().Kind(); kind != KindArray {
		t.Fatalf("NewArray kind = %v, want %v", kind, KindArray)
	}
	if kind := NewObject().Kind(); kind != KindObject {
		t.Fatalf("NewObject kind = %v, want %v", kind, KindObject)
	}
}

func TestObjectOperations(t *testing.T) {
	obj := NewObject()
	obj.ObjectSet("foo", NewString("bar"))
	obj.ObjectSet("baz", NewInt64(7))
	obj.ObjectSet("foo", NewString("updated"))

	if got := obj.ObjectGet("foo"); got == nil || got.Kind() != KindString {
		t.Fatalf("ObjectGet foo = %#v, want KindString", got)
	}
	if obj.ObjectGet("missing") != nil {
		t.Fatalf("ObjectGet missing should be nil")
	}
	keys := obj.ObjectKeys()
	wantOrder := []string{"foo", "baz"}
	if len(keys) != len(wantOrder) {
		t.Fatalf("ObjectKeys length = %d, want %d", len(keys), len(wantOrder))
	}
	for i, key := range wantOrder {
		if keys[i] != key {
			t.Fatalf("ObjectKeys[%d] = %s, want %s", i, keys[i], key)
		}
	}
	if !obj.ObjectDelete("foo") {
		t.Fatalf("ObjectDelete foo = false, want true")
	}
	if obj.ObjectGet("foo") != nil {
		t.Fatalf("foo should be removed")
	}
	if obj.ObjectDelete("does-not-exist") {
		t.Fatalf("delete missing key should be false")
	}
}

func TestArrayOperations(t *testing.T) {
	arr := NewArray()
	arr.ArrayAppend(NewInt64(1))
	arr.ArrayAppend(NewInt64(2))
	arr.ArrayAppend(NewInt64(3))

	if got := arr.ArrayLen(); got != 3 {
		t.Fatalf("ArrayLen = %d, want 3", got)
	}

	if !arr.ArraySet(1, NewString("two")) {
		t.Fatalf("ArraySet index 1 failed")
	}
	if elem := arr.ArrayGet(1); elem == nil || elem.Kind() != KindString {
		t.Fatalf("ArrayGet[1] = %#v, want KindString", elem)
	}

	if arr.ArraySet(5, NewInt64(10)) {
		t.Fatalf("ArraySet out of bounds should be false")
	}

	if !arr.ArrayDelete(0) {
		t.Fatalf("ArrayDelete index 0 failed")
	}
	if arr.ArrayLen() != 2 {
		t.Fatalf("Array length after delete = %d, want 2", arr.ArrayLen())
	}

	if arr.ArrayDelete(10) {
		t.Fatalf("ArrayDelete out of range should be false")
	}

	elements := arr.ArrayElements()
	if len(elements) != arr.ArrayLen() {
		t.Fatalf("ArrayElements length = %d, want %d", len(elements), arr.ArrayLen())
	}
	if &elements[0] == &arr.arrElems[0] {
		t.Fatalf("ArrayElements should return a copy")
	}
}
