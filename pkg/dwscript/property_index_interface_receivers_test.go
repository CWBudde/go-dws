package dwscript

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

const interfaceReceiverPrefix = `type Ints = array of Integer;
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

// The interface probe's ordinary-method fallback preserves public acceptance
// and diagnostic text while analyzing each reached receiver only once.
func TestInterfacePropertyIndexModes_CheckedCompatibilityReceiverDiagnostics(t *testing.T) {
	const prefix = "type TInner = class function Items: array of Integer; begin Result := [1]; end; end;\ntype TOuter = class FInner: TInner; property Inner: TInner read FInner; deprecated 'old'; end;\nvar Obj := new TOuter;\n"
	for _, tc := range []struct {
		name, tail, want string
		accepted         bool
	}{
		{"deprecated", `var A := Obj.Inner.Items()[0];`, `Warning: "Inner" has been deprecated: old [line: 4, column: 14]`, true},
		{"unknown", `var A := Obj.Missing.Items()[0];`, `Syntax Error: There is no accessible member with name "Missing" for type TOuter [line: 4, column: 14]`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := prefix + tc.tail
			engine, err := New()
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(source)
			if tc.accepted {
				if program == nil || err != nil {
					t.Fatalf("Compile = %v, %v; want accepted nonnil Program", program, err)
				}
				if got := program.analyzer.Errors(); len(got) != 1 || got[0] != tc.want {
					t.Fatalf("accepted receiver diagnostics = %q, want [%q]", got, tc.want)
				}
			} else {
				if program != nil || err == nil {
					t.Fatalf("Compile = %v, %v; want rejected nil Program", program, err)
				}
				public := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
				if got := strings.Join(public.DiagnosticStrings(), "\n"); got != tc.want {
					t.Fatalf("public diagnostics = %q, want %q", got, tc.want)
				}
			}
			// Accepted Compile has no public warning result, so supplement its
			// Program assertion with the same pipeline's rendered warning list.
			metadata := frontend.CompileWithOptions(source, frontend.Options{HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(metadata.DiagnosticStrings(), "\n"); got != tc.want {
				t.Fatalf("rendered diagnostics = %q, want %q", got, tc.want)
			}
		})
	}
}

// Earlier receiver brackets and later result brackets must not become getter
// arguments. These two sources fail independently in Compile and Run.
func TestInterfacePropertyIndexModes_CheckedReceivers(t *testing.T) {
	for _, tc := range []struct{ name, tail, want string }{
		{"array", `var Receivers: array of I := [O]; function ReceiverSlot: Integer; begin Trace += 'R'; Result := 0; end; PrintLn(Receivers[ReceiverSlot()][X, Other()][ResultSlot()]); PrintLn(X); PrintLn(Trace);`, "6\n6\nRJGK\n"},
		{"arrayReturn", `function Receivers: array of I; begin Trace += 'R'; Result := [O]; end; PrintLn(Receivers()[0][X, Other()][ResultSlot()]); PrintLn(X); PrintLn(Trace);`, "6\n6\nRJGK\n"},
		{"namedArray", `var Receivers: array of I := [O]; function ReceiverSlot: Integer; begin Trace += 'R'; Result := 0; end; PrintLn(Receivers[ReceiverSlot()].P[X, Other()][ResultSlot()]); PrintLn(X); PrintLn(Trace);`, "6\n6\nRJGK\n"},
		{"identifier", `PrintLn(O[X, Other()][ResultSlot()]); PrintLn(X); PrintLn(Trace);`, "6\n6\nJGK\n"},
		{"field", `type Holder = class R: I; end; var H := new Holder; H.R := O; PrintLn(H.R[X, Other()][ResultSlot()]); PrintLn(X); PrintLn(Trace);`, "6\n6\nJGK\n"},
		{"arrayField", `type Holder = class Receivers: array of I; end; var H := new Holder; H.Receivers := [O]; PrintLn(H.Receivers[0][X, Other()][ResultSlot()]); PrintLn(X); PrintLn(Trace);`, "6\n6\nJGK\n"},
		{"returnedField", `type Holder = class R: I; end; var H := new Holder; H.R := O; function Box: Holder; begin Trace += 'R'; Result := H; end; PrintLn(Box().R[X, Other()][ResultSlot()]); PrintLn(X); PrintLn(Trace);`, "6\n6\nRJGK\n"},
		{"interfaceReturn", `function Receiver: I; begin Trace += 'R'; Result := O; end; PrintLn(Receiver()[X, Other()][ResultSlot()]); PrintLn(X); PrintLn(Trace);`, "6\n6\nRJGK\n"},
		{"aliasCase", `type Alias = I; var Receivers: array of Alias := [O]; PrintLn(rEcEiVeRs[0][x, oThEr()][rEsUlTsLoT()]); PrintLn(X); PrintLn(Trace);`, "6\n6\nJGK\n"},
	} {
		t.Run(tc.name, func(t *testing.T) { runPropertyIndexProgram(t, interfaceReceiverPrefix+tc.tail, tc.want, true) })
	}
}

const interfaceAggregateCapturePrefix = `var Trace := '';
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
 property Q[A: Integer]: Integer read (A + 1000); default;
end;
var First := new T; var Second := new T; var O: I := First;
var Receivers: array of I := [O];
var A: array of Integer := [5]; var Saved := A;
function ReceiverSlot: Integer; begin Trace += 'R'; Result := 0; end;
function Slot: Integer; begin Trace += 'I'; Result := 0; end;
function Other: Integer; begin Trace += 'J'; A := [100]; O := Second; Receivers := [O]; Result := 1; end;
function RHS: Integer; begin Trace += 'V'; A := [200]; O := Second; Receivers := [O]; Result := 3; end;
`

// A later index or RHS can rebind both arrays; the original receiver, original
// writable slot and static interface accessor must survive for reads and writes.
func TestInterfacePropertyIndexModes_CheckedAggregateCapture(t *testing.T) {
	for _, tc := range []struct{ name, tail, want string }{
		{"read", `PrintLn(Receivers[ReceiverSlot()][A[Slot()], Other()]);`, "6\n6\n100\n0\n0\nRIJG\n"},
		{"setter", `Receivers[ReceiverSlot()][A[Slot()], Other()] := RHS();`, "3\n200\n3\n0\nRIJVS\n"},
		{"compound", `Receivers[ReceiverSlot()][A[Slot()], Other()] += RHS();`, "9\n200\n9\n0\nRIJGVS\n"},
		{"namedSetter", `Receivers[ReceiverSlot()].P[A[Slot()], Other()] := RHS();`, "3\n200\n3\n0\nRIJVS\n"},
		{"namedCompound", `Receivers[ReceiverSlot()].P[A[Slot()], Other()] += RHS();`, "9\n200\n9\n0\nRIJGVS\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runPropertyIndexProgram(t, interfaceAggregateCapturePrefix+tc.tail+` PrintLn(Saved[0]); PrintLn(A[0]); PrintLn(First.F); PrintLn(Second.F); PrintLn(Trace);`, tc.want, true)
		})
	}
}

func TestInterfacePropertyIndexModes_CheckedEmptyRecovery(t *testing.T) {
	for _, tc := range []struct{ name, tail, want string }{
		{"directive", "PrintLn(O[]);\n{$ERROR 'late'}", "Syntax Error: More arguments expected [line: 13, column: 10]\nSyntax Error: late [line: 14, column: 3]"},
		{"aggregate", "var Receivers: array of I := [O];\nPrintLn(Receivers[0][]);\n{$ERROR 'late'}", "Syntax Error: More arguments expected [line: 14, column: 21]\nSyntax Error: late [line: 15, column: 3]"},
		{"forward", "procedure Pending; forward;\nPrintLn(O[]);", "Syntax Error: More arguments expected [line: 14, column: 10]\nSyntax Error: The function \"Pending\" was forward declared but not implemented [line: 13, column: 11]"},
		{"ordinaryForward", "procedure Pending; forward;\nvar A: array of Integer;\nPrintLn(A[]);", "Syntax Error: Expression expected [line: 15, column: 11]"},
		{"ordinary", "var A: array of Integer;\nPrintLn(A[]);\n{$ERROR 'late'}", "Syntax Error: Expression expected [line: 14, column: 11]"},
		{"enclosingStop", "PrintLn(O[]);\nvar Broken := ;\n{$ERROR 'late'}", "Syntax Error: More arguments expected [line: 13, column: 10]\nSyntax Error: Expression expected [line: 14, column: 15]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := compilePropertyUseSite(t, interfaceReceiverPrefix+tc.tail, "<test>", semantic.HintsLevelNormal)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tc.want {
				t.Fatalf("diagnostics = %q, want %q", got, tc.want)
			}
		})
	}
	t.Run("parseOnly", func(t *testing.T) {
		engine, err := New()
		if err != nil {
			t.Fatal(err)
		}
		_, err = engine.Parse(interfaceReceiverPrefix + "PrintLn(O[]);\n{$ERROR 'late'}")
		ce, ok := err.(*CompileError)
		if !ok || len(ce.Errors) != 1 || ce.Errors[0].Message != "Expression expected" || ce.Errors[0].Line != 13 || ce.Errors[0].Column != 11 {
			t.Fatalf("Parse diagnostics = %v", err)
		}
	})
}

func TestInterfacePropertyIndexModes_CheckedReceiverProbeUsage(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(interfaceReceiverPrefix + `type Holder = class R: I; end; var H := new Holder; H.R := O;
function Box: Holder; begin Result := H; end;
PrintLn(Box().R[X, Other()][ResultSlot()]);`)
	if err != nil || program == nil {
		t.Fatalf("Compile = %v, %v", program, err)
	}
	box, found := program.analyzer.GetSymbolTable().Resolve("Box")
	if !found || len(box.Usages) != 1 {
		t.Fatalf("Box receiver usages = %v, want one reached receiver", box)
	}
}

func TestInterfacePropertyIndexModes_CheckedDeclinedReceiverUsage(t *testing.T) {
	for _, tc := range []struct {
		name, tail, want string
		uses             int
	}{
		{"one", `PrintLn(O[X]);`, "Syntax Error: Array expected [line: 13, column: 10]", 1},
		{"two", `PrintLn(O[X]); PrintLn(O[X]);`, "Syntax Error: Array expected [line: 13, column: 10]\nSyntax Error: Array expected [line: 13, column: 25]", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := strings.Replace(interfaceReceiverPrefix, "; default;", ";", 1) + tc.tail
			public := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
			if got := strings.Join(public.DiagnosticStrings(), "\n"); got != tc.want {
				t.Fatalf("public diagnostics = %q, want %q", got, tc.want)
			}
			metadata := frontend.CompileWithOptions(source, frontend.Options{HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(metadata.DiagnosticStrings(), "\n"); got != tc.want {
				t.Fatalf("metadata diagnostics = %q, want %q", got, tc.want)
			}
			symbol, found := metadata.Analyzer.GetSymbolTable().Resolve("O")
			if !found || len(symbol.Usages) != tc.uses {
				t.Fatalf("O uses = %v, want %d reached source occurrences", symbol, tc.uses)
			}
			x, found := metadata.Analyzer.GetSymbolTable().Resolve("X")
			if !found || len(x.Usages) != 0 {
				t.Fatalf("unread X uses = %v", x)
			}
		})
	}
}

// These metadata assertions supplement the actual public Compile rejection:
// count checking happens after reached operands, while a ReadTerm stop leaves
// the remaining index operands and result index unread.
func TestInterfacePropertyIndexModes_CheckedReceiverReach(t *testing.T) {
	for _, tc := range []struct {
		name, tail, want string
		uses             [3]int
		stop, after      bool
	}{
		{"extra", `PrintLn(O[X, Other(), ResultSlot()]);`, "Syntax Error: Too many arguments [line: 13, column: 10]", [3]int{1, 1, 1}, false, true},
		{"few", `PrintLn(O[X]);`, "Syntax Error: More arguments expected [line: 13, column: 10]", [3]int{1, 0, 0}, false, true},
		{"storageBeforeCount", `PrintLn(O[5, Other(), ResultSlot()]);`, "Syntax Error: Argument 0 (X) cannot be passed as Var-parameter [line: 13, column: 10]", [3]int{0, 1, 1}, false, true},
		{"aggregateTermStop", "var Receivers: array of I := [O];\nPrintLn(Receivers[0][X + Missing, Other()][ResultSlot()]);", "Syntax Error: \")\" expected [line: 14, column: 24]", [3]int{1, 0, 0}, true, false},
		{"declinedReceiverError", `PrintLn(Missing.R[X, Other()][ResultSlot()]);`, "Syntax Error: Unknown name \"Missing\" [line: 13, column: 9]", [3]int{0, 0, 0}, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := interfaceReceiverPrefix + tc.tail
			public := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
			if got := strings.Join(public.DiagnosticStrings(), "\n"); got != tc.want {
				t.Fatalf("public diagnostics = %q, want %q", got, tc.want)
			}
			result := frontend.CompileWithOptions(source, frontend.Options{HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tc.want {
				t.Fatalf("metadata diagnostics = %q, want %q", got, tc.want)
			}
			structured := result.Analyzer.StructuredErrors()
			if len(result.Diagnostics) != 1 || result.Diagnostics[0].Stop != tc.stop || len(structured) != 1 || structured[0].AfterChildren != tc.after {
				t.Fatalf("diagnostic reach = %+v; structured = %+v; want stop=%t after=%t", result.Diagnostics, structured, tc.stop, tc.after)
			}
			for n, name := range []string{"X", "Other", "ResultSlot"} {
				symbol, found := result.Analyzer.GetSymbolTable().Resolve(name)
				if !found || len(symbol.Usages) != tc.uses[n] {
					t.Fatalf("%s uses = %v, want %d", name, symbol, tc.uses[n])
				}
			}
		})
	}
}

func TestInterfacePropertyIndexModes_CheckedAggregateStops(t *testing.T) {
	const aggregate = `var Receivers: array of I := [O];
function ReceiverSlot: Integer; begin Trace += 'R'; Result := 0; end;
var A: array of Integer := [5];
function BadSlot: Integer; begin Trace += 'I'; raise Exception.Create('original var failure'); end;
`
	for _, tc := range []struct{ name, prefix, tail, want string }{
		{"receiver", strings.Replace(interfaceReceiverPrefix+aggregate, "Trace += 'R'; Result := 0", "Trace += 'R'; raise Exception.Create('original receiver failure')", 1), `try PrintLn(Receivers[ReceiverSlot()][X, Other()][ResultSlot()]); except on E: Exception do PrintLn(E.Message); end;`, "original receiver failure\n5\nR\n"},
		{"varIndex", interfaceReceiverPrefix + aggregate, `try PrintLn(Receivers[ReceiverSlot()][A[BadSlot()], Other()][ResultSlot()]); except on E: Exception do PrintLn(E.Message); end;`, "original var failure\n5\nRI\n"},
		{"constIndex", strings.Replace(interfaceReceiverPrefix+aggregate, "Trace += 'J'; Result := 1", "Trace += 'J'; raise Exception.Create('original const failure')", 1), `try PrintLn(Receivers[ReceiverSlot()][X, Other()][ResultSlot()]); except on E: Exception do PrintLn(E.Message); end;`, "original const failure\n5\nRJ\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runPropertyIndexProgram(t, tc.prefix+tc.tail+` PrintLn(X); PrintLn(Trace);`, tc.want, true)
		})
	}
	t.Run("rhs", func(t *testing.T) {
		source := strings.Replace(interfaceAggregateCapturePrefix, "Trace += 'V'; A := [200]; O := Second; Receivers := [O]; Result := 3", "Trace += 'V'; raise Exception.Create('original RHS failure')", 1)
		runPropertyIndexProgram(t, source+`try Receivers[ReceiverSlot()][A[Slot()], Other()] := RHS(); except on E: Exception do PrintLn(E.Message); end; PrintLn(Saved[0]); PrintLn(A[0]); PrintLn(First.F); PrintLn(Second.F); PrintLn(Trace);`, "original RHS failure\n5\n100\n0\n0\nRIJV\n", true)
	})
	for _, tc := range []struct{ name, prefix, tail, want, message string }{
		{"receiverError", interfaceReceiverPrefix + strings.Replace(aggregate, "Trace += 'R'; Result := 0", "PrintLn('receiver'); Result := 1 div 0", 1), `PrintLn(Receivers[ReceiverSlot()][X, Other()][ResultSlot()]);`, "receiver\n", "Division by zero"},
		{"varStorageError", interfaceReceiverPrefix + aggregate, `PrintLn(Receivers[ReceiverSlot()][A[1], Other()][ResultSlot()]);`, "", "Upper bound exceeded! Index 1"},
		{"nilRead", strings.Replace(interfaceReceiverPrefix+aggregate, "var O: I := new T", "var O: I", 1), `PrintLn(Receivers[ReceiverSlot()][X, Other()][ResultSlot()]);`, "", "interface is nil"},
		{"nilWrite", strings.ReplaceAll(strings.Replace(interfaceAggregateCapturePrefix, "var O: I := First", "var O: I", 1), "Trace +=", "PrintLn('effect'); Trace +="), `Receivers[ReceiverSlot()][A[Slot()], Other()] := RHS();`, "effect\neffect\neffect\neffect\n", "setter cannot be executed on nil interface"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(tc.prefix + tc.tail)
			if err != nil || program == nil {
				t.Fatalf("Compile = %v, %v", program, err)
			}
			result, err := engine.Run(program)
			if err == nil || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("Run = %v, %v, want %q", result, err, tc.message)
			}
			if got := output.String(); got != tc.want {
				t.Fatalf("output = %q, want %q", got, tc.want)
			}
		})
	}
}

// One comma group owns the count diagnostic; no operand may be borrowed from a
// later bracket or silently turned into indexing the getter's array result.
func TestInterfacePropertyIndexModes_CheckedReceiverGroups(t *testing.T) {
	for _, tc := range []struct{ name, tail, message string }{
		{"few", `PrintLn(O[X]);`, "More arguments expected"},
		{"extra", `PrintLn(O[X, Other(), ResultSlot()]);`, "Too many arguments"},
		{"empty", `PrintLn(O[]);`, "More arguments expected"},
		{"split", `PrintLn(O[X][Other()][ResultSlot()]);`, "More arguments expected"},
		{"invalidVarExtra", `PrintLn(O[5, Other(), ResultSlot()]);`, "Argument 0 (X) cannot be passed as Var-parameter"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := compilePropertyUseSite(t, interfaceReceiverPrefix+tc.tail, "<test>", semantic.HintsLevelNormal)
			want := fmt.Sprintf("Syntax Error: %s [line: 13, column: 10]", tc.message)
			got := result.DiagnosticStrings()
			if tc.name == "split" {
				// The source pins the first group's primary, not the complete
				// recovery cascade through malformed later brackets.
				if len(got) == 0 || got[0] != want {
					t.Fatalf("diagnostics = %q, want primary %q", got, want)
				}
			} else if strings.Join(got, "\n") != want {
				t.Fatalf("diagnostics = %q, want %q", got, want)
			}
		})
	}
}

func TestInterfacePropertyIndexModes_CheckedAggregateValidation(t *testing.T) {
	for _, tc := range []struct{ name, prefix, tail, want string }{
		{"storage", interfaceReceiverPrefix, "var Receivers: array of I := [O];\nPrintLn(Receivers[0][5, Other()]);", "Syntax Error: Argument 0 (X) cannot be passed as Var-parameter [line: 14, column: 21]"},
		{"type", interfaceReceiverPrefix, "var Receivers: array of I := [O];\nPrintLn(Receivers[0][1.5, Other()]);", "Syntax Error: Argument 0 expects type \"Integer\" instead of \"Float\" [line: 14, column: 21]"},
		{"groupedValue", interfaceReceiverPrefix, "var Receivers: array of I := [O];\nPrintLn(Receivers[0][(X + 1), Other()]);", "Syntax Error: Argument 0 (X) cannot be passed as Var-parameter [line: 14, column: 21]"},
		{"readonly", strings.Replace(interfaceAggregateCapturePrefix, "read Get write Put; default;", "read Get; default;", 1), `Receivers[0][A[0], 1] := RHS();`, "Syntax Error: Cannot set a value for a read-only property [line: 21, column: 1]"},
		{"writeonly", strings.Replace(interfaceAggregateCapturePrefix, "read Get write Put; default;", "write Put; default;", 1), `PrintLn(Receivers[0][A[0], 1]);`, "Syntax Error: Cannot read a write only property [line: 21, column: 21]\nSyntax Error: Array expected [line: 21, column: 21]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := compilePropertyUseSite(t, tc.prefix+tc.tail, "<test>", semantic.HintsLevelNormal)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tc.want {
				t.Fatalf("diagnostics = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestInterfacePropertyIndexModes_CheckedValueReceiverGroups(t *testing.T) {
	for _, mode := range []struct{ name, declaration string }{{"value", ""}, {"const", "const "}} {
		t.Run(mode.name, func(t *testing.T) {
			prefix := strings.ReplaceAll(interfaceReceiverPrefix, "var X: Integer;", mode.declaration+"X: Integer;")
			prefix = strings.ReplaceAll(prefix, "var A: Integer;", mode.declaration+"A: Integer;")
			prefix = strings.ReplaceAll(prefix, "X += Y;", "")
			for _, tc := range []struct{ name, tail, want string }{
				{"few", `PrintLn(O[X]);`, "Syntax Error: More arguments expected [line: 13, column: 10]"},
				{"extra", `PrintLn(O[X, Other(), ResultSlot()]);`, "Syntax Error: Too many arguments [line: 13, column: 10]"},
				{"typeBeforeCount", `PrintLn(O[True, Other(), ResultSlot()]);`, "Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 13, column: 10]"},
				{"matchedType", `PrintLn(O[True, Other()]);`, "Syntax Error: Array index expected \"Integer\" but got \"Boolean\" [line: 13, column: 11]"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					result := compilePropertyUseSite(t, prefix+tc.tail, "<test>", semantic.HintsLevelNormal)
					if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tc.want {
						t.Fatalf("diagnostics = %q, want %q", got, tc.want)
					}
					if tc.name == "extra" || tc.name == "typeBeforeCount" {
						metadata := frontend.CompileWithOptions(prefix+tc.tail, frontend.Options{HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
						for _, name := range []string{"Other", "ResultSlot"} {
							symbol, found := metadata.Analyzer.GetSymbolTable().Resolve(name)
							if !found || len(symbol.Usages) != 1 {
								t.Fatalf("%s reached uses = %v", name, symbol)
							}
						}
					}
				})
			}
			runPropertyIndexProgram(t, prefix+`PrintLn(O[X, Other()][ResultSlot()]); PrintLn(X); PrintLn(Trace);`, "5\n5\nJGK\n", true)
			t.Run("reachedTypedChild", func(t *testing.T) {
				source := prefix + "function Wrong: Boolean; begin Result := True; end;\nPrintLn(O[Wrong(), Other(), ResultSlot()]);"
				want := "Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 14, column: 10]"
				public := compilePropertyUseSite(t, source, "<test>", semantic.HintsLevelNormal)
				if got := strings.Join(public.DiagnosticStrings(), "\n"); got != want {
					t.Fatalf("diagnostics = %q, want %q", got, want)
				}
				metadata := frontend.CompileWithOptions(source, frontend.Options{HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
				for _, name := range []string{"Wrong", "Other", "ResultSlot"} {
					symbol, found := metadata.Analyzer.GetSymbolTable().Resolve(name)
					if !found || len(symbol.Usages) != 1 {
						t.Fatalf("%s reached uses = %v", name, symbol)
					}
				}
			})
		})
	}
}
