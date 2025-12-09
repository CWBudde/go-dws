package interp

import "github.com/cwbudde/go-dws/internal/interp/runtime"

// Environment is a type alias for runtime.Environment.
// This provides backward compatibility during the migration to unified runtime types.
// Code in internal/interp can continue using Environment as before.
//
// Phase 3.1.2: Temporary alias during environment migration.
// Phase 3.1.5: This alias will be used to route all env access through ExecutionContext.
type Environment = runtime.Environment

// NewEnvironment creates a new root-level environment with no outer scope.
// This is a convenience wrapper around runtime.NewEnvironment().
func NewEnvironment() *Environment {
	return runtime.NewEnvironment()
}

// NewEnclosedEnvironment creates a new environment that is enclosed by the given
// outer environment. This is a convenience wrapper around runtime.NewEnclosedEnvironment().
func NewEnclosedEnvironment(outer *Environment) *Environment {
	return runtime.NewEnclosedEnvironment(outer)
}
