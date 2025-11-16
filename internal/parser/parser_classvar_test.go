package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestClassVarParsing(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantClassVar bool
		wantInit     bool
	}{
		{
			name: "simple class var",
			input: `type TBase = class
				class var Test : Integer;
			end;`,
			wantClassVar: true,
			wantInit:     false,
		},
		{
			name: "class var with initialization",
			input: `type TBase = class
				class var Test : Integer := 456;
			end;`,
			wantClassVar: true,
			wantInit:     true,
		},
		{
			name: "regular instance field",
			input: `type TBase = class
				Test : Integer;
			end;`,
			wantClassVar: false,
			wantInit:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements does not contain 1 statement. got=%d",
					len(program.Statements))
			}

			classDecl, ok := program.Statements[0].(*ast.ClassDecl)
			if !ok {
				t.Fatalf("statement is not *ast.ClassDecl. got=%T", program.Statements[0])
			}

			if len(classDecl.Fields) != 1 {
				t.Fatalf("class does not have 1 field. got=%d", len(classDecl.Fields))
			}

			field := classDecl.Fields[0]
			if field.IsClassVar != tt.wantClassVar {
				t.Errorf("field.IsClassVar = %v, want %v", field.IsClassVar, tt.wantClassVar)
			}

			if (field.InitValue != nil) != tt.wantInit {
				t.Errorf("field.InitValue present = %v, want %v", field.InitValue != nil, tt.wantInit)
			}
		})
	}
}
