package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
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
		if i.ctx.ControlFlow().IsExit() {
			i.ctx.ControlFlow().Clear()
			break // Exit the program
		}
	}

	// If there's an uncaught exception, convert it to an error
	if i.exception != nil {
		exc := i.exception
		return newError("uncaught exception: %s", exc.Inspect())
	}

	// Clean up interface and object references (destructors called for global objects)
	i.cleanupInterfaceReferences(i.Env())

	return result
}

// evalVarDeclStatement evaluates a variable declaration statement.
// It defines a new variable in the current environment.
// External variables are marked with a special ExternalVarValue.
func (i *Interpreter) evalVarDeclStatement(stmt *ast.VarDeclStatement) Value {
	if stmt.IsExternal {
		return i.evalExternalVarDecl(stmt)
	}

	var value Value
	if stmt.Value != nil {
		val := i.evalVarDeclInitializer(stmt)
		if isError(val) {
			return val
		}

		// This is important for operations that raise exceptions (like invalid casts)
		if i.exception != nil {
			return nil
		}
		value = val
	} else {
		// Just for safety, though the loop will regenerate it for each variable
		value = i.createZeroValue(stmt.Type)
	}

	return i.defineVariables(stmt, value)
}

func (i *Interpreter) evalExternalVarDecl(stmt *ast.VarDeclStatement) Value {
	// External variables only apply to single declarations
	if len(stmt.Names) != 1 {
		return newError("external keyword cannot be used with multiple variable names")
	}
	// Store a special marker for external variables
	externalName := stmt.ExternalName
	if externalName == "" {
		externalName = stmt.Names[0].Value
	}
	value := &ExternalVarValue{
		Name:         stmt.Names[0].Value,
		ExternalName: externalName,
	}
	i.Env().Define(stmt.Names[0].Value, value)
	return value
}

func (i *Interpreter) evalVarDeclInitializer(stmt *ast.VarDeclStatement) Value {
	var value Value
	if arrayLit, ok := stmt.Value.(*ast.ArrayLiteralExpression); ok {
		var expected *types.ArrayType
		if stmt.Type != nil {
			arrType, errVal := i.arrayTypeByName(stmt.Type.String(), stmt)
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

		typeName := stmt.Type.String()
		recordTypeKey := "__record_type_" + ident.Normalize(typeName)
		if typeVal, ok := i.Env().Get(recordTypeKey); ok {
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
		if isError(value) {
			return value
		}
	} else {
		value = i.Eval(stmt.Value)
		if isError(value) {
			return value
		}
	}

	return i.postProcessVarInit(value, stmt.Type)
}

func (i *Interpreter) postProcessVarInit(value Value, typeExpr ast.TypeExpression) Value {
	// If declaring a subrange variable with an initializer, wrap and validate
	if typeExpr != nil {
		typeName := typeExpr.String()
		handledSubrange := false
		if subrangeType := i.typeSystem.LookupSubrangeType(typeName); subrangeType != nil {
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
				SubrangeType: subrangeType,
			}
			if err := subrangeVal.ValidateAndSet(intValue); err != nil {
				return &ErrorValue{Message: err.Error()}
			}
			value = subrangeVal
			handledSubrange = true
		}
		if !handledSubrange {
			if converted, ok := i.tryImplicitConversion(value, typeName); ok {
				value = converted
			}
		}

		// Box value if target type is Variant
		if ident.Equal(typeName, "Variant") {
			value = BoxVariant(value)
		}
	}
	return value
}

func (i *Interpreter) defineVariables(stmt *ast.VarDeclStatement, initialValue Value) Value {
	// Define all names with the same value/type
	// For multi-identifier declarations without initializers, each gets its own zero value
	var lastValue = initialValue
	for _, name := range stmt.Names {
		// If there's an initializer, all names share the same value (but parser prevents this for multi-names)
		// If no initializer, need to create separate zero values for each variable
		var nameValue Value
		if stmt.Value != nil {
			// Single name with initializer - use the computed value
			if _, isIndexExpr := stmt.Value.(*ast.IndexExpression); isIndexExpr {
				nameValue = initialValue
			} else {
				nameValue = cloneIfCopyable(initialValue)
			}
			// Check if the type annotation is an interface type, wrap the value in an InterfaceInstance
			if stmt.Type != nil {
				typeName := stmt.Type.String()
				if ifaceInfo := i.lookupInterfaceInfo(typeName); ifaceInfo != nil {
					// Check if the value is already an InterfaceInstance
					if _, alreadyInterface := nameValue.(*InterfaceInstance); !alreadyInterface {
						// Check if the value is an ObjectInstance
						if objInst, isObj := nameValue.(*ObjectInstance); isObj {
							// Validate that the object's class implements the interface
							concreteClass, ok := objInst.Class.(*ClassInfo)
							if !ok {
								return i.newErrorWithLocation(stmt, "object has invalid class type")
							}
							if !classImplementsInterface(concreteClass, ifaceInfo) {
								return i.newErrorWithLocation(stmt, "class '%s' does not implement interface '%s'",
									objInst.Class.GetName(), ifaceInfo.Name)
							}
							// Wrap the object in an InterfaceInstance
							nameValue = NewInterfaceInstance(ifaceInfo, objInst)
						}
					}
				}
			}
		} else {
			// No initializer - create a new zero value for each name
			// Must create separate instances to avoid aliasing
			nameValue = i.createZeroValue(stmt.Type)
		}
		i.Env().Define(name.Value, nameValue)
		lastValue = nameValue
	}

	return lastValue
}

// createZeroValue creates a zero value for the given type
// This is used for multi-identifier declarations where each variable needs its own instance
func (i *Interpreter) createZeroValue(typeExpr ast.TypeExpression) Value {
	if typeExpr == nil {
		return &NilValue{}
	}

	// Check for array types
	if arrayNode, ok := typeExpr.(*ast.ArrayTypeNode); ok {
		arrayType := i.resolveArrayTypeNode(arrayNode)
		if arrayType != nil {
			return NewArrayValue(arrayType)
		}
		return &NilValue{}
	}

	return i.createZeroValueFromString(typeExpr.String())
}

func (i *Interpreter) createZeroValueFromString(typeName string) Value {
	lowerTypeName := ident.Normalize(typeName)

	// Check for inline types
	if val := i.createInlineZeroValue(lowerTypeName); val != nil {
		return val
	}

	// Check for registered types (records, sets, arrays, subranges, interfaces)
	if val := i.createRegisteredZeroValue(typeName); val != nil {
		return val
	}

	// Check for basic types and classes
	return i.createBasicZeroValue(typeName)
}

func (i *Interpreter) createInlineZeroValue(lowerTypeName string) Value {
	// Check for inline array types from string (fallback for older code)
	if strings.HasPrefix(lowerTypeName, "array of ") || strings.HasPrefix(lowerTypeName, "array[") {
		arrayType := i.parseInlineArrayType(lowerTypeName)
		if arrayType != nil {
			return NewArrayValue(arrayType)
		}
		return &NilValue{}
	}

	// Check for inline set types (e.g., "set of TColor")
	if strings.HasPrefix(lowerTypeName, "set of ") {
		setType := i.parseInlineSetType(lowerTypeName)
		if setType != nil {
			return NewSetValue(setType)
		}
		return &NilValue{}
	}

	return nil
}

func (i *Interpreter) createRegisteredZeroValue(typeName string) Value {
	normalizedName := ident.Normalize(typeName)

	// Check if this is a record type
	if typeVal, ok := i.Env().Get("__record_type_" + normalizedName); ok {
		if rtv, ok := typeVal.(*RecordTypeValue); ok {
			// Use createRecordValue for proper nested record initialization
			return i.createRecordValue(rtv.RecordType)
		}
	}

	// Check if this is a named set type (registered during type section evaluation)
	if setTypeVal, ok := i.Env().Get("__set_type_" + normalizedName); ok {
		if stv, ok := setTypeVal.(interface{ GetSetType() *types.SetType }); ok {
			return NewSetValue(stv.GetSetType())
		}
	}

	// Check if this is an array type
	if arrayType := i.typeSystem.LookupArrayType(typeName); arrayType != nil {
		return NewArrayValue(arrayType)
	}

	// Check if this is a subrange type
	if subrangeType := i.typeSystem.LookupSubrangeType(typeName); subrangeType != nil {
		return &SubrangeValue{
			Value:        subrangeType.LowBound,
			SubrangeType: subrangeType,
		}
	}

	// Check if this is an interface type
	if ifaceInfo := i.lookupInterfaceInfo(typeName); ifaceInfo != nil {
		return &InterfaceInstance{
			Interface: ifaceInfo,
			Object:    nil,
		}
	}

	return nil
}

func (i *Interpreter) createBasicZeroValue(typeName string) Value {
	// Initialize basic types with their zero values
	switch ident.Normalize(typeName) {
	case "integer":
		return &IntegerValue{Value: 0}
	case "float":
		return &FloatValue{Value: 0.0}
	case "string":
		return &StringValue{Value: ""}
	case "boolean":
		return &BooleanValue{Value: false}
	case "variant":
		// Initialize Variant with nil/unassigned value
		return &VariantValue{Value: nil, ActualType: nil}
	default:
		// Check if this is a class type and create a typed nil value
		if _, exists := i.classes[ident.Normalize(typeName)]; exists {
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

	// Special handling for anonymous record literals (needs type context)
	if recordLit, ok := stmt.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
		// Anonymous record literal needs explicit type
		if stmt.Type == nil {
			return newError("anonymous record literal requires explicit type annotation")
		}
		typeName := stmt.Type.String()
		recordTypeKey := "__record_type_" + ident.Normalize(typeName)
		if typeVal, ok := i.Env().Get(recordTypeKey); ok {
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
	i.Env().Define(stmt.Name.Value, value)
	return value
}
