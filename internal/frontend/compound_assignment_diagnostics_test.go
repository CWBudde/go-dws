package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompile_CompoundAssignmentFixture(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "FailureScripts")
	source, err := os.ReadFile(filepath.Join(root, "assign_op_incompatible.pas"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(root, "assign_op_incompatible.txt"))
	if err != nil {
		t.Fatal(err)
	}
	assertDiagnostics(t, string(source), "assign_op_incompatible.pas", strings.Split(strings.TrimSpace(string(want)), "\n"))
}

func TestCompile_CompoundAssignmentRecovery(t *testing.T) {
	for _, tt := range []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "numeric target incompatible RHS",
			source: "var value: Integer;\nvalue += 'text';",
			want: []string{
				`Syntax Error: Incompatible operands [line: 2, column: 7]`,
				`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 2, column: 10]`,
			},
		},
		{
			name:   "supported RHS assignment incompatible operator",
			source: "var value: String;\nvalue *= 'text';",
			want:   []string{`Syntax Error: Incompatible operands [line: 2, column: 7]`},
		},
		{
			name:   "unsupported Boolean operator",
			source: "var value: Boolean;\nvalue += false;",
			want:   []string{`Syntax Error: Incompatible operands [line: 2, column: 7]`},
		},
		{
			name:   "Variant does not admit string multiplication",
			source: "var value: Variant;\nvalue *= 'text';",
			want:   []string{`Syntax Error: Incompatible operands [line: 2, column: 7]`},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertDiagnostics(t, tt.source, "<test>", tt.want)
		})
	}
}

func TestCompile_CompoundAssignmentValidPairs(t *testing.T) {
	for _, op := range []string{"+=", "-=", "*=", "/="} {
		for _, pair := range []struct{ left, right string }{
			{"Integer", "Integer"},
			{"Float", "Float"},
			{"Float", "Integer"},
			{"Integer", "Variant"},
			{"Variant", "Float"},
		} {
			t.Run(op+"/"+pair.left+"/"+pair.right, func(t *testing.T) {
				source := "var target: " + pair.left + ";\nvar value: " + pair.right + ";\ntarget " + op + " value;"
				assertDiagnostics(t, source, "<test>", nil)
			})
		}
	}
	for _, op := range []string{"+=", "-=", "*=", "/="} {
		t.Run("Integer rejects Float/"+op, func(t *testing.T) {
			assertDiagnostics(t, "var target: Integer;\nvar value: Float;\ntarget "+op+" value;", "<test>", []string{
				`Syntax Error: Incompatible types: Cannot assign "Float" to "Integer" [line: 3, column: 11]`,
			})
		})
	}
	for _, pair := range []struct{ left, right string }{
		{"String", "String"}, {"String", "Variant"}, {"Variant", "String"},
	} {
		t.Run("string concatenation/"+pair.left+"/"+pair.right, func(t *testing.T) {
			source := "var target: " + pair.left + ";\nvar value: " + pair.right + ";\ntarget += value;"
			assertDiagnostics(t, source, "<test>", nil)
		})
	}
}

func TestCompile_CompoundAssignmentClassOperatorAndAppend(t *testing.T) {
	assertDiagnostics(t, `type TAccumulator = class
   procedure Append(value: Integer);
   class operator += (value: Integer) uses Append;
end;
procedure TAccumulator.Append(value: Integer);
begin
   PrintLn(value);
end;
var accumulator := TAccumulator.Create;
accumulator += 1;`, "<test>", nil)

	assertDiagnostics(t, `var values: array of Integer;
values += 1;
values += [2, 3];
var other: array of Integer := [4];
values += other;
PrintLn(Length(values));`, "<test>", nil)
}

func TestCompile_CompoundAssignmentMissingClassOperatorStops(t *testing.T) {
	assertDiagnostics(t, "var target: TObject;\ntarget -= 1;\nunknown := 2;", "<test>", []string{
		`Syntax Error: Incompatible operands [line: 2, column: 8]`,
	})
}

func TestCompile_CompoundAssignmentFunctionNameClassOperator(t *testing.T) {
	assertDiagnostics(t, `type TAccumulator = class
   procedure Append(value: Integer);
   class operator += (value: Integer) uses Append;
end;
procedure TAccumulator.Append(value: Integer);
begin
   PrintLn(value);
end;
function Build: TAccumulator;
begin
   Build := TAccumulator.Create;
   Build += 1;
end;
var accumulator := Build;
accumulator.Free;`, "<test>", nil)
}
