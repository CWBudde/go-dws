package runtime

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ClassVirtualMethodEntry tracks virtual method dispatch information.
type ClassVirtualMethodEntry = VirtualMethodEntry

// ClassInfo represents runtime class metadata.
// It stores information about a class's structure including fields, methods,
// parent class, and constructor/destructor.
type ClassInfo struct {
	// ClassMetadata owns member and dispatch metadata; Metadata names the same object.
	*ClassMetadata

	// Type is the resolved semantic identity shared with declaration analysis.
	Type *types.ClassType

	// Metadata contains runtime declaration metadata for this class.
	Metadata *ClassMetadata

	// Class identity and hierarchy links remain separate from member metadata.
	Parent             *ClassInfo
	NestedClasses      map[string]*ClassInfo
	ExternalName       string
	Name               string
	DefaultConstructor string
	Interfaces         []*MutableInterfaceInfo
	typeShared         bool
	IsExternalFlag     bool // Renamed to avoid conflict with IsExternal() method
	IsAbstractFlag     bool // Renamed to avoid conflict with IsAbstract() method
	IsPartial          bool
	// IsForwardDecl marks a class registered only via a forward declaration
	// ("type TFoo = class;"). It is cleared when the full definition completes it.
	IsForwardDecl bool
}

// NewClassInfo creates a new ClassInfo with the given name.
// Fields, Methods, ClassVars, ClassMethods, and Properties maps are initialized as empty.
func NewClassInfo(name string) *ClassInfo {
	metadata := NewClassMetadata(name)
	return &ClassInfo{
		ClassMetadata: metadata,
		Type:          types.NewClassType(name, nil),
		Name:          name,
		Parent:        nil,
		Metadata:      metadata,
		NestedClasses: make(map[string]*ClassInfo),
	}
}

// Compile-time assertion that ClassInfo implements IClassInfo
var _ IClassInfo = (*ClassInfo)(nil)

// === IClassInfo Interface Implementation ===
// These methods expose runtime class data to object instances and evaluators.

// SetResolvedType adopts the analyzer's immutable class identity. Runtime
// inheritance binding must not modify this shared object on subsequent runs.
func (c *ClassInfo) SetResolvedType(classType *types.ClassType) {
	if c == nil || classType == nil {
		return
	}
	c.Type = classType
	c.typeShared = true
}

// GetClassType returns this class's resolved semantic identity.
func (c *ClassInfo) GetClassType() *types.ClassType {
	if c == nil {
		return nil
	}
	return c.Type
}

// GetName returns the class name
func (c *ClassInfo) GetName() string {
	if c == nil {
		return ""
	}
	return c.Name
}

// GetParent returns the parent class metadata
func (c *ClassInfo) GetParent() IClassInfo {
	if c == nil || c.Parent == nil {
		return nil
	}
	return c.Parent
}

// GetDefaultConstructor returns the name of the constructor declared with the
// 'default' directive on this class (empty if none was declared here).
func (c *ClassInfo) GetDefaultConstructor() string {
	if c == nil {
		return ""
	}
	return c.DefaultConstructor
}

// GetMetadata returns the AST-free metadata
func (c *ClassInfo) GetMetadata() *ClassMetadata {
	if c == nil {
		return nil
	}
	return c.Metadata
}

// LookupMethod finds a method by name in the class hierarchy
func (c *ClassInfo) LookupMethod(name string) *MethodMetadata {
	return c.lookupMethod(name)
}

// LookupClassMethod finds a class/static method by name in the class hierarchy.
func (c *ClassInfo) LookupClassMethod(name string) *MethodMetadata {
	if c == nil {
		return nil
	}

	normalizedName := ident.Normalize(name)
	for current := c; current != nil; current = current.Parent {
		if method, exists := current.ClassMethods[normalizedName]; exists {
			return method
		}
		if overloads, exists := current.ClassMethodOverloads[normalizedName]; exists && len(overloads) > 0 {
			return overloads[0]
		}
	}

	return nil
}

// LookupProperty finds a property by name in the class hierarchy
func (c *ClassInfo) LookupProperty(name string) *PropertyInfo {
	propInfo := c.lookupProperty(name)
	if propInfo == nil {
		return nil
	}
	return &PropertyInfo{
		Name:      propInfo.Name,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		ReadSpec:  propInfo.ReadSpec,
		WriteSpec: propInfo.WriteSpec,
		Impl:      propInfo,
	}
}

// GetDefaultProperty returns the default property
func (c *ClassInfo) GetDefaultProperty() *PropertyInfo {
	propInfo := c.getDefaultProperty()
	if propInfo == nil {
		return nil
	}
	return &PropertyInfo{
		Name:      propInfo.Name,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		ReadSpec:  propInfo.ReadSpec,
		WriteSpec: propInfo.WriteSpec,
		Impl:      propInfo,
	}
}

// GetOwnProperties returns the properties declared directly on this class
// (not inherited), each wrapped like LookupProperty. The order is
// unspecified (Go map iteration); callers that need a stable order must sort.
// Used by JSON serialization to enumerate a class level's own properties.
func (c *ClassInfo) GetOwnProperties() []*PropertyInfo {
	if c == nil || len(c.Properties) == 0 {
		return nil
	}
	props := make([]*PropertyInfo, 0, len(c.Properties))
	for name, propInfo := range c.Properties {
		// InheritParentPropertyInfos copies each parent property (same pointer)
		// into the child's map; skip those so this returns only properties
		// actually declared (or overridden) on this class level. This keeps
		// JSON serialization ordering correct (most-derived own members first).
		if c.Parent != nil {
			if parentProp, ok := c.Parent.Properties[name]; ok && parentProp == propInfo {
				continue
			}
		}
		props = append(props, &PropertyInfo{
			Name:      propInfo.Name,
			IsIndexed: propInfo.IsIndexed,
			IsDefault: propInfo.IsDefault,
			ReadSpec:  propInfo.ReadSpec,
			WriteSpec: propInfo.WriteSpec,
			Impl:      propInfo,
		})
	}
	return props
}

// FieldExists checks if a field exists
func (c *ClassInfo) FieldExists(name string) bool {
	for current := c; current != nil; current = current.Parent {
		if current.ClassMetadata == nil {
			continue
		}
		for field := range current.Fields {
			if ident.Equal(field, name) {
				return true
			}
		}
	}
	return false
}

// GetFieldsMap returns canonical field metadata owned by this class.
func (c *ClassInfo) GetFieldsMap() map[string]*FieldMetadata {
	if c == nil || c.ClassMetadata == nil {
		return nil
	}
	return c.Fields
}

// GetMethodsMap returns canonical callable metadata owned by this class.
func (c *ClassInfo) GetMethodsMap() map[string]*MethodMetadata {
	if c == nil || c.ClassMetadata == nil {
		return nil
	}
	return c.Methods
}

// LookupClassVar retrieves a class variable by name
func (c *ClassInfo) LookupClassVar(name string) (Value, IClassInfo) {
	value, owningClass := c.lookupClassVar(name)
	if owningClass == nil {
		return nil, nil
	}
	return value, owningClass
}

// GetClassVarsMap returns the map of class variables
func (c *ClassInfo) GetClassVarsMap() map[string]Value {
	if c == nil {
		return nil
	}
	return c.ClassVars
}

// GetVirtualMethodTable returns the virtual method dispatch table
func (c *ClassInfo) GetVirtualMethodTable() map[string]*VirtualMethodEntry {
	if c == nil || c.ClassMetadata == nil {
		return nil
	}
	return c.VirtualMethods
}

// LookupOperator finds an operator overload
// Note: This method doesn't support inheritance checking because it can't access typeSystem
// (interface constraint). Use lookupOperator directly when inheritance checking is needed.
func (c *ClassInfo) LookupOperator(operator string, operandTypes []types.Type) (*OperatorEntry, bool) {
	if c == nil || c.Operators == nil {
		return nil, false
	}
	// Use nil typeSystem for exact match only (no inheritance checking)
	entry, found := c.lookupOperator(operator, operandTypes)
	if !found || entry == nil {
		return nil, false
	}
	// Note: OperatorEntry is currently a placeholder.
	// Operator lookup is only used within interp package, so returning
	// a simplified version here. Full implementation when needed.
	return &OperatorEntry{
		Operator:      entry.Operator,
		OperandTypes:  entry.OperandTypes,
		Class:         entry.Class,
		Method:        nil, // Not needed, BindingName is used instead
		BindingName:   entry.BindingName,
		SelfIndex:     entry.SelfIndex,
		IsClassMethod: entry.IsClassMethod,
	}, true
}

// GetInterfaces returns the list of interfaces this class implements
func (c *ClassInfo) GetInterfaces() []*InterfaceInfo {
	if c == nil {
		return nil
	}
	// For now, return nil as InterfaceInfo interface is not fully defined
	// This will need proper implementation when interface support is added
	return nil
}

// IsAbstract returns true if this class is declared as abstract.
func (c *ClassInfo) IsAbstract() bool {
	if c == nil {
		return false
	}
	return c.IsAbstractFlag
}

// IsExternal returns true if this class is declared as external.
func (c *ClassInfo) IsExternal() bool {
	if c == nil {
		return false
	}
	return c.IsExternalFlag
}

// GetConstructor returns a constructor declaration by name (case-insensitive).
func (c *ClassInfo) GetConstructor(name string) *MethodMetadata {
	if c == nil {
		return nil
	}
	normalizedName := ident.Normalize(name)
	if ctor, ok := c.Constructors[normalizedName]; ok {
		return ctor
	}
	return nil
}

// HasMethodOverloads returns true if the class or any ancestor declares more than
// one instance method with the given name (i.e., the method is overloaded).
func (c *ClassInfo) HasMethodOverloads(name string) bool { return len(c.GetMethodOverloads(name)) > 1 }

// HasClassMethodOverloads returns true if the class or any ancestor declares more
// than one class (static) method with the given name.
func (c *ClassInfo) HasClassMethodOverloads(name string) bool {
	normalizedName := ident.Normalize(name)
	total := 0
	for current := c; current != nil; current = current.Parent {
		total += len(current.ClassMethodOverloads[normalizedName])
		if total > 1 {
			return true
		}
	}
	return false
}

// GetMethodOverloads returns all instance method overloads across the class hierarchy.
func (c *ClassInfo) GetMethodOverloads(name string) []*MethodMetadata {
	normalizedName := ident.Normalize(name)
	var result []*MethodMetadata
	seen := make(map[*MethodMetadata]bool)
	add := func(methods []*MethodMetadata) {
		for _, method := range methods {
			if method != nil && !seen[method] {
				seen[method] = true
				result = append(result, method)
			}
		}
	}
	for current := c; current != nil; current = current.Parent {
		add(current.MethodOverloads[normalizedName])
		add(current.ConstructorOverloads[normalizedName])
	}
	return result
}

// GetClassMethodOverloads returns all class (static) method overloads across the class hierarchy.
func (c *ClassInfo) GetClassMethodOverloads(name string) []*MethodMetadata {
	normalizedName := ident.Normalize(name)
	var result []*MethodMetadata
	for current := c; current != nil; current = current.Parent {
		result = append(result, current.ClassMethodOverloads[normalizedName]...)
	}
	return result
}

// GetConstructorOverloads returns all constructor overloads across the class hierarchy.
func (c *ClassInfo) GetConstructorOverloads(name string) []*MethodMetadata {
	normalizedName := ident.Normalize(name)
	var result []*MethodMetadata
	for current := c; current != nil; current = current.Parent {
		result = append(result, current.ConstructorOverloads[normalizedName]...)
	}
	return result
}

// OwnsMethodDecl reports whether this class (not an ancestor) declares the
// given method/constructor/destructor declaration. Used to determine the
// defining class of an executing method so `inherited` resolves relative to
// the declaration site, not the receiver's dynamic class.
func (c *ClassInfo) OwnsMethodDecl(fn *ast.FunctionDecl) bool {
	if c == nil || fn == nil {
		return false
	}
	for _, overloads := range c.MethodOverloads {
		for _, decl := range overloads {
			if MethodDeclaration(decl) == fn {
				return true
			}
		}
	}
	for _, overloads := range c.ClassMethodOverloads {
		for _, decl := range overloads {
			if MethodDeclaration(decl) == fn {
				return true
			}
		}
	}
	for _, overloads := range c.ConstructorOverloads {
		for _, decl := range overloads {
			if MethodDeclaration(decl) == fn {
				return true
			}
		}
	}
	for _, decl := range c.Methods {
		if MethodDeclaration(decl) == fn {
			return true
		}
	}
	for _, decl := range c.ClassMethods {
		if MethodDeclaration(decl) == fn {
			return true
		}
	}
	for _, decl := range c.Constructors {
		if MethodDeclaration(decl) == fn {
			return true
		}
	}
	return MethodDeclaration(c.Destructor) == fn
}

// === End IClassInfo Interface Implementation ===

// LookupNestedClass returns a nested class by short name (case-insensitive).
func (c *ClassInfo) LookupNestedClass(name string) *ClassInfo {
	if c == nil {
		return nil
	}
	if nested, ok := c.NestedClasses[ident.Normalize(name)]; ok {
		return nested
	}
	return nil
}

// lookupMethod searches for a method in the class hierarchy.
// Walks up the parent chain, returning the first method found or nil.
func (c *ClassInfo) lookupMethod(name string) *MethodMetadata {
	normalizedName := ident.Normalize(name)

	if method, exists := c.Methods[normalizedName]; exists {
		return method
	}

	// Fallback to overload list when Methods map is empty or incomplete.
	if overloads, exists := c.MethodOverloads[normalizedName]; exists && len(overloads) > 0 {
		return overloads[0]
	}

	if constructor := c.Constructors[normalizedName]; constructor != nil {
		return constructor
	}
	if c.Destructor != nil && ident.Equal(c.Destructor.Name, name) {
		return c.Destructor
	}

	// Check parent class (recursive)
	if c.Parent != nil {
		return c.Parent.lookupMethod(name)
	}

	// Not found
	return nil
}

// PR #147: Removed lookupMethodWithVirtualDispatch - it was never called.
// Virtual dispatch is implemented inline in objects_methods.go.

// lookupProperty searches for a property in the class hierarchy.
// It starts with the current class and walks up the parent chain.
// Returns the first property found, or nil if not found.
func (c *ClassInfo) lookupProperty(name string) *types.PropertyInfo {
	// Check current class with case-insensitive match
	for propName, prop := range c.Properties {
		if ident.Equal(propName, name) {
			return prop
		}
	}

	// Check parent class (recursive)
	if c.Parent != nil {
		return c.Parent.lookupProperty(name)
	}

	// Not found
	return nil
}

// lookupClassVar searches for a class variable in the class hierarchy.
// It starts with the current class and walks up the parent chain.
// Returns the class variable value and the ClassInfo that owns it, or (nil, nil) if not found.
func (c *ClassInfo) lookupClassVar(name string) (Value, *ClassInfo) {
	// Check current class with case-insensitive match
	for varName, value := range c.ClassVars {
		if ident.Equal(varName, name) {
			return value, c
		}
	}

	// Check parent class (recursive)
	if c.Parent != nil {
		return c.Parent.lookupClassVar(name)
	}

	// Not found
	return nil, nil
}

// setClassVar sets a class variable value in the class hierarchy.
// Returns true if found and set, false otherwise.
func (c *ClassInfo) setClassVar(name string, value Value) bool {
	// Check current class with case-insensitive match
	for varName := range c.ClassVars {
		if ident.Equal(varName, name) {
			c.ClassVars[varName] = value
			return true
		}
	}

	// Check parent class (recursive)
	if c.Parent != nil {
		return c.Parent.setClassVar(name, value)
	}

	// Not found
	return false
}

// hasClassVar checks if a class variable exists in the class hierarchy.
func (c *ClassInfo) hasClassVar(name string) bool {
	_, owningClass := c.lookupClassVar(name)
	return owningClass != nil
}

// getDefaultProperty searches for the default property in the class hierarchy.
// Returns the default property if found, or nil if no default property exists.
func (c *ClassInfo) getDefaultProperty() *types.PropertyInfo {
	// Check current class
	for _, prop := range c.Properties {
		if prop.IsDefault {
			return prop
		}
	}

	// Check parent class (recursive)
	if c.Parent != nil {
		return c.Parent.getDefaultProperty()
	}

	// Not found
	return nil
}

// lookupOperator searches class operators using assignment-compatible operand types.
func (c *ClassInfo) lookupOperator(operator string, operandTypes []types.Type) (*ClassOperatorEntry, bool) {
	var entries []*ClassOperatorEntry
	for current := c; current != nil; current = current.Parent {
		if current.Operators != nil {
			entries = append(entries, current.Operators.entries[ident.Normalize(operator)]...)
		}
	}
	return types.SelectOperatorOverload(operandTypes, entries, func(entry *ClassOperatorEntry) []types.Type { return entry.OperandTypes })
}

// HasConstructor checks whether the class or its ancestors declare a constructor with the given name.
func (c *ClassInfo) HasConstructor(name string) bool {
	if c == nil {
		return false
	}
	// Case-insensitive search through constructors
	for ctorName := range c.Constructors {
		if ident.Equal(ctorName, name) {
			return true
		}
	}
	// Also check constructor overloads (case-insensitive)
	for ctorName, overloads := range c.ConstructorOverloads {
		if ident.Equal(ctorName, name) && len(overloads) > 0 {
			return true
		}
	}
	if c.Parent != nil {
		return c.Parent.HasConstructor(name)
	}
	return false
}

// InheritsFrom reports whether the class or any of its ancestors has the given name.
// The check is case-sensitive and includes the class itself.
func (c *ClassInfo) InheritsFrom(name string) bool {
	for current := c; current != nil; current = current.Parent {
		if current.Name == name {
			return true
		}
	}
	return false
}

// ============================================================================
// ClassValue - Metaclass Runtime Value
// ============================================================================

// ClassValue represents a class reference (metaclass value) at
// In DWScript, when you write "TMyClass" in an expression, it represents
// a reference to the class type itself, not an instance.
//
// Example usage:
//
//	var cls: class of TAnimal;
//	cls := TDog;              // Assign class reference
//	obj := cls.Create;        // Call constructor through metaclass
type ClassValue struct {
	// ClassInfo points to the class metadata
	ClassInfo *ClassInfo
}

// Type returns "CLASS" to indicate this is a class reference.
func (c *ClassValue) Type() string {
	return "CLASS"
}

// String returns a string representation of the class reference.
// Format: "class TClassName"
func (c *ClassValue) String() string {
	if c.ClassInfo != nil {
		return fmt.Sprintf("class %s", c.ClassInfo.Name)
	}
	return "class <nil>"
}

// GetClassName returns the class name.
func (c *ClassValue) GetClassName() string {
	if c == nil || c.ClassInfo == nil {
		return ""
	}
	return c.ClassInfo.Name
}

// GetClassType returns the class type (metaclass) as a ClassValue.
// For ClassValue, this returns itself.
func (c *ClassValue) GetClassType() Value {
	return c
}

// GetClassVar retrieves a class variable value by name from the class hierarchy.
func (c *ClassValue) GetClassVar(name string) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	value, owningClass := c.ClassInfo.lookupClassVar(name)
	if owningClass == nil {
		return nil, false
	}
	return value, true
}

// GetClassConstant retrieves a class constant value by name from the class hierarchy.
func (c *ClassValue) GetClassConstant(name string) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Check ConstantValues cache first (case-insensitive)
	for constName, value := range c.ClassInfo.Constants {
		if ident.Equal(constName, name) {
			return value, true
		}
	}
	// Check parent class hierarchy
	if c.ClassInfo.Parent != nil {
		parentCV := &ClassValue{ClassInfo: c.ClassInfo.Parent}
		return parentCV.GetClassConstant(name)
	}
	return nil, false
}

// HasClassMethod checks if a class method with the given name exists.
func (c *ClassValue) HasClassMethod(name string) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	normalizedName := ident.Normalize(name)
	// Check single class methods
	if _, exists := c.ClassInfo.ClassMethods[normalizedName]; exists {
		return true
	}
	// Check overloaded class methods
	if overloads, exists := c.ClassInfo.ClassMethodOverloads[normalizedName]; exists && len(overloads) > 0 {
		return true
	}
	// Check parent class hierarchy
	if c.ClassInfo.Parent != nil {
		parentCV := &ClassValue{ClassInfo: c.ClassInfo.Parent}
		return parentCV.HasClassMethod(name)
	}
	return false
}

// HasConstructor checks if a constructor with the given name exists.
func (c *ClassValue) HasConstructor(name string) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	return c.ClassInfo.HasConstructor(name)
}

// InvokeParameterlessClassMethod invokes a parameterless class method.
func (c *ClassValue) InvokeParameterlessClassMethod(name string, executor func(methodDecl *MethodMetadata) Value) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Look up class method in hierarchy
	normalizedName := ident.Normalize(name)
	for current := c.ClassInfo; current != nil; current = current.Parent {
		if method, exists := current.ClassMethods[normalizedName]; exists {
			if len(method.Parameters) == 0 {
				return executor(method), true
			}
			return nil, false // Has parameters
		}
		if overloads, exists := current.ClassMethodOverloads[normalizedName]; exists && len(overloads) > 0 {
			// Check if any overload is parameterless
			for _, m := range overloads {
				if len(m.Parameters) == 0 {
					return executor(m), true
				}
			}
			return nil, false // No parameterless overload
		}
	}
	return nil, false
}

// CreateClassMethodPointer creates a function pointer for a class method.
// A parameterless class method yields a pointer too when this is called (the
// auto-invoke path tries InvokeParameterlessClassMethod first, so this is only
// reached for a parameterless method when a pointer is explicitly wanted, e.g.
// p := TClass.ClassProc).
func (c *ClassValue) CreateClassMethodPointer(name string, creator func(methodDecl *MethodMetadata) Value) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Look up class method in hierarchy
	normalizedName := ident.Normalize(name)
	for current := c.ClassInfo; current != nil; current = current.Parent {
		// A pointer cannot represent an overload set: bind the first declared
		// overload, which is the policy the analyzer records for @TClass.M
		// (firstBindableMethodOverload). ClassMethods holds the last declaration
		// of a name, so the overload list is consulted first.
		if overloads, exists := current.ClassMethodOverloads[normalizedName]; exists && len(overloads) > 0 {
			return creator(overloads[0]), true
		}
		if method, exists := current.ClassMethods[normalizedName]; exists {
			return creator(method), true
		}
	}
	return nil, false
}

// InvokeConstructor invokes a constructor.
func (c *ClassValue) InvokeConstructor(name string, executor func(methodDecl *MethodMetadata) Value) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Look up constructor in hierarchy
	for current := c.ClassInfo; current != nil; current = current.Parent {
		// Check ConstructorOverloads first
		for ctorName, overloads := range current.ConstructorOverloads {
			if ident.Equal(ctorName, name) && len(overloads) > 0 {
				return executor(overloads[0]), true
			}
		}
		// Check single constructors
		for ctorName, ctor := range current.Constructors {
			if ident.Equal(ctorName, name) {
				return executor(ctor), true
			}
		}
	}
	return nil, false
}

// GetNestedClass returns a nested class by name.
func (c *ClassValue) GetNestedClass(name string) Value {
	if c == nil || c.ClassInfo == nil {
		return nil
	}
	nested := c.ClassInfo.LookupNestedClass(name)
	if nested == nil {
		return nil
	}
	return &ClassInfoValue{ClassInfo: nested}
}

// ReadClassProperty reads a class property value using the executor callback.
func (c *ClassValue) ReadClassProperty(name string, executor func(propInfo any) Value) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Look up property in hierarchy
	propInfo := c.ClassInfo.lookupProperty(name)
	if propInfo == nil || !propInfo.IsClassProperty {
		return nil, false
	}
	return executor(propInfo), true
}

// GetClassInfo returns the underlying ClassInfo.
func (c *ClassValue) GetClassInfo() IClassInfo {
	if c == nil {
		return nil
	}
	return c.ClassInfo
}

// IsAssignableTo checks if this class reference can be assigned to a variable
// of the given metaclass type. This implements the assignment compatibility
// rules for metaclasses.
//
// For example:
// - TDog (ClassValue) can be assigned to "class of TAnimal" if TDog inherits from TAnimal
// - TDog cannot be assigned to "class of TCat"
//
// Returns true if assignment is allowed, false otherwise.
func (c *ClassValue) IsAssignableTo(targetClass *ClassInfo) bool {
	if c.ClassInfo == nil || targetClass == nil {
		return false
	}

	// Check if c.ClassInfo is targetClass or derives from it
	current := c.ClassInfo
	for current != nil {
		if current.Name == targetClass.Name {
			return true
		}
		current = current.Parent
	}
	return false
}

// SetClassVar sets a class variable by name in the hierarchy.
func (c *ClassValue) SetClassVar(name string, value Value) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	return c.ClassInfo.setClassVar(name, value)
}

// HasClassVar checks if a class variable exists in the hierarchy.
func (c *ClassValue) HasClassVar(name string) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	return c.ClassInfo.hasClassVar(name)
}

// WriteClassProperty writes to a class property using the executor callback.
func (c *ClassValue) WriteClassProperty(name string, value Value, executor func(propInfo any, value Value) Value) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Look up class property in hierarchy
	propDesc := c.ClassInfo.LookupProperty(name)
	if propDesc == nil {
		return nil, false
	}
	propInfo, ok := propDesc.Impl.(*types.PropertyInfo)
	if !ok || !propInfo.IsClassProperty {
		return nil, false
	}
	return executor(propInfo, value), true
}

// AsClassValue attempts to cast a Value to a ClassValue.
// Returns the ClassValue and true if successful, or nil and false if not.
func AsClassValue(v Value) (*ClassValue, bool) {
	cls, ok := v.(*ClassValue)
	return cls, ok
}

// buildVirtualMethodTable builds the virtual method table for this class.
// Implements virtual/override/reintroduce semantics for method dispatch.
func (c *ClassInfo) buildVirtualMethodTable() {
	c.VirtualMethods = make(map[string]*VirtualMethodEntry)
	if c.Parent != nil {
		for sig, inherited := range c.Parent.VirtualMethods {
			entry := *inherited
			entry.IsReintroduce = false
			c.VirtualMethods[sig] = &entry
		}
	}
	install := func(method *MethodMetadata) {
		if method == nil || method.Declaration == nil || (method.Owner != nil && method.Owner != c) {
			return
		}
		sig := methodSignature(method.Declaration)
		if method.IsVirtual {
			c.VirtualMethods[sig] = &VirtualMethodEntry{OwningClass: c, Method: method, IsVirtual: true}
		} else if method.IsOverride {
			if entry := c.VirtualMethods[sig]; entry != nil {
				entry.Method = method
				entry.IsOverride = true
			}
		}
	}
	for _, group := range c.MethodOverloads {
		for _, method := range group {
			install(method)
		}
	}
	for _, group := range c.ClassMethodOverloads {
		for _, method := range group {
			install(method)
		}
	}
	for _, group := range c.ConstructorOverloads {
		for _, method := range group {
			install(method)
		}
	}
	install(c.Destructor)
}

// methodSignature generates a signature string for a method.
// The signature includes the method name and parameter types to support overloading.
func methodSignature(method *ast.FunctionDecl) string {
	sig := ident.Normalize(method.Name.Value)

	// For now, use a simple signature that includes parameter count
	// In a full implementation, we'd include parameter types
	sig += fmt.Sprintf("_%d", len(method.Parameters))

	return sig
}

// GetClassConstant looks up a class constant by name in the class hierarchy.
// Returns pre-evaluated constant values only.
func (c *ClassInfo) GetClassConstant(name string) (Value, bool) {
	if c == nil {
		return nil, false
	}

	// Check ConstantValues cache (case-insensitive)
	normalizedName := ident.Normalize(name)
	for constName, value := range c.Constants {
		if ident.Normalize(constName) == normalizedName {
			return value, true
		}
	}

	// Check parent class hierarchy
	if c.Parent != nil {
		return c.Parent.GetClassConstant(name)
	}

	return nil, false
}
