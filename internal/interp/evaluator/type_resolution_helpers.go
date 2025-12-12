package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// resolveTypeName resolves a type name string to a types.Type.
// Handles primitives (Integer, Float, String, Boolean, Variant, TDateTime, Nil, Void),
// registered types (enums, records, classes, interfaces), and inline array types.
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

	// Handle primitive types
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
		// "Const" is deprecated, redirect to Variant
		return types.VARIANT, nil

	case "tdatetime":
		return types.DATETIME, nil

	case "nil":
		return types.NIL, nil

	case "void":
		return types.VOID, nil

	default:
		// Try enum type via TypeSystem
		// Check if typeSystem is initialized (defensive programming for tests)
		if e.typeSystem != nil {
			if enumMetadata := e.typeSystem.LookupEnumMetadata(cleanTypeName); enumMetadata != nil {
				if etv, ok := enumMetadata.(EnumTypeValueAccessor); ok {
					return etv.GetEnumType(), nil
				}
				// Found but wrong type - programming error
				return nil, fmt.Errorf("type '%s' is registered as enum but does not provide EnumType (internal error)", typeName)
			}
		}

		// Environment-based lookups (records, type aliases, subranges)
		if ctx.Env() != nil {

			// Try record type (stored in environment with "__record_type_" prefix)
			if recordTypeVal, ok := ctx.Env().Get("__record_type_" + normalizedName); ok {
				// Extract RecordType using interface method
				if recordTypeProvider, ok := recordTypeVal.(interface{ GetRecordType() *types.RecordType }); ok {
					return recordTypeProvider.GetRecordType(), nil
				}
				// Found but wrong type - programming error
				return nil, fmt.Errorf("type '%s' is registered as record but does not provide RecordType (internal error)", typeName)
			}

			// Try type alias (stored in environment with "__type_alias_" prefix)
			if typeAliasVal, ok := ctx.Env().Get("__type_alias_" + normalizedName); ok {
				// Extract aliased type using interface method
				if typeAliasProvider, ok := typeAliasVal.(interface{ GetAliasedType() types.Type }); ok {
					return typeAliasProvider.GetAliasedType(), nil
				}
				// Found but wrong type - programming error
				return nil, fmt.Errorf("type '%s' is registered as type alias but does not provide AliasedType (internal error)", typeName)
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

		// Try array type via TypeSystem
		if e.typeSystem != nil {
			if arrayType := e.typeSystem.LookupArrayType(cleanTypeName); arrayType != nil {
				return arrayType, nil
			}
		}

		// Try inline array type parsing
		if strings.HasPrefix(ident.Normalize(cleanTypeName), "array") {
			if arrayType := e.parseInlineArrayType(cleanTypeName, ctx); arrayType != nil {
				return arrayType, nil
			}
		}

		// Use TypeSystem for subrange type lookup
		if e.typeSystem != nil {
			if subrangeType := e.typeSystem.LookupSubrangeType(typeName); subrangeType != nil {
				return subrangeType, nil
			}
		}

		// Function/method pointer types registered in the type system
		if e.typeSystem != nil {
			if funcPtrType := e.typeSystem.LookupFunctionPointerType(cleanTypeName); funcPtrType != nil {
				return funcPtrType, nil
			}
		}

		// Unknown type
		return nil, fmt.Errorf("unknown type: %s", typeName)
	}
}

// parseInlineArrayType parses inline array type signatures like "array of Type" or "array[low..high] of Type".
// Supports nested arrays and both static and dynamic array syntax.
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

	// Resolve element type recursively (handles nested arrays)
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
// Handles dynamic, static, ordinal-indexed, and nested arrays.
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
