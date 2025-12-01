package interp

import (
	"fmt"
	"math"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// evalBinaryExpression evaluates a binary expression.
func (i *Interpreter) evalBinaryExpression(expr *ast.BinaryExpression) Value {
	// Special handling for operators that require short-circuit evaluation
	if expr.Operator == "??" {
		return i.evalCoalesceExpression(expr)
	}
	if expr.Operator == "and" {
		return i.evalAndExpression(expr)
	}
	if expr.Operator == "or" {
		return i.evalOrExpression(expr)
	}

	left := i.Eval(expr.Left)
	if isError(left) {
		return left
	}
	if left == nil {
		return i.newErrorWithLocation(expr.Left, "left operand evaluated to nil")
	}

	right := i.Eval(expr.Right)
	if isError(right) {
		return right
	}
	if right == nil {
		return i.newErrorWithLocation(expr.Right, "right operand evaluated to nil")
	}

	// Ensure currentNode points to the operator for accurate error locations
	i.currentNode = expr

	if result, ok := i.tryBinaryOperator(expr.Operator, left, right, expr); ok {
		return result
	}

	// Handle 'in' operator for array membership checking
	if expr.Operator == "in" {
		return i.evalInOperator(left, right, expr)
	}

	// Handle operations based on operand types
	// Task 9.4.5: Check for Variant FIRST before type-specific operations
	// This ensures Variant operations take precedence over type-specific handlers
	switch {
	// Handle Variant operations
	case left.Type() == "VARIANT" || right.Type() == "VARIANT":
		return i.evalVariantBinaryOp(expr.Operator, left, right, expr)

	case left.Type() == "INTEGER" && right.Type() == "INTEGER":
		return i.evalIntegerBinaryOp(expr.Operator, left, right)

	case left.Type() == "FLOAT" || right.Type() == "FLOAT":
		return i.evalFloatBinaryOp(expr.Operator, left, right)

	case left.Type() == "STRING" && right.Type() == "STRING":
		return i.evalStringBinaryOp(expr.Operator, left, right)

	case left.Type() == "SET" && right.Type() == "SET":
		leftSet := left.(*SetValue)
		rightSet := right.(*SetValue)
		return i.evalBinarySetOperation(leftSet, rightSet, expr.Operator)

	// Allow string concatenation with RTTI_TYPEINFO
	case (left.Type() == "STRING" && right.Type() == "RTTI_TYPEINFO") || (left.Type() == "RTTI_TYPEINFO" && right.Type() == "STRING"):
		if expr.Operator == "+" {
			// Convert both to strings and concatenate
			leftStr := left.String()
			rightStr := right.String()
			return &StringValue{Value: leftStr + rightStr}
		}
		return i.newErrorWithLocation(expr, "type mismatch: %s %s %s", left.Type(), expr.Operator, right.Type())

	case left.Type() == "BOOLEAN" && right.Type() == "BOOLEAN":
		return i.evalBooleanBinaryOp(expr.Operator, left, right)

	// Handle Enum comparisons (=, <>, <, >, <=, >=)
	case left.Type() == "ENUM" && right.Type() == "ENUM":
		return i.evalEnumBinaryOp(expr.Operator, left, right)

	// Handle object and nil comparisons (=, <>)
	case expr.Operator == "=" || expr.Operator == "<>":
		// Check if either operand is nil or an object instance
		_, leftIsNil := left.(*NilValue)
		_, rightIsNil := right.(*NilValue)
		_, leftIsObj := left.(*ObjectInstance)
		_, rightIsObj := right.(*ObjectInstance)
		leftIntf, leftIsIntf := left.(*InterfaceInstance)
		rightIntf, rightIsIntf := right.(*InterfaceInstance)
		leftClass, leftIsClass := left.(*ClassValue)
		rightClass, rightIsClass := right.(*ClassValue)

		// Handle RTTITypeInfoValue comparisons (for TypeOf results)
		leftRTTI, leftIsRTTI := left.(*RTTITypeInfoValue)
		rightRTTI, rightIsRTTI := right.(*RTTITypeInfoValue)
		if leftIsRTTI && rightIsRTTI {
			// Compare by TypeID (unique identifier for each type)
			result := leftRTTI.TypeID == rightRTTI.TypeID
			if expr.Operator == "=" {
				return &BooleanValue{Value: result}
			} else {
				return &BooleanValue{Value: !result}
			}
		}

		// Handle ClassValue (metaclass) comparisons
		// meta = TBase, meta <> TChild, etc.
		if leftIsClass || rightIsClass {
			// Both are ClassValue - compare by ClassInfo identity
			if leftIsClass && rightIsClass {
				result := leftClass.ClassInfo == rightClass.ClassInfo
				if expr.Operator == "=" {
					return &BooleanValue{Value: result}
				} else {
					return &BooleanValue{Value: !result}
				}
			}
			// One is ClassValue, one is nil
			if leftIsNil || rightIsNil {
				if expr.Operator == "=" {
					return &BooleanValue{Value: false}
				} else {
					return &BooleanValue{Value: true}
				}
			}
		}

		// Handle InterfaceInstance comparisons
		// Two interfaces are equal if they wrap the same underlying object
		if leftIsIntf || rightIsIntf {
			// Both are interfaces - compare underlying objects
			if leftIsIntf && rightIsIntf {
				// Compare by underlying object identity
				result := leftIntf.Object == rightIntf.Object
				if expr.Operator == "=" {
					return &BooleanValue{Value: result}
				} else {
					return &BooleanValue{Value: !result}
				}
			}
			// One is interface, one is nil
			if leftIsNil || rightIsNil {
				// Interface is nil if its Object is nil
				var intfIsNil bool
				if leftIsIntf {
					intfIsNil = leftIntf.Object == nil
				} else {
					intfIsNil = rightIntf.Object == nil
				}
				if expr.Operator == "=" {
					return &BooleanValue{Value: intfIsNil}
				} else {
					return &BooleanValue{Value: !intfIsNil}
				}
			}
		}

		// If either is nil or an object, do object identity comparison
		if leftIsNil || rightIsNil || leftIsObj || rightIsObj {
			// Both nil
			if leftIsNil && rightIsNil {
				if expr.Operator == "=" {
					return &BooleanValue{Value: true}
				} else {
					return &BooleanValue{Value: false}
				}
			}

			// One is nil, one is not
			if leftIsNil || rightIsNil {
				if expr.Operator == "=" {
					return &BooleanValue{Value: false}
				} else {
					return &BooleanValue{Value: true}
				}
			}

			// Both are objects - compare by identity
			if expr.Operator == "=" {
				return &BooleanValue{Value: left == right}
			} else {
				return &BooleanValue{Value: left != right}
			}
		}

		// Check if both are records (by type assertion, not string comparison)
		// Since RecordValue.Type() now returns actual type name (e.g., "TPoint"), not "RECORD"
		if _, leftIsRecord := left.(*RecordValue); leftIsRecord {
			if _, rightIsRecord := right.(*RecordValue); rightIsRecord {
				return i.evalRecordBinaryOp(expr.Operator, left, right)
			}
		}

		// Not object/nil/record comparison - return error
		return i.newErrorWithLocation(expr, "type mismatch: %s %s %s", left.Type(), expr.Operator, right.Type())

	default:
		return i.newErrorWithLocation(expr, "type mismatch: %s %s %s", left.Type(), expr.Operator, right.Type())
	}
}

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
	if entry, found := i.globalOperators.lookup(operator, operandTypes); found {
		return i.invokeRuntimeOperator(entry, operands, node), true
	}
	return nil, false
}

// evalCoalesceExpression evaluates the coalesce operator (??) with short-circuit evaluation.
// Returns the left operand if it's "truthy", otherwise evaluates and returns the right operand.
func (i *Interpreter) evalCoalesceExpression(expr *ast.BinaryExpression) Value {
	// Evaluate left operand
	left := i.Eval(expr.Left)
	if isError(left) {
		return left
	}
	if left == nil {
		return i.newErrorWithLocation(expr.Left, "left operand evaluated to nil")
	}

	// Check if left is "falsey" (default/zero value for its type)
	if !isFalsey(left) {
		// Left is truthy, return it (don't evaluate right)
		return left
	}

	// Left is falsey, evaluate and return right operand
	right := i.Eval(expr.Right)
	if isError(right) {
		return right
	}
	if right == nil {
		return i.newErrorWithLocation(expr.Right, "right operand evaluated to nil")
	}

	return right
}

// evalAndExpression evaluates the 'and' operator with short-circuit evaluation for booleans.
// For integers, it falls back to normal evaluation for bitwise AND.
func (i *Interpreter) evalAndExpression(expr *ast.BinaryExpression) Value {
	// Peek at the left operand type to determine if we need short-circuit evaluation
	// We need to evaluate it first to check the type
	left := i.Eval(expr.Left)
	if isError(left) {
		return left
	}
	if left == nil {
		return i.newErrorWithLocation(expr.Left, "left operand evaluated to nil")
	}

	// If left is a boolean, use short-circuit evaluation
	if leftBool, ok := left.(*BooleanValue); ok {
		// If left is false, return false immediately (don't evaluate right)
		if !leftBool.Value {
			return &BooleanValue{Value: false}
		}

		// Left is true, evaluate right operand
		right := i.Eval(expr.Right)
		if isError(right) {
			return right
		}
		if right == nil {
			return i.newErrorWithLocation(expr.Right, "right operand evaluated to nil")
		}

		// Handle Variant on the right: unwrap and coerce to boolean
		var rightBoolValue bool
		var rightIsVariant bool
		if right.Type() == "VARIANT" {
			rightBoolValue = variantToBool(right)
			rightIsVariant = true
		} else {
			rightBool, ok := right.(*BooleanValue)
			if !ok {
				return i.newErrorWithLocation(expr.Right, "expected boolean, got %s", right.Type())
			}
			rightBoolValue = rightBool.Value
		}

		result := &BooleanValue{Value: rightBoolValue}
		// If right operand was a Variant, wrap the result in a Variant
		if rightIsVariant {
			return BoxVariant(result)
		}
		return result
	}

	// For non-boolean types (like integers for bitwise AND), evaluate both operands normally
	// and let the normal binary operator handling deal with it
	right := i.Eval(expr.Right)
	if isError(right) {
		return right
	}
	if right == nil {
		return i.newErrorWithLocation(expr.Right, "right operand evaluated to nil")
	}

	// Handle operations based on operand types (this duplicates logic from evalBinaryExpression)
	if result, ok := i.tryBinaryOperator(expr.Operator, left, right, expr); ok {
		return result
	}

	switch {
	case left.Type() == "INTEGER" && right.Type() == "INTEGER":
		return i.evalIntegerBinaryOp(expr.Operator, left, right)
	case left.Type() == "ENUM" && right.Type() == "ENUM":
		// Task 1.6: Support enum boolean operations
		return i.evalEnumBinaryOp(expr.Operator, left, right)
	case left.Type() == "VARIANT" || right.Type() == "VARIANT":
		return i.evalVariantBinaryOp(expr.Operator, left, right, expr)
	default:
		return i.newErrorWithLocation(expr, "type mismatch: %s %s %s", left.Type(), expr.Operator, right.Type())
	}
}

// evalOrExpression evaluates the 'or' operator with short-circuit evaluation for booleans.
// For integers, it falls back to normal evaluation for bitwise OR.
func (i *Interpreter) evalOrExpression(expr *ast.BinaryExpression) Value {
	// Peek at the left operand type to determine if we need short-circuit evaluation
	// We need to evaluate it first to check the type
	left := i.Eval(expr.Left)
	if isError(left) {
		return left
	}
	if left == nil {
		return i.newErrorWithLocation(expr.Left, "left operand evaluated to nil")
	}

	// If left is a boolean, use short-circuit evaluation
	if leftBool, ok := left.(*BooleanValue); ok {
		// If left is true, return true immediately (don't evaluate right)
		if leftBool.Value {
			return &BooleanValue{Value: true}
		}

		// Left is false, evaluate right operand
		right := i.Eval(expr.Right)
		if isError(right) {
			return right
		}
		if right == nil {
			return i.newErrorWithLocation(expr.Right, "right operand evaluated to nil")
		}

		// Handle Variant on the right: unwrap and coerce to boolean
		var rightBoolValue bool
		var rightIsVariant bool
		if right.Type() == "VARIANT" {
			rightBoolValue = variantToBool(right)
			rightIsVariant = true
		} else {
			rightBool, ok := right.(*BooleanValue)
			if !ok {
				return i.newErrorWithLocation(expr.Right, "expected boolean, got %s", right.Type())
			}
			rightBoolValue = rightBool.Value
		}

		result := &BooleanValue{Value: rightBoolValue}
		// If right operand was a Variant, wrap the result in a Variant
		if rightIsVariant {
			return BoxVariant(result)
		}
		return result
	}

	// For non-boolean types (like integers for bitwise OR), evaluate both operands normally
	// and let the normal binary operator handling deal with it
	right := i.Eval(expr.Right)
	if isError(right) {
		return right
	}
	if right == nil {
		return i.newErrorWithLocation(expr.Right, "right operand evaluated to nil")
	}

	// Handle operations based on operand types (this duplicates logic from evalBinaryExpression)
	if result, ok := i.tryBinaryOperator(expr.Operator, left, right, expr); ok {
		return result
	}

	switch {
	case left.Type() == "INTEGER" && right.Type() == "INTEGER":
		return i.evalIntegerBinaryOp(expr.Operator, left, right)
	case left.Type() == "ENUM" && right.Type() == "ENUM":
		// Task 1.6: Support enum boolean operations
		return i.evalEnumBinaryOp(expr.Operator, left, right)
	case left.Type() == "VARIANT" || right.Type() == "VARIANT":
		return i.evalVariantBinaryOp(expr.Operator, left, right, expr)
	default:
		return i.newErrorWithLocation(expr, "type mismatch: %s %s %s", left.Type(), expr.Operator, right.Type())
	}
}

// isFalsey checks if a value is considered "falsey" (default/zero value for its type).
// Falsey values: 0 (integer), 0.0 (float), "" (empty string), false (boolean), nil, empty arrays.
func isFalsey(val Value) bool {
	// Task 9.4.4: Handle nil (from unassigned variants)
	if val == nil {
		return true
	}

	switch v := val.(type) {
	case *IntegerValue:
		return v.Value == 0
	case *FloatValue:
		return v.Value == 0.0
	case *StringValue:
		return v.Value == ""
	case *BooleanValue:
		return !v.Value
	case *NilValue:
		return true
	case *NullValue:
		// Task 9.4.1: Null is always falsey
		return true
	case *UnassignedValue:
		// Task 9.4.1: Unassigned is always falsey
		return true
	case *ArrayValue:
		return len(v.Elements) == 0
	case *VariantValue:
		// Task 9.4.4: Unwrap variant and check inner value
		// If variant.Value is nil, the nil check above will return true
		return isFalsey(v.Value)
	default:
		// Other types (objects, classes, etc.) are truthy if non-nil
		return false
	}
}

// evalIntegerBinaryOp evaluates binary operations on integers.
func (i *Interpreter) evalIntegerBinaryOp(op string, left, right Value) Value {
	// Task 9.4.3: Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Safe type assertions with error handling
	leftInt, ok := left.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected integer, got %s", left.Type())
	}
	rightInt, ok := right.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected integer, got %s", right.Type())
	}

	leftVal := leftInt.Value
	rightVal := rightInt.Value

	switch op {
	case "+":
		return &IntegerValue{Value: leftVal + rightVal}
	case "-":
		return &IntegerValue{Value: leftVal - rightVal}
	case "*":
		return &IntegerValue{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			// Enhanced error with operand values
			return i.NewRuntimeError(
				i.currentNode,
				"division_by_zero",
				"Division by zero",
				map[string]string{
					"left":  fmt.Sprintf("%d", leftVal),
					"right": fmt.Sprintf("%d", rightVal),
				},
			)
		}
		// Integer division in DWScript uses / for float division
		// We'll convert to float for division
		return &FloatValue{Value: float64(leftVal) / float64(rightVal)}
	case "div":
		if rightVal == 0 {
			// Enhanced error with operand values
			return i.NewRuntimeError(
				i.currentNode,
				"division_by_zero",
				"Division by zero",
				map[string]string{
					"left":  fmt.Sprintf("%d", leftVal),
					"right": fmt.Sprintf("%d", rightVal),
				},
			)
		}
		return &IntegerValue{Value: leftVal / rightVal}
	case "mod":
		if rightVal == 0 {
			// Enhanced error with operand values
			return i.NewRuntimeError(
				i.currentNode,
				"modulo_by_zero",
				"Division by zero",
				map[string]string{
					"left":  fmt.Sprintf("%d", leftVal),
					"right": fmt.Sprintf("%d", rightVal),
				},
			)
		}
		return &IntegerValue{Value: leftVal % rightVal}
	case "shl":
		if rightVal < 0 {
			return i.newErrorWithLocation(i.currentNode, "negative shift amount")
		}
		// Shift left - multiply by 2^rightVal
		return &IntegerValue{Value: leftVal << uint(rightVal)}
	case "shr":
		if rightVal < 0 {
			return i.newErrorWithLocation(i.currentNode, "negative shift amount")
		}
		// Shift right - divide by 2^rightVal (logical shift)
		return &IntegerValue{Value: leftVal >> uint(rightVal)}
	case "sar":
		if rightVal < 0 {
			return i.newErrorWithLocation(i.currentNode, "negative shift amount")
		}
		// Arithmetic shift right - sign-preserving shift
		// In Go, >> on signed integers is already arithmetic (sign-preserving)
		return &IntegerValue{Value: leftVal >> uint(rightVal)}
	case "and":
		// Bitwise AND for integers
		return &IntegerValue{Value: leftVal & rightVal}
	case "or":
		// Bitwise OR for integers
		return &IntegerValue{Value: leftVal | rightVal}
	case "xor":
		// Bitwise XOR for integers
		return &IntegerValue{Value: leftVal ^ rightVal}
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &BooleanValue{Value: leftVal >= rightVal}
	default:
		return i.newTypeError(i.currentNode, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalFloatBinaryOp evaluates binary operations on floats.
// Handles mixed integer/float operations by converting to float.
func (i *Interpreter) evalFloatBinaryOp(op string, left, right Value) Value {
	var leftVal, rightVal float64

	// Task 9.4.3: Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Convert left operand to float
	switch v := left.(type) {
	case *FloatValue:
		leftVal = v.Value
	case *IntegerValue:
		leftVal = float64(v.Value)
	default:
		return i.newTypeError(i.currentNode, "type error in float operation: expected FLOAT or INTEGER, got %s", left.Type())
	}

	// Convert right operand to float
	switch v := right.(type) {
	case *FloatValue:
		rightVal = v.Value
	case *IntegerValue:
		rightVal = float64(v.Value)
	default:
		return i.newTypeError(i.currentNode, "type error in float operation: expected FLOAT or INTEGER, got %s", right.Type())
	}

	switch op {
	case "+":
		return &FloatValue{Value: leftVal + rightVal}
	case "-":
		return &FloatValue{Value: leftVal - rightVal}
	case "*":
		return &FloatValue{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			// DWScript returns infinity (or NaN when 0/0) for float division by zero
			if leftVal == 0 {
				return &FloatValue{Value: math.NaN()}
			}
			sign := 1.0
			if leftVal < 0 {
				sign = -1.0
			}
			return &FloatValue{Value: math.Inf(int(sign))}
		}
		return &FloatValue{Value: leftVal / rightVal}
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &BooleanValue{Value: leftVal >= rightVal}
	default:
		return i.newTypeError(i.currentNode, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalStringBinaryOp evaluates binary operations on strings.
func (i *Interpreter) evalStringBinaryOp(op string, left, right Value) Value {
	// Task 9.4.3: Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Safe type assertions with error handling
	leftStr, ok := left.(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected string, got %s", left.Type())
	}
	rightStr, ok := right.(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected string, got %s", right.Type())
	}

	leftVal := leftStr.Value
	rightVal := rightStr.Value

	switch op {
	case "+":
		return &StringValue{Value: leftVal + rightVal}
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &BooleanValue{Value: leftVal >= rightVal}
	default:
		return i.newTypeError(i.currentNode, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalBooleanBinaryOp evaluates binary operations on booleans.
func (i *Interpreter) evalBooleanBinaryOp(op string, left, right Value) Value {
	// Task 9.4.3: Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Safe type assertions with error handling
	leftBool, ok := left.(*BooleanValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected boolean, got %s", left.Type())
	}
	rightBool, ok := right.(*BooleanValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected boolean, got %s", right.Type())
	}

	leftVal := leftBool.Value
	rightVal := rightBool.Value

	switch op {
	case "and":
		return &BooleanValue{Value: leftVal && rightVal}
	case "or":
		return &BooleanValue{Value: leftVal || rightVal}
	case "xor":
		return &BooleanValue{Value: leftVal != rightVal}
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	default:
		return i.newTypeError(i.currentNode, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalEnumBinaryOp evaluates binary operations on enum values.
// Enums support comparison operations (=, <>, <, >, <=, >=) and boolean operations (and, or, xor).
func (i *Interpreter) evalEnumBinaryOp(op string, left, right Value) Value {
	// Safe type assertions with error handling
	leftEnum, ok := left.(*EnumValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected enum, got %s", left.Type())
	}
	rightEnum, ok := right.(*EnumValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected enum, got %s", right.Type())
	}

	leftVal := leftEnum.OrdinalValue
	rightVal := rightEnum.OrdinalValue

	switch op {
	// Comparison operators
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &BooleanValue{Value: leftVal >= rightVal}
	// Task 1.6: Boolean operations for enums (especially flags enums)
	case "and":
		// Bitwise AND on enum ordinal values, return enum of same type
		return &EnumValue{
			TypeName:     leftEnum.TypeName,
			ValueName:    "", // No specific name for computed values
			OrdinalValue: leftVal & rightVal,
		}
	case "or":
		// Bitwise OR on enum ordinal values, return enum of same type
		return &EnumValue{
			TypeName:     leftEnum.TypeName,
			ValueName:    "", // No specific name for computed values
			OrdinalValue: leftVal | rightVal,
		}
	case "xor":
		// Bitwise XOR on enum ordinal values, return enum of same type
		return &EnumValue{
			TypeName:     leftEnum.TypeName,
			ValueName:    "", // No specific name for computed values
			OrdinalValue: leftVal ^ rightVal,
		}
	default:
		return i.newTypeError(i.currentNode, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalVariantBinaryOp evaluates binary operations on Variant values.
//
// Variant operations follow these rules:
//   - Unwrap operands to get actual runtime values
//   - Apply numeric promotion (Integer + Float → Float)
//   - Support string concatenation with + operator
//   - Raise runtime error if types are incompatible
func (i *Interpreter) evalVariantBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Task 9.4.3: Check if either operand is an uninitialized Variant (before unwrapping)
	// An uninitialized variant is a VariantValue with Value == nil.
	// After unwrapping via unwrapVariant(), this nil becomes an UnassignedValue object.
	// This is distinct from a VariantValue explicitly containing an UnassignedValue/NullValue/NilValue.
	leftUnassignedVariant := false
	rightUnassignedVariant := false
	if leftVar, ok := left.(*VariantValue); ok && leftVar.Value == nil {
		leftUnassignedVariant = true
	}
	if rightVar, ok := right.(*VariantValue); ok && rightVar.Value == nil {
		rightUnassignedVariant = true
	}

	// Unwrap Variant values to get the actual runtime values
	leftVal := unwrapVariant(left)
	rightVal := unwrapVariant(right)

	// Task 9.4.1: Check for Null/Unassigned/Nil values (after unwrapping)
	_, leftIsNil := leftVal.(*NilValue)
	_, rightIsNil := rightVal.(*NilValue)
	_, leftIsNull := leftVal.(*NullValue)
	_, rightIsNull := rightVal.(*NullValue)
	_, leftIsUnassigned := leftVal.(*UnassignedValue)
	_, rightIsUnassigned := rightVal.(*UnassignedValue)

	leftIsNullish := leftIsNil || leftIsNull || leftIsUnassigned
	rightIsNullish := rightIsNil || rightIsNull || rightIsUnassigned

	// For comparison operators
	// Task 9.4.3: Complex comparison semantics for Null/Unassigned variants:
	// - Uninitialized variant (VariantValue with Value==nil): equals falsey values (0, false, '', etc.)
	//   and also equals other nullish values (Unassigned, Null, Nil).
	// - Explicit Unassigned/Null/Nil value: only equals other nullish values (Unassigned, Null, Nil),
	//   does NOT equal falsey values.
	//   Example: var v: Variant; (uninitialized) will equal 0, but var v: Variant := Unassigned; (explicitly assigned) will NOT equal 0.
	if op == "=" || op == "<>" {
		// Case 1: Both are nullish (Null/nil/Unassigned) or unassigned variants -> equal
		if (leftIsNullish || leftUnassignedVariant) && (rightIsNullish || rightUnassignedVariant) {
			return &BooleanValue{Value: op == "="}
		}

		// Case 2: One is an UNASSIGNED variant (not just nullish), check if other is falsey
		// Only unassigned variants (not Null/nil) equal falsey values
		if leftUnassignedVariant && !rightIsNullish {
			result := isFalsey(rightVal)
			if op == "=" {
				return &BooleanValue{Value: result}
			}
			return &BooleanValue{Value: !result}
		}
		if rightUnassignedVariant && !leftIsNullish {
			result := isFalsey(leftVal)
			if op == "=" {
				return &BooleanValue{Value: result}
			}
			return &BooleanValue{Value: !result}
		}

		// Case 3: One is nullish (but not unassigned variant), the other is not -> not equal
		if leftIsNullish || rightIsNullish {
			return &BooleanValue{Value: op == "<>"}
		}
	}

	// Error if either operand is nullish for non-comparison operators
	if leftIsNullish {
		return i.newErrorWithLocation(node, "cannot perform operation on unassigned Variant")
	}
	if rightIsNullish {
		return i.newErrorWithLocation(node, "cannot perform operation on unassigned Variant")
	}

	leftType := leftVal.Type()
	rightType := rightVal.Type()

	// Dispatch based on unwrapped types
	switch {
	// Both integers
	case leftType == "INTEGER" && rightType == "INTEGER":
		return i.evalIntegerBinaryOp(op, leftVal, rightVal)

	// Either is float → promote to float
	case leftType == "FLOAT" || rightType == "FLOAT":
		return i.evalFloatBinaryOp(op, leftVal, rightVal)

	// Both strings
	case leftType == "STRING" && rightType == "STRING":
		return i.evalStringBinaryOp(op, leftVal, rightVal)

	// Both booleans
	case leftType == "BOOLEAN" && rightType == "BOOLEAN":
		return i.evalBooleanBinaryOp(op, leftVal, rightVal)

	// String + any type → string concatenation (for + operator only)
	case op == "+" && (leftType == "STRING" || rightType == "STRING"):
		leftStr := i.convertToString(leftVal)
		rightStr := i.convertToString(rightVal)
		return &StringValue{Value: leftStr + rightStr}

	// Numeric type mismatch → try conversion
	case isNumericType(leftType) && isNumericType(rightType):
		// This shouldn't happen since we handle Integer and Float above,
		// but included for completeness
		return i.evalFloatBinaryOp(op, leftVal, rightVal)

	// For comparison operators, try comparing as strings
	case (op == "=" || op == "<>" || op == "<" || op == ">" || op == "<=" || op == ">="):
		// Convert both to strings and compare
		leftStr := i.convertToString(leftVal)
		rightStr := i.convertToString(rightVal)
		return i.evalStringBinaryOp(op, &StringValue{Value: leftStr}, &StringValue{Value: rightStr})

	// For boolean operators with mixed numeric/boolean types, coerce to boolean
	case (op == "and" || op == "or" || op == "xor") &&
		((leftType == "BOOLEAN" && (rightType == "INTEGER" || rightType == "FLOAT")) ||
		 (rightType == "BOOLEAN" && (leftType == "INTEGER" || leftType == "FLOAT")) ||
		 ((leftType == "INTEGER" || leftType == "FLOAT") && (rightType == "INTEGER" || rightType == "FLOAT"))):
		// Coerce both operands to boolean
		leftBool := variantToBool(leftVal)
		rightBool := variantToBool(rightVal)
		result := i.evalBooleanBinaryOp(op, &BooleanValue{Value: leftBool}, &BooleanValue{Value: rightBool})
		// Wrap result in Variant since at least one operand was a Variant
		return BoxVariant(result)

	default:
		return i.newErrorWithLocation(node, "incompatible Variant types for operator %s: %s and %s",
			op, leftType, rightType)
	}
}

// isNumericType checks if a type is numeric (INTEGER or FLOAT).
func isNumericType(typeStr string) bool {
	return typeStr == "INTEGER" || typeStr == "FLOAT"
}

// convertToString converts a Value to its string representation.
// Used for Variant string concatenation and comparison.
func (i *Interpreter) convertToString(val Value) string {
	if val == nil {
		return ""
	}
	return val.String()
}
