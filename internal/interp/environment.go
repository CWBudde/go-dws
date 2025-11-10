package interp

import (
	"fmt"
	"strings"
)

// Environment represents a symbol table for variable storage and scope management.
// It supports nested scopes through the outer environment reference, enabling
// proper lexical scoping for DWScript programs.
type Environment struct {
	// store maps variable names to their runtime values
	store map[string]Value
	// outer references the enclosing (parent) environment for nested scopes
	outer *Environment
}

// NewEnvironment creates a new root-level environment with no outer scope.
// This is typically used for the global scope of a program.
func NewEnvironment() *Environment {
	return &Environment{
		store: make(map[string]Value),
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
		store: make(map[string]Value),
		outer: outer,
	}
}

// Get retrieves a variable value by name. It searches the current environment
// first, then recursively searches outer (parent) environments if not found.
// DWScript is case-insensitive, so names are normalized to lowercase.
//
// Returns the value and true if found, or nil and false if the variable is
// undefined in this scope chain.
func (e *Environment) Get(name string) (Value, bool) {
	// Normalize to lowercase for case-insensitive lookup (DWScript is case-insensitive)
	key := strings.ToLower(name)

	// Check current environment
	val, ok := e.store[key]
	if ok {
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
// is defined. DWScript is case-insensitive, so names are normalized to lowercase.
//
// Returns an error if the variable is not defined in any scope in the chain.
// Use Define() to create a new variable in the current scope.
func (e *Environment) Set(name string, val Value) error {
	// Normalize to lowercase for case-insensitive lookup (DWScript is case-insensitive)
	key := strings.ToLower(name)

	// Check if variable exists in current environment
	if _, ok := e.store[key]; ok {
		e.store[key] = val
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
// overwritten (no error is returned). DWScript is case-insensitive, so names
// are normalized to lowercase.
//
// This differs from Set() which only updates existing variables and errors
// if the variable is not found. Define() is used for variable declarations,
// while Set() is used for assignments.
func (e *Environment) Define(name string, val Value) {
	// Normalize to lowercase for case-insensitive storage (DWScript is case-insensitive)
	key := strings.ToLower(name)
	e.store[key] = val
}

// Has checks if a variable is defined in the current environment or any outer scope.
func (e *Environment) Has(name string) bool {
	_, ok := e.Get(name)
	return ok
}

// GetLocal retrieves a variable value only from the current environment,
// without searching outer scopes. This is useful for checking if a variable
// is shadowing an outer variable. DWScript is case-insensitive, so names are
// normalized to lowercase.
func (e *Environment) GetLocal(name string) (Value, bool) {
	// Normalize to lowercase for case-insensitive lookup (DWScript is case-insensitive)
	key := strings.ToLower(name)
	val, ok := e.store[key]
	return val, ok
}

// Size returns the number of variables defined in the current environment
// (not including outer scopes).
func (e *Environment) Size() int {
	return len(e.store)
}
