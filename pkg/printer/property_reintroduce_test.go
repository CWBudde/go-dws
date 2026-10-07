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

func TestPrint_InheritedReintroducedPropertyRoundTrip(t *testing.T) {
	const source = `type TBase = class
 Field: Integer;
 property Prop: Integer read Field reintroduce;
end;
type TChild = class(TBase)
 function Probe: Integer;
 begin Result := inherited Prop(); end;
end;
var Obj := new TChild; PrintLn(Obj.Probe());`
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) != 0 {
		t.Fatalf("parse source: %v", errors)
	}
	rendered := printer.New(printer.DefaultOptions()).Print(program)
	result := frontend.CompileWithOptions(rendered, frontend.Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
	diagnostics := result.DiagnosticStrings()
	if len(diagnostics) != 1 || !strings.HasPrefix(diagnostics[0], `Hint: Property "Prop" reintroduced a method, you should remove empty brackets ()`) {
		t.Fatalf("recompiled output lost empty compatibility brackets: %v:\n%s", diagnostics, rendered)
	}
}

func TestPrint_InheritedMethodArgumentsRoundTrip(t *testing.T) {
	const source = `type TBase = class
 function Sum(A, B: Integer): Integer; begin Result := A+B; end;
end;
type TChild = class(TBase)
 function Probe: Integer; begin Result := inherited Sum(7, 5); end;
end;
var Obj := new TChild; PrintLn(Obj.Probe());`
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) != 0 {
		t.Fatalf("parse source: %v", errors)
	}
	rendered := printer.New(printer.DefaultOptions()).Print(program)
	result := frontend.CompileWithOptions(rendered, frontend.Options{DisableSymbolDictionaryDiagnostics: true})
	if !result.SemanticSuccessful {
		t.Fatalf("printed inherited arguments failed compilation: %v:\n%s", result.DiagnosticStrings(), rendered)
	}
}
