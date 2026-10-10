package frontend

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_UnitSignatureStopRetainsEarlierBodies(t *testing.T) {
	for _, section := range []struct {
		name, source string
		lines        int
	}{
		{"sectionless", "", 0},
		{"implementation", "interface\nimplementation\n", 2},
	} {
		t.Run(section.name, func(t *testing.T) {
			dir := t.TempDir()
			source := "unit Probe;\n" + section.source + "procedure First;\nbegin\nPrintLn(1 as String);\nend;\nprocedure Second;\nbegin\nPrintLn(2 as String);\nend;\nprocedure Later(x: Integer = Missing);\nbegin UnreadBody; end;\nvar unread := UnreadDeclaration;\nend."
			if err := os.WriteFile(filepath.Join(dir, "Probe.pas"), []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			result := Compile("uses Probe;", filepath.Join(dir, "Main.pas"), semantic.HintsLevelPedantic)
			want := []string{
				fmt.Sprintf(`Syntax Error: Cannot cast "Integer" as "String" [line: %d, column: 11]`, 4+section.lines),
				fmt.Sprintf(`Syntax Error: Cannot cast "Integer" as "String" [line: %d, column: 11]`, 8+section.lines),
				fmt.Sprintf(`Syntax Error: Unknown name "Missing" [line: %d, column: 30]`, 10+section.lines),
			}
			if got := result.DiagnosticStrings(); !reflect.DeepEqual(got, want) {
				t.Fatalf("diagnostics = %q, want %q", got, want)
			}
			for i, diag := range result.Diagnostics {
				if diag.Stop != (i == len(want)-1) {
					t.Errorf("diagnostic %d Stop = %v", i, diag.Stop)
				}
			}
		})
	}
}

func TestCompile_SemanticStopCompletionHints(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"routine", "procedure P;\nbegin\nvar unused: Integer;\nMissing;\nPrintLn(unused);\nend;", []string{`Syntax Error: Unknown name "Missing" [line: 4, column: 1]`}},
		{"nested block", "procedure P;\nbegin\nbegin\nvar unused: Integer;\nMissing;\nPrintLn(unused);\nend;\nend;", []string{`Syntax Error: Unknown name "Missing" [line: 5, column: 1]`}},
		{"method", "type TTest = class\nprocedure P;\nbegin\nvar unused: Integer;\nMissing;\nPrintLn(unused);\nend;\nend;", []string{`Syntax Error: Unknown name "Missing" [line: 5, column: 1]`}},
		{"lambda", "var f := lambda(): Integer\nbegin\nvar unused: Integer;\nMissing;\nResult := unused;\nend;", []string{`Syntax Error: Unknown name "Missing" [line: 4, column: 1]`}},
		{"completed earlier routine", "procedure P;\nbegin\nvar unused: Integer;\nend;\nMissing;", []string{`Hint: Variable "unused" declared but not used [line: 3, column: 5]`, `Syntax Error: Unknown name "Missing" [line: 5, column: 1]`}},
		{"completed earlier block", "procedure P;\nbegin\nbegin\nvar unused: Integer;\nend;\nMissing;\nend;", []string{`Hint: Variable "unused" declared but not used [line: 4, column: 5]`, `Syntax Error: Unknown name "Missing" [line: 6, column: 1]`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			if got := result.DiagnosticStrings(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("diagnostics = %q, want %q", got, tt.want)
			}
		})
	}
}
