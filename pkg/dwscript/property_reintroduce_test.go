package dwscript

import (
	"bytes"
	"testing"
)

// A compatibility pair reads the property once; receiver and getter side
// effects must never be duplicated by dispatching the call as a method.
func TestReintroducedProperty_Execution(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"field backed", `type TTest = class
 Field: Integer;
 property Prop: Integer read Field reintroduce;
end;
var Obj := new TTest;
Obj.Field := 123;
PrintLn(Obj.Prop()); PrintLn(Obj.Prop);`, "123\n123\n"},
		{"getter and receiver once", `var Receivers := 0;
type TTest = class
 Reads: Integer;
 function GetProp: Integer;
 begin Inc(Reads); Result := Reads; end;
 property Prop: Integer read GetProp reintroduce;
end;
var Obj := new TTest;
function NextObj: TTest;
begin Inc(Receivers); Result := Obj; end;
PrintLn(NextObj().Prop()); PrintLn(Receivers); PrintLn(Obj.Reads);`, "1\n1\n1\n"},
		{"inherited descriptor", `type TBase = class
 Field: Integer;
 property Prop: Integer read Field reintroduce;
end;
type TChild = class(TBase) end;
var Obj := new TChild; Obj.Field := 9;
PrintLn(Obj.pRoP());`, "9\n"},
		{"class property", `type TTest = class
 class var Field: Integer;
 class property Prop: Integer read Field reintroduce;
end;
TTest.Field := 7; PrintLn(TTest.Prop());`, "7\n"},
		{"descendant method shadows inherited property", `type TBase = class
 Field: Integer;
 property Prop: Integer read Field reintroduce;
end;
type TChild = class(TBase)
 function Prop: Integer; reintroduce;
 begin Result := 12; end;
end;
var Obj := new TChild;
PrintLn(Obj.Prop());`, "12\n"},
		{"helper method shadows scalar property", `type TTest = class
 Field: Integer;
 property Prop: Integer read Field;
end;
type THelper = helper for TTest
 function Prop: Integer; begin Result := 9; end;
end;
var Obj := new TTest;
PrintLn(Obj.Prop());`, "9\n"},
		{"alias receiver", `type TTest = class
 Field: Integer;
 property Prop: Integer read Field reintroduce;
end;
type TAlias = TTest;
var Obj: TAlias := new TTest;
Obj.Field := 11; PrintLn(Obj.Prop());`, "11\n"},
		{"implicit routine receiver", `var Receivers := 0;
type TTest = class
 Reads: Integer;
 function GetProp: Integer;
 begin Inc(Reads); Result := Reads; end;
 property Prop: Integer read GetProp reintroduce;
end;
var Obj := new TTest;
function NextObj: TTest;
begin Inc(Receivers); Result := Obj; end;
PrintLn(NextObj.Prop()); PrintLn(Receivers); PrintLn(Obj.Reads);`, "1\n1\n1\n"},
		{"method control", `type TTest = class
 function Prop: Integer;
 begin Result := 8; end;
 function Plus(X: Integer): Integer;
 begin Result := X+1; end;
end;
var Obj := new TTest;
PrintLn(Obj.Prop()); PrintLn(Obj.Plus(3));`, "8\n4\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			_, err = engine.Eval(tt.source)
			if err != nil {
				t.Fatalf("execution failed: %v", err)
			}
			if got := output.String(); got != tt.want {
				t.Fatalf("output = %q, want %q", got, tt.want)
			}
		})
	}
}
