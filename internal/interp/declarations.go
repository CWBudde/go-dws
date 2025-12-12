package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

func (i *Interpreter) fullClassNameFromDecl(cd *ast.ClassDecl) string {
	if cd.EnclosingClass != nil && cd.EnclosingClass.Value != "" {
		return cd.EnclosingClass.Value + "." + cd.Name.Value
	}
	return cd.Name.Value
}

// evalFunctionDeclaration registers a function in the registry without executing it.
// For methods (fn.ClassName != nil), updates the class/record method maps.
func (i *Interpreter) evalFunctionDeclaration(fn *ast.FunctionDecl) Value {
	// Handle method implementation (e.g., TExample.Method)
	if fn.ClassName != nil {
		typeName := fn.ClassName.Value

		classInfo, isClass := i.classes[ident.Normalize(typeName)]
		if isClass {
			i.evalClassMethodImplementation(fn, classInfo)
			return &NilValue{}
		}

		recordInfo, isRecord := i.records[ident.Normalize(typeName)]
		if isRecord {
			i.evalRecordMethodImplementation(fn, recordInfo)
			return &NilValue{}
		}

		return i.newErrorWithLocation(fn, "type '%s' not found for method '%s'", typeName, fn.Name.Value)
	}

	// Register global function in TypeSystem and legacy map
	i.typeSystem.RegisterFunctionOrReplace(fn.Name.Value, fn)
	funcName := ident.Normalize(fn.Name.Value)

	// Implementations (with body) replace forward declarations; declarations are appended
	if fn.Body != nil {
		existingOverloads := i.functions[funcName]
		i.functions[funcName] = i.replaceMethodInOverloadList(existingOverloads, fn)
	} else {
		i.functions[funcName] = append(i.functions[funcName], fn)
	}

	return &NilValue{}
}

// evalClassMethodImplementation registers a class method implementation, replacing any declaration.
func (i *Interpreter) evalClassMethodImplementation(fn *ast.FunctionDecl, classInfo *ClassInfo) {
	normalizedMethodName := ident.Normalize(fn.Name.Value)

	// Replace declaration with implementation in method maps and overload lists
	if fn.IsClassMethod {
		classInfo.ClassMethods[normalizedMethodName] = fn
		overloads := classInfo.ClassMethodOverloads[normalizedMethodName]
		classInfo.ClassMethodOverloads[normalizedMethodName] = i.replaceMethodInOverloadList(overloads, fn)
	} else {
		classInfo.Methods[normalizedMethodName] = fn
		overloads := classInfo.MethodOverloads[normalizedMethodName]
		classInfo.MethodOverloads[normalizedMethodName] = i.replaceMethodInOverloadList(overloads, fn)
	}

	// Store constructors and destructors
	if fn.IsConstructor {
		normalizedCtorName := ident.Normalize(fn.Name.Value)
		classInfo.Constructors[normalizedCtorName] = fn
		overloads := classInfo.ConstructorOverloads[normalizedCtorName]
		classInfo.ConstructorOverloads[normalizedCtorName] = i.replaceMethodInOverloadList(overloads, fn)
		classInfo.Constructor = fn
	}

	if fn.IsDestructor {
		classInfo.Destructor = fn
	}

	// Rebuild VMT and propagate to descendants
	classInfo.buildVirtualMethodTable()
	i.propagateMethodImplementationToDescendants(classInfo, normalizedMethodName, fn, fn.IsClassMethod)
	i.rebuildDescendantVMTs(classInfo)
}

// rebuildDescendantVMTs rebuilds VMTs for all descendant classes to pick up parent method changes.
func (i *Interpreter) rebuildDescendantVMTs(parentClass *ClassInfo) {
	for _, classInfo := range i.classes {
		if i.isDescendantOf(classInfo, parentClass) {
			classInfo.buildVirtualMethodTable()
		}
	}
}

// isDescendantOf returns true if childClass inherits from ancestorClass (directly or indirectly).
func (i *Interpreter) isDescendantOf(childClass, ancestorClass *ClassInfo) bool {
	current := childClass.Parent
	for current != nil {
		if current == ancestorClass {
			return true
		}
		current = current.Parent
	}
	return false
}

// propagateMethodImplementationToDescendants updates descendant method maps to latest parent implementation
// (unless the descendant provides its own override).
func (i *Interpreter) propagateMethodImplementationToDescendants(parentClass *ClassInfo, normalizedMethodName string, fn *ast.FunctionDecl, isClassMethod bool) {
	for _, classInfo := range i.classes {
		if !i.isDescendantOf(classInfo, parentClass) {
			continue
		}

		if isClassMethod {
			if existing, ok := classInfo.ClassMethods[normalizedMethodName]; ok {
				// Skip if descendant overrides the method
				if existing.ClassName != nil && ident.Equal(existing.ClassName.Value, classInfo.Name) {
					continue
				}
				classInfo.ClassMethods[normalizedMethodName] = fn
			}
		} else {
			if existing, ok := classInfo.Methods[normalizedMethodName]; ok {
				// Skip if descendant overrides the method
				if existing.ClassName != nil && ident.Equal(existing.ClassName.Value, classInfo.Name) {
					continue
				}
				classInfo.Methods[normalizedMethodName] = fn
			}
		}
	}
}

// evalRecordMethodImplementation registers a record method implementation, replacing any declaration.
func (i *Interpreter) evalRecordMethodImplementation(fn *ast.FunctionDecl, recordInfo *RecordTypeValue) {
	normalizedMethodName := ident.Normalize(fn.Name.Value)
	methodMeta := runtime.MethodMetadataFromAST(fn)

	if fn.IsClassMethod {
		// Static method
		recordInfo.ClassMethods[normalizedMethodName] = fn
		overloads := recordInfo.ClassMethodOverloads[normalizedMethodName]
		recordInfo.ClassMethodOverloads[normalizedMethodName] = i.replaceMethodInOverloadList(overloads, fn)

		// Update metadata
		if recordInfo.Metadata != nil {
			recordInfo.Metadata.StaticMethods[normalizedMethodName] = methodMeta
			recordInfo.Metadata.StaticMethodOverloads[normalizedMethodName] = i.replaceMethodMetadataInOverloadList(
				recordInfo.Metadata.StaticMethodOverloads[normalizedMethodName],
				methodMeta,
			)
		}
	} else {
		// Instance method
		recordInfo.Methods[normalizedMethodName] = fn
		overloads := recordInfo.MethodOverloads[normalizedMethodName]
		recordInfo.MethodOverloads[normalizedMethodName] = i.replaceMethodInOverloadList(overloads, fn)

		// Update metadata
		if recordInfo.Metadata != nil {
			recordInfo.Metadata.Methods[normalizedMethodName] = methodMeta
			recordInfo.Metadata.MethodOverloads[normalizedMethodName] = i.replaceMethodMetadataInOverloadList(
				recordInfo.Metadata.MethodOverloads[normalizedMethodName],
				methodMeta,
			)
		}
	}
}

// evalClassDeclaration builds ClassInfo from AST, registers it, and handles inheritance.
func (i *Interpreter) evalClassDeclaration(cd *ast.ClassDecl) Value {
	className := i.fullClassNameFromDecl(cd)

	// Handle partial class merging
	var classInfo *ClassInfo
	existingClass, exists := i.classes[ident.Normalize(className)]

	switch {
	case exists && existingClass.IsPartial && cd.IsPartial:
		classInfo = existingClass
	case exists:
		return i.newErrorWithLocation(cd, "class '%s' already declared", className)
	default:
		classInfo = NewClassInfo(className)
	}

	// Set flags
	if cd.IsPartial {
		classInfo.IsPartial = true
		classInfo.Metadata.IsPartial = true
	}

	if cd.IsAbstract {
		classInfo.IsAbstractFlag = true
		classInfo.Metadata.IsAbstract = true
	}

	if cd.IsExternal {
		classInfo.IsExternalFlag = true
		classInfo.ExternalName = cd.ExternalName
		classInfo.Metadata.IsExternal = true
		classInfo.Metadata.ExternalName = cd.ExternalName
	}

	// Provide current class context for nested type resolution
	defer i.PushScope()()
	i.Env().Define("__CurrentClass__", &ClassInfoValue{ClassInfo: classInfo})

	// Resolve parent class (explicit or implicit TObject)
	var parentClass *ClassInfo
	if cd.Parent != nil {
		var exists bool
		parentClass, exists = i.classes[ident.Normalize(cd.Parent.Value)]
		if !exists {
			return i.newErrorWithLocation(cd, "parent class '%s' not found", cd.Parent.Value)
		}
	} else if !ident.Equal(className, "TObject") && !cd.IsExternal {
		var exists bool
		parentClass, exists = i.classes[ident.Normalize("TObject")]
		if !exists {
			return i.newErrorWithLocation(cd, "implicit parent class 'TObject' not found")
		}
	}

	// Inherit from parent (if not already set for partial classes)
	if parentClass != nil && classInfo.Parent == nil {
		classInfo.Parent = parentClass
		classInfo.Metadata.Parent = parentClass.Metadata
		classInfo.Metadata.ParentName = parentClass.Metadata.Name

		// Copy fields
		for fieldName, fieldType := range parentClass.Fields {
			classInfo.Fields[fieldName] = fieldType
		}
		for fieldName, fieldDecl := range parentClass.FieldDecls {
			classInfo.FieldDecls[fieldName] = fieldDecl
		}

		// Copy methods (direct lookups only; overloads handled separately)
		for methodName, methodDecl := range parentClass.Methods {
			classInfo.Methods[methodName] = methodDecl
		}
		for methodName, methodDecl := range parentClass.ClassMethods {
			classInfo.ClassMethods[methodName] = methodDecl
		}

		// Copy constructors
		for name, constructor := range parentClass.Constructors {
			normalizedName := ident.Normalize(name)
			classInfo.Constructors[normalizedName] = constructor
		}
		for name, overloads := range parentClass.ConstructorOverloads {
			normalizedName := ident.Normalize(name)
			classInfo.ConstructorOverloads[normalizedName] = append([]*ast.FunctionDecl(nil), overloads...)
		}

		if parentClass.DefaultConstructor != "" {
			classInfo.DefaultConstructor = parentClass.DefaultConstructor
		}

		// Copy constructor metadata
		for name, constructor := range parentClass.Metadata.Constructors {
			if classInfo.Metadata.Constructors == nil {
				classInfo.Metadata.Constructors = make(map[string]*runtime.MethodMetadata)
			}
			classInfo.Metadata.Constructors[name] = constructor
		}
		for name, overloads := range parentClass.Metadata.ConstructorOverloads {
			if classInfo.Metadata.ConstructorOverloads == nil {
				classInfo.Metadata.ConstructorOverloads = make(map[string][]*runtime.MethodMetadata)
			}
			classInfo.Metadata.ConstructorOverloads[name] = append([]*runtime.MethodMetadata(nil), overloads...)
		}

		if parentClass.Metadata.DefaultConstructor != "" {
			classInfo.Metadata.DefaultConstructor = parentClass.Metadata.DefaultConstructor
		}

		classInfo.Operators = parentClass.Operators.clone()
	}

	// Process implemented interfaces
	for _, ifaceIdent := range cd.Interfaces {
		iface := i.lookupInterfaceInfo(ifaceIdent.Value)
		if iface == nil {
			return i.newErrorWithLocation(cd, "interface '%s' not found", ifaceIdent.Value)
		}

		classInfo.Interfaces = append(classInfo.Interfaces, iface)
		classInfo.Metadata.Interfaces = append(classInfo.Metadata.Interfaces, ifaceIdent.Value)
	}

	// Process class constants (evaluated eagerly so later constants can reference earlier ones)
	for _, constDecl := range cd.Constants {
		if constDecl == nil {
			continue
		}
		classInfo.Constants[constDecl.Name.Value] = constDecl

		// Evaluate with previously defined constants in scope
		// Phase 3.1.4: unified scope management
		var constValue Value
		func() {
			defer i.PushScope()()
			for cName, cValue := range classInfo.ConstantValues {
				i.Env().Define(cName, cValue)
			}

			constValue = i.Eval(constDecl.Value)
		}()

		if isError(constValue) {
			return constValue
		}

		classInfo.ConstantValues[constDecl.Name.Value] = constValue
	}

	// Inherit parent constants
	if classInfo.Parent != nil {
		for constName, constDecl := range classInfo.Parent.Constants {
			if _, exists := classInfo.Constants[constName]; !exists {
				classInfo.Constants[constName] = constDecl
			}
		}
		for constName, constValue := range classInfo.Parent.ConstantValues {
			if _, exists := classInfo.ConstantValues[constName]; !exists {
				classInfo.ConstantValues[constName] = constValue
			}
		}
	}

	// Register class early for field initializers that may reference the class
	i.classes[ident.Normalize(classInfo.Name)] = classInfo

	// Process nested types before fields so they can be referenced
	for _, nested := range cd.NestedTypes {
		switch n := nested.(type) {
		case *ast.ClassDecl:
			if n.EnclosingClass == nil {
				n.EnclosingClass = &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: cd.Name.Token,
						},
					},
					Value: classInfo.Name,
				}
			}
			if result := i.evalClassDeclaration(n); isError(result) {
				return result
			}
			if nestedInfo, ok := i.classes[ident.Normalize(i.fullClassNameFromDecl(n))]; ok {
				classInfo.NestedClasses[ident.Normalize(n.Name.Value)] = nestedInfo
			}
		default:
			if result := i.Eval(n); isError(result) {
				return result
			}
		}
	}

	// Process fields
	for _, field := range cd.Fields {
		var fieldType types.Type
		var cachedInitValue Value

		// Resolve or infer field type
		switch {
		case field.Type != nil:
			fieldType = i.resolveTypeFromExpression(field.Type)
			if fieldType == nil {
				return i.newErrorWithLocation(field, "unknown or invalid type for field '%s'", field.Name.Value)
			}
		case field.InitValue != nil:
			// Infer type from init value
			var initVal Value
			func() {
				defer i.PushScope()()
				for cName, cValue := range classInfo.ConstantValues {
					i.Env().Define(cName, cValue)
				}

				initVal = i.Eval(field.InitValue)
			}()

			if isError(initVal) {
				return initVal
			}
			cachedInitValue = initVal

			fieldType = i.inferTypeFromValue(initVal)
			if fieldType == nil {
				return i.newErrorWithLocation(field, "cannot infer type for field '%s'", field.Name.Value)
			}
		default:
			return i.newErrorWithLocation(field, "field '%s' has no type annotation", field.Name.Value)
		}

		// Handle class variables vs instance fields
		if field.IsClassVar {
			var classVarValue Value

			if field.InitValue != nil {
				// Reuse cached value to avoid double evaluation
				if cachedInitValue != nil {
					classVarValue = cachedInitValue
				} else {
					var val Value
					func() {
						defer i.PushScope()()
						for cName, cValue := range classInfo.ConstantValues {
							i.Env().Define(cName, cValue)
						}

						val = i.Eval(field.InitValue)
					}()

					if isError(val) {
						return val
					}
					classVarValue = val
				}
			} else {
				// Initialize with default value
				switch fieldType {
				case types.INTEGER:
					classVarValue = &IntegerValue{Value: 0}
				case types.FLOAT:
					classVarValue = &FloatValue{Value: 0.0}
				case types.STRING:
					classVarValue = &StringValue{Value: ""}
				case types.BOOLEAN:
					classVarValue = &BooleanValue{Value: false}
				default:
					classVarValue = &NilValue{}
				}
			}
			classInfo.ClassVars[field.Name.Value] = classVarValue
		} else {
			// Instance field
			classInfo.Fields[field.Name.Value] = fieldType
			classInfo.FieldDecls[field.Name.Value] = field

			fieldMeta := runtime.FieldMetadataFromAST(field)
			fieldMeta.Type = fieldType
			runtime.AddFieldToClass(classInfo.Metadata, fieldMeta)
		}
	}

	// Process methods
	for _, method := range cd.Methods {
		normalizedMethodName := ident.Normalize(method.Name.Value)

		// Auto-detect constructors (methods named "Create" that return the class type)
		if !method.IsConstructor && ident.Equal(method.Name.Value, "Create") && method.ReturnType != nil {
			returnTypeName := method.ReturnType.String()
			if ident.Equal(returnTypeName, cd.Name.Value) {
				method.IsConstructor = true
			}
		}

		// Register method metadata
		methodMeta := runtime.MethodMetadataFromAST(method)
		i.methodRegistry.RegisterMethod(methodMeta)

		// Add to method maps and overload lists
		if method.IsClassMethod {
			classInfo.ClassMethods[normalizedMethodName] = method
			classInfo.ClassMethodOverloads[normalizedMethodName] = append(classInfo.ClassMethodOverloads[normalizedMethodName], method)

			if !method.IsConstructor && !method.IsDestructor {
				runtime.AddMethodToClass(classInfo.Metadata, methodMeta, true)
			}
		} else {
			classInfo.Methods[normalizedMethodName] = method
			classInfo.MethodOverloads[normalizedMethodName] = append(classInfo.MethodOverloads[normalizedMethodName], method)

			if !method.IsConstructor && !method.IsDestructor {
				runtime.AddMethodToClass(classInfo.Metadata, methodMeta, false)
			}
		}

		if method.IsDestructor {
			classInfo.Metadata.Destructor = methodMeta
		}

		if method.IsConstructor {
			normalizedName := ident.Normalize(method.Name.Value)
			classInfo.Constructors[normalizedName] = method
			runtime.AddConstructorToClass(classInfo.Metadata, methodMeta)

			if method.IsDefault {
				classInfo.DefaultConstructor = method.Name.Value
			}

			// Child constructors hide parent constructors with same signature
			existingOverloads := classInfo.ConstructorOverloads[normalizedName]
			replaced := false
			for i, existingMethod := range existingOverloads {
				if parametersMatch(existingMethod.Parameters, method.Parameters) {
					existingOverloads[i] = method
					replaced = true
					break
				}
			}
			if !replaced {
				existingOverloads = append(existingOverloads, method)
			}
			classInfo.ConstructorOverloads[normalizedName] = existingOverloads
		}
	}

	// Identify constructor and destructor
	if constructor, exists := classInfo.Methods["create"]; exists {
		classInfo.Constructor = constructor
	}

	if cd.Constructor != nil {
		normalizedName := ident.Normalize(cd.Constructor.Name.Value)
		classInfo.Constructors[normalizedName] = cd.Constructor

		constructorMeta := runtime.MethodMetadataFromAST(cd.Constructor)
		i.methodRegistry.RegisterMethod(constructorMeta)
		runtime.AddConstructorToClass(classInfo.Metadata, constructorMeta)

		// Child constructors hide parent constructors with same signature
		existingOverloads := classInfo.ConstructorOverloads[normalizedName]
		replaced := false
		for i, existingMethod := range existingOverloads {
			if parametersMatch(existingMethod.Parameters, cd.Constructor.Parameters) {
				existingOverloads[i] = cd.Constructor
				replaced = true
				break
			}
		}
		if !replaced {
			existingOverloads = append(existingOverloads, cd.Constructor)
		}
		classInfo.ConstructorOverloads[normalizedName] = existingOverloads
	}

	if destructor, exists := classInfo.Methods["destroy"]; exists {
		classInfo.Destructor = destructor
	}

	// Inherit destructor from parent if not declared
	if classInfo.Metadata.Destructor == nil && classInfo.Parent != nil && classInfo.Parent.Metadata.Destructor != nil {
		classInfo.Metadata.Destructor = classInfo.Parent.Metadata.Destructor
	}

	// Synthesize implicit parameterless constructor if needed
	i.synthesizeImplicitParameterlessConstructor(classInfo)

	// Process properties
	for _, propDecl := range cd.Properties {
		if propDecl == nil {
			continue
		}

		propInfo := i.convertPropertyDecl(propDecl)
		if propInfo != nil {
			classInfo.Properties[propDecl.Name.Value] = propInfo
		}
	}

	// Inherit parent properties
	if classInfo.Parent != nil {
		for propName, propInfo := range classInfo.Parent.Properties {
			if _, exists := classInfo.Properties[propName]; !exists {
				classInfo.Properties[propName] = propInfo
			}
		}
	}

	// Register class operators
	for _, opDecl := range cd.Operators {
		if opDecl == nil {
			continue
		}
		if errVal := i.registerClassOperator(classInfo, opDecl); isError(errVal) {
			return errVal
		}
	}

	// Build VMT and register in TypeSystem
	classInfo.buildVirtualMethodTable()

	parentName2 := ""
	if classInfo.Parent != nil {
		parentName2 = classInfo.Parent.Name
	}
	i.typeSystem.RegisterClassWithParent(classInfo.Name, classInfo, parentName2)

	return &NilValue{}
}

// convertPropertyDecl converts AST property to PropertyInfo for runtime access.
func (i *Interpreter) convertPropertyDecl(propDecl *ast.PropertyDecl) *types.PropertyInfo {
	// Resolve property type
	var propType types.Type
	switch propDecl.Type.String() {
	case "Integer":
		propType = types.INTEGER
	case "Float":
		propType = types.FLOAT
	case "String":
		propType = types.STRING
	case "Boolean":
		propType = types.BOOLEAN
	default:
		if classInfo := i.resolveClassInfoByName(propDecl.Type.String()); classInfo != nil {
			propType = types.NewClassType(classInfo.Name, nil)
		} else {
			propType = types.NIL
		}
	}

	propInfo := &types.PropertyInfo{
		Name:            propDecl.Name.Value,
		Type:            propType,
		IsIndexed:       len(propDecl.IndexParams) > 0,
		IsDefault:       propDecl.IsDefault,
		IsClassProperty: propDecl.IsClassProperty,
	}

	if propDecl.IndexValue != nil {
		if val, ok := ast.ExtractIntegerLiteral(propDecl.IndexValue); ok {
			propInfo.HasIndexValue = true
			propInfo.IndexValue = val
			propInfo.IndexValueType = types.INTEGER
		}
	}

	// Configure read access
	if propDecl.ReadSpec != nil {
		if ident, ok := propDecl.ReadSpec.(*ast.Identifier); ok {
			propInfo.ReadSpec = ident.Value
			propInfo.ReadKind = types.PropAccessField
		} else {
			propInfo.ReadKind = types.PropAccessExpression
			propInfo.ReadSpec = propDecl.ReadSpec.String()
			propInfo.ReadExpr = propDecl.ReadSpec
		}
	} else {
		propInfo.ReadKind = types.PropAccessNone
	}

	// Configure write access
	if propDecl.WriteSpec != nil {
		if ident, ok := propDecl.WriteSpec.(*ast.Identifier); ok {
			propInfo.WriteSpec = ident.Value
			propInfo.WriteKind = types.PropAccessField
		} else {
			propInfo.WriteKind = types.PropAccessNone
		}
	} else {
		propInfo.WriteKind = types.PropAccessNone
	}

	return propInfo
}

// evalInterfaceDeclaration builds InterfaceInfo from AST and registers it.
func (i *Interpreter) evalInterfaceDeclaration(id *ast.InterfaceDecl) Value {
	interfaceInfo := NewInterfaceInfo(id.Name.Value)

	// Handle interface inheritance
	if id.Parent != nil {
		parentInterface := i.lookupInterfaceInfo(id.Parent.Value)
		if parentInterface == nil {
			return i.newErrorWithLocation(id, "parent interface '%s' not found", id.Parent.Value)
		}
		interfaceInfo.Parent = parentInterface
	}

	// Convert interface methods to FunctionDecl (interface methods have no body)
	for _, methodDecl := range id.Methods {
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
	for _, propDecl := range id.Properties {
		if propDecl == nil {
			continue
		}
		if propInfo := i.convertPropertyDecl(propDecl); propInfo != nil {
			interfaceInfo.Properties[ident.Normalize(propDecl.Name.Value)] = propInfo
		}
	}

	i.typeSystem.RegisterInterface(interfaceInfo.Name, interfaceInfo)

	return &NilValue{}
}

// synthesizeImplicitParameterlessConstructor adds a parameterless constructor when
// at least one constructor has 'overload' but no parameterless version exists.
func (i *Interpreter) synthesizeImplicitParameterlessConstructor(classInfo *ClassInfo) {
	for ctorName, overloads := range classInfo.ConstructorOverloads {
		hasOverloadDirective := false
		hasParameterlessOverload := false

		for _, ctor := range overloads {
			if ctor.IsOverload {
				hasOverloadDirective = true
			}
			if len(ctor.Parameters) == 0 {
				hasParameterlessOverload = true
			}
		}

		// Synthesize parameterless constructor if needed
		if hasOverloadDirective && !hasParameterlessOverload {
			implicitConstructor := &ast.FunctionDecl{
				BaseNode:      ast.BaseNode{},
				Name:          &ast.Identifier{Value: ctorName},
				Parameters:    []*ast.Parameter{},
				ReturnType:    nil,
				Body:          nil,
				IsConstructor: true,
				IsOverload:    true,
			}

			normalizedName := ident.Normalize(ctorName)
			if _, exists := classInfo.Constructors[normalizedName]; !exists {
				classInfo.Constructors[normalizedName] = implicitConstructor
			}
			classInfo.ConstructorOverloads[normalizedName] = append(
				classInfo.ConstructorOverloads[normalizedName],
				implicitConstructor,
			)
		}
	}
}

func (i *Interpreter) evalOperatorDeclaration(decl *ast.OperatorDecl) Value {
	if decl.Kind == ast.OperatorKindClass {
		return &NilValue{}
	}

	if decl.Binding == nil {
		return i.newErrorWithLocation(decl, "operator '%s' missing binding", decl.OperatorSymbol)
	}

	operandTypes := make([]string, len(decl.OperandTypes))
	for idx, operand := range decl.OperandTypes {
		operandTypes[idx] = NormalizeTypeAnnotation(operand.String())
	}

	// Handle conversion operators
	if decl.Kind == ast.OperatorKindConversion {
		if len(operandTypes) != 1 {
			return i.newErrorWithLocation(decl, "conversion operator '%s' requires exactly one operand", decl.OperatorSymbol)
		}
		if decl.ReturnType == nil {
			return i.newErrorWithLocation(decl, "conversion operator '%s' requires a return type", decl.OperatorSymbol)
		}
		targetType := NormalizeTypeAnnotation(decl.ReturnType.String())
		entry := &interptypes.ConversionEntry{
			From:        operandTypes[0],
			To:          targetType,
			BindingName: ident.Normalize(decl.Binding.Value),
			Implicit:    ident.Equal(decl.OperatorSymbol, "implicit"),
		}
		if err := i.typeSystem.Conversions().Register(entry); err != nil {
			return i.newErrorWithLocation(decl, "conversion from %s to %s already defined", operandTypes[0], targetType)
		}
		return &NilValue{}
	}

	// Register global operator
	entry := &runtimeOperatorEntry{
		Operator:      decl.OperatorSymbol,
		OperandTypes:  operandTypes,
		BindingName:   ident.Normalize(decl.Binding.Value),
		Class:         nil,
		SelfIndex:     -1,
		IsClassMethod: false,
	}

	// Convert to types.OperatorEntry for TypeSystem
	tsEntry := &interptypes.OperatorEntry{
		Class:         entry.Class,
		Operator:      entry.Operator,
		BindingName:   entry.BindingName,
		OperandTypes:  entry.OperandTypes,
		SelfIndex:     entry.SelfIndex,
		IsClassMethod: entry.IsClassMethod,
	}
	if err := i.typeSystem.Operators().Register(tsEntry); err != nil {
		return i.newErrorWithLocation(decl, "operator '%s' already defined for operand types (%s)", decl.OperatorSymbol, strings.Join(operandTypes, ", "))
	}

	return &NilValue{}
}

func (i *Interpreter) registerClassOperator(classInfo *ClassInfo, opDecl *ast.OperatorDecl) Value {
	if opDecl.Binding == nil {
		return i.newErrorWithLocation(opDecl, "class operator '%s' missing binding", opDecl.OperatorSymbol)
	}

	// Find binding method
	normalizedBindingName := ident.Normalize(opDecl.Binding.Value)
	method, isClassMethod := classInfo.ClassMethods[normalizedBindingName]
	if !isClassMethod {
		var ok bool
		method, ok = classInfo.Methods[normalizedBindingName]
		if !ok {
			return i.newErrorWithLocation(opDecl, "binding '%s' for class operator '%s' not found in class '%s'", opDecl.Binding.Value, opDecl.OperatorSymbol, classInfo.Name)
		}
	}

	// Build operand type list
	classKey := NormalizeTypeAnnotation(classInfo.Name)
	operandTypes := make([]string, 0, len(opDecl.OperandTypes)+1)
	includesClass := false
	for _, operand := range opDecl.OperandTypes {
		typeName := operand.String()
		resolvedType, err := i.resolveType(typeName)
		var key string
		if err == nil {
			key = NormalizeTypeAnnotation(resolvedType.String())
		} else {
			key = NormalizeTypeAnnotation(typeName)
		}
		if key == classKey {
			includesClass = true
		}
		operandTypes = append(operandTypes, key)
	}

	// Add class to operand types if not present
	if !includesClass {
		if ident.Equal(opDecl.OperatorSymbol, "in") {
			operandTypes = append(operandTypes, classKey)
		} else {
			operandTypes = append([]string{classKey}, operandTypes...)
		}
	}

	// Find self parameter index for instance methods
	selfIndex := -1
	if !isClassMethod {
		for idx, key := range operandTypes {
			if key == classKey {
				selfIndex = idx
				break
			}
		}
		if selfIndex == -1 {
			return i.newErrorWithLocation(opDecl, "unable to determine self operand for class operator '%s'", opDecl.OperatorSymbol)
		}
	}

	entry := &runtimeOperatorEntry{
		Operator:      opDecl.OperatorSymbol,
		OperandTypes:  operandTypes,
		BindingName:   normalizedBindingName,
		Class:         classInfo,
		IsClassMethod: isClassMethod,
		SelfIndex:     selfIndex,
	}

	if err := classInfo.Operators.register(entry); err != nil {
		return i.newErrorWithLocation(opDecl, "class operator '%s' already defined for operand types (%s)", opDecl.OperatorSymbol, strings.Join(operandTypes, ", "))
	}

	if method.IsConstructor {
		classInfo.Constructors[normalizedBindingName] = method
	}

	return &NilValue{}
}

// parametersMatch checks if two parameter lists have matching signatures.
func parametersMatch(params1, params2 []*ast.Parameter) bool {
	if len(params1) != len(params2) {
		return false
	}
	for i := range params1 {
		if params1[i].Type != nil && params2[i].Type != nil {
			if params1[i].Type.String() != params2[i].Type.String() {
				return false
			}
		} else if params1[i].Type != params2[i].Type {
			return false
		}
	}
	return true
}

// replaceMethodInOverloadList replaces a declaration with its implementation, or appends if not found.
func (i *Interpreter) replaceMethodInOverloadList(list []*ast.FunctionDecl, impl *ast.FunctionDecl) []*ast.FunctionDecl {
	for idx, decl := range list {
		if parametersMatch(decl.Parameters, impl.Parameters) {
			// Preserve flags from declaration
			impl.IsVirtual = decl.IsVirtual
			impl.IsOverride = decl.IsOverride
			impl.IsReintroduce = decl.IsReintroduce
			impl.IsAbstract = decl.IsAbstract

			list[idx] = impl
			return list
		}
	}
	return append(list, impl)
}

// replaceMethodMetadataInOverloadList replaces a declaration with its implementation.
func (i *Interpreter) replaceMethodMetadataInOverloadList(list []*runtime.MethodMetadata, impl *runtime.MethodMetadata) []*runtime.MethodMetadata {
	for idx, decl := range list {
		if methodMetadataSignatureMatch(decl, impl) {
			// Preserve flags from declaration
			impl.IsVirtual = decl.IsVirtual
			impl.IsOverride = decl.IsOverride
			impl.IsReintroduce = decl.IsReintroduce
			impl.IsAbstract = decl.IsAbstract
			impl.Visibility = decl.Visibility
			if impl.ReturnTypeName == "" {
				impl.ReturnTypeName = decl.ReturnTypeName
			}
			list[idx] = impl
			return list
		}
	}
	return append(list, impl)
}

// methodMetadataSignatureMatch checks if two MethodMetadata have matching signatures.
func methodMetadataSignatureMatch(a, b *runtime.MethodMetadata) bool {
	if a == nil || b == nil {
		return false
	}

	if !parameterMetadataMatch(a.Parameters, b.Parameters) {
		return false
	}

	if a.ReturnTypeName != "" && b.ReturnTypeName != "" && !ident.Equal(a.ReturnTypeName, b.ReturnTypeName) {
		return false
	}

	return true
}

// parameterMetadataMatch checks if two parameter lists match by type and ByRef flag.
func parameterMetadataMatch(params1, params2 []runtime.ParameterMetadata) bool {
	if len(params1) != len(params2) {
		return false
	}

	for i := range params1 {
		if params1[i].ByRef != params2[i].ByRef {
			return false
		}

		switch {
		case params1[i].TypeName != "" && params2[i].TypeName != "":
			if !ident.Equal(params1[i].TypeName, params2[i].TypeName) {
				return false
			}
		case params1[i].TypeName != params2[i].TypeName:
			return false
		}
	}

	return true
}

// inferTypeFromValue infers a type from a runtime value (for type inference).
func (i *Interpreter) inferTypeFromValue(val Value) types.Type {
	switch val := val.(type) {
	case *IntegerValue:
		return types.INTEGER
	case *FloatValue:
		return types.FLOAT
	case *StringValue:
		return types.STRING
	case *BooleanValue:
		return types.BOOLEAN
	case *ArrayValue:
		if len(val.Elements) > 0 {
			elemType := i.inferTypeFromValue(val.Elements[0])
			if elemType != nil {
				lowBound := 0
				highBound := len(val.Elements) - 1
				return &types.ArrayType{
					ElementType: elemType,
					LowBound:    &lowBound,
					HighBound:   &highBound,
				}
			}
		}
		return nil
	case *ObjectInstance, *NilValue:
		return nil
	default:
		return nil
	}
}
