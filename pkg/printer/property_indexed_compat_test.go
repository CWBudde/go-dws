package printer_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	"github.com/cwbudde/go-dws/pkg/printer"
)

func TestPrint_IndexedCompatibilityRoundTrip(t *testing.T) {
	for _, name := range []string{"basic", "multi", "class_getter", "array_result"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("../../testdata/property_indexed_compat/" + name + ".dws")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("../../testdata/property_indexed_compat/" + name + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			p := parser.New(lexer.New(string(source)))
			program := p.ParseProgram()
			if errs := p.Errors(); len(errs) != 0 {
				t.Fatalf("parse: %v", errs)
			}
			rendered := printer.New(printer.DefaultOptions()).Print(program)
			compiled := frontend.CompileWithOptions(rendered, frontend.Options{HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			hints := 0
			for _, diagnostic := range compiled.DiagnosticStrings() {
				if !strings.HasPrefix(diagnostic, `Hint: Property "`) {
					t.Fatalf("diagnostics: %v\n%s", compiled.DiagnosticStrings(), rendered)
				}
				hints++
			}
			wantHints := 1
			if name == "class_getter" {
				wantHints = 3
			}
			if hints != wantHints {
				t.Fatalf("got %d hints; want %d\n%s", hints, wantHints, rendered)
			}
			var output bytes.Buffer
			engine, err := dwscript.New(dwscript.WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = engine.Eval(rendered); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != string(want) {
				t.Fatalf("output %q; want %q\n%s", got, want, rendered)
			}
		})
	}
}
