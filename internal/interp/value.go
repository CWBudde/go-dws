// Package interp provides the interpreter and runtime for DWScript.
package interp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/internal/types"
)

// Value represents a runtime value in the DWScript interpreter.
// All runtime values must implement this interface.
// This interface does NOT use interface{} to ensure type safety.
type Value interface {
	// Type returns the type name of the value (e.g., "INTEGER", "STRING")
	Type() string
	// String returns the string representation of the value
	String() string
}

// IntegerValue represents an integer value in DWScript.
type IntegerValue struct {
	Value int64
}

// Type returns "INTEGER".
func (i *IntegerValue) Type() string {
	return "INTEGER"
}

// String returns the string representation of the integer.
func (i *IntegerValue) String() string {
	return strconv.FormatInt(i.Value, 10)
}

// FloatValue represents a floating-point value in DWScript.
type FloatValue struct {
	Value float64
}

// Type returns "FLOAT".
func (f *FloatValue) Type() string {
	return "FLOAT"
}

// String returns the string representation of the float.
func (f *FloatValue) String() string {
	return strconv.FormatFloat(f.Value, 'g', -1, 64)
}

// StringValue represents a string value in DWScript.
type StringValue struct {
	Value string
}

// Type returns "STRING".
func (s *StringValue) Type() string {
	return "STRING"
}

// String returns the string value itself.
func (s *StringValue) String() string {
	return s.Value
}

// BooleanValue represents a boolean value in DWScript.
type BooleanValue struct {
	Value bool
}

// Type returns "BOOLEAN".
func (b *BooleanValue) Type() string {
	return "BOOLEAN"
}

// String returns "true" or "false".
func (b *BooleanValue) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

// NilValue represents a nil/null value in DWScript.
type NilValue struct{}

// Type returns "NIL".
func (n *NilValue) Type() string {
	return "NIL"
}

// String returns "nil".
func (n *NilValue) String() string {
	return "nil"
}

// TypeMetaValue represents a type name as a runtime value in DWScript.
// Task 9.133: DWScript allows type names like `Integer` to be used as values.
// This is used for reflection and type-based operations like High(Integer), Low(Integer).
//
// Examples:
//   - High(Integer) where `Integer` is a TypeMetaValue wrapping types.INTEGER
//   - Low(Boolean) where `Boolean` is a TypeMetaValue wrapping types.BOOLEAN
//   - High(TColor) where `TColor` is a TypeMetaValue wrapping the enum type
//
// See reference/dwscript-original/Source/dwsExprs.pas for type info expressions.
type TypeMetaValue struct {
	TypeInfo types.Type // The type metadata (e.g., types.INTEGER, types.FLOAT, enum type)
	TypeName string     // The type name for display (e.g., "Integer", "TColor")
}

// Type returns "TYPE_META".
func (t *TypeMetaValue) Type() string {
	return "TYPE_META"
}

// String returns the type name.
func (t *TypeMetaValue) String() string {
	return t.TypeName
}

// EnumValue represents an enum value in DWScript.
// Task 8.49: Store enum values with their ordinal value and type name.
type EnumValue struct {
	TypeName     string // Enum type name (e.g., "TColor")
	ValueName    string // Enum value name (e.g., "Red")
	OrdinalValue int    // The ordinal value (e.g., 0 for Red if implicit)
}

// Type returns "ENUM".
func (e *EnumValue) Type() string {
	return "ENUM"
}

// String returns the enum value name.
func (e *EnumValue) String() string {
	return e.ValueName
}

// RecordValue represents a record value in DWScript.
// Task 8.73: Store record type metadata and field values.
// Task 9.7: Extended to include method declarations for runtime invocation.
// Records are value types (like structs) with fields and methods.
type RecordValue struct {
	RecordType *types.RecordType            // The record type metadata
	Fields     map[string]Value             // Field name -> runtime value mapping
	Methods    map[string]*ast.FunctionDecl // Method name -> AST declaration (Task 9.7)
}

// Type returns the record type name (e.g., "TFoo") or "RECORD" if unnamed.
func (r *RecordValue) Type() string {
	if r.RecordType != nil && r.RecordType.Name != "" {
		return r.RecordType.Name
	}
	return "RECORD"
}

// String returns the string representation of the record.
func (r *RecordValue) String() string {
	var sb strings.Builder

	// Show type name if available
	if r.RecordType != nil && r.RecordType.Name != "" {
		sb.WriteString(r.RecordType.Name)
		sb.WriteString("(")
	} else {
		sb.WriteString("record(")
	}

	// Sort field names for consistent output
	fieldNames := make([]string, 0, len(r.Fields))
	for name := range r.Fields {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	// Add field values
	for i, name := range fieldNames {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(name)
		sb.WriteString(": ")
		if val := r.Fields[name]; val != nil {
			sb.WriteString(val.String())
		} else {
			sb.WriteString("nil")
		}
	}

	sb.WriteString(")")
	return sb.String()
}

// Copy creates a deep copy of the record value
// Records have value semantics in DWScript, so assignment should copy.
// Task 9.7: Updated to copy methods as well.
func (r *RecordValue) Copy() *RecordValue {
	copiedFields := make(map[string]Value, len(r.Fields))

	// Deep copy all fields
	for name, val := range r.Fields {
		// Check if the value is also a record that needs copying
		if recVal, ok := val.(*RecordValue); ok {
			copiedFields[name] = recVal.Copy()
		} else {
			// For basic types (Integer, String, etc.), they're already immutable or copied by value
			copiedFields[name] = val
		}
	}

	return &RecordValue{
		RecordType: r.RecordType,
		Fields:     copiedFields,
		Methods:    r.Methods, // Methods are shared (AST nodes are immutable)
	}
}

// GetMethod retrieves a method declaration by name.
// Task 9.7: Helper for record method invocation.
func (r *RecordValue) GetMethod(name string) *ast.FunctionDecl {
	if r.Methods == nil {
		return nil
	}
	return r.Methods[name]
}

// HasMethod checks if a method exists on the record.
// Task 9.7: Helper for record method resolution.
func (r *RecordValue) HasMethod(name string) bool {
	return r.GetMethod(name) != nil
}

// ExternalVarValue represents an external variable marker.
// Task 7.144: This is a special marker stored in the environment to indicate
// that a variable is external. Attempting to read or write this value raises an error.
type ExternalVarValue struct {
	Name         string // The variable name in DWScript
	ExternalName string // The external name for FFI binding (may be empty)
}

// VariantValue represents a Variant value in DWScript.
// Task 9.221: Variant is a dynamic type that can hold any runtime value.
//
// The VariantValue wraps another Value and tracks its actual runtime type.
// This enables:
// - Heterogeneous arrays (array of Variant)
// - Dynamic type conversions at runtime
// - Type introspection (VarType, VarIsNull, etc.)
// - Polymorphic function parameters (array of const)
//
// Similar to Delphi's TVarData structure and DWScript's IDataContext for variants.
// The wrapped value can be any Value type: Integer, Float, String, Boolean,
// Array, Record, Object, etc.
//
// Example:
//
//	var v: Variant := 42;
//	// Creates: VariantValue{Value: IntegerValue{42}, ActualType: INTEGER}
//
//	v := 'hello';
//	// Creates: VariantValue{Value: StringValue{'hello'}, ActualType: STRING}
//
// See reference/dwscript-original/Source/dwsVariantFunctions.pas
type VariantValue struct {
	Value      Value      // The wrapped runtime value
	ActualType types.Type // The actual type of the wrapped value (for type checking)
}

// Type returns "VARIANT" to identify this as a Variant value.
func (v *VariantValue) Type() string {
	return "VARIANT"
}

// String returns the string representation by delegating to the wrapped value.
// This allows Variant values to be printed naturally.
func (v *VariantValue) String() string {
	if v.Value == nil {
		return "Unassigned" // Similar to Delphi's unassigned variant
	}
	return v.Value.String()
}

// ============================================================================
// Variant Boxing/Unboxing Helpers
// ============================================================================

// boxVariant wraps any Value in a VariantValue for dynamic typing.
// Task 9.227: Implement VariantValue boxing in interpreter.
//
// Boxing preserves the original value and tracks its type for later unboxing.
// Examples:
//   - boxVariant(&IntegerValue{42}) → VariantValue{Value: IntegerValue{42}, ActualType: INTEGER}
//   - boxVariant(&StringValue{"hello"}) → VariantValue{Value: StringValue{"hello"}, ActualType: STRING}
//   - boxVariant(nil) → VariantValue{Value: nil, ActualType: nil}
func boxVariant(value Value) *VariantValue {
	if value == nil {
		return &VariantValue{Value: nil, ActualType: nil}
	}

	// If already a Variant, return as-is (no double-wrapping)
	if variant, ok := value.(*VariantValue); ok {
		return variant
	}

	// Map runtime Value type to semantic types.Type
	var actualType types.Type
	switch value.Type() {
	case "INTEGER":
		actualType = types.INTEGER
	case "FLOAT":
		actualType = types.FLOAT
	case "STRING":
		actualType = types.STRING
	case "BOOLEAN":
		actualType = types.BOOLEAN
	case "NIL":
		actualType = nil // nil has no type
	// Complex types (arrays, records, objects) will be added as needed
	// For now, we store nil for ActualType and rely on Value.Type()
	default:
		actualType = nil
	}

	return &VariantValue{
		Value:      value,
		ActualType: actualType,
	}
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
//
// Unlike unboxVariant, this always returns a valid Value (never nil, false).
// Examples:
//   - unwrapVariant(VariantValue{Value: IntegerValue{42}}) → IntegerValue{42}
//   - unwrapVariant(IntegerValue{42}) → IntegerValue{42}
//   - unwrapVariant(VariantValue{Value: nil}) → NilValue{}
func unwrapVariant(value Value) Value {
	if variant, ok := value.(*VariantValue); ok {
		if variant.Value == nil {
			return &NilValue{}
		}
		return variant.Value
	}
	return value
}

// Type returns "EXTERNAL_VAR".
func (e *ExternalVarValue) Type() string {
	return "EXTERNAL_VAR"
}

// String returns a description of the external variable.
func (e *ExternalVarValue) String() string {
	if e.ExternalName != "" {
		return fmt.Sprintf("external(%s -> %s)", e.Name, e.ExternalName)
	}
	return fmt.Sprintf("external(%s)", e.Name)
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
		// Task 9.7e1: Handle nested records - recursively create RecordValue instances
		if recordType, ok := t.(*types.RecordType); ok {
			// For record types, create a new RecordValue instance with methods
			var methods map[string]*ast.FunctionDecl
			if methodsLookup != nil {
				methods = methodsLookup(recordType)
			}
			return newRecordValueInternal(recordType, methods, methodsLookup)
		}
		// For other complex types (classes, arrays, etc.), return nil
		return &NilValue{}
	}
}

// newRecordValueInternal is the internal implementation that supports recursive initialization.
func newRecordValueInternal(recordType *types.RecordType, methods map[string]*ast.FunctionDecl, methodsLookup func(*types.RecordType) map[string]*ast.FunctionDecl) *RecordValue {
	fields := make(map[string]Value)

	// Task 9.7e1: Initialize all fields with zero values
	// This ensures fields are accessible in record methods even before being explicitly assigned
	for fieldName, fieldType := range recordType.Fields {
		fields[fieldName] = getZeroValueForType(fieldType, methodsLookup)
	}

	return &RecordValue{
		RecordType: recordType,
		Fields:     fields,
		Methods:    methods,
	}
}

// NewRecordValue creates a new RecordValue with the given record type.
// Task 8.73: Initialize the fields map.
// Task 9.7: Add methods parameter for record method invocation.
// Task 9.7e1: Initialize fields with zero values so they can be accessed in methods.
func NewRecordValue(recordType *types.RecordType, methods map[string]*ast.FunctionDecl) Value {
	return newRecordValueInternal(recordType, methods, nil)
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

// ============================================================================
// SetValue - Runtime representation for set types
// ============================================================================

// SetValue represents a set value in DWScript.
// Sets are based on enum types and use hybrid storage for efficiency.
// Task 9.8: Support both small and large enums:
//   - Small enums (≤64 values): uint64 bitset (fast, 8 bytes)
//   - Large enums (>64 values): map[int]bool (unlimited size)
// The storage strategy is determined by SetType.StorageKind.
type SetValue struct {
	SetType  *types.SetType // The set type metadata (includes storage strategy)
	Elements uint64         // Bitset for small enums (used when StorageKind == Bitmask)
	MapStore map[int]bool   // Map for large enums (used when StorageKind == Map)
}

// Type returns "SET".
func (s *SetValue) Type() string {
	return "SET"
}

// String returns the string representation of the set.
// Format: [element1, element2, ...] or [] for empty set
// Task 9.8: Works with both bitmask and map storage.
func (s *SetValue) String() string {
	// Quick check for empty set (both storage types)
	if s.SetType.StorageKind == types.SetStorageBitmask && s.Elements == 0 {
		return "[]"
	}
	if s.SetType.StorageKind == types.SetStorageMap && len(s.MapStore) == 0 {
		return "[]"
	}

	var elements []string

	// Iterate through all possible enum values in order
	if s.SetType != nil && s.SetType.ElementType != nil {
		for _, name := range s.SetType.ElementType.OrderedNames {
			ordinal := s.SetType.ElementType.Values[name]
			if s.HasElement(ordinal) {
				elements = append(elements, name)
			}
		}
	}

	if len(elements) == 0 {
		return "[]"
	}

	return "[" + strings.Join(elements, ", ") + "]"
}

// HasElement checks if an element with the given ordinal value is in the set.
// Task 9.8: Supports both bitmask and map storage.
func (s *SetValue) HasElement(ordinal int) bool {
	if ordinal < 0 {
		return false // Negative ordinals are invalid
	}

	// Choose storage backend based on set type
	switch s.SetType.StorageKind {
	case types.SetStorageBitmask:
		if ordinal >= 64 {
			return false // Out of range for bitset
		}
		mask := uint64(1) << uint(ordinal)
		return (s.Elements & mask) != 0

	case types.SetStorageMap:
		return s.MapStore[ordinal]

	default:
		return false
	}
}

// AddElement adds an element with the given ordinal value to the set.
// This mutates the set in place (used for Include).
// Task 9.8: Supports both bitmask and map storage.
func (s *SetValue) AddElement(ordinal int) {
	if ordinal < 0 {
		return // Negative ordinals are invalid
	}

	// Choose storage backend based on set type
	switch s.SetType.StorageKind {
	case types.SetStorageBitmask:
		if ordinal >= 64 {
			return // Out of range for bitset
		}
		mask := uint64(1) << uint(ordinal)
		s.Elements |= mask

	case types.SetStorageMap:
		s.MapStore[ordinal] = true
	}
}

// RemoveElement removes an element with the given ordinal value from the set.
// This mutates the set in place (used for Exclude).
// Task 9.8: Supports both bitmask and map storage.
func (s *SetValue) RemoveElement(ordinal int) {
	if ordinal < 0 {
		return // Negative ordinals are invalid
	}

	// Choose storage backend based on set type
	switch s.SetType.StorageKind {
	case types.SetStorageBitmask:
		if ordinal >= 64 {
			return // Out of range for bitset
		}
		mask := uint64(1) << uint(ordinal)
		s.Elements &^= mask // AND NOT to clear the bit

	case types.SetStorageMap:
		delete(s.MapStore, ordinal)
	}
}

// NewSetValue creates a new empty SetValue with the given set type.
// Task 9.8: Initializes the appropriate storage backend (bitmask or map).
func NewSetValue(setType *types.SetType) *SetValue {
	sv := &SetValue{
		SetType:  setType,
		Elements: 0,
	}

	// Initialize map storage if needed for large enums
	if setType.StorageKind == types.SetStorageMap {
		sv.MapStore = make(map[int]bool)
	}

	return sv
}

// ============================================================================
// ArrayValue - Runtime representation for array types
// ============================================================================

// ArrayValue represents an array value in DWScript.
// DWScript supports both static arrays (with fixed bounds) and dynamic arrays (resizable).
// Examples:
//   - Static: array[1..10] of Integer
//   - Dynamic: array of String
type ArrayValue struct {
	ArrayType *types.ArrayType // The array type metadata
	Elements  []Value          // The runtime elements (slice)
}

// Type returns "ARRAY".
func (a *ArrayValue) Type() string {
	return "ARRAY"
}

// String returns the string representation of the array.
// Format: [element1, element2, ...] or [] for empty array
func (a *ArrayValue) String() string {
	if len(a.Elements) == 0 {
		return "[]"
	}

	var elements []string
	for _, elem := range a.Elements {
		if elem != nil {
			elements = append(elements, elem.String())
		} else {
			elements = append(elements, "nil")
		}
	}

	return "[" + strings.Join(elements, ", ") + "]"
}

// NewArrayValue creates a new ArrayValue with the given array type.
// For static arrays, pre-allocates elements (initialized to nil).
// For dynamic arrays, creates an empty array.
func NewArrayValue(arrayType *types.ArrayType) *ArrayValue {
	var elements []Value

	if arrayType.IsStatic() {
		// Static array: pre-allocate with size
		size := arrayType.Size()
		elements = make([]Value, size)

		// Task 9.56: For nested arrays, initialize each element as an array
		if arrayType.ElementType != nil {
			if nestedArrayType, ok := arrayType.ElementType.(*types.ArrayType); ok {
				for i := 0; i < size; i++ {
					elements[i] = NewArrayValue(nestedArrayType)
				}
			}
		}
		// Otherwise elements are nil (will be filled with zero values or explicit assignments)
	} else {
		// Dynamic array: start empty
		elements = make([]Value, 0)
	}

	return &ArrayValue{
		ArrayType: arrayType,
		Elements:  elements,
	}
}

// ============================================================================
// FunctionPointerValue - Runtime representation for function/method pointers
// ============================================================================

// FunctionPointerValue represents a function or procedure pointer in DWScript.
// Task 9.164: Create runtime representation for function pointers.
// Task 9.221: Extended to support lambda expressions/anonymous methods.
//
// Function pointers store a reference to a callable function/procedure along with
// its closure environment. Method pointers additionally capture the Self object.
// Lambdas are also represented using this type, with Lambda field set instead of Function.
//
// Examples:
//   - Function pointer: var f: TFunc; f := @MyFunction;
//   - Method pointer: var m: TMethod; m := @obj.MyMethod; (captures obj as Self)
//   - Lambda: var f := lambda(x: Integer): Integer begin Result := x * 2; end;
type FunctionPointerValue struct {
	// Function is the AST node of the function/procedure being pointed to
	// Either Function OR Lambda will be set, never both
	Function *ast.FunctionDecl

	// Lambda is the AST node of the lambda expression (anonymous method)
	// Either Function OR Lambda will be set, never both
	// Task 9.221: Added for lambda/closure support
	Lambda *ast.LambdaExpression

	// Closure is the environment where the function/lambda was defined
	// For lambdas, this captures all variables from outer scopes
	// For functions, this is typically the global environment
	Closure *Environment

	// SelfObject is the object instance for method pointers (nil for regular functions)
	// When non-nil, this function pointer is a method pointer ("of object")
	SelfObject Value

	// PointerType is the function pointer type information
	PointerType *types.FunctionPointerType
}

// Type returns "FUNCTION_POINTER", "METHOD_POINTER", or "LAMBDA" (closure).
// Task 9.221: Updated to distinguish lambdas.
func (f *FunctionPointerValue) Type() string {
	if f.SelfObject != nil {
		return "METHOD_POINTER"
	}
	if f.Lambda != nil {
		return "LAMBDA"
	}
	return "FUNCTION_POINTER"
}

// String returns the string representation of the function pointer.
// Format: @FunctionName, @Object.MethodName, or <lambda> for closures
// Task 9.221: Updated to handle lambdas.
func (f *FunctionPointerValue) String() string {
	// Lambda closures
	if f.Lambda != nil {
		return "<lambda>"
	}

	// Regular function/method pointers
	if f.Function == nil {
		return "@<nil>"
	}

	if f.SelfObject != nil {
		return "@" + f.SelfObject.String() + "." + f.Function.Name.Value
	}

	return "@" + f.Function.Name.Value
}

// NewFunctionPointerValue creates a new function pointer value.
// For regular functions, selfObject should be nil.
// For method pointers, selfObject should be the instance.
func NewFunctionPointerValue(
	function *ast.FunctionDecl,
	closure *Environment,
	selfObject Value,
	pointerType *types.FunctionPointerType,
) *FunctionPointerValue {
	return &FunctionPointerValue{
		Function:    function,
		Lambda:      nil,
		Closure:     closure,
		SelfObject:  selfObject,
		PointerType: pointerType,
	}
}

// NewLambdaValue creates a new lambda/closure value.
// Task 9.221: Constructor for lambda expressions/anonymous methods.
// The closure environment captures all variables from the scope where the lambda is defined.
func NewLambdaValue(
	lambda *ast.LambdaExpression,
	closure *Environment,
	pointerType *types.FunctionPointerType,
) *FunctionPointerValue {
	return &FunctionPointerValue{
		Function:    nil,
		Lambda:      lambda,
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
type JSONValue struct {
	Value *jsonvalue.Value // The underlying JSON value
}

// Type returns "JSON".
func (j *JSONValue) Type() string {
	return "JSON"
}

// String returns the string representation of the JSON value.
// For primitives, returns the value directly.
// For objects and arrays, returns a JSON-like representation.
func (j *JSONValue) String() string {
	if j.Value == nil {
		return "undefined"
	}

	switch j.Value.Kind() {
	case jsonvalue.KindUndefined:
		return "undefined"
	case jsonvalue.KindNull:
		return "null"
	case jsonvalue.KindBoolean:
		// Extract boolean value for string representation
		// We need to access the primitive payload, but it's not exported
		// For now, use a simple approach
		return j.jsonValueToString(j.Value)
	case jsonvalue.KindString:
		return j.jsonValueToString(j.Value)
	case jsonvalue.KindNumber, jsonvalue.KindInt64:
		return j.jsonValueToString(j.Value)
	case jsonvalue.KindObject:
		return j.jsonObjectToString(j.Value)
	case jsonvalue.KindArray:
		return j.jsonArrayToString(j.Value)
	default:
		return "unknown"
	}
}

// jsonValueToString converts a primitive JSON value to string.
// This is a helper for String() method.
func (j *JSONValue) jsonValueToString(v *jsonvalue.Value) string {
	switch v.Kind() {
	case jsonvalue.KindNull:
		return "null"
	case jsonvalue.KindBoolean:
		if v.BoolValue() {
			return "true"
		}
		return "false"
	case jsonvalue.KindString:
		return v.StringValue()
	case jsonvalue.KindNumber:
		return strconv.FormatFloat(v.NumberValue(), 'g', -1, 64)
	case jsonvalue.KindInt64:
		return strconv.FormatInt(v.Int64Value(), 10)
	default:
		return "undefined"
	}
}

// jsonObjectToString converts a JSON object to string representation.
func (j *JSONValue) jsonObjectToString(v *jsonvalue.Value) string {
	keys := v.ObjectKeys()
	if len(keys) == 0 {
		return "{}"
	}

	var sb strings.Builder
	sb.WriteString("{")
	for i, key := range keys {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(key)
		sb.WriteString(": ")
		child := v.ObjectGet(key)
		if child != nil {
			// Recursively stringify child values
			childJSON := &JSONValue{Value: child}
			sb.WriteString(childJSON.String())
		} else {
			sb.WriteString("undefined")
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// jsonArrayToString converts a JSON array to string representation.
func (j *JSONValue) jsonArrayToString(v *jsonvalue.Value) string {
	length := v.ArrayLen()
	if length == 0 {
		return "[]"
	}

	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < length; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		child := v.ArrayGet(i)
		if child != nil {
			// Recursively stringify child values
			childJSON := &JSONValue{Value: child}
			sb.WriteString(childJSON.String())
		} else {
			sb.WriteString("undefined")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// NewJSONValue creates a new JSONValue wrapping a jsonvalue.Value.
func NewJSONValue(v *jsonvalue.Value) *JSONValue {
	return &JSONValue{Value: v}
}
