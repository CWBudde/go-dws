package frontend

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestBuiltinDeclarationCaseHints(t *testing.T) {
	for _, tt := range []struct {
		name   string
		source string
		want   []string
	}{
		{"bare intrinsic", "var i := low;", []string{`Hint: "low" does not match case of declaration ("Low") [line: 1, column: 10]`}},
		{"invalid intrinsic argument", "var i := low(PrintLn(''));", []string{`Hint: "low" does not match case of declaration ("Low") [line: 1, column: 10]`}},
		{"registry declaration", "PrintLn(format('%d', [1]));", []string{`Hint: "format" does not match case of declaration ("Format") [line: 1, column: 9]`}},
		{"var parameter intrinsic", "var i := 1; inc(i); PrintLn(i);", []string{`Hint: "inc" does not match case of declaration ("Inc") [line: 1, column: 13]`}},
		{"matching declaration", "var i := 1; Inc(i); PrintLn(Format('%d', [i]));", nil},
		{"user declaration shadows builtin", "function format: String; begin Result := 'ok'; end; PrintLn(format());", nil},
		{"unknown names have no declaration", "DoesNotExist();", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "", semantic.HintsLevelPedantic)
			var got []string
			for _, hint := range result.HintStrings() {
				if strings.Contains(hint, "does not match case of declaration") {
					got = append(got, hint)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("case hints = %q, want %q; diagnostics: %v", got, tt.want, result.DiagnosticStrings())
			}
		})
	}
}
