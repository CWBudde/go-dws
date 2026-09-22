package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
)

func TestInterfaceProperties_Fixtures(t *testing.T) {
	for _, category := range []string{"InterfacesPass", "InterfacesFail"} {
		t.Run(category, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, category, "interface_properties.pas")
			if got, detail := runFixtureTest(path, category == "InterfacesFail", fixtureconfig.HintsLevel(category)); got != testResultPassed {
				t.Fatalf("%v: %s", got, detail)
			}
		})
	}
}

func TestInterfaceProperties_InheritedMultiIndex(t *testing.T) {
	const source = `type IBase = interface
 function GetItem(x, y: Integer): Integer;
 procedure SetItem(x, y, v: Integer);
 property Items[x, y: Integer]: Integer read GetItem write SetItem; default;
 end;
 type IChild = interface(IBase) end;
 type TImpl = class(TObject, IChild)
 value: Integer;
 function GetItem(x, y: Integer): Integer; begin Result := value+x+y; end;
 procedure SetItem(x, y, v: Integer); begin value := v-x-y; end;
 end;
 var i: IChild := TImpl.Create;
 i.Items[2, 3] := 10;
 PrintLn(i.items[2, 3]);
 i[4, 5] := 20;
 PrintLn(i[4, 5]);`
	compileAndRunWithHelperTransfer(t, source, "interface_multi.dws", "10\n20\n")
}

func TestInterfaceProperties_ResultIndex(t *testing.T) {
	const source = `type TInts = array of Integer;
 type IItems = interface
 function GetItems(x, y: Integer): TInts;
 property Items[x, y: Integer]: TInts read GetItems; default;
 end;
 type TImpl = class(TObject, IItems)
 data: TInts;
 function GetItems(x, y: Integer): TInts; begin Result := data; end;
 end;
 var obj := TImpl.Create;
 obj.data := [10, 20, 30];
 var i: IItems := obj;
 PrintLn(i.Items[1, 2][1]);
 PrintLn(i[1, 2][2]);
 i.Items[1, 2][1] := 50;
 i[1, 2][2] := 60;
 PrintLn(i.Items[1, 2][1]);
 PrintLn(i[1, 2][2]);`
	compileAndRunWithHelperTransfer(t, source, "interface_result_index.dws", "20\n30\n50\n60\n")
}

func TestInterfaceProperties_EvaluateOnce(t *testing.T) {
	const source = `type IItems = interface
 function GetItem(x, y: Integer): Integer;
 procedure SetItem(x, y, value: Integer);
 property Items[x, y: Integer]: Integer read GetItem write SetItem; default;
 end;
 type TImpl = class(TObject, IItems)
 value: Integer;
 function GetItem(x, y: Integer): Integer; begin Result := value; end;
 procedure SetItem(x, y, v: Integer); begin value := v; end;
 end;
 var i: IItems := TImpl.Create;
 var receivers := 0;
 var indices := 0;
 function Receiver: IItems; begin Inc(receivers); Result := i; end;
 function GetIndex: Integer; begin Inc(indices); Result := 1; end;
 Receiver().Items[GetIndex(), GetIndex()] := 10;
 Receiver()[GetIndex(), GetIndex()] += 5;
 Receiver().Items[GetIndex(), GetIndex()] += 7;
 PrintLn(Receiver()[GetIndex(), GetIndex()]);
 PrintLn(receivers);
 PrintLn(indices);`
	compileAndRunWithHelperTransfer(t, source, "interface_once.dws", "22\n4\n8\n")
}

func TestInterfaceProperties_ResultCompoundEvaluateOnce(t *testing.T) {
	for _, target := range []string{
		"Receiver().Items[GetIndex(), GetIndex()][GetIndex()]",
		"Receiver()[GetIndex(), GetIndex()][GetIndex()]",
		"Receiver.Items[GetIndex(), GetIndex()][GetIndex()]",
		"Receiver[GetIndex(), GetIndex()][GetIndex()]",
	} {
		t.Run(target, func(t *testing.T) {
			const prefix = `type TInts = array of Integer;
 type IItems = interface
 function GetItems(x, y: Integer): TInts;
 property Items[x, y: Integer]: TInts read GetItems; default;
 end;
 var getters := 0;
 type TImpl = class(TObject, IItems)
 data: TInts;
 function GetItems(x, y: Integer): TInts;
 begin Inc(getters); Result := data; end;
 end;
 var first := TImpl.Create;
 first.data := [10];
 var second := TImpl.Create;
 second.data := [100];
 var receivers := 0;
 var indices := 0;
 var rightSides := 0;
 function Receiver: IItems;
 begin
 Inc(receivers);
 if receivers = 1 then Result := first else Result := second;
 end;
 function GetIndex: Integer; begin Inc(indices); Result := 0; end;
 function RightSide: Integer; begin Inc(rightSides); Result := 5; end;
 `
			source := prefix + target + ` += RightSide();
 PrintLn(first.data[0]); PrintLn(second.data[0]);
 PrintLn(receivers); PrintLn(indices); PrintLn(getters); PrintLn(rightSides);`
			compileAndRunWithHelperTransfer(t, source, "interface_result_compound_once.dws", "15\n100\n1\n3\n1\n1\n")
		})
	}
}

func TestInterfaceProperties_ImplicitReceiver(t *testing.T) {
	const source = `type IItems = interface
 function GetItem(x, y: Integer): Integer;
 procedure SetItem(x, y, v: Integer);
 property Items[x, y: Integer]: Integer read GetItem write SetItem; default;
 end;
 type TImpl = class(TObject, IItems)
 value: Integer;
 function GetItem(x, y: Integer): Integer; begin Result := value+x+y; end;
 procedure SetItem(x, y, v: Integer); begin value := v-x-y; end;
 end;
 var i: IItems := TImpl.Create;
 var receivers := 0;
 function Receiver: IItems; begin Inc(receivers); Result := i; end;
 Receiver.Items[1, 2] := 10;
 Receiver[1, 2] += 5;
 PrintLn(Receiver.Items[1, 2]);
 PrintLn(Receiver[1, 2]);
 PrintLn(receivers);`
	compileAndRunWithHelperTransfer(t, source, "interface_implicit_receiver.dws", "15\n15\n4\n")
}

func TestInterfaceProperties_DefaultThroughMember(t *testing.T) {
	const source = `type IItems = interface
 function GetItem(x: Integer): Integer;
 procedure SetItem(x, value: Integer);
 property Items[x: Integer]: Integer read GetItem write SetItem; default;
 end;
 type TImpl = class(TObject, IItems)
 value: Integer;
 function GetItem(x: Integer): Integer; begin Result := value+x; end;
 procedure SetItem(x, v: Integer); begin value := v-x; end;
 end;
 type THolder = class
 item: IItems;
 property Intf: IItems read item;
 end;
 var h := THolder.Create;
 h.item := TImpl.Create;
 h.item[1] := 20;
 PrintLn(h.Intf[1]);
 h.Intf[1] := 30;
 PrintLn(h.item[1]);`
	compileAndRunWithHelperTransfer(t, source, "interface_member_default.dws", "20\n30\n")
}

func TestInterfaceProperties_ChildDefaultOverridesParent(t *testing.T) {
	const source = `type IBase = interface
 function GetBase(x: Integer): Integer;
 property BaseItem[x: Integer]: Integer read GetBase; default;
 end;
 type IChild = interface(IBase)
 function GetChild(x: Integer): Integer;
 property ChildItem[x: Integer]: Integer read GetChild; default;
 end;
 type TImpl = class(TObject, IChild)
 function GetBase(x: Integer): Integer; begin Result := 10+x; end;
 function GetChild(x: Integer): Integer; begin Result := 20+x; end;
 end;
 var child: IChild := TImpl.Create;
 var base: IBase := child;
 PrintLn(child[1]);
 PrintLn(base[1]);
 PrintLn(child.BaseItem[1]);`
	compileAndRunWithHelperTransfer(t, source, "interface_child_default.dws", "21\n11\n11\n")
}

func TestInterfaceProperties_CompoundArrayAppend(t *testing.T) {
	const source = `type TInts = array of Integer;
 type IItems = interface
 function GetValues: TInts;
 procedure SetValues(v: TInts);
 function GetRow(x: Integer): TInts;
 procedure SetRow(x: Integer; v: TInts);
 property Values: TInts read GetValues write SetValues;
 property Rows[x: Integer]: TInts read GetRow write SetRow;
 end;
 type TImpl = class(TObject, IItems)
 data: TInts;
 function GetValues: TInts; begin Result := data; end;
 procedure SetValues(v: TInts); begin data := v; end;
 function GetRow(x: Integer): TInts; begin Result := data; end;
 procedure SetRow(x: Integer; v: TInts); begin data := v; end;
 end;
 var i: IItems := TImpl.Create;
 i.Values += 1;
 i.Values += 2;
 i.Rows[0] += 3;
 PrintLn(i.Values.Length);
 PrintLn(i.Values[2]);`
	compileAndRunWithHelperTransfer(t, source, "interface_compound_append.dws", "3\n3\n")
}
