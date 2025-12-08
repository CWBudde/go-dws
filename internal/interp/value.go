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

// Value represents a runtime value in the DWScript interpreter.
type Value = runtime.Value

// Primitive value type aliases from runtime package
type (
	IntegerValue    = runtime.IntegerValue    // Integer values
	FloatValue      = runtime.FloatValue      // Floating-point values
	StringValue     = runtime.StringValue     // String values
	BooleanValue    = runtime.BooleanValue    // Boolean values
	NilValue        = runtime.NilValue        // Nil/null values
	NullValue       = runtime.NullValue       // Variant Null (explicit null state)
	UnassignedValue = runtime.UnassignedValue // Variant Unassigned (default state)
	EnumValue       = runtime.EnumValue       // Enumerated values
	TypeMetaValue   = runtime.TypeMetaValue   // Type references
	SetValue        = runtime.SetValue        // Set values
	IntRange        = runtime.IntRange        // Integer ranges for sets
	ArrayValue      = runtime.ArrayValue      // Array values
	VariantValue    = runtime.VariantValue    // Variant (dynamic type) values
	SubrangeValue   = runtime.SubrangeValue   // Integer subranges with bounds checking
)

// ============================================================================
// Value Interface Extensions
// ============================================================================

// Type operation interfaces
type (
	NumericValue     = runtime.NumericValue     // Supports numeric operations
	ComparableValue  = runtime.ComparableValue  // Supports equality comparison
	OrderableValue   = runtime.OrderableValue   // Supports ordering comparison
	CopyableValue    = runtime.CopyableValue    // Supports copying
	ReferenceType    = runtime.ReferenceType    // Reference-type values
	IndexableValue   = runtime.IndexableValue   // Supports indexing
	CallableValue    = runtime.CallableValue    // Can be called as functions
	ConvertibleValue = runtime.ConvertibleValue // Supports explicit type conversion
	IterableValue    = runtime.IterableValue    // Supports iteration
	Iterator         = runtime.Iterator         // Collection iterator
)

// ============================================================================
// Complex Value Types
// ============================================================================

// RTTITypeInfoValue represents runtime type information.
// Used by TypeOf() to return unique type identifiers for RTTI operations.
type RTTITypeInfoValue struct {
	TypeInfo types.Type
	TypeName string
	TypeID   int
}

func (r *RTTITypeInfoValue) Type() string   { return "RTTI_TYPEINFO" }
func (r *RTTITypeInfoValue) String() string { return r.TypeName }

// RecordValue represents a record (struct) value
type RecordValue = runtime.RecordValue

// ObjectInstance represents a class instance
type ObjectInstance = runtime.ObjectInstance

// Object construction and type checking helpers
var (
	NewObjectInstance = runtime.NewObjectInstance
	AsObject          = runtime.AsObject
	IsObject          = runtime.IsObject
)

// GetRecordMethod retrieves a method by name from a record (case-insensitive).
func GetRecordMethod(r *RecordValue, name string) *ast.FunctionDecl {
	if r.Metadata == nil {
		return nil
	}

	methodMeta, ok := r.Metadata.Methods[ident.Normalize(name)]
	if !ok || methodMeta.Body == nil {
		return nil
	}

	// Reconstruct FunctionDecl from metadata for compatibility
	blockBody, ok := methodMeta.Body.(*ast.BlockStatement)
	if !ok {
		return nil
	}

	// Reconstruct parameters
	params := make([]*ast.Parameter, len(methodMeta.Parameters))
	for i, paramMeta := range methodMeta.Parameters {
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

	// Reconstruct return type
	var returnType ast.TypeExpression
	if methodMeta.ReturnTypeName != "" {
		returnType = &ast.TypeAnnotation{Name: methodMeta.ReturnTypeName}
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
func RecordHasMethod(r *RecordValue, name string) bool {
	return GetRecordMethod(r, name) != nil
}

// ExternalVarValue marks external variables
type ExternalVarValue = runtime.ExternalVarValue

// ReferenceValue represents a reference to a variable (for var parameters).
// Allows functions to modify caller's variables through by-reference parameters.
type ReferenceValue struct {
	Env     *Environment // Environment containing the variable
	VarName string       // Name of the referenced variable
}

func (r *ReferenceValue) Type() string   { return "REFERENCE" }
func (r *ReferenceValue) String() string { return fmt.Sprintf("&%s", r.VarName) }

// Dereference returns the current value of the referenced variable.
func (r *ReferenceValue) Dereference() (Value, error) {
	val, ok := r.Env.Get(r.VarName)
	if !ok {
		return nil, fmt.Errorf("referenced variable %s not found", r.VarName)
	}
	return val, nil
}

// Assign sets the value of the referenced variable.
func (r *ReferenceValue) Assign(value Value) error {
	return r.Env.Set(r.VarName, value)
}

// ============================================================================
// Variant Boxing/Unboxing
// ============================================================================

// BoxVariant wraps a value in a VariantValue for dynamic typing.
func BoxVariant(value Value) *VariantValue {
	return runtime.BoxVariant(value)
}

// unboxVariant extracts the underlying value from a Variant.
// Returns (value, true) if successful, (nil, false) otherwise.
func unboxVariant(value Value) (Value, bool) {
	variant, ok := value.(*VariantValue)
	if !ok {
		return nil, false
	}
	return variant.Value, true
}

// unwrapVariant returns the underlying value if input is a Variant, otherwise returns input as-is.
// Uninitialized variants (nil value) return UnassignedValue.
func unwrapVariant(value Value) Value {
	if variant, ok := value.(*VariantValue); ok {
		if variant.Value == nil {
			return &UnassignedValue{}
		}
		return variant.Value
	}
	return value
}

// ============================================================================
// Value Constructors
// ============================================================================

// Primitive value constructors
func NewIntegerValue(v int64) Value  { return &IntegerValue{Value: v} }
func NewFloatValue(v float64) Value  { return &FloatValue{Value: v} }
func NewStringValue(v string) Value  { return &StringValue{Value: v} }
func NewBooleanValue(v bool) Value   { return &BooleanValue{Value: v} }
func NewNilValue() Value             { return &NilValue{} }
func NewNullValue() Value            { return &NullValue{} }
func NewUnassignedValue() Value      { return &UnassignedValue{} }

// NewSetValue creates an empty set with the given type.
func NewSetValue(setType *types.SetType) *SetValue {
	return runtime.NewSetValue(setType)
}

// NewTypeMetaValue creates a type reference value.
func NewTypeMetaValue(typeInfo types.Type, typeName string) Value {
	return &TypeMetaValue{TypeInfo: typeInfo, TypeName: typeName}
}

// getZeroValueForType returns the zero value for a given type.
// Recursively initializes nested arrays and records.
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
		if recordType, ok := t.(*types.RecordType); ok {
			return newRecordValueInternal(recordType, nil, methodsLookup)
		}
		return &NilValue{}
	}
}

// newRecordValueInternal creates a record with recursive field initialization.
func newRecordValueInternal(recordType *types.RecordType, metadata *runtime.RecordMetadata, methodsLookup func(*types.RecordType) map[string]*ast.FunctionDecl) *RecordValue {
	initializer := func(fieldName string, fieldType types.Type) runtime.Value {
		return getZeroValueForType(fieldType, methodsLookup)
	}
	return runtime.NewRecordValueWithInitializer(recordType, metadata, initializer)
}

// newRecordValueInternalWithMetadataLookup creates a record with metadata-aware initialization.
func newRecordValueInternalWithMetadataLookup(recordType *types.RecordType, metadata *runtime.RecordMetadata, metadataLookup func(*types.RecordType) *runtime.RecordMetadata) *RecordValue {
	zeroValueProvider := func(t types.Type) runtime.Value {
		return getZeroValueForType(t, nil)
	}
	return runtime.NewRecordValueWithMetadataLookup(recordType, metadata, metadataLookup, zeroValueProvider)
}

// NewRecordValue creates a new record with default field values.
func NewRecordValue(recordType *types.RecordType) Value {
	return newRecordValueInternal(recordType, nil, nil)
}

// NewRecordValueWithMetadata creates a new record using metadata (AST-free).
func NewRecordValueWithMetadata(recordType *types.RecordType, metadata *runtime.RecordMetadata) Value {
	return newRecordValueInternal(recordType, metadata, nil)
}

// ClassInfoValue tracks current class context in class methods.
// Stored as "__CurrentClass__" in the environment during method execution.
type ClassInfoValue struct {
	ClassInfo *ClassInfo
}

func (c *ClassInfoValue) Type() string   { return "CLASSINFO" }
func (c *ClassInfoValue) String() string { return "class " + c.ClassInfo.Name }

// GetClassName returns the class name.
func (c *ClassInfoValue) GetClassName() string {
	if c == nil || c.ClassInfo == nil {
		return ""
	}
	return c.ClassInfo.Name
}

// GetClassType returns the class type (metaclass) as a ClassValue.
func (c *ClassInfoValue) GetClassType() Value {
	if c == nil || c.ClassInfo == nil {
		return nil
	}
	return &ClassValue{ClassInfo: c.ClassInfo}
}

// GetClassVar retrieves a class variable by name from the hierarchy.
func (c *ClassInfoValue) GetClassVar(name string) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	value, owningClass := c.ClassInfo.lookupClassVar(name)
	return value, owningClass != nil
}

// GetClassConstant retrieves a class constant by name from the hierarchy.
func (c *ClassInfoValue) GetClassConstant(name string) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Check current class (case-insensitive)
	for constName, value := range c.ClassInfo.ConstantValues {
		if ident.Equal(constName, name) {
			return value, true
		}
	}
	// Check parent hierarchy
	if c.ClassInfo.Parent != nil {
		parentCIV := &ClassInfoValue{ClassInfo: c.ClassInfo.Parent}
		return parentCIV.GetClassConstant(name)
	}
	return nil, false
}

// HasClassMethod checks if a class method exists in the hierarchy.
func (c *ClassInfoValue) HasClassMethod(name string) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	normalizedName := ident.Normalize(name)
	// Check single methods and overloads
	if _, exists := c.ClassInfo.ClassMethods[normalizedName]; exists {
		return true
	}
	if overloads, exists := c.ClassInfo.ClassMethodOverloads[normalizedName]; exists && len(overloads) > 0 {
		return true
	}
	// Check parent
	if c.ClassInfo.Parent != nil {
		parentCIV := &ClassInfoValue{ClassInfo: c.ClassInfo.Parent}
		return parentCIV.HasClassMethod(name)
	}
	return false
}

// HasConstructor checks if a constructor exists.
func (c *ClassInfoValue) HasConstructor(name string) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	return c.ClassInfo.HasConstructor(name)
}

// TypeCastValue wraps an object with its static type from a cast.
// Preserves static type for member access (e.g., TBase(obj).ClassVar uses TBase's class var).
type TypeCastValue struct {
	Object     Value      // Actual object value
	StaticType *ClassInfo // Static class type from the cast
}

func (t *TypeCastValue) Type() string   { return "TYPE_CAST" }
func (t *TypeCastValue) String() string { return t.Object.String() }

// GetStaticTypeName returns the static type name.
func (t *TypeCastValue) GetStaticTypeName() string {
	if t.StaticType == nil {
		return ""
	}
	return t.StaticType.Name
}

// GetWrappedValue returns the actual wrapped value.
func (t *TypeCastValue) GetWrappedValue() Value {
	return t.Object
}

// GetStaticClassVar retrieves a class variable using the static type.
func (t *TypeCastValue) GetStaticClassVar(name string) (Value, bool) {
	if t.StaticType == nil {
		return nil, false
	}
	value, owningClass := t.StaticType.lookupClassVar(name)
	return value, owningClass != nil
}

// ============================================================================
// Value to Go Type Conversions
// ============================================================================

func GoInt(v Value) (int64, error) {
	if iv, ok := v.(*IntegerValue); ok {
		return iv.Value, nil
	}
	return 0, fmt.Errorf("value is not an integer: %s", v.Type())
}

func GoFloat(v Value) (float64, error) {
	if fv, ok := v.(*FloatValue); ok {
		return fv.Value, nil
	}
	return 0, fmt.Errorf("value is not a float: %s", v.Type())
}

func GoString(v Value) (string, error) {
	if sv, ok := v.(*StringValue); ok {
		return sv.Value, nil
	}
	return "", fmt.Errorf("value is not a string: %s", v.Type())
}

func GoBool(v Value) (bool, error) {
	if bv, ok := v.(*BooleanValue); ok {
		return bv.Value, nil
	}
	return false, fmt.Errorf("value is not a boolean: %s", v.Type())
}

// NewArrayValue creates a new array with the given type.
// Static arrays are pre-allocated; dynamic arrays start empty.
func NewArrayValue(arrayType *types.ArrayType) *ArrayValue {
	initializer := func(elementType types.Type, index int) Value {
		if nestedArrayType, ok := elementType.(*types.ArrayType); ok {
			return NewArrayValue(nestedArrayType)
		}
		if recordType, ok := elementType.(*types.RecordType); ok {
			return NewRecordValue(recordType)
		}
		return nil
	}
	return runtime.NewArrayValue(arrayType, initializer)
}

// ============================================================================
// Function Pointer Values
// ============================================================================

type FunctionPointerValue = runtime.FunctionPointerValue

// NewFunctionPointerValue creates a function pointer (regular or method).
func NewFunctionPointerValue(function *ast.FunctionDecl, closure *Environment, selfObject Value, pointerType *types.FunctionPointerType) *FunctionPointerValue {
	return &FunctionPointerValue{
		MethodID:    runtime.InvalidMethodID,
		Function:    function,
		Lambda:      nil,
		Closure:     closure,
		SelfObject:  selfObject,
		PointerType: pointerType,
	}
}

// NewFunctionPointerValueWithID creates a function pointer using MethodID (AST-free).
func NewFunctionPointerValueWithID(methodID runtime.MethodID, closure *Environment, selfObject Value, pointerType *types.FunctionPointerType) *FunctionPointerValue {
	return &FunctionPointerValue{
		MethodID:    methodID,
		Function:    nil,
		Lambda:      nil,
		Closure:     closure,
		SelfObject:  selfObject,
		PointerType: pointerType,
	}
}

// NewBuiltinFunctionPointerValue creates a function pointer to a built-in function.
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

// NewLambdaValue creates a lambda/closure value.
func NewLambdaValue(lambda *ast.LambdaExpression, closure *Environment, pointerType *types.FunctionPointerType) *FunctionPointerValue {
	return &FunctionPointerValue{
		MethodID:    runtime.InvalidMethodID,
		Function:    nil,
		Lambda:      lambda,
		Closure:     closure,
		SelfObject:  nil,
		PointerType: pointerType,
	}
}

// NewLambdaValueWithID creates a lambda/closure value using MethodID (AST-free).
func NewLambdaValueWithID(methodID runtime.MethodID, closure *Environment, pointerType *types.FunctionPointerType) *FunctionPointerValue {
	return &FunctionPointerValue{
		MethodID:    methodID,
		Function:    nil,
		Lambda:      nil,
		Closure:     closure,
		SelfObject:  nil,
		PointerType: pointerType,
	}
}

// ============================================================================
// JSON Values
// ============================================================================

// JSONValue represents a JSON value (primitives, objects, arrays).
type JSONValue = runtime.JSONValue

var NewJSONValue = runtime.NewJSONValue
