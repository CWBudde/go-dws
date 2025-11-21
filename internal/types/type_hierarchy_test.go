package types

import (
	"testing"
)

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
	tObject.AddMethodOverload("ToString", &MethodInfo{
		Signature: NewFunctionType([]Type{}, STRING),
	})

	tStream := NewClassType("TStream", tObject)
	tStream.Fields["size"] = INTEGER
	tStream.AddMethodOverload("Read", &MethodInfo{
		Signature: NewFunctionType([]Type{INTEGER}, INTEGER),
	})

	tFileStream := NewClassType("TFileStream", tStream)
	tFileStream.Fields["filename"] = STRING

	tMemoryStream := NewClassType("TMemoryStream", tStream)
	tMemoryStream.Fields["memory"] = INTEGER

	tPersistent := NewClassType("TPersistent", tObject)
	tPersistent.AddMethodOverload("Assign", &MethodInfo{
		Signature: NewProcedureType([]Type{}),
	})

	tComponent := NewClassType("TComponent", tPersistent)
	tComponent.Fields["name"] = STRING

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
	tFile.AddMethodOverload("Read", &MethodInfo{
		Signature: NewFunctionType([]Type{}, STRING),
	})
	tFile.AddMethodOverload("Write", &MethodInfo{
		Signature: NewProcedureType([]Type{STRING}),
	})
	tFile.AddMethodOverload("Close", &MethodInfo{
		Signature: NewProcedureType([]Type{}),
	})

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
