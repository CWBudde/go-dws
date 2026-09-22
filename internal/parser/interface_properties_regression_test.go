package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestInterfaceProperties_MethodKeywordAndDefaultPosition(t *testing.T) {
	const source = `type IItems = interface
 method GetItem(x: Integer): Integer;
 method SetItem(x, value: Integer);
 property Items[x: Integer]: Integer read GetItem write SetItem; default;
 end;`
	p := New(lexer.New(source))
	program := p.ParseProgram()
	checkParserErrors(t, p)
	decl := program.Statements[0].(*ast.InterfaceDecl)
	if len(decl.Methods) != 2 || decl.Methods[0].ReturnType == nil || decl.Methods[1].ReturnType != nil {
		t.Fatalf("incorrect method signatures: %#v", decl.Methods)
	}
	if len(decl.Methods[0].Parameters) != 1 || len(decl.Methods[1].Parameters) != 2 {
		t.Fatal("method parameters were not preserved")
	}
	prop := decl.Properties[0]
	line := strings.Split(source, "\n")[3]
	wantColumn := strings.Index(line, "default;") + len("default") + 1
	if !prop.IsDefault || prop.DefaultPos.Line != 4 || prop.DefaultPos.Column != wantColumn {
		t.Fatalf("default position: %v, want 4:%d", prop.DefaultPos, wantColumn)
	}
}
