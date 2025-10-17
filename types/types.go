package types

import "fmt"

// Type represents a DWScript type at compile-time.
// This is used for static type checking during semantic analysis,
// separate from runtime Value representations.
type Type interface {
	// String returns the string representation of the type (e.g., "Integer", "String")
	String() string

	// Equals checks if two types are identical
	Equals(other Type) bool

	// TypeKind returns a unique identifier for the type kind
	// Used for type discrimination without reflection
	TypeKind() string
}

// ============================================================================
// Basic Types
// ============================================================================

// IntegerType represents the Integer type
type IntegerType struct{}

func (t *IntegerType) String() string   { return "Integer" }
func (t *IntegerType) TypeKind() string { return "INTEGER" }
func (t *IntegerType) Equals(other Type) bool {
	_, ok := other.(*IntegerType)
	return ok
}

// FloatType represents the Float type
type FloatType struct{}

func (t *FloatType) String() string   { return "Float" }
func (t *FloatType) TypeKind() string { return "FLOAT" }
func (t *FloatType) Equals(other Type) bool {
	_, ok := other.(*FloatType)
	return ok
}

// StringType represents the String type
type StringType struct{}

func (t *StringType) String() string   { return "String" }
func (t *StringType) TypeKind() string { return "STRING" }
func (t *StringType) Equals(other Type) bool {
	_, ok := other.(*StringType)
	return ok
}

// BooleanType represents the Boolean type
type BooleanType struct{}

func (t *BooleanType) String() string   { return "Boolean" }
func (t *BooleanType) TypeKind() string { return "BOOLEAN" }
func (t *BooleanType) Equals(other Type) bool {
	_, ok := other.(*BooleanType)
	return ok
}

// NilType represents the nil/null type
type NilType struct{}

func (t *NilType) String() string   { return "Nil" }
func (t *NilType) TypeKind() string { return "NIL" }
func (t *NilType) Equals(other Type) bool {
	_, ok := other.(*NilType)
	return ok
}

// VoidType represents the void type (for procedures with no return value)
type VoidType struct{}

func (t *VoidType) String() string   { return "Void" }
func (t *VoidType) TypeKind() string { return "VOID" }
func (t *VoidType) Equals(other Type) bool {
	_, ok := other.(*VoidType)
	return ok
}

// ============================================================================
// Singleton Type Constants
// ============================================================================

// Singleton instances of basic types
// These are used throughout the compiler for type checking
var (
	INTEGER = &IntegerType{}
	FLOAT   = &FloatType{}
	STRING  = &StringType{}
	BOOLEAN = &BooleanType{}
	NIL     = &NilType{}
	VOID    = &VoidType{}
)

// ============================================================================
// Type Utilities
// ============================================================================

// IsBasicType checks if a type is one of the basic types
func IsBasicType(t Type) bool {
	switch t.TypeKind() {
	case "INTEGER", "FLOAT", "STRING", "BOOLEAN":
		return true
	default:
		return false
	}
}

// IsNumericType checks if a type is numeric (Integer or Float)
func IsNumericType(t Type) bool {
	kind := t.TypeKind()
	return kind == "INTEGER" || kind == "FLOAT"
}

// IsOrdinalType checks if a type is an ordinal type (used for loop variables)
// In DWScript, ordinal types are Integer, Boolean, and enumerations
func IsOrdinalType(t Type) bool {
	kind := t.TypeKind()
	return kind == "INTEGER" || kind == "BOOLEAN"
}

// TypeFromString converts a type name string to a Type
// This is useful for parsing type annotations
func TypeFromString(name string) (Type, error) {
	switch name {
	case "Integer":
		return INTEGER, nil
	case "Float":
		return FLOAT, nil
	case "String":
		return STRING, nil
	case "Boolean":
		return BOOLEAN, nil
	case "Void":
		return VOID, nil
	default:
		return nil, fmt.Errorf("unknown type: %s", name)
	}
}
