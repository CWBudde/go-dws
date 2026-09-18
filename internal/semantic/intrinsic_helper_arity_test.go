package semantic

import (
	"strings"
	"testing"
)

func TestIntrinsicHelperImplicitCallArity(t *testing.T) {
	for _, source := range []string{
		"var a: array of Float; a.Offset;",
		"var a: array of Float; a.Multiply;",
		"var a: array of Float; a.MultiplyAdd;",
		"var a: array of Float; a.Offset(1).Multiply;",
		"function Values: array of Float; begin end; Values.Offset;",
		"var n := 1; n.Compare;",
		"var n := 1; n.TestBit;",
	} {
		t.Run(source, func(t *testing.T) {
			diagnostics := analyzeWithHints(t, source, HintsLevelNormal)
			if len(diagnostics) != 1 || !strings.HasPrefix(diagnostics[0], "More arguments expected at ") {
				t.Fatalf("missing arity diagnostic: %v", diagnostics)
			}
		})
	}
}

func TestIntrinsicHelperImplicitCallArity_ReferencesAndOverrides(t *testing.T) {
	for _, source := range []string{
		"var a: array of Float; var callback := a.Offset;",
		"var n := 1; var callback := n.Compare;",
		"var a: array of Float; a.Reciprocal;",
		"var n := 1; n.PopCount;",
		`type TValues = array of Float;
type TCustom = helper for TValues
  procedure Offset(value: Float = 1); begin end;
end;
var a: TValues; a.Offset;`,
		`type TCustom = helper for Integer
  procedure Compare(value: Integer); overload; begin end;
  procedure Compare; overload; begin end;
end;
var n := 1; n.Compare;`,
		`type TRec = record procedure F; begin end; end;
type TCustom = helper for TRec procedure F(value: Integer); begin end; end;
var r: TRec; r.F;`,
	} {
		t.Run(source, func(t *testing.T) {
			diagnostics := analyzeWithHints(t, source, HintsLevelNormal)
			if len(diagnostics) != 0 {
				t.Fatalf("unexpected diagnostic: %v", diagnostics)
			}
		})
	}
}
