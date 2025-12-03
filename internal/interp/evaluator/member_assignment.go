package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Member Assignment Operations
// ============================================================================
//
// Handles member assignment: obj.field := value, TClass.Variable := value
//
// NATIVE handling (via evaluator, no adapter):
// - Record field assignment via RecordFieldSetter interface
// - Interface unwrapping and routing to underlying object
// - Object field assignment via ObjectFieldSetter interface
// - Object property assignment via WriteProperty callback pattern
//
// ADAPTER handling (via EvalNode):
// - Static class access: TClass.Variable (requires ClassInfo lookup)
// - Nil value auto-initialization
// - Record property setter dispatch
// - Class/metaclass assignment
//
// EvalNode reduction: 6 calls → 3-4 calls (~50% reduction)
// ============================================================================

// evalMemberAssignmentDirect attempts to handle member assignment directly.
//
// Handles directly:
// - Record field assignment via RecordFieldSetter interface
// - Interface unwrapping and routing to underlying object
// - Object field assignment via ObjectFieldSetter interface
// - Object property assignment via WriteProperty callback
//
// Delegates to adapter (EvalNode):
// - Static class access: TClass.Variable := value (needs ClassInfo lookup)
// - Nil value auto-initialization
// - Record property setter dispatch (when record has properties)
// - Class/metaclass assignment
func (e *Evaluator) evalMemberAssignmentDirect(
	target *ast.MemberAccessExpression,
	value Value,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	// KEEP EVALNODE: Static class identifier access (TClass.Variable)
	// This requires ClassInfo lookup which is in interp package
	if _, ok := target.Object.(*ast.Identifier); ok {
		// Could be TClass.Variable or TClass.Property
		// Delegate to adapter for class info lookup
		return e.adapter.EvalNode(stmt)
	}

	// Evaluate the object expression
	objVal := e.Eval(target.Object, ctx)
	if isError(objVal) {
		return objVal
	}

	// Check for exception during evaluation
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	// KEEP EVALNODE: Nil value handling (auto-initialization)
	if objVal == nil || objVal.Type() == "NIL" {
		// Delegate to adapter for potential auto-initialization
		return e.adapter.EvalNode(stmt)
	}

	fieldName := target.Member.Value

	// NATIVE: Record field assignment via RecordFieldSetter
	if recInst, ok := objVal.(RecordInstanceValue); ok {
		// Check if record has a property with this name (needs setter dispatch)
		if recInst.HasRecordProperty(fieldName) {
			// Record property setter needs adapter dispatch
			return e.adapter.EvalNode(stmt)
		}
		// Simple field assignment - use RecordFieldSetter if available
		if setter, ok := objVal.(RecordFieldSetter); ok {
			setter.SetRecordField(fieldName, value)
			return value
		}
		// Record doesn't support field assignment directly - delegate
		return e.adapter.EvalNode(stmt)
	}

	// NATIVE: Interface unwrapping - get underlying object
	if intfInst, ok := objVal.(InterfaceInstanceValue); ok {
		underlying := intfInst.GetUnderlyingObjectValue()
		if underlying == nil {
			return e.newError(stmt, "cannot assign to member of nil interface")
		}
		// Route to underlying object
		objVal = underlying
	}

	// NATIVE: Object field/property assignment
	if objValIface, ok := objVal.(ObjectValue); ok {
		// Check if this is a property (has priority over fields)
		if objValIface.HasProperty(fieldName) {
			// Property assignment via callback pattern
			return objValIface.WriteProperty(fieldName, value, func(propInfo any, val Value) Value {
				return e.executePropertyWrite(objVal, propInfo, val, stmt, ctx)
			})
		}

		// Direct field assignment via ObjectFieldSetter
		if setter, ok := objVal.(ObjectFieldSetter); ok {
			setter.SetField(fieldName, value)
			return value
		}

		return e.newError(stmt, "object does not support field assignment")
	}

	// KEEP EVALNODE: Class/metaclass assignment
	objType := objVal.Type()
	if strings.HasPrefix(objType, "CLASS") || objType == "CLASSINFO" {
		return e.adapter.EvalNode(stmt)
	}

	// KEEP EVALNODE: Unknown type fallback
	return e.adapter.EvalNode(stmt)
}
