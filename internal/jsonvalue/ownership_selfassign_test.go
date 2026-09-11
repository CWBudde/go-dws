package jsonvalue

import "testing"

// TestOwnership_SelfAssignmentKeepsSlot covers the `o.a := o.a` / `a[i] := a[i]`
// shape: re-inserting a node into the very slot it already occupies must not
// reorder object keys, drop array elements or leave a stale owner pointer.
func TestOwnership_SelfAssignmentKeepsSlot(t *testing.T) {
	t.Run("object key keeps its position", func(t *testing.T) {
		obj := NewObject()
		obj.ObjectSet("a", NewInt64(1))
		obj.ObjectSet("b", NewInt64(2))
		obj.ObjectSet("a", obj.ObjectGet("a"))
		if got, want := Stringify(obj), `{"a":1,"b":2}`; got != want {
			t.Errorf("Stringify() = %s, want %s", got, want)
		}
		if obj.ObjectGet("a").Owner() != obj {
			t.Error("self-assigned entry lost its owner")
		}
	})

	t.Run("array slot is a no-op", func(t *testing.T) {
		arr := NewArray()
		arr.ArrayAppend(NewInt64(1))
		arr.ArrayAppend(NewInt64(2))
		if !arr.ArraySet(0, arr.ArrayGet(0)) {
			t.Fatal("ArraySet reported failure for a self-assignment")
		}
		if got, want := Stringify(arr), `[1,2]`; got != want {
			t.Errorf("Stringify() = %s, want %s", got, want)
		}
	})
}

// TestOwnership_ArraySetSameArrayMove pins the post-detach layout used when a
// node is moved from one slot of an array into another slot of the same array.
func TestOwnership_ArraySetSameArrayMove(t *testing.T) {
	t.Run("moving forward", func(t *testing.T) {
		arr := NewArray()
		arr.ArrayAppend(NewInt64(1))
		arr.ArrayAppend(NewInt64(2))
		arr.ArrayAppend(NewInt64(3))
		if !arr.ArraySet(2, arr.ArrayGet(0)) {
			t.Fatal("ArraySet reported failure")
		}
		// 1 leaves slot 0, the array collapses to [2,3] and 1 overwrites 3.
		if got, want := Stringify(arr), `[2,1]`; got != want {
			t.Errorf("Stringify() = %s, want %s", got, want)
		}
		for i := 0; i < arr.ArrayLen(); i++ {
			if arr.ArrayGet(i).Owner() != arr {
				t.Errorf("element %d lost its owner", i)
			}
		}
	})

	t.Run("single element array keeps the node", func(t *testing.T) {
		arr := NewArray()
		arr.ArrayAppend(NewInt64(1))
		if !arr.ArraySet(0, arr.ArrayGet(0)) {
			t.Fatal("ArraySet reported failure")
		}
		if got, want := Stringify(arr), `[1]`; got != want {
			t.Errorf("Stringify() = %s, want %s", got, want)
		}
	})
}
