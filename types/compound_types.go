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

// RecordType represents a record (struct) type.
// Records are value types with named fields.
// Example:
//
//	type TPoint = record
//	  X: Integer;
//	  Y: Integer;
//	end;
type RecordType struct {
	Name   string          // Record type name (e.g., "TPoint")
	Fields map[string]Type // Field name -> field type mapping
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

// NewRecordType creates a new record type with the given name and fields
func NewRecordType(name string, fields map[string]Type) *RecordType {
	return &RecordType{
		Name:   name,
		Fields: fields,
	}
}
