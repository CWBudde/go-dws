package interp

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// TestPropertyAccessor_ObjectInstance tests that ObjectInstance implements PropertyAccessor.
func TestPropertyAccessor_ObjectInstance(t *testing.T) {
	// Create a class with a property
	class := NewClassInfo("TestClass")
	class.Properties["TestProp"] = &types.PropertyInfo{
		Name:      "TestProp",
		IsIndexed: false,
		IsDefault: false,
	}
	class.Properties["DefaultProp"] = &types.PropertyInfo{
		Name:      "DefaultProp",
		IsIndexed: true,
		IsDefault: true,
	}

	// Create an object instance
	obj := NewObjectInstance(class)

	// Verify it implements PropertyAccessor
	var accessor runtime.PropertyAccessor = obj

	// Test LookupProperty
	desc := accessor.LookupProperty("TestProp")
	if desc == nil {
		t.Fatal("Expected to find TestProp, got nil")
	}
	if desc.Name != "TestProp" {
		t.Errorf("Expected property name 'TestProp', got '%s'", desc.Name)
	}
	if desc.IsDefault {
		t.Error("Expected IsDefault=false for TestProp")
	}

	// Test GetDefaultProperty
	defaultDesc := accessor.GetDefaultProperty()
	if defaultDesc == nil {
		t.Fatal("Expected to find default property, got nil")
	}
	if defaultDesc.Name != "DefaultProp" {
		t.Errorf("Expected default property name 'DefaultProp', got '%s'", defaultDesc.Name)
	}
	if !defaultDesc.IsDefault {
		t.Error("Expected IsDefault=true for DefaultProp")
	}
	if !defaultDesc.IsIndexed {
		t.Error("Expected IsIndexed=true for DefaultProp")
	}

	// Test lookup of non-existent property
	nilDesc := accessor.LookupProperty("NonExistent")
	if nilDesc != nil {
		t.Errorf("Expected nil for non-existent property, got %+v", nilDesc)
	}
}

// TestPropertyAccessor_InterfaceInstance tests that InterfaceInstance implements PropertyAccessor.
func TestPropertyAccessor_InterfaceInstance(t *testing.T) {
	// Create an interface with a property
	// Note: Properties map keys must be normalized (lowercase)
	iface := NewInterfaceInfo("TestInterface")
	iface.Properties["testprop"] = &types.PropertyInfo{
		Name:      "TestProp",
		IsIndexed: false,
		IsDefault: false,
	}
	iface.Properties["defaultprop"] = &types.PropertyInfo{
		Name:      "DefaultProp",
		IsIndexed: true,
		IsDefault: true,
	}

	// Create a class that implements the interface
	class := NewClassInfo("TestClass")
	class.Interfaces = append(class.Interfaces, iface)

	// Create an interface instance
	obj := NewObjectInstance(class)
	intfInst := NewInterfaceInstance(iface, obj)

	// Verify it implements PropertyAccessor
	var accessor runtime.PropertyAccessor = intfInst

	// Test LookupProperty
	desc := accessor.LookupProperty("TestProp")
	if desc == nil {
		t.Fatal("Expected to find TestProp, got nil")
	}
	if desc.Name != "TestProp" {
		t.Errorf("Expected property name 'TestProp', got '%s'", desc.Name)
	}

	// Test GetDefaultProperty
	defaultDesc := accessor.GetDefaultProperty()
	if defaultDesc == nil {
		t.Fatal("Expected to find default property, got nil")
	}
	if defaultDesc.Name != "DefaultProp" {
		t.Errorf("Expected default property name 'DefaultProp', got '%s'", defaultDesc.Name)
	}
	if !defaultDesc.IsDefault {
		t.Error("Expected IsDefault=true for DefaultProp")
	}
}

// TestPropertyAccessor_RecordValue tests that RecordValue implements PropertyAccessor.
func TestPropertyAccessor_RecordValue(t *testing.T) {
	// Create a record type with a property
	recordType := &types.RecordType{
		Name:   "TestRecord",
		Fields: make(map[string]types.Type),
		Properties: map[string]*types.RecordPropertyInfo{
			"testprop": {
				Name:      "TestProp",
				IsDefault: false,
			},
			"defaultprop": {
				Name:      "DefaultProp",
				IsDefault: true,
			},
		},
	}

	// Create a record value
	record := &RecordValue{
		RecordType: recordType,
		Fields:     make(map[string]Value),
	}

	// Verify it implements PropertyAccessor
	var accessor runtime.PropertyAccessor = record

	// Test LookupProperty
	desc := accessor.LookupProperty("TestProp")
	if desc == nil {
		t.Fatal("Expected to find TestProp, got nil")
	}
	if desc.Name != "TestProp" {
		t.Errorf("Expected property name 'TestProp', got '%s'", desc.Name)
	}
	if desc.IsDefault {
		t.Error("Expected IsDefault=false for TestProp")
	}

	// Test GetDefaultProperty
	defaultDesc := accessor.GetDefaultProperty()
	if defaultDesc == nil {
		t.Fatal("Expected to find default property, got nil")
	}
	if defaultDesc.Name != "DefaultProp" {
		t.Errorf("Expected default property name 'DefaultProp', got '%s'", defaultDesc.Name)
	}
	if !defaultDesc.IsDefault {
		t.Error("Expected IsDefault=true for DefaultProp")
	}

	// Test lookup of non-existent property
	nilDesc := accessor.LookupProperty("NonExistent")
	if nilDesc != nil {
		t.Errorf("Expected nil for non-existent property, got %+v", nilDesc)
	}
}

// TestPropertyAccessor_CaseInsensitive tests that property lookup is case-insensitive.
func TestPropertyAccessor_CaseInsensitive(t *testing.T) {
	// Create a class with a property
	class := NewClassInfo("TestClass")
	class.Properties["myproperty"] = &types.PropertyInfo{
		Name:      "MyProperty",
		IsIndexed: false,
		IsDefault: false,
	}

	obj := NewObjectInstance(class)
	var accessor runtime.PropertyAccessor = obj

	// Test various case variations
	testCases := []string{"MyProperty", "myproperty", "MYPROPERTY", "myProperty", "MyPrOpErTy"}

	for _, name := range testCases {
		desc := accessor.LookupProperty(name)
		if desc == nil {
			t.Errorf("Expected to find property with name '%s', got nil", name)
		} else if desc.Name != "MyProperty" {
			t.Errorf("For lookup '%s', expected property name 'MyProperty', got '%s'", name, desc.Name)
		}
	}
}
