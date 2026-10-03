package interp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These runtime controls are derived from upstream's type-directed AS caster:
// the target expression is discarded after its declared type selects the cast.
func TestAsCast_DeclaredTargetDoesNotExecute(t *testing.T) {
	const declarations = `
type TBase = class
  function Name: String; virtual; begin Result := 'base'; end;
end;
type TChild = class(TBase)
  function Name: String; override; begin Result := 'child'; end;
end;
type TBaseClass = class of TBase;
var Calls := 0;
function Target: TBaseClass;
begin Inc(Calls); Result := TBase; end;
var Obj: TObject := TChild.Create;
var Meta: TBaseClass := TChild;
type THolder = record Meta: TBaseClass; end;
var Holder: THolder;
Holder.Meta := TChild;
`
	for _, rhs := range []string{"Meta", "Target", "Target()", "(Target())", "Holder.Meta"} {
		t.Run(rhs, func(t *testing.T) {
			source := declarations + "var Cast := Obj as " + rhs + ";\nPrintLn(Cast.Name);\nPrintLn(Calls);\nPrintLn((TChild as " + rhs + ") = TChild);\nPrintLn(Calls);"
			assertOutput(t, runQuickwinScript(t, source), "child\n0\nTrue\n0\n")
		})
	}
}

func TestIsCheck_DeclaredTargetDoesNotExecute(t *testing.T) {
	const declarations = `
type TBase = class end;
type TChild = class(TBase) end;
type TBaseClass = class of TBase;
var Calls := 0;
function Target: TBaseClass;
begin Inc(Calls); Result := TChild; end;
var Obj: TObject := TBase.Create;
var Meta: TBaseClass := TChild;
type THolder = record Meta: TBaseClass; end;
var Holder: THolder;
Holder.Meta := TChild;
`
	for _, rhs := range []string{"Meta", "Target", "Target()", "(Target())", "Holder.Meta"} {
		t.Run(rhs, func(t *testing.T) {
			source := declarations + "PrintLn(Obj is " + rhs + ");\nPrintLn(Calls);"
			assertOutput(t, runQuickwinScript(t, source), "True\n0\n")
		})
	}
}

func TestAsCast_SimpleFixtureControls(t *testing.T) {
	for _, name := range []string{"class_cast_meta", "class_of_cast", "variants_as_casts"} {
		t.Run(name, func(t *testing.T) {
			base := filepath.Join("..", "..", "testdata", "fixtures", "SimpleScripts", name)
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

func TestIsCheck_BooleanValueTargets(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `var Left := true; var Right := false;
PrintLn(Left is Right);
PrintLn(Right is Right);
PrintLn(Left is True);
PrintLn(Right is False);`), "False\nTrue\nTrue\nTrue\n")
}

func TestIsCheck_VariantBooleanTargets(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `type TV = Variant; type TB = Boolean;
var V: TV := false;
var B: TB := false;
PrintLn(V is False); PrintLn(V is True); PrintLn(V is B);
V := true; B := true;
PrintLn(V is False); PrintLn(V is True); PrintLn(V is B);`), "True\nFalse\nTrue\nFalse\nTrue\nTrue\n")
}

func TestIsCheck_BooleanCallableIdentity(t *testing.T) {
	for _, tt := range []struct{ name, declarations, reference string }{
		{"function pointer", `type TCheck = function: Boolean;
var Calls := 0;
function Check: Boolean; begin Inc(Calls); Result := false; end;
var P: TCheck := @Check;`, "P"},
		{"bound method", `var Calls := 0;
type TChecker = class
function Check: Boolean; begin Inc(Calls); Result := false; end;
end;
var Obj := TChecker.Create;`, "Obj.Check"},
		{"method pointer", `type TCheck = function: Boolean of object;
var Calls := 0;
type TChecker = class
function Check: Boolean; begin Inc(Calls); Result := false; end;
end;
var Obj := TChecker.Create;
var P: TCheck := @Obj.Check;`, "P"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := tt.declarations + "\nPrintLn(false is " + tt.reference + "); PrintLn(Calls);\nPrintLn(false is (" + tt.reference + ")); PrintLn(Calls);"
			assertOutput(t, runQuickwinScript(t, source), "True\n1\nTrue\n2\n")
		})
	}
}

func TestCastCheck_QualifiedTypeTargets(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `type TOuter = class
 type TInner = class end;
end;
var Obj: TOuter.TInner;
PrintLn(Obj is TOuter.TInner);
PrintLn((Obj as TOuter.TInner) = Obj);
PrintLn(Obj is System.TObject);
PrintLn((Obj as Internal.TObject) = Obj);`), "False\nTrue\nFalse\nTrue\n")
}
