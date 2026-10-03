package frontend

import "testing"

func TestCompile_ParameterDefaultDiagnostics(t *testing.T) {
	assertDiagnostics(t, fixtureSource(t, "params1.pas"), "params1.pas", fixtureExpectation(t, "params1.txt"))
}

func TestCompile_SwapDiagnostics(t *testing.T) {
	assertDiagnostics(t, fixtureSource(t, "swap1.pas"), "swap1.pas", fixtureExpectation(t, "swap1.txt"))
}

func TestCompile_SwapWritableArguments(t *testing.T) {
	for _, source := range []string{
		"var a := 1; var b := 2; Swap(a, b);",
		"var a: array of Integer := [1, 2]; Swap(a[0], a[1]);",
		"type TRec = record x: Integer; end; var a, b: TRec; Swap(a.x, b.x);",
		"type TObj = class x: Integer; end; var a := TObj.Create; Swap(a.x, a.x);",
	} {
		t.Run(source, func(t *testing.T) {
			assertDiagnostics(t, source, "<test>", nil)
		})
	}
}

func TestCompile_ValidParameterDefaults(t *testing.T) {
	for _, source := range []string{
		"procedure Take(v: Float = 2); begin end; Take;",
		"procedure Take(v: TObject = nil); begin end; Take;",
		"procedure Take(v: Integer = (1 + 2)); begin end; Take;",
		"procedure Take(v: Integer = Abs(-2)); begin end; Take;",
		"procedure Take(const v: array of const); begin end; Take([1, 'text']);",
	} {
		t.Run(source, func(t *testing.T) { assertDiagnostics(t, source, "<test>", nil) })
	}
}

func TestCompile_ModifiedDefaultsRemainRequired(t *testing.T) {
	for _, modifier := range []string{"lazy", "var", "const"} {
		source := "procedure Take(" + modifier + " v: Integer = 1); external;\nTake();"
		got := Compile(source, "<test>", 0).DiagnosticStrings()
		if len(got) != 2 || got[1] != `Syntax Error: More arguments expected [line: 2, column: 1]` {
			t.Fatalf("%s diagnostics: %q", modifier, got)
		}
	}
}

func TestCompile_SwapReadOnlyAndPoisonedArguments(t *testing.T) {
	for _, tt := range []struct {
		source string
		want   []string
	}{
		{"const c = 1; var a := 2; Swap(c, a);", []string{`Syntax Error: Variable expected [line: 1, column: 31]`}},
		{"var a := 1; Swap(missing, a);", []string{`Syntax Error: Unknown name "missing" [line: 1, column: 18]`}},
		{"var a := 1; Swap(a, missing);", []string{`Syntax Error: Unknown name "missing" [line: 1, column: 21]`}},
	} {
		t.Run(tt.source, func(t *testing.T) { assertDiagnostics(t, tt.source, "<test>", tt.want) })
	}
}

func TestCompile_ImmutableAssignmentDiagnostics(t *testing.T) {
	for _, name := range []string{"const_param1", "const_param4", "class_const3", "const_2"} {
		t.Run(name, func(t *testing.T) {
			assertDiagnostics(t, fixtureSource(t, name+".pas"), name+".pas", fixtureExpectation(t, name+".txt"))
		})
	}
}

func TestCompile_SwapImmutableData(t *testing.T) {
	for _, expression := range []string{"o.R", "o.C", "a[0]"} {
		source := "type TObj = class const C = 1; function GetR: Integer; begin Result := 1; end; property R: Integer read GetR; end;\n" +
			"var o := TObj.Create; var b := 2; const a = [1, 2];\nSwap(" + expression + ", b);"
		assertDiagnostics(t, source, "<test>", []string{`Syntax Error: Variable expected [line: 3, column: 6]`})
	}
}

func TestCompile_SwapVariantPair(t *testing.T) {
	assertDiagnostics(t, "var a := 1; var b: Variant := 2;\nSwap(a, b);", "<test>", []string{
		`Syntax Error: Incompatible types: "Integer" and "Variant" [line: 2, column: 1]`,
	})
}

func TestCompile_AggregateParameterDefaults(t *testing.T) {
	assertDiagnostics(t, "type TInt = array[0..0] of Integer;\nprocedure Take(v: TInt = [Abs(-2)]); begin end; Take;", "<test>", nil)
}

func TestCompile_DefaultDiagnosticsPrecedeLaterCallOnSameLine(t *testing.T) {
	assertDiagnostics(t, "procedure Take(v: Float = Random); external; Take;", "<test>", []string{
		`Syntax Error: Constant expression expected [line: 1, column: 33]`,
		`Syntax Error: More arguments expected [line: 1, column: 46]`,
	})
}

func TestCompile_StaticArraySwapTemporary(t *testing.T) {
	assertDiagnostics(t, "type TInt = array[0..1] of Integer; function Values: TInt; begin Result := [1,2]; end; var a := 1;\nSwap(Values()[0], a);", "<test>", []string{
		`Syntax Error: Variable expected [line: 2, column: 6]`,
	})
}

func TestCompile_ConstantTypecastDefault(t *testing.T) {
	assertDiagnostics(t, "procedure Take(v: Integer = Integer(Abs(-2))); begin end; Take;", "<test>", nil)
}
