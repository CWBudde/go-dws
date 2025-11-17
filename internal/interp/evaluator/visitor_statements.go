package evaluator

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

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
		case "NIL":
			return false
		case "VARIANT":
			// Nested variant - shouldn't happen in practice
			return false
		default:
			// For objects, arrays, records, etc: non-nil → true
			return true
		}
	}
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
	// Phase 3.5.2: Delegate to interpreter for now
	// This has special logic for auto-invoking function pointers
	return e.adapter.EvalNode(node)
}

// VisitVarDeclStatement evaluates a variable declaration statement.
func (e *Evaluator) VisitVarDeclStatement(node *ast.VarDeclStatement, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitConstDecl evaluates a constant declaration.
func (e *Evaluator) VisitConstDecl(node *ast.ConstDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitAssignmentStatement evaluates an assignment statement.
func (e *Evaluator) VisitAssignmentStatement(node *ast.AssignmentStatement, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
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
func (e *Evaluator) VisitForStatement(node *ast.ForStatement, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitForInStatement evaluates a for-in loop statement.
func (e *Evaluator) VisitForInStatement(node *ast.ForInStatement, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitCaseStatement evaluates a case statement (switch).
func (e *Evaluator) VisitCaseStatement(node *ast.CaseStatement, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitTryStatement evaluates a try-except-finally statement.
func (e *Evaluator) VisitTryStatement(node *ast.TryStatement, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitRaiseStatement evaluates a raise statement (exception throwing).
func (e *Evaluator) VisitRaiseStatement(node *ast.RaiseStatement, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
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
