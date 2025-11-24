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
//
// Task 3.5.39: Migration to AST-free metadata
// - Metadata field contains AST-free runtime metadata (Phase 9)
// - Legacy AST fields maintained for backward compatibility
// - Future: Code will migrate to use Metadata, AST fields deprecated
type ClassInfo struct {
	// === AST-Free Metadata (Task 3.5.39) ===

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
	ExternalName         string
	Name                 string
	DefaultConstructor   string
	Interfaces           []*InterfaceInfo
	IsExternal           bool
	IsAbstract           bool
	IsPartial            bool
}

// NewClassInfo creates a new ClassInfo with the given name.
// Fields, Methods, ClassVars, ClassMethods, and Properties maps are initialized as empty.
// Task 3.5.39: Also initializes AST-free ClassMetadata.
func NewClassInfo(name string) *ClassInfo {
	return &ClassInfo{
		Name:                 name,
		Parent:               nil,
		Metadata:             runtime.NewClassMetadata(name), // Task 3.5.39
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
	}
}

// ObjectInstance represents a runtime instance of a class.
// It implements the Value interface so it can be used as a runtime value.
type ObjectInstance struct {
	// Class points to the class metadata
	Class *ClassInfo

	// Fields maps field names to their runtime values
	Fields map[string]Value

	// RefCount tracks interface references to this object for lifetime management
	// Task 9.1.5: Objects held by interfaces use reference counting
	// - Starts at 0 when object is created; incremented when assigned to variable or interface
	// - Increments when assigned to another variable or interface
	// - Decrements when variable is reassigned, set to nil, or goes out of scope
	// - Destructor called when RefCount reaches 0
	RefCount int
}

// NewObjectInstance creates a new object instance of the given class.
// Field values are initialized as an empty map.
// Task 9.1.5: RefCount starts at 0; incremented when assigned to variable or interface
func NewObjectInstance(class *ClassInfo) *ObjectInstance {
	return &ObjectInstance{
		Class:    class,
		Fields:   make(map[string]Value),
		RefCount: 0, // Start with reference count of 0
	}
}

// GetField retrieves the value of a field by name.
// Returns nil if the field doesn't exist or hasn't been set.
// Task 3.5.40: Uses ClassMetadata for AST-free field lookup.
// During migration, falls back to legacy Fields map if metadata unavailable.
func (o *ObjectInstance) GetField(name string) Value {
	// Guard against nil class
	if o.Class == nil {
		return nil
	}

	normalizedName := ident.Normalize(name)

	// Try metadata first (AST-free path), walking up the inheritance chain.
	// This avoids missing inherited fields when the child class adds its own fields.
	if fieldMeta := lookupFieldMetadata(o.Class.Metadata, normalizedName); fieldMeta != nil {
		if val, exists := o.Fields[normalizedName]; exists {
			return val
		}
		if val, exists := o.Fields[name]; exists {
			return val
		}
		return nil
	}

	// Legacy fallback: Check if field exists in ClassInfo.Fields
	if _, exists := o.Class.Fields[normalizedName]; !exists {
		// Also check non-normalized key for backward compatibility
		if _, exists := o.Class.Fields[name]; !exists {
			return nil
		}
		// Return field with original name key
		return o.Fields[name]
	}

	// Return field with normalized name
	return o.Fields[normalizedName]
}

// SetField sets the value of a field by name.
// The field must be defined in the class's field map.
// Task 3.5.40: Uses ClassMetadata for AST-free field lookup.
// During migration, falls back to legacy Fields map if metadata unavailable.
func (o *ObjectInstance) SetField(name string, value Value) {
	// Guard against nil class
	if o.Class == nil {
		return
	}

	normalizedName := ident.Normalize(name)

	// Try metadata first (AST-free path), walking up the inheritance chain.
	if lookupFieldMetadata(o.Class.Metadata, normalizedName) != nil {
		o.Fields[normalizedName] = value
		return
	}

	// Legacy fallback: Check if field exists in ClassInfo.Fields
	if _, exists := o.Class.Fields[normalizedName]; exists {
		o.Fields[normalizedName] = value
		return
	}

	// Also check non-normalized key for backward compatibility
	if _, exists := o.Class.Fields[name]; exists {
		o.Fields[name] = value
	}
}

// GetMethod looks up a method by name in this object's class.
// It searches the class hierarchy, starting with the object's class
// and walking up through parent classes until the method is found.
// Returns nil if the method is not found in the class hierarchy.
//
// This implements method resolution order (MRO) and supports method overriding:
//   - If a child class defines a method with the same name as a parent class method,
//     the child's method is returned (overriding).
//
// Note: This performs static method resolution (not virtual dispatch).
// Virtual dispatch is implemented inline in objects_methods.go where needed.
//
// Task 3.5.40: Uses ClassMetadata for method lookup (AST-free).
// Returns AST node from metadata for backward compatibility during migration.
func (o *ObjectInstance) GetMethod(name string) *ast.FunctionDecl {
	if o.Class == nil {
		return nil
	}
	return o.Class.lookupMethod(name)
}

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
// Value Interface Implementation
// ============================================================================

// Type returns "OBJECT" to indicate this is an object instance.
func (o *ObjectInstance) Type() string {
	return "OBJECT"
}

// String returns a string representation of the object instance.
// Format: "TClassName instance"
func (o *ObjectInstance) String() string {
	return fmt.Sprintf("%s instance", o.Class.Name)
}

// ClassName returns the class name of this object instance.
// Task 3.5.72: Implements evaluator.ObjectValue interface.
func (o *ObjectInstance) ClassName() string {
	if o == nil || o.Class == nil {
		return ""
	}
	return o.Class.Name
}

// HasProperty checks if this object's class has a property with the given name.
// The check includes the entire class hierarchy.
// Task 3.5.72: Implements evaluator.ObjectValue interface.
func (o *ObjectInstance) HasProperty(name string) bool {
	if o == nil || o.Class == nil {
		return false
	}
	return o.Class.lookupProperty(name) != nil
}

// HasMethod checks if this object's class has a method with the given name.
// Task 3.5.72: Implements evaluator.ObjectValue interface.
func (o *ObjectInstance) HasMethod(name string) bool {
	if o == nil || o.Class == nil {
		return false
	}
	_, exists := o.Class.Methods[ident.Normalize(name)]
	return exists
}

// GetClassVar retrieves a class variable value by name from this object's class hierarchy.
// Returns the value and true if found, nil and false otherwise.
// Task 3.5.86: Implements evaluator.ObjectValue interface for direct class variable access.
func (o *ObjectInstance) GetClassVar(name string) (Value, bool) {
	if o == nil || o.Class == nil {
		return nil, false
	}
	value, owningClass := o.Class.lookupClassVar(name)
	if owningClass == nil {
		return nil, false
	}
	return value, true
}

// IsInstanceOf checks whether the object derives from the given class.
func (o *ObjectInstance) IsInstanceOf(target *ClassInfo) bool {
	if o == nil || o.Class == nil || target == nil {
		return false
	}
	current := o.Class
	for current != nil {
		if current.Name == target.Name {
			return true
		}
		current = current.Parent
	}
	return false
}

// Helper function to check if a value is an ObjectInstance
func isObject(v Value) bool {
	_, ok := v.(*ObjectInstance)
	return ok
}

// AsObject attempts to cast a Value to an ObjectInstance.
// Returns the ObjectInstance and true if successful, or nil and false if not.
func AsObject(v Value) (*ObjectInstance, bool) {
	obj, ok := v.(*ObjectInstance)
	return obj, ok
}

// lookupFieldMetadata searches for a field in the class metadata hierarchy.
// Returns the metadata for the field if found, or nil otherwise.
func lookupFieldMetadata(meta *runtime.ClassMetadata, normalizedName string) *runtime.FieldMetadata {
	for current := meta; current != nil; current = current.Parent {
		if field, ok := current.Fields[normalizedName]; ok {
			return field
		}
	}
	return nil
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

// Helper function to check if a value is a ClassValue
func isClassValue(v Value) bool {
	_, ok := v.(*ClassValue)
	return ok
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
