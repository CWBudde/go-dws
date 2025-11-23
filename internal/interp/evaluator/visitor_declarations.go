package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains visitor methods for declaration AST nodes.
// Phase 3.5.2: Visitor pattern implementation for declarations.
//
// Declarations define types, functions, classes, etc. and register them
// in the appropriate registries.

// ============================================================================
// Local Value Types for Declaration Processing
// ============================================================================
//
// These types are defined locally to avoid circular dependencies between
// evaluator and interp packages. They will eventually be moved to the runtime
// package as part of the AST-free runtime types migration (Phase 3.5.37).

// EnumTypeValue is an internal value type used to store enum type metadata
// in the environment.
type EnumTypeValue struct {
	EnumType *types.EnumType
}

// Type returns "ENUM_TYPE".
func (e *EnumTypeValue) Type() string {
	return "ENUM_TYPE"
}

// String returns the enum type name.
func (e *EnumTypeValue) String() string {
	if e.EnumType == nil {
		return "enum type <nil>"
	}
	return e.EnumType.Name
}

// ArrayTypeValue is an internal value that stores array type metadata in the environment.
type ArrayTypeValue struct {
	ArrayType *types.ArrayType
	Name      string
}

// Type returns "ARRAY_TYPE".
func (a *ArrayTypeValue) Type() string {
	return "ARRAY_TYPE"
}

// String returns the array type name.
func (a *ArrayTypeValue) String() string {
	return "array type " + a.Name
}

// VisitFunctionDecl evaluates a function declaration.
// Task 3.5.51: Delegates to adapter.EvalFunctionDeclaration() for function registration logic.
// This includes registering in i.functions map, handling method implementations (ClassName != nil),
// supporting function overloading, and updating ClassInfo/RecordTypeValue for class/record methods.
func (e *Evaluator) VisitFunctionDecl(node *ast.FunctionDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalFunctionDeclaration(node, ctx)
}

// VisitClassDecl evaluates a class declaration.
// Task 3.5.50: Delegates to adapter.EvalClassDeclaration() for class registration logic.
// This includes building ClassInfo/ClassMetadata, handling inheritance, methods, properties,
// constructors, destructors, operator overloads, and virtual method table construction.
func (e *Evaluator) VisitClassDecl(node *ast.ClassDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalClassDeclaration(node, ctx)
}

// VisitInterfaceDecl evaluates an interface declaration.
// Task 3.5.50: Delegates to adapter.EvalInterfaceDeclaration() for interface registration logic.
// This includes building InterfaceInfo, handling parent interface inheritance,
// method signature registration (InterfaceMethodDecl to FunctionDecl conversion),
// and property registration.
func (e *Evaluator) VisitInterfaceDecl(node *ast.InterfaceDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalInterfaceDeclaration(node, ctx)
}

// VisitOperatorDecl evaluates an operator declaration (operator overloading).
// Task 3.5.51: Delegates to adapter.EvalOperatorDeclaration() for operator registration logic.
// This includes registering conversion operators in i.conversions, global operators in i.globalOperators,
// and skipping class operators (handled during class declaration).
func (e *Evaluator) VisitOperatorDecl(node *ast.OperatorDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalOperatorDeclaration(node, ctx)
}

// VisitEnumDecl evaluates an enum declaration.
// Phase 3.5.48: Migrated from adapter to Evaluator.
// Registers enum type and values in the environment.
func (e *Evaluator) VisitEnumDecl(node *ast.EnumDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil enum declaration")
	}

	enumName := node.Name.Value

	// Build the enum type from the declaration
	enumValues := make(map[string]int)
	orderedNames := make([]string, 0, len(node.Values))

	// Calculate ordinal values (explicit or implicit)
	currentOrdinal := 0
	flagBitPosition := 0 // For flags enums, track the bit position (2^n)

	for _, enumValue := range node.Values {
		valueName := enumValue.Name

		// Determine ordinal value (explicit or implicit)
		var ordinalValue int
		if enumValue.Value != nil {
			// Explicit value provided
			ordinalValue = *enumValue.Value
			if node.Flags {
				// For flags, explicit values must be powers of 2
				if ordinalValue <= 0 || (ordinalValue&(ordinalValue-1)) != 0 {
					return e.newError(node, "enum '%s' value '%s' (%d) must be a power of 2 for flags enum",
						enumName, valueName, ordinalValue)
				}
				// For flags, calculate bit position using bit manipulation
				// This is more efficient than a loop and works for all valid powers of 2
				bitPos := 0
				temp := ordinalValue
				for temp > 1 {
					temp >>= 1
					bitPos++
				}
				flagBitPosition = bitPos + 1
			} else {
				// For regular enums, update current ordinal
				currentOrdinal = ordinalValue + 1
			}
		} else {
			// Implicit value
			if node.Flags {
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
	if node.Scoped || node.Flags {
		enumType = types.NewScopedEnumType(enumName, enumValues, orderedNames, node.Flags)
	} else {
		enumType = types.NewEnumType(enumName, enumValues, orderedNames)
	}

	// Register each enum value in the symbol table as a constant
	// For scoped enums (enum/flags keyword), skip global registration -
	// values are only accessible via qualified access (Type.Value)
	if !node.Scoped {
		for valueName, ordinalValue := range enumValues {
			enumVal := &runtime.EnumValue{
				TypeName:     enumName,
				ValueName:    valueName,
				OrdinalValue: ordinalValue,
			}
			ctx.Env().Define(valueName, enumVal)
		}
	}

	// Store enum type metadata in environment with special key
	// This allows variable declarations to resolve the type
	enumTypeKey := "__enum_type_" + ident.Normalize(enumName)
	ctx.Env().Define(enumTypeKey, &EnumTypeValue{EnumType: enumType})

	// Register enum type name as a TypeMetaValue
	// This allows the type name to be used as a runtime value in expressions
	// like High(TColor) or Low(TColor), just like built-in types (Integer, Float, etc.)
	ctx.Env().Define(enumName, &runtime.TypeMetaValue{
		TypeInfo: enumType,
		TypeName: enumName,
	})

	return &runtime.NilValue{}
}

// VisitRecordDecl evaluates a record declaration.
// Task 3.5.49: Migrated from Interpreter to Evaluator.
// Registers the record type with field definitions, methods, constants, and class variables.
func (e *Evaluator) VisitRecordDecl(node *ast.RecordDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil record declaration")
	}

	recordName := node.Name.Value

	// Build the record type from the declaration
	fields := make(map[string]types.Type)
	fieldDecls := make(map[string]*ast.FieldDecl)

	for _, field := range node.Fields {
		fieldName := field.Name.Value

		// Handle type inference for fields
		var fieldType types.Type
		if field.Type != nil {
			// Explicit type - resolve it
			resolvedType := e.adapter.ResolveTypeFromExpression(field.Type)
			if resolvedType == nil {
				return e.newError(node, "unknown or invalid type for field '%s' in record '%s'", fieldName, recordName)
			}
			var ok bool
			fieldType, ok = resolvedType.(types.Type)
			if !ok {
				return e.newError(node, "internal error: invalid type resolution for field '%s'", fieldName)
			}
		} else if field.InitValue != nil {
			// Type inference from initializer
			initValue := e.Eval(field.InitValue, ctx)
			if isError(initValue) {
				return initValue
			}
			fieldTypeAny := e.adapter.GetValueType(initValue)
			if fieldTypeAny == nil {
				return e.newError(node, "cannot infer type for field '%s' in record '%s'", fieldName, recordName)
			}
			var ok bool
			fieldType, ok = fieldTypeAny.(types.Type)
			if !ok {
				return e.newError(node, "internal error: invalid type inference for field '%s'", fieldName)
			}
		} else {
			return e.newError(node, "field '%s' in record '%s' must have either a type or initializer", fieldName, recordName)
		}

		// Use lowercase key for case-insensitive access
		fieldNameLower := ident.Normalize(fieldName)
		fields[fieldNameLower] = fieldType
		fieldDecls[fieldNameLower] = field
	}

	// Create the record type
	recordType := types.NewRecordType(recordName, fields)

	// Build maps for instance methods and static methods
	methods := make(map[string]*ast.FunctionDecl)
	staticMethods := make(map[string]*ast.FunctionDecl)
	for _, method := range node.Methods {
		methodKey := ident.Normalize(method.Name.Value)
		if method.IsClassMethod {
			staticMethods[methodKey] = method
		} else {
			methods[methodKey] = method
		}
	}

	// Evaluate record constants
	constants := make(map[string]Value)
	for _, constant := range node.Constants {
		constName := constant.Name.Value
		constValue := e.Eval(constant.Value, ctx)
		if isError(constValue) {
			return constValue
		}
		constants[ident.Normalize(constName)] = constValue
	}

	// Initialize class variables
	classVars := make(map[string]Value)
	for _, classVar := range node.ClassVars {
		varName := classVar.Name.Value
		var varValue Value

		if classVar.InitValue != nil {
			// Evaluate the initializer
			varValue = e.Eval(classVar.InitValue, ctx)
			if isError(varValue) {
				return varValue
			}
		} else {
			// Use type to determine zero value
			var varType types.Type
			if classVar.Type != nil {
				resolvedType := e.adapter.ResolveTypeFromExpression(classVar.Type)
				if resolvedType == nil {
					return e.newError(node, "unknown type for class variable '%s' in record '%s'", varName, recordName)
				}
				var ok bool
				varType, ok = resolvedType.(types.Type)
				if !ok {
					return e.newError(node, "internal error: invalid type resolution for class variable '%s'", varName)
				}
			}
			varValue = e.adapter.GetZeroValueForType(varType)
		}

		classVars[ident.Normalize(varName)] = varValue
	}

	// Process properties
	for _, prop := range node.Properties {
		propName := prop.Name.Value
		propNameLower := ident.Normalize(propName)

		// Resolve property type
		propTypeAny := e.adapter.ResolveTypeFromExpression(prop.Type)
		if propTypeAny == nil {
			return e.newError(node, "unknown type for property '%s' in record '%s'", propName, recordName)
		}
		propType, ok := propTypeAny.(types.Type)
		if !ok {
			return e.newError(node, "internal error: invalid type resolution for property '%s'", propName)
		}

		// Create property info
		propInfo := &types.RecordPropertyInfo{
			Name:       propName,
			Type:       propType,
			ReadField:  prop.ReadField,
			WriteField: prop.WriteField,
			IsDefault:  prop.IsDefault,
		}

		// Store in recordType.Properties (case-insensitive)
		recordType.Properties[propNameLower] = propInfo
	}

	// Build RecordTypeValue using adapter
	recordTypeValue := e.adapter.BuildRecordTypeValue(
		recordName,
		recordType,
		fieldDecls,
		methods,
		staticMethods,
		constants,
		classVars,
	)

	// Register in environment and TypeSystem using adapter
	e.adapter.RegisterRecordTypeInEnvironment(recordName, recordTypeValue, ctx)

	return &runtime.NilValue{}
}

// VisitHelperDecl evaluates a helper declaration (type extension).
// Task 3.5.50: Delegates to adapter.EvalHelperDeclaration() for helper registration logic.
// This includes building HelperInfo, resolving target type, handling parent helper inheritance,
// method registration (user-defined and builtin), property registration,
// and class variable/constant initialization.
func (e *Evaluator) VisitHelperDecl(node *ast.HelperDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalHelperDeclaration(node, ctx)
}

// VisitArrayDecl evaluates an array type declaration.
// Phase 3.5.48: Migrated from adapter to Evaluator.
// Example: type TMyArray = array[1..10] of Integer;
func (e *Evaluator) VisitArrayDecl(node *ast.ArrayDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil array declaration")
	}

	arrayName := node.Name.Value

	// Build the array type from the declaration
	arrayTypeAnnotation := node.ArrayType
	if arrayTypeAnnotation == nil {
		return e.newError(node, "invalid array type declaration")
	}

	// Resolve the element type
	elementTypeName := arrayTypeAnnotation.ElementType.String()
	elementType := e.resolveTypeForDeclaration(elementTypeName, ctx)
	if elementType == nil {
		return e.newError(node, "unknown element type '%s'", elementTypeName)
	}

	// Create the array type
	var arrayType *types.ArrayType
	if arrayTypeAnnotation.IsDynamic() {
		arrayType = types.NewDynamicArrayType(elementType)
	} else {
		// Evaluate bound expressions at runtime
		lowBoundVal := e.Eval(arrayTypeAnnotation.LowBound, ctx)
		if isError(lowBoundVal) {
			return lowBoundVal
		}
		highBoundVal := e.Eval(arrayTypeAnnotation.HighBound, ctx)
		if isError(highBoundVal) {
			return highBoundVal
		}

		// Extract integer values
		lowBound, ok := lowBoundVal.(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "array lower bound must be an integer")
		}
		highBound, ok := highBoundVal.(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "array upper bound must be an integer")
		}

		arrayType = types.NewStaticArrayType(elementType, int(lowBound.Value), int(highBound.Value))
	}

	// Store array type in environment with a special prefix
	// This allows var declarations to look up the type
	arrayTypeValue := &ArrayTypeValue{
		Name:      arrayName,
		ArrayType: arrayType,
	}

	// Task 3.5.48: Register in interpreter's i.env via adapter method.
	// This is necessary because IsArrayType and CreateArrayZeroValue look in i.env,
	// not in ctx.Env(). The adapter ensures the type is stored where lookups expect it.
	e.adapter.RegisterArrayTypeInEnvironment(arrayName, arrayTypeValue)

	return &runtime.NilValue{} // Type declarations don't return a value
}

// VisitTypeDeclaration evaluates a type alias declaration.
// Task 3.5.49: Migrated from Interpreter to Evaluator.
// Handles type aliases, subrange types, and function pointer types.
func (e *Evaluator) VisitTypeDeclaration(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil type declaration")
	}

	// Handle subrange types
	if node.IsSubrange {
		// Evaluate low bound
		lowBoundVal := e.Eval(node.LowBound, ctx)
		if isError(lowBoundVal) {
			return lowBoundVal
		}

		// Evaluate high bound
		highBoundVal := e.Eval(node.HighBound, ctx)
		if isError(highBoundVal) {
			return highBoundVal
		}

		// Build and register subrange type using adapter
		subrangeTypeValue, err := e.adapter.BuildSubrangeTypeValue(node.Name.Value, lowBoundVal, highBoundVal)
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}

		e.adapter.RegisterSubrangeTypeInEnvironment(node.Name.Value, subrangeTypeValue, ctx)
		return &runtime.NilValue{}
	}

	// Handle function pointer type declarations
	if node.IsFunctionPointer {
		if node.FunctionPointerType == nil {
			return e.newError(node, "function pointer type declaration has no type information")
		}

		// Function pointer types are validated by the semantic analyzer
		// At runtime, we just register the type name as existing
		typeKey := "__funcptr_type_" + ident.Normalize(node.Name.Value)
		ctx.Env().Define(typeKey, &runtime.StringValue{Value: "function_pointer_type"})

		return &runtime.NilValue{}
	}

	// Handle type aliases
	if node.IsAlias {
		// Check for inline/complex type expressions
		switch aliasType := node.AliasedType.(type) {
		case *ast.ClassOfTypeNode:
			// Metaclass types - semantic analyzer handles them
			return &runtime.NilValue{}
		case *ast.SetTypeNode:
			// Set types - semantic analyzer handles them
			return &runtime.NilValue{}
		case *ast.ArrayTypeNode:
			// Task 3.5.48: Inline array types need runtime registration for variable declarations.
			// Example: type TMyArray = array[1..10] of Integer;
			// The type must be registered so IsArrayType and CreateArrayZeroValue can find it.
			arrayType, err := e.adapter.ResolveArrayTypeNode(aliasType)
			if err != nil {
				return e.newError(node, "failed to resolve array type: %v", err)
			}
			if arrayType != nil {
				// Create ArrayTypeValue and register it
				typedArrayType, ok := arrayType.(*types.ArrayType)
				if !ok {
					return e.newError(node, "resolved type is not an ArrayType")
				}
				arrayTypeValue := &ArrayTypeValue{
					Name:      node.Name.Value,
					ArrayType: typedArrayType,
				}
				e.adapter.RegisterArrayTypeInEnvironment(node.Name.Value, arrayTypeValue)
			}
			return &runtime.NilValue{}
		case *ast.FunctionPointerTypeNode:
			// Function pointer types - already handled earlier
			return &runtime.NilValue{}
		}

		// For TypeAnnotation with InlineType, also skip runtime resolution
		if typeAnnot, ok := node.AliasedType.(*ast.TypeAnnotation); ok && typeAnnot.InlineType != nil {
			return &runtime.NilValue{}
		}

		// Resolve the aliased type by name (handles simple named types only)
		aliasedType, err := e.adapter.GetType(node.AliasedType.String())
		if err != nil {
			return e.newError(node, "unknown type '%s' in type alias", node.AliasedType.String())
		}

		// Build and register type alias using adapter
		typeAliasValue := e.adapter.BuildTypeAliasValue(node.Name.Value, aliasedType)
		e.adapter.RegisterTypeAliasInEnvironment(node.Name.Value, typeAliasValue, ctx)

		return &runtime.NilValue{}
	}

	// Non-alias type declarations (future)
	return e.newError(node, "non-alias type declarations not yet supported")
}

// VisitSetDecl evaluates a set declaration.
// Phase 3.5.48: Migrated from adapter to Evaluator.
// Set types are already registered by the semantic analyzer, so we just return nil.
func (e *Evaluator) VisitSetDecl(node *ast.SetDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil set declaration")
	}

	// Set type already registered by semantic analyzer
	// Just return nil value to indicate successful processing
	return &runtime.NilValue{}
}

// ============================================================================
// Helper Methods for Declaration Processing
// ============================================================================

// resolveTypeForDeclaration resolves a type name to its semantic type.
// This handles built-in types, enum types, record types, array types, type aliases, and subranges.
// Returns nil if the type cannot be resolved.
//
// Phase 3.5.48: Helper for array/enum/record type resolution in declarations.
// Uses adapter.GetType() to delegate to the full Interpreter.resolveType() which handles
// all type kinds including type aliases and subranges.
func (e *Evaluator) resolveTypeForDeclaration(typeName string, ctx *ExecutionContext) types.Type {
	// Delegate to adapter's GetType which calls Interpreter.resolveType()
	// This handles: built-ins, enums, records, arrays, type aliases, subranges, inline arrays
	resolvedType, err := e.adapter.GetType(typeName)
	if err != nil {
		return nil
	}

	// Convert from any to types.Type
	if typ, ok := resolvedType.(types.Type); ok {
		return typ
	}

	return nil
}
