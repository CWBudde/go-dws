package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestUnitQualified_BareCallsAndFunctionValues(t *testing.T) {
	var output bytes.Buffer
	engine := New(&output)
	registry := units.NewUnitRegistry(nil)
	engine.SetUnitRegistry(registry)
	parse := func(source string) *ast.Program {
		p := parser.New(lexer.New(source))
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatal(p.Errors())
		}
		return program
	}
	unit := units.NewUnit("UnitA", "unita.dws")
	unit.InterfaceSection = &ast.BlockStatement{Statements: parse(`
 procedure PrepareTest; begin PrintLn('prepared') end;
 function GetValue: Integer; begin Result := 7 end;
 function DefaultValue(value: Integer = 9): Integer; begin Result := value end;
 function WithArg(value: Integer): Integer; begin Result := value + 1 end;
 `).Statements}
	registry.RegisterUnit(unit.Name, unit)
	if err := engine.ImportUnitSymbols(unit); err != nil {
		t.Fatal(err)
	}
	result := engine.Eval(parse(`
 uNiTa.PrepareTest;
 PrintLn(UnitA.GetValue);
 PrintLn(UnitA.DefaultValue);
 var callback := UnitA.WithArg;
 PrintLn(callback(4));
 `))
	if err, ok := result.(*ErrorValue); ok {
		t.Fatal(err.Message)
	}
	if got := output.String(); got != "prepared\n7\n9\n5\n" {
		t.Fatalf("output = %q", got)
	}
}
