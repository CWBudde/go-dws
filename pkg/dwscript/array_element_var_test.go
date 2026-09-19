package dwscript

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestEngine_ArrayElementVarUnitQualified(t *testing.T) {
	dir := t.TempDir()
	unit := `unit Mutators;
interface
procedure Mutate(var value: Integer);
implementation
procedure Mutate(var value: Integer);
begin value := 42; end;
end.`
	if err := os.WriteFile(filepath.Join(dir, "Mutators.dws"), []byte(unit), 0600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	engine, err := New(WithUnitSearchPaths(dir), WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(`uses Mutators;
var a: array of Integer := [10];
function NextIndex: Integer;
begin PrintLn('index'); Result := 0; end;
Mutators.Mutate(a[NextIndex()]);
PrintLn(a[0]);`)
	if err != nil {
		t.Fatal(err)
	}
	result, err := engine.Run(program)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success || output.String() != "index\n42\n" {
		t.Fatalf("result=%+v output=%q", result, output.String())
	}
}
