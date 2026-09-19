package frontend

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestCompile_DWScriptRecoverySentences pins the complete diagnostic output for inputs
// where go-dws used to add recovery sentences of its own after DWScript's (PLAN.md §4,
// F9). Each case mirrors a fixture; the expectation is the fixture's `.txt` verbatim.
func TestCompile_DWScriptRecoverySentences(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			// "Expression expected" is a compiler stop upstream: nothing follows it.
			name:   "missing operand inside parentheses",
			source: "(5 and );",
			want:   []string{`Syntax Error: Expression expected [line: 1, column: 8]`},
		},
		{
			name:   "missing unary operand in a later statement",
			source: "var  i := 0;\n\ni := (i + +i);\ni := (i + +);",
			want:   []string{`Syntax Error: Expression expected [line: 4, column: 12]`},
		},
		{
			name:   "property getter expression without a branch",
			source: "Type\n\n TObj = Class\n\n  Property Name : String Read (If True Then);\n\n End;",
			want:   []string{`Syntax Error: Expression expected [line: 5, column: 44]`},
		},
		{
			name:   "enum element value missing",
			source: "Type TElement = (etOne=);",
			want:   []string{`Syntax Error: Expression expected [line: 1, column: 24]`},
		},
		{
			name:   "case range missing its upper bound",
			source: "var i : Integer;\n\ncase i of\n   1..;",
			want:   []string{`Syntax Error: Expression expected [line: 4, column: 7]`},
		},
		{
			name:   "if-then-else expression missing its else branch",
			source: "var t := if true then 'a' else;",
			want:   []string{`Syntax Error: Expression expected [line: 1, column: 31]`},
		},
		{
			name:   "var without type or initializer",
			source: "var s;",
			want:   []string{`Syntax Error: Colon ":" expected [line: 1, column: 6]`},
		},
		{
			name:   "multi-name var with an equals initializer",
			source: "var i, j = ;",
			want:   []string{`Syntax Error: Colon ":" expected [line: 1, column: 10]`},
		},
		{
			name:   "enum type name assigned as a value",
			source: "Type TEnum = (name);\nvar i : TEnum;\ni := TEnum;\n",
			want:   []string{`Syntax Error: "(" expected [line: 3, column: 11]`},
		},
		{
			name:   "enum type name as an inferred initializer",
			source: "Type TEnum = (name);\nvar i := TEnum;\n",
			want:   []string{`Syntax Error: "(" expected [line: 2, column: 15]`},
		},
		{
			// Parser errors keep their emission order even when the later one sits
			// further left on the line.
			name:   "set of nothing",
			source: "type TMyEnum = (enumOne, enumTwo);\ntype TMySet = set of ;\n",
			want: []string{
				`Syntax Error: Type expected [line: 2, column: 22]`,
				`Syntax Error: Enumeration expected [line: 2, column: 19]`,
			},
		},
		{
			name: "read-only property assigned at the top level",
			source: "type\n   TMyClass = class\n      Field : Integer;\n      property Prop : Integer read Field;\n   end;\n   \n" +
				"var o := new TMyClass;\no.Prop:=1;",
			want: []string{
				`Syntax Error: Cannot set a value for a read-only property [line: 8, column: 3]`,
				`Syntax Error: Unexpected "Integer Literal" [line: 8, column: 9]`,
			},
		},
		{
			name:   "constant as a property write specifier",
			source: "type\n   TMyClass = class\n      const c = 1;\n      property p : Integer write c;\n   end;\n\nTMyClass.p:=3;",
			want: []string{
				`Syntax Error: Constant "c" cannot be written to [line: 4, column: 34]`,
				`Syntax Error: Cannot set a value for a read-only property [line: 7, column: 10]`,
				`Syntax Error: Unexpected "Integer Literal" [line: 7, column: 13]`,
			},
		},
		{
			// Inside a block upstream's statement loop words the leftover value
			// differently; go-dws does not guess at that sentence.
			name: "read-only property assigned inside a block",
			source: "type\n   TMyClass = class\n      Field : Integer;\n      property Prop : Integer read Field;\n   end;\n" +
				"var o := new TMyClass;\nbegin\n   o.Prop:=1;\nend;",
			want: []string{
				`Syntax Error: Cannot set a value for a read-only property [line: 8, column: 6]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			got := result.DiagnosticStrings()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

// TestCompile_TypeNamesThatAreValues keeps the `"(" expected` rule to non-class type
// names: a class name is a metaclass value, and an enum type name stays usable where
// DWScript accepts a type (High, Low, for-in).
func TestCompile_TypeNamesThatAreValues(t *testing.T) {
	source := "type TEnum = (a, b);\ntype TFoo = class end;\n" +
		"var c : TClass;\nc := TFoo;\nvar h := High(TEnum);\nvar e : TEnum;\nfor e in TEnum do PrintLn(Ord(e));\nPrintLn(h);\n"
	result := Compile(source, "<test>", semantic.HintsLevelPedantic)
	if got := result.DiagnosticStrings(); len(got) != 0 {
		t.Fatalf("expected no diagnostics, got %q", got)
	}
}
