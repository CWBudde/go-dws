package jsonvalue

import "testing"

// TestOwnership_InsertAdoptsAndReparents covers DWScript's single-owner rule: a
// node inserted into a container is removed from wherever it lived before.
func TestOwnership_InsertAdoptsAndReparents(t *testing.T) {
	tests := []struct {
		name  string
		build func() *Value
		want  string
	}{
		{
			name: "object insert sets owner",
			build: func() *Value {
				root := NewObject()
				child := NewObject()
				root.ObjectSet("a", child)
				if child.Owner() != root {
					t.Fatalf("owner not set on ObjectSet")
				}
				return root
			},
			want: `{"a":{}}`,
		},
		{
			name: "move between objects detaches from the first",
			build: func() *Value {
				root := NewObject()
				dati := NewObject()
				campo := NewObject()
				campo.ObjectSet("IDValue", NewInt64(5))
				dati.ObjectSet("Campo", campo)
				root.ObjectSet("Dati", dati)
				root.ObjectSet("SottoOggetto", campo)
				return root
			},
			want: `{"Dati":{},"SottoOggetto":{"IDValue":5}}`,
		},
		{
			name: "clone keeps the original in place",
			build: func() *Value {
				root := NewObject()
				dati := NewObject()
				campo := NewObject()
				campo.ObjectSet("IDValue", NewInt64(5))
				dati.ObjectSet("Campo", campo)
				root.ObjectSet("Dati", dati)
				root.ObjectSet("SottoOggetto", campo.Clone())
				return root
			},
			want: `{"Dati":{"Campo":{"IDValue":5}},"SottoOggetto":{"IDValue":5}}`,
		},
		{
			name: "append to another array removes it from the first",
			build: func() *Value {
				a := NewArray()
				b := NewObject()
				a.ArrayAppend(b)
				a2 := NewArray()
				a2.ArrayAppend(a.ArrayGet(0))
				if a.ArrayLen() != 0 {
					t.Fatalf("source array still holds the node: len=%d", a.ArrayLen())
				}
				return a2
			},
			want: `[{}]`,
		},
		{
			name: "array set of an already-owned node leaves a hole",
			build: func() *Value {
				a := NewArray()
				b := NewObject()
				b.ObjectSet("ID", NewString("r"))
				a.ArrayAppend(b)
				b.Detach()
				for a.ArrayLen() <= 1 {
					a.ArrayAppend(NewNull())
				}
				a.ArraySet(1, b)
				return a
			},
			want: `[null,{"ID":"r"}]`,
		},
		{
			name: "swap does not reparent",
			build: func() *Value {
				a := NewArray()
				a.ArrayAppend(NewInt64(1))
				a.ArrayAppend(NewInt64(2))
				a.ArraySwap(0, 1)
				return a
			},
			want: `[2,1]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Stringify(tt.build())
			if got != tt.want {
				t.Errorf("Stringify() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestOwnership_RemovalClearsOwner(t *testing.T) {
	obj := NewObject()
	child := NewObject()
	obj.ObjectSet("a", child)
	obj.ObjectDelete("a")
	if child.Owner() != nil {
		t.Errorf("ObjectDelete left an owner behind")
	}

	arr := NewArray()
	elem := NewObject()
	arr.ArrayAppend(elem)
	arr.ArrayDelete(0)
	if elem.Owner() != nil {
		t.Errorf("ArrayDelete left an owner behind")
	}

	arr.ArrayAppend(elem)
	arr.ClearArray()
	if elem.Owner() != nil {
		t.Errorf("ClearArray left an owner behind")
	}

	// Replacing an entry releases the displaced node.
	replaced := NewObject()
	obj.ObjectSet("a", replaced)
	obj.ObjectSet("a", NewInt64(1))
	if replaced.Owner() != nil {
		t.Errorf("ObjectSet replacement left an owner behind")
	}
}

func TestOwnership_DetachOfRootIsNoOp(t *testing.T) {
	root := NewObject()
	root.Detach()
	if root.Owner() != nil {
		t.Errorf("root gained an owner")
	}
	var nilValue *Value
	nilValue.Detach() // must not panic
}

func TestOwnership_ParsedChildrenAreOwned(t *testing.T) {
	root, err := Parse(`{"a":{"b":1},"c":[2]}`)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if root.Owner() != nil {
		t.Errorf("parsed root should have no owner")
	}
	a := root.ObjectGet("a")
	if a.Owner() != root {
		t.Errorf("parsed object member is not owned by its container")
	}
	c := root.ObjectGet("c")
	if c.ArrayGet(0).Owner() != c {
		t.Errorf("parsed array element is not owned by its container")
	}
}
