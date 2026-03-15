package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Enum Declaration Evaluation
// ============================================================================

// evalEnumDeclaration evaluates an enum type declaration.
// It registers the enum type and its values in the interpreter's symbol table.
func (i *Interpreter) evalEnumDeclaration(decl *ast.EnumDecl) Value {
	if decl == nil {
		return &ErrorValue{Message: "nil enum declaration"}
	}

	enumName := decl.Name.Value

	// Build the enum type from the declaration
	enumValues := make(map[string]int)
	orderedNames := make([]string, 0, len(decl.Values))

	// Calculate ordinal values (explicit or implicit)
	currentOrdinal := 0
	flagBitPosition := 0 // For flags enums, track the bit position (2^n)

	for _, enumValue := range decl.Values {
		valueName := enumValue.Name

		// Determine ordinal value (explicit or implicit)
		var ordinalValue int

		// Check ValueExpr first (constant expressions)
		if enumValue.ValueExpr != nil {
			// Evaluate constant expression and coerce to an ordinal
			val := i.Eval(enumValue.ValueExpr)
			if isError(val) {
				return &ErrorValue{
					Message: fmt.Sprintf("enum '%s' value '%s': %v", enumName, valueName, val),
				}
			}

			var errVal *ErrorValue
			ordinalValue, errVal = i.extractEnumOrdinal(val, enumName, valueName)
			if errVal != nil {
				return errVal
			}

			if decl.Flags {
				// For flags, explicit values must be powers of 2
				if ordinalValue <= 0 || (ordinalValue&(ordinalValue-1)) != 0 {
					return &ErrorValue{
						Message: fmt.Sprintf("enum '%s' value '%s' (%d) must be a power of 2 for flags enum",
							enumName, valueName, ordinalValue),
					}
				}
				// For flags, update bit position based on explicit value
				for bitPos := 0; bitPos < 64; bitPos++ {
					if (1 << bitPos) == ordinalValue {
						flagBitPosition = bitPos + 1
						break
					}
				}
			} else {
				// For regular enums, update current ordinal
				currentOrdinal = ordinalValue + 1
			}
		} else if enumValue.Value != nil {
			// Backward compatibility: Explicit value provided (simple integer)
			ordinalValue = *enumValue.Value
			if decl.Flags {
				// For flags, explicit values must be powers of 2
				if ordinalValue <= 0 || (ordinalValue&(ordinalValue-1)) != 0 {
					return &ErrorValue{
						Message: fmt.Sprintf("enum '%s' value '%s' (%d) must be a power of 2 for flags enum",
							enumName, valueName, ordinalValue),
					}
				}
				// For flags, update bit position based on explicit value
				for bitPos := 0; bitPos < 64; bitPos++ {
					if (1 << bitPos) == ordinalValue {
						flagBitPosition = bitPos + 1
						break
					}
				}
			} else {
				// For regular enums, update current ordinal
				currentOrdinal = ordinalValue + 1
			}
		} else {
			// Implicit value
			if decl.Flags {
				// Flags use power-of-2 values: 1, 2, 4, 8, 16, ...
				ordinalValue = 1 << flagBitPosition
				flagBitPosition++
			} else {
				// Regular enums use sequential values
				ordinalValue = currentOrdinal
				currentOrdinal++
			}
		}

		// Store the value
		enumValues[valueName] = ordinalValue
		orderedNames = append(orderedNames, valueName)
	}

	// Create the enum type
	var enumType *types.EnumType
	if decl.Scoped || decl.Flags {
		enumType = types.NewScopedEnumType(enumName, enumValues, orderedNames, decl.Flags)
	} else {
		enumType = types.NewEnumType(enumName, enumValues, orderedNames)
	}

	// Register each enum value in the symbol table as a constant
	// For scoped enums (enum/flags keyword), skip global registration -
	// values are only accessible via qualified access (Type.Value)
	if !decl.Scoped {
		for valueName, ordinalValue := range enumValues {
			enumVal := &EnumValue{
				TypeName:     enumName,
				ValueName:    valueName,
				OrdinalValue: ordinalValue,
			}
			i.Env().Define(valueName, enumVal)
		}
	}

	// Store enum type metadata in environment with special key
	// This allows variable declarations to resolve the type
	enumTypeKey := "__enum_type_" + ident.Normalize(enumName)
	i.Env().Define(enumTypeKey, &EnumTypeValue{EnumType: enumType})

	// This enables dual storage during migration (both environment and TypeSystem)
	i.typeSystem.RegisterEnumType(enumName, &EnumTypeValue{EnumType: enumType})

	// Register enum type name as a TypeMetaValue
	// This allows the type name to be used as a runtime value in expressions
	// like High(TColor) or Low(TColor), just like built-in types (Integer, Float, etc.)
	i.Env().Define(enumName, NewTypeMetaValue(enumType, enumName))

	return &NilValue{}
}

// EnumTypeValue is an internal value type used to store enum type metadata
// in the interpreter's environment.
type EnumTypeValue struct {
	EnumType *types.EnumType
}

// Type returns "ENUM_TYPE".
func (e *EnumTypeValue) Type() string {
	return "ENUM_TYPE"
}

// String returns the enum type name.
func (e *EnumTypeValue) String() string {
	return e.EnumType.Name
}

// GetEnumType returns the underlying EnumType.
func (e *EnumTypeValue) GetEnumType() *types.EnumType {
	return e.EnumType
}

// ============================================================================
// Enum Literal Evaluation
// ============================================================================

// extractEnumOrdinal coerces a value produced by an enum ValueExpr into an ordinal integer.
// Supports integers, enums, booleans, single-character strings, subrange values, and
// registered implicit conversions to Integer.
func (i *Interpreter) extractEnumOrdinal(val Value, enumName, valueName string) (int, *ErrorValue) {
	// Unwrap variant containers first
	val = unwrapVariant(val)

	switch v := val.(type) {
	case *IntegerValue:
		return int(v.Value), nil
	case *EnumValue:
		return v.OrdinalValue, nil
	case *BooleanValue:
		if v.Value {
			return 1, nil
		}
		return 0, nil
	case *StringValue:
		runes := []rune(v.Value)
		if len(runes) == 0 {
			return 0, &ErrorValue{
				Message: fmt.Sprintf("enum '%s' value '%s': empty string not valid for enum value", enumName, valueName),
			}
		}
		if len(runes) > 1 {
			return 0, &ErrorValue{
				Message: fmt.Sprintf("enum '%s' value '%s': string '%s' must be a single character", enumName, valueName, v.Value),
			}
		}
		return int(runes[0]), nil
	case *SubrangeValue:
		return v.Value, nil
	}

	// Try implicit conversions registered in the type system (e.g., Enum → Integer)
	if converted, ok := i.tryImplicitConversion(val, "Integer"); ok {
		if errVal, isErr := converted.(*ErrorValue); isErr {
			return 0, errVal
		}
		if intVal, ok := converted.(*IntegerValue); ok {
			return int(intVal.Value), nil
		}
	}

	return 0, &ErrorValue{
		Message: fmt.Sprintf("enum '%s' value '%s': expected ordinal constant, got %s", enumName, valueName, val.Type()),
	}
}
