package types

import (
	"testing"
)

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
	tPerson.AddMethodOverload("CompareTo", &MethodInfo{
		Signature: NewFunctionType([]Type{}, INTEGER),
	})

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
	tFullImpl.AddMethodOverload("CompareTo", &MethodInfo{
		Signature: NewFunctionType([]Type{INTEGER}, INTEGER),
	})
	tFullImpl.AddMethodOverload("Equals", &MethodInfo{
		Signature: NewFunctionType([]Type{INTEGER}, BOOLEAN),
	})

	// Create class that partially implements IComparable
	tPartialImpl := NewClassType("TPartialImpl", nil)
	tPartialImpl.AddMethodOverload("CompareTo", &MethodInfo{
		Signature: NewFunctionType([]Type{INTEGER}, INTEGER),
	})
	// Missing Equals method

	// Create class with wrong signature
	tWrongSig := NewClassType("TWrongSig", nil)
	tWrongSig.AddMethodOverload("CompareTo", &MethodInfo{
		Signature: NewFunctionType([]Type{STRING}, INTEGER), // Wrong param type
	})
	tWrongSig.AddMethodOverload("Equals", &MethodInfo{
		Signature: NewFunctionType([]Type{INTEGER}, BOOLEAN),
	})

	// Create class that implements via inheritance
	tParent := NewClassType("TParent", nil)
	tParent.AddMethodOverload("CompareTo", &MethodInfo{
		Signature: NewFunctionType([]Type{INTEGER}, INTEGER),
	})
	tChild := NewClassType("TChild", tParent)
	tChild.AddMethodOverload("Equals", &MethodInfo{
		Signature: NewFunctionType([]Type{INTEGER}, BOOLEAN),
	})

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
