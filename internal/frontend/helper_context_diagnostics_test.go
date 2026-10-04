package frontend

import "testing"

// Explicit helper names supply Self as an ordinary written argument. Helper
// bodies bind it implicitly, using the same shifted positions as receiver calls.
func TestCompile_HelperContextDiagnostics(t *testing.T) {
	for _, target := range []struct{ name, declaration string }{
		{"primitive", "Integer"},
		{"record", "record value: Integer; end"},
		{"class", "class end"},
		{"interface", "interface end"},
	} {
		t.Run(target.name, func(t *testing.T) {
			for _, tt := range []struct {
				name, member, explicit, body string
				explicitWant, bodyWant       []string
			}{
				{"short", "procedure Take(v: Integer);", "H.Take(item);", "Take();",
					[]string{`Syntax Error: More arguments expected [line: 4, column: 3]`},
					[]string{`Syntax Error: More arguments expected [line: 4, column: 1]`}},
				{"excess", "procedure Take(v: Integer);", "H.Take(item, 1, 2);", "Take(1, 2);",
					[]string{`Syntax Error: Too many arguments [line: 4, column: 3]`},
					[]string{`Syntax Error: Too many arguments [line: 4, column: 1]`}},
				{"parameterless excess", "procedure Take;", "H.Take(item, 1);", "Take(1);",
					[]string{`Syntax Error: Too many arguments [line: 4, column: 3]`},
					[]string{`Syntax Error: Too many arguments [line: 4, column: 1]`}},
				{"type before short", "procedure Take(v: Integer; w: String);", "H.Take(item, true);", "Take(true);",
					[]string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 4, column: 14]`},
					[]string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 4, column: 1]`}},
				{"type before excess", "procedure Take(v: Integer);", "H.Take(item, true, 1);", "Take(true, 1);",
					[]string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 4, column: 14]`},
					[]string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 4, column: 12]`}},
				{"multiline", "procedure Take(v: Integer);", "H.Take(item, true,\n  1);", "Take(true,\n  1);",
					[]string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 4, column: 14]`},
					[]string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 5, column: 3]`}},
				{"child before short", "procedure Take(v: Integer; w: String);", "H.Take(item,\n  Missing);", "Take(\n  Missing);",
					[]string{`Syntax Error: Unknown name "Missing" [line: 5, column: 3]`, `Syntax Error: More arguments expected [line: 4, column: 3]`},
					[]string{`Syntax Error: Unknown name "Missing" [line: 5, column: 3]`, `Syntax Error: More arguments expected [line: 4, column: 1]`}},
				{"excess children", "procedure Take;", "H.Take(item,\n  Missing, Other);", "Take(\n  Missing, Other);",
					[]string{`Syntax Error: Unknown name "Missing" [line: 5, column: 3]`, `Syntax Error: Unknown name "Other" [line: 5, column: 12]`, `Syntax Error: Too many arguments [line: 4, column: 3]`},
					[]string{`Syntax Error: Unknown name "Missing" [line: 5, column: 3]`, `Syntax Error: Unknown name "Other" [line: 5, column: 12]`, `Syntax Error: Too many arguments [line: 4, column: 1]`}},
				{"case insensitive", "procedure Take(v: Integer);", "h.tAKE(item);", "tAKE();",
					[]string{`Syntax Error: More arguments expected [line: 4, column: 3]`},
					[]string{`Syntax Error: More arguments expected [line: 4, column: 1]`}},
				{"valid", "procedure Take(v: Integer);", "H.Take(item, 1);", "Take(1);", nil, nil},
				{"contextual array", "procedure Take(v: array of Integer);", "H.Take(item, []);", "Take([]);", nil, nil},
				{"contextual nil", "procedure Take(v: TObject);", "H.Take(item, nil);", "Take(nil);", nil, nil},
			} {
				t.Run(tt.name, func(t *testing.T) {
					t.Run("explicit", func(t *testing.T) {
						source := "type T = " + target.declaration + ";\ntype H = helper for T " + tt.member + " begin end; end;\nvar item: T;\n" + tt.explicit
						assertHelperCallDiagnostics(t, source, tt.explicitWant)
					})
					t.Run("body", func(t *testing.T) {
						source := "type T = " + target.declaration + ";\ntype H = helper for T " + tt.member + " begin end;\nprocedure Run; begin\n" + tt.body + "\nend; end;"
						assertHelperCallDiagnostics(t, source, tt.bodyWant)
					})
				})
			}
		})
	}
}

func TestCompile_HelperContextReceiverRoles(t *testing.T) {
	for _, tt := range []struct {
		name, target, modifier, directives, explicit, body string
		explicitWant, bodyWant                             []string
	}{
		{"primitive class", "Integer", "class ", "", "H.Take(true, 1);", "Take(true, 1);",
			[]string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 3, column: 8]`},
			[]string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 4, column: 6]`}},
		{"interface class", "interface end", "class ", "", "H.Take(true, 1);", "Take(true, 1);",
			[]string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 3, column: 8]`},
			[]string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 4, column: 6]`}},
		{"record class", "record value: Integer; end", "class ", "", "H.Take(T, true, 1);", "Take(true, 1);",
			[]string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 3, column: 11]`},
			[]string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 4, column: 12]`}},
		{"object class", "class end", "class ", "", "H.Take(T, true, 1);", "Take(true, 1);",
			[]string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 3, column: 11]`},
			[]string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 4, column: 12]`}},
		{"record static", "record value: Integer; end", "class ", "static; ", "H.Take(true, 1);", "Take(true, 1);",
			[]string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 3, column: 8]`},
			[]string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 4, column: 6]`}},
		{"object static", "class end", "class ", "static; ", "H.Take(true, 1);", "Take(true, 1);",
			[]string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 3, column: 8]`},
			[]string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 4, column: 6]`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			prefix := "type T = " + tt.target + ";\ntype H = helper for T " + tt.modifier + "procedure Take(v: Integer); " + tt.directives + "begin end;"
			assertHelperCallDiagnostics(t, prefix+" end;\n"+tt.explicit, tt.explicitWant)
			assertHelperCallDiagnostics(t, prefix+"\n"+tt.modifier+"procedure Run; "+tt.directives+"begin\n"+tt.body+"\nend; end;", tt.bodyWant)
		})
	}
}

func TestCompile_HelperContextBindingAndRecovery(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"wrong explicit receiver before short", "type H = helper for Integer procedure Take(v: Integer); begin end; end;\nH.Take(true);", []string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 2, column: 8]`}},
		{"record class alias and metatype copy", "type T = record x: Integer; end;\ntype A = T;\ntype H = helper for T class procedure Take(v: Integer); begin end; end;\nvar meta := A;\nH.Take(meta, 1);", nil},
		{"Variant record class receiver", "type T = record x: Integer; end;\ntype H = helper for T class procedure Take(v: Integer); begin end; end;\nvar item: Variant := 1;\nH.Take(item, 1);", []string{`Syntax Error: Argument 0 expects type "meta of T" instead of "Variant" [line: 4, column: 8]`}},
		{"wrong record class receiver", "type T = record x: Integer; end;\ntype H = helper for T class procedure Take(v: Integer); begin end; end;\nvar item: T;\nH.Take(item, 1);", []string{`Syntax Error: Argument 0 expects type "meta of T" instead of "T" [line: 4, column: 8]`}},
		{"out of line", "type H = helper for Integer procedure Take(v: Integer); begin end; procedure Run; end;\nprocedure H.Run; begin\nTake(true, 1);\nend;", []string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 3, column: 12]`}},
		{"callable parameter shadows helper", "type H = helper for Integer procedure Take(v: Integer); begin end;\nprocedure Run(Take: procedure); begin\nTake(1);\nend; end;", []string{`Syntax Error: Too many arguments [line: 3, column: 5]`}},
		{"parameter shadows helper", "type H = helper for Integer procedure Take(v: Integer); begin end;\nprocedure Run(Take: Integer); begin\nTake(1);\nend; end;", []string{`Syntax Error: 'Take' is not a function [line: 3, column: 5]`}},
		{"explicit result recovery", "type H = helper for Integer function Take(v: Integer): Integer; begin Result := v; end; end;\nvar s: String;\ns := H.Take(1);", []string{`Syntax Error: More arguments expected [line: 3, column: 8]`, `Syntax Error: Incompatible types: Cannot assign "Integer" to "String" [line: 3, column: 6]`}},
		{"body result recovery", "type H = helper for Integer function Take(v: Integer): Integer; begin Result := v; end;\nprocedure Run; begin\nvar s: String;\ns := Take();\nend; end;", []string{`Syntax Error: More arguments expected [line: 4, column: 6]`, `Syntax Error: Incompatible types: Cannot assign "Integer" to "String" [line: 4, column: 6]`}},
	} {
		t.Run(tt.name, func(t *testing.T) { assertHelperCallDiagnostics(t, tt.source, tt.want) })
	}
}
