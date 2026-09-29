package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// Compare complete diagnostics through the compile path used by the CLI,
// including child errors, recovery, type captions and scanner positions.
func TestCompile_IncompatibleTypesDiagnostics(t *testing.T) {
	for _, name := range []string{
		"array_initialization4", "coalesce_dynarray", "const_1",
		"array_of_proc", "array_of_proc2", "const_procedure_array",
	} {
		t.Run(name, func(t *testing.T) {
			assertDiagnostics(t, fixtureSource(t, name+".pas"), name+".pas", fixtureExpectation(t, name+".txt"))
		})
	}
}

func TestCompile_IncompatibleRoutineSetOperand(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "SetOfFail")
	source, err := os.ReadFile(filepath.Join(root, "invalid_operand.pas"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(root, "invalid_operand.txt"))
	if err != nil {
		t.Fatal(err)
	}
	assertDiagnostics(t, string(source), "invalid_operand.pas", strings.Split(strings.TrimSpace(string(want)), "\n"))
}

func TestCompile_ArrayLiteralStops(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{
			name:   "first token scanned",
			source: "procedure Test(a: array of Float); begin end;\nTest([Test, 2]);\nUnknown;\n",
			want: []string{
				`Syntax Error: More arguments expected [line: 2, column: 7]`,
				`Syntax Error: Incompatible types: "void" and "nil" [line: 2, column: 11]`,
			},
		},
		{
			name:   "cursor before newline",
			source: "procedure Test(a: array of Float); begin end;\nTest([1,\n  Test]);\nUnknown;\n",
			want: []string{
				`Syntax Error: More arguments expected [line: 3, column: 3]`,
				`Syntax Error: Incompatible types: "void" and "Integer" [line: 2, column: 9]`,
			},
		},
		{
			name:   "array inferred from identifier-only literal",
			source: "procedure Test(a: Integer); begin end;\nvar x := [Test];\n",
			want: []string{
				`Syntax Error: More arguments expected [line: 2, column: 11]`,
				`Syntax Error: Incompatible types: "void" and "nil" [line: 2, column: 15]`,
			},
		},
		{
			name:   "later parser diagnostic",
			source: "const Data = [[1, 2], [3]];\nvar x := ;\n",
			want: []string{
				`Syntax Error: Incompatible types: "array [0..0] of Integer" and "array [0..1] of Integer" [line: 1, column: 22]`,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertDiagnostics(t, tt.source, "<test>", tt.want)
		})
	}
}

func TestCompile_CompatibleIncompatibleTypesContexts(t *testing.T) {
	for _, source := range []string{
		"const Data = [[1, 2], [3, 4]]; PrintLn(Data[1, 1]);",
		"var a: array of Variant := [1, 'text', 2.5]; PrintLn(a[0]);",
		"function Value: Integer; begin Result := 1; end; var a: array of Integer := [Value]; PrintLn(a[0]);",
		"procedure Test; begin end; procedure Accept(a: array of procedure); begin end; Accept([Test]);",
		"var f: Float := 1.5; var i: Integer := 2; PrintLn(f ?? i);",
	} {
		result := Compile(source, "<test>", semantic.HintsLevelPedantic)
		if got := result.DiagnosticStrings(); len(got) != 0 {
			t.Errorf("source %q produced diagnostics: %q", source, got)
		}
	}
}

func TestCompile_CoalesceArrayNil(t *testing.T) {
	assertDiagnostics(t, "var a: array of Integer; var b := a ?? nil; PrintLn(b.Length);", "<test>", nil)
}

func TestCompile_CoalesceArrayElementConversions(t *testing.T) {
	for _, tt := range []struct {
		name, source, want string
	}{
		{
			"numeric elements",
			"var a: array of Float := [1.5];\nvar b: array of Integer := [2];\nvar c := a ?? b; PrintLn(c[0]);",
			`Syntax Error: Incompatible types: "array of Float" and "array of Integer" [line: 3, column: 12]`,
		},
		{
			"variant elements",
			"var a: array of Float := [1.5];\nvar b: array of Variant := [2];\nvar c := a ?? b; PrintLn(c[0]);",
			`Syntax Error: Incompatible types: "array of Float" and "array of Variant" [line: 3, column: 12]`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertDiagnostics(t, tt.source, "<test>", []string{tt.want})
		})
	}
}

func TestCompile_CompatibleNestedArrayNumericTypes(t *testing.T) {
	assertDiagnostics(t, "const Data = [[1, 2], [3.5, 4.5]]; PrintLn(Data[1, 1]);", "<test>", nil)
}

func TestCompile_ExplicitRoutineReferencesInArrays(t *testing.T) {
	for _, tt := range []struct {
		name, source string
	}{
		{"procedure", "procedure Foo; begin end; var a := [@Foo]; PrintLn(Length(a));"},
		{"function", "function Foo: Integer; begin Result := 42; end; var a := [@Foo]; PrintLn(a[0]());"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertDiagnostics(t, tt.source, "<test>", nil)
		})
	}
}

func TestCompile_ImportedUnitArrayStops(t *testing.T) {
	for _, tt := range []struct {
		name, declarations string
		want               []string
	}{
		{
			"earlier body error retained",
			"procedure A; begin var x: Integer := 3.4; end;\nconst Data = [[1, 2], [3]];",
			[]string{"Cannot assign", "Incompatible types:"},
		},
		{
			"later declaration error omitted",
			"procedure A; begin var a := [[1, 2], [3]]; end;\nvar x: Integer := 3.4;",
			[]string{"Incompatible types:"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			source := "unit U;\ninterface\nprocedure A;\nimplementation\n" + tt.declarations + "\nend."
			if err := os.WriteFile(filepath.Join(dir, "U.pas"), []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			result := Compile("uses U;", filepath.Join(dir, "Main.pas"), semantic.HintsLevelDisabled)
			got := result.DiagnosticStrings()
			if len(got) != len(tt.want) {
				t.Fatalf("got %q, want diagnostics containing %q in order", got, tt.want)
			}
			for i, want := range tt.want {
				if !strings.Contains(got[i], want) {
					t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want)
				}
			}
		})
	}
}
