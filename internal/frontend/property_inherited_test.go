package frontend

import (
	"os"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_InheritedProperty(t *testing.T) {
	const prefix = "type TBase = class Field: Integer; property Prop: Integer read Field reintroduce; property Plain: Integer read Field; end;\ntype TChild = class(TBase) function Probe: Integer; end;\nfunction TChild.Probe: Integer;\n"
	tests := []struct {
		name, body, want string
		hints            semantic.HintsLevel
	}{
		{"normal", "begin Result := inherited Prop(); end;", `Hint: Property "Prop" reintroduced a method, you should remove empty brackets () [line: 4, column: 31]`, semantic.HintsLevelNormal},
		{"pedantic casing", "begin Result := inherited pRoP(); end;", `Hint: Property "Prop" reintroduced a method, you should remove empty brackets () [line: 4, column: 31]`, semantic.HintsLevelPedantic},
		{"disabled", "begin Result := inherited Prop(); end;", "", semantic.HintsLevelDisabled},
		{"directive disabled", "{$HINTS OFF}\nbegin Result := inherited Prop(); end;", "", semantic.HintsLevelNormal},
		{"bare", "begin Result := inherited Prop; end;", "", semantic.HintsLevelPedantic},
		{"ordinary bare", "begin Result := inherited Plain; end;", "", semantic.HintsLevelNormal},
		{"ordinary stop", "begin Result := inherited Plain(); Missing; end;", `Syntax Error: Not a method [line: 4, column: 32]`, semantic.HintsLevelNormal},
		{"comment anchor", "begin Result := inherited Prop {comment}\n(); end;", `Hint: Property "Prop" reintroduced a method, you should remove empty brackets () [line: 5, column: 1]`, semantic.HintsLevelNormal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(prefix+tt.body, Options{Filename: "<test>", HintsLevel: tt.hints, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_InheritedPropertySelectsParentMarker(t *testing.T) {
	const source = "type TBase = class Field: Integer; property Prop: Integer read Field; end;\ntype TChild = class(TBase)\n property Prop: Integer read Field reintroduce;\n function Probe: Integer; begin Result := inherited Prop(); Missing; end;\nend;"
	result := CompileWithOptions(source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
	const want = `Syntax Error: Not a method [line: 4, column: 57]`
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}

func TestCompile_InheritedWriteOnlyProperty(t *testing.T) {
	const prefix = "type TBase = class Field: Integer; property Prop: Integer write Field MARKER end;\ntype TChild = class(TBase) function Probe: Integer; end;\nfunction TChild.Probe: Integer;\n"
	tests := []struct{ name, marker, body, want string }{
		{"flagged empty", "reintroduce;", "begin Result := inherited Prop(); end;", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 4, column: 31]\nSyntax Error: Cannot read a write only property [line: 4, column: 32]"},
		{"flagged bare next token", "reintroduce;", "begin Result := inherited Prop {comment}\n; end;", "Syntax Error: Cannot read a write only property [line: 5, column: 1]"},
		{"ordinary bare", ";", "begin Result := inherited Prop; end;", "Syntax Error: Cannot read a write only property [line: 4, column: 27]"},
		{"ordinary empty", ";", "begin Result := inherited Prop(); Missing; end;", "Syntax Error: Cannot read a write only property [line: 4, column: 27]\nSyntax Error: Not a method [line: 4, column: 31]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := strings.Replace(prefix, "MARKER", tt.marker, 1) + tt.body
			result := CompileWithOptions(source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_InheritedDeprecatedProperty(t *testing.T) {
	const source = "type TBase = class Field: Integer; property Prop: Integer read Field reintroduce; deprecated 'old'; end;\ntype TChild = class(TBase) function Probe: Integer; end;\nfunction TChild.Probe: Integer;\nbegin Result := inherited pRoP(); end;"
	result := CompileWithOptions(source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: true})
	const want = "Warning: \"Prop\" has been deprecated: old [line: 4, column: 27]\nHint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 4, column: 31]"
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}

func TestCompile_InheritedClassPropertySelectedGetterFlags(t *testing.T) {
	source, err := os.ReadFile("../../testdata/property_inherited/invalid/class_getter.dws")
	if err != nil {
		t.Fatal(err)
	}
	result := CompileWithOptions(string(source), Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
	const want = `Syntax Error: Class method or constructor expected [line: 6, column: 36]`
	diagnostics := result.DiagnosticStrings()
	if result.SemanticSuccessful || len(diagnostics) == 0 || diagnostics[0] != want {
		t.Fatalf("selected getter must reject the declaration; got %v, semantic success %v", diagnostics, result.SemanticSuccessful)
	}
}
