package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains loop statement evaluation (while, repeat, for, for-in, break, continue, exit, return).

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
		result = i.Eval(stmt.Body)
		if isError(result) {
			return result
		}

		// Handle control flow signals
		cf := i.ctx.ControlFlow()
		if cf.IsBreak() {
			cf.Clear()
			break
		}
		if cf.IsContinue() {
			cf.Clear()
			continue
		}
		// Handle exit signal (exit from function while in loop)
		if cf.IsExit() {
			// Don't clear the signal - let the function handle it
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
		result = i.Eval(stmt.Body)
		if isError(result) {
			return result
		}

		// Handle control flow signals
		cf := i.ctx.ControlFlow()
		if cf.IsBreak() {
			cf.Clear()
			break
		}
		if cf.IsContinue() {
			cf.Clear()
			// Continue to condition check
		}
		// Handle exit signal (exit from function while in loop)
		if cf.IsExit() {
			// Don't clear the signal - let the function handle it
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
	startOrd, err := evaluator.GetOrdinalValue(startVal)
	if err != nil {
		return newError("for loop start value must be ordinal, got %s", startVal.Type())
	}
	endOrd, err := evaluator.GetOrdinalValue(endVal)
	if err != nil {
		return newError("for loop end value must be ordinal, got %s", endVal.Type())
	}

	// Task 9.154: Evaluate step expression if present
	stepValue := int64(1) // Default step value
	if stmt.Step != nil {
		stepVal := i.Eval(stmt.Step)
		if isError(stepVal) {
			return stepVal
		}

		stepOrd, err := evaluator.GetOrdinalValue(stepVal)
		if err != nil {
			return newError("for loop step value must be ordinal, got %s", stepVal.Type())
		}

		// Validate step value is strictly positive
		if stepOrd <= 0 {
			return newError("FOR loop STEP should be strictly positive: %d", stepOrd)
		}

		stepValue = int64(stepOrd)
	}

	// Create a new enclosed environment for the loop variable
	// This ensures the loop variable is scoped to the loop body
	// Phase 3.1.4: unified scope management
	defer i.PushScope()()

	// Define the loop variable in the loop environment
	loopVarName := stmt.Variable.Value

	// Helper to rebuild loop variable values with the correct runtime type
	var makeLoopValue func(int64) Value
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

		makeLoopValue = func(ord int64) Value {
			valueName := ""
			if enumType != nil {
				valueName = enumType.GetEnumName(int(ord))
			}
			return &EnumValue{
				TypeName:     v.TypeName,
				ValueName:    valueName,
				OrdinalValue: int(ord),
			}
		}
	case *IntegerValue:
		makeLoopValue = func(ord int64) Value { return &IntegerValue{Value: ord} }
	case *BooleanValue:
		makeLoopValue = func(ord int64) Value { return &BooleanValue{Value: ord != 0} }
	case *StringValue:
		makeLoopValue = func(ord int64) Value { return &StringValue{Value: string(rune(ord))} }
	default:
		return newError("for loop start value must be ordinal, got %s", startVal.Type())
	}

	// Execute the loop based on direction
	if stmt.Direction == ast.ForTo {
		// Task 9.155: Ascending loop with step support
		for current := int64(startOrd); current <= int64(endOrd); current += stepValue {
			// Set the loop variable to the current value
			i.env.Define(loopVarName, makeLoopValue(current))

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				return result
			}

			// Handle control flow signals
			cf := i.ctx.ControlFlow()
			if cf.IsBreak() {
				cf.Clear()
				break
			}
			if cf.IsContinue() {
				cf.Clear()
				continue
			}
			// Handle exit signal (exit from function while in loop)
			if cf.IsExit() {
				// Don't clear the signal - let the function handle it
				break
			}
		}
	} else {
		// Task 9.155: Descending loop with step support
		for current := int64(startOrd); current >= int64(endOrd); current -= stepValue {
			// Set the loop variable to the current value
			i.env.Define(loopVarName, makeLoopValue(current))

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				return result
			}

			// Handle control flow signals
			cf := i.ctx.ControlFlow()
			if cf.IsBreak() {
				cf.Clear()
				break
			}
			if cf.IsContinue() {
				cf.Clear()
				continue
			}
			// Handle exit signal (exit from function while in loop)
			if cf.IsExit() {
				// Don't clear the signal - let the function handle it
				break
			}
		}
	}

	return result
}

// evalForInStatement evaluates a for-in loop statement.
// It iterates over the elements of a collection (array, set, or string).
// The loop variable is assigned each element in turn, and the body is executed.
func (i *Interpreter) evalForInStatement(stmt *ast.ForInStatement) Value {
	var result Value = &NilValue{}

	// Evaluate the collection expression
	collectionVal := i.Eval(stmt.Collection)
	if isError(collectionVal) {
		return collectionVal
	}

	// Create a new enclosed environment for the loop variable
	// This ensures the loop variable is scoped to the loop body
	// Phase 3.1.4: unified scope management
	defer i.PushScope()()

	loopVarName := stmt.Variable.Value

	// Type-switch on the collection type to determine iteration strategy
	switch col := collectionVal.(type) {
	case *ArrayValue:
		// Iterate over array elements
		for _, element := range col.Elements {
			// Assign the current element to the loop variable
			i.env.Define(loopVarName, element)

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				return result
			}

			// Handle control flow signals (break, continue, exit)
			cf := i.ctx.ControlFlow()
			if cf.IsBreak() {
				cf.Clear()
				break
			}
			if cf.IsContinue() {
				cf.Clear()
				continue
			}
			if cf.IsExit() {
				// Don't clear the signal - let the function handle it
				break
			}
		}

	case *SetValue:
		// Iterate over set elements
		// Sets contain enum values; we iterate through the enum's ordered names
		// and check which ones are present in the set
		if col.SetType == nil || col.SetType.ElementType == nil {
			return newError("invalid set type for iteration")
		}

		// Task 9.226: Handle iteration over different set element types
		elementType := col.SetType.ElementType
		ordinals := col.Ordinals()

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
				i.env.Define(loopVarName, enumVal)

				// Execute the body
				result = i.Eval(stmt.Body)
				if isError(result) {
					return result
				}

				// Handle control flow signals (break, continue, exit)
				cf := i.ctx.ControlFlow()
				if cf.IsBreak() {
					cf.Clear()
					break
				}
				if cf.IsContinue() {
					cf.Clear()
					continue
				}
				if cf.IsExit() {
					// Don't clear the signal - let the function handle it
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

				i.env.Define(loopVarName, loopVal)

				// Execute the body
				result = i.Eval(stmt.Body)
				if isError(result) {
					return result
				}

				// Handle control flow signals (break, continue, exit)
				cf := i.ctx.ControlFlow()
				if cf.IsBreak() {
					cf.Clear()
					break
				}
				if cf.IsContinue() {
					cf.Clear()
					continue
				}
				if cf.IsExit() {
					// Don't clear the signal - let the function handle it
					break
				}
			}
		}

	case *StringValue:
		// Iterate over string characters
		// Each character becomes a single-character string
		// Use runes to handle UTF-8 correctly
		runes := []rune(col.Value)
		for idx := 0; idx < len(runes); idx++ {
			// Create a single-character string for this iteration
			charVal := &StringValue{Value: string(runes[idx])}

			// Assign the character to the loop variable
			i.env.Define(loopVarName, charVal)

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				return result
			}

			// Handle control flow signals (break, continue, exit)
			cf := i.ctx.ControlFlow()
			if cf.IsBreak() {
				cf.Clear()
				break
			}
			if cf.IsContinue() {
				cf.Clear()
				continue
			}
			if cf.IsExit() {
				// Don't clear the signal - let the function handle it
				break
			}
		}

	case *TypeMetaValue:
		// Task 9.213: Iterate over enum type values
		// When iterating over an enum type directly (e.g., for var e in TColor do),
		// we iterate over all values of the enum type in declaration order.
		// This is similar to set iteration but without checking membership.
		enumType, ok := col.TypeInfo.(*types.EnumType)
		if !ok {
			return newError("for-in loop: can only iterate over enum types, got %s", col.TypeName)
		}

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
			i.env.Define(loopVarName, enumVal)

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				return result
			}

			// Handle control flow signals (break, continue, exit)
			cf := i.ctx.ControlFlow()
			if cf.IsBreak() {
				cf.Clear()
				break
			}
			if cf.IsContinue() {
				cf.Clear()
				continue
			}
			if cf.IsExit() {
				// Don't clear the signal - let the function handle it
				break
			}
		}

	default:
		// If we reach here, the semantic analyzer missed something
		// This is defensive programming
		return newError("for-in loop: cannot iterate over %s", collectionVal.Type())
	}

	return result
}

// evalBreakStatement evaluates a break statement
// Sets the break signal to exit the innermost loop.
func (i *Interpreter) evalBreakStatement(_ *ast.BreakStatement) Value {
	i.ctx.ControlFlow().SetBreak()
	return &NilValue{}
}

// evalContinueStatement evaluates a continue statement
// Sets the continue signal to skip to the next iteration of the innermost loop.
func (i *Interpreter) evalContinueStatement(_ *ast.ContinueStatement) Value {
	i.ctx.ControlFlow().SetContinue()
	return &NilValue{}
}

// evalExitStatement evaluates an exit statement
// Sets the exit signal to exit the current function.
// If at program level, sets exit signal to terminate the program.
func (i *Interpreter) evalExitStatement(stmt *ast.ExitStatement) Value {
	i.ctx.ControlFlow().SetExit()
	if stmt.ReturnValue != nil {
		value := i.Eval(stmt.ReturnValue)
		if isError(value) {
			return value
		}

		// Assign evaluated value to Result if it exists
		if _, exists := i.env.Get("Result"); exists {
			i.env.Set("Result", value)
		}
		return value
	}
	// No explicit return value; function will rely on Result or default
	return &NilValue{}
}

// evalReturnStatement handles return statements in lambda expressions.
// Task 9.222: Return statements are used in shorthand lambda syntax.
//
// In shorthand lambda syntax, the parser creates a return statement:
//
//	lambda(x) => x * 2
//
// becomes:
//
//	lambda(x) begin return x * 2; end
//
// The return value is assigned to the Result variable if it exists.
func (i *Interpreter) evalReturnStatement(stmt *ast.ReturnStatement) Value {
	// Evaluate the return value
	var returnVal Value
	if stmt.ReturnValue != nil {
		returnVal = i.Eval(stmt.ReturnValue)
		if isError(returnVal) {
			return returnVal
		}
		if returnVal == nil {
			return i.newErrorWithLocation(stmt, "return expression evaluated to nil")
		}
	} else {
		returnVal = &NilValue{}
	}

	// Assign to Result variable if it exists (for functions)
	// This allows the function to return the value
	if _, exists := i.env.Get("Result"); exists {
		i.env.Set("Result", returnVal)
	}

	// Set exit signal to indicate early return
	i.ctx.ControlFlow().SetExit()

	return returnVal
}
