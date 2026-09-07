package ast_test

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
)

// TestWalk_InlineRecordType checks that an inline anonymous record *type*
// (ast.RecordTypeNode) has its fields walked. RecordTypeNode does not embed
// BaseNode, so it is only visited through the generator's knownNodeTypes
// allowlist — an omission there silently skips the whole subtree.
func TestWalk_InlineRecordType(t *testing.T) {
	engine, _ := dwscript.New()
	program, err := engine.Parse(`
		var points: array of record
			X: Integer;
			Y: Integer;
		end;
	`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	var recordTypes, fields int
	visitor := &recordTypeVisitor{recordTypes: &recordTypes, fields: &fields}
	ast.Walk(visitor, program)

	if recordTypes != 1 {
		t.Errorf("expected 1 RecordTypeNode visited, got %d", recordTypes)
	}
	if fields != 2 {
		t.Errorf("expected 2 field declarations visited, got %d", fields)
	}
}

type recordTypeVisitor struct {
	recordTypes *int
	fields      *int
}

func (v *recordTypeVisitor) Visit(node ast.Node) ast.Visitor {
	switch node.(type) {
	case *ast.RecordTypeNode:
		*v.recordTypes++
	case *ast.FieldDecl:
		*v.fields++
	}
	return v
}
