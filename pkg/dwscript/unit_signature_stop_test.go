package dwscript

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEngineCompile_UnitSignatureStopRetainsEarlierBodies(t *testing.T) {
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
			engine, err := New(WithUnitSearchPaths(dir))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile("uses Probe;")
			if program != nil {
				t.Fatal("Compile returned a Program for stopped source")
			}
			compileErr, ok := err.(*CompileError)
			if !ok {
				t.Fatalf("error = %T %v, want CompileError", err, err)
			}
			var got []string
			for _, diag := range compileErr.Errors {
				if diag.Severity != SeverityError {
					t.Errorf("unexpected non-error diagnostic: %v", diag)
				}
				got = append(got, fmt.Sprintf("Syntax Error: %s [line: %d, column: %d]", diag.Message, diag.Line, diag.Column))
			}
			want := []string{
				fmt.Sprintf(`Syntax Error: Cannot cast "Integer" as "String" [line: %d, column: 11]`, 4+section.lines),
				fmt.Sprintf(`Syntax Error: Cannot cast "Integer" as "String" [line: %d, column: 11]`, 8+section.lines),
				fmt.Sprintf(`Syntax Error: Unknown name "Missing" [line: %d, column: 30]`, 10+section.lines),
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("diagnostics = %q, want %q", got, want)
			}
		})
	}
}

func TestEngineCompile_SemanticStopCompletionHints(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"stopped routine", "procedure P;\nbegin\nvar unused: Integer;\nMissing;\nPrintLn(unused);\nend;", []string{`Unknown name "Missing" [line: 4, column: 1]`}},
		{"completed earlier routine", "procedure P;\nbegin\nvar unused: Integer;\nend;\nMissing;", []string{`Hint: Variable "unused" declared but not used [line: 3, column: 5]`, `Unknown name "Missing" [line: 5, column: 1]`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := New()
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(tt.source)
			if program != nil {
				t.Fatal("Compile returned a Program for stopped source")
			}
			compileErr, ok := err.(*CompileError)
			if !ok {
				t.Fatalf("error = %T %v, want CompileError", err, err)
			}
			var got []string
			for _, diag := range compileErr.Errors {
				got = append(got, fmt.Sprintf("%s [line: %d, column: %d]", diag.Message, diag.Line, diag.Column))
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("diagnostics = %q, want %q", got, tt.want)
			}
		})
	}
}
