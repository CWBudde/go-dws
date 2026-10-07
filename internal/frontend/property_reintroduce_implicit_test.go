package frontend

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_ImplicitReintroducedProperty(t *testing.T) {
	const prefix = "type TTest = class Field: Integer; property Prop: Integer read Field reintroduce; property Plain: Integer read Field; function Read: Integer; end;\nfunction TTest.Read: Integer;\n"
	tests := []struct {
		name, body, want string
		hints            semantic.HintsLevel
	}{
		{"normal", "begin Result := Prop(); end;", `Hint: Property "Prop" reintroduced a method, you should remove empty brackets () [line: 3, column: 21]`, semantic.HintsLevelNormal},
		{"pedantic", "begin Result := pRoP(); end;", "Hint: \"pRoP\" does not match case of declaration (\"Prop\") [line: 3, column: 17]\nHint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 3, column: 21]", semantic.HintsLevelPedantic},
		{"disabled", "begin Result := pRoP(); end;", "", semantic.HintsLevelDisabled},
		{"directive disabled", "{$HINTS OFF}\nbegin Result := pRoP(); end;", "", semantic.HintsLevelPedantic},
		{"comment anchor", "begin Result := Prop {comment}\n(); end;", `Hint: Property "Prop" reintroduced a method, you should remove empty brackets () [line: 4, column: 1]`, semantic.HintsLevelNormal},
		{"bare read", "begin Result := Prop; end;", "", semantic.HintsLevelNormal},
		{"ordinary scalar", "begin Result := Plain(); Missing; end;", `Syntax Error: Not a method [line: 3, column: 22]`, semantic.HintsLevelNormal},
		{"grouped scalar", "begin Result := (Prop)(); end;", `Syntax Error: Not a method [line: 3, column: 23]`, semantic.HintsLevelNormal},
		{"write only", "begin Result := WriteOnly(); end;", "Hint: Property \"WriteOnly\" reintroduced a method, you should remove empty brackets () [line: 3, column: 26]\nSyntax Error: Cannot read a write only property [line: 3, column: 17]", semantic.HintsLevelNormal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourcePrefix := prefix
			if tt.name == "write only" {
				sourcePrefix = strings.Replace(prefix, "property Plain: Integer read Field;", "property WriteOnly: Integer write Field reintroduce;", 1)
			}
			result := CompileWithOptions(sourcePrefix+tt.body, Options{Filename: "<test>", HintsLevel: tt.hints, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_StaticReintroducedProperty(t *testing.T) {
	const source = "type TTest = class\n class var Field: Integer; class property Prop: Integer read Field reintroduce;\n class function Read: Integer; static; begin Result := Prop(); Missing; end;\nend;"
	result := CompileWithOptions(source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
	const want = `Syntax Error: Object reference needed to read/write an object field [line: 3, column: 60]`
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}
