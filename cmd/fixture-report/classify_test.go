package main

import (
	"strings"
	"testing"
)

func TestLineDiff_DistanceIsPerHunkEditDistance(t *testing.T) {
	tests := []struct {
		name              string
		expected, got     []string
		missing, spurious int
		distance          int
	}{
		{
			name:     "one line replaced by another is one edit, not two",
			expected: []string{"a", "WRONG", "c"},
			got:      []string{"a", "RIGHT", "c"},
			missing:  1, spurious: 1, distance: 1,
		},
		{
			name:     "two adjacent lines replaced by two others is two edits",
			expected: []string{"a", "X1", "X2"},
			got:      []string{"a", "Y1", "Y2"},
			missing:  2, spurious: 2, distance: 2,
		},
		{
			name:     "an unrelated omission and an unrelated extra are separate hunks",
			expected: []string{"keep", "dropped", "same", "tail"},
			got:      []string{"keep", "same", "added", "tail"},
			missing:  1, spurious: 1, distance: 2,
		},
		{
			name:     "a missing line alone is one edit",
			expected: []string{"a", "b"},
			got:      []string{"a"},
			missing:  1, spurious: 0, distance: 1,
		},
		{
			name:     "identical input has no distance",
			expected: []string{"a", "b"},
			got:      []string{"a", "b"},
			missing:  0, spurious: 0, distance: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missing, spurious, distance := lineDiff(tt.expected, tt.got)
			if len(missing) != tt.missing || len(spurious) != tt.spurious {
				t.Errorf("missing=%v spurious=%v, want %d/%d", missing, spurious, tt.missing, tt.spurious)
			}
			if distance != tt.distance {
				t.Errorf("distance = %d, want %d", distance, tt.distance)
			}
		})
	}
}

func TestLineDiff_FallsBackOnRunawayOutput(t *testing.T) {
	// A fixture that loops until the timeout produces enough lines that the DP table
	// would dominate the run; the fallback must still report the surplus.
	expected := []string{"only line"}
	got := make([]string, lcsCellLimit+1)
	for i := range got {
		got[i] = "spam"
	}
	missing, spurious, distance := lineDiff(expected, got)
	if len(missing) != 1 {
		t.Errorf("missing = %d, want the one expected line", len(missing))
	}
	if len(spurious) != len(got) || distance != len(got) {
		t.Errorf("spurious = %d, distance = %d, want %d for both", len(spurious), distance, len(got))
	}
}

func TestShapeOf_CollapsesNamesAndPositions(t *testing.T) {
	tests := []struct{ line, want string }{
		{
			`Syntax Error: Incompatible types: "Integer" and "String" [line: 3, column: 12]`,
			`Syntax Error: Incompatible types: "X" and "X"`,
		},
		{
			`Syntax Error: Incompatible types: "TFoo" and "TBar" [line: 41, column: 7]`,
			`Syntax Error: Incompatible types: "X" and "X"`,
		},
		{`Hint: Empty FOR loop [line: 5, column: 1]`, `Hint: Empty FOR loop`},
		{`Syntax Error: argument 2 to method 'Add' has type Integer`, `Syntax Error: argument N to method 'X' has type Integer`},
	}
	for _, tt := range tests {
		if got := shapeOf(tt.line); got != tt.want {
			t.Errorf("shapeOf(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
	// Two diagnostics differing only in their names and anchor are one shape, which is
	// the whole point: shapes are what rank the work.
	if shapeOf(tests[0].line) != shapeOf(tests[1].line) {
		t.Error("two Incompatible types diagnostics must collapse to one shape")
	}
}

func TestIsDiagnostic_SeparatesMessagesFromProgramOutput(t *testing.T) {
	diagnostics := []string{
		`Syntax Error: Name expected [line: 1, column: 1]`,
		`Hint: Result is never used [line: 2, column: 3]`,
		`Warning: Assignment to FOR-Loop variable [line: 4, column: 3]`,
		`Runtime Error: Upper bound exceeded!`,
	}
	for _, line := range diagnostics {
		if !isDiagnostic(line) {
			t.Errorf("%q must be a diagnostic", line)
		}
	}
	for _, line := range []string{"Hello, World!", "42", "Errors >>>>", "Result >>>>"} {
		if isDiagnostic(line) {
			t.Errorf("%q is program output, not a diagnostic", line)
		}
	}
}

func TestClassify_BucketsByWhatDiffers(t *testing.T) {
	tests := []struct {
		name          string
		expected, got string
		kind          string
		distance      int
	}{
		{
			name:     "only the diagnostic's wording is wrong",
			expected: "Syntax Error: DO expected [line: 5, column: 12]",
			got:      "Syntax Error: expected 'do' after for-in collection [line: 5, column: 12]",
			kind:     kindDiagnostics, distance: 1,
		},
		{
			name:     "the program printed the wrong value",
			expected: "Errors >>>>\nResult >>>>\n42",
			got:      "Errors >>>>\nResult >>>>\n41",
			kind:     kindOutput, distance: 1,
		},
		{
			name:     "a wrong diagnostic and wrong output together",
			expected: "Syntax Error: Name expected [line: 1, column: 1]\n7",
			got:      "Syntax Error: unexpected token [line: 1, column: 1]\n8",
			kind:     kindMixed, distance: 2,
		},
		{
			name:     "nothing came out at all",
			expected: "Syntax Error: Name expected [line: 1, column: 1]",
			got:      "",
			kind:     kindEmpty, distance: 1,
		},
		{
			name:     "the fixture never terminated",
			expected: "anything",
			got:      timeoutSentinel,
			kind:     kindTimeout, distance: 1,
		},
		{
			name:     "the compiler crashed",
			expected: "Syntax Error: Name expected [line: 1, column: 1]",
			got:      "panic: runtime error: invalid memory address\n\ngoroutine 1 [running]:",
			kind:     kindCrash, distance: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := classify(tt.expected, tt.got)
			if c.kind != tt.kind {
				t.Errorf("kind = %q, want %q", c.kind, tt.kind)
			}
			if c.distance != tt.distance {
				t.Errorf("distance = %d, want %d", c.distance, tt.distance)
			}
		})
	}
}

func TestTallyShapes_RanksByFixturesAndPredictsYield(t *testing.T) {
	// One fixture emitting a shape twenty times is still one fixture, and the pass rate
	// counts fixtures — so the shape blocking more of them ranks higher even though it
	// accounts for fewer lines.
	noisy := make([]string, 0, 20)
	for i := 0; i < 20; i++ {
		noisy = append(noisy, `Syntax Error: Noisy [line: 1, column: 1]`)
	}
	classes := []classification{
		{kind: kindDiagnostics, missing: noisy},
		{kind: kindDiagnostics, missing: []string{`Syntax Error: Widespread [line: 1, column: 1]`}},
		{kind: kindDiagnostics, missing: []string{`Syntax Error: Widespread [line: 2, column: 1]`}},
		// Widespread is here too, but so is another shape, so fixing Widespread alone
		// would not close this one.
		{kind: kindDiagnostics, missing: []string{
			`Syntax Error: Widespread [line: 3, column: 1]`,
			`Syntax Error: Other [line: 4, column: 1]`,
		}},
		// Widespread is the only missing shape, but the fixture also prints the wrong
		// value, so it stays failing whatever happens to the diagnostic.
		{kind: kindMixed, missing: []string{`Syntax Error: Widespread [line: 5, column: 1]`},
			spurious: []string{"41"}},
	}
	tallies := tallyShapes(classes, func(c classification) []string { return c.missing })

	if len(tallies) != 3 {
		t.Fatalf("expected 3 shapes, got %d: %v", len(tallies), tallies)
	}
	if tallies[0].shape != "Syntax Error: Widespread" || tallies[0].fixtures != 4 || tallies[0].lines != 4 {
		t.Errorf("first tally = %+v, want the shape mentioned by 4 fixtures", tallies[0])
	}
	if tallies[0].sole != 2 {
		t.Errorf("sole = %d, want 2: the fixture with a second shape and the one with wrong output do not count",
			tallies[0].sole)
	}
	if tallies[1].fixtures != 1 || tallies[1].lines != 20 || tallies[1].sole != 1 {
		t.Errorf("second tally = %+v, want 1 fixture / 20 lines / sole 1", tallies[1])
	}
}

func TestTallyShapes_IgnoresProgramOutput(t *testing.T) {
	tallies := tallyShapes(
		[]classification{{kind: kindOutput, missing: []string{"42", "Hello"}}},
		func(c classification) []string { return c.missing },
	)
	if len(tallies) != 0 {
		t.Errorf("program output is not a message shape, got %v", tallies)
	}
}

func TestSplitLines_KeepsInteriorBlanks(t *testing.T) {
	// normalize() has already stripped the leading and trailing blanks by the time a
	// line list is built, so a surviving blank is one the program printed. Scoring it
	// away would make a fixture whose only defect is a missing blank line look like a
	// failure at distance 0 (SimpleScripts/print_multi_args).
	if got := splitLines("a\n\nb"); strings.Join(got, "|") != "a||b" {
		t.Errorf("splitLines = %v, want the blank line kept", got)
	}
	if got := splitLines(""); got != nil {
		t.Errorf("splitLines(\"\") = %v, want no lines at all", got)
	}
}

func TestClassify_MissingBlankLineIsOneEdit(t *testing.T) {
	c := classify("hello\n\nworld", "hello\nworld")
	if c.distance != 1 || c.kind != kindOutput {
		t.Errorf("classify = %+v, want one output edit", c)
	}
}
