package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// resolveTypeName resolves a type name string to a types.Type.
// Handles primitives (Integer, Float, String, Boolean, Variant, TDateTime, Nil, Void),
// registered types (enums, records, classes, interfaces), and function pointer aliases.
func (e *Evaluator) resolveTypeName(typeName string, ctx *ExecutionContext) (types.Type, error) {
	if ctx == nil {
		ctx = &ExecutionContext{}
	}
	cleanTypeName := typeName

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
	case "jsonvariant":
		return types.JSON_VARIANT, nil

	case "bytebuffer":
		return types.BYTE_BUFFER, nil

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
				return enumMetadata.GetEnumType(), nil
			}
		}

		// Environment-based lookups (records, type aliases, subranges)
		if ctx.Env() != nil {
			// Try nested class type in current class context.
			if currentClassRaw, ok := ctx.Env().Get("__CurrentClass__"); ok {
				if currentClass, ok := currentClassRaw.(ClassMetaValue); ok && currentClass != nil {
					if ident.Equal(currentClass.GetClassName(), cleanTypeName) {
						return currentClass.GetClassInfo().GetClassType(), nil
					}
					if nestedVal := currentClass.GetNestedClass(cleanTypeName); nestedVal != nil {
						if nestedClass, ok := nestedVal.(ClassMetaValue); ok {
							return nestedClass.GetClassInfo().GetClassType(), nil
						}
					}
				}
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

			if setValue, ok := ctx.Env().Get("__set_type_" + normalizedName); ok {
				if provider, ok := setValue.(interface{ GetSetType() *types.SetType }); ok {
					return provider.GetSetType(), nil
				}
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

		// TClass is the builtin metaclass of TObject
		if normalizedName == "tclass" {
			return types.NewClassOfType(e.buildClassTypeWithHierarchy("TObject")), nil
		}

		// Try class type via TypeSystem
		if e.typeSystem != nil && e.typeSystem.HasClass(cleanTypeName) {
			// Build the class type with its parent chain so overload resolution
			// can rank subclass -> base-class conversions.
			return e.buildClassTypeWithHierarchy(cleanTypeName), nil
		}

		// Try interface type via TypeSystem
		if e.typeSystem != nil && e.typeSystem.HasInterface(cleanTypeName) {
			return e.typeSystem.LookupInterface(cleanTypeName).GetInterfaceType(), nil
		}

		// Try array type via TypeSystem
		if e.typeSystem != nil {
			if arrayType := e.typeSystem.LookupArrayType(cleanTypeName); arrayType != nil {
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

// resolveArrayElementType resolves an array's element type expression (which may
// itself be a nested array).
func (e *Evaluator) resolveArrayElementType(elementExpr ast.TypeExpression, ctx *ExecutionContext) types.Type {
	if elementExpr == nil {
		return nil
	}
	elementType, err := e.ResolveTypeFromAnnotation(elementExpr, ctx)
	if err != nil {
		return nil
	}
	return elementType
}

// resolveAssociativeArrayTypeNode resolves an ArrayTypeNode to an associative
// array type when its index type is not a bounded ordinal. Returns nil when the
// node is not an associative array (dynamic, static, or enum-indexed).
func (e *Evaluator) resolveAssociativeArrayTypeNode(arrayNode *ast.ArrayTypeNode, ctx *ExecutionContext) *types.AssociativeArrayType {
	if arrayNode == nil || !arrayNode.IsEnumIndexed() {
		return nil
	}
	indexType, err := e.ResolveTypeFromAnnotation(arrayNode.IndexType, ctx)
	if err != nil {
		return nil
	}
	if _, _, ok := types.OrdinalBounds(indexType); ok {
		return nil // bounded ordinal => enum-indexed static array, not associative
	}
	elementType := e.resolveArrayElementType(arrayNode.ElementType, ctx)
	if elementType == nil {
		return nil
	}
	return types.NewAssociativeArrayType(indexType, elementType)
}

// resolveArrayTypeNode resolves an ArrayTypeNode directly from the AST.
// Handles dynamic, static, ordinal-indexed, and nested arrays.
func (e *Evaluator) resolveArrayTypeNode(arrayNode *ast.ArrayTypeNode, ctx *ExecutionContext) *types.ArrayType {
	if arrayNode == nil {
		return nil
	}

	// Resolve element type first
	elementType := e.resolveArrayElementType(arrayNode.ElementType, ctx)
	if elementType == nil {
		return nil
	}

	// Check if dynamic or static array
	if arrayNode.IsDynamic() {
		return types.NewDynamicArrayType(elementType)
	}

	// Ordinal-indexed array (enum, boolean, subrange)
	if arrayNode.IsEnumIndexed() {
		indexType, err := e.ResolveTypeFromAnnotation(arrayNode.IndexType, ctx)
		if err != nil {
			return nil
		}

		low, high, ok := types.OrdinalBounds(indexType)
		if !ok {
			return nil
		}

		return types.NewStaticArrayTypeWithIndexType(elementType, indexType, low, high)
	}

	// Static array - evaluate constant bound expressions.
	lowBoundValue, ok := e.resolveStaticArrayBound(arrayNode.LowBound, ctx)
	if !ok {
		return nil
	}
	highBoundValue, ok := e.resolveStaticArrayBound(arrayNode.HighBound, ctx)
	if !ok {
		return nil
	}

	return types.NewStaticArrayType(elementType, lowBoundValue, highBoundValue)
}

func (e *Evaluator) resolveStaticArrayBound(expr ast.Expression, ctx *ExecutionContext) (int, bool) {
	if expr == nil {
		return 0, false
	}

	// Fast paths for simple literal bounds.
	if intLit, ok := expr.(*ast.IntegerLiteral); ok {
		return int(intLit.Value), true
	}
	if unary, ok := expr.(*ast.UnaryExpression); ok && unary.Operator == "-" {
		if intLit, ok := unary.Right.(*ast.IntegerLiteral); ok {
			return -int(intLit.Value), true
		}
	}

	value := e.Eval(expr, ctx)
	if isError(value) {
		return 0, false
	}

	intVal, ok := value.(*runtime.IntegerValue)
	if !ok {
		return 0, false
	}

	return int(intVal.Value), true
}
