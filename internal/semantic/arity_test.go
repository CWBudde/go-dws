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
		{
			name: "overload set with no parameterless member",
			input: "procedure Test(i : Integer); overload;\nbegin\nend;\n" +
				"procedure Test(s : String); overload;\nbegin\nend;\n\nTest;",
			want: "More arguments expected at 8:1",
		},
		{
			name: "method reached through Self",
			input: "type TTest = class\n   procedure Doit(i : Integer);\n   procedure Run;\nend;\n" +
				"procedure TTest.Doit(i : Integer);\nbegin\nend;\n" +
				"procedure TTest.Run;\nbegin\n   Self.Doit;\nend;",
			want: "More arguments expected at 10:9",
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
			name: "overload set with a parameterless member",
			input: "procedure Test; overload;\nbegin\nend;\n" +
				"procedure Test(i : Integer); overload;\nbegin\nend;\n\nTest;",
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

// The count a constructor call is measured against comes from the whole overload
// set, not from whichever signature happens to be declared first: a class with
// both `Create` and `Create(Integer)` accepts 0..1 arguments, so two is over the
// top rather than short of the parameterless one.
func TestConstructorArityBoundsSpanTheOverloadSet(t *testing.T) {
	input := "type T = class\n   constructor Create;\n   constructor Create(i : Integer); overload;\nend;\n" +
		"constructor T.Create;\nbegin\nend;\n" +
		"constructor T.Create(i : Integer);\nbegin\nend;\n\n" +
		"var x := new T(1, 2);"
	got := analyzeWithHints(t, input, HintsLevelPedantic)
	if !hasDiagnosticContaining(got, "Too many arguments") {
		t.Errorf("expected the count to be measured against the whole set, got %v", got)
	}
}

// A routine reference the context rejects is re-read as a call, so a routine
// that needs arguments is short of them before the type error the context goes
// on to report. Recorded in FailureScripts/func_ptr1.txt and func_ptr4.txt.
func TestPointerContextImplicitCallArity(t *testing.T) {
	const procedures = "type TMyProc = procedure;\n\n" +
		"procedure Proc1;\nbegin\nend;\n\n" +
		"procedure Proc2(i : Integer);\nbegin\nend;\n\n" +
		"function Proc4 : String;\nbegin\n   Result := '';\nend;\n\n" +
		"var p : TMyProc;\n"

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "signature does not fit the pointer",
			input: procedures + "p:=Proc2;",
			want:  "More arguments expected at 17:4",
		},
		{
			name:  "address-of a builtin that needs arguments",
			input: "type TMyProc = procedure;\nvar p : TMyProc;\np := @IntToHex;",
			want:  "More arguments expected at 3:7",
		},
		{
			name:  "inside an array literal typed by the parameter",
			input: "procedure Test(const AParams : array of Integer);\nbegin\nend;\n\nTest([Test]);",
			want:  "More arguments expected at 5:7",
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

	// The reference fits, or the call the context falls back to is well-formed.
	// Neither is short of arguments, so neither says so.
	notReported := []struct {
		name  string
		input string
	}{
		{name: "signature fits the pointer", input: procedures + "p:=Proc1;"},
		{
			name:  "wrong result type but no parameters",
			input: procedures + "p:=Proc4;",
		},
		{
			name:  "array helper callbacks keep the reference reading",
			input: "var a : array of Integer;\na.ForEach(IntToStr);",
		},
	}

	for _, tt := range notReported {
		t.Run(tt.name, func(t *testing.T) {
			for _, diagnostic := range analyzeWithHints(t, tt.input, HintsLevelPedantic) {
				if containsDiagnostic([]string{diagnostic}, "More arguments expected") {
					t.Errorf("unexpected arity diagnostic %q", diagnostic)
				}
			}
		})
	}
}

// A function-pointer operand of a comparison is implicitly called too. It has no
// name token, so both diagnostics are anchored at the operator — the only
// position the comparison has. Recorded in FailureScripts/callback_err_vs_nil.txt.
func TestComparisonOperandImplicitCallArity(t *testing.T) {
	input := "type TSomeCallback = procedure(x, y : Integer);\n" +
		"var callback : TSomeCallback;\n" +
		"if callback <> nil then\n   PrintLn('x');"

	got := analyzeWithHints(t, input, HintsLevelPedantic)
	for _, want := range []string{
		"More arguments expected at 3:13",
		"Syntax Error: Invalid Operands at 3:13",
	} {
		if !containsDiagnostic(got, want) {
			t.Errorf("missing %q in %v", want, got)
		}
	}
}

// A call through a function pointer uses the same two sentences as any other
// call site. `No arguments expected` is not among them even when the pointer
// declares no parameters: func_ptr1 calls a `procedure` pointer with one
// argument and gets `Too many arguments`.
func TestFunctionPointerCallArityVocabulary(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "too many for a parameterless pointer",
			input: "type TMyProc = procedure;\nvar p : TMyProc;\n" +
				"p('hello');",
			want: "Too many arguments at 3:2",
		},
		{
			name: "too few",
			input: "type TBinaryOp = function(x, y : Integer) : Integer;\n" +
				"var op : TBinaryOp;\nvar r : Integer;\nr := op(5);",
			want: "More arguments expected at 4:8",
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

// DWScript names the type a context required and nothing else, and anchors the
// message at the first token of the unit that owns the value — the introducer
// where the unit has one. Recorded in loop_nonbool, repeat2,
// ifthenelse_expression1 and contracts_types.
func TestBooleanExpectedAnchor(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "while anchors at the keyword, not the condition",
			input: "while 'hello' do ;",
			want:  "Syntax Error: Boolean expected at 1:1",
		},
		{
			name:  "repeat anchors at until, not at repeat",
			input: "repeat until 'hello';",
			want:  "Syntax Error: Boolean expected at 1:8",
		},
		{
			name:  "the if-then-else expression anchors at if",
			input: "var t1 := if 'bug' then 1 else 2;",
			want:  "Syntax Error: Boolean expected at 1:11",
		},
		{
			name:  "a contract clause has no introducer, so it anchors at the condition",
			input: "procedure Test(i : Integer);\nrequire\n   IntToStr(i) : i;\nbegin\nend;",
			want:  "Syntax Error: Boolean expected at 3:4",
		},
		{
			name:  "a contract message re-uses the condition's anchor",
			input: "procedure Test(i : Integer);\nrequire\n   IntToStr(i) : i;\nbegin\nend;",
			want:  "Syntax Error: String expected at 3:4",
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

// One empty-block hint per `if`. Upstream reports only `Empty THEN block` for
// `if True then else ;` (if_empty_terms) and reaches `Empty ELSE block` only
// where the THEN branch has a body of its own (empty_if_block). A `while` loop
// with an empty body draws no hint at all — `Empty FOR loop` belongs to the FOR
// loops.
func TestEmptyBlockHints(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		notWant []string
	}{
		{
			name:    "an empty else beside an empty then is not reported",
			input:   "if True then else ;",
			want:    []string{"Hint: Empty THEN block [line: 1, column: 14]"},
			notWant: []string{"Empty ELSE block"},
		},
		{
			name:    "an empty else beside a real then is reported",
			input:   "if True then PrintLn('x') else ;",
			want:    []string{"Hint: Empty ELSE block [line: 1, column: 32]"},
			notWant: []string{"Empty THEN block"},
		},
		{
			name:    "an empty while body draws no hint",
			input:   "while True do ;",
			notWant: []string{"Empty FOR loop", "Empty WHILE loop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeWithHints(t, tt.input, HintsLevelPedantic)
			for _, want := range tt.want {
				if !containsDiagnostic(got, want) {
					t.Errorf("missing %q in %v", want, got)
				}
			}
			for _, unwanted := range tt.notWant {
				if containsDiagnostic(got, unwanted) {
					t.Errorf("unexpected %q in %v", unwanted, got)
				}
			}
		})
	}
}
