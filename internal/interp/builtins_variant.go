package interp

import "strconv"

// ============================================================================
// Variant Introspection Functions
// ============================================================================

// VarType constants - DWScript/Delphi-compatible type codes
// See: https://docwiki.embarcadero.com/Libraries/en/System.Variants.VarType
const (
	varEmpty    = 0      // Unassigned/Empty
	varNull     = 1      // Null (SQL NULL)
	varSmallint = 2      // 16-bit signed integer
	varInteger  = 3      // 32-bit signed integer
	varSingle   = 4      // Single precision float
	varDouble   = 5      // Double precision float
	varCurrency = 6      // Currency type
	varDate     = 7      // TDateTime
	varOleStr   = 8      // OLE String
	varDispatch = 9      // IDispatch
	varError    = 10     // Error code
	varBoolean  = 11     // Boolean
	varVariant  = 12     // Variant (nested)
	varUnknown  = 13     // IUnknown
	varDecimal  = 14     // Decimal
	varByte     = 17     // Byte (unsigned 8-bit)
	varWord     = 18     // Word (unsigned 16-bit)
	varLongWord = 19     // LongWord (unsigned 32-bit)
	varInt64    = 20     // 64-bit signed integer
	varUInt64   = 21     // 64-bit unsigned integer
	varString   = 256    // String (Unicode)
	varArray    = 0x2000 // Array flag (ORed with element type)
	varJSON     = 0x1000 // JSON object
)

// builtinVarType implements the VarType() built-in function.
//
// Syntax: VarType(v: Variant): Integer
//
// Returns type codes compatible with Delphi's VarType function:
//   - varEmpty (0): Unassigned/uninitialized Variant
//   - varInteger (3): Integer value
//   - varDouble (5): Float value
//   - varBoolean (11): Boolean value
//   - varString (256): String value
//
// Example:
//
//	var v: Variant := 42;
//	PrintLn(VarType(v));  // Outputs: 3 (varInteger)
func (i *Interpreter) builtinVarType(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "VarType() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// If not a Variant, check the direct type
	// This allows VarType to work with non-Variant values for convenience
	variant, ok := arg.(*VariantValue)
	if !ok {
		// Non-Variant argument - check its actual type
		return i.varTypeFromValue(arg)
	}

	// Variant with nil value is unassigned/empty
	if variant.Value == nil {
		return &IntegerValue{Value: varEmpty}
	}

	// Return type code for the wrapped value
	return i.varTypeFromValue(variant.Value)
}

// varTypeFromValue returns the VarType code for a runtime Value.
func (i *Interpreter) varTypeFromValue(val Value) Value {
	if val == nil {
		return &IntegerValue{Value: varEmpty}
	}

	switch v := val.(type) {
	case *IntegerValue:
		return &IntegerValue{Value: varInteger}
	case *FloatValue:
		return &IntegerValue{Value: varDouble}
	case *StringValue:
		return &IntegerValue{Value: varString}
	case *BooleanValue:
		return &IntegerValue{Value: varBoolean}
	case *NilValue:
		return &IntegerValue{Value: varEmpty}
	case *ArrayValue:
		// Arrays have varArray flag ORed with element type
		// For simplicity, return varArray for now
		return &IntegerValue{Value: varArray}
	case *VariantValue:
		// Nested variant
		return &IntegerValue{Value: varVariant}
	case *JSONValue:
		// Return VarType code based on JSON kind
		typeCode := jsonKindToVarType(v.Value.Kind())
		return &IntegerValue{Value: typeCode}
	default:
		// Unknown type - treat as empty
		return &IntegerValue{Value: varEmpty}
	}
}

// builtinVarIsNull implements the VarIsNull() built-in function.
//
// Syntax: VarIsNull(v: Variant): Boolean
//
// Returns True if the Variant is unassigned (has no value), False otherwise.
// In DWScript, "null" and "empty" are essentially the same for Variants
// (unlike VBScript which distinguishes them).
//
// Example:
//
//	var v: Variant;
//	PrintLn(VarIsNull(v));  // Outputs: true
//	v := 42;
//	PrintLn(VarIsNull(v));  // Outputs: false
func (i *Interpreter) builtinVarIsNull(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "VarIsNull() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Check if it's a Variant
	variant, ok := arg.(*VariantValue)
	if !ok {
		// Non-Variant values are not null
		return &BooleanValue{Value: false}
	}

	// Variant is null if its wrapped value is nil
	return &BooleanValue{Value: variant.Value == nil}
}

// builtinVarIsEmpty implements the VarIsEmpty() built-in function.
//
// Syntax: VarIsEmpty(v: Variant): Boolean
//
// Returns True if the Variant is empty (unassigned), False otherwise.
// In DWScript, VarIsEmpty is equivalent to VarIsNull - both check for unassigned Variants.
// This differs from VBScript where Empty and Null are distinct states.
//
// Example:
//
//	var v: Variant;
//	PrintLn(VarIsEmpty(v));  // Outputs: true
//	v := "hello";
//	PrintLn(VarIsEmpty(v));  // Outputs: false
func (i *Interpreter) builtinVarIsEmpty(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "VarIsEmpty() expects exactly 1 argument, got %d", len(args))
	}

	// VarIsEmpty is the same as VarIsNull in DWScript
	return i.builtinVarIsNull(args)
}

// builtinVarIsNumeric implements the VarIsNumeric() built-in function.
//
// Syntax: VarIsNumeric(v: Variant): Boolean
//
// Returns True if the Variant holds an Integer or Float value, False otherwise.
// Unassigned Variants return False.
//
// Example:
//
//	var v: Variant := 42;
//	PrintLn(VarIsNumeric(v));  // Outputs: true
//	v := "hello";
//	PrintLn(VarIsNumeric(v));  // Outputs: false
func (i *Interpreter) builtinVarIsNumeric(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "VarIsNumeric() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Unwrap if it's a Variant
	val := unwrapVariant(arg)

	// Check if the unwrapped value is numeric
	switch val.(type) {
	case *IntegerValue, *FloatValue:
		return &BooleanValue{Value: true}
	default:
		return &BooleanValue{Value: false}
	}
}

// ============================================================================
// Variant Conversion Functions
// ============================================================================

// builtinVarToStr implements the VarToStr() built-in function.
//
// Syntax: VarToStr(v: Variant): String
//
// Converts the Variant's value to its string representation:
//   - Integer: decimal string (e.g., 42 → "42")
//   - Float: decimal string (e.g., 3.14 → "3.14")
//   - String: returns as-is
//   - Boolean: "True" or "False"
//   - Empty/Null: empty string ""
//
// Example:
//
//	var v: Variant := 42;
//	PrintLn(VarToStr(v));  // Outputs: "42"
func (i *Interpreter) builtinVarToStr(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "VarToStr() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	val := unwrapVariant(arg)

	// Handle nil/empty Variant
	if val == nil {
		return &StringValue{Value: ""}
	}
	if _, ok := val.(*NilValue); ok {
		return &StringValue{Value: ""}
	}

	// Convert based on actual type
	return &StringValue{Value: i.convertToString(val)}
}

// builtinVarToInt implements the VarToInt() built-in function.
//
// Syntax: VarToInt(v: Variant): Integer
//
// Converts the Variant's value to Integer:
//   - Integer: returns as-is
//   - Float: truncates to integer (e.g., 3.9 → 3)
//   - String: parses as integer (e.g., "42" → 42)
//   - Boolean: True → 1, False → 0
//   - Empty/Null: 0
//
// Returns error if conversion is not possible (e.g., "abc" cannot convert to Integer).
//
// Example:
//
//	var v: Variant := 3.14;
//	PrintLn(VarToInt(v));  // Outputs: 3
func (i *Interpreter) builtinVarToInt(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "VarToInt() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	val := unwrapVariant(arg)

	// Handle nil/empty Variant
	if val == nil {
		return &IntegerValue{Value: 0}
	}
	if _, ok := val.(*NilValue); ok {
		return &IntegerValue{Value: 0}
	}

	// Convert based on actual type
	switch v := val.(type) {
	case *IntegerValue:
		return v
	case *FloatValue:
		return &IntegerValue{Value: int64(v.Value)}
	case *BooleanValue:
		if v.Value {
			return &IntegerValue{Value: 1}
		}
		return &IntegerValue{Value: 0}
	case *StringValue:
		// Try to parse as integer
		intVal, err := strconv.ParseInt(v.Value, 10, 64)
		if err != nil {
			return i.newErrorWithLocation(i.currentNode, "cannot convert string '%s' to Integer", v.Value)
		}
		return &IntegerValue{Value: intVal}
	default:
		return i.newErrorWithLocation(i.currentNode, "cannot convert %s to Integer", val.Type())
	}
}

// builtinVarToFloat implements the VarToFloat() built-in function.
//
// Syntax: VarToFloat(v: Variant): Float
//
// Converts the Variant's value to Float:
//   - Float: returns as-is
//   - Integer: converts to float (e.g., 42 → 42.0)
//   - String: parses as float (e.g., "3.14" → 3.14)
//   - Boolean: True → 1.0, False → 0.0
//   - Empty/Null: 0.0
//
// Returns error if conversion is not possible (e.g., "abc" cannot convert to Float).
//
// Example:
//
//	var v: Variant := 42;
//	PrintLn(VarToFloat(v));  // Outputs: 42.0
func (i *Interpreter) builtinVarToFloat(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "VarToFloat() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	val := unwrapVariant(arg)

	// Handle nil/empty Variant
	if val == nil {
		return &FloatValue{Value: 0.0}
	}
	if _, ok := val.(*NilValue); ok {
		return &FloatValue{Value: 0.0}
	}

	// Convert based on actual type
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
		// Try to parse as float
		floatVal, err := strconv.ParseFloat(v.Value, 64)
		if err != nil {
			return i.newErrorWithLocation(i.currentNode, "cannot convert string '%s' to Float", v.Value)
		}
		return &FloatValue{Value: floatVal}
	default:
		return i.newErrorWithLocation(i.currentNode, "cannot convert %s to Float", val.Type())
	}
}

// builtinVarAsType implements the VarAsType() built-in function.
//
// Syntax: VarAsType(v: Variant, varType: Integer): Variant
//
// Converts the Variant to the specified type code:
//   - varInteger (3): Convert to Integer
//   - varDouble (5): Convert to Float
//   - varString (256): Convert to String
//   - varBoolean (11): Convert to Boolean
//
// Returns a new Variant with the converted value.
//
// Example:
//
//	var v: Variant := "42";
//	var i: Variant := VarAsType(v, 3);  // Convert to Integer
//	PrintLn(VarToInt(i));  // Outputs: 42
func (i *Interpreter) builtinVarAsType(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "VarAsType() expects exactly 2 arguments, got %d", len(args))
	}

	arg := args[0]
	typeCodeVal := args[1]

	// Extract type code
	typeCodeInt, ok := typeCodeVal.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "VarAsType() type code must be Integer, got %s", typeCodeVal.Type())
	}
	targetType := typeCodeInt.Value

	val := unwrapVariant(arg)

	// Handle nil/empty Variant - convert to zero value of target type
	if val == nil || val.Type() == "NIL" {
		switch targetType {
		case varInteger:
			return boxVariant(&IntegerValue{Value: 0})
		case varDouble:
			return boxVariant(&FloatValue{Value: 0.0})
		case varString:
			return boxVariant(&StringValue{Value: ""})
		case varBoolean:
			return boxVariant(&BooleanValue{Value: false})
		case varEmpty:
			return &VariantValue{Value: nil, ActualType: nil}
		default:
			return i.newErrorWithLocation(i.currentNode, "unsupported VarType code: %d", targetType)
		}
	}

	// Convert based on target type
	var converted Value
	switch targetType {
	case varInteger:
		// Use VarToInt for conversion
		converted = i.builtinVarToInt([]Value{arg})
		if converted.Type() == "ERROR" {
			return converted
		}
	case varDouble:
		// Use VarToFloat for conversion
		converted = i.builtinVarToFloat([]Value{arg})
		if converted.Type() == "ERROR" {
			return converted
		}
	case varString:
		// Use VarToStr for conversion
		converted = i.builtinVarToStr([]Value{arg})
		if converted.Type() == "ERROR" {
			return converted
		}
	case varBoolean:
		// Convert to boolean
		switch v := val.(type) {
		case *BooleanValue:
			converted = v
		case *IntegerValue:
			converted = &BooleanValue{Value: v.Value != 0}
		case *FloatValue:
			converted = &BooleanValue{Value: v.Value != 0.0}
		case *StringValue:
			// Non-empty string is true
			converted = &BooleanValue{Value: v.Value != ""}
		default:
			return i.newErrorWithLocation(i.currentNode, "cannot convert %s to Boolean", val.Type())
		}
	case varEmpty:
		return &VariantValue{Value: nil, ActualType: nil}
	default:
		return i.newErrorWithLocation(i.currentNode, "unsupported VarType code: %d", targetType)
	}

	// Box the converted value back into a Variant
	return boxVariant(converted)
}

// builtinVarClear implements the VarClear() built-in function.
// Clears a Variant and sets it to empty/uninitialized state.
//
// Syntax: VarClear(v: Variant): Variant
//
// Note: In standard DWScript, VarClear is a procedure with a var parameter.
// This implementation returns an empty Variant that should be assigned back:
//
//	v := VarClear(v);  // Assigns empty Variant to v
//
// Returns: An empty Variant (VarType = varEmpty)
//
// Example:
//
//	var v: Variant := 42;
//	PrintLn(VarType(v));     // Outputs: 3 (varInteger)
//	v := VarClear(v);
//	PrintLn(VarIsNull(v));   // Outputs: True
func (i *Interpreter) builtinVarClear(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "VarClear() expects exactly 1 argument, got %d", len(args))
	}

	// Return an empty Variant regardless of input
	// This clears the Variant to uninitialized state
	return &VariantValue{Value: nil, ActualType: nil}
}
