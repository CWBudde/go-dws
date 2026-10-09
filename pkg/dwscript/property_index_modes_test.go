package dwscript

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestPropertyIndexModes_Runtime(t *testing.T) {
	varScript := `type T = class
 function Get(var I: Integer): Integer;
 begin I += 1; Result := I; end;
 procedure SetIt(var I: Integer; V: Integer);
 begin I += V; end;
 property P[var I: Integer]: Integer read Get write SetIt;
end;
var O := new T;
var X := 5;
PrintLn(O.P[X]);
PrintLn(X);
O.P[X] := 3;
PrintLn(X);`
	cases := []struct{ name, source, want string }{
		{"var", varScript, "6\n6\n9\n"},
		{"value", strings.ReplaceAll(varScript, "var I", "I"), "6\n5\n5\n"},
		{"default", strings.ReplaceAll(strings.ReplaceAll(varScript, "write SetIt;", "write SetIt; default;"), "O.P[", "O["), "6\n6\n9\n"},
		{"const", `type T = class
 Field: Integer;
 function Get(const I: Integer): Integer;
 begin Result := I; end;
 procedure SetIt(const I: Integer; V: Integer);
 begin Field := I + V; end;
 property P[const I: Integer]: Integer read Get write SetIt;
end;
var O := new T;
PrintLn(O.P[2 + 3]);
O.P[5] := 3;
PrintLn(O.Field);`, "5\n8\n"},
		{"setterOrder", `var Trace := '';
type T = class
 function Get(var I: Integer): Integer; begin Result := I; end;
 procedure SetIt(var I: Integer; V: Integer);
 begin Trace += 'S'; I += V; end;
 property P[var I: Integer]: Integer read Get write SetIt;
end;
var O := new T;
var A: array of Integer := [5];
function Receiver: T; begin Trace += 'R'; Result := O; end;
function Container: array of Integer; begin Trace += 'A'; Result := A; end;
function Slot: Integer; begin Trace += 'I'; Result := 0; end;
function RHS: Integer; begin Trace += 'V'; Result := 3; end;
Receiver().P[Container()[Slot()]] := RHS();
PrintLn(Trace); PrintLn(A[0]);`, "RAIVS\n8\n"},
	}
	cases = append(cases, []struct{ name, source, want string }{
		{"forwardedVar", `type T = class
 function Get(var I: Integer): Integer; begin I += 1; Result := I; end;
 procedure SetIt(var I: Integer; V: Integer); begin I += V; end;
 property P[var I: Integer]: Integer read Get write SetIt;
 property Q[var I: Integer]: Integer read P write P;
end;
var O := new T;
procedure Use(var Y: Integer);
begin var Z := 1; PrintLn(O.Q[Y]); O.Q[Y] := 3; PrintLn(O.P[Z]); PrintLn(Z); end;
var X := 5; Use(X); PrintLn(X);`, "6\n2\n2\n9\n"},
		{"mixedCapture", `type T = class
 function Get(var I: Integer; const J: Integer; K: Integer): Integer;
 begin I += J + K; Result := I; end;
 procedure SetIt(var I: Integer; const J: Integer; K: Integer; V: Integer);
 begin I += J + K + V; end;
 property P[var I: Integer; const J: Integer; K: Integer]: Integer read Get write SetIt;
end;
var O := new T;
var A: array of Integer := [5]; var Saved := A;
function Rebind: Integer; begin A := [100]; Result := 2; end;
PrintLn(O.P[A[0], Rebind(), 1]); PrintLn(Saved[0]); PrintLn(A[0]);
A := Saved; O.P[A[0], 0, 0] := Rebind();
PrintLn(Saved[0]); PrintLn(A[0]);`, "8\n8\n100\n10\n100\n"},
		{"fieldCapture", `var Trace := '';
type T = class
 I: Integer;
 function Get(var X: Integer): Integer; begin X += 1; Result := X; end;
 property P[var X: Integer]: Integer read Get;
end;
var O := new T; O.I := 5;
function Receiver: T; begin Trace += 'R'; Result := O; end;
PrintLn(O.P[Receiver().I]); PrintLn(O.I); PrintLn(Trace);`, "6\n6\nR\n"},
		{"classAccessors", `type T = class
 class function Get(var I: Integer): Integer; begin I += 1; Result := I; end;
 class procedure SetIt(var I: Integer; V: Integer); begin I += V; end;
 class property P[var I: Integer]: Integer read Get write SetIt;
end;
var O := new T; var X := 5;
PrintLn(O.P[X]); PrintLn(T.P[X]); T.P[X] := 3; PrintLn(X); O.P[X] := 2; PrintLn(X);`, "6\n7\n10\n12\n"},
		{"expressionAccessor", `type T = class
 property P[var I: Integer]: Integer read (I + 1) write (I := Value);
end;
var O := new T; var X := 5;
PrintLn(O.P[X]); O.P[X] := 9; PrintLn(X);`, "6\n9\n"},
		{"failedCapture", `var Trace := '';
type T = class
 function Get(var I: Integer): Integer; begin Result := I; end;
 procedure SetIt(var I: Integer; V: Integer); begin Trace += 'S'; end;
 property P[var I: Integer]: Integer read Get write SetIt;
end;
var O := new T; var A: array of Integer := [5];
function Receiver: T; begin Trace += 'R'; Result := O; end;
function Container: array of Integer; begin Trace += 'A'; Result := A; end;
function Slot: Integer; begin Trace += 'I'; raise Exception.Create('original index failure'); end;
function RHS: Integer; begin Trace += 'V'; Result := 3; end;
try Receiver().P[Container()[Slot()]] := RHS(); except on E: Exception do PrintLn(E.Message); end;
PrintLn(Trace);`, "original index failure\nRAI\n"},
		{"genericFieldArrayOrder", `var Trace := '';
type T = class A: array of Integer; end;
var O := new T; O.A := [5];
function Receiver: T; begin Trace += 'R'; Result := O; end;
function Slot: Integer; begin Trace += 'I'; Result := 0; end;
function RHS: Integer; begin Trace += 'V'; Result := 3; end;
Receiver().A[Slot()] := RHS(); PrintLn(Trace); PrintLn(O.A[0]);`, "VRI\n3\n"},
	}...)
	orderScript := cases[4].source
	cases = append(cases,
		struct{ name, source, want string }{"boundFailureStopsRHS", strings.ReplaceAll(strings.ReplaceAll(orderScript, "Result := 0; end;", "Result := 1; end;"), "Receiver().P[Container()[Slot()]] := RHS();", "try Receiver().P[Container()[Slot()]] := RHS(); except on E: Exception do PrintLn(E.Message); end;"), "Upper bound exceeded! Index 1 [line: 14, column: 37]\nRAI\n5\n"},
		struct{ name, source, want string }{"defaultSetterOrder", strings.ReplaceAll(strings.ReplaceAll(orderScript, "write SetIt;", "write SetIt; default;"), "Receiver().P[", "Receiver()["), "RAIVS\n8\n"},
		struct{ name, source, want string }{"nestedFieldSetterOrder", strings.ReplaceAll(strings.ReplaceAll(orderScript, "var O := new T;", "var O := new T; type H = class Child: T; end; var Holder := new H; Holder.Child := O;"), "Receiver().P[", "Holder.Child.P["), "AIVS\n8\n"},
		struct{ name, source, want string }{"classSetterOrder", strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(orderScript, "function Get(var", "class function Get(var"), "procedure SetIt(var", "class procedure SetIt(var"), "property P[var", "class property P[var"), "RAIVS\n8\n"},
		struct{ name, source, want string }{"classMetaSetterOrder", strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(orderScript, "function Get(var", "class function Get(var"), "procedure SetIt(var", "class procedure SetIt(var"), "property P[var", "class property P[var"), "Receiver().P[", "T.P["), "AIVS\n8\n"},
		struct{ name, source, want string }{"virtualIdentity", `type T = class
 function Get(var I: Integer): Integer; virtual;
 begin I += 1; Result := I; end;
 property P[var I: Integer]: Integer read Get;
end;
type U = class(T)
 function Get(var I: Integer): Integer; override;
 begin I += 10; Result := I; end;
end;
var O: T := new U; var X := 5;
PrintLn(O.P[X]); PrintLn(X);`, "15\n15\n"},
		struct{ name, source, want string }{"constFailureStopsCapture", `var Trace := '';
type T = class
 function Get(var I: Integer; const J: Integer; K: Integer): Integer; begin Trace += 'G'; Result := I; end;
 procedure SetIt(var I: Integer; const J: Integer; K: Integer; V: Integer); begin Trace += 'S'; end;
 property P[var I: Integer; const J: Integer; K: Integer]: Integer read Get write SetIt;
end;
var O := new T; var X := 5;
function Fail: Integer; begin Trace += 'J'; raise Exception.Create('original const failure'); end;
function Later: Integer; begin Trace += 'K'; Result := 1; end;
function RHS: Integer; begin Trace += 'V'; Result := 3; end;
try O.P[X, Fail(), Later()] := RHS(); except on E: Exception do PrintLn(E.Message); end;
PrintLn(Trace);`, "original const failure\nJ\n"},
	)
	for _, tc := range cases {
		for _, checked := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/checked=%t", tc.name, checked), func(t *testing.T) {
				var output bytes.Buffer
				engine, err := New(WithTypeCheck(checked), WithOutput(&output))
				if err != nil {
					t.Fatal(err)
				}
				if _, err := engine.Eval(tc.source); err != nil {
					t.Fatal(err)
				}
				if got := output.String(); got != tc.want {
					t.Fatalf("output = %q, want %q", got, tc.want)
				}
			})
		}
	}
}

func TestEngineCompile_PropertyIndexDeclarationFixtures(t *testing.T) {
	for _, name := range []string{"array_params1", "array_params2", "array_params3"} {
		t.Run(name, func(t *testing.T) {
			base := "../../testdata/fixtures/FailureScripts/" + name
			source, err := os.ReadFile(base + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			engine, err := New()
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(string(source))
			if program != nil || err == nil {
				t.Fatalf("invalid source compiled: %v, %v", program, err)
			}
			ce, ok := err.(*CompileError)
			if !ok {
				t.Fatalf("error = %T: %v", err, err)
			}
			got := make([]string, len(ce.Errors))
			for i, d := range ce.Errors {
				got[i] = fmt.Sprintf("Syntax Error: %s [line: %d, column: %d]", d.Message, d.Line, d.Column)
				if d.Severity != SeverityError {
					t.Fatalf("unexpected severity %q", d.Severity)
				}
			}
			if strings.Join(got, "\n") != strings.TrimSpace(string(expected)) {
				t.Fatalf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), expected)
			}
		})
	}
}
