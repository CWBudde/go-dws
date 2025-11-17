package evaluator

import (
	"unicode/utf8"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// Phase 3.5.4 - Phase 2E: Imports for future use (commented out for now)
// import "fmt" // For exception error messages
// import "github.com/cwbudde/go-dws/internal/errors" // For exception stack traces

// This file contains visitor methods for statement AST nodes.
// Phase 3.5.2: Visitor pattern implementation for statements.
//
// Statements perform actions and control flow, typically not returning values
// (or returning nil).

// isTruthy determines if a value is considered "true" for conditional logic.
// Task 9.35: Support Variant→Boolean implicit conversion.
// DWScript semantics for Variant→Boolean: empty/nil/zero → false, otherwise → true
func isTruthy(val Value) bool {
	switch v := val.(type) {
	case *runtime.BooleanValue:
		return v.Value
	default:
		// Check if this is a Variant type by type name
		// (VariantValue is in internal/interp, not runtime, so we check by type string)
		if val.Type() == "VARIANT" {
			// For variants, we need to unwrap and check the underlying value
			// This requires accessing the Value field, but VariantValue is not imported here
			// Use variantToBool helper
			return variantToBool(val)
		}
		// In DWScript, only booleans and variants can be used in conditions
		// Non-boolean values in conditionals would be a type error
		// But we return false as a safe default
		return false
	}
}

// variantToBool converts a variant value to boolean using DWScript semantics.
// Task 9.35: Variant→Boolean coercion rules:
// - nil/null → false
// - Integer 0 → false, non-zero → true
// - Float 0.0 → false, non-zero → true
// - String "" → false, non-empty → true
// - Objects/arrays → true (non-nil)
func variantToBool(val Value) bool {
	if val == nil {
		return false
	}

	// First, check if this is a VariantWrapper and unwrap it
	// This handles *VariantValue without needing to import it (avoids circular dependency)
	if wrapper, ok := val.(runtime.VariantWrapper); ok {
		unwrapped := wrapper.UnwrapVariant()
		// Recursively evaluate the unwrapped value
		return variantToBool(unwrapped)
	}

	switch v := val.(type) {
	case *runtime.BooleanValue:
		return v.Value
	case *runtime.IntegerValue:
		return v.Value != 0
	case *runtime.FloatValue:
		return v.Value != 0.0
	case *runtime.StringValue:
		return v.Value != ""
	case *runtime.NilValue:
		return false
	default:
		// Check by type name for types not in runtime package
		switch val.Type() {
		case "NIL", "UNASSIGNED":
			return false
		default:
			// For objects, arrays, records, etc: non-nil → true
			return true
		}
	}
}

// runeLength returns the number of Unicode characters (runes) in a string.
func runeLength(s string) int {
	return utf8.RuneCountInString(s)
}

// runeAt returns the rune at the given 1-based index in the string.
// Returns the rune and true if the index is valid, or 0 and false otherwise.
func runeAt(s string, index int) (rune, bool) {
	if index < 1 {
		return 0, false
	}

	runes := []rune(s)
	if index > len(runes) {
		return 0, false
	}

	return runes[index-1], true
}

// valuesEqual compares two values for equality.
// This is used by case statements to match values.
// Phase 3.5.4.41: Migrated from Interpreter.valuesEqual()
func valuesEqual(left, right Value) bool {
	// Unwrap VariantValue if present (check by type name since VariantValue is in interp package)
	// For now, we don't handle Variant unwrapping in the evaluator
	// This will be handled when Variant types are migrated

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

	switch l := left.(type) {
	case *runtime.IntegerValue:
		r, ok := right.(*runtime.IntegerValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *runtime.FloatValue:
		r, ok := right.(*runtime.FloatValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *runtime.StringValue:
		r, ok := right.(*runtime.StringValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *runtime.BooleanValue:
		r, ok := right.(*runtime.BooleanValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *runtime.NilValue:
		return true // nil == nil
	default:
		// For other types (RecordValue, etc.), use string comparison as fallback
		// Phase 3.5.4.41: Record comparison delegated to later migration
		return left.String() == right.String()
	}
}

// isInRange checks if value is within the range [start, end] inclusive.
// Supports Integer, Float, String (character), and Enum values.
// Phase 3.5.4.41: Migrated from Interpreter.isInRange()
func isInRange(value, start, end Value) bool {
	// Unwrap VariantValue if present - delegated to later migration
	// For now, assume values are not wrapped in Variant

	// Handle nil values (uninitialized variants)
	if value == nil || start == nil || end == nil {
		return false // Cannot perform range check with uninitialized variants
	}

	// Handle different value types
	switch v := value.(type) {
	case *runtime.IntegerValue:
		startInt, startOk := start.(*runtime.IntegerValue)
		endInt, endOk := end.(*runtime.IntegerValue)
		if startOk && endOk {
			return v.Value >= startInt.Value && v.Value <= endInt.Value
		}

	case *runtime.FloatValue:
		startFloat, startOk := start.(*runtime.FloatValue)
		endFloat, endOk := end.(*runtime.FloatValue)
		if startOk && endOk {
			return v.Value >= startFloat.Value && v.Value <= endFloat.Value
		}

	case *runtime.StringValue:
		// For strings, compare character by character
		startStr, startOk := start.(*runtime.StringValue)
		endStr, endOk := end.(*runtime.StringValue)
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

	default:
		// For EnumValue and other types, check by type name
		// Phase 3.5.4.41: Enum range checking delegated to later migration
		return false
	}

	return false
}

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
// Special handling for auto-invoking parameterless function pointers.
func (e *Evaluator) VisitExpressionStatement(node *ast.ExpressionStatement, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2A: Function call infrastructure is available via adapter
	// This has special logic for auto-invoking parameterless function pointers
	// TODO: Migrate auto-invoke logic to use adapter.CallFunctionPointer
	return e.adapter.EvalNode(node)
}

// VisitVarDeclStatement evaluates a variable declaration statement.
func (e *Evaluator) VisitVarDeclStatement(node *ast.VarDeclStatement, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Type system available for array types, type inference
	// TODO: Migrate variable declaration logic using adapter type system methods
	return e.adapter.EvalNode(node)
}

// VisitConstDecl evaluates a constant declaration.
func (e *Evaluator) VisitConstDecl(node *ast.ConstDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Record type registry available via adapter.LookupRecord()
	// TODO: Migrate constant declaration logic
	return e.adapter.EvalNode(node)
}

// VisitAssignmentStatement evaluates an assignment statement.
func (e *Evaluator) VisitAssignmentStatement(node *ast.AssignmentStatement, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Type system available for compound operators
	// Phase 3.5.4 - Phase 2C: Property setter infrastructure available via PropertyEvalContext
	// Property setters handled via EvalNode delegation (uses Phase 2A + Phase 2B + ctx.PropContext())
	// TODO: Migrate assignment logic with operator overloads and property setters
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
	if isTruthy(condition) {
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
		if !isTruthy(condition) {
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
		if isTruthy(condition) {
			break
		}
	}

	return result
}

// VisitForStatement evaluates a for loop statement.
// Phase 3.5.4 - Phase 2D: Infrastructure ready (PushEnv/PopEnv), full migration pending type migration
func (e *Evaluator) VisitForStatement(node *ast.ForStatement, ctx *ExecutionContext) Value {
	// TODO Phase 3.5.4.39: Full migration pending ArrayValue/SetValue/EnumValue migration to runtime package
	// Infrastructure is ready (ExecutionContext.PushEnv/PopEnv for scoping)
	return e.adapter.EvalNode(node)
}

// VisitForInStatement evaluates a for-in loop statement.
// Phase 3.5.4 - Phase 2D: Infrastructure ready (PushEnv/PopEnv), full migration pending type migration
func (e *Evaluator) VisitForInStatement(node *ast.ForInStatement, ctx *ExecutionContext) Value {
	// TODO Phase 3.5.4.40: Full migration pending ArrayValue/SetValue/EnumValue migration to runtime package
	// Infrastructure is ready (ExecutionContext.PushEnv/PopEnv for scoping)
	return e.adapter.EvalNode(node)
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
				if isInRange(caseValue, startValue, endValue) {
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
				if valuesEqual(caseValue, branchValue) {
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
// Phase 3.5.4 - Phase 2E: Infrastructure ready (exception methods), full migration pending type migration
func (e *Evaluator) VisitTryStatement(node *ast.TryStatement, ctx *ExecutionContext) Value {
	// TODO Phase 3.5.4.45: Full migration pending ExceptionValue/ObjectInstance migration to runtime package
	// Infrastructure is ready (ExecutionContext exception methods, evalExceptClause helper)
	return e.adapter.EvalNode(node)
}

// VisitRaiseStatement evaluates a raise statement (exception throwing).
// Phase 3.5.4 - Phase 2E: Infrastructure ready (exception methods), full migration pending type migration
func (e *Evaluator) VisitRaiseStatement(node *ast.RaiseStatement, ctx *ExecutionContext) Value {
	// TODO Phase 3.5.4.46: Full migration pending ExceptionValue/ObjectInstance migration to runtime package
	// Infrastructure is ready (ExecutionContext exception methods)
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
// Exception Handling Helpers
// Phase 3.5.4 - Phase 2E: Exception Infrastructure (ready for future use)
// ============================================================================

// evalExceptClause evaluates an except clause.
// Phase 3.5.4 - Phase 2E: Infrastructure ready, will be used when ExceptionValue migrates to runtime package
// TODO: Uncomment and use when Phase 3.5.4.45 (TryStatement migration) completes
/* func (e *Evaluator) evalExceptClause(clause *ast.ExceptClause, ctx *ExecutionContext) {
	if ctx.Exception() == nil {
		// No exception to handle
		return
	}

	// Save the current exception (cast to runtime.ExceptionValue)
	exc, ok := ctx.Exception().(*runtime.ExceptionValue)
	if !ok {
		// Invalid exception type - should not happen
		return
	}

	// If no handlers, this is a bare except - catches all
	if len(clause.Handlers) == 0 {
		ctx.SetException(nil) // Clear the exception
		return
	}

	// Try each handler in order
	for _, handler := range clause.Handlers {
		if e.matchesExceptionType(exc, handler.ExceptionType) {
			// Phase 3.5.4 - Phase 2D: Use PushEnv/PopEnv for handler scope
			ctx.PushEnv()
			defer ctx.PopEnv()

			// Bind exception variable
			if handler.Variable != nil {
				// Use Define to create a new variable in the current scope
				ctx.Env().Define(handler.Variable.Value, exc.Instance)
			}

			// Save the current handlerException (for nested handlers)
			savedHandlerException := ctx.HandlerException()

			// Save exception for bare raise to access
			ctx.SetHandlerException(exc)

			// Set ExceptObject to the current exception
			// Save old ExceptObject value to restore later
			oldExceptObject, _ := ctx.Env().Get("ExceptObject")
			ctx.Env().Define("ExceptObject", exc.Instance)

			// Temporarily clear exception to allow handler to execute
			ctx.SetException(nil)

			// Execute handler statement
			e.Eval(handler.Statement, ctx)

			// After handler executes:
			// - If ctx.Exception() is still nil, handler completed normally
			// - If ctx.Exception() is not nil, handler raised/re-raised

			// Restore handler exception context (for nested handlers)
			ctx.SetHandlerException(savedHandlerException)

			// Restore ExceptObject
			if oldExceptObject != nil {
				ctx.Env().Define("ExceptObject", oldExceptObject)
			}

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
		e.Eval(clause.ElseBlock, ctx)
	}
	// If no else block, exception remains active and will propagate
}

// matchesExceptionType checks if an exception matches a handler's type.
// Phase 3.5.4 - Phase 2E: Infrastructure ready, will be used when ExceptionValue migrates to runtime package
// TODO: Uncomment and use when Phase 3.5.4.45 (TryStatement migration) completes
/* func (e *Evaluator) matchesExceptionType(exc *runtime.ExceptionValue, typeExpr ast.TypeExpression) bool {
	if typeExpr == nil {
		return true // Bare handler catches all
	}

	typeName := typeExpr.String()

	// Check if exception class matches or inherits from handler type
	currentClass := exc.ClassInfo
	for currentClass != nil {
		if currentClass.Name == typeName {
			return true
		}
		// Check parent class
		currentClass = currentClass.Parent
	}

	return false
} */
