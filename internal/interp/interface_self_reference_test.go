package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
)

func TestInterfaceSelfReference_Fixture(t *testing.T) {
	for _, name := range []string{"intf_self_ref", "intf_delegate"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, "InterfacesPass", name+".pas")
			if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel("InterfacesPass")); got != testResultPassed {
				t.Fatalf("%s: %v: %s", name, got, detail)
			}
		})
	}
}

func TestInterfaceSelfReference_MethodPointer(t *testing.T) {
	const source = `type INode = interface
   function Next: INode;
end;
type TGetNode = function: INode;
var calls := 0;
type TNode = class(TObject, INode)
   function Next: INode;
   begin
      Inc(calls);
      Result := Self;
   end;
end;
var node: INode := TNode.Create;
var nextNode: TGetNode := node.Next;
PrintLn(calls);
var copy: INode := nextNode();
PrintLn(calls);
PrintLn(copy = nil);`
	compileAndRunWithHelperTransfer(t, source, "interface_self_pointer.dws", "0\n1\nFalse\n")
}
