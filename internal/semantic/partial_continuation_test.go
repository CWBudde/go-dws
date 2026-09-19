package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestPartialContinuation_HintPosition(t *testing.T) {
	for _, test := range []struct {
		name, continuation, position string
	}{
		{"visibility", "type TTest = class\n  private\n  F: Integer;\nend;", "line: 3, column: 3"},
		{"member", "type TTest = class\n  F: Integer;\nend;", "line: 3, column: 3"},
		{"empty", "type TTest = class\nend;", "line: 3, column: 1"},
		{"comment", "type TTest = class { ignored }\n\n  end;", "line: 4, column: 3"},
		{"ancestor", "type TTest = class(TObject) end;", "line: 2, column: 19"},
		{"modifier", "type TTest = class abstract\nend;", "line: 3, column: 1"},
		{"modifier ancestor", "type TTest = class abstract(TObject) end;", "line: 2, column: 28"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := "type TTest = partial class end;\n" + test.continuation
			diagnostics := analyzeWithHints(t, source, HintsLevelPedantic)
			want := `Hint: Previous declaration of class was "partial" [` + test.position + "]"
			if !containsDiagnostic(diagnostics, want) {
				t.Fatalf("missing %q in %v", want, diagnostics)
			}
			for _, diagnostic := range analyzeWithHints(t, source, HintsLevelDisabled) {
				if strings.Contains(diagnostic, "Previous declaration") {
					t.Errorf("hint emitted with hints disabled: %s", diagnostic)
				}
			}
		})
	}
}

func TestPartialContinuation_MissingHintPositionFallsBack(t *testing.T) {
	p := parser.New(lexer.New("type TTest = partial class end;\ntype TTest = class end;"))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	decl := program.Statements[1].(*ast.ClassDecl)
	decl.PartialHintPos = token.Position{}
	analyzer := NewAnalyzer()
	if err := analyzer.Analyze(program); err != nil {
		t.Fatalf("semantic errors: %v", err)
	}
	want := `Hint: Previous declaration of class was "partial" [line: 2, column: 14]`
	if !containsDiagnostic(analyzer.Errors(), want) {
		t.Errorf("missing %q in %v", want, analyzer.Errors())
	}
	if !analyzer.GetClasses()["ttest"].IsPartial {
		t.Error("nonpartial continuation closed the partial class")
	}
}
