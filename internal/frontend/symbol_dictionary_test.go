package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestSymbolDictionaryDiagnostics(t *testing.T) {
	for _, tc := range []struct{ name, source, hint string }{
		{"local", "procedure P; var Unused: Integer; begin end;", `Variable "Unused" declared but not used`},
		{"result", "function F: Integer; begin end;", "Result is never used"},
		{"field", "type T = class private FUnused: Integer; end;", `Private field "FUnused" declared but never used`},
		{"method", "type T = class private procedure Unused; begin end; end;", `Private method "Unused" declared but never used`},
		{"reference parameter", "procedure P(var Obj: TObject); begin end;", `"Obj" parameter is a reference type passed as VAR`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, disabled := range []bool{false, true} {
				r := CompileWithOptions(tc.source, Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: disabled})
				if r.HasFatalDiagnostics() {
					t.Fatalf("compile: %v", r.DiagnosticStrings())
				}
				if got := strings.Contains(strings.Join(r.HintStrings(), "\n"), tc.hint); got == disabled {
					t.Errorf("disabled=%v: diagnostics %v", disabled, r.DiagnosticStrings())
				}
			}
		})
	}
}

func TestSymbolDictionaryDiagnostics_PreservesOtherHints(t *testing.T) {
	r := CompileWithOptions("{$HINT 'explicit'}\nvar Foo := 1; PrintLn(foo);", Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: true})
	got := strings.Join(r.HintStrings(), "\n")
	if r.HasFatalDiagnostics() || !strings.Contains(got, "explicit") || !strings.Contains(got, "does not match case") {
		t.Fatalf("diagnostics: %v", r.DiagnosticStrings())
	}
}

func TestSymbolDictionaryDiagnostics_InterfaceImplementation(t *testing.T) {
	for _, source := range []string{
		`type I = interface procedure Hello; end;
type T = class(TObject, I) private procedure Hello; begin end; procedure Unused; begin end; end;`,
		`type I = interface procedure Hello; end;
type TBase = class private procedure Hello; begin end; procedure Unused; begin end; end;
type T = class(TBase, I) end;`,
		`type I = interface procedure Hello; end;
type IChild = interface(I) end;
type T = class(TObject, IChild) private procedure hELLo; begin end; procedure Unused; begin end; end;`,
	} {
		r := Compile(source, "", semantic.HintsLevelPedantic)
		if r.HasFatalDiagnostics() {
			t.Fatalf("compile: %v", r.DiagnosticStrings())
		}
		var private []string
		for _, h := range r.HintStrings() {
			if strings.Contains(h, "Private method") {
				private = append(private, h)
			}
		}
		if len(private) != 1 || !strings.Contains(private[0], `"Unused"`) {
			t.Fatalf("private hints: %v", private)
		}
	}
}

func TestSymbolDictionaryDiagnostics_IntfPrivate(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "fixtures", "InterfacesPass", "intf_private.pas")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, disabled := range []bool{false, true} {
		r := CompileWithOptions(string(source), Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: disabled})
		if r.HasFatalDiagnostics() {
			t.Fatalf("compile: %v", r.DiagnosticStrings())
		}
		want := `Hint: Private method "Unused" declared but never used [line: 10, column: 20]`
		if disabled {
			want = ""
		}
		if got := strings.Join(r.HintStrings(), "\n"); got != want {
			t.Errorf("disabled=%v: got %q, want %q", disabled, got, want)
		}
	}
}

func TestSymbolDictionaryDiagnostics_Units(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "HintsUnit.pas"), []byte(`unit HintsUnit;
interface
procedure P;
implementation
procedure P; var Unused: Integer; begin end;
end.`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, disabled := range []bool{false, true} {
		r := CompileWithOptions("uses HintsUnit; P;", Options{UnitSearchPaths: []string{dir}, HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: disabled})
		if r.HasFatalDiagnostics() {
			t.Fatalf("compile: %v", r.DiagnosticStrings())
		}
		if got := strings.Contains(strings.Join(r.HintStrings(), "\n"), `Variable "Unused"`); got == disabled {
			t.Errorf("disabled=%v: %v", disabled, r.DiagnosticStrings())
		}
	}
}
