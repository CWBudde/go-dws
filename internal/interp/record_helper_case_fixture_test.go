package interp

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestRecordHelperCase_Fixture(t *testing.T) {
	// The expectation includes both X/x hints in upstream revision
	// 1dbf8a90329cc3f2638516e89c0668f916c1ddb9. Keep them alongside the
	// output assertions: record fields still resolve regardless of casing.
	const category = "HelpersPass"
	path := filepath.Join(fixturesRoot, category, "record_array_helper.pas")
	if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel(category)); got != testResultPassed {
		t.Fatalf("record_array_helper: %v: %s", got, detail)
	}
}

func TestRecordHelperCase_DirectFieldAccess(t *testing.T) {
	const source = `type TPoint = record x: Integer; end;
var point: TPoint;
point.X := 7;
PrintLn(point.X);
PrintLn(point.x);`
	compiled := frontend.Compile(source, "record_case.pas", semantic.HintsLevelPedantic)
	if compiled.HasFatalDiagnostics() || !compiled.SemanticSuccessful {
		t.Fatalf("compile diagnostics: %v", compiled.DiagnosticStrings())
	}
	want := []string{
		`Hint: "X" does not match case of declaration ("x") [line: 3, column: 7]`,
		`Hint: "X" does not match case of declaration ("x") [line: 4, column: 15]`,
	}
	if got := compiled.DiagnosticStrings(); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("diagnostics = %v, want %v", got, want)
	}
	var output bytes.Buffer
	engine := New(&output)
	engine.SetSemanticInfo(compiled.SemanticInfo)
	if result := engine.Eval(compiled.Program); result != nil && result.Type() == "ERROR" {
		t.Fatalf("runtime error: %s", result.String())
	}
	if got := output.String(); got != "7\n7\n" {
		t.Fatalf("output = %q, want %q", got, "7\n7\n")
	}
}
