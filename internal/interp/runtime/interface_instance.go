// Package runtime provides runtime value types for the DWScript interpreter.
// This file contains InterfaceInstance, which represents a runtime instance of an interface.
package runtime

// ============================================================================
// Interface Instance
// ============================================================================

// InterfaceInstance represents a runtime instance of an interface.
// It wraps an ObjectInstance and provides interface-based access to it.
// This implements the Value interface so it can be used as a runtime value.
//
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

// interfaceInstanceDisplay is what DWScript prints for a bound interface
// reference. The name is DWScript's own compiler symbol class leaking through
// its Variant-to-string conversion; the fixtures in
// testdata/fixtures/FunctionsVariant record it verbatim, and it does not vary
// with the interface or the implementing class.
const interfaceInstanceDisplay = "TInterfaceSymbol"

// String returns the string representation of the interface instance,
// matching DWScript: an unbound reference prints as "nil".
// Implements the Value interface.
func (ii *InterfaceInstance) String() string {
	if ii.Object == nil {
		return "nil"
	}
	return interfaceInstanceDisplay
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
func (ii *InterfaceInstance) GetUnderlyingObjectValue() Value {
	if ii.Object == nil {
		return nil
	}
	return ii.Object
}

// InterfaceName returns the name of the interface type.
func (ii *InterfaceInstance) InterfaceName() string {
	if ii.Interface == nil {
		return ""
	}
	return ii.Interface.GetName()
}

// HasInterfaceMethod checks if the interface declares a method with the given name.
// The check includes parent interfaces.
func (ii *InterfaceInstance) HasInterfaceMethod(name string) bool {
	if ii.Interface == nil {
		return false
	}
	return ii.Interface.HasMethod(name)
}

// HasInterfaceProperty checks if the interface declares a property with the given name.
// The check includes parent interfaces.
func (ii *InterfaceInstance) HasInterfaceProperty(name string) bool {
	if ii.Interface == nil {
		return false
	}
	return ii.Interface.HasProperty(name)
}

// LookupProperty searches for a property by name in the interface hierarchy.
// Returns a PropertyDescriptor wrapping PropertyInfo, or nil if not found.
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
		ReadSpec:  propInfo.ReadSpec,
		WriteSpec: propInfo.WriteSpec,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		Impl:      propInfo, // Store the original PropertyInfo for later use
	}
}

// GetDefaultProperty returns the default property for this interface, if any.
// Returns a PropertyDescriptor wrapping PropertyInfo, or nil if no default property exists.
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
		ReadSpec:  propInfo.ReadSpec,
		WriteSpec: propInfo.WriteSpec,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		Impl:      propInfo, // Store the original PropertyInfo for later use
	}
}
