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
	parent.Fields["id"] = INTEGER

	child := NewClassType("TPerson", parent)
	child.Fields["name"] = STRING
	child.Fields["age"] = INTEGER

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
	parent.AddMethodOverload("ToString", &MethodInfo{
		Signature: NewFunctionType([]Type{}, STRING),
	})

	child := NewClassType("TPerson", parent)
	child.AddMethodOverload("GetAge", &MethodInfo{
		Signature: NewFunctionType([]Type{}, INTEGER),
	})
	child.AddMethodOverload("SetName", &MethodInfo{
		Signature: NewProcedureType([]Type{STRING}),
	})

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
// External Class Tests
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
