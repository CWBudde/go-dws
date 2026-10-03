package interp

import (
	"fmt"
	"testing"
)

func TestIndexedMemberCallable_CompiledReceiverOnce(t *testing.T) {
	for _, tc := range []struct {
		name, member, initialize, receiver, selection, want string
	}{
		{"factory", "Factories: array of TFactory;", "holder.Factories.SetLength(1); holder.Factories[0] := @Make;", "Receiver()", "var selected: TProc := %s.Factories[Slot()];", "receiver\nindex\nouter\nselected\ninner\n"},
		{"copy", "Factories: array of TFactory;", "holder.Factories.SetLength(1); holder.Factories[0] := @Make;", "Receiver()", "var copied: TFactory := %s.Factories[Slot()]; PrintLn('captured'); var selected: TProc := copied();", "receiver\nindex\ncaptured\nouter\nselected\ninner\n"},
		{"nested array", "Factories: array of array of TFactory;", "var row: array of TFactory; row.SetLength(1); row[0] := @Make; holder.Factories := [row];", "Receiver()", "var selected: TProc := %s.Factories[Slot()][Slot()];", "receiver\nindex\nindex\nouter\nselected\ninner\n"},
		{"implicit receiver", "Factories: array of TFactory;", "holder.Factories.SetLength(1); holder.Factories[0] := @Make;", "receiverPointer", "var selected: TProc := %s.Factories[Slot()];", "receiver\nindex\nouter\nselected\ninner\n"},
		{"array property", "Data: array of TFactory; function GetFactories: array of TFactory; begin PrintLn('getter'); Result := Data; end; property Factories: array of TFactory read GetFactories;", "holder.Data.SetLength(1); holder.Data[0] := @Make;", "Receiver()", "var selected: TProc := %s.Factories[Slot()];", "receiver\ngetter\nindex\nouter\nselected\ninner\n"},
		{"indexed property", "Data: array of TFactory; function GetFactory(i: Integer): TFactory; begin PrintLn('getter'); Result := Data[i]; end; property Factories[i: Integer]: TFactory read GetFactory;", "holder.Data.SetLength(1); holder.Data[0] := @Make;", "Receiver()", "var selected: TProc := %s.Factories[Slot()];", "receiver\nindex\ngetter\nouter\nselected\ninner\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := fmt.Sprintf(`
type TProc = procedure;
type TFactory = function: TProc;
type THolder = class %s end;
procedure Inner; begin PrintLn('inner'); end;
function Make: TProc; begin PrintLn('outer'); Result := @Inner; end;
var holder := THolder.Create;
%s
function Receiver: THolder; begin PrintLn('receiver'); Result := holder; end;
var receiverPointer: function: THolder := @Receiver;
function Slot: Integer; begin PrintLn('index'); Result := 0; end;
%s
PrintLn('selected'); selected();
`, tc.member, tc.initialize, fmt.Sprintf(tc.selection, tc.receiver))
			compileAndRunWithHelperTransfer(t, source, "indexed_member_callable.dws", tc.want)
		})
	}
}

func TestIndexedMemberCallable_CompiledExceptions(t *testing.T) {
	for _, tc := range []struct {
		name, receiver, index, factory, want string
	}{
		{"receiver", "raise Exception.Create('receiver failed');", "", "", "receiver\nreceiver failed\ninner\n"},
		{"index", "", "raise Exception.Create('index failed');", "", "receiver\nindex\nindex failed\ninner\n"},
		{"factory", "", "", "raise Exception.Create('factory failed');", "receiver\nindex\nouter\nfactory failed\ninner\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := fmt.Sprintf(`
type TProc = procedure;
type TFactory = function: TProc;
type THolder = class Factories: array of TFactory; end;
procedure Inner; begin PrintLn('inner'); end;
function Make: TProc; begin PrintLn('outer'); %s Result := @Inner; end;
var holder := THolder.Create;
holder.Factories.SetLength(1); holder.Factories[0] := @Make;
function Receiver: THolder; begin PrintLn('receiver'); %s Result := holder; end;
function Slot: Integer; begin PrintLn('index'); %s Result := 0; end;
var selected: TProc := @Inner;
try selected := Receiver().Factories[Slot()]; except on E: Exception do PrintLn(E.Message); end;
selected();
`, tc.factory, tc.receiver, tc.index)
			compileAndRunWithHelperTransfer(t, source, "indexed_member_exception.dws", tc.want)
		})
	}
}

func TestIndexedMemberCallable_CompiledDefaultAndInterfaceProperties(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type TProc = procedure;
type TFactory = function: TProc;
type TFactories = array of TFactory;
type IHolder = interface
 function GetFactories: TFactories;
 property Factories: TFactories read GetFactories;
end;
type TStore = class
 Data: array of TFactory;
 function GetFactory(i: Integer): TFactory;
 begin PrintLn('default getter'); Result := Data[i]; end;
 property Items[i: Integer]: TFactory read GetFactory; default;
end;
type THolder = class(TObject, IHolder)
 Store: TStore;
 function GetFactories: TFactories;
 begin PrintLn('interface getter'); Result := Store.Data; end;
 property Factories: TFactories read GetFactories;
end;
procedure Inner; begin PrintLn('inner'); end;
function Make: TProc; begin PrintLn('outer'); Result := @Inner; end;
var holder := THolder.Create;
holder.Store := TStore.Create;
holder.Store.Data.SetLength(1); holder.Store.Data[0] := @Make;
function Receiver: THolder; begin PrintLn('receiver'); Result := holder; end;
function InterfaceReceiver: IHolder; begin PrintLn('interface receiver'); Result := holder; end;
function Slot: Integer; begin PrintLn('index'); Result := 0; end;
var selected: TProc := Receiver().Store[Slot()];
PrintLn('selected'); selected();
selected := InterfaceReceiver().Factories[Slot()];
PrintLn('selected'); selected();
`, "indexed_default_interface.dws", "receiver\nindex\ndefault getter\nouter\nselected\ninner\ninterface receiver\ninterface getter\nindex\nouter\nselected\ninner\n")
}

func TestIndexedMemberCallable_RecordDefaultPropertyResult(t *testing.T) {
	for _, tc := range []struct {
		name, source, want string
	}{
		{"original accepted program", `type A = array of Integer;
type R = record
 class function Get(i,j: Integer): A;
 begin Result := [10*i+j]; end;
 property Items[i: Integer; j: Integer]: A read Get; default;
end;
type THolder = class Values := [R]; end;
var holder := THolder.Create;
PrintLn(holder.Values[0][1,2][0]);
`, "12\n"},
		{"receiver and bracket order", `
var trace := '';
type A = array of Integer;
type R = record
 class function Get(i,j: Integer): A;
 begin trace += 'g'; Result := [10*i+j]; end;
 property Items[i: Integer; j: Integer]: A read Get; default;
end;
type THolder = class Values := [R]; end;
var holder := THolder.Create;
function Receiver: THolder; begin trace += 'r'; Result := holder; end;
function Mark(labelText: String; value: Integer): Integer;
begin trace += labelText; Result := value; end;
PrintLn(Receiver().Values[Mark('a',0)][Mark('i',1),Mark('j',2)][Mark('z',0)]);
PrintLn(trace);
`, "12\nraijgz\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			compileAndRunWithHelperTransfer(t, tc.source, "record_default_result.dws", tc.want)
		})
	}
}

func TestIndexedMemberCallable_InterfaceDefaultPropertyResult(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
var trace := '';
type A = array of Integer;
type IValues = interface
 function Get(i,j: Integer): A;
 property Items[i: Integer; j: Integer]: A read Get; default;
end;
type TValues = class(TObject, IValues)
 function Get(i,j: Integer): A;
 begin trace += 'g'; Result := [10*i+j]; end;
 property Items[i: Integer; j: Integer]: A read Get; default;
end;
type THolder = class Value: IValues; end;
var holder := THolder.Create;
holder.Value := TValues.Create;
function Receiver: THolder; begin trace += 'r'; Result := holder; end;
function Mark(labelText: String; value: Integer): Integer;
begin trace += labelText; Result := value; end;
PrintLn(Receiver().Value[Mark('i',1),Mark('j',2)][Mark('z',0)]);
PrintLn(trace);
`, "interface_default_result.dws", "12\nrijgz\n")
}
