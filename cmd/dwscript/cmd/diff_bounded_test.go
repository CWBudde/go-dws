package cmd

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// disjointInputs returns two texts of n lines that share no line at all, so
// that the edit distance is the worst possible 2n.
func disjointInputs(n int) (a, b []string) {
	a = make([]string, n)
	b = make([]string, n)
	for i := 0; i < n; i++ {
		a[i] = fmt.Sprintf("old line %d", i)
		b[i] = fmt.Sprintf("new line %d", i)
	}
	return a, b
}

// TestMyersDiff_BoundedMemoryOnDisjointInputs pins the memory bound: two
// 5,000-line inputs with no line in common used to retain a full trace
// snapshot per edit distance - about 1.6 GB - which is enough to get the
// process killed. The bounded fallback must keep the allocation small and
// still produce a valid whole-file replacement.
func TestMyersDiff_BoundedMemoryOnDisjointInputs(t *testing.T) {
	a, b := disjointInputs(5000)

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	ops := myersDiff(a, b)
	runtime.ReadMemStats(&after)

	allocated := after.TotalAlloc - before.TotalAlloc
	const budget = 256 << 20 // generous: the old code allocated ~1.6 GB here
	if allocated > budget {
		t.Errorf("myersDiff allocated %d bytes, want at most %d", allocated, budget)
	}

	// The fallback deletes every old line and inserts every new one.
	oldCount, newCount := countLines(ops)
	if oldCount != len(a) || newCount != len(b) {
		t.Errorf("edit script covers %d old / %d new lines, want %d / %d",
			oldCount, newCount, len(a), len(b))
	}
	for _, op := range ops {
		if op.Kind == diffEqual {
			t.Fatalf("fallback script must contain no equal lines")
		}
	}
}

// TestMyersDiff_ExactBelowBound checks that inputs below maxDiffEditDistance
// still get the exact minimal edit script rather than the fallback: the
// unchanged lines must be reported as equal.
func TestMyersDiff_ExactBelowBound(t *testing.T) {
	const n = 4000
	a := make([]string, n)
	b := make([]string, n)
	for i := 0; i < n; i++ {
		a[i] = fmt.Sprintf("line %d", i)
		b[i] = a[i]
	}
	// Change 100 scattered lines: edit distance 200, well below the bound.
	for i := 0; i < n; i += 40 {
		b[i] = "changed " + b[i]
	}

	ops := myersDiff(a, b)
	equal := 0
	for _, op := range ops {
		if op.Kind == diffEqual {
			equal++
		}
	}
	if equal != n-100 {
		t.Errorf("got %d equal lines, want %d (the fallback was taken for a small edit distance)", equal, n-100)
	}
}

// TestMyersDiff_Degenerate covers the empty and single-sided inputs, which
// exercise the distance-0 snapshot and the trace bounds.
func TestMyersDiff_Degenerate(t *testing.T) {
	tests := []struct {
		name                 string
		a, b                 []string
		wantOld, wantNew     int
		wantEqualOpsAtLeast1 bool
	}{
		{name: "both empty", a: nil, b: nil},
		{name: "insert into empty", b: []string{"x", "y"}, wantNew: 2},
		{name: "delete to empty", a: []string{"x", "y"}, wantOld: 2},
		{name: "identical", a: []string{"x"}, b: []string{"x"}, wantOld: 1, wantNew: 1, wantEqualOpsAtLeast1: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := myersDiff(tt.a, tt.b)
			oldCount, newCount := countLines(ops)
			if oldCount != tt.wantOld || newCount != tt.wantNew {
				t.Errorf("covered %d old / %d new lines, want %d / %d", oldCount, newCount, tt.wantOld, tt.wantNew)
			}
			if tt.wantEqualOpsAtLeast1 {
				if len(ops) != 1 || ops[0].Kind != diffEqual {
					t.Errorf("identical inputs should yield a single equal op, got %v", ops)
				}
			}
		})
	}
}

// TestFormatUnifiedDiff_FallbackShape checks that the bounded fallback still
// renders a well-formed unified diff: one hunk replacing the whole file.
func TestFormatUnifiedDiff_FallbackShape(t *testing.T) {
	a, b := disjointInputs(5000)
	ops := myersDiff(a, b)
	hunks := buildHunks(ops)
	if len(hunks) != 1 {
		t.Fatalf("expected a single whole-file hunk, got %d", len(hunks))
	}
	h := hunks[0]
	if h.oldStart != 0 || h.newStart != 0 || h.oldCount != len(a) || h.newCount != len(b) {
		t.Errorf("hunk = @@ -%s +%s @@, want the whole file",
			formatRange(h.oldStart, h.oldCount), formatRange(h.newStart, h.newCount))
	}
	if got := formatRange(h.oldStart, h.oldCount); !strings.HasPrefix(got, "1,") {
		t.Errorf("old range = %q, want it to start at line 1", got)
	}
}
