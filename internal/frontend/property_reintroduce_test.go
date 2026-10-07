package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// These cases catch lost property identity, call-argument checking in place of
// property reads, and a missing closer being incorrectly promoted to a stop.
func TestCompile_ReintroducedPropertyReads(t *testing.T) {
	const prefix = "type TTest = class Field: Integer; property Prop: Integer read Field reintroduce; property Plain: Integer read Field; end;\nvar Obj := new TTest;\n"
	tests := []struct {
		name, source, want string
		hints              semantic.HintsLevel
	}{
		{"bare read", "var A := Obj.Prop;", "", semantic.HintsLevelNormal},
		{"empty brackets", "var A := Obj.Prop();", `Hint: Property "Prop" reintroduced a method, you should remove empty brackets () [line: 3, column: 18]`, semantic.HintsLevelNormal},
		{"disabled hints", "var A := Obj.Prop();", "", semantic.HintsLevelDisabled},
		{"source disables hints", "{$HINTS OFF}\nvar A := Obj.Prop();", "", semantic.HintsLevelNormal},
		{"missing closer", "var A := Obj.Prop(;", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 3, column: 18]\nSyntax Error: \")\" expected [line: 3, column: 19]", semantic.HintsLevelNormal},
		{"recovery keeps declaration", "var A := Obj.Prop(;\nPrintLn(A); Missing;", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 3, column: 18]\nSyntax Error: \")\" expected [line: 3, column: 19]\nSyntax Error: Unknown name \"Missing\" [line: 4, column: 13]", semantic.HintsLevelNormal},
		{"comment newline anchor", "var A := Obj.Prop( {note}\n;", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 3, column: 18]\nSyntax Error: \")\" expected [line: 4, column: 1]", semantic.HintsLevelNormal},
		{"case hint before brackets", "var A := Obj.pRoP();", "Hint: \"pRoP\" does not match case of declaration (\"Prop\") [line: 3, column: 14]\nHint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 3, column: 18]", semantic.HintsLevelPedantic},
		{"ordinary property is not callable", "var A := Obj.Plain();", `Syntax Error: Not a method [line: 3, column: 19]`, semantic.HintsLevelNormal},
		{"ordinary property stops before argument", "var A := Obj.Plain(Missing);\nOther;", `Syntax Error: Not a method [line: 3, column: 19]`, semantic.HintsLevelNormal},
		{"ordinary property boundary stop", "var A := Obj.Plain(;", `Syntax Error: Not a method [line: 3, column: 19]`, semantic.HintsLevelNormal},
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

func TestCompile_ReintroducedPropertyFixtures(t *testing.T) {
	for _, name := range []string{"property_reintroduce1", "property_reintroduce2"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "..", "testdata", "fixtures", "FailureScripts", name)
			source, err := os.ReadFile(path + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(path + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			result := Compile(string(source), path+".pas", semantic.HintsLevelPedantic)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.TrimSpace(string(want)) {
				t.Fatalf("got %q; want %q", got, want)
			}
		})
	}
}

// Retaining an unresolved boundary call must preserve ordinary punctuation and
// must not accidentally turn an incomplete namespace/method call into ().
func TestCompile_PropertyCallBoundaryControls(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"array helper", "var A: array of Integer;\nA.Add(1, ;", `Syntax Error: Expression expected [line: 2, column: 10]`},
		{"ordinary method", "type TTest = class procedure Work(X: Integer); begin end; end;\nvar Obj := new TTest; Obj.Work(; Missing;", `Syntax Error: Expression expected [line: 2, column: 32]`},
		{"Default namespace", "Default.PrintLn(; Missing;", `Syntax Error: Expression expected [line: 1, column: 17]`},
		{"JSON namespace", "JSON.Parse(; Missing;", `Syntax Error: Expression expected [line: 1, column: 12]`},
		{"unresolved receiver", "Unknown.Work(; Missing;", `Syntax Error: Expression expected [line: 1, column: 14]`},
		{"nested ordinary call", "type T = class procedure M; begin end; end;\nvar O := new T;\nPrintLn(O.M(;\nMissing;", `Syntax Error: Expression expected [line: 3, column: 13]`},
		{"JSON receiver", "var J := JSON.NewObject;\nJ.Add(; Missing;", `Syntax Error: Expression expected [line: 2, column: 7]`},
		{"ByteBuffer receiver", "var B := new ByteBuffer;\nB.SetLength(; Missing;", `Syntax Error: Expression expected [line: 2, column: 13]`},
		{"explicit helper", "type H = helper for String function Work: Integer; begin Result := 1; end; end;\nH.Work(; Missing;", `Syntax Error: Expression expected [line: 2, column: 8]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelDisabled)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}
