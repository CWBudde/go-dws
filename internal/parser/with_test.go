package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestWithStatement_LocalDeclarations(t *testing.T) {
	input := `with b := "a", d := 3 do begin
	with c := b + "c" do
		PrintLn(c);
	PrintLn(d);
	with c : Integer = d + 1 do
		PrintLn(c);
end;`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.WithStatement)
	if !ok {
		t.Fatalf("statement is not ast.WithStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Declarations) != 2 {
		t.Fatalf("with declaration count = %d, want 2", len(stmt.Declarations))
	}
	if stmt.Declarations[0].Names[0].Value != "b" {
		t.Errorf("first declaration name = %q, want b", stmt.Declarations[0].Names[0].Value)
	}
	if stmt.Declarations[1].Names[0].Value != "d" {
		t.Errorf("second declaration name = %q, want d", stmt.Declarations[1].Names[0].Value)
	}

	body, ok := stmt.Body.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("with body is not ast.BlockStatement. got=%T", stmt.Body)
	}
	if len(body.Statements) != 3 {
		t.Fatalf("with body statement count = %d, want 3", len(body.Statements))
	}

	nested, ok := body.Statements[0].(*ast.WithStatement)
	if !ok {
		t.Fatalf("first body statement is not ast.WithStatement. got=%T", body.Statements[0])
	}
	if len(nested.Declarations) != 1 || nested.Declarations[0].Names[0].Value != "c" {
		t.Fatalf("nested declaration = %#v, want single c declaration", nested.Declarations)
	}

	typed, ok := body.Statements[2].(*ast.WithStatement)
	if !ok {
		t.Fatalf("third body statement is not ast.WithStatement. got=%T", body.Statements[2])
	}
	if typed.Declarations[0].Type == nil {
		t.Fatal("typed with declaration should have a type")
	}
}
