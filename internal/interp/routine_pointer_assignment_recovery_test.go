package interp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoutinePointerAssignment_ReturnedCallable(t *testing.T) {
	for _, rhs := range []string{"Make", "factory"} {
		t.Run(rhs, func(t *testing.T) {
			source := fmt.Sprintf(`
type TProc = procedure;
type TFactory = function: TProc;
var outerCalls, innerCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Make: TProc; begin Inc(outerCalls); Result := Inner; end;
var factory: TFactory := Make;
var target: TProc;
PrintLn(outerCalls); PrintLn(innerCalls);
target := %[1]s;
PrintLn(outerCalls); PrintLn(innerCalls);
var initialized: TProc := %[1]s;
PrintLn(outerCalls); PrintLn(innerCalls);
var explicitResult: TProc := %[1]s();
PrintLn(outerCalls); PrintLn(innerCalls);
var copied: TProc := initialized;
target(); initialized(); explicitResult(); copied();
PrintLn(outerCalls); PrintLn(innerCalls);
`, rhs)
			assertOutput(t, runQuickwinScript(t, source), "0\n0\n1\n0\n2\n0\n3\n0\n3\n4\n")
		})
	}
}

func TestRoutinePointerAssignment_References(t *testing.T) {
	source := `
type TProc = procedure;
type TParam = procedure(x: Integer);
var calls: Integer;
procedure Work; begin Inc(calls); end;
procedure Parameterized(x: Integer); begin calls += x; end;
var p: TProc := Work;
var copy: TProc;
copy := p;
var q: TParam := Parameterized;
q := Parameterized;
var addressed: TProc := @Work;
PrintLn(calls);
p(); copy(); addressed(); q(7);
PrintLn(calls);
`
	assertOutput(t, runQuickwinScript(t, source), "0\n10\n")
}

func TestRoutinePointerAssignment_FactoryExceptionPreservesTarget(t *testing.T) {
	for _, rhs := range []string{"Make", "factory"} {
		t.Run(rhs, func(t *testing.T) {
			source := fmt.Sprintf(`
type TProc = procedure;
type TFactory = function: TProc;
var outerCalls, innerCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Make: TProc;
begin Inc(outerCalls); raise Exception.Create('factory failed'); end;
var factory: TFactory := Make;
var target: TProc := Inner;
try target := %s; except on E: Exception do PrintLn(E.Message); end;
PrintLn(outerCalls); PrintLn(innerCalls);
target(); PrintLn(innerCalls);
`, rhs)
			assertOutput(t, runQuickwinScript(t, source), "factory failed\n1\n0\n1\n")
		})
	}
}

func TestRoutinePointerAssignment_SimpleFixture(t *testing.T) {
	base := filepath.Join("..", "..", "testdata", "fixtures", "SimpleScripts", "func_ptr1")
	source, err := os.ReadFile(base + ".pas")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(base + ".txt")
	if err != nil {
		t.Fatal(err)
	}
	assertOutput(t, runQuickwinScript(t, string(source)), strings.TrimRight(string(want), "\r\n")+"\n")
}

func TestRoutinePointerAssignment_LocalReferencesAndFactories(t *testing.T) {
	for _, rhs := range []string{"Make", "factory"} {
		t.Run(rhs, func(t *testing.T) {
			source := fmt.Sprintf(`
type TProc = procedure;
type TFactory = function: TProc;
var outerCalls, innerCalls: Integer;
procedure Run;
begin
 procedure Inner; begin Inc(innerCalls); end;
 function Make: TProc; begin Inc(outerCalls); Result := Inner; end;
  var factory: TFactory := Make;
  var target: TProc := Inner;
  PrintLn(outerCalls); PrintLn(innerCalls);
  target := %s;
  PrintLn(outerCalls); PrintLn(innerCalls);
  target(); PrintLn(innerCalls);
 end;
Run;
`, rhs)
			assertOutput(t, runQuickwinScript(t, source), "0\n0\n1\n0\n1\n")
		})
	}
}

func TestRoutinePointerAssignment_BoundMethodReference(t *testing.T) {
	source := `
type TProc = procedure(x: Integer) of object;
type TCounter = class
 Value: Integer;
 procedure Add(x: Integer); begin Value += x; end;
end;
var first := TCounter.Create;
var second := TCounter.Create;
var callback: TProc := first.Add;
var copied: TProc;
copied := callback;
callback := second.Add;
PrintLn(first.Value); PrintLn(second.Value);
copied(7); callback(3);
PrintLn(first.Value); PrintLn(second.Value);
`
	assertOutput(t, runQuickwinScript(t, source), "0\n0\n7\n3\n")
}

func TestRoutinePointerAssignment_WrappedFactoryParameters(t *testing.T) {
	for _, tt := range []struct{ modifier, argument, want string }{
		{"var", "factory", "0\n1\n0\n0\n2\n0\n2\n"},
		{"lazy", "GetFactory()", "1\n1\n0\n2\n2\n0\n2\n"},
	} {
		t.Run(tt.modifier, func(t *testing.T) {
			source := fmt.Sprintf(`
type TProc = procedure;
type TFactory = function: TProc;
var forces, outerCalls, innerCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Make: TProc; begin Inc(outerCalls); Result := Inner; end;
function GetFactory: TFactory; begin Inc(forces); Result := Make; end;
procedure Apply(%s source: TFactory);
begin
 var target: TProc;
 target := source;
 PrintLn(forces); PrintLn(outerCalls); PrintLn(innerCalls);
 var initialized: TProc := source;
 PrintLn(forces); PrintLn(outerCalls); PrintLn(innerCalls);
 target(); initialized(); PrintLn(innerCalls);
end;
var factory: TFactory := Make;
Apply(%s);
`, tt.modifier, tt.argument)
			assertOutput(t, runQuickwinScript(t, source), tt.want)
		})
	}
}

func TestRoutinePointerAssignment_LazyFactoryException(t *testing.T) {
	for _, entry := range []string{"Apply", "Relay"} {
		for _, tt := range []struct{ name, makeBody, getBody, want string }{
			{"forcing raises", "Result := Inner;", "Inc(forces); raise Exception.Create('force failed');", "force failed\n1\n0\n0\n1\n"},
			{"factory raises", "raise Exception.Create('factory failed');", "Inc(forces); Result := Make;", "factory failed\n1\n1\n0\n1\n"},
		} {
			t.Run(entry+"/"+tt.name, func(t *testing.T) {
				source := fmt.Sprintf(`
type TProc = procedure;
type TFactory = function: TProc;
var forces, outerCalls, innerCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Make: TProc; begin Inc(outerCalls); %s end;
function GetFactory: TFactory; begin %s end;
var target: TProc := Inner;
procedure Apply(lazy source: TFactory);
begin
 try target := source; except on E: Exception do PrintLn(E.Message); end;
 PrintLn(forces); PrintLn(outerCalls); PrintLn(innerCalls);
 target(); PrintLn(innerCalls);
end;
procedure Relay(lazy value: TFactory); begin Apply(value); end;
%s(GetFactory());
`, tt.makeBody, tt.getBody, entry)
				assertOutput(t, runQuickwinScript(t, source), tt.want)
			})
		}
	}

}

func TestRoutinePointerAssignment_CoalescedLazyFactory(t *testing.T) {
	for _, forwarded := range []string{"value ?? nil", "value ?? Fallback()"} {
		for _, tt := range []struct{ name, makeBody, getBody, want string }{
			{"supplier raises", "Result := Inner;", "Inc(forces); raise Exception.Create('force failed');", "force failed\n1\n0\n0\n0\n1\n"},
			{"factory raises", "raise Exception.Create('factory failed');", "Inc(forces); Result := Make;", "factory failed\n1\n1\n0\n0\n1\n"},
			{"success", "Result := Inner;", "Inc(forces); Result := Make;", "1\n1\n0\n0\n1\n"},
		} {
			t.Run(forwarded+"/"+tt.name, func(t *testing.T) {
				source := fmt.Sprintf(`
type TProc = procedure;
type TFactory = function: TProc;
var forces, outerCalls, innerCalls, fallbackCalls: Integer;
procedure Inner; begin Inc(innerCalls); end;
function Make: TProc; begin Inc(outerCalls); %s end;
function GetFactory: TFactory; begin %s end;
function Fallback: TFactory; begin Inc(fallbackCalls); Result := Make; end;
var target: TProc := Inner;
procedure Apply(lazy source: TFactory);
begin
 try target := source; except on E: Exception do PrintLn(E.Message); end;
 PrintLn(forces); PrintLn(outerCalls); PrintLn(innerCalls); PrintLn(fallbackCalls);
 target(); PrintLn(innerCalls);
end;
procedure Relay(lazy value: TFactory); begin Apply(%s); end;
Relay(GetFactory());
`, tt.makeBody, tt.getBody, forwarded)
				assertOutput(t, runQuickwinScript(t, source), tt.want)
			})
		}
	}
}

func TestRoutinePointerAssignment_OrdinaryLazyExceptions(t *testing.T) {
	for _, tt := range []struct{ name, entry, supplier, fallback, want string }{
		{"direct supplier raises", "Capture", "raise Exception.Create('number failed');", "Result := 9;", "number failed\n99\n1\n0\n"},
		{"coalesced supplier raises", "Relay", "raise Exception.Create('number failed');", "Result := 9;", "number failed\n99\n1\n0\n"},
		{"coalesced fallback raises", "Relay", "Result := 0;", "raise Exception.Create('fallback failed');", "fallback failed\n99\n1\n1\n"},
		{"coalesced success", "Relay", "Result := 7;", "Result := 9;", "7\n1\n0\n"},
		{"coalesced fallback success", "Relay", "Result := 0;", "Result := 9;", "9\n1\n1\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := fmt.Sprintf(`
var forces, fallbackCalls: Integer;
function Number: Integer; begin Inc(forces); %s end;
function Fallback: Integer; begin Inc(fallbackCalls); %s end;
procedure Capture(lazy value: Integer);
begin
 var received: Integer := 99;
 try received := value; except on E: Exception do PrintLn(E.Message); end;
 PrintLn(received); PrintLn(forces); PrintLn(fallbackCalls);
end;
procedure Relay(lazy value: Integer); begin Capture(value ?? Fallback()); end;
%s(Number());
`, tt.supplier, tt.fallback, tt.entry)
			assertOutput(t, runQuickwinScript(t, source), tt.want)
		})
	}
}
