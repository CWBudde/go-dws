package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains control flow statement evaluation (if, case, block).

// evalBlockStatement evaluates a block of statements.
func (i *Interpreter) evalBlockStatement(block *ast.BlockStatement) Value {
	if block == nil {
		return &NilValue{}
	}

	var result Value

	for _, stmt := range block.Statements {
		result = i.Eval(stmt)

		if isError(result) {
			return result
		}

		// Check if exception is active - if so, unwind the stack
		if i.exception != nil {
			return nil
		}

		// Check for control flow signals and propagate them upward
		// These signals should propagate up to the appropriate control structure
		if i.ctx.ControlFlow().IsActive() {
			return nil // Propagate signal upward by returning early
		}
	}

	return result
}

func (i *Interpreter) evalIfStatement(stmt *ast.IfStatement) Value {
	// Evaluate the condition
	condition := i.Eval(stmt.Condition)
	if isError(condition) {
		return condition
	}

	// Convert condition to boolean
	if isTruthy(condition) {
		return i.Eval(stmt.Consequence)
	} else if stmt.Alternative != nil {
		return i.Eval(stmt.Alternative)
	}

	// No alternative and condition was false - return nil
	return &NilValue{}
}

// isTruthy determines if a value is considered "true" for conditional logic.
// DWScript semantics for Variant→Boolean: empty/nil/zero → false, otherwise → true
func isTruthy(val Value) bool {
	switch v := val.(type) {
	case *BooleanValue:
		return v.Value
	case *VariantValue:
		// Variant→Boolean coercion: unassigned → false
		if v.Value == nil {
			return false
		}
		return variantToBool(v.Value)
	default:
		// In DWScript, only booleans and variants can be used in conditions
		// Non-boolean values in conditionals would be a type error
		// For now, treat non-booleans as false
		return false
	}
}

// variantToBool converts a variant's wrapped value to boolean.
// DWScript semantics: empty/nil/zero → false, otherwise → true
func variantToBool(val Value) bool {
	if val == nil {
		return false
	}

	switch v := val.(type) {
	case *BooleanValue:
		return v.Value
	case *IntegerValue:
		return v.Value != 0
	case *FloatValue:
		return v.Value != 0.0
	case *StringValue:
		return v.Value != ""
	case *NilValue:
		return false
	case *VariantValue:
		// Nested variant - recursively unwrap
		return variantToBool(v.Value)
	default:
		// For objects, arrays, records, etc: non-nil → true
		return true
	}
}

// evalCaseStatement evaluates a case statement.
// It evaluates the case expression, then checks each branch in order.
// The first branch with a matching value has its statement executed.
// If no branch matches and there's an else clause, it's executed.
func (i *Interpreter) evalCaseStatement(stmt *ast.CaseStatement) Value {
	// Evaluate the case expression
	caseValue := i.Eval(stmt.Expression)
	if isError(caseValue) {
		return caseValue
	}

	// Check each case branch in order
	for _, branch := range stmt.Cases {
		match, errVal := i.matchCaseBranch(branch, caseValue)
		if isError(errVal) {
			return errVal
		}
		if match {
			// Execute this branch's statement
			return i.Eval(branch.Statement)
		}
	}

	// No branch matched - execute else clause if present
	if stmt.Else != nil {
		return i.Eval(stmt.Else)
	}

	// No match and no else clause - return nil
	return &NilValue{}
}

func (i *Interpreter) matchCaseBranch(branch *ast.CaseBranch, caseValue Value) (bool, Value) {
	// Check each value in this branch
	for _, branchVal := range branch.Values {
		match, errVal := i.matchCaseValue(branchVal, caseValue)
		if isError(errVal) {
			return false, errVal
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func (i *Interpreter) matchCaseValue(branchVal ast.Expression, caseValue Value) (bool, Value) {
	// Check if this is a range expression
	if rangeExpr, isRange := branchVal.(*ast.RangeExpression); isRange {
		return i.matchCaseRange(rangeExpr, caseValue)
	}

	// Regular value comparison
	branchValue := i.Eval(branchVal)
	if isError(branchValue) {
		return false, branchValue
	}

	// Check if values match
	return i.valuesEqual(caseValue, branchValue), nil
}

func (i *Interpreter) matchCaseRange(rangeExpr *ast.RangeExpression, caseValue Value) (bool, Value) {
	// Evaluate start and end of range
	startValue := i.Eval(rangeExpr.Start)
	if isError(startValue) {
		return false, startValue
	}

	endValue := i.Eval(rangeExpr.RangeEnd)
	if isError(endValue) {
		return false, endValue
	}

	// Check if caseValue is within range [start, end]
	return i.isInRange(caseValue, startValue, endValue), nil
}

// valuesEqual compares two values for equality.
// This is used by case statements to match values.
func (i *Interpreter) valuesEqual(left, right Value) bool {
	left, right = i.unwrapVariants(left, right)

	// Handle nil values (uninitialized variants)
	if left == nil && right == nil {
		return true // Both uninitialized variants are equal
	}
	if left == nil || right == nil {
		return false // One is nil, the other is not
	}

	// Handle same type comparisons
	if left.Type() != right.Type() {
		return false
	}

	return i.valuesEqualTyped(left, right)
}

func (i *Interpreter) unwrapVariants(left, right Value) (Value, Value) {
	// Unwrap VariantValue if present
	if varVal, ok := left.(*VariantValue); ok {
		left = varVal.Value
	}
	if varVal, ok := right.(*VariantValue); ok {
		right = varVal.Value
	}
	return left, right
}

func (i *Interpreter) valuesEqualTyped(left, right Value) bool {
	switch l := left.(type) {
	case *IntegerValue:
		r, ok := right.(*IntegerValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *FloatValue:
		r, ok := right.(*FloatValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *StringValue:
		r, ok := right.(*StringValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *BooleanValue:
		r, ok := right.(*BooleanValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *NilValue:
		return true // nil == nil
	case *RecordValue:
		r, ok := right.(*RecordValue)
		if !ok {
			return false
		}
		return i.recordsEqual(l, r)
	default:
		// For other types, use string comparison as fallback
		return left.String() == right.String()
	}
}

// isInRange checks if value is within the range [start, end] inclusive.
// Supports Integer, Float, String (character), and Enum values.
func (i *Interpreter) isInRange(value, start, end Value) bool {
	value, start, end = unwrapVariantsForRange(value, start, end)

	// Handle nil values (uninitialized variants)
	if value == nil || start == nil || end == nil {
		return false // Cannot perform range check with uninitialized variants
	}

	// Handle different value types
	switch v := value.(type) {
	case *IntegerValue:
		return i.isInRangeInteger(v, start, end)
	case *FloatValue:
		return i.isInRangeFloat(v, start, end)
	case *StringValue:
		return i.isInRangeString(v, start, end)
	case *EnumValue:
		return i.isInRangeEnum(v, start, end)
	}

	return false
}

func unwrapVariantsForRange(value, start, end Value) (Value, Value, Value) {
	// Unwrap VariantValue if present
	if varVal, ok := value.(*VariantValue); ok {
		value = varVal.Value
	}
	if varVal, ok := start.(*VariantValue); ok {
		start = varVal.Value
	}
	if varVal, ok := end.(*VariantValue); ok {
		end = varVal.Value
	}
	return value, start, end
}

func (i *Interpreter) isInRangeInteger(v *IntegerValue, start, end Value) bool {
	startInt, startOk := start.(*IntegerValue)
	endInt, endOk := end.(*IntegerValue)
	if startOk && endOk {
		return v.Value >= startInt.Value && v.Value <= endInt.Value
	}
	return false
}

func (i *Interpreter) isInRangeFloat(v *FloatValue, start, end Value) bool {
	startFloat, startOk := start.(*FloatValue)
	endFloat, endOk := end.(*FloatValue)
	if startOk && endOk {
		return v.Value >= startFloat.Value && v.Value <= endFloat.Value
	}
	return false
}

func (i *Interpreter) isInRangeString(v *StringValue, start, end Value) bool {
	// For strings, compare character by character
	startStr, startOk := start.(*StringValue)
	endStr, endOk := end.(*StringValue)
	// Use rune-based comparison to handle UTF-8 correctly
	if startOk && endOk && runeLength(v.Value) == 1 && runeLength(startStr.Value) == 1 && runeLength(endStr.Value) == 1 {
		// Single character comparison (for 'A'..'Z' style ranges)
		charVal, _ := runeAt(v.Value, 1)
		charStart, _ := runeAt(startStr.Value, 1)
		charEnd, _ := runeAt(endStr.Value, 1)
		return charVal >= charStart && charVal <= charEnd
	}
	// Fall back to string comparison for multi-char strings
	if startOk && endOk {
		return v.Value >= startStr.Value && v.Value <= endStr.Value
	}
	return false
}

func (i *Interpreter) isInRangeEnum(v *EnumValue, start, end Value) bool {
	startEnum, startOk := start.(*EnumValue)
	endEnum, endOk := end.(*EnumValue)
	if startOk && endOk && v.TypeName == startEnum.TypeName && v.TypeName == endEnum.TypeName {
		return v.OrdinalValue >= startEnum.OrdinalValue && v.OrdinalValue <= endEnum.OrdinalValue
	}
	return false
}

// tryCallClassOperator tries to call a class operator method for the given operator symbol.
// Returns nil if no operator is defined, otherwise returns the result of the method call (or error).
func (i *Interpreter) tryCallClassOperator(objInst *ObjectInstance, opSymbol string, args []Value, stmt *ast.AssignmentStatement) Value {
	// Look up the operator in the class (check current class and parents)
	classInfo := objInst.Class
	// Need concrete ClassInfo for operator lookup
	concreteClass, ok := classInfo.(*ClassInfo)
	if !ok || concreteClass == nil {
		return nil // No class info
	}

	// Search for the operator in this class and parent classes
	for class := concreteClass; class != nil; class = class.Parent {
		if class.Operators == nil {
			continue
		}

		opEntry := i.findOperatorInClass(class, opSymbol, args)
		if opEntry == nil {
			continue
		}

		return i.executeOperatorMethod(class, opEntry, objInst, args, stmt)
	}

	return nil // No matching operator found
}

func (i *Interpreter) findOperatorInClass(class *ClassInfo, opSymbol string, args []Value) *runtimeOperatorEntry {
	// Convert arg values to type strings for lookup using valueTypeKey
	// When searching parent classes, use the parent class name for matching
	argTypes := make([]string, len(args)+1)           // +1 for the class instance itself
	argTypes[0] = NormalizeTypeAnnotation(class.Name) // Use the current class being searched
	for idx, arg := range args {
		argTypes[idx+1] = valueTypeKey(arg) // Use valueTypeKey for consistent type keys
	}

	opEntry, found := class.Operators.lookup(opSymbol, argTypes, i.typeSystem)
	if !found {
		return nil
	}
	return opEntry
}

func (i *Interpreter) executeOperatorMethod(class *ClassInfo, opEntry *runtimeOperatorEntry, objInst *ObjectInstance, args []Value, stmt *ast.AssignmentStatement) Value {
	// Found the operator - call the bound method
	var method *ast.FunctionDecl
	var exists bool

	// Method names are case-insensitive
	normalizedBindingName := strings.ToLower(opEntry.BindingName)

	if opEntry.IsClassMethod {
		method, exists = class.ClassMethods[normalizedBindingName]
	} else {
		method, exists = class.Methods[normalizedBindingName]
	}

	if !exists {
		return i.newErrorWithLocation(stmt, "operator method '%s' not found in class '%s'",
			opEntry.BindingName, class.Name)
	}

	// Call the method with Self bound
	defer i.PushScope()()

	// Bind Self to the object instance
	i.Env().Define("Self", objInst)

	i.prepareOperatorArgs(method, args)

	i.Eval(method.Body)

	// Check for errors after method execution
	if i.exception != nil {
		return &NilValue{} // Exception is active, return value doesn't matter
	}

	// Extract return value - operator methods may have a return type
	// Check if Result variable was set in the method environment
	if resultVal, exists := i.Env().Get("Result"); exists {
		return resultVal // Return the operator result
	}

	// No explicit return value - return the modified object instance
	return objInst
}

func (i *Interpreter) prepareOperatorArgs(method *ast.FunctionDecl, args []Value) {
	// Bind parameters
	for idx, param := range method.Parameters {
		if idx < len(args) {
			argValue := args[idx]

			// Convert array arguments to array of Variant if parameter is array of const
			if param.Type != nil {
				typeName := param.Type.String()
				// Resolve potential type aliases (same pattern as registerClassOperator)
				resolvedType, err := i.resolveType(typeName)
				var paramTypeName string
				if err == nil {
					// Successfully resolved - use the resolved type's string representation
					paramTypeName = strings.ToLower(resolvedType.String())
				} else {
					// Failed to resolve - use the raw type name
					paramTypeName = strings.ToLower(typeName)
				}

				if strings.HasPrefix(paramTypeName, "array of const") || strings.HasPrefix(paramTypeName, "array of variant") {
					if arrVal, ok := argValue.(*ArrayValue); ok {
						// Convert array elements to Variants
						variantElements := make([]Value, len(arrVal.Elements))
						for idx, elem := range arrVal.Elements {
							variantElements[idx] = BoxVariant(elem)
						}
						// Create new array with Variant elements
						argValue = &ArrayValue{
							Elements:  variantElements,
							ArrayType: types.ARRAY_OF_CONST, // array of Variant
						}
					}
				}
			}

			i.Env().Define(param.Name.Value, argValue)
		}
	}
}
