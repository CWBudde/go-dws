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

	// Program-scope finalization: associative arrays release the keys and
	// values they own, running the destructors of objects the map outlived.
	// Deferred so it also covers the error and uncaught-exception exits below.
	// The environment is resolved inside the closure, because the program's
	// own declarations are only defined after this function has been entered.
	//
	// Destructor bodies do not execute while an exception is still active, so
	// the pending exception is parked for the duration of the finalization and
	// restored afterwards. It has already been turned into the returned error
	// at that point; restoring it only keeps the context state truthful.
	defer func() {
		pending := ctx.Exception()
		ctx.SetException(nil)
		e.releaseAssociativeBindings(ctx.Env())
		ctx.SetException(pending)
	}()

	if result = e.predeclareProgramClassTypes(node, ctx); isError(result) {
		return result
	}

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
		if exc, ok := ctx.Exception().(*runtime.ExceptionValue); ok && exc != nil {
			message := exc.Message
			if message == "" {
				message = exc.Inspect()
			}
			// Match DWScript's unhandled-exception report: explicit raises are
			// prefixed "User defined exception:", the raise position is
			// appended, and each call site is listed labeled by its containing
			// routine (see StackTrace.DWScriptString).
			if exc.UserRaised {
				message = "User defined exception: " + message
			}
			if exc.Position != nil && exc.Position.IsValid() && !strings.Contains(message, "[line:") {
				message = fmt.Sprintf("%s [line: %d, column: %d]", message, exc.Position.Line, exc.Position.Column)
			}
			if trace := exc.CallStack.DWScriptString(); trace != "" {
				message += "\n" + trace
			}
			message = formatDWScriptExceptionMessage(message)
			return &runtime.ErrorValue{Message: message}
		}
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

func formatDWScriptExceptionMessage(message string) string {
	message = strings.TrimSpace(message)
	switch {
	case strings.HasPrefix(message, "cannot cast interface of "):
		return "Cannot cast interface of " + strings.TrimPrefix(message, "cannot cast interface of ")
	case strings.HasPrefix(message, "class \"") && strings.Contains(message, "\" does not implement interface "):
		return "Class " + strings.TrimPrefix(message, "class ")
	default:
		return message
	}
}

func (e *Evaluator) predeclareProgramClassTypes(node *ast.Program, ctx *ExecutionContext) Value {
	if node == nil {
		return &runtime.NilValue{}
	}

	for _, stmt := range node.Statements {
		if result := e.predeclareClassTypesInStatement(stmt, ctx); isError(result) {
			return result
		}
	}

	return &runtime.NilValue{}
}

func (e *Evaluator) predeclareClassTypesInStatement(stmt ast.Statement, ctx *ExecutionContext) Value {
	switch n := stmt.(type) {
	case *ast.BlockStatement:
		for _, inner := range n.Statements {
			if result := e.predeclareClassTypesInStatement(inner, ctx); isError(result) {
				return result
			}
		}
	case *ast.ClassDecl:
		if n.EnclosingClass != nil {
			return &runtime.NilValue{}
		}
		className := e.fullClassNameFromDecl(n)
		if className == "" || e.typeSystem.LookupClass(className) != nil {
			return &runtime.NilValue{}
		}
		rawClassInfo, err := e.typeSystem.NewClassInfo(className)
		if err != nil {
			return e.newError(n, "internal error: %s", err.Error())
		}
		ci, ok := rawClassInfo.(classDeclarationInfo)
		if !ok {
			return e.newError(n, "internal error: invalid class type for '%s'", className)
		}
		ci.SetForwardClass(true)
		ci.RegisterInTypeSystem(e.typeSystem, "")
		ci.DefineInEnv(ctx.Env())
	}

	return &runtime.NilValue{}
}

// VisitUnitDeclaration evaluates a DWScript unit as a top-level compilation target.
// Runtime execution processes interface declarations, then implementation declarations,
// then the initialization section. Finalization is deferred to unit shutdown handling.
func (e *Evaluator) VisitUnitDeclaration(node *ast.UnitDeclaration, ctx *ExecutionContext) Value {
	var result Value = &runtime.NilValue{}

	sections := []*ast.BlockStatement{
		node.InterfaceSection,
		node.ImplementationSection,
		node.InitSection,
	}

	for _, section := range sections {
		if section == nil {
			continue
		}

		result = e.Eval(section, ctx)
		if isError(result) {
			return result
		}
		if ctx.Exception() != nil {
			return result
		}
		if ctx.ControlFlow().IsExit() {
			ctx.ControlFlow().Clear()
			return result
		}
		if ctx.ControlFlow().IsActive() {
			return result
		}
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
			if !strings.Contains(errVal.Message, "line:") && !strings.Contains(errVal.Message, lineMarker) {
				errVal.Message = errVal.Message + "\n " + loc
			}
		}
		return val
	}

	// Auto-invoke parameterless function pointers
	// Example: var fp := @SomeProc; fp; // auto-invokes SomeProc
	// Built-in function pointers (IntToStr, etc.) always take arguments, so they
	// are never auto-invoked here even though their AST-free value reports no
	// declared parameters.
	if funcPtr, ok := val.(FunctionPointerCallable); ok {
		if funcPtr.ParamCount() == 0 && funcPtr.GetBuiltinName() == "" {
			if funcPtr.IsNil() {
				pos := node.Expression.End()
				msg := fmt.Sprintf("Function pointer is nil [line: %d, column: %d]", pos.Line, pos.Column)
				exc := e.createException("Exception", msg, &node.Token.Pos, ctx)
				ctx.SetException(exc)
				return &runtime.NilValue{}
			}
			return e.executeFunctionPointerDirect(val, []Value{}, node, ctx)
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
		if node.Type != nil {
			if resolvedType, err := e.ResolveTypeFromAnnotation(node.Type, ctx); err == nil && resolvedType != nil {
				if resolvedType.TypeKind() == "FUNCTION_POINTER" || resolvedType.TypeKind() == "METHOD_POINTER" {
					if memberAccess, ok := node.Value.(*ast.MemberAccessExpression); ok {
						value = e.buildMethodPointerFromMemberAccess(memberAccess, ctx)
						if isError(value) {
							return value
						}
					}
				}
			}
		}

		if value == nil {
			if arrayLit, ok := node.Value.(*ast.ArrayLiteralExpression); ok {
				if node.Type != nil {
					typeName := node.Type.String()
					resolvedType, err := e.ResolveTypeFromAnnotation(node.Type, ctx)
					if err != nil {
						return e.newError(node, "failed to resolve array type '%s': %v", typeName, err)
					}
					// A bracket literal against a `set of` declaration is a set
					// constructor, not an array one. Only `[]` reaches here as an
					// ArrayLiteralExpression — the parser already classifies a
					// non-empty `[a, b]` as a SetLiteral.
					if setType, isSet := types.GetUnderlyingType(resolvedType).(*types.SetType); isSet {
						value = e.evalBracketLiteralAsSet(arrayLit, setType, ctx)
					} else if arrayType, isArray := resolvedType.(*types.ArrayType); isArray {
						value = e.evalArrayLiteralWithExpectedType(arrayLit, arrayType, ctx)
					} else {
						return e.newError(node, "expected array type, got %s", resolvedType.String())
					}
					if isError(value) {
						return value
					}
				} else {
					value = e.Eval(node.Value, ctx)
				}
			} else if recordLit, ok := node.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
				if node.Type == nil {
					return e.newError(node, "anonymous record literal requires explicit type annotation")
				}
				typeName := node.Type.String()

				if e.recordTypeFromAnnotation(node.Type, ctx) == nil {
					return e.newError(node, "unknown type '%s'", typeName)
				}

				previousRecordType := ctx.RecordTypeContext()
				ctx.SetRecordTypeContext(e.recordTypeFromAnnotation(node.Type, ctx))
				value = e.Eval(recordLit, ctx)
				ctx.SetRecordTypeContext(previousRecordType)
			} else {
				value = e.Eval(node.Value, ctx)
			}
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
				wrappedVal, err := e.wrapInSubrange(value, typeName, node)
				if err != nil {
					return e.newError(node, "%v", err)
				}
				value = wrappedVal
			} else {
				if converted, ok := e.TryImplicitConversionFromAnnotation(value, node.Type, ctx); ok {
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
			// Clone copyable values (arrays, records) unless it's an index expression
			if _, isIndexExpr := node.Value.(*ast.IndexExpression); isIndexExpr {
				nameValue = value
			} else {
				nameValue = cloneIfCopyable(value)
			}

			if node.Type != nil {
				typeName := node.Type.String()
				if e.typeSystem.HasInterface(typeName) {
					if runtime.KindOf(value) != runtime.KindInterface {
						wrapped, err := e.wrapInInterface(value, typeName, node)
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

		nameValue = e.retainValueForBinding(nameValue, ctx)
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

		if e.recordTypeFromAnnotation(node.Type, ctx) == nil {
			return e.newError(node, "unknown type '%s'", typeName)
		}

		previousRecordType := ctx.RecordTypeContext()
		ctx.SetRecordTypeContext(e.recordTypeFromAnnotation(node.Type, ctx))
		value = e.Eval(recordLit, ctx)
		ctx.SetRecordTypeContext(previousRecordType)
	} else {
		value = e.Eval(node.Value, ctx)
	}

	if isError(value) {
		return value
	}

	value = e.retainValueForBinding(value, ctx)
	ctx.Env().Define(node.Name.Value, value)
	return value
}

// VisitAssignmentStatement evaluates an assignment statement.
// Handles: simple assignment, compound operators, index assignment, member assignment.
// Complex cases are routed through evaluator-owned helpers and runtime metadata.
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
					Elements: arrLit.Elements,
					BaseNode: arrLit.BaseNode,
				}

				// Provide type information for empty `[]` inference.
				if e.SemanticInfo() != nil {
					typeName := expectedSetType.String()
					if targetAnnot := e.SemanticInfo().GetType(target); targetAnnot != nil && targetAnnot.Name != "" {
						typeName = targetAnnot.Name
					}
					e.SemanticInfo().SetType(setLit, &ast.TypeAnnotation{Token: setLit.Token, Name: typeName})
					e.SemanticInfo().SetResolvedType(setLit, expectedSetType)
					defer e.SemanticInfo().ClearType(setLit)
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
			if recordType := e.getRecordTypeFromTarget(target, ctx); recordType != nil {
				previousRecordType := ctx.RecordTypeContext()
				ctx.SetRecordTypeContext(recordType)
				defer ctx.SetRecordTypeContext(previousRecordType)
			}
		}

		var value Value
		if memberAccess, ok := node.Value.(*ast.MemberAccessExpression); ok {
			expectedTypeKind := e.expectedTypeKindForIdentifier(target, ctx)
			if expectedTypeKind == "FUNCTION_POINTER" || expectedTypeKind == "METHOD_POINTER" {
				value = e.buildMethodPointerFromMemberAccess(memberAccess, ctx)
				if isError(value) {
					return value
				}
			}
		}
		if value == nil {
			value = e.Eval(node.Value, ctx)
		}
		if isError(value) {
			return value
		}

		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}

		// Auto-box a base scalar assigned to a JSONVariant-typed target as a JSON
		// immediate. This keys off the declared type only: a plain Variant that
		// happens to hold a JSON value must NOT coerce a later scalar assignment.
		if e.expectedTypeKindForIdentifier(target, ctx) == "JSON_VARIANT" {
			value = coerceToJSONVariant(value)
		}

		// Records have value semantics - copy when assigning
		if record, ok := value.(*runtime.RecordValue); ok {
			value = record.Copy()
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

	ctx.PushEnv()
	defer ctx.PopEnv()

	startVal := e.Eval(node.Start, ctx)
	if isError(startVal) {
		return startVal
	}

	loopVarName := node.Variable.Value
	ctx.Env().Define(loopVarName, startVal)

	endVal := e.Eval(node.EndValue, ctx)
	if isError(endVal) {
		return endVal
	}

	startOrdinal, err := runtime.GetOrdinalValue(startVal)
	if err != nil {
		return e.newError(node.Start, "for loop start value must be ordinal, got %s", startVal.Type())
	}

	endOrdinal, err := runtime.GetOrdinalValue(endVal)
	if err != nil {
		return e.newError(node.EndValue, "for loop end value must be ordinal, got %s", endVal.Type())
	}

	stepOrdinal := 1
	if node.Step != nil {
		stepVal := e.Eval(node.Step, ctx)
		if isError(stepVal) {
			return stepVal
		}

		stepOrdinal, err = runtime.GetOrdinalValue(stepVal)
		if err != nil {
			return e.newError(node.Step, "for loop step value must be ordinal, got %s", stepVal.Type())
		}

		if stepOrdinal <= 0 {
			pos := node.Token.Pos
			message := fmt.Sprintf("FOR loop STEP should be strictly positive: %d [line: %d, column: %d]",
				stepOrdinal, pos.Line, pos.Column)
			ctx.SetException(e.createException("Exception", message, &pos, ctx))
			return &runtime.NilValue{}
		}
	}

	if node.Direction == ast.ForTo {
		for current := startOrdinal; current <= endOrdinal; {
			currentVal, err := runtime.RebuildOrdinalValue(startVal, current, e.lookupEnumType)
			if err != nil {
				return e.newError(node, "%s", err.Error())
			}
			ctx.Env().Define(loopVarName, currentVal)

			result = e.Eval(node.Body, ctx)
			if isError(result) {
				return result
			}

			cf := ctx.ControlFlow()
			if cf.IsBreak() {
				cf.Clear()
				break
			}
			if cf.IsContinue() {
				cf.Clear()
			}
			if cf.IsExit() {
				break
			}
			if current > endOrdinal-stepOrdinal {
				break
			}
			current += stepOrdinal
		}
	} else {
		for current := startOrdinal; current >= endOrdinal; {
			currentVal, err := runtime.RebuildOrdinalValue(startVal, current, e.lookupEnumType)
			if err != nil {
				return e.newError(node, "%s", err.Error())
			}
			ctx.Env().Define(loopVarName, currentVal)

			result = e.Eval(node.Body, ctx)
			if isError(result) {
				return result
			}

			cf := ctx.ControlFlow()
			if cf.IsBreak() {
				cf.Clear()
				break
			}
			if cf.IsContinue() {
				cf.Clear()
			}
			if cf.IsExit() {
				break
			}
			if current < endOrdinal+stepOrdinal {
				break
			}
			current -= stepOrdinal
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

	loopVarName := node.Variable.Value
	stepOrdinal := 1
	if node.Step != nil {
		stepVal := e.Eval(node.Step, ctx)
		if isError(stepVal) {
			return stepVal
		}
		var err error
		stepOrdinal, err = runtime.GetOrdinalValue(stepVal)
		if err != nil {
			return e.newError(node.Step, "for-in loop step value must be ordinal, got %s", stepVal.Type())
		}
		if stepOrdinal <= 0 {
			pos := node.Token.Pos
			message := fmt.Sprintf("FOR loop STEP should be strictly positive: %d [line: %d, column: %d]",
				stepOrdinal, pos.Line, pos.Column)
			ctx.SetException(e.createException("Exception", message, &pos, ctx))
			return &runtime.NilValue{}
		}
	}

	stringAsOrdinal := false
	if !node.InlineVar {
		if current, ok := ctx.Env().Get(loopVarName); ok {
			if wrapper, ok := current.(runtime.VariantWrapper); ok {
				current = wrapper.UnwrapVariant()
			}
			_, stringAsOrdinal = current.(*runtime.IntegerValue)
		}
	}

	ctx.PushEnv()
	defer ctx.PopEnv()

	runBody := func(loopValue Value) (bool, Value) {
		if node.InlineVar {
			ctx.Env().Define(loopVarName, loopValue)
		} else if err := ctx.Env().Set(loopVarName, loopValue); err != nil {
			ctx.Env().Define(loopVarName, loopValue)
		}

		result = e.Eval(node.Body, ctx)
		if isError(result) {
			return true, result
		}
		if ctx.Exception() != nil {
			return true, result
		}

		cf := ctx.ControlFlow()
		if cf.IsBreak() {
			cf.Clear()
			return true, result
		}
		if cf.IsContinue() {
			cf.Clear()
			return false, result
		}
		if cf.IsExit() {
			return true, result
		}
		return false, result
	}

	switch col := collectionVal.(type) {
	case *runtime.ArrayValue:
		// Iterate over array elements
		for idx := 0; idx < len(col.Elements); idx += stepOrdinal {
			stop, val := runBody(runtime.CopyValue(col.Elements[idx]))
			if isError(val) {
				return val
			}
			if stop {
				break
			}
		}

	case *runtime.SetValue:
		if col.SetType == nil || col.SetType.ElementType == nil {
			return e.newError(node, "invalid set type for iteration")
		}

		elementType := col.SetType.ElementType

		if enumType, ok := elementType.(*types.EnumType); ok {
			stepIndex := 0
			for _, name := range enumType.OrderedNames {
				ordinal := enumType.Values[name]
				// Check if this enum value is in the set
				if col.HasElement(ordinal) {
					if stepIndex%stepOrdinal != 0 {
						stepIndex++
						continue
					}
					// Create an enum value for this element
					enumVal := &runtime.EnumValue{
						EnumType:     enumType,
						TypeName:     enumType.Name,
						ValueName:    name,
						OrdinalValue: ordinal,
					}

					stop, val := runBody(enumVal)
					if isError(val) {
						return val
					}
					stepIndex++
					if stop {
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
		for idx := 0; idx < len(runes); idx += stepOrdinal {
			var loopValue Value
			if stringAsOrdinal {
				loopValue = &runtime.IntegerValue{Value: int64(runes[idx])}
			} else {
				loopValue = &runtime.StringValue{Value: string(runes[idx])}
			}
			stop, val := runBody(loopValue)
			if isError(val) {
				return val
			}
			if stop {
				break
			}
		}

	case *runtime.TypeMetaValue:
		// Iterate over enum type values by ordinal range (min to max)
		// DWScript iterates over the full range from min to max, not just declared values
		enumType, ok := types.GetUnderlyingType(col.TypeInfo).(*types.EnumType)
		if !ok {
			return e.newError(node, "for-in loop: can only iterate over enum types, got %s", col.TypeName)
		}

		for ordinal := enumType.MinOrdinal(); ordinal <= enumType.MaxOrdinal(); ordinal += stepOrdinal {
			enumVal := runtime.NewEnumValue(enumType.Name, enumType, ordinal)

			stop, val := runBody(enumVal)
			if isError(val) {
				return val
			}
			if stop {
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

			// Save and suspend any pending control-flow signal (Exit/Break/
			// Continue) so the finally block runs completely. Without this,
			// an Exit raised in the try block would make the finally block
			// stop after its first statement, or be swallowed by a function
			// call inside the finally block (see fixtures exit_finally,
			// exit_finally2).
			savedFlow := ctx.ControlFlow().Kind()
			ctx.ControlFlow().Clear()

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

			// Re-arm the suspended control-flow signal unless the finally
			// block itself raised a new one (which takes precedence).
			if !ctx.ControlFlow().IsActive() {
				ctx.ControlFlow().Restore(savedFlow)
			}

			// Restore ExceptObject
			ctx.Env().Set("ExceptObject", oldExceptObject)
		}()
	}

	// Execute try block
	tryResult := e.Eval(node.TryBlock, ctx)

	// Runtime errors (ErrorValue) raised inside the try block are catchable in
	// DWScript: convert them into a script exception so except handlers see them.
	if isError(tryResult) && ctx.Exception() == nil {
		e.raiseErrorValueAsException(tryResult, currentRoutineName(ctx), ctx)
	}

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
		if e.matchesExceptionType(exc, handler.ExceptionType, ctx) {
			// Create new scope for exception variable
			ctx.PushEnv()
			defer ctx.PopEnv()

			// Get exception instance once (for both variable binding and ExceptObject)
			excInstance := e.getExceptionInstance(exc)

			// Populate the exception's StackTrace field from the raise-time call
			// stack so `E.StackTrace` reads the DWScript-format trace.
			if ev, ok := exc.(*runtime.ExceptionValue); ok {
				if inst, ok := excInstance.(*runtime.ObjectInstance); ok {
					inst.SetField("StackTrace", &runtime.StringValue{Value: ev.StackTraceString()})
				}
			}

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

	// Raising a nil exception reference raises "Object not instantiated",
	// reported just past the raised expression (DWScript reports the parser's
	// position after consuming the expression).
	if excVal == nil || runtime.KindOf(excVal) == runtime.KindNil {
		pos := raisedExpressionEndPos(node.Exception)
		message := fmt.Sprintf("Object not instantiated [line: %d, column: %d]", pos.Line, pos.Column)
		ctx.SetException(e.createException("Exception", message, nil, ctx))
		return nil
	}

	// DWScript reports an unhandled raise just past the raised expression (the
	// parser's position after consuming it) ...
	pos := node.Exception.End()
	excObj := e.createExceptionFromObject(excVal, ctx, &pos)
	if excValue, ok := excObj.(*runtime.ExceptionValue); ok {
		excValue.UserRaised = true
		// ... but the innermost stack-trace frame is the site where the
		// exception object was constructed, at the constructor's name token.
		originPos := raiseSitePos(node.Exception)
		excValue.OriginPos = &originPos
	}
	ctx.SetException(excObj)

	return nil
}

// raisedExpressionEndPos approximates the source position immediately after a
// raised expression (identifiers advance by their length; other expressions
// fall back to their start position).
func raisedExpressionEndPos(expr ast.Expression) token.Position {
	pos := expr.Pos()
	if identExpr, ok := expr.(*ast.Identifier); ok {
		pos.Column += len(identExpr.Value)
	}
	return pos
}

// matchesExceptionType checks if an exception matches a handler's exception type.
func (e *Evaluator) matchesExceptionType(exc interface{}, typeExpr ast.TypeExpression, ctx *ExecutionContext) bool {
	if typeExpr == nil {
		return true
	}
	handlerType, err := e.ResolveTypeFromAnnotation(typeExpr, ctx)
	if err != nil {
		return false
	}
	exception, ok := exc.(*runtime.ExceptionValue)
	if !ok {
		return false
	}
	var actual types.Type
	if exception.Instance != nil && exception.Instance.Class != nil {
		actual = exception.Instance.Class.GetClassType()
	} else if exception.Metadata != nil {
		if class := e.typeSystem.LookupClass(exception.Metadata.Name); class != nil {
			actual = class.GetClassType()
		}
	}
	return types.OperatorTypesCompatible(actual, handlerType)
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
	if obj == nil || runtime.KindOf(obj) == runtime.KindNil {
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
		previousRecordType := ctx.RecordTypeContext()
		contextSet := false
		if returnType := ctx.GetCurrentFunctionReturnType(); returnType != nil {
			if recordLit, ok := node.ReturnValue.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
				if recordType, ok := types.GetUnderlyingType(returnType).(*types.RecordType); ok {
					ctx.SetRecordTypeContext(recordType)
					contextSet = true
				}
			}
		}

		value := e.Eval(node.ReturnValue, ctx)

		if contextSet {
			ctx.SetRecordTypeContext(previousRecordType)
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
		previousRecordType := ctx.RecordTypeContext()
		contextSet := false
		if returnType := ctx.GetCurrentFunctionReturnType(); returnType != nil {
			if recordLit, ok := node.ReturnValue.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
				if recordType, ok := types.GetUnderlyingType(returnType).(*types.RecordType); ok {
					ctx.SetRecordTypeContext(recordType)
					contextSet = true
				}
			}
		}

		returnVal = e.Eval(node.ReturnValue, ctx)

		if contextSet {
			ctx.SetRecordTypeContext(previousRecordType)
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

	resolved, err := e.ResolveTypeFromAnnotation(typeExpr, ctx)
	if err != nil {
		return &runtime.NilValue{}
	}
	return e.createZeroValueForResolvedType(resolved, ctx)
}

// createZeroValueForResolvedType preserves declared type metadata when initializing
// variables and function results. Runtime field defaults use the same constructors.
func (e *Evaluator) createZeroValueForResolvedType(resolved types.Type, ctx *ExecutionContext) Value {
	switch typ := types.GetUnderlyingType(resolved).(type) {
	case *types.ArrayType:
		return e.createArrayZeroValue(typ, ctx)
	case *types.SetType:
		return runtime.NewSetValue(typ)
	case *types.SubrangeType:
		return runtime.NewSubrangeValueZero(typ)
	}
	if types.GetUnderlyingType(resolved) == types.JSON_VARIANT {
		return boxJSON(nil)
	}
	// A declared ByteBuffer variable is auto-instantiated: `var b : ByteBuffer;`
	// is immediately usable and starts empty rather than nil.
	if types.IsByteBuffer(resolved) {
		return runtime.NewByteBufferValue()
	}
	if types.GetUnderlyingType(resolved) == types.VARIANT {
		return &runtime.VariantValue{Value: nil, ActualType: nil}
	}
	return e.getZeroValueForType(resolved, ctx)
}

// createArrayZeroValue creates a properly initialized array value.
// For static arrays, it pre-allocates elements and initializes nested arrays/records.
// For dynamic arrays, it creates an empty array.
func (e *Evaluator) createArrayZeroValue(arrayType *types.ArrayType, ctx *ExecutionContext) Value {
	initializer := func(elementType types.Type, index int) runtime.Value {
		if nestedArrayType, ok := elementType.(*types.ArrayType); ok {
			return e.createArrayZeroValue(nestedArrayType, ctx)
		}
		if recordType, ok := elementType.(*types.RecordType); ok {
			return e.createRecordZeroValue(recordType, ctx)
		}
		// For basic types, return nil - runtime will use zero values
		return nil
	}
	return runtime.NewArrayValue(arrayType, initializer)
}

// createRecordZeroValue creates a properly initialized record value.
func (e *Evaluator) createRecordZeroValue(recordType *types.RecordType, ctx *ExecutionContext) Value {
	metadata := e.typeSystem.LookupRecordMetadata(recordType.Name)

	// Create field initializer for nested types
	initializer := func(fieldName string, fieldType types.Type) runtime.Value {
		return e.getZeroValueForType(fieldType, ctx)
	}

	return runtime.NewRecordValueWithInitializer(recordType, metadata, initializer)
}
