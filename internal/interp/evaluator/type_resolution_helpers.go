package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// type_resolution_helpers.go
//
// This file provides type resolution helpers for the evaluator.
// Task 3.5.139a: Initial implementation with primitive types.
//
// These helpers enable the evaluator to resolve type names to their
// corresponding types.Type instances without delegating to the adapter.
// This is essential for array type resolution and other type operations.

// resolveTypeName resolves a type name string to a types.Type.
//
// Task 3.5.139a: Initial implementation handles primitive types:
//   - Integer, Float, String, Boolean
//   - Variant, Const (deprecated, mapped to Variant)
//   - TDateTime, Nil, Void
//
// Task 3.5.139b: Extended to handle registered types:
//   - Enum types (from environment "__enum_type_" prefix)
//   - Record types (from environment "__record_type_" prefix)
//   - Class types (from TypeSystem)
//   - Interface types (from TypeSystem)
//   - Subrange types (from environment "__subrange_type_" prefix)
//
// Task 3.5.139c: Extended to handle array types:
//   - Named array types (from TypeSystem.LookupArrayType)
//
// Task 3.5.139d: Extended to handle inline array types:
//   - Inline array syntax ("array of Type", "array[low..high] of Type")
//   - Nested arrays via recursive resolution
//
// Parameters:
//   - typeName: The name of the type to resolve (case-insensitive)
//   - ctx: ExecutionContext for environment access (used in future tasks)
//
// Returns:
//   - The resolved types.Type
//   - An error if the type cannot be resolved
//
// Examples:
//
//	resolveTypeName("Integer", ctx) → types.INTEGER, nil
//	resolveTypeName("float", ctx)   → types.FLOAT, nil
//	resolveTypeName("STRING", ctx)  → types.STRING, nil
//	resolveTypeName("Variant", ctx) → types.VARIANT, nil
//	resolveTypeName("Unknown", ctx) → nil, error
func (e *Evaluator) resolveTypeName(typeName string, ctx *ExecutionContext) (types.Type, error) {
	// Strip parent qualification from class type strings like "TSub(TBase)"
	// This enables proper resolution using the declared class name.
	cleanTypeName := typeName
	if idx := strings.Index(cleanTypeName, "("); idx != -1 {
		cleanTypeName = strings.TrimSpace(cleanTypeName[:idx])
	}

	// Normalize type name for case-insensitive comparison
	// DWScript (like Pascal) is case-insensitive for all identifiers
	normalizedName := ident.Normalize(cleanTypeName)

	// Task 3.5.139a: Handle primitive types
	switch normalizedName {
	case "integer":
		return types.INTEGER, nil

	case "float":
		return types.FLOAT, nil

	case "string":
		return types.STRING, nil

	case "boolean":
		return types.BOOLEAN, nil

	case "variant":
		return types.VARIANT, nil

	case "const":
		// Task 3.5.139a: "Const" is deprecated, redirect to Variant
		// This matches the interpreter's behavior (record.go:562-565)
		return types.VARIANT, nil

	case "tdatetime":
		return types.DATETIME, nil

	case "nil":
		return types.NIL, nil

	case "void":
		return types.VOID, nil

	default:
		// Task 3.5.139b: Handle registered types (enums, records, classes, interfaces, subranges)

		// Environment-based lookups (enums, records, subranges)
		if ctx.Env() != nil {
			// Try enum type (stored in environment with "__enum_type_" prefix)
			if enumTypeVal, ok := ctx.Env().Get("__enum_type_" + normalizedName); ok {
				// Extract EnumType using interface method
				if enumTypeProvider, ok := enumTypeVal.(interface{ GetEnumType() *types.EnumType }); ok {
					return enumTypeProvider.GetEnumType(), nil
				}
				// Found but wrong type - programming error
				return nil, fmt.Errorf("type '%s' is registered as enum but does not provide EnumType (internal error)", typeName)
			}

			// Try record type (stored in environment with "__record_type_" prefix)
			if recordTypeVal, ok := ctx.Env().Get("__record_type_" + normalizedName); ok {
				// Extract RecordType using interface method
				if recordTypeProvider, ok := recordTypeVal.(interface{ GetRecordType() *types.RecordType }); ok {
					return recordTypeProvider.GetRecordType(), nil
				}
				// Found but wrong type - programming error
				return nil, fmt.Errorf("type '%s' is registered as record but does not provide RecordType (internal error)", typeName)
			}
		}

		// Try class type via TypeSystem
		if e.typeSystem != nil && e.typeSystem.HasClass(cleanTypeName) {
			// Use nominal class type for runtime type information
			// Note: Uses cleanTypeName (without parent qualification) for lookup
			return types.NewClassType(cleanTypeName, nil), nil
		}

		// Try interface type via TypeSystem
		if e.typeSystem != nil && e.typeSystem.HasInterface(cleanTypeName) {
			// Create an InterfaceType with the clean name
			// Note: The TypeSystem stores InterfaceInfo as 'any', so we just create the type directly
			return types.NewInterfaceType(cleanTypeName), nil
		}

		// Task 3.5.139c: Try array type via TypeSystem
		if e.typeSystem != nil {
			if arrayType := e.typeSystem.LookupArrayType(cleanTypeName); arrayType != nil {
				return arrayType, nil
			}
		}

		// Task 3.5.139d: Try inline array type parsing
		// Check if this is an inline array type like "array of T" or "array[low..high] of T"
		if strings.HasPrefix(ident.Normalize(cleanTypeName), "array") {
			if arrayType := e.parseInlineArrayType(cleanTypeName, ctx); arrayType != nil {
				return arrayType, nil
			}
		}

		// Environment-based lookups (enums, records, subranges) - continued
		if ctx.Env() != nil {
			// Try subrange type (stored in environment with "__subrange_type_" prefix)
			if subrangeTypeVal, ok := ctx.Env().Get("__subrange_type_" + normalizedName); ok {
				// Extract SubrangeType using interface method
				if subrangeTypeProvider, ok := subrangeTypeVal.(interface{ GetSubrangeType() *types.SubrangeType }); ok {
					return subrangeTypeProvider.GetSubrangeType(), nil
				}
				// Found but wrong type - programming error
				return nil, fmt.Errorf("type '%s' is registered as subrange but does not provide SubrangeType (internal error)", typeName)
			}
		}

		// Task 3.5.139c: Will add array type lookups
		// Task 3.5.139c: Will add type alias lookups (if needed)

		// Unknown type
		return nil, fmt.Errorf("unknown type: %s", typeName)
	}
}

// parseInlineArrayType parses a DWScript inline array type signature (static or dynamic)
// from a string, extracting bounds and element type information.
//
// Task 3.5.139d: Helper method for parsing inline array syntax.
//
// Supported formats:
//   - "array of Type" → dynamic array
//   - "array[low..high] of Type" → static array with bounds
//
// The element type can be any resolvable type, including nested arrays.
// For example: "array of array of Integer" creates a 2D dynamic array.
//
// Parameters:
//   - signature: The array type signature string (case-insensitive)
//   - ctx: ExecutionContext for type resolution
//
// Returns:
//   - *types.ArrayType if parsing succeeds
//   - nil if the signature doesn't match expected format or element type cannot be resolved
//
// Examples:
//
//	parseInlineArrayType("array of Integer", ctx) → DynamicArrayType with Integer elements
//	parseInlineArrayType("array[0..9] of String", ctx) → StaticArrayType[0..9] with String elements
//	parseInlineArrayType("array of array of Float", ctx) → DynamicArrayType of DynamicArrayType of Float
func (e *Evaluator) parseInlineArrayType(signature string, ctx *ExecutionContext) *types.ArrayType {
	var lowBound, highBound *int

	// Normalize signature for case-insensitive parsing
	lowerSignature := strings.ToLower(signature)

	// Check if this is a static array with bounds
	if strings.HasPrefix(lowerSignature, "array[") {
		// Extract bounds: array[low..high] of Type
		endBracket := strings.Index(signature, "]")
		if endBracket == -1 {
			return nil
		}

		boundsStr := signature[6:endBracket] // Skip "array["
		parts := strings.Split(boundsStr, "..")
		if len(parts) != 2 {
			return nil
		}

		// Parse low bound
		low := 0
		if _, err := fmt.Sscanf(parts[0], "%d", &low); err != nil {
			return nil
		}
		lowBound = &low

		// Parse high bound
		high := 0
		if _, err := fmt.Sscanf(parts[1], "%d", &high); err != nil {
			return nil
		}
		highBound = &high

		// Skip past "] of " in original signature (preserve case for element type)
		signature = signature[endBracket+1:]
		lowerSignature = lowerSignature[endBracket+1:]
	} else if strings.HasPrefix(lowerSignature, "array of ") {
		// Dynamic array: skip "array" to get " of ElementType"
		signature = signature[5:] // Skip "array" (preserve case for element type)
		lowerSignature = lowerSignature[5:]
	} else {
		return nil
	}

	// Now signature should be " of ElementType"
	if !strings.HasPrefix(lowerSignature, " of ") {
		return nil
	}

	// Extract element type name (from original signature to preserve case)
	elementTypeName := strings.TrimSpace(signature[4:]) // Skip " of "

	// Resolve element type recursively
	// Task 3.5.139d: Use resolveTypeName which handles all type categories
	// For nested arrays like "array of array of Integer", this will recursively
	// call parseInlineArrayType via resolveTypeName
	elementType, err := e.resolveTypeName(elementTypeName, ctx)
	if err != nil || elementType == nil {
		return nil
	}

	// Create array type
	if lowBound != nil && highBound != nil {
		return types.NewStaticArrayType(elementType, *lowBound, *highBound)
	}
	return types.NewDynamicArrayType(elementType)
}

// resolveArrayTypeNode resolves an ArrayTypeNode directly from the AST.
// This avoids string conversion issues with parentheses in bound expressions like (-5).
//
// Task 3.5.139e: Helper method for resolving AST ArrayTypeNode.
//
// Handles:
//   - Dynamic arrays (no bounds)
//   - Static arrays with integer literal bounds
//   - Ordinal-indexed arrays (enum, boolean, subrange)
//   - Nested arrays (recursive resolution)
//
// Parameters:
//   - arrayNode: The AST ArrayTypeNode to resolve
//   - ctx: ExecutionContext for type resolution
//
// Returns:
//   - *types.ArrayType if resolution succeeds
//   - nil if the node cannot be resolved or element type is invalid
//
// Examples:
//
//	resolveArrayTypeNode(ArrayTypeNode{ElementType: "Integer", IsDynamic: true}) → DynamicArrayType
//	resolveArrayTypeNode(ArrayTypeNode{LowBound: 0, HighBound: 9, ElementType: "String"}) → StaticArrayType[0..9]
func (e *Evaluator) resolveArrayTypeNode(arrayNode *ast.ArrayTypeNode, ctx *ExecutionContext) *types.ArrayType {
	if arrayNode == nil {
		return nil
	}

	// Resolve element type first
	var elementType types.Type

	// Check if element type is also an array (nested arrays)
	if nestedArray, ok := arrayNode.ElementType.(*ast.ArrayTypeNode); ok {
		elementType = e.resolveArrayTypeNode(nestedArray, ctx)
		if elementType == nil {
			return nil
		}
	} else {
		// Get element type name and resolve it
		var elementTypeName string
		if typeAnnot, ok := arrayNode.ElementType.(*ast.TypeAnnotation); ok {
			elementTypeName = typeAnnot.Name
		} else {
			elementTypeName = arrayNode.ElementType.String()
		}

		var err error
		elementType, err = e.resolveTypeName(elementTypeName, ctx)
		if err != nil || elementType == nil {
			return nil
		}
	}

	// Check if dynamic or static array
	if arrayNode.IsDynamic() {
		return types.NewDynamicArrayType(elementType)
	}

	// Ordinal-indexed array (enum, boolean, subrange)
	if arrayNode.IsEnumIndexed() {
		indexTypeName := arrayNode.IndexType.String()
		indexType, err := e.resolveTypeName(indexTypeName, ctx)
		if err != nil {
			return nil
		}

		low, high, ok := types.OrdinalBounds(indexType)
		if !ok {
			return nil
		}

		return types.NewStaticArrayType(elementType, low, high)
	}

	// Static array - extract bounds from AST
	// Task 3.5.139e: Handle IntegerLiteral bounds directly (most common case)
	// For more complex expressions (like -5), use UnaryExpression with IntegerLiteral

	// Extract low bound
	var lowBoundValue int
	if intLit, ok := arrayNode.LowBound.(*ast.IntegerLiteral); ok {
		lowBoundValue = int(intLit.Value)
	} else if unary, ok := arrayNode.LowBound.(*ast.UnaryExpression); ok {
		// Handle unary minus for negative bounds like -5
		if unary.Operator == "-" {
			if intLit, ok := unary.Right.(*ast.IntegerLiteral); ok {
				lowBoundValue = -int(intLit.Value)
			} else {
				return nil
			}
		} else {
			return nil
		}
	} else {
		return nil
	}

	// Extract high bound
	var highBoundValue int
	if intLit, ok := arrayNode.HighBound.(*ast.IntegerLiteral); ok {
		highBoundValue = int(intLit.Value)
	} else if unary, ok := arrayNode.HighBound.(*ast.UnaryExpression); ok {
		// Handle unary minus for negative bounds
		if unary.Operator == "-" {
			if intLit, ok := unary.Right.(*ast.IntegerLiteral); ok {
				highBoundValue = -int(intLit.Value)
			} else {
				return nil
			}
		} else {
			return nil
		}
	} else {
		return nil
	}

	return types.NewStaticArrayType(elementType, lowBoundValue, highBoundValue)
}
