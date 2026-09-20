package frontend

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestCompile_SetOfFailSentences pins the complete diagnostic output for the
// SetOfFail fixtures whose sentences are produced by semantic analysis
// (PLAN.md §4, F7). Each case mirrors a fixture; the expectation is that
// fixture's `.txt` verbatim.
func TestCompile_SetOfFailSentences(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			// SetOfFail/bracket_left_missing: a set's Include/Exclude is a call,
			// so upstream stops at the missing argument list rather than looking
			// the name up as a member.
			name:   "set mutator without an argument list",
			source: "type TMyEnum = (enumOne, enumTwo);\ntype TMySet = set of TMyEnum;\n\nvar e : TMySet;\n\ne.Include;\n",
			want:   []string{`Syntax Error: "(" expected [line: 6, column: 10]`},
		},
		{
			// SetOfFail/include: Include is read as `Include(set, element)`, so a
			// single argument leaves the separator missing.
			name:   "Include with only the set argument",
			source: "var t : set of (a, b);\n\nInclude(t);\n",
			want:   []string{`Syntax Error: "," expected [line: 3, column: 10]`},
		},
		{
			// SetOfFail/invalid_method: the receiver is named by its declared
			// type symbol, not by the structural `set of TMyEnum`.
			name:   "unknown member on a named set type",
			source: "type TMyEnum = (enumOne, enumTwo);\ntype TMySet = set of TMyEnum;\n\nvar e : TMySet;\n\ne.BugBugBug();\n",
			want: []string{
				`Syntax Error: There is no accessible member with name "BugBugBug" for type TMySet [line: 6, column: 3]`,
			},
		},
		{
			// SetOfFail/invalid_operand, lines 6/8/12. The fixture's fourth case
			// (`Include(s, @Test)`) additionally needs the `unexpected "@"`
			// sentence, which go-dws does not produce anywhere yet.
			name: "set element argument of the wrong type",
			source: "type TMyEnum = (enumOne, enumTwo);\ntype TMySet = set of TMyEnum;\n\nvar s : TMySet;\n\nInclude(s, 1);\n\n" +
				"Include(s, '');\n\nprocedure Test; begin end;\n\nInclude(s, Test);\n",
			want: []string{
				`Syntax Error: Incompatible parameter types - "TMyEnum" expected (instead of "Integer") [line: 6, column: 10]`,
				`Syntax Error: Incompatible parameter types - "TMyEnum" expected (instead of "String") [line: 8, column: 10]`,
				`Syntax Error: Incompatible parameter types - "TMyEnum" expected (instead of "void") [line: 12, column: 10]`,
			},
		},
		{
			// SetOfFail/test_non_variable: both mutator forms need a writable
			// receiver, and a function result is not one. The fixture also
			// expects the `Result is never used` hint first; the assignment here
			// keeps that hint out of the way of the receiver sentences.
			name: "function result as a set mutator receiver",
			source: "type TMyEnum = (enumOne, enumTwo);\ntype TMySet = set of TMyEnum;\n\n" +
				"function Test : TMySet; begin Result := []; end;\n\nInclude(Test, enumOne);\n\nTest.Exclude(enumTwo);\n",
			want: []string{
				`Syntax Error: Variable expected [line: 6, column: 9]`,
				`Syntax Error: Variable expected [line: 8, column: 6]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			got := result.DiagnosticStrings()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

// TestCompile_SetMutatorsStayClean keeps the SetOfFail sentences off the well
// formed calls: both the procedure form and the method form of Include and
// Exclude compile without a diagnostic.
func TestCompile_SetMutatorsStayClean(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "procedure form",
			source: "type TMyEnum = (enumOne, enumTwo);\ntype TMySet = set of TMyEnum;\n\nvar s : TMySet;\nvar e : TMyEnum := enumOne;\n\nInclude(s, e);\nExclude(s, e);\n",
		},
		{
			name:   "method form",
			source: "type TMyEnum = (enumOne, enumTwo);\ntype TMySet = set of TMyEnum;\n\nvar s : TMySet;\nvar e : TMyEnum := enumOne;\n\ns.Include(e);\ns.Exclude(e);\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			if got := result.DiagnosticStrings(); len(got) != 0 {
				t.Fatalf("expected no diagnostics, got %q", got)
			}
		})
	}
}
