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
// Type Resolution
// ============================================================================
//
// Task 3.5.79: Type resolution for the Evaluator.
// Resolves type names to types.Type using direct service access where possible.
// ============================================================================

// ResolveType resolves a type name to a types.Type.
// This method provides direct type resolution using the Evaluator's services,
// reducing adapter dependency where possible.
//
// Resolution order:
//  1. Built-in types (Integer, Float, String, Boolean, Variant, Const)
//  2. Inline array types (array of X, array[...])
//  3. Named array types (via TypeSystem)
//  4. All other types (enum, record, class, alias, subrange) via adapter
//
// The lookup is case-insensitive per DWScript semantics.
//
// Returns:
//   - The resolved types.Type
//   - An error if the type cannot be resolved
func (e *Evaluator) ResolveType(typeName string) (types.Type, error) {
	// Step 1: Handle inline array types first
	if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
		return e.resolveInlineArrayType(typeName)
	}

	// Step 2: Normalize for case-insensitive lookup
	lowerTypeName := ident.Normalize(typeName)

	// Step 3: Check built-in types
	switch lowerTypeName {
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
		// "Const" redirects to VARIANT for dynamic typing
		return types.VARIANT, nil
	}

	// Step 4: Check named array types via TypeSystem (direct access)
	if arrayType := e.typeSystem.LookupArrayType(typeName); arrayType != nil {
		return arrayType, nil
	}

	// Step 5: Delegate to adapter for all other types
	// (enum, record, class, type alias, subrange - stored in environment)
	resolvedType, err := e.adapter.GetType(typeName)
	if err != nil {
		return nil, err
	}

	// Cast the result to types.Type
	if typ, ok := resolvedType.(types.Type); ok {
		return typ, nil
	}

	return nil, fmt.Errorf("resolved type %q is not a types.Type", typeName)
}

// resolveInlineArrayType handles inline array type syntax:
//   - "array of Integer" → dynamic array
//   - "array[1..10] of String" → static array
func (e *Evaluator) resolveInlineArrayType(typeName string) (types.Type, error) {
	// Handle "array of ElementType" (dynamic array)
	if strings.HasPrefix(typeName, "array of ") {
		elementTypeName := strings.TrimPrefix(typeName, "array of ")
		elementType, err := e.ResolveType(elementTypeName)
		if err != nil {
			return nil, fmt.Errorf("invalid array element type: %w", err)
		}
		return types.NewDynamicArrayType(elementType), nil
	}

	// Handle "array[bounds] of ElementType" (static array)
	// Format: array[low..high] of ElementType
	if strings.HasPrefix(typeName, "array[") {
		// Find the "] of " separator
		ofIdx := strings.Index(typeName, "] of ")
		if ofIdx == -1 {
			return nil, fmt.Errorf("invalid array type syntax: %s", typeName)
		}

		boundsStr := typeName[6:ofIdx] // Extract "low..high"
		elementTypeName := typeName[ofIdx+5:]

		// Parse bounds
		parts := strings.Split(boundsStr, "..")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid array bounds: %s", boundsStr)
		}

		var low, high int
		_, err := fmt.Sscanf(parts[0], "%d", &low)
		if err != nil {
			return nil, fmt.Errorf("invalid low bound: %s", parts[0])
		}
		_, err = fmt.Sscanf(parts[1], "%d", &high)
		if err != nil {
			return nil, fmt.Errorf("invalid high bound: %s", parts[1])
		}

		// Resolve element type
		elementType, err := e.ResolveType(elementTypeName)
		if err != nil {
			return nil, fmt.Errorf("invalid array element type: %w", err)
		}

		return types.NewStaticArrayType(elementType, low, high), nil
	}

	return nil, fmt.Errorf("invalid inline array type: %s", typeName)
}

// ResolveTypeFromName is an alias for ResolveType for backward compatibility.
// Deprecated: Use ResolveType instead.
func (e *Evaluator) ResolveTypeFromName(typeName string) (types.Type, error) {
	return e.ResolveType(typeName)
}

// ============================================================================
// Type Annotation Resolution
// ============================================================================
//
// Task 3.5.102g: Resolve types from AST type annotations.
// ============================================================================

// ResolveTypeFromAnnotation resolves a type from an AST TypeExpression.
// This is used for function return types, parameter types, and variable declarations.
//
// Task 3.5.102g: Migrated from Interpreter.resolveTypeFromAnnotation().
func (e *Evaluator) ResolveTypeFromAnnotation(typeExpr ast.TypeExpression) (types.Type, error) {
	if typeExpr == nil {
		return nil, nil
	}

	// Get the type name string from the expression
	typeName := typeExpr.String()

	// Delegate to ResolveType which handles all type resolution
	return e.ResolveType(typeName)
}

// ============================================================================
// Default Value Creation
// ============================================================================
//
// Task 3.5.102g: Create default/zero values for types.
// ============================================================================

// GetDefaultValue returns the default/zero value for a given type.
// This is used for Result variable initialization in functions.
//
// Task 3.5.102g: Migrated from Interpreter.getDefaultValue().
func (e *Evaluator) GetDefaultValue(typ types.Type) Value {
	if typ == nil {
		return e.nilValue()
	}

	switch typ.TypeKind() {
	case "STRING":
		return &runtime.StringValue{Value: ""}
	case "INTEGER":
		return &runtime.IntegerValue{Value: 0}
	case "FLOAT":
		return &runtime.FloatValue{Value: 0.0}
	case "BOOLEAN":
		return &runtime.BooleanValue{Value: false}
	case "CLASS", "INTERFACE", "FUNCTION_POINTER", "METHOD_POINTER":
		return e.nilValue()
	case "ARRAY":
		// Arrays should default to an empty array value of the correct element type.
		if arrType, ok := typ.(*types.ArrayType); ok {
			return runtime.NewArrayValue(arrType, nil)
		}
		return e.nilValue()
	case "RECORD":
		// Records should be initialized with default field values.
		// For now, return NIL (will be enhanced if needed).
		return e.nilValue()
	case "VARIANT":
		// Variants default to Unassigned (nil-like)
		return e.nilValue()
	default:
		// Unknown types default to NIL
		return e.nilValue()
	}
}

// nilValue returns a nil value.
// Task 3.5.102g: Helper for creating nil values.
func (e *Evaluator) nilValue() Value {
	return &runtime.NilValue{}
}
