package dwscript

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// These compile-only controls catch lost record index grammar and mode checks
// without depending on record instance indexing execution.
func TestEngineCompile_RecordIndexDeclarations(t *testing.T) {
	for _, source := range []string{
		`type TR = record
 class function Get(var x,y: Integer; const z: Integer; w: Integer): Integer; begin Result := 0; end;
 class property P[var a,b: Integer; const c: Integer; d: Integer]: Integer read Get; default;
 property Q[var renamed,other: Integer; const amount: Integer; last: Integer]: Integer read P;
end;`,
		`type TI = Integer; type TR = record
 function Get(const x: TI): Integer;
 property P[const a: Integer]: Integer read Get;
end;
function TR.Get(const renamed: TI): Integer; begin Result := 0; end;`,
		`var r: record
 function Get(var x: Integer): Integer; begin Result := 0; end;
 property P[var a: Integer]: Integer read Get;
end;`,
		`var r := record
 function Get(const x: Integer): Integer; begin Result := 0; end;
 property P[const a: Integer]: Integer read Get;
end;`,
	} {
		result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
		if got := result.DiagnosticStrings(); len(got) != 0 {
			t.Fatalf("compile declaration:\n%s\n%v", source, got)
		}
	}
}

func TestEngineCompile_RecordIndexModeMatrix(t *testing.T) {
	modes := []string{"", "var ", "const "}
	details := [][]string{{"", "Value", "Value"}, {"Var", "", "Var"}, {"Const", "Value", ""}}
	for pi, property := range modes {
		for ai, accessor := range modes {
			t.Run(fmt.Sprintf("%d/%d", pi, ai), func(t *testing.T) {
				source := "type TR = record\n function Get(" + accessor + "Different: Integer): Integer; begin Result := 0; end;\n procedure Put(" + accessor + "Other: Integer; const V: Integer); begin end;\n property P[" + property + "a: Integer]: Integer read Get write Put;\nend;"
				var want []string
				if mode := details[pi][ai]; mode != "" {
					readColumn := 42 + len(property)
					writeColumn := 52 + len(property)
					for i, name := range []string{"Get", "Put"} {
						column := readColumn
						if i == 1 {
							column = writeColumn
						}
						want = append(want, fmt.Sprintf("Syntax Error: Parameter 0 (a) - %s-parameter expected [line: 4, column: %d]", mode, column), fmt.Sprintf("Syntax Error: Method %q has incompatible parameters [line: 4, column: %d]", name, column))
					}
				}
				result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
				if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
					t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
				}
			})
		}
	}
}

func TestEngineCompile_RecordIndexSignaturePriority(t *testing.T) {
	for _, tt := range []struct {
		name, method, property string
		want                   []string
	}{
		{"typeBeforeMode", `function Get(const x: String): Integer; begin Result := 0; end;`, `property P[var a: Integer]: Integer read Get;`, []string{`Parameter 0 - Type "Integer" expected (instead of "String")`, `Method "Get" has incompatible parameters`}},
		{"count", `function Get(x: Integer): Integer; begin Result := 0; end;`, `property P[a,b: Integer]: Integer read Get;`, []string{`Method "Get" has incompatible parameters`}},
		{"resultBeforeParameters", `function Get(var x: String): String; begin Result := ''; end;`, `property P[a: Integer]: Integer read Get;`, []string{`Field/method "Get" has an incompatible type`}},
		{"default", `function Get(x: Integer = 1): Integer; begin Result := 0; end;`, `property P[a: Integer]: Integer read Get;`, []string{`Parameter 0 (a) - default value at implementation does not match declaration or forward`, `Method "Get" has incompatible parameters`}},
		{"modeBeforeDefault", `function Get(x: Integer = 1): Integer; begin Result := 0; end;`, `property P[const a: Integer]: Integer read Get;`, []string{`Parameter 0 (a) - Const-parameter expected`, `Method "Get" has incompatible parameters`}},
		{"setterValueMode", `procedure Put(x: Integer; var V: Integer); begin end;`, `property P[a: Integer]: Integer write Put;`, nil},
		{"setterValueType", `procedure Put(x: Integer; V: String); begin end;`, `property P[a: Integer]: Integer write Put;`, []string{`Method "Put" has incompatible parameters`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "type TR = record\n " + tt.method + "\n " + tt.property + "\nend;"
			result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
			column := strings.Index(tt.property, "Get") + 4 + 1
			if strings.Contains(tt.property, "Put") {
				column = strings.Index(tt.property, "Put") + 4 + 1
			}
			var want []string
			for _, m := range tt.want {
				want = append(want, fmt.Sprintf("Syntax Error: %s [line: 3, column: %d]", m, column))
			}
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
				t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
			}
		})
	}
}

func TestEngineCompile_RecordIndexForwardingValidation(t *testing.T) {
	for _, form := range []string{"named", "inline", "anonymous"} {
		t.Run(form, func(t *testing.T) {
			body := `function Get(var x: Integer): Integer; begin Result := 0; end;
 procedure Put(var x: Integer; v: Integer); begin end;
 property P[var a: Integer]: Integer read Get write Put;
 property Q[var renamed: Integer]: Integer read P write P;
 property R[const last: Integer]: Integer read Q write Q;`
			source := "type TR = record\n " + body + "\nend;"
			if form == "inline" {
				source = "var r: record\n " + body + "\nend;"
			}
			if form == "anonymous" {
				source = "var r := record\n " + body + "\nend;"
			}
			want := []string{
				`Syntax Error: Parameter 0 (last) - Value-parameter expected [line: 6, column: 49]`,
				`Syntax Error: Method "Get" has incompatible parameters [line: 6, column: 49]`,
				`Syntax Error: Parameter 0 (last) - Value-parameter expected [line: 6, column: 57]`,
				`Syntax Error: Method "Put" has incompatible parameters [line: 6, column: 57]`,
			}
			result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
				t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
			}
		})
	}
}

func TestEngineCompile_RecordIndexUnknownTypes(t *testing.T) {
	for _, tt := range []struct {
		source string
		column int
	}{
		{"type TR = record property P[a: Missing]: Integer read (1); end;", 32},
		{"var r: record property P[a: Missing]: Integer read (1); end;", 29},
		{"var r := record property P[a: Missing]: Integer read (1); end;", 31},
	} {
		result := compilePropertyUseSite(t, tt.source, "<test>", semantic.HintsLevelNormal)
		want := fmt.Sprintf(`Syntax Error: Unknown name "Missing" [line: 1, column: %d]`, tt.column+7)
		if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestEngineCompile_RecordIndexStops(t *testing.T) {
	for _, tt := range []struct {
		name, list, message string
		column              int
	}{
		{"colon", "alpha", `Colon ":" expected`, 34},
		{"mixed", "var const a: Integer", `Name expected`, 33},
		{"repeat", "var var a: Integer", `Name expected`, 33},
		{"comma", "i: Integer, j: Integer", `"]" expected`, 39},
		{"default", "i: Integer = 1", `"]" expected`, 40},
		{"lazy", "lazy a: Integer", `Name expected`, 29},
		{"constRef", "const(ref) a: Integer", `Name expected`, 34},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "type TR = record property P[" + tt.list + "]: Integer read (1); end; Missing;"
			result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
			want := fmt.Sprintf("Syntax Error: %s [line: 1, column: %d]", tt.message, tt.column)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestEngineCompile_RecordIndexAccessorKind(t *testing.T) {
	for i, tt := range []struct{ body, want string }{
		{`function Put(x: Integer; v: Integer): Integer; begin Result := 0; end;
 property P[a: Integer]: Integer write Put;`, `Syntax Error: Procedure expected [line: 3, column: 43]`},
		{`F: Integer;
 property P[a: Integer]: Integer read F;`, `Syntax Error: Function expected [line: 3, column: 40]`},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			result := compilePropertyUseSite(t, "type TR = record\n "+tt.body+"\nend;", "<test>", semantic.HintsLevelNormal)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEngineCompile_RecordIndexRecovery(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"emptyOrdinary", "type TR = record\n property Dummy[]: Integer read (1);\nend;\nMissing;", []string{
			`Syntax Error: Parameters expected [line: 2, column: 17]`,
			`Syntax Error: Unknown name "Missing" [line: 4, column: 1]`,
		}},
		{"emptyThenStop", "type TR = record\n property Dummy[]: Integer read (1);\n property Test[alpha]: Integer read (1);\nend;\nMissing; {$ERROR 'unread'}", []string{
			`Syntax Error: Parameters expected [line: 2, column: 17]`,
			`Syntax Error: Colon ":" expected [line: 3, column: 21]`,
		}},
		{"eofBracket", "type TR = record property P[a: Integer", []string{`Syntax Error: "]" expected [line: 1, column: 32]`}},
		{"ordinarySignature", "type TR = record\n function Get(x: Integer): Integer; begin Result := 0; end;\n property P[var a: Integer]: Integer read Get;\nend;\nMissing;", []string{
			`Syntax Error: Parameter 0 (a) - Var-parameter expected [line: 3, column: 46]`,
			`Syntax Error: Method "Get" has incompatible parameters [line: 3, column: 46]`,
			`Syntax Error: Unknown name "Missing" [line: 5, column: 1]`,
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := compilePropertyUseSite(t, tt.source, "<test>", semantic.HintsLevelNormal)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(tt.want, "\n") {
				t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(tt.want, "\n"))
			}
		})
	}
}

func TestEngineCompile_RecordIndexExpressionForwarding(t *testing.T) {
	source := `type TR = record
 F: Integer;
 property P[var a: Integer]: Integer read (1) write (F := Value);
 property Q[const b: Integer]: Integer read P write P;
end;`
	want := []string{
		`Syntax Error: Parameter 0 (b) - Value-parameter expected [line: 4, column: 46]`,
		`Syntax Error: Method "" has incompatible parameters [line: 4, column: 46]`,
		`Syntax Error: Parameter 0 (b) - Value-parameter expected [line: 4, column: 54]`,
		`Syntax Error: Method "" has incompatible parameters [line: 4, column: 54]`,
	}
	result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
	}
}

func TestEngineCompile_RecordIndexFieldTypePriority(t *testing.T) {
	source := `type TR = record
 F: String;
 property P[a: Integer]: Integer read F;
end;`
	want := `Syntax Error: Field/method "F" has an incompatible type [line: 3, column: 40]`
	result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestEngineRun_RecordIndexUncheckedUnknownType(t *testing.T) {
	for _, source := range []string{
		"type TR = record property P[a: Missing]: Integer read (1); end;",
		"var r: record property P[a: Missing]: Integer read (1); end;",
		"var r := record property P[a: Missing]: Integer read (1); end;",
	} {
		t.Run(source, func(t *testing.T) {
			engine, err := New(WithTypeCheck(false))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(source)
			if err != nil || program == nil {
				t.Fatalf("unchecked Compile: program %v, error %v", program, err)
			}
			_, err = engine.Run(program)
			if err == nil || !strings.Contains(err.Error(), "record property index parameter 'a'") {
				t.Fatalf("unknown type accepted by Run: %v", err)
			}
		})
	}
}

func TestEngineCompile_RecordIndexGetterSubtype(t *testing.T) {
	source := `type TB = class end;
 type TC = class(TB) end;
 type TR = record
 function Get(x: Integer): TC; begin Result := nil; end;
 property P[a: Integer]: TB read Get;
 end;`
	result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
	if got := result.DiagnosticStrings(); len(got) > 0 {
		t.Fatalf("getter subtype rejected: %v", got)
	}
}

func TestEngineRun_RecordIndexInlineZeroValues(t *testing.T) {
	for _, checked := range []bool{false, true} {
		for _, source := range []string{
			`var r: record F: Integer; property P: Integer read F; end; PrintLn(r.F);`,
			`var r: record F: Integer; function Get(i: Integer): Integer; begin Result := i; end; property P[a: Integer]: Integer read Get; end; PrintLn(r.F);`,
		} {
			var output bytes.Buffer
			engine, err := New(WithTypeCheck(checked), WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(source)
			if err != nil || program == nil {
				t.Fatalf("Compile: program %v, error %v", program, err)
			}
			_, err = engine.Run(program)
			if err != nil {
				t.Fatal(err)
			}
			if output.String() != "0\n" {
				t.Fatalf("lost record zero value: %q", output.String())
			}
		}
	}
}

func TestEngineCompile_RecordIndexMissingTypeRecovery(t *testing.T) {
	source := "type TR = record\n property P[a:]: Integer read (1);\nend;\nMissing;"
	want := []string{
		`Syntax Error: Type expected [line: 2, column: 15]`,
		`Syntax Error: Unknown name "Missing" [line: 4, column: 1]`,
	}
	result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
	}
}

func TestEngineCompile_RecordIndexUnknownRecovery(t *testing.T) {
	source := `type TR = record
 function Get(x: Integer): Integer; begin Result := 0; end;
 property P[var a,b: Missing]: Integer read Get;
end;`
	want := []string{
		`Syntax Error: Unknown name "Missing" [line: 3, column: 29]`,
		`Syntax Error: Method "Get" has incompatible parameters [line: 3, column: 48]`,
	}
	result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
	}
}

func TestEngineCompile_RecordIndexUnknownBeforeStop(t *testing.T) {
	for _, prefix := range []string{"type TR = record", "var r: record", "var r := record"} {
		t.Run(prefix, func(t *testing.T) {
			source := prefix + "\n property P[a: Missing : Integer read (1);\nend;\nTail; {$ERROR 'unread'}"
			want := []string{
				`Syntax Error: Unknown name "Missing" [line: 2, column: 23]`,
				`Syntax Error: "]" expected [line: 2, column: 24]`,
			}
			result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
				t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
			}
		})
	}
}

func TestEngineCompile_RecordIndexDiagnosticOrder(t *testing.T) {
	source := `type TR = record
 function Get(x: String; var y: Integer): Integer; begin Result := 0; end;
 procedure Put(var x: Integer; y: Integer; v: Integer); begin end;
 property P[var a: Integer; const b: Integer]: Integer read Get write Put;
end;`
	want := []string{
		`Syntax Error: Parameter 0 - Type "Integer" expected (instead of "String") [line: 4, column: 64]`,
		`Syntax Error: Parameter 1 (b) - Value-parameter expected [line: 4, column: 64]`,
		`Syntax Error: Method "Get" has incompatible parameters [line: 4, column: 64]`,
		`Syntax Error: Parameter 1 (b) - Const-parameter expected [line: 4, column: 74]`,
		`Syntax Error: Method "Put" has incompatible parameters [line: 4, column: 74]`,
	}
	result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
	}
}

func TestEngineCompile_RecordIndexPriorDiagnostics(t *testing.T) {
	for _, prefix := range []string{"type TR = record", "var r: record", "var r := record"} {
		for _, tt := range []struct {
			name, body string
			want       []string
		}{
			{"type", " function Get(x: String): Integer; begin Result := 0; end;\n property P[var a: Integer]: Integer read Get;", []string{
				`Syntax Error: Parameter 0 - Type "Integer" expected (instead of "String") [line: 3, column: 46]`,
				`Syntax Error: Method "Get" has incompatible parameters [line: 3, column: 46]`,
			}},
			{"mode", " function Get(var x: Integer): Integer; begin Result := 0; end;\n property P[a: Integer]: Integer read Get;", []string{
				`Syntax Error: Parameter 0 (a) - Value-parameter expected [line: 3, column: 42]`,
				`Syntax Error: Method "Get" has incompatible parameters [line: 3, column: 42]`,
			}},
			{"body", " function Get(x: Integer): Integer; begin Result := MissingBody; end;\n property P[a: Integer]: Integer read Get;", []string{
				`Syntax Error: Unknown name "MissingBody" [line: 2, column: 53]`,
			}},
		} {
			t.Run(prefix+"/"+tt.name, func(t *testing.T) {
				source := prefix + "\n" + tt.body + "\n property Q[a: Missing : Integer read (1);\nend;\nTail; {$ERROR 'unread'}"
				want := make([]string, len(tt.want))
				copy(want, tt.want)
				want = append(want, `Syntax Error: Unknown name "Missing" [line: 4, column: 23]`, `Syntax Error: "]" expected [line: 4, column: 24]`)
				result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
				if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
					t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
				}
			})
		}
	}
}

func TestEngineCompile_RecordIndexPriorMethodScope(t *testing.T) {
	for _, prefix := range []string{"type TR = record", "var r: record", "var r := record"} {
		source := prefix + `
 function Base(x: Integer): Integer; begin Result := x; end;
 function Get(x: Integer): Integer; begin Result := Base(x); end;
 property P[a: Integer]: Integer read Get;
 property Q[a: Missing : Integer read (1);
end;
Tail; {$ERROR 'unread'}`
		want := []string{
			`Syntax Error: Unknown name "Missing" [line: 5, column: 23]`,
			`Syntax Error: "]" expected [line: 5, column: 24]`,
		}
		result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
		if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
			t.Fatalf("%s got:\n%s\nwant:\n%s", prefix, got, strings.Join(want, "\n"))
		}
	}
}

func TestEngineCompile_RecordIndexPriorCompletedPropertyError(t *testing.T) {
	for _, tt := range []struct {
		name, body string
		want       []string
	}{
		{"type", " function Get(x: String): Integer; begin Result := 0; end;\n property P[var a: Integer]: Integer read Get;", []string{
			`Syntax Error: Parameter 0 - Type "Integer" expected (instead of "String") [line: 3, column: 46]`,
			`Syntax Error: Method "Get" has incompatible parameters [line: 3, column: 46]`,
		}},
		{"mode", " function Get(var x: Integer): Integer; begin Result := 0; end;\n property P[a: Integer]: Integer read Get;", []string{
			`Syntax Error: Parameter 0 (a) - Value-parameter expected [line: 3, column: 42]`,
			`Syntax Error: Method "Get" has incompatible parameters [line: 3, column: 42]`,
		}},
		{"body", " function Get(x: Integer): Integer; begin Result := MissingBody; end;\n property P[a: Integer]: Integer read Get;", []string{
			`Syntax Error: Unknown name "MissingBody" [line: 2, column: 53]`,
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "var r: record\n" + tt.body + `
 property Q[a: Integer]: MissingType read Get;
 property U[a: MissingIndex : Integer read (1);
end;
Tail; {$ERROR 'unread'}`
			want := make([]string, len(tt.want))
			copy(want, tt.want)
			want = append(want,
				`Syntax Error: unknown type 'MissingType' for property 'Q' in inline record [line: 4, column: 2]`,
				`Syntax Error: Unknown name "MissingIndex" [line: 5, column: 28]`,
				`Syntax Error: "]" expected [line: 5, column: 29]`)
			result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
				t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
			}
		})
	}
}

func TestEngineCompile_RecordIndexCompletePropertyError(t *testing.T) {
	source := `var r: record
 function Get(x: String): Integer; begin Result := 0; end;
 property P[var a: Integer]: Integer read Get;
 property Q[a: Integer]: MissingType read Get;
end;`
	result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
	want := `Syntax Error: unknown type 'record' [line: 1, column: 8]`
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestEngineCompile_RecordIndexUnknownGetterResult(t *testing.T) {
	for _, directive := range []string{"", " default;"} {
		source := `type TR = record
 function Get(var x: Integer): MissingResult; begin end;
 property P[a: Integer]: Integer read Get;` + directive + "\nend;"
		want := []string{
			`Syntax Error: unknown return type 'MissingResult' for method 'Get' [line: 2, column: 2]`,
			`Syntax Error: unknown return type [line: 2, column: 2]`,
			`Syntax Error: Field/method "Get" has an incompatible type [line: 3, column: 42]`,
		}
		result := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
		if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
			t.Fatalf("%q got:\n%s\nwant:\n%s", directive, got, strings.Join(want, "\n"))
		}
	}
}
