package builtins

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// stringVar is a shorthand for a stored String Variant.
func stringVar(s string) GlobalVarValue {
	return GlobalVarValue{Kind: GlobalVarString, Str: s}
}

// intVar is a shorthand for a stored Integer Variant.
func intVar(i int64) GlobalVarValue {
	return GlobalVarValue{Kind: GlobalVarInteger, Int: i}
}

func TestGlobalVarStore_WriteReadDelete(t *testing.T) {
	store := NewGlobalVarStore()

	if _, found := store.Read("missing"); found {
		t.Fatalf("Read of an absent global reported found")
	}

	store.Write("test", stringVar("hello"), 0)
	value, found := store.Read("test")
	if !found || value.Str != "hello" {
		t.Fatalf("Read(test) = %#v, %v; want hello, true", value, found)
	}

	if !store.Delete("test") {
		t.Fatalf("Delete(test) reported the global was absent")
	}
	if store.Delete("test") {
		t.Fatalf("Delete(test) on an already deleted global reported present")
	}
}

func TestGlobalVarStore_NamesAreCaseSensitiveButMasksAreNot(t *testing.T) {
	// DWScript keeps 'hello' and 'Hello' as separate globals while matching
	// the mask 'h*' against both (fixture FunctionsGlobalVars/names).
	store := NewGlobalVarStore()
	store.Write("hello", stringVar("world"), 0)
	store.Write("Hello", stringVar("world"), 0)
	store.Write("Byebye", stringVar("world"), 0)

	tests := []struct {
		name string
		mask string
		want []string
	}{
		{name: "all", mask: "*", want: []string{"Byebye", "Hello", "hello"}},
		{name: "case insensitive prefix", mask: "h*", want: []string{"Hello", "hello"}},
		{name: "no match", mask: "z*", want: []string{}},
		{name: "empty mask means all", mask: "", want: []string{"Byebye", "Hello", "hello"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := store.Names(tc.mask)
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Fatalf("Names(%q) = %v; want %v", tc.mask, got, tc.want)
			}
		})
	}
}

func TestGlobalVarStore_Cleanup(t *testing.T) {
	tests := []struct {
		name string
		mask string
		want string
	}{
		{name: "prefix", mask: "Test.*", want: "Alpha.Test"},
		{name: "contains", mask: "*Alpha*", want: "Test.Beta"},
		{name: "everything", mask: "*", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := NewGlobalVarStore()
			store.Write("Test.Alpha", intVar(1), 0)
			store.Write("Alpha.Test", intVar(1), 0)
			store.Write("Test.Beta", intVar(1), 0)

			store.Cleanup(tc.mask)
			if got := strings.Join(store.Names("*"), ","); got != tc.want {
				t.Fatalf("after Cleanup(%q) names = %q; want %q", tc.mask, got, tc.want)
			}
		})
	}
}

func TestGlobalVarStore_Expiration(t *testing.T) {
	store := NewGlobalVarStore()
	base := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	current := base
	store.SetClock(func() time.Time { return current })

	store.Write("alpha", intVar(10), 0)     // never expires
	store.Write("beta", intVar(20), 0.001)  // expires after 1ms
	store.Write("gamma", intVar(30), 0.001) // expires after 1ms

	// Incrementing with an expiration keeps the lifetime; incrementing without
	// one clears it (fixture FunctionsGlobalVars/inc_expire).
	if got := store.Increment("gamma", 1, 0.001); got != 31 {
		t.Fatalf("Increment(gamma) = %d; want 31", got)
	}
	if got := store.Increment("beta", 1, 0); got != 21 {
		t.Fatalf("Increment(beta) = %d; want 21", got)
	}

	current = base.Add(10 * time.Millisecond)

	if _, found := store.Read("gamma"); found {
		t.Fatalf("gamma should have expired")
	}
	if value, found := store.Read("beta"); !found || value.Int != 21 {
		t.Fatalf("beta = %#v, %v; want 21, true", value, found)
	}
	if value, found := store.Read("alpha"); !found || value.Int != 10 {
		t.Fatalf("alpha = %#v, %v; want 10, true", value, found)
	}
	if got := store.Names("*"); strings.Join(got, ",") != "alpha,beta" {
		t.Fatalf("Names after expiry = %v; want [alpha beta]", got)
	}
}

func TestGlobalVarStore_IncrementCreatesMissingGlobal(t *testing.T) {
	store := NewGlobalVarStore()
	if got := store.Increment("hi", 123, 0); got != 123 {
		t.Fatalf("Increment on a missing global = %d; want 123", got)
	}
	if got := store.Increment("hi", 1, 0); got != 124 {
		t.Fatalf("Increment = %d; want 124", got)
	}
}

func TestGlobalVarStore_CompareExchange(t *testing.T) {
	unassigned := GlobalVarValue{Kind: GlobalVarUnassigned}

	tests := []struct {
		name      string
		initial   *GlobalVarValue
		value     GlobalVarValue
		comparand GlobalVarValue
		wantPrev  GlobalVarValue
		wantAfter GlobalVarValue
	}{
		{
			name:      "absent does not match a value comparand",
			value:     intVar(2),
			comparand: intVar(1),
			wantPrev:  unassigned,
			wantAfter: unassigned,
		},
		{
			name:      "absent matches Unassigned",
			value:     intVar(0),
			comparand: unassigned,
			wantPrev:  unassigned,
			wantAfter: intVar(0),
		},
		{
			name:      "matching comparand exchanges",
			initial:   ptrTo(intVar(1)),
			value:     intVar(2),
			comparand: intVar(1),
			wantPrev:  intVar(1),
			wantAfter: intVar(2),
		},
		{
			name:      "mismatched comparand leaves the value",
			initial:   ptrTo(stringVar("alpha")),
			value:     stringVar("beta"),
			comparand: intVar(2),
			wantPrev:  stringVar("alpha"),
			wantAfter: stringVar("alpha"),
		},
		{
			name:      "string comparand exchanges",
			initial:   ptrTo(stringVar("alpha")),
			value:     stringVar("beta"),
			comparand: stringVar("alpha"),
			wantPrev:  stringVar("alpha"),
			wantAfter: stringVar("beta"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := NewGlobalVarStore()
			if tc.initial != nil {
				store.Write("test", *tc.initial, 0)
			}
			if got := store.CompareExchange("test", tc.value, tc.comparand); got != tc.wantPrev {
				t.Fatalf("CompareExchange previous = %#v; want %#v", got, tc.wantPrev)
			}
			after, _ := store.Read("test")
			if after != tc.wantAfter {
				t.Fatalf("stored value = %#v; want %#v", after, tc.wantAfter)
			}
		})
	}
}

// ptrTo returns a pointer to v, for table-driven optional fields.
func ptrTo(v GlobalVarValue) *GlobalVarValue { return &v }

func TestGlobalVarStore_SaveAndLoadRoundTrip(t *testing.T) {
	store := NewGlobalVarStore()
	empty := store.SaveToString()

	store.Write("test", stringVar("hello world"), 0)
	store.Write("testNum", intVar(123), 0)
	store.Write("testFloat", GlobalVarValue{Kind: GlobalVarFloat, Float: 2.5}, 0)
	store.Write("testBool", GlobalVarValue{Kind: GlobalVarBoolean, Bool: true}, 0)
	full := store.SaveToString()

	if err := store.LoadFromString(empty); err != nil {
		t.Fatalf("LoadFromString(empty snapshot) = %v", err)
	}
	if names := store.Names("*"); len(names) != 0 {
		t.Fatalf("store should be empty after restoring the empty snapshot, got %v", names)
	}

	if err := store.LoadFromString(full); err != nil {
		t.Fatalf("LoadFromString(full snapshot) = %v", err)
	}
	if value, found := store.Read("testFloat"); !found || value.Float != 2.5 {
		t.Fatalf("testFloat = %#v, %v; want 2.5, true", value, found)
	}
	if value, found := store.Read("testBool"); !found || !value.Bool {
		t.Fatalf("testBool = %#v, %v; want true, true", value, found)
	}
}

func TestGlobalVarStore_LoadFromStringErrors(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{name: "empty clears the store", data: "", wantErr: false},
		{name: "unknown tag", data: "Z", wantErr: true},
		{name: "wrong header", data: "NOPE\n[]", wantErr: true},
		{name: "corrupt payload", data: globalVarsFileTag + "\nnot json", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := NewGlobalVarStore()
			store.Write("x", intVar(1), 0)
			err := store.LoadFromString(tc.data)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("LoadFromString(%q) = nil; want an error", tc.data)
				}
				if err.Error() != "Invalid file tag" {
					t.Fatalf("error = %q; want %q", err.Error(), "Invalid file tag")
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadFromString(%q) = %v; want nil", tc.data, err)
			}
		})
	}
}

func TestGlobalVarStore_QueueOrdering(t *testing.T) {
	// Push appends to the back, Insert prepends to the front, Pull removes from
	// the front and Pop removes from the back
	// (fixture FunctionsGlobalVars/queue_basic).
	store := NewGlobalVarStore()
	store.QueuePush("test", intVar(1))
	store.QueuePush("test", intVar(2))
	store.QueueInsert("test", intVar(3))

	if got := store.QueueLength("test"); got != 3 {
		t.Fatalf("QueueLength = %d; want 3", got)
	}

	store.QueueInsert("test", intVar(4))

	front, ok := store.QueuePull("test")
	if !ok || front.Int != 4 {
		t.Fatalf("QueuePull = %#v, %v; want 4, true", front, ok)
	}

	var popped []int64
	for {
		value, ok := store.QueuePop("test")
		if !ok {
			break
		}
		popped = append(popped, value.Int)
	}
	want := []int64{2, 1, 3}
	if len(popped) != len(want) {
		t.Fatalf("popped = %v; want %v", popped, want)
	}
	for i := range want {
		if popped[i] != want[i] {
			t.Fatalf("popped = %v; want %v", popped, want)
		}
	}

	if _, ok := store.QueuePull("test"); ok {
		t.Fatalf("QueuePull on an empty queue reported success")
	}
}

func TestGlobalVarStore_QueuePeekAndFirst(t *testing.T) {
	store := NewGlobalVarStore()
	if _, ok := store.QueuePeek("test"); ok {
		t.Fatalf("QueuePeek on an empty queue reported success")
	}
	if _, ok := store.QueueFirst("test"); ok {
		t.Fatalf("QueueFirst on an empty queue reported success")
	}

	store.QueuePush("test", intVar(456))
	store.QueuePush("test", stringVar("hello"))
	store.QueueInsert("test", stringVar("world"))

	back, _ := store.QueuePeek("test")
	if back.Str != "hello" {
		t.Fatalf("QueuePeek = %#v; want hello", back)
	}
	front, _ := store.QueueFirst("test")
	if front.Str != "world" {
		t.Fatalf("QueueFirst = %#v; want world", front)
	}

	snapshot := store.QueueSnapshot("test")
	if len(snapshot) != 3 || snapshot[0].Str != "world" || snapshot[2].Str != "hello" {
		t.Fatalf("QueueSnapshot = %#v; want world, 456, hello", snapshot)
	}
	// A snapshot must not alias the live queue.
	snapshot[0] = intVar(-1)
	if again, _ := store.QueueFirst("test"); again.Str != "world" {
		t.Fatalf("mutating a snapshot changed the queue")
	}
}

func TestGlobalVarStore_CleanupQueues(t *testing.T) {
	store := NewGlobalVarStore()
	store.QueuePush("alpha", intVar(1))
	store.QueuePush("beta", intVar(1))

	store.CleanupQueues("alp*")
	if store.QueueLength("alpha") != 0 {
		t.Fatalf("alpha queue survived a matching cleanup")
	}
	if store.QueueLength("beta") != 1 {
		t.Fatalf("beta queue was removed by a non-matching cleanup")
	}

	store.CleanupQueues("")
	if store.QueueLength("beta") != 0 {
		t.Fatalf("beta queue survived a full cleanup")
	}
}

func TestFromRuntimeValue(t *testing.T) {
	tests := []struct {
		value   Value
		name    string
		wantErr string
		want    GlobalVarValue
	}{
		{name: "integer", value: &runtime.IntegerValue{Value: 7}, want: intVar(7)},
		{name: "string", value: &runtime.StringValue{Value: "s"}, want: stringVar("s")},
		{name: "float", value: &runtime.FloatValue{Value: 1.5}, want: GlobalVarValue{Kind: GlobalVarFloat, Float: 1.5}},
		{name: "boolean", value: &runtime.BooleanValue{Value: true}, want: GlobalVarValue{Kind: GlobalVarBoolean, Bool: true}},
		{name: "null", value: &runtime.NullValue{}, want: GlobalVarValue{Kind: GlobalVarNull}},
		{name: "unassigned", value: &runtime.UnassignedValue{}, want: GlobalVarValue{Kind: GlobalVarUnassigned}},
		{name: "nil interface becomes Null", value: &runtime.InterfaceInstance{}, want: GlobalVarValue{Kind: GlobalVarNull}},
		{
			name:    "object is rejected",
			value:   &runtime.ObjectInstance{},
			wantErr: "Cannot store global of type TClassSymbol",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FromRuntimeValue(tc.value)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("FromRuntimeValue error = %v; want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("FromRuntimeValue = %v", err)
			}
			if got != tc.want {
				t.Fatalf("FromRuntimeValue = %#v; want %#v", got, tc.want)
			}
		})
	}
}

func TestGlobalVarStore_ConcurrentAccess(t *testing.T) {
	// Run with -race: every exported entry point must be safe to call from
	// several goroutines at once, since the store is process-wide.
	store := NewGlobalVarStore()
	const workers = 8
	const iterations = 200

	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				store.Increment("counter", 1, 0)
				store.Write("shared", intVar(int64(i)), 0)
				store.Read("shared")
				store.Names("*")
				store.CompareExchange("cas", intVar(1), GlobalVarValue{Kind: GlobalVarUnassigned})
				store.QueuePush("queue", intVar(int64(i)))
				store.QueueInsert("queue", intVar(int64(i)))
				store.QueueSnapshot("queue")
				store.QueuePull("queue")
				store.QueuePop("queue")
				store.QueueLength("queue")
			}
		}()
	}
	wg.Wait()

	counter, found := store.Read("counter")
	if !found || counter.Int != workers*iterations {
		t.Fatalf("counter = %#v, %v; want %d", counter, found, workers*iterations)
	}
	if store.QueueLength("queue") != 0 {
		t.Fatalf("queue should be drained, got %d entries", store.QueueLength("queue"))
	}
}
