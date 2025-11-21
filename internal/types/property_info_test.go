package types

import (
	"testing"
)

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
