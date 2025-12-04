package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
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
//
// Task 3.5.20: Now implements runtime.IInterfaceInfo interface to allow
// InterfaceInstance (in runtime package) to reference it without circular imports.
type InterfaceInfo struct {
	Parent     *InterfaceInfo
	Methods    map[string]*ast.FunctionDecl
	Properties map[string]*types.PropertyInfo
	Name       string
}

// Ensure InterfaceInfo implements runtime.IInterfaceInfo at compile time.
var _ runtime.IInterfaceInfo = (*InterfaceInfo)(nil)

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

// GetName returns the interface name.
// Task 3.5.20: Implements runtime.IInterfaceInfo interface.
func (ii *InterfaceInfo) GetName() string {
	return ii.Name
}

// GetParent returns the parent interface, or nil if this is a root interface.
// Task 3.5.20: Implements runtime.IInterfaceInfo interface.
func (ii *InterfaceInfo) GetParent() runtime.IInterfaceInfo {
	if ii.Parent == nil {
		return nil
	}
	return ii.Parent
}

// GetMethod looks up a method by name in this interface.
// It searches the interface hierarchy, starting with this interface
// and walking up through parent interfaces until the method is found.
// Returns nil if the method is not found in the interface hierarchy.
// Task 3.5.20: Returns any (instead of *ast.FunctionDecl) to implement IInterfaceInfo.
func (ii *InterfaceInfo) GetMethod(name string) any {
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

// getMethodDecl is an internal helper that returns the method as *ast.FunctionDecl.
// Used by existing code that needs the concrete type.
func (ii *InterfaceInfo) getMethodDecl(name string) *ast.FunctionDecl {
	method := ii.GetMethod(name)
	if method == nil {
		return nil
	}
	return method.(*ast.FunctionDecl)
}

// HasMethod checks if this interface (or any parent) has a method with the given name.
func (ii *InterfaceInfo) HasMethod(name string) bool {
	return ii.GetMethod(name) != nil
}

// GetProperty looks up a property by name in this interface (case-insensitive).
// It searches the interface hierarchy until found.
// Task 3.5.20: Returns *runtime.PropertyInfo (wrapper) to implement IInterfaceInfo.
func (ii *InterfaceInfo) GetProperty(name string) *runtime.PropertyInfo {
	normalized := ident.Normalize(name)

	if prop, exists := ii.Properties[normalized]; exists {
		// Convert types.PropertyInfo to runtime.PropertyInfo
		return &runtime.PropertyInfo{
			Name:      prop.Name,
			IsIndexed: prop.IsIndexed,
			IsDefault: prop.IsDefault,
			ReadSpec:  prop.ReadSpec,
			WriteSpec: prop.WriteSpec,
			Impl:      prop, // Store original for backward compatibility
		}
	}

	if ii.Parent != nil {
		return ii.Parent.GetProperty(name)
	}

	return nil
}

// getPropertyInfo is an internal helper that returns the property as *types.PropertyInfo.
// Used by existing code that needs the concrete type.
func (ii *InterfaceInfo) getPropertyInfo(name string) *types.PropertyInfo {
	normalized := ident.Normalize(name)
	if prop, exists := ii.Properties[normalized]; exists {
		return prop
	}
	if ii.Parent != nil {
		return ii.Parent.getPropertyInfo(name)
	}
	return nil
}

// HasProperty checks if the interface (or any parent) declares a property.
func (ii *InterfaceInfo) HasProperty(name string) bool {
	return ii.GetProperty(name) != nil
}

// GetDefaultProperty returns the default property defined on the interface hierarchy, if any.
// Task 3.5.20: Renamed from getDefaultProperty and returns *runtime.PropertyInfo to implement IInterfaceInfo.
func (ii *InterfaceInfo) GetDefaultProperty() *runtime.PropertyInfo {
	for _, prop := range ii.AllProperties() {
		if prop.IsDefault {
			return prop
		}
	}
	return nil
}

// AllMethods returns all methods in this interface, including inherited methods.
// Returns a new map containing all methods from this interface and its parents.
// Task 3.5.20: Returns map[string]any (instead of map[string]*ast.FunctionDecl) to implement IInterfaceInfo.
func (ii *InterfaceInfo) AllMethods() map[string]any {
	result := make(map[string]any)

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

// allMethodsDecl is an internal helper that returns methods as map[string]*ast.FunctionDecl.
// Used by existing code that needs the concrete type.
func (ii *InterfaceInfo) allMethodsDecl() map[string]*ast.FunctionDecl {
	result := make(map[string]*ast.FunctionDecl)

	// Add parent methods first (so child methods can override)
	if ii.Parent != nil {
		for name, method := range ii.Parent.allMethodsDecl() {
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
// Task 3.5.20: Returns map[string]*runtime.PropertyInfo to implement IInterfaceInfo.
func (ii *InterfaceInfo) AllProperties() map[string]*runtime.PropertyInfo {
	result := make(map[string]*runtime.PropertyInfo)

	if ii.Parent != nil {
		for name, prop := range ii.Parent.AllProperties() {
			result[name] = prop
		}
	}

	for name, prop := range ii.Properties {
		// Convert types.PropertyInfo to runtime.PropertyInfo
		result[name] = &runtime.PropertyInfo{
			Name:      prop.Name,
			IsIndexed: prop.IsIndexed,
			IsDefault: prop.IsDefault,
			ReadSpec:  prop.ReadSpec,
			WriteSpec: prop.WriteSpec,
			Impl:      prop,
		}
	}

	return result
}

// allPropertiesInfo is an internal helper that returns properties as map[string]*types.PropertyInfo.
// Used by existing code that needs the concrete type.
func (ii *InterfaceInfo) allPropertiesInfo() map[string]*types.PropertyInfo {
	result := make(map[string]*types.PropertyInfo)

	if ii.Parent != nil {
		for name, prop := range ii.Parent.allPropertiesInfo() {
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
// Task 3.5.20: Moved to runtime.InterfaceInstance for bridge constructor elimination.
// Type alias provided for backward compatibility during migration.
type InterfaceInstance = runtime.InterfaceInstance

// NewInterfaceInstance creates a new interface instance wrapping an object.
// Task 3.5.20: Function alias for backward compatibility.
var NewInterfaceInstance = runtime.NewInterfaceInstance

// ImplementsInterface checks if the underlying object's class implements all methods
// of the given interface. This is used for runtime type checking.
// Task 3.5.20: Kept in interp package as it needs ClassInfo access.
func ImplementsInterface(ii *InterfaceInstance, iface *InterfaceInfo) bool {
	if ii.Object == nil {
		return false // nil doesn't implement any interface
	}
	concreteClass, ok := ii.Object.Class.(*ClassInfo)
	if !ok {
		return false
	}
	return classImplementsInterface(concreteClass, iface)
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
	// Task 3.5.20: Use allMethodsDecl() to get concrete types for iteration
	targetMethods := target.allMethodsDecl()

	// Check that the source interface has each required method
	for methodName := range targetMethods {
		if !source.HasMethod(methodName) {
			return false
		}
	}

	return true
}

// runDestructor executes the destructor for an object with lifecycle tracking.
// If node is provided, "Object already destroyed" is raised when called twice.
func (i *Interpreter) runDestructor(obj *ObjectInstance, destructor *ast.FunctionDecl, node ast.Node) Value {
	if obj == nil {
		return &NilValue{}
	}

	// Prevent double-destruction errors for explicit calls
	if obj.Destroyed {
		if node != nil {
			return i.newErrorWithLocation(node, "Object already destroyed")
		}
		return &NilValue{}
	}

	// Reuse existing destructor if not supplied
	if destructor == nil && obj.Class != nil {
		destructor = obj.Class.LookupMethod("Destroy")
	}

	// If no destructor is defined, just mark destroyed
	if destructor == nil {
		obj.Destroyed = true
		// Task 3.5.42: Reset to 0 after destruction - this is a finalization step, not ref counting.
		// Do NOT use RefCountManager here - we're cleaning up a destroyed object.
		obj.RefCount = 0
		return &NilValue{}
	}

	obj.DestroyCallDepth++
	defer func() {
		obj.DestroyCallDepth--
		if obj.DestroyCallDepth == 0 {
			obj.Destroyed = true
			// Task 3.5.42: Reset to 0 after destruction - this is a finalization step, not ref counting.
			// Do NOT use RefCountManager here - we're cleaning up a destroyed object.
			obj.RefCount = 0
		}
	}()

	// Create a temporary environment for the destructor call
	destructorEnv := NewEnclosedEnvironment(i.env)
	prevEnv := i.env
	i.env = destructorEnv

	// Bind Self and class constants
	i.env.Define("Self", obj)
	concreteClass, ok := obj.Class.(*ClassInfo)
	if ok {
		i.bindClassConstantsToEnv(concreteClass)
	}

	// Push call stack for better stack traces
	i.pushCallStack(obj.Class.GetName() + ".Destroy")
	defer i.popCallStack()

	// Execute destructor body
	result := Value(&NilValue{})
	if destructor.Body != nil {
		result = i.Eval(destructor.Body)
	}

	// Restore previous environment
	i.env = prevEnv

	return result
}

// runDestructorForRefCount executes the destructor as a callback from RefCountManager.
// Task 3.5.41: Wrapper for RefCountManager destructor callback pattern.
// This method follows the destructor callback contract:
// 1. Check if obj.Destroyed is true (skip if already destroyed)
// 2. Mark obj.Destroyed = true BEFORE execution (prevent recursion)
// 3. Execute the destructor method
// 4. Reset obj.RefCount = 0 after completion
func (i *Interpreter) runDestructorForRefCount(obj *ObjectInstance) error {
	if obj == nil || obj.Destroyed {
		return nil
	}

	// If we're already inside this object's destructor, skip to avoid infinite recursion
	if obj.DestroyCallDepth > 0 {
		return nil
	}

	// Look up the destructor
	destructor := obj.Class.LookupMethod("Destroy")

	// Execute destructor via runDestructor (handles marking and environment)
	// Pass nil for node since this is automatic ref count cleanup
	i.runDestructor(obj, destructor, nil)

	return nil
}

// callDestructorIfNeeded decrements the reference count of an object and calls its
// destructor if the reference count reaches zero.
// Task 9.1.5: Consolidates destructor logic to reduce code duplication.
// Task 3.5.42: Migrated to use RefCountManager for consistent ref counting across all code paths.
func (i *Interpreter) callDestructorIfNeeded(obj *ObjectInstance) {
	if obj == nil || obj.Destroyed {
		return
	}

	// If we're already inside this object's destructor (eg, destructor clears a
	// global reference pointing back to itself), skip to avoid infinite recursion.
	if obj.DestroyCallDepth > 0 {
		return
	}

	// Task 3.5.42: Use RefCountManager for proper ref counting
	// DecrementRef handles:
	// - Decrementing the ref count
	// - Clamping negative values to 0
	// - Invoking destructor when ref count reaches 0
	i.evaluatorInstance.RefCountManager().DecrementRef(obj)
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

	// Register only in TypeSystem (legacy map removed)
	i.typeSystem.RegisterInterface("IInterface", iinterface)
}
