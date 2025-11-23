package interp

import (
	"bytes"
	"fmt"
	"testing"

	internalTypes "github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"

	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
)

// TestTypeRegistryStandalone tests that the TypeSystem (TypeRegistry) works correctly
// as a shared service between Interpreter and Evaluator.
// Task 3.5.46: This test verifies that type registration and lookup work without
// needing the adapter pattern.
func TestTypeRegistryStandalone(t *testing.T) {
	// Create a shared TypeSystem
	ts := interptypes.NewTypeSystem()

	// Test 1: Register and lookup classes
	t.Run("ClassRegistration", func(t *testing.T) {
		classInfo := &ClassInfo{
			Name:   "TMyClass",
			Parent: nil,
			Fields: make(map[string]internalTypes.Type),
		}

		// Register class
		ts.RegisterClass("TMyClass", classInfo)

		// Verify lookup (case-insensitive)
		if !ts.HasClass("TMyClass") {
			t.Error("HasClass(TMyClass) = false, want true")
		}
		if !ts.HasClass("tmyclass") {
			t.Error("HasClass(tmyclass) = false, want true (case-insensitive)")
		}
		if !ts.HasClass("TMYCLASS") {
			t.Error("HasClass(TMYCLASS) = false, want true (case-insensitive)")
		}

		// Lookup class
		looked := ts.LookupClass("TMyClass")
		if looked == nil {
			t.Fatal("LookupClass(TMyClass) = nil, want ClassInfo")
		}
		lookedClass := looked.(*ClassInfo)
		if lookedClass.Name != "TMyClass" {
			t.Errorf("LookupClass(TMyClass).Name = %s, want TMyClass", lookedClass.Name)
		}
	})

	// Test 2: Register and lookup records
	t.Run("RecordRegistration", func(t *testing.T) {
		recordTypeValue := &RecordTypeValue{
			RecordType: &internalTypes.RecordType{
				Name:   "TMyRecord",
				Fields: make(map[string]internalTypes.Type),
			},
			Methods:       make(map[string]*ast.FunctionDecl),
			StaticMethods: make(map[string]*ast.FunctionDecl),
		}

		// Register record
		ts.RegisterRecord("TMyRecord", recordTypeValue)

		// Verify lookup (case-insensitive)
		if !ts.HasRecord("TMyRecord") {
			t.Error("HasRecord(TMyRecord) = false, want true")
		}
		if !ts.HasRecord("tmyrecord") {
			t.Error("HasRecord(tmyrecord) = false, want true (case-insensitive)")
		}

		// Lookup record
		looked := ts.LookupRecord("TMyRecord")
		if looked == nil {
			t.Fatal("LookupRecord(TMyRecord) = nil, want RecordTypeValue")
		}
	})

	// Test 3: Register and lookup interfaces
	t.Run("InterfaceRegistration", func(t *testing.T) {
		interfaceInfo := &InterfaceInfo{
			Name:    "IMyInterface",
			Methods: make(map[string]*ast.FunctionDecl),
		}

		// Register interface
		ts.RegisterInterface("IMyInterface", interfaceInfo)

		// Verify lookup (case-insensitive)
		if !ts.HasInterface("IMyInterface") {
			t.Error("HasInterface(IMyInterface) = false, want true")
		}
		if !ts.HasInterface("imyinterface") {
			t.Error("HasInterface(imyinterface) = false, want true (case-insensitive)")
		}

		// Lookup interface
		looked := ts.LookupInterface("IMyInterface")
		if looked == nil {
			t.Fatal("LookupInterface(IMyInterface) = nil, want InterfaceInfo")
		}
	})

	// Test 4: Register and lookup helpers
	t.Run("HelperRegistration", func(t *testing.T) {
		helperInfo := &HelperInfo{
			Name:       "TMyHelper",
			TargetType: internalTypes.STRING,
			Methods:    make(map[string]*ast.FunctionDecl),
		}

		// Register helper for String type
		ts.RegisterHelper("String", helperInfo)

		// Verify lookup (case-insensitive)
		if !ts.HasHelpers("String") {
			t.Error("HasHelpers(String) = false, want true")
		}
		if !ts.HasHelpers("string") {
			t.Error("HasHelpers(string) = false, want true (case-insensitive)")
		}

		// Lookup helpers
		helpers := ts.LookupHelpers("String")
		if helpers == nil || len(helpers) == 0 {
			t.Fatal("LookupHelpers(String) = nil or empty, want helper list")
		}
	})

	// Test 5: Type ID allocation
	t.Run("TypeIDAllocation", func(t *testing.T) {
		// Allocate class type ID
		classID1 := ts.GetOrAllocateClassTypeID("TTestClass")
		if classID1 == 0 {
			t.Error("GetOrAllocateClassTypeID(TTestClass) = 0, want non-zero")
		}

		// Should return same ID for same class
		classID2 := ts.GetOrAllocateClassTypeID("TTestClass")
		if classID1 != classID2 {
			t.Errorf("GetOrAllocateClassTypeID returned different IDs: %d vs %d", classID1, classID2)
		}

		// Should be case-insensitive
		classID3 := ts.GetOrAllocateClassTypeID("ttestclass")
		if classID1 != classID3 {
			t.Errorf("GetOrAllocateClassTypeID not case-insensitive: %d vs %d", classID1, classID3)
		}

		// Allocate record type ID
		recordID := ts.GetOrAllocateRecordTypeID("TTestRecord")
		if recordID == 0 {
			t.Error("GetOrAllocateRecordTypeID(TTestRecord) = 0, want non-zero")
		}
		if recordID == classID1 {
			t.Error("Record and class type IDs should be different")
		}

		// Allocate enum type ID
		enumID := ts.GetOrAllocateEnumTypeID("TTestEnum")
		if enumID == 0 {
			t.Error("GetOrAllocateEnumTypeID(TTestEnum) = 0, want non-zero")
		}
		if enumID == classID1 || enumID == recordID {
			t.Error("Enum type ID should be different from class and record IDs")
		}
	})
}

// TestInterpreterEvaluatorSharedTypeSystem tests that both Interpreter and Evaluator
// can access the same TypeSystem instance and see each other's registrations.
// Task 3.5.46: This verifies the goal of the task - shared type registry service.
func TestInterpreterEvaluatorSharedTypeSystem(t *testing.T) {
	// Create interpreter
	out := &bytes.Buffer{}
	interp := New(out)

	// Get the type system from interpreter
	ts := interp.typeSystem

	// Register a class via TypeSystem
	classInfo := &ClassInfo{
		Name:    "TSharedClass",
		Fields:  make(map[string]internalTypes.Type),
		Methods: make(map[string]*ast.FunctionDecl),
	}
	ts.RegisterClass("TSharedClass", classInfo)

	// Test 1: Interpreter should see the registered class
	t.Run("InterpreterCanSeeTypeSystemClass", func(t *testing.T) {
		if !interp.HasClass("TSharedClass") {
			t.Error("Interpreter.HasClass(TSharedClass) = false, want true")
		}

		// Lookup through interpreter
		_, ok := interp.LookupClass("TSharedClass")
		if !ok {
			t.Error("Interpreter.LookupClass(TSharedClass) failed")
		}
	})

	// Test 2: Create evaluator with same TypeSystem
	t.Run("EvaluatorCanSeeTypeSystemClass", func(t *testing.T) {
		// Get evaluator from interpreter
		eval := interp.evaluatorInstance

		// Evaluator should see the class via its typeSystem reference
		if !eval.TypeSystem().HasClass("TSharedClass") {
			t.Error("Evaluator.TypeSystem().HasClass(TSharedClass) = false, want true")
		}

		// Verify it's the same TypeSystem instance
		if eval.TypeSystem() != ts {
			t.Error("Evaluator and Interpreter should share the same TypeSystem instance")
		}
	})

	// Test 3: Case-insensitive access
	t.Run("CaseInsensitiveAccess", func(t *testing.T) {
		// Both should support case-insensitive lookup
		if !interp.HasClass("tsharedclass") {
			t.Error("Interpreter.HasClass(tsharedclass) = false, want true (case-insensitive)")
		}
		if !interp.typeSystem.HasClass("TSHAREDCLASS") {
			t.Error("TypeSystem.HasClass(TSHAREDCLASS) = false, want true (case-insensitive)")
		}
	})
}

// TestTypeRegistryConcurrentAccess tests that the TypeRegistry can be safely accessed
// from multiple goroutines (though DWScript itself is single-threaded, this ensures
// the registry design is sound).
func TestTypeRegistryConcurrentAccess(t *testing.T) {
	ts := interptypes.NewTypeSystem()

	// Register some types
	for i := 0; i < 10; i++ {
		className := fmt.Sprintf("TClass%d", i)
		classInfo := &ClassInfo{
			Name:   className,
			Fields: make(map[string]internalTypes.Type),
		}
		ts.RegisterClass(className, classInfo)
	}

	// Concurrent lookups should work
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			className := fmt.Sprintf("TClass%d", idx)
			if !ts.HasClass(className) {
				t.Errorf("Concurrent HasClass(%s) = false, want true", className)
			}
			looked := ts.LookupClass(className)
			if looked == nil {
				t.Errorf("Concurrent LookupClass(%s) = nil, want ClassInfo", className)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
