package builtins

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// Default implements the Default() built-in function.
// It returns the default/zero value for a given type.
// Default(TypeName): Value
func Default(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Default() expects exactly 1 argument, got %d", len(args))
	}

	// The argument should be a type identifier
	// In practice, during semantic analysis, the type will be resolved
	// For runtime, we'll receive either a StringValue with the type name
	// or we'll work with the type information directly

	// For now, we'll handle common type names as strings
	typeArg, ok := args[0].(*runtime.StringValue)
	if !ok {
		// If not a string, try to get type information from the value
		return ctx.NewError("Default() expects a type name as argument")
	}

	typeName := typeArg.Value

	// Return default values based on type name
	switch typeName {
	case "Integer", "Int64", "Byte", "Word", "Cardinal", "SmallInt", "ShortInt", "LongWord":
		return &runtime.IntegerValue{Value: 0}
	case "Float", "Double", "Single", "Extended", "Currency":
		return &runtime.FloatValue{Value: 0.0}
	case "String", "UnicodeString", "AnsiString":
		return &runtime.StringValue{Value: ""}
	case "Boolean":
		return &runtime.BooleanValue{Value: false}
	case "Variant":
		// For Variant, return nil (unassigned variant)
		return &runtime.NilValue{}
	default:
		// For unknown types, try to return a default value
		// For class types, return nil
		return &runtime.NilValue{}
	}
}
