package bytecode

import (
	"fmt"
)

// binaryIntOp applies an integer binary operation to the top two stack values.
func (vm *VM) binaryIntOp(fn func(a, b int64) int64) error {
	right, err := vm.pop()
	if err != nil {
		return err
	}
	left, err := vm.pop()
	if err != nil {
		return err
	}
	if !left.IsInt() || !right.IsInt() {
		return vm.typeError("integer operation", "Integer", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
	}
	vm.push(IntValue(fn(left.AsInt(), right.AsInt())))
	return nil
}

// binaryIntOpChecked applies an integer binary operation that can fail.
func (vm *VM) binaryIntOpChecked(fn func(a, b int64) (int64, error)) error {
	right, err := vm.pop()
	if err != nil {
		return err
	}
	left, err := vm.pop()
	if err != nil {
		return err
	}
	if !left.IsInt() || !right.IsInt() {
		return vm.typeError("integer operation", "Integer", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
	}
	result, err := fn(left.AsInt(), right.AsInt())
	if err != nil {
		return err
	}
	vm.push(IntValue(result))
	return nil
}

// binaryFloatOp applies a float binary operation to the top two stack values.
func (vm *VM) binaryFloatOp(fn func(a, b float64) float64) error {
	right, err := vm.pop()
	if err != nil {
		return err
	}
	left, err := vm.pop()
	if err != nil {
		return err
	}
	if !left.IsNumber() || !right.IsNumber() {
		return vm.typeError("float operation", "Number", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
	}
	vm.push(FloatValue(fn(left.AsFloat(), right.AsFloat())))
	return nil
}

// binaryFloatOpChecked applies a float binary operation that can fail.
func (vm *VM) binaryFloatOpChecked(fn func(a, b float64) (float64, error)) error {
	right, err := vm.pop()
	if err != nil {
		return err
	}
	left, err := vm.pop()
	if err != nil {
		return err
	}
	if !left.IsNumber() || !right.IsNumber() {
		return vm.typeError("float operation", "Number", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
	}
	result, err := fn(left.AsFloat(), right.AsFloat())
	if err != nil {
		return err
	}
	vm.push(FloatValue(result))
	return nil
}

// requireArray validates that a value is an array and returns it.
func (vm *VM) requireArray(val Value, context string) (*ArrayInstance, error) {
	if !val.IsArray() {
		return nil, vm.typeError(context, "Array", val.Type.String())
	}
	arr := val.AsArray()
	if arr == nil {
		return nil, vm.runtimeError("%s on nil array", context)
	}
	return arr, nil
}

// requireInt validates that a value is an integer and returns it.
func (vm *VM) requireInt(val Value, context string) (int, error) {
	if !val.IsInt() {
		return 0, vm.typeError(context, "Integer", val.Type.String())
	}
	return int(val.AsInt()), nil
}

// compare performs a comparison operation on the top two stack values.
func (vm *VM) compare(op OpCode) (bool, error) {
	right, err := vm.pop()
	if err != nil {
		return false, err
	}
	left, err := vm.pop()
	if err != nil {
		return false, err
	}

	switch {
	case left.IsNumber() && right.IsNumber():
		l := left.AsFloat()
		r := right.AsFloat()
		switch op {
		case OpGreater:
			return l > r, nil
		case OpGreaterEqual:
			return l >= r, nil
		case OpLess:
			return l < r, nil
		case OpLessEqual:
			return l <= r, nil
		}
	case left.IsString() && right.IsString():
		l := left.AsString()
		r := right.AsString()
		switch op {
		case OpGreater:
			return l > r, nil
		case OpGreaterEqual:
			return l >= r, nil
		case OpLess:
			return l < r, nil
		case OpLessEqual:
			return l <= r, nil
		}
	default:
		return false, vm.runtimeError("comparison of incompatible types %s and %s", left.Type.String(), right.Type.String())
	}

	return false, vm.runtimeError("unsupported comparison opcode %v", op)
}

// valuesEqual checks if two values are equal.
func (vm *VM) valuesEqual(a, b Value) bool {
	if a.Type == b.Type {
		switch a.Type {
		case ValueNil:
			return true
		case ValueBool:
			return a.AsBool() == b.AsBool()
		case ValueInt:
			return a.AsInt() == b.AsInt()
		case ValueFloat:
			return a.AsFloat() == b.AsFloat()
		case ValueString:
			return a.AsString() == b.AsString()
		case ValueArray:
			return a.AsArray() == b.AsArray()
		case ValueFunction:
			return a.AsFunction() == b.AsFunction()
		case ValueClosure:
			return a.AsClosure() == b.AsClosure()
		case ValueObject:
			return a.AsObject() == b.AsObject()
		default:
			return false
		}
	}

	if a.IsNumber() && b.IsNumber() {
		return a.AsFloat() == b.AsFloat()
	}

	return false
}

// constantAsString retrieves a string constant from a chunk.
func (vm *VM) constantAsString(chunk *Chunk, index int, context string) (string, error) {
	if chunk == nil {
		return "", vm.runtimeError("%s without chunk", context)
	}
	if index < 0 || index >= len(chunk.Constants) {
		return "", vm.runtimeError("%s constant index %d out of range", context, index)
	}
	val := chunk.Constants[index]
	if !val.IsString() {
		return "", vm.runtimeError("%s constant %d is not a string (got %s)", context, index, val.Type.String())
	}
	return val.AsString(), nil
}

// runtimeError creates a runtime error with stack trace.
func (vm *VM) runtimeError(format string, args ...interface{}) error {
	msg := fmt.Sprintf("vm: %s", fmt.Sprintf(format, args...))
	trace := vm.buildStackTrace()
	return &RuntimeError{
		Message: msg,
		Trace:   trace,
	}
}

// typeError creates a type error.
func (vm *VM) typeError(context, expected, actual string) error {
	return vm.runtimeError("%s expects %s but got %s", context, expected, actual)
}

// isTruthy converts a VM value to a boolean for use in conditionals.
// Task 9.35: Support Variant→Boolean implicit conversion
func isTruthy(val Value) bool {
	switch val.Type {
	case ValueBool:
		return val.AsBool()
	case ValueVariant:
		// Unwrap the variant and check the underlying value
		wrapped := val.AsVariant()
		if wrapped.IsNil() {
			// Unassigned variant → false
			return false
		}
		// Recursively check the wrapped value
		return variantToBool(wrapped)
	default:
		// In DWScript, only booleans and variants can be used in conditions
		// Non-boolean values in conditionals would be a type error
		// For now, treat non-booleans as false
		return false
	}
}

// variantToBool converts a variant's wrapped value to boolean following DWScript semantics.
// Task 9.35: empty/nil/zero → false, otherwise → true
func variantToBool(val Value) bool {
	switch val.Type {
	case ValueNil:
		return false
	case ValueBool:
		return val.AsBool()
	case ValueInt:
		return val.AsInt() != 0
	case ValueFloat:
		return val.AsFloat() != 0.0
	case ValueString:
		return val.AsString() != ""
	case ValueVariant:
		// Nested variant - recursively unwrap
		wrapped := val.AsVariant()
		return variantToBool(wrapped)
	default:
		// For objects, arrays, functions, etc: non-nil → true
		return true
	}
}
