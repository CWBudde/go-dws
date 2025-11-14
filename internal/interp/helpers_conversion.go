package interp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Type Resolution, Conversion, and Default Value Helpers
// ============================================================================

// getDefaultValue returns the default/zero value for a given type.
// This is used for Result variable initialization in functions.
func (i *Interpreter) getDefaultValue(typ types.Type) Value {
	if typ == nil {
		return &NilValue{}
	}

	switch typ.TypeKind() {
	case "STRING":
		return &StringValue{Value: ""}
	case "INTEGER":
		return &IntegerValue{Value: 0}
	case "FLOAT":
		return &FloatValue{Value: 0.0}
	case "BOOLEAN":
		return &BooleanValue{Value: false}
	case "CLASS", "INTERFACE", "FUNCTION_POINTER", "METHOD_POINTER":
		return &NilValue{}
	case "ARRAY":
		// Dynamic arrays default to NIL
		return &NilValue{}
	case "RECORD":
		// Records should be initialized with default field values
		// For now, return NIL (will be enhanced in future tasks if needed)
		return &NilValue{}
	default:
		// Unknown types default to NIL
		return &NilValue{}
	}
}

// resolveTypeFromExpression resolves a type from any TypeExpression.
func (i *Interpreter) resolveTypeFromExpression(typeExpr ast.TypeExpression) types.Type {
	if typeExpr == nil {
		return nil
	}

	// For simple type annotations, delegate to existing function
	if typeAnnot, ok := typeExpr.(*ast.TypeAnnotation); ok {
		return i.resolveTypeFromAnnotation(typeAnnot)
	}

	// For array types, resolve the element type and construct an array type
	if arrayType, ok := typeExpr.(*ast.ArrayTypeNode); ok {
		elementType := i.resolveTypeFromExpression(arrayType.ElementType)
		if elementType == nil {
			return nil
		}

		// Evaluate bound expressions if this is a static array
		if arrayType.IsDynamic() {
			return types.NewDynamicArrayType(elementType)
		}

		// Evaluate low bound
		lowBoundVal := i.Eval(arrayType.LowBound)
		if isError(lowBoundVal) {
			return nil
		}
		lowBound, ok := lowBoundVal.(*IntegerValue)
		if !ok {
			return nil
		}

		// Evaluate high bound
		highBoundVal := i.Eval(arrayType.HighBound)
		if isError(highBoundVal) {
			return nil
		}
		highBound, ok := highBoundVal.(*IntegerValue)
		if !ok {
			return nil
		}

		return types.NewStaticArrayType(elementType, int(lowBound.Value), int(highBound.Value))
	}

	// For function pointer types, we need full type information
	// For now, return a generic function type placeholder
	if _, ok := typeExpr.(*ast.FunctionPointerTypeNode); ok {
		// TODO: Properly construct function pointer type
		return types.NewFunctionType([]types.Type{}, nil)
	}

	return nil
}

// resolveTypeFromAnnotation resolves a type from an AST TypeAnnotation
func (i *Interpreter) resolveTypeFromAnnotation(typeAnnot *ast.TypeAnnotation) types.Type {
	if typeAnnot == nil {
		return nil
	}

	typeName := typeAnnot.Name

	// Normalize type name to lowercase for case-insensitive comparison
	// DWScript (like Pascal) is case-insensitive for all identifiers including type names
	lowerTypeName := strings.ToLower(typeName)

	// Check basic types (case-insensitive)
	switch lowerTypeName {
	case "integer":
		return types.INTEGER
	case "float":
		return types.FLOAT
	case "string":
		return types.STRING
	case "boolean":
		return types.BOOLEAN
	case "const":
		// Migrate Const to Variant for proper dynamic typing
		// "Const" was a temporary workaround, now redirects to VARIANT
		return types.VARIANT
	case "variant":
		// Support Variant type for dynamic values
		return types.VARIANT
	}

	// Check for class types (stored in i.classes map)
	// Preserve original case for custom type lookup
	if classInfo, ok := i.classes[typeName]; ok {
		return types.NewClassType(classInfo.Name, nil)
	}

	// Check for record types (stored with special prefix in environment)
	// Normalize to lowercase for case-insensitive lookups
	recordTypeKey := "__record_type_" + strings.ToLower(typeName)
	if typeVal, ok := i.env.Get(recordTypeKey); ok {
		if recordTypeVal, ok := typeVal.(*RecordTypeValue); ok {
			return recordTypeVal.RecordType
		}
	}

	// Type not found
	return nil
}

// extractSimpleTypeName extracts the simple type name from a full type string
// Examples:
//   - "array of Integer" -> "Integer"
//   - "array[0..10] of String" -> "String"
//   - "MyClass(ParentClass)" -> "MyClass"
//   - "class of MyClass" -> "MyClass"
//   - "Integer" -> "Integer"
func extractSimpleTypeName(typeName string) string {
	// Handle array types: "array of Integer" or "array[0..10] of String"
	if strings.HasPrefix(typeName, "array") {
		if idx := strings.Index(typeName, " of "); idx != -1 {
			return typeName[idx+4:] // Extract everything after " of "
		}
	}

	// Handle metaclass types: "class of MyClass"
	if strings.HasPrefix(typeName, "class of ") {
		return typeName[9:] // Extract everything after "class of "
	}

	// Handle class types with parent: "MyClass(ParentClass)"
	if idx := strings.Index(typeName, "("); idx != -1 {
		return typeName[:idx] // Extract everything before "("
	}

	// Already a simple type name
	return typeName
}

// ============================================================================
// Built-in Helper Method Implementations
// ============================================================================

// evalBuiltinHelperMethod executes a built-in helper method implementation identified by spec.
func (i *Interpreter) evalBuiltinHelperMethod(spec string, selfValue Value, args []Value, node ast.Node) Value {
	switch spec {
	case "__integer_tostring":
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "Integer.ToString does not take arguments")
		}
		intVal, ok := selfValue.(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Integer.ToString requires integer receiver")
		}
		return &StringValue{Value: strconv.FormatInt(intVal.Value, 10)}

	case "__float_tostring_prec":
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "Float.ToString expects exactly 1 argument")
		}
		floatVal, ok := selfValue.(*FloatValue)
		if !ok {
			return i.newErrorWithLocation(node, "Float.ToString requires float receiver")
		}
		precVal, ok := args[0].(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Float.ToString precision must be Integer, got %s", args[0].Type())
		}
		precision := int(precVal.Value)
		if precision < 0 {
			precision = 0
		}
		return &StringValue{Value: fmt.Sprintf("%.*f", precision, floatVal.Value)}

	case "__boolean_tostring":
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "Boolean.ToString does not take arguments")
		}
		boolVal, ok := selfValue.(*BooleanValue)
		if !ok {
			return i.newErrorWithLocation(node, "Boolean.ToString requires boolean receiver")
		}
		if boolVal.Value {
			return &StringValue{Value: "True"}
		}
		return &StringValue{Value: "False"}

	case "__integer_tohexstring":
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "Integer.ToHexString expects exactly 1 argument")
		}
		intVal, ok := selfValue.(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Integer.ToHexString requires integer receiver")
		}
		digitsVal, ok := args[0].(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Integer.ToHexString digits must be Integer, got %s", args[0].Type())
		}
		digits := int(digitsVal.Value)
		if digits < 0 {
			digits = 0
		}
		// Format as uppercase hex with specified width
		hexStr := fmt.Sprintf("%X", intVal.Value)
		// Pad with zeros if needed
		if len(hexStr) < digits {
			hexStr = strings.Repeat("0", digits-len(hexStr)) + hexStr
		}
		return &StringValue{Value: hexStr}

	case "__string_toupper":
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "String.ToUpper does not take arguments")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.ToUpper requires string receiver")
		}
		return &StringValue{Value: strings.ToUpper(strVal.Value)}

	case "__string_tolower":
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "String.ToLower does not take arguments")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.ToLower requires string receiver")
		}
		return &StringValue{Value: strings.ToLower(strVal.Value)}

	case "__string_array_join":
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "String array Join expects exactly 1 argument")
		}
		separator, ok := args[0].(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "Join separator must be String, got %s", args[0].Type())
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Join helper requires string array receiver")
		}
		var builder strings.Builder
		for idx, elem := range arrVal.Elements {
			strElem, ok := elem.(*StringValue)
			if !ok {
				return i.newErrorWithLocation(node, "Join requires elements of type String")
			}
			if idx > 0 {
				builder.WriteString(separator.Value)
			}
			builder.WriteString(strElem.Value)
		}
		return &StringValue{Value: builder.String()}

	case "__array_add":
		// Implements arr.Add(value) - adds an element to a dynamic array
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "Array.Add expects exactly 1 argument")
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Add requires array receiver")
		}

		// Check if it's a dynamic array (static arrays cannot use Add)
		if !arrVal.ArrayType.IsDynamic() {
			return i.newErrorWithLocation(node, "Add() can only be used with dynamic arrays, not static arrays")
		}

		// For dynamic arrays, just append the element
		// Type checking should have been done at semantic analysis
		valueToAdd := args[0]
		arrVal.Elements = append(arrVal.Elements, valueToAdd)

		// Return nil (procedure, not a function)
		return &NilValue{}

	case "__array_setlength":
		// Implements arr.SetLength(newLength) - resizes a dynamic array
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "Array.SetLength expects exactly 1 argument")
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.SetLength requires array receiver")
		}

		// Check if it's a dynamic array (static arrays cannot use SetLength)
		if !arrVal.ArrayType.IsDynamic() {
			return i.newErrorWithLocation(node, "SetLength() can only be used with dynamic arrays, not static arrays")
		}

		// Get the new length
		lengthInt, ok := args[0].(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.SetLength expects integer argument, got %s", args[0].Type())
		}

		newLength := int(lengthInt.Value)
		if newLength < 0 {
			return i.newErrorWithLocation(node, "Array.SetLength expects non-negative length, got %d", newLength)
		}

		currentLength := len(arrVal.Elements)

		if newLength == currentLength {
			// No change
			return &NilValue{}
		}

		if newLength < currentLength {
			// Truncate the slice
			arrVal.Elements = arrVal.Elements[:newLength]
			return &NilValue{}
		}

		// Extend the slice with default values
		elementType := arrVal.ArrayType.ElementType
		for j := currentLength; j < newLength; j++ {
			arrVal.Elements = append(arrVal.Elements, getZeroValueForType(elementType, nil))
		}

		// Return nil (procedure, not a function)
		return &NilValue{}

	case "__array_delete":
		// Implements arr.Delete(index) or arr.Delete(index, count) - removes elements from a dynamic array
		if len(args) < 1 || len(args) > 2 {
			return i.newErrorWithLocation(node, "Array.Delete expects 1 or 2 arguments, got %d", len(args))
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Delete requires array receiver")
		}

		// Check if it's a dynamic array (static arrays cannot use Delete)
		if !arrVal.ArrayType.IsDynamic() {
			return i.newErrorWithLocation(node, "Delete() can only be used with dynamic arrays, not static arrays")
		}

		// Get the index
		indexInt, ok := args[0].(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Delete index must be Integer, got %s", args[0].Type())
		}
		index := int(indexInt.Value)

		// Get the count (default to 1 if not specified)
		count := 1
		if len(args) == 2 {
			countInt, ok := args[1].(*IntegerValue)
			if !ok {
				return i.newErrorWithLocation(node, "Array.Delete count must be Integer, got %s", args[1].Type())
			}
			count = int(countInt.Value)
		}

		// Validate index and count
		arrayLen := len(arrVal.Elements)
		if index < 0 || index >= arrayLen {
			return i.newErrorWithLocation(node, "Array.Delete index %d out of bounds (0..%d)", index, arrayLen-1)
		}
		if count < 0 {
			return i.newErrorWithLocation(node, "Array.Delete count must be non-negative, got %d", count)
		}

		// Calculate end index (don't go beyond array length)
		endIndex := index + count
		if endIndex > arrayLen {
			endIndex = arrayLen
		}

		// Delete elements by slicing (removes elements from index to index+count)
		arrVal.Elements = append(arrVal.Elements[:index], arrVal.Elements[endIndex:]...)

		// Return nil (procedure, not a function)
		return &NilValue{}

	case "__array_indexof":
		// Implements arr.IndexOf(value) or arr.IndexOf(value, startIndex) - finds first occurrence
		if len(args) < 1 || len(args) > 2 {
			return i.newErrorWithLocation(node, "Array.IndexOf expects 1 or 2 arguments, got %d", len(args))
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.IndexOf requires array receiver")
		}

		valueToFind := args[0]

		// Get the start index (default to 0 if not specified)
		startIndex := 0
		if len(args) == 2 {
			startIndexInt, ok := args[1].(*IntegerValue)
			if !ok {
				return i.newErrorWithLocation(node, "Array.IndexOf startIndex must be Integer, got %s", args[1].Type())
			}
			startIndex = int(startIndexInt.Value)
		}

		// Use the existing builtinArrayIndexOf function
		return i.builtinArrayIndexOf(arrVal, valueToFind, startIndex)

	default:
		// Try calling as a builtin function with self as first argument
		allArgs := append([]Value{selfValue}, args...)
		result := i.callBuiltin(spec, allArgs)
		if isError(result) {
			// If it's an unknown function error, report it as unknown helper method
			return i.newErrorWithLocation(node, "unknown built-in helper method '%s'", spec)
		}
		return result
	}
}

// ============================================================================
// Built-in Helper Property Implementations
// ============================================================================

// evalBuiltinHelperProperty evaluates a built-in helper property
// Implements .Length, .High, .Low, .Count for arrays
func (i *Interpreter) evalBuiltinHelperProperty(propSpec string, selfValue Value, node ast.Node) Value {
	switch propSpec {
	case "__array_length", "__array_count", "__array_high", "__array_low":
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "built-in property '%s' can only be used on arrays", propSpec)
		}
		var result Value
		switch propSpec {
		case "__array_length", "__array_count":
			// Task 9.34: .Count is an alias for .Length
			result = &IntegerValue{Value: int64(len(arrVal.Elements))}
		case "__array_high":
			if arrVal.ArrayType.IsStatic() {
				result = &IntegerValue{Value: int64(*arrVal.ArrayType.HighBound)}
			} else {
				result = &IntegerValue{Value: int64(len(arrVal.Elements) - 1)}
			}
		case "__array_low":
			if arrVal.ArrayType.IsStatic() {
				result = &IntegerValue{Value: int64(*arrVal.ArrayType.LowBound)}
			} else {
				result = &IntegerValue{Value: 0}
			}
		}
		return result

	case "__integer_tostring":
		intVal, ok := selfValue.(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Integer.ToString property requires integer receiver")
		}
		return &StringValue{Value: strconv.FormatInt(intVal.Value, 10)}

	case "__float_tostring_default":
		floatVal, ok := selfValue.(*FloatValue)
		if !ok {
			return i.newErrorWithLocation(node, "Float.ToString property requires float receiver")
		}
		return &StringValue{Value: fmt.Sprintf("%g", floatVal.Value)}

	case "__boolean_tostring":
		boolVal, ok := selfValue.(*BooleanValue)
		if !ok {
			return i.newErrorWithLocation(node, "Boolean.ToString property requires boolean receiver")
		}
		if boolVal.Value {
			return &StringValue{Value: "True"}
		}
		return &StringValue{Value: "False"}

	case "__enum_value":
		// Implement enum .Value helper property
		enumVal, ok := selfValue.(*EnumValue)
		if !ok {
			return i.newErrorWithLocation(node, "Enum.Value property requires enum receiver")
		}
		return &IntegerValue{Value: int64(enumVal.OrdinalValue)}

	case "__enum_name":
		// Implement enum .Name helper property
		enumVal, ok := selfValue.(*EnumValue)
		if !ok {
			return i.newErrorWithLocation(node, "Enum.Name property requires enum receiver")
		}
		// Return just the value name (e.g., "Red" for TColor.Red)
		// If the enum value doesn't have a name (invalid ordinal), return "?"
		if enumVal.ValueName == "" {
			return &StringValue{Value: "?"}
		}
		return &StringValue{Value: enumVal.ValueName}

	case "__enum_qualifiedname":
		// Implement enum .QualifiedName helper property
		enumVal, ok := selfValue.(*EnumValue)
		if !ok {
			return i.newErrorWithLocation(node, "Enum.QualifiedName property requires enum receiver")
		}
		// Return TypeName.ValueName (e.g., "TColor.Red")
		// If the enum value doesn't have a name (invalid ordinal), return "TypeName.?"
		valueName := enumVal.ValueName
		if valueName == "" {
			valueName = "?"
		}
		return &StringValue{Value: enumVal.TypeName + "." + valueName}

	case "__string_length":
		// Implement String.Length property
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.Length property requires string receiver")
		}
		return &IntegerValue{Value: int64(len(strVal.Value))}

	case "StripAccents":
		// Implement String.StripAccents property (no-argument method accessed as property)
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.StripAccents property requires string receiver")
		}
		return &StringValue{Value: stripAccents(strVal.Value)}

	default:
		return i.newErrorWithLocation(node, "unknown built-in property '%s'", propSpec)
	}
}
