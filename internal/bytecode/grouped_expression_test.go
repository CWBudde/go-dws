package bytecode

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompiler_GroupedUnsupportedInner(t *testing.T) {
	for _, tt := range []struct{ source, message string }{
		{"var Obj: TObject; PrintLn(((Obj as TObject)));", "unsupported expression type *ast.AsExpression"},
		{"var Obj: TObject; var Meta: TClass; PrintLn(((Obj is Meta)));", "type checking with 'is' operator not yet supported in bytecode mode"},
	} {
		for _, checked := range []bool{false, true} {
			t.Run(tt.message+map[bool]string{false: "/unchecked", true: "/checked"}[checked], func(t *testing.T) {
				p := parser.New(lexer.New(tt.source))
				program := p.ParseProgram()
				if len(p.Errors()) != 0 {
					t.Fatal(p.Errors())
				}
				compiler := NewCompiler("test")
				if checked {
					a := semantic.NewAnalyzer()
					if err := a.Analyze(program); err != nil {
						t.Fatal(err)
					}
					compiler.SetSemanticInfo(a.GetSemanticInfo())
				}
				_, err := compiler.Compile(program)
				if err == nil || !strings.Contains(err.Error(), tt.message) {
					t.Fatalf("wanted original inner-expression rejection, got %v", err)
				}
			})
		}
	}
}
