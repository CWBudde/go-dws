// Package runtime provides the core runtime value system for the DWScript interpreter.
// This package contains value type definitions, interfaces for type operations,
// and utilities for value manipulation.
package runtime

// Value represents a runtime value in the DWScript interpreter.
// All runtime values must implement this interface.
type Value interface {
	// Type returns the type name of the value (e.g., "INTEGER", "STRING")
	Type() string
	// String returns the string representation of the value
	String() string
}

// NumericValue represents values that can be used in numeric operations.
// Values implementing this interface can be converted to integer or float
// for arithmetic operations.
type NumericValue interface {
	Value
	// AsInteger converts the value to an integer if possible.
	// Returns the integer value and true if conversion succeeded, 0 and false otherwise.
	AsInteger() (int64, bool)
	// AsFloat converts the value to a float if possible.
	// Returns the float value and true if conversion succeeded, 0.0 and false otherwise.
	AsFloat() (float64, bool)
}

// ComparableValue represents values that can be compared for equality.
// This interface is used for equality checks (= and <> operators).
type ComparableValue interface {
	Value
	// Equals checks if this value is equal to another value.
	// Returns true if equal, false otherwise, and an error if comparison is not possible.
	Equals(other Value) (bool, error)
}

// OrderableValue represents values that can be ordered (compared with <, >, <=, >=).
// This interface extends ComparableValue and adds ordering operations.
type OrderableValue interface {
	ComparableValue
	// CompareTo compares this value with another value.
	// Returns:
	//   -1 if this < other
	//    0 if this = other
	//    1 if this > other
	// Returns an error if comparison is not possible.
	CompareTo(other Value) (int, error)
}

// CopyableValue represents values that can be copied.
// This interface formalizes copy semantics for value types.
type CopyableValue interface {
	Value
	// Copy creates a deep copy of this value.
	// For value types (integers, strings), this may return the same instance.
	// For reference types (objects, arrays), this creates a new instance with copied data.
	Copy() Value
}

// ReferenceType represents reference-type values (objects, interfaces, arrays, etc.).
// These values have reference semantics - assignment copies the reference, not the data.
// Note: This interface is distinct from the ReferenceValue struct which represents
// a variable reference (var parameter).
type ReferenceType interface {
	Value
	// IsNil checks if this reference is nil.
	IsNil() bool
}

// IndexableValue represents values that can be indexed (arrays, strings).
type IndexableValue interface {
	Value
	// GetIndex retrieves the element at the specified index.
	// Returns the element value and an error if the index is out of bounds.
	GetIndex(index int64) (Value, error)
	// SetIndex sets the element at the specified index.
	// Returns an error if the index is out of bounds or the value is read-only.
	SetIndex(index int64, value Value) error
	// Length returns the number of elements.
	Length() int64
}

// CallableValue represents values that can be called as functions.
type CallableValue interface {
	Value
	// Call invokes the function with the given arguments.
	// Returns the result value and an error if the call fails.
	Call(args []Value) (Value, error)
}

// ConvertibleValue represents values that support explicit type conversion.
type ConvertibleValue interface {
	Value
	// ConvertTo attempts to convert this value to the specified target type.
	// Returns the converted value and an error if conversion is not possible.
	ConvertTo(targetType string) (Value, error)
}

// IterableValue represents values that can be iterated over (arrays, sets, strings).
type IterableValue interface {
	Value
	// Iterator returns an iterator for this value.
	// The iterator can be used in for-in loops.
	Iterator() Iterator
}

// Iterator provides iteration over collection values.
type Iterator interface {
	// Next advances to the next element.
	// Returns false when there are no more elements.
	Next() bool
	// Current returns the current element.
	// Returns nil if Next() hasn't been called or returned false.
	Current() Value
	// Reset resets the iterator to the beginning.
	Reset()
}
