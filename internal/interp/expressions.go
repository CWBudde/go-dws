package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// evalIdentifier looks up an identifier in the environment.
// Special handling for "Self" keyword in method contexts.
// Also handles class variable access from within methods.
func (i *Interpreter) evalIdentifier(node *ast.Identifier) Value {
	// Special case for Self keyword
	if node.Value == "Self" {
		val, ok := i.env.Get("Self")
		if !ok {
			return i.newErrorWithLocation(node, "Self used outside method context")
		}
		return val
	}

	// First, try to find in current environment
	val, ok := i.env.Get(node.Value)
	if ok {
		// Task 7.144: Check if this is an external variable
		if extVar, isExternal := val.(*ExternalVarValue); isExternal {
			return i.newErrorWithLocation(node, "Unsupported external variable access: %s", extVar.Name)
		}

		// Task 9.142: Check if this is a lazy parameter (LazyThunk)
		// If so, force evaluation - each access re-evaluates the expression
		if thunk, isLazy := val.(*LazyThunk); isLazy {
			return thunk.Evaluate()
		}

		return val
	}

	// Not found in environment - check if we're in a method context (Self is bound)
	selfVal, selfOk := i.env.Get("Self")
	if selfOk {
		// We're in an instance method context - check for instance fields first
		if obj, ok := AsObject(selfVal); ok {
			// Check if it's an instance field
			if fieldValue := obj.GetField(node.Value); fieldValue != nil {
				return fieldValue
			}

			// Check if it's a class variable
			if classVarValue, exists := obj.Class.ClassVars[node.Value]; exists {
				return classVarValue
			}

			// Task 9.173: Check if it's a method of the current class
			// This allows methods to reference other methods as method pointers
			if method, exists := obj.Class.Methods[node.Value]; exists {
				// Create a method pointer bound to the current object (self)
				// Build the pointer type
				paramTypes := make([]types.Type, len(method.Parameters))
				for idx, param := range method.Parameters {
					if param.Type != nil {
						paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
					}
				}
				var returnType types.Type
				if method.ReturnType != nil {
					returnType = i.getTypeFromAnnotation(method.ReturnType)
				}
				pointerType := types.NewFunctionPointerType(paramTypes, returnType)

				return NewFunctionPointerValue(method, i.env, obj, pointerType)
			}
		}
	}

	// Check if we're in a class method context (__CurrentClass__ marker)
	currentClassVal, hasCurrentClass := i.env.Get("__CurrentClass__")
	if hasCurrentClass {
		if classInfo, ok := currentClassVal.(*ClassInfoValue); ok {
			// Check if it's a class variable
			if classVarValue, exists := classInfo.ClassInfo.ClassVars[node.Value]; exists {
				return classVarValue
			}
		}
	}

	// Before returning error, check if this is a parameterless function/procedure call
	// In DWScript, you can call parameterless procedures without parentheses: "Test;" instead of "Test();"
	if fn, exists := i.functions[node.Value]; exists && fn != nil {
		// Check if function has zero parameters
		if len(fn.Parameters) == 0 {
			// Auto-invoke the parameterless function/procedure
			return i.callUserFunction(fn, []Value{})
		}
	}

	// Task 9.132: Check if this is a parameterless built-in function
	// Built-in functions like PrintLn can be called without parentheses
	if i.isBuiltinFunction(node.Value) {
		// Call the built-in function with no arguments
		return i.callBuiltinFunction(node.Value, []Value{})
	}

	// Still not found - return error
	return i.newErrorWithLocation(node, "undefined variable '%s'", node.Value)
}

// evalBinaryExpression evaluates a binary expression.
func (i *Interpreter) evalBinaryExpression(expr *ast.BinaryExpression) Value {
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

	if result, ok := i.tryBinaryOperator(expr.Operator, left, right, expr); ok {
		return result
	}

	// Handle 'in' operator for array membership checking
	if expr.Operator == "in" {
		return i.evalInOperator(left, right, expr)
	}

	// Handle operations based on operand types
	switch {
	case left.Type() == "INTEGER" && right.Type() == "INTEGER":
		return i.evalIntegerBinaryOp(expr.Operator, left, right)

	case left.Type() == "FLOAT" || right.Type() == "FLOAT":
		return i.evalFloatBinaryOp(expr.Operator, left, right)

	case left.Type() == "STRING" && right.Type() == "STRING":
		return i.evalStringBinaryOp(expr.Operator, left, right)

	case left.Type() == "BOOLEAN" && right.Type() == "BOOLEAN":
		return i.evalBooleanBinaryOp(expr.Operator, left, right)

	// Task 9.229 & 9.230: Handle Variant operations
	case left.Type() == "VARIANT" || right.Type() == "VARIANT":
		return i.evalVariantBinaryOp(expr.Operator, left, right, expr)

	// Handle object and nil comparisons (=, <>)
	case expr.Operator == "=" || expr.Operator == "<>":
		// Check if either operand is nil or an object instance
		_, leftIsNil := left.(*NilValue)
		_, rightIsNil := right.(*NilValue)
		_, leftIsObj := left.(*ObjectInstance)
		_, rightIsObj := right.(*ObjectInstance)

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
		if entry, found := obj.Class.lookupOperator(operator, operandTypes); found {
			return i.invokeRuntimeOperator(entry, operands, node), true
		}
	}
	if obj, ok := right.(*ObjectInstance); ok {
		if entry, found := obj.Class.lookupOperator(operator, operandTypes); found {
			return i.invokeRuntimeOperator(entry, operands, node), true
		}
	}
	if entry, found := i.globalOperators.lookup(operator, operandTypes); found {
		return i.invokeRuntimeOperator(entry, operands, node), true
	}
	return nil, false
}

// evalIntegerBinaryOp evaluates binary operations on integers.
func (i *Interpreter) evalIntegerBinaryOp(op string, left, right Value) Value {
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
			// Task 9.111: Enhanced error with operand values
			return i.NewRuntimeError(
				i.currentNode,
				"division_by_zero",
				fmt.Sprintf("Division by zero: %d / %d", leftVal, rightVal),
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
			// Task 9.111: Enhanced error with operand values
			return i.NewRuntimeError(
				i.currentNode,
				"division_by_zero",
				fmt.Sprintf("Division by zero: %d div %d", leftVal, rightVal),
				map[string]string{
					"left":  fmt.Sprintf("%d", leftVal),
					"right": fmt.Sprintf("%d", rightVal),
				},
			)
		}
		return &IntegerValue{Value: leftVal / rightVal}
	case "mod":
		if rightVal == 0 {
			// Task 9.111: Enhanced error with operand values
			return i.NewRuntimeError(
				i.currentNode,
				"modulo_by_zero",
				fmt.Sprintf("Modulo by zero: %d mod %d", leftVal, rightVal),
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
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalFloatBinaryOp evaluates binary operations on floats.
// Handles mixed integer/float operations by converting to float.
func (i *Interpreter) evalFloatBinaryOp(op string, left, right Value) Value {
	var leftVal, rightVal float64

	// Convert left operand to float
	switch v := left.(type) {
	case *FloatValue:
		leftVal = v.Value
	case *IntegerValue:
		leftVal = float64(v.Value)
	default:
		return newError("type error in float operation")
	}

	// Convert right operand to float
	switch v := right.(type) {
	case *FloatValue:
		rightVal = v.Value
	case *IntegerValue:
		rightVal = float64(v.Value)
	default:
		return newError("type error in float operation")
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
			// Task 9.111: Enhanced error with operand values
			return i.NewRuntimeError(
				i.currentNode,
				"division_by_zero",
				fmt.Sprintf("Division by zero: %v / %v", leftVal, rightVal),
				map[string]string{
					"left":  fmt.Sprintf("%v", leftVal),
					"right": fmt.Sprintf("%v", rightVal),
				},
			)
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
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalStringBinaryOp evaluates binary operations on strings.
func (i *Interpreter) evalStringBinaryOp(op string, left, right Value) Value {
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
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalBooleanBinaryOp evaluates binary operations on booleans.
func (i *Interpreter) evalBooleanBinaryOp(op string, left, right Value) Value {
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
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalVariantBinaryOp evaluates binary operations on Variant values.
// Task 9.229: Implement Variant arithmetic operators (+, -, *, /, div, mod)
// Task 9.230: Implement Variant comparison operators (=, <>, <, >, <=, >=)
//
// Variant operations follow these rules:
//   - Unwrap operands to get actual runtime values
//   - Apply numeric promotion (Integer + Float → Float)
//   - Support string concatenation with + operator
//   - Raise runtime error if types are incompatible
func (i *Interpreter) evalVariantBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Unwrap Variant values to get the actual runtime values
	leftVal := unwrapVariant(left)
	rightVal := unwrapVariant(right)

	// Handle nil/unassigned Variants
	_, leftIsNil := leftVal.(*NilValue)
	_, rightIsNil := rightVal.(*NilValue)

	// For comparison operators with nil, handle specially
	if (op == "=" || op == "<>") && (leftIsNil || rightIsNil) {
		if leftIsNil && rightIsNil {
			return &BooleanValue{Value: op == "="}
		}
		return &BooleanValue{Value: op == "<>"}
	}

	// Error if either operand is nil for non-comparison operators
	if leftIsNil {
		return i.newErrorWithLocation(node, "cannot perform operation on unassigned Variant")
	}
	if rightIsNil {
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

// evalUnaryExpression evaluates a unary expression.
func (i *Interpreter) evalUnaryExpression(expr *ast.UnaryExpression) Value {
	right := i.Eval(expr.Right)
	if isError(right) {
		return right
	}

	if result, ok := i.tryUnaryOperator(expr.Operator, right, expr); ok {
		return result
	}

	switch expr.Operator {
	case "-":
		return i.evalMinusUnaryOp(right)
	case "+":
		return i.evalPlusUnaryOp(right)
	case "not":
		return i.evalNotUnaryOp(right)
	default:
		return newError("unknown operator: %s%s", expr.Operator, right.Type())
	}
}

func (i *Interpreter) tryUnaryOperator(operator string, operand Value, node ast.Node) (Value, bool) {
	operands := []Value{operand}
	operandTypes := []string{valueTypeKey(operand)}

	if obj, ok := operand.(*ObjectInstance); ok {
		if entry, found := obj.Class.lookupOperator(operator, operandTypes); found {
			return i.invokeRuntimeOperator(entry, operands, node), true
		}
	}

	if entry, found := i.globalOperators.lookup(operator, operandTypes); found {
		return i.invokeRuntimeOperator(entry, operands, node), true
	}

	return nil, false
}

// evalMinusUnaryOp evaluates the unary minus operator.
func (i *Interpreter) evalMinusUnaryOp(right Value) Value {
	switch v := right.(type) {
	case *IntegerValue:
		return &IntegerValue{Value: -v.Value}
	case *FloatValue:
		return &FloatValue{Value: -v.Value}
	default:
		return i.newErrorWithLocation(i.currentNode, "expected integer or float for unary minus, got %s", right.Type())
	}
}

// evalPlusUnaryOp evaluates the unary plus operator.
func (i *Interpreter) evalPlusUnaryOp(right Value) Value {
	switch right.(type) {
	case *IntegerValue, *FloatValue:
		return right
	default:
		return i.newErrorWithLocation(i.currentNode, "expected integer or float for unary plus, got %s", right.Type())
	}
}

// evalNotUnaryOp evaluates the not operator.
func (i *Interpreter) evalNotUnaryOp(right Value) Value {
	// Handle boolean NOT
	if boolVal, ok := right.(*BooleanValue); ok {
		return &BooleanValue{Value: !boolVal.Value}
	}

	// Handle bitwise NOT for integers
	if intVal, ok := right.(*IntegerValue); ok {
		return &IntegerValue{Value: ^intVal.Value}
	}

	return i.newErrorWithLocation(i.currentNode, "NOT operator requires Boolean or Integer operand, got %s", right.Type())
}

// evalAddressOfExpression evaluates an address-of expression (@Function).
// Implement address-of operator evaluation to create function pointers.
//
// This creates a FunctionPointerValue that wraps the target function/procedure.
// For methods, it also captures the Self object to create a method pointer.
func (i *Interpreter) evalAddressOfExpression(expr *ast.AddressOfExpression) Value {
	// The operator should be an identifier (function/procedure name) or member access (for methods)
	switch operand := expr.Operator.(type) {
	case *ast.Identifier:
		// Regular function/procedure pointer: @FunctionName
		return i.evalFunctionPointer(operand.Value, nil, expr)

	case *ast.MemberAccessExpression:
		// Method pointer: @object.MethodName
		// First evaluate the object
		objectVal := i.Eval(operand.Object)
		if isError(objectVal) {
			return objectVal
		}

		// Get the method name
		methodName := operand.Member.Value

		// Create method pointer with the object as Self
		return i.evalFunctionPointer(methodName, objectVal, expr)

	default:
		return newError("address-of operator requires function or method name, got %T", operand)
	}
}

// evalFunctionPointer creates a function pointer value for the named function.
// If selfObject is non-nil, creates a method pointer.
func (i *Interpreter) evalFunctionPointer(name string, selfObject Value, _ ast.Node) Value {
	// Look up the function in the function registry
	function, exists := i.functions[name]
	if !exists {
		return newError("undefined function or procedure: %s", name)
	}

	// Get the function pointer type from the semantic analyzer's type information
	// For now, create a basic function pointer type from the function signature
	var pointerType *types.FunctionPointerType

	// Build parameter types
	paramTypes := make([]types.Type, len(function.Parameters))
	for idx, param := range function.Parameters {
		if param.Type != nil {
			paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
		} else {
			paramTypes[idx] = &types.IntegerType{} // Default fallback
		}
	}

	// Get return type
	var returnType types.Type
	if function.ReturnType != nil {
		returnType = i.getTypeFromAnnotation(function.ReturnType)
	}

	// Create the function pointer type
	// If this is a method pointer, create a MethodPointerType
	if selfObject != nil {
		methodPtr := types.NewMethodPointerType(paramTypes, returnType)
		// Cast to FunctionPointerType for storage
		pointerType = &methodPtr.FunctionPointerType
	} else if returnType != nil {
		pointerType = types.NewFunctionPointerType(paramTypes, returnType)
	} else {
		pointerType = types.NewProcedurePointerType(paramTypes)
	}

	// Create and return the function pointer value
	return NewFunctionPointerValue(function, i.env, selfObject, pointerType)
}

// getTypeFromAnnotation converts a type annotation to a types.Type
// This is a helper to extract type information from AST
func (i *Interpreter) getTypeFromAnnotation(typeAnnotation *ast.TypeAnnotation) types.Type {
	if typeAnnotation == nil {
		return nil
	}

	// TypeAnnotation has a Name field that contains the type name
	typeName := typeAnnotation.Name
	return i.getTypeByName(typeName)
}

// getTypeByName looks up a type by name
func (i *Interpreter) getTypeByName(name string) types.Type {
	switch name {
	case "Integer":
		return &types.IntegerType{}
	case "Float":
		return &types.FloatType{}
	case "String":
		return &types.StringType{}
	case "Boolean":
		return &types.BooleanType{}
	default:
		// Try to find in type registry (for custom types)
		// For now, return integer as placeholder
		return &types.IntegerType{}
	}
}

// evalLambdaExpression evaluates a lambda expression and creates a closure.
// Task Lambda evaluation - creates a closure capturing the current environment.
//
// A lambda expression evaluates to a function pointer value that captures the
// environment where it was created (closure). The closure allows the lambda to
// access variables from outer scopes when it's eventually called.
//
// Examples:
//   - var double := lambda(x: Integer): Integer begin Result := x * 2; end;
//   - var add := lambda(a, b: Integer) => a + b;  // shorthand syntax
//   - Capturing outer variable: var factor := 10;
//     var multiply := lambda(x: Integer) => x * factor;
func (i *Interpreter) evalLambdaExpression(expr *ast.LambdaExpression) Value {
	// The current environment becomes the closure environment
	// This captures all variables accessible at the point where the lambda is defined
	closureEnv := i.env

	// Get the function pointer type from the semantic analyzer
	// The semantic analyzer already computed the type during type checking
	var pointerType *types.FunctionPointerType
	if expr.Type != nil {
		// Extract the type information from the annotation
		// The semantic analyzer stored a FunctionPointerType in expr.Type
		pointerType = i.getFunctionPointerTypeFromAnnotation(expr.Type)
	} else {
		// Fallback: construct type from lambda signature
		// Build parameter types
		paramTypes := make([]types.Type, len(expr.Parameters))
		for idx, param := range expr.Parameters {
			if param.Type != nil {
				paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
			} else {
				paramTypes[idx] = &types.IntegerType{} // Default fallback
			}
		}

		// Get return type
		var returnType types.Type
		if expr.ReturnType != nil {
			returnType = i.getTypeFromAnnotation(expr.ReturnType)
		}

		// Create the function pointer type
		if returnType != nil {
			pointerType = types.NewFunctionPointerType(paramTypes, returnType)
		} else {
			pointerType = types.NewProcedurePointerType(paramTypes)
		}
	}

	// Create and return a lambda value (closure)
	// The lambda captures the current environment (closureEnv) which includes
	// all variables from outer scopes listed in expr.CapturedVars
	return NewLambdaValue(expr, closureEnv, pointerType)
}

// getFunctionPointerTypeFromAnnotation extracts FunctionPointerType from a type annotation.
// Helper for lambda evaluation to get the type computed by semantic analysis.
func (i *Interpreter) getFunctionPointerTypeFromAnnotation(typeAnnotation *ast.TypeAnnotation) *types.FunctionPointerType {
	if typeAnnotation == nil {
		return nil
	}

	// For lambda expressions, the semantic analyzer stores a FunctionPointerType
	// in the Type field. We need to reconstruct it from the annotation.
	// For now, we'll use the type name to determine if it's a function pointer

	// TODO: This is a simplified implementation. In a full implementation,
	// the semantic analyzer should provide a way to get the computed type directly.
	// For now, return nil to trigger the fallback in evalLambdaExpression

	return nil
}

// evalInOperator evaluates the 'in' operator for checking membership in arrays
// Syntax: value in array
// Returns: Boolean indicating whether value is found in the array
func (i *Interpreter) evalInOperator(value Value, container Value, node ast.Node) Value {
	// Check if container is an array
	arrVal, ok := container.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(node, "type mismatch: %s in %s", value.Type(), container.Type())
	}

	// Search for the value in the array
	for _, elem := range arrVal.Elements {
		// Compare values for equality
		if i.valuesEqual(value, elem) {
			return &BooleanValue{Value: true}
		}
	}

	// Value not found
	return &BooleanValue{Value: false}
}
