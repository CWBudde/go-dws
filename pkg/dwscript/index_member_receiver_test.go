package dwscript

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestIndexMemberReceiver_MethodResult(t *testing.T) {
	source := `var Trace := '';
type T = class
 A: array of Integer;
 function GetArray: array of Integer;
 begin Trace += 'G'; Result := A; end;
end;
var One: array of Integer := [5]; var Two: array of Integer := [9];
var First := new T; First.A := One;
var Second := new T; Second.A := Two;
var Calls := 0;
function Receiver: T;
begin Trace += 'R'; Calls += 1; if Calls = 1 then Result := First else Result := Second; end;
function Slot: Integer; begin Trace += 'I'; Result := 0; end;
function RHS: Integer; begin Trace += 'V'; Result := 3; end;
Receiver().GetArray[Slot()] := RHS();
PrintLn(Trace); PrintLn(First.A[0]); PrintLn(Second.A[0]);`
	classSource := strings.ReplaceAll(strings.ReplaceAll(source, " A: array", " class var A: array"), " function GetArray", " class function GetArray")
	defaultSource := strings.ReplaceAll(strings.ReplaceAll(source, "function GetArray:", "function GetArray(N: Integer = 7):"), "Trace += 'G';", "Trace += 'G' + IntToStr(N);")
	overloadSource := strings.ReplaceAll(source, " function GetArray: array of Integer;", " function GetArray(N: Integer): array of Integer; overload; begin Trace += 'W'; Result := A; end;\n function GetArray: array of Integer; overload;")
	exceptionSource := strings.ReplaceAll(strings.ReplaceAll(source, "Trace += 'G'; Result := A;", "Trace += 'G'; raise Exception.Create('original getter failure');"), "Receiver().GetArray[Slot()] := RHS();", "try Receiver().GetArray[Slot()] := RHS(); except on E: Exception do PrintLn(E.Message); end;")
	classSelfSource := strings.ReplaceAll(classSource, "Trace += 'G';", "PrintLn(Self.ClassName); Trace += 'G';")
	helperSource := strings.ReplaceAll(source, "function GetArray: array of Integer;\n begin Trace += 'G'; Result := A; end;\nend;", "end;\ntype H = helper for T\n function GetArray: array of Integer; begin Trace += 'G'; Result := Self.A; end; end;")
	for _, tc := range []struct {
		name, source, want string
		modes              []bool
	}{
		{"default-arguments", defaultSource, "VRG7I\n3\n9\n", []bool{false}},
		{"overload-selected", overloadSource, "VRGI\n3\n9\n", []bool{false}},
		{"original-getter-exception", exceptionSource, "original getter failure\nVRG\n5\n9\n", []bool{true, false}},
		{"class-self", classSelfSource, "T\nVRGI\n3\n3\n", []bool{false}},
		{"helper-call-tail", helperSource, "VRGI\n3\n9\n", []bool{true, false}},
		{"namespace-capture", `var Trace := ''; function Slot: String; begin Trace += 'I'; Result := 'x'; end; function RHS: Integer; begin Trace += 'V'; Result := 3; end; JSON.NewObject[Slot()] := RHS(); PrintLn(Trace);`, "VI\n", []bool{false}},

		{"instance", source, "VRGI\n3\n9\n", []bool{true, false}},
		{"explicit-instance-control", strings.ReplaceAll(source, "GetArray[", "GetArray()["), "VRGI\n3\n9\n", []bool{true, false}},
		{"class", classSource, "VRGI\n3\n3\n", []bool{true, false}},
		{"explicit-class-control", strings.ReplaceAll(classSource, "GetArray[", "GetArray()["), "VRGI\n3\n3\n", []bool{true, false}},
	} {
		for _, checked := range tc.modes {
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

// The runtime's callable probe may return a defaulted overload. Dispatch must
// still prefer the existing exact zero-argument candidate on the captured object.
func TestIndexMemberReceiver_DefaultOverloadPrecedence(t *testing.T) {
	for _, kind := range []string{"instance", "class", "merged"} {
		for _, zeroFirst := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/zero-first=%t", kind, zeroFirst), func(t *testing.T) {
				zero := " function Pick: Integer; overload; begin Result := 0; end;"
				defaults := " function Pick(N: Integer = 7): Integer; overload; begin Result := N; end;"
				if kind != "instance" {
					zero = strings.Replace(zero, " function", " class function", 1)
				}
				if kind == "class" {
					defaults = strings.Replace(defaults, " function", " class function", 1)
				}
				methods := zero + "\n" + defaults
				if !zeroFirst {
					methods = defaults + "\n" + zero
				}
				source := "type T = class\n" + methods + "\nend;\nvar O := new T;\nPrintLn(O.Pick); PrintLn(O.Pick());"
				want := "0\n0\n"
				// Class-method bare reads with parameters already return a pointer.
				if kind == "class" && zeroFirst {
					want = "@class T.Pick\n0\n"
				}
				var output bytes.Buffer
				engine, err := New(WithTypeCheck(false), WithOutput(&output))
				if err != nil {
					t.Fatal(err)
				}
				if _, err := engine.Eval(source); err != nil {
					t.Fatal(err)
				}
				if got := output.String(); got != want {
					t.Fatalf("output = %q, want %q", got, want)
				}
			})
		}
	}
}

func TestIndexMemberReceiver_DefaultOverloadContainer(t *testing.T) {
	for _, zeroFirst := range []bool{true, false} {
		for _, explicit := range []bool{false, true} {
			modes := []bool{false}
			if zeroFirst && !explicit {
				modes = append(modes, true)
			}
			for _, checked := range modes {
				t.Run(fmt.Sprintf("zero-first=%t/explicit=%t/checked=%t", zeroFirst, explicit, checked), func(t *testing.T) {
					zero := " function GetArray: array of Integer; overload; begin Trace += 'G'; Result := A; end;"
					defaults := " function GetArray(N: Integer = 7): array of Integer; overload; begin Trace += 'W'; Result := B; end;"
					methods := zero + "\n" + defaults
					if !zeroFirst {
						methods = defaults + "\n" + zero
					}
					call := "Receiver().GetArray"
					if explicit {
						call += "()"
					}
					source := `var Trace := '';
type T = class
 A, B: array of Integer;
` + methods + `
end;
var One: array of Integer := [5]; var Two: array of Integer := [9];
var WrongOne: array of Integer := [11]; var WrongTwo: array of Integer := [13];
var First := new T; First.A := One; First.B := WrongOne;
var Second := new T; Second.A := Two; Second.B := WrongTwo;
var Calls := 0;
function Receiver: T;
begin Trace += 'R'; Calls += 1; if Calls = 1 then Result := First else Result := Second; end;
function Slot: Integer; begin Trace += 'I'; Result := 0; end;
function RHS: Integer; begin Trace += 'V'; Result := 3; end;
` + call + `[Slot()] := RHS();
PrintLn(Trace); PrintLn(First.A[0]); PrintLn(Second.A[0]);
PrintLn(First.B[0]); PrintLn(Second.B[0]);`
					var output bytes.Buffer
					engine, err := New(WithTypeCheck(checked), WithOutput(&output))
					if err != nil {
						t.Fatal(err)
					}
					if _, err := engine.Eval(source); err != nil {
						t.Fatal(err)
					}
					want := "VRGI\n3\n9\n11\n13\n"
					if got := output.String(); got != want {
						t.Fatalf("output = %q, want %q", got, want)
					}
				})
			}
		}
	}
}
