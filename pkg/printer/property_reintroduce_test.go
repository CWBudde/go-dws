package printer_test

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/printer"
)

func TestPrint_ReintroducedPropertyRoundTrip(t *testing.T) {
	const source = `type TTest = class
  Field: Integer;
  property Prop: Integer read Field reintroduce;
end;
var Obj := new TTest;
PrintLn(Obj.Prop());`
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) != 0 {
		t.Fatalf("parse source: %v", errors)
	}
	rendered := printer.New(printer.DefaultOptions()).Print(program)
	result := frontend.CompileWithOptions(rendered, frontend.Options{
		Filename: "<test>", HintsLevel: semantic.HintsLevelNormal,
		DisableSymbolDictionaryDiagnostics: true,
	})
	diagnostics := result.DiagnosticStrings()
	if len(diagnostics) != 1 || !strings.HasPrefix(diagnostics[0], `Hint: Property "Prop" reintroduced a method, you should remove empty brackets ()`) {
		t.Fatalf("recompiled output has diagnostics %v:\n%s", diagnostics, rendered)
	}
}
