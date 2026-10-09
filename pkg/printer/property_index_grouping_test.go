package printer_test

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	"github.com/cwbudde/go-dws/pkg/printer"
)

func TestPropertyIndexGrouping_PrintReparse(t *testing.T) {
	source := `type T = class
 F: Integer;
 property P[var I: Integer]: Integer read (I + 1) write (I := Value);
 property Q: Integer read F write (F);
end;
var O := new T; var X := 5;
PrintLn(O.P[(X)]); O.P[(X)] := 8;
O.Q := 3; PrintLn(X); PrintLn(O.Q);
PrintLn((1 + 2) * 3);`
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatal(p.Errors())
	}
	for _, rendered := range []string{source, printer.New(printer.DefaultOptions()).Print(program)} {
		again := parser.New(lexer.New(rendered))
		parsed := again.ParseProgram()
		if len(again.Errors()) != 0 {
			t.Fatalf("reparse %s: %v", rendered, again.Errors())
		}
		_ = parsed
		var output bytes.Buffer
		engine, err := dwscript.New(dwscript.WithOutput(&output))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = engine.Eval(rendered); err != nil {
			t.Fatalf("%s: %v", rendered, err)
		}
		if output.String() != "6\n8\n3\n9\n" {
			t.Fatalf("output %q", output.String())
		}
	}
}

func TestPropertyIndexGrouping_WriterKindsPrintReparse(t *testing.T) {
	for _, writer := range []string{"(F)", "(F := Value)", "(Put(Value))"} {
		t.Run(writer, func(t *testing.T) {
			source := `type T = class
 F: Integer;
 procedure Put(V: Integer); begin F := V; end;
 property P: Integer read F write ` + writer + `;
end;
var O := new T; O.P := 7; PrintLn(O.P);`
			parsed := parser.New(lexer.New(source))
			program := parsed.ParseProgram()
			if len(parsed.Errors()) != 0 {
				t.Fatal(parsed.Errors())
			}
			for _, rendered := range []string{source, printer.New(printer.DefaultOptions()).Print(program)} {
				again := parser.New(lexer.New(rendered))
				reparsed := again.ParseProgram()
				if len(again.Errors()) != 0 {
					t.Fatal(again.Errors())
				}
				property := reparsed.Statements[0].(*ast.ClassDecl).Properties[0]
				if property.WriteStmt == nil || property.WriteSpec != nil {
					t.Fatal("printed writer lost expression kind")
				}
				var output bytes.Buffer
				engine, err := dwscript.New(dwscript.WithOutput(&output))
				if err != nil {
					t.Fatal(err)
				}
				if _, err = engine.Eval(rendered); err != nil {
					t.Fatal(err)
				}
				if output.String() != "7\n" {
					t.Fatalf("output %q", output.String())
				}
			}
		})
	}
}
