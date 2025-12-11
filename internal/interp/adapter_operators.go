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

// tryBinaryOperator is the internal implementation that looks up operator overloads.
// It checks in order: left operand's class, right operand's class, global operators.
func (i *Interpreter) tryBinaryOperator(operator string, left, right Value, node ast.Node) (Value, bool) {
	operands := []Value{left, right}
	operandTypes := []string{valueTypeKey(left), valueTypeKey(right)}

	if obj, ok := left.(*ObjectInstance); ok {
		if entry, found := obj.Class.LookupOperator(operator, operandTypes); found {
			// Convert runtime.OperatorEntry to runtimeOperatorEntry
			concreteClass, ok := entry.Class.(*ClassInfo)
			if !ok {
				return i.newErrorWithLocation(node, "invalid class type for operator"), true
			}
			runtimeEntry := &runtimeOperatorEntry{
				Class:         concreteClass,
				Operator:      entry.Operator,
				BindingName:   entry.BindingName,
				OperandTypes:  entry.OperandTypes,
				SelfIndex:     entry.SelfIndex,
				IsClassMethod: entry.IsClassMethod,
			}
			return i.invokeRuntimeOperator(runtimeEntry, operands, node), true
		}
	}
	if obj, ok := right.(*ObjectInstance); ok {
		if entry, found := obj.Class.LookupOperator(operator, operandTypes); found {
			// Convert runtime.OperatorEntry to runtimeOperatorEntry
			concreteClass, ok := entry.Class.(*ClassInfo)
			if !ok {
				return i.newErrorWithLocation(node, "invalid class type for operator"), true
			}
			runtimeEntry := &runtimeOperatorEntry{
				Class:         concreteClass,
				Operator:      entry.Operator,
				BindingName:   entry.BindingName,
				OperandTypes:  entry.OperandTypes,
				SelfIndex:     entry.SelfIndex,
				IsClassMethod: entry.IsClassMethod,
			}
			return i.invokeRuntimeOperator(runtimeEntry, operands, node), true
		}
	}
	if entry, found := i.typeSystem.Operators().Lookup(operator, operandTypes); found {
		// Convert types.OperatorEntry to runtimeOperatorEntry
		runtimeEntry := &runtimeOperatorEntry{
			Class:         entry.Class.(*ClassInfo),
			Operator:      entry.Operator,
			BindingName:   entry.BindingName,
			OperandTypes:  entry.OperandTypes,
			SelfIndex:     entry.SelfIndex,
			IsClassMethod: entry.IsClassMethod,
		}
		return i.invokeRuntimeOperator(runtimeEntry, operands, node), true
	}
	return nil, false
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

// tryUnaryOperator is the internal implementation that looks up unary operator overloads.
// It checks in order: operand's class, global operators.
func (i *Interpreter) tryUnaryOperator(operator string, operand Value, node ast.Node) (Value, bool) {
	operands := []Value{operand}
	operandTypes := []string{valueTypeKey(operand)}

	if obj, ok := operand.(*ObjectInstance); ok {
		if entry, found := obj.Class.LookupOperator(operator, operandTypes); found {
			// Convert runtime.OperatorEntry to runtimeOperatorEntry
			concreteClass, ok := entry.Class.(*ClassInfo)
			if !ok {
				return i.newErrorWithLocation(node, "invalid class type for operator"), true
			}
			runtimeEntry := &runtimeOperatorEntry{
				Class:         concreteClass,
				Operator:      entry.Operator,
				BindingName:   entry.BindingName,
				OperandTypes:  entry.OperandTypes,
				SelfIndex:     entry.SelfIndex,
				IsClassMethod: entry.IsClassMethod,
			}
			return i.invokeRuntimeOperator(runtimeEntry, operands, node), true
		}
	}

	if entry, found := i.typeSystem.Operators().Lookup(operator, operandTypes); found {
		// Convert types.OperatorEntry to runtimeOperatorEntry
		runtimeEntry := &runtimeOperatorEntry{
			Class:         entry.Class.(*ClassInfo),
			Operator:      entry.Operator,
			BindingName:   entry.BindingName,
			OperandTypes:  entry.OperandTypes,
			SelfIndex:     entry.SelfIndex,
			IsClassMethod: entry.IsClassMethod,
		}
		return i.invokeRuntimeOperator(runtimeEntry, operands, node), true
	}

	return nil, false
}
