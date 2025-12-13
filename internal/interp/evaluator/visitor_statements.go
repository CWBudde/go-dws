package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// This file contains visitor methods for statement AST nodes.
// Statements perform actions and control flow.

// VisitProgram evaluates a program (the root node).
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

	// Convert uncaught exceptions to errors
	if ctx.Exception() != nil {
		type ExceptionInspector interface {
			Inspect() string
		}
		if exc, ok := ctx.Exception().(ExceptionInspector); ok && exc != nil {
			return e.newError(node, "uncaught exception: %s", exc.Inspect())
		}
		return e.newError(node, "uncaught exception: %v", ctx.Exception())
	}

	return result
}

// VisitEmptyStatement performs no operation for explicit empty statements (a lone semicolon).
func (e *Evaluator) VisitEmptyStatement(_ *ast.EmptyStatement, _ *ExecutionContext) Value {
	return &runtime.NilValue{}
}

// VisitExpressionStatement evaluates an expression statement.
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

	// Auto-invoke parameterless function pointers
	// Example: var fp := @SomeProc; fp; // auto-invokes SomeProc
	if funcPtr, ok := val.(FunctionPointerCallable); ok {
		if funcPtr.ParamCount() == 0 {
			if funcPtr.IsNil() {
				exc := e.createException("Exception", "Function pointer is nil", &node.Token.Pos, ctx)
				ctx.SetException(exc)
				return &runtime.NilValue{}
			}
			metadata := FunctionPointerMetadata{
				IsLambda:   funcPtr.IsLambda(),
				Lambda:     funcPtr.GetLambdaExpr(),
				Function:   funcPtr.GetFunctionDecl(),
				Closure:    funcPtr.GetClosure(),
				SelfObject: funcPtr.GetSelfObject(),
			}
			return e.oopEngine.ExecuteFunctionPointerCall(metadata, []Value{}, node)
		}
	}

	return val
}

// VisitVarDeclStatement evaluates a variable declaration statement.
// Handles: external variables, multi-identifier declarations, inline types,
// subrange/interface wrapping, zero value initialization, type inference.
func (e *Evaluator) VisitVarDeclStatement(node *ast.VarDeclStatement, ctx *ExecutionContext) Value {
	var value Value

	// Handle external variables
	if node.IsExternal {
		// External variables only apply to single declarations
		if len(node.Names) != 1 {
			return e.newError(node, "external keyword cannot be used with multiple variable names")
		}

		externalName := node.ExternalName
		if externalName == "" {
			externalName = node.Names[0].Value
		}
		value = &runtime.ExternalVarValue{
			Name:         node.Names[0].Value,
			ExternalName: externalName,
		}
		ctx.Env().Define(node.Names[0].Value, value)
		return value
	}

	// Evaluate initializer if present
	if node.Value != nil {
		if arrayLit, ok := node.Value.(*ast.ArrayLiteralExpression); ok {
			if node.Type != nil {
				typeName := node.Type.String()
				resolvedType, err := e.resolveTypeName(typeName, ctx)
				if err != nil {
					return e.newError(node, "failed to resolve array type '%s': %v", typeName, err)
				}
				arrayType, ok := resolvedType.(*types.ArrayType)
				if !ok {
					return e.newError(node, "expected array type, got %s", resolvedType.String())
				}
				value = e.evalArrayLiteralWithExpectedType(arrayLit, arrayType, ctx)
			} else {
				value = e.Eval(node.Value, ctx)
			}
		} else if recordLit, ok := node.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
			if node.Type == nil {
				return e.newError(node, "anonymous record literal requires explicit type annotation")
			}
			typeName := node.Type.String()

			if !e.typeSystem.HasRecord(typeName) {
				return e.newError(node, "unknown type '%s'", typeName)
			}

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

		// Type conversions and wrapping if explicit type declared
		if node.Type != nil {
			typeName := node.Type.String()
			if e.typeSystem.HasSubrangeType(typeName) {
				wrappedVal, err := e.oopEngine.WrapInSubrange(value, typeName, node)
				if err != nil {
					return e.newError(node, "%v", err)
				}
				value = wrappedVal
			} else {
				if converted, ok := e.TryImplicitConversion(value, typeName, ctx); ok {
					value = converted
				}
			}

			if ident.Equal(typeName, "Variant") {
				value = runtime.BoxVariant(value)
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
	var lastValue = value
	for _, name := range node.Names {
		var nameValue Value

		if node.Value != nil {
			nameValue = value

			if node.Type != nil {
				typeName := node.Type.String()
				if e.typeSystem.HasInterface(typeName) {
					if value.Type() != "INTERFACE" {
						wrapped, err := e.oopEngine.WrapInInterface(value, typeName, node)
						if err != nil {
							return e.newError(node, "%v", err)
						}
						nameValue = wrapped
					}
				}
			}
		} else {
			nameValue = e.createZeroValue(node.Type, node, ctx)
			if isError(nameValue) {
				return nameValue
			}
		}

		ctx.Env().Define(name.Value, nameValue)
		lastValue = nameValue
	}

	return lastValue
}

// VisitConstDecl evaluates a constant declaration.
// Supports type inference and anonymous record literals.
// Immutability is enforced by semantic analysis, not at runtime.
func (e *Evaluator) VisitConstDecl(node *ast.ConstDecl, ctx *ExecutionContext) Value {
	if node.Value == nil {
		return e.newError(node, "constant '%s' must have a value", node.Name.Value)
	}

	var value Value

	if recordLit, ok := node.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
		if node.Type == nil {
			return e.newError(node, "anonymous record literal requires explicit type annotation")
		}
		typeName := node.Type.String()

		if !e.typeSystem.HasRecord(typeName) {
			return e.newError(node, "unknown type '%s'", typeName)
		}

		ctx.SetRecordTypeContext(typeName)
		value = e.Eval(recordLit, ctx)
		ctx.ClearRecordTypeContext()
	} else {
		value = e.Eval(node.Value, ctx)
	}

	if isError(value) {
		return value
	}

	ctx.Env().Define(node.Name.Value, value)
	return value
}

// VisitAssignmentStatement evaluates an assignment statement.
// Handles: simple assignment, compound operators, index assignment, member assignment.
// Complex cases (properties, class variables) delegate to adapter.
func (e *Evaluator) VisitAssignmentStatement(node *ast.AssignmentStatement, ctx *ExecutionContext) Value {
	isCompound := node.Operator != token.ASSIGN && node.Operator != token.TokenType(0)

	switch target := node.Target.(type) {
	case *ast.Identifier:
		if isCompound {
			return e.evalCompoundIdentifierAssignment(target, node, ctx)
		}

		// Disambiguation for `[...]` literals: in DWScript, brackets can represent sets.
		// If the target is a set type, evaluate any bracket literal as a set literal.
		if arrLit, ok := node.Value.(*ast.ArrayLiteralExpression); ok {
			if expectedSetType := e.getSetTypeFromTarget(target, ctx); expectedSetType != nil {
				setLit := &ast.SetLiteral{
					Elements:            arrLit.Elements,
					TypedExpressionBase: arrLit.TypedExpressionBase,
				}

				// Provide type information for empty `[]` inference.
				if e.semanticInfo != nil {
					typeName := expectedSetType.String()
					if targetAnnot := e.semanticInfo.GetType(target); targetAnnot != nil && targetAnnot.Name != "" {
						typeName = targetAnnot.Name
					}
					e.semanticInfo.SetType(setLit, &ast.TypeAnnotation{Token: setLit.Token, Name: typeName})
					defer e.semanticInfo.ClearType(setLit)
				}

				value := e.evalSetLiteralDirect(setLit, ctx)
				if isError(value) {
					return value
				}
				if ctx.Exception() != nil {
					return &runtime.NilValue{}
				}
				return e.evalSimpleAssignmentDirect(target, value, node, ctx)
			}
		}

		// Context inference for array literals
		if _, isArrayLit := node.Value.(*ast.ArrayLiteralExpression); isArrayLit {
			if expectedType := e.getArrayTypeFromTarget(target, ctx); expectedType != nil {
				ctx.SetArrayTypeContext(expectedType)
				defer ctx.ClearArrayTypeContext()
			}
		}

		// Context inference for anonymous record literals
		if recordLit, isRecordLit := node.Value.(*ast.RecordLiteralExpression); isRecordLit && recordLit.TypeName == nil {
			if recordTypeName := e.getRecordTypeNameFromTarget(target, ctx); recordTypeName != "" {
				ctx.SetRecordTypeContext(recordTypeName)
				defer ctx.ClearRecordTypeContext()
			}
		}

		value := e.Eval(node.Value, ctx)
		if isError(value) {
			return value
		}

		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}

		// Records have value semantics - copy when assigning
		if value != nil && value.Type() == "RECORD" {
			if copyable, ok := value.(runtime.CopyableValue); ok {
				if copied := copyable.Copy(); copied != nil {
					if copiedValue, ok := copied.(Value); ok {
						value = copiedValue
					}
				}
			}
		}

		return e.evalSimpleAssignmentDirect(target, value, node, ctx)

	case *ast.MemberAccessExpression:
		if isCompound {
			// Compound member assignment (obj.field += value)
			// Pattern: Read current value → apply operation → write back
			return e.evalCompoundMemberAssignment(target, node, ctx)
		}

		value := e.Eval(node.Value, ctx)
		if isError(value) {
			return value
		}

		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}

		return e.evalMemberAssignmentDirect(target, value, node, ctx)

	case *ast.IndexExpression:
		if isCompound {
			// Compound index assignment (arr[i] += value)
			// Pattern: Read current value → apply operation → write back
			return e.evalCompoundIndexAssignment(target, node, ctx)
		}

		value := e.Eval(node.Value, ctx)
		if isError(value) {
			return value
		}

		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}

		return e.evalIndexAssignmentDirect(target, value, node, ctx)

	default:
		return e.newError(node, "invalid assignment target type: %T", target)
	}
}

// VisitBlockStatement evaluates a block statement (begin...end).
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

	stepValue := int64(1) // Default step
	if node.Step != nil {
		stepVal := e.Eval(node.Step, ctx)
		if isError(stepVal) {
			return stepVal
		}

		stepInt, ok := stepVal.(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "for loop step value must be integer, got %s", stepVal.Type())
		}

		if stepInt.Value <= 0 {
			return e.newError(node, "FOR loop STEP should be strictly positive: %d", stepInt.Value)
		}

		stepValue = stepInt.Value
	}

	ctx.PushEnv()
	defer ctx.PopEnv()

	loopVarName := node.Variable.Value

	if node.Direction == ast.ForTo {
		// Ascending loop with step support
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
		// Descending loop with step support
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
// Iterates over arrays, sets, strings, and enum types.
func (e *Evaluator) VisitForInStatement(node *ast.ForInStatement, ctx *ExecutionContext) Value {
	var result Value = &runtime.NilValue{}

	// Evaluate the collection expression
	collectionVal := e.Eval(node.Collection, ctx)
	if isError(collectionVal) {
		return collectionVal
	}

	ctx.PushEnv()
	defer ctx.PopEnv()

	loopVarName := node.Variable.Value

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
		if col.SetType == nil || col.SetType.ElementType == nil {
			return e.newError(node, "invalid set type for iteration")
		}

		elementType := col.SetType.ElementType

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
		// Iterate over string characters as single-character strings
		runes := []rune(col.Value)
		for idx := 0; idx < len(runes); idx++ {
			charVal := &runtime.StringValue{Value: string(runes[idx])}
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
		// Iterate over enum type values in declaration order
		enumType, ok := col.TypeInfo.(*types.EnumType)
		if !ok {
			return e.newError(node, "for-in loop: can only iterate over enum types, got %s", col.TypeName)
		}

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
func (e *Evaluator) VisitTryStatement(node *ast.TryStatement, ctx *ExecutionContext) Value {
	// Set up finally block to run at the end using defer
	if node.FinallyClause != nil {
		defer func() {
			// Save the current exception state
			savedExc := ctx.Exception()

			// Set ExceptObject to the current exception in finally block
			oldExceptObject, _ := ctx.Env().Get("ExceptObject")
			if savedExc != nil {
				excInstance := e.getExceptionInstance(savedExc)
				if excInstance != nil {
					ctx.Env().Set("ExceptObject", excInstance)
				}
			}

			// Clear exception so finally block can execute
			ctx.SetException(nil)

			e.Eval(node.FinallyClause.Block, ctx)

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
	e.Eval(node.TryBlock, ctx)

	// If an exception occurred, try to handle it
	if ctx.Exception() != nil {
		if node.ExceptClause != nil {
			e.evalExceptClause(node.ExceptClause, ctx)
		}
	}

	return nil
}

// evalExceptClause evaluates an except clause.
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
		if e.matchesExceptionType(exc, handler.ExceptionType) {
			// Create new scope for exception variable
			ctx.PushEnv()
			defer ctx.PopEnv()

			// Get exception instance once (for both variable binding and ExceptObject)
			excInstance := e.getExceptionInstance(exc)

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

			e.Eval(handler.Statement, ctx)

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
		ctx.SetException(nil)
		e.Eval(clause.ElseBlock, ctx)
	}
}

// VisitRaiseStatement evaluates a raise statement (exception throwing).
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

	excVal := e.Eval(node.Exception, ctx)
	if isError(excVal) {
		return excVal
	}

	excObj := e.createExceptionFromObject(excVal, ctx, node.Pos())
	ctx.SetException(excObj)

	return nil
}

// matchesExceptionType checks if an exception matches a handler's exception type.
func (e *Evaluator) matchesExceptionType(exc interface{}, typeExpr ast.TypeExpression) bool {
	// Nil type expression means bare handler - catches all
	if typeExpr == nil {
		return true
	}

	// Get the handler's exception type name
	handlerTypeName := typeExpr.String()

	// Get the exception's type name
	// All values implement Type() string method
	type TypedValue interface {
		Type() string
	}

	excVal, ok := exc.(TypedValue)
	if !ok {
		return false
	}

	excTypeName := excVal.Type()

	// Use TypeSystem to check class hierarchy
	// IsClassDescendantOf returns true if excTypeName == handlerTypeName or if excTypeName inherits from handlerTypeName
	return e.typeSystem.IsClassDescendantOf(excTypeName, handlerTypeName)
}

// getExceptionInstance extracts the ObjectInstance from an ExceptionValue.
func (e *Evaluator) getExceptionInstance(exc interface{}) Value {
	// Define local interface to access Instance field without importing parent package.
	// ExceptionValue in parent package implements GetInstance() method.
	type ExceptionWithInstance interface {
		GetInstance() interface{} // Returns *ObjectInstance but we can't import that type
	}

	if excWithInst, ok := exc.(ExceptionWithInstance); ok {
		// GetInstance returns *ObjectInstance which implements Value interface
		instance := excWithInst.GetInstance()
		if instance == nil {
			return nil
		}
		// Type assert to Value (ObjectInstance implements Value)
		if val, ok := instance.(Value); ok {
			return val
		}
	}

	return nil
}

// createExceptionFromObject creates an ExceptionValue from an object instance.
// Handles nil objects by creating a standard "Object not instantiated" exception.
func (e *Evaluator) createExceptionFromObject(obj Value, ctx *ExecutionContext, pos any) any {
	// Handle nil object case -> raise standard "Object not instantiated" exception
	if obj == nil || obj.Type() == "NIL" {
		// Get Exception class from type system
		excClass := e.typeSystem.LookupClass("Exception")
		if excClass == nil {
			panic("runtime error: Exception class not found")
		}

		message := "Object not instantiated"
		if pos != nil {
			message = fmt.Sprintf("Object not instantiated [position: %v]", pos)
		}

		lexerPos, _ := pos.(*lexer.Position)
		return e.createException("Exception", message, lexerPos, ctx)
	}

	lexerPos, _ := pos.(*lexer.Position)
	return e.wrapObjectAsException(obj, lexerPos, ctx)
}

// VisitBreakStatement evaluates a break statement.
func (e *Evaluator) VisitBreakStatement(node *ast.BreakStatement, ctx *ExecutionContext) Value {
	ctx.ControlFlow().SetBreak()
	return &runtime.NilValue{}
}

// VisitContinueStatement evaluates a continue statement.
func (e *Evaluator) VisitContinueStatement(node *ast.ContinueStatement, ctx *ExecutionContext) Value {
	ctx.ControlFlow().SetContinue()
	return &runtime.NilValue{}
}

// VisitExitStatement evaluates an exit statement.
func (e *Evaluator) VisitExitStatement(node *ast.ExitStatement, ctx *ExecutionContext) Value {
	ctx.ControlFlow().SetExit()
	if node.ReturnValue != nil {
		// Set record type context if returning anonymous record literal
		contextSet := false
		if returnType := ctx.GetCurrentFunctionReturnType(); returnType != "" {
			if recordLit, ok := node.ReturnValue.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
				if e.typeSystem.HasRecord(returnType) {
					ctx.SetRecordTypeContext(returnType)
					contextSet = true
				}
			}
		}

		value := e.Eval(node.ReturnValue, ctx)

		if contextSet {
			ctx.ClearRecordTypeContext()
		}

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
// Used in lambda expressions and explicit returns.
func (e *Evaluator) VisitReturnStatement(node *ast.ReturnStatement, ctx *ExecutionContext) Value {
	// Evaluate the return value
	var returnVal Value
	if node.ReturnValue != nil {
		// Set record type context if returning anonymous record literal
		contextSet := false
		if returnType := ctx.GetCurrentFunctionReturnType(); returnType != "" {
			if recordLit, ok := node.ReturnValue.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
				if e.typeSystem.HasRecord(returnType) {
					ctx.SetRecordTypeContext(returnType)
					contextSet = true
				}
			}
		}

		returnVal = e.Eval(node.ReturnValue, ctx)

		if contextSet {
			ctx.ClearRecordTypeContext()
		}

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
// ============================================================================

// createZeroValue creates a zero value for the given type.
func (e *Evaluator) createZeroValue(typeExpr ast.TypeExpression, node ast.Node, ctx *ExecutionContext) Value {
	if typeExpr == nil {
		return &runtime.NilValue{}
	}

	if arrayNode, ok := typeExpr.(*ast.ArrayTypeNode); ok {
		arrayType := e.resolveArrayTypeNode(arrayNode, ctx)
		if arrayType != nil {
			return &runtime.ArrayValue{ArrayType: arrayType, Elements: []runtime.Value{}}
		}
		return &runtime.NilValue{}
	}

	typeName := typeExpr.String()

	if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
		arrayType := e.parseInlineArrayType(typeName, ctx)
		if arrayType != nil {
			return &runtime.ArrayValue{ArrayType: arrayType, Elements: []runtime.Value{}}
		}
		return &runtime.NilValue{}
	}

	if strings.HasPrefix(typeName, "set of ") {
		setType := e.parseInlineSetType(typeName, ctx)
		if setType == nil {
			return &runtime.NilValue{}
		}
		return runtime.NewSetValue(setType)
	}

	if e.typeSystem.HasRecord(typeName) {
		// Look up record type via TypeSystem
		recordTypeAny := e.typeSystem.LookupRecord(typeName)
		if recordTypeAny == nil {
			return &runtime.NilValue{}
		}

		// Type-assert to access RecordType, Metadata, and FieldDecls
		type recordTypeAccess interface {
			GetRecordType() *types.RecordType
			GetMetadata() any
		}

		recordTypeAccessor, ok := recordTypeAny.(recordTypeAccess)
		if !ok {
			return &runtime.NilValue{}
		}

		recordType := recordTypeAccessor.GetRecordType()
		if recordType == nil {
			return &runtime.NilValue{}
		}

		// Extract Metadata (may be nil)
		var metadata *runtime.RecordMetadata
		if mdAny := recordTypeAccessor.GetMetadata(); mdAny != nil {
			if md, ok := mdAny.(*runtime.RecordMetadata); ok {
				metadata = md
			}
		}

		// Extract FieldDecls for field initializer evaluation
		var fieldDecls map[string]*ast.FieldDecl
		type hasFieldDecls interface {
			GetFieldDecls() map[string]*ast.FieldDecl
		}
		if rtVal, ok := recordTypeAny.(hasFieldDecls); ok {
			fieldDecls = rtVal.GetFieldDecls()
		}

		// Create field initializer callback for runtime constructor
		initializer := func(fieldName string, fieldType types.Type) runtime.Value {
			fieldNameNorm := ident.Normalize(fieldName)

			// Check for field initializer expression in FieldDecls
			if fieldDecls != nil {
				if fieldDecl, hasDecl := fieldDecls[fieldNameNorm]; hasDecl && fieldDecl.InitValue != nil {
					// Evaluate the field initializer AST expression directly
					fieldValue := e.Eval(fieldDecl.InitValue, ctx)
					if isError(fieldValue) {
						return fieldValue
					}
					return fieldValue
				}
			}

			// No initializer - generate zero value
			return e.getZeroValueForType(fieldType)
		}

		recordValue := runtime.NewRecordValueWithInitializer(recordType, metadata, initializer)

		return recordValue
	}

	if e.typeSystem.HasArrayType(typeName) {
		arrayType := e.typeSystem.LookupArrayType(typeName)
		if arrayType == nil {
			return &runtime.NilValue{}
		}
		return runtime.NewArrayValue(arrayType, nil)
	}

	if subrangeType := e.typeSystem.LookupSubrangeType(typeName); subrangeType != nil {
		return runtime.NewSubrangeValueZero(subrangeType)
	}

	if e.typeSystem.HasInterface(typeName) {
		// Lookup interface metadata from TypeSystem
		ifaceInfoAny := e.typeSystem.LookupInterface(typeName)
		if ifaceInfoAny == nil {
			return &runtime.NilValue{}
		}
		// Type-assert to IInterfaceInfo interface
		ifaceInfo, ok := ifaceInfoAny.(runtime.IInterfaceInfo)
		if !ok {
			return &runtime.NilValue{}
		}
		// Create nil interface instance directly
		return runtime.NewInterfaceInstance(ifaceInfo, nil)
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
		return runtime.BoxVariant(&runtime.NilValue{})
	default:
		if e.typeSystem.HasClass(typeName) {
			return &runtime.NilValue{ClassType: typeName}
		}
		return &runtime.NilValue{}
	}
}
