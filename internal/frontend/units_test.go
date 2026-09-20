package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_UnitDependencies(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"Numbers.pas": `unit Numbers; interface type TNumber = Integer; const Offset = 2; function Twice(x: TNumber): TNumber; implementation function Twice(x: TNumber): TNumber; begin Result := x * 2; end; end.`,
		"Facade.pas":  `unit Facade; interface function Answer: Integer; implementation uses Numbers; var Secret: Integer := 1; function Answer: Integer; begin Result := Twice(20) + Offset; end; end.`,
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct{ name, source, want string }{
		{"dependency order", `uses Facade; PrintLn(Answer());`, ""},
		{"exported types and constants", `uses Numbers; var n: TNumber := Offset; PrintLn(Twice(n));`, ""},
		{"qualified function", `uses Numbers; PrintLn(Numbers.Twice(21));`, ""},
		{"program type error", `uses Numbers; var n: Integer := 'bad';`, "Cannot assign String to Integer"},
		{"private declaration", `uses Facade; PrintLn(Secret);`, `Unknown name "Secret"`},
		{"unimported transitive declaration", `uses Facade; PrintLn(Offset);`, `Unknown name "Offset"`},
		{"missing unit", `uses Missing;`, "cannot load unit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := Compile(tc.source, filepath.Join(dir, "Main.pas"), semantic.HintsLevelDisabled)
			got := strings.Join(res.DiagnosticStrings(), "\n")
			if tc.want == "" {
				if !res.SemanticSuccessful || res.HasFatalDiagnostics() {
					t.Fatalf("compile failed: %s", got)
				}
			} else if !res.HasFatalDiagnostics() || !strings.Contains(got, tc.want) {
				t.Fatalf("want error %q, got %s", tc.want, got)
			}
		})
	}
}

func TestCompile_UnitBodyError(t *testing.T) {
	dir := t.TempDir()
	source := `unit Broken; interface function Value: Integer; implementation function Value: Integer; begin Result := 'bad'; end; end.`
	if err := os.WriteFile(filepath.Join(dir, "Broken.pas"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	res := Compile(`uses Broken; PrintLn(Value());`, filepath.Join(dir, "Main.pas"), semantic.HintsLevelDisabled)
	if !res.HasFatalDiagnostics() || !strings.Contains(strings.Join(res.DiagnosticStrings(), "\n"), `Cannot assign "String" to "Integer"`) {
		t.Fatalf("unit body error missing: %v", res.DiagnosticStrings())
	}
}

func TestCompileWithOptions_UnitSearchPathsAndIncludes(t *testing.T) {
	dir := t.TempDir()
	for name, source := range map[string]string{
		"U.dws":   "unit U; interface {$INCLUDE api.inc} implementation end.",
		"api.inc": "const Answer = 42;",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	res := CompileWithOptions("uses U; PrintLn(Answer);", Options{UnitSearchPaths: []string{dir}, HintsLevel: semantic.HintsLevelDisabled})
	if res.HasFatalDiagnostics() || !res.SemanticSuccessful {
		t.Fatalf("compile failed: %v", res.DiagnosticStrings())
	}
	if err := os.Remove(filepath.Join(dir, "api.inc")); err != nil {
		t.Fatal(err)
	}
	res = CompileWithOptions("uses U;", Options{UnitSearchPaths: []string{dir}})
	if !res.HasFatalDiagnostics() {
		t.Fatalf("missing include should fail: %v", res.DiagnosticStrings())
	}
}

func TestCompile_UnitCircularDependency(t *testing.T) {
	dir := t.TempDir()
	for name, source := range map[string]string{
		"A.dws": "unit A; interface uses B; implementation end.",
		"B.dws": "unit B; interface uses A; implementation end.",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	res := CompileWithOptions("uses A;", Options{UnitSearchPaths: []string{dir}})
	if !res.HasFatalDiagnostics() || !strings.Contains(strings.Join(res.DiagnosticStrings(), "\n"), "circular dependency") {
		t.Fatalf("expected cycle diagnostic: %v", res.DiagnosticStrings())
	}
}

// TestCompile_UsesUnitUnimplementedForwards checks that a unit loaded through
// a program's uses clause reports its unimplemented interface routines and
// implementation-section forwards with DWScript's forward sentence, after the
// unit's other errors, in name order (overloads latest first), exempting
// externals (fixture BuildScripts/sections_test).
func TestCompile_UsesUnitUnimplementedForwards(t *testing.T) {
	dir := t.TempDir()
	source := `unit Fwd;

interface

procedure Zeta;
procedure Test; overload;
procedure Test(s : String); overload;
function Test(i : Integer) : String; overload;
function Ext(v : Variant) : Integer; external;

implementation

procedure Hidden; forward;

procedure Test;
var i : Integer;
begin
   i := 'bad';
end;

end.`
	if err := os.WriteFile(filepath.Join(dir, "Fwd.pas"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	res := Compile(`uses Fwd;`, filepath.Join(dir, "Main.pas"), semantic.HintsLevelDisabled)
	got := res.DiagnosticStrings()
	want := []string{
		`Cannot assign "String" to "Integer"`,
		`The function "Hidden" was forward declared but not implemented [line: 13, column: 11`,
		`The function "Test" was forward declared but not implemented [line: 8, column: 10`,
		`The function "Test" was forward declared but not implemented [line: 7, column: 11`,
		`The function "Zeta" was forward declared but not implemented [line: 5, column: 11`,
	}
	if len(got) != len(want) {
		t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
	}
	for i := range want {
		if !strings.Contains(got[i], want[i]) {
			t.Fatalf("diagnostic %d mismatch\n got: %q\nwant: %q", i, got, want)
		}
	}
}
