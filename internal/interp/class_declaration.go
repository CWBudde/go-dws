package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

func (c *ClassInfo) IsPartialClass() bool {
	return c != nil && c.IsPartial
}

func (c *ClassInfo) SetPartialClass(isPartial bool) {
	if c != nil {
		c.IsPartial = isPartial
		if c.Metadata != nil {
			c.Metadata.IsPartial = isPartial
		}
	}
}

func (c *ClassInfo) SetAbstractClass(isAbstract bool) {
	if c != nil {
		c.IsAbstractFlag = isAbstract
		if c.Metadata != nil {
			c.Metadata.IsAbstract = isAbstract
		}
	}
}

func (c *ClassInfo) SetExternalClass(isExternal bool, externalName string) {
	if c == nil {
		return
	}
	c.IsExternalFlag = isExternal
	c.ExternalName = externalName
	if c.Metadata != nil {
		c.Metadata.IsExternal = isExternal
		c.Metadata.ExternalName = externalName
	}
}

func (c *ClassInfo) HasNoParentClass() bool {
	return c != nil && c.Parent == nil
}

func (c *ClassInfo) DefineCurrentClassMarker(env *Environment) {
	if c != nil && env != nil {
		env.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: c})
	}
}

func (c *ClassInfo) DefineInEnv(env *Environment) {
	if c != nil && env != nil {
		env.Define(c.Name, &ClassValue{ClassInfo: c})
	}
}

func (c *ClassInfo) SetParentClass(parent any) {
	parentClass, ok := parent.(*ClassInfo)
	if c == nil || !ok || parentClass == nil || c.Parent != nil {
		return
	}

	c.Parent = parentClass
	c.Metadata.Parent = parentClass.Metadata
	c.Metadata.ParentName = parentClass.Metadata.Name

	for fieldName, fieldType := range parentClass.Fields {
		c.Fields[fieldName] = fieldType
	}
	for fieldName, fieldDecl := range parentClass.FieldDecls {
		c.FieldDecls[fieldName] = fieldDecl
	}
	for methodName, methodDecl := range parentClass.Methods {
		c.Methods[methodName] = methodDecl
	}
	for methodName, methodDecl := range parentClass.ClassMethods {
		c.ClassMethods[methodName] = methodDecl
	}
	for name, constructor := range parentClass.Constructors {
		c.Constructors[ident.Normalize(name)] = constructor
	}
	for name, overloads := range parentClass.ConstructorOverloads {
		c.ConstructorOverloads[ident.Normalize(name)] = append([]*ast.FunctionDecl(nil), overloads...)
	}
	if parentClass.DefaultConstructor != "" {
		c.DefaultConstructor = parentClass.DefaultConstructor
	}

	for name, constructor := range parentClass.Metadata.Constructors {
		if c.Metadata.Constructors == nil {
			c.Metadata.Constructors = make(map[string]*runtime.MethodMetadata)
		}
		c.Metadata.Constructors[name] = constructor
	}
	for name, overloads := range parentClass.Metadata.ConstructorOverloads {
		if c.Metadata.ConstructorOverloads == nil {
			c.Metadata.ConstructorOverloads = make(map[string][]*runtime.MethodMetadata)
		}
		c.Metadata.ConstructorOverloads[name] = append([]*runtime.MethodMetadata(nil), overloads...)
	}
	if parentClass.Metadata.DefaultConstructor != "" {
		c.Metadata.DefaultConstructor = parentClass.Metadata.DefaultConstructor
	}

	c.Operators = parentClass.Operators.clone()
}

func (c *ClassInfo) AddImplementedInterface(iface any, ifaceName string) {
	interfaceInfo, ok := iface.(*InterfaceInfo)
	if c == nil || !ok || interfaceInfo == nil {
		return
	}
	c.Interfaces = append(c.Interfaces, interfaceInfo)
	c.Metadata.Interfaces = append(c.Metadata.Interfaces, ifaceName)
}

func (c *ClassInfo) AddConstantValue(constDecl *ast.ConstDecl, value Value) {
	if c == nil || constDecl == nil {
		return
	}
	c.Constants[constDecl.Name.Value] = constDecl
	c.ConstantValues[constDecl.Name.Value] = value
	if c.Metadata.Constants == nil {
		c.Metadata.Constants = make(map[string]any)
	}
	c.Metadata.Constants[constDecl.Name.Value] = value
}

func (c *ClassInfo) ConstantValuesCopy() map[string]Value {
	if c == nil {
		return nil
	}
	result := make(map[string]Value, len(c.ConstantValues))
	for name, val := range c.ConstantValues {
		result[name] = val
	}
	return result
}

func (c *ClassInfo) InheritConstantValuesFrom(parent any) {
	parentClass, ok := parent.(*ClassInfo)
	if c == nil || !ok || parentClass == nil {
		return
	}
	for name, decl := range parentClass.Constants {
		if _, exists := c.Constants[name]; !exists {
			c.Constants[name] = decl
		}
	}
	for name, val := range parentClass.ConstantValues {
		if _, exists := c.ConstantValues[name]; !exists {
			c.ConstantValues[name] = val
		}
	}
	if parentClass.Metadata != nil && parentClass.Metadata.Constants != nil {
		if c.Metadata.Constants == nil {
			c.Metadata.Constants = make(map[string]any)
		}
		for name, val := range parentClass.Metadata.Constants {
			if _, exists := c.Metadata.Constants[name]; !exists {
				c.Metadata.Constants[name] = val
			}
		}
	}
}

func (c *ClassInfo) AddFieldDeclaration(fieldDecl *ast.FieldDecl, fieldType types.Type) {
	if c == nil || fieldDecl == nil {
		return
	}
	c.Fields[fieldDecl.Name.Value] = fieldType
	c.FieldDecls[fieldDecl.Name.Value] = fieldDecl

	fieldMeta := runtime.FieldMetadataFromAST(fieldDecl)
	fieldMeta.Type = fieldType
	runtime.AddFieldToClass(c.Metadata, fieldMeta)
}

func (c *ClassInfo) AddClassVarValue(name string, value Value) {
	if c != nil {
		c.ClassVars[name] = value
	}
}

func (c *ClassInfo) AddNestedClassRef(nestedName string, nestedClass any) {
	nestedInfo, ok := nestedClass.(*ClassInfo)
	if c != nil && ok && nestedInfo != nil {
		c.NestedClasses[ident.Normalize(nestedName)] = nestedInfo
	}
}

func (c *ClassInfo) LookupDeclaredMethod(methodName string, isClassMethod bool) (*ast.FunctionDecl, bool) {
	if c == nil {
		return nil, false
	}
	normalizedName := ident.Normalize(methodName)
	if isClassMethod {
		method, exists := c.ClassMethods[normalizedName]
		return method, exists
	}
	method, exists := c.Methods[normalizedName]
	return method, exists
}

func (c *ClassInfo) SetConstructorDecl(constructor *ast.FunctionDecl) {
	if c != nil {
		c.Constructor = constructor
	}
}

func (c *ClassInfo) SetDestructorDecl(destructor *ast.FunctionDecl) {
	if c != nil {
		c.Destructor = destructor
	}
}

func (c *ClassInfo) InheritDestructorMetadataIfMissing() {
	if c != nil && c.Metadata.Destructor == nil && c.Parent != nil && c.Parent.Metadata.Destructor != nil {
		c.Metadata.Destructor = c.Parent.Metadata.Destructor
	}
}

func (c *ClassInfo) SynthesizeImplicitDefaultConstructor() {
	if c == nil {
		return
	}
	for ctorName, overloads := range c.ConstructorOverloads {
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
		if hasOverloadDirective && !hasParameterlessOverload {
			implicitConstructor := &ast.FunctionDecl{
				BaseNode:      ast.BaseNode{},
				Name:          &ast.Identifier{Value: ctorName},
				Parameters:    []*ast.Parameter{},
				IsConstructor: true,
				IsOverload:    true,
			}
			normalizedName := ident.Normalize(ctorName)
			if _, exists := c.Constructors[normalizedName]; !exists {
				c.Constructors[normalizedName] = implicitConstructor
			}
			c.ConstructorOverloads[normalizedName] = append(c.ConstructorOverloads[normalizedName], implicitConstructor)
		}
	}
}

func (c *ClassInfo) SetPropertyInfo(name string, propInfo *types.PropertyInfo) {
	if c != nil && propInfo != nil {
		c.Properties[name] = propInfo
	}
}

func (c *ClassInfo) DeterminePropertyAccessKind(specName string) types.PropAccessKind {
	if c == nil {
		return types.PropAccessMethod
	}

	normalizedName := ident.Normalize(specName)
	for current := c; current != nil; current = current.Parent {
		if _, isField := current.Fields[normalizedName]; isField {
			return types.PropAccessField
		}
		if _, isField := current.Fields[specName]; isField {
			return types.PropAccessField
		}
		if _, isClassVar := current.ClassVars[specName]; isClassVar {
			return types.PropAccessField
		}
		if _, isClassVar := current.ClassVars[normalizedName]; isClassVar {
			return types.PropAccessField
		}
	}

	for current := c; current != nil; current = current.Parent {
		if _, isMethod := current.Methods[normalizedName]; isMethod {
			return types.PropAccessMethod
		}
		if _, isClassMethod := current.ClassMethods[normalizedName]; isClassMethod {
			return types.PropAccessMethod
		}
	}

	return types.PropAccessNone
}

func (c *ClassInfo) AddMethodDeclaration(method *ast.FunctionDecl, className string, registry *runtime.MethodRegistry) bool {
	if c == nil || method == nil {
		return false
	}

	normalizedMethodName := ident.Normalize(method.Name.Value)

	if !method.IsConstructor && ident.Equal(method.Name.Value, "Create") && method.ReturnType != nil {
		if ident.Equal(method.ReturnType.String(), className) {
			method.IsConstructor = true
		}
	}

	methodMeta := runtime.MethodMetadataFromAST(method)
	if registry != nil {
		registry.RegisterMethod(methodMeta)
	}

	if method.IsClassMethod {
		c.ClassMethods[normalizedMethodName] = method
		c.ClassMethodOverloads[normalizedMethodName] = append(c.ClassMethodOverloads[normalizedMethodName], method)
		if !method.IsConstructor && !method.IsDestructor {
			runtime.AddMethodToClass(c.Metadata, methodMeta, true)
		}
	} else {
		c.Methods[normalizedMethodName] = method
		c.MethodOverloads[normalizedMethodName] = append(c.MethodOverloads[normalizedMethodName], method)
		if !method.IsConstructor && !method.IsDestructor {
			runtime.AddMethodToClass(c.Metadata, methodMeta, false)
		}
	}

	if method.IsDestructor {
		c.Metadata.Destructor = methodMeta
	}

	if method.IsConstructor {
		normalizedName := ident.Normalize(method.Name.Value)
		c.Constructors[normalizedName] = method
		runtime.AddConstructorToClass(c.Metadata, methodMeta)
		if method.IsDefault {
			c.DefaultConstructor = method.Name.Value
		}

		existingOverloads := c.ConstructorOverloads[normalizedName]
		replaced := false
		for idx, existingMethod := range existingOverloads {
			if parametersMatch(existingMethod.Parameters, method.Parameters) {
				existingOverloads[idx] = method
				replaced = true
				break
			}
		}
		if !replaced {
			existingOverloads = append(existingOverloads, method)
		}
		c.ConstructorOverloads[normalizedName] = existingOverloads
	}

	return true
}

func (c *ClassInfo) InheritParentPropertyInfos() {
	if c == nil || c.Parent == nil {
		return
	}
	for propName, propInfo := range c.Parent.Properties {
		if _, exists := c.Properties[propName]; !exists {
			c.Properties[propName] = propInfo
		}
	}
}

func (c *ClassInfo) RegisterOperatorBinding(operatorSymbol, bindingName string, operandTypes []string) error {
	if c == nil {
		return nil
	}

	normalizedBindingName := ident.Normalize(bindingName)
	_, isClassMethod := c.ClassMethods[normalizedBindingName]
	if !isClassMethod {
		if _, ok := c.Methods[normalizedBindingName]; !ok {
			return fmt.Errorf("binding '%s' for class operator '%s' not found in class '%s'", bindingName, operatorSymbol, c.Name)
		}
	}

	classKey := NormalizeTypeAnnotation(c.Name)
	normalizedOperands := make([]string, 0, len(operandTypes)+1)
	includesClass := false
	for _, operandType := range operandTypes {
		key := NormalizeTypeAnnotation(operandType)
		if key == classKey {
			includesClass = true
		}
		normalizedOperands = append(normalizedOperands, key)
	}
	if !includesClass {
		if ident.Equal(operatorSymbol, "in") {
			normalizedOperands = append(normalizedOperands, classKey)
		} else {
			normalizedOperands = append([]string{classKey}, normalizedOperands...)
		}
	}

	selfIndex := -1
	if !isClassMethod {
		for idx, key := range normalizedOperands {
			if key == classKey {
				selfIndex = idx
				break
			}
		}
		if selfIndex == -1 {
			return fmt.Errorf("unable to determine self operand for class operator '%s'", operatorSymbol)
		}
	}

	entry := &runtimeOperatorEntry{
		Operator:      operatorSymbol,
		OperandTypes:  normalizedOperands,
		BindingName:   normalizedBindingName,
		Class:         c,
		IsClassMethod: isClassMethod,
		SelfIndex:     selfIndex,
	}

	if err := c.Operators.register(entry); err != nil {
		return fmt.Errorf("class operator '%s' already defined for operand types (%s)", operatorSymbol, strings.Join(normalizedOperands, ", "))
	}

	return nil
}

func (c *ClassInfo) BuildVirtualMethodTableDirect() {
	if c != nil {
		c.buildVirtualMethodTable()
	}
}

func (c *ClassInfo) RegisterInTypeSystem(ts any, parentName string) {
	typeSystem, ok := ts.(*interptypes.TypeSystem)
	if c != nil && ok && typeSystem != nil {
		typeSystem.RegisterClassWithParent(c.Name, c, parentName)
	}
}

func (c *ClassInfo) RegisterMethodImplementation(fn *ast.FunctionDecl, allClasses map[string]interptypes.ClassInfo) {
	if c == nil || fn == nil {
		return
	}

	normalizedMethodName := ident.Normalize(fn.Name.Value)

	if fn.IsClassMethod {
		c.ClassMethods[normalizedMethodName] = fn
		overloads := c.ClassMethodOverloads[normalizedMethodName]
		c.ClassMethodOverloads[normalizedMethodName] = replaceMethodInOverloadListNoReceiver(overloads, fn)
	} else {
		c.Methods[normalizedMethodName] = fn
		overloads := c.MethodOverloads[normalizedMethodName]
		c.MethodOverloads[normalizedMethodName] = replaceMethodInOverloadListNoReceiver(overloads, fn)
	}

	if fn.IsConstructor {
		normalizedCtorName := ident.Normalize(fn.Name.Value)
		c.Constructors[normalizedCtorName] = fn
		overloads := c.ConstructorOverloads[normalizedCtorName]
		c.ConstructorOverloads[normalizedCtorName] = replaceMethodInOverloadListNoReceiver(overloads, fn)
		c.Constructor = fn
		c.propagateConstructorImplementation(allClasses, fn)
	}

	if fn.IsDestructor {
		c.Destructor = fn
	}

	c.buildVirtualMethodTable()
	c.propagateMethodImplementation(allClasses, normalizedMethodName, fn, fn.IsClassMethod)
	c.rebuildDescendantVMTs(allClasses)
}

func (c *ClassInfo) rebuildDescendantVMTs(allClasses map[string]interptypes.ClassInfo) {
	for _, classInfoAny := range allClasses {
		classInfo, ok := classInfoAny.(*ClassInfo)
		if !ok {
			continue
		}
		if isDescendantOfClass(classInfo, c) {
			classInfo.buildVirtualMethodTable()
		}
	}
}

func isDescendantOfClass(childClass, ancestorClass *ClassInfo) bool {
	current := childClass.Parent
	for current != nil {
		if current == ancestorClass {
			return true
		}
		current = current.Parent
	}
	return false
}

func (c *ClassInfo) propagateMethodImplementation(allClasses map[string]interptypes.ClassInfo, normalizedMethodName string, fn *ast.FunctionDecl, isClassMethod bool) {
	for _, classInfoAny := range allClasses {
		classInfo, ok := classInfoAny.(*ClassInfo)
		if !ok || !isDescendantOfClass(classInfo, c) {
			continue
		}

		if isClassMethod {
			if existing, ok := classInfo.ClassMethods[normalizedMethodName]; ok {
				if existing.ClassName != nil && ident.Equal(existing.ClassName.Value, classInfo.Name) {
					continue
				}
				classInfo.ClassMethods[normalizedMethodName] = fn
			}
		} else {
			if existing, ok := classInfo.Methods[normalizedMethodName]; ok {
				if existing.ClassName != nil && ident.Equal(existing.ClassName.Value, classInfo.Name) {
					continue
				}
				classInfo.Methods[normalizedMethodName] = fn
			}
		}
	}
}

func (c *ClassInfo) propagateConstructorImplementation(allClasses map[string]interptypes.ClassInfo, fn *ast.FunctionDecl) {
	normalizedCtorName := ident.Normalize(fn.Name.Value)

	for _, classInfoAny := range allClasses {
		classInfo, ok := classInfoAny.(*ClassInfo)
		if !ok || !isDescendantOfClass(classInfo, c) {
			continue
		}

		if ctor, ok := classInfo.Constructors[normalizedCtorName]; ok && ctor != nil {
			if ctor.ClassName != nil && ident.Equal(ctor.ClassName.Value, classInfo.Name) {
				continue
			}
			if parametersMatch(ctor.Parameters, fn.Parameters) {
				classInfo.Constructors[normalizedCtorName] = fn
			}
		}

		if overloads, ok := classInfo.ConstructorOverloads[normalizedCtorName]; ok {
			for idx, decl := range overloads {
				if decl == nil {
					continue
				}
				if decl.ClassName != nil && ident.Equal(decl.ClassName.Value, classInfo.Name) {
					continue
				}
				if parametersMatch(decl.Parameters, fn.Parameters) {
					overloads[idx] = fn
				}
			}
			classInfo.ConstructorOverloads[normalizedCtorName] = overloads
		}
	}
}

func replaceMethodInOverloadListNoReceiver(list []*ast.FunctionDecl, impl *ast.FunctionDecl) []*ast.FunctionDecl {
	for idx, decl := range list {
		if parametersMatch(decl.Parameters, impl.Parameters) {
			list[idx] = impl
			return list
		}
	}
	return append(list, impl)
}
