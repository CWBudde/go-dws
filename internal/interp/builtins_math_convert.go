package interp

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// builtinDivMod implements the DivMod() built-in procedure.
// It computes both the quotient and remainder of integer division.
// DivMod(dividend, divisor: Integer; var quotient, remainder: Integer)
// Note: This function is called from functions_builtins.go with special handling for var parameters
func (i *Interpreter) builtinDivMod(args []ast.Expression) Value {
	// Validate argument count (exactly 4 arguments)
	if len(args) != 4 {
		return i.newErrorWithLocation(i.currentNode, "DivMod() expects exactly 4 arguments, got %d", len(args))
	}

	// Evaluate first two arguments (dividend and divisor)
	dividendVal := i.Eval(args[0])
	if isError(dividendVal) {
		return dividendVal
	}
	dividendInt, ok1 := dividendVal.(*IntegerValue)
	if !ok1 {
		return i.newErrorWithLocation(i.currentNode, "DivMod() expects integer as first argument, got %s", dividendVal.Type())
	}

	divisorVal := i.Eval(args[1])
	if isError(divisorVal) {
		return divisorVal
	}
	divisorInt, ok2 := divisorVal.(*IntegerValue)
	if !ok2 {
		return i.newErrorWithLocation(i.currentNode, "DivMod() expects integer as second argument, got %s", divisorVal.Type())
	}

	// Check for division by zero
	if divisorInt.Value == 0 {
		return i.newErrorWithLocation(i.currentNode, "DivMod() division by zero")
	}

	// Calculate quotient and remainder
	quotient := dividendInt.Value / divisorInt.Value
	remainder := dividendInt.Value % divisorInt.Value

	// Last two arguments must be identifiers (variable names for var parameters)
	quotientIdent, ok3 := args[2].(*ast.Identifier)
	if !ok3 {
		return i.newErrorWithLocation(i.currentNode, "DivMod() third argument must be a variable, got %T", args[2])
	}
	remainderIdent, ok4 := args[3].(*ast.Identifier)
	if !ok4 {
		return i.newErrorWithLocation(i.currentNode, "DivMod() fourth argument must be a variable, got %T", args[3])
	}

	// Get variable names
	quotientVarName := quotientIdent.Value
	remainderVarName := remainderIdent.Value

	// Check if variables exist and handle ReferenceValue (var parameters)
	quotientVar, exists1 := i.env.Get(quotientVarName)
	if !exists1 {
		return i.newErrorWithLocation(i.currentNode, "undefined variable: %s", quotientVarName)
	}
	remainderVar, exists2 := i.env.Get(remainderVarName)
	if !exists2 {
		return i.newErrorWithLocation(i.currentNode, "undefined variable: %s", remainderVarName)
	}

	// Handle var parameters (ReferenceValue)
	quotientResult := &IntegerValue{Value: quotient}
	remainderResult := &IntegerValue{Value: remainder}

	if refQuot, isRef := quotientVar.(*ReferenceValue); isRef {
		if err := refQuot.Assign(quotientResult); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", quotientVarName, err)
		}
	} else {
		if err := i.env.Set(quotientVarName, quotientResult); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", quotientVarName, err)
		}
	}

	if refRem, isRef := remainderVar.(*ReferenceValue); isRef {
		if err := refRem.Assign(remainderResult); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", remainderVarName, err)
		}
	} else {
		if err := i.env.Set(remainderVarName, remainderResult); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", remainderVarName, err)
		}
	}

	return &NilValue{}
}
