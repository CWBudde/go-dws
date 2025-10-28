package interp

import (
	"fmt"

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
	// Check this interface's methods first
	if method, exists := ii.Methods[name]; exists {
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
func NewInterfaceInstance(iface *InterfaceInfo, obj *ObjectInstance) *InterfaceInstance {
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
	// Check current class
	if _, exists := class.Methods[methodName]; exists {
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
