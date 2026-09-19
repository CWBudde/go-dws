package frontend

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestCompile_RoutineDeclarationDiagnostics pins DWScript's exact sentences,
// anchors and ordering for overload and forward-declaration errors (PLAN.md §4
// F7). Each case mirrors a fixture under testdata/fixtures.
func TestCompile_RoutineDeclarationDiagnostics(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			// OverloadsFail/default_params: capitalized sentences; the duplicate
			// is anchored after the directives (the `begin`), the ambiguity at
			// the declaration keyword.
			name: "overload sentences and anchors",
			source: `procedure Test; overload;
begin
end;

procedure Test(a : Integer = 0); overload;
begin
end;

procedure Test1(a : String; b : Integer); overload;
begin
end;

procedure Test1(a : String; b : Integer); overload;
begin
end;
`,
			want: []string{
				`Syntax Error: Overload of "Test" will be ambiguous with a previously declared version [line: 5, column: 1]`,
				`Syntax Error: There is already a method with name "Test1" [line: 14, column: 1]`,
			},
		},
		{
			// OverloadsFail/overload_missing: the directive sentence names the
			// kind of the set's first overload.
			name: "missing overload directive",
			source: `procedure Test(i : Integer); overload;
begin
end;

procedure Test(s : String; t : String); overload;
begin
end;

function Test(s : String) : String;
begin
end;
`,
			want: []string{
				`Syntax Error: Overloaded procedure "Test" must be marked with the "overload" directive [line: 9, column: 1]`,
			},
		},
		{
			// FailureScripts/forward_missing1.
			name: "unimplemented forwards",
			source: `function Test1 : String; forward;
procedure Test2; forward;

PrintLn('hello');
`,
			want: []string{
				`Syntax Error: The function "Test1" was forward declared but not implemented [line: 1, column: 10]`,
				`Syntax Error: The function "Test2" was forward declared but not implemented [line: 2, column: 11]`,
			},
		},
		{
			// OverloadsFail/forwards and overload_func_ptr_param: reported after
			// every other error, sorted by name, overloads latest-first; a
			// missing directive is reported before default-parameter ambiguity
			// and leaves the forward unimplemented.
			name: "forward ordering",
			source: `procedure Zeta; forward;
procedure Test(s : String); overload; forward;
function Test(i : Integer) : String; overload; forward;
procedure Alpha; forward;

function Test(i : Integer = 0) : String;
begin
end;
`,
			want: []string{
				`Syntax Error: Overloaded procedure "Test" must be marked with the "overload" directive [line: 6, column: 1]`,
				`Syntax Error: The function "Alpha" was forward declared but not implemented [line: 4, column: 11]`,
				`Syntax Error: The function "Test" was forward declared but not implemented [line: 3, column: 10]`,
				`Syntax Error: The function "Test" was forward declared but not implemented [line: 2, column: 11]`,
				`Syntax Error: The function "Zeta" was forward declared but not implemented [line: 1, column: 11]`,
			},
		},
		{
			// FailureScripts/forward_multiple1.
			name: "duplicate forward",
			source: `procedure Proc1; forward;
procedure Proc1; forward;

procedure Proc1;
begin
end;
`,
			want: []string{
				`Syntax Error: There is already a forward declaration of this function [line: 2, column: 18]`,
			},
		},
		{
			// FailureScripts/declaration_mismatch2: a mismatched implementation
			// still implements its forward.
			name: "mismatched implementation binds",
			source: `procedure Test1; forward;

procedure Test1(i : Integer);
begin
end;
`,
			want: []string{
				`Syntax Error: implementation signature for 'Test1' does not match forward declaration [line: 3, column: 1]`,
			},
		},
		{
			// FailureScripts/class_external: `forward` on an external routine
			// leaves nothing to implement.
			name:   "external forward",
			source: "procedure Ext; external; forward;\n",
			want:   nil,
		},
		{
			// FailureScripts/property_write3: an unknown name in an expression
			// is a compiler stop upstream, so the end-of-program check never runs.
			name: "compiler stop skips forward check",
			source: `procedure Proc1; forward;

PrintLn(bug);
`,
			want: []string{
				`Syntax Error: Unknown name "bug" [line: 3, column: 9]`,
			},
		},
		{
			// AssociativeFail/contains.
			name: "nil spelling",
			source: `var s : array [String] of Boolean;
PrintLn(nil in s);
`,
			want: []string{
				`Syntax Error: Incompatible types: "nil" and "String" [line: 2, column: 13]`,
			},
		},
		{
			// FailureScripts/incorrect_type1: the declared type's own spelling.
			name: "parameter type spelling",
			source: `procedure Log(AStr : string);
begin
end;

Log(1.5);
`,
			want: []string{
				`Syntax Error: Argument 0 expects type "String" instead of "Float" [line: 5, column: 5]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "test.pas", semantic.HintsLevelPedantic)
			var got []string
			for _, d := range result.DiagnosticStrings() {
				if !strings.HasPrefix(d, "Hint:") {
					got = append(got, d)
				}
			}
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

// TestCompile_UnitInterfaceForwards checks that routines declared in a unit's
// interface but never implemented are reported (OverloadsFail/forwards_unit),
// and that interface externals are not.
func TestCompile_UnitInterfaceForwards(t *testing.T) {
	source := `unit test;

interface

procedure Test; overload;
procedure Test(s : String); overload;
function Ext(v : Variant) : Integer; external;

implementation

procedure Test;
begin
end;

end.`
	result := Compile(source, "test.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: The function "Test" was forward declared but not implemented [line: 6, column: 11]`,
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
	}
}
