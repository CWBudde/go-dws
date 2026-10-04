package frontend

import "testing"

func TestCompile_HelperOverloadOwnerDoesNotFallThrough(t *testing.T) {
	assertHelperCallDiagnostics(t, `type H1 = helper for Integer
procedure Take(v: Integer); overload; begin end;
end;
type H2 = helper for Integer
procedure Take(v: String); overload; begin end;
end;
var item: Integer;
item.Take('x');`, []string{`Syntax Error: There is no overloaded version of "Take" that can be called with these arguments [line: 8, column: 6]`})
}

func TestCompile_HelperTypeReceiverUsesDeclaringOwner(t *testing.T) {
	assertHelperCallDiagnostics(t, `type H1 = helper for Integer
function Pick(v: Integer): String; begin Result := 'first'; end;
end;
type H2 = helper for Integer
class function Pick(v: Integer): String; begin Result := 'second'; end;
end;
PrintLn(Integer.Pick(1));`, []string{`Syntax Error: Class method or constructor expected [line: 7, column: 17]`})
}
