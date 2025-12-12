package runtime

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
)

// GetOrdinalValue extracts the ordinal value from an ordinal runtime value.
// Ordinal types include: Integer, Enum, String (single rune), Boolean.
func GetOrdinalValue(val Value) (int, error) {
	switch v := val.(type) {
	case *IntegerValue:
		return int(v.Value), nil
	case *EnumValue:
		return v.OrdinalValue, nil
	case *StringValue:
		runes := []rune(v.Value)
		if len(runes) == 0 {
			return 0, fmt.Errorf("cannot get ordinal value of empty string")
		}
		if len(runes) > 1 {
			return 0, fmt.Errorf("cannot get ordinal value of multi-character string '%s'", v.Value)
		}
		return int(runes[0]), nil
	case *BooleanValue:
		if v.Value {
			return 1, nil
		}
		return 0, nil
	default:
		if val == nil {
			return 0, fmt.Errorf("value of type <nil> is not an ordinal type")
		}
		return 0, fmt.Errorf("value of type %s is not an ordinal type", val.Type())
	}
}

// GetOrdinalType returns the core type used for set element typing.
func GetOrdinalType(val Value) types.Type {
	switch val.(type) {
	case *IntegerValue:
		return types.INTEGER
	case *EnumValue:
		return nil
	case *StringValue:
		return types.STRING
	case *BooleanValue:
		return types.BOOLEAN
	default:
		return nil
	}
}
