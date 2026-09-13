package interp

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestNumericHelpers_Fixtures(t *testing.T) {
	for _, name := range []string{"compare_num", "sort_nums", "testbit", "popcnt"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, "FunctionsMath", name+".pas")
			result, detail := runFixtureTest(path, false, semantic.HintsLevelPedantic)
			if result != testResultPassed {
				t.Fatalf("FunctionsMath/%s: %v: %s", name, result, detail)
			}
		})
	}
}

func TestNumericHelpers_PrecisionAndSpecialValues(t *testing.T) {
	got := runQuickwinScript(t, `
var large: Integer := 9007199254740993;
var smaller: Integer := 9007199254740992;
var rounded: Float := 9007199254740992.0;
PrintLn(large.Compare(smaller));
PrintLn(smaller.Compare(large));
PrintLn(CompareNum(large, smaller));
PrintLn(large.Compare(rounded));
PrintLn(rounded.Compare(large));
var special := NaN;
PrintLn(special.Compare(1));
PrintLn(special.Compare(special));
PrintLn(CompareNum(1, special));
var bits: Integer := -1;
PrintLn(bits.TestBit(-1));
PrintLn(bits.TestBit(63));
PrintLn(bits.TestBit(64));
PrintLn(TestBit(bits, 64));
PrintLn(bits.PopCount);
PrintLn(bits.PopCount());
`)
	assertOutput(t, got, "1\n-1\n1\n0\n0\n1\n1\n1\nFalse\nTrue\nFalse\nFalse\n64\n64\n")
}

func TestNumericHelpers_EvaluateReceiverAndArgumentsOnce(t *testing.T) {
	got := runQuickwinScript(t, `
var receivers := 0;
var arguments := 0;
function Next: Integer;
begin receivers += 1; Result := 7; end;
function Other: Integer;
begin arguments += 1; Result := 8; end;
PrintLn(Next().pOpCoUnT);
PrintLn(Next().cOmPaRe(Other()));
PrintLn(Next().tEsTbIt(Other()));
PrintLn(receivers);
PrintLn(arguments);
`)
	assertOutput(t, got, "3\n-1\nFalse\n3\n2\n")
}

func TestNumericHelpers_UserHelperTakesPrecedence(t *testing.T) {
	got := runQuickwinScript(t, `
type TCustomInteger = helper for Integer
  function PopCount: Integer;
  begin Result := 123; end;
  function Compare(other: Integer): Integer;
  begin Result := 456; end;
  function TestBit(bit: Integer): Boolean;
  begin Result := False; end;
end;
var number := 7;
PrintLn(number.PopCount);
PrintLn(number.Compare(7));
PrintLn(number.TestBit(0));
PrintLn(PopCount(number));
PrintLn(CompareNum(number, 7));
PrintLn(TestBit(number, 0));
`)
	assertOutput(t, got, "123\n456\nFalse\n3\n0\nTrue\n")
}

func TestNumericHelpers_RejectInvalidCalls(t *testing.T) {
	for _, call := range []string{
		"number.Compare()", "number.Compare('text')", "number.Compare(True)",
		"number.TestBit()", "number.TestBit(1.5)", "number.TestBit('text')",
		"number.PopCount(1)", "fraction.PopCount()", "fraction.TestBit(0)",
	} {
		t.Run(call, func(t *testing.T) {
			compiled := frontend.Compile("var number := 7; var fraction := 1.5; "+call+";", "numeric_helpers.pas", semantic.HintsLevelNormal)
			if !compiled.HasFatalDiagnostics() || compiled.SemanticSuccessful {
				t.Fatalf("invalid helper call compiled: %s; diagnostics=%v", call, compiled.DiagnosticStrings())
			}
		})
	}
}

func TestNumericHelpers_AliasReceivers(t *testing.T) {
	got := runQuickwinScript(t, `
type TCount = Integer;
type TRatio = Float;
var count: TCount := 7;
var ratio: TRatio := 7.5;
PrintLn(count.PopCount);
PrintLn(count.TestBit(2));
PrintLn(count.Compare(ratio));
PrintLn(ratio.Compare(count));
`)
	assertOutput(t, got, "3\nTrue\n-1\n1\n")
}

func TestNumericHelpers_FailuresStopCalls(t *testing.T) {
	for _, failure := range []string{
		"raise Exception.Create('numeric failure');",
		"var zero := 0; Result := 1 div zero;",
	} {
		for _, call := range []string{"Fail().Compare(Mark())", "Mark().Compare(Fail())"} {
			t.Run(failure+call, func(t *testing.T) {
				got := runQuickwinScript(t, `
var marked := 0;
function Fail: Integer;
begin `+failure+` end;
function Mark: Integer;
begin marked += 1; Result := 7; end;
try
  PrintLn(`+call+`);
except
  on E: Exception do PrintLn(E.Message);
end;
PrintLn(marked);
`)
				lines := strings.Split(strings.TrimSpace(got), "\n")
				wantCount := "0"
				if strings.HasPrefix(call, "Mark()") {
					wantCount = "1"
				}
				wantMessage := "numeric failure"
				if strings.HasPrefix(failure, "var zero") {
					wantMessage = "Division by zero"
				}
				if len(lines) != 2 || !strings.HasPrefix(lines[0], wantMessage) || lines[1] != wantCount {
					t.Fatalf("failed numeric call output=%q; want original %q and receiver count %s", got, wantMessage, wantCount)
				}
			})
		}
	}
}
