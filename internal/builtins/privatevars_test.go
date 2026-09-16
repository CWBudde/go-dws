package builtins

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPrivateVarStore_IsolationAndPersistence(t *testing.T) {
	store := NewPrivateVarStore()
	store.Write("FirstUnit", "Name", stringVar("first"), 0)
	store.Write("FirstUnit", "name", stringVar("lowercase"), 0)
	store.Write("SecondUnit", "Name", stringVar("second"), 0)
	for _, tc := range []struct{ unit, name, want string }{
		{"firstUNIT", "Name", "first"},
		{"FIRSTUNIT", "name", "lowercase"},
		{"secondunit", "Name", "second"},
	} {
		for range 2 {
			value, ok := store.Read(tc.unit, tc.name)
			if !ok || value != stringVar(tc.want) {
				t.Fatalf("Read(%q, %q) = %#v, %v; want %q", tc.unit, tc.name, value, ok, tc.want)
			}
		}
	}
	if value, ok := store.Read("ThirdUnit", "Name"); ok || !value.IsUnassigned() {
		t.Fatalf("missing unit read = %#v, %v; want Unassigned, false", value, ok)
	}
	store.Cleanup("firstunit", "*")
	if _, ok := store.Read("SecondUnit", "Name"); !ok {
		t.Fatal("cleaning first unit removed second unit's value")
	}
}

func TestPrivateVarStore_WriteResultAndExpiration(t *testing.T) {
	store := NewPrivateVarStore()
	current := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	store.SetClock(func() time.Time { return current })
	if !store.Write("Unit", "entry", intVar(1), 1) {
		t.Fatal("new entry must return true")
	}
	if store.Write("Unit", "entry", intVar(2), 1) {
		t.Fatal("live replacement must return false")
	}
	current = current.Add(time.Second)
	if !store.Write("Unit", "entry", intVar(3), 1) {
		t.Fatal("expired replacement must return true")
	}
	store.Write("Unit", "expires", intVar(4), 1)
	if store.Write("Unit", "entry", intVar(5), 0) {
		t.Fatal("clearing a live entry's expiration must return false")
	}
	current = current.Add(time.Second)
	if _, ok := store.Read("Unit", "expires"); ok {
		t.Fatal("entry should expire at its deadline")
	}
	if value, ok := store.Read("Unit", "entry"); !ok || value.Int != 5 {
		t.Fatalf("entry without expiration = %#v, %v; want 5, true", value, ok)
	}
	store.Write("Other", "old", intVar(1), 1)
	store.SetClock(func() time.Time { return current.Add(2 * time.Second) })
	if got := store.Names("other", ""); len(got) != 0 {
		t.Fatalf("clock update must affect existing unit stores; got names %v", got)
	}
	store.SetClock(nil)
}

func TestPrivateVarStore_NamesAndCleanupMasks(t *testing.T) {
	for _, tc := range []struct {
		name, mask string
		want       []string
	}{
		{"empty name only", "", []string{"Alpha", "alpha", "beta"}},
		{"case insensitive wildcard", "a*", []string{"", "beta"}},
		{"single character wildcard", "?eta", []string{"", "Alpha", "alpha"}},
		{"everything", "*", []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := NewPrivateVarStore()
			for _, name := range []string{"beta", "alpha", "", "Alpha"} {
				store.Write("Unit", name, intVar(1), 0)
			}
			if got, want := store.Names("Unit", ""), []string{"", "Alpha", "alpha", "beta"}; !reflect.DeepEqual(got, want) {
				t.Fatalf("Names with empty mask = %v; want %v", got, want)
			}
			if got, want := store.Names("Unit", "a*"), []string{"Alpha", "alpha"}; !reflect.DeepEqual(got, want) {
				t.Fatalf("Names with wildcard = %v; want %v", got, want)
			}
			store.Cleanup("Unit", tc.mask)
			if got := store.Names("Unit", ""); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Names after Cleanup(%q) = %v; want %v", tc.mask, got, tc.want)
			}
			store.Cleanup("Missing", tc.mask)
		})
	}
}

func TestPrivateVarStore_ConcurrentAccess(t *testing.T) {
	store := NewPrivateVarStore()
	var created atomic.Int64
	var wg sync.WaitGroup
	for worker := range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if store.Write("SharedUnit", "shared", intVar(int64(worker)), 0) {
				created.Add(1)
			}
			unit := fmt.Sprintf("Unit%d", worker)
			for iteration := range 20 {
				if iteration == 10 {
					store.SetClock(time.Now)
				}
				store.Write(unit, "entry", intVar(int64(iteration)), 0)
				store.Read(unit, "entry")
				store.Names(unit, "*")
				store.Cleanup(unit, "entry")
			}
		}()
	}
	wg.Wait()
	if got := created.Load(); got != 1 {
		t.Fatalf("concurrent creation count = %d; want exactly 1", got)
	}
}

func TestPrivateVarStore_GlobalOperationsAreSeparate(t *testing.T) {
	private := NewPrivateVarStore()
	global := NewGlobalVarStore()
	private.Write("Unit", "private", intVar(42), 0)
	global.Write("global", intVar(7), 0)
	if got, want := global.Names("*"), []string{"global"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("global names = %v; want %v", got, want)
	}
	snapshot := global.SaveToString()
	restored := NewGlobalVarStore()
	if err := restored.LoadFromString(snapshot); err != nil {
		t.Fatal(err)
	}
	if got, want := restored.Names("*"), []string{"global"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("global snapshot names = %v; want %v", got, want)
	}
	global.Cleanup("*")
	if value, ok := private.Read("Unit", "private"); !ok || value.Int != 42 {
		t.Fatalf("private after global cleanup = %#v, %v; want 42, true", value, ok)
	}
}
