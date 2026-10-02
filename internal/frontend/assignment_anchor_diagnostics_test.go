package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// Ordinary assignments report at their RHS; a property backed by a field is
// an assignment too, rather than a setter-argument mismatch.
func TestCompile_AssignmentAnchorDiagnostics(t *testing.T) {
	for _, tt := range []struct{ category, name string }{
		{"FailureScripts", "assign_error"},
		{"FailureScripts", "coalesce_class"},
		{"InterfacesFail", "assign_obj_from_intf"},
		{"InterfacesFail", "interface_inheritence2"},
		// Array expressions retain their existing expression-specific anchors.
		{"FailureScripts", "array_concat"},
		{"FailureScripts", "array_dyn_mismatch"},
	} {
		t.Run(tt.category+"/"+tt.name, func(t *testing.T) {
			root := filepath.Join("..", "..", "testdata", "fixtures", tt.category)
			source, err := os.ReadFile(filepath.Join(root, tt.name+".pas"))
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join(root, tt.name+".txt"))
			if err != nil {
				t.Fatal(err)
			}
			assertDiagnostics(t, string(source), tt.name+".pas", strings.Split(strings.TrimSpace(string(want)), "\n"))
		})
	}
}

func TestCompile_AssignmentRHSOnNextLine(t *testing.T) {
	assertDiagnostics(t, "var s: String;\ns :=\n  42;\nPrintLn(s);", "<test>", []string{
		`Syntax Error: Incompatible types: Cannot assign "Integer" to "String" [line: 3, column: 3]`,
	})
}

func TestCompile_ImplicitFieldPropertyAssignmentAnchor(t *testing.T) {
	source := "type TTest = class\n  F: Integer;\n  property Value: Integer write F;\n  procedure Run;\n  begin\n    Value :=\n      'bad';\n  end;\nend;"
	result := Compile(source, "<test>", semantic.HintsLevelDisabled)
	got := result.DiagnosticStrings()
	want := `Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 7, column: 7]`
	if len(got) != 1 || got[0] != want {
		t.Fatalf("diagnostics = %q, want [%q]", got, want)
	}
}

func TestCompile_InheritedPropertyAssignmentAnchors(t *testing.T) {
	source := `type TBase = class
  FValue: Integer;
  procedure SetValue(value: Integer); begin FValue := value; end;
  property Direct: Integer write FValue;
  property Indirect: Integer write SetValue;
end;
type TChild = class(TBase)
  procedure Probe;
end;
procedure TChild.Probe;
begin
  FValue := 'bad';
  Direct := 'bad';
  Indirect := 'bad';
  Self.FValue := 'bad';
  Self.Direct := 'bad';
  Self.Indirect := 'bad';
end;
var c: TChild;
c.FValue := 'bad';
c.Direct := 'bad';
c.Indirect := 'bad';`
	want := []string{
		`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 12, column: 13]`,
		`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 13, column: 13]`,
		`Syntax Error: Argument 0 expects type "Integer" instead of "String" [line: 14, column: 3]`,
		`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 15, column: 18]`,
		`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 16, column: 18]`,
		`Syntax Error: Argument 0 expects type "Integer" instead of "String" [line: 17, column: 8]`,
		`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 20, column: 13]`,
		`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 21, column: 13]`,
		`Syntax Error: Argument 0 expects type "Integer" instead of "String" [line: 22, column: 3]`,
	}
	result := Compile(source, "<test>", semantic.HintsLevelDisabled)
	got := result.DiagnosticStrings()
	if len(got) != len(want) {
		t.Fatalf("diagnostics = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}
