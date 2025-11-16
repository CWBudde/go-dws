package interp

// Options defines the interface for configuring the interpreter.
// This interface breaks the circular dependency between internal/interp and pkg/dwscript.
// The pkg/dwscript.Options concrete type implements this interface.
//
// By using an interface here instead of a concrete type:
// - internal/interp can accept configuration without importing pkg/dwscript
// - pkg/dwscript can configure the interpreter without creating circular imports
// - The reflection hack in NewWithOptions is eliminated
type Options interface {
	// GetExternalFunctions returns the external function registry, or nil if not set.
	GetExternalFunctions() *ExternalFunctionRegistry

	// GetMaxRecursionDepth returns the maximum recursion depth for function calls.
	// Returns 0 if not set (caller should use default).
	GetMaxRecursionDepth() int
}
