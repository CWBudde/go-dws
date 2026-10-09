package dwscript

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPropertyIndexUseSite_GroupedReferences(t *testing.T) {
	source := `type IntAlias = Integer;
var Trace := '';
type T = class
 F: Integer;
 function Get(var Access: IntAlias): Integer; begin Access += 1; Result := Access; end;
 procedure Put(var Store: IntAlias; V: Integer); begin Store := V; end;
 property P[var Declared: IntAlias]: Integer read Get write Put;
end;
var O := new T; var X := 5; var A: array of Integer := [7];
function Container: array of Integer; begin Trace += 'A'; Result := A; end;
function Slot: Integer; begin Trace += 'I'; Result := 0; end;
function Receiver: T; begin Trace += 'R'; Result := O; end;
procedure Ordinary(var Y: Integer); begin Y += 1; end;
procedure ForwardIndex(var Y: Integer);
begin PrintLn(O.P[((Y))]); O.P[(Y)] := 9; Ordinary((Y)); end;
ForwardIndex(X); PrintLn(X);
PrintLn(O.P[(Container()[Slot()])]); PrintLn(A[0]);
O.F := 2; PrintLn(O.P[((Receiver().F))]); PrintLn(O.F);
PrintLn(Trace);`
	for _, checked := range []bool{true, false} {
		t.Run(fmt.Sprint(checked), func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithTypeCheck(checked), WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = engine.Eval(source); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != "6\n10\n8\n8\n3\n3\nAIR\n" {
				t.Fatalf("output %q", got)
			}
		})
	}
}

func TestPropertyIndexUseSite_GroupedOldContracts(t *testing.T) {
	const source = `function Double(X: Integer): Integer;
begin Result := X * 2; end;
ensure Result = ((old X) * 2);
function AddDoubled(A, B: Integer): Integer;
begin Result := Double(A) + Double(B); end;
ensure Result = (((old A) * 2)) + (((old B) * 2));
begin PrintLn(AddDoubled(3, 4)); end.`
	for _, checked := range []bool{true, false} {
		t.Run(fmt.Sprint(checked), func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithTypeCheck(checked), WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = engine.Eval(source); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != "14\n" {
				t.Fatalf("output %q", got)
			}
		})
	}
}
