package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// This file contains variable and constant declaration evaluation.

// evalProgram evaluates all statements in the program.
func (i *Interpreter) evalProgram(program *ast.Program) Value {
	var result Value

	for _, stmt := range program.Statements {
		result = i.Eval(stmt)

		// If we hit an error, stop execution
		if isError(result) {
			return result
		}

		// Check if exception is active - if so, unwind the stack
		if i.exception != nil {
			break
		}

		// Check if exit was called at program level
		if i.exitSignal {
			i.exitSignal = false // Clear signal
			break                // Exit the program
		}
	}

	// If there's an uncaught exception, convert it to an error
	if i.exception != nil {
		exc := i.exception
		return newError("uncaught exception: %s", exc.Inspect())
	}

	return result
}

// evalVarDeclStatement evaluates a variable declaration statement.
// It defines a new variable in the current environment.
// External variables are marked with a special ExternalVarValue.
func (i *Interpreter) evalVarDeclStatement(stmt *ast.VarDeclStatement) Value {
	var value Value

	// Handle multi-identifier declarations
	// All names share the same type, but each needs to be defined separately
	// Note: Parser already validates that multi-name declarations cannot have initializers

	// Handle external variables
	if stmt.IsExternal {
		// External variables only apply to single declarations
		if len(stmt.Names) != 1 {
			return newError("external keyword cannot be used with multiple variable names")
		}
		// Store a special marker for external variables
		externalName := stmt.ExternalName
		if externalName == "" {
			externalName = stmt.Names[0].Value
		}
		value = &ExternalVarValue{
			Name:         stmt.Names[0].Value,
			ExternalName: externalName,
		}
		i.env.Define(stmt.Names[0].Value, value)
		return value
	}

	if stmt.Value != nil {
		if arrayLit, ok := stmt.Value.(*ast.ArrayLiteralExpression); ok {
			var expected *types.ArrayType
			if stmt.Type != nil {
				arrType, errVal := i.arrayTypeByName(stmt.Type.Name, stmt)
				if errVal != nil {
					return errVal
				}
				expected = arrType
			}
			value = i.evalArrayLiteralWithExpected(arrayLit, expected)
			if isError(value) {
				return value
			}
		} else if recordLit, ok := stmt.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
			// Anonymous record literal needs explicit type
			if stmt.Type == nil {
				return newError("anonymous record literal requires explicit type annotation")
			}
			typeName := stmt.Type.Name
			// Task 9.225: Normalize to lowercase for case-insensitive lookups
			recordTypeKey := "__record_type_" + strings.ToLower(typeName)
			if typeVal, ok := i.env.Get(recordTypeKey); ok {
				if rtv, ok := typeVal.(*RecordTypeValue); ok {
					// Temporarily set the type name for evaluation
					recordLit.TypeName = &ast.Identifier{Value: rtv.RecordType.Name}
					value = i.Eval(recordLit)
					recordLit.TypeName = nil
				} else {
					return newError("type '%s' is not a record type", typeName)
				}
			} else {
				return newError("unknown type '%s'", typeName)
			}
		} else {
			value = i.Eval(stmt.Value)
		}
		if isError(value) {
			return value
		}

		// Task 9.1.3: Check if exception was raised during evaluation
		// This is important for operations that raise exceptions (like invalid casts)
		if i.exception != nil {
			return nil
		}

		// If declaring a subrange variable with an initializer, wrap and validate
		if stmt.Type != nil {
			typeName := stmt.Type.Name
			// Task 9.225: Normalize to lowercase for case-insensitive lookups
			subrangeTypeKey := "__subrange_type_" + strings.ToLower(typeName)
			handledSubrange := false
			if typeVal, ok := i.env.Get(subrangeTypeKey); ok {
				if stv, ok := typeVal.(*SubrangeTypeValue); ok {
					// Extract integer value from initializer
					var intValue int
					if intVal, ok := value.(*IntegerValue); ok {
						intValue = int(intVal.Value)
					} else if srcSubrange, ok := value.(*SubrangeValue); ok {
						intValue = srcSubrange.Value
					} else {
						return newError("cannot initialize subrange type %s with %s", typeName, value.Type())
					}

					// Create subrange value and validate
					subrangeVal := &SubrangeValue{
						Value:        0, // Will be set by ValidateAndSet
						SubrangeType: stv.SubrangeType,
					}
					if err := subrangeVal.ValidateAndSet(intValue); err != nil {
						return &ErrorValue{Message: err.Error()}
					}
					value = subrangeVal
					handledSubrange = true
				}
			}
			if !handledSubrange {
				if converted, ok := i.tryImplicitConversion(value, typeName); ok {
					value = converted
				}
			}

			// Task 9.227: Box value if target type is Variant
			if strings.EqualFold(typeName, "Variant") {
				value = boxVariant(value)
			}
		}
	} else {
		// No initializer - check if we need to initialize based on type
		if stmt.Type != nil {
			typeName := stmt.Type.Name

			// Check for inline array types first
			// Inline array types have signatures like "array of Integer" or "array[1..10] of Integer"
			if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
				// Parse inline array type and create array value
				arrayType := i.parseInlineArrayType(typeName)
				if arrayType != nil {
					value = NewArrayValue(arrayType)
				} else {
					value = &NilValue{}
				}
			} else if strings.HasPrefix(typeName, "set of ") {
				// Task 9.214: Inline set types like "set of TColor"
				setType := i.parseInlineSetType(typeName)
				if setType != nil {
					value = NewSetValue(setType)
				} else {
					value = &NilValue{}
				}
			} else if typeVal, ok := i.env.Get("__record_type_" + strings.ToLower(typeName)); ok {
				// Check if this is a record type
				// Task 9.225: Normalize to lowercase for case-insensitive lookups
				if rtv, ok := typeVal.(*RecordTypeValue); ok {
					// Initialize with empty record value
					// Task 9.7e1: Use createRecordValue for proper nested record initialization
					value = i.createRecordValue(rtv.RecordType, rtv.Methods)
				} else {
					value = &NilValue{}
				}
			} else {
				// Check if this is an array type
				// Task 9.225: Normalize to lowercase for case-insensitive lookups
				arrayTypeKey := "__array_type_" + strings.ToLower(typeName)
				if typeVal, ok := i.env.Get(arrayTypeKey); ok {
					if atv, ok := typeVal.(*ArrayTypeValue); ok {
						// Initialize with empty array value
						value = NewArrayValue(atv.ArrayType)
					} else {
						value = &NilValue{}
					}
				} else {
					// Check if this is a subrange type
					// Task 9.225: Normalize to lowercase for case-insensitive lookups
					subrangeTypeKey := "__subrange_type_" + strings.ToLower(typeName)
					if typeVal, ok := i.env.Get(subrangeTypeKey); ok {
						if stv, ok := typeVal.(*SubrangeTypeValue); ok {
							// Initialize with zero value (will be validated if assigned)
							value = &SubrangeValue{
								Value:        0,
								SubrangeType: stv.SubrangeType,
							}
						} else {
							value = &NilValue{}
						}
					} else {
						// Task 9.16.2: Check if this is an interface type
						if ifaceInfo, exists := i.interfaces[strings.ToLower(typeName)]; exists {
							// Initialize with nil interface instance
							value = &InterfaceInstance{
								Interface: ifaceInfo,
								Object:    nil, // nil until assigned
							}
						} else {
							// Initialize basic types with their zero values
							// Proper initialization allows implicit conversions to work with target type
							switch strings.ToLower(typeName) {
							case "integer":
								value = &IntegerValue{Value: 0}
							case "float":
								value = &FloatValue{Value: 0.0}
							case "string":
								value = &StringValue{Value: ""}
							case "boolean":
								value = &BooleanValue{Value: false}
							case "variant":
								// Task 9.227: Initialize Variant with nil/unassigned value
								value = &VariantValue{Value: nil, ActualType: nil}
							default:
								// Task 9.5.4: Check if this is a class type and create a typed nil value
								// This allows accessing class variables via nil instances: var b: TBase; b.ClassVar
								if _, exists := i.classes[typeName]; exists {
									value = &NilValue{ClassType: typeName}
								} else {
									value = &NilValue{}
								}
							}
						}
					}
				}
			}
		} else {
			value = &NilValue{}
		}
	}

	// Define all names with the same value/type
	// For multi-identifier declarations without initializers, each gets its own zero value
	var lastValue = value
	for _, name := range stmt.Names {
		// If there's an initializer, all names share the same value (but parser prevents this for multi-names)
		// If no initializer, need to create separate zero values for each variable
		var nameValue Value
		if stmt.Value != nil {
			// Single name with initializer - use the computed value
			nameValue = value
		} else {
			// No initializer - create a new zero value for each name
			// Must create separate instances to avoid aliasing
			nameValue = i.createZeroValue(stmt.Type)
		}
		i.env.Define(name.Value, nameValue)
		lastValue = nameValue
	}

	return lastValue
}

// createZeroValue creates a zero value for the given type
// This is used for multi-identifier declarations where each variable needs its own instance
func (i *Interpreter) createZeroValue(typeAnnotation *ast.TypeAnnotation) Value {
	if typeAnnotation == nil {
		return &NilValue{}
	}

	// Check for inline types stored in AST (arrays, function pointers)
	// Task: Fix negative array bounds - use AST node directly instead of parsing string
	if typeAnnotation.InlineType != nil {
		if arrayNode, ok := typeAnnotation.InlineType.(*ast.ArrayTypeNode); ok {
			arrayType := i.resolveArrayTypeNode(arrayNode)
			if arrayType != nil {
				return NewArrayValue(arrayType)
			}
			return &NilValue{}
		}
	}

	typeName := typeAnnotation.Name

	// Check for inline array types from string (fallback for older code)
	if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
		arrayType := i.parseInlineArrayType(typeName)
		if arrayType != nil {
			return NewArrayValue(arrayType)
		}
		return &NilValue{}
	}

	// Task 9.214: Check for inline set types
	if strings.HasPrefix(typeName, "set of ") {
		setType := i.parseInlineSetType(typeName)
		if setType != nil {
			return NewSetValue(setType)
		}
		return &NilValue{}
	}

	// Check if this is a record type
	// Task 9.225: Normalize to lowercase for case-insensitive lookups
	if typeVal, ok := i.env.Get("__record_type_" + strings.ToLower(typeName)); ok {
		if rtv, ok := typeVal.(*RecordTypeValue); ok {
			// Task 9.7e1: Use createRecordValue for proper nested record initialization
			return i.createRecordValue(rtv.RecordType, rtv.Methods)
		}
	}

	// Check if this is an array type
	// Task 9.225: Normalize to lowercase for case-insensitive lookups
	arrayTypeKey := "__array_type_" + strings.ToLower(typeName)
	if typeVal, ok := i.env.Get(arrayTypeKey); ok {
		if atv, ok := typeVal.(*ArrayTypeValue); ok {
			return NewArrayValue(atv.ArrayType)
		}
	}

	// Check if this is a subrange type
	// Task 9.225: Normalize to lowercase for case-insensitive lookups
	subrangeTypeKey := "__subrange_type_" + strings.ToLower(typeName)
	if typeVal, ok := i.env.Get(subrangeTypeKey); ok {
		if stv, ok := typeVal.(*SubrangeTypeValue); ok {
			return &SubrangeValue{
				Value:        0,
				SubrangeType: stv.SubrangeType,
			}
		}
	}

	// Task 9.1.3: Check if this is an interface type
	if ifaceInfo, exists := i.interfaces[strings.ToLower(typeName)]; exists {
		return &InterfaceInstance{
			Interface: ifaceInfo,
			Object:    nil,
		}
	}

	// Initialize basic types with their zero values
	switch strings.ToLower(typeName) {
	case "integer":
		return &IntegerValue{Value: 0}
	case "float":
		return &FloatValue{Value: 0.0}
	case "string":
		return &StringValue{Value: ""}
	case "boolean":
		return &BooleanValue{Value: false}
	case "variant":
		// Task 9.227: Initialize Variant with nil/unassigned value
		return &VariantValue{Value: nil, ActualType: nil}
	default:
		// Task 9.5.4: Check if this is a class type and create a typed nil value
		// This allows accessing class variables via nil instances: var b: TBase; b.ClassVar
		if _, exists := i.classes[typeName]; exists {
			return &NilValue{ClassType: typeName}
		}
		return &NilValue{}
	}
}

// evalConstDecl evaluates a const declaration statement
// Constants are immutable values stored in the environment.
// Immutability is enforced at semantic analysis time, so at runtime
// we simply evaluate the value and store it like a variable.
func (i *Interpreter) evalConstDecl(stmt *ast.ConstDecl) Value {
	// Constants must have a value
	if stmt.Value == nil {
		return newError("constant '%s' must have a value", stmt.Name.Value)
	}

	// Evaluate the constant value
	var value Value

	// Task 9.177: Special handling for anonymous record literals - they need type context
	if recordLit, ok := stmt.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
		// Anonymous record literal needs explicit type
		if stmt.Type == nil {
			return newError("anonymous record literal requires explicit type annotation")
		}
		typeName := stmt.Type.Name
		// Task 9.225: Normalize to lowercase for case-insensitive lookups
		recordTypeKey := "__record_type_" + strings.ToLower(typeName)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				// Temporarily set the type name for evaluation
				recordLit.TypeName = &ast.Identifier{Value: rtv.RecordType.Name}
				value = i.Eval(recordLit)
				recordLit.TypeName = nil
			} else {
				return newError("type '%s' is not a record type", typeName)
			}
		} else {
			return newError("unknown type '%s'", typeName)
		}
	} else {
		value = i.Eval(stmt.Value)
	}
	if isError(value) {
		return value
	}

	// Store the constant in the environment
	// Note: Immutability is enforced by semantic analysis, not at runtime
	i.env.Define(stmt.Name.Value, value)
	return value
}
