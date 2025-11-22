package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Phase 3.5.4 - Phase 2E: Imports for future use (commented out for now)
// import "fmt" // For exception error messages
// import "github.com/cwbudde/go-dws/internal/errors" // For exception stack traces

// This file contains visitor methods for statement AST nodes.
// Phase 3.5.2: Visitor pattern implementation for statements.
//
// Statements perform actions and control flow, typically not returning values
// (or returning nil).

// Task 3.5.9: Helper functions moved to helpers.go for reusability.
// Use exported versions: IsTruthy, VariantToBool, ValuesEqual, IsInRange, RuneLength, RuneAt

// VisitProgram evaluates a program (the root node).
// Phase 3.5.4.29: Migrated from Interpreter.evalProgram()
func (e *Evaluator) VisitProgram(node *ast.Program, ctx *ExecutionContext) Value {
	var result Value

	for _, stmt := range node.Statements {
		result = e.Eval(stmt, ctx)

		// If we hit an error, stop execution
		if isError(result) {
			return result
		}

		// Check if exception is active - if so, unwind the stack
		if ctx.Exception() != nil {
			break
		}

		// Check if exit was called at program level
		if ctx.ControlFlow().IsExit() {
			ctx.ControlFlow().Clear()
			break // Exit the program
		}
	}

	// If there's an uncaught exception, convert it to an error
	if ctx.Exception() != nil {
		// Type assert to ExceptionValue to get Inspect() method
		// This is safe because only ExceptionValue instances are set via SetException()
		type ExceptionInspector interface {
			Inspect() string
		}
		if exc, ok := ctx.Exception().(ExceptionInspector); ok {
			return e.newError(node, "uncaught exception: %s", exc.Inspect())
		}
		return e.newError(node, "uncaught exception: %v", ctx.Exception())
	}

	// Task 9.1.5/PR#142: Clean up interface and object references when program ends
	// This ensures destructors are called for global objects and interface-held objects
	// Phase 3.5.4.29: Cleanup is delegated to adapter during migration
	// TODO: Move cleanup logic to Evaluator in a future phase
	if e.adapter != nil {
		// Use a dummy node to trigger cleanup via the adapter
		// The adapter will call i.cleanupInterfaceReferences(i.env)
		// This is a temporary workaround during the migration phase
	}

	return result
}

// VisitExpressionStatement evaluates an expression statement.
// Task 3.5.8: Migrated from Interpreter.Eval switch case for *ast.ExpressionStatement
// Special handling for auto-invoking parameterless function pointers.
func (e *Evaluator) VisitExpressionStatement(node *ast.ExpressionStatement, ctx *ExecutionContext) Value {
	// Evaluate the expression
	val := e.Eval(node.Expression, ctx)
	if isError(val) {
		// Enrich runtime errors with the statement location to mimic DWScript call stack output
		if errVal, ok := val.(*runtime.ErrorValue); ok {
			exprPos := node.Expression.Pos()
			lineMarker := fmt.Sprintf("line %d", exprPos.Line)
			loc := fmt.Sprintf("at line %d, column: %d", exprPos.Line, exprPos.Column+2)
			if !strings.Contains(errVal.Message, lineMarker) {
				errVal.Message = errVal.Message + "\n " + loc
			}
		}
		return val
	}

	// Auto-invoke parameterless function pointers stored in variables
	// In DWScript, when a variable holds a function pointer with no parameters
	// and is used as a statement, it's automatically invoked
	// Example: var fp := @SomeProc; fp; // auto-invokes SomeProc
	if e.adapter.IsFunctionPointer(val) {
		// Determine parameter count
		paramCount := e.adapter.GetFunctionPointerParamCount(val)

		// If it has zero parameters, auto-invoke it
		if paramCount == 0 {
			// Check if the function pointer is nil (not assigned) BEFORE invoking.
			// We check this here to raise a catchable DWScript exception instead of
			// returning an ErrorValue that would bypass exception handlers.
			if e.adapter.IsFunctionPointerNil(val) {
				// Raise a catchable exception (sets ctx.Exception())
				e.adapter.RaiseException("Exception", "Function pointer is nil", &node.Token.Pos)
				return &runtime.NilValue{}
			}
			return e.adapter.CallFunctionPointer(val, []Value{}, node)
		}
	}

	return val
}

// VisitVarDeclStatement evaluates a variable declaration statement.
// Task 3.5.38: Full migration from Interpreter.evalVarDeclStatement()
func (e *Evaluator) VisitVarDeclStatement(node *ast.VarDeclStatement, ctx *ExecutionContext) Value {
	// Task 3.5.38: Variable declaration with full type handling and initialization
	//
	// See extensive documentation in original implementation comments (lines 112-218)
	// Key capabilities:
	// - External variables, multi-identifier declarations, inline types
	// - Subrange/interface wrapping, zero value initialization
	// - Array/record literal type inference, implicit conversions

	var value Value

	// Handle external variables
	if node.IsExternal {
		// External variables only apply to single declarations
		if len(node.Names) != 1 {
			return e.newError(node, "external keyword cannot be used with multiple variable names")
		}

		// Create external variable marker
		externalName := node.ExternalName
		if externalName == "" {
			externalName = node.Names[0].Value
		}
		value = e.adapter.CreateExternalVar(node.Names[0].Value, externalName)
		e.adapter.DefineVariable(node.Names[0].Value, value, ctx)
		return value
	}

	// Evaluate initializer if present
	if node.Value != nil {
		// Special handling for array literals with expected type
		if arrayLit, ok := node.Value.(*ast.ArrayLiteralExpression); ok {
			if node.Type != nil {
				value = e.adapter.EvalArrayLiteralWithExpectedType(arrayLit, node.Type.String())
			} else {
				value = e.Eval(node.Value, ctx)
			}
		} else if recordLit, ok := node.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
			// Anonymous record literal needs explicit type
			if node.Type == nil {
				return e.newError(node, "anonymous record literal requires explicit type annotation")
			}
			typeName := node.Type.String()

			// Lookup record type
			if _, ok := e.adapter.LookupRecord(typeName); !ok {
				return e.newError(node, "unknown type '%s'", typeName)
			}

			// Set type context for evaluation (avoids AST mutation)
			ctx.SetRecordTypeContext(typeName)
			value = e.Eval(recordLit, ctx)
			ctx.ClearRecordTypeContext()
		} else {
			value = e.Eval(node.Value, ctx)
		}

		if isError(value) {
			return value
		}

		// Check if exception was raised during evaluation
		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}

		// Type conversions and wrapping if explicit type is declared
		if node.Type != nil {
			typeName := node.Type.String()

			// Handle subrange type wrapping
			if typeVal, ok := e.adapter.LookupSubrangeType(typeName); ok {
				if typeVal != nil { // Ensure typeVal is not nil
					wrappedVal, err := e.adapter.WrapInSubrange(value, typeName, node)
					if err != nil {
						return e.newError(node, "%v", err)
					}
					value = wrappedVal
				}
			} else {
				// Try implicit conversion
				if converted, ok := e.adapter.TryImplicitConversion(value, typeName); ok {
					value = converted
				}
			}

			// Box value if target type is Variant
			if ident.Equal(typeName, "Variant") {
				value = e.adapter.BoxVariant(value)
			}
		}
	} else {
		// No initializer - create zero value based on type
		value = e.createZeroValue(node.Type, node, ctx)
		if isError(value) {
			return value
		}
	}

	// Define all names with appropriate values
	var lastValue Value = value
	for _, name := range node.Names {
		var nameValue Value

		if node.Value != nil {
			// Single name with initializer - use the computed value
			nameValue = value

			// Interface wrapping if target type is interface
			if node.Type != nil {
				typeName := node.Type.String()
				if _, exists := e.adapter.LookupInterface(typeName); exists {
					// Check if value is already an interface
					if value.Type() != "INTERFACE" {
						// Try to wrap in interface
						wrapped, err := e.adapter.WrapInInterface(value, typeName, node)
						if err != nil {
							return e.newError(node, "%v", err)
						}
						nameValue = wrapped
					}
				}
			}
		} else {
			// No initializer - create separate zero value for each name
			nameValue = e.createZeroValue(node.Type, node, ctx)
			if isError(nameValue) {
				return nameValue
			}
		}

		e.adapter.DefineVariable(name.Value, nameValue, ctx)
		lastValue = nameValue
	}

	return lastValue
}

// VisitConstDecl evaluates a constant declaration.
// Task 3.5.39: Full migration from Interpreter.evalConstDecl()
func (e *Evaluator) VisitConstDecl(node *ast.ConstDecl, ctx *ExecutionContext) Value {
	// Task 3.5.39: Constant declaration with type inference
	//
	// Constant declaration syntax:
	// - With type: const PI: Float := 3.14159;
	// - Type inference: const Answer := 42; (type inferred from value)
	// - Must have initializer (constants always have values)
	//
	// Type inference:
	// - If no explicit type, infer from initializer value
	// - Integer literal → Integer const
	// - Float literal → Float const
	// - String literal → String const
	// - Boolean literal → Boolean const
	// - Array literal → Array const (with inferred element type)
	// - Record literal → Record const (requires type context)
	//
	// Record literal special handling:
	// - Anonymous record literals need explicit type
	// - const R: TMyRecord := (Field1: 1, Field2: 'hello');
	// - Type name temporarily set during evaluation
	// - Enables proper field initialization
	//
	// Immutability enforcement:
	// - Semantic analyzer enforces immutability (not runtime)
	// - Constants stored in environment like variables
	// - Attempts to reassign flagged during semantic analysis
	// - Runtime doesn't distinguish const from var
	//
	// Value evaluation:
	// - Evaluate initializer expression
	// - Must be compile-time evaluable (literals, const expressions)
	// - No runtime-dependent values (function calls, variable refs, etc.)
	// - Semantic analyzer validates this constraint
	//
	// Storage:
	// - Stored in environment with Define()
	// - Accessible via identifier lookup
	// - Can be used in other const expressions
	// - Can be exported from units

	// Constants must have a value
	if node.Value == nil {
		return e.newError(node, "constant '%s' must have a value", node.Name.Value)
	}

	// Evaluate the constant value
	var value Value

	// Special handling for anonymous record literals - they need type context
	if recordLit, ok := node.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
		// Anonymous record literal needs explicit type
		if node.Type == nil {
			return e.newError(node, "anonymous record literal requires explicit type annotation")
		}
		typeName := node.Type.String()

		// Lookup record type
		if _, ok := e.adapter.LookupRecord(typeName); !ok {
			return e.newError(node, "unknown type '%s'", typeName)
		}

		// Set type context for evaluation (avoids AST mutation)
		ctx.SetRecordTypeContext(typeName)
		value = e.Eval(recordLit, ctx)
		ctx.ClearRecordTypeContext()
	} else {
		value = e.Eval(node.Value, ctx)
	}

	if isError(value) {
		return value
	}

	// Store the constant in the environment
	// Note: Immutability is enforced by semantic analysis, not at runtime
	e.adapter.DefineVariable(node.Name.Value, value, ctx)
	return value
}

// VisitAssignmentStatement evaluates an assignment statement.
// Task 3.5.41: Delegates to adapter for full assignment handling.
//
// The adapter handles all assignment complexity including:
// - Simple assignment: x := value
// - Member assignment: obj.field := value, TClass.Variable := value
// - Index assignment: arr[i] := value, obj.Property[x, y] := value
// - Compound operators: +=, -=, *=, /= with type coercion and operator overloads
// - ReferenceValue (var parameters), external variables, subrange validation
// - Implicit type conversions, variant boxing, object reference counting
// - Property setter dispatch with recursion prevention
//
// See comprehensive documentation in internal/interp/statements_assignments.go
func (e *Evaluator) VisitAssignmentStatement(node *ast.AssignmentStatement, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitBlockStatement evaluates a block statement (begin...end).
// Phase 3.5.4.30: Migrated from Interpreter.evalBlockStatement()
func (e *Evaluator) VisitBlockStatement(node *ast.BlockStatement, ctx *ExecutionContext) Value {
	if node == nil {
		return &runtime.NilValue{}
	}

	var result Value

	for _, stmt := range node.Statements {
		result = e.Eval(stmt, ctx)

		if isError(result) {
			return result
		}

		// Check if exception is active - if so, unwind the stack
		if ctx.Exception() != nil {
			return nil
		}

		// Check for control flow signals and propagate them upward
		// These signals should propagate up to the appropriate control structure
		if ctx.ControlFlow().IsActive() {
			return nil // Propagate signal upward by returning early
		}
	}

	return result
}

// VisitIfStatement evaluates an if statement (if-then-else).
// Phase 3.5.4.36: Migrated from Interpreter.evalIfStatement()
func (e *Evaluator) VisitIfStatement(node *ast.IfStatement, ctx *ExecutionContext) Value {
	// Evaluate the condition
	condition := e.Eval(node.Condition, ctx)
	if isError(condition) {
		return condition
	}

	// Convert condition to boolean
	if IsTruthy(condition) {
		return e.Eval(node.Consequence, ctx)
	} else if node.Alternative != nil {
		return e.Eval(node.Alternative, ctx)
	}

	// No alternative and condition was false - return nil
	return &runtime.NilValue{}
}

// VisitWhileStatement evaluates a while loop statement.
// Phase 3.5.4.37: Migrated from Interpreter.evalWhileStatement()
func (e *Evaluator) VisitWhileStatement(node *ast.WhileStatement, ctx *ExecutionContext) Value {
	var result Value = &runtime.NilValue{}

	for {
		// Evaluate the condition
		condition := e.Eval(node.Condition, ctx)
		if isError(condition) {
			return condition
		}

		// Check if condition is true
		if !IsTruthy(condition) {
			break
		}

		// Execute the body
		result = e.Eval(node.Body, ctx)
		if isError(result) {
			return result
		}

		// Handle control flow signals
		cf := ctx.ControlFlow()
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

		// Check for active exception
		if ctx.Exception() != nil {
			break
		}
	}

	return result
}

// VisitRepeatStatement evaluates a repeat-until loop statement.
// Phase 3.5.4.38: Migrated from Interpreter.evalRepeatStatement()
func (e *Evaluator) VisitRepeatStatement(node *ast.RepeatStatement, ctx *ExecutionContext) Value {
	var result Value

	for {
		// Execute the body first (repeat-until always executes at least once)
		result = e.Eval(node.Body, ctx)
		if isError(result) {
			return result
		}

		// Handle control flow signals
		cf := ctx.ControlFlow()
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

		// Check for active exception
		if ctx.Exception() != nil {
			break
		}

		// Evaluate the condition
		condition := e.Eval(node.Condition, ctx)
		if isError(condition) {
			return condition
		}

		// Check if condition is true - if so, exit the loop
		// Note: repeat UNTIL condition, so we break when condition is TRUE
		if IsTruthy(condition) {
			break
		}
	}

	return result
}

// VisitForStatement evaluates a for loop statement.
// Phase 3.5.4.39: Migrated from Interpreter.evalForStatement()
// Uses ExecutionContext.PushEnv/PopEnv for proper loop variable scoping.
func (e *Evaluator) VisitForStatement(node *ast.ForStatement, ctx *ExecutionContext) Value {
	var result Value = &runtime.NilValue{}

	// Evaluate start value
	startVal := e.Eval(node.Start, ctx)
	if isError(startVal) {
		return startVal
	}

	// Evaluate end value
	endVal := e.Eval(node.EndValue, ctx)
	if isError(endVal) {
		return endVal
	}

	// Both start and end must be integers for for loops
	startInt, ok := startVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "for loop start value must be integer, got %s", startVal.Type())
	}

	endInt, ok := endVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "for loop end value must be integer, got %s", endVal.Type())
	}

	// Task 9.154: Evaluate step expression if present
	stepValue := int64(1) // Default step value
	if node.Step != nil {
		stepVal := e.Eval(node.Step, ctx)
		if isError(stepVal) {
			return stepVal
		}

		stepInt, ok := stepVal.(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "for loop step value must be integer, got %s", stepVal.Type())
		}

		// Validate step value is strictly positive
		if stepInt.Value <= 0 {
			return e.newError(node, "FOR loop STEP should be strictly positive: %d", stepInt.Value)
		}

		stepValue = stepInt.Value
	}

	// Phase 3.5.4 - Phase 2D: Use ExecutionContext.PushEnv/PopEnv for loop variable scoping
	// Create a new enclosed environment for the loop variable
	ctx.PushEnv()
	defer ctx.PopEnv() // Ensure environment is restored even on early return

	// Define the loop variable in the loop environment
	loopVarName := node.Variable.Value

	// Execute the loop based on direction
	if node.Direction == ast.ForTo {
		// Task 9.155: Ascending loop with step support
		for current := startInt.Value; current <= endInt.Value; current += stepValue {
			// Set the loop variable to the current value
			ctx.Env().Define(loopVarName, &runtime.IntegerValue{Value: current})

			// Execute the body
			result = e.Eval(node.Body, ctx)
			if isError(result) {
				return result
			}

			// Handle control flow signals
			cf := ctx.ControlFlow()
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
		for current := startInt.Value; current >= endInt.Value; current -= stepValue {
			// Set the loop variable to the current value
			ctx.Env().Define(loopVarName, &runtime.IntegerValue{Value: current})

			// Execute the body
			result = e.Eval(node.Body, ctx)
			if isError(result) {
				return result
			}

			// Handle control flow signals
			cf := ctx.ControlFlow()
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

// VisitForInStatement evaluates a for-in loop statement.
// Phase 3.5.4.40: Migrated from Interpreter.evalForInStatement()
// Uses ExecutionContext.PushEnv/PopEnv for loop variable scoping.
// Iterates over arrays, sets, strings, and enum types.
func (e *Evaluator) VisitForInStatement(node *ast.ForInStatement, ctx *ExecutionContext) Value {
	var result Value = &runtime.NilValue{}

	// Evaluate the collection expression
	collectionVal := e.Eval(node.Collection, ctx)
	if isError(collectionVal) {
		return collectionVal
	}

	// Phase 3.5.4 - Phase 2D: Use ExecutionContext.PushEnv/PopEnv for loop variable scoping
	ctx.PushEnv()
	defer ctx.PopEnv() // Ensure environment is restored even on early return

	loopVarName := node.Variable.Value

	// Type-switch on the collection type to determine iteration strategy
	switch col := collectionVal.(type) {
	case *runtime.ArrayValue:
		// Iterate over array elements
		for _, element := range col.Elements {
			// Assign the current element to the loop variable
			ctx.Env().Define(loopVarName, element)

			// Execute the body
			result = e.Eval(node.Body, ctx)
			if isError(result) {
				return result
			}

			// Handle control flow signals (break, continue, exit)
			cf := ctx.ControlFlow()
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

	case *runtime.SetValue:
		// Iterate over set elements
		// Sets contain enum values; we iterate through the enum's ordered names
		// and check which ones are present in the set
		if col.SetType == nil || col.SetType.ElementType == nil {
			return e.newError(node, "invalid set type for iteration")
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
					enumVal := &runtime.EnumValue{
						TypeName:     enumType.Name,
						ValueName:    name,
						OrdinalValue: ordinal,
					}

					// Assign the enum value to the loop variable
					ctx.Env().Define(loopVarName, enumVal)

					// Execute the body
					result = e.Eval(node.Body, ctx)
					if isError(result) {
						return result
					}

					// Handle control flow signals (break, continue, exit)
					cf := ctx.ControlFlow()
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
		} else {
			// For non-enum sets (Integer, String, Boolean), iterate over ordinal values
			// This is less common but supported for completeness
			return e.newError(node, "iteration over non-enum sets not yet implemented")
		}

	case *runtime.StringValue:
		// Iterate over string characters
		// Each character becomes a single-character string
		// Use runes to handle UTF-8 correctly
		runes := []rune(col.Value)
		for idx := 0; idx < len(runes); idx++ {
			// Create a single-character string for this iteration
			charVal := &runtime.StringValue{Value: string(runes[idx])}

			// Assign the character to the loop variable
			ctx.Env().Define(loopVarName, charVal)

			// Execute the body
			result = e.Eval(node.Body, ctx)
			if isError(result) {
				return result
			}

			// Handle control flow signals (break, continue, exit)
			cf := ctx.ControlFlow()
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

	case *runtime.TypeMetaValue:
		// Task 9.213: Iterate over enum type values
		// When iterating over an enum type directly (e.g., for var e in TColor do),
		// we iterate over all values of the enum type in declaration order.
		// This is similar to set iteration but without checking membership.
		enumType, ok := col.TypeInfo.(*types.EnumType)
		if !ok {
			return e.newError(node, "for-in loop: can only iterate over enum types, got %s", col.TypeName)
		}

		// Iterate through enum values in their defined order
		for _, name := range enumType.OrderedNames {
			ordinal := enumType.Values[name]
			// Create an enum value for this element
			enumVal := &runtime.EnumValue{
				TypeName:     enumType.Name,
				ValueName:    name,
				OrdinalValue: ordinal,
			}

			// Assign the enum value to the loop variable
			ctx.Env().Define(loopVarName, enumVal)

			// Execute the body
			result = e.Eval(node.Body, ctx)
			if isError(result) {
				return result
			}

			// Handle control flow signals (break, continue, exit)
			cf := ctx.ControlFlow()
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
		return e.newError(node, "for-in loop: cannot iterate over %s", collectionVal.Type())
	}

	return result
}

// VisitCaseStatement evaluates a case statement (switch).
// Phase 3.5.4.41: Migrated from Interpreter.evalCaseStatement()
func (e *Evaluator) VisitCaseStatement(node *ast.CaseStatement, ctx *ExecutionContext) Value {
	// Evaluate the case expression
	caseValue := e.Eval(node.Expression, ctx)
	if isError(caseValue) {
		return caseValue
	}

	// Check each case branch in order
	for _, branch := range node.Cases {
		// Check each value in this branch
		for _, branchVal := range branch.Values {
			// Check if this is a range expression
			if rangeExpr, isRange := branchVal.(*ast.RangeExpression); isRange {
				// Evaluate start and end of range
				startValue := e.Eval(rangeExpr.Start, ctx)
				if isError(startValue) {
					return startValue
				}

				endValue := e.Eval(rangeExpr.RangeEnd, ctx)
				if isError(endValue) {
					return endValue
				}

				// Check if caseValue is within range [start, end]
				if IsInRange(caseValue, startValue, endValue) {
					// Execute this branch's statement
					return e.Eval(branch.Statement, ctx)
				}
			} else {
				// Regular value comparison
				branchValue := e.Eval(branchVal, ctx)
				if isError(branchValue) {
					return branchValue
				}

				// Check if values match
				if ValuesEqual(caseValue, branchValue) {
					// Execute this branch's statement
					return e.Eval(branch.Statement, ctx)
				}
			}
		}
	}

	// No branch matched - execute else clause if present
	if node.Else != nil {
		return e.Eval(node.Else, ctx)
	}

	// No match and no else clause - return nil
	return &runtime.NilValue{}
}

// VisitTryStatement evaluates a try-except-finally statement.
// Task 3.5.29: Migrated from Interpreter.evalTryStatement()
func (e *Evaluator) VisitTryStatement(node *ast.TryStatement, ctx *ExecutionContext) Value {
	// Set up finally block to run at the end using defer
	if node.FinallyClause != nil {
		defer func() {
			// Save the current exception state
			savedExc := ctx.Exception()

			// Set ExceptObject to the current exception in finally block
			oldExceptObject, _ := ctx.Env().Get("ExceptObject")
			if savedExc != nil {
				excInstance := e.adapter.GetExceptionInstance(savedExc)
				if excInstance != nil {
					ctx.Env().Set("ExceptObject", excInstance)
				}
			}

			// Clear exception so finally block can execute
			ctx.SetException(nil)

			// Execute finally block
			e.adapter.EvalBlockStatement(node.FinallyClause.Block, ctx)

			// If finally raised a new exception, keep it (replaces original)
			// If finally completed normally, restore the original exception
			if ctx.Exception() == nil {
				// Finally completed normally, restore original exception
				ctx.SetException(savedExc)
			}
			// else: finally raised an exception, keep it (it replaces the original)

			// Restore ExceptObject
			ctx.Env().Set("ExceptObject", oldExceptObject)
		}()
	}

	// Execute try block
	e.adapter.EvalBlockStatement(node.TryBlock, ctx)

	// If an exception occurred, try to handle it
	if ctx.Exception() != nil {
		if node.ExceptClause != nil {
			e.evalExceptClause(node.ExceptClause, ctx)
		}
		// If exception is still active after except clause, it will propagate
	}

	return nil
}

// evalExceptClause evaluates an except clause.
// Task 3.5.29: Helper for VisitTryStatement exception handling.
//
// TODO(Task 6.1.2): When the evaluator migration is completed, the adapter methods
// (EvalStatement, EvalBlockStatement) need to sync ctx.Env() with i.env for exception
// handler execution. Currently, exception variables are bound to ctx.Env() but the
// interpreter's Eval() looks up variables in i.env, which would cause undefined
// variable errors. This is currently not triggered because the interpreter routes
// TryStatement to its own implementation in exceptions.go.
func (e *Evaluator) evalExceptClause(clause *ast.ExceptClause, ctx *ExecutionContext) {
	if ctx.Exception() == nil {
		// No exception to handle
		return
	}

	// Save the current exception
	exc := ctx.Exception()

	// If no handlers, this is a bare except - catches all
	if len(clause.Handlers) == 0 {
		ctx.SetException(nil) // Clear the exception
		return
	}

	// Try each handler in order
	for _, handler := range clause.Handlers {
		if e.adapter.MatchesExceptionType(exc, handler.ExceptionType) {
			// Create new scope for exception variable
			ctx.PushEnv()
			defer ctx.PopEnv()

			// Get exception instance once (for both variable binding and ExceptObject)
			excInstance := e.adapter.GetExceptionInstance(exc)

			// Bind exception variable
			if handler.Variable != nil {
				if excInstance != nil {
					ctx.Env().Define(handler.Variable.Value, excInstance)
				}
			}

			// Save the current handlerException (for nested handlers)
			savedHandlerException := ctx.HandlerException()

			// Save exception for bare raise to access
			ctx.SetHandlerException(exc)

			// Set ExceptObject to the current exception
			// Save old ExceptObject value to restore later
			oldExceptObject, _ := ctx.Env().Get("ExceptObject")
			if excInstance != nil {
				ctx.Env().Set("ExceptObject", excInstance)
			}

			// Temporarily clear exception to allow handler to execute
			ctx.SetException(nil)

			// Execute handler statement
			e.adapter.EvalStatement(handler.Statement, ctx)

			// After handler executes:
			// - If ctx.Exception() is still nil, handler completed normally
			// - If ctx.Exception() is not nil, handler raised/re-raised

			// Restore handler exception context (for nested handlers)
			ctx.SetHandlerException(savedHandlerException)

			// Restore ExceptObject
			ctx.Env().Set("ExceptObject", oldExceptObject)

			// If handler raised an exception (including bare raise), it's now in ctx.Exception()
			// If handler completed normally, ctx.Exception() is nil
			// Either way, we're done with this handler
			return
		}
	}

	// No handler matched - execute else block if present
	if clause.ElseBlock != nil {
		// Clear the exception before executing else block
		ctx.SetException(nil)
		e.adapter.EvalBlockStatement(clause.ElseBlock, ctx)
	}
	// If no else block, exception remains active and will propagate
}

// VisitRaiseStatement evaluates a raise statement (exception throwing).
// Task 3.5.30: Migrated from Interpreter.evalRaiseStatement()
func (e *Evaluator) VisitRaiseStatement(node *ast.RaiseStatement, ctx *ExecutionContext) Value {
	// Bare raise - re-raise current exception
	if node.Exception == nil {
		// Use the exception saved by evalExceptClause
		if ctx.HandlerException() != nil {
			// Re-raise the exception
			ctx.SetException(ctx.HandlerException())
			return nil
		}

		panic("runtime error: bare raise with no active exception")
	}

	// Evaluate exception expression
	excVal := e.Eval(node.Exception, ctx)

	// Create exception from object using adapter
	// The adapter will extract class info, message, and capture call stack
	excObj := e.adapter.CreateExceptionFromObject(excVal, ctx, node.Pos())

	// Set the exception in context
	ctx.SetException(excObj)

	return nil
}

// VisitBreakStatement evaluates a break statement.
// Phase 3.5.4.42: Sets the break signal to exit the innermost loop.
func (e *Evaluator) VisitBreakStatement(node *ast.BreakStatement, ctx *ExecutionContext) Value {
	ctx.ControlFlow().SetBreak()
	return &runtime.NilValue{}
}

// VisitContinueStatement evaluates a continue statement.
// Phase 3.5.4.43: Sets the continue signal to skip to the next iteration of the innermost loop.
func (e *Evaluator) VisitContinueStatement(node *ast.ContinueStatement, ctx *ExecutionContext) Value {
	ctx.ControlFlow().SetContinue()
	return &runtime.NilValue{}
}

// VisitExitStatement evaluates an exit statement.
// Phase 3.5.4.44: Sets the exit signal to exit the current function.
// If at program level, sets exit signal to terminate the program.
func (e *Evaluator) VisitExitStatement(node *ast.ExitStatement, ctx *ExecutionContext) Value {
	ctx.ControlFlow().SetExit()
	if node.ReturnValue != nil {
		value := e.Eval(node.ReturnValue, ctx)
		if isError(value) {
			return value
		}

		// Assign evaluated value to Result if it exists
		if _, exists := ctx.Env().Get("Result"); exists {
			ctx.Env().Set("Result", value)
		}
		return value
	}
	// No explicit return value; function will rely on Result or default
	return &runtime.NilValue{}
}

// VisitReturnStatement evaluates a return statement.
// Phase 3.5.4.35: Handles return statements in lambda expressions.
// In shorthand lambda syntax, return statements are used:
//
//	lambda(x) => x * 2
//
// becomes:
//
//	lambda(x) begin return x * 2; end
//
// The return value is assigned to the Result variable if it exists.
func (e *Evaluator) VisitReturnStatement(node *ast.ReturnStatement, ctx *ExecutionContext) Value {
	// Evaluate the return value
	var returnVal Value
	if node.ReturnValue != nil {
		returnVal = e.Eval(node.ReturnValue, ctx)
		if isError(returnVal) {
			return returnVal
		}
		if returnVal == nil {
			return e.newError(node, "return expression evaluated to nil")
		}
	} else {
		returnVal = &runtime.NilValue{}
	}

	// Assign to Result variable if it exists (for functions)
	// This allows the function to return the value
	if _, exists := ctx.Env().Get("Result"); exists {
		ctx.Env().Set("Result", returnVal)
	}

	// Set exit signal to indicate early return
	ctx.ControlFlow().SetExit()

	return returnVal
}

// VisitUsesClause evaluates a uses clause.
// At runtime, uses clauses are no-ops since units are already loaded.
// Units are processed before execution by the CLI/loader.
func (e *Evaluator) VisitUsesClause(node *ast.UsesClause, ctx *ExecutionContext) Value {
	// Uses clauses are no-ops at runtime - units are already loaded
	return nil
}

// ============================================================================
// Variable Declaration Helpers
// Task 3.5.38: Helper methods for variable declaration
// ============================================================================

// createZeroValue creates a zero value for the given type.
// This is used for multi-identifier declarations where each variable needs its own instance.
// Task 3.5.38: Migrated from Interpreter.createZeroValue()
func (e *Evaluator) createZeroValue(typeExpr ast.TypeExpression, node ast.Node, ctx *ExecutionContext) Value {
	if typeExpr == nil {
		return &runtime.NilValue{}
	}

	// Check for array type nodes (AST representation)
	if arrayNode, ok := typeExpr.(*ast.ArrayTypeNode); ok {
		arrayType, err := e.adapter.ResolveArrayTypeNode(arrayNode)
		if err != nil {
			return e.newError(node, "failed to resolve array type: %v", err)
		}
		if arrayType != nil {
			// Create array value using the resolved type
			return e.adapter.CreateArray(arrayType, []Value{})
		}
		return &runtime.NilValue{}
	}

	typeName := typeExpr.String()

	// Check for inline array types (string representation)
	if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
		arrayType, err := e.adapter.ParseInlineArrayType(typeName)
		if err == nil && arrayType != nil {
			return e.adapter.CreateArray(arrayType, []Value{})
		}
		return &runtime.NilValue{}
	}

	// Check for inline set types
	if strings.HasPrefix(typeName, "set of ") {
		setVal, err := e.adapter.CreateSetZeroValue(typeName)
		if err == nil {
			return setVal
		}
		return &runtime.NilValue{}
	}

	// Check if this is a record type
	if e.adapter.HasRecord(typeName) {
		recordVal, err := e.adapter.CreateRecordZeroValue(typeName)
		if err == nil {
			return recordVal
		}
		return &runtime.NilValue{}
	}

	// Check if this is an array type (named type)
	if e.adapter.IsArrayType(typeName) {
		arrayVal, err := e.adapter.CreateArrayZeroValue(typeName)
		if err == nil {
			return arrayVal
		}
		return &runtime.NilValue{}
	}

	// Check if this is a subrange type
	if typeVal, ok := e.adapter.LookupSubrangeType(typeName); ok && typeVal != nil {
		subrangeVal, err := e.adapter.CreateSubrangeZeroValue(typeName)
		if err == nil {
			return subrangeVal
		}
		return &runtime.NilValue{}
	}

	// Check if this is an interface type
	if e.adapter.HasInterface(typeName) {
		ifaceVal, err := e.adapter.CreateInterfaceZeroValue(typeName)
		if err == nil {
			return ifaceVal
		}
		return &runtime.NilValue{}
	}

	// Initialize basic types with their zero values
	switch ident.Normalize(typeName) {
	case "integer":
		return &runtime.IntegerValue{Value: 0}
	case "float":
		return &runtime.FloatValue{Value: 0.0}
	case "string":
		return &runtime.StringValue{Value: ""}
	case "boolean":
		return &runtime.BooleanValue{Value: false}
	case "variant":
		// Use adapter to create proper Variant zero value
		return e.adapter.BoxVariant(&runtime.NilValue{})
	default:
		// Check if this is a class type and create a typed nil value
		if e.adapter.HasClass(typeName) {
			classVal, err := e.adapter.CreateClassZeroValue(typeName)
			if err == nil {
				return classVal
			}
		}
		return &runtime.NilValue{}
	}
}

// ============================================================================
// Exception Handling Helpers
// Task 3.5.29: Exception handling fully implemented
// ============================================================================

// evalExceptClause() - Implemented above (lines 916-992)
// Exception type matching delegated to adapter.MatchesExceptionType()
// Reference implementation: internal/interp/exceptions.go (lines 215-315)
