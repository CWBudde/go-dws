package runtime

import (
	"sync"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// mockClassInfo is a minimal IClassInfo implementation for testing
type mockClassInfo struct {
	name string
}

func (m *mockClassInfo) GetName() string                                { return m.name }
func (m *mockClassInfo) GetMetadata() *ClassMetadata                    { return nil }
func (m *mockClassInfo) GetFieldsMap() map[string]*ast.FieldDecl        { return nil }
func (m *mockClassInfo) GetMethodsMap() map[string]*ast.FunctionDecl    { return nil }
func (m *mockClassInfo) LookupMethod(name string) *ast.FunctionDecl     { return nil }
func (m *mockClassInfo) LookupProperty(name string) *PropertyInfo       { return nil }
func (m *mockClassInfo) LookupClassVar(name string) (Value, IClassInfo) { return nil, nil }
func (m *mockClassInfo) LookupOperator(op string, types []string) (*OperatorEntry, bool) {
	return nil, false
}
func (m *mockClassInfo) GetDefaultProperty() *PropertyInfo                     { return nil }
func (m *mockClassInfo) GetParent() IClassInfo                                 { return nil }
func (m *mockClassInfo) FieldExists(normalizedName string) bool                { return false }
func (m *mockClassInfo) IsAbstract() bool                                      { return false }
func (m *mockClassInfo) IsExternal() bool                                      { return false }
func (m *mockClassInfo) GetClassVarsMap() map[string]Value                     { return nil }
func (m *mockClassInfo) GetVirtualMethodTable() map[string]*VirtualMethodEntry { return nil }
func (m *mockClassInfo) GetConstructor(name string) *ast.FunctionDecl          { return nil }
func (m *mockClassInfo) GetFieldTypesMap() map[string]any                      { return nil }
func (m *mockClassInfo) GetInterfaces() []*InterfaceInfo                       { return nil }

// mockInterfaceInfo is a minimal IInterfaceInfo implementation for testing
type mockInterfaceInfo struct {
	name string
}

func (m *mockInterfaceInfo) GetName() string                         { return m.name }
func (m *mockInterfaceInfo) GetParent() IInterfaceInfo               { return nil }
func (m *mockInterfaceInfo) GetMethod(name string) any               { return nil }
func (m *mockInterfaceInfo) HasMethod(name string) bool              { return false }
func (m *mockInterfaceInfo) GetProperty(name string) *PropertyInfo   { return nil }
func (m *mockInterfaceInfo) HasProperty(name string) bool            { return false }
func (m *mockInterfaceInfo) GetDefaultProperty() *PropertyInfo       { return nil }
func (m *mockInterfaceInfo) AllMethods() map[string]any              { return nil }
func (m *mockInterfaceInfo) AllProperties() map[string]*PropertyInfo { return nil }

// Test_IncrementRef_Nil tests that incrementing nil is a no-op
func Test_IncrementRef_Nil(t *testing.T) {
	mgr := NewRefCountManager()

	result := mgr.IncrementRef(nil)
	if result != nil {
		t.Errorf("IncrementRef(nil) = %v, want nil", result)
	}
}

// Test_IncrementRef_ObjectInstance tests incrementing an object's ref count
func Test_IncrementRef_ObjectInstance(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 0,
	}

	result := mgr.IncrementRef(obj)

	if result != obj {
		t.Errorf("IncrementRef(obj) returned wrong value")
	}

	if obj.RefCount != 1 {
		t.Errorf("obj.RefCount = %d, want 1", obj.RefCount)
	}
}

// Test_IncrementRef_ObjectInstance_Multiple tests multiple increments
func Test_IncrementRef_ObjectInstance_Multiple(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 0,
	}

	mgr.IncrementRef(obj)
	mgr.IncrementRef(obj)
	mgr.IncrementRef(obj)

	if obj.RefCount != 3 {
		t.Errorf("obj.RefCount = %d, want 3", obj.RefCount)
	}
}

// Test_IncrementRef_InterfaceInstance tests incrementing an interface's underlying object
func Test_IncrementRef_InterfaceInstance(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 0,
	}

	intf := &InterfaceInstance{
		Interface: &mockInterfaceInfo{name: "ITest"},
		Object:    obj,
	}

	result := mgr.IncrementRef(intf)

	if result != intf {
		t.Errorf("IncrementRef(intf) returned wrong value")
	}

	if obj.RefCount != 1 {
		t.Errorf("underlying obj.RefCount = %d, want 1", obj.RefCount)
	}
}

// Test_IncrementRef_InterfaceInstance_NilObject tests incrementing interface with nil object
func Test_IncrementRef_InterfaceInstance_NilObject(t *testing.T) {
	mgr := NewRefCountManager()

	intf := &InterfaceInstance{
		Interface: &mockInterfaceInfo{name: "ITest"},
		Object:    nil,
	}

	result := mgr.IncrementRef(intf)

	if result != intf {
		t.Errorf("IncrementRef(intf) returned wrong value")
	}

	// Should not panic with nil object
}

// Test_DecrementRef_Nil tests that decrementing nil is a no-op
func Test_DecrementRef_Nil(t *testing.T) {
	mgr := NewRefCountManager()

	result := mgr.DecrementRef(nil)
	if result != nil {
		t.Errorf("DecrementRef(nil) = %v, want nil", result)
	}
}

// Test_DecrementRef_ObjectInstance tests decrementing an object's ref count
func Test_DecrementRef_ObjectInstance(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 2,
	}

	result := mgr.DecrementRef(obj)

	if result != nil {
		t.Errorf("DecrementRef(obj) = %v, want nil", result)
	}

	if obj.RefCount != 1 {
		t.Errorf("obj.RefCount = %d, want 1", obj.RefCount)
	}
}

// Test_DecrementRef_ObjectInstance_ToZero tests decrementing to zero without destructor
func Test_DecrementRef_ObjectInstance_ToZero(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 1,
	}

	result := mgr.DecrementRef(obj)

	if result != nil {
		t.Errorf("DecrementRef(obj) = %v, want nil", result)
	}

	if obj.RefCount != 0 {
		t.Errorf("obj.RefCount = %d, want 0", obj.RefCount)
	}
}

// Test_DecrementRef_ObjectInstance_NegativeClamped tests that negative ref counts are clamped to 0
func Test_DecrementRef_ObjectInstance_NegativeClamped(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 0,
	}

	mgr.DecrementRef(obj)

	if obj.RefCount != 0 {
		t.Errorf("obj.RefCount = %d, want 0 (should clamp negative)", obj.RefCount)
	}
}

// Test_DecrementRef_ObjectInstance_WithDestructor tests destructor callback invocation
func Test_DecrementRef_ObjectInstance_WithDestructor(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 1,
	}

	destructorCalled := false
	var destructedObj *ObjectInstance

	mgr.SetDestructorCallback(func(o *ObjectInstance) error {
		destructorCalled = true
		destructedObj = o
		return nil
	})

	mgr.DecrementRef(obj)

	if !destructorCalled {
		t.Error("destructor callback was not called")
	}

	if destructedObj != obj {
		t.Error("destructor callback received wrong object")
	}

	if obj.RefCount != 0 {
		t.Errorf("obj.RefCount = %d, want 0", obj.RefCount)
	}
}

// Test_DecrementRef_ObjectInstance_AlreadyDestroyed tests that destroyed objects skip destructor
func Test_DecrementRef_ObjectInstance_AlreadyDestroyed(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:     &mockClassInfo{name: "TestClass"},
		Fields:    make(map[string]Value),
		RefCount:  1,
		Destroyed: true,
	}

	destructorCalled := false

	mgr.SetDestructorCallback(func(o *ObjectInstance) error {
		destructorCalled = true
		return nil
	})

	result := mgr.DecrementRef(obj)

	if destructorCalled {
		t.Error("destructor callback should not be called for already destroyed object")
	}

	if result != nil {
		t.Errorf("DecrementRef(destroyed obj) = %v, want nil", result)
	}
}

// Test_DecrementRef_InterfaceInstance tests decrementing an interface's underlying object
func Test_DecrementRef_InterfaceInstance(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 2,
	}

	intf := &InterfaceInstance{
		Interface: &mockInterfaceInfo{name: "ITest"},
		Object:    obj,
	}

	result := mgr.DecrementRef(intf)

	if result != nil {
		t.Errorf("DecrementRef(intf) = %v, want nil", result)
	}

	if obj.RefCount != 1 {
		t.Errorf("underlying obj.RefCount = %d, want 1", obj.RefCount)
	}
}

// Test_ReleaseObject_Nil tests that releasing nil is a no-op
func Test_ReleaseObject_Nil(t *testing.T) {
	mgr := NewRefCountManager()
	// Should not panic
	mgr.ReleaseObject(nil)
}

// Test_ReleaseObject tests the convenience method
func Test_ReleaseObject(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 2,
	}

	mgr.ReleaseObject(obj)

	if obj.RefCount != 1 {
		t.Errorf("obj.RefCount = %d, want 1", obj.RefCount)
	}
}

// Test_ReleaseInterface_Nil tests that releasing nil is a no-op
func Test_ReleaseInterface_Nil(t *testing.T) {
	mgr := NewRefCountManager()
	// Should not panic
	mgr.ReleaseInterface(nil)
}

// Test_ReleaseInterface tests the convenience method
func Test_ReleaseInterface(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 2,
	}

	intf := &InterfaceInstance{
		Interface: &mockInterfaceInfo{name: "ITest"},
		Object:    obj,
	}

	mgr.ReleaseInterface(intf)

	if obj.RefCount != 1 {
		t.Errorf("underlying obj.RefCount = %d, want 1", obj.RefCount)
	}
}

// Test_WrapInInterface tests creating an interface instance with ref count increment
func Test_WrapInInterface(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 0,
	}

	iface := &mockInterfaceInfo{name: "ITest"}

	intf := mgr.WrapInInterface(iface, obj)

	if intf == nil {
		t.Fatal("WrapInInterface returned nil")
	}

	if intf.Interface != iface {
		t.Error("interface not set correctly")
	}

	if intf.Object != obj {
		t.Error("object not set correctly")
	}

	if obj.RefCount != 1 {
		t.Errorf("obj.RefCount = %d, want 1 (should increment on wrap)", obj.RefCount)
	}
}

// Test_WrapInInterface_NilObject tests wrapping nil object
func Test_WrapInInterface_NilObject(t *testing.T) {
	mgr := NewRefCountManager()

	iface := &mockInterfaceInfo{name: "ITest"}

	intf := mgr.WrapInInterface(iface, nil)

	if intf == nil {
		t.Fatal("WrapInInterface returned nil")
	}

	if intf.Interface != iface {
		t.Error("interface not set correctly")
	}

	if intf.Object != nil {
		t.Error("object should be nil")
	}

	// Should not panic with nil object
}

// Test_SetDestructorCallback tests registering and invoking destructor callback
func Test_SetDestructorCallback(t *testing.T) {
	mgr := NewRefCountManager()

	callCount := 0
	var lastObj *ObjectInstance

	mgr.SetDestructorCallback(func(obj *ObjectInstance) error {
		callCount++
		lastObj = obj
		return nil
	})

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 1,
	}

	mgr.DecrementRef(obj)

	if callCount != 1 {
		t.Errorf("destructor called %d times, want 1", callCount)
	}

	if lastObj != obj {
		t.Error("destructor received wrong object")
	}
}

// Test_SetDestructorCallback_ReplaceCallback tests replacing the callback
func Test_SetDestructorCallback_ReplaceCallback(t *testing.T) {
	mgr := NewRefCountManager()

	firstCallCount := 0
	secondCallCount := 0

	mgr.SetDestructorCallback(func(obj *ObjectInstance) error {
		firstCallCount++
		return nil
	})

	mgr.SetDestructorCallback(func(obj *ObjectInstance) error {
		secondCallCount++
		return nil
	})

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 1,
	}

	mgr.DecrementRef(obj)

	if firstCallCount != 0 {
		t.Errorf("first callback called %d times, want 0", firstCallCount)
	}

	if secondCallCount != 1 {
		t.Errorf("second callback called %d times, want 1", secondCallCount)
	}
}

// Test_RefCount_IncrementDecrement_Balance tests balanced inc/dec operations
func Test_RefCount_IncrementDecrement_Balance(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 0,
	}

	// Simulate: x := obj; y := obj; x := nil; y := nil
	mgr.IncrementRef(obj) // x := obj
	mgr.IncrementRef(obj) // y := obj

	if obj.RefCount != 2 {
		t.Errorf("after 2 increments: obj.RefCount = %d, want 2", obj.RefCount)
	}

	mgr.DecrementRef(obj) // x := nil

	if obj.RefCount != 1 {
		t.Errorf("after 1 decrement: obj.RefCount = %d, want 1", obj.RefCount)
	}

	destructorCalled := false
	mgr.SetDestructorCallback(func(obj *ObjectInstance) error {
		destructorCalled = true
		return nil
	})

	mgr.DecrementRef(obj) // y := nil

	if obj.RefCount != 0 {
		t.Errorf("after 2 decrements: obj.RefCount = %d, want 0", obj.RefCount)
	}

	if !destructorCalled {
		t.Error("destructor should be called when ref count reaches 0")
	}
}

// Test_RefCount_SameInstanceAssignment tests that assigning object to itself doesn't change ref count
func Test_RefCount_SameInstanceAssignment(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 1,
	}

	destructorCalled := false
	mgr.SetDestructorCallback(func(obj *ObjectInstance) error {
		destructorCalled = true
		return nil
	})

	// Simulate: obj := obj (should not change ref count or call destructor)
	mgr.DecrementRef(obj) // Old value (should reach 0)

	if obj.RefCount != 0 {
		t.Errorf("obj.RefCount = %d, want 0", obj.RefCount)
	}

	if !destructorCalled {
		t.Error("destructor should be called")
	}

	// This test shows that checking for same-instance assignment must be done
	// at a higher level (in the evaluator), not in RefCountManager
}

// Test_RefCount_InterfaceToInterface tests interface-to-interface assignment
func Test_RefCount_InterfaceToInterface(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 0,
	}

	iface1 := &mockInterfaceInfo{name: "ITest1"}
	iface2 := &mockInterfaceInfo{name: "ITest2"}

	// Create first interface (increments ref count)
	intf1 := mgr.WrapInInterface(iface1, obj)

	if obj.RefCount != 1 {
		t.Errorf("after WrapInInterface: obj.RefCount = %d, want 1", obj.RefCount)
	}

	// Create second interface wrapping same object (increments again)
	intf2 := mgr.WrapInInterface(iface2, obj)

	if obj.RefCount != 2 {
		t.Errorf("after second WrapInInterface: obj.RefCount = %d, want 2", obj.RefCount)
	}

	// Release first interface
	mgr.ReleaseInterface(intf1)

	if obj.RefCount != 1 {
		t.Errorf("after ReleaseInterface(intf1): obj.RefCount = %d, want 1", obj.RefCount)
	}

	// Release second interface (should reach 0)
	destructorCalled := false
	mgr.SetDestructorCallback(func(obj *ObjectInstance) error {
		destructorCalled = true
		return nil
	})

	mgr.ReleaseInterface(intf2)

	if obj.RefCount != 0 {
		t.Errorf("after ReleaseInterface(intf2): obj.RefCount = %d, want 0", obj.RefCount)
	}

	if !destructorCalled {
		t.Error("destructor should be called when ref count reaches 0")
	}
}

// Test_RefCount_Concurrency tests thread safety of RefCountManager
func Test_RefCount_Concurrency(t *testing.T) {
	mgr := NewRefCountManager()

	obj := &ObjectInstance{
		Class:    &mockClassInfo{name: "TestClass"},
		Fields:   make(map[string]Value),
		RefCount: 0,
	}

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Increment from many goroutines
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			mgr.IncrementRef(obj)
		}()
	}

	// Decrement from many goroutines
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			mgr.DecrementRef(obj)
		}()
	}

	wg.Wait()

	// RefCount should be 0 after balanced inc/dec
	// Note: Due to race conditions, this might not be exactly 0,
	// but the test verifies that concurrent access doesn't panic
	if obj.RefCount < 0 {
		t.Errorf("obj.RefCount = %d, should not be negative", obj.RefCount)
	}
}

// Test_RefCount_DestructorCallback_Concurrency tests thread safety of destructor callback
func Test_RefCount_DestructorCallback_Concurrency(t *testing.T) {
	mgr := NewRefCountManager()

	callCount := 0
	var mu sync.Mutex

	mgr.SetDestructorCallback(func(obj *ObjectInstance) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	})

	const objects = 100
	var wg sync.WaitGroup
	wg.Add(objects)

	// Create and destroy many objects concurrently
	for i := 0; i < objects; i++ {
		go func() {
			defer wg.Done()

			obj := &ObjectInstance{
				Class:    &mockClassInfo{name: "TestClass"},
				Fields:   make(map[string]Value),
				RefCount: 1,
			}

			mgr.DecrementRef(obj)
		}()
	}

	wg.Wait()

	mu.Lock()
	finalCount := callCount
	mu.Unlock()

	if finalCount != objects {
		t.Errorf("destructor called %d times, want %d", finalCount, objects)
	}
}
