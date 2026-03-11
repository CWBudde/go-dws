package evaluator

import (
	"fmt"
	"math"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Bounds Operation Methods
// ============================================================================
//
// This file implements the bounds operation methods of the builtins.Context
// interface for the Evaluator:
// - GetLowBound(): Get lower bound of array, enum, string, or type meta-value
// - GetHighBound(): Get upper bound of array, enum, string, or type meta-value
//
// Polymorphic behavior:
// - Arrays: Return array bounds from ArrayType or dynamic bounds
// - Enums: Return first/last enum value as bound
// - Type meta-values: Return bounds for built-in types or enum bounds
// - Strings: 1-indexed (Low=1, High=Length)
//
// These implementations are self-contained and do not require callbacks
// to the interpreter.
//
// ============================================================================

// GetLowBound returns the lower bound for arrays, enums, strings, or type meta-values.
// This implements the builtins.Context interface.
func (e *Evaluator) GetLowBound(value Value) (Value, error) {
	// Type meta-values
	if typeMetaVal, ok := value.(*runtime.TypeMetaValue); ok {
		switch typeMetaVal.TypeInfo {
		case types.INTEGER:
			return &runtime.IntegerValue{Value: math.MinInt64}, nil
		case types.FLOAT:
			return &runtime.FloatValue{Value: -math.MaxFloat64}, nil
		case types.BOOLEAN:
			return &runtime.BooleanValue{Value: false}, nil
		}
		if enumType, ok := typeMetaVal.TypeInfo.(*types.EnumType); ok {
			if len(enumType.OrderedNames) == 0 {
				return nil, fmt.Errorf("enum type '%s' has no values", typeMetaVal.TypeName)
			}
			return runtime.EnumValueAtIndex(typeMetaVal.TypeName, enumType, 0)
		}
		return nil, fmt.Errorf("Low() not supported for type %s", typeMetaVal.TypeName)
	}

	// Arrays
	if arrayVal, ok := value.(*runtime.ArrayValue); ok {
		if arrayVal.ArrayType != nil && arrayVal.ArrayType.IsStatic() {
			return &runtime.IntegerValue{Value: int64(*arrayVal.ArrayType.LowBound)}, nil
		}
		return &runtime.IntegerValue{Value: 0}, nil
	}

	// Enum values
	if enumVal, ok := value.(*runtime.EnumValue); ok {
		enumMetadata := e.typeSystem.LookupEnumMetadata(enumVal.TypeName)
		if enumMetadata == nil {
			return nil, fmt.Errorf("enum type '%s' not found", enumVal.TypeName)
		}
		etv, ok := enumMetadata.(EnumTypeValueAccessor)
		if !ok {
			return nil, fmt.Errorf("invalid enum type metadata for '%s'", enumVal.TypeName)
		}
		enumType := etv.GetEnumType()
		if len(enumType.OrderedNames) == 0 {
			return nil, fmt.Errorf("enum type '%s' has no values", enumVal.TypeName)
		}
		return runtime.EnumValueAtIndex(enumVal.TypeName, enumType, 0)
	}

	// Strings are 1-indexed in DWScript
	if _, ok := value.(*runtime.StringValue); ok {
		return &runtime.IntegerValue{Value: 1}, nil
	}

	return nil, fmt.Errorf("Low() expects array, enum, string, or type name, got %s", value.Type())
}

// GetHighBound returns the upper bound for arrays, enums, strings, or type meta-values.
// This implements the builtins.Context interface.
func (e *Evaluator) GetHighBound(value Value) (Value, error) {
	// Type meta-values
	if typeMetaVal, ok := value.(*runtime.TypeMetaValue); ok {
		switch typeMetaVal.TypeInfo {
		case types.INTEGER:
			return &runtime.IntegerValue{Value: math.MaxInt64}, nil
		case types.FLOAT:
			return &runtime.FloatValue{Value: math.MaxFloat64}, nil
		case types.BOOLEAN:
			return &runtime.BooleanValue{Value: true}, nil
		}
		if enumType, ok := typeMetaVal.TypeInfo.(*types.EnumType); ok {
			if len(enumType.OrderedNames) == 0 {
				return nil, fmt.Errorf("enum type '%s' has no values", typeMetaVal.TypeName)
			}
			return runtime.EnumValueAtIndex(typeMetaVal.TypeName, enumType, len(enumType.OrderedNames)-1)
		}
		return nil, fmt.Errorf("High() not supported for type %s", typeMetaVal.TypeName)
	}

	// Arrays
	if arrayVal, ok := value.(*runtime.ArrayValue); ok {
		if arrayVal.ArrayType != nil && arrayVal.ArrayType.IsStatic() {
			return &runtime.IntegerValue{Value: int64(*arrayVal.ArrayType.HighBound)}, nil
		}
		return &runtime.IntegerValue{Value: int64(len(arrayVal.Elements) - 1)}, nil
	}

	// Enum values
	if enumVal, ok := value.(*runtime.EnumValue); ok {
		enumMetadata := e.typeSystem.LookupEnumMetadata(enumVal.TypeName)
		if enumMetadata == nil {
			return nil, fmt.Errorf("enum type '%s' not found", enumVal.TypeName)
		}
		etv, ok := enumMetadata.(EnumTypeValueAccessor)
		if !ok {
			return nil, fmt.Errorf("invalid enum type metadata for '%s'", enumVal.TypeName)
		}
		enumType := etv.GetEnumType()
		if len(enumType.OrderedNames) == 0 {
			return nil, fmt.Errorf("enum type '%s' has no values", enumVal.TypeName)
		}
		return runtime.EnumValueAtIndex(enumVal.TypeName, enumType, len(enumType.OrderedNames)-1)
	}

	// Strings: High(s) = Length(s)
	if strVal, ok := value.(*runtime.StringValue); ok {
		return &runtime.IntegerValue{Value: int64(len([]rune(strVal.Value)))}, nil
	}

	return nil, fmt.Errorf("High() expects array, enum, string, or type name, got %s", value.Type())
}
