package jsonvalue

import (
	"errors"
	"testing"
)

func TestOwnership_CycleRejectedBeforeMutation(t *testing.T) {
	for _, operation := range []string{"object", "append", "replace"} {
		t.Run(operation, func(t *testing.T) {
			root := NewObject()
			child := NewObject()
			array := NewArray()
			root.ObjectSet("child", child)
			child.ObjectSet("array", array)
			array.ArrayAppend(NewInt64(1))
			before := Stringify(root)
			if err := array.ValidateAdoption(root); !errors.Is(err, ErrCircularReference) {
				t.Fatalf("ValidateAdoption(root) = %v, want circular reference", err)
			}
			switch operation {
			case "object":
				child.ObjectSet("array", root)
			case "append":
				array.ArrayAppend(root)
			case "replace":
				if array.ArraySet(0, root) {
					t.Fatal("ArraySet accepted a cycle")
				}
			}
			if root.Owner() != nil || child.Owner() != root || array.Owner() != child {
				t.Fatal("rejected insertion changed ownership")
			}
			if got := Stringify(root); got != before {
				t.Fatalf("rejected insertion changed contents: %s", got)
			}
		})
	}
}

func TestOwnership_AdoptionValidation(t *testing.T) {
	object, array := NewObject(), NewArray()
	for _, value := range []*Value{object, array} {
		if err := value.ValidateAdoption(value); !errors.Is(err, ErrCircularReference) {
			t.Fatalf("self-adoption = %v, want circular reference", err)
		}
		if err := value.ValidateAdoption(nil); err != nil {
			t.Fatalf("nil adoption = %v", err)
		}
	}
	object.ObjectSet("array", array)
	if err := object.ValidateAdoption(array); err != nil {
		t.Fatalf("same-owner child adoption = %v", err)
	}
	if err := NewObject().ValidateAdoption(array); err != nil {
		t.Fatalf("valid reparent = %v", err)
	}
}
