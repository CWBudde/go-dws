package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// VirtualMethodEntry tracks virtual method dispatch information.
// Task 9.14: Support for virtual/override/reintroduce semantics
type VirtualMethodEntry struct {
	// IntroducedBy is the class that first declared this method as virtual
	IntroducedBy *ClassInfo
	// Implementation is the method declaration to actually call for this class
	Implementation *ast.FunctionDecl
	// IsVirtual indicates if this method participates in virtual dispatch
	IsVirtual bool
	// IsReintroduced indicates if this method breaks the virtual dispatch chain
	IsReintroduced bool
}

// ClassInfo represents runtime class metadata.
// It stores information about a class's structure including fields, methods,
// parent class, and constructor/destructor.

type ClassInfo struct {
	// === AST-Free Metadata ===

	// Metadata contains AST-free runtime metadata for this class.
	// Populated during class declaration evaluation.
	// Phase 9: Enables method calls via MethodID without AST access.
	Metadata *runtime.ClassMetadata

	// === Legacy AST Fields (to be deprecated) ===
	// These fields are maintained for backward compatibility during migration.
	// New code should use Metadata instead.

	Constants            map[string]*ast.ConstDecl
	ClassVars            map[string]Value
	ConstructorOverloads map[string][]*ast.FunctionDecl
	VirtualMethodTable   map[string]*VirtualMethodEntry
	Fields               map[string]types.Type
	FieldDecls           map[string]*ast.FieldDecl
	Constructor          *ast.FunctionDecl
	ClassMethodOverloads map[string][]*ast.FunctionDecl
	ClassMethods         map[string]*ast.FunctionDecl
	Methods              map[string]*ast.FunctionDecl
	ConstantValues       map[string]Value
	Constructors         map[string]*ast.FunctionDecl
	Operators            *runtimeOperatorRegistry
	Properties           map[string]*types.PropertyInfo
	Destructor           *ast.FunctionDecl
	Parent               *ClassInfo
	MethodOverloads      map[string][]*ast.FunctionDecl
	NestedClasses        map[string]*ClassInfo
	ExternalName         string
	Name                 string
	DefaultConstructor   string
	Interfaces           []*InterfaceInfo
	IsExternalFlag       bool // Renamed to avoid conflict with IsExternal() method
	IsAbstractFlag       bool // Renamed to avoid conflict with IsAbstract() method
	IsPartial            bool
}

// NewClassInfo creates a new ClassInfo with the given name.
// Fields, Methods, ClassVars, ClassMethods, and Properties maps are initialized as empty.
func NewClassInfo(name string) *ClassInfo {
	return &ClassInfo{
		Name:                 name,
		Parent:               nil,
		Metadata:             runtime.NewClassMetadata(name),
		Fields:               make(map[string]types.Type),
		FieldDecls:           make(map[string]*ast.FieldDecl),
		ClassVars:            make(map[string]Value),
		Constants:            make(map[string]*ast.ConstDecl),
		ConstantValues:       make(map[string]Value),
		Methods:              make(map[string]*ast.FunctionDecl),
		MethodOverloads:      make(map[string][]*ast.FunctionDecl),
		ClassMethods:         make(map[string]*ast.FunctionDecl),
		ClassMethodOverloads: make(map[string][]*ast.FunctionDecl),
		Operators:            newRuntimeOperatorRegistry(),
		Constructors:         make(map[string]*ast.FunctionDecl),
		ConstructorOverloads: make(map[string][]*ast.FunctionDecl),
		Properties:           make(map[string]*types.PropertyInfo),
		VirtualMethodTable:   make(map[string]*VirtualMethodEntry),
		NestedClasses:        make(map[string]*ClassInfo),
	}
}

// Compile-time assertion that ClassInfo implements runtime.IClassInfo
var _ runtime.IClassInfo = (*ClassInfo)(nil)

// === IClassInfo Interface Implementation ===
// These methods allow ObjectInstance (in runtime package) to access ClassInfo
// without creating a circular import dependency.

// GetName returns the class name
func (c *ClassInfo) GetName() string {
	if c == nil {
		return ""
	}
	return c.Name
}

// GetParent returns the parent class metadata
func (c *ClassInfo) GetParent() runtime.IClassInfo {
	if c == nil || c.Parent == nil {
		return nil
	}
	return c.Parent
}

// GetMetadata returns the AST-free metadata
func (c *ClassInfo) GetMetadata() *runtime.ClassMetadata {
	if c == nil {
		return nil
	}
	return c.Metadata
}

// LookupMethod finds a method by name in the class hierarchy
func (c *ClassInfo) LookupMethod(name string) *ast.FunctionDecl {
	return c.lookupMethod(name)
}

// LookupProperty finds a property by name in the class hierarchy
func (c *ClassInfo) LookupProperty(name string) *runtime.PropertyInfo {
	propInfo := c.lookupProperty(name)
	if propInfo == nil {
		return nil
	}
	return &runtime.PropertyInfo{
		Name:      propInfo.Name,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		ReadSpec:  propInfo.ReadSpec,
		WriteSpec: propInfo.WriteSpec,
		Impl:      propInfo,
	}
}

// GetDefaultProperty returns the default property
func (c *ClassInfo) GetDefaultProperty() *runtime.PropertyInfo {
	propInfo := c.getDefaultProperty()
	if propInfo == nil {
		return nil
	}
	return &runtime.PropertyInfo{
		Name:      propInfo.Name,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		ReadSpec:  propInfo.ReadSpec,
		WriteSpec: propInfo.WriteSpec,
		Impl:      propInfo,
	}
}

// FieldExists checks if a field exists
func (c *ClassInfo) FieldExists(normalizedName string) bool {
	if c == nil {
		return false
	}
	_, exists := c.Fields[normalizedName]
	return exists
}

// GetFieldsMap returns the legacy field declarations map
func (c *ClassInfo) GetFieldsMap() map[string]*ast.FieldDecl {
	if c == nil {
		return nil
	}
	return c.FieldDecls
}

// GetMethodsMap returns the legacy methods map
func (c *ClassInfo) GetMethodsMap() map[string]*ast.FunctionDecl {
	if c == nil {
		return nil
	}
	return c.Methods
}

// LookupClassVar retrieves a class variable by name
func (c *ClassInfo) LookupClassVar(name string) (Value, runtime.IClassInfo) {
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
func (c *ClassInfo) GetVirtualMethodTable() map[string]*runtime.VirtualMethodEntry {
	if c == nil || c.VirtualMethodTable == nil {
		return nil
	}
	// Convert from interp.VirtualMethodEntry to runtime.VirtualMethodEntry
	result := make(map[string]*runtime.VirtualMethodEntry, len(c.VirtualMethodTable))
	for sig, entry := range c.VirtualMethodTable {
		if entry != nil {
			result[sig] = &runtime.VirtualMethodEntry{
				Method:        entry.Implementation,
				OwningClass:   entry.IntroducedBy,
				IsVirtual:     entry.IsVirtual,
				IsOverride:    false, // interp.VirtualMethodEntry doesn't track this separately
				IsReintroduce: entry.IsReintroduced,
			}
		}
	}
	return result
}

// LookupOperator finds an operator overload
func (c *ClassInfo) LookupOperator(operator string, operandTypes []string) (*runtime.OperatorEntry, bool) {
	if c == nil || c.Operators == nil {
		return nil, false
	}
	entry, found := c.lookupOperator(operator, operandTypes)
	if !found || entry == nil {
		return nil, false
	}
	// Note: runtime.OperatorEntry is currently a placeholder.
	// Operator lookup is only used within interp package, so returning
	// a simplified version here. Full implementation when needed.
	return &runtime.OperatorEntry{
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
func (c *ClassInfo) GetInterfaces() []*runtime.InterfaceInfo {
	if c == nil {
		return nil
	}
	// For now, return nil as InterfaceInfo interface is not fully defined
	// This will need proper implementation when interface support is added
	return nil
}

// IsAbstract returns true if this class is declared as abstract.
// Task 3.5.22k: Added for CreateObject migration to evaluator.
func (c *ClassInfo) IsAbstract() bool {
	if c == nil {
		return false
	}
	return c.IsAbstractFlag
}

// IsExternal returns true if this class is declared as external.
// Task 3.5.22k: Added for CreateObject migration to evaluator.
func (c *ClassInfo) IsExternal() bool {
	if c == nil {
		return false
	}
	return c.IsExternalFlag
}

// GetConstructor returns a constructor declaration by name (case-insensitive).
// Task 3.5.22k: Added for CreateObject migration to evaluator.
func (c *ClassInfo) GetConstructor(name string) *ast.FunctionDecl {
	if c == nil {
		return nil
	}
	normalizedName := ident.Normalize(name)
	if ctor, ok := c.Constructors[normalizedName]; ok {
		return ctor
	}
	return nil
}

// GetFieldTypesMap returns the field name to type mapping for this class.
// Task 3.5.22k: Added for CreateObject migration to evaluator.
func (c *ClassInfo) GetFieldTypesMap() map[string]any {
	if c == nil {
		return nil
	}
	// Convert map[string]types.Type to map[string]any to avoid import cycle
	result := make(map[string]any, len(c.Fields))
	for name, typ := range c.Fields {
		result[name] = typ
	}
	return result
}

// === End IClassInfo Interface Implementation ===

// ============================================================================
// ObjectInstance - Moved to runtime package
// ============================================================================
// Task 3.5.17: ObjectInstance has been moved to internal/interp/runtime/object.go
// A type alias is provided in internal/interp/value.go for backward compatibility.
//
// ObjectInstance now uses runtime.IClassInfo interface instead of *ClassInfo
// to avoid circular import dependencies.
// ============================================================================

// lookupNestedClass returns a nested class by short name (case-insensitive).
func (c *ClassInfo) lookupNestedClass(name string) *ClassInfo {
	if c == nil {
		return nil
	}
	if nested, ok := c.NestedClasses[ident.Normalize(name)]; ok {
		return nested
	}
	return nil
}

// ============================================================================
// ClassInfo Helper Methods
// ============================================================================

// lookupMethod searches for a method in the class hierarchy.
// It starts with the current class and walks up the parent chain.
// Returns the first method found, or nil if not found.
//
// Task 3.5.40: Uses ClassMetadata for AST-free method lookup.
// During migration, falls back to legacy Methods map if metadata is unavailable.
func (c *ClassInfo) lookupMethod(name string) *ast.FunctionDecl {
	normalizedName := ident.Normalize(name)

	// Task 3.5.40: Try metadata first (AST-free path)
	if c.Metadata != nil {
		if _, exists := c.Metadata.Methods[normalizedName]; exists {
			// Extract AST node from metadata for backward compatibility
			// During migration, MethodMetadata.Body is still ast.Statement
			// We need to return the full FunctionDecl, so fall back to legacy for now
			// TODO: After full migration, return callable instead of AST
			if legacyMethod, legacyExists := c.Methods[normalizedName]; legacyExists {
				return legacyMethod
			}
		}
	}

	// Legacy fallback: Check current class Methods map
	// This is needed during migration when metadata exists but method isn't in Metadata.Methods
	if method, exists := c.Methods[normalizedName]; exists {
		return method
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

// lookupConstant searches for a constant in the class hierarchy.
// It starts with the current class and walks up the parent chain.
// Returns the ConstDecl and the ClassInfo that owns it, or (nil, nil) if not found.
func (c *ClassInfo) lookupConstant(name string) (*ast.ConstDecl, *ClassInfo) {
	// Check current class with case-insensitive match
	for constName, constDecl := range c.Constants {
		if ident.Equal(constName, name) {
			return constDecl, c
		}
	}

	// Check parent class (recursive)
	if c.Parent != nil {
		return c.Parent.lookupConstant(name)
	}

	// Not found
	return nil, nil
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
// Task 3.2.11a: Added for static class variable assignment.
// Returns true if the variable was found and set, false if not found.
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
// Task 3.2.11a: Added for static class variable assignment validation.
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

// lookupOperator searches for a class operator in the hierarchy.
func (c *ClassInfo) lookupOperator(operator string, operandTypes []string) (*runtimeOperatorEntry, bool) {
	if c == nil {
		return nil, false
	}
	if c.Operators != nil {
		if entry, ok := c.Operators.lookup(operator, operandTypes); ok {
			return entry, true
		}
	}
	if c.Parent != nil {
		return c.Parent.lookupOperator(operator, operandTypes)
	}
	return nil, false
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

// ClassValue represents a class reference (metaclass value) at runtime.
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
// Task 3.5.88: Implements evaluator.ClassMetaValue interface.
func (c *ClassValue) GetClassName() string {
	if c == nil || c.ClassInfo == nil {
		return ""
	}
	return c.ClassInfo.Name
}

// GetClassType returns the class type (metaclass) as a ClassValue.
// Task 3.5.157: Implements evaluator.ClassMetaValue interface.
// For ClassValue, this returns itself since it already represents the class type.
func (c *ClassValue) GetClassType() Value {
	return c
}

// GetClassVar retrieves a class variable value by name from the class hierarchy.
// Returns the value and true if found, nil and false otherwise.
// Task 3.5.88: Implements evaluator.ClassMetaValue interface.
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
// Returns the value and true if found, nil and false otherwise.
// Task 3.5.88: Implements evaluator.ClassMetaValue interface.
func (c *ClassValue) GetClassConstant(name string) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Check ConstantValues cache first (case-insensitive)
	for constName, value := range c.ClassInfo.ConstantValues {
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
// Task 3.5.88: Implements evaluator.ClassMetaValue interface.
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
// Task 3.5.88: Implements evaluator.ClassMetaValue interface.
func (c *ClassValue) HasConstructor(name string) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	return c.ClassInfo.HasConstructor(name)
}

// InvokeParameterlessClassMethod invokes a parameterless class method.
// Task 3.2.10: Implements evaluator.ClassMetaValue interface.
func (c *ClassValue) InvokeParameterlessClassMethod(name string, executor func(methodDecl any) Value) (Value, bool) {
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

// CreateClassMethodPointer creates a function pointer for a class method with parameters.
// Task 3.2.10: Implements evaluator.ClassMetaValue interface.
func (c *ClassValue) CreateClassMethodPointer(name string, creator func(methodDecl any) Value) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Look up class method in hierarchy
	normalizedName := ident.Normalize(name)
	for current := c.ClassInfo; current != nil; current = current.Parent {
		if method, exists := current.ClassMethods[normalizedName]; exists {
			if len(method.Parameters) > 0 {
				return creator(method), true
			}
			return nil, false // Parameterless
		}
		if overloads, exists := current.ClassMethodOverloads[normalizedName]; exists && len(overloads) > 0 {
			// Return pointer for first overload with parameters
			for _, m := range overloads {
				if len(m.Parameters) > 0 {
					return creator(m), true
				}
			}
			return nil, false // All parameterless
		}
	}
	return nil, false
}

// InvokeConstructor invokes a constructor.
// Task 3.2.10: Implements evaluator.ClassMetaValue interface.
func (c *ClassValue) InvokeConstructor(name string, executor func(methodDecl any) Value) (Value, bool) {
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
// Task 3.2.10: Implements evaluator.ClassMetaValue interface.
func (c *ClassValue) GetNestedClass(name string) Value {
	if c == nil || c.ClassInfo == nil {
		return nil
	}
	nested := c.ClassInfo.lookupNestedClass(name)
	if nested == nil {
		return nil
	}
	return &ClassInfoValue{ClassInfo: nested}
}

// ReadClassProperty reads a class property value using the executor callback.
// Task 3.2.10: Implements evaluator.ClassMetaValue interface.
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
// Task 3.2.10: Implements evaluator.ClassMetaValue interface.
func (c *ClassValue) GetClassInfo() any {
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

// AsClassValue attempts to cast a Value to a ClassValue.
// Returns the ClassValue and true if successful, or nil and false if not.
func AsClassValue(v Value) (*ClassValue, bool) {
	cls, ok := v.(*ClassValue)
	return cls, ok
}

// buildVirtualMethodTable builds the virtual method table for this class.
// Task 9.14: This implements proper virtual/override/reintroduce semantics.
//
// The VMT tracks which method implementation should be called for each virtual method.
// Rules:
//   - Virtual methods start a dispatch chain (added to VMT)
//   - Override continues the chain (updates the VMT entry to point to new implementation)
//   - Reintroduce does NOT update parent's VMT entry - parent's virtual method stays in VMT
//     (the reintroduced method is static and doesn't participate in virtual dispatch)
func (c *ClassInfo) buildVirtualMethodTable() {
	// First, copy parent's VMT if we have a parent
	// This inherits all virtual methods from the parent
	if c.Parent != nil {
		for sig, entry := range c.Parent.VirtualMethodTable {
			// Copy the entry - child inherits parent's virtual methods
			c.VirtualMethodTable[sig] = &VirtualMethodEntry{
				IntroducedBy:   entry.IntroducedBy,
				Implementation: entry.Implementation,
				IsVirtual:      entry.IsVirtual,
				IsReintroduced: false,
			}
		}
	}

	// Now process this class's own methods
	for _, method := range c.MethodOverloads {
		for _, m := range method {
			sig := methodSignature(m)

			if m.IsVirtual {
				// This method is declared as virtual
				// It starts a new virtual dispatch chain
				c.VirtualMethodTable[sig] = &VirtualMethodEntry{
					IntroducedBy:   c,
					Implementation: m,
					IsVirtual:      true,
					IsReintroduced: false,
				}
			} else if m.IsOverride {
				// This method overrides a parent virtual method
				// Update the VMT entry to point to this override
				if existingEntry, exists := c.VirtualMethodTable[sig]; exists {
					// Keep the IntroducedBy from parent, but update implementation
					existingEntry.Implementation = m
				}
				// If no existing entry, this is an error (should be caught by semantic analysis)
			} else if m.IsReintroduce {
				// Reintroduce does NOT participate in virtual dispatch
				// The parent's virtual method (if any) remains in the VMT
				// The reintroduced method is only callable via static (compile-time) type
				// So we do nothing to the VMT here - parent's entry stays unchanged
			}
			// If none of virtual/override/reintroduce, it's a new non-virtual method
			// We don't add it to the VMT
		}
	}

	// Process class methods (static methods) similarly
	for _, method := range c.ClassMethodOverloads {
		for _, m := range method {
			sig := methodSignature(m)

			if m.IsVirtual {
				c.VirtualMethodTable[sig] = &VirtualMethodEntry{
					IntroducedBy:   c,
					Implementation: m,
					IsVirtual:      true,
					IsReintroduced: false,
				}
			} else if m.IsOverride {
				if existingEntry, exists := c.VirtualMethodTable[sig]; exists {
					existingEntry.Implementation = m
				}
			} else if m.IsReintroduce {
				// Same as instance methods - don't update VMT
			}
		}
	}

	// Process constructors (they can also be virtual/override)
	for _, ctors := range c.ConstructorOverloads {
		for _, ctor := range ctors {
			sig := methodSignature(ctor)

			if ctor.IsVirtual {
				c.VirtualMethodTable[sig] = &VirtualMethodEntry{
					IntroducedBy:   c,
					Implementation: ctor,
					IsVirtual:      true,
					IsReintroduced: false,
				}
			} else if ctor.IsOverride {
				if existingEntry, exists := c.VirtualMethodTable[sig]; exists {
					existingEntry.Implementation = ctor
				}
			}
			// Constructors typically don't use reintroduce
		}
	}
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
// Task 3.5.32: Implements runtime.ClassConstantProvider interface for property reads.
// Returns the pre-evaluated constant value and true if found, nil and false otherwise.
//
// Note: This only returns constants that have been pre-evaluated and stored in ConstantValues.
// Lazy evaluation of constant expressions is handled separately by the interpreter.
func (c *ClassInfo) GetClassConstant(name string) (Value, bool) {
	if c == nil {
		return nil, false
	}

	// Check ConstantValues cache (case-insensitive)
	normalizedName := ident.Normalize(name)
	for constName, value := range c.ConstantValues {
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
