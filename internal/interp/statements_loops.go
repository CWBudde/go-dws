package interp

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
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

		// Handle break/continue signals
		if i.breakSignal {
			i.breakSignal = false // Clear signal
			break
		}
		if i.continueSignal {
			i.continueSignal = false // Clear signal
			continue
		}
		// Handle exit signal (exit from function while in loop)
		if i.exitSignal {
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
	var result Value = &NilValue{}

	for {
		// Execute the body first (repeat-until always executes at least once)
		result = i.Eval(stmt.Body)
		if isError(result) {
			return result
		}

		// Handle break/continue signals
		if i.breakSignal {
			i.breakSignal = false // Clear signal
			break
		}
		if i.continueSignal {
			i.continueSignal = false // Clear signal
			// Continue to condition check
		}
		// Handle exit signal (exit from function while in loop)
		if i.exitSignal {
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

	// Both start and end must be integers for for loops
	startInt, ok := startVal.(*IntegerValue)
	if !ok {
		return newError("for loop start value must be integer, got %s", startVal.Type())
	}

	endInt, ok := endVal.(*IntegerValue)
	if !ok {
		return newError("for loop end value must be integer, got %s", endVal.Type())
	}

	// Task 9.154: Evaluate step expression if present
	stepValue := int64(1) // Default step value
	if stmt.Step != nil {
		stepVal := i.Eval(stmt.Step)
		if isError(stepVal) {
			return stepVal
		}

		stepInt, ok := stepVal.(*IntegerValue)
		if !ok {
			return newError("for loop step value must be integer, got %s", stepVal.Type())
		}

		// Validate step value is strictly positive
		if stepInt.Value <= 0 {
			return newError("FOR loop STEP should be strictly positive: %d", stepInt.Value)
		}

		stepValue = stepInt.Value
	}

	// Create a new enclosed environment for the loop variable
	// This ensures the loop variable is scoped to the loop body
	loopEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = loopEnv

	// Define the loop variable in the loop environment
	loopVarName := stmt.Variable.Value

	// Execute the loop based on direction
	if stmt.Direction == ast.ForTo {
		// Task 9.155: Ascending loop with step support
		for current := startInt.Value; current <= endInt.Value; current += stepValue {
			// Set the loop variable to the current value
			i.env.Define(loopVarName, &IntegerValue{Value: current})

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				i.env = savedEnv // Restore environment before returning
				return result
			}

			// Handle break/continue signals
			if i.breakSignal {
				i.breakSignal = false // Clear signal
				break
			}
			if i.continueSignal {
				i.continueSignal = false // Clear signal
				continue
			}
			// Handle exit signal (exit from function while in loop)
			if i.exitSignal {
				// Don't clear the signal - let the function handle it
				break
			}
		}
	} else {
		// Task 9.155: Descending loop with step support
		for current := startInt.Value; current >= endInt.Value; current -= stepValue {
			// Set the loop variable to the current value
			i.env.Define(loopVarName, &IntegerValue{Value: current})

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				i.env = savedEnv // Restore environment before returning
				return result
			}

			// Handle break/continue signals
			if i.breakSignal {
				i.breakSignal = false // Clear signal
				break
			}
			if i.continueSignal {
				i.continueSignal = false // Clear signal
				continue
			}
			// Handle exit signal (exit from function while in loop)
			if i.exitSignal {
				// Don't clear the signal - let the function handle it
				break
			}
		}
	}

	// Restore the original environment after the loop
	i.env = savedEnv

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
	loopEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = loopEnv

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
				i.env = savedEnv // Restore environment before returning
				return result
			}

			// Handle control flow signals (break, continue, exit)
			if i.breakSignal {
				i.breakSignal = false // Clear signal
				break
			}
			if i.continueSignal {
				i.continueSignal = false // Clear signal
				continue
			}
			if i.exitSignal {
				// Don't clear the signal - let the function handle it
				break
			}
		}

	case *SetValue:
		// Iterate over set elements
		// Sets contain enum values; we iterate through the enum's ordered names
		// and check which ones are present in the set
		if col.SetType == nil || col.SetType.ElementType == nil {
			i.env = savedEnv
			return newError("invalid set type for iteration")
		}

		// Task 9.226: Handle iteration over different set element types
		elementType := col.SetType.ElementType

		// For enum sets, iterate through enum values in their defined order
		if enumType, ok := elementType.(*types.EnumType); ok {
			for _, name := range enumType.OrderedNames {
				ordinal := enumType.Values[name]
				// Check if this enum value is in the set
				if col.HasElement(ordinal) {
					// Create an enum value for this element
					enumVal := &EnumValue{
						TypeName:     enumType.Name,
						ValueName:    name,
						OrdinalValue: ordinal,
					}

					// Assign the enum value to the loop variable
					i.env.Define(loopVarName, enumVal)

					// Execute the body
					result = i.Eval(stmt.Body)
					if isError(result) {
						i.env = savedEnv // Restore environment before returning
						return result
					}

					// Handle control flow signals (break, continue, exit)
					if i.breakSignal {
						i.breakSignal = false // Clear signal
						break
					}
					if i.continueSignal {
						i.continueSignal = false // Clear signal
						continue
					}
					if i.exitSignal {
						// Don't clear the signal - let the function handle it
						break
					}
				}
			}
		} else {
			// For non-enum sets (Integer, String, Boolean), iterate over ordinal values
			// This is less common but supported for completeness
			i.env = savedEnv
			return newError("iteration over non-enum sets not yet implemented")
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
				i.env = savedEnv // Restore environment before returning
				return result
			}

			// Handle control flow signals (break, continue, exit)
			if i.breakSignal {
				i.breakSignal = false // Clear signal
				break
			}
			if i.continueSignal {
				i.continueSignal = false // Clear signal
				continue
			}
			if i.exitSignal {
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
			i.env = savedEnv
			return newError("for-in loop: can only iterate over enum types, got %s", col.TypeName)
		}

		// Iterate through enum values in their defined order
		for _, name := range enumType.OrderedNames {
			ordinal := enumType.Values[name]
			// Create an enum value for this element
			enumVal := &EnumValue{
				TypeName:     enumType.Name,
				ValueName:    name,
				OrdinalValue: ordinal,
			}

			// Assign the enum value to the loop variable
			i.env.Define(loopVarName, enumVal)

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				i.env = savedEnv // Restore environment before returning
				return result
			}

			// Handle control flow signals (break, continue, exit)
			if i.breakSignal {
				i.breakSignal = false // Clear signal
				break
			}
			if i.continueSignal {
				i.continueSignal = false // Clear signal
				continue
			}
			if i.exitSignal {
				// Don't clear the signal - let the function handle it
				break
			}
		}

	default:
		// If we reach here, the semantic analyzer missed something
		// This is defensive programming
		i.env = savedEnv
		return newError("for-in loop: cannot iterate over %s", collectionVal.Type())
	}

	// Restore the original environment after the loop
	i.env = savedEnv

	return result
}

// evalBreakStatement evaluates a break statement
// Sets the break signal to exit the innermost loop.
func (i *Interpreter) evalBreakStatement(_ *ast.BreakStatement) Value {
	i.breakSignal = true
	return &NilValue{}
}

// evalContinueStatement evaluates a continue statement
// Sets the continue signal to skip to the next iteration of the innermost loop.
func (i *Interpreter) evalContinueStatement(_ *ast.ContinueStatement) Value {
	i.continueSignal = true
	return &NilValue{}
}

// evalExitStatement evaluates an exit statement
// Sets the exit signal to exit the current function.
// If at program level, sets exit signal to terminate the program.
func (i *Interpreter) evalExitStatement(stmt *ast.ExitStatement) Value {
	i.exitSignal = true
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
	i.exitSignal = true

	return returnVal
}
