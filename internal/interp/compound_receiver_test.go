package interp

import (
	"strings"
	"testing"
)

const compoundReceiverDeclarations = `
type TTarget = class
  class var Created: Integer;
  class var Reads: Integer;
  class var Writes: Integer;
  class var First: TTarget;
  class var Last: TTarget;
  Value: Integer;
  constructor Create;
  begin
    TTarget.Created += 1;
    TTarget.Last := Self;
    if TTarget.First = nil then TTarget.First := Self;
    Value := 10 * TTarget.Created;
  end;
  function GetValue: Integer;
  begin
    TTarget.Reads += 1;
    Result := Value;
  end;
  procedure SetValue(value: Integer);
  begin
    TTarget.Writes += 1;
    Self.Value := value;
  end;
  property Prop: Integer read GetValue write SetValue;
end;
type TChild = class(TTarget) end;
`

func TestCompoundReceiver_ConstructorEvaluatedOnce(t *testing.T) {
	for _, receiver := range []string{"TTarget.Create", "TTarget.Create()", "TChild.Create"} {
		for _, member := range []string{"Value", "Prop"} {
			for _, operation := range []struct{ operator, want string }{
				{"+=", "12"}, {"-=", "8"}, {"*=", "20"}, {"/=", "5"},
			} {
				t.Run(receiver+"."+member+operation.operator, func(t *testing.T) {
					got := runQuickwinScript(t, compoundReceiverDeclarations+receiver+"."+member+" "+operation.operator+` 2;
PrintLn(TTarget.Created);
PrintLn(TTarget.Reads);
PrintLn(TTarget.Writes);
PrintLn(TTarget.First.Value);
PrintLn(TTarget.First = TTarget.Last);
`)
					counts := "0\n0\n"
					if member == "Prop" {
						counts = "1\n1\n"
					}
					assertOutput(t, got, "1\n"+counts+operation.want+"\nTrue\n")
				})
			}
		}
	}
}

func TestCompoundReceiver_RHSPreservesReceiver(t *testing.T) {
	for _, member := range []string{"Value", "Prop"} {
		t.Run(member, func(t *testing.T) {
			got := runQuickwinScript(t, compoundReceiverDeclarations+`
var target := TTarget.Create;
function ReplaceTarget: Integer;
begin
  target := TTarget.Create;
  Result := 2;
end;
target.`+member+` += ReplaceTarget();
PrintLn(TTarget.Created);
PrintLn(TTarget.First.Value);
PrintLn(TTarget.Last.Value);
`)
			assertOutput(t, got, "2\n12\n20\n")
		})
	}
}

func TestCompoundReceiver_FailureStopsWrite(t *testing.T) {
	for _, tt := range []struct {
		name, constructor, getter, rhs, operation, message, counts string
	}{
		{"receiver exception", "raise Exception.Create('receiver failed');", "", "Result := 2;", "+=", "receiver failed", "1\n0\n0\n0\n"},
		{"receiver runtime error", "var zero := 0; PrintLn(1 div zero);", "", "Result := 2;", "+=", "Division by zero", "1\n0\n0\n0\n"},
		{"getter exception", "", "raise Exception.Create('getter failed');", "Result := 2;", "+=", "getter failed", "1\n1\n0\n0\n"},
		{"getter runtime error", "", "var zero := 0; PrintLn(1 div zero);", "Result := 2;", "+=", "Division by zero", "1\n1\n0\n0\n"},
		{"rhs exception", "", "", "raise Exception.Create('rhs failed');", "+=", "rhs failed", "1\n1\n1\n0\n"},
		{"rhs runtime error", "", "", "var zero := 0; Result := 1 div zero;", "+=", "Division by zero", "1\n1\n1\n0\n"},
		{"operation runtime error", "", "", "Result := 0;", "/=", "Division by zero", "1\n1\n1\n0\n"},
	} {
		for _, receiver := range []string{"TTarget.Create", "TTarget.Create()"} {
			t.Run(tt.name+"/"+receiver, func(t *testing.T) {
				got := runQuickwinScript(t, `
var constructions, reads, rightSides, writes: Integer;
type TTarget = class
  constructor Create;
  begin
    constructions += 1;
    `+tt.constructor+`
  end;
  function GetValue: Integer;
  begin
    reads += 1;
    `+tt.getter+`
    Result := 10;
  end;
  procedure SetValue(value: Integer);
  begin writes += 1; end;
  property Prop: Integer read GetValue write SetValue;
end;
function RightSide: Integer;
begin
  rightSides += 1;
  `+tt.rhs+`
end;
try
  `+receiver+`.Prop `+tt.operation+` RightSide();
except
  on E: Exception do PrintLn(E.Message);
end;
PrintLn(constructions);
PrintLn(reads);
PrintLn(rightSides);
PrintLn(writes);
`)
				message, counts, _ := strings.Cut(got, "\n")
				if !strings.HasPrefix(message, tt.message) || counts != tt.counts {
					t.Fatalf("expected %q and counts %q, got %q", tt.message, tt.counts, got)
				}
			})
		}
	}
}

func TestCompoundReceiver_OperatorExceptionSkipsSetter(t *testing.T) {
	got := runQuickwinScript(t, `
var constructions, reads, writes, operations: Integer;
type TNumber = class
  procedure Add(value: Integer);
  begin
    operations += 1;
    raise Exception.Create('operator failed');
  end;
  class operator += Integer uses Add;
end;
type THolder = class
  Number: TNumber;
  constructor Create;
  begin
    constructions += 1;
    Number := TNumber.Create;
  end;
  function GetNumber: TNumber;
  begin reads += 1; Result := Number; end;
  procedure SetNumber(value: TNumber);
  begin writes += 1; Number := value; end;
  property Prop: TNumber read GetNumber write SetNumber;
end;
try
  THolder.Create.Prop += 2;
except
  on E: Exception do PrintLn(E.Message);
end;
PrintLn(constructions);
PrintLn(reads);
PrintLn(operations);
PrintLn(writes);
`)
	assertOutput(t, got, "operator failed\n1\n1\n1\n0\n")
}

func TestCompoundReceiver_StaticFieldIdentity(t *testing.T) {
	got := runQuickwinScript(t, `
type TBase = class Value: Integer; end;
type TChild = class(TBase) Value: Integer; end;
var calls: Integer;
var target := TChild.Create;
function Receiver: TChild;
begin calls += 1; Result := target; end;
TBase(target).Value := 10;
target.Value := 20;
TBase(Receiver()).Value += 2;
PrintLn(calls);
PrintLn(TBase(target).Value);
PrintLn(target.Value);
`)
	assertOutput(t, got, "1\n12\n20\n")
}

func TestCompoundReceiver_PropertyDispatch(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"helper", compoundReceiverDeclarations + `
type TTargetHelper = class helper for TTarget
  property HelperProp: Integer read (0 + Self.Value) write (Self.Value := Value);
end;
TTarget.Create.HelperProp += 2;
PrintLn(TTarget.Created);
PrintLn(TTarget.First.Value);
`, "1\n12\n"},
		{"interface", `
type IValue = interface
  function GetValue: Integer;
  procedure SetValue(value: Integer);
  property Prop: Integer read GetValue write SetValue;
end;
type TTarget = class(TObject, IValue)
  Value: Integer;
  function GetValue: Integer; begin Result := Value; end;
  procedure SetValue(value: Integer); begin Self.Value := value; end;
end;
var calls: Integer;
var target := TTarget.Create;
target.Value := 10;
function Receiver: IValue;
begin calls += 1; Result := target; end;
Receiver().Prop += 2;
PrintLn(calls);
PrintLn(target.Value);
`, "1\n12\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, tt.source), tt.want)
		})
	}
}
