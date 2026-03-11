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

// EnumTypeResolver resolves enum type metadata by DWScript type name.
type EnumTypeResolver func(typeName string) (*types.EnumType, error)

// RebuildOrdinalValue recreates a runtime value of the same ordinal kind as the template.
// This keeps ordinal-to-value reconstruction centralized instead of duplicating it in
// evaluator execution paths.
func RebuildOrdinalValue(template Value, ordinal int, resolveEnum EnumTypeResolver) (Value, error) {
	switch v := template.(type) {
	case *IntegerValue:
		return &IntegerValue{Value: int64(ordinal)}, nil
	case *EnumValue:
		if resolveEnum == nil {
			return nil, fmt.Errorf("enum ordinal reconstruction requires enum type resolver")
		}
		enumType, err := resolveEnum(v.TypeName)
		if err != nil {
			return nil, err
		}
		return NewEnumValue(v.TypeName, enumType, ordinal), nil
	case *StringValue:
		return &StringValue{Value: string(rune(ordinal))}, nil
	case *BooleanValue:
		return &BooleanValue{Value: ordinal != 0}, nil
	default:
		if template == nil {
			return nil, fmt.Errorf("unsupported ordinal loop variable type <nil>")
		}
		return nil, fmt.Errorf("unsupported ordinal loop variable type %s", template.Type())
	}
}
