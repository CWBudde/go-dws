package frontend

import "testing"

// Bare helper members are calls unless the expected callable type fits. Grouping
// reads the inner member without that context before considering an outer call.
func TestCompile_HelperBareDiagnostics(t *testing.T) {
	for _, target := range []struct{ name, declaration string }{
		{"primitive", "Integer"},
		{"record", "record x: Integer; end"},
		{"class", "class end"},
		{"interface", "interface end"},
	} {
		t.Run(target.name, func(t *testing.T) {
			for _, tt := range []struct {
				name, member, use string
				want              []string
			}{
				{"statement short", "function Take(v: Integer): Integer; begin Result := v; end;", "item.Take;", []string{`Syntax Error: More arguments expected [line: 4, column: 6]`}},
				{"grouped statement short", "function Take(v: Integer): Integer; begin Result := v; end;", "(item.Take);", []string{`Syntax Error: More arguments expected [line: 4, column: 7]`}},
				{"value short", "function Take(v: Integer): Integer; begin Result := v; end;", "PrintLn(item.Take);", []string{`Syntax Error: More arguments expected [line: 4, column: 14]`}},
				{"scalar initializer short", "function Take(v: Integer): Integer; begin Result := v; end;", "var i: Integer := item.Take;", []string{`Syntax Error: More arguments expected [line: 4, column: 24]`}},
				{"inferred short", "function Take(v: Integer): Integer; begin Result := v; end;", "var i := item.Take;", []string{`Syntax Error: More arguments expected [line: 4, column: 15]`}},
				{"compatible reference", "function Take(v: Integer): Integer; begin Result := v; end;", "var p: function(v: Integer): Integer := item.Take;", nil},
				{"parameterless reference", "function Take: Integer; begin Result := 7; end;", "var p: function: Integer := item.Take;", nil},
				{"procedure reference", "procedure Take(v: Integer); begin end;", "var p: procedure(v: Integer) := item.Take;", nil},
				{"grouped scalar call", "function Take(v: Integer): Integer; begin Result := v; end;", "(item.Take)(Missing); Other;", []string{`Syntax Error: More arguments expected [line: 4, column: 7]`, `Syntax Error: Not a method [line: 4, column: 12]`}},
				{"grouped parameterless scalar call", "function Take: Integer; begin Result := 7; end;", "(item.Take)(Missing); Other;", []string{`Syntax Error: Not a method [line: 4, column: 12]`}},
				{"case insensitive", "function Take(v: Integer): Integer; begin Result := v; end;", "item.tAKE;", []string{`Syntax Error: More arguments expected [line: 4, column: 6]`}},
				{"valid parameterless value", "function Take: Integer; begin Result := 7; end;", "var i: Integer := item.Take;", nil},
			} {
				t.Run(tt.name, func(t *testing.T) {
					source := "type T = " + target.declaration + ";\ntype H = helper for T " + tt.member + " end;\nvar item: T;\n" + tt.use
					assertHelperCallDiagnostics(t, source, tt.want)
				})
			}
		})
	}
}

func TestCompile_HelperBareBindingsAndRecovery(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"explicit reference", "type H = helper for Integer function Take(v: Integer): Integer; begin Result := v; end; end;\nvar p: function(s: Integer; v: Integer): Integer := H.Take;", nil},
		{"explicit scalar call short", "type H = helper for Integer function Take(v: Integer): Integer; begin Result := v; end; end;\nvar i: Integer := H.Take;", []string{`Syntax Error: More arguments expected [line: 2, column: 21]`}},
		{"body scalar call short", "type H = helper for Integer function Take(v: Integer): Integer; begin Result := v; end;\nprocedure Run; begin\nvar i: Integer := Take;\nend; end;", []string{`Syntax Error: More arguments expected [line: 3, column: 19]`}},
		{"body reference", "type H = helper for Integer function Take(v: Integer): Integer; begin Result := v; end;\nprocedure Run; begin\nvar p: function(v: Integer): Integer := Take;\nend; end;", nil},
		{"grouped body call", "type H = helper for Integer function Take(v: Integer): Integer; begin Result := v; end;\nprocedure Run; begin\n(Take)(Missing); Other;\nend; end;", []string{`Syntax Error: More arguments expected [line: 3, column: 2]`, `Syntax Error: Not a method [line: 3, column: 7]`}},
		{"inherited helper short", "type H = helper for Integer procedure Take(v: Integer); begin end; end;\ntype J = helper(H) for Integer end;\nvar item: Integer;\nitem.Take;", []string{`Syntax Error: More arguments expected [line: 4, column: 6]`}},
		{"function helper short", "function Take(s: String; v: Integer): Integer; helper; begin Result := v; end;\n('x').Take;", []string{`Syntax Error: More arguments expected [line: 2, column: 7]`}},
		{"native interface owns name", "type T = interface function Take(v: Integer): Integer; end;\ntype H = helper for T function Take: Integer; begin Result := 1; end; end;\nvar item: T;\nitem.Take;", nil},
		{"helper property is a value", "type H = helper for Integer function GetValue: Integer; begin Result := 1; end; property Take: Integer read GetValue; end;\nvar item: Integer; var i: Integer := item.Take;", nil},
		{"helper Result alias", "type H = helper for Integer function Take(v: Integer): Integer; begin Take := v; end; end;", nil},
		{"noncallable body shadow", "type H = helper for Integer function Take(v: Integer): Integer; begin Result := v; end;\nprocedure Run(Take: Integer); begin\nvar i: Integer := Take;\nend; end;", nil},
		{"bare function result recovery", "type H = helper for Integer function Take(v: Integer): String; begin Result := ''; end; end;\nvar item: Integer; var i: Integer;\ni := item.Take;", []string{`Syntax Error: More arguments expected [line: 3, column: 11]`, `Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 3, column: 6]`}},
	} {
		t.Run(tt.name, func(t *testing.T) { assertHelperCallDiagnostics(t, tt.source, tt.want) })
	}
}

func TestCompile_HelperGroupedReturnedCallable(t *testing.T) {
	prefix := "type P = function(v: Integer): Integer;\nfunction Inner(v: Integer): Integer; begin Result := v; end;\ntype H = helper for Integer function Factory: P; begin Result := @Inner; end; end;\nvar item: Integer;\n"
	for _, tt := range []struct {
		name, use string
		want      []string
	}{
		{"valid", "(item.Factory)(1);", nil},
		{"short", "(item.Factory)();", []string{`Syntax Error: More arguments expected [line: 5, column: 15]`}},
		{"type before excess", "(item.Factory)(true, 1);", []string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 5, column: 16]`}},
		{"excess children", "(item.Factory)(1, Missing);", []string{`Syntax Error: Unknown name "Missing" [line: 5, column: 19]`, `Syntax Error: Too many arguments [line: 5, column: 15]`}},
	} {
		t.Run(tt.name, func(t *testing.T) { assertHelperCallDiagnostics(t, prefix+tt.use, tt.want) })
	}
}

// Grouped helper names lose reference context; preserve the initializer checker
// diagnostics here (its vocabulary is a separate declaration/assignment audit).
func TestCompile_HelperGroupedExplicitReferenceContext(t *testing.T) {
	assertHelperCallDiagnostics(t, "type P = function: Integer;\ntype H = helper for Integer class function F: Integer; begin Result := 7; end; end;\nvar p: P := (H.F);", []string{`Syntax Error: Cannot assign Integer to function P: Integer variable 'p' [line: 3, column: 1]`})
	assertHelperCallDiagnostics(t, "type P = procedure;\ntype H = helper for Integer class procedure F; begin end; end;\nvar p: P := (H.F);", []string{`Syntax Error: Cannot assign void to procedure P variable 'p' [line: 3, column: 1]`})
}
