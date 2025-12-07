package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Operator overloading adapter methods.
// These methods implement the InterpreterAdapter interface for operator overloading.

// TryBinaryOperator attempts to find and invoke a binary operator overload.
// Returns (result, true) if operator found, or (nil, false) if not found.
func (i *Interpreter) TryBinaryOperator(operator string, left, right evaluator.Value, node ast.Node) (evaluator.Value, bool) {
	// Convert evaluator.Value to interp.Value (they're the same interface)
	leftVal, _ := left.(Value)
	rightVal, _ := right.(Value)

	// Call the internal tryBinaryOperator method
	result, found := i.tryBinaryOperator(operator, leftVal, rightVal, node)
	return result, found
}

// TryUnaryOperator attempts to find and invoke a unary operator overload.
// Returns (result, true) if operator found, or (nil, false) if not found.
func (i *Interpreter) TryUnaryOperator(operator string, operand evaluator.Value, node ast.Node) (evaluator.Value, bool) {
	// Convert evaluator.Value to interp.Value (they're the same interface)
	operandVal, _ := operand.(Value)

	// Call the internal tryUnaryOperator method
	result, found := i.tryUnaryOperator(operator, operandVal, node)
	return result, found
}
