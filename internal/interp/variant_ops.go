package interp

// variant_ops.go contains Variant-specific binary operations.

// convertToString converts a Value to its string representation.
// Used for Variant string concatenation and comparison.
func (i *Interpreter) convertToString(val Value) string {
	if val == nil {
		return ""
	}
	return val.String()
}

// isFalsey checks if a value is considered "falsey" (default/zero value for its type).
// Falsey values: 0 (integer), 0.0 (float), "" (empty string), false (boolean), nil, empty arrays.
func isFalsey(val Value) bool {
	if val == nil {
		return true
	}

	switch v := val.(type) {
	case *IntegerValue:
		return v.Value == 0
	case *FloatValue:
		return v.Value == 0.0
	case *StringValue:
		return v.Value == ""
	case *BooleanValue:
		return !v.Value
	case *NilValue:
		return true
	case *NullValue:
		return true
	case *UnassignedValue:
		return true
	case *ArrayValue:
		return len(v.Elements) == 0
	case *VariantValue:
		// Unwrap variant and check inner value
		return isFalsey(v.Value)
	default:
		// Other types (objects, classes, etc.) are truthy if non-nil
		return false
	}
}

// isNumericType checks if a type is numeric (INTEGER or FLOAT).
func isNumericType(typeStr string) bool {
	return typeStr == "INTEGER" || typeStr == "FLOAT"
}
