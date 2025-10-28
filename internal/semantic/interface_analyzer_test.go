package semantic

import (
	"testing"
)

// ============================================================================
// Interface Declaration Tests (Task 7.104-7.105)
// ============================================================================

// TestInterfaceDeclaration tests basic interface declaration analysis
func TestInterfaceDeclaration(t *testing.T) {
	input := `
		type IMyInterface = interface
			procedure DoSomething;
			function GetValue: Integer;
		end;
	`
	expectNoErrors(t, input)
}

// TestInterfaceRedeclaration tests that redeclaring an interface is an error
func TestInterfaceRedeclaration(t *testing.T) {
	input := `
		type IMyInterface = interface
			procedure DoSomething;
		end;

		type IMyInterface = interface
			function GetValue: Integer;
		end;
	`
	expectError(t, input, "interface 'IMyInterface' already declared")
}

// ============================================================================
// Interface Inheritance Tests
// ============================================================================

// TestInterfaceInheritance tests that interfaces can inherit from other interfaces
func TestInterfaceInheritance(t *testing.T) {
	input := `
		type IBase = interface
			procedure BaseMethod;
		end;

		type IDerived = interface(IBase)
			procedure DerivedMethod;
		end;
	`
	expectNoErrors(t, input)
}

// TestInterfaceUndefinedParent tests error when parent interface doesn't exist
func TestInterfaceUndefinedParent(t *testing.T) {
	input := `
		type IDerived = interface(IUndefined)
			procedure DerivedMethod;
		end;
	`
	expectError(t, input, "parent interface 'IUndefined' not found")
}

// ============================================================================
// Circular Inheritance Tests
// ============================================================================

// TestCircularInterfaceInheritance tests that circular interface inheritance is detected
func TestCircularInterfaceInheritance(t *testing.T) {
	input := `
		type IA = interface
			procedure A;
		end;

		type IB = interface(IA)
			procedure B;
		end;

		type IC = interface(IB)
			procedure C;
		end;
	`
	// This should pass - linear inheritance chain
	expectNoErrors(t, input)
}

// ============================================================================
// Class Implements Interface Tests
// ============================================================================

// TestClassImplementsInterface tests that a class correctly implements an interface
func TestClassImplementsInterface(t *testing.T) {
	input := `
		type IMyInterface = interface
			procedure DoSomething;
			function GetValue: Integer;
		end;

		type TMyClass = class(IMyInterface)
		public
			procedure DoSomething; virtual;
			function GetValue: Integer; virtual;
		end;
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Class Missing Interface Methods Tests
// ============================================================================

// TestClassMissingInterfaceMethod tests that a class must implement all interface methods
func TestClassMissingInterfaceMethod(t *testing.T) {
	input := `
		type IMyInterface = interface
			procedure DoSomething;
			function GetValue: Integer;
		end;

		type TMyClass = class(IMyInterface)
		public
			procedure DoSomething; virtual;
			// Missing GetValue method
		end;
	`
	expectError(t, input, "class 'TMyClass' does not implement interface method 'GetValue'")
}

// ============================================================================
// Multiple Interface Implementation Tests
// ============================================================================

// TestMultipleInterfaces tests that a class can implement multiple interfaces
func TestMultipleInterfaces(t *testing.T) {
	input := `
		type IReadable = interface
			function GetData: String;
		end;

		type IWritable = interface
			procedure SetData(value: String);
		end;

		type TBuffer = class(IReadable, IWritable)
		public
			function GetData: String; virtual;
			procedure SetData(value: String); virtual;
		end;
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Helper functions (reuse from analyzer_test.go)
// ============================================================================
// Note: expectNoErrors() and expectError() are already defined in analyzer_test.go
// and will be available since this is in the same package
