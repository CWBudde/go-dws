package evaluator

import (
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// Type Conversion Methods (Task 3.5.143g, 3.5.143i)
// ============================================================================
//
// This file implements the type conversion and introspection methods of the
// builtins.Context interface for the Evaluator:
// - UnwrapVariant(): Unwrap Variant values
// - ToInt64(): Convert values to int64
// - ToBool(): Convert values to bool
// - ToFloat64(): Convert values to float64
// - GetTypeOf(): Get type name of a value
// - GetClassOf(): Get class name of an object value
// - ParseInt(): Parse string to int64 (Task 3.5.143i)
// - ParseFloat(): Parse string to float64 (Task 3.5.143i)
//
// These methods enable built-in functions to perform type conversions and
// introspection on DWScript values.
//
// Note: For types defined in the interp package (SubrangeValue, ObjectInstance,
// ClassValue, ClassInfoValue), we delegate to the adapter since the evaluator
// package cannot directly import interp (circular dependency).
// ============================================================================

// UnwrapVariant returns the underlying value if input is a Variant, otherwise returns input as-is.
// This allows built-in functions to work with both direct values and Variant-wrapped values.
//
// This implements the builtins.Context interface method UnwrapVariant().
func (e *Evaluator) UnwrapVariant(value Value) Value {
	if value != nil {
		// Check if it's a VariantValue and unwrap it
		if variant, ok := value.(*runtime.VariantValue); ok {
			if variant.Value == nil {
				return &runtime.UnassignedValue{}
			}
			return variant.Value
		}
	}
	return value
}

// ToInt64 converts a Value to int64, handling SubrangeValue and EnumValue.
// Returns the integer value and true if successful, 0 and false otherwise.
//
// This implements the builtins.Context interface method ToInt64().
func (e *Evaluator) ToInt64(value Value) (int64, bool) {
	// Simple implementations for types available in runtime package
	switch v := value.(type) {
	case *runtime.IntegerValue:
		return v.Value, true
	case *runtime.EnumValue:
		return int64(v.OrdinalValue), true
	case *runtime.BooleanValue:
		if v.Value {
			return 1, true
		}
		return 0, true
	case *runtime.FloatValue:
		return int64(v.Value), true
	default:
		// For SubrangeValue and other interp types, we'd need the adapter
		// But since we're implementing builtins.Context which takes runtime.Value,
		// we shouldn't receive those types here. Return false.
		return 0, false
	}
}

// ToBool converts a Value to bool.
// Returns the boolean value and true if successful, false and false otherwise.
//
// This implements the builtins.Context interface method ToBool().
func (e *Evaluator) ToBool(value Value) (bool, bool) {
	switch v := value.(type) {
	case *runtime.BooleanValue:
		return v.Value, true
	case *runtime.IntegerValue:
		return v.Value != 0, true
	case *runtime.EnumValue:
		return v.OrdinalValue != 0, true
	default:
		return false, false
	}
}

// ToFloat64 converts a Value to float64, handling integer types.
// Returns the float value and true if successful, 0.0 and false otherwise.
//
// This implements the builtins.Context interface method ToFloat64().
func (e *Evaluator) ToFloat64(value Value) (float64, bool) {
	switch v := value.(type) {
	case *runtime.FloatValue:
		return v.Value, true
	case *runtime.IntegerValue:
		return float64(v.Value), true
	case *runtime.EnumValue:
		return float64(v.OrdinalValue), true
	default:
		return 0.0, false
	}
}

// GetTypeOf returns the type name of a value.
// For objects, it returns the class name. For primitives, it returns the type name.
//
// This implements the builtins.Context interface method GetTypeOf().
func (e *Evaluator) GetTypeOf(value Value) string {
	if value == nil {
		return "Null"
	}

	// Check for ObjectValue interface (objects implement ClassName())
	if objVal, ok := value.(ObjectValue); ok {
		return objVal.ClassName()
	}

	// Check for ClassMetaValue interface (class references)
	if classVal, ok := value.(ClassMetaValue); ok {
		return classVal.GetClassName()
	}

	typeName := value.Type()

	// Convert internal type names to DWScript format
	switch typeName {
	case "INTEGER":
		return "Integer"
	case "FLOAT":
		return "Float"
	case "STRING":
		return "String"
	case "BOOLEAN":
		return "Boolean"
	case "NIL", "NULL":
		return "Null"
	case "ARRAY":
		return "Array"
	case "RECORD":
		return "Record"
	default:
		return typeName
	}
}

// GetClassOf returns the class name of an object value.
// Returns empty string if the value is not an object.
//
// This implements the builtins.Context interface method GetClassOf().
func (e *Evaluator) GetClassOf(value Value) string {
	// Check for ObjectValue interface (objects implement ClassName())
	if objVal, ok := value.(ObjectValue); ok {
		return objVal.ClassName()
	}

	// Check for ClassMetaValue interface (class references)
	if classVal, ok := value.(ClassMetaValue); ok {
		return classVal.GetClassName()
	}

	return ""
}

// ============================================================================
// String Parsing Methods (Task 3.5.143i)
// ============================================================================

// ParseInt parses a string to an int64 with the specified base.
// Returns the parsed value and true if successful, 0 and false otherwise.
//
// This implements the builtins.Context interface method ParseInt().
func (e *Evaluator) ParseInt(s string, base int) (int64, bool) {
	// Trim whitespace for lenient parsing
	s = strings.TrimSpace(s)

	// Use strconv.ParseInt for parsing
	intValue, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		return 0, false
	}

	return intValue, true
}

// ParseFloat parses a string to a float64.
// Returns the parsed value and true if successful, 0.0 and false otherwise.
//
// This implements the builtins.Context interface method ParseFloat().
func (e *Evaluator) ParseFloat(s string) (float64, bool) {
	// Trim whitespace for lenient parsing
	s = strings.TrimSpace(s)

	// Use strconv.ParseFloat for parsing
	floatValue, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, false
	}

	return floatValue, true
}
