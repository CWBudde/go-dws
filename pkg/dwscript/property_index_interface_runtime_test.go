package dwscript

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// Losing original index storage or choosing an implementation property instead
// of the interface contract changes both accessor results and caller state.
func TestInterfacePropertyIndexModes_Runtime(t *testing.T) {
	const basic = `type IntAlias = Integer;
type I = interface
 function Get(var Access: IntAlias): Integer;
 procedure Put(var Store: IntAlias; V: Integer);
 property P[var Declared: IntAlias]: Integer read Get write Put; default;
end;
type T = class(TObject, I)
 function Get(var Actual: IntAlias): Integer; begin Actual += 1; Result := Actual; end;
 procedure Put(var Actual: IntAlias; V: Integer); begin Actual += V; end;
end;
var O: I := new T; var X := 5;
PrintLn(O.P[X]); PrintLn(X); O.P[X] := 3; PrintLn(X);`
	const capture = `var Trace := '';
type I = interface
 function Get(var X: Integer; const Y: Integer; Z: Integer): Integer;
 procedure Put(var X: Integer; const Y: Integer; Z: Integer; V: Integer);
 property P[var A: Integer; const B: Integer; C: Integer]: Integer read Get write Put; default;
end;
type T = class(TObject, I)
 F: Integer;
 function Get(var X: Integer; const Y: Integer; Z: Integer): Integer;
 begin Trace += 'G'; X += Y + Z; Result := X; end;
 procedure Put(var X: Integer; const Y: Integer; Z: Integer; V: Integer);
 begin Trace += 'S'; X += Y + Z + V; F += 1; end;
end;
var First := new T; var Second := new T; var O: I := First;
var A: array of Integer := [5]; var Saved := A;
function Rebind: Integer; begin Trace += 'J'; A := [100]; O := Second; Result := 2; end;
function RHS: Integer; begin Trace += 'V'; A := [200]; O := Second; Result := 3; end;
O.P[A[0], Rebind(), 1] := RHS();
PrintLn(Saved[0]); PrintLn(A[0]); PrintLn(First.F); PrintLn(Second.F); PrintLn(Trace);`
	cases := []struct{ name, source, want string }{
		{"named", basic, "6\n6\n9\n"},
		{"default", strings.ReplaceAll(basic, "O.P[", "O["), "6\n6\n9\n"},
		{"value", strings.ReplaceAll(basic, "var ", ""), "6\n5\n5\n"},
		{"capture", capture, "11\n200\n1\n0\nJVS\n"},
		{"defaultCapture", strings.ReplaceAll(capture, "O.P[", "O["), "11\n200\n1\n0\nJVS\n"},
		{"mixedRead", strings.ReplaceAll(capture, "O.P[A[0], Rebind(), 1] := RHS();", "PrintLn(O.P[A[0], Rebind(), 1]);"), "8\n8\n100\n0\n0\nJG\n"},
	}
	// Value parameters are allowed to mutate their local copy, but script variables
	// still require 'var' declarations.
	cases[2].source = strings.ReplaceAll(cases[2].source, "O: I :=", "var O: I :=")
	cases[2].source = strings.ReplaceAll(cases[2].source, "X := 5", "var X := 5")
	for _, tc := range cases {
		for _, checked := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/checked=%t", tc.name, checked), func(t *testing.T) {
				runPropertyIndexProgram(t, tc.source, tc.want, checked)
			})
		}
	}
	for _, target := range []string{"Receiver().P[", "Receiver()["} {
		t.Run(target, func(t *testing.T) {
			source := strings.ReplaceAll(capture, "O.P[A[0]", "function Receiver: I; begin Trace += 'R'; Result := O; end;\n"+target+"A[0]")
			runPropertyIndexProgram(t, source, "11\n200\n1\n0\nRJVS\n", true)
		})
	}

}

func runPropertyIndexProgram(t *testing.T, source, want string, checked bool) {
	t.Helper()
	var output bytes.Buffer
	engine, err := New(WithTypeCheck(checked), WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if program == nil {
		t.Fatal("compile returned nil Program")
	}
	result, err := engine.Run(program)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result == nil || !result.Success {
		t.Fatalf("run result: %v", result)
	}
	if got := output.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

// The lexical parent descriptor must supply var modes even when dynamic Self
// has a different property and a restarted accessor chain.
func TestResolvedPropertyIndexModes_Runtime(t *testing.T) {
	const source = `type B = class
 function Get(var Actual: Integer): Integer; virtual; begin Actual += 1; Result := Actual; end;
 property P[var Declared: Integer]: Integer read Get reintroduce;
end;
type M = class(B)
 function Get(var Actual: Integer): Integer; override; begin Actual += 10; Result := Actual; end;
end;
type C = class(M)
 property P[I: Integer]: Integer read (I + 90) reintroduce;
 function Get(var Actual: Integer): Integer; reintroduce; virtual; begin Actual += 100; Result := Actual; end;
 function Read(var X: Integer): Integer; begin Result := inherited P()[((X))]; end;
end;
type D = class(C)
 function Get(var Actual: Integer): Integer; override; begin Actual += 1000; Result := Actual; end;
end;
var O := new D; var X := 5; PrintLn(O.Read(X)); PrintLn(X);`
	for _, read := range []string{"inherited P()[((X))]", "inherited P[((X))]"} {
		t.Run(read, func(t *testing.T) {
			runPropertyIndexProgram(t, strings.ReplaceAll(source, "inherited P()[((X))]", read), "15\n15\n", true)
		})
	}
	const compat = `type T = class
 function Get(var Actual: Integer): Integer; begin Actual += 1; Result := Actual; end;
 property P[var Declared: Integer]: Integer read Get reintroduce;
 function Read(var X: Integer): Integer; begin Result := P()[((X))]; end;
end;
var O := new T; var X := 5; PrintLn(O.Read(X)); PrintLn(X); PrintLn(O.P()[X]); PrintLn(X);`
	runPropertyIndexProgram(t, compat, "6\n6\n7\n7\n", true)
}

// An interface's selected accessor must remain independent of a dynamic class
// property with the same name, and compound callers must reuse captured slots.
func TestInterfacePropertyIndexModes_Contracts(t *testing.T) {
	const prefix = `var Trace := '';
type I = interface
 function Get(var X: Integer; const Y: Integer): Integer;
 procedure Put(var X: Integer; const Y: Integer; V: Integer);
 property P[var A: Integer; const B: Integer]: Integer read Get write Put; default;
end;
type T = class(TObject, I)
 F: Integer;
 function Get(var X: Integer; const Y: Integer): Integer; begin Trace += 'G'; X += Y; Result := X; end;
 procedure Put(var X: Integer; const Y: Integer; V: Integer); begin Trace += 'S'; X := V; F := V; end;
 property P[A, B: Integer]: Integer read (A + B + 100) write (F := Value);
end;
var First := new T; var Second := new T; var O: I := First;
var A: array of Integer := [5]; var Saved := A;
function Slot: Integer; begin Trace += 'I'; Result := 0; end;
function Other: Integer; begin Trace += 'J'; Result := 1; end;
function RHS: Integer; begin Trace += 'V'; A := [200]; O := Second; Result := 3; end;
`
	cases := []struct{ name, tail, want string }{
		{"staticProperty", `PrintLn(O.P[A[0], 1]); PrintLn(A[0]); O.P[A[0], 1] := 9; PrintLn(A[0]); PrintLn(First.F); PrintLn(Trace);`, "6\n6\n9\n9\nGS\n"},
		{"compound", `O.P[A[Slot()], Other()] += RHS(); PrintLn(Saved[0]); PrintLn(A[0]); PrintLn(First.F); PrintLn(Second.F); PrintLn(Trace);`, "9\n200\n9\n0\nIJGVS\n"},
		{"defaultCompound", `O[A[Slot()], Other()] += RHS(); PrintLn(Saved[0]); PrintLn(A[0]); PrintLn(First.F); PrintLn(Second.F); PrintLn(Trace);`, "9\n200\n9\n0\nIJGVS\n"},
		{"forwarded", `procedure PassIndex(var X: Integer); begin var Local := 3; PrintLn(O.P[((X)), 1]); O.P[(Local), 0] := 9; PrintLn(Local); end; var X := 5; PassIndex(X); PrintLn(X); PrintLn(Trace);`, "6\n9\n6\nGS\n"},
		{"member", `First.F := 5; O.P[First.F, 0] := 9; PrintLn(First.F); PrintLn(Trace);`, "9\nS\n"},
		{"constValue", `var X := 5; PrintLn(O.P[X, 2 + 3]); PrintLn(X); PrintLn(Trace);`, "10\n10\nG\n"},
	}
	for _, tc := range cases {
		for _, checked := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/checked=%t", tc.name, checked), func(t *testing.T) { runPropertyIndexProgram(t, prefix+tc.tail, tc.want, checked) })
		}
	}
}

func TestInterfacePropertyIndexModes_Stops(t *testing.T) {
	const prefix = `var Trace := '';
type I = interface
 function Get(var X: Integer; const Y: Integer; Z: Integer): Integer;
 procedure Put(var X: Integer; const Y: Integer; Z: Integer; V: Integer);
 property P[var A: Integer; const B: Integer; C: Integer]: Integer read Get write Put;
end;
type T = class(TObject, I)
 function Get(var X: Integer; const Y: Integer; Z: Integer): Integer; begin Trace += 'G'; Result := X; end;
 procedure Put(var X: Integer; const Y: Integer; Z: Integer; V: Integer); begin Trace += 'S'; X := V; end;
end;
var O: I := new T; var A: array of Integer := [5];
function Receiver: I; begin Trace += 'R'; Result := O; end;
function Container: array of Integer; begin Trace += 'A'; Result := A; end;
function Slot: Integer; begin Trace += 'I'; raise Exception.Create('original var failure'); end;
function Fail: Integer; begin Trace += 'J'; raise Exception.Create('original const failure'); end;
function Later: Integer; begin Trace += 'K'; Result := 1; end;
function RHS: Integer; begin Trace += 'V'; Result := 3; end;
`
	cases := []struct {
		name, tail, want string
		checkedOnly      bool
	}{
		{"varException", `try O.P[Container()[Slot()], Later(), Later()] := RHS(); except on E: Exception do PrintLn(E.Message); end; PrintLn(Trace); PrintLn(A[0]);`, "original var failure\nAI\n5\n", false},
		{"constException", `try O.P[A[0], Fail(), Later()] := RHS(); except on E: Exception do PrintLn(E.Message); end; PrintLn(Trace); PrintLn(A[0]);`, "original const failure\nJ\n5\n", false},
		{"readException", `try PrintLn(O.P[A[0], Fail(), Later()]); except on E: Exception do PrintLn(E.Message); end; PrintLn(Trace); PrintLn(A[0]);`, "original const failure\nJ\n5\n", false},
		{"rhsException", `try O.P[A[0], 1, Later()] := Fail(); except on E: Exception do PrintLn(E.Message); end; PrintLn(Trace); PrintLn(A[0]);`, "original const failure\nKJ\n5\n", false},
		{"receiverException", `function BadReceiver: I; begin Trace += 'R'; raise Exception.Create('original receiver failure'); end; try BadReceiver().P[Container()[Slot()], Later(), Later()] := RHS(); except on E: Exception do PrintLn(E.Message); end; PrintLn(Trace);`, "original receiver failure\nR\n", true},
	}
	for _, tc := range cases {
		for _, checked := range []bool{true, false} {
			if !checked && tc.checkedOnly {
				continue
			}
			t.Run(fmt.Sprintf("%s/checked=%t", tc.name, checked), func(t *testing.T) { runPropertyIndexProgram(t, prefix+tc.tail, tc.want, checked) })
		}
	}
}

func TestInterfacePropertyIndexModes_Brackets(t *testing.T) {
	const prefix = `type Ints = array of Integer;
var Trace := '';
type I = interface
 function Get(var X: Integer; const Y: Integer): Ints;
 property P[var A: Integer; const B: Integer]: Ints read Get; default;
end;
type T = class(TObject, I)
 function Get(var X: Integer; const Y: Integer): Ints; begin Trace += 'G'; X += Y; Result := [X, 99]; end;
end;
var O: I := new T; var X := 5;
function Other: Integer; begin Trace += 'J'; Result := 1; end;
function ResultSlot: Integer; begin Trace += 'K'; Result := 0; end;
`
	for _, expr := range []string{"O.P[X, Other()][ResultSlot()]", "O[X, Other()][ResultSlot()]"} {
		for _, checked := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/checked=%t", expr, checked), func(t *testing.T) {
				runPropertyIndexProgram(t, prefix+"PrintLn("+expr+"); PrintLn(X); PrintLn(Trace);", "6\n6\nJGK\n", checked)
			})
		}
	}
	runPropertyIndexProgram(t, prefix+`var Receivers: array of I := [O]; function ReceiverSlot: Integer; begin Trace += 'R'; Result := 0; end; PrintLn(Receivers[ReceiverSlot()].P[X, Other()][ResultSlot()]); PrintLn(X); PrintLn(Trace);`, "6\n6\nRJGK\n", true)
}

func TestResolvedPropertyIndexModes_CaptureAndResult(t *testing.T) {
	const source = `type Ints = array of Integer;
var Trace := ''; var A: Ints := [5]; var Saved := A;
function Container: Ints; begin Trace += 'A'; Result := A; end;
function Slot: Integer; begin Trace += 'I'; Result := 0; end;
function Rebind: Integer; begin Trace += 'J'; A := [100]; Result := 1; end;
function ResultSlot: Integer; begin Trace += 'K'; Result := 0; end;
type B = class
 function Get(var Actual: Integer; const Other: Integer): Ints;
 begin Trace += 'G'; Actual += Other; Result := [Actual, 99]; end;
 property P[var Declared: Integer; const Extra: Integer]: Ints read Get reintroduce;
end;
type C = class(B)
 property P[A, B: Integer]: Ints read ([90, 99]) reintroduce;
 function Read: Integer; begin Result := inherited P()[Container()[Slot()], Rebind()][ResultSlot()]; end;
end;
PrintLn(C.Create.Read); PrintLn(Saved[0]); PrintLn(A[0]); PrintLn(Trace);`
	for _, read := range []string{"inherited P()", "inherited P"} {
		t.Run(read, func(t *testing.T) {
			runPropertyIndexProgram(t, strings.ReplaceAll(source, "inherited P()", read), "6\n6\n100\nAIJGK\n", true)
		})
	}
	const exception = `var Trace := '';
type B = class
 function Get(var Actual: Integer; const Other: Integer; Last: Integer): Integer;
 begin Trace += 'G'; Actual += Other; Result := Actual; end;
 property P[var Declared: Integer; const Extra: Integer; Tail: Integer]: Integer read Get reintroduce;
end;
function Fail: Integer; begin Trace += 'J'; raise Exception.Create('original resolved failure'); end;
function Later: Integer; begin Trace += 'K'; Result := 1; end;
type C = class(B)
 function Read(var X: Integer): Integer; begin Result := inherited P()[X, Fail(), Later()]; end;
end;
var O := new C; var X := 5;
try PrintLn(O.Read(X)); except on E: Exception do PrintLn(E.Message); end; PrintLn(X); PrintLn(Trace);`
	runPropertyIndexProgram(t, exception, "original resolved failure\n5\nJ\n", true)
}

func TestInterfacePropertyIndexModes_ErrorAndNil(t *testing.T) {
	const prefix = `type I = interface
 function Get(var X: Integer; const Y: Integer): Integer;
 procedure Put(var X: Integer; const Y: Integer; V: Integer);
 property P[var A: Integer; const B: Integer]: Integer read Get write Put;
end;
type T = class(TObject, I)
 function Get(var X: Integer; const Y: Integer): Integer; begin PrintLn('getter'); Result := X; end;
 procedure Put(var X: Integer; const Y: Integer; V: Integer); begin PrintLn('setter'); X := V; end;
end;
var O: I := new T; var X := 5; var A: array of Integer := [5];
function Slot: Integer; begin PrintLn('index'); Result := 0; end;
function Later: Integer; begin PrintLn('later'); Result := 1; end;
function RHS: Integer; begin PrintLn('rhs'); Result := 3; end;
`
	cases := []struct {
		name, tail, want, errorContains string
		checked                         bool
	}{
		{"invalidIndexStorage", `O.P[A[1], Later()] := RHS();`, "", "Upper bound exceeded! Index 1", true},
		{"indexValueError", `function Bad: Integer; begin PrintLn('bad'); Result := Missing; end; O.P[X, Bad()] := RHS();`, "bad\n", "undefined variable 'Missing'", false},
		{"receiverValueError", `function BadReceiver: I; begin PrintLn('receiver'); var Bad := 1 div 0; Result := O; end; BadReceiver().P[X, Later()] := RHS();`, "receiver\n", "Division by zero", true},
		{"nilRead", `O := nil; PrintLn(O.P[A[Slot()], Later()]);`, "index\nlater\n", "interface is nil", true},
		{"nilWrite", `O := nil; O.P[A[Slot()], Later()] := RHS();`, "index\nlater\nrhs\n", "setter cannot be executed on nil interface", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithTypeCheck(tc.checked), WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(prefix + tc.tail)
			if err != nil || program == nil {
				t.Fatalf("compile: %v, %v", program, err)
			}
			result, err := engine.Run(program)
			if err == nil || !strings.Contains(err.Error(), tc.errorContains) {
				t.Fatalf("result=%v error=%v; want %q", result, err, tc.errorContains)
			}
			if got := output.String(); got != tc.want {
				t.Fatalf("output=%q want=%q", got, tc.want)
			}
		})
	}
}
