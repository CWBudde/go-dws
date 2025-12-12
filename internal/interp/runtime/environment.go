package runtime

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ident"
)

// Environment represents a symbol table for variable storage and scope management.
// It supports nested scopes through the outer environment reference, enabling
// proper lexical scoping for DWScript programs.
//
// Implementation Note:
// DWScript is a case-insensitive language. The store uses ident.Map which
// automatically normalizes keys for case-insensitive lookup. This ensures that
// "myVariable", "MyVariable", and "MYVARIABLE" all refer to the same variable.
// The ident.Map also preserves the original casing of keys for error messages.
type Environment struct {
	// store is a case-insensitive map of variable names to their runtime values.
	// Keys are automatically normalized by ident.Map for case-insensitive lookup
	// in accordance with DWScript semantics.
	store *ident.Map[Value]
	// outer references the enclosing (parent) environment for nested scopes
	outer *Environment
}

// NewEnvironment creates a new root-level environment with no outer scope.
// This is typically used for the global scope of a program.
func NewEnvironment() *Environment {
	return &Environment{
		store: ident.NewMap[Value](),
		outer: nil,
	}
}

// NewEnclosedEnvironment creates a new environment that is enclosed by the given
// outer environment. This is used for creating nested scopes such as function
// bodies, blocks, or control flow structures.
//
// When resolving variables, the inner environment is checked first, then the
// outer environments are searched recursively up the scope chain.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	return &Environment{
		store: ident.NewMap[Value](),
		outer: outer,
	}
}

// NewEnclosed creates a new enclosed environment. This breaks the circular import dependency
// and allows the evaluator package to create proper scopes without importing interp.
func (e *Environment) NewEnclosed() interface{} {
	return NewEnclosedEnvironment(e)
}

// Get retrieves a variable value by name. It searches the current environment
// first, then recursively searches outer (parent) environments if not found.
// DWScript is case-insensitive - ident.Map handles normalization automatically.
//
// Returns the value and true if found, or nil and false if the variable is
// undefined in this scope chain.
func (e *Environment) Get(name string) (Value, bool) {
	// Check current environment (ident.Map handles case-insensitive lookup)
	if val, ok := e.store.Get(name); ok {
		return val, true
	}

	// If not found and we have an outer scope, search there
	if e.outer != nil {
		return e.outer.Get(name)
	}

	// Variable not found in any scope
	return nil, false
}

// Set updates an existing variable's value. It searches the current environment
// first, then recursively searches outer environments to find where the variable
// is defined. DWScript is case-insensitive - ident.Map handles normalization automatically.
//
// Returns an error if the variable is not defined in any scope in the chain.
// Use Define() to create a new variable in the current scope.
func (e *Environment) Set(name string, val Value) error {
	// Check if variable exists in current environment (ident.Map handles normalization)
	if e.store.Has(name) {
		e.store.Set(name, val)
		return nil
	}

	// If not in current scope, try outer scope
	if e.outer != nil {
		return e.outer.Set(name, val)
	}

	// Variable doesn't exist in any scope
	return fmt.Errorf("undefined variable: %s", name)
}

// Define creates a new variable in the current environment's scope.
// If a variable with the same name already exists in this scope, it is
// overwritten (no error is returned). DWScript is case-insensitive -
// ident.Map handles normalization automatically.
//
// This differs from Set() which only updates existing variables and errors
// if the variable is not found. Define() is used for variable declarations,
// while Set() is used for assignments.
func (e *Environment) Define(name string, val Value) {
	// ident.Map handles case-insensitive storage automatically
	e.store.Set(name, val)
}

// Has checks if a variable is defined in the current environment or any outer scope.
func (e *Environment) Has(name string) bool {
	_, ok := e.Get(name)
	return ok
}

// GetLocal retrieves a variable value only from the current environment,
// without searching outer scopes. This is useful for checking if a variable
// is shadowing an outer variable. DWScript is case-insensitive -
// ident.Map handles normalization automatically.
func (e *Environment) GetLocal(name string) (Value, bool) {
	// ident.Map handles case-insensitive lookup automatically
	return e.store.Get(name)
}

// Size returns the number of variables defined in the current environment
// (not including outer scopes).
func (e *Environment) Size() int {
	return e.store.Len()
}

// Range iterates over all variables in the current environment (not including outer scopes).
// The function f is called for each variable with the name and value.
// If f returns false, iteration stops.
//
// This method is used by cleanup routines that need to inspect all variables in a scope.
func (e *Environment) Range(f func(name string, value Value) bool) {
	if e.store == nil {
		return
	}
	e.store.Range(f)
}

// Outer returns the outer (parent) environment, or nil if this is the root environment.
// This is primarily used for testing and debugging.
func (e *Environment) Outer() *Environment {
	return e.outer
}
