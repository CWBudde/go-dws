package interp_test

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readBoundarySource(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s failed: %v", path, err)
	}
	return string(content)
}

// TestInterpDoesNotImportEvaluator verifies the architectural boundary:
// internal/interp must not import internal/interp/evaluator except for the
// canonical construction entry point and test files.
//
// This boundary exists to enable dependency inversion:
// - interp depends on contracts (interfaces)
// - evaluator implements those interfaces
// - interp/new.go owns canonical runtime construction during Phase 4
//
// See docs/architecture/interp-evaluator-boundary.md for rationale.
func TestInterpDoesNotImportEvaluator(t *testing.T) {
	const forbidden = "github.com/cwbudde/go-dws/internal/interp/evaluator"

	fset := token.NewFileSet()
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			// Skip evaluator package itself.
			// This test is about preventing *other* interp packages from importing evaluator.
			if path == "evaluator" {
				return fs.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files - they're allowed to import evaluator.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Allow canonical construction entry points to import evaluator.
		if path == "new.go" || path == filepath.Join("runner", "runner.go") || strings.HasPrefix(path, "runner"+string(filepath.Separator)) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		f, err := parser.ParseFile(fset, path, content, parser.ImportsOnly)
		if err != nil {
			return err
		}

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if importPath == forbidden {
				t.Errorf("%s imports %s - violates interp/evaluator boundary.\n"+
					"Allowed only in new.go, runner/**, and *_test.go.\n"+
					"Use contracts interfaces for all non-construction code.\n"+
					"See docs/architecture/interp-evaluator-boundary.md",
					path, forbidden)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
}

func TestConstructionDoesNotReferenceLegacyBridgeWiring(t *testing.T) {
	t.Parallel()

	checks := []struct {
		path      string
		forbidden []string
	}{
		{
			path:      "new.go",
			forbidden: []string{"SetFocusedInterfaces", "SetRuntimeBridge", "SetEnvironment", "RestoreEnvironment"},
		},
		{
			path:      filepath.Join("runner", "runner.go"),
			forbidden: []string{"SetFocusedInterfaces", "SetRuntimeBridge", "SetEnvironment", "RestoreEnvironment"},
		},
	}

	for _, tc := range checks {
		source := readBoundarySource(t, tc.path)
		for _, needle := range tc.forbidden {
			if strings.Contains(source, needle) {
				t.Fatalf("%s still references legacy construction wiring %q", tc.path, needle)
			}
		}
	}
}

func TestNoLegacyCallbackInterfaceDeclarationsRemain(t *testing.T) {
	t.Parallel()

	type forbiddenDecl struct {
		needle string
		label  string
	}

	forbidden := []forbiddenDecl{
		{needle: "type CoreEvaluator interface", label: "CoreEvaluator"},
		{needle: "type OOPEngine interface", label: "OOPEngine"},
		{needle: "type DeclHandler interface", label: "DeclHandler"},
		{needle: "func (e *Evaluator) SetFocusedInterfaces(", label: "SetFocusedInterfaces"},
	}

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		source := readBoundarySource(t, path)
		for _, decl := range forbidden {
			if strings.Contains(source, decl.needle) {
				t.Errorf("%s still declares legacy callback surface %s", path, decl.label)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
}

func TestNoInterpreterShadowStatementEvaluatorsRemain(t *testing.T) {
	t.Parallel()

	forbidden := map[string]string{
		"evalProgram":                     "program execution belongs to evaluator visitors",
		"evalBlockStatement":              "block execution belongs to evaluator visitors",
		"evalExpressionStatement":         "expression-statement execution belongs to evaluator visitors",
		"evalIfStatement":                 "if-statement execution belongs to evaluator visitors",
		"evalCaseStatement":               "case-statement execution belongs to evaluator visitors",
		"evalWhileStatement":              "while-loop execution belongs to evaluator visitors",
		"evalRepeatStatement":             "repeat-loop execution belongs to evaluator visitors",
		"evalForStatement":                "for-loop execution belongs to evaluator visitors",
		"evalForInStatement":              "for-in execution belongs to evaluator visitors",
		"evalVarDeclStatement":            "var declaration execution belongs to evaluator visitors",
		"evalConstDeclStatement":          "const declaration execution belongs to evaluator visitors",
		"evalAssignmentStatement":         "assignment execution belongs to evaluator visitors",
		"evalCompoundAssignmentStatement": "compound assignment execution belongs to evaluator visitors",
		"evalTryStatement":                "try/except/finally execution belongs to evaluator visitors",
		"evalRaiseStatement":              "raise execution belongs to evaluator visitors",
		"evalBreakStatement":              "break control flow belongs to evaluator visitors",
		"evalContinueStatement":           "continue control flow belongs to evaluator visitors",
		"evalExitStatement":               "exit control flow belongs to evaluator visitors",
	}

	fset := token.NewFileSet()
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*goast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Name == nil {
				continue
			}
			if len(fn.Recv.List) != 1 {
				continue
			}

			star, ok := fn.Recv.List[0].Type.(*goast.StarExpr)
			if !ok {
				continue
			}
			recvIdent, ok := star.X.(*goast.Ident)
			if !ok || recvIdent.Name != "Interpreter" {
				continue
			}

			if reason, forbidden := forbidden[fn.Name.Name]; forbidden {
				t.Errorf("%s still defines Interpreter.%s; %s", path, fn.Name.Name, reason)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
}

func TestDeletedShadowExecutionFilesDoNotReturn(t *testing.T) {
	t.Parallel()

	deleted := []string{
		"array.go",
		"functions_calls.go",
		"objects_instantiation.go",
		"oop_dispatch.go",
		"statements_assignments.go",
		"statements_declarations.go",
		"statements_loops.go",
	}

	for _, path := range deleted {
		if _, err := os.Stat(path); err == nil {
			t.Errorf("%s reappeared; phase 4.9 removed it as dead shadow execution", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s failed: %v", path, err)
		}
	}
}
