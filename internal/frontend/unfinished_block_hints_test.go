package frontend

import (
	"slices"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// Upstream makes a block's unused-symbol hints only on leaving it: ReadProcBody
// calls HintUnusedSymbols after its CPE_EndOfBlockExpected stop
// (dwsCompiler.pas:4154-4157), and ReadBlock after its END (4540). A block whose
// END is never read stops the compile before that point, so it gets no hints,
// while a routine completed before the stop keeps its own.
func TestCompile_UnfinishedBlockDropsCompletionHints(t *testing.T) {
	const unusedU = "Hint: Variable \"u\" declared but not used [line: 2, column: 5]"
	const unusedI = "Hint: Variable \"i\" declared but not used"
	compile := func(source string) []string {
		return CompileWithOptions(source, Options{HintsLevel: semantic.HintsLevelPedantic}).DiagnosticStrings()
	}
	hasHintFor := func(diags []string, prefix string) bool {
		return slices.ContainsFunc(diags, func(d string) bool { return strings.HasPrefix(d, prefix) })
	}

	t.Run("block_unfinished2 fixture", func(t *testing.T) {
		got := compile(fixtureSource(t, "block_unfinished2.pas"))
		if want := fixtureExpectation(t, "block_unfinished2.txt"); !slices.Equal(got, want) {
			t.Fatalf("got %q; want %q", got, want)
		}
	})
	t.Run("unfinished nested block", func(t *testing.T) {
		got := compile("procedure Test;\nbegin\nbegin\nvar i : Integer;\n")
		if hasHintFor(got, unusedI) {
			t.Fatalf("got %q; want no hint for the unfinished block's local", got)
		}
	})
	t.Run("completed routine before unfinished routine", func(t *testing.T) {
		got := compile("procedure P;\nvar u : Integer;\nbegin\nend;\nprocedure Test;\nvar i : Integer;\nbegin\n")
		if !slices.Contains(got, unusedU) || hasHintFor(got, unusedI) {
			t.Fatalf("got %q; want the completed routine's hint only", got)
		}
	})
	t.Run("completed routine without stop", func(t *testing.T) {
		if got, want := compile("procedure P;\nvar u : Integer;\nbegin\nend;\n"), []string{unusedU}; !slices.Equal(got, want) {
			t.Fatalf("got %q; want %q", got, want)
		}
	})
}
