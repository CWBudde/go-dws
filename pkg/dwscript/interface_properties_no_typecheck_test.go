package dwscript

import (
	"bytes"
	"testing"
)

// TestInterfaceIndexedProperties_WithoutTypeCheck guards the runtime-only path:
// with no semantic info, the index contract comes from the interface metadata
// held by the receiver variable.
func TestInterfaceIndexedProperties_WithoutTypeCheck(t *testing.T) {
	var output bytes.Buffer
	engine, err := New(WithTypeCheck(false), WithOutput(&output))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	_, err = engine.Eval(`
type TInts = array of Integer;
type IGrid = interface
  function GetCell(x, y: Integer): Integer;
  procedure SetCell(x, y, v: Integer);
  function GetRow(x, y: Integer): TInts;
  property Cells[x, y: Integer]: Integer read GetCell write SetCell; default;
  property Rows[x, y: Integer]: TInts read GetRow;
end;
type TGrid = class(TObject, IGrid)
  value: Integer;
  function GetCell(x, y: Integer): Integer; begin Result := value + x * 10 + y; end;
  procedure SetCell(x, y, v: Integer); begin value := v; end;
  function GetRow(x, y: Integer): TInts; begin Result := [x, y, x + y]; end;
end;
var g: IGrid := TGrid.Create;
PrintLn(g[1, 2]);
PrintLn(g.Cells[3, 4]);
PrintLn(g.Rows[5, 6][2]);
g[0, 0] := 100;
g.Cells[0, 0] += 5;
PrintLn(g[0, 0]);
`)
	if err != nil {
		t.Fatalf("eval failed: %v", err)
	}
	if want := "12\n34\n11\n105\n"; output.String() != want {
		t.Errorf("output = %q, want %q", output.String(), want)
	}
}
