package types

import (
	"testing"
)

func TestNewHelperRegistry(t *testing.T) {
	registry := NewHelperRegistry()

	if registry == nil {
		t.Fatal("NewHelperRegistry() returned nil")
	}

	if registry.HelperCount() != 0 {
		t.Errorf("New registry should have 0 helpers, got %d", registry.HelperCount())
	}

	if registry.TypeCount() != 0 {
		t.Errorf("New registry should have 0 types, got %d", registry.TypeCount())
	}
}

func TestRegisterHelper(t *testing.T) {
	registry := NewHelperRegistry()

	// Create a helper for String type
	stringType := STRING
	helper := NewHelperType("TStringHelper", stringType, false)
	helper.Methods["toupper"] = &FunctionType{
		ReturnType: STRING,
		Parameters: []Type{},
	}

	// Register the helper
	err := registry.RegisterHelper(helper)
	if err != nil {
		t.Fatalf("RegisterHelper() failed: %v", err)
	}

	// Verify counts
	if registry.HelperCount() != 1 {
		t.Errorf("Expected 1 helper, got %d", registry.HelperCount())
	}

	if registry.TypeCount() != 1 {
		t.Errorf("Expected 1 type with helpers, got %d", registry.TypeCount())
	}

	// Verify we can retrieve it by name
	retrieved, ok := registry.GetHelperByName("TStringHelper")
	if !ok {
		t.Fatal("Could not retrieve helper by name")
	}
	if retrieved.Name != "TStringHelper" {
		t.Errorf("Expected helper name 'TStringHelper', got '%s'", retrieved.Name)
	}
}

func TestRegisterHelperCaseInsensitive(t *testing.T) {
	registry := NewHelperRegistry()

	helper := NewHelperType("TStringHelper", STRING, false)
	err := registry.RegisterHelper(helper)
	if err != nil {
		t.Fatalf("RegisterHelper() failed: %v", err)
	}

	// Test case-insensitive lookup
	testCases := []string{
		"TStringHelper",
		"tstringhelper",
		"TSTRINGHELPER",
		"tStRiNgHeLpEr",
	}

	for _, name := range testCases {
		retrieved, ok := registry.GetHelperByName(name)
		if !ok {
			t.Errorf("Could not retrieve helper with name '%s'", name)
		}
		if retrieved.Name != "TStringHelper" {
			t.Errorf("Expected helper name 'TStringHelper', got '%s'", retrieved.Name)
		}
	}
}

func TestRegisterHelperDuplicateName(t *testing.T) {
	registry := NewHelperRegistry()

	// Register first helper
	helper1 := NewHelperType("THelper", STRING, false)
	err := registry.RegisterHelper(helper1)
	if err != nil {
		t.Fatalf("RegisterHelper() failed: %v", err)
	}

	// Try to register another helper with the same name
	helper2 := NewHelperType("THelper", INTEGER, false)
	err = registry.RegisterHelper(helper2)
	if err == nil {
		t.Error("Expected error when registering duplicate helper name, got nil")
	}
}

func TestRegisterHelperNil(t *testing.T) {
	registry := NewHelperRegistry()

	err := registry.RegisterHelper(nil)
	if err == nil {
		t.Error("Expected error when registering nil helper, got nil")
	}
}

func TestRegisterMultipleHelpersForSameType(t *testing.T) {
	registry := NewHelperRegistry()

	// Register two helpers for String type
	helper1 := NewHelperType("TStringHelper1", STRING, false)
	helper1.Methods["method1"] = &FunctionType{ReturnType: STRING, Parameters: []Type{}}

	helper2 := NewHelperType("TStringHelper2", STRING, false)
	helper2.Methods["method2"] = &FunctionType{ReturnType: INTEGER, Parameters: []Type{}}

	err := registry.RegisterHelper(helper1)
	if err != nil {
		t.Fatalf("RegisterHelper(helper1) failed: %v", err)
	}

	err = registry.RegisterHelper(helper2)
	if err != nil {
		t.Fatalf("RegisterHelper(helper2) failed: %v", err)
	}

	// Verify counts
	if registry.HelperCount() != 2 {
		t.Errorf("Expected 2 helpers, got %d", registry.HelperCount())
	}

	if registry.TypeCount() != 1 {
		t.Errorf("Expected 1 type with helpers, got %d", registry.TypeCount())
	}

	// Get all helpers for String type
	helpers := registry.GetHelpersForType(STRING)
	if len(helpers) != 2 {
		t.Fatalf("Expected 2 helpers for String, got %d", len(helpers))
	}

	// Verify order (should be in registration order)
	if helpers[0].Name != "TStringHelper1" {
		t.Errorf("Expected first helper to be 'TStringHelper1', got '%s'", helpers[0].Name)
	}
	if helpers[1].Name != "TStringHelper2" {
		t.Errorf("Expected second helper to be 'TStringHelper2', got '%s'", helpers[1].Name)
	}
}

func TestGetHelpersForType(t *testing.T) {
	registry := NewHelperRegistry()

	// Register helpers for different types
	stringHelper := NewHelperType("TStringHelper", STRING, false)
	intHelper := NewHelperType("TIntHelper", INTEGER, false)

	registry.RegisterHelper(stringHelper)
	registry.RegisterHelper(intHelper)

	// Get helpers for String
	stringHelpers := registry.GetHelpersForType(STRING)
	if len(stringHelpers) != 1 {
		t.Errorf("Expected 1 helper for String, got %d", len(stringHelpers))
	}
	if stringHelpers[0].Name != "TStringHelper" {
		t.Errorf("Expected 'TStringHelper', got '%s'", stringHelpers[0].Name)
	}

	// Get helpers for Integer
	intHelpers := registry.GetHelpersForType(INTEGER)
	if len(intHelpers) != 1 {
		t.Errorf("Expected 1 helper for Integer, got %d", len(intHelpers))
	}
	if intHelpers[0].Name != "TIntHelper" {
		t.Errorf("Expected 'TIntHelper', got '%s'", intHelpers[0].Name)
	}

	// Get helpers for type with no helpers
	floatHelpers := registry.GetHelpersForType(FLOAT)
	if len(floatHelpers) != 0 {
		t.Errorf("Expected 0 helpers for Float, got %d", len(floatHelpers))
	}
}

func TestGetHelpersForTypeNil(t *testing.T) {
	registry := NewHelperRegistry()

	helpers := registry.GetHelpersForType(nil)
	if helpers != nil {
		t.Errorf("Expected nil for nil type, got %v", helpers)
	}
}

func TestFindMethod(t *testing.T) {
	registry := NewHelperRegistry()

	// Create helper with method
	helper := NewHelperType("TStringHelper", STRING, false)
	helper.Methods["toupper"] = &FunctionType{
		ReturnType: STRING,
		Parameters: []Type{},
	}

	err := registry.RegisterHelper(helper)
	if err != nil {
		t.Fatalf("RegisterHelper() failed: %v", err)
	}

	// Find the method
	method, foundHelper, ok := registry.FindMethod(STRING, "toupper")
	if !ok {
		t.Fatal("Expected to find method 'toupper'")
	}
	if method == nil {
		t.Error("Expected non-nil method")
	}
	if foundHelper == nil {
		t.Error("Expected non-nil helper")
	}
	if foundHelper.Name != "TStringHelper" {
		t.Errorf("Expected helper 'TStringHelper', got '%s'", foundHelper.Name)
	}
}

func TestFindMethodCaseInsensitive(t *testing.T) {
	registry := NewHelperRegistry()

	helper := NewHelperType("TStringHelper", STRING, false)
	helper.Methods["toupper"] = &FunctionType{
		ReturnType: STRING,
		Parameters: []Type{},
	}

	registry.RegisterHelper(helper)

	// Test case-insensitive method lookup
	testCases := []string{
		"toupper",
		"TOUPPER",
		"ToUpper",
		"tOuPpEr",
	}

	for _, name := range testCases {
		method, _, ok := registry.FindMethod(STRING, name)
		if !ok {
			t.Errorf("Could not find method with name '%s'", name)
		}
		if method == nil {
			t.Errorf("Expected non-nil method for name '%s'", name)
		}
	}
}

func TestFindMethodPriority(t *testing.T) {
	registry := NewHelperRegistry()

	// Register two helpers for String with the same method
	helper1 := NewHelperType("THelper1", STRING, false)
	helper1.Methods["test"] = &FunctionType{
		ReturnType: STRING,
		Parameters: []Type{},
	}

	helper2 := NewHelperType("THelper2", STRING, false)
	helper2.Methods["test"] = &FunctionType{
		ReturnType: INTEGER, // Different return type
		Parameters: []Type{},
	}

	registry.RegisterHelper(helper1)
	registry.RegisterHelper(helper2)

	// Find method - should return the one from helper2 (most recent)
	method, foundHelper, ok := registry.FindMethod(STRING, "test")
	if !ok {
		t.Fatal("Expected to find method 'test'")
	}

	// Should be from helper2 (most recent)
	if foundHelper.Name != "THelper2" {
		t.Errorf("Expected helper 'THelper2', got '%s'", foundHelper.Name)
	}

	// Return type should be INTEGER (from helper2)
	if !method.ReturnType.Equals(INTEGER) {
		t.Errorf("Expected return type INTEGER, got %s", method.ReturnType)
	}
}

func TestFindMethodNotFound(t *testing.T) {
	registry := NewHelperRegistry()

	helper := NewHelperType("THelper", STRING, false)
	registry.RegisterHelper(helper)

	// Try to find non-existent method
	method, foundHelper, ok := registry.FindMethod(STRING, "nonexistent")
	if ok {
		t.Error("Expected method not found, but got success")
	}
	if method != nil {
		t.Error("Expected nil method")
	}
	if foundHelper != nil {
		t.Error("Expected nil helper")
	}
}

func TestFindProperty(t *testing.T) {
	registry := NewHelperRegistry()

	// Create helper with property
	helper := NewHelperType("TStringHelper", STRING, false)
	helper.Properties["length"] = &PropertyInfo{
		Type:     INTEGER,
		ReadSpec: "GetLength",
	}

	registry.RegisterHelper(helper)

	// Find the property
	prop, foundHelper, ok := registry.FindProperty(STRING, "length")
	if !ok {
		t.Fatal("Expected to find property 'length'")
	}
	if prop == nil {
		t.Error("Expected non-nil property")
	}
	if foundHelper == nil {
		t.Error("Expected non-nil helper")
	}
	if !prop.Type.Equals(INTEGER) {
		t.Errorf("Expected property type INTEGER, got %s", prop.Type)
	}
}

func TestFindClassVar(t *testing.T) {
	registry := NewHelperRegistry()

	// Create helper with class var
	helper := NewHelperType("THelper", STRING, false)
	helper.ClassVars["defaultencoding"] = STRING

	registry.RegisterHelper(helper)

	// Find the class var
	varType, foundHelper, ok := registry.FindClassVar(STRING, "defaultencoding")
	if !ok {
		t.Fatal("Expected to find class var 'defaultencoding'")
	}
	if varType == nil {
		t.Error("Expected non-nil var type")
	}
	if foundHelper == nil {
		t.Error("Expected non-nil helper")
	}
	if !varType.Equals(STRING) {
		t.Errorf("Expected var type STRING, got %s", varType)
	}
}

func TestFindClassConst(t *testing.T) {
	registry := NewHelperRegistry()

	// Create helper with class const
	helper := NewHelperType("TMathHelper", FLOAT, false)
	helper.ClassConsts["pi"] = 3.14159

	registry.RegisterHelper(helper)

	// Find the class const
	constVal, foundHelper, ok := registry.FindClassConst(FLOAT, "pi")
	if !ok {
		t.Fatal("Expected to find class const 'pi'")
	}
	if constVal == nil {
		t.Error("Expected non-nil const value")
	}
	if foundHelper == nil {
		t.Error("Expected non-nil helper")
	}

	// Check value
	floatVal, ok := constVal.(float64)
	if !ok {
		t.Errorf("Expected float64 value, got %T", constVal)
	}
	if floatVal != 3.14159 {
		t.Errorf("Expected pi value 3.14159, got %f", floatVal)
	}
}

func TestClear(t *testing.T) {
	registry := NewHelperRegistry()

	// Register some helpers
	helper1 := NewHelperType("THelper1", STRING, false)
	helper2 := NewHelperType("THelper2", INTEGER, false)

	registry.RegisterHelper(helper1)
	registry.RegisterHelper(helper2)

	if registry.HelperCount() != 2 {
		t.Errorf("Expected 2 helpers before clear, got %d", registry.HelperCount())
	}

	// Clear the registry
	registry.Clear()

	if registry.HelperCount() != 0 {
		t.Errorf("Expected 0 helpers after clear, got %d", registry.HelperCount())
	}

	if registry.TypeCount() != 0 {
		t.Errorf("Expected 0 types after clear, got %d", registry.TypeCount())
	}

	// Verify helpers are not retrievable
	_, ok := registry.GetHelperByName("THelper1")
	if ok {
		t.Error("Expected helper1 to not be found after clear")
	}

	helpers := registry.GetHelpersForType(STRING)
	if len(helpers) != 0 {
		t.Errorf("Expected 0 helpers for String after clear, got %d", len(helpers))
	}
}
