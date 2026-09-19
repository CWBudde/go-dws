package interp

import "testing"

func TestArrayElementVar_MethodOverloadRetainsStaleReference(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var a: array of Integer := [10];
function ResizeArray: Integer;
begin a.SetLength(0); Result := 0; end;
type TTarget = class
  procedure Mutate(value: String; ignored: Integer); overload;
  begin PrintLn('wrong overload'); end;
  procedure Mutate(var value: Integer; ignored: Integer); overload;
  begin
    PrintLn('body');
    try PrintLn(value);
    except on E: Exception do PrintLn(E.Message); end;
  end;
end;
var target := TTarget.Create;
target.Mutate(a[0], ResizeArray());
`), "body\nUpper bound exceeded! Index 0\n")
}
