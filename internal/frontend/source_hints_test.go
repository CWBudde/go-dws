package frontend

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func caseHintLines(result *Result) []int {
	var lines []int
	for _, diagnostic := range result.Diagnostics {
		if strings.Contains(diagnostic.Message, "does not match case of declaration") {
			lines = append(lines, diagnostic.Line)
		}
	}
	return lines
}

func TestCompile_SourceHintLevels(t *testing.T) {
	const source = `procedure Test; begin end;
test;
{$HINTS OFF}
test;
{$HINTS PEDANTIC}
test;
{$HINTS NORMAL}
test;
{$HINTS STRICT}
test;
{$HINTS OFF}
{$HINTS ON}
test;
{$HINTS PEDANTIC}
test;
{$IFDEF UNDEFINED}
{$HINTS OFF}
{$ENDIF}
test;`
	for _, initial := range []semantic.HintsLevel{semantic.HintsLevelDisabled, semantic.HintsLevelNormal, semantic.HintsLevelPedantic} {
		t.Run(string(rune('0'+initial)), func(t *testing.T) {
			result := Compile(source, "hints.pas", initial)
			if result.HasFatalDiagnostics() {
				t.Fatal(result.DiagnosticStrings())
			}
			want := []int{6, 15, 19}
			if initial == semantic.HintsLevelPedantic {
				want = []int{2, 6, 13, 15, 19}
			}
			if got := caseHintLines(result); !reflect.DeepEqual(got, want) {
				t.Fatalf("case hints on lines %v, want %v: %v", got, want, result.DiagnosticStrings())
			}
		})
	}
}

func TestCompile_SourceHintControlsInclude(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hints.inc"), []byte("test;\n{$HINTS PEDANTIC}\ntest;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := "procedure Test; begin end;\n{$HINTS OFF}\n{$I 'hints.inc'}\ntest;"
	result := Compile(source, filepath.Join(dir, "main.pas"), semantic.HintsLevelPedantic)
	if result.HasFatalDiagnostics() {
		t.Fatal(result.DiagnosticStrings())
	}
	if got, want := caseHintLines(result), []int{3, 4}; !reflect.DeepEqual(got, want) {
		t.Fatalf("case hints on lines %v, want %v: %v", got, want, result.DiagnosticStrings())
	}
}

func TestCompile_SourceHintControlsUnit(t *testing.T) {
	dir := t.TempDir()
	unit := "unit U; interface procedure Test; implementation procedure Test; begin end;\n{$HINTS OFF}\ninitialization test; end."
	if err := os.WriteFile(filepath.Join(dir, "U.pas"), []byte(unit), 0o644); err != nil {
		t.Fatal(err)
	}
	result := Compile("uses U;\ntest;", filepath.Join(dir, "main.pas"), semantic.HintsLevelPedantic)
	if result.HasFatalDiagnostics() {
		t.Fatal(result.DiagnosticStrings())
	}
	if got, want := caseHintLines(result), []int{2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("case hints on lines %v, want %v: %v", got, want, result.DiagnosticStrings())
	}
}

func TestCompile_SourceHintsSuppressSemanticHints(t *testing.T) {
	source := "{$HINTS OFF}\nprocedure Test; stdcall; begin end;\n{$HINTS NORMAL}\nprocedure Again; stdcall; begin end;\nTest; Again;"
	result := Compile(source, "hints.pas", semantic.HintsLevelNormal)
	if result.HasFatalDiagnostics() {
		t.Fatal(result.DiagnosticStrings())
	}
	hints := result.HintStrings()
	if len(hints) != 1 || !strings.Contains(hints[0], "line: 4") {
		t.Fatalf("expected only enabled call convention hint, got %v", hints)
	}
}

func TestCompile_SourceHintsPreservedAcrossParseAnalyze(t *testing.T) {
	source := "procedure Test; begin end;\n{$HINTS OFF}\ntest;\n{$HINTS PEDANTIC}\ntest;"
	opts := Options{HintsLevel: semantic.HintsLevelNormal}
	result := AnalyzeParsed(ParseWithOptions(source, opts), source, opts)
	if result.HasFatalDiagnostics() {
		t.Fatal(result.DiagnosticStrings())
	}
	if got, want := caseHintLines(result), []int{5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("case hints on lines %v, want %v: %v", got, want, result.DiagnosticStrings())
	}
}

func TestCompile_SourceHintsInvalidSwitchKeepsSetting(t *testing.T) {
	source := "procedure Test; begin end;\n{$HINTS OFF}\n{$HINTS INVALID}\ntest;\n{$HINTS PEDANTIC}\ntest;"
	result := Compile(source, "hints.pas", semantic.HintsLevelPedantic)
	if !result.HasFatalDiagnostics() {
		t.Fatal("expected invalid switch error")
	}
	if got, want := caseHintLines(result), []int{6}; !reflect.DeepEqual(got, want) {
		t.Fatalf("case hints on lines %v, want %v: %v", got, want, result.DiagnosticStrings())
	}
}

func TestCompile_ExplicitHintsUseConfiguredDefault(t *testing.T) {
	source := "{$HINT 'initial'}\n{$HINTS OFF}\n{$HINT 'off'}\n{$HINTS ON}\n{$HINT 'restored'}\n{$HINTS NORMAL}\n{$HINT 'explicit'}"
	for _, initial := range []semantic.HintsLevel{semantic.HintsLevelDisabled, semantic.HintsLevelNormal} {
		result := Compile(source, "hints.pas", initial)
		if result.HasFatalDiagnostics() {
			t.Fatal(result.DiagnosticStrings())
		}
		want := 1
		if initial == semantic.HintsLevelNormal {
			want = 3
		}
		if got := result.HintStrings(); len(got) != want {
			t.Fatalf("initial %v: hints %v, want %d", initial, got, want)
		}
	}
}

func TestCompile_ExplicitUnitHintsUseConfiguredDefault(t *testing.T) {
	dir := t.TempDir()
	unit := "unit U;\n{$HINT 'initial'}\n{$HINTS OFF}\n{$HINT 'off'}\n{$HINTS ON}\n{$HINT 'restored'}\n{$HINTS NORMAL}\n{$HINT 'explicit'}\ninterface implementation end."
	if err := os.WriteFile(filepath.Join(dir, "U.pas"), []byte(unit), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, initial := range []semantic.HintsLevel{semantic.HintsLevelDisabled, semantic.HintsLevelNormal} {
		result := Compile("uses U;", filepath.Join(dir, "main.pas"), initial)
		if result.HasFatalDiagnostics() {
			t.Fatal(result.DiagnosticStrings())
		}
		want := 1
		if initial == semantic.HintsLevelNormal {
			want = 3
		}
		if got := result.HintStrings(); len(got) != want {
			t.Fatalf("initial %v: hints %v, want %d", initial, got, want)
		}
	}
}

func TestCompile_UnitDirectiveErrorsStillStopWithHintsDisabled(t *testing.T) {
	for _, directive := range []string{"ERROR", "FATAL"} {
		t.Run(directive, func(t *testing.T) {
			dir := t.TempDir()
			unit := "unit U; {$HINTS OFF} {$" + directive + " 'unit failed'} interface implementation end."
			if err := os.WriteFile(filepath.Join(dir, "U.pas"), []byte(unit), 0o644); err != nil {
				t.Fatal(err)
			}
			result := Compile("uses U;", filepath.Join(dir, "main.pas"), semantic.HintsLevelDisabled)
			if !result.HasFatalDiagnostics() {
				t.Fatalf("expected fatal unit directive: %v", result.DiagnosticStrings())
			}
			if !strings.Contains(strings.Join(result.DiagnosticStrings(), "\n"), "unit failed") {
				t.Fatalf("lost unit error: %v", result.DiagnosticStrings())
			}
		})
	}
}
