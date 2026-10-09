package frontend

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_ExportForwardImplementationStops(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"matching forward", "procedure P; forward;\nprocedure P; export; begin end;", "Syntax Error: BEGIN expected [line: 2, column: 14]"},
		{"EOF after export semicolon", "procedure P; forward;\nprocedure P; export;", "Syntax Error: BEGIN expected [line: 2, column: 14]"},
		{"EOF without export semicolon", "procedure P; forward;\nprocedure P; export", "Syntax Error: BEGIN expected [line: 2, column: 14]"},
		{"EOF after optional name", "procedure P; forward;\nprocedure P; export 'name'", "Syntax Error: BEGIN expected [line: 2, column: 14]"},
		{"bad optional name", "procedure P; forward;\nprocedure P; export 42; begin Bad; end; {$ERROR 'late'} Later;", "Syntax Error: BEGIN expected [line: 2, column: 14]"},
		{"missing body and earlier diagnostic", "PrintLn(Missing);\nprocedure Pending; forward;\nprocedure P; forward;\nprocedure P; {header}\n export; PrintLn('late'); {$ERROR 'late'} Later;", "Syntax Error: Unknown name \"Missing\" [line: 1, column: 9]\nSyntax Error: BEGIN expected [line: 5, column: 2]"},
		{"matching lone overloaded forward", "procedure P(X: Integer); overload; forward;\nprocedure P(X: Integer); overload;\n export; begin end;", "Syntax Error: BEGIN expected [line: 3, column: 2]"},
		{"different lone overload retains pending forward", "procedure P(X: Integer); overload; forward;\nprocedure P(X: String); overload;\n export; begin end;", "Syntax Error: The function \"P\" was forward declared but not implemented [line: 1, column: 11]"},
		// Return-only overloading retains Go's existing policy, not source parity.
		{"Go return only overload retains pending forward", "function F: Integer; overload; forward;\nfunction F: String; overload;\n export; begin end;", "Syntax Error: The function \"F\" was forward declared but not implemented [line: 1, column: 10]"},
		{"omitted default selects forward", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: Integer); overload;\n export; begin end;", "Syntax Error: BEGIN expected [line: 3, column: 2]"},
		{"parameter mismatch precedes stop", "procedure P(X: Integer); forward;\nprocedure P(X: String);\n export; begin Bad; end;", "Syntax Error: implementation signature for 'P' does not match forward declaration [line: 2, column: 1]\nSyntax Error: BEGIN expected [line: 3, column: 2]"},
		{"return mismatch precedes stop", "function F: Integer; forward;\nfunction F: String;\n export; begin Bad; end;", "Syntax Error: implementation return type for 'F' does not match forward declaration [line: 2, column: 1]\nSyntax Error: BEGIN expected [line: 3, column: 2]"},
		{"default mismatch precedes stop", "procedure P(X: Integer = 1); forward;\nprocedure P(X: Integer);\n export; begin Bad; end;", "Syntax Error: implementation signature for 'P' does not match forward declaration [line: 2, column: 1]\nSyntax Error: BEGIN expected [line: 3, column: 2]"},
		{"export helper bare", "function F(X: Integer): Integer; forward;\nfunction F(X: Integer): Integer;\n export; helper; begin Result := X; end;", "Syntax Error: BEGIN expected [line: 3, column: 2]"},
		{"export helper alias overload", "function F(X: Integer): Integer; overload; forward;\nfunction F(X: String): Integer; overload; forward;\nfunction F(X: Integer): Integer; overload;\n export; helper Alias; begin Result := X; end;", "Syntax Error: BEGIN expected [line: 4, column: 2]"},
		{"helper mismatch precedes stop", "function F(X: Integer): Integer; forward;\nfunction F(X: String): Integer;\n export; helper Alias; begin Bad; end;", "Syntax Error: implementation signature for 'F' does not match forward declaration [line: 2, column: 1]\nSyntax Error: BEGIN expected [line: 3, column: 2]"},
		{"case insensitive", "procedure Pending; forward;\nprocedure pENDING; ExPoRt 'Again'; begin end;", "Syntax Error: BEGIN expected [line: 2, column: 20]"},
		{"overload signature", "procedure P(X: Integer); overload; forward;\nprocedure P(X: String); overload; forward;\nprocedure p(X: INTEGER); overload;\n export 'Again'; begin Bad; end;", "Syntax Error: BEGIN expected [line: 4, column: 2]"},
		{"earlier diagnostic and later cutoff", "PrintLn(Missing);\nprocedure Pending; forward;\ntype TForward = class;\nprocedure P; forward;\nprocedure P; export; begin Bad; end;\n{$ERROR 'late'}\nLater;", "Syntax Error: Unknown name \"Missing\" [line: 1, column: 9]\nSyntax Error: BEGIN expected [line: 5, column: 14]"},
		{"earlier deferred body diagnostic", "procedure Before; begin\nPrintLn(Missing); end;\nprocedure P; forward;\nprocedure P; export; begin Bad; end;", "Syntax Error: Unknown name \"Missing\" [line: 2, column: 9]\nSyntax Error: BEGIN expected [line: 4, column: 14]"},
		{"comments and newline anchor", "procedure P; forward;\nprocedure P; {header}\n  ExPoRt {name}\n 'Again'; begin Bad; end;\nLater;", "Syntax Error: BEGIN expected [line: 3, column: 3]"},
		{"exported forward implementation", "procedure P; forward; export 'PublicP';\nprocedure P; export; begin end;", "Syntax Error: BEGIN expected [line: 2, column: 14]"},
		{"unit interface match", "unit U; interface\nprocedure P;\nimplementation\nprocedure P; export; begin Bad; end;\nend.", "Syntax Error: BEGIN expected [line: 4, column: 14]"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(tt.source, Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
