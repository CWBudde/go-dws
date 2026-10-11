package frontend

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_IndexedPropertyCompatibility(t *testing.T) {
	const prefix = "type TTest = class\n function Get(I: Integer): Integer; begin Result := I; end;\n property Prop[I: Integer]: Integer read Get reintroduce; default;\n property Plain[I: Integer]: Integer read Get;\nend;\nvar Obj := new TTest;\n"
	tests := []struct {
		name, source, want string
		hints              semantic.HintsLevel
	}{
		{"empty", "var A := Obj.Prop()[3];", `Hint: Property "Prop" reintroduced a method, you should remove empty brackets () [line: 7, column: 18]`, semantic.HintsLevelNormal},
		{"disabled", "var A := Obj.Prop()[3];", "", semantic.HintsLevelDisabled},
		{"directive", "{$HINTS OFF}\nvar A := Obj.Prop()[3];", "", semantic.HintsLevelNormal},
		{"case", "var A := Obj.pRoP()[3];", "Hint: \"pRoP\" does not match case of declaration (\"Prop\") [line: 7, column: 14]\nHint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 18]", semantic.HintsLevelPedantic},
		{"comment anchor", "var A := Obj.Prop {comment}\n()[3];", `Hint: Property "Prop" reintroduced a method, you should remove empty brackets () [line: 8, column: 1]`, semantic.HintsLevelNormal},
		{"ordinary read", "var A := Obj.Plain[3];", "", semantic.HintsLevelNormal},
		{"default read", "var A := Obj[3];", "", semantic.HintsLevelNormal},
		{"unpaired read", "var A := Obj.Prop[3];", "", semantic.HintsLevelNormal},
		{"ordinary call", "var A := Obj.Plain()[Missing]; Other;", "Syntax Error: More arguments expected [line: 7, column: 14]\nSyntax Error: Not a method [line: 7, column: 19]", semantic.HintsLevelNormal},
		{"trailing pair", "var A := Obj.Prop[3](); Other;", `Syntax Error: Not a method [line: 7, column: 21]`, semantic.HintsLevelNormal},
		{"type", "var A := Obj.Prop()[True];", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 18]\nSyntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 7, column: 19]", semantic.HintsLevelNormal},
		{"extra", "var A := Obj.Prop()[1, 2];", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 18]\nSyntax Error: Too many arguments [line: 7, column: 19]", semantic.HintsLevelNormal},
		{"type before count", "var A := Obj.Prop()[True, 2];", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 18]\nSyntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 7, column: 19]", semantic.HintsLevelNormal},
		{"child before count", "var A := Obj.Prop()[1, Missing];", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 18]\nSyntax Error: Unknown name \"Missing\" [line: 7, column: 24]", semantic.HintsLevelNormal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(prefix+tt.source, Options{Filename: "<test>", HintsLevel: tt.hints, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_IndexedCompatibilityMultipleArguments(t *testing.T) {
	const prefix = "type TTest = class function Get(I, J: Integer): Integer; begin Result := I+J; end; property Pair[I, J: Integer]: Integer read Get reintroduce; end;\nvar Obj := new TTest;\n"
	const hint = "Hint: Property \"Pair\" reintroduced a method, you should remove empty brackets () [line: 3, column: 18]\n"
	tests := []struct{ name, body, want string }{
		{"missing", "var A := Obj.Pair()[1];", "Syntax Error: More arguments expected [line: 3, column: 19]"},
		{"first type suppresses missing", "var A := Obj.Pair()[True];", "Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 19]"},
		{"second type suppresses excess", "var A := Obj.Pair()[1, True, 3];", "Syntax Error: Argument 1 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 19]"},
		{"two types retain order", "var A := Obj.Pair()[True, False];", "Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 19]\nSyntax Error: Argument 1 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 19]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(prefix+tt.body, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != hint+tt.want {
				t.Fatalf("got %q; want %q", got, hint+tt.want)
			}
		})
	}
}
