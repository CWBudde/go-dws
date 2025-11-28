// Package runtime provides runtime value types for the DWScript interpreter.
// This file contains property-related types and interfaces.
//
// Task 3.5.128b: PropertyDescriptor moved from evaluator to runtime
// to enable RecordValue to implement PropertyAccessor without circular imports.
package runtime

// PropertyDescriptor provides metadata about a property.
// Task 3.5.99a: Abstracts property info across classes, interfaces, and records.
// Task 3.5.128b: Moved from evaluator to runtime package.
//
// This allows runtime types to implement PropertyAccessor without depending
// on the evaluator package. The evaluator imports runtime and uses this type.
type PropertyDescriptor struct {
	Name      string // Property name
	IsIndexed bool   // True if this is an indexed property (e.g., property Items[Index: Integer]: String)
	IsDefault bool   // True if this is the default property

	// For implementation reference:
	// - Objects: pointer to types.PropertyInfo
	// - Interfaces: pointer to types.PropertyInfo
	// - Records: pointer to types.RecordPropertyInfo
	// We store as `any` to avoid circular imports and maintain type flexibility
	Impl any
}

// PropertyAccessor is an optional interface for values that support property access.
// Task 3.5.99a: Enables direct property lookup without adapter delegation.
// Task 3.5.128b: Moved from evaluator to runtime package.
//
// This interface is implemented by:
// - ObjectInstance (for class instance properties)
// - InterfaceInstance (for interface properties)
// - RecordValue (for record properties)
type PropertyAccessor interface {
	Value
	// LookupProperty searches for a property by name in the type hierarchy.
	// Returns a PropertyDescriptor with metadata needed for property access.
	// Returns nil if the property is not found.
	// The lookup is case-insensitive and includes parent types where applicable.
	LookupProperty(name string) *PropertyDescriptor

	// GetDefaultProperty returns the default property for this type, if any.
	// Default properties allow indexing syntax: obj[index] instead of obj.Property[index].
	// Returns nil if no default property is defined.
	GetDefaultProperty() *PropertyDescriptor
}
