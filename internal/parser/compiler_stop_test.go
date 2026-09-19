package parser

import (
	"reflect"
	"testing"
)

// TestParser_ExpressionExpectedIsACompilerStop pins that ReadTerm's "Expression
// expected" ends diagnostic output, as DWScript's AddCompilerStop does: the recovery
// errors go-dws's parser would otherwise add afterwards are never recorded.
func TestParser_ExpressionExpectedIsACompilerStop(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"grouped expression missing an operand", "(5 and );"},
		{"enum element value missing", "Type TElement = (etOne=);"},
		{"case range missing its upper bound", "var i : Integer;\ncase i of\n   1..;"},
		{"later statements are not reported either", "x := (1 + );\nvar y : Integer := ;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			p.ParseProgram()
			got := parserErrorMessages(p)
			if want := []string{"Expression expected"}; !reflect.DeepEqual(got, want) {
				t.Fatalf("errors = %q, want %q", got, want)
			}
			if !p.Errors()[0].Stop {
				t.Fatalf("Expression expected should be recorded as a compiler stop")
			}
		})
	}
}

// TestParser_CompilerStopIsUndoneByBacktracking checks that a stop recorded during a
// speculative parse does not outlive the restoreState that discards it.
func TestParser_CompilerStopIsUndoneByBacktracking(t *testing.T) {
	p := testParser("x")
	state := p.saveState()
	p.recordStop(NewParserError(p.cursor.Current().Pos, 1, "Expression expected", ErrInvalidExpression))
	if !p.stopped() {
		t.Fatalf("expected the parser to be stopped after recordStop")
	}
	p.restoreState(state)
	if p.stopped() {
		t.Fatalf("restoreState should discard the speculative stop")
	}
	p.addError("later error", ErrInvalidSyntax)
	if got := parserErrorMessages(p); !reflect.DeepEqual(got, []string{"later error"}) {
		t.Fatalf("errors = %q, want only the post-restore error", got)
	}
}
