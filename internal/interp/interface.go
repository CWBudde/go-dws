package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Interface Runtime Metadata
// ============================================================================

type InterfaceInfo = runtime.MutableInterfaceInfo

var NewInterfaceInfo = runtime.NewMutableInterfaceInfo

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
	targetMethods := target.AllMethods()

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

	i.refCountManager().DecrementRef(obj)
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
// Task 4.0.10: Skip "Result" variable - it's returned to caller and shouldn't be released here.
func (i *Interpreter) cleanupInterfaceReferences(env *Environment) {
	if env == nil {
		return
	}

	// Iterate through all variables in the environment
	env.Range(func(name string, value Value) bool {
		// Skip ReferenceValue entries (like function name aliases)
		if _, isRef := value.(*ReferenceValue); isRef {
			return true // continue
		}

		// Task 4.0.10: Skip "Result" variable during function cleanup.
		// Result is the return value and will be managed by the caller.
		// Releasing it here would cause premature destructor calls.
		if name == "Result" {
			return true // continue (skip)
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
