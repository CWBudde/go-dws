package runtime

// IsFalsey determines if a value is "falsey" (default/zero value for its type).
func IsFalsey(val Value) bool {
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
	case *ArrayValue:
		return len(v.Elements) == 0
	default:
		switch val.Type() {
		case "NIL", "UNASSIGNED", "NULL":
			return true
		case "VARIANT":
			if wrapper, ok := val.(VariantWrapper); ok {
				return IsFalsey(wrapper.UnwrapVariant())
			}
			return false
		default:
			return false
		}
	}
}
