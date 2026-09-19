package interp

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
)

func TestArrayElementVar_CompiledFixtures(t *testing.T) {
	for _, fixture := range []string{
		"ArrayPass/array_element_byref.pas",
		"SimpleScripts/const_array_empty.pas",
		"ArrayPass/array_element_var.pas",
		"ArrayPass/array_element_var2.pas",
		"ArrayPass/array_element_var3.pas",
	} {
		t.Run(fixture, func(t *testing.T) {
			category := filepath.Dir(fixture)
			result, detail := runFixtureTest(filepath.Join(fixturesRoot, fixture), false, fixtureconfig.HintsLevel(category))
			if result != testResultPassed {
				t.Fatalf("%s: %v: %s", fixture, result, detail)
			}
		})
	}
}

func TestArrayElementVar_MemberIndicesWriteThrough(t *testing.T) {
	for _, index := range []string{"i", "2+2", "a.High", "a.Length-1"} {
		t.Run(index, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, `
var a: array of Integer := [0, 0, 0, 0, 10];
var i := 4;
procedure Mutate(var value: Integer);
begin value := value + 7; end;
Mutate(a[`+index+`]);
PrintLn(a[4]);
`), "17\n")
		})
	}
}

func TestArrayElementVar_VariantIndexConversion(t *testing.T) {
	for _, index := range []string{"0", "'0'"} {
		t.Run(index, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, `
var a: array of Integer := [10];
var i: Variant := `+index+`;
procedure Mutate(var value: Integer);
begin value := 42; end;
Mutate(a[i]);
PrintLn(a[0]);
`), "42\n")
		})
	}
	t.Run("invalid string stops call", func(t *testing.T) {
		assertOutput(t, runQuickwinScript(t, `
var a: array of Integer := [10];
var i: Variant := 'invalid';
function Later: Integer;
begin PrintLn('later argument'); Result := 0; end;
procedure Mutate(var value: Integer; last: Integer);
begin PrintLn('body'); value := 42; end;
try Mutate(a[i], Later());
except on E: Exception do PrintLn(E.Message); end;
PrintLn(a[0]);
`), "Could not cast variant from String to Integer\n10\n")
	})
}

func TestArrayElementVar_ReceiverAndIndexEvaluatedOnce(t *testing.T) {
	for _, overload := range []bool{false, true} {
		name, extra := "single", ""
		if overload {
			name = "overloaded"
			extra = "procedure Mutate(value: String); overload; begin PrintLn('wrong overload'); end;"
		}
		t.Run(name, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, `
type TItems = array of Integer;
var a: TItems := [10];
function Items: TItems;
begin PrintLn('receiver'); Result := a; end;
function NextIndex: Integer;
begin PrintLn('index'); Result := 0; end;
procedure Mutate(var value: Integer); overload;
begin PrintLn('body'); value := value + 7; end;
`+extra+`
Mutate(Items()[NextIndex()]);
PrintLn(a[0]);
`), "receiver\nindex\nbody\n17\n")
		})
	}
}

func TestArrayElementVar_ArgumentEvaluationOrder(t *testing.T) {
	for _, overload := range []bool{false, true} {
		name, extra := "single", ""
		if overload {
			name = "overloaded"
			extra = "procedure Mutate(first: Integer; value: String; last: Integer); overload; begin PrintLn('wrong overload'); end;"
		}
		t.Run(name, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, `
var a: array of Integer := [10];
function Mark(labelText: String): Integer;
begin PrintLn(labelText); Result := 0; end;
procedure Mutate(first: Integer; var value: Integer; last: Integer); overload;
begin PrintLn('body'); value := 42; end;
`+extra+`
Mutate(Mark('first'), a[Mark('index')], Mark('last'));
PrintLn(a[0]);
`), "first\nindex\nlast\nbody\n42\n")
		})
	}
}

func TestArrayElementVar_NestedMemberReceiverEvaluatedOnce(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
type THolder = class
  Matrix: array of array of Integer;
end;
var holder := THolder.Create;
holder.Matrix.SetLength(1);
holder.Matrix[0].SetLength(1);
function NextHolder: THolder;
begin PrintLn('receiver'); Result := holder; end;
function NextRow: Integer;
begin PrintLn('row'); Result := 0; end;
function NextCol: Integer;
begin PrintLn('column'); Result := 0; end;
procedure Mutate(var value: Integer);
begin value := 42; end;
Mutate(NextHolder().Matrix[NextRow()][NextCol()]);
PrintLn(holder.Matrix[0][0]);
`), "receiver\nrow\ncolumn\n42\n")
}

func TestArrayElementVar_FailureStopsArgumentPreparation(t *testing.T) {
	for _, tt := range []struct {
		name, target, failure, want string
	}{
		{"receiver exception", "FailArray()[0]", "raise Exception.Create('receiver failed');", "receiver failed"},
		{"index exception", "a[FailIndex()]", "raise Exception.Create('index failed');", "index failed"},
		{"index runtime error", "a[FailIndex()]", "var zero := 0; Result := 1 div zero;", "Division by zero"},
		{"lower bound", "a[-1]", "Result := 0;", "Lower bound exceeded! Index -1"},
		{"upper member bound", "a[a.Length]", "Result := 0;", "Upper bound exceeded! Index 1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			arrayBody := "Result := a;"
			indexBody := tt.failure
			if tt.name == "receiver exception" {
				arrayBody, indexBody = tt.failure, "Result := 0;"
			}
			got := runQuickwinScript(t, `
type TItems = array of Integer;
var a: TItems := [10];
function FailArray: TItems;
begin `+arrayBody+` end;
function FailIndex: Integer;
begin `+indexBody+` end;
function Later: Integer;
begin PrintLn('later argument'); Result := 0; end;
procedure Mutate(var value: Integer; last: Integer);
begin PrintLn('body'); value := 42; end;
try
  Mutate(`+tt.target+`, Later());
except
  on E: Exception do PrintLn(E.Message);
end;
PrintLn(a[0]);
`)
			lines := strings.Split(strings.TrimSpace(got), "\n")
			if len(lines) != 2 || !strings.HasPrefix(lines[0], tt.want) || lines[1] != "10" {
				t.Fatalf("expected original failure, no later argument or body, unchanged array; got %q", got)
			}
		})
	}
}

func TestArrayElementVar_NonzeroStaticBounds(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var a: array [3..5] of Integer;
procedure Mutate(var value: Integer);
begin value := 42; end;
Mutate(a[a.Low]);
Mutate(a[a.High]);
PrintLn(a[3]);
PrintLn(a[4]);
PrintLn(a[5]);
`), "42\n0\n42\n")
}

func TestArrayElementVar_OverloadedFailureStopsArgumentPreparation(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var a: array of Integer := [10];
function FailIndex: Integer;
begin PrintLn('index'); raise Exception.Create('index failed'); end;
function Later: Integer;
begin PrintLn('later argument'); Result := 0; end;
procedure Mutate(var value: Integer; last: Integer); overload;
begin PrintLn('body'); value := 42; end;
procedure Mutate(value: String; last: Integer); overload;
begin PrintLn('wrong overload'); end;
try
  Mutate(a[FailIndex()], Later());
except
  on E: Exception do PrintLn(E.Message);
end;
PrintLn(a[0]);
`), "index\nindex failed\n10\n")
}

func TestArrayElementVar_OverloadValueArgumentStaysByValue(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var a: array of String := ['original'];
function NextIndex: Integer;
begin PrintLn('index'); Result := 0; end;
procedure Mutate(var value: Integer); overload;
begin PrintLn('wrong overload'); end;
procedure Mutate(value: String); overload;
begin value := 'local'; PrintLn(value); end;
Mutate(a[NextIndex()]);
PrintLn(a[0]);
`), "index\nlocal\noriginal\n")
}

func TestArrayElementVar_NestedAssociativeSlot(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
type TItems = array [2..3] of Integer;
var a: array [String] of TItems;
procedure Mutate(var value: Integer);
begin value := 42; end;
Mutate(a['new'][2]);
PrintLn(a['new'][2]);
PrintLn(a.Length);
`), "42\n1\n")
}

func TestArrayElementVar_CallRoutes(t *testing.T) {
	for _, tt := range []struct{ name, declarations, call string }{
		{"method", `type TTarget = class
  procedure Mutate(var value: Integer);
  begin value := 42; end;
end;
var target := TTarget.Create;`, "target.Mutate(a[NextIndex()]);"},
		{"method overload", `type TTarget = class
  procedure Mutate(var value: Integer); overload;
  begin value := 42; end;
  procedure Mutate(value: String); overload;
  begin PrintLn('wrong overload'); end;
end;
var target := TTarget.Create;`, "target.Mutate(a[NextIndex()]);"},
		{"local function", `procedure Outer;
begin
  procedure Mutate(var value: Integer);
  begin value := 42; end;
  Mutate(a[NextIndex()]);
end;`, "Outer();"},
		{"implicit Self", `type TTarget = class
  procedure Mutate(var value: Integer);
  begin value := 42; end;
  procedure Run;
  begin Mutate(a[NextIndex()]); end;
end;
var target := TTarget.Create;`, "target.Run();"},
		{"implicit Self overload", `type TTarget = class
  procedure Mutate(var value: Integer); overload;
  begin value := 42; end;
  procedure Mutate(value: String); overload;
  begin PrintLn('wrong overload'); end;
  procedure Run;
  begin Mutate(a[NextIndex()]); end;
end;
var target := TTarget.Create;`, "target.Run();"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, `
var a: array of Integer := [10];
function NextIndex: Integer;
begin PrintLn('index'); Result := 0; end;
`+tt.declarations+tt.call+` PrintLn(a[0]);`), "index\n42\n")
		})
	}
}

func TestArrayElementVar_LazyArgumentIsNotEvaluated(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var a: array of Integer := [10];
function NextIndex: Integer;
begin PrintLn('index'); Result := 0; end;
function Unused: Integer;
begin PrintLn('lazy argument'); Result := 0; end;
procedure Mutate(var value: Integer; lazy ignored: Integer);
begin value := 42; end;
Mutate(a[NextIndex()], Unused());
PrintLn(a[0]);
`), "index\n42\n")
}

func TestArrayElementVar_ImplicitSelfFailureStopsLaterArgument(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var a: array of Integer := [10];
function FailIndex: Integer;
begin PrintLn('index'); raise Exception.Create('index failed'); end;
function Later: Integer;
begin PrintLn('later argument'); Result := 0; end;
type TTarget = class
  procedure Mutate(var value: Integer; last: Integer);
  begin PrintLn('body'); value := 42; end;
  procedure Run;
  begin
    try Mutate(a[FailIndex()], Later());
    except on E: Exception do PrintLn(E.Message); end;
  end;
end;
var target := TTarget.Create;
target.Run();
PrintLn(a[0]);
`), "index\nindex failed\n10\n")
}

func TestArrayElementVar_RuntimeOverloadDiscriminator(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var a: array of Integer := [10];
var discriminator: Variant := 1;
function NextIndex: Integer;
begin PrintLn('index'); Result := 0; end;
procedure Mutate(var value: Integer; kind: Integer); overload;
begin PrintLn('integer'); value := 42; end;
procedure Mutate(var value: Integer; kind: String); overload;
begin PrintLn('string'); value := 43; end;
procedure Mutate(var value: Integer; kind: Variant); overload;
begin PrintLn('variant'); value := 44; end;
Mutate(a[NextIndex()], discriminator);
PrintLn(a[0]);
discriminator := 'text';
Mutate(a[NextIndex()], discriminator);
PrintLn(a[0]);
`), "index\nvariant\n44\nindex\nvariant\n44\n")
}

func TestArrayElementVar_RuntimeOverloadFailureStopsLaterArgument(t *testing.T) {
	got := runQuickwinScript(t, `
var a: array of Integer := [10];
function Later: Variant;
begin PrintLn('later argument'); Result := 1; end;
procedure Mutate(var value: Integer; kind: Integer); overload;
begin PrintLn('body'); end;
procedure Mutate(var value: Integer; kind: Variant); overload;
begin PrintLn('variant body'); end;
try Mutate(a[a.Length], Later());
except on E: Exception do PrintLn(E.Message); end;
PrintLn(a[0]);
`)
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "Upper bound exceeded! Index 1") || lines[1] != "10" {
		t.Fatalf("expected original bounds error and unchanged array without later argument or body, got %q", got)
	}
}

func TestArrayElementVar_RuntimeOverloadKeepsStaleReference(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var a: array of Integer := [10];
function ResizeArray: Variant;
begin PrintLn('resize'); a.Clear; Result := 1; end;
procedure Mutate(var value: Integer; kind: Variant); overload;
begin
  PrintLn('body');
  try PrintLn(value);
  except on E: Exception do PrintLn(E.Message); end;
end;
procedure Mutate(var value: Integer; kind: String); overload;
begin PrintLn('string body'); end;
Mutate(a[0], ResizeArray());
PrintLn(a.Length);
`), "resize\nbody\nUpper bound exceeded! Index 0\n0\n")
}

func TestArrayElementVar_RuntimeOverloadAssociativeOwnership(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
type TIntegers = array [0..0] of Integer;
type TStrings = array [0..0] of String;
var ints: array [String] of TIntegers;
var strs: array [String] of TStrings;
var discriminator: Variant := 1;
procedure Mutate(var value: Integer; kind: Variant); overload;
begin value := 42; end;
procedure Mutate(value: String; kind: Variant); overload;
begin value := 'local'; PrintLn(value); end;
Mutate(strs['missing'][0], discriminator);
PrintLn(strs.Length);
Mutate(ints['missing'][0], discriminator);
PrintLn(ints.Length);
PrintLn(ints['missing'][0]);
`), "local\n0\n1\n42\n")
}

func TestArrayElementVar_RuntimeOverloadAssociativeKeyEvaluatedOnce(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var a: array [String] of Integer;
var calls: Integer;
var discriminator: Variant := 'text';
function Key: String;
begin Inc(calls); Result := 'k' + IntToStr(calls); end;
procedure Mutate(var value: Integer; kind: String); overload;
begin PrintLn('var'); value := 42; end;
procedure Mutate(value: Variant; kind: Integer); overload;
begin PrintLn('value'); end;
a['k1'] := 1;
Mutate(a[Key()], discriminator);
PrintLn(calls);
PrintLn(a.Length);
PrintLn(a['k1']);
`), "var\n1\n1\n42\n")
}
