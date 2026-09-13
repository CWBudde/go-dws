package types

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// RegisterNodeUnit records lexical ownership of a source tree. Empty unitName
// explicitly denotes the main module; absent nodes are generated or unregistered.
func (ts *TypeSystem) RegisterNodeUnit(root ast.Node, unitName string) {
	if ts == nil || root == nil {
		return
	}
	if ts.nodeUnits == nil {
		ts.nodeUnits = make(map[ast.Node]string)
	}
	unitName = ident.Normalize(unitName)
	ast.Inspect(root, func(node ast.Node) bool {
		if node != nil {
			ts.nodeUnits[node] = unitName
		}
		return true
	})
}

// NodeUnit returns the lexical unit of a source node and whether it was registered.
func (ts *TypeSystem) NodeUnit(node ast.Node) (string, bool) {
	if ts == nil {
		return "", false
	}
	unit, ok := ts.nodeUnits[node]
	return unit, ok
}

// RegisterNodeAlias gives a generated wrapper the lexical ownership of its
// source node without changing the ownership of reused children or contracts.
func (ts *TypeSystem) RegisterNodeAlias(node, source ast.Node) {
	if unit, ok := ts.NodeUnit(source); ok && node != nil {
		ts.nodeUnits[node] = unit
	}
}
