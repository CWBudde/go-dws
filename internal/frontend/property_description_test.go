package frontend

import (
	"os"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_PropertyDescriptionFixture(t *testing.T) {
	source, err := os.ReadFile("../../testdata/fixtures/FailureScripts/property_description1.pas")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../../testdata/fixtures/FailureScripts/property_description1.txt")
	if err != nil {
		t.Fatal(err)
	}
	result := Compile(string(source), "property_description1.pas", semantic.HintsLevelPedantic)
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.TrimSpace(string(want)) {
		t.Fatalf("complete diagnostics:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// ReadPropertyDecl leaves the invalid value for ReadSemiColon and the class
// member loop. ReadNameList rejects numbers; END closes the class; EOF retains
// the hot position at DESCRIPTION. These are full lists, including recovery.
func TestCompile_PropertyDescriptionRecovery(t *testing.T) {
	const prefix = "type T = class\n F: Integer;\n property P: Integer read F description"
	tests := []struct{ name, source, want string }{
		{"semicolon", prefix + ";\nend;", `Syntax Error: String expected [line: 3, column: 40]`},
		{"end", prefix + " end;\nend;", `Syntax Error: String expected [line: 3, column: 41]
Syntax Error: ";" expected [line: 3, column: 41]
Syntax Error: Unexpected END [line: 4, column: 1]`},
		{"number stops later declaration", prefix + " 42;\nend;\nMissing;", `Syntax Error: String expected [line: 3, column: 41]
Syntax Error: ";" expected [line: 3, column: 41]
Syntax Error: Name expected [line: 3, column: 41]`},
		{"name remains field", prefix + " Other;\nend;", `Syntax Error: String expected [line: 3, column: 41]
Syntax Error: ";" expected [line: 3, column: 41]
Syntax Error: Colon ":" expected [line: 3, column: 46]`},
		{"EOF hot position", prefix, `Syntax Error: String expected [line: 3, column: 29]
Syntax Error: ";" expected [line: 3, column: 29]
Syntax Error: Name expected [line: 3, column: 29]`},
		{"comments and newlines", prefix + " {comment}\n //next\n end;", `Syntax Error: String expected [line: 5, column: 2]
Syntax Error: ";" expected [line: 5, column: 2]`},
		{"earlier error and later stop", "Missing;\n" + prefix + ";\n 42;\nend;\nLater;", `Syntax Error: Unknown name "Missing" [line: 1, column: 1]
Syntax Error: String expected [line: 4, column: 40]
Syntax Error: Name expected [line: 5, column: 2]`},
		{"missing semicolon retains property", "type T = class\n F: Integer;\n property P: Integer read F description 'ok'\nend;\nvar O := new T; PrintLn(O.P);", `Syntax Error: ";" expected [line: 4, column: 1]`},
		{"root END after earlier stop", "var X := ; end;", `Syntax Error: Expression expected [line: 1, column: 10]`},
		{"root END is ordinary", "end; Missing;", `Syntax Error: Unexpected END [line: 1, column: 1]
Syntax Error: Unknown name "Missing" [line: 1, column: 6]`},
		{"unrelated numeric member stops", "type T = class 42; end; Missing;", `Syntax Error: Name expected [line: 1, column: 16]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(tt.source, Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_PropertyDescriptionReadonlyFieldFixture(t *testing.T) {
	source, err := os.ReadFile("../../testdata/fixtures/SimpleScripts/readonly_field.pas")
	if err != nil {
		t.Fatal(err)
	}
	result := Compile(string(source), "readonly_field.pas", semantic.HintsLevelPedantic)
	if diagnostics := result.DiagnosticStrings(); len(diagnostics) != 0 {
		t.Fatalf("complete diagnostics: %v", diagnostics)
	}
}

func TestCompile_PropertyDescriptionRecordMethodStopFixture(t *testing.T) {
	for _, name := range []string{"record_method_missing_begin", "record_recursive3"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("../../testdata/fixtures/FailureScripts/" + name + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("../../testdata/fixtures/FailureScripts/" + name + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			result := Compile(string(source), name+".pas", semantic.HintsLevelPedantic)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.TrimSpace(string(want)) {
				t.Fatalf("complete diagnostics:\ngot %s\nwant %s", got, want)
			}
		})
	}
}

func TestCompile_PropertyDescriptionRecordStopBoundaries(t *testing.T) {
	const prefix = "procedure Pending; forward;\ntype TForward = class;\n"
	tests := []struct{ name, source, want string }{
		{"earlier reached lexer diagnostic survives", "{$ERROR 'earlier'}\ntype R = record X: Integer; procedure M PrintLn(X); end; end;\n{$ERROR 'later'}", `Compile Error: earlier [line: 1, column: 3]
Syntax Error: ";" expected [line: 2, column: 41]
Syntax Error: Record fields must be declared before record methods [line: 2, column: 41]`},
		{"malformed header retains earlier error", prefix + "type R = record X: Integer; procedure M PrintLn(X); end; end;\n{$ERROR 'later'}", `Syntax Error: ";" expected [line: 3, column: 41]
Syntax Error: Record fields must be declared before record methods [line: 3, column: 41]`},
		{"misplaced field stops", prefix + "type R = record procedure M; begin end; X: Integer; end;\nend; {$ERROR 'later'}", `Syntax Error: Record fields must be declared before record methods [line: 3, column: 41]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_PropertyDescriptionClassExternalField(t *testing.T) {
	const source = `type T = class
 F: Integer; external "field"; readonly;
 property P: Integer read F description "text";
end;
var O := new T; PrintLn(O.P);`
	result := Compile(source, "<test>", semantic.HintsLevelPedantic)
	if got := result.DiagnosticStrings(); len(got) != 0 {
		t.Fatalf("complete diagnostics: %v", got)
	}
}

func TestCompile_PropertyDescriptionExternalFieldRecovery(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"missing string", "type T = class F: Integer; external; end;", `Syntax Error: String expected [line: 1, column: 36]`},
		{"number stays unconsumed", "type T = class F: Integer; external 42; end; Later;", `Syntax Error: String expected [line: 1, column: 37]
Syntax Error: ";" expected [line: 1, column: 37]
Syntax Error: Name expected [line: 1, column: 37]`},
		{"missing semicolon still reads readonly", `type T = class F: Integer; external "field" readonly; G: Integer; end;`, `Syntax Error: ";" expected [line: 1, column: 45]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}
