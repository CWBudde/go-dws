package interp

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestFunctionPointerOperators_ValueContext(t *testing.T) {
	for _, name := range []string{"booleans", "bitwise", "variants", "methods", "exceptions"} {
		t.Run(name, func(t *testing.T) {
			base := filepath.Join("..", "..", "testdata", "function_pointer_operators", name)
			source, err := os.ReadFile(base + ".dws")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			assertOutput(t, runQuickwinScript(t, string(source)), string(want))
		})
	}
}

func TestFunctionPointerOperators_InvalidOperands(t *testing.T) {
	tests := []struct {
		signature  string
		diagnostic string
	}{
		{"function(x: Integer): Boolean", "Incompatible operands"},
		{"function: String", "Invalid Operands"},
		{"function: Integer", "Invalid Operands"},
		{"procedure", "Incompatible operands"},
	}
	for _, op := range []string{"and", "or"} {
		for _, tt := range tests {
			for _, expr := range []string{"callback %s True", "False %s callback"} {
				t.Run(op+"/"+tt.signature+"/"+expr, func(t *testing.T) {
					source := fmt.Sprintf("var callback: %s; PrintLn(%s);", tt.signature, fmt.Sprintf(expr, op))
					assertCompileError(t, source, tt.diagnostic)
				})
			}
		}
	}
}

func TestFunctionPointerOperators_ComparisonUnchanged(t *testing.T) {
	for _, op := range []string{"=", "<>"} {
		t.Run(op, func(t *testing.T) {
			assertCompileError(t, fmt.Sprintf(`
var callback: function: Boolean;
PrintLn(callback %s callback);
`, op), "requires comparable types")
		})
	}
}
