package types

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// Mock types for testing (these mirror the real types in interp package)
type mockClassInfo struct {
	Name string
}

type mockRecordTypeValue struct {
	Name     string
	Metadata any
}

// GetMetadata implements the interface expected by LookupRecordMetadata.
// Task 3.5.128d: Test support for metadata lookup.
func (m *mockRecordTypeValue) GetMetadata() any {
	return m.Metadata
}

type mockInterfaceInfo struct {
	Name string
}

type mockHelperInfo struct {
	Name string
}

// TestNewTypeSystem verifies that NewTypeSystem initializes all registries correctly.
func TestNewTypeSystem(t *testing.T) {
	ts := NewTypeSystem()
	if ts == nil {
		t.Fatal("NewTypeSystem() returned nil")
	}

	// Verify sub-registries are initialized
	if ts.classRegistry == nil {
		t.Error("classRegistry is nil")
	}
	if ts.functionRegistry == nil {
		t.Error("functionRegistry is nil")
	}
	if ts.records == nil {
		t.Error("records map is nil")
	}
	if ts.interfaces == nil {
		t.Error("interfaces map is nil")
	}
	if ts.helpers == nil {
		t.Error("helpers map is nil")
	}
	if ts.operators == nil {
		t.Error("operators registry is nil")
	}
	if ts.conversions == nil {
		t.Error("conversions registry is nil")
	}

	// Verify RTTI type ID registries are initialized
	if ts.classTypeIDs == nil {
		t.Error("classTypeIDs is nil")
	}
	if ts.recordTypeIDs == nil {
		t.Error("recordTypeIDs is nil")
	}
	if ts.enumTypeIDs == nil {
		t.Error("enumTypeIDs is nil")
	}

	// Verify RTTI type ID starting values
	if ts.nextClassTypeID != 1000 {
		t.Errorf("nextClassTypeID = %d, want 1000", ts.nextClassTypeID)
	}
	if ts.nextRecordTypeID != 200000 {
		t.Errorf("nextRecordTypeID = %d, want 200000", ts.nextRecordTypeID)
	}
	if ts.nextEnumTypeID != 300000 {
		t.Errorf("nextEnumTypeID = %d, want 300000", ts.nextEnumTypeID)
	}
}

// TestClassRegistryDelegation tests TypeSystem's delegation to ClassRegistry.
func TestClassRegistryDelegation(t *testing.T) {
	ts := NewTypeSystem()

	// Test RegisterClass and LookupClass
	// Note: ClassInfo is type alias for 'any', so we can pass any value
	mockClass := &mockClassInfo{Name: "TestClass"}
	ts.RegisterClass("TestClass", mockClass)

	result := ts.LookupClass("TestClass")
	if result == nil {
		t.Error("LookupClass returned nil for registered class")
	}

	// Test case-insensitive lookup
	result = ts.LookupClass("testclass")
	if result == nil {
		t.Error("LookupClass failed case-insensitive lookup")
	}

	result = ts.LookupClass("TESTCLASS")
	if result == nil {
		t.Error("LookupClass failed uppercase lookup")
	}

	// Test HasClass
	if !ts.HasClass("TestClass") {
		t.Error("HasClass returned false for registered class")
	}
	if !ts.HasClass("testclass") {
		t.Error("HasClass failed case-insensitive check")
	}
	if ts.HasClass("NonExistentClass") {
		t.Error("HasClass returned true for non-existent class")
	}

	// Test LookupClass for non-existent class
	result = ts.LookupClass("NonExistent")
	if result != nil {
		t.Error("LookupClass returned non-nil for non-existent class")
	}
}

// TestClassHierarchy tests class hierarchy methods.
func TestClassHierarchy(t *testing.T) {
	ts := NewTypeSystem()

	// Register parent and child classes
	parent := &mockClassInfo{Name: "Parent"}
	child := &mockClassInfo{Name: "Child"}

	ts.RegisterClass("Parent", parent)
	ts.RegisterClassWithParent("Child", child, "Parent")

	// Test IsClassDescendantOf
	if !ts.IsClassDescendantOf("Child", "Parent") {
		t.Error("IsClassDescendantOf returned false for valid inheritance")
	}
	if ts.IsClassDescendantOf("Parent", "Child") {
		t.Error("IsClassDescendantOf returned true for reverse inheritance")
	}

	// Test GetClassDepth
	parentDepth := ts.GetClassDepth("Parent")
	if parentDepth != 0 {
		t.Errorf("Parent depth = %d, want 0", parentDepth)
	}

	childDepth := ts.GetClassDepth("Child")
	if childDepth != 1 {
		t.Errorf("Child depth = %d, want 1", childDepth)
	}

	nonExistentDepth := ts.GetClassDepth("NonExistent")
	if nonExistentDepth != -1 {
		t.Errorf("Non-existent class depth = %d, want -1", nonExistentDepth)
	}
}

// TestRecordRegistry tests record registration and lookup.
func TestRecordRegistry(t *testing.T) {
	ts := NewTypeSystem()

	// Test RegisterRecord and LookupRecord
	mockRecord := &mockRecordTypeValue{Name: "TestRecord"}
	ts.RegisterRecord("TestRecord", mockRecord)

	result := ts.LookupRecord("TestRecord")
	if result == nil {
		t.Error("LookupRecord returned nil for registered record")
	}

	// Test case-insensitive lookup
	result = ts.LookupRecord("testrecord")
	if result == nil {
		t.Error("LookupRecord failed case-insensitive lookup")
	}

	// Test HasRecord
	if !ts.HasRecord("TestRecord") {
		t.Error("HasRecord returned false for registered record")
	}
	if !ts.HasRecord("TESTRECORD") {
		t.Error("HasRecord failed uppercase check")
	}
	if ts.HasRecord("NonExistent") {
		t.Error("HasRecord returned true for non-existent record")
	}

	// Test RegisterRecord with nil (should not panic)
	ts.RegisterRecord("NilRecord", nil)
	if ts.HasRecord("NilRecord") {
		t.Error("HasRecord returned true for nil record")
	}

	// Test LookupRecordMetadata (Task 3.5.128d)
	mockMetadata := "test-metadata-value"
	recordWithMetadata := &mockRecordTypeValue{
		Name:     "RecordWithMetadata",
		Metadata: mockMetadata,
	}
	ts.RegisterRecord("RecordWithMetadata", recordWithMetadata)

	retrievedMetadata := ts.LookupRecordMetadata("RecordWithMetadata")
	if retrievedMetadata == nil {
		t.Error("LookupRecordMetadata returned nil for record with metadata")
	}
	if retrievedMetadata.(string) != mockMetadata {
		t.Errorf("LookupRecordMetadata returned %v, want %v", retrievedMetadata, mockMetadata)
	}

	// Test case-insensitive metadata lookup
	retrievedMetadata = ts.LookupRecordMetadata("recordwithmetadata")
	if retrievedMetadata.(string) != mockMetadata {
		t.Error("LookupRecordMetadata failed case-insensitive lookup")
	}

	// Test LookupRecordMetadata for non-existent record
	nilMetadata := ts.LookupRecordMetadata("NonExistentRecord")
	if nilMetadata != nil {
		t.Error("LookupRecordMetadata should return nil for non-existent record")
	}

	// Test LookupRecordMetadata for record without GetMetadata method
	recordWithoutMetadata := &mockRecordTypeValue{Name: "NoMetadata"}
	ts.RegisterRecord("RecordNoMeta", recordWithoutMetadata)
	metadataResult := ts.LookupRecordMetadata("RecordNoMeta")
	// Should return nil for record with nil metadata
	if metadataResult != nil {
		t.Error("LookupRecordMetadata should return nil for record with nil metadata")
	}
}

// TestInterfaceRegistry tests interface registration and lookup.
func TestInterfaceRegistry(t *testing.T) {
	ts := NewTypeSystem()

	// Test RegisterInterface and LookupInterface
	mockInterface := &mockInterfaceInfo{Name: "TestInterface"}
	ts.RegisterInterface("TestInterface", mockInterface)

	result := ts.LookupInterface("TestInterface")
	if result == nil {
		t.Error("LookupInterface returned nil for registered interface")
	}

	// Test case-insensitive lookup
	result = ts.LookupInterface("testinterface")
	if result == nil {
		t.Error("LookupInterface failed case-insensitive lookup")
	}

	// Test HasInterface
	if !ts.HasInterface("TestInterface") {
		t.Error("HasInterface returned false for registered interface")
	}
	if ts.HasInterface("NonExistent") {
		t.Error("HasInterface returned true for non-existent interface")
	}

	// Test RegisterInterface with nil (should not panic)
	ts.RegisterInterface("NilInterface", nil)
	if ts.HasInterface("NilInterface") {
		t.Error("HasInterface returned true for nil interface")
	}
}

// TestFunctionRegistry tests function registration and lookup.
func TestFunctionRegistry(t *testing.T) {
	ts := NewTypeSystem()

	// Create mock function declaration
	mockFn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
	}

	// Test RegisterFunction
	ts.RegisterFunction("TestFunc", mockFn)

	// Test LookupFunctions
	results := ts.LookupFunctions("TestFunc")
	if len(results) != 1 {
		t.Errorf("LookupFunctions returned %d functions, want 1", len(results))
	}

	// Test case-insensitive lookup
	results = ts.LookupFunctions("testfunc")
	if len(results) != 1 {
		t.Error("LookupFunctions failed case-insensitive lookup")
	}

	// Test HasFunction
	if !ts.HasFunction("TestFunc") {
		t.Error("HasFunction returned false for registered function")
	}
	if ts.HasFunction("NonExistent") {
		t.Error("HasFunction returned true for non-existent function")
	}

	// Test function overloading
	mockFn2 := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestFunc"},
	}
	ts.RegisterFunction("TestFunc", mockFn2)

	results = ts.LookupFunctions("TestFunc")
	if len(results) != 2 {
		t.Errorf("LookupFunctions returned %d overloads, want 2", len(results))
	}

	// Test GetFunctionOverloadCount
	count := ts.GetFunctionOverloadCount("TestFunc")
	if count != 2 {
		t.Errorf("GetFunctionOverloadCount = %d, want 2", count)
	}
}

// TestHelperRegistry tests helper method registration and lookup.
func TestHelperRegistry(t *testing.T) {
	ts := NewTypeSystem()

	// Test RegisterHelper and LookupHelpers
	mockHelper := &mockHelperInfo{Name: "TestHelper"}
	ts.RegisterHelper("String", mockHelper)

	results := ts.LookupHelpers("String")
	if len(results) != 1 {
		t.Errorf("LookupHelpers returned %d helpers, want 1", len(results))
	}

	// Test case-insensitive lookup (helpers use lowercase keys)
	results = ts.LookupHelpers("string")
	if len(results) != 1 {
		t.Error("LookupHelpers failed lowercase lookup")
	}

	results = ts.LookupHelpers("STRING")
	if len(results) != 1 {
		t.Error("LookupHelpers failed uppercase lookup")
	}

	// Test HasHelpers
	if !ts.HasHelpers("String") {
		t.Error("HasHelpers returned false for registered helper")
	}
	if !ts.HasHelpers("string") {
		t.Error("HasHelpers failed lowercase check")
	}
	if ts.HasHelpers("NonExistent") {
		t.Error("HasHelpers returned true for non-existent helper")
	}

	// Test RegisterHelper with nil (should not panic)
	ts.RegisterHelper("NilHelper", nil)
	if ts.HasHelpers("NilHelper") {
		t.Error("HasHelpers returned true for nil helper")
	}

	// Test multiple helpers for same type
	mockHelper2 := &mockHelperInfo{Name: "TestHelper2"}
	ts.RegisterHelper("String", mockHelper2)

	results = ts.LookupHelpers("String")
	if len(results) != 2 {
		t.Errorf("LookupHelpers returned %d helpers after adding second, want 2", len(results))
	}
}

// TestRTTITypeIDs tests RTTI type ID allocation and retrieval.
func TestRTTITypeIDs(t *testing.T) {
	ts := NewTypeSystem()

	// Test class type IDs
	id1 := ts.GetOrAllocateClassTypeID("TestClass1")
	if id1 != 1000 {
		t.Errorf("First class type ID = %d, want 1000", id1)
	}

	id2 := ts.GetOrAllocateClassTypeID("TestClass2")
	if id2 != 1001 {
		t.Errorf("Second class type ID = %d, want 1001", id2)
	}

	// Test case-insensitive allocation (should return same ID)
	id3 := ts.GetOrAllocateClassTypeID("testclass1")
	if id3 != id1 {
		t.Errorf("Case-insensitive lookup returned different ID: %d vs %d", id3, id1)
	}

	// Test GetClassTypeID (without allocating)
	existingID := ts.GetClassTypeID("TestClass1")
	if existingID != id1 {
		t.Errorf("GetClassTypeID returned %d, want %d", existingID, id1)
	}

	nonExistentID := ts.GetClassTypeID("NonExistent")
	if nonExistentID != 0 {
		t.Errorf("GetClassTypeID for non-existent class = %d, want 0", nonExistentID)
	}

	// Test record type IDs
	recID1 := ts.GetOrAllocateRecordTypeID("TestRecord1")
	if recID1 != 200000 {
		t.Errorf("First record type ID = %d, want 200000", recID1)
	}

	recID2 := ts.GetOrAllocateRecordTypeID("TestRecord2")
	if recID2 != 200001 {
		t.Errorf("Second record type ID = %d, want 200001", recID2)
	}

	// Test enum type IDs
	enumID1 := ts.GetOrAllocateEnumTypeID("TestEnum1")
	if enumID1 != 300000 {
		t.Errorf("First enum type ID = %d, want 300000", enumID1)
	}

	enumID2 := ts.GetOrAllocateEnumTypeID("TestEnum2")
	if enumID2 != 300001 {
		t.Errorf("Second enum type ID = %d, want 300001", enumID2)
	}

	// Test GetRecordTypeID and GetEnumTypeID
	if ts.GetRecordTypeID("TestRecord1") != recID1 {
		t.Error("GetRecordTypeID returned wrong ID")
	}
	if ts.GetEnumTypeID("TestEnum1") != enumID1 {
		t.Error("GetEnumTypeID returned wrong ID")
	}
}

// TestOperatorRegistry tests access to operator registry.
func TestOperatorRegistry(t *testing.T) {
	ts := NewTypeSystem()

	opReg := ts.Operators()
	if opReg == nil {
		t.Error("Operators() returned nil")
	}

	// Verify it's the same instance
	if ts.Operators() != opReg {
		t.Error("Operators() returned different instance on second call")
	}
}

// TestConversionRegistry tests access to conversion registry.
func TestConversionRegistry(t *testing.T) {
	ts := NewTypeSystem()

	convReg := ts.Conversions()
	if convReg == nil {
		t.Error("Conversions() returned nil")
	}

	// Verify it's the same instance
	if ts.Conversions() != convReg {
		t.Error("Conversions() returned different instance on second call")
	}
}

// TestAllMethods tests the "All*" getter methods.
func TestAllMethods(t *testing.T) {
	ts := NewTypeSystem()

	// Register some test data
	ts.RegisterClass("Class1", &mockClassInfo{Name: "Class1"})
	ts.RegisterClass("Class2", &mockClassInfo{Name: "Class2"})
	ts.RegisterRecord("Record1", &mockRecordTypeValue{Name: "Record1"})
	ts.RegisterInterface("Interface1", &mockInterfaceInfo{Name: "Interface1"})
	ts.RegisterHelper("String", &mockHelperInfo{Name: "Helper1"})

	mockFn := &ast.FunctionDecl{Name: &ast.Identifier{Value: "Func1"}}
	ts.RegisterFunction("Func1", mockFn)

	// Test AllClasses
	classes := ts.AllClasses()
	if len(classes) != 2 {
		t.Errorf("AllClasses returned %d classes, want 2", len(classes))
	}

	// Test AllRecords
	records := ts.AllRecords()
	if len(records) != 1 {
		t.Errorf("AllRecords returned %d records, want 1", len(records))
	}

	// Test AllInterfaces
	interfaces := ts.AllInterfaces()
	if len(interfaces) != 1 {
		t.Errorf("AllInterfaces returned %d interfaces, want 1", len(interfaces))
	}

	// Test AllHelpers
	helpers := ts.AllHelpers()
	if len(helpers) != 1 {
		t.Errorf("AllHelpers returned %d helper types, want 1", len(helpers))
	}

	// Test AllFunctions
	functions := ts.AllFunctions()
	if len(functions) != 1 {
		t.Errorf("AllFunctions returned %d function names, want 1", len(functions))
	}
}

// TestDirectRegistryAccess tests the Classes() and Functions() accessor methods.
func TestDirectRegistryAccess(t *testing.T) {
	ts := NewTypeSystem()

	// Test Classes() accessor
	classReg := ts.Classes()
	if classReg == nil {
		t.Error("Classes() returned nil")
	}
	if classReg != ts.classRegistry {
		t.Error("Classes() returned different registry than internal")
	}

	// Test Functions() accessor
	funcReg := ts.Functions()
	if funcReg == nil {
		t.Error("Functions() returned nil")
	}
	if funcReg != ts.functionRegistry {
		t.Error("Functions() returned different registry than internal")
	}
}
