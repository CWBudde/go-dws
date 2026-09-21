package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
)

func TestInterfaceImplements_ClassReferences(t *testing.T) {
	source := `
type IBase = interface end;
type IDerived = interface(IBase) end;
type TSupported = class(TObject, IDerived) end;
type TChild = class(TSupported) end;
type TRef = class of TObject;
type TAlias = TSupported;
type IAlias = IDerived;
var ref: TRef := TChild;
var empty: TRef;
PrintLn(TSupported implements IDerived);
PrintLn(TChild implements IDerived);
PrintLn(TAlias implements IDerived);
PrintLn(ref implements IDerived);
PrintLn(TObject implements IDerived);
PrintLn(empty implements IDerived);
PrintLn(nil implements IDerived);
// implements uses explicit declarations, including those on ancestor classes.
PrintLn(TSupported implements IBase);
var obj := TChild.Create;
PrintLn(obj implements IDerived);
PrintLn(TAlias implements IAlias);
`
	assertOutput(t, runQuickwinScript(t, source), "True\nTrue\nTrue\nTrue\nFalse\nFalse\nFalse\nFalse\nTrue\nTrue\n")
}

func TestInterfaceImplements_RejectsInvalidOperands(t *testing.T) {
	for _, tc := range []struct {
		name, source, diagnostic string
	}{
		{"scalar", "type ITest = interface end; PrintLn(42 implements ITest);", "requires class"},
		{"target", "PrintLn(TObject implements Integer);", "requires interface type"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertCompileError(t, tc.source, tc.diagnostic)
		})
	}
}

func TestInterfaceCompatibility_Fixtures(t *testing.T) {
	const category = "InterfacesPass"
	for _, name := range []string{
		"interface_cast_to_obj", "interface_multiple_cast", "interface_nil_cast_from_intf",
		"interface_nil_cast_from_obj",
		"intf_casts", "intf_compare", "interface_implements_intf",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, category, name+".pas")
			if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel(category)); got != testResultPassed {
				t.Fatalf("%s/%s: %v: %s", category, name, got, detail)
			}
		})
	}
}
