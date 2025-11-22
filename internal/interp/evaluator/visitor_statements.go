package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
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
			if strings.EqualFold(typeName, "Variant") {
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
// Task 3.5.17: Migrated from Interpreter.evalTryStatement()
func (e *Evaluator) VisitTryStatement(node *ast.TryStatement, ctx *ExecutionContext) Value {
	// Task 3.5.17: Try-except-finally statement with defer semantics and exception handling
	//
	// Try-except-finally syntax:
	// - try...except: try Block; except Handler; end;
	// - try...finally: try Block; finally Block; end;
	// - try...except...finally: try Block; except Handler; finally Block; end;
	//
	// Execution order:
	// 1. Execute try block
	// 2. If exception occurs:
	//    a. Execute matching except handler (if except clause exists)
	//    b. Exception cleared if handler completes normally
	//    c. Exception propagates if no matching handler
	// 3. Finally block always executes (via defer)
	//    a. Executes even if exception occurs
	//    b. Executes even if try/except raises new exception
	//    c. Can raise its own exception (replaces original)
	//
	// Defer semantics:
	// - Finally block registered with Go's defer
	// - Ensures finally runs even on panic/error
	// - Finally executes after except handlers complete
	// - Exception state saved/restored around finally
	//
	// Except clause structure:
	// - Multiple handlers: try...except on E1 do S1; on E2 do S2; end;
	// - Bare except: try...except ElseBlock; end; (catches all)
	// - Else block: try...except on E do S; else ElseBlock; end;
	//
	// Exception handler matching:
	// - Handlers tried in order
	// - First matching handler executes
	// - Match by exception class type (including inheritance)
	// - Example: on EConvertError catches EConvertError and subclasses
	// - Bare handler (no type) catches all exceptions
	//
	// Handler variable binding:
	// - on E: Exception do S; (E is bound to exception instance)
	// - Variable accessible within handler scope
	// - New scope created for handler execution
	// - Variable is ObjectInstance of exception
	// - Access fields via E.Message, etc.
	//
	// ExceptObject:
	// - Global variable set to current exception
	// - Available in except and finally blocks
	// - Allows exception access without explicit binding
	// - Example: try...finally WriteLn(ExceptObject.Message); end;
	//
	// Handler execution:
	// - Create new scope for handler
	// - Bind exception variable (if specified)
	// - Set ExceptObject to exception instance
	// - Clear exception temporarily
	// - Execute handler statement
	// - If handler completes normally, exception is cleared
	// - If handler raises/re-raises, new exception propagates
	//
	// Exception type matching:
	// - matchesExceptionType() checks class hierarchy
	// - Exception class must match handler type or inherit from it
	// - Example: on Exception catches all exception types
	// - Example: on EDivByZero catches only division by zero
	//
	// Else block:
	// - Executes if no handler matches
	// - Only if exception still active after handler matching
	// - Clears exception before execution
	// - Can raise its own exception
	//
	// Finally block execution:
	// - Always executes via defer
	// - Exception state saved before finally
	// - Exception cleared temporarily for finally execution
	// - ExceptObject set to current exception
	// - If finally completes normally, original exception restored
	// - If finally raises exception, new exception replaces original
	//
	// Nested try statements:
	// - Inner try can have its own handlers
	// - Exceptions propagate outward if not handled
	// - handlerException saved/restored for nested handlers
	// - Each level has independent finally blocks
	//
	// Bare raise in handlers:
	// - Re-raises current exception
	// - Uses handlerException saved by evalExceptClause
	// - Only valid within exception handler
	// - Error if no active exception
	//
	// Exception propagation:
	// - If no matching handler, exception remains active
	// - Exception propagates to outer try or program level
	// - Finally blocks execute during propagation
	// - Uncaught exception terminates program
	//
	// Call stack preservation:
	// - Exception includes call stack at raise point
	// - Stack trace preserved across handler invocations
	// - Available for error reporting and debugging
	//
	// Complexity: Very High - defer semantics, exception matching, state management, nested handlers
	// Full implementation requires:
	// - defer for finally block execution
	// - Exception state save/restore
	// - ExceptObject binding
	// - Handler scope creation
	// - Exception type matching with inheritance
	// - handlerException tracking for bare raise
	// - Else block handling
	// - Exception propagation logic
	//
	// Blocking Dependencies (type migration needed):
	// - ExceptionValue (ClassInfo, Message, Instance, CallStack)
	// - ObjectInstance (Fields map, Class field)
	// - ClassInfo (Name, Parent for hierarchy traversal)
	//
	// Delegate to adapter which handles all exception handling logic

	return e.adapter.EvalNode(node)
}

// VisitRaiseStatement evaluates a raise statement (exception throwing).
// Task 3.5.17: Migrated from Interpreter.evalRaiseStatement()
func (e *Evaluator) VisitRaiseStatement(node *ast.RaiseStatement, ctx *ExecutionContext) Value {
	// Task 3.5.17: Raise statement for exception throwing
	//
	// Raise syntax:
	// - Explicit: raise new Exception('Error message');
	// - Bare: raise; (re-raises current exception in handler)
	//
	// Explicit raise:
	// - Evaluate exception expression (should be object instance)
	// - Extract exception class info
	// - Extract Message field from exception object
	// - Capture current call stack
	// - Create ExceptionValue with all metadata
	// - Set exception in context (ctx.SetException or i.exception)
	// - Return nil to begin stack unwinding
	//
	// Exception object creation:
	// - Usually via new: raise new Exception('message');
	// - Can be variable: var e := Exception.Create('msg'); raise e;
	// - Object must be instance of exception class
	// - Error if not an ObjectInstance
	//
	// Exception class validation:
	// - Must inherit from Exception base class
	// - Semantic analyzer should validate this
	// - Runtime type check for safety
	//
	// Message extraction:
	// - Read Message field from exception object
	// - All exception classes have Message field
	// - Default to empty string if not set
	// - Message is StringValue
	//
	// Call stack capture:
	// - Copy current call stack at raise point
	// - Stack trace includes function names and positions
	// - Used for error reporting and debugging
	// - Preserved across exception propagation
	//
	// Bare raise:
	// - Re-raises exception currently being handled
	// - Only valid within exception handler
	// - Uses handlerException saved by evalExceptClause
	// - Example:
	//   try
	//     DoSomething;
	//   except
	//     on E: Exception do
	//     begin
	//       WriteLn('Caught: ', E.Message);
	//       raise; // Re-raise E
	//     end;
	//   end;
	//
	// handlerException state:
	// - Saved when handler begins execution
	// - Available for bare raise
	// - Restored when handler completes
	// - Supports nested exception handlers
	//
	// Bare raise error handling:
	// - Panic if no active exception
	// - Should never happen if semantic analysis correct
	// - Example error: "bare raise with no active exception"
	//
	// Exception state management:
	// - Exception stored in ctx.Exception() or i.exception
	// - Cleared by exception handlers
	// - Checked after each statement for propagation
	// - Controls execution flow (early return/break)
	//
	// Stack unwinding:
	// - Return from current function immediately
	// - Exception propagates to caller
	// - Finally blocks execute during unwinding
	// - Continue until caught or program terminates
	//
	// Exception value fields:
	// - ClassInfo: Exception class metadata
	// - Instance: Object instance of exception
	// - Message: Error message string
	// - Position: Source position where raised (can be nil)
	// - CallStack: Stack trace at raise point
	//
	// Standard exception classes:
	// - Exception: Base class for all exceptions
	// - EConvertError: Type conversion errors
	// - ERangeError: Array/string bounds errors
	// - EDivByZero: Division by zero
	// - EAssertionFailed: Assertion failures
	// - EInvalidOp: Invalid operations
	// - EScriptStackOverflow: Recursion limit exceeded
	// - EHost: Wrapper for Go runtime errors
	//
	// Custom exception classes:
	// - User can define own exception types
	// - Must inherit from Exception or subclass
	// - Can add custom fields
	// - Caught by type matching in handlers
	//
	// Complexity: Medium-High - object validation, message extraction, stack capture, state management
	// Full implementation requires:
	// - Expression evaluation for exception object
	// - ObjectInstance type validation
	// - Message field extraction
	// - Call stack capture and copy
	// - ExceptionValue creation
	// - Exception state management (ctx or interp field)
	// - handlerException access for bare raise
	// - Error handling for invalid bare raise
	//
	// Blocking Dependencies (type migration needed):
	// - ExceptionValue (for creating exceptions)
	// - ObjectInstance (for exception objects)
	// - ClassInfo (for exception class metadata)
	//
	// Delegate to adapter which handles all raise logic

	return e.adapter.EvalNode(node)
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
	switch strings.ToLower(typeName) {
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
// Phase 3.5.4 - Phase 2E: Exception Infrastructure (ready for future use)
// ============================================================================

// TODO: Implement evalExceptClause() and matchesExceptionType() when Phase 3.5.4.45 (TryStatement migration) completes
// Reference implementation available in internal/interp/exceptions.go (lines 215-315)
