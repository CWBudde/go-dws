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

	// Step 1b: Handle function pointer types (e.g., "function(Integer): String")
	if strings.HasPrefix(typeName, "function(") || strings.HasPrefix(typeName, "procedure(") {
		return e.resolveFunctionPointerType(typeName)
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
	case "jsonvariant":
		return types.JSON_VARIANT, nil
	case "const":
		// "Const" redirects to VARIANT for dynamic typing
		return types.VARIANT, nil
	case "tclass":
		if !e.typeSystem.HasClass("TObject") {
			return nil, fmt.Errorf("unknown class type 'TObject'")
		}
		return types.NewClassOfType(types.NewClassType("TObject", nil)), nil
	}

	// Step 4: Check named array types via TypeSystem (direct access)
	if arrayType := e.typeSystem.LookupArrayType(typeName); arrayType != nil {
		return arrayType, nil
	}

	// Step 5: Use evaluator's resolveTypeName for all other types
	// Use currentContext if available for full resolution including environment-based lookups
	// (records, type aliases, subranges). Falls back to empty context if not in an evaluation.
	ctx := e.currentContext
	if ctx == nil {
		ctx = &ExecutionContext{}
	}
	resolvedType, err := e.resolveTypeName(typeName, ctx)
	if err != nil {
		return nil, err
	}

	return resolvedType, nil
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

		boundsStr := strings.TrimSpace(typeName[6:ofIdx]) // Extract "low..high" or index type
		elementTypeName := strings.TrimSpace(typeName[ofIdx+5:])

		// Parse bounds or ordinal index type
		parts := strings.Split(boundsStr, "..")
		var low, high int
		var hasOrdinalBounds bool

		if len(parts) == 2 {
			_, err := fmt.Sscanf(parts[0], "%d", &low)
			if err != nil {
				return nil, fmt.Errorf("invalid low bound: %s", parts[0])
			}
			_, err = fmt.Sscanf(parts[1], "%d", &high)
			if err != nil {
				return nil, fmt.Errorf("invalid high bound: %s", parts[1])
			}
		} else {
			indexType, err := e.ResolveType(boundsStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index type: %w", err)
			}

			ordLow, ordHigh, ok := types.OrdinalBounds(indexType)
			if !ok {
				return nil, fmt.Errorf("array index type '%s' must be a bounded ordinal", boundsStr)
			}

			low, high = ordLow, ordHigh
			hasOrdinalBounds = true
		}

		// Resolve element type
		elementType, err := e.ResolveType(elementTypeName)
		if err != nil {
			return nil, fmt.Errorf("invalid array element type: %w", err)
		}

		if len(parts) != 2 && !hasOrdinalBounds {
			return nil, fmt.Errorf("invalid array bounds: %s", boundsStr)
		}

		return types.NewStaticArrayType(elementType, low, high), nil
	}

	return nil, fmt.Errorf("invalid inline array type: %s", typeName)
}

// resolveFunctionPointerType handles function pointer type syntax:
//   - "function(Integer): String" → function pointer with Integer param returning String
//   - "procedure(String)" → procedure pointer with String param
//   - "function(): Boolean" → parameterless function returning Boolean
//
// This is used to resolve type annotations stored by the semantic analyzer
// for builtin functions used as function references (e.g., IntToStr in Map).
func (e *Evaluator) resolveFunctionPointerType(typeName string) (types.Type, error) {
	// Determine if it's a function or procedure
	isFunction := strings.HasPrefix(typeName, "function(")
	var prefix string
	if isFunction {
		prefix = "function("
	} else {
		prefix = "procedure("
	}

	// Find the closing paren for parameters
	remaining := strings.TrimPrefix(typeName, prefix)

	// Find matching closing paren (handle nested types)
	parenDepth := 1
	closeIdx := -1
	for i := 0; i < len(remaining); i++ {
		switch remaining[i] {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
			if parenDepth == 0 {
				closeIdx = i
				break
			}
		}
		if closeIdx >= 0 {
			break
		}
	}

	if closeIdx == -1 {
		return nil, fmt.Errorf("invalid function pointer type syntax: %s", typeName)
	}

	// Extract parameter list
	paramsStr := strings.TrimSpace(remaining[:closeIdx])
	afterParams := strings.TrimSpace(remaining[closeIdx+1:])

	// Parse parameters (comma-separated type names)
	var paramTypes []types.Type
	if paramsStr != "" {
		// Split by comma, but be careful about nested types
		paramStrs := splitTypeList(paramsStr)
		for _, paramStr := range paramStrs {
			paramType, err := e.ResolveType(strings.TrimSpace(paramStr))
			if err != nil {
				return nil, fmt.Errorf("invalid parameter type '%s': %w", paramStr, err)
			}
			paramTypes = append(paramTypes, paramType)
		}
	}

	// Parse return type if present (for functions)
	var returnType types.Type
	if isFunction && strings.HasPrefix(afterParams, ":") {
		returnTypeName := strings.TrimSpace(strings.TrimPrefix(afterParams, ":"))
		var err error
		returnType, err = e.ResolveType(returnTypeName)
		if err != nil {
			return nil, fmt.Errorf("invalid return type '%s': %w", returnTypeName, err)
		}
	}

	return types.NewFunctionPointerType(paramTypes, returnType), nil
}

// splitTypeList splits a comma-separated list of types, respecting nested parentheses.
func splitTypeList(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0

	for _, ch := range s {
		switch ch {
		case '(':
			depth++
			current.WriteRune(ch)
		case ')':
			depth--
			current.WriteRune(ch)
		case ',':
			if depth == 0 {
				result = append(result, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	// Don't forget the last part
	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}

// ResolveTypeFromName is an alias for ResolveType for backward compatibility.
// Deprecated: Use ResolveType instead.
func (e *Evaluator) ResolveTypeFromName(typeName string) (types.Type, error) {
	return e.ResolveType(typeName)
}

// ============================================================================
// Context-Aware Type Resolution
// ============================================================================

// ResolveTypeWithContext resolves a type name to a types.Type with environment access.
// This method provides complete direct type resolution without adapter dependency.
//
// Resolution order:
//  1. Built-in types (Integer, Float, String, Boolean, Variant, Const)
//  2. Inline array types (array of X, array[...])
//  3. Named array types (via TypeSystem)
//  4. Enum types (via TypeSystem)
//  5. Record types (via environment: __record_type_<name>)
//  6. Type aliases (via environment: __type_alias_<name>)
//  7. Subrange types (via environment: __subrange_type_<name>)
//
// The lookup is case-insensitive per DWScript semantics.
//
// Returns:
//   - The resolved types.Type
//   - An error if the type cannot be resolved
func (e *Evaluator) ResolveTypeWithContext(typeName string, ctx *ExecutionContext) (types.Type, error) {
	// Step 1: Handle inline array types first
	if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
		return e.resolveInlineArrayTypeWithContext(typeName, ctx)
	}

	// Step 2: Handle inline set types (e.g., "set of TColor")
	// Use case-insensitive prefix matching to respect DWScript semantics.
	lowerTypeName := ident.Normalize(typeName)
	if strings.HasPrefix(lowerTypeName, "set of ") {
		elementTypeName := strings.TrimSpace(lowerTypeName[len("set of "):])
		elementType, err := e.ResolveTypeWithContext(elementTypeName, ctx)
		if err != nil {
			return nil, fmt.Errorf("invalid set element type: %w", err)
		}
		return types.NewSetType(elementType), nil
	}

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
	case "jsonvariant":
		return types.JSON_VARIANT, nil
	case "const":
		// "Const" redirects to VARIANT for dynamic typing
		return types.VARIANT, nil
	}

	// Step 4: Check named array types via TypeSystem (direct access)
	if e.typeSystem != nil {
		if arrayType := e.typeSystem.LookupArrayType(typeName); arrayType != nil {
			return arrayType, nil
		}

		// Step 5: Check enum types via TypeSystem
		if enumMetadata := e.typeSystem.LookupEnumMetadata(typeName); enumMetadata != nil {
			if etv, ok := enumMetadata.(EnumTypeValueAccessor); ok {
				return etv.GetEnumType(), nil
			}
		}
	}

	// Step 6: Check record types in environment
	if recordTypeVal, ok := ctx.Env().Get("__record_type_" + lowerTypeName); ok {
		if rtv, ok := recordTypeVal.(interface{ GetRecordType() *types.RecordType }); ok {
			return rtv.GetRecordType(), nil
		}
	}

	// Step 6b: Check named set types in environment
	if setTypeVal, ok := ctx.Env().Get("__set_type_" + lowerTypeName); ok {
		if stv, ok := setTypeVal.(interface{ GetSetType() *types.SetType }); ok {
			return stv.GetSetType(), nil
		}
	}

	// Step 7: Check type aliases in environment
	if typeAliasVal, ok := ctx.Env().Get("__type_alias_" + lowerTypeName); ok {
		if tav, ok := typeAliasVal.(interface{ GetAliasedType() types.Type }); ok {
			return tav.GetAliasedType(), nil
		}
	}

	// Step 8: Check subrange types in TypeSystem
	if subrangeType := e.typeSystem.LookupSubrangeType(typeName); subrangeType != nil {
		return subrangeType, nil
	}

	// Step 9: Check class and interface types in TypeSystem. Class type
	// strings may carry a parent qualification ("TSub(TBase)") - strip it.
	if e.typeSystem != nil {
		cleanName := typeName
		if idx := strings.Index(cleanName, "("); idx != -1 {
			cleanName = strings.TrimSpace(cleanName[:idx])
		}
		if e.typeSystem.HasClass(cleanName) {
			return e.buildClassTypeWithHierarchy(cleanName), nil
		}
		if e.typeSystem.HasInterface(cleanName) {
			return types.NewInterfaceType(cleanName), nil
		}
	}

	// Step 10: Named function/method pointer types (type TProc = procedure;).
	if e.typeSystem != nil {
		if funcPtrType := e.typeSystem.LookupFunctionPointerType(typeName); funcPtrType != nil {
			return funcPtrType, nil
		}
	}

	// Unknown type
	return nil, fmt.Errorf("unknown type: %s", typeName)
}

// resolveInlineArrayTypeWithContext handles inline array type syntax with context:
//   - "array of Integer" → dynamic array
//   - "array[1..10] of String" → static array
func (e *Evaluator) resolveInlineArrayTypeWithContext(typeName string, ctx *ExecutionContext) (types.Type, error) {
	// Handle "array of ElementType" (dynamic array)
	if strings.HasPrefix(typeName, "array of ") {
		elementTypeName := strings.TrimPrefix(typeName, "array of ")
		elementType, err := e.ResolveTypeWithContext(elementTypeName, ctx)
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

		boundsStr := strings.TrimSpace(typeName[6:ofIdx]) // Extract "low..high" or index type
		elementTypeName := strings.TrimSpace(typeName[ofIdx+5:])

		// Parse bounds
		parts := strings.Split(boundsStr, "..")
		var low, high int
		var hasOrdinalBounds bool

		if len(parts) == 2 {
			_, err := fmt.Sscanf(parts[0], "%d", &low)
			if err != nil {
				return nil, fmt.Errorf("invalid low bound: %s", parts[0])
			}
			_, err = fmt.Sscanf(parts[1], "%d", &high)
			if err != nil {
				return nil, fmt.Errorf("invalid high bound: %s", parts[1])
			}
		} else {
			indexType, err := e.ResolveTypeWithContext(boundsStr, ctx)
			if err != nil {
				return nil, fmt.Errorf("invalid array index type: %w", err)
			}

			ordLow, ordHigh, ok := types.OrdinalBounds(indexType)
			if !ok {
				return nil, fmt.Errorf("array index type '%s' must be a bounded ordinal", boundsStr)
			}

			low, high = ordLow, ordHigh
			hasOrdinalBounds = true
		}

		// Resolve element type using context-aware method
		elementType, err := e.ResolveTypeWithContext(elementTypeName, ctx)
		if err != nil {
			return nil, fmt.Errorf("invalid array element type: %w", err)
		}

		if len(parts) != 2 && !hasOrdinalBounds {
			return nil, fmt.Errorf("invalid array bounds: %s", boundsStr)
		}

		return types.NewStaticArrayType(elementType, low, high), nil
	}

	return nil, fmt.Errorf("invalid inline array type: %s", typeName)
}

// ============================================================================
// Type Annotation Resolution
// ============================================================================

// ResolveTypeFromAnnotation resolves a type from an AST TypeExpression.
// This is used for function return types, parameter types, and variable declarations.
func (e *Evaluator) ResolveTypeFromAnnotation(typeExpr ast.TypeExpression) (types.Type, error) {
	if typeExpr == nil {
		return nil, nil
	}

	switch node := typeExpr.(type) {
	case *ast.TypeAnnotation:
		if node.InlineType != nil {
			return e.ResolveTypeFromAnnotation(node.InlineType)
		}
	case *ast.ClassOfTypeNode:
		baseClassName := ""
		if node.ClassType != nil {
			baseClassName = node.ClassType.String()
		}
		if baseClassName == "" || !e.typeSystem.HasClass(baseClassName) {
			return nil, fmt.Errorf("unknown class type '%s'", baseClassName)
		}
		return types.NewClassOfType(types.NewClassType(baseClassName, nil)), nil
	case *ast.RecordTypeNode:
		return e.resolveRecordTypeNode(node)
	case *ast.ArrayTypeNode:
		ctx := e.currentContext
		if ctx == nil {
			ctx = &ExecutionContext{}
		}
		if assocType := e.resolveAssociativeArrayTypeNode(node, ctx); assocType != nil {
			return assocType, nil
		}
		arrayType := e.resolveArrayTypeNode(node, ctx)
		if arrayType == nil {
			return nil, fmt.Errorf("invalid array type")
		}
		return arrayType, nil
	}

	// Get the type name string from the expression
	typeName := typeExpr.String()

	// Delegate to ResolveType which handles all type resolution
	return e.ResolveType(typeName)
}

func (e *Evaluator) resolveRecordTypeNode(recordNode *ast.RecordTypeNode) (types.Type, error) {
	recordType := types.NewRecordType("", make(map[string]types.Type))

	for _, field := range recordNode.Fields {
		fieldName := field.Name.Value
		fieldKey := ident.Normalize(fieldName)
		if _, exists := recordType.Fields[fieldKey]; exists {
			return nil, fmt.Errorf("field '%s' already exists in inline record", fieldName)
		}

		var fieldType types.Type
		if field.Type != nil {
			resolved, err := e.ResolveTypeFromAnnotation(field.Type)
			if err != nil {
				return nil, err
			}
			fieldType = resolved
		} else if field.InitValue != nil {
			value := e.Eval(field.InitValue, e.currentContext)
			if isError(value) {
				return nil, fmt.Errorf("%s", value.String())
			}
			fieldType = e.getValueType(value)
		}
		if fieldType == nil {
			return nil, fmt.Errorf("field '%s' in inline record must have either a type or initializer", fieldName)
		}

		recordType.AddField(fieldName, fieldType, field.InitValue != nil)
	}

	for _, prop := range recordNode.Properties {
		propType, err := e.ResolveTypeFromAnnotation(prop.Type)
		if err != nil {
			return nil, err
		}
		propKey := ident.Normalize(prop.Name.Value)
		recordType.Properties[propKey] = &types.RecordPropertyInfo{
			Name:       prop.Name.Value,
			Type:       propType,
			ReadField:  prop.ReadField,
			WriteField: prop.WriteField,
			IsDefault:  prop.IsDefault,
			IsIndexed:  len(prop.IndexParams) > 0,
		}
	}

	return recordType, nil
}

// ============================================================================
// Default Value Creation
// ============================================================================

// GetDefaultValue returns the default/zero value for a given type.
// This is used for Result variable initialization in functions.
func (e *Evaluator) GetDefaultValue(typ types.Type) Value {
	if typ == nil {
		return e.nilValue()
	}
	typ = types.GetUnderlyingType(typ)

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
		// Records are value types and must be zero-initialized, especially when
		// dynamic arrays grow via SetLength and allocate new record elements.
		return e.getZeroValueForType(typ)
	case "VARIANT":
		// Variants default to Unassigned (nil-like)
		return e.nilValue()
	default:
		// Unknown types default to NIL
		return e.nilValue()
	}
}

// nilValue returns a nil value.
func (e *Evaluator) nilValue() Value {
	return &runtime.NilValue{}
}
