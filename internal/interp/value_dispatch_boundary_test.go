package interp

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// Display labels must not silently become runtime dispatch keys again.
func TestRuntimeDispatchDoesNotCompareTypeDisplay(t *testing.T) {
	for _, directory := range []string{"runtime", "evaluator"} {
		err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			positions := token.NewFileSet()
			file, err := parser.ParseFile(positions, path, nil, 0)
			if err != nil {
				return err
			}
			checkTypeDisplayComparisons(t, positions, file)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func checkTypeDisplayComparisons(t *testing.T, positions *token.FileSet, file *ast.File) {
	t.Helper()
	ast.Inspect(file, func(node ast.Node) bool {
		var operands []ast.Expr
		switch expression := node.(type) {
		case *ast.BinaryExpr:
			if expression.Op == token.EQL || expression.Op == token.NEQ {
				operands = []ast.Expr{expression.X, expression.Y}
			}
		case *ast.SwitchStmt:
			operands = []ast.Expr{expression.Tag}
		}
		for _, operand := range operands {
			if isTypeDisplayCall(operand) {
				t.Errorf("%s: runtime dispatch compares a display label", positions.Position(operand.Pos()))
			}
		}
		return true
	})
}

func isTypeDisplayCall(expression ast.Expr) bool {
	call, ok := expression.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "Type"
}
