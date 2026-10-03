package dwscript

import (
	"bytes"
	"testing"
)

func TestRecordMeta_CompiledProgramIsolation(t *testing.T) {
	var output bytes.Buffer
	engine, err := New(WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(`
type R = record class var Count: Integer; end; type A = R;
var m := A; m.Count += 1; PrintLn(R.Count);
`)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		output.Reset()
		if _, err := engine.Run(program); err != nil {
			t.Fatal(err)
		}
		if output.String() != "1\n" {
			t.Fatalf("run %d output = %q", i, output.String())
		}
	}
}
