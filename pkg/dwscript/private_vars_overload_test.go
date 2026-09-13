package dwscript

import "testing"

func TestPrivateVars_BuiltinOverloadKeepsDefaultLazy(t *testing.T) {
	for _, tc := range []struct {
		name, overload, want string
	}{
		{"different arity", `function ReadPrivateVar(n: Integer): Variant; overload;
begin Result := n; end;`, "42\n0\n"},
		{"less specific user signature", `function ReadPrivateVar(n, fallback: Variant): Variant; overload;
begin Result := fallback; end;`, "42\n0\n"},
		{"more specific user signature", `function ReadPrivateVar(n: String; fallback: Integer): Variant; overload;
begin Result := fallback; end;`, "99\n1\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine, program, output := compilePrivateVarsProgram(t, map[string]string{
				"PrivateOverloadUnit": `unit PrivateOverloadUnit;
interface
procedure Run;
implementation
var Hits: Integer;
function Fallback: Integer;
begin Hits += 1; Result := 99; end;
procedure Run;
begin
  CleanupPrivateVars;
  WritePrivateVar('present', 42);
  PrintLn(ReadPrivateVar('present', Fallback()));
  PrintLn(Hits);
  CleanupPrivateVars;
end;
end.`,
			}, "uses PrivateOverloadUnit;\n"+tc.overload+"\nPrivateOverloadUnit.Run();")
			runPrivateVarsProgram(t, engine, program, output, tc.want)
		})
	}
}
