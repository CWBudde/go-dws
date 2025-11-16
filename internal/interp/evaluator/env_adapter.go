package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

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
// The underlying environment must implement methods compatible with runtime.Value.
// This function validates that the underlying type is correct and panics if not.
func NewEnvironmentAdapter(underlying interface{}) *EnvironmentAdapter {
	// Validate that the underlying type implements the required methods
	type envInterface interface {
		Define(string, runtime.Value)
		Get(string) (runtime.Value, bool)
		Set(string, runtime.Value) error
	}
	if _, ok := underlying.(envInterface); !ok {
		panic(fmt.Sprintf("NewEnvironmentAdapter: underlying type %T does not implement required methods (Define, Get, Set with runtime.Value)", underlying))
	}
	return &EnvironmentAdapter{underlying: underlying}
}

// Define creates a new variable binding in the current scope.
func (ea *EnvironmentAdapter) Define(name string, value interface{}) {
	// Convert interface{} to runtime.Value
	var val runtime.Value
	if v, ok := value.(runtime.Value); ok {
		val = v
	} else {
		// If not already a Value, this is a programming error
		panic(fmt.Sprintf("Define: value must be runtime.Value, got %T", value))
	}

	// Use type assertion to call the underlying Define method
	if env, ok := ea.underlying.(interface {
		Define(string, runtime.Value)
	}); ok {
		env.Define(name, val)
	}
}

// Get retrieves a variable value by name.
func (ea *EnvironmentAdapter) Get(name string) (interface{}, bool) {
	// Use type assertion to call the underlying Get method
	if env, ok := ea.underlying.(interface {
		Get(string) (runtime.Value, bool)
	}); ok {
		val, found := env.Get(name)
		// runtime.Value is an interface{}, so it can be returned directly
		return val, found
	}
	return nil, false
}

// Set updates an existing variable value.
func (ea *EnvironmentAdapter) Set(name string, value interface{}) bool {
	// Convert interface{} to runtime.Value
	var val runtime.Value
	if v, ok := value.(runtime.Value); ok {
		val = v
	} else {
		// If not already a Value, this is a programming error
		panic(fmt.Sprintf("Set: value must be runtime.Value, got %T", value))
	}

	// Use type assertion to call the underlying Set method
	if env, ok := ea.underlying.(interface {
		Set(string, runtime.Value) error
	}); ok {
		// Set returns error in the underlying Environment, but bool in evaluator.Environment
		return env.Set(name, val) == nil
	}
	return false
}

// NewEnclosedEnvironment creates a new child scope.
func (ea *EnvironmentAdapter) NewEnclosedEnvironment() Environment {
	// The underlying environment type must have a method or function to create enclosed scopes.
	// We check for a method that returns something we can wrap.
	type enclosedEnvCreator interface {
		// This interface matches types that have a method to create child environments
		Define(string, runtime.Value)
		Get(string) (runtime.Value, bool)
		Set(string, runtime.Value) error
	}

	// Try to use reflection-free approach: check if there's a NewEnclosed method
	// that returns an environment-like type
	if envCreator, ok := ea.underlying.(interface {
		NewEnclosed() interface{}
	}); ok {
		newEnv := envCreator.NewEnclosed()
		if _, ok := newEnv.(enclosedEnvCreator); ok {
			return NewEnvironmentAdapter(newEnv)
		}
	}

	// Since we can't call interp.NewEnclosedEnvironment without importing interp,
	// and the underlying type doesn't have a NewEnclosed method, we need to use
	// reflection or return a fallback. For now, return the same adapter.
	// This will be fixed in Phase 3.4 when the environment structure is refactored.
	return ea
}

// Underlying returns the underlying environment implementation.
// This allows callers to access the original environment if needed.
func (ea *EnvironmentAdapter) Underlying() interface{} {
	return ea.underlying
}
