package interp

import (
	"bytes"
	"testing"
)

// Ensure the default Destroy/Free lifecycle works like DWScript TObject.
func TestFreeDestroyLifecycle(t *testing.T) {
	input := `
type TMyObj = class
  destructor Destroy; override;
end;

destructor TMyObj.Destroy;
begin
  PrintLn('destroy');
  inherited destroy;
end;

var o: TMyObj;
begin
  o := TMyObj.Create;
  o.Free;
  try
    o.Free;
  except
    on e: Exception do PrintLn(e.Message);
  end;

  o := nil;
  o.Free;
end.
`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "destroy\nObject already destroyed [line: 17, column: 6]\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}
