package interp

import (
	"fmt"
	"sync"
)

// ExternalFunctionRegistry stores external Go functions registered for DWScript.
// It provides thread-safe registration and lookup of external functions.
type ExternalFunctionRegistry struct {
	functions map[string]*ExternalFunctionValue
	mu        sync.RWMutex
}

// NewExternalFunctionRegistry creates a new empty registry.
func NewExternalFunctionRegistry() *ExternalFunctionRegistry {
	return &ExternalFunctionRegistry{
		functions: make(map[string]*ExternalFunctionValue),
	}
}

// Register adds an external function to the registry.
// Returns an error if a function with the same name is already registered.
func (r *ExternalFunctionRegistry) Register(name string, wrapper ExternalFunctionWrapper) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.functions[name]; exists {
		return fmt.Errorf("function %s is already registered", name)
	}

	r.functions[name] = &ExternalFunctionValue{
		Name:    name,
		Wrapper: wrapper,
	}

	return nil
}

// Get retrieves an external function by name.
// Returns the function and true if found, nil and false otherwise.
func (r *ExternalFunctionRegistry) Get(name string) (*ExternalFunctionValue, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, exists := r.functions[name]
	return fn, exists
}

// Has checks if an external function with the given name is registered.
func (r *ExternalFunctionRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.functions[name]
	return exists
}

// List returns the names of all registered external functions.
func (r *ExternalFunctionRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.functions))
	for name := range r.functions {
		names = append(names, name)
	}
	return names
}

// ExternalFunctionWrapper is an interface for wrapping Go functions.
// The public API (pkg/dwscript) provides the implementation.
type ExternalFunctionWrapper interface {
	// Call invokes the wrapped function with DWScript values.
	Call(args []Value) (Value, error)

	// GetVarParams returns a slice indicating which parameters are by-reference (var parameters).
	GetVarParams() []bool

	// SetInterpreter sets the interpreter reference for callback support.
	SetInterpreter(interp *Interpreter)
}

// ExternalFunctionValue represents an external Go function as a DWScript value.
// It implements the Value interface so it can be stored in the environment.
type ExternalFunctionValue struct {
	Wrapper ExternalFunctionWrapper
	Name    string
}

// Type implements Value.Type
func (e *ExternalFunctionValue) Type() string {
	return "EXTERNAL_FUNCTION"
}

// String implements Value.String
func (e *ExternalFunctionValue) String() string {
	return fmt.Sprintf("<external function %s>", e.Name)
}
