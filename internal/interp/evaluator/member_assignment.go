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
// - Record property assignment via executeRecordPropertyWrite
// - Interface unwrapping and routing to underlying object
// - Object field assignment via ObjectFieldSetter interface
// - Object property assignment via WriteProperty callback pattern
// - Class/metaclass assignment via ClassMetaValue interface (class vars & properties)
// - Nil value auto-initialization
//
// Delegates to adapter (EvalNode):
// - Nil auto-initialization fallback (line 105)
//
// EvalNode reduction: 6 calls → 1 call (83% reduction)
// ============================================================================

// evalMemberAssignmentDirect attempts to handle member assignment directly.
//
// Handles directly:
// - Record field assignment via RecordFieldSetter interface
// - Record property assignment via executeRecordPropertyWrite
// - Interface unwrapping and routing to underlying object
// - Object field assignment via ObjectFieldSetter interface
// - Object property assignment via WriteProperty callback
// - Class/metaclass assignment via ClassMetaValue interface (class vars & properties)
// - Nil value auto-initialization
//
// Delegates to adapter (EvalNode):
// - Nil auto-initialization fallback (when no setter available)
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

	// Dereference ReferenceValue (e.g. function name alias to Result)
	// This allows `GetValue.N := 70` to work when GetValue is a ReferenceValue pointing to Result
	if refVal, isRef := objVal.(ReferenceAccessor); isRef {
		deref, err := refVal.Dereference()
		if err != nil {
			return e.newError(stmt, "failed to dereference: %s", err.Error())
		}
		objVal = deref
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
			// No setter (rvalue), cannot assign to a member of a nil value.
			return e.newError(stmt, "cannot assign to member of a nil value")
		}
	}

	fieldName := target.Member.Value

	// NATIVE: Record field/property assignment
	if recInst, ok := objVal.(RecordInstanceValue); ok {
		// Check if record has a property with this name
		if recInst.HasRecordProperty(fieldName) {
			// Get property info and dispatch to record property write handler
			if recVal, ok := objVal.(*runtime.RecordValue); ok {
				propInfo := recVal.LookupProperty(fieldName)
				if propInfo != nil && propInfo.Impl != nil {
					if recordPropInfo, ok := propInfo.Impl.(*types.RecordPropertyInfo); ok {
						return e.executeRecordPropertyWrite(objVal, recordPropInfo, value, stmt, ctx)
					}
				}
			}
			// Fallback if property lookup failed
			return e.newError(stmt, "property '%s' not found in record", fieldName)
		}
		// Simple field assignment - use RecordFieldSetter if available
		if setter, ok := objVal.(RecordFieldSetter); ok {
			setter.SetRecordField(fieldName, value)
			return value
		}
		// Record doesn't support field assignment directly
		return e.newError(stmt, "record does not support field assignment")
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

	// NATIVE: Type cast unwrapping - get wrapped value
	if typeCastVal, ok := objVal.(TypeCastAccessor); ok {
		wrapped := typeCastVal.GetWrappedValue()
		if wrapped == nil {
			return e.newError(stmt, "cannot assign to member of nil cast value")
		}
		// Route to wrapped object
		objVal = wrapped
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
				// Variable exists but SetClassVar failed
				return e.newError(stmt, "failed to set class variable '%s' in class '%s'", fieldName, classMeta.GetClassName())
			}

			// Check for Class Property
			result, ok := classMeta.WriteClassProperty(fieldName, value, func(propInfo any, val Value) Value {
				return e.coreEvaluator.EvalClassPropertyWrite(classMeta.GetClassInfo(), propInfo, val, stmt)
			})
			if ok {
				return result
			}

			// Neither class variable nor class property found
			return e.newError(stmt, "class member '%s' not found in class '%s'", fieldName, classMeta.GetClassName())
		}
		// Not a ClassMetaValue, but has CLASS type
		return e.newError(stmt, "cannot assign to member of class type value")
	}

	// Unknown type or unsupported member assignment
	return e.newError(stmt, "member assignment not supported for type %s", objType)
}
