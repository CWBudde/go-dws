package dwscript

import (
	"bytes"
	"testing"
)

// TestExplicitHelperCalls_WithoutTypeCheck guards the runtime-only path: with
// no semantic info, the helper's own name binding must not be mistaken for a
// variable shadowing the helper.
func TestExplicitHelperCalls_WithoutTypeCheck(t *testing.T) {
	var output bytes.Buffer
	engine, err := New(WithTypeCheck(false), WithOutput(&output))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	_, err = engine.Eval(`
type TNumber = helper for Integer
   function Add(n: Integer): Integer; begin Result := Self + n; end;
   procedure Show; begin PrintLn(Self); end;
end;
PrintLn(TNumber.Add(40, 2));
TNumber.Show(7);
`)
	if err != nil {
		t.Fatalf("eval failed: %v", err)
	}
	if want := "42\n7\n"; output.String() != want {
		t.Errorf("output = %q, want %q", output.String(), want)
	}
}
