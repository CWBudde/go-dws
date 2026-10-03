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

	if identifier, ok := value.(*ast.Identifier); ok && expected != nil {
		if result, readAsCall := a.analyzeAssignmentIdentifierCall(identifier, expected, readRejectedReference); readAsCall {
			return result
		}
	}
	return a.analyzeExpressionWithExpectedType(value, expected)
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

	var pointer *types.FunctionPointerType
	var actual types.Type
	required := 0
	namedRoutine := false
	switch declared := types.GetUnderlyingType(symbol.Type).(type) {
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
		return nil, false
	}
	if a.assignmentAcceptsCallableReference(actual, expected, namedRoutine) {
		return nil, false
	}
	result := functionPointerCallResult(pointer)
	// Typed pointer initializers select a factory call only when its returned
	// pointer fits; preserve existing rejected-reference initializer diagnostics.
	if !readRejectedReference && (!types.IsPointerType(result) || !a.canAssign(result, expected)) {
		return nil, false
	}
	// Analyze once to retain symbol usage, casing and deprecation diagnostics.
	a.analyzeIdentifier(identifier)
	if required > 0 {
		a.addMoreArgumentsExpected(identifier.Token.Pos)
	}
	a.semanticInfo.SetImplicitCall(identifier)
	a.semanticInfo.SetResolvedType(identifier, result)
	return result, true
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
