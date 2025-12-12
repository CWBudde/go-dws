package evaluator

import "github.com/cwbudde/go-dws/internal/interp/contracts"

// Focused interfaces are defined in contracts package to avoid import cycles.
// These aliases maintain backward compatibility for evaluator-internal usage.
//
// See contracts package for full interface documentation.
// See PLAN.md Phase 4 for the plan to eliminate these callback interfaces.

type (
	// OOPEngine handles runtime OOP operations (method dispatch, constructors, operators).
	// 20 methods, ~37 callback calls from evaluator.
	OOPEngine = contracts.OOPEngine

	// DeclHandler handles type declaration processing (classes, interfaces, helpers).
	// 38 methods, ~41 callback calls from evaluator.
	DeclHandler = contracts.DeclHandler

	// ExceptionManager handles exception creation, propagation, and cleanup.
	// 6 methods, ~6 callback calls from evaluator.
	ExceptionManager = contracts.ExceptionManager

	// CoreEvaluator provides fallback evaluation for cross-cutting concerns.
	// 4 methods, ~18 callback calls from evaluator.
	CoreEvaluator = contracts.CoreEvaluator
)
