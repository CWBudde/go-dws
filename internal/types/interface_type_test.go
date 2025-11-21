package types

import (
	"testing"
)

// ============================================================================
// InterfaceType Tests (Stage 7.3)
// ============================================================================

func TestNewInterfaceType(t *testing.T) {
	iface := NewInterfaceType("IComparable")

	if iface.Name != "IComparable" {
		t.Errorf("Expected interface name 'IComparable', got '%s'", iface.Name)
	}

	if iface.Methods == nil {
		t.Error("Expected Methods map to be initialized")
	}
}

func TestInterfaceTypeString(t *testing.T) {
	iface := NewInterfaceType("IComparable")
	if iface.String() != "IComparable" {
		t.Errorf("Expected 'IComparable', got '%s'", iface.String())
	}
}

func TestInterfaceTypeKind(t *testing.T) {
	iface := NewInterfaceType("IComparable")
	if iface.TypeKind() != "INTERFACE" {
		t.Errorf("Expected TypeKind 'INTERFACE', got '%s'", iface.TypeKind())
	}
}

func TestInterfaceTypeEquals(t *testing.T) {
	iface1 := NewInterfaceType("IComparable")
	iface2 := NewInterfaceType("IComparable")
	iface3 := NewInterfaceType("IEnumerable")

	tests := []struct {
		iface1   Type
		iface2   Type
		name     string
		expected bool
	}{
		{
			name:     "same interface name",
			iface1:   iface1,
			iface2:   iface2,
			expected: true,
		},
		{
			name:     "different interface names",
			iface1:   iface1,
			iface2:   iface3,
			expected: false,
		},
		{
			name:     "interface vs non-interface",
			iface1:   iface1,
			iface2:   STRING,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.iface1.Equals(tt.iface2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestInterfaceTypeMethods(t *testing.T) {
	iface := NewInterfaceType("IComparable")
	iface.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)
	iface.Methods["Equals"] = NewFunctionType([]Type{INTEGER}, BOOLEAN)

	tests := []struct {
		name       string
		methodName string
		hasMethod  bool
	}{
		{
			name:       "existing method",
			methodName: "CompareTo",
			hasMethod:  true,
		},
		{
			name:       "non-existent method",
			methodName: "DoSomething",
			hasMethod:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test HasMethod
			hasMethod := iface.HasMethod(tt.methodName)
			if hasMethod != tt.hasMethod {
				t.Errorf("HasMethod: expected %v, got %v", tt.hasMethod, hasMethod)
			}

			// Test GetMethod
			_, found := iface.GetMethod(tt.methodName)
			if found != tt.hasMethod {
				t.Errorf("GetMethod found: expected %v, got %v", tt.hasMethod, found)
			}
		})
	}
}

// ============================================================================
// Interface Inheritance Tests
// ============================================================================

// Test Equals() with hierarchy support for interfaces
func TestInterfaceEqualsWithHierarchy(t *testing.T) {
	// Create interface hierarchy
	iBase := NewInterfaceType("IBase")
	iDerived := &InterfaceType{
		Name:    "IDerived",
		Parent:  iBase,
		Methods: make(map[string]*FunctionType),
	}

	tests := []struct {
		iface1   Type
		iface2   Type
		name     string
		note     string
		expected bool
	}{
		{
			name:     "exact same interface",
			iface1:   iBase,
			iface2:   iBase,
			expected: true,
			note:     "same interface instance should equal itself",
		},
		{
			name:     "same name different instance",
			iface1:   NewInterfaceType("ITest"),
			iface2:   NewInterfaceType("ITest"),
			expected: true,
			note:     "interfaces with same name should be equal (nominal typing)",
		},
		{
			name:     "different interface names",
			iface1:   iBase,
			iface2:   iDerived,
			expected: false,
			note:     "derived interface is NOT equal to base (exact match required)",
		},
		{
			name:     "interface vs non-interface",
			iface1:   iBase,
			iface2:   INTEGER,
			expected: false,
			note:     "interface should not equal non-interface type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.iface1.Equals(tt.iface2)
			if result != tt.expected {
				t.Errorf("%s: Expected %v, got %v", tt.note, tt.expected, result)
			}
		})
	}
}

// Test IINTERFACE base interface constant
func TestIINTERFACEConstant(t *testing.T) {
	t.Run("IINTERFACE exists", func(t *testing.T) {
		if IINTERFACE == nil {
			t.Error("IINTERFACE should be defined")
		}
		if IINTERFACE.Name != "IInterface" {
			t.Errorf("Expected IINTERFACE name to be 'IInterface', got '%s'", IINTERFACE.Name)
		}
		if IINTERFACE.Parent != nil {
			t.Error("IINTERFACE should have no parent (root interface)")
		}
	})

	t.Run("all interfaces can inherit from IINTERFACE", func(t *testing.T) {
		iCustom := &InterfaceType{
			Name:    "ICustom",
			Parent:  IINTERFACE,
			Methods: make(map[string]*FunctionType),
		}

		if !IsSubinterfaceOf(iCustom, IINTERFACE) {
			t.Error("ICustom should be subinterface of IINTERFACE")
		}
	})
}

// Test interface with Parent, IsExternal, ExternalName
func TestInterfaceTypeWithInheritance(t *testing.T) {
	// Create base interface
	iBase := NewInterfaceType("IBase")
	iBase.Methods["BaseMethod"] = NewProcedureType([]Type{})

	// Create derived interface with parent
	iDerived := &InterfaceType{
		Name:    "IDerived",
		Parent:  iBase,
		Methods: make(map[string]*FunctionType),
	}
	iDerived.Methods["DerivedMethod"] = NewProcedureType([]Type{})

	t.Run("interface has parent", func(t *testing.T) {
		if iDerived.Parent == nil {
			t.Error("IDerived should have IBase as parent")
		}
		if iDerived.Parent.Name != "IBase" {
			t.Errorf("Expected parent name 'IBase', got '%s'", iDerived.Parent.Name)
		}
	})

	t.Run("base interface has no parent", func(t *testing.T) {
		if iBase.Parent != nil {
			t.Error("IBase should have no parent")
		}
	})
}

func TestInterfaceTypeExternal(t *testing.T) {
	// Create external interface (for FFI)
	iExternal := &InterfaceType{
		Name:         "IExternal",
		Methods:      make(map[string]*FunctionType),
		IsExternal:   true,
		ExternalName: "IDispatch",
	}

	t.Run("external interface flag", func(t *testing.T) {
		if !iExternal.IsExternal {
			t.Error("Interface should be marked as external")
		}
		if iExternal.ExternalName != "IDispatch" {
			t.Errorf("Expected external name 'IDispatch', got '%s'", iExternal.ExternalName)
		}
	})

	// Create regular (non-external) interface
	iRegular := NewInterfaceType("IRegular")

	t.Run("regular interface not external", func(t *testing.T) {
		if iRegular.IsExternal {
			t.Error("Regular interface should not be marked as external")
		}
		if iRegular.ExternalName != "" {
			t.Error("Regular interface should have empty external name")
		}
	})
}

// Test interface inheritance checking
func TestIsSubinterfaceOf(t *testing.T) {
	// Create interface hierarchy: IBase -> IMiddle -> IDerived
	iBase := NewInterfaceType("IBase")
	iBase.Methods["BaseMethod"] = NewProcedureType([]Type{})

	iMiddle := &InterfaceType{
		Name:    "IMiddle",
		Parent:  iBase,
		Methods: make(map[string]*FunctionType),
	}
	iMiddle.Methods["MiddleMethod"] = NewProcedureType([]Type{})

	iDerived := &InterfaceType{
		Name:    "IDerived",
		Parent:  iMiddle,
		Methods: make(map[string]*FunctionType),
	}
	iDerived.Methods["DerivedMethod"] = NewProcedureType([]Type{})

	iUnrelated := NewInterfaceType("IUnrelated")

	tests := []struct {
		child    *InterfaceType
		parent   *InterfaceType
		name     string
		expected bool
	}{
		{
			name:     "direct parent",
			child:    iDerived,
			parent:   iMiddle,
			expected: true,
		},
		{
			name:     "grandparent",
			child:    iDerived,
			parent:   iBase,
			expected: true,
		},
		{
			name:     "immediate parent",
			child:    iMiddle,
			parent:   iBase,
			expected: true,
		},
		{
			name:     "same interface",
			child:    iMiddle,
			parent:   iMiddle,
			expected: true,
		},
		{
			name:     "unrelated interfaces",
			child:    iDerived,
			parent:   iUnrelated,
			expected: false,
		},
		{
			name:     "reverse hierarchy",
			child:    iBase,
			parent:   iMiddle,
			expected: false,
		},
		{
			name:     "nil child",
			child:    nil,
			parent:   iBase,
			expected: false,
		},
		{
			name:     "nil parent",
			child:    iDerived,
			parent:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSubinterfaceOf(tt.child, tt.parent)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test circular interface inheritance detection
func TestCircularInterfaceInheritance(t *testing.T) {
	// This test will verify that we detect circular inheritance
	// For now, we'll test the concept - actual implementation may vary
	iA := NewInterfaceType("IA")
	iB := &InterfaceType{
		Name:    "IB",
		Parent:  iA,
		Methods: make(map[string]*FunctionType),
	}
	// Attempting to make iA inherit from iB would create a cycle
	// This should be detected during semantic analysis

	t.Run("no circular inheritance", func(t *testing.T) {
		// Verify simple inheritance works
		if !IsSubinterfaceOf(iB, iA) {
			t.Error("IB should be subinterface of IA")
		}
	})
}

// Test interface method inheritance
func TestInterfaceMethodInheritance(t *testing.T) {
	// Create base interface with methods
	iBase := NewInterfaceType("IBase")
	iBase.Methods["BaseMethod"] = NewProcedureType([]Type{})
	iBase.Methods["GetValue"] = NewFunctionType([]Type{}, INTEGER)

	// Create derived interface
	iDerived := &InterfaceType{
		Name:    "IDerived",
		Parent:  iBase,
		Methods: make(map[string]*FunctionType),
	}
	iDerived.Methods["DerivedMethod"] = NewProcedureType([]Type{})

	t.Run("interface has all inherited methods", func(t *testing.T) {
		// Interface should have access to parent methods
		allMethods := GetAllInterfaceMethods(iDerived)

		if len(allMethods) != 3 {
			t.Errorf("Expected 3 methods (1 own + 2 inherited), got %d", len(allMethods))
		}

		if _, ok := allMethods["BaseMethod"]; !ok {
			t.Error("Should have inherited BaseMethod from parent")
		}
		if _, ok := allMethods["GetValue"]; !ok {
			t.Error("Should have inherited GetValue from parent")
		}
		if _, ok := allMethods["DerivedMethod"]; !ok {
			t.Error("Should have own DerivedMethod")
		}
	})
}

// Test interface-to-interface assignment
func TestInterfaceToInterfaceAssignment(t *testing.T) {
	// Create interface hierarchy
	iBase := NewInterfaceType("IBase")
	iBase.Methods["BaseMethod"] = NewProcedureType([]Type{})

	iDerived := &InterfaceType{
		Name:    "IDerived",
		Parent:  iBase,
		Methods: make(map[string]*FunctionType),
	}
	iDerived.Methods["DerivedMethod"] = NewProcedureType([]Type{})

	iUnrelated := NewInterfaceType("IUnrelated")
	iUnrelated.Methods["UnrelatedMethod"] = NewProcedureType([]Type{})

	tests := []struct {
		target   Type
		source   Type
		name     string
		expected bool
	}{
		{
			name:     "derived to base interface (covariant)",
			target:   iBase,
			source:   iDerived,
			expected: true,
		},
		{
			name:     "same interface",
			target:   iBase,
			source:   iBase,
			expected: true,
		},
		{
			name:     "base to derived (contravariant - not allowed)",
			target:   iDerived,
			source:   iBase,
			expected: false,
		},
		{
			name:     "unrelated interfaces",
			target:   iBase,
			source:   iUnrelated,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAssignableFrom(tt.target, tt.source)
			if result != tt.expected {
				t.Errorf("IsAssignableFrom(%v, %v) = %v, want %v",
					tt.target, tt.source, result, tt.expected)
			}
		})
	}
}

// Test class with multiple interface implementation
func TestClassWithMultipleInterfaces(t *testing.T) {
	// Create multiple interfaces
	iReadable := NewInterfaceType("IReadable")
	iReadable.Methods["Read"] = NewFunctionType([]Type{}, STRING)

	iWritable := NewInterfaceType("IWritable")
	iWritable.Methods["Write"] = NewProcedureType([]Type{STRING})

	// Create class implementing both
	tFile := NewClassType("TFile", nil)
	tFile.AddMethodOverload("Read", &MethodInfo{
		Signature: NewFunctionType([]Type{}, STRING),
	})
	tFile.AddMethodOverload("Write", &MethodInfo{
		Signature: NewProcedureType([]Type{STRING}),
	})
	tFile.Interfaces = []*InterfaceType{iReadable, iWritable}

	t.Run("class tracks implemented interfaces", func(t *testing.T) {
		if len(tFile.Interfaces) != 2 {
			t.Errorf("Expected 2 interfaces, got %d", len(tFile.Interfaces))
		}
	})

	t.Run("class implements all tracked interfaces", func(t *testing.T) {
		for _, iface := range tFile.Interfaces {
			if !ImplementsInterface(tFile, iface) {
				t.Errorf("Class should implement interface %s", iface.Name)
			}
		}
	})
}

// Test that ClassType has Interfaces field
func TestClassTypeInterfacesField(t *testing.T) {
	iComparable := NewInterfaceType("IComparable")
	iComparable.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)

	tPerson := NewClassType("TPerson", nil)
	tPerson.AddMethodOverload("CompareTo", &MethodInfo{
		Signature: NewFunctionType([]Type{INTEGER}, INTEGER),
	})
	tPerson.Interfaces = []*InterfaceType{iComparable}

	t.Run("class has interfaces field", func(t *testing.T) {
		if tPerson.Interfaces == nil {
			t.Error("ClassType should have Interfaces field initialized")
		}
		if len(tPerson.Interfaces) != 1 {
			t.Errorf("Expected 1 interface, got %d", len(tPerson.Interfaces))
		}
		if tPerson.Interfaces[0].Name != "IComparable" {
			t.Errorf("Expected interface IComparable, got %s", tPerson.Interfaces[0].Name)
		}
	})
}
