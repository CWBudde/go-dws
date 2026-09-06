package frontend

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompileWithOptions_TypeCheckOffSkipsAnalyzer(t *testing.T) {
	res := CompileWithOptions("var x: Integer := 'hello';", Options{SkipTypeCheck: true})
	if res.Analyzer != nil || res.SemanticAttempted {
		t.Fatalf("expected no semantic analysis, got attempted=%v", res.SemanticAttempted)
	}
	if res.HasFatalDiagnostics() {
		t.Fatalf("unexpected diagnostics: %v", res.DiagnosticStrings())
	}
	if res.Program == nil {
		t.Fatal("expected a parsed program")
	}
}

func TestCompileWithOptions_IncludeDirControlsIncludes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "inc.inc"), []byte("PrintLn('x');"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := "{$INCLUDE 'inc.inc'}"
	with := CompileWithOptions(src, Options{Filename: "<eval>", IncludeDir: dir, HintsLevel: semantic.HintsLevelNormal})
	if with.HasFatalDiagnostics() {
		t.Fatalf("include should resolve: %v", with.DiagnosticStrings())
	}
	if got := len(with.Program.Statements); got != 1 {
		t.Fatalf("expected the included statement, got %d statements", got)
	}
	// Without a resolver the lexer drops the directive (pre-existing lexer behaviour):
	// the included content must not appear.
	without := CompileWithOptions(src, Options{Filename: "<eval>", HintsLevel: semantic.HintsLevelNormal})
	if got := len(without.Program.Statements); got != 0 {
		t.Fatalf("empty IncludeDir must disable include resolution, got %d statements", got)
	}
}

func TestParseThenAnalyzeEqualsCompile(t *testing.T) {
	src := "var x: Integer := 'hello';"
	opts := Options{HintsLevel: semantic.HintsLevelPedantic}
	twoStep := AnalyzeParsed(ParseWithOptions(src, opts), src, opts)
	oneStep := CompileWithOptions(src, opts)
	legacy := Compile(src, "", semantic.HintsLevelPedantic)
	if got, want := twoStep.DiagnosticStrings(), oneStep.DiagnosticStrings(); !reflect.DeepEqual(got, want) {
		t.Fatalf("two-step %v != one-step %v", got, want)
	}
	if got, want := legacy.DiagnosticStrings(), oneStep.DiagnosticStrings(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Compile %v != CompileWithOptions %v", got, want)
	}
	if len(oneStep.DiagnosticStrings()) == 0 {
		t.Fatal("expected a type-mismatch diagnostic")
	}
}

func TestHintStrings(t *testing.T) {
	res := Compile("var Foo: Integer := 1; PrintLn(foo);", "", semantic.HintsLevelPedantic)
	hints := res.HintStrings()
	if len(hints) == 0 || !strings.HasPrefix(hints[0], "Hint:") {
		t.Fatalf("expected a case-mismatch hint, got %v", res.DiagnosticStrings())
	}
	for _, h := range hints {
		if !strings.HasPrefix(h, "Hint:") && !strings.HasPrefix(h, "Warning:") {
			t.Fatalf("HintStrings must only return hints and warnings, got %q", h)
		}
	}
	errRes := Compile("var x: Integer := 'hello';", "", semantic.HintsLevelPedantic)
	for _, h := range errRes.HintStrings() {
		if strings.HasPrefix(h, "Syntax Error:") {
			t.Fatalf("errors must not appear in HintStrings: %q", h)
		}
	}
}
