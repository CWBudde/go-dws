package runtime

import (
	"sync"
)

// RefCountManager manages object and interface reference counting.
//
// This interface allows the evaluator to manage object lifecycles without
// importing the interpreter package. Destructor invocation uses a callback
// pattern to avoid circular imports.
type RefCountManager interface {
	// IncrementRef increments the reference count for an object or interface.
	// Returns the same value for chaining.
	IncrementRef(val Value) Value

	// DecrementRef decrements the reference count and invokes the destructor
	// callback when the count reaches 0. Returns nil.
	DecrementRef(val Value) Value

	// ReleaseObject decrements the reference count if the object is not nil.
	ReleaseObject(obj *ObjectInstance)

	// ReleaseInterface decrements the reference count on the interface's underlying object.
	ReleaseInterface(intf *InterfaceInstance)

	// WrapInInterface creates an InterfaceInstance and increments ref count.
	// Returns a new InterfaceInstance wrapping the object.
	WrapInInterface(iface InterfaceInfo, obj *ObjectInstance) *InterfaceInstance

	// SetDestructorCallback registers the callback invoked when RefCount reaches 0.
	SetDestructorCallback(callback DestructorCallback)
}

// DestructorCallback is invoked when an object's reference count reaches 0.
// The implementation should look up and execute the "Destroy" method if it exists.
type DestructorCallback func(obj *ObjectInstance) error

// defaultRefCountManager implements RefCountManager with callback-based destructors.
type defaultRefCountManager struct {
	destructorCallback DestructorCallback
	mu                 sync.RWMutex // Protects callback
}

// NewRefCountManager creates a default reference count manager.
func NewRefCountManager() RefCountManager {
	return &defaultRefCountManager{}
}

// IncrementRef increments the reference count for an object or interface.
func (m *defaultRefCountManager) IncrementRef(val Value) Value {
	if val == nil {
		return val
	}

	switch v := val.(type) {
	case *ObjectInstance:
		if v != nil {
			v.RefCount++
		}
	case *InterfaceInstance:
		if v != nil && v.Object != nil {
			v.Object.RefCount++
		}
	}

	return val
}

// DecrementRef decrements the reference count and calls destructor if it reaches 0.
func (m *defaultRefCountManager) DecrementRef(val Value) Value {
	if val == nil {
		return nil
	}

	var obj *ObjectInstance
	switch v := val.(type) {
	case *ObjectInstance:
		obj = v
	case *InterfaceInstance:
		obj = v.Object
	}

	if obj == nil || obj.Destroyed {
		return nil
	}

	obj.RefCount--
	if obj.RefCount < 0 {
		obj.RefCount = 0
	}

	// Invoke destructor when ref count reaches 0
	if obj.RefCount <= 0 {
		m.mu.RLock()
		callback := m.destructorCallback
		m.mu.RUnlock()

		if callback != nil {
			// Ignore error for now - destructor errors are logged by interpreter
			_ = callback(obj)
		}
	}

	return nil
}

// ReleaseObject decrements ref count for an object (convenience method).
func (m *defaultRefCountManager) ReleaseObject(obj *ObjectInstance) {
	if obj != nil {
		m.DecrementRef(obj)
	}
}

// ReleaseInterface decrements ref count for an interface's underlying object.
func (m *defaultRefCountManager) ReleaseInterface(intf *InterfaceInstance) {
	if intf != nil {
		m.DecrementRef(intf)
	}
}

// WrapInInterface creates an InterfaceInstance and increments the object's ref count.
func (m *defaultRefCountManager) WrapInInterface(iface InterfaceInfo, obj *ObjectInstance) *InterfaceInstance {
	intf := &InterfaceInstance{
		Interface: iface,
		Object:    obj,
	}
	if obj != nil {
		obj.RefCount++
	}
	return intf
}

// SetDestructorCallback registers the destructor callback.
func (m *defaultRefCountManager) SetDestructorCallback(callback DestructorCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.destructorCallback = callback
}
