package builtins

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains built-in functions that require AST-level access to modify
// variables in-place (var parameters). These functions use VarParamContext which
// provides extended capabilities beyond the basic Context interface.
//
// Functions here are registered separately from regular builtins and are routed
// through callBuiltinWithVarParam() in functions_builtins.go.

// =============================================================================
// Helper functions for common patterns
// =============================================================================

// getEnumType retrieves enum type metadata for an EnumValue.
// Returns the EnumType and nil on success, or nil and an error Value on failure.
func getEnumType(ctx VarParamContext, val *runtime.EnumValue) (*types.EnumType, Value) {
	enumMetadata := ctx.GetEnumMetadata(val.TypeName)
	if enumMetadata == nil {
		return nil, ctx.NewError("enum type metadata not found for %s", val.TypeName)
	}

	etv, ok := enumMetadata.(EnumTypeValueAccessor)
	if !ok {
		return nil, ctx.NewError("invalid enum type metadata for %s", val.TypeName)
	}
	return etv.GetEnumType(), nil
}

// findEnumPosition finds the position of an enum value in its type's OrderedNames.
// Returns the position (0-based) on success, or -1 and an error Value on failure.
func findEnumPosition(ctx VarParamContext, enumType *types.EnumType, val *runtime.EnumValue) (int, Value) {
	for idx, name := range enumType.OrderedNames {
		if name == val.ValueName {
			return idx, nil
		}
	}
	return -1, ctx.NewError("enum value '%s' not found in type '%s'", val.ValueName, val.TypeName)
}

// assignToVar assigns a value to a variable, handling ReferenceValue transparently.
func assignToVar(ctx VarParamContext, varName string, varValue, newValue Value) Value {
	if ctx.IsReference(varValue) {
		if err := ctx.AssignToReference(varValue, newValue); err != nil {
			return ctx.NewError("failed to update variable %s: %s", varName, err)
		}
	} else {
		if err := ctx.SetVariable(varName, newValue); err != nil {
			return ctx.NewError("failed to update variable %s: %s", varName, err)
		}
	}
	return nil
}

// getActualValue dereferences a value if it's a ReferenceValue.
func getActualValue(ctx VarParamContext, val Value) (Value, Value) {
	if ctx.IsReference(val) {
		deref, err := ctx.DereferenceValue(val)
		if err != nil {
			return nil, ctx.NewError("%s", err.Error())
		}
		return deref, nil
	}
	return val, nil
}

// =============================================================================
// Var-param builtin implementations
// =============================================================================

// Inc implements the Inc() built-in function.
// It increments a variable in place: Inc(x) or Inc(x, delta)
// Supports any lvalue: Inc(x), Inc(arr[i]), Inc(obj.field)
func Inc(ctx VarParamContext, args []ast.Expression) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("Inc() expects 1-2 arguments, got %d", len(args))
	}

	// Get delta (default 1)
	delta := int64(1)
	if len(args) == 2 {
		deltaVal := ctx.Eval(args[1])
		if ctx.IsError(deltaVal) {
			return deltaVal
		}
		deltaInt, ok := deltaVal.(*runtime.IntegerValue)
		if !ok {
			return ctx.NewError("Inc() delta must be Integer, got %s", deltaVal.Type())
		}
		delta = deltaInt.Value
	}

	// Evaluate lvalue once and get both current value and assignment target
	currentVal, assignFunc, err := ctx.EvaluateLValue(args[0])
	if err != nil {
		return ctx.NewError("Inc() failed to evaluate lvalue: %s", err.Error())
	}

	// Unwrap ReferenceValue if needed
	currentVal, err = ctx.DereferenceValue(currentVal)
	if err != nil {
		return ctx.NewError("%s", err.Error())
	}

	// Handle nil values (uninitialized array/record elements default to 0)
	if currentVal == nil {
		currentVal = ctx.CreateIntegerValue(0)
	}

	// Compute new value based on type
	newValue, errVal := incValue(ctx, currentVal, delta)
	if errVal != nil {
		return errVal
	}

	// Assign the new value back using the pre-evaluated assignment function
	if err := assignFunc(newValue); err != nil {
		return ctx.NewError("Inc() failed to assign: %s", err.Error())
	}

	return newValue
}

// incValue computes the incremented value for Inc().
func incValue(ctx VarParamContext, currentVal Value, delta int64) (Value, Value) {
	switch val := currentVal.(type) {
	case *runtime.IntegerValue:
		return ctx.CreateIntegerValue(val.Value + delta), nil

	case *runtime.EnumValue:
		if delta != 1 {
			return nil, ctx.NewError("Inc() with delta not supported for enum types")
		}
		return incEnumValue(ctx, val)

	default:
		return nil, ctx.NewError("Inc() expects Integer or Enum, got %s", val.Type())
	}
}

// incEnumValue computes the successor enum value.
func incEnumValue(ctx VarParamContext, val *runtime.EnumValue) (Value, Value) {
	enumType, errVal := getEnumType(ctx, val)
	if errVal != nil {
		return nil, errVal
	}

	currentPos, errVal := findEnumPosition(ctx, enumType, val)
	if errVal != nil {
		return nil, errVal
	}

	if currentPos >= len(enumType.OrderedNames)-1 {
		return nil, ctx.NewError("Inc() cannot increment enum beyond its maximum value")
	}

	nextValueName := enumType.OrderedNames[currentPos+1]
	nextOrdinal := enumType.Values[nextValueName]
	return ctx.CreateEnumValue(val.TypeName, nextValueName, int64(nextOrdinal)), nil
}

// Dec implements the Dec() built-in function.
// It decrements a variable in place: Dec(x) or Dec(x, delta)
// Supports any lvalue: Dec(x), Dec(arr[i]), Dec(obj.field)
func Dec(ctx VarParamContext, args []ast.Expression) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("Dec() expects 1-2 arguments, got %d", len(args))
	}

	// Get delta (default 1)
	delta := int64(1)
	if len(args) == 2 {
		deltaVal := ctx.Eval(args[1])
		if ctx.IsError(deltaVal) {
			return deltaVal
		}
		deltaInt, ok := deltaVal.(*runtime.IntegerValue)
		if !ok {
			return ctx.NewError("Dec() delta must be Integer, got %s", deltaVal.Type())
		}
		delta = deltaInt.Value
	}

	// Evaluate lvalue once and get both current value and assignment target
	currentVal, assignFunc, err := ctx.EvaluateLValue(args[0])
	if err != nil {
		return ctx.NewError("Dec() failed to evaluate lvalue: %s", err.Error())
	}

	// Unwrap ReferenceValue if needed
	currentVal, err = ctx.DereferenceValue(currentVal)
	if err != nil {
		return ctx.NewError("%s", err.Error())
	}

	// Handle nil values (uninitialized array/record elements default to 0)
	if currentVal == nil {
		currentVal = ctx.CreateIntegerValue(0)
	}

	// Compute new value based on type
	newValue, errVal := decValue(ctx, currentVal, delta)
	if errVal != nil {
		return errVal
	}

	// Assign the new value back using the pre-evaluated assignment function
	if err := assignFunc(newValue); err != nil {
		return ctx.NewError("Dec() failed to assign: %s", err.Error())
	}

	return newValue
}

// decValue computes the decremented value for Dec().
func decValue(ctx VarParamContext, currentVal Value, delta int64) (Value, Value) {
	switch val := currentVal.(type) {
	case *runtime.IntegerValue:
		return ctx.CreateIntegerValue(val.Value - delta), nil

	case *runtime.EnumValue:
		if delta != 1 {
			return nil, ctx.NewError("Dec() with delta not supported for enum types")
		}
		return decEnumValue(ctx, val)

	default:
		return nil, ctx.NewError("Dec() expects Integer or Enum, got %s", val.Type())
	}
}

// decEnumValue computes the predecessor enum value.
func decEnumValue(ctx VarParamContext, val *runtime.EnumValue) (Value, Value) {
	enumType, errVal := getEnumType(ctx, val)
	if errVal != nil {
		return nil, errVal
	}

	currentPos, errVal := findEnumPosition(ctx, enumType, val)
	if errVal != nil {
		return nil, errVal
	}

	if currentPos <= 0 {
		return nil, ctx.NewError("Dec() cannot decrement enum below its minimum value")
	}

	prevValueName := enumType.OrderedNames[currentPos-1]
	prevOrdinal := enumType.Values[prevValueName]
	return ctx.CreateEnumValue(val.TypeName, prevValueName, int64(prevOrdinal)), nil
}

// Insert implements the Insert() built-in function.
// It inserts a source string into a target string at the specified position.
// Insert(source, target, pos) - modifies target in-place (1-based position)
func Insert(ctx VarParamContext, args []ast.Expression) Value {
	if len(args) != 3 {
		return ctx.NewError("Insert() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: source string to insert (evaluate it)
	sourceVal := ctx.Eval(args[0])
	if ctx.IsError(sourceVal) {
		return sourceVal
	}
	sourceStr, ok := sourceVal.(*runtime.StringValue)
	if !ok {
		return ctx.NewError("Insert() expects String as first argument (source), got %s", sourceVal.Type())
	}

	// Second argument: target string variable (must be an identifier)
	targetIdent, ok := args[1].(*ast.Identifier)
	if !ok {
		return ctx.NewError("Insert() second argument (target) must be a variable, got %T", args[1])
	}

	targetName := targetIdent.Value

	// Get current target value from environment
	currentVal, exists := ctx.GetVariable(targetName)
	if !exists {
		return ctx.NewError("undefined variable: %s", targetName)
	}

	targetStr, ok := currentVal.(*runtime.StringValue)
	if !ok {
		return ctx.NewError("Insert() expects target to be String, got %s", currentVal.Type())
	}

	// Third argument: position (1-based index)
	posVal := ctx.Eval(args[2])
	if ctx.IsError(posVal) {
		return posVal
	}
	posInt, ok := posVal.(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("Insert() expects Integer as third argument (position), got %s", posVal.Type())
	}

	pos := int(posInt.Value)
	target := targetStr.Value
	source := sourceStr.Value

	// Use rune-based insertion to handle UTF-8 correctly
	newStr := ctx.RuneInsert(source, target, pos)

	// Update the target variable with the new string
	newValue := ctx.CreateStringValue(newStr)
	if err := ctx.SetVariable(targetName, newValue); err != nil {
		return ctx.NewError("failed to update variable %s: %s", targetName, err)
	}

	return ctx.CreateNilValue()
}

// DeleteString implements the Delete() built-in function for strings.
// It deletes count characters from a string starting at the specified position.
// Delete(s, pos, count) - modifies s in-place (1-based position)
func DeleteString(ctx VarParamContext, args []ast.Expression) Value {
	if len(args) != 3 {
		return ctx.NewError("Delete() for strings expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: string variable (must be an identifier)
	strIdent, ok := args[0].(*ast.Identifier)
	if !ok {
		return ctx.NewError("Delete() first argument must be a variable, got %T", args[0])
	}

	strName := strIdent.Value

	// Get current string value from environment
	currentVal, exists := ctx.GetVariable(strName)
	if !exists {
		return ctx.NewError("undefined variable: %s", strName)
	}

	strVal, ok := currentVal.(*runtime.StringValue)
	if !ok {
		return ctx.NewError("Delete() expects first argument to be String, got %s", currentVal.Type())
	}

	// Second argument: position (1-based index)
	posVal := ctx.Eval(args[1])
	if ctx.IsError(posVal) {
		return posVal
	}
	posInt, ok := posVal.(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("Delete() expects Integer as second argument (position), got %s", posVal.Type())
	}

	// Third argument: count (number of characters to delete)
	countVal := ctx.Eval(args[2])
	if ctx.IsError(countVal) {
		return countVal
	}
	countInt, ok := countVal.(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("Delete() expects Integer as third argument (count), got %s", countVal.Type())
	}

	pos := int(posInt.Value)
	count := int(countInt.Value)
	str := strVal.Value

	// Use rune-based deletion to handle UTF-8 correctly
	newStr := ctx.RuneDelete(str, pos, count)

	// Update the string variable with the new value
	newValue := ctx.CreateStringValue(newStr)
	if err := ctx.SetVariable(strName, newValue); err != nil {
		return ctx.NewError("failed to update variable %s: %s", strName, err)
	}

	return ctx.CreateNilValue()
}

// DivMod implements the DivMod() built-in procedure.
// It computes both the quotient and remainder of integer division.
// DivMod(dividend, divisor: Integer; var quotient, remainder: Integer)
func DivMod(ctx VarParamContext, args []ast.Expression) Value {
	if len(args) != 4 {
		return ctx.NewError("DivMod() expects exactly 4 arguments, got %d", len(args))
	}

	// Evaluate dividend and divisor
	dividend, divisor, errVal := evalDivModOperands(ctx, args[0], args[1])
	if errVal != nil {
		return errVal
	}

	if divisor == 0 {
		return ctx.NewError("DivMod() division by zero")
	}

	// Calculate quotient and remainder
	quotient := dividend / divisor
	remainder := dividend % divisor

	// Assign results to var parameters
	return assignDivModResults(ctx, args[2], args[3], quotient, remainder)
}

// evalDivModOperands evaluates the dividend and divisor arguments.
func evalDivModOperands(ctx VarParamContext, dividendExpr, divisorExpr ast.Expression) (int64, int64, Value) {
	dividendVal := ctx.Eval(dividendExpr)
	if ctx.IsError(dividendVal) {
		return 0, 0, dividendVal
	}
	dividendInt, ok := dividendVal.(*runtime.IntegerValue)
	if !ok {
		return 0, 0, ctx.NewError("DivMod() expects integer as first argument, got %s", dividendVal.Type())
	}

	divisorVal := ctx.Eval(divisorExpr)
	if ctx.IsError(divisorVal) {
		return 0, 0, divisorVal
	}
	divisorInt, ok := divisorVal.(*runtime.IntegerValue)
	if !ok {
		return 0, 0, ctx.NewError("DivMod() expects integer as second argument, got %s", divisorVal.Type())
	}

	return dividendInt.Value, divisorInt.Value, nil
}

// assignDivModResults assigns the quotient and remainder to the var parameters.
func assignDivModResults(ctx VarParamContext, quotientExpr, remainderExpr ast.Expression, quotient, remainder int64) Value {
	quotientIdent, ok := quotientExpr.(*ast.Identifier)
	if !ok {
		return ctx.NewError("DivMod() third argument must be a variable, got %T", quotientExpr)
	}
	remainderIdent, ok := remainderExpr.(*ast.Identifier)
	if !ok {
		return ctx.NewError("DivMod() fourth argument must be a variable, got %T", remainderExpr)
	}

	quotientVar, exists := ctx.GetVariable(quotientIdent.Value)
	if !exists {
		return ctx.NewError("undefined variable: %s", quotientIdent.Value)
	}
	remainderVar, exists := ctx.GetVariable(remainderIdent.Value)
	if !exists {
		return ctx.NewError("undefined variable: %s", remainderIdent.Value)
	}

	if errVal := assignToVar(ctx, quotientIdent.Value, quotientVar, ctx.CreateIntegerValue(quotient)); errVal != nil {
		return errVal
	}
	if errVal := assignToVar(ctx, remainderIdent.Value, remainderVar, ctx.CreateIntegerValue(remainder)); errVal != nil {
		return errVal
	}

	return ctx.CreateNilValue()
}

// Swap implements the Swap() built-in function.
// It swaps the values of two variables: Swap(var a, var b)
func Swap(ctx VarParamContext, args []ast.Expression) Value {
	if len(args) != 2 {
		return ctx.NewError("Swap() expects exactly 2 arguments, got %d", len(args))
	}

	// Both arguments must be identifiers (variable names)
	var1Ident, ok := args[0].(*ast.Identifier)
	if !ok {
		return ctx.NewError("Swap() first argument must be a variable, got %T", args[0])
	}
	var2Ident, ok := args[1].(*ast.Identifier)
	if !ok {
		return ctx.NewError("Swap() second argument must be a variable, got %T", args[1])
	}

	// Get current values from environment
	val1, exists := ctx.GetVariable(var1Ident.Value)
	if !exists {
		return ctx.NewError("undefined variable: %s", var1Ident.Value)
	}
	val2, exists := ctx.GetVariable(var2Ident.Value)
	if !exists {
		return ctx.NewError("undefined variable: %s", var2Ident.Value)
	}

	// Dereference both values
	actualVal1, errVal := getActualValue(ctx, val1)
	if errVal != nil {
		return errVal
	}
	actualVal2, errVal := getActualValue(ctx, val2)
	if errVal != nil {
		return errVal
	}

	// Swap: assign val2's content to var1 and val1's content to var2
	if errVal := assignToVar(ctx, var1Ident.Value, val1, actualVal2); errVal != nil {
		return errVal
	}
	if errVal := assignToVar(ctx, var2Ident.Value, val2, actualVal1); errVal != nil {
		return errVal
	}

	return ctx.CreateNilValue()
}

// SetLengthVarParam implements the SetLength() built-in function for var parameters.
// It resizes a dynamic array or string to the specified length.
// This version takes AST expressions (for var parameter handling) unlike SetLength in array.go.
func SetLengthVarParam(ctx VarParamContext, args []ast.Expression) Value {
	if len(args) != 2 {
		return ctx.NewError("SetLength() expects exactly 2 arguments, got %d", len(args))
	}

	// Use EvaluateLValue to support identifiers, indexed arrays, member access, etc.
	currentVal, assignFunc, err := ctx.EvaluateLValue(args[0])
	if err != nil {
		return ctx.NewError("SetLength() first argument must be a variable: %s", err.Error())
	}

	// Dereference if it's a var parameter (ReferenceValue)
	currentVal, err = ctx.DereferenceValue(currentVal)
	if err != nil {
		return ctx.NewError("%s", err.Error())
	}

	// Evaluate the second argument (new length)
	lengthVal := ctx.Eval(args[1])
	if ctx.IsError(lengthVal) {
		return lengthVal
	}

	lengthInt, ok := lengthVal.(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("SetLength() expects integer as second argument, got %s", lengthVal.Type())
	}

	newLength := int(lengthInt.Value)
	// DWScript/Delphi behavior: negative lengths are treated as 0
	if newLength < 0 {
		newLength = 0
	}

	// Handle arrays
	if arrayVal, ok := currentVal.(*runtime.ArrayValue); ok {
		// Check that it's a dynamic array
		if arrayVal.ArrayType == nil {
			return ctx.NewError("array has no type information")
		}

		if arrayVal.ArrayType.IsStatic() {
			return ctx.NewError("SetLength() can only be used with dynamic arrays, not static arrays")
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

		return ctx.CreateNilValue()
	}

	// Handle strings
	if strVal, ok := currentVal.(*runtime.StringValue); ok {
		// Use rune-based SetLength to handle UTF-8 correctly
		newStr := ctx.RuneSetLength(strVal.Value, newLength)

		// Create new StringValue
		newValue := ctx.CreateStringValue(newStr)

		// Use the assignment function to update the string
		if err := assignFunc(newValue); err != nil {
			return ctx.NewError("failed to update string variable: %s", err)
		}

		return ctx.CreateNilValue()
	}

	return ctx.NewError("SetLength() expects array or string as first argument, got %s", currentVal.Type())
}

// VarParamFunctions maps function names to their implementations.
// This is used by the interpreter to look up var-param functions.
var VarParamFunctions = map[string]VarParamBuiltinFunc{
	"Inc":       Inc,
	"Dec":       Dec,
	"Insert":    Insert,
	"Delete":    DeleteString,
	"DivMod":    DivMod,
	"Swap":      Swap,
	"SetLength": SetLengthVarParam,
}
