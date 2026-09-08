package evaluator

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestEvaluatorDoesNotImportSemantic keeps execution independent of the
// analyzer. Shared type and overload rules belong in internal/types.
func TestEvaluatorDoesNotImportSemantic(t *testing.T) {
	t.Parallel()
	const forbidden = "github.com/cwbudde/go-dws/internal/semantic"
	fset := token.NewFileSet()
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if importPath == forbidden || strings.HasPrefix(importPath, forbidden+"/") {
				t.Errorf("%s imports %s; evaluator must consume shared type rules without depending on semantic analysis", path, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
