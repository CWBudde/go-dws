package interp

import (
	"fmt"
	"testing"
)

func TestAnonymousCallable_CompiledFactories(t *testing.T) {
	for _, named := range []bool{false, true} {
		proc, factory, prefix := "procedure", "function: procedure", ""
		if named {
			prefix = "type TProc = procedure; type TFactory = function: TProc;"
			proc, factory = "TProc", "TFactory"
		}
		for _, supplier := range []string{"Make", "factory"} {
			t.Run(fmt.Sprintf("named=%v/%s", named, supplier), func(t *testing.T) {
				source := fmt.Sprintf(`%s
var outerCalls, innerCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Make: %s; begin Inc(outerCalls); Result := @Inner; end;
var factory: %s := @Make;
var target: %s;
target := %s;
PrintLn(outerCalls); PrintLn(innerCalls);
var initialized: %s := %s;
PrintLn(outerCalls); PrintLn(innerCalls);
var explicitResult := %s();
PrintLn(outerCalls); PrintLn(innerCalls);
var copied: %s := initialized;
target(); initialized(); explicitResult(); copied();
PrintLn(outerCalls); PrintLn(innerCalls);
var inferred := Make;
PrintLn(outerCalls); PrintLn(innerCalls);
inferred(); PrintLn(innerCalls);
`, prefix, proc, factory, proc, supplier, proc, supplier, supplier, proc)
				compileAndRunWithHelperTransfer(t, source, "anonymous_factory.dws", "1\n0\n2\n0\n3\n0\n3\n4\n4\n4\n5\n")
			})
		}
	}
}

func TestAnonymousCallable_CompiledInitializerReads(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
var outerCalls, innerCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Make: procedure; begin Inc(outerCalls); Result := @Inner; end;
var factory: function: procedure := @Make;
var inferred := factory;
PrintLn(outerCalls); PrintLn(innerCalls);
inferred(); PrintLn(innerCalls);
function Scalar: Integer; begin Inc(outerCalls); Result := 7; end;
var scalarFactory: function: Integer := @Scalar;
var typed: Integer := scalarFactory;
var json: JSONVariant := scalarFactory;
PrintLn(typed); PrintLn(json); PrintLn(outerCalls);
`, "initializer_reads.dws", "1\n0\n1\n7\n7\n3\n")
}

func TestAnonymousCallable_CompiledSupplierReads(t *testing.T) {
	for _, supplier := range []string{"holder.Factory", "factories[0]"} {
		t.Run(supplier, func(t *testing.T) {
			source := fmt.Sprintf(`
type TFactory = function: procedure;
type THolder = class Factory: TFactory; end;
var outerCalls, innerCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Make: procedure; begin Inc(outerCalls); Result := @Inner; end;
var holder := THolder.Create; holder.Factory := @Make;
var factories: array[0..0] of TFactory; factories[0] := @Make;
var target: procedure;
target := %[1]s;
PrintLn(outerCalls); PrintLn(innerCalls);
var initialized: procedure := %[1]s;
PrintLn(outerCalls); PrintLn(innerCalls);
var copied: TFactory := %[1]s;
PrintLn(outerCalls); PrintLn(innerCalls);
target(); initialized(); copied()();
PrintLn(outerCalls); PrintLn(innerCalls);
`, supplier)
			compileAndRunWithHelperTransfer(t, source, "supplier_reads.dws", "1\n0\n2\n0\n2\n0\n3\n3\n")
		})
	}
}

func TestAnonymousCallable_CompiledNestedReturns(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
var outerCalls, middleCalls, innerCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Middle: procedure; begin Inc(middleCalls); Result := @Inner; end;
function Outer: function: procedure; begin Inc(outerCalls); Result := @Middle; end;
var outerFactory: function: function: procedure := @Outer;
var middleFactory: function: procedure := outerFactory;
PrintLn(outerCalls); PrintLn(middleCalls); PrintLn(innerCalls);
var innerFactory: procedure := middleFactory;
PrintLn(outerCalls); PrintLn(middleCalls); PrintLn(innerCalls);
innerFactory(); PrintLn(innerCalls);
`, "nested_returns.dws", "1\n0\n0\n1\n1\n0\n1\n")
}

func TestAnonymousCallable_CompiledWrappedFactories(t *testing.T) {
	for _, modifier := range []string{"var", "const", "lazy"} {
		t.Run(modifier, func(t *testing.T) {
			argument, want := "factory", "0\n2\n0\n2\n"
			if modifier == "lazy" {
				argument, want = "GetFactory()", "2\n2\n0\n2\n"
			}
			source := fmt.Sprintf(`
var forces, outerCalls, innerCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Make: procedure; begin Inc(outerCalls); Result := @Inner; end;
function GetFactory: function: procedure; begin Inc(forces); Result := @Make; end;
procedure Apply(%s source: function: procedure);
begin
 var first: procedure := source;
 var second: procedure := source;
 PrintLn(forces); PrintLn(outerCalls); PrintLn(innerCalls);
 first(); second(); PrintLn(innerCalls);
end;
var factory: function: procedure := @Make;
Apply(%s);
`, modifier, argument)
			compileAndRunWithHelperTransfer(t, source, "wrapped_factories.dws", want)
		})
	}
}

func TestAnonymousCallable_CompiledSupplierException(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type TFactory = function: procedure;
type THolder = class Factory: TFactory; end;
var outerCalls, innerCalls, receiverCalls, indexCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Make: procedure; begin Inc(outerCalls); raise Exception.Create('factory failed'); end;
var holder := THolder.Create; holder.Factory := @Make;
var factories: array[0..0] of TFactory; factories[0] := @Make;
function Receiver: THolder; begin Inc(receiverCalls); Result := holder; end;
function Slot: Integer; begin Inc(indexCalls); Result := 0; end;
var target: procedure := @Inner;
try target := Receiver().Factory; except on E: Exception do PrintLn(E.Message); end;
try target := factories[Slot()]; except on E: Exception do PrintLn(E.Message); end;
PrintLn(outerCalls); PrintLn(innerCalls); PrintLn(receiverCalls); PrintLn(indexCalls);
target(); PrintLn(innerCalls);
`, "supplier_exception.dws", "factory failed\nfactory failed\n2\n0\n1\n1\n1\n")
}

func TestAnonymousCallable_CompiledMethodResult(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
var outerCalls, innerCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
type THolder = class
 function Make: procedure;
 begin Inc(outerCalls); Result := @Inner; end;
end;
var holder := THolder.Create;
var inferred := holder.Make;
PrintLn(outerCalls); PrintLn(innerCalls);
var typed: procedure := holder.Make;
PrintLn(outerCalls); PrintLn(innerCalls);
typed := holder.Make;
PrintLn(outerCalls); PrintLn(innerCalls);
inferred(); typed(); PrintLn(innerCalls);
`, "method_result.dws", "1\n0\n2\n0\n3\n0\n2\n")
}

func TestAnonymousCallable_CompiledRecordStaticStorage(t *testing.T) {
	for _, supplier := range []string{"R.Stored", "R.Factory", "item.Stored", "item.Factory", "meta.Stored", "meta.Factory"} {
		t.Run(supplier, func(t *testing.T) {
			source := fmt.Sprintf(`
type TProc = procedure;
type TFactory = function: TProc;
type R = record
 class var Stored: TFactory;
 property Factory: TFactory read Stored write Stored;
end;
var outerCalls, innerCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Make: TProc; begin Inc(outerCalls); Result := @Inner; end;
R.Stored := @Make;
var item: R; var meta := R;
var copied: TFactory := %[1]s;
PrintLn(outerCalls); PrintLn(innerCalls);
var selected: TProc := %[1]s;
PrintLn(outerCalls); PrintLn(innerCalls);
selected := %[1]s;
PrintLn(outerCalls); PrintLn(innerCalls);
selected();
var fromCopy: TProc := copied(); fromCopy();
PrintLn(outerCalls); PrintLn(innerCalls);
`, supplier)
			compileAndRunWithHelperTransfer(t, source, "record_static_callable.dws", "0\n0\n1\n0\n2\n0\n3\n2\n")
		})
	}
}

func TestAnonymousCallable_CompiledPointerReceiverStorage(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type TProc = procedure;
type TFactory = function: TProc;
type THolder = class Factory: TFactory; end;
var receiverCalls, outerCalls, innerCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Make: TProc; begin Inc(outerCalls); Result := @Inner; end;
var holder := THolder.Create; holder.Factory := @Make;
function GetHolder: THolder; begin Inc(receiverCalls); Result := holder; end;
var receiver: function: THolder := @GetHolder;
var copied: TFactory := receiver.Factory;
PrintLn(receiverCalls); PrintLn(outerCalls); PrintLn(innerCalls);
var selected: TProc := receiver.Factory;
PrintLn(receiverCalls); PrintLn(outerCalls); PrintLn(innerCalls);
selected := receiver.Factory;
PrintLn(receiverCalls); PrintLn(outerCalls); PrintLn(innerCalls);
selected();
var fromCopy: TProc := copied(); fromCopy();
PrintLn(receiverCalls); PrintLn(outerCalls); PrintLn(innerCalls);
`, "pointer_receiver_callable.dws", "1\n0\n0\n2\n1\n0\n3\n2\n0\n3\n3\n2\n")
}
