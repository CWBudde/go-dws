package interp

import "testing"

func TestClassArity_DefaultsThroughCompilePath(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type B = class
  procedure Take(v: Integer = 7); begin PrintLn(v); end;
  class procedure StaticTake(v: Integer = 8); begin PrintLn(v); end;
end;
type C = class(B)
  procedure Run;
  begin
    Take();
    Self.Take();
    inherited Take();
    StaticTake();
    inherited StaticTake();
  end;
end;
var obj := C.Create;
obj.Run();
`, "class_arity_defaults.dws", "7\n7\n7\n8\n8\n")
}

func TestClassArity_OverloadSelectionThroughCompilePath(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type B = class
  procedure Pick(v: Integer); overload; begin PrintLn(v); end;
  procedure Pick(v: String); overload; begin PrintLn(v); end;
end;
type C = class(B)
  procedure Run;
  begin
    inherited Pick(1);
    inherited Pick('s');
    Pick(2);
    Self.Pick('t');
  end;
end;
var obj := C.Create;
obj.Run();
type T = class
  constructor Create(v: Integer = 9); begin PrintLn(v); end;
end;
var first := new T();
var second := T.Create(3);
var cls: class of T := T;
var third := cls.Create();

`, "class_arity_selection.dws", "1\ns\n2\nt\n9\n3\n9\n")
}

func TestClassArity_HiddenVarConstructorThroughExpressionReceiver(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type P = class
  constructor Create(var i: Integer); begin i := 42; PrintLn('parent'); end;
end;
type C = class(P)
  constructor Create(v: Float); begin PrintLn('child'); end;
end;
type CC = class of C;
function GetCls: CC;
begin Result := C; end;
var n := 1;
var o := GetCls().Create(n);
PrintLn(n);
`, "class_arity_hidden_var_constructor.dws", "child\n1\n")
}
