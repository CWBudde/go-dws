// Package runtime provides runtime value types for the DWScript interpreter.
// This file contains InterfaceInstance, which represents a runtime instance of an interface.
//
// Task 3.5.20: Moved from internal/interp/interface.go to break circular dependencies
// and enable direct evaluator access without adapter pattern.
package runtime

import (
	"fmt"
)

// ============================================================================
// Interface Instance
// ============================================================================

// InterfaceInstance represents a runtime instance of an interface.
// It wraps an ObjectInstance and provides interface-based access to it.
// This implements the Value interface so it can be used as a runtime value.
//
// Task 3.5.20: Moved to runtime package following the same pattern as ObjectInstance (3.5.17).
// Uses IInterfaceInfo interface to reference InterfaceInfo (in interp package) without import cycle.
type InterfaceInstance struct {
	// Interface points to the interface metadata (via IInterfaceInfo interface)
	// This allows referencing InterfaceInfo without importing interp package
	Interface IInterfaceInfo

	// Object is a reference to the implementing object
	// This allows method dispatch to the actual object implementation
	Object *ObjectInstance
}

// NewInterfaceInstance creates a new interface instance wrapping an object.
// Task 9.1.5: Increments the object's reference count when wrapping it in an interface.
// Task 3.5.20: Constructor moved to runtime package.
func NewInterfaceInstance(iface IInterfaceInfo, obj *ObjectInstance) *InterfaceInstance {
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
		if ii.Interface != nil {
			return fmt.Sprintf("%s instance (nil)", ii.Interface.GetName())
		}
		return "interface instance (nil)"
	}
	if ii.Interface != nil {
		return fmt.Sprintf("%s instance (wrapping %s)", ii.Interface.GetName(), ii.Object.Class.GetName())
	}
	return fmt.Sprintf("interface instance (wrapping %s)", ii.Object.Class.GetName())
}

// GetUnderlyingObject returns the object wrapped by this interface instance.
// This is used for interface-to-object casting.
func (ii *InterfaceInstance) GetUnderlyingObject() *ObjectInstance {
	return ii.Object
}

// GetUnderlyingObjectValue returns the object wrapped by this interface instance.
// Returns nil if the interface wraps a nil object.
//
// Note: This returns Value (interface) while GetUnderlyingObject returns *ObjectInstance (concrete type).
// Task 3.5.20: Implements evaluator.InterfaceInstanceValue interface for direct evaluator access.
func (ii *InterfaceInstance) GetUnderlyingObjectValue() Value {
	if ii.Object == nil {
		return nil
	}
	return ii.Object
}

// InterfaceName returns the name of the interface type.
// Task 3.5.20: Implements evaluator.InterfaceInstanceValue interface.
func (ii *InterfaceInstance) InterfaceName() string {
	if ii.Interface == nil {
		return ""
	}
	return ii.Interface.GetName()
}

// HasInterfaceMethod checks if the interface declares a method with the given name.
// The check includes parent interfaces.
// Task 3.5.20: Implements evaluator.InterfaceInstanceValue interface.
func (ii *InterfaceInstance) HasInterfaceMethod(name string) bool {
	if ii.Interface == nil {
		return false
	}
	return ii.Interface.HasMethod(name)
}

// HasInterfaceProperty checks if the interface declares a property with the given name.
// The check includes parent interfaces.
// Task 3.5.20: Implements evaluator.InterfaceInstanceValue interface.
func (ii *InterfaceInstance) HasInterfaceProperty(name string) bool {
	if ii.Interface == nil {
		return false
	}
	return ii.Interface.HasProperty(name)
}

// LookupProperty searches for a property by name in the interface hierarchy.
// Returns a PropertyDescriptor wrapping PropertyInfo, or nil if not found.
// Task 3.5.20: Implements PropertyAccessor interface.
func (ii *InterfaceInstance) LookupProperty(name string) *PropertyDescriptor {
	if ii.Interface == nil {
		return nil
	}

	propInfo := ii.Interface.GetProperty(name)
	if propInfo == nil {
		return nil
	}

	return &PropertyDescriptor{
		Name:      propInfo.Name,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		Impl:      propInfo, // Store the original PropertyInfo for later use
	}
}

// GetDefaultProperty returns the default property for this interface, if any.
// Returns a PropertyDescriptor wrapping PropertyInfo, or nil if no default property exists.
// Task 3.5.20: Implements PropertyAccessor interface.
func (ii *InterfaceInstance) GetDefaultProperty() *PropertyDescriptor {
	if ii.Interface == nil {
		return nil
	}

	propInfo := ii.Interface.GetDefaultProperty()
	if propInfo == nil {
		return nil
	}

	return &PropertyDescriptor{
		Name:      propInfo.Name,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		Impl:      propInfo, // Store the original PropertyInfo for later use
	}
}
