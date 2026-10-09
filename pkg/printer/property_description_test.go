package printer_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	"github.com/cwbudde/go-dws/pkg/printer"
)

func TestPrint_PropertyDescriptionRoundTrip(t *testing.T) {
	for _, name := range []string{"literal", "empty", "escaped", "multiline"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("../../testdata/property_description/" + name + ".dws")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("../../testdata/property_description/" + name + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			p := parser.New(lexer.New(string(source)))
			tree := p.ParseProgram()
			if errors := p.Errors(); len(errors) != 0 {
				t.Fatalf("parse: %v", errors)
			}
			rendered := printer.New(printer.DefaultOptions()).Print(tree)
			var output bytes.Buffer
			engine, err := dwscript.New(dwscript.WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(rendered)
			if err != nil {
				t.Fatalf("compile: %v\n%s", err, rendered)
			}
			original := tree.Statements[0].(*ast.ClassDecl).Properties
			roundTrip := program.AST().Statements[0].(*ast.ClassDecl).Properties
			if len(roundTrip) != len(original) {
				t.Fatalf("lost properties: %s", rendered)
			}
			for i, prop := range original {
				got := roundTrip[i]
				if prop.HasDescription != got.HasDescription || prop.Description != got.Description || prop.IsReintroduce != got.IsReintroduce || prop.IsDefault != got.IsDefault {
					t.Fatalf("metadata differs: source %+v printed %+v", prop, got)
				}
			}
			if _, err := engine.Run(program); err != nil {
				t.Fatalf("run: %v", err)
			}
			if got := output.String(); got != string(want) {
				t.Fatalf("output %q; want %q\n%s", got, want, rendered)
			}
		})
	}
}
