package interp

import "testing"

// TestVariantToScalarAtDeclaredTypes covers PLAN.md E14: a Variant stored into
// an Integer or Float location takes the destination's type, as DWScript's
// VariantToInt64 does. Float-to-Integer rounds half to even (Delphi's Round).
func TestVariantToScalarAtDeclaredTypes(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"var initializer", `var v : Variant := 2.5; var i : Integer := v; PrintLn(i); PrintLn(i + 1);`, "2\n3\n"},
		{"round half to even", `var v : Variant := 3.5; var i : Integer := v; PrintLn(i);`, "4\n"},
		{"negative fraction", `var v : Variant := -1.7; var i : Integer := v; PrintLn(i);`, "-2\n"},
		{"assignment", `var v : Variant := 2.5; var i : Integer; i := v; PrintLn(i);`, "2\n"},
		{"parameter", `procedure P(x : Integer); begin PrintLn(x); end; var v : Variant := 7.25; P(v);`, "7\n"},
		{"function result", `function F(v : Variant) : Integer; begin Result := v; end; PrintLn(F(9.75));`, "10\n"},
		{"string", `var v : Variant := '12'; var i : Integer := v; PrintLn(i + 1);`, "13\n"},
		{"boolean", `var v : Variant := True; var i : Integer := v; PrintLn(i);`, "1\n"},
		{"integer into Float", `var v : Variant := 2; var f : Float := v; PrintLn(f / 4);`, "0.5\n"},
		{"integer unchanged", `var v : Variant := 5; var i : Integer := v; PrintLn(i * 2);`, "10\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runHelperScript(t, tt.source); got != tt.want {
				t.Fatalf("output mismatch: want %q, got %q", tt.want, got)
			}
		})
	}
}
