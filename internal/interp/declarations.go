package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// evalFunctionDeclaration evaluates a function declaration.
// It registers the function in the function registry without executing its body.
// For method implementations (fn.ClassName != nil), it updates the class's Methods map.
func (i *Interpreter) evalFunctionDeclaration(fn *ast.FunctionDecl) Value {
	// Check if this is a method implementation (has a class name like TExample.Method)
	if fn.ClassName != nil {
		className := fn.ClassName.Value
		classInfo, exists := i.classes[className]
		if !exists {
			return i.newErrorWithLocation(fn, "class '%s' not found for method '%s'", className, fn.Name.Value)
		}

		// Update the method in the class (replacing the declaration with the implementation)
		// Task 9.67: Support method overloading by storing multiple methods per name
		// We need to replace the declaration with the implementation in the overload list
		if fn.IsClassMethod {
			classInfo.ClassMethods[fn.Name.Value] = fn
			// Replace declaration with implementation in overload list
			overloads := classInfo.ClassMethodOverloads[fn.Name.Value]
			classInfo.ClassMethodOverloads[fn.Name.Value] = i.replaceMethodInOverloadList(overloads, fn)
		} else {
			classInfo.Methods[fn.Name.Value] = fn
			// Replace declaration with implementation in overload list
			overloads := classInfo.MethodOverloads[fn.Name.Value]
			classInfo.MethodOverloads[fn.Name.Value] = i.replaceMethodInOverloadList(overloads, fn)
		}

		// Also store constructors
		if fn.IsConstructor {
			classInfo.Constructors[fn.Name.Value] = fn
			// Task 9.67: Replace declaration with implementation in constructor overload list
			overloads := classInfo.ConstructorOverloads[fn.Name.Value]
			classInfo.ConstructorOverloads[fn.Name.Value] = i.replaceMethodInOverloadList(overloads, fn)
			// Always update Constructor to use the implementation (which has the body)
			// This replaces the declaration that was set during class parsing
			classInfo.Constructor = fn
		}

		// Store destructor
		if fn.IsDestructor {
			classInfo.Destructor = fn
		}

		return &NilValue{}
	}

	// Store regular function in the registry
	// Task 9.66: Support overloading by storing multiple functions per name
	funcName := fn.Name.Value

	// Task 9.80: If this function has a body, it may be an implementation that should
	// replace a previous interface declaration (which has no body).
	// This happens when units have separate interface and implementation sections.
	if fn.Body != nil {
		// Replace any existing declaration without a body
		existingOverloads := i.functions[funcName]
		i.functions[funcName] = i.replaceMethodInOverloadList(existingOverloads, fn)
	} else {
		// Interface declaration - append it
		i.functions[funcName] = append(i.functions[funcName], fn)
	}

	return &NilValue{}
}

// evalClassDeclaration evaluates a class declaration.
// It builds a ClassInfo from the AST and registers it in the class registry.
// Handles inheritance by copying parent fields and methods to the child class.
func (i *Interpreter) evalClassDeclaration(cd *ast.ClassDecl) Value {
	// Create new ClassInfo
	classInfo := NewClassInfo(cd.Name.Value)

	// Set abstract flag
	classInfo.IsAbstract = cd.IsAbstract

	// Set external flags
	classInfo.IsExternal = cd.IsExternal
	classInfo.ExternalName = cd.ExternalName

	// Handle inheritance if parent class is specified
	if cd.Parent != nil {
		parentName := cd.Parent.Value
		parentClass, exists := i.classes[parentName]
		if !exists {
			return i.newErrorWithLocation(cd, "parent class '%s' not found", parentName)
		}

		// Set parent reference
		classInfo.Parent = parentClass

		// Copy parent fields (child inherits all parent fields)
		for fieldName, fieldType := range parentClass.Fields {
			classInfo.Fields[fieldName] = fieldType
		}

		// Copy parent methods (child inherits all parent methods)
		// Keep Methods and ClassMethods for backward compatibility (direct lookups)
		for methodName, methodDecl := range parentClass.Methods {
			classInfo.Methods[methodName] = methodDecl
		}
		for methodName, methodDecl := range parentClass.ClassMethods {
			classInfo.ClassMethods[methodName] = methodDecl
		}

		// Task 9.21.6: DON'T copy MethodOverloads/ClassMethodOverloads from parent
		// Each class should only store its OWN method overloads, not inherited ones.
		// getMethodOverloadsInHierarchy will walk the hierarchy to collect them at call time.
		// This prevents duplication when a child class overrides a parent method.

		// Copy constructors
		for name, constructor := range parentClass.Constructors {
			classInfo.Constructors[name] = constructor
		}
		for name, overloads := range parentClass.ConstructorOverloads {
			classInfo.ConstructorOverloads[name] = append([]*ast.FunctionDecl(nil), overloads...)
		}

		// Copy operator overloads
		classInfo.Operators = parentClass.Operators.clone()
	}

	// Add own fields to ClassInfo
	for _, field := range cd.Fields {
		// Get the field type from the type annotation
		if field.Type == nil {
			return i.newErrorWithLocation(field, "field '%s' has no type annotation", field.Name.Value)
		}

		// Resolve field type using type expression
		fieldType := i.resolveTypeFromExpression(field.Type)
		if fieldType == nil {
			return i.newErrorWithLocation(field, "unknown or invalid type for field '%s'", field.Name.Value)
		}

		// Check if this is a class variable (static field) or instance field
		if field.IsClassVar {
			// Initialize class variable with default value based on type
			var defaultValue Value
			switch fieldType {
			case types.INTEGER:
				defaultValue = &IntegerValue{Value: 0}
			case types.FLOAT:
				defaultValue = &FloatValue{Value: 0.0}
			case types.STRING:
				defaultValue = &StringValue{Value: ""}
			case types.BOOLEAN:
				defaultValue = &BooleanValue{Value: false}
			default:
				defaultValue = &NilValue{}
			}
			classInfo.ClassVars[field.Name.Value] = defaultValue
		} else {
			// Store instance field type in ClassInfo
			classInfo.Fields[field.Name.Value] = fieldType
		}
	}

	// Add own methods to ClassInfo (these override parent methods if same name)
	// Task 9.67: Support method overloading by storing multiple methods per name
	for _, method := range cd.Methods {
		// Check if this is a class method (static method) or instance method
		if method.IsClassMethod {
			// Store in ClassMethods map
			classInfo.ClassMethods[method.Name.Value] = method
			// Add to overload list
			classInfo.ClassMethodOverloads[method.Name.Value] = append(classInfo.ClassMethodOverloads[method.Name.Value], method)
		} else {
			// Store in instance Methods map
			classInfo.Methods[method.Name.Value] = method
			// Add to overload list
			classInfo.MethodOverloads[method.Name.Value] = append(classInfo.MethodOverloads[method.Name.Value], method)
		}

		if method.IsConstructor {
			classInfo.Constructors[method.Name.Value] = method
			// Task 9.67: Add to constructor overload list
			// Task 9.73.8: Handle constructor hiding/override
			// In DWScript, a child constructor with the same name and signature HIDES the parent's,
			// regardless of whether it has the `override` keyword or not
			existingOverloads := classInfo.ConstructorOverloads[method.Name.Value]
			replaced := false
			for i, existingMethod := range existingOverloads {
				// Check if signatures match (same number and types of parameters)
				if parametersMatch(existingMethod.Parameters, method.Parameters) {
					// Replace the parent constructor with this child constructor (hiding)
					existingOverloads[i] = method
					replaced = true
					break
				}
			}
			if !replaced {
				// No matching parent constructor found (different signature), just append
				existingOverloads = append(existingOverloads, method)
			}
			// Write the modified slice back to the map
			classInfo.ConstructorOverloads[method.Name.Value] = existingOverloads
		}
	}

	// Identify constructor (method named "Create")
	if constructor, exists := classInfo.Methods["Create"]; exists {
		classInfo.Constructor = constructor
	}
	if cd.Constructor != nil {
		classInfo.Constructors[cd.Constructor.Name.Value] = cd.Constructor
		// Task 9.20: Also add to ConstructorOverloads for consistency
		// Task 9.73.8: Handle constructor hiding/override
		// In DWScript, a child constructor with the same name and signature HIDES the parent's,
		// regardless of whether it has the `override` keyword or not
		existingOverloads := classInfo.ConstructorOverloads[cd.Constructor.Name.Value]
		replaced := false
		for i, existingMethod := range existingOverloads {
			// Check if signatures match (same number and types of parameters)
			if parametersMatch(existingMethod.Parameters, cd.Constructor.Parameters) {
				// Replace the parent constructor with this child constructor (hiding)
				existingOverloads[i] = cd.Constructor
				replaced = true
				break
			}
		}
		if !replaced {
			// No matching parent constructor found (different signature), just append
			existingOverloads = append(existingOverloads, cd.Constructor)
		}
		// Write the modified slice back to the map
		classInfo.ConstructorOverloads[cd.Constructor.Name.Value] = existingOverloads
	}

	// Identify destructor (method named "Destroy")
	if destructor, exists := classInfo.Methods["Destroy"]; exists {
		classInfo.Destructor = destructor
	}

	// Register properties
	// Properties are registered after fields and methods so they can reference them
	for _, propDecl := range cd.Properties {
		if propDecl == nil {
			continue
		}

		// Convert AST property to PropertyInfo
		propInfo := i.convertPropertyDecl(propDecl)
		if propInfo != nil {
			classInfo.Properties[propDecl.Name.Value] = propInfo
		}
	}

	// Copy parent properties (child inherits all parent properties)
	if classInfo.Parent != nil {
		for propName, propInfo := range classInfo.Parent.Properties {
			// Only copy if not already defined in child class
			if _, exists := classInfo.Properties[propName]; !exists {
				classInfo.Properties[propName] = propInfo
			}
		}
	}

	// Register class constants (Task 9.20-9.22)
	// Evaluate constants eagerly in order so they can reference earlier constants
	for _, constDecl := range cd.Constants {
		if constDecl == nil {
			continue
		}
		// Store the constant declaration
		classInfo.Constants[constDecl.Name.Value] = constDecl

		// Evaluate the constant value immediately
		// Create temporary environment with previously evaluated constants
		savedEnv := i.env
		tempEnv := NewEnclosedEnvironment(i.env)
		for cName, cValue := range classInfo.ConstantValues {
			tempEnv.Set(cName, cValue)
		}
		i.env = tempEnv

		constValue := i.Eval(constDecl.Value)
		i.env = savedEnv

		if isError(constValue) {
			return constValue
		}

		// Cache the evaluated value
		classInfo.ConstantValues[constDecl.Name.Value] = constValue
	}

	// Copy parent constants (child inherits all parent constants)
	if classInfo.Parent != nil {
		for constName, constDecl := range classInfo.Parent.Constants {
			// Only copy if not already defined in child class
			if _, exists := classInfo.Constants[constName]; !exists {
				classInfo.Constants[constName] = constDecl
			}
		}
		// Also copy parent constant values
		for constName, constValue := range classInfo.Parent.ConstantValues {
			// Only copy if not already defined in child class
			if _, exists := classInfo.ConstantValues[constName]; !exists {
				classInfo.ConstantValues[constName] = constValue
			}
		}
	}

	// Register class operators (Stage 8)
	for _, opDecl := range cd.Operators {
		if opDecl == nil {
			continue
		}
		if errVal := i.registerClassOperator(classInfo, opDecl); isError(errVal) {
			return errVal
		}
	}

	// Register class in registry
	i.classes[classInfo.Name] = classInfo

	return &NilValue{}
}

// convertPropertyDecl converts an AST property declaration to a PropertyInfo struct.
// This extracts the property metadata for runtime property access handling.
// Note: We mark all identifiers as field access for now and will check at runtime
// whether they're actually fields or methods.
func (i *Interpreter) convertPropertyDecl(propDecl *ast.PropertyDecl) *types.PropertyInfo {
	// Resolve property type
	var propType types.Type
	switch propDecl.Type.Name {
	case "Integer":
		propType = types.INTEGER
	case "Float":
		propType = types.FLOAT
	case "String":
		propType = types.STRING
	case "Boolean":
		propType = types.BOOLEAN
	default:
		// For now, treat unknown types as NIL
		// In a full implementation, we'd look up custom types
		propType = types.NIL
	}

	propInfo := &types.PropertyInfo{
		Name:            propDecl.Name.Value,
		Type:            propType,
		IsIndexed:       len(propDecl.IndexParams) > 0,
		IsDefault:       propDecl.IsDefault,
		IsClassProperty: propDecl.IsClassProperty,
	}

	// Determine read access kind and spec
	if propDecl.ReadSpec != nil {
		if ident, ok := propDecl.ReadSpec.(*ast.Identifier); ok {
			// It's an identifier - store the name, we'll check if it's a field or method at access time
			propInfo.ReadSpec = ident.Value
			// Mark as field for now - evalPropertyRead will check both fields and methods
			propInfo.ReadKind = types.PropAccessField
		} else {
			// It's an expression (Task 9.3c)
			propInfo.ReadKind = types.PropAccessExpression
			propInfo.ReadSpec = propDecl.ReadSpec.String()
			propInfo.ReadExpr = propDecl.ReadSpec // Store AST node for evaluation
		}
	} else {
		propInfo.ReadKind = types.PropAccessNone
	}

	// Determine write access kind and spec
	if propDecl.WriteSpec != nil {
		if ident, ok := propDecl.WriteSpec.(*ast.Identifier); ok {
			// It's an identifier - store the name, we'll check if it's a field or method at access time
			propInfo.WriteSpec = ident.Value
			// Mark as field for now - evalPropertyWrite will check both fields and methods
			propInfo.WriteKind = types.PropAccessField
		} else {
			propInfo.WriteKind = types.PropAccessNone
		}
	} else {
		propInfo.WriteKind = types.PropAccessNone
	}

	return propInfo
}

// evalInterfaceDeclaration evaluates an interface declaration.
// It builds an InterfaceInfo from the AST and registers it in the interface registry.
// Handles inheritance by linking to parent interface and inheriting its methods.
func (i *Interpreter) evalInterfaceDeclaration(id *ast.InterfaceDecl) Value {
	// Create new InterfaceInfo
	interfaceInfo := NewInterfaceInfo(id.Name.Value)

	// Handle inheritance if parent interface is specified
	if id.Parent != nil {
		parentName := id.Parent.Value
		parentInterface, exists := i.interfaces[parentName]
		if !exists {
			return i.newErrorWithLocation(id, "parent interface '%s' not found", parentName)
		}

		// Set parent reference
		interfaceInfo.Parent = parentInterface

		// Note: We don't copy parent methods here because InterfaceInfo.GetMethod()
		// and AllMethods() already handle parent interface traversal
	}

	// Add methods to InterfaceInfo
	// Convert InterfaceMethodDecl nodes to FunctionDecl nodes for consistency
	for _, methodDecl := range id.Methods {
		// Create a FunctionDecl from the InterfaceMethodDecl
		// Interface methods are declarations only (no body)
		funcDecl := &ast.FunctionDecl{
			Token:      methodDecl.Token,
			Name:       methodDecl.Name,
			Parameters: methodDecl.Parameters,
			ReturnType: methodDecl.ReturnType,
			Body:       nil, // Interface methods have no body
		}

		interfaceInfo.Methods[methodDecl.Name.Value] = funcDecl
	}

	// Register interface in registry
	i.interfaces[interfaceInfo.Name] = interfaceInfo

	return &NilValue{}
}

func (i *Interpreter) evalOperatorDeclaration(decl *ast.OperatorDecl) Value {
	if decl.Kind == ast.OperatorKindClass {
		// Class operators are registered during class declaration evaluation
		return &NilValue{}
	}

	if decl.Binding == nil {
		return i.newErrorWithLocation(decl, "operator '%s' missing binding", decl.OperatorSymbol)
	}

	operandTypes := make([]string, len(decl.OperandTypes))
	for idx, operand := range decl.OperandTypes {
		opRand := operand.String()
		operandTypes[idx] = normalizeTypeAnnotation(opRand)
	}

	if decl.Kind == ast.OperatorKindConversion {
		if len(operandTypes) != 1 {
			return i.newErrorWithLocation(decl, "conversion operator '%s' requires exactly one operand", decl.OperatorSymbol)
		}
		if decl.ReturnType == nil {
			return i.newErrorWithLocation(decl, "conversion operator '%s' requires a return type", decl.OperatorSymbol)
		}
		targetType := normalizeTypeAnnotation(decl.ReturnType.String())
		entry := &runtimeConversionEntry{
			From:        operandTypes[0],
			To:          targetType,
			BindingName: decl.Binding.Value,
			Implicit:    strings.EqualFold(decl.OperatorSymbol, "implicit"),
		}
		if err := i.conversions.register(entry); err != nil {
			return i.newErrorWithLocation(decl, "conversion from %s to %s already defined", operandTypes[0], targetType)
		}
		return &NilValue{}
	}

	entry := &runtimeOperatorEntry{
		Operator:     decl.OperatorSymbol,
		OperandTypes: operandTypes,
		BindingName:  decl.Binding.Value,
	}

	if err := i.globalOperators.register(entry); err != nil {
		return i.newErrorWithLocation(decl, "operator '%s' already defined for operand types (%s)", decl.OperatorSymbol, strings.Join(operandTypes, ", "))
	}

	return &NilValue{}
}

func (i *Interpreter) registerClassOperator(classInfo *ClassInfo, opDecl *ast.OperatorDecl) Value {
	if opDecl.Binding == nil {
		return i.newErrorWithLocation(opDecl, "class operator '%s' missing binding", opDecl.OperatorSymbol)
	}

	bindingName := opDecl.Binding.Value
	method, isClassMethod := classInfo.ClassMethods[bindingName]
	if !isClassMethod {
		var ok bool
		method, ok = classInfo.Methods[bindingName]
		if !ok {
			return i.newErrorWithLocation(opDecl, "binding '%s' for class operator '%s' not found in class '%s'", bindingName, opDecl.OperatorSymbol, classInfo.Name)
		}
	}

	classKey := normalizeTypeAnnotation(classInfo.Name)
	operandTypes := make([]string, 0, len(opDecl.OperandTypes)+1)
	includesClass := false
	for _, operand := range opDecl.OperandTypes {
		key := normalizeTypeAnnotation(operand.String())
		if key == classKey {
			includesClass = true
		}
		operandTypes = append(operandTypes, key)
	}
	if !includesClass {
		if strings.EqualFold(opDecl.OperatorSymbol, "in") {
			operandTypes = append(operandTypes, classKey)
		} else {
			operandTypes = append([]string{classKey}, operandTypes...)
		}
	}

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
		BindingName:   bindingName,
		Class:         classInfo,
		IsClassMethod: isClassMethod,
		SelfIndex:     selfIndex,
	}

	if err := classInfo.Operators.register(entry); err != nil {
		return i.newErrorWithLocation(opDecl, "class operator '%s' already defined for operand types (%s)", opDecl.OperatorSymbol, strings.Join(operandTypes, ", "))
	}

	if method.IsConstructor {
		classInfo.Constructors[method.Name.Value] = method
	}

	return &NilValue{}
}

// replaceMethodInOverloadList replaces a method declaration with its implementation in the overload list.
// Task 9.67: Helper for managing method overloads
// This function finds a method with matching signature and replaces it, or appends if not found.
// parametersMatch checks if two parameter lists have matching signatures
// (same count and same parameter types)
func parametersMatch(params1, params2 []*ast.Parameter) bool {
	if len(params1) != len(params2) {
		return false
	}
	for i := range params1 {
		// Compare parameter types
		if params1[i].Type != nil && params2[i].Type != nil {
			if params1[i].Type.Name != params2[i].Type.Name {
				return false
			}
		} else if params1[i].Type != params2[i].Type {
			// One has type, other doesn't
			return false
		}
	}
	return true
}

func (i *Interpreter) replaceMethodInOverloadList(list []*ast.FunctionDecl, impl *ast.FunctionDecl) []*ast.FunctionDecl {
	// Check if we already have a declaration for this overload signature
	for idx, decl := range list {
		// Match by parameter count and types
		if parametersMatch(decl.Parameters, impl.Parameters) {
			// Replace the declaration with the implementation
			list[idx] = impl
			return list
		}
	}
	// No matching declaration found - append the implementation
	return append(list, impl)
}
