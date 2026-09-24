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

func TestCompileWithOptions_ExcludesDriverAndPassesDefinesToUnit(t *testing.T) {
	dir := t.TempDir()
	driver := filepath.Join(dir, "Same.dws")
	unit := filepath.Join(dir, "Same.pas")
	source := "uses Same; {$ifdef CONDITION} PrintLn(Answer); {$else} Missing; {$endif}"
	unitSource := "unit Same; interface {$ifdef CONDITION} const Answer = 7; {$endif} implementation end."
	for path, contents := range map[string]string{driver: source, unit: unitSource} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	result := CompileWithOptions(source, Options{Filename: driver, UnitSearchPaths: []string{dir}, Defines: []string{"CONDITION"}})
	if result.HasFatalDiagnostics() || !result.SemanticSuccessful {
		t.Fatalf("driver and Pascal unit did not compile: %v", result.DiagnosticStrings())
	}
	loaded, ok := result.UnitRegistry.GetUnit("Same")
	if !ok || loaded.FilePath != unit {
		t.Fatalf("loaded unit = %#v, want %s", loaded, unit)
	}
	if clone := result.UnitRegistry.CloneForExecution(); clone == nil {
		t.Fatal("execution registry unavailable")
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

// TestCompile_UnitConstantErrorKeepsItsPosition checks that a malformed string or char
// constant in a used unit is reported as the normalized constant diagnostic with its own
// position, not wrapped in the positionless "compiler directive error" load failure. The
// lexer surfaces constant errors through the directive channel, so the unit loader has to
// tell the two apart.
func TestCompile_UnitConstantErrorKeepsItsPosition(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"BadConst.pas": "unit BadConst;\ninterface\nfunction Foo: String;\nimplementation\nfunction Foo: String;\nbegin\n  Result := 'unterminated;\nend;\nend.\n",
		"BadDir.pas":   "unit BadDir;\n{$ERROR 'nope'}\ninterface\nimplementation\nend.\n",
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct{ name, source, want string }{
		{
			name:   "malformed constant is positioned",
			source: "uses BadConst;\nPrintLn(Foo);",
			want:   `Syntax Error: End of string constant not found (end of line) [line: 7, column: 13]`,
		},
		{
			// A real directive still fails the load, as {$FATAL} truncation requires.
			name:   "directive error still fails the load",
			source: "uses BadDir;\nPrintLn(1);",
			want:   `compiler directive error in unit "BadDir": nope`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := Compile(tc.source, filepath.Join(dir, "Main.pas"), semantic.HintsLevelPedantic)
			got := res.DiagnosticStrings()
			if len(got) != 1 || got[0] != tc.want {
				t.Fatalf("diagnostics = %q, want [%q]", got, tc.want)
			}
			if !res.HasFatalDiagnostics() {
				t.Fatalf("expected a fatal diagnostic")
			}
		})
	}
}
