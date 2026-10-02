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
		{"InterfacesFail", "assign_intf_from_obj"},
		{"InterfacesFail", "assign_intf_from_intf"},
		{"FailureScripts", "multi_dim_dyn_array1"},
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

func TestCompile_ClassToInterfaceAssignment(t *testing.T) {
	prefix := "type ITest = interface end;\ntype TPlain = class end;\nvar i: ITest;\nvar o: TPlain;\n"
	assertDiagnostics(t, prefix+"i :=\n  o;", "<test>", []string{
		`Syntax Error: Class "TPlain" does not implement interface "ITest" [line: 5, column: 3]`,
	})

	// The declaration's type controls compatibility even if the stored object
	// implements the interface. Nil and directly implementing classes remain valid.
	source := `type ITest = interface end;
type TImpl = class(TObject, ITest) end;
type TChild = class(TImpl) end;
type IAlias = ITest;
var i: IAlias;
i := nil;
i := TImpl.Create;
i := TChild.Create;`
	assertDiagnostics(t, source, "<test>", nil)
}

func TestCompile_ClassToInterfacePropertyAssignment(t *testing.T) {
	assertDiagnostics(t, `type ITest = interface end;
type TPlain = class end;
type THolder = class
  F: ITest;
  property Direct: ITest write F;
  procedure SetValue(v: ITest); begin F := v; end;
  property Indirect: ITest write SetValue;
end;
var h: THolder;
var o: TPlain;
h.F := o;
h.Direct := o;
h.Indirect := o;`, "<test>", []string{
		`Syntax Error: Class "TPlain" does not implement interface "ITest" [line: 11, column: 5]`,
		`Syntax Error: Class "TPlain" does not implement interface "ITest" [line: 12, column: 10]`,
		`Syntax Error: Argument 0 expects type "ITest" instead of "TPlain" [line: 13, column: 3]`,
	})
}

func TestCompile_NewArrayAssignmentAnchor(t *testing.T) {
	for _, tt := range []struct {
		name, allocation, want string
	}{
		{"space", "new String  [2];", `Syntax Error: Incompatible types: Cannot assign "array of String" to "array of array of String" [line: 2, column: 18]`},
		{"newline", "new String\n  [2];", `Syntax Error: Incompatible types: Cannot assign "array of String" to "array of array of String" [line: 3, column: 3]`},
		{"comment", "new String{size}[2];", `Syntax Error: Incompatible types: Cannot assign "array of String" to "array of array of String" [line: 2, column: 22]`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertDiagnostics(t, "var a: array of array of String;\na := "+tt.allocation, "<test>", []string{tt.want})
		})
	}
}

func TestCompile_NewArrayDimensionRecovery(t *testing.T) {
	assertDiagnostics(t, "var a := new Integer[False, 'bad', 2];\nvar b: array of array of array of Integer;\nb := a;", "<test>", []string{
		`Syntax Error: Integer expression expected [line: 1, column: 22]`,
		`Syntax Error: Integer expression expected [line: 1, column: 29]`,
	})
}

func TestCompile_NewArrayMemberAssignmentAnchor(t *testing.T) {
	assertDiagnostics(t, `type TArray = array of array of Integer;
type THolder = class
  F: TArray;
  property Value: TArray write F;
end;
var h: THolder;
h.F := new Integer[2];
h.Value := new Integer[2];`, "<test>", []string{
		`Syntax Error: Incompatible types: Cannot assign "array of Integer" to "array of array of Integer" [line: 7, column: 19]`,
		`Syntax Error: Incompatible types: Cannot assign "array of Integer" to "array of array of Integer" [line: 8, column: 23]`,
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
