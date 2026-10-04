package interp

import (
	"bytes"
	"testing"
)

func TestHelperOverloadExecution_UncheckedSelection(t *testing.T) {
	var output bytes.Buffer
	engine := New(&output)
	result := interpret(engine, `type H = helper for Integer
function Pick(v: Integer): Integer; overload; begin Result := 11; end;
function Pick(v: String): Integer; overload; begin Result := 12; end;
end;
type J = helper(H) for Integer end;
var item: Integer;
PrintLn(item.Pick(1)); PrintLn(item.Pick('x')); PrintLn(J.Pick(item, 1));`)
	if isError(result) {
		t.Fatalf("evaluation failed: %s", result.String())
	}
	if got, want := output.String(), "11\n12\n11\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestHelperOverloadExecution_CompiledSelection(t *testing.T) {
	for _, target := range []struct{ name, declaration, helperTarget, setup string }{
		{"primitive", "Integer", "Integer", "var item: T := 5;"},
		{"record", "record x: Integer; end", "T", "var item: T;"},
		{"class", "class end", "T", "var item := T.Create;"},
		{"interface", "interface end", "T", "type C = class(TObject, T) end; var item: T := C.Create;"},
	} {
		for _, reverse := range []bool{false, true} {
			name := target.name + "/integer_first"
			integer := "function Pick(v: Integer): String; overload; begin Result := 'integer'; end;\n"
			str := "function Pick(v: String): String; overload; begin Result := 'string'; end;\n"
			methods := integer + str
			if reverse {
				name = target.name + "/string_first"
				methods = str + integer
			}
			t.Run(name, func(t *testing.T) {
				source := "type T = " + target.declaration + ";\n" +
					"type H = helper for " + target.helperTarget + "\n" + methods +
					"function Run: String; begin Result := Pick(1) + ':' + Pick('x'); end;\nend;\n" +
					target.setup + "\nPrintLn(item.Pick(1)); PrintLn(item.Pick('x'));\n" +
					"PrintLn(H.Pick(item, 1)); PrintLn(H.Pick(item, 'x')); PrintLn(item.Run());"
				compileAndRunWithHelperTransfer(t, source, "helper_overload_execution.dws", "integer\nstring\ninteger\nstring\ninteger:string\n")
			})
		}
	}
}

func TestHelperOverloadExecution_CompiledContracts(t *testing.T) {
	for _, tt := range []struct{ name, source, output string }{
		{"inherited set and declaring storage", `type H = helper for Integer
class var Count: Integer;
function Pick(v: Integer): Integer; overload; begin Count += 1; Result := 11; end;
function Pick(v: String): Integer; overload; begin Count += 10; Result := 12; end;
end;
type J = helper(H) for Integer class var Count: Integer; end;
type K = helper(J) for Integer end;
var item: Integer;
PrintLn(item.Pick(1)); PrintLn(K.Pick(item, 1)); PrintLn(H.Pick(item, 1));
PrintLn(item.Pick('x')); PrintLn(K.Pick(item, 'x')); PrintLn(H.Pick(item, 'x'));
PrintLn(H.Count); PrintLn(J.Count);`, "11\n11\n11\n12\n12\n12\n33\n0\n"},
		{"static Variant selection", `type H = helper for Integer
function Pick(v: String): String; overload; begin Result := 'string'; end;
function Pick(v: Variant): String; overload; begin Result := 'variant'; end;
function Run(v: Variant): String; begin Result := Pick(v); end;
end;
var item: Integer; var v: Variant := 'x';
PrintLn(item.Pick(v)); PrintLn(H.Pick(item, v)); PrintLn(item.Run(v));`, "variant\nvariant\nvariant\n"},
		{"lazy argument timing", `var n := 0;
function Count: Integer; begin n += 1; Result := n; end;
type H = helper for Integer
function Pick(lazy v: Integer): Integer; overload; begin Result := v + v; end;
function Pick(v: String): Integer; overload; begin Result := 99; end;
function Run: Integer; begin Result := Pick(Count()); end;
end;
var item: Integer;
PrintLn(item.Pick(Count())); PrintLn(n);
PrintLn(H.Pick(item, Count())); PrintLn(n);
PrintLn(item.Run()); PrintLn(n);`, "3\n2\n7\n4\n11\n6\n"},
		{"receiver and arguments evaluated once", `var n := 0;
function Receiver: Integer; begin PrintLn('receiver'); Result := 5; end;
function Arg: String; begin n += 1; PrintLn('argument'); Result := 'x'; end;
type H = helper for Integer
function Pick(v: Integer): String; overload; begin Result := 'integer'; end;
function Pick(v: String): String; overload; begin Result := v; end;
end;
PrintLn(Receiver().Pick(Arg())); PrintLn(n);`, "receiver\nargument\nx\n1\n"},
		{"lexical helper body owner", `type H1 = helper for Integer
function Pick(v: Integer): String; overload; begin Result := 'H1 integer'; end;
function Pick(v: String): String; overload; begin Result := 'H1 string'; end;
end;
type H2 = helper for Integer
function Pick(v: Integer): String; overload; begin Result := 'H2 integer'; end;
function Pick(v: String): String; overload; begin Result := 'H2 string'; end;
function Run: String; begin Result := Pick('x'); end;
end;
var item: Integer;
PrintLn(item.Pick('x')); PrintLn(H2.Run(item));`, "H1 string\nH2 string\n"},
		{"out of line declarations", `type H = helper for Integer
function Pick(v: Integer): String; overload;
function Pick(v: String): String; overload;
end;
function H.Pick(v: Integer): String; begin Result := 'integer'; end;
function H.Pick(v: String): String; begin Result := 'string'; end;
var item: Integer;
PrintLn(item.Pick(1)); PrintLn(item.Pick('x')); PrintLn(H.Pick(item, 'x'));`, "integer\nstring\nstring\n"},
		{"function helper Self offset", `function Take(s: String; lazy v: Integer): Integer; helper;
begin Result := v + v; end;
var n := 0;
function Count: Integer; begin n += 1; Result := n; end;
PrintLn(('x').Take(Count())); PrintLn(n);`, "3\n2\n"},
		{"class and static roles", `type T = class end;
type H = helper for T
class function Pick(v: Integer): String; overload; begin Result := 'class integer'; end;
class function Pick(v: String): String; overload; begin Result := 'class string'; end;
class function StaticPick(v: Integer): String; overload; static; begin Result := 'static integer'; end;
class function StaticPick(v: String): String; overload; static; begin Result := 'static string'; end;
end;
PrintLn(T.Pick('x')); PrintLn(H.Pick(T, 1));
PrintLn(T.StaticPick('x')); PrintLn(H.StaticPick(1));`, "class string\nclass integer\nstatic string\nstatic integer\n"},
		{"child local name hides parent", `type H = helper for Integer
function Pick(v: Integer): Integer; overload; begin Result := 11; end;
function Pick(v: String): Integer; overload; begin Result := 12; end;
end;
type J = helper(H) for Integer
function Pick(v: Integer): Integer; begin Result := 21; end;
end;
var item: Integer;
PrintLn(item.Pick(1)); PrintLn(J.Pick(item, 1)); PrintLn(H.Pick(item, 1));`, "21\n21\n11\n"},
		{"precedence retains selected result type", `type H1 = helper for Integer
function Pick(v: Integer): String; begin Result := 'first'; end;
end;
type H2 = helper for Integer
function Pick(v: Integer): Integer; begin Result := 2; end;
end;
var item: Integer; var text: String := item.Pick(1); PrintLn(text);`, "first\n"},
		{"inline body retains earlier signature", `type H = helper for Integer
function Pick(v: Variant): String; overload; begin Result := 'variant'; end;
function Run: String; begin Result := Pick(1); end;
function Pick(v: Integer): String; overload; begin Result := 'integer'; end;
end;
var item: Integer; PrintLn(item.Run()); PrintLn(item.Pick(1));`, "variant\ninteger\n"},
		{"selected var parameters", `type H = helper for Integer
procedure Take(var v: Integer); overload; begin v += 1; end;
procedure Take(v: String); overload; begin end;
procedure Run(var v: Integer); begin Take(v); end;
end;
var item: Integer; var n := 1;
item.Take(n); H.Take(item, n); item.Run(n); PrintLn(n);`, "4\n"},
		{"alias-specific implicit Self", `type A = Integer;
type H = strict helper for A
function Pick(v: Integer): Integer; overload; begin Result := 11; end;
function Pick(v: String): Integer; overload; begin Result := 12; end;
function Run: Integer; begin Result := Pick('x'); end;
end;
var item: A; PrintLn(item.Pick('x')); PrintLn(item.Run());`, "12\n12\n"},
		{"type receiver uses first owner's class role", `type H1 = helper for Integer
class function Pick(v: Integer): String; begin Result := 'first'; end;
end;
type H2 = helper for Integer
function Pick(v: Integer): String; begin Result := 'second'; end;
end;
PrintLn(Integer.Pick(1));`, "first\n"},
		{"function helper labels retain declaration identity", `function StringLabel(s: String): String; helper Label;
begin Result := 'string:' + s; end;
function IntegerLabel(i: Integer): String; helper Label;
begin Result := 'integer:' + IntToStr(i); end;
PrintLn(('x').Label()); PrintLn((5).Label());`, "string:x\ninteger:5\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			compileAndRunWithHelperTransfer(t, tt.source, "helper_overload_contract.dws", tt.output)
		})
	}
}
