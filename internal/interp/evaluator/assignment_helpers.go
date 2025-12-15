package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Assignment Helpers
// ============================================================================

// ReferenceValueAccessor allows dereferencing and assignment to var parameters.
type ReferenceValueAccessor interface {
	Dereference() (Value, error)
	Assign(value Value) error
}

// SubrangeValueAccessor validates and assigns subrange values.
type SubrangeValueAccessor interface {
	ValidateAndSet(intValue int) error
	GetValue() int
	GetTypeName() string
}

// cloneIfCopyable returns a defensive copy for static arrays (value semantics).
// Dynamic arrays keep reference semantics.
func cloneIfCopyable(val Value) Value {
	if val == nil {
		return nil
	}

	// Dynamic arrays keep reference semantics
	if arr, ok := val.(*runtime.ArrayValue); ok {
		if arr.ArrayType == nil || arr.ArrayType.IsDynamic() {
			return val
		}
	}

	// Clone copyable values (static arrays)
	if copyable, ok := val.(runtime.CopyableValue); ok {
		if copied := copyable.Copy(); copied != nil {
			if v, ok := copied.(Value); ok {
				return v
			}
		}
	}

	return val
}

// evalSimpleAssignmentDirect handles simple variable assignment: x := value
// Handles regular variables, var parameters, subranges, interface/object ref counting.
func (e *Evaluator) evalSimpleAssignmentDirect(
	target *ast.Identifier,
	value Value,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	targetName := target.Value

	existingValRaw, exists := ctx.Env().Get(targetName)
	if !exists {
		return e.assignToImplicitTarget(target, targetName, value, ctx)
	}

	existingVal, ok := existingValRaw.(Value)
	if !ok {
		return e.newError(target, "variable '%s' has invalid type (not a Value)", targetName)
	}

	// Dispatch to specialized handlers for special value types
	if refVal, isRef := existingVal.(ReferenceValueAccessor); isRef {
		return e.evalReferenceAssignment(refVal, value, target, stmt, ctx)
	}
	if existingVal.Type() == "EXTERNAL_VAR" {
		return e.errorForExternalVar(existingVal, target)
	}
	if subrangeVal, isSubrange := existingVal.(SubrangeValueAccessor); isSubrange {
		return e.evalSubrangeAssignment(subrangeVal, value, target)
	}
	if ifaceInst, isIface := existingVal.(*runtime.InterfaceInstance); isIface {
		return e.assignToInterfaceVar(ifaceInst, value, targetName, ctx)
	}
	if _, isSrcIface := value.(*runtime.InterfaceInstance); isSrcIface {
		// Assigning interface to non-interface variable
		e.SetVar(ctx, targetName, value)
		return value
	}
	if objInst, isObj := existingVal.(*runtime.ObjectInstance); isObj {
		return e.assignToObjectVar(objInst, value, targetName, ctx)
	}

	// Regular variable assignment with type conversion and cloning
	value = e.prepareValueForAssignment(existingVal, value, stmt, ctx)

	if e.SetVar(ctx, targetName, value) {
		return value
	}
	return e.newError(target, "undefined variable: %s", targetName)
}

// assignToImplicitTarget handles assignment when the target is not in the environment.
// Checks Self context (fields/properties/class vars) and __CurrentClass__ context.
func (e *Evaluator) assignToImplicitTarget(
	target *ast.Identifier,
	targetName string,
	value Value,
	ctx *ExecutionContext,
) Value {
	// Check Self context for fields, class vars, properties
	if selfRaw, ok := ctx.Env().Get("Self"); ok {
		if selfVal, ok := selfRaw.(Value); ok && selfVal.Type() == "OBJECT" {
			if result := e.assignToSelfMember(target, targetName, value, selfVal, ctx); result != nil {
				return result
			}
		}
	}

	// Check __CurrentClass__ for class variables in class methods
	if currentClassRaw, ok := ctx.Env().Get("__CurrentClass__"); ok {
		if classInfoVal, ok := currentClassRaw.(Value); ok && classInfoVal.Type() == "CLASSINFO" {
			if result := e.assignToCurrentClassVar(target, targetName, value, classInfoVal); result != nil {
				return result
			}
		}
	}

	return e.newError(target, "undefined variable '%s'", targetName)
}

// assignToSelfMember tries to assign to a field, class var, or property of Self.
// Returns nil if targetName is not a member of Self.
func (e *Evaluator) assignToSelfMember(
	target *ast.Identifier,
	targetName string,
	value Value,
	selfVal Value,
	ctx *ExecutionContext,
) Value {
	objVal, ok := selfVal.(ObjectValue)
	if !ok {
		return nil
	}

	// Try instance field
	if objVal.GetField(targetName) != nil {
		if objInst, ok := selfVal.(*runtime.ObjectInstance); ok {
			objInst.SetField(targetName, value)
			return value
		}
		return e.newError(target, "cannot assign to field '%s': invalid object type", targetName)
	}

	// Try class variable
	if _, found := objVal.GetClassVar(targetName); found {
		return e.assignToClassVarViaSelf(target, targetName, value, selfVal)
	}

	// Try property
	if objVal.HasProperty(targetName) {
		return objVal.WriteProperty(targetName, value, func(propInfo any, val Value) Value {
			return e.executePropertyWrite(selfVal, propInfo, val, target, ctx)
		})
	}

	return nil
}

// assignToClassVarViaSelf assigns to a class variable accessed through Self.
func (e *Evaluator) assignToClassVarViaSelf(
	target *ast.Identifier,
	targetName string,
	value Value,
	selfVal Value,
) Value {
	objInst, ok := selfVal.(*runtime.ObjectInstance)
	if !ok || objInst.Class == nil {
		return e.newError(target, "cannot assign to class variable '%s': class does not support SetClassVar", targetName)
	}

	className := objInst.Class.GetName()
	if className == "" {
		return e.newError(target, "cannot assign to class variable '%s': class does not support SetClassVar", targetName)
	}

	classMetaVal := e.oopEngine.LookupClassByName(className)
	if classMetaVal == nil {
		return e.newError(target, "cannot assign to class variable '%s': class does not support SetClassVar", targetName)
	}

	if classMetaVal.SetClassVar(targetName, value) {
		return value
	}
	return e.newError(target, "failed to set class variable '%s'", targetName)
}

// assignToCurrentClassVar assigns to a class variable in __CurrentClass__ context.
// Returns nil if targetName is not a class variable.
func (e *Evaluator) assignToCurrentClassVar(
	target *ast.Identifier,
	targetName string,
	value Value,
	classInfoVal Value,
) Value {
	classMetaVal, ok := classInfoVal.(ClassMetaValue)
	if !ok {
		return nil
	}

	if _, found := classMetaVal.GetClassVar(targetName); !found {
		return nil
	}

	if classMetaVal.SetClassVar(targetName, value) {
		return value
	}
	return e.newError(target, "failed to set class variable '%s'", targetName)
}

// errorForExternalVar returns an error for unsupported external variable assignment.
func (e *Evaluator) errorForExternalVar(existingVal Value, target *ast.Identifier) Value {
	if extVar, ok := existingVal.(ExternalVarAccessor); ok {
		return e.newError(target, "Unsupported external variable assignment: %s", extVar.ExternalVarName())
	}
	return e.newError(target, "Unsupported external variable assignment")
}

// assignToInterfaceVar handles assignment to an interface variable with ref counting.
func (e *Evaluator) assignToInterfaceVar(
	ifaceInst *runtime.InterfaceInstance,
	value Value,
	targetName string,
	ctx *ExecutionContext,
) Value {
	refMgr := ctx.RefCountManager()
	refMgr.ReleaseInterface(ifaceInst)

	// Wrap new value in interface based on its type
	switch v := value.(type) {
	case *runtime.ObjectInstance:
		value = refMgr.WrapInInterface(ifaceInst.Interface, v)
	case *runtime.InterfaceInstance:
		// Reuse source interface WITHOUT incrementing refcount
		value = &runtime.InterfaceInstance{
			Interface: ifaceInst.Interface,
			Object:    v.Object,
		}
	case *runtime.NilValue:
		value = &runtime.InterfaceInstance{
			Interface: ifaceInst.Interface,
			Object:    nil,
		}
	}

	e.SetVar(ctx, targetName, value)
	return value
}

// assignToObjectVar handles assignment to an object variable with ref counting.
func (e *Evaluator) assignToObjectVar(
	objInst *runtime.ObjectInstance,
	value Value,
	targetName string,
	ctx *ExecutionContext,
) Value {
	refMgr := ctx.RefCountManager()

	switch v := value.(type) {
	case *runtime.NilValue:
		refMgr.ReleaseObject(objInst)
	case *runtime.ObjectInstance:
		if objInst != v {
			refMgr.ReleaseObject(objInst)
			refMgr.IncrementRef(v)
		}
	default:
		refMgr.ReleaseObject(objInst)
	}

	e.SetVar(ctx, targetName, value)
	return value
}

// prepareValueForAssignment applies type conversion, variant boxing, cloning, and ref counting.
func (e *Evaluator) prepareValueForAssignment(
	existingVal Value,
	value Value,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	if value == nil {
		return value
	}

	targetType := existingVal.Type()
	sourceType := value.Type()

	// Implicit type conversion
	if targetType != sourceType {
		if converted, ok := e.TryImplicitConversion(value, targetType, ctx); ok {
			value = converted
		}
	}

	// Box value if target is Variant
	if targetType == "VARIANT" && sourceType != "VARIANT" {
		value = runtime.BoxVariant(value)
	}

	// Clone copyable values (static arrays), except indexed expressions
	shouldClone := stmt == nil
	if stmt != nil {
		_, isIndexExpr := stmt.Value.(*ast.IndexExpression)
		shouldClone = !isIndexExpr
	}
	if shouldClone {
		value = cloneIfCopyable(value)
	}

	// Increment ref count for new objects (interfaces handle this separately)
	if newObj, isNewObj := value.(*runtime.ObjectInstance); isNewObj {
		if _, isIface := existingVal.(*runtime.InterfaceInstance); !isIface {
			ctx.RefCountManager().IncrementRef(newObj)
		}
	}

	// Increment ref count for method pointer's SelfObject
	if value.Type() == "METHOD_POINTER" {
		if funcPtr, ok := value.(*runtime.FunctionPointerValue); ok && funcPtr.SelfObject != nil {
			ctx.RefCountManager().IncrementRef(funcPtr.SelfObject)
		}
	}

	return value
}

// evalReferenceAssignment handles assignment through a var parameter.
func (e *Evaluator) evalReferenceAssignment(
	refVal ReferenceValueAccessor,
	value Value,
	target *ast.Identifier,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	currentVal, err := refVal.Dereference()
	if err != nil {
		return e.newError(target, "%s", err.Error())
	}

	// Interface/object var parameter - ref counting
	if currentVal.Type() == "INTERFACE" || currentVal.Type() == "OBJECT" {
		refMgr := ctx.RefCountManager()

		// Release old reference
		if oldIntf, isOldIntf := currentVal.(*runtime.InterfaceInstance); isOldIntf {
			refMgr.ReleaseInterface(oldIntf)
		} else if oldObj, isOldObj := currentVal.(*runtime.ObjectInstance); isOldObj {
			refMgr.ReleaseObject(oldObj)
		}

		// Increment new reference
		if value != nil {
			refMgr.IncrementRef(value)
		}

		if err := refVal.Assign(value); err != nil {
			return e.newError(target, "%s", err.Error())
		}

		return value
	}

	// Implicit type conversion
	targetType := currentVal.Type()
	sourceType := value.Type()
	if targetType != sourceType {
		if converted, ok := e.TryImplicitConversion(value, targetType, ctx); ok {
			value = converted
		}
	}

	// Box value if target is Variant
	if targetType == "VARIANT" && sourceType != "VARIANT" {
		value = runtime.BoxVariant(value)
	}

	// Clone copyable values
	value = cloneIfCopyable(value)

	if err := refVal.Assign(value); err != nil {
		return e.newError(target, "%s", err.Error())
	}

	return value
}

// evalSubrangeAssignment validates and assigns to a subrange variable.
func (e *Evaluator) evalSubrangeAssignment(
	subrangeVal SubrangeValueAccessor,
	value Value,
	target *ast.Identifier,
) Value {
	// Extract integer value from source
	var intValue int
	switch v := value.(type) {
	case *runtime.IntegerValue:
		intValue = int(v.Value)
	case SubrangeValueAccessor:
		intValue = v.GetValue()
	default:
		return e.newError(target, "cannot assign %s to subrange type %s", value.Type(), subrangeVal.GetTypeName())
	}

	// Validate the value is in range
	if err := subrangeVal.ValidateAndSet(intValue); err != nil {
		return e.newError(target, "%s", err.Error())
	}

	// Return the subrange value (modified in place)
	return value
}

// evalCompoundIdentifierAssignment handles compound assignment (+=, -=, *=, /=).
// Read current value, evaluate RHS, apply operation, write result back.
func (e *Evaluator) evalCompoundIdentifierAssignment(
	target *ast.Identifier,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	targetName := target.Value

	currentValRaw, exists := ctx.Env().Get(targetName)
	if !exists {
		return e.compoundAssignToImplicitTarget(target, targetName, stmt, ctx)
	}

	currentVal, ok := currentValRaw.(Value)
	if !ok {
		return e.newError(target, "variable '%s' has invalid type (not a Value)", targetName)
	}

	// Var parameter - read-modify-write through reference
	if refVal, isRef := currentVal.(ReferenceValueAccessor); isRef {
		return e.compoundAssignToReference(refVal, target, stmt, ctx)
	}

	// Regular variable - read-modify-write
	rightVal := e.Eval(stmt.Value, ctx)
	if isError(rightVal) {
		return rightVal
	}
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	result := e.applyCompoundOperation(stmt.Operator, currentVal, rightVal, stmt)
	if isError(result) {
		return result
	}

	if e.SetVar(ctx, targetName, result) {
		return result
	}
	return e.newError(target, "undefined variable: %s", targetName)
}

// compoundAssignToImplicitTarget handles compound assignment when target is not in environment.
func (e *Evaluator) compoundAssignToImplicitTarget(
	target *ast.Identifier,
	targetName string,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	// Check Self context
	if selfRaw, ok := ctx.Env().Get("Self"); ok {
		if selfVal, ok := selfRaw.(Value); ok && selfVal.Type() == "OBJECT" {
			if result := e.compoundAssignToSelfMember(target, targetName, stmt, selfVal, ctx); result != nil {
				return result
			}
		}
	}

	// Check __CurrentClass__ context
	if currentClassRaw, ok := ctx.Env().Get("__CurrentClass__"); ok {
		if classInfoVal, ok := currentClassRaw.(Value); ok && classInfoVal.Type() == "CLASSINFO" {
			if result := e.compoundAssignToCurrentClassVar(target, targetName, stmt, classInfoVal); result != nil {
				return result
			}
		}
	}

	return e.newError(target, "undefined variable '%s'", targetName)
}

// compoundAssignToSelfMember handles compound assignment to Self's field, class var, or property.
func (e *Evaluator) compoundAssignToSelfMember(
	target *ast.Identifier,
	targetName string,
	stmt *ast.AssignmentStatement,
	selfVal Value,
	ctx *ExecutionContext,
) Value {
	objVal, ok := selfVal.(ObjectValue)
	if !ok {
		return nil
	}

	// Try instance field
	if fieldValue := objVal.GetField(targetName); fieldValue != nil {
		return e.compoundAssignToField(target, targetName, stmt, fieldValue, selfVal, ctx)
	}

	// Try class variable
	if classVarValue, found := objVal.GetClassVar(targetName); found {
		return e.compoundAssignToClassVarViaSelf(target, targetName, stmt, classVarValue, selfVal, ctx)
	}

	// Try property
	if objVal.HasProperty(targetName) {
		return e.compoundAssignToProperty(target, targetName, stmt, objVal, selfVal, ctx)
	}

	return nil
}

// compoundAssignToField handles compound assignment to an object field.
func (e *Evaluator) compoundAssignToField(
	target *ast.Identifier,
	targetName string,
	stmt *ast.AssignmentStatement,
	fieldValue Value,
	selfVal Value,
	ctx *ExecutionContext,
) Value {
	rightVal := e.Eval(stmt.Value, ctx)
	if isError(rightVal) {
		return rightVal
	}

	result := e.applyCompoundOperation(stmt.Operator, fieldValue, rightVal, target)
	if isError(result) {
		return result
	}

	if objInst, ok := selfVal.(*runtime.ObjectInstance); ok {
		objInst.SetField(targetName, result)
		return result
	}
	return e.newError(target, "cannot assign to field '%s': invalid object type", targetName)
}

// compoundAssignToClassVarViaSelf handles compound assignment to a class variable via Self.
func (e *Evaluator) compoundAssignToClassVarViaSelf(
	target *ast.Identifier,
	targetName string,
	stmt *ast.AssignmentStatement,
	classVarValue Value,
	selfVal Value,
	ctx *ExecutionContext,
) Value {
	rightVal := e.Eval(stmt.Value, ctx)
	if isError(rightVal) {
		return rightVal
	}

	result := e.applyCompoundOperation(stmt.Operator, classVarValue, rightVal, target)
	if isError(result) {
		return result
	}

	objInst, ok := selfVal.(*runtime.ObjectInstance)
	if !ok {
		return e.newError(target, "cannot assign to class variable '%s': class does not support SetClassVar", targetName)
	}

	classMetaVal, ok := objInst.Class.(ClassMetaValue)
	if !ok {
		return e.newError(target, "cannot assign to class variable '%s': class does not support SetClassVar", targetName)
	}

	if classMetaVal.SetClassVar(targetName, result) {
		return result
	}
	return e.newError(target, "failed to set class variable '%s'", targetName)
}

// compoundAssignToProperty handles compound assignment to a property (read-modify-write).
func (e *Evaluator) compoundAssignToProperty(
	target *ast.Identifier,
	targetName string,
	stmt *ast.AssignmentStatement,
	objVal ObjectValue,
	selfVal Value,
	ctx *ExecutionContext,
) Value {
	currentPropValue := objVal.ReadProperty(targetName, func(propInfo any) Value {
		return e.executePropertyRead(selfVal, propInfo, target, ctx)
	})
	if isError(currentPropValue) {
		return currentPropValue
	}

	rightVal := e.Eval(stmt.Value, ctx)
	if isError(rightVal) {
		return rightVal
	}

	result := e.applyCompoundOperation(stmt.Operator, currentPropValue, rightVal, target)
	if isError(result) {
		return result
	}

	return objVal.WriteProperty(targetName, result, func(propInfo any, val Value) Value {
		return e.executePropertyWrite(selfVal, propInfo, val, target, ctx)
	})
}

// compoundAssignToCurrentClassVar handles compound assignment in __CurrentClass__ context.
func (e *Evaluator) compoundAssignToCurrentClassVar(
	target *ast.Identifier,
	targetName string,
	stmt *ast.AssignmentStatement,
	classInfoVal Value,
) Value {
	classMetaVal, ok := classInfoVal.(ClassMetaValue)
	if !ok {
		return nil
	}

	classVarValue, found := classMetaVal.GetClassVar(targetName)
	if !found {
		return nil
	}

	rightVal := e.Eval(stmt.Value, nil) // Note: ctx not needed for class var eval
	if isError(rightVal) {
		return rightVal
	}

	result := e.applyCompoundOperation(stmt.Operator, classVarValue, rightVal, target)
	if isError(result) {
		return result
	}

	if classMetaVal.SetClassVar(targetName, result) {
		return result
	}
	return e.newError(target, "failed to set class variable '%s'", targetName)
}

// compoundAssignToReference handles compound assignment through a var parameter.
func (e *Evaluator) compoundAssignToReference(
	refVal ReferenceValueAccessor,
	target *ast.Identifier,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	derefVal, err := refVal.Dereference()
	if err != nil {
		return e.newError(target, "%s", err.Error())
	}

	rightVal := e.Eval(stmt.Value, ctx)
	if isError(rightVal) {
		return rightVal
	}
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	result := e.applyCompoundOperation(stmt.Operator, derefVal, rightVal, stmt)
	if isError(result) {
		return result
	}

	if err := refVal.Assign(result); err != nil {
		return e.newError(target, "%s", err.Error())
	}
	return result
}

// ============================================================================
// Context Inference Helpers
// ============================================================================
//
// These enable array and record literals to infer their types from the
// target variable during assignment.
// ============================================================================

// getArrayTypeFromTarget extracts ArrayType from target variable.
// This enables context inference for array literals during assignment:
// var arr: array of Integer; arr := [1, 2, 3]; // Literal adopts arr's type
//
// Returns nil if:
// - Target variable doesn't exist
// - Target variable is not an ArrayValue
// - Target variable has no type information
func (e *Evaluator) getArrayTypeFromTarget(target *ast.Identifier, ctx *ExecutionContext) *types.ArrayType {
	existingVal, _ := ctx.Env().Get(target.Value)

	// Direct ArrayValue check
	if arrVal, ok := existingVal.(*runtime.ArrayValue); ok {
		return arrVal.ArrayType
	}

	// Try evaluator lookup (handles aliases, special contexts, wrappers)
	if arrType := e.resolveArrayTypeFromIdentifier(target, ctx); arrType != nil {
		return arrType
	}

	// Try type string from value
	if arrType := e.resolveArrayTypeFromTypeStringer(existingVal, ctx); arrType != nil {
		return arrType
	}

	// Fall back to semantic type information
	return e.resolveArrayTypeFromSemanticInfo(target, ctx)
}

// resolveArrayTypeFromIdentifier resolves array type by evaluating the identifier.
func (e *Evaluator) resolveArrayTypeFromIdentifier(target *ast.Identifier, ctx *ExecutionContext) *types.ArrayType {
	if target == nil {
		return nil
	}

	resolved := e.VisitIdentifier(target, ctx)
	if arrVal, ok := resolved.(*runtime.ArrayValue); ok {
		return arrVal.ArrayType
	}

	if typeStringer, ok := resolved.(interface{ ArrayTypeString() string }); ok {
		return e.resolveArrayTypeFromTypeName(typeStringer.ArrayTypeString(), ctx)
	}
	return nil
}

// resolveArrayTypeFromTypeStringer extracts array type from a value with ArrayTypeString method.
func (e *Evaluator) resolveArrayTypeFromTypeStringer(val any, ctx *ExecutionContext) *types.ArrayType {
	if typeStringer, ok := val.(interface{ ArrayTypeString() string }); ok {
		return e.resolveArrayTypeFromTypeName(typeStringer.ArrayTypeString(), ctx)
	}
	return nil
}

// resolveArrayTypeFromSemanticInfo gets array type from semantic analysis info.
func (e *Evaluator) resolveArrayTypeFromSemanticInfo(target *ast.Identifier, ctx *ExecutionContext) *types.ArrayType {
	if e.semanticInfo == nil {
		return nil
	}

	typeAnnot := e.semanticInfo.GetType(target)
	if typeAnnot == nil || typeAnnot.Name == "" {
		return nil
	}

	return e.resolveArrayTypeFromTypeName(typeAnnot.Name, ctx)
}

// resolveArrayTypeFromTypeName resolves a type name to ArrayType.
func (e *Evaluator) resolveArrayTypeFromTypeName(typeName string, ctx *ExecutionContext) *types.ArrayType {
	if typeName == "" || typeName == "array" {
		return nil
	}

	resolved, err := e.ResolveTypeWithContext(typeName, ctx)
	if err != nil {
		return nil
	}

	if arrType, ok := resolved.(*types.ArrayType); ok {
		return arrType
	}

	if underlying := types.GetUnderlyingType(resolved); underlying != nil {
		if arrType, ok := underlying.(*types.ArrayType); ok {
			return arrType
		}
	}
	return nil
}

// getSetTypeFromTarget extracts SetType from the target variable.
// This enables context inference for set literals during assignment:
// var s: set of TEnum; s := []; // Literal adopts s's type
//
// Returns nil if the set type cannot be determined.
func (e *Evaluator) getSetTypeFromTarget(target *ast.Identifier, ctx *ExecutionContext) *types.SetType {
	existingVal, exists := ctx.Env().Get(target.Value)
	if exists {
		if setVal, ok := existingVal.(*runtime.SetValue); ok {
			return setVal.SetType
		}
	}

	if target != nil {
		resolved := e.VisitIdentifier(target, ctx)
		if setVal, ok := resolved.(*runtime.SetValue); ok {
			return setVal.SetType
		}
	}

	if e.semanticInfo != nil {
		if typeAnnot := e.semanticInfo.GetType(target); typeAnnot != nil && typeAnnot.Name != "" {
			if resolved, err := e.ResolveTypeWithContext(typeAnnot.Name, ctx); err == nil {
				if setType, ok := types.GetUnderlyingType(resolved).(*types.SetType); ok {
					return setType
				}
			}
		}
	}

	return nil
}

// getRecordTypeNameFromTarget extracts record type name from target variable.
// This enables context inference for anonymous record literals during assignment:
// var p: TPoint; p := (x: 1, y: 2); // Literal adopts TPoint type
//
// RecordValue.Type() returns the record type name (e.g., "TPoint") or "RECORD"
// for anonymous records. For context inference, we need the actual type name.
//
// Returns empty string if:
// - Target variable doesn't exist
// - Target variable is not a record
// - Target variable is an anonymous record (Type() returns "RECORD")
// - Type name is not a registered record type
func (e *Evaluator) getRecordTypeNameFromTarget(target *ast.Identifier, ctx *ExecutionContext) string {
	existingVal, exists := ctx.Env().Get(target.Value)
	if !exists || existingVal == nil {
		return ""
	}

	// RecordValue.Type() returns record type name or "RECORD" for anonymous
	typeName := existingVal.Type()
	if typeName != "" && typeName != "RECORD" && e.typeSystem.HasRecord(typeName) {
		return typeName
	}
	return ""
}
