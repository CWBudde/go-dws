package frontend

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

const propertyVarPrefix = `type T = class
  function Get(var Access: Integer): Integer; begin Result := Access; end;
  procedure Put(var Store: Integer; Value: Integer); begin Store := Value; end;
  property P[var Declared: Integer]: Integer read Get write Put; default;
end;
var O := new T;
var I := 1;
`

func TestCompile_PropertyIndexUseSite_TermReading(t *testing.T) {
	cases := []struct {
		name, tail, message string
		column              int
		stop, after         bool
	}{
		{"malformedUnreadTail", `PrintLn(O.P[I + ]); Tail; {$ERROR 'late'}`, `")" expected`, 15, true, false},
		{"outer", `PrintLn(O.P[I + Missing]); Tail; {$ERROR 'late'}`, `")" expected`, 15, true, false},
		{"precedence", `PrintLn(O.P[I * 2 + Missing]);`, `")" expected`, 15, true, false},
		{"prefix", `PrintLn(O.P[-I + Missing]);`, `")" expected`, 16, true, false},
		{"groupOuter", `PrintLn(O.P[(I + 1) + Missing]);`, `")" expected`, 21, true, false},
		{"infixNot", `PrintLn(O.P[I not in [1]]);`, `")" expected`, 15, true, false},
		{"groupValue", `PrintLn(O.P[(I + 1)]);`, `Argument 0 (Access) cannot be passed as Var-parameter`, 11, false, true},
		{"groupStorage", `PrintLn(O.P[(I)]);`, "", 0, false, false},
		{"default", `PrintLn(O[I + Missing]);`, `")" expected`, 13, true, false},
		{"setter", `O.P[I + Missing] := MissingValue; Tail;`, `")" expected`, 7, true, false},
		{"childStop", `PrintLn(O.P[Missing + Tail]);`, `Unknown name "Missing"`, 13, true, false},
		{"setterName", `O.P[1] := 2;`, `Argument 0 (Store) cannot be passed as Var-parameter`, 3, false, true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(propertyVarPrefix+tt.tail, "<test>", semantic.HintsLevelNormal)
			var want []string
			if tt.message != "" {
				want = []string{fmt.Sprintf("Syntax Error: %s [line: 8, column: %d]", tt.message, tt.column)}
			}
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
				t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
			}
			if len(want) > 0 && (result.Diagnostics[0].Stop != tt.stop || result.Diagnostics[0].afterChildren != tt.after) {
				t.Fatalf("flags = stop:%t after:%t", result.Diagnostics[0].Stop, result.Diagnostics[0].afterChildren)
			}
		})
	}
}

func TestCompile_PropertyIndexUseSite_ChildOrder(t *testing.T) {
	for _, tt := range []struct {
		tail string
		want []string
		stop bool
	}{
		{`PrintLn(O.P[H(True) + Tail]);`, []string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 9, column: 15]`, `Syntax Error: ")" expected [line: 9, column: 21]`}, true},
		{`PrintLn(O.P[(H(True) + 1)]);`, []string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 9, column: 16]`, `Syntax Error: Argument 0 (Access) cannot be passed as Var-parameter [line: 9, column: 11]`}, false},
		{`PrintLn(O.P[H(I + 1) + Missing]);`, []string{`Syntax Error: ")" expected [line: 9, column: 22]`}, true},
	} {
		result := Compile(propertyVarPrefix+"function H(V: Integer): Integer; begin Result := V; end;\n"+tt.tail, "<test>", semantic.HintsLevelNormal)
		if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(tt.want, "\n") {
			t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(tt.want, "\n"))
		}
		if result.Diagnostics[len(tt.want)-1].Stop != tt.stop {
			t.Fatal("stop flag")
		}
	}
}

func TestCompile_PropertyIndexUseSite_Conversion(t *testing.T) {
	source := `type T = class
  function ReadFloat(var AccessFloat: Float): Integer;
  begin Result := 0; end;
  function ReadInteger(var AccessInteger: Integer): Integer;
  begin Result := 0; end;
  property PF[var DeclaredFloat: Float]: Integer read ReadFloat;
  property PI[var DeclaredInteger: Integer]: Integer read ReadInteger;
end;
var O := new T;
var I: Integer := 1;
var F: Float := 1.5;
const CI: Integer = 1;
const CF: Float = 1.5;
PrintLn(O.PF[I]);
PrintLn(O.PI[F]);
PrintLn(O.PF[1]);
PrintLn(O.PI[1.5]);
PrintLn(O.PF[CF]);
PrintLn(O.PI[CI]);
PrintLn(O.PF[CI]);
PrintLn(O.PI[CF]);`
	messages := []string{`Argument 0 (AccessFloat) cannot be passed as Var-parameter`, `Argument 0 expects type "Integer" instead of "Float"`, `Argument 0 (AccessFloat) cannot be passed as Var-parameter`, `Argument 0 expects type "Integer" instead of "Float"`, `Argument 0 (AccessFloat) cannot be passed as Var-parameter`, `Argument 0 (AccessInteger) cannot be passed as Var-parameter`, `Argument 0 (AccessFloat) cannot be passed as Var-parameter`, `Argument 0 expects type "Integer" instead of "Float"`}
	want := make([]string, len(messages))
	for i, m := range messages {
		want[i] = fmt.Sprintf("Syntax Error: %s [line: %d, column: 11]", m, 14+i)
	}
	result := Compile(source, "<test>", semantic.HintsLevelNormal)
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
	}
	for _, d := range result.Diagnostics {
		if d.Stop || !d.afterChildren {
			t.Fatal("conversion flags")
		}
	}
}

func TestCompile_PropertyIndexUseSite_SelectedOwners(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"interfaceTerm", `type ITest = interface
  function Get(var Access: Integer): Integer;
  property P[var Declared: Integer]: Integer read Get;
end;
var O: ITest;
var I := 1;
PrintLn(O.P[I + Missing]);`, []string{`Syntax Error: ")" expected [line: 7, column: 15]`}},
		{"interfaceVar", `type ITest = interface
  function Get(var Access: Integer): Integer;
  property P[var Declared: Integer]: Integer read Get;
end;
var O: ITest;
var I := 1;
PrintLn(O.P[(I + 1)]);`, []string{`Syntax Error: Argument 0 (Access) cannot be passed as Var-parameter [line: 7, column: 11]`}},
		{"inheritedCompat", `type TBase = class
  function Get(var Access: Integer): Integer; begin Result := Access; end;
  property P[var Declared: Integer]: Integer read Get reintroduce;
end;
type TChild = class(TBase)
  function Read: Integer;
end;
function TChild.Read: Integer;
begin
  var I := 1;
  Result := inherited P()[I + Missing];
end;`, []string{`Hint: Property "P" reintroduced a method, you should remove empty brackets () [line: 11, column: 24]`, `Syntax Error: ")" expected [line: 11, column: 29]`}},
		{"inheritedGroup", `type TBase = class
  function Get(var Access: Integer): Integer; begin Result := Access; end;
  property P[var Declared: Integer]: Integer read Get reintroduce;
end;
type TChild = class(TBase)
  function Read: Integer;
end;
function TChild.Read: Integer;
begin
  var I := 1;
  Result := inherited P()[(I + 1)];
end;`, []string{`Hint: Property "P" reintroduced a method, you should remove empty brackets () [line: 11, column: 24]`, `Syntax Error: Argument 0 (Access) cannot be passed as Var-parameter [line: 11, column: 25]`}},
		{"expressionReader", `type T = class
  F: Integer;
  property P[var Declared: Integer]: Integer read (F) write F;
end;
var O := new T;
O.P[1] := 2;
PrintLn(O.P[1]);`, []string{`Syntax Error: Argument 0 (Declared) cannot be passed as Var-parameter [line: 7, column: 11]`}},
		{"expressionWriter", `type T = class
  F: Integer;
  property P[var Declared: Integer]: Integer read (F) write (F);
end;
var O := new T;
O.P[1] := 2;`, []string{`Syntax Error: Argument 0 (Declared) cannot be passed as Var-parameter [line: 6, column: 3]`}},
		{"fieldWriter", `type T = class
  F: Integer;
  property P[var Declared: Integer]: Integer read (F) write F;
end;
var O := new T;
O.P[1] := 2;
PrintLn(O.F);`, nil},
		{"fieldReaderRejected", `type T = class
  F: Integer;
  property P[var Declared: Integer]: Integer read F;
end;`, []string{`Syntax Error: Function expected [line: 3, column: 51]`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelNormal)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(tt.want, "\n") {
				t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(tt.want, "\n"))
			}
		})
	}
}

func TestCompile_PropertyIndexUseSite_DeferredChecks(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"laterIndexStop", `type T = class
 function Get(var First, Second: Integer): Integer; begin Result := 0; end;
 property P[var A, B: Integer]: Integer read Get;
end;
var O := new T; var I := 1;
PrintLn(O.P[1, I + Missing]); Tail;`, []string{`Syntax Error: ")" expected [line: 6, column: 18]`}},
		{"setterRHSFirst", propertyVarPrefix + `function H(V: Integer): Integer; begin Result := V; end;
O.P[1] := H(True);`, []string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 9, column: 13]`, `Syntax Error: Argument 0 (Store) cannot be passed as Var-parameter [line: 9, column: 3]`}},
		{"staticGateAfterTerms", propertyVarPrefix + `PrintLn(T.P[I + Missing]);`, []string{`Syntax Error: ")" expected [line: 8, column: 15]`}},
		{"readonlyGateAfterTerms", strings.Replace(propertyVarPrefix, " write Put", "", 1) + `O.P[I + Missing] := MissingValue;`, []string{`Syntax Error: ")" expected [line: 8, column: 7]`}},
		{"arrayTermInternal", propertyVarPrefix + `var A: array of Integer := [1, 2];
PrintLn(O.P[A[I + 1] + Missing]);`, []string{`Syntax Error: ")" expected [line: 9, column: 22]`}},
		{"prefixNot", propertyVarPrefix + `PrintLn(O.P[not True + Missing]);`, []string{`Syntax Error: ")" expected [line: 8, column: 22]`}},
		{"infixNotIs", propertyVarPrefix + `PrintLn(O.P[O not is Missing]);`, []string{`Syntax Error: ")" expected [line: 8, column: 15]`}},
		{"infixNotAs", propertyVarPrefix + `PrintLn(O.P[O not as Missing]);`, []string{`Syntax Error: ")" expected [line: 8, column: 15]`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelNormal)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(tt.want, "\n") {
				t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(tt.want, "\n"))
			}
		})
	}
}

func TestCompile_PropertyIndexUseSite_DefaultCommaGroup(t *testing.T) {
	source := `type T = class
 function Get(var First, Second: Integer): Integer; begin Result := 0; end;
 property P[var A, B: Integer]: Integer read Get; default;
end;
var O := new T; var I := 1;
PrintLn(O[1, I + Missing]); Tail;`
	result := Compile(source, "<test>", semantic.HintsLevelNormal)
	want := `Syntax Error: ")" expected [line: 6, column: 16]`
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if !result.Diagnostics[0].Stop {
		t.Fatal("stop flag")
	}
}

func TestCompile_PropertyIndexUseSite_ContextualGroups(t *testing.T) {
	source := `type T = class V: Integer; end;
function Factory: T; begin Result := new T; end;
function IncIt(I: Integer): Integer; begin Result := I + 1; end;
type TFn = function(I: Integer): Integer;
var F: TFn := (IncIt);
var O := (Factory());
var A: array[0..(1 + 1)] of Integer;
const X = (2 + 3);
PrintLn((F)(X)); PrintLn((O is T)); PrintLn((O as T).V);
PrintLn(A[(1 + 1)]);`
	result := Compile(source, "<test>", semantic.HintsLevelNormal)
	if got := result.DiagnosticStrings(); len(got) != 0 {
		t.Fatal(got)
	}
}

func TestCompile_PropertyIndexUseSite_AliasesAndReadonly(t *testing.T) {
	source := `type IntAlias = Integer; FloatAlias = Float;
type T = class
 function Get(var Access: IntAlias; var Other: FloatAlias): Integer; begin Result := 0; end;
 procedure Put(var Store: IntAlias; var OtherStore: FloatAlias; V: Integer); begin end;
 property P[var A: IntAlias; var B: FloatAlias]: Integer read Get write Put;
end;
var O := new T; var I: IntAlias := 1; var F: FloatAlias := 1.5;
O.P[F, I] := 2;
procedure Use(const C: IntAlias);
begin PrintLn(O.P[C, F]); end;
Use(I);`
	result := Compile(source, "<test>", semantic.HintsLevelNormal)
	want := []string{
		`Syntax Error: Argument 0 expects type "IntAlias" instead of "FloatAlias" [line: 8, column: 3]`,
		`Syntax Error: Argument 1 (OtherStore) cannot be passed as Var-parameter [line: 8, column: 3]`,
		`Syntax Error: Argument 0 (Access) cannot be passed as Var-parameter [line: 10, column: 17]`,
	}
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
	}
	for _, d := range result.Diagnostics {
		if d.Stop || !d.afterChildren {
			t.Fatal("late index error flags")
		}
	}
}

func TestCompile_PropertyIndexUseSite_ClassVarTypeIdentity(t *testing.T) {
	source := `type TBase = class end;
type TChild = class(TBase) end;
type T = class
 function Get(var Access: TBase): Integer; begin Access := nil; Result := 0; end;
 property P[var Declared: TBase]: Integer read Get;
end;
var O := new T; var B: TBase := new TChild; var C := new TChild;
PrintLn(O.P[B]);
PrintLn(O.P[C]);
PrintLn(O.P[new TChild]);`
	result := Compile(source, "<test>", semantic.HintsLevelNormal)
	want := []string{
		`Syntax Error: Argument 0 expects type "TBase" instead of "TChild" [line: 9, column: 11]`,
		`Syntax Error: Argument 0 expects type "TBase" instead of "TChild" [line: 10, column: 11]`,
		`Syntax Error: Argument 0 (Access) cannot be passed as Var-parameter [line: 10, column: 11]`,
	}
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
	}
}

func TestCompile_PropertyIndexUseSite_ChildStopSkipsUsage(t *testing.T) {
	source := propertyVarPrefix + `function H(A, B: Integer): Integer; begin Result := A + B; end;
var Tail := 1;
PrintLn(O.P[H(Missing, Tail) + Later]);`
	result := Compile(source, "<test>", semantic.HintsLevelNormal)
	want := `Syntax Error: Unknown name "Missing" [line: 10, column: 15]`
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	symbol, found := result.Analyzer.GetSymbolTable().Resolve("Tail")
	if !found {
		t.Fatal("Tail storage missing")
	}
	if len(symbol.Usages) != 0 {
		t.Fatalf("unread Tail has usages: %v", symbol.Usages)
	}
	if !result.Diagnostics[0].Stop {
		t.Fatal("child stop flag")
	}
}

func TestCompile_PropertyIndexUseSite_SetterRHSStopAndType(t *testing.T) {
	for _, tt := range []struct {
		tail string
		want []string
		stop bool
	}{
		{`O.P[1] := MissingValue; Tail;`, []string{`Syntax Error: Unknown name "MissingValue" [line: 8, column: 11]`}, true},
		{`O.P[1] := True;`, []string{`Syntax Error: Argument 0 (Store) cannot be passed as Var-parameter [line: 8, column: 3]`, `Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 8, column: 3]`}, false},
	} {
		result := Compile(propertyVarPrefix+tt.tail, "<test>", semantic.HintsLevelNormal)
		if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(tt.want, "\n") {
			t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(tt.want, "\n"))
		}
		if result.Diagnostics[len(tt.want)-1].Stop != tt.stop {
			t.Fatal("RHS stop flags")
		}
	}
}

func TestCompile_PropertyIndexUseSite_SelectedModesAfterDeclarationError(t *testing.T) {
	const prefix = `type T = class
 function Get(Access: Integer): Integer; begin Result := Access; end;
 property P[var Declared: Integer]: Integer read Get;
end;
var O := new T;
var I := 1;
`
	for _, tt := range []struct {
		tail  string
		extra string
	}{
		{`PrintLn(O.P[1]);`, ""},
		{`PrintLn(O.P[I + Missing]);`, `Syntax Error: ")" expected [line: 7, column: 15]`},
	} {
		result := Compile(prefix+tt.tail, "<test>", semantic.HintsLevelNormal)
		want := []string{`Syntax Error: Parameter 0 (Declared) - Var-parameter expected [line: 3, column: 50]`, `Syntax Error: Method "Get" has incompatible parameters [line: 3, column: 50]`}
		if tt.extra != "" {
			want = append(want, tt.extra)
		}
		if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
			t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
		}
	}
}

func TestCompile_PropertyIndexUseSite_MissingReaderAndFieldWriter(t *testing.T) {
	for _, tt := range []struct {
		source string
		want   []string
	}{
		{`type T = class
 procedure Put(var Store: Integer; V: Integer); begin end;
 property P[var Declared: Integer]: Integer write Put;
end;
var O := new T; var I := 1;
PrintLn(O.P[I + Missing]);`, []string{`Syntax Error: Cannot read a write only property [line: 6, column: 11]`, `Syntax Error: Array expected [line: 6, column: 12]`}},
		{`type T = class
 F: Integer;
 property P[var Declared: Integer]: Integer read (F) write F;
end;
var O := new T;
O.P[1.5] := 2;`, nil},
		{`type T = class
 F: Integer;
 property P[var Declared: Integer]: Integer read (F) write F;
end;
var O := new T; var I := 1;
O.P[I + Missing] := MissingValue;`, []string{`Syntax Error: ")" expected [line: 6, column: 7]`}},
	} {
		result := Compile(tt.source, "<test>", semantic.HintsLevelNormal)
		if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(tt.want, "\n") {
			t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(tt.want, "\n"))
		}
	}
}

func TestCompile_PropertyIndexUseSite_LexicalSetterNames(t *testing.T) {
	source := `type TBase = class
 function Get(var BaseAccess: Integer): Integer; virtual; begin Result := BaseAccess; end;
 procedure Put(var BaseStore: Integer; V: Integer); virtual; begin BaseStore := V; end;
 property P[var Declared: Integer]: Integer read Get write Put;
end;
type TChild = class(TBase)
 function Get(var ChildAccess: Integer): Integer; override; begin Result := ChildAccess; end;
 procedure Put(var ChildStore: Integer; V: Integer); override; begin ChildStore := V; end;
end;
var O := new TChild;
PrintLn(O.P[1]);
O.P[1] := 2;`
	variants := []struct {
		name, source        string
		readLine, writeLine int
	}{
		{"inherited", source, 11, 12},
		{"promotion", strings.Replace(source, "type TChild = class(TBase)\n", "type TChild = class(TBase)\n property P;\n", 1), 12, 13},
		{"forwarding", strings.ReplaceAll(strings.Replace(source, "type TChild = class(TBase)\n", "type TChild = class(TBase)\n property Q[var Other: Integer]: Integer read P write P;\n", 1), "O.P[", "O.Q["), 12, 13},
	}
	for _, tt := range variants {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelNormal)
			want := []string{
				fmt.Sprintf("Syntax Error: Argument 0 (BaseAccess) cannot be passed as Var-parameter [line: %d, column: 11]", tt.readLine),
				fmt.Sprintf("Syntax Error: Argument 0 (BaseStore) cannot be passed as Var-parameter [line: %d, column: 3]", tt.writeLine),
			}
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
				t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
			}
		})
	}
}

func TestCompile_PropertyIndexUseSite_ProvisionalArrayAndReachedStops(t *testing.T) {
	for _, tt := range []struct {
		source string
		want   string
	}{
		{`var A: array of Integer; var I := 1;
PrintLn(A[I + ]); Tail; {$ERROR 'late'}`, `Syntax Error: Expression expected [line: 2, column: 15]`},
		{`var A: array of Integer;
PrintLn(A[Missing + ]); Tail; {$ERROR 'late'}`, `Syntax Error: Unknown name "Missing" [line: 2, column: 11]`},
		{propertyVarPrefix + `PrintLn(O.P[(I + )]); Tail; {$ERROR 'late'}`, `Syntax Error: Expression expected [line: 8, column: 18]`},
		{propertyVarPrefix + `function H(V: Integer): Integer; begin Result := V; end;
PrintLn(O.P[H(I + )]); Tail; {$ERROR 'late'}`, `Syntax Error: Expression expected [line: 9, column: 19]`},
	} {
		t.Run(tt.want, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelNormal)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_PropertyIndexUseSite_SelectedAccessorRecovery(t *testing.T) {
	for _, tt := range []struct {
		name, source, accessor string
		late                   []string
		child                  bool
	}{
		{"setterValue", `type T = class function Get(var A: Integer): Float; begin Result := 0; end; procedure Put(var Store: Integer; V: Integer); begin end; property P[var Declared: Integer]: Float read Get write Put; end; var O := new T; var I := 1; O.P[I] := 1.5;`, "Put", []string{`Argument 1 expects type "Integer" instead of "Float"`}, false},
		{"getterFew", `type T = class function Get(var A, B: Integer): Integer; begin Result := 0; end; property P[var Declared: Integer]: Integer read Get; end; var O := new T; var I := 1; PrintLn(O.P[I]);`, "Get", []string{"More arguments expected"}, false},
		{"getterFewStorage", `type T = class function Get(var A, B: Integer): Integer; begin Result := 0; end; property P[var Declared: Integer]: Integer read Get; end; var O := new T; PrintLn(O.P[1]);`, "Get", []string{`Argument 0 (A) cannot be passed as Var-parameter`}, false},
		{"getterMany", `type T = class function Get(var A: Integer): Integer; begin Result := 0; end; property P[var First, Second: Integer]: Integer read Get; end; var O := new T; var I := 1; PrintLn(O.P[I, 2]);`, "Get", []string{"Too many arguments"}, false},
		{"getterManyStorage", `type T = class function Get(var A: Integer): Integer; begin Result := 0; end; property P[var First, Second: Integer]: Integer read Get; end; var O := new T; PrintLn(O.P[1, 2]);`, "Get", []string{`Argument 0 (A) cannot be passed as Var-parameter`}, false},
		{"setterFew", `type T = class procedure Put(var A: Integer; B, V: Integer); begin end; property P[var Declared: Integer]: Integer write Put; end; var O := new T; var I := 1; O.P[I] := 2;`, "Put", []string{"More arguments expected"}, false},
		{"setterFewValueType", `type T = class procedure Put(var A: Integer; B, V: Integer); begin end; property P[var Declared: Integer]: Integer write Put; end; var O := new T; var I := 1; O.P[I] := True;`, "Put", []string{`Argument 1 expects type "Integer" instead of "Boolean"`}, false},
		{"setterFewChildError", `type T = class procedure Put(var A: Integer; B, V: Integer); begin end; property P[var Declared: Integer]: Integer write Put; end; function H(V: Integer): Integer; begin Result := V; end; var O := new T; var I := 1; O.P[I] := H(True);`, "Put", []string{"More arguments expected"}, true},
		{"setterMany", `type T = class procedure Put(var A: Integer; V: Integer); begin end; property P[var First, Second: Integer]: Integer write Put; end; var O := new T; var I := 1; O.P[I, 2] := True;`, "Put", []string{"Too many arguments"}, false},
		{"setterManySuppliedType", `type T = class procedure Put(var A: Integer; V: Integer); begin end; property P[var First, Second: Integer]: Integer write Put; end; var O := new T; var I := 1; O.P[I, True] := 2;`, "Put", []string{`Argument 1 expects type "Integer" instead of "Boolean"`}, false},
		{"interfaceSetterValue", `type ITest = interface function Get(var A: Integer): Float; procedure Put(var Store: Integer; V: Integer); property P[var Declared: Integer]: Float read Get write Put; end; var O: ITest; var I := 1; O.P[I] := 1.5;`, "Put", []string{`Argument 1 expects type "Integer" instead of "Float"`}, false},
		{"interfaceGetterFew", `type ITest = interface function Get(var A, B: Integer): Integer; property P[var Declared: Integer]: Integer read Get; end; var O: ITest; var I := 1; PrintLn(O.P[I]);`, "Get", []string{"More arguments expected"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelNormal)
			accessorColumn := strings.LastIndex(tt.source, tt.accessor) + 1
			want := []string{fmt.Sprintf("Syntax Error: Method %q has incompatible parameters [line: 1, column: %d]", tt.accessor, accessorColumn)}
			if tt.child {
				want = append(want, fmt.Sprintf("Syntax Error: Argument 0 expects type %q instead of %q [line: 1, column: %d]", "Integer", "Boolean", strings.LastIndex(tt.source, "True")+1))
			}
			for _, message := range tt.late {
				want = append(want, fmt.Sprintf("Syntax Error: %s [line: 1, column: %d]", message, strings.LastIndex(tt.source, "O.P")+3))
			}
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
				t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
			}
			for _, diagnostic := range result.Diagnostics[len(want)-len(tt.late):] {
				if diagnostic.Stop || !diagnostic.afterChildren {
					t.Fatal("late checking flags")
				}
			}
		})
	}
}
