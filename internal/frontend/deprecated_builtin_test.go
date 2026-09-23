package frontend

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestCompile_DeprecatedBuiltins pins DWScript's iffDeprecated built-ins
// (RandSeed, CharAt): each reference warns at the name, quoting the declared
// spelling, whether it is called with parentheses or used bare.
func TestCompile_DeprecatedBuiltins(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "bare RandSeed",
			source: "SetRandSeed(1234);\nPrintLn(RandSeed);",
			want:   []string{`Warning: "RandSeed" has been deprecated [line: 2, column: 9]`},
		},
		{
			name:   "RandSeed call, other casing",
			source: "var i := randseed();",
			want:   []string{`Warning: "RandSeed" has been deprecated [line: 1, column: 10]`},
		},
		{
			name:   "CharAt anchors at the name",
			source: "var s := 'ab';\nvar c := CharAt(s, 1);",
			want:   []string{`Warning: "CharAt" has been deprecated [line: 2, column: 10]`},
		},
		{
			name:   "SetRandSeed is not deprecated",
			source: "SetRandSeed(1);",
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "deprecated.pas", semantic.HintsLevelNormal)
			if got := result.DiagnosticStrings(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}
