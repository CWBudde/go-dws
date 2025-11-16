package evaluator

// EnvironmentAdapter adapts any environment implementation to the
// evaluator.Environment interface. This allows the ExecutionContext
// to work with different environment implementations without
// creating circular dependencies.
//
// The adapter uses interface{} internally and relies on the caller
// to ensure type safety. This is a temporary solution for Phase 3.3.1.
// In later phases, the environment will be properly typed.
type EnvironmentAdapter struct {
	underlying interface{}
}

// NewEnvironmentAdapter creates a new environment adapter.
// The underlying environment should implement Get, Set, Define methods.
func NewEnvironmentAdapter(underlying interface{}) *EnvironmentAdapter {
	return &EnvironmentAdapter{underlying: underlying}
}

// Define creates a new variable binding in the current scope.
func (ea *EnvironmentAdapter) Define(name string, value interface{}) {
	// Use type assertion to call the underlying Define method
	if env, ok := ea.underlying.(interface {
		Define(string, interface{})
	}); ok {
		env.Define(name, value)
	}
}

// Get retrieves a variable value by name.
func (ea *EnvironmentAdapter) Get(name string) (interface{}, bool) {
	// Use type assertion to call the underlying Get method
	if env, ok := ea.underlying.(interface {
		Get(string) (interface{}, bool)
	}); ok {
		return env.Get(name)
	}
	return nil, false
}

// Set updates an existing variable value.
func (ea *EnvironmentAdapter) Set(name string, value interface{}) bool {
	// Use type assertion to call the underlying Set method
	if env, ok := ea.underlying.(interface {
		Set(string, interface{}) error
	}); ok {
		// Set returns error in interp.Environment, but bool in evaluator.Environment
		return env.Set(name, value) == nil
	}
	return false
}

// NewEnclosedEnvironment creates a new child scope.
func (ea *EnvironmentAdapter) NewEnclosedEnvironment() Environment {
	// Use type assertion to call the underlying NewEnclosedEnvironment method
	if env, ok := ea.underlying.(interface {
		NewEnclosedEnvironment() interface{}
	}); ok {
		return NewEnvironmentAdapter(env.NewEnclosedEnvironment())
	}
	return ea
}

// Underlying returns the underlying environment implementation.
// This allows callers to access the original environment if needed.
func (ea *EnvironmentAdapter) Underlying() interface{} {
	return ea.underlying
}
