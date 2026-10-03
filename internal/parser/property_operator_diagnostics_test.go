package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestParser_PropertyOperatorStops(t *testing.T) {
	tests := []struct {
		name, source string
		messages     []string
		line, column int
	}{
		{"read name", "type T = class\n  property P: Integer read ;\nend;\nvar bad := ;", []string{"Name expected"}, 2, 28},
		{"write name", "type T = class\n  property P: Integer write ;\nend;\nvar bad := ;", []string{"Name expected"}, 2, 29},
		{"read EOF", "type T = class\n  property P: Integer read", []string{"Name expected"}, 2, 23},
		{"write EOF", "type T = class\n  property P: Integer write", []string{"Name expected"}, 2, 23},
		{"class operator", "type T = class\n  class operator Strange uses P;\nend;\nvar bad := ;", []string{"Overloadable operator expected"}, 2, 18},
		{"global operator", "operator dummy ;\nvar bad := ;", []string{"Overloadable operator expected", `"(" expected`}, 1, 10},
		{"global EOF", "operator", []string{"Overloadable operator expected", `"(" expected`}, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.source)
			p.ParseProgram()
			errors := p.Errors()
			if len(errors) != len(tt.messages) {
				t.Fatalf("errors = %v; want %q", errors, tt.messages)
			}
			for i, message := range tt.messages {
				err := errors[i]
				if err.Message != message || err.Pos.Line != tt.line || err.Pos.Column != tt.column || err.Stop != (i == len(errors)-1) {
					t.Errorf("error %d = %+v; want %q at %d:%d, stop=%v", i, err, message, tt.line, tt.column, i == len(errors)-1)
				}
			}
		})
	}
}

func TestParser_PropertyAccessorStopPreservesEarlierMembers(t *testing.T) {
	p := testParser("type T = class\n  F: Integer;\n  property Good: Integer read F;\n  property Bad: Integer write ;\nend;")
	program := p.ParseProgram()
	if len(program.Statements) != 1 {
		t.Fatalf("statements = %v", program.Statements)
	}
	decl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok || len(decl.Fields) != 1 || len(decl.Properties) != 1 {
		t.Fatalf("earlier members lost: %#v", decl)
	}
	if decl.Properties[0].Name.Value != "Good" {
		t.Fatalf("property = %s", decl.Properties[0].Name.Value)
	}
	if len(p.Errors()) != 1 || !p.Errors()[0].Stop {
		t.Fatalf("errors = %v", p.Errors())
	}
}
