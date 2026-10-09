package frontend

import (
	"os"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_ExportFixture(t *testing.T) {
	source, err := os.ReadFile("../../testdata/fixtures/FailureScripts/export.pas")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../../testdata/fixtures/FailureScripts/export.txt")
	if err != nil {
		t.Fatal(err)
	}
	result := Compile(string(source), "export.pas", semantic.HintsLevelPedantic)
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.TrimSpace(string(want)) {
		t.Fatalf("complete diagnostics:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// Expectations come from pinned ReadProcDecl/ReadProcBody: an invalid optional
// name stays unread for ReadSemiColon and the subsequent BEGIN compiler stop.
func TestCompile_ExportRecovery(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"identifier", "procedure P; export Foo; begin end;", "Syntax Error: \";\" expected [line: 1, column: 21]\nSyntax Error: BEGIN expected [line: 1, column: 21]"},
		{"number", "procedure P; export 42; begin end;", "Syntax Error: \";\" expected [line: 1, column: 21]\nSyntax Error: BEGIN expected [line: 1, column: 21]"},
		{"extra after string", "procedure P; export 'name' extra; begin end;", "Syntax Error: \";\" expected [line: 1, column: 28]\nSyntax Error: BEGIN expected [line: 1, column: 28]"},
		{"BEGIN survives missing semicolon", "procedure P; export begin end;", "Syntax Error: \";\" expected [line: 1, column: 21]"},
		{"missing body", "procedure P; export; PrintLn('late');", "Syntax Error: BEGIN expected [line: 1, column: 22]"},
		{"EOF after semicolon", "procedure P; export;", "Syntax Error: BEGIN expected [line: 1, column: 20]"},
		{"EOF without semicolon", "procedure P; export", "Syntax Error: \";\" expected [line: 1, column: 14]\nSyntax Error: BEGIN expected [line: 1, column: 14]"},
		{"EOF after name without semicolon", "procedure P; export 'name'", "Syntax Error: \";\" expected [line: 1, column: 21]\nSyntax Error: BEGIN expected [line: 1, column: 21]"},
		{"EOF after named semicolon", "procedure P; export 'name';", "Syntax Error: BEGIN expected [line: 1, column: 27]"},
		{"EOF after comments and newlines", "procedure P;\n export; {comment}\n // trailing\n", "Syntax Error: BEGIN expected [line: 2, column: 8]"},
		{"duplicate export", "procedure P; export; export; begin end;", "Syntax Error: BEGIN expected [line: 1, column: 22]"},
		{"export after inline", "procedure P; inline; export; begin end;", "Syntax Error: Expression expected before EXPORT [line: 1, column: 22]"},
		{"export after calling qualifier", "procedure P; cdecl; export; begin end;", "Syntax Error: Expression expected before EXPORT [line: 1, column: 21]\nHint: Call convention \"cdecl\" is not supported and ignored [line: 1, column: 14]"},
		{"export after deprecated", "procedure P; deprecated 'old'; export; begin end;", "Syntax Error: Expression expected before EXPORT [line: 1, column: 32]"},
		{"external after export", "procedure P; export; external; begin end;", "Syntax Error: BEGIN expected [line: 1, column: 22]"},
		{"forward after export", "procedure P; export; forward; begin end;", "Syntax Error: BEGIN expected [line: 1, column: 22]"},
		{"overload after export", "procedure P; export; overload; begin end;", "Syntax Error: BEGIN expected [line: 1, column: 22]"},
		{"earlier semantic and later lexer errors", "PrintLn(Missing);\nprocedure P; export Foo;\nbegin Bad; end;\n{$ERROR 'late'}\nLater;", "Syntax Error: Unknown name \"Missing\" [line: 1, column: 9]\nSyntax Error: \";\" expected [line: 2, column: 21]\nSyntax Error: BEGIN expected [line: 2, column: 21]"},
		{"pending end checks suppressed", "procedure Pending; forward;\ntype TForward = class;\nprocedure P; export Foo; begin end;", "Syntax Error: \";\" expected [line: 3, column: 21]\nSyntax Error: BEGIN expected [line: 3, column: 21]"},
		{"earlier stop wins", "var x := ;\nprocedure P; export Foo; begin Bad; end;\n{$ERROR 'late'}", "Syntax Error: Expression expected [line: 1, column: 10]"},
		{"external control", "procedure P; external 'host';", ""},
		{"forward control", "procedure P; forward; procedure P; begin end; P;", ""},
	}
	for _, level := range []struct {
		name  string
		hints semantic.HintsLevel
	}{{"normal", semantic.HintsLevelNormal}, {"pedantic", semantic.HintsLevelPedantic}} {
		for _, tt := range tests {
			t.Run(level.name+"/"+tt.name, func(t *testing.T) {
				result := CompileWithOptions(tt.source, Options{HintsLevel: level.hints, DisableSymbolDictionaryDiagnostics: true})
				if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
					t.Fatalf("got:\n%s\nwant:\n%s", got, tt.want)
				}
			})
		}
	}
}
