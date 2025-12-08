package interp

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Miscellaneous Built-in Functions
// Length, Copy, array operations, debugging, utilities

// builtinSetLength implements the SetLength() built-in function for AST expressions (var-param version).
// It resizes a dynamic array or string to the specified length.
func (i *Interpreter) builtinSetLength(args []ast.Expression) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "SetLength() expects exactly 2 arguments, got %d", len(args))
	}

	// Use evaluateLValue to support identifiers, indexed arrays, member access, etc.
	currentVal, assignFunc, err := i.evaluateLValue(args[0])
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "SetLength() first argument must be a variable: %s", err.Error())
	}

	// Dereference if it's a var parameter (ReferenceValue)
	if ref, isRef := currentVal.(*ReferenceValue); isRef {
		actualVal, err := ref.Dereference()
		if err != nil {
			return i.newErrorWithLocation(i.currentNode, "%s", err.Error())
		}
		currentVal = actualVal
	}

	// Evaluate the second argument (new length)
	lengthVal := i.Eval(args[1])
	if isError(lengthVal) {
		return lengthVal
	}

	lengthInt, ok := lengthVal.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SetLength() expects integer as second argument, got %s", lengthVal.Type())
	}

	newLength := int(lengthInt.Value)
	// DWScript/Delphi behavior: negative lengths are treated as 0
	if newLength < 0 {
		newLength = 0
	}

	// Handle arrays
	if arrayVal, ok := currentVal.(*ArrayValue); ok {
		// Check that it's a dynamic array
		if arrayVal.ArrayType == nil {
			return i.newErrorWithLocation(i.currentNode, "array has no type information")
		}

		if arrayVal.ArrayType.IsStatic() {
			return i.newErrorWithLocation(i.currentNode, "SetLength() can only be used with dynamic arrays, not static arrays")
		}

		currentLength := len(arrayVal.Elements)

		if newLength != currentLength {
			if newLength < currentLength {
				// Truncate the slice
				arrayVal.Elements = arrayVal.Elements[:newLength]
			} else {
				// Extend the slice with nil values
				additional := make([]Value, newLength-currentLength)
				arrayVal.Elements = append(arrayVal.Elements, additional...)
			}
		}

		return &NilValue{}
	}

	// Handle strings
	if strVal, ok := currentVal.(*StringValue); ok {
		// Use rune-based SetLength to handle UTF-8 correctly
		// This truncates or pads with spaces to match DWScript behavior
		newStr := runeSetLength(strVal.Value, newLength)

		// Create new StringValue
		newValue := &StringValue{Value: newStr}

		// Use the assignment function to update the string
		if err := assignFunc(newValue); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update string variable: %s", err)
		}

		return &NilValue{}
	}

	return i.newErrorWithLocation(i.currentNode, "SetLength() expects array or string as first argument, got %s", currentVal.Type())
}

// builtinSwap implements the Swap() built-in function.
// It swaps the values of two variables: Swap(var a, var b)
func (i *Interpreter) builtinSwap(args []ast.Expression) Value {
	// Validate argument count (exactly 2 arguments)
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Swap() expects exactly 2 arguments, got %d", len(args))
	}

	// Both arguments must be identifiers (variable names)
	var1Ident, ok1 := args[0].(*ast.Identifier)
	var2Ident, ok2 := args[1].(*ast.Identifier)

	if !ok1 {
		return i.newErrorWithLocation(i.currentNode, "Swap() first argument must be a variable, got %T", args[0])
	}
	if !ok2 {
		return i.newErrorWithLocation(i.currentNode, "Swap() second argument must be a variable, got %T", args[1])
	}

	var1Name := var1Ident.Value
	var2Name := var2Ident.Value

	// Get current values from environment
	val1, exists1 := i.env.Get(var1Name)
	if !exists1 {
		return i.newErrorWithLocation(i.currentNode, "undefined variable: %s", var1Name)
	}

	val2, exists2 := i.env.Get(var2Name)
	if !exists2 {
		return i.newErrorWithLocation(i.currentNode, "undefined variable: %s", var2Name)
	}

	// Handle var parameters (ReferenceValue) for first variable
	var ref1 *ReferenceValue
	if ref, isRef := val1.(*ReferenceValue); isRef {
		ref1 = ref
		// Dereference to get the actual value
		actualVal, err := ref.Dereference()
		if err != nil {
			return i.newErrorWithLocation(i.currentNode, "%s", err.Error())
		}
		val1 = actualVal
	}

	// Handle var parameters (ReferenceValue) for second variable
	var ref2 *ReferenceValue
	if ref, isRef := val2.(*ReferenceValue); isRef {
		ref2 = ref
		// Dereference to get the actual value
		actualVal, err := ref.Dereference()
		if err != nil {
			return i.newErrorWithLocation(i.currentNode, "%s", err.Error())
		}
		val2 = actualVal
	}

	// Swap the values
	// If first variable is a var parameter, write through the reference
	if ref1 != nil {
		if err := ref1.Assign(val2); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", var1Name, err)
		}
	} else {
		if err := i.env.Set(var1Name, val2); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", var1Name, err)
		}
	}

	// If second variable is a var parameter, write through the reference
	if ref2 != nil {
		if err := ref2.Assign(val1); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", var2Name, err)
		}
	} else {
		if err := i.env.Set(var2Name, val1); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", var2Name, err)
		}
	}

	return &NilValue{}
}
