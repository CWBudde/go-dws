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

// TestNilReceiverVarParamBindsByRef verifies that a non-virtual method
// dispatched on a nil receiver still binds var parameters by reference:
// DWScript resolves parameter modes statically from the receiver's declared
// type, so a nil Self must not degrade the call to by-value argument passing.
func TestNilReceiverVarParamBindsByRef(t *testing.T) {
	input := `
type TMyObj = class
   procedure SetX(var x : Integer);
end;
procedure TMyObj.SetX(var x : Integer);
begin
   x := 42;
end;
var o : TMyObj;
var v := 1;
o.SetX(v);
PrintLn(v);
`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	if !strings.Contains(output, "42") {
		t.Errorf("expected var parameter write-back through nil dispatch (42), got:\n%s", output)
	}
}

// TestRefcountReclaimedObjectMethodCallStillWorks verifies that method calls on
// an object whose destructor ran through automatic refcount cleanup (not an
// explicit Free/Destroy) still succeed, consistent with the member-access path:
// go-dws can reclaim eagerly where DWScript keeps objects alive, so only
// explicit destruction makes later use an error.
func TestRefcountReclaimedObjectMethodCallStillWorks(t *testing.T) {
	input := `
type TList = class
   Field : array of String;
   function GetCount : Integer;
end;
function TList.GetCount : Integer;
begin
   Result := Field.Length;
end;
type TContainer = class
   A : array of TList;
end;
var c := new TContainer;
var i := new TList;
i.Field.Add('a', 'b');
c.A.Add(i);
i := new TList;
i.Field.Add('c', 'd', 'e');
c.A.Add(i);
for var list in c.A do
   PrintLn(list.GetCount);
`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	if !strings.Contains(output, "2") || !strings.Contains(output, "3") {
		t.Errorf("expected method calls on refcount-reclaimed objects to work (2, 3), got:\n%s", output)
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
