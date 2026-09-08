package main

import (
	goast "go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestGeneratedVisitorIsUpToDate guards against drift between the AST node
// definitions and the checked-in pkg/ast/visitor_generated.go.
//
// Drift is silent: a node type that stops being recognized simply loses its
// traversal, and a newly added one is never walked. Run
// `go run cmd/gen-visitor/main.go` to regenerate when this fails.
func TestGeneratedVisitorIsUpToDate(t *testing.T) {
	astDir := filepath.Join("..", "..", "pkg", "ast")

	nodes, err := parseASTFiles(astDir)
	if err != nil {
		t.Fatalf("parsing AST files: %v", err)
	}

	code, err := generateVisitorCode(nodes)
	if err != nil {
		t.Fatalf("generating code: %v", err)
	}

	formatted, err := format.Source(code)
	if err != nil {
		t.Fatalf("formatting code: %v", err)
	}

	committed, err := os.ReadFile(filepath.Join(astDir, "visitor_generated.go"))
	if err != nil {
		t.Fatalf("reading committed visitor: %v", err)
	}

	if string(formatted) != string(committed) {
		t.Error("pkg/ast/visitor_generated.go is out of date; run: go run cmd/gen-visitor/main.go")
	}
}

// TestParseASTFiles_RecognizesTypeExpressionNodes checks that every
// TypeExpression implementation is recognized as a node type. Those that do not
// embed BaseNode are recognized only through the knownNodeTypes allowlist, so
// omitting one silently drops its traversal (RecordTypeNode was missing until
// 2026-09).
//
// The expected set is derived from the AST sources — the receivers of the
// typeExpressionNode() marker method — rather than restated here, so a newly
// added TypeExpression that nobody remembers to allowlist fails this test
// instead of quietly agreeing with the generator.
func TestParseASTFiles_RecognizesTypeExpressionNodes(t *testing.T) {
	astDir := filepath.Join("..", "..", "pkg", "ast")

	implementations, err := typeExpressionImplementations(astDir)
	if err != nil {
		t.Fatalf("collecting TypeExpression implementations: %v", err)
	}
	if len(implementations) == 0 {
		t.Fatal("no TypeExpression implementations found; the marker method may have been renamed")
	}

	nodes, err := parseASTFiles(astDir)
	if err != nil {
		t.Fatalf("parsing AST files: %v", err)
	}

	found := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		found[node.Name] = true
	}

	for _, name := range implementations {
		if !found[name] {
			t.Errorf("%s implements TypeExpression but was not recognized as a node type; add it to knownNodeTypes", name)
		}
	}
}

// typeExpressionImplementations returns the names of the types declaring the
// typeExpressionNode() marker method in dir, i.e. every TypeExpression
// implementation, sorted.
func typeExpressionImplementations(dir string) ([]string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				funcDecl, ok := decl.(*goast.FuncDecl)
				if !ok || funcDecl.Name.Name != "typeExpressionNode" || funcDecl.Recv == nil {
					continue
				}
				if name := receiverTypeName(funcDecl.Recv); name != "" {
					names = append(names, name)
				}
			}
		}
	}

	sort.Strings(names)
	return names, nil
}

// receiverTypeName extracts the base type name of a method receiver, stripping
// a pointer indirection if present.
func receiverTypeName(recv *goast.FieldList) string {
	if len(recv.List) == 0 {
		return ""
	}
	expr := recv.List[0].Type
	if star, ok := expr.(*goast.StarExpr); ok {
		expr = star.X
	}
	if ident, ok := expr.(*goast.Ident); ok {
		return ident.Name
	}
	return ""
}
