package interp

import (
	"strings"
	"testing"
)

const varReceiverDeclarations = `
type TTarget = class
  class var Created: Integer;
  class var Last: TTarget;
  Value: Integer;
  constructor Create;
  begin
    TTarget.Created += 1;
    TTarget.Last := Self;
    Value := 10;
  end;
  property Prop: Integer read Value write Value;
end;
type TChild = class(TTarget) end;
`

func TestVarReceiver_LiveCapturedField(t *testing.T) {
	for _, receiver := range []string{"TTarget.Create", "TTarget.Create()", "TChild.Create", "TChild.Create()"} {
		for _, call := range []string{"routine", "overload", "method"} {
			t.Run(receiver+"/"+call, func(t *testing.T) {
				declaration := `procedure Mutate(var value: Integer; obj: TTarget);`
				if call == "overload" {
					declaration += " overload;"
				}
				declaration += `begin obj.Value := 40; PrintLn(value); value += 2; end;`
				invocation := "Mutate"
				if call == "overload" {
					declaration += `procedure Mutate(var value: String; obj: TTarget); overload; begin PrintLn('wrong overload'); end;`
				}
				if call == "method" {
					declaration = "type TMutator = class " + declaration + " end; var m := TMutator.Create;"
					invocation = "m.Mutate"
				}
				got := runQuickwinScript(t, varReceiverDeclarations+declaration+invocation+"("+receiver+`.Value, TTarget.Last);
PrintLn(TTarget.Created);
PrintLn(TTarget.Last.Value);
PrintLn(TTarget.Last.ClassName);
`)
				className := "TTarget"
				if strings.HasPrefix(receiver, "TChild") {
					className = "TChild"
				}
				assertOutput(t, got, "40\n1\n42\n"+className+"\n")
			})
		}
	}
}

func TestVarReceiver_LaterArgumentChangesCapturedField(t *testing.T) {
	for _, receiver := range []string{"TTarget.Create", "TTarget.Create()"} {
		t.Run(receiver, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, varReceiverDeclarations+`
function Later: Integer;
begin TTarget.Last.Value := 40; Result := 0; end;
procedure Mutate(var value: Integer; unused: Integer);
begin PrintLn(value); value += 2; end;
Mutate(`+receiver+`.Value, Later());
PrintLn(TTarget.Created);
PrintLn(TTarget.Last.Value);
`), "40\n1\n42\n")
		})
	}
}

func TestVarReceiver_PropertyFallbackEvaluatesOnce(t *testing.T) {
	for _, receiver := range []string{"TTarget.Create", "TTarget.Create()"} {
		t.Run(receiver, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, varReceiverDeclarations+`
procedure Mutate(var value: Integer);
begin value += 2; end;
Mutate(`+receiver+`.Prop);
PrintLn(TTarget.Created);
PrintLn(TTarget.Last.Value);
`), "1\n12\n")
		})
	}
}

func TestVarReceiver_LaterArgumentKeepsOriginalInstance(t *testing.T) {
	for _, receiver := range []string{"TTarget.Create", "TTarget.Create()"} {
		t.Run(receiver, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, varReceiverDeclarations+`
var original: TTarget;
function Later: Integer;
begin original := TTarget.Last; TTarget.Create; Result := 0; end;
procedure Mutate(var value: Integer; unused: Integer);
begin value := 42; end;
Mutate(`+receiver+`.Value, Later());
PrintLn(TTarget.Created);
PrintLn(original.Value);
PrintLn(TTarget.Last.Value);
`), "2\n42\n10\n")
		})
	}
}

func TestVarReceiver_ExceptionSkipsArgumentsAndCall(t *testing.T) {
	for _, receiver := range []string{"TTarget.Create", "TTarget.Create()"} {
		for _, failure := range []string{"raise Exception.Create('receiver failed');", "var zero := 0; PrintLn(1 div zero);"} {
			t.Run(receiver+"/"+failure, func(t *testing.T) {
				declarations := strings.Replace(varReceiverDeclarations, "Value := 10;", "Value := 10; "+failure, 1)
				got := runQuickwinScript(t, declarations+`
function Later: Integer;
begin PrintLn('later argument'); Result := 0; end;
procedure Mutate(var value: Integer; unused: Integer);
begin PrintLn('called'); value := 42; end;
try Mutate(`+receiver+`.Value, Later());
except on E: Exception do PrintLn(E.Message); end;
PrintLn(TTarget.Created);
PrintLn(TTarget.Last.Value);
`)
				if strings.HasPrefix(failure, "raise") {
					assertOutput(t, got, "receiver failed\n1\n10\n")
				} else if !strings.HasPrefix(got, "Division by zero in TTarget.Create ") || !strings.HasSuffix(got, "\n1\n10\n") || strings.Contains(got, "later argument") || strings.Contains(got, "called") {
					t.Fatalf("expected original constructor error and no later arguments/call, got %q", got)
				}
			})
		}
	}
}
