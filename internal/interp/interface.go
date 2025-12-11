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

// InterfaceInfo represents runtime interface metadata: methods, parent interface,
// and properties. Implements runtime.IInterfaceInfo to avoid circular imports.
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
func (ii *InterfaceInfo) GetName() string {
	return ii.Name
}

// GetParent returns the parent interface, or nil if this is a root interface.
func (ii *InterfaceInfo) GetParent() runtime.IInterfaceInfo {
	if ii.Parent == nil {
		return nil
	}
	return ii.Parent
}

// GetMethod looks up a method by name, searching the interface hierarchy.
// Returns nil if the method is not found.
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

// HasMethod checks if this interface (or any parent) has a method with the given name.
func (ii *InterfaceInfo) HasMethod(name string) bool {
	return ii.GetMethod(name) != nil
}

// GetProperty looks up a property by name, searching the interface hierarchy.
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

// HasProperty checks if the interface (or any parent) declares a property.
func (ii *InterfaceInfo) HasProperty(name string) bool {
	return ii.GetProperty(name) != nil
}

// GetDefaultProperty returns the default property from the interface hierarchy, if any.
func (ii *InterfaceInfo) GetDefaultProperty() *runtime.PropertyInfo {
	for _, prop := range ii.AllProperties() {
		if prop.IsDefault {
			return prop
		}
	}
	return nil
}

// AllMethods returns all methods including inherited ones from parent interfaces.
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

// AllProperties returns all properties from this interface and its parents.
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

// ============================================================================
// Interface Instance
// ============================================================================

// InterfaceInstance represents a runtime instance of an interface.
type InterfaceInstance = runtime.InterfaceInstance

// NewInterfaceInstance creates a new interface instance wrapping an object.
var NewInterfaceInstance = runtime.NewInterfaceInstance

// ImplementsInterface checks if the object's class implements all methods of the interface.
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

// classImplementsInterface checks if a class (or its parents) explicitly declares the interface.
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

// interfaceIsCompatible checks if one interface is compatible with another.
// An interface is compatible if it implements all methods of the target interface.
func interfaceIsCompatible(source *InterfaceInfo, target *InterfaceInfo) bool {
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

	if destructor == nil {
		obj.Destroyed = true
		obj.RefCount = 0
		return &NilValue{}
	}

	obj.DestroyCallDepth++
	defer func() {
		obj.DestroyCallDepth--
		if obj.DestroyCallDepth == 0 {
			obj.Destroyed = true
			obj.RefCount = 0
		}
	}()

	// Create a temporary environment for the destructor call
	defer i.PushScope()()

	// Bind Self and class constants
	i.Env().Define("Self", obj)
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

	return result
}

// runDestructorForRefCount executes the destructor as a callback from RefCountManager.
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

// callDestructorIfNeeded calls the destructor if the reference count reaches zero.
func (i *Interpreter) callDestructorIfNeeded(obj *ObjectInstance) {
	if obj == nil || obj.Destroyed {
		return
	}

	if obj.DestroyCallDepth > 0 {
		return // Prevent recursion if destructor clears a reference to itself
	}

	i.evaluatorInstance.RefCountManager().DecrementRef(obj)
}

// ReleaseInterfaceReference decrements the reference count and calls the destructor if needed.
func (i *Interpreter) ReleaseInterfaceReference(intfInst *InterfaceInstance) Value {
	if intfInst == nil || intfInst.Object == nil {
		return &NilValue{}
	}

	i.callDestructorIfNeeded(intfInst.Object)

	return &NilValue{}
}

// cleanupInterfaceReferences releases all interface and object references when a scope ends.
func (i *Interpreter) cleanupInterfaceReferences(env *Environment) {
	if env == nil {
		return
	}

	// Iterate through all variables in the environment
	env.Range(func(_ string, value Value) bool {
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
func (i *Interpreter) registerBuiltinInterfaces() {
	iinterface := NewInterfaceInfo("IInterface")
	i.typeSystem.RegisterInterface("IInterface", iinterface)
}
