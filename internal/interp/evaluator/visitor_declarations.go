package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains visitor methods for declaration AST nodes.
// Phase 3.5.2: Visitor pattern implementation for declarations.
//
// Declarations define types, functions, classes, etc. and register them
// in the appropriate registries.

// VisitFunctionDecl evaluates a function declaration.
// Task 3.5.7: Migrated from Interpreter.evalFunctionDeclaration to Evaluator visitor.
//
// This method handles two cases:
//   - Global functions: Registered in TypeSystem.FunctionRegistry
//   - Method implementations (node.ClassName != nil): Delegated to adapter for ClassInfo handling
//
// The split design is intentional: method implementations require ClassInfo internals
// (VMT rebuild, descendant propagation) which belong in the Interpreter. Global functions
// can be fully handled via the TypeSystem's FunctionRegistry.
func (e *Evaluator) VisitFunctionDecl(node *ast.FunctionDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil function declaration")
	}

	// Method implementations (TClass.Method syntax) are delegated to adapter
	// They require ClassInfo internals: VMT rebuild, descendant propagation
	if node.ClassName != nil {
		return e.adapter.EvalMethodImplementation(node)
	}

	// Global function - register in FunctionRegistry via TypeSystem
	// RegisterFunctionOrReplace handles the interface/implementation pattern:
	// - Forward declarations (no body) are appended
	// - Implementations (with body) replace matching declarations by signature
	e.typeSystem.RegisterFunctionOrReplace(node.Name.Value, node)

	return &runtime.NilValue{}
}

// VisitClassDecl evaluates a class declaration.
func (e *Evaluator) VisitClassDecl(node *ast.ClassDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Class registry available via adapter.LookupClass()
	// TODO: Move class registration logic here (use adapter type system methods)
	return e.adapter.EvalNode(node)
}

// VisitInterfaceDecl evaluates an interface declaration.
func (e *Evaluator) VisitInterfaceDecl(node *ast.InterfaceDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Interface registry available via adapter.LookupInterface()
	// TODO: Move interface registration logic here (use adapter type system methods)
	return e.adapter.EvalNode(node)
}

// VisitOperatorDecl evaluates an operator declaration (operator overloading).
// Task 3.5.14: Migrated from Interpreter.evalOperatorDeclaration to Evaluator visitor.
//
// This method handles global and conversion operator declarations:
//   - Global operators: registered via TypeSystem.Operators()
//   - Conversion operators: registered via TypeSystem.Conversions()
//
// Class operators (Kind == OperatorKindClass) are handled separately during
// class declaration evaluation, so this method returns NilValue for them.
func (e *Evaluator) VisitOperatorDecl(node *ast.OperatorDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil operator declaration")
	}

	// Class operators are registered during class declaration evaluation
	if node.Kind == ast.OperatorKindClass {
		return &runtime.NilValue{}
	}

	// Validate binding exists
	if node.Binding == nil {
		return e.newError(node, "operator '%s' missing binding", node.OperatorSymbol)
	}

	// Normalize operand types for consistent lookup
	operandTypes := make([]string, len(node.OperandTypes))
	for idx, operand := range node.OperandTypes {
		opRand := operand.String()
		operandTypes[idx] = interptypes.NormalizeTypeAnnotation(opRand)
	}

	// Handle conversion operators
	if node.Kind == ast.OperatorKindConversion {
		if len(operandTypes) != 1 {
			return e.newError(node, "conversion operator '%s' requires exactly one operand", node.OperatorSymbol)
		}
		if node.ReturnType == nil {
			return e.newError(node, "conversion operator '%s' requires a return type", node.OperatorSymbol)
		}

		targetType := interptypes.NormalizeTypeAnnotation(node.ReturnType.String())
		entry := &interptypes.ConversionEntry{
			From:        operandTypes[0],
			To:          targetType,
			BindingName: ident.Normalize(node.Binding.Value),
			Implicit:    ident.Equal(node.OperatorSymbol, "implicit"),
		}

		if err := e.typeSystem.Conversions().Register(entry); err != nil {
			return e.newError(node, "conversion from %s to %s already defined", operandTypes[0], targetType)
		}
		return &runtime.NilValue{}
	}

	// Handle global operators
	entry := &interptypes.OperatorEntry{
		Operator:     node.OperatorSymbol,
		OperandTypes: operandTypes,
		BindingName:  ident.Normalize(node.Binding.Value),
	}

	if err := e.typeSystem.Operators().Register(entry); err != nil {
		return e.newError(node, "operator '%s' already defined for operand types (%s)", node.OperatorSymbol, strings.Join(operandTypes, ", "))
	}

	return &runtime.NilValue{}
}

// VisitEnumDecl evaluates an enum declaration.
// Task 3.5.11: Migrated from Interpreter.evalEnumDeclaration to Evaluator visitor.
//
// This method:
//  1. Builds the enum type from the AST declaration
//  2. Calculates ordinal values (explicit or implicit)
//  3. For flags enums, validates values are powers of 2
//  4. For unscoped enums, registers values in the environment
//  5. Registers enum type metadata in the TypeSystem
//  6. Creates a TypeMetaValue for the enum type name
func (e *Evaluator) VisitEnumDecl(node *ast.EnumDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil enum declaration")
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
	enumTypeValue := runtime.NewEnumTypeValue(enumType)
	ctx.Env().Define(enumTypeKey, enumTypeValue)

	// Register in TypeSystem for consistent type lookups
	e.typeSystem.RegisterEnumType(enumName, enumTypeValue)

	// Register enum type name as a TypeMetaValue
	// This allows the type name to be used as a runtime value in expressions
	// like High(TColor) or Low(TColor), just like built-in types (Integer, Float, etc.)
	typeMetaValue := &runtime.TypeMetaValue{
		TypeInfo: enumType,
		TypeName: enumName,
	}
	ctx.Env().Define(enumName, typeMetaValue)

	return &runtime.NilValue{}
}

// VisitRecordDecl evaluates a record declaration.
func (e *Evaluator) VisitRecordDecl(node *ast.RecordDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Record registry available via adapter.LookupRecord()
	// TODO: Move record registration logic here (use adapter type system methods)
	return e.adapter.EvalNode(node)
}

// VisitHelperDecl evaluates a helper declaration (type extension).
// Task 3.5.12: Migrated from Interpreter.evalHelperDeclaration to Evaluator visitor.
//
// This method handles helper declarations of the form:
//
//	type TStringHelper = helper for String ... end;
//	type TPointHelper = record helper for TPoint ... end;
//	type TChildHelper = helper(TParentHelper) for String ... end;
//
// It resolves the target type, handles parent helper inheritance,
// registers methods/properties, and registers the helper in the TypeSystem.
func (e *Evaluator) VisitHelperDecl(node *ast.HelperDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return &runtime.NilValue{}
	}

	// 3.5.12.2: Resolve the target type using evaluator's type resolution
	targetType, err := e.ResolveTypeFromAnnotation(node.ForType)
	if err != nil {
		return e.newError(node, "unknown target type '%s' for helper '%s'",
			node.ForType.String(), node.Name.Value)
	}

	// Create helper info via adapter
	helperInfo := e.adapter.CreateHelperInfo(node.Name.Value, targetType, node.IsRecordHelper)
	if helperInfo == nil {
		return e.newError(node, "failed to create helper info for '%s'", node.Name.Value)
	}

	// 3.5.12.3: Resolve parent helper if specified
	if node.ParentHelper != nil {
		parentHelperName := node.ParentHelper.Value

		// Search all registered helpers for parent by name
		var foundParent interface{}
		for _, helpers := range e.typeSystem.AllHelpers() {
			for _, helper := range helpers {
				if ident.Equal(e.adapter.GetHelperName(helper), parentHelperName) {
					foundParent = helper
					break
				}
			}
			if foundParent != nil {
				break
			}
		}

		if foundParent == nil {
			return e.newError(node.ParentHelper,
				"unknown parent helper '%s' for helper '%s'",
				parentHelperName, node.Name.Value)
		}

		// Verify target type compatibility
		if !e.adapter.VerifyHelperTargetTypeMatch(foundParent, targetType) {
			return e.newError(node.ParentHelper,
				"parent helper '%s' extends different type than child helper '%s'",
				parentHelperName, node.Name.Value)
		}

		// Set up inheritance chain
		e.adapter.SetHelperParent(helperInfo, foundParent)
	}

	// Register methods (case-insensitive keys)
	for _, method := range node.Methods {
		e.adapter.AddHelperMethod(helperInfo, ident.Normalize(method.Name.Value), method)
	}

	// Register properties
	for _, prop := range node.Properties {
		propType, propErr := e.ResolveTypeFromAnnotation(prop.Type)
		if propErr != nil {
			return e.newError(prop, "unknown type '%s' for property '%s'",
				prop.Type.String(), prop.Name.Value)
		}
		e.adapter.AddHelperProperty(helperInfo, prop, propType)
	}

	// Initialize class variables
	for _, classVar := range node.ClassVars {
		varType, varErr := e.resolveTypeName(classVar.Type.String(), ctx)
		if varErr != nil {
			return e.newError(classVar, "unknown type for class variable '%s'",
				classVar.Name.Value)
		}
		defaultValue := e.GetDefaultValue(varType)
		e.adapter.AddHelperClassVar(helperInfo, classVar.Name.Value, defaultValue)
	}

	// Initialize class constants (evaluate values)
	for _, classConst := range node.ClassConsts {
		constValue := e.Eval(classConst.Value, ctx)
		if isError(constValue) {
			return constValue
		}
		e.adapter.AddHelperClassConst(helperInfo, classConst.Name.Value, constValue)
	}

	// 3.5.12.4: Register via TypeSystem.RegisterHelper()
	typeName := ident.Normalize(targetType.String())
	e.typeSystem.RegisterHelper(typeName, helperInfo)

	// Also maintain legacy map for backward compatibility during migration
	e.adapter.RegisterHelperLegacy(typeName, helperInfo)

	// Also register by simple type name for lookup compatibility
	simpleTypeName := ident.Normalize(extractSimpleTypeName(targetType.String()))
	if simpleTypeName != typeName {
		e.typeSystem.RegisterHelper(simpleTypeName, helperInfo)
		e.adapter.RegisterHelperLegacy(simpleTypeName, helperInfo)
	}

	return &runtime.NilValue{}
}

// VisitArrayDecl evaluates an array type declaration.
// Task 3.5.13: Migrated from Interpreter.evalArrayDeclaration to Evaluator visitor.
//
// This method handles array type declarations of the form:
//
//	type TMyArray = array[1..10] of Integer;  // static array
//	type TDynArray = array of String;         // dynamic array
//
// It resolves the element type, evaluates bounds for static arrays,
// creates the appropriate ArrayType, and registers it in the TypeSystem.
func (e *Evaluator) VisitArrayDecl(node *ast.ArrayDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil array declaration")
	}

	arrayName := node.Name.Value

	// Build the array type from the declaration
	arrayTypeAnnotation := node.ArrayType
	if arrayTypeAnnotation == nil {
		return e.newError(node, "invalid array type declaration")
	}

	// Resolve the element type using evaluator's type resolution
	elementTypeName := arrayTypeAnnotation.ElementType.String()
	elementType, err := e.resolveTypeName(elementTypeName, ctx)
	if err != nil {
		return e.newError(node, "unknown element type '%s': %v", elementTypeName, err)
	}

	// Create the array type
	var arrayType *types.ArrayType
	if arrayTypeAnnotation.IsDynamic() {
		arrayType = types.NewDynamicArrayType(elementType)
	} else {
		// Evaluate bound expressions using evaluator
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

	// Register array type in TypeSystem
	e.typeSystem.RegisterArrayType(arrayName, arrayType)

	return &runtime.NilValue{}
}

// VisitTypeDeclaration evaluates a type declaration.
// Task 3.5.16: Migrated from Interpreter.evalTypeDeclaration to Evaluator visitor.
//
// This method handles:
//  1. Subrange types: type TDigit = 0..9;
//  2. Function pointer types: type TCallback = procedure(x: Integer);
//  3. Type aliases: type TUserID = Integer;
//
// Inline/complex type expressions (ClassOfTypeNode, SetTypeNode, ArrayTypeNode,
// FunctionPointerTypeNode) are handled by the semantic analyzer and return NilValue.
func (e *Evaluator) VisitTypeDeclaration(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil type declaration")
	}

	// Handle subrange types
	if node.IsSubrange {
		return e.evalSubrangeType(node, ctx)
	}

	// Handle function pointer type declarations
	if node.IsFunctionPointer {
		return e.evalFunctionPointerType(node, ctx)
	}

	// Handle type aliases
	if node.IsAlias {
		return e.evalTypeAlias(node, ctx)
	}

	// Non-alias type declarations (future)
	return e.newError(node, "non-alias type declarations not yet supported")
}

// evalSubrangeType evaluates a subrange type declaration.
// Example: type TDigit = 0..9;
func (e *Evaluator) evalSubrangeType(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	// Evaluate low bound
	lowBoundVal := e.Eval(node.LowBound, ctx)
	if isError(lowBoundVal) {
		return lowBoundVal
	}
	lowBoundIntVal, ok := lowBoundVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "subrange low bound must be an integer")
	}
	lowBoundInt := int(lowBoundIntVal.Value)

	// Evaluate high bound
	highBoundVal := e.Eval(node.HighBound, ctx)
	if isError(highBoundVal) {
		return highBoundVal
	}
	highBoundIntVal, ok := highBoundVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "subrange high bound must be an integer")
	}
	highBoundInt := int(highBoundIntVal.Value)

	// Validate bounds
	if lowBoundInt > highBoundInt {
		return e.newError(node, "subrange low bound (%d) cannot be greater than high bound (%d)", lowBoundInt, highBoundInt)
	}

	// Create SubrangeType
	subrangeType := &types.SubrangeType{
		BaseType:  types.INTEGER,
		Name:      node.Name.Value,
		LowBound:  lowBoundInt,
		HighBound: highBoundInt,
	}

	// Register in TypeSystem
	e.typeSystem.RegisterSubrangeType(node.Name.Value, subrangeType)

	return &runtime.NilValue{}
}

// evalFunctionPointerType evaluates a function pointer type declaration.
// Example: type TCallback = procedure(x: Integer);
func (e *Evaluator) evalFunctionPointerType(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	if node.FunctionPointerType == nil {
		return e.newError(node, "function pointer type declaration has no type information")
	}

	// Store the type name mapping for type resolution
	// We just need to register that this type name exists
	// The actual type checking is done by the semantic analyzer
	typeKey := "__funcptr_type_" + node.Name.Value
	// Store a simple marker that this is a function pointer type
	ctx.Env().Define(typeKey, &runtime.StringValue{Value: "function_pointer_type"})

	return &runtime.NilValue{}
}

// evalTypeAlias evaluates a type alias declaration.
// Example: type TUserID = Integer;
func (e *Evaluator) evalTypeAlias(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	// Check for inline/complex type expressions
	// For inline types, we don't need to resolve them at runtime because:
	// 1. The semantic analyzer already validated the types during analysis phase
	// 2. Inline type aliases are purely semantic constructs (no runtime storage needed)
	// 3. The interpreter's resolveType() doesn't support complex inline syntax
	switch node.AliasedType.(type) {
	case *ast.ClassOfTypeNode:
		// Metaclass types (class of TBase) - semantic analyzer handles them
		return &runtime.NilValue{}
	case *ast.SetTypeNode:
		// Set types (set of TEnum) - semantic analyzer handles them
		return &runtime.NilValue{}
	case *ast.ArrayTypeNode:
		// Inline array types (array of Integer, array[1..10] of String)
		// Note: These could potentially need runtime storage, but semantic analyzer
		// already validated and stored them.
		return &runtime.NilValue{}
	case *ast.FunctionPointerTypeNode:
		// Function pointer types - already handled earlier in this function
		return &runtime.NilValue{}
	}

	// For TypeAnnotation with InlineType, also skip runtime resolution
	if typeAnnot, ok := node.AliasedType.(*ast.TypeAnnotation); ok && typeAnnot.InlineType != nil {
		// TypeAnnotation wrapping an inline type expression
		return &runtime.NilValue{}
	}

	// Resolve the aliased type by name (handles simple named types only)
	aliasedType, err := e.resolveTypeName(node.AliasedType.String(), ctx)
	if err != nil {
		return e.newError(node, "unknown type '%s' in type alias", node.AliasedType.String())
	}

	// Create TypeAliasValue and register it
	typeAlias := &runtime.TypeAliasValue{
		Name:        node.Name.Value,
		AliasedType: aliasedType,
	}

	// Store in environment with special prefix (case-insensitive)
	typeKey := "__type_alias_" + ident.Normalize(node.Name.Value)
	ctx.Env().Define(typeKey, typeAlias)

	return &runtime.NilValue{}
}

// VisitSetDecl evaluates a set declaration.
func (e *Evaluator) VisitSetDecl(node *ast.SetDecl, ctx *ExecutionContext) Value {
	// Set type already registered by semantic analyzer
	// Delegate to adapter for now (Phase 3 migration)
	return e.adapter.EvalNode(node)
}

// extractSimpleTypeName extracts the simple type name from a possibly qualified type string.
// Task 3.5.12: Used for helper registration with simplified type names.
// Example: "array of Integer" -> "array"
func extractSimpleTypeName(typeName string) string {
	if idx := strings.Index(typeName, " "); idx != -1 {
		return typeName[:idx]
	}
	return typeName
}
