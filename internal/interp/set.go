package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Set Literal Evaluation (Task 8.106-8.107)
// ============================================================================

// evalSetLiteral evaluates a set literal expression.
// Examples: [Red, Blue], [one..five], []
func (i *Interpreter) evalSetLiteral(literal *ast.SetLiteral) Value {
	if literal == nil {
		return &ErrorValue{Message: "nil set literal"}
	}

	// Evaluate all elements and determine the enum type
	var enumType *types.EnumType
	var elements uint64 // Bitset for the set

	for _, elem := range literal.Elements {
		// Check if this is a range expression (one..five)
		if rangeExpr, ok := elem.(*ast.RangeExpression); ok {
			// Evaluate range: expand to all values between start and end
			startVal := i.Eval(rangeExpr.Start)
			endVal := i.Eval(rangeExpr.End)

			// Both must be enum values
			startEnum, ok1 := startVal.(*EnumValue)
			endEnum, ok2 := endVal.(*EnumValue)

			if !ok1 || !ok2 {
				return &ErrorValue{
					Message: fmt.Sprintf("range endpoints must be enum values at %s", rangeExpr.Pos().String()),
				}
			}

			// Types must match
			if startEnum.TypeName != endEnum.TypeName {
				return &ErrorValue{
					Message: fmt.Sprintf("range endpoints must be of same type at %s", rangeExpr.Pos().String()),
				}
			}

			// Set enum type if not set yet
			if enumType == nil {
				// Get enum type from environment
				typeVal, ok := i.env.Get("__enum_type_" + startEnum.TypeName)
				if !ok {
					return &ErrorValue{
						Message: fmt.Sprintf("unknown enum type '%s'", startEnum.TypeName),
					}
				}
				enumTypeVal, ok := typeVal.(*EnumTypeValue)
				if !ok {
					return &ErrorValue{
						Message: fmt.Sprintf("invalid enum type for '%s'", startEnum.TypeName),
					}
				}
				enumType = enumTypeVal.EnumType
			}

			// Add all values in the range [start..end] inclusive
			startOrd := startEnum.OrdinalValue
			endOrd := endEnum.OrdinalValue

			// Handle both forward and reverse ranges
			if startOrd <= endOrd {
				for ord := startOrd; ord <= endOrd; ord++ {
					elements |= (1 << uint(ord))
				}
			} else {
				for ord := startOrd; ord >= endOrd; ord-- {
					elements |= (1 << uint(ord))
				}
			}
		} else {
			// Simple element (not a range)
			elemVal := i.Eval(elem)

			// Must be an enum value
			enumVal, ok := elemVal.(*EnumValue)
			if !ok {
				return &ErrorValue{
					Message: fmt.Sprintf("set element must be an enum value, got %s", elemVal.Type()),
				}
			}

			// Set enum type if not set yet (from first element)
			if enumType == nil {
				// Get enum type from environment
				typeVal, ok := i.env.Get("__enum_type_" + enumVal.TypeName)
				if !ok {
					return &ErrorValue{
						Message: fmt.Sprintf("unknown enum type '%s'", enumVal.TypeName),
					}
				}
				enumTypeVal, ok := typeVal.(*EnumTypeValue)
				if !ok {
					return &ErrorValue{
						Message: fmt.Sprintf("invalid enum type for '%s'", enumVal.TypeName),
					}
				}
				enumType = enumTypeVal.EnumType
			} else {
				// Verify all elements are of the same enum type
				if enumVal.TypeName != enumType.Name {
					return &ErrorValue{
						Message: fmt.Sprintf("type mismatch in set literal: expected %s, got %s",
							enumType.Name, enumVal.TypeName),
					}
				}
			}

			// Add element to bitset
			ordinal := enumVal.OrdinalValue
			if ordinal < 0 || ordinal >= 64 {
				return &ErrorValue{
					Message: fmt.Sprintf("enum ordinal %d out of range for bitset", ordinal),
				}
			}
			elements |= (1 << uint(ordinal))
		}
	}

	// Handle empty set - if no enum type determined, we can't infer it
	// For now, create an empty set with nil type (will be inferred from context)
	if enumType == nil && len(literal.Elements) == 0 {
		// Empty set - try to get type from literal's type annotation
		// For now, return error - empty sets need type context
		return &ErrorValue{
			Message: "cannot infer type for empty set literal",
		}
	}

	// Create the SetType
	setType := types.NewSetType(enumType)

	// Create and return the SetValue
	return &SetValue{
		SetType:  setType,
		Elements: elements,
	}
}

// ============================================================================
// Set Binary Operations (Tasks 8.110-8.112)
// ============================================================================

// evalBinarySetOperation evaluates binary operations on sets.
// Supported operations: + (union), - (difference), * (intersection)
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

	// Return new SetValue with result
	return &SetValue{
		SetType:  left.SetType,
		Elements: resultElements,
	}
}

// ============================================================================
// Set Membership Test (Task 8.113)
// ============================================================================

// evalSetMembership evaluates the 'in' operator for sets.
// Returns true if the element is in the set, false otherwise.
func (i *Interpreter) evalSetMembership(element *EnumValue, set *SetValue) Value {
	if element == nil || set == nil {
		return &ErrorValue{Message: "nil operand in membership test"}
	}

	// Verify the element type matches the set's element type
	if element.TypeName != set.SetType.ElementType.Name {
		return &ErrorValue{
			Message: fmt.Sprintf("type mismatch: %s not in set of %s",
				element.TypeName, set.SetType.ElementType.Name),
		}
	}

	// Check if the element is in the set
	isInSet := set.HasElement(element.OrdinalValue)

	return &BooleanValue{Value: isInSet}
}

// ============================================================================
// Include/Exclude Methods (Tasks 8.108-8.109)
// ============================================================================

// evalSetInclude implements the Include method for sets.
// This mutates the set by adding the element.
func (i *Interpreter) evalSetInclude(set *SetValue, element *EnumValue) Value {
	if set == nil || element == nil {
		return &ErrorValue{Message: "nil operand in Include"}
	}

	// Verify the element type matches the set's element type
	if element.TypeName != set.SetType.ElementType.Name {
		return &ErrorValue{
			Message: fmt.Sprintf("type mismatch: cannot add %s to set of %s",
				element.TypeName, set.SetType.ElementType.Name),
		}
	}

	// Add the element to the set (mutates in place)
	set.AddElement(element.OrdinalValue)

	return &NilValue{}
}

// evalSetExclude implements the Exclude method for sets.
// This mutates the set by removing the element.
func (i *Interpreter) evalSetExclude(set *SetValue, element *EnumValue) Value {
	if set == nil || element == nil {
		return &ErrorValue{Message: "nil operand in Exclude"}
	}

	// Verify the element type matches the set's element type
	if element.TypeName != set.SetType.ElementType.Name {
		return &ErrorValue{
			Message: fmt.Sprintf("type mismatch: cannot remove %s from set of %s",
				element.TypeName, set.SetType.ElementType.Name),
		}
	}

	// Remove the element from the set (mutates in place)
	set.RemoveElement(element.OrdinalValue)

	return &NilValue{}
}
