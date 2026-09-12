package semantic

import "testing"

// A bare routine name in statement position is a call, so a routine that needs
// arguments is short of them. The wording and the anchor — the name, not the
// semicolon — are recorded in testdata/fixtures/FailureScripts/missing_param*.txt.
func TestImplicitCallArity(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "user procedure",
			input: "procedure Test(i : Integer);\nbegin\nend;\n\nTest;",
			want:  "More arguments expected at 5:1",
		},
		{
			name:  "builtin function",
			input: "Sin;",
			want:  "More arguments expected at 1:1",
		},
		{
			name: "class method through the metaclass",
			input: "type TTest = class\n   class procedure Test(i : Integer);\nend;\n\n" +
				"class procedure TTest.Test(i : Integer);\nbegin\nend;\n\nTTest.Test;",
			want: "More arguments expected at 9:7",
		},
		{
			name:  "overloaded builtin names no count",
			input: "Max;",
			want:  `There is no overloaded version of "Max" that can be called with these arguments at 1:1`,
		},
		{
			name:  "array helper anchored past the name",
			input: "var a : array of Integer;\na.SetLength;",
			want:  "More arguments expected at 2:12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeWithHints(t, tt.input, HintsLevelPedantic)
			if !containsDiagnostic(got, tt.want) {
				t.Errorf("missing %q in %v", tt.want, got)
			}
		})
	}
}

// The implicit call is only reported when no call is possible at all. A
// parameterless routine, an overload set with a parameterless member, and a
// helper answerable from the array alone are all complete as written.
func TestImplicitCallArityNotReported(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "parameterless procedure",
			input: "procedure Test;\nbegin\nend;\n\nTest;",
		},
		{
			name: "parameterless overload inherited from a parent",
			input: "type TBase = class\n   procedure Test; begin end;\nend;\n" +
				"type TSub = class(TBase)\n   procedure Test(a : Integer); overload; begin end;\nend;\n" +
				"var s := TSub.Create;\ns.Test;",
		},
		{
			name:  "array helper that needs no argument",
			input: "var a : array of Integer;\na.Clear;",
		},
		{
			name:  "indexed property supplied with its index",
			input: propertyClassSource("   PrintLn(Val[1]);"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, diagnostic := range analyzeWithHints(t, tt.input, HintsLevelPedantic) {
				if hasDiagnosticContaining([]string{diagnostic}, "More arguments expected") {
					t.Errorf("unexpected arity diagnostic: %s", diagnostic)
				}
			}
		})
	}
}

// An indexed property named without its indices reads the accessor with no
// arguments; see testdata/fixtures/FailureScripts/property_error10.txt.
func TestIndexedPropertyWithoutIndices(t *testing.T) {
	got := analyzeWithHints(t, propertyClassSource("   PrintLn(Val);"), HintsLevelPedantic)
	if !containsDiagnostic(got, "More arguments expected at 10:12") {
		t.Errorf("missing indexed-property arity diagnostic in %v", got)
	}
}

// propertyClassSource wraps a method body in a class declaring an indexed
// property, so the body lands on line 10 in every case.
func propertyClassSource(body string) string {
	return "type\n" +
		"   TMyClass = class\n" +
		"      function GetVal(i : Integer) : Integer;\n" +
		"      property Val[i : Integer] : Integer read GetVal;\n" +
		"      procedure Test;\n" +
		"   end;\n\n" +
		"procedure TMyClass.Test;\n" +
		"begin\n" +
		body + "\n" +
		"end;\n"
}

// DWScript names neither the routine nor the counts when a call's argument list
// does not fit; the three sentences below are the whole vocabulary. Recorded in
// testdata/fixtures/FailureScripts/{func_params1,array_error6}.txt.
func TestCanonicalArgumentCountWording(t *testing.T) {
	const decl = "type TMyClass = class\n   function Test(a : Integer; var b : Integer) : Integer;\nend;\n\n" +
		"function TMyClass.Test(a : Integer; var b : Integer) : Integer;\nbegin\n   Result := a;\nend;\n\n" +
		"var c := TMyClass.Create;\nvar d : Integer;\n"

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "too few arguments to a method",
			input: decl + "PrintLn(c.Test(1));",
			want:  "More arguments expected at 12:11",
		},
		{
			name:  "too many arguments to a method",
			input: decl + "PrintLn(c.Test(1, d, 3));",
			want:  "Too many arguments at 12:11",
		},
		{
			name:  "arguments to a routine that declares none",
			input: "var a : array of Integer;\nPrintLn(a.Low(1));",
			want:  "No arguments expected at 2:16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeWithHints(t, tt.input, HintsLevelPedantic)
			if !containsDiagnostic(got, tt.want) {
				t.Errorf("missing %q in %v", tt.want, got)
			}
		})
	}
}

// Upstream type-checks the arguments it was handed before it counts them, so a
// short call whose arguments do not fit reports the type error alone
// (testdata/fixtures/FailureScripts/func_params1.txt).
func TestArgumentTypeErrorOutranksCount(t *testing.T) {
	input := "procedure Test(a : Integer ; b : String);\nbegin\nend;\n\nTest('');"
	got := analyzeWithHints(t, input, HintsLevelPedantic)
	if hasDiagnosticContaining(got, "More arguments expected") {
		t.Errorf("count diagnostic emitted alongside an argument type error: %v", got)
	}
	if !hasDiagnosticContaining(got, `Argument 0 expects type "Integer" instead of "String"`) {
		t.Errorf("missing argument type diagnostic in %v", got)
	}
}

// MaxInt answers from nothing, so naming it discards a constant; Sin cannot be
// called at all and draws the arity diagnostic instead
// (testdata/fixtures/FailureScripts/missing_param1.txt).
func TestBareStatelessBuiltinIsConstantInstruction(t *testing.T) {
	got := analyzeWithHints(t, "Sin;\nMaxInt;", HintsLevelPedantic)
	if !containsDiagnostic(got, "More arguments expected at 1:1") {
		t.Errorf("missing arity diagnostic in %v", got)
	}
	if !containsDiagnostic(got, "Hint: Constant Instruction - has no effect [line: 2, column: 1]") {
		t.Errorf("missing constant-instruction hint in %v", got)
	}
}
