package interp

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestEdge_EmptyInterface tests empty interface (no methods)
// Empty interface (no methods) - declaration, implementation, casting
func TestEdge_EmptyInterface(t *testing.T) {
	t.Run("Declaration", func(t *testing.T) {
		// Create an empty interface
		iface := NewInterfaceInfo("IEmpty")

		if iface.Name != "IEmpty" {
			t.Errorf("Interface name should be IEmpty, got %s", iface.Name)
		}

		if len(iface.Methods) != 0 {
			t.Errorf("Empty interface should have 0 methods, got %d", len(iface.Methods))
		}

		// AllMethods should return empty map
		allMethods := iface.AllMethods()
		if len(allMethods) != 0 {
			t.Errorf("AllMethods should return 0 methods for empty interface, got %d", len(allMethods))
		}
	})

	t.Run("Implementation", func(t *testing.T) {
		// Empty interface should be implementable by any class
		iface := NewInterfaceInfo("IEmpty")

		// Class with no methods
		class1 := NewClassInfo("TEmptyClass")
		if !classImplementsInterface(class1, iface) {
			t.Error("Empty class should implement empty interface")
		}

		// Class with methods
		class2 := NewClassInfo("TClassWithMethods")
		class2.Methods["DoSomething"] = &ast.FunctionDecl{
			Name: &ast.Identifier{Value: "DoSomething"},
		}
		if !classImplementsInterface(class2, iface) {
			t.Error("Class with methods should also implement empty interface")
		}
	})

	t.Run("Casting", func(t *testing.T) {
		iface := NewInterfaceInfo("IEmpty")
		class := NewClassInfo("TAnyClass")
		obj := NewObjectInstance(class)

		// Should be able to cast any object to empty interface
		ifaceInstance := NewInterfaceInstance(iface, obj)

		if ifaceInstance.Interface != iface {
			t.Error("Interface instance should reference empty interface")
		}

		if ifaceInstance.Object != obj {
			t.Error("Interface instance should wrap the object")
		}
	})
}

// TestEdge_InterfaceWithManyMethods tests interface with 10+ methods
// Interface with many methods - verify all accessible
func TestEdge_InterfaceWithManyMethods(t *testing.T) {
	// Create interface with 15 methods
	iface := NewInterfaceInfo("ILargeInterface")

	methodNames := []string{
		"Method01", "Method02", "Method03", "Method04", "Method05",
		"Method06", "Method07", "Method08", "Method09", "Method10",
		"Method11", "Method12", "Method13", "Method14", "Method15",
	}

	for _, name := range methodNames {
		// Store methods with lowercase keys for case-insensitive lookup
		iface.Methods[strings.ToLower(name)] = &ast.FunctionDecl{
			Name: &ast.Identifier{Value: name},
		}
	}

	// Verify all 15 methods are registered
	if len(iface.Methods) != 15 {
		t.Errorf("Interface should have 15 methods, got %d", len(iface.Methods))
	}

	// Verify each method is accessible
	for _, name := range methodNames {
		if !iface.HasMethod(name) {
			t.Errorf("Interface should have method %s", name)
		}

		method := iface.GetMethod(name)
		if method == nil {
			t.Errorf("GetMethod should return method %s", name)
		}

		if method.Name.Value != name {
			t.Errorf("Method name should be %s, got %s", name, method.Name.Value)
		}
	}

	// Verify AllMethods returns all 15
	allMethods := iface.AllMethods()
	if len(allMethods) != 15 {
		t.Errorf("AllMethods should return 15 methods, got %d", len(allMethods))
	}

	// Create class that implements all 15 methods
	class := NewClassInfo("TLargeClass")
	for _, name := range methodNames {
		class.Methods[strings.ToLower(name)] = &ast.FunctionDecl{
			Name: &ast.Identifier{Value: name},
		}
	}

	// Verify class implements interface
	if !classImplementsInterface(class, iface) {
		t.Error("Class with all 15 methods should implement interface")
	}

	// Verify class missing one method doesn't implement interface
	incompleteClass := NewClassInfo("TIncompleteClass")
	for i, name := range methodNames {
		if i < 14 { // Only add 14 methods (missing one)
			incompleteClass.Methods[strings.ToLower(name)] = &ast.FunctionDecl{
				Name: &ast.Identifier{Value: name},
			}
		}
	}

	if classImplementsInterface(incompleteClass, iface) {
		t.Error("Class missing method 15 should NOT implement interface")
	}
}

// TestEdge_DeepInterfaceInheritanceChains tests 5+ level deep inheritance
func TestEdge_DeepInterfaceInheritanceChains(t *testing.T) {
	// Create 7-level deep inheritance chain
	level0 := NewInterfaceInfo("ILevel0")
	level0.Methods["method0"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method0"}}

	level1 := NewInterfaceInfo("ILevel1")
	level1.Parent = level0
	level1.Methods["method1"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method1"}}

	level2 := NewInterfaceInfo("ILevel2")
	level2.Parent = level1
	level2.Methods["method2"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method2"}}

	level3 := NewInterfaceInfo("ILevel3")
	level3.Parent = level2
	level3.Methods["method3"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method3"}}

	level4 := NewInterfaceInfo("ILevel4")
	level4.Parent = level3
	level4.Methods["method4"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method4"}}

	level5 := NewInterfaceInfo("ILevel5")
	level5.Parent = level4
	level5.Methods["method5"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method5"}}

	level6 := NewInterfaceInfo("ILevel6")
	level6.Parent = level5
	level6.Methods["method6"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method6"}}

	// Verify deepest level has access to all inherited methods
	for i := 0; i <= 6; i++ {
		methodName := "Method" + string(rune('0'+i))
		if !level6.HasMethod(methodName) {
			t.Errorf("Level 6 should have inherited method %s", methodName)
		}
	}

	// Verify AllMethods returns all 7 methods
	allMethods := level6.AllMethods()
	if len(allMethods) != 7 {
		t.Errorf("Level 6 should have 7 total methods, got %d", len(allMethods))
	}

	// Verify each level has correct number of methods
	testCases := []struct {
		iface         *InterfaceInfo
		expectedCount int
	}{
		{level0, 1},
		{level1, 2},
		{level2, 3},
		{level3, 4},
		{level4, 5},
		{level5, 6},
		{level6, 7},
	}

	for _, tc := range testCases {
		count := len(tc.iface.AllMethods())
		if count != tc.expectedCount {
			t.Errorf("%s should have %d methods, got %d", tc.iface.Name, tc.expectedCount, count)
		}
	}

	// Verify parent chain is correct
	current := level6
	expectedLevel := 5
	for current.Parent != nil {
		expectedName := "ILevel" + string(rune('0'+expectedLevel))
		if current.Parent.Name != expectedName {
			t.Errorf("Expected parent %s, got %s", expectedName, current.Parent.Name)
		}
		current = current.Parent
		expectedLevel--
	}

	if expectedLevel != -1 {
		t.Errorf("Expected to traverse 7 levels, stopped at level %d", expectedLevel+1)
	}
}

// TestEdge_ConflictingInterfaces tests class implementing interfaces with same method names
func TestEdge_ConflictingInterfaces(t *testing.T) {
	// Create two interfaces with same method name
	iface1 := NewInterfaceInfo("IInterface1")
	iface1.Methods["conflictingmethod"] = &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "ConflictingMethod"},
		ReturnType: &ast.TypeAnnotation{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
			Name:  "Integer",
		},
	}

	iface2 := NewInterfaceInfo("IInterface2")
	iface2.Methods["conflictingmethod"] = &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "ConflictingMethod"},
		ReturnType: &ast.TypeAnnotation{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
			Name:  "String",
		},
	}

	// Create class that implements both interfaces
	// (In DWScript, the single implementation satisfies both)
	class := NewClassInfo("TDualImplementor")
	class.Methods["conflictingmethod"] = &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "ConflictingMethod"},
		ReturnType: &ast.TypeAnnotation{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
			Name:  "Integer",
		},
	}

	// Class implements both interfaces (name match is sufficient)
	if !classImplementsInterface(class, iface1) {
		t.Error("Class should implement IInterface1")
	}

	if !classImplementsInterface(class, iface2) {
		t.Error("Class should implement IInterface2 (method name match)")
	}

	// Create interface instances
	obj := NewObjectInstance(class)
	ifaceInstance1 := NewInterfaceInstance(iface1, obj)
	ifaceInstance2 := NewInterfaceInstance(iface2, obj)

	// Both should wrap same object
	if ifaceInstance1.Object != obj || ifaceInstance2.Object != obj {
		t.Error("Both interfaces should wrap the same object")
	}

	// Different interface types
	if ifaceInstance1.Interface == ifaceInstance2.Interface {
		t.Error("Interface instances should have different interface types")
	}
}

// TestEdge_InterfaceVariablesHoldingNil tests interface variables with nil values
func TestEdge_InterfaceVariablesHoldingNil(t *testing.T) {
	t.Run("NilAssignment", func(t *testing.T) {
		iface := NewInterfaceInfo("ITest")

		// Create interface instance with nil object
		var nilObj *ObjectInstance
		ifaceInstance := NewInterfaceInstance(iface, nilObj)

		// Verify nil is stored
		if ifaceInstance.Object != nil {
			t.Error("Interface instance should hold nil object")
		}

		// Type should still be INTERFACE
		if ifaceInstance.Type() != "INTERFACE" {
			t.Errorf("Type should be INTERFACE, got %s", ifaceInstance.Type())
		}
	})

	t.Run("NilCheck", func(t *testing.T) {
		iface := NewInterfaceInfo("ITest")

		// Nil interface
		var nilObj *ObjectInstance
		nilInstance := NewInterfaceInstance(iface, nilObj)

		// Non-nil interface
		class := NewClassInfo("TTest")
		obj := NewObjectInstance(class)
		nonNilInstance := NewInterfaceInstance(iface, obj)

		// Should be able to distinguish
		if nilInstance.Object != nil {
			t.Error("Nil instance should have nil object")
		}

		if nonNilInstance.Object == nil {
			t.Error("Non-nil instance should have non-nil object")
		}
	})

	t.Run("NilExtraction", func(t *testing.T) {
		iface := NewInterfaceInfo("ITest")
		var nilObj *ObjectInstance
		ifaceInstance := NewInterfaceInstance(iface, nilObj)

		// Extracting underlying object should return nil
		extracted := ifaceInstance.GetUnderlyingObject()
		if extracted != nil {
			t.Error("Extracting from nil interface should return nil")
		}
	})

	t.Run("NilInEnvironment", func(t *testing.T) {
		interp := New(nil)
		iface := NewInterfaceInfo("ITest")

		// Store nil interface in environment
		var nilObj *ObjectInstance
		nilInstance := NewInterfaceInstance(iface, nilObj)
		interp.env.Define("nilInterface", nilInstance)

		// Retrieve and verify
		val, exists := interp.env.Get("nilInterface")
		if !exists {
			t.Fatal("Nil interface should exist in environment")
		}

		retrieved := val.(*InterfaceInstance)
		if retrieved.Object != nil {
			t.Error("Retrieved nil interface should have nil object")
		}
	})

	t.Run("NilComparison", func(t *testing.T) {
		iface := NewInterfaceInfo("ITest")

		// Two nil interfaces
		var nilObj1 *ObjectInstance
		var nilObj2 *ObjectInstance
		nilInstance1 := NewInterfaceInstance(iface, nilObj1)
		nilInstance2 := NewInterfaceInstance(iface, nilObj2)

		// Both should have nil objects
		if nilInstance1.Object != nil || nilInstance2.Object != nil {
			t.Error("Both nil instances should have nil objects")
		}

		// They are different interface instances but both hold nil
		if nilInstance1 == nilInstance2 {
			t.Error("Different instances (even both nil) should be different objects")
		}
	})
}

// TestEdge_InterfaceInstanceImplementsInterface tests ImplementsInterface method
func TestEdge_InterfaceInstanceImplementsInterface(t *testing.T) {
	// Create two interfaces
	iface1 := NewInterfaceInfo("IInterface1")
	iface1.Methods["method1"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method1"}}

	iface2 := NewInterfaceInfo("IInterface2")
	iface2.Methods["method2"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method2"}}

	// Create class that implements iface1 but not iface2
	class := NewClassInfo("TTest")
	class.Methods["method1"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method1"}}

	obj := NewObjectInstance(class)
	ifaceInstance := NewInterfaceInstance(iface1, obj)

	// Should return true for iface1 (class implements it)
	if !ifaceInstance.ImplementsInterface(iface1) {
		t.Error("InterfaceInstance should report it implements iface1")
	}

	// Should return false for iface2 (class doesn't implement it)
	if ifaceInstance.ImplementsInterface(iface2) {
		t.Error("InterfaceInstance should NOT report it implements iface2")
	}
}

// TestEdge_InterfaceCompatibilityEdgeCases tests edge cases in interface compatibility
func TestEdge_InterfaceCompatibilityEdgeCases(t *testing.T) {
	t.Run("SelfCompatibility", func(t *testing.T) {
		// Interface should be compatible with itself
		iface := NewInterfaceInfo("ITest")
		iface.Methods["method1"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method1"}}

		if !interfaceIsCompatible(iface, iface) {
			t.Error("Interface should be compatible with itself")
		}
	})

	t.Run("EmptyToEmpty", func(t *testing.T) {
		// Two empty interfaces should be compatible
		empty1 := NewInterfaceInfo("IEmpty1")
		empty2 := NewInterfaceInfo("IEmpty2")

		if !interfaceIsCompatible(empty1, empty2) {
			t.Error("Empty interfaces should be compatible with each other")
		}

		if !interfaceIsCompatible(empty2, empty1) {
			t.Error("Empty interfaces should be compatible (symmetric)")
		}
	})

	t.Run("SubsetSuperset", func(t *testing.T) {
		// Interface with more methods should be compatible with subset
		superset := NewInterfaceInfo("ISuperset")
		superset.Methods["method1"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method1"}}
		superset.Methods["method2"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method2"}}
		superset.Methods["method3"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method3"}}

		subset := NewInterfaceInfo("ISubset")
		subset.Methods["method1"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method1"}}
		subset.Methods["method2"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "Method2"}}

		// Superset is compatible with subset (can satisfy subset's requirements)
		if !interfaceIsCompatible(superset, subset) {
			t.Error("Interface with more methods should be compatible with subset")
		}

		// But subset is NOT compatible with superset (missing methods)
		if interfaceIsCompatible(subset, superset) {
			t.Error("Interface with fewer methods should NOT be compatible with superset")
		}
	})

	t.Run("UnrelatedInterfaces", func(t *testing.T) {
		// Completely unrelated interfaces
		iface1 := NewInterfaceInfo("IUnrelated1")
		iface1.Methods["methoda"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "MethodA"}}

		iface2 := NewInterfaceInfo("IUnrelated2")
		iface2.Methods["methodb"] = &ast.FunctionDecl{Name: &ast.Identifier{Value: "MethodB"}}

		// Should not be compatible
		if interfaceIsCompatible(iface1, iface2) {
			t.Error("Unrelated interfaces should not be compatible")
		}

		if interfaceIsCompatible(iface2, iface1) {
			t.Error("Unrelated interfaces should not be compatible (symmetric)")
		}
	})
}
