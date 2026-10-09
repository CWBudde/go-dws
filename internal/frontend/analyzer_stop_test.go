package frontend

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestCompile_AnalyzerStop pins that an analyzer compiler stop is a real stop:
// upstream's AddCompilerStop raises ECompileError, so nothing positioned after
// it is reported, whatever phase found it in go-dws, and the end-of-program
// checks never run. Diagnostics of routine bodies declared before the stop are
// still reported, though go-dws analyzes those bodies after the main program.
func TestCompile_AnalyzerStop(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			// HelpersFail/strict: ReportNoMemberForType raises
			// CPE_UnknownMemberForType as a stop, so the parser errors of
			// line 11 are never reached.
			name: "unknown member cuts later parser diagnostics",
			source: `type MyString = String;

type TMyHelper = strict helper for MyString
      const Hello = 'Hello';
   end;

var s : String;

PrintLn(s.Hello);

type TMyBug = strict for String end;
`,
			want: []string{
				`Syntax Error: There is no accessible member with name "Hello" for type String [line: 9, column: 11]`,
			},
		},
		{
			name: "unknown member skips the end-of-program forward check",
			source: `procedure P; forward;
var s : String;
PrintLn(s.Hello);
`,
			want: []string{
				`Syntax Error: There is no accessible member with name "Hello" for type String [line: 3, column: 11]`,
			},
		},
		{
			name: "unknown member cuts later semantic diagnostics",
			source: `var s : String;
PrintLn(s.Hello);
PrintLn(Undefined);
`,
			want: []string{
				`Syntax Error: There is no accessible member with name "Hello" for type String [line: 2, column: 11]`,
			},
		},
		{
			name: "routine bodies before the stop keep their diagnostics",
			source: `procedure P;
begin
  PrintLn(1 as String);
end;
var s : String;
PrintLn(s.Hello);
procedure Q;
begin
  PrintLn(AlsoUndefined);
end;
`,
			want: []string{
				`Syntax Error: Cannot cast "Integer" as "String" [line: 3, column: 13]`,
				`Syntax Error: There is no accessible member with name "Hello" for type String [line: 6, column: 11]`,
			},
		},
		{
			// ReadName raises CPE_UnknownName as a stop (dwsCompiler.pas:4910),
			// so neither the second argument nor the call's arity is checked.
			name: "unknown name in an argument stops the call",
			source: `procedure Take(v: Integer);
begin
end;
Take(Missing1, Missing2);
PrintLn(Undefined);
`,
			want: []string{
				`Syntax Error: Unknown name "Missing1" [line: 4, column: 6]`,
			},
		},
		{
			// ReadName (dwsCompiler.pas:4910) for the callee of a call.
			name:   "unknown routine name stops",
			source: "Missing(1);\nPrintLn(Undefined);\n",
			want:   []string{`Syntax Error: Unknown name "Missing" [line: 1, column: 1]`},
		},
		{
			// ReadAt reads its operand through ReadName.
			name:   "unknown address-of operand stops",
			source: "var p := @Missing;\nPrintLn(Undefined);\n",
			want:   []string{`Syntax Error: Unknown name "Missing" [line: 1, column: 10]`},
		},
		{
			// ReadName on a unit prefix: CPE_UnknownNameDotName (dwsCompiler.pas:4941).
			name:   "unknown unit member stops",
			source: "PrintLn(System.Missing);\nPrintLn(Undefined);\n",
			want:   []string{`Syntax Error: Unknown name "System.Missing" [line: 1, column: 16]`},
		},
		{
			// ReadForIn: CPE_ArrayExpected (dwsCompiler.pas:6526).
			name:   "for-in over a non-container stops",
			source: "var i : Integer;\nfor i in 5 do ;\nPrintLn(Undefined);\n",
			want:   []string{`Syntax Error: Array expected [line: 2, column: 7]`},
		},
		{
			// CPE_NoDefaultProperty (dwsCompiler.pas:5704).
			name:   "indexing a class without default property stops",
			source: "type TObj = class end;\nvar o := TObj.Create;\nPrintLn(o[1]);\nPrintLn(Undefined);\n",
			want:   []string{`Syntax Error: Class "TObj" has no default property [line: 3, column: 10]`},
		},
		{
			// CPE_RecordTypeNotFullyDefined (dwsCompiler.pas:7962).
			name:   "record containing itself stops",
			source: "type R = record\n  Inner : R;\nend;\nPrintLn(Undefined);\n",
			want:   []string{`Syntax Error: Record type "R" is not fully defined [line: 2, column: 12]`},
		},
		{
			name: "a stop in a routine body cuts the later main program",
			source: `var s : String;
procedure P;
begin
  PrintLn(s.Hello);
end;
PrintLn(Undefined);
`,
			want: []string{
				`Syntax Error: There is no accessible member with name "Hello" for type String [line: 4, column: 13]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "test.pas", semantic.HintsLevelPedantic)
			got := result.DiagnosticStrings()
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}
