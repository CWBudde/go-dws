package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ===== Interface Registry =====

// lookupInterfaceInfo finds an interface by name. Returns nil if not found.
func (i *Interpreter) lookupInterfaceInfo(name string) *InterfaceInfo {
	return i.LookupInterfaceInfo(name)
}

// LookupInterfaceInfo finds an interface by name (public API). Returns nil if not found.
func (i *Interpreter) LookupInterfaceInfo(name string) *InterfaceInfo {
	iface := i.typeSystem.LookupInterface(name)
	if iface == nil {
		return nil
	}
	// Type assert from any to *InterfaceInfo
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		return ifaceInfo
	}
	return nil
}

// ===== Interface Declaration Adapters =====

// NewInterfaceInfoAdapter creates a new InterfaceInfo for the evaluator.
func (i *Interpreter) NewInterfaceInfoAdapter(name string) interface{} {
	return NewInterfaceInfo(name)
}

// CastToInterfaceInfo performs type assertion from any to *InterfaceInfo.
func (i *Interpreter) CastToInterfaceInfo(iface interface{}) (interface{}, bool) {
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		return ifaceInfo, true
	}
	return nil, false
}

// SetInterfaceParent sets the parent interface for inheritance.
func (i *Interpreter) SetInterfaceParent(iface interface{}, parent interface{}) {
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		if parentInfo, ok := parent.(*InterfaceInfo); ok {
			ifaceInfo.Parent = parentInfo
		}
	}
}

// GetInterfaceName returns the name of an interface.
func (i *Interpreter) GetInterfaceName(iface interface{}) string {
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		return ifaceInfo.Name
	}
	return ""
}

// GetInterfaceParent returns the parent interface (for hierarchy traversal).
func (i *Interpreter) GetInterfaceParent(iface interface{}) interface{} {
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		return ifaceInfo.Parent
	}
	return nil
}

// AddInterfaceMethod adds a method to an interface.
func (i *Interpreter) AddInterfaceMethod(iface interface{}, normalizedName string, method *ast.FunctionDecl) {
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		ifaceInfo.Methods[normalizedName] = method
	}
}

// AddInterfaceProperty adds a property to an interface.
func (i *Interpreter) AddInterfaceProperty(iface interface{}, normalizedName string, propInfo any) {
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		if prop, ok := propInfo.(*types.PropertyInfo); ok {
			ifaceInfo.Properties[normalizedName] = prop
		}
	}
}

// ===== Helper Declaration Adapters =====

// CreateHelperInfo creates a new HelperInfo for the given target type.
func (i *Interpreter) CreateHelperInfo(name string, targetType any, isRecordHelper bool) interface{} {
	if tt, ok := targetType.(types.Type); ok {
		return NewHelperInfo(name, tt, isRecordHelper)
	}
	return nil
}

// SetHelperParent sets the parent helper for inheritance chain.
func (i *Interpreter) SetHelperParent(helper interface{}, parent interface{}) {
	if h, ok := helper.(*HelperInfo); ok {
		if p, ok := parent.(*HelperInfo); ok {
			h.ParentHelper = p
		}
	}
}

// VerifyHelperTargetTypeMatch checks if parent's target type matches.
func (i *Interpreter) VerifyHelperTargetTypeMatch(parent interface{}, targetType any) bool {
	if p, ok := parent.(*HelperInfo); ok {
		if tt, ok := targetType.(types.Type); ok {
			return p.TargetType.Equals(tt)
		}
	}
	return false
}

// GetHelperName returns the name of a helper.
func (i *Interpreter) GetHelperName(helper interface{}) string {
	if h, ok := helper.(*HelperInfo); ok {
		return h.Name
	}
	return ""
}

// AddHelperMethod registers a method in the helper.
func (i *Interpreter) AddHelperMethod(helper interface{}, normalizedName string, method *ast.FunctionDecl) {
	if h, ok := helper.(*HelperInfo); ok {
		h.Methods[normalizedName] = method
	}
}

// AddHelperProperty registers a property in the helper with read/write accessors.
func (i *Interpreter) AddHelperProperty(helper interface{}, prop *ast.PropertyDecl, propType any) {
	h, ok := helper.(*HelperInfo)
	if !ok {
		return
	}
	pt, _ := propType.(types.Type)
	propInfo := &types.PropertyInfo{Name: prop.Name.Value, Type: pt}

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
	h.Properties[prop.Name.Value] = propInfo
}

// AddHelperClassVar adds a class variable to the helper.
func (i *Interpreter) AddHelperClassVar(helper interface{}, name string, value Value) {
	if h, ok := helper.(*HelperInfo); ok {
		h.ClassVars[ident.Normalize(name)] = value
	}
}

// AddHelperClassConst adds a class constant to the helper.
func (i *Interpreter) AddHelperClassConst(helper interface{}, name string, value Value) {
	if h, ok := helper.(*HelperInfo); ok {
		h.ClassConsts[ident.Normalize(name)] = value
	}
}

// RegisterHelperLegacy registers the helper in the legacy i.helpers map.
func (i *Interpreter) RegisterHelperLegacy(typeName string, helper interface{}) {
	if h, ok := helper.(*HelperInfo); ok {
		i.typeSystem.RegisterHelper(typeName, h)
	}
}

// ===== Type System Adapters =====

// WrapInSubrange wraps an integer value in a subrange type with validation.
func (i *Interpreter) WrapInSubrange(value Value, subrangeTypeName string, node ast.Node) (Value, error) {
	subrangeType := i.typeSystem.LookupSubrangeType(subrangeTypeName)
	if subrangeType == nil {
		return nil, fmt.Errorf("subrange type '%s' not found", subrangeTypeName)
	}

	// Extract integer value
	var intValue int
	if intVal, ok := value.(*IntegerValue); ok {
		intValue = int(intVal.Value)
	} else if srcSubrange, ok := value.(*SubrangeValue); ok {
		intValue = srcSubrange.Value
	} else {
		return nil, fmt.Errorf("cannot convert %s to subrange type %s", value.Type(), subrangeTypeName)
	}

	// Create subrange value and validate
	subrangeVal := &SubrangeValue{
		Value:        0, // Will be set by ValidateAndSet
		SubrangeType: subrangeType,
	}
	if err := subrangeVal.ValidateAndSet(intValue); err != nil {
		return nil, err
	}
	return subrangeVal, nil
}

// WrapInInterface wraps an object value in an interface instance.
func (i *Interpreter) WrapInInterface(value Value, interfaceName string, node ast.Node) (Value, error) {
	ifaceInfo := i.lookupInterfaceInfo(interfaceName)
	if ifaceInfo == nil {
		return nil, fmt.Errorf("interface '%s' not found", interfaceName)
	}

	// Check if the value is already an InterfaceInstance
	if _, alreadyInterface := value.(*InterfaceInstance); alreadyInterface {
		return value, nil
	}

	// Check if the value is an ObjectInstance
	objInst, isObj := value.(*ObjectInstance)
	if !isObj {
		return nil, fmt.Errorf("cannot wrap %s in interface %s", value.Type(), interfaceName)
	}

	// Validate that the object's class implements the interface
	concreteClass, ok := objInst.Class.(*ClassInfo)
	if !ok {
		return nil, fmt.Errorf("object has invalid class type")
	}
	if !classImplementsInterface(concreteClass, ifaceInfo) {
		return nil, fmt.Errorf("class '%s' does not implement interface '%s'",
			objInst.Class.GetName(), ifaceInfo.Name)
	}

	// Wrap the object in an InterfaceInstance
	return NewInterfaceInstance(ifaceInfo, objInst), nil
}

// ExecuteRecordPropertyRead executes a record property getter method.
func (i *Interpreter) ExecuteRecordPropertyRead(record Value, propInfoAny any, indices []Value, node any) Value {
	recordVal, ok := record.(*RecordValue)
	if !ok {
		return &ErrorValue{Message: "ExecuteRecordPropertyRead expects RecordValue"}
	}
	propInfo, ok := propInfoAny.(*types.RecordPropertyInfo)
	if !ok {
		return &ErrorValue{Message: "ExecuteRecordPropertyRead expects *types.RecordPropertyInfo"}
	}
	indexExpr, ok := node.(*ast.IndexExpression)
	if !ok {
		return &ErrorValue{Message: "ExecuteRecordPropertyRead expects *ast.IndexExpression"}
	}
	if propInfo.ReadField == "" {
		return i.newErrorWithLocation(indexExpr, "default property is write-only")
	}

	getterMethod := GetRecordMethod(recordVal, propInfo.ReadField)
	if getterMethod == nil {
		return i.newErrorWithLocation(indexExpr, "default property read accessor '%s' is not a method", propInfo.ReadField)
	}

	convertedIndices := make([]Value, len(indices))
	copy(convertedIndices, indices)

	// Create synthetic method call: record.GetterMethod(index)
	methodCall := &ast.MethodCallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: indexExpr.Token},
		},
		Object: indexExpr.Left,
		Method: &ast.Identifier{
			Value: propInfo.ReadField,
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: indexExpr.Token},
			},
		},
		Arguments: make([]ast.Expression, len(indices)),
	}

	// Bind index values as temporary variables
	for idx := range indices {
		tempVarName := fmt.Sprintf("__temp_default_index_%d__", idx)
		methodCall.Arguments[idx] = &ast.Identifier{
			Value: tempVarName,
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: indexExpr.Token},
			},
		}
		i.Env().Define(tempVarName, convertedIndices[idx])
	}
	return i.evalMethodCall(methodCall)
}

// ===== Class Creation Adapters =====

// NewClassInfoAdapter creates a new ClassInfo with the given name.
func (i *Interpreter) NewClassInfoAdapter(name string) interface{} {
	return NewClassInfo(name)
}

// CastToClassInfo attempts to cast interface{} to *ClassInfo.
func (i *Interpreter) CastToClassInfo(class interface{}) (interface{}, bool) {
	ci, ok := class.(*ClassInfo)
	if !ok {
		return nil, false
	}
	return ci, true
}

// IsClassPartial checks if a ClassInfo is marked as partial.
func (i *Interpreter) IsClassPartial(classInfo interface{}) bool {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return false
	}
	return ci.IsPartial
}

// SetClassPartial sets the IsPartial flag on a ClassInfo.
func (i *Interpreter) SetClassPartial(classInfo interface{}, isPartial bool) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}
	ci.IsPartial = isPartial
}

// SetClassAbstract sets the IsAbstract flag on a ClassInfo.
func (i *Interpreter) SetClassAbstract(classInfo interface{}, isAbstract bool) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}
	ci.IsAbstractFlag = isAbstract
}

// SetClassExternal sets the IsExternal flag and ExternalName on a ClassInfo.
func (i *Interpreter) SetClassExternal(classInfo interface{}, isExternal bool, externalName string) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}
	ci.IsExternalFlag = isExternal
	ci.ExternalName = externalName
}

// ClassHasNoParent checks if a ClassInfo has no parent set yet.
func (i *Interpreter) ClassHasNoParent(classInfo interface{}) bool {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return false
	}
	return ci.Parent == nil
}

// DefineCurrentClassMarker defines a marker for the class being declared (for nested types).
func (i *Interpreter) DefineCurrentClassMarker(env interface{}, classInfo interface{}) {
	// Placeholder: nested class resolution to be implemented later
	_ = classInfo
}

// SetClassParent sets the parent class and copies all inherited members.
func (i *Interpreter) SetClassParent(classInfo interface{}, parentClass interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}

	parent, ok := parentClass.(*ClassInfo)
	if !ok {
		return
	}

	if ci.Parent != nil {
		return
	}

	// Set parent references
	ci.Parent = parent
	ci.Metadata.Parent = parent.Metadata
	ci.Metadata.ParentName = parent.Metadata.Name

	// Copy parent fields
	for fieldName, fieldType := range parent.Fields {
		ci.Fields[fieldName] = fieldType
	}
	for fieldName, fieldDecl := range parent.FieldDecls {
		ci.FieldDecls[fieldName] = fieldDecl
	}

	// Copy parent methods (overloads are NOT copied - collected via hierarchy walk)
	for methodName, methodDecl := range parent.Methods {
		ci.Methods[methodName] = methodDecl
	}
	for methodName, methodDecl := range parent.ClassMethods {
		ci.ClassMethods[methodName] = methodDecl
	}

	// Copy constructors
	for name, constructor := range parent.Constructors {
		normalizedName := ident.Normalize(name)
		ci.Constructors[normalizedName] = constructor
	}
	for name, overloads := range parent.ConstructorOverloads {
		normalizedName := ident.Normalize(name)
		ci.ConstructorOverloads[normalizedName] = append([]*ast.FunctionDecl(nil), overloads...)
	}
	if parent.DefaultConstructor != "" {
		ci.DefaultConstructor = parent.DefaultConstructor
	}

	// Copy parent constructor metadata
	for name, constructor := range parent.Metadata.Constructors {
		if ci.Metadata.Constructors == nil {
			ci.Metadata.Constructors = make(map[string]*runtime.MethodMetadata)
		}
		ci.Metadata.Constructors[name] = constructor
	}
	for name, overloads := range parent.Metadata.ConstructorOverloads {
		if ci.Metadata.ConstructorOverloads == nil {
			ci.Metadata.ConstructorOverloads = make(map[string][]*runtime.MethodMetadata)
		}
		ci.Metadata.ConstructorOverloads[name] = append([]*runtime.MethodMetadata(nil), overloads...)
	}
	if parent.Metadata.DefaultConstructor != "" {
		ci.Metadata.DefaultConstructor = parent.Metadata.DefaultConstructor
	}

	ci.Operators = parent.Operators.clone()
}

// AddInterfaceToClass adds an interface to a class's interface list.
func (i *Interpreter) AddInterfaceToClass(classInfo interface{}, interfaceInfo interface{}, interfaceName string) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}

	iface, ok := interfaceInfo.(*InterfaceInfo)
	if !ok {
		return
	}

	ci.Interfaces = append(ci.Interfaces, iface)
	ci.Metadata.Interfaces = append(ci.Metadata.Interfaces, interfaceName)
}

// ===== Class Method, Property, and Operator Adapters =====

// AddClassMethod adds a method to a ClassInfo (handles constructors, overloads, metadata).
func (i *Interpreter) AddClassMethod(classInfo interface{}, method *ast.FunctionDecl, className string) bool {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return false
	}

	normalizedMethodName := ident.Normalize(method.Name.Value)

	// Auto-detect constructors: methods named "Create" returning the class type
	if !method.IsConstructor && ident.Equal(method.Name.Value, "Create") && method.ReturnType != nil {
		returnTypeName := method.ReturnType.String()
		if ident.Equal(returnTypeName, className) {
			method.IsConstructor = true
		}
	}

	methodMeta := runtime.MethodMetadataFromAST(method)
	i.methodRegistry.RegisterMethod(methodMeta)

	// Register as class method or instance method
	if method.IsClassMethod {
		ci.ClassMethods[normalizedMethodName] = method
		ci.ClassMethodOverloads[normalizedMethodName] = append(ci.ClassMethodOverloads[normalizedMethodName], method)
		if !method.IsConstructor && !method.IsDestructor {
			runtime.AddMethodToClass(ci.Metadata, methodMeta, true)
		}
	} else {
		ci.Methods[normalizedMethodName] = method
		ci.MethodOverloads[normalizedMethodName] = append(ci.MethodOverloads[normalizedMethodName], method)
		if !method.IsConstructor && !method.IsDestructor {
			runtime.AddMethodToClass(ci.Metadata, methodMeta, false)
		}
	}

	if method.IsDestructor {
		ci.Metadata.Destructor = methodMeta
	}

	if method.IsConstructor {
		normalizedName := ident.Normalize(method.Name.Value)
		ci.Constructors[normalizedName] = method
		runtime.AddConstructorToClass(ci.Metadata, methodMeta)
		if method.IsDefault {
			ci.DefaultConstructor = method.Name.Value
		}

		// Child constructor with same signature HIDES parent's (DWScript behavior)
		existingOverloads := ci.ConstructorOverloads[normalizedName]
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
		ci.ConstructorOverloads[normalizedName] = existingOverloads
	}

	return true
}

// SynthesizeDefaultConstructor creates implicit parameterless constructor if needed.
func (i *Interpreter) SynthesizeDefaultConstructor(classInfo interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}

	// For each constructor with 'overload' directive, ensure parameterless version exists
	for ctorName, overloads := range ci.ConstructorOverloads {
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
			if _, exists := ci.Constructors[normalizedName]; !exists {
				ci.Constructors[normalizedName] = implicitConstructor
			}
			ci.ConstructorOverloads[normalizedName] = append(
				ci.ConstructorOverloads[normalizedName],
				implicitConstructor,
			)
		}
	}
}

// AddClassProperty adds a property declaration to a ClassInfo.
func (i *Interpreter) AddClassProperty(classInfo interface{}, propDecl *ast.PropertyDecl) bool {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return false
	}

	propInfo := i.convertPropertyDecl(propDecl)
	if propInfo != nil {
		ci.Properties[propDecl.Name.Value] = propInfo
		return true
	}
	return false
}

// RegisterClassOperator registers an operator overload for a class.
func (i *Interpreter) RegisterClassOperator(classInfo interface{}, opDecl *ast.OperatorDecl) Value {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return i.newErrorWithLocation(opDecl, "internal error: invalid class type")
	}

	if opDecl.Binding == nil {
		return i.newErrorWithLocation(opDecl, "class operator '%s' missing binding", opDecl.OperatorSymbol)
	}

	bindingName := opDecl.Binding.Value
	normalizedBindingName := ident.Normalize(bindingName)
	_, isClassMethod := ci.ClassMethods[normalizedBindingName]
	if !isClassMethod {
		_, ok := ci.Methods[normalizedBindingName]
		if !ok {
			return i.newErrorWithLocation(opDecl, "binding '%s' for class operator '%s' not found in class '%s'", bindingName, opDecl.OperatorSymbol, ci.Name)
		}
	}

	// Build operand type list, resolving aliases
	classKey := NormalizeTypeAnnotation(ci.Name)
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
	if !includesClass {
		if ident.Equal(opDecl.OperatorSymbol, "in") {
			operandTypes = append(operandTypes, classKey)
		} else {
			operandTypes = append([]string{classKey}, operandTypes...)
		}
	}

	// Determine self operand index for instance methods
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
		Class:         ci,
		IsClassMethod: isClassMethod,
		SelfIndex:     selfIndex,
	}

	if err := ci.Operators.register(entry); err != nil {
		return i.newErrorWithLocation(opDecl, "class operator '%s' already defined for operand types (%s)", opDecl.OperatorSymbol, strings.Join(operandTypes, ", "))
	}

	return &NilValue{}
}

// LookupClassMethod looks up a method in a ClassInfo by name.
func (i *Interpreter) LookupClassMethod(classInfo interface{}, methodName string, isClassMethod bool) (interface{}, bool) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return nil, false
	}

	normalizedName := ident.Normalize(methodName)
	if isClassMethod {
		method, exists := ci.ClassMethods[normalizedName]
		return method, exists
	}
	method, exists := ci.Methods[normalizedName]
	return method, exists
}

// SetClassConstructor sets the constructor field on a ClassInfo.
func (i *Interpreter) SetClassConstructor(classInfo interface{}, constructor interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}
	ctor, ok := constructor.(*ast.FunctionDecl)
	if !ok {
		return
	}
	ci.Constructor = ctor
}

// SetClassDestructor sets the destructor field on a ClassInfo.
func (i *Interpreter) SetClassDestructor(classInfo interface{}, destructor interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}
	dtor, ok := destructor.(*ast.FunctionDecl)
	if !ok {
		return
	}
	ci.Destructor = dtor
}

// InheritDestructorIfMissing inherits destructor from parent if not locally declared.
func (i *Interpreter) InheritDestructorIfMissing(classInfo interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}
	if ci.Metadata.Destructor == nil && ci.Parent != nil && ci.Parent.Metadata.Destructor != nil {
		ci.Metadata.Destructor = ci.Parent.Metadata.Destructor
	}
}

// InheritParentProperties copies parent properties to child if not already defined.
func (i *Interpreter) InheritParentProperties(classInfo interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}
	if ci.Parent != nil {
		for propName, propInfo := range ci.Parent.Properties {
			if _, exists := ci.Properties[propName]; !exists {
				ci.Properties[propName] = propInfo
			}
		}
	}
}

// ===== VMT and Registration Adapters =====

// BuildVirtualMethodTable builds the virtual method table for a class.
func (i *Interpreter) BuildVirtualMethodTable(classInfo interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}
	ci.buildVirtualMethodTable()
}

// RegisterClassInTypeSystem registers a class in the TypeSystem.
func (i *Interpreter) RegisterClassInTypeSystem(classInfo interface{}, parentName string) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}
	i.typeSystem.RegisterClassWithParent(ci.Name, ci, parentName)
}

// AddClassConstant registers a class constant and its evaluated value.
func (i *Interpreter) AddClassConstant(classInfo interface{}, constDecl *ast.ConstDecl, value Value) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok || constDecl == nil {
		return
	}

	ci.Constants[constDecl.Name.Value] = constDecl
	if val, ok := value.(Value); ok {
		ci.ConstantValues[constDecl.Name.Value] = val
	}

	if ci.Metadata != nil {
		if ci.Metadata.Constants == nil {
			ci.Metadata.Constants = make(map[string]any)
		}
		ci.Metadata.Constants[constDecl.Name.Value] = value
	}
}

// GetClassConstantValues returns a copy of evaluated class constants.
func (i *Interpreter) GetClassConstantValues(classInfo interface{}) map[string]Value {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return nil
	}

	result := make(map[string]Value, len(ci.ConstantValues))
	for name, val := range ci.ConstantValues {
		result[name] = val
	}
	return result
}

// InheritClassConstants copies constants from the parent class if missing.
func (i *Interpreter) InheritClassConstants(classInfo interface{}, parentClass interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}

	parent, ok := parentClass.(*ClassInfo)
	if !ok || parent == nil {
		return
	}

	for name, decl := range parent.Constants {
		if _, exists := ci.Constants[name]; !exists {
			ci.Constants[name] = decl
		}
	}

	for name, val := range parent.ConstantValues {
		if _, exists := ci.ConstantValues[name]; !exists {
			ci.ConstantValues[name] = val
		}
	}

	if parent.Metadata != nil && parent.Metadata.Constants != nil {
		if ci.Metadata.Constants == nil {
			ci.Metadata.Constants = make(map[string]any)
		}
		for name, val := range parent.Metadata.Constants {
			if _, exists := ci.Metadata.Constants[name]; !exists {
				ci.Metadata.Constants[name] = val
			}
		}
	}
}

// AddClassField registers an instance field with its resolved type.
func (i *Interpreter) AddClassField(classInfo interface{}, fieldDecl *ast.FieldDecl, fieldType types.Type) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok || fieldDecl == nil {
		return
	}

	ci.Fields[fieldDecl.Name.Value] = fieldType
	ci.FieldDecls[fieldDecl.Name.Value] = fieldDecl

	fieldMeta := runtime.FieldMetadataFromAST(fieldDecl)
	fieldMeta.Type = fieldType
	runtime.AddFieldToClass(ci.Metadata, fieldMeta)
}

// AddClassVar registers a class variable (static field) with its value.
func (i *Interpreter) AddClassVar(classInfo interface{}, name string, value Value) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}

	if val, ok := value.(Value); ok {
		ci.ClassVars[name] = val
	}

	if ci.Metadata != nil {
		if ci.Metadata.ClassVars == nil {
			ci.Metadata.ClassVars = make(map[string]any)
		}
		ci.Metadata.ClassVars[name] = value
	}
}

// AddNestedClass registers a nested class reference on the parent.
func (i *Interpreter) AddNestedClass(parentClass interface{}, nestedName string, nestedClass interface{}) {
	parent, ok := parentClass.(*ClassInfo)
	if !ok || parent == nil {
		return
	}

	if nested, ok := nestedClass.(*ClassInfo); ok {
		parent.NestedClasses[ident.Normalize(nestedName)] = nested
	}
}

// ===== Helper Property Adapter =====

// EvalBuiltinHelperProperty evaluates a built-in helper property.
func (i *Interpreter) EvalBuiltinHelperProperty(propSpec string, selfValue Value, node ast.Node) Value {
	val, ok := selfValue.(Value)
	if !ok {
		return i.newErrorWithLocation(node, "invalid value type for built-in property")
	}
	return i.evalBuiltinHelperProperty(propSpec, val, node)
}

// ===== Class Property Adapter =====

// EvalClassPropertyRead evaluates a class property read operation.
func (i *Interpreter) EvalClassPropertyRead(classInfoAny any, propInfoAny any, node ast.Node) Value {
	classInfo, ok := classInfoAny.(*ClassInfo)
	if !ok {
		return i.newErrorWithLocation(node, "invalid class info type for class property read")
	}
	propInfo, ok := propInfoAny.(*types.PropertyInfo)
	if !ok {
		return i.newErrorWithLocation(node, "invalid property info type for class property read")
	}
	return i.evalClassPropertyRead(classInfo, propInfo, node)
}

// EvalClassPropertyWrite evaluates a class property write operation.
func (i *Interpreter) EvalClassPropertyWrite(classInfoAny any, propInfoAny any, value Value, node ast.Node) Value {
	classInfo, ok := classInfoAny.(*ClassInfo)
	if !ok {
		return i.newErrorWithLocation(node, "invalid class info type for class property write")
	}
	propInfo, ok := propInfoAny.(*types.PropertyInfo)
	if !ok {
		return i.newErrorWithLocation(node, "invalid property info type for class property write")
	}
	return i.evalClassPropertyWrite(classInfo, propInfo, value, node)
}
