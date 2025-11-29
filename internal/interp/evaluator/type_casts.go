package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

// evalTypeCast handles type cast expressions: TypeName(expression)
// Examples: Integer(3.14), String(42), Boolean(1), TMyClass(someObject)
// Supported types: Integer, Float, String, Boolean, Variant, Enum types, Class types
func (e *Evaluator) evalTypeCast(typeName string, argExpr ast.Expression, ctx *ExecutionContext) Value {
	// First check if this is actually a type cast before evaluating the argument
	// This prevents double evaluation when it's not a type cast
	isTypeCast := false
	var enumType *types.EnumType
	lowerName := pkgident.Normalize(typeName)

	// Check if it's a built-in type
	switch lowerName {
	case "integer", "float", "string", "boolean", "variant":
		isTypeCast = true
	default:
		// Check if it's a class/interface type
		if e.typeSystem.HasClass(lowerName) {
			isTypeCast = true
		} else {
			// Check if it's an enum type via environment
			enumTypeKey := "__enum_type_" + lowerName
			if typeVal, ok := ctx.Env().Get(enumTypeKey); ok {
				if etv, ok := typeVal.(EnumTypeValueAccessor); ok {
					enumType = etv.GetEnumType()
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
	val := e.Eval(argExpr, ctx)
	if isError(val) {
		return val
	}

	// Perform the type cast
	switch lowerName {
	case "integer":
		return e.castToInteger(val)
	case "float":
		return e.castToFloat(val)
	case "string":
		return e.castToString(val)
	case "boolean":
		return e.castToBoolean(val)
	case "variant":
		// Variant can accept any value - wrap directly
		return e.adapter.BoxVariant(val)
	default:
		// Check if it's an enum type
		if enumType != nil {
			return e.castToEnum(val, enumType, typeName)
		}
		// Must be a class/interface type (we already checked above)
		// Delegate to adapter for complex class casting logic
		return e.adapter.CastToClass(val, typeName, argExpr)
	}
}

// castToInteger converts a value to Integer
func (e *Evaluator) castToInteger(val Value) Value {
	switch v := val.(type) {
	case *runtime.IntegerValue:
		return v
	case *runtime.FloatValue:
		// DWScript Integer() truncates toward zero
		return &runtime.IntegerValue{Value: int64(v.Value)}
	case *runtime.BooleanValue:
		if v.Value {
			return &runtime.IntegerValue{Value: 1}
		}
		return &runtime.IntegerValue{Value: 0}
	case *runtime.StringValue:
		// Try to parse string as integer
		var result int64
		_, err := fmt.Sscanf(v.Value, "%d", &result)
		if err != nil {
			return &runtime.ErrorValue{Message: fmt.Sprintf("cannot convert string '%s' to Integer", v.Value)}
		}
		return &runtime.IntegerValue{Value: result}
	case *runtime.EnumValue:
		// Cast enum to its ordinal value
		return &runtime.IntegerValue{Value: int64(v.OrdinalValue)}
	}

	// Handle Variant by unwrapping (VariantValue is in interp package, not runtime)
	if val.Type() == "VARIANT" {
		if varAccessor, ok := val.(VariantAccessor); ok {
			return e.castToInteger(varAccessor.GetVariantValue())
		}
	}

	return &runtime.ErrorValue{Message: fmt.Sprintf("cannot cast %s to Integer", val.Type())}
}

// castToFloat converts a value to Float
func (e *Evaluator) castToFloat(val Value) Value {
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
		// Try to parse string as float
		var result float64
		_, err := fmt.Sscanf(v.Value, "%f", &result)
		if err != nil {
			return &runtime.ErrorValue{Message: fmt.Sprintf("cannot convert string '%s' to Float", v.Value)}
		}
		return &runtime.FloatValue{Value: result}
	case *runtime.EnumValue:
		// Cast enum to its ordinal value as float
		return &runtime.FloatValue{Value: float64(v.OrdinalValue)}
	}

	// Handle Variant by unwrapping (VariantValue is in interp package, not runtime)
	if val.Type() == "VARIANT" {
		if varAccessor, ok := val.(VariantAccessor); ok {
			return e.castToFloat(varAccessor.GetVariantValue())
		}
	}

	return &runtime.ErrorValue{Message: fmt.Sprintf("cannot cast %s to Float", val.Type())}
}

// castToString converts a value to String
func (e *Evaluator) castToString(val Value) Value {
	switch v := val.(type) {
	case *runtime.StringValue:
		return v
	case *runtime.IntegerValue:
		return &runtime.StringValue{Value: fmt.Sprintf("%d", v.Value)}
	case *runtime.FloatValue:
		return &runtime.StringValue{Value: fmt.Sprintf("%g", v.Value)}
	case *runtime.BooleanValue:
		if v.Value {
			return &runtime.StringValue{Value: "True"}
		}
		return &runtime.StringValue{Value: "False"}
	}

	// Handle Variant by unwrapping (VariantValue is in interp package, not runtime)
	if val.Type() == "VARIANT" {
		if varAccessor, ok := val.(VariantAccessor); ok {
			return e.castToString(varAccessor.GetVariantValue())
		}
	}

	// For other types, use their String() representation
	return &runtime.StringValue{Value: val.String()}
}

// castToBoolean converts a value to Boolean
func (e *Evaluator) castToBoolean(val Value) Value {
	switch v := val.(type) {
	case *runtime.BooleanValue:
		return v
	case *runtime.IntegerValue:
		return &runtime.BooleanValue{Value: v.Value != 0}
	case *runtime.FloatValue:
		return &runtime.BooleanValue{Value: v.Value != 0.0}
	case *runtime.StringValue:
		// Parse string to boolean (DWScript semantics)
		// Recognized as true: "1", "T", "t", "Y", "y", "yes", "true" (case-insensitive)
		// Everything else is false
		s := strings.TrimSpace(v.Value)
		if s == "" {
			return &runtime.BooleanValue{Value: false}
		}
		// Check single character shortcuts
		if len(s) == 1 {
			switch s[0] {
			case '1', 'T', 't', 'Y', 'y':
				return &runtime.BooleanValue{Value: true}
			}
			return &runtime.BooleanValue{Value: false}
		}
		// Check multi-character strings (case-insensitive)
		if pkgident.Equal(s, "yes") || pkgident.Equal(s, "true") {
			return &runtime.BooleanValue{Value: true}
		}
		return &runtime.BooleanValue{Value: false}
	}

	// Handle Variant by unwrapping (VariantValue is in interp package, not runtime)
	if val.Type() == "VARIANT" {
		if varAccessor, ok := val.(VariantAccessor); ok {
			return e.castToBoolean(varAccessor.GetVariantValue())
		}
	}

	return &runtime.ErrorValue{Message: fmt.Sprintf("cannot cast %s to Boolean", val.Type())}
}

// castToEnum casts a value to an enum type.
// Supports Integer → Enum and Enum → Enum (same type) casting.
func (e *Evaluator) castToEnum(val Value, targetEnum *types.EnumType, typeName string) Value {
	switch v := val.(type) {
	case *runtime.IntegerValue:
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

		return &runtime.EnumValue{
			TypeName:     typeName,
			ValueName:    valueName,
			OrdinalValue: ordinal,
		}

	case *runtime.EnumValue:
		// Enum → Enum: Only allow identity cast (same type)
		if pkgident.Equal(v.TypeName, typeName) {
			return v
		}
		return &runtime.ErrorValue{Message: fmt.Sprintf("cannot cast enum %s to %s: incompatible enum types", v.TypeName, typeName)}
	}

	// Handle Variant by unwrapping (VariantValue is in interp package, not runtime)
	if val.Type() == "VARIANT" {
		if varAccessor, ok := val.(VariantAccessor); ok {
			return e.castToEnum(varAccessor.GetVariantValue(), targetEnum, typeName)
		}
	}

	return &runtime.ErrorValue{Message: fmt.Sprintf("cannot cast %s to enum %s", val.Type(), typeName)}
}

// builtinDefault handles the Default() built-in function which expects an unevaluated type identifier.
// Default(Integer) should pass "Integer" as a string, not evaluate it as a variable.
// Returns the default/zero value for the specified type, or nil if not a valid type.
// Task 3.5.94: Migrated from Interpreter.evalDefaultFunction.
func (e *Evaluator) builtinDefault(args []ast.Expression, ctx *ExecutionContext) Value {
	// Check argument count
	if len(args) != 1 {
		return &runtime.ErrorValue{Message: "Default() expects exactly one argument"}
	}

	// The argument should be a type identifier (not evaluated)
	ident, ok := args[0].(*ast.Identifier)
	if !ok {
		return &runtime.ErrorValue{Message: "Default() expects a type name as argument"}
	}

	typeName := ident.Value
	lowerName := pkgident.Normalize(typeName)

	// Return default values based on type name
	switch lowerName {
	case "integer", "int64", "byte", "word", "cardinal", "smallint", "shortint", "longword":
		return &runtime.IntegerValue{Value: 0}
	case "float", "double", "single", "extended", "currency":
		return &runtime.FloatValue{Value: 0.0}
	case "string", "unicodestring", "ansistring":
		return &runtime.StringValue{Value: ""}
	case "boolean":
		return &runtime.BooleanValue{Value: false}
	case "variant":
		return &runtime.NilValue{}
	default:
		// For class types, records, enums, and other reference/complex types, return nil
		// Check if it's a valid type by looking it up
		// For now, return nil (which represents the default value for reference types)
		return &runtime.NilValue{}
	}
}

// EnumTypeValueAccessor provides access to EnumType from EnumTypeValue
type EnumTypeValueAccessor interface {
	GetEnumType() *types.EnumType
}

// VariantAccessor provides access to variant values
type VariantAccessor interface {
	GetVariantValue() Value
}
