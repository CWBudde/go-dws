package interp

import "testing"

func TestArrayMapAndFind_CopyRecordValues(t *testing.T) {
	script := `
type TPoint = record
	X: Integer;
end;

var arr: array of TPoint;
var p: TPoint;
p.X := 1;
arr.Add(p);

var mapped := arr.Map(lambda(x: TPoint): TPoint => x);
var found := Find(arr, lambda(x: TPoint): Boolean => true);

arr[0].X := 99;

PrintLn(mapped[0].X);
PrintLn(found.X);
`

	result, output := testEvalWithOutput(script)
	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("unexpected error: %s", result.String())
	}

	const want = "1\n1\n"
	if output != want {
		t.Fatalf("expected output %q, got %q", want, output)
	}
}
