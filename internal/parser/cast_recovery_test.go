package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestAsExpression_RetainedTerm(t *testing.T) {
	for _, tt := range []struct{ source string }{
		{`var x := obj as "hello"; var after := 1;`},
		{`var x := obj as Target(); var after := 1;`},
		{`var x := obj as (Target()); var after := 1;`},
		{`var x := obj as Targets[0]; var after := 1;`},
	} {
		t.Run(tt.source, func(t *testing.T) {
			p := New(lexer.New(tt.source))
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if len(program.Statements) != 2 {
				t.Fatalf("statements: %d", len(program.Statements))
			}
			decl := program.Statements[0].(*ast.VarDeclStatement)
			cast, ok := decl.Value.(*ast.AsExpression)
			if !ok {
				t.Fatalf("cast: %T", decl.Value)
			}
			if cast.Right == nil {
				t.Fatalf("RHS missing: %s", cast.String())
			}
			if cast.TargetType != nil {
				t.Fatal("expression retained as type")
			}
			if cast.Pos().Column != 10 || cast.End().Offset <= cast.Pos().Offset {
				t.Fatalf("span: %v .. %v", cast.Pos(), cast.End())
			}
		})
	}
}

func TestParserState_RepeatedRestoreAfterLookahead(t *testing.T) {
	p := New(lexer.New("target() + following"))
	state := p.saveState()
	for attempt := 0; attempt < 3; attempt++ {
		for i, want := range ([]string{"target", "(", ")", "+", "following"})[:attempt+2] {
			if got := p.cursor.Peek(i).Literal; got != want {
				t.Fatalf("attempt %d token %d = %q, want %q", attempt, i, got, want)
			}
		}
		p.restoreState(state)
	}
}

func TestAsExpression_FollowingOperatorsAndDelimiters(t *testing.T) {
	for _, source := range []string{
		`var x := obj as "hello" + 1; var after := 1;`,
		`var x := obj as ("hello") * 2; var after := 1;`,
		`var x := obj as Target() and true; var after := 1;`,
		`PrintLn(obj as Target(), 1); var after := 1;`,
		`(obj as TObject).Free; var after := 1;`,
	} {
		t.Run(source, func(t *testing.T) {
			p := New(lexer.New(source))
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if len(program.Statements) != 2 {
				t.Fatalf("statements: %d", len(program.Statements))
			}
			count := 0
			ast.Inspect(program, func(node ast.Node) bool {
				if cast, ok := node.(*ast.AsExpression); ok {
					count++
					if _, binary := cast.Right.(*ast.BinaryExpression); binary {
						t.Fatal("RHS consumed following operator")
					}
				}
				return true
			})
			if count != 1 {
				t.Fatalf("cast count: %d", count)
			}
		})
	}
}
