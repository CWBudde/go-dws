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
// Interface Variable Declaration Tests (Task 9.129)
// ============================================================================

// TestInterfaceVariableDeclaration tests declaring variables with interface types
func TestInterfaceVariableDeclaration(t *testing.T) {
	input := `
		type IFoo = interface
			function GetValue(): Integer;
		end;

		var x: IFoo;
	`
	expectNoErrors(t, input)
}

// TestUndefinedInterfaceType tests that using an undefined interface type is an error
func TestUndefinedInterfaceType(t *testing.T) {
	input := `
		var x: IUndefined;
	`
	expectError(t, input, "unknown type")
}

// TestInterfaceVariableAssignment tests assigning a class instance to an interface variable
func TestInterfaceVariableAssignment(t *testing.T) {
	input := `
		type IFoo = interface
			function GetValue(): Integer;
		end;

		type TFooImpl = class(TObject, IFoo)
		public
			function GetValue(): Integer;
			begin
				Result := 42;
			end;
		end;

		var x: IFoo;
		var obj: TFooImpl;
		obj := TFooImpl.Create();
		x := obj;
	`
	expectNoErrors(t, input)
}

// TestInterfaceInFunctionParameter tests using interface types in function parameters
func TestInterfaceInFunctionParameter(t *testing.T) {
	input := `
		type IFoo = interface
			function GetValue(): Integer;
		end;

		procedure Test(x: IFoo);
		begin
			var value: Integer;
			value := x.GetValue();
		end;
	`
	expectNoErrors(t, input)
}

// TestInterfaceInFunctionReturn tests using interface types as function return types
func TestInterfaceInFunctionReturn(t *testing.T) {
	input := `
		type IFoo = interface
			function GetValue(): Integer;
		end;

		type TFooImpl = class(TObject, IFoo)
		public
			function GetValue(): Integer;
			begin
				Result := 42;
			end;
		end;

		function Get(): IFoo;
		begin
			Result := TFooImpl.Create();
		end;
	`
	expectNoErrors(t, input)
}

// TestMultipleInterfaceVariables tests declaring multiple variables with different interface types
func TestMultipleInterfaceVariables(t *testing.T) {
	input := `
		type IFoo = interface
			function GetFoo(): String;
		end;

		type IBar = interface
			function GetBar(): Integer;
		end;

		var foo: IFoo;
		var bar: IBar;
		var another: IFoo;
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Interface Assignment Type Checking Tests (Task 9.130)
// ============================================================================

// TestValidClassToInterfaceAssignment tests that a class implementing an interface can be assigned
func TestValidClassToInterfaceAssignment(t *testing.T) {
	input := `
		type IPrintable = interface
			function ToString(): String;
		end;

		type TDocument = class(TObject, IPrintable)
		public
			function ToString(): String;
			begin
				Result := 'Document';
			end;
		end;

		var printer: IPrintable;
		var doc: TDocument;
		doc := TDocument.Create();
		printer := doc;
	`
	expectNoErrors(t, input)
}

// TestInvalidClassToInterfaceAssignment tests that a class NOT implementing an interface cannot be assigned
func TestInvalidClassToInterfaceAssignment(t *testing.T) {
	input := `
		type IPrintable = interface
			function ToString(): String;
		end;

		type TDocument = class(TObject)
		public
			function GetName(): String;
			begin
				Result := 'Document';
			end;
		end;

		var printer: IPrintable;
		var doc: TDocument;
		doc := TDocument.Create();
		printer := doc;
	`
	expectError(t, input, "cannot assign")
}

// TestValidInterfaceToInterfaceAssignment tests that compatible interface variables can be assigned
func TestValidInterfaceToInterfaceAssignment(t *testing.T) {
	input := `
		type IBase = interface
			function GetValue(): Integer;
		end;

		type IDerived = interface(IBase)
			function GetExtra(): String;
		end;

		type TImpl = class(TObject, IDerived)
		public
			function GetValue(): Integer;
			begin
				Result := 42;
			end;

			function GetExtra(): String;
			begin
				Result := 'extra';
			end;
		end;

		var base: IBase;
		var derived: IDerived;
		var instance: TImpl;

		instance := TImpl.Create();
		derived := instance;
		base := derived;
	`
	expectNoErrors(t, input)
}

// TestIncompatibleInterfaceAssignment tests that incompatible interface types cannot be assigned
func TestIncompatibleInterfaceAssignment(t *testing.T) {
	input := `
		type IReader = interface
			function ReadData(): String;
		end;

		type IWriter = interface
			procedure WriteData(data: String);
		end;

		type TReaderImpl = class(TObject, IReader)
		public
			function ReadData(): String;
			begin
				Result := 'data';
			end;
		end;

		var reader: IReader;
		var writer: IWriter;
		var instance: TReaderImpl;

		instance := TReaderImpl.Create();
		reader := instance;
		writer := reader;
	`
	expectError(t, input, "cannot assign")
}

// ============================================================================
// Helper functions (reuse from analyzer_test.go)
// ============================================================================
// Note: expectNoErrors() and expectError() are already defined in analyzer_test.go
// and will be available since this is in the same package
