package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Interface Runtime Metadata
// ============================================================================

// InterfaceInfo represents runtime interface metadata.
// It stores information about an interface's structure including methods,
// parent interface, and external binding information.
type InterfaceInfo struct {
	Parent     *InterfaceInfo
	Methods    map[string]*ast.FunctionDecl
	Properties map[string]*types.PropertyInfo
	Name       string
}

// NewInterfaceInfo creates a new InterfaceInfo with the given name.
// Methods map is initialized as empty.
func NewInterfaceInfo(name string) *InterfaceInfo {
	return &InterfaceInfo{
		Name:       name,
		Parent:     nil,
		Methods:    make(map[string]*ast.FunctionDecl),
		Properties: make(map[string]*types.PropertyInfo),
	}
}

// GetMethod looks up a method by name in this interface.
// It searches the interface hierarchy, starting with this interface
// and walking up through parent interfaces until the method is found.
// Returns nil if the method is not found in the interface hierarchy.
func (ii *InterfaceInfo) GetMethod(name string) *ast.FunctionDecl {
	normalizedName := ident.Normalize(name)

	// Check this interface's methods first
	if method, exists := ii.Methods[normalizedName]; exists {
		return method
	}

	// Check parent interface if present
	if ii.Parent != nil {
		return ii.Parent.GetMethod(name)
	}

	// Method not found
	return nil
}

// HasMethod checks if this interface (or any parent) has a method with the given name.
func (ii *InterfaceInfo) HasMethod(name string) bool {
	return ii.GetMethod(name) != nil
}

// GetProperty looks up a property by name in this interface (case-insensitive).
// It searches the interface hierarchy until found.
func (ii *InterfaceInfo) GetProperty(name string) *types.PropertyInfo {
	normalized := ident.Normalize(name)

	if prop, exists := ii.Properties[normalized]; exists {
		return prop
	}

	if ii.Parent != nil {
		return ii.Parent.GetProperty(name)
	}

	return nil
}

// HasProperty checks if the interface (or any parent) declares a property.
func (ii *InterfaceInfo) HasProperty(name string) bool {
	return ii.GetProperty(name) != nil
}

// getDefaultProperty returns the default property defined on the interface hierarchy, if any.
func (ii *InterfaceInfo) getDefaultProperty() *types.PropertyInfo {
	for _, prop := range ii.AllProperties() {
		if prop.IsDefault {
			return prop
		}
	}
	return nil
}

// AllMethods returns all methods in this interface, including inherited methods.
// Returns a new map containing all methods from this interface and its parents.
func (ii *InterfaceInfo) AllMethods() map[string]*ast.FunctionDecl {
	result := make(map[string]*ast.FunctionDecl)

	// Add parent methods first (so child methods can override)
	if ii.Parent != nil {
		for name, method := range ii.Parent.AllMethods() {
			result[name] = method
		}
	}

	// Add this interface's methods
	for name, method := range ii.Methods {
		result[name] = method
	}

	return result
}

// AllProperties returns all properties declared on this interface and its parents.
func (ii *InterfaceInfo) AllProperties() map[string]*types.PropertyInfo {
	result := make(map[string]*types.PropertyInfo)

	if ii.Parent != nil {
		for name, prop := range ii.Parent.AllProperties() {
			result[name] = prop
		}
	}

	for name, prop := range ii.Properties {
		result[name] = prop
	}

	return result
}

// ============================================================================
// Interface Instance
// ============================================================================

// InterfaceInstance represents a runtime instance of an interface.
// It wraps an ObjectInstance and provides interface-based access to it.
// This implements the Value interface so it can be used as a runtime value.
type InterfaceInstance struct {
	// Interface points to the interface metadata
	Interface *InterfaceInfo

	// Object is a reference to the implementing object
	// This allows method dispatch to the actual object implementation
	Object *ObjectInstance
}

// NewInterfaceInstance creates a new interface instance wrapping an object.
// Task 9.1.5: Increments the object's reference count when wrapping it in an interface.
func NewInterfaceInstance(iface *InterfaceInfo, obj *ObjectInstance) *InterfaceInstance {
	// Increment reference count when interface takes ownership of object
	if obj != nil {
		obj.RefCount++
	}

	return &InterfaceInstance{
		Interface: iface,
		Object:    obj,
	}
}

// Type returns "INTERFACE" for interface instances.
// Implements the Value interface.
func (ii *InterfaceInstance) Type() string {
	return "INTERFACE"
}

// String returns the string representation of the interface instance.
// Implements the Value interface.
func (ii *InterfaceInstance) String() string {
	if ii.Object == nil {
		return fmt.Sprintf("%s instance (nil)", ii.Interface.Name)
	}
	return fmt.Sprintf("%s instance (wrapping %s)", ii.Interface.Name, ii.Object.Class.Name)
}

// GetUnderlyingObject returns the object wrapped by this interface instance.
// This is used for interface-to-object casting.
func (ii *InterfaceInstance) GetUnderlyingObject() *ObjectInstance {
	return ii.Object
}

// ImplementsInterface checks if the underlying object's class implements all methods
// of the given interface. This is used for runtime type checking.
func (ii *InterfaceInstance) ImplementsInterface(iface *InterfaceInfo) bool {
	if ii.Object == nil {
		return false // nil doesn't implement any interface
	}
	return classImplementsInterface(ii.Object.Class, iface)
}

// GetUnderlyingObjectValue returns the object wrapped by this interface instance.
// Returns nil if the interface wraps a nil object.
// Task 3.5.87: Implements evaluator.InterfaceInstanceValue interface.
// Note: This returns ObjectValue (interface) while GetUnderlyingObject returns *ObjectInstance.
func (ii *InterfaceInstance) GetUnderlyingObjectValue() Value {
	if ii.Object == nil {
		return nil
	}
	return ii.Object
}

// InterfaceName returns the name of the interface type.
// Task 3.5.87: Implements evaluator.InterfaceInstanceValue interface.
func (ii *InterfaceInstance) InterfaceName() string {
	if ii.Interface == nil {
		return ""
	}
	return ii.Interface.Name
}

// HasInterfaceMethod checks if the interface declares a method with the given name.
// The check includes parent interfaces.
// Task 3.5.87: Implements evaluator.InterfaceInstanceValue interface.
func (ii *InterfaceInstance) HasInterfaceMethod(name string) bool {
	if ii.Interface == nil {
		return false
	}
	return ii.Interface.HasMethod(name)
}

// HasInterfaceProperty checks if the interface declares a property with the given name.
// The check includes parent interfaces.
// Task 3.5.87: Implements evaluator.InterfaceInstanceValue interface.
func (ii *InterfaceInstance) HasInterfaceProperty(name string) bool {
	if ii.Interface == nil {
		return false
	}
	return ii.Interface.HasProperty(name)
}

// LookupProperty searches for a property by name in the interface hierarchy.
// Task 3.5.99a: Implements evaluator.PropertyAccessor interface.
// Returns a PropertyDescriptor wrapping types.PropertyInfo, or nil if not found.
func (ii *InterfaceInstance) LookupProperty(name string) *evaluator.PropertyDescriptor {
	if ii.Interface == nil {
		return nil
	}

	propInfo := ii.Interface.GetProperty(name)
	if propInfo == nil {
		return nil
	}

	return &evaluator.PropertyDescriptor{
		Name:      propInfo.Name,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		Impl:      propInfo, // Store the original PropertyInfo for later use
	}
}

// GetDefaultProperty returns the default property for this interface, if any.
// Task 3.5.99a: Implements evaluator.PropertyAccessor interface.
// Returns a PropertyDescriptor wrapping types.PropertyInfo, or nil if no default property exists.
func (ii *InterfaceInstance) GetDefaultProperty() *evaluator.PropertyDescriptor {
	if ii.Interface == nil {
		return nil
	}

	propInfo := ii.Interface.getDefaultProperty()
	if propInfo == nil {
		return nil
	}

	return &evaluator.PropertyDescriptor{
		Name:      propInfo.Name,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		Impl:      propInfo, // Store the original PropertyInfo for later use
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// classImplementsInterface checks if a class explicitly declares that it implements an interface.
// In DWScript, a class must explicitly declare interface implementation in its class declaration.
// This function checks the class's Interfaces list and recursively checks parent classes in the
// class hierarchy, since interface implementations are inherited from parent classes to child classes.
// Returns true if the class or any of its parents explicitly declares implementation of the interface.
func classImplementsInterface(class *ClassInfo, iface *InterfaceInfo) bool {
	// Defensive check: nil class doesn't implement any interface
	if class == nil {
		return false
	}

	// Check if this class explicitly declares the interface
	for _, implementedIface := range class.Interfaces {
		// Direct match: class declares this exact interface
		if implementedIface == iface {
			return true
		}

		// Check if the declared interface inherits from the target interface
		// (interface inheritance compatibility check)
		if interfaceInheritsFrom(implementedIface, iface) {
			return true
		}
	}

	// Check parent class (interfaces are inherited)
	if class.Parent != nil {
		return classImplementsInterface(class.Parent, iface)
	}

	return false
}

// classExplicitlyImplementsInterface checks whether a class or its parents directly declare the interface
// (without considering interface inheritance). This matches DWScript 'implements' operator behavior.
func classExplicitlyImplementsInterface(class *ClassInfo, iface *InterfaceInfo) bool {
	if class == nil || iface == nil {
		return false
	}

	for _, implemented := range class.Interfaces {
		if implemented == iface {
			return true
		}
	}

	if class.Parent != nil {
		return classExplicitlyImplementsInterface(class.Parent, iface)
	}

	return false
}

// isClassCompatible checks if objClass is the same as or inherits from targetClass.
// This is used for type checking in 'as' operator and other casting operations.
func isClassCompatible(objClass, targetClass *ClassInfo) bool {
	current := objClass
	for current != nil {
		if ident.Equal(current.Name, targetClass.Name) {
			return true
		}
		current = current.Parent
	}
	return false
}

// interfaceInheritsFrom checks if sourceIface inherits from targetIface.
// Returns true if sourceIface is a descendant of targetIface in the interface hierarchy.
func interfaceInheritsFrom(sourceIface *InterfaceInfo, targetIface *InterfaceInfo) bool {
	if sourceIface == nil {
		return false
	}

	// Walk up the interface hierarchy
	current := sourceIface.Parent
	for current != nil {
		if current == targetIface {
			return true
		}
		current = current.Parent
	}

	return false
}

// classHasMethod checks if a class or its parents have a method with the given name.
func classHasMethod(class *ClassInfo, methodName string) bool {
	// Defensive check: nil class doesn't have any methods
	if class == nil {
		return false
	}

	normalizedName := ident.Normalize(methodName)

	// Check current class
	if _, exists := class.Methods[normalizedName]; exists {
		return true
	}

	// Check parent class (recursive)
	if class.Parent != nil {
		return classHasMethod(class.Parent, methodName)
	}

	return false
}

// interfaceIsCompatible checks if one interface is compatible with another.
// An interface is compatible if it implements all methods of the target interface.
// This is used for interface-to-interface casting.
func interfaceIsCompatible(source *InterfaceInfo, target *InterfaceInfo) bool {
	// Get all methods required by the target interface
	targetMethods := target.AllMethods()

	// Check that the source interface has each required method
	for methodName := range targetMethods {
		if !source.HasMethod(methodName) {
			return false
		}
	}

	return true
}

// callDestructorIfNeeded decrements the reference count of an object and calls its
// destructor if the reference count reaches zero.
// Task 9.1.5: Consolidates destructor logic to reduce code duplication.
func (i *Interpreter) callDestructorIfNeeded(obj *ObjectInstance) {
	if obj == nil {
		return
	}

	// DEBUG
	// fmt.Printf("DEBUG callDestructorIfNeeded: RefCount before = %d\n", obj.RefCount)

	// Decrement reference count
	obj.RefCount--

	// DEBUG
	// fmt.Printf("DEBUG callDestructorIfNeeded: RefCount after = %d\n", obj.RefCount)

	// If reference count reaches 0 or below, call the destructor
	if obj.RefCount <= 0 {
		// DEBUG
		// fmt.Printf("DEBUG callDestructorIfNeeded: Calling destructor\n")

		// Look for Destroy method in the class hierarchy
		destructor := obj.Class.lookupMethod("Destroy")
		if destructor != nil {
			// Call the destructor
			// Create a temporary environment for the destructor call
			destructorEnv := NewEnclosedEnvironment(i.env)
			destructorEnv.Define("Self", obj)

			// Save current environment and switch to destructor environment
			prevEnv := i.env
			i.env = destructorEnv

			// Execute destructor body
			i.Eval(destructor.Body)

			// Restore previous environment
			i.env = prevEnv
		}
	}
}

// ReleaseInterfaceReference decrements the reference count of the object wrapped by
// an interface instance and calls the destructor if the reference count reaches zero.
// Task 9.1.5: This implements automatic lifetime management for interface-held objects.
// Returns the destructor result value (or nil) and any error from destructor execution.
func (i *Interpreter) ReleaseInterfaceReference(intfInst *InterfaceInstance) Value {
	if intfInst == nil || intfInst.Object == nil {
		return &NilValue{}
	}

	// Use the consolidated helper method
	i.callDestructorIfNeeded(intfInst.Object)

	return &NilValue{}
}

// cleanupInterfaceReferences iterates through all variables in an environment
// and releases any interface references AND object references.
// This is called when a scope ends (e.g., function returns).
// Task 9.1.5: This implements automatic cleanup of interface-held and object-held references when scope ends.
func (i *Interpreter) cleanupInterfaceReferences(env *Environment) {
	if env == nil || env.store == nil {
		return
	}

	// Iterate through all variables in the environment
	env.store.Range(func(_ string, value Value) bool {
		// Skip ReferenceValue entries (like function name aliases)
		if _, isRef := value.(*ReferenceValue); isRef {
			return true // continue
		}

		if intfInst, ok := value.(*InterfaceInstance); ok {
			// Release the interface reference
			i.ReleaseInterfaceReference(intfInst)
		} else if objInst, ok := value.(*ObjectInstance); ok {
			// Release the object reference (decrement ref count and call destructor if needed)
			i.callDestructorIfNeeded(objInst)
		} else if funcPtr, ok := value.(*FunctionPointerValue); ok {
			// Clean up method pointers that hold object references
			// This complements the RefCount++ in objects_hierarchy.go when creating interface method pointers
			if objInst, isObj := funcPtr.SelfObject.(*ObjectInstance); isObj {
				i.callDestructorIfNeeded(objInst)
			}
		}
		return true // continue
	})
}

// ============================================================================
// Built-in Interface Registration
// ============================================================================

// registerBuiltinInterfaces registers the IInterface base interface.
// IInterface is the root interface type in DWScript, similar to how TObject is the root class.
// Note: Interfaces do NOT automatically inherit from IInterface unless explicitly declared.
// Classes that want to be castable to IInterface must explicitly list it in their interface declarations.
func (i *Interpreter) registerBuiltinInterfaces() {
	// Register IInterface as the root interface available for explicit implementation
	// IInterface is an empty interface with no methods, serving as a marker interface
	// Classes implementing any interface should typically also explicitly declare IInterface
	iinterface := NewInterfaceInfo("IInterface")
	iinterface.Parent = nil // Root of the interface hierarchy

	// Register with lowercase key for case-insensitive lookup
	i.interfaces[ident.Normalize("IInterface")] = iinterface
	// Task 3.5.46: Also register in TypeSystem for shared access
	i.typeSystem.RegisterInterface("IInterface", iinterface)
}
