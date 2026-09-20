package frontend

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestCompile_HelperDiagnostics pins the complete diagnostic output for helper
// declarations and helper member access that DWScript rejects (PLAN.md §4, F5).
// Each case mirrors a fixture in testdata/fixtures/HelpersFail; the expectation is
// that fixture's `.txt` verbatim.
func TestCompile_HelperDiagnostics(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		want    []string
	}{
		{
			name:    "duplicate helper class var and class const",
			fixture: "helper_duplicate_member",
			want: []string{
				`Syntax Error: Name "Test" already exists [line: 5, column: 14]`,
				`Syntax Error: Name "Here" already exists [line: 7, column: 10]`,
			},
		},
		{
			name:    "helper method declared but never implemented",
			fixture: "helper_not_implemented",
			want: []string{
				`Syntax Error: Method "Length" of class "THelper" not implemented [line: 3, column: 11]`,
			},
		},
		{
			name:    "Self is not in scope in a static helper class method",
			fixture: "static_class_method_self",
			want: []string{
				`Syntax Error: Unknown name "Self" [line: 8, column: 12]`,
			},
		},
		{
			name:    "static on a non-class helper method",
			fixture: "helper_static",
			want: []string{
				`Syntax Error: Only non-virtual class methods can be marked as static [line: 5, column: 32]`,
				`Hint: Result is never used [line: 7, column: 3]`,
				`Syntax Error: Class method or constructor expected [line: 13, column: 11]`,
			},
		},
		{
			name:    "instance helper method reached through a class reference",
			fixture: "helper_error4",
			want: []string{
				`Syntax Error: Class method or constructor expected [line: 11, column: 9]`,
			},
		},
		{
			name:    "instance helper method reached through a type name",
			fixture: "integer_helper",
			want: []string{
				`Syntax Error: Class method or constructor expected [line: 18, column: 9]`,
				`Syntax Error: Class method or constructor expected [line: 19, column: 5]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := readHelperFixture(t, tt.fixture)
			result := Compile(source, "<test>", semantic.HintsLevelPedantic)
			got := result.DiagnosticStrings()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

// TestCompile_HelperDiagnosticsNegative keeps the new helper rules off valid code:
// each source below is accepted by DWScript and must stay diagnostic-free.
func TestCompile_HelperDiagnosticsNegative(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "class helper for a class reference with a static class method",
			source: "type\n   TClassHelper = class helper for TClass\n" +
				"      class function Hello : String; static;\n   end;\n" +
				"class function TClassHelper.Hello : String;\nbegin\n   Result := 'hi';\nend;\n" +
				"var c : TClass := TObject;\nPrintLn(c.Hello);\n",
		},
		{
			name: "record helper for a record",
			source: "type TRec = record x : Integer; end;\n" +
				"type TRecHelper = record helper for TRec\n   function Twice : Integer;\n   end;\n" +
				"function TRecHelper.Twice : Integer;\nbegin\n   Result := x * 2;\nend;\n" +
				"var r : TRec;\nr.x := 3;\nPrintLn(r.Twice);\n",
		},
		{
			name: "helper class method reached through the class reference",
			source: "type\n   THelper = helper for TObject\n      class function Hello : String;\n   end;\n" +
				"class function THelper.Hello : String;\nbegin\n   Result := 'hi';\nend;\n" +
				"PrintLn(TObject.Hello);\n",
		},
		{
			// A helper *name* is not the target type: TDummy.Two reaches the
			// helper's class method, and (1).Next the instance method.
			name: "explicit helper-name access is not a bare type name",
			source: "type\n   TDummy = helper for Integer\n      function Next : Integer; begin Result := Self + 1; end;\n" +
				"      class function Two : Integer; begin Result := 2; end;\n   end;\n" +
				"var i := (1).Next;\ni := TDummy.Two;\nPrintLn(i);\n",
		},
		{
			name: "non-static helper class method still sees Self",
			source: "type\n   THelper = helper for TObject\n      class function Hello : String;\n   end;\n" +
				"class function THelper.Hello : String;\nbegin\n   Result := Self.ClassName;\nend;\n" +
				"PrintLn(TObject.Hello);\n",
		},
		{
			name: "helper method with an out-of-line implementation is implemented",
			source: "type\n   THelper = helper for String\n      function Twice : String;\n   end;\n" +
				"function THelper.Twice : String;\nbegin\n   Result := Self + Self;\nend;\n" +
				"PrintLn('ab'.Twice);\n",
		},
		{
			// An inline body may call an overload declared before it.
			name: "overloaded helper methods called with matching arguments",
			source: "type TTest = helper for Integer\n" +
				"   procedure Hello(a, b : Integer); overload; begin PrintLn(a+b); end;\n" +
				"   procedure Hello(a : Integer); overload; begin Hello(a, a); end;\n" +
				"end;\n(1).Hello(2);\n(1).Hello(2, 3);\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			if got := result.DiagnosticStrings(); len(got) != 0 {
				t.Fatalf("expected no diagnostics, got %q", got)
			}
		})
	}
}

// TestCompile_HelperInlineBodySeesEarlierOverloadsOnly pins that an inline helper
// method body is compiled where it is written: HelpersFail/helper_overload_error
// calls a two-argument overload declared *after* the body, which DWScript rejects
// at the call name. go-dws now rejects it too; the exact sentence still reads
// "Too many arguments" because the wording is chosen by the shared argument-count
// path in analyze_function_calls.go, so only the anchor is pinned here.
func TestCompile_HelperInlineBodySeesEarlierOverloadsOnly(t *testing.T) {
	source := readHelperFixture(t, "helper_overload_error")
	got := Compile(source, "<test>", semantic.HintsLevelPedantic).DiagnosticStrings()
	if len(got) != 1 || !strings.Contains(got[0], "[line: 2, column: 50]") {
		t.Fatalf("expected one diagnostic anchored at line 2, column 50, got %q", got)
	}
}

func readHelperFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", "HelpersFail", name+".pas")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}
