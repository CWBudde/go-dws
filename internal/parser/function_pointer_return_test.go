package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Rejecting a parsed callable return type loses the declaration and disrupts
// parsing the next statement, as in JSONConnectorFail/autobox.
func TestParseFunctionPointer_NestedReturn(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantReturn string
		wantEnd    int
	}{
		{"procedure", "var p: function: procedure;", "procedure()", 27},
		{"procedure with parameter", "var p: function: procedure(i: Integer);", "procedure(i: Integer)", 39},
		{"function", "var p: function: function: String;", "function(): String", 34},
		{"two return levels", "var p: function: function: procedure;", "function(): procedure()", 37},
		{"method pointer", "var p: function: procedure of object;", "procedure() of object", 37},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(lexer.New(tt.input + "\nvar tail: Integer;"))
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("unexpected parser errors: %v", p.Errors())
			}
			if len(program.Statements) != 2 {
				t.Fatalf("got %d statements, want both variable declarations", len(program.Statements))
			}
			decl, ok := program.Statements[0].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("first statement is %T, want variable declaration", program.Statements[0])
			}
			outer, ok := decl.Type.(*ast.FunctionPointerTypeNode)
			if !ok {
				t.Fatalf("variable type is %T, want function pointer", decl.Type)
			}
			inner, ok := outer.ReturnType.(*ast.FunctionPointerTypeNode)
			if !ok {
				t.Fatalf("return type is %T, want nested function pointer", outer.ReturnType)
			}
			if got := inner.String(); got != tt.wantReturn {
				t.Errorf("nested return = %q, want %q", got, tt.wantReturn)
			}
			if got := outer.End(); got.Line != 1 || got.Column != tt.wantEnd {
				t.Errorf("outer end = %v, want line 1 column %d", got, tt.wantEnd)
			}
			tail, ok := program.Statements[1].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("following statement is %T, want variable declaration", program.Statements[1])
			}
			if tail.Names[0].Value != "tail" || tail.Type.String() != "Integer" {
				t.Errorf("following declaration = %s, want tail: Integer", tail.String())
			}
		})
	}
}

func TestParseFunctionPointer_NestedReturnMissingResult(t *testing.T) {
	p := New(lexer.New("var p: function: function;\nvar tail: Integer;"))
	program := p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Fatal("nested function without a return type must report a parser error")
	}
	for _, stmt := range program.Statements {
		if decl, ok := stmt.(*ast.VarDeclStatement); ok && len(decl.Names) > 0 && decl.Names[0].Value == "tail" {
			return
		}
	}
	t.Fatal("parser did not recover the declaration following the invalid return type")
}
