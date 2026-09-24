package semantic

import (
	"slices"
	"testing"
)

func TestArrayIndexDiagnostics_ExcessIndices(t *testing.T) {
	for _, suffix := range []string{" := 1;", ";"} {
		t.Run(suffix, func(t *testing.T) {
			analyzer, _ := analyzeSource(t, "var a: array of array of Integer;\na[0, 1, 2, 3]"+suffix)
			want := []string{
				"Too many indices at 2:7",
				"Too many indices at 2:10",
			}
			if got := analyzer.Errors(); !slices.Equal(got, want) {
				t.Fatalf("diagnostics = %q, want %q", got, want)
			}
		})
	}
}

func TestArrayIndexDiagnostics_ValidNeighbors(t *testing.T) {
	for _, input := range []string{
		"var a: array of array of Integer; a[0, 1] := 2; PrintLn(a[0, 1]);",
		"var a: array of array of Integer; a[0][1] := 2; PrintLn(a[0][1]);",
		"var a: array of String; a[0, 1] := 'x'; PrintLn(a[0, 1]);",
		"var s: String; s[1] := 'x'; PrintLn(s[1]);",
		`type T = class
			function GetItem(x, y: Integer): Integer; begin Result := x + y; end;
			procedure SetItem(x, y, value: Integer); begin end;
			property Items[x, y: Integer]: Integer read GetItem write SetItem;
		end;
		var obj := T.Create;
		obj.Items[0, 1] := 2;
		PrintLn(obj.Items[0, 1]);`,
	} {
		t.Run(input, func(t *testing.T) { expectNoErrors(t, input) })
	}
}

func TestArrayIndexDiagnostics_NonArrayBase(t *testing.T) {
	analyzer, _ := analyzeSource(t, "var i: Integer;\ni[1] := 1;\ni := i[1];")
	want := []string{"Array expected at 2:2", "Array expected at 3:7"}
	if got := analyzer.Errors(); !slices.Equal(got, want) {
		t.Fatalf("diagnostics = %q, want %q", got, want)
	}
}
