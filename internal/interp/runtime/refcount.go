package runtime

import (
	"sync"
)

// RefCountManager manages object and interface reference counting.
// Task 3.5.39-3.5.40: Moves reference counting from interpreter to runtime.
//
// This interface allows the evaluator to manage object lifecycles without
// importing the interpreter package. Destructor invocation uses a callback
// pattern to avoid circular imports.
type RefCountManager interface {
	// IncrementRef increments the reference count for an object.
	// Called when:
	// - Creating a new variable reference to an object
	// - Assigning interface to interface (copies reference)
	// - Creating method pointers with SelfObject
	// - Returning interfaces from functions
	//
	// Returns the same value for chaining (v = mgr.IncrementRef(v))
	IncrementRef(val Value) Value

	// DecrementRef decrements the reference count for an object.
	// If the reference count reaches 0, invokes the destructor callback.
	// Called when:
	// - Reassigning object variable (old object released)
	// - Setting object variable to nil
	// - Reassigning interface variable (old interface released)
	//
	// Returns nil (indicating the old reference is gone)
	DecrementRef(val Value) Value

	// ReleaseObject combines decrement + potential destructor call.
	// Convenience method for: if obj != nil { DecrementRef(obj) }
	//
	// Used in assignment operations where the old value is being replaced.
	ReleaseObject(obj *ObjectInstance)

	// ReleaseInterface decrements ref count on the underlying object.
	// Handles nil checks and unwraps the InterfaceInstance.
	//
	// Used when reassigning interface variables or releasing temporary interfaces.
	ReleaseInterface(intf *InterfaceInstance)

	// WrapInInterface creates an InterfaceInstance and increments ref count.
	// Task 9.16.2: Wrapping objects in interfaces increments ref count.
	//
	// Returns a new InterfaceInstance wrapping the object.
	WrapInInterface(iface InterfaceInfo, obj *ObjectInstance) *InterfaceInstance

	// SetDestructorCallback registers the callback for destructor invocation.
	// The callback is invoked when RefCount reaches 0.
	//
	// Signature: func(obj *ObjectInstance) error
	// The callback should:
	//   1. Look up the "Destroy" method in obj.Class
	//   2. Execute the destructor in the interpreter
	//   3. Return any error from destructor execution
	SetDestructorCallback(callback DestructorCallback)
}

// DestructorCallback is invoked when an object's reference count reaches 0.
// Task 3.5.39-3.5.40: Callback pattern avoids runtime importing interpreter.
//
// The callback receives the object and should:
//  1. Check if obj.Destroyed is true (skip if already destroyed)
//  2. Look up the "Destroy" method in obj.Class
//  3. Mark obj.Destroyed = true BEFORE execution (prevent recursion)
//  4. Execute the destructor method
//  5. Reset obj.RefCount = 0 after completion
//  6. Return any error from execution
//
// The interpreter will provide this callback during initialization.
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

	// Decrement reference count
	obj.RefCount--
	if obj.RefCount < 0 {
		obj.RefCount = 0
	}

	// Invoke destructor if ref count reaches 0
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
	// Create interface instance
	intf := &InterfaceInstance{
		Interface: iface,
		Object:    obj,
	}

	// Increment ref count (interface takes ownership)
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
