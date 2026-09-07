package main

import (
	"go/format"
	"os"
	"path/filepath"
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

// TestParseASTFiles_RecognizesTypeExpressionNodes checks the node types that
// implement TypeExpression without embedding BaseNode. They are recognized only
// through the knownNodeTypes allowlist, so omitting one silently drops its
// traversal (RecordTypeNode was missing until 2026-09).
func TestParseASTFiles_RecognizesTypeExpressionNodes(t *testing.T) {
	nodes, err := parseASTFiles(filepath.Join("..", "..", "pkg", "ast"))
	if err != nil {
		t.Fatalf("parsing AST files: %v", err)
	}

	found := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		found[node.Name] = true
	}

	for _, name := range []string{
		"ArrayTypeNode",
		"ClassOfTypeNode",
		"FunctionPointerTypeNode",
		"RecordTypeNode",
		"SetTypeNode",
	} {
		if !found[name] {
			t.Errorf("%s was not recognized as a node type; add it to knownNodeTypes", name)
		}
	}
}
