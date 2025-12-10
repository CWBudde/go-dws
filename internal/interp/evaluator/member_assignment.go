package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
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
// - Class/metaclass assignment via ClassMetaValue interface
// - Nil value auto-initialization
//
// Delegates to adapter (EvalNode):
// - Record property setter dispatch
//
// EvalNode reduction: 6 calls → 1 call (~83% reduction)
// ============================================================================

// evalMemberAssignmentDirect attempts to handle member assignment directly.
//
// Handles directly:
// - Record field assignment via RecordFieldSetter interface
// - Interface unwrapping and routing to underlying object
// - Object field assignment via ObjectFieldSetter interface
// - Object property assignment via WriteProperty callback
// - Class/metaclass assignment via ClassMetaValue interface
//
// Delegates to adapter (EvalNode):
// - Nil value auto-initialization
// - Record property setter dispatch (when record has properties)
func (e *Evaluator) evalMemberAssignmentDirect(
	target *ast.MemberAccessExpression,
	value Value,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	var objVal Value
	var objSetter func(Value) error

	// Try to evaluate as LValue to allow auto-initialization and proper mutation
	if IsVarTarget(target.Object) {
		var err error
		objVal, objSetter, err = e.EvaluateLValue(target.Object, ctx)
		if err != nil {
			return e.newError(stmt, "%s", err.Error())
		}
	} else {
		// Not an LValue (e.g. function call), evaluate as RValue
		objVal = e.Eval(target.Object, ctx)
		if isError(objVal) {
			return objVal
		}
	}

	// Check for exception during evaluation
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	// NATIVE: Nil value handling (auto-initialization)
	if objVal == nil || objVal.Type() == "NIL" {
		// Only attempt auto-initialization if we have a setter (LValue)
		if objSetter != nil {
			// Case: Array element initialization (arr[i].Member := val)
			if indexExpr, ok := target.Object.(*ast.IndexExpression); ok {
				// We need the array type to know what to create
				// Re-evaluate array base to get type info (safe for identifiers)
				arrayVal := e.Eval(indexExpr.Left, ctx)
				if isError(arrayVal) {
					return arrayVal
				}

				if arrVal, ok := arrayVal.(*runtime.ArrayValue); ok {
					if arrVal.ArrayType != nil && arrVal.ArrayType.ElementType != nil {
						// Check if element type is a record
						if recordType, ok := arrVal.ArrayType.ElementType.(*types.RecordType); ok {
							// Create new empty record
							newRecord := runtime.NewRecordValue(recordType, nil)

							// Assign new record to array index using the setter from EvaluateLValue
							if err := objSetter(newRecord); err != nil {
								return e.newError(stmt, "failed to auto-initialize record: %s", err.Error())
							}

							// Update objVal to the new record and proceed with member assignment
							objVal = newRecord
						}
					}
				}
			}
		} else {
			// No setter (rvalue), delegating to adapter for fallback handling
			return e.adapter.EvalNode(stmt)
		}
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

	// NATIVE: Class/metaclass assignment
	objType := objVal.Type()
	if strings.HasPrefix(objType, "CLASS") || objType == "CLASSINFO" {
		if classMeta, ok := objVal.(ClassMetaValue); ok {
			// Check for Class Variable
			if classMeta.HasClassVar(fieldName) {
				if classMeta.SetClassVar(fieldName, value) {
					return value
				}
			}

			// Check for Class Property
			result, ok := classMeta.WriteClassProperty(fieldName, value, func(propInfo any, val Value) Value {
				return e.adapter.EvalClassPropertyWrite(classMeta.GetClassInfo(), propInfo, val, stmt)
			})
			if ok {
				return result
			}
		}
		return e.adapter.EvalNode(stmt)
	}

	// Unknown type or unsupported member assignment
	return e.newError(stmt, "member assignment not supported for type %s", objType)
}
