package dwscript

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDebugBreak_Execution(t *testing.T) {
	for _, typeCheck := range []bool{true, false} {
		t.Run(fmt.Sprintf("typeCheck=%t", typeCheck), func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithTypeCheck(typeCheck), WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			_, err = engine.Eval(`
PrintLn('before');
DebugBreak;
debugbreak {comment} ();
procedure P;
begin
   DebugBreak();
   PrintLn('inside');
end;
P;
type TTest = class
   procedure DebugBreak(X: Integer);
   begin PrintLn(X); end;
end;
var Obj := TTest.Create;
Obj.DebugBreak(7);
for var I := 1 to 2 do begin
   DEBUGBREAK;
   PrintLn(I);
end;
PrintLn('after');
`)
			if err != nil {
				t.Fatalf("execution failed: %v", err)
			}
			if want := "before\ninside\n7\n1\n2\nafter\n"; output.String() != want {
				t.Fatalf("output = %q, want %q", output.String(), want)
			}
		})
	}
}
