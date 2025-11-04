package types

import (
	"testing"
)

// ============================================================================
// ClassType Tests (Stage 7.1-7.2)
// ============================================================================

func TestNewClassType(t *testing.T) {
	class := NewClassType("TPoint", nil)

	if class.Name != "TPoint" {
		t.Errorf("Expected class name 'TPoint', got '%s'", class.Name)
	}

	if class.Parent != nil {
		t.Errorf("Expected nil parent, got %v", class.Parent)
	}

	if class.Fields == nil {
		t.Error("Expected Fields map to be initialized")
	}

	if class.Methods == nil {
		t.Error("Expected Methods map to be initialized")
	}
}

func TestClassTypeString(t *testing.T) {
	tests := []struct {
		name     string
		class    *ClassType
		expected string
	}{
		{
			name:     "root class",
			class:    NewClassType("TObject", nil),
			expected: "TObject",
		},
		{
			name:     "derived class",
			class:    NewClassType("TPerson", NewClassType("TObject", nil)),
			expected: "TPerson(TObject)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.class.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestClassTypeKind(t *testing.T) {
	class := NewClassType("TPoint", nil)
	if class.TypeKind() != "CLASS" {
		t.Errorf("Expected TypeKind 'CLASS', got '%s'", class.TypeKind())
	}
}

func TestClassTypeEquals(t *testing.T) {
	class1 := NewClassType("TPoint", nil)
	class2 := NewClassType("TPoint", nil)
	class3 := NewClassType("TPerson", nil)

	tests := []struct {
		class1   Type
		class2   Type
		name     string
		expected bool
	}{
		{
			name:     "same class name",
			class1:   class1,
			class2:   class2,
			expected: true,
		},
		{
			name:     "different class names",
			class1:   class1,
			class2:   class3,
			expected: false,
		},
		{
			name:     "class vs non-class",
			class1:   class1,
			class2:   INTEGER,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.class1.Equals(tt.class2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestClassTypeFields(t *testing.T) {
	parent := NewClassType("TObject", nil)
	parent.Fields["ID"] = INTEGER

	child := NewClassType("TPerson", parent)
	child.Fields["Name"] = STRING
	child.Fields["Age"] = INTEGER

	tests := []struct {
		fieldType Type
		class     *ClassType
		name      string
		fieldName string
		hasField  bool
	}{
		{
			name:      "own field",
			class:     child,
			fieldName: "Name",
			hasField:  true,
			fieldType: STRING,
		},
		{
			name:      "inherited field",
			class:     child,
			fieldName: "ID",
			hasField:  true,
			fieldType: INTEGER,
		},
		{
			name:      "non-existent field",
			class:     child,
			fieldName: "Email",
			hasField:  false,
			fieldType: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test HasField
			hasField := tt.class.HasField(tt.fieldName)
			if hasField != tt.hasField {
				t.Errorf("HasField: expected %v, got %v", tt.hasField, hasField)
			}

			// Test GetField
			fieldType, found := tt.class.GetField(tt.fieldName)
			if found != tt.hasField {
				t.Errorf("GetField found: expected %v, got %v", tt.hasField, found)
			}

			if tt.hasField && !fieldType.Equals(tt.fieldType) {
				t.Errorf("GetField type: expected %v, got %v", tt.fieldType, fieldType)
			}
		})
	}
}

func TestClassTypeMethods(t *testing.T) {
	parent := NewClassType("TObject", nil)
	parent.Methods["ToString"] = NewFunctionType([]Type{}, STRING)

	child := NewClassType("TPerson", parent)
	child.Methods["GetAge"] = NewFunctionType([]Type{}, INTEGER)
	child.Methods["SetName"] = NewProcedureType([]Type{STRING})

	tests := []struct {
		name       string
		class      *ClassType
		methodName string
		hasMethod  bool
		isFunction bool
	}{
		{
			name:       "own method (function)",
			class:      child,
			methodName: "GetAge",
			hasMethod:  true,
			isFunction: true,
		},
		{
			name:       "own method (procedure)",
			class:      child,
			methodName: "SetName",
			hasMethod:  true,
			isFunction: false,
		},
		{
			name:       "inherited method",
			class:      child,
			methodName: "ToString",
			hasMethod:  true,
			isFunction: true,
		},
		{
			name:       "non-existent method",
			class:      child,
			methodName: "DoSomething",
			hasMethod:  false,
			isFunction: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test HasMethod
			hasMethod := tt.class.HasMethod(tt.methodName)
			if hasMethod != tt.hasMethod {
				t.Errorf("HasMethod: expected %v, got %v", tt.hasMethod, hasMethod)
			}

			// Test GetMethod
			methodType, found := tt.class.GetMethod(tt.methodName)
			if found != tt.hasMethod {
				t.Errorf("GetMethod found: expected %v, got %v", tt.hasMethod, found)
			}

			if tt.hasMethod {
				if methodType.IsFunction() != tt.isFunction {
					t.Errorf("IsFunction: expected %v, got %v", tt.isFunction, methodType.IsFunction())
				}
			}
		})
	}
}

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
// Type Compatibility Tests (Stage 7.4)
// ============================================================================

func TestIsSubclassOf(t *testing.T) {
	// Create class hierarchy: TObject -> TPerson -> TEmployee
	tObject := NewClassType("TObject", nil)
	tPerson := NewClassType("TPerson", tObject)
	tEmployee := NewClassType("TEmployee", tPerson)
	tUnrelated := NewClassType("TPoint", nil)

	tests := []struct {
		child    *ClassType
		parent   *ClassType
		name     string
		expected bool
	}{
		{
			name:     "direct parent",
			child:    tPerson,
			parent:   tObject,
			expected: true,
		},
		{
			name:     "grandparent",
			child:    tEmployee,
			parent:   tObject,
			expected: true,
		},
		{
			name:     "immediate parent",
			child:    tEmployee,
			parent:   tPerson,
			expected: true,
		},
		{
			name:     "same class",
			child:    tPerson,
			parent:   tPerson,
			expected: true,
		},
		{
			name:     "unrelated classes",
			child:    tPerson,
			parent:   tUnrelated,
			expected: false,
		},
		{
			name:     "reverse hierarchy",
			child:    tObject,
			parent:   tPerson,
			expected: false,
		},
		{
			name:     "nil child",
			child:    nil,
			parent:   tObject,
			expected: false,
		},
		{
			name:     "nil parent",
			child:    tPerson,
			parent:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSubclassOf(tt.child, tt.parent)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsAssignableFrom(t *testing.T) {
	// Create class hierarchy
	tObject := NewClassType("TObject", nil)
	tPerson := NewClassType("TPerson", tObject)
	tEmployee := NewClassType("TEmployee", tPerson)

	// Create interface
	iComparable := NewInterfaceType("IComparable")
	iComparable.Methods["CompareTo"] = NewFunctionType([]Type{}, INTEGER)

	// Make TPerson implement IComparable
	tPerson.Methods["CompareTo"] = NewFunctionType([]Type{}, INTEGER)

	tests := []struct {
		target   Type
		source   Type
		name     string
		expected bool
	}{
		{
			name:     "exact type match",
			target:   INTEGER,
			source:   INTEGER,
			expected: true,
		},
		{
			name:     "integer to float coercion",
			target:   FLOAT,
			source:   INTEGER,
			expected: true,
		},
		{
			name:     "float to integer (no coercion)",
			target:   INTEGER,
			source:   FLOAT,
			expected: false,
		},
		{
			name:     "subclass to superclass",
			target:   tObject,
			source:   tPerson,
			expected: true,
		},
		{
			name:     "grandchild to grandparent",
			target:   tObject,
			source:   tEmployee,
			expected: true,
		},
		{
			name:     "superclass to subclass (not allowed)",
			target:   tPerson,
			source:   tObject,
			expected: false,
		},
		{
			name:     "class to implemented interface",
			target:   iComparable,
			source:   tPerson,
			expected: true,
		},
		{
			name:     "class to non-implemented interface",
			target:   iComparable,
			source:   tObject,
			expected: false,
		},
		{
			name:     "incompatible types",
			target:   STRING,
			source:   INTEGER,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAssignableFrom(tt.target, tt.source)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Interface Implementation Tests (Stage 7.5)
// ============================================================================

func TestImplementsInterface(t *testing.T) {
	// Create interface with two methods
	iComparable := NewInterfaceType("IComparable")
	iComparable.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)
	iComparable.Methods["Equals"] = NewFunctionType([]Type{INTEGER}, BOOLEAN)

	// Create interface with one method
	iSimple := NewInterfaceType("ISimple")
	iSimple.Methods["DoSomething"] = NewProcedureType([]Type{})

	// Create class that implements IComparable fully
	tFullImpl := NewClassType("TFullImpl", nil)
	tFullImpl.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)
	tFullImpl.Methods["Equals"] = NewFunctionType([]Type{INTEGER}, BOOLEAN)

	// Create class that partially implements IComparable
	tPartialImpl := NewClassType("TPartialImpl", nil)
	tPartialImpl.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)
	// Missing Equals method

	// Create class with wrong signature
	tWrongSig := NewClassType("TWrongSig", nil)
	tWrongSig.Methods["CompareTo"] = NewFunctionType([]Type{STRING}, INTEGER) // Wrong param type
	tWrongSig.Methods["Equals"] = NewFunctionType([]Type{INTEGER}, BOOLEAN)

	// Create class that implements via inheritance
	tParent := NewClassType("TParent", nil)
	tParent.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)
	tChild := NewClassType("TChild", tParent)
	tChild.Methods["Equals"] = NewFunctionType([]Type{INTEGER}, BOOLEAN)

	tests := []struct {
		class    *ClassType
		iface    *InterfaceType
		name     string
		expected bool
	}{
		{
			name:     "full implementation",
			class:    tFullImpl,
			iface:    iComparable,
			expected: true,
		},
		{
			name:     "partial implementation",
			class:    tPartialImpl,
			iface:    iComparable,
			expected: false,
		},
		{
			name:     "wrong signature",
			class:    tWrongSig,
			iface:    iComparable,
			expected: false,
		},
		{
			name:     "implementation via inheritance",
			class:    tChild,
			iface:    iComparable,
			expected: true,
		},
		{
			name:     "unrelated interface",
			class:    tFullImpl,
			iface:    iSimple,
			expected: false,
		},
		{
			name:     "nil class",
			class:    nil,
			iface:    iComparable,
			expected: false,
		},
		{
			name:     "nil interface",
			class:    tFullImpl,
			iface:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ImplementsInterface(tt.class, tt.iface)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Type Checking Utility Tests
// ============================================================================

func TestIsClassType(t *testing.T) {
	class := NewClassType("TPoint", nil)

	tests := []struct {
		typ      Type
		name     string
		expected bool
	}{
		{
			name:     "class type",
			typ:      class,
			expected: true,
		},
		{
			name:     "integer type",
			typ:      INTEGER,
			expected: false,
		},
		{
			name:     "function type",
			typ:      NewFunctionType([]Type{INTEGER}, STRING),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsClassType(tt.typ)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsInterfaceType(t *testing.T) {
	iface := NewInterfaceType("IComparable")

	tests := []struct {
		typ      Type
		name     string
		expected bool
	}{
		{
			name:     "interface type",
			typ:      iface,
			expected: true,
		},
		{
			name:     "string type",
			typ:      STRING,
			expected: false,
		},
		{
			name:     "class type",
			typ:      NewClassType("TPoint", nil),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInterfaceType(tt.typ)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Interface Inheritance Tests
// ============================================================================

// Task 7.76: Test Equals() with hierarchy support for interfaces
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

// Task 7.75: Test IINTERFACE base interface constant
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

// Task 7.73-7.74: Test interface with Parent, IsExternal, ExternalName
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

// Task 7.77: Test interface inheritance checking
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

// Task 7.77: Test circular interface inheritance detection
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

// Task 7.78: Test interface method inheritance
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

// Task 7.79: Test interface-to-interface assignment
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

// Task 7.80: Test class with multiple interface implementation
func TestClassWithMultipleInterfaces(t *testing.T) {
	// Create multiple interfaces
	iReadable := NewInterfaceType("IReadable")
	iReadable.Methods["Read"] = NewFunctionType([]Type{}, STRING)

	iWritable := NewInterfaceType("IWritable")
	iWritable.Methods["Write"] = NewProcedureType([]Type{STRING})

	// Create class implementing both
	tFile := NewClassType("TFile", nil)
	tFile.Methods["Read"] = NewFunctionType([]Type{}, STRING)
	tFile.Methods["Write"] = NewProcedureType([]Type{STRING})
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

// Task 7.80: Test that ClassType has Interfaces field
func TestClassTypeInterfacesField(t *testing.T) {
	iComparable := NewInterfaceType("IComparable")
	iComparable.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)

	tPerson := NewClassType("TPerson", nil)
	tPerson.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)
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

// ============================================================================
// Complex Hierarchy Tests
// ============================================================================

func TestComplexClassHierarchy(t *testing.T) {
	// Build a more complex hierarchy:
	// TObject
	//   ├─ TStream
	//   │    ├─ TFileStream
	//   │    └─ TMemoryStream
	//   └─ TPersistent
	//        └─ TComponent

	tObject := NewClassType("TObject", nil)
	tObject.Methods["ToString"] = NewFunctionType([]Type{}, STRING)

	tStream := NewClassType("TStream", tObject)
	tStream.Fields["Size"] = INTEGER
	tStream.Methods["Read"] = NewFunctionType([]Type{INTEGER}, INTEGER)

	tFileStream := NewClassType("TFileStream", tStream)
	tFileStream.Fields["FileName"] = STRING

	tMemoryStream := NewClassType("TMemoryStream", tStream)
	tMemoryStream.Fields["Memory"] = INTEGER

	tPersistent := NewClassType("TPersistent", tObject)
	tPersistent.Methods["Assign"] = NewProcedureType([]Type{})

	tComponent := NewClassType("TComponent", tPersistent)
	tComponent.Fields["Name"] = STRING

	// Test field inheritance through multiple levels
	t.Run("field inheritance", func(t *testing.T) {
		if !tFileStream.HasField("Size") {
			t.Error("TFileStream should have Size field from TStream")
		}
		if !tComponent.HasField("Name") {
			t.Error("TComponent should have its own Name field")
		}
	})

	// Test method inheritance through multiple levels
	t.Run("method inheritance", func(t *testing.T) {
		if !tFileStream.HasMethod("ToString") {
			t.Error("TFileStream should have ToString method from TObject")
		}
		if !tComponent.HasMethod("ToString") {
			t.Error("TComponent should have ToString method from TObject")
		}
		if !tComponent.HasMethod("Assign") {
			t.Error("TComponent should have Assign method from TPersistent")
		}
	})

	// Test subclass relationships
	t.Run("subclass relationships", func(t *testing.T) {
		if !IsSubclassOf(tFileStream, tObject) {
			t.Error("TFileStream should be subclass of TObject")
		}
		if !IsSubclassOf(tComponent, tObject) {
			t.Error("TComponent should be subclass of TObject")
		}
		if IsSubclassOf(tFileStream, tComponent) {
			t.Error("TFileStream should not be subclass of TComponent")
		}
	})
}

func TestMultipleInterfaceImplementation(t *testing.T) {
	// Create multiple interfaces
	iReadable := NewInterfaceType("IReadable")
	iReadable.Methods["Read"] = NewFunctionType([]Type{}, STRING)

	iWritable := NewInterfaceType("IWritable")
	iWritable.Methods["Write"] = NewProcedureType([]Type{STRING})

	iCloseable := NewInterfaceType("ICloseable")
	iCloseable.Methods["Close"] = NewProcedureType([]Type{})

	// Create class that implements all three
	tFile := NewClassType("TFile", nil)
	tFile.Methods["Read"] = NewFunctionType([]Type{}, STRING)
	tFile.Methods["Write"] = NewProcedureType([]Type{STRING})
	tFile.Methods["Close"] = NewProcedureType([]Type{})

	// Test each interface
	tests := []struct {
		iface    *InterfaceType
		name     string
		expected bool
	}{
		{
			name:     "implements IReadable",
			iface:    iReadable,
			expected: true,
		},
		{
			name:     "implements IWritable",
			iface:    iWritable,
			expected: true,
		},
		{
			name:     "implements ICloseable",
			iface:    iCloseable,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ImplementsInterface(tFile, tt.iface)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Task 7.137: External Class Tests
// ============================================================================

// TestExternalClassFields tests that ClassType has IsExternal and ExternalName fields
func TestExternalClassFields(t *testing.T) {
	t.Run("regular class is not external", func(t *testing.T) {
		tRegular := NewClassType("TRegular", nil)
		if tRegular.IsExternal {
			t.Error("Regular class should not be marked as external")
		}
		if tRegular.ExternalName != "" {
			t.Errorf("Regular class should have empty ExternalName, got %q", tRegular.ExternalName)
		}
	})

	t.Run("external class without external name", func(t *testing.T) {
		tExternal := NewClassType("TExternal", nil)
		tExternal.IsExternal = true
		if !tExternal.IsExternal {
			t.Error("External class should be marked as external")
		}
		if tExternal.ExternalName != "" {
			t.Errorf("External class without name should have empty ExternalName, got %q", tExternal.ExternalName)
		}
	})

	t.Run("external class with external name", func(t *testing.T) {
		tExternal := NewClassType("TExternal", nil)
		tExternal.IsExternal = true
		tExternal.ExternalName = "MyExternalClass"
		if !tExternal.IsExternal {
			t.Error("External class should be marked as external")
		}
		if tExternal.ExternalName != "MyExternalClass" {
			t.Errorf("Expected ExternalName 'MyExternalClass', got %q", tExternal.ExternalName)
		}
	})

	t.Run("external class with parent", func(t *testing.T) {
		tParent := NewClassType("TParentExternal", nil)
		tParent.IsExternal = true

		tChild := NewClassType("TChildExternal", tParent)
		tChild.IsExternal = true
		tChild.ExternalName = "ChildExternal"

		if !tChild.IsExternal {
			t.Error("Child external class should be marked as external")
		}
		if tChild.ExternalName != "ChildExternal" {
			t.Errorf("Expected ExternalName 'ChildExternal', got %q", tChild.ExternalName)
		}
	})
}

// ============================================================================
// PropertyInfo Tests (Stage 8, Tasks 8.26-8.29)
// ============================================================================

func TestPropertyInfo(t *testing.T) {
	t.Run("create field-backed property", func(t *testing.T) {
		// Property: property Name: String read FName write FName;
		prop := &PropertyInfo{
			Name:      "Name",
			Type:      STRING,
			ReadKind:  PropAccessField,
			ReadSpec:  "FName",
			WriteKind: PropAccessField,
			WriteSpec: "FName",
			IsIndexed: false,
			IsDefault: false,
		}

		if prop.Name != "Name" {
			t.Errorf("Expected Name='Name', got '%s'", prop.Name)
		}
		if !prop.Type.Equals(STRING) {
			t.Errorf("Expected Type=STRING, got %v", prop.Type)
		}
		if prop.ReadKind != PropAccessField {
			t.Errorf("Expected ReadKind=PropAccessField, got %v", prop.ReadKind)
		}
		if prop.ReadSpec != "FName" {
			t.Errorf("Expected ReadSpec='FName', got '%s'", prop.ReadSpec)
		}
		if prop.WriteKind != PropAccessField {
			t.Errorf("Expected WriteKind=PropAccessField, got %v", prop.WriteKind)
		}
		if prop.WriteSpec != "FName" {
			t.Errorf("Expected WriteSpec='FName', got '%s'", prop.WriteSpec)
		}
		if prop.IsIndexed {
			t.Error("Expected IsIndexed=false")
		}
		if prop.IsDefault {
			t.Error("Expected IsDefault=false")
		}
	})

	t.Run("create method-backed property", func(t *testing.T) {
		// Property: property Count: Integer read GetCount write SetCount;
		prop := &PropertyInfo{
			Name:      "Count",
			Type:      INTEGER,
			ReadKind:  PropAccessMethod,
			ReadSpec:  "GetCount",
			WriteKind: PropAccessMethod,
			WriteSpec: "SetCount",
			IsIndexed: false,
			IsDefault: false,
		}

		if prop.ReadKind != PropAccessMethod {
			t.Errorf("Expected ReadKind=PropAccessMethod, got %v", prop.ReadKind)
		}
		if prop.WriteKind != PropAccessMethod {
			t.Errorf("Expected WriteKind=PropAccessMethod, got %v", prop.WriteKind)
		}
	})

	t.Run("create read-only property", func(t *testing.T) {
		// Property: property Size: Integer read FSize;
		prop := &PropertyInfo{
			Name:      "Size",
			Type:      INTEGER,
			ReadKind:  PropAccessField,
			ReadSpec:  "FSize",
			WriteKind: PropAccessNone,
			WriteSpec: "",
			IsIndexed: false,
			IsDefault: false,
		}

		if prop.WriteKind != PropAccessNone {
			t.Errorf("Expected WriteKind=PropAccessNone, got %v", prop.WriteKind)
		}
		if prop.WriteSpec != "" {
			t.Errorf("Expected empty WriteSpec, got '%s'", prop.WriteSpec)
		}
	})

	t.Run("create expression-backed property", func(t *testing.T) {
		// Property: property Double: Integer read (FValue * 2);
		prop := &PropertyInfo{
			Name:      "Double",
			Type:      INTEGER,
			ReadKind:  PropAccessExpression,
			ReadSpec:  "(FValue * 2)",
			WriteKind: PropAccessNone,
			WriteSpec: "",
			IsIndexed: false,
			IsDefault: false,
		}

		if prop.ReadKind != PropAccessExpression {
			t.Errorf("Expected ReadKind=PropAccessExpression, got %v", prop.ReadKind)
		}
	})

	t.Run("create indexed property", func(t *testing.T) {
		// Property: property Items[index: Integer]: String read GetItem write SetItem;
		prop := &PropertyInfo{
			Name:      "Items",
			Type:      STRING,
			ReadKind:  PropAccessMethod,
			ReadSpec:  "GetItem",
			WriteKind: PropAccessMethod,
			WriteSpec: "SetItem",
			IsIndexed: true,
			IsDefault: false,
		}

		if !prop.IsIndexed {
			t.Error("Expected IsIndexed=true")
		}
	})

	t.Run("create default indexed property", func(t *testing.T) {
		// Property: property Items[index: Integer]: String read GetItem write SetItem; default;
		prop := &PropertyInfo{
			Name:      "Items",
			Type:      STRING,
			ReadKind:  PropAccessMethod,
			ReadSpec:  "GetItem",
			WriteKind: PropAccessMethod,
			WriteSpec: "SetItem",
			IsIndexed: true,
			IsDefault: true,
		}

		if !prop.IsDefault {
			t.Error("Expected IsDefault=true")
		}
	})
}

func TestClassTypeProperties(t *testing.T) {
	t.Run("HasProperty and GetProperty", func(t *testing.T) {
		// Create a class with properties
		class := NewClassType("TPerson", nil)
		class.Fields["FName"] = STRING
		class.Fields["FAge"] = INTEGER

		// Add a field-backed property
		class.Properties["Name"] = &PropertyInfo{
			Name:      "Name",
			Type:      STRING,
			ReadKind:  PropAccessField,
			ReadSpec:  "FName",
			WriteKind: PropAccessField,
			WriteSpec: "FName",
		}

		// Add a method-backed property
		class.Properties["Age"] = &PropertyInfo{
			Name:      "Age",
			Type:      INTEGER,
			ReadKind:  PropAccessMethod,
			ReadSpec:  "GetAge",
			WriteKind: PropAccessMethod,
			WriteSpec: "SetAge",
		}

		// Test HasProperty
		if !class.HasProperty("Name") {
			t.Error("Should have property Name")
		}
		if !class.HasProperty("Age") {
			t.Error("Should have property Age")
		}
		if class.HasProperty("Email") {
			t.Error("Should not have property Email")
		}

		// Test GetProperty
		nameProp, found := class.GetProperty("Name")
		if !found {
			t.Error("GetProperty should find Name")
		}
		if nameProp.Name != "Name" {
			t.Errorf("Expected property Name, got %s", nameProp.Name)
		}
		if !nameProp.Type.Equals(STRING) {
			t.Errorf("Expected property type STRING, got %v", nameProp.Type)
		}

		ageProp, found := class.GetProperty("Age")
		if !found {
			t.Error("GetProperty should find Age")
		}
		if ageProp.ReadKind != PropAccessMethod {
			t.Errorf("Expected ReadKind=PropAccessMethod, got %v", ageProp.ReadKind)
		}

		_, found = class.GetProperty("Email")
		if found {
			t.Error("GetProperty should not find Email")
		}
	})

	t.Run("property inheritance", func(t *testing.T) {
		// Create parent class with a property
		parent := NewClassType("TBase", nil)
		parent.Properties["BaseProperty"] = &PropertyInfo{
			Name:      "BaseProperty",
			Type:      INTEGER,
			ReadKind:  PropAccessField,
			ReadSpec:  "FBase",
			WriteKind: PropAccessNone,
			WriteSpec: "",
		}

		// Create child class with its own property
		child := NewClassType("TDerived", parent)
		child.Properties["ChildProperty"] = &PropertyInfo{
			Name:      "ChildProperty",
			Type:      STRING,
			ReadKind:  PropAccessField,
			ReadSpec:  "FChild",
			WriteKind: PropAccessField,
			WriteSpec: "FChild",
		}

		// Test HasProperty with inheritance
		if !child.HasProperty("ChildProperty") {
			t.Error("Should have own property ChildProperty")
		}
		if !child.HasProperty("BaseProperty") {
			t.Error("Should have inherited property BaseProperty")
		}

		// Test GetProperty with inheritance
		baseProp, found := child.GetProperty("BaseProperty")
		if !found {
			t.Error("GetProperty should find inherited BaseProperty")
		}
		if baseProp.Name != "BaseProperty" {
			t.Errorf("Expected property BaseProperty, got %s", baseProp.Name)
		}

		childProp, found := child.GetProperty("ChildProperty")
		if !found {
			t.Error("GetProperty should find own ChildProperty")
		}
		if childProp.Name != "ChildProperty" {
			t.Errorf("Expected property ChildProperty, got %s", childProp.Name)
		}
	})
}
