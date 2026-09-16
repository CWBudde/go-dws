package frontend

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_ArrayIndexTypeCaseHints(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		hints  []string
	}{
		{
			name:   "boolean index",
			source: "var a: array[boolean] of Integer;",
			hints:  []string{`Hint: "boolean" does not match case of declaration ("Boolean") [line: 1, column: 14]`},
		},
		{
			name:   "enum index",
			source: "type TDay = (Monday, Tuesday);\nvar a: array[tday] of Integer;",
			hints:  []string{`Hint: "tday" does not match case of declaration ("TDay") [line: 2, column: 14]`},
		},
		{
			name:   "alias retains declared name",
			source: "type TFlag = Boolean;\nvar a: array[tflag] of Integer;",
			hints:  []string{`Hint: "tflag" does not match case of declaration ("TFlag") [line: 2, column: 14]`},
		},
		{
			name:   "matching spelling",
			source: "var a: array[Boolean] of Integer;",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := Compile(tc.source, "", semantic.HintsLevelPedantic)
			if res.HasFatalDiagnostics() {
				t.Fatalf("unexpected errors: %v", res.DiagnosticStrings())
			}
			got := res.HintStrings()
			if len(got) != len(tc.hints) || (len(got) != 0 && !reflect.DeepEqual(got, tc.hints)) {
				t.Fatalf("hints = %v, want %v", got, tc.hints)
			}
		})
	}
}
