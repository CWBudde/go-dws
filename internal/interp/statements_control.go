package interp

// This file contains shared comparison and range helpers used across the
// interpreter shell while statement execution lives in evaluator.

// valuesEqual compares two values for equality.
// This is used by case statements to match values.
func (i *Interpreter) valuesEqual(left, right Value) bool {
	left, right = i.unwrapVariants(left, right)

	// Handle nil values (uninitialized variants)
	if left == nil && right == nil {
		return true // Both uninitialized variants are equal
	}
	if left == nil || right == nil {
		return false // One is nil, the other is not
	}

	// Handle same type comparisons
	if left.Type() != right.Type() {
		return false
	}

	return i.valuesEqualTyped(left, right)
}

func (i *Interpreter) unwrapVariants(left, right Value) (Value, Value) {
	// Unwrap VariantValue if present
	if varVal, ok := left.(*VariantValue); ok {
		left = varVal.Value
	}
	if varVal, ok := right.(*VariantValue); ok {
		right = varVal.Value
	}
	return left, right
}

func (i *Interpreter) valuesEqualTyped(left, right Value) bool {
	switch l := left.(type) {
	case *IntegerValue:
		r, ok := right.(*IntegerValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *FloatValue:
		r, ok := right.(*FloatValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *StringValue:
		r, ok := right.(*StringValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *BooleanValue:
		r, ok := right.(*BooleanValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *NilValue:
		return true // nil == nil
	case *RecordValue:
		r, ok := right.(*RecordValue)
		if !ok {
			return false
		}
		return i.recordsEqual(l, r)
	default:
		// For other types, use string comparison as fallback
		return left.String() == right.String()
	}
}

// isInRange checks if value is within the range [start, end] inclusive.
// Supports Integer, Float, String (character), and Enum values.
func (i *Interpreter) isInRange(value, start, end Value) bool {
	value, start, end = unwrapVariantsForRange(value, start, end)

	// Handle nil values (uninitialized variants)
	if value == nil || start == nil || end == nil {
		return false // Cannot perform range check with uninitialized variants
	}

	// Handle different value types
	switch v := value.(type) {
	case *IntegerValue:
		return i.isInRangeInteger(v, start, end)
	case *FloatValue:
		return i.isInRangeFloat(v, start, end)
	case *StringValue:
		return i.isInRangeString(v, start, end)
	case *EnumValue:
		return i.isInRangeEnum(v, start, end)
	}

	return false
}

func unwrapVariantsForRange(value, start, end Value) (Value, Value, Value) {
	// Unwrap VariantValue if present
	if varVal, ok := value.(*VariantValue); ok {
		value = varVal.Value
	}
	if varVal, ok := start.(*VariantValue); ok {
		start = varVal.Value
	}
	if varVal, ok := end.(*VariantValue); ok {
		end = varVal.Value
	}
	return value, start, end
}

func (i *Interpreter) isInRangeInteger(v *IntegerValue, start, end Value) bool {
	startInt, startOk := start.(*IntegerValue)
	endInt, endOk := end.(*IntegerValue)
	if startOk && endOk {
		return v.Value >= startInt.Value && v.Value <= endInt.Value
	}
	return false
}

func (i *Interpreter) isInRangeFloat(v *FloatValue, start, end Value) bool {
	startFloat, startOk := start.(*FloatValue)
	endFloat, endOk := end.(*FloatValue)
	if startOk && endOk {
		return v.Value >= startFloat.Value && v.Value <= endFloat.Value
	}
	return false
}

func (i *Interpreter) isInRangeString(v *StringValue, start, end Value) bool {
	// For strings, compare character by character
	startStr, startOk := start.(*StringValue)
	endStr, endOk := end.(*StringValue)
	// Use rune-based comparison to handle UTF-8 correctly
	if startOk && endOk && runeLength(v.Value) == 1 && runeLength(startStr.Value) == 1 && runeLength(endStr.Value) == 1 {
		// Single character comparison (for 'A'..'Z' style ranges)
		charVal, _ := runeAt(v.Value, 1)
		charStart, _ := runeAt(startStr.Value, 1)
		charEnd, _ := runeAt(endStr.Value, 1)
		return charVal >= charStart && charVal <= charEnd
	}
	// Fall back to string comparison for multi-char strings
	if startOk && endOk {
		return v.Value >= startStr.Value && v.Value <= endStr.Value
	}
	return false
}

func (i *Interpreter) isInRangeEnum(v *EnumValue, start, end Value) bool {
	startEnum, startOk := start.(*EnumValue)
	endEnum, endOk := end.(*EnumValue)
	if startOk && endOk && v.TypeName == startEnum.TypeName && v.TypeName == endEnum.TypeName {
		return v.OrdinalValue >= startEnum.OrdinalValue && v.OrdinalValue <= endEnum.OrdinalValue
	}
	return false
}
