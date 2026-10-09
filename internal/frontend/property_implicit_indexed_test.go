package frontend

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_ImplicitIndexedCompatibility(t *testing.T) {
	const prefix = "type TTest = class\n function Get(I: Integer): Integer; begin Result := I; end;\n property Prop[I: Integer]: Integer read Get reintroduce; default;\n property Plain[I: Integer]: Integer read Get;\n function Read: Integer;\nend;\nfunction TTest.Read: Integer;\n"
	const hint = "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 8, column: 21]"
	tests := []struct {
		name, body, want string
		hints            semantic.HintsLevel
	}{
		{"normal", "begin Result := Prop()[3]; end;", hint, semantic.HintsLevelNormal},
		{"disabled", "begin Result := Prop()[3]; end;", "", semantic.HintsLevelDisabled},
		{"directive", "{$HINTS OFF}\nbegin Result := Prop()[3]; end;", "", semantic.HintsLevelNormal},
		{"case", "begin Result := pRoP()[3]; end;", "Hint: \"pRoP\" does not match case of declaration (\"Prop\") [line: 8, column: 17]\n" + hint, semantic.HintsLevelPedantic},
		{"comment", "begin Result := Prop {comment}\n()[3]; end;", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 9, column: 1]", semantic.HintsLevelNormal},
		{"ordinary", "begin Result := Plain()[Missing]; Other; end;", "Syntax Error: More arguments expected [line: 8, column: 17]\nSyntax Error: Not a method [line: 8, column: 22]", semantic.HintsLevelNormal},
		{"type", "begin Result := Prop()[True]; end;", hint + "\nSyntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 8, column: 22]", semantic.HintsLevelNormal},
		{"extra", "begin Result := Prop()[1, 2]; end;", hint + "\nSyntax Error: Too many arguments [line: 8, column: 22]", semantic.HintsLevelNormal},
		{"type suppresses count", "begin Result := Prop()[True, 2]; end;", hint + "\nSyntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 8, column: 22]", semantic.HintsLevelNormal},
		{"child precedes count", "begin Result := Prop()[1, Missing]; end;", hint + "\nSyntax Error: Unknown name \"Missing\" [line: 8, column: 27]", semantic.HintsLevelNormal},
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

func TestCompile_ImplicitIndexedAccessorContexts(t *testing.T) {
	const prefix = "type TTest = class\n function Get(I: Integer): Integer; begin Result := I; end; procedure SetValue(I, V: Integer); begin end;\n property Prop[I: Integer]: Integer read Get reintroduce;\n function Read: Integer;\nend;\nfunction TTest.Read: Integer;\n"
	const hint = "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 21]"
	tests := []struct{ name, source, want string }{
		{"write only", strings.Replace(prefix, "read Get reintroduce", "write SetValue reintroduce", 1) + "begin Result := Prop()[Missing]; end;", hint + "\nSyntax Error: Cannot read a write only property [line: 7, column: 22]\nSyntax Error: Array expected [line: 7, column: 23]"},
		{"deprecated casing order", strings.Replace(prefix, "read Get reintroduce;", "read Get reintroduce; deprecated 'old';", 1) + "begin Result := pRoP()[3]; end;", "Hint: \"pRoP\" does not match case of declaration (\"Prop\") [line: 7, column: 17]\nWarning: \"Prop\" has been deprecated: old [line: 7, column: 17]\n" + hint},
		{"static caller", strings.Replace(strings.Replace(prefix, " function Read: Integer;", " class function Read: Integer; static;", 1), "function TTest.Read", "class function TTest.Read", 1) + "begin Result := Prop()[Missing]; Other; end;", "Syntax Error: Object reference needed to read/write an object field [line: 7, column: 17]"},
		{"instance getter in class caller", strings.Replace(strings.Replace(prefix, " function Read", " class function Read", 1), "function TTest.Read", "class function TTest.Read", 1) + "begin Result := Prop()[3]; end;", hint + "\nSyntax Error: Read access of property should be a static method [line: 7, column: 22]\nSyntax Error: Class method or constructor expected [line: 7, column: 22]"},
		{"child precedes class eligibility", strings.Replace(strings.Replace(prefix, " function Read", " class function Read", 1), "function TTest.Read", "class function TTest.Read", 1) + "begin Result := Prop()[Missing]; end;", hint + "\nSyntax Error: Unknown name \"Missing\" [line: 7, column: 24]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(tt.source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_ImplicitIndexedInlineStaticCaller(t *testing.T) {
	const source = `type TTest = class
 class function Get(I: Integer): Integer; begin Result := I; end;
 class property Prop[I: Integer]: Integer read Get reintroduce;
 class function Read: Integer; static;
 begin Result := Prop()[Missing]; Other; end;
end;`
	result := CompileWithOptions(source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: true})
	const want = "Syntax Error: Object reference needed to read/write an object field [line: 5, column: 18]"
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}

func TestCompile_ImplicitIndexedStaticOverloadIsolation(t *testing.T) {
	const source = `type TTest = class
 class function Get(I: Integer): Integer; begin Result := I; end;
 class property Prop[I: Integer]: Integer read Get reintroduce;
 class function Read(I: Integer): Integer; static; overload;
 class function Read(S: String): Integer; overload;
end;
class function TTest.Read(I: Integer): Integer;
begin Result := I; end;
class function TTest.Read(S: String): Integer;
begin Result := Prop()[3]; end;`
	result := CompileWithOptions(source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
	const want = "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 10, column: 21]"
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}
