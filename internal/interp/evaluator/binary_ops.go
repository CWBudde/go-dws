package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains binary operator evaluation methods for the Evaluator.
// Task 3.5.19: Migrated from Interpreter expressions_binary.go

// ============================================================================
// Short-Circuit Operators
// ============================================================================

// evalCoalesceOp evaluates the coalesce operator (??) with short-circuit evaluation.
// Returns the left operand if it's "truthy", otherwise evaluates and returns the right operand.
func (e *Evaluator) evalCoalesceOp(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
	// Evaluate left operand
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}
	if left == nil {
		return e.newError(node.Left, "left operand evaluated to nil")
	}

	// Check if left is "falsey" (default/zero value for its type)
	if !isFalsey(left) {
		// Left is truthy, return it (don't evaluate right)
		return left
	}

	// Left is falsey, evaluate and return right operand
	right := e.Eval(node.Right, ctx)
	if isError(right) {
		return right
	}
	if right == nil {
		return e.newError(node.Right, "right operand evaluated to nil")
	}

	return right
}

// evalAndOp evaluates the 'and' operator with short-circuit evaluation.
// For booleans: left and right (short-circuit if left is false).
// For integers: bitwise AND.
func (e *Evaluator) evalAndOp(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
	// Evaluate left operand
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}
	if left == nil {
		return e.newError(node.Left, "left operand evaluated to nil")
	}

	// For integers, 'and' is bitwise AND (not short-circuit)
	if left.Type() == "INTEGER" {
		// Need to evaluate right operand for bitwise operation
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}
		return e.evalIntegerBinaryOp("and", left, right, node)
	}

	// For booleans, short-circuit evaluation
	if left.Type() == "BOOLEAN" {
		leftBool, ok := left.(*runtime.BooleanValue)
		if !ok {
			return e.newError(node.Left, "expected boolean for 'and' operator")
		}

		// Short-circuit: if left is false, return false without evaluating right
		if !leftBool.Value {
			return &runtime.BooleanValue{Value: false}
		}

		// Left is true, evaluate right
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}

		rightBool, ok := right.(*runtime.BooleanValue)
		if !ok {
			return e.newError(node.Right, "expected boolean for 'and' operator")
		}

		return &runtime.BooleanValue{Value: rightBool.Value}
	}

	// Handle Variant types
	if left.Type() == "VARIANT" {
		// Unwrap and try again
		left = unwrapVariant(left)
		if left.Type() == "BOOLEAN" {
			leftBool, ok := left.(*runtime.BooleanValue)
			if ok && !leftBool.Value {
				// Short-circuit
				return &runtime.BooleanValue{Value: false}
			}
		}
		// Fall through to evaluate right operand
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}
		return e.evalVariantBinaryOp("and", left, right, node)
	}

	// Handle Enum types - bitwise AND (not short-circuit)
	if left.Type() == "ENUM" {
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}
		return e.evalEnumBinaryOp("and", left, right, node)
	}

	return e.newError(node, "type mismatch: 'and' operator requires boolean or integer operands")
}

// evalOrOp evaluates the 'or' operator with short-circuit evaluation.
// For booleans: left or right (short-circuit if left is true).
// For integers: bitwise OR.
func (e *Evaluator) evalOrOp(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
	// Evaluate left operand
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}
	if left == nil {
		return e.newError(node.Left, "left operand evaluated to nil")
	}

	// For integers, 'or' is bitwise OR (not short-circuit)
	if left.Type() == "INTEGER" {
		// Need to evaluate right operand for bitwise operation
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}
		return e.evalIntegerBinaryOp("or", left, right, node)
	}

	// For booleans, short-circuit evaluation
	if left.Type() == "BOOLEAN" {
		leftBool, ok := left.(*runtime.BooleanValue)
		if !ok {
			return e.newError(node.Left, "expected boolean for 'or' operator")
		}

		// Short-circuit: if left is true, return true without evaluating right
		if leftBool.Value {
			return &runtime.BooleanValue{Value: true}
		}

		// Left is false, evaluate right
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}

		rightBool, ok := right.(*runtime.BooleanValue)
		if !ok {
			return e.newError(node.Right, "expected boolean for 'or' operator")
		}

		return &runtime.BooleanValue{Value: rightBool.Value}
	}

	// Handle Variant types
	if left.Type() == "VARIANT" {
		// Unwrap and try again
		left = unwrapVariant(left)
		if left.Type() == "BOOLEAN" {
			leftBool, ok := left.(*runtime.BooleanValue)
			if ok && leftBool.Value {
				// Short-circuit
				return &runtime.BooleanValue{Value: true}
			}
		}
		// Fall through to evaluate right operand
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}
		return e.evalVariantBinaryOp("or", left, right, node)
	}

	// Handle Enum types - bitwise OR (not short-circuit)
	if left.Type() == "ENUM" {
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}
		return e.evalEnumBinaryOp("or", left, right, node)
	}

	return e.newError(node, "type mismatch: 'or' operator requires boolean or integer operands")
}

// ============================================================================
// Type-Specific Binary Operations
// ============================================================================

// evalIntegerBinaryOp evaluates binary operations on integers.
func (e *Evaluator) evalIntegerBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Safe type assertions with error handling
	leftInt, ok := left.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "expected integer, got %s", left.Type())
	}
	rightInt, ok := right.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "expected integer, got %s", right.Type())
	}

	leftVal := leftInt.Value
	rightVal := rightInt.Value

	switch op {
	case "+":
		return &runtime.IntegerValue{Value: leftVal + rightVal}
	case "-":
		return &runtime.IntegerValue{Value: leftVal - rightVal}
	case "*":
		return &runtime.IntegerValue{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return e.newError(node, "division by zero: %d / %d", leftVal, rightVal)
		}
		// Integer division in DWScript uses / for float division
		return &runtime.FloatValue{Value: float64(leftVal) / float64(rightVal)}
	case "div":
		if rightVal == 0 {
			return e.newError(node, "division by zero: %d div %d", leftVal, rightVal)
		}
		return &runtime.IntegerValue{Value: leftVal / rightVal}
	case "mod":
		if rightVal == 0 {
			return e.newError(node, "modulo by zero: %d mod %d", leftVal, rightVal)
		}
		return &runtime.IntegerValue{Value: leftVal % rightVal}
	case "shl":
		if rightVal < 0 {
			return e.newError(node, "negative shift amount")
		}
		return &runtime.IntegerValue{Value: leftVal << uint(rightVal)}
	case "shr":
		if rightVal < 0 {
			return e.newError(node, "negative shift amount")
		}
		return &runtime.IntegerValue{Value: leftVal >> uint(rightVal)}
	case "sar":
		if rightVal < 0 {
			return e.newError(node, "negative shift amount")
		}
		// Arithmetic shift right (sign-preserving)
		return &runtime.IntegerValue{Value: leftVal >> uint(rightVal)}
	case "and":
		// Bitwise AND for integers
		return &runtime.IntegerValue{Value: leftVal & rightVal}
	case "or":
		// Bitwise OR for integers
		return &runtime.IntegerValue{Value: leftVal | rightVal}
	case "xor":
		// Bitwise XOR for integers
		return &runtime.IntegerValue{Value: leftVal ^ rightVal}
	case "=":
		return &runtime.BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &runtime.BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &runtime.BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &runtime.BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &runtime.BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &runtime.BooleanValue{Value: leftVal >= rightVal}
	default:
		return e.newError(node, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalFloatBinaryOp evaluates binary operations on floats.
// Handles mixed integer/float operations by converting to float.
func (e *Evaluator) evalFloatBinaryOp(op string, left, right Value, node ast.Node) Value {
	var leftVal, rightVal float64

	// Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Convert left operand to float
	switch v := left.(type) {
	case *runtime.FloatValue:
		leftVal = v.Value
	case *runtime.IntegerValue:
		leftVal = float64(v.Value)
	default:
		return e.newError(node, "type error in float operation: expected FLOAT or INTEGER, got %s", left.Type())
	}

	// Convert right operand to float
	switch v := right.(type) {
	case *runtime.FloatValue:
		rightVal = v.Value
	case *runtime.IntegerValue:
		rightVal = float64(v.Value)
	default:
		return e.newError(node, "type error in float operation: expected FLOAT or INTEGER, got %s", right.Type())
	}

	switch op {
	case "+":
		return &runtime.FloatValue{Value: leftVal + rightVal}
	case "-":
		return &runtime.FloatValue{Value: leftVal - rightVal}
	case "*":
		return &runtime.FloatValue{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return e.newError(node, "division by zero: %v / %v", leftVal, rightVal)
		}
		return &runtime.FloatValue{Value: leftVal / rightVal}
	case "=":
		return &runtime.BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &runtime.BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &runtime.BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &runtime.BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &runtime.BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &runtime.BooleanValue{Value: leftVal >= rightVal}
	default:
		return e.newError(node, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalStringBinaryOp evaluates binary operations on strings.
func (e *Evaluator) evalStringBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Safe type assertions with error handling
	leftStr, ok := left.(*runtime.StringValue)
	if !ok {
		return e.newError(node, "expected string, got %s", left.Type())
	}
	rightStr, ok := right.(*runtime.StringValue)
	if !ok {
		return e.newError(node, "expected string, got %s", right.Type())
	}

	leftVal := leftStr.Value
	rightVal := rightStr.Value

	switch op {
	case "+":
		return &runtime.StringValue{Value: leftVal + rightVal}
	case "=":
		return &runtime.BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &runtime.BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &runtime.BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &runtime.BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &runtime.BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &runtime.BooleanValue{Value: leftVal >= rightVal}
	default:
		return e.newError(node, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalBooleanBinaryOp evaluates binary operations on booleans.
func (e *Evaluator) evalBooleanBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Safe type assertions with error handling
	leftBool, ok := left.(*runtime.BooleanValue)
	if !ok {
		return e.newError(node, "expected boolean, got %s", left.Type())
	}
	rightBool, ok := right.(*runtime.BooleanValue)
	if !ok {
		return e.newError(node, "expected boolean, got %s", right.Type())
	}

	leftVal := leftBool.Value
	rightVal := rightBool.Value

	switch op {
	case "and":
		return &runtime.BooleanValue{Value: leftVal && rightVal}
	case "or":
		return &runtime.BooleanValue{Value: leftVal || rightVal}
	case "xor":
		return &runtime.BooleanValue{Value: leftVal != rightVal}
	case "=":
		return &runtime.BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &runtime.BooleanValue{Value: leftVal != rightVal}
	default:
		return e.newError(node, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalEnumBinaryOp evaluates binary operations on enum values.
// Enums support comparison operations (=, <>, <, >, <=, >=) and bitwise operations (and, or, xor).
// Enums are compared by their ordinal values.
// Task 3.5.18: Migrated from Interpreter expressions_binary.go
func (e *Evaluator) evalEnumBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Safe type assertions with error handling
	leftEnum, ok := left.(*runtime.EnumValue)
	if !ok {
		return e.newError(node, "expected enum, got %s", left.Type())
	}
	rightEnum, ok := right.(*runtime.EnumValue)
	if !ok {
		return e.newError(node, "expected enum, got %s", right.Type())
	}

	leftVal := leftEnum.OrdinalValue
	rightVal := rightEnum.OrdinalValue

	switch op {
	// Comparison operators
	case "=":
		return &runtime.BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &runtime.BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &runtime.BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &runtime.BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &runtime.BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &runtime.BooleanValue{Value: leftVal >= rightVal}
	// Bitwise operations for enums (especially flags enums)
	case "and":
		// Bitwise AND on enum ordinal values, return enum of same type
		return &runtime.EnumValue{
			TypeName:     leftEnum.TypeName,
			ValueName:    "", // No specific name for computed values
			OrdinalValue: leftVal & rightVal,
		}
	case "or":
		// Bitwise OR on enum ordinal values, return enum of same type
		return &runtime.EnumValue{
			TypeName:     leftEnum.TypeName,
			ValueName:    "", // No specific name for computed values
			OrdinalValue: leftVal | rightVal,
		}
	case "xor":
		// Bitwise XOR on enum ordinal values, return enum of same type
		return &runtime.EnumValue{
			TypeName:     leftEnum.TypeName,
			ValueName:    "", // No specific name for computed values
			OrdinalValue: leftVal ^ rightVal,
		}
	default:
		return e.newError(node, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// ============================================================================
// Complex Type Comparisons
// ============================================================================

// evalEqualityComparison handles = and <> operators for complex types.
// Supports: nil, objects, interfaces, classes, RTTI, sets, arrays, records.
// Task 3.5.103d: Migrated from Interpreter.EvalEqualityComparison.
func (e *Evaluator) evalEqualityComparison(op string, left, right Value, node ast.Node) Value {
	// Check type names to identify complex types
	leftType := left.Type()
	rightType := right.Type()

	// Handle nil comparisons
	if leftType == "NIL" || rightType == "NIL" {
		// Both nil
		if leftType == "NIL" && rightType == "NIL" {
			if op == "=" {
				return &runtime.BooleanValue{Value: true}
			}
			return &runtime.BooleanValue{Value: false}
		}

		// One is nil, one is not - handle interface special case
		if leftType == "INTERFACE" || rightType == "INTERFACE" {
			// Check if interface wraps nil object
			var intfIsNil bool
			if leftType == "INTERFACE" {
				// Interface on left, nil on right
				// Use string representation check (interface with nil object shows as "nil")
				intfIsNil = left.String() == "nil"
			} else {
				// Nil on left, interface on right
				intfIsNil = right.String() == "nil"
			}
			if op == "=" {
				return &runtime.BooleanValue{Value: intfIsNil}
			}
			return &runtime.BooleanValue{Value: !intfIsNil}
		}

		// Standard nil comparison (one nil, one not)
		if op == "=" {
			return &runtime.BooleanValue{Value: false}
		}
		return &runtime.BooleanValue{Value: true}
	}

	// Handle RTTITypeInfoValue comparisons (TypeOf results)
	if leftType == "RTTI_TYPE_INFO" && rightType == "RTTI_TYPE_INFO" {
		// Compare using string representation (contains TypeID)
		result := left.String() == right.String()
		if op == "=" {
			return &runtime.BooleanValue{Value: result}
		}
		return &runtime.BooleanValue{Value: !result}
	}

	// Handle ClassValue (metaclass) comparisons
	// ClassValue Type() returns "CLASS[ClassName]"
	leftIsClass := len(leftType) > 6 && leftType[:6] == "CLASS[" && leftType[len(leftType)-1] == ']'
	rightIsClass := len(rightType) > 6 && rightType[:6] == "CLASS[" && rightType[len(rightType)-1] == ']'

	if leftIsClass || rightIsClass {
		// Both are ClassValue - compare by string representation
		if leftIsClass && rightIsClass {
			result := left.String() == right.String()
			if op == "=" {
				return &runtime.BooleanValue{Value: result}
			}
			return &runtime.BooleanValue{Value: !result}
		}
		// One is ClassValue, one is nil - already handled above
		if op == "=" {
			return &runtime.BooleanValue{Value: false}
		}
		return &runtime.BooleanValue{Value: true}
	}

	// Handle InterfaceInstance comparisons
	if leftType == "INTERFACE" && rightType == "INTERFACE" {
		// Compare underlying objects by string representation
		result := left.String() == right.String()
		if op == "=" {
			return &runtime.BooleanValue{Value: result}
		}
		return &runtime.BooleanValue{Value: !result}
	}

	// Handle Object comparisons (Type() returns "OBJECT[ClassName]")
	leftIsObj := len(leftType) > 7 && leftType[:7] == "OBJECT[" && leftType[len(leftType)-1] == ']'
	rightIsObj := len(rightType) > 7 && rightType[:7] == "OBJECT[" && rightType[len(rightType)-1] == ']'

	if leftIsObj || rightIsObj {
		// Object identity comparison - compare by pointer equality
		// Use string representation which includes object address
		result := left.String() == right.String()
		if op == "=" {
			return &runtime.BooleanValue{Value: result}
		}
		return &runtime.BooleanValue{Value: !result}
	}

	// Handle Record comparisons
	// RecordValue Type() returns record type name or "RECORD"
	// We check both for named records and anonymous "RECORD" type
	leftIsRecord := leftType == "RECORD" || (leftType != "" && !isSimpleType(leftType))
	rightIsRecord := rightType == "RECORD" || (rightType != "" && !isSimpleType(rightType))

	if leftIsRecord && rightIsRecord {
		// Use RecordsEqual helper (currently uses string comparison)
		result := RecordsEqual(left, right)
		if op == "=" {
			return &runtime.BooleanValue{Value: result}
		}
		return &runtime.BooleanValue{Value: !result}
	}

	// Not a supported equality comparison type - use ValuesEqual as fallback
	result := ValuesEqual(left, right)
	if op == "=" {
		return &runtime.BooleanValue{Value: result}
	}
	return &runtime.BooleanValue{Value: !result}
}

// isSimpleType checks if a type name represents a simple value type.
func isSimpleType(typeName string) bool {
	switch typeName {
	case "INTEGER", "FLOAT", "STRING", "BOOLEAN", "NIL", "SET", "ARRAY", "VARIANT", "ENUM":
		return true
	default:
		return false
	}
}

// ============================================================================
// Operator Overloading and Special Operators
// ============================================================================

// tryBinaryOperator attempts to use custom operator overloading.
// Returns (result, true) if operator found, or (nil, false) if not found.
func (e *Evaluator) tryBinaryOperator(operator string, left, right Value, node ast.Node) (Value, bool) {
	// Operator overloading requires access to:
	// - ObjectInstance.Class.lookupOperator() for instance operators
	// - globalOperators registry for global operators
	// These are in interp package and haven't been migrated yet
	// Delegate to adapter for now
	return nil, false
}

// evalInOperator evaluates the 'in' operator for membership testing.
// Supports: arrays, sets, strings, subranges.
// Task 3.5.103a: Migrated from Interpreter expressions_complex.go:15-63
func (e *Evaluator) evalInOperator(value, container Value, node ast.Node) Value {
	// Handle set membership
	if setVal, ok := container.(*runtime.SetValue); ok {
		// Value must be an ordinal type to be in a set
		ordinal, err := GetOrdinalValue(value)
		if err != nil {
			return e.newError(node, "type mismatch: %s", err.Error())
		}
		// Check if the element is in the set using the ordinal value
		isInSet := setVal.HasElement(ordinal)
		return &runtime.BooleanValue{Value: isInSet}
	}

	// Handle string character/substring membership: 'x' in 'abc'
	if strContainer, ok := container.(*runtime.StringValue); ok {
		// Value must be a string (character)
		strValue, ok := value.(*runtime.StringValue)
		if !ok {
			return e.newError(node, "type mismatch: %s in STRING", value.Type())
		}
		// Check if the container string contains the value string
		if len(strValue.Value) > 0 {
			// Check substring containment
			contains := strings.Contains(strContainer.Value, strValue.Value)
			return &runtime.BooleanValue{Value: contains}
		}
		// Empty string is not in any string
		return &runtime.BooleanValue{Value: false}
	}

	// Handle array membership
	if arrVal, ok := container.(*runtime.ArrayValue); ok {
		// Search for the value in the array using proper equality comparison
		// Task 3.5.103c: Use ValuesEqual helper for comprehensive equality
		for _, elem := range arrVal.Elements {
			if ValuesEqual(value, elem) {
				return &runtime.BooleanValue{Value: true}
			}
		}
		// Value not found in array
		return &runtime.BooleanValue{Value: false}
	}

	return e.newError(node, "type mismatch: %s in %s", value.Type(), container.Type())
}

// evalSetBinaryOp evaluates binary operations on sets.
// Supports: + (union), - (difference), * (intersection).
// Task 3.5.103b: Migrated from Interpreter set.go:259-338
func (e *Evaluator) evalSetBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Type assert to SetValue
	leftSet, leftOk := left.(*runtime.SetValue)
	rightSet, rightOk := right.(*runtime.SetValue)

	if !leftOk || !rightOk {
		return e.newError(node, "set operations require set operands")
	}

	if leftSet == nil || rightSet == nil {
		return e.newError(node, "nil set operand")
	}

	// Verify both sets are of the same type
	if !leftSet.SetType.Equals(rightSet.SetType) {
		return e.newError(node, "type mismatch in set operation: %s vs %s",
			leftSet.SetType.String(), rightSet.SetType.String())
	}

	// Create result set with same type
	result := runtime.NewSetValue(leftSet.SetType)

	// Choose operation based on storage kind
	switch leftSet.SetType.StorageKind {
	case types.SetStorageBitmask:
		// Fast bitwise operations for bitmask storage
		var resultElements uint64

		switch op {
		case "+":
			// Union: bitwise OR
			resultElements = leftSet.Elements | rightSet.Elements

		case "-":
			// Difference: bitwise AND NOT
			resultElements = leftSet.Elements &^ rightSet.Elements

		case "*":
			// Intersection: bitwise AND
			resultElements = leftSet.Elements & rightSet.Elements

		default:
			return e.newError(node, "unsupported set operation: %s", op)
		}

		result.Elements = resultElements

	case types.SetStorageMap:
		// Map-based operations for large sets
		switch op {
		case "+":
			// Union: add all elements from both sets
			for ordinal := range leftSet.MapStore {
				result.MapStore[ordinal] = true
			}
			for ordinal := range rightSet.MapStore {
				result.MapStore[ordinal] = true
			}

		case "-":
			// Difference: elements in left but not in right
			for ordinal := range leftSet.MapStore {
				if !rightSet.MapStore[ordinal] {
					result.MapStore[ordinal] = true
				}
			}

		case "*":
			// Intersection: elements in both sets
			for ordinal := range leftSet.MapStore {
				if rightSet.MapStore[ordinal] {
					result.MapStore[ordinal] = true
				}
			}

		default:
			return e.newError(node, "unsupported set operation: %s", op)
		}
	}

	return result
}

// evalVariantBinaryOp handles binary operations with Variant operands.
// Task 3.5.103f: Migrated from Interpreter expressions_binary.go:746-863.
//
// Variant operations follow these rules:
//   - Unwrap operands to get actual runtime values
//   - Apply numeric promotion (Integer + Float → Float)
//   - Support string concatenation with + operator
//   - Raise runtime error if types are incompatible
//   - Special handling for uninitialized vs explicitly nullish variants
func (e *Evaluator) evalVariantBinaryOp(op string, left, right Value, node ast.Node) Value {
	// An uninitialized variant has Value == nil (detected via IsUninitialized).
	// This is distinct from a VariantValue explicitly containing an UnassignedValue/NullValue/NilValue.
	leftUnassignedVariant := false
	rightUnassignedVariant := false

	if wrapper, ok := left.(runtime.VariantWrapper); ok {
		leftUnassignedVariant = wrapper.IsUninitialized()
	}
	if wrapper, ok := right.(runtime.VariantWrapper); ok {
		rightUnassignedVariant = wrapper.IsUninitialized()
	}

	// Unwrap Variant values to get the actual runtime values
	leftVal := unwrapVariant(left)
	rightVal := unwrapVariant(right)

	// Check for Null/Unassigned/Nil values (after unwrapping)
	leftIsNullish := isNullish(leftVal)
	rightIsNullish := isNullish(rightVal)

	// For comparison operators
	// Complex comparison semantics for Null/Unassigned variants:
	// - Uninitialized variant (VariantValue with Value==nil): equals falsey values (0, false, '', etc.)
	//   and also equals other nullish values (Unassigned, Null, Nil).
	// - Explicit Unassigned/Null/Nil value: only equals other nullish values (Unassigned, Null, Nil),
	//   does NOT equal falsey values.
	if op == "=" || op == "<>" {
		// Case 1: Both are nullish (Null/nil/Unassigned) or unassigned variants -> equal
		if (leftIsNullish || leftUnassignedVariant) && (rightIsNullish || rightUnassignedVariant) {
			return &runtime.BooleanValue{Value: op == "="}
		}

		// Case 2: One is an UNASSIGNED variant (not just nullish), check if other is falsey
		// Only unassigned variants (not Null/nil) equal falsey values
		if leftUnassignedVariant && !rightIsNullish {
			result := isFalsey(rightVal)
			if op == "=" {
				return &runtime.BooleanValue{Value: result}
			}
			return &runtime.BooleanValue{Value: !result}
		}
		if rightUnassignedVariant && !leftIsNullish {
			result := isFalsey(leftVal)
			if op == "=" {
				return &runtime.BooleanValue{Value: result}
			}
			return &runtime.BooleanValue{Value: !result}
		}

		// Case 3: One is nullish (but not unassigned variant), the other is not -> not equal
		if leftIsNullish || rightIsNullish {
			return &runtime.BooleanValue{Value: op == "<>"}
		}
	}

	// Error if either operand is nullish for non-comparison operators
	if leftIsNullish {
		return e.newError(node, "cannot perform operation on unassigned Variant")
	}
	if rightIsNullish {
		return e.newError(node, "cannot perform operation on unassigned Variant")
	}

	leftType := leftVal.Type()
	rightType := rightVal.Type()

	// Dispatch based on unwrapped types
	switch {
	// Both integers
	case leftType == "INTEGER" && rightType == "INTEGER":
		return e.evalIntegerBinaryOp(op, leftVal, rightVal, node)

	// Either is float → promote to float
	case leftType == "FLOAT" || rightType == "FLOAT":
		return e.evalFloatBinaryOp(op, leftVal, rightVal, node)

	// Both strings
	case leftType == "STRING" && rightType == "STRING":
		return e.evalStringBinaryOp(op, leftVal, rightVal, node)

	// Both booleans
	case leftType == "BOOLEAN" && rightType == "BOOLEAN":
		return e.evalBooleanBinaryOp(op, leftVal, rightVal, node)

	// String + any type → string concatenation (for + operator only)
	case op == "+" && (leftType == "STRING" || rightType == "STRING"):
		leftStr := convertToString(leftVal)
		rightStr := convertToString(rightVal)
		return &runtime.StringValue{Value: leftStr + rightStr}

	// Numeric type mismatch → try conversion
	case isNumericTypeName(leftType) && isNumericTypeName(rightType):
		// This shouldn't happen since we handle Integer and Float above,
		// but included for completeness
		return e.evalFloatBinaryOp(op, leftVal, rightVal, node)

	// For comparison operators, try comparing as strings
	case op == "=" || op == "<>" || op == "<" || op == ">" || op == "<=" || op == ">=":
		// Convert both to strings and compare
		leftStr := convertToString(leftVal)
		rightStr := convertToString(rightVal)
		return e.evalStringBinaryOp(op, &runtime.StringValue{Value: leftStr}, &runtime.StringValue{Value: rightStr}, node)

	default:
		return e.newError(node, "incompatible Variant types for operator %s: %s and %s",
			op, leftType, rightType)
	}
}

// isNullish checks if a value represents a null/unassigned/nil state.
// Task 3.5.103f: Helper for variant comparison semantics.
func isNullish(val Value) bool {
	if val == nil {
		return true
	}
	switch val.Type() {
	case "NIL", "NULL", "UNASSIGNED":
		return true
	default:
		return false
	}
}

// convertToString converts a Value to its string representation.
// Task 3.5.103f: Helper for Variant string concatenation and comparison.
func convertToString(val Value) string {
	if val == nil {
		return ""
	}
	return val.String()
}

// isNumericTypeName checks if a type name string is numeric (INTEGER or FLOAT).
// Task 3.5.103f: Helper for variant type coercion.
func isNumericTypeName(typeStr string) bool {
	return typeStr == "INTEGER" || typeStr == "FLOAT"
}

// ============================================================================
// Unary Operators
// ============================================================================
// Task 3.5.20: Migrated from Interpreter expressions_basic.go

// tryUnaryOperator attempts to use custom operator overloading for unary operators.
// Returns (result, true) if operator found, or (nil, false) if not found.
func (e *Evaluator) tryUnaryOperator(operator string, operand Value, node ast.Node) (Value, bool) {
	// Unary operator overloading requires access to:
	// - ObjectInstance.Class.lookupOperator() for instance operators
	// - globalOperators registry for global operators
	// These are in interp package and haven't been migrated yet
	// Delegate to adapter for now
	return nil, false
}

// evalMinusUnaryOp evaluates the unary minus operator (-x).
// Supports Integer and Float, with Variant unwrapping.
func (e *Evaluator) evalMinusUnaryOp(operand Value, node ast.Node) Value {
	// Unwrap Variant if necessary
	operand = unwrapVariant(operand)

	switch v := operand.(type) {
	case *runtime.IntegerValue:
		return &runtime.IntegerValue{Value: -v.Value}
	case *runtime.FloatValue:
		return &runtime.FloatValue{Value: -v.Value}
	default:
		return e.newError(node, "expected integer or float for unary minus, got %s", operand.Type())
	}
}

// evalPlusUnaryOp evaluates the unary plus operator (+x).
// Identity operation for Integer and Float, with Variant unwrapping.
func (e *Evaluator) evalPlusUnaryOp(operand Value, node ast.Node) Value {
	// Unwrap Variant if necessary
	operand = unwrapVariant(operand)

	switch operand.(type) {
	case *runtime.IntegerValue, *runtime.FloatValue:
		return operand
	default:
		return e.newError(node, "expected integer or float for unary plus, got %s", operand.Type())
	}
}

// evalNotUnaryOp evaluates the not operator.
// For Boolean: logical NOT
// For Integer: bitwise NOT
// For Variant: convert to boolean, negate, wrap result in Variant
// Task 3.5.103e: Migrated Variant NOT from adapter delegation to direct implementation.
func (e *Evaluator) evalNotUnaryOp(operand Value, node ast.Node) Value {
	// Handle Variant: convert to bool using DWScript semantics, negate, wrap in Variant
	// Uses VariantToBool from helpers.go which handles unwrapping and type coercion
	if operand.Type() == "VARIANT" {
		boolResult := VariantToBool(operand)
		// Return the negated result as a Variant containing a Boolean
		return runtime.BoxVariant(&runtime.BooleanValue{Value: !boolResult})
	}

	// Handle boolean NOT
	if boolVal, ok := operand.(*runtime.BooleanValue); ok {
		return &runtime.BooleanValue{Value: !boolVal.Value}
	}

	// Handle bitwise NOT for integers
	if intVal, ok := operand.(*runtime.IntegerValue); ok {
		return &runtime.IntegerValue{Value: ^intVal.Value}
	}

	return e.newError(node, "NOT operator requires Boolean or Integer operand, got %s", operand.Type())
}
