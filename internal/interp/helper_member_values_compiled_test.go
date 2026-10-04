package interp

import "testing"

func TestHelperMemberValues_CompiledReference(t *testing.T) {
	for _, tt := range []struct{ name, target, setup, body string }{
		{"primitive", "Integer", "var item: T := 5;", "Result := Self + v;"},
		{"record", "record x: Integer; end", "var item: T; item.x := 5;", "Result := Self.x + v;"},
		{"class", "class x: Integer; end", "var item := T.Create; item.x := 5;", "Result := Self.x + v;"},
		{"interface", "interface end", "var item: T;", "Result := v + 5;"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "type T = " + tt.target + ";\ntype H = helper for T function Take(v: Integer): Integer; begin PrintLn('called'); " + tt.body + " end; end;\n" + tt.setup + "\nvar p: function(v: Integer): Integer := item.Take;\nPrintLn('captured'); PrintLn(p(2));"
			compileAndRunWithHelperTransfer(t, source, "helper_member_reference.dws", "captured\ncalled\n7\n")
		})
	}
}

func TestHelperMemberValues_CompiledParameterlessReference(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type H = helper for Integer
 function Take: Integer; begin PrintLn('called'); Result := Self + 1; end;
end;
var item := 5;
var p: function: Integer := item.Take;
PrintLn('captured');
PrintLn(p());
`, "helper_parameterless_reference.dws", "captured\ncalled\n6\n")
}

func TestHelperMemberValues_CompiledCaptureAndReturnedCallable(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type TProc = procedure;
procedure Inner; begin PrintLn('inner'); end;
type H = helper for Integer
 function Factory: TProc; begin PrintLn('factory'); Result := @Inner; end;
 function Take(v: Integer): Integer; begin PrintLn('called'); Result := Self + v; end;
end;
function Receiver: Integer; begin PrintLn('receiver'); Result := 5; end;
var p: function(v: Integer): Integer := Receiver().Take;
PrintLn('captured'); PrintLn(p(2));
var factoryResult: TProc := (5).Factory;
PrintLn('selected'); factoryResult();
((5).Factory)();
`, "helper_capture_factory.dws", "receiver\ncaptured\ncalled\n7\nfactory\nselected\ninner\nfactory\ninner\n")
}

func TestHelperMemberValues_CompiledExplicitReference(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type H = helper for Integer
 function Take(v: Integer): Integer; begin PrintLn('called'); Result := Self + v; end;
end;
var p: function(s: Integer; v: Integer): Integer := H.Take;
PrintLn('captured'); PrintLn(p(5, 2));
`, "helper_explicit_reference.dws", "captured\ncalled\n7\n")
}

func TestHelperMemberValues_CompiledBodyReference(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type H = helper for Integer
 function Take(v: Integer): Integer; begin PrintLn('called'); Result := Self + v; end;
 procedure Run;
 begin
  var p: function(v: Integer): Integer := Take;
  PrintLn('captured'); PrintLn(p(2));
 end;
end;
(5).Run();
`, "helper_body_reference.dws", "captured\ncalled\n7\n")
}

func TestHelperMemberValues_CompiledOverridesIntrinsicReference(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type T = class end;
type H = helper for T
 function ClassName: String; begin PrintLn('called'); Result := 'helper'; end;
end;
var item := T.Create;
var p: function: String := item.ClassName;
PrintLn('captured'); PrintLn(p());
`, "helper_intrinsic_reference.dws", "captured\ncalled\nhelper\n")
}

func TestHelperMemberValues_CompiledFunctionHelperReference(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
function Take(s: String; v: Integer): String; helper;
begin PrintLn('called'); Result := s + IntToStr(v); end;
var p: function(v: Integer): String := ('x').Take;
PrintLn('captured'); PrintLn(p(2));
`, "function_helper_reference.dws", "captured\ncalled\nx2\n")
}

func TestHelperMemberValues_ReviewRegressions(t *testing.T) {
	for _, tt := range []struct{ name, source, output string }{
		{"grouped ordinary reference", `function F: Integer; begin PrintLn('called'); Result := 7; end;
var p: function: Integer := (F); PrintLn('captured'); PrintLn(p());`, "captured\ncalled\n7\n"},
		{"explicit selected overload", `type H = helper for Integer
function Take(v: Integer): Integer; begin Result := Self + v; end;
function Take(v, w: Integer): Integer; overload; begin Result := Self + v + w; end;
end;
var p: function(s, v: Integer): Integer := H.Take; PrintLn(p(5, 2));`, "7\n"},
		{"explicit lazy argument", `type H = helper for Integer
function Take(lazy v: Integer): Integer; begin Result := v + v; end; end;
var n := 0;
function Count: Integer; begin n += 1; Result := n; end;
var p: function(s: Integer; lazy v: Integer): Integer := H.Take;
PrintLn(p(5, Count())); PrintLn(n);`, "3\n2\n"},
		{"reference retains receiver", `type T = class Name: String; destructor Destroy; override; end;
destructor T.Destroy; begin PrintLn('destroyed'); inherited; end;
type H = helper for T function Take: String; begin Result := Self.Name; end; end;
var obj := T.Create; obj.Name := 'kept';
var p: function: String := obj.Take;
obj := nil; PrintLn('cleared'); PrintLn(p());`, "cleared\nkept\n"},
		{"reference uses dispatch-order helper", `type T = class end;
type H1 = helper for T function Name: String; begin Result := 'H1'; end; end;
type H2 = helper for T function Name: String; begin Result := 'H2'; end; end;
var item := T.Create;
var p: function: String := item.Name; PrintLn(item.Name); PrintLn(p());`, "H1\nH1\n"},
		{"grouped constant", `const A = 1; const B = (A); const C = (A) + 1;
var arr: array [0..(B)] of Integer; PrintLn(B); PrintLn(C); PrintLn(High(arr));`, "1\n2\n1\n"},
	} {
		t.Run(tt.name, func(t *testing.T) { compileAndRunWithHelperTransfer(t, tt.source, "helper_review.dws", tt.output) })
	}
}
