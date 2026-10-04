package frontend

import (
	"slices"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// Expectations derive from DWScript 1dbf8a9's ReadTypeHelper,
// WrapUpFunctionRead and ResolveOverload, independently of selected-signature
// argument checking. A found helper with no matching candidate owns the error.
func TestCompile_HelperOverloadCandidates(t *testing.T) {
	for _, target := range []struct{ name, declaration string }{
		{"primitive", "Integer"},
		{"record", "record x: Integer; end"},
		{"class", "class end"},
		{"interface", "interface end"},
	} {
		t.Run(target.name, func(t *testing.T) {
			for _, tt := range []struct {
				name, members, call string
				want                []string
			}{
				{"marked short", "procedure Take(v: Integer); overload; begin end;", "item.Take();", []string{`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 4, column: 6]`}},
				{"marked excess", "procedure Take(v: Integer); overload; begin end;", "item.Take(1, 2);", []string{`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 4, column: 6]`}},
				{"marked wrong type", "procedure Take(v: Integer); overload; begin end;", "item.Take(true);", []string{`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 4, column: 6]`}},
				{"marked valid", "procedure Take(v: Integer); overload; begin end;", "item.Take(1);", nil},
				{"multiple no match", "procedure Take(v: Integer); overload; begin end; procedure Take(v: String); overload; begin end;", "item.Take(true);", []string{`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 4, column: 6]`}},
				{"multiple short", "procedure Take(v: Integer); overload; begin end; procedure Take(v: String); overload; begin end;", "item.Take();", []string{`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 4, column: 6]`}},
				{"multiple valid", "procedure Take(v: Integer); overload; begin end; procedure Take(v: String); overload; begin end;", "item.Take('x');", nil},
				{"contextual empty array", "procedure Take(v: array of Integer); overload; begin end;", "item.Take([]);", nil},
				{"contextual nil", "procedure Take(v: TObject); overload; begin end;", "item.Take(nil);", nil},
				{"multiple empty array", "procedure Take(v: array of Integer); overload; begin end; procedure Take(v: String); overload; begin end;", "item.Take([]);", nil},
				{"multiple nil", "procedure Take(v: TObject); overload; begin end; procedure Take(v: Integer); overload; begin end;", "item.Take(nil);", nil},
			} {
				t.Run(tt.name, func(t *testing.T) {
					source := "type T = " + target.declaration + ";\ntype H = helper for T " + tt.members + " end;\nvar item: T;\n" + tt.call
					assertHelperCallDiagnostics(t, source, tt.want)
				})
			}
		})
	}
}

func TestCompile_HelperOverloadExplicitAndBody(t *testing.T) {
	for _, tt := range []struct {
		name, members, use string
		want               []string
	}{
		{"explicit marked short", "procedure Take(v: Integer); overload; begin end;", "H.Take(item);", []string{`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 4, column: 3]`}},
		{"explicit marked type", "procedure Take(v: Integer); overload; begin end;", "H.Take(item, true);", []string{`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 4, column: 3]`}},
		{"explicit marked receiver type", "procedure Take(v: Integer); overload; begin end;", "H.Take(true, 1);", []string{`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 4, column: 3]`}},
		{"explicit marked valid", "procedure Take(v: Integer); overload; begin end;", "H.Take(item, 1);", nil},
		{"body marked short", "procedure Take(v: Integer); overload; begin end;\nprocedure Run; begin Take(); end;", "", []string{`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 3, column: 22]`}},
		{"body marked type", "procedure Take(v: Integer); overload; begin end;\nprocedure Run; begin Take(true); end;", "", []string{`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 3, column: 22]`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "type T = Integer;\ntype H = helper for T " + tt.members + " end;\nvar item: T;\n" + tt.use
			assertHelperCallDiagnostics(t, source, tt.want)
		})
	}
}

func TestCompile_HelperOverloadRecoverableChildren(t *testing.T) {
	prefix := "function Other(v: Integer): Integer; begin Result := v; end;\n" +
		"type H = helper for Integer\n" +
		"function Inner(v: Integer): Integer; overload; begin Result := v; end;\n" +
		"procedure Take(v: Integer; w: String); overload; begin end;\n" +
		"procedure Take(v: String; w: Integer); overload; begin end;\n"
	for _, tt := range []struct {
		name, use string
		want      []string
	}{
		{"receiver", "end;\nvar item: Integer;\nitem.Take(H.Inner(item, true), Other(true));", []string{
			`Syntax Error: There is no overloaded version of "Inner" that can be called with these arguments [line: 8, column: 13]`,
			`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 8, column: 38]`,
			`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 8, column: 6]`,
		}},
		{"explicit", "end;\nvar item: Integer;\nH.Take(item, H.Inner(item, true), Other(true));", []string{
			`Syntax Error: There is no overloaded version of "Inner" that can be called with these arguments [line: 8, column: 16]`,
			`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 8, column: 41]`,
			`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 8, column: 3]`,
		}},
		{"body", "procedure Run; begin\nTake(Inner(true), Other(true));\nend; end;", []string{
			`Syntax Error: There is no overloaded version of "Inner" that can be called with these arguments [line: 7, column: 6]`,
			`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 7, column: 25]`,
			`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 7, column: 1]`,
		}},
		{"selected child once", "end;\nvar item: Integer;\nitem.Take(Other(true), 'x');", []string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 8, column: 17]`}},
		{"explicit selected child once", "end;\nvar item: Integer;\nH.Take(item, Other(true), 'x');", []string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 8, column: 20]`}},
	} {
		t.Run(tt.name, func(t *testing.T) { assertHelperCallDiagnostics(t, prefix+tt.use, tt.want) })
	}
}

// Receiver eligibility is diagnosed before ranking, so a failed match must not
// suppress the independent error about an instance method read through a type.
func TestCompile_HelperOverloadTypeReceiverRecovery(t *testing.T) {
	prefix := "type T = class end;\ntype H = helper for T procedure Take(v: Integer); overload; begin end; end;\n"
	assertHelperCallDiagnostics(t, prefix+"T.Take(true);", []string{
		`Syntax Error: Class method or constructor expected [line: 3, column: 3]`,
		`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 3, column: 3]`,
	})
	assertHelperCallDiagnostics(t, prefix+"T.Take(1);", []string{
		`Syntax Error: Class method or constructor expected [line: 3, column: 3]`,
	})
}

func TestCompile_HelperOverloadPrimitiveTypeReceiverRecovery(t *testing.T) {
	prefix := "type H = helper for Integer procedure Take(v: Integer); overload; begin end; end;\n"
	assertHelperCallDiagnostics(t, prefix+"Integer.Take(true);", []string{
		`Syntax Error: Class method or constructor expected [line: 2, column: 9]`,
		`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 2, column: 9]`,
	})
	assertHelperCallDiagnostics(t, prefix+"Integer.Take(1);", []string{
		`Syntax Error: Class method or constructor expected [line: 2, column: 9]`,
	})
}

func TestCompile_HelperOverloadFailedCaseHint(t *testing.T) {
	for _, target := range []string{"Integer", "record x: Integer; end", "class end", "interface end"} {
		source := "type T = " + target + ";\ntype H = helper for T procedure Take(v: Integer); overload; begin end; end;\nvar item: T;\nitem.tAKE(true);"
		result := CompileWithOptions(source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: true})
		want := []string{
			`Hint: "tAKE" does not match case of declaration ("Take") [line: 4, column: 6]`,
			`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 4, column: 6]`,
		}
		got := result.DiagnosticStrings()
		if !slices.Equal(got, want) {
			t.Fatalf("got %q; want %q", got, want)
		}
	}
}
