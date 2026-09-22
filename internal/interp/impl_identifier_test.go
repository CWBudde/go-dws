package interp

import "testing"

func TestImplIdentifier_ScopesAndMembers(t *testing.T) {
	const source = `
type TBox = class
   Impl: Integer;
end;
function Identity(Impl: Integer): Integer;
begin Result := iMpL; end;
var Impl := TBox.Create;
Impl.Impl := Identity(42);
PrintLn(impl.IMPL);
`
	assertOutput(t, runQuickwinScript(t, source), "42\n")
}
