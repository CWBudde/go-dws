package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Testing the analyzer's accumulated diagnostics prevents a frontend display
// filter from masking accidental execution of the program-end hint collector.
func TestAnalyzer_CompilerStopPrivateEndHints(t *testing.T) {
	const source = "type T = class\nprivate\n F: Integer;\n procedure M; begin end;\nend;\nprocedure P; begin var U: Integer; end;"
	const local = "Hint: Variable \"U\" declared but not used [line: 6, column: 24]"
	for _, stopped := range []bool{false, true} {
		name := "completed compilation"
		want := local + "\nHint: Private field \"F\" declared but never used [line: 3, column: 2]\nHint: Private method \"M\" declared but never used [line: 4, column: 12]"
		if stopped {
			name, want = "compiler stop", local
		}
		t.Run(name, func(t *testing.T) {
			p := parser.New(lexer.New(source))
			program := p.ParseProgram()
			if parseErrors := p.Errors(); len(parseErrors) != 0 {
				t.Fatalf("parser errors: %v", parseErrors)
			}
			a := NewAnalyzer()
			a.SetHintsLevel(HintsLevelPedantic)
			a.SetCompileStopped(stopped)
			if err := a.Analyze(program); err != nil {
				t.Fatalf("analysis error: %v", err)
			}
			if got := strings.Join(a.Errors(), "\n"); got != want {
				t.Fatalf("got %q; want %q", got, want)
			}
		})
	}
}

// The deferred-stop API receives node identities, so use real parsed index
// expressions and let semantic property resolution settle each identity.
func TestAnalyzer_DeferredIndexStopPrivateEndHints(t *testing.T) {
	const field = "type T = class\nprivate\n F: Integer;\nend;\n"
	const fieldHint = "Hint: Private field \"F\" declared but never used [line: 3, column: 2]"
	const property = "type R = class function Get(I: Integer): Integer; begin Result := I; end; property Prop[I: Integer]: Integer read Get reintroduce; end;\nvar O := R.Create;\nvar A := O.Prop()[1];"
	const propertyHint = "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 16]"
	for _, tt := range []struct{ name, source, want string }{
		{"unresolved ordinary index", field + "var A: array of Integer;\nvar B := A[1];", ""},
		{"resolved indexed property", field + property, propertyHint + "\n" + fieldHint},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.source))
			program := p.ParseProgram()
			if parseErrors := p.Errors(); len(parseErrors) != 0 {
				t.Fatalf("parser errors: %v", parseErrors)
			}
			decl, ok := program.Statements[len(program.Statements)-1].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("last statement is %T; want variable declaration", program.Statements[len(program.Statements)-1])
			}
			index, ok := decl.Value.(*ast.IndexExpression)
			if !ok {
				t.Fatalf("initializer is %T; want index expression", decl.Value)
			}
			a := NewAnalyzer()
			a.SetHintsLevel(HintsLevelPedantic)
			a.SetDeferredIndexStops([]*ast.IndexExpression{index})
			if err := a.Analyze(program); err != nil {
				t.Fatalf("analysis error: %v", err)
			}
			if got := strings.Join(a.Errors(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}
