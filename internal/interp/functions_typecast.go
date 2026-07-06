package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
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

// resolveInlineFunctionPointerType parses an inline function or method pointer signature.
// Mirrors the semantic analyzer's implementation but uses interpreter type resolution.
func (i *Interpreter) resolveInlineFunctionPointerType(signature string) (types.Type, error) {
	// Check if this is a method pointer ("of object")
	ofObject := strings.HasSuffix(signature, " of object")
	if ofObject {
		signature = strings.TrimSuffix(signature, " of object")
		signature = strings.TrimSpace(signature)
	}

	// Determine if it's a function or procedure
	isFunction := strings.HasPrefix(signature, "function(")

	// Extract the part between ( and )
	openParen := strings.Index(signature, "(")
	closeParen := strings.LastIndex(signature, ")")
	if openParen == -1 || closeParen == -1 || closeParen < openParen {
		return nil, fmt.Errorf("invalid function pointer signature: %s", signature)
	}

	// Extract parameters string
	paramsStr := signature[openParen+1 : closeParen]

	// Parse parameters
	paramTypes, err := i.parseInlineParameters(paramsStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing parameters in '%s': %w", signature, err)
	}

	// Extract return type (if function)
	var returnType types.Type
	if isFunction {
		// Look for ": ReturnType" after the closing )
		remainder := strings.TrimSpace(signature[closeParen+1:])
		if strings.HasPrefix(remainder, ":") {
			returnTypeName := strings.TrimSpace(remainder[1:])
			if returnTypeName != "" {
				returnType, err = i.resolveType(returnTypeName)
				if err != nil {
					return nil, fmt.Errorf("unknown return type '%s' in function pointer", returnTypeName)
				}
			}
		}
	}

	// Create function pointer type
	if ofObject {
		return types.NewMethodPointerType(paramTypes, returnType), nil
	}
	return types.NewFunctionPointerType(paramTypes, returnType), nil
}

// parseInlineParameters parses the parameter list from an inline function pointer signature.
// Supports both named and shorthand parameter formats.
func (i *Interpreter) parseInlineParameters(paramsStr string) ([]types.Type, error) {
	paramsStr = strings.TrimSpace(paramsStr)
	if paramsStr == "" {
		return []types.Type{}, nil
	}

	hasColon := strings.Contains(paramsStr, ":")

	if !hasColon {
		// Shorthand format: just types, no names
		return i.parseShorthandParameters(paramsStr)
	}

	paramTypes := []types.Type{}

	// Split by semicolon to get parameter groups
	groups := strings.Split(paramsStr, ";")

	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}

		// Split by colon to separate names from type
		parts := strings.Split(group, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid parameter group: %s", group)
		}

		// Get the type name
		typeName := strings.TrimSpace(parts[1])

		// Resolve the type
		paramType, err := i.resolveType(typeName)
		if err != nil {
			return nil, fmt.Errorf("unknown parameter type '%s'", typeName)
		}

		// Count how many parameters have this type (by counting commas + 1)
		namesStr := strings.TrimSpace(parts[0])
		namesStr = strings.TrimPrefix(namesStr, "const ")
		namesStr = strings.TrimPrefix(namesStr, "var ")
		namesStr = strings.TrimPrefix(namesStr, "lazy ")
		namesStr = strings.TrimSpace(namesStr)

		numParams := 1
		if namesStr != "" {
			numParams = strings.Count(namesStr, ",") + 1
		}

		for j := 0; j < numParams; j++ {
			paramTypes = append(paramTypes, paramType)
		}
	}

	return paramTypes, nil
}

// parseShorthandParameters parses shorthand parameter syntax (types only, no names).
// Format: "Type1, Type2, ..." or "Type1; Type2; ..."
func (i *Interpreter) parseShorthandParameters(paramsStr string) ([]types.Type, error) {
	paramTypes := []types.Type{}

	// Split by both comma and semicolon
	paramsStr = strings.ReplaceAll(paramsStr, ";", ",")

	typeNames := strings.Split(paramsStr, ",")

	for _, typeName := range typeNames {
		typeName = strings.TrimSpace(typeName)
		if typeName == "" {
			continue
		}

		// Remove modifiers if present
		typeName = strings.TrimPrefix(typeName, "const ")
		typeName = strings.TrimPrefix(typeName, "var ")
		typeName = strings.TrimPrefix(typeName, "lazy ")
		typeName = strings.TrimSpace(typeName)

		// Resolve the type
		paramType, err := i.resolveType(typeName)
		if err != nil {
			return nil, fmt.Errorf("unknown parameter type '%s'", typeName)
		}

		paramTypes = append(paramTypes, paramType)
	}

	return paramTypes, nil
}

// lookupClassInfo looks up a class by name (case-insensitive)
func (i *Interpreter) lookupClassInfo(name string) *ClassInfo {
	return i.lookupRegisteredClassInfo(name)
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
		if pkgident.Equal(s, "yes") || pkgident.Equal(s, "true") {
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
		message := fmt.Sprintf("Cannot cast instance of type \"%s\" to class \"%s\" [line: %d, column: %d]",
			obj.Class.GetName(), targetClass.Name, pos.Line, pos.Column)
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
