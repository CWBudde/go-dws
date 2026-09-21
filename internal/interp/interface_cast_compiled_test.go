package interp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInterfaceAsCast_Fixtures(t *testing.T) {
	for _, name := range []string{"interface_cast_to_obj", "interface_multiple_cast", "interface_nil_cast_from_intf", "interface_nil_cast_from_obj", "intf_casts"} {
		t.Run(name, func(t *testing.T) {
			base := filepath.Join("..", "..", "testdata", "fixtures", "InterfacesPass", name)
			source, err := os.ReadFile(base + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			assertOutput(t, strings.TrimSpace(runQuickwinScript(t, string(source))), strings.TrimSpace(string(want)))
		})
	}
}

func TestInterfaceAsCast_Runtime(t *testing.T) {
	const declarations = `
type IFirst = interface procedure First; end;
type ISecond = interface procedure Second; end;
type TBoth = class(TObject, IFirst, ISecond)
  procedure First; begin PrintLn('first'); end;
  procedure Second; begin PrintLn('second'); end;
end;
type TChild = class(TBoth) end;
`
	for _, tt := range []struct{ name, source, want string }{
		{"inherited identity and alias", `
type TAlias = TBoth;
type IAlias = ISecond;
var Obj: TObject := TChild.Create;
var FirstRef: IFirst := Obj as IFirst;
var SecondRef: IAlias := FirstRef as IAlias;
SecondRef.Second;
PrintLn((SecondRef as TAlias) = Obj);
PrintLn((SecondRef as TObject) = Obj);
`, "second\nTrue\nTrue\n"},
		{"root interface upcast", `
var FirstRef: IFirst := TBoth.Create;
var RootRef: IInterface := FirstRef;
(RootRef as IFirst).First;
PrintLn((RootRef as TObject) = (FirstRef as TObject));
`, "first\nTrue\n"},
		{"root interface alias assignment", `
type IFirstAlias = IFirst;
type IRootAlias = IInterface;
var FirstRef: IFirst := TBoth.Create;
var AliasRef: IFirstAlias := FirstRef;
var RootRef: IInterface := AliasRef;
var RootAliasRef: IRootAlias := FirstRef;
var BothAliasRef: IRootAlias := AliasRef;
(RootRef as IFirst).First;
(RootAliasRef as IFirst).First;
(BothAliasRef as IFirst).First;
`, "first\nfirst\nfirst\n"},
		{"explicit root cast round trip", `
type IRootAlias = IInterface;
var Obj := TChild.Create;
var FirstRef: IFirst := Obj;
var RootRef := FirstRef as IInterface;
(RootRef as IFirst).First;
PrintLn((RootRef as TObject) = Obj);
PrintLn(((Obj as IRootAlias) as TObject) = Obj);
PrintLn(Obj implements IInterface);
`, "first\nTrue\nTrue\nFalse\n"},
		{"plain object root cast rejected", `
try
  var RootRef := TObject.Create as IInterface;
  PrintLn('unreachable');
except
  on E: Exception do PrintLn(Pos('does not implement interface', E.Message) > 0);
end;
PrintLn(TObject implements IInterface);
`, "True\nFalse\n"},
		{"nil object and interface", `
var Obj: TObject;
var FirstRef: IFirst := Obj as IFirst;
PrintLn(FirstRef = nil);
PrintLn((FirstRef as ISecond) = nil);
PrintLn((nil as IFirst) = nil);
PrintLn((TObject(nil) as IFirst) = nil);
`, "True\nTrue\nTrue\nTrue\n"},
		{"nil object fixture equivalent", `
var Ref: IFirst;
var Obj: TObject;
Obj := nil;
Ref := Obj as IFirst;
if Ref = nil then PrintLn('Ok');
`, "Ok\n"},
		{"incompatible concrete object fails at runtime", `
var Calls := 0;
function MakeObject: TObject;
begin Inc(Calls); Result := TObject.Create; end;
try
  var Ref := MakeObject() as IFirst;
  PrintLn('unreachable');
except
  on E: Exception do PrintLn(Pos('does not implement interface', E.Message) > 0);
end;
PrintLn(Calls);
`, "True\n1\n"},
		{"single evaluation", `
var Calls := 0;
function MakeObject: TObject;
begin Inc(Calls); Result := TChild.Create; end;
var FirstRef := MakeObject() as IFirst;
FirstRef.First;
PrintLn(Calls);
`, "first\n1\n"},
		{"operand exception propagates", `
function Fail: TObject;
begin raise Exception.Create('original failure'); end;
try
  var FirstRef := Fail() as IFirst;
  PrintLn('unreachable');
except
  on E: Exception do PrintLn(E.Message);
end;
`, "original failure\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, declarations+tt.source), tt.want)
		})
	}
}

func TestInterfaceAsCast_CompileErrors(t *testing.T) {
	for _, tt := range []struct{ name, source, diagnostic string }{
		{"scalar operand", `type ITest = interface end; var Ref := 1 as ITest;`, "'as' operator requires"},
		{"scalar target", `var Obj: TObject; var Value := Obj as Integer;`, "requires class or interface type"},
		{"unknown target", `var Obj: TObject; var Value := Obj as MissingType;`, "cannot resolve target type"},
		{"unrelated classes", `type TA = class end; type TB = class end; var Obj := TA.Create; var Other := Obj as TB;`, "Incompatible types"},
		{"implicit class assignment", `type ITest = interface end; var Obj: TObject; var Ref: ITest; Ref := Obj;`, "Cannot assign"},
		{"implicit interface assignment", `type IA = interface end; type IB = interface end; var A: IA; var B: IB; B := A;`, "Cannot assign"},
		{"implicit root downcast", `type ITest = interface end; var Root: IInterface; var Ref: ITest; Ref := Root;`, "Cannot assign"},
		{"implicit alias root downcast", `type ITest = interface end; type IRootAlias = IInterface; var Root: IRootAlias; var Ref: ITest; Ref := Root;`, "Cannot assign"},
		{"implicit derived downcast", `type IBase = interface end; type IChild = interface(IBase) end; var Base: IBase; var Child: IChild; Child := Base;`, "Cannot assign"},
		{"incomplete implementation", `type ITest = interface procedure Work; end; type TTest = class(TObject, ITest) end;`, "does not implement interface method"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertCompileError(t, tt.source, tt.diagnostic)
		})
	}
}
