package evaluator

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ExceptionManager handles exception creation, propagation, and cleanup.
// This interface encapsulates exception handling logic without exposing
// the full interpreter internals.
//
// Task 3.4.5: Extracted from monolithic InterpreterAdapter (67 methods)
// to create focused interface for exception management.
type ExceptionManager interface {
	// ===== Exception Creation (2 methods) =====

	// CreateExceptionDirect creates an exception object with metadata.
	// Used by: visitor_statements.go (1 use)
	// For explicit exception construction in raise statements.
	CreateExceptionDirect(classMetadata any, message string, pos any, callStack any) any

	// WrapObjectInException wraps an object instance in an exception.
	// Used by: visitor_statements.go (1 use)
	// For raising object instances as exceptions.
	WrapObjectInException(objInstance Value, pos any, callStack any) any

	// ===== Contract Exceptions (1 method) =====

	// CreateContractException creates an exception for contract violations.
	// Used by: visitor_statements.go (1 use)
	// For require/ensure/invariant failures.
	CreateContractException(className, message string, node ast.Node, classMetadata interface{}, callStack interface{}) interface{}

	// ===== Type Cast Exceptions (1 method) =====

	// RaiseTypeCastException raises an exception for invalid type casts.
	// Used by: visitor_expressions_identifiers.go (1 use)
	// For runtime type cast failures (as operator).
	RaiseTypeCastException(message string, node ast.Node)

	// ===== Assertion Failures (1 method) =====

	// RaiseAssertionFailed raises an exception for failed assertions.
	// Used by: visitor_statements.go (1 use)
	// For assert statement failures.
	RaiseAssertionFailed(customMessage string)

	// ===== Cleanup (1 method) =====

	// CleanupInterfaceReferences cleans up interface reference counts.
	// Used by: visitor_statements.go (1 use)
	// For cleanup in finally blocks or scope exit.
	CleanupInterfaceReferences(env interface{})
}

// Total: 6 methods
// Usage pattern: All methods have 1 caller each
//
// Distribution:
// - visitor_statements.go: 4 uses (exception creation, contracts, assertions, cleanup)
// - visitor_expressions_identifiers.go: 1 use (type cast exceptions)
//
// Design rationale:
// - All 6 methods have single callers, which might suggest inlining
// - However, keeping interface-based design provides:
//   1. Clear separation: exception logic vs evaluation logic
//   2. Future extensibility: custom exception handlers, different exception models
//   3. Testability: can mock exception handling independently
//   4. Consistency: matches OOPEngine and DeclHandler patterns
