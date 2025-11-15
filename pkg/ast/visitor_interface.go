package ast

//go:generate go run ../../cmd/gen-visitor/main.go

// Visitor is the interface for AST traversal using the visitor pattern.
// Implementations should define a Visit method that is called for each node.
// If Visit returns nil, the node's children are not traversed.
// Otherwise, Visit is called recursively for all child nodes.
type Visitor interface {
	Visit(node Node) (w Visitor)
}

// Inspect traverses an AST in depth-first order, calling f for each node.
// If f returns false, traversal of that node's children is skipped.
// Otherwise, Inspect is called recursively for each child.
//
// This is a convenience wrapper around Walk for simple inspection tasks.
func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}

// inspector is a helper type that implements Visitor for the Inspect function.
type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f(node) {
		return f
	}
	return nil
}
