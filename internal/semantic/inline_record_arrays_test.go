package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestInlineRecordArrays_TypeResolution(t *testing.T) {
	for _, tt := range []struct{ name, source string }{
		{"dynamic", `var rows: array of record value: Integer; end; rows.SetLength(1); rows[0].value := 42;`},
		{"static constant", `const rows: array[2..2] of record point: record x, y: Integer; end; end = [(point: (x: 3; y: 4))];`},
		{"nested", `var rows: array of array[-1..1] of record values: array of Integer; end;`},
		{"named dynamic", `type TRows = array of record value: Integer; end; var rows: TRows; rows.SetLength(1); rows[0].value := 42;`},
		{"named static", `type TRows = array[2..3] of record value: Integer; end; var rows: TRows; rows[2].value := 42;`},
	} {
		t.Run(tt.name, func(t *testing.T) { expectNoErrors(t, tt.source) })
	}
}

func TestInlineRecordArrays_InvalidFieldTypes(t *testing.T) {
	for _, tt := range []struct{ source, want string }{
		{`var rows: array of record value: MissingType; end;`, "unknown type"},
		{`var rows: array of record value: Integer; VALUE: String; end;`, "already exists"},
	} {
		t.Run(tt.source, func(t *testing.T) {
			p := parser.New(lexer.New(tt.source))
			program := p.ParseProgram()
			if errs := p.Errors(); len(errs) != 0 {
				t.Fatalf("parse errors: %v", errs)
			}
			decl := program.Statements[0].(*ast.VarDeclStatement)
			_, err := NewAnalyzer().resolveTypeExpression(decl.Type)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("resolution error = %v, want %q", err, tt.want)
			}
		})
	}
	expectError(t, `var rows: array of record value: Integer; end := [(value: 'wrong')];`, "String")
}
