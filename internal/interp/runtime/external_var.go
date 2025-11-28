// Package runtime provides runtime value types for the DWScript interpreter.
// This file contains the ExternalVarValue type for external variable markers.
//
// Task 3.5.130a: Migrated from internal/interp/value.go to enable direct
// evaluator construction without adapter calls.
package runtime

import "fmt"

// ExternalVarValue represents an external variable marker.
// External variables are placeholders for future FFI (Foreign Function Interface).
//
// Task 3.5.130a: Moved to runtime package for direct evaluator access.
type ExternalVarValue struct {
	Name         string // The variable name in DWScript
	ExternalName string // The external name for FFI binding (may be empty)
}

// Type returns "EXTERNAL_VAR".
func (e *ExternalVarValue) Type() string {
	return "EXTERNAL_VAR"
}

// String returns a description of the external variable.
func (e *ExternalVarValue) String() string {
	if e.ExternalName != "" {
		return fmt.Sprintf("external(%s -> %s)", e.Name, e.ExternalName)
	}
	return fmt.Sprintf("external(%s)", e.Name)
}

// ExternalVarName returns the variable name for error reporting.
// Task 3.5.73: Implements ExternalVarAccessor interface for direct access without adapter.
func (e *ExternalVarValue) ExternalVarName() string {
	return e.Name
}
