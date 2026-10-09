package frontend

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestCompile_TruncatedCallKeepsCompletedArguments pins that a call whose argument
// list is cut short by a compiler stop still reports the diagnostics of the
// arguments it did complete. Upstream's ReadArguments compiles each argument as it
// reads it, so their errors precede the stop; the call itself is never resolved,
// so no argument-count or argument-type checks run against it.
func TestCompile_TruncatedCallKeepsCompletedArguments(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			// FailureScripts/missing_parenthesis1
			name:   "function call",
			source: "PrintLn('dummy ('+IntToStr(12+')');",
			want: []string{
				`Syntax Error: Invalid Operands [line: 1, column: 30]`,
				`Syntax Error: ")" expected [line: 1, column: 35]`,
			},
		},
		{
			name:   "function call with an identifier first argument",
			source: "var i : Integer;\nPrintLn(i, 12+'x' 1);",
			want: []string{
				`Syntax Error: Invalid Operands [line: 2, column: 14]`,
				`Syntax Error: ")" expected [line: 2, column: 19]`,
			},
		},
		{
			name:   "constructor call",
			source: "type\n   TMyClass = class\n      constructor Create(i : Integer);\n   end;\n \nnew TMyClass(12+'x' 1);",
			want: []string{
				`Syntax Error: Invalid Operands [line: 6, column: 16]`,
				`Syntax Error: ")" expected [line: 6, column: 21]`,
			},
		},
		{
			name:   "method call",
			source: "var a : array of Integer;\na.Add(12+'x' 1);",
			want: []string{
				`Syntax Error: Invalid Operands [line: 2, column: 9]`,
				`Syntax Error: ")" expected [line: 2, column: 14]`,
			},
		},
		{
			// FailureScripts/array_index_bracket_missing1: the stopped argument is not
			// completed, so neither it nor the call is checked.
			name:   "argument cut short by its own stop",
			source: "procedure Test(Data: array of Float);\nbegin\n  \nend;\n\nTest([1.2, 2.2);",
			want:   []string{`Syntax Error: "]" expected [line: 6, column: 15]`},
		},
		{
			// FailureScripts/constructor_invalid_param: the completed argument is
			// valid; the truncated call is not checked against the constructor.
			name:   "constructor call with a missing comma",
			source: "type\n   TMyClass = class\n      constructor Create(i : Integer);\n   end;\n \nnew TMyClass(1 1);",
			want:   []string{`Syntax Error: ")" expected [line: 6, column: 16]`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			got := result.DiagnosticStrings()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}
