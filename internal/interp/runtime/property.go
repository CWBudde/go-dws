// Package runtime provides runtime value types for the DWScript interpreter.
// This file contains property-related types and interfaces.
package runtime

// PropertyDescriptor provides metadata about a property.
//
// This allows runtime types to implement PropertyAccessor without depending
// on the evaluator package. The evaluator imports runtime and uses this type.
type PropertyDescriptor struct {
	Impl      any    // Implementation reference (types.PropertyInfo, types.RecordPropertyInfo, or runtime.PropertyInfo)
	Name      string // Property name
	ReadSpec  string // Field name or getter method name
	WriteSpec string // Field name or setter method name
	IsIndexed bool   // True if indexed property
	IsDefault bool   // True if default property
}

// PropertyAccessor is an optional interface for values that support property access.
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
