package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Set Literal Evaluation
// ============================================================================
//
// Task 3.5.80: Direct set literal evaluation in the evaluator.
// Migrated from interp/set.go to reduce adapter dependency.
// ============================================================================

// evalSetLiteralDirect evaluates a set literal expression directly.
// Examples: [Red, Blue], [one..five], []
//
// This handles:
//   - Simple elements: [Red, Blue, Green]
//   - Range elements: [1..10], [one..five]
//   - Mixed elements: [1, 5..10, 20]
//   - Empty sets: [] (requires type context)
//
// Returns:
//   - A *runtime.SetValue on success
//   - An error Value if evaluation fails
func (e *Evaluator) evalSetLiteralDirect(node *ast.SetLiteral, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil set literal")
	}

	// Check if this SetLiteral should be treated as an array (array of const)
	// This happens when semantic analyzer determined it's used in array context
	if e.semanticInfo != nil {
		if typeAnnot := e.semanticInfo.GetType(node); typeAnnot != nil && typeAnnot.Name != "" {
			resolvedType, err := e.ResolveTypeWithContext(typeAnnot.Name, ctx)
			if err == nil {
				if arrayType, isArray := resolvedType.(*types.ArrayType); isArray {
					// Task 3.5.104: Evaluate as array literal directly instead of delegating
					// Convert SetLiteral to ArrayLiteralExpression and evaluate directly
					arrayLit := &ast.ArrayLiteralExpression{
						Elements: node.Elements,
					}
					// Pass the array type via semantic info if possible
					return e.evalArrayLiteralWithType(arrayLit, arrayType, ctx)
				}
			}
		}
	}

	// Evaluate all elements and determine the ordinal type
	var elementType types.Type
	var enumType *types.EnumType
	ordinals := make(map[int]bool)
	var lazyRanges []runtime.IntRange

	for _, elem := range node.Elements {
		// Check if this is a range expression
		if rangeExpr, ok := elem.(*ast.RangeExpression); ok {
			result := e.evalSetRangeElement(rangeExpr, ctx, &elementType, &enumType, ordinals, &lazyRanges)
			if isError(result) {
				return result
			}
		} else {
			// Simple element (not a range)
			result := e.evalSetSimpleElement(elem, ctx, &elementType, &enumType, ordinals)
			if isError(result) {
				return result
			}
		}
	}

	// Handle empty set - if no element type determined, we can't infer it
	if elementType == nil && len(node.Elements) == 0 {
		return e.newError(node, "cannot infer type for empty set literal")
	}

	// Create the SetType (automatically selects storage strategy)
	setType := types.NewSetType(elementType)

	// Create SetValue and populate the correct storage backend
	setValue := runtime.NewSetValue(setType)

	// Populate storage based on strategy
	switch setType.StorageKind {
	case types.SetStorageBitmask:
		// Use bitmask - convert map to bitset
		var elements uint64
		for ordinal := range ordinals {
			if ordinal >= 64 {
				return e.newError(node, "enum ordinal %d out of range for bitmask storage", ordinal)
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

// evalSetRangeElement evaluates a range expression as a set element.
// Example: 1..10, 'a'..'z', one..five
func (e *Evaluator) evalSetRangeElement(
	rangeExpr *ast.RangeExpression,
	ctx *ExecutionContext,
	elementType *types.Type,
	enumType **types.EnumType,
	ordinals map[int]bool,
	lazyRanges *[]runtime.IntRange,
) Value {
	// Evaluate range endpoints
	startVal := e.Eval(rangeExpr.Start, ctx)
	if isError(startVal) {
		return startVal
	}

	endVal := e.Eval(rangeExpr.RangeEnd, ctx)
	if isError(endVal) {
		return endVal
	}

	// Extract ordinal values
	startOrd, err1 := GetOrdinalValue(startVal)
	endOrd, err2 := GetOrdinalValue(endVal)

	if err1 != nil {
		return e.newError(rangeExpr.Start, "range start must be ordinal type: %s", err1.Error())
	}
	if err2 != nil {
		return e.newError(rangeExpr.RangeEnd, "range end must be ordinal type: %s", err2.Error())
	}

	// Determine element type from first range
	if *elementType == nil {
		if enumVal, isEnum := startVal.(*runtime.EnumValue); isEnum {
			// Get enum type from environment
			et, err := e.lookupEnumType(enumVal.TypeName, ctx)
			if err != nil {
				return e.newError(rangeExpr, "%s", err.Error())
			}
			*enumType = et
			*elementType = et
		} else {
			*elementType = GetOrdinalType(startVal)
		}
	}

	// Verify both endpoints are same type
	if enumVal1, ok1 := startVal.(*runtime.EnumValue); ok1 {
		if enumVal2, ok2 := endVal.(*runtime.EnumValue); ok2 {
			if enumVal1.TypeName != enumVal2.TypeName {
				return e.newError(rangeExpr, "range endpoints must be same enum type")
			}
		} else {
			return e.newError(rangeExpr, "range endpoints type mismatch")
		}
	}

	// Integer ranges are stored lazily (not expanded)
	// Enum ranges must be expanded for proper set operations
	if *enumType == nil {
		// Store as lazy range (integer types only)
		*lazyRanges = append(*lazyRanges, runtime.IntRange{Start: startOrd, End: endOrd})
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

	return nil
}

// evalSetSimpleElement evaluates a simple (non-range) set element.
func (e *Evaluator) evalSetSimpleElement(
	elem ast.Expression,
	ctx *ExecutionContext,
	elementType *types.Type,
	enumType **types.EnumType,
	ordinals map[int]bool,
) Value {
	elemVal := e.Eval(elem, ctx)
	if isError(elemVal) {
		return elemVal
	}

	// Extract ordinal value
	ordinal, err := GetOrdinalValue(elemVal)
	if err != nil {
		return e.newError(elem, "set element must be ordinal type: %s", err.Error())
	}

	// Determine element type from first element
	if *elementType == nil {
		if enumVal, isEnum := elemVal.(*runtime.EnumValue); isEnum {
			// Get enum type from environment
			et, lookupErr := e.lookupEnumType(enumVal.TypeName, ctx)
			if lookupErr != nil {
				return e.newError(elem, "%s", lookupErr.Error())
			}
			*enumType = et
			*elementType = et
		} else {
			*elementType = GetOrdinalType(elemVal)
		}
	} else {
		// Verify all elements are of the same type
		if *enumType != nil {
			if enumVal, ok := elemVal.(*runtime.EnumValue); ok {
				if enumVal.TypeName != (*enumType).Name {
					return e.newError(elem, "type mismatch in set literal: expected %s, got %s",
						(*enumType).Name, enumVal.TypeName)
				}
			} else {
				return e.newError(elem, "type mismatch in set literal: expected enum %s, got %s",
					(*enumType).Name, elemVal.Type())
			}
		}
	}

	// Add element to ordinals map
	ordinals[ordinal] = true

	return nil
}

// lookupEnumType looks up an enum type by name from the environment.
// Task 3.5.104: Removed adapter.GetType() fallback - use direct environment lookup only.
func (e *Evaluator) lookupEnumType(typeName string, ctx *ExecutionContext) (*types.EnumType, error) {
	// Enum types are stored in environment with "__enum_type_" prefix
	// Use ident.Normalize for case-insensitive lookup (consistent with other type lookups)
	normalizedName := ident.Normalize(typeName)
	typeVal, ok := ctx.Env().Get("__enum_type_" + normalizedName)
	if !ok {
		return nil, fmt.Errorf("unknown enum type '%s'", typeName)
	}

	// The stored value should have an EnumType field
	// We need to extract it - use type assertion via interface
	if enumTypeProvider, ok := typeVal.(interface{ GetEnumType() *types.EnumType }); ok {
		return enumTypeProvider.GetEnumType(), nil
	}

	// The value was found but doesn't implement GetEnumType() - this is a programming error
	return nil, fmt.Errorf("type '%s' is registered but does not provide EnumType (internal error)", typeName)
}

// parseInlineSetType parses inline set type syntax like "set of TEnumType".
// Returns the SetType, or nil if the string doesn't match the expected format.
// Task 3.5.129a: Migrated from Interpreter to enable direct set zero-value creation.
func (e *Evaluator) parseInlineSetType(signature string, ctx *ExecutionContext) *types.SetType {
	// Check for "set of " prefix (case-sensitive per DWScript spec)
	if !strings.HasPrefix(signature, "set of ") {
		return nil
	}

	// Extract enum type name: "set of TColor" → "TColor"
	enumTypeName := strings.TrimSpace(signature[7:]) // Skip "set of "
	if enumTypeName == "" {
		return nil
	}

	// Look up the enum type using existing helper
	enumType, err := e.lookupEnumType(enumTypeName, ctx)
	if err != nil {
		return nil
	}

	// Create and return the set type
	return types.NewSetType(enumType)
}
