package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompile_ArrayAssignmentRecoveryFixtures(t *testing.T) {
	for _, name := range []string{"array_assign_error3", "array_const"} {
		t.Run(name, func(t *testing.T) {
			root := filepath.Join("..", "..", "testdata", "fixtures", "FailureScripts")
			source, err := os.ReadFile(filepath.Join(root, name+".pas"))
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join(root, name+".txt"))
			if err != nil {
				t.Fatal(err)
			}
			assertDiagnostics(t, string(source), name+".pas", strings.Split(strings.TrimSpace(string(want)), "\n"))
		})
	}
}

func TestCompile_ConstantArrayAssignmentRecovery(t *testing.T) {
	assertDiagnostics(t, "const values = [1];\nvalues[0] :=\n  'bad';", "<test>", []string{
		`Syntax Error: Cannot assign a value to the left-side argument [line: 2, column: 11]`,
		`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 2, column: 11]`,
	})
}

func TestCompile_EmptyArrayConstantAssignment(t *testing.T) {
	assertDiagnostics(t, "const empty = [];\nvar integers: array of Integer;\nvar strings: array of String;\nintegers := empty;\nstrings := empty;", "<test>", nil)
	assertDiagnostics(t, "var variants: array of Variant;\nvar integers: array of Integer;\nintegers := variants;", "<test>", []string{
		`Syntax Error: Incompatible types: Cannot assign "array of Variant" to "array of Integer" [line: 3, column: 13]`,
	})
}
