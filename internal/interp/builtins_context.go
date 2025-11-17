package interp

import (
	"math/rand"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/builtins"
)

// Ensure Interpreter implements builtins.Context interface at compile time.
var _ builtins.Context = (*Interpreter)(nil)

// NewError creates an error value with location information from the current node.
// This implements the builtins.Context interface.
func (i *Interpreter) NewError(format string, args ...interface{}) builtins.Value {
	return i.newErrorWithLocation(i.currentNode, format, args...)
}

// CurrentNode returns the AST node currently being evaluated.
// This implements the builtins.Context interface.
func (i *Interpreter) CurrentNode() ast.Node {
	return i.currentNode
}

// RandSource returns the random number generator for built-in functions.
// This implements the builtins.Context interface.
func (i *Interpreter) RandSource() *rand.Rand {
	return i.rand
}

// GetRandSeed returns the current random number generator seed value.
// This implements the builtins.Context interface.
func (i *Interpreter) GetRandSeed() int64 {
	return i.randSeed
}

// SetRandSeed sets the random number generator seed.
// This implements the builtins.Context interface.
func (i *Interpreter) SetRandSeed(seed int64) {
	i.randSeed = seed
	i.rand = rand.New(rand.NewSource(seed))
}

// UnwrapVariant returns the underlying value if input is a Variant, otherwise returns input as-is.
// This implements the builtins.Context interface.
// Task 9.4.5: Support for Variant arguments in built-in functions.
func (i *Interpreter) UnwrapVariant(value builtins.Value) builtins.Value {
	if value != nil {
		// Check if it's a VariantValue and unwrap it
		if variant, ok := value.(*VariantValue); ok {
			if variant.Value == nil {
				return &UnassignedValue{}
			}
			return variant.Value
		}
	}
	return value
}
