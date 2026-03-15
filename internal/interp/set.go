package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// evalBinarySetOperation evaluates binary operations on sets: +, -, *.
// These correspond to union, difference, and intersection respectively.
func (i *Interpreter) evalBinarySetOperation(left, right *SetValue, operator string) Value {
	if left == nil || right == nil {
		return &ErrorValue{Message: "nil set operand"}
	}

	// Verify both sets are of the same type
	if !left.SetType.Equals(right.SetType) {
		return &ErrorValue{
			Message: fmt.Sprintf("type mismatch in set operation: %s vs %s",
				left.SetType.String(), right.SetType.String()),
		}
	}

	// Create result set with same type
	result := NewSetValue(left.SetType)

	// Choose operation based on storage kind
	switch left.SetType.StorageKind {
	case types.SetStorageBitmask:
		// Fast bitwise operations for bitmask storage
		var resultElements uint64

		switch operator {
		case "+":
			// Union: bitwise OR
			resultElements = left.Elements | right.Elements

		case "-":
			// Difference: bitwise AND NOT
			resultElements = left.Elements &^ right.Elements

		case "*":
			// Intersection: bitwise AND
			resultElements = left.Elements & right.Elements

		default:
			return &ErrorValue{
				Message: fmt.Sprintf("unsupported set operation: %s", operator),
			}
		}

		result.Elements = resultElements

	case types.SetStorageMap:
		// Map-based operations for large sets
		switch operator {
		case "+":
			// Union: add all elements from both sets
			for ordinal := range left.MapStore {
				result.MapStore[ordinal] = true
			}
			for ordinal := range right.MapStore {
				result.MapStore[ordinal] = true
			}

		case "-":
			// Difference: elements in left but not in right
			for ordinal := range left.MapStore {
				if !right.MapStore[ordinal] {
					result.MapStore[ordinal] = true
				}
			}

		case "*":
			// Intersection: elements in both sets
			for ordinal := range left.MapStore {
				if right.MapStore[ordinal] {
					result.MapStore[ordinal] = true
				}
			}

		default:
			return &ErrorValue{
				Message: fmt.Sprintf("unsupported set operation: %s", operator),
			}
		}
	}

	return result
}

// evalSetMembership evaluates the 'in' operator for sets.
// Returns true if the element is in the set, false otherwise.
func (i *Interpreter) evalSetMembership(element Value, ordinal int, set *SetValue) Value {
	if element == nil || set == nil {
		return &ErrorValue{Message: "nil operand in membership test"}
	}

	// Type checking is done by semantic analyzer and GetOrdinalValue
	// Just verify element type is compatible with set's element type
	// For enum sets, verify the enum type matches
	if enumVal, ok := element.(*EnumValue); ok {
		if enumType, ok := set.SetType.ElementType.(*types.EnumType); ok {
			if enumVal.TypeName != enumType.Name {
				return &ErrorValue{
					Message: fmt.Sprintf("type mismatch: enum %s not in set of %s",
						enumVal.TypeName, enumType.Name),
				}
			}
		}
	}

	isInSet := set.HasElement(ordinal)
	return &BooleanValue{Value: isInSet}
}

// evalSetInclude implements the Include method for sets.
// This mutates the set by adding the element.
func (i *Interpreter) evalSetInclude(set *SetValue, element Value) Value {
	if set == nil || element == nil {
		return &ErrorValue{Message: "nil operand in Include"}
	}

	ordinal, err := runtime.GetOrdinalValue(element)
	if err != nil {
		return &ErrorValue{
			Message: fmt.Sprintf("Include requires ordinal value: %s", err.Error()),
		}
	}

	if enumVal, ok := element.(*EnumValue); ok {
		if enumType, ok := set.SetType.ElementType.(*types.EnumType); ok {
			if enumVal.TypeName != enumType.Name {
				return &ErrorValue{
					Message: fmt.Sprintf("type mismatch: cannot add enum %s to set of %s",
						enumVal.TypeName, enumType.Name),
				}
			}
		}
	}

	set.AddElement(ordinal)
	return &NilValue{}
}

// evalSetExclude implements the Exclude method for sets.
// This mutates the set by removing the element.
func (i *Interpreter) evalSetExclude(set *SetValue, element Value) Value {
	if set == nil || element == nil {
		return &ErrorValue{Message: "nil operand in Exclude"}
	}

	ordinal, err := runtime.GetOrdinalValue(element)
	if err != nil {
		return &ErrorValue{
			Message: fmt.Sprintf("Exclude requires ordinal value: %s", err.Error()),
		}
	}

	if enumVal, ok := element.(*EnumValue); ok {
		if enumType, ok := set.SetType.ElementType.(*types.EnumType); ok {
			if enumVal.TypeName != enumType.Name {
				return &ErrorValue{
					Message: fmt.Sprintf("type mismatch: cannot remove enum %s from set of %s",
						enumVal.TypeName, enumType.Name),
				}
			}
		}
	}

	set.RemoveElement(ordinal)
	return &NilValue{}
}
