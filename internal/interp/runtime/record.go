// Package runtime provides runtime value types for the DWScript interpreter.
// This file contains the RecordValue type for record instances.
//
// Phase 3.5.4 - Type Migration: Migrated from internal/interp to runtime/
// to enable evaluator package to work with record values directly.
//
// Task 3.5.128b: RecordValue struct moved to runtime package.
package runtime

import (
	"sort"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// RecordValue represents a record value in DWScript.
// Records are value types with named fields and optional methods.
//
// Task 3.5.42: Migrated to use RecordMetadata instead of AST-dependent method map.
// Task 3.5.128a: Removed deprecated Methods field - now uses only Metadata.Methods.
// Task 3.5.128b: Moved to runtime package for direct evaluator access.
//
// Examples:
//
//	type TPoint = record
//	  X, Y: Integer;
//	end;
//
//	var p: TPoint;
//	p.X := 10;
//	p.Y := 20;
type RecordValue struct {
	RecordType *types.RecordType // The record type metadata
	Fields     map[string]Value  // Field name -> runtime value mapping
	Metadata   *RecordMetadata   // Runtime metadata (methods, constants, etc.)
}

// Type returns the record type name (e.g., "TFoo") or "RECORD" if unnamed.
func (r *RecordValue) Type() string {
	if r.RecordType != nil && r.RecordType.Name != "" {
		return r.RecordType.Name
	}
	return "RECORD"
}

// String returns the string representation of the record.
// Format: TypeName(field1: value1, field2: value2) or record(...)
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

// Copy creates a deep copy of the record value.
// Records have value semantics in DWScript, so assignment should copy.
// Task 9.7: Updated to copy methods as well.
// Task 3.5.42: Updated to copy Metadata reference.
// Task 3.5.128a: Removed deprecated Methods field copying.
// Task 3.5.128b: Moved to runtime package.
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
		Metadata:   r.Metadata, // Metadata is shared (immutable)
	}
}

// GetRecordField retrieves a field value by name (case-insensitive lookup).
// Returns the field value and true if found, nil and false otherwise.
// Task 3.5.91: Implements RecordInstanceValue interface for direct field access.
// Task 3.5.128b: Moved to runtime package.
func (r *RecordValue) GetRecordField(name string) (Value, bool) {
	if r.Fields == nil {
		return nil, false
	}
	// Case-insensitive lookup
	for fieldName, value := range r.Fields {
		if ident.Equal(fieldName, name) {
			return value, true
		}
	}
	return nil, false
}

// SetRecordField sets a field value by name (case-insensitive lookup).
// Returns true if the field was found and set, false otherwise.
// Task 3.5.128b: Added for direct field modification support.
func (r *RecordValue) SetRecordField(name string, value Value) bool {
	if r.Fields == nil {
		return false
	}
	// Case-insensitive lookup to find the canonical field name
	for fieldName := range r.Fields {
		if ident.Equal(fieldName, name) {
			r.Fields[fieldName] = value
			return true
		}
	}
	return false
}

// GetRecordTypeName returns the record type name (e.g., "TPoint").
// Returns "RECORD" if the type name is not available.
// Task 3.5.91: Implements RecordInstanceValue interface.
// Task 3.5.128b: Moved to runtime package.
func (r *RecordValue) GetRecordTypeName() string {
	return r.Type()
}

// HasRecordMethod checks if a method with the given name exists on this record type.
// The lookup is case-insensitive.
// Task 3.5.91: Implements RecordInstanceValue interface.
// Task 3.5.128b: Moved to runtime package.
func (r *RecordValue) HasRecordMethod(name string) bool {
	if r.Metadata == nil {
		return false
	}
	normalizedName := ident.Normalize(name)
	_, ok := r.Metadata.Methods[normalizedName]
	return ok
}

// HasRecordProperty checks if a property with the given name exists.
// Note: Records CAN have properties in DWScript (though less common than classes).
// Task 3.5.91: Implements RecordInstanceValue interface.
// Task 3.5.128b: Moved to runtime package.
func (r *RecordValue) HasRecordProperty(name string) bool {
	if r.RecordType == nil || r.RecordType.Properties == nil {
		return false
	}
	normalizedName := ident.Normalize(name)
	_, exists := r.RecordType.Properties[normalizedName]
	return exists
}

// GetRecordType returns the record's type metadata.
// Task 3.5.128b: Added for type introspection support.
func (r *RecordValue) GetRecordType() *types.RecordType {
	return r.RecordType
}

// GetMetadata returns the record's runtime metadata (methods, constants, etc.).
// Task 3.5.128b: Added for method lookup support.
func (r *RecordValue) GetMetadata() *RecordMetadata {
	return r.Metadata
}

// LookupProperty searches for a property by name in the record type.
// Task 3.5.99a: Implements PropertyAccessor interface.
// Task 3.5.128b: Moved to runtime package.
// Returns a PropertyDescriptor wrapping types.RecordPropertyInfo, or nil if not found.
func (r *RecordValue) LookupProperty(name string) *PropertyDescriptor {
	if r.RecordType == nil || r.RecordType.Properties == nil {
		return nil
	}

	normalizedName := ident.Normalize(name)
	propInfo, exists := r.RecordType.Properties[normalizedName]
	if !exists {
		return nil
	}

	return &PropertyDescriptor{
		Name:      propInfo.Name,
		IsIndexed: false, // RecordPropertyInfo doesn't have IsIndexed field, records use simple field-based properties
		IsDefault: propInfo.IsDefault,
		Impl:      propInfo, // Store the original RecordPropertyInfo for later use
	}
}

// GetDefaultProperty returns the default property for this record type, if any.
// Task 3.5.99a: Implements PropertyAccessor interface.
// Task 3.5.128b: Moved to runtime package.
// Returns a PropertyDescriptor wrapping types.RecordPropertyInfo, or nil if no default property exists.
func (r *RecordValue) GetDefaultProperty() *PropertyDescriptor {
	if r.RecordType == nil || r.RecordType.Properties == nil {
		return nil
	}

	// Search for the default property
	for _, propInfo := range r.RecordType.Properties {
		if propInfo.IsDefault {
			return &PropertyDescriptor{
				Name:      propInfo.Name,
				IsIndexed: false, // RecordPropertyInfo doesn't have IsIndexed field
				IsDefault: true,
				Impl:      propInfo, // Store the original RecordPropertyInfo for later use
			}
		}
	}

	return nil
}

// ReadIndexedProperty reads an indexed property value using the provided executor callback.
// Task 3.5.118: Enables direct indexed property access for records without adapter delegation.
// Task 3.5.128b: Moved to runtime package.
// The propInfo is already resolved by PropertyAccessor.LookupProperty or GetDefaultProperty.
func (r *RecordValue) ReadIndexedProperty(propInfo any, indices []Value, propertyExecutor func(propInfo any, indices []Value) Value) Value {
	if r == nil || r.RecordType == nil {
		return &ErrorValue{Message: "record has no type information"}
	}
	// Delegate to executor - record validation is done, propInfo already resolved by caller
	return propertyExecutor(propInfo, indices)
}

// RecordFieldInitializer is a function type for initializing record fields.
// This allows the interp package to provide custom initialization logic
// for complex field types (nested records, arrays, etc.) without creating
// circular dependencies.
//
// Task 3.5.128b: Analogous to ArrayElementInitializer for records.
type RecordFieldInitializer func(fieldName string, fieldType types.Type) Value

// RecordMetadataLookup is a function type for looking up record metadata.
// This allows nested record creation to find metadata for embedded record types
// without circular dependencies on the interpreter.
//
// Task 3.5.128b: Used for nested record initialization.
type RecordMetadataLookup func(recordType *types.RecordType) *RecordMetadata

// NewRecordValue creates a new RecordValue with the given record type.
// Fields are initialized to nil - use NewRecordValueWithInitializer for
// proper zero-value initialization.
//
// Task 3.5.128b: Basic constructor moved to runtime package.
func NewRecordValue(recordType *types.RecordType, metadata *RecordMetadata) *RecordValue {
	fields := make(map[string]Value)

	// Initialize fields to nil (caller should set values)
	for fieldName := range recordType.Fields {
		fields[fieldName] = nil
	}

	return &RecordValue{
		RecordType: recordType,
		Fields:     fields,
		Metadata:   metadata,
	}
}

// NewRecordValueWithInitializer creates a new RecordValue with custom field initialization.
// The initializer function is called for each field to provide the initial value.
// This supports proper zero-value initialization for primitive and complex types.
//
// Task 3.5.128b: Constructor with initializer callback.
func NewRecordValueWithInitializer(
	recordType *types.RecordType,
	metadata *RecordMetadata,
	initializer RecordFieldInitializer,
) *RecordValue {
	fields := make(map[string]Value)

	// Initialize all fields using the provided initializer
	for fieldName, fieldType := range recordType.Fields {
		if initializer != nil {
			fields[fieldName] = initializer(fieldName, fieldType)
		} else {
			fields[fieldName] = nil
		}
	}

	return &RecordValue{
		RecordType: recordType,
		Fields:     fields,
		Metadata:   metadata,
	}
}

// NewRecordValueWithMetadataLookup creates a new RecordValue that can recursively
// initialize nested record fields with proper metadata.
//
// Task 3.5.128b: Constructor for nested record support.
func NewRecordValueWithMetadataLookup(
	recordType *types.RecordType,
	metadata *RecordMetadata,
	metadataLookup RecordMetadataLookup,
	zeroValueProvider func(t types.Type) Value,
) *RecordValue {
	fields := make(map[string]Value)

	// Initialize all fields with zero values
	for fieldName, fieldType := range recordType.Fields {
		// Handle nested record types with metadata lookup
		if nestedRecordType, ok := fieldType.(*types.RecordType); ok && metadataLookup != nil {
			nestedMetadata := metadataLookup(nestedRecordType)
			fields[fieldName] = NewRecordValueWithMetadataLookup(
				nestedRecordType,
				nestedMetadata,
				metadataLookup,
				zeroValueProvider,
			)
		} else if zeroValueProvider != nil {
			fields[fieldName] = zeroValueProvider(fieldType)
		} else {
			fields[fieldName] = nil
		}
	}

	return &RecordValue{
		RecordType: recordType,
		Fields:     fields,
		Metadata:   metadata,
	}
}
