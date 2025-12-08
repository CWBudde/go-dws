package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
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
func (i *Interpreter) AddHelperClassVar(helper interface{}, name string, value evaluator.Value) {
	if h, ok := helper.(*HelperInfo); ok {
		h.ClassVars[ident.Normalize(name)] = value.(Value)
	}
}

// AddHelperClassConst adds a class constant to the helper.
func (i *Interpreter) AddHelperClassConst(helper interface{}, name string, value evaluator.Value) {
	if h, ok := helper.(*HelperInfo); ok {
		h.ClassConsts[ident.Normalize(name)] = value.(Value)
	}
}

// RegisterHelperLegacy registers the helper in the legacy i.helpers map.
func (i *Interpreter) RegisterHelperLegacy(typeName string, helper interface{}) {
	if i.helpers == nil {
		i.helpers = make(map[string][]*HelperInfo)
	}
	if h, ok := helper.(*HelperInfo); ok {
		i.helpers[typeName] = append(i.helpers[typeName], h)
	}
}

// ===== Type System Adapters =====

// WrapInSubrange wraps an integer value in a subrange type with validation.
func (i *Interpreter) WrapInSubrange(value evaluator.Value, subrangeTypeName string, node ast.Node) (evaluator.Value, error) {
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
func (i *Interpreter) WrapInInterface(value evaluator.Value, interfaceName string, node ast.Node) (evaluator.Value, error) {
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
func (i *Interpreter) ExecuteRecordPropertyRead(record evaluator.Value, propInfoAny any, indices []evaluator.Value, node any) evaluator.Value {
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
		i.env.Define(tempVarName, convertedIndices[idx])
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
// Task 3.5.8: Phase 2 adapter for external class marking.
func (i *Interpreter) SetClassExternal(classInfo interface{}, isExternal bool, externalName string) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}
	ci.IsExternalFlag = isExternal
	ci.ExternalName = externalName
}

// ClassHasNoParent checks if a ClassInfo has no parent set yet.
// Task 3.5.8: Phase 3 adapter for parent inheritance check.
// Returns true if the class has no parent, false if it already has a parent.
func (i *Interpreter) ClassHasNoParent(classInfo interface{}) bool {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return false
	}
	return ci.Parent == nil
}

// DefineCurrentClassMarker defines a marker in the environment for the class being declared.
// Task 3.5.8: Phase 2.2 adapter for nested class context setup.
// This enables nested type resolution to reference the enclosing class.
// The marker is stored with a special key that won't conflict with user variables.
func (i *Interpreter) DefineCurrentClassMarker(env interface{}, classInfo interface{}) {
	// The evaluator passes an evaluator.Environment, we need to adapt it
	// For now, we can define a special marker variable
	// This matches the old implementation in declarations.go:236-237
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}

	// Note: The environment type from evaluator needs to be converted
	// For simplicity, we can skip this marker for now since it's primarily
	// used for nested class resolution which will be handled in Phase 5.
	// The marker would be: env.Define("__current_class__", ci)
	// But since env is evaluator.Environment interface, we can't call Define directly.
	// This will be properly implemented when we handle nested classes in Phase 5.
	_ = ci // Prevent unused variable warning
}

// SetClassParent sets the parent class and copies all inherited members.
// This replicates the logic from declarations.go:287-351.
func (i *Interpreter) SetClassParent(classInfo interface{}, parentClass interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}

	parent, ok := parentClass.(*ClassInfo)
	if !ok {
		return
	}

	// Only set parent if not already set
	if ci.Parent != nil {
		return
	}

	// Set parent references
	ci.Parent = parent
	ci.Metadata.Parent = parent.Metadata
	ci.Metadata.ParentName = parent.Metadata.Name

	// Copy parent fields (child inherits all parent fields)
	for fieldName, fieldType := range parent.Fields {
		ci.Fields[fieldName] = fieldType
	}

	// Copy parent field declarations (for initializers)
	for fieldName, fieldDecl := range parent.FieldDecls {
		ci.FieldDecls[fieldName] = fieldDecl
	}

	// Copy parent methods (child inherits all parent methods)
	// Keep Methods and ClassMethods for backward compatibility (direct lookups)
	for methodName, methodDecl := range parent.Methods {
		ci.Methods[methodName] = methodDecl
	}
	for methodName, methodDecl := range parent.ClassMethods {
		ci.ClassMethods[methodName] = methodDecl
	}

	// DON'T copy MethodOverloads/ClassMethodOverloads from parent
	// Each class should only store its OWN method overloads, not inherited ones.
	// getMethodOverloadsInHierarchy will walk the hierarchy to collect them at call time.
	// This prevents duplication when a child class overrides a parent method.

	// Copy constructors
	for name, constructor := range parent.Constructors {
		normalizedName := ident.Normalize(name)
		ci.Constructors[normalizedName] = constructor
	}
	for name, overloads := range parent.ConstructorOverloads {
		normalizedName := ident.Normalize(name)
		ci.ConstructorOverloads[normalizedName] = append([]*ast.FunctionDecl(nil), overloads...)
	}

	// Inherit default constructor if parent has one
	if parent.DefaultConstructor != "" {
		ci.DefaultConstructor = parent.DefaultConstructor
	}

	// Copy parent constructors to metadata
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

	// Inherit default constructor name into metadata
	if parent.Metadata.DefaultConstructor != "" {
		ci.Metadata.DefaultConstructor = parent.Metadata.DefaultConstructor
	}

	// Copy operator overloads
	ci.Operators = parent.Operators.clone()
}

// AddInterfaceToClass adds an interface to a class's interface list.
// This updates both the ClassInfo.Interfaces slice and Metadata.Interfaces.
func (i *Interpreter) AddInterfaceToClass(classInfo interface{}, interfaceInfo interface{}, interfaceName string) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}

	iface, ok := interfaceInfo.(*InterfaceInfo)
	if !ok {
		return
	}

	// Add interface to class's interface list
	ci.Interfaces = append(ci.Interfaces, iface)

	// Add interface name to metadata
	ci.Metadata.Interfaces = append(ci.Metadata.Interfaces, interfaceName)
}

// ===== Task 3.5.8 Phase 6: Method, Property, and Operator Adapters =====

// AddClassMethod adds a method declaration to a ClassInfo.
// Task 3.5.8 Phase 6: Migrated from declarations.go:556-637 method processing loop.
// This handles method registration, constructor detection, overload handling, and metadata creation.
func (i *Interpreter) AddClassMethod(classInfo interface{}, method *ast.FunctionDecl, className string) bool {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return false
	}

	// Normalize method name for case-insensitive lookup
	normalizedMethodName := ident.Normalize(method.Name.Value)

	// Auto-detect constructors: methods named "Create" that return the class type
	// This handles inline constructor declarations like: function Create(...): TClass;
	// Matches semantic analyzer behavior (analyze_classes_decl.go:576-580)
	if !method.IsConstructor && ident.Equal(method.Name.Value, "Create") && method.ReturnType != nil {
		returnTypeName := method.ReturnType.String()
		if ident.Equal(returnTypeName, className) {
			method.IsConstructor = true
		}
	}

	// Create MethodMetadata once for this method
	methodMeta := runtime.MethodMetadataFromAST(method)
	i.methodRegistry.RegisterMethod(methodMeta)

	// Check if this is a class method (static method) or instance method
	if method.IsClassMethod {
		// Store in ClassMethods map
		ci.ClassMethods[normalizedMethodName] = method
		// Add to overload list
		ci.ClassMethodOverloads[normalizedMethodName] = append(ci.ClassMethodOverloads[normalizedMethodName], method)

		// Add to metadata (unless it's a constructor/destructor - those go separately)
		if !method.IsConstructor && !method.IsDestructor {
			runtime.AddMethodToClass(ci.Metadata, methodMeta, true)
		}
	} else {
		// Store in instance Methods map
		ci.Methods[normalizedMethodName] = method
		// Add to overload list
		ci.MethodOverloads[normalizedMethodName] = append(ci.MethodOverloads[normalizedMethodName], method)

		// Add to metadata (unless it's a constructor/destructor - those go separately)
		if !method.IsConstructor && !method.IsDestructor {
			runtime.AddMethodToClass(ci.Metadata, methodMeta, false)
		}
	}

	// Handle destructor
	if method.IsDestructor {
		ci.Metadata.Destructor = methodMeta
	}

	// Handle constructor
	if method.IsConstructor {
		normalizedName := ident.Normalize(method.Name.Value)
		ci.Constructors[normalizedName] = method

		// Add constructor to metadata (reuse methodMeta)
		runtime.AddConstructorToClass(ci.Metadata, methodMeta)

		// Capture default constructor
		if method.IsDefault {
			ci.DefaultConstructor = method.Name.Value
		}

		// In DWScript, a child constructor with the same name and signature HIDES the parent's,
		// regardless of whether it has the `override` keyword or not
		existingOverloads := ci.ConstructorOverloads[normalizedName]
		replaced := false
		for idx, existingMethod := range existingOverloads {
			// Check if signatures match (same number and types of parameters)
			if parametersMatch(existingMethod.Parameters, method.Parameters) {
				// Replace the parent constructor with this child constructor (hiding)
				existingOverloads[idx] = method
				replaced = true
				break
			}
		}
		if !replaced {
			// No matching parent constructor found (different signature), just append
			existingOverloads = append(existingOverloads, method)
		}
		// Write the modified slice back to the map
		ci.ConstructorOverloads[normalizedName] = existingOverloads
	}

	return true
}

// Task 3.5.27: CreateMethodMetadata REMOVED - zero callers

// SynthesizeDefaultConstructor synthesizes an implicit parameterless constructor.
// Task 3.5.8 Phase 6: Migrated from declarations.go:880-923.
// For each constructor name, if it has the 'overload' directive but no parameterless overload,
// synthesize one. This matches DWScript behavior.
func (i *Interpreter) SynthesizeDefaultConstructor(classInfo interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}

	// For each constructor name, check if it has the 'overload' directive
	// If so, ensure there's a parameterless overload
	for ctorName, overloads := range ci.ConstructorOverloads {
		hasOverloadDirective := false
		hasParameterlessOverload := false

		// Check if any overload has the 'overload' directive
		// and if a parameterless overload already exists
		for _, ctor := range overloads {
			if ctor.IsOverload {
				hasOverloadDirective = true
			}
			if len(ctor.Parameters) == 0 {
				hasParameterlessOverload = true
			}
		}

		// If this constructor set has 'overload' but no parameterless version, synthesize one
		if hasOverloadDirective && !hasParameterlessOverload {
			// Create a minimal constructor AST node (just for runtime - no actual body needed)
			// The interpreter will initialize fields with default values when no constructor body exists
			implicitConstructor := &ast.FunctionDecl{
				BaseNode:      ast.BaseNode{},
				Name:          &ast.Identifier{Value: ctorName},
				Parameters:    []*ast.Parameter{}, // No parameters
				ReturnType:    nil,                // Constructors don't have explicit return types
				Body:          nil,                // No body - just field initialization
				IsConstructor: true,
				IsOverload:    true,
			}

			// Add to class constructor maps
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
// Task 3.5.8 Phase 6: Migrated from declarations.go:690-710 property registration.
func (i *Interpreter) AddClassProperty(classInfo interface{}, propDecl *ast.PropertyDecl) bool {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return false
	}

	// Convert AST property to PropertyInfo
	propInfo := i.convertPropertyDecl(propDecl)
	if propInfo != nil {
		ci.Properties[propDecl.Name.Value] = propInfo
		return true
	}
	return false
}

// RegisterClassOperator registers an operator overload for a class.
// Task 3.5.8 Phase 6: Migrated from declarations.go:976-1047 registerClassOperator.
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

	classKey := NormalizeTypeAnnotation(ci.Name)
	operandTypes := make([]string, 0, len(opDecl.OperandTypes)+1)
	includesClass := false
	for _, operand := range opDecl.OperandTypes {
		// Resolve type aliases before normalizing
		// This ensures that "toa" (alias for "array of const") is resolved to "ARRAY OF CONST"
		typeName := operand.String()
		resolvedType, err := i.resolveType(typeName)
		var key string
		if err == nil {
			// Successfully resolved - use the resolved type's string representation
			key = NormalizeTypeAnnotation(resolvedType.String())
		} else {
			// Failed to resolve - use the raw type name (might be a forward reference)
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
// Task 3.5.8 Phase 6: Helper for identifying constructors/destructors.
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

// SetClassConstructor sets the constructor field on a ClassInfo (legacy behavior).
// Task 3.5.8 Phase 6: Maintains backward compatibility with old implementation.
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

// SetClassDestructor sets the destructor field on a ClassInfo (legacy behavior).
// Task 3.5.8 Phase 6: Maintains backward compatibility with old implementation.
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

// InheritDestructorIfMissing inherits destructor from parent if no local destructor declared.
// Task 3.5.8 Phase 6: Migrated from declarations.go:680-683.
func (i *Interpreter) InheritDestructorIfMissing(classInfo interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}

	// Inherit destructor from parent if no local destructor declared
	if ci.Metadata.Destructor == nil && ci.Parent != nil && ci.Parent.Metadata.Destructor != nil {
		ci.Metadata.Destructor = ci.Parent.Metadata.Destructor
	}
}

// InheritParentProperties copies parent properties to child class if not already defined.
// Task 3.5.8 Phase 6: Migrated from declarations.go:702-710.
func (i *Interpreter) InheritParentProperties(classInfo interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}

	// Copy parent properties (child inherits all parent properties)
	if ci.Parent != nil {
		for propName, propInfo := range ci.Parent.Properties {
			// Only copy if not already defined in child class
			if _, exists := ci.Properties[propName]; !exists {
				ci.Properties[propName] = propInfo
			}
		}
	}
}

// ===== Task 3.5.8 Phase 7: VMT and Registration Adapters =====

// BuildVirtualMethodTable builds the virtual method table for a class.
// Task 3.5.8 Phase 7: Delegates to existing ClassInfo.buildVirtualMethodTable().
// This method implements proper virtual/override/reintroduce semantics.
func (i *Interpreter) BuildVirtualMethodTable(classInfo interface{}) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}
	ci.buildVirtualMethodTable()
}

// RegisterClassInTypeSystem registers a class in the TypeSystem after VMT is built.
// Task 3.5.8 Phase 7: Uses TypeSystem.RegisterClassWithParent() for proper hierarchy tracking.
func (i *Interpreter) RegisterClassInTypeSystem(classInfo interface{}, parentName string) {
	ci, ok := classInfo.(*ClassInfo)
	if !ok {
		return
	}
	i.typeSystem.RegisterClassWithParent(ci.Name, ci, parentName)
}

// ===== Task 3.5.37: Helper Property Adapter =====

// EvalBuiltinHelperProperty evaluates a built-in helper property.
// Task 3.5.37: Adapter method delegating to interpreter's evalBuiltinHelperProperty.
// This is called from the evaluator for built-in properties that require interpreter access.
func (i *Interpreter) EvalBuiltinHelperProperty(propSpec string, selfValue evaluator.Value, node ast.Node) evaluator.Value {
	// Convert evaluator.Value to Value for interpreter's method
	val, ok := selfValue.(Value)
	if !ok {
		return i.newErrorWithLocation(node, "invalid value type for built-in property")
	}
	return i.evalBuiltinHelperProperty(propSpec, val, node)
}
