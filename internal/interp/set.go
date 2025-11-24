package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Set Literal Evaluation
// ============================================================================

// evalSetLiteral evaluates a set literal expression.
// Examples: [Red, Blue], [one..five], []
func (i *Interpreter) evalSetLiteral(literal *ast.SetLiteral) Value {
	if literal == nil {
		return &ErrorValue{Message: "nil set literal"}
	}

	// Task 9.156: Check if this SetLiteral should be treated as an array (array of const)
	// This happens when semantic analyzer determined it's used in array context
	var typeAnnot *ast.TypeAnnotation
	if i.semanticInfo != nil {
		typeAnnot = i.semanticInfo.GetType(literal)
	}
	if typeAnnot != nil && typeAnnot.Name != "" {
		// Check if the type is an array type
		resolvedType, err := i.resolveType(typeAnnot.Name)
		if err == nil {
			if _, isArray := resolvedType.(*types.ArrayType); isArray {
				// Evaluate as array literal instead
				arrayLit := &ast.ArrayLiteralExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{Token: literal.Token},
					},
					Elements: literal.Elements,
				}
				// Copy type annotation to array literal in semanticInfo
				if i.semanticInfo != nil {
					i.semanticInfo.SetType(arrayLit, typeAnnot)
				}
				return i.evalArrayLiteral(arrayLit)
			}
		}
	}

	// Task 9.8/9.226: Evaluate all elements and determine the ordinal type
	// Use a temporary map to collect ordinals, then populate the correct storage
	var elementType types.Type
	var enumType *types.EnumType   // For enum-specific handling
	ordinals := make(map[int]bool) // Temporary collection of ordinals
	var lazyRanges []IntRange      // Integer ranges stored without expansion

	for _, elem := range literal.Elements {
		// Check if this is a range expression (e.g., 1..10, 'a'..'z', one..five)
		if rangeExpr, ok := elem.(*ast.RangeExpression); ok {
			// Evaluate range: expand to all values between start and end
			startVal := i.Eval(rangeExpr.Start)
			endVal := i.Eval(rangeExpr.RangeEnd)

			if isError(startVal) {
				return startVal
			}
			if isError(endVal) {
				return endVal
			}

			// Extract ordinal values from start and end
			startOrd, err1 := evaluator.GetOrdinalValue(startVal)
			endOrd, err2 := evaluator.GetOrdinalValue(endVal)

			if err1 != nil {
				return &ErrorValue{
					Message: fmt.Sprintf("range start must be ordinal type: %s at %s",
						err1.Error(), rangeExpr.Start.Pos().String()),
				}
			}
			if err2 != nil {
				return &ErrorValue{
					Message: fmt.Sprintf("range end must be ordinal type: %s at %s",
						err2.Error(), rangeExpr.RangeEnd.Pos().String()),
				}
			}

			// Determine element type from first range
			if elementType == nil {
				// Special handling for enum types
				if enumVal, isEnum := startVal.(*EnumValue); isEnum {
					// Get enum type from environment
					typeVal, ok := i.env.Get("__enum_type_" + strings.ToLower(enumVal.TypeName))
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
					elementType = enumType
				} else {
					// Non-enum ordinal type (Integer, String/Char, Boolean)
					elementType = evaluator.GetOrdinalType(startVal)
				}
			}

			// Verify both endpoints are same type
			if enumVal1, ok1 := startVal.(*EnumValue); ok1 {
				if enumVal2, ok2 := endVal.(*EnumValue); ok2 {
					if enumVal1.TypeName != enumVal2.TypeName {
						return &ErrorValue{
							Message: fmt.Sprintf("range endpoints must be same enum type at %s", rangeExpr.Pos().String()),
						}
					}
				} else {
					return &ErrorValue{
						Message: fmt.Sprintf("range endpoints type mismatch at %s", rangeExpr.Pos().String()),
					}
				}
			}

			// Integer ranges are stored lazily (not expanded)
			// Enum ranges must be expanded for proper set operations
			if enumType == nil {
				// Store as lazy range (integer types only)
				lazyRanges = append(lazyRanges, IntRange{Start: startOrd, End: endOrd})
			} else {
				// Expand enum ranges into ordinals map
				if startOrd <= endOrd {
					for ord := startOrd; ord <= endOrd; ord++ {
						ordinals[ord] = true
					}
				} else {
					for ord := startOrd; ord >= endOrd; ord-- {
						ordinals[ord] = true
					}
				}
			}
		} else {
			// Simple element (not a range)
			elemVal := i.Eval(elem)

			if isError(elemVal) {
				return elemVal
			}

			// Extract ordinal value
			ordinal, err := evaluator.GetOrdinalValue(elemVal)
			if err != nil {
				return &ErrorValue{
					Message: fmt.Sprintf("set element must be ordinal type: %s at %s",
						err.Error(), elem.Pos().String()),
				}
			}

			// Determine element type from first element
			if elementType == nil {
				// Special handling for enum types
				if enumVal, isEnum := elemVal.(*EnumValue); isEnum {
					// Get enum type from environment
					typeVal, ok := i.env.Get("__enum_type_" + strings.ToLower(enumVal.TypeName))
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
					elementType = enumType
				} else {
					// Non-enum ordinal type
					elementType = evaluator.GetOrdinalType(elemVal)
				}
			} else {
				// Verify all elements are of the same type
				if enumType != nil {
					// Expecting enum values
					if enumVal, ok := elemVal.(*EnumValue); ok {
						if enumVal.TypeName != enumType.Name {
							return &ErrorValue{
								Message: fmt.Sprintf("type mismatch in set literal: expected %s, got %s",
									enumType.Name, enumVal.TypeName),
							}
						}
					} else {
						return &ErrorValue{
							Message: fmt.Sprintf("type mismatch in set literal: expected enum %s, got %s",
								enumType.Name, elemVal.Type()),
						}
					}
				}
			}

			// Add element to temporary map
			ordinals[ordinal] = true
		}
	}

	// Handle empty set - if no element type determined, we can't infer it
	if elementType == nil && len(literal.Elements) == 0 {
		// Empty set - try to get type from literal's type annotation
		// For now, return error - empty sets need type context
		return &ErrorValue{
			Message: "cannot infer type for empty set literal",
		}
	}

	// Task 9.8/9.226: Create the SetType (automatically selects storage strategy)
	setType := types.NewSetType(elementType)

	// Task 9.8: Create SetValue and populate the correct storage backend
	setValue := NewSetValue(setType)

	// Populate storage based on strategy
	switch setType.StorageKind {
	case types.SetStorageBitmask:
		// Use bitmask - convert map to bitset
		var elements uint64
		for ordinal := range ordinals {
			if ordinal >= 64 {
				// This shouldn't happen if NewSetType chose bitmask correctly
				return &ErrorValue{
					Message: fmt.Sprintf("enum ordinal %d out of range for bitmask storage", ordinal),
				}
			}
			elements |= (1 << uint(ordinal))
		}
		setValue.Elements = elements

	case types.SetStorageMap:
		// Use map - directly assign the ordinals map
		setValue.MapStore = ordinals
	}

	// Add lazy ranges (for large integer ranges)
	setValue.Ranges = lazyRanges

	return setValue
}

// ============================================================================
// Set Binary Operations
// ============================================================================

// evalBinarySetOperation evaluates binary operations on sets.
// Supported operations: + (union), - (difference), * (intersection)
// Task 9.8: Supports both bitmask and map storage.
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

// ============================================================================
// Set Membership Test
// ============================================================================

// evalSetMembership evaluates the 'in' operator for sets.
// Returns true if the element is in the set, false otherwise.
// Task 9.226: Generalized to accept any ordinal value.
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

	// Check if the element is in the set using the ordinal value
	isInSet := set.HasElement(ordinal)

	return &BooleanValue{Value: isInSet}
}

// ============================================================================
// Include/Exclude Methods
// ============================================================================

// evalSetInclude implements the Include method for sets.
// This mutates the set by adding the element.
// Task 9.226: Generalized to accept any ordinal value.
func (i *Interpreter) evalSetInclude(set *SetValue, element Value) Value {
	if set == nil || element == nil {
		return &ErrorValue{Message: "nil operand in Include"}
	}

	// Extract ordinal value
	ordinal, err := evaluator.GetOrdinalValue(element)
	if err != nil {
		return &ErrorValue{
			Message: fmt.Sprintf("Include requires ordinal value: %s", err.Error()),
		}
	}

	// For enum sets, verify the enum type matches
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

	// Add the element to the set (mutates in place)
	set.AddElement(ordinal)

	return &NilValue{}
}

// evalSetExclude implements the Exclude method for sets.
// This mutates the set by removing the element.
// Task 9.226: Generalized to accept any ordinal value.
func (i *Interpreter) evalSetExclude(set *SetValue, element Value) Value {
	if set == nil || element == nil {
		return &ErrorValue{Message: "nil operand in Exclude"}
	}

	// Extract ordinal value
	ordinal, err := evaluator.GetOrdinalValue(element)
	if err != nil {
		return &ErrorValue{
			Message: fmt.Sprintf("Exclude requires ordinal value: %s", err.Error()),
		}
	}

	// For enum sets, verify the enum type matches
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

	// Remove the element from the set (mutates in place)
	set.RemoveElement(ordinal)

	return &NilValue{}
}

// evalSetDeclaration evaluates a set type declaration.
// The semantic analyzer already registered the set type, so we just acknowledge it.
func (i *Interpreter) evalSetDeclaration(decl *ast.SetDecl) Value {
	if decl == nil {
		return &ErrorValue{Message: "nil set declaration"}
	}

	// Set type already registered by semantic analyzer
	// Just return nil value to indicate successful processing
	return &NilValue{}
}
