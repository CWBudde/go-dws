package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains loop statement evaluation (while, repeat, for, for-in, break, continue, exit, return).

// executeLoopIteration evaluates the loop body and handles control flow signals.
// Returns (result, shouldStop). shouldStop is true if the loop should terminate (break, exit, error).
func (i *Interpreter) executeLoopIteration(body ast.Node) (Value, bool) {
	result := i.Eval(body)
	if isError(result) {
		return result, true
	}

	cf := i.ctx.ControlFlow()
	if cf.IsBreak() {
		cf.Clear()
		return result, true
	}
	if cf.IsContinue() {
		cf.Clear()
		return result, false
	}
	if cf.IsExit() {
		// Don't clear the signal - let the function handle it
		return result, true
	}
	return result, false
}

// evalWhileStatement evaluates a while loop.
// It repeatedly evaluates the condition and executes the body while the condition is true.
func (i *Interpreter) evalWhileStatement(stmt *ast.WhileStatement) Value {
	var result Value = &NilValue{}

	for {
		// Evaluate the condition
		condition := i.Eval(stmt.Condition)
		if isError(condition) {
			return condition
		}

		// Check if condition is true
		if !isTruthy(condition) {
			break
		}

		// Execute the body
		res, stop := i.executeLoopIteration(stmt.Body)
		result = res
		if stop {
			break
		}
	}

	return result
}

// evalRepeatStatement evaluates a repeat-until loop.
// The body executes at least once, then repeats until the condition becomes true.
// This differs from while loops: the body always executes at least once,
// and the loop continues UNTIL the condition is true (not WHILE it's true).
func (i *Interpreter) evalRepeatStatement(stmt *ast.RepeatStatement) Value {
	var result Value

	for {
		// Execute the body first (repeat-until always executes at least once)
		res, stop := i.executeLoopIteration(stmt.Body)
		result = res
		if stop {
			break
		}

		// Evaluate the condition
		condition := i.Eval(stmt.Condition)
		if isError(condition) {
			return condition
		}

		// Check if condition is true - if so, exit the loop
		// Note: repeat UNTIL condition, so we break when condition is TRUE
		if isTruthy(condition) {
			break
		}
	}

	return result
}

// evalForStatement evaluates a for loop.
// DWScript for loops iterate from start to end (or downto), with the loop variable
// scoped to the loop body. The loop variable is automatically created and managed.
func (i *Interpreter) evalForStatement(stmt *ast.ForStatement) Value {
	var result Value = &NilValue{}

	// Evaluate start value
	startVal := i.Eval(stmt.Start)
	if isError(startVal) {
		return startVal
	}

	// Evaluate end value
	endVal := i.Eval(stmt.EndValue)
	if isError(endVal) {
		return endVal
	}

	// Convert start/end to ordinal values
	startOrd, err := runtime.GetOrdinalValue(startVal)
	if err != nil {
		return newError("for loop start value must be ordinal, got %s", startVal.Type())
	}
	endOrd, err := runtime.GetOrdinalValue(endVal)
	if err != nil {
		return newError("for loop end value must be ordinal, got %s", endVal.Type())
	}

	stepValue, errVal := i.getForLoopStep(stmt.Step)
	if errVal != nil {
		return errVal
	}

	makeLoopValue, errVal := i.createLoopValueFactory(startVal)
	if errVal != nil {
		return errVal
	}

	// Create a new enclosed environment for the loop variable
	// This ensures the loop variable is scoped to the loop body
	defer i.PushScope()()

	// Define the loop variable in the loop environment
	loopVarName := stmt.Variable.Value

	// Execute the loop based on direction (ascending with step)
	if stmt.Direction == ast.ForTo {
		for current := int64(startOrd); current <= int64(endOrd); current += stepValue {
			// Set the loop variable to the current value
			i.Env().Define(loopVarName, makeLoopValue(current))

			// Execute the body
			res, stop := i.executeLoopIteration(stmt.Body)
			result = res
			if stop {
				break
			}
		}
	} else {
		// Descending loop with step support
		for current := int64(startOrd); current >= int64(endOrd); current -= stepValue {
			// Set the loop variable to the current value
			i.Env().Define(loopVarName, makeLoopValue(current))

			// Execute the body
			res, stop := i.executeLoopIteration(stmt.Body)
			result = res
			if stop {
				break
			}
		}
	}

	return result
}

func (i *Interpreter) getForLoopStep(stepExpr ast.Expression) (int64, Value) {
	// Evaluate step expression if present (default: 1)
	if stepExpr == nil {
		return 1, nil
	}

	stepVal := i.Eval(stepExpr)
	if isError(stepVal) {
		return 0, stepVal
	}

	stepOrd, err := runtime.GetOrdinalValue(stepVal)
	if err != nil {
		return 0, newError("for loop step value must be ordinal, got %s", stepVal.Type())
	}

	// Validate step value is strictly positive
	if stepOrd <= 0 {
		return 0, newError("FOR loop STEP should be strictly positive: %d", stepOrd)
	}

	return int64(stepOrd), nil
}

func (i *Interpreter) createLoopValueFactory(startVal Value) (func(int64) Value, Value) {
	switch v := startVal.(type) {
	case *EnumValue:
		// Look up enum metadata to preserve type name and optional value names via TypeSystem
		enumType := func(typeName string) *types.EnumType {
			enumMetadata := i.typeSystem.LookupEnumMetadata(typeName)
			if enumMetadata == nil {
				return nil
			}
			if etv, ok := enumMetadata.(*EnumTypeValue); ok {
				return etv.EnumType
			}
			return nil
		}(v.TypeName)

		return func(ord int64) Value {
			valueName := ""
			if enumType != nil {
				valueName = enumType.GetEnumName(int(ord))
			}
			return &EnumValue{
				TypeName:     v.TypeName,
				ValueName:    valueName,
				OrdinalValue: int(ord),
			}
		}, nil
	case *IntegerValue:
		return func(ord int64) Value { return &IntegerValue{Value: ord} }, nil
	case *BooleanValue:
		return func(ord int64) Value { return &BooleanValue{Value: ord != 0} }, nil
	case *StringValue:
		return func(ord int64) Value { return &StringValue{Value: string(rune(ord))} }, nil
	default:
		return nil, newError("for loop start value must be ordinal, got %s", startVal.Type())
	}
}

// evalForInStatement evaluates a for-in loop statement.
// It iterates over the elements of a collection (array, set, or string).
// The loop variable is assigned each element in turn, and the body is executed.
func (i *Interpreter) evalForInStatement(stmt *ast.ForInStatement) Value {
	// Evaluate the collection expression
	collectionVal := i.Eval(stmt.Collection)
	if isError(collectionVal) {
		return collectionVal
	}

	// Create a new enclosed environment for the loop variable
	// This ensures the loop variable is scoped to the loop body
	defer i.PushScope()()

	loopVarName := stmt.Variable.Value

	// Type-switch on the collection type to determine iteration strategy
	switch col := collectionVal.(type) {
	case *ArrayValue:
		return i.evalForInArray(col, loopVarName, stmt.Body)
	case *SetValue:
		return i.evalForInSet(col, loopVarName, stmt.Body)
	case *StringValue:
		return i.evalForInString(col, loopVarName, stmt.Body)
	case *TypeMetaValue:
		return i.evalForInTypeMeta(col, loopVarName, stmt.Body)
	default:
		// If we reach here, the semantic analyzer missed something
		// This is defensive programming
		return newError("for-in loop: cannot iterate over %s", collectionVal.Type())
	}
}

func (i *Interpreter) evalForInArray(col *ArrayValue, loopVarName string, body ast.Node) Value {
	var result Value = &NilValue{}
	// Iterate over array elements
	for _, element := range col.Elements {
		// Assign the current element to the loop variable
		i.Env().Define(loopVarName, element)

		// Execute the body
		res, stop := i.executeLoopIteration(body)
		result = res
		if stop {
			break
		}
	}
	return result
}

func (i *Interpreter) evalForInSet(col *SetValue, loopVarName string, body ast.Node) Value {
	// Iterate over set elements
	// Sets contain enum values; we iterate through the enum's ordered names
	// and check which ones are present in the set
	if col.SetType == nil || col.SetType.ElementType == nil {
		return newError("invalid set type for iteration")
	}

	// Iterate over different set element types
	elementType := col.SetType.ElementType
	ordinals := col.Ordinals()
	var result Value = &NilValue{}

	// For enum sets, iterate through all ordinals present
	if enumType, ok := elementType.(*types.EnumType); ok {
		for _, ordinal := range ordinals {
			// Create an enum value for this element
			enumVal := &EnumValue{
				TypeName:     enumType.Name,
				ValueName:    enumType.GetEnumName(ordinal),
				OrdinalValue: ordinal,
			}

			// Assign the enum value to the loop variable
			i.Env().Define(loopVarName, enumVal)

			// Execute the body
			res, stop := i.executeLoopIteration(body)
			result = res
			if stop {
				break
			}
		}
	} else {
		// For non-enum sets (Integer, String, Boolean), iterate over ordinal values
		for _, ordinal := range ordinals {
			var loopVal Value
			switch elementType.TypeKind() {
			case "INTEGER", "SUBRANGE":
				loopVal = &IntegerValue{Value: int64(ordinal)}
			case "STRING":
				loopVal = &StringValue{Value: string(rune(ordinal))}
			case "BOOLEAN":
				loopVal = &BooleanValue{Value: ordinal != 0}
			default:
				return newError("for-in loop: cannot iterate over set of %s", elementType.String())
			}

			i.Env().Define(loopVarName, loopVal)

			// Execute the body
			res, stop := i.executeLoopIteration(body)
			result = res
			if stop {
				break
			}
		}
	}
	return result
}

func (i *Interpreter) evalForInString(col *StringValue, loopVarName string, body ast.Node) Value {
	// Iterate over string characters
	// Each character becomes a single-character string
	// Use runes to handle UTF-8 correctly
	runes := []rune(col.Value)
	var result Value = &NilValue{}
	for idx := 0; idx < len(runes); idx++ {
		// Create a single-character string for this iteration
		charVal := &StringValue{Value: string(runes[idx])}

		// Assign the character to the loop variable
		i.Env().Define(loopVarName, charVal)

		// Execute the body
		res, stop := i.executeLoopIteration(body)
		result = res
		if stop {
			break
		}
	}
	return result
}

func (i *Interpreter) evalForInTypeMeta(col *TypeMetaValue, loopVarName string, body ast.Node) Value {
	// Iterate over enum type values (e.g., for var e in TColor do)
	enumType, ok := col.TypeInfo.(*types.EnumType)
	if !ok {
		return newError("for-in loop: can only iterate over enum types, got %s", col.TypeName)
	}

	var result Value = &NilValue{}
	// Iterate through enum ordinal range (inclusive)
	// DWScript iterates over the full range from min to max, not just declared values
	// This allows constructs like: enum (Low = 2, High = 1000) to iterate 2..1000
	for ordinal := enumType.MinOrdinal(); ordinal <= enumType.MaxOrdinal(); ordinal++ {
		valueName := enumType.GetEnumName(ordinal)
		// Create an enum value for this element
		enumVal := &EnumValue{
			TypeName:     enumType.Name,
			ValueName:    valueName,
			OrdinalValue: ordinal,
		}

		// Assign the enum value to the loop variable
		i.Env().Define(loopVarName, enumVal)

		// Execute the body
		res, stop := i.executeLoopIteration(body)
		result = res
		if stop {
			break
		}
	}
	return result
}
