package interp

import "testing"

func TestPartialContinuation_CompileAndRun(t *testing.T) {
	for _, test := range []struct {
		name, source, want string
	}{
		{
			name: "nonpartial continuation retains fields and methods",
			source: `type TTest = partial class
  A: Integer = 20;
  function First: Integer; begin Result := A; end;
end;
type TTest = class
  B: Integer = 22;
  function Second: Integer; begin Result := B; end;
end;
var item := new TTest;
PrintLn(item.First + item.Second);`,
			want: "42\n",
		},
		{
			name: "nonpartial continuation keeps class open",
			source: `type TTest = partial class A: Integer = 10; end;
type TTest = class B: Integer = 20; end;
type TTest = partial class C: Integer = 12; end;
var item := new TTest;
PrintLn(item.A + item.B + item.C);`,
			want: "42\n",
		},
		{
			name: "inherited and case insensitive continuation",
			source: `type TBase = class A: Integer = 20; end;
type TTest = partial class(TBase) B: Integer = 10; end;
type ttest = class C: Integer = 12; end;
var item := new TTest;
PrintLn(item.A + item.B + item.C);`,
			want: "42\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, test.source), test.want)
		})
	}
}

func TestPartialContinuation_OrdinaryDuplicateRejected(t *testing.T) {
	assertCompileError(t, `type TTest = class end;
type TTest = class end;`, "already defined")
}
