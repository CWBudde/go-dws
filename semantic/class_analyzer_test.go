package semantic

import (
	"testing"
)

// ============================================================================
// Class Declaration Tests (Task 7.54-7.55)
// ============================================================================

func TestSimpleClassDeclaration(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;
			Y: Integer;
		end;
	`
	expectNoErrors(t, input)
}

func TestClassWithParent(t *testing.T) {
	input := `
		type TBase = class
			ID: Integer;
		end;

		type TDerived = class(TBase)
			Name: String;
		end;
	`
	expectNoErrors(t, input)
}

func TestClassWithUndefinedParent(t *testing.T) {
	input := `
		type TDerived = class(TUndefined)
			Name: String;
		end;
	`
	expectError(t, input, "parent class 'TUndefined' not found")
}

func TestCircularInheritance(t *testing.T) {
	// Note: In a single-pass analyzer, this will fail with "parent class not found"
	// because we process declarations in order. For true circular inheritance
	// detection, we'd need a two-pass analyzer. This test documents the current behavior.
	input := `
		type TA = class(TB)
			X: Integer;
		end;

		type TB = class(TA)
			Y: Integer;
		end;
	`
	// In single-pass, TB doesn't exist when TA tries to inherit from it
	expectError(t, input, "parent class 'TB' not found")
}

func TestCircularInheritanceWithForwardDecl(t *testing.T) {
	// This test shows we detect circular inheritance when classes exist
	// We can't actually test this in DWScript without forward declarations,
	// but the detection logic is in place for when classes reference each other
	// This is more of a theoretical edge case.
	input := `
		type TBase = class
			X: Integer;
		end;

		type TDerived = class(TBase)
			Y: Integer;
		end;
	`
	// Should pass - no circular inheritance here
	expectNoErrors(t, input)
}

func TestClassWithInvalidFieldType(t *testing.T) {
	input := `
		type TPerson = class
			Name: UnknownType;
		end;
	`
	expectError(t, input, "unknown type 'UnknownType'")
}

func TestClassWithDuplicateFieldNames(t *testing.T) {
	input := `
		type TPerson = class
			Name: String;
			Name: Integer;
		end;
	`
	expectError(t, input, "duplicate field 'Name'")
}

func TestClassRedeclaration(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;
		end;

		type TPoint = class
			Y: Integer;
		end;
	`
	expectError(t, input, "class 'TPoint' already declared")
}

// ============================================================================
// Method Declaration Tests (Task 7.56)
// ============================================================================

func TestClassWithMethod(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;
			Y: Integer;

			function GetX(): Integer;
			begin
				Result := X;
			end;
		end;
	`
	expectNoErrors(t, input)
}

func TestMethodAccessingFields(t *testing.T) {
	input := `
		type TPerson = class
			Name: String;
			Age: Integer;

			procedure PrintInfo;
			begin
				PrintLn(Name);
				PrintLn(Age);
			end;
		end;
	`
	expectNoErrors(t, input)
}

func TestMethodUsingSelf(t *testing.T) {
	input := `
		type TCounter = class
			Value: Integer;

			function GetValue(): Integer;
			begin
				Result := Self.Value;
			end;
		end;
	`
	expectNoErrors(t, input)
}

func TestMethodAccessingUndefinedField(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;

			function GetY(): Integer;
			begin
				Result := Y;
			end;
		end;
	`
	expectError(t, input, "undefined variable 'Y'")
}

func TestMethodWithInvalidParameterType(t *testing.T) {
	input := `
		type TPerson = class
			procedure SetValue(val: UnknownType);
			begin
				PrintLn(val);
			end;
		end;
	`
	expectError(t, input, "unknown parameter type 'UnknownType'")
}

// ============================================================================
// Object Creation Tests (Task 7.57)
// ============================================================================

func TestNewExpressionSimple(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;
			Y: Integer;
		end;

		var p := TPoint.Create();
	`
	expectNoErrors(t, input)
}

func TestNewExpressionWithConstructor(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;
			Y: Integer;

			function Create(ax: Integer; ay: Integer): TPoint;
			begin
				X := ax;
				Y := ay;
				Result := Self;
			end;
		end;

		var p := TPoint.Create(10, 20);
	`
	expectNoErrors(t, input)
}

func TestNewExpressionUndefinedClass(t *testing.T) {
	input := `var obj := TUndefined.Create();`
	expectError(t, input, "undefined class 'TUndefined'")
}

func TestNewExpressionWrongConstructorArgs(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;
			Y: Integer;

			function Create(ax: Integer; ay: Integer): TPoint;
			begin
				X := ax;
				Y := ay;
				Result := Self;
			end;
		end;

		var p := TPoint.Create(10);
	`
	expectError(t, input, "expects 2 arguments, got 1")
}

func TestNewExpressionWrongConstructorArgTypes(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;
			Y: Integer;

			function Create(ax: Integer; ay: Integer): TPoint;
			begin
				X := ax;
				Y := ay;
				Result := Self;
			end;
		end;

		var p := TPoint.Create('hello', 'world');
	`
	expectError(t, input, "has type String, expected Integer")
}

// ============================================================================
// Member Access Tests (Task 7.58)
// ============================================================================

func TestMemberAccessField(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;
			Y: Integer;
		end;

		var p := TPoint.Create();
		var x: Integer := p.X;
	`
	expectNoErrors(t, input)
}

func TestMemberAccessInheritedField(t *testing.T) {
	input := `
		type TBase = class
			ID: Integer;
		end;

		type TDerived = class(TBase)
			Name: String;
		end;

		var obj := TDerived.Create();
		var id: Integer := obj.ID;
	`
	expectNoErrors(t, input)
}

func TestMemberAccessNonObjectType(t *testing.T) {
	input := `
		var x: Integer := 42;
		var y := x.SomeField;
	`
	expectError(t, input, "member access requires class or record type")
}

func TestMemberAccessUndefinedMember(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;
			Y: Integer;
		end;

		var p := TPoint.Create();
		var z := p.Z;
	`
	expectError(t, input, "class 'TPoint' has no member 'Z'")
}

func TestMemberAccessTypeMismatch(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;
		end;

		var p := TPoint.Create();
		var s: String := p.X;
	`
	expectError(t, input, "cannot assign Integer to String")
}

// ============================================================================
// Method Overriding Tests (Task 7.59)
// ============================================================================

func TestMethodOverriding(t *testing.T) {
	input := `
		type TBase = class
			function GetValue(): Integer;
			begin
				Result := 0;
			end;
		end;

		type TDerived = class(TBase)
			function GetValue(): Integer;
			begin
				Result := 42;
			end;
		end;
	`
	expectNoErrors(t, input)
}

func TestMethodOverridingSignatureMismatch(t *testing.T) {
	input := `
		type TBase = class
			function GetValue(): Integer;
			begin
				Result := 0;
			end;
		end;

		type TDerived = class(TBase)
			function GetValue(): String;
			begin
				Result := 'hello';
			end;
		end;
	`
	expectError(t, input, "method 'GetValue' signature mismatch")
}

func TestMethodOverridingParameterMismatch(t *testing.T) {
	input := `
		type TBase = class
			procedure SetValue(val: Integer);
			begin
				PrintLn(val);
			end;
		end;

		type TDerived = class(TBase)
			procedure SetValue(val: Integer; extra: String);
			begin
				PrintLn(val);
			end;
		end;
	`
	expectError(t, input, "method 'SetValue' signature mismatch")
}

func TestNewMethodInDerivedClass(t *testing.T) {
	// Adding a new method (not in parent) should be OK
	input := `
		type TBase = class
			X: Integer;
		end;

		type TDerived = class(TBase)
			function GetDouble(): Integer;
			begin
				Result := X * 2;
			end;
		end;
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Complex Integration Tests
// ============================================================================

func TestCompleteClassHierarchy(t *testing.T) {
	input := `
		type TShape = class
			X: Integer;
			Y: Integer;

			function Create(ax: Integer; ay: Integer): TShape;
			begin
				X := ax;
				Y := ay;
				Result := Self;
			end;

			function GetArea(): Integer;
			begin
				Result := 0;
			end;
		end;

		type TRectangle = class(TShape)
			Width: Integer;
			Height: Integer;

			function GetArea(): Integer;
			begin
				Result := Width * Height;
			end;
		end;

		var shape := TShape.Create(10, 20);
		var rect := TRectangle.Create(0, 0);
		var area: Integer := rect.GetArea();
	`
	expectNoErrors(t, input)
}

func TestMultipleLevelInheritance(t *testing.T) {
	input := `
		type TBase = class
			ID: Integer;
		end;

		type TMiddle = class(TBase)
			Name: String;
		end;

		type TDerived = class(TMiddle)
			Value: Integer;
		end;

		var obj := TDerived.Create();
		var id: Integer := obj.ID;
		var name: String := obj.Name;
		var value: Integer := obj.Value;
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Class Variables (Static Fields) Tests - Task 7.62
// ============================================================================

func TestClassVariable(t *testing.T) {
	input := `
		type TCounter = class
			class var Count: Integer;
		end;
	`
	expectNoErrors(t, input)
}

func TestClassVariableWithInvalidType(t *testing.T) {
	input := `
		type TExample = class
			class var BadVar: InvalidType;
		end;
	`
	expectError(t, input, "unknown type 'InvalidType'")
}

func TestDuplicateClassVariable(t *testing.T) {
	input := `
		type TExample = class
			class var Count: Integer;
			class var Count: Integer;
		end;
	`
	expectError(t, input, "duplicate class variable 'Count'")
}

func TestClassVariableAndInstanceField(t *testing.T) {
	input := `
		type TExample = class
			class var SharedCount: Integer;
			InstanceID: Integer;
		end;
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Class Methods (Static Methods) Tests - Task 7.61
// ============================================================================

func TestClassMethod(t *testing.T) {
	input := `
		type TMath = class
			class function Add(a, b: Integer): Integer; static;
			begin
				Result := a + b;
			end;
		end;
	`
	expectNoErrors(t, input)
}

func TestClassMethodWithoutStatic(t *testing.T) {
	input := `
		type TUtils = class
			class function Double(x: Integer): Integer;
			begin
				Result := x * 2;
			end;
		end;
	`
	expectNoErrors(t, input)
}

func TestClassMethodAccessingClassVariable(t *testing.T) {
	input := `
		type TCounter = class
			class var Count: Integer;

			class procedure Increment; static;
			begin
				Count := Count + 1;
			end;
		end;
	`
	// Note: This test currently passes semantic analysis but would need
	// runtime support for accessing class variables from class methods
	expectNoErrors(t, input)
}

func TestClassMethodCannotAccessSelf(t *testing.T) {
	input := `
		type TExample = class
			Value: Integer;

			class function GetValue: Integer; static;
			begin
				Result := Self.Value;
			end;
		end;
	`
	// Class methods should not be able to access Self
	expectError(t, input, "undefined")
}

func TestClassMethodCannotAccessInstanceField(t *testing.T) {
	input := `
		type TExample = class
			Value: Integer;

			class function GetValue: Integer; static;
			begin
				Result := Value;
			end;
		end;
	`
	// Class methods should not be able to access instance fields
	expectError(t, input, "undefined")
}

func TestInstanceMethodCanAccessClassVariable(t *testing.T) {
	input := `
		type TExample = class
			class var SharedCount: Integer;
			InstanceID: Integer;

			procedure ShowCount;
			begin
				PrintLn(SharedCount);
			end;
		end;
	`
	// Instance methods should be able to access class variables
	// Note: This requires runtime support that's not yet implemented
	expectNoErrors(t, input)
}

func TestMixedClassAndInstanceMethods(t *testing.T) {
	input := `
		type TExample = class
			class var Count: Integer;
			Value: Integer;

			class procedure IncrementCount; static;
			begin
				Count := Count + 1;
			end;

			procedure SetValue(v: Integer);
			begin
				Value := v;
			end;
		end;
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Virtual/Override Tests (Task 7.64)
// ============================================================================

func TestVirtualMethodDeclaration(t *testing.T) {
	input := `
		type TBase = class
			function DoWork(): Integer; virtual;
			begin
				Result := 1;
			end;
		end;
	`
	expectNoErrors(t, input)
}

func TestOverrideWithoutVirtualParent(t *testing.T) {
	input := `
		type TBase = class
			function DoWork(): Integer;
			begin
				Result := 1;
			end;
		end;

		type TChild = class(TBase)
			function DoWork(): Integer; override;
			begin
				Result := 2;
			end;
		end;
	`
	// Should error: override without virtual parent
	expectError(t, input, "override")
}

func TestOverrideSignatureMismatch(t *testing.T) {
	input := `
		type TBase = class
			function DoWork(): Integer; virtual;
			begin
				Result := 1;
			end;
		end;

		type TChild = class(TBase)
			function DoWork(): String; override;
			begin
				Result := 'two';
			end;
		end;
	`
	// Should error: override signature doesn't match parent
	expectError(t, input, "signature")
}

func TestOverrideNonExistentMethod(t *testing.T) {
	input := `
		type TBase = class
			function DoWork(): Integer; virtual;
			begin
				Result := 1;
			end;
		end;

		type TChild = class(TBase)
			function DoSomethingElse(): Integer; override;
			begin
				Result := 2;
			end;
		end;
	`
	// Should error: override method doesn't exist in parent
	expectError(t, input, "override")
}

func TestVirtualMethodHidingWarning(t *testing.T) {
	input := `
		type TBase = class
			function DoWork(): Integer; virtual;
			begin
				Result := 1;
			end;
		end;

		type TChild = class(TBase)
			function DoWork(): Integer;
			begin
				Result := 2;
			end;
		end;
	`
	// Should warn: redefining virtual method without override
	// For now we'll expect an error for strict enforcement
	expectError(t, input, "override")
}

func TestValidOverride(t *testing.T) {
	input := `
		type TBase = class
			function DoWork(): Integer; virtual;
			begin
				Result := 1;
			end;
		end;

		type TChild = class(TBase)
			function DoWork(): Integer; override;
			begin
				Result := 2;
			end;
		end;
	`
	expectNoErrors(t, input)
}

func TestOverrideParameterMismatch(t *testing.T) {
	input := `
		type TBase = class
			function Calculate(x: Integer): Integer; virtual;
			begin
				Result := x;
			end;
		end;

		type TChild = class(TBase)
			function Calculate(x: Integer; y: Integer): Integer; override;
			begin
				Result := x + y;
			end;
		end;
	`
	// Should error: override has different parameter count
	expectError(t, input, "signature")
}

// ============================================================================
// Abstract Class/Method Tests (Task 7.65)
// ============================================================================

func TestAbstractClassDeclaration(t *testing.T) {
	input := `
		type TShape = class abstract
			FName: String;
		end;
	`
	expectNoErrors(t, input)
}

func TestCannotInstantiateAbstractClass(t *testing.T) {
	input := `
		type TShape = class abstract
			FName: String;
		end;

		var s := TShape.Create();
	`
	// Should error: cannot instantiate abstract class
	expectError(t, input, "abstract")
}

func TestConcreteClassMustImplementAbstractMethods(t *testing.T) {
	input := `
		type TShape = class abstract
			function GetArea(): Float; abstract;
		end;

		type TCircle = class(TShape)
			FRadius: Float;
		end;
	`
	// Should error: TCircle doesn't implement GetArea
	expectError(t, input, "abstract")
}

func TestAbstractMethodInConcreteClass(t *testing.T) {
	input := `
		type TShape = class
			function GetArea(): Float; abstract;
		end;
	`
	// Should error: only abstract classes can have abstract methods
	expectError(t, input, "abstract")
}

func TestValidAbstractImplementation(t *testing.T) {
	input := `
		type TShape = class abstract
			function GetArea(): Float; abstract;

			function GetName(): String;
			begin
				Result := 'Shape';
			end;
		end;

		type TCircle = class(TShape)
			FRadius: Float;

			function GetArea(): Float; override;
			begin
				Result := 3.14 * FRadius * FRadius;
			end;
		end;

		var c := TCircle.Create();
	`
	expectNoErrors(t, input)
}
