package frontend

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_SingletonOverloadForwardHintAndStopReach(t *testing.T) {
	for _, level := range []struct {
		name  string
		hints semantic.HintsLevel
	}{{"normal", semantic.HintsLevelNormal}, {"pedantic", semantic.HintsLevelPedantic}} {
		for _, tt := range []struct {
			name, source, want string
			stop, fatal        bool
		}{
			{"new member reaches calling hint", "procedure P(x: Integer); overload; forward;\nprocedure P(x: String); overload; export; cdecl; begin end;\nprocedure P(x: Integer); begin end;", "Hint: Call convention \"cdecl\" is not supported and ignored [line: 2, column: 43]", false, false},
			{"selected original rejects export and later hint", "procedure Before; cdecl; begin end;\nprocedure P(x: Integer); overload; forward;\nprocedure P(x: String); overload; export; cdecl; begin end;\nprocedure P(x: Integer); export; cdecl; begin Bad; end; {$ERROR 'late'} Later;", "Hint: Call convention \"cdecl\" is not supported and ignored [line: 1, column: 19]\nHint: Call convention \"cdecl\" is not supported and ignored [line: 3, column: 43]\nSyntax Error: BEGIN expected [line: 4, column: 26]", true, true},
			{"matched EOF", "procedure P(x: Integer); overload; forward;\nprocedure P(x: Integer); export", "Syntax Error: BEGIN expected [line: 2, column: 26]", true, true},
			{"new overload conflict stays nonstopping", "procedure P(x: Integer = 1); overload; forward;\nprocedure P(x: String = 's'); overload; export; begin end;\nprocedure P(x: Integer = 1); begin end;", "Syntax Error: Overload of \"P\" will be ambiguous with a previously declared version [line: 2, column: 1]", false, true},
		} {
			t.Run(level.name+"/"+tt.name, func(t *testing.T) {
				result := CompileWithOptions(tt.source, Options{HintsLevel: level.hints, DisableSymbolDictionaryDiagnostics: true})
				if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
					t.Fatalf("complete diagnostics:\n%s\nwant:\n%s", got, tt.want)
				}
				if len(result.Diagnostics) == 0 || result.Diagnostics[len(result.Diagnostics)-1].Stop != tt.stop {
					t.Fatalf("typed stop: %+v; want %v", result.Diagnostics, tt.stop)
				}
				if result.HasFatalDiagnostics() != tt.fatal {
					t.Fatalf("error state: %+v", result.Diagnostics)
				}
			})
		}
	}
}
