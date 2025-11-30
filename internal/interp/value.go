// Package interp provides the interpreter and runtime for DWScript.
package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Value Interface and Type Aliases
// ============================================================================
//
// Phase 3.2 Refactoring: Primitive value types have been moved to the
// internal/interp/runtime package. For backward compatibility, we provide
// type aliases here. Code can continue using interp.IntegerValue, etc.
//
// Eventually, all code should be updated to import runtime directly.
// ============================================================================

// Value represents a runtime value in the DWScript interpreter.
// All runtime values must implement this interface.
//
// This is aliased from runtime.Value for backward compatibility.
type Value = runtime.Value

// Primitive value type aliases (Phase 3.2.2: moved to runtime/)
type (
	// IntegerValue represents an integer value in DWScript.
	// Moved to runtime.IntegerValue in Phase 3.2.
	IntegerValue = runtime.IntegerValue

	// FloatValue represents a floating-point value in DWScript.
	// Moved to runtime.FloatValue in Phase 3.2.
	FloatValue = runtime.FloatValue

	// StringValue represents a string value in DWScript.
	// Moved to runtime.StringValue in Phase 3.2.
	StringValue = runtime.StringValue

	// BooleanValue represents a boolean value in DWScript.
	// Moved to runtime.BooleanValue in Phase 3.2.
	BooleanValue = runtime.BooleanValue

	// NilValue represents a nil/null value in DWScript.
	// Moved to runtime.NilValue in Phase 3.2.
	NilValue = runtime.NilValue

	// NullValue represents the special Variant Null value in DWScript.
	// Task 9.4.1: Null is a variant-specific value that represents an explicit null state.
	// Moved to runtime.NullValue in Phase 3.2.
	NullValue = runtime.NullValue

	// UnassignedValue represents the special Variant Unassigned value in DWScript.
	// Task 9.4.1: Unassigned is the default state of an uninitialized variant.
	// Moved to runtime.UnassignedValue in Phase 3.2.
	UnassignedValue = runtime.UnassignedValue

	// EnumValue represents an enumerated value in DWScript.
	// Moved to runtime.EnumValue in Phase 3.5.4.
	EnumValue = runtime.EnumValue

	// TypeMetaValue represents a type reference in DWScript.
	// Moved to runtime.TypeMetaValue in Phase 3.5.4.
	TypeMetaValue = runtime.TypeMetaValue

	// SetValue represents a set value in DWScript.
	// Moved to runtime.SetValue in Phase 3.5.4.
	SetValue = runtime.SetValue

	// IntRange represents an integer range for lazy set storage.
	// Moved to runtime.IntRange in Phase 3.5.4.
	IntRange = runtime.IntRange

	// ArrayValue represents an array value in DWScript.
	// Moved to runtime.ArrayValue in Phase 3.5.4.
	ArrayValue = runtime.ArrayValue

	// VariantValue represents a Variant value in DWScript.
	// Task 3.5.139: Moved to runtime.VariantValue for evaluator access.
	VariantValue = runtime.VariantValue
)

// ============================================================================
// Value interface re-exports for type operations (Phase 3.2.1)
// ============================================================================

// NumericValue represents values that can be used in numeric operations.
type NumericValue = runtime.NumericValue

// ComparableValue represents values that can be compared for equality.
type ComparableValue = runtime.ComparableValue

// OrderableValue represents values that can be ordered.
type OrderableValue = runtime.OrderableValue

// CopyableValue represents values that can be copied.
type CopyableValue = runtime.CopyableValue

// ReferenceType represents reference-type values (not to be confused with ReferenceValue struct).
type ReferenceType = runtime.ReferenceType

// IndexableValue represents values that can be indexed.
type IndexableValue = runtime.IndexableValue

// CallableValue represents values that can be called as functions.
type CallableValue = runtime.CallableValue

// ConvertibleValue represents values that support explicit type conversion.
type ConvertibleValue = runtime.ConvertibleValue

// IterableValue represents values that can be iterated over.
type IterableValue = runtime.IterableValue

// Iterator provides iteration over collection values.
type Iterator = runtime.Iterator

// ============================================================================
// Complex Value Types (remaining in interp/ for now)
// ============================================================================
// These will be moved to runtime/ in future phases of the refactoring.
// ============================================================================

// RTTITypeInfoValue represents runtime type information in DWScript.
// Task 9.25: TypeOf(value) returns this value type for RTTI operations.
// This value serves as a unique identifier for a type that can be compared
// and used to look up type metadata.
//
// Examples:
//   - TypeOf(obj) returns RTTITypeInfoValue for obj's runtime type
//   - TypeOf(TMyClass) returns RTTITypeInfoValue for the class type
//   - TypeOf(classRef) returns RTTITypeInfoValue for the class reference's type
type RTTITypeInfoValue struct {
	TypeInfo types.Type
	TypeName string
	TypeID   int
}

// Type returns "RTTI_TYPEINFO".
func (r *RTTITypeInfoValue) Type() string {
	return "RTTI_TYPEINFO"
}

// String returns the type name.
func (r *RTTITypeInfoValue) String() string {
	return r.TypeName
}

// RecordValue represents a record value in DWScript.
// Task 3.5.42: Migrated to use RecordMetadata instead of AST-dependent method map.
// Task 3.5.128a: Removed deprecated Methods field - now uses only Metadata.Methods.
// Task 3.5.128b: Moved to runtime package; type alias for backward compatibility.
type RecordValue = runtime.RecordValue

// GetRecordMethod retrieves a method declaration by name from a RecordValue.
// Task 9.7: Helper for record method invocation.
// Task 9.7.3: Case-insensitive method lookup.
// Task 3.5.42: Updated to use RecordMetadata with fallback to legacy Methods field.
// Task 3.5.128a: Removed legacy fallback - now uses only Metadata.Methods.
// Task 3.5.128b: Changed from method to free function due to type alias.
func GetRecordMethod(r *RecordValue, name string) *ast.FunctionDecl {
	// Use metadata for method lookup
	if r.Metadata == nil {
		return nil
	}

	normalizedName := ident.Normalize(name)
	methodMeta, ok := r.Metadata.Methods[normalizedName]
	if !ok {
		return nil
	}

	// Return the AST body from metadata
	// Note: During migration, MethodMetadata.Body contains the AST node
	if methodMeta.Body == nil {
		return nil
	}

	// Reconstruct FunctionDecl from metadata for compatibility
	// This is temporary until all callers migrate to MethodMetadata
	blockBody, ok := methodMeta.Body.(*ast.BlockStatement)
	if !ok {
		// Body must be a BlockStatement for function declarations
		return nil
	}

	// Reconstruct parameters from metadata
	params := make([]*ast.Parameter, len(methodMeta.Parameters))
	for i, paramMeta := range methodMeta.Parameters {
		// Reconstruct Type from TypeName for implicit conversion support
		var paramType ast.TypeExpression
		if paramMeta.TypeName != "" {
			paramType = &ast.TypeAnnotation{Name: paramMeta.TypeName}
		}
		params[i] = &ast.Parameter{
			Name:         &ast.Identifier{Value: paramMeta.Name},
			Type:         paramType,
			ByRef:        paramMeta.ByRef,
			DefaultValue: paramMeta.DefaultValue,
		}
	}

	// Reconstruct return type if present
	var returnType ast.TypeExpression
	if methodMeta.ReturnTypeName != "" {
		// Create a TypeAnnotation from the type name
		returnType = &ast.TypeAnnotation{
			Name: methodMeta.ReturnTypeName,
		}
	}

	return &ast.FunctionDecl{
		Name:          &ast.Identifier{Value: methodMeta.Name},
		Parameters:    params,
		ReturnType:    returnType,
		Body:          blockBody,
		IsClassMethod: methodMeta.IsClassMethod,
		IsConstructor: methodMeta.IsConstructor,
		IsDestructor:  methodMeta.IsDestructor,
	}
}

// RecordHasMethod checks if a method exists on the record.
// Task 9.7: Helper for record method resolution.
// Task 3.5.128b: Changed from method to free function due to type alias.
func RecordHasMethod(r *RecordValue, name string) bool {
	return GetRecordMethod(r, name) != nil
}

// ExternalVarValue represents an external variable marker.
// Task 3.5.130b: Type alias to runtime.ExternalVarValue for backward compatibility.
type ExternalVarValue = runtime.ExternalVarValue

// ReferenceValue represents a reference to a variable in another environment.
// Task 9.35: This is used to implement var parameters (by-reference parameters).
//
// NOTE (Task 3.5.131c): A new callback-based implementation exists in runtime.ReferenceValue
// that breaks the direct Environment dependency. The evaluator uses runtime.NewReferenceValue()
// directly. This implementation remains for backward compatibility with interpreter code.
//
// When a function has a var parameter, instead of copying the argument value,
// we create a ReferenceValue that points to the original variable in the caller's
// environment. This allows the function to modify the caller's variable.
//
// Example:
//
//	procedure Increment(var x: Integer);
//	begin
//	  x := x + 1;  // Modifies the caller's variable through the reference
//	end;
//
//	var n := 5;
//	Increment(n);  // n becomes 6
//
// Implementation:
//   - When calling Increment(n), instead of passing IntegerValue{5}, we pass
//     ReferenceValue{Env: callerEnv, VarName: "n"}
//   - When the function reads x, it dereferences to get the current value from callerEnv
//   - When the function assigns to x, it writes to the original variable in callerEnv
type ReferenceValue struct {
	Env     *Environment // The environment containing the variable
	VarName string       // The name of the variable being referenced
}

// Type returns "REFERENCE".
func (r *ReferenceValue) Type() string {
	return "REFERENCE"
}

// String returns a description of the reference.
func (r *ReferenceValue) String() string {
	return fmt.Sprintf("&%s", r.VarName)
}

// Dereference returns the current value of the referenced variable.
// Task 9.35: Helper to read through a reference.
func (r *ReferenceValue) Dereference() (Value, error) {
	val, ok := r.Env.Get(r.VarName)
	if !ok {
		return nil, fmt.Errorf("referenced variable %s not found", r.VarName)
	}
	return val, nil
}

// Assign sets the value of the referenced variable.
// Task 9.35: Helper to write through a reference.
func (r *ReferenceValue) Assign(value Value) error {
	return r.Env.Set(r.VarName, value)
}

// Task 3.5.139: VariantValue has been moved to runtime package.
// See internal/interp/runtime/variant.go for the implementation.
// Type alias is provided above for backward compatibility.

// ============================================================================
// Variant Boxing/Unboxing Helpers
// ============================================================================

// BoxVariant wraps any Value in a VariantValue for dynamic typing.
// Task 9.227: Implement VariantValue boxing in interpreter.
// Task 3.5.139: Delegates to runtime.BoxVariant.
//
// This function is kept for backward compatibility with code in the interp package.
// New code should use runtime.BoxVariant directly.
func BoxVariant(value Value) *VariantValue {
	return runtime.BoxVariant(value)
}

// unboxVariant extracts the underlying Value from a VariantValue.
// Task 9.228: Implement VariantValue unboxing in interpreter.
//
// Returns the wrapped value and true if successful, or nil and false if not a Variant.
// Examples:
//   - unboxVariant(VariantValue{Value: IntegerValue{42}}) → (IntegerValue{42}, true)
//   - unboxVariant(IntegerValue{42}) → (nil, false)
func unboxVariant(value Value) (Value, bool) {
	variant, ok := value.(*VariantValue)
	if !ok {
		return nil, false
	}
	return variant.Value, true
}

// unwrapVariant returns the underlying value if input is a Variant, otherwise returns input as-is.
// Task 9.228: Helper for operations that need to work with the actual value.
// Task 9.4.1: Updated to return UnassignedValue for uninitialized variants.
//
// Unlike unboxVariant, this always returns a valid Value (never nil, false).
// Examples:
//   - unwrapVariant(VariantValue{Value: IntegerValue{42}}) → IntegerValue{42}
//   - unwrapVariant(IntegerValue{42}) → IntegerValue{42}
//   - unwrapVariant(VariantValue{Value: nil}) → UnassignedValue{}
func unwrapVariant(value Value) Value {
	if variant, ok := value.(*VariantValue); ok {
		if variant.Value == nil {
			// Task 9.4.1: An uninitialized variant (nil value) is Unassigned
			return &UnassignedValue{}
		}
		return variant.Value
	}
	return value
}

// Helper functions to create values from Go types

// NewIntegerValue creates a new IntegerValue from an int64.
func NewIntegerValue(v int64) Value {
	return &IntegerValue{Value: v}
}

// NewFloatValue creates a new FloatValue from a float64.
func NewFloatValue(v float64) Value {
	return &FloatValue{Value: v}
}

// NewStringValue creates a new StringValue from a string.
func NewStringValue(v string) Value {
	return &StringValue{Value: v}
}

// NewBooleanValue creates a new BooleanValue from a bool.
func NewBooleanValue(v bool) Value {
	return &BooleanValue{Value: v}
}

// NewNilValue creates a new NilValue.
func NewNilValue() Value {
	return &NilValue{}
}

// NewNullValue creates a new NullValue.
// Task 9.4.1: Constructor for variant Null value.
func NewNullValue() Value {
	return &NullValue{}
}

// NewSetValue creates a new empty SetValue with the given set type.
// Task 9.8: Initializes the appropriate storage backend (bitmask or map).
// Phase 3.5.4: Forwarding function to runtime.NewSetValue.
func NewSetValue(setType *types.SetType) *SetValue {
	return runtime.NewSetValue(setType)
}

// NewUnassignedValue creates a new UnassignedValue.
// Task 9.4.1: Constructor for variant Unassigned value.
func NewUnassignedValue() Value {
	return &UnassignedValue{}
}

// NewTypeMetaValue creates a new TypeMetaValue.
// Task 9.133: Constructor for type meta-values.
func NewTypeMetaValue(typeInfo types.Type, typeName string) Value {
	return &TypeMetaValue{
		TypeInfo: typeInfo,
		TypeName: typeName,
	}
}

// getZeroValueForType returns the zero value for a given type.
// Task 9.7e1: Helper to initialize record fields with appropriate zero values.
// For nested records, methods should be provided via the methodsLookup callback.
func getZeroValueForType(t types.Type, methodsLookup func(*types.RecordType) map[string]*ast.FunctionDecl) Value {
	switch t {
	case types.INTEGER:
		return &IntegerValue{Value: 0}
	case types.FLOAT:
		return &FloatValue{Value: 0.0}
	case types.STRING:
		return &StringValue{Value: ""}
	case types.BOOLEAN:
		return &BooleanValue{Value: false}
	default:
		if arrayType, ok := t.(*types.ArrayType); ok {
			initializer := func(elementType types.Type, _ int) Value {
				return getZeroValueForType(elementType, methodsLookup)
			}
			return runtime.NewArrayValue(arrayType, initializer)
		}
		// Task 9.7e1: Handle nested records - recursively create RecordValue instances
		if recordType, ok := t.(*types.RecordType); ok {
			// For record types, create a new RecordValue instance
			// Task 3.5.128a: Pass nil metadata (will be populated later if needed)
			return newRecordValueInternal(recordType, nil, methodsLookup)
		}
		// For other complex types (classes, arrays, etc.), return nil
		return &NilValue{}
	}
}

// newRecordValueInternal is the internal implementation that supports recursive initialization.
// Task 3.5.42: Updated to accept RecordMetadata instead of AST method map.
// Task 3.5.128a: Removed deprecated methods parameter, kept methodsLookup for legacy compatibility.
// Task 3.5.128b: Now delegates to runtime.NewRecordValueWithInitializer.
func newRecordValueInternal(recordType *types.RecordType, metadata *runtime.RecordMetadata, methodsLookup func(*types.RecordType) map[string]*ast.FunctionDecl) *RecordValue {
	// Create initializer that handles zero values for all field types
	initializer := func(fieldName string, fieldType types.Type) runtime.Value {
		return getZeroValueForType(fieldType, methodsLookup)
	}

	return runtime.NewRecordValueWithInitializer(recordType, metadata, initializer)
}

// newRecordValueInternalWithMetadataLookup creates a record with metadata lookup for nested records.
// Task 3.5.128a: New function that uses metadata lookup instead of methods lookup.
// Task 3.5.128b: Now delegates to runtime.NewRecordValueWithMetadataLookup.
func newRecordValueInternalWithMetadataLookup(recordType *types.RecordType, metadata *runtime.RecordMetadata, metadataLookup func(*types.RecordType) *runtime.RecordMetadata) *RecordValue {
	// Create zero value provider for non-record fields
	zeroValueProvider := func(t types.Type) runtime.Value {
		return getZeroValueForType(t, nil)
	}

	return runtime.NewRecordValueWithMetadataLookup(recordType, metadata, metadataLookup, zeroValueProvider)
}

// NewRecordValue creates a new RecordValue with the given record type.
// Deprecated: Use NewRecordValueWithMetadata for AST-free creation.
// Task 3.5.128a: Updated to not require methods parameter (uses metadata only).
// Task 3.5.128b: Now delegates to runtime.NewRecordValue.
func NewRecordValue(recordType *types.RecordType) Value {
	return newRecordValueInternal(recordType, nil, nil)
}

// NewRecordValueWithMetadata creates a new RecordValue using RecordMetadata.
// Task 3.5.42: AST-free record creation using metadata.
// Task 3.5.128b: Now delegates to runtime constructor.
func NewRecordValueWithMetadata(recordType *types.RecordType, metadata *runtime.RecordMetadata) Value {
	return newRecordValueInternal(recordType, metadata, nil)
}

// ClassInfoValue is a special internal value type used to track the current class context
// in class methods. It wraps a ClassInfo pointer and is stored as "__CurrentClass__"
// in the environment when executing class methods.
type ClassInfoValue struct {
	ClassInfo *ClassInfo
}

// Type returns "CLASSINFO".
func (c *ClassInfoValue) Type() string {
	return "CLASSINFO"
}

// String returns the class name.
func (c *ClassInfoValue) String() string {
	return "class " + c.ClassInfo.Name
}

// GetClassName returns the class name.
// Task 3.5.88: Implements evaluator.ClassMetaValue interface.
func (c *ClassInfoValue) GetClassName() string {
	if c == nil || c.ClassInfo == nil {
		return ""
	}
	return c.ClassInfo.Name
}

// GetClassVar retrieves a class variable value by name from the class hierarchy.
// Returns the value and true if found, nil and false otherwise.
// Task 3.5.88: Implements evaluator.ClassMetaValue interface.
func (c *ClassInfoValue) GetClassVar(name string) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	value, owningClass := c.ClassInfo.lookupClassVar(name)
	if owningClass == nil {
		return nil, false
	}
	return value, true
}

// GetClassConstant retrieves a class constant value by name from the class hierarchy.
// Returns the value and true if found, nil and false otherwise.
// Task 3.5.88: Implements evaluator.ClassMetaValue interface.
func (c *ClassInfoValue) GetClassConstant(name string) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Check ConstantValues cache first (case-insensitive)
	for constName, value := range c.ClassInfo.ConstantValues {
		if ident.Equal(constName, name) {
			return value, true
		}
	}
	// Check parent class hierarchy
	if c.ClassInfo.Parent != nil {
		parentCIV := &ClassInfoValue{ClassInfo: c.ClassInfo.Parent}
		return parentCIV.GetClassConstant(name)
	}
	return nil, false
}

// HasClassMethod checks if a class method with the given name exists.
// Task 3.5.88: Implements evaluator.ClassMetaValue interface.
func (c *ClassInfoValue) HasClassMethod(name string) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	normalizedName := ident.Normalize(name)
	// Check single class methods
	if _, exists := c.ClassInfo.ClassMethods[normalizedName]; exists {
		return true
	}
	// Check overloaded class methods
	if overloads, exists := c.ClassInfo.ClassMethodOverloads[normalizedName]; exists && len(overloads) > 0 {
		return true
	}
	// Check parent class hierarchy
	if c.ClassInfo.Parent != nil {
		parentCIV := &ClassInfoValue{ClassInfo: c.ClassInfo.Parent}
		return parentCIV.HasClassMethod(name)
	}
	return false
}

// HasConstructor checks if a constructor with the given name exists.
// Task 3.5.88: Implements evaluator.ClassMetaValue interface.
func (c *ClassInfoValue) HasConstructor(name string) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	return c.ClassInfo.HasConstructor(name)
}

// TypeCastValue wraps an object with its static type from a type cast.
// This preserves the static type for member access, particularly for class variables.
// Example: TBase(childObj).ClassVar should access TBase's class variable, not TChild's.
type TypeCastValue struct {
	Object     Value      // The actual object value (ObjectInstance, nil, etc.)
	StaticType *ClassInfo // The static class type from the cast
}

// Type returns the static type's name.
func (t *TypeCastValue) Type() string {
	return t.StaticType.Name
}

// String delegates to the wrapped object's String().
func (t *TypeCastValue) String() string {
	return t.Object.String()
}

// GetStaticTypeName returns the static type name from the cast.
// Task 3.5.89: Implements evaluator.TypeCastAccessor interface.
func (t *TypeCastValue) GetStaticTypeName() string {
	if t.StaticType == nil {
		return ""
	}
	return t.StaticType.Name
}

// GetWrappedValue returns the actual value wrapped by the type cast.
// Task 3.5.89: Implements evaluator.TypeCastAccessor interface.
func (t *TypeCastValue) GetWrappedValue() Value {
	return t.Object
}

// GetStaticClassVar retrieves a class variable from the static type's class hierarchy.
// Task 3.5.89: Implements evaluator.TypeCastAccessor interface.
// This uses the static type from the cast, not the runtime type of the wrapped object.
func (t *TypeCastValue) GetStaticClassVar(name string) (Value, bool) {
	if t.StaticType == nil {
		return nil, false
	}
	value, owningClass := t.StaticType.lookupClassVar(name)
	if owningClass == nil {
		return nil, false
	}
	return value, true
}

// GoInt converts a Value to a Go int64. Returns error if not an IntegerValue.
func GoInt(v Value) (int64, error) {
	if iv, ok := v.(*IntegerValue); ok {
		return iv.Value, nil
	}
	return 0, fmt.Errorf("value is not an integer: %s", v.Type())
}

// GoFloat converts a Value to a Go float64. Returns error if not a FloatValue.
func GoFloat(v Value) (float64, error) {
	if fv, ok := v.(*FloatValue); ok {
		return fv.Value, nil
	}
	return 0, fmt.Errorf("value is not a float: %s", v.Type())
}

// GoString converts a Value to a Go string. Returns error if not a StringValue.
func GoString(v Value) (string, error) {
	if sv, ok := v.(*StringValue); ok {
		return sv.Value, nil
	}
	return "", fmt.Errorf("value is not a string: %s", v.Type())
}

// GoBool converts a Value to a Go bool. Returns error if not a BooleanValue.
func GoBool(v Value) (bool, error) {
	if bv, ok := v.(*BooleanValue); ok {
		return bv.Value, nil
	}
	return false, fmt.Errorf("value is not a boolean: %s", v.Type())
}

// NewArrayValue creates a new ArrayValue with the given array type.
// For static arrays, pre-allocates elements (initialized to nil).
// For dynamic arrays, creates an empty array.
// Phase 3.5.4: Forwarding function to runtime.NewArrayValue with initializer.
func NewArrayValue(arrayType *types.ArrayType) *ArrayValue {
	// Create an initializer that handles nested arrays and records
	initializer := func(elementType types.Type, index int) Value {
		// Task 9.56: For nested arrays, initialize each element as an array
		if nestedArrayType, ok := elementType.(*types.ArrayType); ok {
			return NewArrayValue(nestedArrayType)
		}
		// Task 9.36: For record elements, initialize each element as a record
		if recordType, ok := elementType.(*types.RecordType); ok {
			return NewRecordValue(recordType)
		}
		return nil
	}

	return runtime.NewArrayValue(arrayType, initializer)
}

// ============================================================================
// FunctionPointerValue - Runtime representation for function/method pointers
// ============================================================================
// Task 3.7.7: This is now an alias to runtime.FunctionPointerValue for backward compatibility.
// The canonical definition is in internal/interp/runtime/primitives.go.

type FunctionPointerValue = runtime.FunctionPointerValue

// NewFunctionPointerValue creates a new function pointer value.
// For regular functions, selfObject should be nil.
// For method pointers, selfObject should be the instance.
// Task 3.5.40: Maintains backward compatibility by accepting both AST node and MethodID.
func NewFunctionPointerValue(
	function *ast.FunctionDecl,
	closure *Environment,
	selfObject Value,
	pointerType *types.FunctionPointerType,
) *FunctionPointerValue {
	return &FunctionPointerValue{
		MethodID:    runtime.InvalidMethodID, // Will be set later if needed
		Function:    function,
		Lambda:      nil,
		Closure:     closure,
		SelfObject:  selfObject,
		PointerType: pointerType,
	}
}

// NewFunctionPointerValueWithID creates a new function pointer value using MethodID.
// This is the AST-free constructor for use after full migration.
// Task 3.5.41: AST-free function pointer creation.
func NewFunctionPointerValueWithID(
	methodID runtime.MethodID,
	closure *Environment,
	selfObject Value,
	pointerType *types.FunctionPointerType,
) *FunctionPointerValue {
	return &FunctionPointerValue{
		MethodID:    methodID,
		Function:    nil, // No AST node needed
		Lambda:      nil,
		Closure:     closure,
		SelfObject:  selfObject,
		PointerType: pointerType,
	}
}

// NewBuiltinFunctionPointerValue creates a function pointer value that targets a built-in function.
func NewBuiltinFunctionPointerValue(name string, pointerType *types.FunctionPointerType) *FunctionPointerValue {
	return &FunctionPointerValue{
		MethodID:    runtime.InvalidMethodID,
		Function:    nil,
		Lambda:      nil,
		Closure:     nil,
		SelfObject:  nil,
		PointerType: pointerType,
		BuiltinName: name,
	}
}

// NewLambdaValue creates a new lambda/closure value.
// Task 9.221: Constructor for lambda expressions/anonymous methods.
// The closure environment captures all variables from the scope where the lambda is defined.
// Task 3.5.41: Maintains backward compatibility during migration.
func NewLambdaValue(
	lambda *ast.LambdaExpression,
	closure *Environment,
	pointerType *types.FunctionPointerType,
) *FunctionPointerValue {
	return &FunctionPointerValue{
		MethodID:    runtime.InvalidMethodID, // Will be set later if needed
		Function:    nil,
		Lambda:      lambda,
		Closure:     closure,
		SelfObject:  nil,
		PointerType: pointerType,
	}
}

// NewLambdaValueWithID creates a new lambda/closure value using MethodID.
// This is the AST-free constructor for use after full migration.
// Task 3.5.41: AST-free lambda creation.
func NewLambdaValueWithID(
	methodID runtime.MethodID,
	closure *Environment,
	pointerType *types.FunctionPointerType,
) *FunctionPointerValue {
	return &FunctionPointerValue{
		MethodID:    methodID,
		Function:    nil,
		Lambda:      nil, // No AST node needed
		Closure:     closure,
		SelfObject:  nil,
		PointerType: pointerType,
	}
}

// ============================================================================
// JSONValue - Runtime representation for JSON values
// ============================================================================

// JSONValue represents a JSON value in DWScript.
// Task 9.89: Wrapper around jsonvalue.Value for DWScript runtime integration.
//
// This type wraps the internal jsonvalue.Value representation and exposes it
// to the DWScript runtime. JSON values can be:
//   - Primitives: null, boolean, number, string
//   - Containers: object, array
//
// JSON values are reference types - they maintain identity and can be mutated.
// They can be boxed in Variants to support heterogeneous collections.
//
// Example:
//
//	var obj := JSON.Parse('{"name": "John", "age": 30}');
//	PrintLn(obj.name);  // Outputs: John
//
// See reference/dwscript-original/Source/dwsJSONConnector.pas for the original
// TdwsJSONConnectorType implementation.
// JSONValue represents a JSON value in DWScript.
// Task 3.5.160a: Moved to runtime.JSONValue for evaluator access.
type JSONValue = runtime.JSONValue

// NewJSONValue creates a new JSONValue wrapping a jsonvalue.Value.
// Task 3.5.160a: This now delegates to runtime.NewJSONValue.
var NewJSONValue = runtime.NewJSONValue

// ============================================================================
// Ordinal Value Utilities
// ============================================================================
//
// Task 3.5.77: GetOrdinalValue and GetOrdinalType moved to evaluator package
// for direct access during evaluation. See: evaluator/ordinal_helpers.go
// ============================================================================
