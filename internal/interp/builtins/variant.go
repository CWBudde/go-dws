package builtins

import (
	"strconv"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// Variant Built-in Functions
// ============================================================================
//
// This file contains Variant introspection and conversion functions that have
// been migrated from internal/interp to use the Context interface pattern.
//
// Functions in this file:
//   - VarType: Returns the type code of a Variant
//   - VarIsNull, VarIsEmpty, VarIsClear: Check if Variant is unassigned
//   - VarIsArray, VarIsStr, VarIsNumeric: Type checking functions
//   - VarToStr, VarToInt, VarToFloat: Conversion functions
//   - VarAsType: Convert Variant to specified type code
//   - VarClear: Clear Variant to unassigned state

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

// VarType returns the type code of a Variant value.
// VarType(v: Variant): Integer
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
func VarType(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("VarType() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Unwrap if it's a Variant
	val := ctx.UnwrapVariant(arg)

	// Handle nil/empty value
	if val == nil {
		return &runtime.IntegerValue{Value: varEmpty}
	}

	// Return type code based on actual type
	return varTypeFromValue(ctx, val)
}

// varTypeFromValue returns the VarType code for a runtime Value.
func varTypeFromValue(ctx Context, val Value) Value {
	if val == nil {
		return &runtime.IntegerValue{Value: varEmpty}
	}

	// Check if it's a JSON value - must come before Type() switch
	// because JSON values need special handling based on their kind
	if typeCode, ok := ctx.GetJSONVarType(val); ok {
		return &runtime.IntegerValue{Value: typeCode}
	}

	switch val.Type() {
	case "INTEGER":
		return &runtime.IntegerValue{Value: varInteger}
	case "FLOAT":
		return &runtime.IntegerValue{Value: varDouble}
	case "STRING":
		return &runtime.IntegerValue{Value: varString}
	case "BOOLEAN":
		return &runtime.IntegerValue{Value: varBoolean}
	case "NIL", "NULL", "UNASSIGNED":
		return &runtime.IntegerValue{Value: varEmpty}
	case "ARRAY":
		return &runtime.IntegerValue{Value: varArray}
	case "VARIANT":
		return &runtime.IntegerValue{Value: varVariant}
	default:
		// Unknown type - treat as empty
		return &runtime.IntegerValue{Value: varEmpty}
	}
}

// VarIsNull checks if a Variant is unassigned (has no value).
// VarIsNull(v: Variant): Boolean
//
// Returns True if the Variant is unassigned, False otherwise.
// In DWScript, "null" and "empty" are essentially the same for Variants.
//
// Example:
//
//	var v: Variant;
//	PrintLn(VarIsNull(v));  // Outputs: true
//	v := 42;
//	PrintLn(VarIsNull(v));  // Outputs: false
func VarIsNull(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("VarIsNull() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Unwrap if it's a Variant
	val := ctx.UnwrapVariant(arg)

	// Variant is null if its wrapped value is nil or is a nil-like value
	if val == nil {
		return &runtime.BooleanValue{Value: true}
	}

	// Check for nil-like types
	switch val.Type() {
	case "NIL", "NULL", "UNASSIGNED":
		return &runtime.BooleanValue{Value: true}
	default:
		return &runtime.BooleanValue{Value: false}
	}
}

// VarIsEmpty checks if a Variant is empty (unassigned).
// VarIsEmpty(v: Variant): Boolean
//
// Returns True if the Variant is empty, False otherwise.
// In DWScript, VarIsEmpty is equivalent to VarIsNull.
//
// Example:
//
//	var v: Variant;
//	PrintLn(VarIsEmpty(v));  // Outputs: true
func VarIsEmpty(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("VarIsEmpty() expects exactly 1 argument, got %d", len(args))
	}

	// VarIsEmpty is the same as VarIsNull in DWScript
	return VarIsNull(ctx, args)
}

// VarIsClear checks if a Variant is cleared (unassigned).
// VarIsClear(v: Variant): Boolean
//
// Returns True if the Variant is cleared, False otherwise.
// In DWScript, VarIsClear is an alias for VarIsEmpty.
//
// Example:
//
//	var v: Variant;
//	PrintLn(VarIsClear(v));  // Outputs: true
func VarIsClear(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("VarIsClear() expects exactly 1 argument, got %d", len(args))
	}

	// VarIsClear is the same as VarIsNull in DWScript
	return VarIsNull(ctx, args)
}

// VarIsArray checks if a Variant holds an array value.
// VarIsArray(v: Variant): Boolean
//
// Returns True if the Variant holds an array, False otherwise.
//
// Example:
//
//	var v: Variant := [1, 2, 3];
//	PrintLn(VarIsArray(v));  // Outputs: true
func VarIsArray(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("VarIsArray() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Unwrap if it's a Variant
	val := ctx.UnwrapVariant(arg)

	// Check if the unwrapped value is an array
	if val != nil && val.Type() == "ARRAY" {
		return &runtime.BooleanValue{Value: true}
	}
	return &runtime.BooleanValue{Value: false}
}

// VarIsStr checks if a Variant holds a string value.
// VarIsStr(v: Variant): Boolean
//
// Returns True if the Variant holds a string, False otherwise.
//
// Example:
//
//	var v: Variant := "hello";
//	PrintLn(VarIsStr(v));  // Outputs: true
func VarIsStr(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("VarIsStr() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Unwrap if it's a Variant
	val := ctx.UnwrapVariant(arg)

	// Check if the unwrapped value is a string
	if val != nil && val.Type() == "STRING" {
		return &runtime.BooleanValue{Value: true}
	}
	return &runtime.BooleanValue{Value: false}
}

// VarIsNumeric checks if a Variant holds a numeric value (Integer or Float).
// VarIsNumeric(v: Variant): Boolean
//
// Returns True if the Variant holds an Integer or Float, False otherwise.
//
// Example:
//
//	var v: Variant := 42;
//	PrintLn(VarIsNumeric(v));  // Outputs: true
func VarIsNumeric(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("VarIsNumeric() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Unwrap if it's a Variant
	val := ctx.UnwrapVariant(arg)

	// Check if the unwrapped value is numeric
	if val != nil {
		switch val.Type() {
		case "INTEGER", "FLOAT":
			return &runtime.BooleanValue{Value: true}
		}
	}
	return &runtime.BooleanValue{Value: false}
}

// VarToStr converts a Variant's value to its string representation.
// VarToStr(v: Variant): String
//
// Converts the Variant's value to string:
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
func VarToStr(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("VarToStr() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	val := ctx.UnwrapVariant(arg)

	// Handle nil/empty Variant
	if val == nil {
		return &runtime.StringValue{Value: ""}
	}
	switch val.Type() {
	case "NIL", "NULL", "UNASSIGNED":
		return &runtime.StringValue{Value: ""}
	}

	// Convert based on actual type using String() method
	return &runtime.StringValue{Value: val.String()}
}

// VarToInt converts a Variant's value to Integer.
// VarToInt(v: Variant): Integer
//
// Converts the Variant's value to Integer:
//   - Integer: returns as-is
//   - Float: truncates to integer (e.g., 3.9 → 3)
//   - String: parses as integer (e.g., "42" → 42)
//   - Boolean: True → 1, False → 0
//   - Empty/Null: 0
//
// Returns error if conversion is not possible.
//
// Example:
//
//	var v: Variant := 3.14;
//	PrintLn(VarToInt(v));  // Outputs: 3
func VarToInt(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("VarToInt() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	val := ctx.UnwrapVariant(arg)

	// Handle nil/empty Variant
	if val == nil {
		return &runtime.IntegerValue{Value: 0}
	}
	switch val.Type() {
	case "NIL", "NULL", "UNASSIGNED":
		return &runtime.IntegerValue{Value: 0}
	}

	// Convert based on actual type
	switch v := val.(type) {
	case *runtime.IntegerValue:
		return v
	case *runtime.FloatValue:
		return &runtime.IntegerValue{Value: int64(v.Value)}
	case *runtime.BooleanValue:
		if v.Value {
			return &runtime.IntegerValue{Value: 1}
		}
		return &runtime.IntegerValue{Value: 0}
	case *runtime.StringValue:
		// Try to parse as integer
		intVal, err := strconv.ParseInt(v.Value, 10, 64)
		if err != nil {
			return ctx.NewError("cannot convert string '%s' to Integer", v.Value)
		}
		return &runtime.IntegerValue{Value: intVal}
	default:
		return ctx.NewError("cannot convert %s to Integer", val.Type())
	}
}

// VarToFloat converts a Variant's value to Float.
// VarToFloat(v: Variant): Float
//
// Converts the Variant's value to Float:
//   - Float: returns as-is
//   - Integer: converts to float (e.g., 42 → 42.0)
//   - String: parses as float (e.g., "3.14" → 3.14)
//   - Boolean: True → 1.0, False → 0.0
//   - Empty/Null: 0.0
//
// Returns error if conversion is not possible.
//
// Example:
//
//	var v: Variant := 42;
//	PrintLn(VarToFloat(v));  // Outputs: 42.0
func VarToFloat(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("VarToFloat() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	val := ctx.UnwrapVariant(arg)

	// Handle nil/empty Variant
	if val == nil {
		return &runtime.FloatValue{Value: 0.0}
	}
	switch val.Type() {
	case "NIL", "NULL", "UNASSIGNED":
		return &runtime.FloatValue{Value: 0.0}
	}

	// Convert based on actual type
	switch v := val.(type) {
	case *runtime.FloatValue:
		return v
	case *runtime.IntegerValue:
		return &runtime.FloatValue{Value: float64(v.Value)}
	case *runtime.BooleanValue:
		if v.Value {
			return &runtime.FloatValue{Value: 1.0}
		}
		return &runtime.FloatValue{Value: 0.0}
	case *runtime.StringValue:
		// Try to parse as float
		floatVal, err := strconv.ParseFloat(v.Value, 64)
		if err != nil {
			return ctx.NewError("cannot convert string '%s' to Float", v.Value)
		}
		return &runtime.FloatValue{Value: floatVal}
	default:
		return ctx.NewError("cannot convert %s to Float", val.Type())
	}
}

// VarAsType converts a Variant to the specified type code.
// VarAsType(v: Variant, varType: Integer): Variant
//
// Converts the Variant to the specified type code:
//   - varInteger (3): Convert to Integer
//   - varDouble (5): Convert to Float
//   - varString (256): Convert to String
//   - varBoolean (11): Convert to Boolean
//   - varEmpty (0): Return empty Variant
//
// Returns a new Variant with the converted value.
//
// Example:
//
//	var v: Variant := "42";
//	var i: Variant := VarAsType(v, 3);  // Convert to Integer
//	PrintLn(VarToInt(i));  // Outputs: 42
func VarAsType(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("VarAsType() expects exactly 2 arguments, got %d", len(args))
	}

	arg := args[0]
	typeCodeVal := args[1]

	// Extract type code
	typeCodeInt, ok := typeCodeVal.(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("VarAsType() type code must be Integer, got %s", typeCodeVal.Type())
	}
	targetType := typeCodeInt.Value

	val := ctx.UnwrapVariant(arg)

	// Handle nil/empty Variant - convert to zero value of target type
	if val == nil || val.Type() == "NIL" || val.Type() == "NULL" || val.Type() == "UNASSIGNED" {
		switch targetType {
		case varInteger:
			return &runtime.IntegerValue{Value: 0}
		case varDouble:
			return &runtime.FloatValue{Value: 0.0}
		case varString:
			return &runtime.StringValue{Value: ""}
		case varBoolean:
			return &runtime.BooleanValue{Value: false}
		case varEmpty:
			// Note: Cannot create VariantValue here without circular dependency
			// Return nil value instead
			return &runtime.NilValue{}
		default:
			return ctx.NewError("unsupported VarType code: %d", targetType)
		}
	}

	// Convert based on target type
	var converted Value
	switch targetType {
	case varInteger:
		// Use VarToInt for conversion
		converted = VarToInt(ctx, []Value{arg})
		if converted.Type() == "ERROR" {
			return converted
		}
	case varDouble:
		// Use VarToFloat for conversion
		converted = VarToFloat(ctx, []Value{arg})
		if converted.Type() == "ERROR" {
			return converted
		}
	case varString:
		// Use VarToStr for conversion
		converted = VarToStr(ctx, []Value{arg})
		if converted.Type() == "ERROR" {
			return converted
		}
	case varBoolean:
		// Convert to boolean
		switch v := val.(type) {
		case *runtime.BooleanValue:
			converted = v
		case *runtime.IntegerValue:
			converted = &runtime.BooleanValue{Value: v.Value != 0}
		case *runtime.FloatValue:
			converted = &runtime.BooleanValue{Value: v.Value != 0.0}
		case *runtime.StringValue:
			// Non-empty string is true
			converted = &runtime.BooleanValue{Value: v.Value != ""}
		default:
			return ctx.NewError("cannot convert %s to Boolean", val.Type())
		}
	case varEmpty:
		return &runtime.NilValue{}
	default:
		return ctx.NewError("unsupported VarType code: %d", targetType)
	}

	return converted
}

// VarClear clears a Variant and sets it to empty/uninitialized state.
// VarClear(v: Variant): Variant
//
// Note: In standard DWScript, VarClear is a procedure with a var parameter.
// This implementation returns an empty Variant that should be assigned back.
//
// Returns: An empty Variant (VarType = varEmpty)
//
// Example:
//
//	var v: Variant := 42;
//	PrintLn(VarType(v));     // Outputs: 3 (varInteger)
//	v := VarClear(v);
//	PrintLn(VarIsNull(v));   // Outputs: True
func VarClear(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("VarClear() expects exactly 1 argument, got %d", len(args))
	}

	// Return a nil value to represent empty Variant
	// The caller will wrap this in a VariantValue if needed
	return &runtime.NilValue{}
}
