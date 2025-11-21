package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// parseInlineArrayType parses a DWScript inline array type signature (static or dynamic)
// from a string, extracting bounds and element type information.
func (i *Interpreter) parseInlineArrayType(signature string) *types.ArrayType {
	var lowBound, highBound *int

	// Check if this is a static array with bounds
	if strings.HasPrefix(signature, "array[") {
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

		// Skip past "] of "
		signature = signature[endBracket+1:]
	} else if strings.HasPrefix(signature, "array of ") {
		// Dynamic array: skip "array" to get " of ElementType"
		signature = signature[5:] // Skip "array"
	} else {
		return nil
	}

	// Now signature should be " of ElementType"
	if !strings.HasPrefix(signature, " of ") {
		return nil
	}

	// Extract element type name
	elementTypeName := strings.TrimSpace(signature[4:]) // Skip " of "

	// Get the element type (resolveType handles recursion for nested arrays)
	elementType, err := i.resolveType(elementTypeName)
	if err != nil || elementType == nil {
		return nil
	}

	// Create array type
	if lowBound != nil && highBound != nil {
		return types.NewStaticArrayType(elementType, *lowBound, *highBound)
	}
	return types.NewDynamicArrayType(elementType)
}

// parseInlineSetType parses inline set type syntax like "set of TEnumType".
// Returns the SetType, or nil if the string doesn't match the expected format.
func (i *Interpreter) parseInlineSetType(signature string) *types.SetType {
	// Check for "set of " prefix
	if !strings.HasPrefix(signature, "set of ") {
		return nil
	}

	// Extract enum type name: "set of TColor" → "TColor"
	enumTypeName := strings.TrimSpace(signature[7:]) // Skip "set of "
	if enumTypeName == "" {
		return nil
	}

	// Look up the enum type in the environment
	// Enum types are stored with "__enum_type_" prefix
	typeKey := "__enum_type_" + strings.ToLower(enumTypeName)
	typeVal, ok := i.env.Get(typeKey)
	if !ok {
		return nil
	}

	// Extract the EnumType from the EnumTypeValue
	enumTypeVal, ok := typeVal.(*EnumTypeValue)
	if !ok {
		return nil
	}

	// Create and return the set type
	return types.NewSetType(enumTypeVal.EnumType)
}

// resolveArrayTypeNode resolves an ArrayTypeNode directly from the AST.
// This avoids string conversion issues with parentheses in bound expressions like (-5).
// Task: Fix negative array bounds like array[-5..5]
func (i *Interpreter) resolveArrayTypeNode(arrayNode *ast.ArrayTypeNode) *types.ArrayType {
	if arrayNode == nil {
		return nil
	}

	// Resolve element type first
	var elementType types.Type

	// Check if element type is also an array (nested arrays)
	if nestedArray, ok := arrayNode.ElementType.(*ast.ArrayTypeNode); ok {
		elementType = i.resolveArrayTypeNode(nestedArray)
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
		elementType, err = i.resolveType(elementTypeName)
		if err != nil || elementType == nil {
			return nil
		}
	}

	// Check if dynamic or static array
	if arrayNode.IsDynamic() {
		return types.NewDynamicArrayType(elementType)
	}

	// Static array - evaluate bounds by interpreting the expressions
	// For constant expressions (literals, unary minus), we can evaluate them directly
	lowVal := i.Eval(arrayNode.LowBound)
	if isError(lowVal) {
		return nil
	}
	lowBound, ok := lowVal.(*IntegerValue)
	if !ok {
		return nil
	}

	highVal := i.Eval(arrayNode.HighBound)
	if isError(highVal) {
		return nil
	}
	highBound, ok := highVal.(*IntegerValue)
	if !ok {
		return nil
	}

	return types.NewStaticArrayType(elementType, int(lowBound.Value), int(highBound.Value))
}

// resolveOverload selects the best matching function overload based on argument types.
func (i *Interpreter) resolveOverload(funcName string, overloads []*ast.FunctionDecl, argExprs []ast.Expression) (*ast.FunctionDecl, []Value, error) {
	// Fast path: if only one overload, check for lazy parameters and skip evaluation
	if len(overloads) == 1 {
		fn := overloads[0]
		argValues := make([]Value, len(argExprs))
		for idx, argExpr := range argExprs {
			// Check if this parameter is lazy
			isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
			if isLazy {
				// Don't evaluate lazy parameters - mark as nil
				argValues[idx] = nil
			} else {
				// Evaluate non-lazy parameters
				val := i.Eval(argExpr)
				if isError(val) {
					return nil, nil, fmt.Errorf("error evaluating argument %d: %v", idx+1, val)
				}
				argValues[idx] = val
			}
		}
		return fn, argValues, nil
	}

	// Multiple overloads: evaluate all arguments for type checking
	// Note: For overloaded functions, lazy parameters will be evaluated twice
	// (once here for overload resolution, once when accessed). This is a known limitation.
	argTypes := make([]types.Type, len(argExprs))
	argValues := make([]Value, len(argExprs))
	for idx, argExpr := range argExprs {
		// Evaluate the argument to get its value and type
		val := i.Eval(argExpr)
		if isError(val) {
			return nil, nil, fmt.Errorf("error evaluating argument %d: %v", idx+1, val)
		}
		argTypes[idx] = i.getValueType(val)
		argValues[idx] = val
	}

	// Convert function declarations to semantic symbols for resolution
	// We need to extract the function types from the AST nodes
	candidates := make([]*semantic.Symbol, len(overloads))
	for idx, fn := range overloads {
		funcType := i.extractFunctionType(fn)
		if funcType == nil {
			return nil, nil, fmt.Errorf("unable to extract function type for overload %d of '%s'", idx+1, funcName)
		}

		candidates[idx] = &semantic.Symbol{
			Name:                 fn.Name.Value,
			Type:                 funcType,
			HasOverloadDirective: fn.IsOverload,
		}
	}

	// Use semantic analyzer's overload resolution
	selected, err := semantic.ResolveOverload(candidates, argTypes)
	if err != nil {
		// Use DWScript-compatible error message
		return nil, nil, fmt.Errorf("There is no overloaded version of \"%s\" that can be called with these arguments", funcName)
	}

	// Find which function declaration corresponds to the selected symbol
	// We can match by comparing the function types
	selectedType := selected.Type.(*types.FunctionType)
	for _, fn := range overloads {
		fnType := i.extractFunctionType(fn)
		if fnType != nil && semantic.SignaturesEqual(fnType, selectedType) &&
			fnType.ReturnType.Equals(selectedType.ReturnType) {
			return fn, argValues, nil
		}
	}

	return nil, nil, fmt.Errorf("internal error: resolved overload not found in candidate list")
}

// extractFunctionType extracts a types.FunctionType from an ast.FunctionDecl
// Helper for overload resolution
func (i *Interpreter) extractFunctionType(fn *ast.FunctionDecl) *types.FunctionType {
	paramTypes := make([]types.Type, len(fn.Parameters))
	paramNames := make([]string, len(fn.Parameters))
	lazyParams := make([]bool, len(fn.Parameters))
	varParams := make([]bool, len(fn.Parameters))
	constParams := make([]bool, len(fn.Parameters))
	defaultValues := make([]interface{}, len(fn.Parameters))

	for idx, param := range fn.Parameters {
		if param.Type == nil {
			return nil // Invalid function
		}

		paramType, err := i.resolveType(param.Type.String())
		if err != nil {
			return nil
		}

		paramTypes[idx] = paramType
		paramNames[idx] = param.Name.Value
		lazyParams[idx] = param.IsLazy
		varParams[idx] = param.ByRef
		constParams[idx] = param.IsConst
		defaultValues[idx] = param.DefaultValue
	}

	var returnType types.Type
	if fn.ReturnType != nil {
		var err error
		returnType, err = i.resolveType(fn.ReturnType.String())
		if err != nil {
			returnType = types.VOID
		}
	} else {
		returnType = types.VOID
	}

	return types.NewFunctionTypeWithMetadata(
		paramTypes, paramNames, defaultValues,
		lazyParams, varParams, constParams,
		returnType,
	)
}

// getValueType returns the types.Type for a runtime Value
// Helper for overload resolution
func (i *Interpreter) getValueType(val Value) types.Type {
	switch v := val.(type) {
	case *IntegerValue:
		return types.INTEGER
	case *FloatValue:
		return types.FLOAT
	case *StringValue:
		return types.STRING
	case *BooleanValue:
		return types.BOOLEAN
	case *NilValue:
		return types.NIL
	case *VariantValue:
		return types.VARIANT
	case *RecordValue:
		if v.RecordType != nil {
			return v.RecordType
		}
		return types.NIL
	default:
		// For objects, arrays, and other complex types, try AsObject
		if obj, ok := AsObject(val); ok && obj.Class != nil {
			// TODO: Create custom type from class name when needed
			return types.NIL
		}
		// For arrays and other types
		return types.NIL
	}
}

// evalTypeCast evaluates a type cast expression like Integer(x), Float(y), or TMyClass(obj).
// Returns the cast value if this is a valid type cast, or nil if not a type cast.
func (i *Interpreter) evalTypeCast(typeName string, arg ast.Expression) Value {
	// First check if this is actually a type cast before evaluating the argument
	// This prevents double evaluation when it's not a type cast
	isTypeCast := false
	var enumType *types.EnumType
	lowerName := strings.ToLower(typeName)

	// Check if it's a built-in type
	switch lowerName {
	case "integer", "float", "string", "boolean", "variant":
		isTypeCast = true
	default:
		// Check if it's a class/interface type
		if i.lookupClassInfo(typeName) != nil {
			isTypeCast = true
		} else {
			// Task 9.15.6: Check if it's an enum type
			enumTypeKey := "__enum_type_" + lowerName
			if typeVal, ok := i.env.Get(enumTypeKey); ok {
				if etv, ok := typeVal.(*EnumTypeValue); ok {
					enumType = etv.EnumType
					isTypeCast = true
				}
			}
		}
	}

	// If it's not a type cast, return nil without evaluating
	if !isTypeCast {
		return nil
	}

	// Now evaluate the argument since we know it's a type cast
	val := i.Eval(arg)
	if isError(val) {
		return val
	}

	// Perform the type cast
	switch lowerName {
	case "integer":
		return i.castToInteger(val)
	case "float":
		return i.castToFloat(val)
	case "string":
		return i.castToString(val)
	case "boolean":
		return i.castToBoolean(val)
	case "variant":
		// Variant can accept any value
		return &VariantValue{Value: val}
	default:
		// Task 9.15.6: Check if it's an enum type
		if enumType != nil {
			return i.castToEnum(val, enumType, typeName)
		}
		// Must be a class/interface type (we already checked above)
		classInfo := i.lookupClassInfo(typeName)
		return i.castToClass(val, classInfo, arg)
	}
}

// lookupClassInfo looks up a class by name (case-insensitive)
func (i *Interpreter) lookupClassInfo(name string) *ClassInfo {
	for className, classInfo := range i.classes {
		if strings.EqualFold(className, name) {
			return classInfo
		}
	}
	return nil
}

// castToInteger converts a value to Integer
func (i *Interpreter) castToInteger(val Value) Value {
	switch v := val.(type) {
	case *IntegerValue:
		return v
	case *FloatValue:
		// DWScript Integer() truncates toward zero
		return &IntegerValue{Value: int64(v.Value)}
	case *BooleanValue:
		if v.Value {
			return &IntegerValue{Value: 1}
		}
		return &IntegerValue{Value: 0}
	case *StringValue:
		// Try to parse string as integer
		var result int64
		_, err := fmt.Sscanf(v.Value, "%d", &result)
		if err != nil {
			return newError("cannot convert string '%s' to Integer", v.Value)
		}
		return &IntegerValue{Value: result}
	case *EnumValue:
		// Cast enum to its ordinal value
		return &IntegerValue{Value: int64(v.OrdinalValue)}
	case *VariantValue:
		// Recursively cast the variant's value
		return i.castToInteger(v.Value)
	default:
		return newError("cannot cast %s to Integer", val.Type())
	}
}

// castToFloat converts a value to Float
func (i *Interpreter) castToFloat(val Value) Value {
	switch v := val.(type) {
	case *FloatValue:
		return v
	case *IntegerValue:
		return &FloatValue{Value: float64(v.Value)}
	case *BooleanValue:
		if v.Value {
			return &FloatValue{Value: 1.0}
		}
		return &FloatValue{Value: 0.0}
	case *StringValue:
		// Try to parse string as float
		var result float64
		_, err := fmt.Sscanf(v.Value, "%f", &result)
		if err != nil {
			return newError("cannot convert string '%s' to Float", v.Value)
		}
		return &FloatValue{Value: result}
	case *EnumValue:
		// Cast enum to its ordinal value as float
		return &FloatValue{Value: float64(v.OrdinalValue)}
	case *VariantValue:
		// Recursively cast the variant's value
		return i.castToFloat(v.Value)
	default:
		return newError("cannot cast %s to Float", val.Type())
	}
}

// castToString converts a value to String
func (i *Interpreter) castToString(val Value) Value {
	switch v := val.(type) {
	case *StringValue:
		return v
	case *IntegerValue:
		return &StringValue{Value: fmt.Sprintf("%d", v.Value)}
	case *FloatValue:
		return &StringValue{Value: fmt.Sprintf("%g", v.Value)}
	case *BooleanValue:
		if v.Value {
			return &StringValue{Value: "True"}
		}
		return &StringValue{Value: "False"}
	case *VariantValue:
		// Recursively cast the variant's value
		return i.castToString(v.Value)
	default:
		// For other types, use their String() representation
		return &StringValue{Value: val.String()}
	}
}

// castToBoolean converts a value to Boolean
func (i *Interpreter) castToBoolean(val Value) Value {
	switch v := val.(type) {
	case *BooleanValue:
		return v
	case *IntegerValue:
		return &BooleanValue{Value: v.Value != 0}
	case *FloatValue:
		return &BooleanValue{Value: v.Value != 0.0}
	case *StringValue:
		// Parse string to boolean (DWScript semantics)
		// Recognized as true: "1", "T", "t", "Y", "y", "yes", "true" (case-insensitive)
		// Everything else is false
		s := strings.TrimSpace(v.Value)
		if s == "" {
			return &BooleanValue{Value: false}
		}
		// Check single character shortcuts
		if len(s) == 1 {
			switch s[0] {
			case '1', 'T', 't', 'Y', 'y':
				return &BooleanValue{Value: true}
			}
			return &BooleanValue{Value: false}
		}
		// Check multi-character strings (case-insensitive)
		if strings.EqualFold(s, "yes") || strings.EqualFold(s, "true") {
			return &BooleanValue{Value: true}
		}
		return &BooleanValue{Value: false}
	case *VariantValue:
		// Recursively cast the variant's value
		return i.castToBoolean(v.Value)
	default:
		return newError("cannot cast %s to Boolean", val.Type())
	}
}

// castToClass performs runtime type checking and casts to a class type
func (i *Interpreter) castToClass(val Value, targetClass *ClassInfo, node ast.Node) Value {
	// Unwrap variant if needed
	if variantVal, ok := val.(*VariantValue); ok {
		val = variantVal.Value
	}

	// Unwrap TypeCastValue if needed (for successive casts like TBase(obj1) then TObject(obj2))
	// This preserves support for successive type casts: obj := TObject(child); TBase(obj)
	if typeCast, ok := val.(*TypeCastValue); ok {
		val = typeCast.Object
	}

	// Handle nil - wrap it with the static type for proper class variable access
	if _, isNil := val.(*NilValue); isNil {
		// Wrap nil in TypeCastValue to preserve static type information
		// This allows TBase(nilChild).ClassVar to access TBase's class variable
		return &TypeCastValue{
			Object:     val,
			StaticType: targetClass,
		}
	}

	// Get the object
	obj, ok := AsObject(val)
	if !ok {
		return i.newErrorWithLocation(node, "cannot cast %s to %s: not an object", val.Type(), targetClass.Name)
	}

	// Check if the object's class is compatible with the target class
	// The object must be an instance of the target class or a derived class
	if !obj.IsInstanceOf(targetClass) {
		pos := node.Pos()
		message := fmt.Sprintf("instance of type \"%s\" cannot be cast to class \"%s\" [line: %d, column: %d]",
			obj.Class.Name, targetClass.Name, pos.Line, pos.Column)
		i.raiseException("Exception", message, &pos)
		return nil
	}

	// Cast is valid - return a TypeCastValue that preserves the static type
	// This is crucial for class variable access: TBase(child).ClassVar should access TBase's class variable
	return &TypeCastValue{
		Object:     val,
		StaticType: targetClass,
	}
}

// castToEnum casts a value to an enum type.
// Task 9.15.6: Supports Integer → Enum and Enum → Enum (same type) casting.
func (i *Interpreter) castToEnum(val Value, targetEnum *types.EnumType, typeName string) Value {
	switch v := val.(type) {
	case *IntegerValue:
		// Integer → Enum: Create an EnumValue with the integer as ordinal
		// Find the enum value name for this ordinal (if it exists)
		ordinal := int(v.Value)
		var valueName string

		// Look up the name for this ordinal value
		for name, ord := range targetEnum.Values {
			if ord == ordinal {
				valueName = name
				break
			}
		}

		// If no matching name found, create a placeholder name using the ordinal value
		// (DWScript allows casting any integer to enum, even if not a valid ordinal)
		if valueName == "" && len(targetEnum.OrderedNames) > 0 {
			// For out-of-bounds ordinals, we still create an EnumValue
			// but with a placeholder name (DWScript behavior)
			valueName = fmt.Sprintf("$%d", ordinal)
		}

		return &EnumValue{
			TypeName:     typeName,
			ValueName:    valueName,
			OrdinalValue: ordinal,
		}

	case *EnumValue:
		// Enum → Enum: Only allow identity cast (same type)
		if strings.EqualFold(v.TypeName, typeName) {
			return v
		}
		return newError("cannot cast enum %s to %s: incompatible enum types", v.TypeName, typeName)

	case *VariantValue:
		// Recursively cast the variant's value
		return i.castToEnum(v.Value, targetEnum, typeName)

	default:
		return newError("cannot cast %s to enum %s", val.Type(), typeName)
	}
}

// evalDefaultFunction handles the Default() built-in function which expects an unevaluated type identifier.
// Default(Integer) should pass "Integer" as a string, not evaluate it as a variable.
// Returns the default/zero value for the specified type, or nil if not a valid type.
func (i *Interpreter) evalDefaultFunction(arg ast.Expression) Value {
	// The argument should be a type identifier
	ident, ok := arg.(*ast.Identifier)
	if !ok {
		return i.newErrorWithLocation(arg, "Default() expects a type name as argument")
	}

	typeName := ident.Value
	lowerName := strings.ToLower(typeName)

	// Return default values based on type name
	switch lowerName {
	case "integer", "int64", "byte", "word", "cardinal", "smallint", "shortint", "longword":
		return &IntegerValue{Value: 0}
	case "float", "double", "single", "extended", "currency":
		return &FloatValue{Value: 0.0}
	case "string", "unicodestring", "ansistring":
		return &StringValue{Value: ""}
	case "boolean":
		return &BooleanValue{Value: false}
	case "variant":
		return &NilValue{}
	default:
		// For class types, records, enums, and other reference/complex types, return nil
		// Check if it's a valid type by looking it up
		// For now, return nil (which represents the default value for reference types)
		return &NilValue{}
	}
}
