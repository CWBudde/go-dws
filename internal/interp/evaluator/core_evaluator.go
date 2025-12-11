package evaluator

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// CoreEvaluator provides fallback evaluation methods for cross-cutting concerns.
// These methods represent functionality that doesn't cleanly fit into OOP, declarations,
// or exceptions, and may indicate areas for future architectural improvements.
//
// Task 3.4.5: Extracted from monolithic InterpreterAdapter (67 methods)
// to isolate remaining cross-cutting concerns.
//
// FUTURE WORK: These methods may be eliminated through:
// - EvalNode: Migrate remaining OOP operations to evaluator (Phase 3.5.37+)
// - EvalBuiltinHelperProperty: Move built-in helper logic to evaluator
// - EvalClassPropertyRead/Write: Migrate class property logic to evaluator
type CoreEvaluator interface {
	// ===== Fallback Evaluation (1 method, moderate usage) =====

	// EvalNode evaluates an AST node via interpreter for OOP operations.
	// Used by: method_dispatch.go (2), user_function_helpers.go (1),
	//          helper_methods.go (1), evaluator.go (1), compound_assignments.go (1)
	// MODERATE USAGE: 6 uses (excluding 5 test file references)
	//
	// Current use cases:
	// - OOP context operations requiring Self/class metadata
	// - Complex method dispatch scenarios
	// - Fallback for operations not yet migrated to evaluator
	//
	// FUTURE: May be eliminated by migrating remaining OOP logic to evaluator
	EvalNode(node ast.Node) Value

	// ===== Helper Properties (1 method, moderate usage) =====

	// EvalBuiltinHelperProperty evaluates a built-in helper property.
	// Used by: helper_methods.go (4 uses)
	// MODERATE USAGE: 4 uses - all in same file
	//
	// Built-in helpers: Length, Low, High, etc. on built-in types
	// FUTURE: Could be migrated to evaluator with built-in helper registry
	EvalBuiltinHelperProperty(propSpec string, selfValue Value, node ast.Node) Value

	// ===== Class Properties (2 methods, low usage) =====

	// EvalClassPropertyRead evaluates a class property read operation.
	// Used by: property_read.go (1 use)
	// For CLASS/CLASSINFO member access (static properties).
	EvalClassPropertyRead(classInfo any, propInfo any, node ast.Node) Value

	// EvalClassPropertyWrite evaluates a class property write operation.
	// Used by: property_write.go (1 use)
	// For CLASS/CLASSINFO member assignment (static properties).
	EvalClassPropertyWrite(classInfo any, propInfo any, value Value, node ast.Node) Value
}

// Total: 4 methods
// High usage: None
// Moderate usage (3+ calls): EvalNode (6), EvalBuiltinHelperProperty (4)
// Low usage (1-2 calls): EvalClassPropertyRead (1), EvalClassPropertyWrite (1)
//
// Design rationale:
// - These methods represent cross-cutting concerns that don't fit cleanly
//   into OOP, declaration, or exception categories
// - EvalNode is a temporary fallback for operations not yet migrated to evaluator
// - Built-in helper properties could be moved to evaluator in future
// - Class property operations are rare but needed for static property access
//
// Migration path:
// - Phase 3.5.37+: Eliminate EvalNode by migrating remaining OOP operations
// - Future: Move built-in helper registry to evaluator, eliminate EvalBuiltinHelperProperty
// - Future: Migrate class property operations to evaluator
