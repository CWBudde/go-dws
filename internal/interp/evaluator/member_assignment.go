package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Member Assignment Operations
// ============================================================================
//
// Handles member assignment: obj.field := value, TClass.Variable := value
//
// Member assignment is HIGH RISK due to complex dispatch logic:
// - Class variable assignment (static)
// - Class property assignment (static with setter dispatch)
// - Record field assignment (with property setter support)
// - Object field assignment
// - Property setter dispatch (with recursion prevention)
// - Interface unwrapping
// - Auto-initialization of record array elements
//
// Most cases require types from the interp package (ClassInfo, ObjectInstance,
// RecordValue, InterfaceInstance) which cannot be imported here. Therefore,
// complex cases are delegated to the adapter.
//
// Simple cases handled directly:
// - Record field assignment via RecordFieldSetter interface (when no property setter)
//
// Complex cases delegated to adapter:
// - Class variable/property assignment
// - Object field/property assignment
// - Property setter dispatch
// - Interface member assignment
// - Auto-initialization
// ============================================================================

// evalMemberAssignmentDirect attempts to handle member assignment directly.
// Complex cases (class variables, properties, interfaces) are delegated to adapter.
//
// Handles directly:
// - Simple record field assignment (when record implements RecordFieldSetter)
//
// Delegates to adapter:
// - Class variable assignment: TClass.Variable := value
// - Class property assignment: TClass.Property := value
// - Object field assignment: obj.field := value
// - Property setter dispatch: obj.Property := value
// - Interface member assignment
// - Auto-initialization of array elements
func (e *Evaluator) evalMemberAssignmentDirect(
	target *ast.MemberAccessExpression,
	value Value,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	// Check if the target object is a class identifier (static access)
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

	// Handle nil values (might need auto-initialization)
	if objVal == nil || objVal.Type() == "NIL" {
		// Delegate to adapter for potential auto-initialization
		return e.adapter.EvalNode(stmt)
	}

	// Check object type and route accordingly
	objType := objVal.Type()

	// Record field assignment
	// Records might have properties with setters, so we need to be careful
	// Check if this is a record type (by type name pattern or explicit RECORD type)
	if objType == "RECORD" || isRecordTypeName(objType) {
		// Records may have properties with setters that require adapter dispatch
		// For safety, delegate all record assignments to adapter
		// Future optimization: use RecordFieldSetter interface for simple field-only records
		return e.adapter.EvalNode(stmt)
	}

	// Interface member assignment
	if strings.HasPrefix(objType, "INTERFACE") {
		return e.adapter.EvalNode(stmt)
	}

	// Object member assignment (fields and properties)
	if strings.HasPrefix(objType, "OBJECT[") {
		return e.adapter.EvalNode(stmt)
	}

	// Class/metaclass assignment
	if strings.HasPrefix(objType, "CLASS") || objType == "CLASSINFO" {
		return e.adapter.EvalNode(stmt)
	}

	// Unknown type - delegate to adapter
	return e.adapter.EvalNode(stmt)
}

// isRecordTypeName checks if a type name looks like a record type.
// Record type names typically start with 'T' followed by uppercase.
func isRecordTypeName(typeName string) bool {
	if len(typeName) < 2 {
		return false
	}
	// Records in DWScript typically have names like TMyRecord
	// This is a heuristic - the adapter will validate properly
	return typeName[0] == 'T' && typeName[1] >= 'A' && typeName[1] <= 'Z'
}

// evalSimpleRecordFieldAssignment handles direct record field assignment
// when no property setter is involved.
// This is a placeholder for future optimization when RecordValue moves to runtime.
//
// Currently, all record assignments go through the adapter because:
// 1. RecordValue is in interp package (can't access directly)
// 2. Records may have properties with setters that need dispatch
// 3. Property lookup requires RecordType which is in types package
//
// When RecordValue moves to runtime package, this can be implemented to:
// 1. Check if the field exists (no property with same name)
// 2. Set the field directly via RecordFieldSetter interface
// 3. Only delegate to adapter if property setter dispatch is needed
func (e *Evaluator) evalSimpleRecordFieldAssignment(
	record Value,
	fieldName string,
	value Value,
	stmt *ast.AssignmentStatement,
) Value {
	// Normalize field name for case-insensitive lookup
	_ = ident.Normalize(fieldName)

	// Check if record implements RecordFieldSetter
	setter, ok := record.(RecordFieldSetter)
	if !ok {
		return e.newError(stmt, "record does not support field assignment")
	}

	// Set the field
	setter.SetRecordField(fieldName, value)
	return value
}
