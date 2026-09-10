package interp

import (
	"testing"
)

// ============================================================================
// Stack-frame positions (PLAN.md §3.3)
//
// A frame is positioned at the *name token* of the thing being called, and a
// method frame is class-qualified. The bottom frame's label is empty: its call
// site lies in the main program.
// ============================================================================

// TestStackTrace_RaiseSiteIsConstructorName pins the raise site to the `Create`
// identifier rather than the class name or the closing paren.
func TestStackTrace_RaiseSiteIsConstructorName(t *testing.T) {
	runScriptTest(t, `
procedure ThisOneBombs;
begin
   raise Exception.Create('boom!');
end;

try
   ThisOneBombs;
except
   on E: Exception do
      PrintLn(E.StackTrace);
end;
`, "ThisOneBombs [line: 4, column: 20]\n [line: 8, column: 4]")
}

// TestStackTrace_MethodCallSiteIsMemberName pins a method frame to the member
// name, not the receiver, and checks the class qualifier on a method declared
// inline in the class body.
func TestStackTrace_MethodCallSiteIsMemberName(t *testing.T) {
	runScriptTest(t, `
type TMyClass = class
   procedure Boom;
   begin
      raise Exception.Create('bang');
   end;
end;

try
   var c := TMyClass.Create;
   c.Boom;
except
   on E: Exception do
      PrintLn(E.StackTrace);
end;
`, "TMyClass.Boom [line: 5, column: 23]\n [line: 11, column: 6]")
}

// TestStackTrace_ContractFailureHasNoRaiseFrame checks that a failed
// pre-condition contributes no innermost frame: the routine and condition are
// already named in the message text.
func TestStackTrace_ContractFailureHasNoRaiseFrame(t *testing.T) {
	runScriptTest(t, `
procedure RequirePositive(a : Integer);
require a > 0 : 'here';
begin
end;

procedure TestProc;
begin
   RequirePositive(-1);
end;

try
   TestProc;
except
   on E : Exception do
      PrintLn(E.StackTrace);
end;
`, "TestProc [line: 9, column: 4]\n [line: 13, column: 4]")
}

// TestExceptObject_StackTraceOutsideExceptBlock checks that DWScript's magic
// StackTrace getter answers ” on the nil ExceptObject instead of raising
// "Object not instantiated".
func TestExceptObject_StackTraceOutsideExceptBlock(t *testing.T) {
	runScriptTest(t, `
PrintLn(ExceptObject = nil);
PrintLn('[' + ExceptObject.StackTrace + ']');
`, "True\n[]")
}
