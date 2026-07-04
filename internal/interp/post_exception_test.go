package interp

import (
	"strings"
	"testing"
)

// TestNilReceiverErrorsAreCatchable verifies DWScript's post-exception
// continuation semantics: a runtime error from touching a nil object
// (field write, method call) is a catchable exception, so execution continues
// with subsequent statements. Non-virtual methods still dispatch on nil (the
// error only surfaces when the body dereferences Self), while virtual methods
// raise "Object not instantiated" at the call site.
func TestNilReceiverErrorsAreCatchable(t *testing.T) {
	input := `
type
   TMyObj = class
      Field : Integer;
      procedure Proc;
      procedure ProcVirtual; virtual;
   end;

procedure TMyObj.Proc;
begin
   PrintLn('Proc');
   PrintLn(Field);
end;

procedure TMyObj.ProcVirtual;
begin
   PrintLn('ProcVirtual');
end;

var o : TMyObj;
o := nil;

try
   o.Field := 456;
except
   on E : Exception do PrintLn(E.Message);
end;

try
   o.Proc;
except
   on E : Exception do PrintLn(E.Message);
end;

try
   o.ProcVirtual;
except
   on E : Exception do PrintLn(E.Message);
end;

PrintLn('done');
`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}

	want := []string{
		"Object not instantiated",                // field write on nil
		"Proc",                                   // non-virtual method runs with Self=nil
		"Object not instantiated in TMyObj.Proc", // ...error only when Field is read
		"Object not instantiated",                // virtual method needs an instance
		"done",                                   // execution continued past every catch
	}
	for _, w := range want {
		if !strings.Contains(output, w) {
			t.Errorf("expected output to contain %q, got:\n%s", w, output)
		}
	}
}

// TestRaiseNilIsCatchable verifies that raising a nil exception reference
// raises a catchable "Object not instantiated" exception.
func TestRaiseNilIsCatchable(t *testing.T) {
	input := `
var e : Exception;
try
   raise e;
except
   on Ex: Exception do PrintLn(Ex.Message);
end;
PrintLn('after');
`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	if !strings.Contains(output, "Object not instantiated") {
		t.Errorf("expected 'Object not instantiated', got:\n%s", output)
	}
	if !strings.Contains(output, "after") {
		t.Errorf("expected execution to continue past the handler, got:\n%s", output)
	}
}

// TestExplicitlyFreedObjectAccessIsCatchable verifies that accessing a member
// of an explicitly freed object raises a catchable "Object already destroyed".
func TestExplicitlyFreedObjectAccessIsCatchable(t *testing.T) {
	input := `
type TMyObj = class
   Field : Integer;
end;
var o := TMyObj.Create;
o.Free;
try
   PrintLn(o.Field);
except
   on E : Exception do PrintLn(E.Message);
end;
PrintLn('after');
`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	if !strings.Contains(output, "Object already destroyed") {
		t.Errorf("expected 'Object already destroyed', got:\n%s", output)
	}
	if !strings.Contains(output, "after") {
		t.Errorf("expected execution to continue past the handler, got:\n%s", output)
	}
}
