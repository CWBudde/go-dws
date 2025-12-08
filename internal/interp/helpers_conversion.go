package interp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/builtins"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
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
		// Arrays should default to an empty array value of the correct element type.
		// If we can resolve the array type, create an empty array; otherwise fall back to nil.
		if arrType, ok := typ.(*types.ArrayType); ok {
			return NewArrayValue(arrType)
		}
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
func (i *Interpreter) resolveTypeFromAnnotation(typeExpr ast.TypeExpression) types.Type {
	if typeExpr == nil {
		return nil
	}

	typeName := typeExpr.String()

	// Normalize type name for case-insensitive comparison
	// DWScript (like Pascal) is case-insensitive for all identifiers including type names
	lowerTypeName := ident.Normalize(typeName)

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

	// Check for class types (supports nested classes via current class context)
	if classInfo := i.resolveClassInfoByName(typeName); classInfo != nil {
		return types.NewClassType(classInfo.Name, nil)
	}

	// Check for interface types
	if interfaceInfo := i.lookupInterfaceInfo(typeName); interfaceInfo != nil {
		return types.NewInterfaceType(interfaceInfo.Name)
	}

	// Check for record types (stored with special prefix in environment)
	recordTypeKey := "__record_type_" + ident.Normalize(typeName)
	if typeVal, ok := i.env.Get(recordTypeKey); ok {
		if recordTypeVal, ok := typeVal.(*RecordTypeValue); ok {
			return recordTypeVal.RecordType
		}
	}

	// Type not found
	return nil
}

// resolveClassInfoByName looks up a class by name, handling both global and nested classes.
func (i *Interpreter) resolveClassInfoByName(name string) *ClassInfo {
	if current := i.currentClassContext(); current != nil {
		if nested := current.lookupNestedClass(name); nested != nil {
			return nested
		}
	}

	if classInfo, ok := i.classes[ident.Normalize(name)]; ok {
		return classInfo
	}

	return nil
}

// currentClassContext inspects the execution environment to find the current class scope.
func (i *Interpreter) currentClassContext() *ClassInfo {
	if val, ok := i.env.Get("__CurrentClass__"); ok {
		if classVal, ok := val.(*ClassInfoValue); ok {
			return classVal.ClassInfo
		}
	}
	if val, ok := i.env.Get("Self"); ok {
		if classVal, ok := val.(*ClassInfoValue); ok {
			return classVal.ClassInfo
		}
		if obj, ok := AsObject(val); ok {
			concreteClass, ok := obj.Class.(*ClassInfo)
			if ok {
				return concreteClass
			}
		}
	}
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
	// Ensure CurrentNode points to the helper invocation for accurate error reporting
	prevNode := i.currentNode
	i.currentNode = node
	defer func() { i.currentNode = prevNode }()

	switch spec {
	case "__integer_tostring":
		intVal, ok := selfValue.(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Integer.ToString requires integer receiver")
		}
		if len(args) > 1 {
			return i.newErrorWithLocation(node, "Integer.ToString expects 0 or 1 argument")
		}

		// Use the IntToStr builtin to share base handling/validation (2..36)
		builtinArgs := []builtins.Value{intVal}
		if len(args) == 1 {
			baseVal, ok := args[0].(*IntegerValue)
			if !ok {
				return i.newErrorWithLocation(node, "Integer.ToString base must be Integer, got %s", args[0].Type())
			}
			builtinArgs = append(builtinArgs, baseVal)
		}
		return builtins.IntToStr(i, builtinArgs)

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

	case "__array_map":
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "Array.Map expects exactly 1 argument")
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Map requires array receiver")
		}
		fn, ok := args[0].(*FunctionPointerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Map expects function pointer as argument, got %s", args[0].Type())
		}
		return i.builtinMap([]Value{arrVal, fn})

	case "__array_join":
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "Array.Join expects exactly 1 argument")
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Join requires array receiver")
		}
		sep, ok := args[0].(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Join separator must be String, got %s", args[0].Type())
		}
		var b strings.Builder
		for idx, elem := range arrVal.Elements {
			if idx > 0 {
				b.WriteString(sep.Value)
			}
			if elem == nil {
				continue
			}
			b.WriteString(elem.String())
		}
		return &StringValue{Value: b.String()}

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

	case "__array_swap":
		if len(args) != 2 {
			return i.newErrorWithLocation(node, "Array.Swap expects exactly 2 arguments, got %d", len(args))
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Swap requires array receiver")
		}

		// Get index i
		iInt, ok := args[0].(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Swap first argument must be Integer, got %s", args[0].Type())
		}
		i_idx := int(iInt.Value)

		// Get index j
		jInt, ok := args[1].(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Swap second argument must be Integer, got %s", args[1].Type())
		}
		j_idx := int(jInt.Value)

		// Validate indices
		arrayLen := len(arrVal.Elements)
		if i_idx < 0 || i_idx >= arrayLen {
			return i.newErrorWithLocation(node, "Array.Swap first index %d out of bounds (0..%d)", i_idx, arrayLen-1)
		}
		if j_idx < 0 || j_idx >= arrayLen {
			return i.newErrorWithLocation(node, "Array.Swap second index %d out of bounds (0..%d)", j_idx, arrayLen-1)
		}

		// Swap elements
		arrVal.Elements[i_idx], arrVal.Elements[j_idx] = arrVal.Elements[j_idx], arrVal.Elements[i_idx]

		// Return nil (procedure, not a function)
		return &NilValue{}

	case "__array_push":
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "Array.Push expects exactly 1 argument")
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Push requires array receiver")
		}

		// Check if it's a dynamic array (static arrays cannot use Push)
		if !arrVal.ArrayType.IsDynamic() {
			return i.newErrorWithLocation(node, "Push() can only be used with dynamic arrays, not static arrays")
		}

		// For dynamic arrays, just append the element
		// Type checking should have been done at semantic analysis
		valueToAdd := args[0]

		// If pushing a record, make a copy to avoid aliasing issues (commit a53517a)
		// Records are value types and should be copied when added to collections
		if recVal, ok := valueToAdd.(*RecordValue); ok {
			valueToAdd = recVal.Copy()
		}

		arrVal.Elements = append(arrVal.Elements, valueToAdd)

		// Return nil (procedure, not a function)
		return &NilValue{}

	case "__array_pop":
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "Array.Pop expects no arguments, got %d", len(args))
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Pop requires array receiver")
		}

		// Check if it's a dynamic array (static arrays cannot use Pop)
		if !arrVal.ArrayType.IsDynamic() {
			return i.newErrorWithLocation(node, "Pop() can only be used with dynamic arrays, not static arrays")
		}

		// Check if array is empty
		if len(arrVal.Elements) == 0 {
			return i.newErrorWithLocation(node, "Pop() called on empty array")
		}

		// Get the last element
		lastElement := arrVal.Elements[len(arrVal.Elements)-1]

		// Remove the last element
		arrVal.Elements = arrVal.Elements[:len(arrVal.Elements)-1]

		// Return the popped element
		return lastElement

	case "__string_tointeger":
		// String.ToInteger() -> StrToInt(self)
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "String.ToInteger does not take arguments")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.ToInteger requires string receiver")
		}
		// Call the builtin StrToInt function
		return builtins.StrToInt(i, []builtins.Value{strVal})

	case "__string_tofloat":
		// String.ToFloat() -> StrToFloat(self)
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "String.ToFloat does not take arguments")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.ToFloat requires string receiver")
		}
		// Call the builtin StrToFloat function
		return builtins.StrToFloat(i, []builtins.Value{strVal})

	case "__string_tostring":
		// String.ToString() -> identity (returns self)
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "String.ToString does not take arguments")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.ToString requires string receiver")
		}
		return strVal

	case "__string_startswith":
		// String.StartsWith(str) -> StrBeginsWith(self, str)
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "String.StartsWith expects exactly 1 argument")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.StartsWith requires string receiver")
		}
		// Call the builtin StrBeginsWith function
		return builtins.StrBeginsWith(i, []builtins.Value{strVal, args[0]})

	case "__string_endswith":
		// String.EndsWith(str) -> StrEndsWith(self, str)
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "String.EndsWith expects exactly 1 argument")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.EndsWith requires string receiver")
		}
		// Call the builtin StrEndsWith function
		return builtins.StrEndsWith(i, []builtins.Value{strVal, args[0]})

	case "__string_contains":
		// String.Contains(str) -> StrContains(self, str)
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "String.Contains expects exactly 1 argument")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.Contains requires string receiver")
		}
		// Call the builtin StrContains function
		return builtins.StrContains(i, []builtins.Value{strVal, args[0]})

	case "__string_indexof":
		// String.IndexOf(substr) -> Pos(substr, self)
		// NOTE: Parameter order is REVERSED! .IndexOf(substr) calls Pos(substr, self)
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "String.IndexOf expects exactly 1 argument")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.IndexOf requires string receiver")
		}
		// Call the builtin Pos function with reversed arguments: Pos(substr, str)
		return builtins.Pos(i, []builtins.Value{args[0], strVal})

	case "__string_matches":
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "String.Matches expects exactly 1 argument")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.Matches requires string receiver")
		}
		return builtins.StrMatches(i, []builtins.Value{strVal, args[0]})

	case "__string_isascii":
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "String.IsASCII does not take arguments")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.IsASCII requires string receiver")
		}
		return builtins.StrIsASCII(i, []builtins.Value{strVal})

	case "__string_copy":
		// String.Copy(start, [len]) -> Copy(self, start, len)
		// Optional second parameter defaults to MaxInt (copy to end)
		if len(args) < 1 || len(args) > 2 {
			return i.newErrorWithLocation(node, "String.Copy expects 1 or 2 arguments, got %d", len(args))
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.Copy requires string receiver")
		}
		// If only one argument, default length to MaxInt
		if len(args) == 1 {
			maxInt := &IntegerValue{Value: 2147483647} // MaxInt
			return builtins.Copy(i, []builtins.Value{strVal, args[0], maxInt})
		}
		// Call the builtin Copy function with 3 arguments: Copy(str, start, len)
		return builtins.Copy(i, []builtins.Value{strVal, args[0], args[1]})

	case "__string_before":
		// String.Before(str) -> StrBefore(self, str)
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "String.Before expects exactly 1 argument")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.Before requires string receiver")
		}
		// Call the builtin StrBefore function
		return builtins.StrBefore(i, []builtins.Value{strVal, args[0]})

	case "__string_after":
		// String.After(str) -> StrAfter(self, str)
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "String.After expects exactly 1 argument")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.After requires string receiver")
		}
		// Call the builtin StrAfter function
		return builtins.StrAfter(i, []builtins.Value{strVal, args[0]})

	case "__string_trim":
		// String.Trim([left,right]) -> Trim variations
		if len(args) != 0 && len(args) != 2 {
			return i.newErrorWithLocation(node, "String.Trim expects 0 or 2 arguments")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.Trim requires string receiver")
		}
		if len(args) == 0 {
			return builtins.Trim(i, []builtins.Value{strVal})
		}
		left, lok := args[0].(*IntegerValue)
		right, rok := args[1].(*IntegerValue)
		if !lok || !rok {
			return i.newErrorWithLocation(node, "String.Trim expects integer counts")
		}
		runes := []rune(strVal.Value)
		start := int(left.Value)
		endTrim := int(right.Value)
		if start < 0 {
			start = 0
		}
		if endTrim < 0 {
			endTrim = 0
		}
		if start+endTrim >= len(runes) {
			return &StringValue{Value: ""}
		}
		return &StringValue{Value: string(runes[start : len(runes)-endTrim])}

	case "__string_trimleft":
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "String.TrimLeft expects exactly 1 argument")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.TrimLeft requires string receiver")
		}
		return builtins.TrimLeft(i, []builtins.Value{strVal, args[0]})

	case "__string_trimright":
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "String.TrimRight expects exactly 1 argument")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.TrimRight requires string receiver")
		}
		return builtins.TrimRight(i, []builtins.Value{strVal, args[0]})

	case "__string_split":
		// String.Split(delimiter) -> StrSplit(self, delimiter)
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "String.Split expects exactly 1 argument")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.Split requires string receiver")
		}
		// Call the builtin StrSplit function
		return builtins.StrSplit(i, []builtins.Value{strVal, args[0]})

	case "__string_tojson":
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "String.ToJSON does not take arguments")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.ToJSON requires string receiver")
		}
		return builtins.StrToJSON(i, []builtins.Value{strVal})

	case "__string_tohtml":
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "String.ToHTML does not take arguments")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.ToHTML requires string receiver")
		}
		return builtins.StrToHtml(i, []builtins.Value{strVal})

	case "__string_tohtmlattribute":
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "String.ToHtmlAttribute does not take arguments")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.ToHtmlAttribute requires string receiver")
		}
		return builtins.StrToHtmlAttribute(i, []builtins.Value{strVal})

	case "__string_tocsstext":
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "String.ToCSSText does not take arguments")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.ToCSSText requires string receiver")
		}
		return builtins.StrToCSSText(i, []builtins.Value{strVal})

	case "__string_toxml":
		// String.ToXML([mode]) -> StrToXML(self, [mode])
		if len(args) > 1 {
			return i.newErrorWithLocation(node, "String.ToXML expects 0 or 1 argument")
		}
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.ToXML requires string receiver")
		}
		// StrToXML accepts an optional mode parameter; forward it when provided.
		allArgs := []builtins.Value{strVal}
		if len(args) == 1 {
			allArgs = append(allArgs, args[0])
		}
		return builtins.StrToXML(i, allArgs)

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

	case "__string_isascii":
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.IsASCII property requires string receiver")
		}
		return builtins.StrIsASCII(i, []builtins.Value{strVal})
	case "__string_trim":
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.Trim property requires string receiver")
		}
		return builtins.Trim(i, []builtins.Value{strVal})
	case "__string_trimleft":
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.TrimLeft property requires string receiver")
		}
		// Property version defaults to trimming whitespace (same as TrimLeft(str))
		return builtins.TrimLeft(i, []builtins.Value{strVal})
	case "__string_trimright":
		strVal, ok := selfValue.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "String.TrimRight property requires string receiver")
		}
		// Property version defaults to trimming whitespace (same as TrimRight(str))
		return builtins.TrimRight(i, []builtins.Value{strVal})

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
		// Return the number of Unicode characters (runes), not byte length
		return &IntegerValue{Value: int64(runeLength(strVal.Value))}

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
