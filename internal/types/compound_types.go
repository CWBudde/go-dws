package types

import (
	"fmt"
	"sort"
	"strings"
)

// ============================================================================
// ArrayType
// ============================================================================

// ArrayType represents an array type.
// DWScript supports both static arrays (with bounds) and dynamic arrays.
// Examples:
//   - array[1..10] of Integer (static, with bounds)
//   - array of String (dynamic, no bounds)
type ArrayType struct {
	ElementType Type // Type of elements in the array
	LowBound    *int // Lower bound (nil for dynamic arrays)
	HighBound   *int // Upper bound (nil for dynamic arrays)
}

// String returns a string representation of the array type
func (at *ArrayType) String() string {
	if at.LowBound != nil && at.HighBound != nil {
		return fmt.Sprintf("array[%d..%d] of %s", *at.LowBound, *at.HighBound, at.ElementType.String())
	}
	return fmt.Sprintf("array of %s", at.ElementType.String())
}

// TypeKind returns "ARRAY" for array types
func (at *ArrayType) TypeKind() string {
	return "ARRAY"
}

// Equals checks if two array types are equal.
// Two array types are equal if they have the same element type and bounds.
func (at *ArrayType) Equals(other Type) bool {
	otherArray, ok := other.(*ArrayType)
	if !ok {
		return false
	}

	// Element types must match
	if !at.ElementType.Equals(otherArray.ElementType) {
		return false
	}

	// Check bounds
	// Both must be static or both must be dynamic
	if (at.LowBound == nil) != (otherArray.LowBound == nil) {
		return false
	}
	if (at.HighBound == nil) != (otherArray.HighBound == nil) {
		return false
	}

	// If both are static, bounds must match
	if at.LowBound != nil && otherArray.LowBound != nil {
		if *at.LowBound != *otherArray.LowBound {
			return false
		}
	}
	if at.HighBound != nil && otherArray.HighBound != nil {
		if *at.HighBound != *otherArray.HighBound {
			return false
		}
	}

	return true
}

// IsDynamic returns true if this is a dynamic array (no bounds)
func (at *ArrayType) IsDynamic() bool {
	return at.LowBound == nil && at.HighBound == nil
}

// IsStatic returns true if this is a static array (with bounds)
func (at *ArrayType) IsStatic() bool {
	return !at.IsDynamic()
}

// Size returns the size of a static array, or -1 for dynamic arrays
func (at *ArrayType) Size() int {
	if at.IsDynamic() {
		return -1
	}
	return *at.HighBound - *at.LowBound + 1
}

// NewDynamicArrayType creates a new dynamic array type
func NewDynamicArrayType(elementType Type) *ArrayType {
	return &ArrayType{
		ElementType: elementType,
		LowBound:    nil,
		HighBound:   nil,
	}
}

// NewStaticArrayType creates a new static array type with bounds
func NewStaticArrayType(elementType Type, lowBound, highBound int) *ArrayType {
	low := lowBound
	high := highBound
	return &ArrayType{
		ElementType: elementType,
		LowBound:    &low,
		HighBound:   &high,
	}
}

// ============================================================================
// RecordType
// ============================================================================

// RecordPropertyInfo represents a property in a record type
// Properties provide controlled access to fields.
// Note: Renamed from PropertyInfo to avoid conflict with class PropertyInfo
type RecordPropertyInfo struct {
	Name       string // Property name
	Type       Type   // Property type
	ReadField  string // Field name for reading (can be empty for write-only)
	WriteField string // Field name for writing (can be empty for read-only)
}

// RecordType represents a record (struct) type.
// Records are value types with named fields.
// Example:
//
//	type TPoint = record
//	  X: Integer;
//	  Y: Integer;
//	end;
type RecordType struct {
	Fields               map[string]Type
	Methods              map[string]*FunctionType // Instance methods (primary signature)
	MethodOverloads      map[string][]*MethodInfo // Instance method overloads
	ClassMethods         map[string]*FunctionType // Static (class) methods (primary signature)
	ClassMethodOverloads map[string][]*MethodInfo // Static method overloads
	Properties           map[string]*RecordPropertyInfo
	Name                 string
}

// String returns a string representation of the record type
func (rt *RecordType) String() string {
	if rt.Name != "" {
		return rt.Name
	}

	// If no name, show fields
	var sb strings.Builder
	sb.WriteString("record { ")

	// Sort field names for consistent output
	fieldNames := make([]string, 0, len(rt.Fields))
	for name := range rt.Fields {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	for i, name := range fieldNames {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(name)
		sb.WriteString(": ")
		sb.WriteString(rt.Fields[name].String())
	}
	sb.WriteString(" }")

	return sb.String()
}

// TypeKind returns "RECORD" for record types
func (rt *RecordType) TypeKind() string {
	return "RECORD"
}

// Equals checks if two record types are equal.
// Two record types are equal if they have the same fields with the same types.
// Named records are only equal if they have the same name.
func (rt *RecordType) Equals(other Type) bool {
	otherRecord, ok := other.(*RecordType)
	if !ok {
		return false
	}

	// If both have names, names must match (nominal typing)
	if rt.Name != "" && otherRecord.Name != "" {
		return rt.Name == otherRecord.Name
	}

	// Structural typing: check fields
	if len(rt.Fields) != len(otherRecord.Fields) {
		return false
	}

	for name, typ := range rt.Fields {
		otherType, exists := otherRecord.Fields[name]
		if !exists {
			return false
		}
		if !typ.Equals(otherType) {
			return false
		}
	}

	return true
}

// HasField checks if the record has a field with the given name
func (rt *RecordType) HasField(name string) bool {
	_, exists := rt.Fields[name]
	return exists
}

// GetFieldType returns the type of a field, or nil if not found
func (rt *RecordType) GetFieldType(name string) Type {
	return rt.Fields[name]
}

// HasMethod checks if the record has a method with the given name
func (rt *RecordType) HasMethod(name string) bool {
	_, exists := rt.Methods[name]
	return exists
}

// GetMethod returns the type of a method, or nil if not found
func (rt *RecordType) GetMethod(name string) *FunctionType {
	return rt.Methods[name]
}

// HasClassMethod checks if the record has a class method with the given name
func (rt *RecordType) HasClassMethod(name string) bool {
	_, exists := rt.ClassMethods[name]
	return exists
}

// GetClassMethod returns the type of a class method, or nil if not found
func (rt *RecordType) GetClassMethod(name string) *FunctionType {
	return rt.ClassMethods[name]
}

// HasProperty checks if the record has a property with the given name
func (rt *RecordType) HasProperty(name string) bool {
	_, exists := rt.Properties[name]
	return exists
}

// GetProperty returns the property info, or nil if not found
func (rt *RecordType) GetProperty(name string) *RecordPropertyInfo {
	return rt.Properties[name]
}

// GetMethodOverloads returns all overload variants for a given method name
func (rt *RecordType) GetMethodOverloads(methodName string) []*MethodInfo {
	return rt.MethodOverloads[methodName]
}

// GetClassMethodOverloads returns all overload variants for a given class method name
func (rt *RecordType) GetClassMethodOverloads(methodName string) []*MethodInfo {
	return rt.ClassMethodOverloads[methodName]
}

// NewRecordType creates a new record type with the given name and fields
func NewRecordType(name string, fields map[string]Type) *RecordType {
	return &RecordType{
		Name:                 name,
		Fields:               fields,
		Methods:              make(map[string]*FunctionType),
		MethodOverloads:      make(map[string][]*MethodInfo),
		ClassMethods:         make(map[string]*FunctionType),
		ClassMethodOverloads: make(map[string][]*MethodInfo),
		Properties:           make(map[string]*RecordPropertyInfo),
	}
}

// ============================================================================
// SetType
// ============================================================================

// SetStorageKind determines the internal storage strategy for set values.
type SetStorageKind int

const (
	// SetStorageBitmask uses a uint64 bitset for small enums (≤64 values).
	// This is the default and most efficient storage for common cases.
	SetStorageBitmask SetStorageKind = iota

	// SetStorageMap uses map[int]bool for large enums (>64 values).
	// This allows unlimited set sizes with reasonable performance.
	SetStorageMap
)

// String returns a string representation of the storage kind
func (sk SetStorageKind) String() string {
	switch sk {
	case SetStorageBitmask:
		return "bitmask"
	case SetStorageMap:
		return "map"
	default:
		return "unknown"
	}
}

// SetType represents a set type.
// Sets are based on ordinal types (enum, integer, char/string, subrange) and support
// operations like Include, Exclude, union (+), difference (-), intersection (*), and
// membership tests (in).
// Examples:
//   - type TDays = set of TWeekday;  // enum-based set
//   - var flags: set of TOption;      // enum-based set
//   - var digits: set of [0..9];      // integer subrange set
//   - var letters: set of ['a'..'z']; // character range set
type SetType struct {
	ElementType Type           // Type of elements in the set (any ordinal type)
	StorageKind SetStorageKind // Storage strategy: bitmask (≤64 values) or map (>64 values)
}

// String returns a string representation of the set type
func (st *SetType) String() string {
	return fmt.Sprintf("set of %s", st.ElementType.String())
}

// TypeKind returns "SET" for set types
func (st *SetType) TypeKind() string {
	return "SET"
}

// Equals checks if two set types are equal.
// Two set types are equal if they have the same element type.
func (st *SetType) Equals(other Type) bool {
	otherSet, ok := other.(*SetType)
	if !ok {
		return false
	}

	// Element types must match
	return st.ElementType.Equals(otherSet.ElementType)
}

// NewSetType creates a new set type with the given element type.
//
// Storage selection:
//   - If element type has ≤64 possible values: uses bitmask (fast, memory-efficient)
//   - If element type has >64 possible values: uses map (unlimited size)
//   - If elementType is nil: defaults to bitmask (will be validated later)
//
// For different element types:
//   - EnumType: count of enum values
//   - SubrangeType: (HighBound - LowBound + 1)
//   - IntegerType: always uses map (unbounded)
//   - StringType: always uses map (unbounded, used for character sets)
func NewSetType(elementType Type) *SetType {
	// Determine storage strategy based on element type size
	storageKind := SetStorageBitmask

	if elementType != nil {
		switch et := elementType.(type) {
		case *EnumType:
			// Use map storage for large enums (>64 values)
			if len(et.OrderedNames) > 64 {
				storageKind = SetStorageMap
			}
		case *SubrangeType:
			// Use map storage for large subranges (>64 values)
			rangeSize := et.HighBound - et.LowBound + 1
			if rangeSize > 64 {
				storageKind = SetStorageMap
			}
		case *IntegerType, *StringType:
			// Unbounded types always use map storage
			// (Integer sets and char/string sets can have arbitrary values)
			storageKind = SetStorageMap
		default:
			// For other types (shouldn't happen for valid ordinal types),
			// default to bitmask and let validation catch it later
			storageKind = SetStorageBitmask
		}
	}

	return &SetType{
		ElementType: elementType,
		StorageKind: storageKind,
	}
}

// ============================================================================
// EnumType
// ============================================================================

// EnumType represents an enumerated type.
// Enums are ordinal types with named constant values.
// Examples:
//
//	type TColor = (Red, Green, Blue);
//	type TEnum = (One = 1, Two = 5);
type EnumType struct {
	Name         string         // Enum type name (e.g., "TColor")
	Values       map[string]int // Value name -> ordinal value mapping (forward lookup)
	OrderedNames []string       // Ordered list of value names for reverse lookup
}

// String returns a string representation of the enum type
func (et *EnumType) String() string {
	return et.Name
}

// TypeKind returns "ENUM" for enum types
func (et *EnumType) TypeKind() string {
	return "ENUM"
}

// Equals checks if two enum types are equal.
// Two enum types are equal if they have the same name (nominal typing).
func (et *EnumType) Equals(other Type) bool {
	otherEnum, ok := other.(*EnumType)
	if !ok {
		return false
	}
	return et.Name == otherEnum.Name
}

// GetEnumValue returns the ordinal value for a given enum value name.
// Returns -1 if the name is not found.
func (et *EnumType) GetEnumValue(name string) int {
	if val, ok := et.Values[name]; ok {
		return val
	}
	return -1
}

// GetEnumName returns the enum value name for a given ordinal value.
// Returns empty string if the value is not found.
func (et *EnumType) GetEnumName(value int) string {
	// Use OrderedNames for reverse lookup
	for _, name := range et.OrderedNames {
		if et.Values[name] == value {
			return name
		}
	}
	return ""
}

// MinOrdinal returns the minimum ordinal value in the enum.
// For example, for type TDay = (Mon=1, Tue, Wed), returns 1.
// Returns 0 if the enum has no values.
func (et *EnumType) MinOrdinal() int {
	if len(et.Values) == 0 {
		return 0
	}

	minVal := int(^uint(0) >> 1) // Max int value
	for _, ordinal := range et.Values {
		if ordinal < minVal {
			minVal = ordinal
		}
	}
	return minVal
}

// MaxOrdinal returns the maximum ordinal value in the enum.
// For example, for type TDay = (Mon=1, Tue, Wed), returns 3.
// Returns 0 if the enum has no values.
func (et *EnumType) MaxOrdinal() int {
	if len(et.Values) == 0 {
		return 0
	}

	maxVal := -int(^uint(0)>>1) - 1 // Min int value
	for _, ordinal := range et.Values {
		if ordinal > maxVal {
			maxVal = ordinal
		}
	}
	return maxVal
}

// NewEnumType creates a new enum type with the given name and values.
// Values should be a map of value names to their ordinal values.
// OrderedNames should list the value names in declaration order.
func NewEnumType(name string, values map[string]int, orderedNames []string) *EnumType {
	return &EnumType{
		Name:         name,
		Values:       values,
		OrderedNames: orderedNames,
	}
}

// ============================================================================
// HelperType
// ============================================================================

// HelperType represents a helper type that extends an existing type with
// additional methods, properties, and class members.
// Helpers cannot modify the original type; they only add new functionality.
//
// DWScript syntax:
//
//	type TStringHelper = helper for String
//	  function ToUpper: String;
//	end;
//
//	type TPointHelper = record helper for TPoint
//	  function Distance: Float;
//	end;
//
// Helpers can also inherit from other helpers:
//
//	type TParentHelper = helper for String
//	  function ToUpper: String;
//	end;
//
//	type TChildHelper = helper(TParentHelper) for String
//	  function ToLower: String;
//	end;
type HelperType struct {
	TargetType     Type
	ParentHelper   *HelperType // Parent helper for inheritance (optional)
	Methods        map[string]*FunctionType
	Properties     map[string]*PropertyInfo
	ClassVars      map[string]Type
	ClassConsts    map[string]interface{}
	BuiltinMethods map[string]string
	Name           string
	IsRecordHelper bool
}

// String returns the string representation of the helper type
func (ht *HelperType) String() string {
	if ht.IsRecordHelper {
		return fmt.Sprintf("record helper for %s", ht.TargetType.String())
	}
	return fmt.Sprintf("helper for %s", ht.TargetType.String())
}

// TypeKind returns "HELPER" for helper types
func (ht *HelperType) TypeKind() string {
	return "HELPER"
}

// Equals checks if two helper types are equal.
// Two helpers are equal if they have the same name and target type.
func (ht *HelperType) Equals(other Type) bool {
	otherHelper, ok := other.(*HelperType)
	if !ok {
		return false
	}

	// Names must match
	if ht.Name != otherHelper.Name {
		return false
	}

	// Target types must match
	if !ht.TargetType.Equals(otherHelper.TargetType) {
		return false
	}

	return true
}

// GetMethod looks up a method by name in this helper.
// If not found in this helper, searches in parent helper (if any).
// Returns the method type and true if found, nil and false otherwise.
func (ht *HelperType) GetMethod(name string) (*FunctionType, bool) {
	// Look in this helper first
	method, ok := ht.Methods[name]
	if ok {
		return method, true
	}

	// If not found and we have a parent, look there
	if ht.ParentHelper != nil {
		return ht.ParentHelper.GetMethod(name)
	}

	return nil, false
}

// GetProperty looks up a property by name in this helper.
// If not found in this helper, searches in parent helper (if any).
// Returns the property info and true if found, nil and false otherwise.
func (ht *HelperType) GetProperty(name string) (*PropertyInfo, bool) {
	// Look in this helper first
	prop, ok := ht.Properties[name]
	if ok {
		return prop, true
	}

	// If not found and we have a parent, look there
	if ht.ParentHelper != nil {
		return ht.ParentHelper.GetProperty(name)
	}

	return nil, false
}

// GetClassVar looks up a class variable by name in this helper.
// If not found in this helper, searches in parent helper (if any).
// Returns the variable type and true if found, nil and false otherwise.
func (ht *HelperType) GetClassVar(name string) (Type, bool) {
	// Look in this helper first
	varType, ok := ht.ClassVars[name]
	if ok {
		return varType, true
	}

	// If not found and we have a parent, look there
	if ht.ParentHelper != nil {
		return ht.ParentHelper.GetClassVar(name)
	}

	return nil, false
}

// GetClassConst looks up a class constant by name in this helper.
// If not found in this helper, searches in parent helper (if any).
// Returns the constant value and true if found, nil and false otherwise.
func (ht *HelperType) GetClassConst(name string) (interface{}, bool) {
	// Look in this helper first
	constVal, ok := ht.ClassConsts[name]
	if ok {
		return constVal, true
	}

	// If not found and we have a parent, look there
	if ht.ParentHelper != nil {
		return ht.ParentHelper.GetClassConst(name)
	}

	return nil, false
}

// NewHelperType creates a new helper type.
func NewHelperType(name string, targetType Type, isRecordHelper bool) *HelperType {
	return &HelperType{
		Name:           name,
		TargetType:     targetType,
		Methods:        make(map[string]*FunctionType),
		Properties:     make(map[string]*PropertyInfo),
		ClassVars:      make(map[string]Type),
		ClassConsts:    make(map[string]interface{}),
		BuiltinMethods: make(map[string]string),
		IsRecordHelper: isRecordHelper,
	}
}
