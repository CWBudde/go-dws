package interp

import (
	"strings"
	"testing"
)

func TestFunctionReceiverAssignment_EvaluatedOnce(t *testing.T) {
	const declarations = `
type TTarget = class
  Value: Integer;
  procedure SetValue(v: Integer);
  begin PrintLn('setter'); Value := v; end;
  property Prop: Integer read Value write SetValue;
end;
var Calls: Integer;
var Last: TTarget;
function Make: TTarget;
begin
  Inc(Calls);
  Last := TTarget.Create;
  Result := Last;
end;
`
	for _, receiver := range []string{"Make", "Make()", "mAkE"} {
		for _, member := range []string{"Value", "Prop"} {
			t.Run(receiver+"."+member, func(t *testing.T) {
				want := "1\n42\n"
				if member == "Prop" {
					want = "setter\n" + want
				}
				assertOutput(t, runQuickwinScript(t, declarations+receiver+"."+member+` := 42;
PrintLn(Calls);
PrintLn(Last.Value);
`), want)
			})
		}
	}
}

func TestFunctionReceiverAssignment_LocalFunction(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
type TTarget = class Value: Integer; end;
procedure Test;
begin
var Last: TTarget;
var Calls: Integer;
function Make: TTarget;
begin Inc(Calls); Last := TTarget.Create; Result := Last; end;
  Make.Value := 42;
  PrintLn(Calls);
  PrintLn(Last.Value);
end;
Test;
`), "1\n42\n")
}

func TestFunctionReceiverAssignment_FailureSkipsWrite(t *testing.T) {
	for _, failure := range []string{"raise Exception.Create('receiver failed');", "var zero := 0; PrintLn(1 div zero);"} {
		for _, receiver := range []string{"Make", "Make()"} {
			t.Run(receiver+failure, func(t *testing.T) {
				got := runQuickwinScript(t, `
type TTarget = class
  Value: Integer;
  procedure SetValue(v: Integer); begin PrintLn('unexpected setter'); end;
  property Prop: Integer read Value write SetValue;
end;
var Calls: Integer;
function Make: TTarget;
begin Inc(Calls); `+failure+` end;
try `+receiver+`.Prop := 42;
except on E: Exception do PrintLn(E.Message); end;
PrintLn(Calls);
`)
				lines := strings.Split(strings.TrimSpace(got), "\n")
				wantMessage := "receiver failed"
				if strings.Contains(failure, "zero") {
					wantMessage = "Division by zero"
				}
				if len(lines) != 2 || !strings.HasPrefix(lines[0], wantMessage) || lines[1] != "1" {
					t.Fatalf("expected original failure and one call, got %q", got)
				}
			})
		}
	}
}

func TestFunctionReceiverAssignment_ResultAndVariableStorage(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
type TTarget = class Value: Integer; end;
function Make: TTarget;
begin
  Result := TTarget.Create;
  Make.Value := 40;
end;
var obj := Make();
obj.Value += 2;
PrintLn(obj.Value);
procedure Test;
var Make := TTarget.Create;
begin Make.Value := 7; PrintLn(Make.Value); end;
Test;
`), "42\n7\n")
}
