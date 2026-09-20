package semantic

import (
	"slices"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// analyzeTolerant runs the analyzer the way the compile path does: parse errors
// are reported by the parser, and the analyzer still walks whatever AST the
// parser recovered. Only the analyzer's diagnostics are returned.
func analyzeTolerant(t *testing.T, input string) []string {
	t.Helper()

	p := parser.New(lexer.New(input))
	program := p.ParseProgram()

	analyzer := NewAnalyzer()
	analyzer.SetHintsLevel(HintsLevelPedantic)
	_ = analyzer.Analyze(program)
	return analyzer.Errors()
}

// Call-argument diagnostics use DWScript's sentences: `Argument N expects type
// "X" instead of "Y"` (0-based N, anchored on the argument), the short form
// without "instead of" when the argument is a procedure call with no value, and
// `Invalid argument type` for intrinsics that reject the argument's kind. The
// expected output is recorded in testdata/fixtures/FailureScripts/*.txt.
func TestCallArgumentDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			// method_param_error1: the unknown name is the only error; the
			// argument it poisons is not type-checked again.
			name: "method argument poisoned by an unknown name",
			input: "type\n   TMyClass = class\n      procedure PrintMe(a : Integer);\n   end;\n   \n" +
				"var o : TMyClass;\n\no.PrintMe(IntToStr(bug));\n",
			want: []string{`Unknown name "bug" at 8:20`},
		},
		{
			// method_param_error2
			name: "class method argument poisoned by an unknown name",
			input: "type\n   TMyClass = class\n      class procedure PrintMe(a : Integer);\n   end;\n   \n" +
				"TMyClass.PrintMe(IntToStr(bug));\n",
			want: []string{`Unknown name "bug" at 6:27`},
		},
		{
			name: "plain call argument poisoned by an unknown name",
			input: "procedure PrintMe(a : Integer);\nbegin\nend;\n" +
				"PrintMe(IntToStr(bug));\n",
			want: []string{`Unknown name "bug" at 4:18`},
		},
		{
			name: "method argument of the wrong type",
			input: "type\n   TMyClass = class\n      procedure PrintMe(a : Integer);\n   end;\n" +
				"var o : TMyClass;\no.PrintMe('x');\n",
			want: []string{`Syntax Error: Argument 0 expects type "Integer" instead of "String" [line: 6, column: 11]`},
		},
		{
			name: "class method argument of the wrong type",
			input: "type\n   TMyClass = class\n      class procedure PrintMe(a, b : Integer);\n   end;\n" +
				"TMyClass.PrintMe(1, 'x');\n",
			want: []string{`Syntax Error: Argument 1 expects type "Integer" instead of "String" [line: 5, column: 21]`},
		},
		{
			// HelpersFail/function_helper: the receiver is argument 0, and the
			// shifted index finds no written argument, so the call anchors it.
			name: "helper method argument of the wrong type",
			input: "type TStrHelper = helper for String\n   procedure Test2(s : String);\nend;\n" +
				"procedure TStrHelper.Test2(s : String);\nbegin\nend;\n('hello').Test2(456);",
			want: []string{`Syntax Error: Argument 1 expects type "String" instead of "Integer" [line: 7, column: 11]`},
		},
		{
			name: "record method argument of the wrong type",
			input: "type TRec = record\n   procedure Test(a : Integer; b : String);\nend;\n" +
				"procedure TRec.Test(a : Integer; b : String);\nbegin\nend;\n" +
				"var r : TRec;\nr.Test('x', 'y');",
			want: []string{`Syntax Error: Argument 1 expects type "Integer" instead of "String" [line: 8, column: 13]`},
		},
		{
			// dyn_array1 / dyn_array_setlength2: the argument list did not
			// parse, so its length says nothing about the call.
			name:  "array helper whose argument list failed to parse",
			input: "var a : array of Integer;\na.Length(;",
			want:  nil,
		},
		{
			// open_array2: a dynamic array is not an open array.
			name: "dynamic array to array of const",
			input: "Procedure Proc(Const AParams : Array Of Const);\nBegin\nEnd;\n\n" +
				"Var Params : Array Of Variant;\n\nProc(Params);",
			want: []string{`Syntax Error: Argument 0 expects type "array of const" instead of "array of Variant" [line: 7, column: 6]`},
		},
		{
			// use_proc_result2
			name:  "procedure result to a builtin",
			input: "Print(Print(''));\n",
			want:  []string{`Syntax Error: Argument 0 expects type "Variant" [line: 1, column: 7]`},
		},
		{
			name:  "procedure result to a user routine",
			input: "procedure P(a : Integer);\nbegin\nend;\nP(Print(''));\n",
			want:  []string{`Syntax Error: Argument 0 expects type "Integer" [line: 4, column: 3]`},
		},
		{
			// foreach_invalid_arg: a parameterless procedure is callable as
			// written, so only its missing value is reported.
			name: "parameterless procedure as a ForEach callback",
			input: "procedure b; \nbegin\nend;\n\n" +
				"var a : array of String = ['a', 'b', 'c'];\na.ForEach(b);\n",
			want: []string{`Incompatible parameter types - "procedure (String)" expected (instead of "void") at 6:11`},
		},
		{
			// assigned
			name:  "Assigned on an Integer",
			input: "var b := Assigned( 123 );",
			want:  []string{"Invalid argument type at 1:20"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeTolerant(t, tt.input)
			if !slices.Equal(got, tt.want) {
				t.Errorf("diagnostics:\n got  %q\n want %q", got, tt.want)
			}
		})
	}
}

// Assigned accepts every kind that can be unassigned.
func TestAssignedAcceptedArguments(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "object", input: "var o : TObject;\nvar b := Assigned(o);"},
		{name: "nil", input: "var b := Assigned(nil);"},
		{name: "class reference", input: "var c : TClass;\nvar b := Assigned(c);"},
		{name: "interface", input: "type IFoo = interface end;\nvar i : IFoo;\nvar b := Assigned(i);"},
		{name: "function pointer", input: "type TProc = procedure;\nvar p : TProc;\nvar b := Assigned(p);"},
		{name: "variant", input: "var v : Variant;\nvar b := Assigned(v);"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := analyzeTolerant(t, tt.input); len(got) != 0 {
				t.Errorf("unexpected diagnostics: %q", got)
			}
		})
	}
}
