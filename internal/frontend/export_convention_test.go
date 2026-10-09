package frontend

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_ExportCallingConventionReach(t *testing.T) {
	for _, level := range []struct {
		name  string
		hints semantic.HintsLevel
	}{{"normal", semantic.HintsLevelNormal}, {"pedantic", semantic.HintsLevelPedantic}} {
		for _, tt := range []struct{ name, source, want string }{
			{"repeated export", "procedure P; forward;\nprocedure P; export; cdecl; begin end;", "Syntax Error: BEGIN expected [line: 2, column: 14]"},
			{"repeated helper alias", "function F(X: Integer): Integer; forward;\nfunction F(X: Integer): Integer; export; helper Alias; cdecl; begin Result := X; end;", "Syntax Error: BEGIN expected [line: 2, column: 34]"},
			{"repeated EOF", "procedure P; forward;\nprocedure P; export; cdecl;", "Syntax Error: BEGIN expected [line: 2, column: 14]"},
			{"repeated malformed body", "procedure P; forward;\nprocedure P; export; cdecl; Bad; {$ERROR 'late'}", "Syntax Error: BEGIN expected [line: 2, column: 14]"},
			{"earlier header mismatch", "procedure P(X: Integer); forward;\nprocedure P(X: String); export; cdecl; begin end;", "Syntax Error: implementation signature for 'P' does not match forward declaration [line: 2, column: 1]\nSyntax Error: BEGIN expected [line: 2, column: 25]"},
			{"ordinary convention", "procedure P; cdecl; begin end;", "Hint: Call convention \"cdecl\" is not supported and ignored [line: 1, column: 14]"},
			{"new exported convention", "procedure P; export; cdecl; begin end;", "Hint: Call convention \"cdecl\" is not supported and ignored [line: 1, column: 22]"},
			{"new exported helper convention", "function F(X: Integer): Integer; export; helper Alias; cdecl; begin Result := X; end;", "Hint: Call convention \"cdecl\" is not supported and ignored [line: 1, column: 56]"},
		} {
			t.Run(level.name+"/"+tt.name, func(t *testing.T) {
				result := CompileWithOptions(tt.source, Options{HintsLevel: level.hints, DisableSymbolDictionaryDiagnostics: true})
				if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
					t.Fatalf("got:\n%s\nwant:\n%s", got, tt.want)
				}
			})
		}
	}
}
