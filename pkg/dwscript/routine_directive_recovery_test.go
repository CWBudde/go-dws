package dwscript

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// Source-derived lists from ReadProcDecl/ReadProcBody/ReadInterface; no Pascal oracle.
func TestEngineCompile_RoutineDirectiveRecovery(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"late body and lexer cutoff", "procedure P; inline; export; begin Bad; end; {$ERROR 'late'} Later;", "Syntax Error: BEGIN expected [line: 1, column: 22]"},
		{"prefix before late forward", "procedure P;\nconst X = Missing;\nforward; begin Bad; end; Later;", "Syntax Error: Unknown name \"Missing\" [line: 2, column: 11]"},
		{"late export forward", "procedure P; export; forward; begin Bad; end; Later;", "Syntax Error: There is already a forward declaration of this function [line: 1, column: 22]"},
		{"inline export", "procedure P; inline; export; begin Bad; end; Later;", "Syntax Error: BEGIN expected [line: 1, column: 22]"},
		{"deprecated export", "procedure P; deprecated 'old'; export; begin Bad; end; Later;", "Syntax Error: BEGIN expected [line: 1, column: 32]"},
		{"cdecl export", "procedure P; cdecl; export; begin Bad; end; Later;", "Hint: Call convention \"cdecl\" is not supported and ignored [line: 1, column: 14]\nSyntax Error: BEGIN expected [line: 1, column: 21]"},
		{"late forward body token", "procedure P; inline; forward; begin Bad; end; Later;", "Syntax Error: There is already a forward declaration of this function [line: 1, column: 22]"},
		{"inline cdecl", "procedure P; inline; cdecl; begin end;", "Syntax Error: BEGIN expected [line: 1, column: 22]"},
		{"repeated cdecl", "procedure P; cdecl; cdecl; begin Bad; end;", "Hint: Call convention \"cdecl\" is not supported and ignored [line: 1, column: 14]\nSyntax Error: BEGIN expected [line: 1, column: 21]"},
		{"repeated forward unreached qualifier", "procedure P; forward;\nprocedure P; forward; cdecl; begin Bad; end; Later;", "Syntax Error: There is already a forward declaration of this function [line: 2, column: 14]"},
		{"repeated forward", "procedure P; forward;\nprocedure P; forward; begin Bad; end; Later;", "Syntax Error: There is already a forward declaration of this function [line: 2, column: 14]"},
		{"late export EOF", "procedure P; inline; export;", "Syntax Error: BEGIN expected [line: 1, column: 22]"},
		{"ordinary EOF", "procedure P;", "Syntax Error: BEGIN expected [line: 1, column: 12]"},
		{"qualified implementation", "type T = class procedure P; end;\nprocedure T.P; export; begin end;", "Syntax Error: BEGIN expected [line: 2, column: 16]"},
		{"local const prefix", "procedure P;\nconst X = Missing;\nexport; begin Bad; end; Later;", "Syntax Error: Unknown name \"Missing\" [line: 2, column: 11]"},
		{"earlier definite stop", "var x := ;\nprocedure P; inline; export; begin Bad; end; {$ERROR 'late'}", "Syntax Error: Expression expected [line: 1, column: 10]"},
		{"earlier hint", "procedure Before; cdecl; begin end;\nprocedure P; inline; export; begin Bad; end;", "Hint: Call convention \"cdecl\" is not supported and ignored [line: 1, column: 19]\nSyntax Error: BEGIN expected [line: 2, column: 22]"},
		{"interface export", "type T = interface procedure P; export; end;", "Syntax Error: END expected [line: 1, column: 33]"},
		{"interface inline", "type T = interface procedure P; inline; end;", "Syntax Error: END expected [line: 1, column: 33]"},
		{"interface overload", "type T = interface procedure P; overload; end;", "Syntax Error: END expected [line: 1, column: 33]"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := compilePropertyUseSite(t, tt.source, "", semantic.HintsLevelPedantic)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestEngineCompile_RoutineRecoveryLegalContexts(t *testing.T) {
	for _, source := range []string{
		"type T = interface procedure P; end;",
		"type T = class procedure P; inline; end; procedure T.P; begin end;",
		"type T = record procedure P; inline; end; procedure T.P; begin end;",
		"type T = helper for Integer procedure P; inline; end; procedure T.P; begin end;",
		"procedure P; forward; procedure P; begin end;",
		"procedure P; external;",
		"procedure P; external; cdecl;",
		"procedure P; forward; cdecl; procedure P; begin end;",
		"unit U; interface procedure P; inline; implementation procedure P; begin end; end.",
		"procedure P; forward; export; procedure P; begin end;",
		"procedure P; cdecl; inline; deprecated 'old'; begin end;",
		"procedure P; const X = 1; var Y: Integer := X; begin PrintLn(Y); end;",
		"procedure P; require True; const X = 1; var Y: Integer := X; begin PrintLn(Y); end;",
	} {
		t.Run(source, func(t *testing.T) {
			engine, err := New()
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(source)
			if err != nil || program == nil {
				t.Fatalf("legal context: program %v error %v", program, err)
			}
		})
	}
}

// Prefix lists derive from the pinned reader. Duplicate/mismatch and complete-body
// unknown-name lists retain the existing Go diagnostic policy.
func TestEngineCompile_RoutineRecoveryScopeControls(t *testing.T) {
	for _, tt := range []struct {
		name, source, want string
		stop               bool
	}{
		{"var prefix", "procedure P;\nvar X: Integer := Missing;\nexport; begin Bad; end; Later;", "Syntax Error: Unknown name \"Missing\" [line: 2, column: 19]", true},
		{"contract prefix", "procedure P;\nrequire Missing;\nvar X: Integer;\nexport; begin Bad; end; Later;", "Syntax Error: Unknown name \"Missing\" [line: 2, column: 9]", true},
		{"valid prefix scope", "procedure P;\nconst X = 1;\nvar Y: Integer := X;\nexport; begin Bad; end; Later;", "Syntax Error: BEGIN expected [line: 4, column: 1]", true},
		{"valid contract prefix", "procedure P;\nrequire True;\nconst X = 1;\nexport; begin Bad; end; Later;", "Warning: Constant condition [line: 2, column: 9]\nSyntax Error: BEGIN expected [line: 4, column: 1]", true},
		{"ordinary complete body unknown unchanged", "procedure P; begin PrintLn(Missing); end;\nLater;", "Syntax Error: Unknown name \"Missing\" [line: 1, column: 28]\nSyntax Error: Unknown name \"Later\" [line: 2, column: 1]", false},
		{"genuine qualified duplicate", "type T = class procedure P; begin end; end;\nprocedure T.P; export; begin end;", "Syntax Error: There is already a method with name \"P\" [line: 2, column: 14]\nSyntax Error: BEGIN expected [line: 2, column: 16]", true},
		{"qualified mismatched signature", "type T = class procedure P(X: Integer); end;\nprocedure T.P(X: String); export; begin end;", "Syntax Error: method 'T.P' parameter 1 has type String in implementation, but Integer in declaration [line: 2, column: 1]\nSyntax Error: BEGIN expected [line: 2, column: 27]", true},
		{"ordinary duplicate remains nonstopping", "procedure P; begin end;\nprocedure P; begin end;\nLater;", "Syntax Error: There is already a method with name \"P\" [line: 2, column: 12]\nSyntax Error: Unknown name \"Later\" [line: 3, column: 1]", false},
		{"nested prefix", "procedure Outer; begin\nprocedure Inner;\nconst X = Missing;\nexport; begin Bad; end; end; Later;", "Syntax Error: Unknown name \"Missing\" [line: 3, column: 11]", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := compilePropertyUseSite(t, tt.source, "", semantic.HintsLevelPedantic)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

// Completed prefixes require BEGIN; contracts_unfinished4 keeps its semicolon-only list.
func TestEngineCompile_RoutineContractEOFRecovery(t *testing.T) {
	for _, tt := range []struct {
		name, source, want string
		stop               bool
	}{
		{"completed require EOF", "procedure P; require True;", "Warning: Constant condition [line: 1, column: 22]\nSyntax Error: BEGIN expected [line: 1, column: 26]", true},
		{"completed require comments EOF", "procedure P;\nrequire True; {comment}\n// trailing\n", "Warning: Constant condition [line: 2, column: 9]\nSyntax Error: BEGIN expected [line: 2, column: 13]", true},
		{"completed require const EOF", "procedure P;\nrequire True;\nconst X = 1;", "Warning: Constant condition [line: 2, column: 9]\nSyntax Error: BEGIN expected [line: 3, column: 12]", true},
		{"completed require var EOF", "procedure P;\nrequire True;\nvar Y: Integer := 1;", "Warning: Constant condition [line: 2, column: 9]\nSyntax Error: BEGIN expected [line: 3, column: 20]", true},
		{"completed require const var EOF", "procedure P;\nrequire True;\nconst X = 1;\nvar Y: Integer := X;", "Warning: Constant condition [line: 2, column: 9]\nSyntax Error: BEGIN expected [line: 4, column: 20]", true},
		{"reached convention require EOF", "procedure P; cdecl; require True;", "Hint: Call convention \"cdecl\" is not supported and ignored [line: 1, column: 14]\nWarning: Constant condition [line: 1, column: 29]\nSyntax Error: BEGIN expected [line: 1, column: 33]", true},
		{"exported unfinished require still needs body", "procedure Test(i : Integer); export;\nrequire\n   i>0", "Syntax Error: \";\" expected [line: 3, column: 6]\nSyntax Error: BEGIN expected [line: 3, column: 6]", true},
		{"exported require EOF", "procedure P; export; require True;", "Warning: Constant condition [line: 1, column: 30]\nSyntax Error: BEGIN expected [line: 1, column: 34]", true},
		{"completed prefix unknown EOF", "procedure P;\nrequire Missing;\nvar X: Integer;", "Syntax Error: Unknown name \"Missing\" [line: 2, column: 9]", true},
		{"unfinished contract preserved", "procedure Test(i : Integer);\nrequire\n   i>0", "Syntax Error: \";\" expected [line: 3, column: 6]", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := compilePropertyUseSite(t, tt.source, "", semantic.HintsLevelPedantic)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
