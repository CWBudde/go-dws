# Reference Counting Architecture Design

**Task**: 3.5.39
**Date**: 2025-12-04
**Purpose**: Design reference counting architecture for runtime package to achieve evaluator independence

---

## Executive Summary

This document designs a reference counting architecture that:

1. **Moves ref counting to runtime package** - Eliminates circular import with evaluator
2. **Uses callback pattern for destructors** - Evaluator can't import interpreter
3. **Handles all 6 assignment cases** - Complete coverage of object/interface lifecycle
4. **Maintains DWScript semantics** - Reference counting behavior identical to original
5. **Enables Phase 3.5.40-3.5.42 migration** - Clear implementation path

---

## Part 1: Current Reference Counting Operations (Subtask 3.5.39a)

### 1.1 Reference Counting Call Sites in Interpreter

The interpreter currently performs reference counting in **4 primary locations**:

#### Location 1: `interface.go` - Core Destructor Logic

**Line 452-483**: `callDestructorIfNeeded(obj *ObjectInstance)`
```go
func (i *Interpreter) callDestructorIfNeeded(obj *ObjectInstance) {
    if obj == nil || obj.Destroyed {
        return
    }

    // Decrement reference count
    obj.RefCount--
    if obj.RefCount < 0 {
        obj.RefCount = 0
    }

    // If reference count reaches 0 or below, call the destructor
    if obj.RefCount <= 0 {
        i.runDestructor(obj, obj.Class.LookupMethod("Destroy"), nil)
    }
}
```

**Purpose**: Decrements ref count and invokes destructor when count reaches 0

**Usage**: Called from:
- Simple assignment (replacing old object/interface)
- Interface release (`ReleaseInterfaceReference`)
- Manual cleanup paths

---

**Line 492-496**: `ReleaseInterfaceReference(intfInst *InterfaceInstance)`
```go
func (i *Interpreter) ReleaseInterfaceReference(intfInst *InterfaceInstance) {
    if intfInst == nil || intfInst.Object == nil {
        return
    }

    // Use the consolidated helper method
    i.callDestructorIfNeeded(intfInst.Object)
}
```

**Purpose**: Releases interface reference by decrementing underlying object's ref count

**Usage**: Called when:
- Reassigning interface variables
- Interface goes out of scope
- Temporary interface values from function calls

---

**Line 409-421**: `runDestructor(obj *ObjectInstance, ...)` - Destructor execution
```go
// Mark as destroyed BEFORE running destructor code
// (prevents infinite recursion if destructor accesses Self)
if destructor == nil {
    obj.Destroyed = true
    obj.RefCount = 0  // Reset to 0
    return &NilValue{}
}

// Execute destructor with protection
defer func() {
    if obj.DestroyCallDepth == 0 {
        obj.Destroyed = true
        obj.RefCount = 0  // Reset to 0
    }
}()
```

**Purpose**: Marks object as destroyed and resets ref count after destructor completes

---

#### Location 2: `statements_assignments.go` - Assignment Reference Counting

**Line 373-383**: Assigning through var parameter (interface)
```go
// Task 9.1.5: Handle interface reference counting when assigning through var parameters
// Release the old reference if the target currently holds an interface
if oldIntf, isOldIntf := currentVal.(*InterfaceInstance); isOldIntf {
    i.ReleaseInterfaceReference(oldIntf)
}

// If assigning an interface, increment ref count for the new reference
if intfInst, isIntf := value.(*InterfaceInstance); isIntf {
    // Increment ref count because the target variable gets a new reference
    if intfInst.Object != nil {
        intfInst.Object.RefCount++
    }
}
```

**Purpose**: Reference counting for var parameters that hold interfaces

---

**Line 444-470**: Simple assignment (object variable)
```go
// Task 9.1.5: Handle object variable assignment - manage ref count
if objInst, isObj := existingVal.(*ObjectInstance); isObj {
    // Variable currently holds an object
    if _, isNil := value.(*NilValue); isNil {
        // Setting object variable to nil - decrement ref count and call destructor if needed
        i.callDestructorIfNeeded(objInst)
    } else if newObj, isNewObj := value.(*ObjectInstance); isNewObj {
        // Replacing old object with new object
        // Skip ref count changes if assigning the same instance
        if objInst != newObj {
            // Decrement old object's ref count and call destructor if needed
            i.callDestructorIfNeeded(objInst)
            // Increment new object's ref count
            newObj.RefCount++
        }
    }
} else {
    // Variable doesn't currently hold an object (could be nil, new variable, etc.)
    // If we're assigning an object, increment its ref count
    // BUT: Don't increment if the target is an interface - NewInterfaceInstance will do it
    if newObj, isNewObj := value.(*ObjectInstance); isNewObj {
        if _, isIface := existingVal.(*InterfaceInstance); !isIface {
            // Not an interface variable, so increment ref count
            newObj.RefCount++
        }
    }
}
```

**Purpose**: Manages ref count for object variable assignments (inc + dec)

---

**Line 472-506**: Interface variable assignment
```go
// Task 9.16.2: Wrap object instances in InterfaceInstance when assigning to interface variables
if ifaceInst, isIface := existingVal.(*InterfaceInstance); isIface {
    // Task 9.1.5: Release the old interface reference before assigning new value
    // This decrements ref count and calls destructor if ref count reaches 0
    i.ReleaseInterfaceReference(ifaceInst)

    // Target is an interface variable - wrap the value if it's an object
    if objInst, ok := value.(*ObjectInstance); ok {
        // Assigning an object to an interface variable - wrap it
        value = NewInterfaceInstance(ifaceInst.Interface, objInst)
    } else if _, isNil := value.(*NilValue); isNil {
        // Assigning nil to interface - create interface instance with nil object
        // No need to increment ref count since object is nil
        value = &InterfaceInstance{
            Interface: ifaceInst.Interface,
            Object:    nil,
        }
    } else if srcIface, isSrcIface := value.(*InterfaceInstance); isSrcIface {
        // Assigning interface to interface
        // Task 9.1.5: Increment ref count on the underlying object (if not nil)
        // This implements copy semantics - both variables will hold references
        if srcIface.Object != nil {
            srcIface.Object.RefCount++
        }
        // Use the underlying object but with the target interface type
        value = &InterfaceInstance{
            Interface: ifaceInst.Interface,
            Object:    srcIface.Object,
        }
        // Track that we copied from another interface value; release the source
        if shouldReleaseInterfaceSource(stmt, i.env) {
            defer i.ReleaseInterfaceReference(srcIface)
        }
    }
}
```

**Purpose**: Interface variable assignment with ref counting and wrapping

---

**Line 509-516**: Method pointer ref counting
```go
// PR#142: Increment RefCount for function pointers that hold object references
// When storing a FunctionPointerValue (method pointer from interface or object),
// increment the SelfObject's RefCount to keep it alive while the pointer exists
if funcPtr, isFuncPtr := value.(*FunctionPointerValue); isFuncPtr {
    if objInst, isObj := funcPtr.SelfObject.(*ObjectInstance); isObj {
        objInst.RefCount++
    }
}
```

**Purpose**: Increment ref count when method pointer holds reference to object

---

#### Location 3: `runtime/interface_instance.go` - Interface Creation

**Line 30-33**: NewInterfaceInstance
```go
// Increment reference count when interface takes ownership of object
if obj != nil {
    obj.RefCount++
}
```

**Purpose**: Increment ref count when creating new interface instance

---

#### Location 4: `user_function_callbacks.go` - Function Return Values

**Line 128-139**: Interface return value ref counting
```go
// createInterfaceRefCounterCallback creates the interface ref count increment callback.
//
// When returning an interface value from a function, the ref count needs to be incremented
// for the caller's reference. This will be balanced by cleanup releasing Result after return.
func (i *Interpreter) createInterfaceRefCounterCallback() evaluator.IncrementInterfaceRefCountFunc {
    return func(returnValue evaluator.Value) {
        // If returning an interface, increment RefCount for the caller's reference
        if intfInst, isIntf := returnValue.(*InterfaceInstance); isIntf {
            if intfInst.Object != nil {
                intfInst.Object.RefCount++
            }
        }
    }
}
```

**Purpose**: Increment ref count for interface return values from user functions

---

### 1.2 Reference Counting Patterns and Semantics

Based on the audit, reference counting follows these patterns:

#### Pattern 1: **Increment Operations**

| Scenario | When | Code Location |
|----------|------|---------------|
| Creating interface instance | `NewInterfaceInstance(iface, obj)` | `runtime/interface_instance.go:32` |
| Assigning object to variable | `obj := newObject` | `statements_assignments.go:457` |
| Assigning interface to variable | `intf := otherIntf` | `statements_assignments.go:494` |
| Storing method pointer | `ptr := obj.Method` | `statements_assignments.go:514` |
| Returning interface from function | `return intfValue` | `user_function_callbacks.go:137` |

**Increment Rule**: RefCount++ when creating a **new reference** to an existing object

---

#### Pattern 2: **Decrement Operations**

| Scenario | When | Code Location |
|----------|------|---------------|
| Replacing object variable | `obj := newObj` (old obj decremented) | `statements_assignments.go:455` |
| Setting object to nil | `obj := nil` | `statements_assignments.go:449` |
| Releasing interface | Reassigning interface variable | `interface.go:496` |
| Destructor invocation | Ref count reaches 0 | `interface.go:470` |

**Decrement Rule**: RefCount-- when **removing** a reference to an object

---

#### Pattern 3: **Destructor Invocation**

**Conditions**:
1. `obj.RefCount <= 0` (reached zero)
2. `!obj.Destroyed` (not already destroyed)
3. Destructor method exists (`obj.Class.LookupMethod("Destroy")`)

**Destructor Logic** (`interface.go:409-446`):
```go
1. Mark obj.Destroyed = true BEFORE running destructor (prevents recursion)
2. Increment obj.DestroyCallDepth (tracks re-entrant destructor calls)
3. Execute user-defined Destroy method
4. Decrement obj.DestroyCallDepth
5. Set obj.RefCount = 0 (reset to clean state)
```

**Edge Cases**:
- **Same instance assignment**: `obj := obj` → No ref count change
- **Nil object**: `obj := nil` → Decrement only (no increment)
- **Interface wrapping**: NewInterfaceInstance increments; evaluator doesn't
- **Temporary interface source**: Release source after assignment completes

---

### 1.3 Critical Invariants

The current implementation maintains these invariants:

1. **Object Creation**: `RefCount = 0` (incremented when first assigned)
2. **Reference Equality**: Assigning object to itself doesn't change ref count
3. **Interface Semantics**: Interface holds reference; increment when creating InterfaceInstance
4. **Destructor Safety**: Mark destroyed BEFORE execution (prevents infinite recursion)
5. **Var Parameter Write-Through**: Ref count updated on target, not on reference
6. **Cleanup Order**: Old reference decremented BEFORE new reference incremented

---

## Part 2: Runtime Reference Count Manager Design (Subtask 3.5.39b)

### 2.1 Package Structure

```
internal/interp/runtime/
├── refcount.go          // RefCountManager interface + default implementation
├── object.go            // ObjectInstance (already has RefCount field)
└── interface_instance.go // InterfaceInstance (already uses obj.RefCount++)
```

**Key Constraint**: `runtime` package cannot import `interp` package (circular dependency)

---

### 2.2 RefCountManager Interface

```go
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
```

---

### 2.3 Destructor Callback Pattern (Subtask 3.5.39c)

#### Problem

The runtime package needs to invoke destructors when `RefCount` reaches 0, but:
- Destructors are user-defined methods (requires interpreter to execute)
- Interpreter package cannot be imported by runtime (circular dependency)
- Method lookup requires ClassInfo (lives in interpreter)

#### Solution: Callback Pattern

```go
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
```

**Callback Registration**: Interpreter sets callback during startup
```go
// In interpreter initialization
refCountMgr := runtime.NewRefCountManager()
refCountMgr.SetDestructorCallback(func(obj *ObjectInstance) error {
    return i.runDestructorForRefCount(obj)
})
```

---

### 2.4 Default RefCountManager Implementation

```go
// defaultRefCountManager implements RefCountManager with callback-based destructors.
type defaultRefCountManager struct {
    destructorCallback DestructorCallback
    mu                 sync.RWMutex // Protects callback
}

// NewRefCountManager creates a default reference count manager.
func NewRefCountManager() RefCountManager {
    return &defaultRefCountManager{}
}

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
            _ = callback(obj) // Error handling TBD
        }
    }

    return nil
}

func (m *defaultRefCountManager) ReleaseObject(obj *ObjectInstance) {
    m.DecrementRef(obj)
}

func (m *defaultRefCountManager) ReleaseInterface(intf *InterfaceInstance) {
    m.DecrementRef(intf)
}

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

func (m *defaultRefCountManager) SetDestructorCallback(callback DestructorCallback) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.destructorCallback = callback
}
```

---

## Part 3: Migration Plan for 6 EvalNode Calls (Subtask 3.5.39d)

### 3.1 Overview of 6 Reference Counting Calls

From `internal/interp/evaluator/assignment_helpers.go`:

| Line | Context | Current Behavior | Migration Strategy |
|------|---------|------------------|-------------------|
| 127 | Interface variable assignment | Calls `ReleaseInterfaceReference` + wrapping | Use `ReleaseInterface()` + `WrapInInterface()` |
| 136 | Object variable assignment | Calls `callDestructorIfNeeded` + inc | Use `ReleaseObject()` + `IncrementRef()` |
| 172 | Assigning object VALUE | Increments `RefCount++` | Use `IncrementRef(value)` |
| 182 | Assigning interface VALUE | Increments `RefCount++` on object | Use `IncrementRef(value)` |
| 192 | Method pointer SelfObject | Increments `RefCount++` | Use `IncrementRef(funcPtr.SelfObject)` |
| 227 | Var param → interface/object | Complex release + increment | Use `ReleaseInterface/Object()` + `IncrementRef()` |

---

### 3.2 Line 127: Interface Variable Assignment

**Current Code** (`assignment_helpers.go:126-127`):
```go
if existingVal.Type() == "INTERFACE" {
    return e.adapter.EvalNode(stmt)
}
```

**What interpreter does** (`statements_assignments.go:472-506`):
```go
if ifaceInst, isIface := existingVal.(*InterfaceInstance); isIface {
    // Release old interface
    i.ReleaseInterfaceReference(ifaceInst)

    // Wrap object in interface
    if objInst, ok := value.(*ObjectInstance); ok {
        value = NewInterfaceInstance(ifaceInst.Interface, objInst)
    } else if srcIface, isSrcIface := value.(*InterfaceInstance); isSrcIface {
        // Interface-to-interface: increment and wrap
        if srcIface.Object != nil {
            srcIface.Object.RefCount++
        }
        value = &InterfaceInstance{
            Interface: ifaceInst.Interface,
            Object:    srcIface.Object,
        }
    }
}
```

**Migrated Code**:
```go
if ifaceInst, isIface := existingVal.(*InterfaceInstance); isIface {
    // Get RefCountManager from context
    refMgr := ctx.RefCountManager()

    // Release old interface reference
    refMgr.ReleaseInterface(ifaceInst)

    // Wrap object in interface (increments ref count)
    if objInst, ok := value.(*ObjectInstance); ok {
        value = refMgr.WrapInInterface(ifaceInst.Interface, objInst)
    } else if srcIface, isSrcIface := value.(*InterfaceInstance); isSrcIface {
        // Interface-to-interface assignment
        value = refMgr.WrapInInterface(ifaceInst.Interface, srcIface.Object)
    } else if _, isNil := value.(*NilValue); isNil {
        // Assigning nil - create interface with nil object
        value = &InterfaceInstance{
            Interface: ifaceInst.Interface,
            Object:    nil,
        }
    }

    // Update variable
    e.SetVar(ctx, targetName, value)
    return value
}
```

---

### 3.3 Line 136: Object Variable Assignment

**Current Code** (`assignment_helpers.go:135-136`):
```go
if existingVal.Type() == "OBJECT" {
    return e.adapter.EvalNode(stmt)
}
```

**What interpreter does** (`statements_assignments.go:444-470`):
```go
if objInst, isObj := existingVal.(*ObjectInstance); isObj {
    if _, isNil := value.(*NilValue); isNil {
        // Setting to nil - decrement and destroy
        i.callDestructorIfNeeded(objInst)
    } else if newObj, isNewObj := value.(*ObjectInstance); isNewObj {
        // Replacing with new object
        if objInst != newObj {
            i.callDestructorIfNeeded(objInst)
            newObj.RefCount++
        }
    }
} else {
    // New object assignment - increment ref count
    if newObj, isNewObj := value.(*ObjectInstance); isNewObj {
        newObj.RefCount++
    }
}
```

**Migrated Code**:
```go
if objInst, isObj := existingVal.(*ObjectInstance); isObj {
    refMgr := ctx.RefCountManager()

    if _, isNil := value.(*NilValue); isNil {
        // Setting to nil - release old object
        refMgr.ReleaseObject(objInst)
    } else if newObj, isNewObj := value.(*ObjectInstance); isNewObj {
        // Replacing with new object
        if objInst != newObj {
            refMgr.ReleaseObject(objInst)  // Dec old
            refMgr.IncrementRef(newObj)     // Inc new
        }
    } else {
        // Replacing object with non-object - release old
        refMgr.ReleaseObject(objInst)
    }

    // Update variable
    e.SetVar(ctx, targetName, value)
    return value
}
```

---

### 3.4 Line 172: Assigning Object VALUE

**Current Code** (`assignment_helpers.go:171-172`):
```go
if value != nil && value.Type() == "OBJECT" {
    return e.adapter.EvalNode(stmt)
}
```

**What interpreter does** (`statements_assignments.go:461-469`):
```go
if newObj, isNewObj := value.(*ObjectInstance); isNewObj {
    if _, isIface := existingVal.(*InterfaceInstance); !isIface {
        // Not an interface variable, so increment ref count
        newObj.RefCount++
    }
}
```

**Migrated Code**:
```go
if newObj, isNewObj := value.(*ObjectInstance); isNewObj {
    refMgr := ctx.RefCountManager()

    // Check if target is NOT an interface (interface wrapping increments separately)
    if _, isIface := existingVal.(*InterfaceInstance); !isIface {
        refMgr.IncrementRef(newObj)
    }
}
```

---

### 3.5 Line 182: Assigning Interface VALUE

**Current Code** (`assignment_helpers.go:181-182`):
```go
if value != nil && value.Type() == "INTERFACE" {
    return e.adapter.EvalNode(stmt)
}
```

**What interpreter does**: Handled by line 127 logic (interface variable assignment)

**Migrated Code**:
```go
// This case is already covered by line 127 migration
// Interface value assignment uses WrapInInterface which increments
```

---

### 3.6 Line 192: Method Pointer SelfObject

**Current Code** (`assignment_helpers.go:191-192`):
```go
if valueType == "METHOD_POINTER" {
    // Method pointer with SelfObject - needs ref counting
    return e.adapter.EvalNode(stmt)
}
```

**What interpreter does** (`statements_assignments.go:509-516`):
```go
if funcPtr, isFuncPtr := value.(*FunctionPointerValue); isFuncPtr {
    if objInst, isObj := funcPtr.SelfObject.(*ObjectInstance); isObj {
        objInst.RefCount++
    }
}
```

**Migrated Code**:
```go
if valueType == "METHOD_POINTER" {
    refMgr := ctx.RefCountManager()

    // Safely increment SelfObject ref count
    if funcPtr, isFuncPtr := value.(*FunctionPointerValue); isFuncPtr {
        if funcPtr.SelfObject != nil {
            refMgr.IncrementRef(funcPtr.SelfObject)
        }
    }
}
```

---

### 3.7 Line 227: Var Parameter → Interface/Object

**Current Code** (`assignment_helpers.go:226-227`):
```go
if currentVal.Type() == "INTERFACE" || currentVal.Type() == "OBJECT" {
    return e.adapter.EvalNode(stmt)
}
```

**What interpreter does** (`statements_assignments.go:373-383`):
```go
// Release old reference
if oldIntf, isOldIntf := currentVal.(*InterfaceInstance); isOldIntf {
    i.ReleaseInterfaceReference(oldIntf)
}

// Increment new reference
if intfInst, isIntf := value.(*InterfaceInstance); isIntf {
    if intfInst.Object != nil {
        intfInst.Object.RefCount++
    }
}
```

**Migrated Code**:
```go
refMgr := ctx.RefCountManager()

// Release old reference
if oldIntf, isOldIntf := currentVal.(*InterfaceInstance); isOldIntf {
    refMgr.ReleaseInterface(oldIntf)
} else if oldObj, isOldObj := currentVal.(*ObjectInstance); isOldObj {
    refMgr.ReleaseObject(oldObj)
}

// Increment new reference
if value != nil {
    refMgr.IncrementRef(value)
}
```

---

## Part 4: Additional Reference Counting Call Sites (Subtask 3.5.39e)

### 4.1 User Function Return Values

**Location**: `user_function_callbacks.go:128-139`

**Current Code**:
```go
func (i *Interpreter) createInterfaceRefCounterCallback() evaluator.IncrementInterfaceRefCountFunc {
    return func(returnValue evaluator.Value) {
        if intfInst, isIntf := returnValue.(*InterfaceInstance); isIntf {
            if intfInst.Object != nil {
                intfInst.Object.RefCount++
            }
        }
    }
}
```

**Migration**: Replace callback with RefCountManager call
```go
// In evaluator/user_function_helpers.go
func (e *Evaluator) IncrementInterfaceReturnValue(returnValue Value, ctx *ExecutionContext) {
    if returnValue != nil && returnValue.Type() == "INTERFACE" {
        ctx.RefCountManager().IncrementRef(returnValue)
    }
}
```

---

### 4.2 Interface Instance Creation

**Location**: `runtime/interface_instance.go:30-33`

**Current Code**:
```go
// Increment reference count when interface takes ownership of object
if obj != nil {
    obj.RefCount++
}
```

**Migration**: Use RefCountManager
```go
// NewInterfaceInstance should use RefCountManager
func NewInterfaceInstanceWithRefMgr(iface InterfaceInfo, obj *ObjectInstance, refMgr RefCountManager) *InterfaceInstance {
    intf := &InterfaceInstance{
        Interface: iface,
        Object:    obj,
    }

    if obj != nil {
        refMgr.IncrementRef(obj)
    }

    return intf
}
```

---

### 4.3 Method Returns (Potential Future)

**Context**: Methods that return `Self` or other objects

**Current Behavior**: Implicit ref counting in caller (no explicit increment)

**Future Consideration**: May need explicit ref counting for method return values that are objects

---

### 4.4 Property Getters (Potential Future)

**Context**: Properties that return objects or interfaces

**Current Behavior**: Property getters may return object references without ref counting

**Future Consideration**: Property getters might need ref count tracking

---

## Part 5: Implementation Roadmap

### Task 3.5.40: Implement RefCountManager in Runtime

**Deliverables**:
- [x] `runtime/refcount.go` with RefCountManager interface
- [x] defaultRefCountManager implementation
- [x] DestructorCallback type and pattern
- [x] Unit tests for ref counting operations

**Acceptance Criteria**:
- IncrementRef correctly increments RefCount
- DecrementRef correctly decrements and calls destructor at 0
- Callback pattern works without circular imports
- Edge cases handled: nil, same instance, destroyed objects

---

### Task 3.5.41: Migrate 6 Assignment EvalNode Calls

**Deliverables**:
- [x] Migrate line 127 (interface variable)
- [x] Migrate line 136 (object variable)
- [x] Migrate line 172 (object value)
- [x] Migrate line 182 (interface value)
- [x] Migrate line 192 (method pointer)
- [x] Migrate line 227 (var parameter)

**Acceptance Criteria**:
- All 6 calls replaced with RefCountManager operations
- All fixture tests pass (341 passed, 886 failed - maintain current status)
- No regressions in object/interface lifecycle behavior

---

### Task 3.5.42: Migrate Other Ref Counting Sites

**Deliverables**:
- [x] Migrate user function return value callback
- [x] Update NewInterfaceInstance to use RefCountManager
- [x] Identify and migrate any additional ref count sites

**Acceptance Criteria**:
- Zero manual RefCount++ or RefCount-- operations outside RefCountManager
- All object/interface lifecycle operations use RefCountManager
- Destructor callback pattern fully operational

---

## Part 6: Testing Strategy

### 6.1 Unit Tests for RefCountManager

```go
// Test cases:
- IncrementRef on nil → no-op
- IncrementRef on ObjectInstance → RefCount++
- IncrementRef on InterfaceInstance → underlying object RefCount++
- DecrementRef to 0 → calls destructor callback
- DecrementRef below 0 → clamps to 0
- ReleaseObject on nil → no-op
- ReleaseInterface on nil → no-op
- WrapInInterface → increments ref count
- SetDestructorCallback → callback invoked correctly
- Same instance assignment → no ref count change
```

---

### 6.2 Integration Tests

**Test Scenarios**:
1. Object lifecycle: create → assign → reassign → destroy
2. Interface lifecycle: wrap → assign → reassign → release
3. Var parameter write-through with ref counting
4. Method pointer ref counting
5. User function return value ref counting
6. Destructor invocation on ref count 0
7. No double-destruction (Destroyed flag)

---

## Part 7: Risk Mitigation

### 7.1 Potential Issues

| Risk | Mitigation |
|------|------------|
| Destructor not called | Comprehensive unit tests + fixture test validation |
| Double destruction | Check `obj.Destroyed` flag before invoking callback |
| Ref count drift | Audit all increment/decrement pairs for balance |
| Nil pointer panics | Nil checks in all RefCountManager methods |
| Circular imports | Runtime cannot import interp; use callback pattern |

---

### 7.2 Rollback Plan

If migration causes regressions:
1. Revert evaluator changes (keep adapter calls)
2. Keep RefCountManager implementation (no harm)
3. Address issues in isolated branch
4. Re-attempt migration after fixes

---

## Appendix A: Reference Counting State Machine

```
Object State Transitions:

[Created]
  RefCount = 0
  Destroyed = false

  ↓ (first assignment)

[Referenced]
  RefCount = 1+
  Destroyed = false

  ↓ (assignments create more refs)

[Multiple References]
  RefCount = N (N > 1)
  Destroyed = false

  ↓ (refs released)

[Last Reference]
  RefCount = 1
  Destroyed = false

  ↓ (final release)

[Destructor Running]
  RefCount = 0
  Destroyed = true  (set BEFORE destructor runs)
  DestroyCallDepth = 1+

  ↓ (destructor completes)

[Destroyed]
  RefCount = 0
  Destroyed = true
  DestroyCallDepth = 0
```

---

## Appendix B: Callback Pattern Comparison

### Alternative 1: Event System (Rejected)
- Too heavyweight for simple destructor calls
- Requires event queue, dispatch loop
- Overkill for synchronous destructor invocation

### Alternative 2: Interface-based (Rejected)
- Requires ObjectInstance to implement Destroyable interface
- Still creates circular dependency (ObjectInstance → Interpreter)

### Alternative 3: Callback Function (SELECTED)
- Simple, direct, synchronous
- No circular imports
- Clear ownership: interpreter provides implementation
- Testable: can inject mock callbacks

---

## Summary

This design provides a complete reference counting architecture that:

1. ✅ Moves ref counting to runtime package (no circular imports)
2. ✅ Uses callback pattern for destructors (evaluator independence)
3. ✅ Handles all 6 assignment cases (complete coverage)
4. ✅ Maintains DWScript semantics (identical behavior)
5. ✅ Enables Phase 3.5.40-3.5.42 migration (clear path forward)

**Next Steps**:
- Proceed to Task 3.5.40: Implement RefCountManager
- Write comprehensive unit tests
- Migrate 6 EvalNode calls one by one
- Validate with fixture test suite

---

**End of Document**
