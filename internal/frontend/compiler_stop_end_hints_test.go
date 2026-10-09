package frontend_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/dwscript"
)

// HintUnusedPrivateSymbols runs after ReadScript and ReadScriptImplementations
// succeed (pinned dwsCompiler.pas:1737). A stop skips those program-end hints,
// while ReadProcBody:4168 retains hints from an earlier completed routine.
func TestCompile_CompilerStopPrivateEndHints(t *testing.T) {
	const field = "type T = class\nprivate\n F: Integer;\nend;\n"
	const method = "type T = class\nprivate\n procedure M; begin end;\nend;\n"
	const fieldHint = "Hint: Private field \"F\" declared but never used [line: 3, column: 2]"
	const methodHint = "Hint: Private method \"M\" declared but never used [line: 3, column: 12]"
	tests := []struct{ name, source, want string }{
		{"field before parser stop", field + "var A := ;", "Syntax Error: Expression expected [line: 5, column: 10]"},
		{"method before parser stop", method + "var A := ;", "Syntax Error: Expression expected [line: 5, column: 10]"},
		{"field without stop", field, fieldHint},
		{"method without stop", method, methodHint},
		{"completed local before parser stop", field + "procedure P; begin var U: Integer; end;\nvar A := ;", "Hint: Variable \"U\" declared but not used [line: 5, column: 24]\nSyntax Error: Expression expected [line: 6, column: 10]"},
		{"semantic special stop", field + "var A := Low;\nMissing;", "Syntax Error: \"(\" expected [line: 5, column: 13]"},
		{"ordinary type punctuation keeps later error", field + "var A := Integer;\nMissing;", "Syntax Error: \"(\" expected [line: 5, column: 17]\nSyntax Error: Unknown name \"Missing\" [line: 6, column: 1]"},
		{"ordinary empty index suppresses end hint", field + "var A: array of Integer;\nvar B := A[];", "Syntax Error: Expression expected [line: 6, column: 12]"},
		{"recovered property empty index keeps later directive", field + "type R = class function Get(I: Integer): Integer; begin Result := I; end; property Prop[I: Integer]: Integer read Get reintroduce; end;\nvar O := R.Create;\nvar A := O.Prop()[];\n{$ERROR 'later'}", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 16]\nSyntax Error: More arguments expected [line: 7, column: 17]\nCompile Error: later [line: 8, column: 3]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := frontend.CompileWithOptions(tt.source, frontend.Options{HintsLevel: semantic.HintsLevelPedantic})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_PrivateEndHintOptions(t *testing.T) {
	const source = "type T = class private F: Integer; procedure M; begin end; end;"
	for _, tt := range []struct {
		name string
		opts frontend.Options
	}{
		{"disabled hints", frontend.Options{HintsLevel: semantic.HintsLevelDisabled}},
		{"normal hints", frontend.Options{HintsLevel: semantic.HintsLevelNormal}},
		{"disabled dictionary", frontend.Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: true}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := frontend.CompileWithOptions(source, tt.opts).DiagnosticStrings(); len(got) != 0 {
				t.Fatalf("got %q; want no diagnostics", got)
			}
		})
	}
}

func TestEngine_CompilerStopSkipsPrivateEndHints(t *testing.T) {
	engine, err := dwscript.New()
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile("type T = class\nprivate\n F: Integer;\nend;\nvar A := ;")
	if program != nil {
		t.Fatal("compiler stop returned a program")
	}
	var compileErr *dwscript.CompileError
	if !errors.As(err, &compileErr) {
		t.Fatalf("got error %v; want CompileError", err)
	}
	if len(compileErr.Errors) != 1 || compileErr.Errors[0].Message != "Expression expected" {
		t.Fatalf("got errors %#v; want only Expression expected", compileErr.Errors)
	}
}
