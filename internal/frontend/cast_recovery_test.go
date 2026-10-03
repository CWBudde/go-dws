package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_CastRecoveryFixtures(t *testing.T) {
	for _, name := range []string{"as_error", "object_relops", "as_invalid_right"} {
		t.Run(name, func(t *testing.T) {
			root := filepath.Join("..", "..", "testdata", "fixtures", "FailureScripts", name)
			source, err := os.ReadFile(root + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(root + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			assertDiagnostics(t, string(source), name+".pas", strings.Split(strings.TrimSpace(string(want)), "\n"))
		})
	}
}

func TestCompile_CastRecoveryCategories(t *testing.T) {
	for _, tt := range []struct {
		source string
		want   []string
	}{
		{"var o: TObject;\nPrintLn(1=o);", []string{`Syntax Error: Object expected [line: 2, column: 10]`}},
		{"var c: TClass;\nPrintLn(1=c);", []string{`Syntax Error: Class reference expected [line: 2, column: 10]`}},
		{"var i: IInterface;\nPrintLn(1=i);", []string{`Syntax Error: Interface expected [line: 2, column: 10]`}},
		{"var c: TClass;\nPrintLn(c as 42);\nPrintLn(Missing);", []string{`Syntax Error: Class reference expected [line: 2, column: 11]`}},
		{"var o: TObject;\nPrintLn(o as Missing);\nPrintLn(AnotherMissing);", []string{`Syntax Error: Unknown name "Missing" [line: 2, column: 14]`}},
	} {
		t.Run(tt.source, func(t *testing.T) { assertDiagnostics(t, tt.source, "<test>", tt.want) })
	}
}

func TestCompile_CastTargetUnknownChildrenStop(t *testing.T) {
	for _, operator := range []string{"as", "is"} {
		t.Run(operator+" qualified", func(t *testing.T) {
			assertDiagnostics(t, "var o: TObject;\nPrintLn(o "+operator+" Missing.Target);\nPrintLn(Later);", "<test>", []string{`Syntax Error: Unknown name "Missing" [line: 2, column: 14]`})
		})
		t.Run(operator+" call argument", func(t *testing.T) {
			source := "var o: TObject;\nfunction Target(x: Integer): TClass; begin Result := TObject; end;\nPrintLn(o " + operator + " Target(Missing));\nPrintLn(Later);"
			assertDiagnostics(t, source, "<test>", []string{`Syntax Error: Unknown name "Missing" [line: 3, column: 21]`})
		})
	}
}

func TestCompile_CastQualifiedValueShadow(t *testing.T) {
	for _, namespace := range []string{"System", "Internal"} {
		for _, source := range []string{
			"type THolder = record TObject: Integer; end;\nvar System: THolder;\nvar Obj: TObject;\nPrintLn(Obj is System.TObject);\nPrintLn(Obj as System.TObject);",
			"type THolder = record TObject: Integer; end;\nprocedure Test(System: THolder);\nbegin var Obj: TObject;\nPrintLn(Obj is System.TObject);\nPrintLn(Obj as System.TObject);\nend;\nvar Holder: THolder; Test(Holder);",
		} {
			assertDiagnostics(t, strings.ReplaceAll(source, "System", namespace), "<test>", []string{
				`Syntax Error: Class reference expected [line: 4, column: 13]`,
				`Syntax Error: Class reference expected [line: 5, column: 13]`,
			})
		}
	}

}

func TestCompile_CastImportedNamespaceAndValueShadow(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Shapes.pas"), []byte("unit Shapes; interface type TTarget = class end; implementation end."), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"namespace", "uses Shapes;\nvar Obj: TObject;\nPrintLn(Obj is Shapes.TTarget);\nPrintLn(Obj as Shapes.TTarget);", nil},
		{"parameter shadows namespace", "uses Shapes;\ntype THolder = record TTarget: Integer; end;\nprocedure Test(Shapes: THolder);\nbegin var Obj: TObject;\nPrintLn(Obj is Shapes.TTarget);\nPrintLn(Obj as Shapes.TTarget);\nend;\nvar Holder: THolder; Test(Holder);", []string{
			`Syntax Error: Class reference expected [line: 5, column: 13]`,
			`Syntax Error: Class reference expected [line: 6, column: 13]`,
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, filepath.Join(dir, "Main.pas"), semantic.HintsLevelDisabled)
			got := result.DiagnosticStrings()
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Fatalf("diagnostics = %v; want %v", got, tt.want)
			}
		})
	}
}
