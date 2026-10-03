package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// analyzeAssignmentExpression preserves accepted callable references and reads
// rejected bare callable names as one call. The call's result may itself be a
// pointer; that result is a value, rather than another implicit invocation.
func (a *Analyzer) analyzeAssignmentExpression(value ast.Expression, expected types.Type, readRejectedReference bool) types.Type {
	previousArray := a.inArrayAssignment
	previousCall := a.assignmentCallRecovery
	a.inArrayAssignment = true
	a.assignmentCallRecovery = nil
	if readRejectedReference {
		if call, ok := value.(*ast.CallExpression); ok {
			a.assignmentCallRecovery = call
		}
	}
	defer func() {
		a.inArrayAssignment = previousArray
		a.assignmentCallRecovery = previousCall
	}()
	return a.analyzeCallableValue(value, expected, readRejectedReference)
}

// analyzeCallableValue selects a reference or one call without changing the
// enclosing context's literal inference or explicit-call recovery policy.
func (a *Analyzer) analyzeCallableValue(value ast.Expression, expected types.Type, readRejectedReference bool) types.Type {
	if member, ok := value.(*ast.MemberAccessExpression); ok {
		if result, handled := a.analyzeAssignmentUnitMember(member, expected, readRejectedReference); handled {
			return result
		}
	}
	if identifier, ok := value.(*ast.Identifier); ok {
		if result, readAsCall := a.analyzeAssignmentIdentifierCall(identifier, expected, readRejectedReference); readAsCall {
			return result
		}
	}
	actual := a.analyzeExpressionWithExpectedType(value, expected)
	// Member analysis may already have invoked a method. Only storage reads or
	// annotated method references supply an original callable here.
	if a.isAssignmentCallableReference(value) {
		if result, required, call := a.assignmentCallableResult(actual, expected, readRejectedReference); call {
			a.recordAssignmentCall(value, result, required)
			return result
		}
	}
	return actual
}

// analyzeAssignmentUnitMember resolves an active namespace before method
// probing can mistake its name for a value. Named routines use the same
// reference-versus-result selection as unqualified identifiers.
func (a *Analyzer) analyzeAssignmentUnitMember(member *ast.MemberAccessExpression, expected types.Type, readRejectedReference bool) (types.Type, bool) {
	name, ok := member.Object.(*ast.Identifier)
	if !ok || a.hasLexicalValueReceiver(member.Object) {
		return nil, false
	}
	unit, active := a.importedUnitNamespace(name.Value)
	if !active {
		return nil, false
	}
	actual := a.analyzeExpression(member)
	symbol, found := unit.Resolve(member.Member.Value)
	if !found || symbol.IsOverloadSet {
		return actual, true
	}
	result, required, call := a.assignmentCallableResult(actual, expected, readRejectedReference)
	if routine, ok := types.GetUnderlyingType(actual).(*types.FunctionType); ok {
		actual = types.FunctionPointerFromFunctionType(routine)
		a.annotateMemberPointerType(member, actual)
	}
	if call {
		a.recordAssignmentCall(member, result, required)
		return result, true
	}
	return actual, true
}

func (a *Analyzer) analyzeAssignmentIdentifierCall(identifier *ast.Identifier, expected types.Type, readRejectedReference bool) (types.Type, bool) {
	symbol, found := a.symbols.Resolve(identifier.Value)
	if !found || symbol.Type == nil || symbol.IsOverloadSet {
		return nil, false
	}
	// A routine's own name is the Result alias in its body.
	if a.currentFunction != nil && ident.Equal(a.currentFunction.Name.Value, identifier.Value) {
		return nil, false
	}

	result, required, call := a.assignmentCallableResult(symbol.Type, expected, readRejectedReference)
	if !call {
		return nil, false
	}
	// Analyze once to retain symbol usage, casing and deprecation diagnostics.
	a.analyzeIdentifier(identifier)
	a.recordAssignmentCall(identifier, result, required)
	return result, true
}

func (a *Analyzer) assignmentCallableResult(actual, expected types.Type, readRejectedReference bool) (types.Type, int, bool) {
	var pointer *types.FunctionPointerType
	required := 0
	namedRoutine := false
	switch declared := types.GetUnderlyingType(actual).(type) {
	case *types.FunctionType:
		pointer = types.FunctionPointerFromFunctionType(declared)
		actual = pointer
		required = requiredParamCount(declared)
		namedRoutine = true
	case *types.FunctionPointerType:
		pointer = declared
		actual = declared
		required = declared.RequiredParamCount()
	case *types.MethodPointerType:
		pointer = &declared.FunctionPointerType
		actual = declared
		required = declared.RequiredParamCount()
	default:
		return nil, 0, false
	}
	if expected == nil {
		// Keep inferred procedure/parameterized references; parameterless
		// functions infer their one result, including a returned callable.
		if required > 0 || pointer.IsProcedure() {
			return nil, 0, false
		}
	} else if a.assignmentAcceptsCallableReference(actual, expected, namedRoutine) {
		return nil, 0, false
	}
	result := functionPointerCallResult(pointer)
	// Typed pointer initializers select a factory call only when its returned
	// pointer fits; preserve existing rejected-reference initializer diagnostics.
	if !readRejectedReference && (!types.IsPointerType(result) || !a.canAssign(result, expected)) {
		return nil, 0, false
	}
	return result, required, true
}

func (a *Analyzer) recordAssignmentCall(value ast.Expression, result types.Type, required int) {
	if required > 0 {
		a.addMoreArgumentsExpected(value.Pos())
	}
	a.semanticInfo.SetImplicitCall(value)
	a.semanticInfo.SetResolvedType(value, result)
}

func (a *Analyzer) isAssignmentCallableReference(value ast.Expression) bool {
	switch expr := value.(type) {
	case *ast.IndexExpression:
		return true
	case *ast.MemberAccessExpression:
		// Pointer-context method analysis records the original routine on the
		// member token; ordinary method reads have already produced their result.
		if types.IsPointerType(a.semanticInfo.GetResolvedType(expr.Member)) {
			return true
		}
		receiver := a.semanticInfo.GetResolvedType(expr.Object)
		receiver = a.implicitCallTypePreview(expr.Object, receiver)
		// Match ordinary member analysis when its receiver is a callable value.
		// This previews ownership only; the evaluator reads that receiver once.
		if result := implicitValueContextType(receiver); result != nil {
			receiver = result
		}
		if class, ok := types.GetUnderlyingType(receiver).(*types.ClassOfType); ok {
			receiver = class.ClassType
		}
		return hasCallableMemberStorage(receiver, expr.Member.Value)
	}
	return false
}

func hasCallableMemberStorage(receiver types.Type, name string) bool {
	if record, isMeta := recordReceiverType(receiver); record != nil {
		_, classVar := record.ClassVars[ident.Normalize(name)]
		property := record.GetProperty(name)
		return classVar || (!isMeta && record.HasField(name)) ||
			(property != nil && (!isMeta || recordPropertyIsStatic(record, property)))
	}
	switch owner := types.GetUnderlyingType(receiver).(type) {
	case *types.ClassType:
		_, field := owner.GetField(name)
		_, classVar := owner.GetClassVar(name)
		_, property := owner.GetProperty(name)
		return field || classVar || property
	case *types.InterfaceType:
		return owner.GetProperty(name) != nil
	}
	return false
}

// reportAssignmentValueRecovery terminates enclosing checks once the RHS has
// no value, and keeps pointer-target operand errors after the child's errors.
func (a *Analyzer) reportAssignmentValueRecovery(stmt *ast.AssignmentStatement, expected, actual types.Type) bool {
	if actual == nil {
		return false
	}
	valueless := types.GetUnderlyingType(actual).TypeKind() == "VOID"
	pointerOperandMismatch := types.IsPointerType(expected) && !types.IsPointerType(actual) && !a.canAssign(actual, expected)
	// Upstream's simple dynamic-array store converts directly to its element
	// type, bypassing the ordinary assignment's no-result and operand guards.
	if a.isSimpleDynamicArraySlot(stmt) && (valueless || pointerOperandMismatch) {
		mismatch := NewCannotAssignTypesError(stmt.Token.Pos, semanticTypeNameForDiagnostic(actual), semanticTypeNameForDiagnostic(expected))
		mismatch.AfterChildren = true
		a.addStructuredError(mismatch)
		return true
	}
	if valueless {
		a.addStructuredError(&SemanticError{
			Type:          ErrorTypeMismatch,
			Message:       "Syntax Error: Assignment's right-side-argument has no return type",
			Pos:           stmt.Token.Pos,
			Severity:      SeverityError,
			AfterChildren: true,
		})
		return true
	}
	if pointerOperandMismatch {
		operandError := NewIncompatibleOperandsError(stmt.Token.Pos)
		operandError.AfterChildren = true
		a.addStructuredError(operandError)
		mismatch := NewCannotAssignTypesError(stmt.Value.Pos(), semanticTypeNameForDiagnostic(actual), semanticTypeNameForDiagnostic(expected))
		mismatch.AfterChildren = true
		a.addStructuredError(mismatch)
		return true
	}
	return false
}

func (a *Analyzer) isSimpleDynamicArraySlot(stmt *ast.AssignmentStatement) bool {
	if stmt.Operator != token.ASSIGN && stmt.Operator != token.TokenType(0) {
		return false
	}
	index, indexed := stmt.Target.(*ast.IndexExpression)
	if !indexed {
		return false
	}
	baseType := a.semanticInfo.GetResolvedType(index.Left)
	baseType = a.implicitCallTypePreview(index.Left, baseType)
	array, isArray := types.GetUnderlyingType(baseType).(*types.ArrayType)
	return isArray && array.IsDynamic()
}

// analyzeCompoundAssignmentExpression keeps context-free literal/class-operator
// inference while retaining assignment-local arity recovery and scalar calls.
func (a *Analyzer) analyzeCompoundAssignmentExpression(value ast.Expression, target types.Type) types.Type {
	previousCall := a.assignmentCallRecovery
	a.assignmentCallRecovery = nil
	if call, ok := value.(*ast.CallExpression); ok {
		a.assignmentCallRecovery = call
	}
	defer func() { a.assignmentCallRecovery = previousCall }()

	if _, classOperatorContext := types.GetUnderlyingType(target).(*types.ClassType); !classOperatorContext {
		if identifier, ok := value.(*ast.Identifier); ok {
			if result, readAsCall := a.analyzeAssignmentIdentifierCall(identifier, target, true); readAsCall {
				return result
			}
		}
	}
	return a.analyzeExpression(value)
}

// assignmentAcceptsCallableReference preserves broad-target pointer copies, while
// a named routine needs a compatible pointer context to remain a reference.
func (a *Analyzer) assignmentAcceptsCallableReference(actual, expected types.Type, namedRoutine bool) bool {
	return a.canAssign(actual, expected) && (types.IsPointerType(expected) || !namedRoutine)
}
