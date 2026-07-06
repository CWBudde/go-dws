package types

import "fmt"

// ============================================================================
// AssociativeArrayType
// ============================================================================

// AssociativeArrayType represents a DWScript associative array (map/dictionary):
//
//	array [KeyType] of ElementType
//
// Unlike a static or enum-indexed ArrayType, the key is an arbitrary type
// (Integer, String, Float, a record, an object, a static array) rather than a
// bounded ordinal, and storage is sparse. Associative arrays are reference
// types (assignment shares the backing map, like dynamic arrays).
//
// It deliberately uses a distinct TypeKind ("ASSOCIATIVE_ARRAY") so that code
// which handles the dense *ArrayType does not silently mistreat an associative
// array as a sliceable dynamic array.
type AssociativeArrayType struct {
	KeyType     Type // Type of the keys
	ElementType Type // Type of the values
}

// NewAssociativeArrayType creates a new associative array type.
func NewAssociativeArrayType(keyType, elementType Type) *AssociativeArrayType {
	return &AssociativeArrayType{KeyType: keyType, ElementType: elementType}
}

// String returns a string representation, e.g. "array [String] of Integer".
func (at *AssociativeArrayType) String() string {
	return fmt.Sprintf("array [%s] of %s", at.KeyType.String(), at.ElementType.String())
}

// TypeKind returns "ASSOCIATIVE_ARRAY".
func (at *AssociativeArrayType) TypeKind() string { return "ASSOCIATIVE_ARRAY" }

// Equals reports whether other is an associative array with the same key and
// element types.
func (at *AssociativeArrayType) Equals(other Type) bool {
	o, ok := other.(*AssociativeArrayType)
	if !ok {
		return false
	}
	if (at.KeyType == nil) != (o.KeyType == nil) {
		return false
	}
	if at.KeyType != nil && !at.KeyType.Equals(o.KeyType) {
		return false
	}
	if (at.ElementType == nil) != (o.ElementType == nil) {
		return false
	}
	if at.ElementType != nil && !at.ElementType.Equals(o.ElementType) {
		return false
	}
	return true
}
