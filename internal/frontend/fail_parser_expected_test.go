package frontend

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestCompile_DWScriptExpectedSentences pins the complete diagnostic output for inputs
// where a delimiter, name, type or keyword is missing (PLAN.md §4, F3 + F8). DWScript
// words these as `"X" expected`, `Name expected`, `Type expected` or `DO expected`,
// anchored at the token found instead of X — or, once the input has run out, at the
// last token of the input. Each case mirrors a fixture; the expectation is the
// fixture's `.txt` verbatim.
func TestCompile_DWScriptExpectedSentences(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		// ---- ")" -------------------------------------------------------------
		{
			// FailureScripts/missing_parenthesis2: a compiler stop, anchored at the ";".
			name:   "parenthesised expression missing its closing parenthesis",
			source: "var a := (1;",
			want:   []string{`Syntax Error: ")" expected [line: 1, column: 12]`},
		},
		{
			// FailureScripts/dyn_array_setlength1: the stop hides nothing before it.
			name:   "method call missing its closing parenthesis",
			source: "var a : array of Integer;\na.SetLength('hello');\na.SetLength(1;",
			want: []string{
				`Syntax Error: Integer expression expected [line: 2, column: 13]`,
				`Syntax Error: ")" expected [line: 3, column: 14]`,
			},
		},
		{
			// FailureScripts/constructor_invalid_param: anchored at the second "1"; the
			// end-of-program method check never runs after a stop.
			name:   "constructor call with a missing comma",
			source: "type\n   TMyClass = class\n      constructor Create(i : Integer);\n   end;\n \nnew TMyClass(1 1);",
			want:   []string{`Syntax Error: ")" expected [line: 6, column: 16]`},
		},
		{
			// SetOfFail/bracket_right_missing
			name:   "set pseudo-method call missing its closing parenthesis",
			source: "type TMyEnum = (enumOne, enumTwo);\ntype TMySet = set of TMyEnum;\n\nvar e : TMySet;\n\ne.Include(enumOne;\n",
			want:   []string{`Syntax Error: ")" expected [line: 6, column: 18]`},
		},
		{
			// HelpersFail/helper_error3: the input runs out, so the anchor is the last token.
			name:   "call unfinished at end of input",
			source: "Type\n\n THelper = Helper For TObject\n Procedure Proc(AString : String);\n Begin\n End;\n End;\n\nTObject.Create.Proc(''",
			want:   []string{`Syntax Error: ")" expected [line: 9, column: 21]`},
		},
		{
			// FailureScripts/enums3
			name:   "enum declaration missing its closing parenthesis",
			source: "type TEnum = (en1;",
			want:   []string{`Syntax Error: ")" expected [line: 1, column: 18]`},
		},
		{
			// FailureScripts/nested_type1
			name:   "nested enum declaration missing its closing parenthesis",
			source: "type C1 = class\n  type E1 = enum(a,b;\nend;\n",
			want:   []string{`Syntax Error: ")" expected [line: 2, column: 21]`},
		},
		{
			// FailureScripts/new_class6: `new (` anchors at the opening parenthesis.
			name:   "new with a parenthesised operand missing its closing parenthesis",
			source: "var i : Integer;\nvar o1 := new (TObject;\n",
			want:   []string{`Syntax Error: ")" expected [line: 2, column: 15]`},
		},
		{
			// FailureScripts/class_error2: Name expected is an error, the ")" a stop.
			name:   "class ancestor list unfinished at end of input",
			source: "type TTest = class(",
			want: []string{
				`Syntax Error: Name expected [line: 1, column: 19]`,
				`Syntax Error: ")" expected [line: 1, column: 19]`,
			},
		},
		{
			// FailureScripts/class_error3
			name:   "class ancestor list ending in a comma",
			source: "type TTest = class(TObject,",
			want: []string{
				`Syntax Error: Name expected [line: 1, column: 27]`,
				`Syntax Error: ")" expected [line: 1, column: 27]`,
			},
		},
		{
			// InterfacesFail/partial_declaration3
			name:   "interface ancestor list missing its closing parenthesis",
			source: "type IIntf1 = Interface \n      method Hello;\n      method GetWorld : Integer;\n      method SetWorld(i : Integer);\n      \n      property World : Integer read GetWorld write SetWorld;\n     \n   end;\n   \ntype IIntf2 = Interface(IIntf1 end;",
			want:   []string{`Syntax Error: ")" expected [line: 10, column: 25]`},
		},
		{
			// OperatorOverloadFail/operator_overload2
			name:   "operator operand list unfinished at end of input",
			source: "operator + (",
			want: []string{
				`Syntax Error: Type expected [line: 1, column: 12]`,
				`Syntax Error: ")" expected [line: 1, column: 12]`,
			},
		},
		{
			// OperatorOverloadFail/operator_overload3
			name:   "operator operand list missing its closing parenthesis",
			source: "operator + (TObject ;",
			want:   []string{`Syntax Error: ")" expected [line: 1, column: 21]`},
		},
		// ---- "]" -------------------------------------------------------------
		{
			// FailureScripts/array_new
			name:   "new array dimension missing its closing bracket",
			source: "var a := new Integer[5;",
			want:   []string{`Syntax Error: "]" expected [line: 1, column: 23]`},
		},
		{
			// FailureScripts/in_operator6: the stop also hides the hint for the empty block.
			name:   "set literal missing its closing bracket",
			source: "if 1 in [1..2 then ;\n",
			want:   []string{`Syntax Error: "]" expected [line: 1, column: 15]`},
		},
		{
			// FailureScripts/array_index_bracket_missing1: the call's argument check
			// never runs after a stop inside its argument list.
			name:   "array literal argument missing its closing bracket",
			source: "procedure Test(Data: array of Float);\nbegin\n  \nend;\n\nTest([1.2, 2.2);",
			want:   []string{`Syntax Error: "]" expected [line: 6, column: 15]`},
		},
		{
			// FailureScripts/array_index_bracket_missing2: an index "]" is an error, not
			// a stop, so the declaration's ";" is still reported; both anchor at the
			// last token of the input.
			name:   "array index unfinished at end of input",
			source: "var a := [1,2,3];\nvar x := 1 / a[1",
			want: []string{
				`Syntax Error: "]" expected [line: 2, column: 16]`,
				`Syntax Error: ";" expected [line: 2, column: 16]`,
			},
		},
		// ---- "(" and "=" ---------------------------------------------------------
		{
			// FailureScripts/enum_scoped2, second line: anchored at the first element.
			name:   "enum keyword without its opening parenthesis",
			source: "type MyEnum = enum a, b);\n",
			want:   []string{`Syntax Error: "(" expected [line: 1, column: 20]`},
		},
		{
			// FailureScripts/const_array2: a stop; the declaration is not analysed.
			name:   "typed constant without a value",
			source: "const a : array [1..2] of Integer;\n",
			want:   []string{`Syntax Error: "=" expected [line: 1, column: 34]`},
		},
		// ---- ";" -------------------------------------------------------------
		{
			// FailureScripts/contracts_unfinished4
			name:   "precondition unfinished at end of input",
			source: "procedure Test(i : Integer);\nrequire\n   i>0",
			want:   []string{`Syntax Error: ";" expected [line: 3, column: 6]`},
		},
		{
			// FailureScripts/missing_semi1 (without its case hints): an error, anchored
			// at the token found (the next "var"), then at the last token once the
			// input has run out.
			name:   "variable declarations missing their semicolons",
			source: "var a:String=IntToStr(2)\nvar b:String=IntToStr(3)",
			want: []string{
				`Syntax Error: ";" expected [line: 2, column: 1]`,
				`Syntax Error: ";" expected [line: 2, column: 24]`,
			},
		},
		{
			// FailureScripts/end_implementation2: the unit header's ";" is an error;
			// parsing goes on to the final "end".
			name:   "unit header missing its semicolon",
			source: "unit test\n\ninterface\n\nimplementation\n\nend;",
			want: []string{
				`Syntax Error: ";" expected [line: 3, column: 1]`,
				`Syntax Error: Dot "." expected [line: 7, column: 4]`,
			},
		},
		{
			// FailureScripts/program: both anchor at the last token of the input.
			name:   "program keyword alone",
			source: "program",
			want: []string{
				`Syntax Error: Name expected [line: 1, column: 1]`,
				`Syntax Error: ";" expected [line: 1, column: 1]`,
			},
		},
		{
			// FailureScripts/class_error8
			name:   "type declaration unfinished after the equals sign",
			source: "type\n  TClassTest =",
			want: []string{
				`Syntax Error: Type expected [line: 2, column: 14]`,
				`Syntax Error: ";" expected [line: 2, column: 14]`,
			},
		},
		{
			// FailureScripts/empty_body, second class: the directive's ";" is an error,
			// then the class body runs out of input.
			name:   "method directive unfinished at end of input",
			source: "type \n   TTest2 = class\n      method Proc; empty",
			want: []string{
				`Syntax Error: ";" expected [line: 3, column: 20]`,
				`Syntax Error: Name expected [line: 3, column: 20]`,
			},
		},
		// ---- keywords ----------------------------------------------------------
		{
			// FailureScripts/case_error2
			name:   "case without of",
			source: "var i : Integer;\n\ncase i ;",
			want:   []string{`Syntax Error: OF expected [line: 3, column: 8]`},
		},
		{
			// SetOfFail/of_missing
			name:   "set declaration without of",
			source: "type TMyEnum = (enumOne, enumTwo);\ntype TMySet = set TMyEnum;\n",
			want:   []string{`Syntax Error: OF expected [line: 2, column: 19]`},
		},
		{
			// FailureScripts/for_in_str1
			name:   "for-in without do",
			source: "var c : Integer;\nvar s : String;\n\nfor c in s ;",
			want:   []string{`Syntax Error: DO expected [line: 4, column: 12]`},
		},
		{
			// SetOfFail/for_in_set_missing_do: an error, not a stop — both are reported.
			name:   "two for-in loops without do",
			source: "type TEnum = (one, two);\nvar s : set of TEnum;\n\nvar e : TEnum;\nfor e in s ;\n\nfor var i in s ;\n",
			want: []string{
				`Syntax Error: DO expected [line: 5, column: 12]`,
				`Syntax Error: DO expected [line: 7, column: 16]`,
			},
		},
		{
			// FailureScripts/for_error2: parsing carries on as if the keyword were there.
			name:   "for loop without to and without do",
			source: "var i : Integer;\n\nfor i:=1 10 step 1;",
			want: []string{
				`Syntax Error: TO or DOWNTO expected [line: 3, column: 10]`,
				`Syntax Error: DO expected [line: 3, column: 19]`,
			},
		},
		{
			// FailureScripts/for_error1: a stop.
			name:   "for loop without assignment",
			source: "var i : Integer;\n\nfor i to 10 do ;",
			want:   []string{`Syntax Error: ":=" expected [line: 3, column: 7]`},
		},
		{
			// FailureScripts/ifthenelse_expression1, second line: parsing carries on.
			name:   "if expression without then",
			source: "var t2 := if 2=2 1 else 2;",
			want:   []string{`Syntax Error: THEN expected [line: 1, column: 18]`},
		},
		{
			// HelpersFail/helper_error1: a stop, anchored at the token found.
			name:   "helper without for",
			source: "type\n   TDummy = helper end;\n",
			want:   []string{`Syntax Error: FOR expected [line: 2, column: 20]`},
		},
		{
			// InterfacesFail/partial_declaration: anchored at the last token.
			name:   "interface unfinished at end of input",
			source: "type IIntf = Interface",
			want:   []string{`Syntax Error: END expected [line: 1, column: 14]`},
		},
		{
			// GenericsFail/declaration_params_error1: errors, and the declarations go on.
			name:   "generic parameter lists missing their closing angle bracket",
			source: "type TTest1<T = array of T;\n\ntype TTest2<T : = array of T;\n",
			want: []string{
				`Syntax Error: ">" expected [line: 1, column: 15]`,
				`Syntax Error: Type expected [line: 3, column: 17]`,
				`Syntax Error: ">" expected [line: 3, column: 17]`,
			},
		},
		{
			// FailureScripts/external1
			name:   "external class with a non-name",
			source: "type\n   TClassA = class external 1\n   end;",
			want:   []string{`Syntax Error: Name expected [line: 2, column: 29]`},
		},
		{
			// FailureScripts/case_error4
			name:   "case branch without its colon",
			source: "var i : Integer;\n\ncase i of\n   1 ;",
			want:   []string{`Syntax Error: Colon ":" expected [line: 4, column: 6]`},
		},
		{
			// FailureScripts/class_error6: the field's colon, then the class body runs
			// out of input; both anchor at the last token.
			name:   "class field without its colon at end of input",
			source: "type TTest = class\n   Field",
			want: []string{
				`Syntax Error: Colon ":" expected [line: 2, column: 4]`,
				`Syntax Error: Name expected [line: 2, column: 4]`,
			},
		},
		{
			// FailureScripts/const_param3: a routine parameter needs its colon, a stop.
			name:   "parameter without its colon",
			source: "procedure Test(const Integer v);\nbegin\n   v:=1;\nend;",
			want:   []string{`Syntax Error: Colon ":" expected [line: 1, column: 30]`},
		},
		{
			// OperatorOverloadFail/operator_overload4
			name:   "operator without its result type",
			source: "operator + (TObject, TObject) ;",
			want:   []string{`Syntax Error: Colon ":" expected [line: 1, column: 31]`},
		},
		{
			// FailureScripts/const_4 and for_unfinished1: Name expected at the last token.
			name:   "const with a number for a name",
			source: "const 4",
			want:   []string{`Syntax Error: Name expected [line: 1, column: 7]`},
		},
		{
			name:   "for keyword alone",
			source: "for",
			want:   []string{`Syntax Error: Name expected [line: 1, column: 1]`},
		},
		{
			// FailureScripts/for_var_error: a stop.
			name:   "for var without a name",
			source: "for var :=1 to 2 do ;",
			want:   []string{`Syntax Error: Name expected [line: 1, column: 9]`},
		},
		{
			// FailureScripts/proc_missing_name: the routines are read on with empty
			// names; the dotted one is a stop.
			name:   "routines without names",
			source: "function (abc : Integer) : Integer;\nbegin\n   Result:=1;\nend;\n\ntype \n   TDummy = class\n      function : Integer;\n   end;\n   \nprocedure TDummy.;\nbegin\nend;",
			want: []string{
				`Syntax Error: Name expected [line: 1, column: 10]`,
				`Syntax Error: Name expected [line: 8, column: 16]`,
				`Syntax Error: Name expected [line: 11, column: 18]`,
			},
		},
		{
			// FailureScripts/method_implem3
			name:   "qualified routine name ending in its dot",
			source: "procedure TObject. ;\nbegin\n   \nend;\n",
			want:   []string{`Syntax Error: Name expected [line: 1, column: 20]`},
		},
		{
			// FailureScripts/resourcestring2, last declaration: anchored at the "=".
			name:   "resourcestring without a name",
			source: "resourcestring = 'bug';",
			want:   []string{`Syntax Error: Name expected [line: 1, column: 16]`},
		},
		{
			// FailureScripts/in_operator4: "not" after an operand wants "in", a stop.
			name:   "not without in",
			source: "if 1 not [+1, 2] then;",
			want:   []string{`Syntax Error: IN expected [line: 1, column: 10]`},
		},
		{
			// InterfacesFail/interface_guid: a GUID wants a string, then its bracket.
			name:   "interface GUIDs",
			source: "type \n\tIMyIntf = interface \n\t\t['whatever'] // GUID isn't checked\n\tend;\n\ntype \n\tIMyIntf2 = interface \n\t\t[123] // GUID isn't checked but must be string\n\tend;\n\t\ntype \n\tIMyIntf3 = interface \n\t\t['hhhh'\n\tend;\t",
			want: []string{
				`Syntax Error: String expected [line: 8, column: 4]`,
				`Syntax Error: "]" expected [line: 14, column: 2]`,
			},
		},
		{
			// FailureScripts/const_record4: the constant cut short by the stop is not analysed.
			name:   "record constant closed with the wrong bracket",
			source: "type \n   TRec = record\n      x, y : Integer;\n   end;\n\nconst c1 : TRec = (x:1);\nconst c2 : TRec = (x:1];",
			want:   []string{`Syntax Error: ")" expected [line: 7, column: 23]`},
		},
		// ---- "X" expected but "Y" found -----------------------------------------
		{
			// FailureScripts/else_unexpected1
			name:   "else inside a begin block",
			source: "if True then begin\n   else\nend;",
			want:   []string{`Syntax Error: "end" expected but "else" found [line: 2, column: 4]`},
		},
		{
			// FailureScripts/else_unexpected3
			name:   "else inside a repeat block",
			source: "repeat\n\telse\n",
			want:   []string{`Syntax Error: "until" expected but "else" found [line: 2, column: 2]`},
		},
		{
			// FailureScripts/block_unfinished3: a nested block names only "end".
			name:   "statement followed by begin without a semicolon",
			source: "procedure Test;\nbegin\n   var d : Integer;\n   while True do begin\n     var i : Integer;\n     Test\n     begin\n\n\n\n\n",
			want:   []string{`Syntax Error: "end" expected but "begin" found [line: 7, column: 6]`},
		},
		{
			// FailureScripts/block_unfinished4
			name:   "statement followed by a stray closing parenthesis",
			source: "var toto:=1;\nif toto=1 then begin\ntoto := toto+ 17);\nPrintln(toto);\nend;",
			want:   []string{`Syntax Error: "end" expected but ")" found [line: 3, column: 17]`},
		},
		{
			// FailureScripts/block_unfinished1: a routine body names "ensure" too, and an
			// identifier is not quoted.
			name:   "statement followed by an identifier in a routine body",
			source: "procedure Test;\nvar i : Integer;\nbegin\n   var i2 : Integer;\n   Test\n   Test;\nend;",
			want:   []string{`Syntax Error: "ensure" or "end" expected but identifier found [line: 6, column: 4]`},
		},
		{
			// FailureScripts/final_dot1
			name:   "final dot inside a routine body",
			source: "program G1;\n\nprocedure test;\nbegin\n  try\n    PrintLn('a1');\n    PrintLn('a2');\n  except\n    PrintLn('a3');\n//  end;\nend;\n\nbegin\n  test();\nend.",
			want:   []string{`Syntax Error: "ensure" or "end" expected but "." found [line: 15, column: 4]`},
		},
		{
			// FailureScripts/double_finalization
			name:   "second finalization section",
			source: "unit test;\n\ninterface\n\nimplementation\n\nfinalization\n\nfinalization",
			want:   []string{`Syntax Error: "end" expected but "finalization" found [line: 9, column: 1]`},
		},
		{
			// FailureScripts/double_initialization
			name:   "second initialization section",
			source: "unit test;\n\ninterface\n\nimplementation\n\ninitialization\n\ninitialization",
			want:   []string{`Syntax Error: "finalization" or "end" expected but "initialization" found [line: 9, column: 1]`},
		},
		// ---- negative cases --------------------------------------------------------
		{
			name:   "well-formed delimiters produce nothing",
			source: "var a := (1);\nvar b := [1, 2];\nvar i : Integer;\nfor i := 1 to 2 do PrintLn(i);\ncase i of 1 : ; end;\nPrintLn(a);\nPrintLn(b[0]);",
			want:   nil,
		},
		{
			name:   "statements separated by semicolons inside a block produce nothing",
			source: "procedure Test;\nbegin\n   Test;\n   Test\nend;",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			got := result.DiagnosticStrings()
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}
