package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
)

// ============================================================================
// Interface Runtime Metadata
// ============================================================================

// InterfaceInfo represents runtime interface metadata.
// It stores information about an interface's structure including methods,
// parent interface, and external binding information.
type InterfaceInfo struct {
	Parent  *InterfaceInfo
	Methods map[string]*ast.FunctionDecl
	Name    string
}

// NewInterfaceInfo creates a new InterfaceInfo with the given name.
// Methods map is initialized as empty.
func NewInterfaceInfo(name string) *InterfaceInfo {
	return &InterfaceInfo{
		Name:    name,
		Parent:  nil,
		Methods: make(map[string]*ast.FunctionDecl),
	}
}

// GetMethod looks up a method by name in this interface.
// It searches the interface hierarchy, starting with this interface
// and walking up through parent interfaces until the method is found.
// Returns nil if the method is not found in the interface hierarchy.
func (ii *InterfaceInfo) GetMethod(name string) *ast.FunctionDecl {
	// Normalize to lowercase for case-insensitive lookup
	normalizedName := strings.ToLower(name)

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

// ============================================================================
// Helper Functions
// ============================================================================

// classImplementsInterface checks if a class implements all methods of an interface.
// This includes checking inherited methods from parent classes.
// Returns true if the class has all required methods with matching signatures.
func classImplementsInterface(class *ClassInfo, iface *InterfaceInfo) bool {
	// Get all methods required by the interface (including inherited)
	requiredMethods := iface.AllMethods()

	// Check that the class has each required method
	for methodName := range requiredMethods {
		// Check in current class's methods
		if !classHasMethod(class, methodName) {
			return false
		}
		// Note: Full signature matching would be done in semantic analysis
		// At runtime, we just check method existence
	}

	return true
}

// classHasMethod checks if a class or its parents have a method with the given name.
func classHasMethod(class *ClassInfo, methodName string) bool {
	// Normalize to lowercase for case-insensitive lookup
	normalizedName := strings.ToLower(methodName)

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
	for _, value := range env.store {
		// Skip ReferenceValue entries (like function name aliases)
		if _, isRef := value.(*ReferenceValue); isRef {
			continue
		}

		if intfInst, ok := value.(*InterfaceInstance); ok {
			// Release the interface reference
			i.ReleaseInterfaceReference(intfInst)
		} else if objInst, ok := value.(*ObjectInstance); ok {
			// Release the object reference (decrement ref count and call destructor if needed)
			i.callDestructorIfNeeded(objInst)
		}
	}
}
