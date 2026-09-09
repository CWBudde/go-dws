package runtime

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// IsPartialClass reports whether the class permits declarations in multiple parts.
func (c *ClassInfo) IsPartialClass() bool {
	return c != nil && c.IsPartial
}

// SetPartialClass updates the partial flag in both class metadata representations.
func (c *ClassInfo) SetPartialClass(isPartial bool) {
	if c != nil {
		c.IsPartial = isPartial
		if c.Metadata != nil {
			c.Metadata.IsPartial = isPartial
		}
	}
}

// IsForwardClass reports whether this class exists only as a forward-declaration
// placeholder awaiting its full definition.
func (c *ClassInfo) IsForwardClass() bool {
	return c != nil && c.IsForwardDecl
}

// SetForwardClass marks (or clears) this class as an incomplete forward declaration.
func (c *ClassInfo) SetForwardClass(isForward bool) {
	if c != nil {
		c.IsForwardDecl = isForward
	}
}

// SetAbstractClass updates the abstract flag in both class metadata representations.
func (c *ClassInfo) SetAbstractClass(isAbstract bool) {
	if c != nil {
		c.IsAbstractFlag = isAbstract
		if c.Metadata != nil {
			c.Metadata.IsAbstract = isAbstract
		}
	}
}

// SetExternalClass records whether the class is external and its host binding name.
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

// HasNoParentClass reports whether an existing class has no parent.
func (c *ClassInfo) HasNoParentClass() bool {
	return c != nil && c.Parent == nil
}

// DefineCurrentClassMarker binds this class as the current class in env.
func (c *ClassInfo) DefineCurrentClassMarker(env *Environment) {
	if c != nil && env != nil {
		env.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: c})
	}
}

// DefineInEnv binds a metaclass value under the class name in env.
func (c *ClassInfo) DefineInEnv(env *Environment) {
	if c != nil && env != nil {
		env.Define(c.Name, &ClassValue{ClassInfo: c})
	}
}

// SetParentClass sets the parent once and inherits its field, method, constructor, and operator bindings.
func (c *ClassInfo) SetParentClass(parent any) {
	parentClass, ok := parent.(*ClassInfo)
	if c == nil || !ok || parentClass == nil || c.Parent != nil {
		return
	}

	c.Parent = parentClass
	if !c.typeShared && c.Type != nil {
		c.Type.Parent = parentClass.GetClassType()
	}
	c.Metadata.Parent = parentClass.Metadata

	for name, constructor := range parentClass.Constructors {
		c.Constructors[ident.Normalize(name)] = constructor
	}
	for name, overloads := range parentClass.ConstructorOverloads {
		c.ConstructorOverloads[ident.Normalize(name)] = append([]*MethodMetadata(nil), overloads...)
	}
	if parentClass.DefaultConstructor != "" {
		c.DefaultConstructor = parentClass.DefaultConstructor
	}

	c.Operators = parentClass.Operators.Clone()
}

// AddImplementedInterface records an implemented runtime interface and its metadata name.
func (c *ClassInfo) AddImplementedInterface(iface any, ifaceName string) {
	interfaceInfo, ok := iface.(*MutableInterfaceInfo)
	if c == nil || !ok || interfaceInfo == nil {
		return
	}
	c.Interfaces = append(c.Interfaces, interfaceInfo)
	c.Metadata.Interfaces = append(c.Metadata.Interfaces, ifaceName)
}

// AddConstantValue stores a constant declaration and its evaluated value.
func (c *ClassInfo) AddConstantValue(constDecl *ast.ConstDecl, value Value) {
	if c == nil || constDecl == nil {
		return
	}
	c.Constants[constDecl.Name.Value] = value
}

// ConstantValuesCopy copies the constant bindings while sharing their runtime values.
func (c *ClassInfo) ConstantValuesCopy() map[string]Value {
	if c == nil {
		return nil
	}
	result := make(map[string]Value, len(c.Constants))
	for name, val := range c.Constants {
		result[name] = val
	}
	return result
}

// InheritConstantValuesFrom copies parent constants that this class has not declared.
func (c *ClassInfo) InheritConstantValuesFrom(parent any) {
	parentClass, ok := parent.(*ClassInfo)
	if c == nil || !ok || parentClass == nil {
		return
	}
	for name, value := range parentClass.Constants {
		if _, exists := c.Constants[name]; !exists {
			c.Constants[name] = value
		}
	}
}

// AddFieldDeclaration registers a field declaration and its resolved type in class metadata.
func (c *ClassInfo) AddFieldDeclaration(fieldDecl *ast.FieldDecl, fieldType types.Type) {
	if c == nil || fieldDecl == nil {
		return
	}

	fieldMeta := FieldMetadataFromAST(fieldDecl)
	fieldMeta.Type = fieldType
	AddFieldToClass(c.Metadata, fieldMeta)
}

// AddClassVarValue stores a class variable value under its declared name.
func (c *ClassInfo) AddClassVarValue(name string, value Value) {
	if c != nil {
		c.ClassVars[name] = value
	}
}

// AddNestedClassRef registers a nested class with case-insensitive lookup.
func (c *ClassInfo) AddNestedClassRef(nestedName string, nestedClass any) {
	nestedInfo, ok := nestedClass.(*ClassInfo)
	if c != nil && ok && nestedInfo != nil {
		c.NestedClasses[ident.Normalize(nestedName)] = nestedInfo
	}
}

// InheritDestructorMetadataIfMissing uses the parent destructor when none is recorded locally.
func (c *ClassInfo) InheritDestructorMetadataIfMissing() {
	if c != nil && c.Metadata.Destructor == nil && c.Parent != nil && c.Parent.Metadata.Destructor != nil {
		c.Metadata.Destructor = c.Parent.Metadata.Destructor
	}
}

// SynthesizeImplicitDefaultConstructor adds a parameterless constructor for an overloaded constructor group that lacks one.
func (c *ClassInfo) SynthesizeImplicitDefaultConstructor() {
	if c == nil {
		return
	}
	for ctorName, overloads := range c.ConstructorOverloads {
		hasOverloadDirective := false
		hasParameterlessOverload := false
		for _, ctor := range overloads {
			if ctor.Declaration != nil && ctor.Declaration.IsOverload {
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
			callable := MethodMetadataFromAST(implicitConstructor)
			callable.Owner = c
			if _, exists := c.Constructors[normalizedName]; !exists {
				c.Constructors[normalizedName] = callable
			}
			c.ConstructorOverloads[normalizedName] = append(c.ConstructorOverloads[normalizedName], callable)
		}
	}
}

// SetPropertyInfo registers a property descriptor under its declared name.
func (c *ClassInfo) SetPropertyInfo(name string, propInfo *types.PropertyInfo) {
	if c != nil && propInfo != nil {
		c.Properties[name] = propInfo
	}
}

// DeterminePropertyAccessKind classifies an accessor name as a field, method, or missing member.
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

// AddMethodDeclaration registers a method and its overload, constructor, or destructor metadata.
func (c *ClassInfo) AddMethodDeclaration(method *ast.FunctionDecl, className string, registry *MethodRegistry) bool {
	if c == nil || method == nil {
		return false
	}
	if !method.IsConstructor && ident.Equal(method.Name.Value, "Create") && method.ReturnType != nil && ident.Equal(method.ReturnType.String(), className) {
		method.IsConstructor = true
	}
	callable := MethodMetadataFromAST(method)
	callable.Owner = c
	if registry != nil {
		registry.RegisterMethod(callable)
	}
	name := ident.Normalize(method.Name.Value)
	switch {
	case method.IsConstructor:
		c.Constructors[name] = callable
		c.ConstructorOverloads[name] = replaceCallable(c.ConstructorOverloads[name], callable)
		if method.IsDefault {
			c.DefaultConstructor = method.Name.Value
			c.Metadata.DefaultConstructor = method.Name.Value
		}
	case method.IsDestructor:
		c.Destructor = callable
	default:
		if method.IsClassMethod {
			c.ClassMethods[name] = callable
			c.ClassMethodOverloads[name] = append(c.ClassMethodOverloads[name], callable)
		} else {
			c.Methods[name] = callable
			c.MethodOverloads[name] = append(c.MethodOverloads[name], callable)
		}
	}
	return true
}

// InheritParentPropertyInfos shares parent property descriptors unless this class overrides them.
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

// RegisterOperatorBinding registers an operator’s operand signature and bound method.
func (c *ClassInfo) RegisterOperatorBinding(operatorSymbol, bindingName string, operandTypes []types.Type) error {
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

	classType := c.GetClassType()
	normalizedOperands := append([]types.Type(nil), operandTypes...)
	includesClass := false
	for _, operand := range operandTypes {
		if types.OperatorTypesEqual(operand, classType) {
			includesClass = true
		}
	}
	if !includesClass {
		if ident.Equal(operatorSymbol, "in") {
			normalizedOperands = append(normalizedOperands, classType)
		} else {
			normalizedOperands = append([]types.Type{classType}, normalizedOperands...)
		}
	}
	selfIndex := -1
	if !isClassMethod {
		for i, operand := range normalizedOperands {
			if types.OperatorTypesEqual(operand, classType) {
				selfIndex = i
				break
			}
		}
		if selfIndex == -1 {
			return fmt.Errorf("unable to determine self operand for class operator '%s'", operatorSymbol)
		}
	}

	entry := &ClassOperatorEntry{
		Operator:      operatorSymbol,
		OperandTypes:  normalizedOperands,
		BindingName:   normalizedBindingName,
		Class:         c,
		IsClassMethod: isClassMethod,
		SelfIndex:     selfIndex,
	}

	if err := c.Operators.Register(entry); err != nil {
		return fmt.Errorf("class operator '%s' already defined for operand types (%s)", operatorSymbol, types.FormatTypeList(normalizedOperands))
	}

	return nil
}

// BuildVirtualMethodTableDirect rebuilds inherited and locally overridden virtual method bindings.
func (c *ClassInfo) BuildVirtualMethodTableDirect() {
	if c != nil {
		c.buildVirtualMethodTable()
	}
}

// RegisterInTypeSystem registers this class and its parent name through the registry interface.
func (c *ClassInfo) RegisterInTypeSystem(ts any, parentName string) {
	typeSystem, ok := ts.(interface {
		RegisterClassWithParent(string, IClassInfo, string)
	})
	if c != nil && ok && typeSystem != nil {
		typeSystem.RegisterClassWithParent(c.Name, c, parentName)
	}
}

// RegisterMethodImplementation replaces an out-of-line declaration and updates inherited bindings.
func (c *ClassInfo) RegisterMethodImplementation(fn *ast.FunctionDecl, allClasses map[string]IClassInfo) {
	if c == nil || fn == nil {
		return
	}
	name := ident.Normalize(fn.Name.Value)
	existing := c.findCallableDeclaration(fn)
	if existing != nil {
		existing.BindImplementation(fn)
	} else {
		c.AddMethodDeclaration(fn, c.Name, nil)
		existing = c.findCallableDeclaration(fn)
	}
	switch {
	case fn.IsConstructor:
		c.Constructors[name] = existing
	case fn.IsDestructor:
		c.Destructor = existing
	case fn.IsClassMethod:
		c.ClassMethods[name] = existing
	default:
		c.Methods[name] = existing
	}
	c.buildVirtualMethodTable()
	c.rebuildDescendantVMTs(allClasses)
}

func (c *ClassInfo) rebuildDescendantVMTs(allClasses map[string]IClassInfo) {
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

// findCallableDeclaration locates the canonical callable for a declaration signature.
func (c *ClassInfo) findCallableDeclaration(fn *ast.FunctionDecl) *MethodMetadata {
	if c == nil || fn == nil {
		return nil
	}
	name := ident.Normalize(fn.Name.Value)
	var candidates []*MethodMetadata
	switch {
	case fn.IsConstructor:
		candidates = c.ConstructorOverloads[name]
	case fn.IsDestructor:
		return c.Destructor
	case fn.IsClassMethod:
		candidates = c.ClassMethodOverloads[name]
	default:
		candidates = c.MethodOverloads[name]
	}
	for _, candidate := range candidates {
		if candidate != nil && candidate.Declaration != nil && parameterTypesEqualFold(candidate.Declaration.Parameters, fn.Parameters) {
			return candidate
		}
	}
	return nil
}

func replaceCallable(list []*MethodMetadata, method *MethodMetadata) []*MethodMetadata {
	for index, existing := range list {
		if existing != nil && parameterTypesEqualFold(existing.Declaration.Parameters, method.Declaration.Parameters) {
			list[index] = method
			return list
		}
	}
	return append(list, method)
}

// parameterTypesEqualFold compares two parameter lists by declared type name,
// case-insensitively (DWScript identifiers are case-insensitive, so a
// declaration using "string" matches an implementation using "String").
func parameterTypesEqualFold(params1, params2 []*ast.Parameter) bool {
	if len(params1) != len(params2) {
		return false
	}
	for i := range params1 {
		if params1[i].Type != nil && params2[i].Type != nil {
			if !ident.Equal(params1[i].Type.String(), params2[i].Type.String()) {
				return false
			}
		} else if params1[i].Type != params2[i].Type {
			return false
		}
	}
	return true
}

// mergeParameterDefaults copies parameter default values from the class
// declaration into an out-of-line implementation that did not respecify them
// (DWScript's "default not respecified" rule).
func mergeParameterDefaults(impl, decl *ast.FunctionDecl) {
	if impl == nil || decl == nil || len(impl.Parameters) != len(decl.Parameters) {
		return
	}
	for i, declParam := range decl.Parameters {
		if impl.Parameters[i].DefaultValue == nil && declParam.DefaultValue != nil {
			impl.Parameters[i].DefaultValue = declParam.DefaultValue
		}
	}
}

// CallableForDeclaration returns the canonical callable for a declaration signature.
func (c *ClassInfo) CallableForDeclaration(declaration *ast.FunctionDecl) *MethodMetadata {
	for current := c; current != nil; current = current.Parent {
		if callable := current.findCallableDeclaration(declaration); callable != nil {
			return callable
		}
	}
	return nil
}
