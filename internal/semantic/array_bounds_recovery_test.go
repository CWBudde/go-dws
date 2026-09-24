package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

func TestArrayBounds_RecoveryDiagnostics(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"procedure bound before missing bracket", "type TProc = procedure;\nvar a: array[1.. TProc;", []string{"Function expected", "Bound isn't of an ordinal type"}},
		{"string type used as bound", "var a: array[String..0] of Integer;", []string{"Bound isn't of an ordinal type", `"]" expected`}},
		{"string bound before missing upper bound", "var a: array['aa'] of Integer;", []string{"Bound isn't of an ordinal type"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.source))
			program := p.ParseProgram()
			a := NewAnalyzer()
			_ = a.Analyze(program)
			got := a.Errors()
			if len(got) != len(tt.want) {
				t.Fatalf("errors = %v, want %v", got, tt.want)
			}
			for i, want := range tt.want {
				if !strings.Contains(got[i], want) {
					t.Errorf("error %d = %q, want %q", i, got[i], want)
				}
			}
		})
	}
}

func TestArrayBounds_ProcedureAnchorIgnoresWhitespace(t *testing.T) {
	for _, source := range []string{
		"type TProc = procedure; var a: array[1..     TProc;",
		"type TProc = procedure; var a: array[1..     TProc] of Integer;",
		"type TProc = procedure; type A = array[1..     TProc] of Integer;",
		"type TProc = procedure; type A = array[1..     TProc;",
	} {
		t.Run(source, func(t *testing.T) {
			p := parser.New(lexer.New(source))
			program := p.ParseProgram()
			a := NewAnalyzer()
			_ = a.Analyze(program)
			got := a.StructuredErrors()
			if len(got) != 2 {
				t.Fatalf("diagnostics = %v, want function and bound errors", got)
			}
			// The bound diagnostic points to the last dot, regardless of intervening spaces.
			want := strings.Index(source, "..") + 2
			if got[1].Pos.Column != want {
				t.Fatalf("bound column = %d, want %d", got[1].Pos.Column, want)
			}
		})
	}
}
