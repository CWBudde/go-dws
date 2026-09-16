package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestArrayMathHelpers_Fixtures(t *testing.T) {
	for _, fixture := range []string{"FunctionsMath/array_funcs.pas", "ArrayPass/string_array_pack.pas"} {
		t.Run(fixture, func(t *testing.T) {
			result, detail := runFixtureTest(filepath.Join(fixturesRoot, filepath.FromSlash(fixture)), false, semantic.HintsLevelPedantic)
			if result != testResultPassed {
				t.Fatalf("result=%v: %s", result, detail)
			}
		})
	}
}

func TestArrayMathHelpers_Execution(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{
			"same receiver and chaining",
			`var values: array of Float := [1, 2];
var calls := 0;
function ValuesOnce: array of Float;
begin Inc(calls); Result := values; end;
var alias := ValuesOnce().Offset(1).Multiply(2).Reciprocal().MultiplyAdd(2, 1);
PrintLn(calls);
PrintLn(values[0]);
alias[0] := 99;
PrintLn(values[0]);`,
			"1\n1.5\n99\n",
		},
		{
			"named arrays and case-insensitive helpers",
			`type TValues = array of Float;
type TNames = array of String;
var values: TValues := [2];
var names: TNames := ['', 'a'];
values.oFfSeT(2).mUlTiPlY(3).mUlTiPlYaDd(2, 1).rEcIpRoCaL;
PrintLn(values[0]);
PrintLn(names.pAcK().Join('|'));`,
			"0.04\na\n",
		},
		{
			"empty arrays skip scalar operands",
			`var values: array of Float;
function Operand(labelText: String; value: Float): Float;
begin PrintLn(labelText); Result := value; end;
values.Offset(Operand('offset skipped', 1)).Multiply(Operand('multiply skipped', 2));
values.MultiplyAdd(Operand('scale skipped', 3), Operand('add skipped', 4)).Reciprocal;
PrintLn(values.Length);
values.Add(2);
values.MultiplyAdd(Operand('scale', 3), Operand('add', 4));
PrintLn(values[0]);`,
			"0\nscale\nadd\n10\n",
		},
		{
			"operand exception leaves receiver unchanged",
			`var values: array of Float := [2, 4];
function Fail: Float;
begin raise Exception.Create('operand failed'); end;
try values.MultiplyAdd(3, Fail()); except on E: Exception do PrintLn(E.Message); end;
PrintLn(values[0]); PrintLn(values[1]);
values.Clear;
values.Offset(Fail());
PrintLn('empty remains silent');`,
			"operand failed\n2\n4\nempty remains silent\n",
		},
		{
			"receiver and first operand exceptions stop evaluation",
			`var values: array of Float := [2];
function Mark: Float;
begin PrintLn('unexpected operand'); Result := 3; end;
function Fail: Float;
begin raise Exception.Create('operand failed'); end;
function FailReceiver: array of Float;
begin raise Exception.Create('receiver failed'); end;
try values.MultiplyAdd(Fail(), Mark()); except on E: Exception do PrintLn(E.Message); end;
try FailReceiver().Offset(Mark()); except on E: Exception do PrintLn(E.Message); end;
PrintLn(values[0]);`,
			"operand failed\nreceiver failed\n2\n",
		},
		{
			"user helper overrides empty-array intrinsic",
			`type TValues = array of Float;
type TCustomHelper = helper for TValues
  function Offset(value: Float): Integer;
  begin PrintLn('override'); Result := 7; end;
end;
function Operand: Float;
begin PrintLn('operand'); Result := 2; end;
var values: TValues;
PrintLn(values.Offset(Operand()));`,
			"operand\noverride\n7\n",
		},
		{
			"global pack accepts literals",
			`PrintLn(StrArrayPack(['', 'a', '']).Join('|'));`,
			"a\n",
		},
		{
			"pack helper and global function share mutation",
			`var values: array of String := ['', 'a', '', ' ', 'b', ''];
var alias := values.Pack;
PrintLn(values.Join('|'));
alias.Add('c');
PrintLn(values.Join('|'));
values.Add('');
var packed := StrArrayPack(values);
packed.Add('d');
PrintLn(values.Join('|'));
PrintLn(values.Pack().Join('|'));`,
			"a| |b\na| |b|c\na| |b|c|d\na| |b|c|d\n",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, tt.source), tt.want)
		})
	}
}

func TestArrayMathHelpers_RejectInvalidCalls(t *testing.T) {
	for _, source := range []string{
		`var a: array of Integer; a.Offset(1);`,
		`var a: array of String; a.Multiply(2);`,
		`var a: array of Float; a.Pack;`,
		`var a: array [0..1] of Float; a.Offset(1);`,
		`var a: array [0..1] of String; a.Pack;`,
		`var a: array of Float; a.Offset;`,
		`var a: array of Float; a.MultiplyAdd(1);`,
		`var a: array of Float; a.Reciprocal(1);`,
		`var a: array of Float; a.Offset('bad');`,
		`var a: array of String; a.Pack(1);`,
	} {
		t.Run(source, func(t *testing.T) { assertCompileError(t, source, "") })
	}
}
