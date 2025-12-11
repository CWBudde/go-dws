package evaluator

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// CoreEvaluator provides fallback evaluation for cross-cutting concerns.
// May be eliminated in future by migrating remaining OOP logic to evaluator.
type CoreEvaluator interface {
	// EvalNode evaluates an AST node via interpreter for OOP operations.
	// Fallback for operations not yet migrated to evaluator.
	EvalNode(node ast.Node) Value

	// EvalBuiltinHelperProperty evaluates a built-in helper property (Length, Low, High, etc).
	EvalBuiltinHelperProperty(propSpec string, selfValue Value, node ast.Node) Value

	// EvalClassPropertyRead evaluates a class property read (static properties).
	EvalClassPropertyRead(classInfo any, propInfo any, node ast.Node) Value

	// EvalClassPropertyWrite evaluates a class property write (static properties).
	EvalClassPropertyWrite(classInfo any, propInfo any, value Value, node ast.Node) Value
}
