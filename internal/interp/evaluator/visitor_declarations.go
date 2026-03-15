package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Visitor methods for declaration AST nodes (functions, classes, interfaces, types, etc.).

// VisitFunctionDecl evaluates a function declaration.
// Global functions are registered in TypeSystem.FunctionRegistry.
// Method implementations (node.ClassName != nil) are delegated to adapter.
func (e *Evaluator) VisitFunctionDecl(node *ast.FunctionDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil function declaration")
	}

	// Method implementations require ClassInfo internals (VMT rebuild, etc.)
	if node.ClassName != nil {
		typeName := node.ClassName.Value

		if classInfoAny := e.typeSystem.LookupClass(typeName); classInfoAny != nil {
			if classInfo, ok := classInfoAny.(classDeclarationInfo); ok {
				classInfo.RegisterMethodImplementation(node, e.typeSystem.AllClasses())
				return &runtime.NilValue{}
			}
			return e.newError(node, "type '%s' not found for method '%s'", typeName, node.Name.Value)
		}

		if recordInfoAny := e.typeSystem.LookupRecord(typeName); recordInfoAny != nil {
			if recordInfo, ok := recordInfoAny.(*runtime.RecordTypeValue); ok {
				recordInfo.RegisterMethodImplementation(node)
				return &runtime.NilValue{}
			}
			return e.newError(node, "type '%s' not found for method '%s'", typeName, node.Name.Value)
		}

		return e.newError(node, "type '%s' not found for method '%s'", typeName, node.Name.Value)
	}

	// Register global function in the canonical function registry.
	e.typeSystem.RegisterFunctionOrReplace(node.Name.Value, node)

	return &runtime.NilValue{}
}

// Returns fully qualified class name (e.g., "Outer.Inner" for nested classes).
func (e *Evaluator) fullClassNameFromDecl(cd *ast.ClassDecl) string {
	if cd.EnclosingClass != nil && cd.EnclosingClass.Value != "" {
		return cd.EnclosingClass.Value + "." + cd.Name.Value
	}
	return cd.Name.Value
}

type classDeclarationInfo interface {
	IsPartialClass() bool
	SetPartialClass(isPartial bool)
	SetAbstractClass(isAbstract bool)
	SetExternalClass(isExternal bool, externalName string)
	DefineCurrentClassMarker(env *runtime.Environment)
	HasNoParentClass() bool
	SetParentClass(parent any)
	AddImplementedInterface(iface any, ifaceName string)
	AddConstantValue(constDecl *ast.ConstDecl, value Value)
	ConstantValuesCopy() map[string]Value
	InheritConstantValuesFrom(parent any)
	AddNestedClassRef(nestedName string, nestedClass any)
	AddFieldDeclaration(fieldDecl *ast.FieldDecl, fieldType types.Type)
	AddClassVarValue(name string, value Value)
	AddMethodDeclaration(method *ast.FunctionDecl, className string, registry *runtime.MethodRegistry) bool
	LookupDeclaredMethod(methodName string, isClassMethod bool) (*ast.FunctionDecl, bool)
	SetConstructorDecl(constructor *ast.FunctionDecl)
	SetDestructorDecl(destructor *ast.FunctionDecl)
	InheritDestructorMetadataIfMissing()
	SynthesizeImplicitDefaultConstructor()
	SetPropertyInfo(name string, propInfo *types.PropertyInfo)
	DeterminePropertyAccessKind(specName string) types.PropAccessKind
	InheritParentPropertyInfos()
	RegisterOperatorBinding(operatorSymbol, bindingName string, operandTypes []string) error
	BuildVirtualMethodTableDirect()
	RegisterInTypeSystem(ts any, parentName string)
	DefineInEnv(env *runtime.Environment)
	RegisterMethodImplementation(fn *ast.FunctionDecl, allClasses map[string]interptypes.ClassInfo)
}

// VisitClassDecl evaluates a class declaration.
// Handles: partial classes, inheritance, interfaces, methods, properties, operators, VMT.
func (e *Evaluator) VisitClassDecl(node *ast.ClassDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil class declaration")
	}

	className := e.fullClassNameFromDecl(node)

	// Handle partial class merging or create new class
	var classInfo classDeclarationInfo
	existingClass := e.typeSystem.LookupClass(className)

	if existingClass != nil {
		ci, ok := existingClass.(classDeclarationInfo)
		if !ok {
			return e.newError(node, "internal error: invalid class type for '%s'", className)
		}

		existingPartial := ci.IsPartialClass()
		if existingPartial && node.IsPartial {
			classInfo = ci
		} else {
			return e.newError(node, "class '%s' already declared", className)
		}
	} else {
		rawClassInfo, err := e.typeSystem.NewClassInfo(className)
		if err != nil {
			return e.newError(node, "internal error: %s", err.Error())
		}
		ci, ok := rawClassInfo.(classDeclarationInfo)
		if !ok {
			return e.newError(node, "internal error: invalid class type for '%s'", className)
		}
		classInfo = ci
	}

	// Set class flags
	if node.IsPartial {
		classInfo.SetPartialClass(true)
	}
	if node.IsAbstract {
		classInfo.SetAbstractClass(true)
	}
	if node.IsExternal {
		classInfo.SetExternalClass(true, node.ExternalName)
	}

	// Setup temporary environment for nested class context
	tempEnv := runtime.NewEnclosedEnvironment(ctx.Env())
	classInfo.DefineCurrentClassMarker(tempEnv)
	savedEnv := ctx.Env()
	ctx.SetEnv(tempEnv)
	defer ctx.SetEnv(savedEnv)

	// Resolve inheritance (explicit parent or implicit TObject)
	var parentClass interface{}
	var parentClassName string
	if node.Parent != nil {
		parentClassName = node.Parent.Value
		parentClass = e.typeSystem.LookupClass(parentClassName)
		if parentClass == nil {
			return e.newError(node, "parent class '%s' not found", parentClassName)
		}
	} else {
		// Implicit TObject inheritance (unless this IS TObject or external)
		if !ident.Equal(className, "TObject") && !node.IsExternal {
			parentClassName = "TObject"
			parentClass = e.typeSystem.LookupClass(parentClassName)
			if parentClass == nil {
				return e.newError(node, "implicit parent class 'TObject' not found")
			}
		}
	}

	// Set parent reference and inherit members
	if parentClass != nil && classInfo.HasNoParentClass() {
		classInfo.SetParentClass(parentClass)
	}

	// Process implemented interfaces
	for _, ifaceIdent := range node.Interfaces {
		ifaceName := ifaceIdent.Value
		iface := e.typeSystem.LookupInterface(ifaceName)
		if iface == nil {
			return e.newError(node, "interface '%s' not found", ifaceName)
		}

		classInfo.AddImplementedInterface(iface, ifaceName)
	}

	// Evaluate class constants (sequentially to allow dependencies)
	classConstValues := classInfo.ConstantValuesCopy()
	if classConstValues == nil {
		classConstValues = make(map[string]Value)
	}

	evalWithClassConsts := func(expr ast.Expression) Value {
		savedEnv := ctx.Env()
		tempEnv := runtime.NewEnclosedEnvironment(savedEnv)
		for name, val := range classConstValues {
			tempEnv.Define(name, val)
		}
		ctx.SetEnv(tempEnv)
		defer ctx.SetEnv(savedEnv)
		return e.Eval(expr, ctx)
	}

	for _, constDecl := range node.Constants {
		if constDecl == nil {
			continue
		}
		constVal := evalWithClassConsts(constDecl.Value)
		if isError(constVal) {
			return constVal
		}
		classConstValues[constDecl.Name.Value] = constVal
		classInfo.AddConstantValue(constDecl, constVal)
	}

	// Inherit parent constants for field initializers
	if parentClass != nil {
		classInfo.InheritConstantValuesFrom(parentClass)
	}

	// Refresh constants after inheritance
	classConstValues = classInfo.ConstantValuesCopy()
	if classConstValues == nil {
		classConstValues = make(map[string]Value)
	}

	// Process nested types before fields so they can be referenced
	for _, nested := range node.NestedTypes {
		switch n := nested.(type) {
		case *ast.ClassDecl:
			if n.EnclosingClass == nil {
				n.EnclosingClass = &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{Token: node.Name.Token},
					},
					Value: className,
				}
			}
			if result := e.Eval(n, ctx); isError(result) {
				return result
			}
			if nestedInfo := e.typeSystem.LookupClass(e.fullClassNameFromDecl(n)); nestedInfo != nil {
				classInfo.AddNestedClassRef(n.Name.Value, nestedInfo)
			}
		default:
			if result := e.Eval(n, ctx); isError(result) {
				return result
			}
		}
	}

	// Process fields and class variables
	for _, field := range node.Fields {
		fieldName := field.Name.Value
		var fieldType types.Type
		var cachedInit Value

		switch {
		case field.Type != nil:
			var err error
			typeName := field.Type.String()
			fieldType, err = e.resolveTypeName(typeName, ctx)
			if err != nil || fieldType == nil {
				return e.newError(node, "unknown or invalid type for field '%s' in class '%s'", fieldName, className)
			}
		case field.InitValue != nil:
			initVal := evalWithClassConsts(field.InitValue)
			if isError(initVal) {
				return initVal
			}
			cachedInit = initVal
			fieldType = e.getValueType(initVal)
			if fieldType == nil {
				return e.newError(node, "cannot infer type for field '%s' in class '%s'", fieldName, className)
			}
		default:
			return e.newError(node, "field '%s' in class '%s' must have either a type or initializer", fieldName, className)
		}

		if field.IsClassVar {
			var varValue Value
			if cachedInit != nil {
				varValue = cachedInit
			} else if field.InitValue != nil {
				varValue = evalWithClassConsts(field.InitValue)
				if isError(varValue) {
					return varValue
				}
			} else {
				varValue = e.GetDefaultValue(fieldType)
			}

			classInfo.AddClassVarValue(fieldName, varValue)
			continue
		}

		classInfo.AddFieldDeclaration(field, fieldType)
	}

	// Add methods (overrides parent methods if same name, supports overloading)
	for _, method := range node.Methods {
		if !classInfo.AddMethodDeclaration(method, className, e.EngineState().MethodRegistry) {
			return e.newError(method, "failed to add method '%s' to class '%s'", method.Name.Value, className)
		}
	}

	// Process explicit constructor
	if node.Constructor != nil {
		if !classInfo.AddMethodDeclaration(node.Constructor, className, e.EngineState().MethodRegistry) {
			return e.newError(node.Constructor, "failed to add constructor to class '%s'", className)
		}
	}

	// Identify constructor ("Create") and destructor ("Destroy")
	if constructor, exists := classInfo.LookupDeclaredMethod("create", false); exists {
		classInfo.SetConstructorDecl(constructor)
	}
	if destructor, exists := classInfo.LookupDeclaredMethod("destroy", false); exists {
		classInfo.SetDestructorDecl(destructor)
	}

	// Inherit destructor from parent if missing
	classInfo.InheritDestructorMetadataIfMissing()

	// Synthesize default constructor if any constructor uses 'overload'
	classInfo.SynthesizeImplicitDefaultConstructor()

	// Register properties (after fields/methods so they can reference them)
	for _, propDecl := range node.Properties {
		if propDecl == nil {
			continue
		}
		propInfo := e.convertPropertyDecl(classInfo, propDecl, ctx)
		if propInfo == nil {
			return e.newError(propDecl, "failed to add property '%s' to class '%s'", propDecl.Name.Value, className)
		}
		classInfo.SetPropertyInfo(propDecl.Name.Value, propInfo)
	}

	// Inherit parent properties
	classInfo.InheritParentPropertyInfos()

	// Register class operators
	for _, opDecl := range node.Operators {
		if opDecl == nil {
			continue
		}
		if opDecl.Binding == nil {
			return e.newError(opDecl, "class operator '%s' missing binding", opDecl.OperatorSymbol)
		}

		operandTypes := make([]string, 0, len(opDecl.OperandTypes))
		for _, operand := range opDecl.OperandTypes {
			typeName := operand.String()
			resolvedType, err := e.resolveTypeName(typeName, ctx)
			if err == nil && resolvedType != nil {
				operandTypes = append(operandTypes, resolvedType.String())
			} else {
				operandTypes = append(operandTypes, typeName)
			}
		}

		if err := classInfo.RegisterOperatorBinding(opDecl.OperatorSymbol, opDecl.Binding.Value, operandTypes); err != nil {
			return e.newError(opDecl, "%s", err.Error())
		}
	}

	// Build VMT and register in TypeSystem
	classInfo.BuildVirtualMethodTableDirect()
	classInfo.RegisterInTypeSystem(e.typeSystem, parentClassName)
	classInfo.DefineInEnv(savedEnv)

	return &runtime.NilValue{}
}

// VisitInterfaceDecl evaluates an interface declaration.
// Creates InterfaceInfo, resolves parent, registers methods/properties.
func (e *Evaluator) VisitInterfaceDecl(node *ast.InterfaceDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil interface declaration")
	}

	interfaceName := node.Name.Value
	interfaceInfo := runtime.NewMutableInterfaceInfo(interfaceName)

	// Resolve parent interface
	if node.Parent != nil {
		parentName := node.Parent.Value
		parentInterface := e.typeSystem.LookupInterface(parentName)

		if parentInterface == nil {
			return e.newError(node.Parent, "parent interface '%s' not found", parentName)
		}

		parentInfo, ok := parentInterface.(*runtime.MutableInterfaceInfo)
		if !ok {
			return e.newError(node.Parent, "invalid parent interface '%s'", parentName)
		}

		interfaceInfo.Parent = parentInfo

		if e.hasCircularInterfaceInheritance(interfaceInfo) {
			return e.newError(node.Parent,
				"circular inheritance detected in interface '%s'", node.Name.Value)
		}
	}

	// Convert interface methods to FunctionDecl (no body)
	for _, methodDecl := range node.Methods {
		funcDecl := &ast.FunctionDecl{
			BaseNode: ast.BaseNode{
				Token: methodDecl.Token,
			},
			Name:       methodDecl.Name,
			Parameters: methodDecl.Parameters,
			ReturnType: methodDecl.ReturnType,
			Body:       nil,
		}

		interfaceInfo.Methods[ident.Normalize(methodDecl.Name.Value)] = funcDecl
	}

	// Register interface properties
	for _, propDecl := range node.Properties {
		if propDecl == nil {
			continue
		}
		propInfo := e.convertPropertyDecl(nil, propDecl, ctx)
		if propInfo != nil {
			interfaceInfo.Properties[ident.Normalize(propDecl.Name.Value)] = propInfo
		}
	}

	e.typeSystem.RegisterInterface(interfaceName, interfaceInfo)

	return &runtime.NilValue{}
}

// Converts AST property declaration to PropertyInfo for runtime access.
// Used by interface, class, and record evaluation.
func (e *Evaluator) convertPropertyDecl(classInfo classDeclarationInfo, propDecl *ast.PropertyDecl, ctx *ExecutionContext) *types.PropertyInfo {
	propType, err := e.ResolveTypeWithContext(propDecl.Type.String(), ctx)
	if err != nil || propType == nil {
		propType = types.NIL
	}

	propInfo := &types.PropertyInfo{
		Name:            propDecl.Name.Value,
		Type:            propType,
		IsIndexed:       len(propDecl.IndexParams) > 0,
		IsDefault:       propDecl.IsDefault,
		IsClassProperty: propDecl.IsClassProperty,
	}

	// Extract index value if present
	if propDecl.IndexValue != nil {
		if val, ok := ast.ExtractIntegerLiteral(propDecl.IndexValue); ok {
			propInfo.HasIndexValue = true
			propInfo.IndexValue = val
			propInfo.IndexValueType = types.INTEGER
		}
	}

	// Determine read access (field, method, or expression)
	if propDecl.ReadSpec != nil {
		if ident, ok := propDecl.ReadSpec.(*ast.Identifier); ok {
			propInfo.ReadSpec = ident.Value
			if classInfo == nil {
				propInfo.ReadKind = types.PropAccessMethod
			} else {
				// Match interpreter behavior: read specs on classes are treated as
				// field-like first, with runtime fallback to constant/class-var/method.
				propInfo.ReadKind = types.PropAccessField
			}
		} else {
			propInfo.ReadKind = types.PropAccessExpression
			propInfo.ReadSpec = propDecl.ReadSpec.String()
			propInfo.ReadExpr = propDecl.ReadSpec
		}
	} else {
		propInfo.ReadKind = types.PropAccessNone
	}

	if propDecl.WriteSpec != nil {
		if ident, ok := propDecl.WriteSpec.(*ast.Identifier); ok {
			propInfo.WriteSpec = ident.Value
			propInfo.WriteKind = e.determinePropertyAccessKind(classInfo, ident.Value)
		} else {
			propInfo.WriteKind = types.PropAccessNone
		}
	} else {
		propInfo.WriteKind = types.PropAccessNone
	}

	return propInfo
}

func (e *Evaluator) determinePropertyAccessKind(classInfo classDeclarationInfo, specName string) types.PropAccessKind {
	if classInfo == nil {
		return types.PropAccessMethod
	}
	return classInfo.DeterminePropertyAccessKind(specName)
}

// Checks for circular interface inheritance to prevent infinite loops.
func (e *Evaluator) hasCircularInterfaceInheritance(iface *runtime.MutableInterfaceInfo) bool {
	seen := make(map[string]bool)
	current := iface

	for current != nil {
		normalizedName := ident.Normalize(current.Name)
		if seen[normalizedName] {
			return true
		}
		seen[normalizedName] = true
		current = current.Parent
	}

	return false
}

// VisitOperatorDecl evaluates operator overloading declarations.
// Handles global operators and conversion operators.
// Class operators are handled during class declaration evaluation.
func (e *Evaluator) VisitOperatorDecl(node *ast.OperatorDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil operator declaration")
	}

	// Class operators handled during class declaration
	if node.Kind == ast.OperatorKindClass {
		return &runtime.NilValue{}
	}

	if node.Binding == nil {
		return e.newError(node, "operator '%s' missing binding", node.OperatorSymbol)
	}

	// Normalize operand types
	operandTypes := make([]string, len(node.OperandTypes))
	for idx, operand := range node.OperandTypes {
		opRand := operand.String()
		operandTypes[idx] = interptypes.NormalizeTypeAnnotation(opRand)
	}

	// Handle conversion operators (implicit/explicit)
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

	// Register global operator
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
// Calculates ordinal values, validates flags (powers of 2), registers in TypeSystem.
func (e *Evaluator) VisitEnumDecl(node *ast.EnumDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil enum declaration")
	}

	enumName := node.Name.Value
	enumValues := make(map[string]int)
	orderedNames := make([]string, 0, len(node.Values))

	// Calculate ordinal values
	currentOrdinal := 0
	flagBitPosition := 0

	for _, enumValue := range node.Values {
		valueName := enumValue.Name
		var ordinalValue int

		if enumValue.ValueExpr != nil {
			val := e.Eval(enumValue.ValueExpr, ctx)
			if isError(val) {
				return e.newError(node, "enum '%s' value '%s': %v", enumName, valueName, val)
			}

			var errVal Value
			ordinalValue, errVal = e.extractEnumOrdinal(val, enumName, valueName, node, ctx)
			if errVal != nil {
				return errVal
			}

			if node.Flags {
				// Validate power of 2 for flags
				if ordinalValue <= 0 || (ordinalValue&(ordinalValue-1)) != 0 {
					return e.newError(node, "enum '%s' value '%s' (%d) must be a power of 2 for flags enum",
						enumName, valueName, ordinalValue)
				}
				// Update bit position
				for bitPos := 0; bitPos < 64; bitPos++ {
					if (1 << bitPos) == ordinalValue {
						flagBitPosition = bitPos + 1
						break
					}
				}
			} else {
				currentOrdinal = ordinalValue + 1
			}
		} else if enumValue.Value != nil {
			ordinalValue = *enumValue.Value
			if node.Flags {
				// Validate power of 2 for flags
				if ordinalValue <= 0 || (ordinalValue&(ordinalValue-1)) != 0 {
					return e.newError(node, "enum '%s' value '%s' (%d) must be a power of 2 for flags enum",
						enumName, valueName, ordinalValue)
				}
				// Update bit position
				for bitPos := 0; bitPos < 64; bitPos++ {
					if (1 << bitPos) == ordinalValue {
						flagBitPosition = bitPos + 1
						break
					}
				}
			} else {
				currentOrdinal = ordinalValue + 1
			}
		} else {
			// Implicit value
			if node.Flags {
				ordinalValue = 1 << flagBitPosition
				flagBitPosition++
			} else {
				ordinalValue = currentOrdinal
				currentOrdinal++
			}
		}

		enumValues[valueName] = ordinalValue
		orderedNames = append(orderedNames, valueName)
	}

	// Create enum type
	var enumType *types.EnumType
	if node.Scoped || node.Flags {
		enumType = types.NewScopedEnumType(enumName, enumValues, orderedNames, node.Flags)
	} else {
		enumType = types.NewEnumType(enumName, enumValues, orderedNames)
	}

	// Register enum values in environment (skip for scoped enums)
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

	// Register enum type metadata
	enumTypeKey := "__enum_type_" + ident.Normalize(enumName)
	enumTypeValue := runtime.NewEnumTypeValue(enumType)
	ctx.Env().Define(enumTypeKey, enumTypeValue)
	e.typeSystem.RegisterEnumType(enumName, enumTypeValue)

	// Register type name as TypeMetaValue for High(TColor), Low(TColor), etc.
	typeMetaValue := &runtime.TypeMetaValue{
		TypeInfo: enumType,
		TypeName: enumName,
	}
	ctx.Env().Define(enumName, typeMetaValue)

	return &runtime.NilValue{}
}

// extractEnumOrdinal coerces a value produced by an enum ValueExpr into an ordinal integer.
func (e *Evaluator) extractEnumOrdinal(val Value, enumName, valueName string, node ast.Node, ctx *ExecutionContext) (int, Value) {
	val = e.UnwrapVariant(val)

	switch v := val.(type) {
	case *runtime.IntegerValue:
		return int(v.Value), nil
	case *runtime.EnumValue:
		return v.OrdinalValue, nil
	case *runtime.BooleanValue:
		if v.Value {
			return 1, nil
		}
		return 0, nil
	case *runtime.StringValue:
		runes := []rune(v.Value)
		if len(runes) == 0 {
			return 0, e.newError(node, "enum '%s' value '%s': empty string not valid for enum value", enumName, valueName)
		}
		if len(runes) > 1 {
			return 0, e.newError(node, "enum '%s' value '%s': string '%s' must be a single character", enumName, valueName, v.Value)
		}
		return int(runes[0]), nil
	case *runtime.SubrangeValue:
		return v.GetValue(), nil
	}

	if converted, ok := e.TryImplicitConversion(val, "Integer", ctx); ok {
		if isError(converted) {
			return 0, converted
		}
		return e.extractEnumOrdinal(converted, enumName, valueName, node, ctx)
	}

	return 0, e.newError(node, "enum '%s' value '%s': value of type %s is not an ordinal type", enumName, valueName, val.Type())
}

// VisitRecordDecl evaluates a record declaration.
// Handles fields, methods, constants, class vars, and properties.
func (e *Evaluator) VisitRecordDecl(node *ast.RecordDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil record declaration")
	}

	recordName := node.Name.Value
	fields := make(map[string]types.Type)
	fieldDecls := make(map[string]*ast.FieldDecl)

	// Process fields with type inference support
	for _, field := range node.Fields {
		fieldName := field.Name.Value
		var fieldType types.Type

		if field.Type != nil {
			var err error
			typeName := field.Type.String()
			fieldType, err = e.resolveTypeName(typeName, ctx)
			if err != nil || fieldType == nil {
				return e.newError(node, "unknown or invalid type for field '%s' in record '%s'", fieldName, recordName)
			}
		} else if field.InitValue != nil {
			initValue := e.Eval(field.InitValue, ctx)
			if isError(initValue) {
				return initValue
			}
			fieldType = e.getValueType(initValue)
			if fieldType == nil {
				return e.newError(node, "cannot infer type for field '%s' in record '%s'", fieldName, recordName)
			}
		} else {
			return e.newError(node, "field '%s' in record '%s' must have either a type or initializer", fieldName, recordName)
		}

		fields[fieldName] = fieldType
		fieldDecls[ident.Normalize(fieldName)] = field
	}

	recordType := types.NewRecordType(recordName, fields)

	// Separate instance and static methods
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

	// Setup temporary environment for constant evaluation (to allow dependencies)
	savedEnv := ctx.Env()
	tempEnv := runtime.NewEnclosedEnvironment(savedEnv)
	ctx.SetEnv(tempEnv)
	defer ctx.SetEnv(savedEnv)

	// Evaluate record constants
	constants := make(map[string]Value)
	for _, constant := range node.Constants {
		constName := constant.Name.Value
		constValue := e.Eval(constant.Value, ctx)
		if isError(constValue) {
			return constValue
		}
		// Define in temp env so subsequent constants/class vars can see it
		tempEnv.Define(constName, constValue)
		constants[ident.Normalize(constName)] = constValue
	}

	// Initialize class variables
	classVars := make(map[string]Value)
	for _, classVar := range node.ClassVars {
		varName := classVar.Name.Value
		var varValue Value

		if classVar.InitValue != nil {
			varValue = e.Eval(classVar.InitValue, ctx)
			if isError(varValue) {
				return varValue
			}
		} else {
			var varType types.Type
			if classVar.Type != nil {
				var err error
				typeName := classVar.Type.String()
				varType, err = e.resolveTypeName(typeName, ctx)
				if err != nil || varType == nil {
					return e.newError(node, "unknown type for class variable '%s' in record '%s'", varName, recordName)
				}
			}
			varValue = e.GetDefaultValue(varType)
		}

		classVars[ident.Normalize(varName)] = varValue
	}

	// Process properties
	for _, prop := range node.Properties {
		propName := prop.Name.Value
		propNameLower := ident.Normalize(propName)

		typeName := prop.Type.String()
		propType, err := e.resolveTypeName(typeName, ctx)
		if err != nil || propType == nil {
			return e.newError(node, "unknown type for property '%s' in record '%s'", propName, recordName)
		}

		propInfo := &types.RecordPropertyInfo{
			Name:       propName,
			Type:       propType,
			ReadField:  prop.ReadField,
			WriteField: prop.WriteField,
			IsDefault:  prop.IsDefault,
		}

		recordType.Properties[propNameLower] = propInfo
	}

	// Build metadata and create record type value
	metadata := e.buildRecordMetadata(recordName, recordType, methods, staticMethods, constants, classVars)

	recordTypeKey := "__record_type_" + ident.Normalize(recordName)
	recordTypeValue := &RecordTypeValue{
		RecordType:           recordType,
		FieldDecls:           fieldDecls,
		Metadata:             metadata,
		Methods:              methods,
		StaticMethods:        staticMethods,
		ClassMethods:         make(map[string]*ast.FunctionDecl),
		ClassMethodOverloads: make(map[string][]*ast.FunctionDecl),
		MethodOverloads:      make(map[string][]*ast.FunctionDecl),
		Constants:            constants,
		ClassVars:            classVars,
	}

	// Initialize ClassMethods with StaticMethods
	for k, v := range staticMethods {
		recordTypeValue.ClassMethods[k] = v
	}

	// Initialize overload lists
	for methodName, methodDecl := range methods {
		recordTypeValue.MethodOverloads[methodName] = []*ast.FunctionDecl{methodDecl}
	}
	for methodName, methodDecl := range staticMethods {
		recordTypeValue.ClassMethodOverloads[methodName] = []*ast.FunctionDecl{methodDecl}
	}

	// Register in environment and TypeSystem
	// Use savedEnv because ctx.Env() is currently tempEnv which will be discarded
	// Register record type under two keys:
	// 1. Internal key (recordTypeKey = "__record_type_" + normalized name) for legacy/compatibility lookup
	// 2. Plain record name for direct user access to the type
	savedEnv.Define(recordTypeKey, recordTypeValue)
	savedEnv.Define(recordName, recordTypeValue)
	e.typeSystem.RegisterRecord(recordName, recordTypeValue)

	return &runtime.NilValue{}
}

// VisitHelperDecl evaluates a helper declaration (type extension).
// Handles helper/record helper, parent inheritance, methods/properties.
func (e *Evaluator) VisitHelperDecl(node *ast.HelperDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return &runtime.NilValue{}
	}

	// Resolve target type
	targetType, err := e.ResolveTypeFromAnnotation(node.ForType)
	if err != nil {
		return e.newError(node, "unknown target type '%s' for helper '%s'",
			node.ForType.String(), node.Name.Value)
	}

	helperInfo := runtime.NewMutableHelperInfo(node.Name.Value, targetType, node.IsRecordHelper)

	// Resolve parent helper
	if node.ParentHelper != nil {
		parentHelperName := node.ParentHelper.Value

		var foundParent *runtime.MutableHelperInfo
		for _, helpers := range e.typeSystem.AllHelpers() {
			for _, helper := range helpers {
				if helperInfoAny, ok := helper.(*runtime.MutableHelperInfo); ok && ident.Equal(helperInfoAny.Name, parentHelperName) {
					foundParent = helperInfoAny
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

		if !foundParent.TargetType.Equals(targetType) {
			return e.newError(node.ParentHelper,
				"parent helper '%s' extends different type than child helper '%s'",
				parentHelperName, node.Name.Value)
		}

		helperInfo.ParentHelper = foundParent
	}

	// Register methods
	for _, method := range node.Methods {
		helperInfo.Methods[ident.Normalize(method.Name.Value)] = method
	}

	// Register properties
	for _, prop := range node.Properties {
		propType, propErr := e.ResolveTypeFromAnnotation(prop.Type)
		if propErr != nil {
			return e.newError(prop, "unknown type '%s' for property '%s'",
				prop.Type.String(), prop.Name.Value)
		}
		propInfo := &types.PropertyInfo{Name: prop.Name.Value, Type: propType}
		if prop.ReadSpec != nil {
			if identExpr, ok := prop.ReadSpec.(*ast.Identifier); ok {
				propInfo.ReadKind = types.PropAccessMethod
				propInfo.ReadSpec = identExpr.Value
			}
		}
		if prop.WriteSpec != nil {
			if identExpr, ok := prop.WriteSpec.(*ast.Identifier); ok {
				propInfo.WriteKind = types.PropAccessMethod
				propInfo.WriteSpec = identExpr.Value
			}
		}
		helperInfo.Properties[prop.Name.Value] = propInfo
	}

	// Initialize class variables
	for _, classVar := range node.ClassVars {
		var (
			varType      types.Type
			initialValue runtime.Value
		)

		if classVar.Type != nil {
			resolvedType, err := e.resolveTypeName(classVar.Type.String(), ctx)
			if err != nil {
				return e.newError(classVar, "unknown type for class variable '%s'",
					classVar.Name.Value)
			}
			varType = resolvedType
		}

		if classVar.InitValue != nil {
			val := e.Eval(classVar.InitValue, ctx)
			if isError(val) {
				return val
			}
			initialValue = val

			// Infer type if not explicit
			if varType == nil {
				switch v := val.(type) {
				case *runtime.IntegerValue:
					varType = types.INTEGER
				case *runtime.FloatValue:
					varType = types.FLOAT
				case *runtime.StringValue:
					varType = types.STRING
				case *runtime.BooleanValue:
					varType = types.BOOLEAN
				case *runtime.ArrayValue:
					varType = v.ArrayType
				}
				if varType == nil {
					return e.newError(classVar, "cannot infer type for class variable '%s'",
						classVar.Name.Value)
				}
			}
		}

		if varType == nil {
			return e.newError(classVar, "unknown type for class variable '%s'",
				classVar.Name.Value)
		}

		if initialValue == nil {
			initialValue = e.GetDefaultValue(varType)
		}

		helperInfo.ClassVars[ident.Normalize(classVar.Name.Value)] = initialValue
	}

	// Evaluate class constants
	for _, classConst := range node.ClassConsts {
		constValue := e.Eval(classConst.Value, ctx)
		if isError(constValue) {
			return constValue
		}
		helperInfo.ClassConsts[ident.Normalize(classConst.Name.Value)] = constValue
	}

	// Register helper in TypeSystem
	typeName := ident.Normalize(targetType.String())
	e.typeSystem.RegisterHelper(typeName, helperInfo)

	// Also register by simple type name
	simpleTypeName := ident.Normalize(extractSimpleTypeName(targetType.String()))
	if simpleTypeName != typeName {
		e.typeSystem.RegisterHelper(simpleTypeName, helperInfo)
	}

	// Expose helper name for static access (THelper.Member)
	ctx.Env().Define(node.Name.Value, &runtime.TypeMetaValue{
		TypeInfo: targetType,
		TypeName: targetType.String(),
	})

	return &runtime.NilValue{}
}

// VisitArrayDecl evaluates an array type declaration.
// Handles dynamic (array of T) and static (array[N..M] of T) arrays.
func (e *Evaluator) VisitArrayDecl(node *ast.ArrayDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil array declaration")
	}

	arrayName := node.Name.Value
	arrayTypeAnnotation := node.ArrayType
	if arrayTypeAnnotation == nil {
		return e.newError(node, "invalid array type declaration")
	}

	// Resolve element type
	elementTypeName := arrayTypeAnnotation.ElementType.String()
	elementType, err := e.resolveTypeName(elementTypeName, ctx)
	if err != nil {
		return e.newError(node, "unknown element type '%s': %v", elementTypeName, err)
	}

	// Create array type (dynamic or static)
	var arrayType *types.ArrayType
	if arrayTypeAnnotation.IsDynamic() {
		arrayType = types.NewDynamicArrayType(elementType)
	} else {
		// Evaluate bounds
		lowBoundVal := e.Eval(arrayTypeAnnotation.LowBound, ctx)
		if isError(lowBoundVal) {
			return lowBoundVal
		}
		highBoundVal := e.Eval(arrayTypeAnnotation.HighBound, ctx)
		if isError(highBoundVal) {
			return highBoundVal
		}

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

	e.typeSystem.RegisterArrayType(arrayName, arrayType)

	return &runtime.NilValue{}
}

// VisitTypeDeclaration evaluates a type declaration.
// Handles subrange types, function pointers, and type aliases.
func (e *Evaluator) VisitTypeDeclaration(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil type declaration")
	}

	if node.IsSubrange {
		return e.evalSubrangeType(node, ctx)
	}
	if node.IsFunctionPointer {
		return e.evalFunctionPointerType(node, ctx)
	}
	if node.IsAlias {
		return e.evalTypeAlias(node, ctx)
	}

	return e.newError(node, "non-alias type declarations not yet supported")
}

// Evaluates subrange type (type TDigit = 0..9).
func (e *Evaluator) evalSubrangeType(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	// Evaluate bounds
	lowBoundVal := e.Eval(node.LowBound, ctx)
	if isError(lowBoundVal) {
		return lowBoundVal
	}
	lowBoundIntVal, ok := lowBoundVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "subrange low bound must be an integer")
	}
	lowBoundInt := int(lowBoundIntVal.Value)

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

	// Create and register subrange type
	subrangeType := &types.SubrangeType{
		BaseType:  types.INTEGER,
		Name:      node.Name.Value,
		LowBound:  lowBoundInt,
		HighBound: highBoundInt,
	}

	e.typeSystem.RegisterSubrangeType(node.Name.Value, subrangeType)

	return &runtime.NilValue{}
}

// Evaluates function pointer type (type TCallback = procedure(x: Integer)).
func (e *Evaluator) evalFunctionPointerType(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	if node.FunctionPointerType == nil {
		return e.newError(node, "function pointer type declaration has no type information")
	}

	funcPtrType := node.FunctionPointerType

	// Resolve parameter types
	paramTypes := make([]types.Type, len(funcPtrType.Parameters))
	for idx, param := range funcPtrType.Parameters {
		var paramType types.Type
		if param.Type != nil {
			var err error
			paramType, err = e.ResolveTypeFromAnnotation(param.Type)
			if err != nil || paramType == nil {
				return e.newError(node, "unknown parameter type '%s' in function pointer '%s'", param.Type.String(), node.Name.Value)
			}
		} else {
			paramType = types.INTEGER
		}
		paramTypes[idx] = paramType
	}

	// Resolve return type (nil for procedures)
	var returnType types.Type
	if funcPtrType.ReturnType != nil {
		var err error
		returnType, err = e.ResolveTypeFromAnnotation(funcPtrType.ReturnType)
		if err != nil {
			return e.newError(node, "unknown return type '%s' in function pointer '%s'", funcPtrType.ReturnType.String(), node.Name.Value)
		}
	}

	// Create function or method pointer type
	var resolvedType types.Type
	if funcPtrType.OfObject {
		resolvedType = types.NewMethodPointerType(paramTypes, returnType)
	} else {
		resolvedType = types.NewFunctionPointerType(paramTypes, returnType)
	}

	// Register in TypeSystem
	if e.typeSystem != nil {
		e.typeSystem.RegisterFunctionPointerType(node.Name.Value, resolvedType)
	}

	// Legacy marker
	typeKey := "__funcptr_type_" + node.Name.Value
	ctx.Env().Define(typeKey, &runtime.StringValue{Value: "function_pointer_type"})

	return &runtime.NilValue{}
}

// Evaluates type alias (type TUserID = Integer).
func (e *Evaluator) evalTypeAlias(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	// Skip inline/complex types (handled by semantic analyzer)
	var (
		aliasedType types.Type
		resolveErr  error
	)

	switch t := node.AliasedType.(type) {
	case *ast.ClassOfTypeNode:
		baseClassName := ""
		if t.ClassType != nil {
			baseClassName = t.ClassType.String()
		}
		if baseClassName == "" || !e.typeSystem.HasClass(baseClassName) {
			return e.newError(node, "unknown type '%s' in type alias", baseClassName)
		}
		classType := types.NewClassType(baseClassName, nil)
		aliasedType = types.NewClassOfType(classType)
	case *ast.SetTypeNode:
		// Set types handled by semantic analyzer.
		return &runtime.NilValue{}
	case *ast.ArrayTypeNode:
		resolvedArray := e.resolveArrayTypeNode(t, ctx)
		if resolvedArray == nil {
			return e.newError(node, "cannot resolve array type in alias '%s'", node.Name.Value)
		}
		aliasedType = resolvedArray
	case *ast.FunctionPointerTypeNode:
		// Function pointer types handled elsewhere.
		return &runtime.NilValue{}
	default:
		if typeAnnot, ok := node.AliasedType.(*ast.TypeAnnotation); ok && typeAnnot.InlineType != nil {
			return &runtime.NilValue{}
		}

		aliasedType, resolveErr = e.resolveTypeName(node.AliasedType.String(), ctx)
		if resolveErr != nil {
			return e.newError(node, "unknown type '%s' in type alias", node.AliasedType.String())
		}
	}

	// Create and register type alias.
	typeAlias := &runtime.TypeAliasValue{
		Name:        node.Name.Value,
		AliasedType: aliasedType,
	}

	// If alias targets an enum, register the alias for scoped enum lookups.
	if enumType, ok := aliasedType.(*types.EnumType); ok {
		e.typeSystem.RegisterEnumType(node.Name.Value, runtime.NewEnumTypeValue(enumType))
	}

	typeKey := "__type_alias_" + ident.Normalize(node.Name.Value)
	ctx.Env().Define(typeKey, typeAlias)
	ctx.Env().Define(node.Name.Value, &runtime.TypeMetaValue{
		TypeInfo: aliasedType,
		TypeName: node.Name.Value,
	})

	return &runtime.NilValue{}
}

// VisitSetDecl evaluates a set declaration.
func (e *Evaluator) VisitSetDecl(node *ast.SetDecl, ctx *ExecutionContext) Value {
	if node == nil || node.Name == nil {
		return e.newError(node, "invalid set declaration")
	}
	if node.ElementType == nil {
		return e.newError(node, "set '%s' missing element type", node.Name.Value)
	}

	// Register named set types in the runtime environment so later phases (var init,
	// empty set literals, etc.) can resolve them.
	elemType, err := e.ResolveTypeWithContext(node.ElementType.String(), ctx)
	if err != nil {
		return e.newError(node, "unknown set element type '%s'", node.ElementType.String())
	}

	setType := types.NewSetType(elemType)
	ctx.Env().Define("__set_type_"+ident.Normalize(node.Name.Value), &runtime.SetTypeValue{
		Name:    node.Name.Value,
		SetType: setType,
	})
	return nil
}

// Extracts simple type name from qualified string ("array of Integer" -> "array").
func extractSimpleTypeName(typeName string) string {
	if idx := strings.Index(typeName, " "); idx != -1 {
		return typeName[:idx]
	}
	return typeName
}
